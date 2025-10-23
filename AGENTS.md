# Agents Architecture

## Overview
Constantine is a multi-exchange cryptocurrency trading bot built with a modular agent-based architecture. Each component acts as an autonomous agent responsible for specific aspects of the trading system.

## Core Agents

### 1. Exchange Agents
Exchange agents provide a unified interface for interacting with different cryptocurrency exchanges.

#### Hyperliquid Agent (`internal/exchanges/hyperliquid/client.go`)
- **Type**: Derivatives Exchange Client
- **Authentication**: Ethereum private key-based
- **Capabilities**:
  - Real-time candle data via WebSocket
  - REST API order placement
  - Account balance queries
  - Position management
- **Key Methods**:
  - `Connect()` - Establishes WebSocket connection
  - `SubscribeToCandles()` - Streams real-time OHLCV data
  - `PlaceOrder()` - Executes market/limit orders
  - `GetPositions()` - Retrieves active positions
  - `GetBalance()` - Fetches account balance

#### Coinbase Agent (`internal/exchanges/coinbase/client.go`)
- **Type**: Spot Exchange Client
- **Authentication**: API Key/Secret
- **Capabilities**:
  - REST-based order execution
  - Limit order support
  - Account balance tracking
  - Position queries
- **Key Methods**:
  - `PlaceOrder()` - Places limit orders
  - `GetPositions()` - Retrieves open positions
  - `GetBalance()` - Gets account balances

### 2. Strategy Agent (`internal/strategy/scalping.go`)

The strategy agent implements technical analysis-based trading logic.

**Technical Indicators**:
- **EMA (Exponential Moving Average)**
  - Fast EMA: 9-period
  - Slow EMA: 21-period
- **RSI (Relative Strength Index)**: 14-period
- **Bollinger Bands**: 20-period, 2.0 std deviation

**Signal Generation**:

**Buy Signals** (3 conditions):
1. Fast EMA > Slow EMA (uptrend)
2. RSI < 35 (oversold)
3. Price near lower Bollinger Band

**Sell Signals** (3 conditions):
1. Fast EMA < Slow EMA (downtrend)
2. RSI > 70 (overbought)
3. Price near upper Bollinger Band

**Risk Management**:
- Stop Loss: 0.4% below entry
- Take Profit: 0.8% above entry
- Confidence scoring based on signal strength

**Key Methods**:
- `GenerateSignal()` - Analyzes candles and produces trading signals
- `calculateEMA()` - Computes exponential moving averages
- `calculateRSI()` - Calculates relative strength index
- `calculateBollingerBands()` - Determines volatility bands

### 3. TUI Agent (`internal/tui/`)

The Terminal User Interface agent manages real-time display and user interaction.

#### Model (`model.go`)
- **State Management**:
  - Connected exchanges
  - Active positions across all exchanges
  - Recent trading signals
  - Open orders
  - Account balance
  - Profit/Loss tracking

#### Update Handler (`update.go`)
- **Event Processing**:
  - Keyboard commands (`q`, `ctrl+c`, `r`)
  - Periodic data refresh (1-second interval)
  - Window resize events
  - Asynchronous position fetching

#### View Renderer (`view.go`)
- **Display Components**:
  - Bot header with branding
  - Balance and PnL summary
  - Recent signals with confidence levels
  - Active positions table
  - Color-coded signal indicators (green=buy, red=sell)

## Agent Communication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         TUI Agent                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Model       │→ │  Update      │→ │  View        │          │
│  │  (State)     │  │  (Logic)     │  │  (Render)    │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────┬──────────────────────────────────────┘
                           │ Commands & Queries
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Exchange Interface                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Hyperliquid  │  │  Coinbase    │  │  dYdX (TODO) │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          ↓                  ↓                  ↓
    WebSocket/REST      REST API          REST/WS
          │                  │                  │
          ↓                  ↓                  ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Strategy Agent                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Market Data → Indicators → Signals → Risk Management    │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Data Structures

### Exchange Interface (`internal/exchanges/interface.go`)

```go
type Exchange interface {
    PlaceOrder(ctx context.Context, order Order) (OrderResponse, error)
    GetPositions(ctx context.Context) ([]Position, error)
    GetBalance(ctx context.Context) (Balance, error)
}
```

**Core Types**:
- `Candle` - OHLCV candlestick data
- `OrderBook` - Bid/ask levels
- `Trade` - Individual trade execution
- `Order` - Order request structure
- `Position` - Active trading position
- `Balance` - Account balance info

## Agent Lifecycle

### 1. Initialization (`cmd/bot/main.go`)
```
1. Load configuration (API keys, private keys)
2. Initialize exchange agents (Hyperliquid, Coinbase)
3. Establish connections (WebSocket for Hyperliquid)
4. Launch TUI agent with connected exchanges
```

