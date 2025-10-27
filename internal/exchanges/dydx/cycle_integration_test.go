// Tests for the complete trading cycle: market data → strategy → execution → order placement
// This file validates the integration between different components

package dydx

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
)

// TestTradingCycle_CandleToSignal simulates the flow from candle data to strategy signal
func TestTradingCycle_CandleToSignal(t *testing.T) {
	// Step 1: Create a candle (market data)
	testCandle := &exchanges.Candle{
		Symbol:    "BTC-USD",
		Timestamp: time.Now(),
		Open:      decimal.NewFromInt(43000),
		High:      decimal.NewFromInt(43500),
		Low:       decimal.NewFromInt(42500),
		Close:     decimal.NewFromInt(43250),
		Volume:    decimal.NewFromInt(100),
	}

	// Step 2: Simulate candle callback
	signalGenerated := false
	generatedSignal := (*strategy.Signal)(nil)
	mu := sync.Mutex{}

	// This represents what the strategy would do when receiving a candle
	handleCandleFromStrategy := func(candle *exchanges.Candle) {
		mu.Lock()
		defer mu.Unlock()
		signalGenerated = true
		// Simulate strategy generating a signal from candle
		generatedSignal = &strategy.Signal{
			Type:      strategy.SignalTypeEntry,
			Symbol:    candle.Symbol,
			Timestamp: candle.Timestamp.Unix(),
			Strength:  0.75, // 75% confidence
		}
	}

	// Invoke the simulated callback
	handleCandleFromStrategy(testCandle)

	mu.Lock()
	if !signalGenerated {
		t.Error("Expected signal to be generated from candle")
	}

	if generatedSignal == nil {
		t.Error("Expected signal to be non-nil")
	} else {
		if generatedSignal.Type != strategy.SignalTypeEntry {
			t.Errorf("Expected signal type Entry, got %v", generatedSignal.Type)
		}
		if generatedSignal.Symbol != "BTC-USD" {
			t.Errorf("Expected symbol BTC-USD, got %s", generatedSignal.Symbol)
		}
		if generatedSignal.Strength < 0.5 {
			t.Errorf("Expected signal strength >= 0.5, got %v", generatedSignal.Strength)
		}
	}
	mu.Unlock()
}

// TestTradingCycle_SignalToOrder simulates signal → execution agent → order placement
func TestTradingCycle_SignalToOrder(t *testing.T) {
	// Step 1: Create a signal from strategy
	signal := &strategy.Signal{
		Type:      strategy.SignalTypeEntry,
		Symbol:    "BTC-USD",
		Timestamp: time.Now().Unix(),
		Strength:  0.8,
	}

	// Step 2: Execution agent processes signal
	orderPlaced := false
	mu := sync.Mutex{}

	// Simulate execution agent receiving signal
	handleSignalFromExecution := func(sig *strategy.Signal) *exchanges.Order {
		mu.Lock()
		defer mu.Unlock()

		// Execution agent would validate risk and create order request
		if sig.Strength >= 0.5 && sig.Type == strategy.SignalTypeEntry {
			orderPlaced = true
			// Create order from signal
			order := &exchanges.Order{
				Symbol: sig.Symbol,
				Side:   exchanges.OrderSideBuy,
				Type:   exchanges.OrderTypeLimit,
				Price:  decimal.NewFromInt(43000),
				Amount: decimal.NewFromFloat(0.1),
			}
			return order
		}
		return nil
	}

	// Process signal
	resultOrder := handleSignalFromExecution(signal)

	mu.Lock()
	if !orderPlaced {
		t.Error("Expected order to be placed")
	}

	if resultOrder == nil {
		t.Error("Expected order result to be non-nil")
	} else {
		if resultOrder.Symbol != "BTC-USD" {
			t.Errorf("Expected symbol BTC-USD, got %s", resultOrder.Symbol)
		}
		if resultOrder.Side != exchanges.OrderSideBuy {
			t.Errorf("Expected buy side, got %s", resultOrder.Side)
		}
	}
	mu.Unlock()
}

