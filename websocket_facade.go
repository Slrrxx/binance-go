package binance

import "context"

// WebSocketService is the python-binance-style stream facade.
type WebSocketService struct{ c *Client }

// Trade opens `{symbol}@trade`.
func (s *WebSocketService) Trade(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamTrade(symbol))
}

// AggTrade opens `{symbol}@aggTrade`.
func (s *WebSocketService) AggTrade(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamAggTrade(symbol))
}

// Ticker opens `{symbol}@ticker`.
func (s *WebSocketService) Ticker(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamTicker(symbol))
}

// MiniTicker opens `{symbol}@miniTicker`.
func (s *WebSocketService) MiniTicker(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamMiniTicker(symbol))
}

// BookTicker opens `{symbol}@bookTicker`.
func (s *WebSocketService) BookTicker(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamBookTicker(symbol))
}

// Depth opens `{symbol}@depth`.
func (s *WebSocketService) Depth(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamDepth(symbol))
}

// Kline opens `{symbol}@kline_{interval}`.
func (s *WebSocketService) Kline(ctx context.Context, symbol string, interval Interval) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamKline(symbol, interval))
}

// Combined opens a combined stream of specs on the same market.
func (s *WebSocketService) Combined(ctx context.Context, specs ...StreamSpec) (*Stream, error) {
	return s.c.Subscribe(ctx, specs...)
}

// FuturesTrade opens a USD-M trade stream.
func (s *WebSocketService) FuturesTrade(ctx context.Context, symbol string) (*Stream, error) {
	return s.c.Subscribe(ctx, StreamFuturesTrade(symbol))
}

// UserData opens the spot user-data stream (listen key + keepalive).
func (s *WebSocketService) UserData(ctx context.Context) (*Stream, error) {
	return s.c.UserData(ctx)
}
