#!/usr/bin/env python3
"""S3 (mega-research 2026-08-26) — one-shot, idempotent backfill of
trader_positions.plan_id / plan_trade_date / plan_session / plan_link_note.

Reconstructs the (date, session, trader, version) join for every closed
position from 2026-08-19 UTC onward. NEVER guesses:
  - exactly one matching plan row  -> stamp plan_id/date/session, note 'backfill:reconstructed'
  - 'off-plan' cite                -> plan_id='UNRESOLVABLE', note 'unresolvable:off-plan-cite'
  - plan_version=0                 -> plan_id='UNRESOLVABLE', note 'unresolvable:no-version'
  - no matching plan row           -> plan_id='UNRESOLVABLE', note 'unresolvable:no-plan-row'
Idempotent: rows already stamped (plan_id != '' and != 'UNRESOLVABLE') or
already noted are skipped. Run: python3 scripts/backfill-position-plan.py [db]
"""
import json
import shutil
import sqlite3
import sys
import time
from datetime import datetime, timedelta
from zoneinfo import ZoneInfo

DB = sys.argv[1] if len(sys.argv) > 1 else "data/data.db"
CHI = ZoneInfo("America/Chicago")
CUTOFF = int(datetime(2026, 8, 19, tzinfo=CHI).timestamp() * 1000)

db = sqlite3.connect(DB)
db.row_factory = sqlite3.Row

def session_of(ms):
    t = datetime.fromtimestamp(ms / 1000, CHI)
    hm = t.hour * 60 + t.minute
    if t.hour >= 17 or t.hour < 2:
        return "ASIA"
    if hm < 510:
        return "LONDON"
    return "NY"

def session_date_of(ms):
    t = datetime.fromtimestamp(ms / 1000, CHI)
    if t.hour < 17:
        t = t - timedelta(days=1)
    return t.strftime("%Y-%m-%d")

def cols():
    return {r["name"] for r in db.execute("PRAGMA table_info(trader_positions)")}

# 1. additive schema (idempotent — the new binary's AutoMigrate is a no-op after this)
have = cols()
for name, ddl in [
    ("plan_id", "TEXT DEFAULT ''"),
    ("plan_trade_date", "TEXT DEFAULT ''"),
    ("plan_session", "TEXT DEFAULT ''"),
    ("plan_link_note", "TEXT DEFAULT ''"),
]:
    if name not in have:
        db.execute(f"ALTER TABLE trader_positions ADD COLUMN {name} {ddl}")
        print(f"added column {name}")
db.commit()

# 2. backfill loop
positions = db.execute(
    """SELECT id, trader_id, entry_time, cited_scenario_id, plan_version, plan_id, plan_link_note
       FROM trader_positions
       WHERE status='CLOSED' AND exit_time IS NOT NULL AND exit_time >= ?""",
    (CUTOFF,),
).fetchall()

plans = {}
for r in db.execute("SELECT plan_id, version, strategy_id, trade_date, session FROM plans"):
    plans[(r["trade_date"], r["session"], r["strategy_id"], r["version"])] = r["plan_id"]

resolved = unresolvable = skipped = 0
reasons = {}
for p in positions:
    pid, note = p["plan_id"], p["plan_link_note"]
    if pid and pid != "UNRESOLVABLE":
        skipped += 1
        continue
    if note:
        skipped += 1
        continue
    td, sess = session_date_of(p["entry_time"]), session_of(p["entry_time"])
    cite = (p["cited_scenario_id"] or "").strip()
    verdict, reason, plan_id, d, s = None, None, "", "", ""
    if cite.lower() == "off-plan":
        verdict, reason = "UNRESOLVABLE", "unresolvable:off-plan-cite"
    elif p["plan_version"] <= 0:
        verdict, reason = "UNRESOLVABLE", "unresolvable:no-version"
    else:
        key = (td, sess, p["trader_id"], p["plan_version"])
        if key in plans:
            verdict, reason = plans[key], "backfill:reconstructed"
            plan_id, d, s = plans[key], td, sess
        else:
            verdict, reason = "UNRESOLVABLE", "unresolvable:no-plan-row"
    if verdict == "UNRESOLVABLE":
        db.execute(
            "UPDATE trader_positions SET plan_id='UNRESOLVABLE', plan_link_note=? WHERE id=?",
            (reason, p["id"]),
        )
        unresolvable += 1
    else:
        db.execute(
            "UPDATE trader_positions SET plan_id=?, plan_trade_date=?, plan_session=?, plan_link_note=? WHERE id=?",
            (plan_id, d, s, reason, p["id"]),
        )
        resolved += 1
    reasons[reason] = reasons.get(reason, 0) + 1

db.commit()
print(f"backfill complete: resolved={resolved} unresolvable={unresolvable} skipped={skipped}")
for k, v in sorted(reasons.items()):
    print(f"  {k}: {v}")
db.close()
