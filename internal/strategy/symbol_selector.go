package strategy

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/guyghost/constantine/internal/config"
)

type SymbolData struct {
	Prices  []decimal.Decimal
	Volumes []decimal.Decimal
}

type RankedSymbol struct {
	Symbol      string
	Score       float64 // Composite opportunity score [0, 1]
	Potential   decimal.Decimal // Gain potential
	Risk        decimal.Decimal // Risk assessment
	SharpeRatio decimal.Decimal // Risk-adjusted return
}

type SelectionEvent struct {
	Timestamp int64
	Symbol    string
	Score     float64
	Reason    string
}

type SymbolSelector struct {
	config         *config.Config
	history        []SelectionEvent
	maxHistorySize int
	mu             sync.RWMutex
}

func NewSymbolSelector(cfg *config.Config) *SymbolSelector {
	return &SymbolSelector{
		config:         cfg,
		history:        make([]SelectionEvent, 0),
		maxHistorySize: 100,
	}
}

func (ss *SymbolSelector) CalculateGainPotential(symbol string, prices []decimal.Decimal, volumes []decimal.Decimal) decimal.Decimal {
	if len(prices) < 21 {
		return decimal.Zero
	}

	// Calculate EMA 9 and EMA 21
	ema9 := ss.calculateEMA(prices, 9)
	ema21 := ss.calculateEMA(prices, 21)

	if ema9 == nil || ema21 == nil || len(ema9) == 0 || len(ema21) == 0 {
		return decimal.Zero
	}

	lastEMA9 := ema9[len(ema9)-1]
	lastEMA21 := ema21[len(ema21)-1]

	var score decimal.Decimal
	if lastEMA9.GreaterThan(lastEMA21) {
		// Uptrend
		// Find recent ATH
		ath := decimal.Zero
		for _, p := range prices {
			if p.GreaterThan(ath) {
				ath = p
			}
		}
		current := prices[len(prices)-1]
		if ath.IsZero() {
			score = decimal.NewFromFloat(0.5)
		} else {
			distance := ath.Sub(current).Div(ath)
			score = decimal.NewFromFloat(0.5).Add(decimal.NewFromFloat(0.5).Mul(decimal.NewFromFloat(1).Sub(distance)))
		}
	} else {
		// Downtrend
		score = decimal.NewFromFloat(0.25)
	}

	// Ensure score is between 0 and 1
	if score.LessThan(decimal.Zero) {
		score = decimal.Zero
	}
	if score.GreaterThan(decimal.NewFromFloat(1)) {
		score = decimal.NewFromFloat(1)
	}

	return score
}

func (ss *SymbolSelector) CalculateRiskAssessment(symbol string, prices []decimal.Decimal, volumes []decimal.Decimal) decimal.Decimal {
	if len(prices) < 20 {
		return decimal.NewFromFloat(1.0) // High risk
	}

	// Calculate volatility (stddev of log returns)
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		ratio, _ := prices[i].Div(prices[i-1]).Float64()
		if ratio > 0 {
			returns[i-1] = math.Log(ratio)
		}
	}
	vol := ss.calculateStdDevFloat(returns)

	// Calculate max drawdown
	dd := ss.calculateMaxDrawdown(prices)

	// Volume score
	volScore := decimal.Zero
	if len(volumes) > 0 {
		sum := decimal.Zero
		for _, v := range volumes {
			sum = sum.Add(v)
		}
		avgVol := sum.Div(decimal.NewFromInt(int64(len(volumes))))
		avgVolFloat, _ := avgVol.Float64()
		if avgVolFloat > 1000 {
			volScore = decimal.NewFromFloat(1.0)
		} else if avgVolFloat < 100 {
			volScore = decimal.NewFromFloat(0.0)
		} else {
			volScore = decimal.NewFromFloat((avgVolFloat - 100) / 900)
		}
	}

	volRiskFloat, _ := volScore.Float64()
	riskVal := vol*0.4 + dd*0.4 + (1-volRiskFloat)*0.2

	// Ensure risk is between 0 and 1
	if riskVal < 0 {
		riskVal = 0
	}
	if riskVal > 1 {
		riskVal = 1
	}

	return decimal.NewFromFloat(riskVal)
}

func (ss *SymbolSelector) CalculateSharpeRatio(symbol string, prices []decimal.Decimal, volumes []decimal.Decimal) decimal.Decimal {
	if len(prices) < 2 {
		return decimal.Zero
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		ratio, _ := prices[i].Div(prices[i-1]).Float64()
		if ratio > 0 {
			returns[i-1] = math.Log(ratio)
		}
	}

	meanReturn := 0.0
	for _, r := range returns {
		meanReturn += r
	}
	meanReturn /= float64(len(returns))

	stdDev := ss.calculateStdDevFloat(returns)

	riskFreeRate := 0.01 // 1% annual

	if stdDev == 0 {
		return decimal.Zero
	}

	sharpe := (meanReturn - riskFreeRate) / stdDev
	return decimal.NewFromFloat(sharpe)
}

