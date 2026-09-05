# VETERAN DEEP REVIEW — SECTION 03: DECISIONS

*Owner: hoang · 2026-09-05 (Saturday, CME closed) · READ-ONLY · branch `docs/vet-03-0905` · session `vet-03-0905`*

I am the thirty-year index-futures trader the dispatch asked for, and I had the live store and the API this time. Every number below was recomputed here against `data/data.db` (mode=ro) or the running process's own log; where I quote a prior report I say so and I say whether I could reproduce it. Evidence labels: **[A]** directly verified, **[B]** inferred, **[C]** speculation; market beliefs **[R]** named study, **[T]** own-tape number with n, **[I]** my experience, untested here. All times CT.

---

## EVIDENCE BASIS

- **Worktree base:** `2a66d91c` (2026-09-05 07:22:45 -0500, "Merge #91: wire RiskForceFlat and BiasArmWarning"), cut detached from `origin/dev`; claim commit `8a8625c4` on `docs/vet-03-0905` (2026-09-05T09:54:43-05:00).
- **Running binary:** rev `36648655cfe0` — `/api/health` → `{"revision":"36648655cfe0","status":"ok"}` (200, 09:5x CT); `deploy/RELEASE` = `36648655`; boot line `09-04 13:25:47 🔐 BOOT INTEGRITY OK — rev 36648655cfe0 · built 2026-09-04T18:22:30Z` (`data/nofx_2026-09-04.log`). **[A]**
- **dev is ahead of the running binary by two merges:** #90 `b70b7d2f` (mcp ctx data races, `mcp/client.go`) and #91 `01ce8088` (`trader/entry_gate.go` +26 lines = Leg D daily force-flat; `kernel/risk_limits.go` +65; `kernel/engine_analysis.go`; `trader/auto_trader_planner.go` +14 = BiasArmWarning call). `git diff --stat 36648655 origin/dev`: 18 files, of which only those four are code that bears on this section. Where the running binary differs from dev I say which I am reading. **[A]**
- **Spec-freshness (`git log -1 --format='%h %ci %s' -- <path>`):**
  - `docs/superpowers/SYSTEM-MAP.md` — `a96224dd 2026-09-04 09:07:37 -0500`
  - `docs/superpowers/AUDIT-CHECKLIST.md` — `15340faa 2026-09-04 13:22:07 -0500`
  - `docs/superpowers/research/INDEX.md` — `4e8e7e1a 2026-09-03 19:37:14 -0500`
  - `docs/superpowers/reports/2026-09-05-veteran-review.md` — `676f239c 2026-09-05 05:51:29 +0000`
  - `docs/superpowers/reports/2026-09-04-two-day-audit.md` — `f3c640c3 2026-09-04 07:26:52 -0500`
  - `docs/superpowers/reports/2026-09-04-research-conformance.md` — `790efbb3 2026-09-04 09:20:28 -0500`
  - `docs/superpowers/reports/2026-09-02-bias-calibration.md` — `2deab3c8 2026-09-02 20:53:20 -0500`
  - `docs/superpowers/reports/2026-09-02-live-bias-replay.md` — `53498adb 2026-09-02 21:02:58 -0500`
  - `docs/superpowers/reports/2026-09-02-belief-census.md` — `ee64a494 2026-09-02 08:50:38 -0500`
  - `docs/superpowers/reports/2026-09-03-mc-drawdown.md` — `77e1cdfc 2026-09-03 00:39:25 -0500`
  - `kernel/planner_prompt.go` — `0e016635 2026-09-04 09:32:21 -0500` · `kernel/plan_doc.go` — `fd3fadcd 2026-09-04 07:29:55 -0500` · `trader/entry_gate.go` — `01ce8088 2026-09-05 12:12:00 +0000` · `trader/armed_executor.go` — `51916172 2026-09-04 09:48:03 -0500` · `kernel/min_sl.go` — `4657560b 2026-09-02 07:33:39 -0500` · `kernel/risk_limits.go` — `01ce8088` · `kernel/regime.go` — `7b19e753 2026-09-03 23:00:08 -0500` · `kernel/plan_confirm.go` — `60faefbd 2026-08-30 16:12:05 -0500` · `kernel/entry_law.go` — `575e9c05 2026-09-02 19:01:33 -0500` · `trader/exit_mechs_suspend.go` — `4657560b`.
  - `docs/superpowers/plans/VL-MASTER-PLAN-v2.md` — **empty log line: the file has never existed in this repository.** Not cited.
