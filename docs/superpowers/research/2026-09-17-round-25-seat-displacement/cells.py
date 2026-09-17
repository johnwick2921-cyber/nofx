#!/usr/bin/env python3
"""R25 SEAT DISPLACEMENT — cells Q1–Q5 (DS-R001).

Consumes harness output out-r25/r25-reads.jsonl (one row per planner read with
the 12-seat reconstruction, the farthest in-band HTF level per side, and the
independently-recomputed big-day flag) + read-only sqlite queries on the DB
copy (plans table: scenario usage per level; bars: did the session reach the
farthest HTF level after the read).

Conventions (Q5 lesson — stated up front, kept fixed):
  SIDE: "above" = level.Price > read price; "below" = level.Price < read price.
        This is a PRICE side, never a trade direction. No trend labels are used
        in this round.
  BAND: band = proximityK * dATR, proximityK = 1.0 (the live
        day_plan.proximity_filter_atr read from the DB copy), dATR =
        kernel.DailyRangeProxy (harness, production call).
  BIG DAY: window complete AND session range >= 250 pt — recomputed here from
        the 1m window bars, independent of Chief's Q4 outputs.
Usage join: a plan scenario is "authored on" a level when the scenario's level
price (plans.doc levels[] by scenarios[].level_id; fallback confirm.ref_price /
arm.entry) is within 1 MNQ tick (0.25) of the seat's price for the same
(trade_date, session), latest plan version, lifecycle active.

Run: python3 cells.py <out-r25-dir> <db-copy-path>
"""
import json
import math
import os
import sqlite3
import sys
from collections import defaultdict

P_NULL = 0.5067
TICK = 0.25
BIG = 250.0


def wilson(h, n, z=1.959963984540054):
    if n == 0:
        return (0.0, 0.0)
    p = h / n
    den = 1 + z * z / n
    c = (p + z * z / (2 * n)) / den
    half = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / den
    return (c - half, c + half)


