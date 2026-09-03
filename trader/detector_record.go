package trader

import (
	"time"

	"nofx/kernel"
	"nofx/store"
)

// ── 1B — THE PRODUCTION CALL PATH ────────────────────────────────────────────
//
// Called once per planner read, from the same place the void scope is resolved,
// so the detector judges the SAME tape the prompt and the validator do. It
// changes no decision: it runs D1′ over the levels the read already produced
// and writes what it finds.
//
// A10: telemetry may WARN, never stop the loop. Every failure here returns
// after logging; nothing propagates to the caller.

// recordDetectorOutputs runs D1′ over the seated levels and persists any NEW
// episodes, then records the candidate pool for this read. Safe to call every
// read: the episode watermark comes from the STORE, so repeats write nothing.
func (at *AutoTrader) recordDetectorOutputs(
	symbol, planID, session string, planVersion int,
	allLevels []kernel.DetectedLevel, seated []kernel.ScoredLevel,
	price, dATR, proximityK float64, maxLevels int, now time.Time,
) {
	if at == nil || at.store == nil {
		return
	}
	defer func() {
		// A10 + class 23: a telemetry panic must never reach the trading loop.
		if r := recover(); r != nil {
			at.logWarnf("🔬 detector recording panicked and was contained: %v", r)
		}
	}()

	scope := kernel.ResolveVoidScope(symbol, now)
	if len(scope.Bars) == 0 {
		at.logInfof("🔬 detector: no bars in scope — nothing recorded this read (not zero episodes; UNMEASURED)")
		return
	}
	delta := kernel.MeanAbsIncrement(scope.Bars)
	if delta <= 0 {
		at.logWarnf("🔬 detector: Δ resolved to 0 — the band would be degenerate; skipping rather than recording a meaningless verdict")
		return
	}
	k, horizon, exitOn := kernel.DetectorK(), kernel.DetectorHorizonBars(), kernel.DetectorExitOn()
	dayMs := kernel.CMESessionDayStart(now).UnixMilli()
	ts := at.store.TouchOutcomes()

	written := 0
	for _, lv := range seated {
		if lv.Price <= 0 {
			continue
		}
		eps := kernel.DetectTouchOutcomes(scope.Bars, lv.Price, k, delta, horizon, exitOn)
		if len(eps) == 0 {
			continue
		}
		last := ts.LastOpenedAtMs(at.id, symbol, lv.Price, dayMs)
		for _, e := range kernel.NewEpisodesSince(eps, last) {
			row := &store.TouchOutcomeRow{
				TraderID: at.id, Symbol: symbol,
				LevelPrice: lv.Price, LevelKind: string(lv.Kind),
				CandidateSeated: true, PlanID: planID, PlanVersion: planVersion, Session: session,
				Ordinal: ts.NextOrdinal(at.id, symbol, lv.Price, dayMs),
				K:       k, Delta: delta, BandPts: k * delta, Horizon: horizon, ExitOn: exitOn,
				EntrySide: e.Entry, ExitSide: e.Exit, Outcome: e.Outcome,
				Ambiguous: e.IsAmbiguous(), BarsToExit: e.BarsToExit,
				MFEPts: e.MFE, MAEPts: e.MAE,
				OpenedAtMs: e.OpenedAtMs, ClosedAtMs: e.ClosedAtMs,
			}
			if err := ts.SaveOutcome(row); err != nil {
				at.logWarnf("🔬 detector: touch_outcomes write failed (level %.2f): %v", lv.Price, err)
				continue
			}
			written++
		}
	}

	// D3 — the candidate pool, including everything that did NOT seat.
	pool := kernel.BuildCandidatePool(allLevels, seated, price, dATR, proximityK, maxLevels)
	rows := make([]store.CandidatePoolRow, 0, len(pool))
	for _, c := range pool {
		rows = append(rows, store.CandidatePoolRow{
			TraderID: at.id, Symbol: symbol, PlanID: planID, PlanVersion: planVersion,
			Session: session, ReadAtMs: now.UnixMilli(),
			LevelPrice: c.Price, LevelKind: c.Kind, Label: c.Label,
			Rank: c.Rank, Seated: c.Seated, CutReason: c.CutReason,
			Score: c.Score, Threshold: c.Threshold, Grade: c.Grade,
			ScoreComponents: c.Components,
		})
	}
	cut := 0
	for _, r := range rows {
		if !r.Seated {
			cut++
		}
	}
	if err := at.store.CandidatePool().SavePool(rows); err != nil {
		at.logWarnf("🔬 detector: candidate_pool write failed (%d rows): %v", len(rows), err)
	}
	at.logInfof("🔬 detector recorded: %d new episode(s) · pool %d candidate(s) (%d seated, %d cut) · k=%.0f Δ=%.2f band=±%.2f H=%d %s",
		written, len(rows), len(rows)-cut, cut, k, delta, k*delta, horizon, exitOn)
}
