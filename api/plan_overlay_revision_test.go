package api

import (
	"nofx/kernel"
	"nofx/store"
	"testing"
	"time"
)

func TestOverlayEditRevisionRefusesStaleDraft(t *testing.T) {
	s, st := askTestServer(t)
	now := time.Date(2026, 9, 14, 10, 0, 0, 0, kernel.CTLocation())
	id := store.MakePlanIDForTrader("t", "2026-09-14", "NY")
	doc := `{"reasoning":"review","bias":{"direction":"neutral","conviction":"low","flip_condition":"n/a"},"levels":[],"scenarios":[{"id":"S1","condition":"reject","direction":"long","quality":"A"}],"no_trade":[],"death_condition":"dead"}`
	seed := func() {
		t.Helper()
		if _, err := st.Plan().AppendPlan(&store.PlanDB{PlanID: id, StrategyID: "t", TradeDate: "2026-09-14", Session: "NY", Doc: doc}); err != nil {
			t.Fatal(err)
		}
	}
	seed()
	zero := 0
	expected := &planOverlayRevision{PlanID: id, PlanVersion: 1, OverlayVersion: &zero}
	_, _, code, msg := s.applyPlanOverlay("t", "MNQ", `[{"op":"replace","path":"/reasoning","value":"owner edit"}]`, "owner", now, expected)
	if code != 0 {
		t.Fatalf("current draft refused: %d %s", code, msg)
	}
	_, _, code, _ = s.applyPlanOverlay("t", "MNQ", `[]`, "owner", now, expected)
	if code != 409 {
		t.Fatalf("stale overlay revision accepted: %d", code)
	}
	seed()
	one := 1
	expected.OverlayVersion = &one
	_, _, code, _ = s.applyPlanOverlay("t", "MNQ", `[]`, "owner", now, expected)
	if code != 409 {
		t.Fatalf("stale plan version accepted: %d", code)
	}
	rows, err := st.Plan().ListOverlays(id, 2)
	if err != nil || len(rows) != 0 {
		t.Fatalf("new plan mutated: %v %v", rows, err)
	}
}
