package binance

import "context"

// AccountRequest is GET /api/v3/account.
type AccountRequest struct {
	OmitZeroBalances bool
	RecvWindow       int64
}

// Account returns current account information.
func (s *SpotService) Account(ctx context.Context, req AccountRequest) (*Account, error) {
	p := newParams()
	p.SetBool("omitZeroBalances", req.OmitZeroBalances)
	var out Account
	if err := s.c.get(ctx, familySpot, "/api/v3/account", p, &out, signed(), weight(20), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// BalanceOf returns a single asset from an account snapshot.
func (a *Account) BalanceOf(asset string) (Balance, bool) {
	for _, b := range a.Balances {
		if b.Asset == asset {
			return b, true
		}
	}
	return Balance{}, false
}

// MyTradesRequest is GET /api/v3/myTrades.
type MyTradesRequest struct {
	Symbol     string
	OrderID    int64
	StartTime  int64
	EndTime    int64
	FromID     int64
	Limit      int
	RecvWindow int64
}

// MyTrades returns account trade history for a symbol.
func (s *SpotService) MyTrades(ctx context.Context, req MyTradesRequest) ([]MyTrade, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt64("fromId", req.FromID)
	p.SetInt("limit", req.Limit)
	var out []MyTrade
	if err := s.c.get(ctx, familySpot, "/api/v3/myTrades", p, &out, signed(), weight(20), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// OrderCountUsage returns the current order count rate limits.
func (s *SpotService) OrderCountUsage(ctx context.Context, recv int64) ([]OrderCount, error) {
	var out []OrderCount
	if err := s.c.get(ctx, familySpot, "/api/v3/rateLimit/order", newParams(), &out, signed(), weight(40), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// AccountStatus is GET /sapi/v1/account/status.
type AccountStatus struct {
	Data string `json:"data"`
}

// AccountAPIStatus is GET /sapi/v1/account/apiTradingStatus.
type AccountAPIStatus struct {
	Data struct {
		IsLocked           bool  `json:"isLocked"`
		PlannedRecoverTime int64 `json:"plannedRecoverTime"`
		TriggerCondition   struct {
			GCR  int `json:"GCR"`
			IFER int `json:"IFER"`
			UFR  int `json:"UFR"`
		} `json:"triggerCondition"`
		UpdateTime TimeMS `json:"updateTime"`
	} `json:"data"`
}

// APITradingStatus returns account API trading status (sapi).
func (s *SpotService) APITradingStatus(ctx context.Context, recv int64) (*AccountAPIStatus, error) {
	var out AccountAPIStatus
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/account/apiTradingStatus", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// SAPIAccountStatus returns account status (sapi).
func (s *SpotService) SAPIAccountStatus(ctx context.Context, recv int64) (*AccountStatus, error) {
	var out AccountStatus
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/account/status", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}
