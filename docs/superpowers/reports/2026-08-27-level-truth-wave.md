# LEVEL-TRUTH WAVE (T1–T9) — 2026-08-27

**Branch:** `fix/level-truth` → merged to dev at `41ce0af8` (base: merged dev incl. the armed-orders E2 wave `56c4f692`). One commit per T. Canon unchanged (SIM-only, additive, zero literals, goldens deliberate).

## Commits

| T | Commit | What |
|---|---|---|
| T1 | `86ac1975` | level_stats writer retry/backoff (4×15s) + skip-reason logs + `BackfillLevelStats` cmd + `runLevelStatsDayOnce` split. Root cause proven: DB-lock swallow by a bare `continue`. |
| T3 | `9b6ac4bd` | SWG-H/SWG-L swing detector (`kernel/levels_swing.go`): 5m+15m structure fractals (k=structureSwingK, min-move ATR), recent lookback 144/96 bars, ≤3/side/TF, `react_zone`, evidence 0.85, anchor ladder. |
| T2 | `0ed9c73c` | Stamp-gap fix, both populations: `ScoreLevelsMinGradeFull` returns the graded PRE-SEAT pool; planner records pool grades + `CarryMachineGrades`; no-trade docs stamp every level. |
| T4 | `c0707881` | `MarkConsumed(key, nowMs)` records the consuming touch (times_tested=1 + last_play_ms when born already-accepted) — consumed rows always carry ≥1 touch. |
| T5 | `deabaaa4` | VWAP±2σ emission (`KindVWAP2S`, evidence 0.85, VWAP family) + golden 3→5. |
| T2/T3 goldens | `ac40cafd` | EQL·15m(HTF) grade-A **reproduced** (confluence ×1.2: 0.70×1.2×1.2=1.008) + no-trade/pool stamp goldens + swing detector test + missed-turns E-proof script. |
| T7 | `eb620cc3` | pnl_corrected root cause: the P0 writer was a ONE-TIME flag-gated pass (only disagreements) with NO close-path writer → NULL forever on agreeing rows. `BackfillPnlCorrectedAll` + `StampPnlCorrectedOnClose` + Δ≥$0.50 class-killer WARN re-armed. |
| T8 | `c9f73fe1` | bar_update frame log → DEBUG, 1-in-N (`BAR_UPDATE_LOG_SAMPLE`=500). 7.5M lines/day was eating the 2G journald cap in <24h. |
| T1/T5/T9 | `dbee0496` | nightly-writer integration proof · spec-doc reconciliation (`docs/README-VL-SYSTEM.md` — deployed values = documented truth; iFVG stays; anchors keep decay) · σ hand-calc pin. |

## Key resolutions

- **EQL·15m (HTF) grade-A mystery — SOLVED, not a bug.** Line-level score = `typeEvidence × freshMult × (1+0.2·conf) × htf` = 0.70 × 1.0 × 1.2 (one distinct confluent family) × 1.2 (HTF) = **1.008 → A**. Without confluence: 0.84 → B. Pinned by `TestT2EQL15mHTFGradeAReproducible`.
- **T9 σ — math CORRECT, no accumulation bug.** `vwapAndStdev` is the exact volume-weighted stdev of typical prices (`sd = √(Σv·(tp−vwap)²/Σv)`). The master-recheck's "87pt implied σ" was a WRONG-WINDOW artifact: session VWAP spans the CME session day (17:00 → read), which legitimately carries ~61pt σ on the 08-26 range day. Hand-calc reproduced the plan's machine VWAP/−1σ to 2dp (29536.20/29475.11 vs 29536.17/29475.09); the plan's +1σ row was model-adjusted (29600.96 vs machine 29597.28) — plan-level edit, admissible.
- **T7 pnl_corrected — root-caused with archaeology.** `store/pnl_correction.go` (P0, ca6e990f) is a one-time flag-guarded pass that writes ONLY |Δ|≥$0.50 rows; nothing ever stamped agreeing rows or new closes → NULL everywhere. E-proof on the last 7 days: **zero Δ≥$0.50 rows** (close-path attribution fixed the class) and corrected == raw on every row — last week's expectancy rulings stand unchanged.
- **T3 E-proof (scripted, DB-driven):** missed-turns baseline 75.0/80.0/82.6% → **60.0/65.0/65.2%** with swing seats (8-seat cap). 50 → 40 missed turns across the last 3 sessions.

## Cutover & E-proofs

- **FLAT-GATE:** zero positions of ANY origin from NT8 truth (TCP positions snapshots `count=0` + `/api/positions` `[]` + DB OPEN=0) before the swap.
- **Deploy:** build at dev HEAD → `deploy/RELEASE` = build sha → mv old binary → `kill -9` (systemd relaunch).
- **Boot line to quote:** `🔐 BOOT INTEGRITY OK — rev <sha> · goldens PASS`.
- **E-proofs:**
  - [x] missed-turns delta (75–83% → 60–65%, scripted)
  - [x] σ hand-calc match (2dp on VWAP/−1σ)
  - [x] T7 Δ-rows = 0 on last-7d closes; corrected == raw
  - [ ] `level_stats` rows growing nightly (post-deploy, next session roll)
  - [ ] unstamped=0 on the first plan written by the new binary
  - [ ] pnl_corrected non-NULL on the newest closes (boot backfill + close-path)
  - [ ] journald volume delta (post-deploy: expect ~20–30k lines/day vs 7.5M)
- **Report:** this file, pinned by commit-ref URL.
