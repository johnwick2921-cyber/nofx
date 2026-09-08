# Placement truth — command clock, pre-send identity, received settlement

Recorded 2026-09-07 23:43 CT. Branch `fix/placement-truth-0907`, locked worktree
`/tmp/nofx-placement-truth-0907`. Implementation:
`7e0c5527f42bec9beee164773c279e6ce077e964` (includes `6262bf42`).

**[A] PROVEN in isolated tests; NOT DEPLOYED.** No production database writes,
Go restart, Windows file copy, NT8 compile/restart, or live order was performed.
The owner assigned this Go wave and explicitly deferred the C# producer.

## Source and scope

Cut from the current `origin/dev` at acceptance:
`e762476fd6e1b1791ebaa6a1dc3207b8bcbdeda9`. Remote claim `7b158c6e` was published
before implementation. A final fetch/merge of `origin/dev` reported already up
to date; the full suite ran on that integrated implementation.

Source freshness at acceptance (`git log -1 <base> -- <path>`):

- Protocol: `c84bd247 feat(F12): the AddOn emits its order book; leg 4 asks the broker; build_id is proven by receipt`.
- Audit checklist: `7ddf73cd docs(report+checklist): session calendar report and class 88 — a liveness signal that is a side effect of activity`.
- Timing audit: `e762476f docs: pin timing audit references and preserve concurrent F2 evidence`.

The latest owner ruling is authoritative: command creation timestamp;
`place_pending` on all four placement paths; received-only promotion; optional
Go rejection reason now, C# next wave. The prior audit is context, not fresh live
proof. Checklist R1–R10, class 81 and the production-call-site pins apply. The
class-81 section now includes this placement recurrence; no new class number was
allocated.

## Implemented behavior

All source coordinates below refer to implementation `7e0c5527`.

1. **Creation clock:** `trader/ninjatrader/tcp_trader.go:470` stamps the limit
   payload with current UTC at command creation, using millisecond precision.
   Market entries (`:401`) and stop entries (`:547`) use the same rule.
   `feedNowUTC` remains a distinct market clock; no bar is made fresh by this
   change. Existing market-data and entry gates remain in force.
2. **Payload age:** `provider/ninjatrader/tcp_server.go:2147` parses and checks
   the actual payload timestamp. It runs before enqueue (`:1057`) and after
   acquiring the writer, immediately before writing (`:2181`). Older-than-60s,
   missing, malformed and future timestamps are refused; logs name signal ID
   and measured age/threshold, or age unavailable. Retries retain the original
   payload and enqueue metadata. A newly written arm row or fresh retry time
   cannot conceal an old payload clock.
3. **Identity before wire:** `store/armed_orders.go:391` atomically stores signal
   ID and `place_pending`, conditional on an unsent `armed` row. Registration
   failure prevents the send. All four paths call it before the wire:

   | Path | Pre-send registration |
   | --- | --- |
   | Normal limit | `trader/armed_executor.go:1067` |
   | Normal stop | `trader/armed_executor.go:1366` |
   | Debug limit | `trader/armed_executor.go:2175` |
   | Debug stop | `trader/armed_executor.go:2232` |

   There is no post-send `working` write. Debug rows are created before sending
   and read back afterward, preserving a reply that beats the call's return.
4. **Received truth:** `trader/armed_executor.go:1706` settles entry rejections;
   `:1744` promotes on a received live entry state using the existing shared
   classifier. Protective-leg updates cannot settle the entry. h1's pre-submit
   rejection arrives as `fill.status=rejected`, handled separately at
   `trader/ninjatrader/tcp_trader.go:227`. Both paths persist through
   `store/armed_orders.go:410`. Late acceptance cannot resurrect a rejected row
   or erase `cancel_pending`.
5. **Reason:** optional `reason` fields are accepted on `fill` and `order_update`
   (`provider/ninjatrader/tcp_framing.go:74`, `:161`). Supplied nonblank text is
   stored verbatim, including surrounding spaces. Missing/blank means
   `reason unavailable (NT8 frame omitted reason)`. A later received reason can
   fill that absence without replacing an already received reason. Logs also
   expose supplied/unavailable reasons. No generic `NT8 reject` is stored.
