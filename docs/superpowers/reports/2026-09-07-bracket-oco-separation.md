# BRACKET-OCO SEPARATION — Section C verification (MEASURE FIRST, A17)

**Status: STOPPED AT C1 — the wave's central premise is REFUTED. No code written.**
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

**A23: stopped here, reporting, waiting for the ruling.** Nothing has been built.
