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

### P0 · FOUNDATIONS ✅ COMPLETE (pushed)
- ✅ **P0.1** `d1851dac` day_plan config JSON — both codecs, ROOT placement, round-trip golden FIRST + MergeStrategyConfig survival (`+config-truth`)
- ✅ **P0.2** `0a974d31` plans/plan_overlays append-only + decision FK + WAL + single-writer goroutine (25-way concurrent proof, -race clean)
- ✅ **P0.3** `6a0d233b` CT-anchored session registry — ASIA/LONDON/NY, NY-only enabled, killzones, half-day hook (dormant); NOT yet wired to live gate (P2)
- ✅ **P0.4** `b51ab5c2` scenario-fact evaluator (keystone) — distance/closes_beyond/sweep/acceptance/reclaim/reject/still_valid, 11 fixture tests
- ✅ **P0.5** `041e4450` level-state table — identity-keyed (times_tested/consumed/freshness A→B→C), cross-session persist; snapshot writer = P1 (RECON #2)
- ✅ **P0 EXIT BAR** — build/vet/`test ./...`/-race(store,kernel) green · goldens untouched · config-truth locked · zero FE touched (tsc/npm N/A)

**P0 zone flag for owner:** NY flat encoded 14:45 CT (= 15:45 ET, pre-close). Admin-editable, not yet live. Confirm at ★ OWNER RESTART 1.

### P1 · THE MAP — 🔧 IN PROGRESS (4/8 items done, pushed) · ★ OWNER RESTART 1 after P2
- ✅ **P1.1** `14adc47e` session-tagged bars + multi-day extractor (PDH/PDL/PDC · RTH/AS/LDN/ON · PW/PM); pure, warms forward
- ✅ **P1.2** `e1cd5993` round numbers (100/50/25 within 1.5×dATR) + gap tracker (fill-state) + OR/IB (+1.5×/2× ext)
- ✅ **P1.5** `3058e131` confluence scorer → graded TOP-8 (type×fresh×confluence×HTF, day-trade lock, priority seating) + KEY LEVELS renderer
- ✅ **P1.6** `15e55520` regime block (7 fields, graceful VIX/RV degrade) + one-line render
- ✅ sample `9436ea79` full detector→scorer→render pipeline block (exit-bar sample)
- ⬜ **P1.3** durable session-profile store + 17:00-CT snapshot writer (RECON #2 MANDATORY) — couples to trader loop → NEXT SESSION
- ⬜ **P1.4** EQH/EQL + S/D zones + FVG/OB (C/confluence-only) — NEXT SESSION
- ⬜ **P1.7** KEY LEVELS block into LIVE executor prompt (goldens deliberate) — the visible payoff; scorer+renderer READY → NEXT SESSION
- ⬜ **P1.8** calendar fetcher (ForexFactory weekly JSON + static T1 fallback, per-day slice stored) — NEXT SESSION
### P2 · THE CLOCK — ⬜ (bar-close cadence, skip-while-open gate BUILD, last_entry/eod_flat, MAE/MFE) · ★ OWNER RESTART 1
### P3 · THE PLANNER — ⬜ (per-session read jobs, planner-model binding, spec input pkg, schema-strict fail-closed, prompt reorder, advisory)
### P4 · THE CARD — ⬜ (SessionTabs/timeline/HandoverBanner/PlanCard, chart overlay, alert center, Studio block, /api/plan/*)
### P5 · THE DOOR — ⬜ (overlay API, edit sheet + bulk-add, conflict chip, Ask-Planner, adherence grade, digest, stats gate, blind-mark prep) · ★ OWNER RESTART 2

---

## Checkpoint log
- **2026-08-14 ~18:10 CT** — STEP 0 PASS. Spec + heartbeat committed. P0 recon workflow launched. Starting P0.1 (golden-first).
- **2026-08-14 ~18:40 CT** — **P0 · FOUNDATIONS COMPLETE.** All 5 items committed (d1851dac / 0a974d31 / 6a0d233b / b51ab5c2 / 041e4450) + config-truth test. EXIT BAR green (build/vet/test/-race). All ADDITIVE + dormant until wired (P1-P4) — the running bot (rev 3624a2a4, PID 363618) is untouched; NEW tables + WAL activate only at the next owner-driven rebuild+restart (★ RESTART 1). Report: docs/superpowers/reports/2026-08-14-dayplan-p0-foundations.md. NEXT: P1 · THE MAP (detectors + scorer + regime + key-levels prompt block + calendar).
- **2026-08-14 ~19:15 CT** — **P1 · THE MAP CHECKPOINT (4/8 items).** Built the pure/deterministic backbone: multi-day extractor (P1.1), round/gap/OR-IB detectors (P1.2), confluence scorer + KEY LEVELS renderer (P1.5), regime block (P1.6), + an end-to-end sample-block test. 5 commits (14adc47e→9436ea79), all pushed. Full suite + -race(kernel,store) green; goldens + FE UNTOUCHED (all new files); running bot still untouched. Sample KEY LEVELS block captured in the report. REMAINING (next session, budget-edge checkpoint — never start an item I can't finish): P1.3 durable store+snapshot writer (loop-coupled), P1.4 zones, P1.7 prompt wiring (goldens deliberate; scorer+renderer ready), P1.8 calendar. Report: docs/superpowers/reports/2026-08-14-dayplan-p1-map.md.
