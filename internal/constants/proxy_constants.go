package constants

import (
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var LoggingLevel = SetLogLevel(GetEnv("LOG_LEVEL", "info"))

// Proxy Configurations
const (
	MaxIdleConns           = 10
	ResponseHeaderTimeout  = 30 * time.Second
	IdleConnTimeout        = 30 * time.Second
	Timeout                = 5 * time.Second
	KeepAlive              = 10 * time.Second
	CORSAllowOrigin        = "*"
	CORSMethods            = "GET, POST, PUT, DELETE, OPTIONS"
	CORSHeaders            = "Content-Type, Authorization"
	RealIPHeader           = "X-Real-IP"
	ForwardedForHeader     = "X-Forwarded-For"
	ForwardedHostHeader    = "X-Forwarded-Host"
	ForwardedProtoHeader   = "X-Forwarded-Proto"
	ForwardedURIHeader     = "X-Forwarded-URI"
	ForwardedMethodHeader  = "X-Forwarded-Method"
	ForwardedPathHeader    = "X-Forwarded-Path"
	ForwardedQueryHeader   = "X-Forwarded-Query"
	ForwardedPortHeader    = "X-Forwarded-Port"
	CORSAllowOriginHeader  = "Access-Control-Allow-Origin"
	CORSAllowMethodsHeader = "Access-Control-Allow-Methods"
	CORSAllowHeadersHeader = "Access-Control-Allow-Headers"
	HeartBeatInterval      = time.Second * 60
	HeartBeatTimeout       = time.Second * 10
	PrometheusPath         = "/metrics"
	PrometheusPort         = "8091"
)

// HTTP Headers
var (
	HeadersMap = map[string]string{
		"Access-Control-Allow-Origin":      CORSAllowOrigin,    // CORS Allow Origin
		"Access-Control-Allow-Methods":     CORSMethods,        // CORS Allow Methods
		"Access-Control-Allow-Headers":     CORSHeaders,        // CORS Allow Headers
		"Content-Type":                     "application/json", // Content Type
		"Connection":                       "keep-alive",       // Keep alive connection
		"Keep-Alive":                       Timeout.String(),   // Timeout for keep alive
		"Cache-Control":                    "no-cache",         // No cache
		"Pragma":                           "no-cache",         // No cache
		"Expires":                          "0",                // Expires
		"Access-Control-Allow-Credentials": "true",             // CORS Allow Credentials
		"Access-Control-Max-Age":           "86400",            // CORS Max Age
	}
)

// prom http metrics
var (
	ProxiedRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "reverseproxy",
		Subsystem: "metrics",
		Name:      "proxied_requests_total",
		Help:      "Total number of requests proxied",
	})
	RequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "reverseproxy",
		Subsystem: "metrics",
		Name:      "request_duration_seconds",
		Help:      "Duration of proxy requests",
		Buckets:   prometheus.DefBuckets,
	})
)

// SetLogLevel sets the logging level for the application.
func SetLogLevel(level string) slog.Level {
	var LoggingLevel slog.Level
	switch level {
	case "debug":
		LoggingLevel = slog.LevelDebug
	case "info":
		LoggingLevel = slog.LevelInfo
	case "warn":
		LoggingLevel = slog.LevelWarn
	case "error":
		LoggingLevel = slog.LevelError
	default:
		LoggingLevel = slog.LevelDebug
	}

	return LoggingLevel
}

// getEnv reads an environment variable or returns a default value.
func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
