# W-PICTURE-HTF — end-to-end verification (2026-09-20, final)

Branch `fix/picture-htf`, base `origin/dev` @ d7ca3846.
**Not merged, not deployed. Cutover HELD pending CTO review of this evidence
and owner-attended NT8 steps.**

This revision supersedes all earlier revisions. Two PRODUCTION corrections
landed this round (c79fbe0b) — the implementation differed from the agreed
behavior; the strategy itself is unchanged. All earlier replay claims were
already recomputed (e55fea50) and are restated here with the sequence proof.

## 1. Reviewed SHA + incorporated dev

- Wave head (this report's subject): **c79fbe0b** (code) — reviewed commits
  c79fbe0b, e55fea50, 0dedc0e8, 939b4507, 5dbf59eb, 845672f9, 1fa15ccb,
  4e8f7931, 38585854, 42c35e2d, efddb329, 1b8bcc8e, f33eb7da, f1099f4b,
  17f9fda6, f92dbb80, 44aa5248, ceb894b9.
- Incorporated dev: `d7ca3846cf8df7610f09162be7583c05b7a85f47`
  (`git merge-base origin/dev HEAD` == the dev tip → the branch contains all
  of current dev; dev has not moved).

## 2. THE TRADING BEHAVIOR — requirement → file:line → test → result

Legend: **[pin]** automated production-path test · **[replay]** Sept-17
historical replay evidence · **[audited]** code-level audit only ·
**[missing]** not yet evidenced (see §7).

| # | Requirement (spec) | Production call site | Evidence | Result |
|---|---|---|---|---|
| L1 | Levels from completed native NT8 4H candle BODIES (top=max(o,c), bot=min(o,c)); wicks kept separately | `kernel/picture_htf.go:56` BodyPivots4H; replay loads native stored bars | [pin] TestBodyPivots4HFindsBothPivotsAndNoLookahead · [replay] 21 levels in force on 09-17 | PASS |
| L2 | Strict 2+2 pivot confirmation; no equal-value plateaus | `kernel/picture_htf.go` (4-neighbor loop, res∧sup→skip) | [pin] TestBodyPivots4HEqualBodiesDoNotFormPivot | PASS |
| L3 | Level usable only after BOTH following candles complete | `kernel/picture_htf.go:141` ActiveLevels (KnowableAt) | [pin] TestH1CloseBreakOneTickRule (not-yet-knowable level must not fire) | PASS |
| L4 | 120-bar discovery window; retirement rules | `kernel/picture_htf.go` (window cap + far-edge-close retirement) | [pin] TestBodyPivots4HRetiresOnFarEdgeClose · [replay] 21→23 level evolution | PASS |
| R1 | 4H wick below support + close back above = bullish rejection context; mirror resistance; ADVISORY only | `kernel/picture_htf.go:158` SweepReclaim4H | [pin] TestSweepReclaim4HSupportRejection | PASS (advisory; never a prerequisite — no gate reads it) |
| H1 | Level known before the confirming H1 opened | H1CloseBreak filter `KnowableAt > newH1.OpenTime → skip` | [pin] TestH1CloseBreakOneTickRule | PASS |
| H2 | Long: prev close ≤ resistance; new COMPLETED close ≥ resistance + 1 tick; mirror short | `kernel/picture_htf.go:199` H1CloseBreak (Close only) | [pin] TestH1CloseBreakOneTickRule + TestH1CloseBreakMirroredShort | PASS |
| H3 | Forming or wick-only crossing does NOT qualify | H1CloseBreak uses Close; wick high is never read | [pin] TestH1CloseBreakOneTickRule ("wick-only crossing must not fire") | PASS |
| H4 | Multiple crossed levels: highest resistance (long) / lowest support (short) wins; rest ride as context | `kernel/picture_htf.go` sort + `res.Crossed` | [pin] TestH1CloseBreakExtremeLevelWins | PASS |
| H5 | Simultaneous H1/4H completion: preserve the pre-close breakout reference; apply retirements before target selection | `trader/picture_htf_evaluator.go:188-214` breakLevels (4H bars completed BEFORE cur.OpenTime) + as-of-now snapshot for the target | [pin] TestPictureHtfSimultaneousH1And4HCompletionPreservesBreakoutReference (both frame orders) | PASS — production corrected this round |
| H6 | Arrival-order independence | both snapshots time-derived from the cache; no frame-order state | [pin] same test, orders "4h-first" and "5m-first" | PASS — production corrected this round |
| E1 | Market entry at the START of the following native 5m interval; no wait for that candle's close / AI cycle / retest | `trader/picture_htf_evaluator.go` intervalStart = newest completed 5m open + 5m; submit seam runs in the same evaluation | [pin] TestPictureHtfH1CloseToNext5mSequenceNativeAlignment · [replay] per-break SEQUENCE block (interval opens == next native 5m bar's open) | PASS |
| E2 | Submission within the 10s window | elapsed vs EntryWindowSec gate | [pin] TestPictureHtfEvaluatorPastWindowExpires + SubmitsOnceAndAdmits (elapsed 0.5s) | PASS |
| E3 | 2-second live receipt freshness | freshest5mAt vs FreshnessSec | [pin] TestPictureHtfEvaluatorLateFrameExpires | PASS |
| E4 | Recheck immediately before send | `trader/picture_htf_send.go` pictureHtfSend (freshness/window/flat/pending/size) | [pin] TestPictureHtfSendRefusesStaleFeed / ClosedWindow / NilState / NonNTTrader | PASS |
| E5 | Late opportunity expires; historical backfill cannot trigger it | freshness + window gates; live-only sink | [pin] LateFrameExpires · TestDrainBarIngestHistoricalNeverFansOut · TestSept17ReplayNeverMintsOpportunities | PASS |
| S1 | Stop = latest qualifying confirmed 5m swing, strict 2+2 WICK pivot, preceding 24 COMPLETED bars, both confirmers completed by the H1 close | `kernel/picture_htf.go:251` StructuralSwing5M (strict wick comparisons, lookback 24, confirmDeadline) | [pin] TestStructuralSwing5MStrictAndDeadline | PASS |
| S2 | ONE tick beyond the swing extreme, SYMMETRICALLY (long −1, short +1) | evaluator long/short branches; harness mirrors | [pin] TestPictureHtfEvaluatorMirroredShortEndToEnd (stop == swing high + 0.25) | PASS — harness corrected earlier, production audited clean |
| S3 | No qualifying swing = refusal | evaluator refusal path | [pin] TestPictureHtfEvaluatorNoSwingRefuses | PASS |
| S4 | Never tighten the stop to manufacture R:R | no code path adjusts the stop after the swing; RR refusals keep the stop | [pin] TestPictureHtfEvaluatorLowRRRefuses (refusal, stop untouched) | PASS |
| T1 | Target = nearest ACTIVE opposing 4H body zone beyond entry; target its NEAR edge | `kernel/picture_htf.go:283` NearestOpposingZone (role-respecting, TargetEdge) | [pin] TestNearestOpposingZonePolarityAndNearest + LowRRRefuses (nearer zone never skipped) | PASS |
| T2 | No eligible target = refusal | evaluator refusal path | [pin] TestPictureHtfEvaluatorNoOpposingZoneRefuses | PASS |
| T3 | Enforce the ACTUAL saved strategy minimum + existing sizing/risk limits; 3R is a displayed reference only | evaluator minRR resolver (strategy → SafeDefault); seam sizing clamp | [pin] LowRRRefuses · TestPictureHtfContractSizeNeverExceedsClamp · [replay] both floors (2.5 default, live 2.0) | PASS |
| M1 | Completed-H1 close-progression momentum; advisory only; cannot reverse/flatten/cancel | `kernel/picture_htf.go` H1MomentumStall + evaluator records it on the row only | [pin] TestH1MomentumStallAdvisory | PASS (no gate reads it) |
| P1 | Durable unique opportunity row ≠ submission license; exactly-once send via atomic ownership | `store/picture_htf.go` PictureHtfClaim + PictureHtfClaimSubmission (UPDATE … WHERE stage='confirmed' AND signal_id='') | [pin] TestPictureHtfClaimSubmissionExactlyOneWinner (12 claimers, 1 winner, -race) · TestPictureHtfClaimDuplicateKeyRefuses | PASS |
| P2 | Reconciliation prevents a second send; never blindly retry an ambiguous command | place_pending blocks re-entry; seam reports "send ambiguous" and does not resend | [pin] TestPictureHtfAmbiguousSendStaysPending (re-entry refused at the store) · [missing] the automatic NT8-order reconciliation sweep is NOT implemented (§7) | PARTIAL |
| P3 | Received broker state moves the row (working/filled/rejected); absent fields stay empty | `store/picture_htf.go` PictureHtfMarkBroker | [pin] TestPictureHtfLifecycleAndBrokerEvidence · [missing] no live consumer of order_update frames for picture rows yet (§7) | PARTIAL |
| P4 | Restart after claim / after send / before ack — no double send | durable claim + atomic ownership | [pin] TestPictureHtfRestartAfterClaimSingleSubmission + ClaimSubmission pin | PASS |

