#!/usr/bin/env bash
# nofx-lock-test.sh — pins for the atomic heartbeat lock.
#
# The lock this replaces failed in three directions in one day (2026-09-03):
# a dead pid under a live owner, a live pid silently overwritten by a second
# writer, and a pid that went stale when its session was resumed. Every pin
# below is one of those, or the rule that makes them un-representable.
set -uo pipefail

LOCK_SH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/nofx-lock.sh"
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
export NOFX_LOCK_DIR="$WORK/nofx-main.lock.d"
L() { NOFX_LOCK_DIR="$NOFX_LOCK_DIR" bash "$LOCK_SH" "$@" 2>&1; }

echo "== a free lock reports free =="
out="$(L status)"; rc=$?
has  "status names it free" "$out" "free"
check "status rc on a free lock" "$rc" "0"

echo "== acquire, then a SECOND acquire FAILS (the replacement class) =="
out="$(L acquire nofx-63 'cutover boot 6' 90)"; check "first acquire rc" "$?" "0"
hasi "first acquire confirms" "$out" "acquired"
out="$(L acquire nofx-b3 'a different cutover' 90)"; rc=$?
check "second acquire rc is nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has  "second acquire refuses"        "$out" "REFUSED"
has  "second acquire names the holder" "$out" "nofx-63"
has  "second acquire names the task"   "$out" "cutover boot 6"

echo "== the lock carries no pid, by construction =="
body="$(cat "$NOFX_LOCK_DIR"/* 2>/dev/null)"
hasnt "no pid= field in the lock"  "$body" "pid="
hasnt "no PID word in the lock"    "$body" "PID"
has   "records the session"        "$body" "nofx-63"
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
  local secs="$1" f="$NOFX_LOCK_DIR/meta"
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
has   "stale status still names the holder" "$out" "nofx-63"
check "check rc on held-stale" "$(L check >/dev/null 2>&1; echo $?)" "2"

echo "== heartbeat refreshes it; a foreign session may not beat =="
L heartbeat nofx-63 >/dev/null
out="$(L status)"
hasnti "beating clears stale" "$out" "stale"
out="$(L heartbeat nofx-b3)"; rc=$?
check "foreign heartbeat rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "foreign heartbeat refuses" "$out" "REFUSED"

echo "== release is owner-scoped, and frees the lock =="
out="$(L release nofx-b3)"; rc=$?
check "foreign release rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
out="$(L release nofx-63)"; check "owner release rc" "$?" "0"
has   "release confirms" "$out" "released"
has   "lock is free again" "$(L status)" "free"
check "lock dir is gone" "$([ -e "$NOFX_LOCK_DIR" ] && echo present || echo gone)" "gone"

echo "== with-heartbeat keeps a long job fresh, and stops when it ends =="
L acquire nofx-63 'long build' 90 >/dev/null
age_lock 660
L with-heartbeat nofx-63 -- true >/dev/null 2>&1
hasnti "with-heartbeat beat at least once" "$(L status)" "stale"
before="$(cat "$NOFX_LOCK_DIR/heartbeat")"
sleep 1
check "beater does not outlive its command" "$(cat "$NOFX_LOCK_DIR/heartbeat")" "$before"
L release nofx-63 >/dev/null

echo "== reclaim: REFUSED on a fresh heartbeat =="
L acquire nofx-b3 'a live cutover' 90 >/dev/null
out="$(L reclaim nofx-63 nofx-b3 'HEAD has not moved in 20m; no build in flight')"; rc=$?
check "reclaim rc on fresh is nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "reclaim refuses a fresh lock" "$out" "REFUSED"
hasi  "refusal says the heartbeat is fresh" "$out" "fresh"
check "holder is unchanged after a refused reclaim" "$(L status | grep -c "nofx-b3")" "1"

