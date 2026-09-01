package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// StreamSpec names a Binance market-data or user-data stream.
type StreamSpec struct {
	Name      string
	Market    Market
	userData  bool
	listenKey string
}

// StreamTrade is `{symbol}@trade`.
func StreamTrade(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@trade"}
}

// StreamAggTrade is `{symbol}@aggTrade`.
func StreamAggTrade(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@aggTrade"}
}

// StreamTicker is `{symbol}@ticker`.
func StreamTicker(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@ticker"}
}

// StreamMiniTicker is `{symbol}@miniTicker`.
func StreamMiniTicker(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@miniTicker"}
}

// StreamBookTicker is `{symbol}@bookTicker`.
func StreamBookTicker(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@bookTicker"}
}

// StreamDepth is `{symbol}@depth` (diff book).
func StreamDepth(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@depth"}
}

// StreamDepth100ms is `{symbol}@depth@100ms`.
func StreamDepth100ms(symbol string) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@depth@100ms"}
}

// StreamPartialDepth is `{symbol}@depth{level}` or `@depth{level}@100ms`.
func StreamPartialDepth(symbol string, level DepthLevel, ms100 bool) StreamSpec {
	n := strings.ToLower(symbol) + "@depth" + string(level)
	if ms100 {
		n += "@100ms"
	}
	return StreamSpec{Name: n}
}

// StreamKline is `{symbol}@kline_{interval}`.
func StreamKline(symbol string, interval Interval) StreamSpec {
	return StreamSpec{Name: strings.ToLower(symbol) + "@kline_" + string(interval)}
}

// StreamFuturesTrade tags a stream as USD-M futures.
func StreamFuturesTrade(symbol string) StreamSpec {
	s := StreamTrade(symbol)
	s.Market = MarketUSDFutures
	return s
}

// StreamCoinFuturesTrade tags a stream as COIN-M futures.
func StreamCoinFuturesTrade(symbol string) StreamSpec {
	s := StreamTrade(symbol)
	s.Market = MarketCoinFutures
	return s
}

// Stream is a reconnecting combined WebSocket subscription.
type Stream struct {
	c         *Client
	specs     []StreamSpec
	market    Market
	events    chan Event
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
	lastMsg   time.Time
}

// Subscribe opens a reconnecting stream for one or more specs on the same market.
func (c *Client) Subscribe(ctx context.Context, specs ...StreamSpec) (*Stream, error) {
	if len(specs) == 0 {
		return nil, &WebsocketError{Op: "subscribe", Err: fmt.Errorf("no streams")}
	}
	market := specs[0].Market
	for _, s := range specs[1:] {
		if s.Market != market {
			return nil, &WebsocketError{Op: "subscribe", Err: fmt.Errorf("mixed markets in one stream")}
		}
	}
	sctx, cancel := context.WithCancel(ctx)
	st := &Stream{
		c:      c,
		specs:  specs,
		market: market,
		events: make(chan Event, c.cfg.wsBuffer),
		ctx:    sctx,
		cancel: cancel,
	}
	go st.loop()
	return st, nil
}

// Next waits for the next event or ctx cancellation.
func (s *Stream) Next(ctx context.Context) (Event, error) {
	select {
	case <-ctx.Done():
		return Event{}, ctx.Err()
	case <-s.ctx.Done():
		select {
		case ev, ok := <-s.events:
			if ok {
				return ev, nil
			}
		default:
		}
		return Event{}, ErrStreamClosed
	case ev, ok := <-s.events:
		if !ok {
			return Event{}, ErrStreamClosed
		}
		return ev, nil
	}
}

// Events returns the receive-only event channel. Do not mix with Next.
func (s *Stream) Events() <-chan Event { return s.events }

// Close stops reconnects and closes the event channel.
func (s *Stream) Close() error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
		s.cancel()
	})
	return nil
}

// Healthy reports whether a message was received recently (2 minutes).
func (s *Stream) Healthy() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastMsg.IsZero() {
		return false
	}
	return time.Since(s.lastMsg) < 2*time.Minute
}

