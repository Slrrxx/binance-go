package binance

import "context"

// CoinFuturesService is COIN-M futures (dapi).
type CoinFuturesService struct{ c *Client }

// Ping tests COIN-M connectivity.
func (s *CoinFuturesService) Ping(ctx context.Context) error {
	return s.c.get(ctx, familyDAPI, "/dapi/v1/ping", nil, nil)
}

// ServerTime returns COIN-M server time.
func (s *CoinFuturesService) ServerTime(ctx context.Context) (int64, error) {
	var st ServerTime
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/time", nil, &st); err != nil {
		return 0, err
	}
	return st.ServerTime, nil
}

// ExchangeInfo returns COIN-M exchange information.
func (s *CoinFuturesService) ExchangeInfo(ctx context.Context) (*FuturesExchangeInfo, error) {
	var out FuturesExchangeInfo
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/exchangeInfo", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OrderBook returns the COIN-M order book.
func (s *CoinFuturesService) OrderBook(ctx context.Context, req OrderBookRequest) (*OrderBook, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt("limit", req.Limit)
	var out OrderBook
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/depth", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Klines returns COIN-M candlesticks.
func (s *CoinFuturesService) Klines(ctx context.Context, req KlinesRequest) ([]Kline, error) {
	var out []Kline
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/klines", klineParams(req), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// TickerPrice returns COIN-M latest prices.
func (s *CoinFuturesService) TickerPrice(ctx context.Context, symbol string) ([]SymbolPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyDAPI, "/dapi/v1/ticker/price", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[SymbolPrice](raw)
}

// BookTicker returns COIN-M best bid/ask.
func (s *CoinFuturesService) BookTicker(ctx context.Context, symbol string) ([]BookTicker, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyDAPI, "/dapi/v1/ticker/bookTicker", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[BookTicker](raw)
}

// Ticker24h returns COIN-M 24h tickers.
func (s *CoinFuturesService) Ticker24h(ctx context.Context, symbol string) ([]Ticker24h, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyDAPI, "/dapi/v1/ticker/24hr", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[Ticker24h](raw)
}

// MarkPrice returns COIN-M mark prices.
func (s *CoinFuturesService) MarkPrice(ctx context.Context, symbol, pair string) ([]MarkPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	p.Set("pair", pair)
	raw, err := s.c.rawGet(ctx, familyDAPI, "/dapi/v1/premiumIndex", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[MarkPrice](raw)
}

// CreateOrder places a COIN-M order. Not retried.
func (s *CoinFuturesService) CreateOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.post(ctx, familyDAPI, "/dapi/v1/order", req.params(), &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrder queries a COIN-M order.
func (s *CoinFuturesService) GetOrder(ctx context.Context, req QueryOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrder cancels a COIN-M order.
func (s *CoinFuturesService) CancelOrder(ctx context.Context, req CancelOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.delete_(ctx, familyDAPI, "/dapi/v1/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelAllOpenOrders cancels all COIN-M open orders for a symbol.
func (s *CoinFuturesService) CancelAllOpenOrders(ctx context.Context, symbol string, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	return s.c.delete_(ctx, familyDAPI, "/dapi/v1/allOpenOrders", p, nil, signed(), recvWindow(recv))
}

// OpenOrders returns COIN-M open orders.
func (s *CoinFuturesService) OpenOrders(ctx context.Context, symbol, pair string, recv int64) ([]FuturesOrder, error) {
	p := newParams()
	p.Set("symbol", symbol)
	p.Set("pair", pair)
	var out []FuturesOrder
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/openOrders", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// AllOrders returns COIN-M order history.
func (s *CoinFuturesService) AllOrders(ctx context.Context, req AllOrdersRequest, pair string) ([]FuturesOrder, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("pair", pair)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	var out []FuturesOrder
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/allOrders", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// Balance returns COIN-M balances.
func (s *CoinFuturesService) Balance(ctx context.Context, recv int64) ([]FuturesBalance, error) {
	var out []FuturesBalance
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/balance", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// Account returns COIN-M account information.
func (s *CoinFuturesService) Account(ctx context.Context, recv int64) (*FuturesAccount, error) {
	var out FuturesAccount
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/account", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Positions returns COIN-M position risk.
func (s *CoinFuturesService) Positions(ctx context.Context, marginAsset, pair string, recv int64) ([]FuturesPosition, error) {
	p := newParams()
	p.Set("marginAsset", marginAsset)
	p.Set("pair", pair)
	var out []FuturesPosition
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/positionRisk", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// ChangeLeverage sets COIN-M leverage. Not retried.
func (s *CoinFuturesService) ChangeLeverage(ctx context.Context, symbol string, leverage int, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	p.SetInt("leverage", leverage)
	return s.c.post(ctx, familyDAPI, "/dapi/v1/leverage", p, nil, signed(), noRetry(), recvWindow(recv))
}

// ChangeMarginType sets COIN-M margin type. Not retried.
func (s *CoinFuturesService) ChangeMarginType(ctx context.Context, symbol string, marginType MarginType, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	p.Set("marginType", string(marginType))
	return s.c.post(ctx, familyDAPI, "/dapi/v1/marginType", p, nil, signed(), noRetry(), recvWindow(recv))
}

// ChangePositionMode enables/disables COIN-M hedge mode. Not retried.
func (s *CoinFuturesService) ChangePositionMode(ctx context.Context, dualSide bool, recv int64) error {
	p := newParams()
	p.SetRaw("dualSidePosition", boolString(dualSide))
	return s.c.post(ctx, familyDAPI, "/dapi/v1/positionSide/dual", p, nil, signed(), noRetry(), recvWindow(recv))
}

// GetPositionMode returns COIN-M hedge mode.
func (s *CoinFuturesService) GetPositionMode(ctx context.Context, recv int64) (*PositionMode, error) {
	var out PositionMode
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/positionSide/dual", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// IncomeHistory returns COIN-M income history.
func (s *CoinFuturesService) IncomeHistory(ctx context.Context, req IncomeHistoryRequest) ([]Income, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("incomeType", req.IncomeType)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("page", req.Page)
	p.SetInt("limit", req.Limit)
	var out []Income
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/income", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// UserTrades returns COIN-M account trades.
func (s *CoinFuturesService) UserTrades(ctx context.Context, req MyTradesRequest, pair string) ([]FuturesUserTrade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("pair", pair)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt64("fromId", req.FromID)
	p.SetInt("limit", req.Limit)
	var out []FuturesUserTrade
	if err := s.c.get(ctx, familyDAPI, "/dapi/v1/userTrades", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// ListenKey creates a COIN-M user-data listen key.
func (s *CoinFuturesService) ListenKey(ctx context.Context) (string, error) {
	var out ListenKeyResponse
	if err := s.c.post(ctx, familyDAPI, "/dapi/v1/listenKey", nil, &out, apiKey()); err != nil {
		return "", err
	}
	return out.ListenKey, nil
}

// KeepListenKey keepalives a COIN-M listen key.
func (s *CoinFuturesService) KeepListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.put(ctx, familyDAPI, "/dapi/v1/listenKey", p, nil, apiKey())
}

// CloseListenKey closes a COIN-M listen key.
func (s *CoinFuturesService) CloseListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.delete_(ctx, familyDAPI, "/dapi/v1/listenKey", p, nil, apiKey())
}
