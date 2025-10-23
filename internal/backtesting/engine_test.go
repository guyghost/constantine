package backtesting

import (
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/guyghost/constantine/internal/testutils"
	"github.com/shopspring/decimal"
)

func TestNewEngine(t *testing.T) {
	config := DefaultBacktestConfig()
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:10], // Use first 10 candles
	}

	engine := NewEngine(config, data)

	testutils.AssertNotNil(t, engine, "Engine should not be nil")
	testutils.AssertEqual(t, config, engine.config, "Config should match")
	testutils.AssertEqual(t, data, engine.data, "Data should match")
	testutils.AssertTrue(t, engine.capital.Equal(config.InitialCapital), "Capital should match initial capital")
	testutils.AssertEqual(t, 0, len(engine.trades), "Trades should be empty initially")
	testutils.AssertEqual(t, 0, len(engine.equityCurve), "Equity curve should be empty initially")
}

func TestEngine_Run(t *testing.T) {
	config := DefaultBacktestConfig()
	config.InitialCapital = decimal.NewFromFloat(1000) // Smaller capital for testing

	// Create sample data
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:50], // Use first 50 candles
	}

	engine := NewEngine(config, data)

	// Create strategy config
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = "BTC-USD"

	// Run backtest
	metrics, err := engine.Run(strategyConfig)
	testutils.AssertNoError(t, err, "Run should not return error")
	testutils.AssertNotNil(t, metrics, "Metrics should not be nil")

	// Basic checks
	testutils.AssertTrue(t, metrics.TotalTrades >= 0, "Total trades should be non-negative")
	testutils.AssertTrue(t, len(metrics.Trades) == metrics.TotalTrades, "Trades slice length should match TotalTrades")
	testutils.AssertTrue(t, len(metrics.EquityCurve) > 0, "Equity curve should not be empty")

	// Check that equity curve starts with initial capital
	if len(metrics.EquityCurve) > 0 {
		testutils.AssertTrue(t, metrics.EquityCurve[0].Equity.Equal(config.InitialCapital),
			"Equity curve should start with initial capital")
	}
}

func TestEngine_Run_NoData(t *testing.T) {
	config := DefaultBacktestConfig()
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: []exchanges.Candle{}, // Empty data
	}

	engine := NewEngine(config, data)
	strategyConfig := strategy.DefaultConfig()

	_, err := engine.Run(strategyConfig)
	testutils.AssertError(t, err, "Run should return error for empty data")
}

func TestEngine_Callbacks(t *testing.T) {
	config := DefaultBacktestConfig()
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:10],
	}

	engine := NewEngine(config, data)

	// Test trade callback
	engine.SetOnTrade(func(trade *Trade) {
		// Trade callback set
	})

	// Test equity update callback
	engine.SetOnEquityUpdate(func(equity decimal.Decimal) {
		// Equity callback set
	})

	// Run a short backtest to trigger callbacks
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = "BTC-USD"

	_, err := engine.Run(strategyConfig)
	testutils.AssertNoError(t, err, "Run should not return error")

	// Note: Callbacks may or may not be triggered depending on strategy signals
	// We just verify the callbacks are set properly
	testutils.AssertNotNil(t, engine.onTrade, "Trade callback should be set")
	testutils.AssertNotNil(t, engine.onEquityUpdate, "Equity callback should be set")
}

func TestEngine_OpenPosition(t *testing.T) {
	config := DefaultBacktestConfig()
	config.InitialCapital = decimal.NewFromFloat(100000) // Increase capital for position sizing
	config.UseFixedAmount = true
	config.FixedAmount = decimal.NewFromFloat(0.1)

	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:5],
	}

	engine := NewEngine(config, data)

	// Create a test signal
	signal := &strategy.Signal{
		Type:     strategy.SignalTypeEntry,
		Side:     exchanges.OrderSideBuy,
		Symbol:   "BTC-USD",
		Price:    decimal.NewFromFloat(50000),
		Strength: 0.8,
	}

	candle := data.Candles[0]

	// Open position
	engine.openPosition(signal, candle)

	// Check that position was opened
	testutils.AssertNotNil(t, engine.position, "Position should be opened")
	testutils.AssertEqual(t, "BTC-USD", engine.position.Symbol, "Position symbol should match")
	testutils.AssertEqual(t, exchanges.OrderSideBuy, engine.position.Side, "Position side should match")
	// Entry price should be signal price with slippage applied
	expectedEntryPrice := signal.Price.Mul(decimal.NewFromInt(1).Add(config.Slippage))
	testutils.AssertTrue(t, engine.position.EntryPrice.Equal(expectedEntryPrice), "Entry price should include slippage")
}

func TestEngine_ClosePosition(t *testing.T) {
	config := DefaultBacktestConfig()
	config.InitialCapital = decimal.NewFromFloat(100000) // Same as open position test
	config.UseFixedAmount = true
	config.FixedAmount = decimal.NewFromFloat(0.1)

	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:5],
	}

	engine := NewEngine(config, data)

	// First open a position
	signal := &strategy.Signal{
		Type:     strategy.SignalTypeEntry,
		Side:     exchanges.OrderSideBuy,
		Symbol:   "BTC-USD",
		Price:    decimal.NewFromFloat(50000),
		Strength: 0.8,
	}

	engine.openPosition(signal, data.Candles[0])
	testutils.AssertNotNil(t, engine.position, "Position should be opened")

	// Close position
	candle := data.Candles[1]
	engine.closePosition(candle, "test")

	// Check that position was closed
	if engine.position != nil {
		t.Errorf("Position should be closed, got %+v", engine.position)
	}
	testutils.AssertEqual(t, 1, len(engine.trades), "Should have 1 trade recorded")
}

