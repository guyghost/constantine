package exchanges

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// ExchangeMultiplexer routes orders and aggregates data across multiple exchanges
type ExchangeMultiplexer struct {
	mu        sync.RWMutex
	exchanges map[string]Exchange // exchange name -> exchange
	symbolMap map[string]string   // symbol -> exchange name
	data      *AggregatedData
}

// NewExchangeMultiplexer creates a new exchange multiplexer
func NewExchangeMultiplexer() *ExchangeMultiplexer {
	return &ExchangeMultiplexer{
		exchanges: make(map[string]Exchange),
		symbolMap: make(map[string]string),
		data: &AggregatedData{
			Exchanges:    make(map[string]*ExchangeData),
			TotalBalance: decimal.Zero,
			TotalPnL:     decimal.Zero,
		},
	}
}

// ConnectAll connects to all exchanges
func (em *ExchangeMultiplexer) ConnectAll(ctx context.Context) error {
	em.mu.RLock()
	exchanges := make(map[string]Exchange)
	for k, v := range em.exchanges {
		exchanges[k] = v
	}
	em.mu.RUnlock()

	for name, exchange := range exchanges {
		if err := exchange.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to %s: %w", name, err)
		}
	}
	return nil
}

// DisconnectAll disconnects from all exchanges
func (em *ExchangeMultiplexer) DisconnectAll() error {
	em.mu.RLock()
	exchanges := make(map[string]Exchange)
	for k, v := range em.exchanges {
		exchanges[k] = v
	}
	em.mu.RUnlock()

	for _, exchange := range exchanges {
		if err := exchange.Disconnect(); err != nil {
			// Log but continue
			continue
		}
	}
	return nil
}

// RefreshData refreshes data from all exchanges
func (em *ExchangeMultiplexer) RefreshData(ctx context.Context) error {
	em.mu.RLock()
	exchanges := make(map[string]Exchange)
	for k, v := range em.exchanges {
		exchanges[k] = v
	}
	em.mu.RUnlock()

	aggregated := &AggregatedData{
		Exchanges:    make(map[string]*ExchangeData),
		TotalBalance: decimal.Zero,
		TotalPnL:     decimal.Zero,
		LastUpdate:   time.Now().Unix(),
	}

	for name, exchange := range exchanges {
		exchangeData := &ExchangeData{
			Name:      name,
			Connected: exchange.IsConnected(),
		}

		// Get balances
		balances, err := exchange.GetBalance(ctx)
		if err != nil {
			exchangeData.Error = err
		} else {
			exchangeData.Balances = balances
			// Aggregate total balance (sum of all assets)
			for _, balance := range balances {
				aggregated.TotalBalance = aggregated.TotalBalance.Add(balance.Total)
			}
		}

		// Get positions
		positions, err := exchange.GetPositions(ctx)
		if err != nil {
			if exchangeData.Error == nil {
				exchangeData.Error = err
			}
		} else {
			exchangeData.Positions = positions
			// Aggregate PnL
			for _, pos := range positions {
				aggregated.TotalPnL = aggregated.TotalPnL.Add(pos.UnrealizedPnL).Add(pos.RealizedPnL)
			}
		}

		// Get open orders
		orders, err := exchange.GetOpenOrders(ctx, "")
		if err != nil {
			if exchangeData.Error == nil {
				exchangeData.Error = err
			}
		} else {
			exchangeData.Orders = orders
		}

		aggregated.Exchanges[name] = exchangeData
	}

	em.mu.Lock()
	em.data = aggregated
	em.mu.Unlock()

	return nil
}

// GetAggregatedData returns the latest aggregated data
func (em *ExchangeMultiplexer) GetAggregatedData() *AggregatedData {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.data
}

// AddExchange adds an exchange to the multiplexer
func (em *ExchangeMultiplexer) AddExchange(name string, exchange Exchange) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.exchanges[name] = exchange
}

// MapSymbol maps a symbol to a specific exchange
func (em *ExchangeMultiplexer) MapSymbol(symbol, exchangeName string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.exchanges[exchangeName]; !exists {
		return fmt.Errorf("exchange %s not found", exchangeName)
	}

	em.symbolMap[symbol] = exchangeName
	return nil
}

// GetExchangeForSymbol returns the exchange for a given symbol
func (em *ExchangeMultiplexer) GetExchangeForSymbol(symbol string) (Exchange, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	exchangeName, exists := em.symbolMap[symbol]
	if !exists {
		return nil, fmt.Errorf("no exchange mapped for symbol %s", symbol)
	}

	exchange, exists := em.exchanges[exchangeName]
	if !exists {
		return nil, fmt.Errorf("exchange %s not found", exchangeName)
	}

	return exchange, nil
}

// PlaceOrder places an order on the appropriate exchange for the symbol
func (em *ExchangeMultiplexer) PlaceOrder(ctx context.Context, order *Order) (*Order, error) {
	exchange, err := em.GetExchangeForSymbol(order.Symbol)
	if err != nil {
		return nil, err
	}

	return exchange.PlaceOrder(ctx, order)
}

// GetPositions aggregates positions from all exchanges
func (em *ExchangeMultiplexer) GetPositions(ctx context.Context) ([]Position, error) {
	em.mu.RLock()
	exchanges := make(map[string]Exchange)
	for k, v := range em.exchanges {
		exchanges[k] = v
	}
	em.mu.RUnlock()

	var allPositions []Position
	for _, exchange := range exchanges {
		positions, err := exchange.GetPositions(ctx)
		if err != nil {
			// Log error but continue with other exchanges
			continue
		}
		allPositions = append(allPositions, positions...)
	}

	return allPositions, nil
}

// GetBalance aggregates balances from all exchanges
func (em *ExchangeMultiplexer) GetBalance(ctx context.Context) ([]Balance, error) {
	em.mu.RLock()
	exchanges := make(map[string]Exchange)
	for k, v := range em.exchanges {
		exchanges[k] = v
	}
	em.mu.RUnlock()

	balanceMap := make(map[string]Balance)

	for _, exchange := range exchanges {
		balances, err := exchange.GetBalance(ctx)
		if err != nil {
			// Log error but continue
			continue
		}

		// Aggregate balances by asset
		for _, balance := range balances {
			if existing, exists := balanceMap[balance.Asset]; exists {
				existing.Total = existing.Total.Add(balance.Total)
				existing.Free = existing.Free.Add(balance.Free)
				existing.Locked = existing.Locked.Add(balance.Locked)
				balanceMap[balance.Asset] = existing
			} else {
				balanceMap[balance.Asset] = balance
			}
		}
	}

	var aggregated []Balance
	for _, balance := range balanceMap {
		aggregated = append(aggregated, balance)
	}

	return aggregated, nil
}

// GetExchanges returns all registered exchanges
func (em *ExchangeMultiplexer) GetExchanges() map[string]Exchange {
	em.mu.RLock()
	defer em.mu.RUnlock()

	exchanges := make(map[string]Exchange)
	for k, v := range em.exchanges {
		exchanges[k] = v
	}
	return exchanges
}

// GetSymbolMap returns the symbol to exchange mapping
func (em *ExchangeMultiplexer) GetSymbolMap() map[string]string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	symbolMap := make(map[string]string)
	for k, v := range em.symbolMap {
		symbolMap[k] = v
	}
	return symbolMap
}
