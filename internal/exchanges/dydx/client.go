package dydx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/guyghost/constantine/internal/exchanges"
	"github.com/guyghost/constantine/internal/telemetry"
	"github.com/shopspring/decimal"
)

const (
	dydxAPIURL = "https://indexer.dydx.trade"
	dydxWSURL  = "wss://indexer.dydx.trade/v4/ws"
)

// Client implements the exchanges.Exchange interface for dYdX
type Client struct {
	apiKey     string
	apiSecret  string
	mnemonic   string
	baseURL    string
	wsURL      string
	connected  bool
	ws         *WebSocketClient
	wallet     *Wallet
	signer     *Signer
	mu         sync.RWMutex
	httpClient *HTTPClient
}

// NewClient creates a new dYdX client
// For dYdX, use apiSecret as the mnemonic phrase
func NewClient(apiKey, apiSecret string) (*Client, error) {
	c := &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseURL:   dydxAPIURL,
		wsURL:     dydxWSURL,
	}

	if apiSecret != "" {
		if err := ValidateMnemonic(apiSecret); err != nil {
			return nil, fmt.Errorf("invalid mnemonic: %w", err)
		}
		c.mnemonic = apiSecret
	}

	c.httpClient = NewHTTPClient(c.baseURL, apiKey, "")
	return c, nil
}

// NewClientWithMnemonic creates a new dYdX client with explicit mnemonic
func NewClientWithMnemonic(mnemonic string, subAccountNumber int) (*Client, error) {
	// Validate mnemonic
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, fmt.Errorf("invalid mnemonic: %w", err)
	}

	// Create wallet from mnemonic
	wallet, err := NewWalletFromMnemonic(mnemonic, subAccountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create signer
	signer := NewSigner(wallet)

	c := &Client{
		mnemonic: mnemonic,
		baseURL:  dydxAPIURL,
		wsURL:    dydxWSURL,
		wallet:   wallet,
		signer:   signer,
	}
	c.httpClient = NewHTTPClient(c.baseURL, "", "")
	return c, nil
}

// NewClientWithURL creates a new dYdX client with custom URLs (for testnet)
func NewClientWithURL(apiKey, apiSecret, baseURL, wsURL string) *Client {
	c := &Client{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		mnemonic:  apiSecret, // apiSecret is the mnemonic for dYdX
		baseURL:   baseURL,
		wsURL:     wsURL,
	}
	c.httpClient = NewHTTPClient(c.baseURL, apiKey, "")
	return c
}

