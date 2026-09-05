import sqlite3, math, json, datetime
c=sqlite3.connect('file:/home/hoang/nofx/data/data.db?mode=ro', uri=True)
era=int(datetime.datetime(2026,8,15,5,0).timestamp()*1000)
# REVISION 2026-09-05: the standing rules require plan_id='UNRESOLVABLE' excluded
# too. Primary figures are the COMPLIANT population; the broader cut is printed
# below it as a labelled sensitivity only.
Q="""SELECT id, source, side, entry_price, exit_price, entry_time, pnl_corrected, mae, mfe, plan_session, entry_order_id
 FROM trader_positions WHERE entry_time>=? AND source<>'e7_farside_test' AND pnl_corrected IS NOT NULL {extra} ORDER BY entry_time"""
rows=c.execute(Q.format(extra="AND plan_id<>'UNRESOLVABLE'"),(era,)).fetchall()
broad=c.execute(Q.format(extra=""),(era,)).fetchall()
unres=c.execute("SELECT id,pnl_corrected FROM trader_positions WHERE entry_time>=? AND source<>'e7_farside_test' AND pnl_corrected IS NOT NULL AND plan_id='UNRESOLVABLE' ORDER BY id",(era,)).fetchall()
n=len(rows); print("COMPLIANT era rows n=",n, "(era>=2026-08-15 00:00 CT; excl e7_farside_test, pnl_corrected NULL, plan_id='UNRESOLVABLE')")
print("excluded plan_id='UNRESOLVABLE':", len(unres), "ids", [u[0] for u in unres], "sum %.2f"%sum(u[1] for u in unres))
print("broader 65-row cut (SENSITIVITY ONLY) n=", len(broad))
excl=c.execute("SELECT COUNT(*) FROM trader_positions WHERE entry_time>=? AND source<>'e7_farside_test' AND pnl_corrected IS NULL",(era,)).fetchone()[0]
print("excluded NULL pnl_corrected:",excl)
w=[r[6] for r in rows if r[6]>0]; l=[r[6] for r in rows if r[6]<0]; z=[r for r in rows if r[6]==0]
print("wins",len(w),"losses",len(l),"scratch",len(z),"sum",round(sum(r[6] for r in rows),2))
def wilson(k,n,z=1.96):
    if n==0: return (0,0)
    p=k/n; d=1+z*z/n; ctr=(p+z*z/(2*n))/d; h=z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))/d
    return (ctr-h, ctr+h)
k=len(w); print("win rate (wins/(wins+losses)) = %d/%d = %.4f Wilson [%.4f,%.4f]"%(k,k+len(l),k/(k+len(l)),*wilson(k,k+len(l))))
print("win rate incl scratch = %d/%d = %.4f"%(k,n,k/n))
aw=sum(w)/len(w); al=-sum(l)/len(l); print("avgW %.2f avgL %.2f payoff %.3f breakeven WR %.4f"%(aw,al,aw/al,1/(1+aw/al)))
# MFE >= 2R: recover the authored stop for system rows from decision_records open action within [-4min,+1min]
mfe2=0; withstop=0; ids2=[]; idsw=[]
for r in rows:
    pid,src,side,ep,xp,et,pc,mae,mfe,sess,eoid=r
    stop=None
    if src in('armed_entry','reconcile'):
        a=c.execute("SELECT stop_px FROM armed_orders WHERE state='filled' AND abs(fill_price-?)<0.01 AND abs(strftime('%s',created_at)*1000-?)<7200000 ORDER BY id DESC LIMIT 1",(ep,et)).fetchone()
        if a: stop=a[0]
    if stop is None:
        t0=datetime.datetime.utcfromtimestamp(et/1000-240).strftime('%Y-%m-%d %H:%M:%S'); t1=datetime.datetime.utcfromtimestamp(et/1000+90).strftime('%Y-%m-%d %H:%M:%S')
        d=c.execute("SELECT decision_json FROM decision_records WHERE timestamp BETWEEN ? AND ? AND decision_json LIKE '%open_%' ORDER BY timestamp DESC LIMIT 1",(t0,t1)).fetchone()
        if d:
            try:
                for a in json.loads(d[0]):
                    if str(a.get('action','')).startswith('open'): stop=a.get('stop_loss')
            except: pass
    if stop and mfe is not None:
        risk=abs(ep-stop)
        if risk>0:
            withstop+=1; idsw.append(pid)
            if mfe>=2*risk: mfe2+=1; ids2.append(pid)
print("rows with recoverable authored stop:",withstop, "ids:",idsw)
print("MFE >= 2R: %d/%d Wilson [%.3f,%.3f] ids=%s"%(mfe2,withstop,*wilson(mfe2,withstop),ids2))
# planned R:R of arms (all authored arms, era)
arms=c.execute("SELECT id, side, entry_px, stop_px, target_px FROM armed_orders WHERE entry_px>0 AND stop_px>0 AND target_px>0").fetchall()
rr=[]
for a in arms:
    risk=abs(a[2]-a[3]); rew=abs(a[4]-a[2])
    if risk>0: rr.append(rew/risk)
rr.sort(); print("arms with geometry n=%d planned R:R median %.2f min %.2f max %.2f"%(len(rr), rr[len(rr)//2], rr[0], rr[-1]))

# sensitivity line on the broader 65-row cut
wb=[r[6] for r in broad if r[6]>0]; lb=[r[6] for r in broad if r[6]<0]
print("SENSITIVITY (65-row cut, includes the 7 UNRESOLVABLE): wins %d losses %d sum %.2f win rate %d/%d = %.4f avgW %.2f avgL %.2f payoff %.3f"%(
  len(wb),len(lb),sum(r[6] for r in broad),len(wb),len(wb)+len(lb),len(wb)/(len(wb)+len(lb)),
  sum(wb)/len(wb), -sum(lb)/len(lb), (sum(wb)/len(wb))/(-sum(lb)/len(lb))))
