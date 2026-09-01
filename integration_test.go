//go:build integration

package binance_test

import (
	"context"
	"os"
	"testing"
	"time"

	binance "github.com/Slrrxx/binance-go"
)

func TestIntegrationPublicTicker(t *testing.T) {
	if os.Getenv("BINANCE_INTEGRATION") == "" {
		t.Skip("set BINANCE_INTEGRATION=1 to run live public tests")
	}
	c := binance.NewClient("", "", binance.WithTimeout(10*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	tk, err := c.GetSymbolTicker(ctx, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	if tk.Price == "" {
		t.Fatal("empty price")
	}
}
