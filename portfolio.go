package binance

import "context"

// PortfolioService is Binance portfolio margin (papi).
type PortfolioService struct{ c *Client }

// Ping tests papi connectivity.
func (s *PortfolioService) Ping(ctx context.Context) error {
	return s.c.get(ctx, familyPAPI, "/papi/v1/ping", nil, nil)
}

// Balance returns portfolio-margin balances.
func (s *PortfolioService) Balance(ctx context.Context, asset string, recv int64) ([]FuturesBalance, error) {
	p := newParams()
	p.Set("asset", asset)
	var out []FuturesBalance
	if err := s.c.get(ctx, familyPAPI, "/papi/v1/balance", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// Account returns the unified portfolio account.
func (s *PortfolioService) Account(ctx context.Context, recv int64) (*FuturesAccount, error) {
	var out FuturesAccount
	if err := s.c.get(ctx, familyPAPI, "/papi/v1/account", newParams(), &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// UMPositions returns USD-M position risk under portfolio margin.
func (s *PortfolioService) UMPositions(ctx context.Context, symbol string, recv int64) ([]FuturesPosition, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []FuturesPosition
	if err := s.c.get(ctx, familyPAPI, "/papi/v1/um/positionRisk", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// CMPositions returns COIN-M position risk under portfolio margin.
func (s *PortfolioService) CMPositions(ctx context.Context, symbol string, recv int64) ([]FuturesPosition, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []FuturesPosition
	if err := s.c.get(ctx, familyPAPI, "/papi/v1/cm/positionRisk", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateUMOrder places a USD-M order on portfolio margin. Not retried.
func (s *PortfolioService) CreateUMOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.post(ctx, familyPAPI, "/papi/v1/um/order", req.params(), &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateCMOrder places a COIN-M order on portfolio margin. Not retried.
func (s *PortfolioService) CreateCMOrder(ctx context.Context, req FuturesOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.post(ctx, familyPAPI, "/papi/v1/cm/order", req.params(), &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelUMOrder cancels a USD-M portfolio-margin order.
func (s *PortfolioService) CancelUMOrder(ctx context.Context, req CancelOrderRequest) (*FuturesOrder, error) {
	var out FuturesOrder
	if err := s.c.delete_(ctx, familyPAPI, "/papi/v1/um/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListenKey creates a portfolio-margin user-data listen key.
func (s *PortfolioService) ListenKey(ctx context.Context) (string, error) {
	var out ListenKeyResponse
	if err := s.c.post(ctx, familyPAPI, "/papi/v1/listenKey", nil, &out, apiKey()); err != nil {
		return "", err
	}
	return out.ListenKey, nil
}

// KeepListenKey keepalives a portfolio-margin listen key.
func (s *PortfolioService) KeepListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.put(ctx, familyPAPI, "/papi/v1/listenKey", p, nil, apiKey())
}

// CloseListenKey closes a portfolio-margin listen key.
func (s *PortfolioService) CloseListenKey(ctx context.Context, listenKey string) error {
	p := newParams()
	p.Set("listenKey", listenKey)
	return s.c.delete_(ctx, familyPAPI, "/papi/v1/listenKey", p, nil, apiKey())
}

// UserData opens a portfolio-margin user-data stream with listen-key keepalive.
func (s *PortfolioService) UserData(ctx context.Context) (*Stream, error) {
	return startUserData(ctx, s.c, MarketPortfolio, s.ListenKey, s.KeepListenKey, s.CloseListenKey)
}

// SetUMLeverage changes USD-M leverage under portfolio margin. Not retried.
func (s *PortfolioService) SetUMLeverage(ctx context.Context, symbol string, leverage int, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	p.SetInt("leverage", leverage)
	return s.c.post(ctx, familyPAPI, "/papi/v1/um/leverage", p, nil, signed(), noRetry(), recvWindow(recv))
}
