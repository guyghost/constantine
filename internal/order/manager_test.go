package order

import (
	"context"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/testutils"
	"github.com/shopspring/decimal"
)

func TestNewManager(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	testutils.AssertNotNil(t, manager, "Manager should not be nil")
	testutils.AssertEqual(t, exchange, manager.exchange, "Exchange should match")
	testutils.AssertNotNil(t, manager.orderBook, "Order book should not be nil")
	testutils.AssertFalse(t, manager.running, "Manager should not be running initially")
}

func TestManager_PlaceOrder(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	req := &OrderRequest{
		Symbol:   "BTC-USD",
		Side:     exchanges.OrderSideBuy,
		Type:     exchanges.OrderTypeLimit,
		Price:    decimal.NewFromFloat(50000),
		Amount:   decimal.NewFromFloat(0.1),
		StopLoss: decimal.NewFromFloat(49500),
	}

	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	order, err := manager.PlaceOrder(ctx, req)

	testutils.AssertNoError(t, err, "PlaceOrder should not return error")
	testutils.AssertNotNil(t, order, "Order should not be nil")
	testutils.AssertEqual(t, "BTC-USD", order.Symbol, "Order symbol should match")
	testutils.AssertEqual(t, exchanges.OrderSideBuy, order.Side, "Order side should match")
	testutils.AssertTrue(t, order.Price.Equal(req.Price), "Order price should match request")
	testutils.AssertTrue(t, order.Amount.Equal(req.Amount), "Order amount should match request")
}

func TestManager_CancelOrder(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// First place an order
	req := &OrderRequest{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
	}

	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	placedOrder, err := manager.PlaceOrder(ctx, req)
	testutils.AssertNoError(t, err, "PlaceOrder should not return error")

	// Cancel the order
	err = manager.CancelOrder(ctx, placedOrder.ID)
	testutils.AssertNoError(t, err, "CancelOrder should not return error")
}

func TestManager_GetOpenOrders(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Initially should have no orders
	orders := manager.GetOpenOrders()
	testutils.AssertEqual(t, 0, len(orders), "Should have no open orders initially")

	// Place an order
	req := &OrderRequest{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
	}

	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	_, err := manager.PlaceOrder(ctx, req)
	testutils.AssertNoError(t, err, "PlaceOrder should not return error")

	// Should now have 1 order
	orders = manager.GetOpenOrders()
	testutils.AssertEqual(t, 1, len(orders), "Should have 1 open order")
	testutils.AssertEqual(t, "BTC-USD", orders[0].Symbol, "Order symbol should match")
}

func TestManager_GetPositions(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Initially should have no positions
	positions := manager.GetPositions()
	testutils.AssertEqual(t, 0, len(positions), "Should have no positions initially")

	// Note: In a real scenario, positions would be created when orders are filled
	// For this test, we verify the method returns an empty slice
}

func TestManager_GetPosition(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Test non-existent position
	position := manager.GetPosition("BTC-USD")
	if position != nil {
		t.Errorf("Should return nil for non-existent position, got %v", position)
	}

	// Note: Position creation would happen through order fulfillment
	// This test verifies the method handles missing positions correctly
}

func TestManager_ClosePosition(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Try to close a non-existent position
	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	err := manager.ClosePosition(ctx, "BTC-USD")
	testutils.AssertError(t, err, "ClosePosition should return error for non-existent position")

	// Note: In a real scenario with actual positions, this would close them
	// This test verifies the method handles missing positions correctly
}

func TestManager_StartStop(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Initially not running
	testutils.AssertFalse(t, manager.running, "Manager should not be running initially")

	// Start
	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	err := manager.Start(ctx)
	testutils.AssertNoError(t, err, "Start should not return error")

	// Should be running
	testutils.AssertTrue(t, manager.running, "Manager should be running after start")

	// Stop
	err = manager.Stop()
	testutils.AssertNoError(t, err, "Stop should not return error")

	// Should not be running
	time.Sleep(100 * time.Millisecond) // Allow time for cleanup
	testutils.AssertFalse(t, manager.running, "Manager should not be running after stop")
}

