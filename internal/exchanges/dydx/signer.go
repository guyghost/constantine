package dydx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Signer handles signing of dYdX requests
type Signer struct {
	wallet *Wallet
}

// NewSigner creates a new signer with a wallet
func NewSigner(wallet *Wallet) *Signer {
	return &Signer{
		wallet: wallet,
	}
}

// SignRequest signs a request with the wallet's private key
func (s *Signer) SignRequest(method, path string, body interface{}) (string, string, error) {
	// Generate timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Create signature payload
	var bodyStr string
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyStr = string(bodyBytes)
	}

	// Concatenate for signature: timestamp + method + path + body
	message := timestamp + method + path + bodyStr

	// Sign with private key
	signature, err := s.sign(message)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign message: %w", err)
	}

	return signature, timestamp, nil
}

// sign creates an HMAC-SHA256 signature
func (s *Signer) sign(message string) (string, error) {
	// Decode private key from hex
	privateKeyBytes, err := hex.DecodeString(s.wallet.PrivateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Create HMAC-SHA256
	h := hmac.New(sha256.New, privateKeyBytes)
	h.Write([]byte(message))
	signature := h.Sum(nil)

	// Encode as base64
	return base64.StdEncoding.EncodeToString(signature), nil
}

// SignOrderPlacement signs an order placement request
func (s *Signer) SignOrderPlacement(order *OrderRequest) (map[string]string, error) {
	// Create order hash for signature
	orderHash := s.hashOrder(order)

	// Sign the order hash
	signature, err := s.sign(orderHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}

	// Return headers
	headers := map[string]string{
		"DYDX-SIGNATURE":  signature,
		"DYDX-TIMESTAMP":  strconv.FormatInt(time.Now().Unix(), 10),
		"DYDX-ETHEREUM-ADDRESS": s.wallet.Address,
	}

	return headers, nil
}

// hashOrder creates a hash of the order for signing
func (s *Signer) hashOrder(order *OrderRequest) string {
	// Create a canonical representation of the order
	parts := []string{
		order.Market,
		order.Side,
		order.Type,
		order.Size.String(),
		order.Price.String(),
		order.TimeInForce,
	}

	// Join and hash
	message := ""
	for _, part := range parts {
		message += part + "|"
	}

	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:])
}

// GetWallet returns the associated wallet
func (s *Signer) GetWallet() *Wallet {
	return s.wallet
}
