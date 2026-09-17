package main

// r25.go — R25 SEAT DISPLACEMENT pass (CTO dispatch 2026-09-17 16:30Z).
//
// Per read: replicate the PLANNER's assemble at the LIVE config
// (day_plan.max_levels=12, proximity_filter_atr=1.0 read from the DB copy's
// strategies table, htf_seats unset → seatHTF default, htf_score_multiplier
// unset → 1.2, min_grade unset) via the exported production call
// kernel.AssembleResearchLevels(traderID="hoang", bars1m, DefaultSessionRegistry,
// "MNQ", 12, nil, 1.2, readTime, 1.0, "") → (seated, pool, price, dATR, raw).
// The freshness provider is NOT installed (harness convention, byte-identical
// to the pre-W11b goldens: kernel levelFreshnessFn default → all-fresh) — the
// SAME convention the accepted S4 gate used. Config-as-of is NOT reconstructed
// (the live config is applied to all reads) — a stated limit.
//
// Proximity band (quoted): kernel/levels_score.go:465 `band := proximityK *
// dATR` inside scoreLevelsPool; proximityK = the planner's
// proximityFilterATR() = day_plan.proximity_filter_atr (1.0 live; the ≤0
// fallback is ActivationWindowK=1.5, levels_score.go:460-463). dATR =
// kernel.DailyRangeProxy(bars, now).
//
// Output r25-reads.jsonl, one row per read: day, session, read_at_ms, contract,
// price, d_atr, band, proximity_k, max_levels, n_seats, seats[] (label, kind,
// tf, htf, price, score, seat_order), farthest in-band HTF per side
// (above/below: label, kind, tf, price, dist, seated), session_range_pts,
// window_complete. Session range = max High − min Low over the session window
// [winStart, flat) 1m bars; complete = the window's bars are all present
// (first open == winStart and last open == flat−60s) — the big-day flag
// (range ≥ 250pt, complete) recomputed INDEPENDENTLY of Chief's Q4 outputs
// per the audit lesson.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"nofx/kernel"
)

// live config (read from the DB copy strategies table; hardcoded here as the
// values that query returned — see README for the query).
const (
	r25MaxLevels = 12
	r25ProxK     = 1.0
	r25HtfMult   = 1.2
	r25MinGrade  = ""
	r25TraderID  = "hoang"
	r25BigDayPts = 250.0
)

type r25Seat struct {
	Label string  `json:"label"`
	Kind  string  `json:"kind"`
	TF    string  `json:"tf"`
	HTF   bool    `json:"htf"`
	Price float64 `json:"price"`
	Score float64 `json:"score"`
	Order int     `json:"seat_order"`
}

type r25Side struct {
	Label  string  `json:"label"`
	Kind   string  `json:"kind"`
	TF     string  `json:"tf"`
	Price  float64 `json:"price"`
	Dist   float64 `json:"dist_pts"`
	Seated bool    `json:"seated"`
}

type r25Read struct {
	Day             string    `json:"day"`
	Session         string    `json:"session"`
	ReadAtMs        int64     `json:"read_at_ms"`
	Contract        string    `json:"contract"`
	Price           float64   `json:"price"`
	DATR            float64   `json:"d_atr"`
	Band            float64   `json:"band_pts"`
	ProxK           float64   `json:"proximity_k"`
	MaxLevels       int       `json:"max_levels"`
	NSeats          int       `json:"n_seats"`
	Seats           []r25Seat `json:"seats"`
	FarthestAbove   *r25Side  `json:"farthest_above,omitempty"`
	FarthestBelow   *r25Side  `json:"farthest_below,omitempty"`
	SessionRangePts float64   `json:"session_range_pts"`
	WindowComplete  bool      `json:"window_complete"`
	BigDay          bool      `json:"big_day"`
}

// runR25 executes the seat-displacement pass over the given reads.
func runR25(bd *barDB, reads []*readSnapshot, outDir string) error {
	out, err := os.Create(filepath.Join(outDir, "r25-reads.jsonl"))
	if err != nil {
		return err
	}
	defer out.Close()
	w := bufio.NewWriterSize(out, 1<<20)
	defer w.Flush()

	n, nBig := 0, 0
	for _, r := range reads {
		row, ok := buildR25(bd, r)
		if !ok {
			continue
		}
		line, err := json.Marshal(row)
		if err != nil {
			return err
		}
		w.Write(line)
		w.WriteByte('\n')
		n++
		if row.BigDay {
			nBig++
		}
	}
	fmt.Printf("r25 reads: %d (big days: %d)\n", n, nBig)
	return nil
}

// buildR25 assembles one read with the production seating call.
func buildR25(bd *barDB, r *readSnapshot) (r25Read, bool) {
	bars := r.Bars1m
	if len(bars) < 2 {
		return r25Read{}, false
	}
	now := r.ReadTime
	seated, pool, price, dATR, _ := kernel.AssembleResearchLevels(
		r25TraderID, bars, kernel.DefaultSessionRegistry(), "MNQ",
		r25MaxLevels, nil, r25HtfMult, now, r25ProxK, r25MinGrade)
	if price <= 0 {
		return r25Read{}, false
	}
	band := r25ProxK * dATR
	row := r25Read{
		Day: r.Day, Session: r.Session, ReadAtMs: now.UnixMilli(),
		Contract: r.Contract, Price: price, DATR: dATR, Band: band,
		ProxK: r25ProxK, MaxLevels: r25MaxLevels, NSeats: len(seated),
	}
	seatedPrice := map[float64]bool{}
	for i, s := range seated {
		seatedPrice[s.Price] = true
		row.Seats = append(row.Seats, r25Seat{
			Label: s.Label, Kind: string(s.Kind), TF: s.TF, HTF: s.HTF,
			Price: s.Price, Score: s.Score, Order: i + 1,
		})
	}
	// farthest in-band HTF per side from the pre-cap pool
	for _, p := range pool {
		if !p.HTF {
			continue
		}
		if p.TF != "1d" && p.TF != "1w" && p.TF != "4h" {
			continue
		}
		dist := p.Price - price
		side := &row.FarthestAbove
		if dist < 0 {
			side = &row.FarthestBelow
			dist = -dist
		}
		cand := &r25Side{
			Label: p.Label, Kind: string(p.Kind), TF: p.TF, Price: p.Price,
			Dist: dist, Seated: seatedPrice[p.Price],
		}
		if *side == nil || dist > (*side).Dist {
			*side = cand
		}
	}
	// session range + completeness (the big-day flag, independent recompute)
	_, winBars := bd.sessionBars1m(r.Contract, r.WinStartMs, r.FlatMs)
	if len(winBars) > 0 {
		hi, lo := winBars[0].High, winBars[0].Low
		for _, b := range winBars {
			if b.High > hi {
				hi = b.High
			}
			if b.Low < lo {
				lo = b.Low
			}
		}
		row.SessionRangePts = hi - lo
		first, last := winBars[0].OpenTime, winBars[len(winBars)-1].OpenTime
		row.WindowComplete = first == r.WinStartMs && last == r.FlatMs-60_000
		row.BigDay = row.WindowComplete && row.SessionRangePts >= r25BigDayPts
	}
	return row, true
}
