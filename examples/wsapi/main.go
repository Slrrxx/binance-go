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
	ws := client.WSAPI()
	defer func() { _ = ws.Close() }()

	ctx := context.Background()
	if err := ws.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	ticker, err := ws.GetSymbolTicker(ctx, "BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ticker.Symbol, ticker.Price)
}
