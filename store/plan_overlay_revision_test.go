package store

import (
	"errors"
	"sync"
	"testing"
)

func TestCheckedOverlaySerializesCompetingDraftsAndPlanner(t *testing.T) {
	st := newPlanTestStore(t)
	ps := st.Plan()
	id := "2026-08-14:NY"
	if _, err := ps.AppendPlan(samplePlan(id)); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	outcomes := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := ps.AppendOverlayChecked(&PlanOverlayDB{PlanID: id, PlanVersion: 1, Patch: `[]`}, 0)
			outcomes <- err
		}()
	}
	wg.Wait()
	close(outcomes)
	successes, conflicts := 0, 0
	for err := range outcomes {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrPlanRevisionConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflict=%d", successes, conflicts)
	}
	if _, err := ps.AppendPlan(samplePlan(id)); err != nil {
		t.Fatal(err)
	}
	if _, err := ps.AppendOverlayChecked(&PlanOverlayDB{PlanID: id, PlanVersion: 1, Patch: `[]`}, 1); !errors.Is(err, ErrPlanRevisionConflict) {
		t.Fatalf("planner append not detected: %v", err)
	}
	if err := ps.UpdatePlanLifecycle(id, 2, "no_trade", "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := ps.AppendOverlayChecked(&PlanOverlayDB{PlanID: id, PlanVersion: 2, Patch: `[]`}, 0); !errors.Is(err, ErrPlanRevisionConflict) {
		t.Fatalf("retired plan editable: %v", err)
	}
}
