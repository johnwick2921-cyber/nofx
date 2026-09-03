package kernel

import (
	"os"
	"strings"
	"testing"
)

// E6 — RETIREMENT. No surface may render a rate from the two biased
// instruments. The Go side is pinned here; the Guide is pinned by its own copy.
func TestBiasedInstrumentsRenderNoRate(t *testing.T) {
	// Both files must carry the retirement notice, so the next reader cannot
	// use them without seeing why they are wrong.
	for _, f := range []string{"level_stats_calc.go", "touch_telemetry.go"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), "NO SURFACE MAY RENDER A RATE FROM IT") {
			t.Errorf("%s must carry the D4 retirement notice", f)
		}
		if !strings.Contains(string(b), "D1′") && !strings.Contains(string(b), "detector_d1prime") {
			t.Errorf("%s must name its calibrated replacement", f)
		}
	}
}

// D7 — the boot line reports REAL counts and names the retirement.
func TestDetectorBootLineIsReadNotLiteral(t *testing.T) {
	line := DetectorBootLine(0, 0)
	for _, want := range []string{"detector: D1′", "k=3", "H=12", "exit_on=close", "touch_outcomes=0", "candidate_pool=0", "legacy rates retired"} {
		if !strings.Contains(line, want) {
			t.Errorf("boot line missing %q: %s", want, line)
		}
	}
	// The counts are arguments, so a populated table shows through.
	if !strings.Contains(DetectorBootLine(41, 820), "touch_outcomes=41") {
		t.Error("the boot line must report the counts it is given, never a literal")
	}
	t.Logf("boot: %s", line)
}
