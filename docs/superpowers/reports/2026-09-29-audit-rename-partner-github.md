# AUDIT-RENAME-PARTNER-GITHUB — READ-ONLY adversarial audit (DS-107, 2026-09-29)

**Branch** `audit/rename-partner-github` · **base** `git log -1 origin/dev` = `9d52f5dc6` "RELEASE 4d538206… — boot 7 …" · plan md5 `b16759336fe0e264d11c1bc885ce5192` ✓ matches the dispatch's `b1675933`.
**Scope:** slice PARTNER REPO + GITHUB RENAME + RELEASE PIPELINE (R4 / D4). READ-ONLY — no code touched; only this report file is new.
**Evidence tiers:** [A] read the exact line / ran it · [B] inferred from strong evidence (GitHub docs behaviour) · [C] speculation.

## Summary

The release pipeline is the plan's thinnest spot: one pinned test and one ordering decision can make the very first release after the rename either fail CI or target a nonexistent repo. No P0s; **2 P1, 6 P2, 5 P3**, each with the exact fix to the plan/dispatch text. The GitHub rename itself is the safest part (everything the machine depends on is repo-ID-keyed or redirects).

---

## P1-1 — `RELEASE_REPO` flips in R1b, but the repo rename is R4: the release job targets a nonexistent repo for the whole window

- **Where:** plan "Phase R1b item 7: release.yml RELEASE_REPO=johnwick2921-cyber/vl" vs "Phase R4 step 1 (rename) + step 2 (first signed release)". Evidence: `.github/workflows/release.yml:39` `RELEASE_REPO: johnwick2921-cyber/nofx` [A]; `:271-275` `gh release create --repo "$RELEASE_REPO"` [A].
- **Why it matters:** between the R1b merge and the R4 rename, a tag release fails hard (repo `vl` does not exist yet). Worse: if the R2 boot misbehaves and the owner needs a HOTFIX release before the rename, the release job is dead — the one tool that exists exactly for that moment.
- **Fix (exact plan text):** delete the RELEASE_REPO line from R1b item 7 and add to R4: "Step 0 (same commit as the rename): flip `RELEASE_REPO` to `johnwick2921-cyber/vl` together with the contract-test literal (P1-2). Until the rename, the job keeps the old name." Better still: `RELEASE_REPO: ${{ github.repository }}` — the workflow then follows the rename automatically and can never point at a nonexistent repo.

## P1-2 — `release_contract_test.go` pins the OLD repo literal; D2 as worded reddens CI at the R1b head

- **Where:** `deploy/release_contract_test.go:66` forbids `vlautoagenttraderv1` AND `:75-80` requires release.yml to contain the literal `johnwick2921-cyber/nofx` [A]. D2 item 7 says only "release_contract_test.go, which still forbids vlautoagenttraderv1" — it names HALF the test.
- **Why it matters:** flipping RELEASE_REPO without flipping the literal pin makes `TestReleaseWorkflowDefaultsToThisRepoAndNeverThePartner` fail at the R1b head — a red CI on the rename PR, guaranteed.
- **Fix:** extend D2 item 7: "update `TestReleaseWorkflowDefaultsToThisRepoAndNeverThePartner`: the `johnwick2921-cyber/nofx` literal becomes `johnwick2921-cyber/vl` (or the `github.repository` expression); the `vlautoagenttraderv1` prohibition stays."

## P2-1 — the updater binary names in the pipeline are not on the rename list

- **Where:** `.github/workflows/release.yml:177-185` builds `-o "$OUT/nofx-updater"` / `-o "$OUT/nofx-updater-bootstrap"` into `updater/` [A]; `deploy/release/package.sh:28` OPTIONAL=`( "updater/nofx-updater" "updater/nofx-updater-bootstrap" … )` [A]. D1 cites package.sh ~19 (the ALLOW line) but not :28; D2 item 1 renames `cmd/` dirs but never names these two build lines.
- **Why it matters:** a post-rename archive would still ship `nofx-updater` binaries. The activation readers accept either NAME (R1a), so it works — but the owner's "one name" goal fails, the census allow-list must then cover two more names, and the archive-name allow-list in `check-archive-paths.sh` (if it pins them) silently conflicts.
- **Fix:** add to D2 item 2: "release.yml:177-185 and package.sh:28 become `updater/vl-updater` + `updater/vl-updater-bootstrap`; check-archive-paths.sh updated in the same commit."

## P2-2 — the cross-rename dbcompat/rollback pair must be explicit and green

