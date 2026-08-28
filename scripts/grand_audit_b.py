#!/usr/bin/env python3
"""GRAND AUDIT 2026-08-28 — PART B engine (read-only).

B1: census ALL gate refusals (arm-gate from log_events; decision-gate from
    decision_records.risk_check_error) + ALL fresh-MET AI-declines, replayed on
    1m bars. Outcome: WON (TP-first) / LOST (SL-first) / AMBIG (same-bar, counts
    AGAINST) / EOD (neither by 14:45 CT flat).
B2: armed-era split (pre/post arm-mandate 2026-08-27 09:25 CT) + $1,763 leak
    gauge re-measured at the wider window.
B3: per-condition expectancy on TRUE pnl_corrected (position_plan_join).
B4: R-floor sweep per play at 1.5/2.0/2.5/3.0.

Usage: python3 scripts/grand_audit_b.py
"""
import json
import re
import sqlite3
from collections import defaultdict
from datetime import datetime, timezone, timedelta

CT = timezone(timedelta(hours=-5))
UTC = timezone.utc
DB = sqlite3.connect("file:/home/hoang/nofx/data/data.db?mode=ro", uri=True)

W0 = datetime(2026, 8, 26, 0, 0, tzinfo=CT)          # B1 window start
ARM_MANDATE = datetime(2026, 8, 27, 9, 25, tzinfo=CT)  # armed cutover (memory: 09:25 CT)
NOW = datetime.now(CT)
W0ms, AMms, NOWms = (int(W0.timestamp() * 1000), int(ARM_MANDATE.timestamp() * 1000),
                     int(NOW.timestamp() * 1000))

# ---- bars: 1m MNQ ----
bars = {}
for r in DB.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' ORDER BY open_time_ms"):
    bars[r[0]] = r[1:]
if not bars:
    raise SystemExit("no bars")

def bar_at(t_ms):
    return bars.get(t_ms - t_ms % 60000)

def eod_for(t_ms):
    d = datetime.fromtimestamp(t_ms / 1000, CT)
    flat = d.replace(hour=14, minute=45, second=0, microsecond=0)
    if d >= flat:
        flat += timedelta(days=1)
    return int(flat.timestamp() * 1000)

def ctime(ms):
    return datetime.fromtimestamp(ms / 1000, CT).strftime("%m-%d %H:%M")

def planned_r(side, e, sl, tp):
    risk, reward = abs(e - sl), abs(tp - e)
    return reward / risk if risk > 0 else 0.0

def replay(t_ms, side, entry, sl, tp):
    """1m-bar replay to next 14:45 CT flat. AMBIG same-bar counts AGAINST."""
    eod = eod_for(t_ms)
    keys = sorted(k for k in bars if t_ms - 60000 < k <= eod)
    if not keys:
        return "EOD", 0.0
    for k in keys:
        b = bars[k]
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
    last = bars[keys[-1]][3]
    p = (last - entry) if side == "long" else (entry - last)
    return "EOD", p

# ---- plan resolver: active version of session at time t ----
_plan_cache = {}
def plan_doc_at(t_ms, session):
    d = datetime.fromtimestamp(t_ms / 1000, CT)
    key = (d.strftime("%Y-%m-%d"), session)
    if key in _plan_cache:
        return _plan_cache[key]
    pat = key[0] + "%:" + session + ":%"
    row = DB.execute(
        "SELECT doc FROM plans WHERE plan_id LIKE ? AND created_at <= ? "
        "ORDER BY created_at DESC LIMIT 1",
        (pat, datetime.fromtimestamp(t_ms / 1000, UTC).strftime("%Y-%m-%d %H:%M:%S"))).fetchone()
    doc = json.loads(row[0]) if row else None
    _plan_cache[key] = doc
    return doc

def scenario_arm(doc, sid):
    if not doc:
        return None
    for s in doc.get("scenarios", []):
        if s.get("id") == sid:
            a = s.get("arm") or {}
            return dict(side=s.get("direction", "long"), entry=a.get("entry"),
                        sl=a.get("stop"), tp=a.get("target"), cond=s.get("condition"),
                        quality=s.get("quality"))
    return None

# ============================================================
# CENSUS A — arm-gate refusals (log_events, deduped)
# ============================================================
def refusal_class(reason):
    r = reason.lower()
    if r.startswith("r:r"):
        return "arm_RR"
    if "too close" in r:
        return "arm_minSL"
    if "veto" in r:
        return "arm_HTFveto"
    return "arm_" + re.split(r"[: ]", r)[0][:20]

