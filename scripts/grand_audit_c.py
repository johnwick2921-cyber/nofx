#!/usr/bin/env python3
"""GRAND AUDIT 2026-08-28 — PART C engine (read-only, report-only, zero config changes).

C1 swing-k + TF mix missed-turns   C2 FVG displacement floor census
C3 HTF-veto TF 1h vs 4h replay     C4 min-conf 60 vs 65 band
C5 trail-mult 2.0/2.5/3.0 sweep    C6 proximity band sweep

Usage: python3 scripts/grand_audit_c.py
"""
import json
import sqlite3
from collections import defaultdict
from datetime import datetime, timezone, timedelta

CT = timezone(timedelta(hours=-5))
DB = sqlite3.connect("file:/home/hoang/nofx/data/data.db?mode=ro", uri=True)

# ---- bars ----
bars1 = {}  # ms -> (o,h,l,c)
for r in DB.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' ORDER BY open_time_ms"):
    bars1[r[0]] = r[1:]
keys1 = sorted(bars1)

def agg_tf(ms_list, mins):
    out = []
    step = mins * 60000
    for i in range(0, len(ms_list) - mins + 1, mins):
        w = ms_list[i:i + mins]
        out.append((w[0], max(bars1[k][1] for k in w), min(bars1[k][2] for k in w)))
    return out

def wilder_atr(b5, period=14):
    if len(b5) < period + 1:
        return 0.0
    trs = []
    for i in range(1, len(b5)):
        h, l = b5[i][1], b5[i][2]
        pc = bars1[b5[i - 1][0]][3]
        trs.append(max(h - l, abs(h - pc), abs(l - pc)))
    a = sum(trs[:period]) / period
    for i in range(period, len(trs)):
        a = (a * (period - 1) + trs[i]) / period
    return a

def session_of(ms):
    d = datetime.fromtimestamp(ms / 1000, CT)
    return d.strftime("%Y-%m-%d") + (" LONDON" if d.hour < 8 else (" NY" if d.hour < 15 else " ASIA"))

# ============================================================
# C1 — swing-k / TF mix missed-turns
# ============================================================
def swings(b5, k):
    out = []
    for i in range(k, len(b5) - k):
        h, l = b5[i][1], b5[i][2]
        ok_h = all(b5[j][1] <= h for j in range(i - k, i + k + 1) if j != i)
        ok_l = all(b5[j][2] >= l for j in range(i - k, i + k + 1) if j != i)
        if ok_h and ok_l:
            if abs(h - l) < 4.0:  # ambiguous micro-fractal
                continue
        if ok_h:
            out.append((b5[i][0], h, "H"))
        if ok_l:
            out.append((b5[i][0], l, "L"))
    return out

def seated_levels(session):
    row = DB.execute("SELECT doc FROM plans WHERE session=? AND lifecycle != 'no_trade' "
                     "ORDER BY created_at DESC LIMIT 1", (session,)).fetchone()
    if not row:
        return []
    d = json.loads(row[0])
    return [l["price"] for l in d.get("levels", []) if l.get("grade")]

def missed(b5, sw, seats, band=8.0):
    miss = 0
    for _, px, _ in sw:
        if not any(abs(px - s) <= band for s in seats):
            miss += 1
    return miss

print("== C1 swing-k × TF missed-turns (band ±8pt, vs each session's actual seats) ==")
sessions = ["LONDON", "NY", "ASIA"]
print("%-16s %-5s %-4s %8s %8s %8s" % ("session", "tf", "k", "n_swings", "missed", "miss%"))
for sess in sessions:
    lo = None
    for d in (26, 27, 28):
        lo = int(datetime(2026, 8, d, 0, 0, tzinfo=CT).timestamp() * 1000)
        hi = lo + 86400000
        sub = [k for k in keys1 if lo <= k < hi]
        if len(sub) > 500:
            break
    if lo is None:
        continue
    seats = seated_levels(sess)
    for mins in (5, 15):
        b5 = agg_tf(sub, mins)
        if len(b5) < 20:
            continue
        for k in (2, 3):
            sw = swings(b5, k)
            m = missed(b5, sw, seats)
            print("%-16s %-5s %-4d %8d %8d %7.0f%%" % (sess, "%dm" % mins, k, len(sw), m, 100 * m / max(len(sw), 1)))

