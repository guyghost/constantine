package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guyghost/constantine/internal/exchanges"
)

// WebSocketClient handles WebSocket connections for Coinbase
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
	var err error
	ws.conn, _, err = websocket.DefaultDialer.DialContext(ctx, ws.url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	// Start message handler
	go ws.handleMessages()

	return nil
}

// Close closes the WebSocket connection
func (ws *WebSocketClient) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.conn != nil {
		close(ws.done)
		return ws.conn.Close()
	}
	return nil
}

// handleMessages processes incoming WebSocket messages
func (ws *WebSocketClient) handleMessages() {
	defer func() {
		if ws.conn != nil {
			ws.conn.Close()
		}
	}()

	for {
		select {
		case <-ws.done:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				// Log error and attempt reconnect
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

	// TODO: Implement proper message routing based on Coinbase's protocol
	channel, ok := msg["channel"].(string)
	if !ok {
		return
	}

	switch channel {
	case "ticker":
		ws.handleTickerMessage(msg)
	case "level2":
		ws.handleOrderBookMessage(msg)
	case "market_trades":
		ws.handleTradeMessage(msg)
	}
}

// handleTickerMessage handles ticker updates
func (ws *WebSocketClient) handleTickerMessage(msg map[string]interface{}) {
	// TODO: Parse ticker data according to Coinbase format
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	events, ok := msg["events"].([]interface{})
	if !ok || len(events) == 0 {
		return
	}

	event := events[0].(map[string]interface{})
	symbol, ok := event["product_id"].(string)
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
	// TODO: Parse order book data according to Coinbase format
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	events, ok := msg["events"].([]interface{})
	if !ok || len(events) == 0 {
		return
	}

	event := events[0].(map[string]interface{})
	symbol, ok := event["product_id"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.orderbookCallbacks[symbol]; exists {
		orderbook := &exchanges.OrderBook{
			Symbol:    symbol,
			Bids:      []exchanges.Level{},
			Asks:      []exchanges.Level{},
			Timestamp: time.Now(),
		}
		callback(orderbook)
	}
}

// handleTradeMessage handles trade updates
func (ws *WebSocketClient) handleTradeMessage(msg map[string]interface{}) {
	// TODO: Parse trade data according to Coinbase format
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	events, ok := msg["events"].([]interface{})
	if !ok || len(events) == 0 {
		return
	}

	event := events[0].(map[string]interface{})
	symbol, ok := event["product_id"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.tradeCallbacks[symbol]; exists {
		trade := &exchanges.Trade{
			Symbol:    symbol,
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
		"type":        "subscribe",
		"product_ids": []string{symbol},
		"channel":     "ticker",
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
		"type":        "subscribe",
		"product_ids": []string{symbol},
		"channel":     "level2",
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
		"type":        "subscribe",
		"product_ids": []string{symbol},
		"channel":     "market_trades",
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
