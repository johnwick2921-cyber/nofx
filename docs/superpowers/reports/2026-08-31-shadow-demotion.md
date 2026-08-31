# 0C — Shadow Demotion: fvg_entry + breakout_retest

Date: 2026-08-31 CT · Wave: 0C-shadow-demotion · Branch: `0c-shadow-demotion` (merged → `dev`)
Live rev at park: `98a9b4cfb479197f55047b31f6cdacc1b565ec85` (PID 1391022) · Staged build rev: `7004a7f1f7266a3d8c354afc7ee27f05b5fda2a4`

**STATUS: STAGED-AND-GREEN — CUTOVER ON HOLD pending the owner's explicit GO.**
No unattended deploys (canon 2026-08-27 + dispatch 9.6). The owner was not
available to ack; nothing was swapped and the live bot is untouched.

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

## Evidence at park (all flat-gate legs quoted fresh)

- ASIA 16:30 CT read on the OLD binary: **zero rows authored** —
  `armed_orders WHERE created_at > '2026-08-31 16:00:00'` → empty; no
  fvg_entry/breakout_retest exposure exists or existed under the old binary.
- DB OPEN = 0: `trader_positions` open=0 (576 CLOSED) · `trader_orders`
  open=0 · `armed_orders` state IN ('armed','working') → empty.
- API positions: `[]` (GET /api/positions, Sim101, 17:18 CT).
- NT8 AddOn snapshot ×2: `positions snapshot account=Sim101 count=0` and
  `account=SimAccount1 count=0` (journal 17:16:21 CT).
- Open-orders endpoint: `[]` (GET /api/open-orders?symbol=MNQ, Sim101,
  17:18 CT).
- Window check: 17:13+ CT (forbidden 16:45–17:10 window already passed).
- Owner frontend active (equity=52216.00, pnl=0.00 at 17:16:38 CT).

## Cutover runbook (when the owner gives GO)

1. `sudo`-free swap: `mv ~/nofx/nofx-bin ~/nofx/nofx-bin.prev.boot` then
   `cp ~/nofx-staged/nofx-0c-bin ~/nofx/nofx-bin` (binary stamp
   `vcs.revision=7004a7f1f7266a3d8c354afc7ee27f05b5fda2a4`, modified=false).
2. `kill -9 <PID>` — SIGKILL; systemd `Restart=on-failure` relaunches the
   new binary.
3. Boot checklist within 90s: rev `7004a7f1` · `🔐 BOOT INTEGRITY OK` ·
   goldens PASS · `🔬 conditions: live […] · shadow [breakout_retest
   fvg_entry]` · per-trader resolved line at first arm cycle.
4. Post-boot: flat-gate re-quote (all four legs) + first natural read proof
   (refusal line + E8 counterfactual row; say plainly if none occurred).

## Shipped vs deferred

- Shipped: items 1–11 above.
- Deferred: nothing structural. (C# AddOn from class-27 wave still awaiting
  owner F5 + NT8 restart — separate wave, no C# touched here.)

## Rollback

```
mv nofx-bin.prev.boot nofx-bin && <revert deploy/RELEASE> && kill -9 <PID>
```

(After the swap, `nofx-bin.prev.boot` holds `98a9b4cfb479…` — the current
live rev — so a single `mv` back + `kill -9` restores exactly today's state.)

## Anything the owner will still see wrong on screen

- Nothing new until GO: the live bot still runs `98a9b4cfb479` with the old
  condition handling (fvg_entry / breakout_retest arms are still placeable).
  The guide page now describes the 0C shadow rule while the running binary
  does not enforce it yet — the drift banner will flag this by design
  (GUIDE_BUILT_REV `7004a7f1…` vs health revision `98a9b4cf…`) until cutover.
- No 0C artifacts are in the live DB; the 16:30 ASIA read left nothing.
