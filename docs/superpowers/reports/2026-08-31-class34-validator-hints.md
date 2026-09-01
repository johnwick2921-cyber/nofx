# CLASS 34 — Validator hints must name only legal conditions

Date: 2026-08-31 CT · Wave: class34-validator-hints · Branch: `class34-validator-hints`

## The bug (tonight's evidence)

Both ASIA chains failed identically:
- Reject: `S3 breakdown_continue: a close came back across 29517.00 — the breakdown is void; author a reject/retest play instead`
- Model authored condition `reject_retest`
- Parse/schema rejected: `scenario[2].condition "reject_retest" invalid`
- Repeated identically in the reset chain → both chains FAIL-CLOSED. The model complied with the hint and was punished for it.
- Compounding: `breakout_retest` is SHADOWED (0C), so the hint steered toward either a nonexistent or a demoted condition.

## Before/after of every changed message

| Site | Before | After |
|---|---|---|
| `kernel/breakdown_continue.go` reclaimed reject | `… the breakdown is void; author a reject/retest play instead` | `… the breakdown is void; author a \`reject\` play instead (do NOT combine condition names; \`reject_retest\` is not a valid condition)` |
| `kernel/breakdown_continue.go` displacement reject | `… not a displacement move, author a normal reject/retest play instead` | `… not a displacement move, author a normal \`reject\` play instead (do NOT combine condition names; \`reject_retest\` is not a valid condition)` |
| `kernel/planner_repair.go` BREAKDOWN-CONTINUE LAW | `… author a reject/retest play instead of breakdown_continue.` | `… author a \`reject\` play instead of breakdown_continue (do NOT combine condition names; \`reject_retest\` is not a valid condition).` |
| `kernel/planner_repair.go` ARM-SPLIT LAW / ENTRY-LAW CONFIRM LAW | inline literals | registry constants `RepairArmSplitLaw` / `RepairEntryConfirmLaw` (text unchanged — they already name only `sweep_reclaim` / `breakdown_continue`) |
| `kernel/plan_doc.go` arm-legs reject | inline `(the split entry is the sweep_reclaim contract; …)` | registry constant `ArmLegsSplitContract` (text unchanged — names only `sweep_reclaim`) |
| `trader/auto_trader_planner.go` reject block | `Fix ONLY this defect, keep the rest structurally identical.` | now also appends `\nValid conditions: [<resolved live list>] (use exactly ONE token from this list; do NOT combine condition names).` |

## The guard

- `kernel.ValidatorHints()` — registry of all 6 hint sites, each declaring the
  condition tokens its text names.
- `kernel.ValidateValidatorHints()` — every token must be in `KnownConditions()`
  AND resolve LIVE under defaults. **The table test IS the guard**
  (`kernel/validator_hints_test.go:TestValidatorHintsNameOnlyLegalLiveConditions`),
  re-run at boot (`kernel/levels_volume_boot.go` logs
  `🧪 validator hints: 6 sites — every condition token legal + live (class 34 guard)`
  or an ERROR if broken).
- The resolved live list the guard checks against (defaults, fixture-verified):
  `[acceptance, breakdown_continue, breakup_continue, hold, reclaim, reject, sweep_reclaim]`
  (the 9-enum minus the 0C shadowed `fvg_entry`, `breakout_retest`).
- The reject block's `Valid conditions:` suffix is computed from the SAME
  resolver with the trader's base + session override maps + env
  (`kernel.ResolvedLiveConditions`) — class-8 style: resolved, never literal.

## Tests (all green)

- `TestValidatorHintsNameOnlyLegalLiveConditions` — the table guard.
- `TestBreakdownHintsNameRejectNotComposite` — tonight's reproduction at the
  hint level: names `` `reject` ``, forbids the composite, never says
  "author a reject/retest".
- `TestResolvedLiveConditionsExcludesShadowed` — live vocabulary sorted,
  shadowed excluded.
- `TestLiveConditionsLine` — the suffix rendering.
- `TestClass34RejectBlockNamesOnlyLiveConditions` — end-to-end reproduction:
  breakdown-void reject → verbatim reason kept + suffix lists only legal live
  tokens.
- `TestClass34RepairLawsNameOnlyLegalLiveConditions` — repair excerpts carry no
  composite instruction; registry validates.
- Goldens PASS · `go build ./...` OK · `go test ./...` green · tsc 0 · vitest
  36 files / 292 tests.

## Stop-lines honored

No enum changes · no new conditions · no validator LOGIC changes — message text
plus a guard only. The planner's output-contract schema still lists all 9
conditions (0C: authoring shadowed conditions is allowed; the ARM seam refuses
them), and the prompt contract was untouched.

## Checklist

Class 34 appended (class 33 is unoccupied — quoted the current highest before
appending: 32).
