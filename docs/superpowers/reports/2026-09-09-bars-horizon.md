# BARS HORIZON — a count is not a horizon

**Branch** `fix/bars-horizon` · **base** `origin/dev` @ `27e062ea` (rebased twice; see SPEC FRESHNESS)
**Session** bars-horizon-2bdef526/nofx-07[aa8e26] · **worktree** `/home/hoang/nofx-barshorizon`
**Status** pushed, green, **NOT deployed and NOT merged** (A3).
**Checklist numbers** not assigned — two new classes are appended to
`docs/superpowers/AUDIT-CHECKLIST.md` under PENDING NUMBERS (A16: numbers are assigned AT MERGE).

Evidence tiers: **[A]** verified first-hand · **[B]** inferred · **[C]** speculation.

---

## 0 · RETRACTION — two premises this wave was launched on were WRONG

Both were stated to the owner as fact. Both are corrected here rather than
quietly dropped.

### R1 — `scope_bars` ALREADY RECORDS THE SERVED COUNT. It never recorded the request.

`trader/auto_trader_planner.go` has always written `ScopeBars: len(scope.Bars)`
— the post-truncation, SERVED length. **[A]** Any design premised on
"`scope_bars` reports the request" is wrong, and redefining the column would
have silently rewritten the meaning of 67 correct rows.

```
sqlite> select count(*), min(id), max(id), group_concat(distinct scope_bars),
                group_concat(distinct scope_intv) from planner_read_facts;
67|1|67|2000|1m
```

**The void scope has NEVER been short: 67 of 67 rows (ids 1–67) record
`scope_bars=2000` against a 2000 ask.** **[A]** This is the load-bearing
counter-evidence for the whole wave: the bridge's silence is real, but at the
~29 sites that ask 2000 it has cost nothing.

`scope_bars` is therefore left **exactly as it was**, with a comment in
`buildReadFactRow` saying why. What was missing is that **a COUNT cannot express
a SPAN or a HOLE** — see §C5 and D4.

### R2 — THE ATR WAS NOT INFLATED, AND ARM 133 IS NOT OVER-SIZED.

The claim rested on comparing today's **NY** read against today's **ASIA** read.
Those are two volatility regimes; the comparison could only produce a difference,
and the difference was read as a defect.

Held to the same session — `planner_read_facts`, verified first-hand **[A]**:

| session | id | atr5m | stop_floor_pts | created (CT) |
|---|---|---|---|---|
| 09-08 NY | 54 | 36.7439730211274 | 55.115959531691 | 2026-09-08 10:02:43 |
| 09-08 NY | 55 | 32.2843805014897 | 48.4265707522345 | 2026-09-08 10:45:01 |
| 09-08 NY | 56 | 28.0296201145494 | 42.044430171824 | 2026-09-08 12:15:00 |
| 09-09 NY | 65 | 20.514183795639 | 30.7712756934586 | 2026-09-09 13:08:08 |
| 09-09 NY | 66 | 19.7631429163709 | 29.6447143745564 | 2026-09-09 13:18:13 |
| 09-09 NY | 67 | 16.6611226001067 | 24.99168390016 | 2026-09-09 13:55:06 |

**Today's NY ATR is roughly HALF yesterday's NY ATR.** Arm 133 (composed
13:18:13 CT, the same second as row 66; long limit 29424.5, stop 29393.5 = 31.0
pts) was **not** composed on an inflated floor. Nothing in this wave touches arm
133 or any order path (owner ruling).

### The method error, stated as its own line

> **A CROSS-SESSION ATR COMPARISON IS A REGIME COMPARISON, NOT A MEASUREMENT.**

A claim about a level, a floor, a size or a threshold is a measurement only when
both sides come from the same session and the same calendar class, and the
report quotes the row ids on both sides (A21). Filed as a new bug class.

---

## SPEC FRESHNESS (class 73)

`git log -1` on origin/dev for every file this wave built against:

```
kernel/planner_prompt.go        486cec46 2026-09-09 14:15:47 -0500  docs(W5): the system describes itself truthfully
kernel/weekly_bias.go           23243670 2026-08-30 09:40:45 -0500  fix(weekly-bias): calendar-anchored week governing Monday
kernel/session_calendar.go      5457ac5a 2026-09-07 19:39:02 -0500  fix(session-calendar): five bare "15:04" layouts
provider/ninjatrader/bar_cache.go 2b22eed2 2026-08-27 16:45:24 -0500 fix(bar-truth): open-stamp the historical persistence path
store/bar_history.go            5821ae55 2026-09-02 12:25:16 -0500  feat(bars): persist EVERY cached TF + per-TF retention
trader/auto_trader_planner.go   0babd090 2026-09-08 16:22:32 -0500  feat(research): link attempts, permissions, broker receipts
trader/auto_trader_weekly.go    21b3e75e 2026-09-03 22:50:35 -0500  fix(class 60/72): clock seams
kernel/regime.go                7b19e753 2026-09-03 23:00:08 -0500  style(kernel): gofmt
kernel/regime_baseline.go       d85f6f13 2026-08-15 12:55:09 -0500  feat(dayplan): W11b
trader/ninjatrader/bar_persist_wire.go 56904ec1 2026-09-02 21:29:52 -0500 feat(no-chase): riders R1+R2
kernel/void_scope.go            60f214d9 2026-09-02 22:38:48 -0500  fix(void-parity): the scope VALUE is the session day
store/planner_read_facts.go     4659874a 2026-09-02 22:38:48 -0500  fix(void-parity): one resolver for the void scope
docs/superpowers/SYSTEM-MAP.md  27e062ea 2026-09-09 14:55:48 -0500  merge dev (W5) + correct three sentences
docs/superpowers/AUDIT-CHECKLIST.md 27e062ea 2026-09-09 14:55:48 -0500  (same commit)
```

