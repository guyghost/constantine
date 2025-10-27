package strategy

import (
	"math"
	"testing"

	"github.com/guyghost/constantine/internal/config"
	"github.com/shopspring/decimal"
)

// TestCalculateVolatility tests volatility calculation using Bollinger Bands
func TestCalculateVolatility(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Create test prices with known volatility
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(101.0),
		decimal.NewFromFloat(102.0),
		decimal.NewFromFloat(101.5),
		decimal.NewFromFloat(103.0),
		decimal.NewFromFloat(102.5),
		decimal.NewFromFloat(104.0),
		decimal.NewFromFloat(103.5),
		decimal.NewFromFloat(105.0),
		decimal.NewFromFloat(104.5),
		decimal.NewFromFloat(106.0),
		decimal.NewFromFloat(105.5),
		decimal.NewFromFloat(107.0),
		decimal.NewFromFloat(106.5),
		decimal.NewFromFloat(108.0),
		decimal.NewFromFloat(107.5),
		decimal.NewFromFloat(109.0),
		decimal.NewFromFloat(108.5),
		decimal.NewFromFloat(110.0),
		decimal.NewFromFloat(109.5),
	}

	volatility := wc.CalculateVolatility(prices)

	// Volatility should be between 0 and 1
	if volatility.LessThan(decimal.Zero) || volatility.GreaterThan(decimal.NewFromInt(1)) {
		t.Errorf("Volatility out of range: %s", volatility.String())
	}

	// Volatility should be positive for this data
	if volatility.LessThanOrEqual(decimal.Zero) {
		t.Errorf("Expected positive volatility, got %s", volatility.String())
	}
}

// TestVolatilityEdgeCases tests edge cases for volatility calculation
func TestVolatilityEdgeCases(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Test with insufficient data
	smallPrices := []decimal.Decimal{
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(101.0),
	}
	volatility := wc.CalculateVolatility(smallPrices)
	if !volatility.Equal(decimal.Zero) {
		t.Errorf("Expected zero volatility for insufficient data, got %s", volatility.String())
	}

	// Test with identical prices (zero volatility)
	samePrices := make([]decimal.Decimal, 20)
	for i := 0; i < 20; i++ {
		samePrices[i] = decimal.NewFromFloat(100.0)
	}
	volatility = wc.CalculateVolatility(samePrices)
	if !volatility.Equal(decimal.Zero) {
		t.Errorf("Expected zero volatility for identical prices, got %s", volatility.String())
	}
}

// TestCalculateTrendStrength tests trend strength calculation
func TestCalculateTrendStrength(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Create uptrending prices
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.5)
	}

	strength := wc.CalculateTrendStrength(prices)

	// Strength should be between 0 and 1
	if strength.LessThan(decimal.Zero) || strength.GreaterThan(decimal.NewFromInt(1)) {
		t.Errorf("Trend strength out of range: %s", strength.String())
	}

	// Strong uptrend should have positive strength
	if strength.LessThanOrEqual(decimal.Zero) {
		t.Errorf("Expected positive strength for uptrend, got %s", strength.String())
	}
}

// TestTrendStrengthNeutral tests neutral trend
func TestTrendStrengthNeutral(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Create sideways prices (no trend)
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
		decimal.NewFromFloat(100.0),
		decimal.NewFromFloat(100.5),
	}

	strength := wc.CalculateTrendStrength(prices)

	// Neutral trend should have low strength
	if strength.GreaterThan(decimal.NewFromFloat(0.3)) {
		t.Errorf("Expected low strength for neutral trend, got %s", strength.String())
	}
}

// TestCalculateRSIMomentum tests RSI momentum calculation
func TestCalculateRSIMomentum(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Test oversold RSI
	rsiOversold := decimal.NewFromFloat(30.0)
	momentum := wc.CalculateRSIMomentum(rsiOversold)
	if momentum.LessThan(decimal.Zero) || momentum.GreaterThan(decimal.NewFromInt(1)) {
		t.Errorf("Momentum out of range: %s", momentum.String())
	}
	if momentum.LessThan(decimal.NewFromFloat(0.5)) {
		t.Errorf("Expected high momentum for oversold, got %s", momentum.String())
	}

	// Test neutral RSI
	rsiNeutral := decimal.NewFromFloat(50.0)
	momentum = wc.CalculateRSIMomentum(rsiNeutral)
	if momentum.GreaterThan(decimal.NewFromFloat(0.3)) {
		t.Errorf("Expected low momentum for neutral RSI, got %s", momentum.String())
	}

	// Test overbought RSI
	rsiOverbought := decimal.NewFromFloat(75.0)
	momentum = wc.CalculateRSIMomentum(rsiOverbought)
	if momentum.LessThan(decimal.NewFromFloat(0.5)) {
		t.Errorf("Expected high momentum for overbought, got %s", momentum.String())
	}
}

