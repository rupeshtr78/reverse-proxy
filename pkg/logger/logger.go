package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

type Logger struct {
	Logger *slog.Logger
}

// NewLogger creates a new logger with the specified name and log level.
func NewLogger(output io.Writer, name string, level slog.Level) *Logger {
	handler := GetHandler(output, level).WithGroup(name)
	logger := slog.New(handler)

	l := &Logger{Logger: logger}

	return l
}

// GetHandler returns a slog.Handler with the specified log level.
func GetHandler(w io.Writer, level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource:   false,
		Level:       level,
		ReplaceAttr: Attrfunc,
	}
	return slog.NewJSONHandler(w, opts)
}

// log function adds caller information and request ID to log messages.
func (l *Logger) log(level slog.Level, msg string, args ...any) {
	attrs := []slog.Attr{}
	ctx := context.Background()

	if _, file, line, ok := runtime.Caller(2); ok {
		attrs = append(attrs, slog.String("file", file), slog.Int("line", line))
	}

	// Append provided arguments as attributes
	// attrs = append(attrs,)
	for _, v := range args {
		// attrs = append(attrs, slog.Any("attrs", v))
		if ia, ok := v.(slog.Attr); ok {
			attrs = append(attrs, ia)
			continue
		}

	}

	l.Logger.LogAttrs(ctx, level, msg, attrs...)
}

func Attrfunc(groups []string, attr slog.Attr) slog.Attr {
	switch attr.Key {
	case slog.TimeKey:
		attr.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05"))
	case "file":
		fullPath := attr.Value.String()
		rootDirPath, err := os.Getwd()
		if err != nil {
			return attr
		}
		rootDir := strings.Split(rootDirPath, "/")[len(strings.Split(rootDirPath, "/"))-1]
		if fullPath != "" {
			idx := strings.Index(fullPath, rootDir)
			if idx != -1 {
				attr.Value = slog.StringValue(fullPath[idx:])
			}
		}
	}
	return attr
}

func (l *Logger) Info(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	l.log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Infof(msg string, err error, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	l.log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	l.log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	l.log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	l.log(slog.LevelError, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	msg = fmt.Sprintf(msg, args...)
	l.log(slog.LevelError, msg, args...)
	os.Exit(1)
}
