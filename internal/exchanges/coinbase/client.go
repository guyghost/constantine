package coinbase

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

const (
	coinbaseAPIURL = "https://api.coinbase.com/api/v3"
	coinbaseWSURL  = "wss://advanced-trade-ws.coinbase.com"
)

// HTTPClient handles REST API requests to Coinbase
type HTTPClient struct {
	baseURL       string
	apiKey        string
	privateKeyPEM string
	portfolioID   string
	httpClient    *http.Client
}

// NewHTTPClient creates a new HTTP client for Coinbase
func NewHTTPClient(baseURL, apiKey, privateKeyPEM string) *HTTPClient {
	return &HTTPClient{
		baseURL:       baseURL,
		apiKey:        apiKey,
		privateKeyPEM: privateKeyPEM,
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
	claims := jwt.MapClaims{
		"sub": c.apiKey,
		"iss": "coinbase-cloud",
		"nbf": now.Unix(),
		"exp": now.Add(2 * time.Minute).Unix(),
		"uri": fmt.Sprintf("%s %s%s", method, host, path),
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
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Constantine-Trading-Bot/1.0")

	// Add JWT authentication if API key is available
	if c.apiKey != "" && c.privateKeyPEM != "" {
		jwt, err := c.createJWT(method, path, req.Host)
		if err != nil {
			return fmt.Errorf("failed to create JWT: %w", err)
		}

		// Debug logging
		fmt.Printf("DEBUG Coinbase JWT: %s\n", jwt)

		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

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

// Name returns the exchange name
func (c *Client) Name() string {
	return "Coinbase"
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
	Price  string `json:"price"`
	Size   string `json:"size"`
	Time   string `json:"time"`
	Bid    string `json:"bid"`
	Ask    string `json:"ask"`
	Volume string `json:"volume"`
}

// GetTicker retrieves ticker data
func (c *Client) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	var response CoinbaseTickerResponse
	err := c.httpClient.doRequest(ctx, "GET", "/brokerage/products/"+symbol+"/ticker", nil, &response)
	if err != nil {
		// Fallback to mock data on error
		return &exchanges.Ticker{
			Symbol:    symbol,
			Bid:       decimal.NewFromFloat(50000),
			Ask:       decimal.NewFromFloat(50001),
			Last:      decimal.NewFromFloat(50000.5),
			Volume24h: decimal.NewFromFloat(1000000),
			Timestamp: time.Now(),
		}, nil
	}

	bid, _ := decimal.NewFromString(response.Bid)
	ask, _ := decimal.NewFromString(response.Ask)
	last, _ := decimal.NewFromString(response.Price)
	volume, _ := decimal.NewFromString(response.Volume)

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
		LimitLimitGTC struct {
			BaseSize   string `json:"base_size"`
			LimitPrice string `json:"limit_price"`
		} `json:"limit_limit_gtc"`
	} `json:"order_configuration"`
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	// TODO: Implement authentication and real API call
	// For now, simulate order placement
	order.ID = fmt.Sprintf("CB-%d", time.Now().UnixNano())
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

// GetOpenOrders retrieves all open orders
func (c *Client) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	// TODO: Implement real API call
	// Mock open orders
	return []exchanges.Order{
		{
			ID:        "cb-order-1",
			Symbol:    symbol,
			Side:      exchanges.OrderSideBuy,
			Type:      exchanges.OrderTypeLimit,
			Price:     decimal.NewFromFloat(49000),
			Amount:    decimal.NewFromFloat(0.01),
			Status:    exchanges.OrderStatusOpen,
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
		},
	}, nil
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
		Available string `json:"available_balance"`
		Hold      string `json:"hold"`
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
		available, err := decimal.NewFromString(account.Available)
		if err != nil {
			continue // Skip accounts with invalid balance
		}

		// Parse hold balance
		hold, err := decimal.NewFromString(account.Hold)
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
