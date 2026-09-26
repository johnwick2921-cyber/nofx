package trader

import (
	"encoding/json"
	"testing"

	"nofx/kernel"
)

// FIX-PLANNER (2026-09-25) item 1 — scenario-level born-dead salvage at the
// production call site (runPlannerReadCoreObserved → validateAuthoredScenariosAt).
// The breach fixture is the same row455/row452 tape as plan_born_check_writesite_test.go;
// "5m close above 31085.00" / "2x5m close above 31080.00" are PROVEN breached
// between read 01:30:27 and publish 01:51:47 (TestW2A2RefusesOnlyWithTheReadClock),
// and the conformant sentences above 31095/31100 are proven clean.

// RED (before the fix): this read is refused born-dead on attempt 1 and burns
// attempts — the surviving scenarios must publish instead, with S1 dropped.
func TestFpSalvageDropsBornDeadScenarioAndPublishesTheRest(t *testing.T) {
	_, bars, read, publish := loadW2WriteFixture(t)
	at := w2Trader(t, bars)
	cand := w2Candidate(t, []string{
		"5m close above 31085.00",    // breached between read and publish → born dead
		"5m close above 31095.00",    // clean
		"2x5m close above 31100.00",  // clean
	}, nil)
	ver, lc, err, prompts := w2Run(at, "LONDON", "2026-09-23", read, publish, cand)
	if err != nil || ver != 1 || lc != "active" || len(prompts) != 1 {
		t.Fatalf("salvage must publish on attempt 1: ver=%d lc=%s calls=%d err=%v", ver, lc, len(prompts), err)
	}
	row, _ := at.store.Plan().GetLatestPlanForSession("2026-09-23", "LONDON")
	if row == nil {
		t.Fatal("no published row")
	}
	var pub kernel.PlanDoc
	if json.Unmarshal([]byte(row.Doc), &pub) != nil {
		t.Fatalf("row doc unparsable: %.120s", row.Doc)
	}
	ids := map[string]bool{}
	for _, s := range pub.Scenarios {
		ids[s.ID] = true
	}
	if len(pub.Scenarios) != 2 || ids["S1"] || !ids["S2"] || !ids["S3"] {
		t.Fatalf("S1 must be dropped and S2/S3 published: got %d scenarios %v", len(pub.Scenarios), ids)
	}
	if c, _ := at.store.PlanLivenessCounts(); c.BornDeadRefusals != 0 {
		t.Fatalf("a salvaged read is NOT a whole-read refusal: %+v", c)
	}
	if row.BornCheck == nil {
		t.Fatal("the born check record must still be stored")
	}
	var bc kernel.BornCheck
	if json.Unmarshal([]byte(*row.BornCheck), &bc) != nil {
		t.Fatalf("born check unparsable: %s", *row.BornCheck)
	}
	s1Judged := false
	for _, v := range bc.Verdicts {
		if v.Subject == "S1" && v.Outcome == "invalidated" {
			s1Judged = true
		}
	}
	if !s1Judged {
		t.Fatalf("the record must show S1 was judged invalidated: %s", *row.BornCheck)
	}
	// The 💀 drop is RECORDED, one event per dropped scenario.
	if c, _ := at.store.PlanLivenessCounts(); c.BornDeadDropped != 1 {
		t.Fatalf("exactly one born-dead dropped event must be recorded: %+v", c)
	}
}

// RED (before the fix): both scenarios die, attempt 2 dies again — today that
// still burns attempt 3. The fix: ONE fresh-tape repair, then fail-closed.
func TestFpAllDeadDiesAndBurnsOnlyOneFreshTapeRepair(t *testing.T) {
	_, bars, read, publish := loadW2WriteFixture(t)
	at := w2Trader(t, bars)
	dead := w2Candidate(t, []string{"5m close above 31085.00", "2x5m close above 31080.00"}, nil)
	_, lc, err, prompts := w2Run(at, "LONDON", "2026-09-23", read, publish, dead, dead)
	if err != nil || lc != "no_trade" || len(prompts) != 2 {
		t.Fatalf("all-dead read must die after ONE fresh-tape repair: lc=%s calls=%d err=%v", lc, len(prompts), err)
	}
	if c, _ := at.store.PlanLivenessCounts(); c.BornDeadRefusals != 2 {
		t.Fatalf("one counted refusal per burned attempt: %+v", c)
	}
}

// The attempt-3 skip rides the SAME planner_fresh_tape knob as A6: OFF
// reproduces today's blind retry byte-identically (three calls, three refusals).
func TestFpFreshTapeOffBurnsAttempt3AsToday(t *testing.T) {
	_, bars, read, publish := loadW2WriteFixture(t)
	at := w2Trader(t, bars)
	off := false
	at.config.StrategyConfig.DayPlan.PlannerFreshTape = &off
	dead := w2Candidate(t, []string{"5m close above 31085.00", "2x5m close above 31080.00"}, nil)
	_, lc, _, prompts := w2Run(at, "LONDON", "2026-09-23", read, publish, dead, dead)
	if lc != "no_trade" || len(prompts) != 3 {
		t.Fatalf("knob OFF must reproduce today's blind retry: lc=%s calls=%d", lc, len(prompts))
	}
}
