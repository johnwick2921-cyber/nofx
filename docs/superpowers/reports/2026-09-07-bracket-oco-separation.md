# BRACKET-OCO SEPARATION — Section C verification (MEASURE FIRST, A17)

**Status: BUILT. C1 was REFUTED and the owner ruled replacements; those were built.**
Part 1 below is the measurement that stopped the original build. Part 2 is what
shipped. Awaiting the owner's GO for F1 (Go boot) and a separate GO for F2 (NT8).
Owner: hoang · agent session `bracket-oco-554049f5` · branch `fix/bracket-oco-separation`
Worktree base: `83314d69f973a22fd0567c5ae7d99e346db55c71` (dev tip at accept, 2026-09-07 09:50:52 -0500)
Running rev measured by me: `vcs.revision=44ea117a02a1d6703003109a3386c92ca55d2bd5`,
`vcs.modified=false`, pid 2745590 → `/home/hoang/nofx/nofx-bin`; `/api/health` → `{"revision":"44ea117a02a1"}`.

A1 spec-freshness: `git log -1 -- ninjascript/VLTraderTCPClient.cs` →
`291a299c6e2d52e2fc3f78538ac0598a507a12a0  2026-09-05 22:37:49 -0500`
("fix(wave B repair): canonical side at the placement entry…"). The file has NOT
been touched since the 09-06 incident. My base is newer than the file; nothing
moved under me.

---

## C1 — REFUTED [A]

> Dispatch: *"Our entry currently shares an OCO id with its stop and target, so the
> 09-06 cancel of entry aa07e583 took the accepted stop 29554 with it."*

**The three do not share an OCO id. They have not for some time.** `Account.CreateOrder`
is positional — after `quantity` come `(limitPrice, stopPrice, oco, name, gtd, customOrder)`
— so the OCO argument is the 9th slot:

| order | file:line | oco argument |
|---|---|---|
| ENTRY | `ninjascript/VLTraderTCPClient.cs:1002-1005` | `string.Empty` — **no group** |
| STOP  | `ninjascript/VLTraderTCPClient.cs:1863-1866` | `exitOco` |
| TARGET| `ninjascript/VLTraderTCPClient.cs:1867-1870` | `exitOco` |

with `string exitOco = signalId + "-exit";` at `:1862`.

The separation is deliberate and documented in the file itself at `:956-961`:

> *"Submit the ENTRY only; the protective SL/TP are placed once the entry fills
> (OnOrderUpdate → SubmitBracketOnEntryFill). Submitting all three in one OCO
> group cancels the SL/TP the instant the Market entry fills … The entry stands
> alone (empty OCO); SL+TP get their own OCO pair."*

**D1 is therefore already shipped. Building it would have changed nothing.**

### What actually killed the stop on 09-06 [A]

Our own code cancels the bracket by hand. `HandleCancelOrder` (`:1657-1712`) is a
**two-part** cancel:

```
:1665-1683   workingEntries[signalId] → acct.Cancel(entry)        // part 1: the entry
:1685-1706   placedBrackets[signalId] → legs = {SlOrder, TpOrder}
             acct.Cancel(legs)                                    // part 2: the CHILDREN
```

Part 2 is unconditional. It does not ask whether the entry filled, whether a
position is open, or what state the legs are in.

At 23:37:02 the entry had already filled, so `workingEntries.Remove(signalId)`
had fired on the fill at `:1394` — part 1 found `working == null` and did
nothing. **Only part 2 ran**, and it cancelled the accepted stop 29554 and the
working target. That is why snapshot 8213 shows `-sl` `CancelSubmitted` and 8214
is empty.

OCO propagation was never involved. Provenance of part 2: `8e3a96c8`
("wave2 phase2: … cancel_order …").

**Consequence for this wave: the real D1 is not "separate the OCO groups" (done)
but "`HandleCancelOrder` cancels the ENTRY ONLY, never `placedBrackets`."** That
is a C# change, so it lands in the F2 half, not F1.

### The Go-side ordering defect, from the live log [A]

`data/nofx_*.log:70047` and neighbours, 09-06:

