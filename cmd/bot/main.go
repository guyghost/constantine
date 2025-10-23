package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/exchanges/coinbase"
	"github.com/guyghost/constantine/internal/exchanges/dydx"
	"github.com/guyghost/constantine/internal/exchanges/hyperliquid"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/guyghost/constantine/internal/tui"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

const (
	defaultAPIKey    = ""
	defaultAPISecret = ""
	defaultBalance   = 10000.0
)

var (
	headless = flag.Bool("headless", false, "Run in headless mode without TUI")
)

// ExchangeConfig holds configuration for a single exchange
type ExchangeConfig struct {
	Name             string
	Enabled          bool
	APIKey           string
	APISecret        string
	WalletAddress    string
	SubAccountNumber int
}

// loadExchangeConfigs loads exchange configurations from environment variables
func loadExchangeConfigs() map[string]*ExchangeConfig {
	configs := make(map[string]*ExchangeConfig)

	// Hyperliquid
	configs["hyperliquid"] = &ExchangeConfig{
		Name:      "hyperliquid",
		Enabled:   getEnvBool("ENABLE_HYPERLIQUID", true),
		APIKey:    os.Getenv("HYPERLIQUID_API_KEY"),
		APISecret: os.Getenv("HYPERLIQUID_API_SECRET"),
	}

	// Coinbase
	configs["coinbase"] = &ExchangeConfig{
		Name:      "coinbase",
		Enabled:   getEnvBool("ENABLE_COINBASE", true),
		APIKey:    os.Getenv("COINBASE_API_KEY"),
		APISecret: os.Getenv("COINBASE_API_SECRET"),
	}

	// dYdX
	configs["dydx"] = &ExchangeConfig{
		Name:             "dydx",
		Enabled:          getEnvBool("ENABLE_DYDX", false),
		APIKey:           os.Getenv("DYDX_API_KEY"),
		APISecret:        os.Getenv("DYDX_MNEMONIC"), // Use mnemonic as APISecret for dYdX
		WalletAddress:    os.Getenv("DYDX_WALLET_ADDRESS"),
		SubAccountNumber: getEnvInt("DYDX_SUBACCOUNT_NUMBER", 0),
	}

	return configs
}

// getEnvBool gets a boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// getEnvInt gets an integer environment variable with default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

func main() {
	// Load .env file if it exists
	godotenv.Load()

	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Initialize components
	aggregator, strategyEngine, orderManager, riskManager, err := initializeBot()
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}

	// Connect to all exchanges
	if err := aggregator.ConnectAll(ctx); err != nil {
		return fmt.Errorf("failed to connect to exchanges: %w", err)
	}
	defer aggregator.DisconnectAll()

	// Setup callbacks
	setupCallbacks(strategyEngine, orderManager, riskManager)

	// Start bot components in background
	go func() {
		if err := startBotComponents(ctx, strategyEngine, orderManager); err != nil {
			log.Printf("Error starting bot components: %v", err)
		}
	}()

	// Run in headless or TUI mode
	if *headless {
		return runHeadless(ctx, aggregator, strategyEngine, orderManager, riskManager)
	}

	// Create TUI model
	model := tui.NewModel(aggregator, strategyEngine, orderManager, riskManager)

	// Start the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// initializeBot initializes all bot components
