# VETERAN DEEP REVIEW — SECTION 01 · THE WAY IT TRADES

**Lane vet-01-0905 · branch `docs/vet-01-0905` · owner hoang · 2026-09-05 (Saturday, CME closed) · READ-ONLY**

I am the reviewer the dispatch describes: thirty years in index futures, NQ since it listed, a discretionary desk and then my own automated books, LLMs in the loop for the last three years. I ran this section against the LIVE store and the running API, not against committed CSVs. Every number below names the query that produced it (`q##_*` under `docs/superpowers/reports/2026-09-05-vet-01-way-it-trades-data/`, mirrored from `~/nofx-analysis/vet-01-0905/`). Engineering claims carry **[A]** (I ran it / read the line), **[B]** (inferred from strong evidence) or **[C]** (speculation). Market beliefs carry **[R]** (study named), **[T]** (this system's own tape, with n), or **[I]** (my experience, untested here). All times CT.

---

## EVIDENCE BASIS

**Tree.** Worktree `/home/hoang/nofx-vet-01`, cut detached from `origin/dev` tip **`2a66d91c`** ("Merge #91: wire RiskForceFlat and BiasArmWarning") at 09:54 CT; claim commit `e4f73d63` on `docs/vet-01-0905` (`deploy/nofx-claim.sh new` printed `CLAIMED docs/vet-01-0905 by vet-01-0905`).

**Resume note (11:10 CT).** The first attempt of this lane ran out of session budget at about 10:30 CT with the draft committed and pushed at `88d88fa5` (10:25 CT). I resumed at 11:10 CT on the same worktree (`deploy/nofx-claim.sh check docs/vet-01-0905` → `OK … claim … vet-01-0905`), re-read every query output the draft cites, re-verified every code citation at dev **and** at the running rev, re-ran the store as-of at 11:12 CT (`q31_asof_resume.out`: every table unchanged except `nt8_order_snapshots`, 4,616), and corrected what did not survive — each correction is marked in place and listed under PREMISES CORRECTED (P3, P20) or in the text (the S1 tally, the `entry_gate.go` and `session_registry.go` line numbers, the R6 anchor). `origin/dev` was still `2a66d91c` at resume (`git fetch origin dev`), so the base did not move.

**Running binary.** `GET /api/health` at 09:54 CT → `{"revision":"36648655cfe0","status":"ok"}` **[A]**. Boot line in `data/nofx_2026-09-04.log`: `09-04 13:25:47 🔐 BOOT INTEGRITY OK — rev 36648655cfe0` **[A]**. Dev is ahead of the running binary by #90 and #91; `git diff 36648655 2a66d91c -- kernel/ trader/` touches six files (`entry_gate.go` +26: the new **Leg D daily force-flat** on both paths; `auto_trader_planner.go` +14: `BiasArmWarning` wired warn-only; `engine_analysis.go`, `risk_limits.go`). When I say "live" I mean 36648655; when I quote a line I say which tree. Where it matters: on the running binary the daily-loss limit does NOT gate the arm path (dev #91 fixes it), and on the bound strategy `daily_loss_enabled=false` anyway, so the leg is inert on this configuration either way (`q24_config_values.out`, `q28_binding.out`) **[A]**.

**Specs and reports relied on** (`git log -1 --format="%h %ci %s" -- <path>` in the worktree):

| path | last commit |
|---|---|
| docs/superpowers/SYSTEM-MAP.md | a96224dd 2026-09-04 09:07:37 -0500 |
| docs/superpowers/AUDIT-CHECKLIST.md | 15340faa 2026-09-04 13:22:07 -0500 |
| docs/superpowers/research/INDEX.md | 4e8e7e1a 2026-09-03 19:37:14 -0500 |
| docs/superpowers/reports/2026-09-05-veteran-review.md (+ parts A–D) | 676f239c 2026-09-05 05:51:29 +0000 |
| docs/superpowers/reports/2026-09-04-two-day-audit.md | f3c640c3 2026-09-04 07:26:52 -0500 |
| docs/superpowers/reports/2026-09-04-research-conformance.md | 790efbb3 2026-09-04 09:20:28 -0500 |
| docs/superpowers/reports/2026-09-02-level-kind-replay.md | 3961f873 2026-09-02 19:03:10 -0500 |
| docs/superpowers/reports/2026-09-02-belief-census.md | ee64a494 2026-09-02 08:50:38 -0500 |
| docs/superpowers/reports/2026-09-03-mc-drawdown.md | 77e1cdfc 2026-09-03 00:39:25 -0500 |
| docs/superpowers/plans/2026-08-22-post-soak-build-plan.md | 62308978 2026-08-22 16:11:57 -0500 |

`docs/superpowers/plans/VL-MASTER-PLAN-v2.md` does not exist (`ls docs/superpowers/plans/ | grep -i master` is empty) — confirmed, not cited. `kernel/exit_*` does not exist; exit logic I read: `trader/exit_mechs_suspend.go`, `kernel/min_sl.go`, `trader/arm_stop_anchor.go`, `trader/auto_trader_clock.go:472-530`, `kernel/risk_limits.go`, `kernel/session_registry.go:83-126`.

**Store as-of** (`q00_asof.out`, read at 09:55:10 CT via `file:/home/hoang/nofx/data/data.db?mode=ro`): trader_positions 587 (last entry 2026-09-03 09:05:14 CT) · armed_orders 67 (last update 2026-09-04 13:27:48 CT) · plans 254 (last 2026-09-04 17:10 UTC) · touch_outcomes 677 · candidate_pool 360 · trade_excursions **0** · decision_records 37,768 · ab_confirm_log 223 · nt8_order_snapshots 4,461 (still ticking on Saturday — 4,468 by 09:58) · bars 79,769 · planner_read_facts 32 · plan_lifecycle_log 7 · level_stats 234 · touch_episodes 1,315.

**API read** (all HTTP 200 at ~09:57 CT, token from `cmd/gate-jwt`, never printed): `/api/health`, `/api/expectancy` (+`?by=session`, `?era=post`), `/api/config/resolved`, `/api/risk/gate-blocks`, `/api/cutover-gate`. Responses (minus config/resolved, 26 KB of paths) are in the data directory.

**Logs read** (read-only): `data/nofx_2026-09-02.log`, `-09-03.log`, `-09-04.log` — remembering the rotation trap (a file is named for the boot that opened it, not the date).

**Sample law.** Every trade claim below rests on **ids 521–591** filtered to `entry_time ≥ 1786770000000` (2026-08-15 00:00 CT), minus the test seam (572–574, `source='e7_farside_test'`) and minus the three rows with `pnl_corrected IS NULL` (576, 577, 579 — `reconcile_flat`/`unresolved`, $0 realized). That is **n = 65** usable rows; the excluded count is 6 and is shown wherever it matters. `pnl_corrected` only.

---

## SUMMARY — one page

**Verdict for this section.** As it stands the book is a **short-biased level-fade book** — 65 trades in 14 session-days, 4.6 a day, one MNQ contract — that wins **33.3 %** of decided trades (21/63, Wilson [22.9, 45.6]) and is paid **1.62 : 1** on the winners, for **−$8.68 a trade, 95 % CI [−33.08, +15.73]**, gross of commission (`q21_stats.out`). Break-even at a 1.62 payoff is a 38.2 % win rate; the book runs five points under it. Nothing in the tape proves an edge; nothing rules one out in exactly one cell — **reject-condition fades (n = 31: 14 W / 16 L / 1 flat, +$586, 46.7 % [30.2, 63.9])** — and everything the system does that is not that cell loses. The shape (resting limit at a seated level, stop = 1.5×ATR5m, target written by the model to clear a 2.0 R:R gate, flat 14:45) **cannot make money on this tape as configured** because the reward side is fiction: the planned target sits at **3.44 × ATR5m** (p50) while the tape's favourable excursion is **1.03 × ATR5m** (p50) and **2.47** at p80 — **8 of 59** trades ever travelled as far as their own target, **10 of 65** exits happened at the target and **37 of 65** at the stop. That is arithmetic, not luck. **[T] n = 65.**

**Three biggest problems.**

1. **The reward is written to the gate, not to the tape.** 47 of 132 armed scenarios since 08-15 carry a planned R:R inside [2.00, 2.30) — 35.6 % [28.0, 44.1] (`q11_planned_rr.out`) — and the prompt literally tells the model the floor and the stop floor (`kernel/planner_prompt.go:733` "R:R … must be ≥ 2.0 (ARM_MIN_RR) AND the stop distance must be ≥ 1.5× the current 5m ATR"; `kernel/class45_feeds_forward.go:199` "author stops AND targets consistent with it, or your R:R will not survive the widening") **[A]**. A gate on a number the model controls is an instruction to write a bigger number.
2. **The entry that produced 78 % of the book is the worst entry.** Decision-path market entries: 51 trades, 13 W / 36 L / 2 F, −$750.43, 26.5 % [16.2, 40.3]; the `1x5m_close` confirm class alone is 29 trades, 5 W / 23 L / 1 F, **−$1,135**, 17.9 % [7.9, 35.6]. Arm fills (touch at the level): 11 trades, 6 W / 5 L, +$94.50 (`q21_stats.out`). Strict mode closed that door on 09-03; the historical book is that door. **[T]**
3. **It trades every tape the same, and the dead tape pays the bill.** ASIA: 16 trades, 2 W / 13 L / 1 F, **−$552.43** (13.3 % [3.7, 37.9]); the low-ATR5m tercile (ATR ≤ 22.5 pts): 21 trades, 4 W / 17 L, **−$795.50**. Mid + high terciles together: 41 trades, +$36.07. The code registry ships ASIA and LONDON `Enabled: false` (`kernel/session_registry.go:93,102`; NY `Enabled: true` at `:114` — the draft's `:95,105` was off by two) and the bound strategy's `sessions[]` turns both back on (`q28_binding.out`) **[A]**. n < 30 everywhere, so "no verdict" by the rule — but the sign of every dead-tape cell is the same.

**Three biggest opportunities.**

1. **Take the target away from the model.** Set the first target from the tape — the nearer of the next seated opposing level and the cell's MFE p60 (for reject, ≈1.4 × ATR5m) — and use R:R as a *filter* against that target, not as a number for the model to satisfy. Replay it on the 65 trades first (1m bars exist for every one of them, `q07_bars_coverage.out`). What I would watch: target-hit share (now 10/65 = 15.4 % [8.6, 26.1]) and payoff.
2. **A no-progress scratch.** Losers die at the stop (38 of 41 with MAE ≥ 0.95 × stop [80.6, 97.5]) after a median 21 minutes; 21 of 41 never printed 0.5 × stop of favourable excursion (20 did; `q25_mfe_atr_targets.out`, `q30_details.out`). A "no 0.3 × stop progress inside 15 minutes → scratch" rule would cut the average loss without touching winners (winners' MAE p50 is 0.37 × stop). Replay before any knob.
3. **Gate the session and the vol regime by a measured switch, and fix the detector writer before ruling on any level kind.** The D1′ corpus in the store is 677 rows but only **370 distinct touches** (the writer re-records every touch on every plan-version read: `q14_touch_dup_mech.out`), so every per-kind number ever quoted from it — including the prior review's "RTH-L breaks 68 % (n = 63)" — is inflated 1.8×. Deduplicated, the only kind with n ≥ 30 is VWAP (56 hold / 37 break, 0.602 [0.501, 0.696]); RTH-L is 4 / 9, n = 13 — **no verdict**.

**The one thing I would stop everything for on Monday** is in SURPRISES, not in the recommendations, because I was told not to act on surprises: on 09-04 between 10:25 and 10:53 CT the NT8 SIM book held **nine identical resting sell-stop entries** for a one-contract system while the ledger called each of them cancelled (`q18_nt8_working.out`, `q22_0904_arms.out`). That was rev `70af663d`, not the running binary, and a boot swept it — but nothing in the checklist names it.

---

## PREMISES CORRECTED

Every statement in the dispatch or a prior report that I found wrong or unmeasurable, with the query that shows it.

| # | premise | what the store says | query |
|---|---|---|---|
| P1 | "227 rows have entry_time ≥ 2026-08-15; 223 pnl non-NULL; 64 cited; 66 of 227 mae/mfe" | **71 rows** (ids 521–591) have `entry_time ≥ 1786770000000` (2026-08-15 00:00 CT = the epoch the dispatch's own `entry_time` note requires). `created_at ≥` the same epoch also gives 71. Of the 71: pnl non-NULL **67**, plan_id **71**, non-empty cited_scenario_id **64** (matches), mae/mfe non-NULL **70**. The dispatch's 227 is not reproducible with a 2026-08-15 boundary; a 2025 epoch (my own first mistake, `q01_era_dist.out`) returns all 587. | `q01b_era_dist.out`, `q08_misc.out` |
| P2 | "touch_outcomes: 677 rows … D1′ records every touch" | 677 rows = **370 distinct (level_kind, opened_at_ms) touches** from **3 plans / 12 plan-versions** (09-03 ASIA v5–6, 09-04 LONDON v0–3, 09-04 NY v0–5), opened 2026-09-02 22:10 → 09-04 11:49 CT. The same RTH-L touch at 29199.25 opened 09-02 22:10 is stored under plan_version 2, 0 and 3, created 12:45:01, 13:00:38, 13:17:32 on 09-04, identical outcome — the writer re-records the whole scope window on every read. Only 6 of 370 groups carry conflicting outcomes; 25 carry >1 price (VWAP snapshots per version). | `q05_touch_dup.out`, `q14_touch_dup_mech.out`, `q14b_touch_conflicts.out`, `q27_touch_dedup_majority.py.out` |
| P3 | "nt8_order_snapshots.orders_json is the broker's book" (for the era) | The table **begins 2026-09-03 21:48:22 CT** (= 09-04 02:48:22 UTC; the draft's first pass printed the UTC value with a CT label — corrected on resume, `q33_snapshot_epoch.out`). The last era trade (591) exited 09-03 09:20 CT, so the conclusion stands. It covers none of the 65 era trades; my per-trade NT8-stop lookup matched 0 rows. Usable only for 09-04. | `q13_enrich.out` (line 1), `q08_misc.out` |
| P4 | "ab_confirm_log (223 rows: mfe, mae, mfe_r …)" as an excursion substitute | Columns exist but: `is_counterfactual = 0` on all 223 (the API labels every row counterfactual regardless); `outcome = 'open'` on 163; `net_pnl` of −116,323 … −117,664 on the 24 `recompute='unrecomputable:fill-bar'` rows; `/api/expectancy` reports usable_n ≤ 18 per cell. Not an excursion source. | `q06_abconfirm.out`, `api/expectancy.json` |
| P5 | "RTH-L breaks 68 % of the time (20/63)" (veteran-review.md:84, part-a:73) | The 20/63 is D5b's snapshot of the SAME duplicated table (`D5b-touch_outcomes-by-kind.csv`: RTH-L rows_all 68, hold 20, brk 43). Deduplicated by (kind, opened_at_ms) with majority outcome: RTH-L **4 hold / 9 break, n = 13, 0.308 [0.127, 0.576]** — no verdict. | `q27_touch_dedup_majority.py.out` |
| P6 | "plans claim a median 2.55:1" (part-a:109) | Over all 253 plans since 08-15 (738 scenarios, 132 with a full arm): planned R:R median **2.36** (p25 2.10, p75 2.93; min 0.56, max 5.55). The 2.55 is the 09-01/09-02 window (n = 59 armed scenarios here: median 2.54) — a two-day slice, not the book. | `q11_planned_rr.out` |
| P7 | "book realises 1.66:1" | With id 591 included (n = 65): avgW $114.67 / avgL −$70.76 = **1.62**. The 1.66 is the n = 64 sample. | `q21_stats.out` |
| P8 | "only 3 of 36 reached 2R off the minimum stop" | On the live store with ATR5m recomputed from the 1m bars (Wilder 14 on 5m aggregates): **5 of 59** reached MFE ≥ 2 × 1.5 × ATR5m (ids 524, 529, 555, 557, 581; 8.5 % [3.7, 18.4]); **14 of 62** reached 2 × their ACTUAL initial stop (22.6 % [14.0, 34.4]); 22 of 59 reached 1 × the minimum stop. | `q21_stats.out` |
| P9 | "winners' MAE never exceeds 0.661× the floor (n = 10)" (part-a:135) | n = 17 winners with ATR: **16 of 17 below 0.67×**, but **id 529** (reject, +$168) went 43.75 pts against a 36.0-pt floor = **1.22×** — the 1.5×ATR floor would have stopped a winner. The shape of the finding survives; the "never" does not. | `q21_stats.out`, `q30_details.out` |
| P10 | "the median winner hands back about 24 % of its best price" (part-a:190) | Paired giveback (MFE − realized, per winner, n = 20): **median 4.0 pts, p75 8.5**. The 24 % came from comparing the median of one distribution to the median of another. Winners exit at their target, near MFE. | `q21_stats.out` |
| P11 | "5 of 15 arms died to the marketable guard" (part-b, part-d:752) | Confirmed for the 09-01 → 09-03 window (arms.csv ids 23–37: 25, 27, 33, 34, 36). All-time non-test: **6 of 61** (id 17 added); fills **9 of 61** (14.8 % [8.0, 25.7]). | `q02b_arms.csv`, `q02b` tail |
| P12 | "113 min silent" (two-day-audit.md:35) | Verified in `data/nofx_2026-09-03.log`: last line **09-03 12:23:33**, next **14:18:24** (the audit's 12:24:33 is from `log_events`). Outside this section's scope beyond verification. | shell in APPENDIX A12 |
| P13 | "~1,810 trades needed" (mc-drawdown.md:173, n = 64) | At n = 65: mean −8.68, sd ≈ 100.4, se 12.45 → n ≈ **1,050** (population sd) / ≈ 1,070 (sample sd; part-b:42 says 1,067). The number moves 40 % with one trade because the effect is ~0.09 sd. | `q21_stats.out` |
| P14 | "AUDIT-CHECKLIST.md has 79 classes" | Highest numbered PART-1 class is **79** ("Silence read as death", line 1820); a `## CLASS 75` header at line 2002 duplicates the numbering scheme. 79 is right; count them yourself as the dispatch said — I did. | shell in APPENDIX A14 |
| P15 | "kernel/min_sl.go:~62 prints the raw ATR where the threshold belongs" | Line 64: `"sl_too_tight: %.1f < %.1f×ATR (%.1f)"` prints dist, mult, raw ATR — the threshold mult×ATR is never printed. Cosmetic. Also line 43's comment says "default 1.0" while the const at :34 is 1.5 (stale comment). **[A]** | `kernel/min_sl.go:34,43,64` (dev = running) |
| P16 | "exits: stop = max(anchor beyond nearest seated level, 1.5×ATR5m)" | The composition exists (`trader/arm_stop_anchor.go:composeArmStop`) **[A]**, but `system_config.arm_stop_unanchored_0b = 220` — the anchor leg found no seated level within 3×ATR5m on the risk side 220 times, and there is no counter for the anchored case. In practice the ATR floor governs **[B]**. | `q28_binding.out` |
| P17 | "2-minute AI executor loop" | `traders.scan_interval_minutes = 2` for the bound trader; SYSTEM-MAP §12 says "default 3, min 3" (`store/trader.go:28-29`). The store holds 2 and the log cadence is 2 min (09:02, 09:04, …). Not a correction of the dispatch — a drift between map and store. | `q28_binding.out`, log grep |
| P18 | "ASIA 16:30 / LONDON 01:30 / NY 08:00 CT reads … sessions disabled in code" | Registry: ASIA/LONDON `Enabled: false`, NY true (`kernel/session_registry.go:83-126`). Bound strategy `a5b7662e` ("MNQ", `is_active = 0` in the strategies table but bound by `traders.strategy_id`): `sessions_enabled = ["NY"]` AND `sessions[] = [NY max_trades 10, ASIA enable:true max_trades 7, LONDON enable:true max_trades 10]`. Plans, arms and trades exist in ASIA and LONDON through 09-04. The per-session list wins over the top-level list **[B]** (I did not trace the resolver; the plans are the proof). | `q28_binding.out`, `q03_plans_recent.out` |
| P19 | "RTH-L carries the highest average score in the candidate pool (1.60) and seats 5 of 5" (part-a:76) | Live pool (360 rows, 3 plans): RTH-L avg 1.45, seated 11 of 11; **PDL 1.92** is the highest seated score; DEMAND seats 22 of 59 at 1.70. Different snapshot, same story: the pool seats the multi-day kinds on sight. | `q23_candidate_pool.out` |
| P20 | draft of this report (88d88fa5): "`entry_gate.go:160-172`" (strict), "`session_registry.go:95,105`", "`trader/close_sync.go:156-196`", "snapshots begin 09-04 02:48:22 CT", "closed by rr (5) and stale-window (12)" | strict is `:160` at the running rev and `:184` on dev (#91 inserted Leg D above it); `Enabled: false` sits at `:93` and `:102`; `close_sync.go` does not exist — the writer is `store/position.go:418/:427` called from `trader/ninjatrader/reconcile.go:259,289,460` (and `:524/:557` for partials); the snapshot table begins 09-03 **21:48:22 CT**; the S2 rows closed 5 / 7 / 8 / 1 by reason. Each is corrected in place. | resume verification, APPENDIX one-liners |

---

## Q1 — THE BOOK, AS I WOULD BRIEF A RISK MANAGER

**What it is.** One MNQ contract (`entry_quantity = 1` on 70 of 71 era rows; the one `2` is test-seam row 574), SIM account Sim101, one trader. Since 2026-08-19 03:22 CT it has closed **65 usable trades** in 14 session-days — **4.64 a day** — the last at 09-03 09:05. Sixty-five percent of them are shorts (42 vs 23). Median hold **26 minutes** (p25 12.2, p75 55.4, max 220); winners hold a median 38 minutes, losers 21. It has never held through a session flat. Gross of commission: `fee = 0` on every row (`q23_candidate_pool.out` tail) — MNQ round trips at Tradovate-class retail rates are roughly $2–3 **[I]**, so add ≈ −$2.5 to every mean below.

**What it fades, what it follows.** By the condition of the scenario each trade cited (`q13_trades_enriched.csv` ⋈ `plans.doc`):

| condition | class | n | W / L / F | Σ P&L | win % [Wilson] | mean R |
|---|---|---|---|---|---|---|
| reject | fade at a level | 31 | 14 / 16 / 1 | **+586.00** | 46.7 [30.2, 63.9] | +0.46 |
| breakout_retest | follow | 9 | 1 / 8 / 0 | −581.50 | 11.1 [2.0, 43.5] | −0.76 |
| reclaim | follow (stop-entry through the level) | 5 | 0 / 5 / 0 | −436.50 | 0.0 [0, 43.4] | −1.00 |
| sweep_reclaim | fade-then-follow | 6 | 1 / 5 / 0 | −207.00 | 16.7 [3.0, 56.4] | −0.54 |
| acceptance | follow | 6 | 1 / 4 / 1 | +4.57 | 20.0 [3.6, 62.4] | −0.20 |
| hold | fade | 1 | 1 / 0 / 0 | +168.00 | — | +2.69 |
| (off-plan, no scenario) | — | 7 | 3 / 4 / 0 | −97.50 | 42.9 [15.8, 75.0] | −0.71 |

Only reject has n ≥ 30. Every follow-class cell (breakout_retest + reclaim + acceptance = 20 trades, 2 W / 17 L / 1 F, −$1,013) is negative; n = 20, no verdict, same sign everywhere. The planner authors far more than it trades: 253 plans since 08-15, 738 scenarios — reject 234, sweep_reclaim 186, breakout_retest 124, hold 72, reclaim 63, acceptance 56, breakdown_continue 3 — and arms **132**, of which **96 are reject** and 96 are short (`q11_planned_rr.out`). So the arm layer is a short-fade layer by construction; the decision path, while it was open, traded whatever the executor fancied.

**When.** By plan session: ASIA 16 (2/13/1) **−$552.43**; LONDON 21 (7/14/0) +$24.00; NY 21 (9/11/1) +$62.00; off-plan 7 −$97.50. By entry hour the losses stack 17:00–01:00 CT (18 trades, 2/15/1, −$626) and 06:00–08:00 (10 trades, 1/9, −$477); the only clean cell is 11:00–13:00 CT, **5 of 5 winners, +$649** (ids 524, 549, 566, 570, 578) — n = 5, an anecdote, but it is the opposite of the prompt's "NY AM 08:30–11:00 is the primary window" (`planner_prompt.go:653`), where the tape shows 17 trades, 6/10/1, −$398. **[T]**

**With what stop.** I recovered the initial stop for 62 of 65 trades: 51 from the `decision_json.stop_loss` of the cycle that opened the position, 7 from `armed_orders.stop_px` (the COMPOSED stop — e.g. arm 35's 29351.63, 66.6 pts, not the plan's authored 55.0), 4 from the plan's `arm.stop` where no ledger row matched (`q13_enrich.py`; stop_src column). No source for 566, 571, 580 (reconcile rows, off-plan). Median stop **33.75 pts** (p25 25.1, p75 44.1; min 3.75 — id 552 — max 92.75). In ATR5m units (Wilder-14 on 5m bars aggregated from the 1m table; n = 59): median **1.30×** (p25 1.01, p75 1.77); 14 stops below 1.0×, 22 in [1.0, 1.5), 13 in [1.5, 2.0), 10 at ≥ 2. Most of the book predates the 09-02 07:49 floor of 1.5× (`kernel/min_sl.go:34`, boot `4175e0b6`); post-0B there are **three** trades, all losses (589 −155, 590 −99, 591 −140).

**With what target.** Planned R:R at the fill (target and stop from the same source as above, n = 62): median **2.70** (p25 2.23, p75 3.41); 8 below 2.0 (pre-floor era). In distance: the planned target sits **3.44 × ATR5m** from entry at p50 (p25 2.72, p75 4.44) — `q25_mfe_atr_targets.out`.

**How it exits.** `close_reason = 'sync'` on all 65 usable rows — the store does not record whether the stop, the target, the EOD flat, an AI close or an invalidation ended the trade. Classifying by price (|exit − planned target| ≤ 2 pts, |exit − initial stop| ≤ 2 pts): **10 at target** (521, 523, 524, 529, 532, 536, 549, 555, 582, 584 — 15.4 % [8.6, 26.1]), **37 at stop** (56.9 % [44.8, 68.2]), **18 other** with no `close_*` decision within four minutes of the exit (`q29_neither_exits.out`) — EOD flats, reconcile materialisations, broker OCO partials; the store cannot say which. `exit_order_id` does not help: an NT8 order UUID on 60 of the 65, the literal `Close` on 4 (526, 530, 538, 575), `netting_fill` on 1 (578) — `q32_exit_order_id.out`.

**The arithmetic.**

```
n 65 · W 21 · L 42 · F 2 · Σ −563.93 · mean −8.68 [−33.08, +15.73] · sd ≈ 100.4
win rate (flats excl.) 21/63 = 0.333 [0.229, 0.456]
avg win +114.67 · avg loss −70.76 · payoff 1.62 · break-even win rate 0.382
by side: short 42 (16/24/2) +310.07 · long 23 (5/18/0) −874.00
```

The longs are the whole loss and then some; the shorts are marginally positive. Whether that is a directional edge or the fact that this fortnight's NQ spent its time rejecting rallies, I cannot tell from 23 longs; the plans' bias labels are absent on 64 of 65 rows (P16/`q21_stats.out`), so I cannot even condition on what the machine thought the regime was.

**Reconciliation with the API.** `/api/expectancy?by=condition` (as-of 2026-09-03 14:20 UTC) reports the same cells I do — reject n 31, 14/16/1, +586, mean 18.9, status FAILS (promotion needs the 95 % interval on the mean to exclude 0); breakout_retest 9, 1/8, −581.5; reclaim 5, 0/5, −436.5; acceptance 6; sweep_reclaim 6; hold 1 — with `excluded: unresolved_pnl 3, unresolvable 7, test_seam 3` (`q26_api_rows.out`). Its win rates use n (flats in) where mine use decided trades; the difference is one trade in two cells.

---

## Q2 — WHERE IS THE EDGE, IF ANYWHERE

The rule I applied: Wilson 95 %, n given, **no verdict below n = 30**. Under that rule the honest one-line answer is: the store cannot yet locate an edge in any cell, and it can rule one out in none — but the signs are not random.

### (a) Fades — touch_outcomes hold by kind × ordinal

**The corpus is not what it says it is** (P2). Deduplicated by (level_kind, opened_at_ms) with a majority vote over versions (`q27_touch_dedup_majority.py.out`): 370 touches, 258 decided, 112 ambiguous (30 %).

| kind | hold | break | amb | n_dec | p(hold) [Wilson] |
|---|---|---|---|---|---|
| VWAP | 56 | 37 | 53 | 93 | **0.602 [0.501, 0.696]** |
| OR-H | 9 | 9 | 4 | 18 | 0.500 — no verdict |
| DEMAND | 9 | 8 | 1 | 17 | 0.529 — no verdict |
| SWG-H | 9 | 6 | 14 | 15 | 0.600 — no verdict |
| SUPPLY / POC | 7 / 9 | 7 / 5 | 8 / 6 | 14 / 14 | 0.500 / 0.643 — no verdict |
| RTH-L | 4 | 9 | 1 | 13 | 0.308 [0.127, 0.576] — no verdict |
| every other kind | ≤ 7 | | | ≤ 11 | no verdict |
| **POOLED** | 145 | 113 | 112 | 258 | **0.562 [0.501, 0.621]** |

D1′ is calibrated at p(hold) = 0.5067 on IID-shuffled tape (`kernel/detector_d1prime.go:23-26`) **[A]**, so the pooled 0.562 is a whisker above chance over a day and a half of tape. Ordinal from this table is meaningless — every row of the same touch carries `ordinal = 1` per version — so I recomputed the touch sequence per kind across the corpus: 1st 7/19, 2nd 8/16, 3rd+ 130/223; the corpus is almost entirely "3rd+" because VWAP is touched all day. No verdict on ordinal from D1′ — and I must record that the little it says points the OTHER way: on the recomputed sequence, 1st touches held 7 of 19 (0.368 [0.191, 0.590]) against 3rd+ at 130 of 223 (0.583 [0.517, 0.646]) (`q27_touch_dedup_majority.py.out`, last block). n = 19; the two instruments disagree on direction; the belief I state below therefore carries **[I] [R]** and a split tape, not **[T]**.

**What the retired instrument says, for direction only.** `touch_episodes` (1,315 episodes, 08-25 → 09-03, 8 session days; `q15_touch_episodes.out`) is the instrument the detector redesign retired because it "called a touch a rejection when the close was still on its starting side — ≈0.69 on IID noise BY CONSTRUCTION" (`detector_d1prime.go:17-19`). Read against that null: ONH 88/97 = 0.907 [0.833, 0.950] and ONL 70/74 = 0.946 [0.869, 0.979] are the only families clearly above 0.69 (n ≥ 30), and the ordinal decline is monotone — 1st 451/567 = 0.795, 2nd 200/257 = 0.778, 3rd+ 325/491 = 0.662 — the same direction as the 1h replay's H8 ("1st 0.688 > 2nd/3rd+ 0.571", `2026-09-02-level-kind-replay.md`, commit 3961f873). I would not build on either level; I would build on the direction: **first touches hold more than later ones** — consistent with Osler's finding that clustered stop/limit orders are consumed on the first test **[R] [T]**.

**Realized fades.** The reject cell (above) is the only thing in this book with a positive sum: 31 trades, 46.7 % [30.2, 63.9], mean R +0.46. By session: reject × NY 12 (7/5) +$281.50; reject × LONDON 8 (5/3) +$538.50; reject × ASIA 11 (2/8/1) −$234.00. The prompt still tells the planner "reject-based setups are best in NY RTH (75 % win, +665 this week)" at `kernel/planner_prompt.go:737` on both the running rev and dev **[A]** — a one-week crown (checklist class 16) that the tape has since revised to 58 % (n = 12).

### (b) Breaks — the same inverted

RTH-L deduplicated: 9 of 13 broke — no verdict, and it is the *only* kind that leans that way at n > 10. The realized follow-class cells are all negative (breakout_retest 1/8, reclaim 0/5, acceptance 1/4/1; −$1,013 on 20). The one mechanically interesting fact: a `reclaim` short is executed as a **sell stop two ticks below the level** ("📌 armed S2 → WORKING stop-entry 29590.52 … offset 2t", `q22_0904_arms.out`; `planner_prompt.go:732`) — a momentum entry through the level, i.e. the system's break trade — and it is 0 for 5 on the decision path. n = 5. No verdict; wrong sign.

### (c) Time of day

Sessions: ASIA 16 −$552 (2/13/1), LONDON 21 +$24, NY 21 +$62 — none reaches n = 30. Hours (`q21_stats.out`, "by hour bucket"): 17:00–01:00 CT 18 trades −$626; 02:00–08:00 22 trades −$255; 08:00–11:00 17 trades −$398; **11:00–13:00 5 trades +$649 (5/5)**; 13:00–15:00 3 trades +$66.50. **No verdict** anywhere; the *only* positive buckets are the two the prompt does not privilege. Intraday-momentum research (Gao, Han, Li, Zhou 2018 on SPY) finds the first half-hour predicts the last half-hour, which argues for trading the open's direction late, not fading the open early **[R]** — this book fades early and is flat by 14:45, the exact window the research says carries the day's momentum.

### (d) Regime

Machine regime label: absent on 64 of 65 rows — `plans.doc.bias_label` starts 09-03 and `planner_read_facts.bias_regime` is `up/NORMAL` on all 32 rows (09-03/09-04). **No verdict on the machine label**, and no variance in it either. `day_type` from the plan doc (n: balance 23 −$736; trend 17 +$321; trend-down 13 −$537; trend_down 4 +$213 — two spellings of one label, see SURPRISES): no verdict.

Realized-vol tercile (ATR5m at entry, cuts 22.46 / 27.93 pts; `q21_stats.out`): low 21 trades, 4/17, **−$795.50** (19.0 % [7.7, 40.0]); mid 20, 7/12/1, +$158.57; high 21, 8/12/1, −$122.50. n = 21 per cell, no verdict — but the low-vol cell holds 141 % of the book's net loss, and it overlaps ASIA almost completely (ASIA's ATR5m in `planner_read_facts` runs 9–21 pts). Lo's adaptive-markets framing says an edge is regime-conditional and decays as it is arbitraged **[R]**; the practical version for this desk is simpler: a 1.5×ATR stop on a 13-point ATR is a 20-point stop on a tape that ticks 2–3 points at a time through a level — the fade is being chopped by the noise band, not defeated by a trend **[I]**.

**Bottom line for Q2.** One cell with the right sign at n = 31 (reject fades, mean $18.90, CI on the mean [−17.65, +55.46] — includes zero). Everything else is n < 25 with the wrong sign. The level table cannot rank kinds until the writer is fixed and four weeks accrue.

---

## Q3 — IS THE SHAPE ONE THAT CAN MAKE MONEY ON NQ

The shape, stated plainly: rest a limit at a level the planner seated; when it fills, hold one contract against a stop of max(anchor, 1.5×ATR5m) and a target the model wrote to clear 2.0 R; no scaling, no trailing, no breakeven (both suspended by the `EXIT_MECHS_SUSPENDED` default, `trader/exit_mechs_suspend.go:35-43`; the key is absent from `.env`; the running binary's boot line reads `09-04 13:25:47 🛑 exits: stop=max(anchor+clr, 1.5×ATR5m) · anchor_max=3.0×ATR5m · BE=off · trail=off · size=1 · re-arm-after-sweep=on (0B)` **[A]** — while the bound strategy row says `breakeven_enabled: true, trailing_enabled: true` (`q28_binding.out`; see S12)); flat at 14:45; the AI loop watches every two minutes but under strict mode may not enter (`trader/entry_gate.go:160-172` at the running rev `36648655`; `:184-196` on dev, where #91 inserted Leg D — the daily force-flat — above it at `:151-173`). Strict is not a paper rule on the running binary: it refused 13 decision-path market entries on 09-04 alone (`risk_check_error` = `entry_gate: refused: strict — plan_mode=strict executes plan scenarios …`, `q16_decision_census.out`) **[A]**.

**What the shape gets right [I] [R].** Resting limits at levels the market has already respected is how a fade should be entered — you are paid the spread rather than paying it, and Osler's clustering result says the orders you are leaning on are really there. One contract and a hard EOD flat are the correct posture for an unproven system. Suspending BE/trail was right; Round-7's ranking of ATR trails at the bottom of 15 exit families matches my own experience that a trail converts winners into scratches faster than it saves losers **[R] [I]**.

**What the shape gets wrong — and I mean the shape, not the numbers.**

1. **The target is a variable the model sets; the gate that judges it is the model's own input.** Planned target 3.44 × ATR5m at p50 against an MFE distribution whose p80 is 2.47 × ATR. Reject fades — the good cell — have MFE p50 1.40 × ATR, p80 2.66, p95 3.48 (n = 29). A target at 3× ATR is, on this tape, the 80th–90th percentile of what the trade will ever show you. The R:R gate then rejects any *honest* target and accepts any *inflated* one. Change the shape: the target comes from the map (next seated opposing level) or from the excursion table (cell p60), and R:R becomes a filter that says "no trade" when the honest target is too close. The prior review said this; I agree and I have the number the fix must move: 10 of 65 at target.

2. **The stop is anchored to volatility, not to the thing being faded.** A fade's structural invalidation is "the level is gone" — price accepts through it by more than the sweep the plan allowed for. The composition tries to anchor to the *next* seated level on the risk side (`arm_stop_anchor.go`), which is the wrong level for a fade: the nearest structure is the one you are fading. And in practice it does not bind: `arm_stop_unanchored_0b = 220`. So the live stop is 1.5 × ATR5m, wherever that lands. On the tape: winners' MAE p50 0.37 × stop, p80 0.71, max 0.97; losers 38/41 die at the stop. The trade is decided in the first bars; the stop is doing the work of a time-stop and doing it expensively (median loser −$70 = 35 pts). The shape change: stop = level ± (sweep allowance measured from the level's own touch history, e.g. the p80 penetration of prior touches — `touch_episodes.penetration_pts` exists) with the ATR floor as a *maximum*, not a minimum — and a scratch rule for "no progress".

3. **One contract forbids the only exit structure that pays a 1.6-payoff fade book.** With one lot you cannot take half at 1R and let the rest run to the map; you must pick a single target, and a single target at the p80 of MFE is a coin you lose four times in five. Either the owner rules a two-lot Stage A (`kernel/risk_limits.go:StageAContractCapDefault` — an owner ruling, and I would not make it before the reward side is fixed) or the target comes down to the p50–p60 of the cell. There is no third option that keeps a 2R fixed target on one lot.

4. **The confirmation-close market entry was a different shape wearing the same clothes.** `1x5m_close` entries pay the confirmation premium (you buy after the close that proves the level held, i.e. after the move you wanted has started) and then sit with a stop that is measured from a worse price. 5 W / 23 L / 1 F. Strict mode already removed it; the shape must not let it back in. The lesson for the arm path is the same: `wait_confirm` arms that chain on a `1x5m_close` (`armed_executor.go:410-426`) re-introduce the premium at the arm layer.

5. **What NT8 SIM hides [I].** Every arm fill in the book filled at the limit price exactly (575/576 at 29437.00, 578 at 29413.00, 584 at 29138.00, 591 at 29285.00; `q02b_arms.csv`, `q09_era_trades.csv`) — SIM fills a resting limit on touch with zero queue. Live, a limit at a level fills only when price trades through it or the queue ahead of you is exhausted, so the fills you *get* are adversely selected toward the touches that break. A fade book's SIM win rate is an upper bound on its live win rate, and the gap is largest exactly at the ONH/ONL/PDH/PDL kinds where the crowd's orders sit. The evidence I would demand before believing any SIM fade statistic: for each filled arm, the 1-minute path after the touch for fills vs. for touches that came within one tick and did not fill.

**Can the shape make money on NQ?** Yes, in one form: *fade a first touch of a seated level with a resting limit, stop beyond the level's measured sweep, target the next level or the cell's p60, scratch on no progress, NY only, size 1.* That is an ordinary professional level-fade and it makes money in the hands of people who do the two things this system does not: take the reward from the map and skip the dead tape. In its current form — model-written targets, volatility stops, every session — no. **[I]** with the numbers above as **[T]**.

---

## Q4 — THE REWARD SIDE

**The three claims, re-measured.**

- "Plans claim median 2.55:1" → **2.36** over all 132 armed scenarios since 08-15 (P6); 2.54 in the 09-01/02 window; **2.70** at the fill for the 62 trades with a recoverable stop and target. And 47 of 132 armed R:Rs sit in [2.00, 2.30) — 35.6 % [28.0, 44.1] — with 15 below 2.0 and only 26 at or above 3.0. That is a model solving an inequality: the prompt states the inequality (`planner_prompt.go:733`) and the floor line tells it how the executor will widen its stop (`class45_feeds_forward.go:199`).
- "Book realises 1.66:1" → **1.62** (n = 65).
- "3 of 36 reached 2R off the minimum stop" → **5 of 59** [3.7, 18.4] off the minimum (1.5×ATR5m), **14 of 62** [14.0, 34.4] off the actual initial stop (P8).

**Where the target should come from.** Four candidates, measured against this tape:

1. *The model.* Planned target 3.44 × ATR5m p50; **8 of 59** trades ever reached their planned distance (13.6 % [7.0, 24.5]). Reject cell: planned 3.01 × ATR vs MFE p50 1.40. Rejected — the model's target is a gate artefact.
2. *The level map.* Untested here — no query in this section computes "distance to the next opposing seated level at entry" — but structurally correct: it is a number the model does not control, it moves with the day, and a level fade's natural target is the opposite edge of the range it lives in. This is the *data first* item in my recommendations.
3. *The MFE distribution.* `trade_excursions` has 0 rows (verified: `q00_asof.out`); `trader_positions.mae/mfe` covers 70 of 71 era rows and I spot-checked three against the 1m bars (532: stored 156.75/61.5 vs recomputed 156.75/61.50; 589: 10.25/80.5 vs 10.25/81.25; 591: 43.5/75.0 vs 43.50/75.00 — `q21_stats.out`). By condition (pts, p50/p80/p95): reject n = 31 **41.5 / 69.75 / 88.25**; breakout_retest 9: 25.75 / 43.5 / 62.25; acceptance 6: 17.6 / 89.75 / 140; sweep_reclaim 6: 20.5 / 25 / 41.7; reclaim 5: 16.25 / 21.4 / 26.4. In ATR5m: reject 1.40 / 2.66 / 3.48; all trades 1.03 / 2.47 / 3.13. Only reject clears n = 30. A target at the cell's p60 (≈1.5 × ATR5m for reject) would have been reached by roughly 45 % of reject trades on this tape **[T] in-sample** — and I say "roughly" because I am reading it off a 31-trade percentile table.
4. *A fixed multiple.* 2 × stop is what the book effectively has, and 14 of 62 reached it. A fixed multiple of ATR (1.5×) is better than a fixed multiple of the stop only because the stop is itself 1.5 × ATR now.

**My ruling as reviewer:** targets from the map, bounded by the cell's MFE p60 in ATR units, with R:R as a filter — and the cell table must reach n ≥ 30 per condition × session before the bound is trusted. The number to watch is the target-hit share (10/65) and, second, the payoff (1.62). Until then, the honest 2R fixed target stays, and the planner's target field is advisory only — which is nearly what `planner_prompt.go:722` already says ("target_chain is GUIDANCE … never enforced at execution (D2 ruling)") while the *arm* target is enforced as the bracket.

**On the giveback question** (part-a §1.3): winners give back a median 4 pts (p75 8.5) — they exit at the target, near their best price. There is no giveback problem to solve with a trail; there is a *target-too-far* problem, which a trail cannot solve either. The suspension of BE/trail stands, on this evidence, and I would keep it suspended.

---

## Q5 — THREE THINGS A THIRTY-YEAR HAND SEES THAT THE BUILDERS DON'T

**1. The confidence field is a constant, and a gate on it is theatre.** Of 107 `open_*` proposals since 08-19, **90 carry a confidence of 62–65** (84.1 % [76.0, 89.8]); 39 say exactly 62; the full range is 60–70 (`q19_confidence.out`). `min_confidence` is 60 on the bound strategy. The executor writes the smallest number that clears the floor plus a little politeness. Nobody should ever size, rank, or gate on that field, and the 60/65 defect the controls census argued about is moot — the model will write 66 the day the floor is 65. Query: `q19_confidence.py` (the confidence Counter).

**2. Every loser dies at the stop and the store cannot tell you why anything died.** 37 of 65 exits at the initial stop, 10 at target, 18 unattributed; `close_reason = 'sync'` on all 65; no `close_*` decision within four minutes of any of the 18. A desk that cannot answer "was that an EOD flat, an invalidation, or the AI panicking?" per trade cannot improve its exits. And the exits are where the money is: losers with MFE ≥ 1 × stop before dying — trades that were a full R onside and came all the way back — number 8 of 41 (533, 534, 539, 548, 552, 554, 563, 567); losers that never showed 0.5 × stop number 21 of 41. Two different losers, one stop. A time/no-progress scratch fixes the second kind; a partial or a map target fixes the first. Queries: `q25_mfe_atr_targets.out` (exit classification), `q29_neither_exits.out`.

**3. The ledger said "cancelled"; the broker held nine.** On 09-04 the NY v3 S2 arm (a `reclaim` short, placed as a sell-stop at 29590.52) was placed at 10:10, 10:15, 10:16, 10:20, 10:25, 10:26, 10:28, 10:30, 10:32, 10:34, 10:36, 10:38, 10:40, 10:42 … 10:52 CT (`📌 armed S2 → WORKING stop-entry 29590.52`, sixteen distinct signal ids in `data/nofx_2026-09-04.log`), each a new ledger row (ids 62–101, `q22_0904_arms.out`). Twenty S2 ledger rows carry a signal id between 10:05 and 10:52 (ids 38 and 62–101 odd), and they were closed as: "gate changed: rr" ×5 (38, 62, 65, 67, 70) · "no order_update within stale window" ×7 (73–85 odd) · "boot_sweep: pre-boot order, process restarted" ×8 (87–101 odd) · "cancelled in NT8" ×1 (102) — `q22_0904_arms.out`. (The draft's "(5) and (12)" tally was loose; this is the row-by-row count.) The broker's book (`nt8_order_snapshots` id 1604, 10:40:01 CT) shows **eight `Accepted` + one `Initialized` sell-stop at 29590.5, quantity 1 each**; `working_count` reads 8 or 9 in 94 consecutive snapshots from 10:38:01 to 10:53:11 (`q18_nt8_working.out`). A boot at 10:53:10 (rev `ebe07d5f`) swept "8 pre-boot arm(s)"; the dead-man watchdog at 10:55:10 still reported "6 working order(s) visible post-reconnect — entry cancellation is a Part A AddOn capability (no-op)". Had price traded 29590.5 between 10:40 and 10:53, a one-contract book would have sold up to nine MNQ. This ran on rev `70af663d` (booted 08:30:11), not the running binary; the reaper was rewritten as class 79 ("silence read as death") — but class 79 is about cancelling live orders, and this is the opposite failure: **placing duplicates while the previous is live**. A 30-year hand does not trust a ledger; he counts the book. That is what the snapshot table is for, and it says the count was nine.

(Three more, briefly, because the dispatch asked for three and I found six: the anchor stop never binds — `arm_stop_unanchored_0b = 220`, no anchored counter, so "structure stop" is a boot-line phrase, not a behaviour; the prompt carries three stale small-n crowns on both trees at `planner_prompt.go:601, 660, 737` after checklist class 16 outlawed them; and the day_type label is written two ways, `trend-down` and `trend_down`, which any downstream cell count will split — canonical-casing law.)

---

## RECOMMENDATIONS — in order

Each: what · why (evidence label) · what it takes · the number I'd watch. I cross-checked AUDIT-CHECKLIST.md (classes 15, 16, 25, 48, 79 are the neighbours) so nothing below is already fixed.

**R1. Take the target from the map and the excursion table; make R:R a filter.**
*Why:* 10/65 exits at target; 8/59 MFE reached the planned distance; planned 3.44 × ATR vs MFE p50 1.03 × ATR; 47/132 armed R:Rs in [2.0, 2.3) **[T]**; the prompt states the floor the model must clear **[A]** `planner_prompt.go:733`.
*What it takes:* **data first** — a replay script over the 65 trades (bars exist) scoring "exit at next seated opposing level" and "exit at cell MFE p60 (ATR units)" against the recorded path; then a prompt change (drop the R:R sentence from the feasibility contract, keep the stop-floor line) and a gate change (compute R:R against the machine target) — an **owner ruling** on whether 2.0 stays the floor once the target is honest.
*Watch:* target-hit share (15.4 % now), payoff (1.62), and the reject cell's mean CI.

**R2. Keep the decision path closed to market entries (strict) — permanently for the confirm-close class.**
*Why:* decision path 13/36/2 −$750; `1x5m_close` 5/23/1 −$1,135 (17.9 % [7.9, 35.6]); arm fills 6/5 +$94.50 **[T]**. Owner ruling 2026-09-03 already did this (`entry_gate.go:160` running / `:184` dev) **[A]**; it refused 13 market entries on 09-04 (`q16_decision_census.out`) **[A]**.
*What it takes:* nothing new; a **ruling** that `wait_confirm` arms may chain on `touch`/`1m_mss` but not on `1x5m_close` (the same premium re-enters at the arm layer, `armed_executor.go:410-426`).
*Watch:* fills by path; arm-fill win rate (6/11 now).

**R3. Session and vol-regime gating by a measured switch: NY only until ASIA/LONDON each reach n ≥ 30 at ≥ break-even.**
*Why:* ASIA 2/13/1 −$552; low-ATR tercile 4/17 −$795 **[T] n < 30 no verdict, sign uniform**; the code ships ASIA/LONDON disabled and the bound strategy re-enables them in `sessions[]` (P18) **[A]**.
*What it takes:* an **owner ruling** reconciling `sessions_enabled=["NY"]` with `sessions[].enable=true` (they disagree today), and a **knob**: minimum ATR5m at arm time (the terciles cut at 22.5 / 27.9 pts).
*Watch:* per-session expectancy at n ≥ 30; per-tercile.

**R4. A no-progress scratch (data first).**
*Why:* 38/41 losers die at the stop; 21/41 never showed 0.5 × stop; losers' median hold 21 min; winners' MAE p50 0.37 × stop **[T]**; a fade that has not moved is a fade that is being absorbed **[I]**.
*What it takes:* replay "MFE < 0.3 × stop after 15 min → scratch at market" on the 65 trades; if the average loss falls without losing more than one winner, a **code** change in the position monitor (a new exit reason — see R6 first).
*Watch:* average loss (−$70.76), share of losers at the stop (92.7 %).

**R5. Fix the D1′ writer before any level-kind verdict; then wait four weeks.**
*Why:* 677 rows = 370 touches; the same touch is written once per plan-version read (P2) **[A]**; ambiguous share 30 %; only VWAP reaches n = 30.
*What it takes:* **code** — dedupe on (trader, symbol, level_kind, opened_at_ms) at the writer (`trader/detector_record.go`), or stamp the read that recorded it so consumers can dedupe; a **prompt/knob** freeze on the seated-kind list until the corpus is clean. The prior review's "retire six kinds" is as unsupported as keeping them (n ≤ 5 each after dedup).
*Watch:* distinct touches per version; per-kind n.

**R6. Record the exit cause.**
*Why:* `close_reason = 'sync'` on 65/65; 18 exits unattributable **[A]**.
*What it takes:* **code** — write `target|stop|eod_flat|invalidation|ai_close|reconcile|oco_partial` into `close_reason` (or a new column) at the close-sync call sites. The draft cited `trader/close_sync.go`, which does not exist; the real anchor is `trader/ninjatrader/reconcile.go:259, 289, 460`, which call `store.PositionStore.ClosePosition(id, exitPx, exitOrderID, pnl, fee, closeReason)` (`store/position.go:418`, write at `:427`) with the literal `"sync"` as `closeReason` and the literal `0` as `fee`; the partial-close path `ReducePositionQuantity` (`store/position.go:524`, write `:557`, via `store/position_builder.go:133`) hard-codes the same `"sync"` **[A]**. `exit_order_id` does not rescue it: on the 65 rows it is an NT8 order UUID on 60, the literal `Close` on 4 (526, 530, 538, 575) and `netting_fill` on 1 (578) — `q32_exit_order_id.out`.
*Watch:* the share of 'other' exits (28 % now) going to zero.

**R7. Anchor the fade's stop to the faded level's measured sweep, with the ATR floor as a ceiling.**
*Why:* `arm_stop_unanchored_0b = 220`, no anchored counter **[A]**; winners' MAE p80 0.71 × stop; `touch_episodes.penetration_pts` already measures how far each level gets swept **[A]**.
*What it takes:* an **owner ruling** (the 1.5× floor is owner-ruled 2026-09-02) and **code** in `arm_stop_anchor.go` to anchor to the scenario's own level; add the anchored counter.
*Watch:* anchored/unanchored ratio; losers' MAE/stop distribution.

**R8. Stop gating on confidence, or calibrate it.**
*Why:* 90/107 proposals at 62–65 **[T]**.
*What it takes:* **knob** (min_confidence → 0) or a **data** task: bin realized outcomes by stated confidence at n ≥ 30 per bin.
*Watch:* the distribution's width.

**R9. Remove the stale crowns from the prompt.**
*Why:* `planner_prompt.go:601, 660, 737` (both trees) quote "75 % win, +665 this week", "ONH entries 14 · 21.4 % win", "0 % win" — class 16 forbids exactly this; the live cells are 58 % (n = 12) and 20 % (n = 6) **[A] [T]**.
*What it takes:* **prompt** — delete, or render the live `/api/expectancy` cell with its n.
*Watch:* nothing; hygiene.

**R10. Book the commission.**
*Why:* `fee = 0` on 71/71 rows **[A]** — the close-sync call sites pass `0` as the fee argument (`trader/ninjatrader/reconcile.go:259, 289, 460`) **[A]**; the expectancy is gross; ≈ −$2.5/RT moves the mean from −8.68 to ≈ −11 **[I]**.
*What it takes:* **config/data** — a synthetic per-side fee at close-sync, or the SIM commission if the AddOn can read it.
*Watch:* net mean.

**Evidence I would demand before anything else:** (i) the R1 replay; (ii) the adverse-selection test in Q3 §5 (fills vs. near-misses); (iii) a clean D1′ corpus of four weeks. All three are read-only weekend work on data that already exists.

---

## DISAGREEMENTS WITH EXISTING REPORTS — with the number

- **veteran-review.md:84–88, part-a:73–76** — "RTH-L breaks 68 % (n = 63), statistically separated": the corpus is duplicated per plan-version; deduped RTH-L is 4/13 hold, no verdict. I disagree with the verdict and with the "retire six kinds" that follows from the same table (part-a §9, `veteran-review.md:397`): n ≤ 5 per kind after dedup.
- **part-a:109** "median 2.55:1" — 2.36 all-era (n = 132), 2.54 in that two-day window; the direction (targets cluster just above the floor) is right and stronger than they said: 35.6 % in [2.0, 2.3).
- **part-a:135** "no winner ever exceeded 0.661× the floor" — id 529 at 1.22×; 16 of 17 below 0.67×.
- **part-a:163** "3 of 36 reached 2R" — 5 of 59 off the minimum stop; 14 of 62 off the actual stop.
- **part-a:190** "median winner hands back 24 %" — paired giveback median 4 pts; no giveback problem.
- **part-a §1.3** "re-enabling BE/trail would not help" — I agree, with the sharper number: BE at +40 would have touched 8 of 41 losers (those with MFE ≥ 1 × stop — the stops here are 25–45 pts) and turned an unknown number of the 20 winners into scratches; the suspension stands.
- **mc-drawdown.md:173** "~1,810 trades needed" — ≈1,050–1,070 at n = 65; the figure is a function of a 0.09-sd effect and moves with every trade.
- **SYSTEM-MAP §12** "scan interval default 3, min 3" — the bound trader row holds 2 and the log runs on 2.
- **two-day-audit.md:35** "113 min 51 s" — verified from the log to the minute (12:23:33 → 14:18:24 in the file; 12:24:33 in `log_events`).

---

## SURPRISES — found, never acted on

- **S1. Nine duplicate resting sell-stop entries on the NT8 SIM book, 09-04 10:38–10:53 CT** (Q5 §3). Ledger rows 73–101 marked cancelled; broker snapshots 1596–1689 show working_count 8–9; rev `70af663d`; swept by the 10:53:10 boot (`ebe07d5f`); watchdog at 10:55:10 still saw 6. Not in AUDIT-CHECKLIST.md (grep for duplicate / working_count / dead-man finds no class). Queries: `q18_nt8_working.out`, `q22_0904_arms.out`, log lines in APPENDIX A22.
- **S2. `nt8_order_snapshots` ticks on Saturday** — 4,433 rows at 09:41, 4,461 at 09:55, 4,468 at 09:58, 4,616 at 11:12 and 4,619 at 11:14 CT, `working_count 0`, `[]` (`q31_asof_resume.out`, `q33_snapshot_epoch.out`). The AddOn is emitting periodic snapshots with the market closed; harmless, and it says the AddOn is alive.
- **S3. `arm_stop_unanchored_0b = 220`** with no anchored counter (P16).
- **S4. The trader is bound to a strategy row whose `is_active = 0`** (`a5b7662e` "MNQ"), while two rows named "MNQ SIM Default" carry `is_active = 1`; `is_active` does not mean bound. And on that row `sessions_enabled = ["NY"]` contradicts `sessions[].enable = true` for ASIA and LONDON (P18).
- **S5. `close_reason = 'sync'` on every usable era row** — the exit cause is not stored (Q5 §2).
- **S6. Two spellings of one day-type label**: `trend-down` (13 trades) and `trend_down` (4) in `plans.doc.day_type` — canonical-casing law (checklist 28).
- **S7. The prompt still carries small-n crowns** at `planner_prompt.go:601, 660, 737` on both the running rev and dev, after class 16 (R9).
- **S8. `bars` holds 12,301 ES 1m/5m rows** the system never trades; the MNQ 1m table has two conventions (`''` to 09-01 12:04, `epoch_floor` from 09-01 06:31) that overlap for six hours — `DISTINCT open_time_ms` is required for any recompute (I used it).
- **S9. The 09-02 18:48–18:52 min-SL refusals quote a threshold of 450.56 pts** (= 1.5 × a DAILY ATR) — the "one gate, two ATRs" bug the two-day audit found; SYSTEM-MAP §6 says fixed; three decision rows (36640–36642) carry the scar (`q19_confidence.out` tail).
- **S10. `decision_records` has 1,301 unparsable `decision_json` rows since 08-19** (of 7,810) — empty or non-JSON; they are not `wait`s and not errors by `error_class` (`q16_decision_census.out`).
- **S11. Row 38 of `armed_orders` stamps the 09-04 NY v3 S2 arm as `sweep_reclaim`; rows 62–102 (same plan, same version, same scenario) stamp it `reclaim`**; the plan doc says `reclaim` (`q22_0904_arms.out`).
- **S12. The Studio says BE and trail are ON; the binary says OFF.** The bound strategy row carries `breakeven_enabled: true, breakeven_trigger_points: 40, trailing_enabled: true` (`q28_binding.out`), and the running binary's boot line says `BE=off · trail=off` because `EXIT_MECHS_SUSPENDED` defaults to suspended (`trader/exit_mechs_suspend.go:35-43`; key absent from `.env`). Both are true at once; a knob that reads ON while the wire seam holds it OFF is a guide-content-law case (a toggle that lies about the running binary). Not acted on; not in the checklist by that name.

---

## APPENDIX — every query and its output, or the committed file that holds it

Scratch directory: `~/nofx-analysis/vet-01-0905/` (scripts and outputs), mirrored to `docs/superpowers/reports/2026-09-05-vet-01-way-it-trades-data/` (59 files, 465 KB, secret-scanned). Store opened as `file:/home/hoang/nofx/data/data.db?mode=ro` throughout; no write, no restart, no config touched.

| id | what | file(s) |
|---|---|---|
| A0 | store as-of: row counts + max timestamps of every table used | `q00_asof.sql` → `q00_asof.out` |
| A1 | era distribution — FIRST run with the WRONG epoch 1755234000000 (2025-08-15), kept as the record of the error; the per-day table exposed it | `q01_era_dist.sql` → `q01_era_dist.out` |
| A1b | era distribution with epoch 1786770000000 (2026-08-15 00:00 CT): by source / session / side / close_reason / trader / correction note; the 20 non-system rows | `q01b_era_dist.sql` → `q01b_era_dist.out` |
| A2 | armed_orders book, non-test, with computed R:R and stop pts, CT-normalised timestamps; state-reason tally | `q02_arms.sql` → `q02_arms.out` (all 67 rows); `q02b_arms.csv` (61 non-test) |
| A3 | plans: latest 12, lifecycle counts, one active doc's schema walk | `q03_plans_recent.out`, `q03_sample_doc.json` |
| A4 | ab_confirm_log by condition/outcome/rule; touch_outcomes by kind × ordinal (RAW, duplicated) and by session/k/band | `q04_abconfirm_touch.sql` → `q04_abconfirm_touch.out` |
| A5 | touch_outcomes duplication probe: distinct touches, top duplicated (kind, price), per-version counts | `q05_touch_dup.sql` → `q05_touch_dup.out` |
| A6 | ab_confirm_log detail: is_counterfactual, net_pnl ranges, the 60 resolved rows | `q06_abconfirm.sql` → `q06_abconfirm.out` |
| A7 | bars coverage by tf/convention/day | `q07_bars_coverage.out` |
| A8 | misc premises: created_at ≥ era, cited non-empty, mae/mfe, bars dup by symbol, NT8 snapshot coverage, orders tables | `q08_misc.sql` → `q08_misc.out` |
| A9 | the 71 era rows with every plan column | `q09_era_trades.csv` |
| A10 | decision_records LIKE-based census (over-counts; superseded by A16) | `q10_decisions.sql` → `q10_decisions.out` |
| A11 | all plans since 08-15 parsed: condition/direction/quality/arm mix; planned R:R distribution; per-session; the 09-01/02 window | `q11_planned_rr.out`, `q11_scenarios.csv` (738 rows, numeric fields only) |
| A12 | level_stats, touch_episodes, planner_read_facts corpora | `q12_other_corpora.sql` → `q12_other_corpora.out` |
| A13 | the enrichment: trade ⋈ plan doc scenario ⋈ armed_orders ⋈ decision_json ⋈ NT8 snapshot; initial stop source per row | `q13_enrich.py` → `q13_trades_enriched.csv`, `q13_enrich.out` |
| A14 | duplication mechanism (one RTH-L touch under three versions) and (kind, opened) dedup; conflict counts | `q14_touch_dup_mech.out`, `q14b_touch_conflicts.out` |
| A15 | touch_episodes hold rates by family × ordinal × session, Wilson | `q15_touch_episodes.py` → `q15_touch_episodes.out` |
| A16 | JSON-parsed decision census: actions, per-day opens, risk_check_error and execution_log refusal classes | `q16_decision_census.py` → `q16_decision_census.out` |
| A17 | `/api/config/resolved` knobs of interest (status only — the endpoint carries no values) | `q17_resolved_knobs.out`, `q20_knob_values.out` |
| A18 | NT8 working_count histogram 09-04; the 9-order book | `q18_nt8_working.out` |
| A19 | confidence distribution of all 107 proposals; the six two-ATR / R:R gate rows | `q19_confidence.out` |
| A21 | THE statistics pass: book, by session/side/condition/path/hour/rule/day_type/bias/ATR-tercile, holds, stops in pts and ATR, planned R:R, excursions in pts/stop/ATR, 2R-reach, giveback, stop counterfactual, per-day, reject × session, mae/mfe spot-check | `q21_stats.py` → `q21_stats.out`, `q21_trades_final.csv` (65 rows) |
| A22 | 09-04 arm rows (kind/condition/state), plan v3/v5/v6 scenarios, the 10:10–10:56 log timeline | `q22_0904_arms.out` |
| A23 | candidate_pool by kind; fee column; strategies table shape | `q23_candidate_pool.out` |
| A24 | strategy config values (sessions, limits, gates) for every strategy row | `q24_config_values.out` |
| A25 | MFE in ATR units by condition; target/stop hit classification; post-0B subset; losers' MFE | `q25_mfe_atr_targets.out` |
| A26 | API expectancy rows (by condition, by session) for cross-check | `q26_api_rows.out`, `expectancy.json`, `expectancy_by_session.json` |
| A27 | majority-vote dedup of touch_outcomes, Wilson by kind, pooled, session, recomputed ordinal | `q27_touch_dedup_majority.py.out` |
| A28 | trader → strategy binding, sessions[], entry quantities, arm-refusal counters | `q28_binding.out` |
| A29 | the 18 "neither target nor stop" exits with nearby close decisions | `q29_neither_exits.out` |
| A30 | the winner above the floor (529), the qty-2 row (574), Wilson intervals for every quoted rate | `q30_details.out` |
| A31 | store as-of re-run at resume (11:12 CT): all tables unchanged except `nt8_order_snapshots` | `q00_asof.sql` → `q31_asof_resume.out` |
| A32 | `exit_order_id` × `close_reason` on the 65 usable rows; the 18 'neither' exits | `q32_exit_order_id.sql` → `q32_exit_order_id.out` |
| A33 | `nt8_order_snapshots` first/last row in UTC and CT; the working_count ≥ 8 window | `q33_snapshot_epoch.sql` → `q33_snapshot_epoch.out` |

Shell one-liners not in a file:

- **A12 (silence)** `grep -hE '^09-03 12:2[0-9]|^09-03 14:1[0-9]' data/nofx_2026-09-03.log | awk '{print substr($0,1,14)}' | sort -u` → `… 09-03 12:23:33 09-03 14:18:24 …`
- **A14 (checklist)** `grep -oE '^[0-9]+\. \*\*' docs/superpowers/AUDIT-CHECKLIST.md | sed 's/\. \*\*//' | sort -n | tail -1` → `79`; `grep -nE '^## CLASS'` → line 2002 `## CLASS 75`.
- **boots** `grep -hE 'BOOT INTEGRITY' data/nofx_2026-09-0[234].log` → 09-04 07:38:40 rev 4f47ed2c · 08:30:11 rev 70af663d · 10:53:10 rev ebe07d5f · 13:25:47 rev 36648655 (running).
- **running vs dev** `git diff --stat 36648655 2a66d91c -- kernel/ trader/ provider/ store/` → 6 files, +286/−1 (entry_gate.go +26, risk_limits.go +65, auto_trader_planner.go +14, engine_analysis.go +13, two pin tests).
- **prompt crowns at the running rev** `git show 36648655:kernel/planner_prompt.go | grep -nE '75% win|21\.4% win|0% win'` → lines 601, 660, 735, 737.
- **stale reaper at the running rev** `git show 36648655:trader/armed_executor.go | grep -n reconcileStaleWorking` → present at :977, :1058, :1072 (same as dev).
- **strict leg at both trees (resume)** `grep -n 'if in.PlanMode == "strict"' trader/entry_gate.go` → dev `:184`; `git show 36648655:trader/entry_gate.go | grep -n …` → `:160` (running file 493 lines, dev 519).
- **exit-mech suspension on the running binary (resume)** `grep -E '🛑 exits' data/nofx_2026-09-04.log` → `09-04 13:25:47 … BE=off · trail=off · size=1 · re-arm-after-sweep=on (0B)`; `grep -c '^EXIT_MECHS_SUSPENDED=' .env` → `0` (key absent; value never printed).
- **checklist cross-check (resume)** `grep -c -i` over AUDIT-CHECKLIST.md for `working_count` → 0, `close_reason` → 0, `commission` → 0, `no-progress` → 0, `duplicate` → 1 (line 1926, unrelated), `dead-man` → 2 (lines 33/35, the AddOn reconnect check); classes 15 (Fantasy-R, line 113) and 16 (small-n crowns, line 120) are the only neighbours of R1/R9.
- **secret scan of the data directory (resume)** pattern grep (`eyJ…`, `sk-…`, `AKIA…`, `BEGIN`, `JWT_SECRET=`, `Bearer`) → empty; every `.env` value ≥ 12 chars grepped against the directory → the only hit is the key `DB_PATH` (a 12-character path, not a secret).

*Written 2026-09-05, Saturday, engine idle. Docs-only branch; no code, config, DB, knob, prompt, env or unit was changed; the main tree and its lock were not touched.*
