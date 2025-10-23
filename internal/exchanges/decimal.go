package exchanges

import "github.com/shopspring/decimal"

// Helper function for creating decimals from floats
func NewDecimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// Helper function for creating decimals from integers
func NewDecimalFromInt(i int64) decimal.Decimal {
	return decimal.NewFromInt(i)
}
