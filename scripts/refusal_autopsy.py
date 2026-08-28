#!/usr/bin/env python3
"""REFUSAL AUTOPSY (2026-08-27) — read-only replay on stored 1m bars."""
import json
import re
import sqlite3
from collections import defaultdict
from datetime import datetime, timezone, timedelta

CT = timezone(timedelta(hours=-5))
UTC = timezone.utc
DB = sqlite3.connect("file:/home/hoang/nofx/data/data.db?mode=ro", uri=True)
W0 = datetime(2026, 8, 26, 0, 0, tzinfo=CT)
NOW = datetime.now(CT)

bars = {}
for r in DB.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' AND open_time_ms >= ? ORDER BY open_time_ms",
                    (int(datetime(2026, 8, 25, 16, 0, tzinfo=CT).timestamp() * 1000),)):
    bars[r[0]] = r[1:]

def ctime(ms): return datetime.fromtimestamp(ms/1000, CT).strftime("%m-%d %H:%M")
def bar_at(t_ms): return bars.get(t_ms - t_ms % 60000)

def eod_for(t_ms):
    d = datetime.fromtimestamp(t_ms/1000, CT)
    flat = d.replace(hour=14, minute=45, second=0, microsecond=0)
    if d >= flat: flat += timedelta(days=1)
    return int(flat.timestamp()*1000)

def planned_r(side, e, sl, tp):
    risk, reward = abs(e-sl), abs(tp-e)
    return reward/risk if risk > 0 else 0.0

def replay(t_ms, side, entry, sl, tp):
    eod = eod_for(t_ms)
    keys = sorted(k for k in bars if t_ms-60000 < k <= eod)
    for k in keys:
        b = bars[k]
        hit_sl = b[2] <= sl if side == "long" else b[1] >= sl
        hit_tp = b[1] >= tp if side == "long" else b[2] <= tp
        if hit_sl and hit_tp:
            p = (sl-entry) if side == "long" else (entry-sl); return "AMBIG", p
        if hit_sl:
            p = (sl-entry) if side == "long" else (entry-sl); return "LOST", p
        if hit_tp:
            p = (tp-entry) if side == "long" else (entry-tp); return "WON", p
    last = bars[keys[-1]][3] if keys else entry
    p = (last-entry) if side == "long" else (entry-last)
    return "EOD", p

def verdict(n, won, lost, ambig, s):
    if n < 5: return "TOO-FEW"
    return "SAVING" if s < 0 else "COSTING"

census = []
for r in DB.execute("SELECT id, timestamp, decision_json, risk_check_error, cited_scenario_id FROM decision_records WHERE timestamp >= ? AND risk_check_error != '' AND risk_check_error != 'superseded_wait' ORDER BY id",
                    (datetime(2026, 8, 26, 5, 0, tzinfo=UTC).strftime("%Y-%m-%d %H:%M:%S"),)):
    rid, ts, dj, err, cited = r
    try: dec = json.loads(dj)[0]
    except Exception: continue
    if not dec.get("stop_loss") or not dec.get("take_profit"): continue
    t_ms = int(datetime.fromisoformat(ts).timestamp()*1000)
    if t_ms < int(W0.timestamp()*1000) or t_ms > int(NOW.timestamp()*1000): continue
    side = "long" if dec["action"] == "open_long" else "short"
    b = bar_at(t_ms); entry = b[3] if b else 0
    if not entry: continue
    gate = err.split(":")[0].replace("_refused","")
    census.append(dict(t=t_ms, side=side, entry=entry, sl=dec["stop_loss"], tp=dec["take_profit"], gate=gate, msg=err, src="risk_check_error"))

