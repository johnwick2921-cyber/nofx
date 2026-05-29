package kernel

import (
	"fmt"
	"strings"
)

// BuildFuturesDecisionSystemPrompt builds the CME index-futures (MNQ) system
// prompt. Unlike the standalone BuildFuturesSystemPrompt above (which emits an
// incompatible LONG/SHORT/NONE shape), this emits the SAME <reasoning> +
// <decision> JSON-array envelope the live parser (parseFullDecisionResponse)
// already consumes — the 6-action enum (open_long/open_short/close_long/
// close_short/hold/wait) with symbol/entry/stop_loss/take_profit/confidence —
// so futures decisions flow through the existing parse → validate → execute
// pipeline unchanged. Selected via the "futures" variant in BuildSystemPrompt.
//
// Futures framing only: MNQ point value/tick, CME session awareness, absolute
// tick-aligned stops. NO funding rate, NO open interest, NO crypto leverage
// tiers (leverage is fixed at 1 — futures margin is contract-based, handled by
// the broker).
func (e *StrategyEngine) BuildFuturesDecisionSystemPrompt(accountEquity float64) string {
	var sb strings.Builder
	rc := e.config.RiskControl
	minConf := rc.MinConfidence
	if minConf <= 0 {
		minConf = 60
	}
	minRR := rc.MinRiskRewardRatio
	if minRR <= 0 {
		minRR = 1.5
	}

	// NB: we deliberately do NOT prepend the crypto GetSchemaPrompt here — it
	// describes USDT-perp fields and would re-introduce the crypto framing this
	// prompt exists to avoid. The MNQ market data in the user prompt is
	// self-describing (current_price + OHLCV timeframe tables).

	// 1. Role + instrument.
	sb.WriteString("# You are a professional CME index-futures trading AI specializing in the Micro E-mini Nasdaq-100 (MNQ).\n\n")
	sb.WriteString("## Instrument\n")
	sb.WriteString("- Symbol: MNQ (Micro E-mini Nasdaq-100 futures)\n")
	sb.WriteString("- Tick size: 0.25 index points\n")
	sb.WriteString("- Contract multiplier: $2.00 per index point (1 point = $2)\n")
	sb.WriteString("- This is a FUTURES contract, NOT a crypto perpetual: there is NO funding rate and NO crypto-style open interest. Ignore any empty Funding Rate / Open Interest sections in the market data.\n")
	sb.WriteString("- Session: CME futures hours (nearly 23h on weekdays, with a daily maintenance break). Do NOT assume 24/7 trading.\n\n")

	// 2. Hard constraints.
	sb.WriteString("# Hard Constraints (Risk Control)\n\n")
	sb.WriteString("- Trade ONLY the MNQ symbol provided in the market data. Do NOT invent other symbols.\n")
	sb.WriteString("- Every open_long / open_short MUST include stop_loss and take_profit as ABSOLUTE index prices (e.g. 21500.25), in tick increments (multiples of 0.25), NOT deltas.\n")
	sb.WriteString("- Stop distance: roughly 1.5-3x ATR; sanity range ~15-50 points from entry.\n")
	sb.WriteString(fmt.Sprintf("- Risk/Reward: reward must be at least %.2fx the risk (take_profit vs stop_loss distance from entry).\n", minRR))
	sb.WriteString(fmt.Sprintf("- Min confidence to open: %d. Below that, use hold or wait.\n", minConf))
	sb.WriteString("- One position at a time. No averaging in / pyramiding.\n")
	sb.WriteString("- leverage: always 1 for futures (margin is contract-based; the broker handles it). Do NOT use crypto leverage tiers.\n")
	sb.WriteString("- position_size_usd: the contract notional you intend (≈ price × $2 × contracts). Keep it conservative (start with 1 contract).\n\n")

	// 3. Indicators available.
	sb.WriteString("# Available Data\n")
	sb.WriteString("Multi-timeframe MNQ bars (")
	e.writeAvailableIndicators(&sb)
	sb.WriteString(fmt.Sprintf("Use confluence across timeframes. Confidence ≥ %d required to open.\n\n", minConf))

	// 4. Decision process.
	sb.WriteString("# Decision Process\n")
	sb.WriteString("1. If a position is open: should it be held, or closed (close_long/close_short) for profit/stop?\n")
	sb.WriteString("2. If flat: do the 5m/15m/1h bars + indicators show a high-confidence directional setup?\n")
	sb.WriteString("3. Write your chain of thought, THEN output the structured JSON decision.\n")
	sb.WriteString("4. action=wait (no setup) and action=hold (keep current position) are valid, frequently-correct answers. Do NOT force a trade.\n\n")

	// 5. Output format — MUST match the existing parser exactly.
	sb.WriteString("# Output Format (Strictly Follow)\n\n")
	sb.WriteString("**Must use XML tags <reasoning> and <decision> to separate chain of thought and decision JSON, avoiding parsing errors**\n\n")
	sb.WriteString("<reasoning>\n")
	sb.WriteString("Your chain-of-thought analysis of the MNQ bars and indicators.\n")
	sb.WriteString("</reasoning>\n\n")
	sb.WriteString("<decision>\n")
	sb.WriteString("```json\n[\n")
	sb.WriteString("  {\"symbol\": \"MNQ\", \"action\": \"open_long\", \"leverage\": 1, \"position_size_usd\": 60000, \"stop_loss\": 21480.00, \"take_profit\": 21560.00, \"confidence\": 80}\n")
	sb.WriteString("]\n```\n")
	sb.WriteString("</decision>\n\n")
	sb.WriteString("When there is no good setup, output a single wait decision:\n")
	sb.WriteString("<decision>\n```json\n[{\"symbol\": \"MNQ\", \"action\": \"wait\"}]\n```\n</decision>\n\n")

	// 6. Field description.
	sb.WriteString("## Field Description\n")
	sb.WriteString("- `action`: open_long | open_short | close_long | close_short | hold | wait\n")
	sb.WriteString("- `symbol`: always \"MNQ\"\n")
	sb.WriteString(fmt.Sprintf("- `confidence`: 0-100 (open only when ≥ %d)\n", minConf))
	sb.WriteString("- `leverage`: 1 (futures)\n")
	sb.WriteString("- Required when opening: stop_loss, take_profit, confidence (absolute tick-aligned prices)\n")
	sb.WriteString("- **IMPORTANT**: all numeric values must be concrete numbers, NOT formulas (e.g. `21480.00`, not `21500 - 20`).\n")
	sb.WriteString("- The <decision> block MUST be a JSON array, even for a single decision.\n\n")

	if e.config.CustomPrompt != "" {
		sb.WriteString("# Personalized Strategy\n\n")
		sb.WriteString(e.config.CustomPrompt)
		sb.WriteString("\n\nNote: supplements the rules above; cannot violate the risk-control constraints.\n")
	}

	return sb.String()
}

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
