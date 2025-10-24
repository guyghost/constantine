package execution

import (
	"context"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
)

// OrderManager defines the minimal behavior required from an order manager.
type OrderManager interface {
	GetPositions() []*order.ManagedPosition
	PlaceOrder(ctx context.Context, req *order.OrderRequest) (*exchanges.Order, error)
	ClosePosition(ctx context.Context, symbol string) error
}

// RiskManager defines the minimal behavior required from a risk manager.
type RiskManager interface {
	CanTrade() (bool, string)
	ValidateOrder(req *order.OrderRequest, openPositions []*order.ManagedPosition) error
	CalculatePositionSize(entryPrice decimal.Decimal, stopLoss decimal.Decimal, accountBalance decimal.Decimal) decimal.Decimal
	GetCurrentBalance() decimal.Decimal
}

// ExecutionAgent handles automated order placement based on trading signals
type ExecutionAgent struct {
	orderManager OrderManager
	riskManager  RiskManager
	config       Config
}

// Config holds configuration for the execution agent
type Config struct {
	// Risk management parameters
	StopLossPercent   decimal.Decimal // e.g., 0.005 for 0.5%
	TakeProfitPercent decimal.Decimal // e.g., 0.01 for 1%

	// Signal thresholds
	MinSignalStrength float64 // Minimum signal strength to execute (0.0-1.0)

	// Execution settings
	AutoExecute bool // Whether to automatically execute orders
}

// DefaultConfig returns default execution configuration
func DefaultConfig() Config {
	return Config{
		StopLossPercent:   decimal.NewFromFloat(0.005), // 0.5%
		TakeProfitPercent: decimal.NewFromFloat(0.01),  // 1%
		MinSignalStrength: 0.5,                         // 50%
		AutoExecute:       true,
	}
}

// NewExecutionAgent creates a new execution agent
func NewExecutionAgent(orderManager OrderManager, riskManager RiskManager, config Config) *ExecutionAgent {
	return &ExecutionAgent{
		orderManager: orderManager,
		riskManager:  riskManager,
		config:       config,
	}
}

// HandleSignal processes a trading signal and executes orders if conditions are met
func (e *ExecutionAgent) HandleSignal(ctx context.Context, signal *strategy.Signal) error {
	// Check if auto-execution is enabled
	if !e.config.AutoExecute {
		return nil
	}

	// Check signal strength threshold
	if signal.Strength < e.config.MinSignalStrength {
		return nil
	}

	switch signal.Type {
	case strategy.SignalTypeEntry:
		canTrade, reason := e.riskManager.CanTrade()
		if !canTrade {
			return &ExecutionError{
				Type:    ExecutionErrorTypeRiskCheckFailed,
				Message: reason,
			}
		}
		return e.handleEntrySignal(ctx, signal)
	case strategy.SignalTypeExit:
		return e.handleExitSignal(ctx, signal)
	default:
		return &ExecutionError{
			Type:    ExecutionErrorTypeInvalidSignal,
			Message: "Unknown signal type",
		}
	}
}

// handleEntrySignal handles entry signals by placing orders
func (e *ExecutionAgent) handleEntrySignal(ctx context.Context, signal *strategy.Signal) error {
	// Calculate stop loss price
	stopLoss := e.calculateStopLoss(signal)

	// Get current balance for position sizing
	balance := e.riskManager.GetCurrentBalance()

	// Calculate position size based on risk management
	positionSize := e.riskManager.CalculatePositionSize(signal.Price, stopLoss, balance)

	// Calculate take profit price
	takeProfit := e.calculateTakeProfit(signal)

	// Create order request
	req := &order.OrderRequest{
		Symbol:     signal.Symbol,
		Side:       signal.Side,
		Type:       exchanges.OrderTypeLimit,
		Price:      signal.Price,
		Amount:     positionSize,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
	}

	// Validate order with risk manager
	positions := e.orderManager.GetPositions()
	if err := e.riskManager.ValidateOrder(req, positions); err != nil {
		return &ExecutionError{
			Type:    ExecutionErrorTypeRiskValidationFailed,
			Message: err.Error(),
		}
	}

	// Place the order
	placedOrder, err := e.orderManager.PlaceOrder(ctx, req)
	if err != nil {
		return &ExecutionError{
			Type:    ExecutionErrorTypeOrderPlacementFailed,
			Message: err.Error(),
		}
	}

	// Log successful order placement
	// (Logging will be handled by the order manager callbacks)

	_ = placedOrder // Use the placed order if needed for logging

	return nil
}

// handleExitSignal handles exit signals by closing positions
func (e *ExecutionAgent) handleExitSignal(ctx context.Context, signal *strategy.Signal) error {
	// Close position for the symbol
	if err := e.orderManager.ClosePosition(ctx, signal.Symbol); err != nil {
		return &ExecutionError{
			Type:    ExecutionErrorTypePositionCloseFailed,
			Message: err.Error(),
		}
	}

	return nil
}

// calculateStopLoss calculates the stop loss price based on signal side
func (e *ExecutionAgent) calculateStopLoss(signal *strategy.Signal) decimal.Decimal {
	if signal.Side == exchanges.OrderSideBuy {
		// For buy orders, stop loss is below entry price
		return signal.Price.Mul(decimal.NewFromInt(1).Sub(e.config.StopLossPercent))
	}
	// For sell orders, stop loss is above entry price
	return signal.Price.Mul(decimal.NewFromInt(1).Add(e.config.StopLossPercent))
}

// calculateTakeProfit calculates the take profit price based on signal side
func (e *ExecutionAgent) calculateTakeProfit(signal *strategy.Signal) decimal.Decimal {
	if signal.Side == exchanges.OrderSideBuy {
		// For buy orders, take profit is above entry price
		return signal.Price.Mul(decimal.NewFromInt(1).Add(e.config.TakeProfitPercent))
	}
	// For sell orders, take profit is below entry price
	return signal.Price.Mul(decimal.NewFromInt(1).Sub(e.config.TakeProfitPercent))
}

// ExecutionError represents an error that occurred during order execution
type ExecutionError struct {
	Type    ExecutionErrorType
	Message string
}

func (e *ExecutionError) Error() string {
	return e.Message
}

// ExecutionErrorType defines the type of execution error
type ExecutionErrorType int

const (
	ExecutionErrorTypeRiskCheckFailed ExecutionErrorType = iota
	ExecutionErrorTypeInvalidSignal
	ExecutionErrorTypeRiskValidationFailed
	ExecutionErrorTypeOrderPlacementFailed
	ExecutionErrorTypePositionCloseFailed
)
