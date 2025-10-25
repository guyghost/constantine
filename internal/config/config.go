package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// ExchangeConfig represents configuration for an exchange integration.
type ExchangeConfig struct {
	Name             string
	Enabled          bool
	APIKey           string
	APISecret        string
	Mnemonic         string
	WalletAddress    string
	SubAccountNumber int
	PortfolioID      string
}

// AppConfig aggregates configuration for the bot runtime.
type AppConfig struct {
	Environment    string
	TelemetryAddr  string
	InitialBalance decimal.Decimal
	StrategySymbol string
	Exchanges      map[string]*ExchangeConfig
}

// Load loads configuration from environment variables and validates it.
func Load() (*AppConfig, error) {
	cfg := &AppConfig{
		Environment:    getEnv("APP_ENV", "development"),
		TelemetryAddr:  getEnv("TELEMETRY_ADDR", ":9100"),
		InitialBalance: getEnvDecimal("INITIAL_BALANCE", decimal.NewFromFloat(10000)),
		StrategySymbol: getEnv("TRADING_SYMBOL", "BTC-USD"),
		Exchanges: map[string]*ExchangeConfig{
			"hyperliquid": {
				Name:      "hyperliquid",
				Enabled:   getEnvBool("ENABLE_HYPERLIQUID", true),
				APIKey:    os.Getenv("HYPERLIQUID_API_KEY"),
				APISecret: os.Getenv("HYPERLIQUID_API_SECRET"),
			},
			"coinbase": {
				Name:        "coinbase",
				Enabled:     getEnvBool("ENABLE_COINBASE", true),
				APIKey:      os.Getenv("COINBASE_API_KEY"),
				APISecret:   os.Getenv("COINBASE_API_SECRET"),
				PortfolioID: os.Getenv("COINBASE_PORTFOLIO_ID"),
			},
			"dydx": {
				Name:             "dydx",
				Enabled:          getEnvBool("ENABLE_DYDX", false),
				APIKey:           os.Getenv("DYDX_API_KEY"),
				APISecret:        os.Getenv("DYDX_API_SECRET"),
				Mnemonic:         os.Getenv("DYDX_MNEMONIC"),
				WalletAddress:    os.Getenv("DYDX_WALLET_ADDRESS"),
				SubAccountNumber: getEnvInt("DYDX_SUBACCOUNT_NUMBER", 0),
			},
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *AppConfig) validate() error {
	var missing []string

	if exchange, ok := c.Exchanges["hyperliquid"]; ok && exchange.Enabled {
		if exchange.APIKey == "" {
			missing = append(missing, "HYPERLIQUID_API_KEY")
		}
		if exchange.APISecret == "" {
			missing = append(missing, "HYPERLIQUID_API_SECRET")
		}
	}

	if exchange, ok := c.Exchanges["coinbase"]; ok && exchange.Enabled {
		if exchange.APIKey == "" {
			missing = append(missing, "COINBASE_API_KEY")
		}
		if exchange.APISecret == "" {
			missing = append(missing, "COINBASE_API_SECRET")
		}
	}

	if exchange, ok := c.Exchanges["dydx"]; ok && exchange.Enabled {
		hasMnemonic := exchange.Mnemonic != ""
		hasAPIKeys := exchange.APIKey != "" && exchange.APISecret != ""
		if !hasMnemonic && !hasAPIKeys {
			missing = append(missing, "DYDX_MNEMONIC or DYDX_API_KEY/DYDX_API_SECRET")
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(value) {
	case "true", "1", "yes", "y", "on":
		return true
	case "false", "0", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

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

func getEnvDecimal(key string, defaultValue decimal.Decimal) decimal.Decimal {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if parsed, err := decimal.NewFromString(value); err == nil {
		return parsed
	}
	return defaultValue
}
