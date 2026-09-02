package trader

import (
	"fmt"
	"strings"
	"time"

	"nofx/market"
	"nofx/store"
)

// BarSourceBootLine (BAR-SOURCE WAVE 2026-09-02) reports, per TF, WHERE the
// resolver actually answers from and the earliest bar it can reach — every
// value READ from the resolver at boot, never a literal (A24). A TF the
// resolver cannot answer says so rather than being omitted.
//
// Shape:
//
//	📊 bars: 1w nt8_agg via 1d since 2020-11-11 (1500) · 1d nt8 since 2020-11-11 (1500) · …
//	   · ladder(1w)=[1d 1m] native 1w EXCLUDED (Fri→Thu stamps) · retention 1m=90d coarse=forever
func BarSourceBootLine(r *market.BarResolver, symbol string, now time.Time) string {
	if r == nil {
		return "📊 bars: resolver unavailable — no source report"
	}
	tfs := []string{"1w", "1d", "4h", "1h", "15m", "5m", "1m"}
	parts := make([]string, 0, len(tfs))
	for _, tf := range tfs {
		s, err := r.CompletedBars(symbol, tf, 0, now.UnixMilli())
		if err != nil || len(s.Bars) == 0 {
			parts = append(parts, fmt.Sprintf("%s UNAVAILABLE", tf))
			continue
		}
		via := string(s.Source)
		if s.FromTF != tf {
			via = fmt.Sprintf("%s via %s", s.Source, s.FromTF)
		}
		parts = append(parts, fmt.Sprintf("%s %s since %s (%d)",
			tf, via, time.UnixMilli(s.EarliestMs).UTC().Format("2006-01-02"), len(s.Bars)))
	}
	excl := ""
	if why := market.ExcludedNative("1w"); why != "" {
		head := why
		if i := strings.Index(head, "."); i > 0 {
			head = head[:i]
		}
		excl = fmt.Sprintf(" · native 1w EXCLUDED from the weekly ladder: %s", head)
	}
	return fmt.Sprintf("📊 bars: %s · ladder(1w)=%v%s · retention 1m=%dd 5m=%dd 15m=%dd coarse=forever",
		strings.Join(parts, " · "), market.LadderFor("1w"), excl,
		store.RetentionDaysFor("1m"), store.RetentionDaysFor("5m"), store.RetentionDaysFor("15m"))
}
