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

	err := client.TestOrder(ctx, binance.OrderRequest{
		Symbol:   "BTCUSDT",
		Side:     binance.SideBuy,
		Type:     binance.OrderTypeMarket,
		Quantity: "0.001",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("test order accepted")
}
