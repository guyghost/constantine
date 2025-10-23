package order

import (
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// OrderRequest represents a request to place an order
type OrderRequest struct {
	Symbol      string
	Side        exchanges.OrderSide
	Type        exchanges.OrderType
	Price       decimal.Decimal
	Amount      decimal.Decimal
	StopLoss    decimal.Decimal
	TakeProfit  decimal.Decimal
	TimeInForce string
	ReduceOnly  bool
}

// OrderUpdate represents an order status update
type OrderUpdate struct {
	Order     *exchanges.Order
	Event     OrderEvent
	Timestamp time.Time
}

// OrderEvent represents the type of order event
type OrderEvent string

const (
	OrderEventCreated         OrderEvent = "created"
	OrderEventFilled          OrderEvent = "filled"
	OrderEventPartiallyFilled OrderEvent = "partially_filled"
	OrderEventCanceled        OrderEvent = "canceled"
	OrderEventRejected        OrderEvent = "rejected"
	OrderEventExpired         OrderEvent = "expired"
)

// OrderStats holds statistics about orders
type OrderStats struct {
	TotalOrders     int
	FilledOrders    int
	CanceledOrders  int
	RejectedOrders  int
	TotalVolume     decimal.Decimal
	TotalFees       decimal.Decimal
	AverageFillTime time.Duration
	SuccessRate     float64
}

// PositionSide represents the side of a position
type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

// PositionStatus represents the status of a position
type PositionStatus string

const (
	PositionStatusOpen   PositionStatus = "open"
	PositionStatusClosed PositionStatus = "closed"
)

// ManagedPosition represents a position managed by the order manager
type ManagedPosition struct {
	ID                string
	Symbol            string
	Side              PositionSide
	EntryPrice        decimal.Decimal
	CurrentPrice      decimal.Decimal
	Amount            decimal.Decimal
	Leverage          decimal.Decimal
	StopLoss          decimal.Decimal
	TakeProfit        decimal.Decimal
	UnrealizedPnL     decimal.Decimal
	RealizedPnL       decimal.Decimal
	EntryTime         time.Time
	ExitTime          *time.Time
	Status            PositionStatus
	EntryOrderID      string
	ExitOrderID       string
	StopLossOrderID   string
	TakeProfitOrderID string
}

// OrderBook represents the current state of orders
type OrderBook struct {
	OpenOrders    map[string]*exchanges.Order
	FilledOrders  []*exchanges.Order
	Positions     map[string]*ManagedPosition
	PendingOrders map[string]*OrderRequest
}

// NewOrderBook creates a new order book
func NewOrderBook() *OrderBook {
	return &OrderBook{
		OpenOrders:    make(map[string]*exchanges.Order),
		FilledOrders:  make([]*exchanges.Order, 0),
		Positions:     make(map[string]*ManagedPosition),
		PendingOrders: make(map[string]*OrderRequest),
	}
}
