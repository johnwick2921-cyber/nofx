# Crypto-removal disposition table — CR-B (DS-107)

- Branch point: `dc630ad8f` (origin/dev tip the branch was cut from)
- Integrator tip: `f195e40eb4c7e8e5899df8397db983b2c7b3d745` (claim head the sweep ran against)
- Generated: 2026-10-01T05:41:58.489339+00:00
- Sweep regex: extracted PROGRAMMATICALLY from plan v10 line 108 (the C13 "THE assembled regex" bullet; GO item 2 = NO, so the BASE literal without alpaca/twelvedata/sina). Never hand-typed; when DS-102's guard lands, the gate asserts this table's regex byte-for-byte against it.
- Swept: 639 tracked non-test files under CR-B paths (git ls-files -z, grep -I semantics); hits in 90 files.
- CANONICAL FORMAT (CTO ruling 2026-10-01 00:40 CDT): one row per hit line, six columns path|line|token|disposition|owner|reason — machine-parseable.
- EXACTLY ONE OWNER PER HIT LINE (blocker 1): CR-A-owned lines in the shared files agent/tools.go + agent/agent.go are ceded rows (owner CR-A, their disposition, parsed from origin/crypto-removal-cr-a:docs/crypto-removal/disposition-CR-A.md at generation time); everything else here is CR-B.
- Dispositions are the first-pass proposal; the CTO + checkers review BEFORE any DELETE/CUT code. KEEP rows are one per file+token; DELETE/CUT rows are per line.
- Overlap note: `telegram/` and residual `api/` files are carried here pending CR-A's table; the union gate reconciles.

