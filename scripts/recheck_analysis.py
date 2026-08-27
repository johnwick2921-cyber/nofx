#!/usr/bin/env python3
"""Master-recheck independent analysis (2026-08-27) — READ-ONLY.
Recomputes detector/grade/pnl facts directly from the SQLite DB with its own
logic (never calls kernel functions under test)."""
import json, math, sqlite3, sys
from datetime import datetime, timezone, timedelta

DB = sqlite3.connect("file:/home/hoang/nofx/data/data.db?mode=ro", uri=True)
CT = timezone(timedelta(hours=-5))

def q(sql, *a):
    return DB.execute(sql, a).fetchall()

print("=== DB sanity ===")
print("bars rows:", q("SELECT COUNT(*), MIN(open_time_ms), MAX(open_time_ms) FROM bars")[0])
print("dup keys:", q("SELECT COUNT(*) FROM (SELECT 1 FROM bars GROUP BY symbol,tf,open_time_ms HAVING COUNT(*)>1)")[0])

# ---------- S2.1 census from LONDON 08-27 v7 doc ----------
doc = json.load(open('tmp/london_v7.json'))
levels = doc.get('levels', [])
print("\n=== 2.1 census LONDON v7 (%d levels) ===" % len(levels))
kinds = {}
for l in levels:
    lbl = l['label']
    # TF parse: "EQL·15m (HTF)" -> 15m HTF ; "Supply·1h" -> 1h
    import re
    m = re.search(r'·(\d+m|1h|4h|D)', lbl)
    tf = m.group(1) if m else '1m'
    ht = ' (HTF)' if 'HTF' in lbl else ''
    k = lbl.split('·')[0] + '·' + tf + ht
    kinds[k] = kinds.get(k, 0) + 1
for k, v in sorted(kinds.items()):
    print(f"  {k}: {v}")
print("seated prices:", sorted(round(l['price'], 2) for l in levels))

# ---------- S2.2 recompute anchors from 1m bars ----------
def bars_range(lo_ms, hi_ms):
    return q("SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol='MNQ' AND tf='1m' AND open_time_ms>=? AND open_time_ms<? ORDER BY open_time_ms", lo_ms, hi_ms)

def day_str(d):
    return d.strftime('%Y-%m-%d')

# LONDON session 08-27: 02:00-07:00 CT (registry LONDON). Verify from bars density later.
def ct(ms): return datetime.fromtimestamp(ms/1000, CT)
def ct_ts(y, mo, d, hh, mm=0):
    return int(datetime(y, mo, d, hh, mm, tzinfo=CT).timestamp()*1000)

print("\n=== 2.2 independent recompute (LONDON 2026-08-27, 02:00-07:00 CT) ===")
start, end = ct_ts(2026,8,27,2,0), ct_ts(2026,8,27,7,0)
bars = bars_range(start, end)
print("bars fetched:", len(bars))
if bars:
    hi = max(b[2] for b in bars); lo = min(b[3] for b in bars)
    # PDC = prior day RTH 15:30-16:00? use 08-26 08:30-15:00 CME RTH close = last bar close of 08-26 RTH
    prev_end = ct_ts(2026,8,26,15,0)
    prev_bars = bars_range(ct_ts(2026,8,26,8,30), prev_end)
    pdc = prev_bars[-1][4] if prev_bars else None
    # PDH/PDL from 08-26 RTH
    pdh = max(b[2] for b in prev_bars) if prev_bars else None
    pdl = min(b[3] for b in prev_bars) if prev_bars else None
    # ONH/ONL: overnight 08-26 15:00 -> 08-27 02:00? use 17:00 CT -> 02:00
    on = bars_range(ct_ts(2026,8,26,17,0), start)
    onh = max(b[2] for b in on) if on else None
    onl = min(b[3] for b in on) if on else None
    # RTH-H/L for the session: 08-27 08:30+ is not in LONDON window; RTH-H/L for LONDON = session's own H/L? Use OR period.
    # OR = first 15m? first 30m? use first 30m (02:00-02:30) hi/lo; IB = first hour
    or30 = bars_range(start, start + 30*60000)
    or_h = max(b[2] for b in or30) if or30 else None
    or_l = min(b[3] for b in or30) if or30 else None
    ib = bars_range(start, start + 60*60000)
    ib_h = max(b[2] for b in ib) if ib else None
    ib_l = min(b[3] for b in ib) if ib else None
    # session VWAP: typical price * volume
    tp_sum = sum((b[2]+b[3]+b[4])/3*b[5] for b in bars)
    v_sum = sum(b[5] for b in bars)
    vwap = tp_sum/v_sum if v_sum else None
    print(f"PDC={pdc} PDH={pdh} PDL={pdl}")
    print(f"ONH={onh} ONL={onl}")
    print(f"OR(30m) H/L={or_h}/{or_l} IB(1h) H/L={ib_h}/{ib_l}")
    print(f"session VWAP (typical×vol)={vwap}  (bars session H/L {hi}/{lo})")
    # anchors the plan seated?
    seated = [round(l['price'],2) for l in levels]
    for name, v in [('PDC',pdc),('PDH',pdh),('PDL',pdl),('ONH',onh),('ONL',onl)]:
        hit = 'SEATED' if v is not None and any(abs(s-v)<=1.0 for s in seated) else ('near ' + str(min((abs(s-v), s) for s in seated)) if v else '')
        print(f"  {name}={v} vs plan: {hit}")

