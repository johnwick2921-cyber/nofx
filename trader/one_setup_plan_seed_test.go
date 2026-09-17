package trader

import (
	"encoding/json"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/levelidentity"
	"nofx/market"
	"nofx/store"
)

// ── REVIEW FIX 1 (W-STRUCTURE-ZONE-SEATS): one-setup judges against the map the
// planner authored on ────────────────────────────────────────────────────────
//
// BEFORE (measured 2026-09-17 17:09 CT on the golden fixture's REAL path): the
// plan's own A-grade seated level PDL@29490 was absent from one-setup's map —
// the live re-assembly runs the 1m detectors only — so the best was a narrow-
// pool ONL@29494 (B) and S1/S2/S3 were ALL level_not_best; zero arms,
// counter one_setup:level=2. The planner authored on a table one-setup never
// saw. Seeding the plan's own levels closes it by construction.

// oneSetupSeededFixtureDoc is the golden fixture doc with S1's level RELABELLED
// (label, kind, tf, band), re-identified (the id is recomputed from the new
// identity inputs and the scenario's level_id follows it) and machine-graded A
// — everything else byte-equal, so the arm that results is the golden's S1 row.
func oneSetupSeededFixtureDoc(t *testing.T, label, kind, tf string, lo, hi float64) string {
	t.Helper()
	var d kernel.PlanDoc
	if err := json.Unmarshal([]byte(oneSetupFixtureDoc()), &d); err != nil {
		t.Fatal(err)
	}
	relabel := func(l *kernel.PlanLevel) *string {
		k, tfv, lov, hiv := kind, tf, lo, hi
		l.Label, l.Kind, l.TF, l.Lo, l.Hi = label, &k, &tfv, &lov, &hiv
		l.ID, _ = levelidentity.ID(kernel.IdentityInputs(*l))
		return l.ID
	}
	var oldID, newID *string
	for i := range d.Levels {
		if d.Levels[i].Price == 29490 {
			oldID = d.Levels[i].ID
			newID = relabel(&d.Levels[i])
			d.Levels[i].MachineGrade = "A"
		}
	}
	for i := range d.IdentityLevels {
		if d.IdentityLevels[i].Price == 29490 {
			relabel(&d.IdentityLevels[i])
		}
	}
	for i := range d.Scenarios {
		if d.Scenarios[i].LevelID != nil && oldID != nil && *d.Scenarios[i].LevelID == *oldID {
			d.Scenarios[i].LevelID = newID
		}
	}
	// the structural-stop provenance (the doc's zone map) follows the label too
	if d.Zones != nil {
		for zi := range d.Zones.Zones {
			for si := range d.Zones.Zones[zi].Sources {
				if d.Zones.Zones[zi].Sources[si].Price == 29490 {
					d.Zones.Zones[zi].Sources[si].Label, d.Zones.Zones[zi].Sources[si].TF = label, tf
				}
			}
		}
	}
	d.Scenarios[0].Trigger = "reject at " + label + " 29490"
	blob, err := json.Marshal(&d)
	if err != nil {
		t.Fatal(err)
	}
	return string(blob)
}

// oneSetupStampedFixtureDoc is the golden fixture doc with PDL@29490 carrying
// the write site's machine grade (A) — the seated, stamped level.
func oneSetupStampedFixtureDoc(t *testing.T) string {
	t.Helper()
	var d kernel.PlanDoc
	if err := json.Unmarshal([]byte(oneSetupFixtureDoc()), &d); err != nil {
		t.Fatal(err)
	}
	for i := range d.Levels {
		if d.Levels[i].Price == 29490 {
			d.Levels[i].MachineGrade = "A"
		}
	}
	blob, err := json.Marshal(&d)
	if err != nil {
		t.Fatal(err)
	}
	return string(blob)
}

// reachBand widens the strategy's own proximity_filter_atr to its clamp (3.0):
// the golden tape is flat (dATR 2.00), so the default band (1.5 × 2 = 3 pt)
// cannot reach the plan's levels at ±5 — one-setup's reachability band is the
// same knob the planner's band reads, never a test-only number.
func reachBand(c *store.StrategyConfig) { c.DayPlan.ProximityFilterATR = 3.0 }

