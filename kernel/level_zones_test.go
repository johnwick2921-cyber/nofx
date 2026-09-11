package kernel

import (
	"encoding/json"
	"nofx/market"
	"os"
	"strings"
	"testing"
	"time"
)

// One frozen NY clock; these are actual detector outputs, not reconstructed chart labels.
func TestLevelZonesOwnerSnapshot(t *testing.T) {
	b, err := os.ReadFile("../docs/superpowers/reports/2026-09-11-level-zones-evidence/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Now        string  `json:"now"`
		ATR        float64 `json:"atr5m"`
		Price      float64 `json:"price"`
		Candidates []struct {
			ID     int           `json:"id"`
			Origin DetectedLevel `json:"raw_origin"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	now, err := time.Parse(time.RFC3339Nano, fixture.Now)
	if err != nil {
		t.Fatal(err)
	}
	var raw []DetectedLevel
	for _, r := range fixture.Candidates {
		raw = append(raw, r.Origin)
	}
	before, _ := json.Marshal(raw)
	view := BuildLevelZones(raw, fixture.Price, fixture.ATR, nil, DefaultZoneOptions(12), now)
	after, _ := json.Marshal(raw)
	if string(before) != string(after) {
		t.Fatal("render changed detector anchors/bounds")
	}
	if view.Broad != 164 || len(view.Zones) != 222 || view.Merged != 44 {
		t.Fatalf("native-bound census: %+v", view.Counts())
	}
	if view.WidestMerged > fixture.ATR || view.WidestMerged != 41.5 {
		t.Fatalf("width %g", view.WidestMerged)
	}
	var pair, daily bool
	for _, z := range view.Zones {
		a, b := false, false
		for _, s := range z.Sources {
			if s.Price == 29475 && s.Kind == KindSWGH {
				a = a || s.TF == "5m"
				b = b || s.TF == "15m"
			}
			if s.Price == 29006.625 && s.Kind == KindDemand {
				daily = z.Broad && len(z.Sources) == 1
			}
		}
		pair = pair || a && b
	}
	if !pair || !daily {
		t.Fatalf("pair=%v daily separate=%v", pair, daily)
	}
	prompt := BuildPlannerPrompt(PlannerInput{Now: now, TradeDate: "2026-09-11", Session: "NY", Price: fixture.Price, Zones: &view})
	for _, name := range []string{"SWG-H·5m", "SWG-H·15m", "Demand·1d", "28810.75", "29202.50"} {
		if !strings.Contains(prompt, name) {
			t.Errorf("model table lost %s", name)
		}
	}
}

func TestLevelZonesRankingIgnoresHTFAndKeepsAllReferences(t *testing.T) {
	now := time.Date(2026, 9, 11, 15, 30, 0, 0, time.UTC)
	raw := []DetectedLevel{{Kind: KindSWGL, TF: "5m", Price: 100, Lo: 100, Hi: 100, Label: "near"}, {Kind: KindSWGL, TF: "1d", HTF: true, Price: 200, Lo: 200, Hi: 200, Label: "far"}}
	count := 100
	inputs := map[string]ZoneWidthInput{ZoneSourceKey(raw[1]): {PriorTouches: &count}}
	o := DefaultZoneOptions(1)
	v := BuildLevelZones(raw, 100, 24, inputs, o, now)
	if len(v.Zones) != 2 || !v.Zones[1].Shortlisted {
		t.Fatal("touch term not connected to shortlist or loser dropped")
	}
	raw[1].HTF = false
	w := BuildLevelZones(raw, 100, 24, inputs, o, now)
	if *v.Zones[1].RankValue != *w.Zones[1].RankValue {
		t.Fatal("HTF affects order")
	}
	inputs = nil
	w = BuildLevelZones(raw, 100, 24, inputs, o, now)
	if !w.Zones[0].Shortlisted {
		t.Fatal("distance term not connected")
	}
}

func TestLevelZonesBootUsesResolvedParameters(t *testing.T) {
	t.Setenv("LEVEL_ZONE_MERGE_ATR", "0.25")
	t.Setenv("LEVEL_ZONE_MAX_WIDTH_ATR", "0.75")
	t.Setenv("LEVEL_ZONE_BROAD_ATR", "0.6")
	line := ZoneBootLine(ResolveZoneOptions(12))
	for _, s := range []string{"merge=0.25", "max-width=0.75", "broad>0.6", "cap=12", "n/a", "BuildLevelZones"} {
		if !strings.Contains(line, s) {
			t.Errorf("missing %s: %s", s, line)
		}
	}
}

func TestLevelZonesInputsExcludeFutureBarsAndRequireFormation(t *testing.T) {
	now := time.Date(2026, 9, 11, 15, 30, 0, 0, time.UTC)
	start := now.Add(-30 * time.Minute).UnixMilli()
	var bars []market.Kline
	for j := 0; j < 30; j++ {
		bars = append(bars, market.Kline{OpenTime: start + int64(j)*60_000, CloseTime: start + int64(j+1)*60_000 - 1, Open: 100, High: 102, Low: 98, Close: 100 + float64(j%2)})
	}
	wick := 2.
	birth := bars[1].CloseTime
	l := DetectedLevel{Kind: KindSWGL, TF: "1m", Price: 100, ZoneDefiningWick: &wick, FormedCloseMs: &birth}
	raw := []DetectedLevel{l}
	a := LevelZoneInputs(raw, map[string][]market.Kline{"1m": bars}, now)[ZoneSourceKey(l)]
	if a.ATR <= 0 || a.Wick == nil || a.PriorTouches == nil {
		t.Fatalf("known inputs missing: %+v", a)
	}
	bars = append(bars, market.Kline{OpenTime: now.UnixMilli(), CloseTime: now.Add(time.Minute).UnixMilli(), Open: 100, High: 1000, Low: 1, Close: 800})
	b := LevelZoneInputs(raw, map[string][]market.Kline{"1m": bars}, now)[ZoneSourceKey(l)]
	if a.ATR != b.ATR || *a.PriorTouches != *b.PriorTouches {
		t.Fatal("future bar entered read")
	}
	l.FormedCloseMs = nil
	c := LevelZoneInputs([]DetectedLevel{l}, map[string][]market.Kline{"1m": bars}, now)[ZoneSourceKey(l)]
	if c.PriorTouches != nil {
		t.Fatal("unknown birth fabricated a count")
	}
}

func TestLevelZonesCompatibilityAndNoChain(t *testing.T) {
	now := time.Date(2026, 9, 11, 15, 30, 0, 0, time.UTC) // 10:30 NY session, CT
	raw := []DetectedLevel{{Kind: KindEQH, Price: 100, Lo: 100, Hi: 100, Label: "a"}, {Kind: KindEQH, Price: 112, Lo: 112, Hi: 112, Label: "b"}, {Kind: KindEQH, Price: 124, Lo: 124, Hi: 124, Label: "c"}}
	opts := DefaultZoneOptions(12)
	v := BuildLevelZones(raw, 110, 24, nil, opts, now)
	if len(v.Zones) != 2 {
		t.Fatalf("non-transitive clusters=%d", len(v.Zones))
	}
	opts.MergeATR = .25
	if v = BuildLevelZones(raw, 110, 24, nil, opts, now); len(v.Zones) != 3 {
		t.Fatal("m=.25 must not merge 12pt apart")
	}
	opts = DefaultZoneOptions(12)
	opts.MaxWidthATR = .25
	if v = BuildLevelZones(raw, 110, 24, nil, opts, now); len(v.Zones) != 3 {
		t.Fatal("width cap ignored")
	}
}

func TestLevelZonesWidthFamilyAndUnknown(t *testing.T) {
	now := time.Date(2026, 9, 11, 15, 30, 0, 0, time.UTC)
	raw := []DetectedLevel{{Kind: KindSWGL, TF: "1d", Price: 100, Lo: 100, Hi: 100, Label: "daily"}, {Kind: KindSWGL, TF: "4h", Price: 100, Lo: 100, Hi: 100, Label: "4h"}, {Kind: KindSWGL, TF: "1h", Price: 100, Lo: 100, Hi: 100, Label: "1h"}, {Kind: KindPOC, Price: 100, Lo: 100, Hi: 100, Label: "POC"}, {Kind: KindRound, Price: 100, Lo: 100, Hi: 100, Label: "round"}, {Kind: KindPDH, Price: 100, Lo: 100, Hi: 100, Label: "PDH"}, {Kind: KindOB, Price: 100, Lo: 98, Hi: 102, Label: "OB"}}
	v := BuildLevelZones(raw, 100, 24, nil, DefaultZoneOptions(12), now)
	if len(v.Zones) != 1 || v.Zones[0].FamilyCount != 3 || len(v.Zones[0].Families) != 5 || len(v.Zones[0].Sources) != 7 {
		t.Fatalf("families/names: %+v", v.Zones)
	}
	if v.NullWidths != 5 {
		t.Fatalf("missing wick/ATR must be NULL, got %d", v.NullWidths)
	}
	if *v.Zones[0].Lo != 98 || *v.Zones[0].Hi != 102 {
		t.Fatal("native OB bounds lost")
	}
	w := 4.0
	widths := map[string]ZoneWidthInput{ZoneSourceKey(raw[0]): {ATR: 20, Wick: &w}}
	v = BuildLevelZones(raw[:1], 100, 24, widths, DefaultZoneOptions(12), now)
	if *v.Zones[0].Lo != 95 || *v.Zones[0].Hi != 105 {
		t.Fatal("max(wick,k*ATR)/2 not applied")
	}
}
