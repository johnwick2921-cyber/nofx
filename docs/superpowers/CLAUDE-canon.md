<!-- CANON. THIS FILE IS TRACKED IN GIT AND IS THE SOURCE OF TRUTH.
     The repo-root CLAUDE.md is gitignored and is only a pointer at this file.
     Owner ruling 2026-09-03: a law that cannot travel on a branch or be guarded
     is not canon — it is a local note that happens to be obeyed. Edit THIS file,
     on a branch, with a review, like code. The tree guard checks its md5 (check 5).
-->

# nofx — AI Trading Bot

## Project state (2026-05-29)

**CURRENT ARCHITECTURE: NT8 is the SINGLE data source AND execution path (Stage 4 live).**

Live flow: NT8 (Tradovate real-time MNQ bars via the VL C# AddOn over TCP) → C# AddOn → TCP → Go bot → BarCache (key `"MNQ"`) → Go kernel (`market/data.go` futures branch reads BarCache, NOT CoinAnk/Binance) → futures system prompt → AI (`deepseek-v4-pro`) → risk gate (futures contract sizing) → signal → NT8 SIM execution (same AddOn).

- **Data source:** NinjaTrader 8 Tradovate real-time bars (via C# AddOn over TCP, cached in Go BarCache).
- **Decision engine:** Go kernel reads NT8 BarCache → futures prompt → AI → risk gate.
- **Execution:** NT8 SIM via the same TCP AddOn (CSV path deprecated).
- **Symbol:** MNQ end-to-end. The `"MNQ 06-26"` expiry form lives ONLY inside the C# `GetInstrument` call; everywhere else it's `MNQ`.
- **Transport:** Bars / account / positions / orders ride TCP (NOT CSV).

**Critical:** If NT8 closed/disconnected → no bars → stale cache → NO real decisions. NT8 is NOT optional; it is both data AND execution.

**Why Databento dropped (locked decision):** Databento Historical tier has ~8h lag (unusable for live trading). Proven 2026-05-28: the real MNQ bar in Stage 3 decision (#279, price 30323.75) came from NT8 Tradovate, not Databento. Do NOT re-wire Databento as a live source.

Plan: [docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md](docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md) (architecture pivot section).

## HARD RULE — NT8 compile/deploy location

NT8 compiles NinjaScript ONLY from the fixed Windows path:

```
C:\Users\hoang\Documents\NinjaTrader 8\bin\Custom\AddOns\
(WSL view: /mnt/c/Users/hoang/Documents/NinjaTrader 8/bin/Custom/AddOns/)
```

**NOT** from the repo (`/home/hoang/nofx/ninjascript/`). Editing a repo `.cs` file alone does NOTHING — NT8 never sees it.

**Deploy procedure for every C# change:**
1. `cp` repo `ninjascript/*.cs` → Documents AddOns folder
2. F5 compile inside NT8
3. **Full NT8 restart** (AddOns do NOT hot-reload)

Skipping any step = NT8 runs the old binary. This is the single biggest NT8 gotcha.

## Repo ownership

- `origin` = **github.com/johnwick2921-cyber/nofx** (this project — user's own)
- `upstream` = github.com/NoFxAiOS/nofx (historical only — NOT a contribution target)
- Do NOT propose contributions back to NoFxAiOS unless the user explicitly asks.

## Historical reference (upstream — useful background only)

- The codebase was originally forked from `NoFxAiOS/nofx`. Their architecture docs at `https://github.com/NoFxAiOS/nofx/tree/dev/docs/architecture` (especially `STRATEGY_MODULE.md`) are useful background for the strategy-engine design.
- **Stale-path warning:** upstream docs cite `decision/engine.go` — that folder doesn't exist locally. Code is in `kernel/engine.go`. Same for engine_analysis.go, engine_position.go, engine_prompt.go.

## Critical security gate

`config/config.go:67-69` defaults `JWTSecret = "default-jwt-secret-change-in-production"` if env unset. The log `🔑 JWT secret configured` at `main.go:90` fires UNCONDITIONALLY and does NOT indicate the default was overridden. Verify via `grep JWT_SECRET .env`. Set with `openssl rand -base64 64` before any non-localhost deploy.

## Build & test

- `go build ./...` — main binary
- `go test ./...` — all tests
- `cd web && npm run dev` — frontend dev server (port 3000, proxies /api to :8080)
- `cd web && npm run build` — production frontend
- `./nofx-bin` — runs locally; SQLite at `data/data.db`. Docker NOT required.
- `go run ./cmd/nq_smoke <sub>` — futures smoke matrix: `prompt` + `roundtrip` are offline-safe; `databento`/`resolver` need `DATABENTO_API_KEY`; the no-arg default exercises the DEPRECATED CSV path. See `cmd/nq_smoke/CLAUDE.md`.

## Trading mode env vars

- `TRADING_MODE=crypto` (default) or `futures` — toggles NQ path
- `DATABENTO_API_KEY` — NOT required for futures trading. Only used by the optional `nq_smoke databento|resolver` sub-smokes (Databento dropped as live source 2026-05-28)
- `DATABENTO_DATASET=GLBX.MDP3` — CME Globex (historical/backfill context only)
- `NINJATRADER_DATA_DIR=/mnt/c/Users/<u>/NofxTrader/data` — legacy CSV bridge only; the live TCP path does not use it

## Supported brokers (exchange field)

- crypto CEX: `binance`, `bybit`, `okx`, `bitget`, `gate`, `kucoin`, `indodax`
- crypto DEX: `hyperliquid`, `aster`, `lighter`
- CME futures: `ninjatrader` — NinjaTrader 8 (real-time bars + SIM execution via TCP bridge; CSV legacy path deprecated)

## Architecture: 4 control surfaces

1. **Config** (Settings) — exchanges + AI models, per-trader config
2. **Dashboard** (Trader) — positions, P&L, AI decisions
3. **Strategy** (Studio) — prompt + indicators + risk control
4. **AgentBeta** (Chat) — conversational AI (NOFXi)

Three pages removed: Data (broken iframe), Strategy Market (crypto community), Competition (public leaderboard).

## High-cascade types — do NOT break

- `market.Kline` struct shape — touched by 20+ files
- `trader/types.Trader` interface — **19 methods exactly**; broker impls use compile-time check `var _ types.Trader = (*Trader)(nil)`
- Decision JSON shape `{action, symbol, entry, stop_loss, take_profit, reasoning, leverage, confidence}` — engine + DB + web UI all read it
- TCP wire schema (`provider/ninjatrader/tcp_framing.go` ⟷ `ninjascript/*.cs`, spec in `ninjascript/vltrader_tcp_PROTOCOL.md`) — Go structs and C# AddOn must change in lockstep
- CSV protocol for `vltrader.cs` (legacy, renamed from `claudetrader.cs` 2026-05-23) — 5-field signals, 3-field fills

## Common gotchas

- `market.Normalize()` (data.go:~620) early-returns CME futures symbols via `IsCMEFuturesSymbol(raw)` BEFORE uppercasing — FIXED (was the data.go:557-558 bug). Invariant: keep that early-return first; `NQ.c.0` needs its lowercase suffix preserved.
- `nofxos.ai` is deprecated; key `cm_568c67eae410d912c54c` returns HTTP 402. (Databento dropped for live trading due to ~8h lag; NT8 Tradovate is the data source.)
- NT8 ATM strategies block `OnOrderUpdate/OnExecutionUpdate` events — pick managed-or-ATM, not both.
- WSL2 mirrored mode (Win11 22H2+) needed for `127.0.0.1` to reach Windows-side NT8. Plain NAT mode requires firewall rules + host IP discovery.
- `ExportCalculateMACD` returns only the MACD line (one float64). Signal/histogram not exposed by current API.

## Memory & plans

- Persistent memory: `~/.claude/projects/-home-hoang-nofx/memory/` (MEMORY.md indexes entries)
- Implementation plan: `docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md`
- Existing agent persona: `agents.md` (Chinese — NOFXi assistant spec)

## Code intelligence & maps (generated 2026-07-10 @ 7a8adce0)

- **Architecture map (start here):** `docs/superpowers/reports/2026-07-10-architecture-codebase-map.md` (+ raw per-subsystem maps in `2026-07-10-subsystem-maps.json`)
- **Knowledge graph:** `.understand-anything/knowledge-graph.json` (3,121 nodes / 9,588 edges / 11 layers / 15-step tour) — query with `/understand-chat`, view with `/understand-dashboard`
- **cgc code graph** (`mcp__cgc__*` tools): repo cleanly indexed 2026-07-10 (781 files / 5,095 functions; `.cgcignore` excludes `.claude/`, `.understand-anything/`, artifacts). Reindex after big changes: `delete_repository` + `add_code_to_graph` (~35 min background job).
- Complexity hotspots (cgc, deduped): `agent/skill_execution_handlers.go` holds the repo top-3 (`strategyConfigFieldDisplayName` cc=204, `executeTraderManagementAction` cc=178, `applyStrategyConfigPatch` cc=163).

## Subsystem CLAUDE.md files

- `provider/CLAUDE.md` — adding data providers
- `provider/ninjatrader/CLAUDE.md` — NT8 TCP bridge, Go side (bars/account/orders wire)
- `trader/CLAUDE.md` — adding brokers (19-method interface)
- `trader/ninjatrader/CLAUDE.md` — NT8 TCPTrader (live futures execution)
- `kernel/CLAUDE.md` — strategy engine + prompts
- `market/CLAUDE.md` — OHLCV + indicators + Normalize invariant
- `web/CLAUDE.md` — React frontend conventions
- `cmd/nq_smoke/CLAUDE.md` — end-to-end smoke matrix

## Agent toolbox & standing rules (hardening C4)

Operating rules for any agent working this repo. Additive reference — codifies the
recurring standing dispatch constraints so they survive across sessions.

**Evidence tiers** — label non-obvious claims: **[A]** directly verified (ran it,
read the exact line, saw the frame); **[B]** inferred from strong evidence; **[C]**
speculation. Audit/review findings ship at full depth, terse in chat.

**WORKTREE LAW (canon 2026-08-27).** The main checkout `~/nofx` belongs to exactly
ONE active dispatch at a time. Any secondary work — read-only audits,
investigations, queued waves — runs in an isolated `git worktree add
../nofx-<task>` directory, NEVER the main tree. Commit early, commit often: no
dispatch holds >30 min of uncommitted work.

**PUSH-EMPTY-AT-ACCEPT (canon 2026-09-03, class 70).** The moment you accept a
dispatch, claim the branch BEFORE you build anything:

```
git checkout -b <branch> origin/dev
git commit --allow-empty -m "claim: <wave> — <session>, $(date -Is)"
git push -u origin <branch>          # rejected, or already there? STOP.
```

If the branch already exists on origin, **another lane has this wave** — stop and
coordinate before writing a line. Fold the empty commit into your first real one
or leave it. Born the day two lanes independently wrote ~250 lines of the same
lock wave inside an hour and a third lane's branch was merged into dev without
its author ever being told; the only thing that surfaced any of it was a
non-fast-forward rejection at push time, after all the work was done. A branch
name on origin is the only claim this protocol has.

**PROVENANCE.** Every commit in this repo carries the identical author identity,
so git CANNOT answer "which lane wrote this". Provenance comes from the branch,
the worktree and the timestamp — never from the author field. Do not attribute a
commit to a lane without one of those three, and ask the lane before naming it.

**SPEC-FRESHNESS LAW (owner ruling 2026-09-03).** A worktree is cut from dev's
**TIP at accept**. A spec on dev is a MOVING artifact, and a worktree cut from an
older base silently freezes it — you read the superseded text as current and
nothing tells you. Before building against any spec or plan: `git log -1 --
<spec>`, compared against your worktree's base. **If the spec moved after your
base, rebase and re-read BEFORE writing code.** Every report quotes that
`git log -1` line for the specs it built from, so the base a wave was built
against is on the record instead of in someone's memory. (Born from the
tree-guard wave: the lock model was rewritten and the tree-guard spec edited in
the SAME commit at 21:48:36; the guard was built at 21:54 from a base six minutes
older, implemented the superseded lock, and would have false-alarmed on every
cutover. Found by a peer asking an unrelated question — not by any test, because
the wave's whole suite asserted the model its author had read. Checklist class
73, read beside 70 and 72.)

**MAIN-TREE LOCK LAW (canon 2026-08-28, from P2 audit S8; lock rebuilt
2026-09-03, class 70).** Before ANY dispatch touches `~/nofx`: the tree must be
porcelain-clean (`git status --porcelain` empty) and HEAD must be the single
allowed branch for that dispatch. Non-deploy work runs ONLY via `git worktree
add` + `git worktree lock`. The lock is acquired FIRST, with:

```
deploy/nofx-lock.sh acquire <session> "<task>" [minutes]   # atomic; REFUSES if held
deploy/nofx-lock.sh heartbeat <session>                    # beat every ~2 min as you work
deploy/nofx-lock.sh with-heartbeat <session> -- <cmd>      # wrap long steps (builds, suites)
deploy/nofx-lock.sh status | check                         # check: rc 0 free · 1 held · 2 stale
deploy/nofx-lock.sh release <session>                      # only the holder may release
```

`git reset` on `dev` is FORBIDDEN outside the deploy-owning dispatch. (Born from
the 2026-08-26 double force-reset of dev + six dirty worktrees.)

**LOCK LIVENESS (canon 2026-08-31 as the PID amendment; REPLACED 2026-09-03,
class 70).** There is NO pid field any more, and nothing checks `kill -0`. The
old flat file failed three ways in one day: a pid dead within a second of being
written while its owner worked on; an active cutover's lock silently clobbered
by a second `>` writer; and a resumed session naming its own former process. A
pid answers "does some process exist" — the question was always "is the OWNER
still working", and only the owner can answer it.

So: acquisition is an atomic `mkdir ~/nofx-main.lock.d` (a second acquire cannot
overwrite, only refuse, and it names the holder), identity is the SESSION NAME,
and liveness is a heartbeat the holder rewrites every 2 min. Older than 5 min
reads **STALE**.

**STALE IS NOT DEAD.** The tool never prints "dead" and never clears a lock
itself, at any age. Corroboration remains MANDATORY: ask the named session, watch
whether HEAD moves, look for a build in flight. An unexpired lock with a stale
heartbeat is the 0C case, and it now reads that way on its face.

**SUCCESSION IS ON THE RECORD.** Taking over an abandoned lock is a verb, not a
`rm`:

```
deploy/nofx-lock.sh reclaim <you> <stale-session> "<what you checked>"
```

Refused while the heartbeat is FRESH — a reclaim that can take a live lock is
replacement with better manners. You must name the holder you are taking over
(a misread of `status` must not become a silent seizure) and state the
corroboration; both empty and wrong are refused. It appends who / from whom /
heartbeat age / why to the lock's `history`, which `status` prints and `release`
prints again before the directory goes. It returns **rc 3**, not acquire's 0, so
a lane can tell "took a free lock" from "inherited an abandoned one" and refuse
to inherit.

**NO UNATTENDED DEPLOYS (canon 2026-08-27).** A cutover requires either (a) the
owner REACHABLE and acking the boot line within minutes, or (b) a TESTED
auto-rollback — boot line not observed within 90s OR goldens fail → restore the
prior binary + RELEASE, restart, alert. Timers are BANNED for deploys outright
(0-for-2 history). Deploys run only on the owner's explicit "go".

**SIM-only.** Every trade path is NinjaTrader SIM. Never enable, size for, or route
to a live NT account. `isAccountTradeable` blocks non-SIM; do not weaken it. The
owner's traders / account bindings / API keys are SACRED — never mutate them.

**GUIDE CONTENT LAW (canon 2026-08-28).** Any wave that changes a knob, play,
chip, gate, or default MUST update `web/src/guide/content/*` in the SAME PR —
including the `GUIDE_BUILT_REV` bump to the shipped rev. A guide that lies about
the running binary is worse than no guide; the drift banner is a failsafe, not a
maintenance strategy.

**AUDIT-PLAYBOOK LAW (canon 2026-08-28).** Every audit/verification dispatch
references `docs/superpowers/AUDIT-CHECKLIST.md` (the 18 bug classes + pre-audit
R1-R9 + pre-cutover protocol) instead of re-deriving the probe list. Every NEW
bug class found is appended to that file in the SAME PR that fixes it.

**Guarded DB writes.** `data/data.db` is READ-ONLY unless a task explicitly
authorizes a write. When authorized: back up first (the C1 timer or a manual
`~/nofx-backups/<name>/` copy), make the change idempotent, and WHERE-scope every
statement. Strategy config is cached at trader-load — a raw DB write is NOT
hot-active; it needs the trader reload (handleUpdateStrategy) or a restart.

**Deploy = rebuild + `kill -9 <PID>`.** `sudo systemctl restart nofx` needs a
password that isn't available here; SIGKILL lets systemd's `Restart=on-failure`
relaunch the new binary (SIGTERM exits 0 and does NOT relaunch). Only restart in a
flat/safe window. Editing a repo `.cs` alone does nothing — NT8 AddOns need the
copy → F5 → full NT8 restart dance (see the HARD RULE above).

**No-sudo ops surfaces (installed 2026-08-13):**
- **DB backups** — `systemctl --user` timer (linger on), 05:00+17:30 CT, online
  `sqlite3.backup()` → `~/nofx-backups/auto/{daily,weekly}`. Restore runbook:
  `deploy/RESTORE.md`. Reinstall: `bash deploy/install-db-backup.sh`.
- **Gate-block counters** — `GET /api/risk/gate-blocks` (per-trader/session-day
  table + daily journal summary line). Every risk gate calls
  `telemetry.IncGateBlock`.
- **Owner-gated (need sudo):** `deploy/install-journald.sh` (journald persist +
  2G cap), `deploy/install-autostart.sh` (boot units).

**CANON LAWS (H2, 2026-09-03).** Ruled in a wave, previously only in reports.

- **Canonical casing.** One canonicalizer per identifier, called where the value ENTERS. (checklist 28)
- **No fabricated values.** An empty computed list is `[]`; an uncomputed one is absent; they must differ. (49, 53)
- **Sample-id law.** Every claim about rows names the ids it rests on. (A21)
- **Corrected-column law.** `pnl_corrected` only; NULL is UNRESOLVED, excluded, and the COUNT is shown. (40)
- **RELEASE ordering, four halves.** RELEASE before the kill · marker after the boot · same tree · marker PUSHED before the lock releases. (39, 40)
- **Main tree is deploy-only (A2b).** No edit in `~/nofx` except `--ff-only` under the lock — a VS Code Save All silently reverted 596 lines of shipped safety code. (45)
- **Waves are named until merge.** Preconditions cite a branch/commit and a boot-line string; numbers are assigned AT MERGE. (A27 — three collisions had to be renumbered during the 09-03 combined boot)
- **Counters record, never infer.** (35)
- **Parity tests exercise production CALL SITES.** A test that builds both sides' inputs proves only self-consistency. (53)
- **Boot lines are READ, never literal**, and a field the process cannot know yet prints `n/a`. (45, 49)
- **A branch green alone is not green merged.** Run the full suite at the MERGED HEAD — the 09-03 boot found a bare time layout that only failed once two green lanes were on one HEAD.

**FORBIDDEN — the 5 Google Drive tools** (`search_files`, `read_file_content`,
`download_file_content`, `get_file_metadata`, `get_file_permissions`): never call
any of them, in the main loop or ANY subagent, under any framing.

**Partner repo (vlauto).** The sibling `vlautoagenttraderv1` mirrors the Go/AddOn
logic. Propagate shared changes via `format-patch → am` (build + goldens green
there, secret-scan) — the OWNER runs the push if the classifier blocks it. Do NOT
push to it directly. **After any `origin` history rewrite (e.g. the 2026-08-29
binary purge): the partner repo and every other clone MUST re-clone fresh —
`format-patch → am` against a rewritten remote is corrupt by construction.**

**Repo target.** `origin` = `johnwick2921-cyber/nofx` (ours). Never propose
contributions to `upstream` (NoFxAiOS/nofx) unless explicitly asked.
