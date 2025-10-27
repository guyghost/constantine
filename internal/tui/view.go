package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/strategy"
	"github.com/shopspring/decimal"
)

var (
	successColor = lipgloss.Color("#00FF87")
	errorColor   = lipgloss.Color("#FF5555")
	mutedColor   = lipgloss.Color("#6272A4")

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Bold(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#6272A4")).
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
	case ViewExchanges:
		content = m.renderExchanges()
	case ViewSettings:
		content = m.renderSettings()
	case ViewSymbols:
		content = m.renderSymbols()
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

	// Show selected symbols count
	selectedCount := len(m.GetSelectedSymbols())
	configuredCount := len(m.tradingSymbols)
	symbolsText := mutedStyle.Render(fmt.Sprintf("Symbols: %d/%d", selectedCount, configuredCount))

	// Show integrated engine status
	engine := m.GetIntegratedEngine()
	engineText := mutedStyle.Render("Engine: Initializing")
	if engine != nil {
		engineText = successStyle.Render("Engine: Active")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			title,
			"  ",
			statusText,
			"  ",
			symbolsText,
			"  ",
			engineText,
		),
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
		"[1-6] Switch view",
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

	// Selected symbols with scores
	selectedSymbolsBox := m.renderSelectedSymbols()

	// Current signal
	signalBox := m.renderCurrentSignal()

	// Recent messages
	messagesBox := m.renderMessages()

	// Arrange in grid - 2x3 layout
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, summary, "  ", selectedSymbolsBox)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, signalBox, "  ", messagesBox)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, "", bottomRow)
}

// renderSummary renders the summary box
func (m Model) renderSummary() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Summary") + "\n\n")

	// Get aggregated data
	data := m.aggregator.GetAggregatedData()

	// Total Balance
	totalBalance := data.TotalBalance.StringFixed(2)
	content.WriteString(fmt.Sprintf("Total Balance: %s\n", successStyle.Render("$"+totalBalance)))

	// Total PnL
	totalPnL := data.TotalPnL.StringFixed(2)
	pnlStyle := successStyle
	if data.TotalPnL.IsNegative() {
		pnlStyle = errorStyle
	}
	content.WriteString(fmt.Sprintf("Total P&L:     %s\n", pnlStyle.Render("$"+totalPnL)))

	// Exchange connections
	connectedCount := 0
	totalCount := len(data.Exchanges)
	for _, exchangeData := range data.Exchanges {
		if exchangeData.Connected {
			connectedCount++
		}
	}
	content.WriteString(fmt.Sprintf("Exchanges:     %d/%d connected\n", connectedCount, totalCount))

	// Selected symbols
	selectedSymbols := m.GetSelectedSymbols()
	configuredSymbols := len(m.tradingSymbols)
	selectionPercent := 0.0
	if configuredSymbols > 0 {
		selectionPercent = float64(len(selectedSymbols)) / float64(configuredSymbols) * 100
	}
	content.WriteString(fmt.Sprintf("Selected:      %d/%d symbols (%.0f%%)\n", len(selectedSymbols), configuredSymbols, selectionPercent))

	// Positions (from primary exchange for now)
	posCount := len(m.positions)
	content.WriteString(fmt.Sprintf("Positions:     %d\n", posCount))

	// Orders
	orderCount := len(m.openOrders)
	content.WriteString(fmt.Sprintf("Open Orders:   %d\n", orderCount))

	return boxStyle.Render(content.String())
}