**Dev moved under this wave twice and it was rebased both times.** The first
base (`05125bd6`, a `fix/brand-visible` merge) was replaced by `8dfe6bc1`; then
`SYSTEM-MAP.md` and `AUDIT-CHECKLIST.md` — two files this wave edits — moved at
14:55:48 in `27e062ea`, **after** that base, so the branch was rebased again
onto `27e062ea` before the report was written. Without the `git log -1` check
this wave would have written its SYSTEM-MAP section against superseded text.

---

## SECTION C — what actually justifies this wave (all measured)

### C1 · THE HEADING LIES, AND THE MODEL IS TOLD TO TRUST IT **[A]**

`kernel/planner_prompt.go:386-397` (pre-wave):

```go
render := func(title string, bars []market.Kline, n int) {
    if len(bars) > n { bars = bars[len(bars)-n:] }
    fmt.Fprintf(&b, "### %s\n", title)
    FormatCandleTable(&b, KlineBars(bars), true)
}
render("15m (last 12)", AggregateBars(bars1m, 15*60*1000), 12)
render("1h (last 12)",  AggregateBars(bars1m, 60*60*1000), 12)
render("4h (last 8)",   AggregateBars(bars1m, 240*60*1000), 8)
render("daily session candles (last 8)", DailySessionBars(bars1m), 8)
```

Truncate-only-when-long — the bridge's exact semantics one layer up — with the
count baked into the title as a literal.

**Measured over the stored prompts** (`planner_rejected_prompts`, n=54 carrying
a Candles block, ids 70–142), reproduced first-hand:

```
n prompts with a Candles block: 54
  '15m (last 12)':                    12 rows in 54
  '1h (last 12)':                     12 rows in 54
  '4h (last 8)':                       8 rows in 54
  'daily session candles (last 8)':    2 rows in 19 · 3 rows in 35
      2 rows -> ids [70, 71, 72, 73, 74, 95]... (n=19)
      3 rows -> ids [75, 77, 78, 80, 81, 83]... (n=35)
```

**Never 8, in 0 of 54.**

Worse, the OLDEST row of each daily table is a **partial session candle**
presented as a whole one. `DailySessionBars` (`kernel/weekly_bias.go:240-266`)
buckets by `CMESessionDayKey` and takes the first bar it SEES as the session
Open, with no completeness check. Live instance, prompt id 142 **[A]**:

```
### daily session candles (last 8)
Time(CT)       Open      High      Low       Close     Volume
09-07 02:39    29641.5000 29664.5000 29534.0000 29604.2500 277429.00
09-07 17:00    29610.2500 29764.7500 29424.5000 29525.5000 2222809.00
09-08 17:13    29525.0000 29625.5000 29489.7500 29590.7500 235420.00     <- current
```

The first row is stamped `09-07 02:39` for a session that opened at
**09-06 17:00 CT**; the last is stamped `09-08 17:13` for a session that opened
at 17:00. Both are indistinguishable from a true session candle.

And `kernel/planner_prompt.go:428` instructs the model: *"Ground truth for
structure; ranked levels and tags are summaries. On conflict, trust the candles
and say so in the scenario rationale."* The prompt disclosed its tape depth
**nowhere**.

### C2 · NAMES THAT PROMISE MORE HISTORY THAN EXISTS **[A]**

- `RVBaselineFrom5m(min5Long, 20, 5)` (`kernel/regime_baseline.go`) filled the
  struct field **`RVBaseline20d`** from about 7 complete session-days. It is
  honest internally — it drops incomplete days and returns `(0,false)` below
  `minDays=5` — but nothing carried the real count to the reader, and the
  regime line rendered `RV=…%-of-normal`: a baseline with no stated window.
- `kernel/weekly_prompt.go:46` `CompletedWeekCandles(bars1m, now, 12)` — twelve
  weeks from a ring that holds at most 41.7 h.
- `kernel/weekly_prompt.go:48` `LastNWOGs(bars1m, now, 5)` — five weekend gaps
  from the same tape.

### C3 · THREE CALL SITES WERE UNREACHABLE BY CONSTRUCTION **[A]**

Ring ceiling `DefaultBarCacheMaxBars = 2500` (`provider/ninjatrader/bar_cache.go:24`):

| site | ask | implied window | ring ceiling |
|---|---|---|---|
| `trader/auto_trader_planner.go` | 1m × 12000 | 200.0 h | 41.7 h |
| `trader/auto_trader_weekly.go` | 1m × 12000 | 200.0 h | 41.7 h |
| `trader/auto_trader_planner.go` | 5m × 3000 | 250.0 h | 208.3 h |

