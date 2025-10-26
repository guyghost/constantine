package risk

import (
	"os"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/order"
	"github.com/shopspring/decimal"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	if !config.MaxPositionSize.Equal(decimal.NewFromFloat(1000)) {
		t.Errorf("expected MaxPositionSize 1000, got %s", config.MaxPositionSize)
	}

	if config.MaxPositions != 3 {
		t.Errorf("expected MaxPositions 3, got %d", config.MaxPositions)
	}

	if !config.MaxDailyLoss.Equal(decimal.NewFromFloat(100)) {
		t.Errorf("expected MaxDailyLoss 100, got %s", config.MaxDailyLoss)
	}
}

func TestNewManager(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)

	manager := NewManager(config, initialBalance)

	if manager == nil {
		t.Fatal("NewManager should not return nil")
	}

	if manager.config != config {
		t.Error("manager config should match provided config")
	}

	if !manager.currentBalance.Equal(initialBalance) {
		t.Errorf("expected current balance %s, got %s", initialBalance, manager.currentBalance)
	}

	if !manager.startingBalance.Equal(initialBalance) {
		t.Errorf("expected initial balance %s, got %s", initialBalance, manager.startingBalance)
	}
}

func TestManager_CanTrade(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	// Should be able to trade initially
	canTrade, reason := manager.CanTrade()
	if !canTrade {
		t.Errorf("should be able to trade initially, but got: %s", reason)
	}

	// Test with insufficient balance
	manager.UpdateBalance(decimal.NewFromFloat(50)) // Below MinAccountBalance
	canTrade, reason = manager.CanTrade()
	if canTrade {
		t.Error("should not be able to trade with insufficient balance")
	}
	if reason == "" {
		t.Error("should provide reason for not being able to trade")
	}

	// Reset balance
	manager.UpdateBalance(decimal.NewFromFloat(10000))
	canTrade, reason = manager.CanTrade()
	if !canTrade {
		t.Errorf("should be able to trade after balance reset, but got: %s", reason)
	}
}

func TestManager_CalculatePositionSize(t *testing.T) {
	config := DefaultConfig()
	config.RiskPerTrade = decimal.NewFromFloat(1) // 1% risk per trade
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	entryPrice := decimal.NewFromFloat(50000)
	stopLoss := decimal.NewFromFloat(49500) // 1% stop loss

	positionSize := manager.CalculatePositionSize(entryPrice, stopLoss, initialBalance)

	// Risk amount = 1% of 10000 = 100
	// Price diff = |50000 - 49500| = 500
	// Position size = 100 / 500 = 0.2
	// But capped at MaxPositionSize / entryPrice = 1000 / 50000 = 0.02
	expectedSize := decimal.NewFromFloat(0.02)
	if !positionSize.Equal(expectedSize) {
		t.Errorf("expected position size %s, got %s", expectedSize, positionSize)
	}

	// Test with stop loss above entry (for buy position, stop loss should be below entry)
	stopLossAbove := decimal.NewFromFloat(50500)
	positionSize = manager.CalculatePositionSize(entryPrice, stopLossAbove, initialBalance)
	expectedSizeAbove := decimal.NewFromFloat(0.02) // Same cap applies: 1000 / 50000 = 0.02
	if !positionSize.Equal(expectedSizeAbove) {
		t.Errorf("expected position size %s for stop loss above entry, got %s", expectedSizeAbove, positionSize)
	}
}

func TestManager_ValidateOrder(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	// Valid order with smaller amount
	req := &order.OrderRequest{
		Symbol:   "BTC-USD",
		Side:     exchanges.OrderSideBuy,
		Type:     exchanges.OrderTypeLimit,
		Price:    decimal.NewFromFloat(50000),
		Amount:   decimal.NewFromFloat(0.01), // Smaller amount to stay under MaxPositionSize
		StopLoss: decimal.NewFromFloat(49500),
	}

	openPositions := []*order.ManagedPosition{}
	err := manager.ValidateOrder(req, openPositions)
	if err != nil {
		t.Errorf("valid order should pass validation, got error: %v", err)
	}

	// Test with position size too large
	req.Amount = decimal.NewFromFloat(1.0) // 1.0 * 50000 = 50000 > 1000 MaxPositionSize
	err = manager.ValidateOrder(req, openPositions)
	if err == nil {
		t.Error("order with position size too large should fail validation")
	}

	// Reset amount
	req.Amount = decimal.NewFromFloat(0.1)

	// Test with too many positions
	config.MaxPositions = 1
	openPositions = []*order.ManagedPosition{
		{Symbol: "ETH-USD", Side: "buy", Amount: decimal.NewFromFloat(1)},
	}
	err = manager.ValidateOrder(req, openPositions)
	if err == nil {
		t.Error("order exceeding max positions should fail validation")
	}
}

