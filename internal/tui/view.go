package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Color scheme
	primaryColor   = lipgloss.Color("#00D9FF")
	secondaryColor = lipgloss.Color("#7D56F4")
	successColor   = lipgloss.Color("#00FF87")
	errorColor     = lipgloss.Color("#FF5555")
	warningColor   = lipgloss.Color("#FFB86C")
	mutedColor     = lipgloss.Color("#6272A4")

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(secondaryColor).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)
)

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string

	switch m.activeView {
	case ViewDashboard:
		content = m.renderDashboard()
	case ViewOrderBook:
		content = m.renderOrderBook()
	case ViewPositions:
		content = m.renderPositions()
	case ViewOrders:
		content = m.renderOrders()
	case ViewSettings:
		content = m.renderSettings()
	}

	// Render header
	header := m.renderHeader()

	// Render status bar
	statusBar := m.renderStatusBar()

	// Render help
	help := m.renderHelp()

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		help,
		statusBar,
	)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	title := titleStyle.Render("⚡ SCALPING BOT")

	status := "STOPPED"
	statusStyle := errorStyle
	if m.running {
		status = "RUNNING"
		statusStyle = successStyle
	}

	statusText := statusStyle.Render(status)

	exchange := mutedStyle.Render(fmt.Sprintf("Exchange: %s", m.exchange.Name()))

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		"  ",
		statusText,
		"  ",
		exchange,
	)
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar() string {
	timestamp := time.Now().Format("15:04:05")

	var errorText string
	if m.lastError != nil && time.Since(m.errorTime) < 5*time.Second {
		errorText = " | " + errorStyle.Render("ERROR: "+m.lastError.Error())
	}

	status := fmt.Sprintf("%s%s", timestamp, errorText)
	return statusBarStyle.Width(m.width).Render(status)
}

// renderHelp renders the help text
func (m Model) renderHelp() string {
	helps := []string{
		"[1-5] Switch view",
		"[s] Start/Stop",
		"[r] Refresh",
		"[c] Clear error",
		"[q] Quit",
	}
	return helpStyle.Render(strings.Join(helps, " • "))
}

// renderDashboard renders the dashboard view
func (m Model) renderDashboard() string {
	// Summary
	summary := m.renderSummary()

	// Current signal
	signalBox := m.renderCurrentSignal()

	// Risk stats
	riskBox := m.renderRiskStats()

	// Recent messages
	messagesBox := m.renderMessages()

	// Arrange in grid
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, summary, "  ", signalBox)
	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, riskBox, "  ", messagesBox)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, "", middleRow)
}

// renderSummary renders the summary box
func (m Model) renderSummary() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Summary") + "\n\n")

	// Balance
	if m.riskStats != nil {
		balance := m.riskStats.CurrentBalance.StringFixed(2)
		content.WriteString(fmt.Sprintf("Balance:     %s\n", successStyle.Render("$"+balance)))

		pnl := m.riskStats.DailyPnL.StringFixed(2)
		pnlStyle := successStyle
		if m.riskStats.DailyPnL.IsNegative() {
			pnlStyle = errorStyle
		}
		content.WriteString(fmt.Sprintf("Daily P&L:   %s\n", pnlStyle.Render("$"+pnl)))

		drawdown := fmt.Sprintf("%.2f%%", m.riskStats.CurrentDrawdown)
		content.WriteString(fmt.Sprintf("Drawdown:    %s\n", warningStyle.Render(drawdown)))
	}

	// Positions
	posCount := len(m.positions)
	content.WriteString(fmt.Sprintf("Positions:   %d\n", posCount))

	// Orders
	orderCount := len(m.openOrders)
	content.WriteString(fmt.Sprintf("Open Orders: %d\n", orderCount))

	return boxStyle.Render(content.String())
}

// renderCurrentSignal renders the current signal
func (m Model) renderCurrentSignal() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Current Signal") + "\n\n")

	if m.currentSignal != nil && m.currentSignal.Type != "none" {
		sideStyle := successStyle
		if m.currentSignal.Side == "sell" {
			sideStyle = errorStyle
		}

		content.WriteString(fmt.Sprintf("Type:     %s\n", m.currentSignal.Type))
		content.WriteString(fmt.Sprintf("Side:     %s\n", sideStyle.Render(string(m.currentSignal.Side))))
		content.WriteString(fmt.Sprintf("Symbol:   %s\n", m.currentSignal.Symbol))
		content.WriteString(fmt.Sprintf("Price:    $%s\n", m.currentSignal.Price.StringFixed(2)))
		content.WriteString(fmt.Sprintf("Strength: %.1f%%\n", m.currentSignal.Strength*100))
		content.WriteString(fmt.Sprintf("Reason:   %s\n", mutedStyle.Render(m.currentSignal.Reason)))
	} else {
		content.WriteString(mutedStyle.Render("No active signal"))
	}

	return boxStyle.Render(content.String())
}

