package portfolio

import (
	"context"
	"sync"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// PortfolioManager manages portfolio-level position tracking across multiple symbols and exchanges
type PortfolioManager struct {
	mu          sync.RWMutex
	multiplexer *exchanges.ExchangeMultiplexer
	positions   map[string][]exchanges.Position // symbol -> positions
	totalPnL    decimal.Decimal
}

// NewPortfolioManager creates a new portfolio manager
func NewPortfolioManager(multiplexer *exchanges.ExchangeMultiplexer) *PortfolioManager {
	return &PortfolioManager{
		multiplexer: multiplexer,
		positions:   make(map[string][]exchanges.Position),
		totalPnL:    decimal.Zero,
	}
}

// UpdatePositions refreshes positions from all exchanges
func (pm *PortfolioManager) UpdatePositions(ctx context.Context) error {
	positions, err := pm.multiplexer.GetPositions(ctx)
	if err != nil {
		return err
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Group positions by symbol
	pm.positions = make(map[string][]exchanges.Position)
	for _, pos := range positions {
		pm.positions[pos.Symbol] = append(pm.positions[pos.Symbol], pos)
	}

	// Calculate total PnL
	totalPnL := decimal.Zero
	for _, posList := range pm.positions {
		for _, pos := range posList {
			totalPnL = totalPnL.Add(pos.UnrealizedPnL).Add(pos.RealizedPnL)
		}
	}
	pm.totalPnL = totalPnL

	return nil
}

// GetPositions returns positions for a specific symbol
func (pm *PortfolioManager) GetPositions(symbol string) []exchanges.Position {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	positions, exists := pm.positions[symbol]
	if !exists {
		return nil
	}

	// Return copy
	dst := make([]exchanges.Position, len(positions))
	copy(dst, positions)
	return dst
}

// GetAllPositions returns all positions
func (pm *PortfolioManager) GetAllPositions() map[string][]exchanges.Position {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	all := make(map[string][]exchanges.Position)
	for symbol, positions := range pm.positions {
		dst := make([]exchanges.Position, len(positions))
		copy(dst, positions)
		all[symbol] = dst
	}
	return all
}

// GetTotalPnL returns the total portfolio PnL
func (pm *PortfolioManager) GetTotalPnL() decimal.Decimal {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.totalPnL
}

// GetActiveSymbols returns symbols with active positions
func (pm *PortfolioManager) GetActiveSymbols() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var symbols []string
	for symbol := range pm.positions {
		symbols = append(symbols, symbol)
	}
	return symbols
}
