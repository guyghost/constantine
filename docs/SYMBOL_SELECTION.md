# Intelligent Symbol Selection System

## Overview

Constantine now includes an **intelligent symbol selection system** that automatically identifies and prioritizes the most promising cryptocurrencies to trade based on objective metrics of gain potential and risk-adjusted returns.

Instead of trading a fixed set of symbols, the bot can now:
- Analyze a pool of candidate symbols
- Score each by gain potential and risk metrics
- Select the N best opportunities dynamically
- Filter by price ranges to avoid micro-cap or ultra-expensive assets

## Architecture

### Core Components

#### SymbolSelector
Calculates opportunity scores and ranks symbols:
- **CalculateGainPotential()**: Trend-based potential [0, 1]
- **CalculateRiskAssessment()**: Risk metrics [0, 1]
- **CalculateSharpeRatio()**: Risk-adjusted returns
- **CalculateOpportunityScore()**: Composite metric
- **RankSymbols()**: Sorts by opportunity
- **SelectBestSymbols()**: Returns top N symbols

#### RankedSymbol
```go
type RankedSymbol struct {
    Symbol      string              // Symbol identifier (e.g., "BTC-USD")
    Score       float64             // Composite opportunity score [0, 1]
    Potential   decimal.Decimal     // Gain potential component
    Risk        decimal.Decimal     // Risk assessment component
    SharpeRatio decimal.Decimal     // Risk-adjusted return
}
```

#### SelectionEvent
```go
type SelectionEvent struct {
    Timestamp int64               // Selection timestamp
    Symbol    string              // Selected symbol
    Score     float64             // Selection score
    Reason    string              // Selection reason
}
```

## Metrics Explained

### 1. Gain Potential [0, 1]

**Algorithm:**
```
IF EMA(9) > EMA(21):  // Uptrend
    Distance to ATH = (ATH - Price) / ATH
    Potential = 0.5 + 0.5 * (1 - Distance)  // Range [0.5, 1.0]
ELSE:  // Downtrend
    Potential = 0.25
```

**Interpretation:**
- 0.0-0.25: Bearish trend, limited upside potential
- 0.25-0.5: Weak uptrend, moderate potential
- 0.5-0.75: Good uptrend, solid potential
- 0.75-1.0: Strong uptrend near ATH, maximum potential

**Example:**
- BTC at $45,000, ATH at $50,000, EMA9 > EMA21
  - Distance = (50,000 - 45,000) / 50,000 = 0.1
  - Potential = 0.5 + 0.5 * (1 - 0.1) = 0.95 (Very high)

### 2. Risk Assessment [0, 1]

**Components:**
- Volatility (stddev of log returns): 40% weight
- Max Drawdown (peak-to-trough): 40% weight
- Volume Quality: 20% weight

**Volume Quality:**
- Avg Volume > 1000: Low risk (1.0)
- Avg Volume 100-1000: Medium risk (scaled)
- Avg Volume < 100: High risk (0.0)

**Interpretation:**
- 0.0-0.2: Very safe, low volatility
- 0.2-0.4: Safe, moderate volatility
- 0.4-0.6: Medium risk, higher volatility
- 0.6-0.8: High risk, significant volatility
- 0.8-1.0: Very high risk, extreme volatility

**Example:**
- Symbol A: Vol=5%, DD=8%, Avg Volume=2000
  - Risk = 0.05*0.4 + 0.08*0.4 + (1-1)*0.2 = 0.052 (Very low)
- Symbol B: Vol=25%, DD=35%, Avg Volume=50
  - Risk = 0.25*0.4 + 0.35*0.4 + (1-0)*0.2 = 0.30 (Moderate)

### 3. Sharpe Ratio

**Formula:**
```
Returns = [log(Price[i] / Price[i-1]) for i in 1..N]
Mean Return = sum(Returns) / N
Volatility = sqrt(variance(Returns))
Sharpe = (Mean Return - Risk Free Rate) / Volatility
```

**Risk-Free Rate:** 0.01 (1% annual assumption)

**Interpretation:**
- < 0: Worse than risk-free rate
- 0-1: Moderate risk-adjusted returns
- 1-3: Good risk-adjusted returns
- > 3: Excellent risk-adjusted returns

### 4. Opportunity Score [0, 1]

**Composite Formula:**
```
Score = (Potential * 0.40) + 
        (Sharpe * 0.35) + 
        (Volume Quality * 0.15) - 
        (Risk * 0.10)

Final = clamp(Score, 0, 1)
```

**Weights:**
- Gain Potential: 40% (trend most important)
- Sharpe Ratio: 35% (risk-adjusted returns)
- Volume: 15% (liquidity confirmation)
- Risk Penalty: -10% (subtract for high risk)

