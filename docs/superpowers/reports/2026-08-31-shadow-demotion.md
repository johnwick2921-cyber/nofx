# 0C — Shadow Demotion: fvg_entry + breakout_retest

Date: 2026-08-31 CT · Wave: 0C-shadow-demotion · Branch: `0c-shadow-demotion`
Live-replaced rev: `98a9b4cfb479…` → this wave: `RELEASE=<final sha>`

## What shipped

Config-driven condition-status map; enforcement at the ARM SEAM only; inert
"shadowed" ledger state; E8 complete counterfactual rows; resolved-map boot line.

| # | Implementation item | Verdict |
|---|---|---|
| 1 | `ConditionStatus` precedence: session > base > env > defaults | PASS (kernel tests 7.x) |
| 2 | Defaults `{fvg_entry: shadow, breakout_retest: shadow}` | PASS |
| 3 | `SHADOW_CONDITIONS` / `LIVE_CONDITIONS` env composition | PASS |
| 4 | Enforcement at arm seam only (no gate at decision/sizing) | PASS (wire-loopback silence test) |
| 5 | Boot sweep: resting working/armed rows for shadowed scenarios cancelled (signal ids) | PASS |
| 6 | Inert `shadowed` ledger state — invisible to `ListNonTerminal`, visible to `ListForPlan`, never placed | PASS |
| 7 | `telemetry.armsRefusedShadowed` counter + deduped refusal log | PASS |
| 8 | E8 complete counterfactual rows (R/ATR units, time bars, net-of-friction, ambiguous) | PASS |
| 9 | Resolved-map boot line (once per trader per boot) | PASS |
| 10 | Guide callout + pre-registered promotion criterion (GUIDE_CONTENT_LAW) | PASS |
| 11 | Checklist: new classes 30 (GORM alias scan) + 31 (inert-but-visible state) appended same wave | PASS |

## Resolved condition map at cutover (all 9 known conditions)

| condition | resolved status | source |
|---|---|---|
| reclaim | live | default |
| hold | live | default |
| sweep_reclaim | live | default (docketed Sep-9, NOT shadowed) |
| reject | live | default |
| acceptance | live | default |
| breakout_retest | shadow | default |
| fvg_entry | shadow | default |
| breakdown_continue | live | default |
| breakup_continue | live | default |

(If the owner sets `SHADOW_CONDITIONS` / `LIVE_CONDITIONS` or a session/base
override, the resolved map recomputes at next boot; the boot line prints it.)

## Pre-registered promotion criterion (verbatim)

> A shadowed condition returns to LIVE only if, at n >= 30 shadow setups on our
> own tape, its net-of-friction expectancy LOWER CONFIDENCE BOUND exceeds zero.
> Otherwise it remains shadowed, or is deleted at the court's discretion.
> No promotion on narrative… No promotion on a point estimate without its
> interval.

## Evidence at cutover (old binary exposure)

- ASIA 16:30 CT read on the OLD binary (98a9b4cf): armed/working rows authored —
  `<filled after live query>`.
- Post-cutover first natural read: refusal line / E8 counterfactual row —
  `<filled after cutover, or "none occurred — said plainly">`.

## Shipped vs deferred

- Shipped: items 1–11 above.
- Deferred: nothing structural. (C# AddOn from class-27 wave still awaiting
  owner F5 + NT8 restart — separate wave, no C# touched here.)

## Rollback

```
mv nofx-bin.prev.boot nofx-bin && <revert deploy/RELEASE> && kill -9 <PID>
```

## Anything the owner will still see wrong on screen

- `<plain statement; "nothing" if none>`.
