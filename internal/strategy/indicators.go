package strategy

import (
	"math"

	"github.com/shopspring/decimal"
)

// EMA calculates the Exponential Moving Average
func EMA(prices []decimal.Decimal, period int) []decimal.Decimal {
	if period <= 0 || len(prices) < period {
		return []decimal.Decimal{}
	}

	result := make([]decimal.Decimal, len(prices))
	multiplier := decimal.NewFromFloat(2.0 / float64(period+1))

	// Calculate initial SMA
	sum := decimal.Zero
	for i := 0; i < period; i++ {
		sum = sum.Add(prices[i])
	}
	result[period-1] = sum.Div(decimal.NewFromInt(int64(period)))

	// Calculate EMA
	for i := period; i < len(prices); i++ {
		ema := prices[i].Sub(result[i-1]).Mul(multiplier).Add(result[i-1])
		result[i] = ema
	}

	return result[period-1:]
}

// SMA calculates the Simple Moving Average
func SMA(prices []decimal.Decimal, period int) []decimal.Decimal {
	if period <= 0 || len(prices) < period {
		return []decimal.Decimal{}
	}

	result := make([]decimal.Decimal, len(prices)-period+1)

	for i := 0; i <= len(prices)-period; i++ {
		sum := decimal.Zero
		for j := 0; j < period; j++ {
			sum = sum.Add(prices[i+j])
		}
		result[i] = sum.Div(decimal.NewFromInt(int64(period)))
	}

	return result
}

// RSI calculates the Relative Strength Index
func RSI(prices []decimal.Decimal, period int) []decimal.Decimal {
	if period <= 0 || len(prices) < period+1 {
		return []decimal.Decimal{}
	}

	gains := make([]decimal.Decimal, len(prices)-1)
	losses := make([]decimal.Decimal, len(prices)-1)

	// Calculate gains and losses
	for i := 1; i < len(prices); i++ {
		change := prices[i].Sub(prices[i-1])
		if change.GreaterThan(decimal.Zero) {
			gains[i-1] = change
			losses[i-1] = decimal.Zero
		} else {
			gains[i-1] = decimal.Zero
			losses[i-1] = change.Abs()
		}
	}

	gainEMA := EMA(gains, period)
	lossEMA := EMA(losses, period)

	length := len(gainEMA)
	if len(lossEMA) < length {
		length = len(lossEMA)
	}
	if length == 0 {
		return []decimal.Decimal{}
	}

	result := make([]decimal.Decimal, length)
	for i := 0; i < length; i++ {
		loss := lossEMA[i]
		if loss.IsZero() {
			result[i] = decimal.NewFromInt(100)
			continue
		}
		rs := gainEMA[i].Div(loss)
		rsi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))
		result[i] = rsi
	}

	return result
}

// MACD calculates the Moving Average Convergence Divergence
func MACD(prices []decimal.Decimal, fastPeriod, slowPeriod, signalPeriod int) (macd, signal, histogram []decimal.Decimal) {
	if fastPeriod <= 0 || slowPeriod <= 0 || signalPeriod <= 0 || len(prices) < slowPeriod {
		return []decimal.Decimal{}, []decimal.Decimal{}, []decimal.Decimal{}
	}

	// Calculate fast and slow EMAs
	fastEMA := EMA(prices, fastPeriod)
	slowEMA := EMA(prices, slowPeriod)

	// Align the EMAs
	offset := len(fastEMA) - len(slowEMA)
	if offset < 0 {
		offset = 0
	}

	// Calculate MACD line
	macdLine := make([]decimal.Decimal, len(slowEMA))
	for i := 0; i < len(slowEMA); i++ {
		macdLine[i] = fastEMA[i+offset].Sub(slowEMA[i])
	}

	// Calculate signal line
	signalLine := EMA(macdLine, signalPeriod)

	// Calculate histogram
	hist := make([]decimal.Decimal, len(signalLine))
	offset = len(macdLine) - len(signalLine)
	for i := 0; i < len(signalLine); i++ {
		hist[i] = macdLine[i+offset].Sub(signalLine[i])
	}

	return macdLine, signalLine, hist
}

// BollingerBands calculates Bollinger Bands
func BollingerBands(prices []decimal.Decimal, period int, stdDev float64) (upper, middle, lower []decimal.Decimal) {
	if period <= 0 || len(prices) < period {
		return []decimal.Decimal{}, []decimal.Decimal{}, []decimal.Decimal{}
	}

	middle = SMA(prices, period)
	upper = make([]decimal.Decimal, len(middle))
	lower = make([]decimal.Decimal, len(middle))

	for i := 0; i < len(middle); i++ {
		// Calculate standard deviation
		sum := 0.0
		for j := 0; j < period; j++ {
			diff, _ := prices[i+j].Sub(middle[i]).Float64()
			sum += diff * diff
		}
		std := math.Sqrt(sum / float64(period))
		stdDecimal := decimal.NewFromFloat(std * stdDev)

		upper[i] = middle[i].Add(stdDecimal)
		lower[i] = middle[i].Sub(stdDecimal)
	}

	return upper, middle, lower
}

// ATR calculates the Average True Range
func ATR(high, low, close []decimal.Decimal, period int) []decimal.Decimal {
	if period <= 0 || len(high) < period+1 || len(low) < period+1 || len(close) < period+1 {
		return []decimal.Decimal{}
	}

	trueRanges := make([]decimal.Decimal, len(high)-1)

	for i := 1; i < len(high); i++ {
		hl := high[i].Sub(low[i])
		hc := high[i].Sub(close[i-1]).Abs()
		lc := low[i].Sub(close[i-1]).Abs()

		tr := hl
		if hc.GreaterThan(tr) {
			tr = hc
		}
		if lc.GreaterThan(tr) {
			tr = lc
		}

		trueRanges[i-1] = tr
	}

	return SMA(trueRanges, period)
}

// VWAP calculates the Volume Weighted Average Price
func VWAP(prices, volumes []decimal.Decimal) decimal.Decimal {
	if len(prices) == 0 || len(volumes) == 0 || len(prices) != len(volumes) {
		return decimal.Zero
	}

	totalPV := decimal.Zero
	totalVolume := decimal.Zero

	for i := 0; i < len(prices); i++ {
		totalPV = totalPV.Add(prices[i].Mul(volumes[i]))
		totalVolume = totalVolume.Add(volumes[i])
	}

	if totalVolume.IsZero() {
		return decimal.Zero
	}

	return totalPV.Div(totalVolume)
}

// Stochastic calculates the Stochastic Oscillator
func Stochastic(high, low, close []decimal.Decimal, period int) []decimal.Decimal {
	if period <= 0 || len(high) < period || len(low) < period || len(close) < period {
		return []decimal.Decimal{}
	}

	result := make([]decimal.Decimal, len(close)-period+1)

	for i := 0; i <= len(close)-period; i++ {
		highest := high[i]
		lowest := low[i]

		for j := 1; j < period; j++ {
			if high[i+j].GreaterThan(highest) {
				highest = high[i+j]
			}
			if low[i+j].LessThan(lowest) {
				lowest = low[i+j]
			}
		}

		currentClose := close[i+period-1]
		if highest.Equal(lowest) {
			result[i] = decimal.NewFromInt(50)
		} else {
			k := currentClose.Sub(lowest).Div(highest.Sub(lowest)).Mul(decimal.NewFromInt(100))
			result[i] = k
		}
	}

	return result
}
