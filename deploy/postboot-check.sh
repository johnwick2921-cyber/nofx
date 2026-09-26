#!/usr/bin/env bash
# postboot-check.sh — BOOT-DAY OBSERVATION KIT (DS-101, 2026-09-26)
#
# READ-ONLY. This script never writes a file, never opens the DB for writing,
# never calls the bot's API, never kills or restarts anything. Every DB touch
# goes through `sqlite3 -readonly "file:...?mode=ro"`.
#
# Usage:
#   deploy/postboot-check.sh [logfile]      # run every check against one log
#   deploy/postboot-check.sh --selfcheck    # verify every expected string this
#                                           # kit greps for still exists in the
#                                           # repo sources (anti-rot)
#
# Env overrides (defaults are the production layout):
#   NOFX_REPO=/home/hoang/nofx   NOFX_DATA=$NOFX_REPO/data
#   NOFX_BIN=$NOFX_REPO/nofx-bin NOFX_ENV=$NOFX_REPO/.env
#
# Exit code = number of FAILs (0 = all green). Run right after the boot line
# sweep and again at Sunday 17:00 CT open (the SUBSET: planner-max, fast-market
# max, salvage counters, no-ERROR — see deploy/POSTBOOT-CHECK.md).
set -u

# The repo this script ships with is the selfcheck/baseline source of truth
# (the deployed tree may still be the OLD head before the weekend boot).
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
KIT_REPO=$(cd "$SCRIPT_DIR/.." && pwd)

NOFX_REPO=${NOFX_REPO:-$KIT_REPO}
NOFX_DATA=${NOFX_DATA:-$NOFX_REPO/data}
NOFX_BIN=${NOFX_BIN:-$NOFX_REPO/nofx-bin}
NOFX_ENV=${NOFX_ENV:-$NOFX_REPO/.env}
DB=$NOFX_DATA/data.db

# The documented `_ =` baseline at dev 35a53d29 (regex below, non-test Go
# files). Bump deliberately in deploy/POSTBOOT-CHECK.md when a reviewed wave
# adds a justified hit.
UNDERSCORE_EQ_BASELINE=432

fail=0
pass=0
ok()  { pass=$((pass+1)); echo "PASS $1"; }
bad() { fail=$((fail+1)); echo "FAIL $1"; }

sql() { sqlite3 -readonly "file:$DB?mode=ro" "$1" 2>/dev/null; }

# ---------------------------------------------------------------------------
# 1. stop-entry (attempted-entry) guard — the boot line reads the SAME
#    predicate the placement branch calls; guard=MISROUTED is a refused boot.
# ---------------------------------------------------------------------------
check_stop_entry() {
  local line
  line=$(grep -m1 '🎯 stop-entry:' "$LOG" 2>/dev/null)
  [ -n "$line" ] || { bad "stop-entry boot line absent"; return; }
  echo "$line" | grep -q 'guard=MISROUTED' && { bad "stop-entry guard MISROUTED: $line"; return; }
  echo "$line" | grep -q 'seam=' || { bad "stop-entry seam not stated: $line"; return; }
  ok "stop-entry guard line present and routed: $line"
}

# ---------------------------------------------------------------------------
# 2. retention off — every table =off(keep all, rows=N); the boot line's N is
#    cross-checked against the LIVE row counts (counters are read, not faked).
# ---------------------------------------------------------------------------
check_retention() {
  local line table want got
  line=$(grep -m1 '🧹 retention:' "$LOG" 2>/dev/null)
  [ -n "$line" ] || { bad "retention boot line absent"; return; }
  for pair in "decision_records:decision_records" "equity_snapshots:equity_snapshots" "nt8_order_snapshots:nt8_order_snapshots" "level_stats:level_stats"; do
    table=${pair%%:*}
    want=$(echo "$line" | grep -oE "$table=off\(keep all, rows=[0-9]+\)" | grep -oE '[0-9]+' | head -1)
    if [ -z "$want" ]; then bad "retention: $table not off(keep all) in boot line"; continue; fi
    got=$(sql "SELECT COUNT(*) FROM $table;")
    [ "$got" = "$want" ] && ok "retention $table off(keep all), live rows $got == boot $want" \
                          || bad "retention $table: boot says rows=$want, live COUNT=$got"
  done
}

