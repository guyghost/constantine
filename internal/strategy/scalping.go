package strategy

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
)

const strategyAPITimeout = 5 * time.Second

func parseIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}
	return defaultValue
}

func parseFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.ParseFloat(value, 64); err == nil {
		return parsed
	}
	return defaultValue
}

// Config holds strategy configuration
type Config struct {
	Symbol            string
	ShortEMAPeriod    int
	LongEMAPeriod     int
	RSIPeriod         int
	RSIOversold       float64
	RSIOverbought     float64
	TakeProfitPercent float64
	StopLossPercent   float64
	MaxPositionSize   decimal.Decimal
	MinPriceMove      decimal.Decimal
	UpdateInterval    time.Duration
}

// DefaultConfig returns default strategy configuration
func DefaultConfig() *Config {
	cfg := &Config{
		Symbol:            "BTC-USD",
		ShortEMAPeriod:    9,
		LongEMAPeriod:     21,
		RSIPeriod:         14,
		RSIOversold:       30.0,
		RSIOverbought:     70.0,
		TakeProfitPercent: 0.5,
		StopLossPercent:   0.25,
		MaxPositionSize:   decimal.NewFromFloat(0.1),
		MinPriceMove:      decimal.NewFromFloat(0.01),
		UpdateInterval:    1 * time.Second,
	}

	if symbol := os.Getenv("STRATEGY_SYMBOL"); symbol != "" {
		cfg.Symbol = symbol
	}
	if val := parseIntEnv("STRATEGY_SHORT_EMA", cfg.ShortEMAPeriod); val > 0 {
		cfg.ShortEMAPeriod = val
	}
	if val := parseIntEnv("STRATEGY_LONG_EMA", cfg.LongEMAPeriod); val > 0 {
		cfg.LongEMAPeriod = val
	}
	if val := parseIntEnv("STRATEGY_RSI_PERIOD", cfg.RSIPeriod); val > 0 {
		cfg.RSIPeriod = val
	}
	if val := parseFloatEnv("STRATEGY_RSI_OVERSOLD", cfg.RSIOversold); val > 0 {
		cfg.RSIOversold = val
	}
	if val := parseFloatEnv("STRATEGY_RSI_OVERBOUGHT", cfg.RSIOverbought); val > 0 {
		cfg.RSIOverbought = val
	}
	if val := parseFloatEnv("STRATEGY_TAKE_PROFIT", cfg.TakeProfitPercent); val > 0 {
		cfg.TakeProfitPercent = val
	}
	if val := parseFloatEnv("STRATEGY_STOP_LOSS", cfg.StopLossPercent); val > 0 {
		cfg.StopLossPercent = val
	}
	if value := os.Getenv("STRATEGY_MAX_POSITION_SIZE"); value != "" {
		if parsed, err := decimal.NewFromString(value); err == nil {
			cfg.MaxPositionSize = parsed
		}
	}
	if value := os.Getenv("STRATEGY_MIN_PRICE_MOVE"); value != "" {
		if parsed, err := decimal.NewFromString(value); err == nil {
			cfg.MinPriceMove = parsed
		}
	}
	if duration := os.Getenv("STRATEGY_UPDATE_INTERVAL"); duration != "" {
		if parsed, err := time.ParseDuration(duration); err == nil {
			cfg.UpdateInterval = parsed
		}
	}

	return cfg
}

// ScalpingStrategy implements a scalping trading strategy
type ScalpingStrategy struct {
	config          *Config
	exchange        exchanges.Exchange
	signalGenerator *SignalGenerator
	mu              sync.RWMutex

	// Market data
	prices     []decimal.Decimal
	volumes    []decimal.Decimal
	orderbook  *exchanges.OrderBook
	lastSignal *Signal

	// Callbacks
	onSignal   func(*Signal)
	onError    func(error)
	onPosition func(*exchanges.Position)

	// Control
	running bool
	done    chan struct{}
	cancel  context.CancelFunc
}

