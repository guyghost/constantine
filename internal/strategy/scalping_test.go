package strategy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/exchanges/dydx"
	"github.com/shopspring/decimal"
)

// MockExchangeForStrategy implements Exchange interface for testing
type MockExchangeForStrategy struct {
	ticker    *exchanges.Ticker
	orderBook *exchanges.OrderBook
}

func (m *MockExchangeForStrategy) Connect(ctx context.Context) error { return nil }
func (m *MockExchangeForStrategy) Disconnect() error                 { return nil }
func (m *MockExchangeForStrategy) IsConnected() bool                 { return true }
func (m *MockExchangeForStrategy) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	return []exchanges.Balance{}, nil
}
func (m *MockExchangeForStrategy) GetPositions(ctx context.Context) ([]exchanges.Position, error) {
	return []exchanges.Position{}, nil
}
func (m *MockExchangeForStrategy) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	return order, nil
}
func (m *MockExchangeForStrategy) CancelOrder(ctx context.Context, orderID string) error { return nil }
func (m *MockExchangeForStrategy) GetOrder(ctx context.Context, orderID string) (*exchanges.Order, error) {
	return nil, nil
}
func (m *MockExchangeForStrategy) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	return []exchanges.Order{}, nil
}
func (m *MockExchangeForStrategy) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]exchanges.Order, error) {
	return []exchanges.Order{}, nil
}
func (m *MockExchangeForStrategy) GetPosition(ctx context.Context, symbol string) (*exchanges.Position, error) {
	return nil, nil
}
func (m *MockExchangeForStrategy) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	return m.ticker, nil
}
func (m *MockExchangeForStrategy) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	return m.orderBook, nil
}
func (m *MockExchangeForStrategy) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	return []exchanges.Candle{}, nil
}
func (m *MockExchangeForStrategy) SubscribeTicker(ctx context.Context, symbol string, callback func(*exchanges.Ticker)) error {
	return nil
}
func (m *MockExchangeForStrategy) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*exchanges.OrderBook)) error {
	return nil
}
func (m *MockExchangeForStrategy) SubscribeTrades(ctx context.Context, symbol string, callback func(*exchanges.Trade)) error {
	return nil
}
func (m *MockExchangeForStrategy) Name() string               { return "mock" }
func (m *MockExchangeForStrategy) SupportedSymbols() []string { return []string{"BTC-USD"} }

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	if config.Symbol != "BTC-USD" {
		t.Errorf("expected symbol BTC-USD, got %s", config.Symbol)
	}

	if config.ShortEMAPeriod != 9 {
		t.Errorf("expected short EMA period 9, got %d", config.ShortEMAPeriod)
	}

	if config.LongEMAPeriod != 21 {
		t.Errorf("expected long EMA period 21, got %d", config.LongEMAPeriod)
	}

	if config.RSIPeriod != 14 {
		t.Errorf("expected RSI period 14, got %d", config.RSIPeriod)
	}

	if config.RSIOversold != 30.0 {
		t.Errorf("expected RSI oversold 30.0, got %f", config.RSIOversold)
	}

	if config.RSIOverbought != 70.0 {
		t.Errorf("expected RSI overbought 70.0, got %f", config.RSIOverbought)
	}
}

func TestNewScalpingStrategy(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}

	strategy := NewScalpingStrategy(config, exchange)

	if strategy == nil {
		t.Fatal("NewScalpingStrategy should not return nil")
	}

	if strategy.config != config {
		t.Error("strategy config should match provided config")
	}

	if strategy.exchange != exchange {
		t.Error("strategy exchange should match provided exchange")
	}

	if strategy.signalGenerator == nil {
		t.Error("signal generator should not be nil")
	}
}

