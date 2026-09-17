package main

// Round 24 Q4 — TARGET REACH harness (read-only research).
//
// Inputs: the JSON exports written by ../export.sh (copy DB + live tail, every
// read through `sqlite3 -readonly`). Never opens a database.
//
// Per stored planner read (one plans row = one read, all trigger classes, sessions
// ASIA/LONDON/NY): replays the detection feed at the read instant on the bars
// that were CLOSED at that instant — LOOKAHEAD RULE: a bar participates at read
// R only if open_time + interval <= R (closedBars: CloseTime < R) — and scans
// the session window [R, flat) forward for target/level reach.
//
// Kernel functions called directly (none reimplemented):
//   kernel.DailyRangeProxy            — the dATR the proximity band is scaled by
//   kernel.ExtractMultiDayLevels      — PDH/PDL/PDC/PWH/PWL/PMH/PML … references
//   kernel.DetectHTFLevels            — per-TF swing/zone detection (G2), configured planner_timeframes
//   kernel.ComputeStructureMap        — S1 D/4h/1h trend (last-3-swing rule)
//   kernel.DefaultSessionRegistry, kernel.CTLocation
//
// Outputs (out-r24/): reads.jsonl, scenarios.jsonl, seats.jsonl, sessions.jsonl, run.json

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
)

type planRow struct {
	PlanID    string `json:"plan_id"`
	Version   int    `json:"version"`
	TradeDate string `json:"trade_date"`
	Session   string `json:"session"`
	Trigger   string `json:"trigger_reason"`
	Lifecycle string `json:"lifecycle"`
	Doc       string `json:"doc"`
	CreatedAt string `json:"created_at"`
	src       string
	at        time.Time
}

type docLevel struct {
	ID    string  `json:"id"`
	Kind  string  `json:"kind"`
	Price float64 `json:"price"`
	Lo    float64 `json:"lo"`
	Hi    float64 `json:"hi"`
	TF    string  `json:"tf"`
	Label string  `json:"label"`
}

type docScenario struct {
	ID          string          `json:"id"`
	LevelID     string          `json:"level_id"`
	Direction   string          `json:"direction"`
	TargetChain json.RawMessage `json:"target_chain"`
}

type planDoc struct {
	Levels    []docLevel    `json:"levels"`
	Scenarios []docScenario `json:"scenarios"`
}

func parseCreated(s string) (time.Time, error) {
	for _, l := range []string{"2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05-07:00", time.RFC3339Nano, "2006-01-02 15:04:05.999999999"} {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparsed created_at %q", s)
}

// targets parses target_chain leniently: numbers, numeric strings, or objects with a price field.
func targets(raw json.RawMessage) []float64 {
	var nums []float64
	if json.Unmarshal(raw, &nums) == nil {
		return nums
	}
	var elems []json.RawMessage
	if json.Unmarshal(raw, &elems) != nil {
		return nil
	}
	for _, e := range elems {
		var f float64
		if json.Unmarshal(e, &f) == nil {
			nums = append(nums, f)
			continue
		}
		var s string
		if json.Unmarshal(e, &s) == nil {
			if v, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
				nums = append(nums, v)
			}
			continue
		}
		var m map[string]any
		if json.Unmarshal(e, &m) == nil {
			for _, k := range []string{"price", "target", "level"} {
				if v, ok := m[k].(float64); ok {
					nums = append(nums, v)
					break
				}
			}
		}
	}
	return nums
}

// sessionWindow — the DefaultSessionRegistry clocks (CT), as round 23 used them.
func sessionWindow(day time.Time, session string) (winStart, winEnd time.Time, ok bool) {
	loc := kernel.CTLocation()
	y, m, d := day.Date()
	switch session {
	case "ASIA":
		return time.Date(y, m, d, 17, 0, 0, 0, loc), time.Date(y, m, d+1, 2, 0, 0, 0, loc), true
	case "LONDON":
		return time.Date(y, m, d, 2, 0, 0, 0, loc), time.Date(y, m, d, 8, 30, 0, 0, loc), true
	case "NY":
		return time.Date(y, m, d, 8, 30, 0, 0, loc), time.Date(y, m, d, 14, 45, 0, 0, loc), true
	}
	return winStart, winEnd, false
}