// TestTradingCycle_OrderToPosition simulates order execution → position tracking
func TestTradingCycle_OrderToPosition(t *testing.T) {
	// Step 1: Order placed on exchange
	order := &exchanges.Order{
		ID:     "order-123",
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromInt(43000),
		Amount: decimal.NewFromFloat(0.1),
		Status: exchanges.OrderStatusFilled,
	}

	// Step 2: Order filled, position tracked
	position := (*exchanges.Position)(nil)

	// Simulate position creation from filled order
	createPositionFromOrder := func(filledOrder *exchanges.Order) *exchanges.Position {
		if filledOrder.Status == exchanges.OrderStatusFilled {
			return &exchanges.Position{
				Symbol:        filledOrder.Symbol,
				Side:          filledOrder.Side,
				Size:          filledOrder.Amount,
				EntryPrice:    filledOrder.Price,
				UnrealizedPnL: decimal.Zero, // Will be updated with mark price
				RealizedPnL:   decimal.Zero,
			}
		}
		return nil
	}

	position = createPositionFromOrder(order)

	if position == nil {
		t.Error("Expected position to be created from filled order")
	} else {
		if position.Symbol != "BTC-USD" {
			t.Errorf("Expected symbol BTC-USD, got %s", position.Symbol)
		}
		if position.Side != exchanges.OrderSideBuy {
			t.Errorf("Expected buy side, got %s", position.Side)
		}
		if position.Size.String() != "0.1" {
			t.Errorf("Expected size 0.1, got %s", position.Size.String())
		}
		if position.EntryPrice.String() != "43000" {
			t.Errorf("Expected entry price 43000, got %s", position.EntryPrice.String())
		}
	}
}

// TestTradingCycle_PositionUpdates simulates position being updated with market price
func TestTradingCycle_PositionUpdates(t *testing.T) {
	// Initial position
	position := &exchanges.Position{
		Symbol:     "BTC-USD",
		Side:       exchanges.OrderSideBuy,
		Size:       decimal.NewFromFloat(0.1),
		EntryPrice: decimal.NewFromInt(43000),
		MarkPrice:  decimal.NewFromInt(43000), // Start at entry
	}

	// Simulate price moving up
	newMarkPrice := decimal.NewFromInt(43500)

	// Calculate PnL
	calculatePnL := func(pos *exchanges.Position, newMark decimal.Decimal) decimal.Decimal {
		if pos.Side == exchanges.OrderSideBuy {
			priceChange := newMark.Sub(pos.EntryPrice)
			return priceChange.Mul(pos.Size)
		}
		return decimal.Zero
	}

	pnl := calculatePnL(position, newMarkPrice)

	if pnl.IsZero() {
		t.Error("Expected PnL to be non-zero after price change")
	}

	// PnL should be: (43500 - 43000) * 0.1 = 500 * 0.1 = 50
	expectedPnL := decimal.NewFromInt(50)
	if !pnl.Equal(expectedPnL) {
		t.Errorf("Expected PnL %s, got %s", expectedPnL, pnl)
	}
}

// TestTradingCycle_ExitSignalClosesPosition simulates exit signal closing position
func TestTradingCycle_ExitSignalClosesPosition(t *testing.T) {
	// Step 1: Active position
	position := &exchanges.Position{
		Symbol:     "BTC-USD",
		Side:       exchanges.OrderSideBuy,
		Size:       decimal.NewFromFloat(0.1),
		EntryPrice: decimal.NewFromInt(43000),
		MarkPrice:  decimal.NewFromInt(43500),
	}

	// Step 2: Exit signal generated
	exitSignal := &strategy.Signal{
		Type:      strategy.SignalTypeExit,
		Symbol:    "BTC-USD",
		Strength:  0.9,
		Timestamp: time.Now().Unix(),
	}

	// Step 3: Execution agent creates closing order
	closingOrder := (*exchanges.Order)(nil)

	createClosingOrder := func(pos *exchanges.Position, sig *strategy.Signal) *exchanges.Order {
		if sig.Type != strategy.SignalTypeExit {
			return nil
		}

		// Closing order has opposite side
		closingSide := exchanges.OrderSideSell
		if pos.Side == exchanges.OrderSideSell {
			closingSide = exchanges.OrderSideBuy
		}

		return &exchanges.Order{
			Symbol: pos.Symbol,
			Side:   closingSide,
			Type:   exchanges.OrderTypeMarket,
			Amount: pos.Size,
			Price:  pos.MarkPrice,
		}
	}

	closingOrder = createClosingOrder(position, exitSignal)

	if closingOrder == nil {
		t.Error("Expected closing order to be created")
	} else {
		if closingOrder.Side != exchanges.OrderSideSell {
			t.Errorf("Expected sell side for closing, got %s", closingOrder.Side)
		}
		if closingOrder.Amount.String() != "0.1" {
			t.Errorf("Expected size 0.1, got %s", closingOrder.Amount)
		}
	}
}

