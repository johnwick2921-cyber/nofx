# W-PICTURE-HTF — wave evidence (2026-09-20, final)

Branch `fix/picture-htf`, base `origin/dev` @ d7ca3846.
**Not merged, not deployed — status: implemented, awaiting verification and activation.**
Every claim below is scoped: code-level proof green; the strategy replay ran on
real stored bars; live SIM activation is the owner's next step.

## Status lines (dispatch protocol)

- **Implemented:** yes — detection (kernel), mode surface (store/knobs),
  opportunity ledger (store), event-driven evaluator + admission (trader),
  NT8 market entry + send-side re-check (trader/ninjatrader), live-bar fan-out
  (provider), AddOn evidence surface (C# `final`/`emitted_at`/rejection reason,
  build `2026-09-20-p1`), labeled replay harness (`cmd/picture_htf_replay`),
  UI mode selector + dashboard ledger, guide entry, boot line.
- **Tests passed:** see "Tests at the wave head" — Go 35/35, web 76 files /
  492 tests, race clean.
- **Active in SIM:** **no** — requires the owner's AddOn deploy (copy → F5 →
  full NT8 restart), merge to dev, and the owner-attended cutover (runbook:
  `docs/superpowers/runbooks/2026-09-20-picture-htf-activation.md`).
- **Natural qualifying opportunity observed:** none on the live tape (the mode
  has never been enabled). **Reported separately; remains pending.**
- **Entry and protective orders verified:** code + loopback frame pins yes;
  real NT8 fill: **not yet applicable** (mode gated off until the AddOn
  capability frame arrives).

## 1. Branch incorporates current dev (proof)

- `git fetch origin dev` at report time; `origin/dev` tip = **d7ca3846**
  (`release: 5cf53c56 live 2026-09-19 …`).
- `git merge-base --is-ancestor origin/dev HEAD` → **exit 0**: the wave head
  contains the entire current dev history (the branch was cut and rebased on
  d7ca3846; dev has not moved since).
- Full suites re-run at the wave head itself (below) — the same discipline as
  the merged-head law, applied to a branch that provably contains dev.

## 2. Tests at the wave head

- `go test ./... -count=1` → **exit 0, 35/35 packages**.
- `cd web && npx tsc --noEmit` → clean; `npx vitest run` → **76 files,
  492 tests, exit 0**.
- Race: evaluator + ledger concurrency surfaces `-race` clean.
- Wave pins: kernel 11 · store 5 (+12-way concurrency under `-race`) ·
  evaluator 10 · seam 6 (+ boot-line pin) · wire/loopback 4 · provider sink +
  labeled Sept-17 mechanism pin 5 · panel 2.

## 3. September 17 strategy replay (requested item 3)

Harness: `cmd/picture_htf_replay` (committed). It re-runs the SAME kernel
functions the live evaluator calls, on real stored bars: a read-only COPY of
the live `data.db` (2.27 GB, `sqlite3 .backup`), window
**2026-09-16 22:00Z → 2026-09-17 22:00Z** (the 09-17 CME session day), 30-day
context ladder for level detection (mirrors the live cache), 5m/1h/4h derived
from 1m on read, partial buckets dropped, levels rebuilt **as-of each H1
boundary** (time-faithful retirement). It never writes the ledger, never fans
out to the evaluator, never sends a frame — the label is printed on every run.
The mechanism proof that historical receipts cannot trigger live entries is
the SEPARATE pin `TestSept17ReplayNeverMintsOpportunities`.

**Tape:** 1,380 in-window 1m bars, contract MNQ 12-26 · 5m=276 · 1h=23 ·
4h=5 full in-window bars (context: 31,501 1m rows).

**Detected 4H levels (25 in force at window start, 26 surviving at end — full
inventory in the replay log):** representative entries —

| role | body | wick | source (Z) | status in window |
|---|---|---|---|---|
| support | 29352.25–29293.75 | 29399.75/29290.50 | 08-21 00:00 | retired 09-16 16:00 |
| resistance | 29500.00–29446.50 | 29539.75/29390.50 | 08-21 08:00 | retired 09-17 16:00 |
| resistance | 29647.00–29538.00 | 29655.75/29432.25 | 08-27 00:00 | retired 09-17 16:00 |
| resistance | 29682.25–29515.25 | 29720.00/29477.75 | 09-04 12:00 | retired 09-17 16:00 |
| resistance | 29742.50–29535.00 | 29764.75/29529.50 | 09-08 04:00 | retired 09-17 16:00 |
| resistance | 29620.00–29449.75 | 29686.00/29334.00 | 09-11 12:00 | retired 09-17 16:00 |
| support | 29016.25–28916.75 | 29036.50/28905.00 | 09-14 04:00 | **active all window** |
| support | 29278.00–29252.00 | 29302.75/29208.75 | 09-15 16:00 | **active all window** |
| resistance | 29500.00–29415.50 | 29552.50/29364.00 | 09-16 12:00 | retired 09-17 16:00 |
| support | 29499.75–29254.50 | 29538.75/29052.75 | 09-16 16:00 | **active all window** |

**H1 confirmations:** 23 completed in-window H1 boundaries evaluated against
levels active at each boundary — **3 one-tick close breaks fired** (all long):