// TestRSIMomentumExtremes tests extreme RSI values
func TestRSIMomentumExtremes(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Test RSI = 0
	rsi0 := decimal.NewFromFloat(0.0)
	momentum := wc.CalculateRSIMomentum(rsi0)
	if momentum.LessThan(decimal.Zero) || momentum.GreaterThan(decimal.NewFromInt(1)) {
		t.Errorf("Extreme RSI 0 momentum out of range: %s", momentum.String())
	}

	// Test RSI = 100
	rsi100 := decimal.NewFromFloat(100.0)
	momentum = wc.CalculateRSIMomentum(rsi100)
	if momentum.LessThan(decimal.Zero) || momentum.GreaterThan(decimal.NewFromInt(1)) {
		t.Errorf("Extreme RSI 100 momentum out of range: %s", momentum.String())
	}
}

// TestCalculateDynamicWeights tests dynamic weight calculation
func TestCalculateDynamicWeights(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Create test data
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.5)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	rsi := decimal.NewFromFloat(45.0)

	weights := wc.CalculateDynamicWeights(prices, volumes, rsi)

	// Check that weights sum to approximately 1.0
	sum := weights.EMA + weights.RSI + weights.Volume + weights.BB
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Weights sum to %f, expected ~1.0", sum)
	}
}

// TestDynamicWeightsConsistency tests weight consistency
func TestDynamicWeightsConsistency(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.5)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	rsi := decimal.NewFromFloat(45.0)

	weights := wc.CalculateDynamicWeights(prices, volumes, rsi)

	// All weights should be between 0 and 1
	if weights.EMA < 0 || weights.EMA > 1 {
		t.Errorf("EMA weight out of range: %f", weights.EMA)
	}
	if weights.RSI < 0 || weights.RSI > 1 {
		t.Errorf("RSI weight out of range: %f", weights.RSI)
	}
	if weights.Volume < 0 || weights.Volume > 1 {
		t.Errorf("Volume weight out of range: %f", weights.Volume)
	}
	if weights.BB < 0 || weights.BB > 1 {
		t.Errorf("BB weight out of range: %f", weights.BB)
	}
}

// TestWeightsAdaptToMarketChanges tests that weights adapt to market evolution
func TestWeightsAdaptToMarketChanges(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Low volatility scenario
	lowVolPrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		lowVolPrices[i] = decimal.NewFromFloat(100.0 + float64(i%2)*0.1)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	rsi := decimal.NewFromFloat(45.0)

	lowVolWeights := wc.CalculateDynamicWeights(lowVolPrices, volumes, rsi)

	// High volatility scenario
	highVolPrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		highVolPrices[i] = decimal.NewFromFloat(100.0 + float64(i)*2.0)
	}

	highVolWeights := wc.CalculateDynamicWeights(highVolPrices, volumes, rsi)

	// Weights should be different between scenarios
	if lowVolWeights.EMA == highVolWeights.EMA && lowVolWeights.RSI == highVolWeights.RSI {
		t.Errorf("Weights didn't adapt to market change")
	}
}

// TestVolatilityImpactOnWeights tests that high volatility impacts weights correctly
func TestVolatilityImpactOnWeights(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Low volatility
	lowVolPrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		lowVolPrices[i] = decimal.NewFromFloat(100.0 + float64(i%2)*0.05)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	rsi := decimal.NewFromFloat(45.0)

	lowVolWeights := wc.CalculateDynamicWeights(lowVolPrices, volumes, rsi)

	// High volatility
	highVolPrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		highVolPrices[i] = decimal.NewFromFloat(100.0 + float64(i)*3.0)
	}

	highVolWeights := wc.CalculateDynamicWeights(highVolPrices, volumes, rsi)

	// High volatility should reduce EMA weight and increase RSI weight
	if highVolWeights.EMA >= lowVolWeights.EMA {
		t.Errorf("High volatility should reduce EMA weight")
	}
	if highVolWeights.RSI <= lowVolWeights.RSI {
		t.Errorf("High volatility should increase RSI weight")
	}
}

// TestLowVolatilityImpact tests that weight calculator handles different volatility levels
func TestLowVolatilityImpact(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Create a base trending scenario
	basePrices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		basePrices[i] = decimal.NewFromFloat(100.0 + float64(i)*0.3)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	rsi := decimal.NewFromFloat(45.0)

	// Get baseline weights
	baseWeights := wc.CalculateDynamicWeights(basePrices, volumes, rsi)

	// Verify baseline weights are normalized
	sum := baseWeights.EMA + baseWeights.RSI + baseWeights.Volume + baseWeights.BB
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Weights should sum to 1.0, got %f", sum)
	}

	// Verify all individual weights are between 0 and 1
	if baseWeights.EMA < 0 || baseWeights.EMA > 1 {
		t.Errorf("EMA weight out of range: %f", baseWeights.EMA)
	}
	if baseWeights.RSI < 0 || baseWeights.RSI > 1 {
		t.Errorf("RSI weight out of range: %f", baseWeights.RSI)
	}
}

