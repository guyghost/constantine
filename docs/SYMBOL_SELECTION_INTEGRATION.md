# Integrating Symbol Selection into Constantine Bot

## Overview

This guide explains how to integrate the intelligent dYdX market symbol selection feature into the Constantine trading bot to automatically discover and trade the best opportunities.

## Quick Start

### 1. Run the Symbol Selector Tool

```bash
./bin/symbol-selector -max=10 -min-quality=0.7 -verbose
```

This will output:
- All markets evaluated (e.g., "Retrieved 150 total markets")
- Quality filtering results (e.g., "Filtered to 45 quality markets")
- Top 10 markets ranked by quality score
- Detailed metrics for each recommended market

### 2. Copy Recommended Symbols

From the output, note the top 3-5 symbols:
- BTC-USD
- ETH-USD
- SOL-USD
- AVAX-USD
- etc.

### 3. Update Bot Configuration

Edit your bot's configuration file to use the recommended symbols:

```yaml
# In bot config or .env file
TRADING_SYMBOLS="BTC-USD,ETH-USD,SOL-USD,AVAX-USD,LINK-USD"

# Or in code
cfg.TradingSymbols = []string{
    "BTC-USD",
    "ETH-USD",
    "SOL-USD",
    "AVAX-USD",
    "LINK-USD",
}
```

### 4. Start the Bot

```bash
./bin/bot
```

The bot will now focus on the best trading opportunities.

## Programmatic Integration

### Example: Automatic Symbol Selection in Bot

```go
package main

import (
	"context"
	"log"
	
	"github.com/guyghost/constantine/internal/exchanges/dydx"
)

func selectTradingSymbols(ctx context.Context) ([]string, error) {
	// Create dYdX client
	client, err := dydx.NewClient("", "")
	if err != nil {
		return nil, err
	}
	
	// Select top 5 high-quality markets
	markets, err := client.SelectBestMarkets(ctx, 5, 0.85)
	if err != nil {
		return nil, err
	}
	
	// Extract symbols
	symbols := make([]string, len(markets))
	for i, market := range markets {
		symbols[i] = market.Symbol
		log.Printf("Selected: %s (Quality: %.1f%%)\n", 
			market.Symbol, market.QualityScore*100)
	}
	
	return symbols, nil
}

func main() {
	ctx := context.Background()
	
	// Get dynamically selected symbols
	symbols, err := selectTradingSymbols(ctx)
	if err != nil {
		log.Fatalf("Failed to select symbols: %v", err)
	}
	
	// Use symbols in bot strategy
	// ... rest of bot logic
}
```

### Example: Adaptive Symbol Selection

Update trading symbols periodically based on current market conditions:

```go
func monitorAndUpdateSymbols(ctx context.Context, bot *TradingBot) {
	ticker := time.NewTicker(1 * time.Hour) // Update every hour
	defer ticker.Stop()
	
	for range ticker.C {
		// Get current best markets
		client, _ := dydx.NewClient("", "")
		markets, err := client.SelectBestMarkets(ctx, 10, 0.70)
		if err != nil {
			log.Printf("Failed to update symbols: %v", err)
			continue
		}
		
		// Extract new symbols
		newSymbols := make([]string, len(markets))
		for i, m := range markets {
			newSymbols[i] = m.Symbol
		}
		
		// Update bot configuration
		bot.UpdateTradingSymbols(newSymbols)
		log.Printf("Updated trading symbols: %v", newSymbols)
	}
}
```

## Quality Thresholds by Strategy

### Scalping Strategy
- **Threshold**: 0.85+ (excellent quality)
- **Count**: Top 3-5 symbols
- **Rationale**: Need very tight spreads and high liquidity

```go
markets, _ := client.SelectBestMarkets(ctx, 5, 0.85)
```

### Swing Trading
- **Threshold**: 0.70+ (good quality)
- **Count**: Top 10-15 symbols
- **Rationale**: Need good trending characteristics

```go
markets, _ := client.SelectBestMarkets(ctx, 15, 0.70)
```

### Mean Reversion
- **Threshold**: 0.60+ (fair quality)
- **Count**: Top 15-20 symbols
- **Rationale**: Need volatile symbols with reversion patterns

```go
markets, _ := client.SelectBestMarkets(ctx, 20, 0.60)
```

### Diversified Portfolio
- **Threshold**: 0.50+ (moderate quality)
- **Count**: Top 20-30 symbols
- **Rationale**: Focus on diversification over individual quality

```go
markets, _ := client.SelectBestMarkets(ctx, 30, 0.50)
```

## Monitoring & Analytics

### Track Quality Changes

Log how market quality evolves:

