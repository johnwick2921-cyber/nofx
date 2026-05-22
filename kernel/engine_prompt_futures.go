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
//
// NOTE: market.ExportCalculateMACD returns only the MACD line (one float64).
// Signal/histogram require extending the indicator API and are deferred.
type FuturesContext struct {
	Symbol       string
	CurrentPrice float64
	// indicator snapshot
	EMA20     float64
	EMA50     float64
	RSI14     float64
	MACD      float64 // MACD line only
	ATR14     float64
	BollUpper float64
	BollLower float64
}

func BuildFuturesSystemPrompt(c FuturesPromptConfig) string {
	var b strings.Builder
	b.WriteString("# You are a professional index-futures trading AI specializing in CME E-mini Nasdaq-100 contracts.\n\n")
	b.WriteString(fmt.Sprintf("## Instrument\n- Symbol: %s\n- Tick size: %.2f points\n- Contract multiplier: $%.2f per point\n\n", c.Symbol, c.TickSize, c.ContractMultiplier))
	b.WriteString("## Hard constraints\n")
	b.WriteString("- Every entry MUST include a stop loss and a take profit, expressed as absolute prices.\n")
	b.WriteString(fmt.Sprintf("- Stop loss distance: minimum %.0f points, maximum %.0f points from entry.\n", c.MinStopPoints, c.MaxStopPoints))
	b.WriteString(fmt.Sprintf("- Minimum risk/reward: %.2f (reward must be at least %.2fx the risk).\n", c.MinRiskReward, c.MinRiskReward))
	b.WriteString("- One position at a time. Do NOT propose averaging in or pyramiding.\n")
	b.WriteString("- Prices must be in tick increments (multiples of " + fmt.Sprintf("%.2f", c.TickSize) + ").\n")
	b.WriteString("- The market session is CME futures hours; do not assume 24/7 trading.\n\n")
	b.WriteString("## Decision output\n")
	b.WriteString("Respond ONLY with JSON of the following exact shape:\n")
	b.WriteString("```json\n")
	b.WriteString(`{"action":"LONG"|"SHORT"|"NONE","entry":0.00,"stop_loss":0.00,"take_profit":0.00,"reasoning":"<one-paragraph explanation>"}`)
	b.WriteString("\n```\n")
	b.WriteString("\n- `action=NONE` is a valid and frequently correct answer. Do not force a trade.\n")
	b.WriteString("- All three price fields are absolute (e.g. 21500.25), not deltas from entry.\n\n")
	b.WriteString("## Trade plan checklist (apply before answering LONG/SHORT)\n")
	b.WriteString("1. Is there a clear directional bias from EMA20 vs EMA50 alignment?\n")
	b.WriteString("2. Does RSI confirm or contradict that bias? (extreme = caution)\n")
	b.WriteString("3. Is MACD positive (for LONG) or negative (for SHORT)?\n")
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
	b.WriteString(fmt.Sprintf("- EMA20: %.2f (current price %s)\n", ctx.EMA20, sidePos(ctx.CurrentPrice, ctx.EMA20)))
	b.WriteString(fmt.Sprintf("- EMA50: %.2f (current price %s)\n", ctx.EMA50, sidePos(ctx.CurrentPrice, ctx.EMA50)))
	b.WriteString(fmt.Sprintf("- EMA20 vs EMA50: %s\n", emaAlignment(ctx.EMA20, ctx.EMA50)))
	b.WriteString(fmt.Sprintf("- RSI14: %.1f (%s)\n", ctx.RSI14, rsiBucket(ctx.RSI14)))
	b.WriteString(fmt.Sprintf("- MACD: %.2f (line only; signal/histogram require extended indicator API)\n", ctx.MACD))
	b.WriteString(fmt.Sprintf("- ATR14: %.2f points\n", ctx.ATR14))
	b.WriteString(fmt.Sprintf("- Bollinger Bands: upper %.2f, lower %.2f, position: %s\n", ctx.BollUpper, ctx.BollLower, bollPosition(ctx.CurrentPrice, ctx.BollUpper, ctx.BollLower)))
	b.WriteString("\n## Decision\nGive me your trade decision in the JSON format specified by the system prompt.\n")
	return b.String()
}

func sidePos(price, ref float64) string {
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
