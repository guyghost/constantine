package coinbase

import (
	"testing"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_key", "test_secret")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.Name() != "Coinbase" {
		t.Errorf("Expected name 'Coinbase', got %s", client.Name())
	}

	if client.apiKey != "test_key" {
		t.Errorf("Expected apiKey 'test_key', got %s", client.apiKey)
	}

	if client.apiSecret != "test_secret" {
		t.Errorf("Expected apiSecret 'test_secret', got %s", client.apiSecret)
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
	client := NewClient("", "")

	balances, err := client.GetBalance(nil)
	if err != nil {
		t.Fatalf("GetBalance returned error: %v", err)
	}

	if len(balances) == 0 {
		t.Fatal("GetBalance returned empty balances")
	}

	// Check USD balance
	usdBalance := balances[0]
	if usdBalance.Asset != "USD" {
		t.Errorf("Expected first balance to be USD, got %s", usdBalance.Asset)
	}

	if !usdBalance.Total.Equal(decimal.NewFromFloat(11000)) {
		t.Errorf("Expected USD total 11000, got %s", usdBalance.Total.String())
	}
}

func TestGetPositions(t *testing.T) {
	client := NewClient("", "")

	positions, err := client.GetPositions(nil)
	if err != nil {
		t.Fatalf("GetPositions returned error: %v", err)
	}

	// Should have at least one position (BTC)
	if len(positions) == 0 {
		t.Fatal("GetPositions returned empty positions")
	}

	btcPosition := positions[0]
	if btcPosition.Symbol != "BTC-USD" {
		t.Errorf("Expected position symbol BTC-USD, got %s", btcPosition.Symbol)
	}

	if !btcPosition.Size.Equal(decimal.NewFromFloat(0.6)) {
		t.Errorf("Expected BTC size 0.6, got %s", btcPosition.Size.String())
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
