package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	// Get API credentials from environment or use defaults
	apiKey := os.Getenv("EXCHANGE_API_KEY")
	if apiKey == "" {
		apiKey = defaultAPIKey
	}

	apiSecret := os.Getenv("EXCHANGE_API_SECRET")
	if apiSecret == "" {
		apiSecret = defaultAPISecret
	}

	// Create all exchange clients
	exchangesMap := make(map[string]exchanges.Exchange)

	// Hyperliquid exchange
	hyperliquidExchange := hyperliquid.NewClient(apiKey, apiSecret)
	exchangesMap["hyperliquid"] = hyperliquidExchange

	// Coinbase exchange
	coinbaseExchange := coinbase.NewClient(apiKey, apiSecret)
	exchangesMap["coinbase"] = coinbaseExchange

	// dYdX exchange
	dydxExchange := dydx.NewClient(apiKey, apiSecret)
	exchangesMap["dydx"] = dydxExchange

	// Create aggregator
	aggregator := exchanges.NewMultiExchangeAggregator(exchangesMap)

	// Create strategy configuration
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = "BTC-USD"

	// Create strategy (using hyperliquid as primary for now)
	strategyEngine := strategy.NewScalpingStrategy(strategyConfig, hyperliquidExchange)

	// Create order manager (using hyperliquid as primary for now)
	orderManager := order.NewManager(hyperliquidExchange)

	// Create risk manager
	riskConfig := risk.DefaultConfig()
	initialBalance := decimal.NewFromFloat(defaultBalance)
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
