#!/usr/bin/env python3
"""Round 24 Q4 — cell aggregation over the harness outputs (out-r24/*.jsonl) and
the exported plans/system_config (out-r24/*.json). Prints the cells with n and
the first 5 sample ids, and writes out-r24/cells.json. Read-only."""
import json, statistics, sys, os, collections
OUT = sys.argv[1] if len(sys.argv) > 1 else 'out-r24'
THRESH = 250.0
def jl(p): return [json.loads(l) for l in open(os.path.join(OUT, p))]
def med(xs): return round(statistics.median(xs), 2) if xs else None
def ids(rows, k=5): return [f"{r['day']} {r['session']} v{r.get('version','')} {r.get('read_at','')}".strip() for r in rows[:k]]
reads, scen, seats, sess = jl('reads.jsonl'), jl('scenarios.jsonl'), jl('seats.jsonl'), jl('sessions.jsonl')
run = json.load(open(os.path.join(OUT, 'run.json')))
plans = json.load(open(os.path.join(OUT, 'copy_plans.json'))) + json.load(open(os.path.join(OUT, 'live_plans_tail.json')))
sysc = {r['key']: r['value'] for r in json.load(open(os.path.join(OUT, 'copy_sysconfig_replans.json'))) + json.load(open(os.path.join(OUT, 'live_sysconfig_replans.json')))}
cells = {}
def key(r): return (r['day'], r['session'])
complete = {key(s) for s in sess if s['session_complete']}
big = {key(s) for s in sess if s['session_complete'] and s['session_range'] >= THRESH}
pops = [('range>=250', big), ('all-complete', complete)]
print(f"basis-suspect reads (doc levels vs tape basis disagree; excluded from (a)/(b)): {[(r['day'], r['session'], r['version'], r['read_at']) for r in reads if r['basis_suspect']]}")
print(f"session-days with >=1 plan read: {len(sess)} · complete windows: {len(complete)} · range>={THRESH:g}: {len(big)}")
print("  big session-days:", sorted(big))
print("  incomplete (excluded):", sorted(key(s) for s in sess if not s['session_complete']))

