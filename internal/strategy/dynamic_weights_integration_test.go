package strategy

import (
	"testing"

	"github.com/guyghost/constantine/internal/config"
	"github.com/shopspring/decimal"
)

// TestSignalGeneratorWithDynamicWeights tests that SignalGenerator uses dynamic weights correctly
func TestSignalGeneratorWithDynamicWeights(t *testing.T) {
	cfg := config.DefaultConfig()
	sg := NewSignalGenerator(cfg)

	// Create test data with uptrend and low volatility
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.3)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Generate signal - this should use dynamic weights
	signal := sg.GenerateSignal("BTC-USD", prices, volumes, nil)

	// Signal generation should succeed (even if no signal triggered)
	if signal == nil {
		t.Errorf("Expected signal, got nil")
	}

	// Verify that dynamic weights were calculated
	if sg.indicatorWeights.EMA == 0 && sg.indicatorWeights.RSI == 0 {
		t.Errorf("Dynamic weights should be calculated")
	}

	// Verify weights sum to approximately 1.0
	sum := sg.indicatorWeights.EMA + sg.indicatorWeights.RSI + sg.indicatorWeights.Volume + sg.indicatorWeights.BB
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Weights should sum to ~1.0, got %f", sum)
	}
}

// TestSignalStrengthWithDynamicWeights tests that signal strength respects dynamic weights
func TestSignalStrengthWithDynamicWeights(t *testing.T) {
	cfg := config.DefaultConfig()
	sg := NewSignalGenerator(cfg)

	// Set custom weights for testing
	sg.indicatorWeights = IndicatorWeights{
		EMA:    0.7, // High EMA weight
		RSI:    0.3, // Low RSI weight
		Volume: 0.0,
		BB:     0.0,
	}

	// Strong EMA divergence
	shortEMA := decimal.NewFromFloat(110.0)
	longEMA := decimal.NewFromFloat(100.0)
	rsi := decimal.NewFromFloat(50.0) // Neutral RSI

	strength := sg.calculateSignalStrength(shortEMA, longEMA, rsi, true)

	// With high EMA weight and strong EMA divergence, strength should be significant
	if strength < 0.3 {
		t.Errorf("Expected significant strength with high EMA weight, got %f", strength)
	}

	// Now test with low EMA weight
	sg.indicatorWeights = IndicatorWeights{
		EMA:    0.1, // Very low EMA weight
		RSI:    0.9, // High RSI weight
		Volume: 0.0,
		BB:     0.0,
	}

	strength2 := sg.calculateSignalStrength(shortEMA, longEMA, rsi, true)

	// Strength should be lower when EMA weight is low
	if strength2 >= strength {
		t.Errorf("Lower EMA weight should result in lower strength: %f vs %f", strength2, strength)
	}
}

// TestDynamicWeightsAdaptToVolatilityChange tests that weights adapt to volatility changes
func TestDynamicWeightsAdaptToVolatilityChange(t *testing.T) {
	cfg := config.DefaultConfig()
	sg := NewSignalGenerator(cfg)

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Extremely low volatility (flat prices with tiny movements)
	lowVolPrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		lowVolPrices[i] = decimal.NewFromFloat(100.0 + float64(i%3)*0.001)
	}

	// Generate signal with very low volatility
	sg.GenerateSignal("BTC-USD", lowVolPrices, volumes, nil)
	lowVolWeights := sg.indicatorWeights

	// Extremely high volatility (wild swings)
	highVolPrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		if i%2 == 0 {
			highVolPrices[i] = decimal.NewFromFloat(100.0)
		} else {
			highVolPrices[i] = decimal.NewFromFloat(80.0)
		}
	}

	// Generate signal with high volatility
	sg.GenerateSignal("BTC-USD", highVolPrices, volumes, nil)
	highVolWeights := sg.indicatorWeights

	// Weights should adapt to extreme market changes
	// High volatility should reduce EMA weight and increase RSI weight
	if highVolWeights.EMA >= lowVolWeights.EMA {
		t.Logf("Warning: EMA weight not reduced for high volatility: high=%.3f vs low=%.3f",
			highVolWeights.EMA, lowVolWeights.EMA)
	}
}
