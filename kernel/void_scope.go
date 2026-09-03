package kernel

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"nofx/market"
)

// ── VOID SCOPE (2026-09-02) — ONE RESOLVED VIEW OF THE TAPE ───────────────────
//
// The prompt's VOID list and the write-site validator answer the SAME question
// ("has a close come back across this level?") and before this file they asked
// it of different tape:
//
//	render   auto_trader_planner.go:2304 — sinceMs = CMESessionDayStart · 12,000 bars
//	validate breakdown_continue.go:248   — sinceMs = 0 (whole slice)    ·  2,000 bars
//
// So a level broken and reclaimed BEFORE the 17:00 CT boundary was void to the
// validator and invisible to the prompt. That is precisely ONL 29141.25 on the
// 2026-09-02 20:58 CT read: eight seated levels listed VOID, ONL absent, and the
// read rejected on ONL. An overnight low is broken overnight, by definition.
//
// The class-45 parity fixture passed 40/40 across 20 tapes because it fed BOTH
// sides the same sinceMs. It pinned the two FUNCTIONS and never the CALL SITES.
//
// RULE: neither caller chooses a window or a bar slice. The resolver decides
// once, and the VALIDATOR'S scope wins — the prompt must list what the validator
// will reject, so the prompt reads the validator's tape, not its own.
type VoidScope struct {
	Bars     []market.Kline
	SinceMs  int64 // 0 = the whole slice; the validator's historical behaviour
	Interval string
	BarCount int
	Source   string // "provider" | "given" — how Bars were obtained
}

// VoidScopeBarCount is the resolved bar depth both sides read. Default
// AISVPBarCount (2,000) — the VALIDATOR's historical depth, deliberately not
// the render path's 12,000: matching the validator is the whole point, and a
// deeper slice would list levels the validator cannot see.
func VoidScopeBarCount() int {
	if v := os.Getenv("VOID_SCOPE_BARS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return AISVPBarCount
}

// VoidScopeInterval is the bar interval both sides read.
func VoidScopeInterval() string { return AISVPBarInterval }

// VoidScopeSinceMs is the window start both sides use. 0 = the whole slice.
// The validator has always judged the whole slice; the prompt now agrees.
func VoidScopeSinceMs() int64 { return 0 }

// ResolveVoidScope fetches the ONE tape both sides read. Nil provider or no
// bars → an empty scope, which renders no VOID list and voids nothing: the
// honest degradation, never a fabricated verdict.
func ResolveVoidScope(symbol string) VoidScope {
	sc := VoidScope{SinceMs: VoidScopeSinceMs(), Interval: VoidScopeInterval(), BarCount: VoidScopeBarCount(), Source: "provider"}
	if market.FuturesBarsProvider == nil {
		return sc
	}
	sc.Bars = market.FuturesBarsProvider(symbol, sc.Interval, sc.BarCount)
	return sc
}

// VoidScopeOf wraps an already-fetched slice in the resolved window. Used by
// fixtures and by any caller that already holds the tape — the WINDOW still
// comes from the resolver, never from the caller.
func VoidScopeOf(bars []market.Kline) VoidScope {
	return VoidScope{Bars: bars, SinceMs: VoidScopeSinceMs(), Interval: VoidScopeInterval(), BarCount: VoidScopeBarCount(), Source: "given"}
}

// VoidScopeBootLine reports the resolved scope, every field READ from its
// resolver — never a literal.
func VoidScopeBootLine() string {
	since := "whole-slice"
	if VoidScopeSinceMs() != 0 {
		since = fmt.Sprintf("since=%d", VoidScopeSinceMs())
	}
	return fmt.Sprintf("void scope: %s · %s×%d · one resolver for prompt AND validator (parity)",
		since, VoidScopeInterval(), VoidScopeBarCount())
}