type dayPlanCfg struct {
	ProximityFilterATR float64  `json:"proximity_filter_atr"`
	MaxLevels          int      `json:"max_levels"`
	ReplanCap          int      `json:"replan_cap"`
	PlannerTimeframes  []string `json:"planner_timeframes"`
}

type readOut struct {
	Src          string  `json:"src"`
	Day          string  `json:"day"`
	Session      string  `json:"session"`
	PlanID       string  `json:"plan_id"`
	Version      int     `json:"version"`
	Trigger      string  `json:"trigger"`
	Lifecycle    string  `json:"lifecycle"`
	ReadAt       string  `json:"read_at"`
	ReadAtMs     int64   `json:"read_at_ms"`
	InWindow     bool    `json:"in_window"` // read inside [start, flat)
	Basis        string  `json:"basis"`          // the basis used: by-time (the tape's measured flip), authoritative
	BasisLevels  string  `json:"basis_levels"`   // diagnostic: basis nearest to the doc's median level price
	BasisSuspect bool    `json:"basis_suspect"`  // the two disagree → the doc's levels and the tape may not share a basis
	Price        float64 `json:"price"`
	DATR         float64 `json:"datr"`
	Band1        float64 `json:"band_k1"`
	Band15       float64 `json:"band_k15"`
	Bars1m       int     `json:"bars_1m"`
	NLevels      int     `json:"n_levels"`
	NScen        int     `json:"n_scenarios"`
	SessRange    float64 `json:"session_range"`
	SessComplete bool    `json:"session_complete"`
	SessBars     int     `json:"session_bars"`
	NetMove      float64 `json:"net_move"`     // session close − session open
	NetDir       string  `json:"net_dir"`      // up|down|flat
	FromReadMove float64 `json:"from_read_move"` // session close − price at read
	// (a)
	AheadNet     bool    `json:"a_ahead_net"`
	AheadNetDist float64 `json:"a_ahead_net_dist"`
	AheadNetLbl  string  `json:"a_ahead_net_label"`
	AheadFR      bool    `json:"a_ahead_fromread"`
	AheadFRDist  float64 `json:"a_ahead_fromread_dist"`
	AheadUp      int     `json:"a_levels_above"`
	AheadDown    int     `json:"a_levels_below"`
	// (a') the table's REACH: farthest seated level ahead in the net-move direction vs the session's excursion from the read price
	FarAheadDist   float64 `json:"a_far_ahead_dist"`   // -1 when no ahead level
	Excursion      float64 `json:"a_excursion"`        // max favorable excursion from price in net_dir over [R, flat)
	TableExhausted bool    `json:"a_table_exhausted"`  // excursion carried price beyond the farthest ahead level
	// (c) trend + counts
	DTrend    string `json:"d_trend"`
	H4Trend   string `json:"h4_trend"`
	H1Trend   string `json:"h1_trend"`
	DBars     int    `json:"d_bars"`
	H4Bars    int    `json:"h4_bars"`
	HTFStrict int    `json:"htf_strict_n"`
	HTFWide   int    `json:"htf_wide_n"`
}

type scenOut struct {
	Src       string  `json:"src"`
	Day       string  `json:"day"`
	Session   string  `json:"session"`
	PlanID    string  `json:"plan_id"`
	Version   int     `json:"version"`
	ReadAt    string  `json:"read_at"`
	ScenID    string  `json:"scenario"`
	Direction string  `json:"direction"`
	NTargets  int     `json:"n_targets"`
	Last      float64 `json:"last_target"`
	Monotonic bool    `json:"monotonic"`
	Price     float64 `json:"price"`
	Exhausted bool    `json:"exhausted"`
	MinToHit  int     `json:"min_to_exhaust"`
	Complete  bool    `json:"session_complete"`
	SessRange float64 `json:"session_range"`
}

