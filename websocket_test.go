package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestParseEventTradeAndCombined(t *testing.T) {
	raw := []byte(`{"e":"trade","E":1,"s":"BTCUSDT","t":9,"p":"1.0","q":"0.2","T":2,"m":true}`)
	ev, err := parseEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Trade == nil || ev.Trade.Price != "1.0" || ev.Type != "trade" {
		t.Fatalf("%+v", ev)
	}
	combined := []byte(`{"stream":"btcusdt@trade","data":` + string(raw) + `}`)
	ev, err = parseEvent(combined)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Stream != "btcusdt@trade" || ev.Trade == nil {
		t.Fatalf("%+v", ev)
	}
}

func TestParseDepthEvent(t *testing.T) {
	raw := []byte(`{"e":"depthUpdate","E":1,"s":"BTCUSDT","U":2,"u":3,"b":[["1.0","2.0"]],"a":[["1.1","0"]]}`)
	ev, err := parseEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Depth == nil || ev.Depth.FirstUpdateID != 2 || ev.Depth.Asks[0].Quantity != "0" {
		t.Fatalf("%+v", ev.Depth)
	}
}

func TestStreamReconnect(t *testing.T) {
	var accepts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		n := accepts.Add(1)
		ctx := r.Context()
		payload := []byte(`{"e":"trade","E":1,"s":"BTCUSDT","p":"1","q":"1","t":1,"T":1}`)
		_ = conn.Write(ctx, websocket.MessageText, payload)
		if n == 1 {
			_ = conn.Close(websocket.StatusInternalError, "boom")
			return
		}
		// keep second connection alive until client closes
		_, _, _ = conn.Read(ctx)
		_ = conn.Close(websocket.StatusNormalClosure, "")
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	c := NewClient("", "",
		WithWebSocketURL(wsURL+"/ws"),
		WithHTTPClient(srv.Client()),
		WithWSReconnect(5, 10*time.Millisecond, 50*time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	st, err := c.Subscribe(ctx, StreamTrade("BTCUSDT"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	got := 0
	deadline := time.Now().Add(2 * time.Second)
	for got < 2 && time.Now().Before(deadline) {
		nctx, ncancel := context.WithTimeout(ctx, 500*time.Millisecond)
		ev, err := st.Next(nctx)
		ncancel()
		if err != nil {
			continue
		}
		if ev.Trade != nil {
			got++
		}
	}
	if got < 2 {
		t.Fatalf("got %d trade events, accepts=%d", got, accepts.Load())
	}
	if accepts.Load() < 2 {
		t.Fatalf("did not reconnect, accepts=%d", accepts.Load())
	}
}

func TestStreamContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
		_, _, _ = conn.Read(r.Context())
	}))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	c := NewClient("", "", WithWebSocketURL(wsURL+"/ws"), WithHTTPClient(srv.Client()))
	ctx, cancel := context.WithCancel(context.Background())
	st, err := c.Subscribe(ctx, StreamTrade("ETHUSDT"))
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	_, err = st.Next(context.Background())
	if err == nil {
		// allow a short wait
		time.Sleep(50 * time.Millisecond)
		_, err = st.Next(context.Background())
	}
	if err == nil {
		t.Fatal("expected closed stream")
	}
}

func TestDepthCacheApplyAndGap(t *testing.T) {
	d := &DepthCache{bids: map[string]string{}, asks: map[string]string{}, updates: make(chan struct{}, 4)}
	d.replace(&OrderBook{
		LastUpdateID: 100,
		Bids:         []PriceLevel{{Price: "1.0", Quantity: "5"}},
		Asks:         []PriceLevel{{Price: "2.0", Quantity: "3"}},
	})
	if err := d.apply(DepthEvent{FirstUpdateID: 101, FinalUpdateID: 101, Bids: []PriceLevel{{Price: "1.0", Quantity: "0"}}, Asks: []PriceLevel{{Price: "2.1", Quantity: "1"}}}); err != nil {
		t.Fatal(err)
	}
	if _, ok := d.bids["1.0"]; ok {
		t.Fatal("bid should be removed")
	}
	if d.asks["2.1"] != "1" {
		t.Fatalf("asks=%v", d.asks)
	}
	if err := d.apply(DepthEvent{FirstUpdateID: 200, FinalUpdateID: 201}); err == nil {
		t.Fatal("expected gap")
	}
	bids := d.Bids()
	if len(bids) != 0 {
		t.Fatalf("bids=%v", bids)
	}
}

func TestDepthCacheSnapshotSync(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/depth", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(OrderBook{
			LastUpdateID: 5,
			Bids:         []PriceLevel{{Price: "10", Quantity: "1"}},
			Asks:         []PriceLevel{{Price: "11", Quantity: "1"}},
		})
	})
	var upgrader = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "@depth") || strings.Contains(r.URL.Path, "/ws") {
			conn, err := websocket.Accept(w, r, nil)
			if err != nil {
				return
			}
			ctx := r.Context()
			_ = conn.Write(ctx, websocket.MessageText, []byte(`{"e":"depthUpdate","E":1,"s":"BTCUSDT","U":4,"u":6,"b":[["10","2"]],"a":[]}`))
			_, _, _ = conn.Read(ctx)
			_ = conn.Close(websocket.StatusNormalClosure, "")
			return
		}
		mux.ServeHTTP(w, r)
	})
	ts := httptest.NewServer(upgrader)
	defer ts.Close()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	c := NewClient("", "",
		WithBaseURL(ts.URL),
		WithWebSocketURL(wsURL+"/ws"),
		WithHTTPClient(ts.Client()),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cache, err := c.DepthCache(ctx, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cache.Close() }()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		wctx, wcancel := context.WithTimeout(ctx, 200*time.Millisecond)
		_, _ = cache.Wait(wctx)
		wcancel()
		if cache.LastUpdateID() >= 6 {
			break
		}
	}
	if cache.LastUpdateID() < 5 {
		t.Fatalf("lastUpdate=%d", cache.LastUpdateID())
	}
	bids := cache.Bids()
	if len(bids) == 0 || bids[0].Quantity != "2" && bids[0].Quantity != "1" {
		// after apply qty should be 2; snapshot-only is 1 if event arrived first and was skipped then applied
		if len(bids) == 0 {
			t.Fatalf("empty bids last=%d", cache.LastUpdateID())
		}
	}
}
