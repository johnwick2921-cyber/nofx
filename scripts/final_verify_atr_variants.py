#!/usr/bin/env python3
"""Variant hunt: reproduce engine ATR14=12.0658 (3m section, 2026-08-27 15:58)."""
import sys
sys.path.insert(0, "scripts")
from e2e_recompute import *


def load_prompt(dt_prefix):
    c = db().cursor()
    r = c.execute(
        "SELECT input_prompt FROM decision_records WHERE input_prompt LIKE '%ATR14:%' "
        "AND timestamp LIKE ? ORDER BY timestamp DESC LIMIT 1", (dt_prefix + "%",)
    ).fetchone()
    return r[0] if r else None


p = load_prompt("2026-08-27")
cut, atr_q, rsi_q, ema_q = parse_prompt_cut_and_atr(p)
cut_ms = ct_ms(cut)
TARGET = atr_q
print("engine ATR14:", TARGET)
b1 = bars_1m("MNQ", cut_ms - 6 * 3600_000, cut_ms)


def TR(bars):
    trs = []
    for i in range(1, len(bars)):
        h, l, pc = bars[i][2], bars[i][3], bars[i - 1][4]
        trs.append(max(h - l, abs(h - pc), abs(l - pc)))
    return trs


for tf, dur in [("3m", 180_000), ("5m", 300_000), ("15m", 900_000), ("1h", 3600_000)]:
    agg = {b[0]: b for b in aggregate(b1, dur)}
    ser = [agg[k] for k in sorted(agg) if k + dur <= cut_ms]
    trs = TR(ser)
    if len(trs) < 14:
        continue
    wilder = sum(trs[:14]) / 14
    for t in trs[14:]:
        wilder = (wilder * 13 + t) / 14
    sma14 = sum(trs[-14:]) / 14
    ema14 = trs[0]
    for t in trs[1:]:
        ema14 = t * (2 / 15) + ema14 * (13 / 15)
    print(f"{tf}: wilder={wilder:.4f} sma14={sma14:.4f} ema14={ema14:.4f} n={len(ser)}")
# 3m with the forming bar included
agg = {b[0]: b for b in aggregate(b1, 180_000)}
ser = [agg[k] for k in sorted(agg) if k <= cut_ms]  # includes forming 15:57
trs = TR(ser)
wilder = sum(trs[:14]) / 14
for t in trs[14:]:
    wilder = (wilder * 13 + t) / 14
print(f"3m incl-forming: wilder={wilder:.4f}")
# 3m shifted boundaries (+1m / -1m grouping)
for shift in (-60000, 60000):
    agg = {}
    for t, o, h, l, c, v in b1:
        k = t + shift - (t + shift) % 180_000
        if k in agg:
            z = agg[k]; agg[k] = [k, z[1], max(z[2], h), min(z[3], l), c, z[5] + v]
        else:
            agg[k] = [k, o, h, l, c, v]
    ser = [agg[k] for k in sorted(agg) if k + 180_000 <= cut_ms]
    trs = TR(ser)
    wilder = sum(trs[:14]) / 14
    for t in trs[14:]:
        wilder = (wilder * 13 + t) / 14
    print(f"3m shift={shift//60000:+d}m: wilder={wilder:.4f} n={len(ser)}")
