package dydx

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
)

const (
	// Cache duration for market data
	marketCacheDuration = 5 * time.Minute

	// Quality score weights
	volumeWeight     = 0.35 // 35% volume
	liquidityWeight  = 0.35 // 35% liquidity
	volatilityWeight = 0.30 // 30% volatility (penalizing extreme values)

	// Minimum volume threshold (in USD)
	minVolumeUSD = 1000000.0 // $1M minimum

	// Maximum volatility threshold
	maxVolatility = 0.8 // 80% maximum volatility
)

// marketCache stores cached market data
type marketCache struct {
	markets   map[string]MarketData
	ticker    map[string]MarketTicker
	timestamp time.Time
	mu        sync.RWMutex
}

// GetAllMarkets retrieves all available markets from dYdX
func (c *Client) GetAllMarkets(ctx context.Context) (map[string]MarketData, error) {
	if c.marketCache != nil {
		c.marketCache.mu.RLock()
		if time.Since(c.marketCache.timestamp) < marketCacheDuration && len(c.marketCache.markets) > 0 {
			defer c.marketCache.mu.RUnlock()
			return c.marketCache.markets, nil
		}
		c.marketCache.mu.RUnlock()
	}

	// Get markets data
	var marketsResp MarketsResponse
	if err := c.httpClient.get(ctx, "/v4/perpetualMarkets", &marketsResp); err != nil {
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	// Cache the result
	if c.marketCache == nil {
		c.marketCache = &marketCache{}
	}
	c.marketCache.mu.Lock()
	c.marketCache.markets = marketsResp.Markets
	c.marketCache.timestamp = time.Now()
	c.marketCache.mu.Unlock()

	return marketsResp.Markets, nil
}

// EvaluateMarketQuality evaluates the quality metrics of a market
func (c *Client) EvaluateMarketQuality(ctx context.Context, symbol string) (*MarketQuality, error) {
	// Get market data
	markets, err := c.GetAllMarkets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	_, exists := markets[symbol]
	if !exists {
		return nil, fmt.Errorf("market %s not found", symbol)
	}

	// Get ticker data for volume and price info
	var tickerResp TickerResponse
	if err := c.httpClient.get(ctx, "/v4/perpetualMarkets", &tickerResp); err != nil {
		return nil, fmt.Errorf("failed to get ticker data: %w", err)
	}

	tickerData, ok := tickerResp.Markets[symbol]
	if !ok {
		return nil, fmt.Errorf("ticker data for %s not found", symbol)
	}

	// Calculate volatility from ticker data (fast)
	volatility := c.estimateVolatilityFromSpread(tickerData)

	// Calculate liquidity from ticker data (fast)
	liquidity := c.estimateLiquidityFromTicker(tickerData)

	// Calculate volume score (normalized)
	volumeUSD, _ := tickerData.Volume24h.Float64()
	volumeScore := c.normalizeVolume(volumeUSD)

	// Calculate composite quality score
	// Volatility is a penalty (higher volatility = lower score)
	volatilityScore := 1.0 - math.Min(volatility, 1.0)

	qualityScore := (volumeScore * volumeWeight) +
		(liquidity * liquidityWeight) +
		(volatilityScore * volatilityWeight)

	return &MarketQuality{
		Symbol:       symbol,
		Volume24h:    tickerData.Volume24h,
		Volatility:   volatility,
		Liquidity:    liquidity,
		QualityScore: qualityScore,
	}, nil
}

// FilterMarketsByQuality filters markets based on minimum quality threshold
func (c *Client) FilterMarketsByQuality(ctx context.Context, minQuality float64) (map[string]MarketQuality, error) {
	markets, err := c.GetAllMarkets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	// Get ticker data once
	var tickerResp TickerResponse
	if err := c.httpClient.get(ctx, "/v4/perpetualMarkets", &tickerResp); err != nil {
		return nil, fmt.Errorf("failed to get ticker data: %w", err)
	}

	filtered := make(map[string]MarketQuality)
	filteredMu := sync.Mutex{}

	// Process markets in parallel with limited concurrency
	semaphore := make(chan struct{}, 5) // Max 5 concurrent evaluations
	var wg sync.WaitGroup

	for symbol := range markets {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			tickerData, ok := tickerResp.Markets[sym]
			if !ok {
				return
			}

			// Calculate volatility from ticker data (fast)
			volatility := c.estimateVolatilityFromSpread(tickerData)

			// Calculate liquidity from ticker data (fast)
			liquidity := c.estimateLiquidityFromTicker(tickerData)

			// Get volume score
			volumeUSD, _ := tickerData.Volume24h.Float64()
			volumeScore := c.normalizeVolume(volumeUSD)

			// Calculate composite quality score
			volatilityScore := 1.0 - math.Min(volatility, 1.0)
			qualityScore := (volumeScore * volumeWeight) +
				(liquidity * liquidityWeight) +
				(volatilityScore * volatilityWeight)

			if qualityScore >= minQuality {
				filteredMu.Lock()
				filtered[sym] = MarketQuality{
					Symbol:       sym,
					Volume24h:    tickerData.Volume24h,
					Volatility:   volatility,
					Liquidity:    liquidity,
					QualityScore: qualityScore,
				}
				filteredMu.Unlock()
			}
		}(symbol)
	}

	wg.Wait()
	return filtered, nil
}

