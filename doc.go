// Package binance is a production-oriented unofficial Binance REST and
// WebSocket SDK for Go.
//
// The public API is intentionally similar to python-binance in spirit
// (one client, automatic timestamps, testnet, historical klines, depth
// cache) while remaining idiomatic Go: context on every network call,
// typed structs, string decimals, and structured errors.
//
//	client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_API_SECRET"))
//	ticker, err := client.GetSymbolTicker(ctx, "BTCUSDT")
//	stream, err := client.WebSocket().Trade(ctx, "BTCUSDT")
package binance
