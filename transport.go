package binance

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type security int

const (
	secNone security = iota
	secAPIKey
	secSigned
)

type apiCall struct {
	method     string
	family     family
	path       string
	params     params
	sec        security
	weight     int
	retryable  bool
	recvWindow int64
}

func (c *Client) get(ctx context.Context, fam family, path string, p params, dest any, opts ...callOpt) error {
	call := apiCall{method: http.MethodGet, family: fam, path: path, params: p, retryable: true, weight: 1}
	for _, o := range opts {
		o(&call)
	}
	return c.do(ctx, call, dest)
}

func (c *Client) post(ctx context.Context, fam family, path string, p params, dest any, opts ...callOpt) error {
	call := apiCall{method: http.MethodPost, family: fam, path: path, params: p, weight: 1}
	for _, o := range opts {
		o(&call)
	}
	return c.do(ctx, call, dest)
}

func (c *Client) put(ctx context.Context, fam family, path string, p params, dest any, opts ...callOpt) error {
	call := apiCall{method: http.MethodPut, family: fam, path: path, params: p, retryable: true, weight: 1}
	for _, o := range opts {
		o(&call)
	}
	return c.do(ctx, call, dest)
}

func (c *Client) delete_(ctx context.Context, fam family, path string, p params, dest any, opts ...callOpt) error {
	call := apiCall{method: http.MethodDelete, family: fam, path: path, params: p, retryable: true, weight: 1}
	for _, o := range opts {
		o(&call)
	}
	return c.do(ctx, call, dest)
}

type callOpt func(*apiCall)

func signed() callOpt { return func(a *apiCall) { a.sec = secSigned } }

func apiKey() callOpt { return func(a *apiCall) { a.sec = secAPIKey } }

func weight(n int) callOpt { return func(a *apiCall) { a.weight = n } }

func noRetry() callOpt { return func(a *apiCall) { a.retryable = false } }

func recvWindow(ms int64) callOpt { return func(a *apiCall) { a.recvWindow = ms } }

func (c *Client) do(ctx context.Context, call apiCall, dest any) error {
	ctx, cancel := c.applyDefaultTimeout(ctx)
	defer cancel()

	if call.sec == secSigned && c.cfg.autoSync && !c.synced.Load() {
		if err := c.SyncTime(ctx); err != nil {
			c.cfg.log.Error("auto time sync failed", "err", err)
		}
	}

	attempts := c.cfg.retry.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	if !call.retryable {
		attempts = 1
	}

	var last error
	for n := 1; n <= attempts; n++ {
		if c.cfg.limiter != nil {
			w := call.weight
			if w <= 0 {
				w = 1
			}
			if err := c.cfg.limiter.Wait(ctx, w); err != nil {
				return err
			}
		}
		err := c.doOnce(ctx, call, dest)
		if err == nil {
			return nil
		}
		last = err
		if n == attempts || !c.shouldRetry(call, err) {
			return err
		}
		delay := c.cfg.retry.backoff(n)
		var rl *RateLimitError
		if errorsAsRateLimit(err, &rl) && rl.RetryAfter > 0 {
			delay = rl.RetryAfter
		}
		c.cfg.log.Info("retrying request", "path", call.path, "attempt", n, "delay", delay.String())
		if err := sleep(ctx, delay); err != nil {
			return err
		}
	}
	return last
}

func errorsAsRateLimit(err error, target **RateLimitError) bool {
	return errorsAs(err, target)
}

func (c *Client) shouldRetry(call apiCall, err error) bool {
	if !call.retryable {
		return false
	}
	if isTransientNetErr(err) {
		return true
	}
	var rl *RateLimitError
	if errorsAs(err, &rl) {
		return c.cfg.retry.Retry429 && rl.HTTPStatus == http.StatusTooManyRequests
	}
	var api *APIError
	if errorsAs(err, &api) {
		return isRetryableStatus(api.HTTPStatus)
	}
	var re *RequestError
	if errorsAs(err, &re) {
		return isTransientNetErr(re.Err)
	}
	return false
}

func (c *Client) doOnce(ctx context.Context, call apiCall, dest any) error {
	q, err := c.signedQuery(call)
	if err != nil {
		return err
	}
	base := c.endpoints.base(call.family)
	full := base + call.path
	if enc := q.Encode(); enc != "" {
		full += "?" + enc
	}
	req, err := http.NewRequestWithContext(ctx, call.method, full, nil)
	if err != nil {
		return &RequestError{Method: call.method, Path: call.path, Err: err}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.userAgent)
	for k, vs := range c.cfg.headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if call.sec != secNone {
		if c.cfg.apiKey == "" {
			return &AuthError{Msg: ErrMissingAPIKey.Error()}
		}
		req.Header.Set(headerAPIKey, c.cfg.apiKey)
	}
	if c.cfg.debug {
		c.cfg.log.Debug("http request",
			"method", call.method,
			"path", call.path,
			"query", RedactQuery(q.Encode()),
			"apiKey", redactHeader(headerAPIKey, req.Header.Get(headerAPIKey)),
		)
	}
	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return &RequestError{Method: call.method, Path: call.path, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return &RequestError{Method: call.method, Path: call.path, Err: err}
	}
	if w := parseUsedWeight(resp.Header); w > 0 {
		c.usedWt.Store(int64(w))
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 418 {
		apiErr := parseAPIError(resp.StatusCode, body)
		return &RateLimitError{
			HTTPStatus: resp.StatusCode,
			RetryAfter: parseRetryAfter(resp.Header),
			UsedWeight: parseUsedWeight(resp.Header),
			Err:        apiErr,
		}
	}
	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, body)
	}
	if dest == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return &DecodeError{Op: call.path, Err: err, Status: resp.StatusCode}
	}
	return nil
}

func parseAPIError(status int, body []byte) error {
	var api APIError
	if err := json.Unmarshal(body, &api); err != nil || (api.Code == 0 && api.Message == "") {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(status)
		}
		api.Message = msg
	}
	api.HTTPStatus = status
	return &api
}

func (c *Client) signedQuery(call apiCall) (url.Values, error) {
	q := call.params.values()
	if q == nil {
		q = url.Values{}
	}
	if call.sec != secSigned {
		return q, nil
	}
	if c.cfg.signer == nil {
		return nil, &AuthError{Msg: ErrMissingAPISecret.Error()}
	}
	q.Set("timestamp", strconv.FormatInt(c.nowMillis(), 10))
	rw := call.recvWindow
	if rw == 0 {
		rw = c.cfg.recvWindow
	}
	if rw > 0 {
		q.Set("recvWindow", strconv.FormatInt(rw, 10))
	}
	payload := q.Encode()
	sig, err := c.cfg.signer.Sign(payload)
	if err != nil {
		return nil, err
	}
	q.Set("signature", sig)
	return q, nil
}

func (c *Client) rawGet(ctx context.Context, fam family, path string, p params, opts ...callOpt) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.get(ctx, fam, path, p, &raw, opts...); err != nil {
		return nil, err
	}
	return raw, nil
}

func unmarshalOneOrMany[T any](raw json.RawMessage) ([]T, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[0] == '[' {
		var out []T
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	var one T
	if err := json.Unmarshal(raw, &one); err != nil {
		return nil, err
	}
	return []T{one}, nil
}