func TestEngine_RecordEquity(t *testing.T) {
	config := DefaultBacktestConfig()
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:5],
	}

	engine := NewEngine(config, data)

	initialEquityPoints := len(engine.equityCurve)

	timestamp := time.Now()
	engine.recordEquity(timestamp)

	testutils.AssertEqual(t, initialEquityPoints+1, len(engine.equityCurve), "Equity curve should have one more point")

	point := engine.equityCurve[len(engine.equityCurve)-1]
	testutils.AssertEqual(t, timestamp.Unix(), point.Time.Unix(), "Timestamp should match")
	testutils.AssertTrue(t, point.Equity.Equal(engine.capital), "Equity should match current capital")
}

func TestEngine_CalculateMetrics_NoTrades(t *testing.T) {
	config := DefaultBacktestConfig()
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:5],
	}

	engine := NewEngine(config, data)

	metrics := engine.calculateMetrics()

	testutils.AssertNotNil(t, metrics, "Metrics should not be nil")
	testutils.AssertEqual(t, 0, metrics.TotalTrades, "Total trades should be 0")
	testutils.AssertEqual(t, 0, metrics.WinningTrades, "Winning trades should be 0")
	testutils.AssertEqual(t, 0, metrics.LosingTrades, "Losing trades should be 0")
	testutils.AssertTrue(t, metrics.TotalReturn.IsZero(), "Total return should be 0")
	testutils.AssertTrue(t, metrics.TotalProfit.IsZero(), "Total profit should be 0")
	testutils.AssertTrue(t, metrics.TotalLoss.IsZero(), "Total loss should be 0")
}

func TestEngine_CalculateMetrics_WithTrades(t *testing.T) {
	config := DefaultBacktestConfig()
	data := &HistoricalData{
		Symbol:  "BTC-USD",
		Candles: testutils.SampleCandles()[:10],
	}

	engine := NewEngine(config, data)

	// Add some test trades
	winningTrade := Trade{
		ID:         "trade1",
		Symbol:     "BTC-USD",
		Side:       exchanges.OrderSideBuy,
		EntryPrice: decimal.NewFromFloat(50000),
		ExitPrice:  decimal.NewFromFloat(51000),
		Amount:     decimal.NewFromFloat(0.1),
		EntryTime:  time.Now().Add(-time.Hour),
		ExitTime:   time.Now(),
		PnL:        decimal.NewFromFloat(100),
		Commission: decimal.NewFromFloat(1),
	}

	losingTrade := Trade{
		ID:         "trade2",
		Symbol:     "BTC-USD",
		Side:       exchanges.OrderSideBuy,
		EntryPrice: decimal.NewFromFloat(51000),
		ExitPrice:  decimal.NewFromFloat(50500),
		Amount:     decimal.NewFromFloat(0.1),
		EntryTime:  time.Now().Add(-30 * time.Minute),
		ExitTime:   time.Now().Add(-15 * time.Minute),
		PnL:        decimal.NewFromFloat(-50),
		Commission: decimal.NewFromFloat(1),
	}

	engine.trades = []Trade{winningTrade, losingTrade}

	metrics := engine.calculateMetrics()

	testutils.AssertEqual(t, 2, metrics.TotalTrades, "Total trades should be 2")
	testutils.AssertEqual(t, 1, metrics.WinningTrades, "Winning trades should be 1")
	testutils.AssertEqual(t, 1, metrics.LosingTrades, "Losing trades should be 1")
	testutils.AssertTrue(t, metrics.TotalProfit.Equal(decimal.NewFromFloat(100)), "Total profit should be 100")
	testutils.AssertTrue(t, metrics.TotalLoss.Equal(decimal.NewFromFloat(50)), "Total loss should be 50")
	testutils.AssertEqual(t, 0.5, metrics.WinRate.Div(decimal.NewFromInt(100)).InexactFloat64(), "Win rate should be 0.5")
}

func TestEngine_Integration_FullBacktest(t *testing.T) {
	// Integration test: Load data, run full backtest, verify results
	config := DefaultBacktestConfig()
	config.InitialCapital = decimal.NewFromFloat(10000)

	// Generate sample data
	dataLoader := NewDataLoader()
	data := dataLoader.GenerateSampleData("BTC-USD", time.Now().Add(-24*time.Hour), 100, 50000)
	testutils.AssertNotNil(t, data, "Data should not be nil")
	testutils.AssertTrue(t, len(data.Candles) > 0, "Should have candles")

	// Create engine
	engine := NewEngine(config, data)

	// Create strategy config
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = "BTC-USD"

	// Run backtest
	metrics, err := engine.Run(strategyConfig)
	testutils.AssertNoError(t, err, "Backtest should run without error")
	testutils.AssertNotNil(t, metrics, "Metrics should be generated")

	// Verify basic metrics
	testutils.AssertTrue(t, metrics.TotalTrades >= 0, "Should have valid trade count")
	testutils.AssertTrue(t, len(metrics.EquityCurve) > 0, "Should have equity curve")
	testutils.AssertTrue(t, len(metrics.Trades) == metrics.TotalTrades, "Trades count should match")

	// Verify equity curve starts with initial capital
	if len(metrics.EquityCurve) > 0 {
		testutils.AssertTrue(t, metrics.EquityCurve[0].Equity.Equal(config.InitialCapital),
			"Equity curve should start with initial capital")
	}

	// Verify final equity is reasonable (not negative, not excessively high)
	testutils.AssertTrue(t, metrics.EquityCurve[len(metrics.EquityCurve)-1].Equity.GreaterThanOrEqual(decimal.Zero),
		"Final equity should not be negative")
	testutils.AssertTrue(t, metrics.EquityCurve[len(metrics.EquityCurve)-1].Equity.LessThan(decimal.NewFromFloat(1000000)),
		"Final equity should be reasonable")
}
