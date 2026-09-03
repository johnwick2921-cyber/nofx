# INVALIDATION-WIRED (checklist class 57)

**Branch:** `fix/invalidation-wired` off `75d923eb` (deployed rev `528edd78`)
**Commits:** `d2805408` · `c52c98e2` (+ this report) · pushed
**Checklist:** entry **57** (highest occupied at merge: **56**)
**Boot line:** `🛡 arm gate: invalidation-wired=on · armed-under surfaces=on …`
**Status:** NOT DEPLOYED. Rides the next boot.

---

## 1. What the four defects have in common

Every one is a record the system writes and nothing reads.

| # | the record | who read it |
|---|---|---|
| E1 | `🎯 scenario S1 → ≈invalidated @ 29285.00` | nobody — labelled "never execution-wired" |
| E2 | `armed_under_version` (attribution wave) | **zero readers** outside the store |
| E3 | `fill_quantity` on the armed row | nothing ever wrote it |
| E4 | the authored-log dedup key | blind to the row's state |

## 2. E1 — the twelve minutes

```
08:50:54  🎯 scenario S1 → ≈invalidated @ 29285.00 (price accepted through the
          level against the trade — display-only estimate, never execution-wired)
09:02:54  ⚔️ armed NY S1 leg 1 short limit 29285.00
09:03:53  filled
09:20:45  stopped at the widened stop — −$140
```

Ledger: `armed_orders` row 35, version 2, S1, short, entry 29285.00, filled.

The system reached its verdict twelve minutes before it armed the trade that
verdict condemned. The label was accurate.

### 2.1 The fix

`EntryGate` gains **leg 3**, arm path only. The verdict comes from
`kernel.EvaluatePlanScenarios` — the same function
`trader/auto_trader_levelstate.go:260` calls for the display line, with the same
`BarsSince(bars, plan.BirthMs)` windowing. There is no second predicate to
drift from the first; that is the void-parity lesson applied here.

It runs **before** the pricing legs deliberately. R:R and min-SL ask whether a
trade is well shaped. This asks whether the setup still exists. A well-shaped
trade into a dead setup is exactly the 09:02 arm.

```
entry_gate: scenario S1 invalidated at 08:50 (accepted through 29285.00)
            — price accepted through the level against the trade
```

Counted under its own class, `invalidated`, placed **first** in
`armRefusalClass` so it cannot fall through to `other` and vanish from the
tally.

### 2.2 Unresolved is not a refusal

No bars, no price, or an `UNEVALUABLE` scenario (the display path refuses to
store a status for those, so the gate must not invent one) returns `ok=false`.
The leg **passes** and logs `invalidation check unavailable for S1 — leg PASSED
(no verdict is not a refusal)`.

### 2.3 The verdict time, not the check time

The evaluator is stateless: it knows a scenario **is** invalidated, not **when**
it became so. Rendering `FormatCT(now)` would have printed the check time under
the words "invalidated at" — and twelve minutes is the entire point of this
wave.

So the evaluator stamps `scenario_invalidated_at:<trader>:<plan>:<scenario>`
**once** on the transition, never overwritten, and the gate reads it. Absent the
stamp it renders "at an earlier cycle", which is true.

## 3. E2 — the card showed one S1 and the owner held the other

The card rendered NY **v3 S1 long**, written 09:15, while the account held a
position armed under **v2 S1 short** — filled 09:03, stopped 09:20. Both
scenarios are called "S1".

`/plan/today` gains `open_position`: the position's own provenance, plus
`version_differs`. `ArmedUnderBlock` renders, in this order:

```
POSITION ARMED UNDER · v2 S1 short @ 29285.00 × 1
PLAN NOW · v3 — the rows below are THIS plan, not the position above
```

Nothing renders when flat, or when the two versions agree — there is nothing to
disambiguate then. The chips also carry `armed_under_version` now.

**A23 note.** `armed_under_version` is **0** on both of today's rows (35 and 36,
created 09:02 and 09:20) because the attribution wave did not boot until 10:28 —
the column existed, nothing wrote it. That is the class-56 disease in a fresh
column: 0 means both "version 0" and "never stamped". The payload sends `null`
plus `armed_under_note`, and the card renders **"version not recorded"**, never
"v0". `armedUnderVersionOf` falls back to the mutable `Version` for the log line
rather than printing zero.

## 4. E3 — filled, quantity zero

Row 35: `state=filled`, `fill_quantity=0`, beside `trader_positions.quantity=1`.
Nothing wrote the column, and 0 is a legal "nothing filled", so the row could
not be read either way.

`SetFillQuantity` is called on the same path that stamps lineage. WHERE-scoped,
idempotent, and **a zero never overwrites a measurement** — pinned, because that
is how the column would silently regain its ambiguity.

## 5. E4 — "armed" four times after it filled

The authored log fires when the ledger lookup finds no existing row. The lookup
is `ListNonTerminal` (`state IN ('armed','working')`), so once a row **fills**
it is invisible, `row.ID` stays 0, and the authored branch runs again.

The dedup value now carries `state=…`, and `armAuthoredLoggable` refuses to log
"armed" for a `filled` / `cancelled` / `expired` row at all — the second guard
is the load-bearing one, since state in the key alone would make a state change
*more* loggable, not less.

