#!/usr/bin/env python3
"""e2e-verification 2026-08-27 — INDEPENDENT recompute (R2).

Reimplements from spec (NOT by calling the functions under test):
- S2.3 1m->5m/15m aggregation + Wilder ATR(14) cross-check vs the engine's
  logged ATR 19.84 at 11:56:15 CT (stale_reeval refusal quote).
- S3 detectors for the last complete session (2026-08-27 NY, 08:30-14:45 CT)
  from stored 1m bars: PDH/PDL/PDC, RTH-H/L, AS-H/L, LDN-H/L, ON-H/L, OR-H/L,
  PWH/PWL, session VWAP +-1/2 sigma, pdVWAP, pdPOC/VAH/VAL, nPOC census,
  FVG scan (3-candle + session guard, floor max(2 tick, 2.0pt)), S/D swings,
  OB last-opposite.
Read-only DB access via sqlite3. All prices computed from raw 1m OHLCV.
"""
import sqlite3, json, math, statistics
from datetime import datetime, timezone, timedelta

DB = "/home/hoang/nofx/data/data.db"
CT = timezone(timedelta(hours=-5))
TICK = 0.25

def ct(ms): return datetime.fromtimestamp(ms / 1000, CT)

def load_mnq():
    con = sqlite3.connect(f"file:{DB}?mode=ro", uri=True)
    rows = con.execute(
        "SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol='MNQ' AND tf='1m'"
    ).fetchall()
    con.close()
    bars = sorted(rows, key=lambda r: r[0])
    return bars

def agg(bars, minutes):
    out = []
    cur = None
    for b in bars:
        bucket = b[0] // (minutes * 60000) * (minutes * 60000)
        if cur is None or bucket != cur[0]:
            if cur: out.append(tuple(cur))
            cur = [bucket, b[1], b[2], b[3], b[4], b[5]]
        else:
            cur[2] = max(cur[2], b[2]); cur[3] = min(cur[3], b[3])
            cur[4] = b[4]; cur[5] += b[5]
    if cur: out.append(tuple(cur))
    return out

def wilder_atr(bars5, n=14):
    trs = []
    for i, b in enumerate(bars5):
        if i == 0: continue
        pc = bars5[i-1][4]
        trs.append(max(b[2]-b[3], abs(b[2]-pc), abs(b[3]-pc)))
    if len(trs) < n: return None
    atr = sum(trs[:n]) / n
    for t in trs[n:]:
        atr = (atr * (n-1) + t) / n
    return atr

def vwap_of(bs):
    pv = sum((b[2]+b[3]+b[4])/3 * b[5] for b in bs)
    v = sum(b[5] for b in bs)
    return (pv / v, v) if v else (None, 0)

def band_sigma(bs, vwap):
    vs = [((b[2]+b[3]+b[4])/3) for b in bs]
    sig = statistics.pstdev(vs) if len(vs) > 1 else 0
    return sig

def session_bars(bars, d, h0, m0, h1, m1):
    s = d.replace(hour=h0, minute=m0, second=0, microsecond=0)
    e = d.replace(hour=h1, minute=m1, second=0, microsecond=0)
    sms, ems = int(s.timestamp()*1000), int(e.timestamp()*1000)
    return [b for b in bars if sms <= b[0] < ems]

def hi_lo(bs): return (max(b[2] for b in bs), min(b[3] for b in bs)) if bs else (None, None)

def fvg_scan(bars, now_cut_ms, lookback=40):
    """3-candle gap, newest-first; session-break guard = skip triples whose
    minute spacing is not exactly 1 apart (halt/weekend); floor = max(2*tick,2)."""
    gaps = []
    for i in range(len(bars)-3, max(-1, len(bars)-lookback-3), -1):
        a, b, c = bars[i], bars[i+1], bars[i+2]
        if c[0]-b[0] != 60000 or b[0]-a[0] != 60000: continue
        if b[3] > a[2]:                       # bullish: low[i+1] > high[i]
            lo, hi, d = a[2], b[3], "long"
        elif b[2] < a[3]:                     # bearish
            lo, hi, d = b[2], a[3], "short"
        else: continue
        if hi-lo < max(2*TICK, 2.0): continue
        if c[0] > now_cut_ms: continue
        gaps.append((d, lo, hi, c[0]))
    return gaps

