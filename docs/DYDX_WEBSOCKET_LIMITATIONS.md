# dYdX v4 WebSocket Limitations

## Summary

**dYdX v4 WebSocket API does NOT support real-time candle/OHLCV streams.**

This is why Constantine uses REST API polling for candles instead of pure WebSocket.

---

## What dYdX WebSocket Supports

### ✅ Real-Time (via WebSocket)

**1. v4_markets (Ticker)**
```json
{
  "type": "channel_data",
  "channel": "v4_markets",
  "id": "BTC-USD",
  "contents": {
    "oraclePrice": "45234.56",
    "volume24H": "12345678.90",
    "priceChange24H": "234.56",
    "trades24H": "123456789"
  }
}
```
- Updates: Real-time (as prices update)
- Use case: Current price, 24h volume/change
- Update frequency: Every few milliseconds

**2. v4_orderbook (Order Book)**
```json
{
  "type": "channel_data",
  "channel": "v4_orderbook",
  "id": "BTC-USD",
  "contents": {
    "bids": [["45200.00", "1.5"], ["45195.00", "2.0"]],
    "asks": [["45205.00", "1.2"], ["45210.00", "3.0"]]
  }
}
```
- Updates: Real-time
- Use case: Order book depth, bid/ask spreads
- Update frequency: Every few milliseconds

**3. v4_trades (Trade Feed)**
```json
{
  "type": "channel_data",
  "channel": "v4_trades",
  "id": "BTC-USD",
  "contents": {
    "trades": [
      {
        "price": "45234.56",
        "size": "0.5",
        "side": "BUY",
        "createdAt": "2025-01-15T10:30:45Z"
      }
    ]
  }
}
```
- Updates: Real-time
- Use case: Individual trade executions
- Update frequency: Every few milliseconds

---

### ❌ NOT Available via WebSocket

**❌ Candles/OHLCV Data**
```
NO v4_candles channel
NO v4_ohlcv channel
NO aggregated price data
```

**Why?** dYdX doesn't aggregate real-time trade data into OHLCV candles on their server. This would require:
1. Maintaining state for open candles
2. Aggregating trades in real-time
3. Broadcasting completed candles
4. Managing multiple timeframes (1m, 5m, 15m, 1h, etc.)

This is expensive and not part of their WebSocket API design.

---

## How Constantine Works Around This

### Current Solution: Hybrid Approach

```
Real-time Feeds (WebSocket):
├─ Ticker updates         (every ms)
├─ Order book changes     (every ms)
└─ Individual trades      (every ms)

Candle Data (REST Polling):
└─ OHLCV candles         (every 10 seconds)
```

### Why 10 Seconds?

**Trade-offs:**
- ✅ 10s: Responsive enough for strategy (compared to original 60s)
- ✅ 10s: Low API rate usage (6 req/min per symbol)
- ✅ 10s: dYdX allows 10 req/sec (plenty of headroom)
- ❌ 10s: Not as fast as real WebSocket would be
- ❌ 10s: May miss volatile candles between polls

### Why Not Build Candles from Trades?

We COULD theoretically build candles from the real-time `v4_trades` stream:

```go
// Pseudo-code: Build candles from trades
trades := <-tradeWebSocketFeed
for trade := range trades {
    currentCandle.high = max(currentCandle.high, trade.price)
    currentCandle.low = min(currentCandle.low, trade.price)
    currentCandle.close = trade.price
    currentCandle.volume += trade.size
    
    if timeIssuedNewCandle() {
        emit(currentCandle)
        currentCandle = newCandle()
    }
}
```

**Why this isn't used:**

1. **Trade data is delayed:** Trades come via WebSocket but are aggregated, not all trades
2. **Gaps possible:** If no trades for 10+ seconds, candle would be incomplete
3. **Volume calculation:** Would include only trades on dYdX, not true 24h volume
4. **Complexity:** Adds significant buffering and state management
5. **Not necessary:** REST candles are sufficient for a 10-second polling strategy

---

## Comparison: Real-Time vs Polling

