package utils

import (
	"math"

	"github.com/shopspring/decimal"
)

// RoundDecimal rounds a decimal to a specific number of decimal places
func RoundDecimal(d decimal.Decimal, places int32) decimal.Decimal {
	return d.Round(places)
}

// MinDecimal returns the minimum of two decimals
func MinDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

// MaxDecimal returns the maximum of two decimals
func MaxDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}

// AbsDecimal returns the absolute value of a decimal
func AbsDecimal(d decimal.Decimal) decimal.Decimal {
	return d.Abs()
}

// PercentChange calculates the percentage change between two values
func PercentChange(oldValue, newValue decimal.Decimal) decimal.Decimal {
	if oldValue.IsZero() {
		return decimal.Zero
	}
	return newValue.Sub(oldValue).Div(oldValue).Mul(decimal.NewFromInt(100))
}

// StandardDeviation calculates the standard deviation of a slice of decimals
func StandardDeviation(values []decimal.Decimal) decimal.Decimal {
	if len(values) == 0 {
		return decimal.Zero
	}

	// Calculate mean
	sum := decimal.Zero
	for _, v := range values {
		sum = sum.Add(v)
	}
	mean := sum.Div(decimal.NewFromInt(int64(len(values))))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		diff, _ := v.Sub(mean).Float64()
		variance += diff * diff
	}
	variance /= float64(len(values))

	return decimal.NewFromFloat(math.Sqrt(variance))
}

// ClampDecimal clamps a decimal value between min and max
func ClampDecimal(value, min, max decimal.Decimal) decimal.Decimal {
	if value.LessThan(min) {
		return min
	}
	if value.GreaterThan(max) {
		return max
	}
	return value
}

// LerpDecimal performs linear interpolation between two decimals
func LerpDecimal(a, b decimal.Decimal, t float64) decimal.Decimal {
	tDecimal := decimal.NewFromFloat(t)
	return a.Add(b.Sub(a).Mul(tDecimal))
}

// IsWithinRange checks if a value is within a range (inclusive)
func IsWithinRange(value, min, max decimal.Decimal) bool {
	return value.GreaterThanOrEqual(min) && value.LessThanOrEqual(max)
}
