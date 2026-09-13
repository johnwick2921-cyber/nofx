package trader

import "testing"

// reconcileBeforeOpenNT flattens an NT8 orphan before opening, routing on
// `held == "long"`. NT8 GetPositions emits an UPPERCASE side, so before the fix a
// held LONG returned "LONG", missed the == "long" branch, and CloseShort was sent
// — the flatten never confirmed flat and every open was refused. ntHeldPosition
// must normalize to lowercase so the routing is correct.
func TestNtHeldPosition_NormalizesUppercaseNT8SideToLower(t *testing.T) {
	m := &MockTrader{positions: []map[string]interface{}{
		{"symbol": "MNQ", "side": "LONG", "positionAmt": 2.0},
	}}
	at := &AutoTrader{trader: m}

	if got, err := at.ntHeldPosition("MNQ"); err != nil || got != "long" {
		t.Fatalf("ntHeldPosition(MNQ) = %q, want \"long\" (uppercase NT8 side must normalize so reconcile routes to CloseLong)", got)
	}
	// No matching symbol → flat.
	if got, err := at.ntHeldPosition("ES"); err != nil || got != "" {
		t.Fatalf("ntHeldPosition(ES) = %q, want \"\" (flat)", got)
	}
}

func TestNtHeldPositionPreservesSignedShortAndRefusesMissingQuantity(t *testing.T) {
	m := &MockTrader{positions: []map[string]interface{}{{"symbol": "MNQ", "side": "SHORT", "positionAmt": -1.0}}}
	at := &AutoTrader{trader: m}
	if side, err := at.ntHeldPosition("MNQ"); err != nil || side != "short" {
		t.Fatalf("short became flat: %s %v", side, err)
	}
	m.positions = []map[string]interface{}{{"symbol": "MNQ", "side": "LONG"}}
	if _, err := at.ntHeldPosition("MNQ"); err == nil {
		t.Fatal("missing quantity became flat")
	}
}