```
23:37:02  ✕ armed cancel REQUESTED (one_live_arm_guard): ASIA S1 leg 1 — pending broker confirmation
23:37:02  ⚡ armed fill S1: armed under v5 S1 LONG () · qty 1
23:37:02  ⚡ armed fill S1 @ 29576.00 (entry_class=armed_fill — stale_reeval NOT applied)
```

The cancel was requested **before** the fill was drained, in the same second —
the stale-ledger read the 09-06 four-part fix addressed.

---

## C2 — CONFIRMED, and stronger than stated [A]

The AddOn carries no connection-provider string (grep for `Rithmic`/`Tradovate`
in the order path: 0 hits). The account is NinjaTrader's **internal simulator**
(SIM-only, standing rule) — orders never leave the PC at all, so OCO here is
100% locally simulated, not merely "most functionality". Both consequences the
dispatch names hold: a sibling is not cancelled if NT8 is down when one leg
fills, and a stray local cancel reaches the whole group.

## C3 — (a) already correct · (b) already fixed · (c) REAL GAP [A]

| claim | verdict | evidence |
|---|---|---|
| (a) `Accepted` is a normal live resting state | already correct | `provider/ninjatrader/order_snapshot.go:46-58` — `Accepted` is **not** in `terminalOrderStates`, so `IsWorking()` returns true |
| (b) dying states must not be recorded as accepted | already fixed | `trader/accepted_risk_hook.go:124-128` refuses `cancelpending`/`cancelsubmitted`/`cancel_pending`/`cancel_submitted` (the 09-06 four-part fix, item 4) |
| (c) `TriggerPending` is local-only, not at the exchange | **NOT CLASSIFIED ANYWHERE** | grep across `*.go` + `*.cs`: **0 occurrences**. It falls through `IsWorking()` and reads as live at the broker when it is held on the PC |

Additional finding, not in the dispatch: `terminalOrderStates` contains
`"unknown": true`. An order whose state we cannot read is therefore classified as
*history*, i.e. not live. That is UNKNOWN taking the non-conservative branch
(A24). Whether it reaches a destructive path depends on the caller; flagged for
a ruling rather than asserted. **[B]**

## C4 — CONFIRMED as a gap [A]

`grep -n "Account.Orders\|lock (account"` in the AddOn: **0 hits**. The AddOn
never reads the broker's order collection. It relies entirely on three
in-memory mirrors — `pendingBrackets`, `workingEntries` (`:131`), `placedBrackets`
(`:127`) — which is exactly the cached-mirror pattern C4 warns against.

## C5 — CONFIRMED [A]

Protective orders are **`TimeInForce.Day`**: stop `:1865`, target `:1869`. The
entry is `Day` too (`:1004`). D6's change (GTC on the protective pair) is real
work.

## C6 — CONFIRMED [A]

`placedBrackets` is written in exactly one place — `SubmitBracketOnEntryFill`
at `:1877` — and nothing repopulates it from the broker. After an NT8 restart
the dictionary is empty, so:

- nothing rebuilds the stop/target pairing;
- `HandleModifyBracket` (`:1734`) and auto-breakeven (`:1800`) find no bracket
  and silently no-op;
- `HandleCancelOrder`'s part 2 finds nothing — which, ironically, is the only
  reason a post-restart cancel is currently safe.

---

## Two cancel senders that bypass the Go guard [A]

The 09-06 wave applied `cancelSafetyFor`/`cancelSignalIfSafe` to seven sites in
`trader/armed_executor.go` (`:368, :478, :531, :564, :893, :1088, :1099`). The
incident's own guard — `one_live_arm_guard`, `armed_executor.go:524-539` — **is**
covered. These are not:

| site | what it is | risk |
|---|---|---|
| `trader/one_contract.go:252` | `one_live_entry` sibling cancel — same class as the incident guard | sends `cancel_order` on a signal whose arm may have filled; C# part 2 then strips the bracket |
| `trader/position_desync.go:92` | class-27 orphan-bracket cleanup, cancels by `row.EntryOrderID` | if that id is the signal id, C# part 2 strips a live position's protection |
| `trader/armed_executor.go:2246` | test seam (`test-arm denied` / `test seam cancel`) | reachable only via the test endpoint; noted, not urgent |

