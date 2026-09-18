package trader

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/levelidentity"
	"nofx/store"
)

// W-GEOMETRY-REFUSAL (2026-09-18) — executor-side tests. (a) the WARN, (b1)
// stable reference ids, (b2) the empty-tf wildcard, (b4) OFF byte-identical.

func refPtr(v float64) *float64 { return &v }
func refStr(v string) *string   { return &v }

// nilLevelIDOverride returns the one-setup fixture doc with S1's level_id set
// to nil — the exact live shape of a reject play authored at ONH/ONL (LONDON
// v1 S1 2026-09-18 [A]).
func nilLevelIDOverride(t *testing.T) string {
	t.Helper()
	var d kernel.PlanDoc
	if err := json.Unmarshal([]byte(oneSetupFixtureDoc()), &d); err != nil {
		t.Fatal(err)
	}
	for i := range d.Scenarios {
		if d.Scenarios[i].ID == "S1" {
			d.Scenarios[i].LevelID = nil
		}
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestGeometryRefusalWarnOncePerKeyChange (a): a reject scenario with LevelID
// nil produces exactly ONE ⚔️ arm REFUSED WARN across TWO cycles (de-duped by
// (geometry key, reason) exactly like the other arm refusals). RED on dev
// b70fc6ca: 0 WARN lines (INFO-only). Driven on a SYNTHETIC in-session clock so
// the test never depends on the wall clock (the shared drive harness skips
// outside a live session window).
func TestGeometryRefusalWarnOncePerKeyChange(t *testing.T) {
	oneSetupFixtureDocOverride = nilLevelIDOverride(t)
	entry := 29010.0
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}
	cfg.RiskControl.MinRiskRewardRatio = 2
	structuralTestPolicy(&cfg, 5)
	oneSetupOff(&cfg) // isolate geometry from the independently tested selection checks
	at, st := resetTrader(t, cfg)
	now := time.Date(2026, 8, 18, 14, 0, 0, 0, time.UTC) // 09:00 CT, inside NY
	testNow = func() time.Time { return now }
	t.Cleanup(func() { testNow = nil })
	sess, ok := at.sessionRegistry(now).ActiveSession(now)
	if !ok {
		t.Fatal("no active session on the synthetic clock")
	}
	cfg.DayPlan.SessionsEnabled = []string{sess.Name}
	trueV := true
	cfg.DayPlan.Sessions = []store.DayPlanSessionOverride{{Session: sess.Name, Enable: &trueV}}
	td, _ := kernel.PlanChainTradeDate(sess, now)
	pid := store.MakePlanIDForTrader(at.id, td, sess.Name)
	if _, err := st.Plan().AppendPlan(&store.PlanDB{PlanID: pid, TradeDate: td, Session: sess.Name, StrategyID: at.id, Lifecycle: "active", Doc: oneSetupFixtureDocOverride, CreatedAt: now.Add(-30 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	installActivePlanProviderAt(at, st, func() time.Time { return now })
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(symbol string, tf string, n int) []market.Kline { return oneSetupFixtureTape(now) }
	t.Cleanup(func() { market.FuturesBarsProvider = prev })
	at.oneSetupFactsForTest = func(now time.Time) oneSetupTestFacts {
		return oneSetupTestFacts{Price: entry, BandPts: 100, Candidates: []kernel.MapCandidate{{Price: entry, Names: []string{"PDL"}, Grade: "A"}}, Permission: map[string]kernel.FadeVerdict{"S1": {Evaluated: true, Permitted: true}}}
	}
	buf := captureTraderLog(t)
	at.maybeManageArmedOrdersAt(nil, now)
	at.maybeManageArmedOrdersAt(nil, now) // second cycle — same key/reason → no second WARN

	log := buf.String()
	if !strings.Contains(log, "geometry_no_provenance (scenario_level_id_missing)") {
		t.Fatalf("the geometry refusal must WARN with its reason; got:\n%s", log)
	}
	if n := strings.Count(log, "⚔️ arm REFUSED "); n != 1 {
		t.Fatalf("exactly ONE WARN across two cycles (de-duped), got %d:\n%s", n, log)
	}
}

// TestEnsureReferenceLevelIDsStableAndScoped (b1): anchor kinds with no
// formation close get a STABLE ref| id; non-anchor kinds stay NULL; strict ids
// are never overwritten.
func TestEnsureReferenceLevelIDsStableAndScoped(t *testing.T) {
	mk := func(kind string) kernel.MapCandidate {
		return kernel.MapCandidate{Identity: kernel.PlanLevel{
			Symbol: refStr("MNQ"), Kind: refStr(kind), Lo: refPtr(29400), Hi: refPtr(29410),
			OriginDate: refStr("2026-09-17"), TF: refStr("1m"), Label: kind, Price: 29405,
		}}
	}
	cs := []kernel.MapCandidate{mk("ONH"), mk("FVG")}
	kernel.EnsureReferenceLevelIDs(cs)
	if cs[0].ID == nil {
		t.Fatal("ONH (anchor kind, no formation close) must get a stable id")
	}
	if !strings.HasPrefix(*cs[0].ID, "ref|") {
		t.Fatalf("the fallback id must carry the ref| prefix: %s", *cs[0].ID)
	}
	if cs[1].ID != nil {
		t.Fatal("a non-anchor kind with NULL id must stay NULL")
	}
	again := []kernel.MapCandidate{mk("ONH")}
	kernel.EnsureReferenceLevelIDs(again)
	if *cs[0].ID != *again[0].ID {
		t.Fatalf("same inputs must yield the SAME id: %s vs %s", *cs[0].ID, *again[0].ID)
	}
	full := mk("ONH")
	nowMs := time.Now().UnixMilli()
	full.Identity.FormedCloseMs = &nowMs
	strict, _ := levelidentity.ID(kernel.IdentityInputs(full.Identity))
	full.ID = strict
	full.Identity.ID = strict
	kernel.EnsureReferenceLevelIDs([]kernel.MapCandidate{full})
	if *full.ID != *strict {
		t.Fatal("a strict id must never be overwritten by the reference id")
	}
}

// refAnchorDoc builds the frozen-map doc for the tf-wildcard tests.
func refAnchorDoc(sourceTF string) (*kernel.PlanDoc, *string) {
	id := kernel.ReferenceLevelID("MNQ", "eVWAP", 29400, 29410, "2026-09-17", "1m")
	identity := kernel.PlanLevel{Symbol: refStr("MNQ"), Kind: refStr("eVWAP"), Lo: refPtr(29400), Hi: refPtr(29410), OriginDate: refStr("2026-09-17"), TF: refStr("1m"), Label: "eVWAP", Price: 29405, ID: id}
	doc := &kernel.PlanDoc{
		IdentityLevels: []kernel.PlanLevel{identity},
		Zones: &kernel.LevelZoneMap{Zones: []kernel.LevelZone{{
			Lo: refPtr(29400), Hi: refPtr(29410),
			Sources: []kernel.ZoneSource{{Price: 29405, Label: "eVWAP", TF: sourceTF}},
		}}},
	}
	return doc, id
}

// TestGeometryTFWildcardExecutorOnly (b2): an EMPTY source tf refuses under the
// legacy contract and with the knob OFF (parity), and resolves with the knob ON.
// A NON-empty mismatched tf still refuses even with the knob ON.
func TestGeometryTFWildcardExecutorOnly(t *testing.T) {
	doc, id := refAnchorDoc("")
	sc := kernel.PlanScenario{ID: "S1", LevelID: id}

	if idx, why := ResolveEntryGeometryZone(doc, sc); why == "" {
		t.Fatalf("legacy must refuse the empty-tf source (today's behaviour), got idx=%d", idx)
	}
	if idx, why := ArmGeometryVerdict(doc, sc, false); why == "" {
		t.Fatalf("knob OFF must refuse (byte-identical parity), got idx=%d", idx)
	}
	if idx, why := ArmGeometryVerdict(doc, sc, true); why != "" || idx != 0 {
		t.Fatalf("knob ON must treat the empty source tf as a wildcard: idx=%d why=%s", idx, why)
	}

	docMismatch, id2 := refAnchorDoc("5m")
	scMismatch := kernel.PlanScenario{ID: "S1", LevelID: id2}
	if idx, why := ArmGeometryVerdict(docMismatch, scMismatch, true); why == "" {
		t.Fatalf("a non-empty mismatched tf must still refuse with the knob ON, got idx=%d", idx)
	}
}

// TestGeometryReferenceLevelsKnobResolution (b4): nil config or nil pointer →
// ON (owner default); explicit false → OFF.
func TestGeometryReferenceLevelsKnobResolution(t *testing.T) {
	if !(*store.DayPlanConfig)(nil).GeometryRefIDsEnabled() {
		t.Fatal("nil config must resolve ON (owner default)")
	}
	if !(&store.DayPlanConfig{}).GeometryRefIDsEnabled() {
		t.Fatal("nil pointer must resolve ON (owner default)")
	}
	off := false
	if (&store.DayPlanConfig{GeometryReferenceLevels: &off}).GeometryRefIDsEnabled() {
		t.Fatal("explicit false must resolve OFF")
	}
	on := true
	if !(&store.DayPlanConfig{GeometryReferenceLevels: &on}).GeometryRefIDsEnabled() {
		t.Fatal("explicit true must resolve ON")
	}
}

// TestGeometryRefBootLineReadsResolvedKnob (c): the per-trader boot line reads
// the resolved knob, never a literal.
func TestGeometryRefBootLineReadsResolvedKnob(t *testing.T) {
	off := false
	line := GeometryRefBootLine(&store.DayPlanConfig{GeometryReferenceLevels: &off})
	if !strings.Contains(line, "geom_ref_ids=off") {
		t.Fatalf("boot line must read the resolved OFF: %s", line)
	}
	if strings.Contains(GeometryRefBootLine(nil), "geom_ref_ids=off") {
		t.Fatal("nil config must read on(default)")
	}
}
