package binance

import (
	"fmt"
	"net/http"
	"net/url"
)

func applyProxy(client *http.Client, raw string) error {
	if client == nil || raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("binance: proxy url: %w", err)
	}
	switch t := client.Transport.(type) {
	case *http.Transport:
		tr := t.Clone()
		tr.Proxy = http.ProxyURL(u)
		client.Transport = tr
	case nil:
		base, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			base = &http.Transport{}
		}
		tr := base.Clone()
		tr.Proxy = http.ProxyURL(u)
		client.Transport = tr
	default:
		return fmt.Errorf("binance: cannot apply proxy to %T transport", client.Transport)
	}
	return nil
}
