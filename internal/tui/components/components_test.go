package components

import (
	"strings"
	"testing"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/shopspring/decimal"
)

func TestRenderBalanceCard(t *testing.T) {
	tests := []struct {
		name        string
		balance     decimal.Decimal
		dailyPnL    decimal.Decimal
		totalPnL    decimal.Decimal
		expectWords []string
	}{
		{
			name:        "positive balance and PnL",
			balance:     decimal.NewFromFloat(10000),
			dailyPnL:    decimal.NewFromFloat(500),
			totalPnL:    decimal.NewFromFloat(1500),
			expectWords: []string{"$10000.00", "$500.00", "$1500.00"},
		},
		{
			name:        "negative PnL",
			balance:     decimal.NewFromFloat(10000),
			dailyPnL:    decimal.NewFromFloat(-200),
			totalPnL:    decimal.NewFromFloat(-500),
			expectWords: []string{"$10000.00", "$-200.00", "$-500.00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderBalanceCard(tt.balance, tt.dailyPnL, tt.totalPnL)

			// Check that result contains expected elements
			if !strings.Contains(result, "💰 Account Balance") {
				t.Error("Balance card should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Balance card should contain %s", word)
				}
			}
		})
	}
}

func TestRenderStatsCard(t *testing.T) {
	tests := []struct {
		name        string
		stats       *risk.Stats
		expectWords []string
	}{
		{
			name: "with stats",
			stats: &risk.Stats{
				TotalTrades:     100,
				WinningTrades:   60,
				LosingTrades:    40,
				WinRate:         60.0,
				ProfitFactor:    1.5,
				CurrentDrawdown: decimal.NewFromFloat(2.5),
			},
			expectWords: []string{"100", "60.0%", "60/40", "1.50", "2.50%"},
		},
		{
			name:        "nil stats",
			stats:       nil,
			expectWords: []string{"No statistics available"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderStatsCard(tt.stats)

			if !strings.Contains(result, "📊 Trading Stats") {
				t.Error("Stats card should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Stats card should contain %s", word)
				}
			}
		})
	}
}

func TestRenderActivityCard(t *testing.T) {
	tests := []struct {
		name        string
		messages    []string
		expectWords []string
	}{
		{
			name:        "with messages",
			messages:    []string{"Order placed", "Position opened"},
			expectWords: []string{"Order placed", "Position opened"},
		},
		{
			name:        "no messages",
			messages:    []string{},
			expectWords: []string{"No recent activity"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderActivityCard(tt.messages)

			if !strings.Contains(result, "📝 Recent Activity") {
				t.Error("Activity card should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Activity card should contain %s", word)
				}
			}
		})
	}
}

func TestRenderRiskCard(t *testing.T) {
	tests := []struct {
		name              string
		canTrade          bool
		reason            string
		consecutiveLosses int
		tradesExecuted    int
		maxTrades         int
		expectWords       []string
	}{
		{
			name:              "trading allowed",
			canTrade:          true,
			reason:            "",
			consecutiveLosses: 0,
			tradesExecuted:    5,
			maxTrades:         10,
			expectWords:       []string{"ALLOWED", "0", "5/10"},
		},
		{
			name:              "trading blocked",
			canTrade:          false,
			reason:            "Max losses reached",
			consecutiveLosses: 3,
			tradesExecuted:    10,
			maxTrades:         10,
			expectWords:       []string{"BLOCKED", "Max losses reached", "3", "10/10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderRiskCard(tt.canTrade, tt.reason, tt.consecutiveLosses, tt.tradesExecuted, tt.maxTrades)

			if !strings.Contains(result, "🛡️  Risk Management") {
				t.Error("Risk card should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Risk card should contain %s", word)
				}
			}
		})
	}
}

func TestRenderPositions(t *testing.T) {
	tests := []struct {
		name        string
		positions   []*order.ManagedPosition
		expectWords []string
	}{
		{
			name: "with positions",
			positions: []*order.ManagedPosition{
				{
					Symbol:        "BTC-USD",
					Side:          order.PositionSideLong,
					EntryPrice:    decimal.NewFromFloat(50000),
					CurrentPrice:  decimal.NewFromFloat(51000),
					Amount:        decimal.NewFromFloat(0.1),
					UnrealizedPnL: decimal.NewFromFloat(100),
				},
			},
			expectWords: []string{"BTC-USD", "LONG", "$50000.00", "$51000.00", "0.1000", "$100.00"},
		},
		{
			name:        "no positions",
			positions:   []*order.ManagedPosition{},
			expectWords: []string{"No open positions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPositions(tt.positions)

			if !strings.Contains(result, "📈 Open Positions") {
				t.Error("Positions should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Positions should contain %s", word)
				}
			}
		})
	}
}

