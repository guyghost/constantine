package exchanges

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// OrderSide represents buy or sell
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeLimit     OrderType = "limit"
	OrderTypeMarket    OrderType = "market"
	OrderTypeStopLimit OrderType = "stop_limit"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusOpen      OrderStatus = "open"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCanceled  OrderStatus = "canceled"
	OrderStatusPartially OrderStatus = "partially_filled"
	OrderStatusExpired   OrderStatus = "expired"
	OrderStatusRejected  OrderStatus = "rejected"
)

// Common errors
var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrPositionNotFound = errors.New("position not found")
	ErrNotConnected     = errors.New("exchange not connected")
	ErrInvalidOrder     = errors.New("invalid order")
)

// Ticker represents market ticker data
type Ticker struct {
	Symbol    string
	Bid       decimal.Decimal
	Ask       decimal.Decimal
	Last      decimal.Decimal
	Volume24h decimal.Decimal
	Timestamp time.Time
}

// OrderBook represents the order book
type OrderBook struct {
	Symbol    string
	Bids      []Level
	Asks      []Level
	Timestamp time.Time
}

// Level represents a price level in the order book
type Level struct {
	Price  decimal.Decimal
	Amount decimal.Decimal
}

// Order represents a trading order
type Order struct {
	ID            string
	ClientOrderID string
	Symbol        string
	Side          OrderSide
	Type          OrderType
	Price         decimal.Decimal
	Amount        decimal.Decimal
	Filled        decimal.Decimal
	Remaining     decimal.Decimal
	Status        OrderStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	// Additional fields for advanced order types
	StopPrice    decimal.Decimal
	FilledAmount decimal.Decimal
	AveragePrice decimal.Decimal
}

// Trade represents a completed trade
type Trade struct {
	ID        string
	OrderID   string
	Symbol    string
	Side      OrderSide
	Price     decimal.Decimal
	Amount    decimal.Decimal
	Fee       decimal.Decimal
	Timestamp time.Time
}

// Position represents an open position
type Position struct {
	Symbol           string
	Side             OrderSide
	Size             decimal.Decimal
	EntryPrice       decimal.Decimal
	MarkPrice        decimal.Decimal
	Leverage         decimal.Decimal
	UnrealizedPnL    decimal.Decimal
	RealizedPnL      decimal.Decimal
	LiquidationPrice decimal.Decimal
}

// Balance represents account balance
type Balance struct {
	Asset     string
	Free      decimal.Decimal
	Locked    decimal.Decimal
	Total     decimal.Decimal
	UpdatedAt time.Time
}

// Candle represents OHLCV data
type Candle struct {
	Symbol    string
	Timestamp time.Time
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
}

// Exchange defines the interface all exchanges must implement
type Exchange interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	// Market data
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)
	GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)
	GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]Candle, error)
	SubscribeTicker(ctx context.Context, symbol string, callback func(*Ticker)) error
	SubscribeOrderBook(ctx context.Context, symbol string, callback func(*OrderBook)) error
	SubscribeTrades(ctx context.Context, symbol string, callback func(*Trade)) error
	SubscribeCandles(ctx context.Context, symbol string, interval string, callback func(*Candle)) error

	// Trading
	PlaceOrder(ctx context.Context, order *Order) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetOpenOrders(ctx context.Context, symbol string) ([]Order, error)
	GetOrderHistory(ctx context.Context, symbol string, limit int) ([]Order, error)

	// Account
	GetBalance(ctx context.Context) ([]Balance, error)
	GetPositions(ctx context.Context) ([]Position, error)
	GetPosition(ctx context.Context, symbol string) (*Position, error)

	// Metadata
	Name() string
	SupportedSymbols() []string
}