arm_rows = []
for ts_utc, msg in DB.execute(
        "SELECT ts_utc, message FROM log_events WHERE message LIKE '%arm REFUSED%' "
        "AND ts_utc BETWEEN ? AND ? ORDER BY ts_utc", (W0ms, NOWms)):
    m = re.search(r"arm REFUSED (\S+) (S\d+): (.+)$", msg)
    if not m:
        continue
    session, sid, reason = m.group(1), m.group(2), m.group(3).strip()
    cls = refusal_class(reason)
    doc = plan_doc_at(ts_utc, session)
    sc = scenario_arm(doc, sid)
    if not sc or not sc["entry"] or not sc["sl"] or not sc["tp"]:
        continue
    arm_rows.append(dict(t=ts_utc, session=session, sid=sid, cls=cls, reason=reason, **sc))

# dedup: first occurrence per (session, sid, cls)
seen, arm_dedup = set(), []
for c in arm_rows:
    k = (c["session"], c["sid"], c["cls"])
    if k in seen:
        continue
    seen.add(k)
    arm_dedup.append(c)

# ============================================================
# CENSUS B — decision-gate refusals (decision_records)
# ============================================================
dec_rows = []
for rid, ts, dj, err in DB.execute(
        "SELECT id, timestamp, decision_json, risk_check_error FROM decision_records "
        "WHERE timestamp >= ? AND risk_check_error != '' AND risk_check_error != 'superseded_wait' "
        "ORDER BY id", (W0.strftime("%Y-%m-%d %H:%M:%S"),)):
    try:
        dec = json.loads(dj)[0]
    except Exception:
        continue
    if not dec.get("stop_loss") or not dec.get("take_profit"):
        continue
    t_ms = int(datetime.fromisoformat(ts).timestamp() * 1000)
    if not (W0ms <= t_ms <= NOWms):
        continue
    b = bar_at(t_ms)
    if not b:
        continue
    side = "long" if dec["action"] == "open_long" else "short"
    gate = err.split(":")[0]
    dec_rows.append(dict(t=t_ms, session="-", sid="-", cls=gate, reason=err,
                         side=side, entry=b[3], sl=dec["stop_loss"],
                         tp=dec["take_profit"], cond="(gate)", quality=""))

# ============================================================
# CENSUS C — fresh-MET AI-declines (superseded_wait + fresh-MET)
# ============================================================
decl_rows = []
denom = 0
for rid, ts, dj, sp, pid, pv in DB.execute(
        "SELECT id, timestamp, decision_json, system_prompt, plan_id, plan_version "
        "FROM decision_records WHERE timestamp >= ? AND risk_check_error='superseded_wait' "
        "ORDER BY id", (W0.strftime("%Y-%m-%d %H:%M:%S"),)):
    if not sp or "CONFLICT" in sp:
        continue
    mets = re.findall(r"(S\d+) confirm: .*? (MET|NOT MET)(?: \((stale|.*?)\))?", sp)
    fresh = [m[0] for m in mets if m[1] == "MET" and m[2] != "stale"]
    if not fresh:
        continue
    t_ms = int(datetime.fromisoformat(ts).timestamp() * 1000)
    if not (W0ms <= t_ms <= NOWms):
        continue
    denom += 1
    row = DB.execute("SELECT doc FROM plans WHERE plan_id=? AND version=?", (pid, pv)).fetchone()
    if not row:
        continue
    doc = json.loads(row[0])
    for sid in fresh:
        sc = next((s for s in doc.get("scenarios", []) if s.get("id") == sid), None)
        if not sc:
            continue
        conf = sc.get("confirm") or {}
        entry = conf.get("ref_price")
        side = sc.get("direction", "long")
        chain = sc.get("target_chain") or []
        nums = [float(x) for x in re.findall(r"[\d]+(?:[.][\d]+)?", sc.get("invalid", ""))
                if float(x) > 1000]
        sl = nums[-1] if nums else None
        tp = chain[0] if chain else None
        if not entry or not sl or not tp or abs(sl - entry) < 2.0:
            continue
        if side == "long" and not (sl < entry < tp):
            continue
        if side == "short" and not (tp < entry < sl):
            continue
        decl_rows.append(dict(t=t_ms, session="-", sid=sid, cls="decline", reason="fresh-MET wait",
                              side=side, entry=entry, sl=sl, tp=tp,
                              cond=sc.get("condition", "?"), quality=sc.get("quality", "")))

seen_d, decl_dedup = set(), []
for c in decl_rows:
    k = (c["sid"], c["side"], c["entry"], c["sl"], c["tp"])
    if k in seen_d:
        continue
    seen_d.add(k)
    decl_dedup.append(c)