// SelectBestMarkets selects the top N markets by quality score
func (c *Client) SelectBestMarkets(ctx context.Context, maxCount int, minQuality float64) ([]MarketQuality, error) {
	filtered, err := c.FilterMarketsByQuality(ctx, minQuality)
	if err != nil {
		return nil, fmt.Errorf("failed to filter markets: %w", err)
	}

	// Convert to slice
	markets := make([]MarketQuality, 0, len(filtered))
	for _, quality := range filtered {
		markets = append(markets, quality)
	}

	// Sort by quality score (descending)
	sort.Slice(markets, func(i, j int) bool {
		return markets[i].QualityScore > markets[j].QualityScore
	})

	// Return top N
	if len(markets) > maxCount {
		markets = markets[:maxCount]
	}

	return markets, nil
}

// calculateVolatility calculates price volatility from candles
func (c *Client) calculateVolatility(candles []exchanges.Candle, ticker MarketTicker) float64 {
	if len(candles) < 2 {
		// Estimate from bid-ask spread
		if ticker.Ask.IsZero() || ticker.Bid.IsZero() {
			return 0.1 // Default low volatility
		}
		spread := ticker.Ask.Sub(ticker.Bid).Div(ticker.Ask)
		spreadFloat, _ := spread.Float64()
		// Convert spread to volatility estimate (spread * 2 as rough estimate)
		return math.Min(spreadFloat*2.0, 0.5)
	}

	// Calculate log returns
	returns := make([]float64, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		if candles[i-1].Close.IsZero() {
			continue
		}
		ratio, _ := candles[i].Close.Div(candles[i-1].Close).Float64()
		if ratio > 0 {
			returns[i-1] = math.Log(ratio)
		}
	}

	if len(returns) == 0 {
		return 0.1
	}

	// Calculate standard deviation
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		diff := r - mean
		variance += diff * diff
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance)
}

// normalizeVolume normalizes 24h volume to [0, 1] score
func (c *Client) normalizeVolume(volumeUSD float64) float64 {
	if volumeUSD < 100000 { // $100K minimum
		return 0.1
	}

	// Log-based normalization: better distribution
	// At $100K = 0.1, at $1M = 0.35, at $10M = 0.60, at $100M+ = 0.85+
	logVolume := math.Log10(volumeUSD)
	baseLog := math.Log10(100000)          // $100K baseline
	score := 0.1 + (logVolume-baseLog)*0.2 // Scale factor
	return math.Min(math.Max(score, 0.1), 1.0)
}

// estimateVolatilityFromSpread estimates volatility from bid-ask spread
func (c *Client) estimateVolatilityFromSpread(ticker MarketTicker) float64 {
	if ticker.Ask.IsZero() || ticker.Bid.IsZero() || ticker.Ask.Cmp(ticker.Bid) <= 0 {
		return 0.1 // Default low volatility if no data
	}

	// Spread as percentage
	spread := ticker.Ask.Sub(ticker.Bid).Div(ticker.Bid)
	spreadFloat, _ := spread.Float64()

	// Estimate volatility from spread
	// 0.1% spread ≈ 0.5% volatility, 0.5% spread ≈ 1.5% volatility
	volatility := spreadFloat * 2.0

	return math.Min(volatility, 1.0)
}

// estimateLiquidityFromTicker estimates liquidity from ticker data
func (c *Client) estimateLiquidityFromTicker(ticker MarketTicker) float64 {
	// Start with base liquidity
	liquidity := 0.3

	// Volume boost: higher volume = better liquidity
	volumeFloat, _ := ticker.Volume24h.Float64()
	if volumeFloat > 10_000_000 {
		liquidity += 0.4 // +40% for >$10M volume
	} else if volumeFloat > 1_000_000 {
		liquidity += 0.2 // +20% for >$1M volume
	}

	// Open interest boost: higher OI = more depth
	oiFloat, _ := ticker.OpenInterest.Float64()
	if oiFloat > 1000 {
		liquidity += 0.2 // +20% for good open interest
	} else if oiFloat > 100 {
		liquidity += 0.1 // +10% for moderate OI
	}

	// Funding rate stability: extreme rates suggest liquidity issues
	fundingRate, _ := ticker.NextFundingRate.Float64()
	if math.Abs(fundingRate) > 0.001 { // >0.1% funding
		liquidity -= 0.1 // Penalize extreme funding
	}

	return math.Min(math.Max(liquidity, 0.1), 1.0)
}

// GetMarketCache returns cached market data if available
func (c *Client) GetMarketCache() map[string]MarketData {
	if c.marketCache == nil {
		return nil
	}
	c.marketCache.mu.RLock()
	defer c.marketCache.mu.RUnlock()

	result := make(map[string]MarketData)
	for k, v := range c.marketCache.markets {
		result[k] = v
	}
	return result
}

// ClearMarketCache clears the market data cache
func (c *Client) ClearMarketCache() {
	if c.marketCache == nil {
		return
	}
	c.marketCache.mu.Lock()
	defer c.marketCache.mu.Unlock()

	c.marketCache.markets = make(map[string]MarketData)
	c.marketCache.ticker = make(map[string]MarketTicker)
	c.marketCache.timestamp = time.Time{}
}
