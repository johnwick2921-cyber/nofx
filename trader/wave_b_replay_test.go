// WAVE B (2026-09-05) — the replay, the A29 call-site pin, and the D4 boot line.

package trader

import (
	"os"
	"strings"
	"testing"

	ntwire "nofx/provider/ninjatrader"
)

// ny0904S2Prices are the 21 cycle prices the executor actually saw on
// 2026-09-04 between 10:05:00 and 10:53:11 CT, one per placement, read from the
// "📏 arm far … from price X" line in data/nofx_2026-09-04.log. They belong to
// armed_orders ids 38, 62, 65, 67, 70, 73, 75, 77, 79, 81, 83, 85, 87, 89, 91,
// 93, 95, 97, 99, 101, 102 — NY / v3 / S2 / SHORT, entry_px 29591.02, wire
// trigger 29590.50 (n=21, and 21 is the INVESTIGATORS' measured submission
// count, not the dispatch's framing of one arm).
var ny0904S2Prices = []float64{
	29515.25, 29524.75, 29540.50, 29536.25, 29501.00, 29500.50, 29505.00,
	29503.75, 29505.50, 29487.75, 29497.00, 29500.75, 29495.50, 29503.50,
	29511.50, 29521.00, 29518.75, 29495.50, 29501.50, 29500.25, 29500.25,
}

const ny0904S2Trigger = 29590.50

// TestReplayNY0904S2ThroughTheFixedGuard — E7.
//
// CORRECTION TO THE DISPATCH, and it is the whole point of shipping D1 with D2:
// the dispatch expects "ONE well-formed submission with stopPrice set". Replayed
// against the MEASURED prices the fixed path yields ZERO submissions. At every
// one of the 21 cycles the market was 50.00-102.75 points BELOW a SELL-stop
// trigger — squarely already-through — so the correct answer is 21 cancels and
// no order at all. The 21 that were sent were sent because the guard was
// inverted; the only reason they did no harm is that D1's zero stop slot made
// them inert. A well-formed submission appears in the valid-side replay below.
func TestReplayNY0904S2ThroughTheFixedGuard(t *testing.T) {
	placed, cancelled := 0, 0
	for i, price := range ny0904S2Prices {
		v, why := stopEntryGuardVerdict("short", ny0904S2Trigger, price)
		switch v {
		case stopGuardThrough:
			cancelled++
			if !strings.Contains(why, "accepted through (stop side)") {
				t.Errorf("cycle %d: reason does not name the guard: %q", i, why)
			}
		case stopGuardRest:
			placed++
			t.Errorf("cycle %d: price %.2f is %.2f pts BELOW a sell-stop trigger %.2f and must never be placed",
				i, price, ny0904S2Trigger-price, ny0904S2Trigger)
		default:
			t.Errorf("cycle %d: unadjudicated with a real price and trigger: %s", i, why)
		}
	}
	if len(ny0904S2Prices) != 21 {
		t.Fatalf("the replay must carry all 21 measured cycles, got %d", len(ny0904S2Prices))
	}
	if placed != 0 || cancelled != 21 {
		t.Fatalf("replay of the 21 measured cycles: placed=%d cancelled=%d, want 0/21", placed, cancelled)
	}

	// The valid side of the same arm: price ABOVE the sell-stop trigger. Exactly
	// one placement, and the frame carries the trigger in stop_price with an
	// empty limit slot (the wire half is pinned in
	// trader/ninjatrader/stop_entry_wire_test.go; here we pin the decision).
	v, why := stopEntryGuardVerdict("short", ny0904S2Trigger, 29650.00)
	if v != stopGuardRest {
		t.Fatalf("a sell stop with the market above its trigger must rest: %v (%s)", v, why)
	}
	if !strings.Contains(why, "rests (stop side)") {
		t.Fatalf("resting reason must say so: %q", why)
	}
}