# ============================================================
# REPLAY + TABLES
# ============================================================
def table(title, rows, key="cls", label=lambda c: c.get("cond") or c["cls"]):
    per = defaultdict(lambda: dict(n=0, won=0, lost=0, ambig=0, eod=0, s=0.0, rows=[]))
    for c in rows:
        out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
        g = per[label(c)]
        g["n"] += 1
        g["rows"].append((c, out, pnl))
        if out == "WON":
            g["won"] += 1
        elif out == "AMBIG":
            g["ambig"] += 1
            g["s"] += pnl * 2.0   # resolved AGAINST
        elif out == "LOST":
            g["lost"] += 1
            g["s"] += pnl * 2.0
        else:
            g["eod"] += 1
            g["s"] += pnl * 2.0
        if out == "WON":
            g["s"] += pnl * 2.0
    print("\n== %s ==" % title)
    print("%-16s %3s %4s %4s %5s %3s %9s  %s" % ("class", "n", "won", "lost", "ambig", "eod", "S$", "verdict"))
    for k in sorted(per):
        g = per[k]
        v = "SAVING" if (g["n"] >= 5 and g["s"] < 0) else ("COSTING" if (g["n"] >= 5 and g["s"] > 0) else "TOO-FEW")
        print("%-16s %3d %4d %4d %5d %3d %9.1f  %s" % (k, g["n"], g["won"], g["lost"], g["ambig"], g["eod"], g["s"], v))
    return per

print("WINDOW: %s -> %s CT | bars %d" % (W0.strftime("%m-%d %H:%M"), NOW.strftime("%m-%d %H:%M"), len(bars)))
print("arm-gate rows raw=%d deduped=%d | decision-gate rows=%d | decline rows raw=%d deduped=%d denom-cycles=%d"
      % (len(arm_rows), len(arm_dedup), len(dec_rows), len(decl_rows), len(decl_dedup), denom))

per_gate = table("B1 PER-GATE (refusals)", arm_dedup + dec_rows, label=lambda c: c["cls"])
per_decl = table("B1 PER-DECLINE-CLASS (fresh-MET AI-declines)", decl_dedup, label=lambda c: "decline:" + c["cond"])

print("\n== B1 arm refusals detail (first 20) ==")
for c in arm_dedup[:20]:
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    print("  %s %s %s %s R=%.2f -> %-5s $%+6.1f | %s" % (
        ctime(c["t"]), c["session"], c["sid"], c["side"], planned_r(c["side"], c["entry"], c["sl"], c["tp"]),
        out, pnl * 2, c["reason"][:60]))

print("\n== B1 declines detail (first 20) ==")
for c in decl_dedup[:20]:
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    print("  %s %-6s %s R=%.2f -> %-5s $%+6.1f | %s" % (
        ctime(c["t"]), c["sid"], c["side"], planned_r(c["side"], c["entry"], c["sl"], c["tp"]),
        out, pnl * 2, c["cond"]))

# ============================================================
# B2 — armed-era split
# ============================================================
print("\n== B2 ARMED-ERA SPLIT (mandate 08-27 09:25 CT) ==")
def agg(rows):
    d = dict(n=0, s=0.0, won=0)
    for c in rows:
        out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
        d["n"] += 1
        d["s"] += pnl * 2.0
        if out == "WON":
            d["won"] += 1
    return d

pre = [c for c in decl_dedup if c["t"] < AMms]
post = [c for c in decl_dedup if c["t"] >= AMms]
for name, rows in (("PRE-mandate declines", pre), ("POST-mandate declines", post)):
    d = agg(rows)
    print("  %-24s n=%2d won=%2d  S=%+8.1f" % (name, d["n"], d["won"], d["s"]))
d_all = agg(decl_dedup)
print("  %-24s n=%2d won=%2d  S=%+8.1f  <-- $1,763 leak gauge re-measured" % ("ALL declines", d_all["n"], d_all["won"], d_all["s"]))

arms_all = DB.execute("SELECT created_at FROM armed_orders WHERE session NOT LIKE 'TEST%'").fetchall()
arms_pre = sum(1 for (c,) in arms_all if int(datetime.fromisoformat(c).timestamp() * 1000) < AMms)
arms_post = len(arms_all) - arms_pre
print("  real arms-authored (armed_orders non-TEST): total=%d pre=%d post=%d" % (len(arms_all), arms_pre, arms_post))
print("  armed_orders rows=%d (incl. %d TEST-E2)" % (DB.execute("SELECT COUNT(*) FROM armed_orders").fetchone()[0],
        DB.execute("SELECT COUNT(*) FROM armed_orders WHERE session LIKE 'TEST%'").fetchone()[0]))
