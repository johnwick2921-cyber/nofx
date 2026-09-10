# W1 — THE EPISODE CONTRACT

**Branch** `fix/episode-contract` · **base** `origin/dev` @ `757eb578` · merged, not rebased (see §F.6)
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

6. **This branch is merged onto dev, not rebased, and that was forced.** Four
   rebases in, a force-push was refused by the classifier. I reconciled with a
   `-s ours` merge after verifying the tree diff was additive-only (2,262
   insertions, 11 deletions, no file lost). That merge re-parented the superseded
   pre-rebase commits, so the NEXT `git rebase` tried to replay all of them and
   conflicted on files it had already resolved. I aborted it and merged dev in
   instead. The branch therefore carries both renderings of its own history. It is
   uglier than a clean rebase and it is on the record here rather than hidden: the
   tree is correct, the duplicated commits are content-identical, and no work was
   lost — but a reviewer reading `git log` will see each early commit twice, and
   should read the merge commits for why.

## G · TWO DEFECTS FOUND, NEITHER MINE, ONE MASKING THE OTHER

Found while verifying my Guide change. **Filed, not fixed — A31 scopes this wave,
and both belong to the rebrand lane.**

### G1 — the FE suite is red at the version the lockfile pins, green at the one the main tree has

**This section was wrong in the first two drafts and is corrected here.** I wrote
that the suite "fails everywhere, main tree included". It does not. A peer
(nofx-8e) measured it GREEN in the main tree and challenged the finding. They were
right about the observation and wrong about the cause; I was right about the
observation in my tree and wrong about the cause. Neither of us was measuring
badly. **We were running different versions of vite.**

**The controlled experiment** — same worktree, same commit, same config, same test
file, only the vite version swapped, `--no-save` so nothing else moved:

| vite | result |
|---|---|
| **6.4.1** | `Test Files 1 passed · Tests 9 passed` |
| **6.4.3** | `Test Files 1 failed · Tests 9 failed` — `Error: Denied ID /…/branding/product.txt?raw` |

That is causation, not correlation: one variable, both directions, reproduced.

**Which version is the repo's?** `web/package-lock.json` is TRACKED and pins
**6.4.3** — the failing one. `web/package.json` declares `"vite": "^6.0.7"`, so the
caret range admits both. The main tree's `node_modules` holds **6.4.1**: a stale
install that predates the lockfile move and has not been reinstalled since.

So the correct statement is the inverse of my first draft:

- **The suite is red at the dependency state this repo actually declares.**
- The main tree's green is an artifact of an install that no longer matches the
  lockfile.
- Anyone running `npm ci` — a fresh clone, a new lane's worktree, the A4
  clean-clone deploy path — gets 6.4.3 and gets 12 red files.

**Mechanism**, now measured rather than assumed. `web/src/constants/branding.ts`
(added by `7d7be486`) does `import productName from '../../../branding/product.txt?raw'`.
Vite's workspace root here is `web/`, confirmed by asking vite itself:

```
searchForWorkspaceRoot('/…/web') = /…/web
```

because the ancestry carries no root marker — no repo-root `package.json`,
`pnpm-workspace.yaml` or `lerna.json` in any tree (and `.git` is commented out of
vite's `ROOT_FILES`, so the worktree-vs-clone distinction I first reached for is
irrelevant). `../branding/` is therefore outside the allow root. 6.4.3 enforces
that on this import path; 6.4.1 did not.

**Blast radius:** 12 test files fail to load. Zero assertion failures. And the
count that matters: **354 tests collected red versus 414 green — 60 tests silently
do not exist**, while the file total reads 58 either way.

`npm run build` is **unaffected at both versions** — rollup does not apply
`server.fs.allow`. Verified at 6.4.3: `✓ built in 4.51s`. **The cutover is not
blocked**, and a build-only CI would never see any of this.

**Fix** is one line in `web/vitest.config.ts`:

```ts
server: { fs: { allow: ['..'] } },
```

Verified at 6.4.3, then reverted: 12 red files → 1, 344 → 413 passing.

**The finding underneath the finding.** Two lanes ran "the suite" on the same
commit, got opposite answers, and each correctly believed their own measurement.
A caret range plus a tracked lockfile plus long-lived `node_modules` directories
means **"the suite passes" is not a property of a commit** — it is a property of a
commit *and* whenever someone last ran install. Neither number is on the record
anywhere. My first draft asserted a defect in another lane's file on the strength
of a measurement whose environment I had not pinned, which is the same error in
the opposite direction.

### G2 — a tamper-guard that was red and legible for 13 hours in a suite nobody on that wave ran

With G1 unblocked, a 13th failure surfaces that had been invisible:

```
FAIL src/brand-scope.test.ts > preserves deploy/nofx-lock.sh byte for byte
  Dispatch 102 protected file changed: deploy/nofx-lock.sh
```

`web/src/test/brand-scope-baseline.json` pins sha256 of 16 protected files.
**Exactly one has drifted** — so the guard is otherwise doing its job precisely,
and would have caught this on the first commit had it been able to run.

