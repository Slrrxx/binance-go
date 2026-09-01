package binance

import "context"

// SpotService exposes Binance spot REST endpoints.
type SpotService struct{ c *Client }

// GetExchangeInfo returns spot exchange information.
func (c *Client) GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	return c.Spot().ExchangeInfo(ctx, ExchangeInfoRequest{})
}

// GetOrderBook returns the spot order book for symbol.
func (c *Client) GetOrderBook(ctx context.Context, symbol string, limit int) (*OrderBook, error) {
	return c.Spot().OrderBook(ctx, OrderBookRequest{Symbol: symbol, Limit: limit})
}

// GetSymbolTicker returns the latest price for a symbol.
func (c *Client) GetSymbolTicker(ctx context.Context, symbol string) (*SymbolPrice, error) {
	out, err := c.Spot().TickerPrice(ctx, TickerPriceRequest{Symbol: symbol})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, &APIError{Code: -1121, Message: "invalid symbol", HTTPStatus: 400}
	}
	return &out[0], nil
}

// GetAccount returns the current spot account.
func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	return c.Spot().Account(ctx, AccountRequest{})
}

// CreateOrder places a spot order. This call is not retried.
func (c *Client) CreateOrder(ctx context.Context, req OrderRequest) (*Order, error) {
	return c.Spot().CreateOrder(ctx, req)
}

// TestOrder validates a spot order without submitting it.
func (c *Client) TestOrder(ctx context.Context, req OrderRequest) error {
	return c.Spot().TestOrder(ctx, req)
}

// CancelOrder cancels a spot order.
func (c *Client) CancelOrder(ctx context.Context, req CancelOrderRequest) (*Order, error) {
	return c.Spot().CancelOrder(ctx, req)
}

// GetOrder queries a spot order.
func (c *Client) GetOrder(ctx context.Context, req QueryOrderRequest) (*Order, error) {
	return c.Spot().GetOrder(ctx, req)
}
