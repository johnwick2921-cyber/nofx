# Round 24 Q4 — TARGET REACH (read-only research)

**Lane:** Chief (`claude-bcc69eb7`, worktree `~/nofx-chief`) · **Branch:** `docs/round-24-q4-target-reach` (claim `9ef3cae4` from `origin/dev` @ `7059bc74`) · **Dispatch:** CTO nofx-2a, 2026-09-17 ~02:10 CT · **Due:** 10:00 CT
**Owner's question:** "on a 400-point day the 12-seat table has no target ahead."

Every claim below carries an evidence tier — **[A]** directly verified (ran it / read the exact line / counted the rows), **[B]** inferred from strong evidence, **[C]** speculation — and every cell names its `n` and its first ids. `NOT MEASURED` is written where it applies; it is never folded into a number.

---

## 0. Verdict in four lines (cells in §3–§6)

1. **"No target ahead" is literally false and materially true.** On the 23 session-days with range ≥ 250 pt, every recorded 12-seat table had a level ahead of price in band (21/21 reads) — but the session's move ran PAST the table's farthest ahead level on **12/21 reads (57%; median overshoot 222 pt)**, versus 19/88 (22%) on sub-250 days. The published `doc.levels` show the same: 41/73 (56%) ran out on big days vs 53/231 (23%) control. [A]
2. **The missing targets were in band and in the detector's own universe.** On the 12 recorded big-day reads whose 12-seat table ran out, an in-band 1h/4h/D level on the move side beyond the table's farthest existed on 15/21 reads and **price reached one on 12/21** — the seats went to nearer levels; nothing had to be found beyond the band. (Oracle cell: it uses the move direction and sizes the miss; it is not a rule.) [A]
3. **The dispatched beyond-band rule barely helps.** The nearest beyond-band HTF level always exists (46/46 trended big-day reads) but sits a median 295 pt away and is reached on 16–36% of reads (7–9 hits per definition); on the ran-out reads it rescues 5–12%. The in-band FARTHEST variant (c′) reaches 19–33% and, aligned with the move, converts 5–11 of 17–24 cases. Neither is a seat rule yet; both are measured. [A] with the Dec-contract parity caveat on 15/73 big-day reads.
4. **Scenarios exhaust their chains on half of big-day reads** (88/182 = 48% vs 41% control; median last target 70 pt out, reached in a median 33 min) and **the replan budget was never the constraint**: 0 of 23 big session-days exhausted (1 spend total, 09-16 NY); level_event re-reads are free (40 on big days, max 7 in one session). The 4 pre-09-01 exhaustion markers all fall on sub-250 days. [A]

---

## 1. What was measured, from what

### 1.1 Sources (all read-only; the live store and the research store were opened by `sqlite3 -readonly` ONLY — no Go process opened a database) [A]

| input | rows | what it is | sha256 (`out-r24/in/SHA256SUMS`) |
|---|---|---|---|
| `bars` (live `data/data.db`) | see manifest | MNQ 1m since 2026-08-10 + 5m/15m/1h/4h/1d since 2025-01-01, `contract<>''`, `source NOT IN ('mixed','replay:off-scale')` — the `BarsBetweenOn` exclusions | manifest |
| `plans` | 397 session rows (400 minus 3 WEEKLY) | one row per planner read (version) over the last 30 trade dates 2026-08-15 → 2026-09-17; `doc` JSON carries `levels[]` and `scenarios[].target_chain` | manifest |
| `research_facts` object=`plan` event=`input` (live `data.db.research.db`, 98 GB) | 127 input snapshots, 2026-09-09 18:08 → 2026-09-17 | the planner's OWN `PlannerInput` per read: `Price`, `DATR`, `Levels` (the seated table with `distance`), `Pool`, `HTFZonesFull` | manifest |
| `system_config` `dayplan_replans_used:*` | 1 row | the class-35 recorded replan counter | manifest |
| `strategies.config.day_plan` | strategy `MNQ` | `proximity_filter_atr=1.0`, `max_levels=12`, `replan_cap=4`, `planner_timeframes=[D,4h,1h,15m,5m]` | manifest |

`plan_lifecycle_log` holds only `dormant`/`active` transitions (52 rows) — the re-reads themselves are the `plans` rows; that table is used for nothing else. `planner_read_facts` rows are the **void-scope** reads (`version=0`, `who:void`), not plan reads, and are not used.

### 1.2 Definitions (production functions wherever one exists) [A]

