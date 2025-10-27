package dydx

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
)

// TestGetAllMarkets tests retrieval of all available markets
func TestGetAllMarkets(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	markets, err := client.GetAllMarkets(ctx)
	if err != nil {
		t.Fatalf("Failed to get all markets: %v", err)
	}

	// Should return at least some markets
	if len(markets) == 0 {
		t.Errorf("Expected at least one market, got 0")
	}

	// Verify market structure
	for symbol, marketData := range markets {
		if symbol == "" {
			t.Error("Market symbol is empty")
		}
		if marketData.Ticker == "" {
			t.Errorf("Market %s has empty ticker", symbol)
		}
		if marketData.Status == "" {
			t.Errorf("Market %s has empty status", symbol)
		}
		if marketData.StepSize.LessThanOrEqual(decimal.Zero) {
			t.Errorf("Market %s has invalid step size: %s", symbol, marketData.StepSize.String())
		}
	}
}

// TestEvaluateMarketQuality tests market quality evaluation
func TestEvaluateMarketQuality(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name        string
		market      string
		shouldError bool
	}{
		{"BTC-USD", "BTC-USD", false},
		{"ETH-USD", "ETH-USD", false},
		{"SOL-USD", "SOL-USD", false},
		{"INVALID", "INVALID-MARKET", true},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality, err := client.EvaluateMarketQuality(ctx, tt.market)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for market %s, got nil", tt.market)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error for market %s: %v", tt.market, err)
			}

			if !tt.shouldError {
				// Verify quality score is in valid range
				if quality.QualityScore < 0 || quality.QualityScore > 1 {
					t.Errorf("Quality score out of range [0, 1]: %f", quality.QualityScore)
				}

				// Volume24h should be non-negative
				if quality.Volume24h.LessThan(decimal.Zero) {
					t.Errorf("Volume24h should be non-negative, got %s", quality.Volume24h.String())
				}

				// Volatility should be reasonable
				if quality.Volatility < 0 || quality.Volatility > 1 {
					t.Errorf("Volatility out of range [0, 1]: %f", quality.Volatility)
				}

				// Liquidity should be reasonable
				if quality.Liquidity < 0 || quality.Liquidity > 1 {
					t.Errorf("Liquidity out of range [0, 1]: %f", quality.Liquidity)
				}
			}
		})
	}
}

// TestFilterMarketsByQuality tests filtering markets by quality criteria
func TestFilterMarketsByQuality(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with minimum quality threshold
	markets, err := client.FilterMarketsByQuality(ctx, 0.3)
	if err != nil {
		t.Fatalf("Failed to filter markets: %v", err)
	}

	// Should return some markets with quality >= 0.3
	if len(markets) == 0 {
		t.Logf("Warning: No markets found with quality >= 0.3")
	}

	// Verify all returned markets meet the threshold
	for symbol, quality := range markets {
		if quality.QualityScore < 0.3 {
			t.Errorf("Market %s has quality %f below threshold 0.3", symbol, quality.QualityScore)
		}
	}
}

// TestSelectBestMarkets tests selection of top markets by quality
func TestSelectBestMarkets(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		maxCount   int
		minQuality float64
	}{
		{"Top 5 markets", 5, 0.0},
		{"Top 10 high quality", 10, 0.5},
		{"Top 1", 1, 0.0},
		{"Top 20 very high quality", 20, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markets, err := client.SelectBestMarkets(ctx, tt.maxCount, tt.minQuality)
			if err != nil {
				t.Fatalf("Failed to select markets: %v", err)
			}

			// Should not exceed maxCount
			if len(markets) > tt.maxCount {
				t.Errorf("Expected at most %d markets, got %d", tt.maxCount, len(markets))
			}

			// Verify all markets meet minimum quality
			for i, market := range markets {
				if market.QualityScore < tt.minQuality {
					t.Errorf("Market %d (%s) has quality %f below minimum %f",
						i, market.Symbol, market.QualityScore, tt.minQuality)
				}

				// Markets should be sorted by quality (descending)
				if i > 0 && market.QualityScore > markets[i-1].QualityScore {
					t.Errorf("Markets not sorted by quality: %s (%.3f) > %s (%.3f)",
						market.Symbol, market.QualityScore,
						markets[i-1].Symbol, markets[i-1].QualityScore)
				}
			}
		})
	}
}

