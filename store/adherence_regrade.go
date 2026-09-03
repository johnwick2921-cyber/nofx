package store

import (
	"fmt"
	"strings"

	"nofx/logger"
)

// ── ADHERENCE REGRADE (owner ruling 2026-09-03) ──────────────────────────────
//
// RepairArmedLineage stamped plan lineage onto positions that closed without
// it, then tried to clear the adherence grade so the analytics would regrade
// them — but it only cleared grade "F", while a close with no citation grades
// "D" (kernel/adherence.go). The predicate waited for a letter that path never
// writes, so late-stamped positions kept a permanent off-plan D while carrying
// full lineage.
//
// The predicate is fixed forward in reconcile.go. This is the backfill for the
// rows already stuck, and the owner ruled it explicit rather than silent: a
// logged, counted correction is not a silent backfill; leaving known-wrong
// grades in a published table is.
//
// SCOPE — exactly four rows, and the narrowing matters. The first count offered
// was SEVEN, from the loose discriminator `cited_scenario_id != ”`:
//
//	530  system           matched=0  cited "off-plan"  → the citation is the
//	                                                     literal string; it is
//	                                                     genuinely off-plan and
//	                                                     KEEPS its D
//	572  e7_farside_test  matched=1  cited "TEST-E7"   → the test seam
//	582  armed_entry      matched=0  cited S2          → direction did not
//	                                                     match; KEEPS its D
//	575 584 586 591       matched=1  S2/S2/S3/S1       → all source=reconcile,
//	                                                     the real ones
//
// Backfilling the seven would have regraded a test-seam row and promoted two
// rows that legitimately earned D.
var AdherenceRegradeIDs = []int64{575, 584, 586, 591}

const adherenceRegradeFlag = "adherence_regrade_2026_09_03_done"

// RegradeStuckAdherence clears the adherence grade on exactly the rows above so
// the analytics regrade them with the lineage now in hand. Idempotent behind a
// system_config flag; WHERE-scoped to the id list AND re-checked per row, so a
// row that no longer matches its recorded shape is skipped and named.
func (s *Store) RegradeStuckAdherence() {
	if v, err := s.GetSystemConfig(adherenceRegradeFlag); err == nil && v == "1" {
		return
	}
	before, err := s.adherenceDistribution()
	if err != nil {
		logger.Warnf("🩹 adherence regrade: could not read the before distribution: %v", err)
		return
	}
	regraded, skipped := 0, []string{}
	for _, id := range AdherenceRegradeIDs {
		var row struct {
			ID       int64
			Grade    string
			Matched  int
			Scenario string
			Source   string
			Version  int
		}
		if err := s.gdb.Raw(`SELECT id, COALESCE(adherence_grade,'') AS grade, COALESCE(plan_matched,0) AS matched,
COALESCE(cited_scenario_id,'') AS scenario, COALESCE(source,'') AS source, COALESCE(plan_version,0) AS version
FROM trader_positions WHERE id = ?`, id).Scan(&row).Error; err != nil || row.ID == 0 {
			skipped = append(skipped, fmt.Sprintf("%d(not found)", id))
			continue
		}
		// Re-verify the shape rather than trusting the id list: stamped lineage,
		// a real citation, direction agreed, and not the test seam.
		switch {
		case row.Grade != "D":
			skipped = append(skipped, fmt.Sprintf("%d(grade=%s, already regraded)", id, row.Grade))
		case row.Matched != 1 || row.Scenario == "" || row.Version == 0:
			skipped = append(skipped, fmt.Sprintf("%d(no longer matches: matched=%d scen=%q v=%d)", id, row.Matched, row.Scenario, row.Version))
		case row.Source == "e7_farside_test":
			skipped = append(skipped, fmt.Sprintf("%d(test seam)", id))
		default:
			if err := s.Position().SetAdherence(id, ""); err != nil {
				skipped = append(skipped, fmt.Sprintf("%d(write failed: %v)", id, err))
				continue
			}
			regraded++
		}
	}
	after, _ := s.adherenceDistribution()
	logger.Infof("🩹 adherence regrade: %d row(s) cleared for regrading (ids %v) · before %s → after %s%s",
		regraded, AdherenceRegradeIDs, before, after, skippedSuffix(skipped))
	_ = s.SetSystemConfig(adherenceRegradeFlag, "1")
}

func skippedSuffix(skipped []string) string {
	if len(skipped) == 0 {
		return ""
	}
	return " · skipped: " + strings.Join(skipped, ", ")
}

// adherenceDistribution renders the closed-position grade histogram, so the
// before/after is in the journal and not only in a report.
func (s *Store) adherenceDistribution() (string, error) {
	type row struct {
		Grade string
		N     int
	}
	var rows []row
	if err := s.gdb.Raw(`SELECT COALESCE(NULLIF(adherence_grade,''),'(none)') AS grade, COUNT(*) AS n
FROM trader_positions WHERE status != 'OPEN' GROUP BY grade ORDER BY grade`).Scan(&rows).Error; err != nil {
		return "", err
	}
	parts := make([]string, 0, len(rows))
	for _, r := range rows {
		parts = append(parts, fmt.Sprintf("%s=%d", r.Grade, r.N))
	}
	return strings.Join(parts, " "), nil
}

// AdherenceRegradeBootLine reports the count the migration is scoped to, READ
// from the id list rather than typed.
func AdherenceRegradeBootLine() string {
	return fmt.Sprintf("adherence regraded=%d (ids %v — late-stamped reconcile closes; a genuinely uncited close keeps its D)",
		len(AdherenceRegradeIDs), AdherenceRegradeIDs)
}
