# TREE-GUARD — the main-tree watcher

**Branch** `fix/tree-guard` (off `origin/dev`)
**Status** BUILT AND GREEN. No Go change, no binary, **no cutover**.
Install is a `systemd --user` unit; the guard itself never writes to the tree.

---

## 1 — THE FOUR CHECKS

Live output against the real deploy tree, `--once`, read-only:

```
tree-guard PASS porcelain: tree clean
tree-guard PASS canary: 12 shipped symbol(s) present
tree-guard PASS release: RELEASE=89673ccc5984eef8ad8168edda24984fd1da50cb == running == HEAD:deploy/RELEASE
tree-guard PASS staleness: HEAD == origin/dev (26594412)
tree-guard: all checks clear
```

| # | check | ALARMS when | why it is not redundant |
|---|---|---|---|
| 1 | porcelain | dirty **and no live lock holder** | the 08:46 signature, caught in one tick |
| 2 | shipped-symbol canary | any listed symbol absent | catches the same loss **after a commit**, when check 1 calls the tree clean |
| 3 | RELEASE vs running vs `HEAD:deploy/RELEASE` | any disagreement | the A19 class — a marker carrying a stale RELEASE while another rev runs |
| 4 | staleness | HEAD **ahead** of `origin/dev`, or behind by > 20 | the unpushed-marker class: three deployed waves lived only on disk for hours |

**The lock earns a second job.** A cutover legitimately dirties the tree (A19 writes
RELEASE before the kill), so checks 1 and 3 downgrade to INFO — but only when the lock
names a pid that is **alive** *and* its heartbeat is fresh. A dead pid buys no silence.
That is the same corroboration rule that stopped a live holder's lock being cleared on
09-03, applied in the other direction: neither a dead process nor a stale heartbeat may
speak for a live one.

INFO still lists the files. Suppression that hides the evidence is not suppression, it is
blindness with a nicer name.

---

## 2 — FILE:LINES

| What | Where |
|---|---|
| The guard (191 lines) | `deploy/nofx-tree-guard.sh` |
| Lock liveness + heartbeat | `:42 THE LOCK'S SECOND JOB` |
| Check 1 porcelain | `:74` · Check 2 canary `:90` · Check 3 release `:116` · Check 4 staleness `:148` |
| State file (outside the tree) | `:173` |
| The canary's vocabulary | `deploy/tree-guard-symbols.txt` (12 symbols) |
| Units | `deploy/systemd-user/nofx-tree-guard.{service,timer}` |
| Installer | `deploy/install-tree-guard.sh` |
| Tests — **20 cases** | `deploy/tree_guard_test.go` |
| Guide | `web/src/guide/content/status.ts` |
| Checklist | class **71** |

---

## 3 — E1 RED → GREEN

**RED**, before the script existed:

```
--- FAIL: TestE5ScriptDeclaresItsReadOnlyContract
    tree_guard_test.go:280: read guard: open nofx-tree-guard.sh: no such file or directory
FAIL	nofx/deploy
```

**GREEN**, 16 cases, including the three that carry the wave:

- `TestE1DirtyTreeWithNoLockAlarmsAndNamesTheFile` — and asserts the file is **named**;
  an alarm you have to investigate to understand is half an alarm.
- `TestE1DirtyTreeUnderADeadPidLockStillAlarms` — a dead pid must not silence the guard.
- `TestE2CommittedRevertAlarmsWhileTheTreeIsClean` — the incident-specific pin: porcelain
  PASSES, the canary ALARMS, and it names `composeArmStop`.

Full suite: Go `./...` **0 failures** · vitest **42 files / 336 tests** · `tsc` clean.

**The tests are Go, not shell, on purpose.** A shell test nobody runs is not a test;
`go test ./...` is what every lane runs before every merge.

---

## 4 — THE CONTRACT, ENFORCED RATHER THAN PROMISED

A31 is the wave. `TestE5ScriptContainsNoWritingGitCommand` greps the script for
`checkout|restore|reset|stash|clean|commit|push|pull|merge` and fails on any of them, and
`TestE4GuardNeverFetches` adds `fetch` — a fetch writes into the `.git` of the tree the
guard is supposed only to observe, which is why check 4 uses `ls-remote`.

A repairing guard would be a second unaccountable writer to the deploy tree. That is the
disease, not the cure, and the contract is in the script's own header so it cannot rot
into a spec nobody opens.

