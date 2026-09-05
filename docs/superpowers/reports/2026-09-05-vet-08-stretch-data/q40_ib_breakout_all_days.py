import sqlite3, datetime, bisect
TICK=0.25
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro', uri=True)
rows=c.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' ORDER BY open_time_ms").fetchall()
def ct(ms): return datetime.datetime.utcfromtimestamp(ms/1000-5*3600)
# 5m ATR (Wilder 14) availability map
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
    if atr is not None: avail[k+300000]=atr
ak=sorted(avail)
def atr_at(ms):
    i=bisect.bisect_right(ak,ms)-1
    return avail[ak[i]] if i>=0 else None
# group by CT day
days={}
for ms,o,h,l,cl in rows:
    d=ct(ms).strftime('%Y-%m-%d'); days.setdefault(d,[]).append((ms,o,h,l,cl))
res=[]
for d in sorted(days):
    bars=days[d]
    def between(a,b): return [x for x in bars if a<=ct(x[0]).strftime('%H:%M')<=b]
    ib=between('08:30','09:29')
    if len(ib)<60: print("%s SKIP (IB bars %d)"%(d,len(ib))); continue
    rth=[x for x in bars if '09:30'<=ct(x[0]).strftime('%H:%M')<='14:45']
    if not rth: print("%s SKIP (no RTH after IB)"%d); continue
    ibh=max(x[2] for x in ib); ibl=min(x[3] for x in ib)
    up=ibh+TICK; dn=ibl-TICK
    trade=None
    for i,(ms,o,h,l,cl) in enumerate(rth):
        hitU = h>=up; hitD = l<=dn
        if hitU and hitD:   # ambiguous bar -> take the one nearer the open
            side='long' if abs(o-up)<abs(o-dn) else 'short'
        elif hitU: side='long'
        elif hitD: side='short'
        else: continue
        entry = up if side=='long' else dn
        a=atr_at(ms); stop_d=1.5*a
        stop = entry-stop_d if side=='long' else entry+stop_d
        tgt  = entry+2*stop_d if side=='long' else entry-2*stop_d
        out=None
        for ms2,o2,h2,l2,c2 in rth[i:]:
            if side=='long':
                s = l2<=stop; t = h2>=tgt
            else:
                s = h2>=stop; t = l2<=tgt
            if s: out=('stop',-stop_d,ms2); break
            if t: out=('target',2*stop_d,ms2); break
        if out is None:
            last=rth[-1][4]
            pts = (last-entry) if side=='long' else (entry-last)
            out=('flat',pts,rth[-1][0])
        trade=dict(d=d,side=side,entry=entry,atr=a,stop=stop,tgt=tgt,fill_ct=ct(ms).strftime('%H:%M'),
                   res=out[0],pts=out[1],exit_ct=ct(out[2]).strftime('%H:%M'))
        break
    if trade is None:
        print("%s  no IB break (IBH %.2f IBL %.2f)"%(d,ibh,ibl)); res.append(None); continue
    print("%s  %-5s fill %s @%.2f  ATR %.2f stop %.2f tgt %.2f -> %-6s %+8.2f at %s"%(
        d,trade['side'],trade['fill_ct'],trade['entry'],trade['atr'],trade['stop'],trade['tgt'],trade['res'],trade['pts'],trade['exit_ct']))
    res.append(trade)
t=[r for r in res if r]
w=[r for r in t if r['pts']>0]; l=[r for r in t if r['pts']<=0]
print("\nn=%d breaks (%d no-break days). W %d / L %d  sum %+.2f pts  mean %+.2f pts/trade = $%.0f/trade"%(
    len(t), sum(1 for r in res if r is None), len(w), len(l), sum(r['pts'] for r in t), sum(r['pts'] for r in t)/len(t), 2*sum(r['pts'] for r in t)/len(t)))