func oneSetupArmSiteSeatArms(t *testing.T, label, kind, tf string, lo, hi float64) {
	t.Helper()
	oneSetupFixtureDocOverride = oneSetupSeededFixtureDoc(t, label, kind, tf, lo, hi)
	g, at, st, _ := driveOneSetupArmPath(t, reachBand, nil)
	rec := findOneSetupRecord(t, st, at.id)
	if rec == nil {
		t.Fatal("no one-setup record written")
	}
	s1 := rec.Scenarios["S1"]
	t.Logf("arm site %s: S1 allowed=%v level=%q best=%q(%s) rows=%d counters=%v", label, s1.Allowed, s1.Level, s1.BestNames, s1.BestGrade, len(g.Rows), g.Counters)
	if !s1.Allowed || s1.Level != "ok" || !strings.Contains(s1.BestNames, label) {
		t.Fatalf("S1 on a %s seat must be ALLOWED with that level best: %+v", label, s1)
	}
	armed := false
	for _, r := range g.Rows {
		if r["scenario"] == "S1" && r["entry"] == 29490.0 && r["state"] == "armed" {
			armed = true
		}
	}
	if !armed {
		t.Fatalf("S1 must reach the ledger armed at 29490: %+v", g.Rows)
	}
}

// A scenario authored on a ZONE-4H-DEMAND seat arms through the REAL arm path
// (maybeManageArmedOrdersAt → oneSetupVerdictsAt → oneSetupConsult → ledger).
func TestOneSetupArmSiteZoneSeatScenarioArms(t *testing.T) {
	oneSetupArmSiteSeatArms(t, "ZONE-4H-DEMAND", "DEMAND", "4h", 29450, 29490)
}

// The pre-existing gap: a scenario on a plain DAILY HTF seat (never in the 1m
// pool) arms the same way.
func TestOneSetupArmSiteDailyHTFSeatScenarioArms(t *testing.T) {
	oneSetupArmSiteSeatArms(t, "Demand·1d", "DEMAND", "1d", 29450, 29490)
}

// THE OFF-PATH PIN, both halves recorded. (a) The pinned golden fixture at its
// default band: dATR 2.00 → band 3 pt → the plan's levels at ±5 are outside
// it and only the live ONL@29494 (B) is reachable — NO decision changes (S1/S2
// level_not_best exactly as before the fix, zero arms). (b) With a band that
// reaches the plan's levels (proximity 3.0 → 6 pt) S1 — the plan's own PDL,
// authored A, absent from the live pool — flips from level_not_best to
// ALLOWED and arms: the fix of the defect, not a regression.
func TestOneSetupFixtureDecisionAfterSeeding(t *testing.T) {
	log := func(tag string, g armPathGolden, rec *store.OneSetupRecord) {
		for id, s := range rec.Scenarios {
			t.Logf("%s %s allowed=%v level=%q best=%q(%s) play=%s perm=%s", tag, id, s.Allowed, s.Level, s.BestNames, s.BestGrade, s.Play, s.Permission)
		}
		t.Logf("%s rows=%+v counters=%v", tag, g.Rows, g.Counters)
	}
	g, at, st, _ := driveOneSetupArmPath(t, nil, nil)
	rec := findOneSetupRecord(t, st, at.id)
	if rec == nil {
		t.Fatal("no record")
	}
	log("DEFAULT-BAND", g, rec)
	if rec.Scenarios["S1"].Allowed || !strings.HasPrefix(rec.Scenarios["S1"].Level, "level_not_best:ONL") || len(g.Rows) != 0 {
		t.Fatalf("default band (3 pt): the plan's levels stay unreachable, decision unchanged from before the fix: %+v rows=%v", rec.Scenarios["S1"], g.Rows)
	}
	// (b1) reach band, fixture UNSTAMPED (MachineGrade ""): the plan's PDL/ONH
	// may only alias; PDL@29490 has no live row within 3 pt → dropped, S1 stays
	// declined (bounded seeding, second re-review).
	g1, at1, st1, _ := driveOneSetupArmPath(t, reachBand, nil)
	rec1 := findOneSetupRecord(t, st1, at1.id)
	if rec1 == nil {
		t.Fatal("no record (band 6, unstamped)")
	}
	log("REACH-BAND-UNSTAMPED", g1, rec1)
	if rec1.Scenarios["S1"].Allowed || rec1.Scenarios["S1"].Level == "ok" {
		t.Fatalf("band 6 pt, unstamped PDL: S1 must NOT be allowed on an authored grade: %+v", rec1.Scenarios["S1"])
	}
	// (b2) reach band, PDL machine-stamped A: the decision change.
	oneSetupFixtureDocOverride = oneSetupStampedFixtureDoc(t)
	g2, at2, st2, _ := driveOneSetupArmPath(t, reachBand, nil)
	rec2 := findOneSetupRecord(t, st2, at2.id)
	if rec2 == nil {
		t.Fatal("no record (band 6)")
	}
	log("REACH-BAND-STAMPED", g2, rec2)
	if !rec2.Scenarios["S1"].Allowed || !strings.Contains(rec2.Scenarios["S1"].BestNames, "PDL") {
		t.Fatalf("band 6 pt: S1 on the plan's own machine-stamped PDL must now be allowed with PDL best: %+v", rec2.Scenarios["S1"])
	}
	armed := false
	for _, r := range g2.Rows {
		if r["scenario"] == "S1" && r["state"] == "armed" {
			armed = true
		}
	}
	if !armed {
		t.Fatalf("band 6 pt: S1 must arm: %+v", g2.Rows)
	}
}

