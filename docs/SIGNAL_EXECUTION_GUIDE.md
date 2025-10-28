# Signal Execution Testing Guide

## Overview

This guide explains how to test trading signals on dYdX using the Constantine bot. It covers:

1. **Artificial Signal Testing** (Safe, no real trades)
2. **Real Signal Testing** (With mock order placement)
3. **Live Trading** (Real orders on dYdX)

## Part 1: Artificial Signal Testing

### Quick Start

```bash
# Build the test tool
go build -o bin/test-signal cmd/test-signal/main.go

# Run the test
./bin/test-signal
```

### What Happens

The tool:
1. Connects to dYdX
2. Fetches BTC-USD market price
3. Checks account balance (uses 5000 USDC mock if insufficient)
4. **Creates an artificial BUY signal with 75% confidence**
5. Runs through ExecutionAgent → RiskManager → OrderManager
6. Places a mock order (no actual trade)
7. Shows position sizing and P&L targets

### Expected Output

```
✅ Signal created: BUY with confidence 75%
✅ Ordre placé en mock:
   Amount: 0.0437 BTC
   Price:  $114,440.78
   StopLoss:   $113,296.37
   TakeProfit: $116,729.59
```

### Understanding the Output

| Field | Meaning |
|-------|---------|
| Amount | How much BTC to buy (calculated from risk formula) |
| Price | Entry price (current market price) |
| StopLoss | Price to exit if trade goes wrong (1% below entry) |
| TakeProfit | Price to exit with profit (2% above entry) |

## Part 2: Real Signal Testing

### Testing with Actual Strategy Signals

To test with real market-based signals instead of artificial ones:

#### Option A: Use the Scalping Strategy

```go
// In cmd/test-signal/main.go, replace ÉTAPE 4

import (
    "github.com/guyghost/constantine/internal/strategy"
)

// Instead of artificial signal:
// artificialSignal := &strategy.Signal{ ... }

// Use real strategy:
sg := strategy.NewSignalGenerator(config)
prices := []decimal.Decimal{ /* fetch from dYdX */ }
volumes := []decimal.Decimal{ /* fetch from dYdX */ }
orderbook := &exchanges.OrderBook{ /* fetch from dYdX */ }

signal := sg.GenerateSignal("BTC-USD", prices, volumes, orderbook)
```

#### Option B: Connect to Live Market Data

```go
// Subscribe to real-time candles
candles := make(chan *exchanges.Candle)
go func() {
    client.SubscribeToCandles(ctx, "BTC-USD", "1m", candles)
}()

// Process incoming candles
for candle := range candles {
    signal := sg.GenerateSignal(...)
    executionAgent.HandleSignal(ctx, signal)
    
    if signal.Type == strategy.SignalTypeEntry {
        // Real order would be placed here
    }
}
```

### Testing Different Signal Types

#### 1. Buy Signal Test

```go
signal := &strategy.Signal{
    Type:      strategy.SignalTypeEntry,
    Side:      exchanges.OrderSideBuy,
    Symbol:    "BTC-USD",
    Price:     decimal.NewFromFloat(114440.78),
    Strength:  0.75,
    Reason:    "EMA crossover + RSI oversold",
}
```

#### 2. Sell Signal Test

```go
signal := &strategy.Signal{
    Type:      strategy.SignalTypeEntry,
    Side:      exchanges.OrderSideSell,
    Symbol:    "BTC-USD",
    Price:     decimal.NewFromFloat(116000.00),
    Strength:  0.80,
    Reason:    "EMA crossover + RSI overbought",
}
```

#### 3. Exit Signal Test

```go
signal := &strategy.Signal{
    Type:      strategy.SignalTypeExit,
    Side:      exchanges.OrderSideBuy,
    Symbol:    "BTC-USD",
    Price:     decimal.NewFromFloat(115500.00),
    Strength:  0.90,
    Reason:    "Take profit reached",
}
```

## Part 3: Transitioning to Real Orders

### Step 1: Replace MockOrderManager

**Before** (test-signal/main.go):
```go
mockOrderManager := NewMockOrderManager()
executionAgent := execution.NewExecutionAgent(mockOrderManager, ...)
```

**After**:
```go
// Use the real dYdX client instead
realOrderManager := order.NewManager(client, config)
executionAgent := execution.NewExecutionAgent(realOrderManager, ...)
```