func initializeBot() (
	*exchanges.MultiExchangeAggregator,
	*strategy.ScalpingStrategy,
	*order.Manager,
	*risk.Manager,
	error,
) {
	// Load exchange configurations
	exchangeConfigs := loadExchangeConfigs()

	// Create all exchange clients based on configuration
	exchangesMap := make(map[string]exchanges.Exchange)

	// Hyperliquid exchange
	if exchangeConfigs["hyperliquid"].Enabled {
		hyperliquidExchange := hyperliquid.NewClient(
			exchangeConfigs["hyperliquid"].APIKey,
			exchangeConfigs["hyperliquid"].APISecret,
		)
		exchangesMap["hyperliquid"] = hyperliquidExchange
		log.Printf("Hyperliquid exchange enabled")
	}

	// Coinbase exchange
	if exchangeConfigs["coinbase"].Enabled {
		portfolioID := os.Getenv("COINBASE_PORTFOLIO_ID")
		var coinbaseExchange *coinbase.Client

		if portfolioID != "" {
			coinbaseExchange = coinbase.NewClientWithPortfolio(
				exchangeConfigs["coinbase"].APIKey,
				exchangeConfigs["coinbase"].APISecret, // Now expects private key PEM
				portfolioID,
			)
		} else {
			coinbaseExchange = coinbase.NewClient(
				exchangeConfigs["coinbase"].APIKey,
				exchangeConfigs["coinbase"].APISecret, // Now expects private key PEM
			)
		}
		exchangesMap["coinbase"] = coinbaseExchange
		log.Printf("Coinbase exchange enabled")
	}

	// dYdX exchange
	if exchangeConfigs["dydx"].Enabled {
		var dydxExchange exchanges.Exchange
		var err error

		// Check if DYDX_MNEMONIC is set (preferred method)
		mnemonic := os.Getenv("DYDX_MNEMONIC")
		if mnemonic != "" {
			// Use mnemonic-based authentication
			dydxExchange, err = dydx.NewClientWithMnemonic(
				mnemonic,
				exchangeConfigs["dydx"].SubAccountNumber,
			)
			if err != nil {
				return nil, nil, nil, nil, fmt.Errorf("failed to create dYdX client with mnemonic: %w", err)
			}
			log.Printf("dYdX exchange enabled (mnemonic authentication)")
		} else if exchangeConfigs["dydx"].APISecret != "" {
			// Use traditional API key authentication
			dydxExchange = dydx.NewClient(
				exchangeConfigs["dydx"].APIKey,
				exchangeConfigs["dydx"].APISecret,
			)
			log.Printf("dYdX exchange enabled (API key authentication)")
		} else {
			return nil, nil, nil, nil, fmt.Errorf("dYdX enabled but no authentication method provided - set DYDX_MNEMONIC or DYDX_API_KEY/DYDX_API_SECRET")
		}

		exchangesMap["dydx"] = dydxExchange
	}

	if len(exchangesMap) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("no exchanges enabled - check ENABLE_* environment variables")
	}

	// Create aggregator
	aggregator := exchanges.NewMultiExchangeAggregator(exchangesMap)

	// Create strategy configuration
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = os.Getenv("TRADING_SYMBOL")
	if strategyConfig.Symbol == "" {
		strategyConfig.Symbol = "BTC-USD"
	}

	// Use the first enabled exchange as primary for strategy and order manager
	var primaryExchange exchanges.Exchange
	for _, exchange := range exchangesMap {
		primaryExchange = exchange
		break
	}

	// Create strategy
	strategyEngine := strategy.NewScalpingStrategy(strategyConfig, primaryExchange)

	// Create order manager
	orderManager := order.NewManager(primaryExchange)

	// Create risk manager
	riskConfig := risk.DefaultConfig()
	initialBalanceStr := os.Getenv("INITIAL_BALANCE")
	initialBalance := decimal.NewFromFloat(defaultBalance)
	if initialBalanceStr != "" {
		if parsed, err := decimal.NewFromString(initialBalanceStr); err == nil {
			initialBalance = parsed
		}
	}
	riskManager := risk.NewManager(riskConfig, initialBalance)

	return aggregator, strategyEngine, orderManager, riskManager, nil
}

// setupCallbacks sets up callbacks between components
func setupCallbacks(
	strategyEngine *strategy.ScalpingStrategy,
	orderManager *order.Manager,
	riskManager *risk.Manager,
) {
	// Strategy signal callback
	strategyEngine.SetSignalCallback(func(signal *strategy.Signal) {
		log.Printf("Signal: %s %s for %s at $%s (strength: %.2f)",
			signal.Type, signal.Side, signal.Symbol,
			signal.Price.StringFixed(2), signal.Strength)

		// Check if we can trade
		canTrade, reason := riskManager.CanTrade()
		if !canTrade {
			log.Printf("Trading blocked: %s", reason)
			return
		}

		// Handle entry signals
		if signal.Type == strategy.SignalTypeEntry && signal.Strength > 0.5 {
			handleEntrySignal(signal, orderManager, riskManager)
		}

		// Handle exit signals
		if signal.Type == strategy.SignalTypeExit {
			handleExitSignal(signal, orderManager)
		}
	})

	// Strategy error callback
	strategyEngine.SetErrorCallback(func(err error) {
		log.Printf("Strategy error: %v", err)
	})

	// Order manager callbacks
	orderManager.SetOrderUpdateCallback(func(update *order.OrderUpdate) {
		log.Printf("Order update: %s - %s", update.Order.ID, update.Event)

		// Record trade in risk manager if filled
		if update.Event == order.OrderEventFilled {
			// TODO: Calculate PnL and record trade
		}
	})

	orderManager.SetPositionUpdateCallback(func(position *order.ManagedPosition) {
		log.Printf("Position update: %s %s - PnL: $%s",
			position.Symbol, position.Side,
			position.UnrealizedPnL.StringFixed(2))
	})

	orderManager.SetErrorCallback(func(err error) {
		log.Printf("Order manager error: %v", err)
	})
}

