package trader

import (
	"encoding/json"
	"testing"
	"time"

	"nofx/kernel"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// Wave 2 armed orders — Phase 1 manager tests: gate-at-arm, ledger upsert,
// cancel on dormant (1.4), cancel on no active plan (2.4).

func armedDoc() string {
	d := kernel.PlanDoc{Bias: kernel.PlanBias{Direction: "long", Conviction: "low", FlipCondition: "n/a"},
		Levels: []kernel.PlanLevel{{Price: 100, Label: "PDH", Grade: "A", Instruction: "fade"}},
		Scenarios: []kernel.PlanScenario{{ID: "S1", Trigger: "t", Condition: "fvg_entry", Direction: "long",
			TargetChain: []float64{110}, Invalid: "i", Quality: "B",
			Fvg: &kernel.PlanFvgEntry{Lo: 98, Hi: 102, CE: 100, EntryMode: "ce", DisplacementATR: 1.5, OriginLevel: "ONH", Direction: "long"},
			Arm: &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 110}},
		},
		NoTrade: []string{}, DeathCondition: "n/a",
	}
	blob, _ := json.Marshal(d)
	return string(blob)
}

func TestArmedOrderUpsertAndGateRR(t *testing.T) {
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}
	cfg.RiskControl = store.RiskControlConfig{MinRiskRewardRatio: 1.5}
	at, st := resetTrader(t, cfg)
	now := time.Now()
	sess, ok := at.sessionRegistry(now).ActiveSession(now)
	if !ok {
		t.Skip("no active session right now")
	}
	cfg.DayPlan.SessionsEnabled = []string{sess.Name}
	trueV := true
	cfg.DayPlan.Sessions = []store.DayPlanSessionOverride{{Session: sess.Name, Enable: &trueV}}
	td, _ := kernel.PlanChainTradeDate(sess, now)
	pid := store.MakePlanIDForTrader(at.id, td, sess.Name)
	if _, err := st.Plan().AppendPlan(&store.PlanDB{PlanID: pid, TradeDate: td, Session: sess.Name, StrategyID: at.id, Lifecycle: "active", Doc: armedDoc(), CreatedAt: now.Add(-30 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	installActivePlanProvider(at, st)

	at.maybeManageArmedOrders(nil)

	rows, err := st.ArmedOrders().ListNonTerminal()
	if err != nil || len(rows) != 1 {
		t.Fatalf("arm rows = %d err=%v want 1 (R:R 2.0 ≥ min 1.5 must pass at arm time)", len(rows), err)
	}
	if rows[0].EntryClass != "armed_fill" || rows[0].Scenario != "S1" {
		t.Fatalf("armed row lineage wrong: %+v", rows[0])
	}
}

func TestArmedGateRefusesBadRR(t *testing.T) {
	at, _ := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}})
	at.config.StrategyConfig.RiskControl = store.RiskControlConfig{MinRiskRewardRatio: 4}
	sc := kernel.PlanScenario{ID: "S1", Condition: "reject", Direction: "long", Quality: "A",
		Arm: &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 110}} // R:R 2.0 < 4.0
	if v := at.armGateVerdict(sc, "long", nil, 0, "", at.config.StrategyConfig); v == "" {
		t.Fatal("R:R 2.0 below min 4.0 must be refused at arm time")
	}
	// quality floor: a C scenario below min_scenario_quality=B must refuse.
	scC := kernel.PlanScenario{ID: "S1", Condition: "reject", Direction: "long", Quality: "C",
		Arm: &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 110}}
	if v := at.armGateVerdict(scC, "long", nil, 0, "B", at.config.StrategyConfig); v == "" {
		t.Fatal("C-quality below min_scenario_quality=B must be refused")
	}
	// plan_mode direction against bias
	at.config.StrategyConfig.DayPlan.PlanMode = "direction"
	sc.Direction = "short"
	if v := at.armGateVerdict(sc, "long", nil, 0, "", at.config.StrategyConfig); v == "" {
		t.Fatal("short arm against long bias under plan_mode=direction must be refused")
	}
}

