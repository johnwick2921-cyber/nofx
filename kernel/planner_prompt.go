package kernel

import (
	"fmt"
	"strings"
	"time"
)

// P3.3 — the planner input package (assembled into ONE prompt for the reasoner).
// Pure + deterministic. The 16:55 closed-market read builds entirely from stored
// data (no live feed), which is what makes it a first-class tested path.

// PlannerCalendarEvent is a session-sliced calendar entry (kernel-local to avoid
// importing the calendar package; the trader maps calendar.Event → this).
type PlannerCalendarEvent struct {
	TimeCT   string // "HH:MM" CT
	Currency string
	Title    string
	Impact   string // T1 | T2
}

// PlannerInput is everything the reasoner reads to write a plan.
type PlannerInput struct {
	TradeDate        string
	Session          string    // NY | ASIA | LONDON
	Now              time.Time // labelled CT clock line (zero → omitted)
	ReadKind         string    // e.g. "closed-market 16:55 CT read (from stored data)"
	Price            float64
	DATR             float64
	Regime           RegimeBlock
	Levels           []ScoredLevel // Go-ranked, graded (P1.5) — the decision-critical block
	StructureSummary []string      // one line per timeframe
	// G2.2 (2026-08-24) — nearest in-band HTF zones (S/D/FVG/OB), graded, for a
	// dedicated prompt section. They exist in the data but lose the top-8 seat
	// race to structural levels (cluster collapse + seat priority), so the model
	// never saw them before. Advisory confluence — never standalone triggers.
	HTFZones []ScoredLevel
	// G5 (regime wave 2026-08-21) — levels already CONSUMED at read time (role-
	// flipped), listed so the planner works around them. Advisory.
	ConsumedLevels []string

	// FreshFVGs (level-truth wave b2, 2026-08-27) — the machine's fresh-gap
	// candidate list the planner may author fvg_entry from. Empty list means
	// NO fvg_entry may be authored (the write-site validator refuses anything
	// else).
	FreshFVGs []FreshFvg

	// Pool (level-truth wave, 2026-08-27) — the graded PRE-SEAT candidate pool
	// the assembly produced. NOT rendered: the write site uses it for the
	// machine-grade stamp map so pool levels that lost the seat race still get
	// stamped when the model copies them into the plan.
	Pool           []ScoredLevel
	OvernightStory string
	PriorDayStory  string
	Calendar       []PlannerCalendarEvent // session-sliced (P1.8)
	DigestChain    []string               // session digests + dailies + one-liners
	OwnerNote      string
	Warming        string // non-empty → cold-start / WARMING annotation
	// FastTape (F3 fast-market wake reads, 2026-08-28) — the read fires while
	// the tape is moving fast; the prompt carries the note and the wire runs
	// reasoning=medium (FAST_MARKET_REASONING).
	FastTape     bool
	FastTapeNote string
	// PriorPlanKiller (P0.4-G, 2026-08-25) — when this read re-plans a DEAD
	// plan, the killer line (e.g. "flip-condition: 2x5m close above X → bias
	// long"). Rendered as a MANDATORY context block: a flip that already fired
	// on machine-evaluated bars must be honored in the new plan's bias, not
	// silently re-written (live bug: ASIA v3 flip fired → bias long, but v4
	// came back short-biased). Empty → no block (first reads / owner resets).
	PriorPlanKiller string
	// PriorPlanLevels (P0.4-G) — the PREVIOUS version's levels, carried into
	// every re-plan so the map keeps continuity instead of being rebuilt from
	// scratch each time (the owner's complaint: every re-run dropped all old
	// levels and the flip anchor kept moving). Rendered as a carry-over block;
	// the model keeps or re-anchors them, never silently loses the set.
	PriorPlanLevels []string
	// W11 — the executor's indicator mirror (per-TF EMA/RSI/ATR/BOLL/MACD, driven by
	// ai_config toggles), rendered once by RenderPlannerIndicatorBlock. Empty → the
	// block is omitted (disabled state = byte-identical prompt). AIConfigHash is the
	// fingerprint of the indicator config that produced it (frozen on the plan row).
	IndicatorsBlock string
	AIConfigHash    string
	// H4/H5 — the RESOLVED caps the planner is asked to write within (max_levels,
	// scenario_cap). 0 → shipped defaults, so default callers render the same
	// "max 8 / 1..3" contract as before.
	MaxLevels   int
	ScenarioCap int
	// BiasCtx (addendum 2, 2026-08-26) — the per-cycle bias-context facts line
	// (price vs VWAP/PDC, value area, nearest magnet/liquidity). Facts only.
	BiasCtx string
	// BiasCtxFacts (A1, planner-contract wave 2026-08-26) — the STRUCTURED bias
	// context the BIAS-TREE section computes from (price, PDH/PDL/PDC, value
	// area, nearest liquidity). Filled by the same site as BiasCtx.
	BiasCtxFacts BiasContext
}

