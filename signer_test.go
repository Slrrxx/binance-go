package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"testing"
)

func TestHMACSignatureKnownVector(t *testing.T) {
	// Official-style: HMAC_SHA256(secret, query) hex digest.
	secret := "secret"
	payload := "symbol=BTCUSDT&side=BUY&type=LIMIT&timestamp=1234567890"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	want := hex.EncodeToString(mac.Sum(nil))

	s := hmacSigner{secret: []byte(secret)}
	got, err := s.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestSignedQueryIncludesTimestampAndSignature(t *testing.T) {
	c := NewClient("key", "secret", WithRecvWindow(5000))
	c.SetTimeOffset(0)
	p := newParams()
	p.Set("symbol", "BTCUSDT")
	q, err := c.signedQuery(apiCall{params: p, sec: secSigned})
	if err != nil {
		t.Fatal(err)
	}
	if q.Get("timestamp") == "" {
		t.Fatal("missing timestamp")
	}
	if q.Get("recvWindow") != "5000" {
		t.Fatalf("recvWindow=%s", q.Get("recvWindow"))
	}
	sig := q.Get("signature")
	if sig == "" {
		t.Fatal("missing signature")
	}
	unsigned := url.Values{}
	for k, vs := range q {
		if k == "signature" {
			continue
		}
		for _, v := range vs {
			unsigned.Add(k, v)
		}
	}
	want, _ := hmacSigner{secret: []byte("secret")}.Sign(unsigned.Encode())
	if sig != want {
		t.Fatalf("signature mismatch\n got %s\nwant %s\npayload %s", sig, want, unsigned.Encode())
	}
}

func TestSignedQueryRequiresSecret(t *testing.T) {
	c := NewClient("key", "")
	_, err := c.signedQuery(apiCall{sec: secSigned, params: newParams()})
	if err == nil {
		t.Fatal("expected auth error")
	}
	var ae *AuthError
	if !errorsAs(err, &ae) {
		t.Fatalf("want AuthError, got %T %v", err, err)
	}
}

func TestRedactQuery(t *testing.T) {
	raw := "symbol=BTCUSDT&signature=supersecret&timestamp=1"
	out := RedactQuery(raw)
	if strings.Contains(out, "supersecret") {
		t.Fatalf("signature leaked: %s", out)
	}
	if !strings.Contains(out, "signature=%5Bredacted%5D") && !strings.Contains(out, "signature=[redacted]") {
		t.Fatalf("expected redacted signature, got %s", out)
	}
}

func TestClientStringHidesKey(t *testing.T) {
	c := NewClient("REALKEY", "REALSECRET")
	s := c.String()
	if strings.Contains(s, "REALKEY") || strings.Contains(s, "REALSECRET") {
		t.Fatalf("leaked credentials: %s", s)
	}
}

func TestNewClientOrderID(t *testing.T) {
	a, b := NewClientOrderID(), NewClientOrderID()
	if a == "" || a == b {
		t.Fatalf("ids not unique: %s %s", a, b)
	}
	if !strings.HasPrefix(a, "x-GOBNCE") {
		t.Fatalf("prefix: %s", a)
	}
}
