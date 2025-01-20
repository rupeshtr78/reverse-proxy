package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reverseproxy/api/heartbeat"
	api "reverseproxy/api/proxyserver"
	"reverseproxy/internal/constants"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var log = logger.NewLogger(os.Stdout, "main", constants.LoggingLevel)

func main() {

	configFile := constants.GetEnv("CONFIG_FILE_PATH", "config/config.yaml")
	// logLevel := constants.GetEnv("LOG_LEVEL", "info")

	flag.StringVar(&configFile, "config", configFile, "config file path")
	// flag.StringVar(&logLevel, "loglevel", logLevel, "log level")

	flag.Parse()

	// extract file path and file name
	configPath, configName := filepath.Split(configFile)

	// Setup Viper
	viper.SetConfigName(configName) // name of config file (without extension)
	viper.SetConfigType("yaml")     // YAML format
	viper.AddConfigPath(configPath) // look for config in the config directory

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Error reading config file", err)
		return
	}

	log.Info("Config file loaded successfully", configFile)
	config := &reverseproxy.Config{}

	err = viper.Unmarshal(config)
	if err != nil {
		panic(err)
	}

	err = config.ValidateConfig()
	if err != nil {
		log.Error("Error validating config", err)
		return
	}

	routes := config.Routes

	hb := newHeartBeat()
	client := newHTTPClient(hb.Timeout)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	hbServer, err := heartbeat.RunHeartBeat(ctx, hb, client)
	if err != nil {
		log.Warn("Error starting heartbeat", err)
	}

	startServers(ctx, hbServer, routes)
	handleSignals(ctx, hbServer)

}

func newHeartBeat() heartbeat.HeartBeat {
	return heartbeat.HeartBeat{
		Enabled:           true,
		Interval:          constants.HeartBeatInterval,
		Timeout:           constants.HeartBeatTimeout,
		RetriesBeforeFail: 3,
		ServerAddrURL:     "localhost:8081",
		ServerStatusPath:  "/heartbeat",
		ServerStatus:      "Proxy Server Live",
		ServerStatusURL:   "http://localhost:8081/heartbeat",
	}
}

func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxConnsPerHost: 1,
		},
	}
}

func startServers(ctx context.Context, hbServer *http.Server, route reverseproxy.Route) {
	go func() {
		if err := hbServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start heartbeat server", err)
		}
	}()

	go func() {
		if err := reverseproxy.StartMetricsServer(ctx, constants.PrometheusPort); err != nil {
			log.Error("Failed to start metrics server", err)
		}
	}()

	errChan := make(chan error, len(route.Targets))

	go func(route reverseproxy.Route) {
		errChan <- api.ProxyServer(ctx, &route)
	}(route)

	go func() {
		for err := range errChan {
			log.Error("Error in proxy server", err)
		}
	}()
}

func handleSignals(ctx context.Context, hbServer *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Info("Received signal %v", sig)
	case <-ctx.Done():
		log.Info("Context done")
	}

	heartbeat.Shutdown(hbServer, ctx)
}