echo "== reclaim: allowed once STALE, and only with corroboration =="
age_lock 660
out="$(L reclaim nofx-63 nofx-b3)"; rc=$?
check "reclaim without corroboration rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
hasi  "refuses when corroboration is missing" "$out" "corroborat"
out="$(L reclaim nofx-63 nofx-ed 'HEAD static; no build')"; rc=$?
check "reclaim naming the WRONG stale session rc nonzero" "$([ $rc -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "refuses when the named session is not the holder" "$out" "REFUSED"
out="$(L reclaim nofx-63 nofx-b3 'HEAD has not moved in 20m; no build in flight')"; rc=$?
# rc 3, NOT acquire's 0 — a script must be able to tell "took a free lock" from
# "inherited an abandoned one", so a lane can refuse to inherit.
check "reclaim rc is 3, distinct from acquire's 0" "$rc" "3"
hasi  "reclaim confirms" "$out" "reclaim"
out="$(L status)"
has   "new holder is named"        "$out" "nofx-63"
hasnti "reclaimed lock is fresh"   "$out" "stale"

echo "== the reclaim is written to the lock's history =="
hist="$(cat "$NOFX_LOCK_DIR/history" 2>/dev/null)"
has  "history names who took over"      "$hist" "nofx-63"
has  "history names who was taken from" "$hist" "nofx-b3"
has  "history carries the corroboration" "$hist" "no build in flight"
has  "history is timestamped"           "$hist" "20"
check "history survives into status" "$(L status | grep -c 'reclaim')" "1"

echo "== a reclaimed lock still beats and releases as the new holder =="
check "old holder may no longer beat" "$(L heartbeat nofx-b3 >/dev/null 2>&1; echo $?)" "1"
L heartbeat nofx-63 >/dev/null; check "new holder may beat" "$?" "0"
L release nofx-63 >/dev/null; has "released" "$(L status)" "free"

echo "== the script cannot express pid liveness =="
src="$(grep -v '^[[:space:]]*#' "$LOCK_SH" | sed 's/[[:space:]]#.*$//')"
hasnt "no kill -0"  "$src" "kill -0"
hasnt "no pgrep"    "$src" "pgrep"
hasnt "no \$\$"     "$src" '$$'

# ── THE KEEPER (2026-09-09) ─────────────────────────────────────────────────
#
# acquire printed "heartbeat every 120s" and started NOTHING. A holder who
# simply WAITED — for a position to close, for an owner to run the kill — went
# STALE at 300s with no writer in existence, and status then printed the reclaim
# recipe over a live cutover. Measured: acquired 21:25:07, and at 21:50:16 the
# meta's heartbeat was byte-identical to `acquired`, never beaten once.
#
# The verb was never broken — a hand-beat revives it. Nothing called it for a
# holder who was not running commands. Class 88's shape: a liveness signal that
# is a side effect of activity, dying the moment the work pauses, when the whole
# point of a lock is to be held while you WAIT.
#
# THE KEEPER IS BOUNDED BY THE DECLARED EXPIRY, and that bound is the design.
# An UNBOUNDED keeper survives its session and beats forever for a holder who is
# gone — turning a 5-minute false STALE (recoverable, and the canon already says
# corroborate) into a permanent false ALIVE that no succession path can reach.
# Bounded, an abandoned lock still goes stale — at the window its holder asked
# for, instead of never.
#
# Thresholds are compressed so these run in seconds, not the 400 the real window
# would need.
echo "== the keeper: acquire starts one, and a WAITING holder stays ALIVE =="
KW="$(mktemp -d)"
K() { NOFX_LOCK_DIR="$KW/lock.d" NOFX_LOCK_STALE_SECONDS=4 NOFX_LOCK_BEAT_SECONDS=1 bash "$LOCK_SH" "$@" 2>&1; }
out="$(K acquire nofx-keeper 'waiting on a position to close' 60)"
check "acquire rc" "$?" "0"
hasi  "acquire confirms" "$out" "acquired"
hasnti "the acquire message no longer promises an unstarted beat" "$out" "heartbeat every"

sleep 6   # LONGER than the stale window, with NO other action of any kind
out="$(K status)"
hasi   "a WAITING holder is still ALIVE past the stale window" "$out" "alive"
hasnti "a waiting holder is not reported STALE" "$out" "stale"
hasi   "status reports the auto-beat" "$out" "auto-beat"

echo "== killing the keeper makes it STALE — the beat is a process, not a fiction =="
kp="$(cat "$KW/lock.d/keeper.pid" 2>/dev/null || echo)"
if [ -n "$kp" ]; then
  ok "a keeper pid is recorded as a stop handle"
  kill "$kp" 2>/dev/null; sleep 6
  hasi "with the keeper dead the lock goes STALE" "$(K status)" "stale"
