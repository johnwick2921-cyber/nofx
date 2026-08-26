package kernel

import (
	"strings"
	"testing"

	"nofx/market"
)

func confirmBars(base int64, closes ...float64) []market.Kline {
	var out []market.Kline
	for i, cl := range closes {
		o := base + int64(i)*60_000
		out = append(out, market.Kline{OpenTime: o, CloseTime: o + 59_999, Open: cl, High: cl + 1, Low: cl - 1, Close: cl})
	}
	return out
}

// C1 — authored rule == evaluated rule (with A2's semantics): 1x5m_close needs
// ONE 5m bucket beyond; 2x5m_close needs two; 15m_close one 15m bucket.
func TestConfirmRuleIdentity(t *testing.T) {
	base := int64(1_700_000_100_000)
	base -= base % 900_000 // 15m-aligned
	// ten 1m bars: first 5m bucket closes below 100, second closes above.
	bars := confirmBars(base, 99, 99, 99, 99, 99, 101, 101, 101, 101, 101)
	oneBelow := PlanConfirm{Rule: "1x5m_close", RefPrice: 100, Side: "below"}
	if v := EvaluateConfirm(oneBelow, bars, base-1, base+10*60_000); !v.Met {
		t.Fatalf("1x5m_close below: one 5m close beyond must be MET (%s)", v.Detail)
	}
	twoBelow := PlanConfirm{Rule: "2x5m_close", RefPrice: 100, Side: "below"}
	if v := EvaluateConfirm(twoBelow, bars, base-1, base+10*60_000); v.Met {
		t.Fatalf("2x5m_close below: ONE close must NOT satisfy two (%s)", v.Detail)
	}
	touch := PlanConfirm{Rule: "touch", RefPrice: 100, Side: "below"}
	if v := EvaluateConfirm(touch, bars, base-1, base+10*60_000); !v.Met {
		t.Fatalf("touch: highs/lows straddle 100 (%s)", v.Detail)
	}
}

// The MET detail line carries the last rule-TF close (the executor-prompt fact).
func TestConfirmDetailCarriesLastClose(t *testing.T) {
	base := int64(1_700_003_600_000)
	base -= base % 900_000
	bars := confirmBars(base, 99, 99, 99, 99, 99)
	v := EvaluateConfirm(PlanConfirm{Rule: "1x5m_close", RefPrice: 100, Side: "below"}, bars, base-1, base+5*60_000)
	if !strings.Contains(v.Detail, "close") {
		t.Fatalf("detail must state the last close: %q", v.Detail)
	}
}

// Validator: confirm enum + object↔prose number agreement (the A3 contract).
func TestConfirmValidator(t *testing.T) {
	d := validBaseDoc()
	d.Scenarios[0].Confirm = &PlanConfirm{Rule: "15m_close", RefPrice: 29648.25, Side: "above"}
	if err := ValidatePlanDocWithCaps(d, 8, 3); err != nil {
		t.Fatalf("coherent confirm must pass: %v", err)
	}
	d.Scenarios[0].Confirm.Rule = "3x1m"
	if err := ValidatePlanDocWithCaps(d, 8, 3); err == nil || !strings.Contains(err.Error(), "confirm.rule") {
		t.Fatalf("bad rule must fail (got %v)", err)
	}
	d.Scenarios[0].Confirm.Rule = "15m_close"
	d.Scenarios[0].Confirm.RefPrice = 29000 // not in the prose
	if err := ValidatePlanDocWithCaps(d, 8, 3); err == nil || !strings.Contains(err.Error(), "confirm.ref_price") {
		t.Fatalf("prose mismatch must fail (got %v)", err)
	}
}

// The prompt advisory renders one line per confirmed scenario, none when absent.
func TestRenderConfirmLines(t *testing.T) {
	doc := *validBaseDoc()
	if RenderConfirmLines(doc, nil, 0, 0, 0, 0) != "" {
		t.Fatal("no confirm objects → no advisory block")
	}
	doc.Scenarios[0].Confirm = &PlanConfirm{Rule: "15m_close", RefPrice: 29648.25, Side: "above"}
	out := RenderConfirmLines(doc, confirmBars(1_700_000_100_000-(1_700_000_100_000%900_000), 99, 99, 99), 1, 1_700_000_400_000, 0, 0)
	if !strings.Contains(out, "S1 confirm:") || !strings.Contains(out, "advisory") {
		t.Fatalf("advisory block malformed:\n%s", out)
	}
}

