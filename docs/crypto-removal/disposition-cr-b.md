# Crypto-removal disposition table — CR-B (DS-107), plan v10 FINAL
# The gate (scripts/crypto-union-gate.sh, DS-102) owns this format: headers below + plain pipe rows.
branch-point: ece4882c88290e1eae2529167a2de4f6754b74e7
integrator-tip: ece4882c88290e1eae2529167a2de4f6754b74e7
paths: manager kernel market config store agent branding internal ninjascript screenshots cmd scripts deploy docker nginx .github patches hook telegram provider/ninjatrader provider/databento trader/ninjatrader SECURITY.md Makefile .env.example docs/superpowers/AUDIT-CHECKLIST.md docs/superpowers/SYSTEM-MAP.md trader/auto_trader_decision.go api/handler_debug.go api/strategy_effective.go api/handler_competition.go api/handler_plan_order_truth.go api/handler_order.go api/handler_trader_config.go trader/protection_reconciler.go
regex: binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续|btc|ethusdt|\beth\b|altcoin|ethereum
# Generated:  @ ece4882c8829 (2026-10-02T00:43:39.648835+00:00); branch + sha only, never a filesystem path
# Swept: 636 tracked non-test files under CR-B paths (git ls-files -z, grep -I semantics); hits in 50 files.
# EXACTLY ONE OWNER PER HIT LINE (Finding 3 file-level ruling): CR-A-owned lines are ceded rows (owner CR-A, disposition mirrored from their canonical table); everything else here is CR-B. One file, one owner.
# Dispositions are reconciled to the tree (CTO ruling 2026-10-01): KEEP =
# futures-core text kept as-is; CUT = crypto content in a KEPT file, the cut is
# owed; DELETE rows appear only with file-level evidence. Nothing defaults.

