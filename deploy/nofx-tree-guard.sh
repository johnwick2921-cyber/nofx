#!/usr/bin/env bash
# nofx-tree-guard.sh — the main-tree watcher.
#
# CONTRACT: OBSERVE AND ALARM. NEVER REPAIR.
# This script runs no git command that writes: no checkout, restore, reset,
# stash, clean, commit, push, pull or merge. An automatic repair here would be a
# SECOND actor mutating the deploy tree, which is the disease this guard exists
# to detect, not the cure. It reports; a human decides. A test asserts this by
# grepping the script itself, so the contract cannot rot into a comment.
#
# WHY IT EXISTS. On 2026-09-02 at 08:46:33-34 an editor Save-All wrote six stale
# buffers over /home/hoang/nofx, deleting 596 lines of shipped safety code across
# four waves. No agent did it and no git command did it, so neither the worktree
# law nor ~/nofx-main.lock could have prevented it — both govern AGENTS, and an
# editor is not an agent. It went unnoticed for 3h20m and was found by accident.
# This guard does not close that hole. It shortens the discovery window from
# hours to one minute. The only real fix is to stop opening the deploy tree in an
# editor, and no repo change substitutes for that.
#
#   bash deploy/nofx-tree-guard.sh --once     # human-readable, all four checks
#
# Exit 0 = every check PASS or INFO. Exit 2 = at least one ALARM.
set -uo pipefail

# ── RESOLVED VALUES (A11): every path is overridable and every default is
# stated here rather than assumed by a caller.
TREE="${TREE_GUARD_TREE:-/home/hoang/nofx}"
# THE LOCK MOVED under this wave (ec2dd8f7, 2026-09-03 21:48): it is now an
# ATOMIC DIRECTORY keyed by SESSION with a heartbeat, and it records NO PID —
# kill -0 was the wrong liveness test. Both are read here: the directory is
# authoritative, and a leftover legacy file is SURFACED rather than honoured,
# because a stale pid-file lying next to the real lock is the transition's
# actual hazard.
LOCK_DIR="${TREE_GUARD_LOCK_DIR:-$HOME/nofx-main.lock.d}"
LOCK="${TREE_GUARD_LOCK:-/home/hoang/nofx-main.lock}"
LOCK_STALE_S="${TREE_GUARD_LOCK_STALE_S:-300}"
STATE="${TREE_GUARD_STATE:-$HOME/nofx-backups/tree-guard/state}"
SYMBOLS="${TREE_GUARD_SYMBOLS:-$TREE/deploy/tree-guard-symbols.txt}"
HEARTBEAT_MAX_S="${TREE_GUARD_HEARTBEAT_MAX_S:-900}"
BEHIND_MAX="${TREE_GUARD_BEHIND_MAX:-20}"

alarms=0
lines=()

say() { lines+=("$1"); echo "$1"; }
alarm() { alarms=$((alarms + 1)); say "🚨 tree-guard ALARM $1"; }
pass()  { say "tree-guard PASS $1"; }
info()  { say "tree-guard INFO $1"; }

# ── THE LOCK'S SECOND JOB ────────────────────────────────────────────────────
# A cutover legitimately dirties the tree (A19 writes deploy/RELEASE before the
# kill). The lock is the DECLARATION that dirt is intentional — but only while
# its heartbeat is FRESH. A stale heartbeat buys no silence.
#
# STALE, NEVER DEAD. This guard does not get to declare a session dead: a pid
# died on 09-03 while its holder kept working, and a peer nearly cleared a live
# lock on that reading. Age is reported; a human corroborates.
lock_live=0
lock_desc=""
legacy_note=""