else
  bad "a keeper pid is recorded as a stop handle" "no $KW/lock.d/keeper.pid"
fi

echo "== a second acquire never succeeds while the lock lives =="
out="$(K acquire nofx-other 'a different cutover' 60)"; rc=$?
check "second acquire rc is a refusal" "$rc" "1"
hasi  "second acquire is REFUSED" "$out" "refused"

echo "== release stops the keeper — no writer outlives its lock =="
# THE PROCESS, NOT THE FILE, AND ON A BEAT LONG ENOUGH TO SEE IT.
#
# Two drafts of this pin failed to bite. The first checked only that keeper.pid
# was gone — which rm -rf "$LOCK_DIR" does anyway, so deleting _stop_keeper left
# it green with an orphan still looping. The second checked the PROCESS but ran
# on a 1s beat, so an unstopped keeper noticed the missing directory and exited
# by itself before the assertion — the pin was measuring the OS's timing, not
# the code. With a 30s beat an unstopped keeper is still sleeping when we look,
# and only _stop_keeper can have ended it.
#
# (The script may not use kill -0 — class 70 — but this test may: the pins forbid
# pid liveness in the TOOL, where a reader could mistake it for an answer, not in
# a test that is asking about a process on purpose.)
KWR="$(mktemp -d)"
R() { NOFX_LOCK_DIR="$KWR/lock.d" NOFX_LOCK_STALE_SECONDS=600 NOFX_LOCK_BEAT_SECONDS=30 bash "$LOCK_SH" "$@" 2>&1; }
R acquire nofx-rel 'a holder that will release' 60 >/dev/null
kp2="$(cat "$KWR/lock.d/keeper.pid" 2>/dev/null || echo)"
hasi "release confirms" "$(R release nofx-rel)" "released"
sleep 1
if [ -n "$kp2" ] && kill -0 "$kp2" 2>/dev/null; then
  bad "release STOPS the keeper process" "pid $kp2 still running after release — a writer outliving its lock"
  kill "$kp2" 2>/dev/null
else
  ok "release STOPS the keeper process"
fi
rm -rf "$KWR"
rm -rf "$KW"

echo "== the keeper OUTLIVES the shell that spawned it =="
# A keeper that dies with its session reproduces the defect one layer down.
KW2="$(mktemp -d)"
( NOFX_LOCK_DIR="$KW2/lock.d" NOFX_LOCK_STALE_SECONDS=4 NOFX_LOCK_BEAT_SECONDS=1 \
    bash "$LOCK_SH" acquire nofx-detached 'spawned by a shell that exits' 60 >/dev/null 2>&1 )
sleep 6
out="$(NOFX_LOCK_DIR="$KW2/lock.d" NOFX_LOCK_STALE_SECONDS=4 bash "$LOCK_SH" status 2>&1)"
hasi "a keeper spawned by an exited shell keeps beating" "$out" "alive"
NOFX_LOCK_DIR="$KW2/lock.d" bash "$LOCK_SH" release nofx-detached >/dev/null 2>&1
rm -rf "$KW2"

echo "== THE BOUND: the keeper stops at the declared expiry, and never extends it =="
KW3="$(mktemp -d)"
# a lock declared for 0 minutes is already expired: the keeper must not beat it
( NOFX_LOCK_DIR="$KW3/lock.d" NOFX_LOCK_STALE_SECONDS=4 NOFX_LOCK_BEAT_SECONDS=1 \
    bash "$LOCK_SH" acquire nofx-brief 'a window that has already closed' 0 >/dev/null 2>&1 )
sleep 6
out="$(NOFX_LOCK_DIR="$KW3/lock.d" NOFX_LOCK_STALE_SECONDS=4 bash "$LOCK_SH" status 2>&1)"
hasi "a lock past its declared expiry goes STALE — the keeper never auto-extends" "$out" "stale"
NOFX_LOCK_DIR="$KW3/lock.d" bash "$LOCK_SH" release nofx-brief >/dev/null 2>&1
rm -rf "$KW3"

echo "== the acquire message says what the keeper actually does =="
src2="$(cat "$LOCK_SH")"
case "$src2" in
  *keeper.pid*) ok "the acquire path records a keeper" ;;
  *) bad "the acquire path records a keeper" "no keeper.pid anywhere — the message would be the defect again" ;;
esac

echo
printf 'pass=%d fail=%d\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
