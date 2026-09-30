# AUDIT2-RENAME-PARTNER-GITHUB — DS-106 (round 2, READ-ONLY)

Slice: PARTNER + GITHUB RENAME + RELEASE PIPELINE (R4/D4) + folds F16, F24, F28. Plan v2 audited: `/home/hoang/rename-vl-plan/2026-09-29-rename-nofx-to-vl-plan-v2.md`, **md5 `7b8dc9a9a196e5554203db7a3b71aa7b`** (matches the dispatch). Code base: `origin/dev` @ `9d52f5dc6`. Round-1 report for this slice: #271 (DS-107) — read; no repeat findings below.

## Fold verdicts

- **F16 — CORRECT but INCOMPLETE.** [A] `internal/updaterworker/release_contract_test.go:39` pins `RELEASE_REPO: johnwick2921-cyber/nofx`; `gh release create --repo "$RELEASE_REPO"` at `.github/workflows/release.yml:275`. Moving the flip to R4 in the same commit as the pin is right. But the SAME ordering constraint applies to `deploy/install-updater-worker.sh:33` `REPO_URL="${NOFX_UPDATER_BUILD_REPO:-https://github.com/johnwick2921-cyber/nofx}"` (used at `:89` `git clone -q "$REPO_URL"`), which F16 does not cover → P1-1 below.
- **F24 — CORRECT and implementable.** [A] The dbcompat mechanism is the release.yml CI job (`:48` `dbcompat`, `:94 needs:[dbcompat]`, `:213-216` computed `ROLLBACK_PAIRS`) + `deploy/release/resolve-rollback-old.sh`, which resolves the OLD from git tags by version and reachability — entirely name-agnostic, so `v2026.09.27.2 → first VL release` is exactly what it will prove. The first VL release's OLD = the highest reachable `v*` tag = `v2026.09.27.2` [B]. No blocker. Consistency gap → P3-1.
- **F28 — CORRECT but INCOMPLETE.** "Partner repo is NOT renamed" and the runbook update assigned to D4 are consistent with D4's text. But the fold stops at the repo name and the census allow-list; it never defines what the partner tree's RELEASE PIPELINE identity becomes → P1-2 below.

## P1-1 — the updater build-script repo URL flips at R1b, but the GitHub repo is renamed at R4: any build from source in that window clones a URL that does not exist

- [A] `deploy/install-updater-worker.sh:33` defaults `REPO_URL=https://github.com/johnwick2921-cyber/nofx`; `:89` `git clone -q "$REPO_URL" "$BUILD_DIR/src"`. D2 (R1b) item 1 says "update every reference, release.yml, install-updater-worker.sh" and item 7 says "install-updater-worker.sh repo URL github.com/johnwick2921-cyber/vl" — so REPO_URL becomes the NEW name at R1b.
- [B] GitHub redirects only in the OLD→NEW direction. Between the R1b merge and the R4 Settings rename, `https://github.com/johnwick2921-cyber/vl` does not exist → the clone REFUSES. R2 step 0's "new updater built from the same sha" is the exact consumer of this path.
- **Exact replacement text (fold into F16):** "F16 becomes: `RELEASE_REPO` AND `install-updater-worker.sh`'s `REPO_URL` default both move from `johnwick2921-cyber/nofx` to `johnwick2921-cyber/vl` in the SAME commit, at R4 (after the Settings rename), together with the `release_contract_test.go:39` literal; until R4 both stay `…/nofx` (redirects cover them afterwards, never before)."

## P1-2 — D4 never defines the partner tree's release-pipeline identity: a byte-identical sync carries `RELEASE_REPO=johnwick2921-cyber/vl` into vlautoagenttraderv1

- [A] `.github/workflows/release.yml` publishes to `RELEASE_REPO` (used at `:275`); `deploy/release_contract_test.go:66` forbids the string `vlautoagenttraderv1` in release.yml — i.e. the contract test only REFUSES the partner name; it does not pin the owner's name. A byte-identical tree sync therefore ships a partner repo whose release workflow would publish signed releases into `johnwick2921-cyber/vl` (the owner's repo) from the partner machine's secrets, and whose own contract test would still PASS (it forbids `vlautoagenttraderv1`, which is absent).
- D4 says "bring vlautoagenttraderv1 fully up to the renamed tree … match table (every path, blob sha)" and F28 says the partner repo is not renamed — neither says what release.yml/RELEASE_REPO/REPO_URL become in the partner tree, nor how PARTNER-SYNC-BOOT4 (vlauto PR #13) handled the same problem.
- **Exact replacement text (append to D4):** "After the tree sync, patch the partner-only identity: `release.yml` `RELEASE_REPO=vlautoagenttraderv1`, `install-updater-worker.sh` `REPO_URL` default `github.com/vlautoagenttraderv1`, and the partner's `release_contract_test.go`/census allow-list entries to match; carve exactly those files out of the match table and name the carve-out. No partner CI run may ever publish into `johnwick2921-cyber/*` — the PR must state this verbatim."

## P2-1 (consistency) — R1b item 7 text still carries the pre-F16 RELEASE_REPO flip

- [A] Plan v2 "Phase R1b" item 7 reads: "… `install-updater-worker.sh` repo URL github.com/johnwick2921-cyber/vl …; `release.yml` RELEASE_REPO=johnwick2921-cyber/vl (+ `release_contract_test.go`, which still forbids `vlautoagenttraderv1`)." This directly contradicts F16 (which moved the flip to R4). The folds header says folds override, so no behavioural risk — but D2 is sent as-written and DS-107 executes the numbered items; the contradiction invites the exact pre-F16 bug.
- **Exact replacement text (R1b item 7):** "… `install-updater-worker.sh` repo URL and `release.yml` RELEASE_REPO stay `…/nofx` through R4 — see F16 — and are NOT touched in this item."

## P3-1 (consistency) — R4 step 2 does not name the F24 precondition

- [A] R4 step 2: "First signed release under the new name … installed with the button … That is also the end-to-end proof of the rename." F24 requires the cross-rename dbcompat pair to pass before the first VL release is advertised; the phase text doesn't mention it (folds override, but the owner reads R4 first).
- **Exact replacement text (R4 step 2):** append — "the release is advertised only after the CI `dbcompat` job proves the `v2026.09.27.2 → this-release` rollback pair (F24)."

## Verified clean (no finding)

- [A] Release signing is repo-name-free: `ssh-keygen -Y verify -f deploy/release_allowed_signers -I release -n release` (`release.yml:252`); the secret is `RELEASE_SIGNING_KEY` (`:235`) — secrets and the `release` environment (verified earlier: 1 environment, required-reviewer rule on the owner) follow the repo through a rename [B].
- [A] The bot has no `api.github.com` dependency (grep: zero hits in Go) — "the update page's release downloads" are the local release inbox, so the R4 redirect claim is moot but harmless.
- [A] Default branch `dev`, `dev` unprotected, 30 open PRs at audit time (from round 1) — redirects keep all of them functional [B].
- [A] Worktree/release-clone remotes keep working through redirects after R4; `resolve-rollback-old.sh` is tag/name-agnostic so F24's pair resolves post-rename.

---

NO NEW P0. NEW P1=2 (P1-1 REPO_URL flip ordering; P1-2 partner release-pipeline identity). Plus P2-1/P3-1 consistency items and three fold verdicts (F16/F28 incomplete-as-written, F24 correct). READ-ONLY throughout; no code changes, no tests (L-AUDIT).
