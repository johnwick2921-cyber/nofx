package kernel

import (
	"testing"
	"time"
)

// --- Strategy Studio Phase 1 — Chunk 2 daily guardrails ---

func TestDailyGuardrails_DailyLoss(t *testing.T) {
	base := DailyGuardrails{MasterEnabled: true, DailyLossEnabled: true, DailyLossLimitUSD: 500}

	// loss exactly AT the limit trips (stop once loss >= $X)
	g := base
	g.DailyRealizedPnL = -500
	if d, err := g.Check(); err == nil || d != RiskForceFlat {
		t.Fatalf("loss=-500 limit=500 should ForceFlat, got d=%v err=%v", d, err)
	}
	// loss past the limit trips
	g = base
	g.DailyRealizedPnL = -600.01
	if _, err := g.Check(); err == nil {
		t.Fatal("loss=-600 should trip the daily-loss limit")
	}
	// loss under the limit passes
	g = base
	g.DailyRealizedPnL = -499
	if _, err := g.Check(); err != nil {
		t.Fatalf("loss=-499 under limit should pass, got %v", err)
	}
	// toggle OFF → not enforced even at a huge loss
	g = base
	g.DailyLossEnabled = false
	g.DailyRealizedPnL = -10000
	if _, err := g.Check(); err != nil {
		t.Fatalf("daily-loss toggle OFF must not block, got %v", err)
	}
}

func TestDailyGuardrails_ProfitTarget(t *testing.T) {
	g := DailyGuardrails{MasterEnabled: true, DailyProfitEnabled: true, DailyProfitTargetUSD: 1000, DailyRealizedPnL: 1000}
	if d, err := g.Check(); err == nil || d != RiskBlockEntry {
		t.Fatalf("profit=1000 target=1000 should BlockEntry, got d=%v err=%v", d, err)
	}
	g.DailyRealizedPnL = 999
	if _, err := g.Check(); err != nil {
		t.Fatalf("profit=999 under target should pass, got %v", err)
	}
	g.DailyProfitEnabled = false
	g.DailyRealizedPnL = 999999
	if _, err := g.Check(); err != nil {
		t.Fatalf("profit toggle OFF must not block, got %v", err)
	}
}

func TestDailyGuardrails_MaxTrades(t *testing.T) {
	g := DailyGuardrails{MasterEnabled: true, MaxDailyTradesEnabled: true, MaxDailyTrades: 3, TradesToday: 3}
	if d, err := g.Check(); err == nil || d != RiskBlockEntry {
		t.Fatalf("trades=3 max=3 should BlockEntry (4th blocked), got d=%v err=%v", d, err)
	}
	g.TradesToday = 2
	if _, err := g.Check(); err != nil {
		t.Fatalf("trades=2 max=3 should pass, got %v", err)
	}
	g.MaxDailyTradesEnabled = false
	g.TradesToday = 99
	if _, err := g.Check(); err != nil {
		t.Fatalf("max-trades toggle OFF must not block, got %v", err)
	}
}

func TestDailyGuardrails_MasterSwitchBypassesAll(t *testing.T) {
	// Master OFF must bypass EVERY guardrail, even ones that would otherwise trip.
	g := DailyGuardrails{
		MasterEnabled:         false,
		DailyLossEnabled:      true, DailyLossLimitUSD: 100, DailyRealizedPnL: -9999,
		MaxDailyTradesEnabled: true, MaxDailyTrades: 1, TradesToday: 99,
	}
	if d, err := g.Check(); err != nil || d != RiskAllow {
		t.Fatalf("master OFF must bypass ALL guardrails (RiskAllow,nil), got d=%v err=%v", d, err)
	}
}

func TestDailyGuardrails_UnsetValueNotEnforced(t *testing.T) {
	// toggle ON but value 0 (unset) → not enforced (no false block)
	g := DailyGuardrails{MasterEnabled: true, DailyLossEnabled: true, DailyLossLimitUSD: 0, DailyRealizedPnL: -100000}
	if _, err := g.Check(); err != nil {
		t.Fatalf("enabled but value=0 must not block, got %v", err)
	}
}

func TestBoolOrDefault(t *testing.T) {
	tr, fa := true, false
	if boolOrDefault(nil, true) != true || boolOrDefault(nil, false) != false {
		t.Fatal("nil should return the default")
	}
	if boolOrDefault(&tr, false) != true || boolOrDefault(&fa, true) != false {
		t.Fatal("non-nil should return the pointed value")
	}
}

func TestFirstPositive(t *testing.T) {
	if firstPositive(0, 500) != 500 {
		t.Fatal("per-strategy 0 → env fallback 500")
	}
	if firstPositive(250, 500) != 250 {
		t.Fatal("per-strategy 250 → overrides env")
	}
	if firstPositive(0, 0) != 0 {
		t.Fatal("both unset → 0")
	}
}

func TestCMESessionDayStart(t *testing.T) {
	chicago, _ := time.LoadLocation("America/Chicago")
	// 18:00 CT → session started 17:00 CT the SAME calendar day
	now := time.Date(2026, 6, 1, 18, 0, 0, 0, chicago)
	s := CMESessionDayStart(now)
	if s.Hour() != 17 || s.Month() != time.June || s.Day() != 1 {
		t.Fatalf("18:00 CT → session start 17:00 same day, got %v", s)
	}
	// 09:00 CT → session started 17:00 CT the PREVIOUS calendar day
	now = time.Date(2026, 6, 1, 9, 0, 0, 0, chicago)
	s = CMESessionDayStart(now)
	if s.Hour() != 17 || s.Month() != time.May || s.Day() != 31 {
		t.Fatalf("09:00 CT → session start 17:00 prev day (May 31), got %v", s)
	}
}

// --- Chunk 3 — max contracts + visible equity×20 cap resolvers ---

func TestResolveMaxContracts(t *testing.T) {
	on, off := true, false
	if got := ResolveMaxContracts(nil, nil, 4, 10); got != 4 { // per-strategy overrides
		t.Fatalf("per-strategy 4 → 4, got %d", got)
	}
	if got := ResolveMaxContracts(nil, nil, 0, 10); got != 10 { // unset → default
		t.Fatalf("unset → default 10, got %d", got)
	}
	if got := ResolveMaxContracts(nil, &off, 4, 10); got != 0 { // toggle off → no clamp
		t.Fatalf("toggle off → 0 (no clamp), got %d", got)
	}
	if got := ResolveMaxContracts(&off, &on, 4, 10); got != 0 { // master off → no clamp
		t.Fatalf("master off → 0 (no clamp), got %d", got)
	}
}

func TestResolveNotionalLeverage(t *testing.T) {
	off := false
	if got := ResolveNotionalLeverage(nil, nil, 50, 20); got != 50 { // per-strategy overrides
		t.Fatalf("per-strategy 50 → 50, got %v", got)
	}
	if got := ResolveNotionalLeverage(nil, nil, 0, 20); got != 20 { // unset → default
		t.Fatalf("unset → default 20, got %v", got)
	}
	if got := ResolveNotionalLeverage(nil, &off, 50, 20); got != 0 { // toggle off → no cap
		t.Fatalf("toggle off → 0 (no cap), got %v", got)
	}
	if got := ResolveNotionalLeverage(&off, nil, 50, 20); got != 0 { // master off → no cap
		t.Fatalf("master off → 0 (no cap), got %v", got)
	}
}
