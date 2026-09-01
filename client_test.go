package binance

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testClient(ts *httptest.Server, opts ...Option) *Client {
	all := []Option{WithBaseURL(ts.URL), WithHTTPClient(ts.Client()), WithTimeout(2 * time.Second)}
	all = append(all, opts...)
	return NewClient("key", "secret", all...)
}

func TestPingAndTicker(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{}"))
	})
	mux.HandleFunc("/api/v3/ticker/price", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("symbol=%s", r.URL.Query().Get("symbol"))
		}
		_ = json.NewEncoder(w).Encode(SymbolPrice{Symbol: "BTCUSDT", Price: "50000.12"})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	c := testClient(ts)
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	tk, err := c.GetSymbolTicker(context.Background(), "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Price != "50000.12" {
		t.Fatalf("price=%s", tk.Price)
	}
}

func TestAPIErrorAs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"code":-2010,"msg":"insufficient balance"}`))
	}))
	defer ts.Close()
	c := testClient(ts)
	_, err := c.GetAccount(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var api *APIError
	if !errors.As(err, &api) {
		t.Fatalf("want APIError got %T %v", err, err)
	}
	if api.Code != -2010 || api.Message != "insufficient balance" || api.HTTPStatus != 400 {
		t.Fatalf("%+v", api)
	}
}

func TestCreateOrderSignsAndDoesNotRetry(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get(headerAPIKey) != "key" {
			t.Errorf("missing api key")
		}
		q := r.URL.Query()
		if q.Get("signature") == "" || q.Get("timestamp") == "" {
			t.Errorf("unsigned: %s", r.URL.RawQuery)
		}
		if strings.Contains(r.URL.RawQuery, "secret") {
			t.Error("secret leaked in query")
		}
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"code":-1001,"msg":"disconnected"}`))
	}))
	defer ts.Close()
	c := testClient(ts, WithRetry(RetryConfig{MaxAttempts: 5, MinBackoff: time.Millisecond, Retry429: true}))
	_, err := c.CreateOrder(context.Background(), OrderRequest{
		Symbol:   "BTCUSDT",
		Side:     SideBuy,
		Type:     OrderTypeMarket,
		Quantity: "0.001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if hits.Load() != 1 {
		t.Fatalf("order retried: hits=%d", hits.Load())
	}
}

func TestIdempotentGetRetries503(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) < 3 {
			w.WriteHeader(503)
			_, _ = w.Write([]byte(`{"code":-1001,"msg":"disconnected"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	c := testClient(ts, WithRetry(RetryConfig{MaxAttempts: 4, MinBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}))
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits=%d", hits.Load())
	}
}

func TestRateLimit429(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerRetryAfter, "1")
		w.Header().Set(headerUsedWeight, "1200")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"code":-1003,"msg":"too many requests"}`))
	}))
	defer ts.Close()
	c := testClient(ts, WithRetry(RetryConfig{MaxAttempts: 1}))
	err := c.Ping(context.Background())
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("want RateLimitError got %T %v", err, err)
	}
	if rl.RetryAfter != time.Second || rl.UsedWeight != 1200 {
		t.Fatalf("%+v", rl)
	}
}

func TestContextCancel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer ts.Close()
	c := testClient(ts)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := c.Ping(ctx)
	if err == nil || !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("want canceled, got %v", err)
	}
}

func TestKlineUnmarshalAndPagination(t *testing.T) {
	var pages atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/klines" {
			http.NotFound(w, r)
			return
		}
		n := pages.Add(1)
		start := r.URL.Query().Get("startTime")
		if n == 1 {
			if start != "1000" {
				t.Errorf("startTime=%s", start)
			}
			_, _ = w.Write([]byte(`[[1000,"1","2","0.5","1.5","10",1999,"10",2,"5","5","0"]]`))
			return
		}
		if n == 2 {
			_, _ = w.Write([]byte(`[[2000,"1","2","0.5","1.5","10",2999,"10",2,"5","5","0"]]`))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer ts.Close()
	c := testClient(ts)
	it := c.HistoricalKlines(KlineQuery{Symbol: "BTCUSDT", Interval: Interval1m, StartTime: 1000, Limit: 1})
	var klines []Kline
	for it.Next(context.Background()) {
		klines = append(klines, it.Kline())
	}
	if err := it.Err(); err != nil {
		t.Fatal(err)
	}
	if len(klines) != 2 {
		t.Fatalf("len=%d", len(klines))
	}
	if klines[0].Open != "1" || klines[1].OpenTime != 2000 {
		t.Fatalf("%+v", klines)
	}
	if pages.Load() < 2 {
		t.Fatalf("pages=%d", pages.Load())
	}
}

