package binance

import "context"

// FuturesService is USDⓈ-M futures (fapi).
type FuturesService struct{ c *Client }

// FuturesExchangeInfo is /fapi/v1/exchangeInfo.
type FuturesExchangeInfo struct {
	Timezone   string              `json:"timezone"`
	ServerTime TimeMS              `json:"serverTime"`
	RateLimits []RateLimit         `json:"rateLimits"`
	Symbols    []FuturesSymbolInfo `json:"symbols"`
}

// FuturesSymbolInfo is a USD-M or COIN-M symbol.
type FuturesSymbolInfo struct {
	Symbol            string         `json:"symbol"`
	Pair              string         `json:"pair"`
	ContractType      ContractType   `json:"contractType"`
	Status            string         `json:"status"`
	BaseAsset         string         `json:"baseAsset"`
	QuoteAsset        string         `json:"quoteAsset"`
	MarginAsset       string         `json:"marginAsset"`
	PricePrecision    int            `json:"pricePrecision"`
	QuantityPrecision int            `json:"quantityPrecision"`
	UnderlyingType    string         `json:"underlyingType"`
	Filters           []SymbolFilter `json:"filters"`
	OrderTypes        []OrderType    `json:"orderTypes"`
	TimeInForce       []TimeInForce  `json:"timeInForce"`
}

// Ping tests USD-M futures connectivity.
func (s *FuturesService) Ping(ctx context.Context) error {
	return s.c.get(ctx, familyFAPI, "/fapi/v1/ping", nil, nil)
}

// ServerTime returns USD-M server time.
func (s *FuturesService) ServerTime(ctx context.Context) (int64, error) {
	var st ServerTime
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/time", nil, &st); err != nil {
		return 0, err
	}
	return st.ServerTime, nil
}

