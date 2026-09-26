# Independent adversarial review — FIX-DOUBLE-ENTRY (DS-102)

- **Reviewer:** DS-101 · READ-ONLY · 90-min timebox · reviewed head `7e9fd26c9` (branch `fix/double-entry-replay`, merge-base `35a53d29`; the branch moved once during review — this report is against `7e9fd26c9`).
- **Method:** source review of the full diff (`attempted_guard.go`, `tcp_server.go`, `order_snapshot.go`, `tcp_trader.go`, `VLTraderTCPClient.cs`, protocol doc, tests) + mutations run via `go test -overlay` in an isolated worktree (no file on the wave's branch was touched).

## Findings (severity-ordered)

### P1 — the guard decides from the WRONG account's book (threats 1+2)

`attempted_guard.go decideAttemptedEntry` step 2 reads `orderSnaps.LatestReceivedAny()` — the newest book among **all accounts** — and never consults `SignalPayload.Account`. Consequences:

- A fresh **EMPTY** book for account B orders a **RESEND** of account A's attempted frame even though A has no fresh book. If A's order actually reached NT8 and only A's book lags B's, the guard resends → **double entry on A**.
- Symmetrically, a fresh book for B containing a matching name would **SETTLE** A's frame — wrong in the other direction (A's live order un-settled → missed entry, the acceptable-but-unnecessary failure).

The per-account accessor `OrderSnapshotCache.LatestReceived(account)` already exists and is unused by the guard. The guard's own comment says `LatestReceivedAny` exists for "a frame's account … empty (legacy) or unknown at guard time" — but the code does not switch to the frame's account book when the account IS known.

**Live reproduction (overlay test, current head):**

```
--- FAIL: TestReviewAttemptedFrameDecidesFromItsOwnAccountBook
    zz_review_test.go:22: REVIEW FINDING: frame account=A resend-decided from
    account B's fresh book (decision=resend). A's own book is not fresh — the
    guard must HOLD until A's book arrives (or settle/drop after the wait),
    never resend on another account's evidence.
--- PASS: TestReviewAttemptedFrameResendsOnlyOnItsOwnFreshBook
```

**Fix (for DS-102):** when `sig.Account != ""`, decide from `LatestReceived(sig.Account)` (fresh empty → resend-once; not fresh → hold → drop after wait); keep `LatestReceivedAny` only for legacy empty accounts. Add both overlay cases above as pins.

### P2 — the echo-settle branch has NO test pin (threat 3)

Mutation M3 deleted the entire step-1 branch (`echoedAfter(sig.SignalID, recAt) → attemptedSettle`) — the fill-echo-during-hold path — and the **whole wave guard suite still passes** (`go test -run 'TestAttempted|TestFreshEntry' ./provider/ninjatrader/` → ok). The settle-on-echo is live code that nothing proves; the 10 s wait racing a fill frame is exactly the threat this branch exists for. **Fix:** one pin using the existing harness — reconnect, `noteEcho(signalID, after-reconnect)`, attempted frame → decision `attemptedSettle`, frame not resent.

### P3 — AddOn seen-set cap is soft; restart loses the set (threat 4)

- `VLTraderTCPClient.cs`: when `seenSignals.Count > SEEN_SIGNAL_CAP` the code removes **only expired** entries; with a high signal rate nothing expires for 10 min, so the set is bounded by TTL×rate, not by the cap. Suggest evicting oldest-first when `> CAP`. (Low practical risk at current rates.)
- An NT8 restart clears the in-memory set — acceptable: the Go guard does not depend on the far-side dedupe (documented), and Go-side echo memory is reconnect-scoped (`echoedAfter` uses `recAt`), so old echoes cannot settle a post-restart frame.
- `duplicate_ignored` misread-as-reject is **closed**: `tcp_trader.go` returns before the reject path and the pin `TestDuplicateIgnoredReplyNeverTriggersRearm` FAILs under M4 (early-return deleted) — non-hollow.

## Mutations run (go test -overlay) — non-hollow confirmed

| # | Mutation | Wave test that must fail | Result |
|---|---|---|---|
| M2 | snapshot-freshness gate removed (`recvAt.After(recAt)` dropped) | `TestAttemptedEntryStaleSnapshotTreatedAsNoSnapshot` | **FAILs** ✓ non-hollow |
| M3 | echo-settle branch deleted | (any) | **none fail** ✗ hollow — see P2 |
| M4 | `duplicate_ignored` early-return deleted (treated as normal fill) | `TestDuplicateIgnoredReplyNeverTriggersRearm` | **FAILs** ✓ non-hollow |
| M1 | (P1 reproduction) cross-account decision | `TestReviewAttemptedFrameDecidesFromItsOwnAccountBook` | **FAILs** on current head — the bug itself |

## Threat-vector sweep

1. **Order reached NT8 but snapshot lags/omits** — covered per account only if the frame's own book is consulted; today the wrong-book case resends (P1). Single-account deployment: correct (own book == the only book).
2. **Stale snapshot as fresh** — closed and pinned (M2).
3. **10 s wait racing fill/position** — echo settle is live but unpinned (P2); drop is fail-closed (missed entry, never double); recheck timer at wait+500 ms re-reads snapshots; `checkSignalAge` still gates every write (verified at enqueue and at the flush write).
4. **AddOn seen set** — TTL 10 min, never cleared on fill ✓; restart loses it but the Go guard is independent ✓; cap soft (P3); `duplicate_ignored` never re-arms ✓ pinned.
5. **market / armed limit / stop-entry / protective** — all three `SendSignal` callers are entries (market, armed, stop-entry); the stop-entry variant is pinned (`TestAttemptedEntryVariantsLimitAndStopEntry`). Protective SL/TP are placed by the AddOn on fill (C# comment) and never ride `SendSignal` → never blocked by the guard ✓.
6. **Error paths B1–B8** — settle/drop WARN with `op`, `trader_id`, `symbol`, `signal_id` (B3) ✓; drop counted `telemetry.IncGateBlock("attempted_entry_unverified")` ✓; env-parse failure WARNs once (B1) ✓; recheck flush error logged, not discarded (B1) ✓. `isEntrySignal` returns true unconditionally — safe today (only entries queue) but brittle if a future non-entry frame is queued; a one-line comment pinning that invariant would help (P3).

## Bottom line

The guard is fail-closed and the double-entry class is materially closed for the single-account deployment. **P1 must be fixed before merge if the AddOn can ever serve more than one account** (the wire already carries per-account routing: `SignalPayload.Account` + `snapKey(account)`). P2 is a one-test pin. P3 is polish.
