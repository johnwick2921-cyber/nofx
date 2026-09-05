# COMMITTED IN THE REVISION PASS (the first draft shipped q35_reclaim_targets.txt with
# no generator, so a reader could not reconcile it against replay_B_arms.csv's rr column;
# the two differ because THIS measures the 1.5xATR floor at the version's BIRTH while the
# replay's rr is the last cycle the version lived — one version late).
import sqlite3, json, datetime, bisect, os
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro',uri=True)
B=c.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' ORDER BY open_time_ms").fetchall()
agg={}
for ms,o,h,l,cl in B:
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
    if atr is not None: avail[k+300000]=atr
ak=sorted(avail)
def atr_at(ms):
    i=bisect.bisect_right(ak,ms)-1
    return avail[ak[i]] if i>=0 else None
rows=c.execute("""SELECT p.trade_date,p.session,p.version,p.created_at,p.doc FROM plans p
 WHERE p.trade_date BETWEEN '2026-09-02' AND '2026-09-04' ORDER BY p.trade_date,p.session,p.version""").fetchall()
out=[]
for td,sess,v,ca,doc in rows:
    ms=int(datetime.datetime.strptime(ca[:19],'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp()*1000)
    if '-05:00' in ca: ms+=5*3600*1000
    a=atr_at(ms)
    try: d=json.loads(doc)
    except: continue
    for sc in d.get('scenarios',[]):
        if sc.get('condition')!='reclaim': continue
        if (sc.get('arm') or {}).get('enabled'): continue   # the 25th reclaim WAS armed (live row 38)
        lvl=sc.get('trigger_price') or (sc.get('confirm') or {}).get('ref_price')
        chain=sc.get('target_chain') or []
        if lvl is None or not chain or not a: continue
        floor=1.5*a
        r0=abs(chain[0]-lvl)/floor; rn=abs(chain[-1]-lvl)/floor
        out.append((td,sess,v,sc.get('id'),sc.get('direction'),lvl,chain[0],chain[-1],round(a,1),round(r0,2),round(rn,2)))
print("counterfactual reclaim scenarios:",len(out))
for r in out: print(" ".join(str(x) for x in r))
p0=sum(1 for r in out if r[9]>=2.0); pn=sum(1 for r in out if r[10]>=2.0)
m0=sorted(r[9] for r in out); mn=sorted(r[10] for r in out)
print("pass 2R with first target: %d/%d; with farthest target: %d/%d (floor = 1.5xATR5m at birth, no anchor)"%(p0,len(out),pn,len(out)))
def med(a):
    n=len(a); return a[n//2] if n%2 else (a[n//2-1]+a[n//2])/2
print("median first-target R: %.2f  median farthest-target R: %.2f  (even n=%d -> mean of the two middle values)"%(med(m0),med(mn),len(m0)))
