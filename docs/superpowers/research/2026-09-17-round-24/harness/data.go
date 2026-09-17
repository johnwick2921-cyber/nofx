package main

// data.go — Round 24 Q4 inputs. Everything here reads the read-only dumps that
// extract.sh produced with `sqlite3 -readonly`; no database is opened by Go.

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"nofx/kernel"
	"nofx/market"
)

var tfMillis = map[string]int64{
	"1m": 60_000, "3m": 180_000, "5m": 300_000, "15m": 900_000, "30m": 1_800_000,
	"1h": 3_600_000, "2h": 7_200_000, "4h": 14_400_000, "1d": 86_400_000,
}

type bar struct {
	openMs     int64
	o, h, l, c float64
	v          float64
}

type seriesKey struct{ contract, tf string }

type barDB struct {
	byKey    map[seriesKey][]bar
	merged1m []struct {
		openMs   int64
		contract string
	}
}

// loadBarsCSV reads one extract (bars.csv or bars_htf.csv) into db.
func loadBarsCSV(db *barDB, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	head, err := r.Read()
	if err != nil {
		return err
	}
	col := map[string]int{}
	for i, h := range head {
		col[h] = i
	}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		tf := rec[col["tf"]]
		contract := rec[col["contract"]]
		open, _ := strconv.ParseInt(rec[col["open_time_ms"]], 10, 64)
		o, _ := strconv.ParseFloat(rec[col["o"]], 64)
		h, _ := strconv.ParseFloat(rec[col["h"]], 64)
		l, _ := strconv.ParseFloat(rec[col["l"]], 64)
		c, _ := strconv.ParseFloat(rec[col["c"]], 64)
		v, _ := strconv.ParseFloat(rec[col["v"]], 64)
		k := seriesKey{contract, tf}
		db.byKey[k] = append(db.byKey[k], bar{open, o, h, l, c, v})
	}
	return nil
}

func newBarDB(paths ...string) (*barDB, error) {
	db := &barDB{byKey: map[seriesKey][]bar{}}
	for _, p := range paths {
		if err := loadBarsCSV(db, p); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
	}
	// dedupe (bars.csv and bars_htf.csv overlap on 5m/15m/1h/4h/1d for the
	// last 38 days) and sort every series.
	for k, rows := range db.byKey {
		sort.Slice(rows, func(i, j int) bool { return rows[i].openMs < rows[j].openMs })
		out := rows[:0]
		for i, r := range rows {
			if i > 0 && rows[i-1].openMs == r.openMs {
				continue
			}
			out = append(out, r)
		}
		db.byKey[k] = out
		if k.tf == "1m" {
			for _, r := range out {
				db.merged1m = append(db.merged1m, struct {
					openMs   int64
					contract string
				}{r.openMs, k.contract})
			}
		}
	}
	sort.Slice(db.merged1m, func(i, j int) bool { return db.merged1m[i].openMs < db.merged1m[j].openMs })
	return db, nil
}

// contractAt emulates store.ContractAt: the contract of the newest usable 1m
// bar at or before ms (same rule as the round-23 harness).
func (b *barDB) contractAt(ms int64) (string, bool) {
	idx := sort.Search(len(b.merged1m), func(i int) bool { return b.merged1m[i].openMs > ms })
	if idx == 0 {
		return "", false
	}
	return b.merged1m[idx-1].contract, true
}

func toKline(rows []bar, tf string) []market.Kline {
	width := tfMillis[tf]
	if width <= 0 {
		width = 60_000
	}
	out := make([]market.Kline, 0, len(rows))
	for _, r := range rows {
		out = append(out, market.Kline{OpenTime: r.openMs, CloseTime: r.openMs + width - 1,
			Open: r.o, High: r.h, Low: r.l, Close: r.c, Volume: r.v})
	}
	return out
}

