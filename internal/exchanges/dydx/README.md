# dYdX v4 Integration

This directory contains the dYdX v4 exchange integration for the Constantine trading bot.

## Overview

dYdX v4 is a decentralized perpetual futures exchange built on Cosmos SDK. Unlike traditional exchanges, it requires blockchain transactions to place and cancel orders.

## Architecture

The integration uses a hybrid approach:

- **Market Data**: Native Go implementation using HTTP and WebSocket APIs
- **Order Placement**: Python wrapper around the official dYdX v4 Python client
- **Authentication**: BIP39 mnemonic-based wallet

## Prerequisites

### Python Dependencies

To enable order placement, you need to install the official dYdX v4 Python client:

```bash
pip3 install dydx-v4-client-py v4-proto
```

Or using a virtual environment (recommended):

```bash
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install dydx-v4-client-py v4-proto
```

### Verification

Test that the Python client is installed correctly:

```bash
python3 -c "from v4_client_py import NodeClient; print('dYdX Python client installed successfully!')"
```

## Configuration

### Using Environment Variables

Set your dYdX credentials in your `.env` file:

```env
# dYdX Configuration
DYDX_MNEMONIC="your twelve or twenty-four word mnemonic phrase here"
DYDX_NETWORK="testnet"  # or "mainnet"
```

### Using Code

```go
import "github.com/guyghost/constantine/internal/exchanges/dydx"

// Create client with mnemonic
client, err := dydx.NewClientWithMnemonic(mnemonic, 0)
if err != nil {
    log.Fatal(err)
}

// For testnet
testnetClient, err := dydx.NewClientWithMnemonicAndURL(
    mnemonic,
    0,  // subaccount number
    "https://indexer.v4testnet.dydx.exchange",
    "wss://indexer.v4testnet.dydx.exchange/v4/ws",
)
```

## Features

### ✅ Implemented

- **Market Data (HTTP)**:
  - Get ticker
  - Get orderbook
  - Get candles (OHLCV)
  - Get balance
  - Get positions

- **Market Data (WebSocket)**:
  - Real-time ticker updates
  - Real-time orderbook updates
  - Real-time trade feed

- **Trading (via Python client)**:
  - Place orders (market, limit, stop loss, take profit)
  - Cancel orders
  - Get balance

### ⚠️ Partially Implemented

- Get open orders (indexer API limitations)
- Get order history (indexer API limitations)

### ❌ Not Implemented

- Modify order (not supported by dYdX v4)
- Native Go transaction signing (requires proto compilation)

## Usage Examples

### Market Data

```go
ctx := context.Background()

// Get ticker
ticker, err := client.GetTicker(ctx, "BTC-USD")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("BTC-USD Price: %s\n", ticker.Last)

// Subscribe to real-time ticker
callback := func(ticker *exchanges.Ticker) {
    fmt.Printf("Price update: %s\n", ticker.Last)
}
client.SubscribeTicker(ctx, "BTC-USD", callback)
```

### Trading

```go
// Place a limit buy order
order := &exchanges.Order{
    Symbol:      "BTC-USD",
    Side:        "buy",
    Type:        "limit",
    Size:        decimal.NewFromFloat(0.01),
    Price:       decimal.NewFromFloat(50000),
    TimeInForce: "GTT", // Good-til-time
}

placedOrder, err := client.PlaceOrder(ctx, order)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order placed: %s\n", placedOrder.ID)

// Cancel order
err = client.CancelOrder(ctx, placedOrder.ID)
if err != nil {
    log.Fatal(err)
}
```

## Networks

### Mainnet

- **Indexer API**: `https://indexer.dydx.trade`
- **Indexer WebSocket**: `wss://indexer.dydx.trade/v4/ws`
- **Chain ID**: `dydx-mainnet-1`

### Testnet

- **Indexer API**: `https://indexer.v4testnet.dydx.exchange`
- **Indexer WebSocket**: `wss://indexer.v4testnet.dydx.exchange/v4/ws`
- **Chain ID**: `dydx-testnet-4`

## Order Types

### Short-Term Orders (20 blocks ~2-3 seconds)
- Market orders
- Immediate-or-cancel (IOC) orders

### Long-Term Orders (stateful)
- Limit orders with GTT (Good-til-time)
- Post-only orders

### Conditional Orders
- Stop loss orders
- Take profit orders

## Market Information

Popular perpetual markets on dYdX v4:
- `BTC-USD`
- `ETH-USD`
- `SOL-USD`
- `AVAX-USD`
- And 50+ more...

## Troubleshooting

### Python Client Not Found

**Error**: `dydx-v4-client-py not installed`

**Solution**: Install the Python client:
```bash
pip3 install dydx-v4-client-py v4-proto
```

### Mnemonic Invalid

**Error**: `invalid mnemonic`

**Solution**: Ensure your mnemonic is a valid BIP39 phrase (12 or 24 words).

### Order Placement Fails

**Error**: `Python client not initialized`

**Solution**: Use `NewClientWithMnemonic()` instead of `NewClient()`:
```go
client, err := dydx.NewClientWithMnemonic(mnemonic, 0)
```

### Script Not Found

**Error**: `Python script error: no such file or directory`

**Solution**: Ensure the script path is correct. The default is:
```
internal/exchanges/dydx/scripts/dydx_client.py
```

## Development Roadmap

### Future Enhancements

1. **Native Go Implementation**
   - Compile dYdX v4 proto definitions
   - Implement Cosmos SDK transaction signing in Go
   - Remove Python dependency

2. **Advanced Features**
   - Leverage management
   - Liquidation monitoring
   - Position averaging

3. **Performance**
   - Connection pooling
   - Request batching
   - Async order placement

## Resources

- [dYdX v4 Documentation](https://docs.dydx.xyz/)
- [dYdX v4 Python Client](https://github.com/dydxprotocol/v4-clients/tree/main/v4-client-py-v2)
- [dYdX Chain](https://github.com/dydxprotocol/v4-chain)
- [Cosmos SDK](https://docs.cosmos.network/)

## License

See main project LICENSE file.

## Support

For issues specific to this integration, please check:
1. This README
2. The main project documentation
3. dYdX official documentation
4. GitHub issues

**⚠️ Important**: This is alpha software. Always test on testnet before using real funds.
