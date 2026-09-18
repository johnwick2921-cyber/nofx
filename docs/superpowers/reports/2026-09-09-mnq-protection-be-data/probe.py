import sqlite3,json,pathlib,datetime,zoneinfo,collections,re
P=pathlib.Path(__file__).parent;z=zoneinfo.ZoneInfo('America/Chicago');ct=lambda x:datetime.datetime.fromtimestamp(x/1000,z).isoformat();since=int(datetime.datetime(2026,9,8,tzinfo=z).timestamp()*1000)
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro',uri=True);c.row_factory=sqlite3.Row;c.execute('BEGIN')
pos=[dict(x) for x in c.execute('SELECT id,side,entry_price,entry_order_id,entry_time,exit_time,status,close_reason,source FROM trader_positions WHERE entry_time>=? ORDER BY id',(since,))]
acc=[dict(x) for x in c.execute('SELECT id,signal_id,accepted_entry_px,accepted_stop_px,ledger_stop_px,accepted_at_ms FROM accepted_risk WHERE accepted_at_ms>=? ORDER BY id',(since,))]
snaps=[dict(x) for x in c.execute('SELECT id,emitted_at_ms,received_at_ms,orders_json FROM nt8_order_snapshots WHERE received_at_ms>=? ORDER BY id',(since,))]
changes=collections.defaultdict(list);counts=collections.Counter();prev={}
for ss in snaps:
 for o in json.loads(ss['orders_json']):
  name=o.get('name','')
  if not name.endswith('-sl'):continue
  sid=name[:-3];counts[sid]+=1
  key=(o.get('order_id'),o.get('stop_price'),o.get('state'))
  if prev.get(sid)!=key:
   changes[sid].append({'snapshot_id':ss['id'],'received_ct':ct(ss['received_at_ms']),'emitted_ct':ct(ss['emitted_at_ms']),'name':name,'order_id':o.get('order_id'),'state':o.get('state'),'stop_price':o.get('stop_price'),'quantity':o.get('quantity')});prev[sid]=key
log=[dict(x) for x in c.execute("SELECT id,ts_utc,message FROM log_events WHERE ts_utc>=? AND (message LIKE '%breakeven%' OR message LIKE '%trailing_moved%' OR message LIKE '%modify_bracket%' OR message LIKE '%move_stop%') ORDER BY id",(since,))]
for x in log:x['ct']=ct(x['ts_utc']);x['message']=re.sub(r'\[trader_id=.*?\] ','',x['message'])
for x in pos:
 x['entry_ct']=ct(x['entry_time']);x['exit_ct']=ct(x['exit_time']) if x['exit_time'] else None
out={'snapshot_ct':datetime.datetime.now(z).isoformat(),'positions':pos,'accepted_risk':acc,'stop_child_changes':dict(changes),'stop_child_snapshot_counts':dict(counts),'log_events':log,'snapshots_n':len(snaps),'snapshots_ids':[x['id'] for x in snaps]}
(P/'store.json').write_text(json.dumps(out,indent=2)+'\n')
print('snapshots',len(snaps));print('child chains',json.dumps(dict(changes),indent=2));print('logs',json.dumps(log,indent=2))