// ADDENDUM S (2026-08-26) — the stale parenthetical math: |now − ref| must be
// STRICTLY greater than STALE_CONFIRM_ATR (default 1.0) × dATR; NOT MET, zero
// dATR, or a missing context suppresses the annotation.
func TestStaleConfirmAnnotationMath(t *testing.T) {
	if got := StaleConfirmATR(); got != 1.0 {
		t.Fatalf("default STALE_CONFIRM_ATR must be 1.0, got %v", got)
	}
	met := func(ref float64) ConfirmVerdict {
		return ConfirmVerdict{Rule: "1x5m_close", RefPrice: ref, Side: "above", Met: true, Detail: "d"}
	}
	s := PlanScenario{ID: "S1", Direction: "long"}
	// 2.0 pts drift vs 1.0 × 1.5 = 1.5 pts threshold → stale.
	if a := staleConfirmAnnotation(s, met(100), 102.0, 1.5); !strings.Contains(a, "stale — written 100.00 context, price now 102.00; treat as expired") {
		t.Fatalf("2.0 > 1.5 must be stale, got %q", a)
	}
	// 1.4 ≤ 1.5 → NOT stale (strict inequality).
	if a := staleConfirmAnnotation(s, met(100), 101.4, 1.5); a != "" {
		t.Fatalf("1.4 ≤ 1.5 must NOT be stale, got %q", a)
	}
	// dATR = 0 (no bars / caller without levels) suppresses the annotation.
	if a := staleConfirmAnnotation(s, met(100), 120, 0); a != "" {
		t.Fatalf("dATR=0 must suppress, got %q", a)
	}
	// NOT MET never gets the stale tag.
	notMet := ConfirmVerdict{Rule: "1x5m_close", RefPrice: 100, Side: "above", Met: false, Detail: "d"}
	if a := staleConfirmAnnotation(s, notMet, 120, 1.0); a != "" {
		t.Fatalf("NOT MET must not be annotated, got %q", a)
	}
}

// ADDENDUM S — render: two opposite-direction MET confirms with drifted price
// get the stale parenthetical AND one CONFLICT trailer; a single-direction MET
// near its ref renders clean (no CONFLICT, no stale).
func TestRenderConfirmLinesStaleConflict(t *testing.T) {
	base := int64(1_700_003_600_000)
	base -= base % 300_000 // 5m-aligned
	bars := confirmBars(base, 105, 105, 105, 105, 105) // above 100 AND below 110
	doc := PlanDoc{Scenarios: []PlanScenario{
		{ID: "S1", Direction: "long", Confirm: &PlanConfirm{Rule: "1x5m_close", RefPrice: 100, Side: "above"}},
		{ID: "S3", Direction: "short", Confirm: &PlanConfirm{Rule: "1x5m_close", RefPrice: 110, Side: "below"}},
	}}
	out := RenderConfirmLines(doc, bars, base-1, base+5*60_000, 103.0, 1.5)
	for _, want := range []string{
		"S1 confirm:", "S3 confirm:",
		"stale — written 100.00 context, price now 103.00; treat as expired",
		"stale — written 110.00 context, price now 103.00; treat as expired",
		"CONFLICT: opposing confirms MET — structural ambiguity, default WAIT unless fresh trigger",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("render missing %q:\n%s", want, out)
		}
	}

	// single direction, price still near its ref → clean render.
	doc2 := PlanDoc{Scenarios: []PlanScenario{
		{ID: "S1", Direction: "long", Confirm: &PlanConfirm{Rule: "1x5m_close", RefPrice: 100, Side: "above"}},
	}}
	out2 := RenderConfirmLines(doc2, bars, base-1, base+5*60_000, 100.5, 1.5)
	if strings.Contains(out2, "CONFLICT:") || strings.Contains(out2, "stale —") {
		t.Fatalf("single-direction near-ref render must be clean:\n%s", out2)
	}
}