---

## Revised plan put to the owner

Unchanged and still real: **D2** (bracket from the execution event's own
parameters), **D3** (one state vocabulary — `TriggerPending` is the actual gap,
plus the `unknown` ruling), **D4** (broker-sourced state before any cancel),
**D5** (reconnect/boot reconciliation — C6 confirms nothing rebuilds the
pairing), **D6** (GTC on the protective pair), **D7** (boot line + Guide +
SYSTEM-MAP).

Replaced: **D1** becomes *"`HandleCancelOrder` cancels the entry only; it never
walks `placedBrackets`"* — plus the two unguarded Go senders above.

Replaced: **E1** cannot fail on current code for the stated reason (the OCO ids
are already separate). The incident pin must instead assert that a `cancel_order`
for a signal whose entry has filled leaves the protective pair untouched — which
does fail today, in the C# half.

**A23: stopped here, reported, waited.** The owner ruled on 2026-09-07 and Part 2
is what was then built.

---
---

# PART 2 — WHAT SHIPPED

Owner ruling, 2026-09-07: *"GO with your replacements. D1 → HandleCancelOrder
cancels the ENTRY only; it never walks placedBrackets… E1 → a cancel_order for a
signal whose entry has FILLED leaves the protective pair untouched. Must fail
today."* Plus the two uncovered Go senders, `TriggerPending` classified LOCAL,
and the ruling that `unknown` is non-terminal.

Branch `fix/bracket-oco-separation`. Merged `origin/dev` in rather than rebasing
(A24 forbids a history rewrite; the branch was already pushed).

## The OCO assignment, before and after — BOTH SIDES

