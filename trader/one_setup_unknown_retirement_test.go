package trader

import (
	"encoding/json"
	"nofx/kernel"
	"nofx/market"
	"nofx/store"
	"testing"
	"time"
)

func TestUnknownOneSetupVerdictCannotPlaceOldAuthorization(t *testing.T) {
	for _, tc := range []struct{ failRetirement, lowQuality bool }{{}, {failRetirement: true}, {lowQuality: true}} {
		failRetirement := tc.failRetirement
		name := "missing verdict"
		if tc.lowQuality {
			name = "current quality refused"
		}
		if failRetirement {
			name = "retirement write fails"
		}
		t.Run(name, func(t *testing.T) {
			now := time.Date(2026, 9, 11, 15, 0, 0, 0, time.UTC)
			cfg := store.StrategyConfig{DayPlan: &store.DayPlanConfig{PlanEnabled: true}}
			cfg.RiskControl.MinRiskRewardRatio = 2
			if tc.lowQuality {
				cfg.DayPlan.MinScenarioQuality = "A"
			}
			structuralTestPolicy(&cfg, .5)
			at, st, sigs, _ := shadowWireHarnessAt(t, cfg, now)
			doc := kernel.PlanDoc{Bias: kernel.PlanBias{Direction: "long", Conviction: "low", FlipCondition: "n/a"}, Levels: []kernel.PlanLevel{{Price: 100, Label: "PDL", Grade: "A", Instruction: "fade"}}, Scenarios: []kernel.PlanScenario{{ID: "S1", Trigger: "t", Condition: "reject", Direction: "long", TargetChain: []float64{110}, Invalid: "i", Quality: "B", Confirm: &kernel.PlanConfirm{Rule: "touch", RefPrice: 100, Side: "above"}, Arm: &kernel.PlanArmSpec{Enabled: true, Entry: 100, Stop: 95, Target: 110}}}, NoTrade: []string{}, DeathCondition: "n/a"}
			structuralTestMap(&doc, structuralTestZone{100, 95.5, 100, "PDL"}, structuralTestZone{110, 110, 111, "target"})
			blob, _ := json.Marshal(doc)
			pid := shadowPlanAtTime(t, at, st, string(blob), now)
			row := store.ArmedOrderDB{TraderID: at.id, PlanID: pid, Version: 1, Session: "X", Scenario: "S1", Side: "long", EntryPx: 100, StopPx: 95, TargetPx: 110, State: store.StateArmed, EntryClass: "armed_fill", Condition: "reject"}
			if err := st.ArmedOrders().UpsertArm(&row); err != nil {
				t.Fatal(err)
			}
			if failRetirement {
				if err := st.GormDB().Exec("CREATE TRIGGER fail_unknown_retire BEFORE UPDATE OF state ON armed_orders WHEN NEW.state = 'cancelled' BEGIN SELECT RAISE(ABORT, 'synthetic retirement failure'); END;").Error; err != nil {
					t.Fatal(err)
				}
			}
			at.oneSetupFactsForTest = func(time.Time) oneSetupTestFacts {
				if !tc.lowQuality {
					panic("synthetic unavailable permission facts")
				}
				id := *structuralTestIdentity(100, "PDL").ID
				return oneSetupTestFacts{Price: 100, BandPts: 50, Candidates: []kernel.MapCandidate{{ID: &id, Identity: kernel.PlanLevel{ID: &id, Price: 100}, Price: 100, Names: []string{"PDL"}, Grade: "A"}}, Permission: map[string]kernel.FadeVerdict{"S1": {Evaluated: true, Permitted: true}}}
			}
			previous := market.FuturesBarsProvider
			market.FuturesBarsProvider = func(string, string, int) []market.Kline { return shadowBarsNearAt(100, now) }
			t.Cleanup(func() { market.FuturesBarsProvider = previous })
			at.maybeManageArmedOrdersAt(nil, now)
			if err := st.GormDB().First(&row, row.ID).Error; err != nil {
				t.Fatal(err)
			}
			if row.SignalID != "" {
				t.Fatalf("unknown verdict registered old authorization: %+v", row)
			}
			if !failRetirement && !tc.lowQuality && row.State != store.StateCancelled {
				t.Fatalf("unknown authorization not retired: %s", row.State)
			}
			select {
			case s := <-sigs:
				t.Fatalf("unknown verdict sent signal: %+v", s)
			case <-time.After(30 * time.Millisecond):
			}
		})
	}
}
