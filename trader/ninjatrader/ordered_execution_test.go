package ninjatrader

import (
	"nofx/store"
	"testing"
)

func TestOrderedFillFanoutCannotRestoreClosedCache(t *testing.T) {
	tr, st, _ := partialCloseFixture(t)
	tr.st = st
	fill := tr.lastFill
	tr.handleFill(fill)
	exit := closeFixtureFrame()
	exit.Quantity = 3
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), exit)
	if tr.hasFill {
		t.Fatal("ordered close did not clear cache")
	}
	// The router may deliver its copy later; the actual fill consumer skips it.
	fill.OrderedHandled = true
	tr.handleFill(fill)
	if tr.hasFill {
		t.Fatal("advisory replay restored closed cache")
	}
	exit.OrderedHandled = true
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), exit)
	if tr.hasFill {
		t.Fatal("advisory exit changed terminal cache")
	}
	// A full wire replay is not marked as already handled. It must still not
	// resurrect the exact closed entry; duplicate close deliberately has no hooks.
	fill.OrderedHandled = false
	tr.handleFill(fill)
	exit.OrderedHandled = false
	tr.recordClose("owned", "ex", "ninjatrader", st, store.NewPositionBuilder(st.Position()), exit)
	if tr.hasFill {
		t.Fatal("wire replay restored fully exited entry cache")
	}
	// The receipt fence still requires authoritative account truth.
	if _, err := tr.GetPositions(); err == nil {
		t.Fatal("cache closure fabricated account flat")
	}
}
