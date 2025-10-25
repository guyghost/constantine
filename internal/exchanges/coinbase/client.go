package coinbase

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/ratelimit"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
)

const (
	coinbaseAPIURL = "https://api.coinbase.com/api/v3"
	coinbaseWSURL  = "wss://advanced-trade-ws.coinbase.com"

	// Coinbase rate limits (conservative estimates)
	// Public endpoints: ~15 requests/second
	// Private endpoints: ~10 requests/second
	coinbasePublicRateLimit  = 10.0 // requests per second
	coinbasePrivateRateLimit = 8.0  // requests per second
)

// HTTPClient handles REST API requests to Coinbase
type HTTPClient struct {
	baseURL       string
	apiKey        string
	privateKeyPEM string
	portfolioID   string
	httpClient    *http.Client
	rateLimiter   ratelimit.Limiter
}

// NewHTTPClient creates a new HTTP client for Coinbase
func NewHTTPClient(baseURL, apiKey, privateKeyPEM string) *HTTPClient {
	// Create rate limiter with burst capability
	// Using private rate limit as it's more restrictive
	limiter := ratelimit.NewTokenBucket(coinbasePrivateRateLimit, int(coinbasePrivateRateLimit*2))

	return &HTTPClient{
		baseURL:       baseURL,
		apiKey:        apiKey,
		privateKeyPEM: privateKeyPEM,
		rateLimiter:   limiter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// createJWT creates a JWT token for Coinbase API authentication
func (c *HTTPClient) createJWT(method, path, host string) (string, error) {
	if c.apiKey == "" || c.privateKeyPEM == "" {
		return "", fmt.Errorf("API key and private key PEM required for authentication")
	}

	// Parse the private key
	keyBytes := []byte(c.privateKeyPEM)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse EC private key: %w", err)
	}

	// Create JWT claims - using full URI as in official example
	now := time.Now()

	// Construct URI in the format: "GET api.coinbase.com/api/v3/brokerage/accounts"
	uri := fmt.Sprintf("%s %s%s", method, host, path)

	claims := jwt.MapClaims{
		"sub": c.apiKey,
		"iss": "coinbase-cloud",
		"nbf": now.Unix(),
		"exp": now.Add(2 * time.Minute).Unix(),
		"uri": uri,
	}

	// Create token with ES256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.apiKey
	token.Header["nonce"] = uuid.New().String()

	// Sign the token
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// doRequest performs an HTTP request
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body any, result any) error {
	// Apply rate limiting before making the request
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	start := time.Now()

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Constantine-Trading-Bot/1.0")

	// Add JWT authentication if API key is available
	if c.apiKey != "" && c.privateKeyPEM != "" {
		// Extract host from baseURL (e.g., "https://api.coinbase.com/api/v3" -> "api.coinbase.com")
		// and construct full path including /api/v3 prefix
		fullPath := "/api/v3" + path
		jwt, err := c.createJWT(method, fullPath, "api.coinbase.com")
		if err != nil {
			return fmt.Errorf("failed to create JWT: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		telemetry.RecordAPIRequest("coinbase", path, time.Since(start))
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		telemetry.RecordAPIRequest("coinbase", path, time.Since(start))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		telemetry.RecordAPIRequest("coinbase", path, time.Since(start))
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			telemetry.RecordAPIRequest("coinbase", path, time.Since(start))
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	// Record successful request
	telemetry.RecordAPIRequest("coinbase", path, time.Since(start))

	return nil
}

// Client implements the exchanges.Exchange interface for Coinbase
type Client struct {
	apiKey        string
	apiSecret     string
	privateKeyPEM string
	portfolioID   string
	baseURL       string
	wsURL         string
	connected     bool
	ws            *WebSocketClient
	mu            sync.RWMutex
	httpClient    *HTTPClient
}

// NewClient creates a new Coinbase client
func NewClient(apiKey, privateKeyPEM string) *Client {
	return NewClientWithPortfolio(apiKey, privateKeyPEM, "")
}

// NewClientWithPortfolio creates a new Coinbase client with portfolio ID
func NewClientWithPortfolio(apiKey, privateKeyPEM, portfolioID string) *Client {
	c := &Client{
		apiKey:        apiKey,
		privateKeyPEM: privateKeyPEM,
		portfolioID:   portfolioID,
		baseURL:       coinbaseAPIURL,
		wsURL:         coinbaseWSURL,
	}
	c.httpClient = NewHTTPClient(c.baseURL, apiKey, privateKeyPEM)
	return c
}

// NewClientWithURL creates a new Coinbase client with custom URLs (for testnet)
func NewClientWithURL(apiKey, privateKeyPEM, baseURL, wsURL string) *Client {
	return NewClientWithPortfolioAndURL(apiKey, privateKeyPEM, "", baseURL, wsURL)
}

// NewClientWithPortfolioAndURL creates a new Coinbase client with portfolio ID and custom URLs
func NewClientWithPortfolioAndURL(apiKey, privateKeyPEM, portfolioID, baseURL, wsURL string) *Client {
	c := &Client{
		apiKey:        apiKey,
		privateKeyPEM: privateKeyPEM,
		portfolioID:   portfolioID,
		baseURL:       baseURL,
		wsURL:         wsURL,
	}
	c.httpClient = NewHTTPClient(c.baseURL, apiKey, privateKeyPEM)
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

// CoinbaseTickerResponse represents the response from Coinbase ticker API
type CoinbaseTickerResponse struct {
	Trades []struct {
		TradeID   string `json:"trade_id"`
		ProductID string `json:"product_id"`
		Price     string `json:"price"`
		Size      string `json:"size"`
		Time      string `json:"time"`
		Side      string `json:"side"`
	} `json:"trades"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
}

// GetTicker retrieves ticker data
func (c *Client) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	var response CoinbaseTickerResponse
	err := c.httpClient.doRequest(ctx, "GET", "/brokerage/products/"+symbol+"/ticker", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	bid, _ := decimal.NewFromString(response.BestBid)
	ask, _ := decimal.NewFromString(response.BestAsk)

	// Last price from the most recent trade
	var last decimal.Decimal
	if len(response.Trades) > 0 {
		last, _ = decimal.NewFromString(response.Trades[0].Price)
	}

	// Calculate 24h volume from recent trades (approximate)
	var volume decimal.Decimal
	for _, trade := range response.Trades {
		size, _ := decimal.NewFromString(trade.Size)
		volume = volume.Add(size)
	}

	return &exchanges.Ticker{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Last:      last,
		Volume24h: volume,
		Timestamp: time.Now(),
	}, nil
}

// CoinbaseOrderBookResponse represents the response from Coinbase order book API
type CoinbaseOrderBookResponse struct {
	PriceBook struct {
		ProductID string     `json:"product_id"`
		Bids      [][]string `json:"bids"` // [[price, size], ...]
		Asks      [][]string `json:"asks"` // [[price, size], ...]
		Time      string     `json:"time"`
	} `json:"pricebook"`
}

// GetOrderBook retrieves order book data
func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	var response CoinbaseOrderBookResponse
	path := fmt.Sprintf("/brokerage/product_book?product_id=%s", symbol)
	err := c.httpClient.doRequest(ctx, "GET", path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	// Parse bids
	bids := make([]exchanges.Level, 0, len(response.PriceBook.Bids))
	for i, bid := range response.PriceBook.Bids {
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
	asks := make([]exchanges.Level, 0, len(response.PriceBook.Asks))
	for i, ask := range response.PriceBook.Asks {
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

// CoinbaseCandlesResponse represents the response from Coinbase candles API
type CoinbaseCandlesResponse struct {
	Candles []struct {
		Start  string `json:"start"`
		Low    string `json:"low"`
		High   string `json:"high"`
		Open   string `json:"open"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	} `json:"candles"`
}

// intervalToGranularity converts interval string to Coinbase granularity
func intervalToGranularity(interval string) string {
	switch interval {
	case "1m":
		return "60"
	case "5m":
		return "300"
	case "15m":
		return "900"
	case "1h":
		return "3600"
	case "6h":
		return "21600"
	case "1d":
		return "86400"
	default:
		return "3600" // default to 1 hour
	}
}

// GetCandles retrieves OHLCV data
func (c *Client) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	granularity := intervalToGranularity(interval)
	path := fmt.Sprintf("/brokerage/products/%s/candles?granularity=%s&limit=%d", symbol, granularity, limit)

	var response CoinbaseCandlesResponse
	err := c.httpClient.doRequest(ctx, "GET", path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles: %w", err)
	}

	candles := make([]exchanges.Candle, 0, len(response.Candles))
	for _, c := range response.Candles {
		timestamp, err := time.Parse(time.RFC3339, c.Start)
		if err != nil {
			continue
		}

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

	// Reverse to get chronological order (Coinbase returns newest first)
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}

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

// CoinbaseOrderRequest represents the request body for placing orders
type CoinbaseOrderRequest struct {
	ClientOrderID string `json:"client_order_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	OrderConfig   struct {
		MarketMarketIOC *struct {
			QuoteSize string `json:"quote_size,omitempty"`
			BaseSize  string `json:"base_size,omitempty"`
		} `json:"market_market_ioc,omitempty"`
		LimitLimitGTC *struct {
			BaseSize   string `json:"base_size"`
			LimitPrice string `json:"limit_price"`
			PostOnly   bool   `json:"post_only,omitempty"`
		} `json:"limit_limit_gtc,omitempty"`
		StopLimitStopLimitGTC *struct {
			BaseSize      string `json:"base_size"`
			LimitPrice    string `json:"limit_price"`
			StopPrice     string `json:"stop_price"`
			StopDirection string `json:"stop_direction"`
		} `json:"stop_limit_stop_limit_gtc,omitempty"`
	} `json:"order_configuration"`
}

// CoinbaseOrderResponse represents the API response for order operations
type CoinbaseOrderResponse struct {
	Success      bool   `json:"success"`
	OrderID      string `json:"order_id"`
	ProductID    string `json:"product_id"`
	Side         string `json:"side"`
	ClientOID    string `json:"client_order_id"`
	ErrorMessage string `json:"error_message,omitempty"`
	Order        struct {
		OrderID            string `json:"order_id"`
		ProductID          string `json:"product_id"`
		Side               string `json:"side"`
		ClientOID          string `json:"client_order_id"`
		Status             string `json:"status"`
		OrderType          string `json:"order_type"`
		CreatedTime        string `json:"created_time"`
		CompletionTime     string `json:"completion_percentage"`
		FilledSize         string `json:"filled_size"`
		AverageFilledPrice string `json:"average_filled_price"`
		Fee                string `json:"fee"`
		NumberOfFills      string `json:"number_of_fills"`
		OrderConfiguration struct {
			MarketMarketIOC *struct {
				QuoteSize string `json:"quote_size,omitempty"`
				BaseSize  string `json:"base_size,omitempty"`
			} `json:"market_market_ioc,omitempty"`
			LimitLimitGTC *struct {
				BaseSize   string `json:"base_size"`
				LimitPrice string `json:"limit_price"`
			} `json:"limit_limit_gtc,omitempty"`
		} `json:"order_configuration"`
	} `json:"order"`
}

// CoinbaseListOrdersResponse represents response for listing orders
type CoinbaseListOrdersResponse struct {
	Orders []struct {
		OrderID            string `json:"order_id"`
		ProductID          string `json:"product_id"`
		UserID             string `json:"user_id"`
		OrderConfiguration struct {
			MarketMarketIOC *struct {
				QuoteSize string `json:"quote_size,omitempty"`
				BaseSize  string `json:"base_size,omitempty"`
			} `json:"market_market_ioc,omitempty"`
			LimitLimitGTC *struct {
				BaseSize   string `json:"base_size"`
				LimitPrice string `json:"limit_price"`
			} `json:"limit_limit_gtc,omitempty"`
		} `json:"order_configuration"`
		Side                 string `json:"side"`
		ClientOrderID        string `json:"client_order_id"`
		Status               string `json:"status"`
		TimeInForce          string `json:"time_in_force"`
		CreatedTime          string `json:"created_time"`
		CompletionPercentage string `json:"completion_percentage"`
		FilledSize           string `json:"filled_size"`
		AverageFilledPrice   string `json:"average_filled_price"`
		Fee                  string `json:"fee"`
		NumberOfFills        string `json:"number_of_fills"`
		FilledValue          string `json:"filled_value"`
		PendingCancel        bool   `json:"pending_cancel"`
		SizeInQuote          bool   `json:"size_in_quote"`
		TotalFees            string `json:"total_fees"`
		SizeInclusiveOfFees  bool   `json:"size_inclusive_of_fees"`
		TotalValueAfterFees  string `json:"total_value_after_fees"`
		TriggerStatus        string `json:"trigger_status"`
		OrderType            string `json:"order_type"`
		RejectReason         string `json:"reject_reason"`
		Settled              bool   `json:"settled"`
		ProductType          string `json:"product_type"`
	} `json:"orders"`
	HasNext bool   `json:"has_next"`
	Cursor  string `json:"cursor"`
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	// Build request
	req := CoinbaseOrderRequest{
		ClientOrderID: uuid.New().String(),
		ProductID:     order.Symbol,
		Side:          mapOrderSideToString(order.Side),
	}

	// Configure order type
	switch order.Type {
	case exchanges.OrderTypeMarket:
		req.OrderConfig.MarketMarketIOC = &struct {
			QuoteSize string `json:"quote_size,omitempty"`
			BaseSize  string `json:"base_size,omitempty"`
		}{
			BaseSize: order.Amount.String(),
		}
	case exchanges.OrderTypeLimit:
		req.OrderConfig.LimitLimitGTC = &struct {
			BaseSize   string `json:"base_size"`
			LimitPrice string `json:"limit_price"`
			PostOnly   bool   `json:"post_only,omitempty"`
		}{
			BaseSize:   order.Amount.String(),
			LimitPrice: order.Price.String(),
		}
	case exchanges.OrderTypeStopLimit:
		if order.StopPrice.IsZero() {
			return nil, fmt.Errorf("stop price required for stop limit orders")
		}
		stopDirection := "STOP_DIRECTION_STOP_DOWN"
		if order.Side == exchanges.OrderSideBuy {
			stopDirection = "STOP_DIRECTION_STOP_UP"
		}
		req.OrderConfig.StopLimitStopLimitGTC = &struct {
			BaseSize      string `json:"base_size"`
			LimitPrice    string `json:"limit_price"`
			StopPrice     string `json:"stop_price"`
			StopDirection string `json:"stop_direction"`
		}{
			BaseSize:      order.Amount.String(),
			LimitPrice:    order.Price.String(),
			StopPrice:     order.StopPrice.String(),
			StopDirection: stopDirection,
		}
	default:
		return nil, fmt.Errorf("unsupported order type: %s", order.Type)
	}

	// Make API request
	var response CoinbaseOrderResponse
	err := c.httpClient.doRequest(ctx, "POST", "/brokerage/orders", req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("order placement failed: %s", response.ErrorMessage)
	}

	// Parse response
	order.ID = response.OrderID
	order.Status = mapCoinbaseStatus(response.Order.Status)
	order.CreatedAt = parseTimeString(response.Order.CreatedTime)
	order.UpdatedAt = time.Now()

	// Parse filled information if available
	if response.Order.FilledSize != "" {
		filledSize, _ := decimal.NewFromString(response.Order.FilledSize)
		order.FilledAmount = filledSize
	}
	if response.Order.AverageFilledPrice != "" {
		avgPrice, _ := decimal.NewFromString(response.Order.AverageFilledPrice)
		order.AveragePrice = avgPrice
	}

	return order, nil
}

// Helper function to map order side
func mapOrderSideToString(side exchanges.OrderSide) string {
	switch side {
	case exchanges.OrderSideBuy:
		return "BUY"
	case exchanges.OrderSideSell:
		return "SELL"
	default:
		return "BUY"
	}
}

// Helper function to map Coinbase status to our status
func mapCoinbaseStatus(status string) exchanges.OrderStatus {
	switch status {
	case "OPEN", "PENDING":
		return exchanges.OrderStatusOpen
	case "FILLED":
		return exchanges.OrderStatusFilled
	case "CANCELLED":
		return exchanges.OrderStatusCanceled
	case "EXPIRED":
		return exchanges.OrderStatusExpired
	case "FAILED", "REJECTED":
		return exchanges.OrderStatusRejected
	default:
		return exchanges.OrderStatusOpen
	}
}

// Helper function to parse time string
func parseTimeString(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Now()
	}
	return t
}

// CancelOrder cancels an existing order
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	type CancelOrderResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message,omitempty"`
	}

	var response CancelOrderResponse
	path := fmt.Sprintf("/brokerage/orders/batch_cancel")
	body := map[string][]string{
		"order_ids": {orderID},
	}

	err := c.httpClient.doRequest(ctx, "POST", path, body, &response)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("cancel order failed: %s", response.Message)
	}

	return nil
}

// GetOrder retrieves order details
func (c *Client) GetOrder(ctx context.Context, orderID string) (*exchanges.Order, error) {
	type GetOrderResponse struct {
		Order struct {
			OrderID            string `json:"order_id"`
			ProductID          string `json:"product_id"`
			Side               string `json:"side"`
			ClientOrderID      string `json:"client_order_id"`
			Status             string `json:"status"`
			OrderType          string `json:"order_type"`
			CreatedTime        string `json:"created_time"`
			FilledSize         string `json:"filled_size"`
			AverageFilledPrice string `json:"average_filled_price"`
			Fee                string `json:"fee"`
			OrderConfiguration struct {
				MarketMarketIOC *struct {
					QuoteSize string `json:"quote_size,omitempty"`
					BaseSize  string `json:"base_size,omitempty"`
				} `json:"market_market_ioc,omitempty"`
				LimitLimitGTC *struct {
					BaseSize   string `json:"base_size"`
					LimitPrice string `json:"limit_price"`
				} `json:"limit_limit_gtc,omitempty"`
				StopLimitStopLimitGTC *struct {
					BaseSize      string `json:"base_size"`
					LimitPrice    string `json:"limit_price"`
					StopPrice     string `json:"stop_price"`
					StopDirection string `json:"stop_direction"`
				} `json:"stop_limit_stop_limit_gtc,omitempty"`
			} `json:"order_configuration"`
		} `json:"order"`
	}

	var response GetOrderResponse
	path := fmt.Sprintf("/brokerage/orders/historical/%s", orderID)
	err := c.httpClient.doRequest(ctx, "GET", path, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Parse order
	order := &exchanges.Order{
		ID:        response.Order.OrderID,
		Symbol:    response.Order.ProductID,
		Status:    mapCoinbaseStatus(response.Order.Status),
		CreatedAt: parseTimeString(response.Order.CreatedTime),
		UpdatedAt: time.Now(),
	}

	// Parse side
	if response.Order.Side == "BUY" {
		order.Side = exchanges.OrderSideBuy
	} else {
		order.Side = exchanges.OrderSideSell
	}

	// Parse order configuration
	if response.Order.OrderConfiguration.MarketMarketIOC != nil {
		order.Type = exchanges.OrderTypeMarket
		if response.Order.OrderConfiguration.MarketMarketIOC.BaseSize != "" {
			order.Amount, _ = decimal.NewFromString(response.Order.OrderConfiguration.MarketMarketIOC.BaseSize)
		}
	} else if response.Order.OrderConfiguration.LimitLimitGTC != nil {
		order.Type = exchanges.OrderTypeLimit
		order.Amount, _ = decimal.NewFromString(response.Order.OrderConfiguration.LimitLimitGTC.BaseSize)
		order.Price, _ = decimal.NewFromString(response.Order.OrderConfiguration.LimitLimitGTC.LimitPrice)
	} else if response.Order.OrderConfiguration.StopLimitStopLimitGTC != nil {
		order.Type = exchanges.OrderTypeStopLimit
		order.Amount, _ = decimal.NewFromString(response.Order.OrderConfiguration.StopLimitStopLimitGTC.BaseSize)
		order.Price, _ = decimal.NewFromString(response.Order.OrderConfiguration.StopLimitStopLimitGTC.LimitPrice)
		order.StopPrice, _ = decimal.NewFromString(response.Order.OrderConfiguration.StopLimitStopLimitGTC.StopPrice)
	}

	// Parse filled information
	if response.Order.FilledSize != "" {
		order.FilledAmount, _ = decimal.NewFromString(response.Order.FilledSize)
	}
	if response.Order.AverageFilledPrice != "" {
		order.AveragePrice, _ = decimal.NewFromString(response.Order.AverageFilledPrice)
	}

	return order, nil
}

// GetOpenOrders retrieves all open orders
func (c *Client) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	var response CoinbaseListOrdersResponse
	path := "/brokerage/orders/historical/batch"

	// Add query parameters for open orders
	queryParams := "?order_status=OPEN"
	if symbol != "" {
		queryParams += "&product_id=" + symbol
	}

	err := c.httpClient.doRequest(ctx, "GET", path+queryParams, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}

	orders := make([]exchanges.Order, 0, len(response.Orders))
	for _, cbOrder := range response.Orders {
		order := exchanges.Order{
			ID:        cbOrder.OrderID,
			Symbol:    cbOrder.ProductID,
			Status:    mapCoinbaseStatus(cbOrder.Status),
			CreatedAt: parseTimeString(cbOrder.CreatedTime),
			UpdatedAt: time.Now(),
		}

		// Parse side
		if cbOrder.Side == "BUY" {
			order.Side = exchanges.OrderSideBuy
		} else {
			order.Side = exchanges.OrderSideSell
		}

		// Parse order configuration
		if cbOrder.OrderConfiguration.MarketMarketIOC != nil {
			order.Type = exchanges.OrderTypeMarket
			if cbOrder.OrderConfiguration.MarketMarketIOC.BaseSize != "" {
				order.Amount, _ = decimal.NewFromString(cbOrder.OrderConfiguration.MarketMarketIOC.BaseSize)
			}
		} else if cbOrder.OrderConfiguration.LimitLimitGTC != nil {
			order.Type = exchanges.OrderTypeLimit
			order.Amount, _ = decimal.NewFromString(cbOrder.OrderConfiguration.LimitLimitGTC.BaseSize)
			order.Price, _ = decimal.NewFromString(cbOrder.OrderConfiguration.LimitLimitGTC.LimitPrice)
		}

		// Parse filled information
		if cbOrder.FilledSize != "" {
			order.FilledAmount, _ = decimal.NewFromString(cbOrder.FilledSize)
		}
		if cbOrder.AverageFilledPrice != "" {
			order.AveragePrice, _ = decimal.NewFromString(cbOrder.AverageFilledPrice)
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

// CoinbaseAccountsResponse represents the response from Coinbase accounts API
type CoinbaseAccountsResponse struct {
	Accounts []struct {
		UUID      string `json:"uuid"`
		Name      string `json:"name"`
		Currency  string `json:"currency"`
		Available struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"available_balance"`
		Hold struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"hold"`
	} `json:"accounts"`
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	var response CoinbaseAccountsResponse
	err := c.httpClient.doRequest(ctx, "GET", "/brokerage/accounts", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	balances := make([]exchanges.Balance, 0, len(response.Accounts))
	for _, account := range response.Accounts {
		// Parse available balance
		available, err := decimal.NewFromString(account.Available.Value)
		if err != nil {
			continue // Skip accounts with invalid balance
		}

		// Parse hold balance
		hold, err := decimal.NewFromString(account.Hold.Value)
		if err != nil {
			continue // Skip accounts with invalid hold
		}

		// Calculate total balance
		total := available.Add(hold)

		// Only include accounts with non-zero balance
		if total.GreaterThan(decimal.Zero) {
			balances = append(balances, exchanges.Balance{
				Asset:     account.Currency,
				Free:      available,
				Locked:    hold,
				Total:     total,
				UpdatedAt: time.Now(),
			})
		}
	}

	// Record balance metrics
	for _, balance := range balances {
		telemetry.RecordBalanceUpdate(balance.Asset, balance.Total.InexactFloat64())
	}

	return balances, nil
}

// GetPositions retrieves all open positions
func (c *Client) GetPositions(ctx context.Context) ([]exchanges.Position, error) {
	// For Coinbase spot trading, positions are just non-zero balances
	balances, err := c.GetBalance(ctx)
	if err != nil {
		return nil, err
	}

	positions := make([]exchanges.Position, 0)
	for _, balance := range balances {
		if balance.Total.GreaterThan(decimal.Zero) && balance.Asset != "USD" {
			// Mock position data - in reality would need market price
			positions = append(positions, exchanges.Position{
				Symbol:        balance.Asset + "-USD",
				Side:          exchanges.OrderSideBuy, // Assume long positions
				Size:          balance.Total,
				EntryPrice:    decimal.NewFromFloat(50000), // Mock price
				MarkPrice:     decimal.NewFromFloat(50000), // Mock price
				Leverage:      decimal.NewFromInt(1),
				UnrealizedPnL: decimal.Zero,
				RealizedPnL:   decimal.Zero,
			})
		}
	}

	// Record position metrics
	for _, position := range positions {
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
	return []string{"BTC-USD", "ETH-USD", "SOL-USD", "LINK-USD"}
}

// Name returns the exchange name
func (c *Client) Name() string {
	return "Coinbase"
}