| path | line | token | disposition | owner | reason |
|---|---|---|---|---|---|
| `.github/ISSUE_TEMPLATE/bug_report.md` | 105 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/ISSUE_TEMPLATE/bug_report.md` | 105 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/ISSUE_TEMPLATE/bug_report.md` | 105 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/ISSUE_TEMPLATE/bug_report.md` | 155 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/SECURITY.md` | 69 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/SECURITY.md` | 69 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/SECURITY.md` | 72 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `.github/SECURITY.md` | 72 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `.github/SECURITY.md` | 73 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `.github/SECURITY.md` | 73 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/SECURITY.md` | 74 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `.github/labeler.yml` | 32 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/labeler.yml` | 33 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/labeler.yml` | 34 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/labeler.yml` | 35 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `.github/labeler.yml` | 36 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 45 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 45 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 90 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 91 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 91 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 92 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 103 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 103 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 110 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 119 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 121 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `CHANGELOG.zh-CN.md` | 151 | `币安` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `CHANGELOG.zh-CN.md` | 170 | `币安` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `README.ja.md` | 47 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 47 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 47 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 49 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 60 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 68 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 69 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 72 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 74 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 76 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 79 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 79 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 86 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 92 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 92 | `asterdex` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 92 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.ja.md` | 95 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 105 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 105 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 105 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 106 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 106 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 106 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 107 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 107 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 107 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 108 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 108 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 108 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 109 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 109 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 109 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 115 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 115 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 115 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 116 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 116 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 116 | `asterdex` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 117 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 117 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 117 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 160 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 170 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 170 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 170 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 180 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 180 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 180 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 186 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 219 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 219 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 234 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 251 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 259 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 259 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 271 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 273 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 275 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 275 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 279 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 283 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 283 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 465 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 465 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 466 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 466 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 480 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 488 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 488 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 488 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 489 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 489 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 494 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 496 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 500 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 501 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 509 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 511 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 511 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 511 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 513 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 518 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 518 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 520 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 526 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 527 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 530 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 531 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 532 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 532 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.ja.md` | 533 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 544 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 545 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 545 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 545 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 546 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 547 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 553 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 555 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 558 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 564 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 566 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 566 | `asterdex` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 566 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.ja.md` | 567 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.ja.md` | 581 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 584 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 605 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 615 | `asterdex` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 615 | `asterdex` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 615 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.ja.md` | 630 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 630 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 631 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 631 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 642 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 642 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 643 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 643 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 653 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 659 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 661 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 672 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 672 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 672 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 673 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 673 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 673 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 674 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 674 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 674 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 675 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 675 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 675 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 676 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 676 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.ja.md` | 676 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 676 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 677 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 688 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 700 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 711 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 713 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 768 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 775 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 782 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 821 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 930 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 934 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 964 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 998 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 999 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1018 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1038 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1039 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1039 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1047 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1067 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1068 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1069 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1070 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1071 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1074 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1075 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1080 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1161 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1190 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1218 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1220 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1221 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1235 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1235 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1237 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1239 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1240 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1240 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1242 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1253 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1253 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1254 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1254 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1263 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1297 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1297 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1342 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1342 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.ja.md` | 1342 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 5 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 14 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 14 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 14 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 14 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 15 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 15 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 15 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 32 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 32 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 34 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 34 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 34 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 34 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 34 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 52 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 56 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 59 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 59 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 62 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 64 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 67 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 67 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 67 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 75 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 88 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 88 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 88 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 89 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 89 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 89 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 90 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 90 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 90 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 91 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 91 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 91 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 92 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 92 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 92 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 98 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 98 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 98 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 99 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 99 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 99 | `asterdex` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 100 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 100 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 100 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 114 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 116 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 116 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 116 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 116 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 185 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 185 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `README.md` | 221 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 222 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 227 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 227 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 227 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 227 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 227 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 228 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 228 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `README.md` | 228 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 143 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 144 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 171 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 177 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 178 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 178 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 220 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 220 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 378 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 379 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 406 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 412 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 413 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `SECURITY.md` | 413 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 29 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 64 | `Wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 64 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 65 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 65 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 65 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 71 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 71 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 71 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 169 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 169 | `usdc` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 169 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 177 | `claw402` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 178 | `x402` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 226 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 236 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 238 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 249 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 249 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 250 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 250 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 277 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 278 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 285 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 285 | `Wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 286 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 289 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 289 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 296 | `USDC` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 298 | `claw402` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 298 | `blockrun` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 326 | `wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 362 | `claw402` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 362 | `claw402` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 459 | `Wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 508 | `Wallet` | CUT | CR-A | ceded to CR-A — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays |
| `agent/agent.go` | 745 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 758 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 761 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 762 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 781 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 907 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 909 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/agent.go` | 985 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/brain.go` | 209 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/brain.go` | 209 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/brain.go` | 210 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/brain.go` | 224 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/brain.go` | 228 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 94 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 94 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 94 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 94 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 94 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 94 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 483 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 856 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 856 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/central_brain.go` | 908 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 26 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 26 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 26 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 75 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 75 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 79 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 101 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 101 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 105 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 156 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 156 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 156 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 156 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 160 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 160 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 204 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 204 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 204 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 223 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 223 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 224 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 224 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 224 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 224 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 235 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 236 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 236 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 248 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 248 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 248 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 248 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 252 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 252 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 281 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 282 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 283 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 283 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 284 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 285 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 286 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 287 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 287 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/config_validation.go` | 288 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 333 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 334 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 448 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/config_validation.go` | 450 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 37 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 37 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/entity_field_catalog.go` | 37 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 37 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/entity_field_catalog.go` | 37 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 37 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/entity_field_catalog.go` | 38 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 39 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 40 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 41 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/entity_field_catalog.go` | 41 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 41 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/entity_field_catalog.go` | 41 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/entity_field_catalog.go` | 42 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 42 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/entity_field_catalog.go` | 43 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 224 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 224 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/llm_flow_extractor.go` | 224 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 224 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/llm_flow_extractor.go` | 228 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/llm_flow_extractor.go` | 228 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/llm_flow_extractor.go` | 249 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 251 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 251 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 251 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 426 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 427 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 427 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 427 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/llm_flow_extractor.go` | 428 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 429 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 429 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/llm_flow_extractor.go` | 429 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/llm_flow_extractor.go` | 430 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 16 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 32 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 33 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 33 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 36 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 39 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 44 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 45 | `BlockRun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 45 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 48 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 51 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 54 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 55 | `BlockRun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 55 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 58 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 61 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 145 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 145 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 145 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 145 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 145 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 145 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 147 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 147 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 147 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 174 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 175 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 176 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 198 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 199 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 199 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 200 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 200 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 213 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 214 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 214 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 214 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 214 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 214 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 215 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 216 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 217 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 218 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 224 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 225 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 225 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 225 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 225 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 225 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 225 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 225 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 225 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 226 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 227 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 227 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 227 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 228 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 229 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_provider_catalog.go` | 229 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_provider_catalog.go` | 229 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 9 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 14 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 14 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 15 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 18 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 19 | `usdc` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 20 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 20 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 26 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 27 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 33 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 35 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 35 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 40 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 47 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 49 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 49 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 53 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 56 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 56 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 57 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 58 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 60 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 62 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 62 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 64 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 64 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 68 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 68 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 68 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 72 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 72 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 75 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 75 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 76 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 77 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 77 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 79 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 80 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/model_wallet_fastpath.go` | 80 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 80 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 83 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/model_wallet_fastpath.go` | 83 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/onboard.go` | 129 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 129 | `币安` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `agent/onboard.go` | 130 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 130 | `欧易` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `agent/onboard.go` | 131 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 132 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 134 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 135 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 138 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 139 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 140 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 141 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 143 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/onboard.go` | 144 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 157 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/planner_runtime.go` | 2670 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/planner_runtime.go` | 3767 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/planner_runtime.go` | 3769 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/sentinel.go` | 148 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_catalog.go` | 32 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_catalog.go` | 34 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_catalog.go` | 51 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_catalog.go` | 56 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_catalog.go` | 88 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_catalog.go` | 91 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 255 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_dispatcher.go` | 256 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 263 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_dispatcher.go` | 263 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_dispatcher.go` | 264 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_dispatcher.go` | 267 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 267 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_dispatcher.go` | 272 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 274 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 294 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_dispatcher.go` | 782 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 782 | `欧易` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `agent/skill_dispatcher.go` | 782 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 783 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_dispatcher.go` | 795 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 30 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 30 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 30 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 31 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 31 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 34 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 35 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 35 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 44 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 44 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 44 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_domain_context.go` | 44 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 45 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 45 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 46 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_domain_context.go` | 48 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 48 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_domain_context.go` | 49 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 49 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 49 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_domain_context.go` | 114 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 124 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 148 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 148 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 148 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 150 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 172 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 172 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 172 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 174 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 182 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 182 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 182 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 182 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 182 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_domain_context.go` | 182 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 21 | `usdt` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 21 | `usdc` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 53 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 155 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 155 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 157 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 159 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 159 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 160 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 162 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 164 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 167 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 169 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 172 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 174 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 177 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 179 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 180 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 182 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 184 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 184 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 187 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 189 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 192 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 194 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 197 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 199 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 368 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 368 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 372 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 379 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 379 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 380 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 403 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 403 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 404 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 404 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 404 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 404 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 415 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 416 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 416 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 445 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 445 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 446 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 446 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 446 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 446 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 457 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 458 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 458 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 483 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 483 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 483 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 483 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 487 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 487 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1544 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 1544 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1544 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 1544 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1545 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 1545 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1556 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1556 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1557 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1592 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1602 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 1602 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 1606 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_execution_handlers.go` | 2178 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2180 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2506 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2510 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2515 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2515 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2577 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_execution_handlers.go` | 2585 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 174 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 174 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 174 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 174 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 174 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 765 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 765 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 767 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 767 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 783 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 787 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 805 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 805 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 805 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 805 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 807 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 807 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 825 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 825 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 825 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 825 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 825 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 825 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 829 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 859 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 859 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 869 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 869 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 869 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 869 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 869 | `永续` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `agent/skill_management_handlers.go` | 869 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 869 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 885 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 885 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 885 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 885 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 885 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 885 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 889 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 889 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1080 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1115 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1133 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1273 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1273 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1273 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1278 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1278 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1278 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1284 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1287 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1287 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1287 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1289 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1291 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1292 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1293 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1295 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1297 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1297 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1298 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1301 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1320 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1320 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1320 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1325 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1325 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1325 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1331 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1334 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1334 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1334 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1334 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1336 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1338 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1339 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1340 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1342 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1344 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1344 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1344 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1345 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1348 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1735 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1735 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1735 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1738 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1738 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1738 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1742 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1744 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1744 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1744 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1745 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1745 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1745 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1745 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1746 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1748 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1749 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1750 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1753 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1754 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1755 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1757 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1759 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1759 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1760 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1761 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1764 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1764 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1764 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1765 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1766 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1821 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1822 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1824 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1825 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1825 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1837 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1838 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1838 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1840 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1841 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1841 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1841 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1976 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1980 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1981 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1981 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1982 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1982 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 1991 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1991 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 1991 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2003 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2003 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 2003 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2003 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 2007 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 2007 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 2026 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2026 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 2026 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skill_management_handlers.go` | 2255 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2262 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2291 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_management_handlers.go` | 2298 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 103 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 142 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 143 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 144 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 145 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 146 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 147 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 150 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 151 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 152 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 153 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 154 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 155 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 156 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 157 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 158 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 159 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 160 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skill_semantic_gate.go` | 161 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 8 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 8 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 8 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 8 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 8 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 8 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 13 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 13 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 13 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 15 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 15 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_diagnosis.json` | 15 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 10 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `币安` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `agent/skills/exchange_management.json` | 11 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `欧易` | CUT | CR-B | Chinese crypto token — reworded away, nothing stays |
| `agent/skills/exchange_management.json` | 11 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `gate.io` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 11 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 31 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 31 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 31 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 32 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 32 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 32 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 32 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 32 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 32 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 44 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 44 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 46 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 47 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 47 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 49 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 52 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 53 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 57 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 58 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 58 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 62 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 63 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 63 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 67 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 68 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 68 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 70 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 72 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 73 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 73 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 77 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 78 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 78 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 82 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 83 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 83 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 89 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 90 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 96 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 97 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 97 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 98 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 98 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 98 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 99 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 99 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 99 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 100 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 100 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 101 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 102 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 102 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 107 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 108 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 109 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 110 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 112 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 113 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 114 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 114 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 114 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 115 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 116 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 116 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 122 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 122 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 122 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 122 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 126 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 126 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 127 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 127 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 127 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 138 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 138 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/exchange_management.json` | 138 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/exchange_management.json` | 138 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/model_diagnosis.json` | 5 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_diagnosis.json` | 8 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_diagnosis.json` | 8 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_diagnosis.json` | 8 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_diagnosis.json` | 13 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_diagnosis.json` | 13 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_diagnosis.json` | 13 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 10 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 10 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 10 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 20 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 20 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 25 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 25 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 38 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 38 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 38 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 43 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 43 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 43 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 44 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 44 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 45 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 45 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 45 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 58 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 59 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 59 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 59 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 60 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 60 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 76 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 76 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 89 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 89 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 130 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 130 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 130 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/model_management.json` | 141 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_diagnosis.json` | 13 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 40 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 40 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 40 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 40 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 40 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 40 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 41 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 41 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 41 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 41 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 41 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 41 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 45 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 45 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 45 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 46 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 46 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 46 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 46 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 51 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 51 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/strategy_management.json` | 121 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trade_execution.json` | 23 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trade_execution.json` | 24 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trade_execution.json` | 28 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_diagnosis.json` | 26 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_diagnosis.json` | 26 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_diagnosis.json` | 26 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 65 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 65 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 67 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 67 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 67 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 67 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 69 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 69 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 69 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/trader_management.json` | 69 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 69 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 69 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/skills/trader_management.json` | 81 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 114 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 114 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 146 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 146 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 158 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/skills/trader_management.json` | 158 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 24 | `aster` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 25 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 26 | `bitget` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 27 | `bybit` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 29 | `hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 29 | `hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 30 | `indodax` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 31 | `kucoin` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 32 | `lighter` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 34 | `okx` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 41 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 41 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 64 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 76 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 79 | `usdt` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 284 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 284 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 284 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 284 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 285 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 285 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 345 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 380 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 380 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 380 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 392 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 392 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 392 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 396 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 396 | `blockrun` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 413 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 425 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 425 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 425 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 427 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 427 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 427 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 427 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 428 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 428 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 429 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 430 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 431 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 432 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 432 | `LIGHTER` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 432 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 433 | `LIGHTER` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 434 | `LIGHTER` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 435 | `LIGHTER` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 534 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 550 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 550 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 550 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 550 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 551 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 551 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 555 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 555 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 695 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 695 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 744 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 761 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 761 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 786 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 786 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 821 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 860 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 860 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 933 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 933 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 933 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 933 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 937 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 937 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 953 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 953 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 954 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 954 | `usdc` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1037 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1037 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1037 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1037 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1041 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1041 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1065 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1066 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1067 | `bybit` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1068 | `bybit` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1068 | `Bybit` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1069 | `okx` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1070 | `okx` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1070 | `OKX` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1071 | `bitget` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1072 | `bitget` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1072 | `Bitget` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1075 | `kucoin` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1076 | `kucoin` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1076 | `KuCoin` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1077 | `indodax` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1078 | `indodax` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1078 | `Indodax` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1079 | `hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1080 | `hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1080 | `Hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1082 | `Hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1082 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1084 | `Hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1086 | `aster` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1087 | `aster` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1092 | `lighter` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1093 | `lighter` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1094 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1110 | `Wallet` | KEEP | CR-A | ceded to CR-A — NT8 account-snapshot shape — same totalWalletBalance key family, the agent balance extractor (CTO ruling) |
| `agent/tools.go` | 1110 | `wallet` | KEEP | CR-A | ceded to CR-A — NT8 account-snapshot shape — same totalWalletBalance key family, the agent balance extractor (CTO ruling) |
| `agent/tools.go` | 1146 | `USDC` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1149 | `wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1149 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1149 | `wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1150 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1150 | `wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1151 | `USDC` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1151 | `wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1152 | `USDC` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1472 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1472 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1472 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1472 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1473 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1473 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1477 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1477 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1505 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1506 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1518 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1518 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1518 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1518 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1522 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1522 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1540 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1540 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1545 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1599 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1600 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1601 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1607 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1607 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/tools.go` | 1607 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `agent/tools.go` | 1608 | `Hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1608 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1609 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1609 | `Hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1609 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1619 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1619 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1620 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1621 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1621 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1653 | `hyperliquid` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1653 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1653 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1657 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1657 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1672 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 1677 | `Wallet` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 2390 | `AI500` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 2413 | `AI500` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 2413 | `AI500` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 2485 | `AI500` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 2485 | `AI500` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3013 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3014 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3050 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3051 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3081 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3093 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3102 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3107 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3248 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3250 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3308 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3309 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3328 | `binance` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3610 | `hyper_all` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3611 | `hyper_main` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3612 | `hyper_main` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3643 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3643 | `USDC` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3645 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3727 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3734 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3735 | `USDC` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3756 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3765 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3766 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3767 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3773 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3791 | `USDT` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/tools.go` | 3791 | `USDC` | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays |
| `agent/trade.go` | 21 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 22 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 60 | `USDT` | CUT | CR-A | ceded to CR-A — crypto symbol example/comment and the USDT-suffix stripping go; the chat-entry admission chain stays |
| `agent/trade.go` | 327 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 328 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 328 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 346 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 346 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 357 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 362 | `USDT` | CUT | CR-A | ceded to CR-A — crypto symbol example/comment and the USDT-suffix stripping go; the chat-entry admission chain stays |
| `agent/trade.go` | 400 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 400 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 413 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 413 | `USDT` | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency |
| `agent/trade.go` | 486 | `USDT` | CUT | CR-A | ceded to CR-A — crypto symbol example/comment and the USDT-suffix stripping go; the chat-entry admission chain stays |
| `agent/trade.go` | 487 | `USDT` | CUT | CR-A | ceded to CR-A — crypto symbol example/comment and the USDT-suffix stripping go; the chat-entry admission chain stays |
| `agent/web.go` | 50 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 56 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 56 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 58 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 208 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 212 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 228 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 228 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 231 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 235 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 243 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 243 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 246 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 246 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 246 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 251 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 251 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 251 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 283 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 288 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 341 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `agent/web.go` | 347 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `api/handler_competition.go` | 154 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `api/handler_competition.go` | 418 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `api/handler_competition.go` | 419 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `api/handler_competition.go` | 420 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `api/handler_competition.go` | 432 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `api/handler_order.go` | 108 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `api/handler_order.go` | 108 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `api/handler_order.go` | 109 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `api/handler_order.go` | 315 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `api/handler_order.go` | 371 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `api/handler_plan_order_truth.go` | 230 | `"mixed"` | KEEP | CR-B | plan-state string (plan C13) |
| `clock-seams.list` | 35 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `clock-seams.list` | 36 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 60 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 60 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 62 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 66 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 68 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 74 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 116 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 117 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 126 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 211 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 211 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/README.md` | 270 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/hooks.go` | 37 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/hooks.go` | 37 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/hooks.go` | 37 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/trader_hook.go` | 7 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/trader_hook.go` | 10 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/trader_hook.go` | 15 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/trader_hook.go` | 17 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `hook/trader_hook.go` | 22 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 15 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 55 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 55 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 331 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 331 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `kernel/engine.go` | 332 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 332 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `kernel/engine.go` | 334 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 334 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `kernel/engine.go` | 398 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 399 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 401 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 401 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 417 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 418 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 420 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 420 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 436 | `"mixed"` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 440 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 443 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 451 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 454 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 507 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 510 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 512 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 517 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 518 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 521 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 524 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 524 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 528 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 535 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 537 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 542 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 543 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 546 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 549 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 549 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine.go` | 554 | `Quant` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_analysis.go` | 835 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_analysis.go` | 837 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 73 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 73 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 94 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 94 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 96 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 96 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 100 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 100 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 109 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 109 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 110 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_position.go` | 112 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 81 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 83 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 86 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 100 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 156 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 262 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 288 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 309 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 320 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 409 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 471 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 471 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 492 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 494 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 499 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 506 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 507 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 508 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 509 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 645 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt.go` | 655 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/engine_prompt_futures.go` | 89 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 108 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 109 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 150 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 151 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 202 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 233 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 235 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 238 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 343 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 344 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 386 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 387 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 437 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 467 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 469 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/formatter.go` | 472 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/grid_engine.go` | 116 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/grid_engine.go` | 150 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/grid_engine.go` | 151 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/grid_engine.go` | 168 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/grid_engine.go` | 205 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/grid_engine.go` | 206 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 85 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 108 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 158 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 164 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 171 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 220 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 243 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 293 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 299 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 306 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/prompt_builder.go` | 319 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/risk_limits.go` | 13 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 67 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 76 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 106 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 113 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 120 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 180 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 189 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 206 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/schema.go` | 213 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/testdata/golden/user_prompt_crypto_change_na.txt` | 8 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/testdata/golden/user_prompt_crypto_change_na.txt` | 8 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/testdata/golden/user_prompt_crypto_change_na.txt` | 8 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `kernel/testdata/golden/user_prompt_crypto_change_na.txt` | 10 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `main.go` | 247 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `main.go` | 253 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `main.go` | 253 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `main.go` | 372 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 569 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 639 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 645 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 645 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 645 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 647 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 648 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 649 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 650 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 680 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 681 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 682 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 683 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 684 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 685 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 686 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 687 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 688 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 689 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 690 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 691 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 692 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 693 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 697 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 698 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 699 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 700 | `KuCoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 701 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 702 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 703 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 703 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 703 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 703 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 704 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 704 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 705 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 709 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 711 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 711 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 715 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 716 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 717 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 741 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 741 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 741 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 779 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 780 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 781 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 782 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 782 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 783 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 791 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 792 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 795 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 795 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 795 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 797 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `manager/trader_manager.go` | 797 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `manager/trader_manager.go` | 800 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `market/api_client.go` | 15 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 16 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 28 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 30 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 36 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 40 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 40 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 40 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 47 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 47 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 88 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 90 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 100 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 103 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 105 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 113 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 113 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 114 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 114 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 125 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 126 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 127 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 129 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 132 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 133 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 135 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 139 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 151 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 152 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 154 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 157 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 159 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 178 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 178 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 184 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 185 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 258 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 258 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 310 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 313 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 334 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 335 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 337 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 341 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 341 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 342 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 342 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 344 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 411 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 412 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 413 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 448 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 493 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 693 | `ASTER` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 716 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 737 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 737 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 747 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 748 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 749 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 752 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 764 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 770 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 770 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 779 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 779 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 779 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 785 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 788 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data.go` | 808 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 7 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 7 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 8 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 8 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 9 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 15 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 17 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 17 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 18 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 19 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 20 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 20 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 23 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 23 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 25 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 25 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 27 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 27 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 29 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 29 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 31 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 31 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 33 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 33 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 35 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 35 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 37 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 37 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 39 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 39 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 41 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 41 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 43 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 43 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 45 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 45 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 47 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 47 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 49 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 49 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 54 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 55 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 55 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 57 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 58 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 58 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 58 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 59 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 60 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 60 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 60 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 61 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 62 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 62 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 63 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 64 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 64 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 64 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 66 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 66 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 67 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 68 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 68 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 68 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 69 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 70 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 70 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 70 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 72 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 73 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 73 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 73 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 76 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 80 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 80 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 80 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 80 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 80 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 81 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 82 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 83 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 83 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 83 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 85 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 85 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 87 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 87 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 89 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 89 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 89 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 89 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 89 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 89 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 91 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 94 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 98 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 99 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 100 | `coinank` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 115 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 115 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 116 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 120 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 121 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 123 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 124 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 130 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 411 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 413 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/data_klines.go` | 413 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/futures_data.go` | 5 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/futures_symbol.go` | 5 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/futures_symbol.go` | 38 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 12 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 12 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 13 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 36 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 44 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 60 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/historical.go` | 98 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/types.go` | 38 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `market/types.go` | 127 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `ninjascript/VLContractResolver.cs` | 66 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `ninjascript/VLContractResolver_VERIFY.md` | 20 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `ninjascript/VLContractResolver_VERIFY.md` | 20 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `ninjascript/VLContractResolver_VERIFY.md` | 123 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `ninjascript/VLContractResolver_VERIFY.md` | 123 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `patches/partner-2026-08-20/MANIFEST.md` | 458 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `provider/ninjatrader/bar_source.go` | 18 | `"mixed"` | KEEP | CR-B | BarSourceMixed = "mixed" — futures bar-source enum (plan C13) |
| `provider/ninjatrader/tcp_framing.go` | 675 | `"mixed"` | KEEP | CR-B | comment describes BarSourceMixed (plan C13) |
| `scripts/replay_write_time_feasibility.py` | 162 | `"mixed"` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 152 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 152 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 153 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 154 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 157 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 158 | `usdc` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 165 | `usdc` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_charge.go` | 166 | `usdc` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 64 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 64 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 65 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 66 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 68 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 70 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 431 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 431 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 431 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 431 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 432 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 433 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 434 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 436 | `Claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 436 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 442 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 443 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 444 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 445 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 445 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 447 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 456 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 456 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/ai_model.go` | 461 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 471 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/ai_model.go` | 471 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/bar_history.go` | 151 | `"mixed"` | KEEP | CR-B | NT8 live+historical census value + boot-line mixed=%d (plan C13) |
| `store/bar_history_across_roll.go` | 82 | `"mixed"` | KEEP | CR-B | comment describes the roll-straddling ("mixed") state (plan C13) |
| `store/exchange.go` | 32 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 32 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 32 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 32 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 32 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 32 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 33 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 33 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 33 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 37 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 37 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 37 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 101 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 101 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 105 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 131 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 131 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 131 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 131 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 131 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 131 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 131 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 145 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 155 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 210 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 211 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 212 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 213 | `Bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 214 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 215 | `OKX` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 216 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 217 | `Bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 218 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 219 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 220 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 221 | `Aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 222 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 223 | `LIGHTER` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 224 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 225 | `Indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 238 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 238 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 238 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 240 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 245 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 245 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 246 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 274 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 274 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 274 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 274 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 275 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 275 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 279 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 279 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 298 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 298 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 298 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 299 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 307 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 307 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 307 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 307 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 308 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 308 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 311 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 311 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 388 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 388 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 391 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 391 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 391 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 391 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 391 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 391 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 391 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 393 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 393 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 409 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 409 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/exchange.go` | 409 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/exchange.go` | 409 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/indicator_fingerprint.go` | 12 | `quant` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/knob_registry_table.go` | 73 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/knob_registry_table.go` | 73 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/knob_registry_table.go` | 169 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/knob_registry_table.go` | 169 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/knob_registry_table.go` | 170 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/knob_registry_table.go` | 170 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/order.go` | 33 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/order.go` | 390 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position.go` | 746 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position.go` | 747 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position.go` | 748 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position.go` | 795 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position.go` | 796 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position.go` | 797 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/position_builder.go` | 146 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 25 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 218 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 225 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 225 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 225 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 261 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 262 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 264 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 266 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 514 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1828 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1832 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1874 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1874 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1874 | `"mixed"` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1880 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1881 | `hyper_all` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1882 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1883 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1884 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1885 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 1960 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2083 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2138 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2189 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2190 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2201 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2202 | `BINANCE` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2414 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2414 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2414 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2419 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2419 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2419 | `oi_low` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2512 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2514 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/strategy.go` | 2676 | `hyper_main` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/trader.go` | 66 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/trader.go` | 66 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/trader.go` | 67 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/trader.go` | 67 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/trader.go` | 163 | `AI500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/trader.go` | 164 | `oi_top` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 5 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 5 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 5 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 7 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 7 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 7 | `indodax` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 12 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 12 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 12 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 18 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 21 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 21 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 21 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 21 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 23 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 29 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 31 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 31 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 80 | `Hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `store/visibility.go` | 80 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `store/visibility.go` | 84 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `telegram/agent/prompt.go` | 79 | `ai500` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 427 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 428 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 443 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 444 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 467 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 467 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 467 | `x402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 468 | `USDC` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `telegram/bot.go` | 469 | `claw402` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 131 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 136 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 136 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 137 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 137 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 150 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 151 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 163 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 182 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 184 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 185 | `Lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 222 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 223 | `wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 223 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 223 | `Wallet` | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| `trader/auto_trader_decision.go` | 255 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `lighter` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `hyperliquid` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `bitget` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `aster` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 363 | `kucoin` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 382 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 524 | `binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 524 | `bybit` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 524 | `okx` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 660 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/auto_trader_decision.go` | 704 | `USDT` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/ninjatrader/bars_market_bridge.go` | 14 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/ninjatrader/reconcile_late_fill.go` | 258 | `USDT` | KEEP | CR-B | CommissionAsset: "USDT" — NT8 SIM snapshot mirror, KEEP byte-identical (CTO ruling, plan C13) |
| `trader/ninjatrader/tcp_trader.go` | 1258 | `Wallet` | KEEP | CR-B | "totalWalletBalance": acct.CashValue — account snapshot shape, KEEP byte-identical (CTO ruling, plan C13) |
| `trader/ninjatrader/tcp_trader.go` | 1359 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/ninjatrader/tcp_trader.go` | 1410 | `Binance` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/ninjatrader/transport.go` | 68 | `CoinAnk` | DELETE | CR-B | crypto-only site (plan v10 C4/C6/C7 family — review before code) |
| `trader/protection_reconciler.go` | 425 | `"MIXED"` | KEEP | CR-B | bracketOCO = "MIXED" — bracket-consistency state, not a coin source (plan C13) |
