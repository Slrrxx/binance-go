package binance

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"
)

// DepthCache is a thread-safe local order book synchronized from a REST
// snapshot plus incremental depth WebSocket events.
type DepthCache struct {
	c          *Client
	symbol     string
	market     Market
	stream     *Stream
	mu         sync.RWMutex
	bids       map[string]string
	asks       map[string]string
	lastID     int64
	updateTime time.Time
	updates    chan struct{}
	err        error
	cancel     context.CancelFunc
}

// DepthCacheOption configures DepthCache.
type DepthCacheOption func(*DepthCache)

// WithDepthMarket selects spot (default), USD-M, or COIN-M books.
func WithDepthMarket(m Market) DepthCacheOption {
	return func(d *DepthCache) { d.market = m }
}

// DepthCache starts a synchronized local order book for symbol.
func (c *Client) DepthCache(ctx context.Context, symbol string, opts ...DepthCacheOption) (*DepthCache, error) {
	dctx, cancel := context.WithCancel(ctx)
	d := &DepthCache{
		c:       c,
		symbol:  symbol,
		bids:    make(map[string]string),
		asks:    make(map[string]string),
		updates: make(chan struct{}, 1),
		cancel:  cancel,
	}
	for _, o := range opts {
		o(d)
	}
	spec := StreamDepth(symbol)
	spec.Market = d.market
	st, err := c.Subscribe(dctx, spec)
	if err != nil {
		cancel()
		return nil, err
	}
	d.stream = st
	go d.run(dctx)
	return d, nil
}

// Close stops the cache and underlying stream.
func (d *DepthCache) Close() error {
	d.cancel()
	if d.stream != nil {
		return d.stream.Close()
	}
	return nil
}

// Wait blocks until the book is updated or ctx is done.
func (d *DepthCache) Wait(ctx context.Context) (struct{}, error) {
	select {
	case <-ctx.Done():
		return struct{}{}, ctx.Err()
	case <-d.updates:
		d.mu.RLock()
		err := d.err
		d.mu.RUnlock()
		return struct{}{}, err
	}
}

// Updates is signaled (non-blocking, coalesced) after each applied book update.
func (d *DepthCache) Updates() <-chan struct{} { return d.updates }

// Bids returns bids sorted by price descending.
func (d *DepthCache) Bids() []PriceLevel {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return sortedLevels(d.bids, true)
}

// Asks returns asks sorted by price ascending.
func (d *DepthCache) Asks() []PriceLevel {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return sortedLevels(d.asks, false)
}

// LastUpdateID is the last applied update id.
func (d *DepthCache) LastUpdateID() int64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastID
}

func (d *DepthCache) run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		err := d.syncOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		d.setErr(err)
		d.c.cfg.log.Info("depth cache resync", "symbol", d.symbol, "err", err)
		if err := sleep(ctx, time.Second); err != nil {
			return
		}
	}
}

func (d *DepthCache) syncOnce(ctx context.Context) error {
	book, err := d.fetchSnapshot(ctx)
	if err != nil {
		return err
	}
	d.mu.Lock()
	d.replace(book)
	d.mu.Unlock()
	d.notify()

	if err := d.drainQueued(book.LastUpdateID); err != nil {
		return err
	}
	for {
		ev, err := d.stream.Next(ctx)
		if err != nil {
			return err
		}
		if ev.Depth == nil {
			continue
		}
		if ev.Depth.FinalUpdateID <= d.LastUpdateID() && ev.Depth.FinalUpdateID <= book.LastUpdateID {
			continue
		}
		if err := d.apply(*ev.Depth); err != nil {
			return err
		}
	}
}

func (d *DepthCache) drainQueued(snapID int64) error {
	for {
		select {
		case ev, ok := <-d.stream.events:
			if !ok {
				return ErrStreamClosed
			}
			if ev.Depth == nil {
				continue
			}
			if ev.Depth.FinalUpdateID <= snapID {
				continue
			}
			if err := d.apply(*ev.Depth); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (d *DepthCache) fetchSnapshot(ctx context.Context) (*OrderBook, error) {
	req := OrderBookRequest{Symbol: d.symbol, Limit: 1000}
	switch d.market {
	case MarketUSDFutures:
		return d.c.Futures().OrderBook(ctx, req)
	case MarketCoinFutures:
		return d.c.CoinFutures().OrderBook(ctx, req)
	default:
		return d.c.Spot().OrderBook(ctx, req)
	}
}

func (d *DepthCache) replace(book *OrderBook) {
	d.bids = make(map[string]string, len(book.Bids))
	d.asks = make(map[string]string, len(book.Asks))
	for _, lv := range book.Bids {
		if lv.Quantity != "0" {
			d.bids[lv.Price] = lv.Quantity
		}
	}
	for _, lv := range book.Asks {
		if lv.Quantity != "0" {
			d.asks[lv.Price] = lv.Quantity
		}
	}
	d.lastID = book.LastUpdateID
	d.updateTime = time.Now()
}

func (d *DepthCache) apply(ev DepthEvent) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lastID == 0 {
		applyLevels(d.bids, ev.Bids)
		applyLevels(d.asks, ev.Asks)
		d.lastID = ev.FinalUpdateID
		d.updateTime = ev.EventTime.Time()
		d.notifyLocked()
		return nil
	}
	if ev.PrevUpdateID != 0 {
		if ev.PrevUpdateID != d.lastID {
			return fmt.Errorf("%w: pu=%d last=%d", ErrDepthDesync, ev.PrevUpdateID, d.lastID)
		}
	} else if ev.FirstUpdateID > d.lastID+1 {
		return fmt.Errorf("%w: U=%d last=%d", ErrDepthDesync, ev.FirstUpdateID, d.lastID)
	}
	if ev.FinalUpdateID < d.lastID {
		return nil
	}
	applyLevels(d.bids, ev.Bids)
	applyLevels(d.asks, ev.Asks)
	d.lastID = ev.FinalUpdateID
	d.updateTime = ev.EventTime.Time()
	d.notifyLocked()
	return nil
}

func applyLevels(book map[string]string, levels []PriceLevel) {
	for _, lv := range levels {
		if lv.Quantity == "0" || lv.Quantity == "0.00000000" {
			delete(book, lv.Price)
			continue
		}
		book[lv.Price] = lv.Quantity
	}
}

func (d *DepthCache) notify() {
	d.mu.Lock()
	d.notifyLocked()
	d.mu.Unlock()
}

func (d *DepthCache) notifyLocked() {
	select {
	case d.updates <- struct{}{}:
	default:
	}
}

func (d *DepthCache) setErr(err error) {
	d.mu.Lock()
	d.err = err
	d.mu.Unlock()
	d.notify()
}

func sortedLevels(m map[string]string, bids bool) []PriceLevel {
	out := make([]PriceLevel, 0, len(m))
	for p, q := range m {
		out = append(out, PriceLevel{Price: p, Quantity: q})
	}
	sort.Slice(out, func(i, j int) bool {
		ai, _ := strconv.ParseFloat(out[i].Price, 64)
		aj, _ := strconv.ParseFloat(out[j].Price, 64)
		if bids {
			return ai > aj
		}
		return ai < aj
	})
	return out
}