func (s *Stream) loop() {
	defer close(s.events)
	attempt := 0
	for {
		if s.ctx.Err() != nil {
			return
		}
		err := s.connectOnce()
		if s.ctx.Err() != nil {
			return
		}
		attempt++
		max := s.c.cfg.wsMaxRetries
		if max > 0 && attempt > max {
			s.emitError(&WebsocketError{Op: "reconnect", Err: fmt.Errorf("exceeded max retries: %w", err)})
			return
		}
		delay := s.c.cfg.wsBackoffMin
		for i := 1; i < attempt; i++ {
			delay *= 2
			if delay > s.c.cfg.wsBackoffMax {
				delay = s.c.cfg.wsBackoffMax
				break
			}
		}
		s.c.cfg.log.Info("websocket reconnecting", "attempt", attempt, "delay", delay.String())
		if err := sleep(s.ctx, delay); err != nil {
			return
		}
	}
}

func (s *Stream) connectOnce() error {
	url := s.dialURL()
	dialCtx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	httpClient := s.c.cfg.httpClient
	if httpClient != nil && httpClient.Timeout > 0 {
		cloned := *httpClient
		cloned.Timeout = 0
		httpClient = &cloned
	}
	conn, _, err := websocket.Dial(dialCtx, url, &websocket.DialOptions{
		HTTPClient: httpClient,
		HTTPHeader: http.Header{"User-Agent": []string{s.c.cfg.userAgent}},
	})
	cancel()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()
	conn.SetReadLimit(1 << 20)
	for {
		_, data, err := conn.Read(s.ctx)
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.lastMsg = time.Now()
		s.mu.Unlock()
		ev, err := parseEvent(data)
		if err != nil {
			s.c.cfg.log.Error("websocket decode", "err", err)
			continue
		}
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case s.events <- ev:
		}
	}
}

func (s *Stream) dialURL() string {
	if len(s.specs) == 1 && (s.specs[0].userData || !strings.Contains(s.specs[0].Name, "@") && s.specs[0].listenKey != "") {
		return strings.TrimRight(s.c.endpoints.ws(s.market, false), "/") + "/" + s.specs[0].Name
	}
	if len(s.specs) == 1 {
		return strings.TrimRight(s.c.endpoints.ws(s.market, false), "/") + "/" + s.specs[0].Name
	}
	names := make([]string, len(s.specs))
	for i, sp := range s.specs {
		names[i] = sp.Name
	}
	return strings.TrimRight(s.c.endpoints.ws(s.market, true), "/") + "?streams=" + strings.Join(names, "/")
}

func (s *Stream) emitError(err error) {
	ev := Event{Type: "error", Err: err}
	select {
	case s.events <- ev:
	default:
	}
}

func parseEvent(data []byte) (Event, error) {
	var wrap struct {
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}
	payload := data
	if err := json.Unmarshal(data, &wrap); err == nil && len(wrap.Data) > 0 && wrap.Stream != "" {
		payload = wrap.Data
	}
	var head struct {
		EventType string `json:"e"`
		EventTime TimeMS `json:"E"`
	}
	_ = json.Unmarshal(payload, &head)
	ev := Event{
		Stream:    wrap.Stream,
		Type:      head.EventType,
		EventTime: head.EventTime.Time(),
		Raw:       append(json.RawMessage(nil), payload...),
	}
	switch head.EventType {
	case "trade":
		var t TradeEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.Trade = &t
	case "aggTrade":
		var t AggTradeEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.AggTrade = &t
	case "24hrTicker":
		var t TickerEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.Ticker = &t
	case "24hrMiniTicker":
		var t MiniTickerEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.MiniTicker = &t
	case "bookTicker":
		var t BookTickerEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.BookTicker = &t
	case "depthUpdate":
		var t DepthEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.Depth = &t
	case "kline":
		var t KlineEvent
		if err := json.Unmarshal(payload, &t); err != nil {
			return ev, err
		}
		ev.Kline = &t
	case "outboundAccountPosition", "balanceUpdate", "executionReport",
		"listStatus", "listenKeyExpired", "ACCOUNT_UPDATE", "ORDER_TRADE_UPDATE",
		"MARGIN_CALL", "ACCOUNT_CONFIG_UPDATE":
		ev.User = payload
	default:
		if head.EventType == "" {
			var t BookTickerEvent
			if err := json.Unmarshal(payload, &t); err == nil && t.Symbol != "" && t.BidPrice != "" {
				ev.Type = "bookTicker"
				ev.BookTicker = &t
			}
		}
	}
	return ev, nil
}
