package binance

import "context"

// EarnService covers Simple Earn flexible/locked products.
type EarnService struct{ c *Client }

// EarnProduct is a Simple Earn product row.
type EarnProduct struct {
	Asset                      string            `json:"asset"`
	LatestAnnualPercentageRate string            `json:"latestAnnualPercentageRate"`
	TierAnnualPercentageRate   map[string]string `json:"tierAnnualPercentageRate"`
	CanPurchase                bool              `json:"canPurchase"`
	CanRedeem                  bool              `json:"canRedeem"`
	IsSoldOut                  bool              `json:"isSoldOut"`
	Hot                        bool              `json:"hot"`
	MinPurchaseAmount          string            `json:"minPurchaseAmount"`
	ProductID                  string            `json:"productId"`
	Status                     string            `json:"status"`
	SubscriptionStartTime      TimeMS            `json:"subscriptionStartTime"`
}

// EarnPage is a paginated Simple Earn list.
type EarnPage struct {
	Rows  []EarnProduct `json:"rows"`
	Total int           `json:"total"`
}

// FlexibleProducts lists flexible Simple Earn products.
func (s *EarnService) FlexibleProducts(ctx context.Context, asset string, recv int64) (*EarnPage, error) {
	p := newParams()
	p.Set("asset", asset)
	var out EarnPage
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/simple-earn/flexible/list", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// LockedProducts lists locked Simple Earn products.
func (s *EarnService) LockedProducts(ctx context.Context, asset string, recv int64) (*EarnPage, error) {
	p := newParams()
	p.Set("asset", asset)
	var out EarnPage
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/simple-earn/locked/list", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubscribeFlexible subscribes to a flexible product. Not retried.
func (s *EarnService) SubscribeFlexible(ctx context.Context, productID, amount string, recv int64) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("productId", productID)
	p.Set("amount", amount)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/simple-earn/flexible/subscribe", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// RedeemFlexible redeems a flexible product. Not retried.
func (s *EarnService) RedeemFlexible(ctx context.Context, productID, amount string, recv int64) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("productId", productID)
	p.Set("amount", amount)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/simple-earn/flexible/redeem", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Account returns Simple Earn account totals.
func (s *EarnService) Account(ctx context.Context, recv int64) (map[string]string, error) {
	var out map[string]string
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/simple-earn/account", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}
