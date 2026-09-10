# W1 — THE EPISODE CONTRACT

**Branch** `fix/episode-contract` · **base** `origin/dev` @ `33e6d008` · 11 commits
**Scope (A31)** RECORDING ONLY. No rule, threshold, gate, order, plan content, level
score or surface behaviour changed. The wave defines the unit of opportunity and
writes it down.

---

## THE HEADLINE: THE BACKFILL RECOMPUTES ZERO ROWS

The dispatch's premise was that history could be re-read into the new unit. It
cannot. Measured against `data/data.db` on 2026-09-10, read-only:

| state | rows |
|---|---|
| in-era rows examined (since `DayPlanEraStart` = 2026-08-15 00:00 CT) | **4,860** |
| `untouched_pre_era` | 0 |
| `unrecomputable:no_formation` | 4,677 |
| `unrecomputable:no_scenario_link` | 183 |
| **RECOMPUTED** | **0** |

`formed_at_ms` is present on **183 of 4,860 rows — 3.77%**. The research that
motivated this wave claimed 26%. That is not a rounding disagreement; it is a
different claim about the archive, and the archive settles it.

The two gates compose to nothing by construction: the 183 rows that survive the
formation gate are *exactly* the rows that then fail the scenario-link gate,
because **no historical row carries a scenario link at all** — the column is born
in this wave. So the ceiling on any retrospective episode study of the current
archive is zero rows. Not "few". Zero.

**This is the research's claim MEASURED, not the backfill failing.** The backfill
is working correctly: it marks every row it cannot recompute, with the reason, and
invents nothing. A backfill that had "recovered" 4,860 episodes would have been
fabricating 4,677 formation times and 4,860 scenario links.

**What it means for the next wave.** Episode-based evaluation starts from rows
written *after* this boot. There is no historical baseline to compare against, and
any experiment that assumes one is measuring its own reconstruction.

---

## A · THE UNIT

A touch was always the row. What was missing was the ladder above it —
`store/opportunity_outcome.go`:

```
never_reached → reached_declined → confirmed_not_armed → armed_not_filled → filled
```

`OpportunityOutcomeFor` derives the rung from four booleans in a single switch,
highest rung first. **Exhaustive by construction, not by a default branch**: there
is no `default:` that silently absorbs a state nobody enumerated. Adding a rung
means adding a case, and a missing case is a compile-visible gap rather than a
row quietly filed under the lowest rung.

`store/opportunity_close.go` selects on `opportunity_outcome IS NULL` — idempotent
by *predicate*, not by a "done" flag that can drift from the rows it describes.
`factsFor` is called **per row** (`opportunity_close.go:48`); a batch cannot smear
one row's facts across its neighbours.

## B · THE LINK IS A HEURISTIC AND THE COLUMN SAYS SO

Ruling (b) as dispatched. A touch has a *price*; nothing in the plan path stamps a
scenario id on it. So the link is recorded as what it is:

- column `scenario_nearest` — **never** `scenario`, which would read as fact
- `ScenarioLinkBasis = price_proximity`, plus distance in **points and in Δ**
- two anchors inside the band → **NULL**, basis `unresolved:two_scenarios_within_band`
- nothing close → `unresolved:nearest_outside_band`
- no plan at the seat → `unresolved:no_scenario_at_seat`

Ambiguity is recorded as ambiguity. There is no nearest-wins path: a tie does not
become a fact because a tie is inconvenient. A zero Δ leaves `dist_delta` NULL
(`scenario_link.go:90` guards `delta > 0`) rather than dividing.

The band is `kernel.LevelClusterTicks × 0.25` = 3.00 pts — **the map's own cluster
width**, not a new tolerance knob. Two levels the map would have merged are exactly
the two this link refuses to choose between.

`trader/scenario_anchor.go` takes `Confirm.RefPrice` when present, falls back to
`Arm.Entry`, and **counts what it cannot anchor** rather than dropping it silently.
`trader/auto_trader_planner.go` now parses the `latest.Doc` it previously fetched
and discarded — only `.Version` was ever read from it.

**The identity wave comes after.** This is the labelled stand-in, and the label is
the point: a downstream reader can tell a heuristic from a fact without asking
anyone.

## C · THE ATTAINABLE ENTRY — MEASURED BEFORE ASSUMED

