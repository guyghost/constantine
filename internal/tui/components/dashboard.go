package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/shopspring/decimal"
)

var (
	successColor = lipgloss.Color("#00FF87")
	errorColor   = lipgloss.Color("#FF5555")
	warningColor = lipgloss.Color("#FFB86C")
	mutedColor   = lipgloss.Color("#6272A4")

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2)
)

// RenderBalanceCard renders the account balance card
func RenderBalanceCard(balance, dailyPnL, totalPnL decimal.Decimal) string {
	var content strings.Builder

	content.WriteString("💰 Account Balance\n\n")

	balanceStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	content.WriteString(fmt.Sprintf("Balance:    %s\n", balanceStyle.Render("$"+balance.StringFixed(2))))

	pnlStyle := lipgloss.NewStyle().Foreground(successColor)
	if dailyPnL.IsNegative() {
		pnlStyle = lipgloss.NewStyle().Foreground(errorColor)
	}
	content.WriteString(fmt.Sprintf("Daily P&L:  %s\n", pnlStyle.Render("$"+dailyPnL.StringFixed(2))))

	totalPnLStyle := lipgloss.NewStyle().Foreground(successColor)
	if totalPnL.IsNegative() {
		totalPnLStyle = lipgloss.NewStyle().Foreground(errorColor)
	}
	content.WriteString(fmt.Sprintf("Total P&L:  %s\n", totalPnLStyle.Render("$"+totalPnL.StringFixed(2))))

	return boxStyle.Render(content.String())
}

// RenderStatsCard renders trading statistics card
func RenderStatsCard(stats *risk.Stats) string {
	var content strings.Builder

	content.WriteString("📊 Trading Stats\n\n")

	if stats != nil {
		winRateStyle := lipgloss.NewStyle().Foreground(successColor)
		if stats.WinRate < 50 {
			winRateStyle = lipgloss.NewStyle().Foreground(warningColor)
		}

		content.WriteString(fmt.Sprintf("Total Trades:  %d\n", stats.TotalTrades))
		content.WriteString(fmt.Sprintf("Win Rate:      %s\n", winRateStyle.Render(fmt.Sprintf("%.1f%%", stats.WinRate))))
		content.WriteString(fmt.Sprintf("Wins/Losses:   %d/%d\n", stats.WinningTrades, stats.LosingTrades))
		content.WriteString(fmt.Sprintf("Profit Factor: %.2f\n", stats.ProfitFactor))

		drawdownStyle := lipgloss.NewStyle().Foreground(warningColor)
		if stats.CurrentDrawdown.GreaterThan(decimal.NewFromFloat(5)) {
			drawdownStyle = lipgloss.NewStyle().Foreground(errorColor)
		}
		content.WriteString(fmt.Sprintf("Drawdown:      %s\n", drawdownStyle.Render(stats.CurrentDrawdown.StringFixed(2)+"%")))
	} else {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		content.WriteString(mutedStyle.Render("No statistics available"))
	}

	return boxStyle.Render(content.String())
}

// RenderActivityCard renders recent activity card
func RenderActivityCard(messages []string) string {
	var content strings.Builder

	content.WriteString("📝 Recent Activity\n\n")

	if len(messages) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		content.WriteString(mutedStyle.Render("No recent activity"))
	} else {
		mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)
		for _, msg := range messages {
			content.WriteString(mutedStyle.Render("• "+msg) + "\n")
		}
	}

	return boxStyle.Render(content.String())
}

// RenderRiskCard renders risk management card
func RenderRiskCard(canTrade bool, reason string, consecutiveLosses int, tradesExecuted int, maxTrades int) string {
	var content strings.Builder

	content.WriteString("🛡️  Risk Management\n\n")

	statusStyle := lipgloss.NewStyle().Foreground(successColor).Bold(true)
	statusText := "ALLOWED"
	if !canTrade {
		statusStyle = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
		statusText = "BLOCKED"
	}

	content.WriteString(fmt.Sprintf("Status:     %s\n", statusStyle.Render(statusText)))

	if !canTrade && reason != "" {
		warningStyle := lipgloss.NewStyle().Foreground(warningColor)
		content.WriteString(fmt.Sprintf("Reason:     %s\n", warningStyle.Render(reason)))
	}

	lossStyle := lipgloss.NewStyle().Foreground(mutedColor)
	if consecutiveLosses > 0 {
		lossStyle = lipgloss.NewStyle().Foreground(warningColor)
	}
	if consecutiveLosses >= 3 {
		lossStyle = lipgloss.NewStyle().Foreground(errorColor)
	}

	content.WriteString(fmt.Sprintf("Consecutive Losses: %s\n", lossStyle.Render(fmt.Sprintf("%d", consecutiveLosses))))
	content.WriteString(fmt.Sprintf("Trades Today:       %d/%d\n", tradesExecuted, maxTrades))

	return boxStyle.Render(content.String())
}
