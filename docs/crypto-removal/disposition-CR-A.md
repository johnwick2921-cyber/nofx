# CR-A disposition table — brokers + payments + providers (crypto removal, part CR-A)

base sha (branch point): db412e61c
integrator tip generated against: a4200b8fc
generated: 2026-10-01T11:15:36-05:00 (branch crypto-removal-integration @ a4200b8fc — provenance: branch + sha only, never a filesystem path)
sweep regex (exported programmatically from branding/no_crypto.go SweepRegexLiteral — the guard's amended literal, never hand-typed):
  `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续|btc|ethusdt|\beth\b|altcoin|ethereum`

| path | line | token | disposition | owner | reason |
|---|---|---|---|---|---|
| agent/tools.go | 848 | `Hyperliquid` | CUT | CR-A | safe-tool DEX/wallet fields — cut deferred until DS-107 (skill surface) + DS-108 (model surface) land their consumer cuts |
| agent/tools.go | 852 | `Wallet` | CUT | CR-A | safe-tool DEX/wallet fields — cut deferred until DS-107 (skill surface) + DS-108 (model surface) land their consumer cuts |
| agent/tools.go | 868 | `Wallet` | CUT | CR-A | safe-tool DEX/wallet fields — cut deferred until DS-107 (skill surface) + DS-108 (model surface) land their consumer cuts |
| agent/tools.go | 869 | `USDC` | CUT | CR-A | safe-tool DEX/wallet fields — cut deferred until DS-107 (skill surface) + DS-108 (model surface) land their consumer cuts |
| agent/tools.go | 952 | `Hyperliquid` | CUT | CR-A | safe-tool DEX/wallet fields — cut deferred until DS-107 (skill surface) + DS-108 (model surface) land their consumer cuts |
| agent/tools.go | 956 | `Wallet` | CUT | CR-A | safe-tool DEX/wallet fields — cut deferred until DS-107 (skill surface) + DS-108 (model surface) land their consumer cuts |
| agent/tools.go | 991 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — totalWalletBalance/wallet_balance key family in the agent balance extractor (CTO ruling, byte-identical) |
| agent/tools.go | 2199 | `btc` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2200 | `altcoin` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2202 | `AI500` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2222 | `BTC` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2223 | `Altcoin` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2225 | `AI500` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2294 | `BTC` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2295 | `Altcoin` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| agent/tools.go | 2297 | `AI500` | CUT | CR-A | store.Trader legacy crypto columns (C1: stored rows must load) — cut deferred until DS-103 drops the fields |
| api/exchange_account_state.go | 200 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — balance-key family in the account-state scan (CTO ruling, byte-identical) |
| api/exchange_account_state.go | 250 | `Wallet` | KEEP | CR-A | NT8 account-snapshot shape — balance-key family in the account-state scan (CTO ruling, byte-identical) |
| market/api_client.go | 15 | `binance` | CUT | CR-A | APIClient binance base URL — live call sites market/data.go:419,464 are DS-101's residue |
| store/ai_charge.go | 152 | `USDC` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_charge.go | 153 | `usdc` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_charge.go | 160 | `usdc` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_charge.go | 161 | `usdc` | KEEP | CR-A | KEEP — EstimateRunway stays (daily-cost/runway estimate keeps calling it; usdcBalance naming byte-identical per CTO ruling) |
| store/ai_model.go | 416 | `Claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 417 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 418 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 419 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 421 | `Claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 427 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 428 | `wallet` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 429 | `wallet` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 430 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 432 | `wallet` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 441 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 446 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/ai_model.go | 456 | `claw402` | CUT | CR-A | ResolveClaw402WalletKey — sole caller manager/trader_manager.go:795 (DS-103) |
| store/exchange.go | 32 | `Hyperliquid` | CUT | CR-A | store.Exchange crypto columns — C1 stored rows must load; column drop pending CTO ruling |
| store/exchange.go | 33 | `Hyperliquid` | CUT | CR-A | store.Exchange crypto columns — C1 stored rows must load; column drop pending CTO ruling |
| store/exchange.go | 37 | `Wallet` | CUT | CR-A | store.Exchange crypto columns — C1 stored rows must load; column drop pending CTO ruling |
| store/exchange.go | 125 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 139 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 149 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 216 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 218 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 250 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 251 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 255 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 274 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 275 | `Wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 283 | `hyperliquid` | CUT | CR-A | Create/Update/Default signatures carry crypto params — pending DS-103's trader_manager cut + the C1 ruling |
| store/exchange.go | 284 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 287 | `wallet` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 364 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 367 | `binance` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 369 | `hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
| store/exchange.go | 385 | `Hyperliquid` | CUT | CR-A | CUT — crypto line (unclassified: review) |
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