## 3. Production path trace (native frame → broker state)

```
NT8 AddOn bar_update (final+emitted_at)      → C# VLBarsSubscriptionManager boundary finalization
 → Go BarCache.Upsert (provider/ninjatrader/tcp_server.go drainBarIngest)
 → fanOutLiveBars (LIVE frames only)          → trader pictureHtfLiveBars → AutoTrader.NotifyLiveBars
 → PictureHtfEvaluator.OnBars/Evaluate
     → capability gate (build ≥ 2026-09-20-p1, proven by receipt)
     → rebuildLevels (4H body pivots, final-marked completed bars)
     → H1 completion scan + advisory momentum
     → BREAK check: pre-close level snapshot (bars completed before the H1's open) + one-tick close rule
     → following 5m interval: intervalStart = newest completed 5m open + 5m; 10s window; 2s freshness
     → geometry: 5m swing stop (±1 tick) + nearest opposing 4H zone; R:R vs the saved minimum
 → shared gate: PictureHtfClaim (unique row) → PictureHtfClaimSubmission (atomic owner)
 → seam re-checks (freshness/window/flat book/no unreconciled pending/1-contract clamp)
 → TCPTrader.MarketEntryWithProtection (SIM + bound-account + B3 guard + identity assert)
 → NT8 executes; received order_update/fill frames carry rejection reason
 → PictureHtfMarkBroker records RECEIVED broker state (consumer NOT yet wired — §7)
```

