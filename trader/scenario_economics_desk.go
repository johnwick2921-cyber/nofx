package trader

import (
	"fmt"
	"nofx/kernel"
	"strings"
	"time"
)

func (at *AutoTrader) deskScenarioEconomics(now time.Time) DeskLine {
	const source = "versioned plan document; authored economics, not broker prices"
	p := kernel.ActivePlanFor(at.id, at.futuresSymbol())
	if p == nil || p.BirthMs <= 0 {
		return deskUnknown(14, "scenarios", "SCENARIOS", source, "no dated active plan")
	}
	lines := make([]string, 0, len(p.Doc.Scenarios))
	unknown := false
	// W2 — every scenario carries its fade label on the strip (D4). The
	// live reading at `now`; the durable per-episode stamp is fixed at open.
	fade := at.FadeLabelsFor(now, &p.Doc, at.LastPriceForDesk(), nil)
	for _, s := range p.Doc.Scenarios {
		lines = append(lines, kernel.EconomicsSummary(s)+" · "+fadeChipText(fade[s.ID]))
		v := kernel.EconomicsFor(s)
		unknown = unknown || v.ObstacleR == nil || v.ArmR == nil
	}
	lines = append(lines, at.FadeCounterToday(now).Text())
	if len(lines) == 0 {
		return deskUnknown(14, "scenarios", "SCENARIOS", source, "no authored scenarios")
	}
	line := DeskLine{N: 14, Key: "scenarios", Label: "SCENARIOS", Source: source, AsOfMs: p.BirthMs, State: "ok", Verified: !unknown, Text: fmt.Sprintf("%s v%d · %s", p.Session, p.Version, strings.Join(lines, " | "))}
	if unknown {
		line.State = "unknown"
		line.Reason = "legacy or missing economics remains UNKNOWN"
	}
	if now.UnixMilli() < p.BirthMs {
		line.State = "unknown"
		line.Verified = false
		line.Reason = "plan timestamp is in the future"
	}
	return line
}
