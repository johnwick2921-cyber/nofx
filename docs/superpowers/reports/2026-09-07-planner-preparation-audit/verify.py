"""Recompute report evidence offline. No database, network, or production imports."""
import csv,json,math,pathlib,collections,datetime
ROOT=pathlib.Path(__file__).resolve().parent
load=lambda n:json.loads((ROOT/n).read_text())
h=load('planner-history.json');m=load('metrics.json');plans=h['plans']
sc=[];arms=[]
for p in plans:
 for s in p['document'].get('scenarios',[]):
  sc.append((p['reference'],s));a=s.get('arm') or {};e,sl,tp=[a.get(k) for k in ['entry','stop','target']]
  if not all(isinstance(v,(int,float)) for v in [e,sl,tp]) or e==sl:continue
  side=1 if s['direction']=='long' else -1;r=abs(e-sl);ts=s.get('target_chain') or []
  arms.append({'ref':p['reference'],'s':s['id'],'stop_ok':side*(e-sl)>0,'target_ok':side*(tp-e)>0,'first_R':side*(ts[0]-e)/r if ts else None,'target_R':side*(tp-e)/r,'off_tick':any(abs(v/.25-round(v/.25))>1e-7 for v in [e,sl,tp]),'chain_match':any(abs(tp-t)<.126 for t in ts)})
assert len(plans)==m['plan_versions']==102
assert len(sc)==m['scenario_documents']==280
assert len(arms)==m['arms_with_complete_geometry']==111
assert sum(not a['stop_ok'] for a in arms)==m['wrong_stop_side']==0
assert sum(not a['target_ok'] for a in arms)==m['wrong_target_side']==0
k=sum(a['first_R'] is not None and 0<a['first_R']<1 for a in arms);n=len(arms)
assert k==m['first_target_under_1R']['count']==45
q=1.959963984540054;p=k/n;den=1+q*q/n;c=(p+q*q/(2*n))/den;d=q*math.sqrt(p*(1-p)/n+q*q/(4*n*n))/den
assert [round(c-d,6),round(c+d,6)]==m['first_target_under_1R']['wilson95_nominal']
assert sum(not a['chain_match'] for a in arms)==m['target_not_chain_half_tick']==4
assert sum(a['off_tick'] for a in arms)==m['off_tick_any_trade_price']==54
with (ROOT/'all-scenarios.csv').open() as f:rows=list(csv.DictReader(f))
assert len(rows)==len(sc) and {(r['plan'],r['scenario']) for r in rows}=={(p,s['id']) for p,s in sc}
with (ROOT/'candidate-selection.csv').open() as f:cp=list(csv.DictReader(f))
assert len(cp)==m['candidate_rows']==600
assert sum(int(r['seated']) for r in cp)==m['candidate_seated']==299
assert all(r['score_components']=='{}' for r in cp)
t=load('touch-evidence.json');valid=[r for r in t if r['validity']=='valid']
assert dict(collections.Counter(r['validity'] for r in t))==m['touch_validity']
assert len(valid)==124 and sum(not r['candidate_seated'] for r in valid)==0
assert collections.Counter(r['outcome'] for r in valid)=={'hold':46,'break':41,'ambiguous_horizon':37}
assert len(h['read_facts'])==42 and all(not r['plan_id'] and not r['version'] for r in h['read_facts'])
bars=load('case-bars.json')['bars']
for row in load('publication-context.json'):
 end=datetime.datetime.fromisoformat(row['created_ct']).timestamp()*1000
 before=[b for b in bars if b['open_time_ms']+60000<=end]
 assert before and before[-1]['c']==row['close']
probe=load('confirmation-probe.json')
assert len(probe)==4 and all(r['expected_met'] is False and r['actual_met'] is True for r in probe)
print('PASS: plan/scenario coverage, geometry, denominators, nominal Wilson interval, candidate/touch validity, publication context, and four recorded counterexamples.')
print('This verifies report arithmetic and recorded output. To rerun production confirmation functions, copy confirmation-probe.go.txt to a scratch .go file and run it from a NOFX checkout pinned to the documented source revision.')
