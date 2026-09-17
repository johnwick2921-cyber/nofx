# Round 25 — SEAT DISPLACEMENT (DS-R001)

Dispatch: CTO 2026-09-17 16:30Z. Traceability per the R24 audit lesson.

## Reconstruction method (Q1)

Per read: `kernel.AssembleResearchLevels("hoang", bars1m, DefaultSessionRegistry,
"MNQ", 12, nil, 1.2, readTime, 1.0, "")` — the exported production path
(kernel/levels_assemble.go:217), called with the LIVE config read from the DB copy:
`SELECT json_extract(config,'$.day_plan.max_levels'), ... FROM strategies ORDER BY
updated_at DESC` → max_levels=12, proximity_filter_atr=1.0; htf_seats unset
(seatHTF default), htf_score_multiplier unset (1.2), min_grade unset.

Limits (stated): (1) config-as-of is NOT reconstructed — the live config applies
to all 3,093 reads; (2) no LevelStateProvider installed → all-fresh (the S4-gate
convention, byte-identical to pre-W11b goldens); (3) bars = the same per-contract
lastClosed(…, 2000, readTime) ring the planner reads; same store copy.

Proximity band (quoted): kernel/levels_score.go:465 `band := proximityK * dATR`;
proximityK = day_plan.proximity_filter_atr (1.0) with the ≤0 fallback
ActivationWindowK=1.5 (levels_score.go:460-463); dATR = kernel.DailyRangeProxy.

Sides (fixed convention, no trend labels): above = level.Price > read price;
below = level.Price < read price. Farthest = max |price − level.Price| among the
pre-cap scored pool rows with HTF=true and TF ∈ {1d, 1w, 4h} (observed set stated
in the report).

Big-day flag (independent of Chief's Q4): window complete AND session range
(max High − min Low over [winStart, flat) 1m bars) ≥ 250 pt.

## Artifacts

- harness/ (r25.go + the round-23 harness base, builds against origin/dev)
- `go build -o /tmp/r25pass harness/ && /tmp/r25pass -db /home/hoang/nofx-r101/data/db.copy.db -out out-r25 -r25`
  → out-r25/r25-reads.jsonl (gitignored)
- cells.py: Q1–Q5 (`python3 cells.py out-r25 <db-copy>`), reads plans (scenario
  usage join, ≤1 tick) and bars (reach) from the copy read-only.
- Manifest in the report: inputs with sha256, commands, outputs.
