package exchanges

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// ExchangeData represents aggregated data from a single exchange
type ExchangeData struct {
	Name      string
	Connected bool
	Balances  []Balance
	Positions []Position
	Orders    []Order
	Error     error
}

// AggregatedData represents data aggregated from all exchanges
type AggregatedData struct {
	Exchanges    map[string]*ExchangeData
	TotalBalance decimal.Decimal
	TotalPnL     decimal.Decimal
	LastUpdate   int64
}

// MultiExchangeAggregator manages multiple exchange clients and aggregates their data
type MultiExchangeAggregator struct {
	exchanges map[string]Exchange
	data      *AggregatedData
	mu        sync.RWMutex
}

// NewMultiExchangeAggregator creates a new aggregator with the given exchanges
func NewMultiExchangeAggregator(exchanges map[string]Exchange) *MultiExchangeAggregator {
	return &MultiExchangeAggregator{
		exchanges: exchanges,
		data: &AggregatedData{
			Exchanges:    make(map[string]*ExchangeData),
			TotalBalance: decimal.Zero,
			TotalPnL:     decimal.Zero,
		},
	}
}

// ConnectAll connects to all exchanges
func (a *MultiExchangeAggregator) ConnectAll(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for name, exchange := range a.exchanges {
		if err := exchange.Connect(ctx); err != nil {
			a.data.Exchanges[name] = &ExchangeData{
				Name:      name,
				Connected: false,
				Error:     err,
			}
			continue
		}

		a.data.Exchanges[name] = &ExchangeData{
			Name:      name,
			Connected: true,
		}
	}

	return nil
}

// DisconnectAll disconnects from all exchanges
func (a *MultiExchangeAggregator) DisconnectAll() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, exchange := range a.exchanges {
		exchange.Disconnect()
	}
}

// RefreshData refreshes data from all exchanges
func (a *MultiExchangeAggregator) RefreshData(ctx context.Context) error {
	a.mu.RLock()
	exchanges := make(map[string]Exchange, len(a.exchanges))
	for name, exchange := range a.exchanges {
		exchanges[name] = exchange
	}
	a.mu.RUnlock()

	totalBalance := decimal.Zero
	totalPnL := decimal.Zero
	results := make(map[string]*ExchangeData, len(exchanges))

	for name, exchange := range exchanges {
		exchangeData := &ExchangeData{
			Name:      name,
			Connected: exchange.IsConnected(),
		}

		balances, err := exchange.GetBalance(ctx)
		if err != nil {
			exchangeData.Error = err
			results[name] = exchangeData
			continue
		}
		exchangeData.Balances = balances

		for _, balance := range balances {
			if balance.Asset == "USD" || balance.Asset == "USDC" {
				totalBalance = totalBalance.Add(balance.Total)
			}
		}

		positions, err := exchange.GetPositions(ctx)
		if err != nil {
			exchangeData.Error = err
			results[name] = exchangeData
			continue
		}
		exchangeData.Positions = positions

		for _, position := range positions {
			totalPnL = totalPnL.Add(position.UnrealizedPnL)
		}

		if exchangeWithOrders, ok := exchange.(interface {
			GetOrders(context.Context) ([]Order, error)
		}); ok {
			if orders, err := exchangeWithOrders.GetOrders(ctx); err == nil {
				exchangeData.Orders = orders
			}
		}

		results[name] = exchangeData
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for name, data := range results {
		a.data.Exchanges[name] = data
	}
	a.data.TotalBalance = totalBalance
	a.data.TotalPnL = totalPnL
	a.data.LastUpdate = time.Now().Unix()

	return nil
}

// GetAggregatedData returns the current aggregated data (thread-safe)
func (a *MultiExchangeAggregator) GetAggregatedData() *AggregatedData {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy to prevent external modifications
	data := &AggregatedData{
		Exchanges:    make(map[string]*ExchangeData),
		TotalBalance: a.data.TotalBalance,
		TotalPnL:     a.data.TotalPnL,
		LastUpdate:   a.data.LastUpdate,
	}

	for name, exchangeData := range a.data.Exchanges {
		data.Exchanges[name] = &ExchangeData{
			Name:      exchangeData.Name,
			Connected: exchangeData.Connected,
			Balances:  append([]Balance(nil), exchangeData.Balances...),
			Positions: append([]Position(nil), exchangeData.Positions...),
			Orders:    append([]Order(nil), exchangeData.Orders...),
			Error:     exchangeData.Error,
		}
	}

	return data
}

// GetExchange returns a specific exchange by name
func (a *MultiExchangeAggregator) GetExchange(name string) (Exchange, bool) {
	exchange, exists := a.exchanges[name]
	return exchange, exists
}

// GetAllExchanges returns all exchanges
func (a *MultiExchangeAggregator) GetAllExchanges() map[string]Exchange {
	a.mu.RLock()
	defer a.mu.RUnlock()

	exchanges := make(map[string]Exchange)
	for name, exchange := range a.exchanges {
		exchanges[name] = exchange
	}
	return exchanges
}

// PlaceOrder places an order on a specific exchange
func (a *MultiExchangeAggregator) PlaceOrder(ctx context.Context, exchangeName string, order *Order) (*Order, error) {
	exchange, exists := a.exchanges[exchangeName]
	if !exists {
		return nil, fmt.Errorf("exchange %s not found", exchangeName)
	}

	return exchange.PlaceOrder(ctx, order)
}

// GetOrderBook gets order book from a specific exchange
func (a *MultiExchangeAggregator) GetOrderBook(ctx context.Context, exchangeName string, symbol string, depth int) (*OrderBook, error) {
	exchange, exists := a.exchanges[exchangeName]
	if !exists {
		return nil, fmt.Errorf("exchange %s not found", exchangeName)
	}

	return exchange.GetOrderBook(ctx, symbol, depth)
}
