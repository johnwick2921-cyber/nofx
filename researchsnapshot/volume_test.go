package researchsnapshot

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

type captureSink struct {
	mu    sync.Mutex
	facts []Fact
}

func (s *captureSink) Save(_ context.Context, facts []Fact) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.facts = append(s.facts, facts...)
	return nil
}

type lineCapture struct {
	mu    sync.Mutex
	lines []string
}

func (l *lineCapture) add(s string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, s)
}

func (l *lineCapture) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]string, len(l.lines))
	copy(out, l.lines)
	return out
}

// RED 1+4: the recorder narrates per-fact at INFO (13.2M lines/day). The volume
// contract: ONE rollup line per RESEARCH_LOG_EVERY_S (default 60s), carrying
// rows-per-object, the real dropped count, and the queue depth — never a
// per-fact line.
func TestVolumeRollupNotPerFact(t *testing.T) {
	t.Setenv("RESEARCH_LOG_EVERY_S", "60")
	sink := &captureSink{}
	log := &lineCapture{}
	r := NewRecorder(sink, 128, log.add)
	defer r.Close()

	now := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	for i := 0; i < 20; i++ {
		f := NewFact("market", "test", nil, Clocks{})
		if !r.offerWithClock(clock, "market", func() []Fact { return []Fact{f} }) {
			t.Fatalf("offer %d refused", i)
		}
	}
	r.drop("queue full: market") // 3 synthetic drops for the dropped= assertion
	r.drop("queue full: market")
	r.drop("queue full: market")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := r.Flush(ctx); err != nil {
		t.Fatalf("flush: %v", err)
	}
	time.Sleep(50 * time.Millisecond) // let the rollup timer/drain settle

	lines := log.snapshot()
	perFact := 0
	rollup := ""
	for _, l := range lines {
		if strings.Contains(l, "research snapshot written:") {
			perFact++
		}
		if strings.Contains(l, "research snapshot rollup") {
			rollup = l
		}
	}
	if perFact > 0 {
		t.Fatalf("per-fact narration must be gone: %d line(s) still contain 'research snapshot written:' (first: %q)", perFact, firstContaining(lines, "research snapshot written:"))
	}
	if rollup == "" {
		t.Fatal("no rollup line emitted; want one per RESEARCH_LOG_EVERY_S with rows/objects/drops/queue")
	}
	if !strings.Contains(rollup, "rows=20") {
		t.Fatalf("rollup must carry the row count; got %q", rollup)
	}
	if !strings.Contains(rollup, "market=20") {
		t.Fatalf("rollup must carry rows per object; got %q", rollup)
	}
	if !strings.Contains(rollup, "drops=3") {
		t.Fatalf("rollup must carry the REAL dropped count (3 forced); got %q", rollup)
	}
	if !strings.Contains(rollup, "queue=0") {
		t.Fatalf("rollup must carry the queue depth; got %q", rollup)
	}
}

// RED 2: drop notices must reach the WARN-level func passed from main.go,
// rate-limited to ONE line per minute carrying the coalesced delta — not one
// INFO line per drop.
func TestVolumeDropNoticeWarnRateLimited(t *testing.T) {
	sink := &captureSink{}
	warn := &lineCapture{}
	r := NewRecorder(sink, 1, warn.add)
	defer r.Close()

	for i := 0; i < 5; i++ {
		r.drop("queue full: market")
	}
	deadline := time.Now().Add(2 * time.Second)
	var lines []string
	for time.Now().Before(deadline) {
		lines = warn.snapshot()
		if len(lines) > 0 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	// settle the coalescing window: no further lines may appear for a burst
	time.Sleep(250 * time.Millisecond)
	lines = warn.snapshot()
	dropLines := 0
	joined := ""
	for _, l := range lines {
		if strings.Contains(l, "research snapshot dropped") {
			dropLines++
			joined = l
		}
	}
	if dropLines == 0 {
		t.Fatal("no drop notice reached the warn func; want a WARN line with the coalesced delta")
	}
	if dropLines > 1 {
		t.Fatalf("drop notices must be rate-limited to one line per minute: got %d line(s); last: %q", dropLines, joined)
	}
	if !strings.Contains(joined, "total=5") {
		t.Fatalf("the single drop line must carry the coalesced total (5); got %q", joined)
	}
}

// RED 3: RESEARCH_SNAPSHOT env gate — unset or 0 means the recorder is not
// started and the boot line says OFF.
func TestVolumeResearchSnapshotEnvGate(t *testing.T) {
	t.Setenv("RESEARCH_SNAPSHOT", "0")
	log := &lineCapture{}
	closeFn := Start(t.TempDir()+"/r.db", log.add, log.add)
	defer closeFn()
	if Active() != nil {
		t.Fatal("recorder must NOT start when RESEARCH_SNAPSHOT=0; Active() is non-nil")
	}
	if !strings.Contains(CurrentBootLine(), "research snapshot: OFF (RESEARCH_SNAPSHOT unset)") {
		t.Fatalf("boot line must say OFF when the gate is closed; got %q", CurrentBootLine())
	}
}

// RED 5: RESEARCH_RETAIN_DAYS (default 7) prunes research_facts by captured_ms
// at boot and daily; VACUUM is never run automatically.
func TestVolumeRetentionPrunesAtStart(t *testing.T) {
	t.Setenv("RESEARCH_SNAPSHOT", "1")
	t.Setenv("RESEARCH_RETAIN_DAYS", "7")
	path := t.TempDir() + "/r.db"
	a, err := Open(path, "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// One row 8 days old, one row now — inserted directly so captured_ms is
	// under test control (Save stamps now).
	old := time.Now().Add(-8 * 24 * time.Hour).UnixMilli()
	if _, err := a.db.Exec(`INSERT INTO research_facts (writer_revision, schema_version, object, snapshot_id, event, captured_ms, fields_json, missing_json, null_fields) VALUES ('', 1, 'market', NULL, 'test', ?, '{}', '{}', 0), ('', 1, 'market', NULL, 'test', ?, '{}', '{}', 0)`, old, time.Now().UnixMilli()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = a.Close()

	log := &lineCapture{}
	closeFn := Start(path, log.add, log.add)
	defer closeFn()

	got := func() int {
		a2, err := Open(path, "")
		if err != nil {
			t.Fatalf("reopen: %v", err)
		}
		defer a2.Close()
		var n int
		if err := a2.db.QueryRow(`SELECT COUNT(*) FROM research_facts WHERE captured_ms < ?`, time.Now().Add(-7*24*time.Hour).UnixMilli()).Scan(&n); err != nil {
			t.Fatalf("count: %v", err)
		}
		return n
	}()
	if got != 0 {
		t.Fatalf("rows older than RESEARCH_RETAIN_DAYS must be pruned at boot; %d remain", got)
	}
}

func firstContaining(lines []string, sub string) string {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return l
		}
	}
	return "<none>"
}
