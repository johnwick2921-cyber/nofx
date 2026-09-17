package kernel

import (
	"fmt"
	"math"
	"strings"
)

// ── W-STRUCTURE-ZONE-SEATS (owner order 2026-09-17, knob OFF by default) ──────
//
// WHY. The S1 structure map (CLASS 131) renders the D/4h/1h zones into the
// planner prompt as CONTEXT ONLY. A zone never becomes a seat in the 12-seat
// table unless its detector row (priced at the zone MIDPOINT — levels.go
// zoneLevel) wins a seat on its own, so no scenario and no entry can be
// authored on a zone edge. The owner's words: "if price near that, that a good
// structure" — he wants a zone near price to be TRADABLE.
//
// ROUND 24 CAVEAT (docs/superpowers/reports/2026-09-17-structure-zone-seats.md):
// the HTF-zone-entry research found zone-at-entry does NOT predict a better
// entry. The owner enabled this anyway to test it live; Round 25 (seat
// displacement) is the measurement that validates or kills it.
//
// THE SEATING RULE (five lines):
//  1. For each structure-map zone (D, 4h, 1h — the map's own top-6 per TF),
//     the candidate price is the zone EDGE nearest to the current price
//     (price inside the zone → still the nearer edge; a tie → the low edge).
//  2. A candidate whose edge is farther than the proximity band from price is
//     dropped — nothing changes for that zone.
//  3. The candidate enters the pool BEFORE scoring as a DetectedLevel that
//     clones the zone's own detector row (kind, TF, Lo/Hi, HTF, pattern,
//     origin) with Price = the edge and Label = ZONE-<TF>-<KIND>; it is graded
//     by the SAME zone grader as every other zone on that TF.
//  4. In the cluster collapse (clusterToleranceFor, 12 ticks = 3.00pt) a ZONE-*
//     row within the tolerance of ANY other survivor merges INTO it — never the
//     other way round: the detector level keeps its seat and its label, and the
//     ZONE-<TF>-<KIND> name rides along on CollapsedNames so the merged map and
//     the card say it is a zone. Otherwise the ZONE-* row stands and competes
//     under the SAME priority rule and the SAME cap — no reserved seat, no
//     multiplier.
//  5. The validator's structural-label rule (plan_doc.go structuralLabels) does
//     not list ZONE-* — a scenario authored on a ZONE seat is a free label and
//     is never rejected as a re-invented anchor (pinned by test).

// ZoneSeatLabelPrefix is the label prefix every injected zone seat carries.
const ZoneSeatLabelPrefix = "ZONE-"

// ZoneSeatLabel renders the seat label: "ZONE-4H-DEMAND", "ZONE-D-OB",
// "ZONE-1H-SUPPLY". TF keys are the structure map's own (D | 4h | 1h),
// upper-cased; the kind is the LevelKind string (already upper-case).
func ZoneSeatLabel(tf string, kind LevelKind) string {
	return ZoneSeatLabelPrefix + strings.ToUpper(strings.TrimSpace(tf)) + "-" + strings.ToUpper(string(kind))
}

// IsZoneSeatLabel reports whether a label is an injected zone seat's.
func IsZoneSeatLabel(label string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(label)), ZoneSeatLabelPrefix)
}

// ZoneSeatReport is what one read RECORDS about the candidate pass (counters
// record, never infer). OutOfBand + InBand == Zones.
type ZoneSeatReport struct {
	Zones      int      // structure-map zones examined (all TFs)
	InBand     int      // whose nearest edge sits inside the band → candidates
	OutOfBand  int      // dropped by rule 2
	NoSource   int      // zones with no matching detector row in the source pool (graded from the zone fields alone; pattern unknown)
	Duplicates int      // in-band zones identical (kind, TF, lo, hi) to one already taken — the map may list the same band more than once; one candidate each
	Candidates []string // the distinct candidate labels with their edge price, in map order
}

// String is the per-read observability line body.
func (r ZoneSeatReport) String() string {
	return fmt.Sprintf("zones=%d in_band=%d out_of_band=%d duplicates=%d no_source=%d", r.Zones, r.InBand, r.OutOfBand, r.Duplicates, r.NoSource)
}

