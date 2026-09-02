# LEVEL-KIND REPLAY — D1′ over the persisted MNQ tape, mechanical level kinds only

**READ-ONLY · no lock · no engine code · scripts in `~/nofx-analysis/level-replay/`**
Owner: hoang · Live rev `e1b1176844aa` · Report opened **2026-09-02 ~18:30 CT**
Instrument source: `docs/superpowers/reports/2026-09-02-detector-redesign.md` (D1′ amendment)
Base scripts (imported, not copied): `~/nofx-analysis/detector-redesign/detectors.py`

---

## 0. PRE-REGISTRATION — written 2026-09-02 ~18:30 CT, BEFORE any real-tape replay ran

Fixed in advance; nothing below this line was observed when this header was written.

### Period (R1)

Query (read-only): `SELECT tf, COUNT(*), MIN(open_time_ms), MAX(open_time_ms) FROM bars
WHERE symbol='MNQ' AND tf='1m'` on `data/data.db?mode=ro` → **1m = 24,103 rows,
2026-08-19 17:00 CT → 2026-09-02 08:28 CT**.
Replay period = every COMPLETED CME session day (17:00 CT roll, `detectors.session_day`
quoted from `detector-redesign/detectors.py`) inside that range:
**2026-08-20 … 2026-09-01 = 13 session days.** 2026-09-02 excluded (incomplete day).
**SURPRISE (stated up front):** the persisted intrabar history is ~14 days, NOT years.
The deep TFs exist (1d back to 2018-12, 1w to 2011-12) but carry no intrabar
granularity, and 5m starts 2026-08-24 (worse than 1m). Per R1 I run 1m over the 13
completed session days and state the n-limits explicitly.

### Instrument (D1′)

`detectors.detect_symmetric_v2(bars, level, k, Δ, H=12, exit_on='close')`, imported
from `~/nofx-analysis/detector-redesign/detectors.py` (path quoted). Barriers anchored
on the level (`L ± k·Δ`); episode opens on a bar that straddles `L` while the previous
bar does not; `entry_side` from the previous bar's close; exit when a CLOSE crosses a
barrier (next bar onward); both in one bar → `ambiguous_span`; horizon `H=12` →
`ambiguous_horizon`. p(hold) = hold/(hold+break); ambiguous recorded, counted,
EXCLUDED from the denominator.

**k and Δ (re-derived from the replay period, per dispatch):** Δ = mean absolute 1m
close increment over the replay period's own bars (the detector-redesign tape_stats
definition, quoted). **k = 3** → band = ±3Δ. If the replay Δ departs >20% from the
calibration Δ (5.3771 pt), the band in points is re-reported and the k=3 band on the
replay Δ is the ONLY band used for verdicts.

### Levels — mechanical kinds, reconstructable per session day (definitions mirrored)

| kind | per session day d | mirrored definition |
|---|---|---|
| PDH / PDL | 1 each | prior completed session day's 1m high / low (session-day window 17:00 CT → 17:00 CT, `detectors.session_day`) |
| PDC | 1 | last 1m close of the prior session day |
| ONH / ONL | 1 each | high/low of bars from session start (prev 17:00 CT) up to <08:30 CT |
| RTH-H / RTH-L | 1 each | high/low of exchange RTH 08:30–15:00 CT bars (engine NY flat is 14:45 CT by owner contract, `kernel/session_registry.go:88-113` quoted — replay uses exchange RTH for the traded range) |
| OR5-H/L · OR15-H/L · OR30-H/L | 3 pairs | high/low of the first 5/15/30 minutes from 08:30 CT |
| VWAP · VWAP±1σ | 3 | mirror `kernel/levels_volume.go:76-92 vwapAndStdev` (typical price (H+L+C)/3, volume-weighted; σ = volume-weighted sd of typical prices), computed on session bars (17:00 CT roll) **through 08:29 CT — the RTH-open snapshot, frozen for the day** (deviation from the live re-emitting line: stated, conservative, lookahead-free) |
| Round x000 / x500 / x250 | every 250-multiple | multiples of 250 within [session-day min − 250, session-day max + 250] |
| Settlement | 1 | mirror `vwapAndStdev` over 14:59:30–15:00 CT bars; if none, last RTH 1m close |

Touch ordinal per level per day: episode counter per (level, session day) from the tape.

### OOS split (R2, fixed before looking)