func TestManager_RecordTrade(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	// Record a winning trade
	winningTrade := TradeResult{
		PnL:   decimal.NewFromFloat(100),
		IsWin: true,
	}

	manager.RecordTrade(winningTrade)

	if manager.tradesExecutedToday != 1 {
		t.Errorf("expected daily trade count 1, got %d", manager.tradesExecutedToday)
	}

	if !manager.dailyPnL.Equal(decimal.NewFromFloat(100)) {
		t.Errorf("expected daily PnL 100, got %s", manager.dailyPnL)
	}

	if manager.consecutiveLosses != 0 {
		t.Errorf("expected consecutive losses 0, got %d", manager.consecutiveLosses)
	}

	// Record a losing trade
	losingTrade := TradeResult{
		PnL:   decimal.NewFromFloat(-50),
		IsWin: false,
	}

	manager.RecordTrade(losingTrade)

	if manager.tradesExecutedToday != 2 {
		t.Errorf("expected daily trade count 2, got %d", manager.tradesExecutedToday)
	}

	if !manager.dailyPnL.Equal(decimal.NewFromFloat(50)) {
		t.Errorf("expected daily PnL 50, got %s", manager.dailyPnL)
	}

	if manager.consecutiveLosses != 1 {
		t.Errorf("expected consecutive losses 1, got %d", manager.consecutiveLosses)
	}
}

func TestManager_UpdateBalance(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	newBalance := decimal.NewFromFloat(10500)
	manager.UpdateBalance(newBalance)

	if !manager.GetCurrentBalance().Equal(newBalance) {
		t.Errorf("expected current balance %s, got %s", newBalance, manager.GetCurrentBalance())
	}
}

func TestManager_Getters(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	// Test GetCurrentBalance
	if !manager.GetCurrentBalance().Equal(initialBalance) {
		t.Errorf("expected current balance %s, got %s", initialBalance, manager.GetCurrentBalance())
	}

	// Test GetDailyPnL (initially zero)
	if !manager.GetDailyPnL().IsZero() {
		t.Errorf("expected daily PnL 0, got %s", manager.GetDailyPnL())
	}

	// Test GetDailyTradeCount (initially zero)
	if manager.GetDailyTradeCount() != 0 {
		t.Errorf("expected daily trade count 0, got %d", manager.GetDailyTradeCount())
	}

	// Test GetConsecutiveLosses (initially zero)
	if manager.GetConsecutiveLosses() != 0 {
		t.Errorf("expected consecutive losses 0, got %d", manager.GetConsecutiveLosses())
	}
}

func TestManager_Cooldown(t *testing.T) {
	config := DefaultConfig()
	config.ConsecutiveLossLimit = 2
	config.CooldownPeriod = 1 * time.Minute
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	// Record consecutive losses to trigger cooldown
	losingTrade := TradeResult{
		PnL:   decimal.NewFromFloat(-50),
		IsWin: false,
	}

	manager.RecordTrade(losingTrade)
	manager.RecordTrade(losingTrade) // Second consecutive loss

	// Should be in cooldown
	if !manager.IsInCooldown() {
		t.Error("should be in cooldown after consecutive losses")
	}

	remaining := manager.GetCooldownRemaining()
	if remaining <= 0 {
		t.Error("should have cooldown time remaining")
	}

	// Should not be able to trade during cooldown
	canTrade, reason := manager.CanTrade()
	if canTrade {
		t.Error("should not be able to trade during cooldown")
	}
	if reason == "" {
		t.Error("should provide reason for not being able to trade during cooldown")
	}
}

func TestManager_GetStats(t *testing.T) {
	config := DefaultConfig()
	initialBalance := decimal.NewFromFloat(10000)
	manager := NewManager(config, initialBalance)

	stats := manager.GetStats()

	if stats == nil {
		t.Fatal("GetStats should not return nil")
	}

	if !stats.CurrentBalance.Equal(initialBalance) {
		t.Errorf("expected current balance %s, got %s", initialBalance, stats.CurrentBalance)
	}

	if !stats.DailyPnL.IsZero() {
		t.Errorf("expected daily PnL 0, got %s", stats.DailyPnL)
	}

	if stats.TotalTrades != 0 {
		t.Errorf("expected total trades 0, got %d", stats.TotalTrades)
	}
}

func TestLoadConfig(t *testing.T) {
	// Test with default values
	config := LoadConfig()
	if config.MinAccountBalance.String() != "100" {
		t.Errorf("Expected default MinAccountBalance to be 100, got %s", config.MinAccountBalance.String())
	}

	// Test with environment variable override
	originalValue := os.Getenv("RISK_MIN_ACCOUNT_BALANCE")
	defer func() {
		if originalValue != "" {
			os.Setenv("RISK_MIN_ACCOUNT_BALANCE", originalValue)
		} else {
			os.Unsetenv("RISK_MIN_ACCOUNT_BALANCE")
		}
	}()

	os.Setenv("RISK_MIN_ACCOUNT_BALANCE", "50")
	config = LoadConfig()
	if config.MinAccountBalance.String() != "50" {
		t.Errorf("Expected MinAccountBalance to be 50 after env override, got %s", config.MinAccountBalance.String())
	}

	// Test integer parsing
	os.Setenv("RISK_MAX_POSITIONS", "5")
	config = LoadConfig()
	if config.MaxPositions != 5 {
		t.Errorf("Expected MaxPositions to be 5, got %d", config.MaxPositions)
	}

	// Test decimal parsing
	os.Setenv("RISK_MAX_POSITION_SIZE", "2000")
	config = LoadConfig()
	if config.MaxPositionSize.String() != "2000" {
		t.Errorf("Expected MaxPositionSize to be 2000, got %s", config.MaxPositionSize.String())
	}

	// Clean up
	os.Unsetenv("RISK_MIN_ACCOUNT_BALANCE")
	os.Unsetenv("RISK_MAX_POSITIONS")
	os.Unsetenv("RISK_MAX_POSITION_SIZE")
}