// TestMarketQualityScoring tests the quality scoring algorithm
func TestMarketQualityScoring(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get a few markets to test scoring
	markets, err := client.SelectBestMarkets(ctx, 10, 0.0)
	if err != nil {
		t.Fatalf("Failed to get markets: %v", err)
	}

	// Test score composition
	for _, market := range markets {
		// Verify quality score is composition of components
		if market.QualityScore < 0 || market.QualityScore > 1 {
			t.Errorf("Quality score out of range: %f", market.QualityScore)
		}

		// Volume and liquidity should contribute to quality
		if market.Volume24h.GreaterThan(decimal.Zero) {
			if market.Liquidity < 0.1 {
				t.Logf("Warning: High volume market %s has low liquidity", market.Symbol)
			}
		}

		// Test volatility impact
		if market.Volatility > 0.8 {
			if market.QualityScore > 0.8 {
				t.Logf("Note: High volatility market %s still has high quality score", market.Symbol)
			}
		}
	}
}

// TestCachedMarketData tests market data caching
func TestCachedMarketData(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// First call
	markets1, err := client.GetAllMarkets(ctx)
	if err != nil {
		t.Fatalf("First GetAllMarkets failed: %v", err)
	}

	// Second call (should use cache)
	markets2, err := client.GetAllMarkets(ctx)
	if err != nil {
		t.Fatalf("Second GetAllMarkets failed: %v", err)
	}

	// Should return same markets
	if len(markets1) != len(markets2) {
		t.Errorf("Market count mismatch: %d vs %d", len(markets1), len(markets2))
	}

	// Spot check some markets
	for symbol := range markets1 {
		if _, exists := markets2[symbol]; !exists {
			t.Errorf("Market %s missing in second call", symbol)
		}
	}
}

// TestMarketDataStructure tests MarketQuality struct
func TestMarketDataStructure(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	quality, err := client.EvaluateMarketQuality(ctx, "BTC-USD")
	if err != nil {
		t.Fatalf("Failed to evaluate quality: %v", err)
	}

	// Verify all fields are populated
	if quality.Symbol == "" {
		t.Error("Symbol is empty")
	}
	if quality.Volume24h.IsZero() {
		t.Logf("Info: Zero volume for %s (might be normal)", quality.Symbol)
	}
	if quality.Volatility < 0 {
		t.Error("Volatility is negative")
	}
	if quality.Liquidity < 0 {
		t.Error("Liquidity is negative")
	}
	if quality.QualityScore < 0 {
		t.Error("Quality score is negative")
	}
}

// TestMarketSelectionConsistency tests consistent ranking across multiple selections
func TestMarketSelectionConsistency(t *testing.T) {
	client, err := NewClient("", "")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get selections multiple times
	selection1, err := client.SelectBestMarkets(ctx, 5, 0.0)
	if err != nil {
		t.Fatalf("First selection failed: %v", err)
	}

	selection2, err := client.SelectBestMarkets(ctx, 5, 0.0)
	if err != nil {
		t.Fatalf("Second selection failed: %v", err)
	}

	// Should return same markets in same order
	if len(selection1) != len(selection2) {
		t.Errorf("Selection count mismatch: %d vs %d", len(selection1), len(selection2))
	}

	for i := range selection1 {
		if selection1[i].Symbol != selection2[i].Symbol {
			t.Errorf("Symbol mismatch at position %d: %s vs %s",
				i, selection1[i].Symbol, selection2[i].Symbol)
		}
	}
}
