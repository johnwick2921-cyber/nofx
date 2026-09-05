# VETERAN REVIEW — SECTION 04: MONITORING

*Owner: hoang · 2026-09-05 (Saturday, CME closed) · session vet-04-0905 · branch docs/vet-04-0905 · READ-ONLY. First-person analytical lens; I am an AI reviewer, not a human veteran and make no claim to a trading career. Evidence tiers: **[A]** directly verified · **[B]** inferred from strong evidence · **[C]** speculation. Market beliefs: **[R]** study · **[T]** own-tape number, n · **[I]** analytical judgment, untested here. All times CT unless a column says otherwise.*

## ONE-PAGE SUMMARY

I would not use the green system header as permission to trade. It answers whether the HTTP process responds, not whether the feed, broker connection, or risk controls are healthy (`api/server.go:626-632`; Q1–Q3). My priorities are feed freshness, explicit protection status, and a review page that cannot silently substitute all-time equity for today's corrected ledger P&L.

**Three problems.** (1) The September 3 record contains a 113.86-minute log gap (log ids 25754→25755), followed by about 48.3 minutes from Go boot to first observed bars. Windows events support an unclean host restart; they do not establish its root cause or who subsequently launched NT8 (`q08`, `q47`). (2) The “no alert” premise is false: feed-down P0 ids 629 and 654 were emitted on September 3 and 4 and remained unacknowledged at audit. Whether the owner saw them is unknown. The alert UI already has a P0 banner and toast; backlog is a usability concern, not proof of invisibility (`web/src/components/plan/AlertCenter.tsx:161-169,196-197`; `q08`, `q45a`). (3) The risk endpoint reports an env-derived limit/armed flag while the bound Studio master and daily-loss flag are OFF (`api/handler_risk.go:178-233`; `q26b`). The positions table omits stop, target and open-risk columns (`web/src/pages/TraderDashboardPage.tsx:880-1020`).

**Three opportunities.** First, use the existing dashboard, P0 banner, boot log and NT8 log more effectively: show feed/link/loop ages, distinguish stale from healthy, and resolve alerts without falsifying acknowledgement. Do not add Telegram push, desktop popups, or another channel. Second, add the Q2 desk strip with explicit OFF/UNKNOWN labels and per-field timestamps. Third, replace premature session digests with a 15:00 CT page whose queries use close-time corrected P&L and a fixed cutoff. The digest timing discrepancy is directly visible in ids 55/56/59/60/63/64 (`q33`; `trader/auto_trader_planner.go:2473-2512`).

**What the corrected 15:00 page says.** September 3: one eligible close, id 591, −$140 corrected; zero wins / n=1, Wilson 95% [0%,79.35%], far too little for inference. September 4: no eligible closes, win rate/interval undefined. Both days have four BOOT INTEGRITY receipts before 15:00; whole-calendar-day counts must not be presented as information available at 15:00 (`q48_eod_verified.json`, exact SQL, parameters, rows and source log lines). No broker-order snapshot exists before September 3's cutoff. An empty working-order book does not prove flat positions.

**Limits.** A watcher inside Windows/WSL cannot report while that host is down. With no new channel, that residual is explicit. Automatic recovery needs a separately reviewed operating change and restart testing; I cannot promise a three-minute recovery. UI behavior here is source-inspected, not observed in a live browser. All recommendations are proposals; this task changes only this report and its evidence.

---

## EVIDENCE BASIS — read this before any number

- **Worktree base:** `/home/hoang/nofx-vet-04` cut detached from `origin/dev` at **2a66d91c** (2026-09-05 07:22:45 -0500, "Merge #91: fix(risk,planner) wire RiskForceFlat and BiasArmWarning — both shipped uncalled"). Claim commit 318ff682 (`deploy/nofx-claim.sh check docs/vet-04-0905` → OK, claim names vet-04-0905). This report was produced across two runs of the same session (budget exhausted at ~10:30, resumed 11:10); the first run's 44 query outputs were kept as drafts and re-used only where each carries its output.
- **Running binary:** rev **36648655cfe0** (`GET /api/health` → `{"revision":"36648655cfe0","status":"ok","time":null}`; `systemctl show nofx` → `ExecMainStartTimestamp=Fri 2026-09-04 13:25:47 CDT`, `NRestarts=12`, `Restart=on-failure`). dev is ahead of the running binary by #90 and #91 (`git log 36648655..origin/dev`, data file `q03_dev_vs_running.out`). **For every file this section reads, running rev and dev tip are byte-identical except three:** `kernel/risk_limits.go` (+65, the RiskForceFlat arm/clear/reason map), `trader/auto_trader_planner.go` (+14, BiasArmWarning), `web/src/guide/types.ts` (GUIDE_BUILT_REV) — `git diff --stat 36648655..origin/dev -- <section files>` in `q44_rev_diff.out`. This is the earlier captured runtime; source inspection does not prove continuous runtime behavior or the loaded UI asset.
- **Spec-freshness lines** (`git log -1 --format="%h %ci %s" -- <path>`), quoted verbatim:
  - `docs/superpowers/SYSTEM-MAP.md` :: a96224dd 2026-09-04 09:07:37 -0500 docs: align SYSTEM-MAP labels with 2026-09-04 conformance corrections
  - `docs/superpowers/AUDIT-CHECKLIST.md` :: 15340faa 2026-09-04 13:22:07 -0500 Merge commit 'feba0ce5' into deploy/merge-claim
  - `docs/superpowers/research/INDEX.md` :: 4e8e7e1a 2026-09-03 19:37:14 -0500 docs(index): the stranded-branch sweep — 25 docs-only merged and indexed unclassified, 11 name-only-docs listed as not merged
  - `docs/superpowers/reports/2026-09-05-veteran-review.md` :: 676f239c 2026-09-05 05:51:29 +0000 docs(review): veteran review COMPLETE — parts C and D, lead sections rewritten (Part B, §4 MONITORING, is the prior text I re-measure against)
  - `docs/superpowers/reports/2026-09-04-two-day-audit.md` :: f3c640c3 2026-09-04 07:26:52 -0500 docs(two-day audit D3): why the blindness went unalerted — a note, not a build
  - `docs/superpowers/reports/2026-09-03-studio-audit.md` :: 35e0991a 2026-09-03 08:34:08 -0500
