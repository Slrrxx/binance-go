package binance

import (
	"net/url"
	"strings"
)

var sensitiveQueryKeys = map[string]struct{}{
	"signature":       {},
	"apiSecret":       {},
	"secret":          {},
	"privateKey":      {},
	"withdrawOrderId": {},
}

// RedactQuery returns a query string safe for logs. Signature and other
// sensitive keys are replaced with [redacted].
func RedactQuery(raw string) string {
	if raw == "" {
		return ""
	}
	q, err := url.ParseQuery(raw)
	if err != nil {
		return "[unparseable-query]"
	}
	for key := range q {
		lk := strings.ToLower(key)
		if _, ok := sensitiveQueryKeys[key]; ok || strings.Contains(lk, "secret") || strings.Contains(lk, "signature") || strings.Contains(lk, "private") {
			q.Set(key, "[redacted]")
		}
	}
	return q.Encode()
}

func redactHeader(name, value string) string {
	switch strings.ToLower(name) {
	case "x-mbx-apikey", "authorization", "proxy-authorization":
		return "[redacted]"
	default:
		if looksLikeSecret(value) {
			return "[redacted]"
		}
		return value
	}
}

func looksLikeSecret(s string) bool {
	return s != "" && strings.Contains(s, "PRIVATE KEY")
}