// renderCurrentSignal renders the current signal
func (m Model) renderCurrentSignal() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Active Signals") + "\n\n")

	if len(m.currentSignals) == 0 {
		content.WriteString(mutedStyle.Render("No active signals"))
	} else {
		for symbol, sig := range m.currentSignals {
			// Type assertion to strategy.Signal
			if signal, ok := sig.(*strategy.Signal); ok {
				// Signal type icon
				signalIcon := "📈"
				if signal.Type == "exit" {
					signalIcon = "📉"
				}

				// Side styling
				sideStyle := successStyle
				sideIcon := "↗️"
				if signal.Side == exchanges.OrderSideSell {
					sideStyle = errorStyle
					sideIcon = "↘️"
				}

				content.WriteString(fmt.Sprintf("%s %s %s\n",
					signalIcon,
					symbol,
					sideStyle.Render(fmt.Sprintf("%s %s", sideIcon, string(signal.Side)))))

				content.WriteString(fmt.Sprintf("  Price: $%s\n", signal.Price.StringFixed(2)))
				content.WriteString(fmt.Sprintf("  Strength: %.1f%%\n", signal.Strength*100))
				content.WriteString(fmt.Sprintf("  Reason: %s\n", signal.Reason))
				content.WriteString("\n")
			}
		}
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

// renderSelectedSymbols renders the selected symbols with scores and weights
func (m Model) renderSelectedSymbols() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Selected Symbols") + "\n\n")

	selectedSymbols := m.GetSelectedSymbols()
	if len(selectedSymbols) == 0 {
		content.WriteString(mutedStyle.Render("⏳ Updating symbol selection...\n"))
		content.WriteString(mutedStyle.Render("This may take a few moments as data is fetched."))
	} else {
		// Create a slice to sort symbols for consistent display
		symbols := make([]string, 0, len(selectedSymbols))
		for symbol := range selectedSymbols {
			symbols = append(symbols, symbol)
		}

		// Sort for consistent display
		for i := 0; i < len(symbols)-1; i++ {
			for j := i + 1; j < len(symbols); j++ {
				if symbols[i] > symbols[j] {
					symbols[i], symbols[j] = symbols[j], symbols[i]
				}
			}
		}

		for _, symbol := range symbols {
			rankedSymbol := selectedSymbols[symbol]

			// Render symbol with score
			scorePercent := rankedSymbol.Score * 100
			scoreStyle := successStyle
			if scorePercent < 30 {
				scoreStyle = errorStyle
			} else if scorePercent < 60 {
				scoreStyle = mutedStyle
			}

			content.WriteString(fmt.Sprintf("📊 %s: %s (%.1f%%)\n",
				rankedSymbol.Symbol,
				scoreStyle.Render(fmt.Sprintf("%.2f", rankedSymbol.Score)),
				scorePercent))

			// Show metrics
			content.WriteString(fmt.Sprintf("  Potential: %s | Sharpe: %s\n",
				rankedSymbol.Potential.StringFixed(4),
				rankedSymbol.SharpeRatio.StringFixed(4)))

			// Show dynamic weights for this symbol if available
			if weights, ok := m.GetDynamicWeights(rankedSymbol.Symbol); ok {
				content.WriteString(fmt.Sprintf("  Weights - EMA: %.0f%% RSI: %.0f%% Vol: %.0f%% BB: %.0f%%\n",
					weights.EMA*100, weights.RSI*100, weights.Volume*100, weights.BB*100))
			}

			content.WriteString("\n")
		}

		// Show refresh time
		refreshDelta := time.Since(m.lastSymbolRefresh)
		content.WriteString(mutedStyle.Render(fmt.Sprintf("Last refresh: %vs ago", int(refreshDelta.Seconds()))))
	}

	return boxStyle.Render(content.String())
}

