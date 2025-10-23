// Package testutils provides shared utilities for testing
package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// TestExchange is a test implementation of the Exchange interface
type TestExchange struct {
	NameValue        string
	ConnectedValue   bool
	BalancesValue    []exchanges.Balance
	PositionsValue   []exchanges.Position
	OrdersValue      []exchanges.Order
	TickerValue      *exchanges.Ticker
	OrderBookValue   *exchanges.OrderBook
	CandlesValue     []exchanges.Candle
	ConnectError     error
	BalanceError     error
	PositionError    error
	OrderError       error
	PlaceOrderError  error
	CancelOrderError error
}

func NewTestExchange(name string) *TestExchange {
	return &TestExchange{
		NameValue:      name,
		ConnectedValue: true,
		BalancesValue: []exchanges.Balance{
			{
				Asset:  "USD",
				Free:   decimal.NewFromFloat(10000),
				Locked: decimal.NewFromFloat(1000),
				Total:  decimal.NewFromFloat(11000),
			},
		},
		PositionsValue: []exchanges.Position{
			{
				Symbol:           "BTC-USD",
				Side:             exchanges.OrderSideBuy,
				Size:             decimal.NewFromFloat(0.5),
				EntryPrice:       decimal.NewFromFloat(50000),
				MarkPrice:        decimal.NewFromFloat(51000),
				UnrealizedPnL:    decimal.NewFromFloat(500),
				RealizedPnL:      decimal.Zero,
				LiquidationPrice: decimal.NewFromFloat(45000),
			},
		},
		OrdersValue: []exchanges.Order{
			{
				ID:        "test-order-1",
				Symbol:    "BTC-USD",
				Side:      exchanges.OrderSideBuy,
				Type:      exchanges.OrderTypeLimit,
				Price:     decimal.NewFromFloat(50000),
				Amount:    decimal.NewFromFloat(0.1),
				Status:    exchanges.OrderStatusOpen,
				CreatedAt: time.Now(),
			},
		},
		TickerValue: &exchanges.Ticker{
			Symbol:    "BTC-USD",
			Bid:       decimal.NewFromFloat(49900),
			Ask:       decimal.NewFromFloat(50100),
			Last:      decimal.NewFromFloat(50000),
			Volume24h: decimal.NewFromFloat(1000),
			Timestamp: time.Now(),
		},
		OrderBookValue: &exchanges.OrderBook{
			Symbol: "BTC-USD",
			Bids: []exchanges.Level{
				{Price: decimal.NewFromFloat(49900), Amount: decimal.NewFromFloat(1.0)},
				{Price: decimal.NewFromFloat(49800), Amount: decimal.NewFromFloat(2.0)},
			},
			Asks: []exchanges.Level{
				{Price: decimal.NewFromFloat(50100), Amount: decimal.NewFromFloat(1.0)},
				{Price: decimal.NewFromFloat(50200), Amount: decimal.NewFromFloat(2.0)},
			},
			Timestamp: time.Now(),
		},
		CandlesValue: []exchanges.Candle{
			{
				Symbol:    "BTC-USD",
				Timestamp: time.Now().Add(-time.Hour),
				Open:      decimal.NewFromFloat(49000),
				High:      decimal.NewFromFloat(51000),
				Low:       decimal.NewFromFloat(48500),
				Close:     decimal.NewFromFloat(50000),
				Volume:    decimal.NewFromFloat(100),
			},
		},
	}
}

func (t *TestExchange) Connect(ctx context.Context) error {
	if t.ConnectError != nil {
		return t.ConnectError
	}
	t.ConnectedValue = true
	return nil
}

func (t *TestExchange) Disconnect() error {
	t.ConnectedValue = false
	return nil
}

func (t *TestExchange) IsConnected() bool {
	return t.ConnectedValue
}

func (t *TestExchange) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	if t.BalanceError != nil {
		return nil, t.BalanceError
	}
	return t.BalancesValue, nil
}