---

## 5 — THE LOCK CHANGED UNDERNEATH THIS WAVE (A23)

**I built against a superseded spec and shipped a guard that would have false-alarmed on
every cutover.** Found only because a peer asked an unrelated provenance question.

At **21:48:36** another lane landed `ec2dd8f7`, which replaced the lock model — from
`~/nofx-main.lock`, a file with a pid, liveness by `kill -0`, to `~/nofx-main.lock.d`, an
atomic directory keyed by SESSION with a heartbeat and **no pid at all** — and edited
`2026-09-02-tree-guard-spec.md` to say so. I created my worktree and read that spec at
**~21:54, six minutes later**, but from a base commit that predates the edit. So I read
the old version and implemented the old model.

The failure this would have produced is the worst shape available: during a cutover under
the new lock there is no legacy file, so the guard would find "no live holder", see the
legitimately dirty tree, and **ALARM — at exactly the moment it is supposed to be
trusted.** It would have kept running and kept printing the whole time. A guard that cries
wolf on every deploy is worse than no guard, because the next real alarm is the one you
scroll past.

**Fixed:** the directory is authoritative (heartbeat age < 300 s), and the legacy file is
**surfaced but never honoured for liveness** — honouring it would reintroduce the exact
`kill -0` test the new lock exists to remove. **STALE, never DEAD**: the guard reports a
heartbeat age and refuses to declare a session dead, because a pid died on 09-03 while its
holder kept working.

One test was superseded and migrated with its reason rather than weakened —
`TestLegacyLockFileNoLongerConfersLiveness` now asserts the opposite of what it originally
asserted, and says why in the test body.

**The lesson is not "read the spec more carefully".** It is that a spec on dev is a moving
artifact, and a worktree cut from an older base silently freezes it. Checking `git log`
on the spec file before building against it costs one command.

## 6 — TWO PLACES I DEPARTED FROM THE DISPATCH, AND WHY

**5.1 — The state file lives OUTSIDE the tree.** Section B says "a state file under
`data/`"; the spec says `~/nofx-backups/tree-guard/state`. I followed the spec.
`data/` **is** gitignored (`.gitignore:40`), so a write there would not dirty the tree —
and that is precisely the objection: the guard would be writing into the tree it guards,
**invisibly to its own porcelain check**. A guard whose own writes it cannot see is the
wrong shape. The clock guard writes to `data/` and that is fine, because it does not guard
the tree.

**5.2 — Check 3 has a third comparison.** The dispatch names RELEASE, the running binary
and `/api/health`. I compare RELEASE, the running binary and **`HEAD:deploy/RELEASE`**, and
skip `/api/health`: the health endpoint reports the same `vcs.revision` the binary already
gives us via `/proc/<pid>/exe`, so asking it adds a second path to one fact and a network
dependency to a guard that should have none. `HEAD:deploy/RELEASE` is a genuinely different
fact — it catches a marker committed from the wrong tree, which happened twice on 09-02.

---

## 7 — A15: WHAT IS STILL WRONG AFTER THIS

**The hole is not closed.** An editor can overwrite the deploy tree at any moment. The
worktree law and the lock govern agents; an editor is not an agent, holds no lock, and its
write is indistinguishable from a legitimate one until someone reads the content. This
guard shortens the discovery window from **3h20m to 60s**. It does not prevent anything.

The only real fix is to stop opening `/home/hoang/nofx` in an editor and open a worktree
instead. That is an owner habit, and no file in this repo substitutes for it. The Guide
says so plainly rather than implying the problem is handled.

**Also still true:** the guard is only as good as its symbol list. A wave that ships
something whose loss would be invisible and does not add a line to
`deploy/tree-guard-symbols.txt` is unguarded, and nothing will tell you that.

---

## 8 — INSTALL + PROOF

Install writes only to `~/.config/systemd/user` and chmods the script — it is not a tree
write, which is why it needs no cutover:

```
bash deploy/install-tree-guard.sh
```

**PROOF owed:** the first real ALARM, whenever it fires, quoted with its cause. Four PASS
lines on a healthy tree prove the guard runs; only an alarm proves it *catches*. Until one
fires, this wave is installed and unproven — and the honest way to read the current output
is "nothing is wrong right now", not "the guard works".
