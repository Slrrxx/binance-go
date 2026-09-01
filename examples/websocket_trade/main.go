package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Slrrxx/binance-go"
)

func main() {
	client := binance.NewClient("", "")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	stream, err := client.Subscribe(ctx, binance.StreamTrade("BTCUSDT"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = stream.Close() }()

	for {
		ev, err := stream.Next(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		if ev.Trade != nil {
			fmt.Printf("%s %s qty=%s\n", ev.Trade.Symbol, ev.Trade.Price, ev.Trade.Quantity)
		}
	}
}
