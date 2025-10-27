# Symbol Selector Tool

A command-line tool for intelligent dYdX market symbol selection and quality evaluation.

## Description

The Symbol Selector analyzes all available dYdX perpetual trading pairs and identifies the best opportunities based on:

- **Volume (35% weight)**: 24-hour trading volume
- **Liquidity (35% weight)**: Order book depth and bid-ask spreads
- **Volatility (30% weight)**: Price stability analysis

Each market receives a composite quality score from 0-1, with higher scores indicating better trading opportunities.

## Installation

```bash
cd constantine
go build -o bin/symbol-selector ./cmd/symbol-selector/main.go
```

## Usage

### Basic Command

```bash
./bin/symbol-selector
```

This outputs the top 10 markets (default) with quality >= 30% (default).

### With Options

```bash
# Select top 20 high-quality markets (70% minimum quality)
./bin/symbol-selector -max=20 -min-quality=0.7

# Verbose output with detailed analysis
./bin/symbol-selector -max=10 -min-quality=0.5 -verbose

# Select only top 5 excellent markets
./bin/symbol-selector -max=5 -min-quality=0.85
```

## Options

- `-max int`: Maximum number of symbols to select (default: 10)
- `-min-quality float`: Minimum quality score [0-1] (default: 0.3)
- `-verbose`: Show detailed analysis for each symbol

## Output

### Standard Output

```
═══════════════════════════════════════════════════════════
          dYdX Market Symbol Selector v1.0
═══════════════════════════════════════════════════════════

Configuration:
  Max Symbols: 10
  Min Quality: 30.0%
  Verbose: false

Step 1: Retrieving all available markets...
✓ Retrieved 150 total markets in 0.08s

Step 2: Filtering markets by quality criteria...
✓ Filtered to 45 quality markets in 12.34s

Step 3: Selecting top markets by quality score...
✓ Selected 10 best markets in 0.01s

═══════════════════════════════════════════════════════════
                    RESULTS
═══════════════════════════════════════════════════════════

RANK      SYMBOL     QUALITY    VOLUME(USD)    LIQUIDITY    VOLATILITY
─────────────────────────────────────────────────────────────
1         BTC-USD    95.2%      $92.7M         98.5%        2.3%
2         ETH-USD    93.8%      $45.2M         96.2%        3.1%
...
```

### Verbose Output

Includes detailed metrics for each symbol:

```bash
./bin/symbol-selector -max=3 -min-quality=0.7 -verbose
```

Output:
```
Top 10 Recommended Symbols:

1. BTC-USD
   Quality Score: 95.18%
   Volume (24h): $92.74M
   Liquidity: 98.54%
   Volatility: 2.30%
   Status: ✅ EXCELLENT - Very high quality trading pair

2. ETH-USD
   Quality Score: 93.82%
   Volume (24h): $45.23M
   Liquidity: 96.21%
   Volatility: 3.12%
   Status: ✅ EXCELLENT - Very high quality trading pair
```

## Quality Score Interpretation

| Score | Classification | Recommendation |
|-------|-----------------|-----------------|
| 0.85+ | Excellent | ✅ Ideal for scalping and active trading |
| 0.70-0.85 | Good | 🟢 Great for swing trading |
| 0.50-0.70 | Fair | 🟡 Suitable for conservative strategies |
| 0.30-0.50 | Poor | 🔴 Research only |
| < 0.30 | Unsuitable | ❌ Not recommended |

## Example Workflows

### 1. Scalping Strategy

Find top 5 ultra-high-quality symbols:

```bash
./bin/symbol-selector -max=5 -min-quality=0.85
```

Use these 5 symbols for high-frequency scalping.

### 2. Swing Trading

Find top 15 high-quality symbols:

```bash
./bin/symbol-selector -max=15 -min-quality=0.70
```

Use these for multi-day swing trades.

### 3. Diversified Portfolio

Find top 30 good-quality symbols:

```bash
./bin/symbol-selector -max=30 -min-quality=0.50
```

Spread capital across 20-30 positions.

### 4. Daily Analysis

Run with verbose mode to get detailed metrics:

```bash
./bin/symbol-selector -max=10 -min-quality=0.65 -verbose
```

Analyze top opportunities and adjust strategy.

## Integration with Constantine Bot

### Automatic Symbol Selection

Update your bot configuration to use recommended symbols:

```bash
# Get recommendations
./bin/symbol-selector -max=10 -min-quality=0.70 > symbol-analysis.txt

# Extract symbols (manually or via script)
# BTC-USD, ETH-USD, SOL-USD, ...

# Update bot configuration
export TRADING_SYMBOLS="BTC-USD,ETH-USD,SOL-USD,AVAX-USD"
./bin/bot
```

### Programmatic Usage

```go
import "github.com/guyghost/constantine/internal/exchanges/dydx"

client, _ := dydx.NewClient("", "")
markets, _ := client.SelectBestMarkets(ctx, 10, 0.70)

for _, m := range markets {
    fmt.Printf("%s (Quality: %.1f%%)\n", m.Symbol, m.QualityScore*100)
}
```

## Performance

- **First run**: ~15 seconds (evaluates ~150 markets)
- **Cached runs**: <1 second (uses 5-minute cache)
- **Network**: Requires API access to https://indexer.dydx.trade

## Requirements

- Go 1.18+
- Network access to dYdX Indexer API
- ~50MB free memory

## Troubleshooting

### No markets found

Problem: Results say "No markets met the selection criteria"

Solution: Lower the quality threshold

```bash
./bin/symbol-selector -min-quality=0.3
```

### Slow execution

Problem: Takes >20 seconds

Solution: This is normal on first run. Subsequent runs use cache.

```bash
# First run (slow)
./bin/symbol-selector -max=10 -min-quality=0.70

# Second run within 5 minutes (fast)
./bin/symbol-selector -max=10 -min-quality=0.70
```

### API errors

Problem: "Failed to get markets" error

Solution: Check internet connection and API availability

```bash
# Test API connectivity
curl https://indexer.dydx.trade/v4/perpetualMarkets
```

## See Also

- [Symbol Selection Guide](../../docs/SYMBOL_SELECTION.md)
- [Integration Guide](../../docs/SYMBOL_SELECTION_INTEGRATION.md)
- [dYdX Integration](../../docs/DYDX_INTEGRATION.md)
- [Constantine Documentation](../../docs/)

## License

Part of Constantine Trading Bot