6. **Pending lifecycle/display:** pending rows remain nonterminal and cannot be
   rewritten or used to mint a replacement. Boot sweep/cancel eligibility and
   split-leg handling include them. The UI distinguishes order authorized,
   placement pending, received working, and rejected with reason. Guide content
   and its revision marker accompany this wave.

## Pins and validation

**[A] Fresh final loopback measurement:** the actual entry composer was given prices from a
1-minute bar whose close was 30 minutes old (bar open 31 minutes old). The
receiver parsed the emitted payload timestamp and independently subtracted it
from receipt time, as h1 does. Results from `/tmp/placement-clock-proof.log`:

| Entry | Rejection reason fixture | Payload age at receiver |
| --- | --- | --- |
| Limit | h1 omitted field | 7.929 ms |
| Limit | supplied verbatim | 7.822 ms |
| Stop | h1 omitted field | 9.340 ms |
| Stop | supplied verbatim | 9.011 ms |

These are **loopback measurements, not live NT8 receipts**. Fresh arm rows were
present in every fixture: checking those alone would have missed the original
715s and 1,824s incidents.

- `TestFourPlacementPathsWaitForEntryReceipt`: all four production paths send a
  real loopback frame and remain pending; protective updates do not promote;
  received entry working promotes; rejection survives later working.
- `TestStopPlacementFastRejectBeforeSendReturns`: deterministic registration →
  received rejection → send-return interleaving; rejection survives.
- Actual inbound fill-frame routing, missing/verbatim reasons, registration
  failure, duplicate placement/re-authorization refusal, cancellation intent,
  reason enrichment, stale payload with fresh queue metadata, and expired retry
  are covered. Existing long/short stop wire twins also pass in the full suite.
- `go test ./...`: PASS on integrated `7e0c5527`;
  `/tmp/placement-merged-go.log` (trader package 95.881s).
- Targeted `go test -race`: PASS for store, provider, TCPTrader and executor
  regressions; `/tmp/placement-race-final.log`.
- Web: 49 files / **365 tests PASS**; production build PASS.
- Clean-clone Go build PASS; `git diff --check` PASS.

The race run first exposed a pre-existing lazy subscription-map initialization
race at the receipt router. `ensureRouters` now initializes all maps before
starting dispatch (`provider/ninjatrader/tcp_server.go:232`). The fill subscription
is installed synchronously before `NewTCPTrader` returns (`tcp_trader.go:173`),
and close-sync establishes its ledger handle before orders can be placed.
No timing sleep was used to conceal the race.

## Artifact and remaining live evidence

Undeployed candidate: `/tmp/nofx-placement-truth-candidate`, built from clean
clone `/tmp/nofx-placement-build/nofx` using `go build -buildvcs=true`.

```
vcs.revision=7e0c5527f42bec9beee164773c279e6ce077e964
vcs.time=2026-09-08T04:39:11Z
vcs.modified=false
sha256=6b005690cf53bcea0482d6ee55d23edd648827a84addc4033df44b47a268337d
```

The repo h1 source remains unchanged: MD5
`d0a604d79163f36557af89edc9f40777`. Its line 44 still states
`STALE_SIGNAL_AGE_SECONDS = 60`. This wave does not claim h1 emits the new reason
field or that Go has been cut over.

Unanswered placements retain `place_pending` and hold their slot. Expired queued
payloads are refused/logged locally; no broker rejection is invented and no
automatic replacement is minted. Existing cancellation/reconciliation policy
still governs unresolved rows. This is intentionally not a general rewrite of
cancellation or a promise of automatic queue-expiry settlement.

**EVENT-WAIT:** watcher snapshot **11119**, received **23:42:25.714 CT**,
`build_id=2026-09-07-h1`, zero orders. The four live h1 wire proofs remain pending:
entry owns its OCO; fill creates stop+target on another shared OCO; protective
legs carry Gtc; cancelling an unfilled entry leaves the bracket untouched.
The previously adopted g2 orders carried **Day**, not GTC; their earlier close
was recorded by the watch report. No new placement, OCO, or TIF claim follows
from an empty snapshot.
