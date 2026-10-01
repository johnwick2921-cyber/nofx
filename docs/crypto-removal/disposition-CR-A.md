# CR-A disposition table — base db412e61c — integrator tip db412e61c — brokers + payments + providers

base sha (branch point): db412e61c
integrator tip generated against: db412e61c
generated: 2026-10-01T00:42:09-05:00 (worktree /home/hoang/nofx-104-cra, branch crypto-removal-cr-a)
sweep regex (exported from plan v10 line 108, never hand-typed):
  `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续`

| path | line | token | disposition | owner | reason |
|---|---|---|---|---|---|
| agent/agent.go | 29 | `wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 64 | `Wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 65 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 71 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 169 | `wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 177 | `claw402` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 178 | `x402` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 226 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 236 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 238 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 249 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 250 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 277 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 278 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 285 | `wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 286 | `wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 289 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 296 | `USDC` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 298 | `claw402` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 326 | `wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 362 | `claw402` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 459 | `Wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 508 | `Wallet` | CUT | CR-A | payment lines go (wallet import, USDC ranking, claw402 provider cases + URL, wallet-balance gate); native provider path stays |
| agent/agent.go | 745 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 758 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 761 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 762 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 781 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 783 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 907 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 909 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/agent.go | 985 | `wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 24 | `aster` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 25 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 26 | `bitget` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 27 | `bybit` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 29 | `hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 30 | `indodax` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 31 | `kucoin` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 32 | `lighter` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 34 | `okx` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 41 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 64 | `wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 76 | `wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 79 | `usdt` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 284 | `hyper_all` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 285 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 343 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 345 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 380 | `claw402` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 392 | `claw402` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 396 | `claw402` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 413 | `binance` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 425 | `OKX` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 427 | `hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 428 | `hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 429 | `Aster` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 430 | `Aster` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 431 | `Aster` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 432 | `wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 433 | `LIGHTER` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 434 | `LIGHTER` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 435 | `LIGHTER` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 534 | `wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 550 | `hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 551 | `hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 555 | `wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 695 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 744 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 761 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 786 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 821 | `AI500` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 860 | `USDT` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 933 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 937 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 953 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 954 | `USDC` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1037 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1041 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1065 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1066 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1067 | `bybit` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1068 | `bybit` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1069 | `okx` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1070 | `okx` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1071 | `bitget` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1072 | `bitget` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1075 | `kucoin` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1076 | `kucoin` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1077 | `indodax` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1078 | `indodax` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1079 | `hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1080 | `hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1082 | `Hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1084 | `Hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1086 | `aster` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1087 | `aster` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1092 | `lighter` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1093 | `lighter` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1094 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1110 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key family in the agent balance extractor (CTO ruling) |
| agent/tools.go | 1146 | `USDC` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1149 | `wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1150 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1151 | `USDC` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1152 | `USDC` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1472 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1473 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1477 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1505 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1506 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1518 | `hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1522 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1540 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1545 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1599 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1600 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1601 | `Hyperliquid` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1607 | `Wallet` | - | CR-B | ceded to CR-B (union coverage) |
| agent/tools.go | 1608 | `Hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1609 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1619 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1620 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1621 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1653 | `hyperliquid` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1657 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1672 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 1677 | `Wallet` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 2390 | `AI500` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 2413 | `AI500` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 2485 | `AI500` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3013 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3014 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3050 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3051 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3081 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3093 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3102 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3107 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3248 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3250 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3308 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3309 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3328 | `binance` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3610 | `hyper_all` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3611 | `hyper_main` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3612 | `hyper_main` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3643 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3645 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3727 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3734 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3735 | `USDC` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3756 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3765 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3766 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3767 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3773 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/tools.go | 3791 | `USDT` | CUT | CR-A | broker factory arms + binanceFuturesAPIBaseURL/get_kline family + model-list USDC block go; ninjatrader arm stays |
| agent/trade.go | 21 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 22 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 60 | `USDT` | CUT | CR-A | crypto symbol example/comment and USDT-suffix stripping go |
| agent/trade.go | 327 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 328 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 346 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 357 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 362 | `USDT` | CUT | CR-A | crypto symbol example/comment and USDT-suffix stripping go |
| agent/trade.go | 400 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 413 | `USDT` | KEEP | CR-A | futures-active chat-entry notional caps; USDT is legacy naming for the value currency |
| agent/trade.go | 486 | `USDT` | CUT | CR-A | crypto symbol example/comment and USDT-suffix stripping go |
| agent/trade.go | 487 | `USDT` | CUT | CR-A | crypto symbol example/comment and USDT-suffix stripping go |
| api/exchange_account_state.go | 14 | `aster` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 15 | `binance` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 16 | `bitget` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 17 | `bybit` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 19 | `hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 20 | `indodax` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 21 | `kucoin` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 22 | `lighter` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 24 | `okx` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 210 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance/wallet_balance keys in the balance scan (CTO ruling) |
| api/exchange_account_state.go | 248 | `binance` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 249 | `binance` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 250 | `bybit` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 251 | `bybit` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 252 | `okx` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 253 | `okx` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 254 | `bitget` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 255 | `bitget` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 258 | `kucoin` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 259 | `kucoin` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 260 | `indodax` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 261 | `indodax` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 262 | `hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 263 | `hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 265 | `Hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 267 | `Hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 269 | `aster` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 270 | `aster` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 275 | `lighter` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 276 | `lighter` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 277 | `Wallet` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 294 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance/wallet_balance keys in the balance scan (CTO ruling) |
| api/exchange_account_state.go | 340 | `hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 341 | `USDC` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 343 | `USDT` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 353 | `Hyperliquid` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/exchange_account_state.go | 357 | `Wallet` | CUT | CR-A | broker imports/factory arms/USDC+USDT quote cases go; NT arm stays |
| api/handler_ai_model.go | 16 | `wallet` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 43 | `Wallet` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 44 | `USDC` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 107 | `claw402` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 109 | `wallet` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 110 | `Wallet` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 111 | `USDC` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 113 | `claw402` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 277 | `blockrun` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 278 | `blockrun` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_ai_model.go | 279 | `claw402` | CUT | CR-A | wallet import, wallet-address/balance columns, blockrun/claw402 list entries go |
| api/handler_competition.go | 154 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_competition.go | 418 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_competition.go | 419 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_competition.go | 420 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_competition.go | 432 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_exchange.go | 31 | `binance` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 40 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 42 | `Aster` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 43 | `Aster` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 44 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 65 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 69 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 83 | `OKX` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 85 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 86 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 90 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 103 | `binance` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 110 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 111 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 115 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 244 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 245 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 246 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 256 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 257 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 258 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 277 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 281 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 298 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 321 | `wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 397 | `binance` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 398 | `hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 410 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 414 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 429 | `Hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 431 | `Wallet` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 496 | `binance` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 497 | `bybit` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 498 | `okx` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 499 | `Gate.io` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 500 | `kucoin` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 501 | `hyperliquid` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 502 | `aster` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_exchange.go | 503 | `lighter` | CUT | CR-A | crypto credential fields/merge/venue rows/create-update arms go |
| api/handler_klines.go | 16 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 17 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 18 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 27 | `Coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 31 | `Coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 35 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 48 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 49 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 54 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 64 | `binance` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 87 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 88 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 89 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 91 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 97 | `CoinAnk` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 111 | `CoinAnk` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 113 | `Coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 115 | `CoinAnk` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 123 | `Coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 124 | `Coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 125 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 126 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 128 | `binance` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 129 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 130 | `bybit` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 131 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 132 | `okx` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 133 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 134 | `bitget` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 135 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 137 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 138 | `aster` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 139 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 140 | `lighter` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 141 | `Lighter` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 142 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 143 | `kucoin` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 144 | `KuCoin` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 145 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 147 | `Binance` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 148 | `Binance` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 149 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 152 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 153 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 156 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 158 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 160 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 162 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 164 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 166 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 168 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 170 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 172 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 174 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 176 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 178 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 180 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 182 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 184 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 186 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 188 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 190 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 192 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 194 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 196 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 200 | `OKX` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 202 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 203 | `USDT` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 204 | `USDT` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 205 | `USDT` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 206 | `USDT` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 210 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 214 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 216 | `OKX` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 217 | `Binance` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 218 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 219 | `CoinAnk` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 220 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 222 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 225 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 229 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 230 | `Coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 231 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 232 | `coinank` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 240 | `USDT` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 322 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 324 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 325 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 326 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 328 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 329 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 331 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 336 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 339 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 501 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 513 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 514 | `Hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 515 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_klines.go | 519 | `hyperliquid` | CUT | CR-A | coinank + hyperliquid branches go; NT8 BarCache + SVP stay (C6) |
| api/handler_onboarding.go | 16 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 30 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 37 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 40 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 42 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 71 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 76 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 78 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 88 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 98 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 104 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 108 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 110 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 111 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 116 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 117 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 122 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 124 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 128 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 129 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 130 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 132 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 137 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 138 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 139 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 141 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 148 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 154 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 160 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 164 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 165 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 170 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 179 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 181 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 185 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 188 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 190 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 195 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 197 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 200 | `USDC` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 202 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 207 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 209 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 213 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 214 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 221 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 229 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 231 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 238 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 240 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 244 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 247 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 249 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 256 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 267 | `Claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 274 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 279 | `claw402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 282 | `wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 299 | `Wallet` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 313 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 314 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_onboarding.go | 315 | `CLAW402` | CUT | CR-A | claw402/wallet onboarding sections go (routes deleted) |
| api/handler_trader.go | 43 | `AI500` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 44 | `oi_top` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 109 | `binance` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 116 | `okx` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 126 | `hyperliquid` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 130 | `Hyperliquid` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 133 | `aster` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 135 | `Aster` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 138 | `Aster` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 141 | `Aster` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 143 | `lighter` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 144 | `Wallet` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 191 | `binance` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 221 | `hyperliquid` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 222 | `hyperliquid` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 223 | `aster` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 224 | `Aster` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 227 | `wallet` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 228 | `hyperliquid` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 359 | `USDT` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 361 | `USDT` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader.go | 522 | `AI500` | CUT | CR-A | crypto credential cases/refusals go; ninjatrader-only for NEW |
| api/handler_trader_status.go | 12 | `aster` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 13 | `binance` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 14 | `bitget` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 15 | `bybit` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 17 | `hyperliquid` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 18 | `kucoin` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 19 | `lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 21 | `okx` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 95 | `USDT` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 112 | `USDT` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 195 | `binance` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 198 | `binance` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 199 | `binance` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 200 | `hyperliquid` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 201 | `hyperliquid` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 203 | `Hyperliquid` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 205 | `Hyperliquid` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 207 | `aster` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 208 | `aster` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 213 | `bybit` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 214 | `bybit` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 218 | `okx` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 219 | `okx` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 224 | `bitget` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 225 | `bitget` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 235 | `kucoin` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 236 | `kucoin` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 241 | `lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 242 | `Wallet` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 243 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 244 | `lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 245 | `Wallet` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 248 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 251 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 319 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 323 | `binance` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 366 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 369 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 413 | `USDT` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 435 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 436 | `lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 488 | `USDT` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 513 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 516 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_trader_status.go | 517 | `Lighter` | CUT | CR-A | broker factory arms + Lighter order-recording path go |
| api/handler_wallet.go | 7 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 15 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 19 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 22 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 23 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 29 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 30 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 32 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 43 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 51 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 60 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 70 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 80 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 81 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 83 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 84 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 86 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 89 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 90 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 96 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 101 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 102 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 105 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 112 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 118 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| api/handler_wallet.go | 120 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| api/server.go | 131 | `Wallet` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 132 | `wallet` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 133 | `wallet` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 196 | `claw402` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 197 | `claw402` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 267 | `USDT` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 301 | `okx` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 305 | `USDT` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 309 | `okx` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 310 | `binance` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 312 | `binance` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 313 | `okx` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 314 | `hyperliquid` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 315 | `aster` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 316 | `lighter` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 319 | `okx` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 364 | `hyper_all` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 365 | `USDT` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/server.go | 389 | `USDT` | CUT | CR-A | wallet + claw402-onboarding routes and crypto doc lines go |
| api/strategy.go | 804 | `claw402` | CUT | CR-A | claw402 case of resolveStrategyDataWalletKey goes |
| api/utils.go | 43 | `Hyperliquid` | CUT | CR-A | hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop |
| api/utils.go | 47 | `Wallet` | CUT | CR-A | hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop |
| api/utils.go | 72 | `Hyperliquid` | CUT | CR-A | hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop |
| api/utils.go | 73 | `hyperliquid` | CUT | CR-A | hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop |
| api/utils.go | 81 | `Wallet` | CUT | CR-A | hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop |
| api/utils.go | 82 | `wallet` | CUT | CR-A | hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop |
| go.mod | 6 | `binance` | CUT | CR-A | SDK module drop (go mod tidy + verify) |
| go.mod | 9 | `bybit` | CUT | CR-A | SDK module drop (go mod tidy + verify) |
| go.mod | 10 | `lighter` | CUT | CR-A | SDK module drop (go mod tidy + verify) |
| go.mod | 12 | `gateio` | CUT | CR-A | SDK module drop (go mod tidy + verify) |
| go.mod | 25 | `hyperliquid` | CUT | CR-A | SDK module drop (go mod tidy + verify) |
| hook/hooks.go | 37 | `BINANCE` | CUT | CR-A | NEW_BINANCE_TRADER registration goes with the broker dirs |
| hook/trader_hook.go | 7 | `binance` | CUT | CR-A | per-broker result types go with the broker dirs (same commit) |
| hook/trader_hook.go | 10 | `Binance` | CUT | CR-A | per-broker result types go with the broker dirs (same commit) |
| hook/trader_hook.go | 15 | `Binance` | CUT | CR-A | per-broker result types go with the broker dirs (same commit) |
| hook/trader_hook.go | 17 | `Binance` | CUT | CR-A | per-broker result types go with the broker dirs (same commit) |
| hook/trader_hook.go | 22 | `Binance` | CUT | CR-A | per-broker result types go with the broker dirs (same commit) |
| kernel/testdata/golden/user_prompt_crypto_change_na.txt | 8 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| kernel/testdata/golden/user_prompt_crypto_change_na.txt | 10 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| main.go | 247 | `CoinAnk` | CUT | CR-A | CoinAnk comment rewords (host census must read 0) |
| main.go | 253 | `BINANCE` | CUT | CR-A | CoinAnk comment rewords (host census must read 0) |
| main.go | 372 | `BINANCE` | CUT | CR-A | CoinAnk comment rewords (host census must read 0) |
| manager/trader_manager.go | 569 | `ai500` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 639 | `ai500` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 645 | `binance` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 647 | `Binance` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 648 | `Binance` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 649 | `Hyperliquid` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 650 | `Hyperliquid` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 680 | `binance` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 681 | `Binance` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 682 | `Binance` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 683 | `bybit` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 684 | `Bybit` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 685 | `Bybit` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 686 | `okx` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 687 | `OKX` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 688 | `OKX` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 689 | `OKX` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 690 | `bitget` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 691 | `Bitget` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 692 | `Bitget` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 693 | `Bitget` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 697 | `kucoin` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 698 | `KuCoin` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 699 | `KuCoin` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 700 | `KuCoin` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 701 | `hyperliquid` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 702 | `Hyperliquid` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 703 | `Hyperliquid` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 704 | `Hyperliquid` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 705 | `aster` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 709 | `lighter` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 711 | `Wallet` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 715 | `indodax` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 716 | `Indodax` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 717 | `Indodax` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 741 | `Claw402` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 779 | `Wallet` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 780 | `claw402` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 781 | `claw402` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 782 | `wallet` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 783 | `wallet` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 791 | `claw402` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 792 | `claw402` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 795 | `wallet` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 797 | `claw402` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| manager/trader_manager.go | 800 | `wallet` | CUT | CR-A | crypto credential map + Claw402WalletKey family go; NT arm stays |
| market/api_client.go | 15 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| market/historical.go | 12 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| market/historical.go | 13 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| market/historical.go | 36 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| market/historical.go | 44 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| market/historical.go | 60 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| market/historical.go | 98 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| mcp/client.go | 66 | `claw402` | CUT | CR-A | claw402 payment channel goes; native provider path stays |
| mcp/client.go | 74 | `claw402` | CUT | CR-A | claw402 payment channel goes; native provider path stays |
| mcp/client.go | 77 | `Claw402` | CUT | CR-A | claw402 payment channel goes; native provider path stays |
| mcp/client.go | 78 | `claw402` | CUT | CR-A | claw402 payment channel goes; native provider path stays |
| mcp/payment/claw402.go | 14 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 25 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 26 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 37 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 52 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 53 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 56 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 57 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 87 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 88 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 92 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 95 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 101 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 103 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 104 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 105 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 108 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 109 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 111 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 112 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 113 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 114 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 115 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 120 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 122 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 128 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 132 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 137 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 143 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 148 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 151 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 156 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 157 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 164 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 167 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 169 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 173 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 176 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 180 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 183 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 185 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 192 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 193 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 194 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 195 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 196 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 200 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 202 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 221 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 222 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 223 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 229 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 244 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 251 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 258 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 265 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 272 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 273 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 277 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/claw402.go | 278 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 26 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 27 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 29 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 31 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 32 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 34 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 39 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 44 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 46 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 48 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 49 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 51 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 54 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 57 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 64 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 78 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 80 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 81 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 82 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 83 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 84 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 87 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 88 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 98 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 99 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 105 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 107 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 109 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 111 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 113 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 124 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 125 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 127 | `Claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 131 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 132 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 135 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 138 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 143 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 145 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 148 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 152 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 155 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 156 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 160 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 194 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 200 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 211 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 212 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 214 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 215 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 244 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 245 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 257 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 258 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 260 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 261 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 264 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 265 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 269 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 272 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 295 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 301 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 305 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 349 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 356 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 367 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 368 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 370 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 371 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 401 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 402 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 414 | `claw402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 416 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 417 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 419 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 420 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 423 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 424 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 428 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 431 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 443 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 445 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 447 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 453 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 468 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 481 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 482 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 489 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 499 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 540 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 541 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 550 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 551 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 553 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 554 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 563 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 572 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 573 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 589 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 598 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 601 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 606 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 626 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 627 | `X402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/payment/x402.go | 700 | `x402` | DELETE | CR-A | whole-file delete (C4) |
| mcp/providers.go | 16 | `Claw402` | CUT | CR-A | ProviderClaw402 goes |
| provider/coinank/base_coin.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/base_coin.go | 4 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/base_coin.go | 6 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 6 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 7 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 10 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 11 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 26 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/base_coin.go | 32 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/depth_ws.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/depth_ws.go | 6 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/depth_ws.go | 12 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/depth_ws.go | 33 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/depth_ws.go | 51 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/depth_ws.go | 83 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 10 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 11 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 16 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 18 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 19 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 20 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 32 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 38 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline.go | 40 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 6 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 7 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 15 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 19 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 39 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 57 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 89 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 106 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 107 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 109 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 125 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_api/kline_ws.go | 136 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/exchange.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/exchange.go | 6 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/exchange.go | 11 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/exchange.go | 13 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/exchange.go | 22 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/exchange.go | 24 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/instrument_agg_sort_by.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/interval.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/product_type.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/side.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/sort_type.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/url.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/url.go | 3 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_enum/url.go | 5 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 15 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 16 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 21 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 22 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 28 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 40 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 41 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 64 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/coinank_http.go | 65 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 6 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 11 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 12 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 17 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 31 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 32 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 41 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instrument_agg_rank.go | 44 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/instruments.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/kline.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/kline.go | 6 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/kline.go | 11 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/kline.go | 13 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/kline.go | 30 | `Coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/liquidation.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/net_positions.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/open_interest.go | 1 | `coinank` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/open_interest.go | 31 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/open_interest.go | 33 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/coinank/open_interest.go | 34 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 16 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 26 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 50 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 62 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 67 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 136 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/coins.go | 153 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 15 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 16 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 19 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 47 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 53 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 63 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 127 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 178 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 215 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 238 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 240 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 241 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 248 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 254 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 264 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 268 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 276 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 319 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 391 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 392 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 393 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 395 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 396 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 397 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| provider/hyperliquid/kline.go | 406 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| store/ai_charge.go | 152 | `Claw402` | CUT | CR-A | crypto lines cut per C4/C5; futures half stays |
| store/ai_charge.go | 153 | `Claw402` | CUT | CR-A | crypto lines cut per C4/C5; futures half stays |
| store/ai_charge.go | 154 | `claw402` | CUT | CR-A | crypto lines cut per C4/C5; futures half stays |
| store/ai_charge.go | 157 | `USDC` | KEEP | CR-A | EstimateRunway stays (daily-cost/runway estimate keeps calling it); usdcBalance naming byte-identical |
| store/ai_charge.go | 158 | `usdc` | KEEP | CR-A | EstimateRunway stays (daily-cost/runway estimate keeps calling it); usdcBalance naming byte-identical |
| store/ai_charge.go | 165 | `usdc` | KEEP | CR-A | EstimateRunway stays (daily-cost/runway estimate keeps calling it); usdcBalance naming byte-identical |
| store/ai_charge.go | 166 | `usdc` | KEEP | CR-A | EstimateRunway stays (daily-cost/runway estimate keeps calling it); usdcBalance naming byte-identical |
| store/ai_model.go | 64 | `Claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 65 | `wallet` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 66 | `Claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 68 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 70 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 431 | `Claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 432 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 433 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 434 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 436 | `Claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 442 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 443 | `wallet` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 444 | `wallet` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 445 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 447 | `wallet` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 456 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 461 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/ai_model.go | 471 | `claw402` | CUT | CR-A | FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY |
| store/exchange.go | 32 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 33 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 37 | `Wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 101 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 105 | `Wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 131 | `binance` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 145 | `binance` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 155 | `binance` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 210 | `binance` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 211 | `Binance` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 212 | `bybit` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 213 | `Bybit` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 214 | `okx` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 215 | `OKX` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 216 | `bitget` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 217 | `Bitget` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 218 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 219 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 220 | `aster` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 221 | `Aster` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 222 | `lighter` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 223 | `LIGHTER` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 224 | `indodax` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 225 | `Indodax` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 238 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 240 | `Wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 245 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 246 | `Wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 274 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 275 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 279 | `Wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 298 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 299 | `Wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 307 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 308 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 311 | `wallet` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 388 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 391 | `binance` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 393 | `hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/exchange.go | 409 | `Hyperliquid` | CUT | CR-A | crypto credential fields go (columns STAY); migrateToMultiAccount deleted |
| store/trader.go | 66 | `AI500` | CUT | CR-A | UseAI500/UseOITop fields go; columns STAY (C1) |
| store/trader.go | 67 | `oi_top` | CUT | CR-A | UseAI500/UseOITop fields go; columns STAY (C1) |
| store/trader.go | 163 | `AI500` | CUT | CR-A | UseAI500/UseOITop fields go; columns STAY (C1) |
| store/trader.go | 164 | `oi_top` | CUT | CR-A | UseAI500/UseOITop fields go; columns STAY (C1) |
| store/visibility.go | 5 | `hyperliquid` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 7 | `binance` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 12 | `okx` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 18 | `hyperliquid` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 21 | `hyperliquid` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 23 | `aster` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 29 | `lighter` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 31 | `wallet` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 80 | `Hyperliquid` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| store/visibility.go | 84 | `Wallet` | CUT | CR-A | crypto cases + DEX wallet fields go; PR #188 D1-D3 cleanup-skip ported |
| telegram/bot.go | 427 | `USDC` | CUT | CR-A | USDC/claw402 payment gate goes |
| telegram/bot.go | 428 | `USDC` | CUT | CR-A | USDC/claw402 payment gate goes |
| telegram/bot.go | 443 | `USDC` | CUT | CR-A | USDC/claw402 payment gate goes |
| telegram/bot.go | 444 | `USDC` | CUT | CR-A | USDC/claw402 payment gate goes |
| telegram/bot.go | 467 | `USDC` | CUT | CR-A | USDC/claw402 payment gate goes |
| telegram/bot.go | 468 | `USDC` | CUT | CR-A | USDC/claw402 payment gate goes |
| telegram/bot.go | 469 | `claw402` | CUT | CR-A | USDC/claw402 payment gate goes |
| telemetry/experience.go | 59 | `claw402` | CUT | CR-A | claw402 channel value goes |
| trader/aster/trader.go | 1 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 26 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 29 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 30 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 31 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 48 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 49 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 50 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 51 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader.go | 78 | `asterdex` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 1 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 28 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 31 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 32 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 35 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 36 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 38 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 45 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 46 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 52 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 53 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 62 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 69 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 89 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 91 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 94 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 97 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 130 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 131 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 187 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 203 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 218 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 224 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_account.go | 259 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_orders.go | 1 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_orders.go | 251 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_orders.go | 623 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_orders.go | 692 | `ASTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_positions.go | 1 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_positions.go | 42 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_positions.go | 49 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_positions.go | 67 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_positions.go | 68 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 1 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 14 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 17 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 26 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 34 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 58 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 81 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 115 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 142 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 147 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 183 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 186 | `aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 189 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/aster/trader_sync.go | 193 | `Aster` | DELETE | CR-A | whole-file delete (C4) |
| trader/auto_trader.go | 21 | `aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 22 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 23 | `bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 24 | `bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 26 | `hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 27 | `indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 28 | `kucoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 29 | `lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 31 | `okx` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 32 | `wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 267 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 270 | `Binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 271 | `Binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 272 | `Binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 274 | `Bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 275 | `Bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 276 | `Bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 278 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 279 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 280 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 281 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 283 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 284 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 285 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 286 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 292 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 293 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 294 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 295 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 297 | `Indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 298 | `Indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 299 | `Indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 306 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 307 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 308 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 309 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 310 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 312 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 313 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 314 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 315 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 317 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 318 | `Wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 319 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 320 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 321 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 333 | `Claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 376 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 484 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 692 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 694 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 707 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 714 | `CoinAnk` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 733 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 734 | `Binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 735 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 736 | `bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 737 | `Bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 738 | `bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 739 | `okx` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 740 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 741 | `okx` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 742 | `bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 743 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 744 | `bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 746 | `Gate.io` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 748 | `kucoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 749 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 750 | `kucoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 751 | `hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 752 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 753 | `hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 755 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 757 | `aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 758 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 759 | `aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 761 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 763 | `lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 764 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 766 | `Wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 767 | `Lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 770 | `Lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 771 | `lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 772 | `Wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 775 | `Lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 778 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 780 | `LIGHTER` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 781 | `indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 782 | `Indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 783 | `indodax` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 843 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key in the auto-fetched balance scan (CTO ruling) |
| trader/auto_trader.go | 853 | `USDT` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1001 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1009 | `Lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1010 | `lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1011 | `lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1013 | `Lighter` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1017 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1018 | `hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1019 | `hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1020 | `hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1021 | `Hyperliquid` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1025 | `Bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1026 | `bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1027 | `bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1028 | `bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1029 | `Bybit` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1033 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1034 | `okx` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1035 | `okx` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1036 | `okx` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1037 | `OKX` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1041 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1042 | `bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1043 | `bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1044 | `bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1045 | `Bitget` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1049 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1050 | `aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1051 | `aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1053 | `Aster` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1057 | `Binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1058 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1059 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1060 | `binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1061 | `Binance` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1073 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1074 | `kucoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1075 | `kucoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1076 | `kucoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1077 | `KuCoin` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1330 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1332 | `Claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1336 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1338 | `wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1341 | `Wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1343 | `claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1344 | `Claw402` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1346 | `USDC` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1347 | `wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1349 | `USDC` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1358 | `USDC` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1362 | `USDC` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1365 | `USDC` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1374 | `Wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader.go | 1375 | `Wallet` | CUT | CR-A | crypto credential struct, broker factory + order-sync arms, claw402 pre-launch go; NT arm stays |
| trader/auto_trader_loop.go | 17 | `wallet` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 350 | `USDC` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 351 | `Claw402` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 352 | `Claw402` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 522 | `USDT` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 709 | `USDT` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1005 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key mirrors tcp_trader.go CashValue snapshot (CTO ruling, byte-identical) |
| trader/auto_trader_loop.go | 1010 | `wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key mirrors tcp_trader.go CashValue snapshot (CTO ruling, byte-identical) |
| trader/auto_trader_loop.go | 1011 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key mirrors tcp_trader.go CashValue snapshot (CTO ruling, byte-identical) |
| trader/auto_trader_loop.go | 1024 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key mirrors tcp_trader.go CashValue snapshot (CTO ruling, byte-identical) |
| trader/auto_trader_loop.go | 1025 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance key mirrors tcp_trader.go CashValue snapshot (CTO ruling, byte-identical) |
| trader/auto_trader_loop.go | 1041 | `Binance` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1088 | `Bybit` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1242 | `Claw402` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1243 | `Claw402` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1252 | `claw402` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1253 | `wallet` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1255 | `USDC` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1260 | `USDC` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1263 | `USDC` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/auto_trader_loop.go | 1270 | `USDC` | CUT | CR-A | wallet import, claw402 balance check, USDT logs/comments go; snapshot keys stay |
| trader/binance/futures.go | 1 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 14 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 45 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 67 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 73 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 110 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 111 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 114 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures.go | 121 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 1 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 25 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 28 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 33 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 37 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 38 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 51 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 52 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_account.go | 110 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 1 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 10 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 39 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 94 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 536 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 654 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 688 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_orders.go | 753 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 1 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 10 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 26 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 47 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 102 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 192 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 194 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/futures_positions.go | 210 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 1 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 18 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 19 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 22 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 25 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 33 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 34 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 35 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 61 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 155 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 248 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 281 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 282 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 283 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 290 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 352 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 355 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 356 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 357 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 358 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 364 | `binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 366 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 367 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/binance/order_sync.go | 371 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 1 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 16 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 17 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 32 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 33 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 38 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 42 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 52 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 53 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 71 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 74 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 75 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 79 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 84 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 87 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 90 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 98 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 107 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 108 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 134 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 141 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 155 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 158 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 159 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 167 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 175 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 216 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 277 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 281 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 282 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 284 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 286 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 287 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/order_sync.go | 291 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 1 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 19 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 21 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 22 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 23 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 24 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 25 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 26 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 27 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 28 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 29 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 30 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 31 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 32 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 35 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 36 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 55 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 63 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 64 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 75 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 76 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 83 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 84 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 90 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 96 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 101 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 104 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 110 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 112 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 116 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 124 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 128 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 129 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 138 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 173 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 201 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 202 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 206 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 207 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 210 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 213 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 214 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 215 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 216 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 221 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 233 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 237 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 266 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 291 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 303 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 313 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader.go | 314 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 1 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 13 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 23 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 26 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 35 | `Usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 45 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 49 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 55 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 71 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 81 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 82 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 86 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 103 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 108 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 109 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 113 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 127 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 132 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 135 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 162 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_account.go | 164 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 1 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 13 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 29 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 31 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 35 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 38 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 40 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 57 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 67 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 83 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 85 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 89 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 92 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 94 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 111 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 121 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 146 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 148 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 153 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 156 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 158 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 174 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 184 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 214 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 216 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 221 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 224 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 226 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 242 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 252 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 253 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 268 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 270 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 278 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 286 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 291 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 292 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 307 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 309 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 317 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 325 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 330 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 335 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 340 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 346 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 369 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 370 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 380 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 386 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 389 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 408 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 409 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 412 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 423 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 430 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 435 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 494 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 501 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 504 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 506 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 547 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 549 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 555 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 619 | `BITGET` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 625 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 631 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 646 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 648 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 654 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 662 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 664 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 678 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 695 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 700 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_orders.go | 709 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 1 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 12 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 22 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 23 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 26 | `bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 96 | `Bitget` | DELETE | CR-A | whole-file delete (C4) |
| trader/bitget/trader_positions.go | 105 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 1 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 21 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 22 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 38 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 39 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 43 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 44 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 47 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 67 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 78 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 100 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 106 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 107 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 108 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 153 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 175 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 178 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 179 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 187 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 195 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 236 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 270 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 297 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 301 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 302 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 304 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 306 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 307 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/order_sync.go | 311 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 1 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 15 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 18 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 19 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 20 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 42 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 43 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 46 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 61 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 69 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 86 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 96 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 99 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 138 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 144 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 169 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader.go | 179 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 1 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 18 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 33 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 35 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 39 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 45 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 50 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 60 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 61 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 62 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 64 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 70 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 71 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 72 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 77 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 92 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 93 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 94 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 98 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 99 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 102 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 122 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 133 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 153 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 159 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 160 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_account.go | 228 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 1 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 16 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 17 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 21 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 23 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 25 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 30 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 45 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 47 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 49 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 59 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 60 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 64 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 66 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 68 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 73 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 88 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 90 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 92 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 102 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 135 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 137 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 147 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 180 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 182 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 192 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 200 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 202 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 217 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 229 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 245 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 251 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 282 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 314 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 323 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 328 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 360 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 369 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 374 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 379 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 384 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 390 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 399 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 401 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 404 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 409 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 417 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 458 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 466 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 473 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 527 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 537 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 573 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 589 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 602 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 628 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 630 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 644 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 647 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 664 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 671 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 677 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 680 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 686 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 692 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_orders.go | 719 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 1 | `bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 13 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 26 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 29 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 31 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 35 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 40 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 87 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 90 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/bybit/trader_positions.go | 99 | `Bybit` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/order_sync.go | 15 | `gateio` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/order_sync.go | 48 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/order_sync.go | 137 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/order_sync.go | 184 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 11 | `gateio` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 14 | `Gate.io` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 56 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 62 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 63 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 64 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 65 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 70 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader.go | 88 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_account.go | 10 | `gateio` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_account.go | 25 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_account.go | 35 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_account.go | 61 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_account.go | 139 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 12 | `gateio` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 19 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 21 | `Gate.io` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 35 | `Gate.io` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 78 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 135 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 203 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 276 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 306 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 362 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 413 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 440 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 448 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 462 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 505 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 570 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/gate/trader_orders.go | 613 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 14 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 17 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 18 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 26 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 34 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 57 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 74 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 108 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 110 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 140 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 142 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 144 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 145 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/order_sync.go | 149 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 13 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 16 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 17 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 18 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 20 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 21 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 64 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 68 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 74 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 86 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 87 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 88 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 93 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 119 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 120 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 121 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 132 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 134 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 137 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 138 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 141 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 142 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 144 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 145 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 146 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 147 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 150 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 151 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 152 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 153 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 154 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 155 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 157 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 158 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 159 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 165 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 171 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 175 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 183 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 184 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 185 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 188 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 192 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 194 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 195 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 196 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 197 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 198 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 199 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 200 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 202 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 203 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 204 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 206 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 207 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 211 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 212 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 217 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 220 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 223 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 233 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 234 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 243 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 265 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 279 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 280 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader.go | 285 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 17 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 18 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 21 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 22 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 27 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 28 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 29 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 36 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 38 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 66 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 76 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 80 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 81 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 82 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 123 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 127 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 130 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 131 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 132 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 134 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 135 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 136 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 138 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 139 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 140 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 146 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 150 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 155 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 156 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 157 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 159 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 161 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 162 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 163 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 165 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 167 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 168 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 169 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 203 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 207 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 217 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 270 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 271 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 297 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 309 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 356 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 358 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 359 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 362 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 365 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 392 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 396 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 402 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 403 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 405 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 418 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 455 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 456 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 459 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 480 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 508 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 513 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 528 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 529 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 563 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_account.go | 564 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 15 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 19 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 25 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 26 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 61 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 66 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 67 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 68 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 91 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 97 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 98 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 133 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 138 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 139 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 140 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 163 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 164 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 165 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 215 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 220 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 221 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 222 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 250 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 251 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 252 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 302 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 307 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 308 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 309 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 336 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 337 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 338 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 340 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 344 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 345 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 346 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 348 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 353 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 354 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 365 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 385 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 386 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 397 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 402 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 427 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 431 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 440 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 495 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 514 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 525 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 527 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 584 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 624 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 630 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 631 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 632 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 638 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 640 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 650 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 663 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 665 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 746 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 786 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 792 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 793 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 796 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 802 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 804 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 814 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 827 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 829 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 900 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 901 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 922 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 927 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 928 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 948 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 949 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 970 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 975 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 976 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 997 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 998 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1004 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1017 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1019 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1024 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1025 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1026 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1037 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1042 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1059 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1060 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_orders.go | 1073 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 11 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 13 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 33 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 34 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 142 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 143 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 154 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 155 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_positions.go | 156 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 1 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 15 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 40 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 41 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 50 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 62 | `hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 100 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/hyperliquid/trader_sync.go | 127 | `Hyperliquid` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 1 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 19 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 21 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 22 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 23 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 26 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 27 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 28 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 30 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 39 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 52 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 53 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 69 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 70 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 77 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 78 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 87 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 88 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 89 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 92 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 93 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 94 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 99 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 105 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 113 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 120 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 121 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 147 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 148 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 181 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 193 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 195 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 198 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 204 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 216 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 218 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 219 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 224 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 234 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 247 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 255 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 264 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 269 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader.go | 293 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 1 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 14 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 15 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 52 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 86 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 87 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 163 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 164 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 167 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_account.go | 194 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 1 | `indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 15 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 45 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 57 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 58 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 59 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 63 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 106 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 118 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 119 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 120 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 123 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 124 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 125 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 129 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 130 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 131 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 136 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 144 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 157 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 158 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 159 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 162 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 163 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 164 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 167 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 168 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 172 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 173 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 178 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 214 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 216 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 223 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 224 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 228 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 229 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 249 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 278 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 294 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/indodax/trader_orders.go | 301 | `Indodax` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 1 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 15 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 16 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 30 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 31 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 36 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 40 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 65 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 78 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 80 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 113 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 139 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 149 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 163 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 164 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 187 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 236 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 254 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 255 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 276 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 279 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 280 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 288 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 296 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 337 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 398 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 402 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 403 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 405 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 407 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 408 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/order_sync.go | 412 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 1 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 20 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 22 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 23 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 24 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 25 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 26 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 27 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 28 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 29 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 30 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 31 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 32 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 33 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 34 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 37 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 38 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 61 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 69 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 70 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 84 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 85 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 91 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 92 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 98 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 104 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 109 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 112 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 116 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 117 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 118 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 150 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 155 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 165 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 166 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 167 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 175 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 182 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 197 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 221 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 229 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 234 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 240 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 241 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 242 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 243 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 244 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 245 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 249 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 252 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 253 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 254 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 265 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 277 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 302 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 331 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 340 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader.go | 345 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_account.go | 1 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_account.go | 11 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_account.go | 20 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_account.go | 41 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_account.go | 48 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 1 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 15 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 43 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 56 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 70 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 98 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 111 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 125 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 129 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 150 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 177 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 205 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 218 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 231 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 258 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 286 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 299 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 312 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 314 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 334 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 364 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 374 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 404 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 414 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 419 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 424 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 428 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 465 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 476 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 479 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 494 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 498 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 511 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 512 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 513 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 518 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 526 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 541 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 557 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 558 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 597 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 605 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 678 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 682 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_orders.go | 740 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_positions.go | 1 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_positions.go | 10 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_positions.go | 19 | `kucoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_positions.go | 54 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/kucoin/trader_positions.go | 110 | `KuCoin` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 1 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 13 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 16 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 38 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 46 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 55 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 91 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 92 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 95 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 97 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 109 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 134 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 153 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 187 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 199 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 207 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 258 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 262 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 285 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 344 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 368 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/account.go | 453 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 1 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 14 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 17 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 26 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 34 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 54 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 117 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 151 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/order_sync.go | 161 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 1 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 12 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 22 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 33 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 44 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 55 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 76 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 90 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 96 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 160 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 161 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 167 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 168 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 199 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 223 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 244 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 246 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 261 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 286 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 318 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 324 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/orders.go | 339 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 1 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 16 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 17 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 22 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 35 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 51 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 60 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 63 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 96 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 98 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 102 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 103 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 104 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 105 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 108 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 109 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 110 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 120 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 123 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 131 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 168 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 170 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 176 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 198 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 201 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 220 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 226 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 234 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 243 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 272 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 340 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 347 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 350 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 391 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 396 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 401 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 450 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 474 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 493 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 498 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trader.go | 540 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 1 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 15 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 25 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 49 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 66 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 90 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 122 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 136 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 166 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 180 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 211 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 298 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 308 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 351 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 353 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 413 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 415 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 430 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 469 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 470 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 474 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 475 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 476 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 477 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 483 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 554 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 593 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 597 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 599 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 621 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 674 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 685 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 740 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 794 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 837 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 861 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 869 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 875 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 923 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 937 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/trading.go | 961 | `LIGHTER` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 1 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 18 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 27 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 40 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 52 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 53 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 72 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 79 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 80 | `lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/lighter/types.go | 117 | `Lighter` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 1 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 16 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 17 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 35 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 36 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 41 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 45 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 57 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 76 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 84 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 95 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 125 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 135 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 149 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 152 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 153 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 161 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 169 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 245 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 258 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 271 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 275 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 276 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 278 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 280 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 281 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/order_sync.go | 285 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 1 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 20 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 22 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 23 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 24 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 25 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 26 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 27 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 28 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 29 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 30 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 31 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 32 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 33 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 34 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 35 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 38 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 39 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 64 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 72 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 73 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 84 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 85 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 91 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 92 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 97 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 98 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 103 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 104 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 111 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 112 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 114 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 120 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 127 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 132 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 139 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 143 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 148 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 156 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 157 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 172 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 179 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 184 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 188 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 194 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 198 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 199 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 207 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 221 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 245 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 246 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 252 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 253 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 256 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 259 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 260 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 261 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 262 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 263 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 264 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 267 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 268 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 269 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 278 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 284 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader.go | 290 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 1 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 14 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 19 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 24 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 25 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 54 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 55 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 57 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 58 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 59 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 67 | `Wallet` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 68 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 69 | `usdt` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 72 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 86 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 89 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 94 | `Binance` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 96 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 100 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 109 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 122 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 137 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 139 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 166 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 167 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 168 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 191 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 212 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 220 | `USDT` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 229 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_account.go | 244 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 1 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 13 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 30 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 35 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 39 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 53 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 54 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 57 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 81 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 92 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 109 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 114 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 118 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 132 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 133 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 136 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 160 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 171 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 191 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 193 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 204 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 211 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 214 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 228 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 237 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 238 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 246 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 269 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 282 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 302 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 304 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 312 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 320 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 329 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 339 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 348 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 349 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 357 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 359 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 379 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 383 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 396 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 428 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 431 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 441 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 473 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 476 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 486 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 491 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 496 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 500 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 524 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 540 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 544 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 565 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 579 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 584 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 626 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 652 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 657 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 662 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 665 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 683 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 706 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 707 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 710 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 792 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 798 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 810 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 836 | `Okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 837 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 845 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 847 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 868 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 871 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 888 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 901 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_orders.go | 907 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 1 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 12 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 17 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 22 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 23 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 47 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 50 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 64 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 71 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 82 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 121 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 129 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 141 | `okx` | DELETE | CR-A | whole-file delete (C4) |
| trader/okx/trader_positions.go | 174 | `OKX` | DELETE | CR-A | whole-file delete (C4) |
| wallet/balance_cache.go | 1 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| wallet/balance_cache.go | 22 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/balance_cache.go | 25 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/balance_cache.go | 50 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 1 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 2 | `wallet` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 17 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 18 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 21 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 24 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 25 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 28 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 30 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 31 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 32 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 39 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 49 | `USDC` | DELETE | CR-A | whole-file delete (C4) |
| wallet/usdc.go | 94 | `USDC` | DELETE | CR-A | whole-file delete (C4) |

## C7 KEEP (settled — GO item 2 declined 2026-10-01)
provider/alpaca, provider/twelvedata and their live branches are out of scope (GO item 2 declined 2026-10-01); zero hits under the base regex, so no rows. If a stock-provider line ever matches a future regex, its disposition is KEEP with this reason.
