# VERIFY-0926 — Trading Pipeline stages 1–3 (DS-104 adversarial second pass)

READ-ONLY. Zero code changes. Scope: DS-106's `2026-09-26-audit-trading-pipeline.md`
(@ `audit/0926-trading-pipeline` `c3ee15ca`), stages 1–3 plus their §4 refutations.

Base per L3: `git log -1` = `04ae1c2f RELEASE 6cac1b896bdb… — boot 3: #227 … booted
19:33:22 CT 2026-09-25`. Worktree `/home/hoang/nofx-ds-104-verify0926` @ `04ae1c2f`,
branch `verify/0926-trading-plan` (claim `599929ac`). DB copy `v.db` via
`sqlite3 -readonly "file:…?mode=ro" ".backup"`; live log `data/nofx_2026-09-25.log`
(50,311 lines) read-only. No tests run (L17; box is DS-101's).

## Verdict

**No new P0/P1.** DS-106's "zero surviving" holds under this pass. Their stage map is
substantially accurate; three claims are PARTIAL (one of them a census mis-scope that
must be corrected in the consolidated report), and all three P0/P1 hunt candidates
refute.

Counts: **CONFIRMED 8 · PARTIAL 3 · REFUTED 3 · NEW 0 (1 census correction).**

## 1. Check-by-check (every dispatch question)

### Stage 1 — Data

**1.1 Stale/forming bar use.** CONFIRMED. `barsToKlines` stamps `CloseTime = T +
dur − 1` (the SCHEDULED close) and `Final` from the wire; the comment-truth is that
the newest bar is usually forming and consumers needing closed bars filter
`CloseTime < now` (`trader/ninjatrader/bars_market_bridge.go:64-84`). [A]

**1.2 Short read is warn-only.** CONFIRMED. `barsFromCache` returns nil on empty,
the tail when the ring holds more, "whatever it has" when less — the horizon WARN
(`bar_horizon_warn.go:214`, WARN not INFO so it reaches `log_events`) is the only
signal; nothing is refused (`bars_market_bridge.go:37-62`). [A]

**1.3 BarCache after reconnect / feed-down.** CONFIRMED. `feed_status` frames set
`TCPServer.feedStatus` (`provider/ninjatrader/history_rerequest.go:112-114`,
`tcp_framing.go:509-512`); `ninjaFeedDown` reads it (`trader/auto_trader.go:64`); the
admission gate refuses only once a frame says down — the documented default-ALLOW
posture (DS-106 NOTE). [A]

**1.4 The 4h shortfall (446/500) — impact.** CONFIRMED as a low-impact, fail-open
observation. Live line 08:56:16 [A]: `bar horizon SHORT: MNQ 4h asked=500 served=446
span=2513h … ring=2500 callers=[kernel/engine_analysis.go:411×3
trader/auto_trader_wake_levels.go:281×4]`. The served set is 89.2% of the ask — the
oldest ~54 4h bars (~9 days) of HTF context are missing for the executor's analysis
context and the wake-levels reads; the planner's own HTF map is not a named caller on
this line. Nothing is refused by design (A10/A24). The line also carries a genuine
datum: "served 446 exceeds the 366 OPEN intervals in this span" — the ring holds more
4h bars than the span's open intervals, i.e. the shortfall is horizon depth, not a
hole. [A]

### Stage 2 — Planner

**2.1 Every read type → model + effort actually sent.** PARTIAL. Code path [A]:
`resolvePlannerClient` (`auto_trader_planner.go:60-133`) pins the exact model
(`pinExactModel`, never a provider alias); the call site re-asserts thinking per call:
scheduled/ordinary reads → `planReasoningWire()` = `AI_PLAN_REASONING`, default
`max` (`config.go:232`, `auto_trader_loop.go:65-73`); a read with `fastTapePending`
set swaps to `fastMarketReasoningWire()` = `FAST_MARKET_REASONING`, default `fast`
(`auto_trader_loop.go:89-97`), applied via `mcp.ApplyThinking` inside the call
closure (`auto_trader_planner.go:1416-1436`). Live evidence [A]: the 14:05:02
structure-flip read (NY v4) ran `reasoning=fast→low wire=enabled/low` — the fast-tape
wire, which was the DEFAULT `fast` before 20:14. **After the 20:14:25 restart
(`BOOT INTEGRITY OK — rev 6cac1b896bdb … pid 21600`, log line 48104) there are ZERO
planner/session/weekly lines in the remaining 2,207 log lines and the newest plan row
in the DB is 2026-09-25 14:06:43 NY v4.** So the claim "post-20:14 reads show max"
has no log instance to quote — it is not refuted, it is UNOBSERVED. The `FAST_MARKET_
REASONING=max` value itself lives in `.env` (unread per L12), exactly as DS-106's §5
disclosed. The code guarantees that when a read does fire, the knob is read at that
instant (env is read per call, not cached). [A]

**2.2 The 30-min hold after a failed death re-read.** PARTIAL (precision, not
error). The class-47 level-wake cooldown default IS 30 min
(`WakeCooldownMinDefault = 30`, `trader/class47_wake_cadence.go:59`). But the DEATH
paths' holds are different: the death REREAD self-backoff is `wake_min_interval_min`
measured from the last launch that wrote nothing (`trader/death_reread.go:265-270`),
its launch waits out the 5-minute flap guard (`DORMANT_MIN_HOLD_MIN`, `:185-205`),
and FLIP reads are exempt from the cooldown entirely (reaction reads,
`auto_trader_planner.go:796-803`). DS-106's "holds + 30-min retry" collapses three
different cadence rules; each of the three is sane. [A]

**2.3 Fail-closed after 3 attempts.** PARTIAL (read-type-dependent, and DS-106's map
misses the live configuration). `runPlannerReadWithTriggerClaimedCtx` takes
`failClosed` per call site [A]: scheduled reads, death re-plans (`runDeathReplan`,
`:1219-1221`, true) and owner re-reads are `true` — exhaustion writes the NO-TRADE
doc (`writeNoTradePlan :1227`, with the fail-closed map obeying min_grade). Wakes
(level_event `auto_trader_wake_levels.go:404`, structure_mss
`auto_trader_transition.go:217`), flips (`:940`) and — the nuance DS-106 did not map —
the death REREAD (`trader/death_reread.go:303`) are `false`: a failed re-read leaves
the still-active/dormant plan standing and writes nothing. Live knob [A]: boot lines
01:40:03 / 19:33:23 print `🧬 death→reread=on(default)`, so the LIVE death path is
dormant + death-reread (fail-open, re-arm-on-close-back design), NOT the legacy
`runDeathReplan` fail-closed path DS-106's map line 3 describes as the mechanism. The
"sits out after 3 failed attempts" answer is true for scheduled/owner/legacy-death
reads, false for wake/flip/death-reread reads — by design (the dormant re-arm
predicate exists precisely so a failed opportunistic read doesn't end the session).

**2.4 Repair path + born-dead + publish clock.** CONFIRMED. Repair: attempts ≥2
render the rejected-reason block verbatim (`plannerRejectBlock :1742`, `:1800-1815`);
live proof 14:05:02-14:06:43: attempt 1/3 rejected A5 → `🧩 planner attempt 2/3
repair: prompt ~1613 tokens` → 2/3 rejected A1 (write-time feasibility) → 3/3 repair
→ `🗓️ PLAN written 2026-09-25 NY v4 (model deepseek-v4-pro, lifecycle active)`.
Born-dead: the born-check records at publish (`:1814-1816` comment), `bornCheckRefused
:1943`, and the death-born birth-wick guard (`deathBornWickActive`, death_reread.go
:143-174). Publish clock: P15 comment — the authoring clock is the caller's instant,
"the publish clock stays live" (`:1430-1434`); the per-read line carries the
measured read→publish latency (`plannerReadLine :1814`). Live [A]: 9 `🧭 planner
read` lines on 09-25, last at 14:06:43 with `attempts=3 reject_classes=A5,A1
read→publish=322936ms lifecycle=active`. [A]

**2.5 Model actually sent.** CONFIRMED from the log: all 09-25 planner calls
`🧠 planner model … pinned "deepseek-v4-pro"` (e.g. 14:01:20) and every `🗓️ PLAN
written` carries `model deepseek-v4-pro`. [A]

### Stage 3 — Arming

**3.1 Policy planned_order vs market_in_zone + zone geometry.** CONFIRMED.
`zoneLegFor` decides per scenario; a `market_in_zone` leg enters at the FAR edge
(worst fill every gate below judges) and composes the stop from the NEAR edge; a bad
zone refuses the leg (`armed_executor.go:530-540`, W3 D7/D9). `planned_order` and
legacy rows fall through byte-identically (`:1361-1363`). Picture legs = zone legs
with a picture R:R floor (D10/D11, `:541-550`). [A]

**3.2 Stop floor/anchor — is the live WIDENED line sane?** CONFIRMED sane.
`composeArmStop` (`trader/arm_stop_anchor.go:90-187`): stop = the WIDEST of
{ATR floor (MIN_SL_ATR_MULT×ATR5m), anchor (nearest seated level on the risk side
within 3.0×ATR5m + tick clearance), authored} — it never TIGHTENS the planner's stop.
When no seated level is in the dead-zone bound the arm is `stop_unanchored`, the ATR
floor governs, a WARN fires, and a counter records (n=719 today) — a level is never
invented. Live [A]: 256 `WIDENED` lines on 09-25, e.g. 01:46:10 `stop 30739.39
(authored 30743.00 WIDENED) · anchor VWAP 30753.58 → beyond 30753.08 · atr_floor
30739.39 (1.5×ATR5m 17.07) · bound=atr_floor` — for the long the VWAP anchor sits
ABOVE entry (wrong side), so its candidate stop (beyond 30753.08) is on the profit
side and correctly loses to the ATR floor; the authored stop was widened by ~3.6 pts
to the 1.5×ATR5m floor. 270 `stop_unanchored` lines, e.g. 07:01:15 LONDON S1 short
`stop 31026.00 · anchor none · atr_floor 31025.70 · bound=authored` — authored won
(was wider than the floor). All sampled compositions obey "widest wins, never
tighter". [A]

**3.3 Re-arm / cancel / rest cap / one_setup.** CONFIRMED. Re-arm never auto-re-arms
(1.4, `armed_executor.go:248-285`); working rows whose new version re-priced the
authored bracket ≥2 ticks or whose zone no longer holds its limit are CANCELLED via
the filled-arm guard (`respecWorkingArm`, `trader/arm_respec.go:195-260`; a
`cancel_pending` row is never placed — no hole); a marketable-through authored limit
is cancelled, NEVER placed (`armed_executor.go:1425-1431`,
`limitMarketableWrongSide`); placement band = `ARM_PLACE_TICKS` default 100
(`:44`, `:1311`); the zone rest cap ends a rested policy limit
(`zone_placement.go:524-537`); `one_setup` is the structural-fade play with
`ResolveStructuralStop` geometry and min-grade default B (`kernel/one_setup.go:29-38`,
`armed_executor.go:550-563`). [A]

### P0/P1 hunt (the three questions)

**H1 — "can a plan arm something the plan did not say?" REFUTED.** Arms iterate ONLY
`doc.Scenarios` of the ACTIVE plan's own doc (`armed_executor.go:309`, `:411`,
`:988`); every row is stamped `plan_id:version:scenario:leg`, and the scenario spec
is consumed at fill/void on the SAME version (fresh authorization = new version).
Shadowed conditions write an inert `shadowed` row that is never placed. [A]

**H2 — "can it arm on stale geometry?" REFUTED, with one named caveat.** (a)
marketable-through limits are cancelled-never-placed (`:1425-1431`); (b) a working arm
whose new-version geometry moved ≥2 ticks is cancelled and re-arms only after the book
confirms (`arm_respec.go:195-260`); (c) zone arms re-evaluate the zone each cycle and a
zone that no longer holds its limit is cancelled (W3 D17); (d) write-time feasibility
refuses geometry with `no_provenance (entry_zone_ambiguous)` (live 14:05:20, attempt
2/3). Caveat (design, not defect): within the 100-tick band a resting limit that is
NOT marketable keeps resting even if price has drifted toward it — that is the
documented place-then-manage behavior, and the drift is bounded by the band and the
per-cycle re-spec. [A]

**H3 — "can an arm outlive its plan?" REFUTED.** `plan == nil` / lifecycle
`no_trade` / `dormant` / row unavailable → `cancelArmedOrdersSync` fires the same
pass, synchronously, before any other placement work (`armed_executor.go:248-285`);
session end (EOD flat) cancels everything; stale UNPLACED arms of superseded versions
expire (class 47 F4, `:290-310`); a newer ACTIVE version either re-specs (cancel) or
re-prices the working arm. [A]

## 2. Census correction (DS-106 §2)

**PARTIAL — the "since midnight UTC" window is wrong for both figures.**
DS-106: "decision_records since midnight UTC: 46,698 rows; 566 carry
risk_check_error". The DB copy [A]:
- 46,698 = the ALL-TIME row count (min `timestamp` 2026-05-30 04:03:43+00:00, max
  2026-09-25 20:59:19+00:00; `created_at` NULL = 0).
- 566 = the ALL-TIME `risk_check_error != ''` count.
- Since 2026-09-25 00:00:00 (the `timestamp` column): **619 rows, 9 with
  `risk_check_error`**.
The two "since midnight" figures DS-106's report implies do not exist in the data.
Their §4.3 support ("Live census: 566 decision rows carry risk_check_error") stays
true as an all-time figure, so the refutation itself is unaffected — but the
consolidated report must scope these numbers correctly (sample-id law: the corrected
window is `timestamp >= '2026-09-25 00:00:00'`).

## 3. Re-check of DS-106 §4 refutations (stages 1–3 items)

- **#1 EOD/T1 sync cancel vs just-filled row** — CONFIRMED at the code: the drain
  counts fills during the sweep separately, cancels by SIGNAL id, records
  `cancel_pending` (never `cancelled`) when the wire is missing
  (`armed_executor.go:2638-2705` region). [A]
- **#2 agent door without EntryGate** — CONFIRMED (out of stage scope but cheap):
  `admitAgent` runs the fail-closed bracket refusal BEFORE EntryGate
  (`entry_admission.go:390-405`). [A]
- **#3 silent refusal** — CONFIRMED: `admitRefuse` always logs + `IncGateBlock`
  (`entry_admission.go:120-136`). [A]
- **#4 double-open latch** — CONFIRMED: `wireNT8EntryLatch` is the ONLY production
  caller of `SetEntryLatchSource` and `NewAutoTrader` calls it unconditionally for
  every `*TCPTrader`; the boot line READS wired/UNWIRED
  (`trader/entry_latch_wiring.go:18-43`). [A]
- **#5 row 619 close** — CONFIRMED in the DB copy: 618 `CLOSED unresolved exit 0.0
  pnl_corrected NULL`; 619 `CLOSED sync exit 30909.25 pnl_corrected −32.0`; receipts
  table 1 row, `applied=1`. [A]

## 4. What I could not verify

- The `.env` values (L12) — `FAST_MARKET_REASONING=max` is UNOBSERVED in the log
  because no planner read has fired since the 20:14 restart.
- The C# AddOn (same as DS-106 — Go-facing contract + wire evidence only).
- No tests run; no `-race` (L16/L17).
- Anything beyond stages 1–3 (admission/execution stages audited by DS-106 but out
  of this dispatch's scope, except the §4 items above).

## 5. Notes for the consolidated report

1. Correct the decision_records census window (§2 above).
2. Name the death-path configuration precisely: live = `death→reread=on(default)`
   (dormant + fail-open reread + re-arm), with `runDeathReplan` as the legacy
   fail-closed path — "fails closed after 3 attempts" must not be stated
   read-type-independently.
3. The "post-20:14 reads show max" verification question has NO observation yet; the
   next natural instance is the 2026-09-26 ASIA read (16:30 CT). If the CTO wants log
   proof of `FAST_MARKET_REASONING=max` today, the only safe avenue is the owner
   quoting `.env` — a lane cannot.
