#!/usr/bin/env python3
"""
ZONE-MATH TOTAL VERIFICATION — independent reimplementation (R2 rule).
Python3 + stdlib ONLY (sqlite3). No Go code invoked.

Verifies the six zone classes for the last 3 complete sessions:
  2026-08-28 NY (v7), 2026-08-28 LONDON (v6), 2026-08-27 NY (v5).

Every rule below was re-derived from the kernel spec sources listed in the
report (read as reference), then reimplemented from scratch here.

Usage: python3 zm_verify.py
Writes: docs/superpowers/reports/2026-08-29-zone-math-total-verification.md
"""

import sqlite3, json, math, os, sys, datetime as dt

DB_PATH = "file:/home/hoang/nofx/data/data.db?mode=ro"
OUT_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                        "docs", "superpowers", "reports",
                        "2026-08-29-zone-math-total-verification.md")

TICK = 0.25
SYMBOL = "MNQ"

# ----------------------------------------------------------------------------
# time zone: America/Chicago (CDT in August, UTC-5)
# ----------------------------------------------------------------------------
try:
    from zoneinfo import ZoneInfo
    CT = ZoneInfo("America/Chicago")
except Exception:
    class _CT(dt.tzinfo):
        def utcoffset(self, d): return dt.timedelta(hours=-5)
        def dst(self, d): return dt.timedelta(0)
    CT = _CT()

def ct_of(ms):
    return dt.datetime.fromtimestamp(ms / 1000.0, dt.timezone.utc).astimezone(CT)

def ms_of_ct(y, mo, d, hh, mm):
    return int(dt.datetime(y, mo, d, hh, mm, tzinfo=CT).timestamp() * 1000)

def cme_day_start_ms(ms):
    """kernel/cme_calendar.go:94-102 — 17:00 CT boundary; <17:00 → prior day."""
    t = ct_of(ms)
    b = dt.datetime(t.year, t.month, t.day, 17, 0, tzinfo=CT)
    if t.hour < 17:
        b -= dt.timedelta(days=1)
    return int(b.timestamp() * 1000)

def cme_day_key(ms):
    t = ct_of(cme_day_start_ms(ms))
    return "%04d-%02d-%02d" % (t.year, t.month, t.day)

# ----------------------------------------------------------------------------
# sessions (kernel/session_registry.go:88-109)
# ----------------------------------------------------------------------------
SESSIONS = {"ASIA": ("17:00", "02:00"), "LONDON": ("02:00", "08:30"),
            "NY": ("08:30", "14:45")}

def active_session(ms):
    t = ct_of(ms)
    hm = t.hour * 60 + t.minute
    if hm >= 17 * 60 or hm < 2 * 60:
        return "ASIA"
    if hm < 8 * 60 + 30:
        return "LONDON"
    if hm < 14 * 60 + 45:
        return "NY"
    return ""

# ----------------------------------------------------------------------------
# bars
# ----------------------------------------------------------------------------
def load_bars(db):
    rows = db.execute(
        "SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol=? AND tf='1m' "
        "ORDER BY open_time_ms", (SYMBOL,)).fetchall()
    bars = [{"t": r[0], "o": r[1], "h": r[2], "l": r[3], "c": r[4], "v": r[5]}
            for r in rows]
    return bars

def closed_1m_slice(bars, plan_ms, count=2000):
    """FuturesBarsProvider(symbol,'1m',2000): last 2000 cache bars; closedBars
    filters CloseTime(=T+60000-1) < now. barsToKlines CloseTime = T+dur-1."""
    cand = [b for b in bars if b["t"] <= plan_ms - 60_000 + 60_000]  # include forming
    if len(cand) > count:
        cand = cand[-count:]
    closed = [b for b in cand if b["t"] + 60_000 - 1 < plan_ms]
    return closed

def aggregate(bars, span_min, plan_ms=None, closed_only=True):
    """Buckets 1m bars to span_min; closed-only keeps buckets with bucket_end-1 < now."""
    span = span_min * 60_000
    out = []
    for b in bars:
        key = b["t"] // span * span
        if out and out[-1]["t"] == key:
            a = out[-1]
            a["h"] = max(a["h"], b["h"]); a["l"] = min(a["l"], b["l"])
            a["c"] = b["c"]; a["v"] += b["v"]
        else:
            out.append({"t": key, "o": b["o"], "h": b["h"], "l": b["l"],
                        "c": b["c"], "v": b["v"]})
    if plan_ms is not None and closed_only:
        out = [b for b in out if b["t"] + span - 1 < plan_ms]
    return out

def wilder_atr14(bars):
    """market/data_indicators.go:86-116 — TR max, seed = mean(TR[1..14]),
    Wilder smoothing. Returns 0 when len <= period."""
    n = len(bars)
    if n <= 14:
        return 0.0
    trs = [0.0] * n
    for i in range(1, n):
        h, l, pc = bars[i]["h"], bars[i]["l"], bars[i - 1]["c"]
        trs[i] = max(h - l, abs(h - pc), abs(l - pc))
    s = sum(trs[1:15])
    atr = s / 14.0
    for i in range(15, n):
        atr = (atr * 13.0 + trs[i]) / 14.0
    return atr

def daily_range_proxy(bars, plan_ms):
    """kernel/levels_assemble.go:285-330 — mean of COMPLETED CME session-day
    ranges (current key excluded); fallback = developing day's range."""
    days = {}
    for b in bars:
        k = cme_day_key(b["t"])
        d = days.setdefault(k, [1e18, -1e18])
        d[0] = min(d[0], b["l"]); d[1] = max(d[1], b["h"])
    nowk = cme_day_key(plan_ms)
    s, n = 0.0, 0
    for k, (lo, hi) in days.items():
        if k == nowk:
            continue
        if hi > lo:
            s += hi - lo; n += 1
    if n > 0:
        return s / n
    if nowk in days and days[nowk][1] > days[nowk][0]:
        return days[nowk][1] - days[nowk][0]
    return 0.0

# ----------------------------------------------------------------------------
# detectors (all reimplemented from kernel/levels_zones.go, levels_swing.go,
# levels_volume.go, levels_intraday.go, levels_multiday.go, fvg_entry.go)
# ----------------------------------------------------------------------------
def fvg_window_contiguous(cb, i):
    """kernel/levels_zones.go:190-207 — span of 3 candles <= 3x the series' own
    min positive interval (last 10 bars)."""
    if i < 2 or i >= len(cb):
        return False
    iv = 0
    for j in range(len(cb) - 1, max(0, len(cb) - 10), -1):
        d = cb[j]["t"] - cb[j - 1]["t"]
        if d > 0 and (iv == 0 or d < iv):
            iv = d
    if iv <= 0:
        return False
    return cb[i]["t"] - cb[i - 2]["t"] <= 3 * iv

