package binance

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter is invoked before a REST request. Weight is the Binance
// request weight (1 when unknown).
type RateLimiter interface {
	Wait(ctx context.Context, weight int) error
}

// WeightLimiter is a simple token-bucket limiter keyed to request weight.
type WeightLimiter struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	rate     float64 // tokens per second
	last     time.Time
}

// NewWeightLimiter allows `weight` units per `window` (for example 1200/minute).
func NewWeightLimiter(weight int, window time.Duration) *WeightLimiter {
	if weight <= 0 {
		weight = 1200
	}
	if window <= 0 {
		window = time.Minute
	}
	w := float64(weight)
	return &WeightLimiter{
		capacity: w,
		tokens:   w,
		rate:     w / window.Seconds(),
		last:     time.Now(),
	}
}

// Wait blocks until `weight` tokens are available.
func (l *WeightLimiter) Wait(ctx context.Context, weight int) error {
	if weight <= 0 {
		weight = 1
	}
	need := float64(weight)
	for {
		l.mu.Lock()
		l.refill()
		if l.tokens >= need {
			l.tokens -= need
			l.mu.Unlock()
			return nil
		}
		missing := need - l.tokens
		wait := time.Duration(missing/l.rate*float64(time.Second)) + time.Millisecond
		l.mu.Unlock()
		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}

func (l *WeightLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.last = now
	l.tokens += elapsed * l.rate
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
}

func parseRetryAfter(h http.Header) time.Duration {
	raw := h.Get(headerRetryAfter)
	if raw == "" {
		return 0
	}
	if n, err := strconv.Atoi(raw); err == nil {
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(raw); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

func parseUsedWeight(h http.Header) int {
	raw := h.Get(headerUsedWeight)
	if raw == "" {
		return 0
	}
	n, _ := strconv.Atoi(raw)
	return n
}