n_arm = 0
for r in DB.execute("SELECT ts_utc FROM log_events WHERE ts_utc >= ? AND message LIKE '%arm REFUSED%'",
                    (int(datetime(2026, 8, 26, 5, 0, tzinfo=UTC).timestamp()*1000),)):
    t_ms = r[0]
    if t_ms < int(W0.timestamp()*1000): continue
    census.append(dict(t=t_ms, side="short", entry=29611.41, sl=29638.0, tp=29557.15, gate="arm_RR", msg="arm REFUSED NY S1: R:R 2.04 below min 3.00", src="log_events"))
    n_arm += 1

print("=== CENSUS: %d refusals (%d decision-gate + %d arm-gate) ===" % (len(census), len(census)-n_arm, n_arm))
arm_seen = False; census_d = []
for c in census:
    if c["gate"] == "arm_RR":
        if arm_seen: continue
        arm_seen = True
    census_d.append(c)
census = census_d
per_gate = defaultdict(lambda: dict(n=0, won=0, lost=0, ambig=0, eod=0, s=0.0, rows=[]))
for c in census:
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    g = per_gate[c["gate"]]; g["n"] += 1
    if out == "WON": g["won"] += 1
    elif out == "AMBIG": g["ambig"] += 1; g["lost"] += 1
    elif out == "LOST": g["lost"] += 1
    else: g["eod"] += 1
    g["s"] += pnl*2.0
    g["rows"].append((c, out, pnl))
    print("  %s %-5s e=%.2f sl=%.2f tp=%.2f R=%.2f -> %-5s $%+.1f | %s" % (ctime(c["t"]), c["side"], c["entry"], c["sl"], c["tp"], planned_r(c["side"],c["entry"],c["sl"],c["tp"]), out, pnl*2, c["gate"]))

print("\n=== TABLE per gate ===")
print("%-16s %2s %4s %4s %5s %3s %8s  %s" % ("gate","n","won","lost","ambig","eod","S$","verdict"))
for gate in sorted(per_gate):
    g = per_gate[gate]
    v = verdict(g["n"], g["won"], g["lost"], g["ambig"], g["s"])
    print("%-16s %2d %4d %4d %5d %3d %8.1f  %s" % (gate, g["n"], g["won"], g["lost"], g["ambig"], g["eod"], g["s"], v))
print("\nTOTAL hypothetical S: $%+.1f" % sum(g["s"] for g in per_gate.values()))

print("\n=== HONEST-WAIT AUTOPSY ===")
declines = []
for r in DB.execute("SELECT id, timestamp, decision_json, system_prompt, cited_scenario_id, plan_id, plan_version FROM decision_records WHERE timestamp >= ? AND risk_check_error='superseded_wait' ORDER BY id",
                    (datetime(2026, 8, 26, 5, 0, tzinfo=UTC).strftime("%Y-%m-%d %H:%M:%S"),)):
    rid, ts, dj, sp, cited, pid, pv = r
    if not sp or "CONFLICT" in sp: continue
    mets = re.findall(r"(S\d+) confirm: .*? (MET|NOT MET)(?: \((stale|.*?)\))?", sp)
    fresh = [m[0] for m in mets if m[1] == "MET" and m[2] != "stale"]
    if not fresh: continue
    t_ms = int(datetime.fromisoformat(ts).timestamp()*1000)
    if t_ms < int(W0.timestamp()*1000) or t_ms > int(NOW.timestamp()*1000): continue
    row = DB.execute("SELECT doc FROM plans WHERE plan_id=? AND version=?", (pid, pv)).fetchone()
    if not row: continue
    doc = json.loads(row[0])
    for sid in fresh:
        sc = next((s for s in doc.get("scenarios",[]) if s.get("id") == sid), None)
        if not sc: continue
        conf = sc.get("confirm") or {}
        entry = conf.get("ref_price"); side = sc.get("direction","long")
        chain = sc.get("target_chain") or []
        nums = [float(x) for x in re.findall(r"[\d]+(?:[.][\d]+)?", sc.get("invalid","")) if float(x) > 1000]
        sl = nums[-1] if nums else None
        tp = chain[0] if chain else None
        if not entry or not sl or not tp: continue
        # side sanity + minimum risk (>=2pt): reject degenerate reconstructions
        if abs(sl - entry) < 2.0: continue
        if side == "long" and not (sl < entry < tp): continue
        if side == "short" and not (tp < entry < sl): continue
        declines.append(dict(t=t_ms, side=side, entry=entry, sl=sl, tp=tp, sid=sid, reason=(json.loads(dj)[0].get("reasoning") or "")[:70]))

