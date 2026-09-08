package trader

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"nofx/kernel"
	"nofx/researchsnapshot"
	"time"
)

// Caller transfers the completed local detector slices; no worker reads a live
// BarCache, trader configuration pointer, or mutable plan document.
func recordResearchCandidates(id, symbol string, raw []kernel.DetectedLevel, seated []kernel.ScoredLevel, observed time.Time) {
	recordResearchCandidatesAt(id, symbol, raw, seated, observed, time.Now())
}
func recordResearchCandidatesAt(id, symbol string, raw []kernel.DetectedLevel, seated []kernel.ScoredLevel, observed, received time.Time) {
	researchsnapshot.Record("candidate:planner_read", func() []researchsnapshot.Fact {
		ranks := map[string]int{}
		key := func(l kernel.DetectedLevel) string { return fmt.Sprintf("%s|%g|%s", l.Kind, l.Price, l.Label) }
		for i, l := range seated {
			ranks[key(l.DetectedLevel)] = i + 1
		}
		out := make([]researchsnapshot.Fact, 0, len(raw))
		for _, l := range raw {
			f := researchsnapshot.NewFact("candidate", "planner_read", researchsnapshot.Value(id), researchsnapshot.Clocks{ObservationMS: researchsnapshot.Value(observed.UnixMilli()), ReceiptMS: researchsnapshot.Value(received.UnixMilli())})
			identity := fmt.Sprintf("%s|%s|%g|%g|%s|%s|%d", symbol, l.Kind, l.Lo, l.Hi, l.OriginDate, l.TF, l.FormedAtMs)
			h := sha256.Sum256([]byte(identity))
			f.Set("stable_id", hex.EncodeToString(h[:]))
			f.Set("identity_basis", "symbol/kind/bounds/origin date/timeframe/formation; unknown formation may alias episodes")
			f.Set("root_symbol", symbol)
			f.Set("raw_origin", l)
			f.Set("price", l.Price)
			f.Set("lo", l.Lo)
			f.Set("hi", l.Hi)
			if l.FormedAtMs > 0 {
				f.Set("formation_ms", l.FormedAtMs)
			}
			f.Set("availability_ms", received.UnixMilli())
			if l.TF != "" {
				f.Set("timeframe", l.TF)
			}
			if l.ZonePattern != "" {
				f.Set("zone_pattern", l.ZonePattern)
			}
			if c := l.Research; c != nil {
				f.Set("family", c.Family)
				f.Set("freshness_at_read", c.Freshness)
				f.Set("confluence_raw", c.ConfluenceRaw)
				f.Set("confluence_capped", c.ConfluenceCapped)
				if c.Score != nil {
					f.Set("raw_score_components", c)
					f.Set("capped_score_components", map[string]any{"confluence": c.ConfluenceCapped, "grade": c.Grade})
					f.Set("final_score", c.Score)
					f.Set("grade", c.Grade)
				}
				f.Set("overrides", c.Overrides)
				f.Set("exclusion_reason", c.Exclusion)
			}
			if rank, ok := ranks[key(l)]; ok {
				f.Set("rank", rank)
				f.Set("selection_outcome", "seated")
				f.Unknown("exclusion_reason", "selected; no exclusion")
			} else {
				f.Set("selection_outcome", "cut")
				if f.Fields["exclusion_reason"] == nil {
					f.Unknown("exclusion_reason", "later selection override did not emit a cause")
				}
			}
			f.Unknown("role", "exclusion does not invalidate the level; target/obstacle/invalidation role is not separately authored at this stage")
			out = append(out, f)
		}
		return out
	})
}

func recordResearchInput(id string, input kernel.PlannerInput, systemPrompt, model string) {
	recordResearchInputAt(id, input, systemPrompt, model, time.Now())
}
func recordResearchInputAt(id string, input kernel.PlannerInput, systemPrompt, model string, received time.Time) {
	researchsnapshot.Record("plan:input", func() []researchsnapshot.Fact {
		f := researchsnapshot.NewFact("plan", "input", researchsnapshot.Value(id), researchsnapshot.Clocks{ObservationMS: researchsnapshot.Value(input.Now.UnixMilli()), ReceiptMS: researchsnapshot.Value(received.UnixMilli())})
		f.Set("input_snapshot_id", id)
		f.Set("input_snapshot", input)
		f.Set("model", model)
		f.Set("system_prompt", systemPrompt)
		f.Set("config_version", input.AIConfigHash)
		return []researchsnapshot.Fact{f}
	})
}

// The caller has already evaluated these verdicts and serialized its metadata.
// No confirmation evaluator is called here.
func recordResearchPermissions(planID string, version int, meta string, observed time.Time) {
	recordResearchPermissionsAt(planID, version, meta, observed, time.Now())
}
func recordResearchPermissionsAt(planID string, version int, meta string, observed, received time.Time) {
	researchsnapshot.Record("scenario:revalidation", func() []researchsnapshot.Fact {
		f := researchsnapshot.NewFact("scenario", "revalidation", nil, researchsnapshot.Clocks{ObservationMS: researchsnapshot.Value(observed.UnixMilli()), ReceiptMS: researchsnapshot.Value(received.UnixMilli()), PermissionMS: researchsnapshot.Value(observed.UnixMilli())})
		f.Set("plan_id", planID)
		f.Set("plan_version", version)
		f.Set("revalidation", json.RawMessage(meta))
		f.Unknown("permission_status", "scenario activation and confirmation verdicts; not an order authorization")
		return []researchsnapshot.Fact{f}
	})
}