func (t *TestExchange) GetPositions(ctx context.Context) ([]exchanges.Position, error) {
	if t.PositionError != nil {
		return nil, t.PositionError
	}
	return t.PositionsValue, nil
}

func (t *TestExchange) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	if t.PlaceOrderError != nil {
		return nil, t.PlaceOrderError
	}
	order.ID = "placed-" + t.NameValue + "-" + order.ID
	order.Status = exchanges.OrderStatusOpen
	return order, nil
}

func (t *TestExchange) CancelOrder(ctx context.Context, orderID string) error {
	return t.CancelOrderError
}

func (t *TestExchange) GetOrder(ctx context.Context, orderID string) (*exchanges.Order, error) {
	for _, order := range t.OrdersValue {
		if order.ID == orderID {
			return &order, nil
		}
	}
	return nil, exchanges.ErrOrderNotFound
}

func (t *TestExchange) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	if t.OrderError != nil {
		return nil, t.OrderError
	}
	return t.OrdersValue, nil
}

func (t *TestExchange) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]exchanges.Order, error) {
	return t.OrdersValue, nil
}

func (t *TestExchange) GetPosition(ctx context.Context, symbol string) (*exchanges.Position, error) {
	for _, pos := range t.PositionsValue {
		if pos.Symbol == symbol {
			return &pos, nil
		}
	}
	return nil, exchanges.ErrPositionNotFound
}

func (t *TestExchange) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	return t.TickerValue, nil
}

func (t *TestExchange) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	return t.OrderBookValue, nil
}

func (t *TestExchange) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	return t.CandlesValue, nil
}

func (t *TestExchange) SubscribeTicker(ctx context.Context, symbol string, callback func(*exchanges.Ticker)) error {
	return nil
}

func (t *TestExchange) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*exchanges.OrderBook)) error {
	return nil
}

func (t *TestExchange) SubscribeTrades(ctx context.Context, symbol string, callback func(*exchanges.Trade)) error {
	return nil
}

func (t *TestExchange) Name() string {
	return t.NameValue
}

func (t *TestExchange) SupportedSymbols() []string {
	return []string{"BTC-USD", "ETH-USD"}
}

// AssertEqual is a helper function for asserting equality in tests
func AssertEqual(t *testing.T, expected, actual any, message string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertTrue is a helper function for asserting boolean true
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true, got false", message)
	}
}

// AssertFalse is a helper function for asserting boolean false
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false, got true", message)
	}
}

// AssertNil is a helper function for asserting nil values
func AssertNil(t *testing.T, value any, message string) {
	t.Helper()
	if value != nil {
		t.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertNotNil is a helper function for asserting non-nil values
func AssertNotNil(t *testing.T, value any, message string) {
	t.Helper()
	if value == nil {
		t.Errorf("%s: expected non-nil value, got nil", message)
	}
}

// AssertNoError is a helper function for asserting no error
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", message, err)
	}
}

// AssertError is a helper function for asserting an error
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error, got nil", message)
	}
}

// CreateTestContext creates a context for testing with timeout
func CreateTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// SampleCandles returns a slice of sample candle data for testing
func SampleCandles() []exchanges.Candle {
	baseTime := time.Now().Add(-24 * time.Hour)
	candles := make([]exchanges.Candle, 100)

	for i := 0; i < 100; i++ {
		price := 50000 + float64(i)*100 // Trending up
		candles[i] = exchanges.Candle{
			Symbol:    "BTC-USD",
			Timestamp: baseTime.Add(time.Duration(i) * time.Hour),
			Open:      decimal.NewFromFloat(price - 50),
			High:      decimal.NewFromFloat(price + 100),
			Low:       decimal.NewFromFloat(price - 100),
			Close:     decimal.NewFromFloat(price),
			Volume:    decimal.NewFromFloat(100 + float64(i)),
		}
	}

	return candles
}
