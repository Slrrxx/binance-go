# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-09-02

### Added

- Idiomatic Go client for Binance REST and WebSocket APIs
- HMAC-SHA256, RSA, and Ed25519 request signing
- Spot market, trading, account, and user-data stream
- USD-M and COIN-M futures
- Margin (cross / isolated) and wallet endpoints
- Sub-account helpers
- Portfolio margin (papi) and vanilla options (eapi)
- Simple Earn, convert, and gift card
- Historical kline pagination and streaming iterator
- Reconnecting market WebSocket streams with typed events
- `WebSocket()` facade (`Trade`, `Depth`, `UserData`, …)
- Spot and USD-M WebSocket API trading (`WSAPI`, `FuturesWSAPI`)
- Futures / COIN-M / portfolio / options user-data + futures depth cache
- Thread-safe depth cache (snapshot + incremental sync)
- Functional options: testnet, demo, proxy, custom HTTP client, logger, retry, rate limiter
- YAML-free endpoint generator (`go generate ./...` → `zz_generated.go`)
- Structured errors (`APIError`, rate-limit, auth, websocket)
- Secret redaction in logs and error URLs
- Unit tests via `httptest` (no live exchange in default CI)
- GitHub Actions: gofmt, generate, vet, test, race, golangci-lint, build
- Dependabot for Go modules and Actions
