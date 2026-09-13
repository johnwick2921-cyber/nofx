package ninjatrader

import (
	"nofx/store"
	"testing"
	"time"
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

func TestRejectedEntryFillPreservesExecutedSliceAndUnknownEvidence(t *testing.T) {
	for _, tc := range []struct {
		name              string
		qty               int
		price             float64
		held, wantPending bool
	}{{"positive", 1, 100, true, false}, {"unknown positive basis", 1, 0, true, true}, {"zero", 0, 0, false, false}, {"known positive then zero", 0, 0, true, false}, {"older positive", 1, 100, true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			tr, _, _ := partialCloseFixture(t)
			if tc.name == "positive" || tc.name == "zero" {
				tr.hasFill = false
				tr.lastFill.Quantity = 0
			}
			tr.pending = map[string]string{"entry-owned": "long"}
			tr.pendingAt = map[string]int64{"entry-owned": time.Now().UnixMilli()}
			fill := tr.lastFill
			fill.Status = "rejected"
			fill.Quantity = tc.qty
			fill.FillPrice = tc.price
			tr.handleFill(fill)
			if tr.hasFill != tc.held {
				t.Fatalf("held=%v want=%v", tr.hasFill, tc.held)
			}
			_, pending := tr.pending[fill.SignalID]
			if pending != tc.wantPending {
				t.Fatalf("pending=%v want=%v", pending, tc.wantPending)
			}
			if tc.name == "positive" && (tr.lastFill.Quantity != 1 || tr.lastFill.Status != "partial") {
				t.Fatalf("actual partial not cached: %+v", tr.lastFill)
			}
		})
	}
}

func TestLegacyGuardRejectionRequestedQuantityStillSettlesRefusal(t *testing.T) {
	tr, _, _ := partialCloseFixture(t)
	tr.pending = map[string]string{"entry-owned": "long"}
	tr.pendingAt = map[string]int64{}
	notified := false
	tr.SetRejectSink(func(signal, reason string) { notified = signal == "entry-owned" })
	fill := tr.lastFill
	fill.Account = ""
	fill.Status = "rejected"
	fill.Quantity = 1
	fill.FillPrice = 0
	tr.handleFill(fill)
	if !notified || len(tr.pending) != 0 {
		t.Fatal("pre-submit refusal receipt lost")
	}
	if !tr.hasFill || !tr.entryReceivedAt.IsZero() {
		t.Fatal("requested size changed existing exposure or execution fence")
	}
}

func TestResidualCacheNormalizationPreservesRawFillRing(t *testing.T) {
	tr, st, row := partialCloseFixture(t)
	tr.st = st
	if err := st.GormDB().Model(row).Updates(map[string]any{"quantity": 2, "entry_price": 95}).Error; err != nil {
		t.Fatal(err)
	}
	tr.handleFill(tr.lastFill)
	if tr.lastFill.Quantity != 2 || tr.lastFill.FillPrice != 95 {
		t.Fatalf("cache not residual: %+v", tr.lastFill)
	}
	if len(tr.recentFills) != 1 {
		t.Fatal("execution evidence missing")
	}
	if tr.recentFills[0].Quantity != 3 || tr.recentFills[0].Price != 100 {
		t.Fatalf("raw execution evidence rewritten: %+v", tr.recentFills[0])
	}
}