// ExchangeInfo returns USD-M exchange information.
func (s *FuturesService) ExchangeInfo(ctx context.Context) (*FuturesExchangeInfo, error) {
	var out FuturesExchangeInfo
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/exchangeInfo", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OrderBook returns the USD-M order book.
func (s *FuturesService) OrderBook(ctx context.Context, req OrderBookRequest) (*OrderBook, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt("limit", req.Limit)
	var out OrderBook
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/depth", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Klines returns USD-M candlesticks.
func (s *FuturesService) Klines(ctx context.Context, req KlinesRequest) ([]Kline, error) {
	var out []Kline
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/klines", klineParams(req), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TickerPrice returns USD-M latest prices.
func (s *FuturesService) TickerPrice(ctx context.Context, symbol string) ([]SymbolPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyFAPI, "/fapi/v1/ticker/price", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[SymbolPrice](raw)
}

// BookTicker returns USD-M best bid/ask.
func (s *FuturesService) BookTicker(ctx context.Context, symbol string) ([]BookTicker, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyFAPI, "/fapi/v1/ticker/bookTicker", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[BookTicker](raw)
}

// Ticker24h returns USD-M 24h tickers.
func (s *FuturesService) Ticker24h(ctx context.Context, symbol string) ([]Ticker24h, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyFAPI, "/fapi/v1/ticker/24hr", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[Ticker24h](raw)
}

// MarkPrice is GET /fapi/v1/premiumIndex.
type MarkPrice struct {
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	NextFundingTime      TimeMS `json:"nextFundingTime"`
	InterestRate         string `json:"interestRate"`
	Time                 TimeMS `json:"time"`
}

// MarkPrice returns mark/index price and funding info.
func (s *FuturesService) MarkPrice(ctx context.Context, symbol string) ([]MarkPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyFAPI, "/fapi/v1/premiumIndex", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[MarkPrice](raw)
}

// FundingRate is GET /fapi/v1/fundingRate.
type FundingRate struct {
	Symbol      string `json:"symbol"`
	FundingRate string `json:"fundingRate"`
	FundingTime TimeMS `json:"fundingTime"`
}

// FundingRateHistory returns funding rate history.
func (s *FuturesService) FundingRateHistory(ctx context.Context, symbol string, start, end int64, limit int) ([]FundingRate, error) {
	p := newParams()
	p.Set("symbol", symbol)
	p.SetInt64("startTime", start)
	p.SetInt64("endTime", end)
	p.SetInt("limit", limit)
	var out []FundingRate
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/fundingRate", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FuturesOrderRequest is POST /fapi/v1/order.
type FuturesOrderRequest struct {
	Symbol                  string
	Side                    Side
	PositionSide            PositionSide
	Type                    OrderType
	TimeInForce             TimeInForce
	Quantity                string
	ReduceOnly              string
	Price                   string
	ClientOrderID           string
	StopPrice               string
	ClosePosition           string
	ActivationPrice         string
	CallbackRate            string
	WorkingType             WorkingType
	PriceProtect            string
	NewOrderRespType        OrderRespType
	PriceMatch              string
	SelfTradePreventionMode SelfTradePreventionMode
	GoodTillDate            int64
	RecvWindow              int64
}

func (r FuturesOrderRequest) params() params {
	p := newParams()
	p.Set("symbol", r.Symbol)
	p.Set("side", string(r.Side))
	p.Set("positionSide", string(r.PositionSide))
	p.Set("type", string(r.Type))
	p.Set("timeInForce", string(r.TimeInForce))
	p.Set("quantity", r.Quantity)
	p.Set("reduceOnly", r.ReduceOnly)
	p.Set("price", r.Price)
	p.Set("newClientOrderId", r.ClientOrderID)
	p.Set("stopPrice", r.StopPrice)
	p.Set("closePosition", r.ClosePosition)
	p.Set("activationPrice", r.ActivationPrice)
	p.Set("callbackRate", r.CallbackRate)
	p.Set("workingType", string(r.WorkingType))
	p.Set("priceProtect", r.PriceProtect)
	p.Set("newOrderRespType", string(r.NewOrderRespType))
	p.Set("priceMatch", r.PriceMatch)
	p.Set("selfTradePreventionMode", string(r.SelfTradePreventionMode))
	p.SetInt64("goodTillDate", r.GoodTillDate)
	return p
}

// FuturesOrder is a USD-M / COIN-M order.
type FuturesOrder struct {
	ClientOrderID string       `json:"clientOrderId"`
	CumQty        string       `json:"cumQty"`
	CumQuote      string       `json:"cumQuote"`
	ExecutedQty   string       `json:"executedQty"`
	OrderID       int64        `json:"orderId"`
	AvgPrice      string       `json:"avgPrice"`
	OrigQty       string       `json:"origQty"`
	Price         string       `json:"price"`
	ReduceOnly    bool         `json:"reduceOnly"`
	Side          Side         `json:"side"`
	PositionSide  PositionSide `json:"positionSide"`
	Status        OrderStatus  `json:"status"`
	StopPrice     string       `json:"stopPrice"`
	ClosePosition bool         `json:"closePosition"`
	Symbol        string       `json:"symbol"`
	TimeInForce   TimeInForce  `json:"timeInForce"`
	Type          OrderType    `json:"type"`
	OrigType      OrderType    `json:"origType"`
	UpdateTime    TimeMS       `json:"updateTime"`
	WorkingType   WorkingType  `json:"workingType"`
	PriceProtect  bool         `json:"priceProtect"`
}

// CreateOrder places a USD-M futures order. Not retried.
func (s *FuturesService) CreateOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.post(ctx, familyFAPI, "/fapi/v1/order", req.params(), &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// TestOrder validates a USD-M order without matching.
func (s *FuturesService) TestOrder(ctx context.Context, req FuturesOrderRequest) error {
	return s.c.post(ctx, familyFAPI, "/fapi/v1/order/test", req.params(), nil, signed(), noRetry(), recvWindow(req.RecvWindow))
}

// GetOrder queries a USD-M order.
func (s *FuturesService) GetOrder(ctx context.Context, req QueryOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrder cancels a USD-M order.
func (s *FuturesService) CancelOrder(ctx context.Context, req CancelOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.delete_(ctx, familyFAPI, "/fapi/v1/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelAllOpenOrders cancels all USD-M open orders for a symbol.
func (s *FuturesService) CancelAllOpenOrders(ctx context.Context, symbol string, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	return s.c.delete_(ctx, familyFAPI, "/fapi/v1/allOpenOrders", p, nil, signed(), recvWindow(recv))
}

// OpenOrders returns USD-M open orders.
func (s *FuturesService) OpenOrders(ctx context.Context, symbol string, recv int64) ([]FuturesOrder, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []FuturesOrder
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/openOrders", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// AllOrders returns USD-M order history for a symbol.
func (s *FuturesService) AllOrders(ctx context.Context, req AllOrdersRequest) ([]FuturesOrder, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	var out []FuturesOrder
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/allOrders", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// FuturesBalance is GET /fapi/v2/balance.
type FuturesBalance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
	MarginAvailable    bool   `json:"marginAvailable"`
	UpdateTime         TimeMS `json:"updateTime"`
}

// Balance returns USD-M futures balances.
func (s *FuturesService) Balance(ctx context.Context, recv int64) ([]FuturesBalance, error) {
	var out []FuturesBalance
	if err := s.c.get(ctx, familyFAPI, "/fapi/v2/balance", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// FuturesAccount is GET /fapi/v2/account.
type FuturesAccount struct {
	FeeTier                     int               `json:"feeTier"`
	CanTrade                    bool              `json:"canTrade"`
	CanDeposit                  bool              `json:"canDeposit"`
	CanWithdraw                 bool              `json:"canWithdraw"`
	UpdateTime                  TimeMS            `json:"updateTime"`
	TotalInitialMargin          string            `json:"totalInitialMargin"`
	TotalMaintMargin            string            `json:"totalMaintMargin"`
	TotalWalletBalance          string            `json:"totalWalletBalance"`
	TotalUnrealizedProfit       string            `json:"totalUnrealizedProfit"`
	TotalMarginBalance          string            `json:"totalMarginBalance"`
	TotalPositionInitialMargin  string            `json:"totalPositionInitialMargin"`
	TotalOpenOrderInitialMargin string            `json:"totalOpenOrderInitialMargin"`
	TotalCrossWalletBalance     string            `json:"totalCrossWalletBalance"`
	TotalCrossUnPnl             string            `json:"totalCrossUnPnl"`
	AvailableBalance            string            `json:"availableBalance"`
	MaxWithdrawAmount           string            `json:"maxWithdrawAmount"`
	Assets                      []FuturesAsset    `json:"assets"`
	Positions                   []FuturesPosition `json:"positions"`
}

// FuturesAsset is one futures wallet asset.
type FuturesAsset struct {
	Asset                  string `json:"asset"`
	WalletBalance          string `json:"walletBalance"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	MarginBalance          string `json:"marginBalance"`
	MaintMargin            string `json:"maintMargin"`
	InitialMargin          string `json:"initialMargin"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	CrossWalletBalance     string `json:"crossWalletBalance"`
	CrossUnPnl             string `json:"crossUnPnl"`
	AvailableBalance       string `json:"availableBalance"`
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
	UpdateTime             TimeMS `json:"updateTime"`
}

// Account returns USD-M account information.
func (s *FuturesService) Account(ctx context.Context, recv int64) (*FuturesAccount, error) {
	var out FuturesAccount
	if err := s.c.get(ctx, familyFAPI, "/fapi/v2/account", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// FuturesPosition is GET /fapi/v2/positionRisk.
type FuturesPosition struct {
	Symbol           string       `json:"symbol"`
	PositionAmt      string       `json:"positionAmt"`
	EntryPrice       string       `json:"entryPrice"`
	BreakEvenPrice   string       `json:"breakEvenPrice"`
	MarkPrice        string       `json:"markPrice"`
	UnRealizedProfit string       `json:"unRealizedProfit"`
	LiquidationPrice string       `json:"liquidationPrice"`
	Leverage         string       `json:"leverage"`
	MaxNotionalValue string       `json:"maxNotionalValue"`
	MarginType       MarginType   `json:"marginType"`
	IsolatedMargin   string       `json:"isolatedMargin"`
	IsAutoAddMargin  string       `json:"isAutoAddMargin"`
	PositionSide     PositionSide `json:"positionSide"`
	Notional         string       `json:"notional"`
	IsolatedWallet   string       `json:"isolatedWallet"`
	UpdateTime       TimeMS       `json:"updateTime"`
}

// Positions returns USD-M position risk.
func (s *FuturesService) Positions(ctx context.Context, symbol string, recv int64) ([]FuturesPosition, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []FuturesPosition
	if err := s.c.get(ctx, familyFAPI, "/fapi/v2/positionRisk", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// ChangeLeverage sets symbol leverage. Not retried.
func (s *FuturesService) ChangeLeverage(ctx context.Context, symbol string, leverage int, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	p.SetInt("leverage", leverage)
	return s.c.post(ctx, familyFAPI, "/fapi/v1/leverage", p, nil, signed(), noRetry(), recvWindow(recv))
}

// ChangeMarginType sets isolated or crossed margin. Not retried.
func (s *FuturesService) ChangeMarginType(ctx context.Context, symbol string, marginType MarginType, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	p.Set("marginType", string(marginType))
	return s.c.post(ctx, familyFAPI, "/fapi/v1/marginType", p, nil, signed(), noRetry(), recvWindow(recv))
}

// ChangePositionMode enables/disables hedge mode. Not retried.
func (s *FuturesService) ChangePositionMode(ctx context.Context, dualSide bool, recv int64) error {
	p := newParams()
	p.SetRaw("dualSidePosition", boolString(dualSide))
	return s.c.post(ctx, familyFAPI, "/fapi/v1/positionSide/dual", p, nil, signed(), noRetry(), recvWindow(recv))
}

// PositionMode is GET /fapi/v1/positionSide/dual.
type PositionMode struct {
	DualSidePosition bool `json:"dualSidePosition"`
}

// GetPositionMode returns whether hedge mode is enabled.
func (s *FuturesService) GetPositionMode(ctx context.Context, recv int64) (*PositionMode, error) {
	var out PositionMode
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/positionSide/dual", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// Income is GET /fapi/v1/income.
type Income struct {
	Symbol     string `json:"symbol"`
	IncomeType string `json:"incomeType"`
	Income     string `json:"income"`
	Asset      string `json:"asset"`
	Info       string `json:"info"`
	Time       TimeMS `json:"time"`
	TranID     int64  `json:"tranId"`
	TradeID    string `json:"tradeId"`
}

// IncomeHistoryRequest is GET /fapi/v1/income.
type IncomeHistoryRequest struct {
	Symbol     string
	IncomeType string
	StartTime  int64
	EndTime    int64
	Page       int
	Limit      int
	RecvWindow int64
}

// IncomeHistory returns USD-M income (funding, realized PnL, ...).
func (s *FuturesService) IncomeHistory(ctx context.Context, req IncomeHistoryRequest) ([]Income, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("incomeType", req.IncomeType)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("page", req.Page)
	p.SetInt("limit", req.Limit)
	var out []Income
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/income", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// FuturesUserTrade is GET /fapi/v1/userTrades.
type FuturesUserTrade struct {
	Buyer           bool         `json:"buyer"`
	Commission      string       `json:"commission"`
	CommissionAsset string       `json:"commissionAsset"`
	ID              int64        `json:"id"`
	Maker           bool         `json:"maker"`
	OrderID         int64        `json:"orderId"`
	Price           string       `json:"price"`
	Qty             string       `json:"qty"`
	QuoteQty        string       `json:"quoteQty"`
	RealizedPnl     string       `json:"realizedPnl"`
	Side            Side         `json:"side"`
	PositionSide    PositionSide `json:"positionSide"`
	Symbol          string       `json:"symbol"`
	Time            TimeMS       `json:"time"`
}

// UserTrades returns USD-M account trades.
func (s *FuturesService) UserTrades(ctx context.Context, req MyTradesRequest) ([]FuturesUserTrade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt64("fromId", req.FromID)
	p.SetInt("limit", req.Limit)
	var out []FuturesUserTrade
	if err := s.c.get(ctx, familyFAPI, "/fapi/v1/userTrades", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// ListenKey creates a USD-M user-data listen key.
func (s *FuturesService) ListenKey(ctx context.Context) (string, error) {
	var out ListenKeyResponse
	if err := s.c.post(ctx, familyFAPI, "/fapi/v1/listenKey", nil, &out, apiKey()); err != nil {
		return "", err
	}
	return out.ListenKey, nil
}

// KeepListenKey keepalives a USD-M listen key.
func (s *FuturesService) KeepListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.put(ctx, familyFAPI, "/fapi/v1/listenKey", p, nil, apiKey())
}

// CloseListenKey closes a USD-M listen key.
func (s *FuturesService) CloseListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.delete_(ctx, familyFAPI, "/fapi/v1/listenKey", p, nil, apiKey())
}
