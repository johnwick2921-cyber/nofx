# Round 25 — SEAT DISPLACEMENT (DS-R001)

Branch `docs/round-25-seat-displacement` (claim `b609a346`, work `f4782941`+).
Dispatch: CTO 2026-09-17 16:30Z, GO 17:58Z/21:00Z. Generated tables are the
record; narrative is commentary. No bot code.

## MANIFEST (inputs → scripts → outputs; anything not here is NOT MEASURED)

| item | value |
|---|---|
| harness | `research/2026-09-17-round-25-seat-displacement/harness/` (r25.go + round-23 base), commit `f4782941` |
| pass command | `/tmp/r25pass -db /home/hoang/nofx-r101/data/db.copy.db -out out-r25 -r25` (nice 19, ionice 3) |
| input db.copy.db sha256[:16] | `998a15ee230f321d` |
| output | `out-r25/r25-reads.jsonl` (3,093 reads; gitignored, on-disk) |
| cells | `cells.py` (this commit), `python3 cells.py out-r25 <db-copy>` |
| lookahead | bars are the lastClosed(…, 2000, readTime) ring — closed strictly before the read; reads precede windows by 30 min (S4-gate proof, 0/4,471,482 violations) |

## Method (Q1) — the 12-seat reconstruction

Per read: the exported production call
`kernel.AssembleResearchLevels("hoang", bars1m, DefaultSessionRegistry, "MNQ", 12, nil, 1.2,
readTime, 1.0, "", extra...)` with the SAME extras the live planner feeds
(`DetectHTFLevels` over [D,4h,1h,15m] via the per-TF closed-bar fetch +
nPOC from session_profiles — auto_trader_planner.go:2502-2518). Live config read
from the DB copy (`max_levels=12`, `proximity_filter_atr=1.0`; htf_seats/mult/min_grade
unset). Band quoted: kernel/levels_score.go:465 `band := proximityK * dATR`.
Limits stated: config-as-of NOT reconstructed; no LevelStateProvider → all-fresh
(the S4-gate convention); lane 93's R24 branch carries no seats reconstruction
(verified) — this is the only one.

Convention (fixed, Q5 lesson): sides are PRICE sides — above = level.Price >
read price, below = < read price. Farthest = max |level − price| among pre-cap
pool rows with HTF=true and TF ∈ {1d, 1w, 4h} (observed set: 1d/4h only — no 1w
rows exist in the pool). Big-day flag recomputed independently of Chief's Q4:
window complete AND session range ≥ 250 pt.

## Q1 — reconstruction + farthest in-band HTF per side

| fact | value |
|---|---|
| reads | 3,093 |
| reads with a farthest in-band HTF level (1d/4h) above / below | 157 / 122 — **all 2026** (the S4c coverage shape: 4h/1d series exist only for 2026 reads) |
| median distance, above / below | 273 pt / 148 pt |
| big days (complete, range ≥ 250) | 612 |

Sample ids: 2022-04-10 LONDON, 2022-04-10 NY, 2022-04-10 ASIA (first three reads;
rows in `out-r25/r25-reads.jsonl`).

## Q2 — is the farthest in-band HTF level already seated?

| side | seated | n | year |
|---|---|---|---|
| above | 29 (18.5%) | 157 | 2026 only |
| below | 50 (41.0%) | 122 | 2026 only |

So the proposed TARGET-ONLY seat would ADD a level in 81.5% (above) / 59.0%
(below) of the measurable reads — i.e., displace something.

## Q3 — the displaced seat and its usage cost

200 displacements (all 2026; a displacement = farthest HTF not seated → the last
seat under the priority sort is evicted). Displaced seat kinds: **OB 1h (41),
OB 4h (25), VWAP (23), DEMAND 1h (14), SWG-L 5m (9), IFVG 1h (8), PDH (7)** —
the rule mostly displaces OTHER already-seated HTF zones (OB 1h/4h = 66 of 200),
not today-references.

Usage (a scenario authored on the displaced level in that session-day's plan,
≤1 tick price match): the plans table covers only recent September session-days,
so usage is measurable on **40 of 200** displacements. **Used: 0 of 40 (0.0%).**
The last seat is never referenced by a scenario in the measurable set. For the
160 June–August displacements: NOT MEASURED (no plan rows exist in the copy).

## Q4 — on the overshot big days, was the farthest in-band HTF level the reached target?

Overshot = a big-day read whose session excursion crossed the farthest SEATED
level on that side (861 overshot sides across the 612 big days). Of those, **65
sides** carry a farthest in-band HTF (1d/4h) level on that side (the rest NOT
MEASURED — 2022–2025 reads have no 4h/1d series, the S4c coverage shape).

| metric | value |
|---|---|
| overshot big-day sides | 861 |
| with a farthest in-band HTF on that side | 65 (all 2026) |
| **reached** | **39 / 65 = 60.0%** |

Reached ids (first 5): 2026-06-03 LONDON below@30689.25 · 2026-06-03 NY
above@30689.25 · 2026-06-08 NY below@29910.75 · 2026-06-10 LONDON below@29005.5 ·
2026-06-11 LONDON below@29694.5. Missed (first 5): 2026-06-07 LONDON
above@29910.75 · 2026-06-07 NY above@30775.875 · 2026-06-08 NY above@30455.5 ·
2026-06-10 NY above@29922.375 · 2026-06-11 LONDON above@30455.5.

## Headline

Where the rule can even apply (2026 reads with a 1d/4h level in band): it adds a
level 81.5%/59.0% of the time, the displaced seat is usually another already-seated
HTF zone (OB 1h/4h) and is **never used** in the measurable subset (0/40), and on
overshot big-day sides the proposed target level WAS reached 60.0% of the time
(39/65). Cells only; no recommendation beyond them.