// Pure pins on the seeding itself.
func TestOneSetupSeedPlanLevels(t *testing.T) {
	price := 29495.0
	live := []kernel.ScoredLevel{{DetectedLevel: kernel.DetectedLevel{Kind: kernel.KindONL, Price: 29494, Lo: 29494, Hi: 29494, Label: "ONL"}, Grade: "B", Score: 1.1, Distance: -1}}
	id := structuralTestIdentity(29480, "PDL")
	id.Grade, id.MachineGrade = "B", "A"
	near := kernel.PlanLevel{Price: 29493, Label: "VWAP", Grade: "A"}
	doc := &kernel.PlanDoc{Levels: []kernel.PlanLevel{id, near, {Price: 0, Label: "junk"}}}
	seeded, dropped := oneSetupSeedPlanLevels(live, doc, price)
	if len(seeded) != 3 || len(live) != 1 || dropped != 0 {
		t.Fatalf("two seeded rows appended, input untouched: %d/%d dropped=%d", len(seeded), len(live), dropped)
	}
	cs := kernel.BuildMapCandidates(seeded, price, 5, kernel.MapCandidateOpts{})
	// the near doc level merges INTO the live keeper (zero score) and lends its name
	mc, ok := kernel.MatchMapCandidate(cs, 29494)
	if !ok || mc.Grade != "B" || len(mc.Names) != 2 || mc.Names[0] != "ONL" || mc.Names[1] != "VWAP" {
		t.Fatalf("live row stays the keeper and gains the name: %+v", mc)
	}
	// the absent level stands alone, MACHINE grade over the authored grade, own id
	pd, ok := kernel.MatchMapCandidate(cs, 29480)
	if !ok || pd.Grade != "A" || pd.Names[0] != "PDL" {
		t.Fatalf("plan level stands as its own candidate with the machine grade: %+v", pd)
	}
	if pd.ID == nil || id.ID == nil || *pd.ID != *id.ID {
		t.Fatalf("the seeded candidate must carry the doc level's own id: got %v want %v", pd.ID, id.ID)
	}
	if got, d := oneSetupSeedPlanLevels(live, nil, price); len(got) != 1 || d != 0 {
		t.Fatal("nil doc → live rows only")
	}
	// BOUNDED: an UNSTAMPED level with no live row within the merge width is
	// dropped and counted, never a candidate on its authored grade; an unstamped
	// level within the width aliases and the live grade wins (`near` above: the
	// merged candidate's grade stayed the live B, not the authored A).
	unstamped := &kernel.PlanDoc{Levels: []kernel.PlanLevel{{Price: 29480, Label: "PDL", Grade: "A"}}}
	got, d := oneSetupSeedPlanLevels(live, unstamped, price)
	if len(got) != 1 || d != 1 {
		t.Fatalf("unstamped solo level must be dropped and counted: rows=%d dropped=%d", len(got), d)
	}
	if _, ok := kernel.MatchMapCandidate(kernel.BuildMapCandidates(got, price, 5, kernel.MapCandidateOpts{}), 29480); ok {
		t.Fatal("an unstamped solo level must not be a candidate")
	}
}