### Step 2: Setup dYdX Python Client

The dYdX Python client is required for order placement:

```bash
# Install the Python client
pip3 install dydx-v4-client-py

# Verify installation
python3 -c "import v4_client_py; print('OK')"
```

### Step 3: Configure Order Placement

Update configuration for real trades:

```go
// Set execution config for real trading
executionConfig := execution.Config{
    StopLossPercent:   decimal.NewFromFloat(0.005),  // 0.5%
    TakeProfitPercent: decimal.NewFromFloat(0.01),   // 1%
    MinSignalStrength: 0.5,                          // 50% minimum
    AutoExecute:       false,  // Require manual confirmation first
}
```

### Step 4: Add Manual Confirmation

```go
// Add user confirmation before executing real trades
fmt.Println("Trade Signal Received:")
fmt.Printf("  Type: %s\n", signal.Type)
fmt.Printf("  Side: %s\n", signal.Side)
fmt.Printf("  Amount: %s\n", positionSize)
fmt.Printf("  Price: %s\n", signal.Price)
fmt.Print("\nExecute trade? (yes/no): ")

var response string
fmt.Scanln(&response)

if response != "yes" {
    fmt.Println("Trade rejected by user")
    return
}

// Execute the trade
executionAgent.HandleSignal(ctx, signal)
```

### Step 5: Implement Order Monitoring

```go
// Monitor order status
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        positions, err := client.GetPositions(ctx)
        if err != nil {
            log.Printf("Error fetching positions: %v", err)
            continue
        }
        
        for _, pos := range positions {
            fmt.Printf("[%s] %s %s @ %s | PnL: %s\n",
                time.Now().Format("15:04:05"),
                pos.Symbol,
                pos.Side,
                pos.EntryPrice,
                pos.UnrealizedPnL)
        }
    }
}
```

## Part 4: Testing Checklist

### Before First Real Trade

- [ ] **Paper Trading**: Run with mock orders first (10+ signals)
- [ ] **Signal Quality**: Verify signal logic works correctly
- [ ] **Risk Limits**: Test that risk manager enforces position limits
- [ ] **Stop Loss**: Confirm stop loss is placed automatically
- [ ] **Take Profit**: Verify take profit target is set
- [ ] **Balance**: Have sufficient balance for position sizing
- [ ] **Network**: Test dYdX connection stability
- [ ] **Python Client**: Ensure Python client is installed and working

### During Testing

- [ ] Monitor for execution errors
- [ ] Check position sizing calculations
- [ ] Verify order IDs are received
- [ ] Confirm positions appear in account
- [ ] Track PnL updates
- [ ] Monitor for API rate limits
- [ ] Test order cancellation

### After Each Trade

- [ ] Verify trade on dYdX: https://v4.testnet.dydx.exchange/
- [ ] Check position entry price
- [ ] Confirm stop loss and take profit levels
- [ ] Monitor PnL changes
- [ ] Review execution logs
- [ ] Document any issues

## Part 5: Risk Management Testing

### Test Position Size Calculation

```go
// Given:
accountBalance := decimal.NewFromFloat(5000)
riskPerTrade := decimal.NewFromFloat(1)  // 1%
entryPrice := decimal.NewFromFloat(114440.78)
stopLoss := decimal.NewFromFloat(113296.37)

// Calculate:
riskAmount := accountBalance.Mul(riskPerTrade).Div(decimal.NewFromInt(100))
priceDiff := entryPrice.Sub(stopLoss).Abs()
positionSize := riskAmount.Div(priceDiff)

// Verify position size is reasonable
fmt.Printf("Position Size: %.6f BTC\n", positionSize)  // Should be ~0.0437
```

### Test Risk Limit Enforcement

```go
// Test 1: Max position size exceeded
riskConfig.MaxPositionSize = decimal.NewFromFloat(100)  // Too small
// Result: ExecutionAgent rejects order ✓

// Test 2: Max positions reached
riskConfig.MaxPositions = 1
// With 1 open position, next order rejected ✓

// Test 3: Daily loss limit exceeded
riskManager.dailyPnL = decimal.NewFromFloat(-150)
riskConfig.MaxDailyLoss = decimal.NewFromFloat(100)
// Result: CanTrade() returns false ✓
```

### Test Multiple Symbols

