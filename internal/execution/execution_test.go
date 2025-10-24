package execution

import (
	"testing"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, decimal.NewFromFloat(0.005), config.StopLossPercent)
	assert.Equal(t, decimal.NewFromFloat(0.01), config.TakeProfitPercent)
	assert.Equal(t, 0.5, config.MinSignalStrength)
	assert.True(t, config.AutoExecute)
}

func TestCalculateStopLoss_BuyOrder(t *testing.T) {
	agent := &ExecutionAgent{
		config: DefaultConfig(),
	}
	signal := &strategy.Signal{
		Side:  exchanges.OrderSideBuy,
		Price: decimal.NewFromFloat(50000),
	}

	stopLoss := agent.calculateStopLoss(signal)
	expected := decimal.NewFromFloat(50000).Mul(decimal.NewFromFloat(0.995)) // 0.5% below

	assert.Equal(t, expected, stopLoss)
}

func TestCalculateStopLoss_SellOrder(t *testing.T) {
	agent := &ExecutionAgent{
		config: DefaultConfig(),
	}
	signal := &strategy.Signal{
		Side:  exchanges.OrderSideSell,
		Price: decimal.NewFromFloat(50000),
	}

	stopLoss := agent.calculateStopLoss(signal)
	expected := decimal.NewFromFloat(50000).Mul(decimal.NewFromFloat(1.005)) // 0.5% above

	assert.Equal(t, expected, stopLoss)
}

func TestCalculateTakeProfit_BuyOrder(t *testing.T) {
	agent := &ExecutionAgent{
		config: DefaultConfig(),
	}
	signal := &strategy.Signal{
		Side:  exchanges.OrderSideBuy,
		Price: decimal.NewFromFloat(50000),
	}

	takeProfit := agent.calculateTakeProfit(signal)
	expected := decimal.NewFromFloat(50000).Mul(decimal.NewFromFloat(1.01)) // 1% above

	assert.Equal(t, expected, takeProfit)
}

func TestCalculateTakeProfit_SellOrder(t *testing.T) {
	agent := &ExecutionAgent{
		config: DefaultConfig(),
	}
	signal := &strategy.Signal{
		Side:  exchanges.OrderSideSell,
		Price: decimal.NewFromFloat(50000),
	}

	takeProfit := agent.calculateTakeProfit(signal)
	expected := decimal.NewFromFloat(50000).Mul(decimal.NewFromFloat(0.99)) // 1% below

	assert.Equal(t, expected, takeProfit)
}

func TestExecutionError_Error(t *testing.T) {
	err := &ExecutionError{
		Type:    ExecutionErrorTypeOrderPlacementFailed,
		Message: "test error",
	}

	assert.Equal(t, "test error", err.Error())
}

func TestExecutionErrorTypes(t *testing.T) {
	assert.Equal(t, ExecutionErrorType(0), ExecutionErrorTypeRiskCheckFailed)
	assert.Equal(t, ExecutionErrorType(1), ExecutionErrorTypeInvalidSignal)
	assert.Equal(t, ExecutionErrorType(2), ExecutionErrorTypeRiskValidationFailed)
	assert.Equal(t, ExecutionErrorType(3), ExecutionErrorTypeOrderPlacementFailed)
	assert.Equal(t, ExecutionErrorType(4), ExecutionErrorTypePositionCloseFailed)
}