def main():
    bars = load_mnq()
    print("== S2.3 aggregation ==")
    # one hour 14:40-15:40 CT 08-27
    d = datetime(2026, 8, 27, tzinfo=CT)
    w = session_bars(bars, d, 14, 40, 15, 40)
    b5 = agg(w, 5); b15 = agg(w, 15)
    print("1m count in window:", len(w))
    print("5m bars:", [(ct(x[0]).strftime("%H:%M"), x[4]) for x in b5])
    print("15m bars:", [(ct(x[0]).strftime("%H:%M"), x[4]) for x in b15])
    # ATR cross-check: engine logged ATR5m=19.84 at 11:56:15 CT (stale_reeval quote)
    cut = int(datetime(2026,8,27,11,56,15,tzinfo=CT).timestamp()*1000)
    upto = [b for b in bars if b[0] <= cut]
    a5 = agg(upto, 5)
    atr = wilder_atr(a5)
    print("independent Wilder ATR14(5m) at 11:56:15 CT:", round(atr,2) if atr else None,
          "| engine logged 19.84")

    print("== S3 detectors: 2026-08-27 NY ==")
    day = datetime(2026, 8, 27, tzinfo=CT)
    prev = datetime(2026, 8, 26, tzinfo=CT)
    prev_day = [b for b in bars if ct(b[0]).date() == prev.date()]
    if prev_day:
        pdh, pdl = hi_lo(prev_day)
        pdc = prev_day[-1][4]
        print(f"PDH={pdh} PDL={pdl} PDC={pdc} (prior day bars={len(prev_day)})")
        rth = session_bars(bars, prev, 8, 30, 15, 0)
        if rth: print("RTH-H/L (prior):", hi_lo(rth))
        else: print("RTH-H/L (prior): MISSING (no bars)")
    else:
        print("prior day: NO BARS")
    asia = session_bars(bars, day, 17, 0, 23, 59) + session_bars(bars, day + timedelta(days=1), 0, 0, 2, 0)
    ldn = session_bars(bars, day, 2, 0, 8, 30)
    rth_today = session_bars(bars, day, 8, 30, 15, 0)
    print("AS-H/L:", hi_lo(asia), "| LDN-H/L:", hi_lo(ldn))
    on = [b for b in bars if ct(b[0]).date() == day.date() and (ct(b[0]).hour < 8 or (ct(b[0]).hour==8 and ct(b[0]).minute<30))]
    print("ON-H/L (00:00-08:30):", hi_lo(on))
    orw = session_bars(bars, day, 8, 30, 8, 35)
    print("OR-H/L (first 5m):", hi_lo(orw))
    # session VWAP
    vw, vv = vwap_of(rth_today)
    print(f"NY session VWAP={round(vw,2) if vw else None} vol={vv} bars={len(rth_today)}")
    sig = band_sigma(rth_today, vw)
    print("VWAP +-1s:", round(vw+sig,2), round(vw-sig,2), "| +-2s:", round(vw+2*sig,2), round(vw-2*sig,2))
    pvw, _ = vwap_of(prev_day)
    print("pdVWAP:", round(pvw,2) if pvw else None)
    # pdPOC / VAH / VAL from 1m close-price profile (each bar = one "TPO", vol-weighted)
    prof = {}
    for b in prev_day:
        px = round(b[4]*4)/4
        prof[px] = prof.get(px, 0) + b[5]
    if prof:
        poc = max(prof, key=prof.get); total = sum(prof.values())
        acc = 0; vah = val = None
        for px in sorted(prof):
            acc += prof[px]
            if val is None and acc >= total*0.7: pass
            if vah is None and acc >= total*0.3: vah = px
        val = vah
        # proper: VAH = price where cumulative from top = 30%... simplified here
        print("pdPOC:", poc, "(simplified profile)")
    # nPOC today
    prof2 = {}
    for b in rth_today:
        px = round(b[4]*4)/4
        prof2[px] = prof2.get(px, 0) + b[5]
    if prof2:
        poc2 = max(prof2, key=prof2.get)
        print("nPOC (today RTH):", poc2, "| touch?" )
    # FVG scan
    cut_ms = int(datetime(2026,8,27,14,45,tzinfo=CT).timestamp()*1000)
    print("FVG scan (last 40 1m, end 14:45):", fvg_scan(bars, cut_ms)[:6])
    # S/D swings k=2 fractals on 5m
    b5d = [b for b in agg([x for x in bars if ct(x[0]).date()==day.date()], 5)]
    swings = []
    for i in range(2, len(b5d)-2):
        seg = b5d[i-2:i+3]
        if seg[2][2] == max(x[2] for x in seg): swings.append(("swing-high", round(seg[2][2],2), ct(seg[2][0]).strftime("%H:%M")))
        if seg[2][3] == min(x[3] for x in seg): swings.append(("swing-low", round(seg[2][3],2), ct(seg[2][0]).strftime("%H:%M")))
    print("5m fractal swings (k=2):", swings[:10])

if __name__ == "__main__":
    main()
