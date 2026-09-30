# AUDIT-RENAME-OPS — DS-106 (operations half, READ-ONLY)

Adversarial audit of the VL rename plan, ops half. Plan audited: `/home/hoang/rename-vl-plan/2026-09-29-rename-nofx-to-vl-plan.md`, **md5 `b16759336fe0e264d11c1bc885ce5192`** (verified against the dispatch's `b1675933`). Spec base for code citations: `origin/dev` @ `9d52f5dc6` (boot 7, 2026-09-28 01:01 CT). Zero writes to anything outside this report; every unit/state observation is `systemctl cat`/`ls`/`readlink` read-only.

## Summary

The plan is right about the shape of the risk (R1a dual readers first, one rehearsed reversible script, symlinks) and the timeline is honest. It has **one P0** (R2 step 2 as written cannot complete: the symlinked release dir is refused by the updater's own guard, and the plan's re-check has no remedy), **two P1s** (the web unit is missing from R2 entirely; the rollback leaves `vl` units enabled → double bot on the next boot), and several P2/P3 wording gaps. All findings below with [A]/[B]/[C] and the exact fix to the plan text.

---

## P0-1 — R2 step 2 cannot complete: symlinking `~/nofx-releases` is refused by the updater's own release-root guard, and the plan's re-check has no remedy

- [A] `~/.config/nofx-updater/env` line 1: `NOFX_RELEASE_DIR=/home/hoang/nofx-releases` (the dir exists — `stat` says `directory`).
- [A] `internal/updaterworker/releaseroot.go:46-55` REFUSES any release root that is "not its own resolved path (it resolves to %s) — name the real directory, with no symlink in it" (`filepath.EvalSymlinks(root)` must equal the configured string).
- After R2 step 2's `mv ~/nofx-releases ~/vl-releases` + `ln -s ~/vl-releases ~/nofx-releases`, the configured `NOFX_RELEASE_DIR=/home/hoang/nofx-releases` resolves to `/home/hoang/vl-releases` ≠ configured → **every future release materialization is refused, and the plan's own "re-run the release-dir-outside-install check … and refuse if it now fails" turns step 2 into a hard stop with no defined remedy** — the step as written is a trap, not a check.
- The plan's "the owner's .env keeps working untouched (NOFX_* fallback)" covers the BOT's `.env`, not the UPDATER env file (`~/.config/nofx-updater/env`, only 2 keys: `NOFX_RELEASE_DIR`, `NOFX_CUTOVER_TOKEN`). The updater env is nowhere in D3.
- **Fix to plan text (R2 step 2):** "…and in the SAME step rewrite the updater env file (`~/.config/vl-updater/env`, created from `~/.config/nofx-updater/env`) so `VL_RELEASE_DIR` (or the dual-read `NOFX_RELEASE_DIR`) names the REAL new path `$HOME/vl-releases` — a symlinked release dir will be refused by releaseroot.go:46-55 by design." Alternatively: do not move `~/nofx-releases` at all (leave the name; only the bot home must be `vl`).

## P1-1 — `nofx-web.service` exists, is ACTIVE and ENABLED, and is missing from R2/D3 entirely

- [A] `/etc/systemd/system/nofx-web.service`: `Description=NOFX frontend (vite dev server :3000)`, `WorkingDirectory=/home/hoang/nofx/web`, `ExecStart=/usr/bin/npm run dev`. `systemctl is-active nofx-web` → `active`; `is-enabled` → `enabled`.
- [A] The live listener is that server: `ss -ltnp` shows `127.0.0.1:3000` owned by node pid 427, and `/proc/427/cwd` → `/home/hoang/nofx/web`. `nginx` is `inactive` — the web UI on this box IS the vite dev server under `nofx-web.service`.
- The plan's R2 handles only `nofx.service` (steps 1, 3). R1b creates `deploy/vl-web.service` but nothing ever installs it or removes the old unit. After the rename the box would run `vl.service` (bot) + a still-enabled `nofx-web.service` — the web keeps working through the symlink, so it is not a break, but it is a permanent `nofx` name the rename was supposed to remove, and it is not covered by any rollback either.
- **Fix to plan text (R2, add step 3b):** "stop and disable `nofx-web`, install `/etc/systemd/system/vl-web.service` from `deploy/vl-web.service` (same placeholders as `install-autostart.sh` fills), `systemctl enable --now vl-web`". And in `--rollback`: "stop+disable `vl-web`, re-enable `nofx-web`".

## P1-2 — `--rollback` leaves the `vl` units enabled → two bots and two backup timers on the next boot

- Plan rollback text: "stop `vl`, `rm` the symlinks, `mv ~/vl ~/nofx` (and the other dirs), restore nofx-bin, re-enable `nofx` + the old user units, start, verify the old boot line."
- It stops `vl` but never `disable`s `vl.service`, never removes `/etc/systemd/system/vl.service`, and never disables the `vl-backup`/`vl-clock-guard`/`vl-updater` user units it enabled in step 4. `nofx.service` is `enabled` today [A]; `vl.service` would be `enabled` after step 3. After a rollback the box has BOTH enabled → next boot starts `vl` AND `nofx` — two bots, double ledger writes, and two backup timers on one DB.
- Also unhandled by the rollback text: `~/bin/vl-updater` and `~/.config/vl-updater/env` (the old `nofx-updater` is never deleted, so restoring = disabling the new unit — must be said), and the rewritten updater env from P0-1 (rollback must restore the old value).
- **Fix to plan text (`--rollback`):** "stop `vl`, `systemctl disable vl`, `rm /etc/systemd/system/vl.service`, `daemon-reload`; disable `vl-backup.timer`, `vl-clock-guard.timer`, `vl-updater`; restore `NOFX_RELEASE_DIR` in `~/.config/nofx-updater/env`; then `rm` the symlinks, `mv ~/vl ~/nofx` …, restore nofx-bin, re-enable `nofx` + the old user units, start, verify the old boot line."

## P2-1 — the lock hand-over has an unlocked window a third session could enter

- Plan step 5: "release the OLD lock (old script), acquire+release once with the new script to prove `~/vl-main.lock.d`". Between release-old and acquire-new, NO lock exists: another session can `deploy/nofx-lock.sh acquire` the old home (`~/nofx-main.lock.d`, still valid via the symlink — lock dir is `$HOME/nofx-main.lock.d` [A], and the dir does NOT currently exist [A]) while the migration session holds nothing. Two sessions can then each believe they own the main tree (one via the old script, one via the new), which is exactly the two-lock-homes failure the R1b refusal rule exists to prevent.
- The R1b rule ("new script REFUSES while the old lock dir exists") makes acquire-new-before-release-old impossible as written — so the window is inherent to the plan's ordering. Mitigated, not closed: the D-FREEZE means no lane acts during R2, and the owner is present — but a two-lock-homes split is exactly the class-70 shape this program exists to prevent.
- **Fix to plan text (R1b + D3 step 5):** the `vl-lock.sh` refusal exempts the migration session: "refuse while `~/nofx-main.lock.d` exists UNLESS its `meta` names the same `--session` that is calling acquire — then and only then the new lock may be acquired FIRST, and the migration releases the old lock immediately after." D3 step 5 becomes: "acquire the NEW lock with the same `--session` (exempted by the old-lock holder match), then release the old lock; a lock home is never empty while the bot is stopped."

## P2-2 — D3 "ONE `sudo` inside" cannot be right as written

- The script must: `systemctl stop nofx` (system unit, needs root), write `/etc/systemd/system/vl.service` (root), `systemctl disable nofx` / `enable --now vl` (root), and `mv ~/nofx ~/vl` (owner-owned, no root). "ONE `sudo` inside for the system unit" is ambiguous: if read literally as one `sudo <cmd>` invocation, steps 1 and 3 conflict.
- **Fix:** "The whole script is invoked ONCE under `sudo` (the preconditions in step 0 run BEFORE elevation, inside the script, and re-verify after); the elevated section performs stop/mv/unit-install/enable as one sequence. No `sudo` appears anywhere else."

## P2-3 — R4 wording gaps (GitHub rename)

- [A] `gh api repos/johnwick2921-cyber/nofx` → public; `/branches/dev/protection` → 404 (dev is NOT protected — "branch protections" in the plan is about a thing that does not exist on dev); open PRs at audit time: **30** (`/pulls?state=open --jq length`). The `release` environment EXISTS (`total_count:1`) with one required-reviewers protection rule (reviewer `johnwick2921-cyber`) [A].
- [B] GitHub redirects (HTTP 301) cover old-name remotes, `gh`, API release listings and raw URLs after the rename — environments/protection rules/open PRs are repo-scoped and follow the rename. The risk is not redirects; it is the things that VERIFY the name: `deploy/install-updater-worker.sh:33` `REPO_URL=https://github.com/johnwick2921-cyber/nofx` (R1b fixes it ✓) and `.github/workflows/release.yml:39` `RELEASE_REPO` (R1b fixes it ✓) — any straggler works through the redirect but should be named, because the partner sync's match table compares URLs.
- [B] "CTO updates `git remote set-url` on every worktree and the release clones" — `git worktree list` shows the main tree + ~40 lane worktrees [A]; the release clones are 7 `~/nofx-release-*` dirs [A]. Hand-updating each is error-prone.
- **Fix:** "CTO runs one loop: `for w in $(git worktree list --porcelain | awk '/^worktree/{print $2}') ~/nofx-release-*; do git -C "$w" remote set-url origin git@github.com:johnwick2921-cyber/vl.git; done` — plus the partner repo's remote on its own machine; redirects are the fallback, never the plan."

## P2-4 — `~/nofx-releases` may not exist on partner machines; D3 must skip missing dirs

- [A] On THIS box `~/nofx-releases` exists, but `~/nofx-releases` is created by the updater materializing releases — a machine that never installed one won't have it. D3 step 2's blanket `mv` list has no "if it exists".
- **Fix:** "for each of `~/nofx-backups`, `~/nofx-releases`, `~/nofx-inbox`, `~/.config/nofx-updater`: `[ -e <path> ] && mv … && ln -s … || log 'skipped (absent)'`" — and per P0-1, `~/nofx-releases` is either not moved or the env file is rewritten in the same step.

## P2-5 — the `~/.claude` "unaffected thanks to the symlink" claim is only half right

- [A] `~/.claude/projects/` holds 40+ `-home-hoang-nofx*` memory dirs (main tree `-home-hoang-nofx`, one per lane worktree). Claude/Codex key memory dirs by the path STRING; the symlink preserves the string, so EXISTING memory keeps working — but any session opened at `~/vl` after the rename starts a NEW, EMPTY `-home-hoang-vl` memory dir. The plan's owner-steps say "~/.claude memory path is unaffected thanks to the symlink" — true for continuity, false for "no change": memory splits unless the owner keeps opening `~/nofx`.
- **Fix:** append to Owner steps: "continue opening the main tree as `~/nofx` (the symlink) in VS Code/Claude for as long as the old memory matters; `~/vl` is a new path and starts empty memory."

## P3-1 — `~/nofx-main.lock.d` does not currently exist; step 0's "old lock held" is a precondition, not a discovery

- [A] `ls -d ~/nofx-main.lock.d` → absent. The migration acquires it in step 0 — correct, but D3 step 5's hand-over must not assume it was held BEFORE the migration started (it wasn't). Minor wording: "release the OLD lock (old script) [held since step 0]".