- **Where:** `.github/workflows/release.yml:48-89` — dbcompat's OLD comes from `resolve-rollback-old.sh` = highest v\* ancestor tag = `v2026.09.28.1`, a PRE-rename commit [A]; `deploy/release/manifest.sh:6-7` advertises unproven pairs `tested:false` [A].
- **Why it matters:** the first post-rename release's rollback target is a pre-rename binary. That is exactly the rollback the owner would want after the R2 boot. The plan never says the pair must be proven; if dbcompat is skipped or fails, the cross-rename rollback is advertised untested.
- **Fix:** add to R4 step 2: "the first new-name release must run its dbcompat job green against the pre-rename boot tag (the v2026.09.28.x ancestor) — the manifest then advertises the cross-rename pair `tested:true`, which is the rename's rollback proof."

## P2-3 — the old updater binary is never retired at R2

- **Where:** D3 step 4 installs `~/bin/vl-updater` but does not move/remove `~/bin/nofx-updater` (plan text). Evidence for the hazard is the plan's own R1a motivation: the old reader accepts only `nofx-bin`.
- **Why it matters:** one stray manual run of the old updater against a post-rename archive rolls back a GOOD install.
- **Fix:** D3 step 4: "`mv ~/bin/nofx-updater ~/bin/nofx-updater.old.<sha>` (and `nofx-activate`) — the pre-rename reader cannot parse post-rename archives; keep the copies as rollback-only artifacts."

## P2-4 — D1's "shell twin" script list omits NOFX_-reading scripts

- **Where:** D1 item 3 lists postboot-check.sh, planner-ab-report.sh, cutover.sh, nofx-db-backup.sh. Actual readers NOT listed: `deploy/install-updater-worker.sh:33` (`NOFX_UPDATER_BUILD_REPO`) [A], `install-autostart.sh`, `start.sh` (10 hits), `install.sh`, `nofx-claim.sh`, `nofx-lock.sh` [A grep]. Distinct env names found outside the dispatch's minimum: `NOFX_REPO, NOFX_DATA, NOFX_BIN, NOFX_ENV, NOFX_DIR, NOFX_INSTALL, NOFX_UNIT, NOFX_GATE_URL, NOFX_HEALTH_URL, NOFX_UPDATER_BUILD_REPO, NOFX_UPDATER_INSTALL_DIR, NOFX_LOCK_* (5), NOFX_KEEP_* (4), NOFX_BACKUP_*, NOFX_BACKEND_PORT, NOFX_FRONTEND_PORT, NOFX_DB, NOFX_DB_RESEARCH, NOFX_M, NOFX_USER__` [A census].
- **Why it matters:** the plan says "full list in the identifier map" but the map is not delivered; a lane taking "at minimum" literally leaves owner-set env values silently unread after R1b.
- **Fix:** D1 item 3: "the shell twin goes into EVERY deploy script that reads NOFX_\* — enumerate: install-updater-worker.sh, install-autostart.sh, start.sh, install.sh, postboot-check.sh, planner-ab-report.sh, cutover.sh, nofx-lock.sh, nofx-claim.sh, nofx-db-backup.sh, nofx-clock-guard.sh; ship the full identifier map in the PR body."

## P2-5 — the partner runbook update is unassigned

- **Where:** R4 step 3 "per the updated partner runbook". `docs/superpowers/runbooks/2026-09-22-vl-partner-update.md` exists ONLY in vlauto (partner-only file, proven in the boot-4 sync match table) [A]; the nofx-side #240 runbook is still an OPEN DRAFT [A gh pr view 240]. No dispatch (D0–D4) is tasked to update the vlauto copy.
- **Why it matters:** the partner machines' steps (unit `vl`, `vl-bin`, `~/vl`, migrate-to-vl.sh, AddOn VL_BUILD_ID, one-order check) would follow a runbook that still says nofx.
- **Fix:** add to D4: "update the partner-only runbook in the same PR — and state that the 6 deploy scripts with hardcoded `/home/hoang` paths (`bars-key-rollback.sh`, `leveltruth-cutover.sh`, `nofx-clock-guard.sh`, `nofx-db-backup.sh`, `planner-ab-report.sh`, `postboot-check.sh`) [A grep] are owner-box-only; on the partner boxes they already require the env overrides, and `migrate-to-vl.sh` must derive every path from `$HOME`/`SUDO_USER` (install-autostart.sh:26-34 already does the portable-user pattern [A])."

## P2-6 — D4 wording ambiguities (the dispatch's check #4)

