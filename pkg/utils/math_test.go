package utils

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestRoundDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    decimal.Decimal
		places   int32
		expected decimal.Decimal
	}{
		{"Round to 2 places", decimal.NewFromFloat(1.23456), 2, decimal.NewFromFloat(1.23)},
		{"Round to 0 places", decimal.NewFromFloat(1.6), 0, decimal.NewFromFloat(2)},
		{"Round to 4 places", decimal.NewFromFloat(1.23456), 4, decimal.NewFromFloat(1.2346)},
		{"No rounding needed", decimal.NewFromFloat(1.23), 2, decimal.NewFromFloat(1.23)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundDecimal(tt.input, tt.places)
			if !result.Equal(tt.expected) {
				t.Errorf("RoundDecimal(%v, %d) = %v, want %v", tt.input, tt.places, result, tt.expected)
			}
		})
	}
}

func TestMinDecimal(t *testing.T) {
	tests := []struct {
		name     string
		a, b     decimal.Decimal
		expected decimal.Decimal
	}{
		{"a < b", decimal.NewFromFloat(1.5), decimal.NewFromFloat(2.5), decimal.NewFromFloat(1.5)},
		{"a > b", decimal.NewFromFloat(3.5), decimal.NewFromFloat(2.5), decimal.NewFromFloat(2.5)},
		{"a == b", decimal.NewFromFloat(2.5), decimal.NewFromFloat(2.5), decimal.NewFromFloat(2.5)},
		{"negative values", decimal.NewFromFloat(-1.5), decimal.NewFromFloat(-2.5), decimal.NewFromFloat(-2.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinDecimal(tt.a, tt.b)
			if !result.Equal(tt.expected) {
				t.Errorf("MinDecimal(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMaxDecimal(t *testing.T) {
	tests := []struct {
		name     string
		a, b     decimal.Decimal
		expected decimal.Decimal
	}{
		{"a < b", decimal.NewFromFloat(1.5), decimal.NewFromFloat(2.5), decimal.NewFromFloat(2.5)},
		{"a > b", decimal.NewFromFloat(3.5), decimal.NewFromFloat(2.5), decimal.NewFromFloat(3.5)},
		{"a == b", decimal.NewFromFloat(2.5), decimal.NewFromFloat(2.5), decimal.NewFromFloat(2.5)},
		{"negative values", decimal.NewFromFloat(-1.5), decimal.NewFromFloat(-2.5), decimal.NewFromFloat(-1.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxDecimal(tt.a, tt.b)
			if !result.Equal(tt.expected) {
				t.Errorf("MaxDecimal(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestAbsDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    decimal.Decimal
		expected decimal.Decimal
	}{
		{"positive", decimal.NewFromFloat(5.5), decimal.NewFromFloat(5.5)},
		{"negative", decimal.NewFromFloat(-3.2), decimal.NewFromFloat(3.2)},
		{"zero", decimal.Zero, decimal.Zero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AbsDecimal(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("AbsDecimal(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPercentChange(t *testing.T) {
	tests := []struct {
		name     string
		oldValue decimal.Decimal
		newValue decimal.Decimal
		expected decimal.Decimal
	}{
		{"increase", decimal.NewFromFloat(100), decimal.NewFromFloat(120), decimal.NewFromFloat(20)},
		{"decrease", decimal.NewFromFloat(100), decimal.NewFromFloat(80), decimal.NewFromFloat(-20)},
		{"no change", decimal.NewFromFloat(100), decimal.NewFromFloat(100), decimal.Zero},
		{"from zero", decimal.Zero, decimal.NewFromFloat(100), decimal.Zero},
		{"to zero", decimal.NewFromFloat(100), decimal.Zero, decimal.NewFromFloat(-100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PercentChange(tt.oldValue, tt.newValue)
			if !result.Equal(tt.expected) {
				t.Errorf("PercentChange(%v, %v) = %v, want %v", tt.oldValue, tt.newValue, result, tt.expected)
			}
		})
	}
}

func TestStandardDeviation(t *testing.T) {
	tests := []struct {
		name     string
		values   []decimal.Decimal
		expected decimal.Decimal
	}{
		{"empty slice", []decimal.Decimal{}, decimal.Zero},
		{"single value", []decimal.Decimal{decimal.NewFromFloat(5)}, decimal.Zero},
		{"two same values", []decimal.Decimal{decimal.NewFromFloat(5), decimal.NewFromFloat(5)}, decimal.Zero},
		{"simple case", []decimal.Decimal{decimal.NewFromFloat(1), decimal.NewFromFloat(2), decimal.NewFromFloat(3)}, decimal.NewFromFloat(0.816496580927726)},
		{"larger spread", []decimal.Decimal{decimal.NewFromFloat(10), decimal.NewFromFloat(20), decimal.NewFromFloat(30)}, decimal.NewFromFloat(8.16496580927726)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StandardDeviation(tt.values)
			// Allow small tolerance for floating point precision
			diff := result.Sub(tt.expected).Abs()
			if diff.GreaterThan(decimal.NewFromFloat(0.01)) {
				t.Errorf("StandardDeviation(%v) = %v, want %v", tt.values, result, tt.expected)
			}
		})
	}
}

func TestClampDecimal(t *testing.T) {
	tests := []struct {
		name     string
		value    decimal.Decimal
		min      decimal.Decimal
		max      decimal.Decimal
		expected decimal.Decimal
	}{
		{"within range", decimal.NewFromFloat(5), decimal.NewFromFloat(0), decimal.NewFromFloat(10), decimal.NewFromFloat(5)},
		{"below min", decimal.NewFromFloat(-5), decimal.NewFromFloat(0), decimal.NewFromFloat(10), decimal.NewFromFloat(0)},
		{"above max", decimal.NewFromFloat(15), decimal.NewFromFloat(0), decimal.NewFromFloat(10), decimal.NewFromFloat(10)},
		{"equal to min", decimal.NewFromFloat(0), decimal.NewFromFloat(0), decimal.NewFromFloat(10), decimal.NewFromFloat(0)},
		{"equal to max", decimal.NewFromFloat(10), decimal.NewFromFloat(0), decimal.NewFromFloat(10), decimal.NewFromFloat(10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClampDecimal(tt.value, tt.min, tt.max)
			if !result.Equal(tt.expected) {
				t.Errorf("ClampDecimal(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestLerpDecimal(t *testing.T) {
	tests := []struct {
		name     string
		a, b     decimal.Decimal
		t        float64
		expected decimal.Decimal
	}{
		{"t=0", decimal.NewFromFloat(0), decimal.NewFromFloat(10), 0.0, decimal.NewFromFloat(0)},
		{"t=1", decimal.NewFromFloat(0), decimal.NewFromFloat(10), 1.0, decimal.NewFromFloat(10)},
		{"t=0.5", decimal.NewFromFloat(0), decimal.NewFromFloat(10), 0.5, decimal.NewFromFloat(5)},
		{"t=0.25", decimal.NewFromFloat(0), decimal.NewFromFloat(10), 0.25, decimal.NewFromFloat(2.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LerpDecimal(tt.a, tt.b, tt.t)
			if !result.Equal(tt.expected) {
				t.Errorf("LerpDecimal(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.t, result, tt.expected)
			}
		})
	}
}

func TestIsWithinRange(t *testing.T) {
	tests := []struct {
		name     string
		value    decimal.Decimal
		min      decimal.Decimal
		max      decimal.Decimal
		expected bool
	}{
		{"within range", decimal.NewFromFloat(5), decimal.NewFromFloat(0), decimal.NewFromFloat(10), true},
		{"below min", decimal.NewFromFloat(-5), decimal.NewFromFloat(0), decimal.NewFromFloat(10), false},
		{"above max", decimal.NewFromFloat(15), decimal.NewFromFloat(0), decimal.NewFromFloat(10), false},
		{"equal to min", decimal.NewFromFloat(0), decimal.NewFromFloat(0), decimal.NewFromFloat(10), true},
		{"equal to max", decimal.NewFromFloat(10), decimal.NewFromFloat(0), decimal.NewFromFloat(10), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWithinRange(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("IsWithinRange(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func BenchmarkStandardDeviation(b *testing.B) {
	// Generate large dataset
	values := make([]decimal.Decimal, 10000)
	for i := 0; i < 10000; i++ {
		values[i] = decimal.NewFromFloat(100 + float64(i)*0.1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StandardDeviation(values)
	}
}

func BenchmarkPercentChange(b *testing.B) {
	oldValue := decimal.NewFromFloat(100)
	newValue := decimal.NewFromFloat(105)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PercentChange(oldValue, newValue)
	}
}
