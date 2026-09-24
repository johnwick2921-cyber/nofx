package trader

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/market"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// ── W1b E1 + E2 — a working arm under a NEW plan version ────────────────────
//
// E1: the churn guard compared the new bracket with itself (the row was built
// from the new leg and only its id was copied from the ledger), so a working
// order kept the stop and target it was placed with through every re-spec.
// E2: nothing compared a working market_in_zone limit with the CURRENT
// version's zone, so a moved zone left the old limit resting until the
// 30-minute rest cap. Both now go through ONE function (arm_respec.go) that
// asks for at most one cancel per row, through the filled-arm guard; the
// normal authoring then re-arms under the new version once the broker's book
// confirms the cancel. Every test below drives maybeManageArmedOrdersAt over
// the real TCP wire (canon 53).

// respecTape is zoneTape with a chosen bar half-range: a wider range raises
// ATR5m, which moves the composed market_in_zone stop (near − 1.5×ATR5m)
// WITHOUT any plan change — the drift the version gate exists to ignore.
func respecTape(last float64, now time.Time, half float64) []market.Kline {
	out := zoneTape(last, now, 0)
	for i := range out {
		out[i].High = out[i].Close + half
		out[i].Low = out[i].Close - half
	}
	return out
}

// restingBook is the AddOn's snapshot with the entry `sid` resting at px.
func (r *zoneRig) restingBook(at time.Time, sid string, px float64) {
	r.srv.OrderSnapshots().PutAt(ntwire.OrderSnapshotPayload{Account: "Sim101", Orders: []ntwire.NT8Order{
		{OrderID: "o-" + sid, Name: sid, Symbol: "MNQ", Action: "buy", Type: "limit", LimitPrice: px, Quantity: 1, State: "Working"},
	}}, at)
}

// persistFlat is the PERSISTED empty snapshot the settlement pass reads as
// evidence that the cancelled order is gone.
func (r *zoneRig) persistFlat(at time.Time) {
	r.t.Helper()
	if err := r.st.NT8OrderSnapshots().Insert(&store.NT8OrderSnapshot{Account: "Sim101", OrdersJSON: "[]", ReceivedMs: at.UnixMilli()}); err != nil {
		r.t.Fatal(err)
	}
}

// placeWorking runs the first pass, returns the placed signal, and moves the
// row to working through the production order_update handler.
func (r *zoneRig) placeWorking(wantLimit float64) ntwire.SignalPayload {
	r.t.Helper()
	r.setTape(zoneTape(101.95, r.now, 0)) // beyond the zone: rests at the far edge
	r.at.maybeManageArmedOrdersAt(nil, r.now)
	sigs, _ := r.drain()
	if len(sigs) != 1 || sigs[0].LimitPrice != wantLimit {
		r.t.Fatalf("fixture: want one limit at %.2f, got %+v", wantLimit, sigs)
	}
	r.at.onArmedOrderUpdate(ntwire.OrderUpdatePayload{SignalID: sigs[0].SignalID, State: "Working", Account: "Sim101"}, r.st.ArmedOrders())
	if row := r.row("S1"); row.State != store.StateWorking || row.Version != 1 {
		r.t.Fatalf("fixture: the placed row must be working under v1: %+v", row)
	}
	return sigs[0]
}

// newVersion appends the next plan version (the provider clock stays fixed).
func (r *zoneRig) newVersion(scs ...kernel.PlanScenario) {
	r.t.Helper()
	blob, _ := json.Marshal(zoneDoc(scs...))
	shadowPlanAtTime(r.t, r.at, r.st, string(blob), r.now)
}

