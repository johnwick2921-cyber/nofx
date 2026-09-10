// W1 item 2 — closing the episode. E6: every one of them, every time.
//
// A row left open after its session-day ends is this wave's own defect wearing
// a different hat. An experiment counting outcomes would skip it in silence and
// the denominator would be quietly wrong — which is exactly what a missing
// never-confirmed row does today. So the closer is exhaustive BY CONSTRUCTION
// (it closes every row with a NULL outcome for the key, not a list someone
// remembered to build) and the pin is "zero open rows".

package store

// Why an episode ended. Always recorded: a closed row with no cause has lost
// the only thing that distinguishes an orderly close from an abandoned one.
const (
	CloseCauseSessionEnd        = "session_close"
	CloseCauseVersionSuperseded = "version_superseded"
	CloseCauseInvalidated       = "invalidated"
	CloseCauseForcedExit        = "forced_exit"
)

// CloseOpenOpportunities closes every still-open episode for the key and
// returns how many it closed.
//
// factsFor is called PER ROW with that row's scenario link (which may be NULL —
// an unlinked touch still closes, and still gets an outcome). The caller owns
// the facts because only it can see the confirm, the arm and the fill; this
// function owns the guarantee that nothing is left open.
//
// Idempotent: it selects on a NULL outcome, so a second run closes nothing and
// cannot overwrite a recorded verdict.
func (s *TouchOutcomeStore) CloseOpenOpportunities(
	traderID, planID string, planVersion int, session, cause string,
	factsFor func(scenario *string) OpportunityFacts,
) (int, error) {
	if s == nil || s.db == nil || factsFor == nil {
		return 0, nil
	}
	var rows []TouchOutcomeRow
	if err := s.db.Where(
		"trader_id = ? AND plan_id = ? AND plan_version = ? AND session = ? AND opportunity_outcome IS NULL",
		traderID, planID, planVersion, session,
	).Find(&rows).Error; err != nil {
		return 0, err
	}

	closed := 0
	for i := range rows {
		outcome := OpportunityOutcomeFor(factsFor(rows[i].ScenarioNearest))
		c := cause
		if err := s.db.Model(&TouchOutcomeRow{}).
			Where("id = ?", rows[i].ID).
			Updates(map[string]any{
				"opportunity_outcome": outcome,
				"close_cause":         c,
			}).Error; err != nil {
			return closed, err
		}
		closed++
	}
	return closed, nil
}

// CountOpenOpportunities is the E6 probe: after a session-day ends this must be
// zero, and the boot line reports it so an unclosed row is visible rather than
// merely absent from a rate.
func (s *TouchOutcomeStore) CountOpenOpportunities(traderID string) (int64, error) {
	if s == nil || s.db == nil {
		return 0, nil
	}
	var n int64
	err := s.db.Model(&TouchOutcomeRow{}).
		Where("trader_id = ? AND opportunity_outcome IS NULL", traderID).
		Count(&n).Error
	return n, err
}

// CountClosedByOutcome feeds the boot line's per-rung counts.
func (s *TouchOutcomeStore) CountClosedByOutcome(traderID string, sinceMs int64) (map[string]int64, error) {
	out := map[string]int64{}
	if s == nil || s.db == nil {
		return out, nil
	}
	type row struct {
		Outcome string
		N       int64
	}
	var rs []row
	if err := s.db.Model(&TouchOutcomeRow{}).
		Select("opportunity_outcome as outcome, COUNT(*) as n").
		Where("trader_id = ? AND opportunity_outcome IS NOT NULL AND opened_at_ms >= ?", traderID, sinceMs).
		Group("opportunity_outcome").Scan(&rs).Error; err != nil {
		return out, err
	}
	for _, r := range rs {
		out[r.Outcome] = r.N
	}
	return out, nil
}
