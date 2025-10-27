package strategy

import (
	"context"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"github.com/guyghost/constantine/internal/config"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/logger"
)

// IntegratedStrategyEngine combines dynamic weights, symbol selection, and signal generation
type IntegratedStrategyEngine struct {
	config                 *config.Config
	tradingSymbols         []string
	symbolSelector         *SymbolSelector
	weightCalculator       *WeightCalculator
	signalGenerator        *SignalGenerator
	scalingStrategy        *ScalpingStrategy
	exchange               exchanges.Exchange
	
	// State
	selectedSymbols        map[string]RankedSymbol
	marketData             map[string]SymbolData
	refreshInterval        time.Duration
	
	// Control
	mu                     sync.RWMutex
	running                bool
	done                   chan struct{}
	cancel                 context.CancelFunc
}

// NewIntegratedStrategyEngine creates a new integrated strategy engine
func NewIntegratedStrategyEngine(
	cfg *config.Config,
	tradingSymbols []string,
	exchange exchanges.Exchange,
	refreshInterval time.Duration,
) *IntegratedStrategyEngine {
	return &IntegratedStrategyEngine{
		config:            cfg,
		tradingSymbols:    tradingSymbols,
		symbolSelector:    NewSymbolSelector(cfg),
		weightCalculator:  NewWeightCalculator(cfg),
		signalGenerator:   NewSignalGenerator(cfg),
		scalingStrategy:   NewScalpingStrategy(cfg, exchange),
		exchange:          exchange,
		selectedSymbols:   make(map[string]RankedSymbol),
		marketData:        make(map[string]SymbolData),
		refreshInterval:   refreshInterval,
		done:              make(chan struct{}),
	}
}

// Start initializes and starts the integrated engine
func (ise *IntegratedStrategyEngine) Start(ctx context.Context) error {
	ise.mu.Lock()
	if ise.running {
		ise.mu.Unlock()
		return nil
	}
	
	engCtx, cancel := context.WithCancel(ctx)
	ise.cancel = cancel
	doneCh := ise.done
	
	ise.running = true
	ise.mu.Unlock()

	// Start strategy
	if err := ise.scalingStrategy.Start(engCtx); err != nil {
		ise.mu.Lock()
		ise.running = false
		ise.mu.Unlock()
		return err
	}

	// Start symbol selection refresh loop
	go ise.refreshSymbolSelection(engCtx, doneCh)

	logger.Component("strategy").Info("integrated strategy engine started")
	return nil
}

// Stop stops the integrated engine
func (ise *IntegratedStrategyEngine) Stop() error {
	ise.mu.Lock()
	if !ise.running {
		ise.mu.Unlock()
		return nil
	}

	ise.running = false
	if ise.cancel != nil {
		ise.cancel()
	}
	
	select {
	case <-ise.done:
	default:
		close(ise.done)
	}
	ise.mu.Unlock()

	return ise.scalingStrategy.Stop()
}

// refreshSymbolSelection periodically refreshes symbol selection
func (ise *IntegratedStrategyEngine) refreshSymbolSelection(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(ise.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			ise.updateSymbolSelection(ctx)
		}
	}
}

