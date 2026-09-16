package main

// main.go — Round 23 ("HTF defines the map, LTF times the entry") harness.
//
// Read-only research: opens a COPY of the live store (never the live file),
// replays the live detection chain at the DefaultSessionRegistry read times
// across the full contract-stamped MNQ era, and runs the D1′ calibrated touch
// instrument (k=3, Δ = trailing-5-day mean |1m close increment| ≈ ±16pt band,
// H=12×1m, exit_on=close, kernel.DetectorK/DetectorHorizonBars) on every
// detected level in the referencing session window.
//
// Outputs (see eval.go): episodes.jsonl, q1_cells.json, q4_ordinals.json.

import (
	"flag"
	"fmt"
	"os"
	"time"

	"nofx/kernel"
)

func main() {
	var (
		dbPath = flag.String("db", "", "path to the DB copy (sqlite, read-only)")
		outDir = flag.String("out", "", "output directory")
		start  = flag.String("start", "2022-04-11", "first CME day YYYY-MM-DD (CT), inclusive")
		end    = flag.String("end", "", "last CME day YYYY-MM-DD (CT), inclusive; empty = last bar day")
		limit  = flag.Int("limit", 0, "debug: stop after N reads (0 = all)")
	)
	flag.Parse()
	if *dbPath == "" || *outDir == "" {
		fmt.Println("usage: -db <copy> -out <dir> [-start YYYY-MM-DD] [-end YYYY-MM-DD] [-limit N]")
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatal("mkdir: %v", err)
	}

	db, err := openCopy(*dbPath)
	if err != nil {
		fatal("open copy: %v", err)
	}
	defer db.Close()

	bd, err := loadBars(db, []string{"1m", "15m", "1h", "4h", "1d"})
	if err != nil {
		fatal("load bars: %v", err)
	}
	if len(bd.merged1m) == 0 {
		fatal("no 1m bars loaded", nil)
	}

	loc := kernel.CTLocation()
	startD, err := time.ParseInLocation("2006-01-02", *start, loc)
	if err != nil {
		fatal("parse -start: %v", err)
	}
	endD := time.UnixMilli(bd.merged1m[len(bd.merged1m)-1].openMs).In(loc)
	endD = time.Date(endD.Year(), endD.Month(), endD.Day(), 0, 0, 0, 0, loc)
	if *end != "" {
		if endD, err = time.ParseInLocation("2006-01-02", *end, loc); err != nil {
			fatal("parse -end: %v", err)
		}
	}
	fmt.Printf("era: %s → %s (CT), %d 1m bars\n",
		startD.Format("2006-01-02"), endD.Format("2006-01-02"), len(bd.merged1m))

	var reads []*readSnapshot
	for day := startD; !day.After(endD); day = day.AddDate(0, 0, 1) {
		for _, sess := range []string{"LONDON", "NY", "ASIA"} {
			r, ok := buildRead(bd, db, len(reads), day, sess)
			if !ok {
				continue
			}
			reads = append(reads, r)
			if *limit > 0 && len(reads) >= *limit {
				goto collected
			}
		}
	}
collected:
	fmt.Printf("reads built: %d\n", len(reads))

	if err := runEval(bd, reads, *outDir); err != nil {
		fatal("eval: %v", err)
	}
	fmt.Println("done")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
