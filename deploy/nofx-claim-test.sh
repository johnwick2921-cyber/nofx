#!/usr/bin/env bash
# Pins for nofx-claim. The load-bearing one is REAL-2: the actual malformed
# claim from 2026-09-04 must FAIL. A checker that passes the message that
# caused the incident is decoration.
set -uo pipefail
CLAIM="$(dirname "$0")/nofx-claim.sh"
RE="$(grep -oP "^CLAIM_RE='\K.*(?='$)" "$CLAIM")"
pass=0; fail=0
t() { # t <name> <expect:ok|bad> <message>
  local name="$1" expect="$2" msg="$3" got
  if printf '%s' "$msg" | grep -qE "$RE"; then got=ok; else got=bad; fi
  if [ "$got" = "$expect" ]; then pass=$((pass+1)); echo "  PASS $name"
  else fail=$((fail+1)); echo "  FAIL $name — expected $expect, got $got"; echo "       msg: $msg"; fi
}

echo "REAL messages from 2026-09-04:"
# The message that caused the incident. MUST be rejected.
t REAL-1-reaper-no-session bad \
  "claim: reaper reads the snapshot, not order_update silence (PART 3 step 0)"
# nofx-47's well-formed one. MUST be accepted.
t REAL-2-nofx47-ok ok \
  "claim: reaper reads the broker snapshot, not order_update silence — nofx-47, 2026-09-04T09:44:00-05:00"
# This wave's own claim.
t REAL-3-self ok \
  "claim: claim-identity enforcement + worktree prune — nofx-b3, 2026-09-04T10:10:00-05:00"

echo "shape pins:"
t no-session          bad "claim: some wave, 2026-09-04T09:44:00-05:00"
t no-timestamp        bad "claim: some wave — nofx-x"
t date-not-iso        bad "claim: some wave — nofx-x, Sep 4 2026"
t iso-utc-z           ok  "claim: some wave — nofx-x, 2026-09-04T09:44:00Z"
t iso-no-colon-offset ok  "claim: some wave — nofx-x, 2026-09-04T09:44:00-0500"
t not-a-claim         bad "fix(thing): unrelated commit"
t empty-session       bad "claim: some wave — , 2026-09-04T09:44:00-05:00"

echo "---- $pass passed, $fail failed"
[ "$fail" -eq 0 ]