# ---------------------------------------------------------------------------
# 3. JWT — the boot line fires UNCONDITIONALLY, so the real proof is .env:
#    a secret ≥24 chars that is NOT the shipped default.
# ---------------------------------------------------------------------------
check_jwt() {
  grep -q '🔑 JWT secret configured' "$LOG" 2>/dev/null || { bad "JWT boot line absent"; return; }
  local v
  v=$(grep -E '^JWT_SECRET=' "$NOFX_ENV" 2>/dev/null | cut -d= -f2-)
  if [ -z "$v" ]; then bad "JWT_SECRET unset in .env (default-jwt-secret in force)"; return; fi
  if [ "$v" = "default-jwt-secret-change-in-production" ]; then bad "JWT_SECRET is the shipped default"; return; fi
  [ ${#v} -ge 24 ] && ok "JWT secret set, length ${#v}" || bad "JWT_SECRET shorter than 24 chars"
}

# ---------------------------------------------------------------------------
# 4. transport — the per-trader line names the resolved transport; the CSV
#    fallback WARN must not exist on the live boot (NT_TRANSPORT=tcp).
# ---------------------------------------------------------------------------
check_transport() {
  grep -q 'Using NinjaTrader (transport via NT_TRANSPORT env' "$LOG" 2>/dev/null \
    || { bad "transport boot line absent"; return; }
  if grep -q '⚠️ transport:' "$LOG" 2>/dev/null; then bad "transport WARN present (CSV fallback): $(grep -m1 '⚠️ transport:' "$LOG")"; else ok "transport line present, no CSV-fallback WARN"; fi
  grep -qE '^NT_TRANSPORT=tcp' "$NOFX_ENV" 2>/dev/null && ok "NT_TRANSPORT=tcp in .env" \
    || bad "NT_TRANSPORT=tcp not found in .env"
}

# ---------------------------------------------------------------------------
# 5. fast-market max — owner rule: MAX everywhere. FAST_MARKET_REASONING must
#    be max, and no read may log a fast-market reasoning downgrade.
# ---------------------------------------------------------------------------
check_fast_market_max() {
  grep -qE '^FAST_MARKET_REASONING=max' "$NOFX_ENV" 2>/dev/null \
    || { bad "FAST_MARKET_REASONING=max not in .env (owner rule: MAX everywhere)"; }
  if grep -q 'planner mode: fast-market.*reasoning downgraded to ' "$LOG" 2>/dev/null; then
    bad "fast-market downgrade observed: $(grep -m1 'planner mode: fast-market' "$LOG")"
  else
    ok "no fast-market reasoning downgrade in log"
  fi
}

# ---------------------------------------------------------------------------
# 6. death-reread retry — mechanism line on (never off) + the per-strategy
#    death_reread_retry_min knob as stored (nil = wake_min_interval_min).
# ---------------------------------------------------------------------------
check_death_reread() {
  grep -q '🧬 death→reread=on' "$LOG" 2>/dev/null || { bad "death→reread not on in boot"; return; }
  ok "death→reread=on"
  local rows
  rows=$(sql "SELECT id, config FROM strategies WHERE config IS NOT NULL AND config != '';")
  if [ -z "$rows" ]; then bad "no strategy rows to read death_reread_retry_min from"; return; fi
  echo "$rows" | while IFS='|' read -r id cfg; do
    v=$(echo "$cfg" | python3 -c 'import json,sys
try:
  d=json.load(sys.stdin).get("day_plan") or {}
  print(d.get("death_reread_retry_min","<nil>"))
except Exception:
  print("<unparsable>")' 2>/dev/null)
    echo "INFO strategy $id day_plan.death_reread_retry_min=$v (nil = wake_min_interval_min)"
  done
}

# ---------------------------------------------------------------------------
# 7. busy_timeout — the shipped binary must embed the per-connection DSN
#    pragma (#246 P1-B). A PRAGMA run on one connection only does not prove
#    the pool; the binary carrying the DSN suffix does.
# ---------------------------------------------------------------------------
check_busy_timeout() {
  if strings -n 8 "$NOFX_BIN" 2>/dev/null | grep -qE '_pragma=busy_timeout\(5000\)|_busy_timeout=5000'; then
    ok "binary carries the per-connection busy_timeout DSN"
  else
    bad "binary does not embed the busy_timeout DSN (or strings unavailable)"
  fi
}

# ---------------------------------------------------------------------------
# 8. first planner read at max — the FIRST planner call line after boot must
#    read reasoning=max, and the per-trader lifecycle line plan_reasoning=max.
# ---------------------------------------------------------------------------
check_planner_max() {
  local first
  first=$(grep -m1 '🧠 planner call (reasoning=' "$LOG" 2>/dev/null)
  [ -n "$first" ] || { bad "no planner call line found"; return; }
  echo "$first" | grep -q 'reasoning=max' && ok "first planner call at max: ${first:0:120}" \
    || bad "first planner call not max: ${first:0:120}"
  grep -m1 '🧬 plan lifecycle:' "$LOG" 2>/dev/null | grep -q 'plan_reasoning=max' \
    && ok "plan lifecycle line reads plan_reasoning=max" \
    || bad "plan lifecycle line missing plan_reasoning=max"
}

# ---------------------------------------------------------------------------
# 9. born-dead salvage counters — boot counter == live event count
#    (counters record; never infer).
# ---------------------------------------------------------------------------
check_salvage_counters() {
  local line want got
  line=$(grep -m1 'plan liveness:' "$LOG" 2>/dev/null)
  [ -n "$line" ] || { bad "plan liveness boot line absent"; return; }
  want=$(echo "$line" | grep -oE 'born-dead dropped=[0-9]+' | grep -oE '[0-9]+' | head -1)
  [ -n "$want" ] || { bad "born-dead dropped counter absent from boot line"; return; }
  got=$(sql "SELECT COUNT(*) FROM system_config WHERE key LIKE 'plan_liveness_event:born_dead_dropped:%';")
  [ "$got" = "$want" ] && ok "born-dead dropped counter boot $want == live $got" \
    || bad "born-dead dropped: boot says $want, live COUNT=$got"
}

# ---------------------------------------------------------------------------
# 10. no ERROR lines — the boot and the trading day must be clean; any [ERRO]
#     is printed (a human must look).
# ---------------------------------------------------------------------------
check_no_errors() {
  local n
  n=$(grep -c '\[ERRO\]' "$LOG" 2>/dev/null || true)
  if [ "$n" = "0" ]; then ok "zero [ERRO] lines"; else
    bad "$n [ERRO] line(s):"
    grep '\[ERRO\]' "$LOG" | head -5 | sed 's/^/    /'
  fi
}

# ---------------------------------------------------------------------------
# 11. no `_ =` regressions — the regex-counted baseline is documented in
#     deploy/POSTBOOT-CHECK.md (432 at dev 35a53d29). An increase fails.
# ---------------------------------------------------------------------------
check_underscore_eq() {
  local n
  n=$(git -C "$NOFX_REPO" grep -E '_ = [a-zA-Z_]' -- '*.go' ':!*_test.go' 2>/dev/null | wc -l | tr -d ' ')
  if [ -n "$n" ] && [ "$n" -le "$UNDERSCORE_EQ_BASELINE" ]; then ok "_ = hits $n ≤ baseline $UNDERSCORE_EQ_BASELINE"; else
    bad "_ = hits $n > baseline $UNDERSCORE_EQ_BASELINE"
  fi
}

# ---------------------------------------------------------------------------
# --selfcheck: every needle the kit greps for must exist in the repo sources,
# so a renamed/removed boot line fails the kit, not the boot.
# ---------------------------------------------------------------------------
selfcheck() {
  local f=0
  while IFS='|' read -r needle where; do
    [ -z "$needle" ] && continue
    read -ra paths <<<"$where"
    if git -C "$KIT_REPO" grep -qF -- "$needle" -- "${paths[@]}" 2>/dev/null; then
      echo "SELFCHECK OK   $needle"
    else
      echo "SELFCHECK FAIL $needle (expected in $where)"
      f=$((f+1))
    fi
  done <<'NEEDLES'
🎯 stop-entry:|trader/*.go
🧹 retention:|main.go internal/retention/*.go
🔑 JWT secret configured|main.go
transport via NT_TRANSPORT env|trader/*.go
⚠️ transport:|trader/ninjatrader/*.go
🧠 planner call (reasoning=|trader/*.go
planner mode: fast-market|trader/*.go
🧬 death→reread=on|trader/*.go
🧬 plan lifecycle:|trader/*.go
born-dead dropped=|trader/*.go
_pragma=busy_timeout(5000)|store/sqlitedriver/*.go
NEEDLES
  exit $f
}

[ "${1:-}" = "--selfcheck" ] && selfcheck

LOG=${1:-}
if [ -z "$LOG" ]; then
  LOG=$(ls -t "$NOFX_DATA"/nofx_*.log 2>/dev/null | head -1)
fi
if [ -z "$LOG" ] || [ ! -r "$LOG" ]; then
  echo "no readable log (pass the path; looked in $NOFX_DATA)" >&2
  exit 2
fi
echo "== postboot-check against $LOG"
for t in sqlite3 strings python3 git; do command -v "$t" >/dev/null || echo "WARN $t missing — related checks will FAIL"; done

check_stop_entry
check_retention
check_jwt
check_transport
check_fast_market_max
check_death_reread
check_busy_timeout
check_planner_max
check_salvage_counters
check_no_errors
check_underscore_eq

echo "== $pass PASS, $fail FAIL"
exit $fail
