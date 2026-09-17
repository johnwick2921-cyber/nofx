package main

// series.go — bars from the exported JSON files (export.sh; every database read
// in this study goes through `sqlite3 -readonly`; the harness never opens a DB).
//
// CONTRACT BASIS. The store carries two price bases (~+290pt Sep→Dec) under
// labels that do NOT track the basis (the AddOn's date rule relabelled to
// "MNQ 12-26" before the live tape actually rolled — memory: bar-source wave).
// Measured against the live 1m tape (probe in the report, §basis):
//   1m/15m : Sep basis until the tape's +294.25pt one-minute jump at
//            2026-09-14 10:26 CT (open 1789399560000); Dec basis from that bar on.
//            (both labels; the 15m "12-26 live" rows before the flip diff 0 vs tape)
//   1h/4h  : label IS basis — "MNQ 09-26" rows Sep, "MNQ 12-26" rows Dec
//            ("12-26 live" 1h/4h rows closed BEFORE the flip sit +289.5..+300 above the tape)
//   1d     : label IS basis (same rule as 1h/4h; [B] — the daily rows cannot be
//            cross-checked minute-by-minute)
//   1m non-live rows: 09-26 historical(_import) Sep (diff 0..3); 12-26 historical(_import) Dec (+282..+292)
// Every series handed to the kernel is expressed in ONE basis per read (the
// basis the plan was written in, detected from its level prices — main.go), by
// shifting rows of the other basis by the measured step. A constant shift
// preserves swing/level ORDER; a level price that came from a shifted row is a
// basis-adjusted price, not a traded one ([B]).
//
// Sources 'mixed' and 'replay:off-scale' are excluded exactly as
// store.BarsBetweenOn excludes them.

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"nofx/market"
)

const (
	contractSep = "MNQ 09-26"
	contractDec = "MNQ 12-26"
	basisSep    = "sep"
	basisDec    = "dec"
	// tapeFlipMs is the open of the first Dec-basis 1m bar on the live tape
	// (2026-09-14 10:26 CT): the bar whose open sits +294.25 above the previous close.
	tapeFlipMs = int64(1789399560000)
)

var tfMillis = map[string]int64{
	"1m": 60_000, "5m": 300_000, "15m": 900_000, "1h": 3_600_000, "4h": 14_400_000, "1d": 86_400_000,
}

type barRow struct {
	TF       string  `json:"tf"`
	OpenMs   int64   `json:"open_time_ms"`
	O        float64 `json:"o"`
	H        float64 `json:"h"`
	L        float64 `json:"l"`
	C        float64 `json:"c"`
	V        float64 `json:"v"`
	Contract string  `json:"contract"`
	Source   string  `json:"source"`
	basis    string
}

// rowBasis applies the measured rule above.
func rowBasis(r barRow) string {
	switch r.TF {
	case "1m", "15m":
		if r.Source == "live" {
			if r.OpenMs >= tapeFlipMs {
				return basisDec
			}
			return basisSep
		}
	}
	if r.Contract == contractDec {
		return basisDec
	}
	return basisSep
}

type barStore struct {
	rows  map[string][]barRow // per tf, sorted by open (one row per open: the store PK)
	step  float64             // Dec minus Sep, points (pooled median, see measureStep)
	stepN int
	stepLo, stepHi float64
}

func loadJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func loadBars(outDir string) (*barStore, error) {
	byKey := map[string]map[int64]barRow{}
	for _, f := range []string{"copy_bars_1m.json", "copy_bars_15m.json", "copy_bars_1h.json", "copy_bars_4h.json", "copy_bars_1d.json", "live_bars_tail.json"} {
		var rows []barRow
		if err := loadJSON(outDir+"/"+f, &rows); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		n := 0
		for _, r := range rows { // live_bars_tail is LAST: a live row replaces the copy's on the same (tf, open)
			if r.Source == "mixed" || r.Source == "replay:off-scale" || r.Contract == "" || r.Contract == "unrecomputable:spans_roll" {
				continue
			}
			if byKey[r.TF] == nil {
				byKey[r.TF] = map[int64]barRow{}
			}
			r.basis = rowBasis(r)
			byKey[r.TF][r.OpenMs] = r
			n++
		}
		fmt.Printf("loaded %s rows=%d kept=%d\n", f, len(rows), n)
	}
	st := &barStore{rows: map[string][]barRow{}}
	for tf, m := range byKey {
		out := make([]barRow, 0, len(m))
		for _, r := range m {
			out = append(out, r)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].OpenMs < out[j].OpenMs })
		st.rows[tf] = out
	}
	st.measureStep()
	return st, nil
}

