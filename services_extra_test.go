package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEarnConvertGiftCardAndGenerated(t *testing.T) {
	var paths []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/sapi/v1/simple-earn/flexible/list":
			_ = json.NewEncoder(w).Encode(EarnPage{Total: 1, Rows: []EarnProduct{{ProductID: "p1", Asset: "USDT"}}})
		case "/sapi/v1/simple-earn/flexible/subscribe":
			_ = json.NewEncoder(w).Encode(UniversalTransferResult{TranID: 3})
		case "/sapi/v1/convert/exchangeInfo":
			_ = json.NewEncoder(w).Encode([]ConvertPair{{FromAsset: "USDT", ToAsset: "BTC"}})
		case "/sapi/v1/convert/getQuote":
			_ = json.NewEncoder(w).Encode(ConvertQuote{QuoteID: "q1", FromAmount: "10"})
		case "/sapi/v1/giftcard/verify":
			_, _ = w.Write([]byte(`{"valid":true}`))
		case "/sapi/v1/asset/dribblet":
			_, _ = w.Write([]byte(`{"total":0}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer ts.Close()
	c := testClient(ts)
	ctx := context.Background()

	page, err := c.Earn().FlexibleProducts(ctx, "USDT", 0)
	if err != nil || page.Total != 1 {
		t.Fatalf("earn list: %+v %v", page, err)
	}
	sub, err := c.Earn().SubscribeFlexible(ctx, "p1", "10", 0)
	if err != nil || sub.TranID != 3 {
		t.Fatalf("earn sub: %+v %v", sub, err)
	}
	pairs, err := c.Convert().ExchangeInfo(ctx, "USDT", "BTC")
	if err != nil || len(pairs) != 1 {
		t.Fatalf("convert info: %+v %v", pairs, err)
	}
	q, err := c.Convert().GetQuote(ctx, "USDT", "BTC", "10", 0)
	if err != nil || q.QuoteID != "q1" {
		t.Fatalf("quote: %+v %v", q, err)
	}
	raw, err := c.GiftCard().Verify(ctx, "ref", 0)
	if err != nil || string(raw) != `{"valid":true}` {
		t.Fatalf("gift: %s %v", raw, err)
	}
	dust, err := c.Wallet().DustLog(ctx, 0, map[string]string{"accountType": "SPOT"})
	if err != nil || string(dust) != `{"total":0}` {
		t.Fatalf("dust: %s %v", dust, err)
	}

	want := []string{
		"GET /sapi/v1/simple-earn/flexible/list",
		"POST /sapi/v1/simple-earn/flexible/subscribe",
		"GET /sapi/v1/convert/exchangeInfo",
		"POST /sapi/v1/convert/getQuote",
		"GET /sapi/v1/giftcard/verify",
		"GET /sapi/v1/asset/dribblet",
	}
	if len(paths) != len(want) {
		t.Fatalf("paths=%v", paths)
	}
	for i, p := range want {
		if paths[i] != p {
			t.Fatalf("path[%d]=%s want %s", i, paths[i], p)
		}
	}
}

func TestGiftCardCreateNotRetried(t *testing.T) {
	hits := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":-1001,"msg":"boom"}`))
	}))
	defer ts.Close()
	c := testClient(ts)
	_, err := c.GiftCard().CreateCode(context.Background(), "USDT", "10", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if hits != 1 {
		t.Fatalf("retries=%d", hits)
	}
}
