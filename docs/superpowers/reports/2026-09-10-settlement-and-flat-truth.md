# Settlement and flat truth — Section G report

**Wave:** settlement-and-flat-truth (owner-pinned Section C)
**Branch:** `fix/settlement-and-flat-truth` · **Claim:** `settlement-554049f5/nofx-8e[88742a]`
**Spec:** `reports/2026-09-09-next-wave-basis-settlement-and-flat-truth.md`,
`git log -1` → **`1718cf75` 2026-09-09 14:28:22 -0500** (SPEC-FRESHNESS LAW),
measured at `954f11b1`, re-verified at `bd295804` — see
`2026-09-10-settlement-c-verification.md`.
**Environment (class 110):** Go suite run **12:43–12:49 CDT, inside the
12:00–13:30 lunch no-trade band**, on the branch's own worktree.

---

## The one sentence under four of the five

> **A word in our ledger was being written from something other than the
> broker's answer.**

`cancelled` from a timeout (C1), from a missing link (C3), from a fill (C4);
`flat` from a local table (C2). C5 is the guard that would have asked the broker,
deliberately absent by the owner's ruling.

The 2026-09-06 wave established the rule — *a cancel is confirmed by the broker's
book, never by the call returning.* These were the five places that still did not.

## C1 · An ack timeout is not a cancellation — FIXED

`trader/armed_executor.go:2198` wrote
`SetState(r.ID, "cancelled", reason+" (ack timeout — flatten proceeds)")`.

A TIMEOUT promoted the row straight to the word that frees the arm slot
(`UpsertArm`) and that cutover leg 4 counts, skipping the entire
`RequestCancel → cancel_pending → ConfirmCancel` lifecycle. **A row promoted on a
timeout can be replaced while its order still rests at the broker.**

Not hearing an ack is not evidence the order is gone — it is the absence of
evidence either way, and `cancelled` is the destructive branch here (A24). The
row now stays `cancel_pending`, keeps its slot, and the settlement pass owns it.

A test existed that named this exact defect and pinned a different function. It
is not a footnote to C1 — it is the reason C1 survived a wave built to end it,
and it has its own section below.

## C2 · "Flat" was a claim about our table — FIXED (read), NOT auto-closed

Five flatten reads, all `at.store.Position().GetOpenPositions(at.id)`:
`auto_trader_clock.go` **80 · 158 · 533 · 648 · 721**.

This codebase documents that table lagging the broker in BOTH directions
(`position_desync.go:18` and `:78`). One direction is merely noisy. **The other
is our table empty while the broker still holds a position** — the flatten
declares "book flat", returns, and the position rides the close with no stop
attached to it. The cutover gate already refuses one number here (leg 1 sqlite,
legs 2/3 broker, quoted separately); the flatten had only the sqlite leg.

`trader/flat_truth.go` adds the second reader. `ProvenFlat()` requires BOTH
readers present and BOTH at zero — one reader agreeing with itself is not
corroboration. `Why()` always renders both sources so a flat claim is auditable.

**The disagreement ALARMS; it does not auto-close.** Closing would mean issuing
exits for positions our store has no row for, through a broker-side close path
that has never run, invented at the close, from a reader we have just discovered
we disagree with. A24: UNKNOWN takes no destructive branch. It refuses the word
"flat", names what each reader said, and raises it for the owner.

## C3 · The no-link fallback wrote `cancelled` with zero broker contact — FIXED

`cancelArmedOrders` is reached exactly when `at.armedTrader()` is nil — the NT8
bridge absent, **which is precisely when resting orders are most likely to
outlive us.** It walked every non-terminal row and wrote `SetState(id,
"cancelled")` with no wire, no book, no state test, and its count fed the
operator-facing `🔒 EOD-FLAT: %d armed order(s) cancelled`. A flat claim with zero
broker contact behind it.

Placed rows now go to `cancel_pending`. Never-placed rows are still retired
truthfully — there is no broker outcome to be wrong about, and holding them would
strand an arm slot on an order that never existed. **The return is split into
`(retired, unsettled)`** so the flat line cannot print "we asked" as "we did", and
both operator lines now say UNCONFIRMED / NOT cancelled.

## C4 · A fill counted as a cancel — FIXED

`acked = !at.armedRowStillActive(ledger, r.ID)` means only *the row left the
non-terminal set*. **Every terminal state satisfies that, including FILLED.** A
limit that filled two seconds before the close was counted by `n++` and logged as
an order we cancelled — understating the book at exactly the moment the flatten
is about to claim it is empty.

The outcome is now READ (`store.StateOf`) and named: fills are counted
separately, logged as a POSITION, and never counted as cancels.

## C5 · The flatten is the last unguarded cancel site — RECORDED, NOT FIXED

`cancelArmedOrdersSyncWith` still contains **0** `cancelSafetyFor` calls, per the
owner's ruling of 2026-09-07, which stands:

