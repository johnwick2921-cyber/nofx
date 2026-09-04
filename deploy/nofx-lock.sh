#!/usr/bin/env bash
# nofx-lock — the main-tree lock, atomic and heartbeat-based.
#
# WHY THIS SHAPE (owner ruling 2026-09-03, from two failures in one day):
#
#  1. The old lock was a plain FILE written with `>`, i.e. read-then-write. Two
#     lanes raced it: nofx-63's lock "replaced mine at 21:43" and nofx-b3 had to
#     re-acquire after. A write that clobbers is not a lock.
#  2. The old lock recorded a PID and liveness was `kill -0`. In a long session
#     the process identity CHANGES: pid 1860416 died while its holder kept
#     working, and a peer nearly cleared a live lock mid-deploy on that reading.
#     A dead PID is not a dead holder.
#
# So: acquisition is an ATOMIC CREATE (mkdir succeeds for exactly one caller,
# fails for every other — no read-then-write anywhere), identity is the SESSION
# NAME rather than a PID, and liveness is a HEARTBEAT the holder rewrites every
# 2 minutes. A heartbeat older than 5 minutes is reported STALE, never DEAD —
# corroboration (HEAD moving, a build in flight, the named session answering)
# remains required before any clearing.
#
# Usage:
#   nofx-lock acquire <session> <task> [minutes]   # atomic; fails if held
#   nofx-lock heartbeat                            # holder rewrites; run every 2m
#   nofx-lock status                               # free | held | STALE + age
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

cmd_acquire() {
  local session="${1:?session required}" task="${2:?task required}" mins="${3:-45}"
  # THE ATOMIC STEP. mkdir either creates the directory or fails; it never
  # overwrites. Exactly one caller can win, whatever the interleaving.
  if ! mkdir "$LOCK_DIR" 2>/dev/null; then
    echo "REFUSED — lock already held:"; cmd_status; return 1
  fi
  { echo "session=$session"
    echo "task=$task"
    echo "acquired=$(_now)"
    echo "expiry=$(date -Is -d "+$mins minutes")"
    echo "heartbeat=$(_now)"
    echo "heartbeat_epoch=$(_epoch)"
  } > "$LOCK_DIR/meta"
  echo "ACQUIRED by $session — heartbeat every ${HEARTBEAT_EVERY_SECONDS}s, stale after ${HEARTBEAT_STALE_SECONDS}s"
}

cmd_heartbeat() {
  [ -d "$LOCK_DIR" ] || { echo "no lock to beat"; return 1; }
  local tmp="$LOCK_DIR/.meta.$$"
  { _meta | grep -v '^heartbeat'; echo "heartbeat=$(_now)"; echo "heartbeat_epoch=$(_epoch)"; } > "$tmp" \
    && mv -f "$tmp" "$LOCK_DIR/meta"
  echo "heartbeat $(_now)"
}

cmd_status() {
  if [ -f "$LEGACY_LOCK" ]; then
    echo "LEGACY lock file present at $LEGACY_LOCK — a lane is still on the old shape:"
    sed 's/^/    /' "$LEGACY_LOCK"
  fi
  if [ ! -d "$LOCK_DIR" ]; then
    [ -f "$LEGACY_LOCK" ] && return 0
    echo "free"; return 0
  fi
  local hb age session task
  hb="$(_field heartbeat_epoch)"; session="$(_field session)"; task="$(_field task)"
  age=$(( $(_epoch) - ${hb:-0} ))
  if [ "$age" -gt "$HEARTBEAT_STALE_SECONDS" ]; then
    # STALE, NOT DEAD. The distinction is the whole point: a stale heartbeat
    # says the holder has not checked in, not that it has stopped. Corroborate
    # before clearing — HEAD moving, a build in flight, the session answering.
    echo "STALE — held by '$session' (task: $task), heartbeat ${age}s old (> ${HEARTBEAT_STALE_SECONDS}s)."
    echo "  DO NOT CLEAR ON THIS ALONE. Corroborate first: is HEAD moving? is a build running?"
    echo "  does '$session' answer? Clear only with a note naming what you checked."
    return 2
  fi
  echo "held by '$session' (task: $task), heartbeat ${age}s old — ALIVE"
}

cmd_release() {
  local session="${1:?session required}"
  [ -d "$LOCK_DIR" ] || { echo "no lock"; return 0; }
  local holder; holder="$(_field session)"
  if [ "$holder" != "$session" ]; then
    echo "REFUSED — '$session' is not the holder ('$holder'). Only the holder releases."; return 1
  fi
  rm -rf "$LOCK_DIR"; echo "released by $session"
}

case "${1:-status}" in
  acquire)   shift; cmd_acquire "$@" ;;
  heartbeat) cmd_heartbeat ;;
  status)    cmd_status ;;
  release)   shift; cmd_release "$@" ;;
  *) echo "usage: nofx-lock {acquire <session> <task> [mins]|heartbeat|status|release <session>}"; exit 64 ;;
esac
