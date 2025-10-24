package order

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	ordererrors "github.com/guyghost/constantine/internal/order/errors"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
)

const (
	// MaxFilledOrdersHistory limits the number of filled orders kept in memory
	MaxFilledOrdersHistory = 1000

	defaultAPICallTimeout = 5 * time.Second
)

// Manager manages orders and positions
type Manager struct {
	exchange  exchanges.Exchange
	orderBook *OrderBook
	mu        sync.RWMutex

	// Callbacks
	onOrderUpdate    func(*OrderUpdate)
	onPositionUpdate func(*ManagedPosition)
	onError          func(error)

	// Control
	running bool
	done    chan struct{}
}

// NewManager creates a new order manager
func NewManager(exchange exchanges.Exchange) *Manager {
	return &Manager{
		exchange:  exchange,
		orderBook: NewOrderBook(),
		done:      make(chan struct{}),
	}
}

// SetOrderUpdateCallback sets the callback for order updates
func (m *Manager) SetOrderUpdateCallback(callback func(*OrderUpdate)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onOrderUpdate = callback
}

// SetPositionUpdateCallback sets the callback for position updates
func (m *Manager) SetPositionUpdateCallback(callback func(*ManagedPosition)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onPositionUpdate = callback
}

// SetErrorCallback sets the callback for errors
func (m *Manager) SetErrorCallback(callback func(error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onError = callback
}

// Start starts the order manager
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("order manager already running")
	}
	if m.done == nil {
		m.done = make(chan struct{})
	} else {
		select {
		case <-m.done:
			m.done = make(chan struct{})
		default:
		}
	}
	doneCh := m.done
	m.running = true
	m.mu.Unlock()

	// Start monitoring loop
	go m.monitor(ctx, doneCh)

	return nil
}

// Stop stops the order manager
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	if m.done != nil {
		select {
		case <-m.done:
		default:
			close(m.done)
		}
		m.done = nil
	}
	m.running = false
	return nil
}

// PlaceOrder places a new order
func (m *Manager) PlaceOrder(ctx context.Context, req *OrderRequest) (*exchanges.Order, error) {
	if err := validateOrderRequest(req); err != nil {
		return nil, err
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAPICallTimeout)
	defer cancel()

	// Create order
	order := &exchanges.Order{
		ClientOrderID: fmt.Sprintf("order-%d", time.Now().UnixNano()),
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Amount:        req.Amount,
	}

	// Place order on exchange
	placedOrder, err := m.exchange.PlaceOrder(callCtx, order)
	if err != nil {
		m.emitError(ordererrors.New(ordererrors.OperationPlace, order.Symbol, err))
		return nil, err
	}

	// Store order
	m.mu.Lock()
	m.orderBook.OpenOrders[placedOrder.ID] = placedOrder
	m.mu.Unlock()

	// Emit order update
	m.emitOrderUpdate(&OrderUpdate{
		Order:     placedOrder,
		Event:     OrderEventCreated,
		Timestamp: time.Now(),
	})

	// Place stop loss and take profit if specified
	if !req.StopLoss.IsZero() {
		if _, err := m.placeStopLoss(ctx, placedOrder, req.StopLoss); err != nil {
			_ = m.CancelOrder(context.WithoutCancel(ctx), placedOrder.ID)
			return nil, ordererrors.New(ordererrors.OperationPlaceStopLoss, placedOrder.Symbol, err)
		}
	}
	if !req.TakeProfit.IsZero() {
		if _, err := m.placeTakeProfit(ctx, placedOrder, req.TakeProfit); err != nil {
			_ = m.CancelOrder(context.WithoutCancel(ctx), placedOrder.ID)
			return nil, ordererrors.New(ordererrors.OperationPlaceTakeProfit, placedOrder.Symbol, err)
		}
	}

	telemetry.RecordOrderPlaced(req.Symbol, string(req.Side))
	return placedOrder, nil
}

// CancelOrder cancels an existing order
func (m *Manager) CancelOrder(ctx context.Context, orderID string) error {
	callCtx, cancel := context.WithTimeout(ctx, defaultAPICallTimeout)
	defer cancel()

	if err := m.exchange.CancelOrder(callCtx, orderID); err != nil {
		m.emitError(ordererrors.New(ordererrors.OperationCancel, orderID, err))
		return err
	}

	// Update order book
	m.mu.Lock()
	if order, exists := m.orderBook.OpenOrders[orderID]; exists {
		order.Status = exchanges.OrderStatusCanceled
		delete(m.orderBook.OpenOrders, orderID)
		m.addFilledOrder(order)
	}
	m.mu.Unlock()

	// Emit order update
	m.emitOrderUpdate(&OrderUpdate{
		Order: &exchanges.Order{
			ID:     orderID,
			Status: exchanges.OrderStatusCanceled,
		},
		Event:     OrderEventCanceled,
		Timestamp: time.Now(),
	})

	return nil
}

