// picture_htf_replay — the LABELED historical replay of the two-picture mode.
//
// This harness re-runs the DETECTION rules (BodyPivots4H → H1CloseBreak →
// StructuralSwing5M → NearestOpposingZone → R:R gate) over a closed window of
// stored bars, the same kernel functions the live evaluator calls, and prints
// detected levels, H1 confirmations, geometry, and eligible/refused
// opportunities WITH REASONS.
//
// IT IS A REPLAY, NOT A LIVE PATH: it never touches the opportunity ledger,
// never fans out to the evaluator, never sends a wire frame. Historical
// receipts cannot trigger live entries — that is the separate mechanism pin
// (TestSept17ReplayNeverMintsOpportunities); THIS harness is the requested
// strategy replay: what the rules would have seen and done on that tape.
//
// Usage:
//
//	go run ./cmd/picture_htf_replay --db /tmp/picture-htf-replay.db \
//	  --start 2026-09-16T22:00:00Z --end 2026-09-17T22:00:00Z
//
// The DB argument must be a COPY of the live data.db (mode=ro). It reads
// 1m rows and derives 5m/1h/4h on read — the same derive-on-read convention
// as the chart path.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	_ "modernc.org/sqlite"

	"nofx/kernel"
	"nofx/market"
)

func main() {
	dbPath := flag.String("db", "", "path to a COPY of data.db (opened read-only)")
	startS := flag.String("start", "2026-09-16T22:00:00Z", "window start (RFC3339, UTC)")
	endS := flag.String("end", "2026-09-17T22:00:00Z", "window end (RFC3339, UTC)")
	symbol := flag.String("symbol", "MNQ", "symbol to replay")
	minRR := flag.Float64("min-rr", 2.5, "R:R gate (the resolved picture_htf minimum)")
	contextDays := flag.Int("context-days", 30, "days of 4H/1H history BEFORE --start used for level detection (mirrors the live cache)")
	flag.Parse()

	if *dbPath == "" {
		fmt.Fprintln(os.Stderr, "HISTORICAL REPLAY: --db is required (a COPY of data.db)")
		os.Exit(2)
	}
	start, err := time.Parse(time.RFC3339, *startS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad --start: %v\n", err)
		os.Exit(2)
	}
	end, err := time.Parse(time.RFC3339, *endS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad --end: %v\n", err)
		os.Exit(2)
	}

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  PICTURE-HTF HISTORICAL REPLAY — feed receipts NOT claimed")
	fmt.Println("  no ledger writes · no evaluator fan-out · no wire frames")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Printf("  db      : %s (read-only)\n", *dbPath)
	fmt.Printf("  symbol  : %s\n", *symbol)
	fmt.Printf("  window  : %s → %s\n", start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339))
	fmt.Printf("  min R:R : %.2f\n", *minRR)
	fmt.Println()

	m1, contracts, err := load1m(*dbPath, *symbol, start.Add(-time.Duration(*contextDays)*24*time.Hour).UnixMilli(), end.UnixMilli())
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("1m bars read: %d (%dd context + window; contracts present: %v)\n", len(m1), *contextDays, contracts)
	if len(m1) == 0 {
		fmt.Println("VERDICT: no 1m rows in the window — nothing to replay (state this honestly).")
		os.Exit(0)
	}

	// Derive higher timeframes from the 1m ladder (derive-on-read).
	fiveM := aggregate(m1, 5*60_000)
	h1 := aggregate(m1, 60*60_000)
	fourH := aggregate(m1, 4*60*60_000)
	fmt.Printf("derived: 5m=%d · 1h=%d · 4h=%d (context included; the OPPORTUNITY window is %s → %s)\n\n",
		len(fiveM), len(h1), len(fourH), start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339))

	// ── Section A: 4H levels in force when the window opened ──
	fmt.Println("── A. DETECTED 4H LEVELS (BodyPivots4H, as-of window start; retirement inside the window noted) ──")
	levelsAtStart := levelsAt(fourH, start.UnixMilli())
	levelsAtEnd := levelsAt(fourH, end.UnixMilli())
	activeEnd := map[int64]bool{}
	for _, l := range levelsAtEnd {
		activeEnd[l.SourceOpen] = true
	}
	for _, l := range levelsAtStart {
		ret := ""
		if !activeEnd[l.SourceOpen] {
			ret = " · RETIRED during the window"
		}
		fmt.Printf("  %-10s body %8.2f–%-8.2f wick %8.2f/%8.2f · source %s · knowable %s%s\n",
			l.Role, l.BodyTop, l.BodyBottom, l.WickHigh, l.WickLow,
			ts(l.SourceOpen), ts(l.KnowableAt), ret)
	}
	fmt.Printf("  in force at window start: %d · surviving at window end: %d\n\n", len(levelsAtStart), len(levelsAtEnd))

	// ── Section B: H1 confirmations ──
	fmt.Println("── B. H1 CONFIRMATIONS (each completed in-window H1 boundary vs levels active AT THAT BOUNDARY) ──")
	type breakEvent struct {
		cur     market.Kline
		prev    market.Kline
		verdict kernel.H1CloseBreakResult
	}
	var breaks []breakEvent
	h1Confirmed := 0
	for i := 1; i < len(h1); i++ {
		prev, cur := h1[i-1], h1[i]
		if cur.CloseTime >= end.UnixMilli() {
			break // the last H1 is still forming inside the window
		}
		if cur.OpenTime < start.UnixMilli() {
			continue // context only — confirmations counted inside the window
		}
		act := kernel.ActiveLevels(levelsAt(fourH, cur.OpenTime), cur.OpenTime)
		v := kernel.H1CloseBreak(act, prev, cur, 0.25)
		h1Confirmed++
		if v.Fired {
			breaks = append(breaks, breakEvent{prev: prev, cur: cur, verdict: v})
			fmt.Printf("  H1 %s: prev %.2f → new %.2f · BREAK %s over %s %.2f (active levels %d, crossed %d)\n",
				ts(cur.CloseTime), prev.Close, cur.Close, v.Direction, levelRoleOf(act, v.LevelIdx), v.Boundary, len(act), len(v.Crossed))
		} else {
			fmt.Printf("  H1 %s: prev %.2f → new %.2f · no break (active levels %d)\n",
				ts(cur.CloseTime), prev.Close, cur.Close, len(act))
		}
	}
	fmt.Printf("  H1 completions evaluated: %d · breaks fired: %d\n\n", h1Confirmed, len(breaks))

	// ── Section C: per-break 5m window geometry + verdict ──
	fmt.Println("── C. OPPORTUNITIES (per break: next 5m interval, swing stop, opposing zone, R:R) ──")
	eligible, refused := 0, 0
	reasonCounts := map[string]int{}
	for bi, b := range breaks {
		intervalStart := alignUp(b.cur.CloseTime+1, 5*60_000)
		// The first COMPLETED 5m bar of that interval.
		idx := -1
		for j, k := range fiveM {
			if k.OpenTime == intervalStart {
				idx = j
				break
			}
		}
		fmt.Printf("\n  #%d %s break over %.2f (H1 confirm %s)\n", bi+1, b.verdict.Direction, b.verdict.Boundary, ts(b.cur.CloseTime))
		if idx < 0 {
			fmt.Printf("    verdict: REFUSED — next 5m interval (%s) has no bar in the tape\n", ts(intervalStart))
			refused++
			reasonCounts["interval_unfilled"]++
			continue
		}
		newest5m := fiveM[idx]
		elapsed := newest5m.OpenTime - intervalStart
		lookback := fiveM[:idx+1]
		stopPx, ok := kernel.StructuralSwing5M(lookback, b.verdict.Direction, 24, b.cur.CloseTime)
		verb := "ELIGIBLE"
		refuseClass := ""
		targetPx := 0.0
		rr := 0.0
		if !ok {
			verb, refuseClass = "REFUSED", "no_swing"
			fmt.Printf("    interval %s (elapsed %dms) · swing: NONE\n", ts(intervalStart), elapsed)
		} else {
			stopAdj := stopPx - 0.25
			if b.verdict.Direction == "short" {
				stopAdj = stopPx + 0.5
			}
			entryRef := newest5m.Close
			targetPx, ok = kernel.NearestOpposingZone(kernel.ActiveLevels(levelsAt(fourH, newest5m.OpenTime), newest5m.OpenTime), entryRef, b.verdict.Direction, newest5m.OpenTime)
			if !ok {
				verb, refuseClass = "REFUSED", "no_opposing_zone"
			} else {
				risk := entryRef - stopAdj
				if risk < 0 {
					risk = -risk
				}
				reward := targetPx - entryRef
				if reward < 0 {
					reward = -reward
				}
				if risk > 0 {
					rr = reward / risk
				}
				if rr < *minRR {
					verb, refuseClass = "REFUSED", "rr_below_min"
				}
				fmt.Printf("    interval %s (elapsed %dms) · entry ref %.2f · swing stop %.2f (adj %.2f) · opposing zone %.2f · R:R %.2f\n",
					ts(intervalStart), elapsed, newest5m.Close, stopPx, stopAdj, targetPx, rr)
			}
		}
		if verb == "ELIGIBLE" {
			fmt.Println("    verdict: ELIGIBLE — would submit in live mode (subject to flat/fresh/unreconciled re-checks)")
			eligible++
		} else {
			fmt.Printf("    verdict: REFUSED — %s\n", refuseClass)
			refused++
			reasonCounts[refuseClass]++
		}
	}

	fmt.Println("\n════════════════════════════════════════════════════════════════")
	fmt.Printf("  SUMMARY: levels(as-of start)=%d · H1 completions=%d · breaks=%d · eligible=%d · refused=%d\n",
		len(levelsAtStart), h1Confirmed, len(breaks), eligible, refused)
	if len(reasonCounts) > 0 {
		fmt.Println("  refusal reasons:")
		for r, n := range reasonCounts {
			fmt.Printf("    - %s ×%d\n", r, n)
		}
	}
	fmt.Println("  LABEL: HISTORICAL REPLAY — feed receipts NOT claimed; no opportunity")
	fmt.Println("  rows were written; live entries require native live frames + the")
	fmt.Println("  AddOn capability (MinAddonBuildPictureHtf) and the flat/fresh/")
	fmt.Println("  unreconciled re-checks at send time.")
	fmt.Println("════════════════════════════════════════════════════════════════")
}

