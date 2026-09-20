# W-PICTURE-HTF — wave evidence (2026-09-20)

Branch `fix/picture-htf`, wave head **8c0ad05c** (code head 939b4507), base `origin/dev` @ d7ca3846 (rebased).
**Not merged, not deployed.** Every "works" claim below is scoped: code-level
proof is green; live SIM activation is the owner's next step.

## Status lines (dispatch protocol)

- **Implemented:** yes — detection (kernel), mode surface (store/knobs),
  opportunity ledger (store), event-driven evaluator + admission (trader),
  NT8 market entry + send-side re-check (trader/ninjatrader), live-bar fan-out
  (provider), AddOn evidence surface (C# `final`/`emitted_at`/rejection reason,
  build `2026-09-20-p1`), UI mode selector + dashboard ledger, guide entry,
  boot line.
- **Tests passed:** Go 35/35 packages (`go test ./...` exit 0 at 939b4507),
  web 76 files / 492 tests (vitest, exit 0). New wave pins: kernel 11, store 5
  (+ race concurrency), evaluator 10, seam 6, wire/loopback 4, provider sink +
  Sept-17 replay 5, boot line 1, panel 2.
- **Active in SIM:** **no** — not yet. Requires the owner's AddOn deploy
  (copy → F5 → full NT8 restart), merge to dev, and the owner-attended cutover.
  The honest terminal state is "enabled, awaiting setup".
- **Natural qualifying opportunity observed:** no live tape yet.
- **Entry and protective orders verified:** code + loopback frame pins yes;
  real NT8 fill: **not yet applicable** (mode is gated off until the AddOn
  capability frame arrives).

## What shipped (fix/picture-htf)

| # | commit | surface |
|---|--------|---------|
| 1 | ceb894b9 | claim + rebased acceptance base |
| 2 | 44aa5248 | kernel detection rules + pins |
| 3 | f92dbb80 | mode + rule surface (`PictureHtfConfirmRule`) |
| 4 | f1099f4b, 17f9fda6 | knob registry (dotted `day_plan.picture_htf.*` paths) |
| 5 | f33eb7da | opportunity ledger + atomic claim |
| 6 | 1b8bcc8e | atomic submission ownership (CTO corrections) |
| 7 | efddb329 | event-driven evaluator + admission (production call sites) |
| 8 | 42c35e2d | `MarketEntryWithProtection` + seam wiring + bar fan-out |
| 9 | 38585854 | AddOn evidence surface (Go half) |
| 10 | 4e8f7931 | C# AddOn: boundary finalization + rejection reason + build bump |
| 11 | 1fa15ccb | UI: studio mode selector + dashboard opportunity ledger |
| 12 | 845672f9 | guide entry + labeled Sept-17 replay + census/scope re-pins |
| 13 | 5dbf59eb | GUIDE_BUILT_REV bump |
| 14 | 939b4507 | boot line: mode, rule v1, SIM, data readiness, capability |

## Verification table (file:line)

