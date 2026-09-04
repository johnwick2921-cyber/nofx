package store

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ── KNOB REGISTRY (settings integrity, 2026-09-03) ───────────────────────────
//
// The Studio audit (2026-09-03-studio-audit.md) found FIFTEEN saved settings
// that cannot take effect: two unreachable clock fields, four parse-only flags,
// a prompt-text-only "limit", a session field with no consumer, two env vars
// with no consumer, a hardcoded crypto override, a dropped per-session
// plan_mode, an env var replacing the Studio R:R floor, an always-on veto, a
// clamped size, a suspended toggle whose wire is untouched, and an
// external-data list with zero engine consumers.
//
// The registry exists so that class cannot recur SILENTLY. The schema is
// enumerated by REFLECTION — it can never drift from the struct — and every
// enumerated field must carry a status. Adding a field and leaving it
// unclassified fails the build. That is the point, not a side effect.
type KnobStatus string

const (
	KnobLive        KnobStatus = "live"         // a consumer reads it and behaviour changes
	KnobSuspended   KnobStatus = "suspended"    // a standing ruling disables it; value preserved
	KnobAdvisory    KnobStatus = "advisory"     // feeds prompt text only — never a gate
	KnobDisplayOnly KnobStatus = "display-only" // rendered, never read by the engine
	KnobDead        KnobStatus = "dead"         // saved value CANNOT take effect (audit's 15)
	KnobInfra       KnobStatus = "infra"        // ports, paths, keys — not a trading knob
)

// KnobEntry is one schema field's classification.
type KnobEntry struct {
	Path      string     // dotted schema path, e.g. "risk_control.min_risk_reward_ratio"
	Status    KnobStatus //
	Consumers []string   // file:line — REQUIRED when live, or the entry is a claim without a reader
	DualLevel bool       // has a per-session override
	Clamp     string     // "" = none; else what the clamp does to the saved value
	Note      string     // REQUIRED when not live: WHY
}

// EnumerateSchemaKnobs walks StrategyConfig by reflection and returns every
// json-tagged leaf as a dotted path. Reflection rather than a hand list: a
// hand-maintained enumeration is exactly how a field goes unclassified.
func EnumerateSchemaKnobs() []string {
	seen := map[string]bool{}
	var out []string
	var walk func(t reflect.Type, prefix string, depth int)
	walk = func(t reflect.Type, prefix string, depth int) {
		if depth > 6 || t == nil {
			return
		}
		for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")[0]
			if tag == "" || tag == "-" {
				continue
			}
			path := tag
			if prefix != "" {
				path = prefix + "." + tag
			}
			ft := f.Type
			for ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Slice {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				walk(ft, path, depth+1)
				continue
			}
			if !seen[path] {
				seen[path] = true
				out = append(out, path)
			}
		}
	}
	walk(reflect.TypeOf(StrategyConfig{}), "", 0)
	sort.Strings(out)
	return out
}

// LookupKnob returns the registry entry for a schema path.
func LookupKnob(path string) (KnobEntry, bool) {
	if e, ok := knobRegistry[path]; ok {
		return e, ok
	}
	// The table is keyed by json LEAF name; reflection yields DOTTED paths.
	// Fall back to the leaf so the two agree. Documented rather than hidden:
	// a leaf name can collide across structs, and where it does the first
	// classification wins — which is why the 2026-09-03 sweep records the
	// enumerated-vs-tagged gap in the report instead of claiming completeness.
	if i := strings.LastIndex(path, "."); i >= 0 {
		e, ok := knobRegistry[path[i+1:]]
		return e, ok
	}
	return KnobEntry{}, false
}

// KnobSummary is what the boot line reports — counted, never typed.
type KnobSummary struct {
	Total, Live, Suspended, Advisory, DisplayOnly, Dead, Infra int
	EnvShadows                                                 int
	EnvShadowPaths                                             []string
}

// KnobStatusSummary counts the registry by status.
func KnobStatusSummary() KnobSummary {
	s := KnobSummary{}
	for _, e := range knobRegistry {
		s.Total++
		switch e.Status {
		case KnobLive:
			s.Live++
		case KnobSuspended:
			s.Suspended++
		case KnobAdvisory:
			s.Advisory++
		case KnobDisplayOnly:
			s.DisplayOnly++
		case KnobDead:
			s.Dead++
		case KnobInfra:
			s.Infra++
		}
	}
	return s
}

// KnobRegistryBootLine reports the registry — every field READ from it.
func KnobRegistryBootLine() string {
	s := KnobStatusSummary()
	fields := len(EnumerateSchemaKnobs())
	unclassified := 0
	for _, p := range EnumerateSchemaKnobs() {
		if _, ok := LookupKnob(p); !ok {
			unclassified++
		}
	}
	warn := ""
	if unclassified > 0 {
		warn = fmt.Sprintf(" · ⚠ %d UNCLASSIFIED", unclassified)
	}
	return fmt.Sprintf("settings: schema=%d classified=%d live=%d suspended=%d advisory=%d display-only=%d dead=%d infra=%d · env-shadows=%d%s",
		fields, s.Total, s.Live, s.Suspended, s.Advisory, s.DisplayOnly, s.Dead, s.Infra, s.EnvShadows, warn)
}

// AuditDeadKnobs2026_09_03 is the audit's fifteen, by schema path, so the
// registry can be checked against the finding that produced it.
var AuditDeadKnobs2026_09_03 = []string{
	"day_plan.last_entry_ct",
	"day_plan.eod_flat_ct",
	"risk_control.max_contracts_enabled",
	"risk_control.notional_cap_enabled",
	"risk_control.max_margin_usage",
	"indicator_config.external_data_sources",
}