## Usage Examples

### Example 1: Ranking Multiple Symbols

```go
selector := NewSymbolSelector(config)

// Prepare data for all symbols
symbolData := map[string]SymbolData{
    "BTC-USD": {Prices: btcPrices, Volumes: btcVolumes},
    "ETH-USD": {Prices: ethPrices, Volumes: ethVolumes},
    "ADA-USD": {Prices: adaPrices, Volumes: adaVolumes},
}

// Rank all symbols
ranked := selector.RankSymbols([]string{"BTC-USD", "ETH-USD", "ADA-USD"}, symbolData)

for i, r := range ranked {
    println(i+1, r.Symbol, r.Score, r.Potential, r.Risk)
}
```

### Example 2: Selecting Top 3 Symbols

```go
// Select top 3 with dynamic threshold
selected := selector.SelectBestSymbols(
    symbols,
    symbolData,
    3,  // Max count
)

for _, sym := range selected {
    println("Trading:", sym.Symbol, "Score:", sym.Score)
}
```

### Example 3: Filtering by Price Range

```go
// Only trade symbols between $10 and $1000
minPrice := decimal.NewFromFloat(10)
maxPrice := decimal.NewFromFloat(1000)

filtered := selector.FilterByPriceRange(
    symbols,
    symbolData,
    minPrice,
    maxPrice,
)
```

## Selection Strategies

### Conservative Strategy
```go
// Select only symbols with scores > 0.6
threshold := 0.6
selected := make([]RankedSymbol, 0)
for _, r := range ranked {
    if r.Score > threshold {
        selected = append(selected, r)
    }
}
```

### Balanced Strategy
```go
// Select top 5 symbols with dynamic threshold
selected := selector.SelectBestSymbols(symbols, data, 5)
```

### Aggressive Strategy
```go
// Select all symbols scoring > 0.4
threshold := 0.4
// Filter ranked symbols
```

## Integration with Bot

### Current Integration Points
1. Strategy Agent can receive selected symbols
2. Backtesting can test selection on historical data
3. TUI can display selected symbols and their scores

### Future Integration
1. Auto-refresh selection every N candles
2. Correlation analysis to avoid redundant picks
3. Portfolio-level optimization
4. Machine learning to optimize weights

## Performance Characteristics

- **Calculation Speed**: ~1-10ms for 100 symbols
- **Memory**: Minimal (only current data stored)
- **Thread Safety**: Full RWMutex protection
- **History**: Last 100 selection events retained

## Configuration

### Adjustable Parameters

```go
// Modify these in SymbolSelector for custom behavior
Gain Potential Weight: 0.40  // Increase for trend-focused
Sharpe Ratio Weight: 0.35    // Increase for risk-adjusted focus
Volume Weight: 0.15          // Increase to prioritize liquidity
Risk Penalty: 0.10           // Increase to be more conservative

Dynamic Threshold: 0.6       // Minimum quality threshold
Min Symbols: 1               // Ensure at least N symbols selected
```

## Test Coverage

**13 Comprehensive Tests:**
- Gain potential calculation
- Risk assessment
- Sharpe ratio computation
- Opportunity score weighting
- Symbol ranking and sorting
- Best symbol selection
- Price range filtering
- Volatility impact on scores
- Volume impact on scores
- Dynamic threshold calculation
- Consistency across runs
- Edge cases (minimal data, single symbol)

**All tests: PASS (100%)**

## Troubleshooting

### All symbols have low scores (< 0.3)
- **Cause**: Market conditions poor or data insufficient
- **Solution**: Lower threshold or wait for better opportunities

### Score doesn't change
- **Cause**: Need more historical data (minimum 30 candles)
- **Solution**: Ensure sufficient price/volume history

### One symbol always selected
- **Cause**: Other symbols lack data or have poor metrics
- **Solution**: Verify data for all symbols, check for NaN/Inf values

## Future Enhancements

1. **Correlation Analysis**: Avoid correlated pairs
2. **ML Optimization**: Learn optimal weights from backtests
3. **Sector Rotation**: Track sector opportunities
4. **Market Regime Detection**: Adjust selection for bull/bear/range
5. **Risk Parity**: Weight portfolio by volatility
6. **Momentum Ranking**: Add momentum to gain potential
7. **Sentiment Analysis**: Incorporate social signals

## References

- Sharpe Ratio: https://en.wikipedia.org/wiki/Sharpe_ratio
- EMA Calculation: https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average
- Log Returns: https://en.wikipedia.org/wiki/Rate_of_return#Logarithmic_return
- Maximum Drawdown: https://en.wikipedia.org/wiki/Drawdown_(economics)
