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
	"github.com/guyghost/constantine/internal/exchanges/hyperliquid"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/guyghost/constantine/internal/tui"
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
	exchange, strategyEngine, orderManager, riskManager, err := initializeBot(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}

	// Connect to exchange
	if err := exchange.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to exchange: %w", err)
	}
	defer exchange.Disconnect()

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
		return runHeadless(ctx, exchange, strategyEngine, orderManager, riskManager)
	}

	// Create TUI model
	model := tui.NewModel(exchange, strategyEngine, orderManager, riskManager)

	// Start the TUI
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// initializeBot initializes all bot components
func initializeBot(ctx context.Context) (
	exchanges.Exchange,
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

	// Choose exchange based on environment variable
	exchangeType := os.Getenv("EXCHANGE")
	if exchangeType == "" {
		exchangeType = "hyperliquid"
	}

	var exchange exchanges.Exchange
	switch exchangeType {
	case "hyperliquid":
		exchange = hyperliquid.NewClient(apiKey, apiSecret)
	// case "coinbase":
	// 	exchange = coinbase.NewClient(apiKey, apiSecret)
	// case "dydx":
	// 	exchange = dydx.NewClient(apiKey, apiSecret)
	default:
		exchange = hyperliquid.NewClient(apiKey, apiSecret)
	}

	// Create strategy configuration
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = "BTC-USD"

	// Create strategy
	strategyEngine := strategy.NewScalpingStrategy(strategyConfig, exchange)

	// Create order manager
	orderManager := order.NewManager(exchange)

	// Create risk manager
	riskConfig := risk.DefaultConfig()
	initialBalance := decimal.NewFromFloat(defaultBalance)
	riskManager := risk.NewManager(riskConfig, initialBalance)

	return exchange, strategyEngine, orderManager, riskManager, nil
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
	exchange exchanges.Exchange,
	strategyEngine *strategy.ScalpingStrategy,
	orderManager *order.Manager,
	riskManager *risk.Manager,
) error {
	log.Println("=== Constantine Trading Bot - Headless Mode ===")
	log.Printf("Exchange: Hyperliquid (Demo Mode)")
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
			// Log periodic status updates
			logStatus(exchange, orderManager, riskManager)
		}
	}
}

// logStatus logs the current status of the bot
func logStatus(
	exchange exchanges.Exchange,
	orderManager *order.Manager,
	riskManager *risk.Manager,
) {
	ctx := context.Background()

	// Get balance
	balances, err := exchange.GetBalance(ctx)
	if err != nil {
		log.Printf("[STATUS] Failed to get balance: %v", err)
	} else if len(balances) > 0 {
		log.Printf("[STATUS] Balance: %s $%s", balances[0].Asset, balances[0].Total.StringFixed(2))
	}

	// Get positions
	positions := orderManager.GetPositions()
	log.Printf("[STATUS] Active Positions: %d", len(positions))

	for _, pos := range positions {
		log.Printf("  - %s %s: Entry=$%s, Current PnL=$%s",
			pos.Symbol, pos.Side,
			pos.EntryPrice.StringFixed(2),
			pos.UnrealizedPnL.StringFixed(2))
	}

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
