# CR-A disposition table — brokers + payments + providers (crypto removal, part CR-A)

base sha (branch point): db412e61c
integrator tip generated against: 9bb6caddf
generated: 2026-10-01T17:35:45-05:00 (branch crypto-removal-integration @ 9bb6caddf — provenance: branch + sha only, never a filesystem path)
sweep regex (exported programmatically from branding/no_crypto.go SweepRegexLiteral — the guard's amended literal, never hand-typed):
  `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续|btc|ethusdt|\beth\b|altcoin|ethereum`

| path | line | token | disposition | owner | reason |
|---|---|---|---|---|---|
| agent/tools.go | 814 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 818 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 834 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 835 | `USDC` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 918 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 922 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 957 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2165 | `btc` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2166 | `altcoin` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2168 | `AI500` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2188 | `BTC` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2189 | `Altcoin` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2191 | `AI500` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2260 | `BTC` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2261 | `Altcoin` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| agent/tools.go | 2263 | `AI500` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| api/exchange_account_state.go | 200 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — balance-key family in the account-state scan (CTO ruling, byte-identical) |
| api/exchange_account_state.go | 250 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — balance-key family in the account-state scan (CTO ruling, byte-identical) |
| store/ai_charge.go | 152 | `USDC` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_charge.go | 153 | `usdc` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_charge.go | 160 | `usdc` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_charge.go | 161 | `usdc` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/exchange.go | 32 | `Hyperliquid` | CUT | CR-A | store.Exchange crypto columns — C1 stored rows must load; column drop pending CTO ruling |
| store/exchange.go | 33 | `Hyperliquid` | CUT | CR-A | store.Exchange crypto columns — C1 stored rows must load; column drop pending CTO ruling |
| store/exchange.go | 37 | `Wallet` | CUT | CR-A | store.Exchange crypto columns — C1 stored rows must load; column drop pending CTO ruling |
| store/exchange.go | 142 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 156 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 166 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 233 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 235 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 267 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 268 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 272 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 291 | `hyperliquid` | CUT | CR-A | Create/Update/Default signatures carry crypto params — pending DS-103's trader_manager cut + the C1 ruling |
| store/exchange.go | 292 | `Wallet` | CUT | CR-A | Create/Update/Default signatures carry crypto params — pending DS-103's trader_manager cut + the C1 ruling |
| store/exchange.go | 300 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 301 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 304 | `wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 381 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 384 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 386 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 402 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/visibility.go | 53 | `Hyperliquid` | CUT | CR-A | MissingRequiredExchangeCredentialFields crypto params/checks — interlocked with store/exchange.go signature |
| store/visibility.go | 57 | `Wallet` | CUT | CR-A | MissingRequiredExchangeCredentialFields crypto params/checks — interlocked with store/exchange.go signature |
| trader/auto_trader.go | 256 | `Binance` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 257 | `Binance` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 258 | `Binance` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 260 | `Bybit` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 261 | `Bybit` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 262 | `Bybit` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 264 | `OKX` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 265 | `OKX` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 266 | `OKX` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 267 | `OKX` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 269 | `Bitget` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 270 | `Bitget` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 271 | `Bitget` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 272 | `Bitget` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 278 | `KuCoin` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 279 | `KuCoin` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 280 | `KuCoin` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 281 | `KuCoin` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 283 | `Indodax` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 284 | `Indodax` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 285 | `Indodax` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 292 | `Hyperliquid` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 293 | `Hyperliquid` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 294 | `Hyperliquid` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 295 | `Hyperliquid` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 296 | `Hyperliquid` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 298 | `Aster` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 299 | `Aster` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 300 | `Aster` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 301 | `Aster` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 303 | `LIGHTER` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 304 | `Wallet` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 305 | `LIGHTER` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 306 | `LIGHTER` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 307 | `LIGHTER` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 319 | `Claw402` | CUT | CR-A | AutoTraderConfig per-broker credential fields — sole populator is manager/trader_manager.go (DS-103) |
| trader/auto_trader.go | 771 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — balance-key family in the auto-fetched balance scan (CTO ruling, byte-identical) |