func levelRoleOf(levels []kernel.PictureHtfLevel, idx int) string {
	if idx < 0 || idx >= len(levels) {
		return "?"
	}
	return levels[idx].Role
}

// levelsAt is the TIME-FAITHFUL snapshot: pivots over the 4H bars COMPLETED
// before `ms` (retirement scans only those bars) — the same semantics as the
// evaluator's rebuild (bars filtered by CloseTime < now).
func levelsAt(fourH []market.Kline, ms int64) []kernel.PictureHtfLevel {
	completed := make([]market.Kline, 0, len(fourH))
	for _, b := range fourH {
		if b.CloseTime < ms {
			completed = append(completed, b)
		}
	}
	return kernel.BodyPivots4H(completed, 120)
}

func ts(ms int64) string {
	return time.UnixMilli(ms).UTC().Format("2006-01-02 15:04:05Z")
}

func alignUp(ms, span int64) int64 {
	return ((ms + span - 1) / span) * span
}

// load1m reads 1m rows for the window, deduped by open_time (max rowid wins —
// the same convention the bar-history reader uses across contracts).
func load1m(dbPath, symbol string, startMs, endMs int64) ([]market.Kline, []string, error) {
	dsn := "file:" + dbPath + "?mode=ro&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()
	rows, err := db.Query(`
		SELECT open_time_ms, o, h, l, c, contract
		FROM bars
		WHERE symbol = ? AND tf = '1m' AND open_time_ms >= ? AND open_time_ms < ?
		GROUP BY open_time_ms
		HAVING rowid = MAX(rowid)
		ORDER BY open_time_ms ASC`, symbol, startMs, endMs)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var out []market.Kline
	contractSet := map[string]bool{}
	for rows.Next() {
		var openMs int64
		var o, h, l, c float64
		var contract string
		if err := rows.Scan(&openMs, &o, &h, &l, &c, &contract); err != nil {
			return nil, nil, err
		}
		contractSet[contract] = true
		out = append(out, market.Kline{
			OpenTime:  openMs,
			CloseTime: openMs + 60_000 - 1,
			Open:      o, High: h, Low: l, Close: c,
			Final: true, // stored rows are closed by definition (the replay label above is the honesty gate)
		})
	}
	var contracts []string
	for c := range contractSet {
		contracts = append(contracts, c)
	}
	sort.Strings(contracts)
	return out, contracts, rows.Err()
}

