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

	stream, err := client.Subscribe(ctx, binance.StreamDepth("BTCUSDT"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = stream.Close() }()

	for i := 0; i < 5; i++ {
		ev, err := stream.Next(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if ev.Depth != nil {
			fmt.Printf("u=%d bids=%d asks=%d\n", ev.Depth.FinalUpdateID, len(ev.Depth.Bids), len(ev.Depth.Asks))
		}
	}
}