// RenderBiasTree (A1, planner-contract wave 2026-08-26) — the machine-computed
// bias decision tree. Facts only + the branch the facts CURRENTLY match; the
// planner states which branch it took (contract requirement). Premium/discount
// = position within the value-area range; the draw = nearest opposing
// liquidity pool (the runner target).
func RenderBiasTree(price float64, levels []ScoredLevel, bc BiasContext) string {
	// S-dispatch (2026-08-27) — the structured bias facts win: post-roll the
	// seated table can drop PDH/PDL (out of the proximity band) and the tree
	// rendered "no PDH/PDL anchor". Fall back to the seated scan for legacy
	// callers/tests that only pass levels.
	pdh, pdl, pdc := bc.PDH, bc.PDL, bc.PDC
	for _, l := range levels {
		switch l.Kind {
		case KindPDH:
			if pdh <= 0 {
				pdh = l.Price
			}
		case KindPDL:
			if pdl <= 0 {
				pdl = l.Price
			}
		case KindPDC:
			if pdc <= 0 {
				pdc = l.Price
			}
		}
	}
	var b strings.Builder
	b.WriteString("## BIAS-TREE (machine branches — your reasoning MUST state the branch you took, e.g. \"bias-tree: inside-day long LOW\")\n")
	b.WriteString("  1. close > PDH → bull-continuation, conviction HIGH\n")
	b.WriteString("  2. sweep of PDH + close back inside → bear, conviction MEDIUM  (mirror: sweep PDL + close inside → bull MEDIUM)\n")
	b.WriteString("  3. inside the day (between PDH/PDL) → direction of close vs PDC (prior close), conviction LOW\n")
	b.WriteString("  4. closed OUTSIDE the prior day's range but now inside → NO bias (write neutral; trade the structure, not a thesis)\n")
	b.WriteString("  5. premium/discount: longs ONLY below the 50% mark of the dealing range, shorts ONLY above it\n")
	b.WriteString("  6. draw-on-liquidity: the runner target is the DRAW — the nearest opposing liquidity pool beyond the first target\n")
	fmt.Fprintf(&b, "  computed: PDH %.2f · PDL %.2f · PDC %.2f", pdh, pdl, pdc)
	// S-dispatch (2026-08-27) — the branch-5 premium/discount anchor is the
	// DEALING RANGE (prior-day swing hi/lo = PDH/PDL), not the value area. The
	// VA can be a few points wide (the 17:46 read printed "376% of range" off
	// a ~30pt VA); a beyond-range price now renders "beyond range (extended)"
	// instead of a >100% percentage. Value area stays as the fallback when
	// the day anchors are unknown.
	lo, hi := pdl, pdh
	if lo <= 0 || hi <= 0 || hi <= lo {
		lo, hi = bc.VAL, bc.VAH
		if lo > hi {
			lo, hi = hi, lo
		}
	}
	if hi > lo && price > 0 {
		fmt.Fprintf(&b, " · dealing range %.2f–%.2f", lo, hi)
		pos := (price - lo) / (hi - lo)
		switch {
		case pos > 1:
			b.WriteString(" · price BEYOND range high (extended) — longs disallowed by branch 5 (premium)")
		case pos < 0:
			b.WriteString(" · price BEYOND range low (extended) — shorts disallowed by branch 5 (discount)")
		case pos >= 0.5:
			fmt.Fprintf(&b, " · price at %.0f%% of the range (PREMIUM — longs disallowed by branch 5)", pos*100)
		default:
			fmt.Fprintf(&b, " · price at %.0f%% of the range (DISCOUNT — shorts disallowed by branch 5)", pos*100)
		}
	}
	if price > 0 && pdh > 0 && price > pdh {
		b.WriteString(" · facts match branch 1 (close > PDH)")
	} else if price > 0 && pdl > 0 && price < pdl {
		b.WriteString(" · facts match branch 1 mirror (close < PDL)")
	} else if price > 0 && pdh > 0 && pdl > 0 && pdc > 0 && price < pdh && price > pdl {
		dir := "flat"
		if price > pdc {
			dir = "long"
		} else if price < pdc {
			dir = "short"
		}
		fmt.Fprintf(&b, " · facts match branch 3 (inside day; close %s PDC → %s LOW)", dir, dir)
	}
	if bc.NearestLiquidity != "" {
		b.WriteString(" · draw/liquidity: " + bc.NearestLiquidity)
	}
	b.WriteString("\n\n")
	return b.String()
}

