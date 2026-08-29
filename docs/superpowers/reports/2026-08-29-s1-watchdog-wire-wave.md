# S1 Watchdog Wire-Up Wave (BUILD & PROVE, NO DEPLOY)

- Branch: `fix/watchdog-wire` (off dev `43ceaf9c`) · Date: 2026-08-29
- Status: **proven & parked — awaiting owner "go cutover"** (deploy this afternoon, hours before 17:00 CT)
- Scope: S1 (the pre-live-fire BROKEN) + the 4 A-items, all folded same wave. Canon: AUDIT-CHECKLIST self-law satisfied (class 19 appended in the same PR).

## S1 — persist-silence watchdog WIRED (was declared-but-dead code)

- `provider/ninjatrader/bar_persist.go`:
  - **Flush stamp:** `persistLastFlushAt.Store(time.Now().Unix())` on EVERY successful flush — the single `flush()` closure serves both the 256-batch and the 300ms-tick paths, so one stamp site covers both (cited `startPersistWorker`).
  - **30s ticker:** the worker loop gained a `watchdog` ticker → `persistWatchdogAlarmAt(now)`; when `now − lastFlush > persistWatchdogSeconds()` (default 60s, min 10) it emits the deduped loud ERROR via `persistAlarmAt` (one alarm per watchdog window). Cold start (zero stamp — boot backfill in flight) never alarms.
- **Behavior fixture (the regression pin, `TestPersistWatchdogAlarmFiresOnceAndRecovers`):** simulated 61s stall → the alarm **FIRES exactly once** with the exact ERROR text `🔕 PERSIST WATCHDOG: no successful bar flush for 61s (queue_drops=%d) — the persist writer may be stalled (the 2026-08-28 GORM-stall class)` · same window re-call is silent (dedup) · resumed flush advances the stamp and **no repeat alarm** · a fresh stall after recovery re-fires (guard is alive, not a one-shot). Existence tests escaped the half-ship before; this one proves firing.

## A3 — HTF-seat injection rows stamped (the 13 Demand·1h escapes)

Root cause: the write-site stamp map knew the cap-4 HTF-zones prompt section, but the model also reads the whole key-levels block and wrote zones the cap hid (e.g. `Demand·1h (HTF) A 29722.62`).
- `kernel/planner_prompt.go`: `PlannerInput.HTFZonesFull` — the FULL in-band graded HTF-zone universe (uncapped). Prompt rendering still uses the cap-4 `HTFZones`.
- `trader/auto_trader_planner.go`: builds `htfZonesFull` (`ScoreLevels(..., len(zones), ...)`) beside the cap-4 build; the write-site record loops are extracted into pure `collectMachineGrades` (Levels + Pool + HTFZones + HTFZonesFull) and the write site calls it.
- Fixture `TestCollectMachineGradesCoversHTFUniverse`: the 1h zone outside the cap-4 section is recorded (grade + label), 4h included, legacy sources + the stronger-grade collision rule regressions pinned.

## A4 — boot line prints BOTH caps

`main.go` AI-params line is now `🧠 AI params in force: model=… client_max_tokens=%d planner_max_tokens=%d …` via exported `trader.PlannerMaxTokens()` (default 65536). The misleading single 32768 is gone — next boot shows `client_max_tokens=32768 planner_max_tokens=65536`.

## A1 — n-values on the two offender lines (class 16)

- `2026-08-27-guide-page.md`: "(94.2% ON — n not recorded in that wave, pre-retention week: treat as anecdote per class-16 law; 75% reject-NY — fresh reconstruction at the same cut: n=7, 4W, +$386.5)".
- `2026-08-27-level-truth-wave.md`: missed-turns delta line now carries the n reference (08-28 recompute at 0.3×ATR: 60.9/59.1/72.7%).

## A5 — main-tree churn reverted + gap noted

`git checkout -- store/position_query.go` in the main tree (the 5-line comment-only churn, byte-verified: 5+/5−). Main tree is now porcelain-clean for tracked files (0 dirty). Gap noted: it sat dirty with no `~/nofx-main.lock` while the lock law was already recorded — the law's marker step was skipped (no collision occurred; process note only).

## AUDIT-CHECKLIST — class 19 appended (same PR, per its own law)

**"Half-shipped guard": declared state without wired call sites; probe = grep every guard atomic/knob for `.Store`/`.Load` call sites + a behavior fixture that makes the alarm FIRE.** Header count 18→19.

## Gates

| Gate | Result |
|------|--------|
| `go build ./...` | ✅ |
| `go vet ./...` | ✅ clean |
| `go test ./...` | ✅ 27 packages ok, 0 fail |
| Goldens (`-run Golden`) | ✅ |
| New fixtures | ✅ `TestPersistWatchdogAlarmFiresOnceAndRecovers` · `TestCollectMachineGradesCoversHTFUniverse` |

No guide-content changes (docs/report/checklist only) → no `GUIDE_BUILT_REV` bump required.

## Cutover checklist (owner "go", this afternoon)

1. Merge `fix/watchdog-wire` → dev; temp-clone build at merge sha (`vcs.revision == merge sha`).
2. Flat-gate all-origin (DB OPEN=0 · NT8 both accounts count=0 · API `[]` · armed non-terminal=0).
3. Marker commit: `deploy/RELEASE=<merge sha>`.
4. Swap → `kill -9` → boot poll: `🔐 BOOT INTEGRITY OK` · goldens PASS · **`🧠 AI params in force: … client_max_tokens=32768 planner_max_tokens=65536 …`** · zero ERRO. (The watchdog has no boot line by design — its live proof is silence while healthy and the quoted ERROR on a real stall.)
