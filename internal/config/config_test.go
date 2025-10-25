package config

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestLoad_SucceedsWithRequiredSecrets(t *testing.T) {
	t.Setenv("HYPERLIQUID_API_KEY", "test-key")
	t.Setenv("HYPERLIQUID_API_SECRET", "test-secret")
	t.Setenv("ENABLE_COINBASE", "false")
	t.Setenv("ENABLE_DYDX", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	hl := cfg.Exchanges["hyperliquid"]
	if hl == nil || hl.APIKey != "test-key" || hl.APISecret != "test-secret" {
		t.Fatalf("hyperliquid config not populated correctly: %+v", hl)
	}
}

func TestLoad_FailsWhenHyperliquidSecretsMissing(t *testing.T) {
	t.Setenv("ENABLE_COINBASE", "false")
	t.Setenv("ENABLE_DYDX", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when hyperliquid secrets are missing")
	}
}

func TestLoad_FailsWhenCoinbaseSecretMissing(t *testing.T) {
	t.Setenv("HYPERLIQUID_API_KEY", "test-key")
	t.Setenv("HYPERLIQUID_API_SECRET", "test-secret")
	t.Setenv("ENABLE_COINBASE", "true")
	t.Setenv("COINBASE_API_KEY", "coinbase-key")
	t.Setenv("COINBASE_API_SECRET", "")
	t.Setenv("ENABLE_DYDX", "false")

	if _, err := Load(); err == nil {
		t.Fatal("expected error with missing coinbase secret")
	}
}

func TestLoad_FailsWhenDydxMissingAuth(t *testing.T) {
	t.Setenv("HYPERLIQUID_API_KEY", "test-key")
	t.Setenv("HYPERLIQUID_API_SECRET", "test-secret")
	t.Setenv("ENABLE_COINBASE", "false")
	t.Setenv("ENABLE_DYDX", "true")

	if _, err := Load(); err == nil {
		t.Fatal("expected error when dydx enabled without credentials")
	}
}

func TestLoad_AllowsTelemetryAndStrategyOverrides(t *testing.T) {
	t.Setenv("HYPERLIQUID_API_KEY", "test-key")
	t.Setenv("HYPERLIQUID_API_SECRET", "test-secret")
	t.Setenv("ENABLE_COINBASE", "false")
	t.Setenv("ENABLE_DYDX", "false")
	t.Setenv("TELEMETRY_ADDR", ":9200")
	t.Setenv("TRADING_SYMBOL", "ETH-USD")
	t.Setenv("INITIAL_BALANCE", "25000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error: %v", err)
	}

	if cfg.TelemetryAddr != ":9200" {
		t.Fatalf("expected telemetry addr override, got %s", cfg.TelemetryAddr)
	}
	if cfg.StrategySymbol != "ETH-USD" {
		t.Fatalf("expected strategy symbol override, got %s", cfg.StrategySymbol)
	}
	if !cfg.InitialBalance.Equal(decimal.NewFromInt(25000)) {
		t.Fatalf("expected initial balance override, got %s", cfg.InitialBalance)
	}
}
