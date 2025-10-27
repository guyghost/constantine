# Automatic Symbol Selection at Startup

## Overview

Constantine now automatically selects the best trading symbols from dYdX at startup. No need to run the `symbol-selector` CLI tool separately - the bot intelligently discovers optimal trading pairs and configures itself.

## Quick Start

### Default Behavior (Auto-Selection Enabled)

```bash
./bin/bot
```

The bot will:
1. Auto-select the top 5 markets with 70%+ quality
2. Connect to those symbols
3. Start trading

No configuration needed!

## Configuration

### Environment Variables

Control auto-selection behavior with environment variables:

```bash
# Enable/disable auto-selection
export AUTO_SELECT_SYMBOLS=true          # Default: true

# Maximum number of symbols to select
export AUTO_SELECT_MAX_SYMBOLS=5         # Default: 10

# Minimum quality threshold (0.0 to 1.0)
export AUTO_SELECT_MIN_QUALITY=0.70      # Default: 0.70

# Then run the bot
./bin/bot
```

### Option 1: Full Auto-Selection (Recommended)

```bash
# Uses all defaults
./bin/bot

# Bot will auto-select top 10 markets with 70%+ quality
```

### Option 2: Customize Auto-Selection

```bash
# Select top 5 markets with 80%+ quality (scalping strategy)
export AUTO_SELECT_MAX_SYMBOLS=5
export AUTO_SELECT_MIN_QUALITY=0.80
./bin/bot
```

### Option 3: Disable Auto-Selection

```bash
# Use pre-configured symbols from .env or defaults
export AUTO_SELECT_SYMBOLS=false
./bin/bot

# Falls back to configured TRADING_SYMBOLS or defaults (BTC-USD, ETH-USD)
```

### Option 4: Manual Symbol Configuration

```bash
# Pre-configure symbols and disable auto-selection
export TRADING_SYMBOLS="BTC-USD,ETH-USD,SOL-USD,AVAX-USD"
export AUTO_SELECT_SYMBOLS=false
./bin/bot

# Bot will use these symbols instead of auto-selecting
```

## How It Works

### Startup Sequence

```
1. Bot starts
   ↓
2. Load configuration from .env / environment
   ↓
3. Check if trading symbols are configured
   ├─ If YES → Use configured symbols
   └─ If NO → Proceed to auto-selection
   ↓
4. Check AUTO_SELECT_SYMBOLS flag
   ├─ If false → Use defaults (BTC-USD, ETH-USD)
   └─ If true → Proceed to auto-selection
   ↓
5. Connect to dYdX
   ↓
6. Run symbol selector algorithm
   ├─ Input: MAX_SYMBOLS (5-10), MIN_QUALITY (0.5-0.8)
   └─ Output: Ranked list of best markets
   ↓
7. Log selected symbols with metrics
   ↓
8. Initialize trading for selected symbols
```

### Quality Scoring

The auto-selection uses the same quality scoring algorithm as the CLI tool:

```
Quality = (Volume × 35%) + (Liquidity × 35%) + ((1 - Volatility) × 30%)
```

**Metrics:**
- **Volume**: 24-hour trading volume (normalized)
- **Liquidity**: Order book depth + spread analysis
- **Volatility**: Price stability (lower = better)

### Performance

- **Selection Time**: ~0.2-1 second
- **API Calls**: 1-2 (cached)
- **Failure Fallback**: Automatic (uses defaults)
- **Timeout Protection**: 30-second maximum

## Use Cases

### Scalping Strategy

Select tight-spread, highly liquid pairs:

```bash
export AUTO_SELECT_MAX_SYMBOLS=5
export AUTO_SELECT_MIN_QUALITY=0.80
./bin/bot
```

**Recommended symbols**: ETH-USD, BTC-USD, SOL-USD

### Swing Trading

Balance quality with diversification:

```bash
export AUTO_SELECT_MAX_SYMBOLS=15
export AUTO_SELECT_MIN_QUALITY=0.60
./bin/bot
```

**Recommended symbols**: Top 15 with 60%+ quality

### Conservative Trading

Highest quality symbols only:

```bash
export AUTO_SELECT_MAX_SYMBOLS=5
export AUTO_SELECT_MIN_QUALITY=0.80
./bin/bot
```

