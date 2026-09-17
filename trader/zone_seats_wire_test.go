package trader

import (
	"strings"
	"testing"
	"time"

	"nofx/kernel"
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
	want := "🗺 zone-seats @NY: zones=3 in_band=2 out_of_band=1 injected=0 aliased=0 no_source=0 → merged injected=1 aliased=1 [ZONE-4H-DEMAND@19980.00 ZONE-1H-SUPPLY@20040.00]"
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