- **Store as-of** (`~/nofx-analysis/vet-03-0905/q01_store_premises.out`): trader_positions 587 rows (max created_at 1788444314627 = 2026-09-03 09:05 CT) · armed_orders 67 (max 2026-09-04 12:11:02 CT) · plans 254 (max 2026-09-04 17:10:35 UTC) · plan_lifecycle_log 7 · touch_outcomes 677 · candidate_pool 360 · trade_excursions **0** · decision_records 37,768 (max 2026-09-04 18:28:01 UTC) · ab_confirm_log 223 · nt8_order_snapshots 4,461 (the dispatch's 4,433 was 09:41 CT; the count moved) · bars 79,769 (last 1m bar 2026-09-04 12:19 CT) · planner_rejected_prompts 64 · planner_read_facts 32 · level_state 675 · level_stats 234 · touch_episodes 1,315 · trader_fills 433.
- **API read (all 200, token via `cmd/gate-jwt`, never printed):** `/api/health`, `/api/expectancy` (default and `?by=session`), `/api/config/resolved` (26,495 B; knobs+summary — summary: schema 57, classified 167, live 144, ineffective 7, candidate-unverified 16), `/api/risk/gate-blocks`, `/api/cutover-gate`. Saved under `docs/superpowers/reports/2026-09-05-vet-03-decisions-data/` except the config dump (it carries knob names such as `api_key`; nothing secret, but I kept it out of the repo).
- **Logs:** `data/nofx_2026-09-01..04.log`, grepped by the `MM-DD HH:MM:SS` prefix inside the files (rotation-on-boot trap respected; the 09-04 morning lives in `nofx_2026-09-03.log`).
- **Scratch:** `~/nofx-analysis/vet-03-0905/` (q01–q14); the decisive outputs are committed beside this report.

---

## ONE-PAGE SUMMARY

**Verdict for this section: the decision layer is three authors fighting over the same three numbers, refereed by a gate chain that is doing its job, on a plan that is rewritten every half hour.** The planner (LLM) writes entry, stop and target; `composeArmStop` rewrites the stop; the R:R leg judges the rewritten pair against the planner's own target; the executor AI is told it may set its own take-profit. Under `plan_mode=strict` the executor cannot enter at all, yet it is still called every two minutes (2,332 calls since 09-01, p50 16.7 s, p90 51 s) to produce intents that leg 0 refuses. The realized book says the same thing the code does: the AI decision path is **−$750.43 on 51 trades (13W/36L)**, the armed-fill lineage is **+$42.50 on 9 (5W/4L)**, and the whole era is **−$563.93 on 65 resolved trades, mean −$8.68**, indistinguishable from zero. The gates are not the problem; the refused set loses money on the tape. The plan is the problem — its shape, its churn, and who owns its numbers.

**Three biggest problems** (each measured here):

1. **The plan is a rolling opinion, not a plan.** Since 09-01, 79 plan versions over 11 session-days; **53 of 79 (67.1%, Wilson [56.1%, 76.4%]) were level-event re-plans**; median time between versions **35.7 min** (n=68); the median scenario is in force for **32 min** before it is superseded (n=592). Only 4 of the 11 session-days carry a version authored by the scheduled read; 4 carry a `planner_fail_closed` marker. **[A]** `q10_churn_durations.out`. On NFP morning (09-04) the NY read started 08:00:38, was rejected three times, **fail-closed at 08:11:45**, the open arrived with a `no_trade` v1, the process restarted at 08:30:11, and the owner reset the plan by hand at ~08:44. **[A]** log.
2. **The stop composer and the R:R floor flap against each other, at the broker, every two minutes.** 09-04 NY S2 (stop-entry reclaim short 29591.02 / target 29481.5 / authored stop 29645.25 = R:R **2.019**): each cycle `composeArmStop` widened the stop by the live 1.5×ATR5m (29645.62 → 29646.99 → 29647.87), the arm gate cancelled "gate changed: rr", the next cycle re-armed at the authored stop, and so on — **20 ledger rows (ids 62–102) in 42 minutes, 20 real broker placements** (nt8_order_snapshots 10:00–11:00: 242 snapshots, 203 with a working order), **0 fills**. **[A]** `q08…out`, `q13…out`, log 10:05–10:53. The two-day audit's D37 ("all 7 R:R-at-arm refusals were authored ≥2.0 and pushed below by composeArmStop") is not a corner case; it is the steady state for any arm authored within ~0.3 of the floor, which is **47 of 132 arms (35.6%)**. Nothing on dev fixes it (`git log --since=2026-09-04 -- trader/arm_stop_anchor.go trader/armed_executor.go` shows the reaper fix only).
3. **The executor is paid to be refused.** 31 open intents in 4,102 cycles since 08-27 (0.76%); 19 of the 31 were refused by a gate; since strict entered force (commit `c8c90dcc`, 09-03 10:43) **13 of 14 decision-path intents were refused by leg 0** — the same ASIA short re-proposed thirteen times between 20:35 and 21:12 on 09-03. The path's only surviving contribution is P&L −$750.43. **[A]** `q04…out`, `q11…out`.

**Three biggest opportunities:**

1. **One scenario, machine numbers.** The LLM's first idea is the only one with a positive record: **S1 n=27 +$321.00 (mean +$11.89, 10W/17L)** versus **S2–S4 n=31 −$787.43 (8W/21L/2 scratch)**. Let it author the narrative, the level shortlist, the invalidation and ONE scenario; let the tape compose entry (the level), stop (already `composeArmStop`) and target (from the store's own MFE distribution: median MFE **0.67R**, p75 **1.79R**; a 2R target is reached **13/57 = 22.8% [13.8%, 35.2%]** of the time against each trade's own authored stop). Judge R:R at write on the composed stop, so the plan that clears the floor at authoring clears it at arm. This kills the flap, the reject loop (44.8% of authoring attempts since 09-01 are rejected, 59.4% of those by the continuation law the prompt itself invites), and the R:R-gaming incentive in one move.
2. **Symmetric arm eligibility, and a direction-mix validator that does not exist yet** (`grep -rn 'direction mix' kernel/` is empty). Arm-enabled share: **short 96/395 (24.3%) vs long 36/343 (10.5%)** across all versions; on the latest version per session-day **long 4/56 (7.1% [2.8%, 17.0%]) vs short 14/72 (19.4%)**. The 09-04 "make reclaim armable to rescue longs" change produced **21 reclaim-family stop-entry arms in the store, all SHORT, zero long**. The executable book is whatever can rest a limit, and that is a fade at resistance.
3. **Put a human at the one place it pays: before the open, on one card, with three buttons — approve · veto the session · re-read — and a halt.** Nothing mid-session. ASIA alone is **−$552.43 on 16 (2W/13L, 13.3% [3.7%, 37.9%])**; the code default already disables it (`kernel/session_registry.go:93 Enabled: false`) and the saved strategy re-enables it. That is a judgment call no model made and no gate can make.

---

## PREMISES CORRECTED

Every statement in the dispatch or a prior report that I found wrong or unmeasurable, with the query that shows it.

1. **"trader_positions: 227 rows have entry_time ≥ 2026-08-15; 223 pnl_corrected non-NULL; 227 have plan_id; 64 cited_scenario_id."** Wrong on three of four. `q01`: `SUM(entry_time >= strftime('%s','2026-08-15 05:00:00')*1000)` = **71**, of which 67 have `pnl_corrected`, **71 have a plan_id (all 71 plan-linked rows in the table are dated 2026-08-19 or later)**, 64 cited. `q02` by month: 2026-08 = 242 rows (218 pnl_corrected, 60 plan_id), 2026-09 = 11. No cut of this table yields 227; rows since 08-09 = 201, since 08-01 = 253. The 64-cited figure is right. `trade_excursions` = 0 rows confirmed; `mae`/`mfe` are populated on **70 of 71** plan-linked rows, not "66 of 227".
2. **"AUDIT-CHECKLIST.md: 79 classes."** `sed -n 11,1845p | grep -cE '^[0-9]+\. \*\*'` = 78 numbered entries, max number 79, **number 46 is skipped**; `## CLASS 75` at line 2002 restates 75. **78 distinct classes.**
3. **"2-minute AI executor loop."** The dispatch is right for the live configuration and the prior veteran review's correction ("not 2-minute: default 3, minimum 3, `store/trader.go:29`") is wrong about the live value: boot line `09-04 13:25:47 📊 Loading trader hoang: ScanIntervalMinutes=2 (source=Studio/DB), cadence=interval 2m0s` and `⚙️ Scan interval: 2m0s`. The store default is 3 (`store/trader.go:29` `default:3`) and the saved row holds 2, so "minimum 3" (SYSTEM-MAP §12) is not enforced on the stored value. **[A]**
4. **The dispatch's gate order** ("strict → one-open-position → invalidation → direction → shadow → R:R at fill → min-SL → daily-limit leg → no-chase warn") **is not the code's order.** At dev (`trader/entry_gate.go:146-320`): Leg D daily force-flat (:169-173, **dev only, #91**) → 0 strict (:184-197) → 1 plan bias, direction mode only (:200-207) → 2 scenario-direction, class 48 (:210-220) → 3 invalidation, ARM PATH ONLY (:229-252) → 4 shadow (:257-259) → 5 R:R at execution price (:261-283) → 6 min-SL ×ATR5m (:286-294) → 7 one-open-position (:301-307) → no-chase WARN (:312-318). **One-open-position runs seventh, not second. The running binary `36648655` has no daily leg at all** — #91's own comment: "a RESTING ARM FILLED STRAIGHT THROUGH A TRIPPED DAILY LOSS LIMIT". **[A]**
5. **`docs/superpowers/plans/VL-MASTER-PLAN-v2.md` and `kernel/exit_*` do not exist** (both confirmed: empty `git log -1`, no such files). Exit logic is in `trader/exit_mechs_suspend.go`, `kernel/min_sl.go`, `kernel/risk_limits.go`, `trader/arm_stop_anchor.go`; the boot line states the live exit contract: `🛑 exits: stop=max(anchor+clr, 1.5×ATR5m) · anchor_max=3.0×ATR5m · BE=off · trail=off · size=1 · re-arm-after-sweep=on (0B)`.
6. **"Bias: both bias systems calibrated non-predictive (weekly anti-predictive; machine tree 0.48–0.54)."** Verified in the reports, not re-derivable from the store: weekly control raw holdout hit **25–28%** (`2026-09-02-bias-calibration.md:97,118`); live legs holdout **0.48–0.54, Wilson lower bounds ≤ 0.40, n = 21–62** (`2026-09-02-live-bias-replay.md:159`); regime leg 0.5435 [0.4018, 0.6785] n=46 (`:118`). What the dispatch did not say: **`planner_read_facts.bias_ai` and `.bias_tree` are empty on all 32 rows** (`q03`), the label lives only in `plans.doc.bias_label`, which **27 of 253 versions carry** (all from 09-04); AI==tree on 21 of 27 (77.8% [59.2%, 89.4%]).
7. **Prior review: "only 3 of 36 trades produced enough MFE to reach a 2.0R target measured off the minimum stop."** Not reproducible here — the CSV cohort is not in the store and the denominator is the 1.5×ATR floor, not the trade. Against each trade's **own** authored stop (system trades: the `stop_loss` in the `open_*` decision within ±3 min of entry; armed trades: `armed_orders.stop_px`), **13 of 57 = 22.8% [13.8%, 35.2%] reached 2.0R; 22 of 57 = 38.6% reached 1.0R**; median MFE 0.67R, p75 1.79R; median R = 34.25 pts; 7 rows unresolvable (568, 569, 571, 575, 580, 582, 585). `q11_mfe_r_rows.csv`. Both numbers say the same thing — a 2R target is a minority outcome — but the number to quote is the trade's own.
8. **Prior review "S1 +$461 (n=26)".** Now **+$321.00 on n=27** — position 591 (−$140, 09-03) was added after that cohort was cut; S2–S4 unchanged at −$787.43. Consistent, superseded.
9. **Two-day audit: "the refused set loses $860.64 (session-flat) / −$1,036.52 (CME-day) over 44 distinct opportunities."** The committed `refusals.csv` (61 rows) sums to **−$726.14 session-flat / −$902.02 CME-day** (`q12`), and the CSV carries no distinct-opportunity key, so the 44-opportunity figure cannot be rebuilt from the committed data. Same sign, same conclusion; the quoted dollar figure is not reproducible from its own artifact. **Disagreement recorded, magnitude −$134.**
10. **Two-day audit: "Fourteen wired gate legs never fired … invalidation … refused nothing."** True for 09-02/03. **On 09-04 leg 3 fired twice** — `🚦 entry-gate REFUSED arm LONDON: entry_gate: scenario S2 invalidated at 2026-09-04 02:00 CT` (02:00:46) and `… arm NY: scenario S1 invalidated at 2026-09-04 09:00 CT` (09:02:09); counters `arm_refusals_0b:…:2026-09-04:LONDON:entry_gate:invalidated=1` and `…:NY:entry_gate:invalidated=1`. **[A]** Both counterfactuals reached TARGET (below).
11. **"113 min silent"** is the 09-03 host reboot (two-day audit D1, 12:24:33→14:18:24 CT) — a Section 4 premise, not a decision fault. One thing to record for whoever owns that section: **the bars table shows 60 bars in every hour of 09-03 including 12:00–14:00** (`q13` gg), because `SeedHistorical` backfills on reconnect; **bars cannot be used to detect an outage**, only logs and decision_records can.
12. **`/api/risk/gate-blocks` → "no gate blocks recorded".** This is an in-memory tally (`telemetry/gate_blocks.go:15-17` "intentionally in-memory and ephemeral … resets at the 17:00 CT CME session-day rollover"), also lost on restart. A zero here on a Saturday is not a measurement.
13. **Prior review: "5 of 15 arms died to the marketable guard; 3 of 15 filled."** Confirmed for the two-day window (ids 23–37). All-time on the store, excluding the 6 test rows (sessions TEST-E2 ×4, TEST-E7 ×2): **real fills 9 of 61 (14.8% [8.0%, 25.7%]); marketable-guard deaths 6 of 61 (9.8% [4.6%, 19.8%])**. The bigger killer all-time is the stale-window reaper: **12 of 61 (19.7% [11.6%, 31.3%])** ("no order_update within stale window"), 8 of them on 09-04 alone, before the reaper fix `51916172` (09:48 CT) went live at the 13:25 boot. (Ledger caveat: `UpsertArm` rewrites a working slot in place, so ledger rows under-report broker placements — the 09-04 flap alone put 20 orders on the book.) `q11`, `q03`.
14. **Prior review §3.1: "direction: long 56 (86.2%) · short 9" on 23 plans.** Window-specific. Across all 738 stored scenarios: **short 395 (53.5%), long 343**; latest version per session-day: short 72, long 56. The imbalance that matters is not authored direction, it is **arm eligibility by direction** (item 2 of the opportunities).

---

## 1 · THE DIVISION OF LABOR — drawn from the code

```
                 ┌───────────────────────── PLANNER (LLM, deepseek-v4-pro) ─────────────────────────┐
 clock ──────────┤ scheduled read at ReadCT: ASIA 16:30 · LONDON 01:30 · NY 08:00                      │
 (session reg.)  │   kernel/session_registry.go:90,99,108 → trader/auto_trader_planner.go:1883       │
 level-event ────┤ wake → runPlannerReadWithTriggerClaimedCtx("level_event")                          │
 (W6 wakes)      │   trader/auto_trader_wake_levels.go:376 (unlimited; spends no replan budget :24)   │
 death ──────────┤ death re-plan, capped replan_cap=4/session  auto_trader_planner.go:292            │
 MSS / fast-mkt ─┤ auto_trader_transition.go (SYSTEM-MAP §11)                                        │
                 │ prompt: kernel/planner_prompt.go (770 lines) — schema :697, scenario-mix law :707,│
                 │   entry law :721, "ARMS FOLLOW THE BIAS" :730, entry-type-by-condition :732,      │
                 │   "target_chain is GUIDANCE … executor AI sets the actual take_profit" :722       │
                 │ validators: kernel/plan_doc.go:588 ValidatePlanDocWithCaps (+entry_law.go,        │
                 │   plan_confirm.go, breakdown law) → REJECT at write → repair attempt 2/3, 3/3     │
                 │ output PlanDoc: bias{direction,conviction,flip_condition} · levels[≤12] ·         │
                 │   scenarios[1..cap]{trigger, condition, direction, target_chain[], invalid,       │
                 │   confirm{rule,ref_price,side}, quality, arm{enabled,entry,stop,target,           │
                 │   wait_confirm}} · death · flip · day_type · no_trade_windows · bias_label        │
                 └──────────────────────────────┬────────────────────────────────────────────────────┘
                                                │ ActivePlan (kernel/plan_render.go:103)
            ┌───────────────────────────────────┴──────────────────────────────────────────┐
            │  ARM SEAM (every 2-min tick)                     │  EXECUTOR (every 2-min tick)  │
            │  trader/armed_executor.go:188 maybeManage…       │  trader/auto_trader_loop.go:174│
            │  for each scenario with arm.enabled:             │  runCycle → prompt with PLAN   │
            │   kind ← kernel/arm_kind.go:48 (limit) / :58     │  BLOCK (engine_prompt_futures  │
            │     (reclaim → stop_entry, 09-04)                │  .go:149) + cited_scenario     │
            │   stop ← composeArmStop (arm_stop_anchor.go:71): │  REQUIRED (:252) → AI JSON     │
            │     max(anchor+2 ticks, 1.5×ATR5m), never        │  {action, stop_loss,           │
            │     tighter than authored                        │  take_profit, …}               │
            │   wait_confirm ← kernel.EvaluateConfirm          │  validateDecision              │
            │   armGateVerdictFor (:1349): spec valid ·        │  (kernel/engine_position.go:43)│
            │     armable · bias(direction mode) · quality ≥   │  incl. min-SL (kernel/min_sl.go│
            │     min_scenario_quality=C · R:R ≥ arm_rr 2.0 ·  │  :56) + confidence ≥ 60        │
            │     min-SL · HTF veto                            │  → entryGateForDecision        │
            │   entryGateForArm (entry_gate.go:352)            │    (entry_gate.go:413)         │
            │   place iff |price−entry| ≤ ARM_PLACE_TICKS×tick │                                │
            │     = 25 pts (:904/:964 per SYSTEM-MAP §5)       │                                │
            │   reaper: broker book (51916172, live in boot 10)│                                │
            └───────────────────┬──────────────────────────────┴───────────────┬───────────────┘
                                │                                              │
                     ┌──────────┴──────────── EntryGate (ONE function, both seams) ┴──────────┐
                     │ trader/entry_gate.go:146  D daily(dev) → 0 strict → 1 bias(dir) →      │
                     │   2 scen-dir → 3 invalidation(arm) → 4 shadow → 5 R:R@fill →           │
                     │   6 min-SL → 7 one-open → no-chase WARN                                 │
                     └─────────────────────────────────┬───────────────────────────────────────┘
                                                       │ NT8 SIM over TCP (trader/ninjatrader/tcp_trader.go)
                                          exits: NT8 bracket stop/target · EOD flat 14:45 · BE/trail OFF
```

**What the store says each box actually did** (all `[A]`, plan-linked rows, `pnl_corrected` only, test seam excluded):

| box | measured | n | source |
|---|---|---|---|
| Planner: versions since 09-01 | 79 versions / 11 session-days; level_event 53 (67.1%), dormant/rearm 12, fail-closed 7, owner_reset 3, scheduled 4 | 79 | `q10a/q10c` |
| Planner: time between versions | median 35.7 min, mean 60.8, min 6.5, max 400.9 | 68 gaps | `q10b` |
| Planner: scenario in-force window | median 32 min (p25 22, p75 56) | 592 | `q06b` |
| Planner: authoring attempts rejected since 09-01 | 64 of 143 = 44.8% [36.8%, 52.9%] | 143 | `q07` |
| Executor: cycles since 08-27 | 4,102; open intents 31 (0.76% [0.5%, 1.1%]); 19 refused | 4,102 | `q04` |
| Executor: AI latency since 09-01 | p50 16,691 ms · p90 51,076 ms · max 600,000 ms | 2,332 | `q10d` |
| Executor: closes it authored | 2 `close_*` decisions since 08-19; 65 of 68 closes are `close_reason=sync` (NT8 stop/target/EOD) | 68 | `q14` |
| Decision-path trades (source=system) | −$750.43 · 13W/36L · win 26.5% [16.2%, 40.3%] | 51 | `q11` |
| Armed-fill lineage (reconcile + plan_band=armed_fill) | +$42.50 · 5W/4L | 9 | `q11` |
| Era total (65 resolved, 3 NULL excluded: ghosts 576/577/579) | −$563.93 · mean −$8.68 · 21W/42L/2 scratch · avg win $114.67 · avg loss −$70.76 · payoff 1.62 · win rate ex-scratch 33.3% [22.9%, 45.6%] · breakeven at 1.62 = 38.2% | 65 | `q09f`, `q11` |
| By slot | S1 +$321.00 (n=27, 37.0% [21.5%, 55.8%]) · S2 −$142.43 (22) · S3 −$467.50 (7) · S4 −$177.50 (2) · off-plan −$71.50 (2) · uncited −$26.00 (5) | 65 | `q09b` |
| By session | ASIA −$552.43 (16; 2W/13L) · LONDON +$24.00 (21) · NY +$62.00 (21 resolved) · none −$97.50 (7) | 65 | `q09d` |
| Holding time | median 26.0 min (p25 12.2, p75 55.4, max 219.7) | 65 | `q14` |

**Would I draw it that way? No, and for three concrete reasons, each with a line.**

*(a) Three authors of the same three numbers.* The planner writes `arm.entry/stop/target` (`plan_doc.go` scenario schema; the prompt demands "EXACT entry/stop/target" `planner_prompt.go:733`). `composeArmStop` overwrites the stop every tick ("never tighter than authored", `arm_stop_anchor.go:71`; log `🛑 arm stop NY S2 leg 1 short: stop 29645.62 (authored 29645.25 WIDENED)`). The R:R leg then judges the machine's stop against the model's target (`entry_gate.go:261-283`; `armGateVerdictFor` "R:R 1.87 below arm min 2.00"). And the executor prompt tells the AI that "`target_chain` is GUIDANCE for the executor AI (which sets the actual take_profit)" (`planner_prompt.go:722`). Nobody owns the trade's geometry. **[A]** The 09-04 flap (§4, leg 5) is what that looks like at the broker.

*(b) The plan is re-authored faster than a scenario can play out.* Median in-force window 32 min against a median holding time of 26 min and a trigger proxy that needs the whole session to fire 75.6% of the time but fires 44.4% inside the in-force window (§2). A "plan" whose scenarios are replaced before their trigger fires is not a plan; it is a commentary stream. **[T]** n=592.

*(c) Under strict, the executor is a cost centre with no output.* Leg 0 refuses every decision-path market entry (`entry_gate.go:184-186`); the executor is still called every two minutes with the full plan block; its p90 latency is 51 s; and the one thing it still does that reaches the store — propose `open_*` — was refused 13 of 14 times since strict entered force, all thirteen the same ASIA short between 20:35 and 21:12 on 09-03. **[A]** `q04`. The only executor outputs that matter under strict are `close_*` (2 in seventeen days) and the citation stamp for adherence grading.

**[I]** On my desks the analyst never touched the order ticket, the trader never rewrote the analyst's levels, and the risk desk touched neither — it only said no. This architecture has the analyst writing the ticket, the trader rewriting it, and the risk desk arguing with both every two minutes.

---

## 2 · WOULD I LET AN LLM AUTHOR PLANS AT ALL?

**Yes — and the store says what it should and should not author.**

### 2.1 What the planner is good at, measured

- **Its first idea.** S1 is the only slot with a positive record: **+$321.00 on 27 (10W/17L)**; S2–S4 are **−$787.43 on 31 (8W/21L, 2 scratch; win rate ex-scratch 27.6% [14.7%, 45.7%])**. The n is thin in the tail (S3 7, S4 2) and I flag it, but the sign flips after S1 and stays flipped. **[T]** `q09b`. **[I]** This is exactly what I expect of a model asked for a list: the first item is its read, the rest is schema-filling.
- **Its narrative is grounded.** `reasoning` is required (`plan_doc.go:600`), the bias-tree branch must be named, and `confirm.ref_price` must match a number in the trigger/invalid prose (`plan_doc.go:669`). That is the right contract for prose.
- **Half its levels are its own.** On the 10 plan versions since 09-01 that have a seated `candidate_pool`, **54 of 105 plan levels (51.4%) match a seated candidate within 1 pt** — the other half the model added or shifted. **[A]** `q13` (qq). Whether the added half is good is Section 2's question, not mine; the point here is that "the levels come from the machine" is not true today.

### 2.2 What it is bad at, measured

**Reject rate by rule since 09-01** (`planner_rejected_prompts` began 09-01 with class 38, so this is the only honest window; `q07`):

| family | n | share |
|---|---|---|
| continuation law (breakdown/breakup void · displacement 0.00 pts · BD_MIN_CLOSES) | **38** | **59.4% [47.1%, 70.5%]** |
| transport (503 Server Overloaded ×3, stream deadline ×2, EOF ×1) | 6 | 9.4% |
| `fade_requires_touch` (reject authored with a close-confirm) | 5 | 7.8% |
| "arm requires entry_mode=pullback" — a rule that first appears 09-04 | 5 | 7.8% |
| arm-legs contract (split/legs) | 4 | 6.3% |
| confirm-rule vocabulary (`2x5m`, `displacement` not in enum) | 2 | 3.1% |
| gap-up trigger law | 2 | 3.1% |
| retest distance > 5×ATR · level cap (13 > 12) | 1 + 1 | 3.1% |

**64 rejects against 79 accepted versions = 44.8% [36.8%, 52.9%] of authoring attempts rejected.** Attempt distribution: 30 first attempts, 20 second, 14 third (`q03`). Three of five rejects are the model authoring `breakdown_continue` on a tape where the machine says the breakdown is void — the same class-45 "prompt feeds forward" loop that was supposedly closed on 09-02, still running on 09-04 (`S1 breakdown_continue: the tape shows NO confirming close beyond 29481.50 yet`, 3 rejects 09-04). **[A]**

**Re-author cost in seconds** (first rejected attempt → next `plans` row for the same session-day; `q07`): **n=26 clusters, median 286 s, min 27 s, max 5,185 s; 6 of 26 ended as `planner_fail_closed` (no plan at all); 2 clusters produced no subsequent row.** The 09-04 NY morning is the worked example: read 08:00:38 → attempt 1 rejected 08:07:20 (`confirm2.rule "touch" not allowed`) → attempt 2 rejected 08:09:03 (breakdown law) → attempt 3 rejected 08:11:45 (`arm requires entry_mode=pullback`) → `🚨 PLANNER FAIL-CLOSED 2026-09-04 NY` → `PLAN written NY v1 lifecycle no_trade` → NY opens 08:30 with no plan → process restart 08:30:11 → `owner_reset` v2 at ~08:44. On NFP day. **[A]** `nofx_2026-09-03.log` (rotation trap: the 09-04 morning lives there).

**Share of scenarios whose trigger ever fired** (plans ⋈ 1m bars, `q06b`). Definition, stated because it matters: a scenario "fires" when any 1m bar's [low, high] contains `confirm.ref_price` — a touch proxy, looser than the confirm rule itself; 128 scenarios carry no `ref_price` (45 of them `hold`) and 18 have no bar coverage, so they are excluded.
- **In-force window** (from the version's creation to the next version or session end): **263 of 592 = 44.4% [40.5%, 48.5%]**; by condition: reject 84/198 (42.4%), sweep_reclaim 58/167 (34.7%), breakout_retest 54/109 (49.5%), reclaim 30/47 (63.8%), acceptance 24/42 (57.1%), hold 11/26.
- **First version per session-day, window to session end:** **62 of 82 = 75.6% [65.3%, 83.6%]**.
The gap between those two numbers is the churn: the level is usually reached during the session, but the version that named it has usually been replaced by then.

**The planned reward is chosen by the party the floor is checked against.** Arm-enabled scenarios, all versions, n=132: planned R:R median **2.36**; bins **<2.0: 15 · [2.0, 2.1): 19 · [2.1, 2.3): 28 · [2.3, 3.0): 44 · ≥3.0: 26** → **47 of 132 (35.6%) sit within 0.3 of the 2.0 floor** (`q13` rr). Realized payoff **1.62** (avg win $114.67 / avg loss $70.76). The prior review's "reward side is fiction" holds on this store, with the number to quote being 2.36 planned vs 1.62 realized, not 2.55 vs 1.66. **[T]** n=132 / n=63.

**Scenario mix, all 738 stored scenarios** (`q06`): reject 234 · sweep_reclaim 186 · **breakout_retest 124 (16.8%) — shadowed by default (`kernel/condition_status.go:26-28`), authored, scored, never placed** · hold 72 · reclaim 63 · acceptance 56 · breakdown_continue 3 (of the dozens authored, 3 survived validation). Quality: B 460 (62.3%), C 160, A 117, A+ 1 — the grade does not discriminate. Arm-enabled 132/738 (17.9%): reject 96, sweep_reclaim 25, breakout_retest 10 (dead on arrival), reclaim 1.

### 2.3 My answer: what the LLM authors, what is mechanical

**The LLM authors:** (1) the read — regime, day type, the bias-tree branch it took, in prose; (2) a level shortlist with a role and a reason for each (react / target / invalidation), drawn from the machine's seated pool, with any addition flagged as the model's own; (3) **one primary scenario** (condition, direction, the level, the invalidation line in prose) and at most one alternate on the other side for a `neutral` day; (4) the no-trade reasons it sees that the calendar does not. That is reading a situation, which is what the model is for.

**Mechanical, from the tape, with no model input:** (1) entry = the level (already the arm's entry); (2) stop = `composeArmStop` as it stands (anchor beyond nearest seated level + 2 ticks, floored at 1.5×ATR5m) — this exists and is right; (3) **target = a rule from the store's own MFE distribution**, e.g. the level chain's next seated level capped at the empirical p75 MFE (1.79R) — not a number the model chose; (4) **R:R judged once, at write, on the composed stop and the composed target** — a plan that clears the floor at authoring clears it at arm, and the flap of §4 cannot occur; (5) direction-mix and arm-symmetry validators (§3); (6) the confirm rule per condition (already the entry law). The model never writes a price it is judged against.

**What this removes:** the R:R-gaming loop (the model cannot inflate a target it does not write), 59% of the reject loop (no continuation prices to void), the three-author stop, and the S2–S4 tail that has cost $787.

**Evidence I would demand before believing any of it works:** the same three numbers on the new contract, pre-registered — S1-only expectancy with its CI, planned-vs-realized payoff gap, reject share — at n ≥ 100 trades for a sign and, per the MC report I reproduce (mean −$8.68, sd ≈ $101, se $12.55 at n=65), **~1,000 trades to separate −$8.68 from zero** (`2026-09-05-veteran-part-b.md:42`, and my own n=65 mean matches to the cent). Nobody should promise a P&L verdict from this desk before Q1 2027 at one contract; what can be verified in weeks is the mechanism — zero flaps, zero fail-closed opens, reject share under 15%.

**[R]** The literature does not settle the LLM question; the relevant part is the level-and-target literature: Osler (2000, 2003) on round-number clustering and stop/take-profit asymmetry, and the volume-at-price work — targets at the next liquidity pool, not at 2R. **[I]** Every discretionary trader I ran learned to write the trade's stop and target before the entry, from the chart, and never to argue with them afterwards; the model should be held to the same rule by not being asked.

---

## 3 · BIAS

### 3.1 What bias does mechanically today (running rev unless noted)

Every consumer of `Bias.Direction` (`grep -rn 'Bias\.Direction' kernel trader`, `q` in the appendix):

| consumer | effect | live under strict? |
|---|---|---|
| `entry_gate.go:200-207` leg 1 | refuses an entry against bias — **direction mode only** | no (plan_mode=strict) |
| `armed_executor.go:430` → `armGateVerdictFor` "against plan bias (plan_mode=direction)" | same, arm seam | no |
| `auto_trader_planner.go:1635` | after a flip fires, the re-plan's bias is **MANDATORY** (reject otherwise) | **yes — mechanical** |
| `kernel/arms_bias_coherent.go:74` `BiasArmWarning` | WARN if the bias side carries no armable scenario | dev only (#91); warn-first — a hard reject would have refused 50/68 longs and 66/103 shorts (`:9-10`) |
| `plan_render.go:157` "Bias: %s (%s) · flips" | the executor AI reads it every cycle | yes (advisory) |
| `planner_prompt.go:183-189` branch 5 | "longs disallowed by branch 5 (premium)" / "shorts disallowed (discount)" in the machine bias-tree the planner must cite | yes — as prompt law |
| `planner_prompt.go:730` "ARMS FOLLOW THE BIAS" | scenarios in the bias direction SHOULD carry arms | yes — as prompt law |
| `rootfix_shadow_ab.go:133` | shadow A/B requiredBias | OFF (boot: `shadow A/B: OFF`) |

So under strict the bias is not a gate; **it acts through which scenarios the planner arms**, and that is where the money is: bias.direction across all versions is **short 109 / neutral 73 / long 71**, and arms are **96 short / 36 long**. The label the card shows (`bias: AI short · tree short · regime up`, `BiasLabelLine` `planner_prompt.go:271-279`, `BiasBlock.tsx:72` "a LABEL not a direction") is honest about what it is. What it is not is populated where the audit trail expects it: **`planner_read_facts.bias_ai` / `bias_tree` are empty on all 32 rows** while `bias_regime` reads `up/NORMAL` on all 32, and `tokens_in` is 0 on all 32 (`q03`). **[A]** Whoever built the facts table wired the regime word and not the two calls beside it.

### 3.2 Calibration, as the reports measured it (not re-derivable from this store)

- Weekly-structure bias (the shipped control): raw holdout hit **25–28%**, "significantly BELOW 50% … anti-predictive"; called weeks 45–51%, Wilson lo far below .50 (`2026-09-02-bias-calibration.md:97`, verdict table `:118`). The report's own D6 bar (`:55`): USABLE iff holdout Wilson lower bound > 0.50 AND net-of-friction t > 2.
- Live legs (`2026-09-02-live-bias-replay.md:118-122,159`): regime 0.5435 [0.4018, 0.6785] n=46; every leg NOT USABLE; plan-stamp legs 0.48–0.54 with lower bounds ≤ 0.40 at n = 21–62.
- The only signal with a real sign edge is Moskowitz–Ooi–Pedersen-style TSMOM at 252 sessions: holdout hit .564, Wilson lo .5265, binom p .0005 — **and net-of-friction t = −2.56** (`bias-calibration.md:75`). **[R]** A real edge at a horizon this book cannot hold.
- Branch 5 (premium/discount veto), which the prompt states as law: measured a coin on this tape, n=21, p=0.476 [0.283, 0.676] (`2026-09-05-veteran-part-c.md:279`), and 17 of 58 plans citing a side veto stamped the forbidden side anyway. **[I] dressed as machine law.**

### 3.3 Should any bias exist?

**Not as a direction. As context, yes.** My ruling, with the evidence bar attached:

1. **Keep the regime block as facts** (`kernel/regime.go` — trend D/1h, dATR percentile, gap; "never a card readout" is already its contract). Facts inform a read; they do not permit or forbid a side.
2. **Remove the two places bias is still mechanical:** the branch-5 veto text in the planner prompt (`planner_prompt.go:183-189`) — it is [I], measured a coin, and it is the one line that makes a rising day short-only at the arm layer — and the post-flip MANDATORY bias (`auto_trader_planner.go:1635`), which forces the next plan's direction from a flip rule nobody has calibrated. Prompt + code; owner ruling because both were rulings.
3. **Replace "bias" with symmetric arm eligibility plus a direction-mix validator** that does not exist (`grep -rn 'direction mix' kernel/` returns nothing; the prompt's own instruction at `:707` "do NOT default to 2 longs + 1 rally-rejection short" is unenforceable prose). Rule: `bias.direction ∈ {long, short}` ⇒ at least one *armable* scenario on that side, else reject at write (turn `BiasArmWarning` into a reject once its warn count is on the record — it is warn-first by the 09-04 ruling and shipped uncalled until #91). `neutral` ⇒ one armable scenario each side.
4. **Evidence bar for ever re-admitting a directional call:** the D6 rule already written down — holdout Wilson lower bound > 0.50 on ≥ 200 called sessions AND net-of-friction t > 2 at the horizon this book actually holds (median 26 min). No signal in the census clears it; TSMOM-252 clears (a) and fails (b). Until one does, a bias is a story, and the card is right to show it as a label. **[T]** [R]

**[I]** In thirty years the only intraday "bias" that paid on NQ was the one written by the first hour's order flow at a level, not the one written at 08:00. The Zarattini–Aziz opening-range result **[R]** is a rule about the first bar's direction with a hard stop, tested on years of data; it is not a daily bias, and this system does not implement it — do not let its name be borrowed for one.

---

## 4 · THE GATE CHAIN, LEG BY LEG

Order as the code runs it at dev (`trader/entry_gate.go:146-320`); the running rev `36648655` is identical minus Leg D. Refusal counts are **since 09-02** unless marked, from three sources laid side by side per the silent-refusal rule: `decision_records.execution_log/risk_check_error` (`q03`), the `arm_refusals_0b` counters (`q03`), the two-day audit's committed `refusals.csv` (`q12`), and the 09-04 log. Counterfactual dollars are the audit's own CSV rows (its horizons), re-summed here; 09-04 counterfactuals are mine from the 1m bars (`q13`, `q14`).

| leg | what it protects | refusals since 09-02 | counterfactual | my ruling |
|---|---|---|---|---|
| **D daily force-flat** (dev only) | money — the daily loss limit | 0 (never live) | — | **Ship it; on the running rev a resting arm fills through a tripped limit.** Order: after one-open. |
| **0 strict** (`:184-197`) | policy — closes the decision path | **13** (09-03 20:35:06→21:12:40, all ASIA S1 short, ids 37304–37322) | all 13 STOP: −$511 as 13 trades, **−$43.50 as the one position they really were** | Protects money by construction: the path it closes is −$750.43 on 51. Keep — **then stop paying the executor to be refused (§1c)**. |
| **1 plan bias, direction mode** (`:200-207`) | nothing under strict | 0 ever | — | Dormant leg. Delete or leave labelled dormant; a dormant leg is a future surprise. |
| **2 scenario–direction, class 48** (`:210-220`) | sanity — an entry must match the scenario it cites | 0 | — | Keep; free. |
| **3 invalidation, arm only** (`:229-252`) | "a well-formed trade into a dead setup" | **2**, both 09-04: LONDON S2 02:00:46; NY S1 09:02:09 (counters `…:entry_gate:invalidated=1` ×2) | **both reached TARGET**: LONDON v1 S2 long 29579.5 touched 02:00 → target 29639.5 at 02:30, **+60 pts**; NY S1 short 29657.38 touched 09:02 → target 29503.38 at 09:56, **+154 pts**; **+$428 forgone**, n=2, **no verdict** | Keep, but it is now the leg with the worst ledger per firing, and its premise is untested. **Data first:** record verdict + 30/60-min counterfactual per firing. |
| **4 shadow map** (`:257-259`) | policy — shadowed conditions never place | 0 at the gate (breakout_retest is written as `shadowed` rows upstream, `armed_executor.go:355-380`) | — | Engineering; keep. |
| **5 R:R at execution price** (`:261-283`) + the arm seam's R:R (`armGateVerdictFor`) | the reward floor | decision path **3** (09-02 19:00:23 · 20:53:52 · 09-03 02:12:56); arm seam **11 counter increments** (09-02: 5 · 09-03: 2 · 09-04: 4) + the 09-04 flap (5 "gate changed: rr" cancels of live broker orders on one scenario) | decision path: 2 STOP 1 TARGET, **−$76.50**; arm seam 09-02/03 (n=7): −$43.14 events / −$14 distinct (audit) | **Money: marginally positive. Mechanism: broken.** Every one of the audit's 7 and the 09-04 flap were authored ≥ 2.0 and pushed under by the stop composer. Fix in §2.3 (4). |
| **6 min-SL ×ATR5m** (`:286-294`) | the stop floor | EntryGate leg: **3** (09-02 18:48–18:52) — all three the two-ATR bug (450.56 = 1.5×dATR 300.4), fixed 09-03; `validateDecision` min-SL (`engine_position.go:229` → `min_sl.go:56`): **34** | leg 6 bug rows: 3 TARGET **+$317 forgone**; validateDecision 34: 25 STOP / 8 TARGET / 1 flat, **−$435.00 session-flat, −$499 CME-day**; biggest single saved loss **09-02 10:36:09 NY long 29196.5 / SL 29145.75 → STOP 10:49, −$101.50** (decision cycle 26500) | **Protects money [T] n=34.** The prior review's "stop floor is calibrated to the losers" reads the same on my rows: winners' MAE median 0.353R (max 0.967R, n=16), losers' median 1.048R (n=39) — the floor is where losers go, not where winners go. Keep at 1.5. |
| **7 one open position** (`:301-307`) | money — no adds, no flips | 0 | — | The most absolute and cheapest leg, running last. **Move first**, so a second entry while in a position is refused for the true reason instead of an R:R message (reason laundering). |
| **no-chase** (`:312-318`) | warn-only | 40 evaluations, all `dist=0.00×ATR run=NULL` (audit D36) | — | Cannot fire on the arm path (`citedLevelFor` returns the arm entry). Fix or delete; today it is theatre. |
| other refusers seen | `last_entry_cutoff` 1 (09-02 01:51:52, cf +$22.50 / −$61.50 CME) · `stale_reeval_refused` 2 (08-27, 09-01) · `session_gate first-5m` 1 (08-27) · `superseded_wait` 30 (not refusals) · marketable guard 5 (09-02/03) · stale reaper 3 (09-02/03) + **12 on 09-04** (ids 39, 73–85) · boot sweep 8 (09-04) | | | The 09-04 reaper deaths are the pre-fix reaper (`51916172` went live 13:25); the audit already ruled its mechanism wrong (D10); fixed. |

**A refusal that saved a loss, from the store, by id:** decision cycle 26500, 09-02 10:36:09 CT, NY, `open_long` 29196.5 / SL 29145.75 / TP 29298.25, refused by `validateDecision` min-SL (dist 50.75 < 1.5×ATR5m); the tape hit 29145.75 at 10:49 — **−$101.50 avoided** (`refusals.csv` row; `q12`). Position 590 (10:37:17, long 29193.25, −$99.00, MFE 1.0 pt) is the trade that got through two cycles later at a wider stop. **[A]**

**A refusal that cost a win, from the store, by id:** arm 39 (NY S1 short limit 29657.38 / SL 29720.5 / TP 29503.38): refused by leg 3 at **09:02:09** ("scenario S1 invalidated at 09:00 CT (accepted through)") — the tape was passing down through 29657.38 in that minute (09:0x bar 29691 → 29644); placed at **09:10:08** (price 29601–29644, level now 13–56 pts above); on the broker book 09:10:08→09:26:06 (`nt8_order_snapshots` working_count 1 with signal `89339298…`); cancelled **09:26:06** by the pre-fix stale-window reaper; price never returned to 29657 and reached 29503.38 at **09:56** (+154 pts = $308). Three mechanisms, one miss, n=1. **[A]** `q13`, `q14`, log.

**Count since 09-02, all legs, all sources** (the number the dispatch asked for): decision-path EntryGate 19 (strict 13 · rr 3 · min-SL-bug 3) · `validateDecision` min-SL 34 · arm-seam R:R 11 counter increments · invalidation 2 · last-entry cutoff 1 · marketable guard 5 (+1 on 09-04 = 6 all-time) · stale reaper 3 + 12 · boot sweep 8 · "gate changed: rr" cancels 5 (09-04 flap; the audit window had 1 min-SL cancel). **09-04 had zero decision-path refusals: 374 cycles, 370 `wait`, 4 empty, 0 open intents.** **[A]** `q03`, `q04`, `q12`.

**Which legs protect money:** min-SL (both copies), one-open, daily (once live), strict (by closing a losing path). **Engineering:** scenario-direction, shadow, no-chase. **Broken mechanism, positive ledger:** R:R — and the fix is upstream (judge at write on the composed stop), not a looser floor. **Untested premise:** invalidation, 0-for-2 on the tape.

**Order I would run:** one-open → daily → strict → scenario-direction → shadow → invalidation (instrumented) → R:R (judged once, at write) → min-SL → no-chase (or nothing). The current order refuses for the pricing reason before the structural one, which is how a ledger ends up saying "rr" when the truth was "you already have a position".

**Cross-check against AUDIT-CHECKLIST (78 classes):** the two-ATR refusals (class 53/no-trade rider), the reaper (D10 → `51916172`), the decision-path silent refusal (D32 → the 🚦 line in `entryGateDecisionTelemetry` at `:512`), the daily-limit hole (#91) are **already fixed**; I do not re-recommend them. The composer-vs-floor flap, the direction-mix validator, the gate reorder, leg-3 instrumentation, and the executor's dormant entry vocabulary are **not** in the checklist (greps: `direction mix` 0, `composeArmStop` 1 mention inside class 48's text, `flap` 1 unrelated, `approval` 0).

---

## 5 · WHERE HUMAN JUDGMENT GOES BACK IN

**Design, in one paragraph.** Before the open, never during it. At ReadCT+15 (08:15 CT for NY; 01:45 LONDON; 16:45 ASIA if ASIA ever trades again) the plan card shows one scenario with the machine's composed entry/stop/target and the reasons, and the owner has exactly three buttons: **Approve** (arms may place when their confirm is met), **Veto the session** (no arms, no re-plans, the veto is recorded as a counter and graded like a trade — what the vetoed plan would have done is written to the store the next morning), and **Re-read** (spends one of the four replans). **If no button is pressed by the open the session is NOT approved** — fail-closed, one WARN, and the card says so — because on 09-04 the alternative was a `no_trade` fail-close, a restart at 08:30:11 and an owner reset at 08:44 on NFP day. During the session the only human act is the **halt**: one button that flattens (`POST /api/risk/force-flat`, which exists — `api/server.go` route, `handleForceFlat`) and blocks new entries on both paths until 17:00 CT (that is `SetDailyForceFlat` from #91 with an operator caller, on dev only today; on the running rev the arm seam does not know the day is over). No "override the gate" button, no mid-session re-arm, no size dial above one contract. Session enablement stays a pre-week ruling, not a daily click: ASIA is −$552.43 on 16 with a 2-of-15 win rate, the code already ships it disabled (`session_registry.go:93`), the saved strategy `MNQ` re-enables all three (`knobs.csv` "sessions in force"), and that is the single owner ruling in this section that pays for itself on the existing record. Everything else — the numbers, the timing, the exits — stays with the machine, because the store shows that when the human-shaped component (the AI executor, which is the closest thing to discretion in the loop) touched the ticket, it lost $750 on 51 trades, and when it did not, the book was flat.

**[I]** The two controls I never regretted on an automated book were the pre-open "not today" and the halt. The one I always regretted was the mid-session "just this once".

---

## RECOMMENDATIONS, in order

1. **One scenario, machine numbers, R:R judged at write.** *What:* planner authors read + shortlist + invalidation + ONE scenario (+ one alternate on neutral days); entry = level; stop = `composeArmStop`; target = rule from the MFE distribution / next seated level; the validator computes R:R on the composed pair and refuses at write. *Why:* S1 +$321 vs S2–S4 −$787 [T n=58]; planned 2.36 vs realized 1.62 [T]; 35.6% of arms within 0.3 of the floor [T n=132]; 20-placement flap 09-04 [A]. *Takes:* prompt + code (`plan_doc.go` scenario cap → 1–2; `composeArmStop` called from the validator; a target composer) + owner ruling (0B ruling 2 said wider stops → more R:R refusals is "the intended trade"; this changes where the refusal happens, not whether). *Watch:* `gate changed: rr` cancels/day (09-04: 5 in 20 min) → 0; reject share 44.8% → < 15%; broker placements per arm (09-04: 20) → 1; S1 expectancy with CI at n=100.
2. **Stop the flap now, independently of (1)** — hysteresis on the working-arm re-check: cancel only when the composed R:R < floor − 0.10 on two consecutive cycles, and never re-arm the same (plan, version, scenario) more than twice per session without a plan change. *Why:* ids 62–102 [A]; D37 [A]. *Takes:* code (`armed_executor.go:440-448`) + owner ruling. *Watch:* rows per (plan, version, scenario) per session; `placement_seq` max (09-04: 20).
3. **Under strict, take the entry vocabulary away from the executor or drop its cadence to `watch`.** *What:* the executor prompt offers `open_long/open_short` (`engine_prompt_futures.go:237`) that leg 0 refuses; either remove them from the contract under strict, or run the executor only on `watch`/`post_exit` triggers (it already has those cycle types: 210 + 27). *Why:* 2,332 calls since 09-01, p90 51 s, 13/14 intents refused, path P&L −$750.43 [A]. *Takes:* prompt + knob. *Watch:* AI calls/day; `open_*` proposals under strict → 0; `close_*` and citation stamps unchanged.
4. **Direction-mix + arm-symmetry validator.** *What:* bias ≠ neutral ⇒ ≥ 1 armable scenario on the bias side, else reject; neutral ⇒ one each side; `BiasArmWarning` → reject after its warn count is recorded (class-35 law). *Why:* long armed 7.1% vs short 19.4% (latest) [T]; 21 reclaim stop-entry arms, all short [A]; no validator exists [A]. *Takes:* code (`plan_doc.go` + `arms_bias_coherent.go`), owner ruling (warn-first was ruled 09-04). *Watch:* armed share by direction → parity within the bias mix; BiasArmWarning count/day.
5. **Bias as label only — remove the branch-5 veto and the post-flip MANDATORY bias.** *Why:* branch 5 a coin at n=21 [T]; weekly control anti-predictive [T]; every live leg NOT USABLE by D6 [T]; `auto_trader_planner.go:1635` forces a direction from an uncalibrated flip [A]. *Takes:* prompt (`planner_prompt.go:183-189`) + code + owner ruling. *Watch:* the D6 rule on ≥ 200 called sessions before any directional call returns; meanwhile `planner_read_facts.bias_ai/bias_tree` must actually be written (data first — they are empty on 32/32).
6. **Reorder the gate and instrument leg 3.** *What:* one-open → daily → strict → … ; leg 3 records verdict + 30/60-min counterfactual per firing. *Why:* leg 3 is 0-for-2 on the tape (+$428 forgone) [T n=2, no verdict]; one-open last launders reasons [A]. *Takes:* code, small. *Watch:* invalidation firings with cf, n → 20 before a ruling.
7. **Pre-open approval card + halt button (§5).** *Takes:* code (three routes; `SetDailyForceFlat` exposed to an operator; card in `web/src/components/plan/`) + owner ruling + guide content in the same PR. *Watch:* sessions vetoed vs approved and their would-have P&L; time from read to approval.
8. **ASIA off.** *Why:* −$552.43 on 16, 2W/13L [T]; code default already disabled [A]. *Takes:* one owner ruling in Studio (`sessions_enabled`). *Watch:* nothing — re-open only at n ≥ 30 positive on LONDON+NY.
9. **A read that cannot land before the open lands a "pending approval" plan, not a fail-close.** *Why:* 09-04 NY [A]; 4 of 11 session-days since 09-01 got a scheduled-read version, 4 got a fail-close marker [A]. *Takes:* code (`auto_trader_planner.go` fail-closed path → lifecycle `pending`) + the card from (7). *Watch:* reads landed ≥ 10 min before open / session-days.
10. **Fix the crypto-shaped example in the executor prompt** (`position_size_usd: 60000, leverage: 1` at `engine_prompt_futures.go:237`) — cosmetic, read every two minutes. *Takes:* prompt. *Watch:* nothing.

**Where I disagree with existing reports, with the number:** (a) prior review §3.3/§1–3 — the loop is 2 min live, not 3 (boot line); (b) prior review "3 of 36" — 13 of 57 on the trade's own stop; (c) two-day audit "−$860.64 over 44 opportunities" — −$726.14 over 61 events from its own CSV, dedup key absent; (d) two-day audit "leg 3 refused nothing" — 2 firings on 09-04, both against winners; (e) prior review §3.5 "the gates are net-positive" — I agree on min-SL and strict and I add that leg 3 and the leg-6 bug rows are the negative part of the ledger (+$317 and +$428 forgone) and that the R:R leg's ledger is positive only because the composer keeps refusing trades the planner had already sized at the floor.

---

## SURPRISES — recorded, never acted on

1. **The 09-04 NY S2 flap:** 20 ledger rows (ids 62–102, 10:10:04→10:52:00), one scenario, one price (29591.02), 20 real stop-entry placements at NT8 (`nt8_order_snapshots` 10:00–11:00: 242 snapshots, 203 with a working order, max 9 working at once), 0 fills; deaths: `gate changed: rr` ×5, stale window ×9, `boot_sweep` ×8 (the 10:53:10 restart), `cancelled in NT8` ×1. The same scenario is labelled `sweep_reclaim` on id 38 and `reclaim` on ids 62+ within plan v3.
2. **The feed died at 12:19 CT on 09-04** (last 1m bar) and was still down at the 13:25:47 boot (`🚨 FEED DOWN: no NT8 bar for 1h6m47s while CME is OPEN` at 13:26:47) — the **second consecutive blind NY afternoon** after the 09-03 host reboot. Two arms (ids 104/105, S2 sweep_reclaim 29720.0, v6) have sat `armed` in the ledger since 12:11:02 with no feed; `/api/cutover-gate` leg 4 reads `broker 0 vs ledger 2 — MISMATCH; ledger[,]` (blank signal ids: never placed) and `ready:false`.
3. **`planner_read_facts` writes the regime word and nothing else:** `bias_ai`, `bias_tree` empty and `tokens_in` = 0 on all 32 rows; `version` = 0 on all 32.
4. **Reclaim-as-stop-entry was shipped 09-04 to rescue longs; every reclaim-family arm in the store is SHORT** (20 `stop_entry reclaim SHORT` + 1 `stop_entry sweep_reclaim SHORT`, 0 long).
5. **The 09-04 process restarted at 08:30:11 — the NY open to the second** — mid-way through the LONDON re-plan's third attempt (08:28:32 `attempt 3/3 reauthor+block`), on NFP day; the boot lines show a deploy, not a crash.
6. **`ab_confirm_log.net_pnl` is unusable on most rows:** `/api/expectancy` counterfactual_e8 blocks report `usable_n=0` for `touch`/ASIA (7), LONDON (5), NY `1x5m_close` (18) and "SHORT ROWS SUSPECT (E8 sign bug) — direction is not stored on ab_confirm_log"; in the table, usable net_pnl on 98 of 223 rows (`q03` rule×condition).
7. **Ghost duplicates:** `armed_entry` rows 576/577/579 (NULL `pnl_corrected`, `close_reason` `reconcile_flat`/`unresolved`) duplicate `reconcile` rows 575/578 — the same fills recorded twice under two sources; they are the 3 NULLs the corrected-column law excludes.
8. **Bars are backfilled across outages** (60/hour through the 09-03 12:24–14:18 reboot), so the bars table cannot testify to a blind window; `decision_records` gaps and logs can.
9. **The executor prompt's example is crypto-shaped** (`position_size_usd: 60000`, `leverage: 1`) on a one-contract MNQ book (`engine_prompt_futures.go:237`).
10. **`/api/config/resolved` lists `eod_flat_ct` as "ineffective — unreachable (auto_trader_clock.go)"** and `wake_min_interval_min` + five `wake_on_*` knobs as "no known reader" — the wake loop the plan churns on has knobs with no consumer (SYSTEM-MAP §11 also records "wake-predicate cutover: NOT FOUND").

---

## APPENDIX — every query and its output

All scripts and outputs are committed under `docs/superpowers/reports/2026-09-05-vet-03-decisions-data/` (35 files, 276 KB, secret-scanned). Scratch originals: `~/nofx-analysis/vet-03-0905/`.

| id | file | what it produced |
|---|---|---|
| q01 | `q01_store_premises.sh` → `.out` | row counts; era counts (71/67/71/64); `typeof(entry_time)=integer`; source/session/status/outcome distributions; max created_at per table |
| q02 | `q02_entry_time_dist.sql` → `.out` | entry_time by month and by CT date; plan_id 71; pnl_corrected 230 all-time |
| q03 | `q03_planner_arms_gates.sql` → `.out` | rejects by reason/session/attempt/date; plans by session×lifecycle, trigger_reason, per session-day; lifecycle log (7 rows); arms by state/reason, kind×condition×side, session, date; decision-path entry_gate refusals by day and reason; risk_check_error classes; intents by day (LIKE, superseded by q04); `planner_read_facts` bias columns; `system_config` counters; ab_confirm_log rule×condition |
| q04 | `q04_decisions_parse.py` → `.out` | actions parsed from decision_json by CT day since 08-27 (4,102 cycles); cycle_type×trigger; the 31 open intents with their refusal text |
| q05 | `q05_plan_doc_peek.py` → `.out` | plan doc schema (NY 09-04 v6) |
| q06 | `q06_plan_corpus.py` → `.out`, `q06_trigger_fired.csv` | corpus mixes (latest per session-day and all versions); arm-enabled by direction/condition; bias_label pairs; planned R:R; first trigger-fired pass (latest version) |
| q06b | `q06b_trigger_inforce.py` → `.out`, `q06b_trigger_inforce.csv`, `q06b_trigger_first.csv` | trigger-fired 263/592 in-force, 62/82 first-version; in-force window median 32 min |
| q07 | `q07_reauthor_cost.py` → `.out` | 28 reject clusters; re-author seconds n=26 median 286; fail-closed 6/26; families |
| q08/q09 | `q08_arms_0904_and_era_positions.sql` → `.out` | 09-04 arm rows; all filled arms; the 68 era rows with slot/source/session/close_reason/mae/mfe; rollups by slot, source, session, close_reason; era totals |
| q10 | `q10_churn_durations.sql` → `.out` | versions per session-day since 09-01 with trigger mix; inter-version minutes; executor ai_request_duration percentiles; cited-scenario distribution |
| q11 | `q11_wilson_pnl_2r.py` → `.out`, `q11_mfe_r_rows.csv` | every Wilson interval quoted; P&L split by path with ids; MFE≥2R 13/57 with per-row R |
| q12 | `q12_refusals_by_leg.py` → `.out` | per-leg counts and counterfactual dollars from the audit's `refusals.csv`; best saved loss / worst missed win per leg |
| q13 | `q13_coverage_leg3_broker_levels.py` → `.out` | bars per hour 09-03/09-04; leg-3 counterfactuals; broker snapshots during the flap; plan levels vs seated pool 54/105; arm R:R bins |
| q14 | `q14_s1_episode_closes_durations.py` → `.out` | 09-04 08:55–10:05 5-min tape; LONDON v1 S2 cf; AI close_* count; holding minutes |
| api | `api_expectancy.json`, `api_expectancy_by_session.json`, `api_risk_gate-blocks.json`, `api_cutover-gate.json` | as read at ~09:55 CT (config dump withheld) |

**Commands quoted in the text (not in a script):**

```
git -C /home/hoang/nofx worktree add --detach /home/hoang/nofx-vet-03 origin/dev      # 2a66d91c
NOFX_SESSION=vet-03-0905 deploy/nofx-claim.sh new docs/vet-03-0905 "…"                 # CLAIMED
git log --oneline 36648655..origin/dev                                                 # #90, #91 + docs
git diff --stat 36648655 origin/dev                                                    # 18 files
sed -n 11,1845p docs/superpowers/AUDIT-CHECKLIST.md | grep -cE '^[0-9]+\. \*\*'         # 78; comm → 46 missing
grep -rnE 'direction mix|long_count|short_count' kernel/plan_doc.go kernel/planner_*.go  # (empty)
grep -rnE 'Bias\.Direction' kernel/*.go trader/*.go | grep -v _test                    # the consumer table in §3.1
grep -hE '^09-04 13:2[5-6]' data/nofx_2026-09-04.log | grep -iE 'scan|strict|…'        # boot lines (ScanIntervalMinutes=2)
grep -hE '^09-04 10:(0[5-9]|[1-5][0-9])' data/nofx_2026-09-0[34].log | grep -iE 'arm|…'   # the flap
grep -hE '^09-04 0(7:[4-5]|8:[0-5])' data/nofx_2026-09-0[34].log | grep -E '📐|🧩|🗓|…'     # NY read timeline
grep -hE '^09-04' data/nofx_2026-09-0[34].log | grep -E '🚦 entry-gate REFUSED|⚔️ arm REFUSED|invalidated at'   # 09-04 arm refusals
sqlite3 "file:…?mode=ro" "SELECT datetime(emitted_at_ms/1000,'unixepoch','-5 hours'), working_count, (orders_json LIKE '%89339298%') FROM nt8_order_snapshots WHERE emitted_at_ms BETWEEN … "   # arm 39 on the book 09:10:08→09:26:06
git merge-base --is-ancestor 51916172 36648655                                         # reaper fix IS in the running rev
```

*End of section 03.*
