package market

import (
	"context"
	"fmt"
	"nofx/provider/coinank/coinank_api"
	"nofx/provider/coinank/coinank_enum"
	"nofx/provider/hyperliquid"
	"strconv"
	"strings"
	"time"
)

// Note: Kline data now uses free/open API (coinank_api.Kline) which doesn't require authentication

// CoinAnkVenue is the ONE mapping from a trader's exchange to its CoinAnk
// kline venue, used by the market reads and by /api/klines alike. An exchange
// with no CoinAnk venue of its own — or no exchange at all — is REFUSED by
// name: its klines are never read from another venue's book (there is no
// default venue and no cross-venue fallback).
func CoinAnkVenue(exchange string) (coinank_enum.Exchange, error) {
	switch strings.ToLower(strings.TrimSpace(exchange)) {
	case "bybit":
		return coinank_enum.Bybit, nil
	case "okx":
		return coinank_enum.Okex, nil
	case "bitget":
		return coinank_enum.Bitget, nil
	case "gate":
		return coinank_enum.Gate, nil
	case "hyperliquid":
		return coinank_enum.Hyperliquid, nil
	case "aster":
		return coinank_enum.Aster, nil
	default:
		return "", fmt.Errorf("no CoinAnk venue for exchange %q — its klines are not read from another venue", exchange)
	}
}

// CoinAnkAPISymbol is the symbol spelling CoinAnk expects on a venue: OKX
// lists perpetuals as "BTC-USDT-SWAP", every other venue as "BTCUSDT".
func CoinAnkAPISymbol(symbol string, venue coinank_enum.Exchange) string {
	if venue == coinank_enum.Okex && strings.HasSuffix(symbol, "USDT") {
		return fmt.Sprintf("%s-USDT-SWAP", strings.TrimSuffix(symbol, "USDT"))
	}
	return symbol
}