print("  declines vs arms-placed ratio: pre %s  post %.2f:1" % (
    "29:0 (no arms pre-mandate)" if arms_pre == 0 else "%.1f:1" % (len(pre) / arms_pre),
    len(post) / max(arms_post, 1)))

# ============================================================
# B3 — per-condition expectancy on pnl_corrected
# ============================================================
print("\n== B3 PER-CONDITION EXPECTANCY (pnl_corrected, closes since %s) ==" % W0.strftime("%m-%d"))
q = """
SELECT p.pnl_corrected, p.realized_pnl, p.side, j.plan_session, j.cited_scenario_id,
       j.plan_link_note, j.doc
FROM position_plan_join j JOIN trader_positions p ON p.id = j.position_id
WHERE p.status='CLOSED' AND p.exit_time >= ?
"""
rows = DB.execute(q, (W0ms,)).fetchall()
unres = DB.execute(
    "SELECT pnl_corrected, side FROM trader_positions p WHERE p.status='CLOSED' AND p.exit_time >= ? "
    "AND NOT EXISTS (SELECT 1 FROM position_plan_join j WHERE j.position_id = p.id)",
    (W0ms,)).fetchall()
cond_agg = defaultdict(lambda: dict(n=0, s=0.0, w=0))
for pnl, rpnl, side, sess, sid, note, doc in rows:
    pc = pnl if pnl is not None else rpnl
    cond = "UNRESOLVABLE"
    if sid and doc:
        try:
            d = json.loads(doc)
            sc = next((s for s in d.get("scenarios", []) if s.get("id") == sid), None)
            if sc:
                cond = sc.get("condition", "?")
                sess = sess or d.get("session")
        except Exception:
            pass
    if "UNRESOLVABLE" in (note or ""):
        cond = "UNRESOLVABLE"
    g = cond_agg[(cond, sess or "?")]
    g["n"] += 1
    g["s"] += pc
    if pc > 0:
        g["w"] += 1
for pnl, side in unres:
    g = cond_agg[("UNRESOLVABLE", "?")]
    g["n"] += 1
    g["s"] += pnl if pnl is not None else 0
    if (pnl or 0) > 0:
        g["w"] += 1
print("%-22s %-9s %3s %5s %8s %6s" % ("condition", "session", "n", "win", "S$", "win%"))
for (cond, sess) in sorted(cond_agg):
    g = cond_agg[(cond, sess)]
    print("%-22s %-9s %3d %5d %+8.1f %5.0f%%" % (cond, sess, g["n"], g["w"], g["s"], 100 * g["w"] / max(g["n"], 1)))

# ============================================================
# B4 — R-floor sweep per play
# ============================================================
print("\n== B4 R-FLOOR SWEEP PER PLAY ==")
all_c = list(decl_dedup) + list(arm_dedup) + list(dec_rows)
for c in all_c:
    c["R"] = planned_r(c["side"], c["entry"], c["sl"], c["tp"])
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    c["out"], c["pnl"] = out, pnl * 2.0
plays = sorted({c["cond"] for c in all_c})
hdr = "%-18s" % "play"
for thr in (0.0, 1.5, 2.0, 2.5, 3.0):
    hdr += " %10s" % ("all" if thr == 0 else "R>=%.1f" % thr)
print(hdr)
for play in plays:
    line = "%-18s" % play
    for thr in (0.0, 1.5, 2.0, 2.5, 3.0):
        sel = [c for c in all_c if c["cond"] == play and c["R"] >= thr - 1e-9]
        s = sum(c["pnl"] for c in sel)
        line += " n=%2d %+7.0f" % (len(sel), s)
    print(line)
sel0 = [c for c in all_c]
print("TOTAL candidates=%d (refusals %d + declines %d)" % (len(all_c), len(arm_dedup) + len(dec_rows), len(decl_dedup)))

print("\n== B1 per-scenario decline outcomes (S3-class re-test) ==")
per_sid = defaultdict(lambda: dict(n=0, w=0, l=0, a=0, s=0.0))
for c in decl_dedup:
    out, pnl = replay(c["t"], c["side"], c["entry"], c["sl"], c["tp"])
    g = per_sid[c["sid"]]
    g["n"] += 1
    if out == "WON":
        g["w"] += 1
        g["s"] += pnl * 2
    elif out == "AMBIG":
        g["a"] += 1
        g["s"] += pnl * 2
    else:
        g["l"] += 1
        g["s"] += pnl * 2
for sid in sorted(per_sid):
    g = per_sid[sid]
    print("  %s: n=%d won=%d lost=%d ambig=%d S=%+.1f" % (sid, g["n"], g["w"], g["l"], g["a"], g["s"]))
