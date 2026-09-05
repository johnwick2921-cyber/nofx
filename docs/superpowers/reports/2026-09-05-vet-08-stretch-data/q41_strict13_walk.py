import sqlite3, json, datetime
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro',uri=True)
def ms(s): return (int(datetime.datetime.strptime(s,'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp())+5*3600)*1000
bars=c.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' AND open_time_ms>=? AND open_time_ms<? ORDER BY open_time_ms",(ms('2026-09-03 20:30:00'),ms('2026-09-04 08:30:00'))).fetchall()
def ct(m): return datetime.datetime.utcfromtimestamp(m/1000-5*3600).strftime('%m-%d %H:%M')
flat02=ms('2026-09-04 02:00:00')
rows=c.execute("SELECT id,timestamp,decision_json FROM decision_records WHERE id BETWEEN 37304 AND 37322 ORDER BY id").fetchall()
tot=0; res=[]
for rid,ts,dj in rows:
    try: acts=json.loads(dj)
    except: continue
    for a in acts:
        if not str(a.get('action','')).startswith('open'): continue
        sl=a['stop_loss']; tp=a['take_profit']; e=(2*sl+tp)/3
        t0=(int(datetime.datetime.strptime(ts[:19],'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp()))*1000
        out=None
        for m,o,h,l,cl in bars:
            if m<t0: continue
            if m>=flat02:
                out=('flat02', e-cl, m); break
            s = h>=sl; t = l<=tp
            if s and t: out=('stop(same bar)', -(sl-e), m); break
            if s: out=('stop', -(sl-e), m); break
            if t: out=('target', e-tp, m); break
        if out is None: out=('open',0,bars[-1][0])
        res.append((rid,e,sl,tp,out))
        print("%d E=%.2f SL=%.2f TP=%.2f -> %-14s %+7.2f at %s"%(rid,e,sl,tp,out[0],out[1],ct(out[2])))
print("\nsum pts if all 13 taken: %+.2f  mean %+.2f"%(sum(r[4][1] for r in res), sum(r[4][1] for r in res)/len(res)))
# ASIA session extremes 17:00 09-03 -> 02:00 09-04, and from 20:35
for lab,a,b in [('ASIA 20:35->02:00','2026-09-03 20:35:00','2026-09-04 02:00:00'),('20:35->08:30','2026-09-03 20:35:00','2026-09-04 08:30:00')]:
    r=c.execute("SELECT MAX(h),MIN(l) FROM bars WHERE symbol='MNQ' AND tf='1m' AND open_time_ms>=? AND open_time_ms<?",(ms(a),ms(b))).fetchone()
    print("%s  high %.2f low %.2f"%(lab,r[0],r[1]))