// NewClientWithMnemonicAndURL creates a new dYdX client with explicit mnemonic and custom URLs (for testnet)
func NewClientWithMnemonicAndURL(mnemonic string, subAccountNumber int, baseURL, wsURL string) (*Client, error) {
	// Validate mnemonic
	if err := ValidateMnemonic(mnemonic); err != nil {
		return nil, fmt.Errorf("invalid mnemonic: %w", err)
	}

	// Create wallet from mnemonic
	wallet, err := NewWalletFromMnemonic(mnemonic, subAccountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create signer
	signer := NewSigner(wallet)

	c := &Client{
		mnemonic: mnemonic,
		baseURL:  baseURL,
		wsURL:    wsURL,
		wallet:   wallet,
		signer:   signer,
	}
	c.httpClient = NewHTTPClient(c.baseURL, "", "")
	return c, nil
}

// Connect establishes connection to the exchange
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Initialize wallet from mnemonic if provided and not already initialized
	if c.wallet == nil && c.mnemonic != "" {
		if err := ValidateMnemonic(c.mnemonic); err != nil {
			return fmt.Errorf("invalid mnemonic: %w", err)
		}
		wallet, err := NewWalletFromMnemonic(c.mnemonic, 0)
		if err != nil {
			return fmt.Errorf("failed to initialize wallet: %w", err)
		}
		c.wallet = wallet
		c.signer = NewSigner(wallet)
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

// GetTicker retrieves ticker data
func (c *Client) GetTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	var resp TickerResponse
	if err := c.httpClient.get(ctx, "/v4/perpetualMarkets", &resp); err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	marketTicker, ok := resp.Markets[symbol]
	if !ok {
		return nil, fmt.Errorf("market %s not found", symbol)
	}

	return &exchanges.Ticker{
		Symbol:    symbol,
		Bid:       marketTicker.Bid,
		Ask:       marketTicker.Ask,
		Last:      marketTicker.Last,
		Volume24h: marketTicker.Volume24h,
		Timestamp: time.Now(),
	}, nil
}

// GetOrderBook retrieves order book data
func (c *Client) GetOrderBook(ctx context.Context, symbol string, depth int) (*exchanges.OrderBook, error) {
	var resp OrderBookResponse
	path := fmt.Sprintf("/v4/orderbooks/perpetualMarket/%s", symbol)
	if err := c.httpClient.get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to get orderbook: %w", err)
	}

	// Convert bids
	bids := make([]exchanges.Level, 0, len(resp.Bids))
	for i, bid := range resp.Bids {
		if i >= depth {
			break
		}
		if len(bid) < 2 {
			continue
		}
		price, _ := decimal.NewFromString(bid[0])
		amount, _ := decimal.NewFromString(bid[1])
		bids = append(bids, exchanges.Level{Price: price, Amount: amount})
	}

	// Convert asks
	asks := make([]exchanges.Level, 0, len(resp.Asks))
	for i, ask := range resp.Asks {
		if i >= depth {
			break
		}
		if len(ask) < 2 {
			continue
		}
		price, _ := decimal.NewFromString(ask[0])
		amount, _ := decimal.NewFromString(ask[1])
		asks = append(asks, exchanges.Level{Price: price, Amount: amount})
	}

	return &exchanges.OrderBook{
		Symbol:    symbol,
		Bids:      bids,
		Asks:      asks,
		Timestamp: time.Now(),
	}, nil
}

// GetCandles retrieves OHLCV data
func (c *Client) GetCandles(ctx context.Context, symbol string, interval string, limit int) ([]exchanges.Candle, error) {
	var resp CandlesResponse
	path := fmt.Sprintf("/v4/candles/perpetualMarkets/%s?resolution=%s&limit=%d", symbol, interval, limit)
	if err := c.httpClient.get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to get candles: %w", err)
	}

	candles := make([]exchanges.Candle, 0, len(resp.Candles))
	for i := range resp.Candles {
		candles = append(candles, exchanges.Candle{
			Symbol:    symbol,
			Timestamp: resp.Candles[i].StartedAt,
			Open:      resp.Candles[i].Open,
			High:      resp.Candles[i].High,
			Low:       resp.Candles[i].Low,
			Close:     resp.Candles[i].Close,
			Volume:    resp.Candles[i].BaseTokenVolume,
		})
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

// PlaceOrder places a new order
// ⚠️ WARNING: NOT IMPLEMENTED - dYdX v4 requires blockchain transactions for order placement
// This requires a full node client implementation which is not yet available.
// DO NOT USE in production - orders will appear successful but nothing will execute.
func (c *Client) PlaceOrder(ctx context.Context, order *exchanges.Order) (*exchanges.Order, error) {
	return nil, fmt.Errorf("PlaceOrder not implemented for dYdX v4 - requires blockchain transaction support")
}

// CancelOrder cancels an existing order
// ⚠️ WARNING: NOT IMPLEMENTED - dYdX v4 requires blockchain transactions for order cancellation
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	return fmt.Errorf("CancelOrder not implemented for dYdX v4 - requires blockchain transaction support")
}

// GetOrder retrieves order details
// ⚠️ WARNING: NOT IMPLEMENTED - dYdX v4 indexer API may not provide individual order queries
func (c *Client) GetOrder(ctx context.Context, orderID string) (*exchanges.Order, error) {
	return nil, fmt.Errorf("GetOrder not implemented for dYdX v4 - indexer API limitations")
}

