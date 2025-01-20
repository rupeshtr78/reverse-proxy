package proxyserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reverseproxy/internal/constants"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
)

var log = logger.NewLogger(os.Stdout, "proxyserver", constants.LoggingLevel)

// ProxyServer creates a new reverse proxy server for the given route.
func ProxyServer(ctx context.Context, route *reverseproxy.Route) error {

	mux := http.NewServeMux()

	proxy, err := reverseproxy.CreateReverseProxy(ctx, route.Targets)
	if err != nil {
		log.Error("Error creating reverse proxy")
	}

	mux.Handle("/", reverseproxy.HandleCORS(proxy))

	log.Info(fmt.Sprintf("Proxy Server started  listening on %s:%d%s", route.ListenHost, route.ListenPort, route.Pattern))

	// 	// Start the server without TLS configuration
	if route.Protocol == "http" {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", route.ListenHost, route.ListenPort), mux)
		if err != nil {
			log.Error("Error starting proxy server")
			return err
		}
	} else if route.Protocol == "https" {
		// Start the server with TLS configuration
		err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", route.ListenHost, route.ListenPort), route.CertFile, route.KeyFile, mux)
		if err != nil {
			log.Error("Error starting proxy server")
			return err
		}
	} else {
		log.Error("Invalid protocol specified")
		return fmt.Errorf("invalid protocol specified")
	}

	return nil
}