- **Session-day** = (`trade_date`, `session`) with ≥1 plan row; window = `DefaultSessionRegistry` (ASIA 17:00→02:00 next day, LONDON 02:00→08:30, NY 08:30→14:45 CT; `trade_date` for ASIA is the calendar date of the 16:30 read — verified against `plans.created_at`). **Range** = max(high) − min(low) over the window's 1m bars on the contract `ContractAt(window start)` (round-23 harness emulation). A session-day with no 1m tape is `NOT MEASURED`.
- **Big day** = range ≥ 250 pt (dispatch threshold).
- **Read** = one `plans` row; read instant = `created_at`.
- **Price at read** = the recorded `PlannerInput.Price` when the research store has that read's snapshot, else the last 1m close before `created_at` (`price_src` names which).
- **DATR** = recorded `PlannerInput.DATR`, else `kernel.DailyRangeProxy(ring of 2000 closed 1m bars, read time)` — the assembler's own function (`levels_assemble.go:615`).
- **Band** = `k × DATR`, `k = kernel.ResolveProximityK(1.0)` — the seating band `ScoreLevels` applies (`levels_score.go:465,475`: `|l.Price − price| ≤ band`).
- **Seated table** = recorded `PlannerInput.Levels` (the 12-seat table the model was shown) — recorded window only. **Doc table** = the plan's published `doc.levels[]` (what the plan card carries) — every read.
- **Eventual move** (`dir_exc`) = the side of the larger excursion from price-at-read over the remaining session `[created_at, flat)`; `dir_close` = sign(flat close − price) is reported as a probe.
- **Ahead** = a table level on the `dir_exc` side; **in band** = `|level − price| ≤ band`. **Ran out** = the excursion exceeded the farthest ahead level (or none was ahead); **overshoot** = excursion − farthest.
- **HTF universe at a read** = `kernel.DetectHTFLevels(fetch, [D,4h,1h,15m,5m], "MNQ", read time)` over the DB bars of the read's contract — the live detector, called the way `auto_trader_planner.go:2213` calls it. It is reconstructed for EVERY read because the recorded `HTFZonesFull` is `ScoreLevels`-filtered to the band and to zone kinds (`auto_trader_planner.go:2264`) and therefore cannot contain a beyond-band level. Parity probe: for recorded reads the reconstruction's in-band zone count is compared with the recorded count.
- **Trend at a read** = `kernel.ComputeStructureState` (the merged S1 engine, `kernel/structure.go:261`) on the read contract's own closed 1h / 4h / 1d ring (500 bars requested; a ring < 60 bars → `n/a`, which after the 09-14 roll is the Dec contract's 4h and 1d rings). `plan_bias` = the plan's own `doc.bias.direction` is reported beside it.
- **Candidate rule (c)** = among HTF-universe levels with tf ∈ {1h, 4h, D} on the trend side and `|level − price| > band`, the nearest; **reached** = a remaining-session 1m bar touches its near edge (zone `lo` for an above level, `hi` for a below level; the line price otherwise); minutes-to-hit from `created_at`. **(c′)** = the same with the FARTHEST **in-band** such level.
- **Scenario exhausted (b)** = a remaining-session 1m bar trades beyond the LAST entry of `target_chain` in the scenario's direction (long: high ≥ target; short: low ≤ target). A chain whose last target is not ahead of price-at-read in its own direction is `malformed` and excluded (counted).
- **Replan budget (d)** = per session-day: `level_event` rows, spending rows (`death_replan`/`owner_reread` — the ONLY class-35 spends per `store.ReplanBudgetBootLine`; `level_event` is free), the recorded counter, cap 4, exhausted = counter ≥ cap.

### 1.3 Lookahead law [A]

At-read fields (price, DATR, seated/doc tables, HTF universe, trend) use only bars closed before `created_at`, or the planner's own recorded snapshot of that read. Outcome fields (excursion, ran out, hit, minutes, exhausted) use only 1m bars with `open ≥ created_at` and `< flat`. One cell — the **oracle** in §5 — uses the eventual move direction on purpose and is labelled as such; it sizes the miss, it is not a rule.

### 1.4 Limits stated up front

- The **seated** table is recorded only from 2026-09-09 18:08 CT (research snapshot start). Before that, only the published `doc.levels[]` exists — a subset the model cited from the seated table. Both are reported; the recorded window is the anchor.
- 2026-09-14 is the Sep→Dec roll day; `2026-09-14:NY` has 102 1m bars (feed gap) and its 562-pt range is partly the contract basis — it is listed but flagged.
- After the roll the Dec contract's own 4h/1d rings are thin (24 / 11 bars at 09-17); stitching contracts would inject the ~290-pt basis as a fake swing, so those trends read `n/a` rather than wrong.
- The reconstruction's HTF universe is the DB's bars, not the live BarCache; the parity probe (§5) bounds the difference.
- `role` (target_only etc.) is assigned at seating, so a beyond-band level has none; the recorded role is carried onto in-band matches only.

---

## 2. Population


#### Population — session-days with a plan, last 30 trade dates (range = 1m high−low over the DefaultSessionRegistry window, own contract)

