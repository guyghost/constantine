package hyperliquid

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
)

// WebSocketClient handles WebSocket connections for Hyperliquid
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
	fmt.Printf("[DEBUG] Hyperliquid WebSocket connected to %s\n", ws.url)

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
				telemetry.RecordWebSocketReconnect("hyperliquid")
				time.Sleep(5 * time.Second)
				continue
			}

			ws.processMessage(message)
		}
	}
}

// processMessage processes a single message
func (ws *WebSocketClient) processMessage(message []byte) {
	var msg map[string]any
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	fmt.Printf("[DEBUG] Hyperliquid received message: %s\n", string(message))

	// Hyperliquid WebSocket messages have different formats
	// Check if it's a subscription response or data update
	if channel, ok := msg["channel"].(string); ok {
		switch channel {
		case "ticker":
			ws.handleTickerMessage(msg)
		case "orderbook":
			ws.handleOrderBookMessage(msg)
		case "trades":
			ws.handleTradeMessage(msg)
		}
	}
}

// handleTickerMessage handles ticker updates
func (ws *WebSocketClient) handleTickerMessage(msg map[string]any) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	data, ok := msg["data"].(map[string]any)
	if !ok {
		return
	}

	symbol, ok := data["coin"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.tickerCallbacks[symbol]; exists {
		// Parse ticker data
		var bid, ask, last decimal.Decimal
		var volume24h decimal.Decimal

		if bids, ok := data["bids"].([]any); ok && len(bids) > 0 {
			if bidData, ok := bids[0].([]any); ok && len(bidData) >= 2 {
				if priceStr, ok := bidData[0].(string); ok {
					bid, _ = decimal.NewFromString(priceStr)
				}
			}
		}

		if asks, ok := data["asks"].([]any); ok && len(asks) > 0 {
			if askData, ok := asks[0].([]any); ok && len(askData) >= 2 {
				if priceStr, ok := askData[0].(string); ok {
					ask, _ = decimal.NewFromString(priceStr)
				}
			}
		}

		if lastStr, ok := data["last"].(string); ok {
			last, _ = decimal.NewFromString(lastStr)
		}

		if volStr, ok := data["volume"].(string); ok {
			volume24h, _ = decimal.NewFromString(volStr)
		}

		ticker := &exchanges.Ticker{
			Symbol:    symbol + "-USD", // Add USD suffix
			Bid:       bid,
			Ask:       ask,
			Last:      last,
			Volume24h: volume24h,
			Timestamp: time.Now(),
		}

		fmt.Printf("[DEBUG] Hyperliquid ticker update for %s: bid=%s, ask=%s, last=%s\n",
			ticker.Symbol, bid.String(), ask.String(), last.String())
		callback(ticker)
	}
}

// handleOrderBookMessage handles order book updates
func (ws *WebSocketClient) handleOrderBookMessage(msg map[string]any) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	data, ok := msg["data"].(map[string]any)
	if !ok {
		return
	}

	symbol, ok := data["coin"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.orderbookCallbacks[symbol]; exists {
		// Parse order book data
		var bids, asks []exchanges.Level

		if bidsData, ok := data["bids"].([]any); ok {
			for _, bid := range bidsData {
				if bidSlice, ok := bid.([]any); ok && len(bidSlice) >= 2 {
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

		if asksData, ok := data["asks"].([]any); ok {
			for _, ask := range asksData {
				if askSlice, ok := ask.([]any); ok && len(askSlice) >= 2 {
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
			Symbol:    symbol + "-USD",
			Bids:      bids,
			Asks:      asks,
			Timestamp: time.Now(),
		}
		callback(orderbook)
	}
}

// handleTradeMessage handles trade updates
func (ws *WebSocketClient) handleTradeMessage(msg map[string]any) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	data, ok := msg["data"].(map[string]any)
	if !ok {
		return
	}

	symbol, ok := data["coin"].(string)
	if !ok {
		return
	}

	if callback, exists := ws.tradeCallbacks[symbol]; exists {
		// Parse trade data
		var price, size decimal.Decimal
		var side exchanges.OrderSide

		if priceStr, ok := data["px"].(string); ok {
			price, _ = decimal.NewFromString(priceStr)
		}
		if sizeStr, ok := data["sz"].(string); ok {
			size, _ = decimal.NewFromString(sizeStr)
		}
		if sideStr, ok := data["side"].(string); ok {
			if sideStr == "B" {
				side = exchanges.OrderSideBuy
			} else {
				side = exchanges.OrderSideSell
			}
		}

		trade := &exchanges.Trade{
			Symbol:    symbol + "-USD",
			Side:      side,
			Price:     price,
			Amount:    size,
			Timestamp: time.Now(),
		}
		callback(trade)
	}
}

// SubscribeTicker subscribes to ticker updates
func (ws *WebSocketClient) SubscribeTicker(ctx context.Context, symbol string, callback func(*exchanges.Ticker)) error {
	ws.mu.Lock()
	coin := strings.Split(symbol, "-")[0] // Extract coin from symbol
	ws.tickerCallbacks[coin] = callback
	ws.mu.Unlock()

	// Send subscription message
	sub := map[string]any{
		"method": "subscribe",
		"params": []string{fmt.Sprintf("ticker.%s", coin)},
	}

	fmt.Printf("[DEBUG] Hyperliquid subscribing to ticker for %s\n", symbol)
	return ws.sendMessage(sub)
}

// SubscribeOrderBook subscribes to order book updates
func (ws *WebSocketClient) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*exchanges.OrderBook)) error {
	ws.mu.Lock()
	coin := strings.Split(symbol, "-")[0]
	ws.orderbookCallbacks[coin] = callback
	ws.mu.Unlock()

	// Send subscription message
	sub := map[string]any{
		"method": "subscribe",
		"params": []string{fmt.Sprintf("orderbook.%s", coin)},
	}

	fmt.Printf("[DEBUG] Hyperliquid subscribing to orderbook for %s\n", symbol)
	return ws.sendMessage(sub)
}

// SubscribeTrades subscribes to trade updates
func (ws *WebSocketClient) SubscribeTrades(ctx context.Context, symbol string, callback func(*exchanges.Trade)) error {
	ws.mu.Lock()
	coin := strings.Split(symbol, "-")[0]
	ws.tradeCallbacks[coin] = callback
	ws.mu.Unlock()

	// Send subscription message
	sub := map[string]any{
		"method": "subscribe",
		"params": []string{fmt.Sprintf("trades.%s", coin)},
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
