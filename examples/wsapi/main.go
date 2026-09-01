package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Slrrxx/binance-go"
)

func main() {
	client := binance.NewClient(
		os.Getenv("BINANCE_API_KEY"),
		os.Getenv("BINANCE_API_SECRET"),
		binance.WithTestnet(),
	)
	defer client.WSAPI().Close()

	ctx := context.Background()
	if err := client.WSAPI().Ping(ctx); err != nil {
		log.Fatal(err)
	}
	ticker, err := client.WSAPI().GetSymbolTicker(ctx, "BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ticker.Symbol, ticker.Price)
}