> *"a guard that refuses without a book, applied to the EOD flatten and the
> news-halt sweep, leaves arms live and re-opens class 33 — worse than the defect
> it guards."*

The harm remains bounded by the FAR side: the deployed AddOn's
`HandleCancelOrder` cancels the resting entry and refuses to read
`placedBrackets`. **That bound lives in C#, not in Go.**

## THE TEST THAT NAMED THE DEFECT AND PINNED A DIFFERENT FUNCTION

**This is the finding that explains why the 2026-09-06 wave's settlement half
never ran**, and it is more valuable than any of the five fixes above.

`trader/cancel_settlement_test.go` opens:

```
// E3 — A CANCEL THAT CANNOT BE CONFIRMED IS NEVER PROMOTED.
//
// This is the branch the old code got wrong in the most consequential way:
// cancelArmedOrdersSyncWith wrote "cancelled" on ACK TIMEOUT with the reason
// "(ack timeout — flatten proceeds)". UNKNOWN took the destructive branch, and
// 'cancelled' is destructive here because it is what unlocks a replacement.
```

That header is **exactly right**. It names the function, the branch, the literal
reason string, and why it matters. Then the body calls
`at.confirmPendingCancels(ledger, fake, …)` — **the settlement pass, not
`cancelArmedOrdersSyncWith`.** It asserts the settlement pass never promotes an
unconfirmable cancel, which was already true and stayed true.

So the defect was: identified, written down in precise language, given a test
file of its own named after it — **and never touched by that test.** It passed on
every run for four days, including every run in this wave before the fix.

### Why this is worse than having no test

A missing test leaves a known gap. **This left a retired suspicion.** Anyone
grepping for whether the ack-timeout promotion was handled found a file named
`cancel_settlement_test.go`, a header describing their exact concern, and a green
result. Every signal available without opening the body said "covered". The
`(ack timeout — flatten proceeds)` string sat in the source the whole time.

It also explains the shape of the 09-06 wave's outcome, which the Section C basis
recorded as *"built ≠ wired, on the wave that was ABOUT settlement"*: the
lifecycle was built correctly, tested correctly, and the path that actually
retires rows at a close was never connected to it. The test made the
disconnection invisible by testing the built half while naming the unwired one.

### The general form

**A test's SUBJECT is the function its body calls, never the function its name or
its comment mentions.** Prose is not an assertion. This is class 53 —
*parity tests exercise production CALL SITES; a test that builds both sides'
inputs proves only self-consistency* — reaching the case where the test does not
touch the production call site at all.

**Probe, four questions:**
1. For every test whose name or header cites a specific function, does its body
   CALL that function? `grep` the body for the identifier in the header. A
   mismatch is the finding, and it takes one command per file.
2. Does the file's name promise coverage of a subsystem rather than a function?
   `cancel_settlement_test.go` promises the settlement of cancels; it delivers
   the settlement PASS. The gap between those two readings is where this lives.
3. Would the test fail if the named defect were reintroduced? Reintroduce it and
   see. That is the only question that distinguishes a pin from a description,
   and it is why every pin in this wave is mutation-verified.
4. Was the test written in the same wave as the fix it describes? Then it was
   written while the author knew what they MEANT, which is exactly when a body
   can drift from a header without anyone noticing.

**Law: a test that names a defect must be able to reproduce it.** If it cannot,
its header is documentation of an intention — and documentation that reads as
coverage is worse than silence, because it stops the next person looking.

## Pins — 9 new, all mutation-verified

`trader/settlement_truth_test.go`. Mutation matrix, **7 applied, 7 killed**:
ack timeout writes `cancelled` · a fill counts as a cancel · no-link writes
`cancelled` · a failed broker read folded into flat · `ProvenFlat` stops requiring
the broker · no link treated as asked-and-empty · (plus the C3 count split).

**Three existing tests asserted the OLD contract and were updated, not deleted** —
`TestSListEODFlatCancelAckTimeoutStillFlattens` and both `TestT1News…`. Each
test's real subject is unchanged and still asserted (the flatten proceeds; the
retry count; the cancel-before-close wire order). Only the ledger claim moved.

## Two process notes worth more than the code

**My first C2 design was untestable.** It called
`at.armedTrader().GetPositions()`, and `armedTrader()` type-asserts to the
concrete `*TCPTrader` — so no fixture could reach it. **Untestable safety code is
how the flatten came to have one reader in the first place.** It now goes through
`gatePositions()` (any broker impl, already panic-guarded per A10) with the
reader as a seam.

**My mutation harness reported a false SURVIVED.** One mutation broke the build;
`go test` then emitted no `--- FAIL` line, and my check read that as "the pin
missed". A build failure and a surviving mutant are indistinguishable to a grep
for failures. The harness now **withholds a verdict unless the build succeeds**.
Every verdict above is build-verified.
