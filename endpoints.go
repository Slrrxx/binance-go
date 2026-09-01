package binance

// REST and WebSocket hosts. Keep all versioned paths in one file so Binance
// URL changes do not scatter across the client.
const (
	Version = "0.1.0"

	userAgent = "binance-go/" + Version

	headerAPIKey     = "X-MBX-APIKEY"
	headerUsedWeight = "X-MBX-USED-WEIGHT-1M"
	headerRetryAfter = "Retry-After"
)

type endpointSet struct {
	SpotREST       string
	SAPI           string
	FAPI           string
	DAPI           string
	FAPIData       string
	DAPIData       string
	PAPI           string
	EAPI           string
	SpotWS         string
	SpotWSCombined string
	FAPIWS         string
	FAPIWSCombined string
	DAPIWS         string
	DAPIWSCombined string
	PAPIWS         string
	EAPIWS         string
	SpotWSAPI      string
	FuturesWSAPI   string
}

func endpointsFor(env Environment, tld string) endpointSet {
	if tld == "" {
		tld = "com"
	}
	switch env {
	case EnvTestnet:
		return endpointSet{
			SpotREST:       "https://testnet.binance.vision",
			SAPI:           "https://testnet.binance.vision",
			FAPI:           "https://testnet.binancefuture.com",
			DAPI:           "https://testnet.binancefuture.com",
			FAPIData:       "https://testnet.binancefuture.com",
			DAPIData:       "https://testnet.binancefuture.com",
			PAPI:           "https://testnet.binancefuture.com",
			EAPI:           "https://testnet.binanceops.com",
			SpotWS:         "wss://stream.testnet.binance.vision/ws",
			SpotWSCombined: "wss://stream.testnet.binance.vision/stream",
			FAPIWS:         "wss://stream.binancefuture.com/ws",
			FAPIWSCombined: "wss://stream.binancefuture.com/stream",
			DAPIWS:         "wss://dstream.binancefuture.com/ws",
			DAPIWSCombined: "wss://dstream.binancefuture.com/stream",
			PAPIWS:         "wss://fstream.binancefuture.com/pm/ws",
			EAPIWS:         "wss://testnetws.binanceops.com/eoptions/ws",
			SpotWSAPI:      "wss://ws-api.testnet.binance.vision/ws-api/v3",
			FuturesWSAPI:   "wss://testnet.binancefuture.com/ws-fapi/v1",
		}
	case EnvDemo:
		return endpointSet{
			SpotREST:       "https://demo-api.binance.com",
			SAPI:           "https://demo-api.binance.com",
			FAPI:           "https://demo-fapi.binance.com",
			DAPI:           "https://demo-dapi.binance.com",
			FAPIData:       "https://demo-fapi.binance.com",
			DAPIData:       "https://demo-dapi.binance.com",
			PAPI:           "https://papi.binance.com",
			EAPI:           "https://eapi.binance.com",
			SpotWS:         "wss://stream.binance.com:9443/ws",
			SpotWSCombined: "wss://stream.binance.com:9443/stream",
			FAPIWS:         "wss://fstream.binance.com/ws",
			FAPIWSCombined: "wss://fstream.binance.com/stream",
			DAPIWS:         "wss://dstream.binance.com/ws",
			DAPIWSCombined: "wss://dstream.binance.com/stream",
			PAPIWS:         "wss://fstream.binance.com/pm/ws",
			EAPIWS:         "wss://nbstream.binance.com/eoptions/ws",
			SpotWSAPI:      "wss://demo-ws-api.binance.com/ws-api/v3",
			FuturesWSAPI:   "wss://testnet.binancefuture.com/ws-fapi/v1",
		}
	default:
		return endpointSet{
			SpotREST:       "https://api.binance." + tld,
			SAPI:           "https://api.binance." + tld,
			FAPI:           "https://fapi.binance." + tld,
			DAPI:           "https://dapi.binance." + tld,
			FAPIData:       "https://fapi.binance." + tld,
			DAPIData:       "https://dapi.binance." + tld,
			PAPI:           "https://papi.binance." + tld,
			EAPI:           "https://eapi.binance." + tld,
			SpotWS:         "wss://stream.binance." + tld + ":9443/ws",
			SpotWSCombined: "wss://stream.binance." + tld + ":9443/stream",
			FAPIWS:         "wss://fstream.binance." + tld + "/ws",
			FAPIWSCombined: "wss://fstream.binance." + tld + "/stream",
			DAPIWS:         "wss://dstream.binance." + tld + "/ws",
			DAPIWSCombined: "wss://dstream.binance." + tld + "/stream",
			PAPIWS:         "wss://fstream.binance." + tld + "/pm/ws",
			EAPIWS:         "wss://nbstream.binance." + tld + "/eoptions/ws",
			SpotWSAPI:      "wss://ws-api.binance." + tld + ":443/ws-api/v3",
			FuturesWSAPI:   "wss://ws-fapi.binance." + tld + "/ws-fapi/v1",
		}
	}
}

type family int

const (
	familySpot family = iota
	familySAPI
	familyFAPI
	familyDAPI
	familyFAPIData
	familyDAPIData
	familyPAPI
	familyEAPI
)

func (e endpointSet) base(f family) string {
	switch f {
	case familySAPI:
		return e.SAPI
	case familyFAPI:
		return e.FAPI
	case familyDAPI:
		return e.DAPI
	case familyFAPIData:
		return e.FAPIData
	case familyDAPIData:
		return e.DAPIData
	case familyPAPI:
		return e.PAPI
	case familyEAPI:
		return e.EAPI
	default:
		return e.SpotREST
	}
}

func (e endpointSet) ws(market Market, combined bool) string {
	switch market {
	case MarketUSDFutures:
		if combined {
			return e.FAPIWSCombined
		}
		return e.FAPIWS
	case MarketCoinFutures:
		if combined {
			return e.DAPIWSCombined
		}
		return e.DAPIWS
	case MarketPortfolio:
		return e.PAPIWS
	case MarketOptions:
		return e.EAPIWS
	default:
		if combined {
			return e.SpotWSCombined
		}
		return e.SpotWS
	}
}
