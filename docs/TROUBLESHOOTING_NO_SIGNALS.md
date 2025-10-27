# Troubleshooting: No Signals/Trades on dYdX

## Problem Summary

Bot is running but generating no trading signals or orders on dYdX.

## Root Cause Analysis (FIXED)

### Issue 1: Candle Data Was Too Stale ✅ FIXED

**Problem:**
- dYdX SubscribeCandles was polling REST API every 60 seconds
- Strategy was updating every 1 second
- Result: 59 strategy updates with identical data, no signals

**Solution Applied:**
- Changed polling frequency to 10 seconds
- Reduced strategy update interval to 5 seconds
- Now signals can be generated with fresh market data

**Commit:** `fix(dydx): improve candle polling frequency from 1m to 10s`

---

## Diagnostic Checklist

### 1. Verify Candles Are Being Received

Watch bot logs for candle messages:
```bash
./bin/bot 2>&1 | grep "candle received"
```

**Expected output (every 10 seconds per symbol):**
```
[info] 📊 candle received symbol=BTC-USD timestamp=15:04:35 open=45123.45 high=45234.56 low=45100.00 close=45200.00 volume=123.45
```

If you **DON'T** see these messages:
- Check if exchanges are connecting
- Verify trading symbols are configured
- Check network connectivity to dYdX

### 2. Verify Strategies Are Running

Watch for strategy startup messages:
```bash
./bin/bot 2>&1 | grep -i "strategy\|subscribed"
```

**Expected:**
```
[info] subscribing to market data symbol=BTC-USD
[info] subscribed to candles
[info] subscribed to ticker
[info] subscribed to orderbook
[info] subscribed to trades
```

### 3. Check Signal Generation

Watch for signal log messages:
```bash
./bin/bot 2>&1 | grep -i "signal\|generated"
```

**Expected output when signal conditions met:**
```
[debug] strategy update symbol=BTC-USD prices_count=25 volumes_count=25 ready_for_signals=true
[info] integrated strategy signal type=ENTRY side=buy symbol=BTC-USD price=45200.00 strength=0.75
```

**If stuck at "insufficient data":**
- Strategy needs 21 candles (LongEMAPeriod) before signals
- Wait ~4 minutes (21 candles × 10 seconds each)
- Then signals should start appearing

### 4. Check Execution

If signals appear but no trades:
```bash
./bin/bot 2>&1 | grep -i "execution\|placement\|order"
```

**Expected:**
```
[info] execution error: [error description]
```

Or for successful orders:
```
[info] order placed symbol=BTC-USD side=buy
```

---

## Common Issues & Solutions

### Issue A: No Candle Messages

**Cause:** Exchanges not connected or subscribing failed

**Fix:**
1. Check exchange connectivity:
   ```bash
   ./bin/bot 2>&1 | grep -i "connected\|connection"
   ```

2. Verify trading symbols configured:
   ```bash
   export TRADING_SYMBOLS="BTC-USD,ETH-USD"
   ./bin/bot
   ```

3. Check for WebSocket errors:
   ```bash
   ./bin/bot 2>&1 | grep -i "websocket\|error"
   ```

### Issue B: Candles Received but No Signals

**Cause:** Insufficient price history (need 21 candles)

**Wait Time:**
- 21 candles × 10 seconds = 210 seconds = 3.5 minutes
- Watch logs for: `ready_for_signals=true`

### Issue C: Signals Generated but No Trades

**Cause:** Order placement not configured or failing

**Check Configuration:**
1. dYdX credentials not set:
   ```bash
   echo $DYDX_MNEMONIC  # Should show 12-word phrase
   echo $DYDX_ADDRESS   # Should show account address
   ```

2. Python client not available:
   ```bash
   python3 -c "from dydx_v4_client import Client"  # Should work
   ```

3. Enable live trading:
   ```bash
   export ENABLE_LIVE_TRADING=true
   ./bin/bot
   ```

### Issue D: Permission/Signature Errors

**Cause:** dYdX wallet setup incomplete

**Setup Steps:**
```bash
# 1. Generate wallet (if new)
python3 -c "from mnemonic import Mnemonic; print(Mnemonic('english').generate())"

# 2. Set mnemonic in .env
echo 'DYDX_MNEMONIC="word1 word2 word3..."' >> .env

# 3. Initialize bot to derive address
./bin/bot

# 4. Copy address from logs, set it
echo 'DYDX_ADDRESS="dydx1..."' >> .env
```

---

## Signal Generation Flow

```
Market Data (WebSocket/REST every 10s)
    ↓ [candle received log]
Price Accumulation (21 candles buffer)
    ↓ [ready_for_signals=true log]
Indicator Calculation (EMA, RSI, Bollinger)
    ↓
Signal Generation (Buy/Sell conditions met)
    ↓ [strategy signal log]
Risk Validation (Stop loss, take profit)
    ↓
Execution Agent (Place/Cancel orders)
    ↓ [execution log]
Position Tracking
```

---

## Minimal Test

To verify the system works end-to-end:

```bash
# 1. Run bot with logging
./bin/bot 2>&1 | tee bot.log

# 2. In another terminal, monitor signals
tail -f bot.log | grep -i "candle\|signal\|execution"

# 3. Wait 3-4 minutes for first signals
```

**Success indicators:**
1. Candle messages every 10 seconds
2. After 3.5 minutes: `ready_for_signals=true`
3. Soon after: Signal messages appear
4. Then: Execution messages (order placement or error)

---

## Performance Tuning

### If signals are too frequent:
- Increase update interval: `UpdateInterval: 10 * time.Second`
- Increase signal threshold: `MinSignalStrength: 0.70`

### If signals are too rare:
- Decrease update interval: `UpdateInterval: 3 * time.Second`
- Lower candle polling: Change `10 * time.Second` to `5 * time.Second`

### If losing signals (data gaps):
- Increase historical candles loaded: `candlesToLoad := 200`

---

## Next Steps

If signals are being generated but no trades:

1. **Check execution errors:**
   ```bash
   ./bin/bot 2>&1 | grep -i "execution error"
   ```

2. **Verify dYdX setup:**
   - Wallet created and funded
   - Address correctly configured
   - Python client working

3. **Enable paper trading first:**
   ```bash
   export ENABLE_LIVE_TRADING=false
   ./bin/bot
   ```

4. **Check order placement logs:**
   ```bash
   ./bin/bot 2>&1 | grep -i "place\|order"
   ```

---

## Support

For further issues:
- Check `docs/DYDX_INTEGRATION.md` for wallet setup
- Review `docs/WHY_NO_TRADES_YET.md` for full setup guide
- Check bot logs for specific error messages
- Ensure network connectivity to dYdX API

Good luck! 🚀
