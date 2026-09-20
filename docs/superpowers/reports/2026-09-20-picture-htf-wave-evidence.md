# W-PICTURE-HTF — wave evidence (2026-09-20, corrected after CTO findings)

Branch `fix/picture-htf`, base `origin/dev` @ d7ca3846.
**Not merged, not deployed — status: implemented, awaiting verification and
activation. Cutover is HELD on the CTO's order.**

This revision REPLACES the replay claims of the previous report revision
(d6fe3341). The CTO's four replay-harness findings were all confirmed,
fixed in the harness (e55fea50), and audited against the PRODUCTION evaluator
first — the production path is clean on all four (see §4). The previous
Sept-17 outcome claim (target touched, stop never) is WITHDRAWN; the
corrected replay reaches the OPPOSITE outcome (stop-first).

## Status lines (dispatch protocol)

- **Implemented:** yes — detection (kernel), mode surface (store/knobs),
  opportunity ledger (store), event-driven evaluator + admission (trader),
  NT8 market entry + send-side re-check (trader/ninjatrader), live-bar fan-out
  (provider), AddOn evidence surface (C# `final`/`emitted_at`/rejection reason,
  build `2026-09-20-p1`), labeled replay harness (`cmd/picture_htf_replay`),
  UI mode selector + dashboard ledger, guide entry, boot line.
- **Tests passed:** Go 35/35 packages exit 0 · web 76 files / 492 tests exit 0
  · `tsc` clean · race clean — all re-run at e55fea50.
- **Active in SIM:** **no** — cutover HELD. Owner AddOn steps + CTO review of
  this corrected evidence required first.
- **Controlled synthetic SIM execution:** loopback wire pins + evaluator
  admission pins only — reported SEPARATELY in §5. It is not natural-market
  evidence and no such claim is made.
- **Natural-market execution evidence:** **none — pending.** The mode has
  never been enabled on a live tape.
- **Entry and protective orders verified:** code + loopback frame pins yes;
  real NT8 fill: not yet applicable (mode gated off until the AddOn
  capability frame arrives).

## 1. Branch incorporates current dev (proof)

- `origin/dev` tip = `d7ca3846cf8df7610f09162be7583c05b7a85f47` (fetched
  fresh); `git merge-base origin/dev HEAD` returns that exact SHA → the wave
  head contains all of current dev; dev has not moved.

## 2. Tests at the wave head (e55fea50)

- `go test ./... -count=1` → **exit 0, 35/35 packages** (`/tmp/gofinal3.log`).
- `npx tsc --noEmit` clean; `npx vitest run` → **76 files, 492 tests, exit 0**
  (`/tmp/webfinal3.log`).
- Race: evaluator + ledger concurrency surfaces `-race` clean.
- Wave pins: kernel 11 · store 5 (+12-way concurrency under `-race`) ·
  evaluator 10 · seam 6 (+ boot-line pin) · wire/loopback 4 · provider sink +
  labeled Sept-17 mechanism pin 5 · panel 2.

## 3. September 17 strategy replay — CORRECTED (e55fea50)

Harness `cmd/picture_htf_replay`, run on a read-only 2.27 GB copy of the live
`data.db`. Window **2026-09-16 22:00Z → 2026-09-17 22:00Z** (the 09-17 CME
session day).

