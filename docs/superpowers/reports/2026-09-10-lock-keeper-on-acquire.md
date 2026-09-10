# Lock keeper on acquire — Section G report

**Wave:** lock keeper on acquire (relayed by lane nofx-07, GO'd by the owner)
**Branch:** `fix/lock-keeper-on-acquire`
**Landed on dev:** `a4c72ff7` (keeper), then `88d40920` (the stderr defect below)
**Suites at the merged head:** lock 82 pass / 0 fail · claim 17 / 0 · Go build OK
**Scope:** `deploy/nofx-lock.sh`, `deploy/nofx-lock-test.sh`, `AUDIT-CHECKLIST.md`.
No Go source, no binary, no boot. SIM untouched.

---

## The defect this wave existed to fix

`deploy/nofx-lock.sh acquire` printed, at line 84:

> heartbeat every 120s

and started no heartbeat. The sentence described a thing the tool did not do.
Liveness in the class-70 lock model is a heartbeat the holder rewrites; a holder
who believed the printed line and did not beat by hand went STALE inside five
minutes while working. Every lane had independently discovered this and was
running a hand-rolled `while true; do ... heartbeat; sleep N; done` — including
this one, for the duration of this wave.

## What shipped

- `acquire` spawns the keeper itself, `setsid`-detached, and records its process
  **group** in `keeper.pid`.
- `keeper.pid` is a **stop handle only**. Nothing reads it as liveness; class
  70's four pins stay green. Liveness remains the heartbeat, and only the
  heartbeat.
- The keeper beats until the lock's **declared expiry**, then stops and records
  why in `keeper.ended`. It never auto-extends. A holder who needs longer
  re-acquires or extends explicitly.
- `release` ends the whole process **group** and WAITS for it (bounded, ~5s)
  before `rm -rf` — never removing the directory while a writer could be inside.
- `status` reports `auto-beat: on (until <expiry>) / ENDED — <why> / off` read
  **from the file**, never by probing a process (class 88).

## Findings

### 1. The expiry was written, printed, and never compared — CLASS 101

`expiry` was set at `acquire` and rendered on every `status` line, ALIVE and
STALE alike. `cmd_heartbeat` refused on exactly two conditions: no lock
directory, and not the holder. **Past its expiry a lock beat happily, forever.**
Lanes quoted the field in prose as though it bound something.

My first design bounded the keeper's own loop. A peer lane (nofx-ed) read the
shipped file and named why that was insufficient: it constrains the keeper THIS
SCRIPT starts, and every lane had a hand-rolled beater precisely because the tool
had never started one. The bound moved to `cmd_heartbeat` — the single place a
heartbeat can be written — so it holds for every writer and terminates the keeper
for free, since the loop breaks on the same non-zero rc.

Verified live: `REFUSED — past the declared expiry (…). No writer extends a
window: re-acquire, or extend explicitly.`

**Caveat, stated plainly.** A lock created before this change carries no
`expiry_epoch` field, and is left **unbounded** rather than retroactively
expired. This is deliberate. Parsing the ISO `expiry` as a fallback would
retroactively bound every lock currently held — potentially refusing the next
heartbeat of a lane mid-cutover, sending its lock stale and inviting a reclaim
under live work. That is a worse failure than the one being fixed. **The guard
binds locks acquired from now on, not the ones lanes are holding right now.**

### 2. The fix rebuilt the lock's own failure (2) one layer down — CLASS 102

Found by an adversarial pass told to attack the fix. `_stop_keeper` killed only
the keeper LOOP. The loop runs `bash "$self" heartbeat` as a foreground CHILD, so
killing the parent ORPHANED that child — and `_write_meta` mv's into
`$LOCK_DIR/meta` by absolute path with no identity check.

An orphan that had passed `_require_holder` while A held the lock could land A's
meta into the lock B created at the same path seconds later. B held the tree
while the lock said A: **B could not release its own lock, and A — holding
nothing — could**, freeing the tree under a live cutover. The atomic `mv` is what
made it silent; the wrong content landed whole, never torn.

That is failure (2) from this file's own header — *a live pid silently
overwritten by a second writer* — rebuilt by the wave whose purpose was to make
holding safer. Fixed with a process-group kill plus a bounded wait.

The pin took four attempts and **the first three passed with the defect fully
present**: it asserted the keeper FILE was gone (`rm -rf` does that anyway), then
the PROCESS but on a beat so short an unstopped keeper exited by itself first
(measuring OS timing, not the code), then SAMPLED the race eight times against a
natural rate near one in sixty. It bites only with a conditional timing shim that
parks exactly one write. A race pin that does not widen its window is testing
luck.

### 3. The keeper's `/proc` read never ran — CLASS 103

Found **after the keeper was already on dev**, by running the shipped verb once
by hand and reading its stderr before announcing it to other lanes.

`_spawn_keeper` reads the process GROUP from `/proc/PID/stat` because `release`
must kill loop and child together, and the comment two lines above says plainly
that `$!` is not reliably the group leader. The line was written with the
`'"'"'` form — correct only INSIDE an already single-quoted string. It was not
inside one. Bash read `{print $5}` in a **double**-quoted region, expanded `$5`
against the function's own empty arguments, and `set -u` aborted the
substitution.

- Every `acquire` printed `line 143: $5: unbound variable` to stderr, while still
  succeeding — noise attached to a working command.
- The `/proc` read never executed once. `pgid` fell through to `$!`, the exact
  value the comment says cannot be trusted. The defensive path was dead code and
  the hazard it defended against was live.

**75 green tests saw nothing** because behaviour was correct: `setsid` execs when
not already a session leader, so `pid == pgid` on this platform and the fallback's
answer happened to equal the right one. The suite was measuring the platform, not
the implementation.

Fixed, and the parse is now counted after `comm` rather than from the start of
the line — `/proc/PID/stat` is `pid (comm) state ppid pgrp …` and `comm` is the
only field that may contain spaces or parens. New pin: **acquire writes nothing
to stderr**, mutation-tested to fail on the shipped script with that exact
message. It is an stderr assertion because no behavioural assertion could catch
it.

---

## Owed / undeliverable

**The landing ping could not be delivered to the lanes that were owed it.** I
committed to nofx-66: *"You get an explicit ping when it is actually on dev."*
By the time the keeper landed, `nofx-07`, `nofx-6b`, `nofx-ed`, `nofx-75` and
`nofx-66` had all ended — nofx-66's socket
(`/run/user/1000/cc-socks/42034.sock`) is gone, and the roster has turned over
entirely. The debt is recorded here instead of delivered. Anything those lanes
were told to expect from me is in this report and in classes 101–103.

**Owed to nofx-66 specifically, now undeliverable to them:**

- The diffs for lock-script pre-existing defects **1** and **3**. Not written —
  those are the next wave (below).
- Confirmation when the arm-state SQL fragment lands. Not built.

**nofx-66's standing caution, recorded because it is right and outlives them.**
On the arm-state fragment: if the SQL fragment is a separate `[]string` or a
hand-written `IN ('filled',…)` constant sitting beside the switch, *the class has
been moved, not retired* — two hand-typed lists in one file diverge exactly as
readily as one in Go and one in SQL. The version that closes it has ONE list as
the single source: a package-level `terminalArmStates` slice, the switch ranging
over it, and the SQL generated from it. Naming is clear at `4c7f1def`:
`TerminalArmStates|terminalArmStatesSQL|ArmTerminalSQL|TerminalArmStateSQL` have
zero hits in `*.go`, and `isTerminalArmState` remains the sole predicate at
`trader/one_contract.go:285`.

**An owner-side edit is needed and I did not make it.** `CLAUDE.md` is
**untracked** — a local file, not in the repo — so this wave could not ship a
change to it. Its MAIN-TREE LOCK LAW block still instructs every lane:

> `deploy/nofx-lock.sh heartbeat <session>   # beat every ~2 min as you work`

That is now wrong: `acquire` starts the keeper, and a lane running its own beater
in addition is a second writer of exactly the kind class 102 is about. The line
should say the keeper is automatic, that it stops at the declared expiry, and
that a lane needing longer re-acquires or extends explicitly.

## Next wave — the three pre-existing lock defects (Section C pinned by the owner)

1. `release` discards `rm`'s exit status.
2. A lock dir with no `meta` is terminal.
3. `status`/`check` read a half-built lock as STALE with an empty holder inside
   the acquire window — 300/300 in that window — and `check` is the tree guard's
   unattended interface.

Plus, folded in from peers: cutover-gate leg 4 counting signal-less arms against
the broker's book (owner ruled the fix on 09-06, still unshipped), and the
arm-state SQL fragment above.
