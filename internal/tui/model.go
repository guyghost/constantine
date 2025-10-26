package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/order"
	"github.com/guyghost/constantine/internal/risk"
	"github.com/guyghost/constantine/internal/strategy"
)

// Model represents the TUI application model
type Model struct {
	// Bot state
	aggregator           *exchanges.ExchangeMultiplexer
	strategyOrchestrator *strategy.StrategyOrchestrator
	orderManager         *order.Manager
	riskManager          *risk.Manager
	running              bool

	// UI state
	width      int
	height     int
	activeView View
	lastUpdate time.Time

	// Data
	tradingSymbols []string // Configured trading symbols
	currentSignals map[string]interface{}
	openOrders     []*exchanges.Order
	positions      []*order.ManagedPosition
	orderbook      *exchanges.OrderBook
	riskStats      *risk.Stats
	orderStats     *order.OrderStats
	messages       []string

	// Error handling
	lastError error
	errorTime time.Time
}

// View represents the active view
type View int

const (
	ViewDashboard View = iota
	ViewOrderBook
	ViewPositions
	ViewOrders
	ViewExchanges
	ViewSettings
	ViewSymbols
)

// NewModel creates a new TUI model
func NewModel(
	aggregator *exchanges.ExchangeMultiplexer,
	strategyOrchestrator *strategy.StrategyOrchestrator,
	orderManager *order.Manager,
	riskManager *risk.Manager,
	tradingSymbols []string,
) Model {
	return Model{
		aggregator:           aggregator,
		strategyOrchestrator: strategyOrchestrator,
		orderManager:         orderManager,
		riskManager:          riskManager,
		tradingSymbols:       tradingSymbols,
		activeView:           ViewDashboard,
		currentSignals:       make(map[string]interface{}),
		messages:             make([]string, 0),
		lastUpdate:           time.Now(),
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.EnterAltScreen,
	)
}

// Message types for the TUI
type tickMsg time.Time
type signalMsg struct {
	symbol string
	signal interface{}
}
type orderUpdateMsg *order.OrderUpdate
type positionUpdateMsg *order.ManagedPosition
type errorMsg error

// tickCmd sends periodic tick messages
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// SignalCmd creates a command for signal updates
func SignalCmd(symbol string, signal interface{}) tea.Cmd {
	return func() tea.Msg {
		return signalMsg{symbol: symbol, signal: signal}
	}
}

// OrderUpdateCmd creates a command for order updates
func OrderUpdateCmd(update *order.OrderUpdate) tea.Cmd {
	return func() tea.Msg {
		return orderUpdateMsg(update)
	}
}

// PositionUpdateCmd creates a command for position updates
func PositionUpdateCmd(position *order.ManagedPosition) tea.Cmd {
	return func() tea.Msg {
		return positionUpdateMsg(position)
	}
}

// ErrorCmd creates a command for errors
func ErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return errorMsg(err)
	}
}

// AddMessage adds a message to the message log
func (m *Model) AddMessage(message string) {
	timestamp := time.Now().Format("15:04:05")
	m.messages = append(m.messages, timestamp+" "+message)

	// Keep only last 100 messages
	if len(m.messages) > 100 {
		m.messages = m.messages[1:]
	}
}

// GetRecentMessages returns the most recent messages
func (m *Model) GetRecentMessages(count int) []string {
	if len(m.messages) <= count {
		return m.messages
	}
	return m.messages[len(m.messages)-count:]
}

// IsRunning returns whether the bot is running
func (m *Model) IsRunning() bool {
	return m.running
}

// SetRunning sets the running state
func (m *Model) SetRunning(running bool) {
	m.running = running
}

// UpdateDimensions updates the terminal dimensions
func (m *Model) UpdateDimensions(width, height int) {
	m.width = width
	m.height = height
}

// GetDimensions returns the terminal dimensions
func (m *Model) GetDimensions() (int, int) {
	return m.width, m.height
}

// SetActiveView sets the active view
func (m *Model) SetActiveView(view View) {
	m.activeView = view
}

// GetActiveView returns the active view
func (m *Model) GetActiveView() View {
	return m.activeView
}

// UpdateSignal updates the current signal for a symbol
func (m *Model) UpdateSignal(symbol string, signal interface{}) {
	if signal == nil {
		delete(m.currentSignals, symbol)
	} else {
		m.currentSignals[symbol] = signal
	}
	m.AddMessage(fmt.Sprintf("New signal for %s", symbol))
}

// UpdateOrders updates the open orders
func (m *Model) UpdateOrders(orders []*exchanges.Order) {
	m.openOrders = orders
}

// UpdatePositions updates the positions
func (m *Model) UpdatePositions(positions []*order.ManagedPosition) {
	m.positions = positions
}

// UpdateOrderBook updates the order book
func (m *Model) UpdateOrderBook(orderbook *exchanges.OrderBook) {
	m.orderbook = orderbook
}

// UpdateRiskStats updates the risk statistics
func (m *Model) UpdateRiskStats(stats *risk.Stats) {
	m.riskStats = stats
}

// UpdateOrderStats updates the order statistics
func (m *Model) UpdateOrderStats(stats *order.OrderStats) {
	m.orderStats = stats
}

// SetError sets the last error
func (m *Model) SetError(err error) {
	m.lastError = err
	m.errorTime = time.Now()
	if err != nil {
		m.AddMessage("Error: " + err.Error())
	}
}

// GetError returns the last error and when it occurred
func (m *Model) GetError() (error, time.Time) {
	return m.lastError, m.errorTime
}

// ClearError clears the last error
func (m *Model) ClearError() {
	m.lastError = nil
}
