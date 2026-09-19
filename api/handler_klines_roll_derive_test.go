package api

// W-ROLL-DAY-CHART (2026-09-19) — the derive+adjust roll stitch. The fixture
// mirrors the real 09-14 shapes: prior-contract 1m rows through the seam, a
// sparse prior 15m rung ending 08:15, the new contract's first live rows, and
// the 09:35–09:58 AddOn-switch gap in the old contract's 1m.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"nofx/auth"
	"nofx/manager"
	"nofx/market"
	"nofx/store"
)

func rollMinute(d, h, min int) int64 {
	return time.Date(2026, 9, d, h, min, 0, 0, rollCT).UnixMilli()
}

// newRollDeriveServer seeds the derive fixture and returns the real Server +
// token. The ring serves the current contract's bars for the asked tf.
func newRollDeriveServer(t *testing.T, withGap bool) (*Server, string) {
	t.Helper()
	orig := market.FuturesBarsProvider
	t.Cleanup(func() { market.FuturesBarsProvider = orig })

	st, err := store.New(filepath.Join(t.TempDir(), "roll-derive.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.BarHistory().Migrate(); err != nil {
		t.Fatal(err)
	}

	var rows []store.BarHistoryDB
	add := func(contract, tf string, at int64, price float64, src string) {
		rows = append(rows, store.BarHistoryDB{
			Symbol: "MNQ", Contract: contract, TF: tf, OpenTimeMs: at,
			O: price, H: price + 1, L: price - 1, C: price, V: 1, Source: src,
		})
	}
	// Prior contract 1m: 08:00–10:33, climbing from 30000.
	p := 30000.0
	for tm := rollMinute(14, 8, 0); tm <= rollMinute(14, 10, 33); tm += 60_000 {
		h, m := time.UnixMilli(tm).In(rollCT).Hour(), time.UnixMilli(tm).In(rollCT).Minute()
		if withGap && h == 9 && m >= 35 && m <= 58 {
			continue
		}
		add("MNQ 09-26", "1m", tm, p, store.BarSourceLive)
		p += 0.25
	}
	// Prior sparse 15m rung ending 08:15 (the legacy stitch's stored rung).
	for tm := rollMinute(14, 7, 0); tm <= rollMinute(14, 8, 15); tm += 15 * 60_000 {
		add("MNQ 09-26", "15m", tm, 29990, store.BarSourceLive)
	}
	// Current contract: first live 1m at 10:34, +290 basis.
	firstNew := rollMinute(14, 10, 34)
	pNew := p + 290
	for tm := firstNew; tm <= rollMinute(14, 11, 5); tm += 60_000 {
		add("MNQ 12-26", "1m", tm, pNew, store.BarSourceLive)
		pNew += 0.25
	}
	// Current sparse 15m from 08:30.
	for tm := rollMinute(14, 8, 30); tm <= rollMinute(14, 11, 0); tm += 15 * 60_000 {
		add("MNQ 12-26", "15m", tm, 30290, store.BarSourceLive)
	}
	// Current sparse 5m from 10:00 (the real 5m seam hour).
	for tm := rollMinute(14, 10, 0); tm <= rollMinute(14, 11, 0); tm += 5 * 60_000 {
		add("MNQ 12-26", "5m", tm, 30290, store.BarSourceLive)
	}
	if err := st.BarHistory().InsertBars(rows); err != nil {
		t.Fatalf("insert fixture: %v", err)
	}
	market.FuturesBarsProvider = func(symbol, tf string, count int) []market.Kline {
		if symbol != "MNQ" {
			return nil
		}
		var out []market.Kline
		for _, r := range rows {
			if r.TF == tf && r.Contract == "MNQ 12-26" && r.Source == store.BarSourceLive {
				out = append(out, market.Kline{OpenTime: r.OpenTimeMs, Open: r.O, High: r.H, Low: r.L, Close: r.C, Volume: r.V})
			}
		}
		if count > 0 && len(out) > count {
			out = out[len(out)-count:]
		}
		return out
	}

	auth.SetJWTSecret("roll-derive-test-secret")
	tok, err := auth.GenerateJWT("u-roll", "roll@test")
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	return NewServer(manager.NewTraderManager(), st, nil, "127.0.0.1", 0), tok
}

func rollDeriveGet(t *testing.T, s *Server, tok, interval string, limit int) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/klines?symbol=MNQ&interval=%s&limit=%d&exchange=ninjatrader", interval, limit), nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	s.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /klines: %d %s", rec.Code, rec.Body.String())
	}
	return rec.Body.Bytes()
}

type rollEnvelope struct {
	Klines []map[string]any `json:"klines"`
	Roll   *struct {
		Derived  bool     `json:"derived"`
		Adjusted bool     `json:"adjusted"`
		Basis    *float64 `json:"basis"`
		Reason   string   `json:"reason"`
	} `json:"roll"`
}

