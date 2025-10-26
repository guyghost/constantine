package backtesting

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/logger"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
)

// Engine is the backtesting engine
type Engine struct {
	config   *BacktestConfig
	data     *HistoricalData
	strategy *strategy.ScalpingStrategy
	exchange *SimulatedExchange

	// State
	currentIndex int
	capital      decimal.Decimal
	position     *Position
	trades       []Trade
	equityCurve  []EquityPoint

	// Callbacks
	onTrade        func(*Trade)
	onEquityUpdate func(decimal.Decimal)
}

// NewEngine creates a new backtesting engine
func NewEngine(config *BacktestConfig, data *HistoricalData) *Engine {
	return &Engine{
		config:      config,
		data:        data,
		capital:     config.InitialCapital,
		trades:      make([]Trade, 0),
		equityCurve: make([]EquityPoint, 0),
	}
}

// SetOnTrade sets the callback for trade execution
func (e *Engine) SetOnTrade(callback func(*Trade)) {
	e.onTrade = callback
}

// SetOnEquityUpdate sets the callback for equity updates
func (e *Engine) SetOnEquityUpdate(callback func(decimal.Decimal)) {
	e.onEquityUpdate = callback
}

// Run executes the backtest
func (e *Engine) Run(strategyConfig *strategy.Config) (*PerformanceMetrics, error) {
	if len(e.data.Candles) == 0 {
		return nil, fmt.Errorf("no historical data to backtest")
	}

	// Create simulated exchange
	e.exchange = NewSimulatedExchange(e.data, e.config)

	// Create strategy with simulated exchange
	e.strategy = strategy.NewScalpingStrategy(strategyConfig, e.exchange)

	// Set up strategy callbacks
	e.setupStrategyCallbacks()

	// Initialize equity curve
	e.recordEquity(e.data.Candles[0].Timestamp)

	// Run through historical data
	ctx := context.Background()

	for e.currentIndex = 0; e.currentIndex < len(e.data.Candles); e.currentIndex++ {
		candle := e.data.Candles[e.currentIndex]

		// Update simulated exchange state
		e.exchange.SetCurrentCandle(e.currentIndex)

		// Check if position should be closed (stop loss / take profit)
		e.checkPositionExit(candle)

		// Feed candle to strategy (this will trigger signals via callback)
		e.feedCandleToStrategy(ctx, candle)

		// Record equity
		e.recordEquity(candle.Timestamp)
	}

	// Close any remaining positions
	if e.position != nil {
		e.closePosition(e.data.Candles[len(e.data.Candles)-1], "end_of_data")
	}

	// Calculate performance metrics
	metrics := e.calculateMetrics()

	return metrics, nil
}

// setupStrategyCallbacks sets up callbacks for the strategy
func (e *Engine) setupStrategyCallbacks() {
	e.strategy.SetSignalCallback(func(signal *strategy.Signal) {
		e.handleSignal(signal)
	})

	e.strategy.SetErrorCallback(func(err error) {
		logger.Component("backtesting").WithError(err).Error("Strategy error occurred")
	})
}

// feedCandleToStrategy feeds a candle to the strategy for analysis
func (e *Engine) feedCandleToStrategy(ctx context.Context, candle exchanges.Candle) {
	// The strategy will analyze the candle and generate signals
	// In a real implementation, we'd need to adapt the strategy to accept candles directly
	// For now, we'll simulate by calling the signal generator

	// Need enough historical data for indicators
	minDataPoints := 25 // Need at least LongEMAPeriod + some buffer
	if e.currentIndex < minDataPoints {
		return // Not enough data yet
	}

	// Get current candles window for analysis
	windowSize := 50 // Should be at least max(LongEMAPeriod, RSIPeriod, BB Period)
	start := e.currentIndex - windowSize + 1
	if start < 0 {
		start = 0
	}

	candles := e.data.Candles[start : e.currentIndex+1]

	// Extract prices and volumes
	prices := make([]decimal.Decimal, len(candles))
	volumes := make([]decimal.Decimal, len(candles))
	for i := range candles {
		prices[i] = candles[i].Close
		volumes[i] = candles[i].Volume
	}

	// Generate signal from candles
	signal := e.strategy.GetSignalGenerator().GenerateSignal(e.data.Symbol, prices, volumes, nil)

	if signal != nil && signal.Type != strategy.SignalTypeNone {
		e.handleSignal(signal)
	}
}

// handleSignal handles trading signals from the strategy
func (e *Engine) handleSignal(signal *strategy.Signal) {
	candle := e.data.Candles[e.currentIndex]

	// Entry signals
	if signal.Type == strategy.SignalTypeEntry {
		if e.position == nil && signal.Strength > 0.1 {
			e.openPosition(signal, candle)
		}
	}

	// Exit signals
	if signal.Type == strategy.SignalTypeExit {
		if e.position != nil && e.position.Side == signal.Side {
			e.closePosition(candle, "signal")
		}
	}
}

