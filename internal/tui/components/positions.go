package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/constantine/internal/order"
	"github.com/shopspring/decimal"
)

// RenderPositions renders the positions list
func RenderPositions(positions []*order.ManagedPosition) string {
	var content strings.Builder

	content.WriteString("📈 Open Positions\n\n")

	if len(positions) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		return boxStyle.Render(content.String() + mutedStyle.Render("No open positions"))
	}

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(mutedColor)
	content.WriteString(headerStyle.Render(
		fmt.Sprintf("%-12s %-8s %-12s %-12s %-12s %-10s\n",
			"Symbol", "Side", "Entry", "Current", "Amount", "PnL")))
	content.WriteString(strings.Repeat("─", 70) + "\n")

	// Positions
	totalPnL := decimal.Zero
	for _, pos := range positions {
		sideStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
		side := "LONG"
		if pos.Side == order.PositionSideShort {
			sideStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
			side = "SHORT"
		}

		pnlStyle := lipgloss.NewStyle().Foreground(successColor)
		if pos.UnrealizedPnL.IsNegative() {
			pnlStyle = lipgloss.NewStyle().Foreground(errorColor)
		}

		totalPnL = totalPnL.Add(pos.UnrealizedPnL)

		line := fmt.Sprintf("%-12s %-8s %-12s %-12s %-12s %s\n",
			pos.Symbol,
			sideStyle.Render(side),
			"$"+pos.EntryPrice.StringFixed(2),
			"$"+pos.CurrentPrice.StringFixed(2),
			pos.Amount.StringFixed(4),
			pnlStyle.Render("$"+pos.UnrealizedPnL.StringFixed(2)))

		content.WriteString(line)
	}

	// Total
	content.WriteString(strings.Repeat("─", 70) + "\n")
	totalStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	if totalPnL.IsNegative() {
		totalStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	}
	content.WriteString(fmt.Sprintf("%-56s %s\n",
		"Total Unrealized PnL:",
		totalStyle.Render("$"+totalPnL.StringFixed(2))))

	return boxStyle.Render(content.String())
}

// RenderPositionDetail renders detailed information for a position
func RenderPositionDetail(pos *order.ManagedPosition) string {
	var content strings.Builder

	content.WriteString("📊 Position Details\n\n")

	if pos == nil {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		return boxStyle.Render(content.String() + mutedStyle.Render("No position selected"))
	}

	sideStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	side := "LONG"
	if pos.Side == order.PositionSideShort {
		sideStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
		side = "SHORT"
	}

	// Basic info
	content.WriteString(fmt.Sprintf("Symbol:        %s\n", pos.Symbol))
	content.WriteString(fmt.Sprintf("Side:          %s\n", sideStyle.Render(side)))
	content.WriteString(fmt.Sprintf("Status:        %s\n", pos.Status))
	content.WriteString("\n")

	// Prices
	content.WriteString(fmt.Sprintf("Entry Price:   $%s\n", pos.EntryPrice.StringFixed(2)))
	content.WriteString(fmt.Sprintf("Current Price: $%s\n", pos.CurrentPrice.StringFixed(2)))
	if !pos.StopLoss.IsZero() {
		content.WriteString(fmt.Sprintf("Stop Loss:     $%s\n", pos.StopLoss.StringFixed(2)))
	}
	if !pos.TakeProfit.IsZero() {
		content.WriteString(fmt.Sprintf("Take Profit:   $%s\n", pos.TakeProfit.StringFixed(2)))
	}
	content.WriteString("\n")

	// Size and leverage
	content.WriteString(fmt.Sprintf("Amount:        %s\n", pos.Amount.StringFixed(4)))
	content.WriteString(fmt.Sprintf("Leverage:      %sx\n", pos.Leverage.StringFixed(0)))
	content.WriteString("\n")

	// PnL
	pnlStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	if pos.UnrealizedPnL.IsNegative() {
		pnlStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	}
	content.WriteString(fmt.Sprintf("Unrealized PnL: %s\n", pnlStyle.Render("$"+pos.UnrealizedPnL.StringFixed(2))))
	content.WriteString(fmt.Sprintf("Realized PnL:   $%s\n", pos.RealizedPnL.StringFixed(2)))

	// Calculate PnL percentage
	pnlPercent := pos.UnrealizedPnL.Div(pos.EntryPrice.Mul(pos.Amount)).Mul(decimal.NewFromInt(100))
	content.WriteString(fmt.Sprintf("PnL %%:          %s\n", pnlStyle.Render(pnlPercent.StringFixed(2)+"%")))
	content.WriteString("\n")

	// Timing
	content.WriteString(fmt.Sprintf("Entry Time:    %s\n", pos.EntryTime.Format("2006-01-02 15:04:05")))
	if pos.ExitTime != nil {
		content.WriteString(fmt.Sprintf("Exit Time:     %s\n", pos.ExitTime.Format("2006-01-02 15:04:05")))
	} else {
		duration := time.Since(pos.EntryTime)
		content.WriteString(fmt.Sprintf("Duration:      %s\n", duration.Round(time.Second)))
	}

	return boxStyle.Render(content.String())
}

// RenderPositionSummary renders a summary of all positions
func RenderPositionSummary(positions []*order.ManagedPosition) string {
	var content strings.Builder

	content.WriteString("💼 Position Summary\n\n")

	totalPositions := len(positions)
	longPositions := 0
	shortPositions := 0
	totalUnrealizedPnL := decimal.Zero
	totalRealizedPnL := decimal.Zero

	for _, pos := range positions {
		if pos.Side == order.PositionSideLong {
			longPositions++
		} else {
			shortPositions++
		}
		totalUnrealizedPnL = totalUnrealizedPnL.Add(pos.UnrealizedPnL)
		totalRealizedPnL = totalRealizedPnL.Add(pos.RealizedPnL)
	}

	content.WriteString(fmt.Sprintf("Total Positions: %d\n", totalPositions))
	content.WriteString(fmt.Sprintf("Long:            %s\n",
		lipgloss.NewStyle().Foreground(successColor).Render(fmt.Sprintf("%d", longPositions))))
	content.WriteString(fmt.Sprintf("Short:           %s\n",
		lipgloss.NewStyle().Foreground(errorColor).Render(fmt.Sprintf("%d", shortPositions))))
	content.WriteString("\n")

	pnlStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	if totalUnrealizedPnL.IsNegative() {
		pnlStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	}

	content.WriteString(fmt.Sprintf("Unrealized PnL: %s\n",
		pnlStyle.Render("$"+totalUnrealizedPnL.StringFixed(2))))
	content.WriteString(fmt.Sprintf("Realized PnL:   $%s\n",
		totalRealizedPnL.StringFixed(2)))

	totalPnL := totalUnrealizedPnL.Add(totalRealizedPnL)
	totalPnLStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	if totalPnL.IsNegative() {
		totalPnLStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
	}
	content.WriteString(fmt.Sprintf("Total PnL:      %s\n",
		totalPnLStyle.Render("$"+totalPnL.StringFixed(2))))

	return boxStyle.Render(content.String())
}