- session-days: 70 · measured (1m tape present): 68 · unmeasured: ['2026-08-15:NY', '2026-08-16:NY']
- plan rows: 397 = model reads 310 + marker rows {'planner_fail_closed': 33, 'demo_seed': 2, 'replans_exhausted': 5, 'owner_reset': 28, 'dormant': 12, 'rearmed': 5, 'e7_incident_kill': 2} (markers excluded from a/b/c, counted in d)
- range ≥ 250 pt: **23** session-days · < 250: 45 · median range (measured): 210.8
| session-day | range | 1m bars | reads | open→close |
| --- | --- | --- | --- | --- |
| 2026-09-14:NY | 562.00 | 102 | 8 | 28905.75→29439.50 |
| 2026-09-16:NY | 499.75 | 375 | 5 | 29386.75→29173.75 |
| 2026-09-01:LONDON | 473.75 | 390 | 6 | 29515.50→29111.50 |
| 2026-09-10:LONDON | 435.75 | 390 | 6 | 29477.25→29089.25 |
| 2026-08-20:LONDON | 432.50 | 390 | 2 | 29672.50→29350.50 |
| 2026-09-03:NY | 385.75 | 375 | 7 | 29249.50→29524.00 |
| 2026-08-28:NY | 374.75 | 375 | 7 | 29628.00→29503.00 |
| 2026-08-19:NY | 366.25 | 375 | 3 | 29715.00→29530.75 |
| 2026-08-23:ASIA | 355.75 | 540 | 4 | 29392.25→29164.00 |
| 2026-08-17:ASIA | 338.75 | 539 | 6 | 30078.75→29894.50 |
| 2026-09-01:NY | 315.50 | 375 | 5 | 29111.75→29117.75 |
| 2026-08-24:NY | 296.00 | 375 | 8 | 29235.50→29131.00 |
| 2026-08-24:ASIA | 271.25 | 540 | 5 | 29140.00→29262.75 |
| 2026-08-25:NY | 270.50 | 375 | 2 | 29295.25→29238.75 |
| 2026-08-30:ASIA | 270.00 | 540 | 3 | 29535.00→29516.75 |
| 2026-08-21:NY | 268.25 | 375 | 1 | 29471.00→29410.75 |
| 2026-08-20:NY | 267.75 | 375 | 2 | 29350.75→29311.25 |
| 2026-09-08:NY | 262.25 | 375 | 5 | 29644.50→29549.00 |
| 2026-08-18:NY | 256.00 | 375 | 1 | 29692.50→29597.25 |
| 2026-08-19:LONDON | 254.75 | 390 | 3 | 29557.00→29713.75 |
| 2026-08-27:NY | 254.25 | 375 | 5 | 29523.75→29633.00 |
| 2026-08-26:ASIA | 253.50 | 540 | 10 | 29500.00→29498.00 |
| 2026-09-15:LONDON | 253.00 | 390 | 2 | 29354.00→29420.75 |


Flags: `2026-09-14:NY` is the Sep→Dec roll day with 102 1m bars (feed gap; the 562-pt range is partly basis). `2026-08-15:NY` and `2026-08-16:NY` have no 1m tape (Saturday/Sunday demo-seed rows) → NOT MEASURED, excluded from every cell.

## 3. (a) — a level AHEAD of price at each read

Two tables per read: the **seated** 12-seat `PlannerInput.Levels` (recorded from 09-09 18:08 CT; n = 21 big-day reads) and the **published** `doc.levels[]` (all 73 big-day reads). "YES" = ≥1 level on the eventual-move side within ±k×DATR (k = 1.0 held on every recorded read — 0/112 seated levels beyond it). "Ran out" = the remaining-session excursion exceeded the farthest ahead level.

### (a) A level AHEAD of price (eventual-move side, within ±k×DATR) — range ≥ 250

- reads measured: 73 (recorded input snapshots: 21); median excursion after read: 155.2 pt; dir(excursion)==dir(close): 67/73
| table | n reads | ahead-in-band YES | fraction | ran out (excursion > farthest ahead) | median overshoot | median table size |
| --- | --- | --- | --- | --- | --- | --- |
| seated 12-seat input (recorded) | 21 | 21 | 1.000 | 12 (0.571) | 221.7 | 12 |
| published doc.levels (all reads) | 73 | 72 | 0.986 | 41 (0.562) | 102.0 | 9 |
- seated per session-day (yes/reads): {'2026-09-10:LONDON': '6/6', '2026-09-14:NY': '8/8', '2026-09-15:LONDON': '2/2', '2026-09-16:NY': '5/5'}
- seated: reads with ≥1 HTF (1h/4h/D) level ahead: 15/21; band probe (seated |distance| > k×DATR): 0 reads
- doc per session-day (yes/reads): {'2026-08-17:ASIA': '2/2', '2026-08-18:NY': '1/1', '2026-08-19:LONDON': '3/3', '2026-08-19:NY': '3/3', '2026-08-20:LONDON': '2/2', '2026-08-20:NY': '2/2', '2026-08-21:NY': '1/1', '2026-08-24:ASIA': '3/3', '2026-08-24:NY': '3/3', '2026-08-25:NY': '2/2', '2026-08-26:ASIA': '7/7', '2026-08-27:NY': '2/2', '2026-08-28:NY': '5/5', '2026-09-01:LONDON': '3/4', '2026-09-01:NY': '2/2', '2026-09-03:NY': '5/5', '2026-09-08:NY': '5/5', '2026-09-10:LONDON': '6/6', '2026-09-14:NY': '8/8', '2026-09-15:LONDON': '2/2', '2026-09-16:NY': '5/5'}
- by trigger (n · doc ahead-yes · doc ran-out · seated n · seated ahead-yes · seated ran-out): scheduled: 30 · 30 · 16 · 4 · 4 · 3; level_event: 40 · 39 · 23 · 16 · 16 · 8; other: 3 · 3 · 2 · 1 · 1 · 1

