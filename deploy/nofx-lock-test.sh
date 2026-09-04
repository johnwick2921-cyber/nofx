#!/usr/bin/env bash
# Pins for the atomic heartbeat lock (owner ruling 2026-09-03).
set -uo pipefail
export NOFX_LOCK_DIR="$(mktemp -d)/lock.d"
export NOFX_LEGACY_LOCK="$(mktemp -d)/legacy-absent"
L="$(dirname "$0")/nofx-lock.sh"
fail=0
check(){ if [ "$2" = "$3" ]; then echo "  PASS  $1"; else echo "  FAIL  $1 — got '$2' want '$3'"; fail=1; fi; }

# PIN 1 — a second acquire FAILS. This is the race that clobbered a live lock
# on 2026-09-03 ("their lock had replaced mine at 21:43").
out1=$("$L" acquire nofx-A "first boot" 45 >/dev/null 2>&1; echo $?)
out2=$("$L" acquire nofx-B "second boot" 45 >/dev/null 2>&1; echo $?)
check "first acquire succeeds"  "$out1" "0"
check "SECOND acquire FAILS"    "$out2" "1"

# PIN 2 — a non-holder cannot release.
r=$("$L" release nofx-B >/dev/null 2>&1; echo $?)
check "non-holder cannot release" "$r" "1"

# PIN 3 — a fresh heartbeat reads ALIVE.
s=$("$L" status 2>&1)
case "$s" in *ALIVE*) echo "  PASS  fresh heartbeat is ALIVE";; *) echo "  FAIL  fresh status: $s"; fail=1;; esac

# PIN 4 — A STALE HEARTBEAT IS REPORTED STALE, NEVER DEAD. The distinction is
# the whole ruling: pid 1860416 died while its holder kept working.
sed -i "s/^heartbeat_epoch=.*/heartbeat_epoch=$(( $(date +%s) - 999 ))/" "$NOFX_LOCK_DIR/meta"
s=$("$L" status 2>&1); rc=$?
case "$s" in
  *STALE*) echo "  PASS  old heartbeat reports STALE";;
  *)       echo "  FAIL  stale status: $s"; fail=1;;
esac
case "$s" in
  *DEAD*|*dead*) echo "  FAIL  status said DEAD — a stale heartbeat is not a dead holder"; fail=1;;
  *)            echo "  PASS  never says DEAD";;
esac
case "$s" in
  *"DO NOT CLEAR ON THIS ALONE"*) echo "  PASS  demands corroboration before clearing";;
  *) echo "  FAIL  stale status must demand corroboration"; fail=1;;
esac
check "stale exits 2 (distinct from free/held)" "$rc" "2"

# PIN 5 — the holder can release, and the lock is then free.
"$L" release nofx-A >/dev/null 2>&1
s=$("$L" status 2>&1)
check "after release the lock is free" "$s" "free"

[ "$fail" = "0" ] && { echo "ALL LOCK PINS PASS"; exit 0; } || { echo "LOCK PINS FAILED"; exit 1; }
