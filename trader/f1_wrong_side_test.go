package trader

import "testing"

// F1 (E7 resend-loop incident 2026-08-30) — the ONE merged wrong-side predicate:
// price has accepted through a resting limit's level ⇒ placing it would fill
// INSTANTLY at market (the S2 re-place loop: fill → stop-out → re-arm → fill…).
// Stored in trader/armed_executor.go as limitMarketableWrongSide; the re-author
// half of the fix lives in store.UpsertArm (manual-cancel-wins: a terminal row
// at the same plan version stays terminal — covered by armed_orders_test.go).

func TestLimitMarketableWrongSideStrictBoundary(t *testing.T) {
	cases := []struct {
		name  string
		side  string
		entry float64
		price float64
		want  bool
	}{
		{"long below market rests", "long", 29371.5, 29400, false},
		// Strict boundary: a limit exactly AT market is marketable too.
		{"long AT market is marketable", "long", 29371.5, 29371.5, true},
		{"long above market is marketable", "long", 29371.5, 29347.25, true},
		{"short above market rests", "short", 29300, 29290, false},
		{"short AT market is marketable", "short", 29300, 29300, true},
		{"short below market is marketable", "short", 29300, 29346.25, true},
		{"no price never refuses", "long", 29371.5, 0, false},
		{"no entry never refuses", "short", 0, 29346.25, false},
		{"side whitespace/case tolerated", " LONG ", 29371.5, 29347.25, true},
	}
	for _, c := range cases {
		if got := limitMarketableWrongSide(c.price, c.entry, c.side); got != c.want {
			t.Errorf("%s: limitMarketableWrongSide(price=%.2f, entry=%.2f, %q)=%v want %v",
				c.name, c.price, c.entry, c.side, got, c.want)
		}
	}
}

func TestThroughWord(t *testing.T) {
	if throughWord("short") != "above" || throughWord("SHORT") != "above" {
		t.Fatal("a short is traded-through when price is ABOVE its entry")
	}
	if throughWord("long") != "below" {
		t.Fatal("a long is traded-through when price is BELOW its entry")
	}
}
