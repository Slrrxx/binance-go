package binance

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Sentinel errors for errors.Is checks.
var (
	ErrMissingAPIKey    = errors.New("binance: api key is required")
	ErrMissingAPISecret = errors.New("binance: api secret or private key is required")
	ErrStreamClosed     = errors.New("binance: websocket stream closed")
	ErrStreamOverflow   = errors.New("binance: websocket event buffer overflow")
	ErrDepthDesync      = errors.New("binance: depth cache sequence gap")
	ErrNoProgress       = errors.New("binance: historical pagination made no progress")
	ErrInvalidInterval  = errors.New("binance: unknown kline interval")
)

// APIError is a JSON error body returned by Binance.
type APIError struct {
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "binance: api error"
	}
	return fmt.Sprintf("binance: api error code=%d status=%d msg=%s", e.Code, e.HTTPStatus, e.Message)
}

// IsRateLimited reports whether the error is a Binance rate-limit or IP ban.
func (e *APIError) IsRateLimited() bool {
	if e == nil {
		return false
	}
	return e.HTTPStatus == http.StatusTooManyRequests || e.HTTPStatus == 418 || e.Code == -1003
}

// DecodeError is returned when a response body cannot be decoded.
type DecodeError struct {
	Op     string
	Err    error
	Status int
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("binance: decode %s: %v", e.Op, e.Err)
}

func (e *DecodeError) Unwrap() error { return e.Err }

// AuthError is a local authentication / signing configuration problem.
type AuthError struct {
	Msg string
}

func (e *AuthError) Error() string { return "binance: auth: " + e.Msg }

// RateLimitError wraps HTTP 429 / 418 with retry metadata.
type RateLimitError struct {
	HTTPStatus int
	RetryAfter time.Duration
	UsedWeight int
	Err        error
}

func (e *RateLimitError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("binance: rate limited status=%d retry_after=%s: %v", e.HTTPStatus, e.RetryAfter, e.Err)
	}
	return fmt.Sprintf("binance: rate limited status=%d retry_after=%s", e.HTTPStatus, e.RetryAfter)
}

func (e *RateLimitError) Unwrap() error { return e.Err }

// WebsocketError is a stream-level failure.
type WebsocketError struct {
	Op  string
	Err error
}

func (e *WebsocketError) Error() string {
	return fmt.Sprintf("binance: websocket %s: %v", e.Op, e.Err)
}

func (e *WebsocketError) Unwrap() error { return e.Err }

// RequestError is a transport failure that is not a decoded API body.
type RequestError struct {
	Method string
	Path   string
	Err    error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("binance: %s %s: %v", e.Method, e.Path, e.Err)
}

func (e *RequestError) Unwrap() error { return e.Err }

func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusServiceUnavailable ||
		code == http.StatusBadGateway || code == http.StatusGatewayTimeout ||
		code == http.StatusInternalServerError
}
