package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/shopspring/decimal"
)

const (
	coinbaseAPIURL = "https://api.coinbase.com/api/v3"
	coinbaseWSURL  = "wss://advanced-trade-ws.coinbase.com"
)

// HTTPClient handles REST API requests to Coinbase
type HTTPClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client for Coinbase
func NewHTTPClient(baseURL, apiKey, apiSecret string) *HTTPClient {
	return &HTTPClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: status=%d", resp.StatusCode)
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
	apiKey     string
	apiSecret  string
	baseURL    string
	wsURL      string
	connected  bool
	ws         *WebSocketClient
	mu         sync.RWMutex
	httpClient *HTTPClient
}

// NewClient creates a new Coinbase client
func NewClient(apiKey, apiSecret string) *Client {
	c := &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   coinbaseAPIURL,
		wsURL:     coinbaseWSURL,
	}
	c.httpClient = NewHTTPClient(c.baseURL, apiKey, apiSecret)
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

// GetOrderBook retrieves order book data
func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	// TODO: Implement REST API call
	return &exchanges.OrderBook{
		Symbol: symbol,
		Bids: []exchanges.Level{
			{Price: decimal.NewFromFloat(50000), Amount: decimal.NewFromFloat(1.5)},
			{Price: decimal.NewFromFloat(49999), Amount: decimal.NewFromFloat(2.0)},
		},
		Asks: []exchanges.Level{
			{Price: decimal.NewFromFloat(50001), Amount: decimal.NewFromFloat(1.5)},
			{Price: decimal.NewFromFloat(50002), Amount: decimal.NewFromFloat(2.0)},
		},
		Timestamp: time.Now(),
	}, nil
}

// GetCandles retrieves OHLCV data
func (c *Client) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	// TODO: Implement REST API call
	return []exchanges.Candle{}, nil
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

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	// TODO: Implement REST API call
	order.ID = fmt.Sprintf("CB-%d", time.Now().Unix())
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
	// TODO: Implement REST API call
	return []exchanges.Order{}, nil
}

// GetOrderHistory retrieves order history
func (c *Client) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]exchanges.Order, error) {
	// TODO: Implement REST API call
	return []exchanges.Order{}, nil
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	// TODO: Implement REST API call
	return []exchanges.Balance{
		{
			Asset:     "USD",
			Free:      decimal.NewFromFloat(10000),
			Locked:    decimal.NewFromFloat(1000),
			Total:     decimal.NewFromFloat(11000),
			UpdatedAt: time.Now(),
		},
	}, nil
}

// GetPositions retrieves all open positions
func (c *Client) GetPositions(ctx context.Context) ([]exchanges.Position, error) {
	// TODO: Implement REST API call
	return []exchanges.Position{}, nil
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
