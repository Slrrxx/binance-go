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

## Deferred (not v0.1)

- [ ] Portfolio margin (papi)
- [ ] Vanilla options (eapi)
- [ ] WebSocket API trading (ws-api / ws-fapi)
- [ ] Simple Earn / staking / mining / convert / gift card
- [ ] Full generated SAPI long tail
- [ ] Broker / copy-trading
- [ ] Code generator (`internal/generator`)
