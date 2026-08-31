package kernel

// WEEKLY-BIAS WAVE (2026-08-30) — W2 Sunday weekly read: the prompt builder,
// the fail-closed validator (r1-r6), the weekly doc schema, and the pure W4/W5
// helpers (invalidation watch, shadow grading, counter clauses, draw-alignment
// tag). Everything here is pure Go — no DB, no LLM, no gates. The trader layer
// (trader/auto_trader_weekly.go) drives it from the EXISTING cycle.
//
// LAW (W5.4): nothing in this file feeds back into seating, grades, gates or
// sizes. Shadow results are LOG/Counter-only views — the real system is frozen.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"nofx/market"
)

// ── facts (what the model actually sees) ───────────────────────────────────

// WeeklyFacts is the rendered evidence package the weekly read is built from.
// FactsHash proves which facts the model saw (sha256 of the rendered facts
// sections — weekly candles + references + NWOG + IPDA + prior-week recap).
type WeeklyFacts struct {
	Now          time.Time
	Price        float64
	Weeks        []WeekCandle // ≤12 completed, oldest→latest
	Refs         WeeklyRefs
	RefsOK       bool
	NWOGs        []NWOG // last 5
	IPDA         []IPDARange
	ThinHistory  bool // CompletedWeekCount < 4
	SectionsText string
	FactsHash    string
}

// ComputeWeeklyFacts derives the full evidence package from stored 1m bars.
// Price ≤ 0 → the last bar close. Thin history is stamped, never faked.
func ComputeWeeklyFacts(bars1m []market.Kline, now time.Time, price float64) WeeklyFacts {
	f := WeeklyFacts{Now: now, Price: price}
	f.Weeks = CompletedWeekCandles(bars1m, now, 12)
	f.Refs, f.RefsOK = PriorWeekRefs(bars1m, now)
	f.NWOGs = LastNWOGs(bars1m, now, 5)
	f.IPDA = IPDA(DailySessionBars(bars1m), price)
	f.ThinHistory = CompletedWeekCount(bars1m, now) < 4
	if f.Price <= 0 && len(bars1m) > 0 {
		f.Price = bars1m[len(bars1m)-1].Close
	}
	f.SectionsText = RenderWeeklyFactsSections(f)
	f.FactsHash = WeeklyFactsHash(f)
	return f
}

// RenderWeeklyFactsSections renders the five facts sections, in spec order:
// Weekly candles · Weekly references · NWOG table · IPDA · Prior week recap.
func RenderWeeklyFactsSections(f WeeklyFacts) string {
	var b strings.Builder

	b.WriteString("## Weekly candles (12 completed weeks, oldest → latest)\n")
	if len(f.Weeks) == 0 {
		b.WriteString("(no completed weeks stored — thin history)\n")
	} else {
		b.WriteString("Time(CT)  Open  High  Low  Close  Volume  Struct\n")
		for _, w := range f.Weeks {
			b.WriteString(w.RenderRow() + "\n")
		}
	}
	b.WriteString("\n")

	b.WriteString("## Weekly references\n")
	if !f.RefsOK {
		b.WriteString("(no completed prior week — thin history)\n")
	} else {
		fmt.Fprintf(&b, "weekly_open %.2f · PWH %.2f · PWL %.2f · PWC %.2f\n",
			f.Refs.WeeklyOpen, f.Refs.PWH, f.Refs.PWL, f.Refs.PWC)
	}
	b.WriteString("\n")

	b.WriteString("## NWOG (last 5 weekend gaps, oldest → latest)\n")
	if len(f.NWOGs) == 0 {
		b.WriteString("(none computed — thin history)\n")
	} else {
		b.WriteString("born  hi  lo  CE  filled\n")
		for _, g := range f.NWOGs {
			filled := "no"
			if g.Filled {
				filled = "yes"
			}
			fmt.Fprintf(&b, "%s  %.2f  %.2f  %.2f  %s\n", g.Born, g.Hi, g.Lo, g.CE, filled)
		}
	}
	b.WriteString("\n")

	b.WriteString("## IPDA (trailing dealing ranges)\n")
	for _, r := range f.IPDA {
		if r.PosPct < 0 {
			fmt.Fprintf(&b, "%dd: insufficient history\n", r.Days)
			continue
		}
		fmt.Fprintf(&b, "%dd: high %.2f · low %.2f · price at %.0f%% of range\n", r.Days, r.High, r.Low, r.PosPct*100)
	}
	b.WriteString("\n")

	b.WriteString("## Prior week recap (facts only)\n")
	if len(f.Weeks) < 2 {
		b.WriteString("(insufficient history for a prior-week recap)\n")
	} else {
		prev := f.Weeks[len(f.Weeks)-1]
		rng := prev.High - prev.Low
		loc := 0.0
		if rng > 0 {
			loc = (prev.Close - prev.Low) / rng * 100
		}
		fmt.Fprintf(&b, "range %.1f pts · close at %.0f%% of range · touched/swept refs: none (touch data not computed)\n",
			rng, loc)
	}
	b.WriteString("\n")
	return b.String()
}

