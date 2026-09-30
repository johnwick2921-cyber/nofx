# AUDIT — VL rename plan, code half (R1a + R1b + R3 + D0/D1/D2)

- **Auditor:** DS-101 (Copilot lane), 2026-09-30 ~01:10–02:35Z
- **Plan under audit:** `/home/hoang/rename-vl-plan/2026-09-29-rename-nofx-to-vl-plan.md` — md5 `b16759336fe0e264d11c1bc885ce5192` (matches the dispatch's expected `b1675933`), 211 lines, read in full.
- **Base:** `origin/dev` @ `9d52f5dc6` — `RELEASE 4d5382069… — boot 7: #266 + #265 booted 01:00:43 CT 2026-09-28` (`git log -1` quoted verbatim above; fetched 2026-09-30 ~01:10Z).
- **Also checked:** `origin/fix/updater-flat-live-ok` @ `21a8b6cd3` (PR #267, for the D0 claim).
- **Disposition:** DONE_WITH_CONCERNS — no P0; the plan is buildable as written, but it has 2 P1 gaps (one live cutover script, one binary-staging line), 2 P2 gaps, and several P3 wording/imprecision fixes. Nothing found blocks D0.

Evidence tiers on every claim: [A] read the exact line this session · [B] inferred from strong evidence · [C] speculation.

---

## CHECK 1 — cited file:lines exist and say what the plan claims

Every R1a/R1b/R3/D0 citation was verified at the base. Corrections are marked.

| Plan cites | Verdict [A] |
|---|---|
| `internal/updaterworker/runner.go:394` | ✓ `filepath.Join(w.cfg.Target.LogDir, "nofx_"+…+".log")` |
| `internal/updaterworker/recovery.go:133` | ✓ shell line greps `BOOT INTEGRITY OK — rev %s · %s/nofx_$(date +%F).log` |
| `internal/updaterworker/recovery.go:153` | ✓ `restartLine` const embeds `systemctl show -p MainPID --value nofx` (a string-built shell command in Go — the R1a dual-read must edit this string, not a glob) |
| `internal/activation/system.go:43` | ✓ `systemctl show -p MainPID --value nofx` |
| `internal/activation/system.go:121` | ✓ `filepath.Glob(filepath.Join(dir, "nofx_*.log"))` |
| `internal/updaterworker/release.go:109` | ✓ `binaryName = "nofx-bin"` |
| `internal/updaterworker/target.go:98` | ✓ `Binary: filepath.Join(t.InstallDir, "nofx-bin")` |
| `internal/updaterworker/runner.go:317` | ✓ `bin := filepath.Join(w.cfg.Target.InstallDir, "nofx-bin")` |
| `internal/updaterworker/snapshot.go:24` | ✓ `Binary: filepath.Join(dir, "nofx-bin")` |
| `internal/activation/activation.go:95` | ✓ `Binary: filepath.Join(dir, "nofx-bin")` |
| `internal/activation/steps.go:491` | ✓ `Binary: filepath.Join(dest, "nofx-bin")` (cited in D1) |
| `internal/updaterworker/host_os.go:34,45` | ✓ `ErrBotCgroup` mentions nofx.service; `bytes.Contains(b, []byte("/nofx.service"))` |
| `internal/updaterworker/host_os.go:56,78-86` | ✓ lock-script path/abs/stat/`check` verb |
| `cmd/nofx-updater/main.go:167` | ✓ `LockScript: filepath.Join(t.InstallDir, "deploy", "nofx-lock.sh")` |
| `cmd/nofx-activate/main.go:37` | ✓ `…empty = the NEWEST data/nofx_*.log…` |
| `agent/tools.go:1249` | ✓ `filepath.Glob(filepath.Join("data", "nofx_*.log"))` |
| `deploy/cutover.sh:176,192` | ✓ both lines reference `data/nofx_*.log` |
| `logger/log_prune.go` | ✓ prunes by `strings.HasPrefix(name, "nofx_")` + `.log` + parsed date (`log_prune.go:45-48`) |
| `internal/updaterworker/release_test.go:997` | ✓ `"nofx/internal/updateauth": …` — a DATA map key, not an import; goimports will NOT rewrite it (see P3-6) |
| `branding/scope_test.go:48` | ✓ `if !strings.HasPrefix(target, "nofx/") || current[target]` — would pass vacuously after the module rename (the plan's concern is real) |
| `store/hold_census_test.go ~175` | ✗ **WRONG PATH** — the file is `internal/updaterworker/hold_census_test.go:175` (`pkg := module + "/internal/updaterworker"` with `module` from `censuswalk.ModulePath`). Pattern exists; folder is wrong. Fix: cite `internal/updaterworker/hold_census_test.go` (P3-1) |
| `auth/auth.go:230` | ✓ `Issuer: "nofxAI"` — and there is NO `WithIssuer`/iss-validation anywhere [A] (grep over auth/ + api/), so the plan's "never checked at parse — no logout" holds |
| `api/handler_updates.go:65,232,310,408` | ✓ `UpdateHeader = "X-NOFX-Update"`; metrics `nofx_updates_refused_total`; the CSRF check at :408 is `r.Header.Values(UpdateHeader); len(v) != 1 || v[0] != "1"` — the "exactly one" rule is real and the dual-accept design preserves it (both → `len==2` → 403) |
| `api/server.go:98` | ✓ CORS allow-list is `"Content-Type, Authorization"` — the update header is deliberately absent, pinned by `TestUpdatePreflightNeverAllowsTheUpdateHeader` |
| `telemetry/metrics.go`, `telemetry/nt8_exit.go` | ✓ `nofx_decisions_total`, `nofx_decision_latency_seconds`, `nofx_databento_errors_total`, `nofx_risk_gate_trips_total`, `nofx_nt8_exit_busy_parks_total`, … |
| `provider/ninjatrader/tcp_server.go:2027` | ✓ `HelloPayload{ProtocolVersion: …, Source: "nofx-go"}` |
| `ninjascript/VLTraderTCPClient.cs:748-765` | ✓ the hello handler checks the protocol version only; it does NOT read `source`. "No AddOn change, no protocol bump" for the hello source is TRUE |
| `ninjascript/VLTraderTCPClient.cs:429,440` | ✓ only NofxTrader mention: `%USERPROFILE%\NofxTrader\account.txt`; **read-only** — nothing writes account.txt [A] (grep `account.txt` over all `.cs` and `.go`), so R3's "copy it across once" is sufficient and there is no writer to dual |
| `VL_BUILD_ID` | ✓ `private const string VL_BUILD_ID = "2026-09-23-m21"` (boot-line build_id at :1596,1653) |
| `agent/planner_runtime.go:2277,2860,3751` / `central_brain.go:67,482` / `llm_skill_router.go:176` / `workflow.go:509` | ✓ all 7 cited lines contain "NOFXi" (3+2+1+1) |
| `.github/workflows/release.yml:134` | ✓ `go build -trimpath -o nofx-bin ./` |
| `deploy/release/package.sh:19` | ✓ `"nofx-bin"` |
| `logger/logger.go` | ✓ writer: `fmt.Sprintf("nofx_%s.log", …)` at `logger/logger.go:90` |
| Web storage | ✓ only TWO `nofx_*` keys exist in current code: `nofx_beginner_wallet_address`, `nofx_beginner_onboarding_completed` (`SetupPage.tsx:67-68`, `AuthContext.tsx:255-256`). `nofx_user_mode` appears NOWHERE in `web/src` [A]. Chat keys ✓: `web/src/lib/agentChatStorage.ts:1,15,19` (`nofxi-agent-chat`, template `nofxi-agent-chat:${userId}`, `…-draft:${userId}`) |
| D0 — `trader/installation_gate.go` `censusVerdict` | ✓ at `origin/fix/updater-flat-live-ok` (branch head `21a8b6cd3`) line 357 appends `" — %d connected non-SIM connection(s), all accounts flat — allowed (owner ruling 2026-09-28)"` to `detail` BEFORE the refusal `why` list is joined — a refusal for an open position also reads "all accounts flat — allowed". The D0 defect is REAL, and D0's fix ("append only when positions==0 && working==0") is the right one. |

---

## CHECK 2 — MISSED READERS

### P1-1 — `deploy/leveltruth-cutover.sh` reads the unit `nofx` and is NOT in the plan anywhere
- **Where [A]:** `deploy/leveltruth-cutover.sh:37` `OLDPID=$(systemctl show -p MainPID --value nofx …)`, `:49` `NEWPID=$(… --value nofx …)`.
- **Why it matters:** `docs/superpowers/SYSTEM-MAP.md:448` documents `leveltruth-cutover.sh` as THE live cutover path (FLAT-GATE → RELEASE marker → binary swap + `kill -9` → poll boot line). R1a item 3 names `postboot-check.sh`, `planner-ab-report.sh`, `cutover.sh` — but not this one. After R2 the unit is `vl`; this script then reads a nonexistent unit → empty/0 pid → `kill -9` targets nothing and the boot-line poll watches a process that either died or was never killed → a live cutover silently breaks, or half-swaps (release marker written, old binary still running). That is the exact "silent safety-stop" class the audit is asked to catch.
- **Fix to the plan:** add to R1a item 3: "`deploy/leveltruth-cutover.sh` (~37, ~49): try `systemctl show -p MainPID --value vl`, else `nofx`" and to R1b item 7: "`leveltruth-cutover.sh` default `UNIT=vl` (same pattern as `cutover.sh`), or a one-line wrapper".

### P1-2 — `deploy/cutover.sh` stages the binary as `nofx-bin` and R1b never tells anyone to change it
- **Where [A]:** `deploy/cutover.sh:69` `cp "$NEW_BIN" "$STAGE_DIR/nofx-bin"`, `:70` `md5sum "$STAGE_DIR/nofx-bin"`.
- **Why it matters:** R1b item 2 renames the shipped binary to `vl-bin` (release.yml, package.sh, Makefile, Dockerfile) but does not name cutover.sh's staging lines (cutover.sh is cited only for log reads in R1a item 3 and UNIT=vl in R1b item 7). Post-R2 a cutover would stage `nofx-bin` into an archive whose layout the updater expects to be `vl-bin` (activation is dual in R1a, so it would still resolve during the compat window — but the moment the wrappers retire, every staged archive is mis-named and a good install gets refused).
- **Fix to the plan:** R1b item 2, add: "`deploy/cutover.sh` (~69-70): stage as `vl-bin` (keep `nofx-bin` acceptance in the dual-reader until the wrapper window closes)".

### P2-1 — the bot's own runtime backup-dir writers (`~/nofx-backups`) are unlisted
- **Where [A]:** `store/adherence_regrade.go:125`, `store/ab_confirm.go:448`, `store/wave_a_migration.go:53`, `store/bar_contract_key.go:225` (the last one `os.MkdirAll`s the dir at runtime), `cmd/dayplan-level-repair/main.go:9`, `cmd/nofx-activate/main.go:83`, `cmd/nofx-updater/main.go:158,164`.
- **Why it matters:** the plan only covers the PRUNER (`nofx-db-backup.sh`) and the R2 folder move with symlinks. The symlink keeps these working, so this is not a break TODAY — but nothing in R1b says whether these writers switch to `~/vl-backups` or keep writing through the symlink forever. If the `~/nofx-backups → ~/vl-backups` symlink is ever removed (rollback path? future cleanup?), `bar_contract_key.go:225` silently recreates a REAL `~/nofx-backups` dir and backups split across two homes. Unlisted writer = silent divergence.
- **Fix to the plan:** R1b add a line: "backup-dir WRITERS (adherence-regrade, e8-backfill, wave-a-record, bar_contract_key pre-migration, dayplan-level-repair, activation/updater backup roots) either move to a `VL_BACKUP_ROOT` default `~/vl-backups` or are explicitly documented as staying on the symlink for one release; the pruner already dual-reads".

### P2-2 — the lock dir env knobs are unnamed in the env list
- **Where [A]:** `deploy/nofx-lock.sh:62` `LOCK_DIR="${NOFX_LOCK_DIR:-$HOME/nofx-main.lock.d}"`, `:63` `LEGACY_LOCK="${NOFX_LEGACY_LOCK:-$HOME/nofx-main.lock}"`.
- **Why it matters:** D1's env list has no `NOFX_LOCK_DIR`/`NOFX_LEGACY_LOCK`. R1b item 6 changes the DEFAULT, but a user/session that exports `NOFX_LOCK_DIR` (the lock tests do: `nofx-lock-test.sh:29,301,488`) keeps pinning the old dir. The new-script-refuses-while-old-dir-exists rule is correct for the default home; the env override is a second, unpinned path into the old lock home.
- **Fix to the plan:** D1 env list add `NOFX_LOCK_DIR`, `NOFX_LEGACY_LOCK`; R1b item 6: "`vl-lock.sh` reads `VL_LOCK_DIR` else `NOFX_LOCK_DIR` else `$HOME/vl-main.lock.d`".

### P3 (cosmetic-to-low-risk missed names; each needs a line in D1/D2 or the census allow-list)

- **P3-2** `deploy/install-clock-guard.sh:21` enables `nofx-clock-guard.timer` — not in R1b item 7's rename list. Add it.
- **P3-3** `deploy/install-updater-worker.sh:105-111` enables unit `nofx-updater` and references `~/.config/systemd/user/nofx-updater.service` — R1b item 7 names the script's repo URL + `$HOME/vl` + `~/.config/vl-updater/env` but not the unit name INSIDE it. Add "enable `vl-updater`" explicitly.
- **P3-4** Env names missing from D1's "at minimum" list: Go `os.Getenv` reads at the base are `NOFX_DEMO_DB`(×2), `NOFX_TRADER_ID`, `NOFX_RELEASE_DIR`, `NOFX_REHEARSAL*` (5), `NOFX_LIVE_TESTS`, `NOFX_EXPECTED_REVISION`, `NOFX_DEMO_VERIFY/TRADER/SEED`, `NOFX_DB_PATH`, `NOFX_CLOCK_STATE`, `NOFX_CHART_ACROSS_ROLL`, `NOFX_CALENDAR_STATIC`, `NOFX_BAR_SCALE_MISMATCH_PCT/_MULT`. Shell reads: `NOFX_UPDATER_BUILD_REPO`/`NOFX_UPDATER_INSTALL_DIR` (`install-updater-worker.sh:33-34`), `NOFX_CUTOVER_TOKEN` (`cutover.sh:127`), `NOFX_RELEASE_DIR` (`cutover.sh:94`), `NOFX_UNIT` (`cutover.sh:36`), `NOFX_BIN`/`NOFX_REPO` (`postboot-check.sh:16,30`), `NOFX_BACKEND_PORT`/`NOFX_FRONTEND_PORT` (`start.sh:178-188`, `install.sh:105-106`, `docker-compose*.yml`, `pr-docker-compose-healthcheck.yml:38-40`). D1's list names NOFX_UPDATER / NOFX_CUTOVER_TOKEN / NOFX_RELEASE_INBOX as if they were `os.Getenv` reads — at the base those three are SHELL/env-file reads, not Go reads [A]. Fix: split the D1 env list into "Go `os.Getenv` names" and "shell names", and add the ones above. Docker-compose `${VAR:-default}` cannot express a fallback chain — use `VL_BACKEND_PORT:-${NOFX_BACKEND_PORT:-8080}`.
- **P3-5** `agent/skills/strategy_management.json` contains `nofxos_api_key` — the census allow-list must cover `agent/skills/*.json` NofxOS references, not just the Go package `provider/nofxos`.
- **P3-6** D1 item 4 lists `reverifier_test.go` and `sshsig_test.go` as files whose "forbidden-import lists" need ModulePath derivation — their `"nofx/…"` literals are IMPORT statements (`reverifier_test.go:15-16`, `sshsig_test.go:18`), which R1b item 1's goimports rewrite handles automatically [A]. Only DATA literals need derivation: `release_test.go:997` (map keys) and `branding/scope_test.go:48` (prefix compare). Reword so DS-103 doesn't hand-edit import blocks.
- **P3-7** `web/src/brand-i18n.test.ts:26` strips `/NOFX_[A-Z_]+/g` from prose before asserting — after R1b the FAQ env text becomes `VL_*`, so the strip line goes dead (harmless). Leave it; just don't let the census guard count it as a violation.
- **P3-8** CLI help strings `nofx-updater resume/fetch` (`steps.go:552`, `nt8.go`, `reverifier.go:79-83`, `main.go:158`) — mechanically renamed with the cmd dirs in R1b item 1; the census guard's "R1a fallbacks/wrappers" bucket should explicitly include CLI prose during the transition.

---

## CHECK 3 — does R1a change runtime behaviour TODAY?

- The design holds except for one self-contradiction in wording. D1 item 5 claims "Byte-identical runtime behaviour today", but the env helper spec in item 2 says "+ one WARN" when only the NOFX_ form is set. On this box the owner's `.env` sets NOFX_* names, so the new WARN **will** appear in logs from day one of R1a. Behaviour (trading, updater decisions) is unchanged; the LOG BYTES are not. This matters because the updater's boot proof anchors on log shape (`bootcheck.go:28-30` comment: matches are ANCHORED to the boot line's logger shape) — a new WARN line is safe for that anchor, but the plan should say so.
- **Fix:** D1 item 5 → "Byte-identical runtime behaviour today except the single WARN line per fallback env name (safe: the boot-line anchor is line-start-anchored on the boot line itself)".

---

## CHECK 4 — R1b risks

- **Vacuous-pass tests:** `branding/scope_test.go:48` (prefix compare) and `release_test.go:997` (map keys) are the real vacuous-pass risks and are addressed; `censuswalk_test.go:107-307` builds SYNTHETIC trees with `"nofx/…"` strings — those remain meaningful after the rename (they test censuswalk mechanics, not the real module) — no change needed [B].
- **Generated/embeds/reflection/string-built:** no `//go:embed` depends on the name (product.txt/persona.txt are already VL; `agent/skills/*.json` embeds carry only `nofxos_api_key`). The one true string-built command is `restartLine` in `recovery.go:153` — already in the R1a cite list. No reflection-built import paths found [A].
- **Header dual-accept:** `api/handler_updates.go:408` enforces exactly-one-value `"1"`. Accepting EITHER name while keeping `len(v)==1` preserves the CSRF strictness exactly as the plan says; keep `api/server.go`'s CORS allow-list untouched. The updater binary does NOT send the update header at all [A] (`grep UpdateHeader cmd/ internal/` → nothing), so the server-side dual-accept cannot break the old updater.
- **Storage migration:** only 2 of the 3 named `nofx_*` keys exist in current code; `nofx_user_mode` is absent [A]. KEEP the migration for it anyway (old browsers may hold it) but word it "if present" — migrating a key that is never present must not log an error or fail the vitest.
- **Census allow-list width:** the plan's buckets are good but too coarse — "the R1a fallbacks/wrappers" as a blanket could excuse whole files. Require per-line pattern matching with the allow-list enumerated as `file:pattern` pairs (see P3-5, P3-7, P3-8 for the additions). Otherwise a lane can bury one new `nofx` in a "wrapper" file and the guard passes.
- **AddOn:** verified — the ONLY nofx-named thing in `ninjascript/*.cs` is the `NofxTrader\account.txt` read (429/440), which is read-only, and `VL_BUILD_ID` is bumpable. The plan's R3 is complete. `grep -rni nofx ninjascript/*.cs` outside NofxTrader returns nothing [A].

---

## CHECK 5 — D0/D1/D2 wording precision

- **D0** is precise and its defect claim is verified on `21a8b6cd3` (see CHECK 1). One nit: "on the pass path" is the same as `positions==0 && working==0` only because the pass path also requires `nonSim==0 && unsettled==0` — state it as the conjunction D0 already gives; no change needed.
- **D1** ambiguities: (a) "one env helper (e.g. `config.Env(name)`)" — the `e.g.` is fine, but say the helper MUST live in one place and every production `os.Getenv("NOFX_…")` routes through it, with the PR listing the before/after grep counts; (b) the env minimum list mixes Go and shell names (see P3-4) — split it; (c) item 4's file list needs the reword in P3-6 and the path fix in P3-1.
- **D2** item 5: add "if present" to the `nofx_user_mode` migration (P3-4-adjacent). Item 3 is precise and matches the code. Item 2 must add the two lines from P1-2. Item 7 must add `leveltruth-cutover.sh` (P1-1) and `install-clock-guard.sh` (P3-2).
- **Plan-level wording:** "brand-scope-baseline.json re-pinned from the renamed bytes" appears in R1b item 7; D1's brand-scope line says "files you edit that are hash-pinned … re-pinned in the SAME PR" — consistent, no fix.

---

## Findings ranked (for the CTO's fold)

| # | Sev | Finding | Fix (one line for the plan) |
|---|---|---|---|
| 1 | P1 | `deploy/leveltruth-cutover.sh:37,49` reads unit `nofx`; the live cutover path is not in the plan | R1a item 3 + R1b item 7: dual unit (`vl` else `nofx`), default `UNIT=vl` |
| 2 | P1 | `deploy/cutover.sh:69-70` stages `nofx-bin`; R1b never renames the staging | R1b item 2: stage as `vl-bin` |
| 3 | P2 | Bot's own `~/nofx-backups` WRITERS unlisted (7 sites) | R1b: list them; move or document symlink-forever |
| 4 | P2 | `NOFX_LOCK_DIR`/`NOFX_LEGACY_LOCK` env override not in the env list | D1 env list + R1b item 6 fallback chain |
| 5 | P2 | Census allow-list buckets too coarse ("fallbacks/wrappers") | Require `file:pattern` pairs, no directory-wide buckets |
| 6 | P3 | `store/hold_census_test.go` wrong path → `internal/updaterworker/hold_census_test.go` | Fix the citation |
| 7 | P3 | D1's env minimum list mixes Go vs shell names; misses `NOFX_UPDATER_BUILD_REPO`, `NOFX_UPDATER_INSTALL_DIR`, `NOFX_UNIT`, `NOFX_BIN`, `NOFX_REPO`, `NOFX_BACKEND_PORT`, `NOFX_FRONTEND_PORT`, demo/rehearsal names, `NOFX_LOCK_DIR` | Split the list, add them |
| 8 | P3 | `nofx_user_mode` storage key absent from code | "if present" wording |
| 9 | P3 | "Byte-identical runtime behaviour today" vs the WARN | Add "except one WARN line per fallback env name" |
| 10 | P3 | `reverifier_test.go`/`sshsig_test.go` don't need ModulePath derivation (imports only) | Restrict item 4 to data literals |
| 11 | P3 | `install-clock-guard.sh`, `install-updater-worker.sh` unit names, `agent/skills/strategy_management.json` `nofxos_api_key` | Name them in R1b item 7 / census allow-list |

## What I did NOT do (L-AUDIT)
No code change, no worktree edit outside this report, no deploy, no stop/start, no DB write, no lock acquire. Only `git fetch` (read-only) and greps/sed reads in the audit worktree. No race run needed (no code built); `go build`/`go test` were not run because nothing was changed.

## Evidence
- Base `git log -1`: `9d52f5dc6 RELEASE 4d5382069643a491403ac0a8575f2bbea3648676 — boot 7: #266 … booted 01:00:43 CT 2026-09-28 …`.
- #267 branch head: `21a8b6cd3 UPDATER-FLAT-LIVE-OK: a FLAT non-SIM connection no longer blocks an update` — `installation_gate.go:357` (branch blob) confirmed the D0 defect.
- All grep/sed commands were run at the base in `/home/hoang/nofx-ds-101-audit` (worktree, HEAD `cc2de4669` = claim commit on `audit/rename-code`).
