#!/usr/bin/env python3
"""Round 24 Q4 "TARGET REACH" — cells + manifest from the harness rows.

Usage: analysis.py <out-r24 dir> [--big 250]
Reads sessions.jsonl / reads.jsonl / scenarios.jsonl / dayplan.jsonl, writes
cells.json + cells.md (the tables the report quotes) + manifest.json.
Every cell carries n and its first 5 ids; no cell is computed without them.
"""
import hashlib, json, os, statistics, subprocess, sys
from collections import Counter, defaultdict

OUT = sys.argv[1]
BIG = float(sys.argv[sys.argv.index('--big') + 1]) if '--big' in sys.argv else 250.0

def jl(name):
    with open(os.path.join(OUT, name)) as f:
        return [json.loads(l) for l in f if l.strip()]

def sha(path):
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        for chunk in iter(lambda: f.read(1 << 20), b''):
            h.update(chunk)
    return h.hexdigest()

def frac(num, den):
    return f"{num}/{den} = {num/den:.3f}" if den else f"{num}/0 = n/a"

def med(xs):
    return f"{statistics.median(xs):.1f}" if xs else "n/a"

def ids(rows, k='id', n=5):
    return [r[k] for r in rows[:n]]

sessions = jl('sessions.jsonl'); allreads = jl('reads.jsonl'); allscens = jl('scenarios.jsonl'); dayplan = jl('dayplan.jsonl')
# (a)/(b)/(c) populations are MODEL READS only; marker rows (owner_reset, dormant:*, rearmed:*,
# planner_fail_closed, replans_exhausted, …) copy the prior doc and are counted in (d) only.
reads = [r for r in allreads if r['is_read']]
READ_IDS = {r['id'] for r in reads}
scens = [s for s in allscens if s['read_id'] in READ_IDS]
MARKERS = Counter(r['trigger'].split(':')[0] for r in allreads if not r['is_read'])
S = {s['key']: s for s in sessions}
big = {k for k, s in S.items() if s['measured'] and s['range'] >= BIG}
small = {k for k, s in S.items() if s['measured'] and s['range'] < BIG}
unmeasured = [k for k, s in S.items() if not s['measured']]

cells = {'big_threshold_pts': BIG, 'population': {}, 'a': {}, 'b': {}, 'c': {}, 'd': {}}
md = []
def H(t): md.append(f"\n### {t}\n")
def row(*cols): md.append("| " + " | ".join(str(c) for c in cols) + " |")

# ── population ──────────────────────────────────────────────────────────
pop = cells['population']
pop['session_days'] = len(sessions); pop['measured'] = len(sessions) - len(unmeasured)
pop['plan_rows'] = len(allreads); pop['model_reads'] = len(reads); pop['marker_rows'] = dict(MARKERS)
pop['unmeasured_ids'] = unmeasured
pop['big_n'] = len(big); pop['big_ids'] = sorted(big); pop['small_n'] = len(small)
bigrows = sorted((S[k] for k in big), key=lambda s: -s['range'])
pop['big_table'] = [{'key': s['key'], 'range': s['range'], 'bars_1m': s['bars_1m'], 'reads': s['reads'],
                     'open': s['open'], 'close': s['close'], 'contract': s['contract']} for s in bigrows]
pop['range_median_all'] = statistics.median([s['range'] for s in sessions if s['measured']]) if pop['measured'] else None
H(f"Population — session-days with a plan, last 30 trade dates (range = 1m high−low over the DefaultSessionRegistry window, own contract)")
md.append(f"- session-days: {len(sessions)} · measured (1m tape present): {pop['measured']} · unmeasured: {unmeasured}")
md.append(f"- plan rows: {pop['plan_rows']} = model reads {pop['model_reads']} + marker rows {pop['marker_rows']} (markers excluded from a/b/c, counted in d)")
md.append(f"- range ≥ {BIG:.0f} pt: **{len(big)}** session-days · < {BIG:.0f}: {len(small)} · median range (measured): {pop['range_median_all']:.1f}")
row('session-day', 'range', '1m bars', 'reads', 'open→close'); row('---', '---', '---', '---', '---')
for s in bigrows:
    row(s['key'], f"{s['range']:.2f}", s['bars_1m'], s['reads'], f"{s['open']:.2f}→{s['close']:.2f}")