### 5.1 CORRECTION — I had this backwards, and my own fix did not work

My first pass recorded, as **[B]**, that a terminal row being invisible to
`ListNonTerminal` meant a filled scenario "can be re-armed". **That was wrong.**
Traced properly:

`UpsertArm` runs its OWN query with no state filter
(`store/armed_orders.go:166`), finds the filled row, and hits MANUAL-CANCEL-WINS
at line 206:

```go
if existing.Version == row.Version && !IsBootSweepReason(existing.StateReason) {
    return nil
}
```

So the re-arm guard **is store-derived**, is a DB read, and holds across a
restart. Nothing was ever re-armed.

What it does is return `nil` having done **nothing**, leaving `row.ID` at 0 —
and the caller then logged `⚔️ armed …` as though it had succeeded. **The guard
was real; the log was the lie.** Measured: five such lines after the 09:03:53
fill, each carrying a different ATR-drifted stop:

```
⚔️ armed NY S1 leg 1 short limit 29285.00 SL 29354.91 …
⚔️ armed NY S1 leg 1 short limit 29285.00 SL 29352.65 …
⚔️ armed NY S1 leg 1 short limit 29285.00 SL 29354.44 …
⚔️ armed NY S1 leg 1 short limit 29285.00 SL 29352.40 …
⚔️ armed NY S1 leg 1 short limit 29285.00 SL 29354.86 …
```

**And my F4 fix was ineffective.** `armAuthoredLoggable(row.State)` read
`row.State` — the DESIRED state the caller had just built as `"armed"`, never
the persisted one — so it passed every time. The drifting stop in the dedup
value made it worse: the value changed each cycle, so the dedup suppressed
nothing either. Two guards, both no-ops.

Corrected: `armedActually(row.ID, row.State)` — the id is the load-bearing
signal, because `UpsertArm` sets it on a create and on a live-row update and
leaves it zero exactly when it declines. The dedup value drops the ATR-derived
prices and keeps the entry, which is the GAR-F6 lesson the refusal path learned
in August and the authored path never did.

### 5.2 The remaining hole, reported

The guard is scoped to the SAME version. `UpsertArm`'s next branch
re-authorizes a terminal row on a **new plan version** — fresh `armed` state,
`fill_quantity` reset, `armed_under_version` re-stamped. With a position still
open, `oneLiveArmGuard` explicitly skips the same side
(`armed_executor.go:601`, "same-side add — outside this guard's scope"), so a
v3 S1 short can arm while the v2 S1 short position is still open and add to it.

Today that did not fire: NY v3 S1 was authored LONG, and the opposite-side
branch of `oneLiveArmGuard` would have refused it. **[A]** for the code path,
**[B]** for the consequence — never observed placing a second same-side arm.
Out of this dispatch's footprint (the owner scoped F2 to "in this version");
worth a ruling.

## 6. Tests

| id | pins | first run |
|---|---|---|
| T1 | the 09:02 arm intent + the 08:50 verdict → REFUSED with the message | **RED quoted**: `got reason=""` with the leg disabled |
| T2 | evaluator unavailable → passes, with the line | RED (undefined) |
| T3 | a live scenario passes; no resolver = leg off | RED (undefined) |
| — | arm-path only, asserted on the REASON | see 6.1 |
| — | the refusal classes as `invalidated`, not `other` | RED |
| T4 | the card's two lines and their ORDER; flat; agreeing; "version not recorded" | RED (no component) |
| T5 | `fill_quantity` stamped, the zero guard, the state dedup, the attribution fallback | RED |
| T6 | `go build ./...` · `go test ./...` · vitest 41 files / 310 tests · `tsc --noEmit` | GREEN |

### 6.1 A test that passed for the wrong reason

The arm-path-scope test first asserted the decision path simply does not refuse.
It failed — because the decision path refused on **R:R 2.00 below floor 3.00**,
nothing to do with invalidation. Had the fixture happened to clear R:R, the test
would have passed while proving nothing about scope. It now asserts on the
refusal **reason** and uses a target that clears the floor, with the reason in a
comment.

### 6.2 My own error, recorded

Demonstrating T1's RED, I disabled the leg with `if false &&`, then reverted with
`git checkout trader/entry_gate.go` — which discarded **the entire file's
uncommitted work**, not just the probe. Rebuilt from the same edits and
re-verified green. It is the class-45 lesson (a blunt revert destroys more than
the thing you aimed at) in a different tool, and the reason A18's "commit ~30
min" exists.

## 7. Boot line

```
🛡 arm gate: invalidation-wired=on · armed-under surfaces=on — the evaluator's
   own ≈invalidated verdict REFUSES an arm (same function as the display path;
   unresolved = pass + a line), and a position states the version it was armed
   under before the live plan's rows are read as its own
```

## 8. Cutover

Rides the next boot alongside the fast-market exemption
(`fix/wake-fastmarket-exempt`). Five-leg gate `ready:true`, owner GO, A13
rollback named by the rev it holds, A19 all four halves with the marker
**pushed** before the lock is released.

**PROOF OWED:** the next arm on an already-invalidated scenario showing the
refusal line, and the next open position's card showing its armed-under version.
