#!/usr/bin/env python3
"""Master-recheck independent analysis v2 (2026-08-27) — READ-ONLY.
Own recomputation from DB rows; never calls kernel funcs under test."""
import json, math, sqlite3
from datetime import datetime, timezone, timedelta
from collections import defaultdict

DB = sqlite3.connect("file:/home/hoang/nofx/data/data.db?mode=ro", uri=True)
CT = timezone(timedelta(hours=-5))

def q(sql, *a): return DB.execute(sql, a).fetchall()
def ct_ts(y, mo, d, hh, mm=0): return int(datetime(y, mo, d, hh, mm, tzinfo=CT).timestamp()*1000)
def doc_of(pid_prefix):
    r = q("SELECT plan_id, version, doc FROM plans WHERE plan_id LIKE ? ORDER BY version DESC LIMIT 1", pid_prefix+'%')
    return r[0] if r else None

def m1(lo, hi):
    return q("SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol='MNQ' AND tf='1m' AND open_time_ms>=? AND open_time_ms<? ORDER BY open_time_ms", lo, hi)

def m5_from_m1(lo, hi):
    b = m1(lo, hi)
    out = []
    for i in range(0, len(b)-4, 5):
        w = b[i:i+5]
        out.append((w[0][0], w[0][1], max(x[2] for x in w), min(x[3] for x in w), w[-1][4], sum(x[5] for x in w)))
    return out

print("=== 2.1 label->price LONDON v7 ===")
p = doc_of('2026-08-27:LONDON:8d5c8af5')
doc = json.loads(p[2])
for l in doc['levels']:
    print(f"  {l['label']:22s} {l['price']:>9} grade={l.get('grade')} mg={l.get('machine_grade','')}")

print("\n=== 2.2 recompute LONDON 2026-08-27 (02:00-08:30 CT) ===")
lo, hi = ct_ts(2026,8,27,2,0), ct_ts(2026,8,27,8,30)
bars = m1(lo, hi)
print("1m bars:", len(bars))
prev = m1(ct_ts(2026,8,26,8,30), ct_ts(2026,8,26,15,0))
pdc = prev[-1][4]; pdh = max(b[2] for b in prev); pdl = min(b[3] for b in prev)
on = m1(ct_ts(2026,8,26,17,0), ct_ts(2026,8,27,2,0))
onh = max(b[2] for b in on); onl = min(b[3] for b in on)
or30 = m1(lo, lo+30*60000); or_h = max(b[2] for b in or30); or_l = min(b[3] for b in or30)
ib = m1(lo, lo+60*60000); ib_h = max(b[2] for b in ib); ib_l = min(b[3] for b in ib)
tp = sum((b[2]+b[3]+b[4])/3*b[5] for b in bars); vs = sum(b[5] for b in bars)
vwap = tp/vs if vs else None
print(f"PDC={pdc} PDH={pdh} PDL={pdl} ONH={onh} ONL={onl}")
print(f"OR30 H/L={or_h}/{or_l} IB60 H/L={ib_h}/{ib_l}")
print(f"session VWAP(typical*vol, 02:00-08:30)={vwap:.2f}")
plan_vwap1s = [l['price'] for l in doc['levels'] if 'VWAP' in l['label']]
print("plan VWAP-1sigma price:", plan_vwap1s)
if vwap and plan_vwap1s:
    print(f"  implied sigma = {vwap-plan_vwap1s[0]:.2f}pt (VWAP - VWAP-1σ)")

print("\n=== 2.6 missed turns (5m from 1m, last 3 sessions) ===")
def swings(b5):
    t = []
    for i in range(2, len(b5)-2):
        hi = max(x[2] for x in b5[i-2:i+3]); lo = min(x[3] for x in b5[i-2:i+3])
        if b5[i][2] >= hi: t.append(('H', b5[i][2]))
        if b5[i][3] <= lo: t.append(('L', b5[i][3]))
    return t
def win(ds, sess):
    d = datetime.strptime(ds, '%Y-%m-%d')
    if sess=='LONDON': return ct_ts(d.year,d.month,d.day,2), ct_ts(d.year,d.month,d.day,8,30)
    if sess=='NY': return ct_ts(d.year,d.month,d.day,8,30), ct_ts(d.year,d.month,d.day,16,0)
    if sess=='ASIA': return ct_ts(d.year,d.month,d.day,17), ct_ts(d.year,d.month,d.day,23,59)