// NewScalpingStrategy creates a new scalping strategy
func NewScalpingStrategy(config *Config, exchange exchanges.Exchange) *ScalpingStrategy {
	return &ScalpingStrategy{
		config:          config,
		exchange:        exchange,
		signalGenerator: NewSignalGenerator(config),
		prices:          make([]decimal.Decimal, 0, 100),
		volumes:         make([]decimal.Decimal, 0, 100),
		done:            make(chan struct{}),
	}
}

// SetSignalCallback sets the callback for signals
func (s *ScalpingStrategy) SetSignalCallback(callback func(*Signal)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSignal = callback
}

// SetErrorCallback sets the callback for errors
func (s *ScalpingStrategy) SetErrorCallback(callback func(error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onError = callback
}

// SetPositionCallback sets the callback for position updates
func (s *ScalpingStrategy) SetPositionCallback(callback func(*exchanges.Position)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onPosition = callback
}

// Start starts the scalping strategy
func (s *ScalpingStrategy) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("strategy already running")
	}

	if s.done == nil {
		s.done = make(chan struct{})
	} else {
		select {
		case <-s.done:
			s.done = make(chan struct{})
		default:
		}
	}
	doneCh := s.done
	strategyCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()

	// Subscribe to market data
	if err := s.subscribeMarketData(strategyCtx); err != nil {
		cancel()
		s.mu.Lock()
		s.cancel = nil
		s.mu.Unlock()
		return fmt.Errorf("failed to subscribe to market data: %w", err)
	}

	s.mu.Lock()
	// Another goroutine could have stopped the strategy while we subscribed
	if s.running {
		s.cancel = nil
		s.mu.Unlock()
		cancel()
		return fmt.Errorf("strategy already running")
	}
	s.running = true
	s.mu.Unlock()

	// Start strategy loop
	go s.run(strategyCtx, doneCh)

	return nil
}

// Stop stops the scalping strategy
func (s *ScalpingStrategy) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if s.done != nil {
		select {
		case <-s.done:
		default:
			close(s.done)
		}
		s.done = nil
	}
	s.running = false
	return nil
}

// IsRunning returns whether the strategy is running
func (s *ScalpingStrategy) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetSignalGenerator returns the signal generator for backtesting
func (s *ScalpingStrategy) GetSignalGenerator() *SignalGenerator {
	return s.signalGenerator
}

// subscribeMarketData subscribes to market data streams
func (s *ScalpingStrategy) subscribeMarketData(ctx context.Context) error {
	// Subscribe to ticker
	tickerCtx, cancel := context.WithTimeout(ctx, strategyAPITimeout)
	if err := s.exchange.SubscribeTicker(tickerCtx, s.config.Symbol, s.handleTicker); err != nil {
		cancel()
		return err
	}
	cancel()

	// Subscribe to order book
	orderBookCtx, cancel := context.WithTimeout(ctx, strategyAPITimeout)
	if err := s.exchange.SubscribeOrderBook(orderBookCtx, s.config.Symbol, s.handleOrderBook); err != nil {
		cancel()
		return err
	}
	cancel()

	// Subscribe to trades
	tradesCtx, cancel := context.WithTimeout(ctx, strategyAPITimeout)
	if err := s.exchange.SubscribeTrades(tradesCtx, s.config.Symbol, s.handleTrade); err != nil {
		cancel()
		return err
	}
	cancel()

	return nil
}

// handleTicker handles ticker updates
func (s *ScalpingStrategy) handleTicker(ticker *exchanges.Ticker) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update price history
	s.prices = append(s.prices, ticker.Last)

	// Keep only last 100 prices
	if len(s.prices) > 100 {
		s.prices = s.prices[1:]
	}
}

// handleOrderBook handles order book updates
func (s *ScalpingStrategy) handleOrderBook(orderbook *exchanges.OrderBook) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orderbook = orderbook
}

