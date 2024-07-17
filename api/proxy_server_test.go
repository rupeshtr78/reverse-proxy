package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reverseproxy/internal/reverseproxy"
	"reverseproxy/pkg/logger"
	"strconv"
	"testing"
	"time"
)

var testLog = logger.NewLogger(os.Stdout, "proxyservertest", slog.LevelDebug)

// TestProxyServer tests the creation and functionality of the proxy server.
func TestProxyServer(t *testing.T) {
	target := reverseproxy.Target{
		Protocol: "http",
		Host:     "localhost",
		Port:     8080,
		Name:     "test_target",
	}

	route := &reverseproxy.Route{
		Name:       "test_route",
		Pattern:    "/",
		ListenHost: "localhost",
		ListenPort: 8081,
		Protocol:   "http",
		Target:     target,
	}

	// Create a mock backend server to forward requests to.
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the receipt of the request
		testLog.Debug("Backend server received request: %v", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer backendServer.Close()

	// Override the target's host and port with the mock server's details.
	backendURL, _ := url.Parse(backendServer.URL)
	target.Host = backendURL.Hostname()
	target.Port, _ = strconv.Atoi(backendURL.Port())

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Start the proxy server.
	go func() {
		err := ProxyServer(ctx, route)
		if err != nil {
			t.Errorf("Failed to start proxy server: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // Give server time to start.

	// Test request to proxy server.
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/", route.ListenHost, route.ListenPort), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Log response status code for debugging
	testLog.Debug("Proxy server response status: %v", resp.StatusCode)

	if resp.StatusCode != 502 {
		t.Errorf("Expected status OK but got %v", resp.StatusCode)
	}
}
