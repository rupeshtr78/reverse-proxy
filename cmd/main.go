package main

import (
	"log/slog"
	"os"
	"reverseproxy/api"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
	"sync"

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

	routes := config.Routes
	// add go routine for each route

	var wg sync.WaitGroup
	wg.Add(len(routes))
	errChan := make(chan error, len(routes))
	defer close(errChan)

	for _, route := range routes {
		go func() {
			err := api.ProxyServer(&route)
			errChan <- err

		}()

	}

	for i := 0; i < len(routes); i++ {
		log.Error("Error in proxy server", <-errChan)
	}

}
