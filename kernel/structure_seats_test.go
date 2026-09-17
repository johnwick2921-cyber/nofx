package kernel

import (
	"reflect"
	"testing"
)

// S3 (2026-09-16) — day_plan.htf_seats drives seatHTF's promotion count:
// 0 = no HTF seating, 2 = the legacy default, higher = more HTF seats.

func htfsSeatFixture() []ScoredLevel {
	// PRE-SORTED like a real scorer output (score desc), all scores EQUAL: at the
	// cut line the stable final sort keeps the promoted HTF in the head and the
	// demoted round number in the tail — promotion is the only observable change.
	scored := make([]ScoredLevel, 0, 13)
	for i := 0; i < 8; i++ { // round numbers fill the head
		scored = append(scored, ScoredLevel{
			DetectedLevel: DetectedLevel{Kind: KindRound, Price: 990 + float64(i), Label: "RN"},
			Score:         0.5, Grade: "C", Fresh: "fresh", Distance: 990 + float64(i) - 1000,
		})
	}
	for i := 0; i < 5; i++ { // HTF reversal zones lost the cut
		scored = append(scored, ScoredLevel{
			DetectedLevel: DetectedLevel{Kind: KindSupply, Price: 1100 + float64(i), Lo: 1098 + float64(i), Hi: 1102 + float64(i), Label: "Supply·4h", HTF: true, TF: "4h", ZonePattern: "reversal"},
			Score:         0.5, Grade: "C", Fresh: "fresh", Distance: 100 + float64(i),
		})
	}
	return scored
}

func htfCountInHead(out []ScoredLevel, maxLevels int) int {
	n := 0
	for _, l := range out[:maxLevels] {
		if isHTFSeatEligible(l) {
			n++
		}
	}
	return n
}

// the dispatch's seat test: htf_seats=4 seats four; 2 seats two; 0 seats none;
// 6 seats all five candidates available.
//
// STOP-LINE (reported to the CTO 2026-09-16): seatHTF's final "restore strict
// seating order" sort uses the SAME comparator as the pre-seat sort, so a
// promoted tail candidate (which loses that comparator to every head member)
// is restored to the tail by construction — the promotion path is nullified
// and no seats value can change the table. The pre-existing
// TestSeatHTFPromotesSwingLevels passes vacuously (its HTF candidates outscore
// the head fillers, so the final sort alone seats them). This test is SKIPPED
// until the CTO rules: make the promotion survive the sort, or ship the knob
// on the no-op mechanism.
func TestSeatHTFSeatsKnob(t *testing.T) {
	t.Skip("STOP-line: seatHTF promotion is nullified by its own final sort — pending CTO ruling (D102-1/S3 report)")
	for seats, want := range map[int]int{4: 4, 2: 2, 0: 0, 6: 5} {
		out := seatHTF(htfsSeatFixture(), 8, seats)
		if got := htfCountInHead(out, 8); got != want {
			t.Errorf("seats=%d: %d HTF in head, want %d", seats, got, want)
		}
	}
}

// seats ≤ 0 must return the input untouched (a legal knob value, not an error).
func TestSeatHTFZeroSeatsNoOp(t *testing.T) {
	in := htfsSeatFixture()
	out := seatHTF(in, 8, 0)
	if !reflect.DeepEqual(out, in) {
		t.Fatal("seats=0 must be a no-op — the table must not move")
	}
}

// parity pins (canon 53): (1) the legacy wrapper and the S3 seats path with the
// legacy count produce IDENTICAL tables; (2) knob-off byte-identity against the
// pre-S3 binary is pinned by the existing goldens that run the production
// Assemble path at LegacyHtfSeats (identity_output_parity_test,
// one_setup_map_pin_test, weekly_shadow_test).
func TestS3HtfSeatsParityWithLegacy(t *testing.T) {
	levels, price, dATR := htfsParityLevels()
	viaWrapper, _ := ScoreLevelsMinGradeFull(levels, price, dATR, nil, 8, 1.5, "")
	viaSeats, _ := ScoreLevelsMinGradeFullSeats(levels, price, dATR, nil, 8, 1.5, "", LegacyHtfSeats)
	if !reflect.DeepEqual(viaWrapper, viaSeats) {
		t.Fatalf("ScoreLevelsMinGradeFull and …FullSeats(LegacyHtfSeats) must be byte-identical — wrapper=%d rows, seats=%d rows", len(viaWrapper), len(viaSeats))
	}
}

// htfsParityLevels builds a small mixed pool where both seat passes run.
func htfsParityLevels() ([]DetectedLevel, float64, float64) {
	price, dATR := 30000.0, 300.0
	out := make([]DetectedLevel, 0, 14)
	for i := 0; i < 10; i++ {
		out = append(out, DetectedLevel{Kind: KindRound, Price: price - 100 + float64(i)*10, Label: "RN"})
	}
	out = append(out,
		DetectedLevel{Kind: KindSupply, Price: price + 80, Lo: price + 75, Hi: price + 85, Label: "Supply·1h", HTF: true, TF: "1h"},
		DetectedLevel{Kind: KindDemand, Price: price - 80, Lo: price - 85, Hi: price - 75, Label: "Demand·4h", HTF: true, TF: "4h"},
		DetectedLevel{Kind: KindPDH, Price: price + 120, Label: "PDH"},
		DetectedLevel{Kind: KindPDL, Price: price - 120, Label: "PDL"},
	)
	return out, price, dATR
}
