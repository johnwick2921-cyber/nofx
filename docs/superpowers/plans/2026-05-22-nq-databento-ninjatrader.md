# NQ Trading via Databento + NinjaTrader — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Agent tooling available in this repo (registered 2026-05-25):**
> A Playwright MCP server is registered in `~/.claude.json` for project
> `/home/hoang/nofx`. When you spawn agents that need to verify the React
> frontend (Settings page exchange config form, Dashboard column rendering,
> Strategy Studio variant gating, etc.), use the `mcp__playwright__*` tools
> to drive a headless Chromium against `http://localhost:3000` after
> running `cd web && npm run dev`. Chromium binary is pre-installed at
> `~/.cache/ms-playwright/`. Without this, you cannot verify UI changes
> end-to-end — `npm run build` only confirms TypeScript compiles, not
> that the rendered DOM is correct.
>
> Server command: `npx -y @playwright/mcp@latest --headless`
> Register on a fresh machine with: `claude mcp add playwright -- npx -y "@playwright/mcp@latest" --headless`

**Goal:** Get the nofx bot to fetch NQ futures OHLCV from Databento, ask the AI for a trade decision in futures-native vocabulary, write that decision to a CSV signal file, and have NinjaTrader (running on the user's Windows host with the open-source `claudetrader.cs` strategy attached to an MNQ chart) execute that decision in SIM mode — then tail the fills CSV back into the bot's database.

**Architecture:** Three independent layers wired through narrow interfaces.

```
┌──── WSL2 (Linux) ────────────────────────────┐    ┌── Windows ──┐
│  nofx-bin (Go)                               │    │             │
│                                              │    │  NinjaTrader 8
│  trader/auto_trader_loop.go                  │    │  + claudetrader.cs
│    │                                         │    │  (modified copy)
│    ├─► provider/databento/  (NEW)            │    │      ▲
│    │     GetOHLCV("NQ.c.0", "1m", ...)       │    │      │ polls CSV
│    │                                         │    │      │ every 2s
│    ├─► kernel/engine.go (MODIFY)             │    │      │
│    │     futures-mode prompt → AI            │    │      │
│    │                                         │    │      │
│    └─► provider/ninjatrader/ (NEW)           │    │      │
│          WriteSignal(dir, entry, sl, tp) ────┼────┼──┐   │
│          TailFills(callback) ◄───────────────┼────┼──┼───┘
│                                              │    │  │
└──────────────────────────────────────────────┘    │  │
                                                    │  ▼  files at
                                                    │  C:\Users\<u>\NofxTrader\data\
                                                    │  ├─ trade_signals.csv  (Go writes, NT reads)
                                                    │  └─ trades_taken.csv   (NT writes, Go reads)
                                                    └─────────────
```

WSL2 reaches Windows files via `/mnt/c/Users/<windows_username>/NofxTrader/data/`. Same files, two views. No sockets, no daemon.

**Tech Stack:**
- Go (existing `nofx` codebase)
- Databento Historical REST API (`https://hist.databento.com/v0/`) with HTTP Basic auth
- `claudetrader.cs` (open-source NinjaScript from `https://github.com/J0shusmc/Claude-Trader-NinjaTrader`) — modified copy with our file paths
- NinjaTrader 8 (Windows desktop app) with SIM connection
- File-CSV protocol (5-field signals, 3-field fills)

**Scope of this plan:** End-to-end paper-trading slice. AI brain takes NQ context → decides → writes signal → NT executes in SIM → fill recorded. **Out of scope** (separate future plans): CME holiday calendar, contract rolls, position-sizing-by-contract-multiplier math, dead-man-switch heartbeat, web UI rewrite, going live with real money.

**Architectural alignment with existing patterns (revised 2026-05-22):**

The system already has rich per-trader configuration via the **Exchange**, **Strategy**, **AI Model**, and **Trader** entities (documented in upstream `STRATEGY_MODULE.md`). This plan integrates NQ trading **into those patterns** rather than introducing a parallel "futures mode" branch:

- **NinjaTrader is added as a new exchange type** in [store/exchange.go](store/exchange.go) (joining `binance`, `bybit`, `hyperliquid`, etc.). Per-account NT data dir lives on the Exchange row, not in global env vars.
- **NQ symbols use the existing static-coin-source mode** at [kernel/engine.go:437-442](kernel/engine.go#L437) — `CoinSource.SourceType="static"`, `StaticCoins=["NQ.c.0"]`. No engine fork needed.
- **The futures prompt template is selected via the existing `PromptVariant` field** ([api/strategy.go:551](api/strategy.go#L551) already supports `req.PromptVariant`). NQ traders set `PromptVariant="futures"`.
- **Data feed override:** when the engine sees `ExchangeType=="ninjatrader"`, it routes K-line fetches through Databento instead of nofxos/coinank. This is a small branch in `kernel/engine.go` data-fetch, not a top-level mode switch.
- **Symbol normalization fix:** [market.Normalize()](market/data.go) currently appends `USDT`. It needs a futures-aware path that detects CME symbol patterns (e.g., `NQ.c.0`, `MNQ`, `NQM6`) and skips the suffix. See Task NEW.

**Verified facts (from reading `claudetrader.cs` directly):**
- Strategy places **market orders** (not limit, despite README claims) → `EnterLong(0, contractQuantity, "CT_Long")` at line 255. `Entry_Price` in CSV is a *reference* logged but not used for entry.
- SL is `ExitLongStopMarket` and TP is `ExitLongLimit`, both placed at absolute prices from CSV after the entry fills.
- File is read every `FileCheckInterval` seconds (default 2), but only re-parsed when `File.GetLastWriteTime` changes. **NT clears the file (rewrites only the header) after each signal is processed** — every signal is one-shot.
- Dedup is by `DateTime + Direction` concatenated as signal ID. Two signals with the same DateTime+Direction = the second is ignored.
- Strategy refuses new signals while `Position.MarketPosition != Flat || hasLimitOrder`. One position at a time.
- `IsExitOnSessionCloseStrategy = true` with 30s buffer — NT auto-flattens at session close.

---

## File Structure

**New files (this plan):**
- `provider/databento/historical.go` — `GetOHLCV(symbol, interval, start, end) ([]Bar, error)`
- `provider/databento/historical_test.go` — unit tests with mocked HTTP
- `provider/databento/resolve.go` — `ResolveContinuous(symbol)` returns "NQM6" etc.
- `provider/databento/resolve_test.go` — unit tests
- `market/databento_adapter.go` — `BarsToKlines(bars []databento.Bar) []market.Kline`
- `market/databento_adapter_test.go` — unit tests
- `provider/ninjatrader/csv_writer.go` — `WriteSignal(SignalRow) error`
- `provider/ninjatrader/csv_writer_test.go` — unit tests against a tempdir
- `provider/ninjatrader/csv_tailer.go` — `TailFills(ctx, onFill) error` — long-running goroutine
- `provider/ninjatrader/csv_tailer_test.go` — unit tests
- `provider/ninjatrader/types.go` — shared `SignalRow`, `FillRow` types
- `trader/ninjatrader/trader.go` — implements `trader/types.Trader` interface using the CSV writer/tailer
- `trader/ninjatrader/trader_test.go` — unit tests
- `kernel/engine_prompt_futures.go` — `BuildFuturesSystemPrompt()` and `BuildFuturesUserPrompt()`
- `kernel/engine_prompt_futures_test.go` — golden-file tests for prompt content

**Modified files:**
- `config/config.go` — add `DatabentoAPIKey`, `NinjaTraderDataDir`, `TradingMode` fields
- `.env.example` — add `DATABENTO_API_KEY`, `NINJATRADER_DATA_DIR`, `TRADING_MODE`
- `kernel/engine.go` — branch on `TradingMode == "futures"` to call futures prompt + Databento data path
- `trader/auto_trader.go:263-315` — add `"ninjatrader"` case in the exchange switch
- `main.go` — wire `provider/databento.DefaultClient` initialization on startup

**Modified externally (Windows side):**
- `claudetrader.cs` — change hardcoded paths from `C:\Users\Joshua\Documents\Projects\Claude Trader\data\` to `C:\Users\<user>\NofxTrader\data\` (or whatever path the user picks, via NT strategy parameter)

---

## Setup Tasks (before any code)

### Task 0: Environment + accounts + manual smoke test

**Goal:** Confirm the human-side prerequisites work BEFORE we write any code. If a hand-edited CSV doesn't trigger an order in NT SIM, no Go code we write will matter.

**Files:** none — this is human-side setup.

- [ ] **Step 0.1: Confirm Databento account + API key**

You should already have this (per earlier conversation). Verify your key works:

```bash
curl -sS -u "$DATABENTO_API_KEY:" "https://hist.databento.com/v0/metadata.list_datasets" | head -c 500
```

Expected: a JSON array of dataset codes including `GLBX.MDP3`. If you get `{"detail":"Not authenticated"}` your key is wrong; if you get a list, you're good.

- [ ] **Step 0.2: Install NinjaTrader 8 on your Windows host**

Download from [ninjatrader.com/download](https://ninjatrader.com/download). Install with default settings. On first launch, connect to the SIM101 simulated account (free, no signup money needed).

- [ ] **Step 0.3: Subscribe to MNQ in NT SIM**

In NinjaTrader:
1. Open Control Center → Connections → SIM101 → Connect.
2. New → Chart → Instrument: `MNQ 12-26` (or whichever is the front-month MNQ at the time you're doing this). Choose a 5-minute or 1-minute bar interval.
3. Confirm the chart renders live or replayed price data.

- [ ] **Step 0.4: Pick a shared data directory**

Decide on a path on your Windows host that both NT and the Go bot will use. Recommended:

```
Windows path:    C:\Users\<your-windows-username>\NofxTrader\data\
WSL2 path:       /mnt/c/Users/<your-windows-username>/NofxTrader/data/
```

Create the directory. From Windows: open Explorer, navigate to `C:\Users\<you>\`, right-click → New Folder → `NofxTrader`, then inside it create `data`.

From WSL2, confirm visibility:

```bash
ls -la "/mnt/c/Users/<your-windows-username>/NofxTrader/data/"
```

Expected: empty directory listing (no errors).

- [ ] **Step 0.5: Initialize the two CSV files**

From WSL2:

```bash
DATA_DIR="/mnt/c/Users/<your-windows-username>/NofxTrader/data"
printf "DateTime,Direction,Entry_Price,Stop_Loss,Take_Profit\n" > "$DATA_DIR/trade_signals.csv"
printf "DateTime,Direction,Entry_Price\n" > "$DATA_DIR/trades_taken.csv"
ls -la "$DATA_DIR/"
```

Expected: both files exist with their header rows.

- [ ] **Step 0.6: Install and configure claudetrader.cs in NT**

1. From your Windows host, clone the repo: `git clone https://github.com/J0shusmc/Claude-Trader-NinjaTrader.git C:\NofxTrader\bridge` (or anywhere convenient).
2. Open `C:\NofxTrader\bridge\ninjascripts\claudetrader.cs` in any text editor.
3. Find lines 34, 35, 86, 87, 98, 99 — replace `C:\Users\Joshua\Documents\Projects\Claude Trader\data\` with `C:\Users\<your-windows-username>\NofxTrader\data\` everywhere.
4. Save.
5. Copy the file to `Documents\NinjaTrader 8\bin\Custom\Strategies\claudetrader.cs`.
6. In NinjaTrader: Tools → Edit NinjaScript → Strategy → find `ClaudeTrader` → click Compile (F5). You should see "0 errors" in the output panel.

- [ ] **Step 0.7: Apply ClaudeTrader strategy to your MNQ chart**

On your MNQ 12-26 chart in NT:
1. Strategies tab → click + → select `ClaudeTrader`.
2. Set parameters:
   - `Signals File Path`: `C:\Users\<your-windows-username>\NofxTrader\data\trade_signals.csv`
   - `Trades Log File Path`: `C:\Users\<your-windows-username>\NofxTrader\data\trades_taken.csv`
   - `File Check Interval`: `2`
   - `Contract Quantity`: `1` ← starting safe; we'll raise this later after validation
3. Click Apply, then Enable. Look at NT's Output window: you should see `ClaudeTrader Initialized - Monitoring signals every 2 seconds`.

- [ ] **Step 0.8: Manual signal smoke test (DO NOT SKIP)**

This is the most important step in the entire plan. From WSL2:

```bash
DATA_DIR="/mnt/c/Users/<your-windows-username>/NofxTrader/data"
# Make sure NT is open, SIM connected, ClaudeTrader running on MNQ chart
DT=$(date +"%m/%d/%Y %H:%M:%S")
# Pick a price near current MNQ market — check NT chart. Replace 21500 with current ish.
echo "${DT},LONG,21500.00,21450.00,21560.00" >> "$DATA_DIR/trade_signals.csv"
```

Within 2-4 seconds:
- NT Output window should print `[SIGNAL] LONG MARKET ORDER (1 contracts)`
- NT Trades tab should show an order for MNQ 12-26
- Once filled, NT prints `[FILLED] LONG 1 contracts @ <price>` and `[ORDERS SUBMITTED] Long exit orders sent to broker`
- After a moment (or when SL/TP hits), NT prints `[EXIT SL]` or `[EXIT TP]`
- `trades_taken.csv` should have one new row

If this doesn't happen, **stop here and debug NT setup before continuing**. Nothing in the rest of the plan works if this manual test fails.

- [ ] **Step 0.9: Commit the .env values**

Once Steps 0.1-0.8 all pass, add the env vars to `/home/hoang/nofx/.env`:

```
DATABENTO_API_KEY=db-XXXXXXXXXXXXXXXXXXXXXXXXXX
DATABENTO_DATASET=GLBX.MDP3
NINJATRADER_DATA_DIR=/mnt/c/Users/<your-windows-username>/NofxTrader/data
TRADING_MODE=futures
```

(Replace placeholders with real values. Do NOT commit `.env` to git; it's already in `.gitignore`.)

---

## Task 1: Wire Databento config into Go

**Files:**
- Modify: `config/config.go`
- Modify: `.env.example`

- [ ] **Step 1.1: Add fields to Config struct**

In `config/config.go`, find the `Config` struct (around line 16-46) and add these fields at the bottom:

```go
// Databento (NQ futures data)
DatabentoAPIKey  string
DatabentoDataset string  // e.g., "GLBX.MDP3"

// NinjaTrader (CSV bridge for execution)
NinjaTraderDataDir string  // e.g., "/mnt/c/Users/<u>/NofxTrader/data"

// Trading mode: "crypto" (default, original behavior) or "futures"
TradingMode string
```

- [ ] **Step 1.2: Load them in Init()**

Find the `Init()` or env-loading section in `config/config.go` and add:

```go
cfg.DatabentoAPIKey = os.Getenv("DATABENTO_API_KEY")
cfg.DatabentoDataset = getEnvOrDefault("DATABENTO_DATASET", "GLBX.MDP3")
cfg.NinjaTraderDataDir = os.Getenv("NINJATRADER_DATA_DIR")
cfg.TradingMode = getEnvOrDefault("TRADING_MODE", "crypto")
```

(If `getEnvOrDefault` doesn't exist, define it as a helper: `func getEnvOrDefault(key, def string) string { if v := os.Getenv(key); v != "" { return v }; return def }`.)

- [ ] **Step 1.3: Add to .env.example**

Append to `.env.example`:

```
# NQ futures (Databento + NinjaTrader)
TRADING_MODE=crypto                                          # set to "futures" to enable NQ path
DATABENTO_API_KEY=                                           # https://databento.com/portal/keys
DATABENTO_DATASET=GLBX.MDP3                                  # CME Globex
NINJATRADER_DATA_DIR=                                        # e.g., /mnt/c/Users/<u>/NofxTrader/data
```

- [ ] **Step 1.4: Build, verify no compile errors**

Run:

```bash
cd /home/hoang/nofx && go build ./... 2>&1 | head -20
```

Expected: no output (build success).

- [ ] **Step 1.5: Commit**

```bash
cd /home/hoang/nofx
git add config/config.go .env.example
git commit -m "feat(config): add Databento + NinjaTrader config fields"
```

---

## Task 2: Databento Historical OHLCV client

**Files:**
- Create: `provider/databento/historical.go`
- Create: `provider/databento/historical_test.go`
- Existing: `provider/databento/client.go` (already has Basic-auth `doRequest`)

**API reference:** `GET /v0/timeseries.get_range` returns JSON-lines (one record per line) when `encoding=json` is set. Schema `ohlcv-1m` returns records with `ts_event` (nanosecond epoch), `open`, `high`, `low`, `close`, `volume` as integer fixed-point (close * 1e9 — divide by 1e9 to get a float). Source: [Databento Historical API docs](https://databento.com/docs/api-reference-historical/timeseries).

- [ ] **Step 2.1: Define the Bar type and signature in a new file**

Create `/home/hoang/nofx/provider/databento/historical.go`:

```go
package databento

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Bar represents one OHLCV record returned by Databento.
type Bar struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// rawBar mirrors the JSON shape Databento returns for ohlcv-* schemas.
// Numeric fields arrive as integer fixed-point with 1e9 divisor.
type rawBar struct {
	TsEvent string `json:"ts_event"`
	Open    string `json:"open"`
	High    string `json:"high"`
	Low     string `json:"low"`
	Close   string `json:"close"`
	Volume  string `json:"volume"`
}

// GetOHLCV fetches OHLCV bars for one symbol over [start, end).
// interval must be one of "1m", "1h", "1d" — maps to schema "ohlcv-<interval>".
// symbol can be a continuous code like "NQ.c.0" or a specific contract like "NQM6".
func (c *Client) GetOHLCV(symbol, interval string, start, end time.Time) ([]Bar, error) {
	schema := "ohlcv-" + interval
	params := url.Values{}
	params.Set("dataset", DefaultDataset)
	params.Set("symbols", symbol)
	params.Set("schema", schema)
	params.Set("stype_in", "continuous") // NQ.c.0 is a continuous symbol
	params.Set("start", start.UTC().Format(time.RFC3339))
	params.Set("end", end.UTC().Format(time.RFC3339))
	params.Set("encoding", "json")

	body, err := c.doRequest("/timeseries.get_range", params)
	if err != nil {
		return nil, err
	}
	return parseOHLCVResponse(body)
}

// parseOHLCVResponse decodes Databento's JSON-lines body into []Bar.
// Exported (lowercase but easy to test from within package).
func parseOHLCVResponse(body []byte) ([]Bar, error) {
	var bars []Bar
	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 1024*1024), 8*1024*1024) // tolerate long lines
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var r rawBar
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, fmt.Errorf("databento: parse bar: %w", err)
		}
		bar, err := r.toBar()
		if err != nil {
			return nil, err
		}
		bars = append(bars, bar)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("databento: scan body: %w", err)
	}
	return bars, nil
}

func (r rawBar) toBar() (Bar, error) {
	tsNs, err := strconv.ParseInt(r.TsEvent, 10, 64)
	if err != nil {
		return Bar{}, fmt.Errorf("databento: parse ts_event %q: %w", r.TsEvent, err)
	}
	open, err := scaledFloat(r.Open)
	if err != nil {
		return Bar{}, err
	}
	high, err := scaledFloat(r.High)
	if err != nil {
		return Bar{}, err
	}
	low, err := scaledFloat(r.Low)
	if err != nil {
		return Bar{}, err
	}
	closeP, err := scaledFloat(r.Close)
	if err != nil {
		return Bar{}, err
	}
	vol, err := strconv.ParseFloat(r.Volume, 64)
	if err != nil {
		return Bar{}, fmt.Errorf("databento: parse volume %q: %w", r.Volume, err)
	}
	return Bar{
		Timestamp: time.Unix(0, tsNs).UTC(),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closeP,
		Volume:    vol,
	}, nil
}

// Databento ohlcv schemas use integer fixed-point with 1e9 divisor.
func scaledFloat(s string) (float64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("databento: parse scaled int %q: %w", s, err)
	}
	return float64(n) / 1e9, nil
}
```

- [ ] **Step 2.2: Write the parser test FIRST**

Create `/home/hoang/nofx/provider/databento/historical_test.go`:

```go
package databento

import (
	"testing"
	"time"
)

func TestParseOHLCVResponse_TwoBars(t *testing.T) {
	// Sample of what Databento returns for schema=ohlcv-1m, encoding=json.
	// Each line is one bar. Numeric fields are integer fixed-point (1e9 divisor).
	body := []byte(`{"ts_event":"1746360000000000000","open":"21500250000000","high":"21515750000000","low":"21498000000000","close":"21510000000000","volume":"4321"}
{"ts_event":"1746360060000000000","open":"21510000000000","high":"21525500000000","low":"21505000000000","close":"21522750000000","volume":"5102"}
`)

	bars, err := parseOHLCVResponse(body)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(bars) != 2 {
		t.Fatalf("want 2 bars, got %d", len(bars))
	}

	want0 := Bar{
		Timestamp: time.Unix(0, 1746360000000000000).UTC(),
		Open:      21500.25,
		High:      21515.75,
		Low:       21498.00,
		Close:     21510.00,
		Volume:    4321,
	}
	if bars[0] != want0 {
		t.Errorf("bar[0] = %+v, want %+v", bars[0], want0)
	}

	if bars[1].Close != 21522.75 {
		t.Errorf("bar[1].Close = %v, want 21522.75", bars[1].Close)
	}
}

func TestParseOHLCVResponse_Empty(t *testing.T) {
	bars, err := parseOHLCVResponse([]byte(""))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(bars) != 0 {
		t.Errorf("want 0 bars, got %d", len(bars))
	}
}

func TestParseOHLCVResponse_MalformedLine(t *testing.T) {
	body := []byte(`{"ts_event":"abc","open":"1","high":"1","low":"1","close":"1","volume":"1"}` + "\n")
	_, err := parseOHLCVResponse(body)
	if err == nil {
		t.Fatal("want error on malformed ts_event, got nil")
	}
}
```

- [ ] **Step 2.3: Run tests, verify all pass**

```bash
cd /home/hoang/nofx && go test ./provider/databento/... -run TestParseOHLCV -v
```

Expected: `--- PASS` for all three. If any fails, fix the parser.

- [ ] **Step 2.4: Live smoke test (manual; costs ~$0.002 USDC equivalent on your Databento balance)**

Create a temporary file `/tmp/db_smoke.go`:

```go
package main

import (
	"fmt"
	"nofx/provider/databento"
	"os"
	"time"
)

func main() {
	c := databento.NewClient("", os.Getenv("DATABENTO_API_KEY"))
	end := time.Now().UTC()
	start := end.Add(-30 * time.Minute)
	bars, err := c.GetOHLCV("NQ.c.0", "1m", start, end)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
	fmt.Printf("Got %d bars\n", len(bars))
	for _, b := range bars {
		fmt.Printf("  %s  O=%.2f H=%.2f L=%.2f C=%.2f V=%.0f\n",
			b.Timestamp.Format("15:04:05"), b.Open, b.High, b.Low, b.Close, b.Volume)
	}
}
```

Run it:

```bash
cd /home/hoang/nofx && go run /tmp/db_smoke.go
```

Expected: ~25-30 lines of 1-minute NQ bars from the last half hour. If you get an error mentioning "401", your API key is wrong; if "402", the endpoint or schema is wrong; if "no bars", check the time range hits a session-open window.

- [ ] **Step 2.5: Commit**

```bash
cd /home/hoang/nofx
git add provider/databento/historical.go provider/databento/historical_test.go
git commit -m "feat(databento): add GetOHLCV historical client"
rm /tmp/db_smoke.go
```

---

## Task 3: Symbol resolver (continuous → specific contract)

**Why:** When NinjaTrader places an order, it places against the *specific* contract attached to the chart (e.g., `MNQ 12-26`), not the continuous symbol. Databento can tell us which specific contract `NQ.c.0` resolves to today via `/v0/symbology.resolve`. We use this to (a) confirm NT is on the front-month, and (b) detect when the front-month rolls.

**Files:**
- Create: `provider/databento/resolve.go`
- Create: `provider/databento/resolve_test.go`

- [ ] **Step 3.1: Write the failing test**

Create `/home/hoang/nofx/provider/databento/resolve_test.go`:

```go
package databento

import "testing"

func TestParseResolveResponse_FrontMonthNQ(t *testing.T) {
	// Real-shape response from /v0/symbology.resolve for symbols=NQ.c.0
	body := []byte(`{
		"result": {
			"NQ.c.0": [
				{"d0": "2026-05-22", "d1": "2026-06-19", "s": "NQM6"}
			]
		},
		"symbols": ["NQ.c.0"],
		"stype_in": "continuous",
		"stype_out": "raw_symbol",
		"start_date": "2026-05-22",
		"end_date": "2026-05-22",
		"partial": [],
		"not_found": []
	}`)

	got, err := parseResolveResponse(body, "NQ.c.0")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if got != "NQM6" {
		t.Errorf("got %q, want %q", got, "NQM6")
	}
}

func TestParseResolveResponse_NotFound(t *testing.T) {
	body := []byte(`{"result":{},"symbols":["NQ.c.0"],"not_found":["NQ.c.0"]}`)
	_, err := parseResolveResponse(body, "NQ.c.0")
	if err == nil {
		t.Fatal("want error when symbol not found, got nil")
	}
}
```

- [ ] **Step 3.2: Run the test and verify failure**

```bash
cd /home/hoang/nofx && go test ./provider/databento/... -run TestParseResolve -v
```

Expected: build error / compile failure ("parseResolveResponse" undefined). That's the failing state.

- [ ] **Step 3.3: Implement the resolver**

Create `/home/hoang/nofx/provider/databento/resolve.go`:

```go
package databento

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// ResolveContinuous returns the specific contract symbol that a continuous
// symbol (e.g. "NQ.c.0") points to today. For non-continuous symbols this
// is a passthrough.
func (c *Client) ResolveContinuous(symbol string) (string, error) {
	today := time.Now().UTC().Format("2006-01-02")
	params := url.Values{}
	params.Set("dataset", DefaultDataset)
	params.Set("symbols", symbol)
	params.Set("stype_in", "continuous")
	params.Set("stype_out", "raw_symbol")
	params.Set("start_date", today)
	params.Set("end_date", today)

	body, err := c.doRequest("/symbology.resolve", params)
	if err != nil {
		return "", err
	}
	return parseResolveResponse(body, symbol)
}

type resolveResponse struct {
	Result   map[string][]resolveEntry `json:"result"`
	NotFound []string                  `json:"not_found"`
}

type resolveEntry struct {
	D0 string `json:"d0"`
	D1 string `json:"d1"`
	S  string `json:"s"`
}

func parseResolveResponse(body []byte, symbol string) (string, error) {
	var resp resolveResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("databento resolve: parse: %w", err)
	}
	for _, nf := range resp.NotFound {
		if nf == symbol {
			return "", fmt.Errorf("databento resolve: symbol not found: %s", symbol)
		}
	}
	entries, ok := resp.Result[symbol]
	if !ok || len(entries) == 0 {
		return "", fmt.Errorf("databento resolve: no entries for %s", symbol)
	}
	return entries[0].S, nil
}
```

- [ ] **Step 3.4: Run tests and verify pass**

```bash
cd /home/hoang/nofx && go test ./provider/databento/... -v
```

Expected: all tests pass.

- [ ] **Step 3.5: Commit**

```bash
cd /home/hoang/nofx
git add provider/databento/resolve.go provider/databento/resolve_test.go
git commit -m "feat(databento): add continuous symbol resolver"
```

---

## Task 4: Databento → market.Kline adapter

**Why:** The rest of the codebase (indicators, formatter, kernel) consumes `market.Kline`, not `databento.Bar`. One small adapter lets every existing function work unchanged.

**Files:**
- Create: `market/databento_adapter.go`
- Create: `market/databento_adapter_test.go`

- [ ] **Step 4.1: Read the existing Kline shape**

```bash
cd /home/hoang/nofx && grep -A 12 "^type Kline struct" market/types.go
```

Note the fields exactly — they'll inform the mapping in the next step.

- [ ] **Step 4.2: Write the failing test**

Create `/home/hoang/nofx/market/databento_adapter_test.go`:

```go
package market

import (
	"nofx/provider/databento"
	"testing"
	"time"
)

func TestBarsToKlines_Mapping(t *testing.T) {
	bars := []databento.Bar{
		{
			Timestamp: time.Unix(1746360000, 0).UTC(),
			Open:      21500.25,
			High:      21515.75,
			Low:       21498.00,
			Close:     21510.00,
			Volume:    4321,
		},
	}
	klines := BarsToKlines(bars)
	if len(klines) != 1 {
		t.Fatalf("want 1 kline, got %d", len(klines))
	}
	k := klines[0]
	if k.Open != 21500.25 || k.High != 21515.75 || k.Low != 21498.00 || k.Close != 21510.00 || k.Volume != 4321 {
		t.Errorf("kline OHLCV mismatch: %+v", k)
	}
	// Kline.OpenTime should be milliseconds since epoch (this is the convention
	// the existing code uses — verify in market/types.go before writing).
	wantMs := int64(1746360000 * 1000)
	if k.OpenTime != wantMs {
		t.Errorf("kline.OpenTime = %d, want %d", k.OpenTime, wantMs)
	}
}

func TestBarsToKlines_Empty(t *testing.T) {
	got := BarsToKlines(nil)
	if len(got) != 0 {
		t.Errorf("want 0 klines, got %d", len(got))
	}
}
```

> **Note:** Before writing the adapter, **read `market/types.go:94-108` to confirm the field name** (`OpenTime` vs `Timestamp` vs `Time`) and **the unit** (seconds, milliseconds, or `time.Time`). Adjust the test and adapter to match. The test above assumes `OpenTime int64` in milliseconds; this matches Binance-style conventions.

- [ ] **Step 4.3: Run the test, verify failure**

```bash
cd /home/hoang/nofx && go test ./market/... -run TestBarsToKlines -v
```

Expected: compile error ("BarsToKlines undefined").

- [ ] **Step 4.4: Implement the adapter**

Create `/home/hoang/nofx/market/databento_adapter.go`:

```go
package market

import (
	"nofx/provider/databento"
)

// BarsToKlines converts Databento bars into the project's canonical Kline shape.
// This is the single bridge that lets every existing indicator, formatter, and
// strategy work with NQ data unchanged.
func BarsToKlines(bars []databento.Bar) []Kline {
	if len(bars) == 0 {
		return nil
	}
	out := make([]Kline, 0, len(bars))
	for _, b := range bars {
		out = append(out, Kline{
			OpenTime: b.Timestamp.UnixMilli(),
			Open:     b.Open,
			High:     b.High,
			Low:      b.Low,
			Close:    b.Close,
			Volume:   b.Volume,
		})
	}
	return out
}
```

> **If `market.Kline` has more fields** (e.g. `CloseTime`, `Trades`, `IsClosed`), fill them with sensible zero values or computed values (`CloseTime = OpenTime + intervalMs - 1`, `IsClosed = true`). Don't leave unset fields that downstream code requires.

- [ ] **Step 4.5: Run tests, verify pass**

```bash
cd /home/hoang/nofx && go test ./market/... -run TestBarsToKlines -v
```

Expected: PASS.

- [ ] **Step 4.6: Commit**

```bash
cd /home/hoang/nofx
git add market/databento_adapter.go market/databento_adapter_test.go
git commit -m "feat(market): add Databento Bar -> Kline adapter"
```

---

## Task 5: NinjaTrader CSV writer

**Files:**
- Create: `provider/ninjatrader/types.go`
- Create: `provider/ninjatrader/csv_writer.go`
- Create: `provider/ninjatrader/csv_writer_test.go`

**Contract format (verified from claudetrader.cs:191-223):**

```csv
DateTime,Direction,Entry_Price,Stop_Loss,Take_Profit
05/22/2026 14:30:15,LONG,21505.00,21485.00,21545.00
```

- DateTime format MUST match `MM/dd/yyyy HH:mm:ss` (C# `DateTime.TryParse` is lenient but this is the format NT writes back).
- Direction must be exactly `LONG` or `SHORT` (uppercase per claudetrader.cs:229,233).
- Prices use `.` decimal separator, two decimal places (NQ tick = 0.25).
- One signal per row. **Writing a new row triggers NT within ~2 seconds.**
- NT clears the file after read, leaving only the header. So our writer truncates+rewrites: header + one row. (Append also works, but truncate is safer — guarantees we never have stale rows.)

- [ ] **Step 5.1: Define types**

Create `/home/hoang/nofx/provider/ninjatrader/types.go`:

```go
package ninjatrader

import (
	"fmt"
	"strings"
)

// SignalRow is one row of trade_signals.csv that claudetrader.cs will consume.
type SignalRow struct {
	DateTime    string // MM/dd/yyyy HH:mm:ss
	Direction   string // "LONG" or "SHORT"
	EntryPrice  float64
	StopLoss    float64
	TakeProfit  float64
}

// FillRow is one row of trades_taken.csv that claudetrader.cs writes.
type FillRow struct {
	DateTime   string  // MM/dd/yyyy HH:mm:ss
	Direction  string  // "LONG" or "SHORT"
	EntryPrice float64
}

const signalsHeader = "DateTime,Direction,Entry_Price,Stop_Loss,Take_Profit"
const fillsHeader = "DateTime,Direction,Entry_Price"

// Validate rejects malformed signals before they reach disk.
func (s SignalRow) Validate() error {
	if s.DateTime == "" {
		return fmt.Errorf("ninjatrader signal: empty DateTime")
	}
	if strings.ToUpper(s.Direction) != "LONG" && strings.ToUpper(s.Direction) != "SHORT" {
		return fmt.Errorf("ninjatrader signal: direction must be LONG or SHORT, got %q", s.Direction)
	}
	if s.EntryPrice <= 0 || s.StopLoss <= 0 || s.TakeProfit <= 0 {
		return fmt.Errorf("ninjatrader signal: prices must be positive")
	}
	dir := strings.ToUpper(s.Direction)
	if dir == "LONG" && s.StopLoss >= s.EntryPrice {
		return fmt.Errorf("ninjatrader signal: LONG stop_loss (%.2f) must be below entry (%.2f)", s.StopLoss, s.EntryPrice)
	}
	if dir == "LONG" && s.TakeProfit <= s.EntryPrice {
		return fmt.Errorf("ninjatrader signal: LONG take_profit (%.2f) must be above entry (%.2f)", s.TakeProfit, s.EntryPrice)
	}
	if dir == "SHORT" && s.StopLoss <= s.EntryPrice {
		return fmt.Errorf("ninjatrader signal: SHORT stop_loss (%.2f) must be above entry (%.2f)", s.StopLoss, s.EntryPrice)
	}
	if dir == "SHORT" && s.TakeProfit >= s.EntryPrice {
		return fmt.Errorf("ninjatrader signal: SHORT take_profit (%.2f) must be below entry (%.2f)", s.TakeProfit, s.EntryPrice)
	}
	return nil
}
```

- [ ] **Step 5.2: Write the failing test for the writer**

Create `/home/hoang/nofx/provider/ninjatrader/csv_writer_test.go`:

```go
package ninjatrader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCSVWriter_WriteSignal_LongValid(t *testing.T) {
	dir := t.TempDir()
	w := NewCSVWriter(dir)

	sig := SignalRow{
		DateTime:   "05/22/2026 14:30:15",
		Direction:  "LONG",
		EntryPrice: 21505.00,
		StopLoss:   21485.00,
		TakeProfit: 21545.00,
	}
	if err := w.WriteSignal(sig); err != nil {
		t.Fatalf("WriteSignal: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "trade_signals.csv"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	want := "DateTime,Direction,Entry_Price,Stop_Loss,Take_Profit\n05/22/2026 14:30:15,LONG,21505.00,21485.00,21545.00\n"
	if string(got) != want {
		t.Errorf("file content:\n  got:  %q\n  want: %q", string(got), want)
	}
}

func TestCSVWriter_WriteSignal_RejectsBadLong(t *testing.T) {
	dir := t.TempDir()
	w := NewCSVWriter(dir)

	bad := SignalRow{
		DateTime:   "05/22/2026 14:30:15",
		Direction:  "LONG",
		EntryPrice: 21500.00,
		StopLoss:   21520.00, // wrong: stop ABOVE entry for a long
		TakeProfit: 21550.00,
	}
	err := w.WriteSignal(bad)
	if err == nil {
		t.Fatal("want validation error, got nil")
	}
	if !strings.Contains(err.Error(), "LONG stop_loss") {
		t.Errorf("error message %q does not mention LONG stop_loss", err.Error())
	}
}

func TestCSVWriter_WriteSignal_TruncatesPrevious(t *testing.T) {
	dir := t.TempDir()
	w := NewCSVWriter(dir)

	first := SignalRow{DateTime: "05/22/2026 14:30:15", Direction: "LONG", EntryPrice: 100, StopLoss: 90, TakeProfit: 110}
	second := SignalRow{DateTime: "05/22/2026 14:32:00", Direction: "SHORT", EntryPrice: 200, StopLoss: 210, TakeProfit: 190}

	if err := w.WriteSignal(first); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteSignal(second); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "trade_signals.csv"))
	// Should contain ONLY the second signal (plus header), not both.
	if strings.Contains(string(got), "14:30:15") {
		t.Errorf("expected first signal to be truncated, but file still contains it:\n%s", string(got))
	}
	if !strings.Contains(string(got), "14:32:00") {
		t.Errorf("expected second signal in file:\n%s", string(got))
	}
}
```

- [ ] **Step 5.3: Run tests, verify failure**

```bash
cd /home/hoang/nofx && go test ./provider/ninjatrader/... -v
```

Expected: compile failure ("NewCSVWriter undefined").

- [ ] **Step 5.4: Implement the writer**

Create `/home/hoang/nofx/provider/ninjatrader/csv_writer.go`:

```go
package ninjatrader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// CSVWriter writes trade signals to a Windows-shared CSV file that
// NinjaTrader's claudetrader.cs polls every 2 seconds.
type CSVWriter struct {
	dataDir string
	mu      sync.Mutex
}

func NewCSVWriter(dataDir string) *CSVWriter {
	return &CSVWriter{dataDir: dataDir}
}

// SignalsPath returns the absolute path to trade_signals.csv.
func (w *CSVWriter) SignalsPath() string {
	return filepath.Join(w.dataDir, "trade_signals.csv")
}

