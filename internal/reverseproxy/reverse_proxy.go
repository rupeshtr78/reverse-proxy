package reverseproxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reverseproxy/pkg/logger"
	"time"
)

var log = logger.NewLogger(os.Stdout, "reverseproxy", slog.LevelDebug)

// ReverseProxy is a struct that holds a Route and a Proxy. It is used to proxy HTTP requests to a target URL.
type ReverseProxy struct {
	Route *Route
	Proxy *httputil.ReverseProxy
}

// NewReverseProxy creates a new ReverseProxy instance that can be used to proxy HTTP requests to a target URL.
// The ReverseProxy is configured with the provided Route, which contains information about the target URL.
// If there is an error parsing the target URL, an error is returned.
func NewReverseProxy(ctx context.Context, route *Route) (*ReverseProxy, error) {

	target := route.Target
	url, err := getTargetURL(target)
	if err != nil {
		log.Info("Error parsing target url", err)
		return nil, err
	}

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
	// transport, err := getTlsTransport(target)
	// if err != nil {
	// 	log.Error("Error setting up TLS configuration", err)
	// 	return nil, err
	// }

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
		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Set("Access-Control-Allow-Origin", "*")
			resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			resp.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			return nil
		},
		Transport: transport,
	}
	reverseProxy := &ReverseProxy{
		Route: route,
		Proxy: proxy,
	}

	return reverseProxy, nil
}

// ServeHTTP is the HTTP handler for the ReverseProxy.
// It sets the "X-Forwarded-Host" header on the incoming request and then passes the request to the underlying ReverseProxy's ServeHTTP method.
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// // Set the X-Forwarded-Host header on the incoming request
	if p.Route.Protocol == "https" {
		r.Header.Set("X-Forwarded-Proto", "https")
	} else {
		r.Header.Set("X-Forwarded-Proto", "http")
	}

	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Forwarded-For", r.RemoteAddr)

	ctx := r.Context()
	p.Proxy.ServeHTTP(w, r.WithContext(ctx))
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)

	})

}

func getTlsTransport(target Target) (*http.Transport, error) {
	if target.Protocol != "https" {
		return nil, nil
	}

	tlsPair, err := tls.LoadX509KeyPair(target.CertFile, target.KeyFile)
	if err != nil {
		log.Error("Error loading certificate files", err)
		return nil, err
	}

	caCert, err := os.ReadFile(target.CertFile)
	if err != nil {
		log.Error("Error reading CA certificate file", err)
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsPair},
			RootCAs:      caCertPool,
		},
	}

	return transport, nil
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
