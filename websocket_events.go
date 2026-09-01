package binance

import (
	"encoding/json"
	"time"
)

// Event is a typed WebSocket payload. Exactly one of the pointer fields is
// typically set; Raw always holds the decoded JSON object.
type Event struct {
	Stream     string
	Type       string
	EventTime  time.Time
	Trade      *TradeEvent
	AggTrade   *AggTradeEvent
	Ticker     *TickerEvent
	MiniTicker *MiniTickerEvent
	BookTicker *BookTickerEvent
	Depth      *DepthEvent
	Kline      *KlineEvent
	User       json.RawMessage
	Raw        json.RawMessage
	Err        error
}

// TradeEvent is a raw trade stream payload.
type TradeEvent struct {
	EventType     string `json:"e"`
	EventTime     TimeMS `json:"E"`
	Symbol        string `json:"s"`
	TradeID       int64  `json:"t"`
	Price         string `json:"p"`
	Quantity      string `json:"q"`
	BuyerOrderID  int64  `json:"b"`
	SellerOrderID int64  `json:"a"`
	TradeTime     TimeMS `json:"T"`
	IsBuyerMaker  bool   `json:"m"`
}

// AggTradeEvent is an aggregate trade stream payload.
type AggTradeEvent struct {
	EventType    string `json:"e"`
	EventTime    TimeMS `json:"E"`
	Symbol       string `json:"s"`
	AggTradeID   int64  `json:"a"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"l"`
	TradeTime    TimeMS `json:"T"`
	IsBuyerMaker bool   `json:"m"`
}

// TickerEvent is a 24hr ticker stream payload.
type TickerEvent struct {
	EventType          string `json:"e"`
	EventTime          TimeMS `json:"E"`
	Symbol             string `json:"s"`
	PriceChange        string `json:"p"`
	PriceChangePercent string `json:"P"`
	WeightedAvgPrice   string `json:"w"`
	FirstTradePrice    string `json:"x"`
	LastPrice          string `json:"c"`
	LastQty            string `json:"Q"`
	BestBidPrice       string `json:"b"`
	BestBidQty         string `json:"B"`
	BestAskPrice       string `json:"a"`
	BestAskQty         string `json:"A"`
	OpenPrice          string `json:"o"`
	HighPrice          string `json:"h"`
	LowPrice           string `json:"l"`
	Volume             string `json:"v"`
	QuoteVolume        string `json:"q"`
	OpenTime           TimeMS `json:"O"`
	CloseTime          TimeMS `json:"C"`
	FirstTradeID       int64  `json:"F"`
	LastTradeID        int64  `json:"L"`
	Count              int64  `json:"n"`
}

// MiniTickerEvent is a mini-ticker stream payload.
type MiniTickerEvent struct {
	EventType   string `json:"e"`
	EventTime   TimeMS `json:"E"`
	Symbol      string `json:"s"`
	ClosePrice  string `json:"c"`
	OpenPrice   string `json:"o"`
	HighPrice   string `json:"h"`
	LowPrice    string `json:"l"`
	Volume      string `json:"v"`
	QuoteVolume string `json:"q"`
}

// BookTickerEvent is a best bid/ask stream payload.
type BookTickerEvent struct {
	UpdateID int64  `json:"u"`
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

// DepthEvent is a diff depth stream payload.
type DepthEvent struct {
	EventType     string       `json:"e"`
	EventTime     TimeMS       `json:"E"`
	Symbol        string       `json:"s"`
	FirstUpdateID int64        `json:"U"`
	FinalUpdateID int64        `json:"u"`
	PrevUpdateID  int64        `json:"pu"`
	Bids          []PriceLevel `json:"b"`
	Asks          []PriceLevel `json:"a"`
}

// KlineEvent is a kline stream payload.
type KlineEvent struct {
	EventType string     `json:"e"`
	EventTime TimeMS     `json:"E"`
	Symbol    string     `json:"s"`
	Kline     KlineFrame `json:"k"`
}

// KlineFrame is the nested kline object.
type KlineFrame struct {
	StartTime     TimeMS   `json:"t"`
	CloseTime     TimeMS   `json:"T"`
	Symbol        string   `json:"s"`
	Interval      Interval `json:"i"`
	FirstTradeID  int64    `json:"f"`
	LastTradeID   int64    `json:"L"`
	Open          string   `json:"o"`
	Close         string   `json:"c"`
	High          string   `json:"h"`
	Low           string   `json:"l"`
	Volume        string   `json:"v"`
	TradeCount    int64    `json:"n"`
	Closed        bool     `json:"x"`
	QuoteVolume   string   `json:"q"`
	TakerBuyBase  string   `json:"V"`
	TakerBuyQuote string   `json:"Q"`
}
