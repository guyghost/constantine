package risk

import (
	"fmt"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/order"
	"github.com/shopspring/decimal"
)

// Config holds risk management configuration
type Config struct {
	MaxPositionSize      decimal.Decimal // Maximum position size per trade
	MaxPositions         int             // Maximum number of concurrent positions
	MaxLeverage          decimal.Decimal // Maximum leverage allowed
	MaxDailyLoss         decimal.Decimal // Maximum daily loss (in base currency)
	MaxDrawdown          decimal.Decimal // Maximum drawdown percentage
	RiskPerTrade         decimal.Decimal // Risk per trade as percentage of capital
	MinAccountBalance    decimal.Decimal // Minimum account balance to trade
	DailyTradingLimit    int             // Maximum trades per day
	CooldownPeriod       time.Duration   // Cooldown period after consecutive losses
	ConsecutiveLossLimit int             // Number of consecutive losses to trigger cooldown
}

// DefaultConfig returns default risk management configuration
func DefaultConfig() *Config {
	return &Config{
		MaxPositionSize:      decimal.NewFromFloat(1000),
		MaxPositions:         3,
		MaxLeverage:          decimal.NewFromInt(5),
		MaxDailyLoss:         decimal.NewFromFloat(100),
		MaxDrawdown:          decimal.NewFromFloat(10), // 10%
		RiskPerTrade:         decimal.NewFromFloat(1),  // 1%
		MinAccountBalance:    decimal.NewFromFloat(100),
		DailyTradingLimit:    50,
		CooldownPeriod:       15 * time.Minute,
		ConsecutiveLossLimit: 3,
	}
}

// Manager manages trading risk
type Manager struct {
	config *Config
	mu     sync.RWMutex

	// State tracking
	dailyPnL            decimal.Decimal
	consecutiveLosses   int
	tradesExecutedToday int
	lastTradeTime       time.Time
	cooldownUntil       time.Time
	startingBalance     decimal.Decimal
	currentBalance      decimal.Decimal
	peakBalance         decimal.Decimal
	tradeHistory        []TradeResult
	lastResetDate       time.Time
}

// TradeResult represents the result of a trade
type TradeResult struct {
	Timestamp  time.Time
	Symbol     string
	Side       exchanges.OrderSide
	EntryPrice decimal.Decimal
	ExitPrice  decimal.Decimal
	Amount     decimal.Decimal
	PnL        decimal.Decimal
	IsWin      bool
}

// NewManager creates a new risk manager
func NewManager(config *Config, initialBalance decimal.Decimal) *Manager {
	now := time.Now()
	return &Manager{
		config:          config,
		dailyPnL:        decimal.Zero,
		startingBalance: initialBalance,
		currentBalance:  initialBalance,
		peakBalance:     initialBalance,
		tradeHistory:    make([]TradeResult, 0),
		lastResetDate:   now,
		lastTradeTime:   now,
	}
}

// CanTrade checks if trading is allowed based on risk parameters
func (m *Manager) CanTrade() (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if in cooldown period
	if time.Now().Before(m.cooldownUntil) {
		remaining := time.Until(m.cooldownUntil)
		return false, fmt.Sprintf("in cooldown period, %v remaining", remaining.Round(time.Second))
	}

	// Check daily loss limit
	if m.dailyPnL.LessThan(m.config.MaxDailyLoss.Neg()) {
		return false, "daily loss limit reached"
	}

	// Check daily trade limit
	if m.tradesExecutedToday >= m.config.DailyTradingLimit {
		return false, "daily trade limit reached"
	}

	// Check minimum account balance
	if m.currentBalance.LessThan(m.config.MinAccountBalance) {
		return false, "account balance below minimum"
	}

	// Check maximum drawdown
	drawdown := m.calculateDrawdown()
	if drawdown.GreaterThan(m.config.MaxDrawdown) {
		return false, fmt.Sprintf("maximum drawdown exceeded: %.2f%%", drawdown)
	}

	return true, ""
}

// ValidateOrder validates an order against risk parameters
func (m *Manager) ValidateOrder(req *order.OrderRequest, openPositions []*order.ManagedPosition) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check max positions
	if len(openPositions) >= m.config.MaxPositions {
		return fmt.Errorf("maximum number of positions (%d) reached", m.config.MaxPositions)
	}

	// Check position size
	positionSize := req.Amount.Mul(req.Price)
	if positionSize.GreaterThan(m.config.MaxPositionSize) {
		return fmt.Errorf("position size %.2f exceeds maximum %.2f",
			positionSize, m.config.MaxPositionSize)
	}

	// Validate position size based on risk per trade
	maxRisk := m.currentBalance.Mul(m.config.RiskPerTrade).Div(decimal.NewFromInt(100))
	if positionSize.GreaterThan(maxRisk.Mul(decimal.NewFromInt(100))) {
		return fmt.Errorf("position size exceeds risk per trade limit")
	}

	// Check if stop loss is set
	if req.StopLoss.IsZero() {
		return fmt.Errorf("stop loss is required")
	}

	return nil
}

// CalculatePositionSize calculates the appropriate position size based on risk
func (m *Manager) CalculatePositionSize(
	entryPrice decimal.Decimal,
	stopLoss decimal.Decimal,
	accountBalance decimal.Decimal,
) decimal.Decimal {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate risk amount
	riskAmount := accountBalance.Mul(m.config.RiskPerTrade).Div(decimal.NewFromInt(100))

	// Calculate price difference
	priceDiff := entryPrice.Sub(stopLoss).Abs()
	if priceDiff.IsZero() {
		return decimal.Zero
	}

	// Calculate position size
	positionSize := riskAmount.Div(priceDiff)

	// Cap at max position size
	maxSize := m.config.MaxPositionSize.Div(entryPrice)
	if positionSize.GreaterThan(maxSize) {
		positionSize = maxSize
	}

	return positionSize
}

