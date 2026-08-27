#!/usr/bin/env python3
"""P1.2 final: 3 cuts x engine ATR14(3m) vs independent recompute, forming bar
included (the engine includes it — proven 2026-08-27)."""
import sys, re
sys.path.insert(0, "scripts")
from e2e_recompute import *


def load_prompt(tm_filter, order="DESC"):
    c = db().cursor()
    sql = ("SELECT input_prompt FROM decision_records WHERE input_prompt LIKE '%ATR14:%' "
           "AND timestamp LIKE ? ORDER BY timestamp " + order + " LIMIT 1")
    r = c.execute(sql, (tm_filter + "%",)).fetchone()
    return r[0] if r else None


def probe(label, p):
    cut, atr_q, rsi_q, ema_q = parse_prompt_cut_and_atr(p)
    cut_ms = ct_ms(cut)
    b1 = bars_1m("MNQ", cut_ms - 6 * 3600_000, cut_ms)
    agg = {b[0]: b for b in aggregate(b1, 180_000)}
    ser = [agg[k] for k in sorted(agg) if k + 180_000 <= cut_ms]
    # parse the table's current/forming row (last row of the 3m table)
    rows3 = re.findall(r"(08-2[67])\s+(\d{1,2}:\d{2})\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)", p)
    forming_ms = cut_ms - cut_ms % 180_000  # floor to the 3m boundary
    forming = None
    for d, hm, o, h, l, cc, v in rows3:
        mm, ss = hm.split(":")
        ms = ct_ms(datetime.datetime(2026, 8, int(d.split("-")[1]), int(mm), int(ss)))
        if ms == forming_ms and (not ser or ms > ser[-1][0]):
            forming = [ms, float(o), float(h), float(l), float(cc), float(v)]
    closed_a = wilder_atr(ser)
    if forming is not None:
        ser2 = ser + [forming]
        incl_a = wilder_atr(ser2)
    else:
        incl_a = closed_a
    m1c = [b for b in b1 if b[0] + 60_000 <= cut_ms]
    a1 = wilder_atr(m1c)
    print(f"{label}: cut={cut} engineATR={atr_q}")
    print(f"   closed-only={closed_a:.4f} incl-forming={incl_a:.4f} "
          f"diff_incl={abs(incl_a-atr_q):.4f} {'2dp-MATCH' if abs(incl_a-atr_q)<=0.005 else 'MISMATCH'}")
    print(f"   twin ATR14(1m)={a1:.4f}")


probe("CUT-A 08-27 mid-RTH", load_prompt("2026-08-27 1", "ASC"))
probe("CUT-B 08-27 last", load_prompt("2026-08-27"))
probe("CUT-C 08-26", load_prompt("2026-08-26"))