```go
signals := []struct {
    symbol   string
    price    decimal.Decimal
    strength float64
}{
    {"BTC-USD", decimal.NewFromFloat(114440.78), 0.75},
    {"ETH-USD", decimal.NewFromFloat(2500.00), 0.70},
    {"SOL-USD", decimal.NewFromFloat(180.00), 0.65},
}

for _, s := range signals {
    signal := &strategy.Signal{
        Symbol:    s.symbol,
        Price:     s.price,
        Strength:  s.strength,
        // ...
    }
    
    executionAgent.HandleSignal(ctx, signal)
}
```

## Part 6: Troubleshooting

### Order Not Placed

**Symptoms**: Signal created but no order appears

**Causes**:
1. RiskManager rejected order (check CanTrade())
2. Position size too large
3. Signal strength too low
4. Insufficient balance

**Fix**:
```go
// Add detailed logging
canTrade, reason := riskManager.CanTrade()
fmt.Printf("Can trade: %v (%s)\n", canTrade, reason)

err := riskManager.ValidateOrder(req, openPositions)
if err != nil {
    fmt.Printf("Order validation failed: %v\n", err)
}
```

### Position Size Wrong

**Symptoms**: Order amount doesn't match expectation

**Causes**:
1. RiskPerTrade incorrect
2. Stop loss not calculated properly
3. Position cap applied

**Fix**:
```go
// Debug position sizing
fmt.Printf("Risk per trade: %.2f%%\n", riskPerTrade)
fmt.Printf("Risk amount: %s USD\n", riskAmount)
fmt.Printf("Price difference: %s\n", priceDiff)
fmt.Printf("Calculated position size: %s\n", positionSize)
```

### dYdX Connection Fails

**Symptoms**: Cannot connect to dYdX

**Causes**:
1. Mnemonic invalid
2. Network unreachable
3. Python client not installed
4. Wrong endpoint URL

**Fix**:
```bash
# Verify mnemonic format
grep DYDX_MNEMONIC .env | wc -w  # Should be ~15 words

# Check network
curl -s https://indexer.dydx.trade/health

# Test Python client
python3 -c "from v4_client_py import *; print('OK')"

# Use correct endpoint
# Testnet: https://indexer.v4testnet.dydx.exchange
# Mainnet: https://indexer.dydx.trade
```

## Part 7: Production Deployment

### Pre-Deployment Checklist

- [ ] 50+ test signals executed successfully
- [ ] All risk checks working correctly
- [ ] Order placement working reliably
- [ ] Position tracking accurate
- [ ] Stop loss and take profit verified
- [ ] PnL calculations correct
- [ ] Error handling robust
- [ ] Logging comprehensive

### Deployment Steps

1. **Test on testnet first** (5-10 real orders)
2. **Monitor for 24 hours** (check for any issues)
3. **Increase capital gradually** (start with small position sizes)
4. **Enable live trading** (set AutoExecute = true)
5. **Monitor continuously** (logs, metrics, positions)

### Monitoring in Production

```go
// Log every signal
logger.Component("trading").Info("signal received",
    "symbol", signal.Symbol,
    "type", signal.Type,
    "side", signal.Side,
    "price", signal.Price,
    "strength", signal.Strength)

// Log every order
logger.Component("trading").Info("order placed",
    "symbol", order.Symbol,
    "side", order.Side,
    "amount", order.Amount,
    "price", order.Price,
    "id", order.ID)

// Monitor positions
logger.Component("trading").Info("position opened",
    "symbol", pos.Symbol,
    "side", pos.Side,
    "entry", pos.EntryPrice,
    "pnl", pos.UnrealizedPnL)
```

## References

- [Constantine Architecture](./AGENTS.md)
- [Risk Management Configuration](./TRADING_RULES.md)
- [dYdX v4 API Documentation](https://dydx.exchange/api)
- [Python Client Repository](https://github.com/dydxprotocol/v4-clients/tree/main/v4-client-py)
- [Test Signal README](../cmd/test-signal/README.md)

## Support

For issues or questions:
1. Check the [troubleshooting section](./TROUBLESHOOTING.md)
2. Review [trading rules documentation](./TRADING_RULES.md)
3. Check Constantine [developer guide](./DEVELOPER.md)
4. Open an issue on GitHub

---

**Last Updated**: October 28, 2025  
**Status**: Active  
**Maintained By**: Constantine Bot Team
