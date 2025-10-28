# Integration: Dynamic Weights & Symbol Selection

## Overview

Constantine now features an integrated strategy engine that combines:
1. **Dynamic Indicator Weights** - Adapts technical indicator emphasis based on market conditions
2. **Intelligent Symbol Selection** - Ranks and selects the best trading opportunities across configured symbols
3. **Real-time Signal Generation** - Generates trading signals with dynamic risk assessment

## Architecture

### Components

#### 1. IntegratedStrategyEngine
Located in `internal/strategy/integration.go`, this is the main orchestrator that:
- Manages symbol selection and refresh cycles
- Coordinates dynamic weight calculations
- Generates signals for selected symbols
- Handles market data fetching and caching
- Provides unified API for the bot and TUI

#### 2. WeightCalculator
Located in `internal/strategy/indicator_weights.go`:
- **Volatility Analysis**: Adjusts weights based on market volatility
  - High volatility → Boost RSI weight (mean reversion tendency)
  - Low volatility → Boost EMA weight (trend-following)
- **Trend Strength**: Weights EMA indicators higher during strong trends
- **Momentum**: Boosts RSI during oversold/overbought conditions
- **Volume**: Adjusts based on trading volume strength

**Default Weights** (before adjustment):
- EMA: 35%
- RSI: 35%
- Volume: 20%
- Bollinger Bands: 10%

#### 3. SymbolSelector
Located in `internal/strategy/symbol_selector.go`:
- **Gain Potential** (40%): Future price appreciation estimate
- **Sharpe Ratio** (35%): Risk-adjusted returns
- **Volume** (15%): Trading volume strength
- **Risk Penalty** (-10%): Volatility discount

**Selection Strategy**:
- Ranks all configured symbols by opportunity score
- Selects approximately 50% of available symbols
- Ensures at least 1 symbol minimum
- Dynamically adjusts thresholds based on available opportunities

#### 4. SignalGenerator
Located in `internal/strategy/signals.go`:
- Generates buy/sell signals using dynamic weights
- Calculates signal strength (0.0-1.0)
- Applies risk management (stop-loss, take-profit)
- Supports multiple indicator combinations

### Data Flow

```
┌──────────────────────────────────────────────────────────────────┐
│                IntegratedStrategyEngine                          │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ refreshSymbolSelection() - Every 30 seconds             │    │
│  │  1. Fetch market data for all configured symbols        │    │
│  │  2. Calculate volatility, trend, volume metrics         │    │
│  │  3. Score each symbol (gain potential, sharpe, etc)     │    │
│  │  4. Rank and select best symbols (50%)                  │    │
│  │  5. Update selected symbols map                         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ updateSymbolSelection() - Called periodically            │    │
│  │  - Fetches price and volume data for candidates          │    │
│  │  - Evaluates using SymbolSelector                        │    │
│  │  - Updates internal state with new selections            │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ For each selected symbol:                               │    │
│  │  - WeightCalculator: Calculate dynamic weights          │    │
│  │  - SignalGenerator: Generate signals with weights       │    │
│  │  - Emit signals to callbacks                            │    │
│  └─────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────┘
         ↓ Signals
┌──────────────────────────────────────────────────────────────────┐
│  ExecutionAgent (in cmd/bot/main.go)                            │
│  - Validates signals against risk parameters                     │
│  - Places orders via OrderManager                                │
│  - Manages position lifecycle                                    │
└──────────────────────────────────────────────────────────────────┘
```

## Usage in Bot

### Initialization

The bot automatically creates and manages the IntegratedStrategyEngine:

```go
// In cmd/bot/main.go initializeBot()
integratedEngine := strategy.NewIntegratedStrategyEngine(
    baseStrategyConfig,
    appConfig.TradingSymbols,  // Configured trading symbols
    primaryExchange,            // For market data fetching
    30 * time.Second,          // Refresh interval
)
```

### Configuration

#### Trading Symbols
Configure in `.env` or via environment:
```
TRADING_SYMBOLS=BTC-USD,ETH-USD,SOL-USD,ADA-USD
```

#### Indicator Settings
Modify in `internal/config/config.go` DefaultConfig():
```go
// EMA periods
ShortEMA:  9
LongEMA:   21

// RSI settings
RSIPeriod:       14
RSIOversold:     35
RSIOverbought:   70

// Bollinger Bands
BBPeriod:    20
BBStdDev:    2.0

// Risk management
TakeProfitPercent: 0.8
StopLossPercent:   0.4
```

### Runtime Operation

1. **Bot Startup**:
   - Load trading symbols from config
   - Create IntegratedStrategyEngine
   - Start symbol refresh loop (30-second interval)

2. **Symbol Refresh Cycle**:
   - Fetch market data for all configured symbols
   - Calculate scoring metrics (gain potential, sharpe ratio, volume)
   - Rank symbols by opportunity score
   - Update selected symbols for trading

3. **Signal Generation** (for selected symbols):
   - Calculate dynamic weights based on market conditions
   - Generate buy/sell signals using weighted indicators
   - Calculate signal strength (0.0-1.0)
   - Emit signals to ExecutionAgent

4. **Execution**:
   - ExecutionAgent validates signals
   - Places orders with risk management
   - Tracks positions
   - Records P&L