type seatOut struct {
	Src      string  `json:"src"`
	Day      string  `json:"day"`
	Session  string  `json:"session"`
	PlanID   string  `json:"plan_id"`
	Version  int     `json:"version"`
	ReadAt   string  `json:"read_at"`
	Variant  string  `json:"variant"` // 4h | D | agree
	Dir      string  `json:"dir"`     // up | down
	BandK    float64 `json:"band_k"`
	Set      string  `json:"set"` // strict | wide
	Seat     int     `json:"seat"`
	Kind     string  `json:"kind"`
	TF       string  `json:"tf"`
	Label    string  `json:"label"`
	Lo       float64 `json:"lo"`
	Hi       float64 `json:"hi"`
	Dist     float64 `json:"dist"`
	Reached  bool    `json:"reached"`
	MinToHit int     `json:"min_to_hit"`
	Complete bool    `json:"session_complete"`
	SessRange float64 `json:"session_range"`
}

type dirSeat struct {
	variant, dir string
}

func main() {
	outDir := flag.String("out", "out-r24", "export/output directory")
	flag.Parse()

	st, err := loadBars(*outDir)
	if err != nil {
		fatal("bars: %v", err)
	}
	fmt.Printf("basis step Dec−Sep = %.2f pt (pooled median, n=%d, min %.2f max %.2f)\n", st.step, st.stepN, st.stepLo, st.stepHi)
	for _, tf := range []string{"1m", "15m", "1h", "4h", "1d"} {
		r := st.rows[tf]
		fmt.Printf("series %s rows=%d first=%d last=%d\n", tf, len(r), r[0].OpenMs, r[len(r)-1].OpenMs)
	}

	var cfgRows []struct {
		ID      string `json:"id"`
		DayPlan string `json:"day_plan"`
	}
	if err := loadJSON(*outDir+"/copy_strategy_dayplan.json", &cfgRows); err != nil || len(cfgRows) == 0 {
		fatal("strategy day_plan: %v", err)
	}
	var cfg dayPlanCfg
	if err := json.Unmarshal([]byte(cfgRows[0].DayPlan), &cfg); err != nil {
		fatal("day_plan parse: %v", err)
	}
	proxK := kernel.ResolveProximityK(cfg.ProximityFilterATR)
	tfs := cfg.PlannerTimeframes
	if len(tfs) == 0 {
		tfs = []string{"D", "4h", "1h", "15m"}
	}
	fmt.Printf("day_plan: proximity_filter_atr=%g (resolved k=%g) max_levels=%d replan_cap=%d planner_timeframes=%v\n", cfg.ProximityFilterATR, proxK, cfg.MaxLevels, cfg.ReplanCap, tfs)

	var plans []planRow
	for _, f := range []string{"copy_plans.json", "live_plans_tail.json"} {
		var rows []planRow
		if err := loadJSON(*outDir+"/"+f, &rows); err != nil {
			fatal("%s: %v", f, err)
		}
		for i := range rows {
			rows[i].src = strings.TrimSuffix(strings.TrimPrefix(f, "copy_"), ".json")
			if strings.HasPrefix(f, "live_") {
				rows[i].src = "live"
			} else {
				rows[i].src = "copy"
			}
			t, err := parseCreated(rows[i].CreatedAt)
			if err != nil {
				fatal("%v", err)
			}
			rows[i].at = t
		}
		plans = append(plans, rows...)
	}
	sort.SliceStable(plans, func(i, j int) bool { return plans[i].at.Before(plans[j].at) })
	fmt.Printf("plans loaded: %d\n", len(plans))

	reg := kernel.DefaultSessionRegistry()
	loc := kernel.CTLocation()
	wReads := mustCreate(*outDir + "/reads.jsonl")
	wScen := mustCreate(*outDir + "/scenarios.jsonl")
	wSeats := mustCreate(*outDir + "/seats.jsonl")
	defer wReads.Close()
	defer wScen.Close()
	defer wSeats.Close()

	type sessKey struct{ day, session string }
	sessions := map[sessKey]map[string]any{}
	strictKinds := map[string]bool{"PDH": true, "PDL": true, "PDC": true, "PWH": true, "PWL": true, "PMH": true, "PML": true}
	nReads, nSkipped := 0, 0
	basisMismatch := 0

	for _, p := range plans {
		if p.Session != "ASIA" && p.Session != "LONDON" && p.Session != "NY" {
			nSkipped++
			continue
		}
		day, err := time.ParseInLocation("2006-01-02", p.TradeDate, loc)
		if err != nil {
			nSkipped++
			continue
		}
		winStart, winEnd, ok := sessionWindow(day, p.Session)
		if !ok {
			nSkipped++
			continue
		}
		var doc planDoc
		_ = json.Unmarshal([]byte(p.Doc), &doc)
		R := p.at
		Rms := R.UnixMilli()

		// --- basis: by time and by the plan's own level prices ---
		basisByTime := basisSep
		if Rms >= tapeFlipMs {
			basisByTime = basisDec
		}
		basis := basisByTime
		basisLevels := ""
		if len(doc.Levels) > 0 {
			cbSep := st.lastClosed("1m", 1, Rms, basisSep)
			if len(cbSep) > 0 {
				prices := make([]float64, 0, len(doc.Levels))
				for _, l := range doc.Levels {
					if l.Price > 0 {
						prices = append(prices, l.Price)
					}
				}
				if len(prices) > 0 {
					sort.Float64s(prices)
					med := prices[len(prices)/2]
					cs := cbSep[len(cbSep)-1].Close
					if math.Abs(med-(cs+st.step)) < math.Abs(med-cs) {
						basisLevels = basisDec
					} else {
						basisLevels = basisSep
					}
				}
			}
		}
		// The level-median check misfires on a one-sided table (17 pre-flip reads have
		// their median level +128..+393 above price while the tape is provably Sep-basis),
		// so a disagreement is SUSPECT only inside the ring's mixed-basis window after the
		// flip: the live 2000-bar 1m ring still held Sep-basis bars for up to 2000 minutes,
		// so a plan written there can carry Sep-basis levels against a Dec-basis price.
		disagree := basisLevels != "" && basisLevels != basisByTime
		suspect := disagree && Rms >= tapeFlipMs && Rms < tapeFlipMs+int64(kernel.AISVPBarCount)*60_000
		if disagree {
			basisMismatch++
		}

		bars1m := st.lastClosed("1m", kernel.AISVPBarCount, Rms, basis)
		if len(bars1m) == 0 {
			nSkipped++
			continue
		}
		price := bars1m[len(bars1m)-1].Close
		dATR := kernel.DailyRangeProxy(bars1m, R)
		if dATR <= 0 {
			dATR = 0.008 * price // the kernel's own fallback (AssembleResearchLevels)
		}
		band1 := proxK * dATR
		band15 := 1.5 * dATR

		// --- session window facts ---
		win := st.between("1m", winStart.UnixMilli(), winEnd.UnixMilli(), basis)
		expected := int((winEnd.UnixMilli() - winStart.UnixMilli()) / 60_000)
		complete := len(win) >= int(0.8*float64(expected)) && len(win) > 0 && win[len(win)-1].OpenTime >= winEnd.UnixMilli()-5*60_000
		var sessRange, netMove, fromRead float64
		netDir := "n/a"
		if len(win) > 0 {
			hi, lo := math.Inf(-1), math.Inf(1)
			for _, b := range win {
				hi = math.Max(hi, b.High)
				lo = math.Min(lo, b.Low)
			}
			sessRange = hi - lo
			netMove = win[len(win)-1].Close - win[0].Open
			fromRead = win[len(win)-1].Close - price
			switch {
			case netMove > 0:
				netDir = "up"
			case netMove < 0:
				netDir = "down"
			default:
				netDir = "flat"
			}
		}
		sk := sessKey{p.TradeDate, p.Session}
		if _, ok := sessions[sk]; !ok {
			sessions[sk] = map[string]any{"day": p.TradeDate, "session": p.Session, "session_range": sessRange, "session_complete": complete,
				"session_bars": len(win), "net_move": netMove, "net_dir": netDir, "win_start": winStart.Format(time.RFC3339), "flat": winEnd.Format(time.RFC3339)}
		}
		scan := st.between("1m", Rms, winEnd.UnixMilli(), basis) // bars opened at/after the read, before the flat

		// --- (a) ahead-of-price levels in the seated table ---
		ro := readOut{Src: p.src, Day: p.TradeDate, Session: p.Session, PlanID: p.PlanID, Version: p.Version, Trigger: p.Trigger, Lifecycle: p.Lifecycle,
			ReadAt: R.In(loc).Format("2006-01-02 15:04:05 CT"), ReadAtMs: Rms, InWindow: Rms >= winStart.UnixMilli() && Rms < winEnd.UnixMilli(),
			Basis: basis, BasisLevels: basisLevels, BasisSuspect: suspect, Price: price, DATR: dATR, Band1: band1, Band15: band15, Bars1m: len(bars1m),
			NLevels: len(doc.Levels), NScen: len(doc.Scenarios), SessRange: sessRange, SessComplete: complete, SessBars: len(win),
			NetMove: netMove, NetDir: netDir, FromReadMove: fromRead}
		nearest := func(dir string) (bool, float64, string) {
			best := math.Inf(1)
			lbl := ""
			for _, l := range doc.Levels {
				lo, hi := l.Lo, l.Hi
				if lo <= 0 || hi <= 0 {
					lo, hi = l.Price, l.Price
				}
				var d float64
				switch dir {
				case "up":
					d = lo - price
				case "down":
					d = price - hi
				default:
					continue
				}
				if d > 0 && d < best {
					best = d
					lbl = fmt.Sprintf("%s %s@%.2f", l.Kind, l.TF, l.Price)
				}
			}
			if math.IsInf(best, 1) {
				return false, -1, "" // absent: -1, never a fabricated 0 (canon: no fabricated values)
			}
			return true, best, lbl
		}
		for _, l := range doc.Levels {
			lo, hi := l.Lo, l.Hi
			if lo <= 0 || hi <= 0 {
				lo, hi = l.Price, l.Price
			}
			if lo > price {
				ro.AheadUp++
			} else if hi < price {
				ro.AheadDown++
			}
		}
		ro.AheadNet, ro.AheadNetDist, ro.AheadNetLbl = nearest(netDir)
		ro.FarAheadDist = -1
		for _, l := range doc.Levels {
			lo, hi := l.Lo, l.Hi
			if lo <= 0 || hi <= 0 {
				lo, hi = l.Price, l.Price
			}
			var d float64
			switch netDir {
			case "up":
				d = hi - price // the far edge: the table is exhausted only once price clears the whole zone
			case "down":
				d = price - lo
			}
			if d > 0 && d > ro.FarAheadDist {
				ro.FarAheadDist = d
			}
		}
		for _, b := range scan {
			switch netDir {
			case "up":
				ro.Excursion = math.Max(ro.Excursion, b.High-price)
			case "down":
				ro.Excursion = math.Max(ro.Excursion, price-b.Low)
			}
		}
		ro.TableExhausted = ro.FarAheadDist > 0 && ro.Excursion > ro.FarAheadDist
		frDir := "flat"
		if fromRead > 0 {
			frDir = "up"
		} else if fromRead < 0 {
			frDir = "down"
		}
		ro.AheadFR, ro.AheadFRDist, _ = nearest(frDir)

		// --- (b) scenarios: target_chain exhausted within [R, flat) ---
		for _, s := range doc.Scenarios {
			tg := targets(s.TargetChain)
			so := scenOut{Src: p.src, Day: p.TradeDate, Session: p.Session, PlanID: p.PlanID, Version: p.Version, ReadAt: ro.ReadAt,
				ScenID: s.ID, Direction: strings.ToLower(s.Direction), NTargets: len(tg), Price: price, Complete: complete, SessRange: sessRange, MinToHit: -1}
			if len(tg) > 0 {
				so.Last = tg[len(tg)-1]
				so.Monotonic = true
				for i := 1; i < len(tg); i++ {
					if (so.Direction == "long" && tg[i] < tg[i-1]) || (so.Direction == "short" && tg[i] > tg[i-1]) {
						so.Monotonic = false
					}
				}
				for _, b := range scan {
					if (so.Direction == "long" && b.High >= so.Last) || (so.Direction == "short" && b.Low <= so.Last) {
						so.Exhausted = true
						so.MinToHit = int((b.OpenTime - Rms) / 60_000)
						break
					}
				}
			}
			writeJSON(wScen, so)
		}

		// --- (c) HTF universe at the read + S1 trend ---
		fetch := func(tf string, count int) []market.Kline {
			return st.lastClosed(tf, count, Rms, basis)
		}
		htf := kernel.DetectHTFLevels(fetch, tfs, "MNQ", R)
		multi := kernel.ExtractMultiDayLevels(bars1m, reg, R)
		var strict, wide []kernel.DetectedLevel
		for _, l := range htf {
			wide = append(wide, l)
			if l.TF == "1d" || l.TF == "4h" {
				strict = append(strict, l)
			}
		}
		for _, l := range multi {
			wide = append(wide, l)
			if strictKinds[string(l.Kind)] {
				strict = append(strict, l)
			}
		}
		ro.HTFStrict, ro.HTFWide = len(strict), len(wide)
		reader := func(tf string) []market.Kline {
			switch tf {
			case "D":
				return st.lastClosed("1d", 0, Rms, basis)
			case "4h":
				return st.lastClosed("4h", 0, Rms, basis)
			default:
				return st.lastClosed("1h", 0, Rms, basis)
			}
		}
		sm := kernel.ComputeStructureMap(reader, "", nil, price, Rms, kernel.DefaultStructureTrendSwings)
		ro.DTrend, ro.H4Trend, ro.H1Trend = "n/a", "n/a", "n/a"
		if sm != nil {
			if t, ok := sm.TFs["D"]; ok {
				ro.DTrend, ro.DBars = t.Trend, t.Bars
			}
			if t, ok := sm.TFs["4h"]; ok {
				ro.H4Trend, ro.H4Bars = t.Trend, t.Bars
			}
			if t, ok := sm.TFs["1h"]; ok {
				ro.H1Trend = t.Trend
			}
		}
		variants := []dirSeat{{"4h", ro.H4Trend}, {"D", ro.DTrend}}
		if ro.H4Trend == ro.DTrend {
			variants = append(variants, dirSeat{"agree", ro.DTrend})
		} else {
			variants = append(variants, dirSeat{"agree", "disagree"})
		}
		for _, v := range variants {
			if v.dir != "up" && v.dir != "down" {
				continue
			}
			for _, bk := range []struct {
				k    float64
				band float64
			}{{proxK, band1}, {1.5, band15}} {
				for _, set := range []struct {
					name string
					lv   []kernel.DetectedLevel
				}{{"strict", strict}, {"wide", wide}} {
					type cand struct {
						l    kernel.DetectedLevel
						dist float64
					}
					var cs []cand
					for _, l := range set.lv {
						lo, hi := l.Lo, l.Hi
						if lo <= 0 || hi <= 0 {
							lo, hi = l.Price, l.Price
						}
						switch v.dir {
						case "up":
							if lo > price+bk.band {
								cs = append(cs, cand{l, lo - price})
							}
						case "down":
							if hi < price-bk.band {
								cs = append(cs, cand{l, price - hi})
							}
						}
					}
					sort.SliceStable(cs, func(i, j int) bool { return cs[i].dist < cs[j].dist })
					for i := 0; i < len(cs) && i < 2; i++ {
						c := cs[i]
						lo, hi := c.l.Lo, c.l.Hi
						if lo <= 0 || hi <= 0 {
							lo, hi = c.l.Price, c.l.Price
						}
						so := seatOut{Src: p.src, Day: p.TradeDate, Session: p.Session, PlanID: p.PlanID, Version: p.Version, ReadAt: ro.ReadAt,
							Variant: v.variant, Dir: v.dir, BandK: bk.k, Set: set.name, Seat: i + 1, Kind: string(c.l.Kind), TF: c.l.TF, Label: c.l.Label,
							Lo: lo, Hi: hi, Dist: c.dist, MinToHit: -1, Complete: complete, SessRange: sessRange}
						for _, b := range scan {
							if (v.dir == "up" && b.High >= lo) || (v.dir == "down" && b.Low <= hi) {
								so.Reached = true
								so.MinToHit = int((b.OpenTime - Rms) / 60_000)
								break
							}
						}
						writeJSON(wSeats, so)
					}
					// an explicit "no candidate" row so n(reads with direction) is recoverable
					if len(cs) == 0 {
						writeJSON(wSeats, seatOut{Src: p.src, Day: p.TradeDate, Session: p.Session, PlanID: p.PlanID, Version: p.Version, ReadAt: ro.ReadAt,
							Variant: v.variant, Dir: v.dir, BandK: bk.k, Set: set.name, Seat: 0, MinToHit: -1, Complete: complete, SessRange: sessRange})
					}
				}
			}
		}
		writeJSON(wReads, ro)
		nReads++
	}
	wSess := mustCreate(*outDir + "/sessions.jsonl")
	keys := make([]sessKey, 0, len(sessions))
	for k := range sessions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].day != keys[j].day {
			return keys[i].day < keys[j].day
		}
		return keys[i].session < keys[j].session
	})
	for _, k := range keys {
		writeJSON(wSess, sessions[k])
	}
	wSess.Close()
	run := map[string]any{
		"reads": nReads, "skipped": nSkipped, "basis_levelcheck_disagree": basisMismatch,
		"basis_step": st.step, "basis_step_n": st.stepN, "basis_step_min": st.stepLo, "basis_step_max": st.stepHi,
		"tape_flip_ms": tapeFlipMs, "proximity_k": proxK, "planner_timeframes": tfs, "max_levels": cfg.MaxLevels, "replan_cap": cfg.ReplanCap,
		"lookahead": "levels at read R use bars with open_time+interval <= R (kernel closedBars: CloseTime < R); reach scans bars with open_time >= R and < session flat",
	}
	b, _ := json.MarshalIndent(run, "", " ")
	os.WriteFile(*outDir+"/run.json", b, 0o644)
	fmt.Printf("reads=%d skipped=%d basis_levelcheck_disagree=%d\n", nReads, nSkipped, basisMismatch)
}

func mustCreate(p string) *os.File {
	f, err := os.Create(p)
	if err != nil {
		fatal("create %s: %v", p, err)
	}
	return f
}

func writeJSON(f *os.File, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		fatal("marshal: %v", err) // a row that cannot be written must stop the run, never vanish
	}
	f.Write(b)
	f.Write([]byte("\n"))
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
