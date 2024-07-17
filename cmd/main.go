package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"reverseproxy/api"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
	"syscall"

	"github.com/spf13/viper"
)

var log = logger.NewLogger(os.Stdout, "main", slog.LevelDebug)

func main() {

	// Setup Viper
	viper.SetConfigName("config")  // name of config file (without extension)
	viper.SetConfigType("yaml")    // YAML format
	viper.AddConfigPath("config/") // look for config in the config directory

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Error reading config file", err)
		return
	}
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

	ctx, cancel := context.WithCancel(context.Background())
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
		log.Info("Received signal %v", sig)
	case err := <-errChan:
		log.Error("Error in proxy server", err)
	}

}
