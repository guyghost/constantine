package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/guyghost/constantine/internal/testutils"
	"log/slog"
)

func TestStartBotComponentsIntegration(t *testing.T) {
	exchange := testutils.NewTestExchange("integration")
	manager := order.NewManager(exchange)
	strat := strategy.NewScalpingStrategy(strategy.DefaultConfig(), exchange)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- startBotComponents(ctx, strat, manager)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("startBotComponents returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("startBotComponents did not exit after cancellation")
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"empty env, default false", "", false, false},
		{"empty env, default true", "", true, true},
		{"true value", "true", false, true},
		{"1 value", "1", false, true},
		{"yes value", "yes", false, true},
		{"false value", "false", true, false},
		{"0 value", "0", true, false},
		{"no value", "no", true, false},
		{"random value", "random", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("TEST_BOOL", tt.envValue)
			} else {
				os.Unsetenv("TEST_BOOL")
			}
			defer os.Unsetenv("TEST_BOOL")

			result := getEnvBool("TEST_BOOL", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBool(%s, %v) = %v, want %v", tt.envValue, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestLoadLoggerConfig(t *testing.T) {
	// Test default config
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")
	os.Unsetenv("LOG_ADD_SOURCE")
	os.Unsetenv("LOG_OUTPUT_PATH")

	config := loadLoggerConfig()
	if config == nil {
		t.Fatal("loadLoggerConfig returned nil")
	}
	if config.Level != slog.LevelInfo {
		t.Errorf("expected default log level Info, got %v", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("expected default format json, got %s", config.Format)
	}

	// Test debug level
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")
	config = loadLoggerConfig()
	if config.Level != slog.LevelDebug {
		t.Errorf("expected debug log level, got %v", config.Level)
	}

	// Test text format
	os.Setenv("LOG_FORMAT", "text")
	defer os.Unsetenv("LOG_FORMAT")
	config = loadLoggerConfig()
	if config.Format != "text" {
		t.Errorf("expected text format, got %s", config.Format)
	}

	// Test add source
	os.Setenv("LOG_ADD_SOURCE", "true")
	defer os.Unsetenv("LOG_ADD_SOURCE")
	config = loadLoggerConfig()
	if !config.AddSource {
		t.Error("expected AddSource to be true")
	}

	// Test output path
	os.Setenv("LOG_OUTPUT_PATH", "/tmp/test.log")
	defer os.Unsetenv("LOG_OUTPUT_PATH")
	config = loadLoggerConfig()
	if config.OutputPath != "/tmp/test.log" {
		t.Errorf("expected output path /tmp/test.log, got %s", config.OutputPath)
	}
}