### (a) A level AHEAD of price (eventual-move side, within ±k×DATR) — range < 250 (control)

- reads measured: 231 (recorded input snapshots: 88); median excursion after read: 92.0 pt; dir(excursion)==dir(close): 181/231
| table | n reads | ahead-in-band YES | fraction | ran out (excursion > farthest ahead) | median overshoot | median table size |
| --- | --- | --- | --- | --- | --- | --- |
| seated 12-seat input (recorded) | 88 | 88 | 1.000 | 19 (0.216) | 49.2 | 12.0 |
| published doc.levels (all reads) | 231 | 230 | 0.996 | 53 (0.229) | 66.0 | 10 |
- seated per session-day (yes/reads): {'2026-09-09:ASIA': '1/1', '2026-09-09:NY': '3/3', '2026-09-10:ASIA': '9/9', '2026-09-11:LONDON': '5/5', '2026-09-11:NY': '6/6', '2026-09-13:ASIA': '15/15', '2026-09-14:ASIA': '8/8', '2026-09-14:LONDON': '4/4', '2026-09-15:ASIA': '9/9', '2026-09-15:NY': '2/2', '2026-09-16:ASIA': '13/13', '2026-09-16:LONDON': '11/11', '2026-09-17:LONDON': '2/2'}
- seated: reads with ≥1 HTF (1h/4h/D) level ahead: 71/88; band probe (seated |distance| > k×DATR): 0 reads
- doc per session-day (yes/reads): {'2026-08-16:ASIA': '5/5', '2026-08-17:LONDON': '1/1', '2026-08-17:NY': '2/2', '2026-08-18:ASIA': '2/2', '2026-08-18:LONDON': '1/1', '2026-08-19:ASIA': '3/3', '2026-08-20:ASIA': '3/3', '2026-08-21:LONDON': '4/4', '2026-08-25:ASIA': '8/8', '2026-08-25:LONDON': '2/2', '2026-08-26:LONDON': '12/12', '2026-08-26:NY': '7/7', '2026-08-27:ASIA': '10/10', '2026-08-27:LONDON': '6/7', '2026-08-28:LONDON': '6/6', '2026-08-31:NY': '3/3', '2026-09-01:ASIA': '4/4', '2026-09-02:ASIA': '12/12', '2026-09-02:LONDON': '4/4', '2026-09-02:NY': '11/11', '2026-09-03:ASIA': '7/7', '2026-09-03:LONDON': '1/1', '2026-09-04:LONDON': '3/3', '2026-09-04:NY': '4/4', '2026-09-06:ASIA': '5/5', '2026-09-07:ASIA': '4/4', '2026-09-08:ASIA': '7/7', '2026-09-08:LONDON': '5/5', '2026-09-09:ASIA': '1/1', '2026-09-09:NY': '3/3', '2026-09-10:ASIA': '9/9', '2026-09-11:LONDON': '5/5', '2026-09-11:NY': '6/6', '2026-09-13:ASIA': '15/15', '2026-09-14:ASIA': '8/8', '2026-09-14:LONDON': '4/4', '2026-09-15:ASIA': '9/9', '2026-09-15:NY': '2/2', '2026-09-16:ASIA': '13/13', '2026-09-16:LONDON': '11/11', '2026-09-17:LONDON': '2/2'}
- by trigger (n · doc ahead-yes · doc ran-out · seated n · seated ahead-yes · seated ran-out): scheduled: 47 · 47 · 16 · 11 · 11 · 3; level_event: 183 · 182 · 37 · 76 · 76 · 16; other: 1 · 1 · 0 · 1 · 1 · 0


Reading: the fraction with SOMETHING ahead is ~1.0 in both pools; the discriminating number is **ran out** — 0.57 (seated) / 0.56 (doc) on big days vs 0.22 / 0.23 on control, with the big-day overshoot (222 / 102 pt) larger than the control's (49 / 66). On big days the scheduled read itself ran out on 16/30 doc reads and the level_event re-reads on 23/40: re-reading did not restore a target ahead. `dir_exc == dir_close` on 67/73 — the direction definition is not doing the work.

## 4. (b) — scenario target chains exhausted before the flat

### (b) Scenarios whose target_chain was EXHAUSTED (price beyond the last target) before the session flat

| pool | scenarios | measured | malformed | exhausted | fraction | first target hit | median min→last | median last-target dist | median #targets |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| range ≥ 250 | 203 | 182 | {'last_target_behind_price': 21} | 88 | 0.484 | 140 | 33 | 69.8 | 3.0 |
| control | 684 | 596 | {'last_target_behind_price': 77} | 247 | 0.414 | 448 | 45 | 69.6 | 3.0 |
- big per session-day (exhausted/measured): {'2026-08-17:ASIA': '2/4', '2026-08-18:NY': '1/3', '2026-08-19:LONDON': '1/9', '2026-08-19:NY': '4/9', '2026-08-20:LONDON': '5/6', '2026-08-20:NY': '5/6', '2026-08-21:NY': '1/3', '2026-08-24:ASIA': '3/10', '2026-08-24:NY': '5/9', '2026-08-25:NY': '1/6', '2026-08-26:ASIA': '21/21', '2026-08-27:NY': '3/7', '2026-08-28:NY': '3/11', '2026-09-01:LONDON': '5/10', '2026-09-01:NY': '6/7', '2026-09-03:NY': '4/12', '2026-09-08:NY': '4/12', '2026-09-10:LONDON': '5/14', '2026-09-14:NY': '1/11', '2026-09-15:LONDON': '2/4', '2026-09-16:NY': '6/8'}