// Through the arm seam: a scenario on an UNSTAMPED solo plan level is declined
// level_no_candidate (min grade A leaves no live candidate) and the drop is
// COUNTED as one_setup:seed_unstamped_dropped.
func TestOneSetupUnstampedSoloLevelIsNotACandidate(t *testing.T) {
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true, OneSetupMinGrade: "A"}}
	at, st := resetTrader(t, cfg)
	now := time.Now()
	bars := oneSetupFixtureTape(now)
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(symbol string, tf string, n int) []market.Kline { return bars }
	t.Cleanup(func() { market.FuturesBarsProvider = prev })
	var doc kernel.PlanDoc
	if err := json.Unmarshal([]byte(oneSetupFixtureDoc()), &doc); err != nil {
		t.Fatal(err)
	}
	for i := range doc.Levels {
		if doc.Levels[i].MachineGrade != "" {
			t.Fatal("fixture levels must be unstamped for this pin")
		}
	}
	plan := &kernel.ActivePlan{PlanID: "unstamped", Version: 1, Session: "NY"}
	c := at.oneSetupVerdictsAt(plan, &doc, bars, 1, &cfg, now)
	v := c.verdicts["S1"]
	t.Logf("unstamped: S1 level=%q best=%q", v.Level, v.BestNames)
	if v.Level != "level_no_candidate" {
		t.Fatalf("S1 on an unstamped solo level must be level_no_candidate: %+v", v)
	}
	var kv []struct{ Key, Value string }
	if err := st.GormDB().Raw("SELECT key, value FROM system_config WHERE key LIKE ?", "arm_refusals_0b:"+at.id+":%"+store.OneSetupClassSeedUnstamped).Scan(&kv).Error; err != nil {
		t.Fatal(err)
	}
	if len(kv) != 1 || kv[0].Value != "2" {
		t.Fatalf("both unstamped solo levels (PDL 29490, ONH 29500) must be counted dropped: %+v", kv)
	}
	// DEBOUNCE: the same plan/version/dropped set on the next cycle counts
	// nothing (2, not 4); a new version with a different dropped set counts
	// again (its one dropped level → 3).
	count := func() string {
		var rows []struct{ Key, Value string }
		if err := st.GormDB().Raw("SELECT key, value FROM system_config WHERE key LIKE ?", "arm_refusals_0b:"+at.id+":%"+store.OneSetupClassSeedUnstamped).Scan(&rows).Error; err != nil || len(rows) != 1 {
			t.Fatalf("counter read: %v %+v", err, rows)
		}
		return rows[0].Value
	}
	_ = at.oneSetupVerdictsAt(plan, &doc, bars, 1, &cfg, now.Add(time.Minute))
	if got := count(); got != "2" {
		t.Fatalf("same plan/version/set on the next cycle must not count again: %s", got)
	}
	plan2 := &kernel.ActivePlan{PlanID: "unstamped", Version: 2, Session: "NY"}
	doc2 := doc
	doc2.Levels = []kernel.PlanLevel{doc.Levels[0]} // one unstamped solo level
	_ = at.oneSetupVerdictsAt(plan2, &doc2, bars, 1, &cfg, now.Add(2*time.Minute))
	if got := count(); got != "3" {
		t.Fatalf("a new version with a different dropped set counts again (2+1): %s", got)
	}
	_ = at.oneSetupVerdictsAt(plan2, &doc2, bars, 1, &cfg, now.Add(3*time.Minute))
	if got := count(); got != "3" {
		t.Fatalf("debounced on v2 too: %s", got)
	}
}

