#!/usr/bin/env bash
# nofx-lock — the main-tree lock, atomic and heartbeat-based.
#
# WHY THIS SHAPE (owner ruling 2026-09-03). The old lock was ONE FLAT FILE
# written with `>`, carrying a PID, read with `kill -0`. It failed in THREE
# directions in a single day, and not one was caught by the file — every one was
# caught by a peer asking:
#
#  1. DEAD PID, LIVE OWNER. Agents recorded the shell's own id, but every tool
#     call is a fresh shell, so it was dead within a second while its owner
#     worked on. A peer found an unexpired lock that did not resolve and
#     correctly did NOT clear it.
#  2. LIVE PID, SILENTLY REPLACED. `>` truncates: a second acquirer clobbered an
#     active cutover's lock with no error and no trace. A write that clobbers is
#     not a lock.
#  3. STALE PID AFTER RESUME. A resumed session wrote a lock naming its own
#     former, now-dead process.
#
# The fault under all three: a process id answers "does some process exist",
# which was never the question. The question is "is the OWNER still working",
# and only the owner can answer it.
#
# So: acquisition is an ATOMIC CREATE (mkdir succeeds for exactly one caller and
# fails for every other — no read-then-write anywhere, so (2) is not detected
# but UNREPRESENTABLE), identity is the SESSION NAME rather than a process id,
# and liveness is a HEARTBEAT the holder rewrites every 2 minutes. There is no
# pid field, so (1) and (3) have nothing to record wrongly.
#
# A heartbeat older than 5 minutes is reported STALE, NEVER DEAD. Corroboration
# (HEAD moving, a build in flight, the named session answering) remains required
# before any clearing, and this script never clears a lock itself at any age.
#
# Usage:
#   nofx-lock acquire <session> <task> [minutes]   # atomic; refuses if held
#   nofx-lock heartbeat <session>                  # holder rewrites; every 2m
#   nofx-lock status                               # free | held | STALE + age
#   nofx-lock check                                # rc 0 free · 1 held · 2 stale
#   nofx-lock with-heartbeat <session> -- <cmd>    # beats for <cmd>'s lifetime
#   nofx-lock reclaim <new> <stale> "<corroboration>"  # ONLY on a stale heartbeat
#   nofx-lock release <session>                    # only the holder may release
set -uo pipefail

LOCK_DIR="${NOFX_LOCK_DIR:-$HOME/nofx-main.lock.d}"
LEGACY_LOCK="${NOFX_LEGACY_LOCK:-$HOME/nofx-main.lock}"
HEARTBEAT_STALE_SECONDS="${NOFX_LOCK_STALE_SECONDS:-300}"   # 5 min
HEARTBEAT_EVERY_SECONDS=120                                  # 2 min

_now()      { date -Is; }
_epoch()    { date +%s; }
_meta()     { cat "$LOCK_DIR/meta" 2>/dev/null; }
_field()    { _meta | grep -m1 "^$1=" | cut -d= -f2-; }

# The meta file is replaced, never edited in place: a reader must never catch a
# half-written heartbeat. mktemp rather than the shell's own id — this file does
# not name processes, and a pin enforces that.
_write_meta() {
  local tmp; tmp="$(mktemp "$LOCK_DIR/.meta.XXXXXX")" || return 1
  cat > "$tmp" && mv -f "$tmp" "$LOCK_DIR/meta"
}

_require_holder() { # _require_holder <session>
  local want="$1" have; have="$(_field session)"
  if [ "$want" != "$have" ]; then
    echo "REFUSED — '$want' is not the holder ('$have', task: $(_field task))."
    return 1
  fi
}

cmd_acquire() {
  local session="${1:?session required}" task="${2:?task required}" mins="${3:-45}"
  # THE ATOMIC STEP. mkdir either creates the directory or fails; it never
  # overwrites. Exactly one caller can win, whatever the interleaving.
  if ! mkdir "$LOCK_DIR" 2>/dev/null; then
    echo "REFUSED — lock already held:"; cmd_status; return 1
  fi
  _write_meta <<META
session=$session
task=$task
acquired=$(_now)
expiry=$(date -Is -d "+$mins minutes")
heartbeat=$(_now)
heartbeat_epoch=$(_epoch)
META
  echo "ACQUIRED by $session — heartbeat every ${HEARTBEAT_EVERY_SECONDS}s, stale after ${HEARTBEAT_STALE_SECONDS}s"
}

# The beat is owner-scoped. An unauthenticated beat would let any lane keep a
# stranger's abandoned lock looking alive, which is failure (1) rebuilt.
cmd_heartbeat() {
  local session="${1:?session required}"
  [ -d "$LOCK_DIR" ] || { echo "REFUSED — no lock to beat."; return 1; }
  _require_holder "$session" || return 1
  { _meta | grep -v '^heartbeat'; echo "heartbeat=$(_now)"; echo "heartbeat_epoch=$(_epoch)"; } | _write_meta
  echo "heartbeat $(_now)"
}

# The history lives in the lock dir and dies with the lock, so it is printed
# wherever the lock is inspected and once more at release — a succession chain
# that vanishes silently would defeat the point of recording it.
_show_history() {
  [ -s "$LOCK_DIR/history" ] || return 0
  echo "  history:"; sed 's/^/    /' "$LOCK_DIR/history"
}

_age() { local hb; hb="$(_field heartbeat_epoch)"; echo $(( $(_epoch) - ${hb:-0} )); }

