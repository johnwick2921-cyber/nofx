#!/usr/bin/env python3
"""S4 analysis — measurement gate for the STRUCTURE-FIRST wave.

Consumes out-s4/{qa,episodes,trends}.jsonl and prints Q-A/Q-B/Q-C/Q-D tables
with n + Wilson CI + exact two-proportion p. Report rules: n per cell, baseline
p_null 0.5067, NOT MEASURED stated where n<200, no recommendation beyond cells.

Run: python3 docs/superpowers/research/2026-09-16-round-23/s4_analysis.py <out-s4-dir>
"""
import json
import math
import sys
from collections import defaultdict

NULL = 0.5067  # D1' IID calibration p(hold)
Z = 1.959963984540054

ZONE_KINDS = {"Supply", "Demand", "FVG", "IFVG", "OB"}  # isZoneKind verbatim
GRADES = ["fresh", "tested-1", "tested-2", "stale"]


def wilson(p, n):
    if n <= 0:
        return (0.0, 0.0)
    den = 1 + Z * Z / n
    c = (p + Z * Z / (2 * n)) / den
    half = Z * math.sqrt(p * (1 - p) / n + Z * Z / (4 * n * n)) / den
    return (c - half, c + half)


def two_prop(h1, n1, h2, n2):
    """Two-proportion z-test (pooled), two-sided. Returns (z, p)."""
    if n1 <= 0 or n2 <= 0:
        return 0.0, 1.0
    p1, p2 = h1 / n1, h2 / n2
    pp = (h1 + h2) / (n1 + n2)
    se = math.sqrt(pp * (1 - pp) * (1 / n1 + 1 / n2))
    if se == 0:
        return 0.0, 1.0
    z = (p1 - p2) / se
    # two-sided normal approximation
    from statistics import NormalDist
    p = 2 * (1 - NormalDist().cdf(abs(z)))
    return z, p


def load_jsonl(path):
    rows = []
    with open(path) as f:
        for line in f:
            if line.strip():
                rows.append(json.loads(line))
    return rows


def hold_of(rows):
    h = sum(1 for r in rows if r["outcome"] == "hold")
    b = sum(1 for r in rows if r["outcome"] == "break")
    return h, h + b


def cell(h, n):
    if n == 0:
        return "NOT MEASURED (n=0)"
    p = h / n
    lo, hi = wilson(p, n)
    return f"{p:.3f} [{lo:.3f},{hi:.3f}] n={n}"


def section(title):
    print(f"\n{'=' * 72}\n{title}\n{'=' * 72}")


