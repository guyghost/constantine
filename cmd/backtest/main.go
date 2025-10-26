package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/guyghost/constantine/internal/backtesting"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
)

var (
	dataFile       = flag.String("data", "", "Path to CSV file with historical data (required)")
	symbol         = flag.String("symbol", "BTC-USD", "Trading symbol")
	initialCapital = flag.Float64("capital", 10000, "Initial capital for backtesting")
	commission     = flag.Float64("commission", 0.001, "Commission rate (e.g., 0.001 for 0.1%)")
	slippage       = flag.Float64("slippage", 0.0005, "Slippage rate (e.g., 0.0005 for 0.05%)")
	riskPerTrade   = flag.Float64("risk", 0.01, "Risk per trade as fraction of capital (e.g., 0.01 for 1%)")
	maxPositions   = flag.Int("max-positions", 1, "Maximum number of concurrent positions")

	// Strategy parameters
	shortEMA      = flag.Int("short-ema", 9, "Short EMA period")
	longEMA       = flag.Int("long-ema", 21, "Long EMA period")
	rsiPeriod     = flag.Int("rsi-period", 14, "RSI period")
	rsiOversold   = flag.Float64("rsi-oversold", 30.0, "RSI oversold threshold")
	rsiOverbought = flag.Float64("rsi-overbought", 70.0, "RSI overbought threshold")
	takeProfit    = flag.Float64("take-profit", 2.0, "Take profit percentage")
	stopLoss      = flag.Float64("stop-loss", 1.0, "Stop loss percentage")

	// Output options
	verbose        = flag.Bool("verbose", false, "Show detailed trade log")
	generateSample = flag.Bool("generate-sample", false, "Generate sample data instead of loading from file")
	sampleCandles  = flag.Int("sample-candles", 1000, "Number of candles to generate for sample data")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Print banner
	printBanner()

	// Load or generate data
	var data *backtesting.HistoricalData
	var err error

	loader := backtesting.NewDataLoader()

	if *generateSample {
		log.Println("📊 Generating sample data...")
		data = loader.GenerateSampleData(*symbol, time.Now().Add(-24*time.Hour*30), *sampleCandles, 50000)
		log.Printf("✓ Generated %d candles\n", len(data.Candles))
	} else {
		if *dataFile == "" {
			return fmt.Errorf("either -data flag or -generate-sample flag is required")
		}

		log.Printf("📂 Loading data from %s...\n", *dataFile)
		data, err = loader.LoadFromCSV(*dataFile, *symbol)
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}
		log.Printf("✓ Loaded %d candles\n", len(data.Candles))
	}

	if len(data.Candles) == 0 {
		return fmt.Errorf("no data loaded")
	}

	// Print data info
	startTime := data.Candles[0].Timestamp
	endTime := data.Candles[len(data.Candles)-1].Timestamp
	log.Printf("📅 Period: %s to %s (%s)\n",
		startTime.Format("2006-01-02"),
		endTime.Format("2006-01-02"),
		endTime.Sub(startTime).Round(time.Hour))

	// Create backtest config
	btConfig := &backtesting.BacktestConfig{
		InitialCapital: decimal.NewFromFloat(*initialCapital),
		CommissionRate: decimal.NewFromFloat(*commission),
		Slippage:       decimal.NewFromFloat(*slippage),
		RiskPerTrade:   decimal.NewFromFloat(*riskPerTrade),
		MaxPositions:   *maxPositions,
		AllowShort:     true,                       // Enable short selling for testing
		UseFixedAmount: true,                       // Use fixed amount instead of risk-based
		FixedAmount:    decimal.NewFromFloat(0.01), // Small fixed amount
		StartTime:      startTime,
		EndTime:        endTime,
	}

	// Create strategy config
	stratConfig := strategy.DefaultConfig()
	stratConfig.Symbol = *symbol
	stratConfig.ShortEMAPeriod = *shortEMA
	stratConfig.LongEMAPeriod = *longEMA
	stratConfig.RSIPeriod = *rsiPeriod
	stratConfig.RSIOversold = *rsiOversold
	stratConfig.RSIOverbought = *rsiOverbought
	stratConfig.TakeProfitPercent = *takeProfit
	stratConfig.StopLossPercent = *stopLoss

	log.Println("\n⚙️  Backtest Configuration:")
	log.Printf("   Initial Capital:  $%.2f\n", *initialCapital)
	log.Printf("   Commission:       %.2f%%\n", *commission*100)
	log.Printf("   Slippage:         %.2f%%\n", *slippage*100)
	log.Printf("   Risk per Trade:   %.2f%%\n", *riskPerTrade*100)
	log.Printf("   Max Positions:    %d\n", *maxPositions)

	log.Println("\n📊 Strategy Parameters:")
	log.Printf("   Short EMA:        %d\n", *shortEMA)
	log.Printf("   Long EMA:         %d\n", *longEMA)
	log.Printf("   RSI Period:       %d\n", *rsiPeriod)
	log.Printf("   RSI Oversold:     %.0f\n", *rsiOversold)
	log.Printf("   RSI Overbought:   %.0f\n", *rsiOverbought)
	log.Printf("   Take Profit:      %.2f%%\n", *takeProfit)
	log.Printf("   Stop Loss:        %.2f%%\n", *stopLoss)

	// Create engine
	engine := backtesting.NewEngine(btConfig, data)

	// Set callbacks for progress
	tradeCount := 0
	engine.SetOnTrade(func(trade *backtesting.Trade) {
		tradeCount++
		if *verbose {
			symbol := "✓"
			if trade.PnL.LessThan(decimal.Zero) {
				symbol = "✗"
			}
			log.Printf("[Trade #%d] %s %s: $%s → $%s = $%s (%.2f%%) [%s]\n",
				tradeCount,
				symbol,
				trade.Side,
				trade.EntryPrice.StringFixed(2),
				trade.ExitPrice.StringFixed(2),
				trade.PnL.StringFixed(2),
				trade.PnLPercent.InexactFloat64(),
				trade.ExitReason,
			)
		}
	})

	// Run backtest
	log.Println("🚀 Running backtest...")
	startRun := time.Now()

	metrics, err := engine.Run(stratConfig)
	if err != nil {
		return fmt.Errorf("backtest failed: %w", err)
	}

	duration := time.Since(startRun)
	log.Printf("✓ Backtest completed in %s\n\n", duration.Round(time.Millisecond))

	// Generate report
	reporter := backtesting.NewReporter()
	report := reporter.GenerateReport(metrics)
	fmt.Println(report)

	// Generate detailed trade log if verbose
	if *verbose && len(metrics.Trades) > 0 {
		tradeLog := reporter.GenerateTradeLog(metrics)
		fmt.Println(tradeLog)
	}

	return nil
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║        CONSTANTINE BACKTESTING FRAMEWORK              ║
║        Multi-Agent Trading System                     ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
