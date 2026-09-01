package binance

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Client is a thread-safe Binance REST and WebSocket client.
type Client struct {
	cfg       config
	offset    atomic.Int64
	synced    atomic.Bool
	usedWt    atomic.Int64
	endpoints endpointSet
}

// NewClient constructs a Client. apiSecret may be empty when an RSA or
// Ed25519 private key option is supplied. Public market-data methods work
// with empty credentials.
func NewClient(apiKey, apiSecret string, opts ...Option) *Client {
	cfg := defaultConfig()
	cfg.apiKey = apiKey
	cfg.apiSecret = apiSecret
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.httpClient == nil {
		cfg.httpClient = httpClientWithTimeout(cfg.timeout)
	}
	if cfg.signer == nil && apiSecret != "" {
		cfg.signer = hmacSigner{secret: []byte(apiSecret)}
	}
	if cfg.log == nil {
		cfg.log = nopLogger{}
	}
	ep := endpointsFor(cfg.env, cfg.tld)
	if cfg.baseOverride != nil {
		ov := *cfg.baseOverride
		if ov.SpotREST != "" {
			ep.SpotREST = ov.SpotREST
		}
		if ov.SAPI != "" {
			ep.SAPI = ov.SAPI
		}
		if ov.FAPI != "" {
			ep.FAPI = ov.FAPI
		}
		if ov.DAPI != "" {
			ep.DAPI = ov.DAPI
		}
		if ov.FAPIData != "" {
			ep.FAPIData = ov.FAPIData
		}
		if ov.DAPIData != "" {
			ep.DAPIData = ov.DAPIData
		}
		if ov.SpotWS != "" {
			ep.SpotWS = ov.SpotWS
		}
		if ov.SpotWSCombined != "" {
			ep.SpotWSCombined = ov.SpotWSCombined
		}
		if ov.FAPIWS != "" {
			ep.FAPIWS = ov.FAPIWS
		}
		if ov.DAPIWS != "" {
			ep.DAPIWS = ov.DAPIWS
		}
		if ov.PAPI != "" {
			ep.PAPI = ov.PAPI
		}
		if ov.EAPI != "" {
			ep.EAPI = ov.EAPI
		}
		if ov.SpotWSAPI != "" {
			ep.SpotWSAPI = ov.SpotWSAPI
		}
		if ov.FuturesWSAPI != "" {
			ep.FuturesWSAPI = ov.FuturesWSAPI
		}
		if ov.PAPIWS != "" {
			ep.PAPIWS = ov.PAPIWS
		}
		if ov.EAPIWS != "" {
			ep.EAPIWS = ov.EAPIWS
		}
	}
	if err := applyProxy(cfg.httpClient, cfg.proxyURL); err != nil {
		cfg.log.Error("proxy configuration failed", "err", err)
	}
	return &Client{cfg: cfg, endpoints: ep}
}

func httpClientWithTimeout(d time.Duration) *http.Client {
	if d <= 0 {
		d = 10 * time.Second
	}
	return &http.Client{Timeout: d}
}

// Spot returns the spot REST service.
func (c *Client) Spot() *SpotService { return &SpotService{c: c} }

// Wallet returns wallet / capital REST methods.
func (c *Client) Wallet() *WalletService { return &WalletService{c: c} }

// Margin returns cross and isolated margin methods.
func (c *Client) Margin() *MarginService { return &MarginService{c: c} }

// Futures returns the USDⓈ-M futures service.
func (c *Client) Futures() *FuturesService { return &FuturesService{c: c} }

// CoinFutures returns the COIN-M futures service.
func (c *Client) CoinFutures() *CoinFuturesService { return &CoinFuturesService{c: c} }

// SubAccount returns sub-account helpers.
func (c *Client) SubAccount() *SubAccountService { return &SubAccountService{c: c} }

// UserStream returns listenKey management for the spot user-data stream.
func (c *Client) UserStream() *UserStreamService { return &UserStreamService{c: c} }

// WebSocket returns market-stream helpers such as Trade and Depth.
func (c *Client) WebSocket() *WebSocketService { return &WebSocketService{c: c} }

// WSAPI returns the spot WebSocket API (order.place and related methods).
func (c *Client) WSAPI() *WSAPI { return &WSAPI{c: c} }

// Portfolio returns the portfolio-margin (papi) service.
func (c *Client) Portfolio() *PortfolioService { return &PortfolioService{c: c} }

// Options returns the vanilla options (eapi) service.
func (c *Client) Options() *OptionsService { return &OptionsService{c: c} }

// Earn returns Simple Earn helpers.
func (c *Client) Earn() *EarnService { return &EarnService{c: c} }

// Convert returns convert-quote helpers.
func (c *Client) Convert() *ConvertService { return &ConvertService{c: c} }

// GiftCard returns gift-card helpers.
func (c *Client) GiftCard() *GiftCardService { return &GiftCardService{c: c} }

// Environment reports the configured environment.
func (c *Client) Environment() Environment { return c.cfg.env }

// UsedWeight returns the last observed X-MBX-USED-WEIGHT-1M value.
func (c *Client) UsedWeight() int { return int(c.usedWt.Load()) }

// TimeOffset returns the server-local clock offset in milliseconds.
func (c *Client) TimeOffset() int64 { return c.offset.Load() }

// SetTimeOffset stores a millisecond offset added to request timestamps.
func (c *Client) SetTimeOffset(ms int64) { c.offset.Store(ms) }

func (c *Client) nowMillis() int64 {
	return time.Now().UnixMilli() + c.offset.Load()
}

func (c *Client) applyDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok || c.cfg.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, c.cfg.timeout)
}

// NewClientOrderID returns a unique client order id (32 hex chars).
func NewClientOrderID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("x-GOBNCE%x", time.Now().UnixNano())
	}
	return "x-GOBNCE" + hex.EncodeToString(b[:])[:16]
}

// String does not expose credentials.
func (c *Client) String() string {
	var buf bytes.Buffer
	buf.WriteString("binance.Client{env=")
	switch c.cfg.env {
	case EnvTestnet:
		buf.WriteString("testnet")
	case EnvDemo:
		buf.WriteString("demo")
	default:
		buf.WriteString("production")
	}
	if c.cfg.apiKey != "" {
		buf.WriteString(", apiKey=***")
	}
	buf.WriteByte('}')
	return buf.String()
}
