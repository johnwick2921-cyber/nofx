#!/usr/bin/env bash
# Round 24 Q4 — export the inputs the harness reads. EVERY database access in this
# study goes through `sqlite3 -readonly`; the Go harness never opens a database.
#   copy DB : /home/hoang/nofx-r101/data/db.copy.db (snapshot of the live store, 2026-09-16 ~15:59 CT)
#   live DB : /home/hoang/nofx/data/data.db (read-only, tail rows newer than the copy)
# Outputs land in out-r24/ (gitignored; sha256 of each file goes in the report MANIFEST).
set -euo pipefail
cd "$(dirname "$0")"
COPY=/home/hoang/nofx-r101/data/db.copy.db
LIVE=/home/hoang/nofx/data/data.db
OUT=out-r24
mkdir -p "$OUT"
q() { sqlite3 -readonly -json "$1" "$2"; }

# --- copy: plans, lifecycle, replan counters, the live trader's day_plan config ---
q "$COPY" "select plan_id, version, strategy_id, trade_date, session, trigger_reason, lifecycle, doc, created_at from plans order by created_at" > "$OUT/copy_plans.json"
q "$COPY" "select id, plan_id, version, event, reason, at from plan_lifecycle_log order by id" > "$OUT/copy_lifecycle.json"
q "$COPY" "select key, value from system_config where key like 'dayplan_replans_used:%' or key like 'dayplan_reset_baseline:%'" > "$OUT/copy_sysconfig_replans.json"
q "$COPY" "select id, json_extract(config,'$.day_plan') as day_plan from strategies where id=(select strategy_id from traders where is_running=1 order by updated_at desc limit 1)" > "$OUT/copy_strategy_dayplan.json"
# --- copy: MNQ bars (all contracts/sources; the harness applies the BarsBetweenOn source exclusions) ---
q "$COPY" "select tf, open_time_ms, o, h, l, c, v, coalesce(contract,'') contract, coalesce(source,'') source from bars where symbol='MNQ' and tf='1m'  and open_time_ms >= 1785000000000 order by open_time_ms" > "$OUT/copy_bars_1m.json"
q "$COPY" "select tf, open_time_ms, o, h, l, c, v, coalesce(contract,'') contract, coalesce(source,'') source from bars where symbol='MNQ' and tf='15m' and open_time_ms >= 1780272000000 order by open_time_ms" > "$OUT/copy_bars_15m.json"
q "$COPY" "select tf, open_time_ms, o, h, l, c, v, coalesce(contract,'') contract, coalesce(source,'') source from bars where symbol='MNQ' and tf='1h'  and open_time_ms >= 1767225600000 order by open_time_ms" > "$OUT/copy_bars_1h.json"
q "$COPY" "select tf, open_time_ms, o, h, l, c, v, coalesce(contract,'') contract, coalesce(source,'') source from bars where symbol='MNQ' and tf='4h' order by open_time_ms" > "$OUT/copy_bars_4h.json"
q "$COPY" "select tf, open_time_ms, o, h, l, c, v, coalesce(contract,'') contract, coalesce(source,'') source from bars where symbol='MNQ' and tf='1d' order by open_time_ms" > "$OUT/copy_bars_1d.json"

# --- live: only rows newer than the copy (plans by created_at; bars from 2026-09-10) ---
COPY_MAX=$(sqlite3 -readonly "$COPY" "select max(created_at) from plans")
sqlite3 -readonly "$LIVE" "select max(created_at) from plans" > "$OUT/live_plans_max_created_at.txt"
q "$LIVE" "select plan_id, version, strategy_id, trade_date, session, trigger_reason, lifecycle, doc, created_at from plans where created_at > '$COPY_MAX' order by created_at" > "$OUT/live_plans_tail.json"
q "$LIVE" "select id, plan_id, version, event, reason, at from plan_lifecycle_log where at > '2026-09-16 14:10:00' order by id" > "$OUT/live_lifecycle_tail.json"
q "$LIVE" "select key, value from system_config where key like 'dayplan_replans_used:%' or key like 'dayplan_reset_baseline:%'" > "$OUT/live_sysconfig_replans.json"
q "$LIVE" "select tf, open_time_ms, o, h, l, c, v, coalesce(contract,'') contract, coalesce(source,'') source from bars where symbol='MNQ' and tf in ('1m','15m','1h','4h','1d') and open_time_ms >= 1789000000000 order by tf, open_time_ms" > "$OUT/live_bars_tail.json"

sha256sum "$COPY" "$OUT"/*.json > "$OUT/SHA256SUMS"
echo "copy_plans_max_created_at=$COPY_MAX"
cat "$OUT/live_plans_max_created_at.txt"
wc -l "$OUT"/*.json | tail -1