## P3-2 — the updater's baked names survive into the new binary's transition window

- [A] `strings ~/bin/nofx-updater` contains 1,299 `nofx` hits including `nofx-backups`, `nofx-lock.sh`, `nofx-updater` (self-name in messages). Step 1 stops it before the move ✓ and step 4 installs `vl-updater` ✓ — but the user unit today runs `%h/bin/nofx-updater --install-dir %h/nofx serve` [A]; the NEW unit must pass `--install-dir %h/vl` (through the symlink `%h/nofx` also resolves, but the unit should carry the real name). **Fix:** name the new unit's `ExecStart=%h/bin/vl-updater --install-dir %h/vl serve` explicitly in D3 step 4.

## P3-3 — lane Stop hooks and VS Code carry the path as a STRING, not a symlink-resolved path

- [A] `~/.agents/hooks/hooks.json` contains no `nofx` path strings; the lane Stop hook that blocks my turns lives on the VS Code side keyed by the workspace folder URI `/home/hoang/nofx` — a string that keeps resolving through the symlink. [B] No VS Code-side break expected. The plan does not need a change; this is recorded so nobody "fixes" a non-problem by editing hooks.

---

## What I did NOT audit

The code half (R1a/R1b/R3 readers, dual-reader completeness) — that is DS-101's half. Partner-machine disks (remote, cannot read) — all partner statements are [C]/plan-text checks. The D0 wording fold on #267 (DS-103's dispatch, not ops). No tests run (L-AUDIT read-only; report-only deliverable).