// updateSymbolSelection updates the selected trading symbols
func (ise *IntegratedStrategyEngine) updateSymbolSelection(ctx context.Context) {
	// Get list of symbols to evaluate
	symbols := ise.tradingSymbols
	if len(symbols) == 0 {
		logger.Component("strategy").Warn("no trading symbols configured")
		return
	}

	// Fetch market data for all symbols
	symbolData := make(map[string]SymbolData)
	for _, symbol := range symbols {
		prices, err := ise.fetchPriceData(ctx, symbol, 30)
		if err != nil {
			logger.Component("strategy").Warn("failed to fetch price data", "symbol", symbol, "error", err)
			continue
		}

		volumes, err := ise.fetchVolumeData(ctx, symbol, 30)
		if err != nil {
			logger.Component("strategy").Warn("failed to fetch volume data", "symbol", symbol, "error", err)
			volumes = make([]decimal.Decimal, len(prices))
			for i := range volumes {
				volumes[i] = decimal.NewFromInt(1000) // Default volume
			}
		}

		if len(prices) >= 20 { // Minimum data requirement
			symbolData[symbol] = SymbolData{
				Prices:  prices,
				Volumes: volumes,
			}
		}
	}

	if len(symbolData) == 0 {
		logger.Component("strategy").Warn("no valid symbol data available")
		return
	}

	// Select best symbols
	selectedCount := 1
	if len(symbols) > 1 {
		selectedCount = (len(symbols) + 1) / 2 // Select 50% of available symbols
		if selectedCount < 1 {
			selectedCount = 1
		}
	}

	selected := ise.symbolSelector.SelectBestSymbols(symbols, symbolData, selectedCount)

	// Update state
	ise.mu.Lock()
	ise.selectedSymbols = make(map[string]RankedSymbol)
	for _, rs := range selected {
		ise.selectedSymbols[rs.Symbol] = rs
	}
	ise.marketData = symbolData
	ise.mu.Unlock()

	// Log selection
	logger.Component("strategy").Info("symbol selection updated",
		"total_symbols", len(symbols),
		"selected_count", len(selected),
		"symbols", formatSelectedSymbols(selected))
}

// fetchPriceData fetches recent price data for a symbol
func (ise *IntegratedStrategyEngine) fetchPriceData(ctx context.Context, symbol string, count int) ([]decimal.Decimal, error) {
	candles, err := ise.exchange.GetCandles(ctx, symbol, "1m", count)
	if err != nil {
		return nil, err
	}

	prices := make([]decimal.Decimal, len(candles))
	for i, candle := range candles {
		prices[i] = candle.Close
	}
	return prices, nil
}

// fetchVolumeData fetches recent volume data for a symbol
func (ise *IntegratedStrategyEngine) fetchVolumeData(ctx context.Context, symbol string, count int) ([]decimal.Decimal, error) {
	candles, err := ise.exchange.GetCandles(ctx, symbol, "1m", count)
	if err != nil {
		return nil, err
	}

	volumes := make([]decimal.Decimal, len(candles))
	for i, candle := range candles {
		volumes[i] = candle.Volume
	}
	return volumes, nil
}

// GetSelectedSymbols returns the currently selected trading symbols
func (ise *IntegratedStrategyEngine) GetSelectedSymbols() map[string]RankedSymbol {
	ise.mu.RLock()
	defer ise.mu.RUnlock()

	result := make(map[string]RankedSymbol)
	for k, v := range ise.selectedSymbols {
		result[k] = v
	}
	return result
}

// GetSignalGenerator returns the signal generator for custom usage
func (ise *IntegratedStrategyEngine) GetSignalGenerator() *SignalGenerator {
	return ise.signalGenerator
}

// GetWeightCalculator returns the weight calculator for custom usage
func (ise *IntegratedStrategyEngine) GetWeightCalculator() *WeightCalculator {
	return ise.weightCalculator
}

// GetScalpingStrategy returns the underlying scalping strategy
func (ise *IntegratedStrategyEngine) GetScalpingStrategy() *ScalpingStrategy {
	return ise.scalingStrategy
}

// SetSignalCallback sets the callback for new signals
func (ise *IntegratedStrategyEngine) SetSignalCallback(callback func(*Signal)) {
	ise.scalingStrategy.SetSignalCallback(callback)
}

// SetErrorCallback sets the callback for errors
func (ise *IntegratedStrategyEngine) SetErrorCallback(callback func(error)) {
	ise.scalingStrategy.SetErrorCallback(callback)
}

// formatSelectedSymbols formats selected symbols for logging
func formatSelectedSymbols(selected []RankedSymbol) string {
	result := ""
	for i, rs := range selected {
		if i > 0 {
			result += ", "
		}
		result += rs.Symbol
	}
	return result
}
