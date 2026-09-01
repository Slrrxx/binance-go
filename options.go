package binance

import (
	"net/http"
	"time"
)

type config struct {
	apiKey       string
	apiSecret    string
	recvWindow   int64
	timeout      time.Duration
	headers      http.Header
	env          Environment
	tld          string
	retry        RetryConfig
	autoSync     bool
	userAgent    string
	httpClient   *http.Client
	signer       Signer
	limiter      RateLimiter
	log          Logger
	debug        bool
	wsMaxRetries int
	wsBackoffMin time.Duration
	wsBackoffMax time.Duration
	wsBuffer     int
	baseOverride *endpointSet
	proxyURL     string
	update       updateConfig
}

func defaultConfig() config {
	return config{
		recvWindow:   5000,
		timeout:      10 * time.Second,
		tld:          "com",
		userAgent:    userAgent,
		log:          nopLogger{},
		retry:        defaultRetryConfig(),
		wsMaxRetries: 0, // 0 = unlimited
		wsBackoffMin: time.Second,
		wsBackoffMax: 60 * time.Second,
		wsBuffer:     256,
		update:       defaultUpdateConfig(),
	}
}

// Option configures a Client.
type Option func(*config)

// WithHTTPClient injects a custom *http.Client (timeouts, proxy, transport).
func WithHTTPClient(c *http.Client) Option {
	return func(cfg *config) { cfg.httpClient = c }
}

// WithProxy sets an HTTP(S) proxy URL on the client's transport
// (for example "http://127.0.0.1:7890"). Applied after WithHTTPClient.
func WithProxy(raw string) Option {
	return func(cfg *config) { cfg.proxyURL = raw }
}

// WithTimeout sets a default per-request timeout when the caller context
// has no deadline.
func WithTimeout(d time.Duration) Option {
	return func(cfg *config) { cfg.timeout = d }
}

// WithBaseURL overrides the spot REST origin (tests, proxies, custom gateways).
func WithBaseURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.SpotREST = raw
		cfg.baseOverride.SAPI = raw
	}
}

// WithFuturesBaseURL overrides the USD-M futures REST origin.
func WithFuturesBaseURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.FAPI = raw
	}
}

// WithCoinFuturesBaseURL overrides the COIN-M futures REST origin.
func WithCoinFuturesBaseURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.DAPI = raw
	}
}

// WithWSAPIURL overrides the spot WebSocket API origin.
func WithWSAPIURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.SpotWSAPI = raw
	}
}

// WithFuturesWSAPIURL overrides the USD-M WebSocket API origin.
func WithFuturesWSAPIURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.FuturesWSAPI = raw
	}
}

// WithPortfolioURL overrides the portfolio-margin (papi) REST origin.
func WithPortfolioURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.PAPI = raw
	}
}

// WithOptionsURL overrides the vanilla options (eapi) REST origin.
func WithOptionsURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.EAPI = raw
	}
}

// WithWebSocketURL overrides the spot WebSocket base (`.../ws`).
func WithWebSocketURL(raw string) Option {
	return func(cfg *config) {
		if cfg.baseOverride == nil {
			e := endpointsFor(cfg.env, cfg.tld)
			cfg.baseOverride = &e
		}
		cfg.baseOverride.SpotWS = raw
	}
}

// WithTestnet points REST and WebSocket hosts at Binance testnet.
func WithTestnet() Option {
	return func(cfg *config) { cfg.env = EnvTestnet }
}

// WithDemo points REST hosts at Binance demo trading.
func WithDemo() Option {
	return func(cfg *config) { cfg.env = EnvDemo }
}

// WithTLD selects a regional website TLD such as "us" or "jp".
func WithTLD(tld string) Option {
	return func(cfg *config) { cfg.tld = tld }
}

// WithRecvWindow sets the default signed-request recvWindow in milliseconds.
func WithRecvWindow(ms int64) Option {
	return func(cfg *config) { cfg.recvWindow = ms }
}

// WithHeader adds a header sent on every REST request.
func WithHeader(key, value string) Option {
	return func(cfg *config) {
		if cfg.headers == nil {
			cfg.headers = make(http.Header)
		}
		cfg.headers.Set(key, value)
	}
}

// WithLogger installs an optional logger. The default is silent.
func WithLogger(l Logger) Option {
	return func(cfg *config) {
		if l != nil {
			cfg.log = l
		}
	}
}

// WithDebug enables verbose (redacted) request logging.
func WithDebug() Option {
	return func(cfg *config) { cfg.debug = true }
}

// WithRateLimiter installs a limiter invoked before each REST call.
func WithRateLimiter(l RateLimiter) Option {
	return func(cfg *config) { cfg.limiter = l }
}

// WithRetry overrides the transient-error retry policy.
func WithRetry(r RetryConfig) Option {
	return func(cfg *config) { cfg.retry = r }
}

// WithAutoTimeSync fetches server time before the first signed request
// and stores a clock offset.
func WithAutoTimeSync() Option {
	return func(cfg *config) { cfg.autoSync = true }
}

// WithRSAPrivateKey uses an RSA PEM private key instead of HMAC.
func WithRSAPrivateKey(pemBytes []byte, password []byte) Option {
	return func(cfg *config) {
		s, err := parsePrivateKeyPEM(pemBytes, password)
		if err != nil {
			cfg.signer = errSigner{err: err}
			return
		}
		if _, ok := s.(rsaSigner); !ok {
			cfg.signer = errSigner{err: &AuthError{Msg: "PEM is not an RSA private key"}}
			return
		}
		cfg.signer = s
	}
}

// WithEd25519PrivateKey uses an Ed25519 PEM (PKCS#8) private key.
func WithEd25519PrivateKey(pemBytes []byte, password []byte) Option {
	return func(cfg *config) {
		s, err := parsePrivateKeyPEM(pemBytes, password)
		if err != nil {
			cfg.signer = errSigner{err: err}
			return
		}
		if _, ok := s.(ed25519Signer); !ok {
			cfg.signer = errSigner{err: &AuthError{Msg: "PEM is not an Ed25519 private key"}}
			return
		}
		cfg.signer = s
	}
}

// WithSigner injects a custom signer.
func WithSigner(s Signer) Option {
	return func(cfg *config) { cfg.signer = s }
}

// WithWSReconnect configures websocket reconnect attempts. maxRetries <= 0
// means retry until the stream context is cancelled.
func WithWSReconnect(maxRetries int, minBackoff, maxBackoff time.Duration) Option {
	return func(cfg *config) {
		cfg.wsMaxRetries = maxRetries
		if minBackoff > 0 {
			cfg.wsBackoffMin = minBackoff
		}
		if maxBackoff > 0 {
			cfg.wsBackoffMax = maxBackoff
		}
	}
}

// WithWSBuffer sets the per-stream event channel size.
func WithWSBuffer(n int) Option {
	return func(cfg *config) {
		if n > 0 {
			cfg.wsBuffer = n
		}
	}
}

type errSigner struct{ err error }

func (e errSigner) Sign(string) (string, error) { return "", e.err }