def fair_value_gaps(cb, min_gap, plan_ms, htf=False, tf_name=""):
    """kernel/levels_zones.go:213-279 — 3-candle gaps; iFVG inversion on a later
    close beyond the far edge; gap floor min_gap (1m: max(2*tick,2.0); HTF:
    that TF's ATR)."""
    out = []
    for i in range(2, len(cb)):
        if not fvg_window_contiguous(cb, i):
            continue
        a, c = cb[i - 2], cb[i]
        if a["h"] < c["l"] and c["l"] - a["h"] >= min_gap:
            gLo, gHi, bullish = a["h"], c["l"], True
        elif a["l"] > c["h"] and a["l"] - c["h"] >= min_gap:
            gLo, gHi, bullish = c["h"], a["l"], False
        else:
            continue
        inv = False
        for j in range(i + 1, len(cb)):
            if bullish and cb[j]["c"] < gLo:
                inv = True; break
            if not bullish and cb[j]["c"] > gHi:
                inv = True; break
        lbl = "iFVG(bear)" if (inv and bullish) else \
              "iFVG(bull)" if inv else "FVG"
        out.append({"kind": "IFVG" if inv else "FVG",
                    "lo": gLo, "hi": gHi, "mid": (gLo + gHi) / 2,
                    "label": lbl + ("·" + tf_name if htf else ""),
                    "bullish": bullish, "t": c["t"], "htf": htf,
                    "tf": tf_name})
    return out

def fresh_fvg_candidates(cb, plan_ms):
    """kernel/fvg_entry.go:69-122 — 3-candle gap, displacement of the impulse
    candle >= 1.5 x Wilder ATR14(5m); lookback 40; newest first."""
    if len(cb) < 3:
        return []
    five = aggregate(cb, 5)
    atr5 = wilder_atr14(five) if len(five) >= 15 else 0.0
    lookback = min(len(cb), 40)
    floor = max(2 * TICK, 2.0)
    out = []
    for i in range(len(cb) - 3, max(-1, len(cb) - lookback - 3), -1):
        if not fvg_window_contiguous(cb, i + 2):
            continue
        if cb[i + 2]["l"] > cb[i]["h"]:
            dirn, gLo, gHi = "long", cb[i]["h"], cb[i + 2]["l"]
        elif cb[i + 2]["h"] < cb[i]["l"]:
            dirn, gLo, gHi = "short", cb[i + 2]["h"], cb[i]["l"]
        else:
            continue
        if gHi - gLo < floor:
            continue
        imp = abs(cb[i + 2]["c"] - cb[i + 2]["o"])
        disp = 0.0
        if atr5 > 0:
            disp = imp / atr5
            if disp < 1.5:
                continue
        out.append({"direction": dirn, "lo": gLo, "hi": gHi,
                    "age": len(cb) - 1 - (i + 2), "disp_atr": round(disp, 2),
                    "t": cb[i + 2]["t"]})
    return out

def order_blocks(cb, atr, plan_ms, htf=False, tf_name=""):
    """kernel/levels_zones.go:257-337 — displacement >=1.5xATR; last opposing
    candle within lookback 8; zone = that candle's [low,high]."""
    if atr <= 0 or len(cb) < 2:
        return []
    disp = 1.5 * atr
    lookback = 8
    out = []
    for i in range(1, len(cb)):
        move = cb[i]["c"] - cb[i]["o"]
        if move >= disp:
            for j in range(i - 1, max(-1, i - 1 - lookback), -1):
                if cb[j]["c"] < cb[j]["o"]:
                    out.append({"kind": "OB", "lo": cb[j]["l"], "hi": cb[j]["h"],
                                "mid": (cb[j]["l"] + cb[j]["h"]) / 2,
                                "label": "OB(bull)" + ("·" + tf_name if htf else ""),
                                "bullish": True, "t": cb[i]["t"],
                                "htf": htf, "tf": tf_name})
                    break
        elif -move >= disp:
            for j in range(i - 1, max(-1, i - 1 - lookback), -1):
                if cb[j]["c"] > cb[j]["o"]:
                    out.append({"kind": "OB", "lo": cb[j]["l"], "hi": cb[j]["h"],
                                "mid": (cb[j]["l"] + cb[j]["h"]) / 2,
                                "label": "OB(bear)" + ("·" + tf_name if htf else ""),
                                "bullish": False, "t": cb[i]["t"],
                                "htf": htf, "tf": tf_name})
                    break
    return out

def sd_zones(cb, atr, plan_ms, htf=False, tf_name=""):
    """kernel/levels_zones.go:99-150 — base of <=6 small-bodied candles
    (body <= 0.5xATR) + departure >= 1.5xATR; zone = base [low,high];
    pattern = reversal if prior leg opposite sign."""
    if atr <= 0 or len(cb) < 3:
        return []
    small = 0.5 * atr; departure = 1.5 * atr
    out = []
    i = 0
    while i < len(cb):
        if abs(cb[i]["c"] - cb[i]["o"]) > small:
            i += 1
            continue
        baseLo, baseHi = cb[i]["l"], cb[i]["h"]
        j = i
        while j + 1 < len(cb) and j - i < 5 and abs(cb[j + 1]["c"] - cb[j + 1]["o"]) <= small:
            j += 1
            baseLo = min(baseLo, cb[j]["l"]); baseHi = max(baseHi, cb[j]["h"])
        if j + 1 < len(cb):
            d = cb[j + 1]
            move = d["c"] - d["o"]
            pat = ""
            if i > 0:
                leg = cb[i - 1]["c"] - cb[i - 1]["o"]
                pat = "reversal" if (leg >= 0) != (move >= 0) else "continuation"
            if move >= departure:
                out.append({"kind": "DEMAND", "lo": baseLo, "hi": baseHi,
                            "mid": (baseLo + baseHi) / 2,
                            "label": "Demand" + ("·" + tf_name if htf else ""),
                            "pattern": pat, "t": d["t"], "htf": htf,
                            "tf": tf_name, "bullish": True})
            elif -move >= departure:
                out.append({"kind": "SUPPLY", "lo": baseLo, "hi": baseHi,
                            "mid": (baseLo + baseHi) / 2,
                            "label": "Supply" + ("·" + tf_name if htf else ""),
                            "pattern": pat, "t": d["t"], "htf": htf,
                            "tf": tf_name, "bullish": False})
        i = j + 1
    return out

def eqhl(cb, tol, htf=False, tf_name=""):
    """kernel/levels_zones.go:36-96 — k=2 strict pivots, clusterEqual
    (sorted; groups within tol of group min; >=2 → level at group max/min)."""
    if tol <= 0 or len(cb) < 5:
        return []
    hi, lo = [], []
    for i in range(2, len(cb) - 2):
        is_h = all(cb[j]["h"] < cb[i]["h"] for j in range(i - 2, i + 3) if j != i)
        is_l = all(cb[j]["l"] > cb[i]["l"] for j in range(i - 2, i + 3) if j != i)
        if is_h:
            hi.append(cb[i]["h"])
        if is_l:
            lo.append(cb[i]["l"])
    out = []
    for prices, high, lbl in ((hi, True, "EQH"), (lo, False, "EQL")):
        prices = sorted(prices)
        k = 0
        while k < len(prices):
            m = k + 1
            while m < len(prices) and prices[m] - prices[k] <= tol:
                m += 1
            if m - k >= 2:
                p = prices[m - 1] if high else prices[k]
                out.append({"kind": "EQL" if not high else "EQH", "price": p,
                            "lo": p, "hi": p,
                            "label": lbl + ("·" + tf_name if htf else ""),
                            "htf": htf, "tf": tf_name})
            k = m
    return out

