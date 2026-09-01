package binance

import "context"

// OptionsService is vanilla options (eapi).
type OptionsService struct{ c *Client }

// OptionsExchangeInfo is /eapi/v1/exchangeInfo.
type OptionsExchangeInfo struct {
	Timezone      string              `json:"timezone"`
	ServerTime    TimeMS              `json:"serverTime"`
	OptionSymbols []FuturesSymbolInfo `json:"optionSymbols"`
}

// Ping tests options connectivity.
func (s *OptionsService) Ping(ctx context.Context) error {
	return s.c.get(ctx, familyEAPI, "/eapi/v1/ping", nil, nil)
}

// ServerTime returns options server time.
func (s *OptionsService) ServerTime(ctx context.Context) (int64, error) {
	var st ServerTime
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/time", nil, &st); err != nil {
		return 0, err
	}
	return st.ServerTime, nil
}

// ExchangeInfo returns options trading rules.
func (s *OptionsService) ExchangeInfo(ctx context.Context) (*OptionsExchangeInfo, error) {
	var out OptionsExchangeInfo
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/exchangeInfo", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// OrderBook returns an options order book.
func (s *OptionsService) OrderBook(ctx context.Context, req OrderBookRequest) (*OrderBook, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt("limit", req.Limit)
	var out OrderBook
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/depth", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Klines returns options candlesticks.
func (s *OptionsService) Klines(ctx context.Context, req KlinesRequest) ([]Kline, error) {
	var out []Kline
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/klines", klineParams(req), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Ticker24h returns options 24h tickers.
func (s *OptionsService) Ticker24h(ctx context.Context, symbol string) ([]Ticker24h, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyEAPI, "/eapi/v1/ticker", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[Ticker24h](raw)
}

// MarkPrice returns options mark prices.
func (s *OptionsService) MarkPrice(ctx context.Context, symbol string) ([]MarkPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	raw, err := s.c.rawGet(ctx, familyEAPI, "/eapi/v1/mark", p)
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[MarkPrice](raw)
}

// Account returns the options account.
func (s *OptionsService) Account(ctx context.Context, recv int64) (*FuturesAccount, error) {
	var out FuturesAccount
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/account", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Positions returns option positions.
func (s *OptionsService) Positions(ctx context.Context, symbol string, recv int64) ([]FuturesPosition, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []FuturesPosition
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/position", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateOrder places an options order. Not retried.
func (s *OptionsService) CreateOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.post(ctx, familyEAPI, "/eapi/v1/order", req.params(), &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrder cancels an options order.
func (s *OptionsService) CancelOrder(ctx context.Context, req CancelOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.delete_(ctx, familyEAPI, "/eapi/v1/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenOrders returns open options orders.
func (s *OptionsService) OpenOrders(ctx context.Context, symbol string, recv int64) ([]FuturesOrder, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []FuturesOrder
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/openOrders", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// ListenKey creates an options user-data listen key.
func (s *OptionsService) ListenKey(ctx context.Context) (string, error) {
	var out ListenKeyResponse
	if err := s.c.post(ctx, familyEAPI, "/eapi/v1/listenKey", nil, &out, apiKey()); err != nil {
		return "", err
	}
	return out.ListenKey, nil
}

// KeepListenKey keepalives an options listen key.
func (s *OptionsService) KeepListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.put(ctx, familyEAPI, "/eapi/v1/listenKey", p, nil, apiKey())
}

// CloseListenKey closes an options listen key.
func (s *OptionsService) CloseListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.delete_(ctx, familyEAPI, "/eapi/v1/listenKey", p, nil, apiKey())
}

// UserData opens an options user-data stream with listen-key keepalive.
func (s *OptionsService) UserData(ctx context.Context) (*Stream, error) {
	return startUserData(ctx, s.c, MarketOptions, s.ListenKey, s.KeepListenKey, s.CloseListenKey)
}

// UserTrades returns options account trades.
func (s *OptionsService) UserTrades(ctx context.Context, req MyTradesRequest) ([]FuturesUserTrade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("fromId", req.FromID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	var out []FuturesUserTrade
	if err := s.c.get(ctx, familyEAPI, "/eapi/v1/userTrades", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}
