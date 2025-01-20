package reverseproxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"reverseproxy/internal/constants"
	"reverseproxy/pkg/logger"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = logger.NewLogger(os.Stdout, "reverseproxy", constants.LoggingLevel)

// ReverseProxy is a struct that holds a Route and a Proxy. It is used to proxy HTTP requests to a target URL.
type ReverseProxy struct {
	Target Target
	Proxy  httputil.ReverseProxy
}

type ReverseProxyFactory interface {
	CreateReverseProxy(ctx context.Context, target Target) (*ReverseProxy, error)
}

// ReverseProxyFactoryImpl is an implementation of the ReverseProxyFactory interface.
type ReverseProxyFactoryImpl struct{}

// NewReverseProxy creates a new ReverseProxy instance that can be used to proxy HTTP requests to a target URL.
// func NewReverseProxy(ctx context.Context, target *Target) (*ReverseProxy, error) {
// 	factory := &ReverseProxyFactoryImpl{}
// 	return factory.CreateReverseProxy(ctx, target)

// }

// CreateReverseProxy creates a new ReverseProxy instance that can be used to proxy HTTP requests to a target URL.
// The ReverseProxy is configured with the provided Route, which contains information about the target URL.
// If there is an error parsing the target URL, an error is returned.
func CreateReverseProxy(ctx context.Context, targets []Target) (*ReverseProxy, error) {

	// Set up the transport for the reverse proxy with a timeout and keep alive time for the connection
	transport := &http.Transport{
		MaxIdleConns:          10,
		ResponseHeaderTimeout: 30 * time.Microsecond,
		IdleConnTimeout:       30 * time.Microsecond,
	}

	// Set up the dialer with a timeout and keep alive time for the connection
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,  // Timeout for establishing connection
		KeepAlive: 10 * time.Second, // Keep alive time
	}

	transport.DialContext = dialer.DialContext
	transport.ResponseHeaderTimeout = 5 * time.Second // Timeout for reading the response headers

	// // Check if protocol is HTTPS and set up TLS configuration
	// tlsConfig, tlsErr := target.GetTlsTransport()

	// transport.TLSClientConfig = tlsConfig
	target := Target{}
	// Setup the reverse proxy
	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			pathParts := strings.Split(req.URL.Path, "/")
			if len(pathParts) < 2 {
				log.Error("Invalid path")
				return
			}

			pathPrefix := pathParts[1]
			log.Debug("Path prefix: ", pathPrefix)

			for _, t := range targets {
				if t.PathPrefix == pathPrefix {
					target = t
					break
				}
			}
			log.Debug("Target path prefix: ", target.PathPrefix)
			if target.PathPrefix != pathPrefix {
				log.Error("Target not found for path prefix: ", pathPrefix)
				return
			}
			req = req.WithContext(ctx)
			req.URL.Scheme = target.Protocol
			req.URL.Host = target.Host + ":" + strconv.Itoa(target.Port)
			req.Host = target.Host
			req.URL.Path = strings.Join(pathParts[2:], "/")

			log.Debug("Proxying request to: ", req.URL.String())
		},
		// Modify the reverse proxy to add the CORS headers:
		// ModifyResponse: func(resp *http.Response) error {
		// 	resp.Header.Set("Access-Control-Allow-Origin", "*")
		// 	resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// 	resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// 	return nil
		// },
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("Error proxying request", err)
			http.Error(w, fmt.Sprintf("Error Proxying request: %v", err), http.StatusBadGateway)
		},
	}

	// Return the ReverseProxy instance

	return &ReverseProxy{
		Target: target,
		Proxy:  proxy,
	}, nil

}

