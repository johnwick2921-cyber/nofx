# AUDIT-RENAME-MODULE — DS-102 (2026-09-29)

- **Slice:** MODULE REWRITE — plan "RENAME nofx → VL" R1b items 1, 7, 10 (plus R1a item 4, which is the module-name-guard half).
- **Plan audited:** `/home/hoang/rename-vl-plan/2026-09-29-rename-nofx-to-vl-plan.md` — `md5sum` = `b16759336fe0e264d11c1bc885ce5192` (matches the dispatch: b1675933). [A]
- **Base:** `origin/dev` `9d52f5dc611ed001fc057bca613685071d77efab` — `RELEASE 4d5382069643a491403ac0a8575f2bbea3648676 — boot 7 (#266 nt8_absent + #265 release build outside tree, booted 01:00:43 CT 2026-09-28)`. [A]
- **Method:** READ-ONLY. Check 1 was the only mutating step, done in a throwaway worktree `/home/hoang/nofx-audit-102` (removed after) exactly as the dispatch orders. Everything else is grep/read in `/home/hoang/nofx-102-renameaudit` at `9d52f5dc6`.

---

## CHECK 1 — the simulated module rewrite (R1b item 1)

Executed in the throwaway worktree:
```
grep -rl '"nofx/' --include='*.go' . | wc -l      → 1128 files
sed -i 's|"nofx/|"vl/|g' (all 1128)  +  go.mod module nofx → module vl
go build ./...  → rc=0, 0 error lines
go vet ./...    → rc=0, 0 diagnostics
```
**Verdict: the mechanical module rename compiles and vets CLEAN with zero breaks that are not plain import paths.** [A] No `//go:embed`, `//go:generate`, ldflags `-X`, build tag, or reflection site references the module path; `censuswalk.ModulePath` reads go.mod at runtime and self-corrects. The plan's claim "gofmt/goimports; go build ./... && go vet ./..." is achievable as written for item 1. The 2,053-imports/1,128-files arithmetic in the plan matches the tree (1,128 files measured). [A]