// GetOpenOrders returns all open orders
func (m *Manager) GetOpenOrders() []*exchanges.Order {
	m.mu.RLock()
	defer m.mu.RUnlock()

	orders := make([]*exchanges.Order, 0, len(m.orderBook.OpenOrders))
	for _, order := range m.orderBook.OpenOrders {
		orders = append(orders, order)
	}
	return orders
}

// GetPositions returns all open positions
func (m *Manager) GetPositions() []*ManagedPosition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	positions := make([]*ManagedPosition, 0, len(m.orderBook.Positions))
	for _, position := range m.orderBook.Positions {
		positions = append(positions, position)
	}
	return positions
}

// GetPosition returns a specific position
func (m *Manager) GetPosition(symbol string) *ManagedPosition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.orderBook.Positions[symbol]
}

// ClosePosition closes a position
func (m *Manager) ClosePosition(ctx context.Context, symbol string) error {
	m.mu.RLock()
	position, exists := m.orderBook.Positions[symbol]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("position not found: %s", symbol)
	}

	// Determine order side (opposite of position side)
	var orderSide exchanges.OrderSide
	if position.Side == PositionSideLong {
		orderSide = exchanges.OrderSideSell
	} else {
		orderSide = exchanges.OrderSideBuy
	}

	// Place market order to close position
	req := &OrderRequest{
		Symbol:     symbol,
		Side:       orderSide,
		Type:       exchanges.OrderTypeMarket,
		Amount:     position.Amount,
		ReduceOnly: true,
	}

	order, err := m.PlaceOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to close position: %w", err)
	}

	// Update position
	m.mu.Lock()
	position.Status = PositionStatusClosed
	exitTime := time.Now()
	position.ExitTime = &exitTime
	position.ExitOrderID = order.ID
	m.mu.Unlock()

	// Emit position update
	m.emitPositionUpdate(position)

	return nil
}

// monitor monitors orders and positions
func (m *Manager) monitor(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			m.updateOrders(ctx)
			m.updatePositions(ctx)
		}
	}
}

// updateOrders updates the status of open orders
func (m *Manager) updateOrders(ctx context.Context) {
	m.mu.RLock()
	orderIDs := make([]string, 0, len(m.orderBook.OpenOrders))
	for id := range m.orderBook.OpenOrders {
		orderIDs = append(orderIDs, id)
	}
	m.mu.RUnlock()

	for _, orderID := range orderIDs {
		callCtx, cancel := context.WithTimeout(ctx, defaultAPICallTimeout)
		order, err := m.exchange.GetOrder(callCtx, orderID)
		cancel()
		if err != nil {
			continue
		}

		m.mu.Lock()
		oldOrder := m.orderBook.OpenOrders[orderID]
		m.mu.Unlock()

		// Check if status changed
		if oldOrder != nil && order.Status != oldOrder.Status {
			m.handleOrderStatusChange(order, oldOrder)
		}
	}
}

// handleOrderStatusChange handles order status changes
func (m *Manager) handleOrderStatusChange(newOrder, oldOrder *exchanges.Order) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var event OrderEvent

	switch newOrder.Status {
	case exchanges.OrderStatusFilled:
		event = OrderEventFilled
		delete(m.orderBook.OpenOrders, newOrder.ID)
		m.addFilledOrder(newOrder)

		// Update or create position
		m.handleFilledOrder(newOrder)

	case exchanges.OrderStatusPartially:
		event = OrderEventPartiallyFilled
		m.orderBook.OpenOrders[newOrder.ID] = newOrder

	case exchanges.OrderStatusCanceled:
		event = OrderEventCanceled
		delete(m.orderBook.OpenOrders, newOrder.ID)
	}

	// Emit order update
	m.emitOrderUpdate(&OrderUpdate{
		Order:     newOrder,
		Event:     event,
		Timestamp: time.Now(),
	})
}

