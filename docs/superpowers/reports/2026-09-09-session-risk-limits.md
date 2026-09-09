# SESSION RISK LIMITS — dispatch 104 (W4)

Owner: hoang · agent `sessionrisk-554049f5/nofx-80[aa87f6]` · branch
`fix/session-risk-limits`, claimed at `9b578c06` off dev's tip `05125bd6`.
Worktree `/home/hoang/nofx-srl`; main tree untouched (A2b).

**Running rev, my own read:** `954f11b15f2e7615678f7d2b708c47895faebf1e`,
`vcs.modified=false`, pid 438, up since 2026-09-09 13:02:45 CT; `/api/health`
agrees. Matches the dispatch.

**Basis, pinned:** `docs/superpowers/research/2026-09-08-range-fade/report.md`
@ `f5927cdc` (requirement 6 and the "do not"); `2026-09-05-vet-06-risk.md`;
the two-day audit @ `f890ea60` D38.

---

# THE HEADLINE, IN PLAIN WORDS

**The $450 daily limit is decorative twice over.** Two switches are off, and
BOTH must be on before it enforces anything:

| knob (`ai_config.risk_control`, strategy `a5b7662e`) | value |
|---|---|
| `guardrails_enabled` — the master | **false** |
| `daily_loss_enabled` — this leg's own toggle | **false** |
| `daily_loss_limit_usd` | 450 `[O]` |
| `consecutive_loss_halt` | **absent** (0 = off) |

Confirmed on the live surface, firing every cycle at
`kernel/engine_analysis.go:173`:

```
⚠️ Strategy Studio: risk guardrails master OFF — daily loss/profit/trade limits
   + blackout NOT enforced this cycle
```

The gate itself is `kernel/risk_limits.go:309` —
`if g.DailyLossEnabled && g.DailyLossLimitUSD > 0 && ...` — so the leg's own
toggle gates it independently of the master. **Until both move, nothing in this
wave makes the daily limit bite.** The desk strip's DAY line and the boot line
now say exactly that, in those words.

**What DOES bite today:** the consecutive-loss breaker. Its own comment says it
is *"a per-strategy circuit breaker, NOT gated by the guardrails master switch"*,
and this wave wires it to the path that actually trades.

---

# SECTION C — THE EVIDENCE

## C1 — the daily limit's four holes

| hole (vet-06) | status at `954f11b1` | file:line |
|---|---|---|
| (a) up-to-4-minute arm window before the check | **reported, not changed** (leg-D logic, out of scope) | `trader/entry_gate.go:171` |
| (b) realized-only input | **reported, not changed** (leg-D logic) | `kernel/risk_limits.go:309` |
| (c) clears only at the first AI cycle after the roll | **WORSE THAN REPORTED — FIXED** (below) | `kernel/risk_limits.go` `MaybeResetDaily` |
| (d) no `daily_force_flat` case in `armRefusalClass` | **CONFIRMED — FIXED** | `trader/armed_executor.go` |

**(c) is not what vet-06 said, and the truth is worse.** The trip did not clear
on the CME roll, and it did not clear at the next AI cycle either: **it never
cleared automatically at all.** `MaybeResetDaily` rolled `lastDailyResetDate` and
left `forceFlatReason` untouched; `clearAllDailyForceFlat` had exactly ONE caller,
`ResetDailyPnLAt`, reachable only from the operator's `POST /api/risk/force-flat`.
A tripped desk stayed blocked across the roll, the next session and the one after,
until a human reset it or the process restarted.

The code has promised otherwise since #91, in the comment above the state:
*"Cleared by the CME session-day reset (ResetDailyPnLAt), so a trip lasts the
session-day and lifts with the daily window."* Nothing wired it. Unseen because
the master is OFF and the trip has never fired — and the latch is an in-memory
map a restart clears, so even a live occurrence would likely have been erased
before anyone correlated it. **Checklist class 96.**

## C2 — the breaker exists, is wired, and guards the wrong path

