#!/usr/bin/env bash
# vl-lock-test.sh — pins for the atomic heartbeat lock.
#
# The lock this replaces failed in three directions in one day (2026-09-03):
# a dead pid under a live owner, a live pid silently overwritten by a second
# writer, and a pid that went stale when its session was resumed. Every pin
# below is one of those, or the rule that makes them un-representable.
set -uo pipefail

LOCK_SH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/vl-lock.sh"
PASS=0; FAIL=0
ok()   { PASS=$((PASS+1)); printf '  ok   %s\n' "$1"; }
bad()  { FAIL=$((FAIL+1)); printf '  FAIL %s\n     %s\n' "$1" "${2:-}"; }
check(){ if [ "$2" = "$3" ]; then ok "$1"; else bad "$1" "want [$3] got [$2]"; fi; }
has()  { case "$2" in *"$3"*) ok "$1";; *) bad "$1" "missing [$3] in: $2";; esac; }
hasnt(){ case "$2" in *"$3"*) bad "$1" "found [$3] in: $2";; *) ok "$1";; esac; }
lower() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }
hasi()  { has  "$1" "$(lower "$2")" "$(lower "$3")"; }
hasnti(){ hasnt "$1" "$(lower "$2")" "$(lower "$3")"; }

# A source pin passes vacuously against a missing file, which is a false green
# of exactly the kind these tests exist to prevent.
if [ ! -f "$LOCK_SH" ]; then
  printf 'FAIL: %s does not exist — every pin below would be vacuous\n' "$LOCK_SH"
  exit 1
fi

WORK="$(mktemp -d)"; trap 'rm -rf "$WORK"' EXIT
export VL_LOCK_DIR="$WORK/vl-main.lock.d"
export VL_OLD_LOCK_DIR="$WORK/old-shape-nonexistent"
L() { VL_LOCK_DIR="$VL_LOCK_DIR" VL_OLD_LOCK_DIR="$VL_OLD_LOCK_DIR" bash "$LOCK_SH" "$@" 2>&1; }

echo "== a free lock reports free =="
out="$(L status)"; rc=$?
has  "status names it free" "$out" "free"
check "status rc on a free lock" "$rc" "0"

echo "== acquire, then a SECOND acquire FAILS (the replacement class) =="
out="$(L acquire vl-63 'cutover boot 6' 90)"; check "first acquire rc" "$?" "0"
hasi "first acquire confirms" "$out" "acquired"
out="$(L acquire vl-b3 'a different cutover' 90)"; rc=$?
check "second acquire rc is nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has  "second acquire refuses"        "$out" "REFUSED"
has  "second acquire names the holder" "$out" "vl-63"
has  "second acquire names the task"   "$out" "cutover boot 6"

echo "== the lock carries no pid, by construction =="
body="$(cat "$VL_LOCK_DIR"/* 2>/dev/null)"
hasnt "no pid= field in the lock"  "$body" "pid="
hasnt "no PID word in the lock"    "$body" "PID"
has   "records the session"        "$body" "vl-63"
has   "records the task"           "$body" "cutover boot 6"
has   "records acquired"           "$body" "acquired="
has   "records expiry"             "$body" "expiry="
has   "records a heartbeat"        "$body" "heartbeat="

echo "== a fresh heartbeat reads held, and never 'stale' =="
out="$(L status)"
has   "fresh status says held"     "$out" "held"
hasnti "fresh status is not stale" "$out" "stale"
check "check rc on held-fresh" "$(L check >/dev/null 2>&1; echo $?)" "1"

echo "== a STALE heartbeat says stale, and NEVER says dead =="
age_lock() { # rewrite the heartbeat the way the script actually reads it
  local secs="$1" f="$VL_LOCK_DIR/meta"
  { grep -v '^heartbeat' "$f"
    printf 'heartbeat=%s\n' "$(date -Is -d "$secs seconds ago")"
    printf 'heartbeat_epoch=%s\n' "$(( $(date +%s) - secs ))"
  } > "$f.tmp" && mv -f "$f.tmp" "$f"
}
age_lock 660
out="$(L status)"
hasi   "stale status says stale"      "$out" "stale"
hasnti "stale status never says dead" "$out" "dead"
hasi  "stale status demands corroboration" "$out" "corroborat"
has   "stale status still names the holder" "$out" "vl-63"
check "check rc on held-stale" "$(L check >/dev/null 2>&1; echo $?)" "2"