# ---------- S2.3 VWAP divergence ----------
print("\n=== 2.3 VWAP divergence ===")
vwap_plan = [l for l in levels if 'VWAP' in l['label']]
print("plan VWAP rows:", [(l['label'], l['price'], l.get('machine_grade')) for l in vwap_plan])

# ---------- S2.5 nPOC ----------
npoc = [l for l in levels if 'nPOC' in l['label'] or 'POC' in l['label']]
print("\n=== 2.5 nPOC seated:", len(npoc), [(l['label'], l['price']) for l in npoc])
print("level_state POC rows:", q("SELECT level_key, times_tested, consumed, freshness FROM level_state WHERE trader_id LIKE '8d5c8af5%' AND (level_type LIKE '%POC%') ORDER BY updated_at DESC LIMIT 5"))

# ---------- S2.6 missed turns (last 3 sessions) ----------
print("\n=== 2.6 missed-turns % (5m fractal swings vs seated, ±8pts) ===")
def session_window(datestr, sess):
    # crude session bounds
    d = datetime.strptime(datestr, '%Y-%m-%d')
    if sess == 'LONDON': lo, hi = 2, 7
    elif sess == 'NY': lo, hi = 8, 16
    elif sess == 'ASIA': lo, hi = 18, 23
    return ct_ts(d.year, d.month, d.day, lo), ct_ts(d.year, d.month, d.day, hi)

def last_plan(pid_prefix):
    rows = q("SELECT plan_id, version, doc FROM plans WHERE plan_id LIKE ? ORDER BY version DESC LIMIT 1", pid_prefix + '%')
    return rows[0] if rows else None

def swing_turns(b5m):
    turns = []
    for i in range(2, len(b5m)-2):
        c = b5m[i]
        if c[2] >= max(x[2] for x in b5m[i-2:i+3]): turns.append(('H', c[2]))
        if c[3] <= min(x[3] for x in b5m[i-2:i+3]): turns.append(('L', c[3]))
    return turns

sessions = [('2026-08-27','LONDON'), ('2026-08-26','NY'), ('2026-08-26','LONDON')]
for ds, sess in sessions:
    lo, hi = session_window(ds, sess)
    b5 = q("SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol='MNQ' AND tf='5m' AND open_time_ms>=? AND open_time_ms<? ORDER BY open_time_ms", lo, hi)
    if not b5: b5 = []
    turns = swing_turns(b5)
    p = last_plan(f'{ds}:{sess}:8d5c8af5%')
    if not p:
        print(f"  {ds} {sess}: no plan"); continue
    doc = json.loads(p[2]); seated = [l['price'] for l in doc.get('levels', [])]
    missed = 0
    for _, px in turns:
        if not any(abs(px-s) <= 8.0 for s in seated): missed += 1
    pct = 100.0*missed/len(turns) if turns else 0
    print(f"  {ds} {sess} v{p[1]}: swings={len(turns)} missed(>8pt)={missed} ({pct:.1f}%) seated={len(seated)}")

# ---------- S3.4 stamp gap ----------
print("\n=== 3.4 stamp gap (machine_grade empty) ===")
for ds, sess in sessions:
    p = last_plan(f'{ds}:{sess}:8d5c8af5%')
    if not p: continue
    doc = json.loads(p[2]); lv = doc.get('levels', [])
    empty = [l['label'] for l in lv if not l.get('machine_grade')]
    print(f"  {ds} {sess} v{p[1]}: {len(empty)}/{len(lv)} unstamped {empty[:6]}")

# ---------- S3.5 reaction-by-grade from level_stats ----------
print("\n=== 3.5 level_stats reaction-by-grade (28 rows, 2026-08-25 session) ===")
rows = q("SELECT * FROM level_stats")
print("cols:", [c[0] for c in DB.execute('PRAGMA table_info(level_stats)')])
from collections import defaultdict
agg = defaultdict(lambda: [0,0,0,0])  # grade -> [touch, react, broke, chop]
for r in rows:
    # schema assumed: trader,date,price,label,kind,grade,role,src,?,touch,react,broke,chop,ts (verify below)
    pass
print("total rows:", len(rows))

# ---------- S7.4 pnl recompute ----------
print("\n=== 7.4 pnl_corrected recompute (newest 5 closed) ===")
print("fills cols:", [c[0] for c in DB.execute('PRAGMA table_info(trader_fills)')][:12])

# ---------- S7.7 expectancy last 7 days ----------
print("\n=== 7.7 closed positions last 7 days ===")
closed = q("SELECT id, side, plan_id, pnl_corrected, pnl_realized, entry_price, exit_price FROM trader_positions WHERE status='CLOSED' AND updated_at > strftime('%s','now')*1000 - 7*86400000 ORDER BY id")
print("n:", len(closed))
for r in closed[:6]: print("  ", r)