// FantasyTargetWarnings (autopsy-response wave, 2026-08-27) — advisory WARN,
// never a fail: an armed scenario whose PLANNED R:R exceeds 6 is a
// fantasy-target flag (the 3.28–22.88 R losers in the refusal autopsy).
func FantasyTargetWarnings(doc PlanDoc) []string {
	var out []string
	for _, s := range doc.Scenarios {
		a := s.Arm
		if a == nil || !a.Enabled || a.Entry <= 0 || a.Stop <= 0 || a.Target <= 0 {
			continue
		}
		risk := a.Entry - a.Stop
		reward := a.Target - a.Entry
		if strings.EqualFold(strings.TrimSpace(s.Direction), "short") {
			risk = a.Stop - a.Entry
			reward = a.Entry - a.Target
		}
		if risk > 0 && reward/risk > 6.0 {
			out = append(out, fmt.Sprintf("%s: planned R %.1f (entry %.2f stop %.2f target %.2f) — fantasy-target flag", s.ID, reward/risk, a.Entry, a.Stop, a.Target))
		}
	}
	return out
}

// FvgDemandWarnings (grand-audit response F5, 2026-08-28) — WARN-only demand
// compliance: fresh machine FVGs existed AND at least one agreed with the plan
// bias AND the plan authored NO fvg_entry scenario → the plan owes a one-line
// reason. The reason is model prose (not reliably parseable), so this is a
// visibility WARN at write, never a fail.
func FvgDemandWarnings(doc PlanDoc, fresh []FreshFvg) []string {
	if len(fresh) == 0 {
		return nil
	}
	authored := false
	for _, s := range doc.Scenarios {
		if strings.EqualFold(strings.TrimSpace(s.Condition), "fvg_entry") {
			authored = true
			break
		}
	}
	if authored {
		return nil
	}
	bias := strings.ToLower(strings.TrimSpace(doc.Bias.Direction))
	match := 0
	for _, g := range fresh {
		if bias == "" || bias == "neutral" || strings.EqualFold(g.Direction, bias) {
			match++
		}
	}
	if match == 0 {
		return nil
	}
	return []string{fmt.Sprintf(
		"%d fresh FVG candidate(s) agree with the %q bias but no fvg_entry was authored — the plan must state why not (one line in reasoning)",
		match, bias)}
}

// ChainWarnings (A2, planner-contract wave) — validator WARN, never a fail:
// an fvg_entry scenario without a chain_after sweep precursor and whose origin
// level is not a fresh A/B zone is the bare-gap setup the research's raw-FVG
// null result warns about (no edge without the sweep → displacement chain).
func ChainWarnings(doc PlanDoc) []string {
	var out []string
	gradeOf := func(label string) string {
		for _, l := range doc.Levels {
			if strings.EqualFold(strings.TrimSpace(l.Label), strings.TrimSpace(label)) {
				return l.Grade
			}
		}
		return ""
	}
	byID := map[string]PlanScenario{}
	for _, s := range doc.Scenarios {
		byID[strings.ToUpper(strings.TrimSpace(s.ID))] = s
	}
	for _, s := range doc.Scenarios {
		if !strings.EqualFold(strings.TrimSpace(s.Condition), "fvg_entry") || s.Fvg == nil {
			continue
		}
		hasPrecursor := false
		if ref := strings.ToUpper(strings.TrimSpace(s.ChainAfter)); ref != "" {
			if p, ok := byID[ref]; ok && strings.EqualFold(strings.TrimSpace(p.Condition), "sweep_reclaim") {
				hasPrecursor = true
			}
		}
		originGrade := strings.ToUpper(gradeOf(s.Fvg.OriginLevel))
		if !hasPrecursor && originGrade != "A" && originGrade != "B" {
			out = append(out, fmt.Sprintf("fvg_entry %s has no sweep_reclaim precursor (chain_after) and origin %q is not a fresh A/B zone — the bare-gap setup has no standalone edge (raw-FVG null, 40k sample); consider chaining it after a sweep play", s.ID, s.Fvg.OriginLevel))
		}
	}
	return out
}

