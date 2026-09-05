# Why 45 of 63 legs never confirmed: distance from the confirm ref to the price at
# the version's BIRTH, in ATR5m units, split by whether the leg ever confirmed.
import sqlite3, csv, datetime, bisect, math, collections, os
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro',uri=True)
B=c.execute("SELECT open_time_ms,o,h,l,c FROM bars WHERE symbol='MNQ' AND tf='1m' ORDER BY open_time_ms").fetchall()
bk=[b[0] for b in B]
def close_at(ms):
    i=bisect.bisect_right(bk,ms)-1
    return B[i][4] if i>=0 else None
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
births={}
for td,sess,v,ca in c.execute("SELECT trade_date,session,version,created_at FROM plans"):
    ms=int(datetime.datetime.strptime(ca[:19],'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp()*1000)
    if '-05:00' in ca: ms+=5*3600*1000
    births[(td,sess,int(v))]=ms
rows=list(csv.DictReader(open(os.path.join(os.path.dirname(os.path.abspath(__file__)),'replay_A_arms.csv'))))
conf=[];never=[];pairs=[]
for r in rows:
    ms=births.get((r['td'],r['sess'],int(r['v'])))
    if ms is None: continue
    px=close_at(ms); a=atr_at(ms)
    try: ref=float(r['c_ref'])
    except: continue
    if px is None or not a: continue
    d=abs(ref-px)/a; met=bool(r['confirm_met_at'].strip())
    pairs.append((d,met)); (conf if met else never).append(d)
def q(a,p):
    a=sorted(a); i=(len(a)-1)*p; lo=int(i); hi=min(lo+1,len(a)-1)
    return a[lo]+(a[hi]-a[lo])*(i-lo)
for lab,a in (('confirmed',conf),('never-confirmed',never)):
    print("%-16s n=%2d  p25 %.2f  median %.2f  p75 %.2f  (xATR5m at version birth)"%(lab,len(a),q(a,.25),q(a,.5),q(a,.75)))
allp=sorted(pairs); n1,n2=len(conf),len(never); rk=[0.0]*len(allp); i=0
while i<len(allp):
    j=i
    while j+1<len(allp) and allp[j+1][0]==allp[i][0]: j+=1
    r=(i+j)/2+1
    for t in range(i,j+1): rk[t]=r
    i=j+1
R1=sum(rk[t] for t in range(len(allp)) if allp[t][1])
U1=R1-n1*(n1+1)/2; mu=n1*n2/2; sd=math.sqrt(n1*n2*(n1+n2+1)/12); z=(U1-mu)/sd
print("Mann-Whitney U1=%.1f z=%.2f two-sided p=%.4f"%(U1,z,2*(1-0.5*(1+math.erf(abs(z)/math.sqrt(2))))))
def wil(k,n,zz=1.96):
    p=k/n; dd=1+zz*zz/n; ctr=(p+zz*zz/(2*n))/dd; h=zz*math.sqrt(p*(1-p)/n+zz*zz/(4*n*n))/dd
    return 100*(ctr-h),100*(ctr+h)
for lab,sel in (('<=1.5xATR',lambda d:d<=1.5),(' >1.5xATR',lambda d:d>1.5)):
    s=[p for p in pairs if sel(p[0])]; k=sum(1 for p in s if p[1])
    print("confirm rate, ref %s: %d/%d = %.1f%% Wilson [%.1f, %.1f]"%(lab,k,len(s),100*k/len(s),*wil(k,len(s))))
byses=collections.defaultdict(list)
for td,sess,v,ca in c.execute("SELECT trade_date,session,version,created_at FROM plans WHERE trade_date BETWEEN '2026-09-02' AND '2026-09-04'"):
    ms=int(datetime.datetime.strptime(ca[:19],'%Y-%m-%d %H:%M:%S').replace(tzinfo=datetime.timezone.utc).timestamp()*1000)
    if '-05:00' in ca: ms+=5*3600*1000
    byses[(td,sess)].append(ms)
gaps=[]
for k,v in byses.items():
    v.sort()
    for a,b in zip(v,v[1:]): gaps.append((b-a)/60000.0)
gaps.sort()
print("version LIFETIME (inter-version gaps in the window) n=%d median %.1f min  mean %.1f min"%(len(gaps),gaps[len(gaps)//2],sum(gaps)/len(gaps)))