// handleFilledOrder handles a filled order and updates positions
func (m *Manager) handleFilledOrder(order *exchanges.Order) {
	position, exists := m.orderBook.Positions[order.Symbol]

	if !exists {
		// Create new position
		var side PositionSide
		if order.Side == exchanges.OrderSideBuy {
			side = PositionSideLong
		} else {
			side = PositionSideShort
		}

		position = &ManagedPosition{
			ID:            fmt.Sprintf("pos-%d", time.Now().UnixNano()),
			Symbol:        order.Symbol,
			Side:          side,
			EntryPrice:    order.Price,
			CurrentPrice:  order.Price,
			Amount:        order.Filled,
			Leverage:      decimal.NewFromInt(1),
			UnrealizedPnL: decimal.Zero,
			RealizedPnL:   decimal.Zero,
			EntryTime:     time.Now(),
			Status:        PositionStatusOpen,
			EntryOrderID:  order.ID,
		}

		m.orderBook.Positions[order.Symbol] = position
		m.emitPositionUpdate(position)
	} else {
		// Update existing position or close it
		if (position.Side == PositionSideLong && order.Side == exchanges.OrderSideSell) ||
			(position.Side == PositionSideShort && order.Side == exchanges.OrderSideBuy) {
			// Closing position
			pnl := m.calculatePnL(position, order.Price)
			position.RealizedPnL = position.RealizedPnL.Add(pnl)
			position.Status = PositionStatusClosed
			exitTime := time.Now()
			position.ExitTime = &exitTime
			position.ExitOrderID = order.ID

			delete(m.orderBook.Positions, order.Symbol)
			m.emitPositionUpdate(position)
		}
	}
}

// calculatePnL calculates profit/loss for a position
func (m *Manager) calculatePnL(position *ManagedPosition, exitPrice decimal.Decimal) decimal.Decimal {
	priceDiff := exitPrice.Sub(position.EntryPrice)
	if position.Side == PositionSideShort {
		priceDiff = priceDiff.Neg()
	}
	leverage := position.Leverage
	if leverage.IsZero() {
		leverage = decimal.NewFromInt(1)
	}
	return priceDiff.Mul(position.Amount).Mul(leverage)
}

// updatePositions updates position information
func (m *Manager) updatePositions(ctx context.Context) {
	callCtx, cancel := context.WithTimeout(ctx, defaultAPICallTimeout)
	defer cancel()

	positions, err := m.exchange.GetPositions(callCtx)
	if err != nil {
		return
	}

	for _, exchangePos := range positions {
		m.mu.Lock()
		managedPos, exists := m.orderBook.Positions[exchangePos.Symbol]
		if exists {
			managedPos.CurrentPrice = exchangePos.MarkPrice
			managedPos.UnrealizedPnL = exchangePos.UnrealizedPnL
		}
		m.mu.Unlock()
	}
}

// placeStopLoss places a stop loss order
func (m *Manager) placeStopLoss(ctx context.Context, order *exchanges.Order, stopLoss decimal.Decimal) (*exchanges.Order, error) {
	if stopLoss.IsZero() {
		return nil, nil
	}

	if stopLoss.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("stop loss price must be positive")
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAPICallTimeout)
	defer cancel()

	// Determine stop loss side (opposite of entry order)
	stopSide := exchanges.OrderSideSell
	if order.Side == exchanges.OrderSideSell {
		stopSide = exchanges.OrderSideBuy
	}

	// Create stop loss order
	stopOrder := &exchanges.Order{
		Symbol:    order.Symbol,
		Side:      stopSide,
		Type:      exchanges.OrderTypeStopLimit,
		Amount:    order.Amount,
		Price:     stopLoss,
		StopPrice: stopLoss,
		Status:    exchanges.OrderStatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Place the stop loss order
	placedOrder, err := m.exchange.PlaceOrder(callCtx, stopOrder)
	if err != nil {
		m.emitError(ordererrors.New(ordererrors.OperationPlaceStopLoss, order.Symbol, err))
		return nil, err
	}

	// Update order book
	m.mu.Lock()
	m.orderBook.OpenOrders[placedOrder.ID] = placedOrder

	// Link to the original position if exists
	for _, pos := range m.orderBook.Positions {
		if pos.Symbol == order.Symbol && pos.EntryOrderID == order.ID {
			pos.StopLossOrderID = placedOrder.ID
			break
		}
	}
	m.mu.Unlock()

	// Emit order update
	m.emitOrderUpdate(&OrderUpdate{
		Order:     placedOrder,
		Event:     OrderEventCreated,
		Timestamp: time.Now(),
	})

	telemetry.RecordStopLossPlaced(order.Symbol)
	return placedOrder, nil
}

