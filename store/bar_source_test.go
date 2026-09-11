package store

import "testing"

// ── BAR-SOURCE WAVE PINS (store half) ────────────────────────────────────────

func bsRow(t int64, c float64, src string) BarHistoryDB {
	return BarHistoryDB{Symbol: "MNQ", TF: "1m", OpenTimeMs: t, O: c, H: c + 1, L: c - 1, C: c, V: 1, Contract: "MNQ 12-26", Source: src}
}

func readOne(t *testing.T, bh *BarHistoryStore, ms int64) BarHistoryDB {
	t.Helper()
	var r BarHistoryDB
	if err := bh.db.Where("symbol='MNQ' AND tf='1m' AND open_time_ms=?", ms).First(&r).Error; err != nil {
		t.Fatalf("read %d: %v", ms, err)
	}
	return r
}

// THE UPSERT RULE — store half. RED on the pre-wave upsert: DO UPDATE was
// unconditional and the replay overwrote the live row.
func TestReplayNeverOverwritesALiveRow(t *testing.T) {
	bh := rollStore(t)
	const ms = int64(1789097820000)
	if err := bh.InsertBars([]BarHistoryDB{bsRow(ms, 29358.25, BarSourceLive)}); err != nil {
		t.Fatal(err)
	}
	if err := bh.InsertBars([]BarHistoryDB{bsRow(ms, 29068.25, BarSourceHistorical)}); err != nil {
		t.Fatal(err)
	}
	r := readOne(t, bh, ms)
	if r.C != 29358.25 || r.Source != BarSourceLive {
		t.Fatalf("the replay overwrote a live row: close=%.2f source=%q — this is the 186-row damage of 2026-09-10 22:39", r.C, r.Source)
	}
}

// Live overwrites historical; historical fills a gap; mixed is kept over a
// later replay (it is the evidence of the seam).
func TestSourcePrecedence(t *testing.T) {
	bh := rollStore(t)
	const a, b, c = int64(1789097820000), int64(1789097880000), int64(1789097940000)
	// historical first, then live: live wins
	_ = bh.InsertBars([]BarHistoryDB{bsRow(a, 29068, BarSourceHistorical)})
	_ = bh.InsertBars([]BarHistoryDB{bsRow(a, 29358, BarSourceLive)})
	if r := readOne(t, bh, a); r.C != 29358 || r.Source != BarSourceLive {
		t.Fatalf("live must overwrite historical: %+v", r)
	}
	// historical into an empty minute: fills
	_ = bh.InsertBars([]BarHistoryDB{bsRow(b, 29068, BarSourceHistorical)})
	if r := readOne(t, bh, b); r.Source != BarSourceHistorical {
		t.Fatalf("historical must fill a gap: %+v", r)
	}
	// mixed, then a replay: mixed stays
	_ = bh.InsertBars([]BarHistoryDB{bsRow(c, 29355, BarSourceMixed)})
	_ = bh.InsertBars([]BarHistoryDB{bsRow(c, 29068, BarSourceHistorical)})
	if r := readOne(t, bh, c); r.Source != BarSourceMixed || r.C != 29355 {
		t.Fatalf("a replay must not erase the mixed evidence: %+v", r)
	}
}

// A bar with no source (or a made-up one) is refused, not guessed.
func TestInsertRefusesUnlabelledSource(t *testing.T) {
	bh := rollStore(t)
	r := bsRow(1789097820000, 29358, "")
	if err := bh.InsertBars([]BarHistoryDB{r}); err == nil {
		t.Fatal("a bar that does not name its feed must be refused")
	}
	r.Source = "replayish"
	if err := bh.InsertBars([]BarHistoryDB{r}); err == nil {
		t.Fatal("an unknown source label must be refused")
	}
	if n, _ := bh.Count(); n != 0 {
		t.Fatalf("refused writes landed %d row(s)", n)
	}
}

// Readers never hand out a mixed bar.
func TestReadersExcludeMixed(t *testing.T) {
	bh := rollStore(t)
	const a, b, c = int64(1789097820000), int64(1789097880000), int64(1789097940000)
	_ = bh.InsertBars([]BarHistoryDB{bsRow(a, 29358, BarSourceLive), bsRow(b, 29355, BarSourceMixed), bsRow(c, 29360, BarSourceLive)})
	last, _ := bh.LastNBarsOn("MNQ", "1m", "MNQ 12-26", 10)
	between, _ := bh.BarsBetweenOn("MNQ", "1m", "MNQ 12-26", a, c+1)
	for _, rows := range [][]BarHistoryDB{last, between} {
		if len(rows) != 2 {
			t.Fatalf("want 2 readable bars, got %d", len(rows))
		}
		for _, r := range rows {
			if r.Source == BarSourceMixed {
				t.Fatalf("a reader handed out a mixed bar at %d", r.OpenTimeMs)
			}
		}
	}
}

// The migration labels what it cannot know CONSERVATIVELY (historical), and
// the roll wave's spans-roll rows as mixed. Idempotent.
func TestSourceBackfillIsConservativeAndIdempotent(t *testing.T) {
	bh := rollStore(t)
	// pre-column rows, written raw
	for _, x := range []struct {
		ms       int64
		contract string
	}{{1789092840000, "MNQ 09-26"}, {1789092900000, ContractMixed}, {1789093080000, "MNQ 12-26"}} {
		if err := bh.db.Exec(`INSERT INTO bars(symbol,tf,open_time_ms,o,h,l,c,v,convention,contract,source) VALUES ('MNQ','1m',?,1,2,0,1,1,'epoch_floor',?,'')`, x.ms, x.contract).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := bh.migrateSourceColumn(); err != nil {
		t.Fatal(err)
	}
	if r := readOne(t, bh, 1789092840000); r.Source != BarSourceHistorical {
		t.Fatalf("an unknown-feed row must be labelled historical (conservative), got %q", r.Source)
	}
	if r := readOne(t, bh, 1789092900000); r.Source != BarSourceMixed {
		t.Fatalf("a spans-roll row is mixed, got %q", r.Source)
	}
	c1, _ := bh.SourceCensus("MNQ")
	if err := bh.migrateSourceColumn(); err != nil {
		t.Fatal(err)
	}
	c2, _ := bh.SourceCensus("MNQ")
	for k, v := range c1 {
		if c2[k] != v {
			t.Fatalf("second run changed %q: %d→%d", k, v, c2[k])
		}
	}
	if c1[""] != 0 {
		t.Fatalf("%d rows left unlabelled", c1[""])
	}
}