# ============================================================
# C2 — FVG displacement floor census (1.25 vs 1.5 × ATR5m)
# ============================================================
print("\n== C2 FVG displacement floor census (5m bars, gap >= 2pt) ==")
print("%-9s %6s %6s" % ("day", "1.25x", "1.5x"))
for d in (24, 25, 26, 27, 28):
    lo = int(datetime(2026, 8, d, 0, 0, tzinfo=CT).timestamp() * 1000)
    hi = lo + 86400000
    sub = [k for k in keys1 if lo <= k < hi]
    if len(sub) < 300:
        print("%-9s (insufficient bars n=%d)" % ("%02d" % d, len(sub)))
        continue
    b5 = agg_tf(sub, 5)
    atr = wilder_atr(b5, 14)
    c_lo = c_hi = 0
    for i in range(2, len(b5)):
        # bull FVG: gap between bar[i-2].high and bar[i].low (price jumps up)
        # bear FVG: gap between bar[i-2].low and bar[i].high
        imp = bars1[b5[i - 1][0]]
        body = abs(imp[3] - imp[0])
        if b5[i][2] - b5[i - 2][1] >= 2.0:   # bull gap >= 2pt floor
            if body >= 1.25 * atr:
                c_lo += 1
            if body >= 1.5 * atr:
                c_hi += 1
        if b5[i - 2][2] - b5[i][1] >= 2.0:   # bear gap >= 2pt floor
            if body >= 1.25 * atr:
                c_lo += 1
            if body >= 1.5 * atr:
                c_hi += 1
    print("%s  %6d %6d" % ("08-%02d" % d, c_lo, c_hi))

# ============================================================
# C3 — HTF veto TF 1h vs 4h replay
# ============================================================
print("\n== C3 HTF-veto TF 1h vs 4h (vetoed arms replayed under each TF) ==")
def tf_trend(t_ms, mins, k=2):
    lo = t_ms - mins * 60000 * 40
    sub = [x for x in keys1 if lo <= x < t_ms]
    if len(sub) < mins * 30:
        return "RANGING"
    b = agg_tf(sub, mins)
    sw = swings(b, k)
    if len(sw) < 4:
        return "RANGING"
    a, b2, c, d2 = sw[-4], sw[-3], sw[-2], sw[-1]
    if a[2] == "H" and b2[2] == "L" and c[2] == "H" and d2[2] == "L":
        if d2[1] > b2[1] and c[1] > a[1]:
            return "TRENDING_UP"
        if d2[1] < b2[1] and c[1] < a[1]:
            return "TRENDING_DOWN"
    if a[2] == "L" and b2[2] == "H" and c[2] == "L" and d2[2] == "H":
        if c[1] > a[1] and d2[1] > b2[1]:
            return "TRENDING_UP"
        if c[1] < a[1] and d2[1] < b2[1]:
            return "TRENDING_DOWN"
    return "RANGING"

def replay(t_ms, side, entry, sl, tp, eod_hours=24):
    eod = t_ms + eod_hours * 3600000
    keys = [k for k in keys1 if t_ms - 60000 < k <= eod]
    for k in keys:
        b = bars1[k]
        hit_sl = b[2] <= sl if side == "long" else b[1] >= sl
        hit_tp = b[1] >= tp if side == "long" else b[2] <= tp
        if hit_sl and hit_tp:
            p = (sl - entry) if side == "long" else (entry - sl)
            return "AMBIG", p
        if hit_sl:
            p = (sl - entry) if side == "long" else (entry - sl)
            return "LOST", p
        if hit_tp:
            p = (tp - entry) if side == "long" else (entry - tp)
            return "WON", p
    last = bars1[keys[-1]][3] if keys else entry
    p = (last - entry) if side == "long" else (entry - last)
    return "EOD", p

