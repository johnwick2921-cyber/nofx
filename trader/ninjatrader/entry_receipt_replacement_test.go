package ninjatrader

import (
	nt "nofx/provider/ninjatrader"
	"testing"
)

func TestNewTCPTraderReplacementRetainsEntrySnapshotEvidence(t *testing.T) {
	for _, fresh := range []bool{false, true} {
		name := "pre_entry_snapshot"
		if fresh {
			name = "post_entry_snapshot_and_duplicate"
		}
		t.Run(name, func(t *testing.T) {
			srv := nt.NewTCPServer(nil)
			srv.SeedPositionsForTest("Sim101", []nt.OpenPosition{})
			old := NewTCPTrader(srv, "MNQ", "Sim101")
			fill := nt.FillPayload{SignalID: "entry-replacement", Symbol: "MNQ", Account: "Sim101", Side: "long", Quantity: 1, FillPrice: 100, Status: "partial"}
			old.handleFill(fill)
			if _, err := old.GetPositions(); err == nil {
				t.Fatal("old empty snapshot accepted after execution")
			}
			if fresh {
				srv.SeedPositionsForTest("Sim101", []nt.OpenPosition{})
			}
			replacement := NewTCPTrader(srv, "MNQ", "Sim101")
			if fresh {
				replacement.handleFill(fill)
			}
			ps, err := replacement.GetPositions()
			if fresh {
				if err != nil || len(ps) != 0 {
					t.Fatalf("newer broker snapshot lost on duplicate after replacement: %+v %v", ps, err)
				}
			} else if err == nil && len(ps) == 0 {
				t.Fatal("replacement forgot entry receipt and invented flat from pre-entry snapshot")
			}
			other := NewTCPTrader(srv, "MNQ", "SimOther")
			srv.SeedPositionsForTest("SimOther", []nt.OpenPosition{})
			if ps, err := other.GetPositions(); err != nil || len(ps) != 0 {
				t.Fatalf("foreign account contaminated: %+v %v", ps, err)
			}
		})
	}
}
