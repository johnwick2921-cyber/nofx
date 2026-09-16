package kernel

import (
	"testing"
	"time"

	"nofx/market"
)

// S2 — timeframe-aware freshness tests (2026-09-16). Parity rule (canon 53):
// the equivalence test below runs through scoreLevelsPool — the production
// call site — not a re-built input.

func mkKline(openMs int64, h, l float64) market.Kline {
	return market.Kline{OpenTime: openMs, High: h, Low: l}
}

func tfFixture() []market.Kline {
	base := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC).UnixMilli()
	step := int64(4 * time.Hour / time.Millisecond)
	return []market.Kline{
		mkKline(base, 100, 98),            // 0: misses the band
		mkKline(base+step, 105, 102),      // 1: trades into [102,106]? see test
		mkKline(base+2*step, 103, 101),    // 2
		mkKline(base+3*step, 106, 104),    // 3
	}
}

func TestLevelFreshnessByTF_CountsAndOrigin(t *testing.T) {
	bars := tfFixture()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	// Zone [102,106] on 4h: bar0 (100/98) misses, bars1-3 hit → 3 tests = stale.
	l := DetectedLevel{Kind: KindSupply, Lo: 102, Hi: 106, TF: "4h", HTF: true, OriginDate: "2026-09-10"}
	grade, n := LevelFreshnessByTF(l, now, bars)
	if grade != "stale" || n != 3 {
		t.Fatalf("got %q/%d, want stale/3", grade, n)
	}
	// Origin after all bars → 0 tests.
	l2 := DetectedLevel{Kind: KindSupply, Lo: 102, Hi: 106, TF: "4h", HTF: true, OriginDate: "2026-09-13"}
	if grade, n := LevelFreshnessByTF(l2, now, bars); grade != "fresh" || n != 0 {
		t.Fatalf("future-origin got %q/%d, want fresh/0", grade, n)
	}
	// Two-touch zone [104.5,105.5]: bar1 (105/102) and bar3 (106/104) trade in.
	l3 := DetectedLevel{Kind: KindSupply, Lo: 104.5, Hi: 105.5, TF: "4h", HTF: true, OriginDate: "2026-09-10"}
	if grade, n := LevelFreshnessByTF(l3, now, bars); grade != "tested-2" || n != 2 {
		t.Fatalf("got %q/%d, want tested-2/2", grade, n)
	}
}

func TestLevelFreshnessByTF_NoOriginMeansFresh(t *testing.T) {
	bars := tfFixture()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	l := DetectedLevel{Kind: KindSupply, Lo: 100, Hi: 106, TF: "4h", HTF: true} // no origin
	grade, n := LevelFreshnessByTF(l, now, bars)
	if grade != "fresh" || n != 0 {
		t.Fatalf("no-origin got %q/%d, want fresh/0", grade, n)
	}
}

func TestNormalizeByTFGrade_IdentityOnLegacy(t *testing.T) {
	// Every pre-existing freshness string must pass through untouched (knob OFF
	// byte-identity). freshMult/zoneFreshMult tables are NOT changed.
	legacy := []string{"", "a", "b", "c", "tested", "done", "consumed", "fresh", "A", "B"}
	for _, f := range legacy {
		if got := normalizeByTFGrade(f); got != f {
			t.Fatalf("normalize(%q) = %q, want identity", f, got)
		}
	}
	if normalizeByTFGrade("tested-1") != "b" || normalizeByTFGrade("tested-2") != "c" || normalizeByTFGrade("stale") != "done" {
		t.Fatal("S2 vocabulary mapping wrong: tested-1→b, tested-2→c, stale→done")
	}
}

// TestScoreLevels_ByTFVocabScoresLikeCanonical — production call site: the same
// level pool graded through the S2 vocabulary must score byte-identically to
// the canonical letters the unchanged tables know (tested-1≡b, tested-2≡c,
// stale≡done). Only Research.Freshness (the display string) differs.
func TestScoreLevels_ByTFVocabScoresLikeCanonical(t *testing.T) {
	levels := []DetectedLevel{
		{Kind: KindPDH, Price: 100, Lo: 100, Hi: 100, Label: "PDH", TF: "1d", HTF: true},
		{Kind: KindSupply, Price: 104, Lo: 102, Hi: 106, Label: "Supply·4h", TF: "4h", HTF: true},
	}
	pairs := []struct{ vocab, canon string }{
		{"tested-1", "b"}, {"tested-2", "c"}, {"stale", "done"}, {"fresh", "a"},
	}
	price, dATR := 101.0, 10.0
	for _, p := range pairs {
		withVocab := scoreLevelsPool(levels, price, dATR, func(DetectedLevel) string { return p.vocab }, 8, 0)
		withCanon := scoreLevelsPool(levels, price, dATR, func(DetectedLevel) string { return p.canon }, 8, 0)
		if len(withVocab) != len(withCanon) {
			t.Fatalf("%s vs %s: different seat counts", p.vocab, p.canon)
		}
		for i := range withVocab {
			if withVocab[i].Score != withCanon[i].Score || withVocab[i].Grade != withCanon[i].Grade {
				t.Fatalf("%s vs %s seat %d: score %.6f/%.6f grade %s/%s diverge",
					p.vocab, p.canon, i, withVocab[i].Score, withCanon[i].Score, withVocab[i].Grade, withCanon[i].Grade)
			}
		}
	}
}

// TestS2_CollapsePin_PDHAbsorbsEQH4h — (c) pin: a today-priority PDH absorbing
// an EQH·4h within the cluster tolerance keeps BOTH names on the map.
func TestS2_CollapsePin_PDHAbsorbsEQH4h(t *testing.T) {
	p := 28910.25
	pair := []ScoredLevel{
		{DetectedLevel: DetectedLevel{Kind: KindPDH, Price: p, Lo: p, Hi: p, Label: "PDH", TF: "1d"}, Score: 0.5, Grade: "C"},
		{DetectedLevel: DetectedLevel{Kind: KindEQH, Price: p + 0.25, Lo: p + 0.25, Hi: p + 0.25, Label: "EQH·4h", TF: "4h"}, Score: 1.5, Grade: "A"},
	}
	collapsed := collapseLevelClusters(pair, clusterToleranceFor(p))
	if len(collapsed) != 1 {
		t.Fatalf("collapse kept %d, want 1", len(collapsed))
	}
	if collapsed[0].Label != "PDH" {
		t.Fatalf("survivor %q, want PDH (today-priority wins regardless of score)", collapsed[0].Label)
	}
	names := namesWithCollapsed(collapsed[0])
	if !namesContain(names, "PDH") || !namesContain(names, "EQH·4h") {
		t.Fatalf("names %v must contain both PDH and EQH·4h", names)
	}
}

func namesContain(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
