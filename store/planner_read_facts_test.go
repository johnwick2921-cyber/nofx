package store

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// VOID PARITY D3 — a read's facts are recorded whether or not the read failed.
// This is the whole point: before it, a working fix erased its own evidence.
func TestReadFactsPersistOnEveryReadAndCap(t *testing.T) {
	st, err := New(filepath.Join(t.TempDir(), "rf.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	defer st.Close()
	rf := st.PlannerReadFacts()

	// An ACCEPTED read (no reject row anywhere) still records its facts.
	if err := rf.SaveReadFact(&PlannerReadFact{
		TraderID: "hoang", TradeDate: "2026-09-02", Session: "ASIA", PromptHash: "h1",
		VoidLevels: EncodeVoidLevels([]VoidLevelRecord{{Price: 29141.25, Short: true, ReclaimedAt: "03:34 CT"}}),
		VoidCount:  1, StopFloorPts: 27.1, ATR5m: 18.09, StopFloorMlt: 1.5,
		ScopeSinceMs: 0, ScopeBars: 2000, ScopeIntv: "1m",
	}); err != nil {
		t.Fatalf("accepted-read write: %v", err)
	}
	got, err := rf.LatestReadFact()
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if got.VoidCount != 1 || got.StopFloorPts != 27.1 || got.ScopeBars != 2000 || got.ScopeSinceMs != 0 {
		t.Errorf("row lost its facts: %+v", got)
	}
	var recs []VoidLevelRecord
	if err := json.Unmarshal([]byte(got.VoidLevels), &recs); err != nil || len(recs) != 1 || recs[0].Price != 29141.25 {
		t.Errorf("void list must round-trip verbatim, got %q (%v)", got.VoidLevels, err)
	}

	// "computed and EMPTY" must not read as "not computed" (A24: no placeholder
	// that reads as data).
	if EncodeVoidLevels(nil) != "[]" {
		t.Errorf("an empty computed list encodes as [], got %q", EncodeVoidLevels(nil))
	}

	// Cap trims oldest, newest survive.
	for i := 0; i < PlannerReadFactsCap+25; i++ {
		if err := rf.SaveReadFact(&PlannerReadFact{TraderID: "hoang", PromptHash: "bulk", VoidLevels: "[]"}); err != nil {
			t.Fatalf("bulk write %d: %v", i, err)
		}
	}
	if n := rf.ReadFactsCount(); n > PlannerReadFactsCap {
		t.Errorf("cap not enforced: %d rows > %d", n, PlannerReadFactsCap)
	}
	last, err := rf.LatestReadFact()
	if err != nil || last.PromptHash != "bulk" {
		t.Errorf("newest row must survive the trim: %+v %v", last, err)
	}
	t.Logf("rows after cap: %d (cap %d)", rf.ReadFactsCount(), PlannerReadFactsCap)
}

// A nil store must be a no-op, never a panic on a planner read (A10).
func TestReadFactsNilStoreIsSafe(t *testing.T) {
	var rf *PlannerReadFactsStore
	if err := rf.SaveReadFact(&PlannerReadFact{}); err != nil {
		t.Errorf("nil store must no-op, got %v", err)
	}
	if n := rf.ReadFactsCount(); n != 0 {
		t.Errorf("nil store count must be 0, got %d", n)
	}
}