func TestArmedCancelOnDormant(t *testing.T) {
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}
	at, st := resetTrader(t, cfg)
	now := time.Now()
	sess, ok := at.sessionRegistry(now).ActiveSession(now)
	if !ok {
		t.Skip("no active session right now")
	}
	cfg.DayPlan.SessionsEnabled = []string{sess.Name}
	trueV := true
	cfg.DayPlan.Sessions = []store.DayPlanSessionOverride{{Session: sess.Name, Enable: &trueV}}
	td, _ := kernel.PlanChainTradeDate(sess, now)
	pid := store.MakePlanIDForTrader(at.id, td, sess.Name)
	v, err := st.Plan().AppendPlan(&store.PlanDB{PlanID: pid, TradeDate: td, Session: sess.Name, StrategyID: at.id, Lifecycle: "active", Doc: armedDoc(), CreatedAt: now.Add(-30 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.ArmedOrders().UpsertArm(&store.ArmedOrderDB{TraderID: at.id, PlanID: pid, Version: v, Session: sess.Name, Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: "armed"}); err != nil {
		t.Fatal(err)
	}
	// go dormant → the manager must cancel instantly (1.4).
	if err := st.Plan().UpdatePlanLifecycle(pid, v, "dormant", "dormant:flip-condition: test"); err != nil {
		t.Fatal(err)
	}
	installActivePlanProvider(at, st)
	at.maybeManageArmedOrders(nil)
	rows, err := st.ArmedOrders().ListNonTerminal()
	if err != nil || len(rows) != 0 {
		t.Fatalf("non-terminal rows = %d err=%v — dormant must cancel ALL arms", len(rows), err)
	}
}

func TestArmedCancelOnNoActivePlan(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}})
	now := time.Now()
	if err := st.ArmedOrders().UpsertArm(&store.ArmedOrderDB{TraderID: at.id, PlanID: "x:NY", Version: 1, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: "armed"}); err != nil {
		t.Fatal(err)
	}
	_ = now
	at.maybeManageArmedOrders(nil)
	if rows, err := st.ArmedOrders().ListNonTerminal(); err != nil || len(rows) != 0 {
		t.Fatalf("no active plan must cancel arms (rows=%d err=%v)", len(rows), err)
	}
}

// PHASE 4 — the order_update event machine: filled → filled+lineage path,
// rejected/cancelled → disarmed with a reason (never silent).
func TestArmedOrderUpdateTransitions(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}})
	ledger := st.ArmedOrders()
	if err := ledger.UpsertArm(&store.ArmedOrderDB{TraderID: at.id, PlanID: "2026-08-27:NY:trader-1", Version: 1, Session: "NY", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: "working", SignalID: "sig-1"}); err != nil {
		t.Fatal(err)
	}
	at.onArmedOrderUpdate(ntwire.OrderUpdatePayload{SignalID: "sig-1", State: "filled", FillPrice: 100.5}, ledger)
	rows, _ := ledger.ListForPlan("2026-08-27:NY:trader-1")
	if len(rows) != 1 || rows[0].State != "filled" {
		t.Fatalf("filled transition: %+v", rows)
	}
	_ = ledger.SetState(rows[0].ID, "working", "")
	at.onArmedOrderUpdate(ntwire.OrderUpdatePayload{SignalID: "sig-1", State: "rejected"}, ledger)
	rows, _ = ledger.ListForPlan("2026-08-27:NY:trader-1")
	if rows[0].State != "cancelled" || rows[0].StateReason == "" {
		t.Fatalf("reject must disarm with a reason: %+v", rows[0])
	}
}

// PHASE 4 — short twin of the R:R arm gate (long side covered elsewhere).
func TestArmedGateRRShortTwin(t *testing.T) {
	at, _ := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}})
	at.config.StrategyConfig.RiskControl = store.RiskControlConfig{MinRiskRewardRatio: 3}
	sc := kernel.PlanScenario{ID: "S1", Condition: "reject", Direction: "short", Quality: "A",
		Arm: &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 105, Target: 90}} // R:R (100−90)/(105−100)=2.0 < 3.0
	if v := at.armGateVerdict(sc, "short", nil, 0, "", at.config.StrategyConfig); v == "" {
		t.Fatal("short arm R:R 2.0 below min 3.0 must be refused")
	}
	sc.Arm.Target = 85 // R:R 3.0 → pass
	if v := at.armGateVerdict(sc, "short", nil, 0, "", at.config.StrategyConfig); v != "" {
		t.Fatalf("short arm R:R 3.0 must pass, got %q", v)
	}
}