func mustEnvelope(t *testing.T, body []byte) rollEnvelope {
	t.Helper()
	var env rollEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("expected the {klines, roll} envelope: %v", err)
	}
	return env
}

// (1) the derived 15m prior side runs up to the 08:15 bucket and the new
// contract starts at 08:30 — NO hole at the seam.
func TestRollDeriveClosesTheSeamHole(t *testing.T) {
	s, tok := newRollDeriveServer(t, true)
	env := mustEnvelope(t, rollDeriveGet(t, s, tok, "15m", 500))
	if env.Roll == nil || !env.Roll.Derived {
		t.Fatalf("roll envelope must say derived: %+v", env.Roll)
	}
	var lastPrior, firstNew int64
	for _, k := range env.Klines {
		if k["contract"] == "MNQ 09-26" {
			if o, ok := k["openTime"].(float64); ok && int64(o) > lastPrior {
				lastPrior = int64(o)
			}
		}
		if k["contract"] == "MNQ 12-26" && firstNew == 0 {
			firstNew = int64(k["openTime"].(float64))
		}
	}
	if lastPrior != rollMinute(14, 8, 15) {
		t.Errorf("last prior 15m bucket = %s, want 08:15", time.UnixMilli(lastPrior).In(rollCT).Format("15:04"))
	}
	if firstNew != rollMinute(14, 8, 30) {
		t.Errorf("first current 15m = %s, want 08:30", time.UnixMilli(firstNew).In(rollCT).Format("15:04"))
	}
	if firstNew-lastPrior != 15*60_000 {
		t.Errorf("seam hole: %s → %s", time.UnixMilli(lastPrior).In(rollCT).Format("15:04"), time.UnixMilli(firstNew).In(rollCT).Format("15:04"))
	}
}

// (2) basis measured and applied to every prior bar; current untouched.
func TestRollDeriveAppliesBasisAndLeavesCurrentAlone(t *testing.T) {
	s, tok := newRollDeriveServer(t, false)
	env := mustEnvelope(t, rollDeriveGet(t, s, tok, "5m", 500))
	if env.Roll == nil || !env.Roll.Adjusted || env.Roll.Basis == nil {
		t.Fatalf("roll must be adjusted with basis: %+v", env.Roll)
	}
	basis := *env.Roll.Basis
	if basis < 289 || basis > 291 {
		t.Errorf("basis = %.2f, want ~290", basis)
	}
	for _, k := range env.Klines {
		if k["contract"] == "MNQ 12-26" {
			if k["adjusted"] != nil {
				t.Errorf("current-contract bar was adjusted: %+v", k)
			}
			continue
		}
		if o, ok := k["openTime"].(float64); ok && int64(o) == rollMinute(14, 8, 0) {
			got := k["close"].(float64)
			want := 30001.00 + basis // 08:00–08:04 five 1m closes end at 30001.00
			if got < want-0.001 || got > want+0.001 {
				t.Errorf("derived 08:00 close = %.3f, want %.3f (raw + basis)", got, want)
			}
		}
	}
}

// (3) the 09:35–09:58 gap stays a gap in the derived series.
func TestRollDerivePreservesTheGap(t *testing.T) {
	s, tok := newRollDeriveServer(t, true)
	env := mustEnvelope(t, rollDeriveGet(t, s, tok, "5m", 500))
	for _, k := range env.Klines {
		if k["contract"] != "MNQ 09-26" {
			continue
		}
		o := int64(k["openTime"].(float64))
		h, m := time.UnixMilli(o).In(rollCT).Hour(), time.UnixMilli(o).In(rollCT).Minute()
		// The skipped minutes are 09:35–09:58; buckets 09:35..09:50 must be
		// ABSENT. The 09:55 bucket legitimately holds the 09:59 1m row.
		if h == 9 && m >= 35 && m <= 50 {
			t.Errorf("derived bar invented in the 09:35–09:58 gap at %02d:%02d", h, m)
		}
	}
}

// (4) CHART_ROLL_STITCH=legacy serves the stored-aggregate stitch byte-
// identically to today: bare array, no envelope, no derive/adjust flags.
func TestRollDeriveLegacyKnobByteIdentical(t *testing.T) {
	old := chartRollStitchDerive
	chartRollStitchDerive = false
	t.Cleanup(func() { chartRollStitchDerive = old })

	s, tok := newRollDeriveServer(t, false)
	body := rollDeriveGet(t, s, tok, "15m", 500)
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		t.Fatalf("legacy knob must serve the bare array: %v", err)
	}
	sawPrior := false
	for _, k := range arr {
		if k["derived"] != nil || k["adjusted"] != nil {
			t.Fatalf("legacy kline carries derive/adjust flags: %+v", k)
		}
		if k["contract"] == "MNQ 09-26" {
			sawPrior = true
		}
	}
	if !sawPrior {
		t.Fatalf("legacy prior segment missing")
	}
}