# ── (a) level ahead of price in the eventual-move direction ─────────────
def a_cells(pool, label):
    rs = [r for r in reads if r['sess_key'] in pool and r['measured']]
    rec = [r for r in rs if r['recorded'] and r.get('seated')]
    out = {'label': label, 'reads_measured': len(rs), 'reads_recorded': len(rec), 'read_ids': ids(rs)}
    # seated (recorded input table) — the 12-seat table the AI was shown
    if rec:
        yes = [r for r in rec if r['seated']['ahead_in_band_n'] > 0]
        ranout = [r for r in rec if r['seated']['ran_out']]
        out['seated'] = {
            'n': len(rec), 'ahead_in_band_yes': len(yes), 'frac': len(yes)/len(rec),
            'yes_ids': ids(yes), 'no_ids': ids([r for r in rec if r not in yes]),
            'ran_out_n': len(ranout), 'ran_out_frac': len(ranout)/len(rec),
            'overshoot_median_pts': statistics.median([r['seated']['overshoot'] for r in ranout]) if ranout else None,
            'ahead_median_n': statistics.median([r['seated']['ahead_n'] for r in rec]),
            'ahead_htf_any': sum(1 for r in rec if r['seated']['ahead_htf_n'] > 0),
            'seated_n_median': statistics.median([r['seated']['n'] for r in rec]),
            'band_probe_max_abs_dist_over_band': sum(1 for r in rec if r['seated_max_abs_dist'] > r['band'] + 1e-6),
        }
        # per session-day fractions
        per = defaultdict(lambda: [0, 0])
        for r in rec:
            per[r['sess_key']][1] += 1
            if r['seated']['ahead_in_band_n'] > 0: per[r['sess_key']][0] += 1
        out['seated']['per_session_day'] = {k: f"{v[0]}/{v[1]}" for k, v in sorted(per.items())}
    # doc (published levels[]) — all reads
    yes = [r for r in rs if r['doc']['ahead_in_band_n'] > 0]
    ranout = [r for r in rs if r['doc']['ran_out']]
    out['doc'] = {
        'n': len(rs), 'ahead_in_band_yes': len(yes), 'frac': len(yes)/len(rs) if rs else None,
        'yes_ids': ids(yes), 'no_ids': ids([r for r in rs if r not in yes]),
        'ran_out_n': len(ranout), 'ran_out_frac': len(ranout)/len(rs) if rs else None,
        'overshoot_median_pts': statistics.median([r['doc']['overshoot'] for r in ranout]) if ranout else None,
        'doc_n_median': statistics.median([r['doc']['n'] for r in rs]) if rs else None,
    }
    per = defaultdict(lambda: [0, 0])
    for r in rs:
        per[r['sess_key']][1] += 1
        if r['doc']['ahead_in_band_n'] > 0: per[r['sess_key']][0] += 1
    out['doc']['per_session_day'] = {k: f"{v[0]}/{v[1]}" for k, v in sorted(per.items())}
    # split by trigger: the scheduled read vs level_event re-reads (doc table, all reads)
    bytrig = {}
    for trig in ('scheduled', 'level_event', 'other'):
        sel = [r for r in rs if (r['trigger'].endswith('_scheduled_read') and trig == 'scheduled') or (r['trigger'] == 'level_event' and trig == 'level_event') or (trig == 'other' and not r['trigger'].endswith('_scheduled_read') and r['trigger'] != 'level_event')]
        y = sum(1 for r in sel if r['doc']['ahead_in_band_n'] > 0)
        ro = sum(1 for r in sel if r['doc']['ran_out'])
        cell = {'n': len(sel), 'doc_ahead_yes': y, 'doc_ran_out': ro, 'ids': ids(sel)}
        if rec:
            selr = [r for r in sel if r['recorded'] and r.get('seated')]
            cell['seated_n'] = len(selr); cell['seated_ahead_yes'] = sum(1 for r in selr if r['seated']['ahead_in_band_n'] > 0); cell['seated_ran_out'] = sum(1 for r in selr if r['seated']['ran_out'])
        bytrig[trig] = cell
    out['by_trigger'] = bytrig
    # direction agreement probe: dominant-excursion vs close-to-close
    agree = sum(1 for r in rs if r['dir_exc'] == r['dir_close'])
    out['dir_probe_exc_eq_close'] = f"{agree}/{len(rs)}"
    out['excursion_median_pts'] = statistics.median([r['excursion'] for r in rs]) if rs else None
    return out

