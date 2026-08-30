# E12 — WEEKLY READ REVIEW (W-1/W-2 dress rehearsal, 2026-08-30)

Evidence tiers: [A] = directly verified (ran it / read the line / recomputed) · [B] = strong inference · [C] = speculation.
Quotes from `e11_weekly_validator.txt` (harness-generated, verbatim).

## Wiring (how the harness drove the SAME builder the scheduler uses)

- The scheduler path (`trader/auto_trader_weekly.go`): `runWeeklyRead` loads stored 1m bars (`:71` weeklyBars1m) → `facts := kernel.ComputeWeeklyFacts(bars, now, price)` (`:163`) → `prompt := kernel.BuildWeeklyPrompt(facts)` (`:164`) → call → `kernel.ParseWeeklyDoc` + `kernel.ValidateWeeklyDoc(doc, kernel.WeeklyRefSet(facts), facts.ThinHistory)` with retry-once (`:178-196`) → stamp `doc.FactsHash`/`doc.ThinHistory` → `store.Plan().AppendPlan` with the exact column set (`:197-207`).
- The harness (cmd/vfverify/d2.go) called exactly these kernel builders over the SCRATCH bars (10,023 stored 1m rows, first 08-19 10:00 CT, last 08-28 15:59 CT) and wrote the row via the REAL `store.Plan AppendPlan` to the SCRATCH DB only. No AutoTrader context was needed — the builders take the minimal inputs the scheduler passes; stated wiring: `kernel.ComputeWeeklyFacts` (kernel/weekly_prompt.go:44), `kernel.BuildWeeklyPrompt` (kernel/weekly_prompt.go:276), `kernel.WeekGoverningMonday` (kernel/weekly_knobs.go:58), `store.Plan().ResolvePlanID/AppendPlan` (store/plan.go:98/248).
- Week check [A]: `WeekGoverningMonday(now)=2026-08-31 — MUST be 2026-08-31: true` (e11). Deadline `2026-08-30 16:30 CT`; rehearsal ran 14:19 CT (before the real one — the live scheduler still owns tonight's read).
- Call: `attempt 1/2: err=<nil> · elapsed=166.8s · raw bytes=472 · finish_reason="stop" · resolved_model="deepseek-v4-pro"` → **ACCEPTED r1-r6 clean on attempt 1 — no retry consumed** [A]. Dispatch total = 2 real calls (1 session + 1 weekly), exactly the budget.

## Facts the model saw [A] (e11, e8)

- 2 completed weeks (08-17 first, 08-24 **outside**), thin_history=true, price=29509.50.
- Refs: weekly_open **0.00** (current week hasn't opened — Sunday pre-open), PWH 29811.75, PWL 28947.75, PWC 29509.50.
- NWOG: 1 gap (born 08-24, CE 29381.12, filled). IPDA: 20d/40d/60d all "insufficient history".
- WeeklyRefSet = 2 draw-eligible refs: [PWH 29811.75, PWL 28947.75] (IPDA insufficient → none; NWOG filled → edges excluded; weekly_open excluded by construction).

## W-2 desk review, rule by rule

**1. Bias from Tier-A evidence ONLY — quoted [A].** The doc is `bias=neutral`. Narrative (single line, full quote):
> "Outside weekly candle swept both prior extremes and closed back inside. No break-and-hold at PWH or PWL resolves direction. Treat as balance until a 1h close beyond either extreme."
- Evidence cited is exactly Tier-A (c) 3-week structure tags — "Outside" matches the facts table row `2026-08-24 … outside` — and Tier-A (b) PWH/PWL break-AND-HOLD vs sweep-reject: the doc explicitly states NO break-and-hold occurred ("swept both prior extremes and closed back inside") [A].
- Tier-A (a) price-vs-weekly_open was NOT usable: weekly_open = 0.00 (the new week had not opened at read time) — not cited, and not needed [A].
- Neutral follows from the facts: an outside week + no break-hold = balance. No Tier-B/C evidence is smuggled into the bias.

**2. Conviction FORCED low by the depth guard — validator r6 proof [A].**
- `facts.ThinHistory=true` (completed weeks = 2 < 4) and `doc.Conviction = "low"` → r6 recompute line: `r6 thin-history depth guard: facts.ThinHistory=true → doc.Conviction MUST be "low" — doc says "low" → true` (e11). The doc says low; the validator would have rejected anything else (`kernel/weekly_prompt.go:225` r6).

**3. draw.px matches a real computed reference (±1 tick = 0.25) [A].**
- `draw: {name: "PWH", px: 29811.75}`. Recompute: nearest ref 29811.75, distance **0.000** → within ±0.25 ✓ (r3 line, e11).
- Neutral-bias rule 4 compliance: "neutral bias → the nearest untested pool on either side" — hand recompute: |PWH−price| = 29811.75 − 29509.50 = 302.25 pts; |PWL−price| = 29509.50 − 28947.75 = 561.75 pts → PWH IS the nearest side ✓ [A].

**4. Invalidation present with basis TF [A].**
- `invalidation: {px: 28947.75, basis: "1h close beyond 28947.75"}` — px > 0, basis non-empty, TF token "1h" present (r2 line: both true). Px = PWL (the other extreme), so the invalidation mirrors the draw symmetrically — coherent with the neutral read.

**5. Narrative ≤3 lines and no day-of-week tokens [A].**
- `r4 narrative lines: 1 (≤3 → true)`; `r5 day-of-week tokens: false` (e11). The narrative contains no weekday/month reasoning.

**6. NWOG/IPDA cited ONLY as draw material, never bias evidence — r-check [A].**
- The bias reasoning quotes (see 1) mention NEITHER NWOG nor IPDA — the only bias evidence is the outside tag + sweep-reject ✓ (rule 2 compliance).
- `weekly_levels` includes `NWOG CE 29381.12` as a LEVEL row — levels are draw/target material, not bias evidence (legal; the CE is drawn verbatim from the facts' NWOG table) [A].
- IPDA is "insufficient" on all three horizons, so nothing IPDA-flavored exists anywhere in the doc ✓.
- Draw itself = PWH (an allowed draw reference), not an NWOG/IPDA row ✓.

**7. Storage-path fidelity (the scratch write) [A].**
- Row: `plan_id="2026-08-31:WEEKLY:8d5c8af5_…_1781246265" version=1 trade_date=2026-08-31 session=WEEKLY trigger="sunday_weekly_read" lifecycle=active model="deepseek-v4-pro" prompt_hash=e68b4814a198e257… doc_len=741 created_at=2026-08-30T19:22:10Z` (e11) — the same column set the scheduler writes (`trader/auto_trader_weekly.go:197-207`).
- Doc stamped: `facts_hash = e68b4814a198e2578907ee4dd9f6cb5d2121b704eb6ecceef1ad62a0a551c117` (= WeeklyFactsHash of the rendered facts sections, e10) and `thin_history: true` ✓.
- Scratch plans 159 → 160; scratch WEEKLY rows 0 → 1. LIVE DB untouched (see e6).

## W-2 verdict

The weekly read rehearses CLEAN end-to-end: correct governing week, correct facts (2 weeks / thin / NWOG / IPDA-insufficient), Tier-A-only neutral bias, depth-guard-forced low conviction, draw exact on a computed reference, invalidation with basis TF, 1-line narrative, NWOG/IPDA used only as level material, and the row written through the REAL store path with the scheduler's exact columns. Tonight's 16:30 CT live read will be the first real WEEKLY row in the LIVE DB (count there = 0, e6).
