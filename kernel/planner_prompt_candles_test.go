package kernel

import (
	"strings"
	"testing"
	"time"

	"nofx/market"
)

// ── W2b — candle-table render + token budget ───────────────────────────────

func TestPlannerCandleTablesRenderAndTokenBudget(t *testing.T) {
	var bars []market.Kline
	start := time.Date(2026, 8, 24, 9, 0, 0, 0, CTLocation())
	for i := 0; i < 5000; i++ {
		ts := start.Add(time.Duration(i) * time.Minute)
		if ts.Weekday() == time.Saturday || ts.Weekday() == time.Sunday {
			continue
		}
		o := 30000.0 + float64(i)*0.01
		bars = append(bars, market.Kline{
			OpenTime: ts.UnixMilli(), Open: o, High: o + 1.5, Low: o - 1.5, Close: o + 0.5, Volume: 10,
		})
	}
	table := BuildPlannerCandleTables(bars)
	for _, want := range []string{"### 15m (last 12)", "### 1h (last 12)", "### 4h (last 8)",
		"### daily session candles (last 8)", "Time(CT)       Open      High      Low       Close     Volume"} {
		if !strings.Contains(table, want) {
			t.Fatalf("proving line: candle table missing %q", want)
		}
	}
	if lines := strings.Count(table, "\n"); lines < 40 {
		t.Fatalf("proving line: candle table too thin — %d lines", lines)
	}

	// Full session prompt render + token budget: chars/4 must be < 50% of 65536.
	prompt := BuildPlannerPrompt(PlannerInput{
		TradeDate:    "2026-08-31",
		Session:      "NY",
		Now:          time.Date(2026, 8, 31, 8, 30, 0, 0, CTLocation()),
		ReadKind:     "fixture read",
		Price:        30100.0,
		DATR:         120.0,
		CandleTables: table,
		WeeklyCtx:    "WEEKLY: refs only — PWH 30500.25 · PWL 29980.00",
	})
	if !strings.Contains(prompt, "## Candles (oldest→latest)") {
		t.Fatalf("proving line: prompt must carry the ## Candles section")
	}
	if !strings.Contains(prompt, "## Weekly Context") || !strings.Contains(prompt, "WEEKLY: refs only") {
		t.Fatalf("proving line: prompt must carry the ## Weekly Context section")
	}
	tokens := len(prompt) / 4
	budget := 65536 / 2
	if tokens >= budget {
		t.Fatalf("proving line: token budget — %d tokens ≥ %d (50%% of 65536); real number must be under", tokens, budget)
	}
	t.Logf("token count: %d chars → %d tokens (budget %d)", len(prompt), tokens, budget)
}

// ── W3 — fail-open no-doc render ───────────────────────────────────────────

func TestPlannerPromptNoWeeklyDocFailOpen(t *testing.T) {
	prompt := BuildPlannerPrompt(PlannerInput{
		TradeDate: "2026-08-31", Session: "NY",
		Now:   time.Date(2026, 8, 31, 8, 30, 0, 0, CTLocation()),
		Price: 30100.0, DATR: 120.0,
		// no CandleTables, no WeeklyCtx — the missing-doc state
	})
	if !strings.Contains(prompt, "## Weekly Context\nWEEKLY: none") {
		t.Fatalf("proving line: fail-open — no-doc renders the none form")
	}
	if !strings.Contains(prompt, "## OUTPUT — one JSON object") {
		t.Fatalf("proving line: fail-open — plan contract still renders, plan remains valid")
	}
}

// ── W2b/W3 playbook lines ──────────────────────────────────────────────────

func TestPlannerPlaybookLines(t *testing.T) {
	prompt := BuildPlannerPrompt(PlannerInput{
		TradeDate: "2026-08-31", Session: "NY",
		Now:   time.Date(2026, 8, 31, 8, 30, 0, 0, CTLocation()),
		Price: 30100.0, DATR: 120.0,
	})
	if !strings.Contains(prompt, "Candles are ground truth for structure; ranked levels and tags are summaries. On conflict, trust the candles and say so in the scenario rationale.") {
		t.Fatalf("proving line: candle ground-truth law missing from the playbook")
	}
	if !strings.Contains(prompt, "counter-weekly scenarios are allowed but must state their justification") {
		t.Fatalf("proving line: counter-weekly guidance missing from the playbook")
	}
}

// ── W2b — executor formatter extraction parity ─────────────────────────────

func TestFormatCandleTableVolumeFlag(t *testing.T) {
	var b strings.Builder
	klines := []market.KlineBar{
		{Time: 1725000000000, Open: 1, High: 2, Low: 0.5, Close: 1.5, Volume: 100},
		{Time: 1725000060000, Open: 1.5, High: 2.5, Low: 1, Close: 2, Volume: 120},
	}
	FormatCandleTable(&b, klines, true)
	if !strings.Contains(b.String(), "Volume") || !strings.Contains(b.String(), "<- current") {
		t.Fatalf("proving line: volume table renders header + current marker")
	}
	b.Reset()
	FormatCandleTable(&b, klines, false)
	if strings.Contains(b.String(), "Volume") || !strings.Contains(b.String(), "Close") {
		t.Fatalf("proving line: volume=false omits the Volume column")
	}
}
