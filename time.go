package binance

import (
	"context"
)

// ServerTime is the /api/v3/time response.
type ServerTime struct {
	ServerTime int64 `json:"serverTime"`
}

// SyncTime sets the local clock offset from Binance server time.
func (c *Client) SyncTime(ctx context.Context) error {
	var st ServerTime
	if err := c.get(ctx, familySpot, "/api/v3/time", nil, &st); err != nil {
		return err
	}
	local := c.nowMillis() - c.offset.Load()
	c.offset.Store(st.ServerTime - local)
	c.synced.Store(true)
	return nil
}

// Ping tests REST connectivity.
func (c *Client) Ping(ctx context.Context) error {
	return c.Spot().Ping(ctx)
}

// GetServerTime returns Binance server time in milliseconds.
func (c *Client) GetServerTime(ctx context.Context) (int64, error) {
	return c.Spot().ServerTime(ctx)
}
