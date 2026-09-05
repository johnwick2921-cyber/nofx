# Veteran deep review — Section 05: EXECUTION

Owner: hoang · 2026-09-05 · branch `docs/vet-05-0905` · analysis only · SIM must stay.

## One-page summary

**Verdict: BROKEN stop-entry construction and guard; demonstrated SIM limit lifecycle; execution readiness remains unproven.** I use the requested veteran analytical lens without claiming a personal trading biography. I would resolve order correctness and observability before changing entry or exit policy. This review changes documentation only.

My three biggest problems:

1. **[A, S/BROKEN] Stop-entry has two separate defects.** C# passes the trigger as `limitPrice` with `stopPrice=0`; Go uses a limit-side predicate for stop entries. The observed stop-entry cohort has **0/21 fills, Wilson 95% 0.0–15.5%**, but those placements are repetitions of one scenario, not independent opportunities. Q6 verifies source, broker logs and snapshot ids. Zero fills alone would not prove the defect; the parameter positions do.
2. **[A, S/BROKEN] The bound strategy has its daily-loss enforcement disabled.** Strategy `a5b7662e-7bf7-49bb-9f09-7efa48f95ac8` has both master and daily-loss switches false. The reviewed dev revision adds an entry-block latch; it does not flatten an open position. A configured dollar amount is not an active cage. Source and fresh selected-field read: Q6 / q32.
3. **[A, A/BROKEN] Entry accounting lags the broker, while excursion telemetry is empty.** Arm 35 filled at 09:03:53, position 591 materialized at 09:05:14, and the arm-fill update was consumed at 09:05:35. The separate fill cache is not wholly blind, but the arm ledger lagged 102 seconds. `trade_excursions` still has zero rows (q31); stored position extrema cannot recover excursion timing.

My three biggest opportunities:

1. **Prove the order at NT8, then process its lifecycle independently of model calls.** Verify both price slots, side, quantity, account, acknowledgement, protective bracket and cancel/fill races. This is engineering correctness, not a forecast of trading edge.
2. **Keep the fade guard while measuring alternatives.** The prior **5/15 cancellations = 33.3% (Wilson 15.2–58.3%)** is a different window. Since September 2 it is **3/37 = 8.1% (2.8–21.3%)**, ids 33, 34, 36. The existing six-event replay cannot prove savings or justify replacing the guard. A four-tick stop-limit cap is an **[I] experimental judgment**, not an established MNQ edge.
3. **Repair measurement before changing exits.** The valid performance cohort contains 21 winners, 42 losers and two scratches; the MAE/MFE cohort loses one winner because its measurements are NULL. Keep these denominators separate. I would shadow a mechanical invalidation exit at one contract, retaining the existing stop/target/EOD baseline and SIM routing.

## Evidence contract and corrected premises

- **[A]** directly read source/store/log; **[B]** inference; **[C]** unverified possibility. Market-policy labels: **[R]** named research, **[T]** own-tape measurement, **[I]** analytical judgment untested here. No [I] proposal below is proven edge. S/A are severity labels; PROVEN/BROKEN describe evidence, not approval to deploy.
- I resumed the existing worktree, preserved its dirty draft and query scripts, and saved the draft in original scratch. Source citations are pinned to **`2a66d91c`**, not whatever dev becomes later. The original session health artifact records running **`36648655`**. `q32_sources.out` verifies the stop-entry/arm/provider execution files are unchanged between those revisions. Risk wiring differs. I did not refresh or change the running process.
- I read `agents.md`, `SYSTEM-MAP.md`, and `AUDIT-CHECKLIST.md` including PART 2 and the reaper/ledger fixes. The user's instruction to keep this existing dev-based worktree takes precedence over the checklist's running-revision worktree convention; I compare revisions explicitly. The parent integration tree and main lock are untouched.
- All database connections use `file:/home/hoang/nofx/data/data.db?mode=ro`. Final q31 also enables query-only. Times are CT (UTC−05 for this cohort), unless explicitly UTC. Performance window: August 15 00:00 through September 5 00:00 CT. Exclude test source, NULL corrected P&L, unresolved/unresolvable rows and correction notes. **71 era rows, 65 eligible**; excluded position ids **572–574, 576, 577, 579**. NULL excursion fields are excluded only from excursion calculations.
- Performance row set **P** = `521–571, 575, 578, 580–591`. Each figure below names ids directly or an explicit keyed manifest. `q31_verified.json` supplies position, fill, bar and scenario keys; `q33_metrics.json` supplies rates and bar-bucket ids. `plans` has a composite `(plan_id, version)` key; scenario/leg labels are not globally unique ids.
- Scratch and data are under `~/nofx-analysis/vet-05-0905/` and `docs/superpowers/reports/2026-09-05-vet-05-execution-data/`. **q31–q34 supersede conflicting q01–q30 conclusions**, which remain preserved as measurement history. The original scripts contain limitations listed below; preserving them is not endorsing their inference.
- Corrected premises: 227 era positions is not reproduced (71 before exclusions, q21/q31). The prior three filled arms are ids **24, 28, 35**, positions **584, 586, 591**, in September 1–3; only **35 → 591** falls in this review's since-September-2 arm cohort. `trade_excursions=0` is reproduced. Snapshot counts continue growing on Saturday: original q01 had 4,461; q31 read 5,178. Neither is a count of trading opportunities.