cmd_status() {
  if [ -f "$LEGACY_LOCK" ]; then
    echo "LEGACY lock file present at $LEGACY_LOCK — a lane is still on the old shape:"
    sed 's/^/    /' "$LEGACY_LOCK"
  fi
  if [ ! -d "$LOCK_DIR" ]; then
    [ -f "$LEGACY_LOCK" ] && return 0
    echo "free"; return 0
  fi
  local age session task; age="$(_age)"; session="$(_field session)"; task="$(_field task)"
  if [ "$age" -gt "$HEARTBEAT_STALE_SECONDS" ]; then
    # STALE, NOT DEAD. The distinction is the whole point: a stale heartbeat
    # says the holder has not checked in, not that it has stopped. Corroborate
    # before clearing — HEAD moving, a build in flight, the session answering.
    echo "STALE — held by '$session' (task: $task), heartbeat ${age}s old (> ${HEARTBEAT_STALE_SECONDS}s), expiry $(_field expiry)."
    echo "  DO NOT CLEAR ON THIS ALONE. Corroborate first: is HEAD moving? is a build running?"
    echo "  does '$session' answer? Clear only with a note naming what you checked."
    echo "  To take it over on the record: nofx-lock reclaim <you> '$session' \"<what you checked>\""
    _show_history
    return 2
  fi
  echo "held by '$session' (task: $task), heartbeat ${age}s old — ALIVE, expiry $(_field expiry)"
  _show_history
}

# For scripts (the tree guard): 0 free · 1 held-fresh · 2 held-stale.
cmd_check() {
  [ -d "$LOCK_DIR" ] || { echo free; return 0; }
  if [ "$(_age)" -gt "$HEARTBEAT_STALE_SECONDS" ]; then echo stale; return 2; fi
  echo held; return 1
}

# Beats while <cmd> runs and STOPS when it ends. The beater's lifetime is
# exactly the work's lifetime; a beater that outlived its job would report a
# gone owner as live, which is failure (1) in a new costume.
cmd_with_heartbeat() {
  local session="${1:?session required}"; shift
  [ "${1:-}" = "--" ] && shift
  [ -d "$LOCK_DIR" ] || { echo "REFUSED — no lock to beat."; return 1; }
  _require_holder "$session" || return 1
  cmd_heartbeat "$session" >/dev/null
  ( while sleep "$HEARTBEAT_EVERY_SECONDS"; do
      [ -d "$LOCK_DIR" ] || exit 0
      [ "$(_field session)" = "$session" ] || exit 0
      cmd_heartbeat "$session" >/dev/null
    done ) &
  local beater=$!
  "$@"; local rc=$?
  kill "$beater" 2>/dev/null; wait "$beater" 2>/dev/null
  return $rc
}

# reclaim — succession, on the record.
#
# The failure this exists for is INVISIBLE SUCCESSION: a lock whose owner is
# gone, taken over by someone else with nothing written down. So every reclaim
# APPENDS to the lock's history — who took it, from whom, when, and what they
# checked — and an unauditable reclaim is worse than none.
#
# It is refused while the heartbeat is FRESH, without exception. A reclaim that
# can take a live lock is replacement with better manners, which is the very
# failure the atomic create removed.
#
# rc 3 is deliberately distinct from acquire's rc 0: a script can tell "took a
# free lock" from "inherited an abandoned one", and a lane that would rather not
# inherit can refuse.
cmd_reclaim() { # reclaim <new_session> <stale_session> <corroboration...>
  local new="${1:?new session required}" stale="${2:-}" ; shift 2 2>/dev/null || true
  local why="${*:-}"
  [ -d "$LOCK_DIR" ] || { echo "REFUSED — no lock to reclaim (use acquire)."; return 1; }
  local holder age; holder="$(_field session)"; age="$(_age)"
  if [ -z "$stale" ] || [ "$stale" != "$holder" ]; then
    echo "REFUSED — name the session you are taking over. The holder is '$holder', you named '${stale:-<nothing>}'."
    return 1
  fi
  if [ "$age" -le "$HEARTBEAT_STALE_SECONDS" ]; then
    echo "REFUSED — '$holder' has a FRESH heartbeat (${age}s old, stale is > ${HEARTBEAT_STALE_SECONDS}s)."
    echo "  A live holder is never reclaimable. Wait, or ask '$holder' to release."
    return 1
  fi
  if [ -z "$why" ]; then
    echo "REFUSED — state the corroboration you checked (HEAD not moving, no build in flight, session not answering)."
    echo "  usage: nofx-lock reclaim <new> <stale> \"<what you checked>\""
    return 1
  fi
  printf '%s reclaim: %s took over from %s (heartbeat was %ss old) — corroboration: %s\n' \
    "$(_now)" "$new" "$stale" "$age" "$why" >> "$LOCK_DIR/history"
  { _meta | grep -vE '^(session|heartbeat)'
    echo "session=$new"
    echo "heartbeat=$(_now)"
    echo "heartbeat_epoch=$(_epoch)"
  } | _write_meta
  echo "RECLAIMED by $new from $stale — logged to the lock history."
  return 3
}

cmd_release() {
  local session="${1:?session required}"
  [ -d "$LOCK_DIR" ] || { echo "no lock"; return 0; }
  _require_holder "$session" || return 1
  _show_history
  rm -rf "$LOCK_DIR"; echo "released by $session"
}

case "${1:-status}" in
  acquire)        shift; cmd_acquire "$@" ;;
  heartbeat)      shift; cmd_heartbeat "$@" ;;
  status)         cmd_status ;;
  check)          cmd_check ;;
  with-heartbeat) shift; cmd_with_heartbeat "$@" ;;
  reclaim)        shift; cmd_reclaim "$@" ;;
  release)        shift; cmd_release "$@" ;;
  *) echo "usage: nofx-lock {acquire <session> <task> [mins]|heartbeat <session>|status|check|with-heartbeat <session> -- <cmd>|reclaim <new> <stale> \"<corroboration>\"|release <session>}"; exit 64 ;;
esac