// placeTakeProfit places a take profit order
func (m *Manager) placeTakeProfit(ctx context.Context, order *exchanges.Order, takeProfit decimal.Decimal) (*exchanges.Order, error) {
	if takeProfit.IsZero() {
		return nil, nil
	}

	if takeProfit.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("take profit price must be positive")
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAPICallTimeout)
	defer cancel()

	// Determine take profit side (opposite of entry order)
	takeProfitSide := exchanges.OrderSideSell
	if order.Side == exchanges.OrderSideSell {
		takeProfitSide = exchanges.OrderSideBuy
	}

	// Create take profit order as limit order
	takeProfitOrder := &exchanges.Order{
		Symbol:    order.Symbol,
		Side:      takeProfitSide,
		Type:      exchanges.OrderTypeLimit,
		Amount:    order.Amount,
		Price:     takeProfit,
		Status:    exchanges.OrderStatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Place the take profit order
	placedOrder, err := m.exchange.PlaceOrder(callCtx, takeProfitOrder)
	if err != nil {
		m.emitError(ordererrors.New(ordererrors.OperationPlaceTakeProfit, order.Symbol, err))
		return nil, err
	}

	// Update order book
	m.mu.Lock()
	m.orderBook.OpenOrders[placedOrder.ID] = placedOrder

	// Link to the original position if exists
	for _, pos := range m.orderBook.Positions {
		if pos.Symbol == order.Symbol && pos.EntryOrderID == order.ID {
			pos.TakeProfitOrderID = placedOrder.ID
			break
		}
	}
	m.mu.Unlock()

	// Emit order update
	m.emitOrderUpdate(&OrderUpdate{
		Order:     placedOrder,
		Event:     OrderEventCreated,
		Timestamp: time.Now(),
	})

	telemetry.RecordTakeProfitPlaced(order.Symbol)
	return placedOrder, nil
}

// addFilledOrder adds an order to the filled orders list with size limit
func (m *Manager) addFilledOrder(order *exchanges.Order) {
	if len(m.orderBook.FilledOrders) >= MaxFilledOrdersHistory {
		trimmed := make([]*exchanges.Order, MaxFilledOrdersHistory-1)
		copy(trimmed, m.orderBook.FilledOrders[len(m.orderBook.FilledOrders)-(MaxFilledOrdersHistory-1):])
		m.orderBook.FilledOrders = trimmed
	}
	m.orderBook.FilledOrders = append(m.orderBook.FilledOrders, order)
}

// emitOrderUpdate emits an order update
func (m *Manager) emitOrderUpdate(update *OrderUpdate) {
	m.mu.RLock()
	callback := m.onOrderUpdate
	m.mu.RUnlock()

	if callback != nil {
		safeInvoke(func() { callback(update) })
	}
}

// emitPositionUpdate emits a position update
func (m *Manager) emitPositionUpdate(position *ManagedPosition) {
	m.mu.RLock()
	callback := m.onPositionUpdate
	m.mu.RUnlock()

	if callback != nil {
		safeInvoke(func() { callback(position) })
	}
}

// emitError emits an error
func (m *Manager) emitError(err error) {
	m.mu.RLock()
	callback := m.onError
	m.mu.RUnlock()

	if callback != nil {
		safeInvoke(func() { callback(err) })
	}
}

func safeInvoke(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			telemetry.RecordCallbackPanic()
		}
	}()
	fn()
}

func validateOrderRequest(req *OrderRequest) error {
	if req == nil {
		return ordererrors.New(ordererrors.OperationValidate, "", errors.New("order request is nil"))
	}
	if req.Symbol == "" {
		return ordererrors.New(ordererrors.OperationValidate, "", errors.New("symbol is required"))
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return ordererrors.New(ordererrors.OperationValidate, req.Symbol, errors.New("amount must be positive"))
	}
	switch req.Type {
	case exchanges.OrderTypeLimit, exchanges.OrderTypeStopLimit:
		if req.Price.LessThanOrEqual(decimal.Zero) {
			return ordererrors.New(ordererrors.OperationValidate, req.Symbol, errors.New("price must be positive for limit orders"))
		}
	}
	return nil
}

// GetStats returns order statistics
func (m *Manager) GetStats() *OrderStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &OrderStats{
		TotalOrders:  len(m.orderBook.FilledOrders) + len(m.orderBook.OpenOrders),
		FilledOrders: len(m.orderBook.FilledOrders),
		TotalVolume:  decimal.Zero,
		TotalFees:    decimal.Zero,
	}

	for _, order := range m.orderBook.FilledOrders {
		stats.TotalVolume = stats.TotalVolume.Add(order.Filled.Mul(order.Price))
		if order.Status == exchanges.OrderStatusCanceled {
			stats.CanceledOrders++
		}
	}

	if stats.TotalOrders > 0 {
		stats.SuccessRate = float64(stats.FilledOrders) / float64(stats.TotalOrders)
	}

	return stats
}