| Requirement | Production call site | Proof |
|---|---|---|
| 4H body pivot, strict 2+2 neighbors, no lookahead, retirement | `kernel/picture_htf.go:56` `BodyPivots4H` | pins: 11 kernel tests |
| Level knowable only after the far-edge candle opens | `kernel/picture_htf.go:141` `ActiveLevels` | kernel pins |
| H1 close break, one tick, extreme crossed level wins | `kernel/picture_htf.go:199` `H1CloseBreak` | kernel pins + evaluator fixture |
| 5m structural swing, strict wicks, deadline-gated | `kernel/picture_htf.go:251` `StructuralSwing5M` | kernel + evaluator pins |
| Nearest opposing zone, role-respecting, nearer never skipped | `kernel/picture_htf.go:283` `NearestOpposingZone` | kernel + evaluator pins (low-RR refuses; never re-targets) |
| Advisory momentum stall | `kernel/picture_htf.go` `H1MomentumStall` | kernel pin |
| Mode surface + rule validation | `store/strategy.go:926` `PictureHtfResolved`, `kernel/entry_law.go` `ValidatePictureHtfRule` | store pins |
| Durable uniqueness: one row per opportunity | `store/picture_htf.go` `PictureHtfOppKey`/`PictureHtfClaim` | store pins incl. 12-way `-race` |
| Exactly one execution owner | `store/picture_htf.go` `PictureHtfClaimSubmission` (atomic `UPDATE … WHERE stage='confirmed' AND signal_id=''`) | 12 claimers, 1 winner, `-race` |
| Broker signal stamped only by the owner | `store/picture_htf.go` `PictureHtfStampSignal` | store pin |
| Event-driven evaluation from native bars | `trader/picture_htf_evaluator.go` `OnBars`/`Evaluate` | 10 evaluator pins at the production call sites |
| Late frame can never enter | `trader/picture_htf_evaluator.go` `evaluateLocked` freshness + window | `LateFrameExpires`, `PastWindowExpires` |
| Duplicate frame/restart single submission | same, durable claim | `SubmitsOnceAndAdmits` |
| Unfinalized bar cannot establish an interval | `trader/picture_htf_evaluator.go` `bars()` (Final-only) | `IgnoresUnfinalizedBars` |
| AddOn capability gate | `trader/picture_htf_evaluator.go` `pictureHtfCapabilityProven` vs `MinAddonBuildPictureHtf` | `CapabilityGateBlocks` |
| Live-only fan-out | `provider/ninjatrader/tcp_server.go` `drainBarIngest` → `fanOutLiveBars` | `DrainBarIngestHistoricalNeverFansOut` |
| Sept-17 replay receipts never claimed | same | `TestSept17ReplayNeverMintsOpportunities` |
| Market entry with protection (concrete type only) | `trader/ninjatrader/tcp_trader.go` `MarketEntryWithProtection` | loopback frame pins (market path, bracket, tick-rounded midpoint, stamp ordering, no-send on stamp failure/incomplete bracket) |
| Send-side re-checks | `trader/picture_htf_send.go` `pictureHtfSend` (freshness, window, flat book, unreconciled pending, 1-contract sizing) | seam pins |
| 19-method `types.Trader` interface untouched | `trader/types.Trader` compile check | full suite |
| Wire evidence fields | `provider/ninjatrader/tcp_framing.go` `Bar.Final`/`EmittedAt`, `OrderUpdatePayload.Reason` | roundtrip pin |
| C# boundary finalization | `ninjascript/VLBarsSubscriptionManager.cs` `OnBarsUpdate` | NOT compiled here — AddOns compile only inside NT8 (owner step) |
| Dashboard ledger | `api/handler_picture_htf.go`, `web/src/components/trader/PictureHtfPanel.tsx` | panel pins + tsc + vitest |

## Rule defaults (resolved)

`enabled=off` · `tick_size=0.25` · `pivot_window=120` (4H bars) ·
`swing_lookback=24` (5m bars) · `entry_window_sec=10` · `freshness_sec=2` ·
`min_rr` inherits `risk_control.min_risk_reward_ratio` (fallback 3.0).
**Owner deviations: none configured yet.**

## Test results (939b4507)

- `go test ./... -count=1` → exit 0, 35 packages.
- `npx vitest run` → 76 files, 492 tests, exit 0.
- Race: evaluator + ledger concurrency `-race` clean.

## Replay results + limitations

- Labeled Sept-17 replay pin: a historical tape that would be a perfect
  two-picture setup is RECEIVED (cache holds the bars) and its receipts are
  NEVER claimed. This proves the mechanism the dispatch demanded.
- Limitation: no live NT8 replay was run (that needs the new AddOn + NT8's
  replay channel). Until a natural live setup or a real NT8 replay fires, the
  end-to-end fill path is proven by loopback pins only — stated honestly.

## One complete opportunity trace (synthetic fixture, not a live fill)

- 4H ladder: old resistance 110 (i=1) → pullback → recent resistance 101 (i=7,
  knowable 09-14 09:00Z). H1 pair 09-14 17:00Z close 101.0 / 18:00Z close 101.5
  (crossing by one tick). 5m ladder ends 18:59:59.999Z; now = 19:00:00.5Z.
- Verdict: `submitted` once. Row: direction long, level resistance 101, stop
  98.25 (swing low 98.5 − tick), target 110, entry ref 101.49, R:R 2.63 vs
  configured 2.50 → stage `place_pending` awaiting broker evidence.
- A duplicate evaluation reports "opportunity already claimed" — zero second
  submissions.

## AddOn build/capability receipts

- **None yet** — the AddOn must be deployed by the owner:
  1. `cp /home/hoang/nofx/ninjascript/*.cs "/mnt/c/Users/hoang/Documents/NinjaTrader 8/bin/Custom/AddOns/"`
     (run from the DEPLOYED tree, not this worktree)
  2. F5 compile inside NT8
  3. full NT8 restart (AddOns do NOT hot-reload)
- Until NT8 reports build `2026-09-20-p1` on the heartbeat, the boot line reads
  `addon=not proven` and the mode refuses every evaluation.

## Outstanding evidence

1. AddOn F5 + restart receipt (owner).
2. Merge `fix/picture-htf` → dev + full suite at the MERGED head.
3. Owner-attended cutover on the deploy tree (the release chicken-and-egg
   pattern + five-leg gate), boot line read from the live log.
4. A natural qualifying opportunity (or a real NT8 replay) producing a row that
   reaches `working`/`filled` from RECEIVED broker frames.