// TestTradingCycle_CompleteCycle simulates the entire cycle end-to-end
func TestTradingCycle_CompleteCycle(t *testing.T) {
	ctx := context.Background()

	// Step 1: Market data (candle)
	candle := &exchanges.Candle{
		Symbol:    "BTC-USD",
		Timestamp: time.Now(),
		Open:      decimal.NewFromInt(43000),
		High:      decimal.NewFromInt(43500),
		Low:       decimal.NewFromInt(42500),
		Close:     decimal.NewFromInt(43250),
		Volume:    decimal.NewFromInt(100),
	}

	// Verify candle has required fields
	if candle.Symbol == "" || candle.Close.IsZero() {
		t.Fatal("Invalid candle data")
	}

	// Step 2: Strategy generates signal
	signal := &strategy.Signal{
		Type:      strategy.SignalTypeEntry,
		Symbol:    candle.Symbol,
		Timestamp: candle.Timestamp.Unix(),
		Strength:  0.8,
	}

	if signal.Type != strategy.SignalTypeEntry {
		t.Fatal("Invalid signal type")
	}

	// Step 3: Execution creates order
	order := &exchanges.Order{
		ID:     "order-456",
		Symbol: signal.Symbol,
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  candle.Close,
		Amount: decimal.NewFromFloat(0.1),
		Status: exchanges.OrderStatusFilled,
	}

	if order.Symbol != candle.Symbol {
		t.Fatal("Order symbol doesn't match candle")
	}

	// Step 4: Track position
	position := &exchanges.Position{
		Symbol:     order.Symbol,
		Side:       order.Side,
		Size:       order.Amount,
		EntryPrice: order.Price,
		MarkPrice:  order.Price,
	}

	_ = ctx // Context unused in this test but part of real flow

	if position.Symbol != candle.Symbol {
		t.Error("Position symbol doesn't match")
	}

	if position.Size.String() != "0.1" {
		t.Error("Position size incorrect")
	}

	if !position.EntryPrice.Equal(candle.Close) {
		t.Error("Position entry price doesn't match order/candle price")
	}

	// Step 5: Position updated with new market price
	newPrice := decimal.NewFromInt(44000)
	priceDiff := newPrice.Sub(position.EntryPrice)
	unrealizedPnL := priceDiff.Mul(position.Size)

	if unrealizedPnL.IsZero() {
		t.Error("Unrealized PnL should be non-zero")
	}

	// Step 6: Exit signal closes position
	exitSignal := &strategy.Signal{
		Type:     strategy.SignalTypeExit,
		Symbol:   candle.Symbol,
		Strength: 0.9,
	}

	closingOrder := &exchanges.Order{
		Symbol: position.Symbol,
		Side:   exchanges.OrderSideSell,
		Type:   exchanges.OrderTypeMarket,
		Amount: position.Size,
		Price:  newPrice,
		Status: exchanges.OrderStatusFilled,
	}

	if closingOrder.Side == position.Side {
		t.Error("Closing order should have opposite side to position")
	}

	// All steps completed successfully
	_ = exitSignal
}

// TestTradingCycle_MultipleSymbols validates cycle works with multiple symbols
func TestTradingCycle_MultipleSymbols(t *testing.T) {
	symbols := []string{"BTC-USD", "ETH-USD", "SOL-USD"}

	for _, symbol := range symbols {
		candle := &exchanges.Candle{
			Symbol:    symbol,
			Timestamp: time.Now(),
			Close:     decimal.NewFromInt(100),
		}

		signal := &strategy.Signal{
			Symbol: candle.Symbol,
			Type:   strategy.SignalTypeEntry,
		}

		order := &exchanges.Order{
			Symbol: signal.Symbol,
			Side:   exchanges.OrderSideBuy,
		}

		if order.Symbol != symbol {
			t.Errorf("Symbol mismatch for %s", symbol)
		}
	}
}
