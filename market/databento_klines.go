// Task 12 / Cluster D — futures data-fetch routing to Databento.
//
// getDataFromDatabento mirrors the per-timeframe assembly logic of
// GetWithTimeframes, but every k-line fetch routes to the Databento
// Historical OHLCV endpoint instead of CoinAnk/Hyperliquid. The branch into
// this function lives at the top of GetWithTimeframes, gated by
// IsCMEFuturesSymbol.
//
// Adjustments (locked in the spec, calibrated to the 2026-05-28 cmd/nq_smoke
// foundation test):
//
//   1. Query the continuous symbol (NQ.c.0) directly via
//      timeseries.get_range. Do NOT call symbology.resolve — that endpoint
//      returns HTTP 422 for GLBX.MDP3 (stype_in=continuous → stype_out=
//      raw_symbol), and the trading cycle does not need the resolved
//      contract code.
//
//   2. Probe the Databento Historical-tier availability window. The tier
//      lag is variable (8h observed on 2026-05-28). The fetch first tries
//      end = now() - 1h safety buffer. On HTTP 422
//      dataset_unavailable_range, we parse `available_end` from the error
//      body and retry once with that timestamp.

package market

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"nofx/logger"
	"nofx/provider/databento"
)

// timeframeToDuration maps a kernel timeframe string to one bar's duration.
// Used to compute the fetch window.
func timeframeToDuration(tf string) time.Duration {
	switch tf {
	case "1m":
		return 1 * time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 1 * time.Hour
	case "2h":
		return 2 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return 1 * time.Minute
	}
}

// databentoInterval maps a kernel timeframe to a Databento OHLCV schema
// interval. Databento natively supports ohlcv-1m, ohlcv-1h, ohlcv-1d.
// Anything sub-hourly falls back to 1m and the downstream consumer
// aggregates (existing indicator code reads bar-by-bar).
func databentoInterval(tf string) string {
	switch tf {
	case "1h", "1d":
		return tf
	default:
		return "1m"
	}
}

// availableEndFromError extracts `available_end` from a Databento 422
// dataset_unavailable_range error body. Returns (timestamp, true) when
// found. The error body shape (verified 2026-05-28):
//
//	{
//	  "detail": {
//	    "case": "dataset_unavailable_range",
//	    "message": "…",
//	    "status_code": 422,
//	    "payload": {
//	      "dataset": "GLBX.MDP3",
//	      "start":   "…",
//	      "end":     "…",
//	      "available_end": "2026-05-28T03:59:43.551313000Z"
//	    }
//	  }
//	}
func availableEndFromError(err error) (time.Time, bool) {
	if err == nil {
		return time.Time{}, false
	}
	var apiErr *databento.APIError
	if !errors.As(err, &apiErr) {
		return time.Time{}, false
	}
	if apiErr.StatusCode != 422 {
		return time.Time{}, false
	}
	var parsed struct {
		Detail struct {
			Case    string `json:"case"`
			Payload struct {
				AvailableEnd string `json:"available_end"`
			} `json:"payload"`
		} `json:"detail"`
	}
	if jerr := json.Unmarshal([]byte(apiErr.Body), &parsed); jerr != nil {
		return time.Time{}, false
	}
	if parsed.Detail.Case != "dataset_unavailable_range" {
		return time.Time{}, false
	}
	if parsed.Detail.Payload.AvailableEnd == "" {
		return time.Time{}, false
	}
	t, perr := time.Parse(time.RFC3339Nano, parsed.Detail.Payload.AvailableEnd)
	if perr != nil {
		// Some Databento responses use plain RFC3339 (no fractional seconds).
		t, perr = time.Parse(time.RFC3339, parsed.Detail.Payload.AvailableEnd)
		if perr != nil {
			return time.Time{}, false
		}
	}
	return t.UTC(), true
}