Residual item-1 work that is NOT a Go import path (the plan covers most): `cmd/nofx-updater`/`cmd/nofx-activate` dir renames, npm name, Makefile/start.sh/install.sh/Dockerfile — all named in the plan. One NOT named: `docker/Dockerfile.backend:52` `go build … -o nofx` (plan's item 1 says "Dockerfile output names" generically — name the file). [A][P3]

## CHECK 2 — module name built as a STRING (breaks or silent-empty)

| site | kind | post-rename behaviour | in plan? |
|---|---|---|---|
| `internal/updateauth/admin.go:62` `const passwordBindingDomain = "nofx-updater/password-binding/v1\x00"` | MAC domain string | Rename it → every enrolled password binding verifies against a NEW domain → owner re-enrollment or broken unlock. Leave it → census guard (item 10) hits it. | **NO — P1** |
| `internal/updaterworker/release.go:109` `binaryName = "nofx-bin"` | string const | R1a dual-accept covers | yes |
| `internal/updaterworker/{snapshot.go:24,target.go:98,runner.go:317}` `filepath.Join(…, "nofx-bin")` | path string | R1a dual | yes (lines named) |
| `internal/updaterworker/recovery.go:153` restartLine template `MainPID --value nofx` | shell template | R1a dual | yes |
| `internal/updaterworker/releasefixture/releasefixture.go:164,175` fixture archive embeds `"nofx-bin"` | test fixture | R1a's dual-reader tests use it; after R1b the fixture should ALSO produce a `vl-bin` case so the dual accept is tested from both sides | **no — P3** |
| `api/handler_updates.go:65` `X-NOFX-Update` | header const | R1b item 3 dual-accept | yes |
| `auth/auth.go:230` `Issuer: "nofxAI"` | JWT | R1b item 2 | yes |
| `provider/ninjatrader/tcp_server.go:2027` hello `Source` | wire | R1b item 2 (`vl-go`; AddOn ignores) | yes |
| `internal/censuswalk/censuswalk.go:233` `ModulePath` | runtime go.mod read | self-corrects | n/a |

## CHECK 3 — tests that would PASS ON EMPTY after the rename

Plan's list (release_test.go ~997, reverifier_test.go, sshsig_test.go, branding/scope_test.go ~48) is **incomplete and one citation is wrong**:

- `internal/updaterworker/release_test.go:997-998` — literal forbidden-import map (`"nofx/internal/updateauth"`, `"nofx/internal/activation"`). True pass-on-empty hazard. Deriving from `censuswalk.ModulePath` fixes it. ✓ plan right.
- `internal/updaterjob/verdict_test.go:29` — `allowedImports` map carries `"nofx/internal/updaterwire": true`. **MISSED by the plan.** Post-rename the parsed import is `vl/internal/updaterwire` → the test fails LOUDLY at the R1b gate (not silent-empty, but a first-pass break the plan should list so R1b is one pass). [A] P2.
- `internal/updaterworker/reverifier_test.go` + `sshsig_test.go` — **plan cites them as literal forbidden-list tests; [A] they are not.** Their `nofx/…` lines are plain import statements (sed handles them) and `sshsig_test.go:381` uses a self-consistent fixture path `/srv/nofx` (any path works; leave or rename, cosmetic). P3: fix the plan text (drop the false citation, add verdict_test.go).
- `branding/scope_test.go:48,62,77` — the flip ✓ plan right. But `:137` asserts `strings.Contains(err.Error(), "nofx/config")` — the flip must also update this literal, and the plan's "update text + logic" is silent about it. Name it. P3.
- `internal/censuswalk/censuswalk_test.go:100,135,267,292` + `hold_census_test.go:155` + ~24 more sites (updaterwire, updateauth, auth, store, trader `*_census_*_test.go`) write synthetic trees with `go.mod = "module nofx"` — **self-consistent fixtures: they pass unchanged after the rename** (their `censuswalk.ModulePath(root)` reads the synthetic go.mod). No test hazard — BUT see CHECK 5: every one of these lines hits the item-10 census guard. P2.

## CHECK 4 — pins, goldens, snapshots the rename will change

- `web/src/test/brand-scope-baseline.json` pins **16 files** by sha256: `auth/auth.go`, `go.mod`, `logger/logger.go`, `ninjascript/VLTraderTCPClient.cs`, `provider/ninjatrader/{tcp_framing.go,tcp_server.go}`, `web/src/lib/agentChatStorage.ts`, `deploy/nofx-claim.sh`, `deploy/nofx-db-backup.sh`, `deploy/nofx-lock.sh`, `deploy/nofx-web.service`, `deploy/nofx.service`, `deploy/systemd-user/nofx-backup.{service,timer}`, `nofx-clock-guard.{service,timer}`. The plan's "re-pin from renamed bytes with a dated owner-ruling line" covers all 16. ✓
- **MISSED by the plan: golden regeneration.** `internal/updaterjob/testdata/job_full.golden.json` (consumed by `internal/updaterjob/golden_test.go`) contains **17 `nofx` hits**: `"blocker": "attended recovery: nofx-updater recovery …"`, `/home/u/nofx-releases/…`, `…/nofx-bin`, `…/web/dist`. After R1b these values become `vl-updater`/`vl-releases`/`vl-bin` → the golden comparison test FAILS until the golden is regenerated. The plan never says "regenerate goldens". P1 for R1b's gate: add an explicit step — regenerate `job_full.golden.json` at the known-good renamed head and DIFF the regen (goldens must not silently absorb unrelated changes). [A]
- `job_requested.golden.json`: 0 hits. `kernel/testdata/futures_mnq_*.golden` (4 files): 0 hits (my earlier tool output listed them from a `find` name match, not a content match — content is clean). [A]
- Fixture `_provenance`/`_source` strings in `trader/testdata/w2-write-truth/plan_row_45{2,5}_subset.json` and `store/testdata/settings_truth_rows.json` carry `nofx` inside provenance comments (1 hit each) — test fixtures, flag for the census guard, no runtime effect. [A]
- No `*.snap` snapshot files exist under `web/src`. [A]

## CHECK 5 — the exact minimal census allow-list (item 10)

Raw census at `9d52f5dc6`: `git grep -Iil nofx` = **2,334 files**. Top buckets: trader 404, docs/superpowers/reports 251, kernel 189, api 120, trader/ninjatrader 65, internal/updaterworker 49, agent 45, store 38, deploy 31. After the rename, most vanish (imports, comments, docs). The guard as written — "`grep -rIic nofx` over the tree must equal an explicit allow-list" — will fire on **identifiers and comments**, not just name tokens: `nofxosClient`/`nofxosLang` (kernel), `NoFxFutures` fields (kucoin/gate/bitget), `provider/nofxos` package docs, `.audit/*.md` (10 tracked files). A raw-count guard is unmanageable; the guard must be **line-anchored to token patterns**, with these buckets:

**Proposed exact allow-list (patterns, not raw grep count):**
1. `provider/nofxos/**` + identifiers `nofxos*`/`NofxOS*` anywhere (third-party provider, plan-excluded). [A]
2. Broker exchange ids: `NoFxFutures` and the gate/kucoin/bitget client-id literals. [A]
3. Upstream history links: `github.com/NoFxAiOS/nofx` in docs/history. [A]
4. R1a compat fallbacks (after R1a lands): `NOFX_*` in `config.Env`, `${VL_X:-${NOFX_X…}}` shell twins, dual globs `nofx_*.log`/`nofx-bin`/`nofx.service` readers, one-line wrapper scripts. [A]
5. Dated history: `docs/superpowers/reports/**`, `CHANGELOG*`, `ninjascript/vltrader_tcp_README.md` legacy mention, `.audit/*.md` (10 files — plan does not name this directory; either allow as history or rewrite). [A]
6. **`internal/updateauth/admin.go:62` `passwordBindingDomain`** — keep the string UNCHANGED (it is a stored-data compatibility surface like a schema) + comment + allow. [A] P1.
7. **Test-fixture literals**: `"module nofx"` in ~30 synthetic-tree fixtures across 16 files (censuswalk/updaterwire/updateauth/auth/store/trader census tests) + `releasefixture`'s embedded `nofx-bin` + `/srv/nofx` + the three testdata `_provenance` hits. Either allow the exact `module nofx`/fixture patterns, or (better) have the fixtures write `censuswalk.ModulePath`-derived or `module vl` synthetic names. [A] P2.
8. Binance referral code. [B] (plan names it; I did not hunt the exact line — grep it during R1b.)

**Would wrongly allow / wrongly rename:** the plan allows "the compat fallbacks" without enumerating them — the fallback LINES live in R1a-touched files and are finite; R1b must pin the exact file list, else `nofx` planted in a R1a file escapes the guard. The plan would wrongly rename `passwordBindingDomain` and the synthetic `module nofx` fixtures (or wrongly red-light them). The plan never mentions `.audit/` or `job_full.golden.json`.

---

## FINDINGS (ranked)

- **P1-1** `internal/updateauth/admin.go:62` — `passwordBindingDomain = "nofx-updater/password-binding/v1\x00"`. Not in the plan. Renaming it breaks every existing password binding (owner re-enrollment); leaving it trips the item-10 guard. Fix to plan: "keep this domain string unchanged — it is a stored-data compatibility surface; add it to the census allow-list with a comment." [A]
- **P1-2** Golden regeneration unnamed. `internal/updaterjob/testdata/job_full.golden.json` (17 hits, consumed by `golden_test.go`) will fail the R1b gate until regenerated. Fix: add an R1b step — "regenerate job_full.golden.json at the renamed head; diff the regen; never regenerate mid-merge". [A]
- **P2-1** `internal/updaterjob/verdict_test.go:29` allowedImports map `"nofx/internal/updaterwire"` — missed by the plan's test list; fails loudly post-rename. Fix: add to the R1a item-4 derived-prefix list. [A]
- **P2-2** Census guard shape: raw `grep -rIic` cannot equal a finite allow-list (identifiers, comments, 2,334 files). Fix: token-pattern guard (forbidden patterns = `"nofx/…` imports, `module nofx`, `nofx-bin`, `nofx.service`, `NOFX_*`, `nofx_*.log`, `nofx-updater`, `X-NOFX-Update`, `nofxAI`, `nofx-go`) + the eight buckets above. [A]
- **P2-3** Env census: plan's "at minimum" list omits `NOFX_LIVE_TESTS`, `NOFX_DEMO_DB`, `NOFX_DEMO_VERIFY`, `NOFX_DEMO_TRADER`, `NOFX_REHEARSAL`/`NOFX_REHEARSAL_{TRADER,OUT,DB,BASE}` (and test-only `NOFX_TEST_*`). Fix: name the full set from `grep -rhoE '"(NOFX_[A-Z0-9_]+)"' --include='*.go'` or declare which are intentionally excluded. [A]
- **P2-4** Synthetic `module nofx` fixtures (~30 lines / 16 files) hit the guard; and `.audit/*.md` (10 tracked) is in no bucket. Fix: derive fixture names or allow-list the exact pattern; bucket `.audit/` as history. [A]
- **P3-1** Plan cites reverifier_test.go + sshsig_test.go as forbidden-list tests — they aren't (imports only). Fix the plan text; name verdict_test.go instead. [A]
- **P3-2** `branding/scope_test.go:137` `"nofx/config"` literal — the flip must update it. [A]
- **P3-3** `releasefixture/releasefixture.go:164,175` fixture `nofx-bin` — add a `vl-bin` case so R1a dual-accept is tested both ways. [A]
- **P3-4** `docker/Dockerfile.backend:52` `-o nofx` — name it in item 1. [A]

## POSITIVES (confirmed [A])

- The module rewrite is mechanically clean: build rc=0, vet rc=0 on 1,128 rewritten files; no embeds/go:generate/ldflags/reflection on the module path; `censuswalk.ModulePath` self-corrects.
- The pass-on-empty hazard list the plan DOES name (release_test.go:997, branding/scope_test.go:48) is real and the `censuswalk.ModulePath` derivation (hold_census pattern) is the right fix.
- The 16-pin re-pin procedure is sufficient for the pinned files; `brand-i18n.test.ts` continuing to forbid NOFX in visible strings is consistent.
- R1a's dual readers cover every `nofx-bin`/`nofx.service`/`nofx_*.log` reader I found in the updater/activation paths.

## NOT done (L-AUDIT)
- No code change anywhere; the simulation worktree was deleted after check 1; report branch carries only this file. No deploy, no unit stop/start, no DB write, no lock.