### Research/Backtesting

See all available symbols:

```bash
export AUTO_SELECT_MAX_SYMBOLS=284
export AUTO_SELECT_MIN_QUALITY=0.0
./bin/bot
```

## Monitoring Startup

Watch the bot select symbols at startup:

```bash
export LOG_LEVEL=info
./bin/bot
```

Output will show:

```
[INFO] auto-selecting best trading symbols...
[INFO] selected symbol symbol=ETH-USD quality=82.5% volume=$84.7M
[INFO] selected symbol symbol=SOL-USD quality=79.5% volume=$31.4M
[INFO] selected symbol symbol=BTC-USD quality=79.3% volume=$92.4M
[INFO] auto-selection complete symbols=[ETH-USD,SOL-USD,BTC-USD] count=3
```

## Example .env File

```bash
# Auto-Selection Configuration
AUTO_SELECT_SYMBOLS=true
AUTO_SELECT_MAX_SYMBOLS=5
AUTO_SELECT_MIN_QUALITY=0.70

# dYdX Configuration (required for auto-selection)
DYDX_ENABLED=true
DYDX_API_KEY=your_api_key
DYDX_API_SECRET=your_api_secret

# Optional: Override with specific symbols
# TRADING_SYMBOLS=BTC-USD,ETH-USD,SOL-USD
# AUTO_SELECT_SYMBOLS=false

# Logging
LOG_LEVEL=info
LOG_FORMAT=text
```

## Fallback Behavior

If auto-selection fails, the bot automatically falls back:

1. **If dYdX unavailable**: Uses default symbols (BTC-USD, ETH-USD)
2. **If no markets found**: Uses default symbols
3. **If timeout**: Uses default symbols
4. **If error**: Logs error and uses default symbols

This ensures the bot always starts trading with some symbols.

## Customization

### Adjust Quality Weights

To modify scoring weights, edit `internal/exchanges/dydx/market_selector.go`:

```go
const (
    volumeWeight     = 0.35    // Increase for volume-focused
    liquidityWeight  = 0.35    // Increase for liquidity-focused
    volatilityWeight = 0.30    // Increase to penalize volatility more
)
```

### Change Scoring Algorithm

Modify the scoring functions:
- `estimateVolatilityFromSpread()`
- `estimateLiquidityFromTicker()`
- `normalizeVolume()`

## Troubleshooting

### "Auto-selection disabled"

Check environment:
```bash
echo $AUTO_SELECT_SYMBOLS
# Should be: true
```

### "dYdX not enabled"

Enable dYdX:
```bash
export DYDX_ENABLED=true
```

### "No markets found"

Lower the quality threshold:
```bash
export AUTO_SELECT_MIN_QUALITY=0.50
```

### "Failed to create dYdX client"

Verify dYdX configuration:
```bash
export DYDX_API_KEY=your_key
export DYDX_API_SECRET=your_secret
```

## Comparison: CLI vs Auto-Selection

| Feature | CLI Tool | Auto-Selection |
|---------|----------|---|
| Command | `./bin/symbol-selector` | `./bin/bot` |
| Frequency | Manual (every 4 hours) | Automatic (at startup) |
| Configuration | Command-line flags | Environment variables |
| Result Type | Display output | Automatic trading |
| Time to Trading | 2-3 min | 30 seconds |

## Best Practices

1. **Monitor first run**: Watch logs to see selected symbols
2. **Adjust quality threshold**: Lower for more symbols, higher for fewer
3. **Test with defaults first**: Start with AUTO_SELECT_SYMBOLS=true
4. **Override when needed**: Use TRADING_SYMBOLS for manual control
5. **Log selection**: Keep LOG_LEVEL=info for visibility
6. **Update periodically**: Rerun bot daily to refresh selections

## Related Documents

- [Symbol Selection Guide](./SYMBOL_SELECTION.md)
- [Bot Configuration](./QUICKSTART.md)
- [dYdX Integration](./DYDX_INTEGRATION.md)

## Questions?

Check the logs:
```bash
export LOG_LEVEL=debug
./bin/bot
```

Detailed debug output will show:
- Market evaluation details
- Quality calculations
- Selection reasoning
- Configuration applied
