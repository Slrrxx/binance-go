package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestWSAPIPingAndOrder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		for {
			_, data, err := conn.Read(r.Context())
			if err != nil {
				return
			}
			var req struct {
				ID     string         `json:"id"`
				Method string         `json:"method"`
				Params map[string]any `json:"params"`
			}
			if err := json.Unmarshal(data, &req); err != nil {
				t.Errorf("bad req: %v", err)
				return
			}
			var result any
			switch req.Method {
			case "ping":
				result = map[string]any{}
			case "order.place":
				if req.Params["signature"] == nil || req.Params["apiKey"] != "key" {
					t.Errorf("unsigned order: %#v", req.Params)
				}
				result = Order{Symbol: "BTCUSDT", OrderID: 7, Status: OrderStatusNew}
			default:
				t.Errorf("unexpected method %s", req.Method)
			}
			raw, _ := json.Marshal(map[string]any{"id": req.ID, "status": 200, "result": result})
			if err := conn.Write(r.Context(), websocket.MessageText, raw); err != nil {
				return
			}
		}
	}))
	defer srv.Close()
	ws := "ws" + strings.TrimPrefix(srv.URL, "http")
	c := NewClient("key", "secret", WithWSAPIURL(ws), WithHTTPClient(srv.Client()), WithTimeout(2*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.WSAPI().Ping(ctx); err != nil {
		t.Fatal(err)
	}
	o, err := c.WSAPI().CreateOrder(ctx, OrderRequest{Symbol: "BTCUSDT", Side: SideBuy, Type: OrderTypeMarket, Quantity: "0.001"})
	if err != nil {
		t.Fatal(err)
	}
	if o.OrderID != 7 {
		t.Fatalf("%+v", o)
	}
}

func TestWithProxySetsTransport(t *testing.T) {
	c := NewClient("", "", WithProxy("http://127.0.0.1:6553"))
	tr, ok := c.cfg.httpClient.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Fatal("proxy not applied")
	}
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	u, err := tr.Proxy(req)
	if err != nil || u == nil || u.Host != "127.0.0.1:6553" {
		t.Fatalf("proxy=%v err=%v", u, err)
	}
}

func TestWebSocketFacadeSubscribe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		_ = conn.Write(r.Context(), websocket.MessageText, []byte(`{"e":"trade","E":1,"s":"BTCUSDT","p":"1","q":"1","t":1,"T":1}`))
		_, _, _ = conn.Read(r.Context())
	}))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	c := NewClient("", "", WithWebSocketURL(wsURL+"/ws"), WithHTTPClient(srv.Client()))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	st, err := c.WebSocket().Trade(ctx, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	ev, err := st.Next(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Trade == nil {
		t.Fatalf("%+v", ev)
	}
}

func TestPortfolioAndOptionsPaths(t *testing.T) {
	var paths []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	c := NewClient("k", "s",
		WithHTTPClient(ts.Client()),
		WithPortfolioURL(ts.URL),
		WithOptionsURL(ts.URL),
	)
	_ = c.Portfolio().Ping(context.Background())
	_ = c.Options().Ping(context.Background())
	if len(paths) != 2 || paths[0] != "/papi/v1/ping" || paths[1] != "/eapi/v1/ping" {
		t.Fatalf("paths=%v", paths)
	}
}
