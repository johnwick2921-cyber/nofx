# Plan 1 — Shipped 2026-05-22

## Validated end-to-end via NT Playback

Manual CSV signal → NT ClaudeTrader → SIM101 fill against playback tape 
→ SL+TP attach → TP hit → position closes → fill row written back.

Two complete round trips:
- 05/22/2026 19:52:04 LONG @ 29706.25
- 05/22/2026 19:53:07 LONG @ 29367.00 (TP target 29396 hit)

## Built and committed

- provider/databento (Historical OHLCV + symbol resolver)
- provider/ninjatrader (CSV writer + fill tailer)  
- trader/ninjatrader (19-method Trader impl, compile-time guarded)
- kernel/engine_prompt_futures.go (futures-mode AI prompt)
- cmd/nq_smoke (one-shot end-to-end runner)
- config + .env.example (Databento + NT data dir + TRADING_MODE)
- Broker switch in trader/auto_trader.go
- Removed: Data, Strategy Market, Leaderboard web pages

11 task commits on branch `nq-databento-ninjatrader-plan`.

## Remaining acceptance step (deferred to Sunday 5pm CT market open)

- [ ] John pastes Databento API key into .env (Step 0.9)
- [ ] go run ./cmd/nq_smoke executes one full cycle against live MNQ
- [ ] trades_taken.csv captures the cmd/nq_smoke fill

## Intentionally NOT done (out of scope, defer until triggered)

- Plan 1.5: NT8 AddOn / TCP bridge (manual close, real balance read, 
  cancel/modify mid-trade, sub-second latency)
- Plan 2: CME calendar, kill-switches, contract roll detection, 
  dead-man-switch, daily loss cap

Plan 1.5 trigger conditions: AI brain validated for several sessions OR
going to funded broker OR CSV race condition fires in SIM.

Plan 2 trigger condition: ready for live (real money) deployment.

## What John should do next

This weekend: nothing. Close the laptop.

Sunday 5pm CT or after: paste Databento key in .env, run 
`go run ./cmd/nq_smoke` against live data, observe fill captured.

Then: let it run in SIM for several sessions before deciding what 
to build next.
