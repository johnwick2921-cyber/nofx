package ninjatrader

import (
	"testing"
	"time"
)

// ── BAR-SOURCE WAVE PINS (ring half) ─────────────────────────────────────────
//
// The fixture reproduces 2026-09-10 22:39 CT: a cold ring seeded by a replay
// on the LOW scale (~29068), then live updates on the HIGH scale (~29358).
// Boundary and scales are FIXTURE CONSTANTS (A28).
const (
	bsT0    int64   = 1789097820000 // 22:37 CT
	bsLow   float64 = 29068.25
	bsHigh  float64 = 29358.25
	bsDelta         = bsHigh - bsLow // 290.00
)

func seededRing(t *testing.T) *BarCache {
	t.Helper()
	c := NewBarCache(2500)
	var seed []Bar
	for i := 0; i < 30; i++ { // 22:07 .. 22:36, replay, low scale
		ms := bsT0 - int64(30-i)*60_000
		seed = append(seed, Bar{T: ms, O: bsLow, H: bsLow + 1, L: bsLow - 1, C: bsLow, V: 1})
	}
	c.SeedHistorical("MNQ", "1m", seed)
	return c
}

// A REPLAY NEVER OVERWRITES A LIVE BAR — ring half. RED on the pre-wave cache:
// mergeBarsByTime took "incoming is freshest" and the live bar was replaced.
func TestReplayNeverOverwritesALiveBarInTheRing(t *testing.T) {
	c := NewBarCache(2500)
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0, O: bsHigh, H: bsHigh + 1, L: bsHigh - 1, C: bsHigh, V: 1}})
	c.SeedHistorical("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsLow + 1, L: bsLow - 1, C: bsLow, V: 1}})
	got := c.Get("MNQ", "1m")
	if len(got) != 1 {
		t.Fatalf("want 1 bar, got %d", len(got))
	}
	if got[0].C != bsHigh || got[0].Source != BarSourceLive {
		t.Fatalf("the replay overwrote a bar that traded: close=%.2f source=%q (want %.2f live)", got[0].C, got[0].Source, bsHigh)
	}
}

// Live DOES overwrite a historical seed — that direction is correct and must
// keep working (the seed is the stale one when they share a scale).
func TestLiveOverwritesAHistoricalSeedOnTheSameScale(t *testing.T) {
	c := NewBarCache(2500)
	c.SeedHistorical("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsLow + 1, L: bsLow - 1, C: bsLow, V: 1}})
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsLow + 2, L: bsLow - 1, C: bsLow + 1.5, V: 2}})
	got := c.Get("MNQ", "1m")
	if got[0].C != bsLow+1.5 || got[0].Source != BarSourceLive {
		t.Fatalf("a same-scale live update must replace the seed: close=%.2f source=%q", got[0].C, got[0].Source)
	}
}

// THE BOOT MINUTE IS NEVER A MIXED BAR. The first live bar after a replay seed
// closes 290 points from the last replay close: the seed is on another scale.
// The seed is DROPPED, the straddling bar is labelled MIXED (values kept), and
// the listener fires ONCE with the numbers.
func TestScaleMismatchDropsTheSeedAndLabelsTheBootMinute(t *testing.T) {
	c := seededRing(t)
	if n := c.Count("MNQ", "1m"); n != 30 {
		t.Fatalf("seed: want 30, got %d", n)
	}
	var got []ScaleMismatch
	OnScaleMismatch(func(m ScaleMismatch) { got = append(got, m) })
	t.Cleanup(func() { scaleListenersMu.Lock(); scaleListeners = nil; scaleListenersMu.Unlock() })

	// 22:37 arrives live: the AddOn copied its open from its own historical
	// series (low) and its close is live (high) — the mixed shape.
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsHigh + 1, L: bsLow - 1, C: bsHigh, V: 3}})
	time.Sleep(20 * time.Millisecond) // listener runs on its own goroutine

	bars := c.Get("MNQ", "1m")
	if len(bars) != 1 {
		t.Fatalf("the historical seed must be DROPPED on a scale mismatch; ring holds %d", len(bars))
	}
	if bars[0].Source != BarSourceMixed {
		t.Fatalf("the straddling bar must be labelled mixed, got %q", bars[0].Source)
	}
	if bars[0].O != bsLow || bars[0].C != bsHigh {
		t.Fatalf("A24: the mixed bar's VALUES must be untouched; got o=%.2f c=%.2f", bars[0].O, bars[0].C)
	}
	if len(got) != 1 || got[0].DeltaPts != bsDelta || got[0].HistoricalDropped != 30 {
		t.Fatalf("listener: want one event Δ=%.2f dropped=30, got %+v", bsDelta, got)
	}
	// A second live bar is ordinary live — no second event, no second mixed.
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0 + 60_000, O: bsHigh, H: bsHigh + 1, L: bsHigh - 1, C: bsHigh + 0.5, V: 1}})
	time.Sleep(20 * time.Millisecond)
	bars = c.Get("MNQ", "1m")
	if len(got) != 1 || bars[len(bars)-1].Source != BarSourceLive {
		t.Fatalf("only the FIRST live bar is checked; events=%d last source=%q", len(got), bars[len(bars)-1].Source)
	}
}

