package heartbeat_test

import (
	"context"
	"net/http"
	"os"
	"reverseproxy/api/heartbeat"
	"testing"
	"time"
)

// var log = logger.NewLogger(os.Stdout, "heartbeat", constants.LoggingLevel)

// TestMain sets up the test environment and runs all tests.
func TestMain(m *testing.M) {
	// Override the logger for testing purposes
	os.Exit(m.Run())
}

// TestHeartBeat_RunHeartBeat tests the RunHeartBeat function.
func TestRunHeartBeat(t *testing.T) {
	type args struct {
		ctx context.Context
		hb  heartbeat.HeartBeat
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Heartbeat enabled",
			args: args{
				ctx: context.Background(),
				hb: heartbeat.HeartBeat{
					Enabled:           true,
					Interval:          30 * time.Second,
					Timeout:           10 * time.Second,
					RetriesBeforeFail: 3,
					ServerPort:        "9090",
					ServerHost:        "localhost",
					ServerAddrURL:     "localhost:9090",
					ServerStatus:      "healthy",
					ServerStatusURL:   "http://localhost:9090/heartbeat",
					ServerStatusPath:  "/heartbeat",
				},
			},
			wantErr: false,
		},
		{
			name: "Test Heartbeat disabled",
			args: args{
				ctx: context.Background(),
				hb: heartbeat.HeartBeat{
					Enabled:           false,
					Interval:          30 * time.Second,
					Timeout:           10 * time.Second,
					RetriesBeforeFail: 3,
					ServerPort:        "9090",
					ServerHost:        "localhost",
					ServerAddrURL:     "localhost:9090",
					ServerStatus:      "healthy",
					ServerStatusURL:   "http://localhost:9090/heartbeat",
					ServerStatusPath:  "/heartbeat",
				},
			},
			wantErr: false,
		},
	}
	// Create a mock HTTP client
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := heartbeat.RunHeartBeat(tt.args.ctx, tt.args.hb, client)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunHeartBeat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestCheckHealthClient tests the checkHealthClient function.
// TestCheckHealthClient tests the checkHealthClient function.
func TestCheckHealthClient(t *testing.T) {
	type args struct {
		ctx context.Context
		hb  heartbeat.HeartBeat
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Check Health Client with valid response",
			args: args{
				ctx: context.Background(),
				hb: heartbeat.HeartBeat{
					Timeout:         10 * time.Second,
					ServerStatusURL: "http://localhost:9090/heartbeat",
				},
			},
			wantErr: true,
		},
		{
			name: "Test Check Health Client with invalid response",
			args: args{
				ctx: context.Background(),
				hb: heartbeat.HeartBeat{
					Timeout:         1 * time.Second,
					ServerStatusURL: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Timeout",
			args: args{
				ctx: context.Background(),
				hb: heartbeat.HeartBeat{
					ServerStatusURL: "http://example.com/heartbeat",
					Timeout:         1 * time.Millisecond, // Very short timeout
				},
			},
			wantErr: true,
		},
	}
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := heartbeat.CheckHealthClient(tt.args.ctx, tt.args.hb, client)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckHealthClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// TestShutdown tests the Shutdown function.
func TestShutdown(t *testing.T) {
	type args struct {
		server *http.Server
		ctx    context.Context
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test Shutdown with valid server",
			args: args{
				server: &http.Server{},
				ctx:    context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heartbeat.Shutdown(tt.args.server, tt.args.ctx)
		})
	}
}