func TestScalpingStrategy_IsRunning(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	// Initially should not be running
	if strategy.IsRunning() {
		t.Error("strategy should not be running initially")
	}

	// Start strategy
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := strategy.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start strategy: %v", err)
	}

	// Should be running now
	if !strategy.IsRunning() {
		t.Error("strategy should be running after start")
	}

	// Stop strategy
	err = strategy.Stop()
	if err != nil {
		t.Fatalf("failed to stop strategy: %v", err)
	}

	// Should not be running anymore
	time.Sleep(100 * time.Millisecond) // Allow some time for cleanup
	if strategy.IsRunning() {
		t.Error("strategy should not be running after stop")
	}
}

func TestScalpingStrategy_Callbacks(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	// Test signal callback
	strategy.SetSignalCallback(func(signal *Signal) {
		// Signal callback set successfully
	})

	// Test error callback
	errorReceived := false
	var receivedError error
	strategy.SetErrorCallback(func(err error) {
		errorReceived = true
		receivedError = err
	})

	// Test position callback
	strategy.SetPositionCallback(func(position *exchanges.Position) {
		// Position callback set successfully
	})

	// Emit test error (this method exists)
	testError := context.Canceled
	strategy.emitError(testError)

	if !errorReceived {
		t.Error("error callback should have been called")
	}
	if receivedError != testError {
		t.Error("received error should match emitted error")
	}

	// For signals and positions, we can't easily test them without running the strategy
	// These would be tested through integration tests or by mocking the internal state
}

func TestScalpingStrategy_Getters(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	// Test GetSignalGenerator
	generator := strategy.GetSignalGenerator()
	if generator == nil {
		t.Error("GetSignalGenerator should not return nil")
	}

	// Test GetLastSignal (initially nil)
	signal := strategy.GetLastSignal()
	if signal != nil {
		t.Error("GetLastSignal should return nil initially")
	}

	// Test GetCurrentPrices (initially empty)
	prices := strategy.GetCurrentPrices()
	if len(prices) != 0 {
		t.Errorf("GetCurrentPrices should return empty slice initially, got %d prices", len(prices))
	}

	// Test GetOrderBook (initially nil)
	orderBook := strategy.GetOrderBook()
	if orderBook != nil {
		t.Error("GetOrderBook should return nil initially")
	}
}

func TestScalpingStrategy_ValidatePrice(t *testing.T) {
	config := DefaultConfig()
	config.MinPrice = decimal.NewFromFloat(10)
	config.MaxPrice = decimal.NewFromFloat(100)
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	tests := []struct {
		name     string
		price    decimal.Decimal
		expected bool
	}{
		{"valid price", decimal.NewFromFloat(50), true},
		{"zero price", decimal.Zero, false},
		{"negative price", decimal.NewFromFloat(-10), false},
		{"below min", decimal.NewFromFloat(5), false},
		{"above max", decimal.NewFromFloat(150), false},
		{"at min", decimal.NewFromFloat(10), true},
		{"at max", decimal.NewFromFloat(100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.validatePrice(tt.price)
			if result != tt.expected {
				t.Errorf("validatePrice(%s) = %v, want %v", tt.price, result, tt.expected)
			}
		})
	}
}

func TestScalpingStrategy_ValidatePriceChange(t *testing.T) {
	config := DefaultConfig()
	config.MaxPriceChangePercent = 5.0
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	tests := []struct {
		name     string
		oldPrice decimal.Decimal
		newPrice decimal.Decimal
		expected bool
	}{
		{"no change", decimal.NewFromFloat(100), decimal.NewFromFloat(100), true},
		{"small change", decimal.NewFromFloat(100), decimal.NewFromFloat(102), true},
		{"large change", decimal.NewFromFloat(100), decimal.NewFromFloat(106), false},
		{"zero old price", decimal.Zero, decimal.NewFromFloat(100), true},
		{"negative change", decimal.NewFromFloat(100), decimal.NewFromFloat(95), true},
		{"large negative change", decimal.NewFromFloat(100), decimal.NewFromFloat(94), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.validatePriceChange(tt.oldPrice, tt.newPrice)
			if result != tt.expected {
				t.Errorf("validatePriceChange(%s, %s) = %v, want %v", tt.oldPrice, tt.newPrice, result, tt.expected)
			}
		})
	}
}

