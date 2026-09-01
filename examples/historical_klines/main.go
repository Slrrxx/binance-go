package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Slrrxx/binance-go"
)

func main() {
	client := binance.NewClient("", "")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	end := time.Now().UTC()
	start := end.Add(-2 * time.Hour)

	it := client.HistoricalKlines(binance.KlineQuery{
		Symbol:    "BTCUSDT",
		Interval:  binance.Interval1m,
		StartTime: start.UnixMilli(),
		EndTime:   end.UnixMilli(),
	})
	n := 0
	for it.Next(ctx) {
		k := it.Kline()
		n++
		if n <= 3 {
			fmt.Printf("%s open=%s close=%s\n", k.OpenTime.Time(), k.Open, k.Close)
		}
	}
	if err := it.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fetched %d candles without loading them all first\n", n)
}