## 4. Adversarial coverage (production-path pins)

| Case | Pin | Result |
|---|---|---|
| Forming candle / unfinalized bar | TestPictureHtfEvaluatorIgnoresUnfinalizedBars + kernel TestBodyPivots4HFormingCandleCannotBePivot | PASS |
| Wick-only crossing | kernel TestH1CloseBreakOneTickRule | PASS |
| Late frame | TestPictureHtfEvaluatorLateFrameExpires | PASS |
| Historical frame | TestDrainBarIngestHistoricalNeverFansOut + TestSept17ReplayNeverMintsOpportunities | PASS |
| Duplicate frame/restart | SubmitsOnceAndAdmits + RestartAfterClaimSingleSubmission | PASS |
| Reordered frames (H1/4H vs 5m arrival) | Simultaneous…PreservesBreakoutReference (both orders) + freshness-watch (no 5m frame → watch) | PASS |
| Missing 5m boundary frame | PastWindowExpires (window elapses fail-closed); zero-frame case now WATCHES, never poisons | PASS |
| Session boundary / halt | same window-expiry path; replay uses native session-aligned rows | PASS |
| Simultaneous H1/4H close | Simultaneous…PreservesBreakoutReference | PASS |
| DST / clock skew | epoch-ms arithmetic (no wall-time parsing); freshness = now − Go-side receipt time (same clock) | AUDITED |
| Mirrored short + exact one-tick buffer | TestPictureHtfEvaluatorMirroredShortEndToEnd | PASS |
| Missing stop / target / wrong-polarity target / insufficient R:R | NoSwingRefuses · NoOpposingZoneRefuses · kernel NearestOpposingZonePolarityAndNearest · LowRRRefuses | PASS |
| Competing old/new executors, same account/instrument | store TestPictureHtfClaimSubmissionExactlyOneWinner (12-way, -race) | PASS |
| Restart after claim / after send / before ack | RestartAfterClaimSingleSubmission · ClaimSubmission pins | PASS |
| Rejection / immediate fill / ambiguous send / partial fill | ambiguous: AmbiguousSendStaysPending · rejection reason on the wire (framing roundtrip) · **live consumption + partial-fill handling: MISSING (§7)** | PARTIAL |
| Missing/rejected protective orders + recovery | **MISSING (§7)** — documented behavior: ambiguous row blocks re-entry until reconciled; the automatic reconciler is not built | MISSING |

## 5. Evidence categories (kept separate)

