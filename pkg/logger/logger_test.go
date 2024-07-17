package logger

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, "testLogger", slog.LevelInfo)

	assert.NotNil(t, log)
	assert.NotNil(t, log.Logger)
}

func TestGetHandler(t *testing.T) {
	var buf bytes.Buffer
	handler, err := GetHandler(&buf, slog.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	handler, err = GetHandler(os.Stdout, slog.LevelInfo)
	assert.NoError(t, err)
	assert.NotNil(t, handler)

	// Testing the case where the file does not exist
	// tempFile, err := os.CreateTemp("", "testfile")
	// assert.NoError(t, err)
	// tempFile.Close()
	// os.Remove(tempFile.Name())

	// file, err := os.Open(tempFile.Name())
	// assert.Error(t, err)
	// _, err = GetHandler(file, slog.LevelInfo)
	// assert.Error(t, err)
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, "testLogger", slog.LevelInfo)

	log.Info("Info message", slog.String("key", "value"))
	output := buf.String()
	assert.Contains(t, output, `"msg":"Info message"`)
	assert.Contains(t, output, `"level":"INFO"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, "testLogger", slog.LevelDebug)

	log.Debug("Debug message", slog.String("key", "value"))
	output := buf.String()
	assert.Contains(t, output, `"msg":"Debug message"`)
	assert.Contains(t, output, `"level":"DEBUG"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, "testLogger", slog.LevelWarn)

	log.Warn("Warn message", slog.String("key", "value"))
	output := buf.String()
	assert.Contains(t, output, `"msg":"Warn message"`)
	assert.Contains(t, output, `"level":"WARN"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, "testLogger", slog.LevelError)

	log.Error("Error message", slog.String("key", "value"))
	output := buf.String()
	assert.Contains(t, output, `"msg":"Error message"`)
	assert.Contains(t, output, `"level":"ERROR"`)
	assert.Contains(t, output, `"key":"value"`)
}

func TestLogger_Fatal(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, "testLogger", slog.LevelError)

	// Wrap to catch os.Exit
	var oldExit = exitFunc
	defer func() {
		exitFunc = oldExit
	}()
	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}

	log.Fatal("Fatal message", slog.String("key", "value"))
	output := buf.String()
	assert.Contains(t, output, `"msg":"Fatal message"`)
	assert.Contains(t, output, `"level":"ERROR"`)
	assert.Contains(t, output, `"key":"value"`)
	assert.Equal(t, 1, exitCode)
}
