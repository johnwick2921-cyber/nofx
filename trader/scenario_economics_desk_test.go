package trader

import (
	"encoding/json"
	"nofx/kernel"
	"os"
	"strings"
	"testing"
	"time"
)

func TestScenarioEconomicsOnProductionDesk(t *testing.T) {
	now := time.Date(2026, 9, 8, 19, 0, 0, 0, time.UTC)
	b, err := os.ReadFile("../kernel/testdata/scenario-economics/plan-265.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc kernel.PlanDoc
	json.Unmarshal(b, &doc)
	s := doc.Scenarios[1]
	p, r1, r2 := 29671.42, 6.92/24.5, 56.75/24.5
	s.Economics = &kernel.ScenarioEconomics{FirstObstacle: &kernel.ScenarioObstacle{Price: &p, Level: "reference", Family: "reference", Response: "pass_through"}, RToObstacle: &r1, RToArmTarget: &r2}
	doc.Scenarios = []kernel.PlanScenario{s}
	plan := &kernel.ActivePlan{Doc: doc, PlanID: "2026-09-08:NY:econ-test", Session: "NY", Version: 2, BirthMs: now.Add(-time.Minute).UnixMilli()}
	kernel.SetTraderPlanProviders("econ-desk", kernel.TraderPlanProviders{ActivePlan: func(string) *kernel.ActivePlan { return plan }})
	defer kernel.SetTraderPlanProviders("econ-desk", kernel.TraderPlanProviders{})
	at := &AutoTrader{id: "econ-desk", store: deskStore(t), config: AutoTraderConfig{NinjaTraderSymbol: "MNQ"}}
	strip := at.DeskStripAt(now)
	found := false
	for _, l := range strip.Lines {
		if l.Key != "scenarios" {
			continue
		}
		found = true
		if !strings.Contains(l.Text, "obstacle R=0.282449") || !strings.Contains(l.Text, "arm R=2.316327") {
			t.Fatalf("Rs absent on production desk: %s", l.Text)
		}
	}
	if !found {
		t.Fatal("economics helper is not wired into DeskStripAt")
	}
	plan.Doc.Scenarios[0].Economics = nil
	for _, l := range at.DeskStripAt(now).Lines {
		if l.Key == "scenarios" && (l.State != "unknown" || !strings.Contains(l.Text, "arm R=UNKNOWN")) {
			t.Fatalf("legacy inferred: %+v", l)
		}
	}
}
