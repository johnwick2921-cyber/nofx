#!/usr/bin/env python3
"""Grand audit 2026-08-28 — DB probes (Part A), READ-ONLY.

Usage: python3 scripts/grand_audit_probes.py [/path/to/data.db]
Opens SQLite in mode=ro. Reproduces:
  A13 dup census + open-stamp convention + NULL census
  A14 pnl_corrected delta / NULL census (post-window)
  A15 level_stats rows by session_day
  A16 touch_episodes census
  A18 stamp-unstamped JSON scan
"""
import json
import sqlite3
import sys
import time

DB_PATH = sys.argv[1] if len(sys.argv) > 1 else "/home/hoang/nofx/data/data.db"
uri = "file:%s?mode=ro" % DB_PATH
db = sqlite3.connect(uri, uri=True)
now_ms = int(time.time() * 1000)
WINDOW = 1787900400000  # 2026-08-28 02:00 CT approx (post-deploy closes)

def q1(sql, args=()):
    return db.execute(sql, args).fetchone()[0]

print("=== A13 bars integrity ===")
print("dup_groups =", q1("SELECT COUNT(*) FROM (SELECT 1 FROM bars "
                        "GROUP BY symbol,tf,open_time_ms HAVING COUNT(*)>1)"))
print("null_open =", q1("SELECT COUNT(*) FROM bars WHERE open_time_ms IS NULL OR open_time_ms=0"))
print("open_stamp_ok =", q1("SELECT COUNT(*) FROM (SELECT open_time_ms FROM bars "
                           "ORDER BY open_time_ms DESC LIMIT 30) WHERE open_time_ms % 60000 = 0"),
      "/ 30")
print("tfs =", [r[0] for r in db.execute("SELECT DISTINCT tf FROM bars")])

print("\n=== A14 pnl_corrected ===")
print("null_pnl_post =", q1("SELECT COUNT(*) FROM trader_positions "
                           "WHERE status='CLOSED' AND pnl_corrected IS NULL AND exit_time >= ?",
                           (WINDOW,)))
print("max_delta = %.2f" % q1("SELECT COALESCE(MAX(ABS(pnl_corrected-realized_pnl)),0) FROM "
                              "trader_positions WHERE status='CLOSED' AND pnl_corrected IS NOT NULL "
                              "AND exit_time >= ?", (WINDOW,)))

print("\n=== A15 level_stats by session_day ===")
for row in db.execute("SELECT session_day, COUNT(*) FROM level_stats "
                      "GROUP BY session_day ORDER BY session_day"):
    print(" ", row[0], row[1])

print("\n=== A16 touch_episodes ===")
print("count =", q1("SELECT COUNT(*) FROM touch_episodes WHERE opened_at_ms >= ?", (WINDOW - 86400000,)),
      "· max_open =", q1("SELECT COALESCE(MAX(opened_at_ms),0) FROM touch_episodes"))

print("\n=== A18 stamp-unstamped (JSON scan) ===")
bad = []
for rowid, pid, ver, doc in db.execute("SELECT rowid, plan_id, version, doc FROM plans "
                                       "WHERE rowid > 130 ORDER BY rowid"):
    try:
        d = json.loads(doc)
    except Exception:
        bad.append((rowid, pid[:20], ver, "UNPARSEABLE"))
        continue
    un = [l.get("price") for l in (d.get("levels") or []) if not l.get("grade")]
    if un:
        bad.append((rowid, pid[12:26], ver, un[:4]))
print("plans_with_ungraded_levels =", len(bad))
for b in bad[:10]:
    print(" ", b)
