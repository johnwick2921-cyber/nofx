// Command nq_smoke is a standalone single-cycle runner for the NQ trading slice.
//
// Pipeline: Databento OHLCV → indicators → futures prompt → stdin-paste AI
// decision → CSV signal → NT fill → console output.
//
// Requires:
//   - NT 8 running on Windows host with claudetrader.cs on an MNQ chart in SIM
//   - .env populated with DATABENTO_API_KEY + NINJATRADER_DATA_DIR
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
	const (
		// SMOKE TEST CONFIG — uses 24h lookback to accommodate Databento
		// Historical tier WITHOUT real-time entitlement. Yesterday's data
		// is guaranteed available regardless of tier, session, or holiday.
		//
		// Production Plan 1.5 will use the documented 15-min embargo when
		// upgraded to real-time-with-embargo tier. See Plan 1.5 Cold Start
		// Sequence spec (commit e15c1dc0) for the canonical numbers.
		smokeLookbackEnd   = 24 * time.Hour
		smokeLookbackRange = 30 * time.Minute
	)
	end := time.Now().UTC().Add(-smokeLookbackEnd)
	start := end.Add(-smokeLookbackRange)
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
		MACD:         market.ExportCalculateMACD(klines),
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

	// 4. AI call — STUBBED. Paste a decision JSON via stdin.
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