cells['a']['big'] = a_cells(big, f'range ≥ {BIG:.0f}')
cells['a']['small'] = a_cells(small, f'range < {BIG:.0f} (control)')
for key in ('big', 'small'):
    c = cells['a'][key]
    H(f"(a) A level AHEAD of price (eventual-move side, within ±k×DATR) — {c['label']}")
    md.append(f"- reads measured: {c['reads_measured']} (recorded input snapshots: {c['reads_recorded']}); median excursion after read: {c['excursion_median_pts']:.1f} pt; dir(excursion)==dir(close): {c['dir_probe_exc_eq_close']}")
    row('table', 'n reads', 'ahead-in-band YES', 'fraction', 'ran out (excursion > farthest ahead)', 'median overshoot', 'median table size'); row('---','---','---','---','---','---','---')
    if 'seated' in c:
        s = c['seated']
        row('seated 12-seat input (recorded)', s['n'], s['ahead_in_band_yes'], f"{s['frac']:.3f}", f"{s['ran_out_n']} ({s['ran_out_frac']:.3f})", f"{s['overshoot_median_pts']:.1f}" if s['overshoot_median_pts'] is not None else 'n/a', s['seated_n_median'])
    d = c['doc']
    row('published doc.levels (all reads)', d['n'], d['ahead_in_band_yes'], f"{d['frac']:.3f}" if d['frac'] is not None else 'n/a', f"{d['ran_out_n']} ({d['ran_out_frac']:.3f})" if d['ran_out_frac'] is not None else 'n/a', f"{d['overshoot_median_pts']:.1f}" if d['overshoot_median_pts'] is not None else 'n/a', d['doc_n_median'])
    if 'seated' in c:
        md.append(f"- seated per session-day (yes/reads): {c['seated']['per_session_day']}")
        md.append(f"- seated: reads with ≥1 HTF (1h/4h/D) level ahead: {c['seated']['ahead_htf_any']}/{c['seated']['n']}; band probe (seated |distance| > k×DATR): {c['seated']['band_probe_max_abs_dist_over_band']} reads")
    md.append(f"- doc per session-day (yes/reads): {d['per_session_day']}")
    md.append("- by trigger (n · doc ahead-yes · doc ran-out · seated n · seated ahead-yes · seated ran-out): " + "; ".join(f"{k}: {v['n']} · {v['doc_ahead_yes']} · {v['doc_ran_out']} · {v.get('seated_n','-')} · {v.get('seated_ahead_yes','-')} · {v.get('seated_ran_out','-')}" for k, v in c['by_trigger'].items()))