seen_decl = set(); declines_d = []
for c in declines:
    key = (c["sid"], c["side"], c["entry"], c["sl"], c["tp"])
    if key in seen_decl: continue
    seen_decl.add(key); declines_d.append(c)
declines = declines_d
dagg = dict(n=0, won=0, lost=0, ambig=0, eod=0, s=0.0); decl_rows=[]
for c in declines:
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    if out == "AMBIG": dagg["ambig"] += 1; dagg["lost"] += 1
    else: dagg[out.lower()] += 1
    dagg["n"] += 1; dagg["s"] += pnl*2.0
    decl_rows.append((c, out, pnl))
    if c["sid"] == "S2":
        print("  ** S2 CASE: %s %s e=%.2f sl=%.2f tp=%.2f -> %s $%+.1f" % (ctime(c["t"]), c["side"], c["entry"], c["sl"], c["tp"], out, pnl*2))
print("fresh-MET declines reconstructable: %d" % len(declines))
print("AI-declines: n=%d won=%d lost=%d ambig=%d eod=%d S=$%+.1f" % (dagg["n"], dagg["won"], dagg["lost"], dagg["ambig"], dagg["eod"], dagg["s"]))
for c, out, pnl in decl_rows[:16]:
    print("  %s %-5s %s e=%.2f sl=%.2f tp=%.2f -> %-5s $%+.1f" % (ctime(c["t"]), c["side"], c["sid"], c["entry"], c["sl"], c["tp"], out, pnl*2))

print("\n=== R:R DEEP-DIVE ===")
all_c = []
for c in census:
    c2 = dict(c); c2["cond"] = c["gate"]; all_c.append(c2)
for c in declines:
    c2 = dict(c); c2["cond"] = c["sid"]; all_c.append(c2)
for c in all_c:
    c["R"] = planned_r(c["side"], c["entry"], c["sl"], c["tp"])
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    c["out"] = out; c["pnl"] = pnl*2.0
rr_rows = [c for c in all_c if c.get("R") and c["R"] > 0]
print("candidates with planned R: %d (realized-R distribution below)" % len(rr_rows))
for c in sorted(rr_rows, key=lambda x:-x["R"])[:24]:
    print("  %s %-5s %-10s R=%.2f -> %-5s $%+.1f" % (ctime(c["t"]), c["side"], c["cond"], c["R"], c["out"], c["pnl"]))
for thr in (2.0, 2.5, 3.0):
    sel = [c for c in rr_rows if c["R"] >= thr]
    s = sum(c["pnl"] for c in sel); w = sum(1 for c in sel if c["out"] == "WON")
    print("  accept-if-R>=%.1f: n=%d won=%d S=$%+.1f" % (thr, len(sel), w, s))
print("per condition-type at thresholds:")
conds = sorted({c["cond"] for c in rr_rows})
for thr in (2.0, 2.5, 3.0):
    line = []
    for cd in conds:
        sel = [c for c in rr_rows if c["cond"] == cd and c["R"] >= thr]
        s = sum(c["pnl"] for c in sel)
        line.append("%s:n%d/S%+.0f" % (cd, len(sel), s))
    print("  R>=%.1f: %s" % (thr, " | ".join(line)))

print("\n=== CAVEAT COUNTS ===")
print("refusals=%d declines=%d candidates=%d ambig=%d eod=%d" % (
    len(census), len(declines), len(all_c),
    sum(1 for c in all_c if c["out"]=="AMBIG"),
    sum(1 for c in all_c if c["out"]=="EOD")))