The timeline is the finding, and the order of the two events is the whole point:

| when (CT) | commit | what |
|---|---|---|
| 09-08 18:57 | `66e2c09a` | baseline written. Pin `bcd82c52…` **correct**. Guard GREEN. |
| **09-09 13:33** | **`7d7be486`** | **G1 breaks the runner. The guard is still GREEN at this moment.** |
| 09-09 22:12 | `87ef772d` | lock keeper wave — sha becomes `9bacd04a…`. **Guard goes RED. Nobody hears.** |
| 09-10 07:16 | `963ea975` | keeper process-group fix. Still red, still unheard. |
| 09-10 08:05 | `88d40920` | keeper `/proc` read fix. |
| 09-10 10:26 | `97a6525c` | `/proc` read SIGTERM fix. |
| 09-10 11:43 | `ace51598` | INCOMPLETE-lock wave. |

Current sha256 at dev tip `757eb578`: `46fcbf76…` — against a pin of `bcd82c52…`.

**Corrected after nofx-8e's challenge, and the correction is against my own
framing.** I first wrote that G1 had blinded this guard. In the environment where
8e actually worked — the main tree, vite 6.4.1 — **the guard ran fine and was
plainly RED, naming their file, for the whole 13h31m.** It was not muted there. It
was legible and unread, because that wave ran Go and the lock suite every time and
treated those as "the suite". 8e states this plainly as their own miss, and it is
the more useful reading: a protected-file guard lived in a suite the wave had
decided, without ever deciding, was not theirs.

Both things are true at once, and which one you hit depends on your installed vite:

- at **6.4.1** (main tree): the guard **runs and is red** — a legible signal nobody read
- at **6.4.3** (lockfile-pinned): the guard's file **never loads** — no signal to read

The second is strictly worse, and it is the state anyone gets from a fresh install.

None of the six commits is at fault. Each changed a protected file for good reason,
and the guard exists precisely so a human ratifies that change by updating the
baseline. **The pin has since been ratified** at `e79bf298` by another lane, and 8e
fast-forwarded the main tree to `a8b66cd0` and re-ran: 58 files, 414 tests, green.
I did not touch their baseline and should not have — ratifying another lane's
protected-file change is exactly the conversation the guard exists to force.

8e also checked what my report had not: whether the unread window hid drift in any
**other** protected file. It did not — 1 of 16 drifted, and it was theirs.

## H · VERIFICATION

| gate | result |
|---|---|
| `go build ./...` | OK at merged HEAD |
| `go vet ./store/... ./trader/... ./kernel/...` | OK |
| `go test ./...` | **30 packages ok, 0 FAIL** at merged HEAD |
| `npx tsc --noEmit` | OK |
| `npm run build` | OK — 4.60s |
| `npx vitest run` @ vite 6.4.3 (lockfile) | 12 files red from G1; 354 collected |
| `npx vitest run` @ vite 6.4.1 (main tree's stale install) | green — see §G1 |

**My Guide change is not verified by the suite at the lockfile-pinned vite**,
because 12 files including the guide's do not load there — G1, not my change. Verified instead by: `tsc --noEmit` clean, `npm run build` clean,
and the guide tests passing under the temporary G1 unblock before I reverted it.
Stated here rather than reported as green.

dev moved **five times** under this branch during the wave (`557494c7` →
`c16a182d` → `cefcf08d` → `33e6d008` → `757eb578`); the suite was re-run at each
merged HEAD rather than carried forward — a branch green alone is not green merged.

## I · CLASSES FILED

- **108 — THE UNIT AN EXPERIMENT NEEDS, WHICH THE RECORD NEVER HELD.** Every
  experiment measures value per opportunity; the record held only fills.
- **109 — A CENSUS THAT CANNOT SEE ITS OWN THIRD FORMAT.** The checklist has three
  entry shapes; a two-format census reported the ceiling as 93 while 104 existed.
  Now carries the 105→106→108 renumber chain as its worked example.

**Two more classes are owed from §G and are NOT filed here**, because each belongs
to a lane that owns the file and because I got the first one wrong twice before
measuring it properly:

- **A suite's result is a property of a commit AND an install date.** A caret
  range plus a tracked lockfile plus long-lived `node_modules` means two lanes can
  run "the suite" on one commit and get opposite answers, both honestly. Neither
  the installed versions nor the install date appear in any report. The remedy is
  cheap: print the resolved version of the runner alongside the pass count, and
  fail when `node_modules` disagrees with the lockfile.
- **Pin the suite's own file and test counts.** Red here reads
  `12 failed | 46 passed (58)` and green reads `58 passed (58)` — but the test
  totals are **354 versus 414**. Sixty tests vanish and no number in the default
  output says so. This is the remedy for the general shape I sent 8e and which
  survives all the corrections above: *a test that cannot run reports the same
  colour as a test that passes.* 8e is taking it into the settlement wave's pins,
  and notes it is the same shape as their class 103 — a fallback producing a
  plausible value so the failure behind it stays invisible.
