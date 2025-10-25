package dydx

import (
	"time"

	"github.com/shopspring/decimal"
)

// TickerResponse represents the ticker response from dYdX
type TickerResponse struct {
	Markets map[string]MarketTicker `json:"markets"`
}

// MarketTicker represents ticker data for a market
type MarketTicker struct {
	Market          string          `json:"market"`
	Bid             decimal.Decimal `json:"bid"`
	Ask             decimal.Decimal `json:"ask"`
	Last            decimal.Decimal `json:"oraclePrice"`
	Volume24h       decimal.Decimal `json:"volume24H"`
	Trades24h       int             `json:"trades24H"`
	NextFundingRate decimal.Decimal `json:"nextFundingRate"`
	OpenInterest    decimal.Decimal `json:"openInterest"`
}

// OrderBookResponse represents the orderbook response
type OrderBookResponse struct {
	Bids [][]string `json:"bids"` // [price, size]
	Asks [][]string `json:"asks"` // [price, size]
}

// CandlesResponse represents historical candles
type CandlesResponse struct {
	Candles []CandleData `json:"candles"`
}

// CandleData represents a single candle
type CandleData struct {
	StartedAt       time.Time       `json:"startedAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	Market          string          `json:"market"`
	Resolution      string          `json:"resolution"`
	Low             decimal.Decimal `json:"low"`
	High            decimal.Decimal `json:"high"`
	Open            decimal.Decimal `json:"open"`
	Close           decimal.Decimal `json:"close"`
	BaseTokenVolume decimal.Decimal `json:"baseTokenVolume"`
	UsdVolume       decimal.Decimal `json:"usdVolume"`
	Trades          int             `json:"trades"`
}

// OrderRequest represents an order placement request
type OrderRequest struct {
	Market       string          `json:"market"`
	Side         string          `json:"side"`        // BUY or SELL
	Type         string          `json:"type"`        // LIMIT, MARKET, STOP_LIMIT, STOP_MARKET
	TimeInForce  string          `json:"timeInForce"` // GTT, FOK, IOC
	Size         decimal.Decimal `json:"size"`
	Price        decimal.Decimal `json:"price,omitempty"`
	LimitFee     decimal.Decimal `json:"limitFee,omitempty"`
	Expiration   string          `json:"expiration,omitempty"`
	PostOnly     bool            `json:"postOnly,omitempty"`
	ReduceOnly   bool            `json:"reduceOnly,omitempty"`
	TriggerPrice decimal.Decimal `json:"triggerPrice,omitempty"`
	ClientID     string          `json:"clientId,omitempty"`
}

// OrderResponse represents an order response
type OrderResponse struct {
	Order OrderData `json:"order"`
}

// OrderData represents order data
type OrderData struct {
	ID             string          `json:"id"`
	ClientID       string          `json:"clientId"`
	Market         string          `json:"market"`
	Side           string          `json:"side"`
	Price          decimal.Decimal `json:"price"`
	TriggerPrice   decimal.Decimal `json:"triggerPrice,omitempty"`
	Size           decimal.Decimal `json:"size"`
	RemainingSize  decimal.Decimal `json:"remainingSize"`
	Type           string          `json:"type"`
	Status         string          `json:"status"`
	TimeInForce    string          `json:"timeInForce"`
	PostOnly       bool            `json:"postOnly"`
	ReduceOnly     bool            `json:"reduceOnly"`
	OrderFlags     string          `json:"orderFlags"`
	GoodTilBlock   int64           `json:"goodTilBlock,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UnfillableAt   *time.Time      `json:"unfillableAt,omitempty"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	ClientMetadata string          `json:"clientMetadata,omitempty"`
}

// OrdersResponse represents multiple orders response
type OrdersResponse struct {
	Orders []OrderData `json:"orders"`
}

// AccountResponse represents account data
type AccountResponse struct {
	SubAccounts []SubAccount `json:"subaccounts"`
}

// SubAccount represents a subaccount
type SubAccount struct {
	Address                string                   `json:"address"`
	SubAccountNumber       int                      `json:"subaccountNumber"`
	Equity                 decimal.Decimal          `json:"equity"`
	FreeCollateral         decimal.Decimal          `json:"freeCollateral"`
	OpenPerpetualPositions map[string]PositionData  `json:"openPerpetualPositions"`
	AssetPositions         map[string]AssetPosition `json:"assetPositions"`
}

// PositionData represents position data
type PositionData struct {
	Market        string          `json:"market"`
	Status        string          `json:"status"`
	Side          string          `json:"side"`
	Size          decimal.Decimal `json:"size"`
	MaxSize       decimal.Decimal `json:"maxSize"`
	EntryPrice    decimal.Decimal `json:"entryPrice"`
	ExitPrice     decimal.Decimal `json:"exitPrice,omitempty"`
	RealizedPnl   decimal.Decimal `json:"realizedPnl"`
	UnrealizedPnl decimal.Decimal `json:"unrealizedPnl"`
	CreatedAt     time.Time       `json:"createdAt"`
	ClosedAt      *time.Time      `json:"closedAt,omitempty"`
	NetFunding    decimal.Decimal `json:"netFunding"`
	SumOpen       decimal.Decimal `json:"sumOpen"`
	SumClose      decimal.Decimal `json:"sumClose"`
}

// AssetPosition represents an asset position (e.g., USDC balance)
type AssetPosition struct {
	Symbol string          `json:"symbol"`
	Side   string          `json:"side"`
	Size   decimal.Decimal `json:"size"`
}

// TradeResponse represents trade data
type TradeResponse struct {
	Trades []TradeData `json:"trades"`
}

// TradeData represents a single trade
type TradeData struct {
	ID        string          `json:"id"`
	Side      string          `json:"side"`
	Size      decimal.Decimal `json:"size"`
	Price     decimal.Decimal `json:"price"`
	Type      string          `json:"type"`
	CreatedAt time.Time       `json:"createdAt"`
	OrderID   string          `json:"orderId,omitempty"`
	Liquidity string          `json:"liquidity,omitempty"`
}

// MarketsResponse represents markets data
type MarketsResponse struct {
	Markets map[string]MarketData `json:"markets"`
}

// MarketData represents market configuration
type MarketData struct {
	Ticker                    string          `json:"ticker"`
	MarketID                  int             `json:"marketId"`
	Status                    string          `json:"status"`
	BaseAsset                 string          `json:"baseAsset"`
	QuoteAsset                string          `json:"quoteAsset"`
	StepSize                  decimal.Decimal `json:"stepSize"`
	TickSize                  decimal.Decimal `json:"tickSize"`
	MinOrderSize              decimal.Decimal `json:"minOrderSize"`
	InitialMarginFraction     decimal.Decimal `json:"initialMarginFraction"`
	MaintenanceMarginFraction decimal.Decimal `json:"maintenanceMarginFraction"`
}