// WeeklyFactsHash is the sha256 hex of the rendered facts sections — the audit
// proof of exactly which facts the model saw (stored on the weekly plan row).
func WeeklyFactsHash(f WeeklyFacts) string {
	sum := sha256.Sum256([]byte(RenderWeeklyFactsSections(f)))
	return hex.EncodeToString(sum[:])
}

// WeeklyRefSet returns every computed reference the validator's r3 rule tests
// the draw against: IPDA extremes (computed ranges only), PWH, PWL, and the
// edges of UNFILLED weekend gaps — exactly the spec's r3 set (weekly_open is
// NOT draw material; it is a level-only reference).
func WeeklyRefSet(f WeeklyFacts) []float64 {
	seen := map[float64]bool{}
	add := func(p float64) {
		if p > 0 {
			seen[p] = true
		}
	}
	for _, r := range f.IPDA {
		if r.PosPct >= 0 {
			add(r.High)
			add(r.Low)
		}
	}
	add(f.Refs.PWH)
	add(f.Refs.PWL)
	for _, g := range f.NWOGs {
		if !g.Filled {
			add(g.Hi)
			add(g.Lo)
		}
	}
	out := make([]float64, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

// ── the weekly doc ──────────────────────────────────────────────────────────

// WeeklyDoc is the JSON the Sunday read stores on the plans row
// (session="WEEKLY"). Schema EXACT per spec W2.2.
type WeeklyDoc struct {
	Bias         string             `json:"bias"`       // bull | bear | neutral
	Conviction   string             `json:"conviction"` // low | med | high
	Draw         WeeklyDraw         `json:"draw"`       // draw_on_liquidity
	Invalidation WeeklyInvalidation `json:"invalidation"`
	WeeklyLevels []WeeklyLevel      `json:"weekly_levels"`
	Narrative    string             `json:"narrative"` // ≤3 lines, plain auction language
	// stamped at write (never model-written):
	FactsHash     string `json:"facts_hash,omitempty"`
	ThinHistory   bool   `json:"thin_history,omitempty"`
	InvalidatedAt string `json:"invalidated_at,omitempty"` // W4 mid-week invalidation stamp
}

// WeeklyDraw is the draw_on_liquidity object.
type WeeklyDraw struct {
	Name string  `json:"name"` // the reference name (PWH / PWL / IPDA-20d-high / NWOG edge …)
	Px   float64 `json:"px"`
}

// WeeklyInvalidation is the MANDATORY invalidation price + basis.
type WeeklyInvalidation struct {
	Px    float64 `json:"px"`
	Basis string  `json:"basis"` // e.g. "1h close beyond 30300.00"
}

// WeeklyLevel is one weekly-class level row.
type WeeklyLevel struct {
	Name string  `json:"name"`
	Px   float64 `json:"px"`
}

// ParseWeeklyDoc extracts the weekly doc JSON from a raw model reply (tolerant
// of markdown fences / prose around the object).
func ParseWeeklyDoc(raw string) (*WeeklyDoc, error) {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		if j := strings.Index(s, "```"); j > 0 {
			s = s[:j]
		}
	}
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
		if j := strings.LastIndex(s, "}"); j > i {
			s = s[:j+1]
		}
	}
	var d WeeklyDoc
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return nil, fmt.Errorf("weekly doc JSON malformed: %v", err)
	}
	return &d, nil
}