// handleEntrySignal handles entry signals from the strategy
func handleEntrySignal(
	signal *strategy.Signal,
	orderManager *order.Manager,
	riskManager *risk.Manager,
) {
	// Calculate position size
	stopLoss := signal.Price.Mul(decimal.NewFromFloat(0.995)) // 0.5% stop loss
	if signal.Side == exchanges.OrderSideSell {
		stopLoss = signal.Price.Mul(decimal.NewFromFloat(1.005))
	}

	balance := riskManager.GetCurrentBalance()
	positionSize := riskManager.CalculatePositionSize(signal.Price, stopLoss, balance)

	// Create order request
	req := &order.OrderRequest{
		Symbol:     signal.Symbol,
		Side:       signal.Side,
		Type:       exchanges.OrderTypeLimit,
		Price:      signal.Price,
		Amount:     positionSize,
		StopLoss:   stopLoss,
		TakeProfit: signal.Price.Mul(decimal.NewFromFloat(1.01)), // 1% take profit
	}

	// Validate order
	positions := orderManager.GetPositions()
	if err := riskManager.ValidateOrder(req, positions); err != nil {
		log.Printf("Order validation failed: %v", err)
		return
	}

	// Place order
	ctx := context.Background()
	placedOrder, err := orderManager.PlaceOrder(ctx, req)
	if err != nil {
		log.Printf("Failed to place order: %v", err)
		return
	}

	log.Printf("Order placed: %s - %s %s @ $%s",
		placedOrder.ID, placedOrder.Side, placedOrder.Symbol,
		placedOrder.Price.StringFixed(2))
}

// handleExitSignal handles exit signals from the strategy
func handleExitSignal(
	signal *strategy.Signal,
	orderManager *order.Manager,
) {
	// Close position for the symbol
	ctx := context.Background()
	if err := orderManager.ClosePosition(ctx, signal.Symbol); err != nil {
		log.Printf("Failed to close position: %v", err)
		return
	}

	log.Printf("Position closed: %s", signal.Symbol)
}

// startBotComponents starts the bot components
func startBotComponents(
	ctx context.Context,
	strategyEngine *strategy.ScalpingStrategy,
	orderManager *order.Manager,
) error {
	// Start order manager
	if err := orderManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start order manager: %w", err)
	}

	// Start strategy
	if err := strategyEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start strategy: %w", err)
	}

	log.Println("Bot components started successfully")

	// Wait for context cancellation
	<-ctx.Done()

	// Stop components
	strategyEngine.Stop()
	orderManager.Stop()

	log.Println("Bot components stopped")

	return nil
}

// runHeadless runs the bot in headless mode with periodic status updates
func runHeadless(
	ctx context.Context,
	aggregator *exchanges.MultiExchangeAggregator,
	strategyEngine *strategy.ScalpingStrategy,
	orderManager *order.Manager,
	riskManager *risk.Manager,
) error {
	log.Println("=== Constantine Trading Bot - Headless Mode ===")
	log.Printf("Multi-Exchange Mode: Connected to Hyperliquid, Coinbase, dYdX")
	log.Printf("Strategy: Scalping with EMA/RSI/Bollinger Bands")
	log.Println("Press Ctrl+C to stop")
	log.Println("===============================================")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down headless mode...")
			return nil

		case <-ticker.C:
			// Refresh data from all exchanges
			if err := aggregator.RefreshData(ctx); err != nil {
				log.Printf("[STATUS] Failed to refresh data: %v", err)
			}

			// Log periodic status updates
			logAggregatedStatus(aggregator, orderManager, riskManager)
		}
	}
}

// logAggregatedStatus logs the current aggregated status of all exchanges
func logAggregatedStatus(
	aggregator *exchanges.MultiExchangeAggregator,
	orderManager *order.Manager,
	riskManager *risk.Manager,
) {
	data := aggregator.GetAggregatedData()

	log.Printf("[STATUS] Total Balance: $%s", data.TotalBalance.StringFixed(2))
	log.Printf("[STATUS] Total PnL: $%s", data.TotalPnL.StringFixed(2))

	// Log each exchange status
	for name, exchangeData := range data.Exchanges {
		status := "✓"
		if !exchangeData.Connected {
			status = "✗"
		}

		log.Printf("[STATUS] %s %s: Connected=%v", status, name, exchangeData.Connected)

		if exchangeData.Error != nil {
			log.Printf("  - Error: %v", exchangeData.Error)
		}

		// Log balances
		for _, balance := range exchangeData.Balances {
			if balance.Total.GreaterThan(decimal.Zero) {
				log.Printf("  - Balance: %s $%s", balance.Asset, balance.Total.StringFixed(2))
			}
		}

		// Log positions
		for _, position := range exchangeData.Positions {
			log.Printf("  - Position: %s %s @ $%s, PnL: $%s",
				position.Symbol, position.Side,
				position.EntryPrice.StringFixed(2),
				position.UnrealizedPnL.StringFixed(2))
		}
	}

	// Get positions from order manager (primary exchange)
	positions := orderManager.GetPositions()
	log.Printf("[STATUS] Active Positions: %d", len(positions))

	// Get pending orders
	orders := orderManager.GetOpenOrders()
	log.Printf("[STATUS] Pending Orders: %d", len(orders))

	// Risk stats
	currentBalance := riskManager.GetCurrentBalance()
	canTrade, reason := riskManager.CanTrade()
	log.Printf("[STATUS] Current Balance: $%s", currentBalance.StringFixed(2))
	log.Printf("[STATUS] Can Trade: %v", canTrade)
	if !canTrade {
		log.Printf("[STATUS] Trading blocked: %s", reason)
	}

	log.Println("-----------------------------------------------")
}
