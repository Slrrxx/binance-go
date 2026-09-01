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
	client := binance.NewClient(
		os.Getenv("BINANCE_API_KEY"),
		os.Getenv("BINANCE_API_SECRET"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	acc, err := client.GetAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("canTrade=%v accountType=%s balances=%d\n", acc.CanTrade, acc.AccountType, len(acc.Balances))
	if b, ok := acc.BalanceOf("USDT"); ok {
		fmt.Printf("USDT free=%s locked=%s\n", b.Free, b.Locked)
	}
}