1. 09-17 **07:59:59Z**: prev 29468.75 → new 29517.50, break long over
   resistance **29500.00**.
2. 09-17 **11:59:59Z**: prev 29585.50 → new 29620.75, break long over
   resistance **29620.00** (2 levels crossed — the extreme wins).
3. 09-17 **15:59:59Z**: prev 29680.00 → new 29743.25, break long over
   resistance **29742.50**.

**Geometry + verdicts** (min R:R = the resolved picture default **2.5**):

| # | break | entry ref | swing stop (adj) | opposing zone | R:R | verdict |
|---|---|---|---|---|---|---|
| 1 | 29500.00 long | 29528.75 | 29376.00 (29375.75) | 29604.50 | 0.50 | **refused** — rr_below_min |
| 2 | 29620.00 long | 29619.00 | 29565.00 (29564.75) | 29742.50 | 2.28 | **refused** — rr_below_min |
| 3 | 29742.50 long | — | — | none above (top of range) | — | **refused** — no_opposing_zone |

At the live configuration floor (`ai_config.risk_control.min_risk_reward_ratio`
= **2.0** on the bound strategy a5b7662e-7bf7…, read from the DB copy;
`day_plan.picture_htf` absent → resolved defaults): **#2 becomes ELIGIBLE** —
entry ref 29619.00, stop 29564.75, target 29742.50, R:R 2.28. Result:
**eligible=1, refused=2** (rr_below_min ×1, no_opposing_zone ×1).

**Outcome of the one eligible setup on the real tape (informational, not a
claim of a fill):** from the 12:00Z interval, target 29742.50 was first
touched at **12:32Z** (+123.50 pts); the stop 29564.75 was **never** touched
(min 1m low after 12:00Z = 29596.00). Target-first, no stop intrusion. Entry
would be a market fill near the 12:00 5m close (slippage not modeled).

**Replay limitations (stated honestly):** stored 1m rows only (NT8's own
higher-TF bars were not re-derived from its ladder); level detection sees the
STORED tape, not the live cache; freshness is simulated (an eligible setup
still needs a fresh live frame + the send-side re-checks to actually submit).

## 4. AddOn backup/copy + cutover procedure

Exact commands and the owner-attended cutover steps:
`docs/superpowers/runbooks/2026-09-20-picture-htf-activation.md` — backup the
running AddOn sources first, copy `ninjascript/*.cs` to the NT8 AddOns folder
with md5 verification, F5 compile, FULL NT8 restart, merge at the merged-head
suite, five-leg flat gate, clean-clone build (vcs stamp), swap + `kill -9`,
boot-line read (`📷 picture-htf: …`), post-boot marker before lock release.

## 5. After activation — evidence still required (pending)

1. Received AddOn build/capability evidence (heartbeat `build=2026-09-20-p1`).
2. Actual Go boot/readiness output (integrity line, picture boot line, UI
   bundle match).
3. Selected strategy mode + saved knob values (quoted).
4. Controlled SIM entry + protection receipts (opportunity row, signal frame,
   received order_update/fill frames with bracket legs) — from a natural
   qualifying setup; the loopback pins are the pre-activation proof.
5. Natural-market setup evidence — **reported separately; none yet, pending.**

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

## Verification table (file:line)

| Requirement | Production call site | Proof |
|---|---|---|
| 4H body pivot, strict 2+2 neighbors, no lookahead, retirement | `kernel/picture_htf.go:56` `BodyPivots4H` | 11 kernel pins + replay levels |
| Level knowable only after the far-edge candle opens | `kernel/picture_htf.go:141` `ActiveLevels` | kernel pins |
| H1 close break, one tick, extreme crossed level wins | `kernel/picture_htf.go:199` `H1CloseBreak` | kernel + evaluator pins + replay breaks |
| 5m structural swing, strict wicks, deadline-gated | `kernel/picture_htf.go:251` `StructuralSwing5M` | kernel + evaluator pins + replay geometry |
| Nearest opposing zone, role-respecting, nearer never skipped | `kernel/picture_htf.go:283` `NearestOpposingZone` | kernel + evaluator pins + replay (break #3: no zone above → refused) |
| Advisory momentum stall | `kernel/picture_htf.go` `H1MomentumStall` | kernel pin |
| Mode surface + rule validation | `store/strategy.go:926` `PictureHtfResolved`, `kernel/entry_law.go` `ValidatePictureHtfRule` | store pins |
| Durable uniqueness: one row per opportunity | `store/picture_htf.go` `PictureHtfOppKey`/`PictureHtfClaim` | store pins incl. 12-way `-race` |
| Exactly one execution owner | `store/picture_htf.go` `PictureHtfClaimSubmission` (atomic `UPDATE … WHERE stage='confirmed' AND signal_id=''`) | 12 claimers, 1 winner, `-race` |
| Broker signal stamped only by the owner | `store/picture_htf.go` `PictureHtfStampSignal` | store pin |
| Event-driven evaluation from native bars | `trader/picture_htf_evaluator.go` `OnBars`/`Evaluate` | 10 evaluator pins at the production call sites |
| Late frame can never enter | `trader/picture_htf_evaluator.go` `evaluateLocked` freshness + window | `LateFrameExpires`, `PastWindowExpires` |
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
