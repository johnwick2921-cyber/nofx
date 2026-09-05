# Per-kind verdict on the COMBINED evidence — 1h replay + live D1′

Owner brief (SECTION 2 addendum, 2026-09-05): merge the 1h-replay per-kind results
(`2026-09-02-level-kind-replay.md` Part 2, n=28–54 for PDH/PDL/PDC/ONH/ONL) with the live
D1′ counts, state the instrument difference, and rule per kind on the combined evidence with
**the 1h caveat printed beside every number**.

## LINE 1 — All five kinds remain TOO FEW. The combined evidence does not merely fail to rank them; it is now ACTIVELY CONTRADICTORY on PDH and PDL, where the two instruments point in opposite directions. Combined Σn is 37–58 against a floor of 200.

## 1. The instrument difference — measured, not asserted

The two bodies of evidence are **not two samples of one quantity**. Measured on the same tape:

| | 1h replay (Part 2) | live D1′ (production) |
|---|---|---|
| bar interval | 1h | 1m |
| Δ = mean\|Δclose\| | **55.459 pt** (n=2,041) | **5.267 pt** (n=16,702); per-read resolved Δ ≈ 4.25–4.35 |
| band at k=3 | **±166.4 pt** | **±12.7–13.0 pt** (recorded per row) |
| horizon H | 6 bars = **6 hours** | 12 bars = **12 minutes** |
| level source | rebuilt from 1h bars (intra-hour extremes hidden) | the constructor's seated levels at the real read |
| exit rule | `exit_on=close` | `exit_on=close` (same) |

The band differs by **~13×** and the horizon by **~30×**. A 1h "hold" means price left a ±166 pt
corridor on its entry side within six hours; a live "hold" means it left a ±13 pt corridor
within twelve minutes. Those are different questions about the same level.

**Therefore Σn is reported but MUST NOT be read as a sample size.** Adding 54 one-hour
episodes to 4 one-minute episodes does not yield 58 observations of anything; the column exists
only to show that even the most generous arithmetic stays far below the n=200 floor. No pooled
p(hold) across instruments is computed here, because it would have no defined estimand.

## 2. The combined table (⚠1h marks every 1h-instrument number, per the brief)

| kind | 1h holdout n | 1h p(hold) ⚠1h | 1h Wilson ⚠1h | 1h null | live 1m n | live p(hold) | live Wilson | Σn | verdict |
|---|---|---|---|---|---|---|---|---|---|
| ONL | 54 | 0.6852 ⚠1h | [0.5526,0.7932] ⚠1h | 0.5540 | 4 | 0.7500 | [0.3006,0.9544] | 58 | **TOO FEW** |
| ONH | 53 | 0.7547 ⚠1h | [0.6243,0.8507] ⚠1h | 0.5449 | 3 | 1.0000 | [0.4385,1.0000] | 56 | **TOO FEW** |
| PDC | 45 | 0.5333 ⚠1h | [0.3908,0.6707] ⚠1h | 0.5515 | 11 | 0.4545 | [0.2127,0.7199] | 56 | **TOO FEW** |
| PDL | 36 | 0.4444 ⚠1h | [0.2954,0.6042] ⚠1h | 0.5258 | 8 | 1.0000 | [0.6756,1.0000] | 44 | **TOO FEW** |
| PDH | 28 | 0.7500 ⚠1h | [0.5664,0.8732] ⚠1h | 0.5319 | 9 | 0.4444 | [0.1888,0.7334] | 37 | **TOO FEW** |

`⚠1h` = coarser instrument: ±166 pt band, 6-hour horizon, levels rebuilt from 1h bars.
`1h null` = that kind's stationary-bootstrap holdout baseline from Part 2 §2.4.

## 3. Ruling per kind, on the combined evidence

- **ONL — TOO FEW.** 1h 0.6852 ⚠1h (n=54) and live 0.7500 (n=4) agree in direction, but the
  live cell is four episodes and its interval [0.3006,0.9544] spans almost the whole unit
  square. Agreement between a coarse estimate and a four-sample one is not corroboration.
- **ONH — TOO FEW.** 1h 0.7547 ⚠1h (n=53) is the highest 1h cell; live is 3/3 hold, p=1.0000
  [0.4385,1.0000]. A three-episode 100% is what a coin flip produces one time in eight.
- **PDC — TOO FEW.** The only kind where both instruments sit near their nulls (1h 0.5333 ⚠1h
  vs null 0.5515; live 0.4545). Consistent, and consistently unremarkable.
- **PDL — TOO FEW, AND THE TWO INSTRUMENTS CONTRADICT.** 1h says 0.4444 ⚠1h (n=36, below its
  0.5258 null); live says **1.0000** (8/8, n=8). One instrument's best guess is "worse than a
  coin"; the other's is "never breaks".
- **PDH — TOO FEW, AND CONTRADICTORY IN THE OPPOSITE DIRECTION.** 1h says 0.7500 ⚠1h (n=28);
  live says 0.4444 (n=9). The kind the 1h replay ranked highest is the kind the live
  instrument ranks below a coin flip.

**PDH and PDL are the finding.** Two calibrated instruments, run on the same market, disagree
in *direction* on both. That is the signature of two tiny samples of noise, and it is stronger
evidence for "no rank" than either instrument's silence alone. H7 (PDL shows no hold edge)
survives in the sense that matters: nothing stable is visible.

## 4. Live-side caveats that bound its half of the evidence

- **677 episodes, but only 3 plan_ids across 37.6 hours and 50 distinct level prices.** These
  are not independent draws: the same levels are re-touched within a session, so effective n is
  materially below nominal n. The per-kind cells above are upper bounds on information.
- **Pooled live D1′ = 0.4981 [0.4555,0.5407], n=526** — indistinguishable from a coin flip and
  from the instrument's own 0.5067 calibration point. The live tape shows no pooled level edge,
  matching Part 1 and Part 2's pooled findings.
- Kinds the 1h replay never covered now have the largest live cells: **VWAP n=132 (0.6212),
  DEMAND n=128 (0.4375), RTH-L n=124 (0.3387 [0.2614,0.4257])**. RTH-L's interval excludes 0.50
  from BELOW — the only live cell that separates from a coin flip — but it is one instrument,
  one day, unpre-registered, and 20 kinds were examined. It is a hypothesis for the next window,
  not a finding, and it is not one of the five kinds this brief rules on.

## 5. What was checked in the source report, and stands

Part 2 §2.4 already handled the one number that could have been over-read: pooled 1h holdout
0.6389 [0.5729,0.7000] does clear the per-kind bootstrap means (0.526–0.554), but sits at the
**65.6th percentile of the Osler random-level null**, whose spread (sd 0.2216) means the same
figure arises from randomly-placed prices about one day in three. That refusal to claim a
verdict was correct and is not disturbed by the live data.

## 6. Verdict

**No kind ranks. The grader ruling stands; every ladder term keeps its [I] label.** Adding the
live D1′ evidence changed the *reason* rather than the answer: before, the cells were merely
too small; now the two instruments actively disagree on the two kinds with the most extreme 1h
readings. The honest lever remains time and retention — 1m at 90 days plus 1h growing — not a
lower floor, not a pooled n across instruments, and not a threshold change.
