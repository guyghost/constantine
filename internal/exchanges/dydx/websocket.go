package dydx

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/logger"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
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

	// Debug log for connection
	logger.Exchange("dydx").Debug("WebSocket connected", "url", ws.url)

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

	// Route messages based on dYdX v4 WebSocket protocol
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
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// dYdX v4 format: {"type": "channel_data", "channel": "v4_markets", "id": "BTC-USD", "contents": {...}}
	id, ok := msg["id"].(string)
	if !ok {
		return
	}

	contents, ok := msg["contents"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract ticker data from contents
	// dYdX provides: oraclePrice, priceChange24H, trades24H, volume24H, etc.
	ticker := &exchanges.Ticker{
		Symbol:    id,
		Timestamp: time.Now(),
	}

	// Parse oracle price (last price)
	if oraclePrice, ok := contents["oraclePrice"].(string); ok {
		if price, err := decimal.NewFromString(oraclePrice); err == nil {
			ticker.Last = price
			// Approximate bid/ask from oracle price (dYdX doesn't provide in ticker)
			ticker.Bid = price.Sub(decimal.NewFromFloat(0.5))
			ticker.Ask = price.Add(decimal.NewFromFloat(0.5))
		}
	}

	// Parse 24h volume
	if volume24h, ok := contents["trades24H"].(string); ok {
		if vol, err := decimal.NewFromString(volume24h); err == nil {
			ticker.Volume24h = vol
		}
	}

	// Invoke callback if registered for this symbol
	if callback, exists := ws.tickerCallbacks[id]; exists {
		callback(ticker)
	}
}

// handleOrderBookMessage handles order book updates
func (ws *WebSocketClient) handleOrderBookMessage(msg map[string]interface{}) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// dYdX v4 format: {"type": "channel_data", "channel": "v4_orderbook", "id": "BTC-USD", "contents": {...}}
	id, ok := msg["id"].(string)
	if !ok {
		return
	}

	contents, ok := msg["contents"].(map[string]interface{})
	if !ok {
		return
	}

	orderbook := &exchanges.OrderBook{
		Symbol:    id,
		Bids:      []exchanges.Level{},
		Asks:      []exchanges.Level{},
		Timestamp: time.Now(),
	}

	// Parse bids: [[price, size], ...]
	if bidsData, ok := contents["bids"].([]interface{}); ok {
		for _, bidData := range bidsData {
			if bidArray, ok := bidData.([]interface{}); ok && len(bidArray) >= 2 {
				priceStr, ok1 := bidArray[0].(string)
				sizeStr, ok2 := bidArray[1].(string)
				if ok1 && ok2 {
					price, err1 := decimal.NewFromString(priceStr)
					size, err2 := decimal.NewFromString(sizeStr)
					if err1 == nil && err2 == nil {
						orderbook.Bids = append(orderbook.Bids, exchanges.Level{
							Price:  price,
							Amount: size,
						})
					}
				}
			}
		}
	}

	// Parse asks: [[price, size], ...]
	if asksData, ok := contents["asks"].([]interface{}); ok {
		for _, askData := range asksData {
			if askArray, ok := askData.([]interface{}); ok && len(askArray) >= 2 {
				priceStr, ok1 := askArray[0].(string)
				sizeStr, ok2 := askArray[1].(string)
				if ok1 && ok2 {
					price, err1 := decimal.NewFromString(priceStr)
					size, err2 := decimal.NewFromString(sizeStr)
					if err1 == nil && err2 == nil {
						orderbook.Asks = append(orderbook.Asks, exchanges.Level{
							Price:  price,
							Amount: size,
						})
					}
				}
			}
		}
	}

	// Invoke callback if registered for this symbol
	if callback, exists := ws.orderbookCallbacks[id]; exists {
		callback(orderbook)
	}
}

// handleTradeMessage handles trade updates
func (ws *WebSocketClient) handleTradeMessage(msg map[string]interface{}) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// dYdX v4 format: {"type": "channel_data", "channel": "v4_trades", "id": "BTC-USD", "contents": {...}}
	id, ok := msg["id"].(string)
	if !ok {
		return
	}

	contents, ok := msg["contents"].(map[string]interface{})
	if !ok {
		return
	}

	// dYdX sends an array of recent trades
	if tradesData, ok := contents["trades"].([]interface{}); ok {
		for _, tradeData := range tradesData {
			if tradeMap, ok := tradeData.(map[string]interface{}); ok {
				trade := &exchanges.Trade{
					Symbol:    id,
					Timestamp: time.Now(),
				}

				// Parse price
				if priceStr, ok := tradeMap["price"].(string); ok {
					if price, err := decimal.NewFromString(priceStr); err == nil {
						trade.Price = price
					}
				}

				// Parse size/amount
				if sizeStr, ok := tradeMap["size"].(string); ok {
					if size, err := decimal.NewFromString(sizeStr); err == nil {
						trade.Amount = size
					}
				}

				// Parse side
				if sideStr, ok := tradeMap["side"].(string); ok {
					if sideStr == "BUY" || sideStr == "buy" {
						trade.Side = exchanges.OrderSideBuy
					} else {
						trade.Side = exchanges.OrderSideSell
					}
				}

				// Parse timestamp if available
				if createdAt, ok := tradeMap["createdAt"].(string); ok {
					if ts, err := time.Parse(time.RFC3339, createdAt); err == nil {
						trade.Timestamp = ts
					}
				}

				// Invoke callback if registered for this symbol
				if callback, exists := ws.tradeCallbacks[id]; exists {
					callback(trade)
				}
			}
		}
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
