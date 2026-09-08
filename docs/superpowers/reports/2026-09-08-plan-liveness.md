# Plan liveness — measurement STOP, no implementation

Owner dispatch: PLAN LIVENESS, 2026-09-08. Branch: `fix/plan-liveness`.
Session: `plan-liveness-c22ee052/root[unlisted]`.

The dispatch remains authoritative for scope. The owner's follow-up explicitly
requires a STOP and correction if a Section C premise fails at the running rev.
**C1's purported pre-publication death timestamp is cross-version contamination;
C3's assertion that S2 has no timestamp is false. No fixtures, implementation,
build, migration, or deployment were performed.**

## Source and runtime provenance

[A] Initial dev tip: `171bebee5a1d29d7493ce4ac58b96dddf2f40c95`.
Claim verified through `ls-remote`:
`a66a2fe02cc6194aed14d7afb56a3126a2a79258 refs/heads/fix/plan-liveness`.
On resumption, origin/dev had advanced to
`9c2ab980986f5b9d07d7a7c21e833a9659ebe296`; it was merged into the isolated
worktree, preserving the claim history, at
`d2538a1e6eddcb7a415ff0d8aae90eeb4a177879`. The inherited placement changes are
not this dispatch's implementation. The main checkout was not modified.

[A] At 2026-09-08 00:03–00:06 CT, `/api/health` returned
`{"revision":"317388e7ab50","status":"ok","time":null}`. Service PID was
`3201079`; `/proc/3201079/exe` resolved to `/home/hoang/nofx/nofx-bin`.
`go version -m /proc/3201079/exe` returned:

```
vcs.revision=317388e7ab50d9ccb4ee4c8a92036b2ed30959da
vcs.modified=false
```

