package binance

import (
	"context"
	"encoding/json"
)

// GiftCardService covers Binance gift-card endpoints.
type GiftCardService struct{ c *Client }

// CreateCode issues a gift card. Not retried.
func (s *GiftCardService) CreateCode(ctx context.Context, token, amount string, recv int64) (json.RawMessage, error) {
	p := newParams()
	p.Set("token", token)
	p.Set("amount", amount)
	var out json.RawMessage
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/giftcard/createCode", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// Redeem redeems a gift card. Not retried.
func (s *GiftCardService) Redeem(ctx context.Context, code string, recv int64) (json.RawMessage, error) {
	p := newParams()
	p.Set("code", code)
	var out json.RawMessage
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/giftcard/redeemCode", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// Verify checks a gift-card reference number.
func (s *GiftCardService) Verify(ctx context.Context, referenceNo string, recv int64) (json.RawMessage, error) {
	p := newParams()
	p.Set("referenceNo", referenceNo)
	var out json.RawMessage
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/giftcard/verify", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// TokenLimit returns gift-card buy limits for a token.
func (s *GiftCardService) TokenLimit(ctx context.Context, baseToken string, recv int64) (json.RawMessage, error) {
	p := newParams()
	p.Set("baseToken", baseToken)
	var out json.RawMessage
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/giftcard/buyCode/token-limit", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}