// getKlinesFromCoinAnk fetches kline data from CoinAnk API (replacement for WSMonitorCli)
func getKlinesFromCoinAnk(symbol, interval, exchange string, limit int) ([]Kline, error) {
	// Map interval string to coinank enum
	var coinankInterval coinank_enum.Interval
	switch interval {
	case "1m":
		coinankInterval = coinank_enum.Minute1
	case "3m":
		coinankInterval = coinank_enum.Minute3
	case "5m":
		coinankInterval = coinank_enum.Minute5
	case "15m":
		coinankInterval = coinank_enum.Minute15
	case "30m":
		coinankInterval = coinank_enum.Minute30
	case "1h":
		coinankInterval = coinank_enum.Hour1
	case "2h":
		coinankInterval = coinank_enum.Hour2
	case "4h":
		coinankInterval = coinank_enum.Hour4
	case "6h":
		coinankInterval = coinank_enum.Hour6
	case "8h":
		coinankInterval = coinank_enum.Hour8
	case "12h":
		coinankInterval = coinank_enum.Hour12
	case "1d":
		coinankInterval = coinank_enum.Day1
	case "3d":
		coinankInterval = coinank_enum.Day3
	case "1w":
		coinankInterval = coinank_enum.Week1
	default:
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// The venue is resolved by the ONE mapper (CoinAnkVenue): an exchange with
	// no CoinAnk venue is REFUSED by name before any request — never read from
	// another venue's book.
	coinankExchange, err := CoinAnkVenue(exchange)
	if err != nil {
		return nil, err
	}

	// Call CoinAnk free/open API (no authentication required). One request,
	// one venue: a failed or empty answer is an error, never a second request
	// to a different venue.
	ctx := context.Background()
	ts := time.Now().UnixMilli()
	// Use "To" side to search backward from current time (get historical klines)
	coinankKlines, err := coinank_api.Kline(ctx, CoinAnkAPISymbol(symbol, coinankExchange), coinankExchange, ts, coinank_enum.To, limit, coinankInterval)
	if err != nil {
		return nil, fmt.Errorf("CoinAnk API error (%s): %w", coinankExchange, err)
	}
	if len(coinankKlines) == 0 {
		return nil, fmt.Errorf("CoinAnk returned no %s klines for %s on %s", interval, symbol, coinankExchange)
	}

	// Convert coinank kline format to market.Kline format
	klines := make([]Kline, len(coinankKlines))
	for i, ck := range coinankKlines {
		klines[i] = Kline{
			OpenTime:  ck.StartTime,
			Open:      ck.Open,
			High:      ck.High,
			Low:       ck.Low,
			Close:     ck.Close,
			Volume:    ck.Volume,
			CloseTime: ck.EndTime,
		}
	}

	return klines, nil
}

// getKlinesFromHyperliquid fetches kline data from Hyperliquid API for xyz dex assets
func getKlinesFromHyperliquid(symbol, interval string, limit int) ([]Kline, error) {
	// Remove xyz: prefix if present for the API call
	baseCoin := strings.TrimPrefix(symbol, "xyz:")

	// Map interval to Hyperliquid format
	hlInterval := hyperliquid.MapTimeframe(interval)

	// Create Hyperliquid client
	client := hyperliquid.NewClient()

	// Fetch candles
	ctx := context.Background()
	candles, err := client.GetCandles(ctx, baseCoin, hlInterval, limit)
	if err != nil {
		return nil, fmt.Errorf("Hyperliquid API error: %w", err)
	}

	// Convert to market.Kline format
	klines := make([]Kline, len(candles))
	for i, c := range candles {
		open, _ := strconv.ParseFloat(c.Open, 64)
		high, _ := strconv.ParseFloat(c.High, 64)
		low, _ := strconv.ParseFloat(c.Low, 64)
		closePrice, _ := strconv.ParseFloat(c.Close, 64)
		volume, _ := strconv.ParseFloat(c.Volume, 64)

		klines[i] = Kline{
			OpenTime:  c.OpenTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    volume,
			CloseTime: c.CloseTime,
		}
	}

	return klines, nil
}

// calculateTimeframeSeries calculates series data for a single timeframe.
// ip carries the strategy-CONFIGURED indicator periods (optional); for each
// non-empty list a series is computed per period into the *ByPeriod maps, IN
// ADDITION to the legacy fixed fields (EMA20/50, BOLL 20, RSI 7/14, ATR 14),
// which are always computed for back-compat readers. With an empty/zero ip the
// output is byte-identical to the pre-change code.
func calculateTimeframeSeries(klines []Kline, timeframe string, count int, ip IndicatorPeriods) *TimeframeSeriesData {
	if count <= 0 {
		count = 10 // default
	}

	data := &TimeframeSeriesData{
		Timeframe:   timeframe,
		Klines:      make([]KlineBar, 0, count),
		MidPrices:   make([]float64, 0, count),
		EMA20Values: make([]float64, 0, count),
		EMA50Values: make([]float64, 0, count),
		MACDValues:  make([]float64, 0, count),
		RSI7Values:  make([]float64, 0, count),
		RSI14Values: make([]float64, 0, count),
		Volume:      make([]float64, 0, count),
		BOLLUpper:   make([]float64, 0, count),
		BOLLMiddle:  make([]float64, 0, count),
		BOLLLower:   make([]float64, 0, count),
	}
	if len(ip.EMA) > 0 {
		data.EMAByPeriod = make(map[int][]float64, len(ip.EMA))
		for _, p := range ip.EMA {
			if p > 0 {
				data.EMAByPeriod[p] = make([]float64, 0, count)
			}
		}
	}
	if len(ip.RSI) > 0 {
		data.RSIByPeriod = make(map[int][]float64, len(ip.RSI))
		for _, p := range ip.RSI {
			if p > 0 {
				data.RSIByPeriod[p] = make([]float64, 0, count)
			}
		}
	}
	if len(ip.BOLL) > 0 {
		data.BOLLByPeriod = make(map[int]*BollBands, len(ip.BOLL))
		for _, p := range ip.BOLL {
			if p > 0 {
				data.BOLLByPeriod[p] = &BollBands{}
			}
		}
	}

	// Get latest N data points based on count from config
	start := len(klines) - count
	if start < 0 {
		start = 0
	}

	for i := start; i < len(klines); i++ {
		// Store full OHLCV kline data
		data.Klines = append(data.Klines, KlineBar{
			Time:   klines[i].OpenTime,
			Open:   klines[i].Open,
			High:   klines[i].High,
			Low:    klines[i].Low,
			Close:  klines[i].Close,
			Volume: klines[i].Volume,
		})

		// Keep MidPrices and Volume for backward compatibility
		data.MidPrices = append(data.MidPrices, klines[i].Close)
		data.Volume = append(data.Volume, klines[i].Volume)

		// Calculate EMA20 for each point
		if i >= 19 {
			ema20 := calculateEMA(klines[:i+1], 20)
			data.EMA20Values = append(data.EMA20Values, ema20)
		}

		// Calculate EMA50 for each point
		if i >= 49 {
			ema50 := calculateEMA(klines[:i+1], 50)
			data.EMA50Values = append(data.EMA50Values, ema50)
		}

		// Calculate the strategy-CONFIGURED EMA periods (e.g. 9/21/200) for each
		// point — same "need `period` bars" guard as the fixed EMAs above.
		for _, p := range ip.EMA {
			if p > 0 && i >= p-1 {
				data.EMAByPeriod[p] = append(data.EMAByPeriod[p], calculateEMA(klines[:i+1], p))
			}
		}

		// Calculate MACD for each point
		if i >= 25 {
			macd := calculateMACD(klines[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}

		// Calculate RSI for each point
		if i >= 7 {
			rsi7 := calculateRSI(klines[:i+1], 7)
			data.RSI7Values = append(data.RSI7Values, rsi7)
		}
		if i >= 14 {
			rsi14 := calculateRSI(klines[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}

		// Configured RSI periods (same i>=period guard as RSI7/RSI14 above).
		for _, p := range ip.RSI {
			if p > 0 && i >= p {
				data.RSIByPeriod[p] = append(data.RSIByPeriod[p], calculateRSI(klines[:i+1], p))
			}
		}

		// Calculate Bollinger Bands (period 20, std dev multiplier 2)
		if i >= 19 {
			upper, middle, lower := calculateBOLL(klines[:i+1], 20, 2.0)
			data.BOLLUpper = append(data.BOLLUpper, upper)
			data.BOLLMiddle = append(data.BOLLMiddle, middle)
			data.BOLLLower = append(data.BOLLLower, lower)
		}

		// Configured BOLL periods (std-dev multiplier fixed at 2; same i>=period-1
		// guard as the fixed period-20 BOLL above).
		for _, p := range ip.BOLL {
			if p > 0 && i >= p-1 {
				u, m, l := calculateBOLL(klines[:i+1], p, 2.0)
				if b := data.BOLLByPeriod[p]; b != nil {
					b.Upper = append(b.Upper, u)
					b.Middle = append(b.Middle, m)
					b.Lower = append(b.Lower, l)
				}
			}
		}
	}

	// Calculate ATR14 (legacy fixed) + the configured ATR periods (latest value).
	data.ATR14 = calculateATR(klines, 14)
	if len(ip.ATR) > 0 {
		data.ATRByPeriod = make(map[int]float64, len(ip.ATR))
		for _, p := range ip.ATR {
			if p > 0 {
				data.ATRByPeriod[p] = calculateATR(klines, p)
			}
		}
	}

	return data
}

// calculatePriceChangeByBars calculates how many K-lines to look back for price change based on timeframe
func calculatePriceChangeByBars(klines []Kline, timeframe string, targetMinutes int) float64 {
	if len(klines) < 2 {
		return 0
	}

	// Parse timeframe to minutes
	tfMinutes := parseTimeframeToMinutes(timeframe)
	if tfMinutes <= 0 {
		return 0
	}

	// Calculate how many K-lines to look back
	barsBack := targetMinutes / tfMinutes
	if barsBack < 1 {
		barsBack = 1
	}

	currentPrice := klines[len(klines)-1].Close
	idx := len(klines) - 1 - barsBack
	if idx < 0 {
		idx = 0
	}

	oldPrice := klines[idx].Close
	if oldPrice > 0 {
		return ((currentPrice - oldPrice) / oldPrice) * 100
	}
	return 0
}

// parseTimeframeToMinutes parses timeframe string to minutes
func parseTimeframeToMinutes(tf string) int {
	switch tf {
	case "1m":
		return 1
	case "3m":
		return 3
	case "5m":
		return 5
	case "15m":
		return 15
	case "30m":
		return 30
	case "1h":
		return 60
	case "2h":
		return 120
	case "4h":
		return 240
	case "6h":
		return 360
	case "8h":
		return 480
	case "12h":
		return 720
	case "1d":
		return 1440
	case "3d":
		return 4320
	case "1w":
		return 10080
	default:
		return 0
	}
}

// calculateIntradaySeries calculates intraday series data
func calculateIntradaySeries(klines []Kline) *IntradayData {
	data := &IntradayData{
		MidPrices:   make([]float64, 0, 10),
		EMA20Values: make([]float64, 0, 10),
		MACDValues:  make([]float64, 0, 10),
		RSI7Values:  make([]float64, 0, 10),
		RSI14Values: make([]float64, 0, 10),
		Volume:      make([]float64, 0, 10),
	}

	// Get latest 10 data points
	start := len(klines) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(klines); i++ {
		data.MidPrices = append(data.MidPrices, klines[i].Close)
		data.Volume = append(data.Volume, klines[i].Volume)

		// Calculate EMA20 for each point
		if i >= 19 {
			ema20 := calculateEMA(klines[:i+1], 20)
			data.EMA20Values = append(data.EMA20Values, ema20)
		}

		// Calculate MACD for each point
		if i >= 25 {
			macd := calculateMACD(klines[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}

		// Calculate RSI for each point
		if i >= 7 {
			rsi7 := calculateRSI(klines[:i+1], 7)
			data.RSI7Values = append(data.RSI7Values, rsi7)
		}
		if i >= 14 {
			rsi14 := calculateRSI(klines[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}
	}

	// Calculate 3m ATR14
	data.ATR14 = calculateATR(klines, 14)

	return data
}

// calculateLongerTermData calculates longer-term data
func calculateLongerTermData(klines []Kline) *LongerTermData {
	data := &LongerTermData{
		MACDValues:  make([]float64, 0, 10),
		RSI14Values: make([]float64, 0, 10),
	}

	// Calculate EMA
	data.EMA20 = calculateEMA(klines, 20)
	data.EMA50 = calculateEMA(klines, 50)

	// Calculate ATR
	data.ATR3 = calculateATR(klines, 3)
	data.ATR14 = calculateATR(klines, 14)

	// Calculate volume
	if len(klines) > 0 {
		data.CurrentVolume = klines[len(klines)-1].Volume
		// Calculate average volume
		sum := 0.0
		for _, k := range klines {
			sum += k.Volume
		}
		data.AverageVolume = sum / float64(len(klines))
	}

	// Calculate MACD and RSI series
	start := len(klines) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(klines); i++ {
		if i >= 25 {
			macd := calculateMACD(klines[:i+1])
			data.MACDValues = append(data.MACDValues, macd)
		}
		if i >= 14 {
			rsi14 := calculateRSI(klines[:i+1], 14)
			data.RSI14Values = append(data.RSI14Values, rsi14)
		}
	}

	return data
}

// GetBoxData fetches 1h klines and calculates box data for a symbol on the
// trader's venue. The read goes through marketRouteFor like every other market
// read: a CME futures symbol has no box source here and is REFUSED (the grid
// is a crypto strategy), the NinjaTrader venue with a non-CME symbol is
// REFUSED, an xyz asset reads Hyperliquid, and a crypto symbol reads its own
// venue's CoinAnk book — an exchange with no CoinAnk venue is refused by name.
func GetBoxData(symbol, venue string) (*BoxData, error) {
	symbol = Normalize(symbol)

	// Fetch 500 1h klines
	var klines []Kline
	var err error

	switch marketRouteFor(symbol, venue) {
	case routeFutures:
		return nil, fmt.Errorf("%s: box data is a crypto grid read — a CME futures symbol has no box source (refused)", symbol)
	case routeRefusedVenue:
		return nil, fmt.Errorf("%s: the NinjaTrader venue reads CME futures only — a non-CME symbol is refused (no crypto market read on the futures venue)", symbol)
	}

	if IsXyzDexAsset(symbol) {
		klines, err = getKlinesFromHyperliquid(symbol, "1h", LongBoxPeriod)
	} else {
		klines, err = getKlinesFromCoinAnk(symbol, "1h", venue, LongBoxPeriod)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get 1h klines: %w", err)
	}

	if len(klines) == 0 {
		return nil, fmt.Errorf("no kline data available")
	}

	currentPrice := klines[len(klines)-1].Close

	return calculateBoxData(klines, currentPrice), nil
}
