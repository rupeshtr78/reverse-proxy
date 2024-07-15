package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"

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
	route1 := routes[0] // for testing //TODO remove

	proxy, err := reverseproxy.NewReverseProxy(&route1)
	if err != nil {
		log.Error("Error creating proxy", err, route1.Name)
	}

	http.Handle("/", proxy)

	log.Info("Staring Porxy Server on port %s\n", config.Routes[0].ListenPort)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Routes[0].ListenPort), nil)

	if err != nil {
		log.Error("Error starting proxy server")
	}
}
