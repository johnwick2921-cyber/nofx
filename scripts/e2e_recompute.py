#!/usr/bin/env python3
"""FINAL VERIFICATION SWEEP v2 — independent recompute helpers (R2: never calls
the engine's functions; replicates the documented Wilder math from the raw
bars table only). Committed as evidence tooling on docs/final-verify."""
import sqlite3, re, sys, datetime

DB = "file:/home/hoang/nofx/data/data.db?mode=ro"

def db():
    return sqlite3.connect(DB, uri=True)

def wilder_atr(bars, period=14):
    """Wilder ATR: TR = max(H-L, |H-pc|, |L-pc|); seed = SMA(TR, period);
    then RMA. Mirrors market/data_indicators.go:86-115 (read, not called)."""
    trs = [0.0]
    for i in range(1, len(bars)):
        h, l, pc = bars[i][2], bars[i][3], bars[i - 1][4]
        trs.append(max(h - l, abs(h - pc), abs(l - pc)))
    if len(trs) <= period:
        return 0.0
    atr = sum(trs[1:period + 1]) / period
    for i in range(period + 1, len(trs)):
        atr = (atr * (period - 1) + trs[i]) / period
    return atr

def rsi14(closes):
    if len(closes) < 15:
        return 0.0
    gains, losses = [], []
    for i in range(1, len(closes)):
        d = closes[i] - closes[i - 1]
        gains.append(max(d, 0.0)); losses.append(max(-d, 0.0))
    ag = sum(gains[:14]) / 14; al = sum(losses[:14]) / 14
    for i in range(14, len(gains)):
        ag = (ag * 13 + gains[i]) / 14
        al = (al * 13 + losses[i]) / 14
    if al == 0:
        return 100.0
    return 100 - 100 / (1 + ag / al)

def ema(closes, period):
    k = 2 / (period + 1)
    e = closes[0]
    for c in closes[1:]:
        e = c * k + e * (1 - k)
    return e

def bars_1m(symbol, since_ms, until_ms):
    c = db().cursor()
    c.execute("SELECT open_time_ms,o,h,l,c,v FROM bars WHERE symbol=? AND tf='1m' "
              "AND open_time_ms>=? AND open_time_ms<? ORDER BY open_time_ms", (symbol, since_ms, until_ms))
    return [(r[0], r[1], r[2], r[3], r[4], r[5]) for r in c.fetchall()]

def aggregate(bars_1m, dur_ms):
    out = {}
    for t, o, h, l, c, v in bars_1m:
        k = t - t % dur_ms
        if k in out:
            p = out[k]; out[k] = [k, p[1], max(p[2], h), min(p[3], l), c, p[5] + v]
        else:
            out[k] = [k, o, h, l, c, v]
    return [out[k] for k in sorted(out)]

def parse_prompt_cut_and_atr(prompt_text):
    t = re.search(r"Time: (\d{4}-\d{2}-\d{2} \d{2}:\d{2}) CT", prompt_text)
    a = re.search(r"ATR14: ([\d.]+)", prompt_text)
    r = re.search(r"RSI14: \[([^\]]*)\]", prompt_text)
    e50 = re.search(r"EMA50: \[([^\]]*)\]", prompt_text)
    cut = None
    if t:
        cut = datetime.datetime.strptime(t.group(1), "%Y-%m-%d %H:%M")
    atr = float(a.group(1)) if a else None
    rsi_last = float(r.group(1).split(",")[-1].strip()) if r else None
    ema_last = float(e50.group(1).split(",")[-1].strip()) if e50 else None
    return cut, atr, rsi_last, ema_last

def ct_ms(dt):
    return int(dt.replace(tzinfo=datetime.timezone(datetime.timedelta(hours=-5))).timestamp() * 1000)