- **Store as-of** (`file:/home/hoang/nofx/data/data.db?mode=ro`, WAL; `q01_store_asof.out`, 09:55 CT; bell count re-read the q45 audit capture): trader_positions 587 (max created_at epoch-ms 1788444314627 = 2026-09-03 09:05 CT) · armed_orders 67 (max created 2026-09-04 12:11:02 CT) · plans 254 (max 2026-09-04 17:10:35 UTC) · plan_lifecycle_log 7 · touch_outcomes 677 · candidate_pool 360 · trade_excursions **0** · decision_records 37,768 (max 2026-09-04 18:28:01 UTC = 13:28 CT) · ab_confirm_log 223 · nt8_order_snapshots **4,461 and rising** (the AddOn emits a snapshot every ~30 s on a closed Saturday) · bars 79,769 (max MNQ 1m open 2026-09-04 12:19 CT) · planner_rejected_prompts 64 · planner_read_facts 32 · day_plan_alerts 655 · log_events 29,609 · trader_equity_snapshots 27,923 · day_plan_digests 65 · watchdog_fires 0 · telegram_configs 1.
- **API endpoints read (GET only, token from `cmd/gate-jwt`, never printed):** `/api/health` (no token) · `/api/expectancy` (200, 7,057 B) · `/api/config/resolved` (200, 26,495 B) · `/api/risk/gate-blocks` (200, 113 B) · `/api/cutover-gate` (200, 931 B) · `/api/risk/status` (200) · `/api/plan/today` (200) · `/api/plan/alerts` (200, 16,271 B) · `/api/status` (200) · `/api/positions` (200, `[]`). Payloads committed under `2026-09-05-vet-04-monitoring-data/` (`q04_*.json`, `q25_*.json`).
- **Logs:** `/home/hoang/nofx/data/nofx_2026-09-0{1,2,3,4}.log` read-only, always grepped by the `MM-DD HH:MM:SS` prefix across ALL files (rotation-on-boot trap confirmed: the last line of `nofx_2026-09-04.log` is stamped `09-05 10:11:24`). `journalctl -u nofx` read-only (retention starts 2026-09-04 17:27:17 CT). **NT8's own log on the Windows side**, `/mnt/c/Users/hoang/Documents/NinjaTrader 8/log/log.2026090{3,4,5}.*.txt`, read-only — this is the source the prior reports did not have, and it changes the 09-03 story (Q4). **Windows System event log** via `powershell.exe Get-WinEvent` (read-only), same reason.
- **Scratch:** every query/script lives in `~/nofx-analysis/vet-04-0905/q01…q49`; every figure below names its file. Decisive outputs are committed in `docs/superpowers/reports/2026-09-05-vet-04-monitoring-data/` (inventory expanded by this continuation; no configuration payload copied).
- **What I could NOT do:** drive the live UI (no browser against the live server — every UI claim is read from `web/src` at the running rev, which is byte-identical to dev for those files); read journald before 2026-09-04 17:27:17 CT; see what the owner's screen showed at any moment (I infer it from the components and the store).

---

## PREMISES CORRECTED

Every statement in the dispatch or a prior report that I found wrong or unmeasurable, with the query that shows it.