func (ss *SymbolSelector) CalculateOpportunityScore(symbol string, prices []decimal.Decimal, volumes []decimal.Decimal) float64 {
	gp := ss.CalculateGainPotential(symbol, prices, volumes)
	risk := ss.CalculateRiskAssessment(symbol, prices, volumes)
	sharpe := ss.CalculateSharpeRatio(symbol, prices, volumes)

	// Volume confirmation
	volScore := 0.0
	if len(volumes) > 0 {
		sum := decimal.Zero
		for _, v := range volumes {
			sum = sum.Add(v)
		}
		avgVol := sum.Div(decimal.NewFromInt(int64(len(volumes))))
		avgVolFloat, _ := avgVol.Float64()
		volScore = avgVolFloat / 1000.0
		if volScore > 1 {
			volScore = 1
		}
	}

	gpFloat, _ := gp.Float64()
	sharpeFloat, _ := sharpe.Float64()
	riskFloat, _ := risk.Float64()

	score := gpFloat*0.4 + sharpeFloat*0.35 + volScore*0.15 - riskFloat*0.10

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

func (ss *SymbolSelector) RankSymbols(symbols []string, symbolData map[string]SymbolData) []RankedSymbol {
	ranked := make([]RankedSymbol, 0, len(symbols))
	for _, sym := range symbols {
		data, ok := symbolData[sym]
		if !ok {
			continue
		}
		score := ss.CalculateOpportunityScore(sym, data.Prices, data.Volumes)
		gp := ss.CalculateGainPotential(sym, data.Prices, data.Volumes)
		risk := ss.CalculateRiskAssessment(sym, data.Prices, data.Volumes)
		sharpe := ss.CalculateSharpeRatio(sym, data.Prices, data.Volumes)
		ranked = append(ranked, RankedSymbol{
			Symbol:      sym,
			Score:       score,
			Potential:   gp,
			Risk:        risk,
			SharpeRatio: sharpe,
		})
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})
	return ranked
}

func (ss *SymbolSelector) SelectBestSymbols(symbols []string, symbolData map[string]SymbolData, maxCount int) []RankedSymbol {
	ranked := ss.RankSymbols(symbols, symbolData)
	threshold := ss.CalculateDynamicThreshold(symbols, symbolData, maxCount)
	selected := make([]RankedSymbol, 0)
	for _, r := range ranked {
		if r.Score >= threshold && len(selected) < maxCount {
			selected = append(selected, r)
			ss.addToHistory(SelectionEvent{
				Timestamp: time.Now().Unix(),
				Symbol:    r.Symbol,
				Score:     r.Score,
				Reason:    "Selected based on opportunity score",
			})
		}
	}
	return selected
}

func (ss *SymbolSelector) CalculateDynamicThreshold(symbols []string, symbolData map[string]SymbolData, minSymbols int) float64 {
	ranked := ss.RankSymbols(symbols, symbolData)
	if len(ranked) == 0 {
		return 0.0
	}
	if len(ranked) <= minSymbols {
		return 0.0
	}

	// If all > 0.6, threshold 0.6
	minScore := 1.0
	for _, r := range ranked {
		if r.Score < minScore {
			minScore = r.Score
		}
	}
	if minScore > 0.6 {
		return 0.6
	}

	// Else, find threshold to get at least minSymbols
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})
	if len(ranked) >= minSymbols {
		return ranked[minSymbols-1].Score
	}
	return 0.0
}

func (ss *SymbolSelector) FilterByPriceRange(symbols []string, symbolData map[string]SymbolData, minPrice, maxPrice decimal.Decimal) []string {
	filtered := make([]string, 0)
	for _, sym := range symbols {
		data, ok := symbolData[sym]
		if !ok || len(data.Prices) == 0 {
			continue
		}
		currentPrice := data.Prices[len(data.Prices)-1]
		if currentPrice.GreaterThanOrEqual(minPrice) && currentPrice.LessThanOrEqual(maxPrice) {
			filtered = append(filtered, sym)
		}
	}
	return filtered
}

func (ss *SymbolSelector) GetSelectionHistory() []SelectionEvent {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	history := make([]SelectionEvent, len(ss.history))
	copy(history, ss.history)
	return history
}

func (ss *SymbolSelector) calculateEMA(prices []decimal.Decimal, period int) []decimal.Decimal {
	if len(prices) < period {
		return nil
	}
	multiplier := decimal.NewFromFloat(2).Div(decimal.NewFromInt(int64(period)).Add(decimal.NewFromInt(1)))
	ema := make([]decimal.Decimal, len(prices))
	for i := period - 1; i < len(prices); i++ {
		if i == period-1 {
			sum := decimal.Zero
			for j := 0; j < period; j++ {
				sum = sum.Add(prices[j])
			}
			ema[i] = sum.Div(decimal.NewFromInt(int64(period)))
		} else {
			ema[i] = prices[i].Mul(multiplier).Add(ema[i-1].Mul(decimal.NewFromInt(1).Sub(multiplier)))
		}
	}
	return ema
}

func (ss *SymbolSelector) calculateStdDevFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))
	
	sumSq := 0.0
	for _, v := range values {
		diff := v - mean
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(values))
	return math.Sqrt(variance)
}

func (ss *SymbolSelector) calculateMaxDrawdown(prices []decimal.Decimal) float64 {
	if len(prices) == 0 {
		return 0
	}
	maxPrice := prices[0]
	maxDD := 0.0
	for _, p := range prices {
		if p.GreaterThan(maxPrice) {
			maxPrice = p
		}
		dd := maxPrice.Sub(p).Div(maxPrice)
		ddFloat, _ := dd.Float64()
		if ddFloat > maxDD {
			maxDD = ddFloat
		}
	}
	return maxDD
}

func (ss *SymbolSelector) addToHistory(event SelectionEvent) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.history = append(ss.history, event)
	if len(ss.history) > ss.maxHistorySize {
		ss.history = ss.history[1:]
	}
}