if [ -f "$LOCK_DIR/meta" ]; then
  l_session="$(grep -m1 '^session=' "$LOCK_DIR/meta" 2>/dev/null | cut -d= -f2- || true)"
  l_task="$(grep -m1 '^task=' "$LOCK_DIR/meta" 2>/dev/null | cut -d= -f2- || true)"
  l_hb="$(grep -m1 '^heartbeat_epoch=' "$LOCK_DIR/meta" 2>/dev/null | cut -d= -f2- || true)"
  if [ -n "${l_hb:-}" ]; then
    age=$(( $(date +%s) - l_hb ))
    if [ "$age" -le "$LOCK_STALE_S" ]; then
      lock_live=1
      lock_desc="session '${l_session:-?}' task=${l_task:-?} heartbeat ${age}s old"
    else
      lock_desc="session '${l_session:-?}' task=${l_task:-?} heartbeat ${age}s old — STALE (> ${LOCK_STALE_S}s); STALE is not DEAD, corroborate before clearing"
    fi
  else
    lock_desc="lock dir present but no heartbeat_epoch — unreadable, treated as NOT live"
  fi
fi

# The legacy pid-file: surfaced either way. Honouring it would reintroduce the
# kill -0 liveness test the new lock exists to remove.
if [ -f "$LOCK" ]; then
  legacy_note=" · NOTE: a legacy lock FILE also exists at $LOCK (the lock moved to $LOCK_DIR on 2026-09-03) — it is NOT honoured for liveness; remove it once its owner is corroborated"
fi

# ── CHECK 1 — PORCELAIN ──────────────────────────────────────────────────────
dirty="$(git -C "$TREE" status --porcelain 2>/dev/null)"
if [ -z "$dirty" ]; then
  pass "porcelain: tree clean$legacy_note"
else
  n=$(printf '%s\n' "$dirty" | wc -l | tr -d ' ')
  files="$(printf '%s\n' "$dirty" | awk '{print $NF}' | paste -sd' ' - | cut -c1-400)"
  if [ "$lock_live" = "1" ]; then
    # Expected-dirty suppression — INFO, never silence: the files are still
    # named, so a cutover that dirties something unexpected is still readable.
    info "porcelain: $n path(s) dirty UNDER A LIVE LOCK ($lock_desc) — expected during a cutover · $files$legacy_note"
  else
    alarm "porcelain: $n path(s) dirty with NO live lock holder (${lock_desc:-no lock}) — this is the 2026-09-02 08:46 signature · $files$legacy_note"
  fi
fi

# ── CHECK 2 — SHIPPED-SYMBOL CANARY ──────────────────────────────────────────
# The check that would have caught 08:46 even if someone had COMMITTED it, which
# is when check 1 would call the tree clean.
if [ ! -r "$SYMBOLS" ]; then
  # An absent list is not "nothing to check" — it is a canary that cannot sing.
  alarm "canary: symbol list unreadable at $SYMBOLS — an empty canary is a check that cannot fail"