### Real-Time WebSocket (Ideal)
```
Trade occurs
  ↓ (milliseconds)
Event in WebSocket
  ↓ (milliseconds)
Candle updated
  ↓
Signal generated

Latency: <100ms
```

### Current Polling (Acceptable)
```
Current candle period
  ↓
After 10 seconds
API polls GET /candles
  ↓ (milliseconds)
Response received
  ↓
Candle processed
  ↓
Signal generated

Latency: 0-10 seconds
```

**For scalping strategy:** 10-second latency is acceptable because:
- Strategy update interval: 5 seconds
- Price changes: Slow enough to catch in 10s windows
- Not high-frequency trading (would need <100ms)

---

## dYdX API Endpoints Used

### For Candles (Polled every 10s)
```
GET /v4/candles/trades
Query params:
  marketId=BTC-USD
  resolution=1m  (or 5m, 15m, 1h, 4h, 1d)
  limit=100

Response:
  Returns [
    {"t": timestamp, "o": open, "h": high, "l": low, "c": close, "v": volume},
    ...
  ]
```

### For Ticker (Real-time WebSocket)
```
WebSocket: wss://indexer.dydx.trade/v4/ws
Subscribe:
  {"type": "subscribe", "channel": "v4_markets", "id": "BTC-USD"}
```

### For Order Book (Real-time WebSocket)
```
WebSocket: wss://indexer.dydx.trade/v4/ws
Subscribe:
  {"type": "subscribe", "channel": "v4_orderbook", "id": "BTC-USD"}
```

### For Trades (Real-time WebSocket)
```
WebSocket: wss://indexer.dydx.trade/v4/ws
Subscribe:
  {"type": "subscribe", "channel": "v4_trades", "id": "BTC-USD"}
```

---

## Impact on Strategy

### Positive
- ✅ Ticker data real-time (no 10s delay on price)
- ✅ Order book real-time (know liquidity instantly)
- ✅ Trades real-time (see market activity)
- ✅ Only candles delayed by max 10s

### Negative
- ❌ Signals based on 10-second-old candle data
- ❌ May miss intra-candle spikes
- ❌ Slightly slower response vs true real-time

### Workaround
Could use **ticker data to supplement candles**:
- Get high/low from ticker over 10s window
- Get close from latest ticker
- Would give more real-time OHLC

Current implementation doesn't do this, but could be improved.

---

## Future Improvements

### Option 1: Build Candles from Trades (Medium Effort)
```go
// Subscribe to v4_trades
// Accumulate trade data into candles
// Emit completed candles in real-time
// Complexity: Medium (state management, buffering)
// Benefit: 100ms candle latency instead of 10s
```

### Option 2: Hybrid Ticker+REST Candles (Low Effort)
```go
// Keep REST candles as foundation (OHLCV values)
// Use ticker to update Close in real-time
// Use order book to estimate High/Low
// Complexity: Low (data fusion)
// Benefit: More responsive without WebSocket complexity
```

### Option 3: Upgrade to WebSocket Candles (Not Possible)
```
Requires dYdX to add v4_candles channel
Status: Not supported by dYdX
Timeline: Unknown
```

---

## Recommendation

**Current implementation (10s REST polling) is good enough for:**
- ✅ Scalping (5-10 minute trades)
- ✅ Mean reversion (wait for RSI extremes)
- ✅ Trend following (EMA crosses)

**Would need improvement for:**
- ❌ High-frequency trading (<1 second)
- ❌ Sub-minute scalping
- ❌ Arbitrage (latency-sensitive)

Since Constantine is designed for scalping (not ultra-high-frequency), the current approach is solid.

---

## Summary

| Data Type | Method | Latency | Update Rate | Status |
|-----------|--------|---------|-------------|--------|
| Ticker | WebSocket | <100ms | Continuous | ✅ Real-time |
| Order Book | WebSocket | <100ms | Continuous | ✅ Real-time |
| Trades | WebSocket | <100ms | Per trade | ✅ Real-time |
| Candles | REST Poll | 0-10s | Every 10s | ⚠️ Acceptable |

dYdX API is not the bottleneck for strategy execution. The 10-second candle latency is acceptable for the Constantine bot's use case.
