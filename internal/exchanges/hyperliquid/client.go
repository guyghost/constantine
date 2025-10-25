package hyperliquid

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/ratelimit"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
)

const (
	hyperliquidAPIURL = "https://api.hyperliquid.xyz"
	hyperliquidWSURL  = "wss://api.hyperliquid.xyz/ws"

	// Hyperliquid rate limits (conservative estimates)
	// Generally ~50 requests per second according to docs
	hyperliquidRateLimit = 40.0 // requests per second (conservative)
)

// extractCoinFromSymbol extracts the coin name from a symbol (e.g., "BTC-USD" -> "BTC")
func extractCoinFromSymbol(symbol string) string {
	// Simple implementation - split on "-" and take first part
	parts := strings.Split(symbol, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return symbol
}

// HTTPClient handles REST API requests to Hyperliquid
type HTTPClient struct {
	baseURL     string
	apiKey      string
	apiSecret   string
	httpClient  *http.Client
	rateLimiter ratelimit.Limiter
}

// NewHTTPClient creates a new HTTP client for Hyperliquid
func NewHTTPClient(baseURL, apiKey, apiSecret string) *HTTPClient {
	// Create rate limiter with burst capability
	limiter := ratelimit.NewTokenBucket(hyperliquidRateLimit, int(hyperliquidRateLimit*2))

	return &HTTPClient{
		baseURL:     baseURL,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		rateLimiter: limiter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// createSignature creates an HMAC-SHA256 signature for Hyperliquid
func (c *HTTPClient) createSignature(message string) string {
	if c.apiSecret == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature
}

// createAuthHeaders creates authentication headers for Hyperliquid
func (c *HTTPClient) createAuthHeaders(method, path string, body []byte) map[string]string {
	headers := make(map[string]string)

	if c.apiKey == "" || c.apiSecret == "" {
		return headers
	}

	// Create timestamp
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Create message to sign: method + path + body + timestamp
	message := method + path + string(body) + timestamp

	// Create signature
	signature := c.createSignature(message)

	// Set headers
	headers["HL-API-KEY"] = c.apiKey
	headers["HL-API-SIGNATURE"] = signature
	headers["HL-API-TIMESTAMP"] = timestamp

	return headers
}

// doRequest performs an HTTP request
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body any, result any) error {
	// Apply rate limiting before making the request
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	start := time.Now()

	var reqBody []byte
	var reqBodyReader io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = jsonData
		reqBodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add authentication headers for exchange endpoints
	if strings.Contains(path, "/exchange") && c.apiKey != "" && c.apiSecret != "" {
		authHeaders := c.createAuthHeaders(method, path, reqBody)
		for key, value := range authHeaders {
			req.Header.Set(key, value)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		telemetry.RecordAPIRequest("hyperliquid", path, time.Since(start))
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		telemetry.RecordAPIRequest("hyperliquid", path, time.Since(start))
		return fmt.Errorf("API error: status=%d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			telemetry.RecordAPIRequest("hyperliquid", path, time.Since(start))
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	// Record successful request
	telemetry.RecordAPIRequest("hyperliquid", path, time.Since(start))

	return nil
}

// Client implements the exchanges.Exchange interface for Hyperliquid
type Client struct {
	apiKey     string
	apiSecret  string
	baseURL    string
	wsURL      string
	connected  bool
	ws         *WebSocketClient
	mu         sync.RWMutex
	httpClient *HTTPClient
}

// NewClient creates a new Hyperliquid client
func NewClient(apiKey, apiSecret string) *Client {
	c := &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   hyperliquidAPIURL,
		wsURL:     hyperliquidWSURL,
	}
	c.httpClient = NewHTTPClient(c.baseURL, apiKey, apiSecret)
	return c
}

// NewClientWithURL creates a new Hyperliquid client with custom URLs (for testnet)
func NewClientWithURL(apiKey, apiSecret, baseURL, wsURL string) *Client {
	c := &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   baseURL,
		wsURL:     wsURL,
	}
	c.httpClient = NewHTTPClient(c.baseURL, apiKey, apiSecret)
	return c
}

// Connect establishes connection to the exchange
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Initialize WebSocket client
	c.ws = NewWebSocketClient(c.wsURL, c.apiKey, c.apiSecret)
	if err := c.ws.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect websocket: %w", err)
	}

	c.connected = true
	return nil
}

// Disconnect closes connection to the exchange
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	if c.ws != nil {
		c.ws.Close()
	}

	c.connected = false
	return nil
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// HyperliquidTickerResponse represents the response from Hyperliquid ticker API
type HyperliquidTickerResponse []struct {
	Coin string `json:"coin"`
	Mid  string `json:"mid"`
}

// GetTicker retrieves ticker data
func (c *Client) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	// Extract coin from symbol (e.g., "BTC-USD" -> "BTC")
	coin := extractCoinFromSymbol(symbol)

	request := map[string]any{
		"type": "allMids",
	}

	var response HyperliquidTickerResponse
	err := c.httpClient.doRequest(ctx, "POST", "/info", request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	// Find the coin in the response
	for _, ticker := range response {
		if ticker.Coin == coin {
			mid, err := decimal.NewFromString(ticker.Mid)
			if err != nil {
				continue
			}

			// For simplicity, use mid price as last price
			// In a real implementation, you'd want to get bid/ask separately
			return &exchanges.Ticker{
				Symbol:    symbol,
				Bid:       mid.Sub(decimal.NewFromFloat(0.5)), // Mock bid
				Ask:       mid.Add(decimal.NewFromFloat(0.5)), // Mock ask
				Last:      mid,
				Volume24h: decimal.NewFromFloat(1000000), // Mock volume
				Timestamp: time.Now(),
			}, nil
		}
	}

	return nil, fmt.Errorf("ticker not found for coin: %s", coin)
}

// HyperliquidOrderBookResponse represents the response from Hyperliquid order book API
type HyperliquidOrderBookResponse struct {
	Coin   string `json:"coin"`
	Levels []struct {
		Bids [][]string `json:"bids"` // [[price, size], ...]
		Asks [][]string `json:"asks"` // [[price, size], ...]
	} `json:"levels"`
}

// GetOrderBook retrieves order book data
func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	coin := extractCoinFromSymbol(symbol)

	request := map[string]any{
		"type": "l2Book",
		"coin": coin,
	}

	var response HyperliquidOrderBookResponse
	err := c.httpClient.doRequest(ctx, "POST", "/info", request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	if len(response.Levels) == 0 {
		return nil, fmt.Errorf("no order book levels returned")
	}

	levels := response.Levels[0] // Take the first (and typically only) level

	// Parse bids
	bids := make([]exchanges.Level, 0, len(levels.Bids))
	for i, bid := range levels.Bids {
		if i >= depth {
			break
		}
		if len(bid) >= 2 {
			price, err := decimal.NewFromString(bid[0])
			if err != nil {
				continue
			}
			size, err := decimal.NewFromString(bid[1])
			if err != nil {
				continue
			}
			bids = append(bids, exchanges.Level{
				Price:  price,
				Amount: size,
			})
		}
	}

	// Parse asks
	asks := make([]exchanges.Level, 0, len(levels.Asks))
	for i, ask := range levels.Asks {
		if i >= depth {
			break
		}
		if len(ask) >= 2 {
			price, err := decimal.NewFromString(ask[0])
			if err != nil {
				continue
			}
			size, err := decimal.NewFromString(ask[1])
			if err != nil {
				continue
			}
			asks = append(asks, exchanges.Level{
				Price:  price,
				Amount: size,
			})
		}
	}

	return &exchanges.OrderBook{
		Symbol:    symbol,
		Bids:      bids,
		Asks:      asks,
		Timestamp: time.Now(),
	}, nil
}

// HyperliquidCandlesResponse represents the response from Hyperliquid candles API
type HyperliquidCandlesResponse []struct {
	Timestamp int64  `json:"t"` // Unix timestamp in milliseconds
	Open      string `json:"o"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Close     string `json:"c"`
	Volume    string `json:"v"`
}

// intervalToHyperliquidInterval converts interval string to Hyperliquid interval
func intervalToHyperliquidInterval(interval string) string {
	switch interval {
	case "1m":
		return "1m"
	case "5m":
		return "5m"
	case "15m":
		return "15m"
	case "1h":
		return "1h"
	case "4h":
		return "4h"
	case "1d":
		return "1d"
	default:
		return "1h" // default to 1 hour
	}
}

// GetCandles retrieves OHLCV data
func (c *Client) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	coin := extractCoinFromSymbol(symbol)
	hlInterval := intervalToHyperliquidInterval(interval)

	request := map[string]any{
		"type": "candleSnapshot",
		"req": map[string]any{
			"coin":      coin,
			"interval":  hlInterval,
			"startTime": time.Now().Add(-time.Duration(limit) * time.Hour).UnixMilli(), // Approximate
			"endTime":   time.Now().UnixMilli(),
		},
	}

	var response HyperliquidCandlesResponse
	err := c.httpClient.doRequest(ctx, "POST", "/info", request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles: %w", err)
	}

	candles := make([]exchanges.Candle, 0, len(response))
	for _, c := range response {
		timestamp := time.UnixMilli(c.Timestamp)

		open, err := decimal.NewFromString(c.Open)
		if err != nil {
			continue
		}
		high, err := decimal.NewFromString(c.High)
		if err != nil {
			continue
		}
		low, err := decimal.NewFromString(c.Low)
		if err != nil {
			continue
		}
		close, err := decimal.NewFromString(c.Close)
		if err != nil {
			continue
		}
		volume, err := decimal.NewFromString(c.Volume)
		if err != nil {
			continue
		}

		candles = append(candles, exchanges.Candle{
			Symbol:    symbol,
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	// Sort by timestamp (oldest first) using efficient sort
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Timestamp.Before(candles[j].Timestamp)
	})

	return candles, nil
}

// SubscribeTicker subscribes to ticker updates
func (c *Client) SubscribeTicker(ctx context.Context, symbol string, callback func(*exchanges.Ticker)) error {
	if c.ws == nil {
		return fmt.Errorf("websocket not connected")
	}
	return c.ws.SubscribeTicker(ctx, symbol, callback)
}

// SubscribeOrderBook subscribes to order book updates
func (c *Client) SubscribeOrderBook(ctx context.Context, symbol string, callback func(*exchanges.OrderBook)) error {
	if c.ws == nil {
		return fmt.Errorf("websocket not connected")
	}
	return c.ws.SubscribeOrderBook(ctx, symbol, callback)
}

// SubscribeTrades subscribes to trade updates
func (c *Client) SubscribeTrades(ctx context.Context, symbol string, callback func(*exchanges.Trade)) error {
	if c.ws == nil {
		return fmt.Errorf("websocket not connected")
	}
	return c.ws.SubscribeTrades(ctx, symbol, callback)
}

// HyperliquidOrderRequest represents the request body for placing orders
type HyperliquidOrderRequest struct {
	Type   string `json:"type"`
	Orders []struct {
		Coin      string `json:"coin"`
		IsBuy     bool   `json:"isBuy"`
		LimitPx   string `json:"limitPx"`
		Size      string `json:"sz"`
		OrderType struct {
			Limit struct {
				Tif string `json:"tif"` // Time in force: "Gtc" (Good till cancel)
			} `json:"limit"`
		} `json:"orderType"`
	} `json:"orders"`
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	// TODO: Implement authentication and real API call
	// For now, simulate order placement
	order.ID = fmt.Sprintf("HL-%d", time.Now().UnixNano())
	order.Status = exchanges.OrderStatusOpen
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	return order, nil
}

// CancelOrder cancels an existing order
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	// TODO: Implement REST API call
	return nil
}

// GetOrder retrieves order details
func (c *Client) GetOrder(ctx context.Context, orderID string) (*exchanges.Order, error) {
	// TODO: Implement REST API call
	return nil, nil
}

// HyperliquidOpenOrdersResponse represents the response from open orders API
type HyperliquidOpenOrdersResponse []struct {
	Coin      string `json:"coin"`
	LimitPx   string `json:"limitPx"`
	Oid       int64  `json:"oid"` // Order ID
	Side      string `json:"side"`
	Sz        string `json:"sz"`        // Size
	Timestamp int64  `json:"timestamp"` // Unix timestamp in ms
}

// GetOpenOrders retrieves all open orders
func (c *Client) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("hyperliquid requires an ethereum address (set as API key) to query open orders")
	}

	request := map[string]any{
		"type": "openOrders",
		"user": c.apiKey,
	}

	var response HyperliquidOpenOrdersResponse
	err := c.httpClient.doRequest(ctx, "POST", "/info", request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	orders := make([]exchanges.Order, 0, len(response))

	for _, o := range response {
		// Filter by symbol if specified
		orderSymbol := o.Coin + "-USD"
		if symbol != "" && orderSymbol != symbol {
			continue
		}

		// Parse price
		price, err := decimal.NewFromString(o.LimitPx)
		if err != nil {
			continue
		}

		// Parse size
		size, err := decimal.NewFromString(o.Sz)
		if err != nil {
			continue
		}

		// Determine side
		var side exchanges.OrderSide
		if o.Side == "B" || o.Side == "buy" {
			side = exchanges.OrderSideBuy
		} else {
			side = exchanges.OrderSideSell
		}

		// Convert timestamp
		timestamp := time.UnixMilli(o.Timestamp)

		order := exchanges.Order{
			ID:        fmt.Sprintf("%d", o.Oid),
			Symbol:    orderSymbol,
			Side:      side,
			Type:      exchanges.OrderTypeLimit,
			Price:     price,
			Amount:    size,
			Filled:    decimal.Zero, // Not provided in this response
			Status:    exchanges.OrderStatusOpen,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderHistory retrieves order history
func (c *Client) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]exchanges.Order, error) {
	// TODO: Implement REST API call
	return []exchanges.Order{}, nil
}

// HyperliquidBalanceResponse represents the response from Hyperliquid balance API
type HyperliquidBalanceResponse []struct {
	Coin  string `json:"coin"`
	Hold  string `json:"hold"`
	Total string `json:"total"`
	Free  string `json:"free"`
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	// Note: For Hyperliquid, we need a user address to query balance
	// apiKey should be set to the Ethereum address
	if c.apiKey == "" {
		return nil, fmt.Errorf("hyperliquid requires an ethereum address (set as API key) to query balance")
	}

	request := map[string]any{
		"type": "clearinghouseState",
		"user": c.apiKey, // apiKey should be the Ethereum address
	}

	var response struct {
		AssetPositions []struct {
			Position struct {
				Coin     string                 `json:"coin"`
				EntryPx  string                 `json:"entryPx"`
				Leverage map[string]interface{} `json:"leverage"`
				Szi      string                 `json:"szi"`
			} `json:"position"`
		} `json:"assetPositions"`
		MarginSummary struct {
			AccountValue    string `json:"accountValue"`
			TotalMarginUsed string `json:"totalMarginUsed"`
			TotalNtlPos     string `json:"totalNtlPos"`
			TotalRawUsd     string `json:"totalRawUsd"`
			Withdrawable    string `json:"withdrawable"`
		} `json:"marginSummary"`
	}

	err := c.httpClient.doRequest(ctx, "POST", "/info", request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Parse account value (total balance)
	accountValue, err := decimal.NewFromString(response.MarginSummary.AccountValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse account value: %w", err)
	}

	// Parse margin used (locked)
	marginUsed, err := decimal.NewFromString(response.MarginSummary.TotalMarginUsed)
	if err != nil {
		marginUsed = decimal.Zero
	}

	// Free = Total - Locked
	free := accountValue.Sub(marginUsed)
	if free.IsNegative() {
		free = decimal.Zero
	}

	balances := []exchanges.Balance{
		{
			Asset:     "USDC", // Hyperliquid uses USDC as collateral
			Free:      free,
			Locked:    marginUsed,
			Total:     accountValue,
			UpdatedAt: time.Now(),
		},
	}

	// Record balance metrics
	for _, balance := range balances {
		telemetry.RecordBalanceUpdate(balance.Asset, balance.Total.InexactFloat64())
	}

	return balances, nil
}

// HyperliquidPositionsResponse represents the response from Hyperliquid positions API
type HyperliquidPositionsResponse struct {
	MarginSummary struct {
		AccountValue string `json:"accountValue"`
	} `json:"marginSummary"`
	AssetPositions []struct {
		Position struct {
			Coin     string `json:"coin"`
			EntryPx  string `json:"entryPx"`
			Leverage struct {
				Value int `json:"value"`
			} `json:"leverage"`
			LiquidationPx  string `json:"liquidationPx"`
			MarginUsed     string `json:"marginUsed"`
			PositionValue  string `json:"positionValue"`
			ReturnOnEquity string `json:"returnOnEquity"`
			Szi            string `json:"szi"` // Size
			UnrealizedPnl  string `json:"unrealizedPnl"`
		} `json:"position"`
	} `json:"assetPositions"`
}

// GetPositions retrieves all open positions
func (c *Client) GetPositions(ctx context.Context) ([]exchanges.Position, error) {
	// apiKey should be set to the Ethereum address
	if c.apiKey == "" {
		return nil, fmt.Errorf("hyperliquid requires an ethereum address (set as API key) to query positions")
	}

	request := map[string]any{
		"type": "clearinghouseState",
		"user": c.apiKey,
	}

	var response HyperliquidPositionsResponse
	err := c.httpClient.doRequest(ctx, "POST", "/info", request, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	positions := make([]exchanges.Position, 0, len(response.AssetPositions))

	for _, assetPos := range response.AssetPositions {
		pos := assetPos.Position

		// Parse size
		size, err := decimal.NewFromString(pos.Szi)
		if err != nil {
			continue
		}

		// Skip zero positions
		if size.IsZero() {
			continue
		}

		// Determine side based on size sign
		side := exchanges.OrderSideBuy
		if size.IsNegative() {
			side = exchanges.OrderSideSell
			size = size.Abs() // Make size positive
		}

		// Parse entry price
		entryPrice, err := decimal.NewFromString(pos.EntryPx)
		if err != nil {
			continue
		}

		// Parse unrealized PnL
		unrealizedPnL, err := decimal.NewFromString(pos.UnrealizedPnl)
		if err != nil {
			unrealizedPnL = decimal.Zero
		}

		// Get leverage
		leverage := decimal.NewFromInt(1) // Default to 1x
		if pos.Leverage.Value > 0 {
			leverage = decimal.NewFromInt(int64(pos.Leverage.Value))
		}

		// Get current mark price (we'll fetch separately or approximate)
		// For now, calculate from unrealized PnL
		markPrice := entryPrice // Default to entry price

		// Construct symbol (coin + "-USD")
		symbol := pos.Coin + "-USD"

		position := exchanges.Position{
			Symbol:        symbol,
			Side:          side,
			Size:          size,
			EntryPrice:    entryPrice,
			MarkPrice:     markPrice,
			Leverage:      leverage,
			UnrealizedPnL: unrealizedPnL,
			RealizedPnL:   decimal.Zero, // Not provided in this response
		}

		positions = append(positions, position)

		// Record position metrics
		telemetry.RecordPositionUpdate(position.Symbol, "size", position.Size.InexactFloat64())
		telemetry.RecordPositionUpdate(position.Symbol, "unrealized_pnl", position.UnrealizedPnL.InexactFloat64())
		telemetry.RecordPositionUpdate(position.Symbol, "entry_price", position.EntryPrice.InexactFloat64())
		telemetry.RecordPositionUpdate(position.Symbol, "mark_price", position.MarkPrice.InexactFloat64())
		telemetry.RecordPnLUpdate(position.Symbol, position.UnrealizedPnL.InexactFloat64())
	}

	return positions, nil
}

// GetPosition retrieves a specific position
func (c *Client) GetPosition(ctx context.Context, symbol string) (*exchanges.Position, error) {
	// TODO: Implement REST API call
	return nil, nil
}

// SupportedSymbols returns list of supported trading symbols
func (c *Client) SupportedSymbols() []string {
	return []string{"BTC-USD", "ETH-USD", "SOL-USD", "ARB-USD"}
}

// Name returns the exchange name
func (c *Client) Name() string {
	return "Hyperliquid"
}
