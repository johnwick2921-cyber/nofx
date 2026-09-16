package researchsnapshot

import (
	"fmt"
	"nofx/telemetry"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Volume control (dispatch 103, 2026-09-16): the archive keeps every fact —
// only the narration changes. Rollups carry rows/objects/drops/queue; drop
// notices are WARN-level and rate-limited to one line per minute with the
// coalesced delta.

// envDur reads a duration env knob (seconds, default fallback).
func envDur(name string, fallback time.Duration) time.Duration {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return fallback
}

func (r *Recorder) noteRows(facts []Fact) {
	r.rowsMu.Lock()
	for _, f := range facts {
		r.rows[f.Object]++
	}
	r.rowsMu.Unlock()
	telemetry.AddResearchSnapshotRows(uint64(len(facts)))
}

// coalesceDrop accumulates a drop into the minute window. The first drop of a
// window arms a short coalescing timer so a burst prints ONE WARN line with the
// full delta instead of one line per drop (the 13.2M-lines/day failure mode).
func (r *Recorder) coalesceDrop(reason string) {
	r.warnMu.Lock()
	if r.warnDelta == 0 {
		time.AfterFunc(r.dropWindow, func() { r.emitDropWarn(false) })
	}
	r.warnDelta++
	if r.warnReason == "" {
		r.warnReason = reason
	}
	r.warnMu.Unlock()
}

// emitDropWarn prints at most ONE WARN line per minute. force=true is used by
// the barrier flush and the rollup ticker so pending drops are never lost.
func (r *Recorder) emitDropWarn(force bool) {
	r.warnMu.Lock()
	defer r.warnMu.Unlock()
	if r.warnDelta == 0 {
		return
	}
	now := time.Now()
	if !force && !r.warnAt.IsZero() && now.Sub(r.warnAt) < time.Minute {
		return // rate-limited; the ticker will force it out within 60s
	}
	line := fmt.Sprintf("WARN research snapshot dropped: +%d since last minute (reason: %s); total=%d", r.warnDelta, r.warnReason, r.Dropped())
	r.warnDelta = 0
	r.warnReason = ""
	r.warnAt = now
	if r.warn != nil {
		r.warn(line)
	} else if r.info != nil {
		r.info(line)
	}
}

// emitRollup prints the volume summary: rows per object, the live dropped
// count, and the queue depth. Rate-limited to RESEARCH_LOG_EVERY_S unless
// forced (the administrative Flush barrier forces one so tests and shutdowns
// see a final number).
func (r *Recorder) emitRollup(force bool) {
	r.rowsMu.Lock()
	rows := make(map[string]uint64, len(r.rows))
	var total uint64
	for k, v := range r.rows {
		rows[k] = v
		total += v
	}
	r.rowsMu.Unlock()
	r.rollupMu.Lock()
	defer r.rollupMu.Unlock()
	if !force && !r.lastRollupAt.IsZero() && time.Since(r.lastRollupAt) < r.rollupEvery {
		return
	}
	if !force && total == 0 && r.Dropped() == 0 {
		return
	}
	r.lastRollupAt = time.Now()
	parts := make([]string, 0, len(Objects))
	keys := make([]string, 0, len(Objects))
	for _, k := range Objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, rows[k]))
	}
	line := fmt.Sprintf("🗄 research snapshot rollup: rows=%d objects={%s} drops=%d queue=%d", total, strings.Join(parts, " "), r.Dropped(), len(r.queue))
	if r.info != nil {
		r.info(line)
	} else if r.warn != nil {
		r.warn(line)
	}
}

// pruneOldFacts deletes research rows older than retainDays. VACUUM is NEVER
// run automatically: on a ~77 GB archive a VACUUM rewrites the whole file on
// the trading DB's disk — the owner decides when to reclaim space manually.
func pruneOldFacts(a *Archive, retainDays int) (int64, error) {
	cutoff := time.Now().Add(-time.Duration(retainDays) * 24 * time.Hour).UnixMilli()
	res, err := a.db.Exec(`DELETE FROM research_facts WHERE captured_ms < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func retainDays() int {
	if v := os.Getenv("RESEARCH_RETAIN_DAYS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return 7
}

var (
	statusMu   sync.Mutex
	statusNote = "OFF (RESEARCH_SNAPSHOT unset)"
)

func setStatusNote(s string) {
	statusMu.Lock()
	defer statusMu.Unlock()
	statusNote = s
}

func currentStatusNote() string {
	statusMu.Lock()
	defer statusMu.Unlock()
	return statusNote
}
