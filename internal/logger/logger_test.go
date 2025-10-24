package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	config := &Config{
		Level:  slog.LevelDebug,
		Format: "json",
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	if logger.Logger == nil {
		t.Fatal("Expected internal slog.Logger to be set")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != slog.LevelInfo {
		t.Errorf("Expected default level Info, got %v", config.Level)
	}

	if config.Format != "json" {
		t.Errorf("Expected default format json, got %s", config.Format)
	}

	if config.AddSource {
		t.Error("Expected AddSource to be false by default")
	}
}

func TestWithField(t *testing.T) {
	logger := New(DefaultConfig())
	newLogger := logger.WithField("key", "value")

	if newLogger == logger {
		t.Error("WithField should return a new logger instance")
	}

	if newLogger.Logger == nil {
		t.Error("New logger should have internal logger set")
	}
}

func TestWithFields(t *testing.T) {
	logger := New(DefaultConfig())
	fields := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	newLogger := logger.WithFields(fields)

	if newLogger == logger {
		t.Error("WithFields should return a new logger instance")
	}
}

func TestWithError(t *testing.T) {
	logger := New(DefaultConfig())

	// Test with nil error
	newLogger := logger.WithError(nil)
	if newLogger != logger {
		t.Error("WithError(nil) should return same logger")
	}

	// Test with actual error
	err := &testError{msg: "test error"}
	newLogger = logger.WithError(err)
	if newLogger == logger {
		t.Error("WithError should return a new logger instance")
	}
}

func TestComponent(t *testing.T) {
	logger := New(DefaultConfig())
	componentLogger := logger.Component("trading-engine")

	if componentLogger == logger {
		t.Error("Component should return a new logger instance")
	}
}

func TestExchange(t *testing.T) {
	logger := New(DefaultConfig())
	exchangeLogger := logger.Exchange("coinbase")

	if exchangeLogger == logger {
		t.Error("Exchange should return a new logger instance")
	}
}

func TestSymbol(t *testing.T) {
	logger := New(DefaultConfig())
	symbolLogger := logger.Symbol("BTC-USD")

	if symbolLogger == logger {
		t.Error("Symbol should return a new logger instance")
	}
}

func TestGlobalLogger(t *testing.T) {
	defaultLog := Default()
	if defaultLog == nil {
		t.Fatal("Default logger should not be nil")
	}

	// Create custom logger and set as default
	customLogger := New(&Config{
		Level:  slog.LevelDebug,
		Format: "text",
	})

	SetDefault(customLogger)
	newDefault := Default()

	if newDefault != customLogger {
		t.Error("SetDefault should update the global logger")
	}

	// Restore original default for other tests
	SetDefault(defaultLog)
}

func TestConvenienceFunctions(t *testing.T) {
	// These should not panic
	Debug("debug message", "key", "value")
	Info("info message", "key", "value")
	Warn("warn message", "key", "value")
	Error("error message", "key", "value")

	// Test convenience constructors
	_ = WithField("test", "value")
	_ = WithFields(map[string]any{"key": "value"})
	_ = Component("test")
	_ = Exchange("test")
	_ = Symbol("BTC-USD")
}

func TestTradeLogging(t *testing.T) {
	logger := New(DefaultConfig())

	// Should not panic
	logger.Trade(map[string]any{
		"side":   "buy",
		"price":  50000.0,
		"amount": 0.1,
	})
}

func TestOrderLogging(t *testing.T) {
	logger := New(DefaultConfig())

	// Should not panic
	logger.Order(map[string]any{
		"order_id": "123",
		"status":   "filled",
	})
}

func TestRiskLogging(t *testing.T) {
	logger := New(DefaultConfig())

	// Should not panic
	logger.Risk(map[string]any{
		"event":    "drawdown_limit",
		"drawdown": 0.15,
	})
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Output should contain the log message")
	}

	// Verify it's valid JSON
	var jsonData map[string]any
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}

	if jsonData["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", jsonData["msg"])
	}

	if jsonData["key"] != "value" {
		t.Errorf("Expected key='value', got %v", jsonData["key"])
	}
}

func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := &Logger{
		Logger: slog.New(handler),
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Output should contain the log message")
	}

	if !strings.Contains(output, "key=value") {
		t.Error("Output should contain key=value pair")
	}
}

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := &Logger{
		Logger: slog.New(handler),
	}

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := buf.String()

	if !strings.Contains(output, "debug") {
		t.Error("Should contain debug message")
	}
	if !strings.Contains(output, "info") {
		t.Error("Should contain info message")
	}
	if !strings.Contains(output, "warn") {
		t.Error("Should contain warn message")
	}
	if !strings.Contains(output, "error") {
		t.Error("Should contain error message")
	}
}

// Helper type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