// ── REVIEW FIX 2: a ZONE-* clone resolves freshness under its SOURCE row ─────
func TestZoneSeatCloneFreshnessKeyedOnSourceRow(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "fresh.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	at := &AutoTrader{id: "fresh-t", exchange: "ninjatrader", store: st, config: AutoTraderConfig{NinjaTraderSymbol: "MNQ", StrategyConfig: &store.StrategyConfig{DayPlan: &store.DayPlanConfig{}}}}
	prev := kernel.LevelStateProvider
	installLevelStateProvider(at, st)
	t.Cleanup(func() { kernel.LevelStateProvider = prev })

	price, dATR := 20000.0, 100.0
	src := zoneLevelForTest(19940, 19980, "4h") // mid 19960, edge 19980 → different 1.25-pt bins
	now := time.Now()
	// mark the SOURCE zone consumed in level state under the W7 key
	key := store.MakeLevelKey(at.id, "MNQ", kernel.LevelTypeFromLabel(src.Label), "", kernel.LevelBinIndex(src.Price))
	if err := st.LevelState().EnsureLevel(&store.LevelStateDB{LevelKey: key, TraderID: at.id, Symbol: "MNQ", LevelType: kernel.LevelTypeFromLabel(src.Label), BinIndex: kernel.LevelBinIndex(src.Price), Price: src.Price}); err != nil {
		t.Fatal(err)
	}
	if err := st.LevelState().MarkConsumed(key, now.UnixMilli()); err != nil {
		t.Fatal(err)
	}
	scoredSrc := kernel.ScoreLevels([]kernel.DetectedLevel{src}, price, dATR, nil, 1, 2)
	m := &kernel.StructureMap{TFs: map[string]kernel.StructureTF{"4h": {Zones: []kernel.StructureZone{{Kind: string(src.Kind), Lo: src.Lo, Hi: src.Hi, TF: "4h"}}}}}
	cands, _ := kernel.ZoneSeatCandidates(m, scoredSrc, price, 2*dATR)
	if len(cands) != 1 || cands[0].StateKeyLabel != src.Label || cands[0].StateKeyPrice != src.Price {
		t.Fatalf("clone must carry the source's state key: %+v", cands)
	}
	if kernel.LevelBinIndex(cands[0].Price) == kernel.LevelBinIndex(src.Price) {
		t.Fatal("fixture: edge and midpoint must fall in different price bins")
	}
	gotSrc := kernel.LevelStateProvider(at.id, "MNQ", src, now)
	gotClone := kernel.LevelStateProvider(at.id, "MNQ", cands[0], now)
	if gotSrc != store.FreshnessDone || gotClone != gotSrc {
		t.Fatalf("clone freshness must equal the source's: src=%q clone=%q", gotSrc, gotClone)
	}
	// through the scorer: both are cut as consumed (zoneFreshMult done=0.15 → seated flipped, never fresh)
	fn := func(l kernel.DetectedLevel) string { return kernel.LevelStateProvider(at.id, "MNQ", l, now) }
	for _, s := range kernel.ScoreLevels(append([]kernel.DetectedLevel{src}, cands...), price, dATR, fn, 12, 2) {
		if s.Fresh != "flipped" {
			t.Fatalf("%s must grade as the consumed source does (flipped), got %q", s.Label, s.Fresh)
		}
	}
	// a clone without a key (no source row) is its own identity: fresh
	noKey := cands[0]
	noKey.StateKeyLabel, noKey.StateKeyPrice = "", 0
	if got := kernel.LevelStateProvider(at.id, "MNQ", noKey, now); got != "" {
		t.Fatalf("keyless clone reads its own (absent) state: %q", got)
	}
	_ = math.Abs
	_ = market.Kline{}
}

func zoneLevelForTest(lo, hi float64, tf string) kernel.DetectedLevel {
	return kernel.DetectedLevel{Kind: kernel.KindDemand, Price: (lo + hi) / 2, Lo: lo, Hi: hi, Label: "Demand·" + tf, HTF: true, TF: tf, ZonePattern: "reversal", OriginDate: "2030-09-11"}
}