func TestManager_Callbacks(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Test error callback
	errorReceived := false
	var receivedError error
	manager.SetErrorCallback(func(err error) {
		errorReceived = true
		receivedError = err
	})

	// Emit an error to test callback
	testError := context.Canceled
	manager.emitError(testError)

	testutils.AssertTrue(t, errorReceived, "Error callback should have been called")
	testutils.AssertEqual(t, testError, receivedError, "Received error should match emitted error")

	// Note: Order and position callbacks would be tested through actual order/position operations
	// which would require more complex setup with filled orders
}

func TestManager_OrderStatusChanges(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Place an order
	req := &OrderRequest{
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Type:   exchanges.OrderTypeLimit,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
	}

	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	order, err := manager.PlaceOrder(ctx, req)
	testutils.AssertNoError(t, err, "PlaceOrder should not return error")

	// Simulate order being filled
	filledOrder := *order
	filledOrder.Status = exchanges.OrderStatusFilled
	filledOrder.Filled = order.Amount

	manager.handleOrderStatusChange(&filledOrder, order)

	// Check that order was moved to filled orders
	testutils.AssertEqual(t, 0, len(manager.GetOpenOrders()), "Should have no open orders after fill")
	testutils.AssertEqual(t, 1, len(manager.orderBook.FilledOrders), "Should have 1 filled order")

	// Check that position was created
	positions := manager.GetPositions()
	testutils.AssertEqual(t, 1, len(positions), "Should have 1 position after fill")
	testutils.AssertEqual(t, "BTC-USD", positions[0].Symbol, "Position symbol should match")
	testutils.AssertEqual(t, PositionSideLong, positions[0].Side, "Position side should be long")
	testutils.AssertTrue(t, positions[0].EntryPrice.Equal(order.Price), "Position entry price should match order price")
}

func TestManager_CalculatePnL(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Create a long position
	position := &ManagedPosition{
		Symbol:     "BTC-USD",
		Side:       PositionSideLong,
		EntryPrice: decimal.NewFromFloat(50000),
		Amount:     decimal.NewFromFloat(0.1),
	}

	// Test profit scenario
	exitPrice := decimal.NewFromFloat(51000)
	pnl := manager.calculatePnL(position, exitPrice)
	expectedPnL := decimal.NewFromFloat(100) // (51000 - 50000) * 0.1
	testutils.AssertTrue(t, pnl.Equal(expectedPnL), "PnL should be positive for profitable long position")

	// Test loss scenario
	exitPrice = decimal.NewFromFloat(49000)
	pnl = manager.calculatePnL(position, exitPrice)
	expectedPnL = decimal.NewFromFloat(-100) // (49000 - 50000) * 0.1
	testutils.AssertTrue(t, pnl.Equal(expectedPnL), "PnL should be negative for losing long position")

	// Test short position profit
	position.Side = PositionSideShort
	exitPrice = decimal.NewFromFloat(49000)
	pnl = manager.calculatePnL(position, exitPrice)
	expectedPnL = decimal.NewFromFloat(100) // (50000 - 49000) * 0.1
	testutils.AssertTrue(t, pnl.Equal(expectedPnL), "PnL should be positive for profitable short position")
}

