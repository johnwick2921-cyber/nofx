# AUDIT2-RENAME-OPS — READ-ONLY adversarial audit, round 2 (DS-107, 2026-09-29)

**Branch** `audit2/rename-ops` · **base** `git log -1 origin/dev` = `9d52f5dc6` ("RELEASE 4d538206… — boot 7 …") · plan v2 md5 `7b8dc9a9a196e5554203db7a3b71aa7b` ✓ matches `7b8dc9a9…`.
**Slice:** OUTSIDE-THE-REPO + lock hand-over + D-FREEZE + canon docs + folds **F5, F7, F25, F26**. Round-1 report read first: #272 (DS-106, `audit/rename-ops`) — its P1-1/P1-2/P2-1/P2-2 map to F5/F25/F7/F26.
**L-AUDIT honored:** no code change, no unit touch, no DB, no lock; only this report file.

## Verdict

The four folds are **correct in substance** (each re-verified against the cited code at origin/dev) but the plan's own promise — "Every dispatch below is sent as written (updated with the audit folds)" — is **false for D3 and D-FREEZE**: the texts shown in v2 still contain the exact steps F1–F8 and F26 correct. A lane reading only its dispatch would rebuild four of the bugs round 1 found. **NEW P0=0 P1=2** (both are fold-integration gaps), plus 2 P2 and 3 P3.

---

## NEW P1-1 — D3 (and the R2 phase text) contradicts binding folds F1, F2, F5, F7, F8 in five concrete steps

