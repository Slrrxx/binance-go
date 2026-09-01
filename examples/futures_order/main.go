package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Slrrxx/binance-go"
)

func main() {
	client := binance.NewClient(
		must("BINANCE_API_KEY"),
		must("BINANCE_API_SECRET"),
		binance.WithTestnet(),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	order, err := client.Futures().CreateOrder(ctx, binance.FuturesOrderRequest{
		Symbol:        "BTCUSDT",
		Side:          binance.SideBuy,
		Type:          binance.OrderTypeMarket,
		Quantity:      envOr("BINANCE_QTY", "0.001"),
		ClientOrderID: binance.NewClientOrderID(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("futures order id=%d status=%s\n", order.OrderID, order.Status)
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
