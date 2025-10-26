package symbolmanager

import (
	"fmt"
	"sync"

	"github.com/guyghost/constantine/internal/config"
)

// SymbolConfig holds configuration for a trading symbol
type SymbolConfig struct {
	Symbol           string
	StrategyConfig   *config.Config
	RiskLimits       RiskLimits
	ExchangePriority []string // Order of priority for exchanges
	Enabled          bool
}

// RiskLimits defines risk parameters for a symbol
type RiskLimits struct {
	MaxPositionSize  float64 // Maximum position size as % of capital
	MaxDrawdown      float64 // Maximum drawdown allowed
	DailyLossLimit   float64 // Daily loss limit as % of capital
	CorrelationLimit float64 // Maximum correlation with other symbols
}

// SymbolManager manages active trading symbols and their configurations
type SymbolManager struct {
	mu            sync.RWMutex
	symbols       map[string]*SymbolConfig
	activeSymbols []string
}

// NewSymbolManager creates a new symbol manager
func NewSymbolManager() *SymbolManager {
	return &SymbolManager{
		symbols:       make(map[string]*SymbolConfig),
		activeSymbols: make([]string, 0),
	}
}

// AddSymbol adds a new symbol with its configuration
func (sm *SymbolManager) AddSymbol(symbol string, config SymbolConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.symbols[symbol]; exists {
		return fmt.Errorf("symbol %s already exists", symbol)
	}

	config.Symbol = symbol // Ensure symbol is set in config
	sm.symbols[symbol] = &config

	if config.Enabled {
		sm.activeSymbols = append(sm.activeSymbols, symbol)
	}

	return nil
}

// RemoveSymbol removes a symbol from management
func (sm *SymbolManager) RemoveSymbol(symbol string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.symbols[symbol]; !exists {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	delete(sm.symbols, symbol)

	// Remove from active symbols if present
	for i, active := range sm.activeSymbols {
		if active == symbol {
			sm.activeSymbols = append(sm.activeSymbols[:i], sm.activeSymbols[i+1:]...)
			break
		}
	}

	return nil
}

// GetActiveSymbols returns a copy of currently active symbols
func (sm *SymbolManager) GetActiveSymbols() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	active := make([]string, len(sm.activeSymbols))
	copy(active, sm.activeSymbols)
	return active
}

// GetSymbolConfig returns the configuration for a specific symbol
func (sm *SymbolManager) GetSymbolConfig(symbol string) (*SymbolConfig, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	config, exists := sm.symbols[symbol]
	if !exists {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	// Return a copy to prevent external modifications
	configCopy := *config
	return &configCopy, nil
}

// IsSymbolActive checks if a symbol is currently active for trading
func (sm *SymbolManager) IsSymbolActive(symbol string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, active := range sm.activeSymbols {
		if active == symbol {
			return true
		}
	}
	return false
}

// EnableSymbol enables trading for a symbol
func (sm *SymbolManager) EnableSymbol(symbol string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	config, exists := sm.symbols[symbol]
	if !exists {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	if config.Enabled {
		return nil // Already enabled
	}

	config.Enabled = true
	sm.activeSymbols = append(sm.activeSymbols, symbol)
	return nil
}

// DisableSymbol disables trading for a symbol
func (sm *SymbolManager) DisableSymbol(symbol string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	config, exists := sm.symbols[symbol]
	if !exists {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	if !config.Enabled {
		return nil // Already disabled
	}

	config.Enabled = false

	// Remove from active symbols
	for i, active := range sm.activeSymbols {
		if active == symbol {
			sm.activeSymbols = append(sm.activeSymbols[:i], sm.activeSymbols[i+1:]...)
			break
		}
	}

	return nil
}

// GetAllSymbols returns all configured symbols (active and inactive)
func (sm *SymbolManager) GetAllSymbols() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	symbols := make([]string, 0, len(sm.symbols))
	for symbol := range sm.symbols {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// UpdateSymbolConfig updates the configuration for an existing symbol
func (sm *SymbolManager) UpdateSymbolConfig(symbol string, config SymbolConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	existing, exists := sm.symbols[symbol]
	if !exists {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	config.Symbol = symbol // Ensure symbol is set
	wasEnabled := existing.Enabled
	*existing = config

	// Handle enable/disable state change
	if config.Enabled && !wasEnabled {
		sm.activeSymbols = append(sm.activeSymbols, symbol)
	} else if !config.Enabled && wasEnabled {
		for i, active := range sm.activeSymbols {
			if active == symbol {
				sm.activeSymbols = append(sm.activeSymbols[:i], sm.activeSymbols[i+1:]...)
				break
			}
		}
	}

	return nil
}
