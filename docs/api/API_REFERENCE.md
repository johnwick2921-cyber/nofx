# VL API Reference

> This file replaced the crypto-era external coin-data API reference
> (coin-score / coin-data / netflow endpoints) as part of the crypto removal
> wave (2026-10). Those external crypto data endpoints were removed with the
> wave and no longer exist in the running binary.

The current system is futures-only: NinjaTrader 8 is the single data source and
execution venue (real-time bars + SIM execution over the TCP bridge).

For the operator-facing surface, see:

- The built-in guide (dashboard → Guide) — verified against code
- `docs/README-VL-SYSTEM.md` — operator's manual + full UI reference
- `docs/guides/TROUBLESHOOTING.md` — diagnostics
- `docs/architecture/README.md` — module overview

HTTP surfaces of the running service are exercised through the web UI
(`http://127.0.0.1:3000`, proxying `/api` to the Go backend). Key groups:

| Group | Purpose |
|-------|---------|
| `/api/auth/*` | JWT auth, sessions |
| `/api/traders/*` | trader CRUD, start/stop, decisions |
| `/api/strategy/*` | strategy config |
| `/api/exchanges/*`, `/api/models/*` | exchange + AI model config |
| `/api/positions`, `/api/balance`, `/api/trade-history` | live state |
| `/api/risk/gate-blocks` | risk-gate block counters |
