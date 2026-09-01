# Security

If you find a vulnerability in this client (secret leakage, signing bugs, unsafe defaults), open a **private** GitHub security advisory on this repository. Do not file a public issue for credential leaks.

## Using the SDK safely

- Store keys in environment variables or a secret manager — never in git.
- Restrict API keys by IP and disable withdrawal unless you need it.
- `Wallet().Withdraw` moves real funds and is not retried.
- Prefer `NewClientOrderID()` (or your own id) on live orders.
- Debug logging redacts signatures; do not wrap this client with a logger that dumps raw URLs.

This project is unofficial and unaffiliated with Binance.
