package binance

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// WSAPI is the spot WebSocket API (wss://ws-api.binance.com/ws-api/v3).
type WSAPI struct {
	c       *Client
	futures bool

	mu      sync.Mutex
	conn    *websocket.Conn
	pending map[string]chan wsAPIEnvelope
	readRun bool
}

type wsAPIEnvelope struct {
	ID     string          `json:"id"`
	Status int             `json:"status"`
	Result json.RawMessage `json:"result"`
	Error  *APIError       `json:"error"`
}

// FuturesWSAPI returns the USD-M futures WebSocket API client.
func (c *Client) FuturesWSAPI() *WSAPI {
	return &WSAPI{c: c, futures: true}
}

func (w *WSAPI) endpoint() string {
	if w.futures {
		return w.c.endpoints.FuturesWSAPI
	}
	return w.c.endpoints.SpotWSAPI
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}

func (w *WSAPI) do(ctx context.Context, method string, signed bool, p params, dest any) error {
	ctx, cancel := w.c.applyDefaultTimeout(ctx)
	defer cancel()
	id := newRequestID()
	params := map[string]string{}
	for k, vs := range p.values() {
		if len(vs) > 0 {
			params[k] = vs[0]
		}
	}
	if signed {
		if w.c.cfg.signer == nil {
			return &AuthError{Msg: ErrMissingAPISecret.Error()}
		}
		if w.c.cfg.apiKey == "" {
			return &AuthError{Msg: ErrMissingAPIKey.Error()}
		}
		params["apiKey"] = w.c.cfg.apiKey
		params["timestamp"] = strconv.FormatInt(w.c.nowMillis(), 10)
		if w.c.cfg.recvWindow > 0 {
			params["recvWindow"] = strconv.FormatInt(w.c.cfg.recvWindow, 10)
		}
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		sig, err := w.c.cfg.signer.Sign(q.Encode())
		if err != nil {
			return err
		}
		params["signature"] = sig
	}
	payload := map[string]any{"id": id, "method": method}
	if len(params) > 0 {
		payload["params"] = params
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ch, err := w.send(ctx, id, raw)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case env, ok := <-ch:
		if !ok {
			return &WebsocketError{Op: method, Err: ErrStreamClosed}
		}
		if env.Error != nil && (env.Error.Code != 0 || env.Error.Message != "") {
			env.Error.HTTPStatus = env.Status
			return env.Error
		}
		if env.Status >= 400 {
			return parseAPIError(env.Status, env.Result)
		}
		if dest == nil || len(env.Result) == 0 {
			return nil
		}
		if err := json.Unmarshal(env.Result, dest); err != nil {
			return &DecodeError{Op: method, Err: err, Status: env.Status}
		}
		return nil
	}
}

func (w *WSAPI) send(ctx context.Context, id string, raw []byte) (<-chan wsAPIEnvelope, error) {
	w.mu.Lock()
	if w.pending == nil {
		w.pending = make(map[string]chan wsAPIEnvelope)
	}
	if w.conn == nil {
		httpClient := w.c.cfg.httpClient
		if httpClient != nil && httpClient.Timeout > 0 {
			cloned := *httpClient
			cloned.Timeout = 0
			httpClient = &cloned
		}
		dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		conn, _, err := websocket.Dial(dialCtx, w.endpoint(), &websocket.DialOptions{
			HTTPClient: httpClient,
			HTTPHeader: http.Header{"User-Agent": []string{w.c.cfg.userAgent}},
		})
		cancel()
		if err != nil {
			w.mu.Unlock()
			return nil, &WebsocketError{Op: "wsapi dial", Err: err}
		}
		w.conn = conn
		if !w.readRun {
			w.readRun = true
			go w.readLoop()
		}
	}
	ch := make(chan wsAPIEnvelope, 1)
	w.pending[id] = ch
	conn := w.conn
	w.mu.Unlock()
	if err := conn.Write(ctx, websocket.MessageText, raw); err != nil {
		w.mu.Lock()
		delete(w.pending, id)
		_ = conn.Close(websocket.StatusInternalError, "write")
		w.conn = nil
		w.mu.Unlock()
		return nil, &WebsocketError{Op: "wsapi write", Err: err}
	}
	return ch, nil
}

func (w *WSAPI) readLoop() {
	for {
		w.mu.Lock()
		conn := w.conn
		w.mu.Unlock()
		if conn == nil {
			return
		}
		_, data, err := conn.Read(context.Background())
		if err != nil {
			w.failAll(err)
			return
		}
		var env wsAPIEnvelope
		if err := json.Unmarshal(data, &env); err != nil {
			continue
		}
		w.mu.Lock()
		ch, ok := w.pending[env.ID]
		if ok {
			delete(w.pending, env.ID)
		}
		w.mu.Unlock()
		if ok {
			ch <- env
			close(ch)
		}
	}
}

func (w *WSAPI) failAll(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn != nil {
		_ = w.conn.Close(websocket.StatusGoingAway, "")
		w.conn = nil
	}
	w.readRun = false
	for id, ch := range w.pending {
		ch <- wsAPIEnvelope{ID: id, Status: 503, Error: &APIError{Message: err.Error(), HTTPStatus: 503}}
		close(ch)
		delete(w.pending, id)
	}
}

// Close closes the underlying WebSocket API connection.
func (w *WSAPI) Close() error {
	w.failAll(ErrStreamClosed)
	return nil
}

// Ping tests WebSocket API connectivity.
func (w *WSAPI) Ping(ctx context.Context) error {
	return w.do(ctx, "ping", false, nil, nil)
}

// ServerTime returns exchange time over the WebSocket API.
func (w *WSAPI) ServerTime(ctx context.Context) (int64, error) {
	var st ServerTime
	method := "time"
	if w.futures {
		method = "time"
	}
	if err := w.do(ctx, method, false, nil, &st); err != nil {
		return 0, err
	}
	return st.ServerTime, nil
}

// CreateOrder places an order over the WebSocket API. Not retried.
func (w *WSAPI) CreateOrder(ctx context.Context, req OrderRequest) (*Order, error) {
	var out Order
	if err := w.do(ctx, "order.place", true, req.params(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TestOrder validates an order over the WebSocket API.
func (w *WSAPI) TestOrder(ctx context.Context, req OrderRequest) error {
	return w.do(ctx, "order.test", true, req.params(), nil)
}

// CancelOrder cancels an order over the WebSocket API.
func (w *WSAPI) CancelOrder(ctx context.Context, req CancelOrderRequest) (*Order, error) {
	var out Order
	if err := w.do(ctx, "order.cancel", true, req.params(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrder queries an order over the WebSocket API.
func (w *WSAPI) GetOrder(ctx context.Context, req QueryOrderRequest) (*Order, error) {
	var out Order
	if err := w.do(ctx, "order.status", true, req.params(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAccount returns spot account information over the WebSocket API.
func (w *WSAPI) GetAccount(ctx context.Context) (*Account, error) {
	var out Account
	if err := w.do(ctx, "account.status", true, newParams(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSymbolTicker returns a price ticker over the WebSocket API.
func (w *WSAPI) GetSymbolTicker(ctx context.Context, symbol string) (*SymbolPrice, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out SymbolPrice
	if err := w.do(ctx, "ticker.price", false, p, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFuturesOrder places a USD-M order over the futures WebSocket API.
func (w *WSAPI) CreateFuturesOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := w.do(ctx, "order.place", true, req.params(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