func TestExchangeInfoAndOrderBook(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v3/exchangeInfo", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ExchangeInfo{
			Timezone:   "UTC",
			ServerTime: 1,
			Symbols:    []SymbolInfo{{Symbol: "BTCUSDT", Status: "TRADING", BaseAsset: "BTC", QuoteAsset: "USDT"}},
		})
	})
	mux.HandleFunc("/api/v3/depth", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"lastUpdateId":10,"bids":[["1.0","2.0"]],"asks":[["1.1","3.0"]]}`))
	})
	mux.HandleFunc("/api/v3/time", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ServerTime{ServerTime: 1_700_000_000_000})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	c := testClient(ts)
	info, err := c.GetExchangeInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(info.Symbols) != 1 || info.Symbols[0].Symbol != "BTCUSDT" {
		t.Fatalf("%+v", info)
	}
	book, err := c.GetOrderBook(context.Background(), "BTCUSDT", 5)
	if err != nil {
		t.Fatal(err)
	}
	if book.LastUpdateID != 10 || book.Bids[0].Price != "1.0" {
		t.Fatalf("%+v", book)
	}
	if err := c.SyncTime(context.Background()); err != nil {
		t.Fatal(err)
	}
	if c.TimeOffset() == 0 {
		// possible but unlikely; offset is server - local
	}
}

func TestAccountAndUsedWeight(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("signature") == "" {
			t.Error("unsigned account")
		}
		w.Header().Set(headerUsedWeight, "42")
		_ = json.NewEncoder(w).Encode(Account{
			CanTrade: true,
			Balances: []Balance{{Asset: "BTC", Free: "0.1", Locked: "0"}},
		})
	}))
	defer ts.Close()
	c := testClient(ts)
	acc, err := c.GetAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b, ok := acc.BalanceOf("BTC")
	if !ok || b.Free != "0.1" {
		t.Fatalf("%+v", acc)
	}
	if c.UsedWeight() != 42 {
		t.Fatalf("used=%d", c.UsedWeight())
	}
}

func TestDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "not-json")
	}))
	defer ts.Close()
	c := testClient(ts)
	_, err := c.GetSymbolTicker(context.Background(), "BTCUSDT")
	var de *DecodeError
	if !errors.As(err, &de) {
		t.Fatalf("want DecodeError got %T %v", err, err)
	}
}

func TestWeightLimiter(t *testing.T) {
	l := NewWeightLimiter(2, 50*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := l.Wait(ctx, 2); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if err := l.Wait(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) < 10*time.Millisecond {
		t.Fatal("expected limiter to wait")
	}
}

func TestFuturesOrderUsesFAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/order" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(FuturesOrder{Symbol: "BTCUSDT", OrderID: 9, Status: OrderStatusNew})
	}))
	defer ts.Close()
	c := NewClient("key", "secret", WithFuturesBaseURL(ts.URL), WithHTTPClient(ts.Client()))
	o, err := c.Futures().CreateOrder(context.Background(), FuturesOrderRequest{
		Symbol: "BTCUSDT", Side: SideBuy, Type: OrderTypeMarket, Quantity: "0.01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if o.OrderID != 9 {
		t.Fatalf("%+v", o)
	}
}

func TestMarginBorrowNotRetried(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"code":-1001,"msg":"x"}`))
	}))
	defer ts.Close()
	c := testClient(ts, WithRetry(RetryConfig{MaxAttempts: 4, MinBackoff: time.Millisecond}))
	_, err := c.Margin().Borrow(context.Background(), BorrowRepayRequest{Asset: "USDT", Amount: "1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if hits.Load() != 1 {
		t.Fatalf("hits=%d", hits.Load())
	}
}
