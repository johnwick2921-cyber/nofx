package kernel

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"nofx/market"
)

// ── W-STRUCTURE-ZONE-SEATS — pins ────────────────────────────────────────────
//
// The production seat path is AssembleResearchLevels (planner read, knob OFF)
// and AssembleResearchLevelsZoneSeats (knob ON). Both run here on the SAME
// bars the identity golden pins (identity_output_parity_test.go), so "OFF is
// byte-identical" is asserted against origin/dev's own function on a fixture
// whose output is already frozen in testdata.

// zoneSeatBars is the identity-golden tape: 48h of 1m bars around 20000
// (1,023 raw levels — a dense map, ~2pt apart in band).
func zoneSeatBars(now time.Time) []market.Kline { return zoneSeatTape(now, 60, 20, 13, 71) }

// zoneSeatSparseBars is a slower, quieter tape (61 raw levels): dense enough
// to fill 12 seats, sparse enough that a zone edge can stand 3pt clear of every
// detector row INSIDE the 12-nearest cut — the injection case needs both.
func zoneSeatSparseBars(now time.Time) []market.Kline { return zoneSeatTape(now, 30, 10, 41, 199) }

func zoneSeatTape(now time.Time, amp, amp2, f1, f2 float64) []market.Kline {
	start := now.Add(-48 * time.Hour)
	bars := make([]market.Kline, 48*60)
	for i := range bars {
		s := start.Add(time.Duration(i) * time.Minute)
		p := 20000 + amp*math.Sin(float64(i)/f1) + amp2*math.Cos(float64(i)/f2)
		bars[i] = market.Kline{OpenTime: s.UnixMilli(), CloseTime: s.Add(time.Minute).UnixMilli() - 1, Open: p - 1, High: p + 3, Low: p - 3, Close: p + 1, Volume: float64(100 + i%31)}
	}
	return bars
}

func zoneSeatNow() time.Time { return time.Date(2030, 9, 12, 12, 0, 0, 0, CTLocation()) }

// zoneSeatFreeEdge finds the NEAREST edge price below `price` that no raw
// detector level sits within `clear` points of (clear > the 3.00pt cluster
// tolerance), so the candidate stands as its own row, deterministically. It
// must be near: the production seat rule is "top-2×max by priority/score, then
// the max NEAREST" (ScoreLevelsMinGradeFullSeats), so an edge farther than the
// 12th-nearest pooled row never seats however strong — the same rule every
// level obeys.
func zoneSeatFreeEdge(t *testing.T, raw []DetectedLevel, price, band, clear float64) float64 {
	t.Helper()
	for d := 3.5; d < band-12; d += 0.25 {
		edge := price - d
		free := true
		for _, l := range raw {
			if math.Abs(l.Price-edge) < clear {
				free = false
				break
			}
		}
		if free {
			return edge
		}
	}
	t.Fatal("no free edge inside the band")
	return 0
}

// a 4h demand zone whose NEAREST edge (the high, price is above) is `edge`,
// 40 points tall — wide enough that its midpoint row and its edge row are
// distinct references (> the 3.00pt cluster tolerance).
func zoneSeatDemand4h(edge float64) DetectedLevel {
	z := zoneLevel(KindDemand, edge-40, edge, "Demand", "2030-09-11")
	z = tagHTFLevel(z, "4h", 120)
	z.ZonePattern = "reversal"
	return z
}

func zoneSeatMapFor(zones []ScoredLevel, tf string) *StructureMap {
	m := &StructureMap{AsOf: 1, Contract: "MNQ 12-30", TFs: map[string]StructureTF{}}
	m.TFs[tf] = StructureTF{Trend: "up", Bars: 50, Zones: structureZonesFor(zones, tf)}
	return m
}

