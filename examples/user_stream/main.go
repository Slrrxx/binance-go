package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Slrrxx/binance-go"
)

func main() {
	client := binance.NewClient(
		must("BINANCE_API_KEY"),
		must("BINANCE_API_SECRET"),
	)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	stream, err := client.UserData(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = stream.Close() }()

	fmt.Println("listening for user-data events; Ctrl+C to exit")
	for {
		ev, err := stream.Next(ctx)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%s %s %s\n", ev.Type, ev.EventTime.Format(time.RFC3339), ev.Raw)
	}
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}
