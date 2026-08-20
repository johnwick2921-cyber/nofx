# P0 — WRONG PnL RECORDED: root cause, correction layer, class-killer (2026-08-20)

**Branch:** `hotfix/pnl-record-integrity` (base = `fix/fail-register-close` head `c28a1c9e`, deployed binary `4e32cd5a` at dispatch time).
**Deployed:** `17cd52e2` at **07:39:19 CT Aug 20** in a verified flat window (0 open rows). Boot: `🔐 BOOT INTEGRITY OK — rev 17cd52e21924 … goldens PASS`, preceded one second earlier by the correction pass: `⚖️ pnl-correction complete: 37 row(s) corrected additively (pnl_corrected + note; originals preserved) of 171 candidates since Aug 6.`

---

## 1. The trade — truth table [A]

Position **#526** — the ">$1K loss" that triggered this dispatch:

| source | side/qty | entry | exit | PnL |
|---|---|---|---|---|
| **Recorded** (DB row, before fix) | SHORT 1.0 | 29626.25 | 29660.964286 | **−$1,458.00** |
| **NT8 wire truth** (frame at 21:16:27 CT) | qty=**21** (whole-account flatten) | — | 29660.96 (21-lot average, non-tick) | −1458 across ALL 21 contracts |
| **Recomputed** (stored prices × ROW qty × $2/pt) | SHORT 1.0 | 29626.25 | 29660.964286 | **−$69.43** |
| **Corrected** (DB now) | — | — | — | `pnl_corrected = −69.43`, original −1458 preserved + note |

The row's own stored prices never supported −1458: (29626.25 − 29660.964286) × 1 × $2 = **−69.43**. The recorded number was **overstated ~$1,389 (21×)**.

## 2. Error class + root cause [A]

**Class: E6 — quantity attribution** (E4-adjacent). The owner manually flattened Sim101 at 21:16 CT Aug 19 (two parked LONG frames at 21:16:23/24 confirm manual activity). NT8 emits **one** `position_close` frame for the whole flattened size: `qty=21, exit=29660.96 (averaged), pnl=-1458`. `recordClose` computed PnL with the **frame's** qty against the bot's **1-lot** row.

**Root cause file:line:** `trader/ninjatrader/close_sync.go` — the PnL formula used frame `qty` instead of the owning row's quantity. Log fingerprint: `📕 NT position closed: MNQ SHORT qty=21 exit=29660.96 reason=manual pnl=-1458.00`. The reconcile fallback (`trader/ninjatrader/reconcile.go:185-198`) already used `row.Quantity` — clean, untouched.

**Corruption fingerprint:** a non-tick stored price (not a multiple of 0.25) = an aggregate-frame average. #526's exit 29660.964286 and #510's entry both carry it; their notes say so.

## 3. Scope [A] — SYSTEMIC historically, one row post-Aug-12

**37 of 171** CLOSED MNQ rows since Aug 6 disagreed with their own stored prices × row qty by >$0.50. Aug 9–12 = the known two-trader contamination era (same class: shared frames mis-attributed); **only #526 after Aug 12**. Top deltas (recorded − recomputed):

| id | side | recorded | corrected | Δ | exit (CT) |
|---|---|---|---|---|---|
| 526 | SHORT | −1458.00 | −69.43 | −1388.57 | 08-19 21:16 |
| 489 | SHORT | +1460.00 | +182.50 | +1277.50 | 08-11 12:13 |
| 459 | LONG | +859.50 | +0.25 | +859.25 | 08-10 20:29 |
| 410 | LONG | +775.00 | +35.38 | +739.63 | 08-10 02:35 |
| 389 | LONG | +500.50 | +50.05 | +450.45 | 08-09 19:10 |
| 413 | LONG | +397.00 | −4.67 | +401.67 | 08-10 04:08 |
| 484/462/414/458/460/461 | LONG | +80…+120 | −62…−27 | +137…+182 | 08-10/11 |

Note the sign: most historical corruption **overstated profits**; #526 overstated a loss. All 37 corrected additively; every note carries the recorded value, the delta, and the non-tick-average suffix where the stored price itself is suspect.