// zoneNearestEdge is rule 1: the edge of [lo,hi] closest to price; a tie
// returns lo.
func zoneNearestEdge(lo, hi, price float64) float64 {
	if math.Abs(hi-price) < math.Abs(lo-price) {
		return hi
	}
	return lo
}

// ZoneSeatCandidates builds the candidate rows (rules 1–2) from the structure
// map. `src` is the pool the map's zones were drawn from (the read's uncapped
// in-band HTF zone universe): a zone's own detector row is looked up there by
// (kind, TF, lo, hi) so the candidate carries the SAME grading inputs the
// detector row has (ZonePattern, origin, formation). A zone with no source row
// is built from the map fields alone (HTF, TF, band) and counted in NoSource —
// its pattern is unknown, so isHTFSeatEligible will not protect it and the
// grade is whatever the zone grader gives those inputs; never invented.
// band ≤ 0 or a nil/empty map → no candidates. Pure.
func ZoneSeatCandidates(m *StructureMap, src []ScoredLevel, price, band float64) ([]DetectedLevel, ZoneSeatReport) {
	var rep ZoneSeatReport
	if m == nil || len(m.TFs) == 0 || price <= 0 || band <= 0 {
		return nil, rep
	}
	var out []DetectedLevel
	type zoneKey struct {
		kind, tf string
		lo, hi   float64
	}
	taken := map[zoneKey]bool{}
	for _, tf := range StructureMapTFs {
		st, ok := m.TFs[tf]
		if !ok {
			continue
		}
		for _, z := range st.Zones {
			if z.Hi <= z.Lo {
				continue // a line is not a zone (the map already filters these)
			}
			rep.Zones++
			edge := zoneNearestEdge(z.Lo, z.Hi, price)
			if math.Abs(edge-price) > band {
				rep.OutOfBand++
				continue
			}
			rep.InBand++
			if k := (zoneKey{z.Kind, tf, z.Lo, z.Hi}); taken[k] {
				rep.Duplicates++
				continue
			} else {
				taken[k] = true
			}
			label := ZoneSeatLabel(tf, LevelKind(z.Kind))
			var cand DetectedLevel
			found := false
			for _, s := range src {
				if string(s.Kind) == z.Kind && structureMapTFKey(s.TF) == tf && s.Lo == z.Lo && s.Hi == z.Hi {
					cand = s.DetectedLevel
					found = true
					break
				}
			}
			if !found {
				rep.NoSource++
				cand = DetectedLevel{Kind: LevelKind(z.Kind), Lo: z.Lo, Hi: z.Hi, TF: tf}
			}
			// Never carry the source row's record-only provenance or its merged
			// names into a new row: the candidate is its own reference.
			cand.Research = nil
			cand.CollapsedNames = nil
			cand.Price = edge
			cand.Label = label
			cand.HTF = true
			cand.Info = fmt.Sprintf("zone edge nearest price; zone %.2f–%.2f", z.Lo, z.Hi)
			out = append(out, cand)
			rep.Candidates = append(rep.Candidates, fmt.Sprintf("%s@%.2f", label, edge))
		}
	}
	return out, rep
}

// ZoneSeatOutcome is what the scorer's output RECORDS about the candidates
// (counters record, never infer): injected = ZONE-* rows standing in the graded
// pool; aliased = ZONE-* names that collapse folded into a detector row
// (CollapsedNames); seated = ZONE-* rows in the seated table. Rule 3 lives in
// collapseLevelClusters — a ZONE-* row within the cluster tolerance of ANY
// other survivor merges INTO it (never the other way round), so the alias is
// decided on survivors, after every cut, and can never be lost to a dedupe.
func ZoneSeatOutcome(pool, seated []ScoredLevel) (injected, aliased, seatedN int) {
	for _, p := range pool {
		if IsZoneSeatLabel(p.Label) {
			injected++
		}
		for _, n := range p.CollapsedNames {
			if IsZoneSeatLabel(n) {
				aliased++
			}
		}
	}
	for _, s := range seated {
		if IsZoneSeatLabel(s.Label) {
			seatedN++
		}
	}
	return injected, aliased, seatedN
}
