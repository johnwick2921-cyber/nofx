package researchsnapshot

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

var active atomic.Pointer[Recorder]

// Install is called before any producer starts. A failed archive leaves
// telemetry disabled; it cannot fail application startup.
func Install(r *Recorder) {
	active.Store(r)
}
func Active() *Recorder {
	return active.Load()
}
func Record(name string, build func() []Fact) (accepted bool) {
	r := Active()
	if r == nil {
		return false
	}
	defer func() {
		if recover() != nil {
			r.drop("producer admission panic: " + name)
			accepted = false
		}
	}()
	return r.Offer(name, build)
}

func Start(path string, log func(string), warn func(string)) (closeRecorder func()) {
	closeRecorder = func() {}
	emit := func(message string) {
		defer func() { _ = recover() }()
		if log != nil {
			log(message)
		}
	}
	warnEmit := func(message string) {
		defer func() { _ = recover() }()
		if warn != nil {
			warn(message)
		} else if log != nil {
			log(message)
		}
	}
	defer func() {
		if recover() != nil {
			warnEmit("WARN research snapshot initialization panic; capture unavailable")
		}
	}()
	// RESEARCH_SNAPSHOT gate (dispatch 103): unset or 0 leaves the recorder off.
	if v := os.Getenv("RESEARCH_SNAPSHOT"); v == "" || v == "0" || strings.EqualFold(v, "false") {
		setStatusNote("OFF (RESEARCH_SNAPSHOT unset)")
		emit("research snapshot: OFF (RESEARCH_SNAPSHOT unset)")
		return
	}
	rev := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				rev = s.Value
			}
		}
	}
	a, err := Open(path, rev)
	if err != nil {
		setStatusNote("OFF (archive unavailable)")
		warnEmit("WARN research snapshot archive unavailable; capture disabled")
		return
	}
	// Retention at boot (RESEARCH_RETAIN_DAYS, default 7). VACUUM is never run
	// automatically: on a ~77 GB archive it would rewrite the whole file on the
	// trading DB's disk — space reclamation is the owner's call.
	if n, err := pruneOldFacts(a, retainDays()); err == nil && n > 0 {
		emit(fmt.Sprintf("research snapshot: pruned %d row(s) older than %dd at boot (no auto-VACUUM — owner decides when to reclaim disk)", n, retainDays()))
	}
	r := NewRecorder(a, 128, warnEmit)
	r.SetInfoLog(emit)
	Install(r)
	retentionStop := make(chan struct{})
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if _, err := pruneOldFacts(a, retainDays()); err != nil {
					warnEmit(fmt.Sprintf("WARN research snapshot daily prune failed: %v", err))
				}
			case <-retentionStop:
				return
			}
		}
	}()
	return func() { close(retentionStop); Install(nil); r.Close(); _ = a.Close() }
}

func BootLineAt(a *Archive, r *Recorder, now time.Time) string {
	schema := "UNKNOWN"
	counts := map[string]string{}
	for _, k := range Objects {
		counts[k] = "UNKNOWN"
	}
	nulls := "UNKNOWN"
	dropped := "UNKNOWN"
	latency := "UNKNOWN"
	if r != nil {
		dropped = fmt.Sprint(r.Dropped())
		if n := r.LatencyP50MS(); n != nil {
			latency = fmt.Sprintf("≤%.3fms (admission)", *n)
		}
	}
	if a != nil {
		loc, err := time.LoadLocation("America/Chicago")
		if err == nil {
			ct := now.In(loc)
			from := time.Date(ct.Year(), ct.Month(), ct.Day(), 0, 0, 0, 0, loc)
			to := from.AddDate(0, 0, 1)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			var version int
			if a.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version) == nil {
				schema = fmt.Sprint(version)
			}
			rows, e := a.db.QueryContext(ctx, "SELECT object,count(*),sum(null_fields) FROM research_facts WHERE captured_ms>=? AND captured_ms<? GROUP BY object", from.UnixMilli(), to.UnixMilli())
			if e == nil {
				for _, k := range Objects {
					counts[k] = "0"
				}
				var total int64
				for rows.Next() {
					var k string
					var n, missing int64
					if rows.Scan(&k, &n, &missing) != nil {
						e = fmt.Errorf("scan failed")
						break
					}
					counts[k] = fmt.Sprint(n)
					total += missing
				}
				if rows.Err() != nil {
					e = rows.Err()
				}
				rows.Close()
				if e == nil {
					nulls = fmt.Sprint(total)
				} else {
					for _, k := range Objects {
						counts[k] = "UNKNOWN"
					}
				}
			}
		}
	}
	return fmt.Sprintf("🗄 research snapshot: schema=%s · objects=%d · rows today market=%s candidate=%s plan=%s scenario=%s exec=%s · null-fields=%s · dropped=%s · added latency p50=%s", schema, len(Objects), counts["market"], counts["candidate"], counts["plan"], counts["scenario"], counts["exec"], nulls, dropped, latency)
}
func CurrentBootLine() string {
	return CurrentBootLineAt(time.Now())
}
func CurrentBootLineAt(now time.Time) string {
	r := Active()
	if r == nil {
		return "research snapshot: " + currentStatusNote()
	}
	a, _ := r.sink.(*Archive)
	return BootLineAt(a, r, now)
}

// Contain is deferred by producer adapters as well as the worker. Even a
// custom error/string conversion fault cannot escape back into a decision.
func Contain(name string) {
	if recover() != nil {
		if r := Active(); r != nil {
			r.drop("producer panic: " + name)
		}
	}
}