// renderTradingSymbols renders the trading symbols status (all configured symbols)
func (m Model) renderTradingSymbols() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Trading Symbols") + "\n\n")

	// Show all configured trading symbols with their signal status
	if len(m.tradingSymbols) == 0 {
		content.WriteString(mutedStyle.Render("No trading symbols configured"))
	} else {
		for _, symbol := range m.tradingSymbols {
			// Check if this symbol has an active signal
			sig, hasSignal := m.currentSignals[symbol]

			statusIcon := "⚫"
			statusStyle := mutedStyle
			statusText := "No signal"

			if hasSignal {
				if signal, ok := sig.(*strategy.Signal); ok {
					switch signal.Type {
					case "entry":
						if signal.Side == exchanges.OrderSideBuy {
							statusIcon = "🟢"
							statusStyle = successStyle
							statusText = fmt.Sprintf("BUY %.1f%%", signal.Strength*100)
						} else {
							statusIcon = "🔴"
							statusStyle = errorStyle
							statusText = fmt.Sprintf("SELL %.1f%%", signal.Strength*100)
						}
					case "exit":
						statusIcon = "⚪"
						statusStyle = mutedStyle
						statusText = "EXIT"
					default:
						statusIcon = "🟡"
						statusStyle = mutedStyle
						statusText = "UNKNOWN"
					}
				}
			}

			content.WriteString(fmt.Sprintf("%s %s: %s\n",
				statusIcon,
				symbol,
				statusStyle.Render(statusText)))
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

	// Get aggregated data
	data := m.aggregator.GetAggregatedData()

	allPositions := make([]*exchanges.Position, 0)

	// Collect positions from all exchanges
	for exchangeName, exchangeData := range data.Exchanges {
		for _, pos := range exchangeData.Positions {
			// Add exchange name to distinguish positions
			posWithExchange := pos
			posWithExchange.Symbol = fmt.Sprintf("%s (%s)", pos.Symbol, exchangeName)
			allPositions = append(allPositions, &posWithExchange)
		}
	}

	if len(allPositions) == 0 {
		content.WriteString(mutedStyle.Render("No open positions"))
	} else {
		for _, pos := range allPositions {
			sideStyle := successStyle
			if pos.Side == exchanges.OrderSideSell {
				sideStyle = errorStyle
			}

			content.WriteString(fmt.Sprintf("%s %s\n",
				pos.Symbol,
				sideStyle.Render(string(pos.Side))))
			content.WriteString(fmt.Sprintf("  Entry:  $%s\n", pos.EntryPrice.StringFixed(2)))
			content.WriteString(fmt.Sprintf("  Size:   %s\n", pos.Size.StringFixed(4)))
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

// renderExchanges renders the exchanges status view
func (m Model) renderExchanges() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Exchange Status") + "\n\n")

	// Get aggregated data
	data := m.aggregator.GetAggregatedData()

	for exchangeName, exchangeData := range data.Exchanges {
		status := "✗ DISCONNECTED"
		statusStyle := errorStyle
		if exchangeData.Connected {
			status = "✓ CONNECTED"
			statusStyle = successStyle
		}

		content.WriteString(fmt.Sprintf("%s: %s\n", exchangeName, statusStyle.Render(status)))

		if exchangeData.Error != nil {
			content.WriteString(fmt.Sprintf("  Error: %s\n", errorStyle.Render(exchangeData.Error.Error())))
		}

		// Show balances
		for _, balance := range exchangeData.Balances {
			if balance.Total.GreaterThan(decimal.Zero) {
				content.WriteString(fmt.Sprintf("  Balance: %s $%s\n",
					balance.Asset, successStyle.Render(balance.Total.StringFixed(2))))
			}
		}

		// Show positions count
		posCount := len(exchangeData.Positions)
		if posCount > 0 {
			content.WriteString(fmt.Sprintf("  Positions: %d\n", posCount))
		}

		// Show orders count
		orderCount := len(exchangeData.Orders)
		if orderCount > 0 {
			content.WriteString(fmt.Sprintf("  Orders: %d\n", orderCount))
		}

		content.WriteString("\n")
	}

	return boxStyle.Render(content.String())
}

// renderSymbols renders the symbols view with detailed information
func (m Model) renderSymbols() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Symbols Analysis") + "\n\n")

	selectedSymbols := m.GetSelectedSymbols()

	if len(selectedSymbols) == 0 {
		content.WriteString(mutedStyle.Render("No symbols selected yet"))
	} else {
		content.WriteString(successStyle.Render("SELECTED SYMBOLS FOR TRADING") + "\n\n")

		for _, rankedSymbol := range selectedSymbols {
			// Header with score
			scorePercent := rankedSymbol.Score * 100
			scoreStyle := successStyle
			if scorePercent < 30 {
				scoreStyle = errorStyle
			} else if scorePercent < 60 {
				scoreStyle = mutedStyle
			}

			content.WriteString(fmt.Sprintf("📊 %s\n", rankedSymbol.Symbol))
			content.WriteString(fmt.Sprintf("  Opportunity Score: %s (%.1f%%)\n",
				scoreStyle.Render(fmt.Sprintf("%.4f", rankedSymbol.Score)),
				scorePercent))

			// Metrics
			content.WriteString(fmt.Sprintf("  Gain Potential:    %s\n", rankedSymbol.Potential.StringFixed(6)))
			content.WriteString(fmt.Sprintf("  Risk Assessment:   %s\n", rankedSymbol.Risk.StringFixed(6)))
			content.WriteString(fmt.Sprintf("  Sharpe Ratio:      %s\n", rankedSymbol.SharpeRatio.StringFixed(6)))

			// Dynamic weights
			if weights, ok := m.GetDynamicWeights(rankedSymbol.Symbol); ok {
				content.WriteString("\n  Dynamic Weights:\n")
				content.WriteString(fmt.Sprintf("    EMA:            %.1f%%\n", weights.EMA*100))
				content.WriteString(fmt.Sprintf("    RSI:            %.1f%%\n", weights.RSI*100))
				content.WriteString(fmt.Sprintf("    Volume:         %.1f%%\n", weights.Volume*100))
				content.WriteString(fmt.Sprintf("    Bollinger Bands %.1f%%\n", weights.BB*100))
			}

			// Current signal if available
			if sig, hasSignal := m.currentSignals[rankedSymbol.Symbol]; hasSignal {
				if signal, ok := sig.(*strategy.Signal); ok {
					sideStyle := successStyle
					if signal.Side == exchanges.OrderSideSell {
						sideStyle = errorStyle
					}
					content.WriteString(fmt.Sprintf("\n  Current Signal: %s %s (%.1f%%)\n",
						signal.Type,
						sideStyle.Render(string(signal.Side)),
						signal.Strength*100))
				}
			}

			content.WriteString("\n")
		}

		// Show refresh time
		refreshDelta := time.Since(m.lastSymbolRefresh)
		content.WriteString(mutedStyle.Render(fmt.Sprintf("\nSymbol selection updated: %vs ago", int(refreshDelta.Seconds()))))
	}

	// Show all configured symbols count
	content.WriteString(fmt.Sprintf("\n%s (%d configured, %d selected)\n",
		headerStyle.Render("Summary"),
		len(m.tradingSymbols),
		len(selectedSymbols)))

	return boxStyle.Render(content.String())
}

// renderSettings renders the settings view with engine configuration
func (m Model) renderSettings() string {
	var content strings.Builder

	content.WriteString(headerStyle.Render("Strategy Configuration") + "\n\n")

	// Integrated Engine Status
	engine := m.GetIntegratedEngine()
	if engine == nil {
		content.WriteString(errorStyle.Render("⚠ Integrated Strategy Engine not initialized"))
	} else {
		content.WriteString(successStyle.Render("✓ Integrated Strategy Engine Active") + "\n\n")

		// Engine Features
		content.WriteString(headerStyle.Render("Features Enabled:") + "\n")
		content.WriteString("  ✓ Dynamic Indicator Weights\n")
		content.WriteString("  ✓ Intelligent Symbol Selection\n")
		content.WriteString("  ✓ Real-time Signal Generation\n")
		content.WriteString("  ✓ Automatic Risk Management\n\n")

		// Symbol Selection
		selectedSymbols := m.GetSelectedSymbols()
		content.WriteString(headerStyle.Render("Symbol Selection:") + "\n")
		content.WriteString(fmt.Sprintf("  Configured:  %d symbols\n", len(m.tradingSymbols)))
		content.WriteString(fmt.Sprintf("  Selected:    %d symbols (%.0f%%)\n",
			len(selectedSymbols),
			float64(len(selectedSymbols))/float64(len(m.tradingSymbols))*100))
		content.WriteString(fmt.Sprintf("  Last Update: %vs ago\n\n", int(time.Since(m.lastSymbolRefresh).Seconds())))

		// Weight Configuration
		content.WriteString(headerStyle.Render("Indicator Weights:") + "\n")
		content.WriteString("  Base Allocation:\n")
		content.WriteString("    EMA:            35%\n")
		content.WriteString("    RSI:            35%\n")
		content.WriteString("    Volume:         20%\n")
		content.WriteString("    Bollinger Band: 10%\n\n")
		content.WriteString(mutedStyle.Render("  * Weights adjust dynamically based on market volatility, trend, and momentum\n"))

		// Risk Management
		content.WriteString(headerStyle.Render("Risk Management:") + "\n")
		content.WriteString("  Stop Loss:      0.40%\n")
		content.WriteString("  Take Profit:    0.80%\n")
		content.WriteString("  Min Signal:     0.50 (50%)\n\n")

		// Refresh Interval
		content.WriteString(headerStyle.Render("Symbol Refresh:") + "\n")
		content.WriteString("  Interval:       30 seconds\n")
	}

	return boxStyle.Render(content.String())
}