// aggregate derives a higher timeframe from the 1m ladder: bucket = open −
// open%span; OHLC over the bucket's minutes. A bucket is INCLUDED only when
// every minute of it is present (a partial edge bucket would distort the
// pivot/break math — partial bars are dropped and the count is printed).
func aggregate(m1 []market.Kline, span int64) []market.Kline {
	if len(m1) == 0 {
		return nil
	}
	type acc struct {
		o, h, l, c float64
		open       int64
		n          int
	}
	m := map[int64]*acc{}
	var order []int64
	for _, b := range m1 {
		bucket := b.OpenTime - b.OpenTime%span
		a, ok := m[bucket]
		if !ok {
			m[bucket] = &acc{o: b.Open, h: b.High, l: b.Low, c: b.Close, open: bucket, n: 1}
			order = append(order, bucket)
			continue
		}
		if b.High > a.h {
			a.h = b.High
		}
		if b.Low < a.l {
			a.l = b.Low
		}
		a.c = b.Close
		a.n++
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	full := span / 60_000
	out := make([]market.Kline, 0, len(order))
	for _, b := range order {
		a := m[b]
		if a.n < int(full) {
			continue // partial bucket — not a real closed bar
		}
		out = append(out, market.Kline{
			OpenTime: a.open, CloseTime: a.open + span - 1,
			Open: a.o, High: a.h, Low: a.l, Close: a.c, Final: true,
		})
	}
	return out
}