// ServeHTTP is the HTTP handler for the ReverseProxy.
// implements the http.Handler interface.
// It sets the "X-Forwarded-Host" header on the incoming request and then passes the request to the underlying ReverseProxy's ServeHTTP method.
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Info("Incoming request: ", r.Method, r.URL.Path)

	// // Set the X-Forwarded-Host header on the incoming request
	if p.Target.Protocol == "https" {
		r.Header.Set(constants.ForwardedProtoHeader, "https")
	} else {
		r.Header.Set(constants.ForwardedProtoHeader, "http")
	}

	// Set the X-Forwarded-Host header on the incoming request
	r.Header.Set(constants.ForwardedHostHeader, r.Host)
	r.Header.Set(constants.ForwardedForHeader, r.RemoteAddr)
	r.Header.Set(constants.ForwardedMethodHeader, r.Method)
	r.Header.Set(constants.ForwardedPathHeader, r.URL.Path)
	r.Header.Set(constants.ForwardedQueryHeader, r.URL.RawQuery)
	r.Header.Set(constants.ForwardedPortHeader, r.URL.Port())

	// Prometheus metrics
	constants.ProxiedRequestsTotal.Inc()
	start := time.Now()
	defer func() {
		constants.RequestDuration.Observe(time.Since(start).Seconds())
	}()

	ctx := r.Context()
	p.Proxy.ServeHTTP(w, r.WithContext(ctx))

	// http.Error(w, "Not Found", http.StatusNotFound)
}

// HandleCORS is a middleware function that adds CORS headers to the response.
func HandleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(constants.CORSAllowHeadersHeader, constants.CORSHeaders)
		w.Header().Set(constants.CORSAllowMethodsHeader, constants.CORSMethods)
		w.Header().Set(constants.CORSAllowOriginHeader, constants.CORSAllowOrigin)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)

	})

}

// getTlsTransport returns a TLS configuration for the provided target. // TODO move this to the target struct
func getTlsTransport(target Target) (*tls.Config, error) {
	if target.Protocol != "https" {
		return nil, nil
	}

	_, err := os.Stat(target.CertFile)
	if err != nil {
		log.Error("Error reading certificate file", err)
		return nil, err
	}

	_, err = os.Stat(target.KeyFile)
	if err != nil {
		log.Error("Error reading key file", err)
		return nil, err
	}

	_, err = os.Stat(target.CaCert)
	if err != nil {
		log.Error("Error reading CA certificate file", err)
		return nil, err
	}

	tlsPair, err := tls.LoadX509KeyPair(target.CertFile, target.KeyFile)
	if err != nil {
		log.Error("Error loading certificate files", err)
		return nil, err
	}

	caCert, err := os.ReadFile(target.CaCert)
	if err != nil {
		log.Error("Error reading CA certificate file", err)
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsPair},
		RootCAs:      caCertPool,
	}

	return tlsConfig, nil
}

// getTargetURI parses the provided Target struct into a URL that can be used by the reverse proxy.
// If there is an error parsing the target URL, an error is returned.
func getTargetURI(r *http.Request, pattern string) (string, bool) {

	// get first part of the path
	log.Debug("Request received with URI: %v", r.URL.RequestURI())
	pathParts := strings.Split(r.URL.Path, "/")
	log.Debug("Request received with Path: ", pathParts)
	if len(pathParts) < 2 {
		log.Debug("Path len expected:  got: ", pattern, pathParts)
		return "", false
	}

	log.Debug("Request received with First Path: ", pathParts[1])

	if pathParts[1] != pattern {
		log.Debug("Path expected:  got: ", pattern, pathParts[1])
		return "", false
	}

	newRequestURI := strings.Join(pathParts[2:], "/")

	return newRequestURI, true

}

func StartMetricsServer(ctx context.Context, metricsPort string) error {

	err := prometheus.Register(constants.ProxiedRequestsTotal)
	if err != nil {
		log.Info("Prometheus metric already registered")
	}
	err = prometheus.Register(constants.RequestDuration)
	if err != nil {
		log.Info("Prometheus metric already registered")
	}

	// start a server that serves prometheus metrics on the configured port
	http.Handle(constants.PrometheusPath, promhttp.Handler())

	errChan := make(chan error, 1)

	go func() {
		log.Info("Starting metrics server on port: ", metricsPort)
		err := http.ListenAndServe(":"+metricsPort, nil)
		if err != nil {
			errChan <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down metrics server")
	default:
		err := <-errChan
		if err != nil {
			return err
		}
	}

	return nil
}
