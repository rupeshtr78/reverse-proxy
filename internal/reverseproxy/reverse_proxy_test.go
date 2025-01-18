package reverseproxy

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"net/url"
// 	"testing"
// )

// // TestReverseProxy tests the creation of a ReverseProxy and its basic functionalities.
// func TestReverseProxy(t *testing.T) {
// 	target := Target{
// 		Protocol: "http",
// 		Host:     "localhost",
// 		Port:     8080,
// 	}
// 	route := &Route{
// 		Pattern: "/",
// 		Target:  target,
// 	}

// 	ctx := context.Background()
// 	proxy, err := NewReverseProxy(ctx, route)
// 	if err != nil {
// 		t.Fatalf("Failed to create reverse proxy: %v", err)
// 	}

// 	// Create test server that the reverse proxy will forward requests to
// 	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}))
// 	defer backendServer.Close()

// 	backendURL, _ := url.Parse(backendServer.URL)
// 	proxy.Proxy.Director = func(req *http.Request) {
// 		req.URL.Scheme = backendURL.Scheme
// 		req.URL.Host = backendURL.Host
// 	}

// 	// Create a request to the proxy
// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	w := httptest.NewRecorder()

// 	// Handle the request using the proxy
// 	proxy.ServeHTTP(w, req)

// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status OK but got %v", resp.Status)
// 	}
// }

// // TestHandleCORS tests the CORS middleware functionality.
// func TestHandleCORS(t *testing.T) {
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	})

// 	corsHandler := HandleCORS(handler)

// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	w := httptest.NewRecorder()
// 	corsHandler.ServeHTTP(w, req)

// 	resp := w.Result()
// 	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
// 		t.Errorf("Expected Access-Control-Allow-Origin '*' but got %v", resp.Header.Get("Access-Control-Allow-Origin"))
// 	}
// }

// // TestGetTlsTransport tests the getTlsTransport function for various scenarios.
// func TestGetTlsTransport(t *testing.T) {
// 	target := Target{
// 		Protocol: "http",
// 	}

// 	_, err := getTlsTransport(target)
// 	if err != nil {
// 		t.Errorf("Expected no error but got %v", err)
// 	}

// 	// Assume the necessary files are not present for the HTTPS test
// 	target.Protocol = "https"
// 	target.CertFile = "nonexistent_cert.pem"
// 	target.KeyFile = "nonexistent_key.pem"
// 	target.CaCert = "nonexistent_ca_cert.pem"

// 	_, err = getTlsTransport(target)
// 	if err == nil {
// 		t.Errorf("Expected error due to nonexistent certificate files but got none")
// 	}
// }