# ── (b) scenario target_chain exhausted before flat ─────────────────────
def b_cells(pool, label):
    ss = [s for s in scens if s['sess_key'] in pool]
    m = [s for s in ss if s['measured']]
    ex = [s for s in m if s['exhausted']]
    out = {'label': label, 'scenarios': len(ss), 'measured': len(m), 'malformed': Counter(s.get('malformed','') for s in ss if s.get('malformed','')),
           'exhausted_n': len(ex), 'exhausted_frac': len(ex)/len(m) if m else None,
           'hit_first_n': sum(1 for s in m if s['hit_first']), 'ids_exhausted': [f"{s['read_id']}#{s['scen_id']}" for s in ex[:5]],
           'ids_measured': [f"{s['read_id']}#{s['scen_id']}" for s in m[:5]],
           'minutes_to_last_median': statistics.median([s['minutes_to_last'] for s in ex]) if ex else None,
           'last_dist_median_pts': statistics.median([s['last_dist'] for s in m]) if m else None,
           'n_targets_median': statistics.median([s['n_targets'] for s in m]) if m else None}
    per = defaultdict(lambda: [0, 0])
    for s in m:
        per[s['sess_key']][1] += 1
        if s['exhausted']: per[s['sess_key']][0] += 1
    out['per_session_day'] = {k: f"{v[0]}/{v[1]}" for k, v in sorted(per.items())}
    return out
cells['b']['big'] = b_cells(big, f'range ≥ {BIG:.0f}'); cells['b']['small'] = b_cells(small, 'control')
H("(b) Scenarios whose target_chain was EXHAUSTED (price beyond the last target) before the session flat")
row('pool', 'scenarios', 'measured', 'malformed', 'exhausted', 'fraction', 'first target hit', 'median min→last', 'median last-target dist', 'median #targets'); row('---','---','---','---','---','---','---','---','---','---')
for key in ('big', 'small'):
    c = cells['b'][key]
    row(c['label'], c['scenarios'], c['measured'], dict(c['malformed']), c['exhausted_n'], f"{c['exhausted_frac']:.3f}" if c['exhausted_frac'] is not None else 'n/a', c['hit_first_n'], f"{c['minutes_to_last_median']:.0f}" if c['minutes_to_last_median'] is not None else 'n/a', f"{c['last_dist_median_pts']:.1f}" if c['last_dist_median_pts'] is not None else 'n/a', c['n_targets_median'])
md.append(f"- big per session-day (exhausted/measured): {cells['b']['big']['per_session_day']}")

# ── (c) candidate rule: nearest beyond-band HTF level in the trend direction ──
def c_cells(pool, label):
    rs = [r for r in reads if r['sess_key'] in pool and r['measured']]
    out = {'label': label, 'reads': len(rs), 'defs': {}}
    for d in ('s1_1h', 's1_4h', 's1_D', 'plan_bias'):
        cs = [(r, c) for r in rs for c in r['cands'] if c['trend_def'] == d]
        trended = [(r, c) for r, c in cs if c['trend'] in ('up', 'down')]
        found = [(r, c) for r, c in trended if c['found']]
        hit = [(r, c) for r, c in found if c['hit']]
        out['defs'][d] = {
            'reads': len(cs), 'trend_counts': Counter(c['trend'] for _, c in cs),
            'trended': len(trended), 'found': len(found), 'found_frac': len(found)/len(trended) if trended else None,
            'hit': len(hit), 'hit_frac': len(hit)/len(found) if found else None,
            'minutes_median': statistics.median([c.get('minutes',0) for _, c in hit]) if hit else None,
            'dist_median': statistics.median([c.get('dist',0) for _, c in found]) if found else None,
            'tf_of_found': Counter(c.get('tf','') for _, c in found), 'role_of_found': Counter(c.get('role') or '(unassigned)' for _, c in found),
            'ids_found': [r['id'] for r, _ in found[:5]], 'ids_hit': [r['id'] for r, _ in hit[:5]],
            'ids_none': [r['id'] for r, c in trended if not c['found']][:5],
            'ring_median': statistics.median([c.get('ring_n',0) for _, c in cs]) if cs else None,
        }
    out['htf_src'] = Counter(r['htf_src'] for r in rs)
    out['htf_beyond_band_median'] = statistics.median([r['htf_beyond_band_n'] for r in rs]) if rs else None
    return out