- **Where:** D4: "bring vlautoagenttraderv1 fully up to the renamed tree (the booted sha)".
- **Why:** three reads are possible: (a) "booted sha" = R1b merge sha? stamp tip? R2 boot build sha?; (b) is the PARTNER REPO also renamed? (the plan never says; vlauto keeps its name); (c) the preserved partner-only docs contain `~/nofx`/`nofx.service` — the renamed tree ships the census guard, which will hit those docs on vlauto unless they are allow-listed.
- **Fix:** rewrite D4's first paragraph: "…up to the renamed tree at the R2 BOOT BUILD SHA (the R1b merge commit the R2 boot was built from — name it when dispatched). The partner repo `vlautoagenttraderv1` is NOT renamed; only its tree changes. The two preserved partner-only docs stay and are added to the census allow-list as dated history (docs/ pattern)."

## P3-1 — cross-rename rollback keeps unit `vl` running a nofx-named binary

Functional via R1a's dual readers [A plan text], but state explicitly: "the dual readers stay until the owner declares the old names dead — they are never silently removed in a later cleanup" (otherwise a future hygiene wave re-breaks the cross-rename rollback).

## P3-2 — GitHub rename facts the plan should state for the owner [B]

- Redirects: web URLs, git clone/pull/push, release-asset download URLs, `raw.githubusercontent.com` — all redirect for the old name; API calls return 301 (gh and curl -L follow) [B, GitHub "Renaming a repository" docs].
- NOT covered by redirects: the old name becomes available for re-creation IMMEDIATELY — if anyone recreates `johnwick2921-cyber/nofx`, old URLs silently serve the wrong repo. Add to R4: "do NOT recreate the old repo name."
- Survives the rename (repo-ID-keyed) [B]: Actions environments (`release`, release.yml:98 [A]) and secrets (`RELEASE_SIGNING_KEY`, release.yml:235 [A]), branch protection, required reviewers, open PRs (their URLs redirect).
- Tags: `v2026.09.28.x` history moves with the repo [B], so `resolve-rollback-old.sh` keeps finding the pre-rename ancestor [A release.yml:81].
- `gh` on the owner's box follows the redirect for old names; the plan's "CTO updates git remotes" is for git, not gh [B].

## P3-3 — FAQ/README install-command links

`FAQContent.tsx:93,102` old repo links are covered by D2 item 9 [A]. `translations.ts:739/2136/3521` and `README.md:44/191/201` install commands point at UPSTREAM `NoFxAiOS/nofx` raw URLs — the plan correctly keeps upstream links, but those commands install the upstream product; pre-existing mismatch, out of rename scope (note only).

## P3-4 — db-compat temp names

`deploy/release/db-compat.sh:152` `OLD_BIN="$WORK/nofx-old"` / `NEW_BIN="$WORK/nofx-new"` [A] — cosmetic variable names; rename them in the deploy-tests sweep (D2 item 7) so the census allow-list does not grow for nothing.

## P3-5 — signing key files and secret-scan prose

`deploy/release/README.md:26-31` key FILE names (`nofx-release-key`) are local-only and shredded; the committed `release_allowed_signers` ident is `release` [A] — **no signing change is needed across the rename** (state this in R4 so the owner does not rotate the key). `secret-scan.sh:13` comment mentions `~/nofx-backups` — cosmetic, covered by item 9's docs sweep.

## Check answers (dispatch CHECK 1–4)

1. **GitHub rename:** see P3-2 + P1-1/P1-2. The machine-facing surface (updater) has NO network source — `cmd/nofx-updater/main.go:75`: "fetch reads `<inbox>/<release_id>.tar.gz`. There is no network source in v1 (brief C9)" [A] — the owner's `gh release download` is the only GitHub touchpoint, and it redirects.
2. **Release pipeline names:** release.yml:39 (covered), :134 `nofx-bin` (covered, item 2), :177-185 updater names (P2-1), package.sh:19 (covered) + :28 (P2-1); contract-test literal (P1-2). Cross-release install: AFTER-updater + PRE-rename archive works (R1a readers) [A plan]; PRE-updater + POST-rename archive is prevented only if the old binary is retired (P2-3). dbcompat cross-rename pair: P2-2.
3. **Partner sync after rename:** repo not renamed (P2-6b); partner-only docs vs the census guard (P2-6c); runbook unassigned (P2-5); migrate-to-vl.sh is user-portable IF it follows the install-autostart.sh pattern — make it explicit (P2-5); hardcoded `/home/hoang` scripts are owner-box-only (P2-5).
4. **D4 wording:** P2-6(a–c) + P3-1.

## What this audit did NOT do

No code change, no unit stop/start, no DB access, no lock acquire, no GitHub actions taken, no merge. Only this report file on a dedicated branch. L-AUDIT honored.