# ---------- (a) ----------
print("\n== (a) seated table (doc.levels) has a level AHEAD of price in the session's net-move direction ==")
for pname, pop in pops:
    rs = [r for r in reads if key(r) in pop and r['net_dir'] in ('up', 'down') and not r['basis_suspect']]
    with_lv = [r for r in rs if r['n_levels'] > 0]
    yes = [r for r in with_lv if r['a_ahead_net']]
    no = [r for r in with_lv if not r['a_ahead_net']]
    per_day = collections.defaultdict(lambda: [0, 0])
    for r in with_lv: per_day[key(r)][0 if r['a_ahead_net'] else 1] += 1
    fr_yes = [r for r in with_lv if r['a_ahead_fromread']]
    c = {'n_reads': len(rs), 'n_reads_with_levels': len(with_lv), 'n_reads_zero_levels': len(rs) - len(with_lv),
         'yes': len(yes), 'no': len(no), 'frac_yes': round(len(yes) / len(with_lv), 3) if with_lv else None,
         'median_dist_pts_when_present': med([r['a_ahead_net_dist'] for r in yes]),
         'p25_p75_dist': (round(statistics.quantiles([r['a_ahead_net_dist'] for r in yes], n=4)[0], 1), round(statistics.quantiles([r['a_ahead_net_dist'] for r in yes], n=4)[2], 1)) if len(yes) >= 4 else None,
         'variant_fromread_frac_yes': round(len(fr_yes) / len(with_lv), 3) if with_lv else None,
         'per_session_day': {f"{d} {s}": {'yes': v[0], 'no': v[1], 'frac': round(v[0] / (v[0] + v[1]), 2)} for (d, s), v in sorted(per_day.items())},
         'first5_yes': ids(yes), 'first5_no': ids(no)}
    by_sess = {}
    for S in ('ASIA', 'LONDON', 'NY'):
        w = [r for r in with_lv if r['session'] == S]; y = [r for r in w if r['a_ahead_net']]
        by_sess[S] = {'n': len(w), 'yes': len(y), 'frac': round(len(y) / len(w), 3) if w else None, 'median_dist': med([r['a_ahead_net_dist'] for r in y])}
    c['by_session'] = by_sess
    # (a') table reach: farthest ahead level vs the session's excursion from the read price
    far = [r for r in with_lv if r['a_far_ahead_dist'] > 0]
    exh = [r for r in far if r['a_table_exhausted']]
    c['a_prime'] = {'n_reads_with_ahead_level': len(far), 'table_exhausted': len(exh), 'frac_exhausted': round(len(exh) / len(far), 3) if far else None,
                    'median_far_ahead_dist_pts': med([r['a_far_ahead_dist'] for r in far]), 'median_excursion_pts': med([r['a_excursion'] for r in far]),
                    'by_session': {S: {'n': len([r for r in far if r['session'] == S]), 'exhausted': len([r for r in exh if r['session'] == S])} for S in ('ASIA', 'LONDON', 'NY')},
                    'first5_exhausted': [f"{r['day']} {r['session']} v{r['version']} {r['net_dir']} far={r['a_far_ahead_dist']:.1f} excursion={r['a_excursion']:.1f} @{r['read_at']}" for r in exh[:5]]}
    cells[f'a:{pname}'] = c
    print(f"[{pname}] reads n={len(rs)} (with levels {len(with_lv)}, zero-level rows {len(rs)-len(with_lv)}) · ahead YES {len(yes)} / NO {len(no)} → frac {c['frac_yes']} · median dist {c['median_dist_pts_when_present']} pt · from-read variant frac {c['variant_fromread_frac_yes']}")
    for S, v in by_sess.items(): print(f"    {S}: n={v['n']} yes={v['yes']} frac={v['frac']} median_dist={v['median_dist']}")
    print("    per session-day:", {k: v['frac'] for k, v in c['per_session_day'].items()})
    print("    first5 yes:", c['first5_yes']); print("    first5 no:", c['first5_no'])
    ap = c['a_prime']
    print(f"    (a') table reach: reads with an ahead level n={ap['n_reads_with_ahead_level']} · farthest-ahead median {ap['median_far_ahead_dist_pts']} pt · excursion median {ap['median_excursion_pts']} pt · TABLE EXHAUSTED {ap['table_exhausted']} → frac {ap['frac_exhausted']} · by session {ap['by_session']}")
    print("    first5 exhausted:", ap['first5_exhausted'])

# ---------- (b) ----------
print("\n== (b) scenarios whose target_chain was EXHAUSTED (price traded beyond the last target) before the flat ==")
for pname, pop in pops:
    suspect_reads = {(r['plan_id'], r['version']) for r in reads if r['basis_suspect']}
    ss = [s for s in scen if key(s) in pop and s['n_targets'] > 0 and s['direction'] in ('long', 'short') and (s['plan_id'], s['version']) not in suspect_reads]
    ex = [s for s in ss if s['exhausted']]
    c = {'n_scenarios': len(ss), 'n_no_targets_or_dir': len([s for s in scen if key(s) in pop]) - len(ss), 'exhausted': len(ex),
         'frac': round(len(ex) / len(ss), 3) if ss else None, 'median_min_to_exhaust': med([s['min_to_exhaust'] for s in ex]),
         'non_monotonic_chains': sum(1 for s in ss if not s['monotonic']),
         'by_direction': {d: {'n': len([s for s in ss if s['direction'] == d]), 'exhausted': len([s for s in ex if s['direction'] == d])} for d in ('long', 'short')},
         'by_session': {S: {'n': len([s for s in ss if s['session'] == S]), 'exhausted': len([s for s in ex if s['session'] == S])} for S in ('ASIA', 'LONDON', 'NY')},
         'first5_exhausted': [f"{s['day']} {s['session']} v{s['version']} {s['scenario']} {s['direction']} last={s['last_target']} @{s['read_at']}" for s in ex[:5]],
         'first5_not': [f"{s['day']} {s['session']} v{s['version']} {s['scenario']} {s['direction']} last={s['last_target']} @{s['read_at']}" for s in ss if not s['exhausted']][:5]}
    cells[f'b:{pname}'] = c
    print(f"[{pname}] scenarios n={len(ss)} · exhausted {len(ex)} → frac {c['frac']} · median min-to-exhaust {c['median_min_to_exhaust']} · non-monotonic chains {c['non_monotonic_chains']}")
    print("    by dir:", c['by_direction'], "by session:", c['by_session'])
    print("    first5 exhausted:", c['first5_exhausted'])