cells['c']['big'] = c_cells(big, f'range ≥ {BIG:.0f}'); cells['c']['small'] = c_cells(small, 'control')
for key in ('big', 'small'):
    c = cells['c'][key]
    H(f"(c) Candidate rule — nearest BEYOND-band HTF (1h/4h/D) level in the trend direction — {c['label']}")
    md.append(f"- reads: {c['reads']}; HTF universe source: {dict(c['htf_src'])}; median beyond-band HTF levels per read: {c['htf_beyond_band_median']}")
    row('trend def', 'trend counts', 'trended reads', 'level existed', 'fraction', 'reached before flat', 'hit rate', 'median min→hit', 'median dist (pt)', 'tf of level', 'role (recorded only)'); row('---','---','---','---','---','---','---','---','---','---','---')
    for d, v in c['defs'].items():
        row(d, dict(v['trend_counts']), v['trended'], v['found'], f"{v['found_frac']:.3f}" if v['found_frac'] is not None else 'n/a', v['hit'], f"{v['hit_frac']:.3f}" if v['hit_frac'] is not None else 'n/a', f"{v['minutes_median']:.0f}" if v['minutes_median'] is not None else 'n/a', f"{v['dist_median']:.1f}" if v['dist_median'] is not None else 'n/a', dict(v['tf_of_found']), dict(v['role_of_found']))

# ── (c-inband) alternative rule: FARTHEST in-band HTF level in the trend direction ──
def cin_cells(pool, label):
    rs = [r for r in reads if r['sess_key'] in pool and r['measured']]
    out = {'label': label, 'reads': len(rs), 'defs': {}}
    for d in ('s1_1h', 's1_4h', 's1_D', 'plan_bias'):
        cs = [(r, c) for r in rs for c in r.get('cands_inband', []) if c['trend_def'] == d]
        trended = [(r, c) for r, c in cs if c['trend'] in ('up', 'down')]
        found = [(r, c) for r, c in trended if c['found']]
        hit = [(r, c) for r, c in found if c['hit']]
        beyond_table = [(r, c) for r, c in found if r['doc'] and c.get('dist', 0) > r['doc']['farthest_ahead'] and c['trend'] == r['dir_exc']]
        out['defs'][d] = {'trended': len(trended), 'found': len(found), 'found_frac': len(found)/len(trended) if trended else None,
                          'hit': len(hit), 'hit_frac': len(hit)/len(found) if found else None,
                          'minutes_median': statistics.median([c.get('minutes', 0) for _, c in hit]) if hit else None,
                          'dist_median': statistics.median([c.get('dist', 0) for _, c in found]) if found else None,
                          'tf_of_found': Counter(c.get('tf', '') for _, c in found),
                          'farther_than_table_and_aligned': len(beyond_table), 'farther_hit': sum(1 for _, c in beyond_table if c['hit']),
                          'ids_found': [r['id'] for r, _ in found[:5]], 'ids_hit': [r['id'] for r, _ in hit[:5]]}
    om = [r for r in rs if r.get('oracle_missed_inband_htf_n', 0) > 0]
    rec = [r for r in rs if r['recorded'] and r.get('seated')]
    out['oracle_seated'] = {'recorded_reads': len(rec), 'ran_out': sum(1 for r in rec if r['seated']['ran_out']),
                            'reads_with_missed': sum(1 for r in rec if r.get('oracle_seated_missed_n', 0) > 0),
                            'reads_where_a_missed_level_was_reached': sum(1 for r in rec if r.get('oracle_seated_missed_hit_n', 0) > 0),
                            'ids': [r['id'] for r in rec if r.get('oracle_seated_missed_hit_n', 0) > 0][:5]}
    out['oracle'] = {'reads_with_missed_inband_htf': len(om), 'frac': len(om)/len(rs) if rs else None,
                     'missed_total': sum(r['oracle_missed_inband_htf_n'] for r in rs), 'missed_hit_total': sum(r['oracle_missed_inband_htf_hit_n'] for r in rs),
                     'reads_where_a_missed_level_was_reached': sum(1 for r in rs if r['oracle_missed_inband_htf_hit_n'] > 0),
                     'ids': [f"{r['id']} {r.get('oracle_missed_ids')}" for r in om[:5]]}
    return out