def simple_atr14(h, l, c):
    """kernel/structure.go:157-193 — Wilder ATR14 on a series (same smoothing)."""
    n = len(c)
    if n == 0:
        return 0.0
    if n == 1:
        return h[0] - l[0]
    trs = [h[0] - l[0]]
    for i in range(1, n):
        trs.append(max(h[i] - l[i], abs(h[i] - c[i - 1]), abs(l[i] - c[i - 1])))
    if n <= 14:
        return sum(trs[1:]) / (n - 1)
    s = sum(trs[1:15]); atr = s / 14.0
    for i in range(15, n):
        atr = (atr * 13 + trs[i]) / 14.0
    return atr

def swing_points(agg, tf_min, plan_ms):
    """kernel/levels_swing.go:48-140 — k=2 fractal, same-side keep-more-extreme,
    min-move 0.25xATR(tf) vs prior opposite swing, newest-first, lookback
    144 bars (5m) / 96 (15m), <=3 per side."""
    closed = [b for b in agg if b["t"] + tf_min * 60_000 - 1 < plan_ms]
    n = len(closed)
    if n < 5:
        return []
    highs = [b["h"] for b in closed]; lows = [b["l"] for b in closed]
    closes = [b["c"] for b in closed]
    atr = simple_atr14(highs, lows, closes)
    if atr <= 0:
        return []
    swings = []
    for i in range(2, n - 2):
        is_h = all(highs[j] < highs[i] for j in range(i - 2, i + 3) if j != i)
        is_l = all(lows[j] > lows[i] for j in range(i - 2, i + 3) if j != i)
        if not is_h and not is_l:
            continue
        price = highs[i] if is_h else lows[i]
        hi = is_h
        t = closed[i]["t"] + tf_min * 60_000
        if swings and swings[-1]["high"] == hi:
            if (hi and price <= swings[-1]["price"]) or (not hi and price >= swings[-1]["price"]):
                continue
            swings[-1] = {"price": price, "t": t, "high": hi}
            continue
        if swings:
            move = abs(price - swings[-1]["price"])
            if move < 0.25 * atr:
                continue
        swings.append({"price": price, "t": t, "high": hi})
    if not swings:
        return []
    lb = 144 if tf_min == 5 else 96
    cutoff = plan_ms - lb * tf_min * 60_000
    out = []
    added_h = added_l = 0
    for s in reversed(swings):
        if added_h >= 3 and added_l >= 3:
            break
        if s["t"] < cutoff:
            continue
        if s["high"]:
            if added_h >= 3:
                continue
            added_h += 1
            out.append({"kind": "SWG-H", "price": s["price"], "label": "SWG-H·" + str(tf_min) + "m", "t": s["t"]})
        else:
            if added_l >= 3:
                continue
            added_l += 1
            out.append({"kind": "SWG-L", "price": s["price"], "label": "SWG-L·" + str(tf_min) + "m", "t": s["t"]})
    return out

def vwap_stdev(bars):
    """kernel/levels_volume.go:66-83 — TP=(H+L+C)/3 volume-weighted; sd = sqrt(Σv d²/Σv)."""
    pv = v = 0.0
    for b in bars:
        tp = (b["h"] + b["l"] + b["c"]) / 3
        pv += tp * b["v"]; v += b["v"]
    if v <= 0:
        return 0.0, 0.0
    vw = pv / v
    acc = 0.0
    for b in bars:
        tp = (b["h"] + b["l"] + b["c"]) / 3
        acc += b["v"] * (tp - vw) ** 2
    return vw, math.sqrt(acc / v)

def profile_levels(bars):
    """kernel/levels_volume.go:194-240 — 120 bins by CLOSE; POC = lo+(idx+.5)*bin;
    70% value area expanding from POC (up>=down → up); VAH=lo+(hiIdx+1)*bin,
    VAL=lo+loIdx*bin."""
    if not bars:
        return None
    hi = max(b["h"] for b in bars); lo = min(b["l"] for b in bars)
    if hi <= lo:
        return None
    binw = (hi - lo) / 120.0
    if binw <= 0:
        binw = TICK
    vols = [0.0] * 120
    total = 0.0
    for b in bars:
        idx = int((b["c"] - lo) / binw)
        idx = min(max(idx, 0), 119)
        vols[idx] += b["v"]; total += b["v"]
    if total <= 0:
        return None
    poc_idx = max(range(120), key=lambda i: vols[i])
    poc = lo + (poc_idx + 0.5) * binw
    need = 0.70 * total
    acc = vols[poc_idx]; li = hi_i = poc_idx
    while acc < need:
        below = li - 1; above = hi_i + 1
        up = vols[above] if above < 120 else 0.0
        down = vols[below] if below >= 0 else 0.0
        if up >= down and above < 120:
            hi_i = above; acc += up
        elif below >= 0:
            li = below; acc += down
        elif above < 120:
            hi_i = above; acc += up
        else:
            break
    return {"poc": poc, "vah": lo + (hi_i + 1) * binw, "val": lo + li * binw}

# ----------------------------------------------------------------------------
# full assembly (levels_assemble.go:71-112 + DetectHTFLevels:222-271)
# ----------------------------------------------------------------------------
KIND = {}

def _kind(name):
    KIND[name] = name
    return name

KPDH, KPDL, KPDC = _kind("PDH"), _kind("PDL"), _kind("PDC")
KRTHH, KRTHL, KORH, KORL = _kind("RTH-H"), _kind("RTH-L"), _kind("OR-H"), _kind("OR-L")
KONH, KONL = _kind("ONH"), _kind("ONL")
KASH, KASL, KLDNH, KLDNL = _kind("AS-H"), _kind("AS-L"), _kind("LDN-H"), _kind("LDN-L")
KPWH, KPWL, KPMH, KPML = _kind("PWH"), _kind("PWL"), _kind("PMH"), _kind("PML")
KEQH, KEQL = _kind("EQH"), _kind("EQL")
KSUPPLY, KDEMAND, KFVG, KIFVG, KOB = (_kind(x) for x in ("SUPPLY", "DEMAND", "FVG", "IFVG", "OB"))
KVWAP, KEVWAP, KPDVWAP, KVWAP2S = _kind("VWAP"), _kind("eVWAP"), _kind("pdVWAP"), _kind("VWAP±2σ")
KPOC, KNPOC, KVAH, KVAL, KSETT, KMIDO = (_kind(x) for x in ("POC", "nPOC", "VAH", "VAL", "SETT", "MID-O"))
KSWGH, KSWGL = _kind("SWG-H"), _kind("SWG-L")
KROUND, KGAP, KIBH, KIBL = _kind("RN"), _kind("GAP"), _kind("IB-H"), _kind("IB-L")