- **A. Automated production-path tests** — §2 matrix; Go 35/35, web 76/492, tsc clean, race clean (commands in §6).
- **B. Historical strategy replay** — Sept-17, labeled, read-only (§6).
- **C. Controlled synthetic NT8 SIM execution** — loopback wire pins (market entry frame, bracket, stamp ordering, no-send on stamp failure/incomplete bracket). Loopback is NOT NT8 execution proof; it proves the Go-side command composition only.
- **D. Naturally occurring market opportunity** — **none. Pending.**

## 6. Corrected replay (Sept-17) + exact test commands

**Replay inputs:** read-only COPY of live `data.db` (2.27 GB, `sqlite3 .backup`);
window 2026-09-16 22:00Z → 2026-09-17 22:00Z; contract **MNQ 12-26**, enforced
in the WHERE clause for the whole 30-day context ladder; 5m/1h = STORED
NT8-native rows (session-aligned, grid opens at :00 from the ETH 22:00Z
session start); 4h = disclosed ETH-grid proxy from native 1h (no stored
native 4h exists); entry price = close of the 5m bar completed at the entry
instant; symmetric ±1 tick stop; per-break SEQUENCE proof printed.

**Results (identical at 2.5 default and the live 2.0 floor):** 21 levels in
force at start · 22 H1 completions · 2 breaks (07:59Z over 29479.25 refused
R:R 0.45 · 10:59Z over 29581.50 ELIGIBLE: entry 29585.50, stop 29561.25,
target 29767.00, R:R 7.48). Outcome on the real tape: **stop-first loss**
(stop touched 11:44Z, low 29555; target touched 17:00Z). Slippage not
modeled. The pre-correction claims were withdrawn.

**Limitations:** stored rows predate the final/emitted_at wave → receipt
freshness is SIMULATED, not measured; native 4h absent (proxy, disclosed);
no market-fill reconstruction.

**Exact commands** (all run at c79fbe0b, logs quoted):
```
go test ./... -count=1                          → /tmp/gofinal5.log  exit 0, 35/35 packages
go test ./trader/ -count=1 -race -run TestPictureHtf … (race surfaces) → clean
cd web && npx tsc --noEmit                     → clean
cd web && npx vitest run                       → /tmp/webfinal5.log 76 files / 492 tests, exit 0
go run ./cmd/picture_htf_replay --db /tmp/picture-htf-replay.db \
  --start 2026-09-16T22:00:00Z --end 2026-09-17T22:00:00Z [--min-rr 2.5|2.0]
                                               → /tmp/replay-final.log
```

## 7. Missing evidence and remaining owner decisions

1. **Live broker-state consumption** — no consumer yet routes received
   order_update/fill frames into `PictureHtfMarkBroker` for picture rows.
   Until it is built, post-submit states are not observable in the ledger
   automatically (the AddOn's rejection-reason field is on the wire and
   roundtrip-pinned; the consumer is the missing link). Owner decision:
   build it in this wave or record it as the activation gate.
2. **Automatic reconciliation sweep** — `PictureHtfPendingByTrader` exists and
   an ambiguous row BLOCKS re-entry (never blindly resent — pinned), but the
   sweep that reconciles place_pending rows against NT8 order snapshots is
   not implemented.
3. **Partial-fill handling** — not implemented (SIM market entries fill or
   reject; state honestly).
4. **Native 4h stored bars** — absent from the store (live cache only). The
   replay's 4h is a disclosed proxy. Owner decision: accept the proxy or
   import native 4h history.
5. **AddOn build/capability receipt** — requires the owner's copy → F5 → full
   NT8 restart; until then the boot line reads `addon=not proven`.
6. **Merged-HEAD/release checks + attended cutover** — pending per the held
   order; runbook: `docs/superpowers/runbooks/2026-09-20-picture-htf-activation.md`.
7. **Natural-market opportunity (category D)** — none; pending.

## 8. Rule defaults

`enabled=off` · `tick_size=0.25` · `pivot_window=120` · `swing_lookback=24` ·
`entry_window_sec=10` · `freshness_sec=2` · `min_rr` inherits the saved
strategy minimum (live bound strategy: **2.0**). Owner deviations: none
configured. Engineering defaults are agreed initial choices, not proven
optimums; a faithful implementation does not establish profitability.

## 9. Report pinning

This file is committed on `fix/picture-htf`; the raw URL pinned to the full
commit SHA with the byte count is quoted in the dispatch reply.