# ---------- (c) ----------
print("\n== (c) TARGET REACH: nearest 2 BEYOND-band HTF levels in the 4h/D trend direction — existed? reached within the session? ==")
trend_counts = collections.Counter((r['h4_trend'], r['d_trend']) for r in reads if key(r) in complete)
print("  trend pairs (4h, D) over complete-session reads:", dict(trend_counts))
for pname, pop in pops:
    for variant in ('4h', 'D', 'agree'):
        for setname in ('strict', 'wide'):
            for bk in sorted({s['band_k'] for s in seats}):
                rows = [s for s in seats if key(s) in pop and s['variant'] == variant and s['set'] == setname and s['band_k'] == bk]
                dir_reads = {(s['plan_id'], s['version']) for s in rows}
                seat1 = [s for s in rows if s['seat'] == 1]; seat2 = [s for s in rows if s['seat'] == 2]
                nocand = [s for s in rows if s['seat'] == 0]
                reads_pop = [r for r in reads if key(r) in pop]
                c = {'n_reads': len(reads_pop), 'n_reads_with_direction': len(dir_reads), 'n_reads_no_candidate': len(nocand),
                     'seat1': {'n': len(seat1), 'reached': sum(s['reached'] for s in seat1), 'hit_rate': round(sum(s['reached'] for s in seat1) / len(seat1), 3) if seat1 else None,
                               'median_min_to_hit': med([s['min_to_hit'] for s in seat1 if s['reached']]), 'median_dist_pts': med([s['dist'] for s in seat1])},
                     'seat2': {'n': len(seat2), 'reached': sum(s['reached'] for s in seat2), 'hit_rate': round(sum(s['reached'] for s in seat2) / len(seat2), 3) if seat2 else None,
                               'median_min_to_hit': med([s['min_to_hit'] for s in seat2 if s['reached']]), 'median_dist_pts': med([s['dist'] for s in seat2])},
                     'first5_seat1': [f"{s['day']} {s['session']} v{s['version']} {s['dir']} {s['kind']}·{s['tf']} {s['lo']:.2f}-{s['hi']:.2f} dist={s['dist']:.1f} reached={s['reached']} min={s['min_to_hit']} @{s['read_at']}" for s in seat1[:5]]}
                cells[f'c:{pname}:{variant}:{setname}:k{bk:g}'] = c
                if bk == run['proximity_k'] or setname == 'strict':
                    s1, s2 = c['seat1'], c['seat2']
                    print(f"[{pname}] dir={variant:5s} set={setname:6s} k={bk:g}: reads {len(reads_pop)}, with direction {len(dir_reads)}, no candidate {len(nocand)} · seat1 n={s1['n']} hit {s1['reached']} ({s1['hit_rate']}) med-min {s1['median_min_to_hit']} med-dist {s1['median_dist_pts']} · seat2 n={s2['n']} hit {s2['reached']} ({s2['hit_rate']}) med-min {s2['median_min_to_hit']} med-dist {s2['median_dist_pts']}")
    print("    first5 seat1 (4h, strict, live k):", cells[f"c:{pname}:4h:strict:k{run['proximity_k']:g}"]['first5_seat1'])