| path | line | token | disposition | owner | reason |
|---|---|---|---|---|---|
| .github/ISSUE_TEMPLATE/bug_report.md | 118 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/ISSUE_TEMPLATE/bug_report.md | 119 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| CHANGELOG.zh-CN.md | 45 | Binance | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 90 | USDT | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 91 | USDT | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 92 | USDT | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 103 | BTC | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 110 | USDT | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 119 | Aster | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 121 | Aster | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 151 | 币安 | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| CHANGELOG.zh-CN.md | 170 | 币安 | KEEP | CR-B | historical changelog entry — records what shipped, not what ships |
| README.ja.md | 365 | btc | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 366 | altcoin | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 446 | btc | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 447 | altcoin | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 468 | btc | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 469 | altcoin | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 492 | btc | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 493 | altcoin | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 500 | btc | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 501 | altcoin | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| README.ja.md | 508 | altcoin | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| agent/central_brain.go | 1177 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/central_brain.go | 1178 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/central_brain.go | 1200 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/central_brain.go | 1201 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_domain_context.go | 41 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 315 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 319 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 326 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 327 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 350 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 351 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 362 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 363 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 392 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 393 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 404 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 405 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 430 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 434 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 561 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 563 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_execution_handlers.go | 565 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_execution_handlers.go | 566 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 570 | altcoin | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_execution_handlers.go | 571 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 573 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 575 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 576 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 580 | altcoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 812 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 813 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 814 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 817 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_execution_handlers.go | 819 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 820 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 825 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 826 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 828 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 1514 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 1524 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 1528 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_execution_handlers.go | 2396 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 2399 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 2401 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 2403 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_execution_handlers.go | 2437 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 2438 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 2500 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_execution_handlers.go | 2508 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 513 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 514 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 515 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 602 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 603 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 684 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 685 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 686 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 687 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 771 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 811 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 812 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_management_handlers.go | 813 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 868 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 873 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 959 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 960 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_management_handlers.go | 961 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1106 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1107 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1111 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1112 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1553 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_management_handlers.go | 1554 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1601 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_management_handlers.go | 1602 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 2149 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 2150 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skills/exchange_management.json | 31 | okx | CUT | CR-B | agent skill content — reword to futures (ninjatrader), the skill file stays |
| agent/skills/strategy_diagnosis.json | 16 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skills/strategy_management.json | 68 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skills/strategy_management.json | 72 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skills/strategy_management.json | 74 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skills/trader_diagnosis.json | 21 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/strategy_field_catalog.go | 43 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/strategy_field_catalog.go | 44 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/strategy_field_catalog.go | 102 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/strategy_field_catalog.go | 103 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 826 | Wallet | KEEP | CR-A | ceded to CR-A — skill/tool surface — cut deferred until DS-107 skill-handler consumer cuts land |
| agent/tools.go | 827 | USDC | KEEP | CR-A | ceded to CR-A — skill/tool surface — cut deferred until DS-107 skill-handler consumer cuts land |
| agent/tools.go | 941 | Wallet | KEEP | CR-A | ceded to CR-A — NT8 account-snapshot shape — balance key family (CTO ruling, byte-identical) |
| agent/tools.go | 2141 | btc | KEEP | CR-A | ceded to CR-A — risk-cap knob family (union-gate risk-cap canary) |
| agent/tools.go | 2142 | altcoin | KEEP | CR-A | ceded to CR-A — risk-cap knob family (union-gate risk-cap canary) |
| agent/tools.go | 2144 | AI500 | KEEP | CR-A | ceded to CR-A — skill/tool surface — cut deferred until DS-107 skill-handler consumer cuts land |
| agent/tools.go | 2164 | BTC | KEEP | CR-A | ceded to CR-A — risk-cap knob family (union-gate risk-cap canary) |
| agent/tools.go | 2165 | Altcoin | KEEP | CR-A | ceded to CR-A — risk-cap knob family (union-gate risk-cap canary) |
| agent/tools.go | 2167 | AI500 | KEEP | CR-A | ceded to CR-A — skill/tool surface — cut deferred until DS-107 skill-handler consumer cuts land |
| agent/tools.go | 2236 | BTC | KEEP | CR-A | ceded to CR-A — risk-cap knob family (union-gate risk-cap canary) |
| agent/tools.go | 2237 | Altcoin | KEEP | CR-A | ceded to CR-A — risk-cap knob family (union-gate risk-cap canary) |
| agent/tools.go | 2239 | AI500 | KEEP | CR-A | ceded to CR-A — skill/tool surface — cut deferred until DS-107 skill-handler consumer cuts land |
| agent/trade.go | 21 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 22 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 327 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 328 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 346 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 357 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 379 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/trade.go | 380 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/trade.go | 396 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 405 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agents.md | 219 | BTC | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| api/handler_competition.go | 154 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 418 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 419 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 420 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 432 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_order.go | 102 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_order.go | 103 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_plan_order_truth.go | 230 | "mixed" | KEEP | CR-B | plan-state string (plan C13) |
| branding/no_crypto.go | 16 | aster | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 17 | quant | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 18 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 19 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 20 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 24 | btc | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 25 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 26 | btc | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 28 | usdt | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 30 | ETH | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 31 | eth | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 32 | altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 34 | ethereum | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 39 | binance | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 44 | binance | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 51 | aster | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 52 | hyperliquid | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 53 | lighter | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 54 | coinank | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 55 | wallet | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 62 | binance | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 63 | lighter | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 69 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 78 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 79 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 80 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 81 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 82 | btc | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 83 | altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 84 | btc | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 85 | altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 87 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 88 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 89 | BTC | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 90 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 92 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 93 | Altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 95 | altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 96 | altcoin | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 100 | "mixed" | KEEP | CR-B | the guard's own regex/amendment text |
| branding/no_crypto.go | 113 | wallet | KEEP | CR-B | the guard's own regex/amendment text |
| clock-seams.list | 35 | binance | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| clock-seams.list | 36 | BINANCE | KEEP | CR-A | ceded to CR-A — historical doc/config file — not shipped code, KEEP byte-identical |
| cmd/picture_htf_replay/main.go | 21 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| cmd/picture_htf_replay/main.go | 109 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| cmd/picture_htf_replay/main.go | 112 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| cmd/picture_htf_replay/main.go | 355 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| cmd/picture_htf_replay/main.go | 356 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| cmd/picture_htf_replay/main.go | 360 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| hook/README.md | 60 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 62 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 66 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 68 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 74 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 116 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 117 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 126 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 211 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| hook/README.md | 270 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| kernel/engine.go | 136 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine.go | 137 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_analysis.go | 559 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 560 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 561 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 562 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 836 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| kernel/engine_analysis.go | 862 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_analysis.go | 877 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 31 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 33 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 43 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 58 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 59 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 73 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 74 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 75 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 92 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 94 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 95 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 96 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 100 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 109 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 110 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_position.go | 112 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 69 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_prompt.go | 70 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 71 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 73 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_prompt.go | 74 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 75 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 81 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 82 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 83 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 84 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 86 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 89 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 90 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_prompt.go | 100 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 101 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 154 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 155 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 156 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 157 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_prompt.go | 158 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 262 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| kernel/engine_prompt.go | 302 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 313 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 402 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 464 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 621 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| kernel/engine_prompt.go | 631 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| main.go | 252 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| main.go | 371 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 13 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 21 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 87 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 148 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 354 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data_freshness.go | 18 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 28 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 144 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 148 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/types.go | 38 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| ninjascript/VLBarsSubscriptionManager.cs | 81 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLBarsSubscriptionManager.cs | 96 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLBarsSubscriptionManager.cs | 349 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLBarsSubscriptionManager.cs | 631 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLContractResolver.cs | 66 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLContractResolver_VERIFY.md | 20 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLContractResolver_VERIFY.md | 123 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLContractResolver_VERIFY.md | 124 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/VLHistoryPull.cs | 132 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| ninjascript/vltrader_tcp_PROTOCOL.md | 333 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| patches/partner-2026-08-20/MANIFEST.md | 458 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| provider/ninjatrader/bar_source.go | 18 | "mixed" | KEEP | CR-B | BarSourceMixed = "mixed" — futures bar-source enum (plan C13) |
| provider/ninjatrader/tcp_framing.go | 675 | "mixed" | KEEP | CR-B | comment describes BarSourceMixed (plan C13) |
| scripts/replay_write_time_feasibility.py | 162 | "mixed" | KEEP | CR-B | replay-source rank key 'mixed', not a crypto source |
| store/ai_charge.go | 152 | USDC | KEEP | CR-A | ceded to CR-A — EstimateRunway stays per CTO ruling — daily-cost/runway estimate; usdcBalance naming byte-identical (exported API; zero callers after the dead claw402 pre-launch check was cut) |
| store/ai_charge.go | 153 | usdc | KEEP | CR-A | ceded to CR-A — EstimateRunway stays per CTO ruling — daily-cost/runway estimate; usdcBalance naming byte-identical (exported API; zero callers after the dead claw402 pre-launch check was cut) |
| store/ai_charge.go | 160 | usdc | KEEP | CR-A | ceded to CR-A — EstimateRunway stays per CTO ruling — daily-cost/runway estimate; usdcBalance naming byte-identical (exported API; zero callers after the dead claw402 pre-launch check was cut) |
| store/ai_charge.go | 161 | usdc | KEEP | CR-A | ceded to CR-A — EstimateRunway stays per CTO ruling — daily-cost/runway estimate; usdcBalance naming byte-identical (exported API; zero callers after the dead claw402 pre-launch check was cut) |
| store/bar_history.go | 151 | "mixed" | KEEP | CR-B | NT8 live+historical census value + boot-line mixed=%d (plan C13) |
| store/bar_history_across_roll.go | 82 | "mixed" | KEEP | CR-B | comment describes the roll-straddling ("mixed") state (plan C13) |
| store/exchange.go | 133 | binance | KEEP | CR-A | ceded to CR-A — legacy crypto columns/params — stored rows must load (C1); column drop pending CTO ruling |
| store/exchange.go | 147 | binance | KEEP | CR-A | ceded to CR-A — legacy crypto columns/params — stored rows must load (C1); column drop pending CTO ruling |
| store/exchange.go | 157 | binance | KEEP | CR-A | ceded to CR-A — legacy crypto columns/params — stored rows must load (C1); column drop pending CTO ruling |
| store/knob_registry_table.go | 14 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 15 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 26 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 27 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 49 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 153 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 154 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 156 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 157 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 159 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 160 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 162 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 163 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 167 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 168 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 170 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 171 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 173 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 174 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 176 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 177 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 217 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 224 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 260 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 261 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 263 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 265 | oi_low | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 627 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 628 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 629 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 630 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1941 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1942 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1943 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1944 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1946 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1947 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1948 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1949 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2119 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2120 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2121 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2122 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2174 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| store/strategy.go | 2186 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| store/strategy.go | 2390 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 63 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 64 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 66 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 67 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 160 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 161 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 163 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 164 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 131 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 136 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 137 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 150 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 151 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 163 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 182 | Lighter | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 184 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 185 | Lighter | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 222 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 223 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| trader/auto_trader_decision.go | 255 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 363 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 382 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 524 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 660 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/auto_trader_decision.go | 704 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/ninjatrader/reconcile_late_fill.go | 258 | USDT | KEEP | CR-B | CommissionAsset: "USDT" — NT8 SIM snapshot mirror, KEEP byte-identical (CTO ruling, plan C13) |
| trader/ninjatrader/tcp_trader.go | 1258 | Wallet | KEEP | CR-B | "totalWalletBalance": acct.CashValue — account snapshot shape, KEEP byte-identical (CTO ruling, plan C13) |
| trader/protection_reconciler.go | 425 | "MIXED" | KEEP | CR-B | bracketOCO = "MIXED" — bracket-consistency state, not a coin source (plan C13) |
