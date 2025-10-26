package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Config holds strategy configuration
type Config struct {
	Symbol            string
	ShortEMAPeriod    int
	LongEMAPeriod     int
	RSIPeriod         int
	RSIOversold       float64
	RSIOverbought     float64
	TakeProfitPercent float64
	StopLossPercent   float64
	MaxPositionSize   decimal.Decimal
	MinPriceMove      decimal.Decimal
	UpdateInterval    time.Duration
	// Price sanity checks
	MaxPriceChangePercent float64 // Maximum allowed price change between updates (default: 5%)
	MinPrice              decimal.Decimal
	MaxPrice              decimal.Decimal
}

// ExchangeConfig holds configuration for an exchange
type ExchangeConfig struct {
	Enabled          bool
	APIKey           string
	APISecret        string
	PortfolioID      string // For Coinbase
	Mnemonic         string // For dYdX
	SubAccountNumber int    // For dYdX
}

// AppConfig holds application-wide configuration
type AppConfig struct {
	TelemetryAddr  string
	StrategySymbol string
	TradingSymbols []string // Multi-symbol support
	InitialBalance decimal.Decimal
	Exchanges      map[string]ExchangeConfig
}

// DefaultConfig returns default strategy configuration
func DefaultConfig() *Config {
	cfg := &Config{
		Symbol:                "BTC-USD",
		ShortEMAPeriod:        9,
		LongEMAPeriod:         21,
		RSIPeriod:             14,
		RSIOversold:           30.0,
		RSIOverbought:         70.0,
		TakeProfitPercent:     2.0, // Updated to 2%
		StopLossPercent:       1.0, // Updated to 1%
		MaxPositionSize:       decimal.NewFromFloat(0.1),
		MinPriceMove:          decimal.NewFromFloat(0.01),
		UpdateInterval:        1 * time.Second,
		MaxPriceChangePercent: 5.0,                           // 5% max price change
		MinPrice:              decimal.NewFromFloat(0.01),    // Minimum valid price
		MaxPrice:              decimal.NewFromFloat(1000000), // Maximum valid price
	}

	if symbol := os.Getenv("STRATEGY_SYMBOL"); symbol != "" {
		cfg.Symbol = symbol
	}
	if val := parseIntEnv("STRATEGY_SHORT_EMA", cfg.ShortEMAPeriod); val > 0 {
		cfg.ShortEMAPeriod = val
	}
	if val := parseIntEnv("STRATEGY_LONG_EMA", cfg.LongEMAPeriod); val > 0 {
		cfg.LongEMAPeriod = val
	}
	if val := parseIntEnv("STRATEGY_RSI_PERIOD", cfg.RSIPeriod); val > 0 {
		cfg.RSIPeriod = val
	}
	if val := parseFloatEnv("STRATEGY_RSI_OVERSOLD", cfg.RSIOversold); val > 0 {
		cfg.RSIOversold = val
	}
	if val := parseFloatEnv("STRATEGY_RSI_OVERBOUGHT", cfg.RSIOverbought); val > 0 {
		cfg.RSIOverbought = val
	}
	if val := parseFloatEnv("STRATEGY_TAKE_PROFIT", cfg.TakeProfitPercent); val > 0 {
		cfg.TakeProfitPercent = val
	}
	if val := parseFloatEnv("STRATEGY_STOP_LOSS", cfg.StopLossPercent); val > 0 {
		cfg.StopLossPercent = val
	}
	if value := os.Getenv("STRATEGY_MAX_POSITION_SIZE"); value != "" {
		if parsed, err := decimal.NewFromString(value); err == nil {
			cfg.MaxPositionSize = parsed
		}
	}
	if value := os.Getenv("STRATEGY_MIN_PRICE_MOVE"); value != "" {
		if parsed, err := decimal.NewFromString(value); err == nil {
			cfg.MinPriceMove = parsed
		}
	}
	if duration := os.Getenv("STRATEGY_UPDATE_INTERVAL"); duration != "" {
		if parsed, err := time.ParseDuration(duration); err == nil {
			cfg.UpdateInterval = parsed
		}
	}
	if val := parseFloatEnv("STRATEGY_MAX_PRICE_CHANGE_PERCENT", cfg.MaxPriceChangePercent); val > 0 {
		cfg.MaxPriceChangePercent = val
	}
	if value := os.Getenv("STRATEGY_MIN_PRICE"); value != "" {
		if parsed, err := decimal.NewFromString(value); err == nil && parsed.GreaterThan(decimal.Zero) {
			cfg.MinPrice = parsed
		}
	}
	if value := os.Getenv("STRATEGY_MAX_PRICE"); value != "" {
		if parsed, err := decimal.NewFromString(value); err == nil && parsed.GreaterThan(decimal.Zero) {
			cfg.MaxPrice = parsed
		}
	}

	return cfg
}