cells['c_inband'] = {'big': cin_cells(big, f'range ≥ {BIG:.0f}'), 'small': cin_cells(small, 'control')}
for key in ('big', 'small'):
    c = cells['c_inband'][key]
    H(f"(c′) Alternative rule — FARTHEST in-band HTF (1h/4h/D) level in the trend direction — {c['label']}")
    row('trend def', 'trended reads', 'level existed', 'fraction', 'reached', 'hit rate', 'median min→hit', 'median dist', 'tf', 'farther than doc table & trend==move', '…of which reached'); row('---','---','---','---','---','---','---','---','---','---','---')
    for d, v in c['defs'].items():
        row(d, v['trended'], v['found'], f"{v['found_frac']:.3f}" if v['found_frac'] is not None else 'n/a', v['hit'], f"{v['hit_frac']:.3f}" if v['hit_frac'] is not None else 'n/a', f"{v['minutes_median']:.0f}" if v['minutes_median'] is not None else 'n/a', f"{v['dist_median']:.1f}" if v['dist_median'] is not None else 'n/a', dict(v['tf_of_found']), v['farther_than_table_and_aligned'], v['farther_hit'])
    os_ = c['oracle_seated']
    md.append(f"- ORACLE vs the recorded 12-SEAT table (recorded reads only): {os_['recorded_reads']} reads, table ran out on {os_['ran_out']}; reads with ≥1 in-band HTF level on the move side beyond the seated farthest: {os_['reads_with_missed']}; reads where ≥1 such level was REACHED: {os_['reads_where_a_missed_level_was_reached']} {os_['ids']}")
    o = c['oracle']
    md.append(f"- ORACLE (move direction known — lookahead, for sizing the miss only): reads with ≥1 in-band HTF level on the move side beyond the doc table's farthest: {o['reads_with_missed_inband_htf']}/{c['reads']} ({o['frac']:.3f}); such levels total {o['missed_total']}, reached {o['missed_hit_total']}; reads where ≥1 was reached: {o['reads_where_a_missed_level_was_reached']}; e.g. {o['ids'][:3]}")

# ── (c2) the crux: reads whose table RAN OUT — did the candidate rule have a level, and was it reached? ──
def c2_cells(pool, label, table):
    rs = [r for r in reads if r['sess_key'] in pool and r['measured'] and r.get(table) and r[table]['ran_out']]
    out = {'label': label, 'table': table, 'ran_out_reads': len(rs), 'ids': ids(rs), 'defs': {}}
    for d in ('s1_1h', 's1_4h', 's1_D', 'plan_bias'):
        cs = [(r, c) for r in rs for c in r['cands'] if c['trend_def'] == d]
        # the rule only helps when the trend points the way price actually went
        aligned = [(r, c) for r, c in cs if c['trend'] == r['dir_exc']]
        found = [(r, c) for r, c in aligned if c['found']]
        hit = [(r, c) for r, c in found if c['hit']]
        out['defs'][d] = {'trend_aligned_with_move': len(aligned), 'found': len(found), 'hit': len(hit),
                          'hit_frac_of_ran_out': len(hit)/len(rs) if rs else None,
                          'minutes_median': statistics.median([c.get('minutes',0) for _, c in hit]) if hit else None,
                          'ids_hit': [r['id'] for r, _ in hit[:5]]}
    return out
cells['c2'] = {'big_doc': c2_cells(big, f'range ≥ {BIG:.0f}', 'doc'), 'big_seated': c2_cells(big, f'range ≥ {BIG:.0f}', 'seated'),
               'small_doc': c2_cells(small, 'control', 'doc')}
