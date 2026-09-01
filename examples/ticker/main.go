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
	client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_API_SECRET"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ticker, err := client.GetSymbolTicker(ctx, "BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ticker.Symbol, ticker.Price)
}
