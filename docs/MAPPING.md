# python-binance → binance-go mapping

Naming follows Go conventions (`GetX`, exported types). This is a learning map, not a 1:1 port.

| python-binance | binance-go |
|---|---|
| `Client(api_key, api_secret)` | `NewClient(apiKey, apiSecret)` |
| `Client(..., testnet=True)` | `NewClient(..., WithTestnet())` |
| `Client(..., demo=True)` | `NewClient(..., WithDemo())` |
| `Client(..., tld='us')` | `NewClient(..., WithTLD("us"))` |
| `ping()` | `Ping(ctx)` |
| `get_server_time()` | `GetServerTime(ctx)` |
| `get_exchange_info()` | `GetExchangeInfo(ctx)` / `Spot().ExchangeInfo` |
| `get_symbol_info` | filter `ExchangeInfo.Symbols` |
| `get_order_book` | `GetOrderBook(ctx, symbol, limit)` |
| `get_recent_trades` | `Spot().RecentTrades` |
| `get_historical_trades` | `Spot().HistoricalTrades` |
| `get_aggregate_trades` | `Spot().AggTrades` |
| `aggregate_trade_iter` | `AggTrades(query)` iterator |
| `get_klines` / `get_ui_klines` | `Spot().Klines` / `UIKlines` |
| `get_historical_klines` | `GetHistoricalKlines` |
| `get_historical_klines_generator` | `HistoricalKlines` / `KlineIterator` |
| `get_avg_price` | `Spot().AvgPrice` |
| `get_ticker` | `Spot().Ticker24h` |
| `get_symbol_ticker` | `GetSymbolTicker` / `Spot().TickerPrice` |
| `get_symbol_ticker_window` | `Spot().RollingTicker` |
| `get_orderbook_ticker` | `Spot().BookTicker` |
| `create_order` | `CreateOrder(ctx, OrderRequest{...})` |
| `create_test_order` | `TestOrder` |
| `order_limit_buy` / `order_market_buy` | compose `OrderRequest` |
| `get_order` | `GetOrder` |
| `get_all_orders` | `Spot().AllOrders` |
| `get_open_orders` | `Spot().OpenOrders` |
| `cancel_order` | `CancelOrder` |
| `cancel_all_open_orders` | `Spot().CancelAllOpenOrders` |
| `cancel_replace_order` | `Spot().CancelReplace` |
| `create_oco_order` | `Spot().CreateOCO` |
| `get_account` | `GetAccount` |
| `get_asset_balance` | `Account.BalanceOf` |
| `get_my_trades` | `Spot().MyTrades` |
| `get_current_order_count` | `Spot().OrderCountUsage` |
| `stream_get_listen_key` | `UserStream().Create` |
| `stream_keepalive` | `UserStream().Keepalive` |
| `stream_close` | `UserStream().Close` |
| `get_system_status` | `Wallet().SystemStatus` |
| `get_deposit_address` | `Wallet().DepositAddress` |
| `get_deposit_history` | `Wallet().DepositHistory` |
| `withdraw` | `Wallet().Withdraw` |
| `get_withdraw_history` | `Wallet().WithdrawHistory` |
| `get_account_snapshot` | `Wallet().AccountSnapshot` |
| `get_asset_details` | `Wallet().AssetDetail` |
| `get_trade_fee` | `Wallet().TradeFee` |
| `make_universal_transfer` | `Wallet().UniversalTransfer` |
| `get_margin_account` | `Margin().Account` |
| `get_isolated_margin_account` | `Margin().IsolatedAccount` |
| `create_margin_order` | `Margin().CreateOrder` |
| `cancel_margin_order` | `Margin().CancelOrder` |
| `create_margin_loan` | `Margin().Borrow` |
| `repay_margin_loan` | `Margin().Repay` |
| `get_margin_interest_history` | `Margin().InterestHistory` |
| `futures_ping` / `futures_time` | `Futures().Ping` / `ServerTime` |
| `futures_exchange_info` | `Futures().ExchangeInfo` |
| `futures_order_book` | `Futures().OrderBook` |
| `futures_klines` | `Futures().Klines` |
| `futures_create_order` | `Futures().CreateOrder` |
| `futures_get_order` | `Futures().GetOrder` |
| `futures_cancel_order` | `Futures().CancelOrder` |
| `futures_account` | `Futures().Account` |
| `futures_account_balance` | `Futures().Balance` |
| `futures_position_information` | `Futures().Positions` |
| `futures_change_leverage` | `Futures().ChangeLeverage` |
| `futures_change_margin_type` | `Futures().ChangeMarginType` |
| `futures_change_position_mode` | `Futures().ChangePositionMode` |
| `futures_income_history` | `Futures().IncomeHistory` |
| `futures_account_trades` | `Futures().UserTrades` |
| `futures_coin_*` | `CoinFutures().*` |
| `BinanceSocketManager.trade_socket` | `WebSocket().Trade(ctx, symbol)` / `Subscribe(ctx, StreamTrade(symbol))` |
| `kline_socket` | `WebSocket().Kline` / `StreamKline` |
| `depth_socket` | `WebSocket().Depth` / `StreamDepth` |
| `DepthCacheManager` | `DepthCache(ctx, symbol)` / `Futures().DepthCache` |
| `BinanceSocketManager.user_socket` | `UserData` / `Futures().UserData` |
| `ws.create_order` (WS API) | `WSAPI().CreateOrder` |
| `papi_*` | `Portfolio().*` |
| `options_*` / eapi | `Options().*` |
| `get_simple_earn_*` | `Earn().*` |
| `convert_*` | `Convert().*` |
| `gift_card_*` | `GiftCard().*` |
| `requests_params` proxy | `WithProxy(url)` |
| `BinanceAPIException` | `*APIError` + `errors.As` |

Go-only additions: `context.Context` on every call, functional options, `RateLimiter`, typed decimals as `string`, structured `Event`, iterator APIs, `go generate` SAPI extras.