# ---------- (d) ----------
print("\n== (d) replan budget: level_event re-reads and budget exhaustion per session-day (cap from day_plan; class-35: only death_replan/owner_reread SPEND) ==")
cap = run['replan_cap']
bykey = collections.defaultdict(list)
for p in plans:
    if p['session'] in ('ASIA', 'LONDON', 'NY'): bykey[(p['trade_date'], p['session'])].append(p)
for pname, pop in pops:
    days = sorted(k for k in bykey if k in pop)
    rows = []
    for k in days:
        ps = bykey[k]
        trig = collections.Counter(p['trigger_reason'] for p in ps)
        spend = trig.get('death_replan', 0) + trig.get('owner_reread', 0)
        counter = [v for kk, v in sysc.items() if kk.startswith('dayplan_replans_used:') and f":{k[0]}:{k[1]}:" in kk]
        exhausted_marker = trig.get('replans_exhausted', 0) > 0
        rows.append({'day': k[0], 'session': k[1], 'versions': len(ps), 'level_event': trig.get('level_event', 0), 'structure_mss': trig.get('structure_mss', 0),
                     'death_replan': trig.get('death_replan', 0), 'owner_reread': trig.get('owner_reread', 0), 'owner_reset': trig.get('owner_reset', 0),
                     'planner_fail_closed': trig.get('planner_fail_closed', 0), 'spends_by_trigger': spend, 'recorded_counter': counter,
                     'replans_exhausted_marker': exhausted_marker, 'no_trade_rows': sum(1 for p in ps if p['lifecycle'] == 'no_trade')})
    n_ex = sum(1 for r in rows if r['replans_exhausted_marker'])
    n_cap = sum(1 for r in rows if r['spends_by_trigger'] >= cap or any(int(v) >= cap for v in r['recorded_counter']))
    c = {'cap': cap, 'n_session_days': len(rows), 'level_event_reads_total': sum(r['level_event'] for r in rows), 'median_level_event_per_day': med([r['level_event'] for r in rows]),
         'max_level_event_per_day': max((r['level_event'] for r in rows), default=None), 'days_with_replans_exhausted_marker': n_ex, 'days_spends_at_cap': n_cap,
         'death_replans_total': sum(r['death_replan'] for r in rows), 'owner_rereads_total': sum(r['owner_reread'] for r in rows), 'rows': rows,
         'first5': [f"{r['day']} {r['session']} versions={r['versions']} level_event={r['level_event']} death={r['death_replan']} exhausted_marker={r['replans_exhausted_marker']}" for r in rows[:5]]}
    cells[f'd:{pname}'] = c
    print(f"[{pname}] session-days n={len(rows)} · level_event re-reads total {c['level_event_reads_total']} (median/day {c['median_level_event_per_day']}, max {c['max_level_event_per_day']}) · death_replan {c['death_replans_total']} · owner_reread {c['owner_rereads_total']} · days with replans_exhausted marker {n_ex} · days with spends>=cap({cap}) {n_cap}")
    for r in rows: print(f"    {r['day']} {r['session']}: versions={r['versions']} level_event={r['level_event']} mss={r['structure_mss']} death={r['death_replan']} reread={r['owner_reread']} reset={r['owner_reset']} fail_closed={r['planner_fail_closed']} counter={r['recorded_counter']} exhausted_marker={r['replans_exhausted_marker']} no_trade_rows={r['no_trade_rows']}")

cells['_run'] = run
cells['_populations'] = {'range>=250': sorted(f"{d} {s}" for d, s in big), 'all-complete': sorted(f"{d} {s}" for d, s in complete)}
json.dump(cells, open(os.path.join(OUT, 'cells.json'), 'w'), indent=1, default=str)
print("\nwrote", os.path.join(OUT, 'cells.json'))