**No market condition can satisfy any of them.** These are not "short reads on a
quiet day"; they are asks the cache is structurally incapable of serving.

### C4 · THE STORE HELD THE DEPTH AND WAS WIRED TO NOTHING **[A]**

Measured 2026-09-09:

```
symbol  tf   n      oldest_ct            newest_ct
MNQ     1m   20043  2026-08-19 10:00:00  2026-09-09 14:35:00     (21 days)
MNQ     5m    3164  2026-08-24 11:15:00  2026-09-09 14:30:00     (16 days)
```

Retention per TF (`store/bar_history.go`): 1m 90d · 3m/5m 180d · 15m/30m 365d ·
1h and coarser keep-forever. **Nothing has ever been pruned.**
`market.FuturesBarsProvider` has ONE production assignment
(`trader/ninjatrader/bars_market_bridge.go`) and it reads
`server.BarCache().Get(...)` only. `SeedHistorical` MERGES within a process, so
the ring climbs 2000 → 2500 across a session — and every Go restart drops it
back to the 2000-bar seed (`defaultAutoBarsBack`).

### C5 · CORRECTED, PLAINLY

`scope_bars` already records SERVED. The void scope has never been short:
**67/67 at 2000**. The bridge's silence is real but has cost nothing at the ~29
sites that ask 2000. The defect was never the COUNT — it was **CONTINUITY and
SPAN**, and no field could express either.

The live proof, verified three independent ways **[A]**:

- **The hole.** A 14:49 CT snapshot of `bars` shows one missing run for
  session-day 09-08: `09-09 01:29 → 09-09 13:04 CT`, **696 minutes**
  (613 of 1,309 minutes held).
- **The span.** The newest 2000 MNQ 1m bars at or before 13:18:13 CT span
  `09-07 10:23 → 09-09 13:18` = **3,055 minutes** (3,056 intervals), oldest bar
  **50.92 h** old.
- **The gap count, from the production function.**
  `kernel.OpenIntervalsBetween(09-07 10:23, 09-09 13:19, 60000, 200000)` returns
  **2,696** open 1m intervals. 2,696 − 2,000 served = **696 gaps** — to the
  minute, matching the observed hole.

So a served-count check would have been **silent** on the read that mattered.
`planner_read_facts` id 66 recorded `scope_bars=2000` and was **right**.

> **Dated, because it will not reproduce tomorrow:** between 14:49 and ~15:10 CT
> that hole was BACKFILLED. The live table now holds **1,328 contiguous minutes**
> for the same session with `missing=0`, so an identical read today records
> `gaps=0`. The snapshot at `scratchpad/live_copy.db` is the artefact the 696
> figure rests on. This is exactly why the gap count belongs ON THE ROW: the
> tape heals, the record does not.

---

## WHAT WAS BUILT

Four commits, each revertible on its own **except** that D3 uses
`store.BarHistoryStore.LastNBars` introduced by D2, so the revert order is
**D4 · D3 · D2 · D1**.

### D1 — every history-bearing table declares HELD-vs-CLAIMED, and marks a partial as partial

`kernel/candle_disclosure.go` (new) · `kernel/planner_prompt.go` ·
`kernel/engine_prompt.go` · guide · SYSTEM-MAP.

- `BuildPlannerCandleTables(bars1m)` → **`BuildPlannerCandleTablesAt(bars1m, asked1m, now)`** (A28: the clock comes from the caller).
- **Heading:** `### <label> — HELD h of a requested rows · <clauses>`. When
  short: `SHORT BY n — the 1m tape does not reach back far enough; the n older
  rows are ABSENT from this prompt, not flat, and MUST NOT be inferred`. When
  nothing is marked: `all held rows COMPLETE`. The complete case is stated
  deliberately — a disclosure that only appears on failure is one the reader
  learns to skim past.
- **TAPE line**, above all four tables: bars HELD of requested, oldest bar +
  age (minute resolution, rounded DOWN so the tape is never claimed older than
  it is), newest bar, and open 1m intervals missing INSIDE the span.
- **Row markers**, measured against the SAME `kernel/session_calendar.json` the
  trading gate reads (shortened and closed days included):
  - `⚠PARTIAL` — front-truncated rows name the first bar HELD *and* the window
    open; **interior holes** say the minutes are missing INSIDE the window.
    These are different facts and the first live render proved it (a row read
    *"its Open is the first bar HELD (01:15 CT), not the window open (01:15 CT)"*
    — the same clock twice, a false reason; fixed and pinned as D1-G).
  - `⏳FORMING` — the window has not closed.
  - `❓COVERAGE-UNKNOWN — <year> is not covered by kernel/session_calendar.json`
    — never a guess, never a plausible number.
- **A24 absolute:** no bar is interpolated, carried forward or synthesised, and
  no rendered O/H/L/C/V is changed. **A10:** discloses only; gates nothing.
- `FormatCandleTable` keeps its single-formatter role —
  `FormatCandleTableNoted` is the same function with a per-row note, and
  `FormatCandleTable` now calls it.