`consecutive_loss_halt` is registered `KnobLive`
(`store/knob_registry_table.go:30`) with reader `consecutiveLossHalted`
(`trader/auto_trader_orders.go:112`) and a production call site at `:250`, inside
`executeDecisionWithRecord`. Resolved value: **0 (OFF)** — the knob is absent
from the live config.

**But the wiring is on the DECISION path**, and `plan_mode=strict` — the live
setting — routes every entry through the ARM path: *"plan_mode=strict executes
plan scenarios on the ARM path only, and this is a %s-path market entry"*
(`trader/entry_gate.go:186`). The breaker guarded a door nobody walks through.

## C3 — REFUTED AS STATED, and the real holes are narrower and worse

The dispatch asked whether the 14:45 flatten cancels resting arms. **It already
did** — the S-LIST CLOSER (2026-08-27) cancels working arms synchronously,
ack-waited, before flattening, and `TestSListEODFlatCancelsArmsBeforeFlatten`
pins it. E1 as written ("MUST FAIL today, arms survive") would not have failed
for that reason.

Two real holes, both leaving a live order at the broker while the ledger reads
flat:

**(a) The early return.** `enforceEODFlatAt` read the open positions and
returned on `len(positions) == 0` **before** reaching the cancel. A session that
ended flat BY LUCK — nothing filled — left every resting arm alive past the
close, into the next session's tape. Nothing said so, because the position was
zero so the book "looked" flat. This is precisely the research's do-not:
*"Canceling remaining entries is part of being flat."*

**(b) `place_pending` settled without a send.** `cancelArmedOrdersSyncWith` read

```go
if r.State != "working" || r.SignalID == "" || cancelFn == nil || src == nil {
    _ = ledger.SetState(r.ID, "cancelled", reason)
    continue
}
```

`BeginPlacement` (`store/armed_orders.go`) sets `signal_id` **and**
`state = place_pending` in ONE update, **before** the order reaches the broker —
so a `place_pending` row ALWAYS carries a signal id and its order may already be
resting. This wrote `cancelled` **with no wire cancel at all**. Class 81 reached
by omitting the send.

**Count of arms non-terminal after a 14:45 flatten in the retained sample:** the
dispatch asked for this and it is not answerable from the record — `armed_orders`
is overwritten in place (no history table), so a row cancelled later is
indistinguishable from one cancelled at the close. Stated as UNKNOWN rather than
estimated. The holes are established from the code, not from a count.

## C4 — runs of losses on this tape

n = **73** usable. EXCLUDED and counted: **516** pre-era (< 2026-08-15 CT),
**3** `pnl_corrected` NULL, **3** `e7_farside_test`.

| measure | value |
|---|---|
| longest losing streak | **7** — ids **585, 586, 587, 588, 589, 590, 591** |
| session-days with a 3+ run | **7 of 15** (08-20 ×4, 08-21 ×3, 08-24 ×4, 08-25 ×3, 08-26 ×3, 09-02 ×4, 09-08 ×3) |
| losers | 47/73 = **64.4%** |
| worst single close | **−155.00** |
| 8 in a row ever reached | **NO** · 5 in a row: YES |

**So N = 8 `[I]` never fires on this tape** — it sits one above the observed
maximum, the same shape as the $450 that trips 0 of 12 days. That sentence is on
the boot line, not buried here. M = 5 `[I]` does fire.

## C5 — post-loss behaviour, and n is the finding

Stop-outs in the era: **4**. Minutes to the next arm: p50 **81.4**, p80 141.7,
min 0.0, max 141.7. Re-armed within 30 min: **2 of 4**. **Same level as the
stop: 0.**

The reason n is 4: of 47 losers, **42 close as `sync`**, 4 `stop`, 1 `manual`.
A cool-down keyed on `close_reason='stop'` would observe four events and read as
"this never happens" — so D3 triggers on a **losing close** (P&L sign), which
needs no attribution to work.

---

# THE EXIT-CAUSE FINDING — AND THE CORRECTION TO MY OWN CLAIM

The owner asked me to file the 89%-unattributed figure as a finding against Wave
A's claim that exit cause is "recorded, not inferred". **That framing is wrong,
and the correction is the finding.**