// openPosition opens a new position
func (e *Engine) openPosition(signal *strategy.Signal, candle exchanges.Candle) {
	// Check if we can open a position
	if e.config.MaxPositions > 0 && e.position != nil {
		return // Already have a position
	}

	if !e.config.AllowShort && signal.Side == exchanges.OrderSideSell {
		return // Short selling not allowed
	}

	// Calculate stop loss and take profit based on strategy configuration
	// These values are now pulled from the strategy config instead of being hardcoded
	stopLossPercent := decimal.NewFromFloat(e.strategy.GetConfig().StopLossPercent)
	takeProfitPercent := decimal.NewFromFloat(e.strategy.GetConfig().TakeProfitPercent)

	var stopLoss, takeProfit decimal.Decimal
	if signal.Side == exchanges.OrderSideBuy {
		stopLoss = signal.Price.Mul(decimal.NewFromInt(1).Sub(stopLossPercent))
		takeProfit = signal.Price.Mul(decimal.NewFromInt(1).Add(takeProfitPercent))
	} else {
		stopLoss = signal.Price.Mul(decimal.NewFromInt(1).Add(stopLossPercent))
		takeProfit = signal.Price.Mul(decimal.NewFromInt(1).Sub(takeProfitPercent))
	}

	// Calculate position size
	var amount decimal.Decimal
	if e.config.UseFixedAmount {
		amount = e.config.FixedAmount
	} else {
		// Risk-based position sizing
		riskAmount := e.capital.Mul(e.config.RiskPerTrade)
		stopDistance := signal.Price.Sub(stopLoss).Abs()
		if stopDistance.IsZero() {
			return
		}
		amount = riskAmount.Div(stopDistance)
	}

	// Apply slippage to entry price
	entryPrice := signal.Price
	if signal.Side == exchanges.OrderSideBuy {
		entryPrice = entryPrice.Mul(decimal.NewFromInt(1).Add(e.config.Slippage))
	} else {
		entryPrice = entryPrice.Mul(decimal.NewFromInt(1).Sub(e.config.Slippage))
	}

	// Calculate commission
	commission := entryPrice.Mul(amount).Mul(e.config.CommissionRate)

	// Check if we have enough capital
	requiredCapital := entryPrice.Mul(amount).Add(commission)
	if requiredCapital.GreaterThan(e.capital) {
		return // Not enough capital
	}

	// Open position
	e.position = &Position{
		Symbol:     signal.Symbol,
		Side:       signal.Side,
		EntryPrice: entryPrice,
		Amount:     amount,
		EntryTime:  candle.Timestamp,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	}

	// Deduct capital
	e.capital = e.capital.Sub(commission)
}

// closePosition closes the current position
func (e *Engine) closePosition(candle exchanges.Candle, reason string) {
	if e.position == nil {
		return
	}

	// Apply slippage to exit price
	exitPrice := candle.Close
	if e.position.Side == exchanges.OrderSideBuy {
		exitPrice = exitPrice.Mul(decimal.NewFromInt(1).Sub(e.config.Slippage))
	} else {
		exitPrice = exitPrice.Mul(decimal.NewFromInt(1).Add(e.config.Slippage))
	}

	// Calculate P&L
	var pnl decimal.Decimal
	if e.position.Side == exchanges.OrderSideBuy {
		pnl = exitPrice.Sub(e.position.EntryPrice).Mul(e.position.Amount)
	} else {
		pnl = e.position.EntryPrice.Sub(exitPrice).Mul(e.position.Amount)
	}

	// Calculate commission
	commission := exitPrice.Mul(e.position.Amount).Mul(e.config.CommissionRate)
	pnl = pnl.Sub(commission)

	// Calculate P&L percentage
	pnlPercent := pnl.Div(e.position.EntryPrice.Mul(e.position.Amount)).Mul(decimal.NewFromInt(100))

	// Create trade record
	trade := Trade{
		ID:         uuid.New().String(),
		Symbol:     e.position.Symbol,
		Side:       e.position.Side,
		EntryPrice: e.position.EntryPrice,
		ExitPrice:  exitPrice,
		Amount:     e.position.Amount,
		EntryTime:  e.position.EntryTime,
		ExitTime:   candle.Timestamp,
		PnL:        pnl,
		PnLPercent: pnlPercent,
		Commission: commission.Mul(decimal.NewFromInt(2)), // Entry + Exit
		StopLoss:   e.position.StopLoss,
		TakeProfit: e.position.TakeProfit,
		ExitReason: reason,
	}

	e.trades = append(e.trades, trade)

	// Update capital
	e.capital = e.capital.Add(pnl)

	// Callback
	if e.onTrade != nil {
		e.onTrade(&trade)
	}

	// Clear position
	e.position = nil
}

// checkPositionExit checks if position should be exited due to stop loss or take profit
func (e *Engine) checkPositionExit(candle exchanges.Candle) {
	if e.position == nil {
		return
	}

	// Check stop loss
	if e.position.Side == exchanges.OrderSideBuy {
		if candle.Low.LessThanOrEqual(e.position.StopLoss) {
			e.closePosition(candle, "stop_loss")
			return
		}
		if candle.High.GreaterThanOrEqual(e.position.TakeProfit) {
			e.closePosition(candle, "take_profit")
			return
		}
	} else {
		if candle.High.GreaterThanOrEqual(e.position.StopLoss) {
			e.closePosition(candle, "stop_loss")
			return
		}
		if candle.Low.LessThanOrEqual(e.position.TakeProfit) {
			e.closePosition(candle, "take_profit")
			return
		}
	}
}

