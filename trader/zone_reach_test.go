package trader

import (
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/store"
	ntTrader "nofx/trader/ninjatrader"
)

// ── WAVE PLANNER LANE B, B1 — the zone arm's reach ──────────────────────────
//
// 09-24 evidence [A], rows from the live DB (DS-106-copy0924.db): armed_orders
// 187 (zone 30814.25–30820, eval 30666.75, ~148 pts away) and 188 (zone
// 30640.25–30645, eval 30732, +87 pts) were PLACED with the market that far
// from their zone, rested at the far edge, and died at the 30-min rest cap;
// 184 (zone 30721.25–30724.25, eval 30748.75, +24.5) expired the same way.
// A market_in_zone arm whose zone is farther than the placement proximity
// bound from the eval price must stay armed-unplaced (no rest clock) and
// place when price comes within the bound; a rest-cap expiry must return the
// row to armed-unplaced (re-placeable), not dismantle it.
//
// The bound is the existing limit proximity: the armed placement band
// (armedPlaceTicks = 100 ticks = 25 pts on MNQ). The behaviour sits behind
// day_plan.zone_place_within_pts (nil → 25 = ON, the wave's L4 ruling;
// 0 = OFF = today, byte-identical).

// fptr is a float64 pointer helper for knob fields.
func fptr(f float64) *float64 { return &f }

// zoneReachRig builds the standard zone rig with zone_place_within_pts set.
func zoneReachRig(t *testing.T, id string, pts float64) *zoneRig {
	t.Helper()
	r := newZoneRig(t, id, zoneDoc(zoneScenario("S1", kernel.EntryPolicyMarketInZone, zone, false)))
	r.at.config.StrategyConfig.DayPlan.ZonePlaceWithinPts = fptr(pts)
	return r
}

// A market_in_zone row whose zone is farther than the placement band from the
// eval price stays armed-unplaced — no wire order, no rest clock — and places
// on a later pass once price comes within the band. Fixture shapes: rows 187
// (+148) and 188 (+87) at 09-24 [A]; today both place and rest far away.
func TestZoneArmBeyondProximityWaitsUnplaced(t *testing.T) {
	r := zoneReachRig(t, "w3-reach-beyond", 25)
	r.armZoneRow() // price short of the zone: the row is authored and armed
	if row := r.row("S1"); row.State != store.StateArmed {
		t.Fatalf("fixture: the row is armed-unplaced: %+v", row)
	}
	// Price far above the zone (beyond by ~59.5 pts > the 25-pt band) — the
	// 187/188 shape. Today this places; the reach contract keeps it armed.
	r.setTape(zoneTape(160.0, r.now, 0))
	r.at.maybeManageArmedOrdersAt(nil, r.now)
	sigs, cancels := r.drain()
	if len(sigs) != 0 || len(cancels) != 0 {
		t.Fatalf("a zone arm far beyond the band must send NOTHING (stays armed-unplaced): sigs=%+v cancels=%+v", sigs, cancels)
	}
	row := r.row("S1")
	if row.State != store.StateArmed || row.SignalID != "" || row.PlacedAtMs != nil {
		t.Fatalf("the far arm stays armed-unplaced with no rest clock: %+v", row)
	}
	if !strings.Contains(row.LastVerdict, "proximity") {
		t.Fatalf("the wait is recorded as a proximity verdict: %+v", row)
	}
	// Price returns within the band of the zone → the SAME row places.
	r.setTape(zoneTape(101.95, r.now.Add(time.Minute), 0))
	r.at.maybeManageArmedOrdersAt(nil, r.now.Add(time.Minute))
	sigs, _ = r.drain()
	if len(sigs) != 1 || sigs[0].LimitPrice != 100.5 {
		t.Fatalf("once price is within the band the arm places at the far edge 100.50: %+v", sigs)
	}
}

// The knob OFF (0) reproduces today byte-identically: a zone arm far beyond
// the band places immediately, no waiting.
func TestZoneReachKnobOffPlacesBeyondTheBand(t *testing.T) {
	r := zoneReachRig(t, "w3-reach-off", 0)
	r.armZoneRow()
	r.setTape(zoneTape(160.0, r.now, 0))
	r.at.maybeManageArmedOrdersAt(nil, r.now)
	sigs, _ := r.drain()
	if len(sigs) != 1 || sigs[0].LimitPrice != 100.5 {
		t.Fatalf("zone_place_within_pts=0 is the legacy behaviour: the far arm places at once: %+v", sigs)
	}
}

