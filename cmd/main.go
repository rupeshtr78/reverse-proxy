package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"reverseproxy/api"
	"reverseproxy/internal/constants"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

var log = logger.NewLogger(os.Stdout, "main", constants.LoggingLevel)

func main() {

	var configFile = flag.String("config", "config/config.yaml", "config file path")
	var logLevel = flag.String("loglevel", "info", "log level")

	flag.Parse()

	// extract file path and file name
	configPath, configName := filepath.Split(*configFile)
	// set log level
	constants.LoggingLevel = constants.SetLogLevel(*logLevel)

	// Setup Viper
	viper.SetConfigName(configName) // name of config file (without extension)
	viper.SetConfigType("yaml")     // YAML format
	viper.AddConfigPath(configPath) // look for config in the config directory

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Error reading config file", err)
		return
	}

	log.Info("Config file loaded successfully", *configFile)
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100)*time.Millisecond)
	defer cancel()

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
	case err := <-errChan:
		log.Error("Error in proxy server", err)
	}

}