// Load loads application configuration from environment variables
func Load() (*AppConfig, error) {
	cfg := &AppConfig{
		TelemetryAddr:  ":9090", // Default telemetry address
		StrategySymbol: "BTC-USD",
		TradingSymbols: []string{"BTC-USD"},         // Default single symbol
		InitialBalance: decimal.NewFromFloat(10000), // Default $10,000
		Exchanges:      make(map[string]ExchangeConfig),
	}

	// Load telemetry address
	if addr := os.Getenv("TELEMETRY_ADDR"); addr != "" {
		cfg.TelemetryAddr = addr
	}

	// Load strategy symbol (single symbol, for backward compatibility)
	if symbol := os.Getenv("STRATEGY_SYMBOL"); symbol != "" {
		cfg.StrategySymbol = symbol
		if len(cfg.TradingSymbols) == 1 && cfg.TradingSymbols[0] == "BTC-USD" {
			cfg.TradingSymbols = []string{symbol}
		}
	}

	// Load trading symbols (multi-symbol support)
	if symbols := os.Getenv("TRADING_SYMBOLS"); symbols != "" {
		// Parse comma-separated list
		symbolList := strings.Split(strings.TrimSpace(symbols), ",")
		var validSymbols []string
		for _, s := range symbolList {
			s = strings.TrimSpace(s)
			if s != "" {
				validSymbols = append(validSymbols, s)
			}
		}
		if len(validSymbols) > 0 {
			cfg.TradingSymbols = validSymbols
			// Set primary symbol to first in list
			cfg.StrategySymbol = validSymbols[0]
		}
	}

	// Load initial balance
	if balance := os.Getenv("INITIAL_BALANCE"); balance != "" {
		if parsed, err := decimal.NewFromString(balance); err == nil {
			cfg.InitialBalance = parsed
		}
	}

	// Load exchange configurations
	cfg.Exchanges["hyperliquid"] = ExchangeConfig{
		Enabled:   os.Getenv("ENABLE_HYPERLIQUID") == "true",
		APIKey:    os.Getenv("HYPERLIQUID_API_KEY"),
		APISecret: os.Getenv("HYPERLIQUID_API_SECRET"),
	}

	cfg.Exchanges["coinbase"] = ExchangeConfig{
		Enabled:     os.Getenv("ENABLE_COINBASE") == "true",
		APIKey:      os.Getenv("COINBASE_API_KEY"),
		APISecret:   os.Getenv("COINBASE_API_SECRET"),
		PortfolioID: os.Getenv("COINBASE_PORTFOLIO_ID"),
	}

	cfg.Exchanges["dydx"] = ExchangeConfig{
		Enabled:          os.Getenv("ENABLE_DYDX") == "true",
		APIKey:           os.Getenv("DYDX_API_KEY"),
		APISecret:        os.Getenv("DYDX_API_SECRET"),
		Mnemonic:         os.Getenv("DYDX_MNEMONIC"),
		SubAccountNumber: parseIntEnv("DYDX_SUB_ACCOUNT_NUMBER", 0),
	}

	return cfg, nil
}

// parseIntEnv parses an integer environment variable
func parseIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseFloatEnv parses a float environment variable
func parseFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