**Data law applied (the CTO's four corrections):**
1. **Knowable entry price:** entryRef = close of the 5m bar COMPLETED at the
   entry boundary (knowable at the entry instant) — the previous revision used
   the entry interval's own bar close, a price known five minutes later.
2. **Native bars:** 5m and 1h ladders are the STORED NT8-native rows
   (session-aligned by NT8 trading hours). 4h is derived from the native 1h
   ladder on the ETH 22:00Z grid (4×1h, all four present) — a DISCLOSED proxy;
   no stored native 4h exists. UTC-modulo 1m aggregation is gone; the harness
   refuses eligibility claims outright when native 5m/1h rows are absent.
3. **Symmetric one-tick stop buffer:** long −1 tick, short +1 tick (the
   previous revision used +2 ticks for shorts).
4. **Contract purity:** the whole context ladder is `WHERE contract =
   'MNQ 12-26'` (the window's dominant contract) — no cross-contract timestamp
   merging anywhere.

**Tape:** 1m(window)=1,380 · native 5m=1,757 (context incl.) · native
1h=529 · 4h(proxy)=118, all contract-pure MNQ 12-26.

**Detected 4H levels — 21 in force at window start, 23 surviving at end**
(native 1h-derived ETH-grid 4h; full inventory in the replay log):

| role | body | wick | source (Z) |
|---|---|---|---|
| resistance | 29951.25–29892.25 | 29970.00/29889.00 | 08-20 02:00 |
| support | 29600.75–29418.00 | 29605.75/29385.50 | 08-25 22:00 |
| resistance | 29973.00–29894.00 | 30001.25/29894.00 | 08-28 02:00 |
| resistance | 29860.00–29503.50 | 29862.25/29475.25 | 09-01 06:00 |
| support | 29355.75–29260.00 | 29441.25/29218.50 | 09-02 10:00 |
| resistance | 29986.00–29961.25 | 30000.00/29800.00 | 09-04 10:00 |
| resistance | 30004.00–29919.50 | 30006.75/29821.25 | 09-07 22:00 |
| resistance | 29779.00–29703.75 | 29802.00/29668.00 | 09-10 02:00 |
| support | 29478.50–29387.25 | 29566.25/29387.25 | 09-10 14:00 |
| resistance | 29767.00–29607.25 | 29794.00/29500.00 | 09-11 10:00 |
| resistance | 29581.50–29292.75 | 29604.25/29246.50 | 09-14 14:00 |
| support | 29268.50–29229.75 | 29290.00/29225.75 | 09-15 22:00 |

**H1 confirmations:** 22 completed in-window boundaries, levels rebuilt
as-of each boundary — **2 one-tick breaks fired** (both long):

1. 09-17 **07:59:59Z**: prev 29468.75 → new 29517.50, break over resistance
   **29479.25**.
2. 09-17 **10:59:59Z**: prev 29558.25 → new 29585.50, break over resistance
   **29581.50**.

**Geometry + verdicts** (identical at the 2.5 default AND the live 2.0 floor):

| # | break | entry ref (knowable at entry) | swing stop (adj) | opposing zone | R:R | verdict |
|---|---|---|---|---|---|---|
| 1 | 29479.25 long | 29517.50 | 29376.00 (29375.75) | 29581.50 | 0.45 | **refused** — rr_below_min |
| 2 | 29581.50 long | 29585.50 | 29561.50 (29561.25) | 29767.00 | 7.48 | **ELIGIBLE** |

Result: **eligible=1 · refused=1** at both floors.

**Outcome of the one eligible setup on the real tape (informational — not a
claim of a fill; slippage not modeled):** the STOP 29561.25 was touched first
at **11:44Z** (min 1m low after 11:00Z = 29555.00); the target 29767.00 was
only touched at 17:00Z. **Stop-first loss.** This REPLACES the previous
revision's outcome claim (target touched 12:32Z, stop never), which was an
artifact of the corrected defects. The previous claim is withdrawn.

**Replay limitations (stated honestly):** 4h is a disclosed proxy (derived
from native 1h on the ETH grid — no stored native 4h); freshness is simulated
(an eligible setup still needs a fresh live frame + the send-side re-checks
to actually submit); the mechanism proof that historical receipts cannot
trigger live entries remains the SEPARATE pin
`TestSept17ReplayNeverMintsOpportunities`.

## 4. The four CTO findings — production evaluator audit (file:line)

The findings were replay-harness defects. The PRODUCTION evaluator was
audited against each before any fix was written:

| # | CTO finding | Harness (now fixed) | Production evaluator |
|---|---|---|---|
| (a) | entry uses a price known 5 min later while elapsed=0 | FIXED: entryRef = close of the 5m bar completed at the boundary | **CLEAN**: `trader/picture_htf_evaluator.go` — `newest5m` is the last bar with `CloseTime < nowMs` (completed BEFORE now), so its close is knowable at the entry instant; the entry interval's own forming bar is excluded by the same filter |
| (b) | UTC-modulo aggregation ≠ native session-aligned candles | FIXED: native stored 5m/1h rows; 4h disclosed ETH-grid proxy; refuses when native rows absent | **CLEAN**: the evaluator reads `market.FuturesBarsProvider` → `trader/ninjatrader/bars_market_bridge.go` `barsFromCache` — the live NT8 BarCache (AddOn session-aligned bars). No aggregation anywhere on the live path |
| (c) | short buffer +2 ticks | FIXED: symmetric ±1 tick | **CLEAN**: `trader/picture_htf_evaluator.go` — long `stopPx −= TickSize`; short `stopPx −= TickSize` then `+= 2*TickSize` → net **+1 tick** both sides |
| (d) | cross-contract timestamp merging (largest rowid) | FIXED: contract in the WHERE clause for the whole ladder | **CLEAN by construction**: the evaluator reads the cache keyed by the subscribed symbol; NT8 serves the platform's front-month contract; the opportunity key binds `currentContract(symbol)` (`trader/picture_htf_evaluator.go`, `OppKey`). No SQL timestamp merging exists on the live path |

## 5. Execution evidence — two SEPARATE tracks (not conflated)

**A. Controlled synthetic SIM execution (harness/pins — NOT market evidence):**
- Loopback wire pins (`trader/ninjatrader/market_entry_wire_test.go`): the
  market entry frame, bracket prices, tick-rounded midpoint, stamp ordering,
  no-send on stamp failure/incomplete bracket.
- Evaluator admission pins (`trader/picture_htf_evaluator_test.go`): the
  synthetic fixture trace (101 level, 110 target) — a CONSTRUCTED tape
  proving the admission sequence, not a market event.
- Sept-17 replay (§3): labeled, read-only, receipts not claimed.

**B. Natural-market execution evidence:**
- **None — pending.** No live setup has been observed (the mode has never been
  enabled). It will be reported separately when one occurs.

## 6. AddOn backup/copy + cutover procedure

`docs/superpowers/runbooks/2026-09-20-picture-htf-activation.md` — backup the
running AddOn sources first, copy `ninjascript/*.cs` to the NT8 AddOns folder
with md5 verification, F5 compile, FULL NT8 restart, merge at the merged-head
suite, five-leg flat gate, clean-clone build (vcs stamp), swap + `kill -9`,
boot-line read, post-boot marker before lock release.

**Cutover is HELD** until the CTO reviews this corrected evidence and the
owner is present for the NT8 compile/restart.

## 7. After activation — evidence still required (pending)

1. Received AddOn build/capability evidence (heartbeat `build=2026-09-20-p1`).
2. Actual Go boot/readiness output (integrity line, picture boot line, UI
   bundle match).
3. Selected strategy mode + saved knob values (quoted).
4. Controlled SIM entry + protection receipts (opportunity row, signal frame,
   received order_update/fill frames with bracket legs) — track A only.
5. Natural-market setup evidence — track B; none yet, pending.

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
| 15 | 0dedc0e8 | labeled historical replay harness (`cmd/picture_htf_replay`) |
| 16 | e55fea50 | corrected replay: native contract-pure bars, knowable entry price, symmetric buffer |

## Verification table (file:line)

| Requirement | Production call site | Proof |
|---|---|---|
| 4H body pivot, strict 2+2 neighbors, no lookahead, retirement | `kernel/picture_htf.go:56` `BodyPivots4H` | 11 kernel pins + replay levels |
| Level knowable only after the far-edge candle opens | `kernel/picture_htf.go:141` `ActiveLevels` | kernel pins |
| H1 close break, one tick, extreme crossed level wins | `kernel/picture_htf.go:199` `H1CloseBreak` | kernel + evaluator pins + replay breaks |
| 5m structural swing, strict wicks, deadline-gated | `kernel/picture_htf.go:251` `StructuralSwing5M` | kernel + evaluator pins + replay geometry |
| Nearest opposing zone, role-respecting, nearer never skipped | `kernel/picture_htf.go:283` `NearestOpposingZone` | kernel + evaluator pins + replay (nearer zone honored) |
| Advisory momentum stall | `kernel/picture_htf.go` `H1MomentumStall` | kernel pin |
| Mode surface + rule validation | `store/strategy.go:926` `PictureHtfResolved`, `kernel/entry_law.go` `ValidatePictureHtfRule` | store pins |
| Durable uniqueness: one row per opportunity | `store/picture_htf.go` `PictureHtfOppKey`/`PictureHtfClaim` | store pins incl. 12-way `-race` |
| Exactly one execution owner | `store/picture_htf.go` `PictureHtfClaimSubmission` (atomic `UPDATE … WHERE stage='confirmed' AND signal_id=''`) | 12 claimers, 1 winner, `-race` |
| Broker signal stamped only by the owner | `store/picture_htf.go` `PictureHtfStampSignal` | store pin |
| Event-driven evaluation from native bars | `trader/picture_htf_evaluator.go` `OnBars`/`Evaluate` | 10 evaluator pins at the production call sites |
| Entry price knowable at entry time | `trader/picture_htf_evaluator.go` — `newest5m` filtered `CloseTime < nowMs` | audited §4(a); harness corrected |
| Native session-aligned bars, no UTC aggregation | `trader/ninjatrader/bars_market_bridge.go` `barsFromCache` (live cache) | audited §4(b); harness uses native rows |
| Symmetric one-tick stop buffer | `trader/picture_htf_evaluator.go` long/short branches | audited §4(c); harness corrected |
| Contract isolation | AddOn front-month + `currentContract(symbol)` in `OppKey` | audited §4(d); harness contract-pure |
| Late frame can never enter | `trader/picture_htf_evaluator.go` freshness + window | `LateFrameExpires`, `PastWindowExpires` |
| Duplicate frame/restart single submission | durable claim | `SubmitsOnceAndAdmits` |
| Unfinalized bar cannot establish an interval | `trader/picture_htf_evaluator.go` `bars()` (Final-only) | `IgnoresUnfinalizedBars` |
| AddOn capability gate | `trader/picture_htf_evaluator.go` `pictureHtfCapabilityProven` vs `MinAddonBuildPictureHtf` | `CapabilityGateBlocks` |
| Live-only fan-out | `provider/ninjatrader/tcp_server.go` `drainBarIngest` → `fanOutLiveBars` | `DrainBarIngestHistoricalNeverFansOut` |
| Sept-17 replay receipts never claimed | same | `TestSept17ReplayNeverMintsOpportunities` |
| Market entry with protection (concrete type only) | `trader/ninjatrader/tcp_trader.go` `MarketEntryWithProtection` | loopback frame pins |
| Send-side re-checks | `trader/picture_htf_send.go` `pictureHtfSend` | seam pins |
| 19-method `types.Trader` interface untouched | `trader/types.Trader` compile check | full suite |
| Wire evidence fields | `provider/ninjatrader/tcp_framing.go` `Bar.Final`/`EmittedAt`, `OrderUpdatePayload.Reason` | roundtrip pin |
| C# boundary finalization | `ninjascript/VLBarsSubscriptionManager.cs` `OnBarsUpdate` | NOT compiled here — AddOns compile only inside NT8 (owner step) |
| Dashboard ledger | `api/handler_picture_htf.go`, `web/src/components/trader/PictureHtfPanel.tsx` | panel pins + tsc + vitest |

## Rule defaults (resolved)

`enabled=off` · `tick_size=0.25` · `pivot_window=120` (4H bars) ·
`swing_lookback=24` (5m bars) · `entry_window_sec=10` · `freshness_sec=2` ·
`min_rr` inherits `risk_control.min_risk_reward_ratio` (live bound strategy:
**2.0**; picture default if unset: 2.5 in the replay; code fallback 3.0).
**Owner deviations: none configured yet** (`day_plan.picture_htf` absent on
the bound strategy — enabling materializes the resolved defaults).
