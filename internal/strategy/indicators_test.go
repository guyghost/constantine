package strategy

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestEMA(t *testing.T) {
	// Test data: simple price series
	prices := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(11),
		decimal.NewFromFloat(12),
		decimal.NewFromFloat(13),
		decimal.NewFromFloat(14),
		decimal.NewFromFloat(15),
	}

	// Test with period 3
	result := EMA(prices, 3)

	// Should return 4 values (len(prices) - period + 1 = 6 - 3 + 1 = 4)
	if len(result) != 4 {
		t.Errorf("expected 4 EMA values, got %d", len(result))
	}

	// Test insufficient data
	shortPrices := []decimal.Decimal{decimal.NewFromFloat(10), decimal.NewFromFloat(11)}
	result = EMA(shortPrices, 3)
	if len(result) != 0 {
		t.Errorf("expected empty result for insufficient data, got %d values", len(result))
	}
}

func TestSMA(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(11),
		decimal.NewFromFloat(12),
		decimal.NewFromFloat(13),
		decimal.NewFromFloat(14),
	}

	// Test with period 3
	result := SMA(prices, 3)

	// Should return 3 values (len(prices) - period + 1 = 5 - 3 + 1 = 3)
	if len(result) != 3 {
		t.Errorf("expected 3 SMA values, got %d", len(result))
	}

	// Check first SMA: (10 + 11 + 12) / 3 = 11
	expected := decimal.NewFromFloat(11)
	if !result[0].Equal(expected) {
		t.Errorf("expected first SMA %s, got %s", expected, result[0])
	}

	// Check second SMA: (11 + 12 + 13) / 3 = 12
	expected = decimal.NewFromFloat(12)
	if !result[1].Equal(expected) {
		t.Errorf("expected second SMA %s, got %s", expected, result[1])
	}

	// Test insufficient data
	shortPrices := []decimal.Decimal{decimal.NewFromFloat(10), decimal.NewFromFloat(11)}
	result = SMA(shortPrices, 3)
	if len(result) != 0 {
		t.Errorf("expected empty result for insufficient data, got %d values", len(result))
	}
}

func TestRSI(t *testing.T) {
	// Create a price series with known RSI values
	prices := []decimal.Decimal{
		decimal.NewFromFloat(44.34),
		decimal.NewFromFloat(44.09),
		decimal.NewFromFloat(44.15),
		decimal.NewFromFloat(43.61),
		decimal.NewFromFloat(44.33),
		decimal.NewFromFloat(44.38),
		decimal.NewFromFloat(44.11),
		decimal.NewFromFloat(43.75),
		decimal.NewFromFloat(44.24),
		decimal.NewFromFloat(44.44),
		decimal.NewFromFloat(44.38),
		decimal.NewFromFloat(44.15),
		decimal.NewFromFloat(43.79),
		decimal.NewFromFloat(43.90),
		decimal.NewFromFloat(44.14),
	}

	result := RSI(prices, 14)

	// Should return 1 value (len(prices) - period = 15 - 14 = 1)
	if len(result) != 1 {
		t.Errorf("expected 1 RSI value, got %d", len(result))
	}

	// RSI should be between 0 and 100
	if result[0].LessThan(decimal.Zero) || result[0].GreaterThan(decimal.NewFromInt(100)) {
		t.Errorf("RSI should be between 0 and 100, got %s", result[0])
	}

	// Test insufficient data
	shortPrices := []decimal.Decimal{decimal.NewFromFloat(10), decimal.NewFromFloat(11)}
	result = RSI(shortPrices, 14)
	if len(result) != 0 {
		t.Errorf("expected empty result for insufficient data, got %d values", len(result))
	}
}