// settleAndReArm confirms the requested cancel from a fresh persisted book at
// t2, then runs the re-arm pass at t3 on a flat book and returns its frames.
func (r *zoneRig) settleAndReArm() ([]ntwire.SignalPayload, []ntwire.CancelOrderPayload) {
	r.t.Helper()
	t2 := r.now.Add(2 * time.Minute)
	r.persistFlat(t2)
	r.flatBook(t2)
	r.setTape(zoneTape(101.95, t2, 0))
	r.at.maybeManageArmedOrdersAt(nil, t2)
	if sigs, cancels := r.drain(); len(sigs) != 0 || len(cancels) != 0 {
		r.t.Fatalf("the settlement pass must send nothing: sigs=%+v cancels=%+v", sigs, cancels)
	}
	var first store.ArmedOrderDB
	for _, x := range r.rows() {
		if x.Scenario == "S1" && x.PlacementSeq == 0 {
			first = x
		}
	}
	if first.State != store.StateCancelled {
		r.t.Fatalf("a fresh flat persisted book must settle the re-spec cancel: %+v", first)
	}
	// The adapter's B3 duplicate guard runs on the WALL clock; the fixture
	// clock says minutes passed, so the adapter is re-made to stand for that.
	r.at.trader = ntTrader.NewTCPTrader(r.srv, "MNQ", "Sim101")
	t3 := r.now.Add(3 * time.Minute)
	r.flatBook(t3)
	r.setTape(zoneTape(101.95, t3, 0))
	r.at.maybeManageArmedOrdersAt(nil, t3)
	return r.drain()
}

func onGridWithin(got, want, tick float64) bool {
	steps := got / tick
	return math.Abs(steps-math.Round(steps)) < 1e-6 && math.Abs(got-want) <= tick+1e-9
}

// E1 — a new version that moves the bracket ≥ 2 ticks cancels the working
// order (the AddOn cannot modify a resting entry's bracket) and the scenario
// re-arms under the new version with the NEW stop and target.
func TestWorkingArmBracketRespecByNewVersionReplacesTheOrder(t *testing.T) {
	r := newZoneRig(t, "w1b-respec-e1", zoneDoc(zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)))
	first := r.placeWorking(100.5)
	sid := first.SignalID
	if first.StopLoss <= 96.5 {
		t.Fatalf("fixture: the v1 stop must sit above the v2 stop by ≥ 2 ticks, got %.2f", first.StopLoss)
	}
	before, _ := store.SystemCounter(r.st, "arm:respec_cancel")

	// v2: the same zone, the bracket re-spec'd to SL 96 / TP 112.
	sc := zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)
	sc.Arm.Stop, sc.Arm.Target = 96, 112
	sc.TargetChain = []float64{112}
	r.newVersion(sc)
	t1 := r.now.Add(time.Minute)
	r.restingBook(t1, sid, 100.5)
	r.setTape(zoneTape(101.95, t1, 0))
	r.at.maybeManageArmedOrdersAt(nil, t1)
	sigs, cancels := r.drain()
	if len(cancels) != 1 || cancels[0].SignalID != sid || len(sigs) != 0 {
		t.Fatalf("a re-spec'd bracket must cancel the working order %s exactly once and place nothing: sigs=%+v cancels=%+v", sid, sigs, cancels)
	}
	row := r.row("S1")
	if row.State != store.StateCancelPending || !strings.Contains(row.StateReason, "bracket re-spec by v2") {
		t.Fatalf("the working row must be cancel_pending 'bracket re-spec by v2': %+v", row)
	}
	if n, _ := store.SystemCounter(r.st, "arm:respec_cancel"); n != before+1 {
		t.Fatalf("the re-spec cancel must be counted once (%d → %d)", before, n)
	}

	sigs, cancels = r.settleAndReArm()
	if len(cancels) != 0 || len(sigs) != 1 {
		t.Fatalf("the settled scenario must re-arm and place once under v2: sigs=%+v cancels=%+v", sigs, cancels)
	}
	if !onGridWithin(sigs[0].StopLoss, 96, 0.25) || sigs[0].StopLoss > 96 || !onGridWithin(sigs[0].TakeProfit, 112, 0.25) {
		t.Fatalf("the re-placed order must carry the v2 bracket (SL ≤ 96 on grid, TP 112): %+v", sigs[0])
	}
	if sigs[0].SignalID == sid {
		t.Fatalf("the re-placement must be a NEW signal, got the old %s", sid)
	}
	rows := r.rows()
	if len(rows) != 2 {
		t.Fatalf("the re-arm mints ONE successor row: %+v", rows)
	}
	nw := r.row("S1")
	if nw.PlacementSeq != 1 || nw.Version != 2 || nw.SignalID != sigs[0].SignalID {
		t.Fatalf("the successor must be placement_seq 1 under v2 carrying the new signal: %+v", nw)
	}
}

