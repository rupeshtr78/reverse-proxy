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
	// add go routine for each route

	hb := heartbeat.HeartBeat{
		Enabled:           true,
		Interval:          constants.HeartBeatInterval,
		Timeout:           constants.HeartBeatTimeout,
		RetriesBeforeFail: 3,
		ServerAddrURL:     "localhost:8081",
		ServerStatusPath:  "/heartbeat",
		ServerStatus:      "Proxy Server Live",
		ServerStatusURL:   "http://localhost:8081/heartbeat",
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	hbServer, err := heartbeat.RunHeartBeat(ctx, hb)
	if err != nil {
		log.Warn("Error starting heartbeat", err)
	}

	go func() {
		err := hbServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start heartbeat server: %v", err)
		}
	}()

	err = reverseproxy.StartMetricsServer(ctx, constants.PrometheusPort)
	if err != nil {
		log.Error("Failed to start metrics server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer close(sigChan)

	// var wg sync.WaitGroup
	// wg.Add(len(routes))
	errChan := make(chan error, len(routes))
	defer close(errChan)

	for _, route := range routes {
		go func(route reverseproxy.Route) {
			err := api.ProxyServer(ctx, &route)
			errChan <- err

		}(route)

	}

	select {
	case sig := <-sigChan:
		log.Infof("Received signal %v", nil, sig)
		heartbeat.Shutdown(hbServer, ctx)
	case err := <-errChan:
		log.Error("Error in proxy server", err)
		heartbeat.Shutdown(hbServer, ctx)
	case <-ctx.Done():
		heartbeat.Shutdown(hbServer, ctx)

	}

}