// OFF = byte-identical. The knob-off call site calls AssembleResearchLevels
// (untouched signature); the knob-on function with NO candidates must produce
// the identical seats, pool, raw universe and rendered table.
func TestZoneSeatsOffIsByteIdentical(t *testing.T) {
	now := zoneSeatNow()
	bars := zoneSeatBars(now)
	extra := []DetectedLevel{zoneSeatDemand4h(19980)}
	seatedOff, poolOff, priceOff, datrOff, rawOff := AssembleResearchLevels("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "", extra...)
	seatedOn, poolOn, priceOn, datrOn, rawOn, inj, al := AssembleResearchLevelsZoneSeats("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "", nil, extra...)
	if inj != 0 || al != 0 {
		t.Fatalf("no candidates must record injected=0 aliased=0, got %d/%d", inj, al)
	}
	if priceOff != priceOn || datrOff != datrOn {
		t.Fatalf("price/dATR moved: %v/%v vs %v/%v", priceOff, datrOff, priceOn, datrOn)
	}
	if !reflect.DeepEqual(seatedOff, seatedOn) {
		t.Fatalf("seated tables differ with the knob off:\noff=%+v\non =%+v", seatedOff, seatedOn)
	}
	if !reflect.DeepEqual(poolOff, poolOn) || !reflect.DeepEqual(rawOff, rawOn) {
		t.Fatal("graded pool / raw universe differ with the knob off")
	}
	if len(seatedOff) != 12 {
		t.Fatalf("fixture must fill all 12 seats so the cap is exercised, got %d", len(seatedOff))
	}
	if a, b := RenderKeyLevelsBlock(seatedOff, priceOff), RenderKeyLevelsBlock(seatedOn, priceOn); a != b {
		t.Fatalf("rendered table differs with the knob off:\n%s\n---\n%s", a, b)
	}
	// PlannerPriceAndRange is the SAME derivation the assemble path uses.
	if p, d := PlannerPriceAndRange(bars, now); p != priceOff || d != datrOff {
		t.Fatalf("PlannerPriceAndRange %v/%v != assemble %v/%v", p, d, priceOff, datrOff)
	}
}

// ON: a 4h demand zone inside the band → its nearest edge seats as
// ZONE-4H-DEMAND under the same scorer and cap; a scenario authored on it
// passes the write-site validators; the one-setup selection can pick it.
func TestZoneSeatsOnSeatsInBandZoneAtEdgeAndValidates(t *testing.T) {
	now := zoneSeatNow()
	bars := zoneSeatSparseBars(now)
	price, dATR := PlannerPriceAndRange(bars, now)
	band := 2 * dATR
	_, _, _, _, rawOff := AssembleResearchLevels("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "")
	edge := zoneSeatFreeEdge(t, rawOff, price, band, 3.1)
	zone := zoneSeatDemand4h(edge)
	// the read's uncapped HTF-zone universe (the SAME ScoreLevels call the
	// planner makes for htfZonesFull) and the structure map drawn from it.
	src := ScoreLevels([]DetectedLevel{zone}, price, dATR, nil, 1, 2)
	if len(src) != 1 {
		t.Fatalf("zone must be in band for the map: %d", len(src))
	}
	m := zoneSeatMapFor(src, "4h")
	cands, rep := ZoneSeatCandidates(m, src, price, band)
	if len(cands) != 1 || rep.InBand != 1 || rep.OutOfBand != 0 || rep.NoSource != 0 {
		t.Fatalf("one in-band candidate expected: %d %+v", len(cands), rep)
	}
	c := cands[0]
	if c.Price != edge || c.Label != "ZONE-4H-DEMAND" || c.Kind != KindDemand || c.TF != "4h" || !c.HTF || c.ZonePattern != "reversal" || c.Lo != zone.Lo || c.Hi != zone.Hi {
		t.Fatalf("candidate must clone the detector row at the nearest edge: %+v", c)
	}
	if c.StateKeyLabel != zone.Label || c.StateKeyPrice != zone.Price {
		t.Fatalf("the clone's freshness key must be the SOURCE row's (label %q price %.2f): %+v", zone.Label, zone.Price, c)
	}
	seated, pool, _, _, _, inj, al := AssembleResearchLevelsZoneSeats("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "", cands, zone)
	if inj != 1 || al != 0 {
		t.Fatalf("the free edge must be INJECTED (1/0), got injected=%d aliased=%d", inj, al)
	}
	var seat *ScoredLevel
	for i := range seated {
		if seated[i].Label == "ZONE-4H-DEMAND" {
			seat = &seated[i]
		}
	}
	if seat == nil {
		var pl []string
		for _, p := range pool {
			pl = append(pl, fmt.Sprintf("%s@%.2f %s %.3f pri=%v", p.Label, p.Price, p.Grade, p.Score, isTodayPriority(p.Kind)))
		}
		t.Fatalf("ZONE-4H-DEMAND@%.2f did not win a seat under the same rules; seated=%s\npool=%s", edge, RenderKeyLevelsBlock(seated, price), strings.Join(pl, "\n"))
	}
	if seat.Price != edge || seat.Grade == "" {
		t.Fatalf("seat must sit at the edge with a machine grade: %+v", seat)
	}
	// the detector's own midpoint row is graded by the same grader — the zone
	// seat never invents a grade of its own.
	for _, p := range pool {
		if p.Label == "Demand·4h" && p.Grade != seat.Grade {
			t.Fatalf("zone seat grade %s differs from the detector row's %s — grading must be the detectors' own", seat.Grade, p.Grade)
		}
	}
	if !strings.Contains(RenderKeyLevelsBlock(seated, price), "ZONE-4H-DEMAND") {
		t.Fatal("the rendered table must carry the ZONE label")
	}
	if len(seated) != 12 {
		t.Fatalf("the cap is unchanged: %d seats", len(seated))
	}

	// A plan authored on the zone seat: schema, label provenance, facts.
	var above, below PlanLevel
	for _, s := range seated {
		if s.Price > price && above.Price == 0 {
			above = PlanLevel{Price: s.Price, Label: s.Label, Grade: s.Grade, Instruction: "fade"}
		}
	}
	below = PlanLevel{Price: seat.Price, Label: seat.Label, Grade: seat.Grade, Instruction: "reclaim-long"}
	doc := &PlanDoc{
		Reasoning:      "long the 4h demand edge",
		Bias:           PlanBias{Direction: "long", Conviction: "medium", FlipCondition: "2x5m below the zone"},
		Levels:         []PlanLevel{above, below},
		Scenarios:      []PlanScenario{{ID: "S1", Trigger: fmt.Sprintf("reject at ZONE-4H-DEMAND %.2f", seat.Price), Condition: "reject", Direction: "long", TargetChain: []float64{above.Price}, Invalid: fmt.Sprintf("2x5m below %.2f", seat.Price-5), Quality: "B", Confirm: &PlanConfirm{Rule: "touch", RefPrice: seat.Price, Side: "above"}}},
		DeathCondition: "acceptance below the zone",
		DayType:        "balance",
	}
	if err := ValidatePlanDocWithCaps(doc, 12, 3); err != nil {
		t.Fatalf("ValidatePlanDocWithCaps rejected a plan on a ZONE seat: %v", err)
	}
	machine := map[float64]string{}
	for _, s := range seated {
		machine[math.Round(s.Price*100)/100] = s.Label
	}
	if mis := MislabeledStructuralLevels(doc, machine); len(mis) > 0 {
		t.Fatalf("label provenance must accept ZONE-*: %v", mis)
	}
	if err := ValidatePlanDocWithFactsMachine(doc, PlanFacts{Price: price, DATR: dATR}, machine, 12, 3); err != nil {
		t.Fatalf("write-site validator rejected a plan on a ZONE seat: %v", err)
	}

	// One-setup: the seat is a merged-map candidate the predicate COMPARES —
	// resolved by price, never "level_no_candidate"/unresolved. Whether it is
	// the BEST is the map's call (grade first, distance second): on this tape
	// OR-L (A) sits 0.6pt from price and legitimately outranks it. The
	// "zone edge IS best" case is TestZoneSeatOneSetupPicksZoneSeatWhenBest.
	cs := BuildMapCandidates(seated, price, dATR/20, MapCandidateOpts{})
	mc, ok := MatchMapCandidate(cs, edge)
	if !ok || !zsHas(mc.Names, "ZONE-4H-DEMAND") || mc.Grade != seat.Grade {
		t.Fatalf("the zone seat must be a merged-map candidate with its grade: ok=%v names=%v grade=%s", ok, mc.Names, mc.Grade)
	}
	verdict := OneSetupAllowsAt(now, doc.Scenarios[0],
		OneSetupLevelFacts{Price: price, BandPts: band, Candidates: cs, Scenario: OneSetupLevelRef{Price: edge, Basis: LevelBasisPriceProximity}},
		FadeVerdict{Evaluated: true, Permitted: true}, OneSetupConfig{Enabled: true, MinGrade: "C"})
	if verdict.Play != "ok" || verdict.Permission != "ok" || (verdict.Level != "ok" && !strings.HasPrefix(verdict.Level, "level_not_best:")) {
		t.Fatalf("one-setup must compare the zone seat as a resolved candidate: %+v", verdict)
	}
}

// The one-setup selection PICKS a zone seat when it is the best merged
// candidate inside the reachability band — the production functions
// (scoreLevelsPool → BuildMapCandidates → OneSetupAllowsAt) on a pool where the
// Tier-1 anchors sit outside the band and the zone edge is the nearest row.
func TestZoneSeatOneSetupPicksZoneSeatWhenBest(t *testing.T) {
	price, dATR := 20000.0, 100.0
	zone := zoneSeatDemand4h(price - 10) // edge at −10; the band [−50,−10] holds no Tier-1 anchor → grade C by rule B2
	src := ScoreLevels([]DetectedLevel{zone}, price, dATR, nil, 1, 2)
	edge, _ := ZoneSeatCandidates(zoneSeatMapFor(src, "4h"), src, price, 2*dATR)
	if len(edge) != 1 {
		t.Fatalf("one candidate: %d", len(edge))
	}
	pool := []DetectedLevel{
		{Kind: KindPDH, Price: price + 60, Lo: price + 60, Hi: price + 60, Label: "PDH"},
		{Kind: KindPDL, Price: price - 60, Lo: price - 60, Hi: price - 60, Label: "PDL"},
		{Kind: KindRound, Price: price + 30, Lo: price + 30, Hi: price + 30, Label: "RN 20030"},
		zone,
	}
	pool = append(pool, edge...)
	seated, graded := ScoreLevelsMinGradeFullSeats(pool, price, dATR, nil, 12, 2, "", nil, HTFScoreMultiplier)
	inj, al, seatedN := ZoneSeatOutcome(graded, seated)
	if inj != 1 || al != 0 || seatedN != 1 {
		t.Fatalf("the edge must stand as its own seat: injected=%d aliased=%d seated=%d", inj, al, seatedN)
	}
	cs := BuildMapCandidates(seated, price, 5, MapCandidateOpts{})
	sc := PlanScenario{ID: "S1", Trigger: "reject at ZONE-4H-DEMAND 19990.00", Condition: "reject", Direction: "long", TargetChain: []float64{price + 30}, Invalid: "2x5m below 19985.00", Quality: "B", Confirm: &PlanConfirm{Rule: "touch", RefPrice: price - 10, Side: "above"}}
	v := OneSetupAllowsAt(time.Now(), sc,
		OneSetupLevelFacts{Price: price, BandPts: 20, Candidates: cs, Scenario: OneSetupLevelRef{Price: price - 10, Basis: LevelBasisPriceProximity}},
		FadeVerdict{Evaluated: true, Permitted: true}, OneSetupConfig{Enabled: true, MinGrade: "C"})
	if !v.Allowed || !strings.Contains(v.BestNames, "ZONE-4H-DEMAND") {
		t.Fatalf("one-setup must pick the zone seat when it is best in band: %+v", v)
	}
}

// A zone whose edge coincides with a detector level (within the cluster
// tolerance) is NOT double-seated: the detector row keeps its seat and label;
// the ZONE name rides along as a merged name.
func TestZoneSeatsCoincidentZoneIsNotDoubleSeated(t *testing.T) {
	now := zoneSeatNow()
	bars := zoneSeatBars(now)
	price, dATR := PlannerPriceAndRange(bars, now)
	seatedOff, _, _, _, _ := AssembleResearchLevels("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "")
	// take a SEATED detector row below price and put the zone's edge 1pt from it
	var host ScoredLevel
	for _, s := range seatedOff {
		if s.Price < price && !isZoneKind(s.Kind) {
			host = s
			break
		}
	}
	if host.Price == 0 {
		t.Fatal("fixture has no seated non-zone row below price")
	}
	edge := host.Price + 1
	zone := zoneSeatDemand4h(edge)
	src := ScoreLevels([]DetectedLevel{zone}, price, dATR, nil, 1, 2)
	cands, _ := ZoneSeatCandidates(zoneSeatMapFor(src, "4h"), src, price, 2*dATR)
	if len(cands) != 1 {
		t.Fatalf("one candidate: %d", len(cands))
	}
	seated, _, _, _, _, inj, al := AssembleResearchLevelsZoneSeats("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "", cands, zone)
	if inj != 0 || al != 1 {
		t.Fatalf("production path: injected=%d aliased=%d, want 0/1", inj, al)
	}
	near := 0
	var hostOn *ScoredLevel
	for i := range seated {
		if math.Abs(seated[i].Price-edge) <= clusterToleranceFor(price) {
			near++
			hostOn = &seated[i]
		}
		if seated[i].Label == "ZONE-4H-DEMAND" {
			t.Fatal("an aliased zone must not seat as its own row")
		}
	}
	if near != 1 || hostOn == nil || hostOn.Label != host.Label {
		t.Fatalf("exactly one seat near the edge, the detector's (%s): near=%d host=%+v", host.Label, near, hostOn)
	}
	if !zsHas(hostOn.CollapsedNames, "ZONE-4H-DEMAND") {
		t.Fatalf("the host row must carry the zone name: %v", hostOn.CollapsedNames)
	}
	mc, ok := MatchMapCandidate(BuildMapCandidates(seated, price, dATR/20, MapCandidateOpts{}), host.Price)
	if !ok || !zsHas(mc.Names, "ZONE-4H-DEMAND") || !zsHas(mc.Names, host.Label) {
		t.Fatalf("the merged map must show both names on one candidate: %v", mc.Names)
	}
}

// A zone outside the band contributes nothing: no candidate, and the seats are
// exactly the knob-off seats.
func TestZoneSeatsOutOfBandChangesNothing(t *testing.T) {
	now := zoneSeatNow()
	bars := zoneSeatBars(now)
	price, dATR := PlannerPriceAndRange(bars, now)
	band := 2 * dATR
	far := zoneSeatDemand4h(price - band - 50)
	// the map may still list it (a wide map universe); the candidate pass drops it
	src := []ScoredLevel{{DetectedLevel: far, Grade: "B", Fresh: "fresh", Score: 1, Distance: far.Price - price}}
	cands, rep := ZoneSeatCandidates(zoneSeatMapFor(src, "4h"), src, price, band)
	if len(cands) != 0 || rep.OutOfBand != 1 || rep.InBand != 0 {
		t.Fatalf("out-of-band zone must yield no candidate: %d %+v", len(cands), rep)
	}
	seatedOff, _, _, _, _ := AssembleResearchLevels("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "", far)
	seatedOn, _, _, _, _, inj, al := AssembleResearchLevelsZoneSeats("zone-seats", bars, DefaultSessionRegistry(), "MNQ", 12, nil, HTFScoreMultiplier, now, 2, "", cands, far)
	if inj != 0 || al != 0 || !reflect.DeepEqual(seatedOff, seatedOn) {
		t.Fatal("out-of-band zone must leave the table byte-identical")
	}
}

// Rule 4 in the collapse itself: a ZONE-* row is NEVER the keeper over a
// detector row, even when it scores higher; two ZONE-* rows within the
// tolerance fold into one (the stronger keeps); a plain zone (Demand·4h) stays
// exempt exactly as before.
func TestZoneSeatCollapseKeepsDetectorRowAndFoldsZoneEdges(t *testing.T) {
	mk := func(kind LevelKind, price, score float64, label string) ScoredLevel {
		return ScoredLevel{DetectedLevel: DetectedLevel{Kind: kind, Price: price, Lo: price, Hi: price, Label: label, TF: "4h"}, Score: score, Grade: "B", Distance: price - 1000}
	}
	in := []ScoredLevel{
		mk(KindVWAP, 980, 0.5, "VWAP"),
		mk(KindDemand, 981, 1.4, "ZONE-4H-DEMAND"),  // stronger than VWAP, within 3pt → merges INTO VWAP
		mk(KindDemand, 1050, 1.2, "ZONE-4H-DEMAND"), // stands alone
		mk(KindDemand, 1052, 0.9, "ZONE-1H-DEMAND"), // folds into the 4h edge
		mk(KindSupply, 1051, 0.7, "Supply·1h"),      // a plain zone: exempt, survives beside them
	}
	out := collapseLevelClusters(in, 3)
	byLabel := map[string]ScoredLevel{}
	for _, o := range out {
		byLabel[o.Label+"@"+trimFloat(o.Price)] = o
	}
	if len(out) != 3 {
		t.Fatalf("want 3 survivors (VWAP, ZONE-4H@1050, Supply·1h), got %d: %+v", len(out), out)
	}
	v, ok := byLabel["VWAP@980"]
	if !ok || !zsHas(v.CollapsedNames, "ZONE-4H-DEMAND") || v.Confluence != 1 {
		t.Fatalf("the detector row must keep the seat and carry the zone name: %+v", v)
	}
	z, ok := byLabel["ZONE-4H-DEMAND@1050"]
	if !ok || !zsHas(z.CollapsedNames, "ZONE-1H-DEMAND") {
		t.Fatalf("the 1h edge must fold into the 4h edge: %+v", z)
	}
	if _, ok := byLabel["Supply·1h@1051"]; !ok {
		t.Fatal("a plain zone stays exempt from collapse")
	}
	// OFF path inertness: with no ZONE-* rows the comparator key never fires.
	plain := []ScoredLevel{mk(KindVWAP, 980, 0.5, "VWAP"), mk(KindRound, 981, 1.4, "RN"), mk(KindSupply, 1051, 0.7, "Supply·1h")}
	got := collapseLevelClusters(plain, 3)
	if len(got) != 2 || got[0].Label != "RN" || got[0].Confluence != 1 || len(got[0].CollapsedNames) != 0 {
		t.Fatalf("plain collapse unchanged (same-TF duplicate carries no name): %+v", got)
	}
}

// Rule 1 — the edge nearest to price, price inside included; tie → the low.
func TestZoneNearestEdge(t *testing.T) {
	cases := []struct{ lo, hi, price, want float64 }{
		{100, 110, 120, 110}, {100, 110, 90, 100}, {100, 110, 108, 110}, {100, 110, 102, 100}, {100, 110, 105, 100},
	}
	for _, c := range cases {
		if got := zoneNearestEdge(c.lo, c.hi, c.price); got != c.want {
			t.Errorf("edge(%v,%v,%v)=%v want %v", c.lo, c.hi, c.price, got, c.want)
		}
	}
}

// Rule 5 — ZONE-* is a free label: never a structural prefix, so a plan level
// copying it is accepted; a structural label re-invented over a ZONE row is
// still rejected (the rule keeps biting where it should).
func TestZoneSeatLabelIsNotStructural(t *testing.T) {
	for _, l := range []string{"ZONE-4H-DEMAND", "ZONE-D-OB", "ZONE-1H-SUPPLY", "ZONE-4H-DEMAND · Demand·4h"} {
		if structuralPrefix(l) != "" {
			t.Fatalf("%q must not be a structural label", l)
		}
		if !IsZoneSeatLabel(l) {
			t.Fatalf("%q must be recognised as a zone seat label", l)
		}
	}
	machine := map[float64]string{19980: "ZONE-4H-DEMAND"}
	ok := &PlanDoc{Levels: []PlanLevel{{Price: 19980, Label: "ZONE-4H-DEMAND"}}}
	if mis := MislabeledStructuralLevels(ok, machine); len(mis) > 0 {
		t.Fatalf("copying the ZONE label must pass: %v", mis)
	}
	bad := &PlanDoc{Levels: []PlanLevel{{Price: 19980, Label: "PDH"}}}
	if mis := MislabeledStructuralLevels(bad, machine); len(mis) != 1 {
		t.Fatalf("a structural label over a ZONE row must still be flagged: %v", mis)
	}
	if ZoneSeatLabel("4h", KindDemand) != "ZONE-4H-DEMAND" || ZoneSeatLabel("D", KindOB) != "ZONE-D-OB" || ZoneSeatLabel("1h", KindSupply) != "ZONE-1H-SUPPLY" {
		t.Fatal("label format")
	}
}

// A zone with no detector row in the source pool is built from the map fields
// alone and COUNTED (NoSource); its pattern is unknown, never invented.
func TestZoneSeatCandidatesNoSourceIsCountedNotInvented(t *testing.T) {
	m := &StructureMap{TFs: map[string]StructureTF{"D": {Zones: []StructureZone{{Kind: "OB", Lo: 990, Hi: 1000, TF: "D"}}}}}
	cands, rep := ZoneSeatCandidates(m, nil, 1010, 100)
	if len(cands) != 1 || rep.NoSource != 1 || cands[0].ZonePattern != "" || cands[0].TF != "D" || !cands[0].HTF || cands[0].Price != 1000 || cands[0].Label != "ZONE-D-OB" || cands[0].StateKeyLabel != "" {
		t.Fatalf("no-source candidate (own identity, no state key): %+v %+v", cands, rep)
	}
	// the map may list one band twice (a periodic tape): one candidate, counted
	m.TFs["D"] = StructureTF{Zones: []StructureZone{{Kind: "OB", Lo: 990, Hi: 1000, TF: "D"}, {Kind: "OB", Lo: 990, Hi: 1000, TF: "D"}}}
	if c, r := ZoneSeatCandidates(m, nil, 1010, 100); len(c) != 1 || r.Duplicates != 1 || r.InBand != 2 || len(r.Candidates) != 1 {
		t.Fatalf("duplicate zone must yield one candidate and be counted: %d %+v", len(c), r)
	}
	if c, r := ZoneSeatCandidates(nil, nil, 1010, 100); c != nil || r.Zones != 0 {
		t.Fatal("nil map → nothing")
	}
	if c, _ := ZoneSeatCandidates(m, nil, 1010, 0); c != nil {
		t.Fatal("no band → nothing")
	}
}

func zsHas(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