else
  missing=()
  checked=0
  while IFS= read -r sym; do
    sym="$(printf '%s' "$sym" | sed 's/#.*//; s/^[[:space:]]*//; s/[[:space:]]*$//')"
    [ -z "$sym" ] && continue
    checked=$((checked + 1))
    if ! grep -rqF --include='*.go' --include='*.cs' --include='*.ts' --include='*.tsx' -- "$sym" "$TREE" 2>/dev/null; then
      missing+=("$sym")
    fi
  done < "$SYMBOLS"
  if [ "$checked" = "0" ]; then
    alarm "canary: symbol list at $SYMBOLS is empty — a check that cannot fail is not a check"
  elif [ ${#missing[@]} -gt 0 ]; then
    alarm "canary: ${#missing[@]} of $checked shipped symbol(s) MISSING from the tree: ${missing[*]} — a committed revert looks clean to porcelain"
  else
    pass "canary: $checked shipped symbol(s) present"
  fi
fi

# ── CHECK 3 — RELEASE VS RUNNING ─────────────────────────────────────────────
if [ "${TREE_GUARD_SKIP_RUNNING:-0}" = "1" ]; then
  info "release: skipped (TREE_GUARD_SKIP_RUNNING=1)"
else
  rel="$(cat "$TREE/deploy/RELEASE" 2>/dev/null | tr -d '[:space:]')"
  head_rel="$(git -C "$TREE" show HEAD:deploy/RELEASE 2>/dev/null | tr -d '[:space:]')"
  running="${TREE_GUARD_RUNNING_REV:-}"
  if [ -z "$running" ]; then
    pid="$(pgrep -x nofx-bin 2>/dev/null | head -1 || true)"
    if [ -n "$pid" ]; then
      running="$(go version -m "/proc/$pid/exe" 2>/dev/null | awk '$1=="build" && $2=="vcs.revision"{print $3}' | head -1)"
      [ -z "$running" ] && running="$(go version -m "/proc/$pid/exe" 2>/dev/null | grep -oE 'vcs\.revision=[0-9a-f]+' | head -1 | cut -d= -f2)"
    fi
  fi
  if [ -z "$rel" ]; then
    alarm "release: deploy/RELEASE is empty or unreadable"
  elif [ -z "$running" ]; then
    # A value the guard cannot know prints as unknown, never as agreement.
    info "release: RELEASE=$rel · running=unknown (no nofx-bin process) — cannot compare"
  elif [ "$rel" != "$running" ]; then
    if [ "$lock_live" = "1" ]; then
      info "release: RELEASE=$rel vs running=$running — MISMATCH under a live lock ($lock_desc), cutover in progress"
    else
      alarm "release: RELEASE=$rel · running=$running · HEAD:deploy/RELEASE=${head_rel:-unknown} — the file and the binary disagree with no cutover in flight"
    fi
  elif [ -n "$head_rel" ] && [ "$head_rel" != "$rel" ]; then
    alarm "release: RELEASE=$rel matches the binary but HEAD:deploy/RELEASE=$head_rel — the marker was never committed from this tree"
  else
    pass "release: RELEASE=$rel == running == HEAD:deploy/RELEASE"
  fi
fi

# ── CHECK 4 — STALENESS ──────────────────────────────────────────────────────
if [ "${TREE_GUARD_SKIP_REMOTE:-0}" = "1" ]; then
  info "staleness: skipped (TREE_GUARD_SKIP_REMOTE=1)"
else
  # READ-ONLY: ls-remote, never fetch. A fetch would write into .git of the tree
  # this guard is supposed to only observe.
  remote_sha="$(git -C "$TREE" ls-remote origin dev 2>/dev/null | awk '{print $1}' | head -1)"
  head_sha="$(git -C "$TREE" rev-parse HEAD 2>/dev/null)"
  if [ -z "$remote_sha" ] || [ -z "$head_sha" ]; then
    info "staleness: cannot resolve HEAD or origin/dev — not compared"
  elif [ "$remote_sha" = "$head_sha" ]; then
    pass "staleness: HEAD == origin/dev ($(printf '%.8s' "$head_sha"))"
  else
    behind="$(git -C "$TREE" rev-list --count "HEAD..$remote_sha" 2>/dev/null || echo '?')"
    ahead="$(git -C "$TREE" rev-list --count "$remote_sha..HEAD" 2>/dev/null || echo '?')"
    if [ "$ahead" != "0" ] && [ "$ahead" != "?" ]; then
      alarm "staleness: HEAD is $ahead commit(s) AHEAD of origin/dev — an unpushed marker (the A19 class); behind by $behind"
    elif [ "$behind" != "?" ] && [ "$behind" -gt "$BEHIND_MAX" ]; then
      alarm "staleness: HEAD is $behind commit(s) behind origin/dev (limit $BEHIND_MAX)"
    else
      info "staleness: HEAD is $behind behind / $ahead ahead of origin/dev"
    fi
  fi
fi

# ── STATE FILE ───────────────────────────────────────────────────────────────
# Deliberately OUTSIDE the tree. A guard that writes into the thing it guards is
# the wrong shape, and data/ is gitignored — so such a write would be invisible
# to this guard's own porcelain check, which is worse than merely untidy.
verdict="PASS"; [ "$alarms" -gt 0 ] && verdict="ALARM"
mkdir -p "$(dirname "$STATE")" 2>/dev/null || true
{
  echo "verdict=$verdict"
  echo "alarms=$alarms"
  echo "checked_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  printf '%s\n' "${lines[@]}"
} > "$STATE.partial" 2>/dev/null && mv "$STATE.partial" "$STATE" 2>/dev/null || true

if [ "$alarms" -gt 0 ]; then
  echo "🚨 tree-guard: $alarms ALARM(s) — the deploy tree does not match what shipped. This guard does NOT repair; a human decides."
  exit 2
fi
echo "tree-guard: all checks clear"
exit 0
