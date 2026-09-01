# Coverage checklist

Status key: **done** in v0.1.0, **partial**, **deferred**.

## Spot

- [x] Ping / server time / exchangeInfo
- [x] Order book
- [x] Recent trades
- [x] Historical trades (API key)
- [x] Aggregate trades + iterator
- [x] Klines / UI klines
- [x] Average price
- [x] Ticker price / book ticker / 24h / rolling window
- [x] Create / test / cancel / cancel-all / query / open / all orders
- [x] Cancel-replace
- [x] OCO create / cancel / open lists
- [x] Account / balances / my trades / order count
- [x] User data stream listenKey + keepalive + close + WS
- [x] Historical klines (slice + iterator)

## Wallet

- [x] System status
- [x] Coin configuration
- [x] Deposit address / history
- [x] Withdraw / withdraw history (dangerous; documented)
- [x] Account snapshot
- [x] Asset detail / trade fee
- [x] Universal transfer
- [x] Dust transfer (minimal)

## Margin

- [x] Cross account
- [x] Isolated account / enable / disable
- [x] Margin order create / cancel / query / open / all
- [x] Borrow / repay
- [x] Cross + isolated transfer
- [x] Interest history
- [x] Max borrowable
- [x] Margin listenKey

## USD-M futures

- [x] Exchange info / book / klines / tickers / mark price / funding
- [x] Create / test / query / cancel / cancel-all / open / all
- [x] Account / balance / positions
- [x] Leverage / margin type / position mode
- [x] Income / user trades
- [x] Listen key

## COIN-M futures

- [x] Parallel market + trading + account surface (dapi)

## WebSocket

- [x] trade, aggTrade, ticker, miniTicker, bookTicker, depth, kline
- [x] Combined streams
- [x] Futures stream helpers
- [x] Typed events
- [x] Reconnect + backoff + cancel
- [x] User data stream

## Depth cache

- [x] REST snapshot + incremental updates
- [x] Sequence / pu gap detection + resync

## Portfolio margin

- [x] Ping / account / balances
- [x] UM + CM position risk
- [x] UM + CM order create / cancel
- [x] UM leverage
- [x] Listen key + user-data stream

## Vanilla options

- [x] Ping / time / exchangeInfo / book / klines / ticker / mark
- [x] Account / positions / orders / trades
- [x] Listen key + user-data stream

## WebSocket API

- [x] Spot ping / time / ticker / account
- [x] Spot order place / test / cancel / status
- [x] USD-M `FuturesWSAPI()` order place

## Earn / convert / gift card

- [x] Simple Earn flexible + locked list, subscribe, redeem, account
- [x] Convert exchangeInfo, quote, accept, order status
- [x] Gift card create / redeem / verify / token limit

## Generated SAPI extras

- [x] `internal/generator` + `zz_generated.go` (`go generate ./...`)
- [x] Wallet dust log, API restrictions, delist schedule, trading status, dividends
- [x] Earn flexible / locked position

## WebSocket facade + proxy

- [x] `client.WebSocket().Trade` and related helpers
- [x] Futures + COIN-M user-data and depth cache
- [x] `WithProxy`

## Deferred

- [ ] Mining
- [ ] Broker / copy-trading
- [ ] Full Binance SAPI catalog (add rows to `internal/generator/endpoints.json`)
