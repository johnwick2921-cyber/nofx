import sqlite3, math, datetime
from collections import defaultdict
def wilson(k,n,z=1.96):
    p=k/n; d=1+z*z/n; c=p+z*z/(2*n); h=z*math.sqrt(p*(1-p)/n+z*z/(4*n*n)); return (round(100*(c-h)/d,1),round(100*(c+h)/d,1))
con=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro',uri=True)
rows=con.execute("SELECT open_time_ms,o,c FROM bars WHERE tf='1d' AND symbol='MNQ' ORDER BY open_time_ms").fetchall()
print("1d MNQ bars:",len(rows))
for drop_flat in (True,False):
    up=defaultdict(int); n=defaultdict(int); flat=defaultdict(int)
    for t,o,c in rows:
        dt=datetime.datetime.fromtimestamp(t/1000, datetime.UTC)-datetime.timedelta(hours=5)+datetime.timedelta(days=1)
        wd=dt.strftime('%a')
        if c==o:
            flat[wd]+=1
            if drop_flat: continue
        n[wd]+=1; up[wd]+=(c>o)
    print("drop_flat=",drop_flat)
    for wd in ['Mon','Tue','Wed','Thu','Fri']:
        print("  ",wd,"n=",n[wd],"up=",up[wd],"share=",round(100*up[wd]/n[wd],1),wilson(up[wd],n[wd]),"flat=",flat[wd])
