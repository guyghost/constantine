# Detailed Diagnostic: Why No Signals Are Being Generated

## Quick Test Sequence

Run these tests in order to find exactly where the signal generation flow breaks:

### Test 1: Can dYdX API be reached?

```bash
# Test connectivity to dYdX REST API
curl -s https://dydx.trade/v4/candles/trades \
  -H "Accept: application/json" | head -50
```

**Expected:** JSON response with market data (not error 404 or timeout)

---

### Test 2: Are candles being polled?

```bash
# Run isolated candle polling test (30 seconds)
./test_dydx_candles.sh
```

**Expected:** 3 candles received (one every ~10 seconds)
**If fails:** Network issue or dYdX API problem

---

### Test 3: Is bot connecting properly?

```bash
# Run bot and watch for connection messages
LOG_LEVEL=info ./bin/bot 2>&1 | head -50
```

**Look for:**
```
[info] exchange enabled exchange=dydx
[info] Connected to exchange dydx
[info] subscribing to market data symbol=BTC-USD
[info] subscribed to candles
```

**If missing:** Exchange connection failed

---

### Test 4: Are candles arriving?

```bash
# Run bot and filter for candle messages (should appear every 10s)
LOG_LEVEL=info ./bin/bot 2>&1 | grep "candle received" &
sleep 30  # Wait 30 seconds
```

**Expected:** 3 "candle received" messages
**If none:** Candle polling not working

---

### Test 5: Is data accumulating?

```bash
# Run bot and look for "ready_for_signals" (appears after 21 candles)
LOG_LEVEL=debug ./bin/bot 2>&1 | grep "ready_for_signals" &
sleep 220  # Wait 3.5 minutes (21 candles × 10 seconds)
```

**Expected after 3.5 min:** `ready_for_signals=true`
**If not:** Data not being accumulated

---

### Test 6: Are signals being generated?

```bash
# Run bot and look for signal messages
LOG_LEVEL=info ./bin/bot 2>&1 | grep "strategy signal" &
sleep 300  # Wait 5 minutes
```

**Expected:** Signal messages when conditions met
**If none:** Strategy not generating signals

---

## Detailed Flow Diagram

```
┌─────────────────────────────────────────────┐
│ Bot Starts (./bin/bot)                      │
└────────────────┬────────────────────────────┘
                 │
                 ▼
        ┌────────────────┐
        │ Load Config    │
        └────────┬───────┘
                 │
                 ▼
        ┌────────────────────────────┐
        │ Auto-select Symbols        │
        │ (BTC-USD, ETH-USD, etc)    │
        └────────┬───────────────────┘
                 │
                 ▼
        ┌────────────────────────────┐
        │ Create Exchanges           │
        │ (Hyperliquid, dYdX, etc)   │
        └────────┬───────────────────┘
                 │
                 ▼
        ┌────────────────────────────┐
        │ Connect Exchanges          │◄─ [TEST 3]
        │ (WebSocket, REST)          │
        └────────┬───────────────────┘
                 │
                 ▼
        ┌────────────────────────────┐
        │ Start Strategies           │
        │ (per symbol)               │
        └────────┬───────────────────┘
                 │
                 ▼
        ┌────────────────────────────┐
        │ Subscribe to Candles       │◄─ [TEST 4]
        │ (polling every 10s)        │
        └────────┬───────────────────┘
                 │
                 ▼
    ┌────────────────────────────────────┐
    │ Candle Loop (10s poll interval)    │
    │ ┌──────────────────────────────┐   │
    │ │ Receive Candle              │   │◄─ [TEST 5]
    │ │ Add to price buffer         │   │
    │ │ if prices.length < 21:      │   │
    │ │    → wait for more          │   │
    │ │ else:                       │   │
    │ │    → READY FOR SIGNALS      │   │
    │ └──────────────────────────────┘   │
    └────────┬─────────────────────────────┘
             │
             ▼
    ┌────────────────────────────────────┐
    │ Strategy Update (every 5s)         │
    │ ┌──────────────────────────────┐   │
    │ │ Calculate Indicators:        │   │
    │ │ - EMA 9                      │   │
    │ │ - EMA 21                     │   │
    │ │ - RSI 14                     │   │
    │ │ - Bollinger Bands            │   │
    │ │                              │   │
    │ │ Check Buy Signal:            │   │
    │ │ if EMA9 > EMA21 AND          │   │
    │ │    RSI < 35 AND              │   │
    │ │    Price < Lower BB:         │   │
    │ │    → GENERATE BUY SIGNAL     │   │
    │ │                              │   │
    │ │ Check Sell Signal:           │   │
    │ │ if EMA9 < EMA21 AND          │   │
    │ │    RSI > 70 AND              │   │
    │ │    Price > Upper BB:         │   │
    │ │    → GENERATE SELL SIGNAL    │   │
    │ └──────────────────────────────┘   │◄─ [TEST 6]
    └────────┬─────────────────────────────┘
             │
             ▼
    ┌────────────────────────────────────┐
    │ Execution Agent (if signal)        │
    │ ┌──────────────────────────────┐   │
    │ │ Validate Risk Parameters     │   │
    │ │ Check Balance                │   │
    │ │ Place Order                  │   │
    │ │ Track Position               │   │
    │ └──────────────────────────────┘   │
    └────────────────────────────────────┘
```

---

## Common Failure Points & Solutions

