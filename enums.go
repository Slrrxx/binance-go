package binance

// Side is an order side.
type Side string

const (
	// SideBuy is a buy order.
	SideBuy Side = "BUY"
	// SideSell is a sell order.
	SideSell Side = "SELL"
)

// OrderType is a spot or futures order type.
type OrderType string

const (
	OrderTypeLimit              OrderType = "LIMIT"
	OrderTypeMarket             OrderType = "MARKET"
	OrderTypeStopLoss           OrderType = "STOP_LOSS"
	OrderTypeStopLossLimit      OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit         OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit    OrderType = "TAKE_PROFIT_LIMIT"
	OrderTypeLimitMaker         OrderType = "LIMIT_MAKER"
	OrderTypeStop               OrderType = "STOP"
	OrderTypeStopMarket         OrderType = "STOP_MARKET"
	OrderTypeTakeProfitMarket   OrderType = "TAKE_PROFIT_MARKET"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"
)

// TimeInForce controls how long an order remains active.
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC"
	TimeInForceIOC TimeInForce = "IOC"
	TimeInForceFOK TimeInForce = "FOK"
	TimeInForceGTX TimeInForce = "GTX"
	TimeInForceGTD TimeInForce = "GTD"
)

// OrderStatus is a Binance order status.
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
	OrderStatusExpiredInMatch  OrderStatus = "EXPIRED_IN_MATCH"
)

// OrderRespType selects how much detail create-order returns.
type OrderRespType string

const (
	OrderRespACK    OrderRespType = "ACK"
	OrderRespResult OrderRespType = "RESULT"
	OrderRespFull   OrderRespType = "FULL"
)

// PositionSide is a futures position side.
type PositionSide string

const (
	PositionSideBoth  PositionSide = "BOTH"
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

// WorkingType is a futures stop working type.
type WorkingType string

const (
	WorkingTypeMarkPrice     WorkingType = "MARK_PRICE"
	WorkingTypeContractPrice WorkingType = "CONTRACT_PRICE"
)

// MarginType is a futures margin type.
type MarginType string

const (
	MarginTypeIsolated MarginType = "ISOLATED"
	MarginTypeCrossed  MarginType = "CROSSED"
)

// ContractType is a futures contract type.
type ContractType string

const (
	ContractPerpetual      ContractType = "PERPETUAL"
	ContractCurrentQuarter ContractType = "CURRENT_QUARTER"
	ContractNextQuarter    ContractType = "NEXT_QUARTER"
	ContractCurrentMonth   ContractType = "CURRENT_MONTH"
	ContractNextMonth      ContractType = "NEXT_MONTH"
)

// Interval is a kline/candlestick interval.
type Interval string

const (
	Interval1s  Interval = "1s"
	Interval1m  Interval = "1m"
	Interval3m  Interval = "3m"
	Interval5m  Interval = "5m"
	Interval15m Interval = "15m"
	Interval30m Interval = "30m"
	Interval1h  Interval = "1h"
	Interval2h  Interval = "2h"
	Interval4h  Interval = "4h"
	Interval6h  Interval = "6h"
	Interval8h  Interval = "8h"
	Interval12h Interval = "12h"
	Interval1d  Interval = "1d"
	Interval3d  Interval = "3d"
	Interval1w  Interval = "1w"
	Interval1M  Interval = "1M"
)

// IntervalDuration returns the interval length. The 1M interval is treated
// as 30 days for pagination only; callers should prefer an explicit EndTime.
func IntervalDuration(interval Interval) (int64, bool) {
	ms := map[Interval]int64{
		Interval1s:  1000,
		Interval1m:  60_000,
		Interval3m:  3 * 60_000,
		Interval5m:  5 * 60_000,
		Interval15m: 15 * 60_000,
		Interval30m: 30 * 60_000,
		Interval1h:  60 * 60_000,
		Interval2h:  2 * 60 * 60_000,
		Interval4h:  4 * 60 * 60_000,
		Interval6h:  6 * 60 * 60_000,
		Interval8h:  8 * 60 * 60_000,
		Interval12h: 12 * 60 * 60_000,
		Interval1d:  24 * 60 * 60_000,
		Interval3d:  3 * 24 * 60 * 60_000,
		Interval1w:  7 * 24 * 60 * 60_000,
		Interval1M:  30 * 24 * 60 * 60_000,
	}
	v, ok := ms[interval]
	return v, ok
}

// SelfTradePreventionMode is a spot STP mode.
type SelfTradePreventionMode string

const (
	STPExpireMaker     SelfTradePreventionMode = "EXPIRE_MAKER"
	STPExpireTaker     SelfTradePreventionMode = "EXPIRE_TAKER"
	STPExpireBoth      SelfTradePreventionMode = "EXPIRE_BOTH"
	STPNone            SelfTradePreventionMode = "NONE"
	STPDecrementCancel SelfTradePreventionMode = "DECREMENT"
)

// TransferType is a universal (wallet) transfer type.
type TransferType string

const (
	TransferMainToUMFuture   TransferType = "MAIN_UMFUTURE"
	TransferMainToCMFuture   TransferType = "MAIN_CMFUTURE"
	TransferMainToMargin     TransferType = "MAIN_MARGIN"
	TransferUMFutureToMain   TransferType = "UMFUTURE_MAIN"
	TransferCMFutureToMain   TransferType = "CMFUTURE_MAIN"
	TransferMarginToMain     TransferType = "MARGIN_MAIN"
	TransferMarginToUMFuture TransferType = "MARGIN_UMFUTURE"
	TransferUMFutureToMargin TransferType = "UMFUTURE_MARGIN"
	TransferMainToFunding    TransferType = "MAIN_FUNDING"
	TransferFundingToMain    TransferType = "FUNDING_MAIN"
)

// Environment selects production, testnet, or demo hosts.
type Environment int

const (
	// EnvProduction is the live Binance environment.
	EnvProduction Environment = iota
	// EnvTestnet is the public testnet.
	EnvTestnet
	// EnvDemo is Binance demo trading.
	EnvDemo
)

// Market is a product family used by streams and historical helpers.
type Market int

const (
	MarketSpot Market = iota
	MarketUSDFutures
	MarketCoinFutures
	MarketPortfolio
	MarketOptions
)

// DepthLevel is a partial book depth for WebSocket streams.
type DepthLevel string

const (
	DepthLevel5  DepthLevel = "5"
	DepthLevel10 DepthLevel = "10"
	DepthLevel20 DepthLevel = "20"
)
