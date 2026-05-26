package kernel

// Plan 3 Task 21 — Risk limits + force-flat kill switch.
//
// Hard server-side guard rails enforced BEFORE expensive prompt assembly.
// The AI prompt itself only contains soft hints ("don't risk more than 1%"),
// which the model may ignore. These limits are non-negotiable.
//
// Design goals:
//   - Plain primitives in CheckPreTrade so the function is trivially testable
//     without constructing a full *Context / *Decision (the plan's illustrative
//     types do not exactly match the real ones).
//   - Decoupled from concrete broker packages (ninjatrader, binance, etc.) so
//     that the engine can call ForceFlat via a tiny interface.
//   - Loaded from env once at startup via LoadRiskLimitsFromConfig().

import (
	"fmt"
	"nofx/config"
	"nofx/logger"
	"sync"
	"time"
)

// RiskLimits caps the risk a strategy can take before the engine intervenes.
// Zero values disable the corresponding check (so partial config is safe).
type RiskLimits struct {
	MaxDailyLossUSD      float64 // hard stop for the trading day (USD, positive number)
	MaxConcurrentTrades  int     // open position cap (count)
	MaxNotionalUSD       float64 // sum of (entry price * quantity) cap across open positions
	MaxContractsPerOrder int     // single-order contract size cap
}

// RiskLimitDecision is the qualitative outcome of a pre-trade check.
// Callers may use the typed err for logging and the decision for action.
type RiskLimitDecision int

const (
	RiskAllow      RiskLimitDecision = iota // no constraints hit
	RiskBlockEntry                          // block this entry, hold existing positions
	RiskForceFlat                           // close everything (daily-loss tripped)
)

// CheckPreTrade evaluates the proposed trade against all enabled limits.
// Inputs are plain primitives so this can be exercised in a unit test
// without a full engine.Context.
//
//   totalPnL          — account.TotalPnL (negative = losing)
//   openPositions     — len(ctx.Positions)
//   requestedNotional — abs(decision.PositionSizeUSD) for the proposed entry
//   existingNotional  — sum of MarkPrice*Quantity over existing positions
//
// Semantics:
//   - PnL strictly LESS THAN -MaxDailyLossUSD trips the daily-loss limit.
//     PnL == -MaxDailyLossUSD does NOT trip (boundary is permissive, matches
//     the plan's `< -r.MaxDailyLossUSD` source). PnL == limit slightly past
//     trips immediately.
//   - openPositions >= cap trips (cap-equals-trip).
//   - existing + requested > cap trips (strict exceed).
//   - When MaxDailyLossUSD == 0 the daily-loss check is disabled; same for
//     MaxConcurrentTrades and MaxNotionalUSD (so partial config is safe).
func (r *RiskLimits) CheckPreTrade(totalPnL float64, openPositions int, requestedNotional float64, existingNotional float64) error {
	if r == nil {
		return nil
	}
	if r.MaxDailyLossUSD > 0 && totalPnL < -r.MaxDailyLossUSD {
		return fmt.Errorf("risk: daily loss limit hit (pnl=%.2f, limit=-%.2f)", totalPnL, r.MaxDailyLossUSD)
	}
	if r.MaxConcurrentTrades > 0 && openPositions >= r.MaxConcurrentTrades {
		return fmt.Errorf("risk: concurrent trade cap reached (open=%d, cap=%d)", openPositions, r.MaxConcurrentTrades)
	}
	if r.MaxNotionalUSD > 0 {
		total := existingNotional + requestedNotional
		if total > r.MaxNotionalUSD {
			return fmt.Errorf("risk: notional cap exceeded (existing=%.2f + requested=%.2f > cap=%.2f)", existingNotional, requestedNotional, r.MaxNotionalUSD)
		}
	}
	return nil
}

// Classify returns the qualitative decision class without re-running checks.
// daily-loss trip → ForceFlat (close everything).
// other trips     → BlockEntry (hold existing, refuse new).
// nil err         → Allow.
func (r *RiskLimits) Classify(totalPnL float64, openPositions int, requestedNotional float64, existingNotional float64) (RiskLimitDecision, error) {
	if r == nil {
		return RiskAllow, nil
	}
	if r.MaxDailyLossUSD > 0 && totalPnL < -r.MaxDailyLossUSD {
		return RiskForceFlat, fmt.Errorf("risk: daily loss limit hit (pnl=%.2f, limit=-%.2f)", totalPnL, r.MaxDailyLossUSD)
	}
	if err := r.CheckPreTrade(totalPnL, openPositions, requestedNotional, existingNotional); err != nil {
		return RiskBlockEntry, err
	}
	return RiskAllow, nil
}

// LoadRiskLimitsFromConfig pulls the four limits from config.Get().
// Called once at startup or each cycle (cheap — just struct copy).
func LoadRiskLimitsFromConfig() RiskLimits {
	c := config.Get()
	return RiskLimits{
		MaxDailyLossUSD:      c.RiskMaxDailyLossUSD,
		MaxConcurrentTrades:  c.RiskMaxConcurrentTrades,
		MaxNotionalUSD:       c.RiskMaxNotionalUSD,
		MaxContractsPerOrder: c.RiskMaxContractsPerOrder,
	}
}

// ============================================================================
// Daily PnL reset
// ============================================================================
//
// The MaxDailyLossUSD limit is evaluated against ctx.Account.TotalPnL, which
// is a session-cumulative figure tracked by the trader. A true "daily" reset
// requires either:
//   (a) the trader to reset its TotalPnL at the CME session boundary
//       (Sunday 18:00 ET); or
//   (b) the engine to remember the start-of-day equity and compute
//       (current_equity - day_open_equity) on each cycle.
//
// For Plan 3 v1 we ship a SIMPLE day-boundary tracker here so the API
// endpoint and tests have something to call. It DOES NOT yet hook into
// CME session boundaries (deferred to a follow-up alongside T22 drift).
//
// Manual reset path: ResetDailyPnL() — called by the operator via the
// (future) force-flat API endpoint after they have wired in a fresh
// session.
//
// Automatic check-and-reset path: MaybeResetDaily(now) — called at the
// top of each decision cycle. If the wall-clock date has changed since
// last reset, the tracker zeroes itself. This is a coarse approximation
// of session boundaries (UTC date rollover, not 18:00 ET) and is
// adequate for SIM-mode paper trading; live trading should switch to
// CME-aware reset by checking IsCMEOpen + transition detection.

var (
	dailyResetMu       sync.Mutex
	lastDailyResetDate string // YYYY-MM-DD in UTC
)

// ResetDailyPnL marks "now" as the new day-start. Cheap, idempotent.
// Operators call this from the force-flat API endpoint or at the
// CME Sunday 18:00 ET open.
func ResetDailyPnL() {
	dailyResetMu.Lock()
	defer dailyResetMu.Unlock()
	lastDailyResetDate = time.Now().UTC().Format("2006-01-02")
	logger.Infof("Plan 3 T21: daily PnL window reset at %s UTC", lastDailyResetDate)
}

// MaybeResetDaily checks for a UTC-date rollover since the last reset and
// resets the daily window when it has. Returns true if a reset fired.
func MaybeResetDaily(now time.Time) bool {
	today := now.UTC().Format("2006-01-02")
	dailyResetMu.Lock()
	last := lastDailyResetDate
	dailyResetMu.Unlock()
	if last == "" || last != today {
		ResetDailyPnL()
		return true
	}
	return false
}
