# CR-A disposition table — base db412e61c — integrator tip db412e61c — brokers + payments + providers

base sha (branch point): db412e61c
integrator tip generated against: db412e61c
generated: 2026-10-01T00:30:15-05:00 (worktree /home/hoang/nofx-104-cra, branch crypto-removal-cr-a)
sweep regex (exported literal, plan v10 line 108):
  `binance|bybit|okx|bitget|kucoin|gate\.io|gateio|indodax|hyperliquid|\baster\b|asterdex|\blighter\b|coinank|usdc|usdt|x402|claw402|blockrun|wallet|ai500|hyper_all|hyper_main|oi_top|oi_low|netflow|quant\b|price ranking|"mixed"|币安|欧易|火币|U本位|永续`
scope: CR-A paths per plan v10 build-split; agent/agent.go + agent/tools.go are shared with CR-B — their CR-B-owned hits are listed separately for union coverage.
## Whole-file DELETE (C4)
DELETE trader/aster/trader.go — whole dir under trader/aster (broker/payment/provider code)
DELETE trader/aster/trader_account.go — whole dir under trader/aster (broker/payment/provider code)
DELETE trader/aster/trader_orders.go — whole dir under trader/aster (broker/payment/provider code)
DELETE trader/aster/trader_positions.go — whole dir under trader/aster (broker/payment/provider code)
DELETE trader/aster/trader_sync.go — whole dir under trader/aster (broker/payment/provider code)
DELETE trader/binance/futures.go — whole dir under trader/binance (broker/payment/provider code)
DELETE trader/binance/futures_account.go — whole dir under trader/binance (broker/payment/provider code)
DELETE trader/binance/futures_orders.go — whole dir under trader/binance (broker/payment/provider code)
DELETE trader/binance/futures_positions.go — whole dir under trader/binance (broker/payment/provider code)
DELETE trader/binance/order_sync.go — whole dir under trader/binance (broker/payment/provider code)
DELETE trader/bitget/order_sync.go — whole dir under trader/bitget (broker/payment/provider code)
DELETE trader/bitget/trader.go — whole dir under trader/bitget (broker/payment/provider code)
DELETE trader/bitget/trader_account.go — whole dir under trader/bitget (broker/payment/provider code)
DELETE trader/bitget/trader_orders.go — whole dir under trader/bitget (broker/payment/provider code)
DELETE trader/bitget/trader_positions.go — whole dir under trader/bitget (broker/payment/provider code)
DELETE trader/bybit/order_sync.go — whole dir under trader/bybit (broker/payment/provider code)
DELETE trader/bybit/trader.go — whole dir under trader/bybit (broker/payment/provider code)
DELETE trader/bybit/trader_account.go — whole dir under trader/bybit (broker/payment/provider code)
DELETE trader/bybit/trader_orders.go — whole dir under trader/bybit (broker/payment/provider code)
DELETE trader/bybit/trader_positions.go — whole dir under trader/bybit (broker/payment/provider code)
DELETE trader/gate/order_sync.go — whole dir under trader/gate (broker/payment/provider code)
DELETE trader/gate/trader.go — whole dir under trader/gate (broker/payment/provider code)
DELETE trader/gate/trader_account.go — whole dir under trader/gate (broker/payment/provider code)
DELETE trader/gate/trader_orders.go — whole dir under trader/gate (broker/payment/provider code)
DELETE trader/hyperliquid/order_sync.go — whole dir under trader/hyperliquid (broker/payment/provider code)
DELETE trader/hyperliquid/trader.go — whole dir under trader/hyperliquid (broker/payment/provider code)
DELETE trader/hyperliquid/trader_account.go — whole dir under trader/hyperliquid (broker/payment/provider code)
DELETE trader/hyperliquid/trader_orders.go — whole dir under trader/hyperliquid (broker/payment/provider code)
DELETE trader/hyperliquid/trader_positions.go — whole dir under trader/hyperliquid (broker/payment/provider code)
DELETE trader/hyperliquid/trader_sync.go — whole dir under trader/hyperliquid (broker/payment/provider code)
DELETE trader/indodax/trader.go — whole dir under trader/indodax (broker/payment/provider code)
DELETE trader/indodax/trader_account.go — whole dir under trader/indodax (broker/payment/provider code)
DELETE trader/indodax/trader_orders.go — whole dir under trader/indodax (broker/payment/provider code)
DELETE trader/kucoin/order_sync.go — whole dir under trader/kucoin (broker/payment/provider code)
DELETE trader/kucoin/trader.go — whole dir under trader/kucoin (broker/payment/provider code)
DELETE trader/kucoin/trader_account.go — whole dir under trader/kucoin (broker/payment/provider code)
DELETE trader/kucoin/trader_orders.go — whole dir under trader/kucoin (broker/payment/provider code)
DELETE trader/kucoin/trader_positions.go — whole dir under trader/kucoin (broker/payment/provider code)
DELETE trader/lighter/account.go — whole dir under trader/lighter (broker/payment/provider code)
DELETE trader/lighter/order_sync.go — whole dir under trader/lighter (broker/payment/provider code)
DELETE trader/lighter/orders.go — whole dir under trader/lighter (broker/payment/provider code)
DELETE trader/lighter/trader.go — whole dir under trader/lighter (broker/payment/provider code)
DELETE trader/lighter/trading.go — whole dir under trader/lighter (broker/payment/provider code)
DELETE trader/lighter/types.go — whole dir under trader/lighter (broker/payment/provider code)
DELETE trader/okx/order_sync.go — whole dir under trader/okx (broker/payment/provider code)
DELETE trader/okx/trader.go — whole dir under trader/okx (broker/payment/provider code)
DELETE trader/okx/trader_account.go — whole dir under trader/okx (broker/payment/provider code)
DELETE trader/okx/trader_orders.go — whole dir under trader/okx (broker/payment/provider code)
DELETE trader/okx/trader_positions.go — whole dir under trader/okx (broker/payment/provider code)
DELETE provider/coinank/base_coin.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_api/base_coin.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_api/depth_ws.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_api/kline.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_api/kline_ws.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/exchange.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/instrument_agg_sort_by.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/interval.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/product_type.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/side.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/sort_type.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_enum/url.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/coinank_http.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/instrument_agg_rank.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/instruments.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/kline.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/liquidation.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/net_positions.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/coinank/open_interest.go — whole dir under provider/coinank (broker/payment/provider code)
DELETE provider/hyperliquid/coins.go — whole dir under provider/hyperliquid (broker/payment/provider code)
DELETE provider/hyperliquid/kline.go — whole dir under provider/hyperliquid (broker/payment/provider code)
DELETE mcp/payment/claw402.go — whole dir under mcp/payment (broker/payment/provider code)
DELETE mcp/payment/x402.go — whole dir under mcp/payment (broker/payment/provider code)
DELETE wallet/balance_cache.go — whole dir under wallet/ (broker/payment/provider code)
DELETE wallet/usdc.go — whole dir under wallet/ (broker/payment/provider code)
DELETE api/handler_wallet.go — C4 named delete (kernel/price_change_render_test.go goes with it, test file not a row)
DELETE market/api_client.go — C4 named delete (kernel/price_change_render_test.go goes with it, test file not a row)
DELETE market/historical.go — C4 named delete (kernel/price_change_render_test.go goes with it, test file not a row)
DELETE api/handler_competition.go — C4 named delete (kernel/price_change_render_test.go goes with it, test file not a row)
DELETE kernel/testdata/golden/user_prompt_crypto_change_na.txt — C4 named delete (kernel/price_change_render_test.go goes with it, test file not a row)
CUT go.mod — drop go-binance, bybit.go.api, gateapi-go (+antihax/optional), lighter-go (+poseidon), go-hyperliquid, go-ethereum (+indirect cascade), then go mod tidy + verify
CUT go.sum — the dropped SDK module lines; go mod tidy rewrites it
## KEEP rows (one per file + token)
KEEP store/ai_charge.go:157 (token: see line) — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP store/ai_charge.go:158 (token: see line) — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP store/ai_charge.go:165 (token: see line) — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP store/ai_charge.go:166 (token: see line) — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP trader/auto_trader_loop.go:1005 (token: see line) — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1010 (token: see line) — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1011 (token: see line) — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1024 (token: see line) — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1025 (token: see line) — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP agent/tools.go:1110 (token: see line) — NT8 account-snapshot shape — same totalWalletBalance key family, the agent balance extractor (CTO ruling)
KEEP api/exchange_account_state.go:210 (token: see line) — NT8 account-snapshot shape — totalWalletBalance/wallet_balance keys in the balance scan (CTO ruling)
KEEP api/exchange_account_state.go:294 (token: see line) — NT8 account-snapshot shape — totalWalletBalance/wallet_balance keys in the balance scan (CTO ruling)
KEEP agent/trade.go:21 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:22 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:327 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:328 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:346 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:357 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:400 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:413 (token: see line) — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP trader/auto_trader.go:843 (token: see line) — NT8 account-snapshot shape — the totalWalletBalance key in the auto-fetched balance scan (CTO ruling)
## CUT rows (grouped ranges; every line listed)
CUT agent/agent.go [29,64-65,169,177-178,226,236,238,249-250,277-278,285-286,289,296,298,326,362,459,508] — payment lines (wallet import, USDC balance ranking, claw402 provider cases + URL, wallet-balance question gate) go; the native provider path stays
CR-B agent/agent.go:71 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:745 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:758 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:761 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:762 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:781 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:783 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:907 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:909 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/agent.go:985 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CUT agent/tools.go [24-27,29-32,34,41,1065-1072,1075-1080,1082,1084,1086-1087,1092-1094,1146,1149-1152,1608-1609,1619-1621,1653,1657,1672,1677,2390,2413,2485,3013-3014,3050-3051,3081,3093,3102,3107,3248,3250,3308-3309,3328,3610-3612,3643,3645,3727,3734-3735,3756,3765-3767,3773,3791] — broker factory arms + binanceFuturesAPIBaseURL + the model-list USDC block go; the ninjatrader factory arm stays
KEEP agent/tools.go:1110 — NT8 account-snapshot shape — same totalWalletBalance key family, the agent balance extractor (CTO ruling)
CR-B agent/tools.go:64 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:76 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:79 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:284 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:285 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:343 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:345 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:380 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:392 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:396 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:413 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:425 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:427 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:428 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:429 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:430 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:431 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:432 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:433 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:434 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:435 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:534 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:550 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:551 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:555 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:695 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:744 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:761 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:786 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:821 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:860 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:933 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:937 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:953 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:954 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1037 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1041 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1472 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1473 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1477 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1505 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1506 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1518 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1522 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1540 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1545 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1599 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1600 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1601 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CR-B agent/tools.go:1607 — owned by CR-B (agent tools/skills non-payment); listed for union coverage
CUT agent/trade.go [60,362,486-487] — crypto symbol example/comment and the USDT-suffix stripping go; the chat-entry admission chain stays
KEEP agent/trade.go:21 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:22 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:327 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:328 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:346 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:357 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:400 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
KEEP agent/trade.go:413 — futures-active chat-entry notional caps (tradeLargeOrderNotionalUSDT/tradeHardMaxOrderNotionalUSDT + hard-cap/min-size messages) — the MNQ path uses the same admission chain; USDT here is legacy naming for the value currency
CUT api/exchange_account_state.go [14-17,19-22,24,248-255,258-263,265,267,269-270,275-277,340-341,343,353,357] — broker imports + factory arms + USDC/USDT quote-asset cases go; the ninjatrader arm stays
KEEP api/exchange_account_state.go:210 — NT8 account-snapshot shape — totalWalletBalance/wallet_balance keys in the balance scan (CTO ruling)
KEEP api/exchange_account_state.go:294 — NT8 account-snapshot shape — totalWalletBalance/wallet_balance keys in the balance scan (CTO ruling)
CUT api/handler_ai_model.go [all 11 hits] — wallet import, wallet-address/balance columns and the blockrun/claw402 list entries go; native rows stay
CUT api/handler_competition.go [154,418-420,432] — crypto lines cut per C4/C5; the futures half stays
CUT api/handler_exchange.go [31,40,42-44,65,69,83,85-86,90,103,110-111,115,244-246,256-258,277,281,298,321,397-398,410,414,429,431,496-503] — crypto credential fields, effective-merge, venue catalog crypto rows and create/update arms go; ninjatrader-only for NEW; stored rows load (C1)
CUT api/handler_klines.go [all 99 hits] — coinank + hyperliquid branches go; the NT8 BarCache path and the SVP endpoint stay (C6)
CUT api/handler_onboarding.go [all 62 hits] — the claw402/wallet onboarding sections go (routes are already refused on the futures build; they get deleted from api/server.go)
CUT api/handler_trader.go [all 22 hits] — crypto credential cases, accepted-type list and refusal strings go; ninjatrader-only for NEW; stored rows still load (C1)
CUT api/handler_trader_status.go [all 45 hits] — broker factory arms and the Lighter order-recording path go; the ninjatrader status path stays
CUT api/handler_wallet.go [7,15,19,22-23,29-30,32,43,51,60,70,80-81,83-84,86,89-90,96,101-102,105,112,118,120] — crypto lines cut per C4/C5; the futures half stays
CUT api/server.go [all 19 hits] — wallet + claw402-onboarding routes and crypto venue/coin-source doc lines go; the futures routes stay
CUT api/strategy.go [all 1 hits] — the claw402 case of resolveStrategyDataWalletKey goes; kernel.NewStrategyEngine loses the claw402WalletKey param
CUT api/utils.go [all 6 hits] — hyperliquid/lighter wallet-addr fields go with the store.Exchange field drop
CUT go.mod [6,9-10,12,25] — crypto lines cut per C4/C5; the futures half stays
CUT hook/hooks.go [all 1 hits] — the NEW_BINANCE_TRADER registration goes with the broker dirs
CUT hook/trader_hook.go [all 5 hits] — per-broker result types go in the SAME commit as the broker dirs; the ninjatrader hook result stays
CUT kernel/testdata/golden/user_prompt_crypto_change_na.txt [8,10] — crypto lines cut per C4/C5; the futures half stays
CUT main.go [all 3 hits] — comments reworded — CoinAnk naming must be gone for the host census
CUT manager/trader_manager.go [all 46 hits] — the crypto credential map + Claw402WalletKey family go; the ninjatrader case + NT data-dir wiring stay
CUT market/api_client.go [15] — crypto lines cut per C4/C5; the futures half stays
CUT market/historical.go [12-13,36,44,60,98] — crypto lines cut per C4/C5; the futures half stays
CUT mcp/client.go [all 4 hits] — the claw402 payment channel goes; the native provider path stays
CUT mcp/providers.go [all 1 hits] — ProviderClaw402 goes
CUT store/ai_charge.go [152-154] — crypto lines cut per C4/C5; the futures half stays
KEEP store/ai_charge.go:157 — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP store/ai_charge.go:158 — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP store/ai_charge.go:165 — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
KEEP store/ai_charge.go:166 — EstimateRunway stays (daily-cost/runway estimate; the auto_trader_loop daily-cost line keeps calling it); its usdcBalance param name stays byte-identical
CUT store/ai_model.go [all 18 hits] — FindOrphanClaw402 + ResolveClaw402WalletKey go; ai_models columns STAY (C1)
CUT store/exchange.go [all 40 hits] — crypto credential fields go (columns STAY), migrateToMultiAccount deleted wholesale, crypto display-name/factory arms go; ninjatrader arm + backfillDefaultAccountName stay
CUT store/trader.go [all 4 hits] — UseAI500/UseOITop fields go; columns use_coin_pool/use_oi_top STAY (C1)
CUT store/visibility.go [all 10 hits] — crypto cases + hyperliquid/aster/lighter wallet fields go; PR #188 D1-D3 ported — cleanup SKIPS unsupported rows
CUT telegram/bot.go [all 7 hits] — the USDC/claw402 payment gate goes
CUT telemetry/experience.go [all 1 hits] — the claw402 channel value goes; comment reworded
CUT trader/auto_trader.go [21-24,26-29,31-32,267,270-272,274-276,278-281,283-286,292-295,297-299,306-310,312-315,317-321,333,376,484,692,694,707,714,733-744,746,748-753,755,757-759,761,763-764,766-767,770-772,775,778,780-783,853,1001,1009-1011,1013,1017-1021,1025-1029,1033-1037,1041-1045,1049-1051,1053,1057-1061,1073-1077,1330,1332,1336,1338,1341,1343-1344,1346-1347,1349,1358,1362,1365,1374-1375] — crypto credential struct, broker factory + order-sync arms, claw402 pre-launch block go; the ninjatrader arm stays
KEEP trader/auto_trader.go:843 — NT8 account-snapshot shape — the totalWalletBalance key in the auto-fetched balance scan (CTO ruling)
CUT trader/auto_trader_loop.go [17,350-352,522,709,1041,1088,1242-1243,1252-1253,1255,1260,1263,1270] — wallet import, claw402 balance check, USDT equity logs and crypto comments go; the NT8 account-snapshot keys stay (KEEP rows below)
KEEP trader/auto_trader_loop.go:1005 — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1010 — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1011 — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1024 — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
KEEP trader/auto_trader_loop.go:1025 — NT8 account-snapshot shape — the totalWalletBalance key mirrors tcp_trader.go's CashValue snapshot the engine + account-state parity tests read (CTO ruling: renaming needs its own wave)
## C7 KEEP (settled — GO item 2 declined 2026-10-01)

KEEP provider/alpaca, provider/twelvedata and their live branches (api/handler_klines.go alpaca/twelvedata arms, agent search_stock, agent/trade.go isAlpaca, config.go keys, handler_exchange venue rows) — out of scope, GO item 2 declined 2026-10-01. The sweep regex stays the BASE literal (no |alpaca|twelvedata|sina); these paths have ZERO hits under it (verified by the generator sweep), so no rows are emitted. If a stock-provider line ever matches a future regex, its disposition is KEEP with this same reason.
