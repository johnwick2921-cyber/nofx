package store

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// E1 — REGISTRY COMPLETENESS. Every schema field reflection can reach must
// carry a classification; a "live" one must name a consumer. Adding a field and
// leaving it unclassified FAILS THE BUILD, which is the point.
func TestKnobRegistryIsComplete(t *testing.T) {
	fields := EnumerateSchemaKnobs()
	if len(fields) < 40 {
		t.Fatalf("reflection found only %d fields — the enumerator is broken, not the registry", len(fields))
	}
	var missing, liveNoConsumer []string
	for _, p := range fields {
		e, ok := LookupKnob(p)
		if !ok {
			missing = append(missing, p)
			continue
		}
		if e.Status == KnobLive && len(e.Consumers) == 0 {
			liveNoConsumer = append(liveNoConsumer, p)
		}
	}
	sort.Strings(missing)
	sort.Strings(liveNoConsumer)
	if len(missing) > 0 {
		show := missing
		if len(show) > 10 {
			show = show[:10]
		}
		t.Errorf("%d schema field(s) unclassified — a setting nobody classified may not take effect:\n  %s",
			len(missing), strings.Join(show, "\n  "))
	}
	if len(liveNoConsumer) > 0 {
		t.Errorf("%d knob(s) marked LIVE with no consumer — the audit's dead-knob signature:\n  %s",
			len(liveNoConsumer), strings.Join(liveNoConsumer, "\n  "))
	}
	t.Logf("registry: %d reflected fields, all classified · %s", len(fields), KnobRegistryBootLine())
}

// The audit's dead knobs were REMOVED (FIX-KNOBS A, 2026-09-26) — the registry
// must no longer know them: a removed dead knob that still classifies is a
// knob that looks active and does nothing.
func TestRegistryDoesNotCallTheAuditsDeadKnobsLive(t *testing.T) {
	for _, p := range AuditDeadKnobs2026_09_03 {
		if e, ok := LookupKnob(p); ok {
			t.Errorf("%s: the audit named it and FIX-KNOBS A removed it — the registry must not still classify it (%+v)", p, e)
		}
	}
}

// The boot line is counted from the registry, never typed. W1 (f): schema= is
// READ from the production enumeration (not a literal), and env-shadows — a
// counter nothing writes — reads n/a instead of a fabricated 0 (L7).
func TestKnobBootLineIsCounted(t *testing.T) {
	line := KnobRegistryBootLine() // the call main.go logs as "⚙ %s"
	schema := fmt.Sprintf("settings: schema=%d ", len(EnumerateSchemaKnobs()))
	envShadows := "env-shadows=n/a (not counted)"
	if s := KnobStatusSummary(); s.EnvShadows != nil {
		envShadows = fmt.Sprintf("env-shadows=%d", *s.EnvShadows)
	}
	// The ruling's two labels must BOTH appear — conflating them is the defect.
	for _, want := range []string{schema, "classified=", "live=", "ineffective=", "candidate-unverified=", envShadows} {
		if !strings.Contains(line, want) {
			t.Errorf("boot line missing %q: %s", want, line)
		}
	}
	if strings.Contains(line, "env-shadows=0") {
		t.Errorf("env-shadows printed 0 from a counter with no writer: %s", line)
	}
	if strings.Contains(line, "UNCLASSIFIED") {
		t.Errorf("the registry has unclassified fields: %s", line)
	}
	t.Logf("boot: %s", line)
}
