# binance-go

[![CI](https://github.com/Slrrxx/binance-go/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Slrrxx/binance-go/actions/workflows/ci.yml?query=branch%3Amain)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/Slrrxx/binance-go.svg)](https://pkg.go.dev/github.com/Slrrxx/binance-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![codecov](https://codecov.io/gh/Slrrxx/binance-go/graph/badge.svg)](https://codecov.io/gh/Slrrxx/binance-go)

Unofficial **Binance REST + WebSocket SDK** for Go.

One client, automatic timestamps, typed models, string decimals, and reconnecting streams — the python-binance experience, written as idiomatic Go.

```bash
go get github.com/Slrrxx/binance-go
```

> Not affiliated with Binance. Trading and withdrawals can lose money. Use at your own risk.

## Why this SDK

| | |
|---|---|
| **Day-one DX** | `NewClient` → ticker, account, or order in a few lines |
| **Safe money types** | prices and quantities stay `string` — no `float64` |
| **Production HTTP** | `context.Context`, retries that skip order creates, 429 / 418 handling |
| **Live markets** | typed WebSocket events, reconnect + backoff, depth cache |
| **Tested in CI** | `gofmt`, `vet`, `test`, `-race`, lint — no live exchange in default CI |

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/Slrrxx/binance-go"
)

func main() {
    client := binance.NewClient(
        os.Getenv("BINANCE_API_KEY"),
        os.Getenv("BINANCE_API_SECRET"),
    )

    ticker, err := client.GetSymbolTicker(context.Background(), "BTCUSDT")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(ticker.Price)
}
```

```go
account, err := client.GetAccount(ctx)

order, err := client.CreateOrder(ctx, binance.OrderRequest{
    Symbol:        "BTCUSDT",
    Side:          binance.SideBuy,
    Type:          binance.OrderTypeMarket,
    Quantity:      "0.001",
    ClientOrderID: binance.NewClientOrderID(),
})
```

Public market data works with empty credentials. Signed calls need `BINANCE_API_KEY` / `BINANCE_API_SECRET`.

## Features

- Spot market data, trading, account, and user-data stream
- USD-M and COIN-M futures, portfolio margin (papi), vanilla options (eapi)
- Cross / isolated margin and wallet (withdrawals documented as dangerous)
- Simple Earn, convert, and gift card
- HMAC-SHA256, RSA, and Ed25519 signing
- Automatic timestamps, configurable `recvWindow`, optional server-time sync
- Testnet, demo, TLD, and HTTP proxy: `WithTestnet()`, `WithDemo()`, `WithProxy`
- Historical klines — slice helper **and** memory-efficient iterator
- Reconnecting market WebSockets (`client.WebSocket().Trade`) with typed events
- WebSocket API trading (`client.WSAPI()`, `client.FuturesWSAPI()`)
- Thread-safe depth cache (snapshot + incremental sync) for spot and futures
- Optional retry and rate limiter; secrets never logged

## Authentication

```go
client := binance.NewClient(
    os.Getenv("BINANCE_API_KEY"),
    os.Getenv("BINANCE_API_SECRET"),
    binance.WithRecvWindow(5000),
    binance.WithAutoTimeSync(),
    binance.WithTestnet(),
)
```

RSA / Ed25519:

```go
client := binance.NewClient(apiKey, "",
    binance.WithEd25519PrivateKey(pemBytes, nil),
)
```

## Market data · orders · streams

```go
info, _ := client.GetExchangeInfo(ctx)
book, _ := client.GetOrderBook(ctx, "BTCUSDT", 100)

err := client.TestOrder(ctx, binance.OrderRequest{
    Symbol: "BTCUSDT", Side: binance.SideBuy,
    Type: binance.OrderTypeLimit, TimeInForce: binance.TimeInForceGTC,
    Quantity: "0.001", Price: "50000",
})
```

```go
stream, err := client.WebSocket().Trade(ctx, "BTCUSDT")
// or: client.Subscribe(ctx, binance.StreamTrade("BTCUSDT"))
defer stream.Close()

for {
    ev, err := stream.Next(ctx)
    if err != nil {
        break
    }
    if ev.Trade != nil {
        fmt.Println(ev.Trade.Price, ev.Trade.Quantity)
    }
}
```

```go
it := client.HistoricalKlines(binance.KlineQuery{
    Symbol: "BTCUSDT", Interval: binance.Interval1m, StartTime: start, EndTime: end,
})
for it.Next(ctx) {
    _ = it.Kline()
}
```

```go
cache, err := client.DepthCache(ctx, "BTCUSDT")
defer cache.Close()
_, _ = cache.Wait(ctx)
bids, asks := cache.Bids(), cache.Asks()
```

Futures: `client.Futures()` · COIN-M: `client.CoinFutures()` · margin / wallet: `client.Margin()`, `client.Wallet()` · portfolio / options: `client.Portfolio()`, `client.Options()` · earn / convert: `client.Earn()`, `client.Convert()`.

```go
err := client.WSAPI().Ping(ctx)
order, err := client.WSAPI().CreateOrder(ctx, binance.OrderRequest{
    Symbol: "BTCUSDT", Side: binance.SideBuy, Type: binance.OrderTypeMarket, Quantity: "0.001",
})
```

Order creates over the WebSocket API are **not retried**.

## Errors

```go
var apiErr *binance.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.Code, apiErr.Message, apiErr.HTTPStatus)
}
```

Also: `AuthError`, `DecodeError`, `RequestError`, `RateLimitError`, `WebsocketError`.

Order creates, withdrawals, borrow/repay, and transfers are **not retried**.

## Examples

```bash
export BINANCE_API_KEY=
export BINANCE_API_SECRET=
go run ./examples/ticker
```

| Example | |
|---|---|
| [`ticker`](examples/ticker) | public price |
| [`order`](examples/order) / [`test_order`](examples/test_order) | live / validate-only order |
| [`account`](examples/account) | spot account |
| [`historical_klines`](examples/historical_klines) | candle iterator |
| [`websocket_trade`](examples/websocket_trade) / [`websocket_depth`](examples/websocket_depth) | market streams |
| [`user_stream`](examples/user_stream) | authenticated user data |
| [`futures_order`](examples/futures_order) | USD-M on testnet |
| [`margin`](examples/margin) | margin account |
| [`testnet`](examples/testnet) | ping + test order |
| [`websocket_facade`](examples/websocket_facade) | `WebSocket().Trade` |
| [`wsapi`](examples/wsapi) | spot WebSocket API ping + ticker |

Never commit real keys.

## Docs

- [Roadmap & architecture](ROADMAP.md)
- [API coverage](docs/COVERAGE.md)
- [python-binance → binance-go](docs/MAPPING.md)
- [Changelog](CHANGELOG.md)
- [Contributing](CONTRIBUTING.md)
- [Social preview](docs/social-preview.png) — set under GitHub → Settings → General → Social preview

```bash
go test ./...
go test -race ./...
```

Live public endpoints: `BINANCE_INTEGRATION=1 go test -tags=integration ./...`

## License

[MIT](LICENSE) — original code. Not a port of python-binance source.

v1 module: `github.com/Slrrxx/binance-go`. Breaking changes go to `/v2`.
