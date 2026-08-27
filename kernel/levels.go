package kernel

import (
	"time"
)

// P1 — THE MAP: shared level types.
//
// Every detector (multi-day extractor, round numbers, gaps, OR/IB, naked-POC,
// EQH/EQL, S/D zones, FVG/OB) emits DetectedLevel so the confluence scorer and
// the KEY LEVELS prompt block consume one uniform list. Detectors are pure and
// deterministic (facts=Go); nothing here calls an LLM.

// LevelKind identifies a reference level's type (drives grading + display).
type LevelKind string

const (
	KindPDH    LevelKind = "PDH"    // prior calendar-day high
	KindPDL    LevelKind = "PDL"    // prior calendar-day low
	KindPDC    LevelKind = "PDC"    // prior calendar-day close
	KindRTHH   LevelKind = "RTH-H"  // prior regular-hours (NY) high
	KindRTHL   LevelKind = "RTH-L"  // prior regular-hours (NY) low
	KindASH    LevelKind = "AS-H"   // overnight Asia high
	KindASL    LevelKind = "AS-L"   // overnight Asia low
	KindLDNH   LevelKind = "LDN-H"  // overnight London high
	KindLDNL   LevelKind = "LDN-L"  // overnight London low
	KindONH    LevelKind = "ONH"    // overnight (Asia+London) composite high
	KindONL    LevelKind = "ONL"    // overnight composite low
	KindPWH    LevelKind = "PWH"    // prior week high
	KindPWL    LevelKind = "PWL"    // prior week low
	KindPMH    LevelKind = "PMH"    // prior month high
	KindPML    LevelKind = "PML"    // prior month low
	KindRound  LevelKind = "RN"     // round number
	KindGap    LevelKind = "GAP"    // unfilled gap edge
	KindORH    LevelKind = "OR-H"   // opening-range high
	KindORL    LevelKind = "OR-L"   // opening-range low
	KindIBH    LevelKind = "IB-H"   // initial-balance high
	KindIBL    LevelKind = "IB-L"   // initial-balance low
	KindNPOC   LevelKind = "nPOC"   // naked point of control
	// Pack B (owner override 2026-08-26) — VOLUME FAMILY kinds. Forward-validated
	// via the B4 level_stats table; weights are provisional until the 2-week verdict.
	KindVWAP   LevelKind = "VWAP"   // session-anchored VWAP (+±1σ band lines share this kind)
	KindEVWAP  LevelKind = "eVWAP"  // extended VWAP (15:00 CT cash-close anchor — A2 re-anchor 2026-08-26)
	KindPOC    LevelKind = "POC"    // prior-day point of control (max-volume price)
	KindVAH    LevelKind = "VAH"    // prior-day value-area high (70%)
	KindVAL    LevelKind = "VAL"    // prior-day value-area low (70%)
	KindPDVWAP LevelKind = "pdVWAP" // prior-day session VWAP
	KindSETT   LevelKind = "SETT"   // prior settlement (16:00 CT close)
	KindMIDO   LevelKind = "MID-O"  // overnight range midpoint
	KindEQH    LevelKind = "EQH"    // equal highs (liquidity)
	KindEQL    LevelKind = "EQL"    // equal lows (liquidity)
	// Level-truth wave (2026-08-27) — SWING-POINT kinds: recent 5m/15m fractal
	// swing highs/lows (the structure engine's swing detector, same k/min-move).
	// THE 43%-missed-turns fix: intraday swing turns get seats as react_zone.
	KindSWGH LevelKind = "SWG-H" // recent swing high (5m/15m fractal extreme)
	KindSWGL LevelKind = "SWG-L" // recent swing low
	KindSupply LevelKind = "SUPPLY" // supply zone
	KindDemand LevelKind = "DEMAND" // demand zone
	KindFVG    LevelKind = "FVG"    // fair-value gap
	KindIFVG   LevelKind = "IFVG"   // inverse FVG (filled gap, inverted polarity)
	KindOB     LevelKind = "OB"     // order block
	KindOwner  LevelKind = "OWNER"  // sticky owner-set level (P3.6-C)
)

// DetectedLevel is the uniform output of every detector (a line or a zone). For
// a line, Lo == Hi == Price. HTF marks a higher-timeframe origin (grading input).
type DetectedLevel struct {
	Kind       LevelKind `json:"kind"`
	Price      float64   `json:"price"`       // line price (zones: midpoint)
	Lo         float64   `json:"lo"`          // zone bottom (== Price for a line)
	Hi         float64   `json:"hi"`          // zone top (== Price for a line)
	Label      string    `json:"label"`       // display label, e.g. "PDH", "RN 15500", "nPOC·Tue"
	OriginDate string    `json:"origin_date"` // YYYY-MM-DD of formation
	HTF        bool      `json:"htf"`         // higher-timeframe origin
	// TF is the DETECTION timeframe ("1m"…"4h"; "" = the 1m slice). Drives the
	// v3 zone evidence tiers (owner-approved 2026-08-24).
	TF string `json:"tf,omitempty"`
	// ZonePattern classifies an S/D zone as "reversal" (RBD/DBR — strongest) or
	// "continuation" (RBR/DBD — weaker). "" = unknown (older detections).
	ZonePattern string `json:"zone_pattern,omitempty"`
	Info        string `json:"info,omitempty"` // extra (gap size, fill state, ...)
	// FormedAtMs is the formation birth instant (bar open time of the candle
	// that completed the pattern), in unix ms. 0 = unknown/older detections.
	// The W6 wake loop (2026-08-25) diffs this against the plan row's birth
	// time to find events the plan never saw.
	FormedAtMs int64 `json:"formed_at_ms,omitempty"`
}

// lineLevel builds a single-price DetectedLevel (Lo==Hi==price).
func lineLevel(kind LevelKind, price float64, label, origin string, htf bool) DetectedLevel {
	return DetectedLevel{Kind: kind, Price: price, Lo: price, Hi: price, Label: label, OriginDate: origin, HTF: htf}
}

// zoneLevel builds a banded DetectedLevel (Price = midpoint of [lo,hi]).
func zoneLevel(kind LevelKind, lo, hi float64, label, origin string) DetectedLevel {
	if hi < lo {
		lo, hi = hi, lo
	}
	return DetectedLevel{Kind: kind, Price: (lo + hi) / 2, Lo: lo, Hi: hi, Label: label, OriginDate: origin}
}

// chicago is the legacy name for the canonical CT location — it delegates
// to CTLocation() (kernel/tz.go), the SINGLE timezone source every renderer
// must go through (owner rule 2026-08-19: CT is canonical everywhere).
func chicago() *time.Location { return CTLocation() }