// TestOversoldMomentumBoosting tests RSI oversold momentum boosting
func TestOversoldMomentumBoosting(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Neutral RSI
	neutralRSI := decimal.NewFromFloat(50.0)
	neutralWeights := wc.CalculateDynamicWeights(prices, volumes, neutralRSI)

	// Oversold RSI
	oversoldRSI := decimal.NewFromFloat(25.0)
	oversoldWeights := wc.CalculateDynamicWeights(prices, volumes, oversoldRSI)

	// Oversold should boost RSI weight
	if oversoldWeights.RSI <= neutralWeights.RSI {
		t.Errorf("Oversold RSI should boost RSI weight")
	}
}

// TestOverboughtMomentumBoosting tests RSI overbought momentum boosting
func TestOverboughtMomentumBoosting(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Neutral RSI
	neutralRSI := decimal.NewFromFloat(50.0)
	neutralWeights := wc.CalculateDynamicWeights(prices, volumes, neutralRSI)

	// Overbought RSI
	overboughtRSI := decimal.NewFromFloat(80.0)
	overboughtWeights := wc.CalculateDynamicWeights(prices, volumes, overboughtRSI)

	// Overbought should boost RSI weight
	if overboughtWeights.RSI <= neutralWeights.RSI {
		t.Errorf("Overbought RSI should boost RSI weight")
	}
}

// TestNeutralMomentumAdjustment tests neutral RSI momentum adjustment
func TestNeutralMomentumAdjustment(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	// Neutral RSI
	neutralRSI := decimal.NewFromFloat(50.0)
	neutralWeights := wc.CalculateDynamicWeights(prices, volumes, neutralRSI)

	// Extreme RSI (should have higher RSI weight)
	extremeRSI := decimal.NewFromFloat(20.0)
	extremeWeights := wc.CalculateDynamicWeights(prices, volumes, extremeRSI)

	// Neutral RSI should have lower RSI weight than extreme
	if neutralWeights.RSI >= extremeWeights.RSI {
		t.Errorf("Neutral RSI should have lower weight than extreme RSI")
	}
}

// TestWeightsIntegration tests integration of all influences
func TestWeightsIntegration(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Complex scenario: high volatility + strong uptrend + overbought
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volatilityComponent := float64(i%3) * 0.8 // Creates spikes
		trendComponent := float64(i) * 1.2        // Strong uptrend
		prices[i] = decimal.NewFromFloat(100.0 + trendComponent + volatilityComponent)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0 + float64(i%5)*500.0)
	}

	overboughtRSI := decimal.NewFromFloat(80.0)

	weights := wc.CalculateDynamicWeights(prices, volumes, overboughtRSI)

	// Weights should sum to 1.0
	sum := weights.EMA + weights.RSI + weights.Volume + weights.BB
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Integration: weights sum to %f, expected ~1.0", sum)
	}

	// All individual weights should be valid
	if weights.EMA < 0 || weights.EMA > 1 {
		t.Errorf("Integration: EMA weight out of range: %f", weights.EMA)
	}
	if weights.RSI < 0 || weights.RSI > 1 {
		t.Errorf("Integration: RSI weight out of range: %f", weights.RSI)
	}
}

// TestWeightsStability tests weights stability under stable conditions
func TestWeightsStability(t *testing.T) {
	cfg := config.DefaultConfig()
	wc := NewWeightCalculator(cfg)

	// Stable market conditions
	prices := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		prices[i] = decimal.NewFromFloat(100.0)
	}

	volumes := make([]decimal.Decimal, 30)
	for i := 0; i < 30; i++ {
		volumes[i] = decimal.NewFromFloat(1000.0)
	}

	neutralRSI := decimal.NewFromFloat(50.0)

	// Get weights multiple times
	weights1 := wc.CalculateDynamicWeights(prices, volumes, neutralRSI)
	weights2 := wc.CalculateDynamicWeights(prices, volumes, neutralRSI)
	weights3 := wc.CalculateDynamicWeights(prices, volumes, neutralRSI)

	// Weights should be stable (within tolerance)
	const tolerance = 0.001
	if math.Abs(weights1.EMA-weights2.EMA) > tolerance || math.Abs(weights2.EMA-weights3.EMA) > tolerance {
		t.Errorf("Weights not stable under stable conditions: %f vs %f vs %f (tolerance: %f)",
			weights1.EMA, weights2.EMA, weights3.EMA, tolerance)
	}
}