// ValidateWeeklyDoc is the FAIL-CLOSED validator. Returns "" when the doc is
// accepted, otherwise the reject reason (r1..r6 per spec W2).
func ValidateWeeklyDoc(d *WeeklyDoc, refs []float64, thinHistory bool) string {
	if d == nil {
		return "r1: empty doc"
	}
	// r1 — bias / conviction enums.
	switch strings.ToLower(strings.TrimSpace(d.Bias)) {
	case "bull", "bear", "neutral":
	default:
		return fmt.Sprintf("r1: bias %q not in {bull, bear, neutral}", d.Bias)
	}
	switch strings.ToLower(strings.TrimSpace(d.Conviction)) {
	case "low", "med", "high":
	default:
		return fmt.Sprintf("r1: conviction %q not in {low, med, high}", d.Conviction)
	}
	// r2 — invalidation: px > 0 AND basis non-empty.
	if d.Invalidation.Px <= 0 {
		return "r2: invalidation missing/malformed (px must be > 0)"
	}
	if strings.TrimSpace(d.Invalidation.Basis) == "" {
		return "r2: invalidation missing/malformed (basis must be non-empty)"
	}
	// r6 — thin history forces low conviction.
	if thinHistory && strings.ToLower(strings.TrimSpace(d.Conviction)) != "low" {
		return fmt.Sprintf("r6: thin_history (completed weeks < 4) but conviction %q != low", d.Conviction)
	}
	// r3 — the draw MUST be a computed reference within ±1 tick (0.25).
	const tick = 0.25
	matched := false
	for _, p := range refs {
		if math.Abs(d.Draw.Px-p) <= tick {
			matched = true
			break
		}
	}
	if !matched {
		return fmt.Sprintf("r3: draw %.2f matches NO computed reference within ±%.2f (refs: %d)", d.Draw.Px, tick, len(refs))
	}
	// r4 — narrative ≤ 3 lines.
	if n := strings.Count(strings.TrimSpace(d.Narrative), "\n") + 1; n > 3 {
		return fmt.Sprintf("r4: narrative has %d lines (max 3)", n)
	}
	// r5 — day-of-week reasoning tokens FORBIDDEN.
	if HasDayOfWeekTokens(d.Narrative) {
		return "r5: narrative contains day-of-week reasoning (FORBIDDEN)"
	}
	return ""
}

// BuildWeeklyPrompt assembles the full Sunday read prompt: facts sections +
// the Instructions block + the exact output JSON schema.
func BuildWeeklyPrompt(f WeeklyFacts) string {
	var b strings.Builder
	b.WriteString("# WEEKLY READ — CME MNQ futures\n")
	b.WriteString("You are a disciplined weekly-bias reasoner. Read the facts ONCE and write ONE weekly bias doc for the coming week. Facts below are Go-computed from stored 1m bars; your job is JUDGMENT, not re-deriving the data.\n\n")
	fmt.Fprintf(&b, "read time: %s (all times CT) · last stored price %.2f\n\n", ClockCTAndUTC(f.Now), f.Price)

	b.WriteString(f.SectionsText)

	b.WriteString("## Instructions (THE RULES — violations are rejected)\n")
	b.WriteString("1. Tier-A evidence ONLY for bias: (a) price vs weekly_open with acceptance, (b) PWH/PWL break-AND-HOLD vs sweep-and-reject, (c) the 3-week structure tags (HH/outside/LL/LH/inside). Nothing else may justify bias.\n")
	b.WriteString("2. NWOG and IPDA are DRAW/TARGET material only — citing them as bias evidence = reject.\n")
	b.WriteString("3. Day-of-week reasoning is FORBIDDEN (any weekday token in the narrative = reject).\n")
	b.WriteString("4. draw_on_liquidity MUST equal an IPDA extreme, PWH, PWL, or an unfilled-NWOG edge in the bias direction; neutral bias → the nearest untested pool on either side.\n")
	b.WriteString("5. MANDATORY invalidation: a price AND a basis time-frame written as \"1h close beyond <px>\" (a CLOSED 1h bar beyond the price invalidates the bias mid-week).\n")
	b.WriteString("6. narrative ≤ 3 lines, plain auction language, no marketing.\n\n")

	b.WriteString("## OUTPUT — one JSON object, no prose outside it\n")
	b.WriteString("```json\n")
	b.WriteString(`{"bias":"bull|bear|neutral","conviction":"low|med|high","draw":{"name":"<ref name>","px":<n>},"invalidation":{"px":<n>,"basis":"1h close beyond <px>"},"weekly_levels":[{"name":"<ref name>","px":<n>}],"narrative":"≤3 lines, plain auction language"}`)
	b.WriteString("\n```\n")
	b.WriteString("weekly_levels: the 3-6 weekly-class references the week should respect (drawn from the facts above).\n")
	return b.String()
}

