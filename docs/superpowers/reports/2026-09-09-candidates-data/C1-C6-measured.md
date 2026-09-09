# C1–C6 measured at the running rev — 2026-09-09, session nofx-66

Running rev verified three ways: `/api/health` → `954f11b15f2e` · `/proc/438/exe` →
`/home/hoang/nofx/nofx-bin`, `vcs.revision=954f11b15f2e7615678f7d2b708c47895faebf1e`,
`vcs.modified=false` · `deploy/RELEASE` = `954f11b1`. All agree.

SPEC-FRESHNESS: the dispatch's pins `f5927cdc` (range-fade) and `982091d4` (trading-policy)
are the MERGE commits; `git log -1 -- <file>` shows the authoring commits `b2ba8c21` / `0ec5bd2c`.
`git diff <pin> HEAD -- <file>` is EMPTY for both — no drift, specs read as current.

| # | premise | verdict | measured |
|---|---|---|---|
| C1 | max_levels=12; 12/12 in 14 of 15 reads; 181 cut rows all score 0 | **PARTLY** | resolved `max_levels`=**12** ✓ (file const `DefaultMaxLevels=8`, `levels_score.go:54` — the boot line prints the const, not the resolver). Live reads show **both** `seated 12/…` (111×) **and** `seated 24/…` (34+32+16×). Cut reasons now: **581** `max_levels: 12 seated, cap 12` + **8** min-grade. The "181 all score 0" was the **09-04 sample**; per-day zeros: 09-04 181, 09-06 36, 09-07 48, 09-08 240, **09-09 84 rows, ALL real scores, zero zeros** — Stage A works, no A24 regression. Today's cap-cut rows include **grade A at score 1.344–1.36**. |
| C2 | 22% of seats >100 pts; touched 31% vs 79% | **NOT REPRODUCED — worse** | today n=84 seats (ids 1009–1176): ≤25pt **14** · 25–50 **20** · 50–100 **25** · **>100pt 25 = 29.8%** (avg 150.1 pt). `touch_outcomes` n=3931 (ids 1–3931): hold 1627 (41.4%) · break 1477 (37.6%) · ambiguous 827 (21.0%). |
| C3 | confluence counts names, not independent evidence | **CONFIRMED + SHARPENED** | `conf` counts **distinct families** within `confBand` (`levels_score.go:462-476`); same-family already excluded (`seenFamilies`). Credit `(1 + 0.20*effConf)`, `ConfluenceCap()`=3 (`:194-200`), applied at `:502`/`:508`. **`confBand := 0.10 * dATR` (`:427`) ≈ ±20.3 pt today, while the cluster COLLAPSE width is fixed 3.00 pt (`LevelClusterTicks=12`, `:717`, used `:570`) — a ~6.8× mismatch, and BOTH are commented "cluster tolerance".** The collapse (`:751-779`) runs **after** scoring and its own override note says `"cluster confluence display %d -> %d; score unchanged"`. |
| C4 | OB 8/117 (6.8%), SUPPLY 5/23, DEMAND 22/59, EQH 0/4 | **NOT REPRODUCED** | **OB 34/187 = 18.2%** · SUPPLY 12/36 = 33.3% · DEMAND 30/75 = 40.0% · **EQH 17/146 = 11.6%** (not 0/4). Tier-1 anchors seat 100%: ONH 34/34, ONL 40/40, PDL 35/35, PDC 33/33; VWAP 64/68 = 94.1%. |
| C5 | nothing exists beyond the map | **CONFIRMED** | grep counts (kernel/+trader/, excl. tests): measured-move **0** · ATR-projected extreme **0** · prior-week 8 · round-number 15 · "projection" 2. No projector exists. |
| C6 | PWH/PWL can never seat (4,320-bar guard on a 33h ring) | **CONFIRMED + SHARPENED** | `priorWeekMinBars = 4320` (`levels_multiday.go:224`), gate at `:198` via `pwCovered`. **The const block's own comment (`:219-221`) says: "with the ~33h 1m ring these can never pass, which is CORRECT: those anchors must come from a multi-day source, not the ring."** `candidate_pool` rows for PWH/PWL/PMH/PML: **0, ever**. A daily source EXISTS: `bars` tf=`1d` n=1901 through 2026-09-08. |