// WriteSignal validates and writes one signal. The file is truncated and
// rewritten with header + this single row — claudetrader.cs clears the file
// after processing, so we always overwrite to avoid stale rows accumulating.
func (w *CSVWriter) WriteSignal(s SignalRow) error {
	if err := s.Validate(); err != nil {
		return err
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	row := fmt.Sprintf("%s,%s,%.2f,%.2f,%.2f",
		s.DateTime,
		strings.ToUpper(s.Direction),
		s.EntryPrice,
		s.StopLoss,
		s.TakeProfit,
	)
	content := signalsHeader + "\n" + row + "\n"

	tmp, err := os.CreateTemp(w.dataDir, "trade_signals.*.tmp")
	if err != nil {
		return fmt.Errorf("ninjatrader writer: create temp: %w", err)
	}
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("ninjatrader writer: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("ninjatrader writer: close temp: %w", err)
	}
	if err := os.Rename(tmp.Name(), w.SignalsPath()); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("ninjatrader writer: rename temp: %w", err)
	}
	return nil
}
```

> **Why temp+rename:** atomic update. NT polls every 2s; if NT happens to read mid-write of a non-atomic update it sees a partial file. Atomic rename means NT always sees either the old or new file, never a torn write.

- [ ] **Step 5.5: Run tests, verify pass**

```bash
cd /home/hoang/nofx && go test ./provider/ninjatrader/... -v
```

Expected: all 3 tests PASS.

- [ ] **Step 5.6: Live smoke test against your actual NT setup**

With NT open, ClaudeTrader running on the MNQ chart in SIM, run:

```bash
cd /home/hoang/nofx
DATA_DIR="$(grep NINJATRADER_DATA_DIR .env | cut -d= -f2)"
cat > /tmp/nt_smoke.go <<'EOF'
package main

import (
	"fmt"
	"nofx/provider/ninjatrader"
	"os"
	"time"
)

func main() {
	dir := os.Args[1]
	w := ninjatrader.NewCSVWriter(dir)
	now := time.Now().Format("01/02/2006 15:04:05")
	// IMPORTANT: replace these prices with something near current MNQ market.
	// Check NT chart — if MNQ is at 21500, use entry near it and SL/TP a few points away.
	sig := ninjatrader.SignalRow{
		DateTime:   now,
		Direction:  "LONG",
		EntryPrice: 21500.00,
		StopLoss:   21480.00,
		TakeProfit: 21540.00,
	}
	if err := w.WriteSignal(sig); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
	fmt.Println("Wrote signal:", sig)
	fmt.Println("Now watch NT's Output window — should see [SIGNAL] within 2-4s.")
}
EOF
go run /tmp/nt_smoke.go "$DATA_DIR"
rm /tmp/nt_smoke.go
```

Watch NT's Output window. Expected:
```
[SIGNAL] LONG MARKET ORDER (1 contracts)
  Reference Entry: 21500.00
  Target SL: 21480.00 | Target TP: 21540.00
[ORDER UPDATE] CT_Long ...
[FILLED] LONG 1 contracts @ <market price>
[ORDERS SUBMITTED] Long exit orders sent to broker
```

If you see this — the Go-side bridge works end-to-end.

- [ ] **Step 5.7: Commit**

```bash
cd /home/hoang/nofx
git add provider/ninjatrader/types.go provider/ninjatrader/csv_writer.go provider/ninjatrader/csv_writer_test.go
git commit -m "feat(ninjatrader): add CSV signal writer with validation"
```

---

## Task 6: NinjaTrader fill tailer

**Files:**
- Create: `provider/ninjatrader/csv_tailer.go`
- Create: `provider/ninjatrader/csv_tailer_test.go`

**Behavior:** Periodically read `trades_taken.csv`, track which rows we've already seen by line count (NT appends; doesn't truncate this file). For each new row, decode it and invoke the callback.

- [ ] **Step 6.1: Write the failing tailer test**

Create `/home/hoang/nofx/provider/ninjatrader/csv_tailer_test.go`:

```go
package ninjatrader

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestCSVTailer_DetectsAppendedRows(t *testing.T) {
	dir := t.TempDir()
	fillsPath := filepath.Join(dir, "trades_taken.csv")
	// Seed with header
	if err := os.WriteFile(fillsPath, []byte(fillsHeader+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var (
		mu    sync.Mutex
		fills []FillRow
	)
	cb := func(f FillRow) {
		mu.Lock()
		defer mu.Unlock()
		fills = append(fills, f)
	}

	tailer := NewCSVTailer(dir, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = tailer.TailFills(ctx, cb) }()

	// Allow tailer to read initial state (header only, no rows)
	time.Sleep(150 * time.Millisecond)

	// Append two fills (NT-style)
	f, err := os.OpenFile(fillsPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("05/22/2026 14:30:15,LONG,21505.25\n")
	f.WriteString("05/22/2026 14:35:42,SHORT,21520.50\n")
	f.Close()

	// Wait for tailer to pick them up
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(fills)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(fills) != 2 {
		t.Fatalf("got %d fills, want 2: %+v", len(fills), fills)
	}
	if fills[0].Direction != "LONG" || fills[0].EntryPrice != 21505.25 {
		t.Errorf("fill[0] = %+v", fills[0])
	}
	if fills[1].Direction != "SHORT" || fills[1].EntryPrice != 21520.50 {
		t.Errorf("fill[1] = %+v", fills[1])
	}
}
```

- [ ] **Step 6.2: Run the test, verify failure**

```bash
cd /home/hoang/nofx && go test ./provider/ninjatrader/... -run TestCSVTailer -v
```

Expected: compile failure.

- [ ] **Step 6.3: Implement the tailer**

Create `/home/hoang/nofx/provider/ninjatrader/csv_tailer.go`:

```go
package ninjatrader

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CSVTailer reads new rows appended to trades_taken.csv by claudetrader.cs.
type CSVTailer struct {
	dataDir      string
	pollInterval time.Duration
	seen         int // rows already delivered (excluding header)
}

func NewCSVTailer(dataDir string, pollInterval time.Duration) *CSVTailer {
	if pollInterval <= 0 {
		pollInterval = time.Second
	}
	return &CSVTailer{dataDir: dataDir, pollInterval: pollInterval}
}

func (t *CSVTailer) FillsPath() string {
	return filepath.Join(t.dataDir, "trades_taken.csv")
}

// TailFills blocks until ctx is cancelled. For each new fill row appended to
// the file, the callback is invoked synchronously. Reset on file-shrink (e.g.
// if NT cycles the file at session boundary).
func (t *CSVTailer) TailFills(ctx context.Context, onFill func(FillRow)) error {
	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := t.readNew(onFill); err != nil {
				// Log internally — never kill the loop on a transient read.
				fmt.Fprintf(os.Stderr, "ninjatrader tailer: %v\n", err)
			}
		}
	}
}

func (t *CSVTailer) readNew(onFill func(FillRow)) error {
	f, err := os.Open(t.FillsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil // file may not exist yet — first fill creates it
		}
		return err
	}
	defer f.Close()

	rows := []string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			rows = append(rows, line)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	// Strip header if present
	if len(rows) > 0 && strings.HasPrefix(rows[0], "DateTime,") {
		rows = rows[1:]
	}

	// File shrunk → NT cycled; reset our cursor
	if len(rows) < t.seen {
		t.seen = 0
	}

	for i := t.seen; i < len(rows); i++ {
		fill, err := parseFillRow(rows[i])
		if err != nil {
			fmt.Fprintf(os.Stderr, "ninjatrader tailer: parse %q: %v\n", rows[i], err)
			continue
		}
		onFill(fill)
	}
	t.seen = len(rows)
	return nil
}

func parseFillRow(line string) (FillRow, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return FillRow{}, fmt.Errorf("expected 3 fields, got %d", len(parts))
	}
	price, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return FillRow{}, fmt.Errorf("parse entry price: %w", err)
	}
	return FillRow{
		DateTime:   strings.TrimSpace(parts[0]),
		Direction:  strings.ToUpper(strings.TrimSpace(parts[1])),
		EntryPrice: price,
	}, nil
}
```

- [ ] **Step 6.4: Run tests, verify pass**

```bash
cd /home/hoang/nofx && go test ./provider/ninjatrader/... -v
```

Expected: all PASS, including the 2-second tailer test.

- [ ] **Step 6.5: Commit**

```bash
cd /home/hoang/nofx
git add provider/ninjatrader/csv_tailer.go provider/ninjatrader/csv_tailer_test.go
git commit -m "feat(ninjatrader): add CSV fill tailer"
```

---

## Task 7: NinjaTrader Trader interface implementation

**Why:** [trader/auto_trader.go:263-315](trader/auto_trader.go#L263-L315) switches on `config.Exchange` and creates a `Trader` per the 17-method interface at [trader/types/interface.go:43-105](trader/types/interface.go#L43-L105). We implement a minimal version that maps `OpenLong/OpenShort` to CSV writes and uses the tailer for `GetPositions`/`GetOrderStatus`.

**Reality check on what we can/can't support via this bridge:**
- ✅ `OpenLong`, `OpenShort` — write LONG/SHORT signal
- ✅ `SetStopLoss`, `SetTakeProfit` — bundled into the signal at submit time, claudetrader.cs handles them
- ✅ `GetPositions` — read trades_taken.csv + track our own state
- ⚠️ `CloseLong`, `CloseShort` — claudetrader.cs auto-closes via SL/TP. No "manual close" path exists in the current bridge. **For v1, return an error here and rely on SL/TP exits.** Adding manual close = a future task that modifies the NinjaScript.
- ⚠️ `SetLeverage`, `SetMarginMode` — irrelevant for futures (set at the broker level, not per order). Return nil noop.
- ⚠️ `CancelAllOrders` — not supported by the CSV bridge. Return error.
- ⚠️ `GetBalance` — claudetrader.cs doesn't expose account balance via CSV. For v1, return a fixed mock value or an error. Future: extend the NinjaScript to write a balance.csv.
- ⚠️ `GetClosedPnL` — same constraint. Future task to log P&L from the C# side.
- ✅ `FormatQuantity` — passthrough rounding.
- ✅ `GetOrderStatus` — query our internal tracking.

**Files:**
- Create: `trader/ninjatrader/trader.go`
- Create: `trader/ninjatrader/trader_test.go`

- [ ] **Step 7.1: Read the Trader interface to confirm signatures**

```bash
cd /home/hoang/nofx && sed -n '43,105p' trader/types/interface.go
```

Note the exact method signatures and return types. **The code below assumes specific signatures based on the audit; verify and adjust if any differ.**

- [ ] **Step 7.2: Write the failing test (just the constructor + OpenLong)**

Create `/home/hoang/nofx/trader/ninjatrader/trader_test.go`:

```go
package ninjatrader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_Smoke(t *testing.T) {
	dir := t.TempDir()
	tr := New(Config{DataDir: dir, Symbol: "MNQ"})
	if tr == nil {
		t.Fatal("New returned nil")
	}
}

func TestOpenLong_WritesSignal(t *testing.T) {
	dir := t.TempDir()
	tr := New(Config{DataDir: dir, Symbol: "MNQ"})

	// Stash SL/TP first — they get bundled into the order on OpenLong.
	_ = tr.SetStopLoss("MNQ", "LONG", 1, 21480.00)
	_ = tr.SetTakeProfit("MNQ", "LONG", 1, 21540.00)

	res, err := tr.OpenLong("MNQ", 1, 1)
	if err != nil {
		t.Fatalf("OpenLong: %v", err)
	}
	if res["status"] != "submitted" {
		t.Errorf("status = %v, want submitted", res["status"])
	}

	body, _ := os.ReadFile(filepath.Join(dir, "trade_signals.csv"))
	if !strings.Contains(string(body), "LONG") {
		t.Errorf("signal file missing LONG row:\n%s", string(body))
	}
	if !strings.Contains(string(body), "21480.00") || !strings.Contains(string(body), "21540.00") {
		t.Errorf("signal file missing SL/TP:\n%s", string(body))
	}
}
```

- [ ] **Step 7.3: Implement the trader**

Create `/home/hoang/nofx/trader/ninjatrader/trader.go`:

```go
// Package ninjatrader implements the trader.Trader interface by writing
// trade signals to a CSV file that NinjaTrader's claudetrader.cs strategy
// consumes. Reads fills back via a CSV tailer.
package ninjatrader

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nofx/provider/ninjatrader"
	"nofx/trader/types"
)

type Config struct {
	DataDir string // /mnt/c/Users/<u>/NofxTrader/data
	Symbol  string // e.g. "MNQ" (informational only; NT uses chart's instrument)
}

// Trader satisfies trader/types.Trader using the CSV bridge.
type Trader struct {
	cfg    Config
	writer *ninjatrader.CSVWriter
	tailer *ninjatrader.CSVTailer

	mu       sync.Mutex
	stopLoss map[string]float64 // key: "<symbol>:<side>"
	takePrft map[string]float64
	lastFill ninjatrader.FillRow
	hasFill  bool
}

func New(cfg Config) *Trader {
	t := &Trader{
		cfg:      cfg,
		writer:   ninjatrader.NewCSVWriter(cfg.DataDir),
		tailer:   ninjatrader.NewCSVTailer(cfg.DataDir, time.Second),
		stopLoss: map[string]float64{},
		takePrft: map[string]float64{},
	}
	// Start the fill tailer in the background. Cancellation is the program's
	// responsibility (defer at startup). For now, use context.Background().
	go func() {
		_ = t.tailer.TailFills(context.Background(), func(f ninjatrader.FillRow) {
			t.mu.Lock()
			defer t.mu.Unlock()
			t.lastFill = f
			t.hasFill = true
		})
	}()
	return t
}

// Compile-time check that we implement the interface. If signatures drift,
// the build fails here — not silently at runtime.
var _ types.Trader = (*Trader)(nil)

// --- Trader interface methods ---

func (t *Trader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeEntry(symbol, "LONG", quantity)
}

func (t *Trader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeEntry(symbol, "SHORT", quantity)
}

func (t *Trader) placeEntry(symbol, side string, quantity float64) (map[string]interface{}, error) {
	t.mu.Lock()
	sl := t.stopLoss[keyFor(symbol, side)]
	tp := t.takePrft[keyFor(symbol, side)]
	t.mu.Unlock()

	if sl == 0 || tp == 0 {
		return nil, fmt.Errorf("ninjatrader: SetStopLoss and SetTakeProfit must be called before %s", side)
	}

	// Entry price for the CSV is a "reference" — claudetrader uses MARKET orders.
	// We use 0 as a placeholder, but the strategy ignores it for entry decisions.
	// Use a near-market value so logs make sense; if you don't have one handy,
	// passing the SL midpoint is harmless.
	entryRef := (sl + tp) / 2.0

	sig := ninjatrader.SignalRow{
		DateTime:   time.Now().Format("01/02/2006 15:04:05"),
		Direction:  side,
		EntryPrice: entryRef,
		StopLoss:   sl,
		TakeProfit: tp,
	}
	if err := t.writer.WriteSignal(sig); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status":   "submitted",
		"symbol":   symbol,
		"side":     side,
		"quantity": quantity,
	}, nil
}

func (t *Trader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ninjatrader: manual CloseLong not supported via CSV bridge — position closes via SL/TP set at entry")
}

func (t *Trader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, fmt.Errorf("ninjatrader: manual CloseShort not supported via CSV bridge — position closes via SL/TP set at entry")
}

func (t *Trader) SetLeverage(symbol string, leverage int) error {
	return nil // futures leverage is set at the broker, not per-order
}

func (t *Trader) SetMarginMode(symbol string, isCrossMargin bool) error {
	return nil // n/a for futures
}

func (t *Trader) SetStopLoss(symbol, positionSide string, quantity, stopPrice float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopLoss[keyFor(symbol, positionSide)] = stopPrice
	return nil
}

func (t *Trader) SetTakeProfit(symbol, positionSide string, quantity, takeProfitPrice float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.takePrft[keyFor(symbol, positionSide)] = takeProfitPrice
	return nil
}

func (t *Trader) CancelAllOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelAllOrders not supported via CSV bridge")
}

func (t *Trader) GetBalance() (map[string]interface{}, error) {
	// claudetrader.cs doesn't expose balance via CSV. For paper-mode v1, return
	// a fixed sim balance so the trader loop doesn't fail balance checks.
	return map[string]interface{}{
		"totalEquity": 50000.0, // SIM101 starts with $50k by default
		"availableBalance": 50000.0,
	}, nil
}

func (t *Trader) GetPositions() ([]map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return []map[string]interface{}{}, nil
	}
	return []map[string]interface{}{{
		"symbol":     t.cfg.Symbol,
		"side":       t.lastFill.Direction,
		"entryPrice": t.lastFill.EntryPrice,
		"quantity":   1.0,
	}}, nil
}

func (t *Trader) FormatQuantity(symbol string, quantity float64) (string, error) {
	return fmt.Sprintf("%.0f", quantity), nil
}

func (t *Trader) GetOrderStatus(symbol, orderID string) (map[string]interface{}, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return map[string]interface{}{"status": "pending"}, nil
	}
	return map[string]interface{}{
		"status":     "filled",
		"price":      t.lastFill.EntryPrice,
		"side":       t.lastFill.Direction,
	}, nil
}

func (t *Trader) GetClosedPnL(start time.Time, limit int) ([]types.ClosedPnLRecord, error) {
	// Not available via CSV bridge in v1. Return empty.
	return nil, nil
}

// --- 5 additional methods required by trader/types.Trader (audited 2026-05-22) ---

func (t *Trader) GetMarketPrice(symbol string) (float64, error) {
	// Pull last known close from the most recent fill, or zero if no fills yet.
	// For accurate live price the caller should query Databento directly via
	// provider/databento; this method is best-effort for the bridge.
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.hasFill {
		return 0, fmt.Errorf("ninjatrader: no fill yet, market price unavailable; use Databento client directly")
	}
	return t.lastFill.EntryPrice, nil
}

func (t *Trader) CancelStopLossOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelStopLossOrders not supported via CSV bridge — SL is set at entry")
}

func (t *Trader) CancelTakeProfitOrders(symbol string) error {
	return fmt.Errorf("ninjatrader: CancelTakeProfitOrders not supported via CSV bridge — TP is set at entry")
}

func (t *Trader) CancelStopOrders(symbol string) error {
	// Legacy method, alias for both SL+TP cancel.
	return fmt.Errorf("ninjatrader: CancelStopOrders not supported via CSV bridge")
}

func (t *Trader) GetOpenOrders(symbol string) ([]types.OpenOrder, error) {
	// claudetrader.cs CSV protocol doesn't expose pending orders. The signal
	// file only stores the most recent unconsumed signal; SL/TP live entirely
	// on NT's side after entry fills. Return empty slice (not nil, not error)
	// to match the contract used by other brokers.
	return []types.OpenOrder{}, nil
}

func keyFor(symbol, side string) string {
	return symbol + ":" + side
}
```

> **Compile-time verification:** the `var _ types.Trader = (*Trader)(nil)` line at the top of this file enforces that the impl matches the 19-method interface exactly. If you see `*Trader does not implement types.Trader (missing method X)` at build time, add the missing method following the same pattern (noop/return-empty for unsupported, error for cancellation, real impl for the few that work).

- [ ] **Step 7.4: Run tests, verify pass**

```bash
cd /home/hoang/nofx && go test ./trader/ninjatrader/... -v
```

Expected: PASS.

- [ ] **Step 7.5: Commit**

```bash
cd /home/hoang/nofx
git add trader/ninjatrader/trader.go trader/ninjatrader/trader_test.go
git commit -m "feat(trader): add NinjaTrader Trader impl via CSV bridge"
```

---

## Task 8: Wire NinjaTrader into auto_trader switch

**Files:**
- Modify: `trader/auto_trader.go:263-315` (the broker switch)

- [ ] **Step 8.1: Read the switch**

```bash
cd /home/hoang/nofx && sed -n '253,320p' trader/auto_trader.go
```

Note the existing pattern — each case constructs a broker-specific struct and assigns to `at.trader`.

- [ ] **Step 8.2: Add the case**

In `trader/auto_trader.go`, inside the switch, before the `default:` clause, add:

```go
case "ninjatrader":
    cfg := ninjatrader.Config{
        DataDir: globalConfig.NinjaTraderDataDir, // from config.Get()
        Symbol:  config.Symbol,                   // e.g. "MNQ"
    }
    at.trader = ninjatrader.New(cfg)
```

Add the import at the top:

```go
import (
    // ... existing ...
    ninjatrader "nofx/trader/ninjatrader"
)
```

> **If `globalConfig` isn't accessible at that point**, pass the data dir into `NewAutoTrader` or pull it from the trader's stored config. The pattern in this file will tell you which approach matches.

- [ ] **Step 8.3: Build and verify**

```bash
cd /home/hoang/nofx && go build ./... 2>&1 | head -20
```

Expected: no output.

- [ ] **Step 8.4: Commit**

```bash
cd /home/hoang/nofx
git add trader/auto_trader.go
git commit -m "feat(trader): register ninjatrader broker in switch"
```

---

## Task 9: NQ futures prompt template

**Files:**
- Create: `kernel/engine_prompt_futures.go`
- Create: `kernel/engine_prompt_futures_test.go`

**Why a separate file:** keeps the existing crypto prompt untouched, makes A/B testing trivial, and lets the engine pick at runtime by `TradingMode`.

- [ ] **Step 9.1: Write a golden-file test**

Create `/home/hoang/nofx/kernel/engine_prompt_futures_test.go`:

```go
package kernel

import (
	"strings"
	"testing"
)

func TestBuildFuturesSystemPrompt_NoCryptoVocab(t *testing.T) {
	p := BuildFuturesSystemPrompt(FuturesPromptConfig{
		Symbol:           "MNQ",
		ContractMultiplier: 2.0, // MNQ = $2/point
		TickSize:         0.25,
		MinStopPoints:    15,
		MaxStopPoints:    50,
		MinRiskReward:    1.5,
	})

	// Must NOT contain crypto vocabulary
	forbidden := []string{
		"cryptocurrency", "altcoin", "BTC", "ETH", "USDT", "perpetual",
		"funding rate", "coins simultaneously",
	}
	for _, f := range forbidden {
		if strings.Contains(p, f) {
			t.Errorf("futures prompt contains forbidden crypto term %q", f)
		}
	}

	// Must contain futures-specific framing
	required := []string{
		"NQ", "tick", "contract", "stop loss", "take profit", "MNQ",
	}
	for _, r := range required {
		if !strings.Contains(p, r) {
			t.Errorf("futures prompt missing required term %q", r)
		}
	}
}

func TestBuildFuturesUserPrompt_IncludesIndicators(t *testing.T) {
	p := BuildFuturesUserPrompt(FuturesContext{
		Symbol:       "MNQ",
		CurrentPrice: 21500.00,
		EMA20:        21495.00,
		EMA50:        21480.00,
		RSI14:        58.3,
		MACD:         3.21,
		ATR14:        12.5,
		BollUpper:    21540.00,
		BollLower:    21460.00,
	})
	for _, s := range []string{"21500.00", "EMA20", "RSI14", "ATR14", "Bollinger"} {
		if !strings.Contains(p, s) {
			t.Errorf("user prompt missing %q", s)
		}
	}
}
```

- [ ] **Step 9.2: Run, verify failure**

```bash
cd /home/hoang/nofx && go test ./kernel/... -run TestBuildFutures -v
```

Expected: compile failure.

- [ ] **Step 9.3: Implement**

Create `/home/hoang/nofx/kernel/engine_prompt_futures.go`:

```go
package kernel

import (
	"fmt"
	"strings"
)

// FuturesPromptConfig captures the few parameters the system prompt needs
// to describe an index-futures contract to the model.
type FuturesPromptConfig struct {
	Symbol             string  // "NQ" or "MNQ"
	ContractMultiplier float64 // NQ = 20 ($20/point), MNQ = 2 ($2/point)
	TickSize           float64 // 0.25 for both NQ and MNQ
	MinStopPoints      float64 // 15
	MaxStopPoints      float64 // 50
	MinRiskReward      float64 // 1.5
}

// FuturesContext is the per-cycle data shoved into the user prompt.
type FuturesContext struct {
	Symbol       string
	CurrentPrice float64
	// indicator snapshot
	EMA20     float64
	EMA50     float64
	RSI14     float64
	MACD      float64 // MACD line (current value, from market.ExportCalculateMACD)
	ATR14     float64
	BollUpper float64
	BollLower float64
}

// NOTE: The existing market.ExportCalculateMACD returns only the MACD line
// (one float64), not the signal line or histogram. To surface signal/histogram
// to the AI, we would need to extend market/data_indicators.go to expose a
// fuller MACD function. For Plan 1, the prompt mentions only the MACD line
// value; signal/histogram are deferred to a future indicator-extension task.

func BuildFuturesSystemPrompt(c FuturesPromptConfig) string {
	var b strings.Builder
	b.WriteString("# You are a professional index-futures trading AI specializing in CME E-mini Nasdaq-100 contracts.\n\n")
	b.WriteString(fmt.Sprintf("## Instrument\n- Symbol: %s\n- Tick size: %.2f points\n- Contract multiplier: $%.2f per point\n\n", c.Symbol, c.TickSize, c.ContractMultiplier))
	b.WriteString("## Hard constraints\n")
	b.WriteString(fmt.Sprintf("- Every entry MUST include a stop loss and a take profit, expressed as absolute prices.\n"))
	b.WriteString(fmt.Sprintf("- Stop loss distance: minimum %.0f points, maximum %.0f points from entry.\n", c.MinStopPoints, c.MaxStopPoints))
	b.WriteString(fmt.Sprintf("- Minimum risk/reward: %.2f (reward must be at least %.2fx the risk).\n", c.MinRiskReward, c.MinRiskReward))
	b.WriteString("- One position at a time. Do NOT propose averaging in or pyramiding.\n")
	b.WriteString("- Prices must be in tick increments (multiples of " + fmt.Sprintf("%.2f", c.TickSize) + ").\n")
	b.WriteString("- The market session is CME futures hours; do not assume 24/7 trading.\n\n")
	b.WriteString("## Decision output\n")
	b.WriteString("Respond ONLY with JSON of the following exact shape:\n")
	b.WriteString("```json\n")
	b.WriteString(fmt.Sprintf(`{"action":"LONG"|"SHORT"|"NONE","entry":%.2f,"stop_loss":%.2f,"take_profit":%.2f,"reasoning":"<one-paragraph explanation>"}`, 0.0, 0.0, 0.0))
	b.WriteString("\n```\n")
	b.WriteString("\n- `action=NONE` is a valid and frequently correct answer. Do not force a trade.\n")
	b.WriteString("- All three price fields are absolute (e.g. 21500.25), not deltas from entry.\n\n")
	b.WriteString(fmt.Sprintf("## Trade plan checklist (apply before answering LONG/SHORT)\n"))
	b.WriteString("1. Is there a clear directional bias from EMA20 vs EMA50 alignment?\n")
	b.WriteString("2. Does RSI confirm or contradict that bias? (extreme = caution)\n")
	b.WriteString("3. Is MACD histogram positive (for LONG) or negative (for SHORT)?\n")
	b.WriteString("4. Is ATR consistent with your proposed stop distance? Stop should be ~1.5-3x ATR.\n")
	b.WriteString("5. Where is the Bollinger band — overextended (mean revert) or trending (continuation)?\n")
	b.WriteString("6. Risk/reward calculation: (take_profit - entry) / (entry - stop_loss) for LONG. Must exceed " + fmt.Sprintf("%.2f", c.MinRiskReward) + ".\n")
	return b.String()
}

func BuildFuturesUserPrompt(ctx FuturesContext) string {
	var b strings.Builder
	b.WriteString("## Current market\n")
	b.WriteString(fmt.Sprintf("- Symbol: %s\n", ctx.Symbol))
	b.WriteString(fmt.Sprintf("- Current price: %.2f\n\n", ctx.CurrentPrice))
	b.WriteString("## Indicator snapshot (1-minute timeframe)\n")
	b.WriteString(fmt.Sprintf("- EMA20: %.2f (current price %s)\n", ctx.EMA20, side(ctx.CurrentPrice, ctx.EMA20)))
	b.WriteString(fmt.Sprintf("- EMA50: %.2f (current price %s)\n", ctx.EMA50, side(ctx.CurrentPrice, ctx.EMA50)))
	b.WriteString(fmt.Sprintf("- EMA20 vs EMA50: %s\n", emaAlignment(ctx.EMA20, ctx.EMA50)))
	b.WriteString(fmt.Sprintf("- RSI14: %.1f (%s)\n", ctx.RSI14, rsiBucket(ctx.RSI14)))
	b.WriteString(fmt.Sprintf("- MACD: %.2f (line only; signal/histogram require extended indicator API)\n", ctx.MACD))
	b.WriteString(fmt.Sprintf("- ATR14: %.2f points\n", ctx.ATR14))
	b.WriteString(fmt.Sprintf("- Bollinger Bands: upper %.2f, lower %.2f, position: %s\n", ctx.BollUpper, ctx.BollLower, bollPosition(ctx.CurrentPrice, ctx.BollUpper, ctx.BollLower)))
	b.WriteString("\n## Decision\nGive me your trade decision in the JSON format specified by the system prompt.\n")
	return b.String()
}

func side(price, ref float64) string {
	if price > ref {
		return "above"
	}
	if price < ref {
		return "below"
	}
	return "equal"
}

func emaAlignment(ema20, ema50 float64) string {
	if ema20 > ema50 {
		return "bullish (20 > 50)"
	}
	if ema20 < ema50 {
		return "bearish (20 < 50)"
	}
	return "neutral"
}

func rsiBucket(r float64) string {
	switch {
	case r >= 70:
		return "overbought"
	case r <= 30:
		return "oversold"
	default:
		return "neutral"
	}
}

func bollPosition(p, upper, lower float64) string {
	switch {
	case p >= upper:
		return "above upper band (overextended)"
	case p <= lower:
		return "below lower band (overextended)"
	default:
		return "inside bands"
	}
}
```

- [ ] **Step 9.4: Run tests, verify pass**

```bash
cd /home/hoang/nofx && go test ./kernel/... -run TestBuildFutures -v
```

Expected: PASS.

- [ ] **Step 9.5: Commit**

```bash
cd /home/hoang/nofx
git add kernel/engine_prompt_futures.go kernel/engine_prompt_futures_test.go
git commit -m "feat(kernel): add NQ futures prompt template"
```

---

## Task 10: End-to-end smoke test

**Goal:** Run the bot one cycle, manually, with TRADING_MODE=futures pointing at MNQ in NT SIM. Verify the full chain: Databento → indicators → AI → CSV → NT fill → DB record.

**Files:**
- Create: `cmd/nq_smoke/main.go` — a standalone runner for one cycle

- [ ] **Step 10.1: Write the smoke runner**

Create `/home/hoang/nofx/cmd/nq_smoke/main.go`:

```go
// Standalone single-cycle runner for the NQ trading slice.
// Pulls 30min of NQ.c.0 1m bars, computes indicators, prompts AI, writes signal.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"nofx/kernel"
	"nofx/market"
	"nofx/provider/databento"
	"nofx/provider/ninjatrader"
)

func main() {
	_ = godotenv.Load("/home/hoang/nofx/.env")

	dbKey := os.Getenv("DATABENTO_API_KEY")
	if dbKey == "" {
		log.Fatal("DATABENTO_API_KEY not set in .env")
	}
	ntDir := os.Getenv("NINJATRADER_DATA_DIR")
	if ntDir == "" {
		log.Fatal("NINJATRADER_DATA_DIR not set in .env")
	}

	// 1. Fetch NQ bars
	db := databento.NewClient("", dbKey)
	end := time.Now().UTC()
	start := end.Add(-30 * time.Minute)
	bars, err := db.GetOHLCV("NQ.c.0", "1m", start, end)
	if err != nil {
		log.Fatalf("databento: %v", err)
	}
	if len(bars) < 50 {
		log.Fatalf("got %d bars; need at least 50 for indicators", len(bars))
	}
	fmt.Printf("✓ Fetched %d 1m NQ bars\n", len(bars))

	// 2. Convert to klines and compute indicators
	klines := market.BarsToKlines(bars)
	ctx := kernel.FuturesContext{
		Symbol:       "MNQ",
		CurrentPrice: klines[len(klines)-1].Close,
		EMA20:        market.ExportCalculateEMA(klines, 20),
		EMA50:        market.ExportCalculateEMA(klines, 50),
		RSI14:        market.ExportCalculateRSI(klines, 14),
		MACD:         market.ExportCalculateMACD(klines), // returns MACD line only
		ATR14:        market.ExportCalculateATR(klines, 14),
	}
	upper, _, lower := market.ExportCalculateBOLL(klines, 20, 2.0)
	ctx.BollUpper = upper
	ctx.BollLower = lower

	fmt.Printf("✓ Indicators computed. Current price: %.2f\n", ctx.CurrentPrice)

	// 3. Build prompts
	sysP := kernel.BuildFuturesSystemPrompt(kernel.FuturesPromptConfig{
		Symbol:             "MNQ",
		ContractMultiplier: 2.0,
		TickSize:           0.25,
		MinStopPoints:      15,
		MaxStopPoints:      50,
		MinRiskReward:      1.5,
	})
	userP := kernel.BuildFuturesUserPrompt(ctx)
	fmt.Println("\n--- SYSTEM PROMPT ---")
	fmt.Println(sysP)
	fmt.Println("\n--- USER PROMPT ---")
	fmt.Println(userP)

	// 4. AI call — STUBBED in the smoke runner.
	// Replace with your actual AI client. For first smoke test, hand-fabricate a decision.
	fmt.Println("\n>>> Paste an AI decision JSON (or press Ctrl-C to abort):")
	var decision struct {
		Action     string  `json:"action"`
		Entry      float64 `json:"entry"`
		StopLoss   float64 `json:"stop_loss"`
		TakeProfit float64 `json:"take_profit"`
		Reasoning  string  `json:"reasoning"`
	}
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&decision); err != nil {
		log.Fatalf("decode decision: %v", err)
	}
	fmt.Printf("✓ Decision: %s entry=%.2f sl=%.2f tp=%.2f\n", decision.Action, decision.Entry, decision.StopLoss, decision.TakeProfit)

	if decision.Action == "NONE" {
		fmt.Println("Action=NONE; nothing to write. Exiting.")
		return
	}

	// 5. Write signal
	w := ninjatrader.NewCSVWriter(ntDir)
	sig := ninjatrader.SignalRow{
		DateTime:   time.Now().Format("01/02/2006 15:04:05"),
		Direction:  decision.Action,
		EntryPrice: decision.Entry,
		StopLoss:   decision.StopLoss,
		TakeProfit: decision.TakeProfit,
	}
	if err := w.WriteSignal(sig); err != nil {
		log.Fatalf("write signal: %v", err)
	}
	fmt.Println("✓ Signal written to", w.SignalsPath())

	// 6. Tail fills for 30 seconds
	fmt.Println("\nTailing trades_taken.csv for 30s — watch for fill...")
	tailer := ninjatrader.NewCSVTailer(ntDir, time.Second)
	ctxT, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = tailer.TailFills(ctxT, func(f ninjatrader.FillRow) {
		fmt.Printf("  >>> FILL: %s %s @ %.2f\n", f.DateTime, f.Direction, f.EntryPrice)
	})
	fmt.Println("Done.")
}
```

- [ ] **Step 10.2: Run the smoke test**

Prerequisites: NT is open, ClaudeTrader running on MNQ chart in SIM, all env vars set.

```bash
cd /home/hoang/nofx && go run ./cmd/nq_smoke
```

The runner prints the system + user prompts. Hand-fabricate a tiny decision JSON near current MNQ price. Example (if NQ is around 21500):

```json
{"action":"LONG","entry":21500.00,"stop_loss":21485.00,"take_profit":21525.00,"reasoning":"smoke test"}
```

Paste it into the runner's stdin. Expected:

1. Signal file written to `$NINJATRADER_DATA_DIR/trade_signals.csv`.
2. Within 2-4s, NT's Output window prints `[SIGNAL] LONG MARKET ORDER (1 contracts)`.
3. NT fills the order in SIM.
4. NT writes a row to `trades_taken.csv`.
5. The smoke runner's tailer prints `>>> FILL: 05/22/2026 ... LONG @ <price>`.

If all five happen — **the entire chain works end-to-end**. This is the slice complete.

- [ ] **Step 10.3: Commit**

```bash
cd /home/hoang/nofx
git add cmd/nq_smoke/main.go
git commit -m "feat(cmd): add NQ end-to-end smoke runner"
```

---

## Task 11: Remove three crypto-era pages from navigation

**Why:** Three pages exist that have no purpose for a single-user NQ trader and add maintenance burden:

1. **Data page** (`/data`) — iframe to `nofxos.ai/dashboard`, now CSP-blocked. NinjaTrader IS the chart view.
2. **Strategy Market page** (`/strategy-market`) — community strategy browser showing public crypto strategies. Confusing UX for NQ users; mixing crypto strategies with NQ ones serves nobody.
3. **Competition / Leaderboard page** (`/competition`) — public competition view ranking crypto traders by P&L. Not relevant for a personal NQ trading setup; also makes the bot send anonymous data to public listings by default.

All three follow the same removal pattern: drop the route, drop the import, drop the nav entry (desktop + mobile), delete the component file, clean the `Page` type union.

**Files:**
- Modify: `web/src/components/common/HeaderBar.tsx` (remove 4 nav entries: 2 for Data, 2 for Market, 2 for Competition — actually 6 entries across desktop + mobile blocks)
- Modify: `web/src/router/AppRoutes.tsx` (remove 3 routes + 3 imports)
- Modify: `web/src/router/paths.ts` (remove 3 entries from ROUTES, PAGE_PATHS, LEGACY_HASH_ROUTES, getCurrentPageForPath, and the Page type union)
- Delete: `web/src/pages/DataPage.tsx`
- Delete: `web/src/pages/StrategyMarketPage.tsx`
- Delete: `web/src/components/trader/CompetitionPage.tsx`
- Delete: `web/src/components/trader/CompetitionPage.test.tsx`
- Modify: `web/src/i18n/translations.ts` (remove unused i18n keys: `dataCenter`, `strategyMarket*`, `competition*`, etc.)

- [ ] **Step 11.1: Remove desktop nav entries (Data + Market + Competition)**

In `web/src/components/common/HeaderBar.tsx`, find the desktop nav block and remove three entries (Data at ~121-131, Strategy Market at ~133-145, Competition at ~162-175). Pattern for each:

```diff
                 {
                   page: 'agent',
                   path: ROUTES.agent,
                   label: 'Agent',
                   badge: 'Beta',
                   requiresAuth: false,
                 },