H("(c2) The crux — reads whose level table RAN OUT (excursion beyond the farthest ahead level): would the candidate rule's level have been there and been reached?")
row('pool · table', 'ran-out reads', 'trend def', 'trend aligned with move', 'level existed', 'reached', 'reached / ran-out reads', 'median min→hit'); row('---','---','---','---','---','---','---','---')
for k, c in cells['c2'].items():
    for d, v in c['defs'].items():
        row(f"{c['label']} · {c['table']}", c['ran_out_reads'], d, v['trend_aligned_with_move'], v['found'], v['hit'], f"{v['hit_frac_of_ran_out']:.3f}" if v['hit_frac_of_ran_out'] is not None else 'n/a', f"{v['minutes_median']:.0f}" if v['minutes_median'] is not None else 'n/a')

# ── (d) replan budget ────────────────────────────────────────────────────
CLASS35_DATE = '2026-09-01'  # recorded replan counter deployed ec6632f9 2026-09-01 17:24 CT; before it, exhaustion shows as a replans_exhausted marker row
def d_cells(pool, label):
    ds = [d for d in dayplan if d['key'] in pool]
    pre = [d for d in ds if d['trade_date'] < CLASS35_DATE]; post = [d for d in ds if d['trade_date'] >= CLASS35_DATE]
    return {'label': label, 'session_days': len(ds),
            'era_pre_class35': {'session_days': len(pre), 'exhausted_marker_n': sum(1 for d in pre if d['triggers'].get('replans_exhausted', 0) > 0),
                                'exhausted_marker_ids': [d['key'] for d in pre if d['triggers'].get('replans_exhausted', 0) > 0],
                                'level_event_total': sum(d['level_event'] for d in pre)},
            'era_post_class35': {'session_days': len(post), 'counter_exhausted_n': sum(1 for d in post if d['exhausted']),
                                 'counter_used_total': sum(d['counter_used'] for d in post), 'level_event_total': sum(d['level_event'] for d in post),
                                 'marker_n': sum(1 for d in post if d['triggers'].get('replans_exhausted', 0) > 0)},
            'level_event_total': sum(d['level_event'] for d in ds),
            'level_event_median': statistics.median([d['level_event'] for d in ds]) if ds else None,
            'level_event_max': max((d['level_event'] for d in ds), default=0),
            'spending_total': sum(d['spending_reads'] for d in ds),
            'counter_used_total': sum(d['counter_used'] for d in ds),
            'exhausted_n': sum(1 for d in ds if d['exhausted']), 'exhausted_ids': [d['key'] for d in ds if d['exhausted']],
            'would_exhaust_if_counted_n': sum(1 for d in ds if d['would_exhaust_if_level_events_counted']),
            'would_exhaust_ids': [d['key'] for d in ds if d['would_exhaust_if_level_events_counted']][:5],
            'cap': ds[0]['cap'] if ds else None,
            'per_session_day': {d['key']: {'reads': d['reads'], 'level_event': d['level_event'], 'spending': d['spending_reads'], 'counter': d['counter_used'], 'triggers': d['triggers']} for d in ds}}
cells['d']['big'] = d_cells(big, f'range ≥ {BIG:.0f}'); cells['d']['small'] = d_cells(small, 'control')
H("(d) Replan budget on those days (class 35: only death_replan/owner_reread SPEND; level_event is FREE)")
row('pool', 'session-days', 'level_event re-reads (total/median/max)', 'spending reads', 'recorded counter used (sum)', 'cap', 'exhausted (counter ≥ cap)', 'would exhaust if level_events counted'); row('---','---','---','---','---','---','---','---')
for key in ('big', 'small'):
    c = cells['d'][key]
    row(c['label'], c['session_days'], f"{c['level_event_total']}/{c['level_event_median']}/{c['level_event_max']}", c['spending_total'], c['counter_used_total'], c['cap'], f"{c['exhausted_n']} {c['exhausted_ids']}", f"{c['would_exhaust_if_counted_n']} {c['would_exhaust_ids']}")
