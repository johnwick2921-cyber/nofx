#!/usr/bin/env bash
# Round 24 Q4 — read-only extraction of every input, via sqlite3 -readonly ONLY.
# Live store + research store are never opened by anything but sqlite3 -readonly.
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
IN="$HERE/out-r24/in"; mkdir -p "$IN"
DB=/home/hoang/nofx/data/data.db
RDB=/home/hoang/nofx/data/data.db.research.db
SINCE_MS=$(( $(date -d '2026-08-10 00:00:00 CDT' +%s) * 1000 ))

# 1. MNQ bars, every tf the planner/harness uses, last ~38 days, BarsBetweenOn exclusions.
sqlite3 -readonly -csv -header "$DB" "
  SELECT tf, contract, open_time_ms, o, h, l, c, v, COALESCE(source,'') source
  FROM bars WHERE symbol='MNQ' AND open_time_ms >= $SINCE_MS
   AND contract<>'' AND contract<>'unrecomputable:spans_roll'
   AND COALESCE(source,'') NOT IN ('mixed','replay:off-scale')
  ORDER BY tf, contract, open_time_ms" > "$IN/bars.csv"

# 1b. HTF bars back to 2025-01-01 (the detector fetches 500 closed bars per tf; 1d needs ~2y).
sqlite3 -readonly -csv -header "$DB" "
  SELECT tf, contract, open_time_ms, o, h, l, c, v, COALESCE(source,'') source
  FROM bars WHERE symbol='MNQ' AND tf IN ('5m','15m','1h','4h','1d') AND open_time_ms >= $(( $(date -d '2025-01-01 00:00:00 CST' +%s) * 1000 ))
   AND contract<>'' AND contract<>'unrecomputable:spans_roll'
   AND COALESCE(source,'') NOT IN ('mixed','replay:off-scale')
  ORDER BY tf, contract, open_time_ms" > "$IN/bars_htf.csv"

# 2. plans (every version = one planner read), 30-date window.
sqlite3 -readonly -json "$DB" "
  SELECT plan_id, version, strategy_id, trade_date, session, trigger_reason, lifecycle,
         model_id, created_at, degraded, doc
  FROM plans WHERE trade_date >= '2026-08-15' ORDER BY trade_date, session, version" > "$IN/plans.json"

# 3. lifecycle log + recorded replan counters + strategy knobs (no secrets in these rows).
sqlite3 -readonly -json "$DB" "SELECT * FROM plan_lifecycle_log ORDER BY id" > "$IN/plan_lifecycle_log.json"
sqlite3 -readonly -json "$DB" "SELECT key, value FROM system_config WHERE key LIKE 'dayplan_replans_used:%' OR key LIKE 'dayplan_reset_baseline:%' ORDER BY key" > "$IN/replan_counters.json"
sqlite3 -readonly -json "$DB" "
  SELECT id, name, is_active, updated_at,
         json_extract(config,'$.day_plan') day_plan
  FROM strategies" > "$IN/strategy_dayplan.json"

# 4. research store: the planner's recorded input snapshot per read (object=plan).
sqlite3 -readonly -json "$RDB" "
  SELECT id, event, snapshot_id, captured_ms, fields_json
  FROM research_facts
  WHERE captured_ms >= $(( $(date -d '2026-09-01 00:00:00 CDT' +%s) * 1000 )) AND object='plan'
  ORDER BY id" > "$IN/research_plan.json"   # captured_ms bound → (captured_ms,object) index range; object=plan first appears 2026-09-09

cd "$IN" && sha256sum bars.csv bars_htf.csv plans.json plan_lifecycle_log.json replan_counters.json strategy_dayplan.json research_plan.json > SHA256SUMS
wc -l bars.csv; ls -la "$IN"; cat SHA256SUMS
