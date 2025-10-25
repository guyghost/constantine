package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
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

	backoff := time.Second
	maxBackoff := time.Minute
	const maxRetries = 10
	retries := 0

	for {
		select {
		case <-done:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				retries++
				if retries >= maxRetries {
					// Max retries exceeded, exit
					return
				}

				// Record reconnect attempt
				telemetry.RecordWebSocketReconnect("coinbase")

				// Exponential backoff with context check
				select {
				case <-ws.done:
					return
				case <-time.After(backoff):
					backoff = min(backoff*2, maxBackoff)
					continue
				}
			}

			// Reset backoff on successful read
			backoff = time.Second
			retries = 0
			ws.processMessage(message)
		}
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// processMessage processes a single message
func (ws *WebSocketClient) processMessage(message []byte) {
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	// Route messages based on Coinbase Advanced Trade WebSocket protocol
	// Supported channels: ticker, level2, market_trades
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
func (ws *WebSocketClient) handleTickerMessage(msg map[string]any) {
	events, ok := msg["events"].([]interface{})
	if !ok || len(events) == 0 {
		return
	}

	event, ok := events[0].(map[string]interface{})
	if !ok {
		return
	}

	symbol, ok := event["product_id"].(string)
	if !ok {
		return
	}

	// Get callback outside the data parsing
	ws.mu.RLock()
	callback, exists := ws.tickerCallbacks[symbol]
	ws.mu.RUnlock()

	if exists {
		// Parse ticker data
		var bid, ask, last decimal.Decimal
		var volume24h decimal.Decimal

		if bestBid, ok := event["best_bid"].(string); ok {
			bid, _ = decimal.NewFromString(bestBid)
		}
		if bestAsk, ok := event["best_ask"].(string); ok {
			ask, _ = decimal.NewFromString(bestAsk)
		}
		if price, ok := event["price"].(string); ok {
			last, _ = decimal.NewFromString(price)
		}
		if size, ok := event["size"].(string); ok {
			volume24h, _ = decimal.NewFromString(size)
		}

		ticker := &exchanges.Ticker{
			Symbol:    symbol,
			Bid:       bid,
			Ask:       ask,
			Last:      last,
			Volume24h: volume24h,
			Timestamp: time.Now(),
		}

		// Execute callback outside the lock
		callback(ticker)
	}
}

// handleOrderBookMessage handles order book updates
func (ws *WebSocketClient) handleOrderBookMessage(msg map[string]any) {
	events, ok := msg["events"].([]interface{})
	if !ok || len(events) == 0 {
		return
	}

	event, ok := events[0].(map[string]interface{})
	if !ok {
		return
	}

	symbol, ok := event["product_id"].(string)
	if !ok {
		return
	}

	// Get callback outside the data parsing
	ws.mu.RLock()
	callback, exists := ws.orderbookCallbacks[symbol]
	ws.mu.RUnlock()

	if exists {
		// Parse order book data
		var bids, asks []exchanges.Level

		if bidsData, ok := event["bids"].([]interface{}); ok {
			for _, bid := range bidsData {
				if bidSlice, ok := bid.([]interface{}); ok && len(bidSlice) >= 2 {
					if priceStr, ok := bidSlice[0].(string); ok {
						if sizeStr, ok := bidSlice[1].(string); ok {
							price, _ := decimal.NewFromString(priceStr)
							size, _ := decimal.NewFromString(sizeStr)
							bids = append(bids, exchanges.Level{Price: price, Amount: size})
						}
					}
				}
			}
		}

		if asksData, ok := event["asks"].([]interface{}); ok {
			for _, ask := range asksData {
				if askSlice, ok := ask.([]interface{}); ok && len(askSlice) >= 2 {
					if priceStr, ok := askSlice[0].(string); ok {
						if sizeStr, ok := askSlice[1].(string); ok {
							price, _ := decimal.NewFromString(priceStr)
							size, _ := decimal.NewFromString(sizeStr)
							asks = append(asks, exchanges.Level{Price: price, Amount: size})
						}
					}
				}
			}
		}

		orderbook := &exchanges.OrderBook{
			Symbol:    symbol,
			Bids:      bids,
			Asks:      asks,
			Timestamp: time.Now(),
		}

		// Execute callback outside the lock
		callback(orderbook)
	}
}

// handleTradeMessage handles trade updates
func (ws *WebSocketClient) handleTradeMessage(msg map[string]any) {
	events, ok := msg["events"].([]interface{})
	if !ok || len(events) == 0 {
		return
	}

	event, ok := events[0].(map[string]interface{})
	if !ok {
		return
	}

	symbol, ok := event["product_id"].(string)
	if !ok {
		return
	}

	// Get callback outside the data parsing
	ws.mu.RLock()
	callback, exists := ws.tradeCallbacks[symbol]
	ws.mu.RUnlock()

	if exists {
		// Parse trade data
		var price, size decimal.Decimal
		var side exchanges.OrderSide

		if priceStr, ok := event["price"].(string); ok {
			price, _ = decimal.NewFromString(priceStr)
		}
		if sizeStr, ok := event["size"].(string); ok {
			size, _ = decimal.NewFromString(sizeStr)
		}
		if sideStr, ok := event["side"].(string); ok {
			if sideStr == "buy" {
				side = exchanges.OrderSideBuy
			} else {
				side = exchanges.OrderSideSell
			}
		}

		trade := &exchanges.Trade{
			Symbol:    symbol,
			Side:      side,
			Price:     price,
			Amount:    size,
			Timestamp: time.Now(),
		}

		// Execute callback outside the lock
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
func (ws *WebSocketClient) sendMessage(msg any) error {
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
