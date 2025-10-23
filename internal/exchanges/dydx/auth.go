package dydx

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

// Wallet represents a dYdX wallet derived from mnemonic
type Wallet struct {
	Mnemonic         string
	Address          string
	PrivateKeyHex    string
	SubAccountNumber int
}

// NewWalletFromMnemonic creates a wallet from a mnemonic phrase
func NewWalletFromMnemonic(mnemonic string, subAccountNumber int) (*Wallet, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic phrase")
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Create master key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// dYdX uses Cosmos derivation path: m/44'/118'/0'/0/0
	// 118 is the coin type for Cosmos
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, fmt.Errorf("failed to derive purpose: %w", err)
	}

	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + 118) // Cosmos coin type
	if err != nil {
		return nil, fmt.Errorf("failed to derive coin type: %w", err)
	}

	account, err := coinType.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account: %w", err)
	}

	change, err := account.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive change: %w", err)
	}

	addressIndex, err := change.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address index: %w", err)
	}

	// Get private key
	privKey, err := addressIndex.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	privKeyHex := hex.EncodeToString(privKey.Serialize())

	// Derive dYdX address from public key (Cosmos format)
	pubKey := privKey.PubKey()
	address := deriveCosmosAddress(pubKey.SerializeCompressed())

	return &Wallet{
		Mnemonic:         mnemonic,
		Address:          address,
		PrivateKeyHex:    privKeyHex,
		SubAccountNumber: subAccountNumber,
	}, nil
}

// deriveCosmosAddress derives a Cosmos address from a public key
func deriveCosmosAddress(pubKey []byte) string {
	// SHA256 hash of public key
	sha256Hash := sha256.Sum256(pubKey)

	// RIPEMD160 hash of SHA256
	ripemd := ripemd160.New()
	ripemd.Write(sha256Hash[:])
	addressBytes := ripemd.Sum(nil)

	// Convert to bech32 with "dydx" prefix
	address, err := bech32.ConvertAndEncode("dydx", addressBytes)
	if err != nil {
		// Fallback to hex if bech32 fails
		return "dydx" + hex.EncodeToString(addressBytes)
	}

	return address
}

// GenerateMnemonic generates a new random mnemonic phrase
func GenerateMnemonic() (string, error) {
	// Generate 256-bit entropy (24 words)
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// ValidateMnemonic validates a mnemonic phrase
func ValidateMnemonic(mnemonic string) error {
	// Normalize mnemonic
	normalized := strings.TrimSpace(mnemonic)
	words := strings.Fields(normalized)

	// Check word count (12 or 24 words)
	if len(words) != 12 && len(words) != 24 {
		return fmt.Errorf("mnemonic must be 12 or 24 words, got %d", len(words))
	}

	// Validate using BIP39
	if !bip39.IsMnemonicValid(normalized) {
		return fmt.Errorf("invalid mnemonic phrase")
	}

	return nil
}

// SubAccountAddress returns the subaccount address
func (w *Wallet) SubAccountAddress() string {
	return fmt.Sprintf("%s/%d", w.Address, w.SubAccountNumber)
}
