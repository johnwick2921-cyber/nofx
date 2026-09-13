package trader

import (
	"encoding/json"
	"nofx/kernel"
	"nofx/market"
	"nofx/store"
	"testing"
	"time"
)

func TestStructuralStopF4RefusalRetiresAuthorizationAndSendsNoOrder(t *testing.T) {
	now := time.Date(2026, 9, 11, 15, 0, 0, 0, time.UTC)
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}
	oneSetupOff(&cfg)
	structuralTestPolicy(&cfg, 5)
	cfg.RiskControl.MinRiskRewardRatio = 2
	at, st, signals, cancels := shadowWireHarnessAt(t, cfg, now)
	doc := kernel.PlanDoc{Bias: kernel.PlanBias{Direction: "long"}, Levels: []kernel.PlanLevel{{Price: 100, Label: "PDL", Grade: "A", Instruction: "fade"}}, Scenarios: []kernel.PlanScenario{{ID: "S1", Condition: "reject", Direction: "long", Quality: "A", Confirm: &kernel.PlanConfirm{Rule: "touch", RefPrice: 100, Side: "above"}, Arm: &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 77, Target: 104}}}}
	structuralTestMap(&doc, structuralTestZone{100, 82, 100, "PDL"}, structuralTestZone{104, 104, 105, "target"})
	raw, _ := json.Marshal(doc)
	pid := shadowPlanAtTime(t, at, st, string(raw), now)
	row := &store.ArmedOrderDB{TraderID: at.id, PlanID: pid, Version: 1, Scenario: "S1", Side: "LONG", EntryPx: 100, StopPx: 77, TargetPx: 104, State: store.StateArmed, Kind: "limit", Condition: "reject", BootID: store.ProcessBootID(), CreatedAt: now, UpdatedAt: now}
	if err := st.ArmedOrders().UpsertArm(row); err != nil {
		t.Fatal(err)
	}
	previous := market.FuturesBarsProvider
	market.FuturesBarsProvider = func(string, string, int) []market.Kline { return shadowBarsNearAt(100, now) }
	t.Cleanup(func() { market.FuturesBarsProvider = previous })
	at.maybeManageArmedOrdersAt(nil, now)
	select {
	case signal := <-signals:
		t.Fatalf("geometry refusal reached broker: %+v", signal)
	case <-time.After(300 * time.Millisecond):
	}
	select {
	case cancel := <-cancels:
		t.Fatalf("unplaced authorization must retire locally, not send cancel: %+v", cancel)
	default:
	}
	rows, err := st.ArmedOrders().ListForPlan(pid)
	if err != nil || len(rows) != 1 || rows[0].State != store.StateCancelled {
		t.Fatalf("bad old authorization remains placeable: %+v %v", rows, err)
	}
	r := structuralRecord(t, st)
	if r["quantity"] != float64(0) || r["reason"] != "rr" || r["stop"] != 77.0 || r["target"] != 104.0 {
		t.Fatalf("frozen refusal record wrong: %+v", r)
	}
}
