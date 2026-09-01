package binance

import "context"

// ConvertService covers the convert (quote) API.
type ConvertService struct{ c *Client }

// ConvertPair is GET /sapi/v1/convert/exchangeInfo.
type ConvertPair struct {
	FromAsset          string `json:"fromAsset"`
	ToAsset            string `json:"toAsset"`
	FromAssetMinAmount string `json:"fromAssetMinAmount"`
	FromAssetMaxAmount string `json:"fromAssetMaxAmount"`
	ToAssetMinAmount   string `json:"toAssetMinAmount"`
	ToAssetMaxAmount   string `json:"toAssetMaxAmount"`
}

// ExchangeInfo returns convert pairs.
func (s *ConvertService) ExchangeInfo(ctx context.Context, fromAsset, toAsset string) ([]ConvertPair, error) {
	p := newParams()
	p.Set("fromAsset", fromAsset)
	p.Set("toAsset", toAsset)
	var out []ConvertPair
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/convert/exchangeInfo", p, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ConvertQuote is a convert quote ticket.
type ConvertQuote struct {
	QuoteID        string `json:"quoteId"`
	Ratio          string `json:"ratio"`
	InverseRatio   string `json:"inverseRatio"`
	ValidTimestamp int64  `json:"validTimestamp"`
	ToAmount       string `json:"toAmount"`
	FromAmount     string `json:"fromAmount"`
}

// GetQuote requests a convert quote. Not retried.
func (s *ConvertService) GetQuote(ctx context.Context, fromAsset, toAsset, fromAmount string, recv int64) (*ConvertQuote, error) {
	p := newParams()
	p.Set("fromAsset", fromAsset)
	p.Set("toAsset", toAsset)
	p.Set("fromAmount", fromAmount)
	var out ConvertQuote
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/convert/getQuote", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// AcceptQuote accepts a convert quote. Not retried.
func (s *ConvertService) AcceptQuote(ctx context.Context, quoteID string, recv int64) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("quoteId", quoteID)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/convert/acceptQuote", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}