func TestScalpingStrategy_HandleOrderBook(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	orderBook := &exchanges.OrderBook{
		Symbol: "BTC-USD",
		Bids: []exchanges.Level{
			{Price: decimal.NewFromFloat(50000), Amount: decimal.NewFromFloat(1.0)},
		},
		Asks: []exchanges.Level{
			{Price: decimal.NewFromFloat(50100), Amount: decimal.NewFromFloat(1.0)},
		},
	}

	strategy.handleOrderBook(orderBook)

	// Check that order book was stored
	stored := strategy.GetOrderBook()
	if stored == nil {
		t.Fatal("Order book should be stored")
	}
	if stored.Symbol != orderBook.Symbol {
		t.Errorf("Expected symbol %s, got %s", orderBook.Symbol, stored.Symbol)
	}
	if len(stored.Bids) != len(orderBook.Bids) {
		t.Errorf("Expected %d bids, got %d", len(orderBook.Bids), len(stored.Bids))
	}
}

func TestScalpingStrategy_HandleTrade(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	trade := &exchanges.Trade{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
	}

	strategy.handleTrade(trade)

	// Check that volume was added
	prices := strategy.GetCurrentPrices()
	if len(prices) != 0 {
		t.Error("Prices should still be empty")
	}

	// We can't directly check volumes since they're private, but we can check that the function doesn't panic
}

func TestScalpingStrategy_EmitError(t *testing.T) {
	config := DefaultConfig()
	exchange := &MockExchangeForStrategy{}
	strategy := NewScalpingStrategy(config, exchange)

	errorReceived := false
	var receivedError error
	strategy.SetErrorCallback(func(err error) {
		errorReceived = true
		receivedError = err
	})

	testError := fmt.Errorf("test error")
	strategy.emitError(testError)

	if !errorReceived {
		t.Error("Error callback should have been called")
	}
	if receivedError != testError {
		t.Error("Received error should match emitted error")
	}
}

func TestSafeInvoke(t *testing.T) {
	// Test normal function
	called := false
	safeInvoke(func() {
		called = true
	})
	if !called {
		t.Error("Function should have been called")
	}

	// Test panicking function - safeInvoke should catch the panic
	safeInvoke(func() {
		panic("test panic")
	})
	// If we reach here, the panic was caught successfully
}

func TestScalpingStrategy_WithDYDXExchange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dYdX integration test in short mode")
	}

	// Create dYdX client for testnet
	client, err := dydx.NewClientWithMnemonicAndURL(
		"test test test test test test test test test test test junk", // Test mnemonic
		0,
		"https://indexer.v4testnet.dydx.exchange",
		"wss://indexer.v4testnet.dydx.exchange/v4/ws",
	)
	if err != nil {
		t.Skipf("Failed to create dYdX client: %v", err)
	}

	config := DefaultConfig()
	config.Symbol = "BTC-USD"

	strategy := NewScalpingStrategy(config, client)

	if strategy == nil {
		t.Fatal("NewScalpingStrategy should not return nil")
	}

	// Test that strategy can get ticker data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker, err := client.GetTicker(ctx, "BTC-USD")
	if err != nil {
		t.Skipf("Failed to get ticker from dYdX: %v", err)
	}

	if ticker.Symbol != "BTC-USD" {
		t.Errorf("Expected ticker symbol BTC-USD, got %s", ticker.Symbol)
	}

	// Test that strategy can get candles
	candles, err := client.GetCandles(ctx, "BTC-USD", "1MIN", 10)
	if err != nil {
		t.Skipf("Failed to get candles from dYdX: %v", err)
	}

	if len(candles) == 0 {
		t.Error("Should have received candles from dYdX")
	}

	// Test basic strategy functionality
	strategy.SetSignalCallback(func(signal *Signal) {
		t.Logf("Received signal: %s %s at %s", signal.Type, signal.Side, signal.Symbol)
	})

	t.Logf("Successfully integrated with dYdX: ticker=%.2f, candles=%d",
		ticker.Last.InexactFloat64(), len(candles))
}