for ds, sess in [('2026-08-27','LONDON'),('2026-08-26','NY'),('2026-08-26','LONDON')]:
    lo, hi = win(ds, sess)
    b5 = m5_from_m1(lo, hi)
    p = doc_of(f'{ds}:{sess}:8d5c8af5')
    if not p: print(f"  {ds} {sess}: no plan"); continue
    doc = json.loads(p[2]); seated = [l['price'] for l in doc.get('levels',[])]
    t = swings(b5); missed = sum(1 for _,px in t if not any(abs(px-s)<=8 for s in seated))
    print(f"  {ds} {sess} v{p[1]}: 5m bars={len(b5)} swings={len(t)} missed={missed} ({100*missed/max(1,len(t)):.1f}%) seated={len(seated)}")

print("\n=== 2.5 nPOC seats in recent plans ===")
for row in q("SELECT plan_id, version FROM plans WHERE plan_id LIKE '2026-08-2%' ORDER BY created_at DESC LIMIT 12"):
    d = json.loads(q("SELECT doc FROM plans WHERE plan_id=? AND version=?", row[0], row[1])[0][0])
    np = [l for l in d.get('levels',[]) if 'POC' in l['label']]
    print(f"  {row[0]} v{row[1]}: nPOC rows = {[(l['label'],l['price']) for l in np]}")

print("\n=== 3.5 reaction by grade (level_stats, session 2026-08-25) ===")
agg = defaultdict(lambda: [0,0,0,0,0])
for r in q("SELECT grade, touched, reacted, broke_clean, chopped FROM level_stats"):
    g = r[0]; agg[g][0]+=1; agg[g][1]+=r[1]; agg[g][2]+=r[2]; agg[g][3]+=r[3]; agg[g][4]+=r[4]
for g in sorted(agg):
    n, t, r2, b, c = agg[g]
    print(f"  {g}: n={n} touched={t} reacted={r2} ({100*r2/max(1,t):.0f}%) broke={b} chop={c}")

print("\n=== 3.4 stamp gap across last plans ===")
for row in q("SELECT plan_id, version, lifecycle FROM plans WHERE plan_id LIKE '2026-08-2%' AND lifecycle='active' ORDER BY created_at DESC LIMIT 6"):
    d = json.loads(q("SELECT doc FROM plans WHERE plan_id=? AND version=?", row[0], row[1])[0][0])
    lv = d.get('levels',[])
    empty = [l['label'] for l in lv if not l.get('machine_grade')]
    print(f"  {row[0]} v{row[1]}: {len(empty)}/{len(lv)} unstamped {empty[:5]}")

print("\n=== 7.4 pnl recompute (newest 5 closed) ===")
rows = q("SELECT id, entry_order_id, exit_order_id, entry_price, exit_price, entry_quantity, pnl_corrected, pnl_correction_note, close_reason FROM trader_positions WHERE status='CLOSED' ORDER BY id DESC LIMIT 5")
for r in rows:
    pid, eo, xo, ep, xp, qt, pnl, note, cr = r
    fills = q("SELECT order_id, side, price, quantity, realized_pnl FROM trader_fills WHERE order_id IN (?,?)", eo or '', xo or '')
    sum_f = sum(f[4] or 0 for f in fills)
    calc = (xp-ep)*qt*2 if (ep and xp and qt) else None   # MNQ $2/pt
    print(f"  pos {pid}: entry={ep} exit={xp} qty={qt} calc(×2)={calc} stored_pnl_corrected={pnl} fills_realized={sum_f} reason={cr} note={note}")

print("\n=== 7.7 expectancy last 7 days by scenario condition ===")
# join positions -> plan doc scenario via cited_scenario_id
rows = q("SELECT id, plan_id, plan_version, cited_scenario_id, pnl_corrected, plan_band FROM trader_positions WHERE status='CLOSED' AND updated_at > strftime('%s','now')*1000 - 7*86400000 ORDER BY id")
cond = defaultdict(lambda: [0,0.0,0])
for r in rows:
    pid, pv, sid, pnl, band = r[1], r[2], r[3], r[4] or 0, r[5]
    key = 'UNRESOLVABLE' if not pid else 'no_scenario'
    if pid and sid:
        d = q("SELECT doc FROM plans WHERE plan_id=? AND version=?", pid, pv)
        if d:
            doc = json.loads(d[0][0])
            sc = next((s for s in doc.get('scenarios',[]) if s.get('id')==sid), None)
            if sc: key = (sc.get('condition') or '?') + ('·' + sc.get('direction','') if sc.get('direction') else '')
    cond[key][0]+=1; cond[key][1]+=pnl; cond[key][2]+=1 if pnl>0 else 0
for k in sorted(cond, key=lambda k: -cond[k][0]):
    n, s, w = cond[k]
    print(f"  {k:22s} n={n:2d} Σ={s:8.1f} win={w}")
