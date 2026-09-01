package binance

import "context"

// WalletService covers capital / wallet REST endpoints (sapi).
type WalletService struct{ c *Client }

// SystemStatus is GET /sapi/v1/system/status.
type SystemStatus struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

// SystemStatus returns 0 when the system is normal.
func (s *WalletService) SystemStatus(ctx context.Context) (*SystemStatus, error) {
	var out SystemStatus
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/system/status", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CoinInfo is one entry from GET /sapi/v1/capital/config/getall.
type CoinInfo struct {
	Coin              string        `json:"coin"`
	DepositAllEnable  bool          `json:"depositAllEnable"`
	WithdrawAllEnable bool          `json:"withdrawAllEnable"`
	Name              string        `json:"name"`
	Free              string        `json:"free"`
	Locked            string        `json:"locked"`
	Freeze            string        `json:"freeze"`
	Withdrawing       string        `json:"withdrawing"`
	Ipoing            string        `json:"ipoing"`
	Ipoable           string        `json:"ipoable"`
	Storage           string        `json:"storage"`
	IsLegalMoney      bool          `json:"isLegalMoney"`
	Trading           bool          `json:"trading"`
	NetworkList       []CoinNetwork `json:"networkList"`
}

// CoinNetwork describes a withdraw/deposit network.
type CoinNetwork struct {
	Network                 string `json:"network"`
	Coin                    string `json:"coin"`
	WithdrawIntegerMultiple string `json:"withdrawIntegerMultiple"`
	IsDefault               bool   `json:"isDefault"`
	DepositEnable           bool   `json:"depositEnable"`
	WithdrawEnable          bool   `json:"withdrawEnable"`
	DepositDesc             string `json:"depositDesc"`
	WithdrawDesc            string `json:"withdrawDesc"`
	SpecialTips             string `json:"specialTips"`
	Name                    string `json:"name"`
	ResetAddressStatus      bool   `json:"resetAddressStatus"`
	AddressRegex            string `json:"addressRegex"`
	MemoRegex               string `json:"memoRegex"`
	WithdrawFee             string `json:"withdrawFee"`
	WithdrawMin             string `json:"withdrawMin"`
	WithdrawMax             string `json:"withdrawMax"`
	MinConfirm              int    `json:"minConfirm"`
	UnLockConfirm           int    `json:"unLockConfirm"`
}

// Coins returns all coins' configuration and balances.
func (s *WalletService) Coins(ctx context.Context, recv int64) ([]CoinInfo, error) {
	var out []CoinInfo
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/capital/config/getall", newParams(), &out, signed(), weight(10), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// DepositAddressRequest is GET /sapi/v1/capital/deposit/address.
type DepositAddressRequest struct {
	Coin       string
	Network    string
	RecvWindow int64
}

// DepositAddress is a deposit address.
type DepositAddress struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Tag     string `json:"tag"`
	URL     string `json:"url"`
}

// DepositAddress returns a deposit address for coin/network.
func (s *WalletService) DepositAddress(ctx context.Context, req DepositAddressRequest) (*DepositAddress, error) {
	p := newParams()
	p.Set("coin", req.Coin)
	p.Set("network", req.Network)
	var out DepositAddress
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/capital/deposit/address", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// DepositHistoryRequest is GET /sapi/v1/capital/deposit/hisrec.
type DepositHistoryRequest struct {
	Coin       string
	Status     int
	StartTime  int64
	EndTime    int64
	Offset     int
	Limit      int
	TxID       string
	RecvWindow int64
}

// Deposit is one deposit history row.
type Deposit struct {
	ID            string `json:"id"`
	Amount        string `json:"amount"`
	Coin          string `json:"coin"`
	Network       string `json:"network"`
	Status        int    `json:"status"`
	Address       string `json:"address"`
	AddressTag    string `json:"addressTag"`
	TxID          string `json:"txId"`
	InsertTime    TimeMS `json:"insertTime"`
	TransferType  int    `json:"transferType"`
	ConfirmTimes  string `json:"confirmTimes"`
	UnlockConfirm int    `json:"unlockConfirm"`
	WalletType    int    `json:"walletType"`
}

// DepositHistory returns deposit records.
func (s *WalletService) DepositHistory(ctx context.Context, req DepositHistoryRequest) ([]Deposit, error) {
	p := newParams()
	p.Set("coin", req.Coin)
	if req.Status != 0 {
		p.SetInt("status", req.Status)
	}
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("offset", req.Offset)
	p.SetInt("limit", req.Limit)
	p.Set("txId", req.TxID)
	var out []Deposit
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/capital/deposit/hisrec", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// WithdrawRequest is POST /sapi/v1/capital/withdraw/apply.
//
// Security: this moves real funds. Restrict API keys by IP, enable
// withdrawal whitelist, and never commit credentials. This method is
// not retried.
type WithdrawRequest struct {
	Coin               string
	WithdrawOrderID    string
	Network            string
	Address            string
	AddressTag         string
	Amount             string
	TransactionFeeFlag bool
	Name               string
	WalletType         int
	RecvWindow         int64
}

// WithdrawResult is the withdraw ticket id.
type WithdrawResult struct {
	ID string `json:"id"`
}

// Withdraw submits a withdrawal. This call is not retried.
func (s *WalletService) Withdraw(ctx context.Context, req WithdrawRequest) (*WithdrawResult, error) {
	p := newParams()
	p.Set("coin", req.Coin)
	p.Set("withdrawOrderId", req.WithdrawOrderID)
	p.Set("network", req.Network)
	p.Set("address", req.Address)
	p.Set("addressTag", req.AddressTag)
	p.Set("amount", req.Amount)
	p.SetBool("transactionFeeFlag", req.TransactionFeeFlag)
	p.Set("name", req.Name)
	p.SetInt("walletType", req.WalletType)
	var out WithdrawResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/capital/withdraw/apply", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// WithdrawHistoryRequest is GET /sapi/v1/capital/withdraw/history.
type WithdrawHistoryRequest struct {
	Coin            string
	WithdrawOrderID string
	Status          int
	StartTime       int64
	EndTime         int64
	Offset          int
	Limit           int
	RecvWindow      int64
}

// Withdrawal is one withdraw history row.
type Withdrawal struct {
	ID              string `json:"id"`
	Amount          string `json:"amount"`
	TransactionFee  string `json:"transactionFee"`
	Coin            string `json:"coin"`
	Status          int    `json:"status"`
	Address         string `json:"address"`
	TxID            string `json:"txId"`
	ApplyTime       string `json:"applyTime"`
	Network         string `json:"network"`
	TransferType    int    `json:"transferType"`
	WithdrawOrderID string `json:"withdrawOrderId"`
	Info            string `json:"info"`
	ConfirmNo       int    `json:"confirmNo"`
	WalletType      int    `json:"walletType"`
	TxKey           string `json:"txKey"`
}

// WithdrawHistory returns withdrawal records.
func (s *WalletService) WithdrawHistory(ctx context.Context, req WithdrawHistoryRequest) ([]Withdrawal, error) {
	p := newParams()
	p.Set("coin", req.Coin)
	p.Set("withdrawOrderId", req.WithdrawOrderID)
	p.SetInt("status", req.Status)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("offset", req.Offset)
	p.SetInt("limit", req.Limit)
	var out []Withdrawal
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/capital/withdraw/history", p, &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// AccountSnapshotRequest is GET /sapi/v1/accountSnapshot.
type AccountSnapshotRequest struct {
	Type       string // SPOT, MARGIN, FUTURES
	StartTime  int64
	EndTime    int64
	Limit      int
	RecvWindow int64
}

// AccountSnapshot is a daily account snapshot.
type AccountSnapshot struct {
	Code        int                      `json:"code"`
	Msg         string                   `json:"msg"`
	SnapshotVos []AccountSnapshotPayload `json:"snapshotVos"`
}

// AccountSnapshotPayload is one snapshot day.
type AccountSnapshotPayload struct {
	Type       string          `json:"type"`
	UpdateTime TimeMS          `json:"updateTime"`
	Data       jsonRawSnapshot `json:"data"`
}

// jsonRawSnapshot keeps snapshot data typed enough without over-fitting variants.
type jsonRawSnapshot struct {
	Balances           []Balance `json:"balances"`
	TotalAssetOfBtc    string    `json:"totalAssetOfBtc"`
	MarginLevel        string    `json:"marginLevel"`
	TotalNetAssetOfBtc string    `json:"totalNetAssetOfBtc"`
}

// AccountSnapshot returns a daily account snapshot.
func (s *WalletService) AccountSnapshot(ctx context.Context, req AccountSnapshotRequest) (*AccountSnapshot, error) {
	p := newParams()
	p.Set("type", req.Type)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	var out AccountSnapshot
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/accountSnapshot", p, &out, signed(), weight(2400), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// AssetDetail is GET /sapi/v1/asset/assetDetail.
type AssetDetail map[string]AssetDetailEntry

// AssetDetailEntry is one asset's withdraw/deposit flags.
type AssetDetailEntry struct {
	MinWithdrawAmount string `json:"minWithdrawAmount"`
	DepositStatus     bool   `json:"depositStatus"`
	WithdrawFee       any    `json:"withdrawFee"`
	WithdrawStatus    bool   `json:"withdrawStatus"`
	DepositTip        string `json:"depositTip"`
}

// AssetDetail returns asset deposit/withdraw details.
func (s *WalletService) AssetDetail(ctx context.Context, asset string, recv int64) (AssetDetail, error) {
	p := newParams()
	p.Set("asset", asset)
	var out AssetDetail
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/asset/assetDetail", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// TradeFee is GET /sapi/v1/asset/tradeFee.
type TradeFee struct {
	Symbol          string `json:"symbol"`
	MakerCommission string `json:"makerCommission"`
	TakerCommission string `json:"takerCommission"`
}

// TradeFee returns trade fees, optionally filtered by symbol.
func (s *WalletService) TradeFee(ctx context.Context, symbol string, recv int64) ([]TradeFee, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []TradeFee
	if err := s.c.get(ctx, familySAPI, "/sapi/v1/asset/tradeFee", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// UniversalTransferRequest is POST /sapi/v1/asset/transfer.
type UniversalTransferRequest struct {
	Type       TransferType
	Asset      string
	Amount     string
	FromSymbol string
	ToSymbol   string
	RecvWindow int64
}

// UniversalTransferResult is a transfer id.
type UniversalTransferResult struct {
	TranID int64 `json:"tranId"`
}

// UniversalTransfer moves assets between Binance wallets. Not retried.
func (s *WalletService) UniversalTransfer(ctx context.Context, req UniversalTransferRequest) (*UniversalTransferResult, error) {
	p := newParams()
	p.Set("type", string(req.Type))
	p.Set("asset", req.Asset)
	p.Set("amount", req.Amount)
	p.Set("fromSymbol", req.FromSymbol)
	p.Set("toSymbol", req.ToSymbol)
	var out UniversalTransferResult
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/asset/transfer", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// DustTransfer converts small balances to BNB. Not retried.
func (s *WalletService) DustTransfer(ctx context.Context, assets []string, recv int64) (jsonRawSnapshot, error) {
	p := newParams()
	if len(assets) > 0 {
		raw, err := marshalJSONArray(assets)
		if err != nil {
			return jsonRawSnapshot{}, err
		}
		p.SetRaw("asset", raw)
	}
	var out jsonRawSnapshot
	if err := s.c.post(ctx, familySAPI, "/sapi/v1/asset/dust", p, &out, signed(), noRetry(), recvWindow(recv)); err != nil {
		return jsonRawSnapshot{}, err
	}
	return out, nil
}