// TestReplayArm35LimitUnchanged — E4. Arm 35 (NY, S1, SHORT, entry_px 29285.00,
// stop_px 29351.63, target_px 29144.50, state filled, 2026-09-03) is a LIMIT
// arm. Its adjudication must be byte-for-byte the behaviour it had before this
// wave: the limit predicate is untouched and is still what the limit branch
// calls. Nothing about the stop-side fix may reach it.
func TestReplayArm35LimitUnchanged(t *testing.T) {
	const arm35Entry = 29285.00
	for _, c := range []struct {
		price float64
		want  bool
		note  string
	}{
		{29280.00, false, "market below a sell limit — rests"},
		{29285.00, false, "exactly at a sell limit — still rests (STRICT boundary, unlike a stop)"},
		{29290.00, true, "market above a sell limit — marketable"},
	} {
		if got := limitMarketableWrongSide(c.price, arm35Entry, "short"); got != c.want {
			t.Errorf("arm 35 limit replay price=%.2f (%s): got %v want %v", c.price, c.note, got, c.want)
		}
	}
	// The two predicates disagree at exactly the level, which is the difference
	// this wave exists to preserve: a limit AT its price rests, a stop AT its
	// trigger fires.
	if limitMarketableWrongSide(arm35Entry, arm35Entry, "short") {
		t.Error("a limit at its own price must rest")
	}
	if !stopEntryMarketableWrongSide("short", arm35Entry, arm35Entry) {
		t.Error("a stop at its own trigger must read as already through")
	}
}

// TestStopEntryGuardHasAProductionCallSite — E8 / A29. Built is not wired. The
// stop-entry placement branch must call the STOP guard, and must no longer call
// the limit predicate with the trigger. Removing the call breaks this test.
func TestStopEntryGuardHasAProductionCallSite(t *testing.T) {
	b, err := os.ReadFile("armed_executor.go")
	if err != nil {
		t.Fatalf("cannot read the placement source: %v", err)
	}
	src := string(b)
	if !strings.Contains(src, "verdict, why := stopEntryGuardVerdict(r.Side, trigger, price)") {
		t.Error("the stop-entry branch does not call stopEntryGuardVerdict — the guard is built but not wired")
	}
	if strings.Contains(src, "limitMarketableWrongSide(price, trigger,") {
		t.Error("the stop-entry branch still calls the LIMIT predicate with the trigger — the inversion is back")
	}
	// The limit branch must keep its own predicate, with the ENTRY argument.
	if !strings.Contains(src, "limitMarketableWrongSide(price, r.EntryPx, r.Side)") {
		t.Error("the limit branch lost limitMarketableWrongSide — the limit path was not supposed to move")
	}
	// D5's refusal must be counted, not swallowed.
	if !strings.Contains(src, "errors.Is(perr, ntwire.ErrAddonBuildTooOld)") {
		t.Error("a build refusal is not distinguished from a transport failure — it cannot be counted")
	}
}

// TestStopEntryBootLineIsRead — D4 / A11. Every field is resolved from the code
// that enforces it; none is a literal. A proven build reads slots=stop_price and
// match=yes, an unproven one must say so on both.
func TestStopEntryBootLineIsRead(t *testing.T) {
	proven := StopEntryBootLine(ntwire.MinAddonBuildStopSlot, ntwire.ExpectedAddonBuild)
	for _, want := range []string{"🎯 stop-entry:", "slots=stop_price", "guard=stop-side", "unknown=no-op", "match=yes"} {
		if !strings.Contains(proven, want) {
			t.Errorf("proven-build line missing %q: %s", want, proven)
		}
	}
	if !strings.Contains(proven, "build_id="+ntwire.MinAddonBuildStopSlot) {
		t.Errorf("the line must name the RECEIVED build id: %s", proven)
	}

	// The build NT8 is running today (the one whose CreateOrder put the trigger
	// in the limit slot) must not be able to render as proven.
	old := StopEntryBootLine("2026-09-03-f12", ntwire.ExpectedAddonBuild)
	if strings.Contains(old, "slots=stop_price") {
		t.Errorf("a pre-stop-slot build must not claim slots=stop_price: %s", old)
	}
	if !strings.Contains(old, "match=NO") || !strings.Contains(old, "build_id=2026-09-03-f12") {
		t.Errorf("a stale DLL must read match=NO with its own id: %s", old)
	}

	// No frame received yet: "none", never an empty string that reads as data.
	none := StopEntryBootLine("", ntwire.ExpectedAddonBuild)
	if !strings.Contains(none, "build_id=none") || strings.Contains(none, "build_id= ") {
		t.Errorf("an unknown build must render as none: %s", none)
	}
	if strings.Contains(none, "slots=stop_price") {
		t.Errorf("an unheard-from AddOn cannot prove the slot: %s", none)
	}
	// The expected half is READ from the source constant, never restated.
	if !strings.Contains(proven, "expected="+ntwire.ExpectedAddonBuild) {
		t.Errorf("expected= must be the source constant: %s", proven)
	}
}
