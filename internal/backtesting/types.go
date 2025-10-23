package backtesting

import (
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// HistoricalData represents a collection of historical candles
type HistoricalData struct {
	Symbol  string
	Candles []exchanges.Candle
}

// Trade represents a backtesting trade execution
type Trade struct {
	ID          string
	Symbol      string
	Side        exchanges.OrderSide
	EntryPrice  decimal.Decimal
	ExitPrice   decimal.Decimal
	Amount      decimal.Decimal
	EntryTime   time.Time
	ExitTime    time.Time
	PnL         decimal.Decimal
	PnLPercent  decimal.Decimal
	Commission  decimal.Decimal
	StopLoss    decimal.Decimal
	TakeProfit  decimal.Decimal
	ExitReason  string // "stop_loss", "take_profit", "signal", "end_of_data"
}

// Position represents an open position during backtesting
type Position struct {
	Symbol      string
	Side        exchanges.OrderSide
	EntryPrice  decimal.Decimal
	Amount      decimal.Decimal
	EntryTime   time.Time
	StopLoss    decimal.Decimal
	TakeProfit  decimal.Decimal
}

// BacktestConfig holds configuration for backtesting
type BacktestConfig struct {
	// Capital
	InitialCapital  decimal.Decimal
	CommissionRate  decimal.Decimal // e.g., 0.001 for 0.1%
	Slippage        decimal.Decimal // e.g., 0.0005 for 0.05%

	// Position sizing
	UseFixedAmount  bool
	FixedAmount     decimal.Decimal
	RiskPerTrade    decimal.Decimal // e.g., 0.01 for 1% of capital

	// Constraints
	MaxPositions    int
	AllowShort      bool

	// Time range
	StartTime       time.Time
	EndTime         time.Time
}

// DefaultBacktestConfig returns default backtesting configuration
func DefaultBacktestConfig() *BacktestConfig {
	return &BacktestConfig{
		InitialCapital:  decimal.NewFromFloat(10000),
		CommissionRate:  decimal.NewFromFloat(0.001), // 0.1%
		Slippage:        decimal.NewFromFloat(0.0005), // 0.05%
		UseFixedAmount:  false,
		RiskPerTrade:    decimal.NewFromFloat(0.01), // 1%
		MaxPositions:    1,
		AllowShort:      false,
	}
}

// PerformanceMetrics contains backtesting results
type PerformanceMetrics struct {
	// Overall performance
	TotalReturn       decimal.Decimal
	TotalReturnPct    decimal.Decimal
	AnnualizedReturn  decimal.Decimal

	// Trade statistics
	TotalTrades       int
	WinningTrades     int
	LosingTrades      int
	WinRate           decimal.Decimal

	// Profit/Loss
	TotalProfit       decimal.Decimal
	TotalLoss         decimal.Decimal
	AverageProfitWin  decimal.Decimal
	AverageLossLose   decimal.Decimal
	LargestWin        decimal.Decimal
	LargestLoss       decimal.Decimal
	ProfitFactor      decimal.Decimal

	// Risk metrics
	MaxDrawdown       decimal.Decimal
	MaxDrawdownPct    decimal.Decimal
	SharpeRatio       decimal.Decimal

	// Time analysis
	AvgTradeDuration  time.Duration
	TotalDuration     time.Duration

	// Detailed records
	Trades            []Trade
	EquityCurve       []EquityPoint
}

// EquityPoint represents a point in the equity curve
type EquityPoint struct {
	Time   time.Time
	Equity decimal.Decimal
}
