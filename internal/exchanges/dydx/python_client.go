package dydx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// PythonClient wraps the official dYdX v4 Python client for order placement
// This is a temporary solution until we have native Go proto support
type PythonClient struct {
	pythonPath string
	scriptPath string
	network    string // "testnet" or "mainnet"
	mnemonic   string
}

// PythonClientConfig contains configuration for the Python client wrapper
type PythonClientConfig struct {
	PythonPath string // Path to python3 executable (default: "python3")
	ScriptPath string // Path to the Python script (default: internal script)
	Network    string // "testnet" or "mainnet"
	Mnemonic   string // BIP39 mnemonic for wallet
}

// NewPythonClient creates a new Python client wrapper
func NewPythonClient(config *PythonClientConfig) *PythonClient {
	pythonPath := config.PythonPath
	if pythonPath == "" {
		pythonPath = "python3"
	}

	return &PythonClient{
		pythonPath: pythonPath,
		scriptPath: config.ScriptPath,
		network:    config.Network,
		mnemonic:   config.Mnemonic,
	}
}

// PlaceOrderRequest represents a Python client order request
type PythonOrderRequest struct {
	Market      string  `json:"market"`
	Side        string  `json:"side"`
	Type        string  `json:"type"`
	Size        float64 `json:"size"`
	Price       float64 `json:"price"`
	TimeInForce string  `json:"timeInForce,omitempty"`
	ReduceOnly  bool    `json:"reduceOnly,omitempty"`
	PostOnly    bool    `json:"postOnly,omitempty"`
	ClientID    string  `json:"clientId,omitempty"`
}

// PythonOrderResponse represents the response from Python client
type PythonOrderResponse struct {
	Success  bool   `json:"success"`
	OrderID  string `json:"orderId,omitempty"`
	ClientID string `json:"clientId,omitempty"`
	Error    string `json:"error,omitempty"`
	TxHash   string `json:"txHash,omitempty"`
}

// PlaceOrder places an order using the Python client
func (c *PythonClient) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	// Convert order to Python request format
	side := "BUY"
	if order.Side == "sell" || order.Side == "SELL" {
		side = "SELL"
	}

	orderType := "LIMIT"
	if order.Type == "market" || order.Type == "MARKET" {
		orderType = "MARKET"
	}

	size, _ := order.Amount.Float64()
	price, _ := order.Price.Float64()

	pyRequest := PythonOrderRequest{
		Market:   order.Symbol,
		Side:     side,
		Type:     orderType,
		Size:     size,
		Price:    price,
		ClientID: order.ID,
	}

	// Execute Python script
	response, err := c.executePythonScript(ctx, "place_order", pyRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to execute Python script: %w", err)
	}

	var pyResponse PythonOrderResponse
	if err := json.Unmarshal(response, &pyResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Python response: %w", err)
	}

	if !pyResponse.Success {
		return nil, fmt.Errorf("order placement failed: %s", pyResponse.Error)
	}

	// Update order with response data
	order.ID = pyResponse.OrderID
	if pyResponse.ClientID != "" {
		order.ClientOrderID = pyResponse.ClientID
	}
	order.Status = "open"
	order.CreatedAt = time.Now()

	return order, nil
}

// CancelOrder cancels an order using the Python client
func (c *PythonClient) CancelOrder(ctx context.Context, orderID string) error {
	request := map[string]string{
		"orderId": orderID,
	}

	response, err := c.executePythonScript(ctx, "cancel_order", request)
	if err != nil {
		return fmt.Errorf("failed to execute Python script: %w", err)
	}

	var pyResponse PythonOrderResponse
	if err := json.Unmarshal(response, &pyResponse); err != nil {
		return fmt.Errorf("failed to parse Python response: %w", err)
	}

	if !pyResponse.Success {
		return fmt.Errorf("order cancellation failed: %s", pyResponse.Error)
	}

	return nil
}

// executePythonScript executes a Python script with the given command and data
func (c *PythonClient) executePythonScript(ctx context.Context, command string, data interface{}) ([]byte, error) {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// Prepare script input
	input := map[string]interface{}{
		"command":  command,
		"network":  c.network,
		"mnemonic": c.mnemonic,
		"data":     json.RawMessage(jsonData),
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Execute Python script
	cmd := exec.CommandContext(ctx, c.pythonPath, c.scriptPath)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Python script error: %s\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// GetBalance gets account balance (for compatibility)
func (c *PythonClient) GetBalance(ctx context.Context) (map[string]decimal.Decimal, error) {
	request := map[string]string{}

	response, err := c.executePythonScript(ctx, "get_balance", request)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	var result struct {
		Success bool                       `json:"success"`
		Balance map[string]decimal.Decimal `json:"balance"`
		Error   string                     `json:"error"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("get balance failed: %s", result.Error)
	}

	return result.Balance, nil
}