// BuildPlannerPrompt assembles the planner prompt: reasoning-first instruction,
// the input tables, and the schema-strict JSON contract. High-salience blocks
// (ranked levels, T1 blackouts) are positioned prominently per the contract.
func BuildPlannerPrompt(in PlannerInput) string {
	var b strings.Builder

	b.WriteString("# DAY-PLAN READER — CME MNQ futures\n")
	b.WriteString("You are a disciplined pro reading the market ONCE, carefully. Write ONE structured day plan for this session. Facts below are Go-computed; your job is JUDGMENT, not re-deriving the map.\n\n")

	fmt.Fprintf(&b, "## Session\ntrade_date %s · session %s · %s · price %.2f · dATR %.1f\n",
		in.TradeDate, in.Session, in.ReadKind, in.Price, in.DATR)
	if !in.Now.IsZero() {
		fmt.Fprintf(&b, "clock %s — EVERY time in this prompt is CT (America/Chicago): session windows, read/flat times and the lunch no-trade (12:00–13:30 CT) are CT wall-clock. Never apply these numbers to a UTC clock.\n",
			ClockCTAndUTC(in.Now))
	}
	if in.Warming != "" {
		fmt.Fprintf(&b, "WARMING: %s (first-week honesty — narrate the machinery, not an edge).\n", in.Warming)
	}
	if in.FastTape {
		fmt.Fprintf(&b, "⚡ FAST TAPE — %s. Write the SHORTEST valid plan you can: fewer scenarios, tight prose. A fast market does not reward long deliberation.\n", in.FastTapeNote)
	}
	b.WriteString("\n")

	b.WriteString("## Regime\n" + in.Regime.Render() + "\n\n")

	// W11 — INDICATORS mirror: the SAME per-timeframe indicator state the executor
	// sees, so the planner reasons on the exact indicators the trader trades on.
	// Omitted when no indicator is enabled (disabled state = pre-W11 prompt).
	if in.IndicatorsBlock != "" {
		b.WriteString("## Indicators (executor mirror — your ai_config toggles, computed values)\n")
		b.WriteString(in.IndicatorsBlock + "\n\n")
	}

	// Ranked level table — the decision-critical block, high-salience.
	b.WriteString("## Ranked levels (Go-graded; you never re-sort)\n")

	// G5 (regime wave 2026-08-21) — consumed levels listed explicitly: the
	// planner must plan AROUND them (a re-test is a NEW setup, never a fresh
	// tag). Advisory — Go facts, AI judgment.
	if len(in.ConsumedLevels) > 0 {
		b.WriteString("## Consumed levels (already role-flipped — plan AROUND them; a re-test is a NEW setup)\n")
		for _, s := range in.ConsumedLevels {
			b.WriteString("- " + s + "\n")
		}
		b.WriteString("\n")
	}

	// Level-truth wave b2 (2026-08-27) — the machine's fresh-gap candidates.
	// The write-site validator re-checks every declared fvg{} against exactly
	// this list; the model must author ONLY from it (it used to invent stale
	// gaps and every read failed closed).
	b.WriteString("## FRESH FVGs (machine-computed candidates — author fvg_entry ONLY from this list; if empty, do NOT author any fvg_entry; if NON-empty and a candidate's direction agrees with your bias, an fvg_entry is EXPECTED unless you state why not in one line)\n")
	if len(in.FreshFVGs) == 0 {
		b.WriteString("(none fresh right now)\n\n")
	} else {
		for _, g := range in.FreshFVGs {
			b.WriteString(fmt.Sprintf("  %s %.2f–%.2f (age %d bars, displacement %.2f×ATR5m)\n",
				g.Direction, g.Lo, g.Hi, g.AgeBars, g.DispATR))
		}
		b.WriteString("\n")
	}
	if len(in.Levels) == 0 {
		b.WriteString("(none in range — warming forward)\n")
	} else {
		for _, l := range in.Levels {
			sign := "+"
			if l.Distance < 0 {
				sign = "-"
			}
			label := l.Label
			if isHTFSwingZone(l) {
				label = label + " (HTF)"
			}
			role := string(l.Role)
			if role == "" {
				role = string(RoleReactZone)
			}
			fmt.Fprintf(&b, "  %-9.2f %-20s grade %s  %-8s %-15s %s%.1f\n", l.Price, label, l.Grade, l.Fresh, role, sign, absF(l.Distance))
		}
	}
	b.WriteString("\n")

	// ADDENDUM (1) — the role playbook (machine facts; your judgment stays).
	b.WriteString("## Level roles (machine-assigned, 5-line playbook)\n")
	b.WriteString(RoleLegend + "\n")

	// ADDENDUM (2) — bias-context: facts only, AI judges.
	if in.BiasCtx != "" {
		b.WriteString(in.BiasCtx + "\n\n")
	}

	if len(in.StructureSummary) > 0 {
		b.WriteString("## Structure (1 line/TF)\n")
		for _, s := range in.StructureSummary {
			b.WriteString("  " + s + "\n")
		}
		b.WriteString("\n")
	}

	// G2.2 (2026-08-24) — HTF zones as their own section: the top-8 seat race
	// hides them from the ranked table, but a 1h/4h base is exactly the
	// large-account reference the plan should know about.
	if len(in.HTFZones) > 0 {
		b.WriteString("## HTF zones (nearest first — confluence references, NEVER standalone triggers)\n")
		for _, z := range in.HTFZones {
			sign := "+"
			if z.Distance < 0 {
				sign = "-"
			}
			fmt.Fprintf(&b, "  %-9.2f %-14s grade %s  %s%.1f (HTF zone)\n", z.Price, z.Label, z.Grade, sign, absF(z.Distance))
		}
		b.WriteString("\n")
	}

	if in.OvernightStory != "" || in.PriorDayStory != "" {
		b.WriteString("## Auction story\n")
		if in.PriorDayStory != "" {
			b.WriteString("  prior day: " + in.PriorDayStory + "\n")
		}
		if in.OvernightStory != "" {
			b.WriteString("  overnight: " + in.OvernightStory + "\n")
		}
		b.WriteString("\n")
	}

	// Session-sliced calendar: T1 = HARD blackout, T2 = caution. Times CT.
	b.WriteString("## Calendar (this session's window — times CT)\n")
	if len(in.Calendar) == 0 {
		b.WriteString("  (no filtered events)\n")
	} else {
		for _, e := range in.Calendar {
			tag := "caution — NOT a no-trade blackout; keep trading with normal discretion"
			if e.Impact == "T1" {
				tag = "HARD no-trade blackout — MUST be added to no_trade"
			}
			fmt.Fprintf(&b, "  %s %s %s — %s (%s)\n", e.TimeCT, e.Currency, e.Impact, e.Title, tag)
		}
	}
	b.WriteString("\n")

	if len(in.DigestChain) > 0 {
		b.WriteString("## Recent context (digest chain)\n")
		for _, d := range in.DigestChain {
			b.WriteString("  " + d + "\n")
		}
		b.WriteString("\n")
	}
	if strings.TrimSpace(in.OwnerNote) != "" {
		b.WriteString("## Owner note\n  " + in.OwnerNote + "\n\n")
	}
	if strings.TrimSpace(in.PriorPlanKiller) != "" {
		b.WriteString("## Prior plan invalidation (MANDATORY context)\n  " + in.PriorPlanKiller + "\n")
		if FlipToDirection(in.PriorPlanKiller) != "" {
			b.WriteString("  This is a flip-condition — the bias flip has ALREADY FIRED on machine-evaluated bars; your new plan MUST use that flipped direction, never the old bias.\n")
		}
		b.WriteString("\n")
	}
	if len(in.PriorPlanLevels) > 0 {
		b.WriteString("## Prior plan levels (continuity — carry over or re-anchor by price, do not silently drop the set)\n")
		for _, l := range in.PriorPlanLevels {
			b.WriteString("- " + l + "\n")
		}
		b.WriteString("\n")
	}

	// A1 (2026-08-26) — anchor-role truth: ONH/ONL are liquidity references,
	// not fade walls. Week evidence: ONH entries 14 · 21.4% win · −131 — the
	// fade framing lost; the research stat (broken intraday ~94% of days,
	// 2,827-day NQ sample) says they break nearly always. Advisory — no block.
	b.WriteString("## Anchor roles (week evidence, advisory)\n")
	b.WriteString("  ONH/ONL = LIQUIDITY/BREAKOUT references — broken intraday ~94% of days (2,827-day NQ sample). ")
	b.WriteString("Fade ONLY on a confirmed sweep-reclaim (wick through + close back inside on the decision TF). ")
	b.WriteString("Otherwise treat them as targets / breakout-retest anchors, never fade walls.\n\n")

	// A1 (planner-contract wave 2026-08-26) — BIAS DECISION TREE replaces
	// free-form bias guidance. Machine-computed branches; the planner names
	// the branch it took (contract requirement).
	if in.Price > 0 {
		b.WriteString(RenderBiasTree(in.Price, in.Levels, in.BiasCtxFacts))
	}

	// A2 — priority setup chain: the sweep → displacement → FVG-retrace play.
	b.WriteString("## Priority setup — THE CHAIN (A2, week evidence, advisory)\n")
	b.WriteString("  The A-setup: sweep of a pool → displacement → FVG retrace at a Tier-1/fresh-zone origin, inside a killzone. ")
	b.WriteString("fvg_entry scenarios SHOULD declare \"chain_after\": \"S#\" naming the sweep_reclaim that swept the origin pool. ")
	b.WriteString("entry_mode=ce is the DEFAULT; edge only for A-grade HTF-confluent origins. Stop beyond the sweep extreme; T1 = first opposing pool; the runner = the draw. ")
	b.WriteString("Citation: raw-FVG carries no standalone edge (40k-sample null) — the edge is conditional on the sweep→displacement chain.\n\n")

	// A3 — no-trade gates (≤8 lines, advisory — the plan declares them; the
	// executor still sees every cycle).
	b.WriteString("## No-trade gates (advisory — declare in no_trade or skip the day)\n")
	b.WriteString("  - balance-day (open inside prior value area AND VAs overlap) → edges-only, or skip\n")
	b.WriteString("  - opening gap >1.2×ATR or open outside the prior range → NEVER fade; the gap is a target\n")
	b.WriteString("  - no A/B zone in reach AND no pool swept by 10:30 ET → declare the skip in the plan\n")
	b.WriteString("  - lunch 11:30–13:30 ET: no new entries (the system hard-gates 12:00–13:30 CT)\n")
	b.WriteString("  - Tier-1 news → stand aside until a fresh post-news swing prints\n\n")

	// A4 — killzone weighting (advisory, not a gate).
	b.WriteString("## Killzone weighting (advisory)\n")
	b.WriteString("  NY AM 08:30–11:00 ET is the primary window; 10:00–11:00 ET is the premium FVG window; mind the macro minutes. ")
	b.WriteString("Conviction: down on Monday, up Thursday/Friday.\n\n")

	// A5 — stop-doing line: bare acceptance entries have no edge.
	b.WriteString("## STOP-DOING (week evidence, advisory)\n")
	b.WriteString("  acceptance entries WITHOUT a prior sweep + displacement are 0% win evidence — skip them (A5).\n\n")

	// 1h wave (2026-08-25) — the conditional 1h S/D mandate: emitted only when
	// a 1h supply/demand zone is actually rendered in the HTF zones section
	// (same conditional pattern as the G2.2 HTF mandate fix — a rule that asks
	// for something absent burns retries).
	has1HSD := false
	for _, z := range in.HTFZones {
		if is1HSDZone(z) {
			has1HSD = true
			break
		}
	}

	b.WriteString(plannerOutputContract(in.MaxLevels, in.ScenarioCap, len(in.HTFZones) > 0, has1HSD))
	return b.String()
}

