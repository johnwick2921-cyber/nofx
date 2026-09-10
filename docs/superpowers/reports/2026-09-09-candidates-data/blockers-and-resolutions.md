# W3 — blockers found by the scout, and how each is resolved

Eight read-only scouts mapped the surfaces. Two findings change the wave's shape; both verified by me
directly, not taken on the scout's word.

## B1 (CRITICAL) — the scoring golden pins the FULL `ScoredLevel`, byte-for-byte

`kernel/stage_a_parity_test.go:21-46` marshals `struct{ Seated, Pool []ScoredLevel }` over **64
fixture combinations** (4 grades × 4 caps × 4 freshness) and byte-compares to
`kernel/testdata/stage_a_score_legacy.json` — **685,898 bytes, 2,680 `"role"` keys**.

```go
seated, pool := ScoreLevelsMinGradeFull(stageAParityLevels(), 30000, 300, func(DetectedLevel) string { return fresh }, cap, 1.5, grade)
results = append(results, struct{ Seated, Pool []ScoredLevel }{seated, pool})
data, err := json.Marshal(results)
```

**Consequence:** adding ANY exported field to `ScoredLevel` (merged names, map role, projection,
entry-candidacy) changes the marshalled JSON and breaks E7's required-empty golden diff.

**RESOLUTION — the new data never touches `ScoredLevel`.** D2/D1/D6 build a **render-time view**
(`MapCandidate`) from `[]ScoredLevel` at the point of rendering. `ScoredLevel` is untouched, the
parity golden stays byte-identical, and E7 passes by construction rather than by re-baselining.
This is also the strictest possible reading of the owner's ruling ("presentation and ordering only").

## B2 (CRITICAL — a hazard the dispatch did not anticipate) — D4 reordering would corrupt owner edits

`web/src/components/plan/EditSheet.tsx` patches plan levels **by array position**:

```ts
:136   [{ op: 'replace', path: `/levels/${levelIndex}`, value }],
:188   [{ op: 'remove',  path: `/levels/${levelIndex}` }],
```

`levelIndex` is the row's index in the rendered list. **If D4 reorders the list the owner is looking
at, an in-flight edit replaces or DELETES a different level than the one clicked** — silent
data corruption of the plan document, and `remove` is not recoverable from the card.

**RESOLUTION — D4 orders the ENTRY SHORTLIST ONLY.** `plan.doc.levels` keeps its existing order and
its existing indices; the reachability ordering applies to the shortlist view (and the model's table),
which carries no edit affordance. This satisfies D4 ("the entry shortlist orders by reachability")
and D1 ("the map is kept whole") without touching the array EditSheet indexes. **No reorder of any
array the card offers an edit control on.**

## B3 — D1's headline is already law: cite class 93, do not re-file it

`docs/superpowers/AUDIT-CHECKLIST.md:2776` already contains **verbatim** "Exclusion is not
invalidation." Filing a new class restating it would re-file 93 and claim another lane's work (A24).
**RESOLUTION:** the new class covers **D3's candidacy refusal** only, and cites 93 for the principle.

## B4 — the A16 census: the dispatch's "highest is 91" does not reproduce

The checklist has **no markdown tables** (`grep -cE '^\|'` → 0), so the dispatch's census command
returns zero rows. Using the two-format census the file itself mandates at `:2568-2573`, the highest
occupied class is **94** on origin/dev, not 91. Also: `:13-14` still reads "Highest occupied class:
**53**" — 41 classes stale — which is the likely origin of a wrong premise. Classes **75, 76 and 77
each appear twice** (once per format); A16 forbids repairing another lane's numbering.
**RESOLUTION:** number AT MERGE from a fresh two-format census, quote it, touch no existing entry.

## B5 — D7's per-read fields cannot be known at boot

`detected` / `merged` / `entry-candidates` / `no-target refused` / `projections` are all **per-read**;
at boot no planner read has happened. Printing `0` would assert a measurement that was never taken
(canon 49; `kernel/detector_d1prime.go:270-271`).
**RESOLUTION — canon already answers this:** CLAUDE.md, *"Boot lines are READ, never literal, and a
field the process cannot know yet prints `n/a`."* The boot line prints `n/a` for the per-read fields
and real values for the resolved ones (`cap`, `order`, `pwh/pwl seatable`); the counts are emitted on
the **per-read** map line. `cap=<n>` is read as the resolved `max_levels`, not `CandidatePoolCap`.

## B6 — practical: no `node_modules` in the worktree

`/home/hoang/nofx-cand/web` has none, and the main tree is deploy-only (A2b), so vitest cannot run
here yet. **RESOLUTION:** `npm ci` inside the worktree before the A12 vitest run — my own worktree,
never the main tree.

## B7 — `GUIDE_BUILT_REV` is stamped at cutover, not in the wave commit

`web/scripts/stamp-guide-rev.sh:19-23` reads the revision from a **running** binary via `/api/health`
and refuses to guess. Per the boot-5 order this is correct: build binary → boot it → stamp → rebuild
`dist`. **RESOLUTION:** the Guide section ships in the wave commit; the rev stamp + `dist` rebuild
happen in the cutover sequence (A4), not before.

## Pre-existing drift found in scope, NOT repaired by this wave (recorded only)

- `GuidePage.tsx:491` renders the literal `12 sections` while `GUIDE_SECTIONS` holds **14** (this
  wave makes 15) — a literal where a read value belongs.
- `faq.ts:7` says "the fourteen questions" while `faq.ts` holds **19**.
- `glossary.ts:122-125` still defines "Thin side" though `faq.ts:106-109` and `levels.ts:123` both
  record it as REMOVED on 2026-08-31.
- `SYSTEM-MAP.md:51` says levels "seat up to 8 per side" — the cap is a **TOTAL**
  (`levels_score.go:603-605`), and 8 is the package default while the bound strategy resolves 12.
- `SYSTEM-MAP.md:86` carries five stale line refs (proximity `:414`→`:423`, confluence `:415-418`→
  `:427`, cluster tolerance `:678-685`→`:719`).
- `kernel/levels_score.go:1148-1150` and `kernel/planner_prompt.go:507-509` both fabricate an
  uncomputed role as `react_zone` (`if role == "" { role = string(RoleReactZone) }`) — a class-49
  fabricated value. The new map-role column will print `n/a`, and the existing column is left alone.
