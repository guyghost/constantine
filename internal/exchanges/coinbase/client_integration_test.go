//go:build integration
// +build integration

package coinbase

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
)

func TestIntegration_GetTicker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("COINBASE_API_KEY")
	privateKeyPEM := os.Getenv("COINBASE_PRIVATE_KEY_PEM")

	if apiKey == "" || privateKeyPEM == "" {
		t.Skip("COINBASE_API_KEY and COINBASE_PRIVATE_KEY_PEM not set")
	}

	// Use sandbox for integration tests
	baseURL := os.Getenv("COINBASE_BASE_URL")
	wsURL := os.Getenv("COINBASE_WS_URL")
	if baseURL == "" {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
	}
	if wsURL == "" {
		wsURL = "wss://ws-direct.sandbox.exchange.coinbase.com"
	}

	client := NewClientWithURL(apiKey, privateKeyPEM, baseURL, wsURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker, err := client.GetTicker(ctx, "BTC-USD")
	if err != nil {
		t.Fatalf("GetTicker failed: %v", err)
	}

	if ticker.Symbol != "BTC-USD" {
		t.Errorf("Expected symbol BTC-USD, got %s", ticker.Symbol)
	}

	if ticker.Last.IsZero() {
		t.Error("Expected non-zero last price")
	}

	if ticker.Bid.IsZero() {
		t.Error("Expected non-zero bid price")
	}

	if ticker.Ask.IsZero() {
		t.Error("Expected non-zero ask price")
	}
}

func TestIntegration_GetOrderBook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("COINBASE_API_KEY")
	privateKeyPEM := os.Getenv("COINBASE_PRIVATE_KEY_PEM")

	if apiKey == "" || privateKeyPEM == "" {
		t.Skip("COINBASE_API_KEY and COINBASE_PRIVATE_KEY_PEM not set")
	}

	baseURL := os.Getenv("COINBASE_BASE_URL")
	wsURL := os.Getenv("COINBASE_WS_URL")
	if baseURL == "" {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
	}
	if wsURL == "" {
		wsURL = "wss://ws-direct.sandbox.exchange.coinbase.com"
	}

	client := NewClientWithURL(apiKey, privateKeyPEM, baseURL, wsURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderBook, err := client.GetOrderBook(ctx, "BTC-USD", 10)
	if err != nil {
		t.Fatalf("GetOrderBook failed: %v", err)
	}

	if orderBook.Symbol != "BTC-USD" {
		t.Errorf("Expected symbol BTC-USD, got %s", orderBook.Symbol)
	}

	if len(orderBook.Bids) == 0 {
		t.Error("Expected non-empty bids")
	}

	if len(orderBook.Asks) == 0 {
		t.Error("Expected non-empty asks")
	}
}

func TestIntegration_GetCandles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("COINBASE_API_KEY")
	privateKeyPEM := os.Getenv("COINBASE_PRIVATE_KEY_PEM")

	if apiKey == "" || privateKeyPEM == "" {
		t.Skip("COINBASE_API_KEY and COINBASE_PRIVATE_KEY_PEM not set")
	}

	baseURL := os.Getenv("COINBASE_BASE_URL")
	wsURL := os.Getenv("COINBASE_WS_URL")
	if baseURL == "" {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
	}
	if wsURL == "" {
		wsURL = "wss://ws-direct.sandbox.exchange.coinbase.com"
	}

	client := NewClientWithURL(apiKey, privateKeyPEM, baseURL, wsURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	candles, err := client.GetCandles(ctx, "BTC-USD", "1h", 10)
	if err != nil {
		t.Fatalf("GetCandles failed: %v", err)
	}

	if len(candles) == 0 {
		t.Error("Expected non-empty candles")
	}

	for i, candle := range candles {
		if candle.Symbol != "BTC-USD" {
			t.Errorf("Candle %d: expected symbol BTC-USD, got %s", i, candle.Symbol)
		}
		if candle.Open.IsZero() {
			t.Errorf("Candle %d: expected non-zero open price", i)
		}
		if candle.Close.IsZero() {
			t.Errorf("Candle %d: expected non-zero close price", i)
		}
	}
}

func TestIntegration_GetBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("COINBASE_API_KEY")
	privateKeyPEM := os.Getenv("COINBASE_PRIVATE_KEY_PEM")

	if apiKey == "" || privateKeyPEM == "" {
		t.Skip("COINBASE_API_KEY and COINBASE_PRIVATE_KEY_PEM not set")
	}

	baseURL := os.Getenv("COINBASE_BASE_URL")
	wsURL := os.Getenv("COINBASE_WS_URL")
	if baseURL == "" {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
	}
	if wsURL == "" {
		wsURL = "wss://ws-direct.sandbox.exchange.coinbase.com"
	}

	client := NewClientWithURL(apiKey, privateKeyPEM, baseURL, wsURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	balances, err := client.GetBalance(ctx)
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}

	// Sandbox might have empty balances, so just check that the call succeeds
	t.Logf("Retrieved %d balances", len(balances))
}

func TestIntegration_Connect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	apiKey := os.Getenv("COINBASE_API_KEY")
	privateKeyPEM := os.Getenv("COINBASE_PRIVATE_KEY_PEM")

	if apiKey == "" || privateKeyPEM == "" {
		t.Skip("COINBASE_API_KEY and COINBASE_PRIVATE_KEY_PEM not set")
	}

	baseURL := os.Getenv("COINBASE_BASE_URL")
	wsURL := os.Getenv("COINBASE_WS_URL")
	if baseURL == "" {
		baseURL = "https://api-public.sandbox.exchange.coinbase.com"
	}
	if wsURL == "" {
		wsURL = "wss://ws-direct.sandbox.exchange.coinbase.com"
	}

	client := NewClientWithURL(apiKey, privateKeyPEM, baseURL, wsURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}

	client.Disconnect()
	if client.IsConnected() {
		t.Error("Expected client to be disconnected")
	}
}
