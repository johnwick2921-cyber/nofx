// W5 / E1 — the GUIDE may not type a clock the machine resolves.
//
// TestNoTradeWindowsHaveNoSurfaceLiterals already forbids a hardcoded lunch
// window in Go source. It scans Go files only, so the Guide — the surface the
// OWNER actually reads — was the one place the window could drift unwatched,
// and it did: web/src/guide/content/plays.ts told the operator
//
//	lunch 11:30–13:30 ET → no entries
//
// while kernel.LunchWindowCT() resolves 12:00–13:30 CT. That is an hour off, in
// the wrong clock, describing a window where entries are genuinely refused.
//
// The same line typed "10:30 ET" twice, where the prompt renders the identical
// cue through ETtoCT(). One class, one fixture: the Guide states a machine
// window by NAMING ITS RESOLVER, never by typing digits.

package kernel

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// guideContentFiles are read relative to the kernel package (../web/...).
func guideContentFiles(t *testing.T) map[string]string {
	t.Helper()
	dir := filepath.Join("..", "web", "src", "guide", "content")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read guide content dir: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".ts") || strings.HasSuffix(e.Name(), ".test.ts") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[e.Name()] = string(b)
	}
	if len(out) == 0 {
		t.Fatal("no guide content files found — the scan would pass vacuously")
	}
	return out
}

// ctPlusOneHour renders a CT "HH:MM" in Eastern. Test-local by design.
func ctPlusOneHour(hhmm string) string {
	var h, m int
	if _, err := fmt.Sscanf(hhmm, "%d:%d", &h, &m); err != nil {
		return hhmm
	}
	return fmt.Sprintf("%02d:%02d", (h+1)%24, m)
}

// A typed lunch window is allowed ONLY in a file that also names the resolver.
//
// Four Guide files typed the CORRECT window and cited nothing; one typed the
// WRONG one. Rewriting the four true sentences to look busy would be worse than
// leaving them (H), so the rule is not "never type it" — it is "never type it
// unsourced". tradingDay.ts already does this correctly and is the model:
// it renders the window and names kernel.LunchWindowCT beside it, so a reader
// who doubts the number knows exactly where to check it.
func TestGuideDoesNotTypeTheLunchWindow(t *testing.T) {
	ls, le := LunchWindowCT()
	// ET is CT+1. Computed here, in the test, rather than adding a production
	// converter with no production caller (A29).
	etStart, etEnd := ctPlusOneHour(ls), ctPlusOneHour(le)

	for name, src := range guideContentFiles(t) {
		typed := ""
		for _, lit := range []string{
			ls + "\u2013" + le, ls + "-" + le,
			etStart + "\u2013" + etEnd, etStart + "-" + etEnd,
			"11:30\u201313:30", "11:30-13:30",
		} {
			if strings.Contains(src, lit) {
				typed = lit
				break
			}
		}
		if typed == "" {
			continue
		}
		// The wrong window is a false claim wherever it appears.
		if strings.HasPrefix(typed, "11:30") {
			t.Errorf("%s states the lunch window as %q — kernel.LunchWindowCT() resolves %s\u2013%s CT; the typed copy is an hour off and in the wrong clock", name, typed, ls, le)
			continue
		}
		// A correct window still has to say where it came from.
		if !strings.Contains(src, "LunchWindowCT") {
			t.Errorf("%s types the lunch window %q without naming its resolver — cite kernel.LunchWindowCT so the number can be checked (tradingDay.ts is the pattern)", name, typed)
		}
	}
}

// No raw Eastern clock may appear in Guide content. The machine speaks CT; an
// ET time in the Guide is either a second clock the reader must convert or a
// window that has silently drifted from its resolver.
var guideETClock = regexp.MustCompile(`\d{1,2}:\d{2}\s*ET\b`)

func TestGuideStatesNoRawEasternClock(t *testing.T) {
	for name, src := range guideContentFiles(t) {
		if hits := guideETClock.FindAllString(src, -1); len(hits) > 0 {
			t.Errorf("%s states %d raw Eastern clock(s) %v — the prompt renders the same cues through ETtoCT(); the Guide must state CT or name the resolver", name, len(hits), hits)
		}
	}
}