ZONE_KINDS = {KSUPPLY, KDEMAND, KFVG, KIFVG, KOB}
TIER1 = {KPDH, KPDL, KPDC, KRTHH, KRTHL, KORH, KORL, KONH, KONL,
         KPWH, KPWL, KPMH, KPML, KVAH, KVAL, KSETT, KNPOC}

TYPE_EVIDENCE = {
    KPDH: 1.0, KPDL: 1.0, KPDC: 1.0, KRTHH: 1.0, KRTHL: 1.0,
    KPWH: 1.0, KPWL: 1.0, KPMH: 1.0, KPML: 1.0,
    KONH: 0.85, KONL: 0.85, KNPOC: 0.85,
    KVWAP: 0.90, KPOC: 0.90, KVWAP2S: 0.85, KSWGH: 0.85, KSWGL: 0.85,
    KEVWAP: 0.85, KPDVWAP: 0.85, KVAH: 0.80, KVAL: 0.80, KSETT: 0.80,
    KMIDO: 0.60,
    KASH: 0.70, KASL: 0.70, KLDNH: 0.70, KLDNL: 0.70, KORH: 0.70, KORL: 0.70,
    KIBH: 0.70, KIBL: 0.70, KEQH: 0.70, KEQL: 0.70,
    KROUND: 0.55, KGAP: 0.55,
    KSUPPLY: 0.30, KDEMAND: 0.30, KFVG: 0.30, KIFVG: 0.30, KOB: 0.30,
}

ZONE_EVIDENCE = {
    KOB: {"1m": 0.40, "15m": 0.50, "1h": 0.70, "4h": 0.72},
    KFVG: {"1m": 0.35, "15m": 0.45, "1h": 0.65, "4h": 0.65},
    KIFVG: {"1m": 0.35, "15m": 0.45, "1h": 0.65, "4h": 0.65},
    KSUPPLY: {"1m": 0.35, "15m": 0.45, "1h": 0.65, "4h": 0.65},
    KDEMAND: {"1m": 0.35, "15m": 0.45, "1h": 0.65, "4h": 0.65},
}
ZONE_TF_MULT = {"1m": 1.0, "15m": 1.1, "1h": 1.2, "4h": 1.3}
ZONE_REVERSAL_BONUS = 1.1

def zone_tier(tf):
    t = (tf or "").lower()
    if t in ("", "1m", "3m", "5m"):
        return "1m"
    if t == "30m":
        return "15m"
    if t == "2h":
        return "1h"
    if t in ("6h", "8h", "12h"):
        return "4h"
    if t in ("15m", "1h", "4h"):
        return t
    return "1m"

def zone_size_mult(lo, hi, atr):
    """kernel/levels_score.go:201-227 — banded 0.5..1.25 in daily-ATR units."""
    if lo <= 0 or hi < lo or atr <= 0:
        return 1.0
    size = (hi - lo) / atr
    if size <= 0.30:
        return 1.25
    if size <= 0.60:
        return 1.10
    if size <= 1.00:
        return 1.0
    if size <= 1.50:
        return 0.85
    if size <= 2.50:
        return 0.70
    return 0.50

def level_family(kind):
    if kind in (KVWAP, KEVWAP, KPDVWAP, KVWAP2S):
        return "vwap"
    if kind in (KPOC, KNPOC, KVAH, KVAL):
        return "profile"
    if kind in (KPDH, KPDL, KPDC, KRTHH, KRTHL, KSETT):
        return "prior"
    if kind in (KPWH, KPWL, KPMH, KPML):
        return "anchor"
    if kind in (KONH, KONL, KASH, KASL, KLDNH, KLDNL, KORH, KORL, KIBH, KIBL, KMIDO):
        return "overnight"
    if kind in (KEQH, KEQL):
        return "liquidity"
    if kind in ZONE_KINDS:
        return "zone"
    if kind == KROUND:
        return "round"
    if kind == KGAP:
        return "gap"
    return "other"

def today_priority(kind):
    return kind in (KPDH, KPDL, KPDC, KRTHH, KRTHL, KORH, KORL, KONH, KONL)

def zone_fresh_mult(f):
    return {"": 1.0, "a": 1.0, "fresh": 1.0, "b": 0.6, "c": 0.3, "tested": 0.3,
            "done": 0.15, "consumed": 0.15}.get(f.lower(), 1.0)

def fresh_mult(f):
    return {"": 1.0, "a": 1.0, "fresh": 1.0, "b": 0.8, "c": 0.6, "tested": 0.6,
            "done": 0.5, "consumed": 0.5}.get(f.lower(), 1.0)

def grade_from_score(s):
    if s >= 1.0:
        return "A"
    if s >= 0.70:
        return "B"
    return "C"

def score_pool(all_levels, price, datr, freshness_fn, max_levels, prox_k):
    """kernel/levels_score.go:410-578 — scoreLevelsPool."""
    if price <= 0 or datr <= 0:
        return []
    band = prox_k * datr
    conf_band = 0.10 * datr
    inband = [l for l in all_levels if abs(l["price"] - price) <= band]
    scored = []
    for l in inband:
        fr = freshness_fn(l) if freshness_fn else ""
        fm = fresh_mult(fr)
        if l["kind"] in ZONE_KINDS:
            fm = zone_fresh_mult(fr)
        if fm == 0:
            continue
        conf = 0
        seen = set()
        for o in inband:
            if o["price"] == l["price"] and o["kind"] == l["kind"] and o["label"] == l["label"]:
                continue
            if abs(o["price"] - l["price"]) <= conf_band:
                fam = level_family(o["kind"])
                if fam != level_family(l["kind"]) and fam not in seen:
                    seen.add(fam); conf += 1
        eff = min(conf, 3)
        if l["kind"] in ZONE_KINDS and conf == 0 and not l.get("htf"):
            continue
        htf = 1.2 if l.get("htf") else 1.0
        tier = zone_tier(l.get("tf", ""))
        if l["kind"] in ZONE_KINDS:
            base = ZONE_EVIDENCE[l["kind"]][tier]
            if l.get("pattern") == "reversal":
                base *= ZONE_REVERSAL_BONUS
            score = base * zone_size_mult(l.get("lo", 0), l.get("hi", 0), datr) \
                * fm * (1 + 0.20 * eff) * ZONE_TF_MULT[tier]
        else:
            score = TYPE_EVIDENCE.get(l["kind"], 0.50) * fm * (1 + 0.20 * eff) * htf
        grade = grade_from_score(score)
        if l["kind"] in ZONE_KINDS:
            if tier == "15m":
                grade = "B"
            elif tier in ("1h", "4h"):
                if grade == "C":
                    grade = "B"
            else:
                grade = "C"
            if grade != "C" and not within_tier1(l, inband):
                grade = "C"
        scored.append(dict(l, grade=grade, score=score, conf=conf,
                           fresh=fr, distance=l["price"] - price))
    # cluster collapse (P0.4)
    order = sorted(scored, key=lambda x: (not today_priority(x["kind"]), -x["score"],
                                          abs(x["distance"]), x["price"]))
    kept = []
    for cand in order:
        if cand["kind"] in ZONE_KINDS:
            kept.append(cand); continue
        merged = False
        for k in kept:
            if k["kind"] in ZONE_KINDS:
                continue
            if abs(k["price"] - cand["price"]) <= 3.0:
                k["conf"] += cand["conf"] + 1
                merged = True
                break
        if not merged:
            kept.append(cand)
    kept.sort(key=lambda x: (not today_priority(x["kind"]), -x["score"],
                             abs(x["distance"]), x["price"]))
    return kept

