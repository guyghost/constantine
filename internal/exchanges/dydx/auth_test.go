package dydx

import (
	"testing"
)

func TestAddressGeneration(t *testing.T) {
	// Test mnemonic from BIP39 test vectors
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	wallet, err := NewWalletFromMnemonic(mnemonic, 0)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	expectedAddress := "dydx19rl4cm2hmr8afy4kldpxz3fka4jguq0a4erelz"
	if wallet.Address != expectedAddress {
		t.Errorf("Address mismatch: got %s, expected %s", wallet.Address, expectedAddress)
	}

	// Verify address starts with dydx1
	if len(wallet.Address) < 5 || wallet.Address[:5] != "dydx1" {
		t.Errorf("Address should start with dydx1: %s", wallet.Address)
	}

	// Verify private key is present
	if wallet.PrivateKeyHex == "" {
		t.Error("Private key should not be empty")
	}
}
