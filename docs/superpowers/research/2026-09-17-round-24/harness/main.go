package main

// main.go — Round 24 Q4 "TARGET REACH" harness (read-only research).
//
// Inputs: the sqlite3 -readonly dumps in out-r24/in (see ../extract.sh).
// Outputs (out-r24/): sessions.jsonl · reads.jsonl · scenarios.jsonl ·
// dayplan.jsonl — one row per session-day / planner read / scenario /
// session-day budget. analysis.py turns these into the report's cells.
//
// LOOKAHEAD RULE: every "outcome" field (excursion, hit, minutes-to-hit,
// exhausted) is computed ONLY from 1m bars with open >= the read's created_at
// and < the session's flat. Every "at-read" field (price, DATR, seated table,
// HTF universe, trend) uses ONLY bars closed before created_at, or the
// planner's own recorded snapshot of that read.

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
)

// live strategy knobs (out-r24/in/strategy_dayplan.json, strategy "MNQ"):
// proximity_filter_atr=1.0, max_levels=12, replan_cap=4, planner_timeframes
// D/4h/1h/15m/5m. Read at runtime from the dump; these are the fallbacks.
type dayPlanCfg struct {
	ProximityK  float64
	MaxLevels   int
	ReplanCap   int
	Timeframes  []string
	StrategyID  string
	StrategyUpd string
}

func loadDayPlanCfg(path string) dayPlanCfg {
	cfg := dayPlanCfg{ProximityK: 1.0, MaxLevels: 12, ReplanCap: 4, Timeframes: []string{"D", "4h", "1h", "15m", "5m"}}
	var rows []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsActive  any    `json:"is_active"`
		UpdatedAt string `json:"updated_at"`
		DayPlan   string `json:"day_plan"`
	}
	if err := loadJSON(path, &rows); err != nil {
		return cfg
	}
	for _, r := range rows {
		if r.Name != "MNQ" {
			continue
		}
		var dp struct {
			ProximityFilterATR float64  `json:"proximity_filter_atr"`
			MaxLevels          int      `json:"max_levels"`
			ReplanCap          int      `json:"replan_cap"`
			PlannerTimeframes  []string `json:"planner_timeframes"`
		}
		if json.Unmarshal([]byte(r.DayPlan), &dp) == nil {
			cfg.ProximityK = kernel.ResolveProximityK(dp.ProximityFilterATR)
			if dp.MaxLevels > 0 {
				cfg.MaxLevels = dp.MaxLevels
			}
			if dp.ReplanCap > 0 {
				cfg.ReplanCap = dp.ReplanCap
			}
			if len(dp.PlannerTimeframes) > 0 {
				cfg.Timeframes = dp.PlannerTimeframes
			}
			cfg.StrategyID, cfg.StrategyUpd = r.ID, r.UpdatedAt
		}
	}
	return cfg
}

// ── output rows ─────────────────────────────────────────────────────────────

