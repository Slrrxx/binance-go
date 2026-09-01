# Slrrxx

I write production-shaped Go libraries. Right now that means a batteries-included Binance SDK: one client, string decimals, reconnecting streams, and no surprise retries on order creates.

<p align="left">
  <a href="https://github.com/Slrrxx/binance-go">
    <img src="https://img.shields.io/github/stars/Slrrxx/binance-go?style=flat-square" alt="stars">
  </a>
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/focus-trading%20APIs-111?style=flat-square" alt="focus">
</p>

## Featured

**[binance-go](https://github.com/Slrrxx/binance-go)** — unofficial Binance REST + WebSocket SDK

```go
client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_API_SECRET"))
ticker, _ := client.GetSymbolTicker(ctx, "BTCUSDT")
stream, _ := client.WebSocket().Trade(ctx, "BTCUSDT")
```

- Spot, USD-M / COIN-M futures, margin, portfolio, options
- WebSocket API trading and a depth cache
- HMAC, RSA, Ed25519 — secrets never logged

[Docs](https://pkg.go.dev/github.com/Slrrxx/binance-go) · [Coverage](https://github.com/Slrrxx/binance-go/blob/main/docs/COVERAGE.md)

## Stack

Go · HTTP/WebSocket clients · exchange APIs · CI

---

Pin **binance-go** on this profile (GitHub → Customize your pins).
