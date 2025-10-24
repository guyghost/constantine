package execution

import (
	"context"
	"errors"
	"testing"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockOrderManager struct {
	getPositionsFunc  func() []*order.ManagedPosition
	placeOrderFunc    func(ctx context.Context, req *order.OrderRequest) (*exchanges.Order, error)
	closePositionFunc func(ctx context.Context, symbol string) error
}

func (m *mockOrderManager) GetPositions() []*order.ManagedPosition {
	if m.getPositionsFunc != nil {
		return m.getPositionsFunc()
	}
	return nil
}

func (m *mockOrderManager) PlaceOrder(ctx context.Context, req *order.OrderRequest) (*exchanges.Order, error) {
	if m.placeOrderFunc != nil {
		return m.placeOrderFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockOrderManager) ClosePosition(ctx context.Context, symbol string) error {
	if m.closePositionFunc != nil {
		return m.closePositionFunc(ctx, symbol)
	}
	return nil
}

type mockRiskManager struct {
	canTradeFunc              func() (bool, string)
	validateOrderFunc         func(req *order.OrderRequest, openPositions []*order.ManagedPosition) error
	calculatePositionSizeFunc func(entryPrice, stopLoss, accountBalance decimal.Decimal) decimal.Decimal
	getCurrentBalanceFunc     func() decimal.Decimal
}

func (m *mockRiskManager) CanTrade() (bool, string) {
	if m.canTradeFunc != nil {
		return m.canTradeFunc()
	}
	return true, ""
}

func (m *mockRiskManager) ValidateOrder(req *order.OrderRequest, openPositions []*order.ManagedPosition) error {
	if m.validateOrderFunc != nil {
		return m.validateOrderFunc(req, openPositions)
	}
	return nil
}

func (m *mockRiskManager) CalculatePositionSize(entryPrice, stopLoss, accountBalance decimal.Decimal) decimal.Decimal {
	if m.calculatePositionSizeFunc != nil {
		return m.calculatePositionSizeFunc(entryPrice, stopLoss, accountBalance)
	}
	return decimal.Zero
}

func (m *mockRiskManager) GetCurrentBalance() decimal.Decimal {
	if m.getCurrentBalanceFunc != nil {
		return m.getCurrentBalanceFunc()
	}
	return decimal.Zero
}

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

func TestHandleSignal_EntryRiskCheckFailure(t *testing.T) {
	agent := &ExecutionAgent{
		orderManager: &mockOrderManager{},
		riskManager: &mockRiskManager{
			canTradeFunc: func() (bool, string) {
				return false, "cooldown"
			},
		},
		config: Config{
			AutoExecute:       true,
			MinSignalStrength: 0.1,
		},
	}

	err := agent.HandleSignal(context.Background(), &strategy.Signal{
		Type:     strategy.SignalTypeEntry,
		Strength: 1,
	})

	var execErr *ExecutionError
	assert.ErrorAs(t, err, &execErr)
	assert.Equal(t, ExecutionErrorTypeRiskCheckFailed, execErr.Type)
	assert.Equal(t, "cooldown", execErr.Message)
}

func TestHandleSignal_EntrySuccess(t *testing.T) {
	var capturedRequest *order.OrderRequest
	agent := &ExecutionAgent{
		orderManager: &mockOrderManager{
			getPositionsFunc: func() []*order.ManagedPosition {
				return nil
			},
			placeOrderFunc: func(ctx context.Context, req *order.OrderRequest) (*exchanges.Order, error) {
				capturedRequest = req
				return &exchanges.Order{ID: "order-1"}, nil
			},
		},
		riskManager: &mockRiskManager{
			canTradeFunc: func() (bool, string) {
				return true, ""
			},
			validateOrderFunc: func(req *order.OrderRequest, openPositions []*order.ManagedPosition) error {
				return nil
			},
			calculatePositionSizeFunc: func(entryPrice, stopLoss, accountBalance decimal.Decimal) decimal.Decimal {
				return decimal.NewFromFloat(0.1)
			},
			getCurrentBalanceFunc: func() decimal.Decimal {
				return decimal.NewFromInt(1000)
			},
		},
		config: Config{
			AutoExecute:       true,
			MinSignalStrength: 0.1,
			StopLossPercent:   decimal.NewFromFloat(0.01),
			TakeProfitPercent: decimal.NewFromFloat(0.02),
		},
	}

	price := decimal.NewFromInt(100)
	signal := &strategy.Signal{
		Type:     strategy.SignalTypeEntry,
		Strength: 0.5,
		Side:     exchanges.OrderSideBuy,
		Price:    price,
		Symbol:   "BTC-USD",
	}

	err := agent.HandleSignal(context.Background(), signal)

	assert.NoError(t, err)
	if assert.NotNil(t, capturedRequest) {
		assert.Equal(t, signal.Symbol, capturedRequest.Symbol)
		assert.Equal(t, signal.Side, capturedRequest.Side)
		assert.Equal(t, price, capturedRequest.Price)
		assert.Equal(t, decimal.NewFromFloat(0.1), capturedRequest.Amount)
		assert.Equal(t, price.Mul(decimal.NewFromInt(1).Sub(agent.config.StopLossPercent)), capturedRequest.StopLoss)
		assert.Equal(t, price.Mul(decimal.NewFromInt(1).Add(agent.config.TakeProfitPercent)), capturedRequest.TakeProfit)
	}
}

func TestHandleSignal_EntryValidationFailure(t *testing.T) {
	validationErr := errors.New("validation failed")
	agent := &ExecutionAgent{
		orderManager: &mockOrderManager{
			getPositionsFunc: func() []*order.ManagedPosition {
				return nil
			},
		},
		riskManager: &mockRiskManager{
			canTradeFunc: func() (bool, string) {
				return true, ""
			},
			validateOrderFunc: func(req *order.OrderRequest, openPositions []*order.ManagedPosition) error {
				return validationErr
			},
			calculatePositionSizeFunc: func(entryPrice, stopLoss, accountBalance decimal.Decimal) decimal.Decimal {
				return decimal.NewFromFloat(0.1)
			},
			getCurrentBalanceFunc: func() decimal.Decimal {
				return decimal.NewFromInt(1000)
			},
		},
		config: Config{
			AutoExecute:       true,
			MinSignalStrength: 0,
			StopLossPercent:   decimal.NewFromFloat(0.01),
		},
	}

	signal := &strategy.Signal{
		Type:     strategy.SignalTypeEntry,
		Strength: 1,
		Side:     exchanges.OrderSideBuy,
		Price:    decimal.NewFromInt(100),
		Symbol:   "BTC-USD",
	}

	err := agent.HandleSignal(context.Background(), signal)

	var execErr *ExecutionError
	assert.ErrorAs(t, err, &execErr)
	assert.Equal(t, ExecutionErrorTypeRiskValidationFailed, execErr.Type)
	assert.Equal(t, validationErr.Error(), execErr.Message)
}

func TestHandleSignal_ExitBypassesRiskCheck(t *testing.T) {
	closed := false
	agent := &ExecutionAgent{
		orderManager: &mockOrderManager{
			closePositionFunc: func(ctx context.Context, symbol string) error {
				closed = true
				assert.Equal(t, "BTC-USD", symbol)
				return nil
			},
		},
		riskManager: &mockRiskManager{
			canTradeFunc: func() (bool, string) {
				t.Fatalf("risk check should not run for exit signals")
				return false, ""
			},
		},
		config: Config{
			AutoExecute:       true,
			MinSignalStrength: 0,
		},
	}

	err := agent.HandleSignal(context.Background(), &strategy.Signal{
		Type:     strategy.SignalTypeExit,
		Strength: 1,
		Symbol:   "BTC-USD",
	})

	assert.NoError(t, err)
	assert.True(t, closed)
}

func TestHandleSignal_ExitCloseError(t *testing.T) {
	agent := &ExecutionAgent{
		orderManager: &mockOrderManager{
			closePositionFunc: func(ctx context.Context, symbol string) error {
				return errors.New("close failed")
			},
		},
		riskManager: &mockRiskManager{},
		config: Config{
			AutoExecute:       true,
			MinSignalStrength: 0,
		},
	}

	err := agent.HandleSignal(context.Background(), &strategy.Signal{
		Type:     strategy.SignalTypeExit,
		Strength: 1,
		Symbol:   "BTC-USD",
	})

	var execErr *ExecutionError
	assert.ErrorAs(t, err, &execErr)
	assert.Equal(t, ExecutionErrorTypePositionCloseFailed, execErr.Type)
	assert.Equal(t, "close failed", execErr.Message)
}
