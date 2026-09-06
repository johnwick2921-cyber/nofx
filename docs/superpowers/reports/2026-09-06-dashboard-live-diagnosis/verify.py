from pathlib import Path
import json,datetime,zoneinfo,math
p=Path(__file__).parent; e=json.loads((p/'evidence.json').read_text());g=e['candle_gap'];z=zoneinfo.ZoneInfo(g['timezone'])
a={r['openTime']:r for r in g['api_rows']};b={r['open_time_ms']:r for r in g['db_rows']}
start=int(datetime.datetime(2026,9,6,17,1,tzinfo=z).timestamp()*1000)
missing=[start+i*60000 for i in range(5) if start+i*60000 not in a]
assert missing==g['missing_open_time_ms']==[start+i*60000 for i in [1,2,3]]
assert missing==[start+i*60000 for i in range(5) if start+i*60000 not in b]
for t in sorted(a.keys() & b.keys()):
 for ak,bk in [('open','o'),('high','h'),('low','l'),('close','c'),('volume','v')]:assert a[t][ak]==b[t][bk],(t,ak)
for s in e['planner']['scenarios']:
 risk=abs(s['entry']-s['stop']);assert math.isclose(risk,s['risk_points']);assert math.isclose(abs(s['target']-s['entry'])/risk,s['planned_r']);assert math.isclose(abs(s['target_chain'][0]-s['entry'])/risk,s['first_target_r'])
assert e['arm']['id']==106 and e['arm']['signal_id']=='' and e['arm']['placement_seq']==0
assert e['config']['STOP_ENTRY_SEAM']=='off'
print('PASS: exact three-minute gap in API and DB; shared closed OHLCV matches; all three scenario calculations; unsent arm evidence. Does not prove live recovery or an implementation fix.')
