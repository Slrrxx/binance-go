package binance

import (
	"encoding/json"
	"fmt"
	"time"
)

// TimeMS is a Unix timestamp in milliseconds.
type TimeMS int64

// Time converts to time.Time in UTC.
func (t TimeMS) Time() time.Time {
	if t == 0 {
		return time.Time{}
	}
	return time.UnixMilli(int64(t)).UTC()
}

// PriceLevel is a [price, quantity] pair from an order book.
type PriceLevel struct {
	Price    string
	Quantity string
}

// UnmarshalJSON accepts ["price","qty"].
func (p *PriceLevel) UnmarshalJSON(b []byte) error {
	var pair []string
	if err := json.Unmarshal(b, &pair); err != nil {
		return err
	}
	if len(pair) < 2 {
		return fmt.Errorf("binance: price level needs 2 elements")
	}
	p.Price = pair[0]
	p.Quantity = pair[1]
	return nil
}

// MarshalJSON emits ["price","qty"].
func (p PriceLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]string{p.Price, p.Quantity})
}

// OrderBook is a REST depth snapshot.
type OrderBook struct {
	LastUpdateID int64        `json:"lastUpdateId"`
	Bids         []PriceLevel `json:"bids"`
	Asks         []PriceLevel `json:"asks"`
	E            int64        `json:"E"`
	T            int64        `json:"T"`
}

// Trade is a public recent/historical trade.
type Trade struct {
	ID           int64  `json:"id"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	QuoteQty     string `json:"quoteQty"`
	Time         TimeMS `json:"time"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
	IsBestMatch  bool   `json:"isBestMatch"`
}

// AggTrade is a compressed aggregate trade.
type AggTrade struct {
	AggTradeID   int64  `json:"a"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"l"`
	Timestamp    TimeMS `json:"T"`
	IsBuyerMaker bool   `json:"m"`
	IsBestMatch  bool   `json:"M"`
}

// Kline is a candlestick. Prices and volumes are decimal strings.
type Kline struct {
	OpenTime                 TimeMS
	Open                     string
	High                     string
	Low                      string
	Close                    string
	Volume                   string
	CloseTime                TimeMS
	QuoteAssetVolume         string
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
}

// UnmarshalJSON accepts the Binance kline array encoding.
func (k *Kline) UnmarshalJSON(b []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if len(raw) < 11 {
		return fmt.Errorf("binance: kline array too short")
	}
	if err := json.Unmarshal(raw[0], &k.OpenTime); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &k.Open); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[2], &k.High); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[3], &k.Low); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[4], &k.Close); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[5], &k.Volume); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[6], &k.CloseTime); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[7], &k.QuoteAssetVolume); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[8], &k.NumberOfTrades); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[9], &k.TakerBuyBaseAssetVolume); err != nil {
		return err
	}
	return json.Unmarshal(raw[10], &k.TakerBuyQuoteAssetVolume)
}

// AvgPrice is the current average price.
type AvgPrice struct {
	Mins      int    `json:"mins"`
	Price     string `json:"price"`
	CloseTime TimeMS `json:"closeTime"`
}

// SymbolPrice is a symbol price ticker.
type SymbolPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// BookTicker is the best bid/ask.
type BookTicker struct {
	Symbol   string `json:"symbol"`
	BidPrice string `json:"bidPrice"`
	BidQty   string `json:"bidQty"`
	AskPrice string `json:"askPrice"`
	AskQty   string `json:"askQty"`
}

// Ticker24h is a 24-hour ticker statistics object.
type Ticker24h struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           TimeMS `json:"openTime"`
	CloseTime          TimeMS `json:"closeTime"`
	FirstID            int64  `json:"firstId"`
	LastID             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}

