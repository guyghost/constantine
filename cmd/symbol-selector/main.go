package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/guyghost/constantine/internal/exchanges/dydx"
)

func main() {
	maxSymbols := flag.Int("max", 10, "Maximum number of symbols to select")
	minQuality := flag.Float64("min-quality", 0.3, "Minimum quality score [0, 1]")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("          dYdX Market Symbol Selector v1.0")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Printf("\nConfiguration:\n")
	fmt.Printf("  Max Symbols: %d\n", *maxSymbols)
	fmt.Printf("  Min Quality: %.1f%%\n", *minQuality*100)
	fmt.Printf("  Verbose: %v\n\n", *verbose)

	// Create client
	client, err := dydx.NewClient("", "")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Step 1: Retrieve all available markets
	fmt.Println("Step 1: Retrieving all available markets...")
	start := time.Now()
	allMarkets, err := client.GetAllMarkets(ctx)
	if err != nil {
		log.Fatalf("Failed to get markets: %v", err)
	}
	fmt.Printf("✓ Retrieved %d total markets in %.2fs\n\n", len(allMarkets), time.Since(start).Seconds())

	// Step 2: Filter by quality
	fmt.Println("Step 2: Filtering markets by quality criteria...")
	fmt.Println("  Evaluation criteria:")
	fmt.Println("    - 24h Volume (35% weight)")
	fmt.Println("    - Liquidity Score (35% weight)")
	fmt.Println("    - Price Volatility (30% weight)")
	fmt.Println()

	start = time.Now()
	filtered, err := client.FilterMarketsByQuality(ctx, *minQuality)
	if err != nil {
		log.Fatalf("Failed to filter markets: %v", err)
	}
	fmt.Printf("✓ Filtered to %d quality markets in %.2fs\n\n", len(filtered), time.Since(start).Seconds())

	// Step 3: Select best markets
	fmt.Println("Step 3: Selecting top markets by quality score...")
	start = time.Now()
	bestMarkets, err := client.SelectBestMarkets(ctx, *maxSymbols, *minQuality)
	if err != nil {
		log.Fatalf("Failed to select best markets: %v", err)
	}
	fmt.Printf("✓ Selected %d best markets in %.2fs\n\n", len(bestMarkets), time.Since(start).Seconds())

	// Display results
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("                    RESULTS")
	fmt.Println("═══════════════════════════════════════════════════════════\n")

	if len(bestMarkets) == 0 {
		fmt.Println("⚠️  No markets met the selection criteria")
		os.Exit(0)
	}

	fmt.Printf("%-10s %-10s %-15s %-12s %-12s %-12s\n",
		"RANK", "SYMBOL", "QUALITY", "VOLUME(USD)", "LIQUIDITY", "VOLATILITY")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for i, market := range bestMarkets {
		volumeStr := formatVolume(market.Volume24h)

		fmt.Printf("%-10d %-10s %-10.1f%% %-15s %-12.1f%% %-12.1f%%\n",
			i+1,
			market.Symbol,
			market.QualityScore*100,
			volumeStr,
			market.Liquidity*100,
			market.Volatility*100,
		)
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════")

	// Verbose output
	if *verbose {
		fmt.Println("\n📊 DETAILED ANALYSIS\n")
		fmt.Println("Top 10 Recommended Symbols:")
		for i, market := range bestMarkets {
			if i >= 10 {
				break
			}
			fmt.Printf("\n%d. %s\n", i+1, market.Symbol)
			fmt.Printf("   Quality Score: %.2f%%\n", market.QualityScore*100)
			volumeStr := market.Volume24h.String()
			volumeFloat := 0.0
			fmt.Sscanf(volumeStr, "%f", &volumeFloat)
			fmt.Printf("   Volume (24h): $%.1fM\n", volumeFloat/1_000_000)
			fmt.Printf("   Liquidity: %.2f%%\n", market.Liquidity*100)
			fmt.Printf("   Volatility: %.2f%%\n", market.Volatility*100)

			if market.QualityScore >= 0.7 {
				fmt.Printf("   Status: ✅ EXCELLENT - Very high quality trading pair\n")
			} else if market.QualityScore >= 0.5 {
				fmt.Printf("   Status: 🟢 GOOD - Good quality trading pair\n")
			} else if market.QualityScore >= 0.3 {
				fmt.Printf("   Status: 🟡 FAIR - Moderate quality trading pair\n")
			} else {
				fmt.Printf("   Status: 🔴 POOR - Lower quality trading pair\n")
			}
		}
	}

	fmt.Println("\n✨ Recommendation: Start with the top 3-5 symbols for optimal trading")
}

func formatVolume(volume interface{}) string {
	var volumeUSD float64

	// Handle decimal type
	switch v := volume.(type) {
	case string:
		fmt.Sscanf(v, "%f", &volumeUSD)
	default:
		// Assume it's already a float or can be converted
		if f, ok := v.(float64); ok {
			volumeUSD = f
		} else if dec, ok := v.(any); ok {
			// Try to get string representation and parse
			str := fmt.Sprintf("%v", dec)
			fmt.Sscanf(str, "%f", &volumeUSD)
		}
	}

	switch {
	case volumeUSD >= 1_000_000_000:
		return fmt.Sprintf("$%.1fB", volumeUSD/1_000_000_000)
	case volumeUSD >= 1_000_000:
		return fmt.Sprintf("$%.0fM", volumeUSD/1_000_000)
	case volumeUSD >= 1_000:
		return fmt.Sprintf("$%.0fK", volumeUSD/1_000)
	default:
		return fmt.Sprintf("$%.0f", volumeUSD)
	}
}

func volumeToMillions(volume interface{}) float64 {
	var f float64

	// Handle decimal type
	switch v := volume.(type) {
	case string:
		fmt.Sscanf(v, "%f", &f)
	default:
		str := fmt.Sprintf("%v", v)
		fmt.Sscanf(str, "%f", &f)
	}

	return f / 1_000_000
}