## Q1 — One arm end to end, with row ids

Arm **armed_orders.id 35** → position **trader_positions.id 591**, 2026-09-03, NY session.

**Authored [A].** `plans` row `plan_id=2026-09-03:NY:8d5c8af5…`, `version=2`,
`lifecycle=active`, `created_at=2026-09-03 13:45:05 UTC` (08:45:05 CT; `plans.created_at` is
UTC). Bias short, conviction medium. Scenario **S1**: condition `reject`, direction short,
trigger "A retest of 29285.00 OR-H stalls after the 08:30 failure; short the touch", confirm
`{rule: touch, ref_price: 29285, side: below}`, invalid "A 5m close above 29293.00 ONH voids
the fade", targets [29233.08, 29222.75, 29144.5], quality B, `arm {enabled: true, entry: 29285,
stop: 29340, target: 29144.5, wait_confirm: true}` (`q20_plan_scenarios.out`).

**Composed (0B stop) [A].** Go log 09-03 09:02:54 CT (in `data/nofx_2026-09-02.log` — the
file rotates on process start, not on date):
`🛑 arm stop NY S1 leg 1 short: stop 29354.91 (authored 29340.00 WIDENED) · anchor ONH
29293.00 → beyond 29293.50`. A later-cycle floor observation (not the placement-time ATR): `📓 read facts: void=20 · floor=67.6 pts
(1.5×ATR5m 45.10)` (09:06:54). The placement stop implies 69.91 pts of risk; the later 67.6-pt floor is a different observation. Composition drifted between
29349.90 (09:00:54) and 29354.91 (09:02:54). Code: `kernel/min_sl.go:33` `MinSLATRMultDefault =
1.5`; `:39` `MinSLTickClearance = 2` (the "beyond 29293.50").

**Gated [A].** Same second: `⚔️ arm S1 leg 1 wait_confirm MET (touch) — arming`; `🚫 no-chase:
arm S1 has no recorded touch — run stored NULL, judged on dist alone (0.00×ATR)`;
`armGateVerdictFor` (`trader/armed_executor.go:430`) returned "" (no `⚔️ arm REFUSED` line);
`⚔️ armed NY S1 leg 1 short limit 29285.00 SL 29354.91 TP 29144.50`. R:R at composition =
(29285 − 29144.5) / (29354.91 − 29285) = 140.5 / 69.91 = **2.01** — close to
the 2.0 floor (this matters in Q3).

**Placed [A].** `📌 armed S1 → WORKING limit 29285.00 signal=f2b1eb20-… (band ±100t)` —
`armedPlaceTicks()` returns 100 (`trader/armed_executor.go:51`) → a 25-pt band around the level;
`runArmedPlacement` `:963-972` → `PlaceLimitEntry` (`trader/ninjatrader/tcp_trader.go:433`).
Ledger row 35: `state=working`, `signal_id=f2b1eb20`, `kind=limit`, `created_at 09:02:54 CT`.

**NT8 [A]** (`log.20260903.00000.txt`): `09:02:54:664 signal f2b1eb20 routed to account
'Sim101' (resolved, sim, connected)`; `:672 Submitted … Action='Sell short' Limit price=29285
Stop price=0 Quantity=1 Type='Limit' Time in force=DAY`; `:780 Accepted` then `Working`;
**`09:03:53:694 Filled`** (59 s after placement); `:702` bracket `-sl` `Type='Stop Market'
Limit price=0 Stop price=29355` and `:707` `-tp` `Type='Limit' Limit price=29144.5`; both
Accepted `:811`. Note the SL at **29355** = round-to-tick of the composed 29354.91.

**order_update → Go [A].** `📡 armed order_update summary (1-line/min): frames=1
initialized=1` and `⚡ armed fill S1 @ 29285.00 (entry_class=armed_fill — stale_reeval NOT
applied)` at **09:05:35** — 102 s after NT8's Filled — because the cycle that placed the order
went straight into `🤖 Requesting AI analysis…` at 09:02:54 and came back at 09:05:35 (`⏱️ AI
call (reasoning=fast→low) duration: 157.03 seconds`; `⏱ cycle overran the scan interval
(2m41.166s > 2m0s)`). `consumeArmedOrderUpdates` is called only from `trader/armed_executor.go:978`.
Ledger row 35 → `filled`, `fill_price 29285` (`:1206-1207`).

**Materialized [A].** `trader/ninjatrader/reconcile.go:395` 09:03:54 `NT8 holds UNTRACKED position
MNQ SHORT @ avg 29285.00 (no open row) — materializing after 60s if it persists` →
`trader/ninjatrader/reconcile.go:436` **09:05:14** `MATERIALIZED untracked NT8 position MNQ SHORT qty=1 @
29285.00 (acct=Sim101)` → `trader_positions.id 591`, `source=reconcile`, `entry_time
09:05:14 CT` (81 s after the fill — the reconcile time-stamp trap), `plan_band=armed_fill`,
`cited_scenario_id=S1`, `plan_matched=1`, `adherence_grade=A`. The executor's own
`materializeArmedEntry` (`:1233`) ran at 09:05:35 and reported `position row not materialized
yet — stamp pending (reconcile completes it)` — 21 s after the reconcile row already existed
[B: the stamp's lookup did not find it; the final row is stamped anyway].

**Ledger vs broker stop [A].** `armed_orders.35.stop_px = 29351.6284728996` while NT8's stop
is 29355: the row's prices were rewritten in place by the 09:05:35 cycle's `UpsertArm` (stop
29351.63 in that cycle's composition line) after the broker already held 29355. This is the
in-place-overwrite class the two-day audit called D12/D40; fixed the next morning by
`3b8d6cd6` (`store/armed_orders.go:201-208`, "refusing to rewrite — the row is working").

**Invalidation while open [A].** 09:15:01 `🎯 scenario S1 → ≈invalidated @ 29285.00 (price
accepted through the level against the trade — … display-only estimate, never exec`. The
09:10 5m bar (q34.trace_5m_bar, explicit bar id) closed 29301.50 above ONH 29293. Position open at ≈ −16 pts. Nothing acted.

**Excursions.** `trade_excursions`: 0 rows. `trader_positions.591`: `mae=75.0`, `mfe=43.5`.
Close-time grader line 09:20:47: `📐🎓 MNQ close: MAE 75.00 / MFE 43.50 pts · adherence D
(off-plan) cited="" matched=false` — the stored row says A / S1 / matched=1 (see Surprises).

**Exit [A].** NT8 `09:20:45:677 … -sl New state='Filled' … Stop price=29355 Type='Stop
Market'`; `-tp Cancelled :782`. Go, same second: `trader/ninjatrader/close_sync.go:180 📊 exit fill recorded: MNQ
SHORT qty=1.00 @ 29355.00 (tick-exact, pnl -140.00)`; `store/position_builder.go:162 ✅ Full close
… PnL: -140.00`; `trader/ninjatrader/close_sync.go:196 📕 NT position closed … reason=sl`. `trader_fills.433`
(`nt8-exit-591`, BUY 29355, realized −140). `trader_positions.591`: `exit_price 29355`,
`pnl_corrected −140.0`, `close_reason 'sync'`, `pnl_correction_note "T7 close-path:
pnl_corrected = recompute (Δ+0.00 vs realized)"`. The 09:20 1m bar: o 29343.5 **h 29360.0**
l 29342.25 c 29348.0 (`q30`) — SIM filled the stop-market at exactly 29355 while the bar traded
5 pts (20 ticks) beyond it. The logs show same-second close synchronization in this case; the generic fill consumer also
updates its cache asynchronously (`trader/ninjatrader/tcp_trader.go:232-245`). I distinguish that
cache from the arm ledger, whose `order_update` drain is cycle-gated.

**What the trace teaches.** Nine hops, four clocks (plan UTC, log CT, NT8 local, epoch-ms), one
real defect surfaced by the timing (arm ledger updated 102 s late), one by the prices (ledger stop ≠
broker stop, since fixed), one by the plan itself (invalidation with no exit).


**Integration correction — actual stop slippage [A].** For position **591**, the accepted far-side protective order `f2b1eb20-f89e-4a88-bcff-5ce971d17861-sl` had **Stop price 29355** at **09:03:53.811**, and filled at **29355** at **09:20:45.677**. Raw NT8 `log.20260903.00000.en.txt:6085,6260` is retained with line numbers in q32. Thus **29355 − 29355 = 0 pts / 0 ticks** against the broker-authorized stop. **29355 − armed_orders.35.stop_px 29351.62847289964 = 3.3715271 pts** is ledger-to-broker geometry drift, not execution slippage. The earlier composition log was 29354.91, rounded to the accepted 29355. I disagree with any section-04/08 label of this 3.37-point difference as slippage; the far-side stop is the required reference. q34 independently records the subtraction and ids. The malformed *entry* stops on September 4 are a separate defect from this correctly formed September 3 protective bracket.

## Q2 — Fill quality, slippage and SIM limits

**[T] Coverage.** I audit each eligible position's entry and exit, not all historical `trader_fills` rows indiscriminately. `q31.fill_rows` contains both legs for every id in P, with price, timestamp method, matched fill id where available and containing `bars.rowid`. There are 65 entries and 65 exits; each leg has 62 containing bars. Missing bars belong to positions **521–523**. Entry timestamps use 51 fill-record matches, nine NT8-log matches and five explicitly late materialization proxies. Exit timestamps use 40 price/side/nearest-time fill matches and 25 position-close timestamps. Approximate joins are not execution-id proofs.

| Containing-bar statistic | Entry | Exit |
|---|---|---|
| Inside range | 58/62 = 93.5%, Wilson 84.6–97.5% | 61/62 = 98.4%, Wilson 91.4–99.7% |
| Adverse extreme | 1/62 = 1.6%, Wilson 0.3–8.6%; position 537 at low | 1/62 = 1.6%, Wilson 0.3–8.6%; position 585 at high |
| Favorable extreme | 0/62 = 0.0%, Wilson 0.0–5.8% | 0/62 = 0.0%, Wilson 0.0–5.8% |

All numerator and denominator ids are in `q31.fill_summary` / `fill_rows`. Extreme tolerance is half a tick (0.126 pts to avoid floating-point boundary loss). The four outside-range entries are **566, 569, 571, 580**, all late materialization proxies. The outside-range exit is **578**, a reconstructed netting close. These are timestamp/provenance flags, not proof that NT8 fabricated a fill. The entire containing bar includes trading before and after the fill: a high reached later is not a price available at submission.

The original q11 grouped fourteen non-system entries as “armed(limit)” without proving an order-type join for all fourteen. I withdraw its “14/14 first-touch limit fills” claim. **Arm 35** is a directly verified limit fill at its submitted price; q32 retains the source log. There is no queue-position measurement.

**[A] `slippageTicks` is transported but not used as a production Go metric.** C# stores the submitted reference (`VLTraderTCPClient.cs:947-954`), computes `(AverageFillPrice − origEntry)/tick` at **1376–1384**, and emits `slippage_ticks` at **1430**. Go declares the JSON field at `provider/ninjatrader/tcp_framing.go:86`; a production search finds no consumer of that field, only its declaration and a mock initialization (`q32`). This is signed raw price difference: adverse slippage for a short reverses the sign. Deserialization of the payload does not mean the metric is recorded or acted on.

**The NT8 logs do not report a slippage statistic.** A fresh scan of the ten September 3–4 log files, including localized duplicates, finds zero occurrences of “slippage” (`q32`, filenames and numbered matches retained). Arm 35's entry is exactly 29285 and stop exit exactly 29355. That establishes prices for this order, not a market-order slippage distribution. q15b's cycle-start-bar-to-fill drift is a separate proxy: it is neither the submitted reference price nor the broker's slippage field. I do not equate them or attribute all price movement to model latency.

**[I] What remains unproven in SIM.** Queue priority and actual available depth determine whether a live touched limit fills; a 1m range does not reconstruct either. A stop-market trigger does not cap its eventual execution price, while a stop-limit cap can leave an order unfilled. I do not estimate a live fill or slippage rate from this SIM sample. The official [NinjaTrader order-entry documentation](https://ninjatrader.com/support/helpguides/nt8/submitting_orders4.htm) distinguishes the stop trigger and limit offset; it does not validate a four-tick strategy cap.

Partial entries **are** emitted as `status="partial"` (`VLTraderTCPClient.cs:1341-1347`) and cached by Go (`tcp_trader.go:241-242`). The protective bracket is submitted only on full `Filled` (**1353–1355**). Therefore the already-filled portion, not the unfilled remainder, can await protection when quantity exceeds one. A one-contract record cannot validate multi-contract partial handling; I propose no size increase.

The 14:45 NY flat uses the session clock and flatten path (`trader/auto_trader_clock.go:454,512,557`). q33's **120 bars** in 14:40–14:49 average **2,394 volume / 9.91 pts range**, versus **120 opening-bucket bars**, **13,381 / 38.24**; exact bar ids are `q33.bar_buckets`. This is volume/range, not order-book depth or expected flat slippage. q26's sampled overnight buckets are lower-volume, but its 01:30 sample is not the 02:00 flat, and its 08:30 sample is the high-volume NY open. I withdraw the draft's claimed overnight-flat cost and its recommendation to change that knob.

## Q3 — Marketable guard and bounded alternatives

**[A] The fade guard remains intentional.** `trader/armed_executor.go:955-961` cancels a marketable limit before placement; **985–996** returns true for a long when price is below entry and a short when above entry. This prevents immediate entry after the level has been crossed. A marketable limit may receive price improvement; “worse than the limit price” in the code comment is not a correct limit-order guarantee. The policy concern is timing/invalidation, not permission to fill beyond a limit.

The prior window has **15 rows, ids 23–37**, including guard ids **25,27,33,34,36**: **5/15 = 33.3% (Wilson 15.2–58.3%)**. Its three fills are **24,28,35**, **3/15 = 20.0% (7.0–45.2%)**. Since September 2 the calendar-created cohort has **37 rows**, guard ids **33,34,36**, **3/37 = 8.1% (2.8–21.3%)**. q16b and q31 provide row manifests. Distinct leg keys give **3/11 = 27.3% (9.7–56.6%)**, but legs are not independent scenarios.

The “1.7 points through” premise is not established. For **33**, q16b's containing-minute close is 29167.50 versus entry 29166.80, a **0.70-point** difference; its separate bar-extreme calculation yields **1.70**. Neither records the exact price passed to Go at cancellation. The containing minute may finish after cancellation. Arms **27,36** even show the opposite sign against that minute's final close. I cannot substitute these proxies for the executor's actual decision price.

**[T, exploratory only] Guard replay audit.** The original q27/q28 covers **17,25,27,33,34,36**. Under its assumptions it reports six returns within five minutes (**6/6, Wilson 61.0–100%**) and five subsequent simulated stops (**5/6, 43.6–97.0%**). These intervals describe the modeled labels only. Its resting-limit total **−92.79 pts** and nominal “bounded market” total **−61.43 pts** are not actionable counterfactual P&L: B never enforces a bound, uses the final close of the cancel-containing bar, includes an invalid stop relationship for arm 17, and A skips the fill bar before checking exits. It assumes fills without queue evidence and allows a 120-minute horizon irrespective of intervening cancellation/session policy. I retain the scripts but reject “the guard saved $186/$123” and “every arm would have filled” as proven claims.

**My ruling [I]: retain cancellation for fades; shadow alternatives, no change now.** For a future continuation experiment I would compare caps **N=0, 2, 4 and 8 ticks**, with **4 ticks (1 point)** the provisional central case. This is a bounded experimental grid, not a research-derived MNQ threshold. A market order cannot guarantee an N-tick maximum; a stop-limit can cap price but sacrifices completion. Fresh executable bid/ask, valid stop-side geometry, tick rounding and R:R at the worst permitted fill must all pass before submission. Cancellation remains a real candidate and opportunity cost must include nonfills.

I would not place an unconfirmed arm early merely to improve apparent fill rate: that changes the plan contract. Measure first with timestamps, actual trigger, quotes, rejection reason and later tape; score both sides, retain all candidates, and bound same-bar ambiguity. A larger sample and held-out session-days are required before any policy recommendation (Q5 evidence gate). No cited research proves the chosen cap.

## Q4 — Funnel since September 2

**Use one currency before interpreting attrition.** The authored cohort is plans with `trade_date` September 2–4: **58 versions of eight plan ids**, **177 scenario-versions**, **57 enabled-arm versions**, **22 distinct `(plan_id, scenario)` opportunities**. Keys are `q31.funnel_scenarios`. Enabled-arm-version share is **57/177 = 32.2% (Wilson 25.8–39.4%)**; correlated revisions make the interval descriptive only.

| Stage, aligned to those plan dates | Count and evidence |
|---|---|
| Authored enabled-arm opportunities | 22 distinct scenario keys |
| Armed in ledger | 7 of those scenario keys; 10 leg keys; 36 ledger rows |
| Placed/sent | 4 scenario keys / 4 leg keys; 24 signal-bearing rows |
| Reached, whole-minute range proxy | 2 signal-bearing rows: 35 and 102; a boundary-inclusive proxy adds 37 |
| Filled | 1: arm 35 → position 591 |
| Won | 0: position 591 has `pnl_corrected=-140` |

Aligned row ids and keys are `q31.aligned_plan_dates`; range-touch bar ids are `q31.reached_proxy`. Authored opportunity → ledger **7/22 = 31.8% (Wilson 16.4–52.7%)**; ledger scenario → sent **4/7 = 57.1% (25.0–84.2%)**; sent rows → fill **1/24 = 4.2% (0.7–20.2%)**; filled → won **0/1 = 0.0% (0.0–79.3%)**. The touch-proxy sensitivity is **2/24 = 8.3% (2.3–25.8%)** versus **3/24 = 12.5% (4.3–31.0%)**; neither is the election rate of valid stops.

For the operational calendar view, include **arm 31**, created September 2 but belonging to the September 1 ASIA plan: **37 rows, 11 leg keys, 25 sent rows, five sent leg keys**. q16's “22 → 11” mixes scenarios with legs and this carry-in. It is not a valid authored conversion. Signal existence is evidence of sending, not by itself broker acceptance; the stop log and arm-35 trace provide the corroboration for the main findings.

The largest numerical loss in aligned opportunity count is **22 → 7**, but this does not identify a defective gate: some plan versions were superseded, scenarios can change across versions, and a version appearing in the table need not have been current for a placement pass. q10's rr/invalidated counters are repeated refusal observations, not a disjoint attribution of the fifteen missing opportunities. Conditions without enabled arms (`hold`, `acceptance`, `breakout_retest` here) cannot explain losses from a cohort already restricted to enabled arms.

The largest downstream operational defect is repeated stop-entry submission: the same September 4 NY S2 generates **21 of the 24 aligned sent rows** (ids in q31 `funnel_arms`, `kind=stop_entry`). The final order is **102 / signal 931a761a**. By unique opportunity there is one malformed stop-entry, not 21 failed trade ideas. The remaining aligned limit sends are **35,37,39**; the calendar cohort adds **31**.

“Reached” is deliberately a proxy. Ledger `updated_at` is not a broker cancellation timestamp, and arm 35's final ledger update extends beyond its fill. Whole-minute bars omit boundary touches (37), while containing bars can include pre-placement/post-cancel prices. I do not use either definition to promise fill probability. The original log also shows R:R cancellation/re-arm activity as ATR changes (q `log_0904_loop`, source `trader/armed_executor.go:430-450`); freezing authorized prices is a design candidate, not permission to relax the floor.

## Q5 — Exits, excursion proxies and the evidence gate

**[A] Current design.** Stop composition uses the farther anchor clearance and ATR floor; `kernel/min_sl.go:33,39` defines **1.5×ATR5m** and **two ticks**. Q1 gives the actual composed and broker-rounded prices. BE/trail wire changes are suspended by `trader/exit_mechs_suspend.go:14-44`, despite strategy toggles being true (fresh q32). The session-registry EOD close remains part of the baseline. I do not re-enable any mechanism.

**[T] The requested `trade_excursions` answer is unavailable: zero rows.** Stored `trader_positions.mae/mfe` is a separate bar-based proxy. `trader/auto_trader_clock.go:752-770` includes the containing fill bar in its path calculation; pre-fill and post-exit extrema can contaminate a boundary bar, and late reconciliation can omit genuine early movement. These values do not establish whether MFE preceded MAE, or how quickly a winner declared itself.

| Cohort | MAE p50 / p80 / p95, pts | MFE p50 / p80 / p95, pts |
|---|---|---|
| Winners with measurements, n=20 | 10.875 / 18.300 / 44.638 | 67.750 / 86.600 / 138.463 |
| Losers with measurements, n=42 | 40.375 / 50.800 / 75.000 | 16.625 / 28.400 / 47.775 |

Exact winner/loser ids are `q31.excursions`; percentiles use linear interpolation. Missing winner **566** has corrected profit **$97** and NULL extrema. It belongs in performance: **21/63 wins excluding scratches = 33.3% (Wilson 22.9–45.6%)**, average win **$114.67**, average loss **−$70.76**, payoff **1.6205**, sum **−$563.93**, mean **−$8.68** across all **65** ids P. These are corrected row P&Ls, not a new all-in commission model. q14's 20/62 win rate and 1.63 payoff improperly dropped 566.

Stop/target classification uses q14's nearest exit-price/side/time log match, independently of P&L sign. `q31.exit_reasons` lists every numerator id: **stop 44/65 = 67.7% (Wilson 55.6–77.8%)**, **target 13/65 = 20.0% (12.1–31.3%)**; four manual, three limit-close and **578 unknown** account for the remainder. Stops include profitable exits **557,570** and scratches **542,551**, so “stop-hit losers” is an incorrect label for all 44. Generic `close_reason=sync` is not an OCO reason; these reconstructed classifications retain matching uncertainty.

**Floor question: outside observed winner MAE in the small available proxy sample; causal verdict unavailable.** q14's ATR calculation included an unfinished 5m bar. q31 instead uses fifteen fully closed bars, a simple mean of fourteen true ranges, and excludes stale latest bars from the winner comparison. This is explicitly a reconstruction, not the exact entry-time engine ATR snapshot. Winner ids **555,557,560,569,570,575,578,580,581,582,584** have MAE/floor p50/p80/p95 **0.124 / 0.457 / 0.671**; **0/11** exceed the proxy floor (**Wilson 0.0–25.9%**). Each contributing bar id and freshness value is in `q31.floor_rows`.

This does not prove that the floor “saved winners,” that a tighter stop would help, or that winners move early. Selection on realized winners, changing historical exit rules and absent initial broker stops prevent those claims. Post-0B positions **589,590,591** lost **$155, $99, $140**, respectively: **−$394** across three trades. That is too little evidence to optimize the floor. Stop-out distance is not necessarily initial risk, particularly with historical BE/trailing and execution slippage.

**My exit design [I], shadow only.** Keep one contract and stop+target+EOD as the control. Add a strictly defined closed-bar invalidation candidate, with the original emergency protective stop always present; examine a time-stop only after excursion timing exists. Position **591** supplies a test case: authored invalidation is a 5m close above 29293 and a later close is 29301.50. Executing there would be hypothetical: a display invalidation is not a proven executable exit at that close. I do not claim the draft's $108 improvement, scale half a one-contract position, or increase to two contracts to enable scaling.

**Proposed evidence gate [I].** First demand 30 consecutive eligible SIM closes with broker fill time, initial accepted stop/target, entry ATR, MAE/MFE timestamps, boundary ambiguity and full lifecycle joins; report missingness by id. Thirty is a minimum instrumentation acceptance sample, not statistical proof. Then compare the unchanged baseline and candidate on identical held-out signals over multiple session-days, recording nonfills, forced-flat outcomes, slippage/fees and same-bar best/worst bounds. Advance only with a positive lower 95% session-day-bootstrap confidence bound on incremental net P&L, no breach of agreed loss limits, and zero unresolved lifecycle mismatches. Continue gathering data if uncertainty remains; report every rate with n and Wilson intervals. No live-money authorization follows from this gate.

**Regime and day-count boundary [A].** Independent mode=ro q34 groups the same **65 positions over 12 CME session-days**, versus **14 CT calendar dates**, for **−$563.93**. CME days roll at 17:00 CT; q34 labels them by the opening date and lists every position id/P&L in each bucket. The last stored position remains **591**, materialized September 3 at 09:05 CT (actual fill 09:03:53). Against the parent's supplied strict-boot boundary **September 3 11:10 CT**, the store contains **zero subsequent positions**. This verifies absence of later recorded trades, not the boot mode itself. There is no realized strict-regime performance sample (n=0; win rate/Wilson undefined). Neither the 65-trade mixed historical sample nor arm 35 proves strict execution profitability.

## Q6 — Technical live-money breakpoints (analysis only; SIM remains)

1. **[A, S/BROKEN] Stop-entry type/price contract.** `VLTraderTCPClient.cs:972-978` selects `StopMarket` but calls `CreateOrder(... qty, orderPx, 0, ...)`. The official [Account.CreateOrder signature](https://docs.ninjatrader.com/ninjascript/createorder) orders those arguments as limit price then stop price; the bracket code correctly uses `0, b.Sl` at **1822–1824**. q32's numbered NT8 records for signals **e38f1774** and **931a761a** show `Limit price=29590.5 Stop price=0 Type='Stop Market'`. q19 lists first snapshot ids for every distinct signal/type/price state; q31 maps all 21 stop ledger rows. Order 102 remained Accepted while whole-minute bars **445398,445402,445403,445406** spanned its intended trigger. **PROVEN malformed construction; observed 0/21 fills (Wilson 0.0–15.5%).** I do not assert how every live broker would reject or handle it.
2. **[A, S/BROKEN] Stop-side guard.** `trader/armed_executor.go:940` calls `limitMarketableWrongSide`; **985–996** uses `long: price<entry`, `short: price>entry`. A resting buy stop needs market below trigger; a resting sell stop needs market above trigger. Thus the predicate refuses the proper resting side and passes the already-crossed side. Equality is also unguarded. This is a two-sided source proof independent of the C# bug. `tcp_framing.go:221-228` uses a build-id comparison for far-side capability; it does not validate these price slots or stop geometry.
3. **[A, S/BROKEN] Daily loss.** Fresh q32 reads the bound strategy's master=false, daily=false, limit=450. `kernel/risk_limits.go:305-306` returns allow when the master is off. At `36648655`, analysis discarded the returned decision and could hold the AI cycle; at `2a66d91c`, `kernel/engine_analysis.go:183-199` records `RiskForceFlat` in the entry-block latch. `kernel/risk_limits.go:151-175` explicitly describes blocking entries, not closing positions. No revision activates false strategy switches. Owner policy must distinguish refusing new risk from flattening existing risk; a variable name is not protection.
4. **[A/B, A] Fill, size and AddOn bracket lifecycle.** Q1 proves lag to the arm ledger, not total broker blindness: Go has an asynchronous fill cache, and NT8 places the full-fill bracket locally. Arm placement hard-codes quantity one (`trader/armed_executor.go:945,965`). For larger sizes the entry partial is emitted before bracket submission (Q2). Code is not evidence of tested partial/cancel/fill ordering or bracket rejection recovery. Preserve one contract; demand deterministic event traces before any sizing discussion.
5. **[A/B, A] Reconnects and snapshots.** `trader/f12_leg4.go:209` defines a 30-second periodic snapshot; state-change snapshots supplement it. `q19` has actual snapshot ids and order prices. Reaper class 79 already replaced silence-as-cancellation with broker-book evidence; D5 already prevents rewriting a working ledger row (`store/armed_orders.go:201-208`). These fixes are in the reviewed running revision and should not be recommended as missing. They do not prove that in-memory AddOn pending bracket state survives an NT8 restart or that a bracket is exchange-hosted across every connection type. Those are **UNVERIFIED recovery cases**, requiring controlled SIM proof in a separately authorized implementation task.
6. **[I, A/UNVERIFIED] Execution clock and stale policy inputs.** The observed cycle-coupled ledger delay motivates an independent lifecycle consumer and fresh market inputs for mechanical checks. That does not justify arbitrary faster polling, relaxing gates, or an assumed latency target. Measure broker-fill→cache→ledger→bracket acknowledgement separately, retain tail timestamps, and test duplicates/out-of-order updates. “Two-minute cycle plus AI call” is not a universal timing formula; Q1's measured 102-second ledger delay is the evidence.

## Ranked recommendations — proposals only, none applied

| Rank / what | Why and label | What it takes | Number or event I would watch |
|---|---|---|---|
| 1. Repair stop price slots and both stop-side predicates | [A] Q6, malformed NT8 order and inverted guard; engineering correctness | C# + Go code, owner-coordinated future SIM verification | Every submitted stop's read-back trigger/side/qty; unresolved mismatches = 0; full fill/bracket/cancel trace |
| 2. Resolve daily-loss policy and enforce the chosen behavior | [A] fresh false flags and block-only latch; flatten choice [I] | Owner ruling, then scoped config/code work | Triggered-loss event blocks all entry paths; existing-position behavior matches ruling |
| 3. Separate arm lifecycle accounting from model latency | [T] arm 35 / position 591; architecture choice [I] | Idempotent Go lifecycle consumer and replay fixtures in a future code task | Fill-to-ledger tail latency; missing/duplicate state transitions by signal id |
| 4. Make excursion and slippage records complete | [A] empty table, unread field; measurement before policy | Diagnose existing hooks, persist signed/side-normalized slippage and initial broker prices | Missing eligible rows and unmatched ids; 30-close instrumentation gate from Q5 |
| 5. Keep the fade guard; remove the claimed replay “savings” from decisions | [T] only six modeled events; [I] retain control | Data first, bounded quote-aware shadow comparison | Paired held-out incremental P&L, nonfills, ambiguity; n and session-days |
| 6. Shadow closed-bar invalidation exits at one contract | [T] 591's authored invalidation; policy [I] | Explicit mechanical spec, telemetry, owner ruling before any change | Q5 gate; net incremental P&L confidence bound and loss breaches |
| 7. Exercise partial, restart and bracket-reject recovery in SIM | [A] partial bracket timing; recovery consequences [B] | Controlled future SIM harness/runbook, not an action in this review | Unprotected filled quantity/time, orphan orders, reconciliation mismatches |

I do not recommend removing already-fixed reaper/ledger safeguards, lowering the R:R gate, increasing size, turning BE/trail on, or replacing scheduled market flats on this evidence. Removing unsupported “first-touch fill,” “guard savings” and “winners move early” conclusions is itself necessary before policy selection.

## Surprises found, never acted on

- Both section-8 stop-entry suspicions reproduce independently in source and NT8 evidence; the order stack also masks a malformed-order problem.
- **566** is profitable but lacks extrema. The inherited draft incorrectly excluded it from win rate and payoff while including its dollars in total P&L.
- **31** is the September 1 plan carry-in; scenario keys and split legs were mixed in the inherited funnel. The aligned authored-to-ledger count is **22 → 7**, not 22 → 11.
- **37** changes “reached” classification when boundary bars are excluded. A ledger lifetime and a 1m range are not an order event timeline.
- The inherited four-tick proposal was supported by a replay that never actually enforced a tick cap. Its bar close also looked beyond the cancellation instant.
- Position **591** has stored A/S1/matched attribution while the close log reported D/empty/unmatched (`log_arm35`, q08); ledger stop rewriting is separately historical and fixed. Neither mismatch licenses rewriting history.
- Excursion hooks already exist (`trader/trade_excursion_hook.go:40,75`; `trader/armed_executor.go:1287`), but the table remains empty. “Add a writer” is premature until the missing path is diagnosed; a function's presence is not proof it ran.

## Reproduction and handoff

Run `python3 q31_verified.py`, `python3 q32_sources.py`, and `python3 q33_metrics.py`, and `python3 q34_integration.py` **from the original scratch directory**. q31 intentionally reads the preserved q11 entry-time and q14 exit-reason CSVs; their matching limitations are disclosed in Q2/Q5. q32 verifies source and selected raw NT8 lines afresh. No application test or broker experiment was run because this is documentation/read-only analysis.

The data directory preserves q01–q30 scripts/outputs, original slim API reads and log excerpts. New q31–q34 add the corrected cohorts, exact ids, closed-bar floor proxy, both fill legs, aligned funnel, source verification, Wilson intervals, CME-day grouping and broker-referenced stop slippage. q19 and q20 carry the broker-snapshot and authored-plan details for Q1/Q6. Scripts that predate q31 remain historical artifacts; use this report's qualified conclusions.

Limitations for the parent: one repeatedly submitted stop scenario; no actual trade_excursions rows; no live-money fills; some inferred fill/close times; no quote queue reconstruction; no valid capped counterfactual; source review pinned to the requested revisions. Fresh measurements support technical defects, not trading edge. **SIM stays, no code/config/env/DB/unit changes, no restarts/cancels, no parent-tree or main-lock operations.** The section branch is the deliverable; integration is the parent's task.

Validation: independent cohort/funnel/slippage assertions and Python syntax checks passed. Review copies have trailing whitespace normalized; original scratch evidence remains preserved. Only section documentation and evidence paths are staged. Application hooks/tests were not invoked for this docs-only evidence commit.
