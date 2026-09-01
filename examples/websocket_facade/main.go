package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Slrrxx/binance-go"
)

func main() {
	client := binance.NewClient("", "")
	stream, err := client.WebSocket().Trade(context.Background(), "BTCUSDT")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = stream.Close() }()

	ev, err := stream.Next(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if ev.Trade != nil {
		fmt.Println(ev.Trade.Price, ev.Trade.Quantity)
	}
}
