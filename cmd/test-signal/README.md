# Test Signal - Artificial Buy Signal Execution Test

## Overview

This tool tests the complete trading signal execution flow on dYdX with an artificially generated buy signal for BTC-USD. It demonstrates:

1. **Connection** to dYdX testnet/mainnet
2. **Market Data** retrieval (ticker prices)
3. **Account Balance** checking
4. **Signal Generation** - artificially created buy signal with 75% confidence
5. **Risk Management** validation
6. **Order Execution** - creating a mock order through the execution agent

## Building

```bash
cd constantine
go build -o bin/test-signal cmd/test-signal/main.go
```

## Running

### Prerequisites

1. Set up environment variables in `.env`:
   ```
   DYDX_MNEMONIC="your 12-word mnemonic here"
   ```

2. (Optional) Add API key for testnet:
   ```
   DYDX_API_KEY="your-api-key"
   ```

### Execution

```bash
./bin/test-signal
```

## Output Example

```
🚀 Test Signal d'Achat Artificiel - dYdX BTC-USD
================================================

📡 ÉTAPE 1: Connexion à dYdX...
✅ Connecté à dYdX

💰 ÉTAPE 2: Récupération des prix BTC-USD...
✅ Prix BTC-USD: 114440.77744
   Bid: 0, Ask: 0

💵 ÉTAPE 3: Configuration du solde...
⚠️  Solde réel insuffisant (21.261754 USDC), utilisation de 5000 USDC pour le test
   Solde utilisé: 5000 USDC (mocké)

📊 ÉTAPE 4: Création du signal d'achat artificiel...
✅ Signal créé:
   Type:      entry
   Side:      buy
   Symbol:    BTC-USD
   Price:     114440.77744
   Strength:  0.75
   Reason:    Signal artificiel de test

⚙️  ÉTAPE 5: Initialisation des managers...
✅ Managers initialisés

🎯 ÉTAPE 6: Exécution du signal via ExecutionAgent...
✅ Ordre placé en mock:
   ID:     MOCK-1761641791430043000
   Symbol: BTC-USD
   Side:   buy
   Amount: 0.0436907203170779
   Price:  114440.77744
   StopLoss:   113296.3696656
   TakeProfit: 116729.5929888

✅ Signal exécuté avec succès!
```

## What Happens Step-by-Step

### 1. Connection (ÉTAPE 1)
- Creates a dYdX client using the mnemonic
- Connects to dYdX testnet/mainnet

### 2. Market Data (ÉTAPE 2)
- Fetches the current BTC-USD ticker
- Shows bid/ask/last prices

### 3. Balance Check (ÉTAPE 3)
- Retrieves real account balance from dYdX
- If insufficient (<100 USDC), uses 5000 USDC mock for testing
- Displays the balance being used

### 4. Signal Creation (ÉTAPE 4)
- Creates an artificial buy signal with:
  - Type: Entry
  - Side: Buy
  - Symbol: BTC-USD
  - Price: Current market price
  - Strength: 0.75 (75% confidence)
  - Reason: "Signal artificiel de test"

### 5. Risk Manager Setup (ÉTAPE 5)
- Initializes risk management configuration:
  - MaxPositionSize: $10,100
  - MaxPositions: 3
  - MaxDrawdown: 20%
  - RiskPerTrade: 1%
  - etc.

### 6. Signal Execution (ÉTAPE 6)
- ExecutionAgent receives the signal
- Risk Manager validates:
  - Trading is allowed
  - Position size is valid
  - Stop loss & take profit are set
  - Risk per trade is acceptable
- OrderManager places the order
- Mock order is created with calculated amounts

## Key Calculations

### Position Sizing
```
riskAmount = accountBalance * RiskPerTrade / 100
           = 5000 * 1 / 100 = 50 USD

priceDiff = entryPrice - stopLoss
          = 114440.78 - 113296.37 = 1144.41

positionSize = riskAmount / priceDiff
             = 50 / 1144.41 = 0.0437 BTC

position Value = positionSize * entryPrice
               = 0.0437 * 114440.78 ≈ 5000 USD
```

### Stop Loss & Take Profit
```
StopLoss = entryPrice * (1 - StopLossPercent)
         = 114440.78 * (1 - 0.01) = 113296.37

TakeProfit = entryPrice * (1 + TakeProfitPercent)
           = 114440.78 * (1 + 0.02) = 116729.59
```

## Risk Manager Configuration

The test uses these risk parameters:

| Parameter | Value | Purpose |
|-----------|-------|---------|
| MaxPositionSize | $10,100 | Maximum USD value per position |
| MaxPositions | 3 | Maximum concurrent open positions |
| MaxDrawdown | 20% | Maximum portfolio drawdown |
| RiskPerTrade | 1% | Risk allocation per trade (% of balance) |
| MinAccountBalance | $10 | Minimum balance to trade |
| MaxExposurePerSymbol | 100% | Max exposure to single symbol |
| MaxSameSymbolPositions | 2 | Max positions per symbol |

## Failure Modes Handled

The test demonstrates how the system handles:

1. **Insufficient Balance** → Uses mock balance of 5000 USDC
2. **API Errors** → Gracefully continues with default values
3. **Risk Violations** → Execution agent rejects unsafe orders
4. **Position Limits** → Risk manager enforces concurrent position limits

## Next Steps

### Real Trading Setup
To use this with real orders on dYdX testnet:

1. Get testnet USDC from the faucet at https://v4.testnet.dydx.exchange/
2. Modify the mock order manager to use the real dYdX client
3. Ensure the Python client is installed:
   ```bash
   pip3 install dydx-v4-client-py
   ```

### Integration Testing
- Test with actual strategy signals (not artificial)
- Test sell signals and position closing
- Test multiple positions simultaneously
- Monitor for execution failures and retries

## Architecture

```
Signal Creation
    ↓
ExecutionAgent.HandleSignal()
    ↓
RiskManager.CanTrade()  ← Check if trading allowed
    ↓
RiskManager.ValidateOrder()  ← Validate position parameters
    ↓
RiskManager.CalculatePositionSize()  ← Calculate amount to buy
    ↓
OrderManager.PlaceOrder()  ← Create order (mock in test)
    ↓
Order Result
```

## Troubleshooting

### "DYDX_MNEMONIC not defined"
- Add to .env file or export as environment variable
- Mnemonic must be 12+ words

### "No subaccounts found"
- Mnemonic is valid but doesn't have a testnet subaccount
- Create one at https://v4.testnet.dydx.exchange/

### "Position size exceeds maximum"
- Increase MaxPositionSize in test
- Or reduce RiskPerTrade percentage
- Or use lower account balance for testing

### "maximum positions for symbol reached"
- MaxSameSymbolPositions limit exceeded
- Close existing positions for the symbol
- Or increase MaxSameSymbolPositions

## References

- dYdX v4 API: https://dydx.exchange/api
- Constantine Architecture: ../../AGENTS.md
- Risk Management: ../../docs/TRADING_RULES.md