def within_tier1(l, inband):
    tol = 12 * TICK
    for o in inband:
        if o["kind"] not in TIER1:
            continue
        if abs(o["price"] - l["price"]) <= tol:
            return True
        if l["lo"] < l["hi"]:
            if l["lo"] <= o["price"] <= l["hi"]:
                return True
            if 0 < l["lo"] - o["price"] <= tol or 0 < o["price"] - l["hi"] <= tol:
                return True
    return False

def is_htf_swing_zone(l):
    return l.get("htf") and l["kind"] in (KEQH, KEQL, KSUPPLY, KDEMAND, KFVG, KIFVG, KOB)

def is_htf_seat_eligible(l):
    if not l.get("htf"):
        return False
    if l["kind"] in TIER1:
        return True
    return l["kind"] in ZONE_KINDS and l.get("pattern") == "reversal"

def is_vol_family(l):
    return l["kind"] in (KVWAP, KEVWAP, KPDVWAP, KVWAP2S, KPOC, KNPOC, KVAH, KVAL,
                         KSETT, KMIDO, KSWGH, KSWGL)

def seat_htf(scored, max_levels):
    if len(scored) <= max_levels:
        return scored
    head = scored[:max_levels]; tail = scored[max_levels:]
    seated = sum(1 for l in head if is_htf_seat_eligible(l))
    need = 2 - seated
    cands = [l for l in tail if is_htf_seat_eligible(l)]
    while need > 0 and cands:
        drop = -1
        for i in range(len(head) - 1, -1, -1):
            if today_priority(head[i]["kind"]) or is_htf_seat_eligible(head[i]):
                continue
            drop = i; break
        if drop < 0:
            break
        cand = cands.pop(0)
        tail = [t for t in tail if not (t["price"] == cand["price"] and t["kind"] == cand["kind"]
                                        and t["label"] == cand["label"] and t.get("htf") == cand.get("htf"))]
        tail.append(head.pop(drop))
        head.append(cand)
        need -= 1
    out = head + tail
    out.sort(key=lambda x: (not today_priority(x["kind"]), -x["score"],
                            abs(x["distance"]), x["price"]))
    return out

def seat_volume_family(scored, max_levels):
    if len(scored) <= max_levels:
        return scored
    head = scored[:max_levels]; tail = scored[max_levels:]
    if any(is_vol_family(l) for l in head):
        return scored
    best = -1
    for i, l in enumerate(tail):
        if is_vol_family(l) and (best < 0 or l["score"] > tail[best]["score"]):
            best = i
    if best < 0:
        return scored
    cand = tail[best]
    drop = -1
    for i in range(len(head) - 1, -1, -1):
        if (head[i]["kind"] in TIER1 and head[i]["grade"] == "A") or is_htf_seat_eligible(head[i]) or is_vol_family(head[i]):
            continue
        drop = i; break
    if drop < 0:
        return scored
    tail = tail[:best] + tail[best + 1:] + [head[drop]]
    head = head[:drop] + head[drop + 1:] + [cand]
    return head + tail

def seat_1h_zone(scored, max_levels):
    def is_1h_sd(l):
        return zone_tier(l.get("tf", "")) == "1h" and l["kind"] in (KSUPPLY, KDEMAND)
    if len(scored) <= max_levels:
        return scored
    head = scored[:max_levels]; tail = scored[max_levels:]
    if any(is_1h_sd(l) for l in head):
        return scored
    best = -1
    for i, l in enumerate(tail):
        if is_1h_sd(l) and (best < 0 or l["score"] > tail[best]["score"]):
            best = i
    if best < 0:
        return scored
    cand = tail[best]
    drop = -1
    for i in range(len(head) - 1, -1, -1):
        if today_priority(head[i]["kind"]) or is_htf_seat_eligible(head[i]):
            continue
        drop = i; break
    if drop < 0:
        return scored
    tail = tail[:best] + tail[best + 1:] + [head[drop]]
    head = head[:drop] + head[drop + 1:] + [cand]
    out = head + tail
    out.sort(key=lambda x: (not today_priority(x["kind"]), -x["score"],
                            abs(x["distance"]), x["price"]))
    return out

def seat_both_sides(scored, max_levels):
    n = len(scored)
    if n <= max_levels or max_levels < 6:
        return scored
    seated = scored[:max_levels]; rest = scored[max_levels:]
    for below in (True, False):
        cnt = sum(1 for l in seated if (l["distance"] < 0) == below)
        need = 3 - cnt
        cands = [l for l in rest if (l["distance"] < 0) == below]
        while cands and need > 0:
            drop = -1
            for i in range(len(seated) - 1, -1, -1):
                if (seated[i]["distance"] < 0) != below and not is_vol_family(seated[i]):
                    drop = i; break
            if drop < 0:
                break
            rest.append(seated.pop(drop))
            seated.append(cands.pop(0))
            need -= 1
    seated.sort(key=lambda x: (not today_priority(x["kind"]), -x["score"],
                               abs(x["distance"]), x["price"]))
    return seated