`store/attainable_entry.go`, in precedence order:

| basis | meaning |
|---|---|
| `observed_fill` | MEASURED — a fill exists |
| `resting_limit:assumed_fill_at_entry` | ASSUMED, and says so |
| `stop_entry:assumed_fill_at_trigger` | ASSUMED, and says so |
| `first_tradeable_after_confirm` | ASSUMED, and says so |
| `none:never_confirmed_never_armed` | there was no entry to attain |
| `not_captured:confirmed_but_no_price_recorded` | there was one; we failed to record it |

The last two are deliberately **different values**. "No entry existed" and "an entry
existed and we lost it" are different facts about the system, and collapsing them
would hide a recording defect inside a legitimate outcome.

**The level price is never a branch.** What the level *said* is not what the tape
*offered*, and no rung falls back to the level price to avoid a NULL.

## D · WHAT IS NOT DONE, DELIBERATELY

- **No scenario identity on the touch.** Ruled to the next wave.
- **No episode table.** The ruling was to extend `touch_outcomes`; a second table
  would have been a second key for the same unit — and a second key is how the
  system ends up with two readers that happen to agree (class 97).
- **No historical terms.** `armed_orders` is a state row **mutated in place**, not
  an event log. Historical terms are unrecomputable by construction and are marked
  `unrecomputable:terms_mutated_in_place` rather than reconstructed from
  `updated_at` — which would have produced a number for every row and a true one
  for none.
- **`trade_excursions` not extended.** It is `UNIQUE(position_id)` — **filled-only
  by construction**. It cannot hold the four rungs below `filled`, which are the
  rungs the wave exists to measure. Wrong key, not a missing column.
- **The ordinal was already true** and was dropped rather than re-implemented.

## E · THE BOOT LINE NAMES ITS RESOLVER

Join key `🎫 episodes:`. Counts are READ, never literal.

```
🎫 episodes: open=2 · closed today=7 (never_reached=3 reached_declined=2
   confirmed_not_armed=1 armed_not_filled=1 filled=0) · backfill recomputed=0
   unrecomputable=4860 · k=3[I] Δ=resolved-per-read
   (kernel.MeanAbsIncrement, the tape's own scale) H=12[I]
```

Δ prints its **resolver**, not a value. Δ is resolved per read; a number printed
once at boot is a literal wearing a measurement's clothes, and would go stale
silently between the boot line and the first decision. `k` and `H` print `[I]` to
mark their env source.

**One honest seam, stated rather than buried:** `store` cannot import `kernel`
(kernel imports store). So `store/episode_detector_scope.go` reads `DETECTOR_K`
and `DETECTOR_HORIZON_BARS` through the **identical env names**. That is two
readers of one source — the closest this dependency direction allows — and
`grep DETECTOR_K` finds both sites, so a drift between them is visible rather
than silent. It is not the single-reader ideal of class 97 and is not claimed to be.

## F · CORRECTIONS TO MY OWN PREMISES

1. **The C1 cross product.** I claimed a join existed, on 1,205 rows. There was no
   join: it was 481 touches × 9 arms × 7 positions. Both the wrong measurement and
   the correction are on the record, because the wrong one is what the method
   produces if nobody checks the row count against the inputs.
2. **`touch_outcomes` has ONE writer**, at `detector_record.go:102`. My earlier
   report said two; the second site (`:115`) writes `candidate_pool`.
3. **`formed_at_ms` is 3.8%, not 26%** — see the headline.
4. **My own SYSTEM-MAP draft overstated the schema.** It claimed *every* added
   column is a pointer. `ScenarioLinkBasis` is a plain `string`, because
   `ResolveScenarioLink` always returns a basis — so an empty basis marks a
   **pre-wave row**, not a missing reading. Caught by grepping my own prose against
   the code before committing. This is class 105's lesson applied to the commit
   that introduced class 105.
5. **My classes 105/106 landed as 108/109 — renumbered twice.** Dispatch 103 took
   105 while this branch rebased; then 106 and 107 went too (a peer's
   generalisation of class 104, and the boot-sweep `cancel_pending` wave). Four
   dev tips in one wave. That is A27 working exactly as written, and class 109
   now carries it as the worked example: a census tells you the ceiling, only the
   merge assigns the number. The alternative — reserving a number at accept — is
   what produced the 75/76/77/92/93 duplicates.

