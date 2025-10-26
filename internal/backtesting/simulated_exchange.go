package backtesting

import (
	"context"
	"fmt"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// SimulatedExchange simulates an exchange for backtesting
type SimulatedExchange struct {
	data         *HistoricalData
	config       *BacktestConfig
	currentIndex int
}

// NewSimulatedExchange creates a new simulated exchange
func NewSimulatedExchange(data *HistoricalData, config *BacktestConfig) *SimulatedExchange {
	return &SimulatedExchange{
		data:   data,
		config: config,
	}
}

// SetCurrentCandle sets the current candle index
func (s *SimulatedExchange) SetCurrentCandle(index int) {
	s.currentIndex = index
}

// Connect simulates connecting to exchange
func (s *SimulatedExchange) Connect(ctx context.Context) error {
	return nil
}

// Disconnect simulates disconnecting from exchange
func (s *SimulatedExchange) Disconnect() error {
	return nil
}

// IsConnected always returns true for simulated exchange
func (s *SimulatedExchange) IsConnected() bool {
	return true
}

// GetTicker returns the current ticker based on current candle
func (s *SimulatedExchange) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	if s.currentIndex >= len(s.data.Candles) {
		return nil, fmt.Errorf("no more data")
	}

	candle := s.data.Candles[s.currentIndex]

	return &exchanges.Ticker{
		Symbol:    symbol,
		Bid:       candle.Close,
		Ask:       candle.Close,
		Last:      candle.Close,
		Volume24h: candle.Volume,
		Timestamp: candle.Timestamp,
	}, nil
}

// GetOrderBook returns a simulated order book
func (s *SimulatedExchange) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	if s.currentIndex >= len(s.data.Candles) {
		return nil, fmt.Errorf("no more data")
	}

	candle := s.data.Candles[s.currentIndex]
	spread := candle.Close.Mul(decimal.NewFromFloat(0.0001)) // 0.01% spread

	// Create simple order book with current price
	return &exchanges.OrderBook{
		Symbol: symbol,
		Bids: []exchanges.Level{
			{Price: candle.Close.Sub(spread), Amount: decimal.NewFromFloat(10)},
		},
		Asks: []exchanges.Level{
			{Price: candle.Close.Add(spread), Amount: decimal.NewFromFloat(10)},
		},
		Timestamp: candle.Timestamp,
	}, nil
}

// GetCandles returns historical candles up to current index
func (s *SimulatedExchange) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	if s.currentIndex >= len(s.data.Candles) {
		return nil, fmt.Errorf("no more data")
	}

	start := s.currentIndex - limit + 1
	if start < 0 {
		start = 0
	}

	return s.data.Candles[start : s.currentIndex+1], nil
}

// SubscribeTicker not implemented for simulated exchange
func (s *SimulatedExchange) SubscribeTicker(ctx context.Context, symbol string, callback func(*exchanges.Ticker)) error {
	return fmt.Errorf("not implemented for simulated exchange")
}

// SubscribeOrderBook not implemented for simulated exchange
func (s *SimulatedExchange) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*exchanges.OrderBook)) error {
	return fmt.Errorf("not implemented for simulated exchange")
}

// SubscribeTrades not implemented for simulated exchange
func (s *SimulatedExchange) SubscribeTrades(ctx context.Context, symbol string, callback func(*exchanges.Trade)) error {
	return fmt.Errorf("not implemented for simulated exchange")
}

// SubscribeCandles not implemented for simulated exchange
func (s *SimulatedExchange) SubscribeCandles(ctx context.Context, symbol string, interval string, callback func(*exchanges.Candle)) error {
	return fmt.Errorf("not implemented for simulated exchange")
}

// PlaceOrder simulates placing an order (not used in backtesting)
func (s *SimulatedExchange) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	return order, nil
}

// CancelOrder simulates canceling an order
func (s *SimulatedExchange) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}

// GetOrder simulates getting an order
func (s *SimulatedExchange) GetOrder(ctx context.Context, orderID string) (*exchanges.Order, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetOpenOrders simulates getting open orders
func (s *SimulatedExchange) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	return []exchanges.Order{}, nil
}

// GetOrderHistory simulates getting order history
func (s *SimulatedExchange) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]exchanges.Order, error) {
	return []exchanges.Order{}, nil
}

// GetBalance returns simulated balance
func (s *SimulatedExchange) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	return []exchanges.Balance{
		{
			Asset:  "USDC",
			Free:   s.config.InitialCapital,
			Locked: decimal.Zero,
			Total:  s.config.InitialCapital,
		},
	}, nil
}

// GetPositions returns empty positions for simulated exchange
func (s *SimulatedExchange) GetPositions(ctx context.Context) ([]exchanges.Position, error) {
	return []exchanges.Position{}, nil
}

// GetPosition returns nil for simulated exchange
func (s *SimulatedExchange) GetPosition(ctx context.Context, symbol string) (*exchanges.Position, error) {
	return nil, nil
}

// Name returns the exchange name
func (s *SimulatedExchange) Name() string {
	return "SimulatedExchange"
}

// SupportedSymbols returns supported symbols
func (s *SimulatedExchange) SupportedSymbols() []string {
	return []string{s.data.Symbol}
}
