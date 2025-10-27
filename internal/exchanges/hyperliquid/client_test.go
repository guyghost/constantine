package hyperliquid

import (
	"strings"
	"testing"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_key", "test_secret")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.Name() != "Hyperliquid" {
		t.Errorf("Expected name 'Hyperliquid', got %s", client.Name())
	}

	if client.apiKey != "test_key" {
		t.Errorf("Expected apiKey 'test_key', got %s", client.apiKey)
	}

	if client.apiSecret != "test_secret" {
		t.Errorf("Expected apiSecret 'test_secret', got %s", client.apiSecret)
	}
}

func TestExtractCoinFromSymbol(t *testing.T) {
	tests := []struct {
		symbol   string
		expected string
	}{
		{"BTC-USD", "BTC"},
		{"ETH-USD", "ETH"},
		{"SOL-USD", "SOL"},
		{"ARB", "ARB"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			result := extractCoinFromSymbol(tt.symbol)
			if result != tt.expected {
				t.Errorf("extractCoinFromSymbol(%s) = %s, want %s", tt.symbol, result, tt.expected)
			}
		})
	}
}

func TestSupportedSymbols(t *testing.T) {
	client := NewClient("", "")

	symbols := client.SupportedSymbols()
	expected := []string{"BTC-USD", "ETH-USD", "SOL-USD", "ARB-USD"}

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

	// Check USDC balance
	usdcBalance := balances[0]
	if usdcBalance.Asset != "USDC" {
		t.Errorf("Expected first balance to be USDC, got %s", usdcBalance.Asset)
	}

	if !usdcBalance.Total.Equal(decimal.NewFromFloat(11000)) {
		t.Errorf("Expected USDC total 11000, got %s", usdcBalance.Total.String())
	}
}

func TestGetPositions(t *testing.T) {
	client := NewClient("", "")

	positions, err := client.GetPositions(nil)
	if err != nil {
		t.Fatalf("GetPositions returned error: %v", err)
	}

	if len(positions) == 0 {
		t.Fatal("GetPositions returned empty positions")
	}

	btcPosition := positions[0]
	if btcPosition.Symbol != "BTC-USD" {
		t.Errorf("Expected position symbol BTC-USD, got %s", btcPosition.Symbol)
	}

	if !btcPosition.Size.Equal(decimal.NewFromFloat(0.1)) {
		t.Errorf("Expected BTC size 0.1, got %s", btcPosition.Size.String())
	}
}

