// deep-verify-22 trader fixtures (package trader — unexported armGateVerdict).
package trader

import (
	"os"
	"strings"
	"testing"
	"time"

	"nofx/kernel"
	ntwire "nofx/provider/ninjatrader"
	"nofx/store"
)

// G1 1.1 arm side — a 2.5-R arm passes under the default ARM_MIN_RR (2.0) while
// the GLOBAL config floor is 4.0; ARM_MIN_RR=3 flips ONLY the arm. Twin.
func TestDeepG1ArmFloorIsolation(t *testing.T) {
	cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}
	cfg.RiskControl = store.RiskControlConfig{MinRiskRewardRatio: 4.0}
	at, _ := resetTrader(t, cfg)
	for _, side := range []struct {
		dir                  string
		entry, stop, target  float64
	}{
		{"long", 100, 96, 110},  // R 2.5
		{"short", 100, 104, 90}, // R 2.5
	} {
		sc := kernel.PlanScenario{ID: "S1", Condition: "reject", Direction: side.dir, Quality: "A",
			Arm: &kernel.PlanArmSpec{Enabled: true, Entry: side.entry, Stop: side.stop, Target: side.target}}
		// default ARM_MIN_RR=2.0: passes even though the global floor is 4.0
		os.Unsetenv("ARM_MIN_RR")
		if v := at.armGateVerdict(sc, side.dir, nil, 0, "", at.config.StrategyConfig); v != "" {
			t.Fatalf("%s R=2.5 arm must pass under default ARM_MIN_RR despite global 4.0: %q", side.dir, v)
		}
		// env override flips ONLY the arm gate
		t.Setenv("ARM_MIN_RR", "3")
		if v := at.armGateVerdict(sc, side.dir, nil, 0, "", at.config.StrategyConfig); v == "" || !strings.Contains(v, "arm min") {
			t.Fatalf("%s R=2.5 arm must be refused at ARM_MIN_RR=3: %q", side.dir, v)
		}
	}
}

// G2 2.3 — armed fill chain: order_update filled → ledger filled + entry_class
// armed_fill survives (the stale_reeval exemption marker).
func TestDeepG2ArmedFillChain(t *testing.T) {
	at, st := resetTrader(t, store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}})
	ledger := st.ArmedOrders()
	row := &store.ArmedOrderDB{TraderID: at.id, PlanID: "2026-08-27:NY:trader-1", Version: 1, Session: "NY",
		Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: "working",
		SignalID: "sig-1", EntryClass: "armed_fill", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := ledger.UpsertArm(row); err != nil {
		t.Fatal(err)
	}
	at.onArmedOrderUpdate(ntwire.OrderUpdatePayload{SignalID: "sig-1", State: "filled", FillPrice: 100.5}, ledger)
	rows, err := ledger.ListForPlan("2026-08-27:NY:trader-1")
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}
	if rows[0].State != "filled" || rows[0].EntryClass != "armed_fill" {
		t.Fatalf("armed fill lineage wrong: state=%s class=%s", rows[0].State, rows[0].EntryClass)
	}
}
