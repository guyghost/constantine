#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  dYdX Candle Polling Test${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}\n"

echo -e "${YELLOW}Testing dYdX candle polling directly...${NC}\n"

# Create a simple Go test program
cat > /tmp/test_candles.go << 'GOEOF'
package main

import (
	"context"
	"fmt"
	"time"
	"os"

	"github.com/guyghost/constantine/internal/exchanges/dydx"
)

func main() {
	fmt.Println("🔍 Testing dYdX candle polling (30 second test)...")
	fmt.Println("Symbol: BTC-USD")
	fmt.Println("Interval: 1m\n")

	client := dydx.NewClient("", "")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	candleCount := 0
	startTime := time.Now()

	err := client.SubscribeCandles(ctx, "BTC-USD", "1m", func(candle *dydx.exchanges.Candle) {
		candleCount++
		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%6.1fs] 📊 Candle #%d: %s Close=%.2f High=%.2f Low=%.2f Vol=%.0f\n",
			elapsed, candleCount,
			candle.Timestamp.Format("15:04:05"),
			candle.Close.InexactFloat64(),
			candle.High.InexactFloat64(),
			candle.Low.InexactFloat64(),
			candle.Volume.InexactFloat64())
	})

	if err != nil {
		fmt.Printf("\n❌ Error subscribing to candles: %v\n", err)
		os.Exit(1)
	}

	<-ctx.Done()

	elapsed := time.Since(startTime).Seconds()
	fmt.Printf("\n📊 Test Results:\n")
	fmt.Printf("   Candles received: %d\n", candleCount)
	fmt.Printf("   Time elapsed: %.1f seconds\n", elapsed)

	if candleCount == 0 {
		fmt.Println("\n   ❌ NO CANDLES RECEIVED!")
		fmt.Println("\n   Check:")
		fmt.Println("   • Network connectivity to dYdX API")
		fmt.Println("   • API rate limits not exceeded")
		fmt.Println("   • Symbol BTC-USD exists on dYdX")
		os.Exit(1)
	} else {
		avgInterval := elapsed / float64(candleCount)
		fmt.Printf("   Average interval: %.1f seconds per candle\n", avgInterval)
		if avgInterval <= 15 {
			fmt.Println("\n   ✅ Candles are flowing at good rate (≤15s apart)")
			fmt.Println("   ✅ Ready to generate signals")
		} else if avgInterval <= 60 {
			fmt.Println("\n   ⚠️  Candles are flowing but slow (>15s apart)")
			fmt.Println("   ⚠️  Signals may be delayed")
		}
	}
}
GOEOF

# Run the test
cd /Users/guy/Developer/guyghost/constantine
go run /tmp/test_candles.go
