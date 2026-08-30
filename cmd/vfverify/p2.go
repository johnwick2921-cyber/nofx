package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"nofx/kernel"
)

// runP2 — GRADE SYSTEM SPOT-VERIFY (read-only).
func runP2() {
	db := openRO(liveDBPath)
	defer db.Close()

	// ── A. weight-table zero-drift check: CODE vs DOCUMENTED TRUTH ─────────
	fmt.Println("=== P2-A: WEIGHT TABLE — code matrix vs documented truth ===")
	probe := func(k kernel.LevelKind, tf, pattern string, htf bool, fresh func(kernel.DetectedLevel) string, extra ...kernel.DetectedLevel) (float64, string) {
		levels := append([]kernel.DetectedLevel{{
			Kind: k, TF: tf, ZonePattern: pattern, HTF: htf,
			Price: 30000, Lo: 0, Hi: 0, Label: "probe",
		}}, extra...)
		out := kernel.ScoreLevels(levels, 30000, 100, fresh, 8, 1.5)
		for _, s := range out {
			if s.Price == 30000 && s.Kind == k {
				return s.Score, s.Grade
			}
		}
		return 0, "DROPPED"
	}
	// zone evidence (HTF=true survives the conf==0 standalone filter; Lo/Hi=0
	// → sizeMult neutral 1.0; conf=0; fresh → 1.0): score = evidence × tfmult.
	fmt.Println("zoneEvidence (score = evidence·size·fresh·(1+.2·0)·tfmult):")
	for _, k := range []kernel.LevelKind{kernel.KindOB, kernel.KindFVG, kernel.KindSupply} {
		for _, tf := range []string{"1m", "15m", "1h", "4h"} {
			s, g := probe(k, tf, "", true, nil)
			fmt.Printf("  %-7s %-3s score=%.4f grade=%s\n", k, tf, s, g)
		}
	}
	sRev, _ := probe(kernel.KindOB, "1h", "reversal", true, nil)
	sCon, _ := probe(kernel.KindOB, "1h", "continuation", true, nil)
	fmt.Printf("reversal ×1.1 (OB 1h): cont=%.4f rev=%.4f ratio=%.3f\n", sCon, sRev, sRev/sCon)

	fmt.Println("B2 Tier-1 proximity (same zone probes + a PDH anchor within 12 ticks):")
	tier1 := kernel.DetectedLevel{Kind: kernel.KindPDH, Price: 30000.5, Lo: 0, Hi: 0, Label: "PDH"}
	for _, tf := range []string{"1m", "15m", "1h", "4h"} {
		s, g := probe(kernel.KindOB, tf, "", true, nil, tier1)
		fmt.Printf("  OB %-3s +PDH: score=%.4f grade=%s\n", tf, s, g)
	}
	// 4h with confluence next to Tier-1 → may reach A
	sA, gA := probe(kernel.KindOB, "4h", "", true, nil,
		tier1, kernel.DetectedLevel{Kind: kernel.KindRound, Price: 30001, Lo: 0, Hi: 0, Label: "RN", HTF: true})
	fmt.Printf("  OB 4h +PDH +1 conf family: score=%.4f grade=%s (4h may reach A)\n", sA, gA)

	fmt.Println("typeEvidence (non-zone, conf 0, fresh, non-HTF): score = evidence")
	for _, k := range []kernel.LevelKind{kernel.KindPDH, kernel.KindVWAP, kernel.KindSWGH, kernel.KindVAH,
		kernel.KindONH, kernel.KindMIDO, kernel.KindASH, kernel.KindRound, kernel.KindGap, kernel.KindEQL} {
		s, g := probe(k, "", "", false, nil)
		fmt.Printf("  %-7s score=%.2f grade=%s\n", k, s, g)
	}

	fmt.Println("grade thresholds (PDH typeEvidence 1.0, conf in cap window, DISTINCT families):")
	famKinds := []kernel.LevelKind{kernel.KindRound, kernel.KindGap, kernel.KindMIDO, kernel.KindASH, kernel.KindEQH}
	for _, c := range []int{0, 1, 2, 3, 4} {
		extras := []kernel.DetectedLevel{}
		for i := 0; i < c; i++ {
			extras = append(extras, kernel.DetectedLevel{
				Kind: famKinds[i], Price: 30000 + float64(2+i), Lo: 0, Hi: 0,
				Label: "probe", TF: "1m", HTF: true,
			})
		}
		s, g := probe(kernel.KindPDH, "", "", false, nil, extras...)
		fmt.Printf("  conf=%d → score=%.2f grade=%s\n", c, s, g)
	}
	sHTF, _ := probe(kernel.KindPDH, "", "", true, nil)
	sNo, _ := probe(kernel.KindPDH, "", "", false, nil)
	fmt.Printf("HTF ×1.2 (PDH): non-HTF=%.2f HTF=%.2f ratio=%.2f\n", sNo, sHTF, sHTF/sNo)

	fmt.Println("freshness ladders (anchor PDH vs zone OB·1h):")
	for _, f := range []string{"", "b", "c", "tested", "done"} {
		sa, _ := probe(kernel.KindPDH, "", "", true, func(kernel.DetectedLevel) string { return f })
		sb, _ := probe(kernel.KindOB, "1h", "", true, func(kernel.DetectedLevel) string { return f })
		fmt.Printf("  fresh=%-7q anchor=%.4f zone=%.4f\n", f, sa, sb)
	}

	fmt.Println("zoneSizeMult bands (OB 1h, dATR 100):")
	for _, sz := range []float64{0, 10, 25, 50, 80, 120, 220, 300} {
		l := kernel.DetectedLevel{Kind: kernel.KindOB, TF: "1h", HTF: true, Price: 30000, Lo: 30000, Hi: 30000 + sz, Label: "probe"}
		out := kernel.ScoreLevels([]kernel.DetectedLevel{l}, 30000, 100, nil, 8, 1.5)
		for _, s := range out {
			if s.Price == 30000 {
				fmt.Printf("  size=%.0f (%.2f×ATR) score=%.4f\n", sz, sz/100, s.Score)
			}
		}
	}

	fmt.Println("shadow knobs (env-unset defaults):")
	fmt.Printf("  WEEKLY_CONFLUENCE_BAND_ATR=%.2f · WEEKLY_SHADOW_MULT=%.2f (GradeRank A=%d B=%d C=%d)\n",
		kernel.WeeklyConfluenceBandATR(), kernel.WeeklyShadowMult(),
		kernel.GradeRank("A"), kernel.GradeRank("B"), kernel.GradeRank("C"))
	if v := os.Getenv("CONFLUENCE_CAP"); v != "" {
		fmt.Println("  confluence cap: env CONFLUENCE_CAP=" + v)
	} else {
		fmt.Println("  confluence cap: default 3 (env unset)")
	}

	// ── B. machine-grade recompute for the 2026-08-28 NY v7 plan ──────────
	fmt.Println("\n=== P2-B: 2026-08-28 NY v7 — machine-grade recompute at write time ===")
	var rowDoc, rowCreated string
	if err := db.QueryRow(`SELECT doc, created_at FROM plans WHERE trade_date='2026-08-28' AND session='NY' ORDER BY version DESC LIMIT 1`).
		Scan(&rowDoc, &rowCreated); err != nil {
		must(err)
	}
	writeAt, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", rowCreated)
	if err != nil {
		writeAt, err = time.Parse(time.RFC3339Nano, rowCreated)
		must(err)
	}
	cutoffMs := writeAt.UnixMilli()
	fmt.Printf("plan v7 created_at=%s (UTC) → recompute at cutoff %d ms\n", rowCreated, cutoffMs)

	var doc struct {
		Levels []struct {
			Price        float64 `json:"price"`
			Label        string  `json:"label"`
			Grade        string  `json:"grade"`
			MachineGrade string  `json:"machine_grade"`
			Instruction  string  `json:"instruction"`
		} `json:"levels"`
		Reasoning string `json:"reasoning"`
		Bias      struct {
			Direction  string `json:"direction"`
			Conviction string `json:"conviction"`
		} `json:"bias"`
	}
	must(json.Unmarshal([]byte(rowDoc), &doc))
	fmt.Printf("doc bias=%s/%s · %d levels · reasoning opens: %q\n",
		doc.Bias.Direction, doc.Bias.Conviction, len(doc.Levels), firstWords(doc.Reasoning, 60))

	// recompute the machine-grade stamp map from the same detector/scorer
	// pipeline, at the write instant, from stored bars ≤ cutoff.
	as := assemble(db, "NY", "2026-08-28", writeAt, cutoffMs)
	_, minGrade, timeframes := resolvePlanCfg(as.dp, "NY")
	fmt.Printf("recomputed read context: price=%.2f dATR=%.1f proximity=%.2f maxLevels=%d minGrade=%s timeframes=%v\n",
		as.price, as.dATR, kernel.ResolveProximityK(as.dp.ProximityFilterATR),
		as.in.MaxLevels, minGrade, timeframes)

	nearest := func(px float64) *kernel.ScoredLevel {
		var best *kernel.ScoredLevel
		bd := 1e9
		for i := range as.pool {
			d := math.Abs(as.pool[i].Price - px)
			if d < bd {
				bd = d
				best = &as.pool[i]
			}
		}
		if bd > 0.011 {
			return nil
		}
		return best
	}
	fmt.Println("\nprice        label             docG  machG  recomputed  verdict")
	exact, delta, missing := 0, 0, 0
	for _, l := range doc.Levels {
		stamp, ok := as.machineGrades[math.Round(l.Price*100)/100]
		verdict := "EXACT"
		switch {
		case !ok:
			stamp = "—"
			verdict = "NO-STAMP (no pool row at this price)"
			missing++
		case stamp != l.MachineGrade:
			verdict = fmt.Sprintf("DELTA (recomputed %s ≠ stored %s)", stamp, l.MachineGrade)
			delta++
		default:
			exact++
		}
		fmt.Printf("%-12.2f %-16s %-4s  %-5s  %-9s  %s\n", l.Price, l.Label, l.Grade, l.MachineGrade, stamp, verdict)
	}
	fmt.Printf("\nsummary: EXACT=%d DELTA=%d NO-STAMP=%d / %d rows\n", exact, delta, missing, len(doc.Levels))
	fmt.Println("doc-carried components per row: price/label/grade/instruction/machine_grade ONLY —")
	fmt.Println("score · distance · sweep · evidence-counts · freshness · confluence are NOT stored in the plan doc")
	fmt.Println("→ every row is NOT-RECOMPUTABLE-FROM-DOC for those fields (missing component named above);")
	fmt.Println("  the machine-path recompute above (detector+scorer on stored bars ≤ write time) is the independent check.")
	fmt.Println("\npool components for the seated rows (recomputed, write-time):")
	for _, l := range doc.Levels {
		if p := nearest(l.Price); p != nil {
			fmt.Printf("  %-10.2f %-16s kind=%-9s tf=%-3q htf=%-5v fresh=%-6q conf=%d score=%.3f → grade %s\n",
				l.Price, l.Label, p.Kind, p.TF, p.HTF, p.Fresh, p.Confluence, p.Score, p.Grade)
		} else {
			fmt.Printf("  %-10.2f %-16s NOT IN RECOMPUTED POOL\n", l.Price, l.Label)
		}
	}
	fmt.Println("\nrecomputed pool (top-24 by seating):")
	for _, p := range as.pool {
		fmt.Printf("  %-10.2f kind=%-9s label=%-16s grade=%s fresh=%-6q conf=%d score=%.3f\n",
			p.Price, p.Kind, p.Label, p.Grade, p.Fresh, p.Confluence, p.Score)
	}
	fmt.Printf("recomputed HTFZonesFull: %d rows\n", len(as.in.HTFZonesFull))
	for _, z := range as.in.HTFZonesFull {
		fmt.Printf("  %-10.2f kind=%-9s label=%-16s grade=%s tf=%q\n", z.Price, z.Kind, z.Label, z.Grade, z.TF)
	}

	// ── C. weekly-class shadow case ───────────────────────────────────────
	fmt.Println("\n=== P2-C: WEEKLY shadow band (0.25×ATR5m) on the seated doc levels ===")
	atr5m := kernel.StaleConfirmATR5m(as.bars1m)
	band := kernel.WeeklyConfluenceBandATR() * atr5m
	refs := kernel.WeeklyShadowRefs(as.bars1m, writeAt, as.price)
	f := kernel.ComputeWeeklyFacts(as.bars1m, writeAt, as.price)
	fmt.Printf("ATR5m=%.2f (StaleConfirmATR5m — 5m-bucket Wilder ATR14 from stored 1m bars) · band=±%.2f pts\n", atr5m, band)
	fmt.Printf("weekly refs: weekly_open=%.2f PWH=%.2f PWL=%.2f (RefsOK=%v) + IPDA extremes + unfilled NWOG edges → %d shadow refs\n",
		f.Refs.WeeklyOpen, f.Refs.PWH, f.Refs.PWL, f.RefsOK, len(refs))
	fmt.Println("NOTE: the plan's indicators_block carries per-TF ATR14 only — no ATR5m row is stored;")
	fmt.Println("      ATR5m is recomputed by the repo's own StaleConfirmATR5m (kernel/plan_confirm.go).")
	anyIn := false
	for _, l := range doc.Levels {
		lv := kernel.WeeklyShadowLevel{Price: l.Price, Label: l.Label, Grade: l.MachineGrade}
		if kernel.WeeklyConfluent(lv, refs, kernel.WeeklyConfluenceBandATR(), atr5m) {
			anyIn = true
			sh := float64(kernel.GradeRank(lv.Grade)) * kernel.WeeklyShadowMult()
			fmt.Printf("  🌗 SHADOW-confluent: %-12.2f %-14s grade=%s → shadow %.1f (g×%.2f) — shadow math verified\n",
				l.Price, l.Label, lv.Grade, sh, kernel.WeeklyShadowMult())
		}
	}
	if !anyIn {
		fmt.Println("  NO seated doc level within ±0.25×ATR5m of PWH/PWL/weekly_open — shadow band empty; stated explicitly.")
	}
}

func firstWords(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