def main(out):
    qa = load_jsonl(f"{out}/qa.jsonl")
    eps = load_jsonl(f"{out}/episodes.jsonl")
    trends = load_jsonl(f"{out}/trends.jsonl")
    print(f"qa rows: {len(qa)}  episodes: {len(eps)}  trend rows: {len(trends)}")

    # ── Q-A: reaction rate by freshness grade, BOTH modes ─────────────────
    section("Q-A: HTF level-scan ordinal-1 reaction rate by grade, 1m-touch vs own-TF")
    o1_qa = [r for r in qa]  # qa rows are ordinal-1 by construction
    for mode, key in [("1m-touch grading (today's live ladder A/B/C/done)", "grade_1m"),
                      ("own-TF grading (S2 levels_fresh_by_tf)", "grade_tf")]:
        print(f"\n-- {mode} --")
        by = defaultdict(list)
        for r in o1_qa:
            g = r.get(key) or "fresh"
            by[g].append(r)
        for g in GRADES:
            rows = by.get(g, [])
            h, n = hold_of(rows)
            tag = "" if n >= 200 else "  [NOT MEASURED n<200]"
            print(f"  {g:9s}  hold {cell(h, n)}{tag}")
        fr = by.get("fresh", [])
        st = by.get("stale", [])
        hf, nf = hold_of(fr)
        hs, ns = hold_of(st)
        if nf > 0 and ns > 0:
            z, p = two_prop(hf, nf, hs, ns)
            print(f"  fresh-vs-stale two-prop: z={z:+.2f} p={p:.4f} (fresh {cell(hf, nf)} | stale {cell(hs, ns)})")

    # cross-tab: grade_tf within each grade_1m bucket
    print("\n-- cross: own-TF grade within 1m-grade buckets --")
    cross = defaultdict(lambda: defaultdict(list))
    for r in o1_qa:
        cross[r.get("grade_1m") or "fresh"][r.get("grade_tf") or "fresh"].append(r)
    for g1 in GRADES:
        inner = cross.get(g1, {})
        parts = []
        for g2 in GRADES:
            h, n = hold_of(inner.get(g2, []))
            if n:
                parts.append(f"{g2}={h}/{n}({h/n:.3f})")
        print(f"  1m-grade {g1:9s}: " + " ".join(parts) if parts else f"  1m-grade {g1:9s}: (none)")

    # ── Q-B: direction agree/oppose vs S1 4h and D trend ──────────────────
    section("Q-B: ordinal-1 reaction by agreement with S1 structure state (4h and D)")
    tr = {(t["day"], t["session"]): t for t in trends}
    def direction(e):
        return "long" if e.get("entry") == "below" else "short"
    def agree(e, tftrend):
        d = direction(e)
        if tftrend == "up":
            return "agree" if d == "long" else "oppose"
        if tftrend == "down":
            return "agree" if d == "short" else "oppose"
        return "range"
    o1 = [e for e in eps if e["ordinal"] == 1]
    for tf in ["4h", "D"]:
        key = "h4_trend" if tf == "4h" else "d_trend"
        by = defaultdict(list)
        for e in o1:
            t = tr.get((e["day"], e["session"]))
            if t is None or t.get(key) is None:
                continue
            by[agree(e, t[key])].append(e)
        print(f"\n-- vs {tf} structure trend --")
        for g in ["agree", "oppose", "range"]:
            h, n = hold_of(by.get(g, []))
            tag = "" if n >= 200 else "  [NOT MEASURED n<200]"
            print(f"  {g:7s} hold {cell(h, n)}{tag}")
        ha, na = hold_of(by.get("agree", []))
        ho, no = hold_of(by.get("oppose", []))
        if na > 0 and no > 0:
            z, p = two_prop(ha, na, ho, no)
            print(f"  agree-vs-oppose two-prop: z={z:+.2f} p={p:.4f}")

    # ── Q-C: HTFScoreMultiplier ×1.2 separation ───────────────────────────
    section("Q-C: ×1.2 multiplier applies to non-zone kinds only (levels_score.go:512-517 [A]) — does the promoted group react differently?")
    def zone(e):
        return e["kind"] in ZONE_KINDS
    groups = {
        "HTF non-zone (×1.2 applies)": [e for e in o1 if e["htf"] and not zone(e)],
        "intraday non-zone (×1.0)": [e for e in o1 if not e["htf"] and not zone(e)],
        "HTF zone (×1.2 does NOT apply)": [e for e in o1 if e["htf"] and zone(e)],
        "intraday zone (×1.0)": [e for e in o1 if not e["htf"] and zone(e)],
    }
    for name, rows in groups.items():
        h, n = hold_of(rows)
        print(f"  {name:32s} hold {cell(h, n)}")
    r1, r2 = groups["HTF non-zone (×1.2 applies)"], groups["intraday non-zone (×1.0)"]
    h1, n1 = hold_of(r1)
    h2, n2 = hold_of(r2)
    z, p = two_prop(h1, n1, h2, n2)
    print(f"  HTF-nonzone vs intraday-nonzone two-prop: z={z:+.2f} p={p:.4f}")

    # ── Q-D: 5m zones vs 1m zones (both sit in the '1m' tier today) ───────
    section("Q-D: 5m vs 1m zone reaction (zoneTierFor maps 5m → '1m' tier today)")
    kind5 = sorted({e["kind"] for e in o1 if e["tf"] == "5m"})
    print(f"  kinds present at 5m: {kind5}")
    for scope, sel in [
        ("same kinds (5m kinds vs 1m same kinds)", lambda e: e["tf"] == "5m" or (e["tf"] == "1m" and e["kind"] in kind5)),
        ("all kinds at each tf", lambda e: e["tf"] in ("5m", "1m")),
    ]:
        m5 = [e for e in o1 if e["tf"] == "5m" and sel(e)]
        m1 = [e for e in o1 if e["tf"] == "1m" and sel(e)]
        h5, n5 = hold_of(m5)
        h1, n1 = hold_of(m1)
        z, p = two_prop(h5, n5, h1, n1)
        print(f"  [{scope}]")
        print(f"    5m: {cell(h5, n5)}   1m: {cell(h1, n1)}   two-prop z={z:+.2f} p={p:.4f}")
    # distance-matched: 5m vs 1m same-kind in each dist bucket
    print("\n  distance-bucket matched (same kinds):")
    for lo, hi in [(0, 25), (25, 50), (50, 100), (100, 200), (200, 10 ** 9)]:
        m5 = [e for e in o1 if e["tf"] == "5m" and (lo <= e["dist_at_read"] < hi)]
        m1 = [e for e in o1 if e["tf"] == "1m" and e["kind"] in kind5 and (lo <= e["dist_at_read"] < hi)]
        h5, n5 = hold_of(m5)
        h1, n1 = hold_of(m1)
        if n5 == 0 and n1 == 0:
            continue
        print(f"    dist {lo:>3}-{hi:<4}: 5m {cell(h5, n5)}  1m {cell(h1, n1)}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print(__doc__)
        sys.exit(2)
    main(sys.argv[1])
