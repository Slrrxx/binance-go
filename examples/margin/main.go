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
	client := binance.NewClient(must("BINANCE_API_KEY"), must("BINANCE_API_SECRET"))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	acc, err := client.Margin().Account(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("marginLevel=%s trade=%v assets=%d\n", acc.MarginLevel, acc.TradeEnabled, len(acc.UserAssets))
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}
