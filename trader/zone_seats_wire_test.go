package trader

import (
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

// W-STRUCTURE-ZONE-SEATS — the boot line is READ from the resolver: off unless
// the strategy saved structure_zone_seats=true; nil config is off, not a panic.
func TestZoneSeatsBootLineReadsResolver(t *testing.T) {
	if got := ZoneSeatsBootLine(nil); got != "🗺 zone-seats=off(default) (W-STRUCTURE-ZONE-SEATS)" {
		t.Fatalf("nil config: %q", got)
	}
	if got := ZoneSeatsBootLine(&store.DayPlanConfig{}); !strings.Contains(got, "zone-seats=off(default)") {
		t.Fatalf("unset knob: %q", got)
	}
	if got := ZoneSeatsBootLine(&store.DayPlanConfig{StructureZoneSeats: true}); got != "🗺 zone-seats=on(saved) (W-STRUCTURE-ZONE-SEATS)" {
		t.Fatalf("saved on: %q", got)
	}
}

// The read line prints what the pass RECORDED, never a literal.
func TestZoneSeatsReadLineRecordsCounters(t *testing.T) {
	rep := kernel.ZoneSeatReport{Zones: 3, InBand: 2, OutOfBand: 1, Candidates: []string{"ZONE-4H-DEMAND@19980.00", "ZONE-1H-SUPPLY@20040.00"}}
	got := zoneSeatsReadLine("NY", rep, 1, 1)
	want := "🗺 zone-seats @NY: zones=3 in_band=2 out_of_band=1 duplicates=0 no_source=0 → merged injected=1 aliased=1 [ZONE-4H-DEMAND@19980.00 ZONE-1H-SUPPLY@20040.00]"
	if got != want {
		t.Fatalf("read line:\n got %q\nwant %q", got, want)
	}
}

// The resolver is nil-safe on the trader and reads the bound strategy.
func TestZoneSeatsEnabledReadsBoundStrategy(t *testing.T) {
	var nilAT *AutoTrader
	if nilAT.zoneSeatsEnabled() {
		t.Fatal("nil trader must read off")
	}
	at := &AutoTrader{}
	if at.zoneSeatsEnabled() {
		t.Fatal("no strategy must read off")
	}
	at.config.StrategyConfig = &store.StrategyConfig{DayPlan: &store.DayPlanConfig{StructureZoneSeats: true}}
	if !at.zoneSeatsEnabled() {
		t.Fatal("saved on must read on")
	}
	// knob off → the candidate pass returns nothing without touching bars
	at.config.StrategyConfig.DayPlan.StructureZoneSeats = false
	if c, m, _ := at.zoneSeatCandidatesForRead("MNQ", "MNQ 12-26", []kernel.DetectedLevel{{Kind: kernel.KindDemand}}, nil, time.Now()); c != nil || m != nil {
		t.Fatal("knob off must return no candidates and no map")
	}
}

// THE PRODUCTION CALL SITE — assemblePlannerInput on a stub tape (the harness
// TestPlannerReadWritesDetectorStores uses), log captured (CLASS 142 sync buf).
// OFF: no zone-seats line, no ZONE row. ON: exactly one read line per read, the
// STRUCTURE section stays gated by ITS knob (Structure nil), every ZONE seat is
// inside the band, and the line's `merged injected` equals the ZONE rows the
// read's pool actually carries (recorded, never inferred). ON + structure_map:
// the section gets the pass's map (non-nil).
func TestZoneSeatsAtThePlannerReadCallSite(t *testing.T) {
	st, err := store.New(filepath.Join(t.TempDir(), "zoneseats.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	now := time.Now()
	bars := oscillatingTape(29141.25, now.Add(-700*time.Minute), 700)
	prev := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline { return bars }
	t.Cleanup(func() { market.FuturesBarsProvider = prev })
	logBuf := captureTraderLog(t)

	dp := &store.DayPlanConfig{PlanEnabled: true, PlannerTimeframes: []string{"D", "4h", "1h", "15m"}}
	at := &AutoTrader{
		id: "zone-seats", exchange: "ninjatrader", store: st,
		config: AutoTraderConfig{NinjaTraderSymbol: "MNQ", StrategyConfig: &store.StrategyConfig{DayPlan: dp}},
	}
	tradeDate := kernel.CMESessionDayStart(now).Format("2006-01-02")
	zoneRows := func(ls []kernel.ScoredLevel) int {
		n := 0
		for _, l := range ls {
			if kernel.IsZoneSeatLabel(l.Label) {
				n++
			}
		}
		return n
	}

	// OFF
	off := at.assemblePlannerInput("NY", tradeDate)
	if strings.Contains(logBuf.String(), "zone-seats @") {
		t.Fatal("knob off must not log a zone-seats read line")
	}
	if zoneRows(off.Levels) != 0 || zoneRows(off.Pool) != 0 {
		t.Fatal("knob off must seat no ZONE row")
	}

	// ON (structure_map unset → the prompt section stays off)
	logBuf.Reset()
	dp.StructureZoneSeats = true
	on := at.assemblePlannerInput("NY", tradeDate)
	lines := strings.Count(logBuf.String(), "🗺 zone-seats @NY:")
	if lines != 1 {
		t.Fatalf("knob on must log exactly one zone-seats read line, got %d:\n%s", lines, logBuf.String())
	}
	if on.Structure != nil {
		t.Fatal("the STRUCTURE section is gated by day_plan.structure_map, not by zone seats")
	}
	line := logBuf.String()[strings.Index(logBuf.String(), "🗺 zone-seats @NY:"):]
	line = strings.SplitN(line, "\n", 2)[0]
	var recInj, recAl int
	if _, err := fmt.Sscanf(line[strings.Index(line, "→ merged"):], "→ merged injected=%d aliased=%d", &recInj, &recAl); err != nil {
		t.Fatalf("read line format: %q (%v)", line, err)
	}
	if got := zoneRows(on.Pool); got != recInj {
		t.Fatalf("read line says injected=%d but the pool carries %d ZONE rows", recInj, got)
	}
	band := at.proximityFilterATR() * on.DATR
	for _, l := range on.Levels {
		if kernel.IsZoneSeatLabel(l.Label) && math.Abs(l.Price-on.Price) > band {
			t.Fatalf("ZONE seat %s@%.2f outside the band ±%.1f of %.2f", l.Label, l.Price, band, on.Price)
		}
	}
	if len(on.Levels) > kernel.DefaultMaxLevels {
		t.Fatalf("the cap is unchanged: %d seats", len(on.Levels))
	}
	if recInj == 0 && !reflect.DeepEqual(stripResearch(off.Levels), stripResearch(on.Levels)) {
		t.Fatal("with nothing injected the ON table must equal the OFF table")
	}
	t.Logf("call site: %s", line)

	// ON + structure_map → the section reuses the pass's map
	logBuf.Reset()
	sm := true
	dp.StructureMap = &sm
	both := at.assemblePlannerInput("NY", tradeDate)
	if both.Structure == nil {
		t.Fatal("structure_map on: the read must carry the map the zone-seat pass computed")
	}
	if strings.Count(logBuf.String(), "🗺 zone-seats @NY:") != 1 || !strings.Contains(logBuf.String(), "🗺 structure @NY:") {
		t.Fatalf("both knobs on: one zone-seats line and the structure line expected:\n%s", logBuf.String())
	}
}

// stripResearch drops the record-only provenance pointer (fresh per read) so
// two reads' tables compare on what the planner is shown.
func stripResearch(ls []kernel.ScoredLevel) []kernel.ScoredLevel {
	out := make([]kernel.ScoredLevel, len(ls))
	for i, l := range ls {
		l.Research = nil
		out[i] = l
	}
	return out
}
