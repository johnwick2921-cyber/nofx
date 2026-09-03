# 2026-09-02 — Bias calibration: three evidence-backed signals vs the shipped weekly-structure bias, pre-registered, out-of-sample

**Dispatch:** BIAS CALIBRATION — round-9 signals on 8 years of NT8 daily bars. READ-ONLY on the engine · no lock · no live calls.
**Tree:** worktree `~/nofx-biascal`, branch `docs/bias-calibration-0902`, base `61b11d5f` (dev HEAD at start). Scripts: `~/nofx-analysis/bias-calibration/`. Bars source: the `bars` table (SQLite, read-only), NT8-native rows persisted by the bar-source wave (completed bars only — the persister writes `ClosedBarsOnly`). CSVs: `docs/superpowers/reports/2026-09-02-bias-calibration-csvs/`.
**Evidence tiers:** [A] directly verified · [B] inferred · [C] speculation.

---

## PRE-REGISTRATION HEADER (fixed 2026-09-02 19:20 CT, BEFORE any signal computation)

### P1 — Data inventory (measured, not looked-at-for-results)
Query: `SELECT tf, symbol, COUNT(*), MIN(open_time_ms), MAX(open_time_ms) FROM bars WHERE tf IN (...) GROUP BY tf, symbol` against `~/nofx/data/data.db` (read-only), 2026-09-02 19:18 CT.

| TF | MNQ range (UTC) | n MNQ | ES range (UTC) | n ES |
|---|---|---|---|---|
| 1d | 2019-05-02 → 2026-09-01 | 1,896 | 2018-12-05 → 2026-09-01 | 1,999 |
| 1h | 2026-05-04 → 2026-09-03 | 2,002 | same | 2,002 |
| 30m | 2026-07-03 → 2026-09-03 | 2,005 | same | 2,005 |
| 15m | 2026-08-04 → 2026-09-03 | 2,010 | same | 2,010 |
| 1w | 2019-04-26 → 2026-08-21 | 383 | 2004-12-03 → 2026-08-21 | 948 |

Stamp convention: bar `open_time_ms` is the bar's OPEN time, epoch-ms UTC; 1d/1h/30m/15m stamps sit on CT clock grids (verified: first 1d stamp 2018-12-05 06:00 UTC = 00:00 CT December). 1d bars are calendar-day CT (open 00:00 CT, close ≈ last pre-16:00 CT RTH print). Native `1w` is NOT used — NT8 stamps it Fri→Thu (class 7; the resolver's `ExcludedNative("1w")`), and our weeks are Monday-governed. **Premise nit:** daily bars reach 2018-12 for ES and 2019-05 for MNQ (the dispatch's parenthetical had it backwards) — ranges used accordingly.

### P2 — Signals (definitions fixed, no peeking after this header)

**S1 TSMOM regime** (freq: session). Lookback L ∈ {21, 63, 252} sessions. At each session close t:
`mom(t) = sign(Σ_{i=0..L−1} r_{t−i})` where `r = close-to-close log return`, vol-scaled copy computed as `mom(t) / (σ63(t) · √L)` for the position-size/return-normalization variant (sign identical by construction). Target: `sign(r_{t+1})`. Position for PnL: `+1 / −1` per mom sign (never flat).

**S2 Overnight → RTH reversal** (freq: session-day). `overnight(t) = RTHopen(t) / RTHclose(t−1) − 1`; `rth(t) = RTHclose(t) / RTHopen(t) − 1`. RTH open = open of the bar stamped 08:30 CT; RTH close = close of the bar stamped 15:45 CT (15m grid) / 15:30 CT (30m grid). Signal: `pred(t) = −sign(overnight(t))` (reversal hypothesis). Target: `sign(rth(t))`. Runs: (a) 15m MNQ 2026-08-04→ (exact RTH stamps), (b) 30m MNQ 2026-07-03→ (same exact stamps, longer range) — both stated, both holdout-only. Reversal rate: fraction of days with `sign(overnight) ≠ sign(rth)` tested against 0.5 (Wilson). Interaction: strata by realized-vol tercile (P6).

**S3 Intraday momentum** (freq: session-day). `first30(t) = open(09:00 CT bar) / open(08:30 CT bar) − 1`; `lastHr(t) = open(14:45 CT bar) / open(13:45 CT bar) − 1` (open-to-open of the boundary bars). Signal: `pred(t) = sign(first30(t))`. Target: `sign(lastHr(t))`. Run: 15m MNQ 2026-08-04→ only (n stated plainly). 5m exists but adds nothing to alignment; not run.

**CONTROL — shipped weekly-structure bias, reconstructed** (freq: week). Weekly candles from 1d by `weekStartMonday` (shipped `kernel/weekly_bias.go`): open = first daily-bar open of the week (Sunday 00:00 CT print — proxy for the 17:00 CT first print, caveat stated), close = last bar's close (Friday ≈16:00 CT), H/L. At each completed week w: `weekly_open(w)` = w's first daily open; `PWH/PWL` = prior completed week's H/L. Rule (shipped Tier-A, `kernel/weekly_prompt.go:285` "PWH/PWL break-AND-HOLD vs sweep-and-reject", "price vs weekly_open"):
- **bull** iff `close_w > weekly_open_w` AND `H_w > PWH_w` AND `close_w > PWH_w`
- **bear** iff `close_w < weekly_open_w` AND `L_w < PWL_w` AND `close_w < PWL_w`
- else **neutral** (flat).
Position for week w+1 = +1/−1/0. Target: `sign(close_{w+1} − close_w)`. This is the null the three signals are measured against.

### P3 — Split
Exploration 2018-12/2019-05 → 2023-12-31 (ES/MNQ respectively); **holdout 2024-01-01 → last completed bar**. Verdicts from holdout only. S2/S3 ranges fall entirely in holdout (state n).

### P4 — Statistics per signal (D2)
1. Hit rate `ĥ = #(pred sign == actual sign) / n`, 95% **Wilson** interval.
2. Mean next-period return conditional on signal sign (+1/−1 buckets), with Student **t-stat** (`t = mean / (std/√n)`, Welch against zero).
3. Long/flat/short PnL-in-points series (position × next-period close delta), no sizing; max drawdown (peak-to-trough in points).
4. Same series **with friction: 2-pt round trip = 1 pt per side, charged as `1pt × |Δpos|`** at each re-position.

### P5 — Multiplicity
Benjamini–Hochberg across the 5 tests (S1×3 lookbacks, S2, S3, CONTROL) applied to the holdout hit-rate p-values (one-sided vs 0.5); raw and adjusted reported.

### P6 — Strata (descriptive, flagged as such)
Realized-vol tercile (trailing 20-session std of daily close-to-close returns, computed over the signal's own sample) and year. Reported as tables only — no verdicts drawn from strata.

### P7 — PASS/FAIL (fixed before looking)
A signal is **USABLE** iff on HOLD OUT: (a) hit-rate Wilson **lower bound > 0.50**, AND (b) net-of-friction mean next-period return **t > 2**. Otherwise **NOT USABLE at this n** — stated per signal.

### P8 — Method honesty
All bars from the `bars` table only (no resolver call, no live system, no engine imports). Closed bars only by construction (persister writes closed bars; last 1d bar is the prior session). Scripts are stdlib-Python (sqlite3 + math), no seeds involved (no randomness anywhere). CSVs carry every observation.

---

## RESULTS (filled after P1–P8 were frozen)
