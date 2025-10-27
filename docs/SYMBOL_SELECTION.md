# dYdX Market Symbol Selection Guide

## Overview

Constantine now includes intelligent market symbol selection for dYdX perpetual futures. This guide explains how to leverage this feature to automatically discover and rank the best trading opportunities.

## Features

### 1. **Market Discovery**
Retrieve all available perpetual trading pairs from dYdX with comprehensive market data:
- Market status and trading conditions
- 24-hour volume
- Open interest
- Funding rates
- Price volatility metrics

### 2. **Quality Evaluation**
Evaluate each market based on three key metrics:
- **Volume Score (35% weight)**: Higher trading volume indicates better liquidity
- **Liquidity Score (35% weight)**: Measure of order book depth and spread
- **Volatility Score (30% weight)**: Price stability assessment (lower volatility = higher quality score)

### 3. **Intelligent Selection**
Select the top N markets based on composite quality score:
- Filter by minimum quality threshold
- Rank by trading potential
- Cache results to minimize API calls
- Parallel evaluation for performance

## API Usage

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	
	"github.com/guyghost/constantine/internal/exchanges/dydx"
)

func main() {
	// Create dYdX client
	client, err := dydx.NewClient("", "")
	if err != nil {
		panic(err)
	}
	
	ctx := context.Background()
	
	// Get all available markets
	markets, err := client.GetAllMarkets(ctx)
	fmt.Printf("Total markets: %d\n", len(markets))
	
	// Evaluate quality of a specific market
	quality, err := client.EvaluateMarketQuality(ctx, "BTC-USD")
	fmt.Printf("BTC-USD Quality: %.1f%%\n", quality.QualityScore*100)
	
	// Filter by minimum quality
	filtered, err := client.FilterMarketsByQuality(ctx, 0.3) // 30% minimum quality
	fmt.Printf("Markets above 30% quality: %d\n", len(filtered))
	
	// Select top 10 markets
	best, err := client.SelectBestMarkets(ctx, 10, 0.3)
	for i, market := range best {
		fmt.Printf("%d. %s - Quality: %.1f%%\n", 
			i+1, market.Symbol, market.QualityScore*100)
	}
}
```

## CLI Tool

### Installation

```bash
cd constantine
go build -o bin/symbol-selector ./cmd/symbol-selector/main.go
```

### Usage

```bash
# Select top 10 markets with 30% minimum quality
./bin/symbol-selector -max=10 -min-quality=0.3

# Verbose output with detailed analysis
./bin/symbol-selector -max=10 -min-quality=0.5 -verbose

# Select top 20 high-quality markets
./bin/symbol-selector -max=20 -min-quality=0.6
```

## Quality Metrics Explained

### Volume Score
- Raw volume is normalized using a sigmoid function
- Below $1M: Score reduced to 0.2
- $1M-$100M range: Linear scaling
- $100M+: Asymptotic approach to 1.0

### Liquidity Score
Calculated from order book analysis:
- Order book depth (sum of bid+ask volumes at top 20 levels)
- Bid-ask spread (narrower spread = better liquidity)
- Score normalized to [0, 1] range

### Volatility Score
Calculated from historical price movements:
- Log returns from recent candles
- Standard deviation of returns
- Converted to penalty: lower volatility = higher quality score

### Composite Quality Score

```
Quality = (Volume × 0.35) + (Liquidity × 0.35) + (Volatility_Score × 0.30)
```

## Quality Thresholds

| Quality Score | Classification | Recommended Use |
|---|---|---|
| 0.85+ | Excellent | Primary trading symbols, scalping |
| 0.70-0.85 | Good | Active trading, swing trading |
| 0.50-0.70 | Fair | Conservative strategies, hedging |
| 0.30-0.50 | Poor | Research only, avoid in live trading |
| < 0.30 | Unsuitable | Not recommended for trading |

## Suggested Quality Thresholds

**Scalping Strategy**: Use top 5 markets with 0.85+ quality
```go
markets, _ := client.SelectBestMarkets(ctx, 5, 0.85)
```

**Swing Trading**: Use top 15 markets with 0.70+ quality
```go
markets, _ := client.SelectBestMarkets(ctx, 15, 0.70)
```

**Diversified Portfolio**: Use top 20 markets with 0.50+ quality
```go
markets, _ := client.SelectBestMarkets(ctx, 20, 0.50)
```

## Implementation Details

### Caching Strategy

Market data is cached with a 5-minute TTL to:
- Minimize API calls
- Improve performance
- Reduce rate limit pressure

Cache can be manually cleared:
```go
client.ClearMarketCache()
```

### Parallel Processing

Quality evaluation uses parallel processing with concurrency control (max 5 concurrent evaluations).

## See Also

- [dYdX Integration Guide](./DYDX_INTEGRATION.md)
- [Strategy Configuration](./DYNAMIC_WEIGHTS.md)
- [API Reference](./INTEGRATION.md)