// E2 — a new version whose zone no longer contains the working limit cancels
// it ("zone moved by v2") and re-arms at the NEW zone's far edge.
func TestZoneMovedByNewVersionCancelsTheWorkingLimitAndReArms(t *testing.T) {
	r := newZoneRig(t, "w1b-respec-e2", zoneDoc(zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)))
	sid := r.placeWorking(100.5).SignalID
	before, _ := store.SystemCounter(r.st, "market_in_zone:zone_moved")

	r.newVersion(zoneScenario("S1", kernel.EntryPolicyMarketInZone, []float64{99.0, 100.0}, false))
	t1 := r.now.Add(time.Minute)
	r.restingBook(t1, sid, 100.5)
	r.setTape(zoneTape(101.95, t1, 0))
	r.at.maybeManageArmedOrdersAt(nil, t1)
	sigs, cancels := r.drain()
	if len(cancels) != 1 || cancels[0].SignalID != sid || len(sigs) != 0 {
		t.Fatalf("a moved zone must cancel the working limit %s exactly once and place nothing: sigs=%+v cancels=%+v", sid, sigs, cancels)
	}
	row := r.row("S1")
	if row.State != store.StateCancelPending || !strings.Contains(row.StateReason, "zone moved by v2") {
		t.Fatalf("the working row must be cancel_pending 'zone moved by v2': %+v", row)
	}
	if n, _ := store.SystemCounter(r.st, "market_in_zone:zone_moved"); n != before+1 {
		t.Fatalf("the zone-moved cancel must be counted once (%d → %d)", before, n)
	}

	sigs, cancels = r.settleAndReArm()
	if len(cancels) != 0 || len(sigs) != 1 || sigs[0].LimitPrice != 100.0 {
		t.Fatalf("the settled scenario must re-arm at the NEW far edge 100.00: sigs=%+v cancels=%+v", sigs, cancels)
	}
	rows := r.rows()
	if len(rows) != 2 {
		t.Fatalf("the re-arm mints ONE successor row: %+v", rows)
	}
	nw := r.row("S1")
	if nw.PlacementSeq != 1 || nw.Version != 2 || nw.State != store.StatePlacePending || nw.EntryPx != 100.0 {
		t.Fatalf("the successor must be placement_seq 1, v2, place_pending at 100.00: %+v", nw)
	}
}

// The version gate: live ATR moves the composed stop by ≥ 2 ticks inside ONE
// plan version — that is not a re-spec, and nothing is sent.
func TestRespecIgnoresATRDriftWithinOneVersion(t *testing.T) {
	r := newZoneRig(t, "w1b-respec-atr", zoneDoc(zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)))
	first := r.placeWorking(100.5)
	t1 := r.now.Add(time.Minute)
	r.restingBook(t1, first.SignalID, 100.5)
	r.setTape(respecTape(101.95, t1, 0.8)) // ATR5m up → composed stop ≥ 2 ticks lower
	r.at.maybeManageArmedOrdersAt(nil, t1)
	if sigs, cancels := r.drain(); len(sigs) != 0 || len(cancels) != 0 {
		t.Fatalf("ATR drift inside one version must send nothing: sigs=%+v cancels=%+v", sigs, cancels)
	}
	if row := r.row("S1"); row.State != store.StateWorking || row.Version != 1 {
		t.Fatalf("the working row must stay working under v1: %+v", row)
	}
}

// A new version whose zone still contains the limit and whose bracket is
// unchanged leaves the working order alone.
func TestRespecLeavesAContainedUnchangedOrderAlone(t *testing.T) {
	r := newZoneRig(t, "w1b-respec-same", zoneDoc(zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)))
	sid := r.placeWorking(100.5).SignalID
	r.newVersion(zoneScenario("S1", kernel.EntryPolicyMarketInZone, []float64{99.5, 100.75}, false))
	t1 := r.now.Add(time.Minute)
	r.restingBook(t1, sid, 100.5)
	r.setTape(zoneTape(101.95, t1, 0))
	r.at.maybeManageArmedOrdersAt(nil, t1)
	if sigs, cancels := r.drain(); len(sigs) != 0 || len(cancels) != 0 {
		t.Fatalf("a contained limit with an unchanged bracket must send nothing: sigs=%+v cancels=%+v", sigs, cancels)
	}
	if row := r.row("S1"); row.State != store.StateWorking {
		t.Fatalf("the working row must stay working: %+v", row)
	}
}