// RecordTrade records a trade result and updates statistics
func (m *Manager) RecordTrade(result TradeResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to trade history
	m.tradeHistory = append(m.tradeHistory, result)

	// Update daily PnL
	m.dailyPnL = m.dailyPnL.Add(result.PnL)

	// Update balance
	m.currentBalance = m.currentBalance.Add(result.PnL)

	// Update peak balance
	if m.currentBalance.GreaterThan(m.peakBalance) {
		m.peakBalance = m.currentBalance
	}

	// Update consecutive losses
	if result.IsWin {
		m.consecutiveLosses = 0
	} else {
		m.consecutiveLosses++

		// Check if cooldown should be triggered
		if m.consecutiveLosses >= m.config.ConsecutiveLossLimit {
			m.cooldownUntil = time.Now().Add(m.config.CooldownPeriod)
			m.consecutiveLosses = 0 // Reset after triggering cooldown
		}
	}

	// Update trade count
	m.tradesExecutedToday++
	m.lastTradeTime = time.Now()

	// Check if we need to reset daily statistics
	m.checkDailyReset()
}

// UpdateBalance updates the current account balance
func (m *Manager) UpdateBalance(balance decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentBalance = balance

	if balance.GreaterThan(m.peakBalance) {
		m.peakBalance = balance
	}
}

// GetCurrentBalance returns the current account balance
func (m *Manager) GetCurrentBalance() decimal.Decimal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentBalance
}

// GetDailyPnL returns the daily profit/loss
func (m *Manager) GetDailyPnL() decimal.Decimal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dailyPnL
}

// GetDailyTradeCount returns the number of trades executed today
func (m *Manager) GetDailyTradeCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tradesExecutedToday
}

// GetConsecutiveLosses returns the current consecutive loss count
func (m *Manager) GetConsecutiveLosses() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.consecutiveLosses
}

// IsInCooldown returns whether trading is in cooldown period
func (m *Manager) IsInCooldown() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Now().Before(m.cooldownUntil)
}

// GetCooldownRemaining returns the remaining cooldown time
func (m *Manager) GetCooldownRemaining() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if time.Now().Before(m.cooldownUntil) {
		return time.Until(m.cooldownUntil)
	}
	return 0
}

// calculateDrawdown calculates the current drawdown percentage
func (m *Manager) calculateDrawdown() decimal.Decimal {
	if m.peakBalance.IsZero() {
		return decimal.Zero
	}

	drawdown := m.peakBalance.Sub(m.currentBalance).Div(m.peakBalance).Mul(decimal.NewFromInt(100))
	return drawdown
}

// GetDrawdown returns the current drawdown percentage
func (m *Manager) GetDrawdown() decimal.Decimal {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.calculateDrawdown()
}

// checkDailyReset checks if daily statistics should be reset
func (m *Manager) checkDailyReset() {
	now := time.Now()
	if now.Day() != m.lastResetDate.Day() {
		m.dailyPnL = decimal.Zero
		m.tradesExecutedToday = 0
		m.lastResetDate = now
	}
}

// GetStats returns risk management statistics
func (m *Manager) GetStats() *Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalTrades := len(m.tradeHistory)
	winningTrades := 0
	losingTrades := 0
	totalProfit := decimal.Zero
	totalLoss := decimal.Zero

	for _, trade := range m.tradeHistory {
		if trade.IsWin {
			winningTrades++
			totalProfit = totalProfit.Add(trade.PnL)
		} else {
			losingTrades++
			totalLoss = totalLoss.Add(trade.PnL.Abs())
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades) * 100
	}

	profitFactor := 0.0
	if !totalLoss.IsZero() {
		pf, _ := totalProfit.Div(totalLoss.Abs()).Float64()
		profitFactor = pf
	}

	return &Stats{
		TotalTrades:         totalTrades,
		WinningTrades:       winningTrades,
		LosingTrades:        losingTrades,
		WinRate:             winRate,
		TotalProfit:         totalProfit,
		TotalLoss:           totalLoss,
		NetPnL:              totalProfit.Add(totalLoss),
		ProfitFactor:        profitFactor,
		CurrentDrawdown:     m.calculateDrawdown(),
		ConsecutiveLosses:   m.consecutiveLosses,
		DailyPnL:            m.dailyPnL,
		TradesExecutedToday: m.tradesExecutedToday,
		CurrentBalance:      m.currentBalance,
		StartingBalance:     m.startingBalance,
		PeakBalance:         m.peakBalance,
	}
}

// Stats holds risk management statistics
type Stats struct {
	TotalTrades         int
	WinningTrades       int
	LosingTrades        int
	WinRate             float64
	TotalProfit         decimal.Decimal
	TotalLoss           decimal.Decimal
	NetPnL              decimal.Decimal
	ProfitFactor        float64
	CurrentDrawdown     decimal.Decimal
	ConsecutiveLosses   int
	DailyPnL            decimal.Decimal
	TradesExecutedToday int
	CurrentBalance      decimal.Decimal
	StartingBalance     decimal.Decimal
	PeakBalance         decimal.Decimal
}