# ----------------------------------------------------------------------------
# other detectors
# ----------------------------------------------------------------------------
def multi_day_levels(cb, plan_ms):
    out = []
    cal = {}
    nowf = cme_day_key(plan_ms)
    as_h = as_l = ldn_h = ldn_l = None
    has_as = has_ldn = False
    for b in cb:
        t = ct_of(b["t"])
        key = "%04d-%02d-%02d" % (t.year, t.month, t.day)
        a = cal.setdefault(key, {"hi": -1e18, "lo": 1e18, "c": 0.0, "ct": 0,
                                 "rthh": -1e18, "rthl": 1e18, "hasrth": False})
        a["hi"] = max(a["hi"], b["h"]); a["lo"] = min(a["lo"], b["l"])
        if b["t"] >= a["ct"]:
            a["c"] = b["c"]; a["ct"] = b["t"]
        if active_session(b["t"]) == "NY":
            a["hasrth"] = True
            a["rthh"] = max(a["rthh"], b["h"]); a["rthl"] = min(a["rthl"], b["l"])
        if cme_day_key(b["t"]) == nowf:
            s = active_session(b["t"])
            if s == "ASIA":
                has_as = True
                as_h = b["h"] if as_h is None else max(as_h, b["h"])
                as_l = b["l"] if as_l is None else min(as_l, b["l"])
            if s == "LONDON":
                has_ldn = True
                ldn_h = b["h"] if ldn_h is None else max(ldn_h, b["h"])
                ldn_l = b["l"] if ldn_l is None else min(ldn_l, b["l"])
    tnow = ct_of(plan_ms)
    today_cal = "%04d-%02d-%02d" % (tnow.year, tnow.month, tnow.day)
    counts = {}
    for b in cb:
        t = ct_of(b["t"])
        key = "%04d-%02d-%02d" % (t.year, t.month, t.day)
        counts[key] = counts.get(key, 0) + 1
    # most recent prior calendar day with >=900 closed bars
    prior = None
    for d in range(1, 8):
        dk = tnow - dt.timedelta(days=d)
        key = "%04d-%02d-%02d" % (dk.year, dk.month, dk.day)
        if key in counts and counts[key] >= 900:
            prior = key
            break
    if prior:
        a = cal[prior]
        out.append({"kind": KPDH, "price": a["hi"], "label": "PDH", "htf": True})
        out.append({"kind": KPDL, "price": a["lo"], "label": "PDL", "htf": True})
        out.append({"kind": KPDC, "price": a["c"], "label": "PDC", "htf": True})
        if a["hasrth"]:
            out.append({"kind": KRTHH, "price": a["rthh"], "label": "RTH-H"})
            out.append({"kind": KRTHL, "price": a["rthl"], "label": "RTH-L"})
    if has_as:
        out.append({"kind": KASH, "price": as_h, "label": "AS-H"})
        out.append({"kind": KASL, "price": as_l, "label": "AS-L"})
    if has_ldn:
        out.append({"kind": KLDNH, "price": ldn_h, "label": "LDN-H"})
        out.append({"kind": KLDNL, "price": ldn_l, "label": "LDN-L"})
    if has_as or has_ldn:
        on_h = max([x for x in (as_h, ldn_h) if x is not None])
        on_l = min([x for x in (as_l, ldn_l) if x is not None])
        out.append({"kind": KONH, "price": on_h, "label": "ONH"})
        out.append({"kind": KONL, "price": on_l, "label": "ONL"})
    return out

def opening_range_levels(cb, plan_ms):
    out = []
    tnow = ct_of(plan_ms)
    rth_open = dt.datetime(tnow.year, tnow.month, tnow.day, 8, 30, tzinfo=CT)
    if tnow < rth_open:
        return out
    or_end = rth_open + dt.timedelta(minutes=5)
    ib_end = rth_open + dt.timedelta(minutes=60)
    or_h = or_l = ib_h = ib_l = None
    has_or = has_ib = False
    for b in cb:
        bt = ct_of(b["t"])
        if rth_open <= bt < or_end:
            has_or = True
            or_h = b["h"] if or_h is None else max(or_h, b["h"])
            or_l = b["l"] if or_l is None else min(or_l, b["l"])
        if rth_open <= bt < ib_end:
            has_ib = True
            ib_h = b["h"] if ib_h is None else max(ib_h, b["h"])
            ib_l = b["l"] if ib_l is None else min(ib_l, b["l"])
    if has_or:
        out.append({"kind": KORH, "price": or_h, "label": "OR-H"})
        out.append({"kind": KORL, "price": or_l, "label": "OR-L"})
    if has_ib:
        r = ib_h - ib_l
        out.append({"kind": KIBH, "price": ib_h, "label": "IB-H"})
        out.append({"kind": KIBL, "price": ib_l, "label": "IB-L"})
    return out

def gap_levels(cb, atr):
    if atr <= 0:
        return []
    min_gap = 1.0 * atr
    out = []
    for i in range(1, len(cb)):
        p, c = cb[i - 1], cb[i]
        if c["l"] > p["h"] and c["l"] - p["h"] >= min_gap:
            gLo, gHi, up = p["h"], c["l"], True
        elif c["h"] < p["l"] and p["l"] - c["h"] >= min_gap:
            gLo, gHi, up = c["h"], p["l"], False
        else:
            continue
        filled = False
        for j in range(i + 1, len(cb)):
            if up and cb[j]["l"] <= gLo:
                filled = True; break
            if not up and cb[j]["h"] >= gHi:
                filled = True; break
        if not filled:
            out.append({"kind": KGAP, "price": gLo if up else gHi,
                        "label": "GAP"})
    return out

def round_number_levels(price, datr, prox_k):
    out = []
    if price <= 0 or datr <= 0:
        return out
    band = prox_k * datr
    lo, hi = price - band, price + band
    seen = set()
    for step, tag in ((100, "100"), (50, "50"), (25, "25")):
        n0 = int(math.ceil(lo / step))
        n = n0
        while n * step <= hi:
            m = n * step
            key = int(round(m * 100))
            if key not in seen:
                seen.add(key)
                out.append({"kind": KROUND, "price": m, "label": "RN %.0f (%s)" % (m, tag)})
            n += 1
    out.sort(key=lambda x: x["price"])
    return out

def volume_levels(cb, plan_ms):
    """SessionVWAP(±1σ,±2σ) · eVWAP · pd profile · nPOC · pdVWAP · SETT · MID-O."""
    out = []
    if not cb:
        return out
    sess_start = cme_day_start_ms(plan_ms)
    session = [b for b in cb if b["t"] >= sess_start]
    vwap, sd = vwap_stdev(session)
    if vwap > 0 and len(session) >= 2:
        out.append({"kind": KVWAP, "price": vwap, "label": "VWAP"})
        out.append({"kind": KVWAP, "price": vwap + sd, "label": "VWAP+1σ"})
        out.append({"kind": KVWAP, "price": vwap - sd, "label": "VWAP−1σ"})
        out.append({"kind": KVWAP2S, "price": vwap + 2 * sd, "label": "VWAP+2σ"})
        out.append({"kind": KVWAP2S, "price": vwap - 2 * sd, "label": "VWAP−2σ"})
    # eVWAP: most recent 15:00 CT
    tnow = ct_of(plan_ms)
    anchor = dt.datetime(tnow.year, tnow.month, tnow.day, 15, 0, tzinfo=CT)
    if tnow.hour < 15:
        anchor -= dt.timedelta(days=1)
    win = [b for b in cb if b["t"] >= int(anchor.timestamp() * 1000)]
    ev, _ = vwap_stdev(win)
    if ev > 0 and len(win) >= 2:
        out.append({"kind": KEVWAP, "price": ev, "label": "eVWAP"})
    # prior CME session-day profile
    cur_day = cme_day_start_ms(plan_ms)
    prior = [b for b in cb if cur_day - 86400000 <= b["t"] < cur_day]
    if prior:
        prof = profile_levels(prior)
        if prof:
            out.append({"kind": KPOC, "price": prof["poc"], "label": "POC"})
            out.append({"kind": KVAH, "price": prof["vah"], "label": "VAH"})
            out.append({"kind": KVAL, "price": prof["val"], "label": "VAL"})
    if prior:
        pv, _ = vwap_stdev(prior)
        if pv > 0 and len(prior) >= 2:
            out.append({"kind": KPDVWAP, "price": pv, "label": "pdVWAP"})
        out.append({"kind": KSETT, "price": prior[-1]["c"], "label": "SETT"})
    # MID-O
    tnow2 = ct_of(plan_ms)
    cut = dt.datetime(tnow2.year, tnow2.month, tnow2.day, 8, 30, tzinfo=CT)
    if tnow2.hour < 8 or (tnow2.hour == 8 and tnow2.minute < 30):
        cut = tnow2
    mids = [b for b in cb if sess_start <= b["t"] <= int(cut.timestamp() * 1000)]
    if mids:
        hi = max(b["h"] for b in mids); lo = min(b["l"] for b in mids)
        out.append({"kind": KMIDO, "price": (hi + lo) / 2, "label": "MID-O"})
    # nPOC from in-slice prior days (NakedPOCLevels)
    for d in range(1, 11):
        ds = cur_day - d * 86400000
        de = ds + 86400000
        db = [b for b in cb if ds <= b["t"] < de]
        if not db:
            continue
        prof = profile_levels(db)
        if not prof:
            continue
        poc = prof["poc"]
        birth = db[-1]["t"] + 60_000 - 1
        touched = False
        for b in cb:
            if b["t"] <= birth:
                continue
            if b["l"] <= poc - 0.25 and b["h"] >= poc + 0.25:
                touched = True; break
        if not touched:
            out.append({"kind": KNPOC, "price": poc, "label": "nPOC·" + cme_day_key(ds)})
    return out