type sessionRow struct {
	Key       string  `json:"key"` // trade_date:session
	TradeDate string  `json:"trade_date"`
	Session   string  `json:"session"`
	Contract  string  `json:"contract"`
	WinStart  int64   `json:"win_start_ms"`
	WinEnd    int64   `json:"win_end_ms"`
	Bars      int     `json:"bars_1m"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Range     float64 `json:"range"`
	Reads     int     `json:"reads"`
	Measured  bool    `json:"measured"` // false → no 1m tape for the window
}

type tableStat struct {
	N           int     `json:"n"`
	AheadN      int     `json:"ahead_n"`         // levels on the eventual-move side
	AheadInBand int     `json:"ahead_in_band_n"` // …and within ±k×DATR
	Farthest    float64 `json:"farthest_ahead"`  // pts from price, 0 when none
	RanOut      bool    `json:"ran_out"`         // excursion > farthest ahead (or none ahead)
	Overshoot   float64 `json:"overshoot"`       // excursion − farthest (when ran out)
	AheadHTF    int     `json:"ahead_htf_n"`     // ahead levels with tf ∈ {1h,4h,D}
}

type candidate struct {
	TrendDef string  `json:"trend_def"` // s1_1h | s1_4h | s1_D | plan_bias
	Trend    string  `json:"trend"`     // up | down | range | n/a
	Found    bool    `json:"found"`
	Kind     string  `json:"kind,omitempty"`
	TF       string  `json:"tf,omitempty"`
	Role     string  `json:"role,omitempty"`
	Price    float64 `json:"price,omitempty"`
	Dist     float64 `json:"dist,omitempty"` // |level − price|, > band by construction
	Edge     float64 `json:"edge,omitempty"` // near edge used for the hit test
	Hit      bool    `json:"hit"`
	Minutes  float64 `json:"minutes,omitempty"`
	Measured bool    `json:"measured"` // false → trend ring too thin / no outcome bars
	RingN    int     `json:"ring_n,omitempty"`
}

type readRow struct {
	ID        string  `json:"id"` // plan_id@version
	PlanID    string  `json:"plan_id"`
	Version   int     `json:"version"`
	TradeDate string  `json:"trade_date"`
	Session   string  `json:"session"`
	SessKey   string  `json:"sess_key"`
	Trigger   string  `json:"trigger"`
	IsRead    bool    `json:"is_read"` // a model read (scheduled/level_event/death_replan/structure_mss/owner_reread) — markers copy the prior doc
	Lifecycle string  `json:"lifecycle"`
	ReadMs    int64   `json:"read_ms"`
	ReadCT    string  `json:"read_ct"`
	Contract  string  `json:"contract"`
	Recorded  bool    `json:"recorded"` // planner input snapshot found in the research store
	SnapID    string  `json:"snap_id,omitempty"`
	SnapLagS  float64 `json:"snap_lag_s,omitempty"`
	Price     float64 `json:"price"`
	PriceSrc  string  `json:"price_src"` // snapshot | bar_close
	DATR      float64 `json:"datr"`
	DATRSrc   string  `json:"datr_src"` // snapshot | DailyRangeProxy
	BandK     float64 `json:"band_k"`
	Band      float64 `json:"band"`
	Bias      string  `json:"bias"`

	// outcome (post-read, pre-flat)
	OutBars   int     `json:"out_bars"`
	ExcUp     float64 `json:"exc_up"`
	ExcDown   float64 `json:"exc_down"`
	DirExc    string  `json:"dir_exc"`   // up | down (dominant excursion)
	DirClose  string  `json:"dir_close"` // up | down | flat (flat close vs price)
	Excursion float64 `json:"excursion"` // in DirExc
	Measured  bool    `json:"measured"`

	Seated           *tableStat `json:"seated,omitempty"`              // recorded 12-seat input table
	Doc              *tableStat `json:"doc"`                           // the plan's published levels[]
	SeatedMaxAbsDist float64    `json:"seated_max_abs_dist,omitempty"` // band consistency probe

	HTFSrc             string            `json:"htf_src"` // snapshot | DetectHTFLevels
	HTFN               int               `json:"htf_n"`
	HTFParityRecordedN int               `json:"htf_parity_recorded_n,omitempty"` // len(snapshot.HTFZonesFull): in-band zones only
	HTFParityReconN    int               `json:"htf_parity_recon_inband_zones_n,omitempty"`
	HTFBeyondBand      int               `json:"htf_beyond_band_n"`
	Trend              map[string]string `json:"trend"` // s1_1h/s1_4h/s1_D → up/down/range/n/a
	TrendRing          map[string]int    `json:"trend_ring"`
	Cands              []candidate       `json:"cands"`
	// CandsInBand: the same rule with the FARTHEST in-band HTF level in the
	// trend direction (the alternative seat rule the beyond-band numbers point at).
	CandsInBand []candidate `json:"cands_inband"`
	// Oracle (uses the eventual move direction — lookahead by design, labelled):
	// in-band HTF (1h/4h/D) levels on the move side that lie BEYOND the doc
	// table's farthest-ahead level, and how many of them price reached.
	OracleMissedN    int      `json:"oracle_missed_inband_htf_n"`
	OracleMissedHitN int      `json:"oracle_missed_inband_htf_hit_n"`
	OracleMissedIDs  []string `json:"oracle_missed_ids,omitempty"`
	// same oracle against the recorded 12-seat SEATED table (recorded reads only)
	OracleSeatedMissedN    int `json:"oracle_seated_missed_n,omitempty"`
	OracleSeatedMissedHitN int `json:"oracle_seated_missed_hit_n,omitempty"`
}

type scenarioRow struct {
	ReadID    string  `json:"read_id"`
	TradeDate string  `json:"trade_date"`
	Session   string  `json:"session"`
	SessKey   string  `json:"sess_key"`
	ReadMs    int64   `json:"read_ms"`
	ScenID    string  `json:"scen_id"`
	Direction string  `json:"direction"`
	Price     float64 `json:"price_at_read"`
	NTargets  int     `json:"n_targets"`
	First     float64 `json:"first_target"`
	Last      float64 `json:"last_target"`
	LastDist  float64 `json:"last_dist"` // signed toward direction
	HitFirst  bool    `json:"hit_first"`
	Exhausted bool    `json:"exhausted"` // price beyond the LAST target before flat
	MinLast   float64 `json:"minutes_to_last,omitempty"`
	Measured  bool    `json:"measured"`
	Malformed string  `json:"malformed,omitempty"` // target on the wrong side, etc.
}

type dayplanRow struct {
	Key                              string         `json:"key"`
	TradeDate                        string         `json:"trade_date"`
	Session                          string         `json:"session"`
	Reads                            int            `json:"reads"`
	Triggers                         map[string]int `json:"triggers"`
	LevelEvent                       int            `json:"level_event"`
	Spending                         int            `json:"spending_reads"` // death_replan + owner_reread (class 35 spends)
	CounterUsed                      int            `json:"counter_used"`   // recorded system_config counter (max over baselines)
	CounterKeys                      []string       `json:"counter_keys"`
	Cap                              int            `json:"cap"`
	Exhausted                        bool           `json:"exhausted"` // counter_used >= cap
	WouldExhaustIfLevelEventsCounted bool           `json:"would_exhaust_if_level_events_counted"`
}

// ── helpers ─────────────────────────────────────────────────────────────────

func sign(x float64) string {
	switch {
	case x > 0:
		return "up"
	case x < 0:
		return "down"
	}
	return "flat"
}

func trendWord(s string) string {
	switch s {
	case "TRENDING_UP":
		return "up"
	case "TRENDING_DOWN":
		return "down"
	case "RANGING":
		return "range"
	}
	return "n/a"
}

func toKB(ks []market.Kline) []market.KlineBar {
	out := make([]market.KlineBar, 0, len(ks))
	for _, k := range ks {
		out = append(out, market.KlineBar{Time: k.OpenTime, Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume})
	}
	return out
}

// tableStats scores one level table against the read's eventual move.
func tableStats(prices []float64, tfs []string, p, band, exc float64, dir string) *tableStat {
	st := &tableStat{N: len(prices)}
	for i, lp := range prices {
		d := lp - p
		if sign(d) != dir {
			continue
		}
		st.AheadN++
		if math.Abs(d) <= band {
			st.AheadInBand++
		}
		if math.Abs(d) > st.Farthest {
			st.Farthest = math.Abs(d)
		}
		if i < len(tfs) {
			switch tfs[i] {
			case "1h", "4h", "D", "1d":
				st.AheadHTF++
			}
		}
	}
	st.RanOut = st.AheadN == 0 || exc > st.Farthest
	if st.RanOut {
		st.Overshoot = exc - st.Farthest
	}
	return st
}

// firstTouch scans outcome bars for the first bar whose range reaches `edge`
// on side `dir` (up: high >= edge; down: low <= edge). Returns minutes since
// readMs, or -1.
func firstTouch(out []market.Kline, readMs int64, edge float64, dir string) float64 {
	for _, b := range out {
		if (dir == "up" && b.High >= edge) || (dir == "down" && b.Low <= edge) {
			return float64(b.OpenTime-readMs) / 60_000
		}
	}
	return -1
}

func main() {
	inDir := flag.String("in", "", "out-r24/in directory (extract.sh output)")
	outDir := flag.String("out", "", "out-r24 directory")
	flag.Parse()
	if *inDir == "" || *outDir == "" {
		fmt.Fprintln(os.Stderr, "usage: r24 -in out-r24/in -out out-r24")
		os.Exit(2)
	}
	must := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, "FATAL:", err)
			os.Exit(1)
		}
	}
	cfg := loadDayPlanCfg(filepath.Join(*inDir, "strategy_dayplan.json"))
	bars, err := newBarDB(filepath.Join(*inDir, "bars.csv"), filepath.Join(*inDir, "bars_htf.csv"))
	must(err)
	plans, err := loadPlans(filepath.Join(*inDir, "plans.json"))
	must(err)
	snaps, err := loadSnapshots(filepath.Join(*inDir, "research_plan.json"))
	must(err)
	var counters []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	must(loadJSON(filepath.Join(*inDir, "replan_counters.json"), &counters))
	fmt.Printf("cfg: strategy=%s k=%.2f max_levels=%d cap=%d tfs=%v\n", cfg.StrategyID, cfg.ProximityK, cfg.MaxLevels, cfg.ReplanCap, cfg.Timeframes)
	fmt.Printf("bars: %d series, merged1m=%d · plans=%d · snapshots=%d · counters=%d\n", len(bars.byKey), len(bars.merged1m), len(plans), len(snaps), len(counters))

	loc := kernel.CTLocation()

	// ── sessions ────────────────────────────────────────────────────────
	sessions := map[string]*sessionRow{}
	var sessKeys []string
	for _, p := range plans {
		key := p.TradeDate + ":" + p.Session
		if sessions[key] != nil {
			sessions[key].Reads++
			continue
		}
		_, ws, we := sessionWindow(dayOf(p.TradeDate), p.Session)
		contract, _ := bars.contractAt(ws.UnixMilli())
		sb := bars.rangeBars1m(contract, ws.UnixMilli(), we.UnixMilli())
		row := &sessionRow{Key: key, TradeDate: p.TradeDate, Session: p.Session, Contract: contract,
			WinStart: ws.UnixMilli(), WinEnd: we.UnixMilli(), Bars: len(sb), Reads: 1}
		if len(sb) > 0 {
			row.Measured = true
			row.Open, row.Close = sb[0].Open, sb[len(sb)-1].Close
			row.High, row.Low = math.Inf(-1), math.Inf(1)
			for _, b := range sb {
				row.High = math.Max(row.High, b.High)
				row.Low = math.Min(row.Low, b.Low)
			}
			row.Range = row.High - row.Low
		}
		sessions[key] = row
		sessKeys = append(sessKeys, key)
	}
	sort.Strings(sessKeys)

	// ── reads ───────────────────────────────────────────────────────────
	// snapshot join: for each plan row, the newest input snapshot of the same
	// (TradeDate, Session) captured at or before created_at and AFTER the
	// previous plan row of that session-day (each read consumes one input).
	prevReadMs := map[string]int64{}
	var reads []readRow
	var scens []scenarioRow
	for _, p := range plans {
		key := p.TradeDate + ":" + p.Session
		sess := sessions[key]
		readT := time.UnixMilli(p.createdMs).In(loc)
		contract, _ := bars.contractAt(p.createdMs)
		isRead := strings.HasSuffix(p.TriggerReason, "_scheduled_read") || p.TriggerReason == "level_event" ||
			p.TriggerReason == "death_replan" || p.TriggerReason == "structure_mss" || p.TriggerReason == "owner_reread"
		r := readRow{ID: fmt.Sprintf("%s@%d", p.PlanID, p.Version), PlanID: p.PlanID, Version: p.Version, IsRead: isRead,
			TradeDate: p.TradeDate, Session: p.Session, SessKey: key, Trigger: p.TriggerReason, Lifecycle: p.Lifecycle,
			ReadMs: p.createdMs, ReadCT: readT.Format("2006-01-02 15:04:05 MST"), Contract: contract, Bias: string(p.doc.Bias),
			BandK: cfg.ProximityK, Trend: map[string]string{}, TrendRing: map[string]int{}}

		var snap *inputSnapshot
		for i := len(snaps) - 1; i >= 0; i-- {
			s := &snaps[i]
			if s.TradeDate != p.TradeDate || s.Session != p.Session || s.CapturedMs > p.createdMs {
				continue
			}
			if s.CapturedMs <= prevReadMs[key] {
				break
			}
			snap = s
			break
		}
		prevReadMs[key] = p.createdMs

		ring := bars.lastClosed(contract, "1m", kernel.AISVPBarCount, readT)
		if snap != nil {
			r.Recorded, r.SnapID, r.SnapLagS = true, snap.ID, float64(p.createdMs-snap.CapturedMs)/1000
			r.Price, r.PriceSrc = snap.Price, "snapshot"
			r.DATR, r.DATRSrc = snap.DATR, "snapshot"
		} else {
			if len(ring) > 0 {
				r.Price, r.PriceSrc = ring[len(ring)-1].Close, "bar_close"
			}
			r.DATR, r.DATRSrc = kernel.DailyRangeProxy(ring, readT), "DailyRangeProxy"
			if r.DATR <= 0 && r.Price > 0 {
				r.DATR = 0.008 * r.Price // the assembler's own cold fallback
			}
		}
		r.Band = r.BandK * r.DATR

		// outcome bars: [max(read, winStart), winEnd)
		from := p.createdMs
		if sess.WinStart > from {
			from = sess.WinStart
		}
		out := bars.rangeBars1m(contract, from, sess.WinEnd)
		r.OutBars = len(out)
		if len(out) > 0 && r.Price > 0 {
			hi, lo := math.Inf(-1), math.Inf(1)
			for _, b := range out {
				hi = math.Max(hi, b.High)
				lo = math.Min(lo, b.Low)
			}
			r.ExcUp, r.ExcDown = hi-r.Price, r.Price-lo
			if r.ExcUp >= r.ExcDown {
				r.DirExc, r.Excursion = "up", r.ExcUp
			} else {
				r.DirExc, r.Excursion = "down", r.ExcDown
			}
			r.DirClose = sign(out[len(out)-1].Close - r.Price)
			r.Measured = true
		}

		// level tables
		if r.Measured {
			if snap != nil {
				ps, tfs := []float64{}, []string{}
				for _, l := range snap.Levels {
					ps = append(ps, l.Price)
					tfs = append(tfs, l.TF)
					if a := math.Abs(l.Distance); a > r.SeatedMaxAbsDist {
						r.SeatedMaxAbsDist = a
					}
				}
				r.Seated = tableStats(ps, tfs, r.Price, r.Band, r.Excursion, r.DirExc)
			}
			ps, tfs := []float64{}, []string{}
			for _, l := range p.doc.Levels {
				ps = append(ps, l.Price)
				tfs = append(tfs, l.TF)
			}
			r.Doc = tableStats(ps, tfs, r.Price, r.Band, r.Excursion, r.DirExc)
		}

		// HTF universe at the read — ALWAYS reconstructed with the production
		// detector over the DB bars: the planner's recorded HTFZonesFull is
		// ScoreLevels-filtered to |price−p| ≤ band and zones-only
		// (auto_trader_planner.go:2264, levels_score.go:475), so by construction
		// it holds no beyond-band level. For recorded reads the in-band zone
		// count of this reconstruction is compared with the recorded count
		// (parity probe), and the recorded roles are carried onto matches.
		type htfL struct {
			kind, tf, role string
			price, lo, hi  float64
		}
		var htf []htfL
		r.HTFSrc = "DetectHTFLevels"
		fetch := func(tf string, count int) []market.Kline {
			if tf == "D" {
				tf = "1d"
			}
			return bars.lastClosed(contract, tf, count, readT)
		}
		for _, l := range kernel.DetectHTFLevels(fetch, cfg.Timeframes, "MNQ", readT) {
			htf = append(htf, htfL{string(l.Kind), l.TF, "", l.Price, l.Lo, l.Hi})
		}
		if snap != nil {
			r.HTFParityRecordedN = len(snap.HTFZonesFull)
			for i := range htf {
				l := &htf[i]
				switch l.kind {
				case "SUPPLY", "DEMAND", "FVG", "OB":
					if math.Abs(l.price-r.Price) <= r.Band {
						r.HTFParityReconN++
					}
				}
				for _, s := range snap.HTFZonesFull {
					if s.TF == l.tf && math.Abs(s.Price-l.price) < 0.01 {
						l.role = s.Role
						break
					}
				}
			}
		}
		r.HTFN = len(htf)
		for _, l := range htf {
			if math.Abs(l.price-r.Price) > r.Band {
				r.HTFBeyondBand++
			}
		}

		// S1 trend per TF ring (production ComputeStructureState), own contract only.
		for _, tf := range []string{"1h", "4h", "1d"} {
			minutes := int(tfMillis[tf] / 60_000)
			ks := bars.lastClosed(contract, tf, 500, readT)
			name := "s1_" + map[string]string{"1h": "1h", "4h": "4h", "1d": "D"}[tf]
			r.TrendRing[name] = len(ks)
			if len(ks) < 60 {
				r.Trend[name] = "n/a"
				continue
			}
			r.Trend[name] = trendWord(kernel.ComputeStructureState(toKB(ks), minutes, 0, p.createdMs).Trend)
		}
		r.Trend["plan_bias"] = map[string]string{"long": "up", "short": "down"}[strings.ToLower(string(p.doc.Bias))]
		if r.Trend["plan_bias"] == "" {
			r.Trend["plan_bias"] = "range"
		}

		// candidate rule: nearest BEYOND-band HTF level (tf 1h/4h/D) in the trend direction.
		for _, def := range []string{"s1_1h", "s1_4h", "s1_D", "plan_bias"} {
			c := candidate{TrendDef: def, Trend: r.Trend[def], RingN: r.TrendRing[def]}
			if c.Trend == "up" || c.Trend == "down" {
				best := -1.0
				for _, l := range htf {
					switch l.tf {
					case "1h", "4h", "D", "1d":
					default:
						continue
					}
					d := l.price - r.Price
					if sign(d) != c.Trend || math.Abs(d) <= r.Band {
						continue
					}
					if best < 0 || math.Abs(d) < best {
						best = math.Abs(d)
						c.Found, c.Kind, c.TF, c.Role, c.Price, c.Dist = true, l.kind, l.tf, l.role, l.price, math.Abs(d)
						// near edge of a zone for the touch test
						c.Edge = l.price
						if l.lo > 0 && l.hi > 0 && l.hi >= l.lo {
							if c.Trend == "up" {
								c.Edge = l.lo
							} else {
								c.Edge = l.hi
							}
						}
					}
				}
				if c.Found && r.Measured {
					c.Measured = true
					if m := firstTouch(out, p.createdMs, c.Edge, c.Trend); m >= 0 {
						c.Hit, c.Minutes = true, m
					}
				} else if !c.Found {
					c.Measured = r.Measured // trend known, no such level: measured "none"
				}
			}
			r.Cands = append(r.Cands, c)
		}

		// in-band FARTHEST HTF level in the trend direction (alternative rule)
		for _, def := range []string{"s1_1h", "s1_4h", "s1_D", "plan_bias"} {
			c := candidate{TrendDef: def, Trend: r.Trend[def], RingN: r.TrendRing[def]}
			if c.Trend == "up" || c.Trend == "down" {
				best := -1.0
				for _, l := range htf {
					switch l.tf {
					case "1h", "4h", "D", "1d":
					default:
						continue
					}
					d := l.price - r.Price
					if sign(d) != c.Trend || math.Abs(d) > r.Band {
						continue
					}
					if math.Abs(d) > best {
						best = math.Abs(d)
						c.Found, c.Kind, c.TF, c.Role, c.Price, c.Dist = true, l.kind, l.tf, l.role, l.price, math.Abs(d)
						c.Edge = l.price
						if l.lo > 0 && l.hi > 0 && l.hi >= l.lo {
							if c.Trend == "up" {
								c.Edge = l.lo
							} else {
								c.Edge = l.hi
							}
						}
					}
				}
				if c.Found && r.Measured {
					c.Measured = true
					if mnt := firstTouch(out, p.createdMs, c.Edge, c.Trend); mnt >= 0 {
						c.Hit, c.Minutes = true, mnt
					}
				} else if !c.Found {
					c.Measured = r.Measured
				}
			}
			r.CandsInBand = append(r.CandsInBand, c)
		}

		// oracle: in-band HTF levels on the eventual-move side beyond the doc
		// table's farthest ahead — the targets that existed and were not seated.
		if r.Measured && r.Doc != nil {
			for _, l := range htf {
				switch l.tf {
				case "1h", "4h", "D", "1d":
				default:
					continue
				}
				d := l.price - r.Price
				if sign(d) != r.DirExc || math.Abs(d) > r.Band || math.Abs(d) <= r.Doc.Farthest {
					continue
				}
				r.OracleMissedN++
				edge := l.price
				if l.lo > 0 && l.hi > 0 && l.hi >= l.lo {
					if r.DirExc == "up" {
						edge = l.lo
					} else {
						edge = l.hi
					}
				}
				hit := firstTouch(out, p.createdMs, edge, r.DirExc) >= 0
				if hit {
					r.OracleMissedHitN++
				}
				if len(r.OracleMissedIDs) < 3 {
					r.OracleMissedIDs = append(r.OracleMissedIDs, fmt.Sprintf("%s·%s@%.2f%s", l.kind, l.tf, l.price, map[bool]string{true: "✓", false: ""}[hit]))
				}
			}
		}

		if r.Measured && r.Seated != nil {
			for _, l := range htf {
				switch l.tf {
				case "1h", "4h", "D", "1d":
				default:
					continue
				}
				d := l.price - r.Price
				if sign(d) != r.DirExc || math.Abs(d) > r.Band || math.Abs(d) <= r.Seated.Farthest {
					continue
				}
				r.OracleSeatedMissedN++
				edge := l.price
				if l.lo > 0 && l.hi > 0 && l.hi >= l.lo {
					if r.DirExc == "up" {
						edge = l.lo
					} else {
						edge = l.hi
					}
				}
				if firstTouch(out, p.createdMs, edge, r.DirExc) >= 0 {
					r.OracleSeatedMissedHitN++
				}
			}
		}

		// scenarios (b)
		for _, sc := range p.doc.Scenarios {
			if len(sc.TargetChain) == 0 {
				continue
			}
			dir := map[string]string{"long": "up", "short": "down"}[strings.ToLower(sc.Direction)]
			s := scenarioRow{ReadID: r.ID, TradeDate: p.TradeDate, Session: p.Session, SessKey: key, ReadMs: p.createdMs,
				ScenID: sc.ID, Direction: sc.Direction, Price: r.Price, NTargets: len(sc.TargetChain),
				First: sc.TargetChain[0], Last: sc.TargetChain[len(sc.TargetChain)-1]}
			if dir == "" {
				s.Malformed = "direction"
			} else {
				s.LastDist = s.Last - r.Price
				if dir == "down" {
					s.LastDist = -s.LastDist
				}
				if s.LastDist <= 0 {
					s.Malformed = "last_target_behind_price"
				}
			}
			if r.Measured && s.Malformed == "" {
				s.Measured = true
				s.HitFirst = firstTouch(out, p.createdMs, s.First, dir) >= 0
				if m := firstTouch(out, p.createdMs, s.Last, dir); m >= 0 {
					s.Exhausted, s.MinLast = true, m
				}
			}
			scens = append(scens, s)
		}
		reads = append(reads, r)
	}

	// ── day-plan budget (d) ─────────────────────────────────────────────
	var dayplan []dayplanRow
	for _, key := range sessKeys {
		s := sessions[key]
		d := dayplanRow{Key: key, TradeDate: s.TradeDate, Session: s.Session, Triggers: map[string]int{}, Cap: cfg.ReplanCap}
		for _, p := range plans {
			if p.TradeDate+":"+p.Session != key {
				continue
			}
			d.Reads++
			t := p.TriggerReason
			switch {
			case t == "level_event":
				d.LevelEvent++
			case t == "death_replan" || t == "owner_reread":
				d.Spending++
			case strings.HasPrefix(t, "dormant:") || strings.HasPrefix(t, "rearmed:"):
				t = strings.SplitN(t, ":", 3)[0] + ":" + strings.SplitN(t+"::", ":", 3)[1]
			}
			d.Triggers[t]++
		}
		for _, c := range counters {
			if !strings.HasPrefix(c.Key, "dayplan_replans_used:") {
				continue
			}
			parts := strings.Split(c.Key, ":")
			if len(parts) >= 5 && parts[2] == s.TradeDate && parts[3] == s.Session {
				var n int
				fmt.Sscanf(c.Value, "%d", &n)
				if n > d.CounterUsed {
					d.CounterUsed = n
				}
				d.CounterKeys = append(d.CounterKeys, parts[2]+":"+parts[3]+":"+parts[4])
			}
		}
		d.Exhausted = d.CounterUsed >= d.Cap
		d.WouldExhaustIfLevelEventsCounted = d.LevelEvent+d.Spending >= d.Cap
		dayplan = append(dayplan, d)
	}

	// ── write ───────────────────────────────────────────────────────────
	writeJSONL := func(name string, rows any) {
		f, err := os.Create(filepath.Join(*outDir, name))
		must(err)
		defer f.Close()
		enc := json.NewEncoder(f)
		switch v := rows.(type) {
		case []*sessionRow:
			for _, r := range v {
				must(enc.Encode(r))
			}
		case []readRow:
			for _, r := range v {
				must(enc.Encode(r))
			}
		case []scenarioRow:
			for _, r := range v {
				must(enc.Encode(r))
			}
		case []dayplanRow:
			for _, r := range v {
				must(enc.Encode(r))
			}
		}
	}
	var sessList []*sessionRow
	for _, k := range sessKeys {
		sessList = append(sessList, sessions[k])
	}
	writeJSONL("sessions.jsonl", sessList)
	writeJSONL("reads.jsonl", reads)
	writeJSONL("scenarios.jsonl", scens)
	writeJSONL("dayplan.jsonl", dayplan)
	rec := 0
	for _, r := range reads {
		if r.Recorded {
			rec++
		}
	}
	fmt.Printf("sessions=%d reads=%d (recorded snapshots %d) scenarios=%d dayplan=%d\n", len(sessList), len(reads), rec, len(scens), len(dayplan))
}
