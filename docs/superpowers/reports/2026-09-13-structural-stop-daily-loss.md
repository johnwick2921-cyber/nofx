# Automatic structural stops; owner-controlled daily loss

[A] The owner clarified **“i mean daily loss”** after the initial structural-stop
SIM boot. The agent had interpreted the earlier owner response as requiring an
additional per-trade dollar cap and refused all fades when that new cap was
blank. That interpretation was wrong. The correction removes the extra cap;
the system continues to choose the stop and target automatically from structure.

## Scope and unchanged risk controls

[A] Branch `fix/structural-stop-daily-loss`, claimed by
`structuraldaily-22fba7ca/Codex[unlisted]`, starts from current dev
`587148a6c52f5f4379c82cfa22a3f28ce7f44097`.
The extra `max_trade_loss_usd` config field, registry entry and UI input are
removed, as are `risk_cap_missing` / `risk_cap` admission decisions. Modeled
one-contract dollar exposure remains recorded using the instrument's point value.
The record reader retains its optional historical cap field for old evidence.
Old cap-refusal counters, if present, remain in the boot line's `other` total;
there is no new cap requirement hidden behind them.

[A] The structural prices, measured buffer, first eligible zone, net-gain test,
unchanged 2R policy, one-setup selection and existing entry gates are preserved.
Daily loss is enforced by the existing `EntryGate` daily-force-flat resolver;
this wave does not replace it with a per-trade proxy or change its switches.
The Guide, boot display, Rulebook, system map and checklist class 126 now state
the corrected owner intent. Checklist class 1 (self-imposed caps) also applies.

[A] Read-only bound-strategy measurement at acceptance: **one bound row**;
`daily_loss_limit_usd=450`, `guardrails_enabled=false`,
`daily_loss_enabled=false`, `max_trade_loss_usd` absent. Thus the saved $450
limit is currently **disabled**, not enforced. No live settings, account bindings,
owner values or switches are changed by this correction. Config keys and binding
were read through `traders.strategy_id`, never an unrelated active strategy.

## Behavioral evidence

[A] RED before production edit: the actual production-arm fixture computed
entry 29010, entry zone [29000,29010], buffer 5, stop 28995, target 29120,
modeled loss $34 including costs, then recorded quantity 0 with
`risk_cap_missing`; no arm row existed. The initial attempted reproduction still
passed because JSON decoding an omitted field retained a prepopulated fixture
cap. Clearing that field explicitly produced the quoted behavioral RED.
[RED](2026-09-13-structural-stop-daily-loss-evidence/red-missing-cap.log).

[A] GREEN: the same arm-cycle long/short fixtures now succeed without any per-trade
cap. A separate production-arm test trips the existing daily-loss resolver and
still gets quantity 0, `entry_gate` / `daily_force_flat`, with stop 28995 and target
29120 unchanged. Dollar exposure is still $16 at a $2 point value and $160 at $20
for the same six-point risk plus two-point cost fixture. This is contract arithmetic,
not a claim that MNQ calibration transfers to NQ.
[Targeted tests](2026-09-13-structural-stop-daily-loss-evidence/green-targeted.log).

[A] Two mutations ran through `scripts/mutate.sh`, with edits confirmed and
mutants building: reintroducing the extra cap refusal is KILLED by the automatic
arm test; bypassing the daily-force-flat leg is KILLED by the production daily-trip
test. The latter mutation is restored; no production daily guard code is changed.
[Cap mutant](2026-09-13-structural-stop-daily-loss-evidence/mutant-extra-cap.log),
[daily guard mutant](2026-09-13-structural-stop-daily-loss-evidence/mutant-daily-guard.log).
The UI fixture edits the existing daily-loss field and verifies that the added
per-trade input is absent.
[UI test](2026-09-13-structural-stop-daily-loss-evidence/ui-targeted.log).

## Research and deployment limits

[A] No level detector, zone merge, target selection, stop buffer, fill model or
post-entry exit is changed. The original **geometry-only** replay already separated
the missing-cap configuration refusal from candidate geometry. Its negative
expectancy remains the research result; removing a mistaken admission requirement
is not new evidence of profitability. Original artifacts must be reproduced at
their pinned source revision because their configured-admission labels described
the superseded cap policy.

[A] At the start of this correction the running binary was
`4127979f2fcc5615f4e8b17540f7aba74bb4ea92`, booted with integrity/goldens PASS,
MD5 `300dc70535f708be9f6278d252124ba4`. The book was flat and the weekend gate
reported its next open as Sunday 17:00 CT. First real composition/refusal proof
awaits an open-market cycle; fixtures are not live proof. The owner's deployment
GO and explicit market-closed A7 exception remain recorded. Corrected cutover
still requires fresh gates, backups, clean build, marker and observed boot.

## Source freshness before correction

- `docs/superpowers/AUDIT-CHECKLIST.md`: `4127979f merge: structural-stop candidate with verified CI corrections and class 126`
- `docs/superpowers/SYSTEM-MAP.md`: `c9147d54 fix: withhold placement when geometry refusal cannot retire an old arm`
- `docs/superpowers/VL-TRADING-RULEBOOK-v1.md`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `docs/superpowers/reports/2026-09-12-structural-stop.md`: `587148a6 docs: record structural-stop SIM boot and pending market-open proof`
- `store/knob_registry_table.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `store/strategy.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `store/structural_geometry.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `store/structural_geometry_test.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `trader/structural_fixture_test.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `trader/structural_geometry.go`: `c9147d54 fix: withhold placement when geometry refusal cannot retire an old arm`
- `trader/structural_geometry_boot.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `trader/structural_geometry_boot_test.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `trader/structural_stop_seam_test.go`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `web/src/components/strategy/RiskControlEditor.tsx`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `web/src/guide/content/guards.ts`: `c9147d54 fix: withhold placement when geometry refusal cannot retire an old arm`
- `web/src/guide/content/settings.ts`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`
- `web/src/guide/types.ts`: `8f4790ca release: prepare verified structural-stop binary and UI; hold cutover under A7`
- `web/src/types/strategy.ts`: `540c9e8d fix: compose level-fade risk from frozen zones and refuse invalid geometry`


[A] Frontend verification: **64 files / 430 tests PASS**, TypeScript PASS.
[Full frontend](2026-09-13-structural-stop-daily-loss-evidence/vitest-full.log),
[TypeScript](2026-09-13-structural-stop-daily-loss-evidence/tsc.log).