# rebuild veto census (same source as Part B)
def plan_doc_at(t_ms, session):
    d = datetime.fromtimestamp(t_ms / 1000, CT)
    pat = d.strftime("%Y-%m-%d") + "%:" + session + ":%"
    row = DB.execute("SELECT doc FROM plans WHERE plan_id LIKE ? AND created_at <= ? "
                     "ORDER BY created_at DESC LIMIT 1",
                     (pat, datetime.fromtimestamp(t_ms / 1000, timezone.utc).strftime("%Y-%m-%d %H:%M:%S"))).fetchone()
    return json.loads(row[0]) if row else None

vetos = []
W0ms = int(datetime(2026, 8, 26, 0, 0, tzinfo=CT).timestamp() * 1000)
for ts_utc, msg in DB.execute(
        "SELECT ts_utc, message FROM log_events WHERE message LIKE '%arm REFUSED%HTF%' "
        "AND ts_utc >= ? ORDER BY ts_utc", (W0ms,)):
    import re
    m = re.search(r"arm REFUSED (\S+) (S\d+): ", msg)
    if not m:
        continue
    doc = plan_doc_at(ts_utc, m.group(1))
    sc = next((s for s in doc.get("scenarios", []) if s.get("id") == m.group(2)), None) if doc else None
    a = (sc or {}).get("arm") or {}
    if not a.get("entry") or not a.get("stop") or not a.get("target"):
        continue
    vetos.append(dict(t=ts_utc, side=sc.get("direction", "long"),
                      entry=a["entry"], sl=a["stop"], tp=a["target"], msg=msg))
seen = set()
for v in vetos:
    k = (v["side"], v["entry"], v["sl"], v["tp"])
    if k in seen:
        continue
    seen.add(k)
    row = []
    for tf in (60, 240):
        tr = tf_trend(v["t"], tf)
        oppose = (v["side"] == "long" and tr == "TRENDING_DOWN") or (v["side"] == "short" and tr == "TRENDING_UP")
        out, p = replay(v["t"], v["side"], v["entry"], v["sl"], v["tp"]) if not oppose else ("VETOED", 0.0)
        row.append((tf, tr, out, p * 2))
    t1, t4 = row[0], row[1]
    print("  %s %-5s e=%.2f | 1h:%s 4h:%s | 1h->%s $%+.0f | 4h->%s $%+.0f" % (
        datetime.fromtimestamp(v["t"] / 1000, CT).strftime("%m-%d %H:%M"), v["side"], v["entry"],
        t1[1], t4[1], t1[2], t1[3], t4[2], t4[3]))

# ============================================================
# C4 — min-conf 60 vs 65
# ============================================================
print("\n== C4 min-confidence 60 vs 65 (closed, pnl_corrected, entry_confidence) ==")
lo = int(datetime(2026, 8, 26, 0, 0, tzinfo=CT).timestamp() * 1000)
bands = defaultdict(lambda: dict(n=0, s=0.0))
for conf, pc in DB.execute(
        "SELECT entry_confidence, pnl_corrected FROM trader_positions "
        "WHERE status='CLOSED' AND entry_time >= ? AND pnl_corrected IS NOT NULL", (lo,)):
    c = conf or 0
    key = "60-64" if 60 <= c < 65 else ("65+" if c >= 65 else "<60/unknown")
    bands[key]["n"] += 1
    bands[key]["s"] += pc
for k in ("60-64", "65+", "<60/unknown"):
    g = bands[k]
    print("  %-12s n=%2d  S=%+8.1f" % (k, g["n"], g["s"]))

# ============================================================
# C5 — trail mult sweep on closed trades
# ============================================================
print("\n== C5 trail-mult sweep (2.0 / 2.5 / 3.0, ATR14(5m), arm after-BE+40) ==")
trades = DB.execute(
    "SELECT id, side, entry_price, exit_price, entry_time, exit_time, mfe, pnl_corrected "
    "FROM trader_positions WHERE status='CLOSED' AND entry_time >= ? AND symbol='MNQ' "
    "AND pnl_corrected IS NOT NULL ORDER BY id", (lo,)).fetchall()
