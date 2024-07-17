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

// ProxyServer creates a new reverse proxy server for the given route.
func ProxyServer(ctx context.Context, route *reverseproxy.Route) error {
	proxy, err := reverseproxy.NewReverseProxy(ctx, route)
	if err != nil {
		log.Error("Error creating proxy", err, route.Name)
		return err
	}

	// http.Handle("/", proxy) // testing

	mux, err := reverseproxy.NewServeMux(ctx, route, proxy)
	if err != nil {
		log.Error("Error creating ServeMux", err)
		return err
	}

	log.Info("Proxy Server started for target name: %s, listening on 0.0.0.0:%d%s", route.Target.Name, route.ListenPort, route.Pattern)

	// 	// Start the server without TLS configuration
	err = http.ListenAndServe(fmt.Sprintf("10.0.0.213:%d", route.ListenPort), reverseproxy.HandleCORS(mux))
	if err != nil {
		log.Error("Error starting proxy server")
		return err
	}

	return nil
}
