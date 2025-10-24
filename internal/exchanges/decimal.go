package exchanges

import (
	"log"

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
		log.Printf("WARNING: Failed to parse decimal value '%s': %v", s, err)
		return decimal.Zero
	}
	return d
}

// ParseDecimalOrDefault parses a decimal string and returns default value on error with logging
func ParseDecimalOrDefault(s string, defaultValue decimal.Decimal) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		log.Printf("WARNING: Failed to parse decimal value '%s': %v, using default", s, err)
		return defaultValue
	}
	return d
}
