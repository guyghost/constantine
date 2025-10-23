package coinbase

import (
	"context"
	"testing"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_key", "test_private_key_pem")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.Name() != "Coinbase" {
		t.Errorf("Expected name 'Coinbase', got %s", client.Name())
	}

	if client.apiKey != "test_key" {
		t.Errorf("Expected apiKey 'test_key', got %s", client.apiKey)
	}

	if client.privateKeyPEM != "test_private_key_pem" {
		t.Errorf("Expected privateKeyPEM 'test_private_key_pem', got %s", client.privateKeyPEM)
	}
}

func TestIntervalToGranularity(t *testing.T) {
	tests := []struct {
		interval string
		expected string
	}{
		{"1m", "60"},
		{"5m", "300"},
		{"15m", "900"},
		{"1h", "3600"},
		{"6h", "21600"},
		{"1d", "86400"},
		{"unknown", "3600"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			result := intervalToGranularity(tt.interval)
			if result != tt.expected {
				t.Errorf("intervalToGranularity(%s) = %s, want %s", tt.interval, result, tt.expected)
			}
		})
	}
}

func TestSupportedSymbols(t *testing.T) {
	client := NewClient("", "")

	symbols := client.SupportedSymbols()
	expected := []string{"BTC-USD", "ETH-USD", "SOL-USD", "LINK-USD"}

	if len(symbols) != len(expected) {
		t.Fatalf("Expected %d symbols, got %d", len(expected), len(symbols))
	}

	for i, symbol := range symbols {
		if symbol != expected[i] {
			t.Errorf("Expected symbol %s at index %d, got %s", expected[i], i, symbol)
		}
	}
}

func TestGetBalance(t *testing.T) {
	client := NewClient("", "") // No API keys for testing

	// Test with valid context but no API keys - should fail with auth error
	ctx := context.Background()
	balances, err := client.GetBalance(ctx)

	// Should fail because no API keys provided
	if err == nil {
		t.Error("Expected GetBalance to fail without API keys")
	}

	// If somehow it succeeds (mock data), check structure
	if err == nil && len(balances) > 0 {
		for _, balance := range balances {
			if balance.Asset == "" {
				t.Error("Balance should have non-empty asset")
			}
			if balance.Total.LessThan(decimal.Zero) {
				t.Error("Balance total should not be negative")
			}
			if balance.UpdatedAt.IsZero() {
				t.Error("Balance should have UpdatedAt timestamp")
			}
		}
	}
}

func TestGetPositions(t *testing.T) {
	client := NewClient("", "") // No API keys for testing

	// Test with valid context but no API keys - should fail with auth error
	ctx := context.Background()
	positions, err := client.GetPositions(ctx)

	// Should fail because no API keys provided
	if err == nil {
		t.Error("Expected GetPositions to fail without API keys")
	}

	// If somehow it succeeds, check structure
	if err == nil && len(positions) > 0 {
		for _, position := range positions {
			if position.Symbol == "" {
				t.Error("Position should have non-empty symbol")
			}
			if position.Size.LessThanOrEqual(decimal.Zero) {
				t.Error("Position size should be positive")
			}
		}
	}
}

func TestPlaceOrder(t *testing.T) {
	client := NewClient("", "")

	order := &exchanges.Order{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.01),
	}

	placedOrder, err := client.PlaceOrder(nil, order)
	if err != nil {
		t.Fatalf("PlaceOrder returned error: %v", err)
	}

	if placedOrder.ID == "" {
		t.Error("Placed order should have an ID")
	}

	if placedOrder.Status != exchanges.OrderStatusOpen {
		t.Errorf("Expected order status Open, got %s", placedOrder.Status)
	}

	if placedOrder.CreatedAt.IsZero() {
		t.Error("Placed order should have CreatedAt timestamp")
	}
}

func TestGetOpenOrders(t *testing.T) {
	client := NewClient("", "")

	orders, err := client.GetOpenOrders(nil, "BTC-USD")
	if err != nil {
		t.Fatalf("GetOpenOrders returned error: %v", err)
	}

	if len(orders) == 0 {
		t.Fatal("GetOpenOrders returned empty orders")
	}

	order := orders[0]
	if order.Symbol != "BTC-USD" {
		t.Errorf("Expected order symbol BTC-USD, got %s", order.Symbol)
	}

	if order.Status != exchanges.OrderStatusOpen {
		t.Errorf("Expected order status Open, got %s", order.Status)
	}
}
