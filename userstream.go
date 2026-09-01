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
	key, err := c.UserStream().Create(ctx)
	if err != nil {
		return nil, err
	}
	st, err := c.Subscribe(ctx, StreamSpec{Name: key, Market: MarketSpot, userData: true, listenKey: key})
	if err != nil {
		return nil, err
	}
	go keepaliveLoop(st.ctx, c, key)
	return st, nil
}

func keepaliveLoop(ctx context.Context, c *Client, key string) {
	t := time.NewTicker(30 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			kctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = c.UserStream().Close(kctx, key)
			cancel()
			return
		case <-t.C:
			kctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := c.UserStream().Keepalive(kctx, key); err != nil {
				c.cfg.log.Error("listen key keepalive failed", "err", err)
			}
			cancel()
		}
	}
}
