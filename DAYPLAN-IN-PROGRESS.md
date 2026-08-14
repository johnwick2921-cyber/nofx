# DAY-PLAN CAMPAIGN — IN PROGRESS

Owner-approved 2026-08-14. Build contract: [docs/VL-DAYPLAN-FULL-SPEC.md](docs/VL-DAYPLAN-FULL-SPEC.md)
(v1 FINAL, code-verified recon @ ca1f38c6). Multi-session, checkpointed (hardening pattern).

**Standing rules:** own commit per item · push-per-part · ADDITIVE · SIM-only · NO agent restarts
(owner restarts at ★ points only) · guardrails master untouched (OFF = owner learning mode).

**STEP 0 (this session, HEAD c051b975):** PASS — HEAD ≥ 3624a2a4 ✓ · tree clean ✓ ·
running rev 3624a2a4 (HEAD ancestor, FE-only since) ✓ · both traders cycling (hoang #440, 15m #84) ✓.

---

## Part ledger

Legend: ⬜ not started · 🔧 in progress · ✅ done+committed · 🚀 pushed

### P0 · FOUNDATIONS
- 🔧 **P0.1** day_plan config JSON — both hand-rolled codecs (store/strategy.go:731-823), ROOT placement, round-trip golden FIRST
- ⬜ **P0.2** plans / plan_overlays tables + decisions FK — append-only, PK(plan_id,version), json_valid, WAL + single-writer goroutine
- ⬜ **P0.3** session registry — CT-anchored ASIA/LONDON/NY (read/flat/killzones/enabled), NY-only; half-day truth hook
- ⬜ **P0.4** scenario-fact evaluator (Go) — distance/closes_beyond/sweep/acceptance/reclaim/reject/level_still_valid, fixtures FIRST
- ⬜ **P0.5** level-state table — identity-keyed (times_tested, consumed, freshness) + durable session-profile store (RECON #2)
- ⬜ **P0 EXIT BAR** — suite green (build/vet/test/-race/tsc/npm) · goldens deliberate-only · config-truth 4-step · push · checkpoint report

### P1 · THE MAP — ⬜ (detectors, scorer, regime, key-levels into executor prompt, calendar) · ★ OWNER RESTART 1 after clock
### P2 · THE CLOCK — ⬜ (bar-close cadence, skip-while-open gate BUILD, last_entry/eod_flat, MAE/MFE) · ★ OWNER RESTART 1
### P3 · THE PLANNER — ⬜ (per-session read jobs, planner-model binding, spec input pkg, schema-strict fail-closed, prompt reorder, advisory)
### P4 · THE CARD — ⬜ (SessionTabs/timeline/HandoverBanner/PlanCard, chart overlay, alert center, Studio block, /api/plan/*)
### P5 · THE DOOR — ⬜ (overlay API, edit sheet + bulk-add, conflict chip, Ask-Planner, adherence grade, digest, stats gate, blind-mark prep) · ★ OWNER RESTART 2

---

## Checkpoint log
- **2026-08-14 ~18:10 CT** — STEP 0 PASS. Spec + heartbeat committed. P0 recon workflow launched. Starting P0.1 (golden-first).
