package trader

import (
	"fmt"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
	"nofx/store"
)

// ── W-STRUCTURE-ZONE-SEATS — THE PRE-SCORING PASS (2026-09-17) ──────────────
//
// Knob-gated (day_plan.structure_zone_seats, default OFF). ON: before the read
// seats its table, the structure map's D/4h/1h zones are turned into seat
// CANDIDATES (kernel/zone_seats.go — nearest edge, inside the band, alias
// within the cluster tolerance) and merged into the pool the scorer sees.
// OFF: the read calls the same AssembleResearchLevels it always did — the
// candidate pass never runs and the table is byte-identical.
//
// The map is computed here UNGATED by day_plan.structure_map: the zone seats
// must not depend on whether the prompt SECTION is rendered. The prompt
// section, its log line and the doc stamp stay gated exactly as S1 shipped
// them; when both knobs are on, the read reuses this map instead of computing
// it twice (identical inputs → identical map).

// zoneSeatsEnabled resolves the knob from THE STRATEGY BOUND TO THIS TRADER.
func (at *AutoTrader) zoneSeatsEnabled() bool {
	if at == nil || at.config.StrategyConfig == nil {
		return false
	}
	return at.config.StrategyConfig.DayPlan.StructureZoneSeatsEnabled()
}

// ZoneSeatsBootLine renders the resolved knob at trader load — READ from the
// resolver, never a literal: off unless the strategy saved
// structure_zone_seats=true. Printed beside the S3 htf line.
func ZoneSeatsBootLine(dp *store.DayPlanConfig) string {
	if dp.StructureZoneSeatsEnabled() {
		return "🗺 zone-seats=on(saved) (W-STRUCTURE-ZONE-SEATS)"
	}
	return "🗺 zone-seats=off(default) (W-STRUCTURE-ZONE-SEATS)"
}

// zoneSeatCandidatesForRead is the ONE production candidate pass. It derives
// price and the band exactly as the assemble path will (kernel.
// PlannerPriceAndRange — the same function), scores the read's HTF zones into
// the same uncapped in-band universe the structure map reads (the SAME
// ScoreLevels call as the read's htfZonesFull), computes the map, and returns
// the candidates + the map + what it recorded. Nil candidates when the knob is
// off, no bars, no zones, or no map.
func (at *AutoTrader) zoneSeatCandidatesForRead(symbol, contract string, htfLevels []kernel.DetectedLevel, bars []market.Kline, now time.Time) (cands []kernel.DetectedLevel, m *kernel.StructureMap, rep kernel.ZoneSeatReport) {
	if !at.zoneSeatsEnabled() || len(bars) == 0 || len(htfLevels) == 0 {
		return nil, nil, rep
	}
	price, dATR := kernel.PlannerPriceAndRange(bars, now)
	if price <= 0 || dATR <= 0 {
		return nil, nil, rep
	}
	zones := make([]kernel.DetectedLevel, 0, len(htfLevels))
	for _, l := range htfLevels {
		switch l.Kind {
		case kernel.KindSupply, kernel.KindDemand, kernel.KindFVG, kernel.KindOB:
			zones = append(zones, l)
		}
	}
	if len(zones) == 0 {
		return nil, nil, rep
	}
	proximityK := at.proximityFilterATR()
	zonesFull := kernel.ScoreLevels(zones, price, dATR, nil, len(zones), proximityK)
	m = at.structureMapCompute(symbol, contract, zonesFull, price, now)
	if m == nil {
		return nil, nil, rep
	}
	cands, rep = kernel.ZoneSeatCandidates(m, zonesFull, price, proximityK*dATR)
	return cands, m, rep
}

// zoneSeatsReadLine is the per-read observability line — every number is what
// the pass RECORDED (counters record, never infer).
func zoneSeatsReadLine(session string, rep kernel.ZoneSeatReport, injected, aliased int) string {
	line := fmt.Sprintf("🗺 zone-seats @%s: %s → merged injected=%d aliased=%d", session, rep.String(), injected, aliased)
	if len(rep.Candidates) > 0 {
		line += " [" + strings.Join(rep.Candidates, " ") + "]"
	}
	return line
}