// recordEquity records the current equity in the equity curve
func (e *Engine) recordEquity(timestamp time.Time) {
	equity := e.capital

	// Add unrealized P&L from open position
	if e.position != nil {
		candle := e.data.Candles[e.currentIndex]
		var unrealizedPnL decimal.Decimal
		if e.position.Side == exchanges.OrderSideBuy {
			unrealizedPnL = candle.Close.Sub(e.position.EntryPrice).Mul(e.position.Amount)
		} else {
			unrealizedPnL = e.position.EntryPrice.Sub(candle.Close).Mul(e.position.Amount)
		}
		equity = equity.Add(unrealizedPnL)
	}

	e.equityCurve = append(e.equityCurve, EquityPoint{
		Time:   timestamp,
		Equity: equity,
	})

	if e.onEquityUpdate != nil {
		e.onEquityUpdate(equity)
	}
}

// calculateMetrics calculates performance metrics from the backtest results
func (e *Engine) calculateMetrics() *PerformanceMetrics {
	metrics := &PerformanceMetrics{
		Trades:      e.trades,
		EquityCurve: e.equityCurve,
		TotalTrades: len(e.trades),
	}

	if len(e.trades) == 0 {
		return metrics
	}

	// Calculate returns
	finalEquity := e.capital
	totalReturn := finalEquity.Sub(e.config.InitialCapital)
	metrics.TotalReturn = totalReturn
	metrics.TotalReturnPct = totalReturn.Div(e.config.InitialCapital).Mul(decimal.NewFromInt(100))

	// Calculate win/loss statistics
	var totalProfit, totalLoss decimal.Decimal
	var largestWin, largestLoss decimal.Decimal
	var totalDuration time.Duration

	for _, trade := range e.trades {
		duration := trade.ExitTime.Sub(trade.EntryTime)
		totalDuration += duration

		if trade.PnL.GreaterThan(decimal.Zero) {
			metrics.WinningTrades++
			totalProfit = totalProfit.Add(trade.PnL)
			if trade.PnL.GreaterThan(largestWin) {
				largestWin = trade.PnL
			}
		} else {
			metrics.LosingTrades++
			totalLoss = totalLoss.Add(trade.PnL.Abs())
			if trade.PnL.Abs().GreaterThan(largestLoss) {
				largestLoss = trade.PnL.Abs()
			}
		}
	}

	metrics.TotalProfit = totalProfit
	metrics.TotalLoss = totalLoss
	metrics.LargestWin = largestWin
	metrics.LargestLoss = largestLoss

	if metrics.TotalTrades > 0 {
		metrics.WinRate = decimal.NewFromInt(int64(metrics.WinningTrades)).Div(decimal.NewFromInt(int64(metrics.TotalTrades))).Mul(decimal.NewFromInt(100))
		metrics.AvgTradeDuration = totalDuration / time.Duration(metrics.TotalTrades)
	}

	if metrics.WinningTrades > 0 {
		metrics.AverageProfitWin = totalProfit.Div(decimal.NewFromInt(int64(metrics.WinningTrades)))
	}

	if metrics.LosingTrades > 0 {
		metrics.AverageLossLose = totalLoss.Div(decimal.NewFromInt(int64(metrics.LosingTrades)))
	}

	if !totalLoss.IsZero() {
		metrics.ProfitFactor = totalProfit.Div(totalLoss)
	}

	// Calculate max drawdown
	metrics.MaxDrawdown, metrics.MaxDrawdownPct = e.calculateMaxDrawdown()

	// Calculate annualized return
	if len(e.data.Candles) > 0 {
		startTime := e.data.Candles[0].Timestamp
		endTime := e.data.Candles[len(e.data.Candles)-1].Timestamp
		years := endTime.Sub(startTime).Hours() / 24 / 365.25
		if years > 0 {
			metrics.AnnualizedReturn = metrics.TotalReturnPct.Div(decimal.NewFromFloat(years))
		}
		metrics.TotalDuration = endTime.Sub(startTime)
	}

	return metrics
}

// calculateMaxDrawdown calculates the maximum drawdown
func (e *Engine) calculateMaxDrawdown() (decimal.Decimal, decimal.Decimal) {
	var maxDrawdown, maxDrawdownPct decimal.Decimal
	peak := e.config.InitialCapital

	for _, point := range e.equityCurve {
		if point.Equity.GreaterThan(peak) {
			peak = point.Equity
		}

		drawdown := peak.Sub(point.Equity)
		if drawdown.GreaterThan(maxDrawdown) {
			maxDrawdown = drawdown
			if !peak.IsZero() {
				maxDrawdownPct = drawdown.Div(peak).Mul(decimal.NewFromInt(100))
			}
		}
	}

	return maxDrawdown, maxDrawdownPct
}
