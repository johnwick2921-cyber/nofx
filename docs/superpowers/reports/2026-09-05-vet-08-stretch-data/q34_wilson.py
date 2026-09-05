import math
def wilson(k,n,z=1.96):
    if n==0: return (0,0,0)
    p=k/n; d=1+z*z/n; c=p+z*z/(2*n); h=z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))
    return p,(c-h)/d,(c+h)/d
rows=[
("arm-legs authored in window that ever confirmed (touch/close) while their version lived",18,63),
("arm-legs whose confirm was NEVER met while their version lived",45,63),
("arm-legs filled / authored (Run A, replay_out.txt)",3,63),
("arm-legs filled / confirmed (Run A)",3,18),
("live ledger arms filled / rows, ids 29-37",1,9),
("live ledger arms filled / rows ids 5-37 excl the 2 TEST-E7 seams (ids 15,16)",9,31),
("live ledger arms marketable-cancelled ids 23-37 (Part D's window)",5,15),
("live ledger arms marketable-cancelled ids 5-37 excl the 2 TEST-E7 seams",6,31),
("live ledger arms marketable-cancelled ids 29-37",3,9),
("counterfactual reclaim stop-entries passing R:R>=2 on target_chain[0] (Run B, q35)",4,24),
("counterfactual reclaim stop-entries passing R:R>=2 on target_chain[-1] (q35)",12,24),
("counterfactual reclaim stop-entries filled (Run B)",0,24),
("decision-path open intents executed 09-02..09-04",4,24),
("decision-path open intents refused by strict 09-03",13,13),
("RTH-L break / (hold+break) all-time touch_outcomes",82,124),
("RTH-L break / (hold+break) LONDON window 09-02..09-04",18,26),
("win rate COMPLIANT era rows 58 (wins/(wins+losses))",18,56),
("win rate broader 65-row cut incl 7 UNRESOLVABLE (sensitivity only)",21,63),
("MFE>=2R COMPLIANT era rows with recoverable stop (51)",13,51),
("MFE>=2R broader 55-row cut (sensitivity only)",13,55),
("long scenarios arm-enabled 09-03",2,28),
("short scenarios arm-enabled 09-03",8,20),
("long scenarios arm-enabled 09-02",9,64),
("short scenarios arm-enabled 09-02",27,43),
("09-03 decisions proposing open_long / decisions",0,596),
("stops within 9 pts of the next-60m extreme (589,590) — 591 is a continuation",2,5),
("stop-entry submissions on 09-04 with Stop price=0 at NT8 (q37)",21,21),
]
for name,k,n in rows:
    p,lo,hi=wilson(k,n)
    print(f"{name}: {k}/{n} = {p:.3f} Wilson95 [{lo:.3f}, {hi:.3f}]")

rows2=[("confirm ref sited <=1.5xATR5m from the birth close: ever confirmed",11,20),
("confirm ref sited >1.5xATR5m from the birth close: ever confirmed",7,43),
("IB-break days qualifying under a day_type=trend + matching-bias gate",2,10)]
for name,k,n in rows2:
    p_,lo,hi=wilson(k,n); print(f"{name}: {k}/{n} = {p_:.3f} Wilson95 [{lo:.3f}, {hi:.3f}]")