print("%-3s %-5s %8s %8s %5s %7s | %7s %7s %7s | %6s" % (
    "id", "side", "entry", "exit", "mfe", "actual$", "t2.0$", "t2.5$", "t3.0$", "gb2.0"))
for tid, side, entry, exitp, et, xt, mfe, pc in trades:
    long = side == "LONG"
    ks = [k for k in keys1 if et <= k <= xt]
    res = []
    for mult in (2.0, 2.5, 3.0):
        best = entry
        stop = None
        realized = None
        for k in ks:
            b = bars1[k]
            hi, lo, cl = b[1], b[2], b[3]
            if long:
                best = max(best, hi)
                pnl = cl - entry
            else:
                best = min(best, lo)
                pnl = entry - cl
            # rolling ATR14(5m) up to this bar
            sub = [x for x in keys1 if x <= k]
            b5 = agg_tf(sub[-5 * 16:], 5) if len(sub) >= 75 else []
            atr = wilder_atr(b5, 14) if b5 else 0.0
            if pnl >= 40 and atr > 0:   # arm after BE+40 (live be_trigger=40)
                tgt = best - mult * atr if long else best + mult * atr
                if stop is None or (long and tgt > stop) or (not long and tgt < stop):
                    stop = tgt
            if stop is not None:
                hit = lo <= stop if long else hi >= stop
                if hit:
                    realized = stop
                    break
        if realized is None:
            realized = exitp
        res.append((realized - entry) * 2.0 if long else (entry - realized) * 2.0)
    gb = (mfe or 0) * 2.0 - res[0]
    print("%-3d %-5s %8.2f %8.2f %5.1f %+7.0f | %+7.0f %+7.0f %+7.0f | %+6.0f" % (
        tid, side, entry, exitp, (mfe or 0), (pc or 0), res[0], res[1], res[2], gb))

# ============================================================
# C6 — proximity band sweep
# ============================================================
print("\n== C6 proximity band sweep (seated pool + excluded entry-levels per plan) ==")
rows = DB.execute(
    "SELECT doc, created_at FROM plans WHERE lifecycle != 'no_trade' AND created_at >= ? "
    "ORDER BY created_at", ("2026-08-26",)).fetchall()
stats = defaultdict(lambda: dict(n=0, pool=0, excl=0, entry=0))
for doc, created in rows:
    d = json.loads(doc)
    lv = [l["price"] for l in d.get("levels", []) if l.get("price")]
    entries = [s.get("arm", {}).get("entry") for s in d.get("scenarios", [])]
    entries = [e for e in entries if e]
    t_ms = int(datetime.fromisoformat(created).timestamp() * 1000)
    sub = [k for k in keys1 if k <= t_ms]
    if len(sub) < 1440:
        continue
    price = bars1[sub[-1]][3]
    # prior completed CME session-day range (17:00 CT prev day -> 16:00 CT)
    d0 = datetime.fromtimestamp(t_ms / 1000, CT)
    prev_open = int(datetime(d0.year, d0.month, d0.day, 17, 0, tzinfo=CT).timestamp() * 1000) - 86400000
    prev_close = prev_open + 23 * 3600000
    pw = [bars1[k] for k in keys1 if prev_open <= k < prev_close]
    rng = (max(b[1] for b in pw) - min(b[2] for b in pw)) if pw else 0.0
    if rng <= 0:
        continue
    for k in (0.2, 0.3, 0.4, 1.5):
        band = k * rng
        pool = sum(1 for p in lv if abs(p - price) <= band)
        excl = sum(1 for e in entries if abs(e - price) > band)
        g = stats[k]
        g["n"] += 1
        g["pool"] += pool
        g["excl"] += excl
        g["entry"] += len(entries)
print("%-6s %3s %10s %12s %10s" % ("band", "n", "mean_pool", "mean_excl", "entries"))
for k in (0.2, 0.3, 0.4, 1.5):
    g = stats[k]
    print("%-6s %3d %10.1f %12.1f %10d" % ("%.1fx" % k, g["n"], g["pool"] / max(g["n"], 1),
                                            g["excl"] / max(g["n"], 1), g["entry"]))