| | before this wave | after |
|---|---|---|
| ENTRY oco (C#) | `string.Empty` | `string.Empty` — **unchanged, it was already right** |
| STOP oco (C#) | `signalId + "-exit"` | unchanged |
| TARGET oco (C#) | `signalId + "-exit"` | unchanged |
| **what a cancel_order reached** | the entry **AND** `placedBrackets`' `SlOrder` + `TpOrder`, unconditionally | **the entry only**; `placedBrackets` is neither walked nor read |
| protective TIF | `TimeInForce.Day` | `TimeInForce.Gtc` |
| bracket quantity | `b.Qty`, cached at submit | `filledQty`, from `OrderEventArgs.Filled` |
| part-filled entry | **no bracket at all** | bracketed for the filled part, amended as more arrives |
| `Unknown` in the book | dropped by the AddOn before Go saw it | shipped |

Nothing in the Go tree assigns an OCO id; the AddOn owns it end to end. That is
why C1 could only be answered by reading the `.cs`.

## E1 — red, then green, then mutation-tested

Redefined per the ruling: *a cancel_order for a signal whose entry has FILLED
leaves the protective pair untouched.*

```
RED   TestCancellingAnEntryNeverTouchesItsBracket
      HandleCancelOrder still reaches placedBrackets
      HandleCancelOrder still reaches SlOrder
      HandleCancelOrder still reaches TpOrder

GREEN after D1

MUTATION (the bracket walk put back)
      --- FAIL: TestCancellingAnEntryNeverTouchesItsBracket
RESTORED
      ok  nofx/trader
```

The pin's first draft failed on **its own explanatory comment**, which is the
inverse of the ordering pin that a comment once satisfied. It now strips C#
comments and judges executable code only.

## The state vocabulary, before and after

| state | before | after | decided by |
|---|---|---|---|
| `Accepted` | not terminal → "working" | **LIVE** at the exchange | `ClassifyOrderState` |
| `Working` | not terminal | LIVE | |
| `Suspended`, `PartFilled` | not terminal | LIVE | |
| `Initialized`, `Submitted`, `Change*` | not terminal | **PENDING** — standing, but nothing agreed | |
| `TriggerPending` | **0 occurrences in the entire tree** — fell through as live | **LOCAL** — held on this PC, never protection | |
| `CancelPending`, `CancelSubmitted` | dying only in `accepted_risk`, 4 spellings inline | **DYING** everywhere, one definition | |
| `Filled`/`Cancelled`/`Rejected`/`Expired` | terminal | terminal | |
| `unknown` | **TERMINAL** — an unreadable order read as a closed one | **non-terminal**, and takes no destructive branch | owner ruling |
| an unrecognised string | not terminal (accidentally) | UNKNOWN, deliberately | |

Readers: `IsWorking()` = not terminal (the flat gate's question, conservative);
`IsLiveAtExchange()` = LIVE only (the reconciler's question); `IsStateReadable()`
lets a caller refuse to act. One bool could not answer both questions, which is
why the classification is graded rather than binary.

## The finding that made the UNKNOWN ruling real — CLASS 78

The Go reclassification was **unreachable**. The AddOn filtered first:

```csharp
if (st == "Filled" || st == "Cancelled" || st == "Rejected" ||
    st == "Expired" || st == "Unknown")
    continue;
```

An `Unknown` order was not shipped as unknown; it was not shipped at all. And
absence is what every "this order is gone" branch keys on — the stale reaper,
cancel settlement, `entryIsResting`, cutover leg 4, and the new D5 reconciler,
which would have placed a **second stop beside an invisible live one**. Every Go
test passed throughout. Found by reading the producer, not by a test. `Unknown`
is now shipped; the genuinely terminal four are still filtered.

## The cancel-sender census — four found, two guarded, two deliberately not

The 09-06 wave guarded seven sites in `armed_executor.go`, reviewing each alone.
The pin is now a census over every reference to the NT8 trader's `CancelOrder`:

| sender | verdict |
|---|---|
| `one_contract.go` `cancelOtherArmsInPlan` | **guarded** (owner-named) |
| `position_desync.go` `skipGateDesync` | **guarded** (owner-named), position-aware |
| `armed_executor.go` `cancelArmedOrdersSyncWith` | tried, **reverted** — see below |
| `class33_boot_sweep.go` `sweepPreBootArmsWith` | tried, **reverted** — see below |
| `armed_executor.go` `TestArmCancel` | exempt, test seam (owner: "stays") |
| `reconcileStaleWorking`, `confirmPendingCancels` | already guarded — they receive a closure that adjudicates (`armed_executor.go:1088`, `:1099`) |

**The two reverts are the honest part of this report.** I guarded two senders the
owner did not name. It turned 14 tests red, and the tests were right:
`cancelArmedOrdersSyncWith` is what flattens at session close and on a news halt,
and `sweepPreBootArmsWith` retires a dead process's orphans. The guard refuses
when there is no broker book — so on those paths a refusal leaves arms live into
an EOD flatten, and leaves the class-33 double-order of 2026-09-02 00:16 CT
alive. Refusing there is the more dangerous failure, and the boot sweep already
has its own answer to a missing link (it DEFERS without latching and retries).
Both are now named in the census's exemption list with that reasoning, so the
trade is visible rather than silent. **This is a deviation from the owner's
instruction and is put to him here.**

**Two shapes, one meaning apart.** Children-with-no-entry means opposite things:
with a position OPEN they are the protection (09-06); with the broker FLAT they
are orphans whose stop fires later and opens a naked position (class 27,
proven live at 26 minutes). So `adjudicateArmCancelWith` takes a
`positionContext` whose **zero value means "unknown" and is read as OPEN** —
ignorance keeps taking the non-destructive branch. A broker-confirmed FLAT
account short-circuits every refusal including the no-book one, because every
refusal exists to protect a live position and there is none.

## D5 — the reconciler, and the wire command that did not exist

`SetStopLoss` on the NT8 path writes a local map that a later `placeEntry`
reads. It puts nothing on the wire. `move_stop` needs a stop that already
exists; `modify_bracket` needs a live bracket. **Before 2026-09-07 this process
could not place a stop for an already-open position at all** — part of why 592
stayed naked: nothing was looking, and nothing could have fixed it if it had
been. `place_protective_stop` is new on both sides, gated on
`MinAddonBuildProtectiveStop = 2026-09-07-h1`.

Runs on the 1-minute `monitorTick` and the reconnect edge — **not** in
`maybeManageArmedOrders`, which returns early unless day_plan is enabled. A
position is unprotected regardless of day_plan.

Three A24 unknowns each say so and act on nothing: no/stale book · a protective
order in an unreadable state · a live stop the book carries no quantity for. A
PARTIALLY covered position is raised, never patched. Where no price exists in
either `accepted_risk` or the plan, a P0 is raised and nothing is placed.

## What is now impossible that was possible

1. **A cancel cannot reach a protective order.** `HandleCancelOrder` neither
   walks nor reads `placedBrackets`; the pin fails if it does either.
2. **A protective order cannot expire while its position lives.** GTC.
3. **A part-filled entry cannot sit unprotected.** It is bracketed for what
   filled, and amended.
4. **An unreadable order state cannot read as a closed one** — at either end.
5. **A locally-held stop cannot be counted as protection**, or written to
   `accepted_risk` as an accepted price.
6. **A position cannot be unprotected without anyone knowing.** Every minute,
   and on reconnect.

Still possible, and stated plainly (A15) — see the next section.

## A15 — what the owner will still see wrong

- **Until F2, the AddOn half is not live.** With the old AddOn running: a cancel
  of a filled entry STILL kills its bracket (D1 is C# code), protective orders
  are still Day, part-fills get no bracket, `Unknown` orders are still dropped,
  and `place_protective_stop` is REFUSED by the build gate — so D5 detects and
  raises a P0 but cannot place. The boot line will read
  `can-place-stop=no (addon 2026-09-05-g2 < 2026-09-07-h1)`. The Go-side cancel
  guards do hold from F1, which is the defence-in-depth that matters in that
  window.
- **The boot line reads mostly `n/a` at startup and that is correct** — those
  fields describe the AddOn and there is no book yet.
- **`guards.ts:51` still says the stop-entry seam is OFF "until a wave lands
  that confirms a broker cancel instead of assuming one."** The 09-06 wave landed
  that confirmation. Flipping the seam is an owner ruling and out of A31 scope,
  so the sentence is untouched — but it now reads as a promise already kept.
- **`GetOpenOrders` is served by the ledger, not the broker**, so the dead-man
  watchdog's "clean reconciliation" still proves nothing about the real book.
  Untouched by this wave (A31).
- **One commit message lost a fragment.** In `8e6cf957` bash consumed a backticked
  expression, so the line reads `(VLTraderTCPClient.cs, ).` The intended text is
  the filter quoted above. No history rewrite was made to fix it.
- **The class-75 SYSTEM-MAP contract test still does not exist** (filed as owed
  in the Wave A report; the map itself is updated).

## Rollback — BOTH halves

**Go (F1):** `nofx-bin.old.<rev it holds>` preserved beside the binary; restore
it, restore `deploy/RELEASE`, `kill -9` the pid, systemd relaunches. Everything
in this wave is additive except `HandleCancelOrder` (C#) and the two
`accepted_risk` narrowings; nothing in the Go half changes what is authored,
which conditions arm, the stop composition, the R:R floor, or any gate.

**NT8 (F2):** the `.cs` and `NinjaTrader.Custom.dll` are copied to
`~/nofx-backups/nt8-addon/` named by build id with md5s quoted BEFORE anything is
copied in (A13). To roll back: copy the backed-up `.cs` back, F5, full NT8
restart. Reverting the AddOn alone is SAFE with the new Go binary running: the
build gate then refuses `place_protective_stop` and D5 degrades to detect-and-
alert, which is exactly the F1-window behaviour described above.

## F2 proof (A20) — NOT YET RECEIVED

Owed, each with a timestamp, after the owner's separate GO in a flat window with
the book empty:

- the hello frame carrying `build_id=2026-09-07-h1`
- an entry order at the broker with its OWN (empty) oco id
- on its fill, a stop and a target with a shared DIFFERENT oco id, both `Gtc`
- a cancel of an unfilled entry leaving no orphan and touching no bracket
- a restart with an open position → the reconciler finds or places the stop
- an `Unknown` order appearing in an order_snapshot rather than being dropped
