package kernel

import (
	"strings"
	"testing"
)

// P0-relax (2026-08-27) — the machine-aware side-quota validator. Owner ruling:
// ≥3 fail-closed the whole 08-26 ASIA session over a machine-map shortage.
// Matrix: machine-thin → WARN+write · AI-omission → reject · zero-side → fail ·
// empty map → fail · nil map → legacy hard fail · HTF+owner rows counted ·
// knob 2 vs 3 · above/below symmetry.

func quotaDoc(below, above int, price float64) *PlanDoc {
	d := &PlanDoc{
		Reasoning:      "quota test",
		Bias:           PlanBias{Direction: "neutral", Conviction: "low", FlipCondition: "none"},
		DeathCondition: "all levels consumed",
		Scenarios: []PlanScenario{
			{ID: "S1", Trigger: "hold", Condition: "hold", Direction: "long",
				TargetChain: []float64{price + 20}, Invalid: "2x5m below", Quality: "B"},
		},
	}
	for i := 0; i < below; i++ {
		d.Levels = append(d.Levels, PlanLevel{Price: price - float64(10*(i+1)), Label: "RN", Grade: "B", Instruction: "reclaim"})
	}
	for i := 0; i < above; i++ {
		d.Levels = append(d.Levels, PlanLevel{Price: price + float64(10*(i+1)), Label: "RN", Grade: "B", Instruction: "fade"})
	}
	return d
}

func machineMap(prices ...float64) map[float64]string {
	m := make(map[float64]string, len(prices))
	for _, p := range prices {
		m[p] = "RN"
	}
	return m
}

func TestSideQuotaMachineThinWarnAndProceed(t *testing.T) {
	// Plan 1 above / 3 below; machine map also has only 1 above → machine-caused.
	d := quotaDoc(3, 1, 100)
	thin, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 97, 101), 2, 8, 3)
	if err != nil {
		t.Fatalf("machine-thin must WARN, not fail: %v", err)
	}
	if len(thin) != 1 || !strings.Contains(thin[0], "above") {
		t.Fatalf("expected one 'above' thin note, got %v", thin)
	}
	if !strings.Contains(thin[0], "machine map 1") {
		t.Fatalf("note must carry the machine count, got %q", thin[0])
	}
}

func TestSideQuotaMachineThinBelowSymmetry(t *testing.T) {
	// Mirrored: thin BELOW must behave identically (long/short symmetry).
	d := quotaDoc(1, 3, 100)
	thin, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 101, 102, 103), 2, 8, 3)
	if err != nil {
		t.Fatalf("machine-thin below must WARN, not fail: %v", err)
	}
	if len(thin) != 1 || !strings.Contains(thin[0], "below") {
		t.Fatalf("expected one 'below' thin note, got %v", thin)
	}
}

func TestSideQuotaAICausedOmissionRejected(t *testing.T) {
	// Plan 1 above but the machine map offered 3 above → the AI dropped levels.
	d := quotaDoc(3, 1, 100)
	_, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 97, 101, 102, 103), 2, 8, 3)
	if err == nil || !strings.Contains(err.Error(), "AI dropped levels") {
		t.Fatalf("AI-caused omission must be rejected, got err=%v", err)
	}
}

func TestSideQuotaZeroSideFails(t *testing.T) {
	// 0 on a side = the 2026-08-18 one-sided pathology — always a hard fail,
	// even when the machine map has rows on that side.
	d := quotaDoc(4, 0, 100)
	if _, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 97, 96, 101), 2, 8, 3); err == nil || !strings.Contains(err.Error(), "0 levels above") {
		t.Fatalf("0-above plan must fail, got err=%v", err)
	}
	d = quotaDoc(0, 4, 100)
	if _, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(101, 102, 103, 104, 99), 2, 8, 3); err == nil || !strings.Contains(err.Error(), "0 levels below") {
		t.Fatalf("0-below plan must fail, got err=%v", err)
	}
}

