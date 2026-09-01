package binance

import "context"

// SubAccountService covers a useful subset of sub-account endpoints.
type SubAccountService struct{ c *Client }

// SubAccount is one sub-account list row.
type SubAccount struct {
	Email                       string `json:"email"`
	IsFreeze                    bool   `json:"isFreeze"`
	CreateTime                  TimeMS `json:"createTime"`
	IsManagedSubAccount         bool   `json:"isManagedSubAccount"`
	IsAssetManagementSubAccount bool   `json:"isAssetManagementSubAccount"`
}

// List returns sub-accounts.
func (s *SubAccountService) List(ctx context.Context, email string, recv int64) ([]SubAccount, error) {
	p := newParams()
	p.Set("email", email)
	var wrap struct {
		SubAccounts []SubAccount `json:"subAccounts"`
	}
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/sub-account/list", p, &wrap, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return wrap.SubAccounts, nil
}

// Assets returns a sub-account's spot assets.
func (s *SubAccountService) Assets(ctx context.Context, email string, recv int64) ([]Balance, error) {
	p := newParams()
	p.Set("email", email)
	var wrap struct {
		Balances []Balance `json:"balances"`
	}
	if err := s.c.get(ctx, familySAPI, "/sapi/v3/sub-account/assets", p, &wrap, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return wrap.Balances, nil
}

// SubAccountTransferRequest is POST /sapi/v1/sub-account/universalTransfer.
type SubAccountTransferRequest struct {
	FromEmail       string
	ToEmail         string
	FromAccountType string
	ToAccountType   string
	Asset           string
	Amount          string
	RecvWindow      int64
}

// UniversalTransfer transfers between master/sub accounts. Not retried.
func (s *SubAccountService) UniversalTransfer(ctx context.Context, req SubAccountTransferRequest) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("fromEmail", req.FromEmail)
	p.Set("toEmail", req.ToEmail)
	p.Set("fromAccountType", req.FromAccountType)
	p.Set("toAccountType", req.ToAccountType)
	p.Set("asset", req.Asset)
	p.Set("amount", req.Amount)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/sub-account/universalTransfer", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}
