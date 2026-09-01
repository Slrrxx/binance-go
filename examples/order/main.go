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
		mustEnv("BINANCE_API_KEY"),
		mustEnv("BINANCE_API_SECRET"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	order, err := client.CreateOrder(ctx, binance.OrderRequest{
		Symbol:        "BTCUSDT",
		Side:          binance.SideBuy,
		Type:          binance.OrderTypeMarket,
		Quantity:      envOr("BINANCE_QTY", "0.001"),
		ClientOrderID: binance.NewClientOrderID(),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("order id=%d status=%s executed=%s\n", order.OrderID, order.Status, order.ExecutedQty)
}

func mustEnv(k string) string {
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