### 2. Runtime Operation
```
Exchange Agents:
├── Stream market data via WebSocket
├── Process order placement requests
└── Provide balance/position queries

Strategy Agent:
├── Consume candle data from exchanges
├── Calculate technical indicators
├── Generate buy/sell signals
└── Apply risk management rules

TUI Agent:
├── Poll positions every second
├── Display real-time updates
├── Handle user commands
└── Aggregate multi-exchange data
```

### 3. Signal Execution Flow
```
1. Exchange Agent receives new candle data
2. Strategy Agent analyzes candles with indicators
3. Signal generated (BUY/SELL) with confidence score
4. TUI Agent displays signal to user
5. (Future) Execution Agent places order automatically
6. Position tracked across all exchanges
7. TUI updates with new position and PnL
```

## Extension Points

### Adding New Exchange Agents
1. Implement the `Exchange` interface in `internal/exchanges/`
2. Create client with authentication mechanism
3. Implement required methods (PlaceOrder, GetPositions, GetBalance)
4. Register in `cmd/bot/main.go`

### Adding New Strategy Agents
1. Create new file in `internal/strategy/`
2. Define indicator calculations
3. Implement signal generation logic
4. Configure risk management parameters
5. Integrate with TUI for signal display

### 4. Backtesting Agent (`internal/backtesting/`)

The backtesting agent validates trading strategies against historical data.

#### Engine (`engine.go`)
- **State Management**:
  - Current capital and equity tracking
  - Open positions and trade history
  - Equity curve generation
- **Trade Execution**:
  - Signal-based position opening/closing
  - Stop loss and take profit management
  - Commission and slippage simulation
- **Key Methods**:
  - `Run()` - Executes backtest on historical data
  - `handleSignal()` - Processes trading signals
  - `openPosition()` - Opens new positions with risk management
  - `closePosition()` - Closes positions and records trades
  - `calculateMetrics()` - Generates performance statistics

#### Data Loader (`data_loader.go`)
- **Data Sources**:
  - CSV file import (timestamp, OHLCV format)
  - Sample data generation for testing
- **Timestamp Parsing**:
  - Unix timestamps (seconds/milliseconds)
  - RFC3339 and common date formats
- **Key Methods**:
  - `LoadFromCSV()` - Loads historical candles from CSV
  - `GenerateSampleData()` - Creates synthetic test data

#### Simulated Exchange (`simulated_exchange.go`)
- **Purpose**: Mock exchange for backtesting
- **Capabilities**:
  - Historical candle replay
  - Simulated order book
  - Balance and position tracking
- **Implementation**: Implements full `Exchange` interface

#### Reporter (`reporter.go`)
- **Report Types**:
  - Performance summary report
  - Detailed trade log
  - Quick summary metrics
- **Metrics Displayed**:
  - Total return, annualized return
  - Win rate, profit factor
  - Max drawdown
  - Trade statistics

#### CLI Tool (`cmd/backtest/`)
- **Usage**: `./bin/backtest --data=file.csv [options]`
- **Options**:
  - `--data`: Path to CSV file with historical data
  - `--symbol`: Trading symbol (default: BTC-USD)
  - `--capital`: Initial capital (default: $10,000)
  - `--commission`: Commission rate (default: 0.1%)
  - `--slippage`: Slippage rate (default: 0.05%)
  - `--generate-sample`: Generate sample data instead of loading
  - `--verbose`: Show detailed trade log
- **Strategy Parameters**:
  - `--short-ema`, `--long-ema`: EMA periods
  - `--rsi-period`, `--rsi-oversold`, `--rsi-overbought`: RSI settings
  - `--take-profit`, `--stop-loss`: Exit thresholds

### Future Agent Opportunities
- **Risk Manager Agent**: Portfolio-level risk monitoring ✓ (Implemented in internal/risk/)
- **Execution Agent**: Automated order placement based on signals
- **Backtesting Agent**: Historical strategy validation ✓ (Implemented)
- **Alert Agent**: Notification system for critical events
- **Analytics Agent**: Performance metrics and reporting

## Best Practices

1. **Interface Isolation**: Each exchange agent implements common interface
2. **Async Communication**: Use channels for real-time data streams
3. **Error Handling**: Comprehensive error propagation and logging
4. **State Management**: TUI agent centralizes application state
5. **Modularity**: Easy to add/remove exchange agents without breaking system
6. **Testing**: Interface-based design enables easy mocking for tests

## Configuration

Exchange agents require specific credentials:
- **Hyperliquid**: Ethereum private key (hex format)
- **Coinbase**: API Key + Secret
- **dYdX** (planned): API credentials

Store credentials securely using environment variables or encrypted configuration files.

## Monitoring & Observability

The TUI agent provides real-time visibility into:
- Active positions across all exchanges
- Recent trading signals with timestamps
- Account balance and unrealized PnL
- Order execution status
- System health indicators

Future enhancements could include:
- Structured logging (zerolog integration)
- Metrics export (Prometheus)
- Trade history persistence
- Performance analytics dashboard
