package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/constantine/internal/config"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/exchanges/coinbase"
	"github.com/guyghost/constantine/internal/exchanges/dydx"
	"github.com/guyghost/constantine/internal/exchanges/hyperliquid"
	"github.com/guyghost/constantine/internal/execution"
	"github.com/guyghost/constantine/internal/logger"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/guyghost/constantine/internal/tui"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

var (
	headless = flag.Bool("headless", false, "Run in headless mode without TUI")
)

// getEnvBool gets a boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

func loadLoggerConfig() *logger.Config {
	cfg := logger.DefaultConfig()

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		cfg.Level = slog.LevelDebug
	case "warn", "warning":
		cfg.Level = slog.LevelWarn
	case "error":
		cfg.Level = slog.LevelError
	case "info", "":
		cfg.Level = slog.LevelInfo
	}

	format := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if format == "text" {
		cfg.Format = "text"
	} else if format == "json" {
		cfg.Format = "json"
	}

	cfg.AddSource = getEnvBool("LOG_ADD_SOURCE", false)
	if output := os.Getenv("LOG_OUTPUT_PATH"); output != "" {
		cfg.OutputPath = output
	}

	return cfg
}

func main() {
	// Load .env file if it exists
	godotenv.Load()

	flag.Parse()

	logger.SetDefault(logger.New(loadLoggerConfig()))

	if err := run(); err != nil {
		logger.Default().Error("bot exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	appConfig, err := config.Load()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	metricsServer := telemetry.NewServer(appConfig.TelemetryAddr)
	if metricsServer != nil {
		if err := metricsServer.Start(); err != nil {
			return fmt.Errorf("failed to start telemetry server: %w", err)
		}
		defer func() {
			metricsServer.SetReady(false)
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			_ = metricsServer.Shutdown(shutdownCtx)
		}()
	}

	defer func() {
		cancel()
		wg.Wait()
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Initialize components
	aggregator, strategyEngine, orderManager, riskManager, executionAgent, err := initializeBot(appConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}

	// Connect to all exchanges
	if err := aggregator.ConnectAll(ctx); err != nil {
		return fmt.Errorf("failed to connect to exchanges: %w", err)
	}
	defer aggregator.DisconnectAll()

	// Setup callbacks
	setupCallbacks(strategyEngine, orderManager, riskManager, executionAgent)

	// Start bot components in background
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startBotComponents(ctx, strategyEngine, orderManager); err != nil {
			botLogger().Error("failed to start bot components", "error", err)
		}
	}()

	if metricsServer != nil {
		metricsServer.SetReady(true)
	}

	// Run in headless or TUI mode
	if *headless {
		return runHeadless(ctx, aggregator, strategyEngine, orderManager, riskManager, executionAgent)
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

func botLogger() *logger.Logger {
	return logger.Default().Component("bot")
}

// initializeBot initializes all bot components
func initializeBot(appConfig *config.AppConfig) (
	*exchanges.MultiExchangeAggregator,
	*strategy.ScalpingStrategy,
	*order.Manager,
	*risk.Manager,
	*execution.ExecutionAgent,
	error,
) {
	// Create all exchange clients based on configuration
	exchangesMap := make(map[string]exchanges.Exchange)

	// Hyperliquid exchange
	if hyperCfg, ok := appConfig.Exchanges["hyperliquid"]; ok && hyperCfg.Enabled {
		hyperliquidExchange := hyperliquid.NewClient(
			hyperCfg.APIKey,
			hyperCfg.APISecret,
		)
		exchangesMap["hyperliquid"] = hyperliquidExchange
		botLogger().Info("exchange enabled", "exchange", "hyperliquid")
	}

	// Coinbase exchange
	if coinbaseCfg, ok := appConfig.Exchanges["coinbase"]; ok && coinbaseCfg.Enabled {
		var coinbaseExchange *coinbase.Client

		if coinbaseCfg.PortfolioID != "" {
			coinbaseExchange = coinbase.NewClientWithPortfolio(
				coinbaseCfg.APIKey,
				coinbaseCfg.APISecret, // Now expects private key PEM
				coinbaseCfg.PortfolioID,
			)
		} else {
			coinbaseExchange = coinbase.NewClient(
				coinbaseCfg.APIKey,
				coinbaseCfg.APISecret, // Now expects private key PEM
			)
		}
		exchangesMap["coinbase"] = coinbaseExchange
		botLogger().Info("exchange enabled", "exchange", "coinbase")
	}

	// dYdX exchange
	if dydxCfg, ok := appConfig.Exchanges["dydx"]; ok && dydxCfg.Enabled {
		var dydxExchange exchanges.Exchange
		var err error

		// Check if mnemonic is set (preferred method)
		if dydxCfg.Mnemonic != "" {
			// Use mnemonic-based authentication
			dydxExchange, err = dydx.NewClientWithMnemonic(
				dydxCfg.Mnemonic,
				dydxCfg.SubAccountNumber,
			)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to create dYdX client with mnemonic: %w", err)
			}
			botLogger().Info("exchange enabled", "exchange", "dydx", "auth", "mnemonic")
		} else if dydxCfg.APISecret != "" {
			// Use traditional API key authentication
			client, err := dydx.NewClient(
				dydxCfg.APIKey,
				dydxCfg.APISecret,
			)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("failed to create dYdX client: %w", err)
			}
			dydxExchange = client
			botLogger().Info("exchange enabled", "exchange", "dydx", "auth", "api_key")
		} else {
			return nil, nil, nil, nil, nil, fmt.Errorf("dYdX enabled but no authentication method provided - set DYDX_MNEMONIC or DYDX_API_KEY/DYDX_API_SECRET")
		}

		exchangesMap["dydx"] = dydxExchange
	}

	if len(exchangesMap) == 0 {
		return nil, nil, nil, nil, nil, fmt.Errorf("no exchanges enabled - check ENABLE_* environment variables")
	}

	// Create aggregator
	aggregator := exchanges.NewMultiExchangeAggregator(exchangesMap)

	// Create strategy configuration
	strategyConfig := strategy.DefaultConfig()
	strategyConfig.Symbol = appConfig.StrategySymbol

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
	riskConfig := risk.LoadConfig()
	riskManager := risk.NewManager(riskConfig, appConfig.InitialBalance)

	// Create execution agent
	executionConfig := execution.DefaultConfig()
	executionAgent := execution.NewExecutionAgent(orderManager, riskManager, executionConfig)

	return aggregator, strategyEngine, orderManager, riskManager, executionAgent, nil
}

// setupCallbacks sets up callbacks between components
func setupCallbacks(
	strategyEngine *strategy.ScalpingStrategy,
	orderManager *order.Manager,
	riskManager *risk.Manager,
	executionAgent *execution.ExecutionAgent,
) {
	log := botLogger()

	// Strategy signal callback
	strategyEngine.SetSignalCallback(func(signal *strategy.Signal) {
		logSensitive := getEnvBool("LOG_SENSITIVE_DATA", false)

		if logSensitive {
			log.Info("strategy signal",
				"type", signal.Type,
				"side", signal.Side,
				"symbol", signal.Symbol,
				"price", signal.Price.StringFixed(2),
				"strength", signal.Strength,
			)
		} else {
			log.Info("strategy signal",
				"type", signal.Type,
				"side", signal.Side,
				"symbol", signal.Symbol,
				"price", "[REDACTED]",
				"strength", signal.Strength,
			)
		}

		// Handle signal with execution agent
		ctx := context.Background()
		if err := executionAgent.HandleSignal(ctx, signal); err != nil {
			log.Error("execution error", "error", err)
		}
	})

	// Strategy error callback
	strategyEngine.SetErrorCallback(func(err error) {
		log.Error("strategy error", "error", err)
	})

	// Order manager callbacks
	orderManager.SetPositionUpdateCallback(func(position *order.ManagedPosition) {
		logSensitive := getEnvBool("LOG_SENSITIVE_DATA", false)

		if logSensitive {
			log.Info("position update",
				"symbol", position.Symbol,
				"side", position.Side,
				"unrealized_pnl", position.UnrealizedPnL.StringFixed(2),
				"realized_pnl", position.RealizedPnL.StringFixed(2),
			)
		} else {
			log.Info("position update",
				"symbol", position.Symbol,
				"side", position.Side,
				"unrealized_pnl", "[REDACTED]",
				"realized_pnl", "[REDACTED]",
			)
		}
	})

	orderManager.SetPositionUpdateCallback(func(position *order.ManagedPosition) {
		log.Info("position update",
			"symbol", position.Symbol,
			"side", position.Side,
			"unrealized_pnl", position.UnrealizedPnL.StringFixed(2),
			"realized_pnl", position.RealizedPnL.StringFixed(2),
		)
	})

	orderManager.SetErrorCallback(func(err error) {
		log.Error("order manager error", "error", err)
	})
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

	botLogger().Info("bot components started")

	// Wait for context cancellation
	<-ctx.Done()

	// Stop components
	strategyEngine.Stop()
	orderManager.Stop()

	botLogger().Info("bot components stopped")

	return nil
}

// runHeadless runs the bot in headless mode with periodic status updates
func runHeadless(
	ctx context.Context,
	aggregator *exchanges.MultiExchangeAggregator,
	strategyEngine *strategy.ScalpingStrategy,
	orderManager *order.Manager,
	riskManager *risk.Manager,
	executionAgent *execution.ExecutionAgent,
) error {
	log := botLogger()
	log.Info("headless mode initialized",
		"exchanges", []string{"hyperliquid", "coinbase", "dydx"},
		"strategy", "scalping",
	)
	log.Info("headless mode awaiting shutdown signal")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("shutting down headless mode")
			return nil

		case <-ticker.C:
			// Refresh data from all exchanges
			if err := aggregator.RefreshData(ctx); err != nil {
				log.Error("headless refresh failed", "error", err)
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
	log := botLogger()

	// Check if sensitive data logging is enabled
	logSensitive := getEnvBool("LOG_SENSITIVE_DATA", false)

	// Update risk manager with current total balance from exchanges
	riskManager.UpdateBalance(data.TotalBalance)

	if logSensitive {
		log.Info("portfolio status",
			"total_balance", data.TotalBalance.StringFixed(2),
			"total_pnl", data.TotalPnL.StringFixed(2),
		)
	} else {
		log.Info("portfolio status",
			"total_balance", "[REDACTED]",
			"total_pnl", "[REDACTED]",
		)
	}

	// Log each exchange status
	for name, exchangeData := range data.Exchanges {
		entry := log.Component("exchange").WithField("exchange", name)
		entry.Info("exchange status", "connected", exchangeData.Connected)

		if exchangeData.Error != nil {
			entry.Warn("exchange error", "error", exchangeData.Error)
		}

		// Log balances (only if sensitive logging is enabled)
		if logSensitive {
			for _, balance := range exchangeData.Balances {
				if balance.Total.GreaterThan(decimal.Zero) {
					entry.Info("balance snapshot",
						"asset", balance.Asset,
						"total", balance.Total.StringFixed(2),
					)
				}
			}
		}

		// Log positions (only if sensitive logging is enabled)
		if logSensitive {
			for _, position := range exchangeData.Positions {
				entry.Info("position snapshot",
					"symbol", position.Symbol,
					"side", position.Side,
					"entry_price", position.EntryPrice.StringFixed(2),
					"unrealized_pnl", position.UnrealizedPnL.StringFixed(2),
				)
			}
		}
	}

	// Get positions from order manager (primary exchange)
	positions := orderManager.GetPositions()

	// Get pending orders
	orders := orderManager.GetOpenOrders()

	// Risk stats
	currentBalance := riskManager.GetCurrentBalance()
	canTrade, reason := riskManager.CanTrade()
	fields := []any{
		"active_positions", len(positions),
		"pending_orders", len(orders),
	}

	if logSensitive {
		fields = append(fields, "current_balance", currentBalance.StringFixed(2))
	} else {
		fields = append(fields, "current_balance", "[REDACTED]")
	}

	fields = append(fields, "can_trade", canTrade)
	if !canTrade {
		fields = append(fields, "blocked_reason", reason)
	}
	log.Info("risk status", fields...)
}

// calculateAndRecordPnL calculates PnL for filled orders and records trades
func calculateAndRecordPnL(update *order.OrderUpdate, orderManager *order.Manager, riskManager *risk.Manager) {
	filledOrder := update.Order

	// Get current positions to determine if this closes a position
	positions := orderManager.GetPositions()

	// Find if this order closes an existing position
	for _, pos := range positions {
		if pos.Symbol == filledOrder.Symbol {
			// Check if order side closes the position
			orderClosesPosition := false
			if (filledOrder.Side == exchanges.OrderSideSell && pos.Side == order.PositionSideLong) ||
				(filledOrder.Side == exchanges.OrderSideBuy && pos.Side == order.PositionSideShort) {
				orderClosesPosition = true
			}

			if orderClosesPosition {
				// Calculate PnL for closed position
				var pnl decimal.Decimal
				if pos.Side == order.PositionSideLong {
					// Long position closed by sell order
					pnl = filledOrder.Price.Sub(pos.EntryPrice).Mul(pos.Amount)
				} else {
					// Short position closed by buy order
					pnl = pos.EntryPrice.Sub(filledOrder.Price).Mul(pos.Amount)
				}

				// Record the trade
				tradeResult := risk.TradeResult{
					Timestamp:  update.Timestamp,
					Symbol:     filledOrder.Symbol,
					Side:       filledOrder.Side,
					EntryPrice: pos.EntryPrice,
					ExitPrice:  filledOrder.Price,
					Amount:     pos.Amount,
					PnL:        pnl,
					IsWin:      pnl.GreaterThan(decimal.Zero),
				}

				riskManager.RecordTrade(tradeResult)

				logSensitive := getEnvBool("LOG_SENSITIVE_DATA", false)
				if logSensitive {
					botLogger().Info("trade recorded",
						"symbol", filledOrder.Symbol,
						"side", filledOrder.Side,
						"entry_price", pos.EntryPrice.StringFixed(2),
						"exit_price", filledOrder.Price.StringFixed(2),
						"amount", pos.Amount.StringFixed(4),
						"pnl", pnl.StringFixed(2),
						"is_win", tradeResult.IsWin,
					)
				} else {
					botLogger().Info("trade recorded",
						"symbol", filledOrder.Symbol,
						"side", filledOrder.Side,
						"entry_price", "[REDACTED]",
						"exit_price", "[REDACTED]",
						"amount", pos.Amount.StringFixed(4),
						"pnl", "[REDACTED]",
						"is_win", tradeResult.IsWin,
					)
				}
				break
			}
		}
	}
}
