package strategy

import (
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// Signal represents a trading signal
type Signal struct {
	Type      SignalType
	Side      exchanges.OrderSide
	Symbol    string
	Price     decimal.Decimal
	Strength  float64 // 0.0 to 1.0
	Reason    string
	Timestamp int64
}

// SignalType represents the type of signal
type SignalType string

const (
	SignalTypeEntry  SignalType = "entry"
	SignalTypeExit   SignalType = "exit"
	SignalTypeNone   SignalType = "none"
	SignalTypeStrong SignalType = "strong"
	SignalTypeWeak   SignalType = "weak"
)

// SignalGenerator generates trading signals
type SignalGenerator struct {
	config *Config
}

// NewSignalGenerator creates a new signal generator
func NewSignalGenerator(config *Config) *SignalGenerator {
	return &SignalGenerator{
		config: config,
	}
}

// GenerateSignal generates a trading signal based on market data and indicators
func (sg *SignalGenerator) GenerateSignal(
	symbol string,
	prices []decimal.Decimal,
	volumes []decimal.Decimal,
	orderbook *exchanges.OrderBook,
) *Signal {
	if len(prices) < sg.config.LongEMAPeriod {
		return &Signal{Type: SignalTypeNone}
	}

	// Calculate indicators
	shortEMA := EMA(prices, sg.config.ShortEMAPeriod)
	longEMA := EMA(prices, sg.config.LongEMAPeriod)
	rsi := RSI(prices, sg.config.RSIPeriod)

	if len(shortEMA) == 0 || len(longEMA) == 0 || len(rsi) == 0 {
		return &Signal{Type: SignalTypeNone}
	}

	currentShortEMA := shortEMA[len(shortEMA)-1]
	currentLongEMA := longEMA[len(longEMA)-1]
	currentRSI := rsi[len(rsi)-1]
	currentPrice := prices[len(prices)-1]

	// Check for buy signal
	if sg.isBuySignal(currentShortEMA, currentLongEMA, currentRSI, orderbook) {
		strength := sg.calculateSignalStrength(currentShortEMA, currentLongEMA, currentRSI, true)
		return &Signal{
			Type:     SignalTypeEntry,
			Side:     exchanges.OrderSideBuy,
			Symbol:   symbol,
			Price:    currentPrice,
			Strength: strength,
			Reason:   "EMA crossover + RSI oversold",
		}
	}

	// Check for sell signal
	if sg.isSellSignal(currentShortEMA, currentLongEMA, currentRSI, orderbook) {
		strength := sg.calculateSignalStrength(currentShortEMA, currentLongEMA, currentRSI, false)
		return &Signal{
			Type:     SignalTypeEntry,
			Side:     exchanges.OrderSideSell,
			Symbol:   symbol,
			Price:    currentPrice,
			Strength: strength,
			Reason:   "EMA crossover + RSI overbought",
		}
	}

	return &Signal{Type: SignalTypeNone}
}

// isBuySignal checks if current conditions indicate a buy signal
func (sg *SignalGenerator) isBuySignal(
	shortEMA, longEMA, rsi decimal.Decimal,
	orderbook *exchanges.OrderBook,
) bool {
	// EMA crossover: short EMA crosses above long EMA
	emaCrossover := shortEMA.GreaterThan(longEMA)

	// RSI indicates oversold
	rsiOversold := rsi.LessThan(decimal.NewFromFloat(sg.config.RSIOversold))

	// Check order book imbalance (more bids than asks)
	orderbookImbalance := sg.checkOrderbookImbalance(orderbook, true)

	return emaCrossover && (rsiOversold || orderbookImbalance)
}

// isSellSignal checks if current conditions indicate a sell signal
func (sg *SignalGenerator) isSellSignal(
	shortEMA, longEMA, rsi decimal.Decimal,
	orderbook *exchanges.OrderBook,
) bool {
	// EMA crossover: short EMA crosses below long EMA
	emaCrossover := shortEMA.LessThan(longEMA)

	// RSI indicates overbought
	rsiOverbought := rsi.GreaterThan(decimal.NewFromFloat(sg.config.RSIOverbought))

	// Check order book imbalance (more asks than bids)
	orderbookImbalance := sg.checkOrderbookImbalance(orderbook, false)

	return emaCrossover && (rsiOverbought || orderbookImbalance)
}

// checkOrderbookImbalance checks for order book imbalance
func (sg *SignalGenerator) checkOrderbookImbalance(orderbook *exchanges.OrderBook, buyDirection bool) bool {
	if orderbook == nil || len(orderbook.Bids) == 0 || len(orderbook.Asks) == 0 {
		return false
	}

	// Calculate total bid and ask volumes in top levels
	bidVolume := decimal.Zero
	askVolume := decimal.Zero

	depth := 5
	if len(orderbook.Bids) < depth {
		depth = len(orderbook.Bids)
	}

	for i := 0; i < depth; i++ {
		bidVolume = bidVolume.Add(orderbook.Bids[i].Amount)
	}

	if len(orderbook.Asks) < depth {
		depth = len(orderbook.Asks)
	}

	for i := 0; i < depth; i++ {
		askVolume = askVolume.Add(orderbook.Asks[i].Amount)
	}

	if askVolume.IsZero() || bidVolume.IsZero() {
		return false
	}

	// Calculate imbalance ratio
	if buyDirection {
		ratio := bidVolume.Div(askVolume)
		return ratio.GreaterThan(decimal.NewFromFloat(1.5)) // 50% more bids than asks
	}

	ratio := askVolume.Div(bidVolume)
	return ratio.GreaterThan(decimal.NewFromFloat(1.5)) // 50% more asks than bids
}

// calculateSignalStrength calculates the strength of a signal (0.0 to 1.0)
func (sg *SignalGenerator) calculateSignalStrength(
	shortEMA, longEMA, rsi decimal.Decimal,
	isBuy bool,
) float64 {
	strength := 0.0

	// EMA divergence strength (max 0.4)
	emaDiff := shortEMA.Sub(longEMA).Abs()
	emaStrength := 0.0
	if !longEMA.IsZero() {
		emaDivergence := emaDiff.Div(longEMA)
		emaStrength, _ = emaDivergence.Mul(decimal.NewFromInt(100)).Float64()
	}
	if emaStrength > 0.4 {
		emaStrength = 0.4
	}
	strength += emaStrength

	// RSI strength (max 0.6)
	rsiFloat, _ := rsi.Float64()
	var rsiStrength float64
	if isBuy {
		// For buy: the lower the RSI, the stronger the signal
		rsiStrength = (30.0 - rsiFloat) / 30.0 * 0.6
	} else {
		// For sell: the higher the RSI, the stronger the signal
		rsiStrength = (rsiFloat - 70.0) / 30.0 * 0.6
	}

	if rsiStrength < 0 {
		rsiStrength = 0
	}
	if rsiStrength > 0.6 {
		rsiStrength = 0.6
	}
	strength += rsiStrength

	if strength > 1.0 {
		strength = 1.0
	}
	if strength < 0.0 {
		strength = 0.0
	}

	return strength
}

// ShouldExit determines if a position should be exited
func (sg *SignalGenerator) ShouldExit(
	position *exchanges.Position,
	currentPrice decimal.Decimal,
	rsi decimal.Decimal,
) bool {
	if position == nil {
		return false
	}

	// Calculate PnL percentage
	pnlPercent := currentPrice.Sub(position.EntryPrice).Div(position.EntryPrice).Mul(decimal.NewFromInt(100))

	// Take profit
	if position.Side == exchanges.OrderSideBuy {
		if pnlPercent.GreaterThanOrEqual(decimal.NewFromFloat(sg.config.TakeProfitPercent)) {
			return true
		}
		// Stop loss
		if pnlPercent.LessThanOrEqual(decimal.NewFromFloat(-sg.config.StopLossPercent)) {
			return true
		}
		// Exit if RSI is overbought
		if rsi.GreaterThan(decimal.NewFromFloat(sg.config.RSIOverbought)) {
			return true
		}
	} else {
		if pnlPercent.LessThanOrEqual(decimal.NewFromFloat(-sg.config.TakeProfitPercent)) {
			return true
		}
		// Stop loss
		if pnlPercent.GreaterThanOrEqual(decimal.NewFromFloat(sg.config.StopLossPercent)) {
			return true
		}
		// Exit if RSI is oversold
		if rsi.LessThan(decimal.NewFromFloat(sg.config.RSIOversold)) {
			return true
		}
	}

	return false
}
