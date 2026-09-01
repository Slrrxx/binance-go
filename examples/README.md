# Examples

All examples use the module `github.com/Slrrxx/binance-go`.

```bash
export BINANCE_API_KEY=
export BINANCE_API_SECRET=
go run ./examples/ticker
```

Never commit real keys. Public market examples work with empty credentials.

| Example | |
|---|---|
| [`ticker`](ticker) | public price |
| [`order`](order) / [`test_order`](test_order) | live / validate-only order |
| [`account`](account) | spot account |
| [`historical_klines`](historical_klines) | candle iterator |
| [`websocket_trade`](websocket_trade) / [`websocket_depth`](websocket_depth) | market streams |
| [`websocket_facade`](websocket_facade) | `WebSocket().Trade` |
| [`wsapi`](wsapi) | WebSocket API |
| [`user_stream`](user_stream) | authenticated user data |
| [`futures_order`](futures_order) | USD-M on testnet |
| [`margin`](margin) | margin account |
| [`testnet`](testnet) | ping + test order |