// RollingTicker is a rolling-window ticker (`/api/v3/ticker`).
type RollingTicker struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	LastPrice          string `json:"lastPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           TimeMS `json:"openTime"`
	CloseTime          TimeMS `json:"closeTime"`
	FirstID            int64  `json:"firstId"`
	LastID             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}

// ExchangeInfo is /api/v3/exchangeInfo.
type ExchangeInfo struct {
	Timezone        string           `json:"timezone"`
	ServerTime      TimeMS           `json:"serverTime"`
	RateLimits      []RateLimit      `json:"rateLimits"`
	ExchangeFilters []map[string]any `json:"exchangeFilters"`
	Symbols         []SymbolInfo     `json:"symbols"`
}

// RateLimit describes an exchange rate limit.
type RateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
}

// SymbolInfo is one tradable symbol.
type SymbolInfo struct {
	Symbol                          string         `json:"symbol"`
	Status                          string         `json:"status"`
	BaseAsset                       string         `json:"baseAsset"`
	BaseAssetPrecision              int            `json:"baseAssetPrecision"`
	QuoteAsset                      string         `json:"quoteAsset"`
	QuotePrecision                  int            `json:"quotePrecision"`
	QuoteAssetPrecision             int            `json:"quoteAssetPrecision"`
	OrderTypes                      []OrderType    `json:"orderTypes"`
	IcebergAllowed                  bool           `json:"icebergAllowed"`
	OCOAllowed                      bool           `json:"ocoAllowed"`
	OTOAllowed                      bool           `json:"otoAllowed"`
	QuoteOrderQtyMarketAllowed      bool           `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop               bool           `json:"allowTrailingStop"`
	CancelReplaceAllowed            bool           `json:"cancelReplaceAllowed"`
	IsSpotTradingAllowed            bool           `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed          bool           `json:"isMarginTradingAllowed"`
	Filters                         []SymbolFilter `json:"filters"`
	Permissions                     []string       `json:"permissions"`
	DefaultSelfTradePreventionMode  string         `json:"defaultSelfTradePreventionMode"`
	AllowedSelfTradePreventionModes []string       `json:"allowedSelfTradePreventionModes"`
}

// SymbolFilter is a lot/price/notional filter. Unknown fields stay in Extra.
type SymbolFilter struct {
	FilterType       string `json:"filterType"`
	MinPrice         string `json:"minPrice"`
	MaxPrice         string `json:"maxPrice"`
	TickSize         string `json:"tickSize"`
	MinQty           string `json:"minQty"`
	MaxQty           string `json:"maxQty"`
	StepSize         string `json:"stepSize"`
	MinNotional      string `json:"minNotional"`
	ApplyMinToMarket bool   `json:"applyMinToMarket"`
	MaxNumOrders     int    `json:"maxNumOrders"`
}

// Account is a spot account snapshot.
type Account struct {
	MakerCommission            int       `json:"makerCommission"`
	TakerCommission            int       `json:"takerCommission"`
	BuyerCommission            int       `json:"buyerCommission"`
	SellerCommission           int       `json:"sellerCommission"`
	CanTrade                   bool      `json:"canTrade"`
	CanWithdraw                bool      `json:"canWithdraw"`
	CanDeposit                 bool      `json:"canDeposit"`
	Brokered                   bool      `json:"brokered"`
	RequireSelfTradePrevention bool      `json:"requireSelfTradePrevention"`
	PreventSor                 bool      `json:"preventSor"`
	UpdateTime                 TimeMS    `json:"updateTime"`
	AccountType                string    `json:"accountType"`
	Balances                   []Balance `json:"balances"`
	Permissions                []string  `json:"permissions"`
	UID                        int64     `json:"uid"`
}

// Balance is a spot asset balance. Free and Locked are decimal strings.
type Balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// Order is a spot order.
type Order struct {
	Symbol                  string      `json:"symbol"`
	OrderID                 int64       `json:"orderId"`
	OrderListID             int64       `json:"orderListId"`
	ClientOrderID           string      `json:"clientOrderId"`
	TransactTime            TimeMS      `json:"transactTime"`
	Price                   string      `json:"price"`
	OrigQty                 string      `json:"origQty"`
	ExecutedQty             string      `json:"executedQty"`
	CummulativeQuoteQty     string      `json:"cummulativeQuoteQty"`
	Status                  OrderStatus `json:"status"`
	TimeInForce             TimeInForce `json:"timeInForce"`
	Type                    OrderType   `json:"type"`
	Side                    Side        `json:"side"`
	StopPrice               string      `json:"stopPrice"`
	IcebergQty              string      `json:"icebergQty"`
	Time                    TimeMS      `json:"time"`
	UpdateTime              TimeMS      `json:"updateTime"`
	IsWorking               bool        `json:"isWorking"`
	WorkingTime             TimeMS      `json:"workingTime"`
	OrigQuoteOrderQty       string      `json:"origQuoteOrderQty"`
	SelfTradePreventionMode string      `json:"selfTradePreventionMode"`
	Fills                   []Fill      `json:"fills"`
}

// Fill is an execution fill on an order.
type Fill struct {
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	TradeID         int64  `json:"tradeId"`
}

// MyTrade is an account trade.
type MyTrade struct {
	Symbol          string `json:"symbol"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	OrderListID     int64  `json:"orderListId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            TimeMS `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	IsBestMatch     bool   `json:"isBestMatch"`
}

// ListenKeyResponse is a user-data stream listen key.
type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

// OrderCount is the current order rate-limit usage.
type OrderCount struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
	Count         int    `json:"count"`
}