func TestBollingerBands(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(11),
		decimal.NewFromFloat(12),
		decimal.NewFromFloat(13),
		decimal.NewFromFloat(14),
		decimal.NewFromFloat(15),
		decimal.NewFromFloat(16),
		decimal.NewFromFloat(17),
		decimal.NewFromFloat(18),
		decimal.NewFromFloat(19),
	}

	upper, middle, lower := BollingerBands(prices, 5, 2.0)

	// Should return 6 values (len(prices) - period + 1 = 10 - 5 + 1 = 6)
	if len(upper) != 6 || len(middle) != 6 || len(lower) != 6 {
		t.Errorf("expected 6 values for each band, got upper: %d, middle: %d, lower: %d",
			len(upper), len(middle), len(lower))
	}

	// Upper should be greater than middle, middle greater than lower
	for i := 0; i < len(middle); i++ {
		if !upper[i].GreaterThan(middle[i]) {
			t.Errorf("upper band should be greater than middle band at index %d", i)
		}
		if !middle[i].GreaterThan(lower[i]) {
			t.Errorf("middle band should be greater than lower band at index %d", i)
		}
	}

	// Test insufficient data
	shortPrices := []decimal.Decimal{decimal.NewFromFloat(10), decimal.NewFromFloat(11)}
	upper, middle, lower = BollingerBands(shortPrices, 5, 2.0)
	if len(upper) != 0 || len(middle) != 0 || len(lower) != 0 {
		t.Errorf("expected empty results for insufficient data, got upper: %d, middle: %d, lower: %d",
			len(upper), len(middle), len(lower))
	}
}

func TestVWAP(t *testing.T) {
	prices := []decimal.Decimal{
		decimal.NewFromFloat(100),
		decimal.NewFromFloat(101),
		decimal.NewFromFloat(102),
	}

	volumes := []decimal.Decimal{
		decimal.NewFromFloat(10),
		decimal.NewFromFloat(20),
		decimal.NewFromFloat(30),
	}

	result := VWAP(prices, volumes)

	// VWAP = (100*10 + 101*20 + 102*30) / (10 + 20 + 30) = (1000 + 2020 + 3060) / 60 = 6080 / 60 = 101.333...
	expected := decimal.NewFromFloat(101.33333333333333)
	if !result.Round(5).Equal(expected.Round(5)) {
		t.Errorf("expected VWAP %s, got %s", expected, result)
	}

	// Test empty data
	result = VWAP([]decimal.Decimal{}, []decimal.Decimal{})
	if !result.Equal(decimal.Zero) {
		t.Errorf("expected zero VWAP for empty data, got %s", result)
	}

	// Test mismatched lengths
	result = VWAP([]decimal.Decimal{decimal.NewFromFloat(100)}, []decimal.Decimal{})
	if !result.Equal(decimal.Zero) {
		t.Errorf("expected zero VWAP for mismatched lengths, got %s", result)
	}

	// Test zero volume
	prices = []decimal.Decimal{decimal.NewFromFloat(100)}
	volumes = []decimal.Decimal{decimal.Zero}
	result = VWAP(prices, volumes)
	if !result.Equal(decimal.Zero) {
		t.Errorf("expected zero VWAP for zero volume, got %s", result)
	}
}

func TestStochastic(t *testing.T) {
	high := []decimal.Decimal{
		decimal.NewFromFloat(127.01),
		decimal.NewFromFloat(127.62),
		decimal.NewFromFloat(126.59),
		decimal.NewFromFloat(127.35),
		decimal.NewFromFloat(128.17),
	}

	low := []decimal.Decimal{
		decimal.NewFromFloat(125.36),
		decimal.NewFromFloat(126.16),
		decimal.NewFromFloat(124.93),
		decimal.NewFromFloat(126.09),
		decimal.NewFromFloat(126.82),
	}

	close := []decimal.Decimal{
		decimal.NewFromFloat(126.00),
		decimal.NewFromFloat(126.60),
		decimal.NewFromFloat(127.10),
		decimal.NewFromFloat(127.29),
		decimal.NewFromFloat(127.18),
	}

	result := Stochastic(high, low, close, 5)

	// Should return 1 value (len(close) - period + 1 = 5 - 5 + 1 = 1)
	if len(result) != 1 {
		t.Errorf("expected 1 stochastic value, got %d", len(result))
	}

	// Stochastic should be between 0 and 100
	if result[0].LessThan(decimal.Zero) || result[0].GreaterThan(decimal.NewFromInt(100)) {
		t.Errorf("stochastic should be between 0 and 100, got %s", result[0])
	}

	// Test insufficient data
	shortHigh := []decimal.Decimal{decimal.NewFromFloat(127)}
	shortLow := []decimal.Decimal{decimal.NewFromFloat(125)}
	shortClose := []decimal.Decimal{decimal.NewFromFloat(126)}
	result = Stochastic(shortHigh, shortLow, shortClose, 5)
	if len(result) != 0 {
		t.Errorf("expected empty result for insufficient data, got %d values", len(result))
	}
}
