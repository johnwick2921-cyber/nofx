# S2 — Timeframe-Aware Freshness + Name-Preserving Collapse (DS-101)

- **Lane:** DS-101 · **Branch:** `fix/levels-fresh-by-tf` · **Claim:** `f3d46176` · **Worktree:** `/home/hoang/nofx-ds101`
- **Base:** `git log -1 -- docs/superpowers/AUDIT-CHECKLIST.md` on dev at cut = `f6465143` (dev tip at claim time; spec-freshness: the S2 dispatch spec is a bridge message, not a tracked file — no tracked spec moved after my base).
- **Report:** 2026-09-16 ~19:30 CT. Dispatch due 21:00 CT.

## What shipped (files)

| File | Change |
|---|---|
| `kernel/levels_fresh_by_tf.go` (NEW) | `LevelFreshnessByTF` grader + `normalizeByTFGrade` + `IsHTFFreshTF` |
| `kernel/levels_score.go` | 2 lines: `freshMult(normalizeByTFGrade(fRaw))`, `zoneFreshMult(normalizeByTFGrade(fRaw))` — identity when OFF |
| `kernel/levels_assemble.go` | `DetectHTFLevelsExport` — exported replay wrapper, no production caller |
| `store/strategy.go` | `DayPlanConfig.LevelsFreshByTF bool json:"levels_fresh_by_tf,omitempty"` + `LevelsFreshByTFEnabled()` |
| `trader/auto_trader_dayplan.go` | knob routing in `installLevelStateProvider` (HTF branch only) + `levelTFBars` + `🧮 freshness:` boot line |
| `kernel/levels_fresh_by_tf_test.go` (NEW) | grader units, normalize identity, scoring-equivalence at `scoreLevelsPool`, collapse pin |
| `web/src/guide/content/settings.ts` | knob entry (`HTF freshness by own timeframe`, default OFF) |
| `web/src/guide/content/levels.ts` | freshness paragraph (two-table context) |
| `docs/superpowers/AUDIT-CHECKLIST.md` | `## CLASS NN (assigned at merge) — S2 by-TF freshness` |
| `docs/superpowers/research/2026-09-16-s2-replay/` (NEW) | read-only replay harness + output |

## (a) Knob

`day_plan.levels_fresh_by_tf` bool, default **false** (`store/strategy.go`, method `LevelsFreshByTFEnabled`, nil-safe → OFF). OFF = the grader is never on the path; scores/prompts byte-identical (proved below). No api/strategy.go range validation applies — it is a plain bool with an omitempty zero (the api file validates warnings only, [A] read).

## (b) Own-timeframe freshness

