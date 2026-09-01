# Contributing

Issues and pull requests are welcome.

## Rules

- Keep the public API small and idiomatic Go.
- Document every exported symbol.
- Never log API secrets, signatures, or PEM material.
- Cover new request paths with `httptest` — do not hit live Binance in unit tests.
- Do not blindly retry order creates, withdrawals, or transfers.

## Dev loop

```bash
gofmt -w .
go vet ./...
go test -race ./...
```

Optional live public checks:

```bash
BINANCE_INTEGRATION=1 go test -tags=integration ./...
```