Days sorted: **exploration = first 8 session days (08-20 … 08-27); holdout = last 5
(08-28 … 09-01).** Both are reported; VERDICTS COME FROM HOLDOUT ONLY.

### Verdict rule (pass/fail, pre-registered)

Per kind, on HOLDOUT: p̂ = p(hold), Wilson 95%. Baselines per kind on holdout:
(a) **stationary bootstrap** — geometric blocks, mean length 10, the SAME per-day level
lines re-run on 40 shuffled paths per holdout day (same method as E2, seed fixed);
(b) **Osler** — B=1000 random level sets per holdout day with the same per-kind counts,
prices uniform in the day's bar range → the real p̂'s percentile.
- **BEATS BASELINE ON HOLDOUT** iff holdout n ≥ 200 AND Wilson lower bound > baseline p̂
  AND the one-sided test (normal approx on p̂ − p̂0) survives Benjamini-Hochberg q=0.05
  across the tested kinds.
- **DOES NOT** iff n ≥ 200 but the interval overlaps the baseline (or BH kills it).
- **TOO FEW** iff holdout n < 200 — descriptive only; no ranking below n=200 is a verdict.
Cohen's h(0.60 vs 0.50) = 0.2014 quoted as the yardstick (n≈193/group for 80% power).

### Strata (pre-registered)

R5 ordinal: 1st / 2nd / 3rd+ touch per level per day, pooled per period; holdout used
where n ≥ 30, else descriptive. R6 session: ASIA 17:00–02:00 / LONDON 02:00–08:30 /
NY 08:30–14:45 CT (engine registry windows, quoted) assigned by the touching bar's time.

### Bias side table (R7, same tape, exploration + holdout)

1. **TSMOM 1/3/12-month** state = sign of cumulative return over the prior 21/63/252
   session-day daily closes (1d table, quoted) vs sign of the next session day's RTH
   return (15:00 close − 08:30 open). 12m state uses the deep daily history; the
   prediction set is the replay period (n=13).
2. **Overnight → RTH**: sign(08:30 open − prev 17:00 open) vs sign(15:00 close − 08:30
   open) same day.
3. **First-30 → last-hour**: sign(09:00 close − 08:30 open) vs sign(15:00 close −
   14:00 open) same day.
All reported with n, Wilson on sign agreement, and the holdout subset (n=5, descriptive).

### Hypotheses (pre-registered, directional where stated)

- H1 (pooled): holdout pooled p(hold) ≈ stationary-bootstrap baseline (no level edge) —
  carried from the detector redesign verdict.
- H2 (PDL lead): the E3 PDL cell (0.6970 [0.5266, 0.8262], n=33) will NOT replicate on
  holdout.
- H3 (consumed touch): if the belief is real, p(hold) DECLINES with ordinal
  (1st > 2nd > 3rd+).
- H4 (rounds): x000/x500/x250 p(hold) ≈ Osler random-level null.
- H5 (VWAP): VWAP / ±1σ indistinguishable from baseline on holdout.

### Grader recommendation rule (pre-registered)

BEATS → keep (graded normally); DOES NOT → neutralize (grade ×1.0, no bonus); TOO FEW →
keep with the current [I] label intact (no change justified). Applied per kind to
`kernel/levels_score.go` ladder terms.

**Nothing below this line was known when the criteria above were written.**

### AMENDMENT (post-hoc, written 2026-09-02 ~18:45 CT before the replay ran — declared, not hidden)

The tape exported from the store at run time differs from the range quoted in the
header: the live bot's bar persistence has been appending since the 18:05 CT class-48
boot. Actual export: **14,258 MNQ 1m bars, 2026-08-19 10:00 CT → 2026-09-02 18:34 CT**
(same read-only query, `symbol='MNQ' AND tf='1m'`). Completed session days (17:00 CT
first bar, ≥15:00 last bar, ≥1000 bars): **2026-08-20 … 2026-09-02 = 10 days**.
The pre-registered 60/40 rule is applied mechanically to this actual list:
**exploration = first 6 (08-20 … 08-25), holdout = last 4 (08-26 … 09-02)**.
The header's illustrative day labels (08-20..08-27 / 08-28..09-01) were computed from a
stale row count; the SPLIT RULE was not changed, its inputs were.
Replay Δ = **5.3707 pt** (vs calibration 5.3771, −0.1% ⇒ band ±16.11 pt at k=3,
no re-band needed). lag1 ρ = −0.0403. Nothing below this line was known when the
criteria above were written.
