#!/usr/bin/env python3
"""Find the engine's exact 3m series window: brute-force K (series length) +
session-slice variants so RSI14/EMA50/ATR14 all reproduce the engine's quotes."""
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
b1 = bars_1m("MNQ", cut_ms - 8 * 3600_000, cut_ms)
m3 = {b[0]: b for b in aggregate(b1, 180_000)}
m3c = [m3[k] for k in sorted(m3) if k + 180_000 <= cut_ms]
print(f"engine: ATR14={atr_q} RSI14last={rsi_q} EMA50last={ema_q}  (n3m={len(m3c)})")
print("--- brute force K (series = last K 3m bars) ---")
best = []
for K in range(15, len(m3c) + 1):
    s = m3c[-K:]
    a = wilder_atr(s)
    r = rsi14([b[4] for b in s])
    e = ema([b[4] for b in s], 50)
    score = abs(a - atr_q) + abs(r - rsi_q) / 10 + abs(e - ema_q) / 100
    best.append((score, K, a, r, e))
best.sort()
for score, K, a, r, e in best[:5]:
    print(f"  K={K:4d} score={score:.4f} ATR={a:.4f} RSI={r:.4f} EMA={e:.4f}")
print("--- session-slice variants ---")
for hh, mm in [(8, 30), (9, 30), (12, 0), (17, 0)]:
    sess_ms = ct_ms(datetime.datetime(2026, 8, 27, hh, mm))
    s = [b for b in m3c if b[0] >= sess_ms]
    a = wilder_atr(s)
    r = rsi14([b[4] for b in s])
    e = ema([b[4] for b in s], 50)
    print(f"  since {hh:02d}:{mm:02d}: n={len(s)} ATR={a:.4f} RSI={r:.4f} EMA={e:.4f}")
print("--- prev-day roll: series ending at cut, starting at prev session 17:00 ---")
for day, hh, mm in [(26, 17, 0), (26, 18, 0), (26, 19, 0)]:
    sess_ms = ct_ms(datetime.datetime(2026, 8, day, hh, mm))
    s = [b for b in m3c if b[0] >= sess_ms]
    a = wilder_atr(s)
    r = rsi14([b[4] for b in s])
    e = ema([b[4] for b in s], 50)
    print(f"  since 08-{day} {hh:02d}:{mm:02d}: n={len(s)} ATR={a:.4f} RSI={r:.4f} EMA={e:.4f}")