`malformed` = the chain's last target is not ahead of price-at-read in the scenario's own direction (a pullback-entry scenario whose targets sit between entry and current price); excluded, counted. `2026-08-26:ASIA` (21/21 exhausted) is a 253-pt day whose every scenario's chain was overrun.

## 5. (c) — the dispatched candidate rule, and what it points at

HTF universe reconstructed per read with `kernel.DetectHTFLevels` over the DB bars (§1.2). **Parity caveat, per cell:** for recorded reads the reconstruction's in-band zone count matches the planner's recorded `HTFZonesFull` at ratio ≈1.0 on the Sep contract (09-09→09-14: 1.07 / 1.03 / 1.00 / 0.95 / 0.94 by date) and only **0.17–0.44 on the Dec contract** after the 09-14 roll (the `bars` table holds 11 daily / 24 4h / 73 1h Dec bars; the live BarCache had deeper history). Big-day reads on the Dec contract: 15 of 73 (`2026-09-14:NY` 8, `2026-09-15:LONDON` 2, `2026-09-16:NY` 5) — on those the universe is UNDER-detected, so "level existed" is a lower bound and "reached" is unaffected once a level exists. The `s1_4h` / `s1_D` trend rings on the Dec contract are < 60 bars → `n/a` (13 big-day reads).

### (c) Candidate rule — nearest BEYOND-band HTF (1h/4h/D) level in the trend direction — range ≥ 250

- reads: 73; HTF universe source: {'DetectHTFLevels': 73}; median beyond-band HTF levels per read: 192
| trend def | trend counts | trended reads | level existed | fraction | reached before flat | hit rate | median min→hit | median dist (pt) | tf of level | role (recorded only) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | {'range': 17, 'down': 12, 'up': 34, 'n/a': 10} | 46 | 46 | 1.000 | 9 | 0.196 | 1 | 298.5 | {'4h': 29, '1h': 14, '1d': 3} | {'(unassigned)': 46} |
| s1_4h | {'up': 22, 'range': 17, 'down': 21, 'n/a': 13} | 43 | 43 | 1.000 | 7 | 0.163 | 208 | 294.9 | {'4h': 24, '1h': 17, '1d': 2} | {'(unassigned)': 43} |
| s1_D | {'range': 46, 'up': 14, 'n/a': 13} | 14 | 14 | 1.000 | 5 | 0.357 | 1 | 295.9 | {'4h': 10, '1h': 4} | {'(unassigned)': 14} |
| plan_bias | {'up': 19, 'range': 17, 'down': 37} | 56 | 54 | 0.964 | 9 | 0.167 | 38 | 307.2 | {'4h': 37, '1d': 4, '1h': 13} | {'(unassigned)': 54} |

### (c) Candidate rule — nearest BEYOND-band HTF (1h/4h/D) level in the trend direction — control

- reads: 231; HTF universe source: {'DetectHTFLevels': 231}; median beyond-band HTF levels per read: 198
| trend def | trend counts | trended reads | level existed | fraction | reached before flat | hit rate | median min→hit | median dist (pt) | tf of level | role (recorded only) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | {'range': 106, 'down': 49, 'up': 41, 'n/a': 35} | 90 | 90 | 1.000 | 8 | 0.089 | 87 | 320.8 | {'1h': 26, '4h': 60, '1d': 4} | {'(unassigned)': 90} |
| s1_4h | {'up': 29, 'down': 104, 'range': 48, 'n/a': 50} | 133 | 133 | 1.000 | 14 | 0.105 | 1 | 354.8 | {'4h': 102, '1d': 13, '1h': 18} | {'(unassigned)': 133} |
| s1_D | {'range': 133, 'up': 48, 'n/a': 50} | 48 | 48 | 1.000 | 5 | 0.104 | 45 | 308.1 | {'1h': 19, '4h': 29} | {'(unassigned)': 48} |
| plan_bias | {'up': 94, 'range': 38, 'down': 99} | 193 | 177 | 0.917 | 28 | 0.158 | 1 | 327.5 | {'4h': 96, '1h': 61, '1d': 20} | {'(unassigned)': 177} |

### (c′) Alternative rule — FARTHEST in-band HTF (1h/4h/D) level in the trend direction — range ≥ 250