[A] The basis audit is **not yet on dev**. Its pinned source is
[planner-preparation audit, sections 4, 7 and 8](https://github.com/johnwick2921-cyber/nofx/blob/6095ca58fe5901ba398be374e4f9d3488d0bed6b/docs/superpowers/reports/2026-09-07-planner-preparation-audit/README.md).
The extracted git blob is 77,345 bytes. Its `git log -1 <sha> -- <path>` is:

```
6095ca58fe5901ba398be374e4f9d3488d0bed6b 2026-09-07T23:09:02-05:00 docs: record owner approval for public audit evidence publication
```

The checklist used is `docs/superpowers/AUDIT-CHECKLIST.md`; its latest commit
on the refreshed dev base is:

```
7e0c5527f42bec9beee164773c279e6ce077e964 2026-09-07T23:39:11-05:00 fix: install receipt routing before entries and preserve rejection evidence
```

All source line references below refer to the **running revision**, inspected
with `git show 317388e7ab50:<path>`. Freshness ledger, each from
`git log -1 317388e7ab50 --format='%H %cI %s' -- <path>`:

| File | Latest commit at running revision |
|---|---|
| `trader/invalidation_resolver.go` | `d280540835f2ddfc09a17497df45359084a8f95f 2026-09-03T10:54:48-05:00 feat(invalidation-wired): the system's own verdict refuses the arm; a position states the version it was armed under` |
| `trader/auto_trader_levelstate.go` | `d280540835f2ddfc09a17497df45359084a8f95f 2026-09-03T10:54:48-05:00 feat(invalidation-wired): the system's own verdict refuses the arm; a position states the version it was armed under` |
| `store/strategy.go` | `d3f711617d58176fb33b3bc3beda9b5bb897e19a 2026-09-03T22:33:00-05:00 feat(settings D6): saved → resolved · source, from the shipped resolvers` |
| `trader/entry_gate.go` | `01ce808839becd61120140e178bffa7cbc225d30 2026-09-05T12:12:00+00:00 fix(risk,planner): wire RiskForceFlat and BiasArmWarning — both shipped uncalled` |
| `kernel/scenario_state.go` | `2eaf7ab59ff1cf89ce88d0317d71a3f3390eff74 2026-08-26T15:21:56-05:00 FVG entry model — 5th scenario condition (pure-math play) (#79)` |
| `trader/auto_trader_wake_levels.go` | `fa86029e95fdd4ff82e093a7b721e1fb41c12372 2026-09-03T14:47:08-05:00 fix(wake): a clock seam — the enforcing cutoff made a fixed-fixture test time-of-day dependent` |
| `store/plan.go` | `4e901261778399f2e09e3f7f9b7e5f4168afae8b 2026-09-03T17:40:18-05:00 feat(plan): a lifecycle log, so trigger_reason stops answering the wrong question` |

## C1 — publication confirmed; claimed death time refuted

[A] Live SQLite was opened using `mode=ro`. The case is one version:
`plans.rowid=267`, logical ID `2026-09-07:ASIA:v4`, lifecycle `active`,
trigger `level_event`. Creation is `2026-09-08 04:40:33.408927189+00:00`,
or **2026-09-07 23:40:33.408927189 CT**. It contains exactly two scenarios:

| Scenario | Direction | Trigger anchor | Authored invalidation |
|---|---|---:|---|
| S1 | short | 29753.25 | `5m close above 29761.62 invalidates rejection` |
| S2 | long | 29761.62 | `5m close back below 29753.25 negates breakout` |

[A] The claimed refusal really appears at 23:45:18 CT:

```
entry-gate REFUSED arm ASIA: entry_gate: scenario S1 invalidated at 2026-09-07 20:51 CT (accepted through 29753.25) — price accepted through the level against the trade — it flipped roles
```

**That line does not establish acceptance through 29753.25 at 20:51.**
At 20:51:39 CT the historical log instead says:

```
scenario S1 → ≈invalidated @ 29687.50 (price accepted through the level against the trade — it flipped roles — display-only estimate, never execution-wired)
scenario S3 → ≈invalidated @ 29664.50 (price accepted through the level against the trade — it flipped roles — display-only estimate, never execution-wired)
```

[A] Those anchors match v1, `plans.rowid=264`, published 20:47:57 CT.
The v4 refusal combines an earlier stored timestamp with a different current
anchor. `store/strategy.go:1340–1341` keys the timestamp by trader, plan ID and
scenario ID, **omitting version**. `trader/auto_trader_levelstate.go:256–270`
stamps each evaluated invalidated scenario once and never overwrites a prior
nonempty stamp. `trader/invalidation_resolver.go:64–79` reads that stamp but
supplies the current evaluation's anchor and reason.

[A] `system_config.rowid=190` holds S1's `2026-09-07 20:51 CT` timestamp.
Row 191 holds S3's same timestamp even though v4 has no S3. This is additional
direct evidence that the namespace survives scenario reuse across versions.

[A] The evaluator selects trigger text before invalidation text
(`kernel/scenario_state.go:96–132`) and applies a generic dangerous-direction
acceptance verdict (`:188–210`). It does not parse S1's authored 29761.62 rule.
The resolver evaluates bars windowed since the current plan's birth
(`trader/invalidation_resolver.go:45–50`).

**Correction:** E1's proposed “accepted through 2h49m earlier” live fixture
cannot use this gate line as its evidence. Whether v4 was already invalid by
its *own written conditions* at publication remains **NOT ESTABLISHED**.
The complete write-time validation path and historical tape replay were not
completed after the required STOP.

## C2 — current all-dead premise not reproduced

[A] During the 00:05 CT measurement, one status row,
`system_config.rowid=188`, held:

```json
{"S1":"armed","S2":"invalidated"}
```

Companion metadata row 189 held v4's confirmation references 29753.25 and
29761.62. These are mutable, unversioned status records, not a durable history.
The measurement does not disprove an earlier both-invalidated observation,
but it **does not reproduce “both dead” now**. No tradeable count is inferred
from these heuristic labels. A complete search of all UI/count surfaces is
not completed.

## C3 — lifecycle count confirmed; absence of timestamps refuted

[A] `plan_lifecycle_log` contains **n=11**, IDs **1–11**, all `active` or
`dormant`. Latest is ID **11**, v3, dormant by flip at
`2026-09-07 23:33:19.779859054-05:00`.

[A] There are already persisted scenario invalidation timestamps outside that
table: S1 at row **190**, S3 at row **191**, and **S2 at row 197**, whose value
is **`2026-09-07 23:45 CT`**. These stamps lack the required version, condition,
price and cause record, but scenario death is not exclusively log text and
S2's stored timestamp is not absent.

[A] `kernel/scenario_state.go:235–265` evaluates all scenarios. The gate resolver
also calls that full evaluator before selecting the cited ID. The recorder
loops over **all** evaluations (`trader/auto_trader_levelstate.go:260–270`).
An entry refusal returns for its cited scenario; that does not prevent the
separate recorder from stamping S2. The claim that S2 has no time *because the
gate stops at S1* is therefore unsupported and contradicted by row 197.

The `display-only estimate, never execution-wired` log text is stale:
the arm gate actually consumes the same evaluator's verdict
(`trader/entry_gate.go:223–250`, `trader/invalidation_resolver.go:49–50`).

## C4 — four suppressions confirmed, prices and governing check corrected

[A] The narrowly selected journal interval 23:39–23:49 CT contains these
**n=4** level-wake suppressions, identified by their exact timestamps:

| Timestamp CT | Close | Seated level | Elapsed printed | Refusal |
|---|---:|---:|---:|---|
| 23:41:18 | 29746.25 | 29708.25 | 4m | `wake_min_interval_min (30m)` |
| 23:43:18 | 29746.25 | 29708.25 | 6m | `wake_min_interval_min (30m)` |
| 23:45:18 | 29758.50 | 29708.25 | 8m | `wake_min_interval_min (30m)` |
| 23:47:17 | 29758.50 | 29708.25 | 10m | `wake_min_interval_min (30m)` |

All identify seated `OB(bear)·4h (HTF)`. The first two do **not** print the
dispatch's claimed 29758.50 close. None of these four cites the cutoff.

[A] `trader/auto_trader_wake_levels.go:274–277` returns at this minimum-interval
check, before the class-47 cadence decision and fast-market measurement at
`:289–320`. The latter's bypass handling is at `:329–343`, followed by its
separate cooldown refusal. The two cooldown concepts must be distinguished
when specifying the exhaustion bypass. The claimed earlier 1.9×ATR event and
governing boot line were not verified before the STOP.

## C5, implementation and live proof status

The exhaustive running-revision wake inventory remains **UNVERIFIED**. No
conclusion is claimed from an incomplete search. The STOP was triggered by
the reproduced C1/C3 contradictions, not merely the missing unmerged audit.

No RED/GREEN pins, mutations, goldens, full Go suite, vitest, tsc, build,
checklist-number allocation, main-tree merge, RELEASE change, binary swap,
process signal, migration or live proof occurred. No new reject, wake trigger,
gate change or durable death recorder exists from this dispatch. The deployed
behavior is unchanged. There is no implementation rollback to perform.

**What the owner still sees wrong:** a gate refusal can attach an older
version's death time to a newly authored scenario; the “display-only” label
misstates an execution-wired heuristic; there is no new versioned death history
or tradeable-count surface from this work. No exhaustion wake or born-dead
refusal has been implemented or proven live.

**Required owner correction before resuming:** distinguish (1) versioned
recording of the existing evaluator's verdict from (2) D3 validation of the
scenario's own authored invalidation. Do not turn the contaminated 20:51 stamp
into a born-dead fixture or treat an anchor heuristic as the authored condition.
Confirm that exhaustion bypasses the minimum-interval throttle observed here
as well as the separate class-47 cooldown, while retaining cutoff and budget.

This report is preserved on the dispatch branch for review. It is not yet
merged to dev; the dispatch is paused under the owner's explicit STOP rule.
