package test

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/spf13/viper"
)

func ProxyPoc() {
	// Setup Viper
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // YAML format
	viper.AddConfigPath(".")      // look for config in the working directory

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	// Setup the reverse proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			routes := viper.GetStringMapString("routes")
			target, ok := routes[req.Host]
			if !ok {
				return
			}
			url, _ := url.Parse(target)
			req.URL.Scheme = url.Scheme
			req.URL.Host = url.Host
			req.URL.Path = url.Path + req.URL.Path
		},
	}

	// Start the server
	http.HandleFunc("/", proxy.ServeHTTP)
	listenPort := viper.GetInt("listen_port")
	err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), nil)
	if err != nil {
		panic(err)
	}
}