for key in ('big', 'small'):
    c = cells['d'][key]
    md.append(f"- {c['label']} — era split: pre-class-35 (< {CLASS35_DATE}): {c['era_pre_class35']['session_days']} session-days, exhausted-marker rows on {c['era_pre_class35']['exhausted_marker_n']} {c['era_pre_class35']['exhausted_marker_ids']}, level_event re-reads {c['era_pre_class35']['level_event_total']}; post (recorded counter): {c['era_post_class35']['session_days']} session-days, counter-exhausted {c['era_post_class35']['counter_exhausted_n']}, counter used total {c['era_post_class35']['counter_used_total']}, level_event re-reads {c['era_post_class35']['level_event_total']}, exhausted markers {c['era_post_class35']['marker_n']}")
md.append("- big per session-day: " + json.dumps({k: {kk: vv for kk, vv in v.items() if kk != 'triggers'} for k, v in cells['d']['big']['per_session_day'].items()}))

# ── manifest ────────────────────────────────────────────────────────────
here = os.path.dirname(os.path.abspath(__file__))
commit = subprocess.run(['git', 'rev-parse', 'HEAD'], capture_output=True, text=True, cwd=here).stdout.strip()
inputs = {}
sums = os.path.join(OUT, 'in', 'SHA256SUMS')
if os.path.exists(sums):
    for line in open(sums):
        h, _, name = line.strip().partition('  ')
        inputs[name] = h
manifest = {
    'commit': commit,
    'scripts': {'extract': 'docs/superpowers/research/2026-09-17-round-24/extract.sh',
                'harness': 'docs/superpowers/research/2026-09-17-round-24/harness/ (go run ./… -in out-r24/in -out out-r24)',
                'analysis': 'docs/superpowers/research/2026-09-17-round-24/analysis.py'},
    'inputs_sha256': inputs,
    'outputs_sha256': {n: sha(os.path.join(OUT, n)) for n in ('sessions.jsonl', 'reads.jsonl', 'scenarios.jsonl', 'dayplan.jsonl')},
    'lookahead': {
        'at_read': 'price/DATR/seated table/HTF universe/trend use only bars closed before created_at, or the planner\'s own recorded input snapshot for that read (research store, object=plan/input)',
        'outcome': 'excursion, ran_out, hit, minutes, exhausted use only 1m bars with open >= created_at and < session flat',
        'direction': 'dir_exc = side of the larger excursion from price-at-read over the remaining session (primary); dir_close = sign(flat close − price) reported as a probe',
    },
    'cells': {
        'population': {'n': len(sessions), 'big_n': len(big), 'ids': sorted(big)[:5]},
        'a_big_seated': {'n': cells['a']['big'].get('seated', {}).get('n', 0), 'ids': cells['a']['big'].get('seated', {}).get('yes_ids', [])},
        'a_big_doc': {'n': cells['a']['big']['doc']['n'], 'ids': cells['a']['big']['doc']['yes_ids']},
        'b_big': {'n': cells['b']['big']['measured'], 'ids': cells['b']['big']['ids_measured']},
        'c_big': {d: {'n': v['found'], 'ids': v['ids_found']} for d, v in cells['c']['big']['defs'].items()},
        'c2_big_doc': {'n': cells['c2']['big_doc']['ran_out_reads'], 'ids': cells['c2']['big_doc']['ids']},
        'd_big': {'n': cells['d']['big']['session_days'], 'ids': sorted(big)[:5]},
    },
}
json.dump(cells, open(os.path.join(OUT, 'cells.json'), 'w'), indent=1, default=lambda o: dict(o) if isinstance(o, Counter) else str(o))
json.dump(manifest, open(os.path.join(OUT, 'manifest.json'), 'w'), indent=1)
open(os.path.join(OUT, 'cells.md'), 'w').write("\n".join(md) + "\n")
print("\n".join(md))
print(f"\nwrote cells.json cells.md manifest.json (commit {commit[:8]})")
