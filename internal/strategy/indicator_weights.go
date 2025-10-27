package strategy

import (
	"math"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/config"
	"github.com/guyghost/constantine/internal/logger"
	"github.com/shopspring/decimal"
)

// IndicatorWeights represents the dynamic weights for technical indicators
type IndicatorWeights struct {
	EMA    float64 // Weight for EMA (crossover trend)
	RSI    float64 // Weight for RSI (momentum)
	Volume float64 // Weight for volume
	BB     float64 // Weight for Bollinger Bands
}

// MarketCondition represents market conditions at a specific timestamp
type MarketCondition struct {
	Timestamp     int64
	Volatility    decimal.Decimal
	TrendStrength decimal.Decimal
	RSIMomentum   decimal.Decimal
	VolumeRatio   decimal.Decimal
}

// WeightCalculator manages dynamic weight calculations based on market conditions
type WeightCalculator struct {
	config         *config.Config
	history        []MarketCondition
	maxHistorySize int
	mu             sync.RWMutex
}

// NewWeightCalculator creates a new WeightCalculator instance
func NewWeightCalculator(cfg *config.Config) *WeightCalculator {
	return &WeightCalculator{
		config:         cfg,
		history:        make([]MarketCondition, 0),
		maxHistorySize: 50,
	}
}

// CalculateDynamicWeights calculates adaptive weights based on current market conditions
func (wc *WeightCalculator) CalculateDynamicWeights(
	prices []decimal.Decimal,
	volumes []decimal.Decimal,
	rsi decimal.Decimal,
) IndicatorWeights {
	if len(prices) < 20 || len(volumes) < 20 {
		logger.Component("strategy").Debug("Insufficient data for weight calculation, using defaults")
		return IndicatorWeights{EMA: 0.35, RSI: 0.35, Volume: 0.15, BB: 0.15}
	}

	volatility := wc.CalculateVolatility(prices)
	trendStrength := wc.CalculateTrendStrength(prices)
	rsiMomentum := wc.CalculateRSIMomentum(rsi)
	volumeRatio := wc.CalculateVolumeRatio(volumes)

	// Record market condition
	condition := MarketCondition{
		Timestamp:     time.Now().Unix(),
		Volatility:    volatility,
		TrendStrength: trendStrength,
		RSIMomentum:   rsiMomentum,
		VolumeRatio:   volumeRatio,
	}
	wc.addToHistory(condition)

	// Start with default weights
	weights := map[string]float64{
		"EMA":    0.35,
		"RSI":    0.35,
		"Volume": 0.15,
		"BB":     0.15,
	}

	// Apply volatility adjustments
	volFloat, _ := volatility.Float64()
	if volFloat < 0.2 {
		weights["EMA"] += 0.2
		weights["RSI"] -= 0.1
	} else if volFloat > 0.8 {
		weights["EMA"] -= 0.2
		weights["RSI"] += 0.15
		weights["BB"] += 0.1
	}

	// Apply trend strength adjustments
	trendFloat, _ := trendStrength.Float64()
	if trendFloat > 0.7 {
		weights["EMA"] += 0.15
	} else if trendFloat < 0.3 {
		weights["EMA"] -= 0.2
	}

	// Apply RSI momentum adjustments
	rsiFloat, _ := rsiMomentum.Float64()
	if rsiFloat > 0.8 {
		weights["RSI"] += 0.2
	} else if rsiFloat < 0.3 {
		weights["RSI"] -= 0.15
	}

	// Apply volume ratio adjustments
	volRatioFloat, _ := volumeRatio.Float64()
	if volRatioFloat > 0.7 {
		weights["Volume"] += 0.3
	} else if volRatioFloat < 0.3 {
		weights["Volume"] -= 0.2
	}

	// Clamp weights to [0, 1]
	for k, v := range weights {
		if v < 0 {
			weights[k] = 0
		} else if v > 1 {
			weights[k] = 1
		}
	}

	normalized := wc.NormalizeWeights(weights)

	// Log significant changes
	if math.Abs(normalized.EMA-0.35) > 0.1 || math.Abs(normalized.RSI-0.35) > 0.1 {
		logger.Component("strategy").Debug("Significant weight adjustment",
			"EMA", normalized.EMA,
			"RSI", normalized.RSI,
			"Volume", normalized.Volume,
			"BB", normalized.BB)
	}

	return normalized
}