// renderRiskStats renders risk statistics
func (m Model) renderRiskStats() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Risk Management") + "\n\n")

	if m.riskStats != nil {
		winRate := fmt.Sprintf("%.1f%%", m.riskStats.WinRate)
		content.WriteString(fmt.Sprintf("Win Rate:    %s\n", successStyle.Render(winRate)))

		content.WriteString(fmt.Sprintf("Total Trades: %d\n", m.riskStats.TotalTrades))
		content.WriteString(fmt.Sprintf("Wins:        %s\n", successStyle.Render(fmt.Sprintf("%d", m.riskStats.WinningTrades))))
		content.WriteString(fmt.Sprintf("Losses:      %s\n", errorStyle.Render(fmt.Sprintf("%d", m.riskStats.LosingTrades))))
		content.WriteString(fmt.Sprintf("Consecutive: %d\n", m.riskStats.ConsecutiveLosses))

		pf := fmt.Sprintf("%.2f", m.riskStats.ProfitFactor)
		content.WriteString(fmt.Sprintf("Profit Factor: %s\n", successStyle.Render(pf)))
	} else {
		content.WriteString(mutedStyle.Render("No risk data available"))
	}

	return boxStyle.Render(content.String())
}

// renderMessages renders recent messages
func (m Model) renderMessages() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Messages") + "\n\n")

	messages := m.GetRecentMessages(5)
	if len(messages) == 0 {
		content.WriteString(mutedStyle.Render("No messages"))
	} else {
		for _, msg := range messages {
			content.WriteString(mutedStyle.Render(msg) + "\n")
		}
	}

	return boxStyle.Render(content.String())
}

// renderOrderBook renders the order book view
func (m Model) renderOrderBook() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Order Book") + "\n\n")

	if m.orderbook != nil {
		// Asks
		content.WriteString(errorStyle.Render("ASKS") + "\n")
		asks := m.orderbook.Asks
		if len(asks) > 10 {
			asks = asks[:10]
		}
		for i := len(asks) - 1; i >= 0; i-- {
			content.WriteString(fmt.Sprintf("  %s  %s\n",
				asks[i].Price.StringFixed(2),
				asks[i].Amount.StringFixed(4)))
		}

		content.WriteString("\n")

		// Bids
		content.WriteString(successStyle.Render("BIDS") + "\n")
		bids := m.orderbook.Bids
		if len(bids) > 10 {
			bids = bids[:10]
		}
		for _, bid := range bids {
			content.WriteString(fmt.Sprintf("  %s  %s\n",
				bid.Price.StringFixed(2),
				bid.Amount.StringFixed(4)))
		}
	} else {
		content.WriteString(mutedStyle.Render("No order book data"))
	}

	return boxStyle.Render(content.String())
}

// renderPositions renders the positions view
func (m Model) renderPositions() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Open Positions") + "\n\n")

	if len(m.positions) == 0 {
		content.WriteString(mutedStyle.Render("No open positions"))
	} else {
		for _, pos := range m.positions {
			sideStyle := successStyle
			if pos.Side == "short" {
				sideStyle = errorStyle
			}

			content.WriteString(fmt.Sprintf("%s %s\n",
				pos.Symbol,
				sideStyle.Render(string(pos.Side))))
			content.WriteString(fmt.Sprintf("  Entry:  $%s\n", pos.EntryPrice.StringFixed(2)))
			content.WriteString(fmt.Sprintf("  Amount: %s\n", pos.Amount.StringFixed(4)))
			content.WriteString(fmt.Sprintf("  PnL:    $%s\n", pos.UnrealizedPnL.StringFixed(2)))
			content.WriteString("\n")
		}
	}

	return boxStyle.Render(content.String())
}

// renderOrders renders the orders view
func (m Model) renderOrders() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Open Orders") + "\n\n")

	if len(m.openOrders) == 0 {
		content.WriteString(mutedStyle.Render("No open orders"))
	} else {
		for _, order := range m.openOrders {
			sideStyle := successStyle
			if order.Side == "sell" {
				sideStyle = errorStyle
			}

			content.WriteString(fmt.Sprintf("%s %s %s\n",
				order.Symbol,
				sideStyle.Render(string(order.Side)),
				order.Type))
			content.WriteString(fmt.Sprintf("  Price:  $%s\n", order.Price.StringFixed(2)))
			content.WriteString(fmt.Sprintf("  Amount: %s\n", order.Amount.StringFixed(4)))
			content.WriteString(fmt.Sprintf("  Status: %s\n", order.Status))
			content.WriteString("\n")
		}
	}

	return boxStyle.Render(content.String())
}

// renderSettings renders the settings view
func (m Model) renderSettings() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Settings") + "\n\n")
	content.WriteString(mutedStyle.Render("Settings view - coming soon"))

	return boxStyle.Render(content.String())
}
