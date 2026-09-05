# REGENERATED 2026-09-05 (revision): the first version double-converted CT->UTC in
# atr_at(), so every lookup read the ATR ~5h late and the 09-04 rows clamped to the
# last bar (23.38 repeated at three different times). Fixed: parse planner_read_facts
# .created_at as UTC-with-offset, compare in epoch ms, no naive-local round trip.
import sqlite3, datetime, bisect
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro', uri=True)
rows=c.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' ORDER BY open_time_ms").fetchall()
agg={}
for ms,o,h,l,cl in rows:
    k=ms-(ms%300000)
    if k not in agg: agg[k]=[o,h,l,cl]
    else:
        a=agg[k]; a[1]=max(a[1],h); a[2]=min(a[2],l); a[3]=cl
keys=sorted(agg); atr=None; prev=None; trs=[]; avail={}
for k in keys:
    o,h,l,cl=agg[k]
    tr=h-l if prev is None else max(h-l,abs(h-prev),abs(l-prev)); prev=cl
    if atr is None:
        trs.append(tr)
        if len(trs)==14: atr=sum(trs)/14
    else: atr=(atr*13+tr)/14
    if atr is not None: avail[k+300000]=atr   # available once the bucket closes
ak=sorted(avail)
def atr_at(ms):
    i=bisect.bisect_right(ak,ms)-1
    return avail[ak[i]] if i>=0 else None
print("5m buckets aggregated from 1m:", len(keys))
facts=c.execute("SELECT id,trade_date,session,created_at,atr5m,stop_floor_pts FROM planner_read_facts ORDER BY id").fetchall()
d=[]
print("%3s %-10s %-7s %-19s %8s %8s %8s %9s"%("id","td","sess","read_ct","recon","recorded","diff","1.5x rec"))
for rid,td,sess,ca,a5,fl in facts:
    ms=int(datetime.datetime.strptime(ca[:19],'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp()*1000)
    ct=datetime.datetime.fromtimestamp(ms/1000-5*3600, datetime.timezone.utc).strftime('%Y-%m-%d %H:%M:%S')
    r=atr_at(ms)
    if r is None: continue
    d.append(r-a5)
    print("%3d %-10s %-7s %-19s %8.2f %8.2f %+8.2f %9.2f"%(rid,td,sess,ct,r,a5,r-a5,fl))
print("n=%d  min %+.2f  max %+.2f  mean %+.2f"%(len(d),min(d),max(d),sum(d)/len(d)))
for ct in ['2026-09-03 10:00:00','2026-09-03 09:15:31','2026-09-03 09:44:33','2026-09-03 10:14:45','2026-09-03 10:49:57','2026-09-03 11:27:53']:
    ms=int(datetime.datetime.strptime(ct,'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp()*1000)+5*3600*1000
    a=atr_at(ms); print("%s CT  ATR5m %.2f  1.5x %.2f"%(ct,a,1.5*a))