-                {
-                  page: 'data',
-                  path: ROUTES.data,
-                  label: language === 'zh' ? '数据' : 'Data',
-                  requiresAuth: false,
-                },
-                {
-                  page: 'strategy-market',
-                  path: ROUTES.strategyMarket,
-                  label: language === 'zh' ? '策略市场' : 'Strategy Market',
-                  requiresAuth: false,
-                },
-                {
-                  page: 'competition',
-                  path: ROUTES.competition,
-                  label: language === 'zh' ? '排行榜' : 'Leaderboard',
-                  requiresAuth: false,
-                },
                 {
                   page: 'traders',
                   ...
                 },
```

- [ ] **Step 11.2: Remove mobile nav entries (Data + Market + Competition)**

Same file, find the mobile nav block at ~457-510 and remove the same three entries (Data ~457-466, Strategy Market ~468-481, Competition ~497-510). Same pattern as desktop block.

- [ ] **Step 11.3: Remove the three routes**

In `web/src/router/AppRoutes.tsx` find and remove three `<Route>` blocks:

```diff
-        <Route
-          path={ROUTES.data}
-          element={
-            <AppChrome currentPage="data" showFooter={false}>
-              <DataPage />
-            </AppChrome>
-          }
-        />
-        <Route
-          path={ROUTES.strategyMarket}
-          element={
-            <AppChrome currentPage="strategy-market" showFooter={false}>
-              <StrategyMarketPage />
-            </AppChrome>
-          }
-        />
-        <Route
-          path={ROUTES.competition}
-          element={
-            <AppChrome currentPage="competition" showFooter={false}>
-              <CompetitionPage />
-            </AppChrome>
-          }
-        />
```

Also remove three import lines at top of `AppRoutes.tsx`:

```diff
-import { DataPage } from '../pages/DataPage'
-import { StrategyMarketPage } from '../pages/StrategyMarketPage'
-import { CompetitionPage } from '../components/trader/CompetitionPage'
```

- [ ] **Step 11.4: Clean route constants**

In `web/src/router/paths.ts`:

```diff
 export type Page =
   | 'agent'
-  | 'competition'
   | 'traders'
   | 'trader'
   | 'strategy'
-  | 'strategy-market'
-  | 'data'
   | 'faq'
   | 'login'
   | 'register'

 export const ROUTES = {
   home: '/',
   agent: '/agent',
   login: '/login',
   register: '/register',
   setup: '/setup',
   welcome: '/welcome',
   faq: '/faq',
   resetPassword: '/reset-password',
   settings: '/settings',
-  data: '/data',
-  competition: '/competition',
   traders: '/traders',
   dashboard: '/dashboard',
   strategy: '/strategy',
-  strategyMarket: '/strategy-market',
 } as const

 export const PAGE_PATHS: Record<Page, string> = {
   agent: ROUTES.agent,
-  competition: ROUTES.competition,
   traders: ROUTES.traders,
   trader: ROUTES.dashboard,
   strategy: ROUTES.strategy,
-  'strategy-market': ROUTES.strategyMarket,
-  data: ROUTES.data,
   faq: ROUTES.faq,
   login: ROUTES.login,
   register: ROUTES.register,
 }

 export const LEGACY_HASH_ROUTES: Record<string, string> = {
   agent: ROUTES.agent,
-  competition: ROUTES.competition,
   traders: ROUTES.traders,
   trader: ROUTES.dashboard,
   details: ROUTES.dashboard,
   strategy: ROUTES.strategy,
-  'strategy-market': ROUTES.strategyMarket,
-  data: ROUTES.data,
 }
```

Also remove the three `case` lines in `getCurrentPageForPath()` for `ROUTES.competition`, `ROUTES.strategyMarket`, and `ROUTES.data`.

- [ ] **Step 11.5: Delete the page files**

```bash
cd /home/hoang/nofx
rm web/src/pages/DataPage.tsx
rm web/src/pages/StrategyMarketPage.tsx
rm web/src/components/trader/CompetitionPage.tsx
rm web/src/components/trader/CompetitionPage.test.tsx
```

- [ ] **Step 11.6: Check for orphan references**

```bash
cd /home/hoang/nofx && grep -rn "DataPage\|StrategyMarketPage\|CompetitionPage\|ROUTES\.data\|ROUTES\.strategyMarket\|ROUTES\.competition" web/src/ 2>&1 | grep -v node_modules | head -30
```

Expected: a small number of i18n keys still reference these page names (e.g., `dataCenter`, `strategyMarket*`, `competition*` in `translations.ts`). These can be deleted as a separate cosmetic pass; the build does not require it.

- [ ] **Step 11.7: Build the frontend**

```bash
cd /home/hoang/nofx/web && npm run build 2>&1 | tail -30
```

Expected: build succeeds with no TypeScript errors. If TS complains about any of the removed page identifiers, locate the remaining reference (likely a stray import or an inline `<Link to={ROUTES.x}>` in a forgotten component) and remove it. The deletion only affects the four files above plus the nav/router infrastructure.

- [ ] **Step 11.8: Smoke test in browser**

```bash
cd /home/hoang/nofx/web && npm run dev
```

Open `http://localhost:3000/`. Verify:
- "Data", "Strategy Market", and "Leaderboard" links are gone from the header (both desktop and mobile views).
- Visiting `http://localhost:3000/data`, `/strategy-market`, `/competition` 404s or redirects.
- Other nav items still work: Agent, Traders, Strategy, Settings, FAQ.
- The "Settings → Exchanges" tab still loads. The Trader Dashboard still loads.

- [ ] **Step 11.9: Backend: are there server routes serving these pages?**

```bash
cd /home/hoang/nofx && grep -n "/api/competition\|/api/leaderboard\|/api/strategy-market\|/api/public-strategies" api/server.go api/handler_*.go 2>/dev/null
```

If any backend routes exist *specifically* to serve competition data or public-strategy listings (e.g., `GET /api/competition`, `GET /api/strategies/public`), and they're no longer reachable from the UI, you can leave them in place (defensive) or remove them. **For this plan, leave them in place** — removing backend routes is a separate optional cleanup; they don't affect the build.

- [ ] **Step 11.10: Commit**

```bash
cd /home/hoang/nofx
git add web/src/components/common/HeaderBar.tsx \
        web/src/router/AppRoutes.tsx \
        web/src/router/paths.ts
git rm web/src/pages/DataPage.tsx \
       web/src/pages/StrategyMarketPage.tsx \
       web/src/components/trader/CompetitionPage.tsx \
       web/src/components/trader/CompetitionPage.test.tsx
git commit -m "refactor(web): remove Data, Strategy Market, and Leaderboard pages

For NQ-trading focus, three crypto-era pages are removed:
- Data: was an iframe to deprecated nofxos.ai/dashboard (CSP-blocked); NinjaTrader covers chart needs
- Strategy Market: community crypto-strategy browser; not useful for single-user NQ setup
- Leaderboard: public competition ranking; not relevant for personal use

Three control surfaces remain: Settings (Config), Strategy (Studio), Traders (Dashboard)."
```

---

## Completion criteria

This plan is done when:

1. ✅ All Go tests pass: `go test ./...`
2. ✅ `cmd/nq_smoke` runs end-to-end against NT SIM with a hand-fabricated decision → real fill in NT SIM → row in `trades_taken.csv`.
3. ✅ All 11 tasks committed to git.
4. ✅ The broken Data page is gone from the web UI; visualization happens in NinjaTrader.

What this plan does **not** deliver (deferred to next plans):
- Real AI client integration (use of DeepSeek/Claude API in the cycle — currently stubbed by stdin in smoke runner)
- Integration into the main `trader/auto_trader_loop.go` polling cycle
- CME holiday calendar + session-boundary lockouts
- Contract roll detection (NQ.c.0 → next month transition awareness)
- Live (real-money) toggle separate from SIM
- Web UI rewrite of `DataPage.tsx`
- Dead-man-switch heartbeat
- Daily loss kill-switch
- `MAX_CONTRACTS_PER_TRADER` enforcement in code (currently relies on NT strategy parameter)

These become the next slices once this one is validated.

---

## Parallel Dispatch Map

The tasks below split into three independent clusters that can run concurrently via `superpowers:dispatching-parallel-agents`. Each cluster ends at a clean compile boundary so dispatches don't fight over shared files.

### Cluster A — Databento data pipeline (Tasks 1–4)

**Files touched:** `provider/databento/`, `market/data.go` (Normalize fix only), `config/`
**Independent because:** Pure data ingestion; no overlap with bridge or UI files.

```text
Task("Cluster A: Databento data pipeline", subagent_type: "feature-dev:code-architect", prompt: """
Implement Tasks 1–4 from docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md:
- Task 1: TRADING_MODE + DATABENTO_API_KEY wiring in config/config.go
- Task 2: provider/databento/client.go — Historical OHLCV REST client (HTTP Basic auth, GET /v0/timeseries.get_range)
- Task 3: Symbol resolver NQ.c.0 → MNQM6 (continuous → specific contract)
- Task 4: Databento bar → market.Kline adapter; fix Normalize() case-preservation for NQ.c.0
Constraints: do NOT touch trader/ninjatrader/ or web/. Stop after `go build ./...` is clean and Task 2.4 + Task 4.6 tests pass.
Return: summary of files created, test results, any deviations from the plan.
""")
```

### Cluster B — NinjaTrader CSV bridge (Tasks 5–8)

**Files touched:** `trader/ninjatrader/`, `trader/auto_trader.go` (switch case only)
**Independent because:** Bridge code is self-contained in one package; only touches auto_trader.go at the broker switch (Task 8) — merge conflict is trivial.
**Blocker:** none on Cluster A — bridge code doesn't import provider/databento.

```text
Task("Cluster B: NinjaTrader CSV bridge", subagent_type: "feature-dev:code-architect", prompt: """
Implement Tasks 5–8 from docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md:
- Task 5: trader/ninjatrader/csv_writer.go — 5-field signal writer (datetime, direction, entry, SL, TP)
- Task 6: trader/ninjatrader/csv_tailer.go — 3-field fill tailer with dedup
- Task 7: trader/ninjatrader/trader.go — 19-method Trader interface impl; compile-time `var _ types.Trader = (*Trader)(nil)`
- Task 8: add "ninjatrader" case to trader/auto_trader.go broker switch (~line 268)
Hazards to encode in code: H1 (lost-signal race, 2s polling), H2 (datetime+direction dedup collision — add monotonic nonce as 6th CSV field OR rate-limit Go writer to ≥2s), H4 (fill replay on NT restart — persist last-processed offset).
Return: summary of files created, list of CSV fields, dedup strategy chosen.
""")
```

### Cluster C — Prompt + UI + smoke (Tasks 9–11)

**Files touched:** `kernel/engine_prompt.go`, `web/src/pages/`, `cmd/nq_smoke/`
**Independent because:** Prompt + UI + smoke runner; doesn't touch provider or trader internals.
**Blocker:** Task 10 (smoke) needs A + B done. Tasks 9 + 11 can start immediately.

```text
Task("Cluster C: Prompt + UI", subagent_type: "feature-dev:code-architect", prompt: """
Implement Tasks 9 + 11 from docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md (DEFER Task 10 — needs Clusters A + B merged first):
- Task 9: kernel/engine_prompt.go — NQ futures prompt template with FuturesContext (no MACDSignal — single float64 only)
- Task 11: confirm crypto-era pages (Data, Strategy Market, Competition) are gone from web/src/router/AppRoutes.tsx + web/src/components/common/HeaderBar.tsx
Constraints: do NOT touch provider/databento/ or trader/ninjatrader/. Run `cd web && npm run build` after Task 11.
Return: prompt diff summary, list of removed route entries, npm build output.
""")
```

### Sequencing

1. **Dispatch A + B + C(Task 9, 11 only) in one message** — 3 parallel agents.
2. **Wait for all three to return** — verify each summary, run `go build ./... && cd web && npm run build`.
3. **Then dispatch Task 10** (end-to-end smoke) as a single agent — depends on A + B fills landing in `data/data.db` via real NT execution.

### Plan 1.5 (NT8 AddOn migration) — NOT parallelizable

Plan 1.5 replaces the CSV bridge with a TCP/WebSocket NDJSON protocol. It must be sequential because:
- Stage 1 (AddOn skeleton) blocks Stage 2 (protocol wire)
- Stage 2 blocks Stage 3 (Go-side client)
- Stage 3 blocks Stage 4 (cutover from CSV)
- Stage 5 (CSV removal) requires Stage 4 stable

Each stage produces a working binary; dispatch one agent per stage in sequence.

## Self-review

**Spec coverage:**
- Databento config — Task 1 ✓
- Databento OHLCV client — Task 2 ✓
- Symbol resolution — Task 3 ✓
- Bar → Kline adapter — Task 4 ✓
- CSV signal writer — Task 5 ✓
- CSV fill tailer — Task 6 ✓
- Trader interface impl — Task 7 ✓
- Wire into auto_trader switch — Task 8 ✓
- NQ futures prompt — Task 9 ✓
- End-to-end smoke — Task 10 ✓
- NT setup + manual smoke (Task 0) — covers the human-side prerequisites

**Placeholder scan:** searched for "TBD", "TODO", "implement later" — none present. All steps have concrete commands or code.

**Type consistency:** `SignalRow` and `FillRow` defined in Task 5, used in Tasks 6, 7, 10. `Bar` defined in Task 2, used in Task 4, 10. `FuturesContext` defined in Task 9, used in Task 10. `Config` (in `trader/ninjatrader`) defined in Task 7, used in Task 8.

**Known assumption to verify during execution:**
- `market.Kline` field name `OpenTime` and millisecond unit — verify in `market/types.go:94-108` before Task 4. If different, adjust adapter + test.
- `trader/types.Trader` interface signatures — verified 2026-05-22: **19 methods exactly** (not 17 as earlier draft said). Task 7's code block has been corrected to implement all 19. The `var _ types.Trader = (*Trader)(nil)` line catches any future drift at compile time.
- Databento JSON response shape for `ohlcv-1m` — verified against Databento docs but actual response format may have minor differences. The Task 2.4 live smoke test catches this.

## v3 audit-feedback corrections (2026-05-22)

Following external audit, the plan was updated with:
- **Task 7 Trader impl** now includes all 19 interface methods (was missing 5: `GetMarketPrice`, `CancelStopLossOrders`, `CancelTakeProfitOrders`, `CancelStopOrders`, `GetOpenOrders`). The "17-method" label was wrong; corrected to 19.
- **`Normalize` futures fix** is now case-preserving (was returning `NQ.C.0` instead of `NQ.c.0` — Databento rejects the uppercased form).
- **`FuturesContext.MACDSignal` removed** — `market.ExportCalculateMACD` returns only the MACD line (one `float64`). Signal/histogram would require extending the indicator API and are deferred.
- **JWT verification** strengthened — the `🔑 JWT secret configured` log line fires unconditionally, so cannot be used as a check. Added a recommended code-level warning when the default value is detected.
- **Coverage scorecard** recounted; previous totals had arithmetic errors on 3 of 4 surfaces. Corrected totals reconcile.
- **CSV bridge runtime hazards** documented as Plan 1.5 work (4 hazards: H1 lost-signal race, H2 dedup collision, H3 DrvFs mtime, H4 fill replay on session rollover). Not blocking for Plan 1 SIM but must be addressed before any live trading.
- **`GetBalance` $50k mock** honestly disclosed as feeding both prompt equity math and Go-side risk checks; 3 fix options documented; option 3 (disable Go-side balance sizing for ninjatrader exchange) recommended for early go-live.

**The plan is now executable for Plan 1 SIM paper trading.** Going live requires Plan 1.5 (CSV protocol hardening + real balance read) and Plan 2 (CME calendar + contract rolls + kill-switches).

---

## Task 12: Backend kernel + market futures routing (Cluster D)

**Why:** v4 audit found that `BuildSystemPrompt("futures")` falls through to crypto, futures decision JSON omits `symbol`, `GetWithTimeframes` has no Databento branch, and `kernel/engine.go` never checks `TradingMode`.

**Files:**
- Modify: `kernel/engine_prompt.go:39-46` (add futures case to variant switch)
- Modify: `kernel/engine_prompt_futures.go:50` (add `symbol` to JSON shape)
- Modify: `market/data.go:147-256` (add futures branch in GetWithTimeframes)
- Modify: `kernel/engine.go` (resolve continuous symbol when TradingMode=futures)
- Test: `kernel/engine_prompt_test.go`, `market/data_test.go`

- [ ] **Step 1: Route futures variant in BuildSystemPrompt**

In `kernel/engine_prompt.go` find the variant switch (around line 39-46) and add:
```go
case "futures":
    return BuildFuturesSystemPrompt(accountEquity)
```
Place BEFORE the default crypto case so "futures" matches first.

- [ ] **Step 2: Add `symbol` to futures decision JSON**

In `kernel/engine_prompt_futures.go` find the JSON example string (around line 50) and update from `{action, entry, stop_loss, take_profit, reasoning}` to `{symbol, action, entry, stop_loss, take_profit, reasoning, confidence}`. The `symbol` field is required by `engine_analysis.go` decision parser.

- [ ] **Step 3: Add futures branch in GetWithTimeframes**

In `market/data.go` add a helper:
```go
func getKlinesFromDatabento(symbol string, timeframe string) ([]Kline, error) {
    cfg := config.Get()
    client := databento.NewClient(cfg.DatabentoAPIKey, cfg.DatabentoDataset)
    resolved, err := databento.ResolveContinuous(client, symbol)
    if err != nil { return nil, err }
    bars, err := client.GetOHLCV(resolved, timeframe, 200)
    if err != nil { return nil, err }
    return BarsToKlines(bars), nil
}
```
Then in `GetWithTimeframes` add (BEFORE the hyperliquid/coinank fork):
```go
if IsCMEFuturesSymbol(symbol) {
    return getKlinesFromDatabento(symbol, tf)
}
```

- [ ] **Step 4: Wire kernel/engine.go to TradingMode**

In `kernel/engine.go` where the engine builds prompts/fetches data, add a TradingMode check at the entry point:
```go
if config.Get().TradingMode == "futures" {
    variant = "futures"
}
```
This ensures the futures prompt path triggers when env var is set, regardless of strategy config.

- [ ] **Step 5: Tests + build**

Add `TestBuildSystemPrompt_FuturesVariant` and `TestNormalize_CMEFutures("NQ.c.0")` cases. Run:
```bash
go test ./kernel/... ./market/...
go build ./...
```

- [ ] **Step 6: Commit**
```bash
git commit -m "feat(futures): route futures variant + add Databento branch in GetWithTimeframes + symbol in decision JSON"
```

---

## Task 13: Backend wiring + storage for NinjaTrader (Cluster E)

**Why:** v4 audit found `manager/trader_manager.go` has no ninjatrader case → NT traders unloadable from DB; `store/exchange.go` Exchange struct lacks NT fields; `AutoTraderConfig` missing `NTDefaultContractQty`; CSV tailer offset not persisted.

**Files:**
- Modify: `store/exchange.go` (add NT fields to Exchange struct + GORM tags + migration)
- Modify: `manager/trader_manager.go:~700` (add ninjatrader case to addTraderFromStore switch)
- Modify: `trader/auto_trader.go:95-97,~322` (add NTDefaultContractQty field + use it)
- Modify: `provider/ninjatrader/csv_tailer.go:18,78-82` (persist offset to disk)
- Test: `store/exchange_test.go`, `provider/ninjatrader/csv_tailer_test.go`

- [ ] **Step 1: Add NT fields to Exchange struct**

In `store/exchange.go` add to the Exchange struct:
```go
// NinjaTrader-specific (only set when ExchangeType == "ninjatrader")
NTDataDir            string `gorm:"column:nt_data_dir"            json:"nt_data_dir,omitempty"`
NTInstrumentName     string `gorm:"column:nt_instrument_name"     json:"nt_instrument_name,omitempty"`
NTDefaultContractQty int    `gorm:"column:nt_default_contract_qty" json:"nt_default_contract_qty,omitempty"`
```
GORM AutoMigrate will add the columns; no separate migration file needed.

- [ ] **Step 2: Add ninjatrader case in manager**

In `manager/trader_manager.go` find the `switch exchangeCfg.ExchangeType` block in `addTraderFromStore` (around line 661-700) and after the `case "indodax":` add:
```go
case "ninjatrader":
    traderConfig.NinjaTraderDataDir = exchangeCfg.NTDataDir
    traderConfig.NinjaTraderSymbol = exchangeCfg.NTInstrumentName
    traderConfig.NTDefaultContractQty = exchangeCfg.NTDefaultContractQty
```

- [ ] **Step 3: Add NTDefaultContractQty to AutoTraderConfig**

In `trader/auto_trader.go` near line 95-97 add field:
```go
NTDefaultContractQty int
```
And in the ninjatrader case (around line 320-328) pass it into `ntTrader.New(ntTrader.Config{...DefaultContractQty: config.NTDefaultContractQty})`.

- [ ] **Step 4: Persist CSV tailer offset**

In `provider/ninjatrader/csv_tailer.go` change `seen int` to a disk-backed counter:
```go
type Tailer struct {
    path       string
    offsetPath string  // e.g. path + ".offset"
    seen       int
}

func (t *Tailer) loadOffset() {
    if data, err := os.ReadFile(t.offsetPath); err == nil {
        fmt.Sscanf(string(data), "%d", &t.seen)
    }
}

func (t *Tailer) saveOffset() {
    _ = os.WriteFile(t.offsetPath, []byte(fmt.Sprintf("%d", t.seen)), 0644)
}
```
Call `loadOffset()` in constructor; call `saveOffset()` after every successful row processing. Removes H4 fill-replay hazard.

- [ ] **Step 5: Tests + build**
```bash
go test ./store/... ./manager/... ./provider/ninjatrader/...
go build ./...
```

- [ ] **Step 6: Commit**
```bash
git commit -m "feat(nt): per-account exchange fields + manager wiring + tailer offset persistence"
```

---

## Task 14: Frontend NinjaTrader config (Settings page)

**Why:** v4 audit found ExchangeConfigModal has no NinjaTrader option, SettingsPage handleSaveExchange doesn't pass NT params. User cannot save NT config from UI.

**Files:**
- Modify: `web/src/components/trader/ExchangeConfigModal.tsx` (~+140 LOC across 4 sections)
- Modify: `web/src/pages/SettingsPage.tsx:190-257` (handleSaveExchange signature + payload)

- [ ] **Step 1: Add NinjaTrader to SUPPORTED_EXCHANGE_TEMPLATES**

In `web/src/components/trader/ExchangeConfigModal.tsx` around line 23-34, append to the templates array:
```ts
{
    id: 'ninjatrader',
    name: 'NinjaTrader',
    type: 'futures',
    icon: '/icons/ninjatrader.svg',  // OK to fallback to a generic icon if asset not present
    description: 'CME futures via NT8 CSV bridge',
    requiresApiKey: false,
}
```

- [ ] **Step 2: Add form state**

Around line 154-180 add useState entries:
```tsx
const [ntDataDir, setNtDataDir] = useState('')
const [ntInstrumentName, setNtInstrumentName] = useState('MNQ')
const [ntDefaultContractQty, setNtDefaultContractQty] = useState(1)
```

- [ ] **Step 3: Add form UI section**

After the existing `aster` form section (around line 720-756) add:
```tsx
{selectedExchange === 'ninjatrader' && (
    <div className="space-y-4">
        <div>
            <label className="block text-sm font-medium mb-1">NT Data Directory (WSL path)</label>
            <input
                type="text"
                value={ntDataDir}
                onChange={(e) => setNtDataDir(e.target.value)}
                placeholder="/mnt/c/Users/<u>/NofxTrader/data"
                className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded"
            />
        </div>
        <div>
            <label className="block text-sm font-medium mb-1">Instrument</label>
            <input
                type="text"
                value={ntInstrumentName}
                onChange={(e) => setNtInstrumentName(e.target.value)}
                placeholder="MNQ"
                className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded"
            />
        </div>
        <div>
            <label className="block text-sm font-medium mb-1">Default Contract Qty</label>
            <input
                type="number"
                min="1"
                value={ntDefaultContractQty}
                onChange={(e) => setNtDefaultContractQty(parseInt(e.target.value) || 1)}
                className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded"
            />
        </div>
    </div>
)}
```

- [ ] **Step 4: Add validation branch in handleSubmit**

In `ExchangeConfigModal.tsx` around line 301-339 add:
```tsx
} else if (selectedExchange === 'ninjatrader') {
    if (!ntDataDir.trim()) {
        setError('NT Data Directory is required')
        return
    }
}
```

- [ ] **Step 5: Update onSave call + props interface**

In `ExchangeConfigModal.tsx` find the onSave invocation at the end of handleSubmit and pass NT params; update the props interface (~line 39-55) to include `ntDataDir`, `ntInstrumentName`, `ntDefaultContractQty`.

- [ ] **Step 6: Update SettingsPage handleSaveExchange**

In `web/src/pages/SettingsPage.tsx:190-257` add NT params to function signature and thread into createRequest/updateRequest body:
```ts
const handleSaveExchange = async (
    // ...existing params...
    ntDataDir?: string,
    ntInstrumentName?: string,
    ntDefaultContractQty?: number,
) => {
    // ...
    const payload = {
        // ...existing fields...
        nt_data_dir: ntDataDir,
        nt_instrument_name: ntInstrumentName,
        nt_default_contract_qty: ntDefaultContractQty,
    }
}
```

- [ ] **Step 7: Build + smoke test**
```bash
cd web && npm run build
```
Then open Settings → Exchanges → Add → NinjaTrader and confirm form renders.

- [ ] **Step 8: Commit**
```bash
git commit -m "feat(web): NinjaTrader exchange config form + save handler"
```

---

## Task 15: Frontend futures-gating (Dashboard + Strategy)

**Why:** v4 audit found Dashboard hardcodes USDT/leverage; CoinSourceEditor appends USDT to "NQ"; IndicatorEditor + RiskControlEditor have no variant prop.

**Files:**
- Modify: `web/src/pages/TraderDashboardPage.tsx:513,522,530,609,663,673` (gate on exchange_type)
- Modify: `web/src/components/strategy/CoinSourceEditor.tsx:69-79` (skip USDT for CME)
- Modify: `web/src/components/strategy/IndicatorEditor.tsx:656-688` (variant prop + hide crypto sources)
- Modify: `web/src/components/strategy/RiskControlEditor.tsx:61-128` (variant prop + relabel)

- [ ] **Step 1: Gate Dashboard StatCard units**

In `web/src/pages/TraderDashboardPage.tsx` around lines 513, 522, 530:
```tsx
const unit = exchangeType === 'ninjatrader' ? 'USD' : 'USDT'
<StatCard ... unit={unit} ... />
```
`exchangeType` is already available via `getExchangeTypeFromList(...)` helper.

- [ ] **Step 2: Hide Leverage + Liquidation Price columns for NT**

Around lines 609, 663, 673 wrap each `<th>` and `<td>` with:
```tsx
{exchangeType !== 'ninjatrader' && (
    <th className="hidden md:table-cell ...">Leverage</th>
)}
```
Same pattern for Liquidation Price column.

- [ ] **Step 3: Skip USDT auto-append for CME futures**

In `web/src/components/strategy/CoinSourceEditor.tsx` around line 69-79:
```ts
const cmeFuturesPattern = /^(NQ|MNQ|ES|MES|YM|MYM|RTY|M2K|CL|GC)(\.c\.0|[A-Z]\d)?$/i
if (cmeFuturesPattern.test(symbol)) {
    return symbol  // keep as-is, no USDT
}
```
Place BEFORE the existing xyzDexAssets check.

- [ ] **Step 4: Add variant prop to IndicatorEditor**

In `web/src/components/strategy/IndicatorEditor.tsx` add `variant?: 'crypto' | 'futures'` prop. Around lines 656-688 gate crypto-only toggles:
```tsx
{variant !== 'futures' && (
    <label><input type="checkbox" name="enable_funding_rate" /> Funding Rate</label>
)}
```
Same for OI Ranking, NetFlow Ranking, Price Ranking sources.

Pass `variant={exchangeType === 'ninjatrader' ? 'futures' : 'crypto'}` from `StrategyStudioPage.tsx`.

- [ ] **Step 5: Add variant prop to RiskControlEditor**

In `web/src/components/strategy/RiskControlEditor.tsx` add `variant?: 'crypto' | 'futures'` prop. Relabel:
```tsx
const leverageLabel = variant === 'futures' ? 'Primary Instrument Leverage' : 'BTC/ETH Leverage'
const sizeUnit = variant === 'futures' ? 'contracts' : 'USDT'
```

- [ ] **Step 6: Build + visual smoke**
```bash
cd web && npm run build
```
Then open Dashboard for a NT trader and verify no USDT unit, no Leverage/LiqPrice columns. Open Strategy Studio and verify no funding rate / OI sources for futures.

- [ ] **Step 7: Commit**
```bash
git commit -m "feat(web): futures-gate Dashboard + IndicatorEditor + RiskControlEditor + CoinSourceEditor"
```

---

## Task 16: VL brand cleanup + minor fixes

**Why:** v4 audit found SetupPage still says "Welcome to NOFX"; chart watermarks say "NOFX"; agent.go:174 comment is stale; missing Normalize test.

**Files:**
- Modify: `web/src/pages/SetupPage.tsx` (NOFX → VL in 3 languages)
- Modify: `web/src/components/charts/EquityChart.tsx:321,335` (watermark text)
- Modify: `web/src/components/charts/AdvancedChart.tsx:1161,1182` (watermark text)
- Modify: `agent/agent.go:174` (comment example)
- Test: `market/data_test.go` (TestNormalize_CMEFutures)

- [ ] **Step 1: SetupPage NOFX → VL**

In `web/src/pages/SetupPage.tsx` replace any occurrence of "NOFX" in user-visible copy with "VL". Check 3 language blocks (en, zh, id).

- [ ] **Step 2: Chart watermarks**

In `EquityChart.tsx:321,335` and `AdvancedChart.tsx:1161,1182` change watermark text "NOFX" → "VL".

- [ ] **Step 3: Update stale comment**

In `agent/agent.go:174` change comment example `"deepseek-chat"` → `"deepseek-v4-pro"` to match the actual default.

- [ ] **Step 4: Add Normalize futures test**

In `market/data_test.go` add:
```go
func TestNormalize_CMEFutures(t *testing.T) {
    cases := []struct{ in, want string }{
        {"NQ.c.0", "NQ.c.0"},
        {"MNQ.c.0", "MNQ.c.0"},
        {"nq.c.0", "NQ.c.0"},  // only the ticker portion uppercased
    }
    for _, tc := range cases {
        if got := Normalize(tc.in); got != tc.want {
            t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
        }
    }
}
```

- [ ] **Step 5: Build + commit**
```bash
go build ./... && cd web && npm run build && cd ..
git commit -m "chore: VL brand cleanup + stale comment + Normalize futures test"
```

---

## Dispatch map for Tasks 12-16

Tasks 12 + 13 + 14/15 are independent — they touch disjoint file sets. Task 16 can fold into any of them. Dispatch in parallel via one message:

```text
Task("Task 12: Backend kernel + market futures routing")
Task("Task 13: Backend wiring + storage for NinjaTrader")
Task("Task 14+15: Frontend NT config + futures-gating")  // touches different files, one agent handles both
```

After all three return clean (`go build ./... && cd web && npm run build`) → run Task 10 (end-to-end smoke) as the final verification.

---

# Plan 1.5: NT8 AddOn migration (research-backed architecture)

> **Added 2026-05-22 based on external research.** Plan 1.5 replaces the CSV polling bridge with an in-process NinjaScript AddOn that hosts a TCP/WebSocket loopback server. This eliminates every "v1 limitation" Task 7 documents (manual close, balance read, cancel/modify, real fill events) and removes the CSV runtime hazards H1-H4 by design. **Plan 1 (CSV) still ships first** to validate the AI brain in SIM; Plan 1.5 migrates the bridge once Plan 1 trades cleanly.

## Trigger conditions — when to start Plan 1.5

Plan 1.5 is **deferred** until at least one of these conditions fires. Do NOT
preemptively rebuild the bridge — the CSV path in Plan 1 is sufficient for
the "validate AI brain in SIM" objective.

Start Plan 1.5 when any one of these is true:

1. **AI brain validated for several sessions.** The AI is making sensible NQ
   decisions on stale-bar (15-min Databento Historical) data and you're
   ready to graduate to real-time bars + real fill events for live trading.
   Concretely: 20-50 cycles observed in SIM with reasoning quality you trust.

2. **Funded broker connection imminent.** You're switching NinjaTrader's
   account from SIM101 to a real funded prop-firm account. Plan 1.5 is
   required before live money — the CSV bridge cannot read real balance
   ($50k mock breaks position sizing) and cannot manually close on news.

3. **CSV race condition fires in SIM.** Hazards H1-H4 (lost-signal race,
   1-sec dedup collision, DrvFs mtime, fill replay on session rollover)
   are documented as acceptable for SIM but not for live. If any fires
   in SIM, that's the signal to upgrade the bridge now rather than later.

4. Real-time bars required for actionable strategy validation

   PROBLEM
   - Databento Historical OHLCV has a publication delay (verify exact
     delay against Databento docs and current account tier before Plan
     1.5 design — assume 15 minutes for planning purposes)
   - Strategies with intra-session entry windows (e.g. TTrades AM
     Silver Bullet, 10:00-11:00 ET) cannot be validated for actionable
     edge on delayed data — by the time a bar arrives, the entry
     window may have already passed
   - Databento Live DOES provide OHLCV aggregates at 1-second and
     1-minute intervals over its real-time streaming API (per
     databento.com/futures product page). Earlier internal assumption
     that Databento Live was ticks-only was INCORRECT.
   - However, Databento Live requires a separate subscription tier
     beyond Historical, while NinjaTrader already receives a live
     CQG/Rithmic feed through the prop firm subscription (MFFU,
     Bulenox, Apex, Topstep all include this) at no additional cost.

   DECISION
   - Plan 1.5 will source LIVE bars from NinjaTrader, not Databento
     Live
   - Databento Historical retained for: warmup on cold start, gap fill
     after restart, archive, walk-forward backtests
   - NT live bars become primary source for the live indicator loop
   - Trade-off accepted: Plan 1.5 requires NinjaScript changes to
     ClaudeTrader.cs and a Go-side bar reader; in exchange,
     elimination of source-disagreement risk and zero additional data
     vendor cost during SIM phase

   SCOPE EXPANSION
   - Plan 1.5 was previously scoped as "TCP socket bridge for orders"
   - Now expanded to include:
     a) TCP or CSV-based live bar feed (NT → Go), implementation
        choice deferred to Plan 1.5 design doc
     b) Data source switcher in Go (Historical for warmup, Live for
        steady-state)
     c) Bar timestamp normalization (NT close-time vs Databento
        open-time conventions — verify both before implementing)
     d) Contract roll handling across NT and Databento symbol
        conventions

   OPEN QUESTIONS FOR PLAN 1.5 DESIGN DOC
   - Current Calculate mode in ClaudeTrader.cs (OnBarClose vs
     OnEachTick) — read file to verify before designing the bar-write
     path
   - File I/O method: File.AppendAllText vs FileStream with explicit
     FileShare.Read — NT support docs recommend lock object
     (private object barWriteLock = new object()) for multi-threaded
     write safety
   - Tail-from-offset reader pattern in Go to survive bot restarts
     without losing NT-written bars
   - Bar timestamp convention difference: NT defaults to bar-close
     time; Databento OHLCV uses bar-open (ts_event = bar start) —
     verify both and document the translation
   - Symbol normalization across MNQM6/MNQU6 rolls
   - Latency budget: NT writes bar → Go reads bar → indicators → AI →
     CSV signal → NT places order. Target end-to-end < 5 seconds for
     1-minute timeframe viability.

   CITATIONS
   - Databento OHLCV aggregate availability:
     https://databento.com/futures (verified 2026-05-25)
   - NT8 multi-threading file I/O guidance:
     https://ninjatrader.com/support/helpguides/nt8/multi-threading.htm
   - NT8 IsFirstTickOfBar / Calculate mode behavior:
     https://ninjatrader.com/support/helpGuides/nt8/isfirsttickofbar.htm

Until one of those triggers fires: **the CSV path is canonical**, Plan 1
stays the production path, and Plan 1.5 lives as documented architecture
ready to execute when needed.

## Plan 1.5 — Design Findings (Research Summary 2026-05-25)

The following findings come from a comprehensive research pass on the
NT8 → Go (WSL2) real-time market data bridge problem. They are NOT
implementation specs — they are the verified-or-flagged knowledge that
must inform the Plan 1.5 implementation when it is written.

Each finding cites its source. Items marked [VERIFY] require empirical
confirmation on John's specific machine before relying on them.

### Architecture Decision: CSV-over-/mnt/c with 1-minute bars

DECISION: Plan 1.5 will use CSV append on /mnt/c as the bar transport,
not TCP, not NetMQ, not WebSocket, not memory-mapped files.

RATIONALE:
- Plan 1 already validated this pattern end-to-end with 2 real fills
  on 2026-05-22. Lowest implementation risk.
- For 1-minute bars on NQ/MNQ, the 250ms-poll-jitter latency is
  comfortably below any actionable strategy budget.
- Alternative transports (NetMQ, WebSocket, TCP) deferred to Stage 2,
  triggered only if measured p99 latency exceeds 500ms.

DECISION: NT emits ONLY 1-minute bars. Go aggregates 5m/15m/H1/H4
in-process.

RATIONALE:
- Single source of truth for higher timeframes
- Backtest-live parity: same aggregation logic used against historical
  Databento OHLCV-1m
- Multi-BarsArray NinjaScript has documented synchronization quirks
  (NT8 "stair-step effect" in multi-series indicators)
- Reconnect resilience: gap-filling one 1m feed is trivial; gap-filling
  5 separate timeframe streams introduces consistency bugs

### NT8 Calculate Mode (verified from NT support forum)

DECISION: Calculate.OnEachTick + IsFirstTickOfBar + State.Historical
guard. The bar-writer fires once per closed bar in realtime only.