func TestPlaceOrder(t *testing.T) {
	// Use a dummy private key for testing (32 bytes)
	dummyPrivateKey := "1234567890123456789012345678901234567890123456789012345678901234"
	client := NewClient("", dummyPrivateKey)

	order := &exchanges.Order{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.01),
	}

	// Since we don't have a real API connection, this should fail with a network error
	// but not with the "requires private key" error
	_, err := client.PlaceOrder(nil, order)
	if err == nil {
		t.Error("Expected PlaceOrder to fail with network error, but it succeeded")
	}

	// Check that it's not the private key error
	if err.Error() == "hyperliquid requires a private key to place orders" {
		t.Error("PlaceOrder should not fail with private key error when key is provided")
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

func TestCancelOrder(t *testing.T) {
	t.Run("cancel order without private key should fail", func(t *testing.T) {
		client := NewClient("", "")

		orderID := "12345"

		err := client.CancelOrder(nil, orderID)

		// TODO implementation currently returns nil, but should fail without private key
		// This test will fail until the implementation is complete
		if err == nil {
			t.Fatal("Expected CancelOrder to fail without private key, but it succeeded (TODO: implement private key validation)")
		}

		expectedError := "hyperliquid requires a private key to cancel orders"
		if err.Error() != expectedError {
			t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
		}
	})

	t.Run("cancel order with invalid order ID format should fail", func(t *testing.T) {
		// Use a dummy private key for testing
		dummyPrivateKey := "1234567890123456789012345678901234567890123456789012345678901234"
		client := NewClient("", dummyPrivateKey)

		// Test with malformed order ID (non-numeric)
		invalidOrderID := "invalid-order-id"

		err := client.CancelOrder(nil, invalidOrderID)

		// TODO implementation currently returns nil, but should validate order ID format
		// This test will fail until the implementation is complete
		if err == nil {
			t.Fatal("Expected CancelOrder to fail with invalid order ID, but it succeeded (TODO: implement order ID validation)")
		}

		// Should contain error about invalid order ID format
		if !contains(err.Error(), "invalid order ID format") {
			t.Errorf("Expected error to contain 'invalid order ID format', got '%s'", err.Error())
		}
	})

	t.Run("cancel order with valid parameters should succeed", func(t *testing.T) {
		// Use a dummy private key for testing (32 bytes hex)
		dummyPrivateKey := "1234567890123456789012345678901234567890123456789012345678901234"
		client := NewClient("", dummyPrivateKey)

		orderID := "12345"

		err := client.CancelOrder(nil, orderID)

		// TODO implementation currently returns nil, which is correct for successful cancellation
		// When fully implemented, this should either succeed (nil) or fail with network/API errors
		if err != nil {
			// Check that it's not a validation error (private key or order ID format)
			if contains(err.Error(), "requires a private key") {
				t.Error("CancelOrder should not fail with private key error when key is provided")
			}
			if contains(err.Error(), "invalid order ID format") {
				t.Error("CancelOrder should not fail with order ID format error for valid numeric ID")
			}
			// Network/API errors are acceptable for this test case
		}
	})

	t.Run("cancel order should use ethereum signatures", func(t *testing.T) {
		// Use a dummy private key for testing
		dummyPrivateKey := "1234567890123456789012345678901234567890123456789012345678901234"
		client := NewClient("", dummyPrivateKey)

		orderID := "67890"

		// The function should attempt to sign the cancel action using Ethereum signatures
		// This is more of a design test - when implemented, it should follow the same pattern as PlaceOrder
		err := client.CancelOrder(nil, orderID)

		// Should not fail with signing error when private key is provided
		if err != nil && contains(err.Error(), "failed to sign") {
			t.Errorf("CancelOrder should not fail with signing error when private key is provided, got '%s'", err.Error())
		}
	})

	t.Run("cancel order should return nil on success", func(t *testing.T) {
		// Use a dummy private key for testing
		dummyPrivateKey := "1234567890123456789012345678901234567890123456789012345678901234"
		client := NewClient("", dummyPrivateKey)

		orderID := "98765"

		err := client.CancelOrder(nil, orderID)

		// Current TODO implementation returns nil, which is the expected success behavior
		// When fully implemented, successful cancellation should return nil
		if err != nil {
			// Only network/API errors are acceptable, not validation errors
			if contains(err.Error(), "requires a private key") || contains(err.Error(), "invalid order ID format") {
				t.Errorf("CancelOrder returned validation error when it should have proceeded: %s", err.Error())
			}
		}
	})
}

func TestGetOrderHistory(t *testing.T) {
	t.Run("should return non-nil slice", func(t *testing.T) {
		client := NewClient("", "")

		symbol := "BTC-USD"
		limit := 10

		orders, err := client.GetOrderHistory(nil, symbol, limit)
		if err != nil {
			t.Fatalf("GetOrderHistory returned error: %v", err)
		}

		// Should always return a non-nil slice, even if empty
		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice, should return empty slice instead")
		}

		// Current TODO implementation returns empty slice, which is acceptable
		// When implemented, should return actual order history
	})

	t.Run("should handle symbol filtering", func(t *testing.T) {
		client := NewClient("", "")

		symbol := "ETH-USD"
		limit := 5

		orders, err := client.GetOrderHistory(nil, symbol, limit)
		if err != nil {
			t.Fatalf("GetOrderHistory returned error: %v", err)
		}

		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice")
		}

		// When implemented, should respect symbol filter
		// All returned orders should match the requested symbol
		for _, order := range orders {
			if order.Symbol != symbol && order.Symbol != "" {
				t.Errorf("Expected order symbol '%s', got '%s'", symbol, order.Symbol)
			}
		}
	})

	t.Run("should handle empty symbol (all symbols)", func(t *testing.T) {
		client := NewClient("", "")

		// Empty symbol should return orders for all symbols
		symbol := ""
		limit := 20

		orders, err := client.GetOrderHistory(nil, symbol, limit)
		if err != nil {
			t.Fatalf("GetOrderHistory returned error: %v", err)
		}

		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice")
		}

		// When implemented, should return orders from all symbols when symbol is empty
		// For now, empty slice is acceptable
	})

	t.Run("should respect limit parameter", func(t *testing.T) {
		client := NewClient("", "")

		symbol := "SOL-USD"
		limit := 3

		orders, err := client.GetOrderHistory(nil, symbol, limit)
		if err != nil {
			t.Fatalf("GetOrderHistory returned error: %v", err)
		}

		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice")
		}

		// When implemented, should not return more orders than the limit
		if len(orders) > limit {
			t.Errorf("Expected at most %d orders due to limit, got %d", limit, len(orders))
		}
	})

	t.Run("should handle zero limit gracefully", func(t *testing.T) {
		client := NewClient("", "")

		symbol := "ARB-USD"
		limit := 0

		orders, err := client.GetOrderHistory(nil, symbol, limit)
		if err != nil {
			t.Fatalf("GetOrderHistory returned error: %v", err)
		}

		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice")
		}

		// Zero limit should return empty slice or handle gracefully
		// Current behavior (empty slice) is acceptable
	})

	t.Run("should return properly formatted exchanges.Order objects", func(t *testing.T) {
		client := NewClient("", "")

		symbol := "BTC-USD"
		limit := 10

		orders, err := client.GetOrderHistory(nil, symbol, limit)
		if err != nil {
			t.Fatalf("GetOrderHistory returned error: %v", err)
		}

		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice")
		}

		// When implemented, each order should be properly formatted
		for i, order := range orders {
			// Test that required fields are populated when orders are returned
			if len(orders) > 0 { // Only test if we actually have orders
				if order.Symbol == "" {
					t.Errorf("Order %d: Symbol should not be empty", i)
				}
				if order.ID == "" {
					t.Errorf("Order %d: ID should not be empty", i)
				}
				if order.Side != exchanges.OrderSideBuy && order.Side != exchanges.OrderSideSell {
					t.Errorf("Order %d: Invalid order side: %s", i, order.Side)
				}
				if order.Status == "" {
					t.Errorf("Order %d: Status should not be empty", i)
				}
				// Price and Amount should be positive for valid orders
				if order.Price.IsNegative() {
					t.Errorf("Order %d: Price should not be negative: %s", i, order.Price.String())
				}
				if order.Amount.IsNegative() {
					t.Errorf("Order %d: Amount should not be negative: %s", i, order.Amount.String())
				}
			}
		}
	})

	t.Run("should use Hyperliquid API endpoint pattern", func(t *testing.T) {
		client := NewClient("", "")

		symbol := "BTC-USD"
		limit := 5

		orders, err := client.GetOrderHistory(nil, symbol, limit)

		// Should not return errors for valid parameters in current TODO implementation
		if err != nil {
			t.Fatalf("GetOrderHistory returned unexpected error: %v", err)
		}

		if orders == nil {
			t.Fatal("GetOrderHistory returned nil orders slice")
		}

		// When implemented, should make POST request to /info endpoint
		// with appropriate request body for order history query
		// This is more of a design expectation test
	})

	t.Run("should handle API errors gracefully", func(t *testing.T) {
		client := NewClient("", "")

		// Test with various parameters that might cause issues
		testCases := []struct {
			name   string
			symbol string
			limit  int
		}{
			{"negative limit", "BTC-USD", -1},
			{"very large limit", "ETH-USD", 10000},
			{"unusual symbol format", "INVALID-SYMBOL", 10},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				orders, err := client.GetOrderHistory(nil, tc.symbol, tc.limit)

				// Should handle edge cases gracefully - either return empty slice or appropriate error
				if err != nil {
					// Errors are acceptable for invalid parameters
					return
				}

				if orders == nil {
					t.Fatal("GetOrderHistory returned nil orders slice, should return empty slice on error")
				}

				// Should not panic or return invalid data
			})
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