// handleTrade handles trade updates
func (s *ScalpingStrategy) handleTrade(trade *exchanges.Trade) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update volume history
	s.volumes = append(s.volumes, trade.Amount)

	// Keep only last 100 volumes
	if len(s.volumes) > 100 {
		s.volumes = s.volumes[1:]
	}
}

// run is the main strategy loop
func (s *ScalpingStrategy) run(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			s.update(ctx)
		}
	}
}

// update performs strategy analysis and generates signals
func (s *ScalpingStrategy) update(ctx context.Context) {
	s.mu.RLock()
	prices := make([]decimal.Decimal, len(s.prices))
	copy(prices, s.prices)
	volumes := make([]decimal.Decimal, len(s.volumes))
	copy(volumes, s.volumes)
	orderbook := s.orderbook
	s.mu.RUnlock()

	// Need enough data for analysis
	if len(prices) < s.config.LongEMAPeriod {
		return
	}

	// Generate signal
	signal := s.signalGenerator.GenerateSignal(
		s.config.Symbol,
		prices,
		volumes,
		orderbook,
	)

	// Skip if no signal
	if signal.Type == SignalTypeNone {
		return
	}

	// Check if we should emit this signal
	s.mu.Lock()
	shouldEmit := s.lastSignal == nil || signal.Type != s.lastSignal.Type || signal.Side != s.lastSignal.Side
	if shouldEmit {
		s.lastSignal = signal
	}
	callback := s.onSignal
	s.mu.Unlock()

	// Emit signal
	if shouldEmit && callback != nil {
		safeInvoke(func() { callback(signal) })
	}

	// Check exit conditions for existing positions
	s.checkExitConditions(ctx, prices)
}

// checkExitConditions checks if any positions should be exited
func (s *ScalpingStrategy) checkExitConditions(ctx context.Context, prices []decimal.Decimal) {
	callCtx, cancel := context.WithTimeout(ctx, strategyAPITimeout)
	defer cancel()

	positions, err := s.exchange.GetPositions(callCtx)
	if err != nil {
		s.emitError(fmt.Errorf("failed to get positions: %w", err))
		return
	}

	if len(prices) < s.config.RSIPeriod {
		return
	}

	rsi := RSI(prices, s.config.RSIPeriod)
	if len(rsi) == 0 {
		return
	}
	currentRSI := rsi[len(rsi)-1]
	currentPrice := prices[len(prices)-1]

	for _, position := range positions {
		if position.Symbol != s.config.Symbol {
			continue
		}

		if s.signalGenerator.ShouldExit(&position, currentPrice, currentRSI) {
			// Generate exit signal
			signal := &Signal{
				Type:     SignalTypeExit,
				Side:     position.Side,
				Symbol:   position.Symbol,
				Price:    currentPrice,
				Strength: 1.0,
				Reason:   "Stop loss or take profit triggered",
			}

			s.mu.RLock()
			callback := s.onSignal
			s.mu.RUnlock()

			if callback != nil {
				safeInvoke(func() { callback(signal) })
			}
		}
	}
}

// emitError emits an error through the error callback
func (s *ScalpingStrategy) emitError(err error) {
	s.mu.RLock()
	callback := s.onError
	s.mu.RUnlock()

	if callback != nil {
		safeInvoke(func() { callback(err) })
	}
}

// GetLastSignal returns the last generated signal
func (s *ScalpingStrategy) GetLastSignal() *Signal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSignal
}

// GetCurrentPrices returns the current price history
func (s *ScalpingStrategy) GetCurrentPrices() []decimal.Decimal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prices := make([]decimal.Decimal, len(s.prices))
	copy(prices, s.prices)
	return prices
}

// GetOrderBook returns the current order book
func (s *ScalpingStrategy) GetOrderBook() *exchanges.OrderBook {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orderbook
}

func safeInvoke(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			telemetry.RecordCallbackPanic()
		}
	}()
	fn()
}