// A resting market_in_zone limit past the rest cap is returned to
// armed-unplaced (re-placeable) instead of dismantled: the wire cancel is
// sent exactly as today, the row resets (state=armed, signal/eval/placed_at
// cleared, placement_seq+1), and it re-places once the broker book clears
// and price is within the band. Fixture shape: row 184 (+24.5, placed and
// killed by the rest cap on 09-24) [A].
func TestZoneRestExpiryReturnsTheRowToArmedUnplaced(t *testing.T) {
	r := zoneReachRig(t, "w3-reach-restexpiry", 25)
	first := r.placeWorking(100.5)
	sid := first.SignalID
	t31 := r.now.Add(31 * time.Minute)
	r.restingBook(t31, sid, 100.5) // the broker still lists the resting limit
	r.setTape(zoneTape(101.95, t31, 0))
	r.at.maybeManageArmedOrdersAt(nil, t31)
	sigs, cancels := r.drain()
	if len(sigs) != 0 {
		t.Fatalf("an expiry pass places nothing: %+v", sigs)
	}
	if len(cancels) != 1 || cancels[0].SignalID != sid {
		t.Fatalf("the wire cancel is still sent for the resting order: %+v", cancels)
	}
	row := r.row("S1")
	if row.State != store.StateArmed {
		t.Fatalf("the expiry returns the row to armed-unplaced, got %q (%q): %+v", row.State, row.StateReason, row)
	}
	if row.SignalID != "" || row.PlacedAtMs != nil || row.EvalPrice != nil || row.EvalBarMs != nil {
		t.Fatalf("the reset clears the placement stamp (no rest clock): %+v", row)
	}
	if row.PlacementSeq != 1 {
		t.Fatalf("the next broker placement gets a fresh seq (0 authored, +1 on reset), got %d: %+v", row.PlacementSeq, row)
	}
	if !strings.Contains(row.StateReason, "zone rest expired") {
		t.Fatalf("the reason names the rest expiry: %+v", row)
	}
	// Re-placeable: the broker book confirms the old order is gone and price
	// is within the band → the same row places again under a new signal.
	// The adapter's B3 duplicate guard runs on the WALL clock while the
	// fixture clock says minutes passed — the adapter is re-made to stand for
	// that (the settleAndReArm pattern).
	t32 := t31.Add(time.Minute)
	r.flatBook(t32)
	r.at.trader = ntTrader.NewTCPTrader(r.srv, "MNQ", "Sim101")
	r.setTape(zoneTape(101.95, t32, 0))
	r.at.maybeManageArmedOrdersAt(nil, t32)
	sigs, _ = r.drain()
	if len(sigs) != 1 || sigs[0].SignalID == sid {
		t.Fatalf("after the book clears the reset row re-places under a NEW signal: %+v", sigs)
	}
	if sigs[0].LimitPrice != 100.5 {
		t.Fatalf("the re-place sits at the far edge 100.50: %+v", sigs)
	}
}

// The knob OFF (0) keeps today's rest-cap semantics byte-identically: the
// expiry dismantles the row (cancel_pending with reason "zone rest expired").
func TestZoneRestExpiryKnobOffKeepsToday(t *testing.T) {
	r := zoneReachRig(t, "w3-reach-restexpiry-off", 0)
	first := r.placeWorking(100.5)
	sid := first.SignalID
	t31 := r.now.Add(31 * time.Minute)
	r.restingBook(t31, sid, 100.5)
	r.setTape(zoneTape(101.95, t31, 0))
	r.at.maybeManageArmedOrdersAt(nil, t31)
	_, cancels := r.drain()
	if len(cancels) != 1 {
		t.Fatalf("knob OFF still sends the wire cancel: %+v", cancels)
	}
	row := r.row("S1")
	if row.State != store.StateCancelPending || row.StateReason != "zone rest expired" {
		t.Fatalf("zone_place_within_pts=0 keeps the legacy dismantle: %+v", row)
	}
}