func TestManager_HandleFilledOrder(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Test creating new position
	order := &exchanges.Order{
		ID:     "test-order-1",
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
		Filled: decimal.NewFromFloat(0.1),
		Status: exchanges.OrderStatusFilled,
	}

	manager.handleFilledOrder(order)

	positions := manager.GetPositions()
	testutils.AssertEqual(t, 1, len(positions), "Should have 1 position")
	position := positions[0]
	testutils.AssertEqual(t, "BTC-USD", position.Symbol, "Position symbol should match")
	testutils.AssertEqual(t, PositionSideLong, position.Side, "Position side should be long")
	testutils.AssertTrue(t, position.EntryPrice.Equal(order.Price), "Entry price should match order price")
	testutils.AssertTrue(t, position.Amount.Equal(order.Filled), "Position amount should match filled amount")

	// Test closing position with opposite order
	closeOrder := &exchanges.Order{
		ID:     "test-order-2",
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideSell,
		Price:  decimal.NewFromFloat(51000),
		Amount: decimal.NewFromFloat(0.1),
		Filled: decimal.NewFromFloat(0.1),
		Status: exchanges.OrderStatusFilled,
	}

	manager.handleFilledOrder(closeOrder)

	positions = manager.GetPositions()
	testutils.AssertEqual(t, 0, len(positions), "Should have no positions after closing")
}

func TestManager_UpdatePositions(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	// Create a position manually
	position := &ManagedPosition{
		ID:            "test-pos",
		Symbol:        "BTC-USD",
		Side:          PositionSideLong,
		EntryPrice:    decimal.NewFromFloat(50000),
		CurrentPrice:  decimal.NewFromFloat(50000),
		Amount:        decimal.NewFromFloat(0.1),
		UnrealizedPnL: decimal.Zero,
		Status:        PositionStatusOpen,
	}

	manager.orderBook.Positions["BTC-USD"] = position

	// Mock exchange positions
	exchangePositions := []exchanges.Position{
		{
			Symbol:        "BTC-USD",
			MarkPrice:     decimal.NewFromFloat(50500),
			UnrealizedPnL: decimal.NewFromFloat(50),
		},
	}

	// Set up the mock to return these positions
	testExchange := exchange
	testExchange.PositionsValue = exchangePositions

	ctx, cancel := testutils.CreateTestContext()
	defer cancel()
	manager.updatePositions(ctx)

	// Check that position was updated
	updatedPosition := manager.GetPosition("BTC-USD")
	testutils.AssertNotNil(t, updatedPosition, "Position should exist")
	testutils.AssertTrue(t, updatedPosition.CurrentPrice.Equal(decimal.NewFromFloat(50500)), "Current price should be updated")
	testutils.AssertTrue(t, updatedPosition.UnrealizedPnL.Equal(decimal.NewFromFloat(50)), "Unrealized PnL should be updated")
}

func TestManager_GetStats(t *testing.T) {
	exchange := testutils.NewTestExchange("test-exchange")
	manager := NewManager(exchange)

	stats := manager.GetStats()

	testutils.AssertNotNil(t, stats, "Stats should not be nil")
	testutils.AssertEqual(t, 0, stats.TotalOrders, "Total orders should be 0 initially")
	testutils.AssertEqual(t, 0, stats.FilledOrders, "Filled orders should be 0 initially")
	testutils.AssertEqual(t, 0, stats.CanceledOrders, "Cancelled orders should be 0 initially")

	// Add some filled orders
	filledOrder := &exchanges.Order{
		ID:     "filled-1",
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
		Filled: decimal.NewFromFloat(0.1),
		Status: exchanges.OrderStatusFilled,
	}

	cancelledOrder := &exchanges.Order{
		ID:     "cancelled-1",
		Symbol: "BTC-USD",
		Side:   exchanges.OrderSideBuy,
		Price:  decimal.NewFromFloat(50000),
		Amount: decimal.NewFromFloat(0.1),
		Filled: decimal.Zero,
		Status: exchanges.OrderStatusCanceled,
	}

	manager.orderBook.FilledOrders = append(manager.orderBook.FilledOrders, filledOrder, cancelledOrder)

	stats = manager.GetStats()
	testutils.AssertEqual(t, 2, stats.TotalOrders, "Total orders should be 2")
	testutils.AssertEqual(t, 2, stats.FilledOrders, "Filled orders should be 2")
	testutils.AssertEqual(t, 1, stats.CanceledOrders, "Cancelled orders should be 1")
	testutils.AssertEqual(t, 1.0, stats.SuccessRate, "Success rate should be 1.0")
}