## 4. Corrected today-total + honest would-trip verdict [A]

Session-day (since 17:00 CT Aug 19), active trader, CLOSED excl. reconcile_flat, **5 trades**:

- **Raw recorded: −$1,518.50** → fed last night's soft-audit `daily loss would trip (realized today=-1589.00, limit=-450.00)` drama.
- **Corrected (now what every reader sees): −$129.93.**

**Verdict: the −450 daily-loss cage would NOT have tripped on true PnL.** The would-trip line was ~91% phantom — one corrupt row (#526) manufactured it. The guardrail telemetry itself behaved correctly on the data it was given.

## 5. What shipped (commits, one fix each)

| commit | change |
|---|---|
| `b9196b93` | **Writer fix** — `recordClose` attributes ONLY `owner.Quantity` (clamp ≥1); frame qty > row qty WARNs `⚖️ pnl-attribution …` as foreign/manual activity. + `pnl_attribution_test.go` pinning wrong-math=−1458 / right-math=−69.43 and distinguishing ticks-mixup and sign-flip classes. |
| `ca6e990f` | **Correction layer** — additive `pnl_corrected` + `pnl_correction_note` columns; `EffectivePnL()`; one-time flag-guarded `CorrectHistoricalPnL()` (WHERE-scoped, `pnl_corrected IS NULL` guard, originals NEVER edited) wired at boot; session-day guardrail SUM + full-stats SQL → `COALESCE(pnl_corrected, realized_pnl)`; close alert / outcome sign / watcher BackfillClose → `EffectivePnL()`. |
| `f024955a` | **Class-killer guard** — every close recomputes from stored prices × row qty × `market.FuturesPointValue`; Δ > $0.50 → WARN `⚖️ pnl_integrity MISMATCH …` to log_events + `pnl_integrity_mismatch` gate-block counter. Crypto (pv=0) skipped. |
| `5bc5c945` | fix-forward: missing telemetry import (the guard commit went out unbuildable — caught same minute). |
| `17cd52e2` | **Reader sweep** — loss-streak halt (`CountConsecutiveLossesSince`), full stats loop, RecentTrades, symbol/holding-time/direction stats, history summary + streaks all read `EffectivePnL()`. + `pnl_correction_test.go` (V2). |

**2.1 estimate-close ruling — no `pnl_source` column needed [A]:** the only close that lacks a real fill price (`reconcile_flat` orphan) records **pnl=0** and is ALREADY excluded from every aggregate ("unknown P&L"); the priced-reconcile path uses real parked fills with `row.Quantity`. No path writes a mark-price *estimate* into realized_pnl, so a source column would label nothing — `close_reason` already IS the source marker. Documented instead of built (YAGNI).

## 6. Verification

- **V1 [A]:** #526 shows `pnl_corrected=-69.43` + note + original −1458 intact (DB read post-boot); 37/37 corrected = exactly the Phase-1 scope set; session-day sum now −129.93 via the COALESCE reader that feeds the soft-audit/guardrail line.
- **V2 [A]:** `TestPnLAttributionMath` (frame-qty vs row-qty vs ticks-mixup vs sign-flip), `TestCorrectHistoricalPnL` (#526 shape: additive, note, original preserved, agreeing row untouched, idempotent flag, corrected session-day sum), `TestPnLIntegrityGuardMath/Wired`.
- **V3 [A]:** seeded $1 discrepancy trips the ±$0.50 guard (unit); exact record does not; the WARN line + counter are source-contract-pinned so a refactor can't silently drop them.
- **V4 [A]:** full `go test ./store/ ./trader/ ./trader/ninjatrader/ ./kernel/` green pre-deploy; 30-min post-deploy soak — see chat delivery for the soak table (zero pnl_integrity WARNs expected while no closes occur; any close during soak must match NT8 to the cent).

## PR

Base = `fix/fail-register-close` for a clean diff. Number in chat delivery (parsed from `gh pr create` output URL).