func TestRenderPositionDetail(t *testing.T) {
	tests := []struct {
		name        string
		position    *order.ManagedPosition
		expectWords []string
	}{
		{
			name: "with position",
			position: &order.ManagedPosition{
				Symbol:        "BTC-USD",
				Side:          order.PositionSideLong,
				Status:        order.PositionStatusOpen,
				EntryPrice:    decimal.NewFromFloat(50000),
				CurrentPrice:  decimal.NewFromFloat(51000),
				Amount:        decimal.NewFromFloat(0.1),
				Leverage:      decimal.NewFromInt(1),
				UnrealizedPnL: decimal.NewFromFloat(100),
				RealizedPnL:   decimal.NewFromFloat(50),
				EntryTime:     time.Now().Add(-time.Hour),
			},
			expectWords: []string{"BTC-USD", "LONG", "open", "$50000.00", "$51000.00", "0.1000", "1x", "$100.00", "$50.00"},
		},
		{
			name:        "nil position",
			position:    nil,
			expectWords: []string{"No position selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPositionDetail(tt.position)

			if !strings.Contains(result, "📊 Position Details") {
				t.Error("Position detail should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Position detail should contain %s", word)
				}
			}
		})
	}
}

func TestRenderPositionSummary(t *testing.T) {
	tests := []struct {
		name        string
		positions   []*order.ManagedPosition
		expectWords []string
	}{
		{
			name: "mixed positions",
			positions: []*order.ManagedPosition{
				{
					Side:          order.PositionSideLong,
					UnrealizedPnL: decimal.NewFromFloat(100),
					RealizedPnL:   decimal.NewFromFloat(50),
				},
				{
					Side:          order.PositionSideShort,
					UnrealizedPnL: decimal.NewFromFloat(-50),
					RealizedPnL:   decimal.NewFromFloat(25),
				},
			},
			expectWords: []string{"2", "1", "1", "$50.00", "$75.00", "$125.00"},
		},
		{
			name:        "no positions",
			positions:   []*order.ManagedPosition{},
			expectWords: []string{"0", "$0.00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPositionSummary(tt.positions)

			if !strings.Contains(result, "💼 Position Summary") {
				t.Error("Position summary should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Position summary should contain %s", word)
				}
			}
		})
	}
}

func TestRenderOrderBook(t *testing.T) {
	tests := []struct {
		name        string
		orderbook   *exchanges.OrderBook
		depth       int
		expectWords []string
	}{
		{
			name: "with orderbook",
			orderbook: &exchanges.OrderBook{
				Bids: []exchanges.Level{
					{Price: decimal.NewFromFloat(49900), Amount: decimal.NewFromFloat(1.0)},
					{Price: decimal.NewFromFloat(49800), Amount: decimal.NewFromFloat(2.0)},
				},
				Asks: []exchanges.Level{
					{Price: decimal.NewFromFloat(50100), Amount: decimal.NewFromFloat(1.0)},
					{Price: decimal.NewFromFloat(50200), Amount: decimal.NewFromFloat(2.0)},
				},
			},
			depth:       2,
			expectWords: []string{"49900.00", "49800.00", "50100.00", "50200.00", "Spread"},
		},
		{
			name:        "nil orderbook",
			orderbook:   nil,
			depth:       5,
			expectWords: []string{"No order book data available"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderOrderBook(tt.orderbook, tt.depth)

			if !strings.Contains(result, "📖 Order Book") {
				t.Error("Order book should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Order book should contain %s", word)
				}
			}
		})
	}
}

func TestRenderOrderBookDepth(t *testing.T) {
	tests := []struct {
		name        string
		orderbook   *exchanges.OrderBook
		maxDepth    int
		expectWords []string
	}{
		{
			name: "with orderbook",
			orderbook: &exchanges.OrderBook{
				Bids: []exchanges.Level{
					{Price: decimal.NewFromFloat(49900), Amount: decimal.NewFromFloat(1.0)},
				},
				Asks: []exchanges.Level{
					{Price: decimal.NewFromFloat(50100), Amount: decimal.NewFromFloat(1.0)},
				},
			},
			maxDepth:    1,
			expectWords: []string{"49900.00", "50100.00"},
		},
		{
			name:        "nil orderbook",
			orderbook:   nil,
			maxDepth:    5,
			expectWords: []string{"No data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderOrderBookDepth(tt.orderbook, tt.maxDepth)

			if !strings.Contains(result, "📊 Order Book Depth") {
				t.Error("Order book depth should contain header")
			}

			for _, word := range tt.expectWords {
				if !strings.Contains(result, word) {
					t.Errorf("Order book depth should contain %s", word)
				}
			}
		})
	}
}
