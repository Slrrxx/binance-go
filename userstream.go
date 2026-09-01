package binance

import (
	"context"
	"time"
)

// UserStreamService manages spot listen keys.
type UserStreamService struct{ c *Client }

// Create opens a new listen key (valid ~60 minutes).
func (s *UserStreamService) Create(ctx context.Context) (string, error) {
	var out ListenKeyResponse
	if err := s.c.post(ctx, familySpot, "/api/v3/userDataStream", nil, &out, apiKey()); err != nil {
		return "", err
	}
	return out.ListenKey, nil
}

// Keepalive extends a listen key's validity.
func (s *UserStreamService) Keepalive(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.put(ctx, familySpot, "/api/v3/userDataStream", p, nil, apiKey())
}

// Close invalidates a listen key.
func (s *UserStreamService) Close(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.delete_(ctx, familySpot, "/api/v3/userDataStream", p, nil, apiKey())
}

// UserData opens a reconnecting user-data stream and keepalives the listen key.
func (c *Client) UserData(ctx context.Context) (*Stream, error) {
	return startUserData(ctx, c, MarketSpot, c.UserStream().Create, c.UserStream().Keepalive, c.UserStream().Close)
}

// UserData opens a USD-M futures user-data stream with listen-key keepalive.
func (s *FuturesService) UserData(ctx context.Context) (*Stream, error) {
	return startUserData(ctx, s.c, MarketUSDFutures, s.ListenKey, s.KeepListenKey, s.CloseListenKey)
}

// DepthCache starts a USD-M local order book.
func (s *FuturesService) DepthCache(ctx context.Context, symbol string) (*DepthCache, error) {
	return s.c.DepthCache(ctx, symbol, WithDepthMarket(MarketUSDFutures))
}

// UserData opens a COIN-M futures user-data stream with listen-key keepalive.
func (s *CoinFuturesService) UserData(ctx context.Context) (*Stream, error) {
	return startUserData(ctx, s.c, MarketCoinFutures, s.ListenKey, s.KeepListenKey, s.CloseListenKey)
}

// DepthCache starts a COIN-M local order book.
func (s *CoinFuturesService) DepthCache(ctx context.Context, symbol string) (*DepthCache, error) {
	return s.c.DepthCache(ctx, symbol, WithDepthMarket(MarketCoinFutures))
}

func startUserData(
	ctx context.Context,
	c *Client,
	market Market,
	create func(context.Context) (string, error),
	keep func(context.Context, string) error,
	closeFn func(context.Context, string) error,
) (*Stream, error) {
	key, err := create(ctx)
	if err != nil {
		return nil, err
	}
	st, err := c.Subscribe(ctx, StreamSpec{Name: key, Market: market, userData: true, listenKey: key})
	if err != nil {
		return nil, err
	}
	go keepaliveLoop(st.ctx, c.cfg.log, key, keep, closeFn)
	return st, nil
}

func keepaliveLoop(ctx context.Context, log Logger, key string, keep func(context.Context, string) error, closeFn func(context.Context, string) error) {
	t := time.NewTicker(30 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			kctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = closeFn(kctx, key)
			cancel()
			return
		case <-t.C:
			kctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := keep(kctx, key); err != nil {
				log.Error("listen key keepalive failed", "err", err)
			}
			cancel()
		}
	}
}
