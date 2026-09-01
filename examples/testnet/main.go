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
		os.Getenv("BINANCE_API_KEY"),
		os.Getenv("BINANCE_API_SECRET"),
		binance.WithTestnet(),
		binance.WithTimeout(10*time.Second),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	ticker, err := client.GetSymbolTicker(ctx, "BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("testnet ticker", ticker.Price)

	if os.Getenv("BINANCE_API_KEY") != "" {
		if err := client.TestOrder(ctx, binance.OrderRequest{
			Symbol:   "BTCUSDT",
			Side:     binance.SideBuy,
			Type:     binance.OrderTypeMarket,
			Quantity: "0.001",
		}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("testnet test order ok")
	}
}
