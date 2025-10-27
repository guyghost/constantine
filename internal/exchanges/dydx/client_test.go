package dydx

import (
	"context"
	"testing"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// TestClient_IsConnected tests connection status checking
func TestClient_IsConnected(t *testing.T) {
	tests := []struct {
		name      string
		connected bool
		expected  bool
	}{
		{
			name:      "connected",
			connected: true,
			expected:  true,
		},
		{
			name:      "disconnected",
			connected: false,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				connected: tt.connected,
			}

			result := client.IsConnected()
			if result != tt.expected {
				t.Errorf("Expected IsConnected() = %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestClient_Disconnect tests disconnection
func TestClient_Disconnect(t *testing.T) {
	client := &Client{
		connected: true,
		ws:        &WebSocketClient{url: "wss://test.com"},
	}

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("Expected connected to be false after disconnect")
	}
}

// TestClient_Name tests exchange name
func TestClient_Name(t *testing.T) {
	client := &Client{}
	name := client.Name()

	if name != "dYdX" {
		t.Errorf("Expected name 'dYdX', got %s", name)
	}
}

// TestClient_SupportedSymbols tests supported symbols list
func TestClient_SupportedSymbols(t *testing.T) {
	client := &Client{}
	symbols := client.SupportedSymbols()

	if len(symbols) == 0 {
		t.Error("Expected at least one supported symbol")
	}

	expectedSymbols := map[string]bool{
		"BTC-USD":  true,
		"ETH-USD":  true,
		"SOL-USD":  true,
		"AVAX-USD": true,
	}

	for _, symbol := range symbols {
		if !expectedSymbols[symbol] {
			t.Errorf("Unexpected symbol: %s", symbol)
		}
	}
}

// TestClient_IsAuthenticated tests authentication status
func TestClient_IsAuthenticated(t *testing.T) {
	tests := []struct {
		name         string
		wallet       *Wallet
		signer       *Signer
		expectedAuth bool
	}{
		{
			name:         "not authenticated - no wallet",
			wallet:       nil,
			signer:       nil,
			expectedAuth: false,
		},
		{
			name:         "not authenticated - wallet only",
			wallet:       &Wallet{Address: "cosmos1test"},
			signer:       nil,
			expectedAuth: false,
		},
		{
			name:         "not authenticated - signer only",
			wallet:       nil,
			signer:       &Signer{},
			expectedAuth: false,
		},
		{
			name:         "authenticated - wallet and signer",
			wallet:       &Wallet{Address: "cosmos1test"},
			signer:       &Signer{},
			expectedAuth: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				wallet: tt.wallet,
				signer: tt.signer,
			}

			result := client.IsAuthenticated()
			if result != tt.expectedAuth {
				t.Errorf("Expected IsAuthenticated() = %v, got %v", tt.expectedAuth, result)
			}
		})
	}
}

// TestClient_GetWalletAddress tests wallet address retrieval
func TestClient_GetWalletAddress(t *testing.T) {
	tests := []struct {
		name     string
		wallet   *Wallet
		expected string
	}{
		{
			name:     "no wallet",
			wallet:   nil,
			expected: "",
		},
		{
			name:     "with wallet",
			wallet:   &Wallet{Address: "cosmos1abc123"},
			expected: "cosmos1abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				wallet: tt.wallet,
			}

			result := client.GetWalletAddress()
			if result != tt.expected {
				t.Errorf("Expected GetWalletAddress() = %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestClient_GetOrder_NotImplemented tests that GetOrder returns not implemented error
func TestClient_GetOrder_NotImplemented(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	_, err := client.GetOrder(ctx, "order-123")
	if err == nil {
		t.Error("Expected error for GetOrder (not implemented)")
	}

	if !contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}
}

// TestClient_GetOrderHistory_NotImplemented tests that GetOrderHistory returns not implemented error
func TestClient_GetOrderHistory_NotImplemented(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	_, err := client.GetOrderHistory(ctx, "BTC-USD", 10)
	if err == nil {
		t.Error("Expected error for GetOrderHistory (not implemented)")
	}

	if !contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}
}

// TestClient_GetPosition_NotImplemented tests that GetPosition returns not implemented error
func TestClient_GetPosition_NotImplemented(t *testing.T) {
	client := &Client{}
	ctx := context.Background()

	_, err := client.GetPosition(ctx, "BTC-USD")
	if err == nil {
		t.Error("Expected error for GetPosition (not implemented)")
	}

	if !contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}
}

// TestClient_PlaceOrder_NoPythonClient tests PlaceOrder without Python client
func TestClient_PlaceOrder_NoPythonClient(t *testing.T) {
	client := &Client{
		pythonClient: nil,
	}

	ctx := context.Background()
	order := &exchanges.Order{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromInt(43000),
		Amount: decimal.NewFromFloat(0.1),
	}

	_, err := client.PlaceOrder(ctx, order)
	if err == nil {
		t.Error("Expected error when Python client not initialized")
	}

	if !contains(err.Error(), "Python client not initialized") {
		t.Errorf("Expected 'Python client not initialized' error, got: %v", err)
	}
}

// TestClient_CancelOrder_NoPythonClient tests CancelOrder without Python client
func TestClient_CancelOrder_NoPythonClient(t *testing.T) {
	client := &Client{
		pythonClient: nil,
	}

	ctx := context.Background()
	err := client.CancelOrder(ctx, "order-123")
	if err == nil {
		t.Error("Expected error when Python client not initialized")
	}

	if !contains(err.Error(), "Python client not initialized") {
		t.Errorf("Expected 'Python client not initialized' error, got: %v", err)
	}
}

// TestClient_GetBalance_NoWallet tests GetBalance without wallet initialization
func TestClient_GetBalance_NoWallet(t *testing.T) {
	client := &Client{
		wallet: nil,
	}

	ctx := context.Background()
	_, err := client.GetBalance(ctx)
	if err == nil {
		t.Error("Expected error when wallet not initialized")
	}

	if !contains(err.Error(), "wallet not initialized") {
		t.Errorf("Expected 'wallet not initialized' error, got: %v", err)
	}
}

// TestClient_GetPositions_NoWallet tests GetPositions without wallet initialization
func TestClient_GetPositions_NoWallet(t *testing.T) {
	client := &Client{
		wallet: nil,
	}

	ctx := context.Background()
	_, err := client.GetPositions(ctx)
	if err == nil {
		t.Error("Expected error when wallet not initialized")
	}

	if !contains(err.Error(), "wallet not initialized") {
		t.Errorf("Expected 'wallet not initialized' error, got: %v", err)
	}
}

// TestClient_GetOpenOrders_NoWallet tests GetOpenOrders without wallet initialization
func TestClient_GetOpenOrders_NoWallet(t *testing.T) {
	client := &Client{
		wallet: nil,
	}

	ctx := context.Background()
	_, err := client.GetOpenOrders(ctx, "BTC-USD")
	if err == nil {
		t.Error("Expected error when wallet not initialized")
	}

	if !contains(err.Error(), "wallet not initialized") {
		t.Errorf("Expected 'wallet not initialized' error, got: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	if len(str) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