// The filled-arm guard decides: a book that shows only the signal's bracket
// children (the entry filled) REFUSES the cancel — nothing is sent, the row
// stays working — and the next pass, on a book with the entry resting, sends it.
func TestRespecCancelRefusedByTheFilledArmGuardRetriesNextPass(t *testing.T) {
	r := newZoneRig(t, "w1b-respec-guard", zoneDoc(zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)))
	sid := r.placeWorking(100.5).SignalID
	r.newVersion(zoneScenario("S1", kernel.EntryPolicyMarketInZone, []float64{99.0, 100.0}, false))
	t1 := r.now.Add(time.Minute)
	r.srv.OrderSnapshots().PutAt(ntwire.OrderSnapshotPayload{Account: "Sim101", Orders: []ntwire.NT8Order{
		{OrderID: "sl", Name: sid + "-SL", Symbol: "MNQ", Action: "sell", Type: "stop", StopPrice: 97.5, Quantity: 1, State: "Working"},
	}}, t1)
	r.setTape(zoneTape(101.95, t1, 0))
	r.at.maybeManageArmedOrdersAt(nil, t1)
	if _, cancels := r.drain(); len(cancels) != 0 {
		t.Fatalf("the filled-arm guard must refuse a cancel that would reach the protections: %+v", cancels)
	}
	if row := r.row("S1"); row.State != store.StateWorking {
		t.Fatalf("a refused cancel leaves the row working: %+v", row)
	}
	t2 := t1.Add(30 * time.Second)
	r.restingBook(t2, sid, 100.5)
	r.setTape(zoneTape(101.95, t2, 0))
	r.at.maybeManageArmedOrdersAt(nil, t2)
	if _, cancels := r.drain(); len(cancels) != 1 || cancels[0].SignalID != sid {
		t.Fatalf("the next pass, entry resting, must send the cancel once: %+v", cancels)
	}
}

// The refresh-write skip (critic correction to E1 step 6): a non-armed row's
// refresh write is skipped ONLY when it carries the same opportunity. A
// WORKING row holding opportunity A under a key a later version gives to B
// still reaches UpsertArm, so W5 R13(a)'s typed refusal is named (WARN +
// counter, once) and B's admit is withdrawn — nothing is placed for B, and
// A's working row is neither rewritten nor cancelled (it is a Picture row).
func TestRespecSkipKeepsTheTypedSourceRefusalForAWorkingRow(t *testing.T) {
	r, epoch := newPicRig(t, "w1b-respec-r13", nil)
	scA := picScenario("P1", "opp-respec-a", r.now, epoch, picDefault)
	picPlan(r, scA)
	seed := &store.ArmedOrderDB{TraderID: r.at.id, PlanID: r.pid, Version: 1, Session: "TEST", Scenario: "P1", Side: "long",
		EntryPx: 100.5, StopPx: 97, TargetPx: 110, State: store.StateArmed, EntryClass: "armed_fill", Kind: "limit",
		Policy: store.ArmPolicyMarketInZone, CreatedAt: r.now, UpdatedAt: r.now}
	stampPictureSource(seed, scA)
	ledger := r.st.ArmedOrders()
	if err := ledger.UpsertArm(seed); err != nil {
		t.Fatal(err)
	}
	a := r.rows()[0]
	if err := ledger.BeginPlacement(a.ID, "sig-respec-a"); err != nil {
		t.Fatal(err)
	}
	if err := ledger.ApplyPlacementReceipt(r.at.id, "sig-respec-a", store.StateWorking, "fixture"); err != nil {
		t.Fatal(err)
	}
	if row := r.rows()[0]; row.State != store.StateWorking || row.SourceRef != "opp-respec-a" {
		t.Fatalf("fixture: A must be working under its own opportunity: %+v", row)
	}
	g := picDefault
	g.stop = 96.5
	picPlan(r, picScenario("P1", "opp-respec-b", r.now, epoch, g)) // v2: B under A's id
	logs := captureTraderLog(t)
	workingBook(r, r.now.Add(5*time.Second), "sig-respec-a", 100.5)
	picPass(r, 5*time.Second, 100.25)
	workingBook(r, r.now.Add(10*time.Second), "sig-respec-a", 100.5)
	picPass(r, 10*time.Second, 100.25)
	if sigs, cancels := r.drain(); len(sigs) != 0 || len(cancels) != 0 {
		t.Fatalf("nothing may be placed or cancelled on B's admission of A's working row: sigs=%+v cancels=%+v", sigs, cancels)
	}
	if after := r.rows(); len(after) != 1 || after[0].State != store.StateWorking || after[0].SourceRef != "opp-respec-a" || after[0].StopPx != 97 {
		t.Fatalf("A's working row must be untouched: %+v", after)
	}
	if n := strings.Count(logs.String(), "NOT authored — arm_source_mismatch"); n != 1 {
		t.Fatalf("the typed refusal must still be named once for a WORKING row, got %d:\n%s", n, logs.String())
	}
	if n := store.ArmRefusalCount(r.st, r.at.id, "2026-09-11", "TEST", "arm_source_mismatch"); n != 1 {
		t.Fatalf("the typed refusal is counted once: %d", n)
	}
}