// WeeklyContextLine renders the ≤3-line injection block for the session
// planner prompt (W3). Forms: active bias · none · neutral (invalidated) ·
// thin history.
func WeeklyContextLine(d *WeeklyDoc, thinWeeks int) string {
	if d == nil {
		return "WEEKLY: none"
	}
	if strings.TrimSpace(d.InvalidatedAt) != "" {
		return "WEEKLY: neutral (invalidated " + d.InvalidatedAt + ")"
	}
	if d.ThinHistory {
		return fmt.Sprintf("WEEKLY: thin history (%dw) — low conviction", thinWeeks)
	}
	return fmt.Sprintf("WEEKLY: %s/%s · draw %s %.2f · invalid %.2f (%s)",
		strings.ToLower(d.Bias), strings.ToLower(d.Conviction), d.Draw.Name, d.Draw.Px, d.Invalidation.Px, d.Invalidation.Basis)
}

// WeeklyExecutorLine renders the one executor-prompt context line (W3).
// F1 (2026-08-30): an INVALIDATED doc still exists and must stay visible —
// the executor reads "WEEKLY: neutral (invalidated …)", never silence (the
// silent "" hid the invalidated-bear doc during the −204pt sell-off).
func WeeklyExecutorLine(d *WeeklyDoc) string {
	if d == nil {
		return ""
	}
	if strings.TrimSpace(d.InvalidatedAt) != "" {
		return "WEEKLY: neutral (invalidated " + d.InvalidatedAt + ")"
	}
	return fmt.Sprintf("WEEKLY: %s/%s · draw %.2f", strings.ToLower(d.Bias), strings.ToLower(d.Conviction), d.Draw.Px)
}

// ApplyWeeklyDOA (F5, 2026-08-30) — the breach-at-write guard: if the weekly
// bias's own invalidation basis is ALREADY crossed at write time, stamp
// neutral + invalidated_at NOW instead of writing a stillborn doc the watch
// kills moments later (the 17:07:15 bear lived 3 seconds; the invalidated
// bear would have been RIGHT by 250pt). Returns true when it stamped neutral.
func ApplyWeeklyDOA(doc *WeeklyDoc, bars []market.Kline, now time.Time) bool {
	if doc == nil || strings.TrimSpace(doc.InvalidatedAt) != "" {
		return false
	}
	bias := strings.ToLower(strings.TrimSpace(doc.Bias))
	if (bias != "bull" && bias != "bear") || doc.Invalidation.Px <= 0 {
		return false
	}
	if !WeeklyInvalidationCrossed(bias, doc.Invalidation.Px, bars) {
		return false
	}
	doc.Bias = "neutral"
	doc.InvalidatedAt = FormatCT(now)
	return true
}

// ── W4 — mid-week invalidation watch (pure) ─────────────────────────────────

// WeeklyInvalidationBasisTF extracts the timeframe token from the basis string
// ("1h close beyond 30300.00" → "1h"). Falls back to the caller's default.
func WeeklyInvalidationBasisTF(basis string) string {
	f := strings.Fields(strings.TrimSpace(basis))
	if len(f) > 0 {
		tok := strings.ToLower(f[0])
		if strings.ContainsAny(tok, "mhd") {
			return tok
		}
	}
	return ""
}

// WeeklyInvalidationCrossed reports whether any CLOSED bar of the basis TF has
// crossed the invalidation price for the given bias (bull → a close below px;
// bear → a close above px). Neutral / px ≤ 0 → never crossed.
func WeeklyInvalidationCrossed(bias string, px float64, bars []market.Kline) bool {
	if px <= 0 {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(bias)) {
	case "bull":
		for _, k := range bars {
			if k.Close < px {
				return true
			}
		}
	case "bear":
		for _, k := range bars {
			if k.Close > px {
				return true
			}
		}
	}
	return false
}

// ── W5 — shadow mode (view-only, THE LAW: real system frozen) ──────────────

// WeeklyShadowLevel is the lightweight seating view the shadow re-rank uses.
type WeeklyShadowLevel struct {
	Price float64
	Label string
	Grade string
}

