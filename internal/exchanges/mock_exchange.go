package exchanges

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// MockExchange implements the Exchange interface for testing
type MockExchange struct {
	name          string
	connected     bool
	balances      []Balance
	positions     []Position
	orders        []Order
	connectError  error
	balanceError  error
	positionError error
	orderError    error
}

func NewMockExchange(name string) *MockExchange {
	return &MockExchange{
		name:      name,
		connected: true,
		balances: []Balance{
			{
				Asset:  "USD",
				Free:   decimal.NewFromFloat(1000),
				Locked: decimal.NewFromFloat(100),
				Total:  decimal.NewFromFloat(1100),
			},
		},
		positions: []Position{
			{
				Symbol:           "BTC-USD",
				Side:             OrderSideBuy,
				Size:             decimal.NewFromFloat(0.1),
				EntryPrice:       decimal.NewFromFloat(50000),
				MarkPrice:        decimal.NewFromFloat(51000),
				UnrealizedPnL:    decimal.NewFromFloat(100),
				RealizedPnL:      decimal.Zero,
				LiquidationPrice: decimal.NewFromFloat(45000),
			},
		},
		orders: []Order{
			{
				ID:        "order1",
				Symbol:    "BTC-USD",
				Side:      OrderSideBuy,
				Type:      OrderTypeLimit,
				Price:     decimal.NewFromFloat(50000),
				Amount:    decimal.NewFromFloat(0.1),
				Status:    OrderStatusOpen,
				CreatedAt: time.Now(),
			},
		},
	}
}

func (m *MockExchange) Connect(ctx context.Context) error {
	if m.connectError != nil {
		return m.connectError
	}
	m.connected = true
	return nil
}

func (m *MockExchange) Disconnect() error {
	m.connected = false
	return nil
}

func (m *MockExchange) IsConnected() bool {
	return m.connected
}

func (m *MockExchange) GetBalance(ctx context.Context) ([]Balance, error) {
	if m.balanceError != nil {
		return nil, m.balanceError
	}
	return m.balances, nil
}

func (m *MockExchange) GetPositions(ctx context.Context) ([]Position, error) {
	if m.positionError != nil {
		return nil, m.positionError
	}
	return m.positions, nil
}

func (m *MockExchange) PlaceOrder(ctx context.Context, order *Order) (*Order, error) {
	order.ID = "new_order_" + m.name
	order.Status = OrderStatusOpen
	return order, nil
}

func (m *MockExchange) CancelOrder(ctx context.Context, orderID string) error {
	return nil
}

func (m *MockExchange) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	for _, order := range m.orders {
		if order.ID == orderID {
			return &order, nil
		}
	}
	return nil, errors.New("order not found")
}

func (m *MockExchange) GetOpenOrders(ctx context.Context, symbol string) ([]Order, error) {
	if m.orderError != nil {
		return nil, m.orderError
	}
	return m.orders, nil
}

func (m *MockExchange) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]Order, error) {
	return m.orders, nil
}

func (m *MockExchange) GetPosition(ctx context.Context, symbol string) (*Position, error) {
	for _, pos := range m.positions {
		if pos.Symbol == symbol {
			return &pos, nil
		}
	}
	return nil, errors.New("position not found")
}

func (m *MockExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	return &Ticker{
		Symbol: symbol,
		Bid:    decimal.NewFromFloat(50000),
		Ask:    decimal.NewFromFloat(50100),
		Last:   decimal.NewFromFloat(50050),
	}, nil
}

func (m *MockExchange) GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	return &OrderBook{
		Symbol: symbol,
		Bids: []Level{
			{Price: decimal.NewFromFloat(50000), Amount: decimal.NewFromFloat(1.0)},
		},
		Asks: []Level{
			{Price: decimal.NewFromFloat(50100), Amount: decimal.NewFromFloat(1.0)},
		},
	}, nil
}

func (m *MockExchange) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]Candle, error) {
	return []Candle{
		{
			Symbol: symbol,
			Open:   decimal.NewFromFloat(50000),
			High:   decimal.NewFromFloat(51000),
			Low:    decimal.NewFromFloat(49000),
			Close:  decimal.NewFromFloat(50500),
			Volume: decimal.NewFromFloat(100),
		},
	}, nil
}

func (m *MockExchange) SubscribeTicker(ctx context.Context, symbol string, callback func(*Ticker)) error {
	return nil
}

func (m *MockExchange) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*OrderBook)) error {
	return nil
}

func (m *MockExchange) SubscribeTrades(ctx context.Context, symbol string, callback func(*Trade)) error {
	return nil
}

func (m *MockExchange) Name() string {
	return m.name
}

func (m *MockExchange) SupportedSymbols() []string {
	return []string{"BTC-USD", "ETH-USD"}
}

// SetConnectError sets the error to return on Connect
func (m *MockExchange) SetConnectError(err error) {
	m.connectError = err
}

// SetBalanceError sets the error to return on GetBalance
func (m *MockExchange) SetBalanceError(err error) {
	m.balanceError = err
}

// SetPositionError sets the error to return on GetPositions
func (m *MockExchange) SetPositionError(err error) {
	m.positionError = err
}

// SetOrderError sets the error to return on GetOrders
func (m *MockExchange) SetOrderError(err error) {
	m.orderError = err
}
