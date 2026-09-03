# 2026-09-02 — LIVE intraday bias replay: the plan-stamping bias reconstructed and tested out-of-sample

**Dispatch:** the live intraday bias — "4H-dominant hierarchy (kernel bias tree + regime)" — gets the identical test as the bias calibration (2deab3c8, P7/D6). READ-ONLY on the engine · no lock · no live calls.
**Tree:** worktree `~/nofx-biastest`, branch `docs/live-bias-replay-0902`, base `bb8b5419` (dev HEAD at start). Scripts: `~/nofx-analysis/live-bias-replay/`. Bars source: the `bars` table (SQLite, read-only). CSVs: `docs/superpowers/reports/2026-09-02-live-bias-replay-csvs/`.

---

## PRE-REGISTRATION HEADER (fixed 2026-09-02 ~21:10 CT, BEFORE any computation)

### P0 — What "the live intraday bias" IS (quoted, no inference)

The plan's `Bias.Direction` is AI-authored; its MACHINE inputs are exactly two
blocks the planner prompt renders every session read:
1. **BIAS-TREE** — `kernel/planner_prompt.go:129 RenderBiasTree`: branch 1
   `close > PDH → bull (HIGH)`; branch 2 sweep of PDH/PDL + close back inside →
   bear/bull (MEDIUM); branch 3 inside the day → `close vs PDC → long/short
   (LOW)`; branch 4 outside-prior-range-but-inside → neutral; branch 5
   premium/discount disallows the premium side long / discount side short.
2. **REGIME** — `kernel/regime.go:50 ComputeRegime`: `trend_daily` = price vs
   EMA200 on daily bars, `trend_1h` = price vs EMA50 on 1h bars (up/down/flat,
   ±0.05% dead-band, `regimeTrendEps`).

**"4H-dominant hierarchy" note:** no literal 4h trend leg exists in the planner's
machine sections (grep-verified; the HTF-level pass uses 1h+4h but feeds the
level ladder, not the bias stamp). The reconstruction therefore covers exactly
the tree + regime legs, in the hierarchy the prompt presents them. Stated, not
hidden.

### P1 — Data inventory (measured, not looked-at-for-results)

`bars` table, read-only: 1d MNQ 2019-05-02 → 2026-09-01 (1,896) · 1h MNQ
2026-05-04 → 2026-09-02 (2,001) · 4h MNQ back to ~2025-05 (not needed by the
legs above; listed for completeness). The binding range for session-level
open/close targets is the 1h table.

### P2 — The reconstructed call (definitions fixed, no peeking after this header)

Unit = one session-plan instance = (session-day, session), sessions ASIA / LONDON
/ NY, over completed session days with 1h coverage (2026-05-05 … 2026-09-02).
Call evaluated at the session's READ time (engine registry `ReadCT`, quoted:
ASIA 16:30 · LONDON 01:30 · NY 08:00 CT) using ONLY bars closed before it:
- **price** = close of the last 1h bar with `open_time` < read time.
- **PDH / PDL / PDC** = prior session-day's 1d H / L / C (1d table).
- **tree call**: branch 1 → long; branch 1-mirror (close < PDL) → short;
  branch 3 inside → long iff price > PDC else short; branch 4 → neutral.
  Branch 2 (sweep + close back inside) and branch 5 (premium/discount veto)
  applied post-hoc as stated below — branch 5 veto: inside-premium (≥50% of
  [PDL,PDH]) disallows long, inside-discount disallows short → neutral.
- **regime call**: `up` iff price > EMA200_daily × 1.0005 AND price >
  EMA50_1h × 1.0005; `down` iff both < ×0.9995; else neutral. EMA50_1h requires
  50 prior 1h closes — the first days are `warming` and the leg is n/a → regime
  neutral (counted).
- **composite (the reconstructed stamp)**: tree call if long/short; else regime
  call; else neutral. Neutral rows are counted separately (called-only rows also
  reported).
- Legs reported individually too (tree-only, regime-only) for the D6 table.

**Target:** sign of the session's own window open→close from 1h bars:
ASIA 17:00→02:00 · LONDON 02:00→08:30 · NY 08:30→14:45 CT. `open` = close of the
last 1h bar with open_time < window start; `close` = close of the last 1h bar
with open_time < window end. **Stated intrabar error:** 1h stamps sit on the
hour; the NY "open" is the 09:00 print (includes 08:30–09:00) and the NY "close"
is the 14:00 print (excludes 14:00–14:45). Same error class as the bias
calibration's S2/S3.

### P3 — Split (fixed before looking)

Session days ordered; **exploration = first 60% (2026-05-05 … <cutoff), holdout
= last 40% (cutoff … 2026-09-02)** — the exact cutoff date computed from the
actual completed-day list and stated in the results. Verdicts from holdout only.

### P4 — Statistics (identical to the bias calibration P4/D2)

1. Hit rate `ĥ = #(call sign == realized sign)/n` on CALLED rows, 95% Wilson.
2. Mean session-window return (points) conditional on call sign, Student t vs 0.
3. Position series: +1/−1/0 per call, PnL in points, max drawdown.
4. **Net of friction:** 1 pt per side charged on |Δpos| at each re-position;
   net t-stat.
5. Neutral share quoted per leg.

### P5 — Multiplicity

Benjamini–Hochberg across the tested calls (tree-only, regime-only, composite,
× ASIA/LONDON/NY strata) on holdout hit-rate one-sided p-values.

### P6 — Strata (descriptive)

Per session (ASIA/LONDON/NY) and per call-vs-realized contingency. Tables only.

### P7 — PASS/FAIL (the SAME D6 rule as 2deab3c8, fixed before looking)

The reconstructed bias is **USABLE** iff on HOLD OUT: (a) called-rows hit-rate
Wilson **lower bound > 0.50**, AND (b) net-of-friction mean session-window
return **t > 2**. Otherwise **NOT USABLE at this n**.

### P8 — The four-trade check (descriptive)

Quote what the reconstruction stamps for the four session-plans behind today's
losers (ASIA 09-02, LONDON 09-02, NY 09-02) and whether it matches the plan
`Bias.Direction` the store records.

**Nothing below this line was known when the criteria above were written.**