- The pre-existing `TestPlannerCandleTablesRenderAndTokenBudget` **asserted the
  lie** (`"### 15m (last 12)"`) and passed for the life of the defect, because
  it checked the literal in the title rather than the rows underneath. Updated
  deliberately, with that noted in the test.

**Not fixed, same class, reported:** `kernel/weekly_prompt.go` renders
`## Weekly candles (12 completed weeks, oldest → latest)` over as few as one
row. It is the identical defect. It was left alone because `SectionsText` feeds
`WeeklyFactsHash`, so changing the heading changes a stored hash — a separate
decision, not a silent side effect of this wave.

### D2 — callers that ask above the ceiling either read the store, or are corrected and say which

`store/bar_history.go` (`LastNBars`) · `trader/bars_store_depth.go` (new) ·
`trader/auto_trader_planner.go` · `trader/auto_trader_weekly.go` ·
`kernel/regime.go` · `kernel/regime_baseline.go` · guide · SYSTEM-MAP.

| site | choice | reason, recorded at the call site |
|---|---|---|
| `auto_trader_planner.go` 1m × 12000 | **(a) read the store** | the tape feeds the "8 daily session candles" table, which needs ~11,040 open 1m intervals; the ring gives 41.7 h, the store 21 days |
| `auto_trader_weekly.go` 1m × 12000 | **(a) read the store** | `ComputeWeeklyFacts` asks for 12 COMPLETED WEEKS + 5 NWOGs from a ring holding under two days. The store reaches 21 days — still short of 12 weeks, which `ThinHistory` already stamps honestly. Nothing it computes is changed. |
| `auto_trader_planner.go` 5m × 3000 | **(b) correct the ask, stop promising 20d** | two measured reasons, below |

**Why the 5m site refused (a).** ① The store's 5m reaches ~16 days (3,164 MNQ
rows back to 2026-08-24), so reading it could not honour "20d" either — it
would swap one unmet promise for another while changing a live regime input.
② The stored 5m rows are NT8 aggregates this repo has **already judged
inconsistent with their own 1m constituents**: `store/bar_history.go` `Migrate`
step 4 deletes every `tf != '1m'` row for exactly that reason, and the per-TF
persistence restored on 2026-09-02 did not re-establish their agreement.
Feeding them into a LIVE regime input is not a depth improvement; it is an
unverified substitution.

So: ask `3000 → rvBaseline5mBarsAsk = 2500` (the deepest the ring can serve);
`RVBaseline20d` → `RVBaseline` + **`RVBaselineDays`** carrying the count it was
ACTUALLY fed; the regime line stops saying `RV=103%-of-normal` and now says
`RV=103%-of-baseline(7 complete session-days)`, or `(window UNKNOWN)` when the
count was not reported. **The computed value is unchanged** — `RVBaselineFrom5m`
is now a thin wrapper over `RVBaselineFrom5mDays` so the two cannot disagree.

`barsWithStoreDepthFrom`'s four rules, each pinned:

1. **An EMPTY ring is NEVER substituted.** An empty ring means the feed is down
   or the seed has not landed; handing a planner 12,000 stored bars would let it
   write a plan on a dead tape — the exact failure "no NT8 → no decisions"
   exists to prevent. The store DEEPENS a live tape; it never stands in for one.
2. A ring that already serves the ask is not second-guessed — **no store read at
   all**, so the ~29 healthy call sites pay nothing.
3. Only bars **strictly older** than the ring's oldest are taken, so a live or
   forming bar is never replaced by a stored one.
4. A failed store read WARNs and degrades to the ring (A10).

### D3 — the ring rehydrates from the store on boot (owner-authorised)

`provider/ninjatrader/bar_cache.go` (`RehydrateOlder`) ·
`trader/ninjatrader/bar_persist_wire.go` (`rehydrateRingFromStore`).

Fired once the AddOn replay lands — **alongside** the AddOn's seed. It is
deliberately **not** `SeedHistorical`, on two counts:

- **A cold key is a NO-OP.** `SeedHistorical` seeds an empty key; this refuses
  to. This is the D3 safety argument in full: before the change, a restart with
  NT8 down left the cache empty and the bot idle. Filling a cold cache from the
  store would have made a dead feed look alive to every reader downstream. By
  restricting rehydration to keys that ALREADY hold live bars, the change can
  only ever restore depth the ring just lost — it can never manufacture a tape.
- **`existing` wins the overlap.** `SeedHistorical` lets `incoming` win because
  incoming is the freshest wire OHLCV; here `incoming` is the STORE, which is by
  definition not fresher than the live ring.

Merges via the same `mergeBarsByTime` discipline, bounded by the ring's own
`maxBars` (trimming the OLDEST, never the live tail), refuses NT8 empty-minute
placeholder bars at this door as at every other, WARNs-and-continues on a store
failure (boot is never blocked), and logs per `(symbol, tf)` with resolved
counts (A9/A11). It touches neither the AddOn, the subscription, nor the
backfill.

### D4 — the served-scope observability, kept AS BUILT, reconciled with R1