- fRaw production today [A]: `kernel/levels_score.go:462-472` — the injected `freshness` callback (`levelFreshnessFn` → `LevelStateProvider` installed at `trader/auto_trader_dayplan.go:181`, reading W7 persisted `store.LevelState` grades).
- Routing [A]: ONLY `l.HTF && IsHTFFreshTF(l.TF)` enters `LevelFreshnessByTF`; every non-HTF level and every OFF config takes the unchanged persisted ladder.
- Grader: a bar of the level's TF that traded into `[Lo,Hi]` (`b.Low <= hi && b.High >= lo`) at/after origin (`OriginDate`, else `FormedAtMs`) is one test. Grade by count: **fresh (0) / tested-1 (1) / tested-2 (2) / stale (≥3)**.
- Scoring: `normalizeByTFGrade` maps tested-1→b, tested-2→c, stale→done before `freshMult`/`zoneFreshMult` — **the two tables are unchanged** (dispatch's hard line). `Research.Freshness` keeps the S2 display vocabulary.
- Bars source: live ring (`FuturesBarsProvider`, 500) + persisted NT8 history leg for the latest contract (same combination `installNakedPOCProvider` uses), assembled in `levelTFBars`.

## (c) Collapse names — verified, pinned

`collapseLevelClusters` (`levels_score.go:750`) ALREADY keeps names (fix/collapse-keeps-names 2026-09-10): loser labels ride `json:"-" CollapsedNames`, rendered by `namesWithCollapsed` (`map_candidates.go:512`). New pin `TestS2_CollapsePin_PDHAbsorbsEQH4h`: PDH (score 0.5, today-priority) absorbs EQH·4h (score 1.5) → survivor `PDH`, names = [PDH, EQH·4h]. PASS [A]. No completion work needed — the spec's "if partial" branch does not apply.

## (d) Boot line

`trader/auto_trader_dayplan.go` — after the 🗺️ knobs line: `🧮 freshness: fresh=<1m|by-tf> htf×=<%g of kernel.HTFScoreMultiplier>` — both READ (knob resolved + const read), never typed. Example: `🧮 freshness: fresh=1m htf×=1.2`.

## (e) Replay table — TODAY's 16:31 CT G2 detection

Harness `docs/superpowers/research/2026-09-16-s2-replay/harness/main.go` opens `data.db` **read-only** (`mode=ro`), re-runs `DetectHTFLevelsExport` over stored bars, grades each level OFF (persisted `level_state` via `AgedFreshness`) and ON (S2 grader). Output: `out/replay-2026-09-16-1631.txt` (full per-level table).

**Coverage caveat, stated not hidden:** the live 16:31 detection ran over the in-memory NT8 ring; the store holds only ~230×15m bars for MNQ 12-26 (roll ≈09-14). Current-contract-only replay reproduces 50 levels (244 logged). Combined with the previous contract's history (09-26, the same historical-leg pattern the nPOC provider uses — basis-adjacent proxy [B]): **417 levels**. The 244 was produced from a bar universe between those two; the per-level table below is the combined-contract proxy. Sample ids are `label@price` rows in the output file.

**Grade shift (HTF levels with TF∈S2 set, n=339, combined-contract proxy):**

| | fresh | B | tested | flipped |
|---|---|---|---|---|
| OFF (1m ladder) | 239 | 83 | 11 | 6 |
| ON (by-TF) | 104 (fresh) | — | 13 (tested-1) + 9 (tested-2) | 213 (stale) |

Direction of the shift: by-TF grading finds most HTF zones tested 3+ times on their OWN timeframe and grades them **stale** (→ `done` in the unchanged ladder → 0.5 anchor / 0.15 zone multiplier). That is the OPPOSITE direction from "1m-touch over-decays HTF": by-TF is stricter, not laxer. This table is the number S4 measures against; **no recommendation is made beyond the cells** — whether stricter HTF decay helps is S4's measurement, not this lane's call.

## Tests

- NEW RED→GREEN: none were ever RED on the branch (grader was new) — all new tests GREEN at first run: `TestLevelFreshnessByTF_CountsAndOrigin`, `TestLevelFreshnessByTF_NoOriginMeansFresh`, `TestNormalizeByTFGrade_IdentityOnLegacy`, `TestScoreLevels_ByTFVocabScoresLikeCanonical` (production call site `scoreLevelsPool`), `TestS2_CollapsePin_PDHAbsorbsEQH4h`.
- OFF-proof: existing goldens re-run green at this HEAD: `TestEnginePromptGolden`, `TestFuturesPrompt*`, `TestStructuralPromptContract`, `TestCollapse*` [A].
- Full `go vet ./...` + `go test ./...` at HEAD: see final tail (ran 19:20 CT, background) — quote below in the DONE message.
- Web: `npm ci && npm run build` in worktree (background, guide changes only) — quote rc in the DONE message.

## What I did NOT do

- No writes under `/home/hoang/nofx` (main tree untouched). DB opened read-only only.
- No change to `freshMult`/`zoneFreshMult` tables, no scoring-constant moves, no seating changes (S3's scope), no S1 structure-block fields.
- `GUIDE_BUILT_REV` NOT bumped (deploy lane's step).
- No deploys, no RELEASE edits, no `.env`, no trader/account/key touches (SIM-only law intact).
- Did not edit any other lane's files.