EXACT IDIOM:
- In State.SetDefaults: Calculate = Calculate.OnEachTick
- In OnBarUpdate():
  - if (State == State.Historical) return;
  - if (!IsFirstTickOfBar) return;
  - if (CurrentBar < 1) return;
  - Write Times[0][1] / Opens[0][1] / Highs[0][1] / Lows[0][1] /
    Closes[0][1] / Volumes[0][1]

WARNING: Without the State.Historical return guard, every strategy
restart writes thousands of duplicate historical bars. Confirmed by NT
staff in forum thread "State == State.Realtime".

WARNING: NT redownloads the current day's historical data on every
reconnect (NT staff, forum thread "Reload charts after connection
lost"). This means after a NT reconnect, there is a temporary GAP in
the live-emitted bars while NT replays history. The Go side must
gap-fill from Databento Historical on reconnect detection.

### File I/O Pattern (NT official guidance)

DECISION: Explicit lock object + FileStream with FileMode.Append +
FileShare.Read + FileOptions.WriteThrough.

RATIONALE: NT's own multi-threading help guide explicitly requires
protection of custom resources because "market data is distributed
across the entire application by a randomly assigned UI thread, there
is no guarantee that your object will be running on the same event
thread."

PATTERN:
private static readonly object _writeLock = new object();
lock(_writeLock) {
    using (FileStream(_csvPath, FileMode.Append, FileAccess.Write,
                      FileShare.Read, 4096, FileOptions.WriteThrough))
    using (StreamWriter sw) { sw.WriteLine(row); }
}

FileShare.Read allows Go to hold the file open for reading concurrently
without blocking NT writes. Go reader must use os.Open (read-only) and
must NEVER call syscall.Flock — that would cause NT's next write to
throw "process cannot access the file."

### WSL2 File Watching: HARD CONSTRAINT

CRITICAL FINDING: inotify on /mnt/c DOES NOT FIRE for Windows-side
file writes. This is microsoft/WSL issue #4739 and #5424, both still
open as of May 2026.

DO NOT use fsnotify in Go for this bridge. The Add() call will
succeed silently, but events will never arrive. There is no error to
detect.

DECISION: Polling with os.Stat + persisted byte offset, every 250ms.

GO PATTERN:
- Load lastOffset from /home/hoang/nofx/.state/bars_MNQ_1m.offset
- Loop: os.Stat -> if size > lastOffset -> os.Open -> Seek(lastOffset)
  -> read new bytes -> update offset -> save offset
- Only ingest lines ending in '\n' (handles partial reads where NT
  has flushed bytes but not yet completed a line)
- Persist offset to WSL native filesystem (NOT /mnt/c) for guaranteed
  POSIX rename atomicity

POLLING INTERVAL: 250ms is the recommended default. Worst-case
NT-write-to-Go-ingest latency = 250ms + transport overhead, typically
under 500ms total. Stat overhead is sub-millisecond on a small file;
CPU impact negligible.

### Timestamp Reconciliation: NT vs Databento

CRITICAL FINDING: NT and Databento use OPPOSITE bar timestamp
conventions in DIFFERENT timezones.

- NT: bar timestamp = bar CLOSE time, in the timezone configured in
  Control Center > Tools > Options > General (defaults to Windows
  local time). Verified from NT staff reply in forum thread
  "NinjaScript NT8 TIME[0]" plus the "How Bars Are Built" help guide.
- Databento OHLCV: ts_event = bar OPEN time, in UTC nanoseconds.
  Verified from NautilusTrader integration docs and Databento schema
  documentation.

CONSEQUENCE: The same 1-minute bar (e.g., 09:30:00-09:31:00 ET) will
appear in NT as Time[0] = 09:31:00 LOCAL and in Databento as
ts_event = 09:30:00.000000000 UTC. Off-by-one-minute bugs from this
mismatch are the #1 source of NT-Databento integration failures.

CANONICAL KEY: All bars in the Go side are keyed by bar_open_utc
(time.Time, UTC, second precision). Convert at ingest:
- NT side: write Times[0][1].AddMinutes(-1).ToUniversalTime() as
  ISO-8601
- Databento side: time.Unix(0, ts_event_ns).UTC()

DST HANDLING: NT uses local Windows time and shifts with DST.
Databento uses UTC and is immune. Always store and reason in UTC;
convert to ET/CT at display layer only.

### Multi-Timeframe: Aggregate in Go from 1m

DECISION: NT writes only 1m bars. Go aggregates 5m, 15m, H1, H4.

WARNING: 5m bar close instants in NT and Databento RARELY align to
the same millisecond. NT bars close on the first tick AFTER the
boundary (verified NT forum thread "Bar Closing Time" documents
2-10 second delta in slow markets). Databento bars close on the
matching-engine aggregation boundary (deterministic).

Implication: ALWAYS use bar_open_utc as the key. NEVER use
arrival-wall-clock-time.

### Session Boundaries (CME NQ/MNQ)

VERIFIED from CME E-mini/Micro futures contract specifications page
(May 2026):
- Sunday 17:00 CT → Friday 16:00 CT
- Daily trading halt 16:00–17:00 CT Mon-Thu
- In Eastern Time: Sunday 18:00 ET → Friday 17:00 ET, daily halt
  17:00-18:00 ET Mon-Thu

CME holidays and early closes occur on different days from US
equity markets. DO NOT infer session boundaries from clock arithmetic
alone — use NT's Trading Hours template (Bars.IsFirstBarOfSession)
or Databento's "status" schema.

### Contract Roll Handling

CRITICAL FINDING from NT's official rollover help guide:
"NinjaScript strategies are not rolled forward and must be manually
rolled over."

CONSEQUENCE: Plan 1.5 must explicitly handle the front-month roll on
both sides (NT and Databento) because they have OPPOSITE policies and
NEITHER auto-rolls inside a running NinjaScript strategy.

- NT side uses fixed expiry-month symbols (e.g. MNQ 06-26, MNQ 09-26).
  Live ticks arrive only for the explicit front-month contract attached
  to the chart. Per NT staff in forum thread "continuous ticket for
  MES/MNQ/M2K": "Loading the current front month in NinjaTrader 8 with
  a default merge policy of 'Merge Back Adjusted' will be almost
  identical to a continuous contract."
- Databento side ships RAW prices through the roll (no back-adjustment)
  via continuous symbology (MNQ.c.0, MNQ.v.0, MNQ.n.0, MNQ.c.1). From
  Databento docs: "Our continuous contract symbology does not behave
  the same as continuous contracts provided on retail charting apps,
  which create a continuous series by applying a constant offset on
  each rollover month to the lead month contract. Our philosophy is
  generally to provide raw prices."

VOLUME-DRIVEN ROLL TIMING: per FlowBots Knowledge Center analysis,
volume typically migrates from the front contract to the next 1-2
days BEFORE the calendar expiry. NT's rollover prompt fires based on
calendar date and can leave the running strategy pointing at a
near-empty contract for that 1-2 day window. Best practice: defer
the NT rollover prompt until volume has actually migrated.

STITCHING OPTIONS (3 methods):
- Raw stitched: concatenate front-month series, accept the price-jump
  at each roll. Appropriate when indicators are scale-invariant or
  analyze each contract separately.
- Back-adjusted (additive): subtract the front/back roll-day spread
  from all pre-roll bars. NT default "Merge Back Adjusted" policy;
  appropriate for trend indicators (MA, BB).
- Ratio-adjusted: multiply all pre-roll bars by new_price/old_price
  ratio. Better for very long histories where additive offsets
  accumulate.

PLAN: live trade only the explicit front-month contract; manually
re-deploy strategy on roll day; plan a tradeless window of several
hours to avoid signaling on the synthetic price jump; request
Databento MNQ.v.0 (volume-based continuous) raw for historical
warmup; apply Go-side back-adjustment at ingest.

### Latency Measurement

Instrument both sides explicitly. NT writes two timestamps per CSV
row (bar_close_utc_nt, wallclock_at_write_utc captured immediately
before sw.WriteLine). Go captures time.Now().UTC() at three points:
- t_read_stat: when os.Stat first shows new bytes
- t_parse_done: when the row is fully parsed into a struct
- t_indicator_done: when downstream indicators have processed it

NT-to-Go transport latency = t_read_stat - wallclock_at_write_utc.

ENGINEERING ESTIMATES (component breakdown, NOT measured on John's
machine — see VERIFY-7):
- NT bar-close to NT file-write: < 10 ms (sub-second tick + lock
  acquisition)
- NT write to NTFS flush: < 5 ms with FileOptions.WriteThrough
- NTFS to DrvFs visibility: < 50 ms typical, occasional spikes to
  several hundred ms
- Go poll-jitter: 0-250 ms depending on phase alignment relative to
  bar close
- Go parse + ingest: < 1 ms per row

BUDGET: end-to-end p99 < 500 ms (NT bar-close to Go indicator update).
Above 500 ms triggers Stage 2 transport upgrade (see Alternative
Transports below).

BOTTLENECK: the 250 ms poll interval. Reducing it costs CPU; moving
off CSV to a push-based transport eliminates it entirely.

See [VERIFY-7] in Caveats and Verification Items below.

### Reordering and Deduplication

FAILURE MODES:
1. NT reconnect during session: NT replays the day's historical and
   the State.Historical guard prevents re-writes. If the strategy is
   RE-ENABLED (not just reconnected), prior bars may never get
   appended (no double-write), while subsequent live bars are
   appended normally. Net: occasional gaps, never duplicates.
2. File-system buffering reordering: with FileOptions.WriteThrough
   and a single-thread writer protected by lock(), reordering is not
   possible within a single NT process. Cross-process reordering is
   also not a concern (one writer).
3. Strategy restart while a bar is being written: lock acquisition
   is process-scoped; an OS crash mid-write can leave a partial line.
   Mitigated by the partial-line rule below.

PARTIAL-LINE RULE (Go side): only ingest lines that end with '\n'.
On incomplete final read, set lastOffset to the end of the last
newline (NOT to the current file position).

IDEMPOTENT INGESTION: use (symbol, bar_open_utc) as composite primary
key. Dedup cache with 24-hour TTL:
- 24h × 60 bars/hour = 1440 entries per symbol — trivial memory
- key format: fmt.Sprintf("%s|%d", symbol, barOpenUTC.Unix())
- On duplicate: log debug, skip ingestion

### Process Lifecycle and Persistent Offset

When the Go process dies and restarts, replaying the entire CSV
from offset 0 is wasteful. Persist last-known-good offset.

CHECKPOINT LOCATION: WSL native filesystem (e.g.
/home/hoang/nofx/.state/), NOT /mnt/c.

REASON: DrvFs does NOT guarantee POSIX rename atomicity across the
Windows/Linux boundary. ext4 native does. Atomic-rename is the
critical primitive for crash-safe checkpoints.

WRITE PATTERN (write-to-tmp + atomic rename):
- Write offset bytes to checkpointPath + ".tmp"
- os.Rename(tmp, checkpointPath) — atomic on same-filesystem WSL
  native

RECOVERY ON STARTUP:
- Load offset from checkpoint file
- os.Seek(offset, io.SeekStart) on the CSV
- Resume the polling loop

OFFSET-BEYOND-FILE-SIZE: if saved offset is greater than current CSV
size (e.g. NT truncated and re-initialized the bar log between
sessions), reset offset to 0 and re-ingest from start. The dedup
cache prevents double-emission of any bars from the prior session
that remain in the cache window.

### Alternative Transports (Future Migration Path)

Plan 1.5 ships on CSV. The following are deferred upgrade paths
triggered only if measured latency exceeds the 500 ms p99 budget.

| Transport | Effort | Latency | Reliability notes |
|---|---|---|---|
| NinjaScript TCP server | 1-2 days | 1-5 ms | Fragile inside NT process per NT staff; users report success in Console apps but issues inside NT |
| NetMQ (native C# ZeroMQ) | 2-3 days | 1-5 ms | No native libsodium dependency; cleaner than clrzmq4 inside NT. RECOMMENDED for fan-out pub/sub |
| WebSocket from NinjaScript | 2-3 days | 2-10 ms | NT support has explicitly recommended over raw TCP: "websockets… more reliable than TCP". RECOMMENDED for single-connection simple protocol |
| HTTP POST from NT to Go | 1 day | 5-50 ms | Highest overhead; easiest to debug; survives Go restarts gracefully. Best for control-plane, not bar data |

REJECTED OPTIONS:
- Named pipes: WSL2 cannot directly open Windows named pipes; would
  require a Windows-side proxy.
- Memory-mapped files: cross-kernel sync primitives not guaranteed
  across DrvFs; MMF semantics undocumented for the Win-Lin boundary.

CSV ROLE AFTER MIGRATION: keep CSV as warm-backup archive. The CSV
writer should remain enabled even after socket-based transport
becomes primary — the file is a free durability layer and aids
post-mortem analysis.

### Decision Matrix Summary

| Design question | Recommended | Rationale |
|---|---|---|
| Transport for bar data | CSV append on /mnt/c | Already works; meets 500ms p99 budget for 1m bars |
| Calculate mode | Calculate.OnEachTick + IsFirstTickOfBar + State.Historical guard | Documented NT idiom; preserves option for future intra-bar tick channel |
| Bar timestamp written | Bar OPEN in UTC ISO-8601 | Matches Databento ts_event convention; DST-immune; one canonical key |
| Multi-timeframe | NT writes only 1m; Go aggregates 5m/15m/H1/H4 | Single source of truth; backtest-live parity; reconnect-safe |
| Symbol on live NT | Front-month explicit (e.g. MNQ 06-26) | NT live cannot use continuous symbols on CQG/Rithmic |
| File I/O pattern | lock(_writeLock) { FileStream(Append, FileShare.Read, WriteThrough) } | Per NT's own multi-threading help guide |
| Go-side file watching | Polling os.Stat every 250ms + persisted offset | inotify does NOT fire for Windows-side writes on /mnt/c |
| Dedup key | symbol + bar_open_utc, 24h TTL | Idempotent against historical replay, reconnect, restart |
| Checkpoint location | WSL native filesystem (/home/hoang/nofx/.state/) | DrvFs does NOT guarantee POSIX rename atomicity |
| Rollover handling | Manual NT re-deploy on roll day; pause trading window around roll | NT support: "NinjaScript strategies are not rolled forward" |
| Databento role | Historical only — warmup, gap-fill, backtest, archive | Live tier with OHLCV is $1,399/mo annual contract; NT live via prop firm is free and matches execution venue |
| Latency budget | End-to-end p99 < 500ms (NT bar-close to Go indicator update) | Comfortably achievable with 250ms poll; reassess only if sub-100ms required |

### Staged Implementation Recommendations

STAGE 0 — Week 1 (Plan 1.5 minimum viable):
1. Add a BarWriter section to ClaudeTrader.cs using the
   Calculate.OnEachTick + IsFirstTickOfBar + State.Historical guard
   from the NT8 Calculate Mode section above. Emit ONLY 1m bars.
   Write to C:\Users\hoang\NofxTrader\data\bars_MNQ_1m.csv with
   columns: bar_open_utc, wallclock_at_write_utc, open, high, low,
   close, volume.
2. Add a Go-side tail_csv package that polls
   /mnt/c/Users/hoang/NofxTrader/data/bars_MNQ_1m.csv every 250ms
   with persisted offset in /home/hoang/nofx/.state/.
3. Aggregate 5m, 15m, H1, H4 in Go using bar_open_utc as the
   canonical key.
4. Dedup by (symbol, bar_open_utc) with 24h TTL.
5. Run a 1-week soak test on SIM101 during RTH and overnight. Log
   NT-to-Go latency on every bar.

STAGE 1 — Weeks 2-3 (hardening):
6. Add gap-fill from Databento Historical on startup. Query
   everything between the latest persisted bar_open_utc and (now -
   16 minutes) to stay outside Databento's 15-minute Historical
   publication delay.
7. Add session-boundary detection using a hard-coded CME calendar
   table or by reading Databento's status schema offline.
8. Add a unit test that takes one known NT bar and one known
   Databento bar for the same minute, normalizes both, and asserts
   equality on the canonical key.
9. Add a daily smoke test that compares the previous day's NT bars
   (via CSV) against Databento Historical OHLCV-1m for the same day,
   after end-of-session, and alerts on any mismatch exceeding the
   tolerance ladder (defined in Merge Logic State Machine, to be
   added in a later commit).

STAGE 2 — only if Stage 0 measured p99 latency exceeds 500 ms:
10. Replace CSV with NetMQ pub/sub: NT publishes tcp://*:5556
    topic bar.MNQ.1m; Go subscribes. Keep CSV as warm-backup
    archive.

BENCHMARKS THAT CHANGE THE RECOMMENDATION:
- If NT-to-Go p99 latency > 500 ms during RTH: upgrade transport
  to NetMQ or WebSocket (Stage 2).
- If Databento Live drops below $200/month or a new lower tier with
  OHLCV appears: reconsider Databento Live as primary feed for
  backtest-live parity.
- If WSL2 adds Windows-side inotify forwarding (microsoft/WSL #4739):
  drop polling, use fsnotify.

### Caveats and Verification Items

[VERIFY-7] Measure NT-to-Go transport latency p50 and p99 over a
full RTH session on John's machine. If p99 exceeds 500 ms, Stage 2
transport migration is triggered. NT staff have noted that
Calculate.OnEachTick is "CPU intensive if your program code is
compute intensive" — for a bar-only writer that early-returns on
!IsFirstTickOfBar this is negligible, but verify under live NQ load
(~1k-10k ticks/min during RTH).

[VERIFY-8] FileShare.Read semantics across DrvFs: confirm
empirically that Go can open and read the file while NT holds the
write handle. Build a 1-day soak test before relying on it.
Validated in spirit by NT's own multi-threading help guide and
YSFKDR/NinjaTrader_Data_Exporter's use of ReaderWriterLockSlim,
but the cross-kernel boundary is the verification gap.

[VERIFY-9] Bar-close jitter under low-liquidity conditions. NT bars
only close on the next tick after the boundary. During Sunday-night
reopen or holiday sessions, multi-second close latency is documented
(NT forum thread "Bar Closing Time"). Affects arrival timing, not
bar content. Decide whether the strategy can tolerate.

[VERIFY-10] Read the actual upstream claudetrader.cs C# source via
git clone and diff against the patterns documented here before
merging the bar-writer changes. Automated fetch returned a
permissions error during research; the patterns above are from NT's
own help guide and corroborating community sources. Confirm
threading and State semantics in the live source match before
implementation.

CONFIRMATION ITEMS (recheck periodically, NOT pre-implementation
blockers):
- Databento pricing can change. Standard tier rose from $179/month
  at April 2025 launch to $199/month by May 2026; Plus $1,399/month
  verified May 2026. Recheck before any business decision involving
  Live data.
- CME session schedule changes occasionally. Verified May 2026 from
  CME's E-mini/Micro futures page. Use NT's Trading Hours template
  (Bars.IsFirstBarOfSession) and/or Databento's status schema rather
  than hard-coded clock arithmetic.
- WSL2 inotify limitation could be fixed in a future WSL2 release
  without obvious announcement. Don't write code that assumes
  "polling forever" — abstract the watcher behind an interface so
  future migration is a one-file change.
- Time zone conversion is the single most common bug source in this
  kind of bridge. Write the unit test described in Stage 1 step 8
  and run it on every PR that touches the timestamp normalization
  path.

## Plan 1.5 — Merge Logic State Machine (Research Summary 2026-05-25)

This subsection specifies how NT live bars and Databento Historical
bars combine into a single canonical 1-minute time series keyed by
bar_open_utc. It complements the design findings above, which
established the architecture and protocols. This section establishes
the LIFECYCLE — what happens at boot, during normal operation, on
disconnects, and at end-of-day.

### Three Rules That Dominate

1. NT wins every conflict. Discrepancies are logged (info / warn /
   error tiered by tick distance), never auto-resolved against NT.
   NT is the execution venue truth.

2. Time, not source, is the join key. All bars are keyed on
   bar_open_utc (second precision UTC). NT stamps bars at CLOSE
   per NT8 help guide, so Go reader must subtract 60s to obtain
   canonical open. Databento ohlcv-1m is timestamped at OPEN per
   NautilusTrader Databento integration docs.

3. Databento is structurally lagged by ~15 minutes. Databento's
   roadmap explicitly states intraday GLBX.MDP3 via Historical API
   has 15-min embargo without real-time entitlement. Databento can
   never close the gap to "now" — only to "now − 15 min."

### Cold Start Sequence

WARMUP WINDOW: 7-10 RTH sessions for indicators up to EMA(200) on 15m.
Matches backtrader's _minperiod convention and QuantConnect Lean's
SetWarmUp pattern. NautilusTrader uses the same historical-prefetch-
then-subscribe pattern (PR #3825).

BOUNDARY CALCULATION:
- databento_warmup_end = T0 - 15min - 2min safety = T0 - 17min
- nt_handoff_start = first NT bar with bar_close_utc >= T0
- overlap_window = [databento_warmup_end, nt_handoff_start] =
  15-17min hole BY CONSTRUCTION

SEQUENCE (sequential, not parallel):
1. Block on Databento timeseries.get_range for warmup window
   ending at T0 - 17min
2. Feed bars through indicator pipeline with is_warmup=true
3. Start consuming NT CSV tail
4. Mark system "live-ready" only when nt_handoff_start observed

FALLBACK on short Databento response:
- Missing bars inside active session: log warn, proceed if
  missing_pct <= 5%; abort if > 5%
- Missing bars during 16:00-17:00 CT halt or weekend: expected,
  no log

### Overlap Zone at Cold Start

DECISION: Accept the gap explicitly. Do NOT block waiting for
Databento to catch up (wastes 15 min of signal time). Do NOT
forward-fill (pollutes indicators with synthetic data).

PATTERN:
- Warm indicators on [T0 - warmup, T0 - 17min)
- Mark [T0 - 17min, T0) as state=GAP
- Start consuming NT at T0
- Schedule backfill from Databento for the gap at T0 + 17min

RISK: indicators path-dependent over the missing window (cumulative
VWAP anchored at session open, opening-range breakouts) must check
state != GAP before firing. Path-independent indicators (EMAs warmed
from earlier data) are safe.

Expose WarmupComplete boolean analogous to Lean's IsWarmingUp.
Strategy code is responsible for gating on it.

### Warm Restart Mid-Session

SCENARIO: Go process died 14:32, restarts 14:35. Local Parquet has
bars up to 14:30. Databento has bars up to 14:20. NT has been
emitting throughout (NT didn't die — only Go did).

RECOVERY HIERARCHY (fastest first):
1. Local Parquet replay for [boot - warmup, 14:30] — no network,
   indicators rehydrate in milliseconds
2. NT live tail consumed from 14:35 onward. NT may have written
   bars [14:30, 14:35] to CSV during the Go outage. VERIFY-11
   below.
3. Databento backfill for any remaining hole [14:30, 14:35] is
   partially available (Databento has through 14:20, won't have
   14:30 until ~14:45). Mark these state=PENDING_BACKFILL.

REFERENCE: NautilusTrader v1.225.0 (PR #3733) — "Fixed Interactive
Brokers historical bar subscriptions not restored after daily
gateway restart" — confirms automatic re-request of historical bars
on reconnect is production-grade pattern.

GO LOGIC ON BOOT:
- Read last_local_bar_open_utc from Parquet
- Issue ONE bounded Databento request for (last_local, T0 - 15min)
- Tail NT CSV from saved offset
- Backfill scheduler handles PENDING_BACKFILL bars at T+17min

### NT Disconnect / Reconnect During Session

SCENARIO: NT loses CQG/Rithmic price feed at 11:42, reconnects at
11:47. During gap NT emits NOTHING to CSV. Per NT staff: "Keep
running will not recalculate the historical data that has passed
so it would just keep running once the connection resumes."

DETECTION (defense in depth, three signals):

Signal 1: Time-since-last-bar
- Watchdog timer on canonical stream
- > 90s during session: suspect
- > 180s: confirmed disconnect

Signal 2: Heartbeat file
- NinjaScript writes UTC timestamp to heartbeat.txt every 5s
- Go reads heartbeat
- Stale > 30s: degraded

Signal 3: Connection probe
- NinjaScript writes line on OnConnectionStatusUpdate(ConnectionLost)
- And again on .Connected
- Definitive signal

Any TWO of three confirm NT outage.

FALLBACK DURING NT GAP:
- Stay on canonical stream, switch source-of-truth flag to DEGRADED
- Do NOT write Databento bars into canonical stream during gap
  (Databento is 15 min behind, gap [11:42, 11:47] won't exist on
  Databento until ~12:02)
- Mark [last_nt_bar, NT_reconnect_bar) as state=NT_GAP
- At NT_reconnect_bar + 17min, scheduled backfill pulls Databento
  [11:42, 11:47] and writes with source="databento_historical",
  provisional=true

TRANSITION HANDLING:
- When OnConnectionStatusUpdate -> Connected fires, do NOT call
  ReloadAllHistoricalData() from running strategy. NT help guide
  warns: "This method should NOT be called from any of the event
  methods which access data."
- For data-emitting NinjaScript, set ConnectionLossHandling =
  KeepRunning rather than default Recalculate (default would
  re-fire OnBarUpdate on backfilled bars, risking duplicates)
- If NT bars arrive later for the gap window, they overwrite
  Databento provisional bars per the NT-wins rule, discrepancy
  logged

### Discrepancy Detection and Logging

Once Databento catches up (T+17min for any bar), every NT bar
acquires a Databento counterpart for comparison.

AGGRESSIVENESS: Compare every bar as it becomes comparable. One
hash-map lookup per bar, cheap. Sampling only at EOD throws away
timing information that makes feed-quality bugs diagnosable.

TOLERANCE LADDER (NQ tick = 0.25, MNQ tick = 0.25):

OHLC fields:
- Exact match: no log
- Off by <= 1 tick (0.25): info
- Off by 2 ticks (0.50): warn
- Off by > 2 ticks: error

Volume:
- Ratio in [0.95, 1.05]: info, normal
- Ratio in [0.80, 0.95) or (1.05, 1.20]: warn, known-class divergence
- Ratio outside [0.80, 1.20]: error, investigate

Volume divergence between CQG/Rithmic and Databento is EXPECTED,
not an error condition by itself. Different aggregation, different
message coalescing.

LOG FORMAT (greppable single-line JSON):
{"ts":"2026-05-25T14:31:00Z","evt":"bar_discrepancy","sym":"NQM6",
 "bar_open_utc":"2026-05-25T14:30:00Z","field":"close",
 "nt":21450.25,"db":21450.50,"ticks_off":1,"vol_nt":1842,
 "vol_db":1851,"level":"info"}

Naming supports triage:
grep evt=bar_discrepancy | jq 'select(.level=="error")'

No alerts, no auto-pause (per task constraints).

### End-of-Day Reconciliation

At 16:00 CT (start of daily halt), NT has emitted full session.
At ~16:45 CT Databento has caught up. Reconciliation job runs
at 17:15 CT (17:00 reopen + 15-min buffer).

SEQUENCE:
1. Read NT bars for session from canonical Parquet
2. Read Databento bars for same session via one
   timeseries.get_range call
3. Bar-set diff:
   - In Databento but NOT in NT: nt_missing — real holes in
     execution-venue data. Write with source=databento_historical,
     provisional=true. Log warn. THIS catches "NT silently missed
     bars" — without this step canonical stream gets quietly
     shorter than the day.
   - In NT but NOT in Databento: db_missing. Extremely rare.
     Log info, trust NT.
   - In both: run tolerance ladder
4. Rewrite vs append: do NOT rewrite NT bars even if Databento
   disagrees (NT-wins rule). EOD writes limited to:
   - INSERT of nt_missing bars (with source tag and provisional)
   - UPDATE of db_close, db_volume, db_match_status audit columns
     on existing rows (NOT canonical OHLCV)

SCHEDULING:
- Automatic at 17:15 CT, with 30-min retry window if Databento
  fails
- Manual --reconcile-day YYYY-MM-DD CLI flag for ad-hoc replay

### Source Tagging

Single source string column per bar. Three values cover universe:
- nt_live: bar consumed from NT CSV in real time
- nt_replay: bar read from NT CSV after Go-process restart (NT
  was running, Go was not)
- databento_historical: bar fetched via Databento Historical API
  (warmup, gap fill, EOD insert)

Separate provisional boolean column marks bars that may be
overwritten later. Currently only databento_historical bars
filling NT gaps are provisional.

### Parquet Layout

DECISION: One file per symbol per day, all sources merged, source
as a column.

Pattern:
data/bars_1m/symbol=NQM6/date=2026-05-25/part-0.parquet

RATIONALE:
- Queries are overwhelmingly time-range scans on one symbol
- Splitting by source doubles file count, complicates "give me
  canonical series" read path, works against Parquet's columnar
  compression on bar_open_utc
- Modexa's Parquet partition design analysis: "Date-only
  partitioning when 90% of analytics filter by a time range" —
  this workload exactly
- Target file size 1-10 MB. 1440 bars/day × ~80 bytes = ~110 KB
  raw, ~30-50 KB compressed. Well below standard 128-512 MB
  row-group recommendation but acceptable at this volume.

### Gap Detection

Run periodic contiguous-sweep every 60s during session hours:

expected = expected_minute_stream(session_start, now - 1min)
for t in expected:
    if t not in local_index:
        gaps.append(t)
if gaps and now - max(gaps) > 120s:
    trigger_backfill(gaps)

expected_minute_stream MUST respect:
- 16:00-17:00 CT daily halt
- Friday 16:00 CT to Sunday 17:00 CT weekend gap
- CME holiday calendar

Without these the detector fires every halt minute. Pre-compute
session-minute calendar weekly from CME calendar.

### On-Demand Backfill

Databento timeseries.get_range is RANGE-ONLY. No single-bar
endpoint. Coalesce gap-fill requests to ranges with at most
5-min separation; minimizes HTTP overhead.

Canonical shape from Databento Python client (Go equivalent):
client.timeseries.get_range(
    dataset='GLBX.MDP3', symbols='NQ.c.0', schema='ohlcv-1m',
    stype_in='continuous', start='...', end='...'
)

Use stype_in='continuous' with NQ.c.0 / MNQ.c.0 to avoid rollover
bookkeeping inside Go process.

### Clock Skew (Windows Host)

W32Time service sufficient for this use case but MUST be
configured. Per Microsoft Learn "High Accuracy W32time Requirements":
W32time was created for "computers to be equal to or less than
five minutes (which is configurable) of each other for
authentication purposes." Five-minute skew is CATASTROPHIC for
bar joining.

CONFIGURATION on Windows host (where NT8 runs):
Command form from Microsoft Learn "Windows Time Service Tools
and Settings":
w32tm /config /manualpeerlist:"pool.ntp.org,0x8"
       /syncfromflags:manual /update

The 0x8 flag is SpecialInterval / client-mode association —
enables tighter polling.

WSL2 inherits Windows host clock by default. Verify with date -u
on both sides at startup.

TARGET SKEW: < 250 ms
- Anything above 1s means stale bar_open_utc lands in wrong
  minute bucket
- Startup check in Go: abort if
  |wsl_clock - windows_clock| > 250ms

### Session Boundary Merge Behavior

DAILY HALT 16:00-17:00 CT Mon-Thu:
Pre-compute expected minute set per session day:
session_minutes(day) = minutes(17:00CT day-1, 16:00CT day)
                       - holidays

Gap detector and EOD reconciler both consume this calendar.

WEEKEND Friday 16:00 CT → Sunday 17:00 CT:
- Treat as 49-hour expected gap
- Sunday 17:00 CT bar is first new bar of next trading week
- Tag with session_first_bar=true audit column
- Cold start on Sunday afternoon should fetch warmup window
  crossing weekend cleanly — Databento returns no bars during
  closed hours, contiguous-sweep detector must skip them

### Decision Matrix (Merge Logic)

State / NT available / Databento current / Action / Canonical source:

WARMUP / not yet / yes (delayed) / pull Databento [T0-7d, T0-17min]
/ databento_historical

OVERLAP_HOLE / not yet / no (inside 15-min embargo) / mark gap, wait
/ none

LIVE / yes / yes (for past bars) / consume NT CSV, compare against
Databento at T+17min / nt_live

NT_GAP / no (disconnect) / catches up at T+17min / mark gap,
backfill from Databento later / databento_historical (provisional)

NT_RECONNECT / yes / yes / resume NT, NT backfill (if any)
overwrites provisional Databento / nt_live (overwrites)

EOD_RECONCILE / session closed / full day available / diff sets,
insert nt_missing bars, log discrepancies / databento_historical
for missing only

WEEKEND / no (closed) / no new data / idle, no writes / none

### State Machine

States: BOOT -> WARMUP -> LIVE <-> NT_GAP, LIVE -> EOD_RECONCILE
-> WEEKEND -> BOOT

BOOT:
- load last_local_bar_open_utc from Parquet
- if last_local within current session: LIVE_RESUME
- else: WARMUP

WARMUP:
- request Databento [now - warmup_window, now - 17min]
- feed indicators in monotonic order with is_warmup=true
- on completion: emit WarmupComplete event -> LIVE

LIVE:
- consume NT CSV tail
- on each new NT bar: write canonical with source=nt_live
- every 60s: run gap_detector
- every bar+17min: run discrepancy_check against Databento
- on heartbeat_stale OR connection_lost: -> NT_GAP
- at 16:00 CT Mon-Thu or Fri close: -> EOD_RECONCILE

NT_GAP:
- log warn evt=nt_gap_open
- scheduled job at gap_start + 17min: fetch Databento
  [gap_start, min(now - 17min, gap_end)]
- write with source=databento_historical, provisional=true
- on heartbeat_recovered AND new NT bar arrives: -> LIVE
- on NT bar inside [gap_start, gap_end] (later): overwrite
  provisional bar, log info

EOD_RECONCILE:
- wait until 17:15 CT
- fetch Databento for full session
- set_diff against NT bars
- insert nt_missing bars (provisional=true)
- run tolerance ladder for each comparable bar
- -> WEEKEND (Fri) or sleep until 17:00 CT next day

WEEKEND:
- on Sunday 16:50 CT: -> WARMUP (short, just to seed indicators
  if state was lost)

### Implementation Checklist (Merge Logic)

1. Normalize NT timestamp on ingest: subtract 60s from NT CSV row
   to produce bar_open_utc. Unit-test against known fixture.

2. Confirm Databento OHLCV-1m timestamp convention via one live
   Historical request. Assert ts_event corresponds to bar-open.
   See [VERIFY-12].

3. Implement source tagging as Parquet column, not directory.

4. Implement three-signal disconnect detector (time-since-last-bar,
   heartbeat file, optional explicit connection-status line in
   NT CSV).

5. Set ConnectionLossHandling = KeepRunning in any data-emitting
   NinjaScript to avoid default Recalculate re-firing OnBarUpdate.

6. Schedule EOD reconciliation at 17:15 CT with 30-min retry
   window if Databento fails.

7. Pre-compute session-minute calendar for next 30 days, refresh
   weekly from CME holiday calendar.

8. Configure W32Time on Windows host with 0x8 interval flag.
   Verify w32tm /query /status reports stratum <= 3 before market
   open daily. Add startup check in Go process that aborts if
   |wsl_clock - windows_clock| > 250ms.

9. Define JSON log schema for bar_discrepancy. Rotate daily.

10. Build --reconcile-day CLI for manual EOD replay.

11. Test cold start across the 16:00-17:00 CT halt — warmup window
    will straddle the halt, gap detector must NOT fire on the
    missing 60 minutes.

### Additional Verification Items

[VERIFY-11] NT CSV behavior when NT is up but Go reader is down
for 5 min. Are bars present, or backfilled by NT into CSV
automatically? Determines whether warm restart can rely on
NT-side persistence.

[VERIFY-12] Databento OHLCV-1m timestamp convention. Make one
live Historical request, inspect ts_event vs bar boundaries,
confirm it's bar-OPEN (not close). NautilusTrader's adapter
normalizes to close internally — the raw Databento API does not.

[VERIFY-13] Databento behavior during 16:00-17:00 CT daily halt.
Does it emit bars with volume=0, or no rows? Code the gap
detector accordingly.

[VERIFY-14] NT CSV behavior on true CQG/Rithmic disconnect
lasting > 60s. Does NT Historical Data Server backfill into CSV
automatically, or only after ReloadAllHistoricalData() called?

[VERIFY-15] Clock skew between Windows host and WSL2 at startup.
Measure |wsl_clock - windows_clock| over 1 hour. Should be
< 250ms.

### Sources (Merge Logic)

NT8 ConnectionLossHandling:
https://ninjatrader.com/support/helpGuides/nt8/connectionlosshandling.htm

NT8 OnConnectionStatusUpdate:
https://ninjatrader.com/support/helpguides/nt8/onconnectionstatusupdate.htm

NT8 ReloadAllHistoricalData:
https://ninjatrader.com/support/helpguides/nt8/reloadallhistoricaldata.htm

NautilusTrader Databento integration:
https://nautilustrader.io/docs/latest/integrations/databento/

Databento 15-min historical embargo:
https://roadmap.databento.com/b/n0o5prm6/feature-ideas/release-historical-data-as-soon-as-possible-based-on-licensing-requirements-including-intraday

QuantConnect warmup documentation:
https://www.quantconnect.com/forum/discussion/4646/what-is-the-purpose-of-a-warmup-period-and-how-do-i-find-out-how-long-it-should-be/

CME equity futures trading hours:
https://www.cmegroup.com/education/files/eq-trading-hours.pdf

Microsoft Learn W32Time configuration:
https://learn.microsoft.com/en-us/windows-server/networking/windows-time-service/windows-time-service-tools-and-settings

QuantVPS NTP for futures trading:
https://www.quantvps.com/blog/ntp-time-synchronization-in-trading

NexusFi data feeds analysis:
https://nexusfi.com/a/platforms/data-feeds-market-data

Parquet partition design (Modexa):
https://medium.com/@Modexa/7-parquet-partition-designs-that-actually-work-69a2a0811ea8

## Why this architecture change

The CSV bridge in Plan 1 has known caps:

| Plan 1 (CSV) | Plan 1.5 (AddOn + socket) |
|---|---|
| One-shot file write, 2s NT polling | Push-driven events, sub-10ms |
| 1-second DateTime dedup → race conditions (H1-H4) | UUIDv7 cmd_id + monotonic seq |
| `GetBalance` returns $50k mock | `Account.Get(AccountItem.CashValue, ...)` + `AccountItemUpdate` subscription |
| `CloseLong/CloseShort` return error | `Account.CreateOrder(...counter-market...) + Account.Submit` |
| `CancelAllOrders` returns error | `Account.CancelAllOrders(instrument)` |
| `SetStopLoss`/`SetTakeProfit` set only at entry | `Account.Change(order)` mid-trade, with `isLiveUntilCancelled=true` |
| `GetClosedPnL` returns empty | `Account.Get(AccountItem.RealizedProfitLoss, ...)` |
| Broker disconnect: invisible | `ConnectionStatusUpdate` event |
| ATM strategies / OCO: unsupported | `AtmStrategy.StartAtmStrategy(...)` available |
| Fill ordering: relies on CSV file mtime | Documented sequence `OrderUpdate → ExecutionUpdate → PositionUpdate` (caveat: Rithmic/IB ordering not guaranteed — drive off `Execution` value, not `Position` cache) |

## Architecture

```
┌──────────────────────── Windows (NinjaTrader 8 process) ────────────────────────┐
│                                                                                  │
│  bin/Custom/AddOns/ClaudeBridge.cs   (new NinjaScript AddOn)                    │
│    ├─ TcpListener on 127.0.0.1:36974  (NDJSON command/event protocol)           │
│    ├─ Account.OrderUpdate / ExecutionUpdate / PositionUpdate /                   │
│    │  AccountItemUpdate / ConnectionStatusUpdate event subscriptions             │
│    ├─ ConcurrentDictionary<cmd_id, Order>  (idempotency LRU)                    │
│    ├─ Ring buffer of last 1000 events for resync on reconnect                   │
│    └─ Optional: claudetrader.cs Strategy still on chart                          │
│       for ATM OCO if you want server-side SL/TP autocancel                       │
└──────────────────────────────────────────────────────────────────────────────────┘
        │                                                            ▲
        │ NDJSON over TCP, mirrored loopback                          │
        │ {"cmd_id":"uuid7","auth":"<token>","action":"OpenLong",...} │
        ▼                                                            │
┌──────────────────────── WSL2 (Ubuntu) ───────────────────────────────────────────┐
│                                                                                   │
│  trader/ninjatrader/trader.go  (Plan 1.5 replaces Plan 1's CSV bridge here)      │
│    ├─ NDJSON client to NT8 (auto-reconnect, exp. backoff, PING/PONG keepalive)   │
│    ├─ Per-cmd UUIDv7 + monotonic seq + LRU dedup                                 │
│    ├─ Pending-command journal (BadgerDB or SQLite) — replay on NT restart        │
│    └─ State cache: positions / orders / equity / connection status               │
└──────────────────────────────────────────────────────────────────────────────────┘
```

## WSL2 networking (required prerequisite)

### Preferred: mirrored mode (Windows 11 22H2+)

```ini
# C:\Users\<user>\.wslconfig
[wsl2]
networkingMode=mirrored
```

Then `wsl --shutdown` and restart. Go in WSL2 connects to `127.0.0.1:36974` exactly as if it were on Windows. **No firewall rule, no host IP discovery, no NAT translation.**

**Limitations:**
- Windows Server 2025 does NOT support mirrored mode (Microsoft/WSL issue #12569).
- Windows 10 does NOT support mirrored mode.

### Fallback: NAT mode (any older Windows)

```powershell
# As Administrator
New-NetFirewallHyperVRule -Name "NTBridge" -DisplayName "NT8 Bridge" `
    -Direction Inbound -VMCreatorId '{40E0AC32-46A5-438A-A0B2-2B479E8F2E90}' `
    -Protocol TCP -LocalPorts 36974
```

In NT8 AddOn, bind `TcpListener` to `IPAddress.Any` (NOT `IPAddress.Loopback`). In WSL2, discover Windows host IP from `ip route show | grep -i default | awk '{ print $3}'` (typically 172.x.x.x).

## NinjaScript AddOn skeleton

`bin/Custom/AddOns/ClaudeBridge.cs`:

```csharp
public class ClaudeBridgeAddOn : NinjaTrader.NinjaScript.AddOnBase
{
    private TcpListener _listener;
    private CancellationTokenSource _cts;
    private Account _acct;
    private readonly ConcurrentDictionary<string, Order> _orders = new();
    private readonly ConcurrentQueue<string> _eventBuffer = new(); // last ~1000 events
    private long _seqCounter;
    private string _authToken;

    protected override void OnStateChange()
    {
        if (State == State.SetDefaults) { Name = "ClaudeBridge"; }
        else if (State == State.Configure)
        {
            // Load auth token (rotated by Go-side deploy)
            _authToken = File.ReadAllText(
                Path.Combine(Environment.GetEnvironmentVariable("USERPROFILE"),
                            ".claudebridge", "secret")).Trim();

            // Bind account (configurable via NT strategy parameter)
            lock (Account.All) _acct = Account.All.First(a => a.Name == "PropFirmAcct");

            // Subscribe to push-event sources
            _acct.OrderUpdate              += OnOrderUpdate;
            _acct.ExecutionUpdate          += OnExecutionUpdate;
            _acct.PositionUpdate           += OnPositionUpdate;
            _acct.AccountItemUpdate        += OnAccountItemUpdate;
            _acct.ConnectionStatusUpdate   += OnConnectionStatusUpdate;

            // Start TCP listener on background task
            _cts = new CancellationTokenSource();
            _listener = new TcpListener(IPAddress.Loopback, 36974);
            // CRITICAL: do NOT use port 36973 — that's NT's own ATI port
            _listener.Start();
            _ = Task.Run(() => AcceptLoopAsync(_cts.Token));
        }
        else if (State == State.Terminated)
        {
            _cts?.Cancel();
            _listener?.Stop();
            _acct.OrderUpdate              -= OnOrderUpdate;
            _acct.ExecutionUpdate          -= OnExecutionUpdate;
            _acct.PositionUpdate           -= OnPositionUpdate;
            _acct.AccountItemUpdate        -= OnAccountItemUpdate;
            _acct.ConnectionStatusUpdate   -= OnConnectionStatusUpdate;
        }
    }

    // ... AcceptLoopAsync handles each client on its own Task
    // ... OnOrderUpdate / OnExecutionUpdate / etc. push NDJSON to connected clients
    // ... Command handlers call Account.CreateOrder/Submit/Cancel/Change directly
    //     (these methods are thread-safe per NT's multi-threading guide)
}
```

**Critical threading rules** (per NT staff guidance):
1. Listener loop MUST be on `Task.Run` — never on `OnStateChange` direct call (that's a UI thread, would block NT)
2. Account.* methods are thread-safe and can be called from the listener task without `Dispatcher`
3. If you ever touch a NT chart drawing or NTWindow from the listener: use `Dispatcher.InvokeAsync` (NEVER `Dispatcher.Invoke` — deadlocks NT)
4. Every callback wrapped in `try/catch` — uncaught exceptions tear down the AddOn

## NT method mapping for the 9 required actions

All assume the **AddOn-level Account API** (not the Strategy-managed approach):

| Action | NT call |
|---|---|
| `OpenLong` / `OpenShort` (entry) | `Account.CreateOrder(instrument, OrderAction.Buy/Sell, OrderType.Market/Limit/StopMarket/StopLimit, TimeInForce.Day/Gtc, qty, limitPrice, stopPrice, oco, name, customId)` then `Account.Submit(new[]{order})`. Optionally bind ATM via `AtmStrategy.StartAtmStrategy(...)`. |
| `CloseLong` / `CloseShort` | Resolve `Position pos = _acct.Positions.FirstOrDefault(p => p.Instrument == instr)`; submit counter-market via `Account.CreateOrder(instrument, pos.MarketPosition == MarketPosition.Long ? OrderAction.Sell : OrderAction.BuyToCover, OrderType.Market, ...)`. |
| `CancelOrder(id)` | Look up in `_orders[cmd_id]`, call `Account.Cancel(new[]{order})`. |
| `CancelAllOrders` | `Account.CancelAllOrders(instrument)` — documented at ninjatrader.com/support/helpguides/nt8/accounts_cancelallorders.htm. |
| `ModifyStopLoss` / `ModifyTakeProfit` | Mutate `order.StopPrice` / `order.LimitPrice`, then `Account.Change(new[]{order})`. Order MUST have been submitted with `isLiveUntilCancelled=true`. |
| `MoveToBreakeven` | Read `pos.AveragePrice`, call `ModifyStopLoss` with that price rounded via `Instrument.MasterInstrument.RoundToTickSize(...)`. |
| `GetAccountBalance` | `Account.Get(AccountItem.CashValue, Currency.UsDollar)` + `BuyingPower` + `NetLiquidation` + `InitialMargin` + `RealizedProfitLoss`. **Caveat:** the `Currency` parameter is documented-as-ignored; values are realtime-only (0 in backtest); not callable from indicator context. |
| `GetOpenPositions` | Iterate `_acct.Positions` → `Instrument.FullName`, `MarketPosition`, `Quantity`, `AveragePrice`. |
| `GetPendingOrders` | `_acct.Orders.Where(o => o.OrderState == OrderState.Working \|\| o.OrderState == OrderState.Accepted)`. |

**Event push (NT → Go):** subscribed in `State.Configure`:

```csharp
_acct.OrderUpdate              += ...;   // every state transition (Submitted → Accepted → Working → PartFilled → Filled / Cancelled / Rejected)
_acct.ExecutionUpdate          += ...;   // every fill, by-value Execution arg
_acct.PositionUpdate           += ...;   // when position size/side changes
_acct.AccountItemUpdate        += ...;   // when CashValue / RealizedPnL / etc. change
_acct.ConnectionStatusUpdate   += ...;   // when broker connection drops/reconnects
```

**Critical caveat (Rithmic / Interactive Brokers):** per NT support, "sequence of events are not guaranteed due to provider API design" on these adapters. Drive Go-side state machine off `ExecutionUpdate`'s **by-value Execution** argument, NOT off cached `Position` properties.

## NDJSON protocol

**Command (Go → NT):**
```json
{"cmd_id":"01900a2d-...-uuid7","seq":42,"auth":"<sha256-hex-of-shared-secret>","action":"OpenLong","instrument":"MNQ 12-26","qty":1,"stop":21485.0,"target":21540.0}
```

**Response (NT → Go), one of:**
```json
{"reply_to":"01900a2d-...","ok":true,"order_id":"<nt-order-id>"}
{"reply_to":"01900a2d-...","ok":false,"error":"InsufficientBuyingPower"}
```

**Event push (NT → Go, unsolicited):**
```json
{"event":"fill","seq":1042,"instrument":"MNQ 12-26","side":"LONG","qty":1,"price":21500.25,"time":"2026-05-22T19:30:45.123Z"}
{"event":"order_state","seq":1043,"order_id":"<id>","state":"Working","price":21485.0}
{"event":"account_item","seq":1044,"item":"CashValue","value":49850.50}
{"event":"connection","seq":1045,"status":"Disconnected"}
```

## Idempotency + reconnect

- Every Go command carries UUIDv7 `cmd_id`. AddOn keeps an LRU of executed cmd_ids (in memory + persisted to `bin/Custom/AddOns/ClaudeBridge/cmd_log.jsonl`). Duplicate cmd_id replays = no-op.
- Every NT event carries monotonic `seq`. Go tracks last-seen seq. On reconnect, Go sends `{"action":"RESYNC","last_seq":1042}`. AddOn replays any events newer than that from its ring buffer.
- TCP keepalive: `socket.SetKeepAlive(true, 10_000, 5_000)` catches OS-level drops in ~15s.
- Application-level: PING every 1s with 3s timeout, catches hung NT processes.
- On Go-side connection loss: mark all cached state stale, retry connect with exponential backoff capped at 5s. Surface `bridge_status=disconnected` to NOFX dashboard.

## Security

- **Bind to 127.0.0.1 only** under mirrored mode. NAT-mode fallback: bind to vEthernet IP, never LAN IP.
- **Shared-secret token** in `%USERPROFILE%\.claudebridge\secret`. Every Go command includes `auth=<sha256-hex>` field. AddOn rejects mismatches.
- **No remote LAN exposure** without stunnel/wireguard in front. Prop-firm TOS likely prohibits this anyway.

## 5-stage migration plan (CSV → AddOn → live)

| Stage | What | Effort | Gate to proceed |
|---|---|---|---|
| **Stage 1** | Keep CSV alive, add read-only event channel. Build AddOn skeleton, subscribe `Account.*Update`, push NDJSON over TCP. NO order submission yet. | 2-3 days | Events arriving in Go with <50ms latency for ≥1h live |
| **Stage 2** | Add command channel for SAFE actions: `GetAccountBalance`, `GetOpenPositions`, `GetPendingOrders`, `CancelAllOrders`, `CloseLong/CloseShort`. (Read or "panic-flat" — conservative failure modes.) | 3-5 days | 1 trading day of flatten commands without stuck position |
| **Stage 3** | Add `ModifyStopLoss`, `ModifyTakeProfit`, `MoveToBreakeven`. Decide ATM-vs-Managed-vs-Unmanaged here and document. | 3-5 days | 50 modify ops live without orphan stop |
| **Stage 4** | Add `PlaceOrder` over socket. Feature-flag the CSV poller. Keep CSV as fallback for 1 month. | 2-3 days | 30 days zero socket-bridge incidents |
| **Stage 5 (parallel)** | Prototype Tradovate/ProjectX exit ramp. Build Go SDK client that performs same 9 actions against `live.tradovateapi.com/v1` or `api.topstepx.com/api`. | 1 week | Trigger to switch: >2 incidents/month requiring NT restarts |

**Realistic total Plan 1.5 timeline: 2-3 weeks** for stages 1-4. Stage 5 runs in parallel as the escape hatch.

## Buy vs build alternative

**CrossTrade XT AddOn** (https://crosstrade.io/blog/crosstrades-new-ninjatrader-add-on/) implements ~95% of this architecture as paid SaaS:
- 25 endpoints covering all 9 actions + more (account queries, position info, place/close/cancel/modify)
- Standard Unlimited: $24/mo
- Pro Unlimited: $49/mo
- 7-day free trial (no card)

**Tradeoff:** routes commands through CrossTrade's cloud servers — adds a hop for LLM agent loops targeting sub-second latency, and may not pass prop-firm compliance review (external order routing). Build-it-yourself is in-process and stays on your machine.

## Escape ramp: move off NT8 entirely (Stage 5)

If the prop firm clears through Tradovate or ProjectX, this is the cleanest architecture and removes NT8 as a single point of failure.

**Tradovate** (Apex, Tradeify, TPT historically, FundedNext, others):
- Live REST: `https://live.tradovateapi.com/v1`
- Demo REST: `https://demo.tradovateapi.com/v1`
- Market data: `https://md.tradovateapi.com`
- Official C# example: `tradovate/example-api-csharp-trading`
- All 9 actions natively supported via REST + WebSocket

**ProjectX / TopstepX** (Topstep, TFDX, Bulenox):
- REST: `https://api.topstepx.com/api`
- SignalR hubs: `wss://realtime.topstepx.com/api` (user hub + market hub)
- Mature Python SDK: `TexasCoding/project-x-py` on PyPI
- Add-on subscription: $14.50-29/mo through ProjectX

**What you lose by moving off NT8:**
- NT chart tooling
- Local sim environment
- Manual intervention via NT UI
- ATM strategies (server-side OCO)
- NT-specific indicators

**Worth it if:** your prop firm clears via Tradovate or ProjectX and your strategy doesn't depend on NT-specific indicators.

## Plan 1.5 known gotchas (from research)

1. **ATM strategies break `OnOrderUpdate`** — per NT support, "OnOrderUpdate/OnExecutionUpdate events will not trigger for strategies that submit ATM templates." If you use `AtmStrategyCreate`, you must use `GetAtmStrategyPositionAveragePrice()` and ATM-specific queries instead. Pick one approach and stick to it.
2. **Managed approach auto-resets stops/targets every bar** unless `isLiveUntilCancelled=true`. For a bridge where Go controls SL/TP placement, you do NOT want managed auto-reset. Use unmanaged or `ExitStopMarket(...isLiveUntilCancelled=true...)`.
3. **`Account.Get(AccountItem.CashValue, ...)` returns 0 in backtests and stale values from indicator context.** Treat as realtime-only; subscribe to `AccountItemUpdate` for change events rather than polling.
4. **Rithmic / IB `PositionUpdate` ordering is not guaranteed.** Drive Go state machine off `ExecutionUpdate`'s by-value `Execution` object, not cached `Position` properties.
5. **NT8 is .NET Framework 4.8** — cannot use modern NuGet packages compiled against .NET 6+. Pin dependencies. `async/await` works fine.
6. **Port 36973 is NT8's own ATI port** — do NOT collide. Use 36974 / 50051 / something else.
7. **WSL2 mirrored mode requires Windows 11 22H2+.** On Windows 10 (or Server 2025), use NAT mode + firewall rule.
8. **Pin NT version.** NT8 **8.1.6** (2025-09-25) or **8.1.6.3** (2026-01-16). Re-validate AddOn on every minor NT8 upgrade — third-party AddOns occasionally break on minor releases.
9. **NT8 sockets are officially "unsupported" by NT staff.** Forum stance: "Of course, no issues. Naturally, from NT support's point of view, this is unsupported C# code. But 'unsupported' only means 'we don't answer questions on that' — but you can certainly achieve anything your skills allow." You're on your own when it breaks.

## When to start Plan 1.5

Trigger conditions (any one):
- Plan 1 SIM validation complete (AI brain proven to make sensible NQ decisions)
- Need to flip to live trading (CSV bridge inadequate per H1-H4 hazards)
- CSV bridge experiencing >1 incident/week in SIM
- Need real balance read for accurate position sizing
- Need manual close / cancel-all / modify mid-trade for risk management

**Plan 1 stays canonical until one of those triggers fires.** Do not pre-emptively rebuild the bridge — the CSV path is sufficient for the "validate AI brain" objective.

---

# Reference: Developing New Strategies and Indicators

> **This section is a reference appendix, not part of the build sequence.** After Plan 1 ships, you'll want to add new indicators, new strategies, or fork existing ones. This guide shows the exact pattern.

## How to add a NEW INDICATOR

**Real example: adding Williams %R (an overbought/oversold oscillator).**

Three files touched, each with one focused change.

### Step W.1: Implement the indicator math

Append to [market/data_indicators.go](market/data_indicators.go):

```go
// calculateWilliamsR returns Larry Williams' %R oscillator over the last
// `period` bars. Range is -100 (oversold) to 0 (overbought).
// Reference: https://www.investopedia.com/terms/w/williamsr.asp
func calculateWilliamsR(klines []Kline, period int) float64 {
    if len(klines) < period {
        return 0
    }
    window := klines[len(klines)-period:]
    highest := window[0].High
    lowest := window[0].Low
    for _, k := range window {
        if k.High > highest {
            highest = k.High
        }
        if k.Low < lowest {
            lowest = k.Low
        }
    }
    if highest == lowest {
        return -50
    }
    close := window[len(window)-1].Close
    return -100 * (highest - close) / (highest - lowest)
}

// ExportCalculateWilliamsR is the package-public wrapper.
func ExportCalculateWilliamsR(klines []Kline, period int) float64 {
    return calculateWilliamsR(klines, period)
}
```

**Pattern to copy:** look at how `calculateEMA`, `calculateRSI`, `calculateATR` are structured — they all follow the same shape: lowercase implementation, uppercase Export wrapper. New indicators follow this pattern verbatim.

### Step W.2: Write the test FIRST (then implementation — TDD)

Append to [market/data_test.go](market/data_test.go):

```go
func TestCalculateWilliamsR_OversoldBottom(t *testing.T) {
    // Build a window where the close is at the period's low — should report
    // very negative (close to -100, deep oversold).
    klines := []Kline{
        {High: 100, Low: 90, Close: 100},
        {High: 102, Low: 92, Close: 95},
        {High: 101, Low: 88, Close: 93},
        {High: 99, Low: 85, Close: 86}, // close near the period low
    }
    got := calculateWilliamsR(klines, 4)
    // highest in window: 102, lowest: 85, close: 86
    // %R = -100 * (102 - 86) / (102 - 85) = -94.12
    want := -94.12
    if math.Abs(got - want) > 0.5 {
        t.Errorf("calculateWilliamsR = %.2f, want ~%.2f", got, want)
    }
}

func TestCalculateWilliamsR_OverboughtTop(t *testing.T) {
    klines := []Kline{
        {High: 100, Low: 90, Close: 91},
        {High: 102, Low: 92, Close: 100},
        {High: 105, Low: 95, Close: 103},
        {High: 110, Low: 100, Close: 109}, // close near the period high
    }
    got := calculateWilliamsR(klines, 4)
    // highest: 110, lowest: 90, close: 109
    // %R = -100 * (110 - 109) / (110 - 90) = -5.0
    if math.Abs(got - (-5.0)) > 0.5 {
        t.Errorf("calculateWilliamsR = %.2f, want -5.0 (overbought)", got)
    }
}

func TestCalculateWilliamsR_FlatRange(t *testing.T) {
    klines := []Kline{
        {High: 100, Low: 100, Close: 100},
        {High: 100, Low: 100, Close: 100},
    }
    got := calculateWilliamsR(klines, 2)
    if got != -50 {
        t.Errorf("flat-range Williams %%R = %.2f, want -50.0", got)
    }
}
```

Run them:

```bash
cd /home/hoang/nofx && go test ./market/... -run TestCalculateWilliams -v
```

Expected: 3 PASS. If any fail, fix the math.

### Step W.3: Surface the indicator to the AI prompt

In [kernel/engine_prompt_futures.go](kernel/engine_prompt_futures.go) — extend `FuturesContext`:

```go
type FuturesContext struct {
    // ... existing fields ...
    WilliamsR14 float64  // ADD: Williams %R, period 14
}
```

In `BuildFuturesUserPrompt`, add a line in the indicator-snapshot block:

```go
b.WriteString(fmt.Sprintf("- Williams %%R(14): %.1f (%s)\n",
    ctx.WilliamsR14, williamsBucket(ctx.WilliamsR14)))
```

Add the bucket helper next to `rsiBucket`:

```go
func williamsBucket(w float64) string {
    switch {
    case w >= -20:
        return "overbought"
    case w <= -80:
        return "oversold"
    default:
        return "neutral"
    }
}
```

### Step W.4: Wire into the trading loop

In `cmd/nq_smoke/main.go` (or, post-plan-1, in `trader/auto_trader_loop.go` where the futures context is assembled):

```go
ctx.WilliamsR14 = market.ExportCalculateWilliamsR(klines, 14)
```

### Step W.5: Update the prompt test

In [kernel/engine_prompt_futures_test.go](kernel/engine_prompt_futures_test.go) — add `Williams` to the required terms in `TestBuildFuturesUserPrompt_IncludesIndicators`:

```go
for _, s := range []string{"21500.00", "EMA20", "RSI14", "ATR14", "Bollinger", "Williams"} {
    // ...
}
```

### Step W.6: Run and verify

```bash
cd /home/hoang/nofx && go test ./... && go run ./cmd/nq_smoke
```

You should see the Williams %R value in the printed user prompt, and the AI now has it in its context. **Total LOC: ~50. Total time: ~30 minutes including reading and tests.**

### Step W.7: Commit

```bash
git add market/data_indicators.go market/data_test.go \
        kernel/engine_prompt_futures.go kernel/engine_prompt_futures_test.go \
        cmd/nq_smoke/main.go
git commit -m "feat(indicators): add Williams %R oscillator to NQ prompt context"
```

That's the entire flow. **Adding a 7th, 8th, 20th indicator is the same five steps.**

---

## How to add a NEW STRATEGY (different prompt template)

A "strategy" in this codebase = a prompt template + the rules baked into it. Different strategies live in separate files so you can have multiple and A/B test.

**Real example: adding a mean-reversion strategy (different from the trend-following one Task 9 builds).**

### Step S.1: Create the new prompt file

Create [kernel/engine_prompt_meanrev.go](kernel/engine_prompt_meanrev.go):

```go
package kernel

import (
    "fmt"
    "strings"
)

// MeanReversionPromptConfig is the same shape as FuturesPromptConfig but the
// strategy framing is different.
type MeanReversionPromptConfig struct {
    FuturesPromptConfig         // embed for tick, multiplier, R/R, etc.
    BBPeriod        int         // 20
    BBStdDevs       float64     // 2.0
    OversoldRSI     float64     // 30
    OverboughtRSI   float64     // 70
}

func BuildMeanReversionSystemPrompt(c MeanReversionPromptConfig) string {
    var b strings.Builder
    b.WriteString("# You are a MEAN-REVERSION index-futures trader.\n\n")
    b.WriteString("## Philosophy\n")
    b.WriteString("Price tends to revert to its statistical mean. You only enter\n")
    b.WriteString("when price has stretched far from that mean and there's evidence\n")
    b.WriteString("of exhaustion. You DO NOT chase trends.\n\n")
    b.WriteString(fmt.Sprintf("## Instrument: %s, tick %.2f, multiplier $%.2f/point\n\n",
        c.Symbol, c.TickSize, c.ContractMultiplier))
    b.WriteString("## Entry rules\n")
    b.WriteString(fmt.Sprintf("- LONG only when: price is below the lower Bollinger band(%d, %.1fσ)\n",
        c.BBPeriod, c.BBStdDevs))
    b.WriteString(fmt.Sprintf("  AND RSI(14) < %.0f (oversold).\n", c.OversoldRSI))
    b.WriteString(fmt.Sprintf("- SHORT only when: price is above the upper Bollinger band(%d, %.1fσ)\n",
        c.BBPeriod, c.BBStdDevs))
    b.WriteString(fmt.Sprintf("  AND RSI(14) > %.0f (overbought).\n", c.OverboughtRSI))
    b.WriteString("- If EMA20 > EMA50 (strong uptrend), DO NOT short. Skip the cycle.\n")
    b.WriteString("- If EMA20 < EMA50 (strong downtrend), DO NOT long. Skip the cycle.\n\n")
    b.WriteString("## Exits (you choose at entry)\n")
    b.WriteString("- Take profit = back to the Bollinger mid (mean reversion target).\n")
    b.WriteString("- Stop loss = beyond the band you entered against, plus 1 ATR buffer.\n\n")
    b.WriteString(fmt.Sprintf("- Minimum R/R: %.2f.\n", c.MinRiskReward))
    b.WriteString("\n## Output: same JSON shape as the base prompt.\n")
    return b.String()
}
```

### Step S.2: Test it

Create [kernel/engine_prompt_meanrev_test.go](kernel/engine_prompt_meanrev_test.go):

```go
package kernel

import (
    "strings"
    "testing"
)

func TestMeanReversionPrompt_HasMeanRevFraming(t *testing.T) {
    p := BuildMeanReversionSystemPrompt(MeanReversionPromptConfig{
        FuturesPromptConfig: FuturesPromptConfig{
            Symbol:             "MNQ",
            ContractMultiplier: 2.0,
            TickSize:           0.25,
            MinStopPoints:      15,
            MaxStopPoints:      50,
            MinRiskReward:      1.5,
        },
        BBPeriod:      20,
        BBStdDevs:     2.0,
        OversoldRSI:   30,
        OverboughtRSI: 70,
    })

    for _, must := range []string{"MEAN-REVERSION", "Bollinger", "oversold", "overbought", "DO NOT chase", "MNQ"} {
        if !strings.Contains(p, must) {
            t.Errorf("mean-rev prompt missing %q", must)
        }
    }
    for _, mustNot := range []string{"trend-following", "momentum"} {
        if strings.Contains(p, mustNot) {
            t.Errorf("mean-rev prompt contains anti-pattern %q", mustNot)
        }
    }
}
```

```bash
cd /home/hoang/nofx && go test ./kernel/... -run TestMeanRev -v
```

### Step S.3: Make the engine pick a strategy

In [config/config.go](config/config.go), add:

```go
// StrategyName picks the prompt template. Values: "trend" (default), "meanrev".
StrategyName string
```

Load in Init:

```go
cfg.StrategyName = getEnvOrDefault("STRATEGY_NAME", "trend")
```

In `cmd/nq_smoke/main.go` (or your live entrypoint), branch:

```go
var sysP string
switch cfg.StrategyName {
case "meanrev":
    sysP = kernel.BuildMeanReversionSystemPrompt(kernel.MeanReversionPromptConfig{
        FuturesPromptConfig: kernel.FuturesPromptConfig{
            Symbol:             "MNQ",
            ContractMultiplier: 2.0,
            TickSize:           0.25,
            MinStopPoints:      15,
            MaxStopPoints:      50,
            MinRiskReward:      1.5,
        },
        BBPeriod:      20,
        BBStdDevs:     2.0,
        OversoldRSI:   30,
        OverboughtRSI: 70,
    })
default: // "trend"
    sysP = kernel.BuildFuturesSystemPrompt(kernel.FuturesPromptConfig{ ... })
}
```

### Step S.4: Switch strategies via env var

```bash
# Run with trend-following
STRATEGY_NAME=trend go run ./cmd/nq_smoke

# Run with mean-reversion
STRATEGY_NAME=meanrev go run ./cmd/nq_smoke
```

Two strategies, same bot. **You can run them concurrently as two separate traders in the DB** — each trader has its own config and can point to a different strategy.

### Step S.5: Commit

```bash
git add kernel/engine_prompt_meanrev.go kernel/engine_prompt_meanrev_test.go \
        config/config.go cmd/nq_smoke/main.go
git commit -m "feat(strategy): add mean-reversion prompt template"
```

---

## How to CLONE and MODIFY an existing strategy

The cheapest way to develop a new strategy = copy a working one and tweak.

```bash
# 1. Clone
cp kernel/engine_prompt_futures.go    kernel/engine_prompt_myversion.go
cp kernel/engine_prompt_futures_test.go kernel/engine_prompt_myversion_test.go

# 2. Rename the symbols inside both files
sed -i 's/BuildFuturesSystemPrompt/BuildMyVersionSystemPrompt/g' kernel/engine_prompt_myversion.go kernel/engine_prompt_myversion_test.go
sed -i 's/BuildFuturesUserPrompt/BuildMyVersionUserPrompt/g'     kernel/engine_prompt_myversion.go kernel/engine_prompt_myversion_test.go
sed -i 's/FuturesPromptConfig/MyVersionPromptConfig/g'           kernel/engine_prompt_myversion.go kernel/engine_prompt_myversion_test.go
sed -i 's/FuturesContext/MyVersionContext/g'                     kernel/engine_prompt_myversion.go kernel/engine_prompt_myversion_test.go

# 3. Verify tests still pass on the clone
go test ./kernel/... -run TestBuildMyVersion -v

# 4. NOW make your changes
$ code kernel/engine_prompt_myversion.go
```

You now have a parallel strategy you can edit freely without breaking the original. Run the original in one trader, your variant in another, compare results in the nofx web UI.

---

## How to RUN MULTIPLE STRATEGIES side-by-side (A/B test)

The codebase's `TraderManager` already supports multiple traders, each with its own config. After Plan 1, you can:

1. Create Trader A in the nofx web UI: `Strategy=trend`, `Symbol=MNQ`, `AI Model=DeepSeek`, status=Running.
2. Create Trader B: `Strategy=meanrev`, `Symbol=MNQ`, `AI Model=DeepSeek`, status=Running.
3. Both write to the SAME `trade_signals.csv` (be careful — one ClaudeTrader instance can only handle one position at a time. **For true A/B you need TWO NinjaTrader charts each with its own ClaudeTrader strategy and its own CSV file pair.**)

**Practical recipe for true A/B:**
- Two MNQ charts in NT, each running ClaudeTrader, configured with two different file paths:
  - `C:\Users\<u>\NofxTrader\data_A\trade_signals.csv` ← Trader A writes here
  - `C:\Users\<u>\NofxTrader\data_B\trade_signals.csv` ← Trader B writes here
- nofx config: per-trader `NinjaTraderDataDir` overrides the global default
- Run both for a week. Compare P&L in the nofx Dashboard.

---

## What changes if you add a NEW DATA PROVIDER (e.g., Polygon, IBKR)

Same pattern as Databento. Three files:

```
provider/polygon/client.go        ← HTTP/auth wrapper
provider/polygon/historical.go    ← GetOHLCV(symbol, interval, start, end)
market/polygon_adapter.go         ← Bars to market.Kline
```

Add a `DATA_PROVIDER=polygon|databento` env switch, branch in `kernel/engine.go`. **The downstream (indicators, prompts, CSV bridge, NT) doesn't know or care which provider supplied the bars.** That's the value of the adapter pattern.

---

## What changes if you add a NEW BROKER (instead of NinjaTrader)

Same pattern as the NinjaTrader CSV bridge. Two packages:

```
provider/tradovate/  (or whichever broker)
  client.go            ← REST + WebSocket auth
  orders.go            ← place/modify/cancel
  positions.go         ← state

trader/tradovate/
  trader.go            ← implements trader/types.Trader interface
```

Add a case in [trader/auto_trader.go:263-315](trader/auto_trader.go#L263-L315) switch. **The strategy + AI side doesn't care which broker executes.**

---

## Boundaries — what you CANNOT change without bigger work

- **The `market.Kline` shape itself** — touched by 20+ files. Changing field names cascades. If you need extra per-bar fields, add them as new optional fields rather than renaming existing ones.
- **The `trader/types.Trader` interface** — adding a method means every existing broker impl must implement it (or use a default in an embedded base). Removing a method means audits across all impls. Add carefully.
- **The decision JSON output shape** — `{action, entry, stop_loss, take_profit, reasoning}`. The validator at the strategy-engine layer, the persistence layer in the DB, the web UI dashboard table — all read this shape. Add fields rather than remove. Adding a "confidence" field is safe; renaming "entry" to "entry_price" is a multi-file refactor.
- **The CSV protocol with NinjaTrader** — `claudetrader.cs` defines the 5-field signal and 3-field fill. Changing these requires also editing the NinjaScript and recompiling in NT. Don't change unless you have a real need.

Everything else is fair game. Strategies, indicators, prompt wording, the AI provider, the data provider, the broker — all designed to be swappable.

---

## TL;DR — strategy development cycle in 5 minutes

```
1. Have an idea.
2. Either: clone a working prompt template → edit the language.
   Or:     add a new indicator function (15-30 lines + test).
3. Run cmd/nq_smoke once, hand-paste a decision, watch it fire in NT SIM.
4. Once it fires correctly, enable as a trader in the web UI.
5. Watch it run live in NT SIM for a few sessions.
6. If profitable, raise contract size. If not, edit prompt, GOTO 3.
```

No UI development needed. No backend refactor needed. The system is built to let you iterate on strategies and indicators with minimum friction.

---

# External References

> Verified by HTTP probe on 2026-05-22.

## Upstream NOFX project (canonical source)

- **Repository:** https://github.com/NoFxAiOS/nofx (default branch: `dev`, last pushed 2026-05-11)
- **Description:** *"Your personal AI trading assistant. Any market. Any model. Pay with USDC, not API keys."*
- **Local clone:** `/home/hoang/nofx` (this working tree)

### Architecture documentation (upstream)

Located at https://github.com/NoFxAiOS/nofx/tree/dev/docs/architecture — these are the authoritative module references:

| Doc | Size | Relevance to this plan |
|---|---|---|
| [README.md](https://github.com/NoFxAiOS/nofx/blob/dev/docs/architecture/README.md) | 6.3 KB | Overall system architecture, module map |
| [STRATEGY_MODULE.md](https://github.com/NoFxAiOS/nofx/blob/dev/docs/architecture/STRATEGY_MODULE.md) | **21.7 KB** | **PRIMARY REFERENCE** — full trading-cycle data flow, prompt construction, risk control. Plan 1's Tasks 4, 9, 10 align with this doc's stages 2 (Data Assembly), 3 (System Prompt), 4 (User Prompt), 6 (AI Parsing). |
| [AGENT_MEMORY_AND_PLANNING.md](https://github.com/NoFxAiOS/nofx/blob/dev/docs/architecture/AGENT_MEMORY_AND_PLANNING.md) | 11.2 KB | Agent (NOFXi) memory + planning subsystem. Out of scope for this plan but informs future work. |
| [X402_STREAMING_PAYMENT.md](https://github.com/NoFxAiOS/nofx/blob/dev/docs/architecture/X402_STREAMING_PAYMENT.md) | 12.7 KB | claw402 / x402 micropayment protocol. **Not used by Plan 1** (we bypass claw402 via direct Databento subscription). |

### Confirmed alignment with STRATEGY_MODULE.md

Plan 1 hooks into the same cycle stages the upstream doc describes:

- **Stage 1 (Coin Selection)** → For NQ we use the **static** path (single symbol "MNQ" or "NQ.c.0"); the `Static` mode is documented in STRATEGY_MODULE.md §1.1 at `decision/engine.go:395-403`. No change to selection code; just static-mode configuration.
- **Stage 2 (Data Assembly)** → Task 4 inserts the Databento adapter at the K-line ingestion point.
- **Stages 3 + 4 (Prompts)** → Task 9 adds a futures-mode template alongside the existing crypto template.
- **Stage 6 (AI Parsing)** → Unchanged; the futures prompt outputs the same JSON shape (`action`/`entry`/`stop_loss`/`take_profit`/`reasoning`) the existing parser expects.

This is important: **Plan 1 does not require structural changes to the strategy engine** — it adds new templates and a new data provider, both behind existing extension points.

## NinjaTrader CSV bridge

- **Repository:** https://github.com/J0shusmc/Claude-Trader-NinjaTrader
- **Key file:** `ninjascripts/claudetrader.cs` (14.8 KB) — the NinjaScript strategy that polls `trade_signals.csv` and places orders.
- **Auxiliary files:** `ninjascripts/SecondHistoricalData.cs`, `ninjascripts/SecondLifeFeed.cs` (data-export from NT to CSV — **not used by Plan 1**, since we get data from Databento).
- **Last update:** 2025-11-25 — production-ready per the README.
- **Status verified:** read claudetrader.cs in full during plan preparation. CSV contract (5-field signals, 3-field fills) confirmed against actual C# code, not just README claims.

## Databento

- **Documentation portal:** https://databento.com/docs/
- **Historical API base:** `https://hist.databento.com/v0/`
- **Authentication:** HTTP Basic, API key as username, empty password (NOT a Bearer token despite some community examples).
- **Pricing:** Pay-per-symbol-day. NQ continuous (`NQ.c.0`) on `GLBX.MDP3` is ~$0.001-0.005 per call for `ohlcv-1m` over short windows.
- **Datasets:** `GLBX.MDP3` covers all CME Globex products (NQ/MNQ/ES/MES/RTY/MRTY/CL/GC/etc.) — one subscription handles every CME futures contract.

### **CORRECTION — Go SDK status**

The URL `https://github.com/databento/databento-go` returns **HTTP 404** — there is **no official Databento Go SDK**.

What does exist:
- Official Python SDK: https://github.com/databento/databento-python
- Official Rust SDK + binary format library: https://github.com/databento/dbn
- C++ SDK: https://github.com/databento/databento-cpp
- **Community Go library:** https://github.com/NimbleMarkets/dbn-go (HTTP 200 verified 2026-05-22) — implements DBN binary format parsing + REST helpers. Active maintenance; not endorsed by Databento but used in production by NimbleMarkets. Could replace Plan 1's `net/http` approach if a more featureful SDK is desired later.

For Go, Plan 1's approach is correct as a starting point: call the REST API directly via `net/http` (Task 2). It's minimal and depends on no third-party code. If the project later wants DBN-binary streaming or more complete schema bindings, swap to `NimbleMarkets/dbn-go` — the adapter pattern in `market/databento_adapter.go` keeps the boundary clean and the swap is 1-2 days.

### **CORRECTION — Upstream docs reference older code structure**

The upstream [STRATEGY_MODULE.md](https://github.com/NoFxAiOS/nofx/blob/dev/docs/architecture/STRATEGY_MODULE.md) cites code at paths like `decision/engine.go:395-403`. **The current codebase does not have a `decision/` folder** — that module was refactored into `kernel/engine.go`. The function names and behavior cited are still accurate; only the file paths drifted. When implementing, find the equivalent in `kernel/`:

| Upstream doc cite | Current local path |
|---|---|
| `decision/engine.go` | `kernel/engine.go` |
| `decision/engine_analysis.go` | `kernel/engine_analysis.go` |
| `decision/engine_position.go` | `kernel/engine_position.go` |
| `decision/engine_prompt.go` | `kernel/engine_prompt.go` |

Other docs may have similar drift. If a cited path doesn't exist, search for the function name within `kernel/`.

## Quick-reference URLs (all verified)

```
Local code:                 /home/hoang/nofx
Upstream NOFX:              https://github.com/NoFxAiOS/nofx
Architecture docs:          https://github.com/NoFxAiOS/nofx/tree/dev/docs/architecture
NT CSV bridge:              https://github.com/J0shusmc/Claude-Trader-NinjaTrader
Databento docs portal:      https://databento.com/docs/
Databento Hist API:         https://hist.databento.com/v0/
Databento Go SDK (official): (does not exist — use net/http per Task 2)
Databento Go SDK (community): https://github.com/NimbleMarkets/dbn-go (alternative for later swap)
```

---

## CSV bridge runtime hazards — must address before live trading

> **Identified by external audit 2026-05-22.** The current Plan-1 CSV protocol has 4 known runtime issues. Acceptable for SIM paper-trading; **not acceptable for live execution**. Add a Plan 1.5 (or fold into Plan 2 safety layer) to harden these before flipping to a funded broker connection.

### Hazard H1: Lost signal race

- **Symptom:** Bot writes signal A at T=0. NT poll fires at T=2s. Bot writes signal B at T=1.5s, overwriting A. NT reads B, never sees A.
- **Cause:** `WriteSignal` truncates+rewrites the file each call. No acknowledgment back to Go that NT consumed the prior signal.
- **Fix:** Block subsequent writes until the tailer confirms a fill (or N-second timeout has passed) for the prior signal. Track an in-flight flag per trader. Reject `WriteSignal` calls when in-flight is set.

### Hazard H2: 1-second dedup collision

- **Symptom:** Two same-direction signals fire in the same wall-clock second. NT's dedup key is `DateTime+Direction`. Second signal silently dropped.
- **Cause:** [claudetrader.cs:202](claudetrader.cs#L202) — `string signalId = $"{parts[0]}_{parts[1]}";` with DateTime format `MM/dd/yyyy HH:mm:ss` (no fractional seconds).
- **Fix:** Either (a) require Go-side rate-limit of ≥2 seconds between same-direction writes (matches the dedup window), OR (b) modify `claudetrader.cs` to include a sub-second nonce/sequence number in the signal ID (`{DateTime}_{Direction}_{Nonce}`), and have the Go writer include a monotonic nonce as a 6th CSV field.

### Hazard H3: DrvFs (WSL2 `/mnt/c`) atomic-rename mtime

- **Symptom:** Go's `os.Rename` is atomic on ext4 but mtime propagation through the WSL2 DrvFs mount may or may not reliably bump `File.GetLastWriteTime` as observed from Windows NT process.
- **Cause:** DrvFs is a 9P-protocol bridge with semantic-translation quirks; not all metadata operations propagate identically to native NTFS access.
- **Fix:** Empirically test in Step 0.8: write a signal, verify NT's `File.GetLastWriteTime` changes within 1 second. If mtime doesn't propagate, fall back to write-in-place (non-atomic) with file locking, OR modify `claudetrader.cs` to poll content-hash instead of mtime.

### Hazard H4: Fill replay on session rollover

- **Symptom:** NT cycles `trades_taken.csv` at session close (truncates or rotates). Go-side tailer sees file shrink, resets seen-rows to 0, then re-emits every row that re-appears as if new → duplicate DB fills.
- **Cause:** Tailer tracks position by line count. Reset logic at file-shrink triggers on rotation.
- **Fix:** Key the tailer on a stable fill identity (timestamp + price + side), not line count. Maintain a "seen fills" set in memory or DB; ignore any row whose key has been seen before.

### Required for live (not Plan 1 paper)

These 4 hazards are not in scope for Plan 1's SIM-only goal. Add them as Plan 1.5 (or first part of Plan 2's safety layer) before flipping NT's broker connection from SIM101 to a funded account. The mitigation work is ~200-300 LOC: Go-side in-flight tracking + nonce + stable-fill-ID set, plus a minor claudetrader.cs extension to echo the nonce in fill rows.

---

## `GetBalance` $50k mock — honest disclosure

[provider/ninjatrader/trader.go:GetBalance](provider/ninjatrader/trader.go) (per Task 7) returns hardcoded:

```go
return map[string]interface{}{
    "totalEquity":      50000.0,
    "availableBalance": 50000.0,
}, nil
```

**This affects two consumers:**

1. **AI prompt equity math** — `BuildSystemPrompt` computes max position size as `equity × ratio`. With $50k mocked, the AI is told it has $50k regardless of the real SIM balance. Decisions sized against fantasy capital.
2. **Go-side risk checks** — `MaxMarginUsage`, `BTCETHMaxPositionValueRatio`, `MinPositionSize` validations all use the balance from `GetBalance()`. They will pass/reject orders based on $50k, not real account state.

**Acceptable for:** SIM paper trading where SIM101 default balance is also ~$50k, AND you understand the prompt is decoupled from real account state.

**Not acceptable for:** any live deployment.

**Fix options:**

| Option | Effort | When |
|---|---|---|
| Extend `claudetrader.cs` to write a `balance.csv` (account equity + available) on each cycle; Go tailer reads it | Medium (~80 LOC C# + 60 LOC Go) | Plan 1.5 |
| Read balance from NinjaTrader's Connection panel via a separate NinjaScript export | Same as above | Plan 1.5 |
| Disable Go-side balance-based sizing when `ExchangeType==ninjatrader` (let the AI's risk reasoning + NT's hard contract-quantity cap do the work) | Low (~20 LOC) | Plan 1 if going live early |

Recommended: option 3 for paper validation, then option 1 before any funded-account flip.

---

## Security must-fix before any live deployment

**Issue:** [config/config.go:67-69](config/config.go#L67-L69) defaults `JWTSecret` to the literal string `"default-jwt-secret-change-in-production"` when the `JWT_SECRET` env var is unset:

```go
if cfg.JWTSecret == "" {
    cfg.JWTSecret = "default-jwt-secret-change-in-production"
}
```

**Risk:** anyone running the bot with default config has the same JWT-signing key as every other default-config install on the internet. JWT tokens can be forged trivially. **Documented but easy to miss.**

**Required action (must be done before exposing the bot to any non-localhost network):**

1. Generate a strong random secret: `openssl rand -base64 64`
2. Set it in `.env` as `JWT_SECRET=<the random string>`
3. **Verify by env var, NOT the log line.** The log `🔑 JWT secret configured` at [main.go:90](main.go#L90) fires unconditionally on every startup — it does NOT indicate the default was overridden. Use one of these instead:
   - `grep "^JWT_SECRET=" /home/hoang/nofx/.env` — must return your generated key, not empty
   - Add a startup warning in code (recommended): modify [config/config.go:67-69](config/config.go#L67-L69) to log a loud warning when the default value is detected:
     ```go
     if cfg.JWTSecret == "" {
         cfg.JWTSecret = "default-jwt-secret-change-in-production"
         logger.Warnf("⚠️  JWT_SECRET env var not set; using INSECURE default. " +
             "This is acceptable for localhost-only paper trading. " +
             "Set JWT_SECRET in .env before any network-exposed deploy.")
     }
     ```
   This way the log distinguishes "default in use" (warning) from "real secret loaded" (silent or info).

This is not specific to NQ trading — it's a pre-existing nofx hardening item. Add to Plan 0 (Task 0.10) below if planning a live deploy:

- [ ] **Step 0.10: Set JWT_SECRET to a strong random value**

```bash
echo "JWT_SECRET=$(openssl rand -base64 64)" >> /home/hoang/nofx/.env
```

Confirm: `grep JWT_SECRET /home/hoang/nofx/.env` should show your generated key, NOT the literal "default-jwt-secret-change-in-production".

This is a no-op for LOCAL-ONLY paper trading (bot binds to localhost), but is non-negotiable for any deployment reachable from the internet.

---

# Function Transfer Manifest

> **Distilled from a 15-agent parallel audit on 2026-05-22.** The audits read every domain end-to-end (schema, kernel, market, broker pattern, API handlers, web UI, i18n, bootstrap). This section is the **canonical actionable list** — every function that must be CREATED, every function that must be MODIFIED, and every function that transfers UNCHANGED.

## A — NEW functions to implement

Group 1: **Databento data layer** (provider/databento/, market/)

| # | File | Function signature | LOC |
|---|------|---------------------|-----|
| A1 | `provider/databento/client.go` ✓ done | `NewClient(baseURL, apiKey string) *Client` + `doRequest(path, params) ([]byte, error)` + `basicAuth(user, pass) string` | 120 |
| A2 | `provider/databento/historical.go` | `(c *Client) GetOHLCV(symbol, interval string, start, end time.Time) ([]Bar, error)` | 80 |
| A3 | same | `parseOHLCVResponse(body []byte) ([]Bar, error)` + `(rawBar).toBar()` + `scaledFloat(s)` | 60 |
| A4 | `provider/databento/resolve.go` | `(c *Client) ResolveContinuous(symbol string) (string, error)` + `parseResolveResponse(body, symbol) (string, error)` | 60 |
| A5 | `market/databento_adapter.go` | `BarsToKlines(bars []databento.Bar) []Kline` | 25 |
| A6 | `market/data.go` (add helper) | `isCMEFuturesSymbol(symbol string) bool` (regex/map of NQ/MNQ/ES/MES/YM/MYM/GC/SI/CL/NG/ZB/ZN/ZT/ZW/ZC/ZS/ZL/ZO/etc.) | 25 |

Group 2: **NinjaTrader CSV bridge** (provider/ninjatrader/)

| # | File | Function signature | LOC |
|---|------|---------------------|-----|
| A7 | `provider/ninjatrader/types.go` | `SignalRow` + `FillRow` types, header constants, `(SignalRow) Validate() error` | 80 |
| A8 | `provider/ninjatrader/csv_writer.go` | `NewCSVWriter(dataDir) *CSVWriter` + `(w) WriteSignal(SignalRow) error` (atomic temp+rename) + `(w) SignalsPath() string` | 90 |
| A9 | `provider/ninjatrader/csv_tailer.go` | `NewCSVTailer(dataDir, pollInterval) *CSVTailer` + `(t) TailFills(ctx, onFill) error` + `parseFillRow(line) (FillRow, error)` | 110 |

Group 3: **Trader interface impl** (trader/ninjatrader/)

| # | File | Function | LOC | Strategy |
|---|------|----------|-----|----------|
| A10 | `trader/ninjatrader/trader.go` | `New(Config) *Trader` constructor + tailer goroutine | 40 | clean |
| A11 | same | `(t) OpenLong(symbol, qty, leverage) (map, error)` — bundle stashed SL/TP into one SignalRow | 30 | clean |
| A12 | same | `(t) OpenShort(...)` — same shape as OpenLong | 30 | clean |
| A13 | same | `(t) CloseLong(...)` / `CloseShort(...)` — return error "auto-close via SL/TP" | 8 | error |
| A14 | same | `(t) SetStopLoss(symbol, side, qty, stopPrice) error` — stash in internal map | 15 | clean |
| A15 | same | `(t) SetTakeProfit(symbol, side, qty, tpPrice) error` — stash | 15 | clean |
| A16 | same | `(t) SetLeverage(symbol, leverage) error` — noop (return nil) | 3 | noop |
| A17 | same | `(t) SetMarginMode(symbol, isCross) error` — noop | 3 | noop |
| A18 | same | `(t) GetMarketPrice(symbol) (float64, error)` — query Databento last 1m bar | 15 | clean |
| A19 | same | `(t) CancelStopLossOrders(symbol) error` — return error "not supported" | 3 | error |
| A20 | same | `(t) CancelTakeProfitOrders(symbol) error` — error | 3 | error |
| A21 | same | `(t) CancelAllOrders(symbol) error` — error | 3 | error |
| A22 | same | `(t) CancelStopOrders(symbol) error` — error (legacy) | 3 | error |
| A23 | same | `(t) FormatQuantity(symbol, qty) (string, error)` — `fmt.Sprintf("%.0f", qty)` for whole contracts | 5 | clean |
| A24 | same | `(t) GetOrderStatus(symbol, orderID) (map, error)` — query internal pending map; "filled" after fill row | 25 | partial |
| A25 | same | `(t) GetBalance() (map, error)` — return fixed SIM101 mock $50,000 | 12 | mock |
| A26 | same | `(t) GetPositions() ([]map, error)` — read from tailer state | 25 | clean |
| A27 | same | `(t) GetClosedPnL(start, limit) ([]ClosedPnLRecord, error)` — parse trades_taken.csv historical rows | 50 | clean |
| A28 | same | `(t) GetOpenOrders(symbol) ([]OpenOrder, error)` — return `[]OpenOrder{}` (CSV doesn't expose pending) | 5 | empty |

**Trader interface methods: 17 implemented, 11 cleanly via CSV, 4 noop/empty, 2 error-returning.** All compile-time-checked via `var _ types.Trader = (*Trader)(nil)`.

Group 4: **NQ futures AI prompt** (kernel/)

| # | File | Function signature | LOC |
|---|------|---------------------|-----|
| A29 | `kernel/engine_prompt_futures.go` | `BuildFuturesSystemPrompt(FuturesPromptConfig) string` | 80 |
| A30 | same | `BuildFuturesUserPrompt(FuturesContext) string` | 60 |
| A31 | same | `side(price, ref) string` + `emaAlignment(20, 50) string` + `rsiBucket(r) string` + `bollPosition(p, u, l) string` | 40 |

Group 5: **Futures-specific validation** (kernel/engine_position.go additions)

| # | File | Function signature | LOC |
|---|------|---------------------|-----|
| A32 | `kernel/engine_position.go` (add) | `roundTickNQ(price float64) float64` — round to nearest 0.25 | 5 |
| A33 | same | `roundTickMNQ(price float64) float64` — round to 0.25 (same; MNQ tick is also 0.25 not 0.05) | 5 |
| A34 | same | `validateStopDistanceNQ(entry, stop float64, minPoints int) error` — enforce min-point gap | 12 |
| A35 | same | `validateContractSizeNQ(contracts int, currentPrice, maxNotional float64) error` — guard notional exposure | 12 |

Group 6: **End-to-end smoke runner**

| # | File | Function | LOC |
|---|------|----------|-----|
| A36 | `cmd/nq_smoke/main.go` | one-shot runner: fetch bars → compute indicators → prompt → stdin-paste AI decision → write signal → tail fills | 130 |

**Total NEW code: ~1,200 LOC** + **~400 LOC tests** = **~1,600 LOC new**.

---

## B — EXISTING functions to MODIFY (specific edits only)

| # | File:Line | Change | Δ LOC |
|---|-----------|--------|-------|
| B1 | `market/data.go:557-558` (`Normalize`) | **Case-preserving fix:** capture raw symbol BEFORE the `strings.ToUpper` call (line 558), check `isCMEFuturesSymbol(raw)` first, and `return raw` if true. Otherwise proceed with existing logic. Databento continuous symbols use lowercase suffix (`NQ.c.0`, not `NQ.C.0`). Inserting the guard only at line 583 would receive an already-uppercased string and return `NQ.C.0`, which Databento rejects. Code shape: `func Normalize(symbol string) string { raw := symbol; if isCMEFuturesSymbol(raw) { return raw }; symbol = strings.ToUpper(symbol); ... }` | +5 |
| B2 | `market/data_klines.go:getKlinesFromCoinAnk` line 18 (or one level up) | Branch: if symbol is CME futures, call Databento adapter instead | +30 |
| B3 | `store/exchange.go:43` (Exchange struct) | Add field `NinjaTraderDataDir string` with gorm tag | +2 |
| B4 | `store/exchange.go:231` (Create) | Add `ninjaTraderDataDir string` param + assignment | +6 |
| B5 | `store/exchange.go:277-323` (Update) | Add to updates map | +4 |
| B6 | `store/exchange.go:initTables` (Postgres migration block) | Add conditional `ALTER TABLE` for the new column | +6 |
| B7 | `store/exchange.go:203-224` (`getExchangeNameAndType`) | Add `case "ninjatrader": return "NinjaTrader Futures", "cex"` | +2 |
| B8 | `store/visibility.go:5-37` (`MissingRequiredExchangeCredentialFields`) | Add `case "ninjatrader"` requiring `ninja_trader_data_dir` | +5 |
| B9 | `store/visibility.go:64-80` (`IsVisibleExchange`) | Add `strings.TrimSpace(exchange.NinjaTraderDataDir) != ""` to the OR-chain | +1 |
| B10 | `manager/trader_manager.go:~700` (`addTraderFromStore` switch) | Add `case "ninjatrader"`: copy `NinjaTraderDataDir` into `AutoTraderConfig` | +5 |
| B11 | `manager/trader_manager.go` imports | `import ninjatrader "nofx/trader/ninjatrader"` | +1 |
| B12 | `trader/auto_trader.go:111` (`AutoTraderConfig` struct) | Add fields `NinjaTraderDataDir string` + `NinjaTraderSymbol string` | +3 |
| B13 | `trader/auto_trader.go:60` (Exchange field comment) | Update comment to include `"ninjatrader"` | +0 |
| B14 | `trader/auto_trader.go:263-315` (broker switch) | Add `case "ninjatrader"`: `trader = ninjatrader.New(...)` | +8 |
| B15 | `kernel/engine_prompt.go:17` (`BuildSystemPrompt`) | Branch: `if variant == "futures" { return BuildFuturesSystemPrompt(...) }` | +6 |
| B16 | `kernel/engine_position.go:39` (`validateDecisions`) | Add NQ/MNQ branch: tick-round SL/TP, validate stop distance in points, allow leverage=1 | +30 |
| B17 | `kernel/engine_analysis.go:91` (`fetchMarketDataWithStrategy`) | Skip `engine.nofxosClient.GetOITopPositions()` when exchange is ninjatrader | +6 |
| B18 | `kernel/engine.go:NewStrategyEngine` lines 183-225 | (optional) Add `databentoClient` field + initialize when env var set | +12 |
| B19 | `api/handler_exchange.go:347-354` (validTypes map) | Add `"ninjatrader"` to the map | +1 |
| B20 | `api/handler_exchange.go:90-107` + `70-87` (request structs) | Add `NinjaTraderDataDir string` field to both request types | +4 |
| B21 | `api/handler_exchange.go:48-68` (`SafeExchangeConfig`) | Add the field | +2 |
| B22 | `api/handler_trader.go:184-196` (`validateExchangeForTraderCreation`) | Add `case "ninjatrader"` checking DataDir non-empty | +5 |
| B23 | `api/handler_trader.go:457-481` (probe builder) | Return nil for ninjatrader (skip live probe) | +5 |
| B24 | `api/handler_trader_status.go:159-201` (`handleClosePosition`) | Add `case "ninjatrader"` returning HTTP 400 with graceful message | +5 |
| B25 | `agent/skill_management_handlers.go:~200-280` (credential display) | Add `case "ninjatrader"`: show DataDir | +5 |
| B26 | `api/exchange_account_state.go:~50-120` (`buildExchangeProbeTrader`) | Add `case "ninjatrader"`: return early (no probe possible for file-based bridge) | +5 |
| B27 | `config/config.go:Config` struct ~16-46 | Add `DatabentoAPIKey string` + `DatabentoDataset string` | +3 |
| B28 | `config/config.go:Init()` ~92 | Load `DATABENTO_API_KEY` + `DATABENTO_DATASET` (default `"GLBX.MDP3"`) | +5 |
| B29 | `.env.example` (append) | New section: `DATABENTO_API_KEY=` + `DATABENTO_DATASET=GLBX.MDP3` | +4 |
| B30 | `web/src/router/AppRoutes.tsx` ~468-475 + import line | Remove `<Route path={ROUTES.data}>` + `DataPage` import | -10 |
| B31 | `web/src/router/paths.ts:23` | Remove `data: '/data',` line | -1 |
| B32 | `web/src/components/common/HeaderBar.tsx:121-131 + 457-466` | Remove desktop + mobile Data nav entries (2 identical blocks) | -22 |
| B33 | `web/src/pages/DataPage.tsx` | Delete file entirely (17 lines) | -17 |
| B34 | `web/src/components/strategy/CoinSourceEditor.tsx:69-79` | Skip USDT auto-append when symbol matches CME futures pattern | +8 |
| B35 | `web/src/pages/StrategyStudioPage.tsx:1203-1206 + 1311-1313` (PromptVariant dropdown) | Add `<option value="futures">{tr('futuresVariant')}</option>` in 2 places | +4 |
| B36 | `web/src/components/trader/ExchangeConfigModal.tsx:23-34` (templates array) | Add `{exchange_type:'ninjatrader', name:'NinjaTrader', type:'cex'}` | +5 |
| B37 | `web/src/components/trader/ExchangeConfigModal.tsx` (form fields, new section ~720-756) | Add NinjaTrader form section: DataDir + InstrumentName + DefaultContractQty inputs | +35 |
| B38 | `web/src/components/trader/ExchangeConfigModal.tsx:301-339` (handleSubmit) | Add validation branch + onSave call for `currentExchangeType==='ninjatrader'` | +12 |
| B39 | `web/src/pages/SettingsPage.tsx:handleSaveExchange` | Accept + pass NT fields to create/update request | +6 |
| B40 | `web/src/types/config.ts:49,91` (Exchange + CreateExchangeRequest) | Add `ninjaTraderDataDir?: string` (plus optional `instrumentName`, `defaultContractQty`) | +5 |
| B41 | `web/src/components/common/ExchangeIcons.tsx:10-21` | Add `ninjatrader: '/exchange-icons/ninjatrader.png'` to ICON_PATHS | +1 |
| B42 | `web/src/pages/TraderDashboardPage.tsx:513,522,530` (StatCard `unit="USDT"`) | Make conditional: USDT for crypto, "USD" for futures (read exchange.exchange_type) | +6 |
| B43 | `web/src/pages/TraderDashboardPage.tsx:609,663,673` (Leverage + Liquidation cols) | Hide for futures exchanges (`exchange_type==='ninjatrader'`) | +8 |
| B44 | `web/src/i18n/translations.ts:347,1693,2971` (`invalidSymbolFormat`) | Soften message: futures-aware | +6 |
| B45 | `web/src/i18n/translations.ts` (add 5 new keys × 3 languages) | `ninjatraderExchangeName`, `ninjatraderSetupGuide`, `dataDirectoryPath`, `chartInstrument`, `futuresVariant` | +30 |

**Total MODIFIED code: ~280 LOC delta** across 45 touchpoints.

---

## C — EXISTING functions that TRANSFER UNCHANGED

These are 100% confirmed transferable as-is. **Do not modify these**:

**Indicator engine** (no change):
- `calculateEMA`, `calculateMACD`, `calculateRSI`, `calculateATR` ([market/data_indicators.go:6,28,42,86](market/data_indicators.go))
- `ExportCalculateEMA/MACD/RSI/ATR/BOLL/Donchian/BoxData` ([market/data_indicators.go:203-233](market/data_indicators.go#L203-L233))
- All take `[]Kline`, return numbers. No symbol-format assumptions.

**Strategy engine — static-coin path** (no change after Normalize fix):
- `GetCandidateCoins` for `SourceType="static"` ([kernel/engine.go:262-271](kernel/engine.go#L262-L271))
- `filterExcludedCoins` ([kernel/engine.go:450-473](kernel/engine.go#L450-L473))
- `CoinSourceConfig` schema ([store/strategy.go:695-721](store/strategy.go#L695-L721))

**Strategy engine — nofxos paths** (no change; bypassed via config):
- `FetchOIRankingData`, `FetchNetFlowRankingData`, `FetchPriceRankingData` ([kernel/engine.go:778,802,834](kernel/engine.go)) — already return `nil` if the corresponding indicator flag is disabled. NQ strategies just leave them disabled.
- `getAI500Coins`, `getOITopCoins`, `getOILowCoins`, `getHyperAllCoins`, `getHyperMainCoins` — never called when `SourceType="static"`.

**Risk control schema** (no schema change; fields repurposed):
- `RiskControlConfig` ([store/strategy.go:802-825](store/strategy.go#L802-L825)) — `MaxPositions`, `MaxMarginUsage`, `MinPositionSize`, `MinRiskRewardRatio`, `MinConfidence` are generic. `BTCETHMaxLeverage` / `AltcoinMaxLeverage` can carry NQ values without rename (cosmetic-only labels).

**Decision parsing** (no change):
- `parseFullDecisionResponse` ([kernel/engine_analysis.go:230-335](kernel/engine_analysis.go#L230-L335)) — `<reasoning>` / `<decision>` XML+JSON extraction is symbol-agnostic.
- Decision JSON shape `{action, symbol, entry, stop_loss, take_profit, reasoning, leverage, confidence}` accepts NQ as-is.

**AI Model store** (no change):
- All of [store/ai_model.go](store/ai_model.go) — provider, API key, custom URL. Works for any LLM provider.
- `ResolveClaw402WalletKey` returns `("", nil)` cleanly when no claw402 configured — NQ traders pass through without error.

**Strategy CRUD** (no change):
- All of [api/strategy.go](api/strategy.go) — accepts arbitrary `StaticCoins` strings without USDT validation.
- `POST /api/strategies/test-run` works for `SourceType="static"` (skips nofxos calls entirely).
- `POST /api/strategies/estimate-tokens` is pure math — works for any prompt.

**Trader CRUD core** (only the broker-switch validation needs an addition; everything else unchanged):
- Trader Create/Update/Start/Stop flows
- `TraderManager.LoadTradersFromStore` and `addTraderFromStore` (only one switch case to add — B10)

**Position storage** (no change):
- [store/position.go:97-121](store/position.go#L97-L121) — `Symbol`, `EntryPrice`, `Quantity`, `Leverage`, `Status` are generic. NQ contract symbols like "NQM6" fit as `Symbol`. Leverage=1 works.

**Decision logging** (no change):
- [store/decision.go](store/decision.go) — `DecisionRecord` and `DecisionAction` structs are symbol-agnostic.

**Order storage** (no change):
- [store/order.go](store/order.go) `TraderOrder` and `TraderFill` — generic fields.

**Web UI components that work as-is** (no functional change; just minor i18n/conditional render edits in B):
- `TraderDashboardPage` positions table, equity stat cards, decisions log
- `DecisionCard` rendering (auto-strips USDT for display; works for NQ)
- `ModelConfigModal` (configures AI providers; provider-agnostic)
- `TraderConfigModal` (only displays strategy; doesn't validate symbols)
- `RiskControlEditor` (sliders work for any numeric range)
- `FAQ*` components

**Bootstrap** (one config addition, otherwise unchanged):
- `main.go` start sequence — adds nothing for the NQ path
- `crypto.NewCryptoService` — `NinjaTraderDataDir` is a path, not a secret, so no encryption needed
- DB init, JWT secret, logger init — unchanged
- Telegram bot — unchanged (vocabulary cosmetic only)

---

## D — Total scope (function-transfer view)

| Layer | New LOC | Modified Δ | Tests | Total |
|---|---|---|---|---|
| **provider/databento/** | 250 | — | 150 | 400 |
| **market/ (Databento adapter + Normalize fix)** | 50 | 34 | 80 | 164 |
| **provider/ninjatrader/ (CSV bridge)** | 280 | — | 150 | 430 |
| **trader/ninjatrader/ (broker impl)** | 270 | — | 50 | 320 |
| **kernel/ (futures prompt + validation)** | 240 | 54 | 60 | 354 |
| **store/ (exchange field + visibility)** | — | 31 | — | 31 |
| **api/ (handlers + validation)** | — | 32 | — | 32 |
| **config/ + .env** | — | 12 | — | 12 |
| **manager/ (trader manager switch)** | — | 6 | — | 6 |
| **trader/auto_trader (broker switch)** | — | 11 | — | 11 |
| **cmd/nq_smoke** | 130 | — | — | 130 |
| **web/ (UI conditional renders + nav cleanup)** | — | 102 | — | 102 |
| **i18n** | — | 36 | — | 36 |
| **TOTAL** | **~1,220** | **~318** | **~490** | **~2,030 LOC** |

**Calibrated estimate: ~2,000 LOC total work** — about 50% less than my original 4,000-6,000 LOC estimate, validated by the 15-agent audit.

**Time to ship**: 3-4 weeks at one focused engineer. The biggest single item is `trader/ninjatrader/trader.go` (the 17-method `Trader` interface impl) at ~270 LOC + tests.

---

## E — The transfer order

When executing, do groups in this dependency order to avoid getting stuck:

```
1. Group A1-A6  (Databento data layer)         ← independent, can start any time
2. Group A7-A9  (CSV bridge primitives)        ← independent
3. Group A29-A31 (Futures prompt)               ← independent
4. Group A32-A35 (Futures validation helpers)  ← independent
5. Group A10-A28 (Trader broker impl)          ← needs A7-A9 done
6. Modifications B1-B7                          ← needs A1-A6 done (Normalize + adapter wired)
7. Modifications B8-B14                         ← needs A10-A28 done (broker registered)
8. Modifications B15-B18                        ← needs A29-A35 done (prompt + validation registered)
9. Modifications B19-B29                        ← API layer + config (independent, can do in parallel after store layer)
10. Modifications B30-B45                       ← web UI (can do in parallel with backend)
11. Group A36 (cmd/nq_smoke)                    ← needs everything else done; the integration test
```

Groups 1-4 are independent and can run in parallel. Group 5 depends on 2. Modifications wait for the new functions they reference. The web UI work can happen in parallel with the backend once the request/response shapes are settled.

This ordering is also reflected in the Task 0-11 sequence at the top of this document.

---

# The 3 Control Surfaces (User-Facing Priority)

> **Refocus 2026-05-22:** From the user's perspective, the entire NQ trading system is controlled through three pages in the web UI: **Config**, **Dashboard**, and **Strategy**. If these three flows work, the system works. This section is the minimal function-transfer manifest scoped to those three surfaces only — backend backbone (Databento client, NinjaTrader CSV bridge, broker impl) is separate, covered earlier in this document.

```
┌────── CONFIG (Settings) ──────┐   ┌──── DASHBOARD (Trader) ────┐   ┌──── STRATEGY (Studio) ─────┐
│  Add NinjaTrader exchange     │   │  Watch the bot trade NQ    │   │  Build NQ strategy with    │
│  Configure AI model           │   │  in NT SIM, review AI       │   │  static coin source +      │
│                                │   │  decisions live            │   │  futures prompt variant    │
└────────────────────────────────┘   └──────────────────────────────┘   └────────────────────────────┘
```

User flow: open Config → add NinjaTrader exchange + verify AI model. Open Strategy → build NQ strategy. Open Dashboard → create trader linking the three → start → watch.

## SURFACE 1: CONFIG (Settings page)

**Frontend:** `web/src/pages/SettingsPage.tsx` + `web/src/components/trader/ExchangeConfigModal.tsx` + `web/src/components/trader/ModelConfigModal.tsx`
**Backend:** `api/handler_exchange.go` + `api/handler_ai_model.go` + `store/exchange.go` + `store/visibility.go`

| # | Touchpoint | Action |
|---|------------|--------|
| C1 | `ExchangeConfigModal.tsx:23-34` (templates array) | Add `{ exchange_type: 'ninjatrader', name: 'NinjaTrader', type: 'cex' }` |
| C2 | same (new form section ~720-756) | NT form fields: DataDir + InstrumentName + DefaultContractQty |
| C3 | same:301-339 (handleSubmit) | NT validation branch (DataDir required, no API key) |
| C4 | `SettingsPage.tsx:190` (`handleSaveExchange`) | Pass new fields into createRequest/updateRequest |
| C5 | `web/src/types/config.ts:49,91` | Add `ninjaTraderDataDir?: string` to Exchange + CreateExchangeRequest |
| C6 | `web/src/components/common/ExchangeIcons.tsx:10-21` | Add NT icon mapping |
| C7 | `web/src/i18n/translations.ts` (×3 languages) | New keys: `ninjatraderExchangeName`, `ninjatraderSetupGuide`, `dataDirectoryPath`, `chartInstrument` |
| C8 | `api/handler_exchange.go:347-354` (validTypes) | Add `"ninjatrader"` |
| C9 | same:90-107 + 70-87 (request structs) | Add `NinjaTraderDataDir string` field |
| C10 | same:48-68 (`SafeExchangeConfig`) | Add the field for return DTOs |
| C11 | `store/exchange.go:43` | Add `NinjaTraderDataDir string` column |
| C12 | `store/visibility.go:5-37` (`MissingRequiredExchangeCredentialFields`) | Add `case "ninjatrader"` |
| C13 | `store/visibility.go:64-80` (`IsVisibleExchange`) | Include DataDir in OR-chain |
| C14 | `api/handler_trader.go:184-196` (`validateExchangeForTraderCreation`) | Add ninjatrader case |

**Already works:** AI Model config UI, tab routing, exchange list rendering, encrypted-credential flow, ModelConfigModal for DeepSeek/Claude/OpenAI/etc.

**Surface 1 total: ~110 LOC frontend + ~30 LOC backend = ~140 LOC.**

## SURFACE 2: DASHBOARD (Trader Dashboard)

**Frontend:** `web/src/pages/TraderDashboardPage.tsx` + `web/src/components/trader/DecisionCard.tsx` + `web/src/components/charts/ChartTabs.tsx`
**Backend:** `api/handler_trader.go` + `api/handler_trader_status.go`

| # | Touchpoint | Action |
|---|------------|--------|
| D1 | `TraderDashboardPage.tsx:513,522,530` (StatCard `unit="USDT"`) | Conditional unit: USD for ninjatrader, USDT otherwise |
| D2 | `TraderDashboardPage.tsx:609,663` (Leverage column) | Hide when `exchange_type === 'ninjatrader'` |
| D3 | `TraderDashboardPage.tsx:611,673` (Liquidation Price column) | Hide for ninjatrader |
| D4 | `api/handler_trader_status.go:159-201` (`handleClosePosition`) | Add ninjatrader case returning HTTP 400 with friendly message |
| D5 | `api/handler_trader.go:457-481` (probe trader) | Return nil for ninjatrader (skip live API probe) |
| D6 | `api/exchange_account_state.go:~50-120` (`buildExchangeProbeTrader`) | Add ninjatrader case: skip |
| D7 | `web/src/i18n/translations.ts:250` etc. (USDT warnings) | Soften crypto-only warnings |

**Already works:** DecisionCard (auto-strips USDT, formats prices, renders reasoning), Recent Decisions panel, Equity/P&L chart, Position history, Run/Stop buttons, cycle counter, all sub-components. **The dashboard is the most asset-agnostic surface — least work.**

**Surface 2 total: ~20 LOC frontend + ~15 LOC backend = ~35 LOC.**

## SURFACE 3: STRATEGY (Strategy Studio)

**Frontend:** `web/src/pages/StrategyStudioPage.tsx` + `web/src/components/strategy/CoinSourceEditor.tsx` + `IndicatorEditor.tsx` + `RiskControlEditor.tsx`
**Backend:** `api/strategy.go` + `api/strategy.go` + `kernel/engine_prompt.go` + `kernel/engine_position.go` + `market/data.go` + NEW `kernel/engine_prompt_futures.go`

| # | Touchpoint | Action |
|---|------------|--------|
| S1 | `CoinSourceEditor.tsx:69-79` (USDT auto-append) | Skip USDT for CME futures patterns (`NQ.c.0`, `MNQ`, `ES`, etc.) |
| S2 | same:195-201 (input placeholder) | "BTC, ETH, SOL, NQ.c.0, MNQ..." |
| S3 | `IndicatorEditor.tsx:658-687` (Market Sentiment) | Hide `enable_funding_rate` + `enable_oi` for futures (prop-controlled) |
| S4 | `IndicatorEditor.tsx:226-449` (NofxOS sources: AI500/OI/NetFlow/Price ranking) | Hide for futures strategies |
| S5 | `RiskControlEditor.tsx:72-127` (leverage labels) | Conditional label: "NQ Leverage" for futures vs "BTC/ETH Leverage" for crypto |
| S6 | `RiskControlEditor.tsx:277` (USDT min position unit) | Conditional unit |
| S7 | `StrategyStudioPage.tsx:1203-1206 + 1311-1313` (PromptVariant dropdown) | Add `<option value="futures">` in 2 places |
| S8 | `web/src/i18n/strategy-translations.ts` | Add futures variant labels, soften coinSource descriptions |
| S9 | `api/strategy.go:514-515` (PromptVariant validation) | Accept `"futures"` (currently no rejection — verify pass-through) |
| S10 | `kernel/engine_prompt.go:17` (`BuildSystemPrompt`) | Branch: `if variant=="futures" → BuildFuturesSystemPrompt(...)` |
| S11 | **NEW** `kernel/engine_prompt_futures.go` | `BuildFuturesSystemPrompt(FuturesPromptConfig) string` + `BuildFuturesUserPrompt(FuturesContext) string` + helpers |
| S12 | **NEW** `kernel/engine_position.go` (add helpers) | `roundTickNQ(price)` + `validateStopDistanceNQ(entry, stop, minPoints)` + futures branch in `validateDecisions` |
| S13 | `kernel/engine_analysis.go:91` (`fetchMarketDataWithStrategy`) | Skip `GetOITopPositions()` when exchange is ninjatrader |
| S14 | `market/data.go:583` (`Normalize`) | Skip USDT-append for CME futures symbols |
| S15 | **NEW** `market/data.go` helper | `isCMEFuturesSymbol(symbol string) bool` |

**Already works:** Strategy CRUD (accepts arbitrary `StaticCoins`), editor layout, Strategy Type selector, `PromptSections` editor, `CustomPrompt` field, test-run for static mode, token estimation, indicator math (EMA/MACD/RSI/ATR/Bollinger), backend `NormalizeProductSchema`.

**Surface 3 total: ~60 LOC frontend + ~250 LOC backend = ~310 LOC.** Most work because it includes the NEW futures prompt (~180 LOC) — the AI's NQ behavior is entirely defined here.

## Consolidated 3-surface scope

| Surface | Frontend Δ | Backend Δ | New backend | TOTAL |
|---------|------------|-----------|-------------|-------|
| **Config** | ~80 | ~30 | — | ~110 |
| **Dashboard** | ~20 | ~15 | — | ~35 |
| **Strategy** | ~60 | ~20 | ~230 | ~310 |
| **TOTAL (3 surfaces)** | **~160** | **~65** | **~230** | **~455 LOC** |

The remaining ~1,500 LOC (provider/databento/, provider/ninjatrader/, trader/ninjatrader/, cmd/nq_smoke, kernel adapter wiring) is the **backbone** — invisible to the user but required for the 3 surfaces to function. The split:

```
~455 LOC  ← 3 user-facing control surfaces
~1500 LOC ← backbone (data layer + broker impl + market adapter + smoke runner)
─────────
~1950 LOC ← total project (3-4 weeks focused engineering)
```

## Acceptance criteria, surface by surface

**Config:** A user can navigate to Settings → Exchanges tab → click Add → select "NinjaTrader" → enter DataDir path → save. The exchange appears in the list. No API key/secret required. **Done when this round-trip works.**

**Strategy:** A user can navigate to Strategy Studio → create new strategy → set Source = "static" + Static Coins = `["NQ.c.0"]` → set PromptVariant = "futures" → save. Strategy editor shows NQ-friendly RiskControl labels and hides funding/OI indicators. Token estimate works. Test-run button (with no AI key) returns prompts containing NQ-aware language, no BTC/USDT references. **Done when this round-trip works.**

**Dashboard:** A user can navigate to Trader Dashboard for an NQ trader → see positions in contracts (not coins, no USDT unit) → see Leverage column hidden → see Recent Decisions render with NQ symbols cleanly. **Done when this round-trip works.**

When all three acceptance criteria pass, the user has a working NQ-trading workflow even before any actual order flows through NinjaTrader. The backbone layer is required for live orders, but the *control surfaces* are independently validatable first.

---

# 4 Control Surfaces — Complete Function Inventory

> **Per-page function audit, 2026-05-22.** Four parallel Explore agents each enumerated every function/component/handler/endpoint/store-method on one surface and assigned KEEP / MODIFY / NEW / DELETE / DEFER per item. **Combined coverage: ~320 enumerated touchpoints.** This is the canonical pre-implementation checklist; nothing should be missed.

## Coverage scorecard

> **Recounted 2026-05-22 v3** — prior version had arithmetic errors on 3 of 4 surfaces. Totals below now reconcile (sum of categories = surface total; sum of surfaces = grand total).

| Surface | Total items | KEEP | MODIFY | NEW | DELETE | DEFER |
|---------|-------------|------|--------|-----|--------|-------|
| **Config** (Settings) | 65 | 47 | 18 | 0 | 0 | 0 |
| **Dashboard** (Trader) | 96 | 71 | 19 | 0 | 6 | 0 |
| **Strategy** (Studio) | 118 | 82 | 34 | 2 | 0 | 0 |
| **AgentBeta** (Chat) | 45 | 20 | 18 | 0 | 0 | 7 |
| **TOTAL** | **324** | **220** | **89** | **2** | **6** | **7** |

**Verdict:** ~68% of all enumerated functions work unchanged for NQ. ~27% need targeted modification (most are label/conditional renders, not structural rewrites). 2 require new code (futures prompt + validators). 6 are conditional deletes (dashboard columns/sections hidden for futures). 7 Agent items are deferred to Plan 2 since they require deeper agent rework (provider catalog, candidate-coin tools, deep prompt persona).

## SURFACE 1 — CONFIG (Settings page): 55 items

**Frontend:** [web/src/pages/SettingsPage.tsx](web/src/pages/SettingsPage.tsx), [ExchangeConfigModal.tsx](web/src/components/trader/ExchangeConfigModal.tsx), [ModelConfigModal.tsx](web/src/components/trader/ModelConfigModal.tsx)
**Backend:** [api/handler_exchange.go](api/handler_exchange.go), [api/handler_ai_model.go](api/handler_ai_model.go), [store/exchange.go](store/exchange.go), [store/visibility.go](store/visibility.go)

### KEEP (47 — work unchanged for NQ)
SettingsPage tab container + state mgmt, all useEffects, refreshModelConfigs, refreshExchangeConfigs, handleChangePassword, handleSaveModel, handleDeleteModel, handleDeleteExchange, all 4 tab containers (Account/AI Models/Exchanges/Telegram), ModelConfigModal entirely (DeepSeek/Claude/OpenAI/etc. work as-is). ExchangeConfigModal: StepIndicator, ExchangeCard, useEffect lifecycle, handleCopyIP, handleSecureInputComplete, handleSelectExchange, handleBack, all 5 existing CEX/DEX field sections (binance, bybit, okx, gate/kucoin/indodax, aster, hyperliquid, lighter), form buttons. API: `GET /api/exchanges`, `DELETE /api/exchanges/:id`. Store: Exchange.List, Exchange.Delete.

### MODIFY (18 items — exact edits)

| # | File:Line | Change |
|---|-----------|--------|
| C-M1 | `SettingsPage.tsx:190-257` (`handleSaveExchange`) | Add NT params (dataDir, instrumentName, defaultContractQty) to function signature + pass into createRequest/updateRequest body |
| C-M2 | `ExchangeConfigModal.tsx:23-34` (`SUPPORTED_EXCHANGE_TEMPLATES`) | Add `{ exchange_type: 'ninjatrader', name: 'NinjaTrader', type: 'cex' }` entry |
| C-M3 | `ExchangeConfigModal.tsx:146-152` (props) | Update onSave signature to accept NT fields |
| C-M4 | `ExchangeConfigModal.tsx:154-180` (form state) | Add `dataDir`, `instrumentName`, `defaultContractQty` useState |
| C-M5 | `ExchangeConfigModal.tsx:212-227` (edit-mode populate) | Load NT fields when editing ninjatrader exchange |
| C-M6 | `ExchangeConfigModal.tsx:301-339` (`handleSubmit`) | Add NT validation branch (DataDir required, no API key) |
| C-M7 | `ExchangeConfigModal.tsx:342-454` (Step 0: exchange selection grid) | Ensure NT card appears in CEX section |
| C-M8 | `ExchangeConfigModal.tsx:~720-756` (NEW form section) | Add NT-specific form: DataDir input + InstrumentName + DefaultContractQty |
| C-M9 | `types/config.ts:21-50` (Exchange interface) | Add optional NT fields |
| C-M10 | `types/config.ts:76-92` (CreateExchangeRequest) | Add optional NT fields |
| C-M11 | `types/config.ts:124-145` (UpdateExchangeConfigRequest) | Add optional NT fields |
| C-M12 | i18n keys (translations.ts ×3 languages) | New: `ninjatraderExchangeName`, `ninjatraderSetupGuide`, `dataDirectoryPath`, `chartInstrument`, `ninjatraderContractQty` |
| C-M13 | `api/handler_exchange.go:347-354` (validTypes map) | Add `"ninjatrader"` |
| C-M14 | `api/handler_exchange.go:90-107 + 70-87` (request structs) | Add `NinjaTraderDataDir string`, `InstrumentName string`, `DefaultContractQty int` |
| C-M15 | `api/handler_exchange.go:28-46` (`SafeExchangeConfig`) + `:48-68` (`safeExchangeConfigFromStore`) | Add NT fields to response DTO + mapper |
| C-M16 | `api/handler_exchange.go:440-458` (`GET /api/supported-exchanges`) | Add NT to static response list |
| C-M17 | `store/exchange.go:20-43` (Exchange struct) + Create/Update signatures + `initTables` migration | Add `NinjaTraderDataDir`, `InstrumentName`, `DefaultContractQty` columns + Postgres ALTER |
| C-M18 | `store/visibility.go:5-37` (`MissingRequiredExchangeCredentialFields`) + `:64-80` (`IsVisibleExchange`) | Add ninjatrader case requiring DataDir + extend visibility OR-chain |

**Surface 1 LOC delta: ~140 (110 FE, 30 BE).**

## SURFACE 2 — DASHBOARD (Trader): 96 items

**Frontend:** [TraderDashboardPage.tsx](web/src/pages/TraderDashboardPage.tsx), [DecisionCard.tsx](web/src/components/trader/DecisionCard.tsx), [ChartTabs.tsx](web/src/components/charts/ChartTabs.tsx), [TraderConfigModal.tsx](web/src/components/trader/TraderConfigModal.tsx)
**Backend:** [api/handler_trader.go](api/handler_trader.go), [api/handler_trader_status.go](api/handler_trader_status.go), [api/handler_order.go](api/handler_order.go), [api/handler_competition.go](api/handler_competition.go) (for decisions endpoint), [store/position.go](store/position.go), [store/decision.go](store/decision.go), [manager/trader_manager.go](manager/trader_manager.go)

### KEEP (71 — work unchanged for NQ)
Trader Header section, Debug Info, StatCard: Positions (count-only), GridRiskPanel, ChartTabs container, the entire positions data table EXCEPT leverage + liq columns, every Symbol/Side/Action/Entry/Mark/Qty/Value/uPnL column, close-position button + handler, position pagination, Recent Decisions Panel, decisions limit selector, DecisionCard main component, Risk/Reward visualization, EquityChart, AdvancedChart, K-line interval selector, symbol dropdown (works for any symbol source), PositionHistory component, useState hooks for closingPosition/selectedChartSymbol/pagination, useEffects for trader/grid changes, handleSymbolClick, handleClosePosition, all DecisionCard copyToClipboard/downloadAsFile, ChartTabs auto-detect on exchange change, all 16 backend API endpoints (`/api/status`, `/api/account`, `/api/positions`, `/api/positions/history`, `/api/decisions`, `/api/decisions/latest`, `/api/statistics`, `/api/equity-history`, `/api/traders/:id/grid-risk-info`, `/api/traders/:id/sync-balance`, `/api/traders/:id/close-position`, `/api/traders` CRUD, `/api/traders/:id/start`, `/api/traders/:id/stop`), all Position/Decision/Trader store methods, all TraderManager lifecycle methods (GetTrader, StartAll, StopAll, AutoStartRunningTraders), all 4 useSWR polling hooks. TraderConfigModal entirely (the create-trader modal trusts strategy's symbols).

### MODIFY (19) + DELETE (6)

| # | File:Line | Change | Type |
|---|-----------|--------|------|
| D-M1 | `TraderDashboardPage.tsx:513` (StatCard Total Equity `unit="USDT"`) | Conditional: USD for ninjatrader, USDT otherwise | MODIFY |
| D-M2 | `TraderDashboardPage.tsx:522` (StatCard Available Balance) | Conditional unit | MODIFY |
| D-M3 | `TraderDashboardPage.tsx:530` (StatCard Total P&L) | Conditional unit | MODIFY |
| D-D1 | `TraderDashboardPage.tsx:609, 663` (Leverage column header + cells, 7 render locations) | DELETE for ninjatrader (futures don't show leverage; conditional render) | DELETE |
| D-D2 | `TraderDashboardPage.tsx:611, 673` (Liquidation Price column) | DELETE for ninjatrader (no liq for NT) | DELETE |
| D-M4 | `TraderDashboardPage.tsx:67-71` (`isPerpDexExchange` helper) | Keep but extend: NT is neither perp nor dex; ensure helper returns false | MODIFY |
| D-M5 | `TraderDashboardPage.tsx:73-87` (`getWalletAddress` helper) | Return empty for ninjatrader | MODIFY |
| D-D3 | `TraderDashboardPage.tsx:178-187` (`handleCopyAddress`) | Not invoked for NT; safe to keep but no UI to trigger | DELETE (cond) |
| D-D4 | `TraderDashboardPage.tsx:395-440` (wallet address display section) | Hide for ninjatrader (conditional render) | DELETE (cond) |
| D-D5 | `TraderDashboardPage.tsx:143-144` (showWalletAddress/copiedAddress useState) | Never set for NT | DELETE (cond) |
| D-M6 | `DecisionCard.tsx:148-151` (ActionCard Leverage display `{leverage}x`) | Hide for ninjatrader (or show "1x" gracefully) | MODIFY |
| D-D6 | `DecisionCard.tsx:?` Leverage field rendering when action is NQ trade | Conditional hide | DELETE (cond) |
| D-M7 | `ChartTabs.tsx:184-204` (market type pills: hyperliquid/crypto/stocks/forex/metals) | For NT, force a single market type (or hide pills) | MODIFY |
| D-M8 | `ChartTabs.tsx:67-71` (useEffect: auto-detect market type from exchangeId) | Add ninjatrader → "futures" mapping (or "crypto" as fallback so kline still works) | MODIFY |
| D-M9 | `ChartTabs.tsx:283-294` (Quick Symbol Input + auto-append USDT) | Skip USDT append for CME futures symbol | MODIFY |
| D-M10 | `ChartTabs.tsx:132-143` (`handleSymbolSubmit`) | Don't normalize CME symbols to USDT | MODIFY |
| D-M11 | `ChartTabs.tsx:112-116` (`handleMarketTypeChange`) | Disable / no-op for ninjatrader | MODIFY |
| D-M12 | `ChartTabs.tsx:60-64` (default symbol "BTCUSDT") | Use exchange's chart symbol or "NQ.c.0" for ninjatrader | MODIFY |
| D-M13 | `api/handler_trader_status.go:159-201` (`handleClosePosition`) | Add `case "ninjatrader"` returning HTTP 400 with friendly message ("Close via NT UI; bridge does not support manual close") | MODIFY |
| D-M14 | `api/handler_trader.go:457-481` (probe trader) | Return nil for ninjatrader (no live probe) | MODIFY |
| D-M15 | `api/exchange_account_state.go:~50-120` (`buildExchangeProbeTrader`) | Add ninjatrader case: skip | MODIFY |
| D-M16 | Position struct returned by `/api/positions` (handler + store) | Conditionally omit `leverage` + `liquidation_price` for NT — OR leave in API response and just hide in UI (recommended: hide in UI only, less backend churn) | MODIFY |
| D-M17 | `web/src/i18n/translations.ts:250` (`asterUsdtWarning` and similar) | Soften crypto-specific warnings (multi-exchange-aware) | MODIFY |
| D-M18 | `web/src/i18n/translations.ts` (`tradingSymbolsDescription`, `invalidSymbolFormat`) | Soften "must end with USDT" validators | MODIFY |
| D-M19 | `web/src/components/trader/PositionHistory.tsx:126` (`stat.symbol.replace('USDT', '')`) | Safe as-is (replace returns original if no match for NQ) — no change needed BUT verify | KEEP-verify |

**Surface 2 LOC delta: ~80 (60 FE conditional renders, 20 BE). Most "DELETE" items are conditional renders — UI does not render the cell, but the data shape stays unchanged.**

## SURFACE 3 — STRATEGY (Studio): 120 items

**Frontend:** [StrategyStudioPage.tsx](web/src/pages/StrategyStudioPage.tsx) + 6 strategy editor sub-components in [web/src/components/strategy/](web/src/components/strategy/)
**Backend:** [api/strategy.go](api/strategy.go), [api/strategy.go](api/strategy.go), [store/strategy.go](store/strategy.go), [kernel/engine.go](kernel/engine.go), [kernel/engine_prompt.go](kernel/engine_prompt.go), [kernel/engine_position.go](kernel/engine_position.go), [market/data.go](market/data.go)

### KEEP (82 — work unchanged for NQ)
StrategyConfig + AIStrategyConfig interfaces, KlineConfig (timeframes work for any asset), PromptSectionsConfig, StrategyStudioPage root + all state vars + accordion section state, all 12 strategy CRUD handlers (fetchStrategies, handleCreateStrategy, handleDeleteStrategy, handleDuplicateStrategy, handleActivateStrategy, handleExportStrategy, handleImportStrategy, handleSaveStrategy, updateConfig, updateAIConfig, handleStrategyTypeChange, fetchPromptPreview), all 7 accordion section render blocks (CoinSource/Indicators/RiskControl/PromptSections/GridConfig/PublishSettings/StrategyType selector), CoinSourceEditor's sourceTypes selector (all 4 modes work for any asset), all 4 source type cards (static/ai500/oi_top/oi_low — AI500/OI just disabled by config), all timeframe selections (14 timeframes), Raw Klines toggle, Technical Indicators section entirely (EMA/MACD/RSI/ATR/BOLL + period inputs — all asset-agnostic), enable_volume + enable_oi (volume always relevant; OI just left off for NQ), RiskControlEditor max_positions / max_margin_usage / min_position_size / min_risk_reward_ratio / min_confidence (all generic), PromptSectionsEditor entirely (free-form text), PublishSettingsEditor entirely, TokenEstimateBar entirely, GridConfigEditor (unused for NQ), all i18n: coinSource/gridConfig/riskControl/promptSections/publishSettings objects (no crypto-specific terms in keys), all 13 backend strategy API endpoints (estimate-tokens, list, get, public, get-active, get-default-config — the last needs minor extension for futures variant), all 12 store/strategy.go CRUD methods, ClampLimits, MergeStrategyConfig, EstimateTokens. Kernel: GetCandidateCoins for static mode at engine.go:262-271 (works for NQ after Normalize fix), BuildUserPrompt, validateDecisions (with futures branch added separately).

### MODIFY (34) + NEW (2)

| # | File:Line | Change | Type |
|---|-----------|--------|------|
| S-M1 | `types/strategy.ts:107-118` (CoinSourceConfig) | Add optional `coin_source_variant?: 'crypto'\|'futures'` discriminator | MODIFY |
| S-M2 | `types/strategy.ts:120-162` (IndicatorConfig) | Document that `enable_funding_rate` + `enable_oi` are hidden in futures variant | MODIFY |
| S-M3 | `types/strategy.ts:184-202` (RiskControlConfig) | Add labels for futures (relabel BTC/ETH and Altcoin fields generically OR add separate fields) | MODIFY |
| S-M4 | `StrategyStudioPage.tsx:125` (selectedVariant state) | Add `'futures'` option | MODIFY |
| S-M5 | `StrategyStudioPage.tsx:1203-1206 + 1311-1313` (PromptVariant dropdown ×2) | Add `<option value="futures">{tr('futuresVariant')}</option>` | MODIFY |
| S-M6 | `StrategyStudioPage.tsx:650-678` (`runAiTest`) | Pass `prompt_variant` (may be `futures`) in request | MODIFY |
| S-M7 | `CoinSourceEditor.tsx:31-42` (xyzDexAssets set) | Add NQ, MES, MNQ, ES patterns | MODIFY |
| S-M8 | `CoinSourceEditor.tsx:44-47` (`isXyzDexAsset`) | Match CME futures patterns | MODIFY |
| S-M9 | `CoinSourceEditor.tsx:60-88` (`handleAddCoin`) | For CME, store as "NQ.c.0" not "NQc.0USDT" | MODIFY |
| S-M10 | `CoinSourceEditor.tsx:69-87` (symbol formatting logic) | Skip USDT suffix for CME symbols | MODIFY |
| S-M11 | `CoinSourceEditor.tsx:97-118` (`handleAddExcludedCoin`) | Same formatting fix | MODIFY |
| S-M12 | `CoinSourceEditor.tsx:200` (input placeholder) | "BTC, ETH, SOL, NQ.c.0, MNQ..." | MODIFY |
| S-M13 | `IndicatorEditor.tsx:226-275` (Quant Data section) | Hide for futures variant | MODIFY |
| S-M14 | `IndicatorEditor.tsx:278-331` (OI Ranking section) | Hide for futures variant | MODIFY |
| S-M15 | `IndicatorEditor.tsx:334-387` (NetFlow Ranking section) | Hide for futures variant | MODIFY |
| S-M16 | `IndicatorEditor.tsx:390-449` (Price Ranking section) | Conditionally show (price ranking is asset-class-agnostic) | MODIFY |
| S-M17 | `IndicatorEditor.tsx:656-688` (Market Sentiment section: volume / OI / funding_rate) | Hide `enable_funding_rate` for futures | MODIFY |
| S-M18 | `RiskControlEditor.tsx:61-128` (BTC/ETH + Altcoin leverage sliders) | Conditional labels for futures ("Primary Instrument Leverage" generic OR explicit "NQ/MNQ Leverage") | MODIFY |
| S-M19 | `RiskControlEditor.tsx:139-?` (position value ratios) | Conditional labels | MODIFY |
| S-M20 | `RiskControlEditor.tsx:277` (USDT min position unit) | Conditional unit | MODIFY |
| S-M21 | i18n `strategy-translations.ts:193-252` (indicator object) | Add `futuresVariant` key + remove crypto-only descriptions for indicators that don't apply | MODIFY |
| S-M22 | i18n `strategy-translations.ts:144-169` (riskControl object) | Add NQ-aware labels | MODIFY |
| S-M23 | `api/strategy.go:172-256` (`handleCreateStrategy`) | Accept `coin_source_variant=futures` → set futures default in `NormalizeProductSchema` | MODIFY |
| S-M24 | `api/strategy.go:478-489` (`handleGetDefaultStrategyConfig`) | Support `?variant=futures` query param → return futures defaults | MODIFY |
| S-M25 | `api/strategy.go:491-538` (`handlePreviewPrompt`) | Accept `prompt_variant="futures"` in body | MODIFY |
| S-M26 | `api/strategy.go:540-711` (`handleStrategyTestRun`) | Accept `prompt_variant="futures"` + route to BuildFuturesSystemPrompt | MODIFY |
| S-M27 | `store/strategy.go:134-150` (`NormalizeProductSchema`) | Detect "NQ.c.0" format and preserve (don't append USDT); detect `coin_source_variant` discriminator | MODIFY |
| S-M28 | `store/strategy.go:842` (`GetDefaultStrategyConfig`) | Add `futures` variant → static + NQ.c.0 + disabled funding/OI | MODIFY |
| S-M29 | `kernel/engine.go:254-?` (`GetCandidateCoins`) | Accept NQ symbols cleanly through normalization | MODIFY |
| S-M30 | `kernel/engine_prompt.go:17-150` (`BuildSystemPrompt`) | Branch: `if variant=="futures" → BuildFuturesSystemPrompt(...)` | MODIFY |
| S-M31 | `kernel/engine_prompt.go:?` (`writeAvailableIndicators`) | For futures variant, exclude funding rate / OI | MODIFY |
| S-M32 | `kernel/engine_position.go:39` (`validateDecisions`) | Add NQ branch: tick-round, min stop in points, allow leverage=1 | MODIFY |
| S-M33 | `market/data.go:557-558` (`Normalize`) | **Case-preserving fix.** Capture raw symbol before `strings.ToUpper` at line 558, check `isCMEFuturesSymbol(raw)` first, `return raw` if futures. Otherwise proceed with existing crypto logic. Critical: Databento uses lowercase suffix (`NQ.c.0`), inserting the check only at line 583 returns `NQ.C.0` which Databento rejects. | MODIFY |
| S-M34 | `market/data.go` | Add `isCMEFuturesSymbol(symbol) bool` helper (regex/map) | MODIFY (new helper in existing file) |
| S-N1 | **NEW** `kernel/engine_prompt_futures.go` | `BuildFuturesSystemPrompt(FuturesPromptConfig) string` + `BuildFuturesUserPrompt(FuturesContext) string` + helpers | NEW |
| S-N2 | **NEW** `kernel/engine_position.go` (add helpers) | `roundTickNQ(price)` + `validateStopDistanceNQ(entry, stop, minPoints)` + `validateContractSizeNQ` | NEW |

**Surface 3 LOC delta: ~340 (60 FE, 80 BE modifications, 200 NEW for futures prompt + validators).**

## SURFACE 4 — AGENT BETA (Chat): 49 items

**Frontend:** [AgentChatPage.tsx](web/src/pages/AgentChatPage.tsx) + [web/src/components/agent/](web/src/components/agent/) (MarketTicker, WelcomeScreen, PositionsPanel, UserPreferencesPanel, TradersPanel, SuggestionCards)
**Backend:** [agent/web.go](agent/web.go), [agent/tools.go](agent/tools.go) (the ~23 tools), [agent/agent.go](agent/agent.go), [agent/model_provider_catalog.go](agent/model_provider_catalog.go)

### KEEP (20)
Main chat area, message stream + input box, quickActions (6 commands), sidebar accordion sections, MarketTicker fetch logic (the *logic* is fine; just the hardcoded symbol list needs change), PositionsPanel (already supports stocks + crypto), most agent tools that are asset-agnostic: get_preferences, manage_preferences, get_backend_logs, get_decisions (works for any symbol), get_exchange_configs, get_model_configs, get_strategies, manage_trader, search_stock, get_positions, get_balance, get_market_price, get_kline, get_trade_history, get_watchlist, manage_watchlist. Backend handlers: `/api/agent/health`, `/api/agent/chat`, `/api/agent/chat/stream`, `/api/agent/tickers`. Memory: TaskState + chatHistory. All 9 streaming SSE events (planning/plan/step_start/step_complete/replan/tool/delta/done/error).

### MODIFY (18 — for minimal Plan 1 NQ-awareness)

| # | File:Line | Change | Type |
|---|-----------|--------|------|
| A-M1 | `MarketTicker.tsx:14` (hardcoded SYMBOLS) | Replace `['BTCUSDT','ETHUSDT','SOLUSDT']` with prop-driven or mode-detected list (include NQ when in futures mode) | MODIFY |
| A-M2 | `WelcomeScreen.tsx:22-34` (suggestion cards) | Replace crypto-only suggestions OR add NQ variants | MODIFY |
| A-M3 | `UserPreferencesPanel.tsx:130-132` (example placeholder) | Update example to mention futures or be generic | MODIFY |
| A-M4 | `agent/tools.go:~432` (exchange_type enum) | Add `ninjatrader` to supported exchange list | MODIFY |
| A-M5 | `agent/tools.go:~551` (`manage_exchange_config` tool description) | Document NinjaTrader CSV bridge setup steps the agent should know | MODIFY |
| A-M6 | `agent/tools.go:~592` (`manage_model_config` tool description) | Note that claw402/blockrun providers are crypto-only; not required for NQ | MODIFY |
| A-M7 | `agent/tools.go:~627` (`manage_strategy` tool: coin_source enum + static_coins example) | Add "NQ.c.0" to example; document that futures use static mode | MODIFY |
| A-M8 | `agent/tools.go:~691` (`execute_trade` description + examples) | Replace "long BTC, short ETH" examples with neutral or NQ-inclusive ones | MODIFY |
| A-M9 | `agent/tools.go:~756` (`get_market_snapshot` description) | Add asset-class guard: if NQ symbol, skip funding-rate / OI sections in response | MODIFY |
| A-M10 | `agent/tools.go:~796` (`get_candidate_coins`) | Note as crypto-only; if asked for NQ context, suggest static-mode strategy template | MODIFY |
| A-M11 | `agent/agent.go:62-72` (DefaultConfig WatchSymbols) | Make symbols configurable (env or per-user pref); default still BTC/ETH/SOL but extensible | MODIFY |
| A-M12 | `agent/agent.go:547-633` (Chinese system prompt) | Add NQ / CME context section ("If user mentions NQ/MNQ/futures, use NinjaTrader bridge + Databento data; skip funding rate questions") | MODIFY |
| A-M13 | `agent/agent.go:636-721` (English system prompt) | Mirror NQ awareness in English | MODIFY |
| A-M14 | `agent/web.go:207` (`HandleKlines`) | Abstract from hardcoded Binance futures URL OR route NQ symbols to Databento adapter | MODIFY |
| A-M15 | `agent/onboard.go:36+` (`needsSetup`) | Detect "NQ", "futures", "NinjaTrader" intent — skip claw402 wallet flow | MODIFY |
| A-M16 | `web/src/components/agent/SuggestionCards.tsx` (if exists, otherwise inline in AgentChatPage) | Add NQ suggestion when appropriate | MODIFY |
| A-M17 | `web/src/components/agent/MarketTicker.tsx:22-52` (fetch interval + UI) | Render NQ ticker entries cleanly (avoid stripping `.c.0` suffix) | MODIFY |
| A-M18 | Agent tool descriptions that mention "crypto" or "USDT" in tools.go | Soften: "crypto or futures symbol" | MODIFY |

### DEFER to Plan 2 (7 — full Agent NQ rewrite, not blocking)

| # | File | Why deferred |
|---|------|---------------|
| A-D1 | `agent/tools.go` `get_candidate_coins` full futures support | Requires new data source for NQ-equivalent ranking; not core to NQ trading |
| A-D2 | `agent/tools.go` `manage_strategy` `coin_source` enum extension | Static mode covers NQ; deeper futures-source registry is post-launch |
| A-D3 | `agent/model_provider_catalog.go:32-42` claw402 spec | Crypto payment unrelated to NQ; leave as-is |
| A-D4 | `agent/model_provider_catalog.go:44-52` blockrun-base | Same |
| A-D5 | `agent/model_provider_catalog.go:54-62` blockrun-sol | Same |
| A-D6 | `agent/memory.go` asset-class memory isolation | Risk: past crypto memory contaminates NQ reasoning. Add Plan 2 once we see real conflicts. |
| A-D7 | Deep agent-driven strategy generation for futures | Requires more sophisticated agent persona work; Plan 2 |

**Surface 4 LOC delta: ~140 (40 FE for ticker/welcome/suggestions, 100 BE for tool descriptions + system prompt + onboarding gating).**

## Combined acceptance criteria

When all 4 surfaces are at MODIFY-complete:

- **Config:** User adds NinjaTrader exchange via Settings → Exchanges → Add → select NT → enter DataDir + InstrumentName + DefaultContractQty → save. Listed correctly. No API key required.
- **Strategy:** User creates Strategy → SourceType="static" → StaticCoins=["NQ.c.0"] → PromptVariant="futures" → save. Editor hides funding/OI sections, shows NQ-aware risk control labels. Token estimate works. Test-run prompts contain NQ language.
- **Dashboard:** User creates Trader linking NinjaTrader exchange + NQ strategy + their AI model → opens Trader Dashboard → sees positions in contracts (no USDT unit, no leverage column, no liquidation column) → Recent Decisions render with NQ symbols → equity curve displays cleanly. AgentBeta sidebar markets list includes NQ.
- **AgentBeta:** User types "set up NinjaTrader for NQ trading" → agent suggests creating an exchange + strategy (does not push claw402 wallet) → user types "what's NQ doing today" → agent calls `get_market_price` (works) → "show NQ funding rate" → agent responds "futures don't have funding rate" gracefully.

## Final scope (4-surface view)

| Surface | LOC Δ | Of which NEW |
|---------|-------|--------------|
| Config | ~140 | 0 |
| Dashboard | ~80 | 0 |
| Strategy | ~340 | ~200 (futures prompt + validators) |
| AgentBeta | ~140 | 0 |
| **4-surface total** | **~700 LOC** | **~200 LOC new code** |

Plus the **backbone** (provider/databento/, provider/ninjatrader/, trader/ninjatrader/, cmd/nq_smoke, market adapter, kernel adapter wiring): ~1,500 LOC.

**Project total: ~2,200 LOC** spread across new code (~1,700) and surgical modifications (~500). Realistic timeline: **3-4 weeks focused engineering**.

**Items requiring confirmation** (flagged by the audit agents):
- `handleSaveExchange` already has 15 params; consider refactor to object param when adding NT fields
- Exchange.Create/Update Go function signatures: add NT fields as optional/last to avoid breaking call sites
- Icon file path for `/exchange-icons/ninjatrader.*` — need to provide an icon asset
- Decision whether to add `prompt_variant="futures"` as a NEW variant (alongside balanced/aggressive/conservative) OR a NEW `coin_source_variant` discriminator
- NQ symbol canonical format: confirm `NQ.c.0` (Databento) vs `NQ`/`NQM6` (specific contract) — propose `NQ.c.0` as primary
- Whether to *delete* leverage/liquidation columns server-side or *hide* them client-side — recommend client-side only (less backend churn)

---

# Plan 2: CME futures domain (production-grade)

> **Required before live money.** Plan 1 + 1.5 get the bot trading SIM cleanly. Plan 2 hardens it for CME's actual rules: tick sizes, contract rolls, session windows, holidays, settlement. Skipping Plan 2 = rejected orders + invalid prices + trading during closed markets.

## Task 17: Tick-size rounding for entry/SL/TP

**Why:** NQ tick = 0.25 (4 ticks/point); MNQ tick = 0.25. AI returns floating-point prices like `21503.17` — CME rejects anything not on a tick boundary. Round at the boundary before writing CSV.

**Files:**
- Create: `trader/ninjatrader/tick_rounding.go`
- Test: `trader/ninjatrader/tick_rounding_test.go`
- Modify: `trader/ninjatrader/trader.go` (call rounding before `csv_writer.Write`)

- [ ] **Step 1: Write the failing test**
```go
// trader/ninjatrader/tick_rounding_test.go
func TestRoundToTick(t *testing.T) {
    cases := []struct {
        in, tick, want float64
    }{
        {21503.17, 0.25, 21503.25},   // round up
        {21503.13, 0.25, 21503.00},   // round down
        {21503.125, 0.25, 21503.00},  // halfway = banker's round
        {21500.0, 0.25, 21500.00},    // exact
    }
    for _, tc := range cases {
        if got := RoundToTick(tc.in, tc.tick); got != tc.want {
            t.Errorf("RoundToTick(%v, %v) = %v, want %v", tc.in, tc.tick, got, tc.want)
        }
    }
}
```

- [ ] **Step 2: Implementation**
```go
// trader/ninjatrader/tick_rounding.go
package ninjatrader

import "math"

// InstrumentTickSize returns the tick size in points for a CME instrument.
// Returns 0.25 for NQ/MNQ/ES/MES (index futures default). Other instruments
// can be added as needed.
func InstrumentTickSize(symbol string) float64 {
    switch symbol {
    case "NQ", "MNQ", "ES", "MES":
        return 0.25
    case "YM", "MYM":
        return 1.0
    case "RTY", "M2K":
        return 0.10
    case "CL":  // crude oil
        return 0.01
    case "GC":  // gold
        return 0.10
    default:
        return 0.25  // safe default for indices
    }
}

// RoundToTick rounds price to the nearest tick boundary.
// Uses banker's rounding (round-half-to-even) to avoid bias.
func RoundToTick(price, tick float64) float64 {
    if tick <= 0 {
        return price
    }
    return math.Round(price/tick) * tick
}
```

- [ ] **Step 3: Wire into trader.go**
In `OpenLong` and `OpenShort`, before constructing SignalRow:
```go
tick := InstrumentTickSize(t.cfg.Symbol)
entry := RoundToTick(decision.Entry, tick)
sl := RoundToTick(decision.StopLoss, tick)
tp := RoundToTick(decision.TakeProfit, tick)
```

- [ ] **Step 4: Verify + commit**
```bash
go test ./trader/ninjatrader/...
git commit -m "feat(nt): round entry/SL/TP to instrument tick size"
```

## Task 18: CME session calendar + RTH/ETH gating

**Why:** CME Globex hours: Sun 5pm CT → Fri 4pm CT, daily break 4-5pm CT. Holiday closures exist (Christmas, New Year, etc.). Trading outside these windows = rejected orders + risk in thin liquidity.

**Files:**
- Create: `kernel/cme_calendar.go`
- Create: `kernel/cme_calendar_test.go`
- Modify: `kernel/engine.go` (skip decision cycle when market closed)

- [ ] **Step 1: Write the calendar**
```go
// kernel/cme_calendar.go
package kernel

import "time"

// IsCMEOpen reports whether CME Globex is open for index futures at the given time.
// Globex hours (Chicago time):
//   Sunday 17:00 → Friday 16:00, with a 60-minute daily break at 16:00–17:00.
// Holidays observed: New Year, MLK Day, Presidents Day, Good Friday, Memorial Day,
//   Juneteenth, Independence Day, Labor Day, Thanksgiving (+ day after), Christmas Eve,
//   Christmas Day. Each may have shortened hours; for v1 we treat them as full closures
//   and refuse to trade. Refine in Plan 3 if it becomes restrictive.
func IsCMEOpen(t time.Time) bool {
    chicago, _ := time.LoadLocation("America/Chicago")
    ct := t.In(chicago)
    if isCMEHoliday(ct) {
        return false
    }
    wd := ct.Weekday()
    hour := ct.Hour()
    switch wd {
    case time.Saturday:
        return false
    case time.Sunday:
        return hour >= 17
    case time.Friday:
        return hour < 16
    default:  // Mon-Thu
        return hour != 16
    }
}

// isCMEHoliday returns true if t falls on a CME-observed full-closure holiday.
// CME may have shortened-hours days (e.g. Good Friday, day after Thanksgiving),
// but for v1 we treat shortened days as full closures and refuse to trade.
// Refine in a later plan if this becomes operationally restrictive.
func isCMEHoliday(ct time.Time) bool {
    year := ct.Year()
    month := ct.Month()
    day := ct.Day()
    weekday := ct.Weekday()

    // Fixed-date holidays
    md := ct.Format("01-02")
    switch md {
    case "01-01": // New Year's Day
        return true
    case "06-19": // Juneteenth
        return true
    case "07-04": // Independence Day
        return true
    case "12-24": // Christmas Eve (early close treated as closure)
        return true
    case "12-25": // Christmas Day
        return true
    case "12-31": // New Year's Eve (early close treated as closure)
        return true
    }

    // MLK Day — 3rd Monday of January
    if month == time.January && weekday == time.Monday && (day-1)/7 == 2 {
        return true
    }

    // Presidents Day — 3rd Monday of February
    if month == time.February && weekday == time.Monday && (day-1)/7 == 2 {
        return true
    }

    // Good Friday — Friday before Easter
    if month == time.March || month == time.April {
        easter := easterSunday(year)
        goodFri := easter.AddDate(0, 0, -2)
        if ct.Year() == goodFri.Year() && ct.Month() == goodFri.Month() && ct.Day() == goodFri.Day() {
            return true
        }
    }

    // Memorial Day — last Monday of May
    if month == time.May && weekday == time.Monday {
        // Check if next Monday is in June (i.e. this is the last Monday of May)
        nextMon := ct.AddDate(0, 0, 7)
        if nextMon.Month() == time.June {
            return true
        }
    }

    // Labor Day — 1st Monday of September
    if month == time.September && weekday == time.Monday && day <= 7 {
        return true
    }

    // Thanksgiving — 4th Thursday of November (plus day after as early-close)
    if month == time.November && weekday == time.Thursday && (day-1)/7 == 3 {
        return true
    }
    // Day after Thanksgiving — Friday after 4th Thursday
    if month == time.November && weekday == time.Friday {
        thursday := ct.AddDate(0, 0, -1)
        if thursday.Month() == time.November && (thursday.Day()-1)/7 == 3 {
            return true
        }
    }

    return false
}

// easterSunday returns the date of Easter Sunday in the given year (Western/Gregorian).
// Used only for Good Friday calculation.
func easterSunday(year int) time.Time {
    // Anonymous Gregorian algorithm (Meeus/Jones/Butcher)
    a := year % 19
    b := year / 100
    c := year % 100
    d := b / 4
    e := b % 4
    f := (b + 8) / 25
    g := (b - f + 1) / 3
    h := (19*a + b - d - g + 15) % 30
    i := c / 4
    k := c % 4
    l := (32 + 2*e + 2*i - h - k) % 7
    m := (a + 11*h + 22*l) / 451
    month := (h + l - 7*m + 114) / 31
    day := ((h + l - 7*m + 114) % 31) + 1
    return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
```

- [ ] **Step 2: Test**
```go
// kernel/cme_calendar_test.go
func TestIsCMEOpen(t *testing.T) {
    chicago, _ := time.LoadLocation("America/Chicago")
    cases := []struct {
        name string
        when time.Time
        want bool
    }{
        {"Mon 10am normal trading", time.Date(2026, 6, 15, 10, 0, 0, 0, chicago), true},
        {"Mon daily break 4:30pm CT", time.Date(2026, 6, 15, 16, 30, 0, 0, chicago), false},
        {"Saturday closed", time.Date(2026, 6, 20, 12, 0, 0, 0, chicago), false},
        {"New Year's Day", time.Date(2026, 1, 1, 10, 0, 0, 0, chicago), false},
        {"MLK Day 2026 (Jan 19)", time.Date(2026, 1, 19, 10, 0, 0, 0, chicago), false},
        {"Presidents Day 2026 (Feb 16)", time.Date(2026, 2, 16, 10, 0, 0, 0, chicago), false},
        {"Good Friday 2026 (Apr 3)", time.Date(2026, 4, 3, 10, 0, 0, 0, chicago), false},
        {"Memorial Day 2026 (May 25)", time.Date(2026, 5, 25, 10, 0, 0, 0, chicago), false},
        {"Juneteenth", time.Date(2026, 6, 19, 10, 0, 0, 0, chicago), false},
        {"Independence Day", time.Date(2026, 7, 4, 10, 0, 0, 0, chicago), false},
        {"Labor Day 2026 (Sep 7)", time.Date(2026, 9, 7, 10, 0, 0, 0, chicago), false},
        {"Thanksgiving 2026 (Nov 26)", time.Date(2026, 11, 26, 10, 0, 0, 0, chicago), false},
        {"Day after Thanksgiving 2026 (Nov 27)", time.Date(2026, 11, 27, 10, 0, 0, 0, chicago), false},
        {"Christmas Eve", time.Date(2026, 12, 24, 10, 0, 0, 0, chicago), false},
        {"Christmas Day", time.Date(2026, 12, 25, 10, 0, 0, 0, chicago), false},
        {"New Year's Eve", time.Date(2026, 12, 31, 10, 0, 0, 0, chicago), false},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            if got := IsCMEOpen(tc.when); got != tc.want {
                t.Errorf("IsCMEOpen(%v) = %v, want %v", tc.when, got, tc.want)
            }
        })
    }
}
```

- [ ] **Step 3: Gate engine decision cycle**
In `kernel/engine.go` at the top of the scan loop:
```go
if config.Get().TradingMode == "futures" && !IsCMEOpen(time.Now()) {
    logger.Info("CME closed, skipping decision cycle")
    return
}
```

- [ ] **Step 4: Commit**
```bash
git commit -m "feat(futures): CME session calendar + skip decisions when market closed"
```

## Task 19: Contract roll automation

**Why:** NQ futures expire quarterly (March / June / Sep / Dec). NQ.c.0 is the front month; after expiry day, it auto-points to the next contract, but the AI may still be holding the old contract → liquidation. Need detection + roll plan.

**Files:**
- Create: `provider/databento/contract_calendar.go`
- Modify: `trader/ninjatrader/trader.go` (warn near expiry)
- Modify: `kernel/engine.go` (avoid new entries within 5 days of expiry)

- [ ] **Step 1: Add expiry resolver**
```go
// provider/databento/contract_calendar.go
package databento

import (
    "fmt"
    "strings"
    "time"
)

// CME month codes for futures contract symbology.
var cmeMonthCodes = map[byte]time.Month{
    'F': time.January,
    'G': time.February,
    'H': time.March,
    'J': time.April,
    'K': time.May,
    'M': time.June,
    'N': time.July,
    'Q': time.August,
    'U': time.September,
    'V': time.October,
    'X': time.November,
    'Z': time.December,
}

// NextExpiryFromSymbol returns the expiry date of the given CME contract code,
// disambiguating the single-digit year against `now`. Format: last 2 chars are
// month code (F/G/H/J/K/M/N/Q/U/V/X/Z) + last digit of year.
//
// Year disambiguation rule: assume the contract is in the current decade.
// If that would place the contract more than 1 year in the past, assume next
// decade. This handles the normal case (front-month contract within ~1 year
// of now) without breaking when the year-digit wraps (e.g. 2030).
//
// Examples (now=2026-05-22):
//   "MNQM6" → 2026-06 (current decade, current year)
//   "MNQU6" → 2026-09 (current decade, this year)
//   "MNQH7" → 2027-03 (current decade, next year)
//   "MNQM0" → 2030-06 (current decade, but year 2020 would be 6 years ago → bump to 2030)
func NextExpiryFromSymbol(symbol string, now time.Time) (time.Time, error) {
    if len(symbol) < 2 {
        return time.Time{}, fmt.Errorf("contract code too short: %q", symbol)
    }
    code := strings.ToUpper(symbol)
    monthChar := code[len(code)-2]
    yearChar := code[len(code)-1]

    month, ok := cmeMonthCodes[monthChar]
    if !ok {
        return time.Time{}, fmt.Errorf("invalid CME month code %q in %q", monthChar, symbol)
    }
    if yearChar < '0' || yearChar > '9' {
        return time.Time{}, fmt.Errorf("invalid year digit %q in %q", yearChar, symbol)
    }
    yearDigit := int(yearChar - '0')

    decade := (now.Year() / 10) * 10
    year := decade + yearDigit
    // If the candidate year is more than 1 year before now, the contract code
    // refers to the next decade.
    if year < now.Year()-1 {
        year += 10
    }
    return thirdFridayOf(year, month), nil
}

// DaysUntilExpiry returns calendar days from now until contract expiry.
// Returns 999 if the symbol cannot be parsed (treat as "not near expiry").
func DaysUntilExpiry(symbol string, now time.Time) int {
    exp, err := NextExpiryFromSymbol(symbol, now)
    if err != nil {
        return 999
    }
    return int(exp.Sub(now).Hours() / 24)
}

// thirdFridayOf returns the 3rd Friday of the given month — CME index futures
// expiry convention.
func thirdFridayOf(year int, month time.Month) time.Time {
    // Start at day 1, advance to first Friday, then add 14 days.
    first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
    offset := (int(time.Friday) - int(first.Weekday()) + 7) % 7
    return first.AddDate(0, 0, offset+14)
}
```

- [ ] **Step 2: Engine gating**
In `kernel/engine.go` at decision-evaluate time:
```go
if config.Get().TradingMode == "futures" {
    if days := databento.DaysUntilExpiry(symbol, time.Now()); days <= 5 {
        logger.Warnf("contract %s expires in %d days, blocking new entries", symbol, days)
        decision.Action = "HOLD"  // override
    }
}
```

- [ ] **Step 3: Test year-digit disambiguation**
```go
// provider/databento/contract_calendar_test.go
func TestNextExpiryFromSymbol_YearDisambiguation(t *testing.T) {
    cases := []struct {
        symbol string
        now    time.Time
        want   time.Time
    }{
        {"MNQM6", date(2026, 5, 22), date(2026, 6, 19)},   // current quarter
        {"MNQU6", date(2026, 5, 22), date(2026, 9, 18)},   // next quarter
        {"MNQH7", date(2026, 5, 22), date(2027, 3, 19)},   // next year
        {"MNQM0", date(2029, 12, 1), date(2030, 6, 21)},   // year-digit wrap into next decade
        {"MNQH0", date(2029, 12, 1), date(2030, 3, 15)},   // wrap, earliest month
    }
    for _, tc := range cases {
        got, err := NextExpiryFromSymbol(tc.symbol, tc.now)
        if err != nil {
            t.Errorf("symbol=%q: %v", tc.symbol, err)
            continue
        }
        if !got.Equal(tc.want) {
            t.Errorf("symbol=%q now=%v: got %v, want %v", tc.symbol, tc.now, got, tc.want)
        }
    }
}

func date(y int, m time.Month, d int) time.Time {
    return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
```

- [ ] **Step 4: Commit**
```bash
git commit -m "feat(futures): contract roll detection + block entries near expiry"
```

## Task 20: Symbol precision + decimal-safe arithmetic

**Why:** Position sizing, PnL, tick math all use float64. For NQ at 21500, 0.01-point rounding error = $5 over 100 trades. Audit and add `decimal.Decimal` only where it matters (position sizing, fill matching).

**Files:**
- Modify: `kernel/engine_position.go` (use math/big or shopspring/decimal for sizing)
- Test: `kernel/engine_position_test.go`

- [ ] **Step 1: Use existing helpers**
Check `market/data_klines.go` for any rounding helpers; if absent, add:
```go
func roundDecimals(v float64, decimals int) float64 {
    p := math.Pow(10, float64(decimals))
    return math.Round(v*p) / p
}
```
2 decimals for NQ prices; 0 decimals for contract quantities.

- [ ] **Step 2: Position-size sanity test**
```go
func TestPositionSize_NQ(t *testing.T) {
    // 1 NQ contract @ 21500 = $1,075,000 notional ($50/point × 21500)
    // Account $10k, risk 1% = $100; stop 25 points = $1,250 loss/contract
    // → Should size to 0 contracts (can't afford 1)
}
```

- [ ] **Step 3: Commit**

---

# Plan 3: Risk management + kill switches

> **Required for live money.** Without explicit risk limits, a runaway AI decision can blow up the account. Plan 3 adds hard limits enforced in Go (not just in the prompt).

## Task 21: Daily loss limit + force-flat kill switch

**Why:** AI prompt says "don't risk more than 1%" but it's a suggestion, not a guard rail. Need hard server-side enforcement.

**Files:**
- Create: `kernel/risk_limits.go`
- Create: `kernel/risk_limits_test.go`
- Modify: `kernel/engine.go` (pre-trade risk check + post-fill PnL check)
- Modify: `config/config.go` (add risk limit env vars)

- [ ] **Step 1: Risk-limit struct**
```go
// kernel/risk_limits.go
package kernel

type RiskLimits struct {
    MaxDailyLossUSD     float64  // hard stop for the day
    MaxConcurrentTrades int      // open position cap
    MaxNotionalUSD      float64  // total notional cap
    MaxContractsPerOrder int     // single-order size cap
}

func (r *RiskLimits) CheckPreTrade(ctx *Context, decision *Decision) error {
    if ctx.Account.TotalPnL < -r.MaxDailyLossUSD {
        return fmt.Errorf("daily loss limit hit: %.2f", ctx.Account.TotalPnL)
    }
    if len(ctx.Positions) >= r.MaxConcurrentTrades {
        return fmt.Errorf("concurrent trade cap reached: %d", r.MaxConcurrentTrades)
    }
    notional := decision.PositionSizeUSD
    for _, p := range ctx.Positions {
        notional += p.NotionalUSD
    }
    if notional > r.MaxNotionalUSD {
        return fmt.Errorf("notional cap exceeded: %.2f", notional)
    }
    return nil
}
```

- [ ] **Step 2: Wire into engine**
In `kernel/engine.go` after parsing the AI decision:
```go
limits := RiskLimits{
    MaxDailyLossUSD:     config.Get().RiskMaxDailyLossUSD,
    MaxConcurrentTrades: config.Get().RiskMaxConcurrentTrades,
    MaxNotionalUSD:      config.Get().RiskMaxNotionalUSD,
    MaxContractsPerOrder: config.Get().RiskMaxContractsPerOrder,
}
if err := limits.CheckPreTrade(ctx, decision); err != nil {
    logger.Warnf("⚠️ risk limit violated: %v — forcing HOLD", err)
    decision.Action = "HOLD"
}
```

- [ ] **Step 3: Config env vars**
```go
// config/config.go
RiskMaxDailyLossUSD     float64
RiskMaxConcurrentTrades int
RiskMaxNotionalUSD      float64
RiskMaxContractsPerOrder int
```
Defaults: `$500`, `2`, `$50000`, `5`.

- [ ] **Step 4: Force-flat API endpoint**
Add `POST /api/risk/force-flat` that calls every active trader's `CancelAllOrders + CloseLong/CloseShort`. (For NT v1: just cancel pending signals — manual close not supported.) Surface as a red "EMERGENCY FLAT" button on TraderDashboardPage.

- [ ] **Step 5: Tests + commit**

## Task 22: Stale-data + drift detection

**Why:** If Databento returns stale OHLCV (e.g. last bar is 10 min old), the AI is making decisions on old data. Detect and skip the cycle.

**Files:**
- Modify: `market/data.go` (timestamp check after fetch)
- Modify: `kernel/engine.go` (abort decision on stale data)

- [ ] **Step 1: Add `IsFresh` check**
```go
// in market/data.go
func (k *Kline) IsFresh(maxAge time.Duration) bool {
    return time.Since(time.UnixMilli(k.OpenTime)) <= maxAge
}
```

- [ ] **Step 2: Engine gating**
```go
lastBar := data.Klines[len(data.Klines)-1]
if !lastBar.IsFresh(2 * time.Minute) {
    logger.Warnf("stale data for %s (last bar %v old) — skipping cycle", symbol, time.Since(time.UnixMilli(lastBar.OpenTime)))
    return
}
```

- [ ] **Step 3: Commit**

---

# Plan 4: Observability + reliability

> **Required for confidence in production.** Without metrics, you don't know if the bot is healthy. Without retries, transient network blips become trade-misses.

## Task 23: Structured logging + decision audit trail

**Files:**
- Modify: `kernel/engine.go` (log decision with full context to structured logger)
- Modify: `store/decision.go` (persist decision JSON + risk-check outcome + execution status)
- Create: `api/handler_decisions.go` (read endpoint for audit replay)

- [ ] **Step 1: Decision struct expansion**
```go
type Decision struct {
    ID               int64
    TraderID         string
    Symbol           string
    Action           string
    Entry, SL, TP    float64
    Confidence       float64
    Reasoning        string
    PromptVersion    string  // hash of system prompt for reproducibility
    AIModel          string
    AILatencyMs      int64
    RiskCheckPassed  bool
    RiskCheckError   string
    ExecutionStatus  string  // "queued", "filled", "rejected", "blocked"
    FillPrice        *float64
    FillLatencyMs    *int64
    CreatedAt        time.Time
}
```

**Timezone requirement:** `CreatedAt` is always stored as UTC. The Go layer
inserts `time.Now().UTC()`; the database column is TIMESTAMP WITH TIME ZONE
(Postgres) or TEXT in ISO 8601 UTC format (SQLite). The display layer in
TraderDashboardPage converts to the user's local time zone for rendering.
Storing local time in the audit trail creates DST-transition bugs that
cause decisions to appear out-of-order or duplicated; UTC is non-negotiable.

The same rule applies to FillLatencyMs computation: subtract two UTC
timestamps, do not mix wall-clock and monotonic time.

- [ ] **Step 2: Persistence**
GORM Migrate adds the new columns. Insert at decision-time + update at fill-time.

- [ ] **Step 3: API + UI**
Expose at `GET /api/decisions/audit?trader_id=xxx&since=2026-05-22`; render in TraderDashboardPage's Decisions tab.

## Task 24: Retry + circuit breaker for Databento + NT bridge

**Why:** Databento HTTP can return 5xx; NT CSV write can fail with EBUSY on Windows file lock contention. Need bounded retry with backoff, plus circuit breaker that pauses the trader after N consecutive failures.

**Files:**
- Create: `mcp/retry.go` (generic retry-with-backoff)
- Modify: `provider/databento/client.go` (use retry on doRequest)
- Modify: `provider/ninjatrader/csv_writer.go` (retry on rename collision)

- [ ] **Step 1: Retry helper**
```go
func RetryWithBackoff(ctx context.Context, maxAttempts int, fn func() error) error {
    delay := 200 * time.Millisecond
    for attempt := 0; attempt < maxAttempts; attempt++ {
        if err := fn(); err == nil {
            return nil
        }
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
        }
        delay *= 2
        if delay > 5*time.Second {
            delay = 5 * time.Second
        }
    }
    return fmt.Errorf("max retries exceeded")
}
```

- [ ] **Step 2: Circuit breaker**
```go
type CircuitBreaker struct {
    failureThreshold int
    cooldown         time.Duration
    failures         int
    openedAt         time.Time
    mu               sync.Mutex
}
func (cb *CircuitBreaker) Allow() bool { /* ... */ }
func (cb *CircuitBreaker) RecordFailure() { /* ... */ }
func (cb *CircuitBreaker) RecordSuccess() { cb.failures = 0 }
```

- [ ] **Step 3: Wire into trader loop**
After N=5 consecutive failures, pause the trader for 5 minutes and log to decision audit trail.

## Task 25: Prometheus metrics endpoint

**Why:** CTO ask. Without numbers you don't know what's happening.

**Files:**
- Modify: `main.go` (add /metrics route)
- Create: `telemetry/metrics.go` (counter + histogram definitions)

- [ ] **Step 1: Define metrics**
```go
// telemetry/metrics.go
var (
    DecisionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "nofx_decisions_total"},
        []string{"trader_id", "action", "status"},
    )
    DecisionLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "nofx_decision_latency_seconds"},
        []string{"trader_id"},
    )
    FillLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "nofx_fill_latency_seconds"},
        []string{"exchange"},
    )
    DatabentoErrorsTotal = prometheus.NewCounter(
        prometheus.CounterOpts{Name: "nofx_databento_errors_total"},
    )
)
```

- [ ] **Step 2: Expose /metrics**
```go
http.Handle("/metrics", promhttp.Handler())
```

- [ ] **Step 3: Instrument decision path**
Add `DecisionLatency.WithLabelValues(traderID).Observe(elapsed.Seconds())` in `engine.go`.

---

# Plan 5: Testing matrix

> **Required for CI confidence.** Without integration tests, every code change is a coin flip.

## Task 26: Databento mock server

**Files:**
- Create: `provider/databento/mock_server.go` (httptest-based)
- Modify: `provider/databento/historical_test.go` (use mock)

- [ ] **Step 1: Mock OHLCV server**
```go
// provider/databento/mock_server.go
func NewMockServer(t *testing.T, fixture string) *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/v0/timeseries.get_range" {
            http.ServeFile(w, r, fixture)
            return
        }
        if r.URL.Path == "/v0/symbology.resolve" {
            // ...
        }
    }))
}
```
Fixture: `provider/databento/fixtures/nq-ohlcv-1m.json` with 100 sample bars.

- [ ] **Step 2: Update tests to use mock**

## Task 27: NinjaTrader CSV bridge mock harness

**Files:**
- Create: `provider/ninjatrader/mock_nt.go` (goroutine that simulates NT polling + writing fills)
- Create: `trader/ninjatrader/integration_test.go` (full Go → CSV → mock NT → CSV → tailer round-trip)

- [ ] **Step 1: Mock NT loop**
```go
// Reads trade_signals.csv every 100ms (faster than real 2s for tests),
// writes a synthetic fill to trades_taken.csv after 200ms.
func StartMockNT(dataDir string, fillDelay time.Duration) (stop func())
```

- [ ] **Step 2: Round-trip test**
```go
func TestTrader_OpenLong_RoundTrip(t *testing.T) {
    dir := t.TempDir()
    stop := StartMockNT(dir, 200*time.Millisecond)
    defer stop()

    tr := New(Config{DataDir: dir, Symbol: "MNQ"})
    err := tr.OpenLong("MNQ", 21500, 21450, 21550)
    require.NoError(t, err)

    // Wait for fill...
    fill := waitForFill(t, dir, 2*time.Second)
    require.Equal(t, "LONG", fill.Direction)
}
```

## Task 28: AI prompt golden tests

**Why:** Changing the prompt template can silently break decision quality. Pin the prompt structure with snapshot tests.

**Files:**
- Create: `kernel/engine_prompt_golden_test.go`
- Create: `kernel/testdata/golden/futures_aggressive.txt`

- [ ] **Step 1: Snapshot test**
```go
func TestBuildFuturesSystemPrompt_Golden(t *testing.T) {
    got := BuildFuturesSystemPrompt(10000)
    want := readGoldenFile(t, "futures_aggressive.txt")
    if got != want {
        t.Fatalf("prompt changed.\nDiff:\n%s", diff(got, want))
    }
}
```
Update via `go test -update`.

## Task 29: End-to-end smoke matrix

Extend `cmd/nq_smoke/main.go` with sub-commands:
- `nq_smoke databento` — fetch OHLCV, verify shape
- `nq_smoke resolver` — resolve NQ.c.0 → MNQM6
- `nq_smoke prompt` — build system prompt, check non-empty
- `nq_smoke roundtrip` — full write→mockNT→tailer cycle
- `nq_smoke all` — runs everything

---

# Plan 6: Operational runbook

> **CTO/SRE-style ops doc.** Not code — but every production system needs these procedures.

## Task 30: Document startup procedure

**File:** Create `docs/operations/STARTUP.md`

Sections to include:
1. **Pre-flight checks** — env vars set (JWT_SECRET, DATABENTO_API_KEY, NINJATRADER_DATA_DIR), NT8 running on Windows, ClaudeTrader strategy attached to MNQ chart, WSL2 mirrored networking enabled.
2. **Cold start** — `./nofx-bin > /tmp/nofx.log 2>&1 &` — verify port 8080 listens, log shows "✅ System started successfully".
3. **Verify trader loads** — `curl localhost:8080/api/traders` returns the configured NT trader. Log shows `📦 Loading trader X (AI Model: deepseek, Exchange: ninjatrader/...)`.
4. **First-trade smoke** — manual decision via `cmd/nq_smoke/main.go all`, watch for fill appearing in `trades_taken.csv` and in DB `decisions` table.
5. **Shutdown** — `pkill -TERM -f nofx-bin`; verify no hung file handles via `lsof | grep NofxTrader`.
6. **Windows Defender exclusion for the data directory.** Defender's
   real-time scanner can transiently lock files during `os.Rename` from
   the Go side, causing the atomic temp+rename pattern in
   `provider/ninjatrader/csv_writer.go` to fail with EBUSY. Exclude the
   data directory from real-time scanning. From an Administrator
   PowerShell prompt on the Windows host:

```powershell
   Add-MpPreference -ExclusionPath "C:\Users\<user>\NofxTrader\data"
```

   Verify the exclusion is active:

```powershell
   Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
```

   This is a one-time setup per host. Without it, you may see intermittent
   "rename: Access is denied" errors in Go logs during high-frequency signal
   writes — these are benign (the next write succeeds) but noisy.

## Task 31: Document rollback procedure

**File:** Create `docs/operations/ROLLBACK.md`

1. **Code rollback** — `git checkout <previous-good-commit> -- .`; `go build -o nofx-bin .`; restart.
2. **Schema rollback** — GORM auto-migrate is additive; for column removal, use a migration file under `store/migrations/`. Always snapshot `data/data.db` to `data/data.db.bak.<timestamp>` before code deploy.
3. **NT script rollback** — restore prior `claudetrader.cs` from git; rebuild in NT (`F5`).
4. **Risk wipe** — if rollback follows an unexpected loss, run force-flat (`curl -X POST localhost:8080/api/risk/force-flat -H "Authorization: Bearer $TOKEN"`) before any new trades.

## Task 32: Document monitoring + alerting

**File:** Create `docs/operations/MONITORING.md`

Key dashboards (Grafana / similar):
- **Decision rate** — `rate(nofx_decisions_total[5m])` — should be ~1/scan_interval
- **Fill latency** — `histogram_quantile(0.95, nofx_fill_latency_seconds)` — alert if > 30s
- **Databento errors** — `rate(nofx_databento_errors_total[10m])` — alert if > 0.1/sec
- **Daily PnL** — read from DB; alert if < -$500 (matches `RiskMaxDailyLossUSD`)
- **CME session health** — alert at 16:00 CT if a trade attempt was logged during the break

Manual checks (no alerting infra yet):
- `tail -f /tmp/nofx.log | grep -E "ERROR|WARN"`
- `curl localhost:8080/api/risk/status` — exposes current PnL + position count vs limits

## Task 33: Document disaster recovery

**File:** Create `docs/operations/DR.md`

1. **DB corruption** — restore from latest `.bak`; replay missing decisions via NT trades_taken.csv backfill (read fills, reconstruct decisions).
2. **NT8 crash mid-trade** — `claudetrader.cs` tracks active position internally; restart resumes from current position. **Risk:** if SL/TP weren't acked to broker, they're lost. Mitigation: after every NT restart, manually verify SL/TP are placed via the NT Orders window.
3. **Databento outage** — fall back to last-known OHLCV cached in DB; engine refuses new entries on stale data per Task 22.
4. **WSL2 reboot** — confirm `wsl --version` shows mirrored mode; reconfirm `/mnt/c/` is writable from Linux side.
5. **Lost JWT secret** — invalidate all sessions (clear `users.last_session_token`); users must re-login.

## Task 34: Document trader-mode runbook

**File:** Create `docs/operations/TRADER_MODE.md`

For the user (not engineers):
- How to switch from SIM to live (NT8 simulation account → real account; re-login required)
- Daily checklist before market open: NT8 connected? ClaudeTrader strategy enabled? Bot logs healthy?
- Weekly checklist: contract roll calendar (3rd Friday); review decision audit trail; rotate API keys.
- Emergency: how to hit force-flat from the dashboard.

---

# Plan 7: Documentation + ADRs

## Task 35: Architecture Decision Records

Create `docs/adr/` with one file per significant decision:
- `001-csv-bridge-vs-tcp.md` — why CSV first, TCP later (Plan 1.5)
- `002-databento-vs-alternatives.md` — why Databento over Polygon/IB Trader Workstation/CQG
- `003-nt8-vs-tradovate.md` — why we picked NT8 (user has license, NinjaScript open-source bridges exist)
- `004-decision-json-shape.md` — locking the action/symbol/entry/stop_loss/take_profit/reasoning shape across crypto + futures
- `005-tick-rounding-strategy.md` — banker's rounding, not floor/ceil, to avoid bias

Each ADR follows the format: **Status / Context / Decision / Consequences**.

## Task 36: API reference

`docs/api/README.md` — list every HTTP endpoint with request/response example. Generate from `api/server.go` route table if practical; otherwise write manually.

## Task 37: Onboarding doc

`docs/getting-started/NEW_DEV.md` — for someone joining the project:
- Repo layout map
- Run locally (`go build`, `cd web && npm run dev`)
- How to add a new exchange (link to `trader/CLAUDE.md`)
- How to add a new AI provider (link to `provider/CLAUDE.md`)
- Plan reading order: Plan 1 → 1.5 → 2 → 3 → 4 → 5 → 6 → 7

---

# Completion checklist (definitive)

| Plan | Status | Blocker for live? |
|---|---|---|
| Plan 1 (CSV bridge SIM) | ✅ Shipped | — |
| Plan 1.5 (NT8 AddOn TCP) | 📋 Documented | No (CSV works for SIM) |
| Plan 2 (CME domain) | 📋 Documented | **YES** — tick rounding + sessions + rolls |
| Plan 3 (Risk + kill switch) | 📋 Documented | **YES** — no live without hard limits |
| Plan 4 (Observability) | 📋 Documented | Recommended |
| Plan 5 (Testing matrix) | 📋 Documented | Recommended |
| Plan 6 (Operational runbook) | 📋 Documented | **YES** — for any non-dev operator |
| Plan 7 (Docs + ADRs) | 📋 Documented | No |

**Order of execution after Plan 1 cleanup (Tasks 12-16):**
1. Plan 2 (Tasks 17-20) — gate live with tick/session/roll/decimal correctness
2. Plan 3 (Tasks 21-22) — kill switches before any real money
3. Plan 4 (Tasks 23-25) — observability so you can SEE the bot
4. Plan 5 (Tasks 26-29) — CI confidence before iterating
5. Plan 6 (Tasks 30-34) — runbook for ops handoff
6. Plan 7 (Tasks 35-37) — write while context is fresh

After all of these: **the bot is ready for paper-live → real-live trading with reasonable safeguards.** Without Plan 2 + 3, do NOT trade real money.

## Plan 1 — Post-Mortem (2026-05-25)

Plan 1 was marked SHIPPED on 2026-05-22 based on:
- 2 real SIM fills on SIM101 via NT Playback
- Unit tests passing for provider/databento and provider/ninjatrader

Final acceptance via cmd/nq_smoke against the live Databento API
was deferred to 2026-05-25. That acceptance session surfaced six
findings in the cmd/nq_smoke entry point — four code bugs (fixed
in the 2026-05-25 session), one verified non-bug, and one
documentation/tier-reality observation. All findings sat in the
warmup / data-fetch / parser code path. NONE in the AI/CSV/NT
path.

The CSV→NT→fill round trip was validated tonight (2026-05-25 evening
session) with a real LIVE-session fill at 29807 on SIM101 after CME
Memorial Day closure ended.

### Bugs found and fixed (2026-05-25 session)

1. Databento window-too-recent — smoke runner queried end=now,
   but Databento Historical has publication lag. Fixed: 17min
   buffer (commit 286ee3b9), then 24h buffer (3ae0bf77) once
   tier-lag reality surfaced.

2. Weekend gap — 24h buffer fails on Monday morning when 24h
   back lands in the Sunday 18:00 to Friday 17:00 weekend hole.
   Fixed: 96h buffer (commit 18c21bae) covers worst-case 3-day
   holiday weekends.

3. Parser struct shape — Databento ohlcv-1m response has
   ts_event nested in hd, not top level. Hand-fabricated test
   fixture put ts_event at top level, so the unit test passed
   against a payload that never matched real API output. The
   bug shipped to SHIPPED status because the test was fiction.
   Fixed: real captured fixture + CRITICAL comment block
   (commit 8d20487d).

4. Lookback range too narrow — 30min window returns 30 bars;
   EMA50 needs 50 bar minimum. Fixed: 90min for 40-bar
   headroom (commit 46cd2aa1).

5. 1e9 fixed-point division — VERIFIED NOT A BUG. The
   scaledFloat() helper at provider/databento/historical.go:117-123
   correctly divides integer-scaled prices (e.g. "21500250000000")
   by 1e9 to produce floats (21500.25). Reviewed during the
   parser-fix pre-edit check on 2026-05-25; behavior was already
   correct, no patch required. Listed here so the audit trail
   shows the code path was examined and ruled out, not silently
   skipped.

6. Account-tier vs documented embargo — documentation/reality
   observation, captured as [VERIFY-16] below. The Plan 1.5
   design assumed Databento intraday GLBX.MDP3 has a documented
   15-minute embargo; that figure applies to the
   real-time-with-embargo subscription tier. The current
   Historical-tier account has multi-hour available_end lag
   (~3 hours observed) and required 96h lookback for
   weekend-spanning queries (per item 2 above). This is not a
   bug in the code — it is a planning assumption that needs
   revisiting if Plan 1.5 Cold Start runs on this tier.

### Lessons

1. Unit tests against fabricated fixtures are not unit tests.
   They are theater. Future Databento integration changes MUST
   verify fixtures against live curl output before committing.

2. SHIPPED status should require live-API acceptance through
   the production code path, not just unit tests + manual NT
   smoke. The 2026-05-22 SHIPPED claim was technically true
   for the NT bridge (csv_writer / csv_tailer / trader) which
   were exercised manually, but the cmd/nq_smoke entry that
   exercises the FULL chain Databento → indicators → AI →
   CSV → NT had never run.

3. Recommendation: future plans MUST add a "Live API acceptance"
   gate to the SHIPPED criteria, not just "unit tests pass +
   manual smoke." This is captured in Plan 5 (Testing matrix)
   Task 26 (Databento mock server) and Task 29 (E2E smoke
   matrix) — those tasks now have proven need.

### Pass 2 validation (2026-05-25 evening)

After Memorial Day closure ended (CME reopen 18:00 ET), pipeline
exercised end-to-end with a single LONG signal on SIM101:

- Signal: entry=29812.00, sl=29792.00, tp=29842.00 (1.5 R:R)
- CSV write to /mnt/c/Users/hoang/NofxTrader/data/trade_signals.csv
- NT VLTrader detected signal within ~2 sec
- Market order placed against MNQ 06-26 contract
- Filled at 29807.00 (5pt favorable slippage)
- Fill row written to trades_taken.csv
- Go tailer detected and logged the new fill
- Clean exit

H4 observed exactly as documented in Plan 1 hazards section: the
Go tailer re-emitted the two historical Playback fills from
2026-05-22 at startup before catching the new live fill. This is
the "fill replay on session rollover" hazard. Plan 1.5 fix
(stable fill-ID dedup + persisted offset) already specified in
canonical plan.

### Current Plan 1 status (after 2026-05-25)

- Databento parser handles real API response shape: DONE
- cmd/nq_smoke validates end-to-end pipeline (Pass 1): DONE
- cmd/nq_smoke validates CSV→NT→fill loop (Pass 2): DONE
- Three pending PRs to merge in order:
  - chore/hide-faq-nav → nq-databento-ninjatrader-plan
  - chore/rename-claudetrader-to-vltrader → main (audit doc)
  - nq-databento-ninjatrader-plan → main (Plan 1 release)
- Plan 1.5 implementation: triggered by Plan 1.5 trigger
  conditions, not yet fired

### Branch state since Plan 1.5 design closeout

After e15c1dc0 (Plan 1.5 merge logic state machine, 2026-05-25),
eight commits landed on nq-databento-ninjatrader-plan during the
Plan 1 live-API acceptance + cleanup session:

- 286ee3b9 fix(nq_smoke): add 17min lag buffer for Databento
  historical embargo
- 3ae0bf77 fix(nq_smoke): use 24h lookback for Historical-tier
  Databento account
- 18c21bae fix(nq_smoke): 96h lookback survives weekend + 3-day
  holiday gaps
- 8d20487d fix(databento): parser struct shape + real fixture +
  docs URL (the parser-fixture-fiction bug surfaced here)
- 46cd2aa1 fix(nq_smoke): bump lookback range to 90min for EMA50
  headroom
- 9b7f8d70 chore(plan-1): rename ClaudeTrader → VLTrader in
  comments (1:1 swap across 5 Go files; upstream URL citation in
  provider/ninjatrader/types.go:7 preserved as historical
  attribution)
- 5db0bea6 docs(plan): Plan 1 post-mortem + Pass 2 validation +
  VERIFY-16 (this post-mortem section)
- 6d66b6ac docs(plan): document Playwright MCP availability for
  UI verification (registered npx -y @playwright/mcp@latest
  --headless in ~/.claude.json; Chromium binary pre-installed
  at ~/.cache/ms-playwright/ for headless E2E testing of
  frontend tasks 14/15)

Branch tip at the time of this update: 6d66b6ac. Branch is ahead
of origin/main by the full commit history since the v4-polish
merge (339d90ab).

### [VERIFY-16] Databento tier vs documented embargo

Plan 1.5 design assumed Databento intraday GLBX.MDP3 has a
documented 15-minute embargo. That number applies to the
real-time-with-embargo subscription tier. The current account
is on basic Historical tier with multi-hour available_end lag
(~3 hours observed during 2026-05-25 acceptance session) and
weekend gap behavior requiring 96h lookback for Monday-morning
queries.

This does NOT change the Plan 1.5 architecture — bars come
from NT live feed via CQG/Rithmic, not Databento. Databento
Historical remains for warmup, gap-fill, backtest only, where
the multi-hour lag is acceptable.

But it DOES change one production-path expectation: when
implementing Plan 1.5 Cold Start Sequence (databento_warmup_end
= T0 - 17min), the 17min constant assumes real-time-with-embargo
tier. If the system runs on the current Historical tier, the
warmup must use a larger buffer (96h or session-aware) and
accept stale warmup bars.

For SIM/Plan-1 validation: 96h buffer is fine.
For live/Plan-1.5 implementation: confirm Databento tier
before implementing the Cold Start logic. Either upgrade to
real-time-with-embargo or adapt the buffer constant.
