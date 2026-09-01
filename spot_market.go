package binance

import "context"

// ExchangeInfoRequest filters GET /api/v3/exchangeInfo.
type ExchangeInfoRequest struct {
	Symbol  string
	Symbols []string
}

// ExchangeInfo returns trading rules and symbol information.
func (s *SpotService) ExchangeInfo(ctx context.Context, req ExchangeInfoRequest) (*ExchangeInfo, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	if len(req.Symbols) > 0 {
		raw, err := marshalJSONArray(req.Symbols)
		if err != nil {
			return nil, err
		}
		p.SetRaw("symbols", raw)
	}
	var out ExchangeInfo
	if err := s.c.get(ctx, familySpot, "/api/v3/exchangeInfo", p, &out, weight(20)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ping tests connectivity to the spot REST API.
func (s *SpotService) Ping(ctx context.Context) error {
	return s.c.get(ctx, familySpot, "/api/v3/ping", nil, nil)
}

// ServerTime returns the Binance server time in milliseconds.
func (s *SpotService) ServerTime(ctx context.Context) (int64, error) {
	var st ServerTime
	if err := s.c.get(ctx, familySpot, "/api/v3/time", nil, &st); err != nil {
		return 0, err
	}
	return st.ServerTime, nil
}

// OrderBookRequest is GET /api/v3/depth.
type OrderBookRequest struct {
	Symbol string
	Limit  int
}

// OrderBook returns the current order book.
func (s *SpotService) OrderBook(ctx context.Context, req OrderBookRequest) (*OrderBook, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt("limit", req.Limit)
	w := 5
	switch {
	case req.Limit >= 1000:
		w = 50
	case req.Limit >= 500:
		w = 25
	case req.Limit >= 100:
		w = 10
	}
	var out OrderBook
	if err := s.c.get(ctx, familySpot, "/api/v3/depth", p, &out, weight(w)); err != nil {
		return nil, err
	}
	return &out, nil
}

// RecentTradesRequest is GET /api/v3/trades.
type RecentTradesRequest struct {
	Symbol string
	Limit  int
}

// RecentTrades returns the most recent trades.
func (s *SpotService) RecentTrades(ctx context.Context, req RecentTradesRequest) ([]Trade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt("limit", req.Limit)
	var out []Trade
	if err := s.c.get(ctx, familySpot, "/api/v3/trades", p, &out, weight(10)); err != nil {
		return nil, err
	}
	return out, nil
}

// HistoricalTradesRequest is GET /api/v3/historicalTrades (API key).
type HistoricalTradesRequest struct {
	Symbol string
	Limit  int
	FromID int64
}

// HistoricalTrades returns older public trades. Requires an API key.
func (s *SpotService) HistoricalTrades(ctx context.Context, req HistoricalTradesRequest) ([]Trade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt("limit", req.Limit)
	p.SetInt64("fromId", req.FromID)
	var out []Trade
	if err := s.c.get(ctx, familySpot, "/api/v3/historicalTrades", p, &out, apiKey(), weight(25)); err != nil {
		return nil, err
	}
	return out, nil
}

// AggTradesRequest is GET /api/v3/aggTrades.
type AggTradesRequest struct {
	Symbol    string
	FromID    int64
	StartTime int64
	EndTime   int64
	Limit     int
}

// AggTrades returns compressed aggregate trades.
func (s *SpotService) AggTrades(ctx context.Context, req AggTradesRequest) ([]AggTrade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("fromId", req.FromID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	var out []AggTrade
	if err := s.c.get(ctx, familySpot, "/api/v3/aggTrades", p, &out, weight(4)); err != nil {
		return nil, err
	}
	return out, nil
}

// KlinesRequest is GET /api/v3/klines.
type KlinesRequest struct {
	Symbol    string
	Interval  Interval
	StartTime int64
	EndTime   int64
	Limit     int
	TimeZone  string
}

func klineParams(req KlinesRequest) params {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("interval", string(req.Interval))
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	p.Set("timeZone", req.TimeZone)
	return p
}

// Klines returns candlestick bars.
func (s *SpotService) Klines(ctx context.Context, req KlinesRequest) ([]Kline, error) {
	var out []Kline
	if err := s.c.get(ctx, familySpot, "/api/v3/klines", klineParams(req), &out, weight(2)); err != nil {
		return nil, err
	}
	return out, nil
}

// UIKlines returns candlesticks optimized for presentation.
func (s *SpotService) UIKlines(ctx context.Context, req KlinesRequest) ([]Kline, error) {
	var out []Kline
	if err := s.c.get(ctx, familySpot, "/api/v3/uiKlines", klineParams(req), &out, weight(2)); err != nil {
		return nil, err
	}
	return out, nil
}

// AvgPrice returns the current average price for a symbol.
func (s *SpotService) AvgPrice(ctx context.Context, symbol string) (*AvgPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out AvgPrice
	if err := s.c.get(ctx, familySpot, "/api/v3/avgPrice", p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ticker24hRequest is GET /api/v3/ticker/24hr.
type Ticker24hRequest struct {
	Symbol  string
	Symbols []string
	Type    string
}

// Ticker24h returns 24-hour rolling window price change statistics.
func (s *SpotService) Ticker24h(ctx context.Context, req Ticker24hRequest) ([]Ticker24h, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("type", req.Type)
	if len(req.Symbols) > 0 {
		raw, err := marshalJSONArray(req.Symbols)
		if err != nil {
			return nil, err
		}
		p.SetRaw("symbols", raw)
	}
	raw, err := s.c.rawGet(ctx, familySpot, "/api/v3/ticker/24hr", p, weight(2))
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[Ticker24h](raw)
}

// TickerPriceRequest is GET /api/v3/ticker/price.
type TickerPriceRequest struct {
	Symbol  string
	Symbols []string
}

// TickerPrice returns the latest price for one or all symbols.
func (s *SpotService) TickerPrice(ctx context.Context, req TickerPriceRequest) ([]SymbolPrice, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	if len(req.Symbols) > 0 {
		raw, err := marshalJSONArray(req.Symbols)
		if err != nil {
			return nil, err
		}
		p.SetRaw("symbols", raw)
	}
	raw, err := s.c.rawGet(ctx, familySpot, "/api/v3/ticker/price", p, weight(2))
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[SymbolPrice](raw)
}

// BookTickerRequest is GET /api/v3/ticker/bookTicker.
type BookTickerRequest struct {
	Symbol  string
	Symbols []string
}

// BookTicker returns the best bid/ask for one or all symbols.
func (s *SpotService) BookTicker(ctx context.Context, req BookTickerRequest) ([]BookTicker, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	if len(req.Symbols) > 0 {
		raw, err := marshalJSONArray(req.Symbols)
		if err != nil {
			return nil, err
		}
		p.SetRaw("symbols", raw)
	}
	raw, err := s.c.rawGet(ctx, familySpot, "/api/v3/ticker/bookTicker", p, weight(2))
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[BookTicker](raw)
}

// RollingTickerRequest is GET /api/v3/ticker.
type RollingTickerRequest struct {
	Symbol     string
	Symbols    []string
	WindowSize string
	Type       string
}

// RollingTicker returns a rolling-window ticker (default 1d when omitted).
func (s *SpotService) RollingTicker(ctx context.Context, req RollingTickerRequest) ([]RollingTicker, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("windowSize", req.WindowSize)
	p.Set("type", req.Type)
	if len(req.Symbols) > 0 {
		raw, err := marshalJSONArray(req.Symbols)
		if err != nil {
			return nil, err
		}
		p.SetRaw("symbols", raw)
	}
	raw, err := s.c.rawGet(ctx, familySpot, "/api/v3/ticker", p, weight(4))
	if err != nil {
		return nil, err
	}
	return unmarshalOneOrMany[RollingTicker](raw)
}
