# Scalping Bot

A high-frequency trading bot with support for multiple cryptocurrency exchanges, featuring a beautiful terminal UI built with Bubble Tea.

## Features

- **Multi-Exchange Support**: Hyperliquid, Coinbase, and dYdX
- **Advanced Trading Strategy**: Scalping strategy with technical indicators (EMA, RSI, MACD, Bollinger Bands, etc.)
- **Risk Management**: Comprehensive risk controls including position limits, drawdown protection, and cooldown periods
- **Real-time Market Data**: WebSocket connections for live price feeds and order book updates
- **Order Management**: Automated order placement, tracking, and position management
- **Beautiful TUI**: Interactive terminal interface with real-time updates using Bubble Tea
- **Performance Monitoring**: Track P&L, win rate, and other key metrics

## Architecture

```
scalping-bot/
├── cmd/bot/main.go              # Application entry point
├── internal/
│   ├── exchanges/               # Exchange integrations
│   │   ├── interface.go         # Common exchange interface
│   │   ├── hyperliquid/         # Hyperliquid implementation
│   │   ├── coinbase/            # Coinbase implementation
│   │   └── dydx/                # dYdX implementation
│   ├── strategy/                # Trading strategy
│   │   ├── scalping.go          # Main scalping strategy
│   │   ├── indicators.go        # Technical indicators
│   │   └── signals.go           # Signal generation
│   ├── order/                   # Order management
│   │   ├── manager.go           # Order manager
│   │   └── types.go             # Order types
│   ├── risk/                    # Risk management
│   │   └── manager.go           # Risk manager
│   └── tui/                     # Terminal UI
│       ├── model.go             # TUI model
│       ├── update.go            # Update logic
│       ├── view.go              # View rendering
│       └── components/          # UI components
└── pkg/utils/                   # Utility functions
```

## Installation

### Prerequisites

- Go 1.22 or higher
- API keys for your chosen exchange(s)

### Build

```bash
cd docs
go mod download
go build -o docs cmd/bot/main.go
```

## Configuration

Set environment variables:

```bash
# Exchange selection (hyperliquid, coinbase, or dydx)
export EXCHANGE=hyperliquid

# API credentials
export EXCHANGE_API_KEY=your_api_key
export EXCHANGE_API_SECRET=your_api_secret
```

## Usage

Run the bot:

```bash
./docs
```

### Keyboard Shortcuts

- `1-5`: Switch between views (Dashboard, Order Book, Positions, Orders, Settings)
- `s`: Start/Stop the bot
- `r`: Refresh data
- `c`: Clear error messages
- `q` or `Ctrl+C`: Quit

## Trading Strategy

The bot uses a scalping strategy based on:

### Technical Indicators
- **EMA (Exponential Moving Average)**: Fast (9) and Slow (21) periods for trend detection
- **RSI (Relative Strength Index)**: Overbought/oversold conditions (70/30 thresholds)
- **Order Book Analysis**: Bid/ask imbalance detection
- **MACD, Bollinger Bands, ATR**: Additional confirmation signals

### Entry Conditions
- EMA crossover (short crosses above/below long)
- RSI confirmation (oversold for buys, overbought for sells)
- Order book imbalance in favor of the trade direction
- Signal strength > 50%

### Exit Conditions
- Take profit: 0.5% (configurable)
- Stop loss: 0.25% (configurable)
- RSI extreme levels
- Opposite signal generated

## Risk Management

The bot includes comprehensive risk management:

- **Position Sizing**: Calculated based on account balance and risk per trade
- **Maximum Positions**: Limit concurrent open positions
- **Daily Loss Limit**: Stop trading after reaching daily loss threshold
- **Drawdown Protection**: Monitor and limit maximum drawdown
- **Cooldown Period**: Pause trading after consecutive losses
- **Trade Limits**: Maximum trades per day
- **Leverage Control**: Maximum leverage per position

### Default Risk Parameters

```go
MaxPositionSize:      $1000
MaxPositions:         3
MaxLeverage:          5x
MaxDailyLoss:         $100
MaxDrawdown:          10%
RiskPerTrade:         1%
MinAccountBalance:    $100
DailyTradingLimit:    50
CooldownPeriod:       15 minutes
ConsecutiveLossLimit: 3
```

## Views

### Dashboard
- Account balance and daily P&L
- Current trading signal
- Risk management statistics
- Recent activity log

### Order Book
- Real-time bid/ask levels
- Spread information
- Volume visualization

### Positions
- Open positions with entry prices
- Unrealized P&L
- Position details and duration

### Orders
- Open orders
- Order history
- Order statistics

## Technical Indicators

All indicators are implemented in `internal/strategy/indicators.go`:

- **EMA**: Exponential Moving Average
- **SMA**: Simple Moving Average
- **RSI**: Relative Strength Index
- **MACD**: Moving Average Convergence Divergence
- **Bollinger Bands**: Volatility bands
- **ATR**: Average True Range
- **VWAP**: Volume Weighted Average Price
- **Stochastic**: Stochastic Oscillator

## Development

### Adding a New Exchange

1. Create a new directory under `internal/exchanges/`
2. Implement the `exchanges.Exchange` interface
3. Add WebSocket support for real-time data
4. Register in `cmd/bot/main.go`

### Customizing the Strategy

Edit `internal/strategy/scalping.go` to modify:
- Entry/exit conditions
- Indicator parameters
- Signal strength calculation
- Position sizing logic

### Modifying Risk Parameters

Edit `internal/risk/manager.go` to adjust:
- Risk limits
- Position sizing algorithm
- Cooldown logic
- Drawdown calculation

## Safety Notice

⚠️ **IMPORTANT**: This bot is for educational purposes. Cryptocurrency trading carries significant risk:

- Start with small amounts or paper trading
- Thoroughly test the strategy before live trading
- Monitor the bot regularly
- Set appropriate risk limits
- Never invest more than you can afford to lose
- The bot makes automated trading decisions - understand the risks

## Disclaimer

This software is provided "as is" without warranty of any kind. The authors are not responsible for any financial losses incurred through the use of this bot. Cryptocurrency trading is risky, and past performance does not guarantee future results.

## License

MIT License - See LICENSE file for details

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Support

For issues and questions:
- Open an issue on GitHub
- Check existing documentation
- Review the code comments

## Roadmap

- [ ] Backtesting framework
- [ ] Paper trading mode
- [ ] More exchange integrations (Binance, Kraken, etc.)
- [ ] Additional trading strategies
- [ ] Web-based dashboard
- [ ] Machine learning signal enhancement
- [ ] Advanced charting
- [ ] Telegram notifications
- [ ] Database persistence
- [ ] Strategy optimization tools