// Same scale → no mismatch, nothing dropped, nothing labelled.
func TestNoMismatchOnTheSameScale(t *testing.T) {
	c := seededRing(t)
	fired := 0
	OnScaleMismatch(func(ScaleMismatch) { fired++ })
	t.Cleanup(func() { scaleListenersMu.Lock(); scaleListeners = nil; scaleListenersMu.Unlock() })
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsLow + 2, L: bsLow - 1, C: bsLow + 1, V: 1}})
	time.Sleep(20 * time.Millisecond)
	if fired != 0 || c.Count("MNQ", "1m") != 31 {
		t.Fatalf("same scale: fired=%d ring=%d (want 0, 31)", fired, c.Count("MNQ", "1m"))
	}
	for _, b := range c.Get("MNQ", "1m") {
		if b.Source == BarSourceMixed {
			t.Fatal("nothing may be labelled mixed on a same-scale boot")
		}
	}
}

// Every bar the ring holds names its source — a replay is historical, an
// update is live, and nothing is unlabelled.
func TestEveryRingBarNamesItsSource(t *testing.T) {
	c := seededRing(t)
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsLow + 1, L: bsLow - 1, C: bsLow + 0.25, V: 1}})
	for _, b := range c.Get("MNQ", "1m") {
		if b.Source != BarSourceLive && b.Source != BarSourceHistorical {
			t.Fatalf("unlabelled bar at %d: %q", b.T, b.Source)
		}
	}
}

// A MID-SESSION RECONNECT RE-ARMS THE CHECK. Boot on one scale (no mismatch),
// live for a while, then a reconnect replays on a DIFFERENT scale: the first
// live bar after THAT seed must catch it. RED with a per-process guard.
func TestReconnectReseedIsCheckedAgain(t *testing.T) {
	c := seededRing(t) // low-scale seed
	fired := 0
	OnScaleMismatch(func(ScaleMismatch) { fired++ })
	t.Cleanup(func() { scaleListenersMu.Lock(); scaleListeners = nil; scaleListenersMu.Unlock() })

	// live on the SAME (low) scale — a clean boot, nothing fires
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0, O: bsLow, H: bsLow + 1, L: bsLow - 1, C: bsLow + 0.5, V: 1}})
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0 + 60_000, O: bsLow, H: bsLow + 1, L: bsLow - 1, C: bsLow + 0.25, V: 1}})
	time.Sleep(20 * time.Millisecond)
	if fired != 0 {
		t.Fatalf("same-scale boot must not fire; fired=%d", fired)
	}
	// reconnect: the replay comes back on the HIGH scale for later minutes
	var re []Bar
	for i := 2; i < 6; i++ {
		re = append(re, Bar{T: bsT0 + int64(i)*60_000, O: bsHigh, H: bsHigh + 1, L: bsHigh - 1, C: bsHigh, V: 1})
	}
	c.SeedHistorical("MNQ", "1m", re)
	// first live bar after the re-seed, on the LOW scale (live never moved)
	c.Upsert("MNQ", "1m", []Bar{{T: bsT0 + 6*60_000, O: bsHigh, H: bsHigh + 1, L: bsLow - 1, C: bsLow, V: 1}})
	time.Sleep(20 * time.Millisecond)
	if fired != 1 {
		t.Fatalf("a reconnect re-seed on another scale must be caught by the first live bar after it; fired=%d", fired)
	}
}
