// THE REAPER READS THE BROKER, NOT SILENCE.
//
// `reconcileStaleWorking` cancelled any working row that had seen no
// order_update for the stale window. order_update is an EVENT frame: a resting
// limit that nobody touches emits nothing, so a perfectly healthy order looks
// identical to a dead one after N minutes. The reaper was reading SILENCE AS
// DEATH — and then acting on it, by cancelling at the broker.
//
// F12 put the broker's own book on the wire. The question "is this order still
// alive?" now has an answer that does not depend on anything having happened.

package trader

import (
	"strings"
	"testing"
	"time"

	nt "nofx/provider/ninjatrader"
	"nofx/store"
)

func snapWithOrders(orders ...nt.NT8Order) nt.OrderSnapshotPayload {
	for i := range orders {
		if orders[i].Symbol == "" {
			orders[i].Symbol = "MNQ"
		}
	}
	return nt.OrderSnapshotPayload{Account: "Sim101", BuildID: "2026-09-03-f12", Orders: orders}
}

// THE PIN THE DISPATCH NAMES. A resting limit with 20 minutes of order_update
// silence, listed in a FRESH snapshot, must not be reaped.
func TestRestingLimitWith20mSilenceButPresentInFreshSnapshotIsNotReaped(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	c := nt.NewOrderSnapshotCache()
	c.PutAt(snapWithOrders(nt.NT8Order{
		OrderID: "NT-1", Name: "sig-S1", Type: "limit", LimitPrice: 29450,
		Quantity: 1, State: "Working",
	}), now.Add(-20*time.Second))

	row := store.ArmedOrderDB{
		TraderID: "t1", State: "working", Scenario: "S1", SignalID: "sig-S1",
		UpdatedAt: now.Add(-20 * time.Minute), // twenty minutes of silence
	}

	v, why := reaperVerdictAt(c, "Sim101", "MNQ", row, 30*time.Second, now)
	if v != reaperAlive {
		t.Fatalf("a fresh snapshot listing the order means ALIVE, got %v (%s)", v, why)
	}
	if !strings.Contains(why, "broker") {
		t.Errorf("the reason must say the broker answered: %q", why)
	}
}

// Absent from a fresh snapshot: the broker's word is that it is gone. Reconcile
// from that, do not re-cancel something the broker no longer has.
func TestAbsentFromAFreshSnapshotReconcilesFromTheBroker(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	c := nt.NewOrderSnapshotCache()
	c.PutAt(snapWithOrders(), now.Add(-10*time.Second)) // empty book, fresh

	row := store.ArmedOrderDB{TraderID: "t1", State: "working", Scenario: "S1", SignalID: "sig-S1",
		UpdatedAt: now.Add(-1 * time.Minute)}

	v, why := reaperVerdictAt(c, "Sim101", "MNQ", row, 30*time.Second, now)
	if v != reaperGone {
		t.Fatalf("absent from a fresh book = GONE, got %v (%s)", v, why)
	}
	if !strings.Contains(strings.ToLower(why), "broker") {
		t.Errorf("the reason must attribute it to the broker: %q", why)
	}
}

// No snapshot at all: we know nothing. WARN, never cancel — cancelling on
// ignorance is how the old reaper killed live orders.
func TestNoSnapshotWarnsAndNeverCancels(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	c := nt.NewOrderSnapshotCache()

	row := store.ArmedOrderDB{TraderID: "t1", State: "working", Scenario: "S1", SignalID: "sig-S1",
		UpdatedAt: now.Add(-60 * time.Minute)}

	v, why := reaperVerdictAt(c, "Sim101", "MNQ", row, 30*time.Second, now)
	if v != reaperUnknown {
		t.Fatalf("no snapshot = UNKNOWN, got %v (%s)", v, why)
	}
	if !strings.Contains(strings.ToLower(why), "link stale") {
		t.Errorf("the warn must say link stale: %q", why)
	}
}

// A STALE snapshot is not evidence either. An old book that happens to list the
// order does not prove it is alive, and one that omits it does not prove it is
// dead.
func TestStaleSnapshotIsUnknownWhicheverWayItReads(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	row := store.ArmedOrderDB{TraderID: "t1", State: "working", Scenario: "S1", SignalID: "sig-S1",
		UpdatedAt: now.Add(-60 * time.Minute)}

	listing := nt.NewOrderSnapshotCache()
	listing.PutAt(snapWithOrders(nt.NT8Order{OrderID: "NT-1", Name: "sig-S1", State: "Working", Type: "limit"}),
		now.Add(-5*time.Minute)) // > 2 × 30s
	if v, why := reaperVerdictAt(listing, "Sim101", "MNQ", row, 30*time.Second, now); v != reaperUnknown {
		t.Errorf("a stale book LISTING it is still unknown, got %v (%s)", v, why)
	}

	omitting := nt.NewOrderSnapshotCache()
	omitting.PutAt(snapWithOrders(), now.Add(-5*time.Minute))
	if v, why := reaperVerdictAt(omitting, "Sim101", "MNQ", row, 30*time.Second, now); v != reaperUnknown {
		t.Errorf("a stale book OMITTING it is still unknown, got %v (%s)", v, why)
	}
}

// Matching is by signal id OR order id — the ledger keys on signal id and the
// broker may report either, and a match that only worked one way would read a
// live order as gone.
func TestOrderMatchesOnEitherIdentifier(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	row := store.ArmedOrderDB{TraderID: "t1", State: "working", Scenario: "S1", SignalID: "sig-S1",
		UpdatedAt: now.Add(-30 * time.Minute)}

	byName := nt.NewOrderSnapshotCache()
	byName.PutAt(snapWithOrders(nt.NT8Order{OrderID: "NT-9", Name: "sig-S1", State: "Working"}), now)
	if v, _ := reaperVerdictAt(byName, "Sim101", "MNQ", row, 30*time.Second, now); v != reaperAlive {
		t.Errorf("a name match must count as alive")
	}

	// bracket legs carry the -sl/-tp suffix on the same signal
	byLeg := nt.NewOrderSnapshotCache()
	byLeg.PutAt(snapWithOrders(nt.NT8Order{OrderID: "NT-9", Name: "sig-S1-sl", State: "Accepted"}), now)
	if v, _ := reaperVerdictAt(byLeg, "Sim101", "MNQ", row, 30*time.Second, now); v != reaperAlive {
		t.Errorf("a bracket leg of the same signal must count as alive")
	}
}

// A TERMINAL order in the book is not alive. "Present" must mean present AND
// working, or a filled row would keep a ledger entry open forever.
func TestTerminalOrderInTheBookIsNotAlive(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	c := nt.NewOrderSnapshotCache()
	c.PutAt(snapWithOrders(nt.NT8Order{OrderID: "NT-1", Name: "sig-S1", State: "Filled"}), now)

	row := store.ArmedOrderDB{TraderID: "t1", State: "working", Scenario: "S1", SignalID: "sig-S1",
		UpdatedAt: now.Add(-30 * time.Minute)}
	if v, _ := reaperVerdictAt(c, "Sim101", "MNQ", row, 30*time.Second, now); v != reaperGone {
		t.Errorf("a Filled order in the book is GONE, not alive")
	}
}