## G · TWO DEFECTS FOUND, NEITHER MINE, ONE MASKING THE OTHER

Found while verifying my Guide change. **Filed, not fixed — A31 scopes this wave,
and both belong to the rebrand lane.**

### G1 — the FE suite has been red repo-wide for ~22 hours

`7d7be486` ("fix(brand): share visible VL names without renaming identifiers",
2026-09-09 13:33 CT, **on dev**) added to `web/src/constants/branding.ts`:

```ts
import productName from '../../../branding/product.txt?raw'
```

That path is **outside vite's workspace root**, so vitest's module runner denies it:

```
Error: Denied ID /…/branding/product.txt?raw
```

**12 test files / 10 tests fail, all from this single cause. Zero assertion failures.**

Bisected: `08f9f88f` (the commit before) → **9 passed**. `7d7be486` and every commit
after → **9 failed**, same file, same test.

**It is not a worktree artifact.** I first hypothesised the linked worktree's `.git`
file; that was wrong — `.git` is commented out of vite's `ROOT_FILES` in this
version. The root resolves via `searchForPackageRoot` to the nearest `package.json`,
which is `web/`. There is **no repo-root `package.json`, `pnpm-workspace.yaml` or
`lerna.json`** in any tree, main included. So the workspace root is `web/`
everywhere and this fails everywhere.

`npm run build` is **unaffected** — rollup does not apply `server.fs.allow`. Verified:
`✓ built in 4.60s`. So the shipped bundle is fine and CI-by-build would never have
caught it.

**Fix is one line** in `web/vitest.config.ts`:

```ts
server: { fs: { allow: ['..'] } },
```

Verified locally, then reverted: **12 files red → 1**, 344 → 413 passing.

### G2 — a tripped tamper-guard that nothing could report

With G1 unblocked, a 13th failure surfaces that had been invisible:

```
FAIL src/brand-scope.test.ts > preserves deploy/nofx-lock.sh byte for byte
  Dispatch 102 protected file changed: deploy/nofx-lock.sh
```

`brand-scope-baseline.json` pins sha256 of protected files. `deploy/nofx-lock.sh`
legitimately changed on dev — `88d40920` and `97a6525c`, the keeper `/proc` fixes —
and the baseline was never updated. Current sha256:
`670a405be9a7c80b866bb20855fbf774227bc4f374152bfc74963265e257387b`.

**The shape is the finding.** A tamper-guard fired correctly and nobody heard it,
because an unrelated import broke the runner that would have reported it. G1 did not
merely break 12 files — it **muted a guard** for 22 hours. A guard that cannot run is
indistinguishable from a guard that passes, and the suite reported the same colour
either way.

## H · VERIFICATION

| gate | result |
|---|---|
| `go build ./...` | OK at merged HEAD |
| `go vet ./store/... ./trader/... ./kernel/...` | OK |
| `go test ./...` | **30 packages ok, 0 FAIL** at merged HEAD |
| `npx tsc --noEmit` | OK |
| `npm run build` | OK — 4.60s |
| `npx vitest run` | 12 files red **from G1 alone**; with G1 unblocked locally, 413/414 pass, the 1 being G2 |

**My Guide change is not verified by the suite**, because the suite cannot run —
G1, not my change. Verified instead by: `tsc --noEmit` clean, `npm run build` clean,
and the guide tests passing under the temporary G1 unblock before I reverted it.
Stated here rather than reported as green.

dev moved **four times** under this branch during the wave (`557494c7` →
`c16a182d` → `cefcf08d` → `33e6d008`); rebased onto each, full suite re-run at the
final merged HEAD — a branch green alone is not green merged.

## I · CLASSES FILED

- **108 — THE UNIT AN EXPERIMENT NEEDS, WHICH THE RECORD NEVER HELD.** Every
  experiment measures value per opportunity; the record held only fills.
- **109 — A CENSUS THAT CANNOT SEE ITS OWN THIRD FORMAT.** The checklist has three
  entry shapes; a two-format census reported the ceiling as 93 while 104 existed.
  Now carries the 105→106→108 renumber chain as its worked example.

Recommended for the rebrand lane, from G1/G2: *a test that cannot run reports the
same colour as a test that passes* — the suite's file count is itself a number that
must be pinned, or a suite can lose 12 files and still look like a suite.