| # | Premise (source) | What the store/code/host says | Evidence |
|---|---|---|---|
| P1 | "113 minutes silent" (two-day audit §5; dispatch Q4) | **Confirmed recording-gap duration: 113.86 min** — `log_events` id 25754 (12:24:33 CT) → 25755 (14:18:24 CT) — the host-restart evidence is: EventLog 6008 "previous system shutdown at 12:24:02 PM was unexpected", Kernel-Power 41 at 12:25:22, OS started 12:25:13 CT. NT8's last log line is 12:24:50, the Go bot's 12:24:33. Note `log_events.ts_utc` is **epoch-MILLISECONDS despite the name**. | `q08`, `q47a`, `q47b` |
| P2 | "50 minutes blind" (dispatch); "~40 min" / "52 minutes" (audit, two figures) | **The available NT8 files show a startup after the gap.** Go booted 14:18:26 (tree `cc2`), 14:59:03, 15:02:18 (tree `cc3`); NT8 launched at **15:06:15** ("Session Break (Version 8.1.8.1)", vendor assemblies loading), AddOn CONNECTED 15:06:39, Tradovate Connected 15:06:43; first bars-ingest 15:06:45; first decision 15:08:51. Alive-but-blind = 14:18:26→15:06:45 = **48.3 min**, 27 min inside the NY window. Decision drought 12:22:44→15:08:51 = **166.1 min** (`decision_records` 37169→37170). | `q05`, `q09`, `q11`, `q47a` |
| P3 | "no alert" (dispatch Q4) | **An in-app P0 DID fire**: `day_plan_alerts` id 629, `feed-down`, 14:28:39 CT, `acked=0, dismissed=0`. Same on 09-04 (id 654, 12:30:01, unacked). And on 09-04 the **AddOn logged the loss 9m45s before Go did**: NT8 log 12:20:16 "VLTraderTCPClient: data feed lost (PriceStatus=ConnectionLost)". Owner attention is not recorded. | `q07`, `q08`, `q47e` |
| P4 | `trade_excursions` has 0 rows (dispatch) | **Confirmed** (0). `trader_positions.mae/mfe` populated on **70 of 71** rows with `entry_time ≥ 2026-08-15 CT` (not "66 of 227"). | `q01`, `q37` |
| P5 | "227 rows have entry_time ≥ 2026-08-15; 223 pnl_corrected non-NULL" (dispatch) | **Not reproducible.** `entry_time` is epoch-ms on all 587 rows. ≥ 2026-08-15 05:00 UTC → **71** rows (67 pnl non-NULL, 64 cited); ≥ 2026-08-01 → 253. The dispatch's `cited_scenario_id = 64` matches mine exactly, but the reason for its population mismatch is unverified. | `q38`, `q39` |
| P6 | Expectancy row "drops nine computed fields on the floor", incl. `stop_hit_share`, `target_hit_share`, `median_mfe` (veteran Part B §4.1) | Half right. `sum_pnl_corrected`, `sd`, `wilson_lo/hi`, `t_stat`, `flats` are non-null and NOT rendered (`web/src/components/plan/ExpectancyPanel.tsx:183-262` renders row/n/w-l/mean/mean 95%/win%/MAE/status/ids). But `median_mfe`, `stop_hit_share`, `target_hit_share` are **null in 6 of 6 rows** — uncomputed (trade_excursions empty), not "dropped". Shipping them today renders em dashes. | `q22_expectancy_rows.out`, `web/src/components/plan/ExpectancyPanel.tsx:16-44` |
| P7 | "the two cancel paths at :941 and :960 call nothing" (Part B §4.4) | Half right at dev tip: the marketable cancel counts nothing; the **stale-arm expiry DOES count** (`arm_superseded_unplaced_class47 = 2`). But that expiry is **version-triggered**, so a never-placed arm on a dead plan's LAST version never expires — ids 104/105 read `armed` since 09-04 12:11 CT, **23.1 h old on a Saturday**, and hold cutover leg 4 at MISMATCH (broker 0 vs ledger 2). | `q28`, `q42`, `q45c2`, `q04_api_cutover-gate.json` |
| P8 | GateBlocksPanel "does not know the names of your risk gates" (Part B §4.4, four examples) | Measured: **41** distinct `IncGateBlock` literals in Go, **14** panel labels, **11** overlap, **30 Go gate names render raw**; the 3 panel-only names (`b3_order_dedup`, `b3_rate_breaker`, `clock_skew_observed`) ARE counted (via variable / `telemetry/gate_blocks.go:51`). | `q40` |
| P9 | "`nt8_order_snapshots` is the broker's book" (dispatch, time traps) | The first available order snapshot is **2026-09-03 21:48:22 CT**. No earlier snapshot is available for this reconstruction; this is an order book, not a position book. | `q06` |
| P10 | `day_plan_alerts.created_at` is epoch seconds (implicit) | 652 rows seconds, **3 rows milliseconds** (ids 3–5, early era). A naive `datetime(created_at,'unixepoch')` yields invalid/out-of-range SQLite dates for those three. | `q07` |
| P11 | AUDIT-CHECKLIST "79 classes" (dispatch) | PART 1 holds **78** numbered items whose max number is **79** (one number is absent), plus one `## CLASS 75` header. I quote "≤79" and never "79". | `q21` |
| P12 | "kernel/exit_*" — confirmed absent; exit logic where the dispatch says. | `ls kernel/exit_*` → none. | — |
| P13 | `GUIDE_BUILT_REV` drift | **The source revision matches the captured health revision; loaded browser asset parity was not verified**: `web/src/guide/types.ts:6` = `36648655cfe0…` = running rev. | `q02` |
| P14 | Two-day audit D3: "why the blindness went unalerted — a note, not a build" (the file's own subject line) | Host-restart receipts and the application-log gap support investigating recovery outside Go. Exact cause, historical launch coverage, and recovery-time improvement are not established; see Q4. | `q47a-d` |
| P15 | My own first-run draft: "420 unacked" | **421** unacked-and-undismissed at the q45 audit capture (90 P0 + 331 P1). Same query, later read. | `q45a` |
| P16 | "the AddOn's own 30 s heartbeat/ack loop" as something to *add* a watchdog to (my first-run summary) | The AddOn already HAS a bar-stall watchdog (`ninjascript/VLBarsSubscriptionManager.cs:106-113`: 15 s cadence, 20 s fast-stall, 75 min backstop) and a connection-status handler; both fired on 09-04. What is missing is not detection — it is any output louder than a log line. | `q47e`, `ninjascript/VLBarsSubscriptionManager.cs:593-712` |

---

## Q1 — INVENTORY: what the owner sees, and what each surface reads

File:line references use the section source base 2a66d91c; q44 records inspected differences from running revision 36648655. Treat runtime behavior as captured evidence, not a new live verification. I list the surface, its source, what it reads, and what it does NOT show.

### 1.1 The dashboard header strip (`web/src/pages/TraderDashboardPage.tsx`)

- `SYSTEM_STATUS::<status> · REV::<revision>` — polls `GET /api/health` every 30 s (`:159-177`). The handler (`api/server.go:626-632`) returns `{"status":"ok","time":null,"revision":...}` unconditionally. **It is a process-liveness ping only** — its handler can return "ok" while no bars arrive; there is no continuous HTTP observation for that interval.
- `EQ::<total_equity>` · `PNL::<total_pnl>` (title "NT8-native total P&L (equity − initial balance)") · `<LedgerDayPnl>` (`:662-666`). The ledger chip (`web/src/components/trader/LedgerDayPnl.tsx:11-27`) renders `LEDGER_DAY::+x.xx (n resolved, m unresolved excluded)` from `account.ledger_day_pnl`, which `api/handler_order.go:157-167` computes via `GetLedgerDayTotal` over rows closed TODAY CT (`dayStart` = midnight CT, `:156`), strict `pnl_corrected`. This explicitly labels corrected daily ledger money, with population differences documented in Q5. **[A]**
- Stat cards: Total equity · Available balance (+ "% free") · Total P&L · Positions (+ "margin: x%") (`:684-741`). Margin and "% free" are crypto-margin concepts; on an NT8 SIM account they render `--` or a meaningless ratio.
- `call_count`, `runtime_minutes` (`:602-609`; source `trader/auto_trader_decision.go:84`) — minutes since process start, resets on every boot (12 boots on 09-02, 11 on 09-03, 4 on 09-04; `q46a`).
- Positions table columns (`:880-1020`): symbol · side · entry · mark · qty · value · **lev** · uPnL · **liq** · close button. No stop, no target, no distance, no open risk, no time-in-trade. `liquidation_price` and `leverage` are Binance fields; NT8 SIM has neither.
- Pause chip (`:426-431`, `status.stop_until`), `EmergencyFlatButton` (`POST /api/risk/force-flat`, `web/src/components/trader/EmergencyFlatButton.tsx:29-30`), `DecisionAudit` (`:800`), `ChartTabs` (`:832`), `DecisionCard` (`:1187`), `PositionHistory` (`:1232`).
- 402 banner: polls `GET /api/risk/errors` every 30 s (`:183-199`) for `ai_payment_402`.

### 1.2 The plan card (`web/src/components/plan/PlanCard.tsx`, mounts at `web/src/pages/TraderDashboardPage.tsx:816`)

Composition order (`web/src/components/plan/PlanCard.tsx:83-176`): `AlertCenter` → `SessionTimelineStrip` → `SessionTabs` → `GateBlocksPanel` → `ExpectancyPanel` → `InstrumentsDrawer` → Approve/Re-read/Reset buttons → `SessionPlanCard`. Data: `usePlanToday` (`web/src/components/plan/usePlan.ts:28-36`, SWR, `PLAN_REFRESH_MS`, deduped 5 s) → `GET /api/plan/today?trader_id&symbol&session&version`; `usePlanVersions` (`:53-56`); `usePlanAlerts` (`:67-70`, `ALERTS_REFRESH_MS`).

`SessionPlanCard.tsx` renders, in order: reading banner (`:252`, `plan.reading`), night panel (`:261`), no-plan panel (`:273`), replan-in-flight chip (`:364-378`), dormant block (`:383`), title + `VersionChips` + `LifecycleChip` + `WeeklyChip` + degraded badge (`:416-433`), Ask/Edit buttons (`:465-479`), mode badges advisory/warming/uncalibrated (`:488-497`), NO-TRADE banner (`:514-544`), fail-closed banner (`:548-561`), historical-version block with death reason/condition/"what changed" (`:585-648`), death history (`:675-700`), uncarried owner levels (`:723`), `BiasBlock` (`:740`), `PlanMiniChart` (`:749`), `ZoneTable` (`:771`), `RealignPanel` (`:788`), `ScenarioList` (`:829`), `ArmedUnderBlock` (`:884`), `RulesBlock` (`:893`), `PlanFooter` (`:901-912`: version chips · day type · re-reads left · model).

The arm chip (`web/src/components/plan/ScenarioList.tsx:223-290`): `⏳ armed` · `📌 working` (pulsing) · `⚡ filled` · `✕ cancelled · <reason>`, with per-leg glyphs. **No entry price, no placement time, no age, no distance from price.** Ids 104/105 would show `⏳ armed` today exactly as they did at 12:11 yesterday.

`/api/plan/today` payload today (`q25_api_plan_today.json`): `active_session:"NY"`, `found:false`, `is_active:true`, `mode:"strict"`, `reading:false`, `replan_in_flight:false`, `runnable_sessions`, `trade_date:"2026-09-05"`, weekly block (PWH 29811.75 / PWL 28927.25, `refs_only:true`). So the card DOES carry `reading` and `replan_in_flight` — two of the planner-state fields Q2 asks for — but SessionPlanCard renders them only as a banner/chip when true, never "next read at HH:MM".

### 1.3 AlertCenter (`web/src/components/plan/AlertCenter.tsx`)

In-app only by design (`:1-4`: "NO external push"). Bell with unacked badge (`:204-236`), dropdown feed, persistent banner for the top unacked P0 (`:196-197, :254`), a toast per NEW P0 (`:161-169`, never the initial backlog). Ack and dismiss via `POST /api/plan/alert-ack` / `alert-dismiss` (`q24_routes.out`). The producer is `emitAlert` (`trader/auto_trader_alerts.go:15-34`) → `store.Alert().Emit` with `EventID` dedupe. Levels: P0 pop-up+banner, P1 feed, P2 digest. Prune: acked P2 older than 7 days, once per session day (`:41-63`) — P0/P1 never auto-prune. Kinds all-time (`q06`): armed 216 · level-burned 136 · close 71 · fill 51 · regime-dark 45 · owner-reset 44 · decision-stale-bar 35 · read-fail 19 · guardrail-would-have-tripped 15 · planner-preflight 6 · plan-died 4 · feed-down 3 · ai-payment 2 · calendar-slice-missing 2 · decision-unparseable 2 · owner-reread 2 · plan-death-streak 2. Ack rate all-time: P0 62/152 acked, P1 160/500 (`q08`). **Today: 421 unacked-undismissed (90 P0, 331 P1)** (`q45a`).

### 1.4 GateBlocksPanel (`web/src/components/plan/GateBlocksPanel.tsx`)

Polls `GET /api/risk/gate-blocks` every 20 s (`:52-70`), merges `by_trader[traderId]` and `by_trader[""]` (`:54-57`), labels via `GATE_LABELS` (`:20-50`, 14 entries), raw name fallback (`:113`). Today's payload: `{"by_trader":{},"session_day_utc":"2026-09-03T22:00:00Z","summary":"no gate blocks recorded"}` (`q04_api_risk_gate-blocks.json`). Label coverage: 30 of 41 Go gate names render raw (`q40`), including `entry_gate`, `rr_gate`, `min_sl_gate`, `htf_veto`, `stale_data`, `plan_off_plan` — the ones that fire.

### 1.5 ExpectancyPanel (`web/src/components/plan/ExpectancyPanel.tsx`)

Polls `GET /api/expectancy` every 60 s (`:106`). Columns rendered (`:183-262`): row · n · w/l · mean · mean 95% · win% · MAE · status · ids; "DESCRIPTIVE ONLY" when `descriptive_only` (`:262`); an honesty ledger line with the exclusions and `Floor: n ≥ min_n` (`:292-302`); the E8 counterfactual block (`:305-336`). Today's payload (`q22`): `min_n 30`, `as_of_utc 2026-09-03T14:20:45Z`, excluded `{unresolved_pnl 3, unresolvable 7, test_seam 3}`, six rows: acceptance n=6 · breakout_retest n=9 · hold n=1 · reclaim n=5 · **reject n=31 FAILS** (mean +18.9, 95% [−18.3, 56.1], win 45.2% Wilson [29.2%, 62.2%], ids listed) · sweep_reclaim n=6. `median_mfe`, `stop_hit_share`, `target_hit_share`: null on 6/6. The `as_of` is two days stale on a Saturday — its meaning needs checking: it equals the last cited close time, so this report does not establish it as cache-computation time.

### 1.6 InstrumentsDrawer (`web/src/components/plan/InstrumentsDrawer.tsx`)

Collapsed below the table (`:104-110`). Three instruments: discipline (`GET /api/plan/trades` adherence summary, `:134-164`), MAE/MFE (`GET /api/expectancy` medians, `:170-200`, "source: trade_excursions" — which is empty, so it renders dashes), and the retired random-level gate (`GET /api/plan/stats` `level_stats`, `:210-258`, labelled BIASED and kept only for its frozen weekly verdict).

### 1.7 Boot lines (`main.go`, `kernel/levels_volume_boot.go`, `trader/auto_trader*.go`)

The 09-04 13:25:47 boot block is 301 lines, of which the emoji-led owner-facing lines are in `q16_boot_block_0904.out` (`q15_boot_lines.out` maps each glyph to its source line). The ones a trader reads: `🔐 BOOT INTEGRITY OK — rev … goldens PASS` (`main.go:327`), `🧾 ledger boot: sessions[…] stop_until=none` (`trader/auto_trader_pause.go:202`), `💰 Initial balance: 50000.00 USDT` (`trader/auto_trader.go:834`), `🛑 min-sl guard: atr_mult=1.5` (`trader/auto_trader_dayplan.go:57`), `⚔️ armed_orders=on place_band=100t stale_working=15m … arm_rr=2.0`, `🛡 cutover safety (class 33): gate legs=5 · leg4=ledger (no snapshot yet)`, `🔌 nt8 addon: build_id=2026-09-03-f12 expected=… match=yes`, `🧾 Track record: −217.18 over 110 resolved trades (115 unresolved excluded), 36.4% win rate, PF=0.95` (`trader/auto_trader_loop.go:1404`), `🛡 clock-guard [boot] rtc_vs_go=…` (`kernel/clock_health.go:164`), and — in that same block, 60 s after start — `🚨 FEED DOWN: no NT8 bar for 1h6m47s while CME is OPEN`. That boot was recorded as passed (`b2d3826e deploy: boot 10 marker … after the passed boot`). **[A]** A boot block that prints its own feed-down line and is accepted anyway is the class-49 "instrument theatre" probe failing in the other direction: the feed warning and the passed marker conflict; who read it is unknown.

### 1.8 Guide (`web/src/guide/content/*.ts`, `GuidePage.tsx`)

Fourteen content modules (buttons, expectancy, faq, glossary, guards, levels, planCard, plays, routines, settings, status, tradingDay, weeklyBias, welcome). The drift banner compares `GUIDE_BUILT_REV` (`web/src/guide/types.ts:6`) with `/api/health.revision` (`web/src/guide/GuidePage.tsx:442-470`); source revision matches the captured health revision; loaded assets were not inspected. What the Guide says about monitoring: a status table (`web/src/guide/content/status.ts:100-116`) with rows "NT8 feed — TCP bridge alive + bars flowing", "Boot integrity", "**Dead-man watchdog — Kernel heartbeats ok.**", "Trader frozen", "Clock drift", "402 banner"; a traffic light ("GREEN = bot running … RED = feed down / frozen / boot mismatch", `:139`); a pre-market checklist that asks "NT8 running? TCP bridge connected (SYSTEM_STATUS green, 'NT8 feed')?" (`web/src/guide/content/routines.ts:16`); an EOD checklist of three lines ("All positions flat · Day summary: fills, gate-block counter, would-have-tripped · Note anything to fix", `:36-42`); and an emergency card "Feed / NT8 down" (`:73-78`). **Two of those claims have no surface behind them** [A]: there is no "Dead-man watchdog: kernel heartbeats ok" indicator anywhere in `web/src` (the phrase is a gate-block LABEL for `dead_man`, which renders only after a refusal), and `SYSTEM_STATUS` green is `/api/health`, which does not know whether bars flow. The pre-market checklist therefore tells the owner to confirm the feed with an indicator that cannot confirm it.

### 1.9 Telegram / alerts outside the app

`telegram_configs` has one bound row (token + chat_id, bound 2026-06-11; `q06`). `telegram/bot.go:23 Start` runs `GetUpdatesChan` (`:109-114`) and replies to inbound messages; `sendMsg`/`sendMarkdownMsg` (`:254-263`) are called only from that loop. `grep -rn "telegram\." trader/ kernel/` → nothing: **no trader/kernel Telegram push call was found in the inspected source** [A]. The feed watch says why: "in-app only — the owner's no-external-push rule" (`trader/auto_trader_feedwatch.go:63`). `watchdog_fires` has 0 rows.

### 1.10 Logs the owner could read but doesn't

- `data/nofx_<boot-date>.log`: WARN+ lines are also shipped to `log_events` (`main.go:105`, "WARN+ → log_events"); INFO never reaches the store. The Go-side "🚨 FEED DOWN" is ERROR, so it is in `log_events` — but nothing reads `log_events` for the owner except by SQL.
- journald: `tcp_server` logs through `slog.Default()` (`provider/ninjatrader/tcp_server.go:483`), so the link events ("hello handshake OK", `:1713`) go to journald and **never to `nofx_*.log`** (0 occurrences in four files, `q12`) — where they drown: 908 of 1,429 journal lines in the last hour are `received frame type=…` at INFO on a closed Saturday (`q13`), and retention is 16.5 hours. The class-12 log-flood probe, re-run, fails.
- NT8's own log (`…/NinjaTrader 8/log/`): the AddOn's watchdog writes "most-stale <key> bar age Ns; dead-subscriptions=N" every minute (`ninjascript/VLBarsSubscriptionManager.cs:619-624`) and the connection handler writes "data feed lost" (`ninjascript/VLTraderTCPClient.cs:315+`). On 09-04 that log had the answer at 12:20:16 CT. Nobody on the WSL side reads it, and the Windows side has no reader.

---

## Q2 — What a trader needs on screen DURING a session that isn't there

The table distinguishes missing information from partial existing displays. For each I say where the number already exists in the system, so the strip is wiring, not research.

| Need | Exists today? | Where it would come from |
|---|---|---|
| Open risk in $ (contracts × (entry − stop) × $2/pt for MNQ) | No. The positions table has no stop column. | `trader_positions` open row + the working stop from `nt8_order_snapshots.orders_json` (or the arm's `stop_px`); MNQ point value 2.0 |
| Distance to stop / to target, pts and ticks | No | mark − stop, target − mark; tick 0.25 |
| Session P&L vs the limit | Half. `LEDGER_DAY::` is the day's realized on `pnl_corrected` (`api/handler_order.go:157`). The limit is not shown; `/api/risk/status` shows the wrong one. | Studio `risk_control.daily_loss_limit_usd` + `daily_loss_enabled` + master flag (`q26b`) — show "limit OFF (450 configured)" when it is off |
| Arms resting, with age and distance | No (chip shows state only, `web/src/components/plan/ScenarioList.tsx:223-290`) | `armed_orders` where state in (armed, working): `entry_px`, `created_at` (CT-normalised!), price − entry_px in pts and ×ATR5m |
| Broker book vs ledger | No (only in `GET /api/cutover-gate` leg 4, for which no dashboard poller was found) | `nt8_order_snapshots` latest (`working_count`, age) × `armed_orders` working count; today: broker 0 vs ledger 2 |
| Last fill slippage | No | `armed_orders.fill_price − entry_px` on the last filled arm (real series so far: ids 20/21/24/28/35 → 0.0, 0.0, 0.0, 0.0, −0.62 pt; `q34`) and exit `exit_price − stop_px` on the last stop (pos 591: 29355.00 vs 29351.63 = **3.37 pt = 13.5 ticks adverse**; `q35`) |
| Planner state: in flight / next read / wake suppressed | Half. `reading` and `replan_in_flight` are in `/api/plan/today` (`q25`); "next scheduled read" and "wake suppressed until" are not. | session registry (`ledger boot` line: ASIA 16:30 / LONDON 01:30 / NY 08:00 reads), `plannerReadInFlight` claim, the class-47 wake throttle (`WARN-first cutoff 25m / cooldown 30m`) |
| Regime | No (only inside the plan doc's bias block) | `planner_read_facts.bias_regime` ("up/NORMAL") + `regime-dark` alerts |
| Day range vs ATR | No | `bars` 1m RTH high−low vs daily ATR(14): 09-03 = 385.75 pts / 410.45 = **0.94** (`q45b`) |
| Feed age / link age | No. `SYSTEM_STATUS::ok` is process-only. | `feedNewestBarAge` (`trader/auto_trader_feedwatch.go:52-61`); last heartbeat ack age from `tcp_server`; `nt8_order_snapshots.received_at_ms` age (12 s at 10:20 today, `q42`) |

**Mock — the DESK STRIP, one line each, monospace, above the plan card, refreshed every 5 s from one endpoint (`GET /api/desk`):**

```
FEED  1m bar age 0:42 · ack 0:11 · snapshot 0:12 (build f12)        [green]
LINK  NT8 CONNECTED since 13:25:47 · Tradovate Connected · Sim101   [green]
LOOP  last decision 12:11:00 (n/a — no_new_data ×15) · next read NY 08:00 (done) · wake: suppressed 25m (class 47) until 12:40
POS   FLAT                                        — or —
POS   SHORT 1 MNQ @ 29285.00 · mark 29301.25 · uPnL −32.50 · stop 29351.63 (+50.4 pt / 202 t) · tgt 29200.00 (−101.3 pt) · risk $133 · in 00:14:31
ARMS  2 resting: NY S2 SHORT 29720 (age 1h17m, +185.8 pt / 8.3×ATR5m from price — FAR) · broker book 0 vs ledger 2 — MISMATCH
DAY   LEDGER_DAY −140.00 (1 resolved, 0 unresolved excl.) · limit OFF (450 configured, master OFF) · RTH range 385.75 pt = 0.94×ATR14
FILL  last entry 29285.00 vs arm 29285.00 (0.0 pt) · last stop exit 29355.00 vs stop 29351.63 (+3.37 pt / 13.5 t adverse)
REG   up/NORMAL · ATR5m 39.9 · stop floor 59.8 pt · plan NY v7 active · 0 re-reads left
ALRT  P0 unacked: 1 (feed-down 12:30:01) · P1 unacked: 11                       [P0 → existing red banner]
```

The proposed implementation must give every field an as-of or an age; a field the process cannot know prints `n/a` with the reason (the boot-line law). This is a synthetic layout combining historical examples with invented ages, counts and state to illustrate the requested fields; it is NOT one observed snapshot. Historical price examples are in q28/q29/q35/q42/q45. The `ARMS … FAR` flag is the existing "📏 arm far" WARN (`q20`: "S2 short entry 29720.00 is 185.75 pts / 8.3×ATR5m from price 29534.25 (counted, not refused)") surfaced.

What a desk would demand first: the FEED and LINK rows. On 09-04 the FEED row would have read "1m bar age 11:02" at 12:30 and stayed red for 3.5 hours; the LINK row would have read "Tradovate Disconnected" from 12:22:29. Everything else on the strip is for trading; those two communicate whether inputs are fresh. The historical layout is counterfactual, not an observed screen.

---

## Q3 — What's on screen that is noise or misleading

1. **`SYSTEM_STATUS::ok`** (`web/src/pages/TraderDashboardPage.tsx:650-657` ← `api/server.go:626-632`). A literal. It can return "ok" during an alive-but-blind phase; continuous historical HTTP responses were not captured. A trader reads it as "the system is ok". Misleading. [A]
2. **`PNL::` "NT8-native total P&L (equity − initial balance)"** and the **Total P&L card** — an all-time number from the SIM account's equity line (51906.50 − 50000 = +1906.50), sitting beside `LEDGER_DAY` and the Track-record line's **−217.18 over 110 resolved** (`q46d`). These are different populations; I would label their scope rather than compare them as equivalent returns. [A]
3. **`/api/risk/status`**: `daily_loss_limit_usd 500 · kill_switch_armed true · last_reset_utc <now>` — a limit that is not enforced (`kernel/engine_analysis.go:122-124`), an "armed" that means "env var > 0", a reset time that is the request time. Unread by the UI; read by anyone auditing via curl, who will believe it. Delete or fix. [A]
4. **Positions table `lev` / `liq` / margin %** — crypto fields (`:880-1020`); on NT8 `liquidation_price` is absent and leverage is a strategy constant. Replace with stop/target/distance.
5. **Crypto vestige log lines in the futures loop** — `Strategy leverage config: BTC/ETH=5x, Altcoin=5x` (146/day), `Strategy engine fetched candidate coins: 1` (146/day), `Account equity: 51906.50 USDT` (145/day), `Initial balance: 50000.00 USDT` — 1,129 to 2,048 lines per day (`q46c`). Not on the dashboard, but they are in the log the emergency checklist tells the owner to read, and every one of them is wrong for MNQ.
6. **"DESCRIPTIVE ONLY" and "NOT ENOUGH DATA" rows** (`web/src/components/plan/ExpectancyPanel.tsx:262`; `q22b`): five of six rows are below the floor of 30 (n = 6, 9, 1, 5, 6) and the sixth (reject, n=31) reads FAILS with a 95% mean interval of [−18.3, +56.1] that straddles zero. The floor is right; the table is still rendered above the fold every cycle where it cannot rule on anything. A trader cannot use a row whose Wilson band is [3%, 56%]. Collapse the table by default until any row passes; keep the honesty ledger line visible.
7. **`median_mfe`, `stop_hit_share`, `target_hit_share`** in the payload as null on 6/6 rows (`q22`): fields that read as "measured, absent" are in fact "never computed" (trade_excursions empty). The InstrumentsDrawer's MAE/MFE instrument cites "source: trade_excursions" (`:200`) and renders dashes. Meanwhile `trader_positions.mae/mfe` is populated on 70 of 71 era rows (`q37`) and is not used by either.
8. **Expectancy timestamp semantics:** the payload as-of equals the last cited close (q22). Add separate computed-at and latest-close labels only when both are actually measured.
9. **The Guide's status table row "Dead-man watchdog — Kernel heartbeats ok."** (`web/src/guide/content/status.ts:108`) and the pre-market step "TCP bridge connected (SYSTEM_STATUS green, 'NT8 feed')" (`web/src/guide/content/routines.ts:16`): both describe indicators that do not exist on any page. The guide-content law says a guide that lies about the running binary is worse than none; this one lies about the running dashboard.
10. **Arm chip without age** (`web/src/components/plan/ScenarioList.tsx:246`): `⏳ armed` on 104/105 for 23.1 h reads identically to an arm placed a minute ago. A stale label. (`q45c2`)
11. **Session digests written before the session** — "NY 2026-09-04 — 0 entries · realized +0.0 · flat" (`q31`, `q33`): shown in DayPlanEditor and AlertCenter and fed to the planner. Wrong by construction (Q1.3 / summary problem 3).
12. **Bell count 421**: 216 of all-time alerts are `armed` P1s that mean "a plan version was written" — one per re-plan, never resolved when the next version supersedes it. The existing P0 banner still exists; I judge the backlog distracting, but cannot establish that it hid either alert. (`q06`, `q45a`)
13. **Gate names rendered raw** (30 of 41, `q40`): `entry_gate`, `rr_gate`, `stale_reeval_refused`, `plan_off_plan` etc. appear as identifiers. Not misleading, but not readable at 09:31 either.
14. **`GET /api/plan/alerts` `created_at` unit mix** (3 rows in ms, `q07`): harmless for the UI, a trap for any EOD script.
15. **Track-record line labels** "Sharpe=−0.02, DD=9.9%" on 110 trades (`q46d`): a Sharpe on 110 SIM trades is decoration. PF and win rate with n are enough.

---

## Q4 — September 3 silence: verify the premise and stay within existing channels

I distinguish missing records, process recovery, and human notification. These are different measurements.

| Observation | Evidence and limit |
|---|---|
| Log silence 113.86 min | `log_events` ids 25754 at 12:24:33 → 25755 at 14:18:24, epoch milliseconds (`q08`). This proves a recording gap. |
| Unclean host restart | Kernel-General startup 12:25:13, Kernel-Power 41 12:25:22, EventLog 6008 reporting unexpected shutdown 12:24:02 (`q47b`). Shutdown time precedes surviving application stamps by seconds; cross-clock precision and root cause remain unresolved. |
| About 48.3 min alive before bars | Go boot 14:18:26 → first observed bars 15:06:45 (`q09`, `q11`). NT8 file starts 15:06:15 and connection is logged 15:06:43 (`q47a`). That supports late application recovery, not proof of a human launch. |
| Decision drought 166.1 min | ids 37169→37170, 12:22:44→15:08:51 (`q05`). At 15:00 the later row was not yet available; q48 reports the trailing gap to cutoff instead. |
| In-app alert existed | P0 `feed-down` id 629 at 14:28:39, unacked/undismissed at audit (`q08`). It could not report during process downtime. No claim about owner attention follows from an ack flag. |
| September 4 is a feed/link failure | NT8 price loss 12:20:16 and DNS/login failures thereafter (`q47e`); Go P0 id 654 at 12:30:01 (`q08`). Detection difference is **9m45s**, not the draft's 14 minutes. The cause of network/authentication failure is unresolved. |

The Windows startup census in q47d found no matching scheduled task or Startup item. It was not a complete historical audit of all services, registry launch paths or user actions. WSL uptime and wtmp disagree by about 19 minutes. I therefore infer a recovery gap; I do not claim Windows sat idle for exactly 113 minutes or that all applications were absent throughout it. Seven Kernel-Power event dates are receipts of unclean restarts, not a probability of next month's failure or proof of hardware faults.

**My monitoring design within the owner's existing channels:** show last bar, last broker frame and last decision separately in the dashboard; retain the P0 banner across reloads with the original event time; show a persistent boot/feed failure in the existing boot log; surface the AddOn's existing disconnect/stall messages in its existing log (`ninjascript/VLBarsSubscriptionManager.cs:593-712`; `ninjascript/VLTraderTCPClient.cs:315`). Reuse the clock-guard timer pattern only as a proposed process-independent collector for existing dashboard/log surfaces (`deploy/nofx-clock-guard.sh:74-78`, `kernel/clock_health.go:121-164`). It is still host-dependent. Proposed detection thresholds (90 seconds without ack, three minutes of stale subscriptions) need normal-session and market-closure testing; they are not measured service guarantees.

I do **not** recommend new Telegram pushes, Windows popups, or sound under the declined-channel constraint. Telegram's observed code is reply-oriented and no trader/kernel push caller was found (`telegram/bot.go:23,109-114,254-263`; `trader/auto_trader_feedwatch.go:63`). That establishes the inspected implementation, not a lifetime delivery audit. An off-host alarm would cover host loss but lies outside this scope. Automatic startup may shorten recovery, but is an operating proposal, not an alert channel or a proven cure; evaluate account/session prerequisites and protection reconciliation before any separate implementation.

The September 4 first-monitor FEED DOWN receipt should have prevented an operational “healthy” assessment (`q16`; marker `b2d3826e`). I would require explicit fresh-feed and broker evidence before declaring readiness. No restart, cancellation, trading configuration, or startup change was performed here.

---

## Q5 — The one page I would read at 15:00 CT, with exact queries

The executable query specification is [q48_eod_verified.py](2026-09-05-vet-04-monitoring-data/q48_eod_verified.py); [q48_eod_verified.json](2026-09-05-vet-04-monitoring-data/q48_eod_verified.json) stores **each exact SQL statement, bound parameters, returned row IDs and output**, plus dated source-log lines. Run `python3 /home/hoang/nofx-analysis/vet-04-0905/q48_eod_verified.py`. It opens only `file:/home/hoang/nofx/data/data.db?mode=ro` in one read transaction and writes its output only to the original scratch directory. Python `zoneinfo` derives midnight and 15:00 America/Chicago bounds; the hour-display SQL's −5 is explicitly September CDT only. Period is [midnight,15:00), not a CME session-day.

**One-page view, reconstructed at 15:00:**

| Block | September 3 | September 4 | Exact receipt |
|---|---|---|---|
| 1. Closes and exclusions | id 591, n=1, corrected −$140; wins 0/1, Wilson 95% [0%,79.35%] | n=0, no eligible closes; rate and interval undefined | q48 block 1; exclusion categories and IDs included |
| 2. Arms created | 3 IDs | 30 IDs | q48 block 2; state updated after cutoff becomes UNKNOWN_AT_CUTOFF |
| 3. Refusal-bearing cycles | 1 | 0 | q48 block 3; excerpts and IDs, not distinct refused orders |
| 4. Feed/loop | Trailing decision drought to 15:00; no assumption of 15:08 recovery yet | Interior plus trailing decision gaps; last observed 1m bar in 12 CT hour | q48 blocks 4/4b; timestamps, adjacent IDs, counts by hour |
| 5. Touch outcomes available | 0 rows satisfying creation and close cutoffs | 20 rows | q48 block 5; per-kind n, ambiguous count, Wilson and IDs; later recorded outcomes excluded |
| 6. Planner | 11 fact rows, 9 plan versions available | 15 fact rows, 9 plan versions available | q48 blocks 6/6b; separate populations, session labels and IDs |
| 7. Operational log receipts | 4 BOOT INTEGRITY, 31 FEED DOWN, 0 matching no_new_data lines | 4 BOOT INTEGRITY, 149 FEED DOWN, 77 matching no_new_data lines | q48 block 7; rotation files scanned by timestamp prefix, source path and line |

The zero log count is a result for that exact text and cutoff, not proof that cycles were healthy. The whole-day q41/q43/q45/q46 figures remain historical context, **not** 15:00 values. The earlier q29 entry-date query is retained as draft evidence and superseded for EOD; its broad literal epoch cutoff and raw-P&L display must not be reused as the authoritative recipe.

For block 1, use CLOSED rows by **exit_time**, strict `pnl_corrected`; never substitute `realized_pnl`. Categorize in order: exact source `e7_farside_test`, pre-Aug15 entry era, `plan_id='UNRESOLVABLE'`, null corrected P&L, excluded reconcile/unresolved/test-seam close reason, then eligible. This exclusive order makes totals reconcile; the JSON lists each excluded ID. NULL source is not silently dropped by `source <> ...`. UNRESOLVABLE is an attribution sentinel and distinct from null money: this review conservatively excludes it from its eligible performance population, while the current ledger-day handler has its own broader population (`store/pnl_surface_guard.go:60-91`). Do not assert those populations are identical.

I would also print **broker/ledger protection UNKNOWN unless independently evidenced**. September 3 has no prior broker-order snapshot; September 4's latest pre-cutoff snapshot has its ID and age in q48. Working orders are not position quantity. The current ledger OPEN count and Saturday snapshots cannot reconstruct historical flatness. Arm terminal states and alert acknowledgements can change after cutoff; q48 withholds later arm state and does not pretend current ack flags were known at 15:00. Plans may have later lifecycle updates, so the query emits versions and creation receipts, not reconstructed historical lifecycle. Mutable daily counters cannot be recovered as-of from current `system_config`; current totals in q42 are supplementary only.

This is the page I would use to investigate readiness and recording gaps, not a recommendation to trade. A current-table retrospective cannot fully reconstruct the owner's historical screen or all event transitions.

---

## RECOMMENDATIONS (ordered, proposals only)

1. **Make readiness explicit on existing surfaces.** Add feed/link/loop ages to the dashboard and boot log; never infer feed health from `/api/health`. Show stale/unknown independently. Validate against the q08/q16/q47 outage timeline. A same-host monitor cannot cover host downtime.
2. **Correct risk and position displays.** Read bound Studio master/per-limit enablement and value; show OFF and corrected ledger-day population. Add stop/target, signed distance, position age and open risk. Verify stop quantities and broker evidence; no stop receipt means UNKNOWN risk, not zero (`api/handler_risk.go:178-233`; Q2).
3. **Resolve without forging acknowledgement.** Keep P0 prominent; mark superseded arms and recovered feed events resolved with timestamps, preserving the record and actual owner ack. Do not auto-ack on the owner's behalf (`web/src/components/plan/AlertCenter.tsx:161-169,196-197`; q45a).
4. **Ship the cutoff-correct EOD view.** Use q48's exact queries and exclusions; add immutable transition history before claiming complete historical state. Acceptance: no after-cutoff leakage, no raw P&L fallback, IDs and unknowns visible.
5. **Repair session digests and planner facts.** Write a session digest only after its actual window ends; validate dates before SaveIfAbsent. Populate or explicitly label unwritten fact columns (`trader/auto_trader_planner.go:2448-2512`; q33/q43). Do not repair historical DB rows in this task.
6. **Reduce monitoring noise.** Label the 30 raw gate names (q40); separate stale cache/data timestamps; remove irrelevant crypto units; retain low-n expectancy evidence in a collapsed panel with counts and Wilson intervals. Win rate alone is not expectancy or a release criterion.
7. **Investigate recovery separately.** q47 supports unclean host restarts and an application recovery gap, not a hardware diagnosis. Review Windows startup coverage and NT8 login dependencies before proposing automatic recovery. No elapsed-time promise follows from a one-line launcher. All startup, service and sound changes remain outside this documentation task and the existing-channel proposal.
8. **Bring Guide wording into parity.** Correct `web/src/guide/content/status.ts:108` and `web/src/guide/content/routines.ts:16` together with the actual status surface; keep the revision comparison. Verify the loaded production asset separately rather than assuming source revision proves it.

Success measures I would request in later implementation: feed-loss-to-banner latency, stale fields explicitly marked, unacknowledged vs resolved P0 counts separately, no accepted readiness check with stale feed, and EOD exclusion totals reconciling to row IDs. These are proposed checks, not measurements already achieved.

---

## SURPRISES (found, not acted on)

1. **Seven Kernel-Power 41 unclean reboots on the host in 60 days** (07-16, 07-23, 07-25, 08-16, 08-19, 08-21, 09-03; `q47c`). No checklist class covers the host; the autostart memory says the owner-side install is still pending.
2. **NT8's Tradovate connection never came back on 09-04** and "There was a problem authenticating account Google Simulation" repeats every 30 s into Saturday (`q47e`, 09-05 log head). The Windows box lost DNS at 12:22 ("could not be resolved: license2.ninjatrader.com", "api.apexinvesting.com"). That is a historical connection failure; current and future connection state were not rechecked here.
3. **`uptime -s` (13:58:49) and wtmp (`last reboot` 14:18) disagree by 19 minutes** on the 09-03 WSL boot — the clock-skew class in a new place.
4. **The daily digest appends the error line twice** — `if el := telemetry.ErrorDigestLine(at.id)` is pasted twice at `trader/auto_trader_planner.go:2528-2533`; `q31` shows "errors today: 1 …" duplicated.
5. **`decision_records.cycle_type` is NULL on 588 of 596 rows on 09-03** (only `watch` × 8 populated; `q45f`) — the column an EOD page would group by is unwritten for normal cycles.
6. **`ab_confirm_log.fill_px == entry_px` on all 10 latest real rows** (`q45i`) — it is not a slippage instrument; the confirm rule stamps the trigger price as the fill.
7. **`nofx-web.service` (vite dev server on :3000) is active** in production beside the Go-served `web/dist` (`q14` area; `systemctl is-active nofx-web` → active).
8. **A P0 "AI CREDIT EXHAUSTED — no decisions until topped up" at 07:34 on 09-04** (alert 643) and a P0 "NY planner fail-closed — NO-TRADE" at 08:11 (646), both unacked (`q45j`).
9. **The E8 counterfactual block cannot compute a mean for ASIA/LONDON/NY touch cells** ("NO USABLE net_pnl in 7/5/18 rows — uncomputed zero") and flags every short-bearing cell `short_suspect` (`q22` head) — the panel renders these honestly, which is the right outcome, but the table is 60% caveat by area.
10. **Arm working ages**: cancelled n=50 avg 27.8 min (max 254.5), filled n=9 avg 167.1 min (5.5–566.5) (`q45c`). These are created-to-last-update ages, not verified placement-to-fill latency; filled-row average is almost three hours; the stale-working window is 15 m (`⚔️ armed_orders … stale_working=15m`) — 8 cancels on 09-04 alone were `no order_update within stale window`. Worth a wave of its own; not this section's.
11. **The Guide's pre-market checklist** tells the owner to check "Guardrails master: intentional position (ON/OFF) — noted?" — OFF messages recur in the sampled four-day logs ("⚠️ Strategy Studio: risk guardrails master OFF", `q07`, `q20`).
12. **`day_plan_alerts` level P2 has 3 rows all-time, all acked** (`q08`) — the "digest" tier is unused.

---

## APPENDIX — every query and where its output lives

All scripts and outputs: `~/nofx-analysis/vet-04-0905/`; the committed copies (including q48 verification) are under `docs/superpowers/reports/2026-09-05-vet-04-monitoring-data/`. Store URI everywhere: `file:/home/hoang/nofx/data/data.db?mode=ro`.

| File | What it holds |
|---|---|
| `q01_store_asof.sh/.out` | row counts + max created_at per table |
| `q02_health.out` | `/api/health` payload |
| `q03_dev_vs_running.out` | `git log 36648655..origin/dev` |
| `q04_api_expectancy.json`, `q04_api_cutover-gate.json`, `q04_api_risk_gate-blocks.json` | API payloads (config/resolved kept in scratch only — 26 KB) |
| `q05_outage_0903.sql/.out` | decision_records around the gap; gaps > 10 min |
| `q06_outage_sources.sql/.out` | snapshots per day, equity-snapshot gaps, alert kinds, telegram binding |
| `q07_logevents_alerts_0903.sql/.out` | log_events unit check, alerts 09-03 CT, epoch-mix count |
| `q08_logevents_gap_ack.sql/.out` | the 113.86-min gap; ack rates; feed-down bodies; alerts per day |
| `q09_log_0903_postboot.out`, `q10_log_0903_link.out`, `q11_log_0903_reconnect.out` | Go log 09-03 14:18→15:02 |
| `q12_journal_hello.out`, `q13_journal_volume.out`, `q14_journal_persist_timers.out` | journald: frame flood, retention, user timers, nofx.service |
| `q15_boot_lines.out`, `q16_boot_block_0904.out` | boot-line sources; the 13:25:47 boot block |
| `q17_0904_feed.sql/.out`, `q18_log_0904_feed.out`, `q19_bars_per_hour.out`, `q20_log_0904_1219.out` | 09-04 feed death from the store and Go log |
| `q21_checklist_count.out` | AUDIT-CHECKLIST class count |
| `q22_expectancy_rows.out`, `q22b_expectancy_labels.out` | expectancy rows with Wilson/means, ids |
| `q23_config_resolved.out` | config/resolved summary |
| `q24_routes.out` | API routes |
| `q25_api_plan_today.json`, `q25_api_risk_status.json`, `q25_api_status.json` | API payloads |
| `q26b_strategy_mnq_guardrails.out` | Studio risk_control for strategy MNQ |
| `q27_log_0904_afternoon.out` | FEED DOWN first/last, skip lines |
| `q28_ledger_arms.out` | armed_orders latest rows + state counts |
| `q29_eod_closes.sql/.out` | EOD Block 1 |
| `q30_digests.out`, `q31_digest_samples.out`, `q32_digest_account.out`, `q33_digest_created.out` | digests: kinds, samples, writer times |
| `q34_fill_slippage.out`, `q35_exit_slippage.out` | fill/exit slippage series |
| `q36_deadman_0904.out` | dead-man lines 09-04 |
| `q37_mae_coverage.out`, `q38_entry_time_units.out`, `q39_era_cutoffs.out` | era rows, mae coverage, unit checks |
| `q40_gate_label_sets.out` | Go gate literals vs panel labels |
| `q41_backfill_skips.out` | excursion backfill lines; no_new_data per day |
| `q42_eod_inputs_0903.sql/.out` | touches, read facts, plans, refusals, counters, snapshot freshness |
| `q43_touches_readfacts.out` | touches per day/kind; read-facts null census |
| `q44_rev_diff.out` | `git diff --stat 36648655..origin/dev` for section files |
| `q45_store_batch.sql/.out` | bell count; RTH range vs ATR; arm ages; arms by reason; decision gaps; cycle_type; reads per session; refusal samples; ab_confirm fills; alerts 09-04 |
| `q46_log_eod.out` | boots per day with revs; FEED DOWN per day; crypto-vestige lines; track record; digest-written lines; first Go-side symptom 09-04 |
| `q47_host_nt8_evidence.out` | NT8 log boundaries 09-03/09-04; Windows System events 09-03; Kernel-Power 41 history; WSL boots; autostart census; NT8's view of the 09-04 feed death |

Commands not in a file (each run once, output quoted inline above): `grep -c 'Alert(\|PlaySound' ninjascript/*.cs` → 0; `grep -rn "telegram\." trader/ kernel/` → none; `grep -rn "risk/status" web/src` → none; `sed -n 626,632p api/server.go` (health handler); `sed -n 178,233p api/handler_risk.go`; `sed -n 100,130p kernel/engine_analysis.go`; `sed -n 2432,2545p trader/auto_trader_planner.go`; `sed -n 525,560p; 2350,2400p ninjascript/VLTraderTCPClient.cs`; `sed -n 593,640p ninjascript/VLBarsSubscriptionManager.cs`; `systemctl is-enabled nofx`; `grep systemd /etc/wsl.conf`.

## FINAL VERIFICATION AND STATISTICAL RECEIPTS

This continuation preserves the incoming draft in original scratch and supersedes its stronger causal language and entry-date EOD recipe. The retained inventory is source analysis; unobserved historical UI behavior remains inference. “Nobody watched,” guaranteed recovery times, human-career anecdotes, and new-channel proposals are not findings of this report. q48 is authoritative for the 15:00 page; q01–q47 are the earlier audit captures. Numeric log excerpts describe what the product printed, not independently validated performance. In particular the boot track-record win rate and PF are quoted UI/log output, not accepted performance estimates.

The post-Aug15 inventory and the eligible performance population are distinct. q48's exclusive full-table categories (with every ID): UNRESOLVABLE n=7, eligible n=58, null_pnl n=3, pre_aug15 n=516, test_seam n=3. The 516 pre-era rows are not pooled into current performance. NULL and UNRESOLVABLE are not zero returns.

| Recorded acknowledgement population | k/n | Wilson 95% |
|---|---|---|
| P0 | 62/152 | [33.30%, 48.74%] |
| P1 | 160/500 | [28.06%, 36.21%] |
| P2 | 3/3 | [43.85%, 100.00%] |

MAE/MFE inventory coverage: 70/71, Wilson 95% [92.44%, 99.75%], IDs in q48. These are descriptive census intervals, not independent-trial inference or alert-delivery probabilities. Touch percentages use the per-kind n and Wilson in q48; the whole-day VWAP example 61/134 is historical context, not a 15:00 measurement.

Expectancy cells below reproduce captured API estimates; all per-cell IDs are in q04_api_expectancy.json. These cells are overlapping descriptive classifications, not a new P&L recomputation or trading recommendation.

| Condition | wins/n | Wilson 95% | IDs |
|---|---|---|---|
| {'condition': 'acceptance'} | 1/6 | [3.01%, 56.35%] | [522, 526, 527, 532, 542, 543] |
| {'condition': 'breakout_retest'} | 1/9 | [1.99%, 43.50%] | [528, 534, 556, 557, 558, 562, 563, 589, 590] |
| {'condition': 'hold'} | 1/1 | [20.65%, 100.00%] | [555] |
| {'condition': 'reclaim'} | 0/5 | [0.00%, 43.45%] | [525, 531, 540, 541, 588] |
| {'condition': 'reject'} | 14/31 | [29.16%, 62.23%] | [521, 523, 524, 529, 533, 535, 536, 537, 538, 544, 547, 548, 549, 550, 551, 552, 553, 554, 560, 564, 565, 567, 569, 570, 575, 581, 582, 584, 585, 587, 591] |
| {'condition': 'sweep_reclaim'} | 1/6 | [3.01%, 56.35%] | [559, 561, 568, 578, 583, 586] |
