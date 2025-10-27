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

// MarketData represents market configuration from dYdX API
type MarketData struct {
	ClobPairID                string          `json:"clobPairId"`
	Ticker                    string          `json:"ticker"`
	Status                    string          `json:"status"`
	OraclePrice               decimal.Decimal `json:"oraclePrice"`
	PriceChange24H            decimal.Decimal `json:"priceChange24H"`
	Volume24H                 decimal.Decimal `json:"volume24H"`
	Trades24H                 int             `json:"trades24H"`
	NextFundingRate           decimal.Decimal `json:"nextFundingRate"`
	InitialMarginFraction     decimal.Decimal `json:"initialMarginFraction"`
	MaintenanceMarginFraction decimal.Decimal `json:"maintenanceMarginFraction"`
	OpenInterest              decimal.Decimal `json:"openInterest"`
	AtomicResolution          int             `json:"atomicResolution"`
	QuantumConversionExponent int             `json:"quantumConversionExponent"`
	TickSize                  decimal.Decimal `json:"tickSize"`
	StepSize                  decimal.Decimal `json:"stepSize"`
	StepBaseQuantums          int64           `json:"stepBaseQuantums"`
	SubticksPerTick           int64           `json:"subticksPerTick"`
	MarketType                string          `json:"marketType"`
	OpenInterestLowerCap      decimal.Decimal `json:"openInterestLowerCap"`
	OpenInterestUpperCap      decimal.Decimal `json:"openInterestUpperCap"`
	BaseOpenInterest          decimal.Decimal `json:"baseOpenInterest"`
	DefaultFundingRate1H      decimal.Decimal `json:"defaultFundingRate1H"`
}

// MarketQuality represents quality metrics for a market
type MarketQuality struct {
	Symbol       string          // Market symbol (e.g., "BTC-USD")
	Volume24h    decimal.Decimal // 24-hour trading volume
	Volatility   float64         // Price volatility [0, 1]
	Liquidity    float64         // Liquidity score [0, 1]
	QualityScore float64         // Composite quality score [0, 1]
}