## Example Market Condition Adaptations

### Low Volatility Uptrend
- **Market**: Consistent price increase, low volatility
- **Adaptation**:
  - EMA weight: +20% (strong trend following)
  - RSI weight: -15% (less useful in trending markets)
  - Result: System favors trend-following signals

### High Volatility Ranging
- **Market**: Price oscillating between levels, high volatility
- **Adaptation**:
  - RSI weight: +25% (mean reversion signals)
  - EMA weight: -10% (trends less reliable)
  - BB weight: +15% (band bounces more pronounced)
  - Result: System favors mean reversion signals

### Strong Uptrend
- **Market**: Consistent higher highs and higher lows
- **Adaptation**:
  - EMA weight: +30% (trend strength indicator)
  - RSI weight: -5% (doesn't signal in trending markets)
  - Result: System generates more buy signals

## Performance Metrics

### Testing
All components include comprehensive unit and integration tests:

```bash
# Run strategy tests
go test ./internal/strategy/... -v

# Run all project tests
go test ./... --timeout 60s
```

**Test Coverage**:
- Dynamic weights: 21 unit tests + 3 integration tests
- Symbol selection: 13 unit tests
- Integration: 8 integration tests
- **Total**: 45+ tests, 100% pass rate

### Backtesting
Test strategies against historical data:

```bash
./bin/backtest --data=testdata/sample_btc.csv \
  --symbol=BTC-USD \
  --capital=10000 \
  --take-profit=0.8 \
  --stop-loss=0.4
```

## Enabling/Disabling Features

### Run with Traditional Strategies Only
```bash
# Disable integrated engine (comment out engine.Start() in main.go)
# or use individual StrategyOrchestrator strategies
```

### Enable Dynamic Weights
```go
// In IntegratedStrategyEngine, weights are always calculated dynamically
// No configuration needed - automatic based on market conditions
```

### Enable Symbol Selection
```bash
# Symbol selection is automatic in IntegratedStrategyEngine
# All configured symbols are evaluated every 30 seconds
# Adjust refresh interval:
symbolRefreshInterval := 60 * time.Second  // Change in initializeBot()
```

### Custom Symbol List
```bash
# Override in .env
TRADING_SYMBOLS=BTC-USD,ETH-USD,SOL-USD
```

## Troubleshooting

### Symbol Selection Not Updating
- Check logs for "symbol selection updated" message
- Verify primary exchange has market data endpoints
- Confirm trading symbols configured
- Check network connectivity

### Dynamic Weights Not Adapting
- Weights are always calculated dynamically
- If seeing constant weights, check price data quality
- Ensure sufficient candle data (minimum 30 periods)

### No Signals Generated
- Verify selected symbols (check logs)
- Check indicator parameters
- Confirm price data is valid
- Check signal strength threshold settings

### High API Rate Limiting
- Increase symbol refresh interval
- Reduce number of trading symbols
- Optimize batch data fetching

## Integration with Existing Systems

### Exchange Multiplexer
The IntegratedStrategyEngine works with any Exchange implementation:
- Hyperliquid (derivatives)
- Coinbase (spot)
- dYdX (decentralized)

### Order Manager
Signals flow directly to ExecutionAgent → OrderManager:
```
IntegratedEngine.Signal → ExecutionAgent.HandleSignal 
→ OrderManager.PlaceOrder
```

### Risk Manager
Risk validation applied per signal:
- Position size limits
- Daily loss limits
- Portfolio risk metrics

## Future Enhancements

1. **Machine Learning Integration**
   - Adaptive weight optimization using historical performance
   - Neural network-based signal generation

2. **Cross-exchange Arbitrage**
   - Monitor price differences across exchanges
   - Execute arbitrage opportunities

3. **Advanced Risk Management**
   - Portfolio-level hedging
   - Correlation-based position sizing
   - Dynamic stop-loss adjustment

4. **TUI Enhancements**
   - Display selected symbols and scores
   - Show dynamic weight visualization
   - Real-time signal confidence charts

## File Structure

```
internal/strategy/
├── integration.go                    # IntegratedStrategyEngine
├── integration_test.go               # Integration tests (8 tests)
├── indicator_weights.go              # WeightCalculator
├── indicator_weights_test.go         # Weight tests (21 tests)
├── dynamic_weights_integration_test.go  # Integration tests (3 tests)
├── symbol_selector.go                # SymbolSelector
├── symbol_selector_test.go           # Symbol selection tests (13 tests)
├── signals.go                        # SignalGenerator (modified)
├── scalping.go                       # ScalpingStrategy
└── [other strategy files]

docs/
├── INTEGRATION.md                    # This file
├── DYNAMIC_WEIGHTS.md                # Weight calculation details
└── SYMBOL_SELECTION.md               # Selection algorithm details
```

## References

- [Dynamic Weights Documentation](./DYNAMIC_WEIGHTS.md)
- [Symbol Selection Documentation](./SYMBOL_SELECTION.md)
- [Strategy Guide](./QUICKSTART.md)
- [Testing Guide](TESTING_GUIDE.md)

## Support

For issues or questions:
1. Check logs for error messages
2. Review test cases for usage examples
3. See troubleshooting section above
4. Check DYNAMIC_WEIGHTS.md and SYMBOL_SELECTION.md for technical details
