package heartbeat

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reverseproxy/internal/constants"
	"reverseproxy/pkg/logger"
	"time"
)

var log = logger.NewLogger(os.Stdout, "heartbeat", constants.LoggingLevel)

type HeartBeat struct {
	Enabled           bool          // Determines whether to enable or disable the heartbeat endpoint.
	Interval          time.Duration // The interval at which heartbeats are sent (e.g., 30 seconds).
	Timeout           time.Duration // The maximum time allowed for a response from the server before considering it unhealthy (e.g., 10 seconds).
	RetriesBeforeFail int           // Number of consecutive failed heartbeat responses before marking the server as unhealthy (e.g., 3 retries).
	ServerPort        string        // The port on which the server is running (e.g., 9090).
	ServerHost        string        // The host on which the server is running (e.g., localhost).
	ServerAddrURL     string        // The URL on which the server is running (e.g., localhost:9090).
	ServerStatus      string        // The status of the server (e.g., healthy).
	ServerStatusURL   string        // The URL on which the server status is available (e.g., http://localhost:9090/heartbeat).
	ServerStatusPath  string        // The path on which the server status is available (e.g., /heartbeat).

}

// RunHeartBeat runs the heartbeat server and returns a pointer to it
func RunHeartBeat(ctx context.Context, hb HeartBeat, client *http.Client) (*http.Server, error) {
	if !hb.Enabled {
		log.Info("Heartbeat is disabled")
	}

	mux := http.NewServeMux()

	// Register heartbeat endpoint
	mux.HandleFunc(hb.ServerStatusPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, hb.ServerStatus)
	})

	server := &http.Server{
		Addr:    hb.ServerAddrURL,
		Handler: mux,
	}

	// Start a separate goroutine to send periodic heartbeats using an HTTP client with configured timeouts and retries
	go func() {
		for {
			err := CheckHealthClient(ctx, hb, client)
			if err != nil {
				log.Error("Failed to send heartbeat", err)
			} else {
				log.Info("Heartbeat sent successfully")
			}
			time.Sleep(hb.Interval)
		}
	}()

	return server, nil
}

func CheckHealthClient(ctx context.Context, hb HeartBeat, client *http.Client) error {

	if hb.ServerStatusURL == "" {
		return fmt.Errorf("server status URL is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, hb.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		hb.ServerStatusURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("heartbeat request failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func Shutdown(server *http.Server, ctx context.Context) {
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Failed to shutdown heart beat server", err)
	} else {
		log.Info("HeartBeat Server shut down")
	}
}