### ❌ Problem: No Exchanges Connecting

**Symptoms:**
```
[error] failed to initialize bot: failed to connect to dydx
[error] failed to connect websocket: ...
```

**Causes:**
1. Network not reachable to dYdX
2. dYdX WebSocket URL wrong
3. API credentials malformed

**Check:**
```bash
# Test network
ping -c 1 dydx.trade

# Check env
echo $DYDX_API_KEY
echo $DYDX_API_SECRET
echo $DYDX_MNEMONIC
```

---

### ❌ Problem: Strategies Not Starting

**Symptoms:**
```
# No "subscribing to market data" messages
# No "subscribed to candles" messages
```

**Causes:**
1. `TRADING_SYMBOLS` empty or wrong format
2. StrategyOrchestrator failed to initialize
3. Symbol not valid on dYdX

**Check:**
```bash
# Verify symbols configured
echo $TRADING_SYMBOLS

# Or run with explicit symbols
TRADING_SYMBOLS="BTC-USD,ETH-USD" ./bin/bot
```

---

### ❌ Problem: Candles Not Arriving (Most Common!)

**Symptoms:**
```
[info] subscribed to candles
[info] 📊 candle received...     ← MISSING!
```

**Root Cause:** dYdX REST API polling not returning data

**Debug:**
```bash
# Test candle polling directly
./test_dydx_candles.sh

# If no candles:
# Check API health: https://status.dydx.trade/
# Check rate limits: dYdX allows 10 req/sec
```

**Solutions:**
1. Verify dYdX API is up: https://status.dydx.trade/
2. Wait if rate limited (response code 429)
3. Check network firewall/proxy blocking dYdX

---

### ❌ Problem: Data Accumulated But No Signals

**Symptoms:**
```
[debug] ready_for_signals=true
[debug] strategy update symbol=BTC-USD prices_count=21
# But no "strategy signal" messages
```

**Causes:**
1. Market conditions don't meet signal criteria
2. RSI never goes below 35 (not oversold)
3. RSI never goes above 70 (not overbought)
4. EMA9 and EMA21 don't cross

**This is NORMAL behavior!**
- Signals only generate when specific conditions met
- May take hours or days depending on market

**To verify strategy works:**
```bash
# Check historical backtest generates signals
./bin/backtest --symbol=BTC-USD --generate-sample --verbose
```

---

### ❌ Problem: Signals Generated But No Trades

**Symptoms:**
```
[info] integrated strategy signal type=ENTRY side=buy symbol=BTC-USD
[error] execution error: ...
```

**Causes:**
1. dYdX wallet not set up
2. No USDC balance
3. Order placement failed
4. Python client not available

**Check:**
```bash
# Verify wallet setup
echo $DYDX_MNEMONIC    # Should be 12 words
echo $DYDX_ADDRESS     # Should be dydx1...

# Verify Python client
python3 -c "from dydx_v4_client import Client; print('OK')"

# Check logs for specific error
./bin/bot 2>&1 | grep -i "execution error"
```

---

## Environment Variables Checklist

For signals to be generated, these must be set:

```bash
# Required for dYdX
DYDX_API_KEY=""           # Read-only API key (optional)
DYDX_API_SECRET=""        # Or use mnemonic instead
DYDX_MNEMONIC=""          # 12-word BIP39 phrase (for trading)

# Optional symbol config
TRADING_SYMBOLS="BTC-USD,ETH-USD"  # Otherwise auto-selects

# Optional logging
LOG_LEVEL=info            # info, debug, warn, error
LOG_FORMAT=text           # text or json

# Optional trading
ENABLE_LIVE_TRADING=false # false=paper trading, true=real
```

---

## Step-by-Step Debug Process

```bash
# 1. Start with clean environment
export LOG_LEVEL=info
export TRADING_SYMBOLS="BTC-USD"

# 2. Run bot and capture first 60 seconds
timeout 60 ./bin/bot 2>&1 | tee bot_test.log

# 3. Check each milestone
echo "=== Exchanges ==="
grep "exchange enabled\|Connected" bot_test.log

echo "=== Strategies ==="
grep "subscribing\|subscribed" bot_test.log

echo "=== Candles ==="
grep "candle received" bot_test.log | wc -l

echo "=== Readiness ==="
grep "ready_for_signals" bot_test.log

echo "=== Errors ==="
grep -i "error\|failed" bot_test.log
```

---

## Expected Timeline

If everything works:

```
T+0s     Bot starts
T+2s     Exchanges connecting...
T+3s     Strategies starting...
T+5s     First candle arriving (📊 candle received)
T+15s    Second candle
T+25s    Third candle
...
T+210s   21st candle = ready_for_signals=true
T+215s   Strategy analysis starts generating signals
T+300s   First signal should appear (if market conditions met)
```

If you don't see candle messages by T+15s → **Candle polling issue** (See Test 4)
If you don't see ready flag by T+210s → **Data accumulation issue** (See Test 5)
If you see ready flag but no signals → **Market conditions not met** (normal, see Test 6)

---

## Next Steps

1. Run **Test 1** (curl) to verify API reachable
2. Run **Test 2** (candle polling) to verify polling works
3. Run **Test 3-4** (bot startup) to verify bot connects and receives candles
4. Wait **3.5 minutes** for data accumulation
5. Look for signals after data ready

If you get stuck at any test, post the output and error message!
