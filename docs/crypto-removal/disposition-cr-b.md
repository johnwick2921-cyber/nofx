# Crypto-removal disposition table — CR-B (DS-107), plan v10 FINAL
# The gate (scripts/crypto-union-gate.sh, DS-102) owns this format: headers below + plain pipe rows.
branch-point: e7e7cc3d20621f9bf4ad4884ef62ebde2b4d828f
integrator-tip: e7e7cc3d20621f9bf4ad4884ef62ebde2b4d828f
paths: manager kernel market config store agent branding internal ninjascript screenshots cmd scripts deploy docker nginx .github patches hook telegram provider/ninjatrader provider/databento trader/ninjatrader SECURITY.md Makefile .env.example docs/superpowers/AUDIT-CHECKLIST.md docs/superpowers/SYSTEM-MAP.md trader/auto_trader_decision.go api/handler_debug.go api/strategy_effective.go api/handler_competition.go api/handler_plan_order_truth.go api/handler_order.go api/handler_trader_config.go trader/protection_reconciler.go
regex: binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续|btc|ethusdt|\beth\b|altcoin|ethereum
# Generated: crypto-removal-cr-b-table @ e7e7cc3d2062 (2026-10-01T16:10:23.630899+00:00); branch + sha only, never a filesystem path
# Swept: 640 tracked non-test files under CR-B paths (git ls-files -z, grep -I semantics); hits in 88 files.
# EXACTLY ONE OWNER PER HIT LINE (Finding 3 file-level ruling): CR-A-owned lines are ceded rows (owner CR-A, disposition mirrored from their canonical table); everything else here is CR-B. One file, one owner.
# Dispositions are reconciled to the tree (CTO ruling 2026-10-01): KEEP =
# futures-core text kept as-is; CUT = crypto content in a KEPT file, the cut is
# owed; DELETE rows appear only with file-level evidence. Nothing defaults.