// WeeklyShadowRefs returns the weekly-class reference set for the shadow
// confluence band, computed cheaply from 1m bars: PWH, PWL, weekly_open,
// computed IPDA extremes and unfilled NWOG edges (the W5.1 set — weekly_open
// IS included here, unlike the r3 draw set).
func WeeklyShadowRefs(bars1m []market.Kline, now time.Time, price float64) []float64 {
	f := ComputeWeeklyFacts(bars1m, now, price)
	seen := map[float64]bool{}
	for _, p := range WeeklyRefSet(f) {
		seen[p] = true
	}
	if f.Refs.WeeklyOpen > 0 {
		seen[f.Refs.WeeklyOpen] = true
	}
	out := make([]float64, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

// WeeklyConfluent reports whether a level sits within `band` (in ATR5m units
// × the caller's ATR) of any weekly-class reference.
func WeeklyConfluent(lv WeeklyShadowLevel, refs []float64, band, atr5m float64) bool {
	if atr5m <= 0 || band <= 0 {
		return false
	}
	for _, p := range refs {
		if math.Abs(lv.Price-p) <= band*atr5m {
			return true
		}
	}
	return false
}

// WeeklyShadowReorder computes the shadow view of the seated list: levels
// within the weekly confluence band get grade×mult. Returns the confluent
// count and how many of the top-N positions WOULD change under a stable
// re-sort by shadow score (a reorder count > 0 = the shadow seating differs
// from the real seating). PURE VIEW: the input is never mutated, nothing is
// returned that could re-enter the real seating path.
func WeeklyShadowReorder(levels []WeeklyShadowLevel, refs []float64, band, atr5m, mult float64) (confluent, reorder int) {
	if len(levels) == 0 {
		return 0, 0
	}
	scored := make([]float64, len(levels))
	for i, lv := range levels {
		s := float64(GradeRank(lv.Grade))
		if WeeklyConfluent(lv, refs, band, atr5m) {
			confluent++
			s *= mult
		}
		scored[i] = s
	}
	// stable re-rank by shadow score desc; count index changes vs the real order.
	order := make([]int, len(levels))
	for i := range order {
		order[i] = i
	}
	for i := 1; i < len(order); i++ {
		for j := i; j > 0 && scored[order[j]] > scored[order[j-1]]; j-- {
			order[j], order[j-1] = order[j-1], order[j]
		}
	}
	for i, idx := range order {
		if idx != i {
			reorder++
		}
	}
	return confluent, reorder
}

// WeeklyCounterClauses renders the ⚖️ WEEKLY-COUNTER annotation clauses for an
// entry opposing the weekly bias — ONLY the clauses that would have changed
// the trade under the hypothetical Sep-9 hard rules. Aligned entries and
// neutral/low-conviction docs return nil (silent). grade: the scenario/level
// grade ("A"|"A+"|"B"|"C"…); rr: the planned risk:reward (0 = unknown → no
// RR clause).
func WeeklyCounterClauses(bias, conviction, side, grade string, rr float64) []string {
	bias = strings.ToLower(strings.TrimSpace(bias))
	conviction = strings.ToLower(strings.TrimSpace(conviction))
	if bias != "bull" && bias != "bear" {
		return nil // neutral weekly → nothing to oppose
	}
	if conviction != "med" && conviction != "high" {
		return nil // low conviction → the hypothetical rules would not fire
	}
	side = strings.ToLower(strings.TrimSpace(side))
	aligned := (bias == "bull" && side == "long") || (bias == "bear" && side == "short")
	if side != "long" && side != "short" {
		return nil // non-entry → silent
	}
	if aligned {
		return nil // aligned entry → silent
	}
	var clauses []string
	clauses = append(clauses, "would-halve-size")
	g := strings.ToUpper(strings.TrimSpace(grade))
	if g != "A" && g != "A+" {
		clauses = append(clauses, "would-require-A-grade")
	}
	if rr > 0 && rr < 4.0 {
		clauses = append(clauses, "would-need-RR≥4.0")
	}
	return clauses
}

// WeeklyDrawAlignTags returns the draw-alignment tag for a decision row:
// toward_draw | away | neutral. Longs with the draw above the entry are
// toward_draw; shorts with the draw below are toward_draw; anything else
// with an active directional weekly doc is away; neutral/no-doc → neutral.
func WeeklyDrawAlignTag(bias, side string, drawPx, entry float64) string {
	bias = strings.ToLower(strings.TrimSpace(bias))
	if (bias != "bull" && bias != "bear") || drawPx <= 0 || entry <= 0 {
		return "neutral"
	}
	side = strings.ToLower(strings.TrimSpace(side))
	switch side {
	case "long":
		if entry < drawPx {
			return "toward_draw"
		}
		return "away"
	case "short":
		if entry > drawPx {
			return "toward_draw"
		}
		return "away"
	}
	return "neutral"
}