// getKlinesFromDatabento fetches OHLCV klines for a single timeframe of a
// CME futures symbol. Implements Adjustments 1 + 2 above.
func getKlinesFromDatabento(symbol, timeframe string, limit int) ([]Kline, error) {
	if limit <= 0 {
		limit = 200
	}
	client := databento.DefaultClient()
	if client.APIKey == "" {
		return nil, fmt.Errorf("databento: missing DATABENTO_API_KEY")
	}

	interval := databentoInterval(timeframe)
	barDur := timeframeToDuration(timeframe)
	if interval == "1m" {
		// When the caller wants a larger timeframe we still fetch 1m
		// bars; ensure the window is wide enough to satisfy the larger
		// bar count after aggregation.
		barDur = 1 * time.Minute
	}
	// limit+10 bars of headroom for session-boundary gaps + sparse OHLCV
	// at low-liquidity hours (matches the cmd/nq_smoke convention).
	windowSize := time.Duration(limit+10) * barDur

	// First attempt: end at now() minus a 1-hour safety buffer. Usually
	// hits the available_end window directly. If the tier is more
	// lagged today, the retry below corrects.
	end := time.Now().UTC().Add(-1 * time.Hour)
	start := end.Add(-windowSize)
	bars, err := client.GetOHLCV(symbol, interval, start, end)
	if err != nil {
		if availEnd, ok := availableEndFromError(err); ok {
			logger.Infof("ℹ️ databento tier lag detected for %s (available_end=%s); retrying with adjusted window",
				symbol, availEnd.Format(time.RFC3339))
			end = availEnd
			start = end.Add(-windowSize)
			bars, err = client.GetOHLCV(symbol, interval, start, end)
			if err != nil {
				return nil, fmt.Errorf("databento %s %s retry after avail-end probe: %w",
					symbol, interval, err)
			}
		} else {
			return nil, fmt.Errorf("databento %s %s: %w", symbol, interval, err)
		}
	}
	return BarsToKlines(bars), nil
}

// getDataFromDatabento builds the same *Data shape that GetWithTimeframes
// produces for crypto, but every per-timeframe fetch routes to Databento.
// CME futures don't have Binance-style Open Interest or funding rate; both
// fields are zero values (downstream code already tolerates this — Plan 1
// futures prompt does not reference OI/funding).
func getDataFromDatabento(symbol string, timeframes []string, primaryTimeframe string, count int) (*Data, error) {
	timeframeData := make(map[string]*TimeframeSeriesData)
	var primaryKlines []Kline

	for _, tf := range timeframes {
		klines, err := getKlinesFromDatabento(symbol, tf, 200)
		if err != nil {
			logger.Infof("⚠️ Failed to get %s %s K-line from Databento: %v", symbol, tf, err)
			continue
		}
		if len(klines) == 0 {
			logger.Infof("⚠️ %s %s Databento K-line data is empty", symbol, tf)
			continue
		}
		if tf == primaryTimeframe {
			primaryKlines = klines
		}
		seriesData := calculateTimeframeSeries(klines, tf, count)
		timeframeData[tf] = seriesData
	}

	if len(primaryKlines) == 0 {
		return nil, fmt.Errorf("primary timeframe %s K-line data is empty for %s", primaryTimeframe, symbol)
	}
	if isStaleData(primaryKlines, symbol) {
		logger.Infof("⚠️ WARNING: %s detected stale data (consecutive price freeze)", symbol)
		return nil, fmt.Errorf("%s data is stale, possible cache failure", symbol)
	}

	currentPrice := primaryKlines[len(primaryKlines)-1].Close
	currentEMA20 := calculateEMA(primaryKlines, 20)
	currentMACD := calculateMACD(primaryKlines)
	currentRSI7 := calculateRSI(primaryKlines, 7)
	priceChange1h := calculatePriceChangeByBars(primaryKlines, primaryTimeframe, 60)
	priceChange4h := calculatePriceChangeByBars(primaryKlines, primaryTimeframe, 240)

	return &Data{
		Symbol:        symbol,
		CurrentPrice:  currentPrice,
		PriceChange1h: priceChange1h,
		PriceChange4h: priceChange4h,
		CurrentEMA20:  currentEMA20,
		CurrentMACD:   currentMACD,
		CurrentRSI7:   currentRSI7,
		OpenInterest:  &OIData{Latest: 0, Average: 0}, // CME futures: no Binance-style OI
		FundingRate:   0,                              // CME futures: no funding rate
		TimeframeData: timeframeData,
	}, nil
}