echo "== heartbeat refreshes it; a foreign session may not beat =="
L heartbeat vl-63 >/dev/null
out="$(L status)"
hasnti "beating clears stale" "$out" "stale"
out="$(L heartbeat vl-b3)"; rc=$?
check "foreign heartbeat rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "foreign heartbeat refuses" "$out" "REFUSED"

echo "== release is owner-scoped, and frees the lock =="
out="$(L release vl-b3)"; rc=$?
check "foreign release rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
out="$(L release vl-63)"; check "owner release rc" "$?" "0"
has   "release confirms" "$out" "released"
has   "lock is free again" "$(L status)" "free"
check "lock dir is gone" "$([ -e "$VL_LOCK_DIR" ] && echo present || echo gone)" "gone"

echo "== with-heartbeat keeps a long job fresh, and stops when it ends =="
L acquire vl-63 'long build' 90 >/dev/null
age_lock 660
L with-heartbeat vl-63 -- true >/dev/null 2>&1
hasnti "with-heartbeat beat at least once" "$(L status)" "stale"
before="$(cat "$VL_LOCK_DIR/heartbeat")"
sleep 1
check "beater does not outlive its command" "$(cat "$VL_LOCK_DIR/heartbeat")" "$before"
L release vl-63 >/dev/null

echo "== reclaim: REFUSED on a fresh heartbeat =="
L acquire vl-b3 'a live cutover' 90 >/dev/null
out="$(L reclaim vl-63 vl-b3 'HEAD has not moved in 20m; no build in flight')"; rc=$?
check "reclaim rc on fresh is nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "reclaim refuses a fresh lock" "$out" "REFUSED"
hasi  "refusal says the heartbeat is fresh" "$out" "fresh"
check "holder is unchanged after a refused reclaim" "$(L status | grep -c "vl-b3")" "1"

echo "== reclaim: allowed once STALE, and only with corroboration =="
age_lock 660
out="$(L reclaim vl-63 vl-b3)"; rc=$?
check "reclaim without corroboration rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
hasi  "refuses when corroboration is missing" "$out" "corroborat"
out="$(L reclaim vl-63 vl-ed 'HEAD static; no build')"; rc=$?
check "reclaim naming the WRONG stale session rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "refuses when the named session is not the holder" "$out" "REFUSED"
out="$(L reclaim vl-63 vl-b3 'HEAD has not moved in 20m; no build in flight')"; rc=$?
# rc 3, NOT acquire's 0 — a script must be able to tell "took a free lock" from
# "inherited an abandoned one", so a lane can refuse to inherit.
check "reclaim rc is 3, distinct from acquire's 0" "$rc" "3"
hasi  "reclaim confirms" "$out" "reclaim"
out="$(L status)"
has   "new holder is named"        "$out" "vl-63"
hasnti "reclaimed lock is fresh"   "$out" "stale"

echo "== the reclaim is written to the lock's history =="
hist="$(cat "$VL_LOCK_DIR/history" 2>/dev/null)"
has  "history names who took over"      "$hist" "vl-63"
has  "history names who was taken from" "$hist" "vl-b3"
has  "history carries the corroboration" "$hist" "no build in flight"
has  "history is timestamped"           "$hist" "20"
check "history survives into status" "$(L status | grep -c 'reclaim')" "1"

echo "== a reclaimed lock still beats and releases as the new holder =="
check "old holder may no longer beat" "$(L heartbeat vl-b3 >/dev/null 2>&1; echo $?)" "1"
L heartbeat vl-63 >/dev/null; check "new holder may beat" "$?" "0"
L release vl-63 >/dev/null; has "released" "$(L status)" "free"

echo "== the script cannot express pid liveness =="
src="$(grep -v '^[[:space:]]*#' "$LOCK_SH" | sed 's/[[:space:]]#.*$//')"
hasnt "no kill -0"  "$src" "kill -0"
hasnt "no pgrep"    "$src" "pgrep"
hasnt "no \$\$"     "$src" '$$'

echo
printf 'pass=%d fail=%d\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