// The predicate's edges (pure): only a WORKING planner row with a signal is
// ever judged, and a version bump that both moves the zone and re-specs the
// bracket yields ONE decision (E2 first) — at most one cancel per row.
func TestArmRespecForEdges(t *testing.T) {
	zl := zoneLeg{on: true, v: kernel.ZoneVerdict{Lo: 99, Hi: 100, Far: 100, Near: 99}}
	leg := kernel.PlanArmLeg{Entry: 100, Stop: 96, Target: 112}
	base := store.ArmedOrderDB{State: store.StateWorking, SignalID: "sig", Version: 1, Policy: kernel.EntryPolicyMarketInZone,
		EntryPx: 100.5, StopPx: 97.7, TargetPx: 110}
	if c, ok := armRespecFor(2, base, leg, zl, 0.25); !ok || !strings.HasPrefix(c.reason, "zone moved by v2") || c.counter != "market_in_zone:zone_moved" {
		t.Fatalf("zone moved AND bracket re-spec'd → ONE decision, the zone's: %+v %v", c, ok)
	}
	for name, mut := range map[string]func(*store.ArmedOrderDB){
		"armed":          func(r *store.ArmedOrderDB) { r.State = store.StateArmed },
		"place_pending":  func(r *store.ArmedOrderDB) { r.State = store.StatePlacePending },
		"cancel_pending": func(r *store.ArmedOrderDB) { r.State = store.StateCancelPending },
		"no signal":      func(r *store.ArmedOrderDB) { r.SignalID = " " },
		"picture row":    func(r *store.ArmedOrderDB) { r.Source = "picture" },
	} {
		p := base
		mut(&p)
		if c, ok := armRespecFor(2, p, leg, zl, 0.25); ok {
			t.Fatalf("%s: must never be judged, got %+v", name, c)
		}
	}
	in := base
	in.EntryPx = 100
	if c, ok := armRespecFor(2, in, leg, zl, 0.25); !ok || !strings.HasPrefix(c.reason, "bracket re-spec by v2: SL 97.70→96.00 TP 110.00→112.00") {
		t.Fatalf("a contained limit with a re-spec'd bracket under a newer version is E1: %+v %v", c, ok)
	}
	if c, ok := armRespecFor(1, in, leg, zl, 0.25); ok {
		t.Fatalf("the same version never re-specs a bracket (live ATR drift): %+v", c)
	}
	legacy := in
	legacy.Policy = ""
	if c, ok := armRespecFor(2, legacy, leg, zoneLeg{}, 0.25); !ok || c.counter != "arm:respec_cancel" {
		t.Fatalf("a legacy working row re-spec'd by a newer version is E1 too: %+v %v", c, ok)
	}
	small := leg
	small.Stop, small.Target = 97.5, 110.25
	if c, ok := armRespecFor(2, in, small, zl, 0.25); ok {
		t.Fatalf("a < 2-tick move is not a re-spec: %+v", c)
	}
}