func TestSideQuotaEmptyMachineMapFails(t *testing.T) {
	d := quotaDoc(3, 3, 100)
	if _, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		map[float64]string{}, 2, 8, 3); err == nil || !strings.Contains(err.Error(), "EMPTY") {
		t.Fatalf("empty machine map must fail-closed, got err=%v", err)
	}
}

func TestSideQuotaNilMachineLegacyHard(t *testing.T) {
	// Legacy callers (nil map) keep the pre-relax hard behavior at the quota.
	d := quotaDoc(2, 3, 100)
	if _, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		nil, 3, 8, 3); err == nil || !strings.Contains(err.Error(), "only 2 levels below") {
		t.Fatalf("legacy nil-map must hard-fail below quota, got err=%v", err)
	}
	// At quota → pass.
	d = quotaDoc(3, 3, 100)
	if thin, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		nil, 3, 8, 3); err != nil || len(thin) != 0 {
		t.Fatalf("legacy at-quota must pass cleanly, thin=%v err=%v", thin, err)
	}
}

func TestSideQuotaHTFAndOwnerRowsCounted(t *testing.T) {
	// Plan 1 above (101) / 3 below. Machine map above = 101 (seated) + 105
	// (HTF-section row) + 110 (owner-sticky row). With HTF+owner counted the
	// map has 3 above → AI omission → REJECT. If they were ignored the map
	// would show 1 above → machine-thin WARN. The reject proves they count.
	d := quotaDoc(3, 1, 100)
	_, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 97, 101, 105, 110), 2, 8, 3)
	if err == nil || !strings.Contains(err.Error(), "AI dropped levels") {
		t.Fatalf("HTF+owner rows must count toward the machine map (expected reject), got err=%v", err)
	}
}

func TestSideQuotaKnobTwoVsThree(t *testing.T) {
	// Quota 2: a plan with 2/2 against a 2/2 map passes with no notes.
	d := quotaDoc(2, 2, 100)
	thin, err := ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 101, 102), 2, 8, 3)
	if err != nil || len(thin) != 0 {
		t.Fatalf("quota 2 with 2/2 must pass cleanly, thin=%v err=%v", thin, err)
	}
	// Quota 3 (old behavior): the same plan now fails — machine thin on both sides.
	_, err = ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 101, 102), 3, 8, 3)
	if err != nil {
		// With quota 3 and machine map 2/2, both sides are MACHINE-thin → WARN
		// notes, not a fail (the old hard-3 is restored only via the legacy nil
		// path; the knob keeps the machine-aware WARN semantics).
		t.Fatalf("quota 3 with 2/2 machine map must still WARN (not fail): %v", err)
	}
	// Quota 3 with machine map 3/3 but plan 2/2 → AI omission → reject.
	_, err = ValidatePlanDocWithFactsMachine(d, PlanFacts{Price: 100, DATR: 50},
		machineMap(99, 98, 97, 101, 102, 103), 3, 8, 3)
	if err == nil || !strings.Contains(err.Error(), "AI dropped levels") {
		t.Fatalf("quota 3 with a full map and a 2/2 plan must reject, got err=%v", err)
	}
}

func TestSideQuotaFromEnv(t *testing.T) {
	t.Setenv("MIN_SIDE_LEVELS", "")
	if got := SideQuotaFromEnv(); got != DefaultSideQuota {
		t.Fatalf("unset → default %d, got %d", DefaultSideQuota, got)
	}
	t.Setenv("MIN_SIDE_LEVELS", "3")
	if got := SideQuotaFromEnv(); got != 3 {
		t.Fatalf("env 3 → 3, got %d", got)
	}
	t.Setenv("MIN_SIDE_LEVELS", "garbage")
	if got := SideQuotaFromEnv(); got != DefaultSideQuota {
		t.Fatalf("garbage → default %d, got %d", DefaultSideQuota, got)
	}
	t.Setenv("MIN_SIDE_LEVELS", "0")
	if got := SideQuotaFromEnv(); got != DefaultSideQuota {
		t.Fatalf("0 → default %d, got %d", DefaultSideQuota, got)
	}
}