// lastClosed: the newest n bars of (contract, tf) whose close is before `before`.
func (b *barDB) lastClosed(contract, tf string, n int, before time.Time) []market.Kline {
	rows := b.byKey[seriesKey{contract, tf}]
	if len(rows) == 0 {
		return nil
	}
	limit := before.UnixMilli()
	width := tfMillis[tf]
	hi := sort.Search(len(rows), func(i int) bool { return rows[i].openMs+width-1 >= limit })
	lo := hi - n
	if lo < 0 {
		lo = 0
	}
	return toKline(rows[lo:hi], tf)
}

// rangeBars1m: 1m bars of contract with fromMs <= open < toMs.
func (b *barDB) rangeBars1m(contract string, fromMs, toMs int64) []market.Kline {
	rows := b.byKey[seriesKey{contract, "1m"}]
	s := sort.Search(len(rows), func(i int) bool { return rows[i].openMs >= fromMs })
	e := sort.Search(len(rows), func(i int) bool { return rows[i].openMs >= toMs })
	if e <= s {
		return nil
	}
	return toKline(rows[s:e], "1m")
}

// sessionWindow — DefaultSessionRegistry values, CT-anchored, same convention
// as the round-23 harness and as plans.trade_date (ASIA: the calendar date of
// the 16:30 read; window 17:00 → 02:00 next day).
func sessionWindow(day time.Time, session string) (readTime, winStart, winEnd time.Time) {
	loc := kernel.CTLocation()
	y, m, d := day.Date()
	switch session {
	case "ASIA":
		readTime = time.Date(y, m, d, 16, 30, 0, 0, loc)
		winStart = time.Date(y, m, d, 17, 0, 0, 0, loc)
		winEnd = time.Date(y, m, d+1, 2, 0, 0, 0, loc)
	case "LONDON":
		readTime = time.Date(y, m, d, 1, 30, 0, 0, loc)
		winStart = time.Date(y, m, d, 2, 0, 0, 0, loc)
		winEnd = time.Date(y, m, d, 8, 30, 0, 0, loc)
	case "NY":
		readTime = time.Date(y, m, d, 8, 0, 0, 0, loc)
		winStart = time.Date(y, m, d, 8, 30, 0, 0, loc)
		winEnd = time.Date(y, m, d, 14, 45, 0, 0, loc)
	}
	return
}

// ── plans.json ──────────────────────────────────────────────────────────────

type planRow struct {
	PlanID        string `json:"plan_id"`
	Version       int    `json:"version"`
	StrategyID    string `json:"strategy_id"`
	TradeDate     string `json:"trade_date"`
	Session       string `json:"session"`
	TriggerReason string `json:"trigger_reason"`
	Lifecycle     string `json:"lifecycle"`
	ModelID       string `json:"model_id"`
	CreatedAt     string `json:"created_at"`
	Degraded      int    `json:"degraded"`
	Doc           string `json:"doc"`

	createdMs int64
	doc       planDoc
}

type docLevel struct {
	Price float64  `json:"price"`
	Label string   `json:"label"`
	Grade string   `json:"grade"`
	TF    string   `json:"tf"`
	Lo    *float64 `json:"lo"`
	Hi    *float64 `json:"hi"`
}

type docScenario struct {
	ID          string    `json:"id"`
	Direction   string    `json:"direction"`
	Condition   string    `json:"condition"`
	TargetChain []float64 `json:"target_chain"`
	Quality     string    `json:"quality"`
}

type planDoc struct {
	Bias      biasField     `json:"bias"`
	BiasLabel string        `json:"bias_label"`
	DayType   string        `json:"day_type"`
	Levels    []docLevel    `json:"levels"`
	Scenarios []docScenario `json:"scenarios"`
}

// biasField tolerates both doc shapes: a bare string ("neutral") in the oldest
// docs and {direction, conviction} everywhere else. Only direction is kept.
type biasField string

func (b *biasField) UnmarshalJSON(raw []byte) error {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		*b = biasField(s)
		return nil
	}
	var o struct {
		Direction string `json:"direction"`
	}
	if err := json.Unmarshal(raw, &o); err != nil {
		return nil // unknown shape → empty, never a parse failure
	}
	*b = biasField(o.Direction)
	return nil
}