| trend def | trended reads | level existed | fraction | reached | hit rate | median min→hit | median dist | tf | farther than doc table & trend==move | …of which reached |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | 46 | 46 | 1.000 | 12 | 0.261 | 11 | 281.6 | {'4h': 22, '1h': 21, '1d': 3} | 17 | 6 |
| s1_4h | 43 | 43 | 1.000 | 8 | 0.186 | 127 | 281.6 | {'4h': 24, '1h': 17, '1d': 2} | 17 | 5 |
| s1_D | 14 | 14 | 1.000 | 1 | 0.071 | 2 | 282.8 | {'4h': 9, '1h': 4, '1d': 1} | 6 | 0 |
| plan_bias | 56 | 55 | 0.982 | 18 | 0.327 | 18 | 270.5 | {'4h': 32, '1h': 18, '1d': 5} | 24 | 11 |
- ORACLE vs the recorded 12-SEAT table (recorded reads only): 21 reads, table ran out on 12; reads with ≥1 in-band HTF level on the move side beyond the seated farthest: 15; reads where ≥1 such level was REACHED: 12 ['2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@1', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@2', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@3', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@4', '2026-09-10:LONDON:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@5']
- ORACLE (move direction known — lookahead, for sizing the miss only): reads with ≥1 in-band HTF level on the move side beyond the doc table's farthest: 61/73 (0.836); such levels total 881, reached 539; reads where ≥1 was reached: 52; e.g. ["2026-08-17:ASIA@1 ['EQH·1d@30076.75✓', 'EQH·4h@30076.75✓', 'SUPPLY·4h@30147.75✓']", "2026-08-17:ASIA@2 ['EQH·4h@29887.00✓', 'EQH·4h@29985.00✓', 'EQH·4h@30010.00✓']", "2026-08-18:NY:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@1 ['EQL·4h@29518.25✓', 'EQL·4h@29577.25✓', 'EQL·4h@29666.00✓']"]

### (c′) Alternative rule — FARTHEST in-band HTF (1h/4h/D) level in the trend direction — control

| trend def | trended reads | level existed | fraction | reached | hit rate | median min→hit | median dist | tf | farther than doc table & trend==move | …of which reached |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| s1_1h | 90 | 90 | 1.000 | 14 | 0.156 | 7 | 301.0 | {'4h': 57, '1h': 28, '1d': 5} | 47 | 9 |
| s1_4h | 133 | 133 | 1.000 | 40 | 0.301 | 18 | 293.2 | {'4h': 106, '1h': 25, '1d': 2} | 71 | 28 |
| s1_D | 48 | 48 | 1.000 | 4 | 0.083 | 128 | 292.8 | {'1d': 1, '4h': 24, '1h': 23} | 18 | 1 |
| plan_bias | 193 | 189 | 0.979 | 35 | 0.185 | 53 | 284.2 | {'4h': 105, '1h': 79, '1d': 5} | 98 | 22 |
- ORACLE vs the recorded 12-SEAT table (recorded reads only): 88 reads, table ran out on 19; reads with ≥1 in-band HTF level on the move side beyond the seated farthest: 52; reads where ≥1 such level was REACHED: 32 ['2026-09-09:NY:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@1', '2026-09-09:NY:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@3', '2026-09-09:ASIA:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@2', '2026-09-10:ASIA:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@2', '2026-09-10:ASIA:8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265@3']
- ORACLE (move direction known — lookahead, for sizing the miss only): reads with ≥1 in-band HTF level on the move side beyond the doc table's farthest: 193/231 (0.835); such levels total 2732, reached 868; reads where ≥1 was reached: 143; e.g. ["2026-08-16:ASIA@1 ['EQH·4h@30275.00✓', 'EQL·4h@30336.75', 'DEMAND·4h@30273.50✓']", "2026-08-16:ASIA@2 ['EQH·4h@30275.00✓', 'EQL·4h@30336.75', 'DEMAND·4h@30273.50✓']", "2026-08-16:ASIA@3 ['EQH·4h@30275.00✓', 'EQL·4h@30336.75', 'DEMAND·4h@30273.50✓']"]


### (c2) The crux — reads whose level table RAN OUT (excursion beyond the farthest ahead level): would the candidate rule's level have been there and been reached?

| pool · table | ran-out reads | trend def | trend aligned with move | level existed | reached | reached / ran-out reads | median min→hit |
| --- | --- | --- | --- | --- | --- | --- | --- |
| range ≥ 250 · doc | 41 | s1_1h | 10 | 10 | 4 | 0.098 | 39 |
| range ≥ 250 · doc | 41 | s1_4h | 9 | 9 | 5 | 0.122 | 249 |
| range ≥ 250 · doc | 41 | s1_D | 3 | 3 | 2 | 0.049 | 20 |
| range ≥ 250 · doc | 41 | plan_bias | 16 | 15 | 5 | 0.122 | 40 |
| range ≥ 250 · seated | 12 | s1_1h | 0 | 0 | 0 | 0.000 | n/a |
| range ≥ 250 · seated | 12 | s1_4h | 6 | 6 | 5 | 0.417 | 249 |
| range ≥ 250 · seated | 12 | s1_D | 0 | 0 | 0 | 0.000 | n/a |
| range ≥ 250 · seated | 12 | plan_bias | 1 | 0 | 0 | 0.000 | n/a |
| control · doc | 53 | s1_1h | 14 | 14 | 5 | 0.094 | 103 |
| control · doc | 53 | s1_4h | 21 | 21 | 6 | 0.113 | 49 |
| control · doc | 53 | s1_D | 5 | 5 | 2 | 0.038 | 23 |
| control · doc | 53 | plan_bias | 24 | 24 | 11 | 0.208 | 1 |


Reading (c): the nearest beyond-band HTF level always exists — but "beyond band" means beyond ±1.0×DATR ≈ 245–360 pt from price, and the median big-day excursion after a read is 155 pt. That is why it is reached on only 7–9 of 43–56 big-day reads, and rescues 4–5 of the 41 ran-out reads. (c′) — the FARTHEST in-band HTF level in the trend direction — reaches 12/46 (1h) · 8/43 (4h) · 18/55 (plan bias) on big days, and where its distance exceeds the doc table's farthest AND the trend matched the move, 6/17 · 5/17 · 11/24 were reached. The oracle rows size the ceiling: **61/73 big-day reads had an in-band HTF level on the move side that the published table did not carry, and on 52/73 price reached one; against the recorded 12-seat table, 12/21.** A seat rule that keeps the farthest in-band HTF level per side is what these cells argue for; it is NOT tested here as a rule (it needs the seat displacement modelled — which nearer level it evicts and what that costs).

## 6. (d) — replan budget on those days

### (d) Replan budget on those days (class 35: only death_replan/owner_reread SPEND; level_event is FREE)

| pool | session-days | level_event re-reads (total/median/max) | spending reads | recorded counter used (sum) | cap | exhausted (counter ≥ cap) | would exhaust if level_events counted |
| --- | --- | --- | --- | --- | --- | --- | --- |
| range ≥ 250 | 23 | 40/0/7 | 1 | 1 | 4 | 0 [] | 8 ['2026-08-26:ASIA', '2026-08-28:NY', '2026-09-01:LONDON', '2026-09-03:NY', '2026-09-08:NY'] |
| control | 45 | 184/4/15 | 0 | 0 | 4 | 0 [] | 23 ['2026-08-25:ASIA', '2026-08-26:LONDON', '2026-08-26:NY', '2026-08-27:ASIA', '2026-08-27:LONDON'] |
- range ≥ 250 — era split: pre-class-35 (< 2026-09-01): 15 session-days, exhausted-marker rows on 0 [], level_event re-reads 10; post (recorded counter): 8 session-days, counter-exhausted 0, counter used total 1, level_event re-reads 30, exhausted markers 0
- control — era split: pre-class-35 (< 2026-09-01): 19 session-days, exhausted-marker rows on 4 ['2026-08-16:ASIA', '2026-08-25:ASIA', '2026-08-26:LONDON', '2026-08-26:NY'], level_event re-reads 47; post (recorded counter): 26 session-days, counter-exhausted 0, counter used total 0, level_event re-reads 137, exhausted markers 0
- big per session-day: {"2026-08-17:ASIA": {"reads": 6, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-18:NY": {"reads": 1, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-19:LONDON": {"reads": 3, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-19:NY": {"reads": 3, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-20:LONDON": {"reads": 2, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-20:NY": {"reads": 2, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-21:NY": {"reads": 1, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-23:ASIA": {"reads": 4, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-24:ASIA": {"reads": 5, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-24:NY": {"reads": 8, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-25:NY": {"reads": 2, "level_event": 0, "spending": 0, "counter": 0}, "2026-08-26:ASIA": {"reads": 10, "level_event": 4, "spending": 0, "counter": 0}, "2026-08-27:NY": {"reads": 5, "level_event": 2, "spending": 0, "counter": 0}, "2026-08-28:NY": {"reads": 7, "level_event": 4, "spending": 0, "counter": 0}, "2026-08-30:ASIA": {"reads": 3, "level_event": 0, "spending": 0, "counter": 0}, "2026-09-01:LONDON": {"reads": 6, "level_event": 4, "spending": 0, "counter": 0}, "2026-09-01:NY": {"reads": 5, "level_event": 1, "spending": 0, "counter": 0}, "2026-09-03:NY": {"reads": 7, "level_event": 5, "spending": 0, "counter": 0}, "2026-09-08:NY": {"reads": 5, "level_event": 4, "spending": 0, "counter": 0}, "2026-09-10:LONDON": {"reads": 6, "level_event": 5, "spending": 0, "counter": 0}, "2026-09-14:NY": {"reads": 8, "level_event": 7, "spending": 0, "counter": 0}, "2026-09-15:LONDON": {"reads": 2, "level_event": 1, "spending": 0, "counter": 0}, "2026-09-16:NY": {"reads": 5, "level_event": 3, "spending": 1, "counter": 1}}

Reading: under class 35 (deployed `ec6632f9` 2026-09-01 17:24 CT) only `death_replan`/`owner_reread` spend; the recorded counter reached 1 on one big session-day (`2026-09-16:NY`, the sole `death_replan`), and 0 of 23 big days exhausted. Pre-09-01 the exhaustion markers (`replans_exhausted` rows, 5) fall on four sub-250 session-days (`2026-08-16:ASIA`, `2026-08-25:ASIA`, `2026-08-26:LONDON` ×2, `2026-08-26:NY`) — none on a big day. Had level_events counted against the cap of 4, 8/23 big days would have exhausted (max 7 level_event re-reads in `2026-09-08:NY`).

## 7. TRACE — manifest

- **Commit:** `9a37f1910f69` (branch `docs/round-24-q4-target-reach-chief`, sibling of the shared branch; see §9)
- **Scripts:** `docs/superpowers/research/2026-09-17-round-24/extract.sh` · `docs/superpowers/research/2026-09-17-round-24/harness/ (go run ./… -in out-r24/in -out out-r24)` · `docs/superpowers/research/2026-09-17-round-24/analysis.py`
- **Inputs (sha256, `out-r24/in/SHA256SUMS`; the `in/` dumps are gitignored — 260 MB — and reproduce from `extract.sh`):**
  - `bars.csv` `c99a31acb9e9e365cc44d10e7cd76def292eb552aa3c1707e48ca7b59e425058`
  - `bars_htf.csv` `90d61a9356d02d8faaaa078c229774dfaef24afa2e6fe3b7adda6362986c29ca`
  - `plans.json` `505fa846e49321f0630cf7c868bdbb14c409d009a23b22566ac963854685e6b9`
  - `plan_lifecycle_log.json` `f41eec2915e71b6686232e68b7685b3d731e585856c97d4fa030e0365f953a9b`
  - `replan_counters.json` `77b583040de7ff99acb2c00c197910be27568bafcfe7232a4a97baa6196e0468`
  - `strategy_dayplan.json` `76ddf39aa06321ffa83bdc10bc62370649968bbaf656483cca86341169122443`
  - `research_plan.json` `37402825e66820a729e9f88522e4d1894ad4c5fbe11d8189b271fa8ec65a4c3d`
- **Outputs (sha256, committed under `out-r24/`):**
  - `sessions.jsonl` `5106733e04717c0bab8c99b6ebdf7a7dc8e1d4157e7ea172772e3f96ce4c3e05`
  - `reads.jsonl` `9f20558e88104d4151acc3cb08d5a17553fb500571cc51c9306ae91968296492`
  - `scenarios.jsonl` `41294851764043fa8b4875fc8729552b45f498682be460b0e65e86f1602d92a2`
  - `dayplan.jsonl` `c6ab78478139daf7d647dec00c523c0b765822849a955f20645d4dd09aea151d`
- **Per-cell n and first ids:** `out-r24/manifest.json` → `cells`; every per-session-day breakdown above names its ids inline; `out-r24/reads.jsonl` / `scenarios.jsonl` carry every row with its `id` (`plan_id@version`, `#scenario`).
- **Lookahead statement:** price/DATR/seated table/HTF universe/trend use only bars closed before created_at, or the planner's own recorded input snapshot for that read (research store, object=plan/input) · excursion, ran_out, hit, minutes, exhausted use only 1m bars with open >= created_at and < session flat · dir_exc = side of the larger excursion from price-at-read over the remaining session (primary); dir_close = sign(flat close − price) reported as a probe

## 8. Recorded-window join and probes [A]

- 130 `plan/input` snapshots; 114 joined to plan rows (111 model reads + 3 markers) by (TradeDate, Session) with capture ≤ `created_at` and after the previous row; capture→row lag median 354 s, max 1509 s (the planner stream deadline). 16 snapshots have no plan row (fail-closed attempts).
- Seated band: 0/112 seated levels beyond 1.0×DATR → k = 1.0 throughout the recorded window; seated table size median 12.
- HTF parity by (date, contract): see §5.

## 9. Provenance and incidents (this lane)

- 01:54 claim `3a7c319b` (W-ENV, merged as CLASS 137) · 02:11 claim `9ef3cae4` of `docs/round-24-q4-target-reach` from `origin/dev` @ `7059bc74`.
- 02:59 → 08:44 CT: this session was torn down with every other lane in the 08:21 reboot (per CTO nofx-4c); the research-store dump died mid-write. Salvaged 926 complete rows, fetched the tail (146 rows) with a `captured_ms` bound.
- 08:34 the CTO's workflow lane (`~/nofx-q4`, Fable 5.1) checked out the same branch and merged `origin/dev` (`f1ed1338`), then committed its own harness (`9a37f191`, distinct dir `…/2026-09-17-round-24-q4/`). This lane never pushed to the shared branch after the claim; it committed on the sibling `docs/round-24-q4-target-reach-chief` cut from `9a37f191`. Two `git reset --hard HEAD` in `~/nofx-chief` (08:50, 08:59) resynced this worktree's index after the branch ref moved under it; no other lane's file was ever present here. `~/nofx` untouched.
- **Load incident, owned:** the 98 GB research-store scans ran 02:21–02:59 (killed) and 08:45–08:57 CT — the second straddled the NY open, during which the CTO reports the bot stalled under research load. Every later run was `nice -n 19`; the extract script now carries the bound that makes the plan-object query an index range. The live `data.db` was read with `sqlite3 -readonly` as the original dispatch specified; the CTO's later rule (use `db.copy.db`) post-dates these runs.
- No DB write, no deploy, no lock, no key/account printed (plan ids carry the trader's opaque id string only).