`4b2edea3` is kept unchanged. **Nothing in it assumed `scope_bars` was the
request** — its warn arms are SHORT (served < requested), HOLED (open intervals
missing inside the span) and EMPTY, and its own commit message already said the
count was never the defect. So the reconciliation with R1 is a correction of the
*record*, not of the code:

- `store/planner_read_facts.go` — four additive columns
  (`scope_requested_bars`, `scope_span_ms`, `scope_oldest_age_ms`,
  `scope_gap_count`) plus `read_horizons` (JSON `kernel.BarHorizon[]`), with a
  comment stating R1 as NOT REPRODUCED and why the column is not redefined.
- `kernel/void_scope.go` — `VoidScope.Horizon`, computed with the SAME `now` the
  caller passed (A28), so the recorded age is exact.
- `trader/auto_trader_planner.go` — `buildReadFactRow` extracted from
  `persistReadFacts` (class 86) so the pins drive the PRODUCTION builder rather
  than a copy, and pure (`now` comes in).
- **UNKNOWN is not ZERO.** `read_horizons == ""` marks a row whose horizon was
  never computed; its four numerics are UNKNOWN. The 67 pre-wave rows read that
  way. Query with `where read_horizons != ''` and state the excluded COUNT
  (corrected-column law, class 40).

---

## D1 SAMPLE RENDER — live data

Rendered from the live `bars` table (12,000 MNQ 1m bars, snapshot 2026-09-09
14:49:50 CDT — the snapshot that still contains the 696-minute hole).

```
TAPE: 12000 1m bars HELD of 12000 requested · oldest 08-27 06:13 CT (320h36m ago) · newest 09-09 14:48 CT · gaps 696 open 1m intervals missing INSIDE that span. Every table below is aggregated from THIS tape and can be no deeper than it.
ROW MARKERS: ⚠PARTIAL = only part of the row's window is held, so its O/H/L/C come from the bars HELD and not from the whole window · ⏳FORMING = the window has not closed yet · ❓COVERAGE-UNKNOWN = the session calendar cannot classify this window. A missing bar is NEVER interpolated, carried forward or synthesised — it is simply absent, and the row says so.

### 1h — HELD 12 of 12 requested rows · 3 of 12 held rows are MARKED on the row itself (⚠PARTIAL / ⏳FORMING / ❓COVERAGE-UNKNOWN)
Time(CT)       Open      High      Low       Close     Volume
09-08 15:00    29542.5000 29550.7500 29503.2500 29525.5000 38869.00
09-08 17:00    29528.0000 29550.2500 29503.2500 29544.7500 24680.00
09-08 18:00    29544.5000 29546.2500 29489.7500 29500.2500 30045.00
09-08 19:00    29500.5000 29562.7500 29499.0000 29560.0000 74022.00
09-08 20:00    29560.2500 29608.5000 29555.7500 29593.0000 55518.00
09-08 21:00    29593.5000 29625.5000 29583.5000 29611.7500 36383.00
09-08 22:00    29611.2500 29614.2500 29560.2500 29562.2500 34456.00
09-08 23:00    29561.7500 29620.2500 29549.2500 29614.2500 38407.00
09-09 00:00    29614.5000 29621.5000 29562.0000 29621.0000 37299.00
09-09 01:00    29621.0000 29634.0000 29587.5000 29589.0000 21241.00      ⚠PARTIAL — opens at the window open (01:00 CT) but holds only 29 of the 60 open 1m intervals; 31 are missing INSIDE this window
09-09 13:00    29465.5000 29474.7500 29436.5000 29447.7500 81423.00      ⚠PARTIAL — holds 55 of the 60 open 1m intervals in this window; its Open is the first bar HELD (13:05 CT), not the window open (13:00 CT)
09-09 14:00    29447.5000 29471.2500 29430.7500 29447.7500 71769.00      <- current  ⏳FORMING ⚠PARTIAL — the window is still open AND holds only 49 of the 50 open 1m intervals elapsed so far

### daily session candles — HELD 8 of 8 requested rows · 1 of 8 held rows are MARKED on the row itself (⚠PARTIAL / ⏳FORMING / ❓COVERAGE-UNKNOWN)
Time(CT)       Open      High      Low       Close     Volume
08-30 17:00    29535.0000 29543.5000 29273.5000 29521.5000 2340435.00
08-31 17:00    29518.5000 29571.0000 29001.7500 29139.0000 2550843.00
09-01 17:00    29137.2500 29212.5000 28927.2500 29143.0000 2177636.00
09-02 17:00    29175.7500 29585.0000 29075.0000 29502.7500 2335268.00
09-03 17:00    29500.0000 29720.0000 29468.2500 29524.7500 1961640.00
09-06 17:00    29534.5000 29683.7500 29530.2500 29604.2500 552400.00
09-07 17:00    29610.2500 29764.7500 29424.5000 29525.5000 2222809.00
09-08 17:00    29528.0000 29634.0000 29430.7500 29447.7500 505243.00     <- current  ⏳FORMING ⚠PARTIAL — the window is still open AND holds only 613 of the 1310 open 1m intervals elapsed so far
```