`ExitCauseFromBroker` (`store/position.go:138-152`) landed in `455dce5a`
(2026-09-05 21:37:17 CT), which **is** an ancestor of the running rev. Splitting
the record at that moment:

| window | n | `sync` | attributed |
|---|---|---|---|
| era ≥ 2026-08-15 | 73 | 66 (**90%**) | stop 4 · manual 2 · target 1 |
| **since Wave A** | 8 | 1 (**12%**) | stop 4 · manual 2 · target 1 |
| since Wave A, **losers only** | 5 | **0** | stop 4 · manual 1 |

**Every attributed close in the entire record happened after Wave A landed.** The
90% is pre-Wave-A history, when the string was a hardcoded literal. Wave A works;
the claim stands. What the figure actually measures is that only 8 closes have
happened since it shipped.

**A real gap remains, unfired.** The AddOn emits `exitReason = "limit"` for an
`-lx` limit exit (`ninjascript/VLTraderTCPClient.cs:1327`).
`ExitCauseFromBroker` maps sl/tp/manual and **has no `limit` case**, so such a
close lands in the `default` arm as `sync` — an attributed exit recorded as
unattributed. Zero occurrences so far. **Reported, not fixed: A31 puts the
recording path outside this wave's footprint.**

**Two further findings from the adversarial pass, also outside the footprint:**
post-Wave-A causes never reach the expectancy engine (5 of 8 post-Wave-A closes
have no `trade_excursions` row; `target` and `manual` have never appeared there),
and `hitOf` in `expectancy/aggregate.go` fabricates `false` rather than absent —
a canon class 49/53 shape. Both belong to whoever owns the record.

---

# SECTION D — WHAT SHIPPED

**D1 — flat means the book.** The cancel runs unconditionally and FIRST; positions
are read after it, so a fill that won the race is still flattened; a failed
position read logs `flatness UNVERIFIED` and never claims flat. The cancel guard
keys on the **signal id** — the only evidence anything could be at the broker —
so `place_pending` gets a wire cancel. A row that never got a signal id may go
terminal without asking; a duplicate cancel is idempotent where a missed one is a
live order we stopped watching.

**D2 — the breaker on both paths.** One resolution (`breakerHaltN`) shared by the
decision and arm paths: the owner's knob when set, else `8` `[I]`;
`BREAKER_HALT_N=0` is the explicit off switch. `M = 5` `[I]` WARNs without
refusing. Clears at the CME roll (`CMESessionDayStart`). Adjudicated ONCE per
cycle before the scenario loop — both are session-level facts, and a per-leg
re-query could answer the same question differently within one cycle.

**UNKNOWN never lengthens a run.** `CountConsecutiveLossesSince` excluded
unresolved rows in its WHERE and called them "never counted either way".
Excluding a row from the scan makes it **transparent**, not neutral: 3 losses, an
unresolvable close, 3 more counted as **six**. Bridging makes a halt MORE likely,
and blocking is this counter's destructive branch — the direction A24 forbids. An
unknown close now ENDS the run. The `e7` seam stays excluded: a synthetic row is
not a trade, so it neither counts nor breaks.

**The no-trade band on the arm path** (added mid-wave). `armed_executor.go` held
**zero** references to `InLunchNoTrade`, `InFirstNoTradeMinutes` or
`sessionEntryBlocked`; the sole enforcement was `auto_trader_orders.go:281`, on
the path strict forbids. An arm inside the band is refused; an arm already
resting when the band opens is **cancelled, not grandfathered**, through the same
seam the close uses. `sessionEntryBlocked` gains the A28 `At(now)` seam so both
paths read one band with one clock.

**D3 — post-loss counter, never a gate.** Triggered on a losing close; labels the
arm card and increments a per-(trader, session-day, session) counter.
K = `30` min `[I]`.

**D4 — (a)/(b) reported unchanged; (c) fixed (class 96); (d) `armRefusalClass`
gains `daily_force_flat`, plus `consecutive_loss` and `no_trade_band`.**

**D5 — boot line, Guide, SYSTEM-MAP** in the same commits.

