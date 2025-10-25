package config

import "testing"

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