| path | line | token | disposition | owner | reason |
|---|---|---|---|---|---|
| .github/ISSUE_TEMPLATE/bug_report.md | 105 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/ISSUE_TEMPLATE/bug_report.md | 118 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/ISSUE_TEMPLATE/bug_report.md | 119 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/ISSUE_TEMPLATE/bug_report.md | 155 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/SECURITY.md | 69 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/SECURITY.md | 72 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| .github/SECURITY.md | 73 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| .github/SECURITY.md | 74 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| .github/labeler.yml | 32 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/labeler.yml | 33 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/labeler.yml | 34 | aster | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/labeler.yml | 35 | okx | CUT | CR-B | crypto wording — rewrite futures-only |
| .github/labeler.yml | 36 | bybit | CUT | CR-B | crypto wording — rewrite futures-only |
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
| README.ja.md | 47 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 49 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 58 | Ethereum | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 60 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 68 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 69 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 72 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 74 | Aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 76 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 79 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 83 | Ethereum | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 86 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 92 | Aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 95 | aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 105 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 106 | bybit | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 107 | okx | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 108 | bitget | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 109 | kucoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 115 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 116 | aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 117 | lighter | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 160 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 170 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 173 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 180 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 186 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 219 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 234 | AI500 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 251 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 259 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 271 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 273 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 275 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 279 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 283 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 465 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 466 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 475 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 476 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 480 | oi_top | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 488 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 489 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 494 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 496 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 500 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 501 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 509 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 511 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 513 | Ethereum | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 515 | Ethereum | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 518 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 520 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 526 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 527 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 530 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 531 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 532 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 533 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 544 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 545 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 546 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 547 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 553 | Aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 555 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 558 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 561 | ETH | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 564 | Aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 566 | Aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 567 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| README.ja.md | 581 | Aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 584 | aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 598 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 599 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 605 | aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 615 | asterdex | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 630 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 631 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 642 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 643 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 653 | oi_top | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 659 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 661 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 672 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 673 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 674 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 675 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 676 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 677 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 684 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 685 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 688 | oi_top | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 692 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 700 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 706 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 707 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 711 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 713 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 714 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 719 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 730 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 731 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 738 | btc | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 739 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 746 | altcoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 768 | oi_top | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 775 | oi_top | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 782 | oi_top | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 821 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 930 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 934 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 963 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 964 | AI500 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 995 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 998 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 999 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1018 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1038 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1039 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1047 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1067 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1068 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1069 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1070 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1071 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1074 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1075 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1080 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1081 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1161 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1190 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1210 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1218 | AI500 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1220 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1221 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1235 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1237 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1239 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1240 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1242 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1253 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1254 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1263 | USDT | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1297 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.ja.md | 1342 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 5 | USDC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 14 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 15 | claw402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 32 | USDC | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 34 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 52 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 56 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 59 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| README.md | 62 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| README.md | 64 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 67 | claw402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 75 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 88 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 89 | bybit | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 90 | okx | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 91 | bitget | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 92 | kucoin | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 98 | hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 99 | aster | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 100 | lighter | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 114 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 116 | Claw402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 185 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 221 | x402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 222 | Claw402 | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 227 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| README.md | 228 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 143 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 144 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 171 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 177 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 178 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 220 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 378 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 379 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 406 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 412 | Hyperliquid | CUT | CR-B | crypto wording — rewrite futures-only |
| SECURITY.md | 413 | Binance | CUT | CR-B | crypto wording — rewrite futures-only |
| agent/agent.go | 63 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 486 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 498 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 515 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 524 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 586 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 603 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 665 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 667 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 678 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 681 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 682 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 701 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 703 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 825 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 827 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 829 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 843 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 845 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/agent.go | 905 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/brain.go | 207 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 208 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 209 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 210 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 224 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 225 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 226 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 228 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 234 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/brain.go | 235 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/central_brain.go | 94 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/central_brain.go | 483 | AI500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/central_brain.go | 856 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/central_brain.go | 908 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/central_brain.go | 1177 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/central_brain.go | 1178 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/central_brain.go | 1200 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/central_brain.go | 1201 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/config_validation.go | 26 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 75 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 79 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/config_validation.go | 102 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 106 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/config_validation.go | 158 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 162 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/config_validation.go | 207 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 226 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 227 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 238 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/config_validation.go | 239 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/config_validation.go | 251 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 255 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/config_validation.go | 284 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 285 | OKX | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 286 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 287 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 288 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 289 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 290 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 291 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 336 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 337 | OKX | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 451 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/config_validation.go | 453 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/entity_field_catalog.go | 37 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/entity_field_catalog.go | 38 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/entity_field_catalog.go | 39 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/entity_field_catalog.go | 40 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/entity_field_catalog.go | 41 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/entity_field_catalog.go | 42 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/entity_field_catalog.go | 43 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 8 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 11 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 12 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 17 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 20 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 21 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 59 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/i18n.go | 60 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 224 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 228 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/llm_flow_extractor.go | 249 | AI500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 251 | ai500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 426 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 427 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 428 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 429 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/llm_flow_extractor.go | 430 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 16 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 32 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 33 | Claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 36 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 39 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 44 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 45 | BlockRun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 48 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 51 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 54 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 55 | BlockRun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 58 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 61 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_provider_catalog.go | 145 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 147 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 174 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 175 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 176 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 198 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 199 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 200 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 213 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 214 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 215 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 216 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 217 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 218 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 224 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 225 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 226 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 227 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 228 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_provider_catalog.go | 229 | blockrun | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 9 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 14 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 15 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 18 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 19 | usdc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 20 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 26 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 27 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 33 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 35 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 40 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 47 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 49 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 53 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 56 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 57 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 58 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 60 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 62 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 64 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 68 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 72 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 75 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 76 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 77 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 79 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/model_wallet_fastpath.go | 80 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/model_wallet_fastpath.go | 83 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 129 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 130 | OKX | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 131 | Bybit | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 132 | Bitget | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 134 | KuCoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 135 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 138 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 139 | OKX | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 140 | Bybit | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 141 | Bitget | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 143 | KuCoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/onboard.go | 144 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/planner_runtime.go | 157 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/planner_runtime.go | 1752 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/planner_runtime.go | 2260 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/planner_runtime.go | 2670 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/planner_runtime.go | 3767 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/planner_runtime.go | 3769 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/sentinel.go | 116 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/sentinel.go | 118 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/sentinel.go | 148 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
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
| agent/skill_execution_handlers.go | 563 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
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
| agent/skill_execution_handlers.go | 817 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
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
| agent/skill_management_handlers.go | 174 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 514 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 515 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 516 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 603 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 604 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 685 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 686 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 687 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 688 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 765 | hyper_all | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 767 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 772 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 783 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 787 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 805 | hyper_all | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 807 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 812 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 813 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 814 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 825 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 829 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 859 | hyper_all | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 869 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 874 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 885 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 889 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 960 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 961 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 962 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1080 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1107 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1108 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1112 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1113 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1115 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1133 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1273 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1278 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1284 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1287 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1289 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1291 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1292 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1293 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1295 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1297 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1298 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1301 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1320 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1325 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1331 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1334 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1336 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1338 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1339 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1340 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1342 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1344 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1345 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1348 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1614 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1615 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1662 | BTC | KEEP | CR-B | label naming the live futures risk caps — KEEP byte-identical (owner removal-only ruling 10-01 11:0x) |
| agent/skill_management_handlers.go | 1663 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 1735 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1738 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1742 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1744 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1745 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1746 | aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1748 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1749 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1750 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1753 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1754 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1755 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1757 | lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1759 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1760 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1761 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1764 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1765 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1766 | Lighter | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1821 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_management_handlers.go | 1822 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_management_handlers.go | 1824 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1825 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1837 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_management_handlers.go | 1838 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_management_handlers.go | 1840 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1841 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_management_handlers.go | 1976 | okx | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1980 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1981 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 1982 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 2003 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 2007 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/skill_management_handlers.go | 2027 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 2256 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 2263 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 2276 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 2277 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/skill_management_handlers.go | 2292 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/skill_management_handlers.go | 2299 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
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
| agent/tools.go | 31 | binance | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 54 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 66 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 69 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 274 | hyper_all | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 275 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 312 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 313 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 333 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 335 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 370 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 382 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 386 | claw402 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 403 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 415 | OKX | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 417 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 418 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 419 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 420 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 421 | Aster | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 422 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 423 | LIGHTER | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 424 | LIGHTER | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 425 | LIGHTER | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 524 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 540 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 541 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 545 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 685 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 734 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 751 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 776 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 811 | AI500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 850 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 923 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 927 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 943 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 944 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1027 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1031 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1066 | Wallet | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1417 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1418 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1422 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1453 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1454 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1466 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1470 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1489 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1494 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1550 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1551 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1552 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1558 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1559 | Hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1560 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1570 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1571 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1572 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1616 | hyperliquid | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 1620 | Wallet | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1636 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 1641 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| agent/tools.go | 2351 | btc | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2352 | altcoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2354 | AI500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2374 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2375 | Altcoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2377 | AI500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2446 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2447 | Altcoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2449 | AI500 | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2977 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 2978 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3014 | USDT | CUT | CR-A | ceded to CR-A — broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3015 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3045 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3057 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3066 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3071 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3195 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 3197 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3199 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3200 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 3204 | Altcoin | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3212 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3214 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3231 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 3239 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/tools.go | 3272 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3273 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3292 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3574 | hyper_all | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3575 | hyper_main | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3576 | hyper_main | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3607 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3609 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3690 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3691 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3693 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3698 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3699 | USDC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3720 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3729 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3730 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3731 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3737 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3738 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3749 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/tools.go | 3755 | USDT | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/trade.go | 21 | USDT | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 22 | USDT | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 327 | USDT | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 328 | USDT | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 346 | USDT | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 357 | USDT | KEEP | CR-A | ceded to CR-A — futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 379 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/trade.go | 380 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| agent/trade.go | 396 | USDT | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| agent/trade.go | 405 | USDT | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| agent/web.go | 50 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 56 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 58 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 208 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 212 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 228 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 231 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 235 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 243 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 246 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 251 | BTC | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 283 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 288 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 341 | Binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agent/web.go | 347 | binance | CUT | CR-B | agent crypto text/branch — cut with the crypto exchange layer |
| agents.md | 219 | BTC | CUT | CR-B | crypto wording — rewrite futures-only |
| api/handler_competition.go | 154 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 418 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 419 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 420 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_competition.go | 432 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| api/handler_order.go | 102 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_order.go | 103 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_order.go | 108 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_order.go | 109 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_order.go | 315 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_order.go | 371 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| api/handler_plan_order_truth.go | 230 | "mixed" | KEEP | CR-B | plan-state string (plan C13) |
| branding/no_crypto.go | 16 | aster | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 17 | quant | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 18 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 19 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 20 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 23 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 24 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| branding/no_crypto.go | 25 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 27 | usdt | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 29 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 30 | eth | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 31 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| branding/no_crypto.go | 33 | ethereum | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 37 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 42 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 49 | aster | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 50 | hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 51 | lighter | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 52 | coinank | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 53 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| branding/no_crypto.go | 60 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 61 | lighter | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 65 | "mixed" | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| branding/no_crypto.go | 78 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| clock-seams.list | 35 | binance | CUT | CR-B | crypto wording — rewrite futures-only |
| clock-seams.list | 36 | BINANCE | CUT | CR-B | crypto wording — rewrite futures-only |
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
| hook/trader_hook.go | 3 | Binance | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| kernel/engine.go | 53 | hyper_all | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine.go | 137 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine.go | 138 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine.go | 329 | claw402 | KEEP | CR-B | compat shim: the claw402WalletKey param is retained for caller compatibility (D2-DEAD item 12) |
| kernel/engine.go | 330 | claw402 | KEEP | CR-B | compat shim: the claw402WalletKey param is retained for caller compatibility (D2-DEAD item 12) |
| kernel/engine.go | 332 | claw402 | KEEP | CR-B | compat shim: the claw402WalletKey param is retained for caller compatibility (D2-DEAD item 12) |
| kernel/engine.go | 429 | Quant | KEEP | CR-B | English word 'quant' (data section header), not the quant feed |
| kernel/engine_analysis.go | 560 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 561 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 562 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 563 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| kernel/engine_analysis.go | 835 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_analysis.go | 837 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| kernel/engine_analysis.go | 863 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_analysis.go | 878 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
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
| kernel/engine_prompt.go | 485 | hyper_all | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 487 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 492 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 499 | hyper_all | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 500 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 501 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 502 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/engine_prompt.go | 638 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| kernel/engine_prompt.go | 648 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| kernel/engine_prompt_futures.go | 89 | USDT | KEEP | CR-B | documents why the USDT-perp framing stays excluded from the futures prompt |
| kernel/formatter.go | 108 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 109 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 150 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 151 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 202 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 233 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 235 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 238 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 343 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 344 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 386 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 387 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 437 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 467 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 469 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/formatter.go | 472 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/grid_engine.go | 116 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/grid_engine.go | 150 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/grid_engine.go | 151 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/grid_engine.go | 168 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/grid_engine.go | 205 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/grid_engine.go | 206 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 85 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 108 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 158 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 164 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 171 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 220 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 243 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 293 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 299 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 306 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/prompt_builder.go | 319 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/risk_limits.go | 13 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 67 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 76 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 106 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 113 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 120 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 180 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 189 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 206 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| kernel/schema.go | 213 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| main.go | 246 | CoinAnk | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| main.go | 252 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| main.go | 371 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| manager/trader_manager.go | 569 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 639 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 645 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 647 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 648 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 649 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 650 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 680 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 681 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 682 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 683 | bybit | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 684 | Bybit | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 685 | Bybit | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 686 | okx | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 687 | OKX | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 688 | OKX | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 689 | OKX | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 690 | bitget | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 691 | Bitget | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 692 | Bitget | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 693 | Bitget | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 697 | kucoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 698 | KuCoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 699 | KuCoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 700 | KuCoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 701 | hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 702 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 703 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 704 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 705 | aster | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 709 | lighter | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 711 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| manager/trader_manager.go | 715 | indodax | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 716 | Indodax | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 717 | Indodax | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 741 | Claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 779 | Wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| manager/trader_manager.go | 780 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 781 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 782 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| manager/trader_manager.go | 783 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| manager/trader_manager.go | 791 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 792 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 795 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| manager/trader_manager.go | 797 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| manager/trader_manager.go | 800 | wallet | CUT | CR-B | wallet wording — reworded to the futures meaning (what stays: the account/margin text) |
| market/api_client.go | 15 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/data.go | 16 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 28 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 30 | Binance | KEEP | CR-B | documents why the futures path reports funding/OI absent |
| market/data.go | 36 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 40 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 47 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 88 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 90 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 101 | CoinAnk | KEEP | CR-B | documents the BarCache-only futures read |
| market/data.go | 103 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 128 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 159 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 165 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 166 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 239 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 292 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 380 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 381 | Binance | KEEP | CR-B | documents why the futures path reports funding/OI absent |
| market/data.go | 382 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/data.go | 417 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 462 | binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 646 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 662 | ASTER | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 670 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 685 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 706 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 716 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 717 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 718 | USDT | KEEP | CR-B | documents the CME Normalize invariant (early-return first) |
| market/data.go | 721 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 733 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 739 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 748 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 754 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 757 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data.go | 777 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 18 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 28 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 144 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/data_freshness.go | 148 | ETH | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/futures_data.go | 5 | CoinAnk | KEEP | CR-B | documents the futures bypass of the crypto market read |
| market/futures_symbol.go | 5 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/futures_symbol.go | 15 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/futures_symbol.go | 38 | USDT | KEEP | CR-B | documents the CME Normalize invariant (early-return first) |
| market/futures_symbol.go | 48 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| market/historical.go | 12 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/historical.go | 13 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/historical.go | 36 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/historical.go | 44 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/historical.go | 60 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/historical.go | 98 | binance | DELETE | CR-A | ceded to CR-A — whole-file delete (C4) |
| market/types.go | 38 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| market/types.go | 127 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
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
| store/ai_charge.go | 152 | USDC | CUT | CR-A | ceded to CR-A — crypto lines cut per C4/C5; futures half stays |
| store/ai_charge.go | 153 | usdc | CUT | CR-A | ceded to CR-A — crypto lines cut per C4/C5; futures half stays |
| store/ai_charge.go | 160 | usdc | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/ai_charge.go | 161 | usdc | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/ai_model.go | 64 | Claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 65 | wallet | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 66 | Claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 68 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 70 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 431 | Claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 432 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 433 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 434 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 436 | Claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 442 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 443 | wallet | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 444 | wallet | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 445 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 447 | wallet | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 456 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 461 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 471 | claw402 | CUT | CR-A | ceded to CR-A — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/bar_history.go | 151 | "mixed" | KEEP | CR-B | NT8 live+historical census value + boot-line mixed=%d (plan C13) |
| store/bar_history_across_roll.go | 82 | "mixed" | KEEP | CR-B | comment describes the roll-straddling ("mixed") state (plan C13) |
| store/exchange.go | 32 | Hyperliquid | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 33 | Hyperliquid | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 37 | Wallet | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 101 | Hyperliquid | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 105 | Wallet | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 131 | binance | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 145 | binance | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 155 | binance | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 222 | hyperliquid | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 224 | Wallet | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 229 | hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 230 | Wallet | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 258 | Hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 259 | Hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 263 | Wallet | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 282 | hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 283 | Wallet | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 291 | hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 292 | hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 295 | wallet | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 372 | hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 375 | binance | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 377 | hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/exchange.go | 393 | Hyperliquid | CUT | CR-A | ceded to CR-A — crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/indicator_fingerprint.go | 12 | quant | KEEP | CR-B | records the D2-DEAD removal — historical note, keep |
| store/knob_registry_table.go | 14 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 15 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 26 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 27 | btc | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/knob_registry_table.go | 73 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/knob_registry_table.go | 169 | hyper_all | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/knob_registry_table.go | 170 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/order.go | 33 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/order.go | 390 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position.go | 746 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position.go | 747 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position.go | 748 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position.go | 795 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position.go | 796 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position.go | 797 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/position_builder.go | 146 | Lighter | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 25 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 50 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 154 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 155 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 157 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 158 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 160 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 161 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 163 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 164 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 168 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 169 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 171 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 172 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 174 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 175 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 177 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 178 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 218 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 225 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 261 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 262 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 264 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 266 | oi_low | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 514 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 628 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 629 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 630 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 631 | altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1828 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1832 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1874 | hyper_all | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1880 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1881 | hyper_all | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1882 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1883 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1884 | Hyperliquid | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1885 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1948 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1949 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1950 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1951 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1953 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1954 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1955 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 1956 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 1960 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2083 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2133 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2134 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2135 | BTC | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2136 | Altcoin | KEEP | CR-B | live futures risk caps (crypto-named) — KEEP byte-identical; the rename is a separate owner-gated wave (CTO ruling 10-01 10:5x) |
| store/strategy.go | 2138 | USDT | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2189 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2190 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| store/strategy.go | 2201 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2202 | BINANCE | KEEP | CR-B | guard-name comment (W-NO-BINANCE A) — documents the crypto-read refusal; nothing to cut |
| store/strategy.go | 2414 | ai500 | KEEP | CR-B | records the D2-DEAD removal — historical note, keep |
| store/strategy.go | 2419 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2512 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2514 | claw402 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2564 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/strategy.go | 2676 | hyper_main | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 63 | BTC | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 64 | Altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 66 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 67 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 160 | btc | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 161 | altcoin | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 163 | AI500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/trader.go | 164 | oi_top | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| store/visibility.go | 5 | hyperliquid | CUT | CR-A | ceded to CR-A — crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 53 | Hyperliquid | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| store/visibility.go | 57 | Wallet | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| telegram/agent/prompt.go | 79 | ai500 | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| telegram/bot.go | 517 | BTC | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| telegram/bot.go | 528 | BTC | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| telegram/bot.go | 562 | BTC | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| telegram/bot.go | 563 | BTC | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| telegram/bot.go | 583 | BTC | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
| telegram/bot.go | 584 | BTC | CUT | CR-A | ceded to CR-A — whole file is CR-A's per the Finding-3 file-level ruling |
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
| trader/ninjatrader/bars_market_bridge.go | 14 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/ninjatrader/reconcile_late_fill.go | 258 | USDT | KEEP | CR-B | CommissionAsset: "USDT" — NT8 SIM snapshot mirror, KEEP byte-identical (CTO ruling, plan C13) |
| trader/ninjatrader/tcp_trader.go | 1258 | Wallet | KEEP | CR-B | "totalWalletBalance": acct.CashValue — account snapshot shape, KEEP byte-identical (CTO ruling, plan C13) |
| trader/ninjatrader/tcp_trader.go | 1359 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/ninjatrader/tcp_trader.go | 1410 | Binance | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/ninjatrader/transport.go | 68 | CoinAnk | CUT | CR-B | crypto wording/branch in a kept file — cut to futures-only text |
| trader/protection_reconciler.go | 425 | "MIXED" | KEEP | CR-B | bracketOCO = "MIXED" — bracket-consistency state, not a coin source (plan C13) |
