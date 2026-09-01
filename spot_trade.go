package binance

import (
	"context"
	"encoding/json"
)

func marshalJSONArray(ss []string) (string, error) {
	b, err := json.Marshal(ss)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// OrderRequest is a spot new-order payload. Quantity and Price are decimal strings.
type OrderRequest struct {
	Symbol                  string
	Side                    Side
	Type                    OrderType
	TimeInForce             TimeInForce
	Quantity                string
	QuoteOrderQty           string
	Price                   string
	ClientOrderID           string
	StrategyID              int64
	StrategyType            int
	StopPrice               string
	TrailingDelta           int64
	IcebergQty              string
	NewOrderRespType        OrderRespType
	SelfTradePreventionMode SelfTradePreventionMode
	RecvWindow              int64
}

func (r OrderRequest) params() params {
	p := newParams()
	p.Set("symbol", r.Symbol)
	p.Set("side", string(r.Side))
	p.Set("type", string(r.Type))
	p.Set("timeInForce", string(r.TimeInForce))
	p.Set("quantity", r.Quantity)
	p.Set("quoteOrderQty", r.QuoteOrderQty)
	p.Set("price", r.Price)
	p.Set("newClientOrderId", r.ClientOrderID)
	p.SetInt64("strategyId", r.StrategyID)
	p.SetInt("strategyType", r.StrategyType)
	p.Set("stopPrice", r.StopPrice)
	p.SetInt64("trailingDelta", r.TrailingDelta)
	p.Set("icebergQty", r.IcebergQty)
	p.Set("newOrderRespType", string(r.NewOrderRespType))
	p.Set("selfTradePreventionMode", string(r.SelfTradePreventionMode))
	return p
}

// CreateOrder submits a new spot order. Not retried (non-idempotent).
func (s *SpotService) CreateOrder(ctx context.Context, req OrderRequest) (*Order, error) {
	var out Order
	opts := []callOpt{signed(), noRetry(), weight(1), recvWindow(req.RecvWindow)}
	if err := s.c.post(ctx, familySpot, "/api/v3/order", req.params(), &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// TestOrder validates a new order without sending it to the matching engine.
func (s *SpotService) TestOrder(ctx context.Context, req OrderRequest) error {
	return s.c.post(ctx, familySpot, "/api/v3/order/test", req.params(), nil, signed(), noRetry(), recvWindow(req.RecvWindow))
}

// QueryOrderRequest is GET /api/v3/order.
type QueryOrderRequest struct {
	Symbol            string
	OrderID           int64
	OrigClientOrderID string
	RecvWindow        int64
}

func (r QueryOrderRequest) params() params {
	p := newParams()
	p.Set("symbol", r.Symbol)
	p.SetInt64("orderId", r.OrderID)
	p.Set("origClientOrderId", r.OrigClientOrderID)
	return p
}

// GetOrder queries an order's status.
func (s *SpotService) GetOrder(ctx context.Context, req QueryOrderRequest) (*Order, error) {
	var out Order
	if err := s.c.get(ctx, familySpot, "/api/v3/order", req.params(), &out, signed(), weight(4), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrderRequest is DELETE /api/v3/order.
type CancelOrderRequest struct {
	Symbol             string
	OrderID            int64
	OrigClientOrderID  string
	NewClientOrderID   string
	CancelRestrictions string
	RecvWindow         int64
}

func (r CancelOrderRequest) params() params {
	p := newParams()
	p.Set("symbol", r.Symbol)
	p.SetInt64("orderId", r.OrderID)
	p.Set("origClientOrderId", r.OrigClientOrderID)
	p.Set("newClientOrderId", r.NewClientOrderID)
	p.Set("cancelRestrictions", r.CancelRestrictions)
	return p
}

// CancelOrder cancels an active order.
func (s *SpotService) CancelOrder(ctx context.Context, req CancelOrderRequest) (*Order, error) {
	var out Order
	if err := s.c.delete_(ctx, familySpot, "/api/v3/order", req.params(), &out, signed(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelAllOpenOrders cancels all open orders on a symbol.
func (s *SpotService) CancelAllOpenOrders(ctx context.Context, symbol string, recv int64) ([]Order, error) {
	p := newParams()
	p.Set("symbol", symbol)
	var out []Order
	if err := s.c.delete_(ctx, familySpot, "/api/v3/openOrders", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}

// OpenOrdersRequest is GET /api/v3/openOrders.
type OpenOrdersRequest struct {
	Symbol     string
	RecvWindow int64
}

// OpenOrders returns all open orders; omit Symbol to query all symbols (heavier weight).
func (s *SpotService) OpenOrders(ctx context.Context, req OpenOrdersRequest) ([]Order, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	w := 80
	if req.Symbol != "" {
		w = 6
	}
	var out []Order
	if err := s.c.get(ctx, familySpot, "/api/v3/openOrders", p, &out, signed(), weight(w), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// AllOrdersRequest is GET /api/v3/allOrders.
type AllOrdersRequest struct {
	Symbol     string
	OrderID    int64
	StartTime  int64
	EndTime    int64
	Limit      int
	RecvWindow int64
}

// AllOrders returns all account orders (active, canceled, filled) for a symbol.
func (s *SpotService) AllOrders(ctx context.Context, req AllOrdersRequest) ([]Order, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.SetInt64("orderId", req.OrderID)
	p.SetInt64("startTime", req.StartTime)
	p.SetInt64("endTime", req.EndTime)
	p.SetInt("limit", req.Limit)
	var out []Order
	if err := s.c.get(ctx, familySpot, "/api/v3/allOrders", p, &out, signed(), weight(20), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return out, nil
}

// CancelReplaceRequest is POST /api/v3/order/cancelReplace.
type CancelReplaceRequest struct {
	OrderRequest
	CancelReplaceMode   string
	CancelOrderID       int64
	CancelClientOrderID string
}

// CancelReplace cancels an order and places a new one atomically.
func (s *SpotService) CancelReplace(ctx context.Context, req CancelReplaceRequest) (*Order, error) {
	p := req.params()
	p.Set("cancelReplaceMode", req.CancelReplaceMode)
	p.SetInt64("cancelOrderId", req.CancelOrderID)
	p.Set("cancelOrigClientOrderId", req.CancelClientOrderID)
	var out Order
	if err := s.c.post(ctx, familySpot, "/api/v3/order/cancelReplace", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// OCORequest is POST /api/v3/orderList/oco.
type OCORequest struct {
	Symbol                  string
	Side                    Side
	Quantity                string
	ListClientOrderID       string
	AboveType               OrderType
	AbovePrice              string
	AboveStopPrice          string
	AboveClientOrderID      string
	AboveTimeInForce        TimeInForce
	BelowType               OrderType
	BelowPrice              string
	BelowStopPrice          string
	BelowClientOrderID      string
	BelowTimeInForce        TimeInForce
	NewOrderRespType        OrderRespType
	SelfTradePreventionMode SelfTradePreventionMode
	RecvWindow              int64
}

// OrderList is an OCO/OTO response.
type OrderList struct {
	OrderListID       int64   `json:"orderListId"`
	ContingencyType   string  `json:"contingencyType"`
	ListStatusType    string  `json:"listStatusType"`
	ListOrderStatus   string  `json:"listOrderStatus"`
	ListClientOrderID string  `json:"listClientOrderId"`
	TransactionTime   TimeMS  `json:"transactionTime"`
	Symbol            string  `json:"symbol"`
	Orders            []Order `json:"orders"`
	OrderReports      []Order `json:"orderReports"`
}

// CreateOCO places an OCO order list.
func (s *SpotService) CreateOCO(ctx context.Context, req OCORequest) (*OrderList, error) {
	p := newParams()
	p.Set("symbol", req.Symbol)
	p.Set("side", string(req.Side))
	p.Set("quantity", req.Quantity)
	p.Set("listClientOrderId", req.ListClientOrderID)
	p.Set("aboveType", string(req.AboveType))
	p.Set("abovePrice", req.AbovePrice)
	p.Set("aboveStopPrice", req.AboveStopPrice)
	p.Set("aboveClientOrderId", req.AboveClientOrderID)
	p.Set("aboveTimeInForce", string(req.AboveTimeInForce))
	p.Set("belowType", string(req.BelowType))
	p.Set("belowPrice", req.BelowPrice)
	p.Set("belowStopPrice", req.BelowStopPrice)
	p.Set("belowClientOrderId", req.BelowClientOrderID)
	p.Set("belowTimeInForce", string(req.BelowTimeInForce))
	p.Set("newOrderRespType", string(req.NewOrderRespType))
	p.Set("selfTradePreventionMode", string(req.SelfTradePreventionMode))
	var out OrderList
	if err := s.c.post(ctx, familySpot, "/api/v3/orderList/oco", p, &out, signed(), noRetry(), recvWindow(req.RecvWindow)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CancelOrderList cancels an OCO/OTO list.
func (s *SpotService) CancelOrderList(ctx context.Context, symbol string, orderListID int64, listClientOrderID string, recv int64) (*OrderList, error) {
	p := newParams()
	p.Set("symbol", symbol)
	p.SetInt64("orderListId", orderListID)
	p.Set("listClientOrderId", listClientOrderID)
	var out OrderList
	if err := s.c.delete_(ctx, familySpot, "/api/v3/orderList", p, &out, signed(), recvWindow(recv)); err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenOrderLists returns open OCO orders.
func (s *SpotService) OpenOrderLists(ctx context.Context, recv int64) ([]OrderList, error) {
	var out []OrderList
	if err := s.c.get(ctx, familySpot, "/api/v3/openOrderList", newParams(), &out, signed(), weight(6), recvWindow(recv)); err != nil {
		return nil, err
	}
	return out, nil
}
