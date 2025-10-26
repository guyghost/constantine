package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.UpdateDimensions(msg.Width, msg.Height)
		return m, nil

	case tickMsg:
		// Fetch latest data
		cmds = append(cmds, m.fetchData())

		// Schedule next tick
		cmds = append(cmds, tickCmd())
		return m, tea.Batch(cmds...)

	case signalMsg:
		if msg.signal != nil {
			m.UpdateSignal(msg.symbol, msg.signal)
		}
		return m, nil

	case orderUpdateMsg:
		if msg != nil {
			m.AddMessage("Order update: " + string(msg.Event))

			// Refresh orders
			orders := m.orderManager.GetOpenOrders()
			m.UpdateOrders(orders)

			// Refresh order stats
			stats := m.orderManager.GetStats()
			m.UpdateOrderStats(stats)
		}
		return m, nil

	case positionUpdateMsg:
		if msg != nil {
			m.AddMessage("Position update: " + msg.Symbol)

			// Refresh positions
			positions := m.orderManager.GetPositions()
			m.UpdatePositions(positions)
		}
		return m, nil

	case errorMsg:
		m.SetError(msg)
		return m, nil
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		// Quit the application
		return m, tea.Quit

	case "1":
		// Switch to dashboard view
		m.SetActiveView(ViewDashboard)
		return m, nil

	case "2":
		// Switch to order book view
		m.SetActiveView(ViewOrderBook)
		return m, nil

	case "3":
		// Switch to positions view
		m.SetActiveView(ViewPositions)
		return m, nil

	case "4":
		// Switch to orders view
		m.SetActiveView(ViewOrders)
		return m, nil

	case "5":
		// Switch to exchanges view
		m.SetActiveView(ViewExchanges)
		return m, nil

	case "6":
		// Switch to settings view
		m.SetActiveView(ViewSettings)
		return m, nil

	case "s":
		// Start/stop the bot
		if m.IsRunning() {
			m.SetRunning(false)
			if m.strategy != nil {
				m.strategy.Stop()
			}
			m.AddMessage("Bot stopped")
		} else {
			m.SetRunning(true)
			m.AddMessage("Bot started")
		}
		return m, nil

	case "c":
		// Clear error
		m.ClearError()
		return m, nil

	case "r":
		// Refresh data
		return m, m.fetchData()
	}

	return m, nil
}

// fetchData fetches latest data from the bot
func (m Model) fetchData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Refresh data from all exchanges
		if err := m.aggregator.RefreshData(ctx); err != nil {
			m.SetError(err)
			return nil
		}

		// Update positions from order manager (primary exchange)
		positions := m.orderManager.GetPositions()
		m.UpdatePositions(positions)

		// Update orders from order manager
		orders := m.orderManager.GetOpenOrders()
		m.UpdateOrders(orders)

		// Update risk stats
		if m.riskManager != nil {
			stats := m.riskManager.GetStats()
			m.UpdateRiskStats(stats)
		}

		// Update order stats
		if m.orderManager != nil {
			stats := m.orderManager.GetStats()
			m.UpdateOrderStats(stats)
		}

		return nil
	}
}
