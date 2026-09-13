package api

import (
	"nofx/kernel"
	"time"
)

// Resolve the same wrap-aware chain date used by plan reads before dereferencing
// a session. Gaps are a refusal, not a nil session passed to downstream readers.
func (s *Server) planMutationSessionAt(traderID string, now time.Time) (*kernel.SessionDef, string, bool) {
	session, ok := s.planRegistry().ActiveSession(now)
	if !ok || session == nil {
		return nil, "", false
	}
	date, ok := kernel.PlanChainTradeDate(session, now)
	if !ok {
		return nil, "", false
	}
	if s.traderManager != nil {
		if at, err := s.traderManager.GetTrader(traderID); err == nil && at != nil {
			if runnable, _ := at.SessionRunnable(session); !runnable {
				return nil, "", false
			}
		}
	}
	return session, date, true
}