- **Where:** plan v2 "Phase R2 steps 1–6" + "D3 JOB steps 0–6" vs folds F1/F2/F5/F7/F8 (header says folds are BINDING and "override any phase/dispatch text that disagrees" — but omissions and orderings are not disagreements a diff would catch).
- **Why it matters:** the D3 text is what DS-104 and the owner's script actually execute. Five contradictions, each with the evidence:
  1. **D3 step 2** says `~/nofx-releases` gets the "same pattern … (→ vl names + symlinks)". **F1** forbids symlinking the release dir: `internal/updaterworker/releaseroot.go:46-55` [A, CTO-confirmed in the fold] REFUSES a release root that is not its own resolved path — a symlinked `~/nofx-releases` breaks every fetch. Fix per F1 = move to `~/vl-releases` and rewrite the updater env to the real path.
  2. **D3 step 3 vs step 6 order:** the shown text disables `nofx` at step 3 and verifies at step 6. **F2** says stop nofx → start vl → **VERIFY** → only then disable nofx and park `nofx-bin`.
  3. **D3 step 5** says "release the old lock … then acquire+release once with the new script". **F7** says acquire the NEW lock FIRST (same-session exemption) then release the old — the shown order is exactly the unlocked window round 1 found.
  4. **D3 step 4** does not retire the old updater binary. **F8** requires renaming `~/bin/nofx-updater` to `.old` at R2.
  5. **D3 omits the web unit entirely.** **F5** (DS-106's P1-1): `nofx-web.service` is the live :3000 page [A: `deploy/nofx-web.service` template exists in-tree; the fold is CTO-confirmed] — the shown D3 never stops/disables it nor installs `vl-web.service`.
- **Fix (exact replacement):** regenerate the D3 JOB text (and Phase R2 steps 1–6) from F1–F8 verbatim — the five bullets above, in F2's order — and add to the plan header: "the dispatch texts below were regenerated from the folds; where they disagree, the fold text wins and is quoted inline." This must be verified before D3 is sent, not left to the lane.

## NEW P1-2 — D-FREEZE (with F26) contradicts D3, which is dispatched in parallel to a lane ON the freeze list

- **Where:** D-FREEZE addressee list = "DS-101, DS-102, DS-104, DS-105, DS-106, DS-108" and its rule "push nothing to a branch you intend to merge"; the timeline dispatches **D3 to DS-104 "in parallel with D1/D2"**, and D3's own text says "rebase onto the R1b merge before your PR" — i.e. DS-104 must push a branch it intends to merge, while frozen.
- **Why it matters:** as shown, DS-104 either violates the freeze or never delivers the migration script that R2 depends on. F26 only exempts the audit PRs #268–#275.
- **Fix:** add to D-FREEZE (and to F26): "**DS-104's `feat/rename-vl-migrate` is exempt until its PR opens** — its merge is gated on the R1b merge sha and it touches only new files. The audit report PRs #268–#275 remain exempt (report-only)." Alternatively move D3's push window to after the R1b merge; pick one and make both texts say it.

---

## P2-1 — F7 incomplete: R1b item 6 and D2's mutant still contradict the same-session exemption

- **Where:** plan v2 "Phase R1b item 6" still reads: "The new lock script REFUSES while the old dir exists ('old lock present — release it with the old script first')"; D2's mutant list says "run vl-lock.sh with the old lock dir present (must refuse)".
- **Why it matters:** under F7 a same-session acquire MUST succeed (it is the hand-over's first half) — a lane implementing item 6 literally re-creates the unlocked window, and the CTO's mutant would then be asserted against its own fix. The lock's identity is the session name in the meta [A: `deploy/nofx-lock.sh` header comments — "identity is the SESSION NAME"]; the exemption is implementable.
- **Fix:** replace the item-6 sentence with: "`vl-lock.sh` REFUSES while `~/nofx-main.lock.d` exists **unless the old lock's meta names the SAME session** (then acquire; the migration releases the old lock immediately after)." Mutant becomes: "old lock dir present, held by a **different** session → must refuse."

## P2-2 — F7 missing the fail-closed rule for a meta-less old lock

- **Where:** F7 text, and D3 step 0 ("the OLD lock held by the session named in `--session`").
- **Why:** the canon's own `check` rc-4 case is an acquire that died before writing meta (an INCOMPLETE lock) — and `clear-incomplete` is the only verb for it. A same-session exemption that skips the meta read would let a migration session treat an incomplete old dir as "held by me".
- **Fix:** add to F7: "the exemption requires the old dir's meta to EXIST and to name exactly `--session`; missing, unreadable, or mismatched meta → REFUSE (the operator clears an incomplete old lock with the old script's `clear-incomplete` first)."

## P3-1 — F25 names the wrong file for the persona doc

- **Where:** F25: "Tracked canon docs updated in R1b: `docs/superpowers/CLAUDE-canon.md` (17 refs …), **AGENTS.md**, `AUDIT-CHECKLIST.md:106`".
- **Evidence:** `git ls-files` — the tracked persona doc is **`agents.md`** (lowercase; its line 1 and 5 say "NOFXi" [A]); the root `AGENTS.md` is **gitignored** [A: not in `git ls-files`; only `web/CLAUDE.md` is tracked under that pattern]. `docs/superpowers/CLAUDE-canon.md` = 17 nofx refs ✓ and `AUDIT-CHECKLIST.md:106` is the worktree example ✓ [A] — the rest of F25 checks out.
- **Fix:** F25 becomes: "…`agents.md` (the persona spec — its NOFXi lines become VL; note it is NOT on D2 item 9's docs list — add it), `docs/superpowers/CLAUDE-canon.md` (17 refs, worktree example), `AUDIT-CHECKLIST.md:106`; the root `AGENTS.md`/`CLAUDE.md` are gitignored owner files (CTO drafts, owner applies — already covered by the Owner-steps)."

## P3-2 — `Codex-canon.md` is not on dev yet; say so or it resurfaces

- **Where:** F25's canon list (completeness).
- **Evidence:** `docs/superpowers/Codex-canon.md` does not exist at origin/dev [A: `ls docs/superpowers/` — only `CLAUDE-canon.md`]; the repo-root AGENTS.md header says the Codex canon lives on branch `fix/tree-guard` and "becomes the two-line pointer" only when it lands on dev.
- **Fix:** append to F25: "`docs/superpowers/Codex-canon.md` is not on dev yet; when it lands, it needs the same nofx sweep (or the census guard must allow-list it by name)."

## P3-3 — D-FREEZE timing leaves the R1a window unprotected (accepted, but say it)

- **Where:** timeline: freeze "sent with D2".
- **Why:** dev merges during R1a (D1, ~2 h) force DS-103 rebases — cheaper than re-running R1b, so this is a deliberate choice; but D-FREEZE's text does not say the freeze starts at D2, so a lane could read it as "frozen from now".
- **Fix:** D-FREEZE first line: "This freeze starts when D2 is dispatched (the R1b window); during D1/R1a, merging to dev costs only a rebase and is still discouraged."

## Fold verdicts (job (a), explicit)

- **F5** CORRECT [A: unit template verified] · incomplete only via P1-1-5.
- **F7** CORRECT [A lock meta + B flow] · incomplete via P2-1/P2-2.
- **F25** CORRECT in substance [A counts verified] · wrong file name via P3-1/P3-2.
- **F26** CORRECT direction · not integrated into the shown D-FREEZE text (the exemption sentences are absent there) and conflicts with D3 via P1-2.

## Outside-the-repo sweep (deeper than round 1, job (b))

- `crontab -l`: **zero** nofx entries [A] — round 1 found none; still none.
- `~/.agent-bridge` + lane hooks: keyed by path strings under `~/.agent-bridge`, not `~/nofx` [A] — survive the move; no new finding (round-1 P3-1 stands).
- Lock hand-over: covered by P2-1/P2-2 above; additionally, `deploy/nofx-lock.sh` exists and is executable [A], so the "old script" side of the hand-over is real, not a placeholder.

## What this audit did NOT do

No code change, no unit stop/start, no DB access, no lock acquire, no GitHub action, no merge. Only this report file on a dedicated branch.
