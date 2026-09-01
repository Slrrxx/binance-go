package binance

import (
	"net/url"
	"strconv"
)

type params url.Values

func newParams() params {
	return params(make(url.Values))
}

func (p params) values() url.Values {
	if p == nil {
		return url.Values{}
	}
	return url.Values(p)
}

func (p params) Set(key, value string) {
	if value == "" {
		return
	}
	url.Values(p).Set(key, value)
}

func (p params) SetRaw(key, value string) {
	url.Values(p).Set(key, value)
}

func (p params) SetInt(key string, n int) {
	if n == 0 {
		return
	}
	url.Values(p).Set(key, strconv.Itoa(n))
}

func (p params) SetInt64(key string, n int64) {
	if n == 0 {
		return
	}
	url.Values(p).Set(key, strconv.FormatInt(n, 10))
}

func (p params) SetBool(key string, v bool) {
	if !v {
		return
	}
	url.Values(p).Set(key, "true")
}

func (p params) SetBoolPtr(key string, v *bool) {
	if v == nil {
		return
	}
	if *v {
		url.Values(p).Set(key, "true")
	} else {
		url.Values(p).Set(key, "false")
	}
}
