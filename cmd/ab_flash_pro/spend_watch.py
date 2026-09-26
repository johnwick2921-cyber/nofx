#!/usr/bin/env python3
# RESEARCH-FLASH-AB spend watcher (DS-101). External enforcement: no binary
# change, no restart — reads the crash-safe artifacts the harness already
# writes, prices every completed call with the OFFICIAL DeepSeek rates the CTO
# gave (2026-09-26 03:56Z, read 09-25 [A]), peak/off-peak by the call's UTC
# start time, and SIGTERMs the harness when the cumulative spend crosses the
# stop line (set BELOW the $40 cap to leave headroom for in-flight calls).
#
# Rates (USD per 1M tokens), PEAK (01:00–04:00 and 06:00–10:00 UTC Mon–Fri),
# HALF off-peak:
#   deepseek-v4-pro : in-miss 1.32 · in-hit 0.044 · out 3.96
#   deepseek-flash  : in-miss 0.30 · in-hit 0.006 · out 1.20
# Prompt tokens are priced at the cache-MISS rate (the harness sends full
# prompts; DeepSeek's own cache discount applies automatically and is NOT
# guessed here — the miss rate is the conservative bound).
import json, os, signal, subprocess, sys, time
from datetime import datetime, timezone

CAP = 40.0
STOP_LINE = 38.0  # headroom for in-flight requests

RATES = {
    "deepseek-v4-pro": (1.32, 0.044, 3.96),
    "deepseek-flash": (0.30, 0.006, 1.20),
}

def peak(utc: datetime) -> bool:
    if utc.weekday() >= 5:
        return False  # Mon–Fri only
    h = utc.hour + utc.minute / 60.0
    return (1.0 <= h < 4.0) or (6.0 <= h < 10.0)

def price(model: str, at: datetime) -> tuple:
    p_in, p_hit, p_out = RATES.get(model, (0.0, 0.0, 0.0))
    if peak(at):
        return p_in, p_hit, p_out
    return p_in / 2, p_hit / 2, p_out / 2

def cost_of(rec: dict) -> tuple:
    wall = float(rec.get("wall_s") or 0)
    mtime = os.path.getmtime(rec["path"])
    at = datetime.fromtimestamp(mtime - wall, tz=timezone.utc)
    p_in, _, p_out = price(rec.get("model", ""), at)
    c = rec.get("prompt_tok", 0) * p_in / 1e6 + rec.get("completion_tok", 0) * p_out / 1e6
    return c, at

def main():
    root = os.path.dirname(os.path.abspath(__file__)) + "/../.."
    raw_dir = os.path.join(root, "ab_raw")
    seen = {}
    total = 0.0
    # the smoke call (row 219 arm a) is NOT in ab_raw: priced once from its CSV
    smoke = 0.226  # computed: 56,703×1.32 + 38,184×3.96 /1e6 at peak 03:43Z
    total += smoke
    seen["smoke"] = True
    log = open("/tmp/ab_spend.log", "a")
    log.write(f"[spend] boot: smoke {smoke:.3f} (peak) · total {total:.3f} · cap {CAP} · stop {STOP_LINE}\n")
    log.flush()
    while True:
        files = sorted(os.listdir(raw_dir)) if os.path.isdir(raw_dir) else []
        for fn in files:
            if fn in seen or not fn.endswith(".json"):
                continue
            seen[fn] = True
            try:
                rec = json.load(open(os.path.join(raw_dir, fn)))
            except Exception:
                continue
            rec["path"] = os.path.join(raw_dir, fn)
            c, at = cost_of(rec)
            total += c
            tag = "PEAK" if peak(at) else "offpeak"
            log.write(f"[spend] {fn}: {c:.4f} ({tag}, {at.isoformat()}) · total {total:.3f}\n")
            log.flush()
        if total >= STOP_LINE:
            log.write(f"[spend] STOP LINE {STOP_LINE} crossed at {total:.2f} — SIGTERM to the harness\n")
            log.flush()
            subprocess.run(["pkill", "-f", "ab_flash_pro"], check=False)
            print(f"STOPPED at ${total:.2f}")
            return
        # harness alive? if it finished normally, exit
        if not subprocess.run(["pgrep", "-f", "ab_flash_pro"], capture_output=True).returncode == 0:
            log.write(f"[spend] harness exited · final total {total:.3f}\n")
            log.flush()
            print(f"harness done · total ${total:.2f}")
            return
        time.sleep(60)

if __name__ == "__main__":
    main()
