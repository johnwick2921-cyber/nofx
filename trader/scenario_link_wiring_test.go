// W1 — the link must actually reach the row. A29: built is not wired.
//
// scenario_anchor.go and store.ResolveScenarioLink can be perfect and still be
// worth nothing if the recorder never calls them. These pins fail if the wiring
// is removed, which is the only reason they exist.

package trader

import (
	"testing"
	"time"

	"nofx/store"
)

// The happy path: one anchor inside the band lands on the row WITH its basis
// and both distances.
func TestRecordedRowCarriesTheScenarioLink(t *testing.T) {
	const level = 29141.25
	at, st, now := recorderFixture(t, level, 2000, time.Now())
	formed := now.Add(-6 * time.Hour).UnixMilli()

	at.recordDetectorOutputs("MNQ", "P1", "NY", 1,
		nil, seatedAt(level, 0, formed), level, 10, 2.0, 12, now,
		[]store.ScenarioAnchor{{ID: "S1", Price: level + 1.0}})

	rows, err := st.TouchOutcomes().AllOutcomes()
	if err != nil || len(rows) == 0 {
		t.Fatalf("no rows written: %v", err)
	}
	r := rows[len(rows)-1]
	if r.ScenarioNearest == nil || *r.ScenarioNearest != "S1" {
		t.Fatalf("the row must carry the nearest scenario, got %v", r.ScenarioNearest)
	}
	if r.ScenarioLinkBasis != store.ScenarioLinkPriceProximity {
		t.Errorf("basis = %q, want %q", r.ScenarioLinkBasis, store.ScenarioLinkPriceProximity)
	}
	if r.ScenarioLinkDistPts == nil || *r.ScenarioLinkDistPts != 1.0 {
		t.Errorf("distance in points = %v, want 1", r.ScenarioLinkDistPts)
	}
}

// No anchors is NOT a blank row: the basis says WHY the link is NULL, so a
// reader can tell "nothing authored" from "nothing close".
func TestRecordedRowStatesWhyTheLinkIsNull(t *testing.T) {
	const level = 29141.25
	at, st, now := recorderFixture(t, level, 2000, time.Now())
	formed := now.Add(-6 * time.Hour).UnixMilli()

	at.recordDetectorOutputs("MNQ", "P1", "NY", 1,
		nil, seatedAt(level, 0, formed), level, 10, 2.0, 12, now, nil)

	rows, _ := st.TouchOutcomes().AllOutcomes()
	if len(rows) == 0 {
		t.Fatal("no rows written")
	}
	r := rows[len(rows)-1]
	if r.ScenarioNearest != nil {
		t.Fatalf("with no anchors the link must be NULL, got %q", *r.ScenarioNearest)
	}
	if r.ScenarioLinkBasis != store.ScenarioLinkNoScenario {
		t.Errorf("basis = %q, want %q — a bare NULL loses the reason", r.ScenarioLinkBasis, store.ScenarioLinkNoScenario)
	}
}

// AMBIGUITY REACHES THE ROW AS NULL. Two anchors inside the map's own cluster
// width means the record cannot say which scenario this touch belongs to.
func TestAmbiguousLinkReachesTheRowAsNull(t *testing.T) {
	const level = 29141.25
	at, st, now := recorderFixture(t, level, 2000, time.Now())
	formed := now.Add(-6 * time.Hour).UnixMilli()

	at.recordDetectorOutputs("MNQ", "P1", "NY", 1,
		nil, seatedAt(level, 0, formed), level, 10, 2.0, 12, now,
		[]store.ScenarioAnchor{{ID: "S1", Price: level + 1.0}, {ID: "S2", Price: level - 1.0}})

	rows, _ := st.TouchOutcomes().AllOutcomes()
	r := rows[len(rows)-1]
	if r.ScenarioNearest != nil {
		t.Fatalf("two anchors within the band must resolve NULL, got %q", *r.ScenarioNearest)
	}
	if r.ScenarioLinkBasis != store.ScenarioLinkAmbiguous {
		t.Errorf("basis = %q, want %q", r.ScenarioLinkBasis, store.ScenarioLinkAmbiguous)
	}
}