# ----------------------------------------------------------------------------
# scoring + seating driver (mirrors assemblePlannerInputWithCtx)
# ----------------------------------------------------------------------------
def stored_atrs(db, td, sess):
    """Parse indicators_block sections for stored per-TF ATR14 values."""
    row = db.execute("SELECT indicators_block FROM plans WHERE trade_date=? AND session=? "
                     "ORDER BY version DESC LIMIT 1", (td, sess)).fetchone()
    out = {}
    if row:
        import re
        for p in re.split(r"###\s+", row[0])[1:]:
            m = re.match(r"(\S+)\s.*?ATR14:\s*([0-9.]+)", p, re.S)
            if m:
                out[m.group(1)] = float(m.group(2))
    return out


def assemble(trader_id, plan_ms, bars_all, plan_doc, htf_extra=None, atrs=None):
    """Full planner pipeline: 1m detectors + HTF (15m/1h/4h) + nPOC extras →
    ScoreLevelsMinGradeFull(prox 0.3, minGrade B, max 12) → Seat1HZone."""
    cb = closed_1m_slice(bars_all, plan_ms, 2000)
    price = cb[-1]["c"] if cb else 0.0
    datr = daily_range_proxy(cb, plan_ms)
    atr1 = wilder_atr14(cb)
    if datr <= 0:
        datr = 0.008 * price
    if atr1 <= 0:
        atr1 = datr / 20.0
    tol = 3 * TICK
    all_levels = []
    all_levels += multi_day_levels(cb, plan_ms)
    all_levels += round_number_levels(price, datr, 0.3)
    all_levels += opening_range_levels(cb, plan_ms)
    all_levels += gap_levels(cb, atr1)
    for l in eqhl(cb, tol):
        all_levels.append({"kind": l["kind"], "price": l["price"], "lo": l["price"],
                           "hi": l["price"], "label": l["label"]})
    for z in sd_zones(cb, atr1, plan_ms):
        all_levels.append({"kind": z["kind"], "price": z["mid"], "lo": z["lo"],
                           "hi": z["hi"], "label": z["label"], "pattern": z["pattern"],
                           "htf": False, "tf": "", "t": z["t"]})
    for f in fair_value_gaps(cb, max(2 * TICK, 2.0), plan_ms):
        all_levels.append({"kind": f["kind"], "price": f["mid"], "lo": f["lo"],
                           "hi": f["hi"], "label": f["label"], "htf": False, "tf": "", "t": f["t"]})
    for o in order_blocks(cb, atr1, plan_ms):
        all_levels.append({"kind": KOB, "price": o["mid"], "lo": o["lo"],
                           "hi": o["hi"], "label": o["label"], "htf": False, "tf": "", "t": o["t"]})
    for v in volume_levels(cb, plan_ms):
        all_levels.append({"kind": v["kind"], "price": v["price"], "lo": v["price"],
                           "hi": v["price"], "label": v["label"]})
    for s in swing_points(aggregate(cb, 5), 5, plan_ms) + swing_points(aggregate(cb, 15), 15, plan_ms):
        all_levels.append({"kind": s["kind"], "price": s["price"], "lo": s["price"],
                           "hi": s["price"], "label": s["label"]})
    # HTF pass (planner_timeframes = D,4h,1h,15m,5m → isHTFDetectionTF keeps 15m,1h,4h)
    # ATR input: the engine's detector series is the NT8-native tf cache (up to
    # 500 bars, pre-08-19 history included) — NOT reconstructible from the 1m
    # table alone. The stored indicators_block ATR14(tf) is the engine's own
    # series ATR → injected as the detector ATR (basis documented in report).
    htf_det = []
    for tf in ("15m", "1h", "4h"):
        tfb = aggregate(bars_all, 15 if tf == "15m" else 60 if tf == "1h" else 240,
                        plan_ms)
        if len(tfb) > 500:
            tfb = tfb[-500:]
        if len(tfb) < 5:
            continue
        atf = wilder_atr14(tfb)
        if atrs and tf in atrs and atrs[tf] > 0:
            atf = atrs[tf]  # engine's own series ATR (stored evidence)
        if atf <= 0:
            atf = 0.002 * tfb[-1]["c"]
        tol2 = max(3 * TICK, 0.15 * atf)
        for l in eqhl(tfb, tol2):
            htf_det.append({"kind": l["kind"], "price": l["price"], "lo": l["price"],
                            "hi": l["price"], "label": l["label"] + "·" + tf,
                            "htf": True, "tf": tf})
        for z in sd_zones(tfb, atf, plan_ms):
            htf_det.append({"kind": z["kind"], "price": z["mid"], "lo": z["lo"],
                            "hi": z["hi"], "label": z["label"] + "·" + tf,
                            "pattern": z["pattern"], "htf": True, "tf": tf, "t": z["t"]})
        for f in fair_value_gaps(tfb, atf, plan_ms):
            htf_det.append({"kind": f["kind"], "price": f["mid"], "lo": f["lo"],
                            "hi": f["hi"], "label": f["label"] + "·" + tf,
                            "htf": True, "tf": tf, "t": f["t"]})
        for o in order_blocks(tfb, atf, plan_ms):
            htf_det.append({"kind": KOB, "price": o["mid"], "lo": o["lo"],
                            "hi": o["hi"], "label": o["label"] + "·" + tf,
                            "htf": True, "tf": tf, "t": o["t"]})
    all_levels += htf_det
    if htf_extra:
        all_levels += htf_extra
    # dedupeSameKind within 1 tick
    ded = []
    for l in all_levels:
        dup = any(o["kind"] == l["kind"] and abs(o["price"] - l["price"]) <= TICK
                  for o in ded)
        if not dup:
            ded.append(l)
    freshness = make_freshness_fn(trader_id, plan_ms)
    pool24 = score_pool(ded, price, datr, freshness, 24, 0.3)
    filtered = [l for l in pool24 if l["grade"] in ("A", "B") or l["kind"] in TIER1]
    seated = filtered
    if len(filtered) > 12:
        seated = seat_htf(sorted(filtered, key=lambda x: (not today_priority(x["kind"]),
                                                          -x["score"], abs(x["distance"]), x["price"])), 12)
        seated = seat_volume_family(seated, 12)
        seated = seat_both_sides(seated, 12)
        seated = seated[:12]
    seated = seat_1h_zone(seated, 12)
    seated.sort(key=lambda x: abs(x["distance"]))
    return {"price": price, "datr": datr, "atr1m": atr1, "cb": cb,
            "pool": pool24, "seated": seated, "all": ded, "htf": htf_det,
            "n_closed": len(cb)}

