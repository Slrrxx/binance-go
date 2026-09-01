package binance_test

import (
	"os"

	"github.com/Slrrxx/binance-go"
)

func ExampleNewClient() {
	client := binance.NewClient(
		os.Getenv("BINANCE_API_KEY"),
		os.Getenv("BINANCE_API_SECRET"),
		binance.WithTestnet(),
		binance.WithProxy("http://127.0.0.1:7890"),
	)
	_ = client.Spot()
	_ = client.Futures()
	_ = client.CoinFutures()
	_ = client.Margin()
	_ = client.Wallet()
	_ = client.WebSocket()
	_ = client.WSAPI()
	_ = client.Portfolio()
	_ = client.Options()
	_ = client.Earn()
	_ = client.Convert()
	_ = client.GiftCard()
}
