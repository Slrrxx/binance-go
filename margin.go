package binance

import "context"

// MarginService covers cross and isolated margin.
type MarginService struct{ c *Client }

// MarginAccount is GET /sapi/v1/margin/account.
type MarginAccount struct {
	BorrowEnabled       bool          `json:"borrowEnabled"`
	MarginLevel         string        `json:"marginLevel"`
	TotalAssetOfBtc     string        `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc string        `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  string        `json:"totalNetAssetOfBtc"`
	TradeEnabled        bool          `json:"tradeEnabled"`
	TransferEnabled     bool          `json:"transferEnabled"`
	AccountType         string        `json:"accountType"`
	UserAssets          []MarginAsset `json:"userAssets"`
}

// MarginAsset is one cross-margin asset.
type MarginAsset struct {
	Asset    string `json:"asset"`
	Borrowed string `json:"borrowed"`
	Free     string `json:"free"`
	Interest string `json:"interest"`
	Locked   string `json:"locked"`
	NetAsset string `json:"netAsset"`
}

// Account returns the cross-margin account.
func (s *MarginService) Account(ctx context.Context, recv int64) (*MarginAccount, error) {
	var out MarginAccount
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/account", newParams(), &out, signed(), weight(10), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// IsolatedMarginAccount is GET /sapi/v1/margin/isolated/account.
type IsolatedMarginAccount struct {
	Assets              []IsolatedMarginPair `json:"assets"`
	TotalAssetOfBtc     string               `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc string               `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  string               `json:"totalNetAssetOfBtc"`
}

// IsolatedMarginPair is one isolated symbol bucket.
type IsolatedMarginPair struct {
	Symbol            string      `json:"symbol"`
	IsolatedCreated   bool        `json:"isolatedCreated"`
	Enabled           bool        `json:"enabled"`
	MarginLevel       string      `json:"marginLevel"`
	MarginLevelStatus string      `json:"marginLevelStatus"`
	MarginRatio       string      `json:"marginRatio"`
	IndexPrice        string      `json:"indexPrice"`
	LiquidatePrice    string      `json:"liquidatePrice"`
	LiquidateRate     string      `json:"liquidateRate"`
	TradeEnabled      bool        `json:"tradeEnabled"`
	BaseAsset         MarginAsset `json:"baseAsset"`
	QuoteAsset        MarginAsset `json:"quoteAsset"`
}

// IsolatedAccount returns isolated margin account information.
func (s *MarginService) IsolatedAccount(ctx context.Context, symbols string, recv int64) (*IsolatedMarginAccount, error) {
	p := newParams()
	p.Set("symbols", symbols)
	var out IsolatedMarginAccount
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/isolated/account", p, &out, signed(), weight(10), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// MarginOrderRequest is POST /sapi/v1/margin/order.
type MarginOrderRequest struct {
	OrderRequest
	IsIsolated        string
	SideEffectType    string
	AutoRepayAtCancel *bool
}

func (r MarginOrderRequest) params() params {
	p := r.OrderRequest.params()
	p.Set("isIsolated", r.IsIsolated)
	p.Set("sideEffectType", r.SideEffectType)
	p.SetBoolPtr("autoRepayAtCancel", r.AutoRepayAtCancel)
	return p
}

// CreateOrder places a margin order. Not retried.
func (s *MarginService) CreateOrder(ctx context.Context, req MarginOrderRequest) (*Order, error) {
	var out Order
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/margin/order", req.params(), &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrder cancels a margin order.
func (s *MarginService) CancelOrder(ctx context.Context, req CancelOrderRequest, isolated string) (*Order, error) {
	p := req.params()
	p.Set("isIsolated", isolated)
	var out Order
	if err := s.c.delete_(ctx, familySAPI, "/sapi/v1/margin/order", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetOrder queries a margin order.
func (s *MarginService) GetOrder(ctx context.Context, req QueryOrderRequest, isolated string) (*Order, error) {
	p := req.params()
	p.Set("isIsolated", isolated)
	var out Order
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/order", p, &out, signed(), weight(10), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenOrders returns open margin orders.
func (s *MarginService) OpenOrders(ctx context.Context, symbol, isolated string, recv int64) ([]Order, error) {
	p := newParams()
	p.Set("symbol", symbol)
	p.Set("isIsolated", isolated)
	var out []Order
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/openOrders", p, &out, signed(), weight(10), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// AllOrders returns all margin orders for a symbol.
func (s *MarginService) AllOrders(ctx context.Context, req AllOrdersRequest, isolated string) ([]Order, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	p.Set("isIsolated", isolated)
	var out []Order
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/allOrders", p, &out, signed(), weight(200), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// BorrowRepayRequest is POST /sapi/v1/margin/borrow-repay.
type BorrowRepayRequest struct {
	Asset      string
	IsIsolated string
	Symbol     string
	Amount     string
	Type       string // BORROW or REPAY
	RecvWindow int64
}

// BorrowRepayResult is a margin loan/repay ticket.
type BorrowRepayResult struct {
	TranID int64 `json:"tranId"`
}

func (s *MarginService) borrowRepay(ctx context.Context, req BorrowRepayRequest) (*BorrowRepayResult, error) {
	p := newParams()
	p.Set("asset", req.Asset)
	p.Set("isIsolated", req.IsIsolated)
	p.Set("symbol", req.Symbol)
	p.Set("amount", req.Amount)
	p.Set("type", req.Type)
	var out BorrowRepayResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/margin/borrow-repay", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// Borrow borrows an asset. Not retried.
func (s *MarginService) Borrow(ctx context.Context, req BorrowRepayRequest) (*BorrowRepayResult, error) {
	req.Type = "BORROW"
	return s.borrowRepay(ctx, req)
}

// Repay repays a margin loan. Not retried.
func (s *MarginService) Repay(ctx context.Context, req BorrowRepayRequest) (*BorrowRepayResult, error) {
	req.Type = "REPAY"
	return s.borrowRepay(ctx, req)
}

// TransferRequest is POST /sapi/v1/margin/transfer (cross).
type TransferRequest struct {
	Asset      string
	Amount     string
	Type       int // 1: main->margin, 2: margin->main
	RecvWindow int64
}

// Transfer moves funds between spot and cross margin. Not retried.
func (s *MarginService) Transfer(ctx context.Context, req TransferRequest) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("asset", req.Asset)
	p.Set("amount", req.Amount)
	p.SetInt("type", req.Type)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/margin/transfer", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// IsolatedTransferRequest is POST /sapi/v1/margin/isolated/transfer.
type IsolatedTransferRequest struct {
	Asset      string
	Symbol     string
	TransFrom  string
	TransTo    string
	Amount     string
	RecvWindow int64
}

// IsolatedTransfer moves funds for isolated margin. Not retried.
func (s *MarginService) IsolatedTransfer(ctx context.Context, req IsolatedTransferRequest) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("asset", req.Asset)
	p.Set("symbol", req.Symbol)
	p.Set("transFrom", req.TransFrom)
	p.Set("transTo", req.TransTo)
	p.Set("amount", req.Amount)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/margin/isolated/transfer", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// InterestHistoryRequest is GET /sapi/v1/margin/interestHistory.
type InterestHistoryRequest struct {
	Asset          string
	IsolatedSymbol string
	StartTime      int64
	EndTime        int64
	Current        int
	Size           int
	RecvWindow     int64
}

// InterestHistoryPage is a paginated interest history.
type InterestHistoryPage struct {
	Rows  []InterestRecord `json:"rows"`
	Total int              `json:"total"`
}

// InterestRecord is one interest row.
type InterestRecord struct {
	TxID                int64  `json:"txId"`
	InterestAccuredTime TimeMS `json:"interestAccuredTime"`
	Asset               string `json:"asset"`
	RawAsset            string `json:"rawAsset"`
	Principal           string `json:"principal"`
	Interest            string `json:"interest"`
	InterestRate        string `json:"interestRate"`
	Type                string `json:"type"`
	IsolatedSymbol      string `json:"isolatedSymbol"`
}

// InterestHistory returns margin interest records.
func (s *MarginService) InterestHistory(ctx context.Context, req InterestHistoryRequest) (*InterestHistoryPage, error) {
	p := newParams()
	p.Set("asset", req.Asset)
	p.Set("isolatedSymbol", req.IsolatedSymbol)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("current", req.Current)
	p.SetInt("size", req.Size)
	var out InterestHistoryPage
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/interestHistory", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// MaxBorrowable is GET /sapi/v1/margin/maxBorrowable.
type MaxBorrowable struct {
	Amount      string `json:"amount"`
	BorrowLimit string `json:"borrowLimit"`
}

// MaxBorrowable returns the max borrowable amount.
func (s *MarginService) MaxBorrowable(ctx context.Context, asset, isolatedSymbol string, recv int64) (*MaxBorrowable, error) {
	p := newParams()
	p.Set("asset", asset)
	p.Set("isolatedSymbol", isolatedSymbol)
	var out MaxBorrowable
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/margin/maxBorrowable", p, &out, signed(), weight(50), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// EnableIsolated enables an isolated margin account for a symbol.
func (s *MarginService) EnableIsolated(ctx context.Context, symbol string, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	return s.c.post(ctx, familySAPI, "/sapi/v1/margin/isolated/account", p, nil, signed(), recvWindow(recv))
}

// DisableIsolated disables an isolated margin account for a symbol.
func (s *MarginService) DisableIsolated(ctx context.Context, symbol string, recv int64) error {
	p := newParams()
	p.Set("symbol", symbol)
	return s.c.delete_(ctx, familySAPI, "/sapi/v1/margin/isolated/account", p, nil, signed(), recvWindow(recv))
}

// ListenKey creates a cross-margin user-data listen key.
func (s *MarginService) ListenKey(ctx context.Context) (string, error) {
	var out ListenKeyResponse
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/userDataStream", nil, &out, apiKey()); err != nil {
		return "", err
	}
	return out.ListenKey, nil
}
