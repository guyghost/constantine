//go:build integration
// +build integration

package dydx

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestIntegration_GetTicker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mnemonic := os.Getenv("DYDX_MNEMONIC")
	subAccountNumber := 0

	if mnemonic == "" {
		t.Skip("DYDX_MNEMONIC not set")
	}

	// Use testnet for integration tests
	baseURL := os.Getenv("DYDX_BASE_URL")
	wsURL := os.Getenv("DYDX_WS_URL")
	if baseURL == "" {
		baseURL = "https://indexer.v4testnet.dydx.exchange"
	}
	if wsURL == "" {
		wsURL = "wss://indexer.v4testnet.dydx.exchange/v4/ws"
	}

	client, err := NewClientWithMnemonicAndURL(mnemonic, subAccountNumber, baseURL, wsURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

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

	mnemonic := os.Getenv("DYDX_MNEMONIC")
	subAccountNumber := 0

	if mnemonic == "" {
		t.Skip("DYDX_MNEMONIC not set")
	}

	baseURL := os.Getenv("DYDX_BASE_URL")
	wsURL := os.Getenv("DYDX_WS_URL")
	if baseURL == "" {
		baseURL = "https://indexer.v4testnet.dydx.exchange"
	}
	if wsURL == "" {
		wsURL = "wss://indexer.v4testnet.dydx.exchange/v4/ws"
	}

	client, err := NewClientWithMnemonicAndURL(mnemonic, subAccountNumber, baseURL, wsURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

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

	mnemonic := os.Getenv("DYDX_MNEMONIC")
	subAccountNumber := 0

	if mnemonic == "" {
		t.Skip("DYDX_MNEMONIC not set")
	}

	baseURL := os.Getenv("DYDX_BASE_URL")
	wsURL := os.Getenv("DYDX_WS_URL")
	if baseURL == "" {
		baseURL = "https://indexer.v4testnet.dydx.exchange"
	}
	if wsURL == "" {
		wsURL = "wss://indexer.v4testnet.dydx.exchange/v4/ws"
	}

	client, err := NewClientWithMnemonicAndURL(mnemonic, subAccountNumber, baseURL, wsURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	candles, err := client.GetCandles(ctx, "BTC-USD", "1m", 10)
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

	mnemonic := os.Getenv("DYDX_MNEMONIC")
	subAccountNumber := 0

	if mnemonic == "" {
		t.Skip("DYDX_MNEMONIC not set")
	}

	baseURL := os.Getenv("DYDX_BASE_URL")
	wsURL := os.Getenv("DYDX_WS_URL")
	if baseURL == "" {
		baseURL = "https://indexer.v4testnet.dydx.exchange"
	}
	if wsURL == "" {
		wsURL = "wss://indexer.v4testnet.dydx.exchange/v4/ws"
	}

	client, err := NewClientWithMnemonicAndURL(mnemonic, subAccountNumber, baseURL, wsURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	balances, err := client.GetBalance(ctx)
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}

	// Testnet might have empty balances, so just check that the call succeeds
	t.Logf("Retrieved %d balances", len(balances))
}

func TestIntegration_Connect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mnemonic := os.Getenv("DYDX_MNEMONIC")
	subAccountNumber := 0

	if mnemonic == "" {
		t.Skip("DYDX_MNEMONIC not set")
	}

	baseURL := os.Getenv("DYDX_BASE_URL")
	wsURL := os.Getenv("DYDX_WS_URL")
	if baseURL == "" {
		baseURL = "https://indexer.v4testnet.dydx.exchange"
	}
	if wsURL == "" {
		wsURL = "wss://indexer.v4testnet.dydx.exchange/v4/ws"
	}

	client, err := NewClientWithMnemonicAndURL(mnemonic, subAccountNumber, baseURL, wsURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
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
