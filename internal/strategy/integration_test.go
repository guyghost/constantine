package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/config"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/testutils"
	"github.com/shopspring/decimal"
)

// TestIntegratedStrategyEngineInitialization tests engine initialization
func TestIntegratedStrategyEngineInitialization(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD", "ETH-USD"}

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)

	if engine == nil {
		t.Errorf("Expected non-nil engine")
	}

	if engine.signalGenerator == nil {
		t.Errorf("Expected signal generator")
	}

	if engine.symbolSelector == nil {
		t.Errorf("Expected symbol selector")
	}

	if engine.weightCalculator == nil {
		t.Errorf("Expected weight calculator")
	}
}

// TestIntegratedStrategyEngineStart tests engine startup
func TestIntegratedStrategyEngineStart(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD"}

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 10*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := engine.Start(ctx); err != nil {
		t.Errorf("Failed to start engine: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	if !engine.running {
		t.Errorf("Engine should be running")
	}

	if err := engine.Stop(); err != nil {
		t.Errorf("Failed to stop engine: %v", err)
	}
}

// TestIntegratedStrategyEngineSymbolSelection tests symbol selection
func TestIntegratedStrategyEngineSymbolSelection(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD", "ETH-USD"}

	// Create candles with proper volume data for test exchange
	candles := make([]exchanges.Candle, 30)
	for i := 0; i < 30; i++ {
		candles[i] = exchanges.Candle{
			Symbol:    "BTC-USD",
			Timestamp: time.Now().Add(-time.Duration(30-i) * time.Minute),
			Open:      decimal.NewFromFloat(100.0 + float64(i)*0.5),
			High:      decimal.NewFromFloat(101.0 + float64(i)*0.5),
			Low:       decimal.NewFromFloat(99.0 + float64(i)*0.5),
			Close:     decimal.NewFromFloat(100.5 + float64(i)*0.5),
			Volume:    decimal.NewFromFloat(1000.0),
		}
	}
	exchange.CandlesValue = candles

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)

	ctx := context.Background()

	// Manually trigger symbol selection
	engine.updateSymbolSelection(ctx)

	selected := engine.GetSelectedSymbols()

	// Should have selected some symbols
	if len(selected) == 0 {
		t.Errorf("Expected some symbols selected")
	}

	// All selected symbols should be in config
	for symbol := range selected {
		found := false
		for _, cs := range tradingSymbols {
			if cs == symbol {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Selected symbol not in config: %s", symbol)
		}
	}
}

// TestIntegratedStrategyEngineWeightCalculation tests weight integration
func TestIntegratedStrategyEngineWeightCalculation(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD"}

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)

	// Create test prices
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.5)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Test weight calculation
	wc := engine.GetWeightCalculator()
	weights := wc.CalculateDynamicWeights(prices, volumes, decimal.NewFromFloat(45.0))

	if weights.EMA == 0 && weights.RSI == 0 {
		t.Errorf("Weights should be calculated")
	}

	sum := weights.EMA + weights.RSI + weights.Volume + weights.BB
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Weights should sum to 1.0, got %f", sum)
	}
}

// TestIntegratedStrategyEngineSignalGeneration tests signal generation
func TestIntegratedStrategyEngineSignalGeneration(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD"}

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)

	// Create test data
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.8)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Generate signal
	sg := engine.GetSignalGenerator()
	signal := sg.GenerateSignal("BTC-USD", prices, volumes, nil)

	if signal == nil {
		t.Errorf("Expected signal")
	}
}

// TestIntegratedStrategyEngineMultipleSymbols tests with multiple symbols
func TestIntegratedStrategyEngineMultipleSymbols(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD", "ETH-USD", "ADA-USD", "SOL-USD"}

	// Create candles with proper volume data for test exchange
	candles := make([]exchanges.Candle, 30)
	for i := 0; i < 30; i++ {
		candles[i] = exchanges.Candle{
			Symbol:    "BTC-USD",
			Timestamp: time.Now().Add(-time.Duration(30-i) * time.Minute),
			Open:      decimal.NewFromFloat(100.0 + float64(i)*0.5),
			High:      decimal.NewFromFloat(101.0 + float64(i)*0.5),
			Low:       decimal.NewFromFloat(99.0 + float64(i)*0.5),
			Close:     decimal.NewFromFloat(100.5 + float64(i)*0.5),
			Volume:    decimal.NewFromFloat(1000.0),
		}
	}
	exchange.CandlesValue = candles

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)

	ctx := context.Background()

	// Update symbol selection
	engine.updateSymbolSelection(ctx)

	selected := engine.GetSelectedSymbols()

	// Should select roughly 50% (2 out of 4)
	if len(selected) < 1 {
		t.Errorf("Expected at least 1 symbol selected")
	}

	if len(selected) > len(tradingSymbols) {
		t.Errorf("Should not select more symbols than available")
	}
}

// TestIntegratedStrategyEngineCallbacks tests callback mechanism
func TestIntegratedStrategyEngineCallbacks(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD"}

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)

	// Verify callbacks can be set on underlying strategy
	engine.SetSignalCallback(func(s *Signal) {
		// Callback set
	})

	engine.SetErrorCallback(func(e error) {
		// Callback set
	})

	// Callbacks should be set on underlying strategy
	strategy := engine.GetScalpingStrategy()
	if strategy == nil {
		t.Errorf("Expected underlying strategy")
	}

	// Verify that GetScalpingStrategy returns a valid strategy
	if strategy.config == nil {
		t.Errorf("Strategy should have config")
	}
}

// TestIntegratedStrategyEngineWeightDistribution tests weight distribution
func TestIntegratedStrategyEngineWeightDistribution(t *testing.T) {
	cfg := config.DefaultConfig()
	exchange := testutils.NewTestExchange("test")
	tradingSymbols := []string{"BTC-USD"}

	engine := NewIntegratedStrategyEngine(cfg, tradingSymbols, exchange, 1*time.Second)
	wc := engine.GetWeightCalculator()

	// Test various market conditions
	testCases := []struct {
		name    string
		vol     float64
		trend   float64
		prices  []decimal.Decimal
		volumes []decimal.Decimal
	}{
		{
			name:  "Low volatility uptrend",
			vol:   0.05,
			trend: 0.8,
			prices: func() []decimal.Decimal {
				p := make([]decimal.Decimal, 30)
				for i := 0; i < 30; i++ {
					p[i] = decimal.NewFromFloat(100.0 + float64(i)*0.5)
				}
				return p
			}(),
			volumes: func() []decimal.Decimal {
				v := make([]decimal.Decimal, 30)
				for i := 0; i < 30; i++ {
					v[i] = decimal.NewFromFloat(1000.0)
				}
				return v
			}(),
		},
	}

	for _, tc := range testCases {
		weights := wc.CalculateDynamicWeights(tc.prices, tc.volumes, decimal.NewFromFloat(45.0))

		if weights.EMA < 0 || weights.EMA > 1 {
			t.Errorf("Test %s: EMA weight out of range: %f", tc.name, weights.EMA)
		}

		if weights.RSI < 0 || weights.RSI > 1 {
			t.Errorf("Test %s: RSI weight out of range: %f", tc.name, weights.RSI)
		}

		sum := weights.EMA + weights.RSI + weights.Volume + weights.BB
		if sum < 0.99 || sum > 1.01 {
			t.Errorf("Test %s: Weights don't sum to 1.0: %f", tc.name, sum)
		}
	}
}