# ----------------------------------------------------------------------------
# freshness replay from level_state
# ----------------------------------------------------------------------------
_level_state_rows = {}

def make_freshness_fn(trader_id, plan_ms):
    def fn(l):
        typ = level_type_from_label(l["label"])
        binidx = int(math.floor(l["price"] / 1.25))
        key = "%s|%s|%s|%s|%d" % (trader_id, SYMBOL, typ, "", binidx)
        rows = _level_state_rows.get(key, [])
        best = None
        for r in rows:
            # state evolves IN PLACE (updated_at) — only a row whose latest
            # update precedes the plan instant reflects the plan-time state.
            if r["created"] <= plan_ms and r["updated"] <= plan_ms and \
                    (best is None or r["created"] > best["created"]):
                best = r
        if best is None:
            return ""
        f = best["freshness"]
        return "" if f == "A" else f
    return fn

def level_type_from_label(label):
    L = label.upper()
    if "PDH" in L or "PDL" in L or "PDC" in L:
        return "PDH/PDL/PDC"
    if "ONH" in L or "ONL" in L:
        return "ONH/ONL"
    if "POC" in L:
        return "naked-POC"
    if "PW" in L or "PM" in L or "WK" in L or "MO" in L:
        return "prior-wk/mo"
    if "OR" in L or "IB" in L:
        return "opening-range"
    if "RN" in L:
        return "round-number"
    if "EQH" in L or "EQL" in L or "EQ" in L:
        return "equal-H/L"
    if "SWG" in L:
        return "swing-point"
    if any(x in L for x in ("FVG", "OB", "S/D", "SUPPLY", "DEMAND", "ZONE")):
        return "S/D+FVG/OB"
    return "other"

def load_level_state(db):
    rows = db.execute("SELECT level_key, price, times_tested, consumed, freshness, "
                      "created_at, updated_at FROM level_state").fetchall()
    out = {}
    for r in rows:
        key, price, tt, cons, fresh, ca, ua = r
        ca_ms = int(dt.datetime.fromisoformat(ca).timestamp() * 1000)
        ua_ms = int(dt.datetime.fromisoformat(ua).timestamp() * 1000)
        out.setdefault(key, []).append({"created": ca_ms, "updated": ua_ms,
                                        "freshness": fresh, "consumed": bool(cons)})
    return out

# ----------------------------------------------------------------------------
# census helpers
# ----------------------------------------------------------------------------
def match_stored(mine, stored_price, tick_tol=0.51):
    best = None
    for m in mine:
        d = abs(m["price"] - stored_price)
        if best is None or d < best[1]:
            best = (m, d)
    if best and best[1] <= tick_tol:
        return best[0], best[1]
    return None, None

def ticks(p):
    return round(p / TICK)

# ----------------------------------------------------------------------------
# consumed transition check (ConsumedSince, kernel/plan_lifecycle.go:180-186 +
# scenario_facts.go:384-400; rule 2x5m → 5m TF, need 2)
# ----------------------------------------------------------------------------
def consumed_since(bars_all, level, plan_ms, end_ms):
    """touched on the rule-TF (5m) series in-window AND not still-valid
    (2 consecutive 5m closes beyond on one side)."""
    w = [b for b in bars_all if plan_ms <= b["t"] < end_ms]
    five = aggregate(w, 5, end_ms)
    touched = any(b["l"] <= level <= b["h"] for b in five)
    if not touched:
        return False
    beyond_above = beyond_below = 0
    for b in reversed(five):
        if b["c"] > level:
            beyond_above += 1; beyond_below = 0
        elif b["c"] < level:
            beyond_below += 1; beyond_above = 0
        else:
            beyond_above = beyond_below = 0
    valid = beyond_above < 2 and beyond_below < 2
    return touched and not valid

# ----------------------------------------------------------------------------
# main
# ----------------------------------------------------------------------------
def main():
    db = sqlite3.connect(DB_PATH, uri=True)
    global _level_state_rows
    _level_state_rows = load_level_state(db)
    bars_all = load_bars(db)
    print("[A] bars loaded: %d MNQ 1m rows, %s → %s (UTC)" % (
        len(bars_all),
        dt.datetime.fromtimestamp(bars_all[0]["t"] / 1000, dt.timezone.utc),
        dt.datetime.fromtimestamp(bars_all[-1]["t"] / 1000, dt.timezone.utc)))

    SESSIONS_TO_VERIFY = [("2026-08-28", "NY"), ("2026-08-28", "LONDON"),
                          ("2026-08-27", "NY")]
    results = {}
    for trade_date, sess in SESSIONS_TO_VERIFY:
        row = db.execute(
            "SELECT version, doc, created_at FROM plans WHERE trade_date=? AND session=? "
            "ORDER BY version DESC LIMIT 1", (trade_date, sess)).fetchone()
        if not row:
            print("NO PLAN for", trade_date, sess); continue
        ver, doc, created = row
        plan_ms = int(dt.datetime.fromisoformat(created).timestamp() * 1000)
        d = json.loads(doc)
        trader_id = "8d5c8af5_8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek_1781246265"
        atrs = stored_atrs(db, trade_date, sess)
        res = assemble(trader_id, plan_ms, bars_all, d, atrs=atrs)
        results[(trade_date, sess)] = {"ver": ver, "plan_ms": plan_ms,
                                       "doc": d, "atrs": atrs, **res}
        print("[A] %s %s v%d @ %s price=%.2f dATR=%.2f atr1m=%.2f closed=%d pool=%d seated=%d"
              % (trade_date, sess, ver, created, res["price"], res["datr"],
                 res["atr1m"], res["n_closed"], len(res["pool"]), len(res["seated"])))

    # ---------------- CENSUS per class ----------------
    import zm_report
    report = zm_report.build_report(db, results, bars_all)
    os.makedirs(os.path.dirname(OUT_PATH), exist_ok=True)
    with open(OUT_PATH, "w") as f:
        f.write(report)
    print("[A] report written:", OUT_PATH)



if __name__ == "__main__":
    main()
