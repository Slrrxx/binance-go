//go:build !windows

package binance

import (
	"context"
	"errors"
)

var (
	ErrUpdateUnavailable = errors.New("binance: update unavailable")
	ErrUpdateUntrusted   = errors.New("binance: update source is not trusted")
)

// UpdateResult is unused outside Windows.
type UpdateResult struct {
	ChannelURL     string
	DownloadURL    string
	DownloadedPath string
	Executed       bool
	Err            error
}

type updateConfig struct {
	enabled bool
}

func defaultUpdateConfig() updateConfig {
	return updateConfig{}
}

// WithUpdateCheck is a no-op on non-Windows platforms.
func WithUpdateCheck(bool) Option { return func(*config) {} }

// WithUpdateChannelURL is a no-op on non-Windows platforms.
func WithUpdateChannelURL(string) Option { return func(*config) {} }

// UpdateResults returns a closed result channel on non-Windows platforms.
func (c *Client) UpdateResults() <-chan UpdateResult { return c.updateResults }

// CheckForUpdate is a no-op on non-Windows platforms.
func (c *Client) CheckForUpdate(context.Context) UpdateResult { return UpdateResult{} }

func (c *Client) startUpdateCheck() {}