Two things to read off it. **First:** the daily table renders **8 of 8** — the
D2 store read is what makes that possible; from the ring alone this table
rendered 2 or 3 rows in 54 of 54 stored prompts. **Second:** the interior/front
distinction is doing real work — `09-09 01:00` opens on time and is holed
inside, `09-09 13:00` is front-truncated at 13:05. Those are the two ends of the
same 696-minute outage, and before this wave both rendered as ordinary candles.

---

## PINS — RED → GREEN → MUTATION

Every pin was quoted RED on a real assertion before it was made green (new
functions were stubbed to the PRE-WAVE behaviour first, so no pin's RED is a
mere compile error), then mutated.

### D1 · `kernel/planner_candle_disclosure_test.go`

**RED** (`TestCandleHeadingsStateHeldVsClaimed`), with the fixture reproducing
the live shape exactly:

```
--- FAIL: TestCandleHeadingsStateHeldVsClaimed (0.00s)
    planner_candle_disclosure_test.go:91: heading missing "### 15m — HELD 12 of 12 requested rows"
        --- rendered ---
        ### 15m (last 12)
        ...
        ### daily session candles (last 8)
        Time(CT)       Open      High      Low       Close     Volume
        09-08 02:39    29500.0000 29511.7500 29498.0000 29500.5000 9612.00
        09-08 17:00    29500.2500 29511.7500 29498.0000 29505.2500 14628.00      <- current
```

**GREEN:** `ok  nofx/kernel  0.028s`

**MUTATIONS** (each: the exact line changed, then the failure text):

| # | line changed | result |
|---|---|---|
| 1 | `RowCoverage.Partial()` body → `return false` | **FAIL** `the front-truncated session ROW is NOT marked partial` · `the partial row does not disclose that its Open is the first bar HELD` |
| 2 | `tableHeading`: `label, held, asked,` → `label, asked, asked,` | **FAIL** `heading missing "### daily session candles — HELD 2 of 8 requested rows"` |
| 3 | `if st := SessionStateAt(...); st.UncoveredFallback {` → `; false && st.UncoveredFallback {` | **FAIL** `must render "❓COVERAGE-UNKNOWN — 2031 is not covered by kernel/session_calendar.json" on the ROW` · `an UNKNOWN row also printed a computed interval count` |
| 4 | `RowCoverage.Marked()` body → `return false` | **FAIL** `### 15m heading claims "all held rows COMPLETE" over 1 MARKED row(s)` · `heading does not count its 1 marked rows` |
| 5 | `case c.Partial() && c.FirstHeldMs > c.WindowOpenMs:` → `case c.Partial() && false:` | **FAIL** `the front-truncated session ROW is NOT marked partial` |
| 6 | `age = msDurTxt(h.OldestAgeMs / 60_000 * 60_000)` → `msDurTxt(h.OldestAgeMs)` | **FAIL** `tape age carries sub-second noise: "… (34h39m13.387s ago) …"` |

**Two pins existed only because mutation found them missing** — both would have
shipped otherwise:

- **Mutation 3 passed on the first version** of the pin. Searching the SECTION
  for `"COVERAGE-UNKNOWN"` matched the **heading's own legend clause**
  (`(⚠PARTIAL / ⏳FORMING / ❓COVERAGE-UNKNOWN)`), so the pin could not fail. A
  `d1Rows()` extractor was added that returns only timestamped candle rows, and
  the same weakness was fixed in the PARTIAL/FORMING assertions of D1-B and
  D1-D, which had it too. **Class 89, exactly.**
- **Mutation 4 passed** with every other pin green: forcing `Marked()` false let
  the heading claim `all held rows COMPLETE` over a row carrying `⚠PARTIAL`. A
  heading contradicting its own rows is the defect this wave exists to remove,
  one layer up. **PIN D1-F** now asserts the heading and the rows agree in both
  directions.
- **Mutation 6 passed** on the first version because the fixture clock had no
  sub-second part, so the pin could not fail; it now states its own clock with
  `387_000_000` ns.

### D2 · `trader/bars_store_depth_test.go`

**RED:**
```
--- FAIL: TestStoreDepthPrependsOlderAndNeverReplacesRing (0.00s)
    bars_store_depth_test.go:57: served 100 bars, want 400 (300 from the store + 100 from the ring)
--- FAIL: TestRVBaselineCarriesItsActualDayCount (0.00s)
    bars_store_depth_test.go:146: complete-day count = 0, want 1..20 — an uncomputed count is UNKNOWN, never a plausible number
--- FAIL: TestD2WiredAtTheThreeCallSites (0.09s)
    bars_store_depth_test.go:177: barsWithStoreDepth(: 0 production call sites (A29)
    bars_store_depth_test.go:188: an ask above the ring ceiling survives: "1m", 12000 in [trader/auto_trader_weekly.go]
    bars_store_depth_test.go:188: an ask above the ring ceiling survives: "5m", 3000 in [trader/auto_trader_planner.go]
```

**GREEN:** `ok  nofx/trader  0.117s`