def main(outdir, dbpath):
    reads = [json.loads(l) for l in open(os.path.join(outdir, "r25-reads.jsonl"))]
    print(f"reads: {len(reads)}")
    big = [r for r in reads if r["big_day"]]
    print(f"big days (complete, range>=250): {len(big)}")

    # ── Q1: reconstruction coverage + farthest HTF presence ────────────────
    print("\n== Q1: reconstruction + farthest in-band HTF per side ==")
    with_fa = [r for r in reads if r.get("farthest_above")]
    with_fb = [r for r in reads if r.get("farthest_below")]
    print(f"  reads with farthest_above: {len(with_fa)}  farthest_below: {len(with_fb)}")
    med = lambda xs: round(sorted(xs)[len(xs)//2], 2) if xs else None
    print(f"  farthest_above dist median: {med([r['farthest_above']['dist_pts'] for r in with_fa])} pt")
    print(f"  farthest_below dist median: {med([r['farthest_below']['dist_pts'] for r in with_fb])} pt")
    print("  sample id:", " ".join(f"{r['day']} {r['session']}" for r in reads[:3]))

    # ── Q2: how often is the farthest HTF level already seated ─────────────
    print("\n== Q2: farthest in-band HTF level already seated? ==")
    for side in ("above", "below"):
        per_year = defaultdict(lambda: [0, 0])
        for r in reads:
            s = r.get("farthest_" + side)
            if not s:
                continue
            per_year[r["day"][:4]][0] += 1
            if s["seated"]:
                per_year[r["day"][:4]][1] += 1
        tot = sum(v[0] for v in per_year.values())
        sat = sum(v[1] for v in per_year.values())
        print(f"  {side}: seated {sat}/{tot} ({sat/tot:.1%})" if tot else f"  {side}: NOT MEASURED")
        for y in sorted(per_year):
            a, b = per_year[y]
            print(f"    {y}: {b}/{a} ({b/a:.1%})" if a else f"    {y}: NOT MEASURED")

    # ── Q3: displaced seat when not seated + usage cost ────────────────────
    print("\n== Q3: which seat gets displaced, and was it used ==")
    displaced = defaultdict(lambda: [0, 0])  # (kind,tf) -> [n, n_used]
    by_year = defaultdict(lambda: defaultdict(int))
    conn = sqlite3.connect(f"file:{dbpath}?mode=ro", uri=True)
    for r in reads:
        for side in ("above", "below"):
            s = r.get("farthest_" + side)
            if not s or s["seated"] or not r["seats"]:
                continue
            last = r["seats"][-1]  # lowest priority row under the sort rule
            key = (last["kind"], last["tf"])
            displaced[key][0] += 1
            by_year[r["day"][:4]][key] += 1
            # usage: scenario authored on the displaced level in that session's plan
            if level_used(conn, r["day"], r["session"], last["price"]):
                displaced[key][1] += 1
    for (k, tf), (n, u) in sorted(displaced.items(), key=lambda x: -x[1][0]):
        print(f"  {k} tf={tf or '-'}: displaced {n} times, used {u} ({u/n:.1%})" if n else "")
    print("  per year:")
    for y in sorted(by_year):
        parts = ", ".join(f"{k[0]}:{v}" for k, v in sorted(by_year[y].items(), key=lambda x: -x[1]))
        print(f"    {y}: {parts}")

    # ── Q4: on big days, was the farthest in-band HTF level the reached target ──
    print("\n== Q4: big-day reach of the farthest in-band HTF level ==")
    per_year = defaultdict(lambda: [0, 0])
    ids_reached, ids_missed = [], []
    for r in big:
        for side in ("above", "below"):
            s = r.get("farthest_" + side)
            if not s:
                continue
            reached = level_reached(conn, r, side, s["price"])
            y = r["day"][:4]
            per_year[y][0] += 1
            if reached:
                per_year[y][1] += 1
                if len(ids_reached) < 5:
                    ids_reached.append(f"{r['day']} {r['session']} {side}@{s['price']}")
            else:
                if len(ids_missed) < 5:
                    ids_missed.append(f"{r['day']} {r['session']} {side}@{s['price']}")
    tot = sum(v[0] for v in per_year.values())
    hit = sum(v[1] for v in per_year.values())
    print(f"  reached {hit}/{tot} ({hit/tot:.1%})" if tot else "  NOT MEASURED")
    for y in sorted(per_year):
        a, b = per_year[y]
        print(f"    {y}: {b}/{a} ({b/a:.1%})" if a else f"    {y}: NOT MEASURED")
    print("  reached ids:", ids_reached)
    print("  missed ids:", ids_missed)
    conn.close()


def level_used(conn, day, session, price):
    """A plan scenario authored on the level at (day, session), latest version."""
    rows = conn.execute(
        """SELECT doc FROM plans WHERE trade_date=? AND session=? AND lifecycle='active'
           ORDER BY version DESC LIMIT 1""", (day, session)).fetchall()
    for (doc,) in rows:
        try:
            d = json.loads(doc)
        except Exception:
            continue
        lv = {}
        for l in d.get("levels", []):
            if isinstance(l.get("price"), (int, float)) and l["price"] > 0:
                lv[l.get("id")] = l["price"]
        for sc in d.get("scenarios", []):
            p = lv.get(sc.get("level_id"))
            if not p:
                p = (sc.get("confirm") or {}).get("ref_price") or (sc.get("arm") or {}).get("entry")
            if isinstance(p, (int, float)) and abs(p - price) <= TICK:
                return True
    return False


FLAT_DELTA_MIN = {"LONDON": 7 * 60, "NY": 6 * 60 + 45, "ASIA": 9 * 60 + 30}


def level_reached(conn, r, side, price):
    """Did the session's 1m tape trade TO the level after the read (high>=price
    for above, low<=price for below) within [read, session flat)?"""
    flat_ms = r["read_at_ms"] + FLAT_DELTA_MIN[r["session"]] * 60_000
    q = conn.execute(
        """SELECT open_time_ms, h, l FROM bars WHERE symbol='MNQ' AND tf='1m'
           AND contract=? AND open_time_ms > ? AND open_time_ms < ?
           ORDER BY open_time_ms ASC""",
        (r["contract"], r["read_at_ms"], flat_ms))
    for t, h, l in q.fetchall():
        if side == "above" and h >= price:
            return True
        if side == "below" and l <= price:
            return True
    return False


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(2)
    main(sys.argv[1], sys.argv[2])