// CalculateVolatility calculates normalized volatility [0, 1] using Bollinger Bands width
func (wc *WeightCalculator) CalculateVolatility(prices []decimal.Decimal) decimal.Decimal {
	if len(prices) < 20 {
		return decimal.Zero
	}

	// Calculate simple moving average and standard deviation
	sum := decimal.Zero
	for _, p := range prices {
		sum = sum.Add(p)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(prices))))

	sumSq := decimal.Zero
	for _, p := range prices {
		diff := p.Sub(mean)
		sumSq = sumSq.Add(diff.Mul(diff))
	}
	variance := sumSq.Div(decimal.NewFromInt(int64(len(prices))))
	stdDev := decimal.NewFromFloat(math.Sqrt(variance.InexactFloat64()))

	// Normalized volatility: (stdDev / mean) with sigmoid-like scaling
	if mean.IsZero() {
		return decimal.Zero
	}

	ratio := stdDev.Div(mean)
	ratioFloat, _ := ratio.Float64()

	// Scale using exponential decay to normalize to [0, 1]
	// Lower volatility coefficient gives more granular distinction
	volatilityFloat := 1.0 - math.Exp(-ratioFloat*2.0) // More sensitive to volatility changes

	// Clamp to [0, 1]
	if volatilityFloat > 1.0 {
		volatilityFloat = 1.0
	} else if volatilityFloat < 0.0 {
		volatilityFloat = 0.0
	}

	return decimal.NewFromFloat(volatilityFloat)
}

// CalculateTrendStrength calculates trend strength [0, 1] using EMA divergence
func (wc *WeightCalculator) CalculateTrendStrength(prices []decimal.Decimal) decimal.Decimal {
	if len(prices) < 21 {
		return decimal.Zero
	}

	// Calculate short EMA (9) and long EMA (21)
	shortEMA := wc.calculateEMA(prices, 9)
	longEMA := wc.calculateEMA(prices, 21)

	// Trend strength: |shortEMA - longEMA| / longEMA
	if longEMA.IsZero() {
		return decimal.Zero
	}

	divergence := shortEMA.Sub(longEMA).Abs()
	strength := divergence.Div(longEMA)

	// Clamp to [0, 1]
	if strength.GreaterThan(decimal.NewFromInt(1)) {
		strength = decimal.NewFromInt(1)
	}

	return strength
}

// CalculateRSIMomentum calculates RSI momentum [0, 1]
func (wc *WeightCalculator) CalculateRSIMomentum(rsi decimal.Decimal) decimal.Decimal {
	rsiFloat, _ := rsi.Float64()

	var momentum float64
	if rsiFloat < 35 {
		momentum = 1 - (35-rsiFloat)/35
	} else if rsiFloat <= 70 {
		momentum = 0.2
	} else {
		momentum = 1 - (rsiFloat-70)/30
	}

	// Clamp to [0, 1]
	if momentum < 0 {
		momentum = 0
	} else if momentum > 1 {
		momentum = 1
	}

	return decimal.NewFromFloat(momentum)
}

// CalculateVolumeRatio calculates volume ratio [0, 1] vs 20-period average
func (wc *WeightCalculator) CalculateVolumeRatio(volumes []decimal.Decimal) decimal.Decimal {
	if len(volumes) < 21 {
		return decimal.Zero
	}

	// Calculate average of last 20 volumes
	sum := decimal.Zero
	for i := len(volumes) - 21; i < len(volumes)-1; i++ {
		sum = sum.Add(volumes[i])
	}
	avg := sum.Div(decimal.NewFromInt(20))

	if avg.IsZero() {
		return decimal.Zero
	}

	// Current volume ratio
	current := volumes[len(volumes)-1]
	ratio := current.Div(avg)

	// Clamp to [0, 1]
	if ratio.GreaterThan(decimal.NewFromInt(1)) {
		ratio = decimal.NewFromInt(1)
	} else if ratio.LessThan(decimal.Zero) {
		ratio = decimal.Zero
	}

	return ratio
}

// NormalizeWeights normalizes weights so they sum to 1.0
func (wc *WeightCalculator) NormalizeWeights(weights map[string]float64) IndicatorWeights {
	total := 0.0
	for _, v := range weights {
		total += v
	}

	if total == 0 {
		return IndicatorWeights{EMA: 0.25, RSI: 0.25, Volume: 0.25, BB: 0.25}
	}

	return IndicatorWeights{
		EMA:    weights["EMA"] / total,
		RSI:    weights["RSI"] / total,
		Volume: weights["Volume"] / total,
		BB:     weights["BB"] / total,
	}
}

// calculateEMA calculates exponential moving average
func (wc *WeightCalculator) calculateEMA(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period {
		return decimal.Zero
	}

	multiplier := decimal.NewFromFloat(2.0 / (float64(period) + 1.0))
	ema := prices[0]

	for i := 1; i < len(prices); i++ {
		ema = prices[i].Mul(multiplier).Add(ema.Mul(decimal.NewFromInt(1).Sub(multiplier)))
	}

	return ema
}

// addToHistory adds a market condition to history (thread-safe)
func (wc *WeightCalculator) addToHistory(condition MarketCondition) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.history = append(wc.history, condition)
	if len(wc.history) > wc.maxHistorySize {
		wc.history = wc.history[1:]
	}
}

// GetHistory returns the market condition history (thread-safe)
func (wc *WeightCalculator) GetHistory() []MarketCondition {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	history := make([]MarketCondition, len(wc.history))
	copy(history, wc.history)
	return history
}
