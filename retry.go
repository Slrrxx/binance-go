package binance

import (
	"context"
	"errors"
	"io"
	"net"
	"time"
)

// RetryConfig controls retries for idempotent REST calls.
type RetryConfig struct {
	// MaxAttempts includes the original try. 1 disables retries.
	MaxAttempts int
	// MinBackoff is the first delay after a failed attempt.
	MinBackoff time.Duration
	// MaxBackoff caps exponential backoff.
	MaxBackoff time.Duration
	// Retry429 retries HTTP 429 after honoring Retry-After when present.
	Retry429 bool
}

func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		MinBackoff:  200 * time.Millisecond,
		MaxBackoff:  2 * time.Second,
		Retry429:    true,
	}
}

func (c RetryConfig) backoff(attempt int) time.Duration {
	d := c.MinBackoff
	if d <= 0 {
		d = 200 * time.Millisecond
	}
	for i := 1; i < attempt; i++ {
		d *= 2
		if c.MaxBackoff > 0 && d > c.MaxBackoff {
			return c.MaxBackoff
		}
	}
	if c.MaxBackoff > 0 && d > c.MaxBackoff {
		return c.MaxBackoff
	}
	return d
}

func isTransientNetErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	return false
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