```go
func trackMarketQuality(ctx context.Context, symbols []string) {
	client, _ := dydx.NewClient("", "")
	
	for _, symbol := range symbols {
		quality, err := client.EvaluateMarketQuality(ctx, symbol)
		if err != nil {
			continue
		}
		
		log.Printf("Market: %s\n", symbol)
		log.Printf("  Quality: %.1f%%\n", quality.QualityScore*100)
		log.Printf("  Volume: $%.2fM\n", toMillions(quality.Volume24h))
		log.Printf("  Liquidity: %.1f%%\n", quality.Liquidity*100)
		log.Printf("  Volatility: %.1f%%\n", quality.Volatility*100)
	}
}
```

### Export Quality Metrics

For analysis and backtesting:

```go
func exportMarketMetrics(ctx context.Context, filename string) error {
	client, _ := dydx.NewClient("", "")
	
	filtered, _ := client.FilterMarketsByQuality(ctx, 0.0)
	
	// Export to CSV/JSON
	for symbol, quality := range filtered {
		// Write to file
		// symbol, quality.QualityScore, quality.Volume24h, etc.
	}
	
	return nil
}
```

## Performance Considerations

### API Call Optimization

- **First run**: ~15 seconds (evaluates all 150+ markets)
- **Cached runs**: <1 second (uses 5-minute cache)
- **Selective evaluation**: <10 seconds (if evaluating subset)

### Recommendations

1. **Cache best practice**: Call once per hour
2. **In production**: Use cached results, refresh periodically
3. **Backtesting**: Pre-evaluate symbols once, reuse cache
4. **Paper trading**: Update symbols every 4-6 hours

## Troubleshooting

### Issue: "No markets meet selection criteria"

**Cause**: Quality threshold too high

**Solution**: Lower the minimum quality threshold

```go
// Try 0.50 instead of 0.85
markets, _ := client.SelectBestMarkets(ctx, 10, 0.50)
```

### Issue: Evaluation takes too long

**Cause**: First run or cache expired

**Solution**: Use results from cache or wait

```go
// Check cache first
if cached := client.GetMarketCache(); cached != nil {
	// Use cached results
} else {
	// Evaluation will take time
	markets, _ := client.SelectBestMarkets(ctx, 10, 0.70)
}
```

### Issue: Missing required data for a market

**Cause**: Market too new or has insufficient history

**Solution**: Market will be skipped automatically, only include in results if sufficient data

## Best Practices

### 1. Regular Updates
```go
// Update symbols every 4 hours
ticker := time.NewTicker(4 * time.Hour)
for range ticker.C {
	updateSymbols()
}
```

### 2. Quality Monitoring
```go
// Log quality metrics daily
dailyQualityReport()
```

### 3. Gradual Transitions
```go
// Don't switch all symbols at once
// Gradually transition to new symbols
for _, newSymbol := range newSymbols {
	bot.AddSymbol(newSymbol)
	time.Sleep(1 * time.Minute)
}
```

### 4. Fallback Symbols
```go
// Always keep some stable symbols
stableSymbols := []string{"BTC-USD", "ETH-USD"}

// Add selected symbols
selectedSymbols, _ := client.SelectBestMarkets(ctx, 8, 0.70)

// Combine
allSymbols := append(stableSymbols, extractSymbols(selectedSymbols)...)
```

## Next Steps

1. **Run the symbol selector** to see current recommendations
2. **Test with top 3-5 symbols** to validate approach
3. **Monitor performance** over 1-2 weeks
4. **Adjust quality threshold** based on results
5. **Automate updates** when satisfied with approach

## Related Documentation

- [Symbol Selection Guide](./SYMBOL_SELECTION.md)
- [dYdX Integration](./DYDX_INTEGRATION.md)
- [Strategy Configuration](./DYNAMIC_WEIGHTS.md)
- [Bot Configuration](./QUICKSTART.md)

## API Reference

### Client Methods

```go
// Get all markets
markets, err := client.GetAllMarkets(ctx)

// Evaluate specific market
quality, err := client.EvaluateMarketQuality(ctx, "BTC-USD")

// Filter by quality threshold
filtered, err := client.FilterMarketsByQuality(ctx, 0.70)

// Select top N
best, err := client.SelectBestMarkets(ctx, 10, 0.70)

// Cache management
client.GetMarketCache()
client.ClearMarketCache()
```

### Data Structure

```go
type MarketQuality struct {
	Symbol       string          // "BTC-USD", "ETH-USD", etc.
	Volume24h    decimal.Decimal // 24-hour volume in USD
	Volatility   float64         // [0, 1] - price volatility
	Liquidity    float64         // [0, 1] - liquidity score
	QualityScore float64         // [0, 1] - composite score
}
```
