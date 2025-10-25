package dydx

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/telemetry"
)

// WebSocketClient handles WebSocket connections for dYdX
type WebSocketClient struct {
	url       string
	apiKey    string
	apiSecret string
	conn      *websocket.Conn
	mu        sync.RWMutex

	tickerCallbacks    map[string]func(*exchanges.Ticker)
	orderbookCallbacks map[string]func(*exchanges.OrderBook)
	tradeCallbacks     map[string]func(*exchanges.Trade)

	done chan struct{}
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(url, apiKey, apiSecret string) *WebSocketClient {
	return &WebSocketClient{
		url:                url,
		apiKey:             apiKey,
		apiSecret:          apiSecret,
		tickerCallbacks:    make(map[string]func(*exchanges.Ticker)),
		orderbookCallbacks: make(map[string]func(*exchanges.OrderBook)),
		tradeCallbacks:     make(map[string]func(*exchanges.Trade)),
		done:               make(chan struct{}),
	}
}

// Connect establishes WebSocket connection
func (ws *WebSocketClient) Connect(ctx context.Context) error {
	ws.mu.Lock()
	if ws.done == nil {
		ws.done = make(chan struct{})
	} else {
		select {
		case <-ws.done:
			ws.done = make(chan struct{})
		default:
		}
	}
	done := ws.done
	ws.mu.Unlock()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, ws.url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.mu.Unlock()

	// Start message handler
	go ws.handleMessages(done)

	return nil
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn != nil {
		if ws.done != nil {
			select {
			case <-ws.done:
			default:
				close(ws.done)
			}
			ws.done = nil
		}
		err := ws.conn.Close()
		ws.conn = nil
		return err
	}
	return nil
}

// handleMessages processes incoming WebSocket messages
func (ws *WebSocketClient) handleMessages(done <-chan struct{}) {
	defer func() {
		ws.mu.Lock()
		if ws.conn != nil {
			ws.conn.Close()
			ws.conn = nil
		}
		ws.mu.Unlock()
	}()

	for {
		select {
		case <-done:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				// Log error and attempt reconnect
				telemetry.RecordWebSocketReconnect("dydx")
				time.Sleep(5 * time.Second)
				continue
			}

			ws.processMessage(message)
		}
	}
}

// processMessage processes a single message
func (ws *WebSocketClient) processMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	// TODO: Implement proper message routing based on dYdX's protocol
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "subscribed":
		// Handle subscription confirmation
		return
	case "channel_data":
		channel, ok := msg["channel"].(string)
		if !ok {
			return
		}

		switch channel {
		case "v4_markets":
			ws.handleTickerMessage(msg)
		case "v4_orderbook":
			ws.handleOrderBookMessage(msg)
		case "v4_trades":
			ws.handleTradeMessage(msg)
		}
	}
}

// handleTickerMessage handles ticker updates
func (ws *WebSocketClient) handleTickerMessage(msg map[string]interface{}) {
	// TODO: Parse ticker data according to dYdX format
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	contents, ok := msg["contents"].(map[string]interface{})
	if !ok {
		return
	}

	symbol, ok := contents["ticker"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.tickerCallbacks[symbol]; exists {
		ticker := &exchanges.Ticker{
			Symbol:    symbol,
			Timestamp: time.Now(),
		}
		callback(ticker)
	}
}

// handleOrderBookMessage handles order book updates
func (ws *WebSocketClient) handleOrderBookMessage(msg map[string]interface{}) {
	// TODO: Parse order book data according to dYdX format
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	id, ok := msg["id"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.orderbookCallbacks[id]; exists {
		orderbook := &exchanges.OrderBook{
			Symbol:    id,
			Bids:      []exchanges.Level{},
			Asks:      []exchanges.Level{},
			Timestamp: time.Now(),
		}
		callback(orderbook)
	}
}

// handleTradeMessage handles trade updates
func (ws *WebSocketClient) handleTradeMessage(msg map[string]interface{}) {
	// TODO: Parse trade data according to dYdX format
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	id, ok := msg["id"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.tradeCallbacks[id]; exists {
		trade := &exchanges.Trade{
			Symbol:    id,
			Timestamp: time.Now(),
		}
		callback(trade)
	}
}

// SubscribeTicker subscribes to ticker updates
func (ws *WebSocketClient) SubscribeTicker(ctx context.Context, symbol string, callback func(*exchanges.Ticker)) error {
	ws.mu.Lock()
	ws.tickerCallbacks[symbol] = callback
	ws.mu.Unlock()

	// Send subscription message
	sub := map[string]interface{}{
		"type":    "subscribe",
		"channel": "v4_markets",
		"id":      symbol,
	}

	return ws.sendMessage(sub)
}

// SubscribeOrderBook subscribes to order book updates
func (ws *WebSocketClient) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*exchanges.OrderBook)) error {
	ws.mu.Lock()
	ws.orderbookCallbacks[symbol] = callback
	ws.mu.Unlock()

	// Send subscription message
	sub := map[string]interface{}{
		"type":    "subscribe",
		"channel": "v4_orderbook",
		"id":      symbol,
	}

	return ws.sendMessage(sub)
}

// SubscribeTrades subscribes to trade updates
func (ws *WebSocketClient) SubscribeTrades(ctx context.Context, symbol string, callback func(*exchanges.Trade)) error {
	ws.mu.Lock()
	ws.tradeCallbacks[symbol] = callback
	ws.mu.Unlock()

	// Send subscription message
	sub := map[string]interface{}{
		"type":    "subscribe",
		"channel": "v4_trades",
		"id":      symbol,
	}

	return ws.sendMessage(sub)
}

// sendMessage sends a message through the WebSocket
func (ws *WebSocketClient) sendMessage(msg interface{}) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn == nil {
		return fmt.Errorf("websocket not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return ws.conn.WriteMessage(websocket.TextMessage, data)
}
