package binance

import (
	"crypto"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// Signer produces a Binance request signature for a canonical payload.
type Signer interface {
	Sign(payload string) (string, error)
}

type hmacSigner struct {
	secret []byte
}

func (s hmacSigner) Sign(payload string) (string, error) {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

type rsaSigner struct {
	key *rsa.PrivateKey
}

func (s rsaSigner) Sign(payload string) (string, error) {
	sum := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

type ed25519Signer struct {
	key ed25519.PrivateKey
}

func (s ed25519Signer) Sign(payload string) (string, error) {
	sig := ed25519.Sign(s.key, []byte(payload))
	return base64.StdEncoding.EncodeToString(sig), nil
}

func parsePrivateKeyPEM(pemBytes, password []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, &AuthError{Msg: "invalid PEM private key"}
	}
	der := block.Bytes
	if x509.IsEncryptedPEMBlock(block) || strings.Contains(block.Type, "ENCRYPTED") { //nolint:staticcheck
		if len(password) == 0 {
			return nil, &AuthError{Msg: "encrypted private key requires a password"}
		}
		var err error
		der, err = x509.DecryptPEMBlock(block, password) //nolint:staticcheck
		if err != nil {
			return nil, fmt.Errorf("binance: decrypt private key: %w", err)
		}
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		return signerFromKey(key)
	}
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return rsaSigner{key: key}, nil
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return signerFromKey(key)
	}
	return nil, &AuthError{Msg: "unsupported private key type"}
}

func signerFromKey(key any) (Signer, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return rsaSigner{key: k}, nil
	case ed25519.PrivateKey:
		return ed25519Signer{key: k}, nil
	default:
		return nil, fmt.Errorf("%w: %T", errors.New("binance: unsupported private key"), key)
	}
}