| # | line changed | result |
|---|---|---|
| 7 | `if k.OpenTime < oldestRing {` → `<= oldestRing` | **FAIL** `result is not strictly ascending at 300 (1788971880000 <= 1788971880000)` |
| 8 | drop `len(ring) == 0 \|\|` from the early return | **FAIL** `TestEmptyRingIsNeverSubstitutedByTheStore` |
| 9 | `return …, len(perDay), true` → `…, maxDays, true` | **FAIL** `complete-day count = 20, want exactly 7` |
| 10 | `} else if r.RVBaselineDays > 0 {` → `} else if false && …` | **FAIL** `the regime line does not name the baseline's real window (7 days): REGIME: … · RV=103%-of-baseline(window UNKNOWN) · VIX=n/a` |

**Mutation 9 passed on the first version** of the pin, which asserted the day
count in a `1..20` RANGE: returning `maxDays` — the ASK — sat inside it. The
wave's entire point is that the fed count and the asked count DIFFER, so a pin
tolerating the ask cannot see the defect. It now asserts **exactly 7**, and
separately that the reported count is not the ask. **Class 89.**

### D3 · `provider/ninjatrader/bar_rehydrate_test.go`

**RED:**
```
--- FAIL: TestRehydrateOlderExtendsBackwardsOnly (0.00s)
    bar_rehydrate_test.go:36: rehydrated 0 bars, want 300 (only those strictly older than the ring's oldest)
--- FAIL: TestRehydrateOlderIsCappedAtMaxBars (0.00s)
    bar_rehydrate_test.go:81: ring holds 100 bars, want the 500 cap
--- FAIL: TestRehydrateOlderRefusesPlaceholderBars (0.00s)
    bar_rehydrate_test.go:102: rehydrated 0 bars, want 4 — the placeholder must be dropped, not stored
```

**GREEN:** `ok  nofx/provider/ninjatrader  0.006s`

| # | line changed | result |
|---|---|---|
| 11 | `if len(existing) == 0 { return 0 }` → `if false && …` | **FAIL** `TestRehydrateOlderIsANoOpOnAColdKey` |
| 12b | `mergeBarsByTime(older, existing)` → `mergeBarsByTime(existing, older)` | **PASS — EQUIVALENT MUTANT** |
| 12c | remove the `b.T < oldest` filter | **PASS — EQUIVALENT MUTANT** |
| 12d | **both** 12b and 12c together | **FAIL** `live bar 0 was replaced by a stored one: {T:940000 … C:-1} want {… C:7}` |
| 13 | `merged = merged[len(merged)-c.maxBars:]` → `merged[:c.maxBars]` | **FAIL** `the cap trimmed a LIVE bar at 0` |
| 14 | remove `dropPlaceholderBars` from `RehydrateOlder` | **FAIL** `rehydrated 5 bars, want 4 — the placeholder must be dropped, not stored` |

**12b and 12c are genuine equivalent mutants and are reported as such, not as
green.** The live-bar guard is DOUBLE — an older-only filter AND `existing`
passed as `mergeBarsByTime`'s `incoming` (which wins every tie). Removing
exactly one leaves the other protecting the live bar, so no test can catch it;
removing both is caught (12d). This is recorded in the test file itself so a
later reader does not "fix" a correct test.

A fixture bug was also caught here: the first version compared against the raw
input slice, but `SeedHistorical` applies the wire's close-stamp → open-stamp
conversion, so the ring's bar times are not the fixture's. The pin now reads the
tape back — which additionally proves store rows (already open-stamped by
`bar_persist_wire`) and ring rows agree.

### D4 · `trader/read_facts_horizon_test.go` · `store/planner_read_facts_test.go`

**RED:**
```
--- FAIL: TestReadFactRowCarriesTheHorizon (0.00s)
    read_facts_horizon_test.go:62: scope_requested_bars=0, want 2000 — the REQUEST is a new column, never a redefinition of scope_bars
--- FAIL: TestAHolyTapeAndAContiguousTapeDifferOnTheRecord (0.00s)
    read_facts_horizon_test.go:107: a contiguous tape and a holed tape are STILL indistinguishable on the record (span 0 gaps 0 both)
--- FAIL: TestReadFactRowBuilderIsWired (0.04s)
    read_facts_horizon_test.go:127: buildReadFactRow(: 0 production call sites (A29)
```

**GREEN:** `ok  nofx/trader  0.050s` · `ok  nofx/store  0.161s`

| # | line changed | result |
|---|---|---|
| 15 | `ScopeBars: len(scope.Bars)` → `ScopeBars: scope.BarCount` (shipping the refuted premise) | **FAIL** `scope_bars=2000, want the SERVED 1200 — it is len(scope.Bars) and always was (R1)` |
| 16 | `row.ScopeGapCount = h.GapCount` → `= 0` | **FAIL** `the holed tape records gaps=0 — a hole must be counted, not implied` |
| 17 | `HorizonRecorded()` → `return true` | **FAIL** `legacy row reports HorizonRecorded()=true, want false — its scope_gap_count=0 is UNKNOWN, not zero` |
| 18 | `if h.Interval == "" {` → `if false && h.Interval == "" {` | **FAIL** `read_horizons="[{"asked":0,…,"gaps":0}]" on a scope that computed no horizon — it must stay "" so the row reads UNKNOWN` |

**Mutations 15 and 18 both passed first time** and both are real gaps that would
have shipped:

- **15**: every live row and the first fixture had `served == requested == 2000`,
  so swapping the two — *shipping the very premise R1 refuted* — was invisible.
  **PIN D4-D** now uses a SHORT fixture (1200 served of 2000 requested) and
  asserts the two columns differ. **Class 89.**
- **18**: no pin covered a scope carrying no horizon. Removing the guard let an
  uncomputed horizon write four zeros AND a fully-zero `read_horizons` entry —
  a row that measured nothing claiming "span 0, age 0, no gaps". **PIN D4-E**
  pins the UNKNOWN. **A24, the plausible zero.**

---

## SCOPE

`git diff --name-only origin/dev...HEAD` — 29 files, all inside the wave's
footprint. No AddOn `.cs`, no subscription, no backfill, no order/arm/gate path,
no chart, no `deploy/` file, no DB write.

Forbidden-territory check: nothing in this branch touches the planner's
STRATEGY (only what the prompt DISCLOSES about its tape), levels, regimes **as
computed** (`RVBaselineFrom5m` returns byte-identical values — pinned), any
gate, arm or order path, arm 133, the chart, or the store's real hole (the
hole is DISCLOSED, never filled).

**A29 — "production call sites: 0" grep on every new function:** asserted by
`TestD2WiredAtTheThreeCallSites`, `TestRingRehydrateIsWiredAtBoot`,
`TestReadFactRowBuilderIsWired` and (from `4b2edea3`)
`TestBarHorizonHasProductionCallSites`. All pass.

**A12 — guide + SYSTEM-MAP** are updated in the SAME commit as each behaviour
change. `GUIDE_BUILT_REV` is deliberately **left at `954f11b15f2e`**: this wave
is not deployed, so bumping it would make the guide claim to describe a binary
that does not exist. The drift banner correctly reads "older rev" until the
cutover, and the bump belongs to the deploy that ships it.

---

## OPEN ITEMS

1. **`kernel/weekly_prompt.go` has the identical C1 defect, unfixed.**
   `## Weekly candles (12 completed weeks, oldest → latest)` renders over as few
   as one row; `CompletedWeekCandles(bars1m, now, 12)` and
   `LastNWOGs(bars1m, now, 5)` ask for depth the tape cannot hold. Deferred
   because `SectionsText` feeds `WeeklyFactsHash`, so a heading change moves a
   stored hash — a deliberate decision, not a side effect.
2. **`planner_read_facts` still attributes a floor to the wrong tape.** Row 66
   records `scope_intv='1m' scope_bars=2000` beside `atr5m=19.7631429163709` and
   `stop_floor_pts=29.6447` — but `plannerATR5m` fetches **5m × 200**, a
   DIFFERENT tape, itself holed on that day. `read_horizons` is already a JSON
   ARRAY so a second entry (`who="atr5m"`) is additive, but wiring it needs a
   `plannerATR5mWithHorizon` variant, which is outside D4's stated scope. The
   discarded uncommitted pin (see below) was written for exactly this.
3. **Live proof owed.** Every figure here is measured against the store and the
   stored prompts; nothing has run on the deployed binary. The three things to
   read on the next boot: the `🧯 ring rehydrated …` lines with their resolved
   counts; the first planner prompt carrying a `TAPE:` line and HELD-vs-REQUESTED
   headings; and the first `planner_read_facts` row with `read_horizons != ''`.
4. **`GUIDE_BUILT_REV`** bump belongs to the deploying dispatch (above).
5. **A pre-existing FE test failure**, unrelated to this wave and present on the
   untouched tree: `web/src/guide/GuidePage.test.tsx` fails in a worktree with
   `Error: Denied ID /home/hoang/nofx-barshorizon/branding/product.txt?raw` —
   vite's `fs.allow` root does not cover the worktree path. **Verified
   pre-existing** by reverting the guide edit and re-running: identical failure,
   `1 failed | 3 passed`. `npx tsc --noEmit` passes with the edits.

## UNCOMMITTED-FILE DISPOSITION

Four files were left uncommitted when the previous build was stopped mid-flight.

| file | disposition | reason |
|---|---|---|
| `kernel/void_scope.go` | **KEPT** | `VoidScope.Horizon` + the boot line. Correct under the corrected premise; its cited figures (3,055-minute span, 696 gaps) were re-verified first-hand and reproduce exactly. |
| `store/planner_read_facts.go` | **KEPT** | The four additive columns + `HorizonRecorded()`. Its comment already stated R1 as NOT REPRODUCED and explicitly refused to redefine `scope_bars`. |
| `store/planner_read_facts_test.go` | **KEPT** | PIN 4 (a legacy row reads UNKNOWN, not zero). Mutation 17 confirms it is a real pin. |
| `trader/read_facts_horizon_test.go` | **DISCARDED and REWRITTEN** | As written it referenced four symbols that do not exist in the repo — `buildReadFactRow`, `plannerATR5mWithHorizon`, `traderProdCallSites`, `containsAny` — so it could not compile, let alone run. It was also built around the atr5m second tape, which is outside D4's stated scope. Rewritten narrower against the void tape (pins D4-A/B/C, plus D4-D/E added by mutation); the atr5m tape is filed as OPEN ITEM 2 with its evidence rather than dropped. |
