package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

// RenderOrderBook renders the order book component
func RenderOrderBook(orderbook *exchanges.OrderBook, depth int) string {
	var content strings.Builder

	content.WriteString("📖 Order Book\n\n")

	if orderbook == nil || len(orderbook.Asks) == 0 || len(orderbook.Bids) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		return boxStyle.Render(content.String() + mutedStyle.Render("No order book data available"))
	}

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(mutedColor)
	content.WriteString(headerStyle.Render(fmt.Sprintf("%-15s %-15s %-15s\n", "Price", "Amount", "Total")))
	content.WriteString(strings.Repeat("─", 45) + "\n\n")

	// Display asks (sell orders) in reverse order
	askStyle := lipgloss.NewStyle().Foreground(errorColor)
	asks := orderbook.Asks
	if len(asks) > depth {
		asks = asks[:depth]
	}

	// Reverse asks for display
	for i := len(asks) - 1; i >= 0; i-- {
		ask := asks[i]
		total := ask.Price.Mul(ask.Amount)
		line := fmt.Sprintf("%-15s %-15s %-15s\n",
			ask.Price.StringFixed(2),
			ask.Amount.StringFixed(4),
			total.StringFixed(2))
		content.WriteString(askStyle.Render(line))
	}

	// Spread separator
	if len(asks) > 0 && len(orderbook.Bids) > 0 {
		spread := orderbook.Asks[0].Price.Sub(orderbook.Bids[0].Price)
		spreadPercent := spread.Div(orderbook.Bids[0].Price).Mul(decimalFromFloat(100))

		content.WriteString("\n")
		spreadStyle := lipgloss.NewStyle().Foreground(warningColor).Bold(true)
		content.WriteString(spreadStyle.Render(fmt.Sprintf("Spread: $%s (%.4f%%)\n",
			spread.StringFixed(2),
			spreadPercent)))
		content.WriteString("\n")
	}

	// Display bids (buy orders)
	bidStyle := lipgloss.NewStyle().Foreground(successColor)
	bids := orderbook.Bids
	if len(bids) > depth {
		bids = bids[:depth]
	}

	for _, bid := range bids {
		total := bid.Price.Mul(bid.Amount)
		line := fmt.Sprintf("%-15s %-15s %-15s\n",
			bid.Price.StringFixed(2),
			bid.Amount.StringFixed(4),
			total.StringFixed(2))
		content.WriteString(bidStyle.Render(line))
	}

	return boxStyle.Render(content.String())
}

// RenderOrderBookDepth renders order book depth visualization
func RenderOrderBookDepth(orderbook *exchanges.OrderBook, maxDepth int) string {
	var content strings.Builder

	content.WriteString("📊 Order Book Depth\n\n")

	if orderbook == nil || len(orderbook.Asks) == 0 || len(orderbook.Bids) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		return boxStyle.Render(content.String() + mutedStyle.Render("No data"))
	}

	// Calculate max volume for scaling
	maxVolume := decimalFromFloat(0)
	depth := maxDepth
	if len(orderbook.Bids) < depth {
		depth = len(orderbook.Bids)
	}
	if len(orderbook.Asks) < depth {
		depth = len(orderbook.Asks)
	}

	for i := 0; i < depth; i++ {
		if orderbook.Bids[i].Amount.GreaterThan(maxVolume) {
			maxVolume = orderbook.Bids[i].Amount
		}
		if orderbook.Asks[i].Amount.GreaterThan(maxVolume) {
			maxVolume = orderbook.Asks[i].Amount
		}
	}

	// Render asks
	askStyle := lipgloss.NewStyle().Foreground(errorColor)
	for i := depth - 1; i >= 0; i-- {
		ask := orderbook.Asks[i]
		barLength := int(ask.Amount.Div(maxVolume).Mul(decimalFromFloat(20)).IntPart())
		bar := strings.Repeat("▓", barLength)
		content.WriteString(askStyle.Render(fmt.Sprintf("%10s │ %s\n",
			ask.Price.StringFixed(2), bar)))
	}

	content.WriteString(strings.Repeat("─", 35) + "\n")

	// Render bids
	bidStyle := lipgloss.NewStyle().Foreground(successColor)
	for i := 0; i < depth; i++ {
		bid := orderbook.Bids[i]
		barLength := int(bid.Amount.Div(maxVolume).Mul(decimalFromFloat(20)).IntPart())
		bar := strings.Repeat("▓", barLength)
		content.WriteString(bidStyle.Render(fmt.Sprintf("%10s │ %s\n",
			bid.Price.StringFixed(2), bar)))
	}

	return boxStyle.Render(content.String())
}

// Helper function to create decimal from float
func decimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}