// plannerOutputContract is the schema-strict, reasoning-first output spec. The
// level/scenario caps are the RESOLVED config values (H4/H5) — the prompt must
// ask for what validation will accept, so a raised max_levels/scenario_cap both
// gets requested AND passes instead of fail-closing every read against a
// hardcoded 8/3.
func plannerOutputContract(maxLevels, maxScenarios int, hasHTFZones, has1HSDZone bool) string {
	maxL, maxS := resolvePlanCaps(maxLevels, maxScenarios)
	htfRule := ""
	if hasHTFZones {
		htfRule = " — plus the HTF zones section, where you MUST include at least ONE HTF zone row in your levels as a confluence reference, never as a standalone trigger"
	}
	if has1HSDZone {
		htfRule += " — the nearest 1h supply/demand zone row in that section MUST be one of your included rows (setup-rung context, never a standalone trigger)"
	}
	return "## OUTPUT — one JSON object, reasoning FIRST, no prose outside it\n" +
		"{\n" +
		`  "reasoning": "<your read: what the auction is doing and why this plan — ≤200 words, decision-focused>",` + "\n" +
		`  "bias": {"direction": "long|short|neutral", "conviction": "high|medium|low", "flip_condition": "<explicit>"},` + "\n" +
		fmt.Sprintf(`  "levels": [{"price": <n>, "label": "<PDH|ONH|nPOC…>", "grade": "A|B|C", "instruction": "<verb>"}],  // max %d, MUST include ≥3 below AND ≥3 above the current price`, maxL) + "\n" +
		fmt.Sprintf(`  "scenarios": [{"id": "S1", "trigger": "<setup>", "condition": "reclaim|hold|sweep_reclaim|reject|acceptance|breakout_retest|fvg_entry|breakdown_continue|breakup_continue", "direction": "long|short", "target_chain": [<n>,…], "invalid": "<line>", "quality": "A+|A|B|C", "chain_after": "<S# of the sweep_reclaim this fvg_entry follows, or omit>", "confirm": {"rule": "touch|1x5m_close|2x5m_close|15m_close", "ref_price": <n>, "side": "above|below"}, "confirm2": {"rule": "<leg 2 rule>", "ref_price": <n>, "side": "above|below"} (OPTIONAL second trigger leg — a two-leg setup MUST carry both legs; the machine renders EVERY leg and a partial NEVER reads MET), "fvg": {"fvg_lo": <n>, "fvg_hi": <n>, "entry_mode": "edge|ce", "displacement_atr": <n>, "origin_level": "<label>", "direction": "long|short"}, "breakdown": {"level": <n>, "level_label": "<label>", "entry_mode": "pullback|immediate"} (REQUIRED iff condition==breakdown_continue|breakup_continue — see the WATERFALL PLAY rule), "arm": {"enabled": true, "entry": <n>, "stop": <n>, "target": <n>, "wait_confirm": true}}],  // 1..%d — confirm{} is REQUIRED per scenario; fvg{} REQUIRED iff condition=="fvg_entry" (ce is COMPUTED, never written); breakdown{} REQUIRED iff waterfall-class; chain_after is OPTIONAL; arm{} is OPTIONAL (see the ARMED ORDERS rule)`, maxS) + "\n" +
		`  "no_trade": ["first 5m (CT)", "12:00-13:30 CT lunch", "<calendar blackouts>"],` + "\n" +
		`  "death_condition": "<the single line that invalidates this whole plan>",` + "\n" +
		`  "death": {"price": <level>, "side": "below|above", "rule": "2x5m|15m_close|5m_close"},` + "\n" +
		`  "flip": {"price": <level>, "side": "below|above", "rule": "2x5m|15m_close|5m_close", "flip_to": "long|short"},` + "\n" +
		`  "day_type": "trend|balance|<optional>"` + "\n" +
		"}\n" +
		"Rules: levels chosen ONLY from the ranked table above" + htfRule + "; levels MUST be at least 3 points apart — near-duplicates are merged by the system; S/D & FVG are confluence, never standalone. " +
		"Copy the EXACT label from the table row for the price you choose — never re-label a table level as a different anchor (a zone price relabeled 'PDH/PDL/PDC' is a phantom level and is REJECTED at write). " +
		"Quality: A+ = highest-conviction setup, A = strong, B = workable, C = machine-demoted (trigger level consumed at write — G5) — use C honestly for a demoted setup, never as a default. " +
		"The scenario MIX must follow the regime + day_type: a trend-down day gets breakdown/pullback-short plays, a trend-up day the reverse, balance days get two-sided plays — do NOT default to 2 longs + 1 rally-rejection short on every day. " +
		"If price sits BELOW PDL you MUST write a continuation short; ABOVE PDH, a continuation long. " +
		"A1: your reasoning MUST open by naming the bias-tree branch you took (e.g. \"bias-tree: inside-day long LOW\"), then argue from it. " +
		"A2: an fvg_entry SHOULD chain after a sweep_reclaim (chain_after: S#) — bare gaps at non-A/B origins get a WARN at write, not a reject. " + "A2b (machine grounding, 2026-08-27): author an fvg_entry scenario ONLY from the ## FRESH FVGs list above — copy its direction and lo–hi EXACTLY. If the list is empty, do NOT author any fvg_entry (invented/stale gaps are REJECTED at write). " + "A2c (FVG demand, 2026-08-28): when ## FRESH FVGs is NON-empty and at least one candidate's direction agrees with your bias, you SHOULD author an fvg_entry from that candidate; if you decide not to, state the reason in ONE line in your reasoning (e.g. 'no fvg_entry: nearest fresh gap is 30pt away — outside my reach'). " + "death.flip objects are MACHINE-EVALUATED — choose levels from your level list and a rule; they must match the prose lines. " +
		"The flip and death MUST be DIFFERENT events: never the same level AND same rule for both (a flip at the same tick death fires is void). A short-biased plan's flip sits BELOW its death line or uses a stricter rule, so the flip can actually fire. " +
		"Every scenario's confirm{} is MACHINE-EVALUATED the same way: rule + ref_price + side, and ref_price MUST equal a number written in that scenario's trigger/invalid prose. " +
		"target_chain is GUIDANCE for the executor AI (which sets the actual take_profit) — it is validated for reachability at write time but never enforced at execution (D2 ruling). " +
		// WAVE 2 armed orders (2026-08-27) — the arming authorization. The LLM
		// chooses WHAT to arm; Go manages WHEN it fills (tick-level).
		// Autopsy-response wave: armable A/B setups SHOULD carry arm{} (the
		// resting order is the fast path); sweep_reclaim retraces chain via
		// wait_confirm; fantasy targets (planned R > 6) are WARN-flagged.
		"ARMED ORDERS (the resting order IS the fast path — prefer it over a 2-minute debate at the touch): every fvg_entry / reject scenario you author at quality A or B SHOULD carry arm{} — enabled:true + EXACT entry/stop/target (breakout_retest stays a normal AI play: the machine never arms it — GAR-F4). A setup the planner believes in gets a resting order, not a mid-touch argument. Long: stop < entry < target. Short: target < entry < stop. CHAINED ARMS: when a sweep_reclaim you believe in confirms, its RETRACE entry should already be resting — author that retrace as its own arm with wait_confirm:true, or add wait_confirm:true to the sweep scenario's arm: the system holds the arm dormant until the scenario's confirm{} is machine-MET, then places it (the sweep fast path). NEVER arm acceptance or a raw sweep WITHOUT the wait_confirm chain. Keep targets REAL: a planned R:R above ~6 is a fantasy target and gets WARN-flagged at write. The system places arms within a tick band, manages them tick-level, and cancels on veto/dormant/session-end. FEASIBILITY CONTRACT: an arm{} MUST be gate-feasible or it is REFUSED every cycle and learns nothing — R:R = |target\u2212entry| \u00f7 |stop\u2212entry| must be \u2265 2.0 (ARM_MIN_RR) AND the stop distance must be \u2265 1.0\u00d7 the current 5m ATR (the facts list the session ATR5m — cite the live value; a 10-point stop when ATR5m is ~16 is an instant refuse). If your setup cannot meet BOTH, OMIT arm{} and let the AI path take it. WATERFALL ARMS (F1): a breakdown_continue / breakup_continue at quality A or B SHOULD carry arm{} with wait_confirm:true + entry_mode=pullback — the resting limit sits AT the broken level and chains on confirm leg 1, so the pullback-that-fails FILLS it (immediate-mode waterfall plays stay on the AI path). " +
		// A2 (2026-08-26) — condition×session guidance from the week ledger:
		// reject 75% win +665 in NY RTH vs acceptance 0% −157 and sweep_reclaim
		// 0% −192. Advisory truth, not a hard rule.
		"Condition×session guidance (week evidence): reject-based setups are best in NY RTH (75% win, +665 this week); acceptance needs a clear displacement or skip (0% win this week); sweep_reclaim requires the reclaim CLOSE on the decision TF, never the wick alone (0% win this week). " +
		// WATERFALL PLAY (F1, 2026-08-28) — the momentum-follow class the 2026-08-28
		// -347pt crash exposed (missed-200pt report: every authored short was a
		// retest-FADE; no continuation entry existed, $0-by-own-rules). Author a
		// breakdown_continue (short) / breakup_continue (long) when: one-sided
		// delivery, a >1.2×ATR gap-and-go, or a waterfall after a failed rally.
		// breakdown{} carries the broken LEVEL (the retest) + entry_mode
		// (pullback = the shallow-retrace-that-fails entry, armable; immediate =
		// enter on the 2nd confirming close, AI path). The confirm{} is leg 1
		// (N closes beyond the level — machine-verified displacement ≥
		// BD_MIN_DISP_ATR×ATR5m, no reclaim close at write); confirm2{} is leg 2
		// (the retest that fails to reclaim). The machine renders both legs and a
		// partial NEVER reads MET. Targets chain to the next liquidity below
		// (above for longs); SL beyond the failed pullback extreme, ≥1×ATR5m.
		"WATERFALL PLAY (F1): author breakdown_continue|breakup_continue when the tape shows one-sided delivery, a >1.2×ATR gap-and-go, or a waterfall after a failed rally — the momentum-follow class. breakdown{} = broken level + entry_mode; confirm = leg 1 (breakdown), confirm2 = leg 2 (failed retest). " + // B3 (2026-08-26) — the ≤5-line noise-filter gate: the plan may include		// at most 5 near-duplicate LINE rows (within 3 points of each other);
		// keep the strongest anchor, drop the crowd.
		"NOISE FILTER (≤5): at most 5 of your included level rows may be line-levels clustered within 3 points of each other — keep the strongest of any such cluster, never a crowd. " +
		// FVG ENTRY MODEL (2026-08-26) — the 5th condition's ≤6-line playbook.
		"FVG ENTRY (fvg_entry): author when displacement off a Tier-1 level (VWAP/POC/PDH/PDL/ON/IB/fresh S/D) leaves a FRESH gap toward the trade direction · prefer the FIRST retrace (the freshness ladder applies to the gap as a zone) · entry_mode=ce for gaps > 20pts, edge for tighter · SL beyond the DISTAL edge · targets chain to the next liquidity (EQH/EQL, ON, PDH/PDL). Citations: NQ gap sweet spot 20–80pts; 1h+ fill ~70–80%; displacement floor per MSS research. " +
		"no_trade may contain ONLY the fixed session windows (first 5m, lunch) plus T1 HARD-blackout lines from the calendar — a T2 caution event is NEVER added to no_trade and never stops entries. " +
		"Respect the no-trade windows. If you cannot form a credible plan, say so in reasoning and output a neutral/no-trade plan.\n"
}

func absF(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
