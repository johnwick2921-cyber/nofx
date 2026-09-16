package store

import "testing"

// ── 101 D2 — THE CHART SHOWS STORED HISTORY ACROSS CONTRACT ROLLS ───────────
//
// Owner ruling 2026-09-16. A DISPLAY-ONLY reader: prior contracts fill the
// time STRICTLY BEFORE the current contract's first live row, so the series is
// continuous with exactly one visible step at the real roll — never
// back-adjusted, never smoothed (research law). The contract-scoped readers
// the kernel, levels and arm path use (LastNBarsOn, BarsBetweenOn) are not
// touched; decision truth stays current-contract only.

const fiveMin = int64(5 * 60_000)

func rollSeed(t *testing.T, bh *BarHistoryStore) (boundaryMs int64) {
	t.Helper()
	var rows []BarHistoryDB
	base := int64(1_789_000_000_000)
	// 09-26: 3,000 live bars, older
	for i := 0; i < 3000; i++ {
		ts := base + int64(i)*fiveMin
		rows = append(rows, BarHistoryDB{Symbol: "MNQ", TF: "5m", OpenTimeMs: ts, O: 29000, H: 29010, L: 28990, C: 29005, V: 1, Contract: "MNQ 09-26", Source: BarSourceLive})
	}
	// 12-26: sparse daily imports OVERLAPPING the 09-26 window (the 09-07..09-14
	// shape). Imports enter through ImportBars — InsertBars refuses the source.
	var imports []BarHistoryDB
	for i := 0; i < 5; i++ {
		ts := base + int64(2000+i*288)*fiveMin
		imports = append(imports, BarHistoryDB{Symbol: "MNQ", TF: "5m", OpenTimeMs: ts, O: 29290, H: 29300, L: 29280, C: 29295, V: 1, Contract: "MNQ 12-26", Source: BarSourceHistoricalImport})
	}
	// 12-26: 500 live bars starting right after the 09-26 series — the roll
	boundaryMs = base + 3000*fiveMin
	for i := 0; i < 500; i++ {
		ts := boundaryMs + int64(i)*fiveMin
		rows = append(rows, BarHistoryDB{Symbol: "MNQ", TF: "5m", OpenTimeMs: ts, O: 29290, H: 29300, L: 29280, C: 29295, V: 1, Contract: "MNQ 12-26", Source: BarSourceLive})
	}
	if err := bh.InsertBars(rows); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, _, err := bh.ImportBars(imports); err != nil {
		t.Fatalf("seed imports: %v", err)
	}
	return boundaryMs
}

// E3 (store half): 5,000 asked, 500 current-contract rows → the reader fills
// the OLDER 4,500 from 09-26, each row labelled, time-ordered, and the
// overlapping 12-26 imports do NOT interleave into the 09-26 series.
func TestPriorContractsFillBehindTheCurrentContract(t *testing.T) {
	bh := newBarStore(t)
	boundary := rollSeed(t, bh)

	rows, err := bh.PriorContractBarsBefore("MNQ", "5m", boundary, 4500)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3000 {
		t.Fatalf("want all 3000 prior 09-26 rows (only 3000 exist before the boundary), got %d", len(rows))
	}
	for i, r := range rows {
		if r.Contract != "MNQ 09-26" {
			t.Fatalf("row %d is %s — a 12-26 import interleaved into the 09-26 series (the overlap must not mix)", i, r.Contract)
		}
		if r.OpenTimeMs >= boundary {
			t.Fatalf("row %d at %d is not strictly before the boundary %d", i, r.OpenTimeMs, boundary)
		}
		if i > 0 && rows[i].OpenTimeMs <= rows[i-1].OpenTimeMs {
			t.Fatalf("rows must be ascending by time; %d then %d", rows[i-1].OpenTimeMs, rows[i].OpenTimeMs)
		}
	}
	// NEVER back-adjusted: the 09-26 closes are the 09-26 closes
	if rows[len(rows)-1].C != 29005 {
		t.Fatalf("prior-contract price was altered: %.2f (research law: never back-adjust across a roll)", rows[len(rows)-1].C)
	}
}

// E4 (store half): the decision readers are untouched and still contract-pure.
func TestDecisionReadersStayCurrentContractOnly(t *testing.T) {
	bh := newBarStore(t)
	boundary := rollSeed(t, bh)
	rows, err := bh.LastNBarsOn("MNQ", "5m", "MNQ 12-26", 5000)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.Contract != "MNQ 12-26" {
			t.Fatalf("LastNBarsOn leaked a %s row — a decision reader must be contract-pure", r.Contract)
		}
	}
	// 500, not 505: the decision reader already EXCLUDES historical_import
	// rows (its own source filter) — measured here, not assumed. That is the
	// CTO's amendment-2 guard (iii) already in force for planner rings.
	if got := len(rows); got != 500 {
		t.Fatalf("LastNBarsOn(12-26) = %d rows, want 500 (live only; imports excluded) — the current-contract reader must be byte-identical in behaviour", got)
	}
	_ = boundary
}

// FirstLiveOn: the boundary is the current contract's first LIVE row — sparse
// imports before it do not move the roll.
func TestRollBoundaryIsTheFirstLiveRowNotTheFirstImport(t *testing.T) {
	bh := newBarStore(t)
	boundary := rollSeed(t, bh)
	got, ok, err := bh.FirstLiveOn("MNQ", "5m", "MNQ 12-26")
	if err != nil || !ok {
		t.Fatalf("FirstLiveOn: ok=%v err=%v", ok, err)
	}
	if got != boundary {
		t.Fatalf("boundary = %d, want the first LIVE 12-26 row %d, not the first import", got, boundary)
	}
}
