# Order truth — plan, ledger, dashboard, NinjaTrader

Owner: hoang. Read-only audit. Branch: `docs/order-truth-0907`.

**Checkpoint, not final conformance verdict.** At 2026-09-07 20:53:09 CT the authenticated live API reported no positions and no working broker orders: `/api/positions` returned `[]`; `/api/cutover-gate` legs 1–4 independently reported zero ledger positions, zero API positions, zero NT8 snapshot positions, and zero working broker orders (book age 6 seconds). No live unprotected position was observed in that sample. This is a dated sample, not an enduring safety certificate.

The running Go executable is `5457ac5accd97c3519bf6d16ead147a0db2ab0d0`, `vcs.modified=false`, measured from `/proc/3058590/exe` at 20:49:30 CT. `/api/health` reported `5457ac5accd9`; service start was 20:39:39 CT. The received AddOn handshake at that start and stored `nt8_order_snapshots.id=10764` report **`2026-09-03-f12`**, below `2026-09-07-h1`. Source changes must not be described as deployed C# behavior.

Worktree base: dev tip `7ddf73cd99b9cd7fece7791c33a15bc56e33192e` at acceptance. The deploy checkout was porcelain-clean on `dev` at `af3472ee7224227dba735d48550e2848f44c982b`. The worktree is isolated and locked at `/tmp/nofx-order-truth-0907`; no main-tree lock was acquired or changed. Source differences from the running revision are documentation, RELEASE and the guide revision marker; executable order-path sources match.

Confirmed so far:

- `plans.rowid=259` (2026-09-06 ASIA v2, S1) has `target_chain=[29545,29566.02,29587.75]` and `arm.target=29575.48`; `armed_orders.id=107` carries the latter. `trader/ninjatrader/tcp_trader.go:447` rounds it to 29575.50. NT8 `log.20260906.00006.txt:1325` records that target reserved for the fill. The card reads `target_chain` at `web/src/components/plan/ScenarioList.tsx:201`.
- The latest real filled entry is arm `111`, signal `aa07e583-6df3-4148-9d57-667630f1b155`, associated with position `592`. NT8 `log.20260906.00006.txt:4907` records its fill at **23:35:06.939 CT**. The older “23:35:22 / 102 seconds” premise requires correction against this timestamp.
- The latest broker-confirmed cancelled entry is arm `109`, signal `c5d8bdde-22a2-494c-8943-e9c6a0999324`. NT8 `log.20260906.00006.txt:3801` records `Cancelled` at 21:58:17.508 CT; the ledger reconciliation log is `log_events.id=30724` at 21:58:57.013 CT.
- The latest retained order-gate refusal before the evidence cutoff is `log_events.id=31082`, 23:37:02.286 CT on 09-06: the S1 reauthorization was refused because position `592` was already open. A refused candidate does not acquire a new broker order ID.
- C6 needs a transport distinction: HTTP `/api/plan/today.armed` omits stop and target (`api/handler_plan.go:205`); the TCP `signal` carries `stop_loss` and `take_profit` (`trader/ninjatrader/tcp_trader.go:453`).
- The desk has accepted-stop, accepted-target and drift renderers (`trader/desk_facts.go:401`, `:421`, `:446`), so “no surface distinguishes intended and accepted” is too broad. Whether those values can establish *current* protection is being audited.

Evidence gaps are explicit: no historical DOM capture from the order moments has been established; local headless Chromium cannot start because `libnspr4.so` is absent. No packages or environment settings were changed. Current API payloads and production renderers are available; a TypeScript declaration alone is not treated as rendering proof.

The final report will include the complete field table, three traces, state vocabulary, placement/cancel census, measured timing, visibility gaps, ranked proposals and per-file freshness ledger. No code, configuration, database, AddOn or trading action is authorized by this report.
