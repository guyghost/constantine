package exchanges

import (
	"github.com/guyghost/constantine/internal/logger"
	"github.com/shopspring/decimal"
)

// Helper function for creating decimals from floats
func NewDecimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// Helper function for creating decimals from integers
func NewDecimalFromInt(i int64) decimal.Decimal {
	return decimal.NewFromInt(i)
}

// ParseDecimalOrZero parses a decimal string and returns zero on error with logging
func ParseDecimalOrZero(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		logger.Warn("Failed to parse decimal value",
			"value", s,
			"error", err.Error())
		return decimal.Zero
	}
	return d
}

// ParseDecimalOrDefault parses a decimal string and returns default value on error with logging
func ParseDecimalOrDefault(s string, defaultValue decimal.Decimal) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		logger.Warn("Failed to parse decimal value, using default",
			"value", s,
			"default", defaultValue.String(),
			"error", err.Error())
		return defaultValue
	}
	return d
}