// measureStep: pooled median of (row.close − tape close just before the row's
// close) over the 15m/1h/4h rows that are Dec-basis while the 1m tape is still
// Sep-basis (row closes before tapeFlipMs), plus the tape's own flip jump.
func (s *barStore) measureStep() {
	tape := s.rows["1m"]
	closeBefore := func(ms int64) (float64, bool) {
		i := sort.Search(len(tape), func(i int) bool { return tape[i].OpenMs >= ms }) - 1
		if i < 0 || ms-tape[i].OpenMs > 15*60_000 || tape[i].basis != basisSep {
			return 0, false
		}
		return tape[i].C, true
	}
	var d []float64
	for _, tf := range []string{"15m", "1h", "4h"} {
		w := tfMillis[tf]
		for _, r := range s.rows[tf] {
			end := r.OpenMs + w
			if r.basis != basisDec || end > tapeFlipMs {
				continue
			}
			if c, ok := closeBefore(end); ok {
				d = append(d, r.C-c)
			}
		}
	}
	// the tape's own one-minute flip jump
	i := sort.Search(len(tape), func(i int) bool { return tape[i].OpenMs >= tapeFlipMs })
	if i > 0 && i < len(tape) && tape[i].OpenMs == tapeFlipMs {
		d = append(d, tape[i].O-tape[i-1].C)
	}
	if len(d) == 0 {
		panic("basis step unmeasurable")
	}
	sort.Float64s(d)
	s.step = d[len(d)/2]
	s.stepN = len(d)
	s.stepLo, s.stepHi = d[0], d[len(d)-1]
}

func (s *barStore) shift(rowBasis, basis string) float64 {
	switch {
	case rowBasis == basis:
		return 0
	case rowBasis == basisSep && basis == basisDec:
		return s.step
	default:
		return -s.step
	}
}

func (s *barStore) kline(r barRow, basis string) market.Kline {
	w := tfMillis[r.TF]
	d := s.shift(r.basis, basis)
	return market.Kline{OpenTime: r.OpenMs, CloseTime: r.OpenMs + w - 1, Open: r.O + d, High: r.H + d, Low: r.L + d, Close: r.C + d, Volume: r.V}
}

// lastClosed returns the last n bars of tf with CloseTime < beforeMs, i.e.
// open_time + interval <= read — the lookahead rule: no bar still forming at
// the read participates. Ascending, in the given basis. n<=0 → all.
func (s *barStore) lastClosed(tf string, n int, beforeMs int64, basis string) []market.Kline {
	rows := s.rows[tf]
	w := tfMillis[tf]
	hi := sort.Search(len(rows), func(i int) bool { return rows[i].OpenMs+w-1 >= beforeMs })
	lo := hi - n
	if n <= 0 || lo < 0 {
		lo = 0
	}
	out := make([]market.Kline, 0, hi-lo)
	for _, r := range rows[lo:hi] {
		out = append(out, s.kline(r, basis))
	}
	return out
}

// between returns tf bars with fromMs <= open < toMs, ascending, in basis.
func (s *barStore) between(tf string, fromMs, toMs int64, basis string) []market.Kline {
	rows := s.rows[tf]
	a := sort.Search(len(rows), func(i int) bool { return rows[i].OpenMs >= fromMs })
	b := sort.Search(len(rows), func(i int) bool { return rows[i].OpenMs >= toMs })
	if b <= a { // a read written after the flat has no forward window
		return nil
	}
	out := make([]market.Kline, 0, b-a)
	for _, r := range rows[a:b] {
		out = append(out, s.kline(r, basis))
	}
	return out
}