var createdLayouts = []string{
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05-07:00",
	"2006-01-02T15:04:05.999999999-07:00",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04:05.999999999+00:00",
}

func parseCreated(s string) (int64, error) {
	for _, l := range createdLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UnixMilli(), nil
		}
	}
	return 0, fmt.Errorf("unparsed created_at %q", s)
}

func loadPlans(path string) ([]planRow, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []planRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	out := rows[:0]
	for _, r := range rows {
		if r.Session != "ASIA" && r.Session != "LONDON" && r.Session != "NY" {
			continue // WEEKLY reads are not session plans
		}
		ms, err := parseCreated(r.CreatedAt)
		if err != nil {
			return nil, err
		}
		r.createdMs = ms
		// target_chain may carry nulls/strings in old docs — tolerate via a
		// loose pass; a chain that fails to parse counts as empty.
		var d planDoc
		if err := json.Unmarshal([]byte(r.Doc), &d); err != nil {
			var loose struct {
				Bias      biasField         `json:"bias"`
				Levels    []docLevel        `json:"levels"`
				Scenarios []json.RawMessage `json:"scenarios"`
			}
			_ = json.Unmarshal([]byte(r.Doc), &loose)
			d = planDoc{Bias: loose.Bias, Levels: loose.Levels} // a partial strict fill must not double the scenarios
			for _, sc := range loose.Scenarios {
				var s docScenario
				if json.Unmarshal(sc, &s) == nil {
					d.Scenarios = append(d.Scenarios, s)
				}
			}
		}
		r.doc = d
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].createdMs < out[j].createdMs })
	return out, nil
}

// ── research_plan.json (object=plan; the planner's own input snapshot) ──────

type snapLevel struct {
	Kind     string  `json:"kind"`
	Price    float64 `json:"price"`
	Lo       float64 `json:"lo"`
	Hi       float64 `json:"hi"`
	Label    string  `json:"label"`
	TF       string  `json:"tf"`
	HTF      bool    `json:"htf"`
	Grade    string  `json:"grade"`
	Distance float64 `json:"distance"`
	Role     string  `json:"role"`
	Score    float64 `json:"score"`
}

type inputSnapshot struct {
	ID               string
	CapturedMs       int64
	TradeDate        string          `json:"TradeDate"`
	Session          string          `json:"Session"`
	Now              string          `json:"Now"`
	Price            float64         `json:"Price"`
	DATR             float64         `json:"DATR"`
	ATR5m            float64         `json:"ATR5m"`
	MaxLevels        int             `json:"MaxLevels"`
	Levels           []snapLevel     `json:"Levels"`
	Pool             []snapLevel     `json:"Pool"`
	HTFZonesFull     []snapLevel     `json:"HTFZonesFull"`
	Regime           json.RawMessage `json:"Regime"`
	StructureSummary json.RawMessage `json:"StructureSummary"`
}

func loadSnapshots(path string) ([]inputSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		ID         int64  `json:"id"`
		Event      string `json:"event"`
		SnapshotID string `json:"snapshot_id"`
		CapturedMs int64  `json:"captured_ms"`
		Fields     string `json:"fields_json"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	var out []inputSnapshot
	for _, r := range rows {
		if r.Event != "input" {
			continue
		}
		var f struct {
			SnapshotID string          `json:"input_snapshot_id"`
			Snapshot   json.RawMessage `json:"input_snapshot"`
		}
		if err := json.Unmarshal([]byte(r.Fields), &f); err != nil || len(f.Snapshot) == 0 || string(f.Snapshot) == "null" {
			continue
		}
		var s inputSnapshot
		if err := json.Unmarshal(f.Snapshot, &s); err != nil {
			continue
		}
		s.ID, s.CapturedMs = f.SnapshotID, r.CapturedMs
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CapturedMs < out[j].CapturedMs })
	return out, nil
}

// ── small json helpers ──────────────────────────────────────────────────────

func loadJSON(path string, v any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

func dayOf(tradeDate string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", strings.TrimSpace(tradeDate), kernel.CTLocation())
	return t
}
