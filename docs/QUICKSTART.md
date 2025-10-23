# Quick Start Guide

Get your scalping bot up and running in minutes!

## Prerequisites

- Go 1.22+ installed
- API credentials from your chosen exchange
- Terminal with at least 80x24 character dimensions

## Installation Steps

### 1. Navigate to the project

```bash
cd docs
```

### 2. Install dependencies

```bash
make install-deps
```

or

```bash
go mod download
go mod tidy
```

### 3. Configure environment

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and add your credentials:

```bash
EXCHANGE=hyperliquid
EXCHANGE_API_KEY=your_actual_api_key
EXCHANGE_API_SECRET=your_actual_api_secret
```

### 4. Build the bot

```bash
make build
```

or

```bash
go build -o docs cmd/bot/main.go
```

### 5. Run the bot

```bash
./docs
```

or

```bash
make run
```

## First Run

When you start the bot for the first time:

1. **Dashboard View** will appear showing:
   - Account balance (default: $10,000)
   - Current trading signal
   - Risk management status
   - Recent activity log

2. **Bot Status**: The bot starts in STOPPED mode
   - Press `s` to start trading
   - Press `s` again to stop

3. **Navigate Views**:
   - Press `1` for Dashboard
   - Press `2` for Order Book
   - Press `3` for Positions
   - Press `4` for Orders
   - Press `5` for Settings

## Testing Without Real Money

### Demo Mode

The bot includes mock implementations for testing:

1. Leave API credentials empty or use dummy values
2. The bot will use simulated data
3. All orders will be mock orders (no real trades)

### Paper Trading

To test with real market data but no actual trades:

1. Use read-only API keys (if your exchange supports them)
2. Monitor signals without executing trades
3. Track performance metrics

## Configuration Tips

### Conservative Settings (Recommended for Beginners)

```env
# Small positions
MAX_POSITION_SIZE=100

# Fewer concurrent trades
MAX_POSITIONS=1

# Tight risk controls
MAX_DAILY_LOSS=50
STOP_LOSS_PERCENT=0.5
TAKE_PROFIT_PERCENT=1.0

# Lower trade frequency
DAILY_TRADING_LIMIT=10
```

### Aggressive Settings (For Experienced Traders)

```env
# Larger positions
MAX_POSITION_SIZE=5000

# More concurrent trades
MAX_POSITIONS=5

# Wider risk tolerance
MAX_DAILY_LOSS=500
STOP_LOSS_PERCENT=0.25
TAKE_PROFIT_PERCENT=0.5

# Higher trade frequency
DAILY_TRADING_LIMIT=100
```

## Monitoring Your Bot

### Key Metrics to Watch

1. **Daily P&L**: Should stay within your risk limits
2. **Win Rate**: Target 50%+ for profitable scalping
3. **Drawdown**: Keep below 10% for safety
4. **Consecutive Losses**: Bot auto-pauses after 3 losses

### Warning Signs

Stop the bot if you see:
- Rapidly increasing losses
- Drawdown > 15%
- Consistent failed orders
- Unusual market conditions

## Common Issues

### Bot won't start

**Problem**: Connection errors
**Solution**: Check API credentials and network connection

**Problem**: Build errors
**Solution**: Run `go mod tidy` and rebuild

### No trades executing

**Problem**: Risk manager blocking trades
**Solution**: Check risk limits in configuration

**Problem**: No signals generated
**Solution**: Verify market data connection

### Orders failing

**Problem**: Insufficient balance
**Solution**: Reduce position size or add funds

**Problem**: Invalid API permissions
**Solution**: Ensure API key has trading permissions

## Next Steps

1. **Read the full README.md** for detailed documentation
2. **Customize the strategy** in `internal/strategy/`
3. **Adjust risk parameters** in `internal/risk/`
4. **Monitor and optimize** based on performance

## Safety Reminders

⚠️ **Important Guidelines**:

- Start with minimum position sizes
- Never risk more than 1-2% per trade
- Set strict daily loss limits
- Monitor the bot regularly
- Keep API keys secure
- Use strong passwords
- Enable 2FA on exchange accounts
- Test thoroughly before live trading

## Getting Help

- Check logs for error messages
- Review the code comments
- Open an issue on GitHub
- Join the community discussions

## Quick Commands

```bash
# Build and run
make run

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean

# View available commands
make help
```

## Development Mode

For development and testing:

```bash
# Run without building
make dev

# Watch for changes (requires external tool)
air  # or use your preferred hot-reload tool
```

## Monitoring Commands

While the bot is running:

- `r` - Refresh data manually
- `c` - Clear error messages
- `1-5` - Switch between different views
- `q` - Quit the application

## Performance Tips

1. **Stable Internet**: Use wired connection for reliability
2. **Low Latency**: Run bot close to exchange servers
3. **Resource Monitoring**: Ensure adequate CPU/RAM
4. **Regular Updates**: Keep dependencies up to date

## Scaling Up

Once comfortable with the bot:

1. Gradually increase position sizes
2. Add more concurrent positions
3. Experiment with different symbols
4. Try multiple exchange connections
5. Optimize strategy parameters

---

**Remember**: Start small, test thoroughly, and scale gradually. Never invest more than you can afford to lose!

Happy trading! 🚀
