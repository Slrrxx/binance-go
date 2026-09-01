package binance

import (
	"context"
	"fmt"
)

const klinePageLimit = 1000

// KlineQuery describes a historical kline range.
type KlineQuery struct {
	Symbol    string
	Interval  Interval
	StartTime int64
	EndTime   int64
	Market    Market
	Limit     int // per-request page size, default 1000
}

func (q KlineQuery) pageSize() int {
	if q.Limit <= 0 || q.Limit > klinePageLimit {
		return klinePageLimit
	}
	return q.Limit
}

// GetHistoricalKlines fetches all klines in [start, end] into memory.
// Prefer HistoricalKlines for large ranges.
func (c *Client) GetHistoricalKlines(ctx context.Context, symbol string, interval Interval, startTime, endTime int64) ([]Kline, error) {
	it := c.HistoricalKlines(KlineQuery{
		Symbol:    symbol,
		Interval:  interval,
		StartTime: startTime,
		EndTime:   endTime,
		Market:    MarketSpot,
	})
	var out []Kline
	for it.Next(ctx) {
		out = append(out, it.Kline())
	}
	return out, it.Err()
}

// HistoricalKlines returns a memory-efficient iterator over paginated klines.
func (c *Client) HistoricalKlines(q KlineQuery) *KlineIterator {
	return &KlineIterator{c: c, q: q}
}

// KlineIterator pages through historical candlesticks like database/sql Rows.
type KlineIterator struct {
	c       *Client
	q       KlineQuery
	buf     []Kline
	idx     int
	nextT   int64
	done    bool
	err     error
	started bool
}

// Next loads the next kline. It returns false on exhaustion or error.
func (it *KlineIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}
	if !it.started {
		it.started = true
		it.nextT = it.q.StartTime
	}
	if it.idx >= len(it.buf) {
		if it.done {
			return false
		}
		if err := it.fetch(ctx); err != nil {
			it.err = err
			return false
		}
		if len(it.buf) == 0 {
			it.done = true
			return false
		}
		it.idx = 0
	}
	it.idx++
	return true
}

// Kline returns the current kline after a successful Next.
func (it *KlineIterator) Kline() Kline {
	if it.idx <= 0 || it.idx > len(it.buf) {
		return Kline{}
	}
	return it.buf[it.idx-1]
}

// Err returns the first iterator error.
func (it *KlineIterator) Err() error { return it.err }

func (it *KlineIterator) fetch(ctx context.Context) error {
	req := KlinesRequest{
		Symbol:    it.q.Symbol,
		Interval:  it.q.Interval,
		StartTime: it.nextT,
		EndTime:   it.q.EndTime,
		Limit:     it.q.pageSize(),
	}
	var (
		page []Kline
		err  error
	)
	switch it.q.Market {
	case MarketUSDFutures:
		page, err = it.c.Futures().Klines(ctx, req)
	case MarketCoinFutures:
		page, err = it.c.CoinFutures().Klines(ctx, req)
	default:
		page, err = it.c.Spot().Klines(ctx, req)
	}
	if err != nil {
		return err
	}
	if len(page) == 0 {
		it.buf = nil
		return nil
	}
	if it.nextT != 0 && int64(page[0].OpenTime) < it.nextT {
		return fmt.Errorf("%w: start moved backwards", ErrNoProgress)
	}
	last := page[len(page)-1]
	next := int64(last.CloseTime) + 1
	if it.nextT != 0 && next <= it.nextT {
		return ErrNoProgress
	}
	it.nextT = next
	if it.q.EndTime > 0 && int64(last.OpenTime) >= it.q.EndTime {
		it.done = true
	}
	if len(page) < it.q.pageSize() {
		it.done = true
	}
	it.buf = page
	return nil
}

// AggTradeQuery pages public aggregate trades.
type AggTradeQuery struct {
	Symbol    string
	StartTime int64
	EndTime   int64
	FromID    int64
	Limit     int
}

// AggTradeIterator pages aggregate trades.
type AggTradeIterator struct {
	c       *Client
	q       AggTradeQuery
	buf     []AggTrade
	idx     int
	fromID  int64
	done    bool
	err     error
	started bool
}

// AggTrades returns an iterator over public aggregate trades.
func (c *Client) AggTrades(q AggTradeQuery) *AggTradeIterator {
	return &AggTradeIterator{c: c, q: q, fromID: q.FromID}
}

// Next advances the aggregate-trade iterator.
func (it *AggTradeIterator) Next(ctx context.Context) bool {
	if it.err != nil {
		return false
	}
	if it.idx >= len(it.buf) && it.done {
		return false
	}
	it.started = true
	if it.idx >= len(it.buf) {
		limit := it.q.Limit
		if limit <= 0 || limit > 1000 {
			limit = 1000
		}
		page, err := it.c.Spot().AggTrades(ctx, AggTradesRequest{
			Symbol:    it.q.Symbol,
			FromID:    it.fromID,
			StartTime: it.q.StartTime,
			EndTime:   it.q.EndTime,
			Limit:     limit,
		})
		if err != nil {
			it.err = err
			return false
		}
		if len(page) == 0 {
			it.done = true
			return false
		}
		last := page[len(page)-1]
		next := last.AggTradeID + 1
		if it.fromID != 0 && next <= it.fromID {
			it.err = ErrNoProgress
			return false
		}
		it.fromID = next
		it.q.StartTime = 0 // subsequent pages use fromId
		if len(page) < limit {
			it.done = true
		}
		it.buf = page
		it.idx = 0
	}
	it.idx++
	return true
}

// AggTrade returns the current aggregate trade.
func (it *AggTradeIterator) AggTrade() AggTrade {
	if it.idx <= 0 || it.idx > len(it.buf) {
		return AggTrade{}
	}
	return it.buf[it.idx-1]
}

// Err returns the iterator error.
func (it *AggTradeIterator) Err() error { return it.err }
