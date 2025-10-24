package errors

import (
	"errors"
	"fmt"
)

// Operation identifies the type of order-manager operation that generated an error.
type Operation string

const (
	OperationValidate        Operation = "validate"
	OperationPlace           Operation = "place_order"
	OperationPlaceStopLoss   Operation = "place_stop_loss"
	OperationPlaceTakeProfit Operation = "place_take_profit"
	OperationCancel          Operation = "cancel_order"
)

// OrderError provides additional context for order-related failures.
type OrderError struct {
	Op     Operation
	Target string
	Err    error
}

// Error implements the error interface.
func (e *OrderError) Error() string {
	if e == nil {
		return ""
	}
	if e.Target != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Target, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the wrapped error.
func (e *OrderError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// New constructs a new OrderError.
func New(op Operation, target string, err error) error {
	if err == nil {
		return nil
	}
	var oe *OrderError
	if errors.As(err, &oe) {
		return err
	}
	return &OrderError{Op: op, Target: target, Err: err}
}
