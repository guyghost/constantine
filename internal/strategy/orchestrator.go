package strategy

import (
	"context"
	"fmt"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/symbolmanager"
)

// SymbolManagerInterface defines the interface for symbol management
type SymbolManagerInterface interface {
	GetActiveSymbols() []string
	GetSymbolConfig(symbol string) (*symbolmanager.SymbolConfig, error)
	IsSymbolActive(symbol string) bool
}

// StrategyOrchestrator manages multiple strategy instances for different symbols
type StrategyOrchestrator struct {
	strategies    map[string]*ScalpingStrategy
	symbolManager SymbolManagerInterface
}

// NewStrategyOrchestrator creates a new strategy orchestrator
func NewStrategyOrchestrator(symbolManager SymbolManagerInterface) *StrategyOrchestrator {
	return &StrategyOrchestrator{
		strategies:    make(map[string]*ScalpingStrategy),
		symbolManager: symbolManager,
	}
}

// StartSymbol initializes and starts a strategy for a specific symbol
func (so *StrategyOrchestrator) StartSymbol(ctx context.Context, symbol string) error {
	// Check if symbol is active
	if !so.symbolManager.IsSymbolActive(symbol) {
		return fmt.Errorf("symbol %s is not active", symbol)
	}

	// Check if strategy already exists
	if _, exists := so.strategies[symbol]; exists {
		return fmt.Errorf("strategy for symbol %s already exists", symbol)
	}

	// Get symbol configuration
	symbolConfig, err := so.symbolManager.GetSymbolConfig(symbol)
	if err != nil {
		return fmt.Errorf("failed to get config for symbol %s: %w", symbol, err)
	}

	// Create strategy instance with mock exchange for now
	// TODO: Pass appropriate exchange based on symbol
	mockExchange := &exchanges.MockExchange{}
	strategy := NewScalpingStrategy(symbolConfig.StrategyConfig, mockExchange)

	so.strategies[symbol] = strategy

	return nil
}

// StopSymbol stops and removes the strategy for a specific symbol
func (so *StrategyOrchestrator) StopSymbol(symbol string) error {
	if _, exists := so.strategies[symbol]; !exists {
		return fmt.Errorf("strategy for symbol %s not found", symbol)
	}

	delete(so.strategies, symbol)

	return nil
}

// GetSymbolStrategy returns the strategy instance for a specific symbol
func (so *StrategyOrchestrator) GetSymbolStrategy(symbol string) (*ScalpingStrategy, error) {
	strategy, exists := so.strategies[symbol]
	if !exists {
		return nil, fmt.Errorf("strategy for symbol %s not found", symbol)
	}

	return strategy, nil
}

// GetActiveStrategies returns all currently active strategy instances
func (so *StrategyOrchestrator) GetActiveStrategies() map[string]*ScalpingStrategy {
	active := make(map[string]*ScalpingStrategy)
	for symbol, strategy := range so.strategies {
		active[symbol] = strategy
	}

	return active
}

// ProcessMarketData processes market data for all active symbols
func (so *StrategyOrchestrator) ProcessMarketData(ctx context.Context, symbol string, candle exchanges.Candle) error {
	if _, exists := so.strategies[symbol]; !exists {
		// Symbol not active, skip
		return nil
	}

	strategy, exists := so.strategies[symbol]
	if !exists {
		return nil
	}

	// Process the candle data through the strategy
	strategy.ProcessCandle(candle)

	return nil
}

// GenerateSignals generates trading signals for all active symbols
func (so *StrategyOrchestrator) GenerateSignals(ctx context.Context) map[string]*Signal {
	signals := make(map[string]*Signal)

	for symbol := range so.strategies {
		// Generate signal for this symbol
		// This would use the strategy's signal generator with current market data
		// For now, return nil (no signal)

		signals[symbol] = nil
	}

	return signals
}

// UpdateActiveSymbols synchronizes active strategies with symbol manager
func (so *StrategyOrchestrator) UpdateActiveSymbols(ctx context.Context) error {
	activeSymbols := so.symbolManager.GetActiveSymbols()

	// Start strategies for new active symbols
	for _, symbol := range activeSymbols {
		if _, exists := so.strategies[symbol]; !exists {
			// Start new strategy
			if err := so.startSymbolLocked(ctx, symbol); err != nil {
				// Log error but continue
				fmt.Printf("Failed to start strategy for symbol %s: %v\n", symbol, err)
			}
		}
	}

	// Stop strategies for inactive symbols
	for symbol := range so.strategies {
		found := false
		for _, active := range activeSymbols {
			if active == symbol {
				found = true
				break
			}
		}
		if !found {
			// Stop strategy
			delete(so.strategies, symbol)
		}
	}

	return nil
}

// startSymbolLocked starts a strategy for a symbol (assumes lock is held)
func (so *StrategyOrchestrator) startSymbolLocked(ctx context.Context, symbol string) error {
	// Get symbol configuration
	symbolConfig, err := so.symbolManager.GetSymbolConfig(symbol)
	if err != nil {
		return fmt.Errorf("failed to get config for symbol %s: %w", symbol, err)
	}

	// Create strategy instance with mock exchange for now
	// TODO: Pass appropriate exchange based on symbol
	mockExchange := &exchanges.MockExchange{}
	strategy := NewScalpingStrategy(symbolConfig.StrategyConfig, mockExchange)

	so.strategies[symbol] = strategy

	return nil
}

// GetStrategyMetrics returns performance metrics for all strategies
func (so *StrategyOrchestrator) GetStrategyMetrics() map[string]StrategyMetrics {
	metrics := make(map[string]StrategyMetrics)

	for symbol := range so.strategies {
		// Get metrics from strategy
		// This would include signal counts, performance, etc.
		metrics[symbol] = StrategyMetrics{
			Symbol: symbol,
			// Populate with actual metrics
		}
	}

	return metrics
}

// StrategyMetrics holds performance metrics for a strategy
type StrategyMetrics struct {
	Symbol           string
	SignalsGenerated int
	SignalsExecuted  int
	WinRate          float64
	TotalPnL         float64
}
