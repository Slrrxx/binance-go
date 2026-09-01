# binance-go Roadmap

**Status (2026-09-02):** v0.1 implemented. `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...` pass. Examples compile. Live exchange tests remain opt-in (`BINANCE_INTEGRATION=1`).


Production-ready unofficial Binance SDK for Go, inspired by [python-binance](https://github.com/sammchardy/python-binance) in **UX and feature surface**, implemented as original idiomatic Go.

Module path (replace after you publish):

```text
github.com/Slrrxx/binance-go
```

```go
import "github.com/Slrrxx/binance-go"
```

python-binance is MIT-licensed. This project does **not** copy its source. We reuse public Binance API contracts and the *idea* of a batteries-included client (auto timestamp, testnet, historical klines, depth cache, reconnecting websockets).

---

## 1. python-binance feature / API coverage matrix

Source: python-binance v1.0.37 (`Client`, `AsyncClient`, `BaseClient`, `binance/ws/*`). The Python client is a single god-class plus a large generated tail (`margin_v1_*`, `futures_v1_*`, `options_v1_*`, `papi_*`).

| Area | python-binance | binance-go v1 | Notes |
|---|---|---|---|
| Spot general (ping, time, exchangeInfo) | Yes | **Yes** | Convenience methods on `Client` |
| Spot market data | Yes | **Yes** | depth, trades, aggTrades, klines, tickers, avgPrice, rolling window |
| Spot trading | Yes | **Yes** | create / test / cancel / query / open / all / OCO / cancel-replace |
| Spot account | Yes | **Yes** | account, balances helper, myTrades, order count |
| User data stream (listenKey) | Yes | **Yes** | create / keepalive / close + WS + auto-keepalive |
| Historical klines + generator | Yes | **Yes** | slice helper + `KlineIterator` |
| Agg trade iterator | Yes | **Yes** | `AggTradeIterator` |
| Wallet / capital | Yes | **Yes** | status, coins, deposit, **withdraw**, snapshot, asset detail |
| Margin (cross + isolated) | Yes | **Yes** | account, order, borrow/repay, transfer, interest |
| USD-M futures (fapi) | Yes | **Yes** | market, trade, account, positions, leverage, income |
| COIN-M futures (dapi) | Yes | **Yes** | parallel service, shared types where possible |
| Futures data (open interest hist, L/S ratios) | Yes | Partial | Core market + account first; extras marked in `docs/COVERAGE.md` |
| WebSocket market streams | Yes | **Yes** | trade, aggTrade, ticker, miniTicker, bookTicker, depth, kline, combined |
| WebSocket reconnect / backoff | Yes (5 retries, exp backoff) | **Yes** | configurable, context-aware, resubscribe |
| Depth cache | Yes | **Yes** | REST snapshot + diff sync + gap resync |
| HMAC SHA256 | Yes | **Yes** | query-string signing |
| RSA / Ed25519 | Yes | **Yes** | PEM private key options |
| Auto timestamp + recvWindow | Yes | **Yes** | optional server-time offset sync |
| Testnet | Yes | **Yes** | `WithTestnet()` |
| Demo trading | Yes | **Yes** | `WithDemo()` |
| Custom HTTP / proxy / headers | Yes | **Yes** | functional options |
| Rate-limit awareness | Partial | **Yes** | 429/418, Retry-After, optional limiter |
| Retry (non-idempotent safe) | Informal | **Yes** | never blind-retry order creates |
| Threaded WS managers | Yes | N/A | Go uses goroutines + `context` instead |
| Async twin client | `AsyncClient` | N/A | One client; all calls take `context.Context` |
| WebSocket API trading (ws-api) | Yes | Deferred | REST is v1 source of truth; WS-API later |
| Portfolio margin (papi) | Yes | Deferred | See coverage checklist |
| Vanilla options (eapi) | Yes | Deferred | |
| Simple Earn / staking / mining / convert / gift card | Yes | Deferred | |
| Sub-account (core) | Yes | **Yes** | list, assets, transfers |
| TLD variants (.us / .jp) | Yes | **Yes** | `WithTLD("us")` |
| Orjson / extra JSON libs | Yes | No | `encoding/json` only |
| Typed request/response models | Mostly `dict` | **Yes** | structs + string decimals |
| Structured errors | Exception classes | **Yes** | `errors.As` / `APIError` |

### python-binance methods we treat as the UX north star

| python-binance | binance-go |
|---|---|
| `Client(api_key, api_secret, testnet=True)` | `NewClient(key, secret, WithTestnet())` |
| `get_account()` | `GetAccount(ctx)` / `Spot().Account(ctx)` |
| `get_symbol_ticker(symbol=...)` | `GetSymbolTicker(ctx, "BTCUSDT")` |
| `create_order(...)` | `CreateOrder(ctx, OrderRequest{...})` |
| `create_test_order(...)` | `TestOrder(ctx, OrderRequest{...})` |
| `get_historical_klines(...)` | `GetHistoricalKlines(ctx, ...)` |
| `get_historical_klines_generator(...)` | `HistoricalKlines(ctx, ...)` iterator |
| `stream_get_listen_key()` | `UserStream().Create(ctx)` |
| `BinanceSocketManager.trade_socket` | `Subscribe(ctx, StreamTrade("BTCUSDT"))` |
| `DepthCacheManager` | `DepthCache(ctx, "BTCUSDT")` |
| `futures_create_order` | `Futures().CreateOrder(ctx, ...)` |
| `futures_coin_*` | `CoinFutures().*` |
| `withdraw` | `Wallet().Withdraw(ctx, ...)` (documented as dangerous) |

Full mapping: [`docs/MAPPING.md`](docs/MAPPING.md). Live checklist: [`docs/COVERAGE.md`](docs/COVERAGE.md).

---

## 2. Recommended Go repository architecture

Single module, **one public package** (`binance`) so the import path matches the python-binance mental model: `import ".../binance-go"` then `binance.NewClient`.

Domain files keep the package from becoming an unreadable blob. HTTP, signing, and retry stay unexported.

```text
binance-go-api/
â”œâ”€â”€ doc.go                 # package overview
â”œâ”€â”€ client.go              # Client, services, convenience methods
â”œâ”€â”€ options.go             # functional options
â”œâ”€â”€ endpoints.go           # centralized URLs / versions
â”œâ”€â”€ enums.go               # Side, OrderType, Interval, ...
â”œâ”€â”€ types.go               # shared request/response models
â”œâ”€â”€ errors.go              # APIError and friends
â”œâ”€â”€ logger.go              # Logger interface, nop default
â”œâ”€â”€ redact.go              # secret redaction
â”œâ”€â”€ signer.go              # HMAC / RSA / Ed25519
â”œâ”€â”€ params.go              # query encoding
â”œâ”€â”€ transport.go           # net/http request cycle
â”œâ”€â”€ retry.go               # idempotent retry policy
â”œâ”€â”€ ratelimit.go           # RateLimiter + header parsing
â”œâ”€â”€ time.go                # server-time offset
â”œâ”€â”€ clientid.go            # NewClientOrderID
â”œâ”€â”€ spot.go                # SpotService
â”œâ”€â”€ spot_market.go
â”œâ”€â”€ spot_trade.go
â”œâ”€â”€ spot_account.go
â”œâ”€â”€ userstream.go
â”œâ”€â”€ wallet.go
â”œâ”€â”€ margin.go
â”œâ”€â”€ subaccount.go
â”œâ”€â”€ futures.go             # USD-M
â”œâ”€â”€ coinfutures.go         # COIN-M
â”œâ”€â”€ klines.go              # historical pagination + iterator
â”œâ”€â”€ websocket.go           # streams + reconnect
â”œâ”€â”€ websocket_events.go    # typed events
â”œâ”€â”€ depthcache.go
â”œâ”€â”€ *_test.go              # httptest unit tests (no live Binance)
â”œâ”€â”€ integration_test.go    # //go:build integration
â”œâ”€â”€ examples/
â”œâ”€â”€ docs/
â”œâ”€â”€ .github/workflows/ci.yml
â”œâ”€â”€ .golangci.yml
â”œâ”€â”€ go.mod
â”œâ”€â”€ LICENSE                # MIT
â”œâ”€â”€ CHANGELOG.md
â”œâ”€â”€ README.md
â””â”€â”€ ROADMAP.md
```

Why not `spot/`, `futures/` subpackages? They would force users to import several packages for types that belong together (`OrderRequest`, `Side`, `APIError`). Stripe-style and `google.golang.org` clients that want a Python-like DX keep one package and split files.

Why not generate every Binance path in v1? python-binance's generated tail is huge and drifts. v1 is handwritten for the trading core. A generator can be added later under `internal/generator` without breaking the public API.

**Dependency policy:** standard library + `github.com/coder/websocket` (context-native, maintained). No gorilla/websocket. No decimal library: Binance decimals stay `string`.

---

## 3. Public API design

### Construction

```go
client := binance.NewClient(apiKey, apiSecret,
    binance.WithTestnet(),
    binance.WithTimeout(10*time.Second),
    binance.WithHTTPClient(httpClient),
    binance.WithRecvWindow(5000),
    binance.WithAutoTimeSync(),
    binance.WithRateLimiter(binance.NewWeightLimiter(1200, time.Minute)),
    binance.WithLogger(myLogger),
)
```

`Client` is safe for concurrent use. Secrets are never printed by `fmt` or debug logs.

### Two layers (convenience + services)

God-class avoidance: full surface lives on small services. The methods python users reach for first are also on `Client`.

```go
client.GetAccount(ctx)
client.GetSymbolTicker(ctx, "BTCUSDT")
client.CreateOrder(ctx, binance.OrderRequest{...})

client.Spot()
client.Futures()      // USDâ“ˆ-M
client.CoinFutures()  // COIN-M
client.Margin()
client.Wallet()
client.SubAccount()
client.UserStream()
```

### Orders (type-safe, string decimals)

```go
order, err := client.CreateOrder(ctx, binance.OrderRequest{
    Symbol:      "BTCUSDT",
    Side:        binance.SideBuy,
    Type:        binance.OrderTypeLimit,
    TimeInForce: binance.TimeInForceGTC,
    Quantity:    "0.001",
    Price:       "50000",
    ClientOrderID: binance.NewClientOrderID(),
})
```

### Errors

```go
var apiErr *binance.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.Code, apiErr.Message, apiErr.HTTPStatus)
}
```

Distinct types: `APIError`, `DecodeError`, `AuthError`, `RateLimitError`, `WebsocketError`, plus wrapped `context` / network errors.

### WebSocket

```go
stream, err := client.Subscribe(ctx, binance.StreamTrade("BTCUSDT"))
defer stream.Close()

for {
    ev, err := stream.Next(ctx)
    if err != nil { return err }
    if ev.Trade != nil {
        fmt.Println(ev.Trade.Price, ev.Trade.Quantity)
    }
}
```

Reconnect, backoff, and subscription restore are internal.

### Depth cache

```go
cache, err := client.DepthCache(ctx, "BTCUSDT")
defer cache.Close()

for {
    if _, err := cache.Wait(ctx); err != nil { return err }
    bids, asks := cache.Bids(), cache.Asks()
    _ = bids
    _ = asks
}
```

### Historical klines

```go
klines, err := client.GetHistoricalKlines(ctx, "BTCUSDT", binance.Interval1m, start, end)

it := client.HistoricalKlines(ctx, binance.KlineQuery{
    Symbol: "BTCUSDT", Interval: binance.Interval1m, StartTime: start, EndTime: end,
})
for it.Next(ctx) {
    k := it.Kline()
    _ = k
}
if err := it.Err(); err != nil { return err }
```

---

## 4. Implementation roadmap

| Phase | Scope | Exit criteria |
|---|---|---|
| 1 Research | python-binance + Binance REST/WS contracts | This document |
| 2 Architecture | Package layout, options, services | This document + `docs/` |
| 3 Core | Client, transport, signer, errors, options, httptest | `go test` for auth/request/errors |
| 4 Spot | Market + trading + account + listenKey | Typed methods + tests |
| 5 WebSocket | Streams, typed events, reconnect | httptest WS + reconnect test |
| 6 Historical data | Kline pagination + iterator | Pagination tests, no giant RAM requirement |
| 7 Futures | USD-M + COIN-M | Shared helpers, no copy-paste explosion |
| 8 Margin / Wallet / Sub | Requested endpoints | Withdrawals documented as dangerous |
| 9 Depth cache | Snapshot + diffs + gap detect | Sequence tests |
| 10 Tests | Unit (always) + integration (opt-in) | `go test ./...` and `-race` |
| 11 Docs | README, examples, mapping, coverage | Examples compile |
| 12 CI | gofmt, vet, test, race, golangci-lint, build | GitHub Actions |

v1 **does not** implement: Options, Portfolio Margin, Simple Earn, Gift Card, WS-API trading, full generated SAPI long tail.

After each phase: compile, test, race, review public API, update coverage/README.

---

## 5. Risks and technical decisions

| Decision | Choice | Why / risk |
|---|---|---|
| One package `binance` | Yes | Best DX; file-split for maintainability. Risk: large package â€” mitigated by services + files. |
| Decimals as `string` | Yes | Avoid float64; no extra dependency. Risk: callers must parse if they need math. |
| `coder/websocket` | Only third-party dep | Context-first, maintained. Risk: API less familiar than gorilla. |
| Query-string signing for all HTTP methods | Yes | Matches current Binance + other Go clients; one code path. |
| Do not retry `POST /order` | Hard rule | Duplicate orders are worse than a failed call. Use `ClientOrderID`. |
| Auto timestamp | Always | Optional `WithAutoTimeSync` for clock skew (`-1021`). |
| recvWindow default 5000ms | Configurable | Align with Binance docs; python-binance uses 10000. |
| Rate limiter optional | Interface + default weight limiter | Do not surprise HFT users with hidden sleeps. Always parse 429/418. |
| Depth cache gap = resync | Yes | Never apply a broken book. |
| No live unit tests | httptest only | CI must be deterministic. Integration is `//go:build integration`. |
| Secrets | Redact in logs/errors | Never log `signature`, secret, PEM, or raw signed query. |
| Module path | `github.com/Slrrxx/binance-go` | v2 would be `/v2`. |
| License | MIT | Same class as python-binance; commercial + OSS friendly. **Original code.** |
| Code generation | Not in v1 | Prefer reviewable handwritten core. Generator is a later additive tool. |
| WS-API / papi / options | Deferred | Scope control; REST + market WS covers the python "few minutes to first trade" path. |

### Security notes (withdrawals)

`Wallet().Withdraw` is implemented because python-binance has it and production bots need it. README and GoDoc state:

- restrict API keys (IP + permission)
- never commit keys
- prefer a whitelist address
- treat withdraw tests as production-dangerous even on "small" amounts

### Compatibility

Go 1.24+ (developed on Go 1.26). Breaking API changes go to `/v2`.
