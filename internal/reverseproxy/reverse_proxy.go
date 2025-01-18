package reverseproxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reverseproxy/internal/constants"
	"reverseproxy/pkg/logger"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var log = logger.NewLogger(os.Stdout, "reverseproxy", constants.LoggingLevel)

// ReverseProxy is a struct that holds a Route and a Proxy. It is used to proxy HTTP requests to a target URL.
type ReverseProxy struct {
	Route *Route
	Proxy map[string]*httputil.ReverseProxy
}

type ReverseProxyFactory interface {
	CreateReverseProxy(ctx context.Context, route *Route) (*ReverseProxy, error)
}

// ReverseProxyFactoryImpl is an implementation of the ReverseProxyFactory interface.
type ReverseProxyFactoryImpl struct{}

// NewReverseProxy creates a new ReverseProxy instance that can be used to proxy HTTP requests to a target URL.
func NewReverseProxy(ctx context.Context, route *Route) (*ReverseProxy, error) {
	factory := &ReverseProxyFactoryImpl{}
	return factory.CreateReverseProxy(ctx, route)
}

// CreateReverseProxy creates a new ReverseProxy instance that can be used to proxy HTTP requests to a target URL.
// The ReverseProxy is configured with the provided Route, which contains information about the target URL.
// If there is an error parsing the target URL, an error is returned.
func (f *ReverseProxyFactoryImpl) CreateReverseProxy(ctx context.Context, route *Route) (*ReverseProxy, error) {

	proxies := make(map[string]*httputil.ReverseProxy)

	targets := route.Target
	for _, target := range targets {
		url, urlErr := getTargetURL(target)

		transport := &http.Transport{
			MaxIdleConns:          10,
			ResponseHeaderTimeout: 30 * time.Microsecond,
			IdleConnTimeout:       30 * time.Microsecond,
		}

		dialer := &net.Dialer{
			Timeout:   5 * time.Second,  // Timeout for establishing connection
			KeepAlive: 10 * time.Second, // Keep alive time
		}

		transport.DialContext = dialer.DialContext
		transport.ResponseHeaderTimeout = 5 * time.Second // Timeout for reading the response headers

		// // Check if protocol is HTTPS and set up TLS configuration
		tlsConfig, tlsErr := target.GetTlsTransport()

		transport.TLSClientConfig = tlsConfig

		// Setup the reverse proxy
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req = req.WithContext(ctx)
				req.URL.Scheme = url.Scheme
				req.URL.Host = url.Host
				req.URL.Path = url.Path + req.URL.Path // adds proxy path plus request url path
				log.Debug("Request proxied to %s%s", req.URL.Host, req.URL.Path)
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
				if urlErr != nil {
					log.Error("Error parsing target url", urlErr)
					http.Error(w, "Error parsing target url", http.StatusBadGateway)
				}
				if tlsErr != nil {
					log.Error("Error setting up TLS configuration", tlsErr)
					http.Error(w, "Error setting up TLS configuration", http.StatusBadGateway)
				}

				log.Error("Error proxying request", err)
				http.Error(w, fmt.Sprintf("Error Proxying request %v", http.StatusBadGateway), http.StatusBadGateway)
			},
		}

		proxies[target.PathPrefix] = proxy
		reverseProxy := &ReverseProxy{
			Route: route,
			Proxy: proxies,
		}

		return reverseProxy, nil
	}

	return nil, fmt.Errorf("no target found for route %s", route.Name)
}

// ServeHTTP is the HTTP handler for the ReverseProxy.
// It sets the "X-Forwarded-Host" header on the incoming request and then passes the request to the underlying ReverseProxy's ServeHTTP method.
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// // Set the X-Forwarded-Host header on the incoming request
	if p.Route.Protocol == "https" {
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
	constants.RequestDuration.Observe(time.Since(time.Now()).Seconds())

	ctx := r.Context()
	// p.Proxy.ServeHTTP(w, r.WithContext(ctx))
	for path, proxy := range p.Proxy {
		// if first path matches the request path, proxy the request
		log.Debug("Request path: %s", r.URL.Path)
		if r.URL.Path == path {
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}
	http.Error(w, "Not Found", http.StatusNotFound)
}

// NewServeMux creates a new HTTP request multiplexer (ServeMux) that will route incoming requests to the provided handler.
// The mux is configured to handle all requests to the root path ("/") and forward them to the provided handler.
func (p *ReverseProxy) NewServeMux(ctx context.Context, route *Route, handler http.Handler) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	// create a new route with the target path
	mux.Handle(route.Pattern, handler)
	return mux, nil
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

// getTargetURL parses the provided Target struct into a URL that can be used by the reverse proxy.
// If there is an error parsing the target URL, an error is returned.
func getTargetURL(target Target) (*url.URL, error) {

	urlString := fmt.Sprintf("%s://%s:%d", target.Protocol, target.Host, target.Port)
	targetUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	return targetUrl, nil
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