// GetOpenOrders retrieves all open orders
// ⚠️ WARNING: NOT IMPLEMENTED - dYdX v4 indexer API may not provide open orders queries
func (c *Client) GetOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	return nil, fmt.Errorf("GetOpenOrders not implemented for dYdX v4 - indexer API limitations")
}

// GetOrderHistory retrieves order history
// ⚠️ WARNING: NOT IMPLEMENTED - dYdX v4 indexer API may not provide order history queries
func (c *Client) GetOrderHistory(ctx context.Context, symbol string, limit int) ([]exchanges.Order, error) {
	return nil, fmt.Errorf("GetOrderHistory not implemented for dYdX v4 - indexer API limitations")
}

// GetBalance retrieves account balance
func (c *Client) GetBalance(ctx context.Context) ([]exchanges.Balance, error) {
	if c.wallet == nil {
		return nil, fmt.Errorf("wallet not initialized - provide mnemonic to access account data")
	}

	// Get subaccount data
	var resp AccountResponse
	path := fmt.Sprintf("/v4/addresses/%s", c.wallet.Address)
	if err := c.httpClient.get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	balances := make([]exchanges.Balance, 0)
	for _, subAccount := range resp.SubAccounts {
		if subAccount.SubAccountNumber == c.wallet.SubAccountNumber {
			// Parse asset positions (USDC balance)
			for symbol, assetPos := range subAccount.AssetPositions {
				balances = append(balances, exchanges.Balance{
					Asset:     symbol,
					Free:      assetPos.Size,
					Locked:    decimal.Zero,
					Total:     assetPos.Size,
					UpdatedAt: time.Now(),
				})
			}

			// If no specific assets, use equity
			if len(balances) == 0 {
				balances = append(balances, exchanges.Balance{
					Asset:     "USDC",
					Free:      subAccount.FreeCollateral,
					Locked:    subAccount.Equity.Sub(subAccount.FreeCollateral),
					Total:     subAccount.Equity,
					UpdatedAt: time.Now(),
				})
			}
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
	if c.wallet == nil {
		return nil, fmt.Errorf("wallet not initialized - provide mnemonic to access account data")
	}

	// Get subaccount data
	var resp AccountResponse
	path := fmt.Sprintf("/v4/addresses/%s", c.wallet.Address)
	if err := c.httpClient.get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	positions := make([]exchanges.Position, 0)
	for _, subAccount := range resp.SubAccounts {
		if subAccount.SubAccountNumber == c.wallet.SubAccountNumber {
			for _, posData := range subAccount.OpenPerpetualPositions {
				var side exchanges.OrderSide
				if posData.Side == "LONG" {
					side = exchanges.OrderSideBuy
				} else {
					side = exchanges.OrderSideSell
				}

				positions = append(positions, exchanges.Position{
					Symbol:        posData.Market,
					Side:          side,
					Size:          posData.Size,
					EntryPrice:    posData.EntryPrice,
					MarkPrice:     decimal.Zero, // Would need market data
					Leverage:      decimal.NewFromInt(1),
					UnrealizedPnL: posData.UnrealizedPnl,
					RealizedPnL:   posData.RealizedPnl,
				})
			}
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
	// Note: dYdX v4 requires address/subaccount parameters
	return nil, fmt.Errorf("not implemented - requires subaccount address")
}

// SupportedSymbols returns list of supported trading symbols
func (c *Client) SupportedSymbols() []string {
	return []string{"BTC-USD", "ETH-USD", "SOL-USD", "AVAX-USD"}
}

// Name returns the exchange name
func (c *Client) Name() string {
	return "dYdX"
}

// GetWalletAddress returns the wallet address if initialized
func (c *Client) GetWalletAddress() string {
	if c.wallet == nil {
		return ""
	}
	return c.wallet.Address
}

// GetSubAccountAddress returns the subaccount address if wallet is initialized
func (c *Client) GetSubAccountAddress() string {
	if c.wallet == nil {
		return ""
	}
	return c.wallet.SubAccountAddress()
}

// IsAuthenticated returns true if the client has a wallet initialized
func (c *Client) IsAuthenticated() bool {
	return c.wallet != nil && c.signer != nil
}