**The desk strip, class 82 caught before it bit.** `deskGuardrail` checked only
the master while the gate requires both toggles. The moment the owner turned the
master ON and left `daily_loss_enabled` off, the strip would have reported the
$450 limit **ENFORCED** while the gate ignored it. Absent stays true, matching
`engine_analysis.go`'s `boolOrDefault(..., true)`, so no desk that never set the
field is silently disarmed.

---

# SECTION E — RED, GREEN, MUTATION

**E1(a)** — flatten with no position, one working arm:
```
RED   wire=[]  ·  1 arm(s) still non-terminal after the close with no position open
GREEN after D1
MUT   restore `if len(ps)==0 { return false }` → --- FAIL   (restored → PASS)
```

**E1(b)** — flatten with a `place_pending` arm:
```
RED   wire=[close_long:MNQ cancel_stops:MNQ]   ← position flattened, no cancel:sig-pending
GREEN after D1
MUT   restore `r.State != "working"` → --- FAIL   (restored → PASS)
```

**E2** — breaker table (8 cases), threshold resolution, roll clearing, UNKNOWN:
```
RED   run = 6, want 3 — an UNRESOLVABLE close bridged two runs of 3 into one of 6
GREEN after the position_query.go change
MUT   halt at `> n` instead of `>= n`  → FAIL at_the_halt
MUT   band no longer precedes the record → FAIL inside_the_no-trade_band, band_wins_over_a_halt
MUT   remove `at.sessionRiskGateAt(now)` from the arm path → FAIL TestArmPathConsultsTheSessionRiskGate
```

**E3** cool-down counted, never refused · **E4** `daily_force_flat` has its own
class · **E5** shortened-day pull-in + the flatten's cancel-before-read ordering
· **D4(c)** roll lifts the trip (RED: *"the trip SURVIVED the CME roll"*),
race-checked.

**A pin of mine that did not bite, and was fixed.** The DAY-line test called
`deskDailyLimitText` directly, so reverting `deskDay` to its old literal left it
GREEN. Built is not wired; it now asserts the call site and fails on that
mutation.

**A contract test updated, not deleted.** `TestCountConsecutiveLossesSince`
asserted that an unknown-P&L close at the tail was excluded so an earlier loss
still governed. Under the owner's UNKNOWN ruling it now ends the run. Updated in
place with the reasoning, plus a new assertion that a later loss starts a fresh
run — so the break is a break, not a permanent mute.

**Suite at the branch head:** Go **20 packages ok, 0 FAIL**.

---

# SECTION A15 — WHAT THE OWNER WILL STILL SEE WRONG

- **The daily limit still enforces nothing.** Two switches, both off. This wave
  makes the state legible and the refusal countable; it does not turn the limit
  on — that is the owner's knob (A31).
- **The breaker will not fire either, on this tape.** N=8 is one above the
  observed maximum. It is ON by default now, and honest about that on the boot
  line. M=5 will fire and WARN.
- **`place_pending` cancels are ack-waited, not book-confirmed.** D1 sends the
  wire cancel the old code omitted; the settlement pass (`confirmPendingCancels`)
  is what confirms against a fresh snapshot. The "flat" claim is therefore as
  strong as the ack plus the next settlement pass, not stronger.
- **The `-lx` → `limit` mapping gap** and the two expectancy findings are
  reported and unfixed, outside the footprint.
- **The count of arms surviving past a historical 14:45** is UNKNOWN and stated
  so — `armed_orders` has no history table.
- **75/76/77 remain duplicated** across the checklist's two numbering formats,
  from waves before this one.

---

# ROLLBACK

Single Go boot, no AddOn change. Preserve the running binary as
`nofx-bin.old.954f11b1` (verified with `go version -m`), restore
`deploy/RELEASE`, `mv` it back, owner runs `kill -9`; systemd relaunches.

Everything here is additive except three behaviour changes, each independently
revertable: the flatten's ordering (`auto_trader_clock.go`), the cancel guard's
key (`armed_executor.go`), and the loss-run's UNKNOWN semantics
(`store/position_query.go`). `BREAKER_HALT_N=0` disables the breaker at runtime
without a rebuild.
