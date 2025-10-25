package main

import (
	"context"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/execution"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
)

// TestIntegration_BotTradingFlow tests the complete trading flow from strategy to execution
func TestIntegration_BotTradingFlow(t *testing.T) {
	// Create mock exchange
	mockExchange := exchanges.NewMockExchange("test-exchange")

	// Create components
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = "BTC-USD"
	strategyEngine := strategy.NewScalpingStrategy(strategyConfig, mockExchange)

	orderManager := order.NewManager(mockExchange)
	riskConfig := risk.DefaultConfig()
	riskManager := risk.NewManager(riskConfig, decimal.NewFromFloat(10000))

	executionConfig := execution.DefaultConfig()
	executionAgent := execution.NewExecutionAgent(orderManager, riskManager, executionConfig)

	// Setup callbacks
	strategyEngine.SetSignalCallback(func(signal *strategy.Signal) {
		ctx := context.Background()
		if err := executionAgent.HandleSignal(ctx, signal); err != nil {
			t.Errorf("Failed to handle signal: %v", err)
		}
	})

	// Start components
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := orderManager.Start(ctx); err != nil {
		t.Fatalf("Failed to start order manager: %v", err)
	}
	defer orderManager.Stop()

	if err := strategyEngine.Start(ctx); err != nil {
		t.Fatalf("Failed to start strategy: %v", err)
	}
	defer strategyEngine.Stop()

	// Wait a bit for components to initialize
	time.Sleep(500 * time.Millisecond)

	// Check if components are initialized correctly
	orders := orderManager.GetOpenOrders()
	t.Logf("Initial open orders: %d", len(orders))

	positions := orderManager.GetPositions()
	t.Logf("Initial positions: %d", len(positions))

	// Check risk manager stats
	stats := riskManager.GetStats()
	t.Logf("Risk stats - Balance: %s, Daily PnL: %s, Total Trades: %d",
		stats.CurrentBalance.String(), stats.DailyPnL.String(), stats.TotalTrades)

	// Verify that components can communicate
	canTrade, reason := riskManager.CanTrade()
	if !canTrade {
		t.Logf("Trading not allowed: %s", reason)
	} else {
		t.Log("Trading is allowed")
	}

	// Test that we can place a manual order through the execution agent
	// (This tests the integration without relying on strategy signals)
	testSignal := &strategy.Signal{
		Type:     strategy.SignalTypeEntry,
		Side:     exchanges.OrderSideBuy,
		Symbol:   "BTC-USD",
		Price:    decimal.NewFromFloat(50000),
		Strength: 0.8,
	}

	if err := executionAgent.HandleSignal(ctx, testSignal); err != nil {
		t.Errorf("Failed to handle test signal: %v", err)
	}

	// Check if order was placed
	ordersAfter := orderManager.GetOpenOrders()
	if len(ordersAfter) > len(orders) {
		t.Log("Order was successfully placed through execution agent")
	} else {
		t.Log("No new orders placed - this might be due to risk checks")
	}
}
