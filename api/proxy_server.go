package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
)

var log = logger.NewLogger(os.Stdout, "proxyserver", slog.LevelDebug)

// ProxyServer sets up and starts a reverse proxy server for the given route.
// It creates a new reverse proxy using the provided route configuration,
// and then creates a new ServeMux to handle the routing for the proxy.
// The server is then started and listens on the specified port.
// If any errors occur during the setup or startup, they are logged and returned.
func ProxyServer(ctx context.Context, route *reverseproxy.Route) error {
	proxy, err := reverseproxy.NewReverseProxy(ctx, route)
	if err != nil {
		log.Error("Error creating proxy", err, route.Name)
		return err
	}

	// http.Handle("/", proxy) // testing

	mux, err := reverseproxy.NewServeMux(ctx, route, proxy)
	if err != nil {
		log.Error("Error creating mux", err)
		return err
	}

	log.Info("Proxy Server created for route", route.Name, route.ListenPort)

	// err = http.ListenAndServe(fmt.Sprintf(":%d", route.ListenPort), reverseproxy.HandleCORS(mux))
	// if err != nil {
	// 	log.Error("Error starting proxy server")
	// 	return err
	// }

	// Check if the target protocol is HTTPS
	if route.Target.Protocol == "https" {
		// Start the server with TLS configuration
		err = http.ListenAndServeTLS(fmt.Sprintf(":%d", route.ListenPort), route.Target.CertFile, route.Target.KeyFile, reverseproxy.HandleCORS(mux))
		if err != nil {
			log.Error("Error starting proxy server")
			return err
		}
	} else {
		// Start the server without TLS configuration
		err = http.ListenAndServe(fmt.Sprintf(":%d", route.ListenPort), reverseproxy.HandleCORS(mux))
		if err != nil {
			log.Error("Error starting proxy server")
			return err
		}
	}

	return nil
}
