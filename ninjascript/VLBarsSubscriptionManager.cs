// VLBarsSubscriptionManager.cs — Plan 4.4 Stage 1 (C# side of the NT8 bar feed).
//
// Subscribes to NT8 bars across multiple timeframes (native, one BarsRequest
// per timeframe — operator-locked Option 1: no Go-side aggregation), and
// streams `bars_historical` + `bar_update` frames over the EXISTING Plan 1.5
// TCP channel managed by VLTraderTCPClient.
//
// Architecture:
//   - Isolated in a separate class so the proven signal/fill/heartbeat path
//     in VLTraderTCPClient.cs stays byte-stable (ADR-007 spirit). The main
//     AddOn gets only: a field, a constructor call, two switch cases.
//   - Sends through VLTraderTCPClient.SendFrame() (now internal), which
//     reuses the existing writeLock + hand-rolled JSON encoder. NO second
//     unsynchronized socket writer.
//
// NT8 gotchas handled (each one is a separate "bug if missed"):
//   1. TIMEZONE  — bars.GetTime() returns LOCAL time. We convert to UTC
//                  epoch ms using bars.TradingHours.TimeZoneInfo before emit.
//   2. MULTI-BAR — a single .Update event can touch MinIndex..MaxIndex. We
//                  iterate that whole range and emit every bar (NEVER just
//                  the last bar).
//   3. HISTORICAL-DEDUP — track each subscription's last-emitted bar time so
//                  reconnect/restart doesn't re-emit bars the Go side
//                  already has.
//   4. RECONNECT — Connection.ConnectionStatusUpdate is hooked; on a
//                  reconnect, all active BarsRequests are disposed and
//                  recreated (NT8 silently stops updating them otherwise).
//   5. JSON      — no Newtonsoft. Payloads are built as
//                  Dictionary<string, object> and serialized by the same
//                  encoder VLTraderTCPClient uses for signal/fill frames.
//   6. WRITELOCK — every send goes through VLTraderTCPClient.SendFrame()
//                  which already takes writeLock; bar frames cannot
//                  interleave with signal/fill/heartbeat bytes on the wire.
//
// Spec: docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md
//       (Plan 4.4 Deep Spec — "Wire protocol extension"; "NT8 SDK gotchas")
// Protocol: ninjascript/vltrader_tcp_PROTOCOL.md (frames 5-8)

#region Using declarations
using System;
using System.Collections.Generic;
using NinjaTrader.Cbi;
using NinjaTrader.Data;
#endregion

namespace NinjaTrader.NinjaScript.AddOns
{
    /// <summary>
    /// Owns the BarsRequest subscriptions and the per-subscription state needed
    /// to stream OHLCV over the VLTrader TCP wire. Constructed by
    /// VLTraderTCPClient at AddOn-Active time; disposed at AddOn-Terminated.
    /// </summary>
    public class VLBarsSubscriptionManager
    {
        // === Inbound dependencies (injected by VLTraderTCPClient) ===

        /// <summary>
        /// Sends a frame through the parent's writeLock + hand-rolled encoder.
        /// We hold a delegate rather than a back-reference so this class never
        /// reaches into the parent's network state directly.
        /// </summary>
        private readonly Action<string, Dictionary<string, object>> sendFrame;

        private readonly Action<string> logInfo;
        private readonly Action<string> logWarn;

        // === Subscription registry ===

        /// <summary>
        /// Active subscriptions keyed by "SYMBOL|TIMEFRAME". Synchronized via
        /// subsLock for ADD/REMOVE; individual entries are owned by NT8's
        /// data thread once the BarsRequest is firing.
        /// </summary>
        private readonly Dictionary<string, BarsRequestEntry> active =
            new Dictionary<string, BarsRequestEntry>();
        private readonly object subsLock = new object();

        // Default historical depth when bars_subscribe.bars_back is omitted.
        // The Plan 4.4 deep spec recommends 500 (~8.3 ETH hours at 1m,
        // sufficient for EMA200/RSI14 warmup).
        private const int DEFAULT_BARS_BACK = 500;

        public VLBarsSubscriptionManager(
            Action<string, Dictionary<string, object>> sendFrame,
            Action<string> logInfo,
            Action<string> logWarn)
        {
            this.sendFrame = sendFrame;
            this.logInfo   = logInfo  ?? (s => { });
            this.logWarn   = logWarn ?? (s => { });
        }

        // ==============================================================
        // Inbound frame handlers — called by VLTraderTCPClient.HandleFrame
        // ==============================================================

        /// <summary>
        /// Handle a `bars_subscribe` envelope. For each requested timeframe
        /// the manager opens (or reuses) one BarsRequest against NT8's data
        /// engine. Idempotent: re-subscribing returns immediately.
        /// </summary>
        public void HandleBarsSubscribe(Dictionary<string, object> payload)
        {
            if (payload == null)
            {
                logWarn("VLBarsSubscriptionManager: bars_subscribe payload missing");
                return;
            }

            string symbol = TryGetString(payload, "symbol");
            if (string.IsNullOrEmpty(symbol))
            {
                logWarn("VLBarsSubscriptionManager: bars_subscribe missing symbol");
                return;
            }

            int barsBack = TryGetInt(payload, "bars_back", DEFAULT_BARS_BACK);
            if (barsBack <= 0) barsBack = DEFAULT_BARS_BACK;

            var timeframes = TryGetStringList(payload, "timeframes");
            if (timeframes == null || timeframes.Count == 0)
            {
                logWarn("VLBarsSubscriptionManager: bars_subscribe missing timeframes");
                return;
            }

            // Resolve the instrument once per subscribe batch. NT8 caches
            // instruments internally so repeated GetInstrument calls are cheap.
            // FLAG: NT8 API — Instrument.GetInstrument(string) is the same call
            // VLTraderTCPClient.HandleSignal uses; if NT8 ever requires the
            // contract-month suffix here (e.g. "MNQ 03-26"), the Go side will
            // need to send the fully-qualified symbol on subscribe too. For
            // Stage 1 we assume bare root ("MNQ") resolves via the operator's
            // configured trading hours template, matching the signal path.
            var instrument = Instrument.GetInstrument(symbol);
            if (instrument == null)
            {
                logWarn("VLBarsSubscriptionManager: instrument " + symbol + " not found");
                return;
            }

            foreach (var tf in timeframes)
            {
                Subscribe(symbol, tf, barsBack, instrument);
            }
        }

        /// <summary>
        /// Handle a `bars_unsubscribe` envelope. Omitting timeframes disposes
        /// every subscription for the named symbol.
        /// </summary>
        public void HandleBarsUnsubscribe(Dictionary<string, object> payload)
        {
            if (payload == null)
            {
                logWarn("VLBarsSubscriptionManager: bars_unsubscribe payload missing");
                return;
            }

            string symbol = TryGetString(payload, "symbol");
            if (string.IsNullOrEmpty(symbol))
            {
                logWarn("VLBarsSubscriptionManager: bars_unsubscribe missing symbol");
                return;
            }

            var timeframes = TryGetStringList(payload, "timeframes");

            lock (subsLock)
            {
                var toRemove = new List<string>();
                foreach (var kv in active)
                {
                    if (!kv.Value.Symbol.Equals(symbol, StringComparison.OrdinalIgnoreCase))
                        continue;
                    if (timeframes != null && timeframes.Count > 0
                        && !timeframes.Contains(kv.Value.Timeframe))
                        continue;
                    toRemove.Add(kv.Key);
                }
                foreach (var key in toRemove)
                {
                    DisposeEntry(active[key]);
                    active.Remove(key);
                    logInfo("VLBarsSubscriptionManager: unsubscribed " + key);
                }
            }
        }

        // ==============================================================
        // Per-subscription lifecycle
        // ==============================================================

        private void Subscribe(string symbol, string timeframe, int barsBack, Instrument instrument)
        {
            string key = (symbol + "|" + timeframe).ToUpperInvariant();

            lock (subsLock)
            {
                if (active.ContainsKey(key))
                {
                    logInfo("VLBarsSubscriptionManager: subscribe " + key + " — already active, ignoring");
                    return;
                }

                BarsPeriod period;
                try
                {
                    period = MapTimeframe(timeframe);
                }
                catch (Exception ex)
                {
                    logWarn("VLBarsSubscriptionManager: unsupported timeframe " + timeframe + ": " + ex.Message);
                    return;
                }

                // FLAG: NT8 API — BarsRequest constructor + how to set BarsPeriod
                // + how to wire .Update vary slightly across NT8 8.0.x and
                // 8.1.x. The shape below matches the NT8 8.1.6 docs:
                //
                //   var req = new BarsRequest(instrument, barsBack);
                //   req.BarsPeriod = period;
                //   req.TradingHoursInstance = TradingHours.Get("CME US Index Futures ETH");
                //   req.Request((bars, err, msg) => { ... });
                //   req.Update += (s, e) => { ... };
                //
                // If the operator's NT8 build rejects this constructor, the
                // alternative is the two-argument (DateTime from, DateTime to)
                // form — flagged here so the compile error is immediately
                // diagnosable.
                BarsRequest request;
                try
                {
                    request = new BarsRequest(instrument, barsBack);
                    request.BarsPeriod = period;
                }
                catch (Exception ex)
                {
                    logWarn("VLBarsSubscriptionManager: BarsRequest ctor failed for "
                            + key + ": " + ex.Message);
                    return;
                }

                var entry = new BarsRequestEntry
                {
                    Key       = key,
                    Symbol    = symbol,
                    Timeframe = timeframe,
                    Request   = request,
                    LastEmittedTimeUtcMs = 0,
                    HistoricalSent = false
                };
                active[key] = entry;

                // Wire the streaming-update handler BEFORE firing the
                // historical Request — otherwise an event fired between the
                // historical load and the handler attach is lost.
                request.Update += (sender, args) => OnBarsUpdate(entry, args);

                logInfo("VLBarsSubscriptionManager: subscribing " + key + " barsBack=" + barsBack);

                // Fire the historical load.
                try
                {
                    // NT8 8.1: the callback's first arg is the BarsRequest
                    // itself (NOT a Bars collection). The actual Bars data
                    // hangs off request.Bars. We capture `entry` and read
                    // entry.Request.Bars inside EmitHistorical so the path
                    // is consistent with OnBarsUpdate.
                    request.Request((req, errorCode, errorMessage) =>
                    {
                        if (errorCode != ErrorCode.NoError)
                        {
                            logWarn("VLBarsSubscriptionManager: historical load failed for "
                                    + key + ": " + errorCode + " — " + errorMessage);
                            return;
                        }
                        EmitHistorical(entry);
                    });
                }
                catch (Exception ex)
                {
                    logWarn("VLBarsSubscriptionManager: Request() threw for "
                            + key + ": " + ex.Message);
                }
            }
        }

        private void DisposeEntry(BarsRequestEntry entry)
        {
            if (entry == null) return;
            try { entry.Request.Dispose(); } catch { }
        }

        // ==============================================================
        // Historical + streaming emit
        // ==============================================================

        private void EmitHistorical(BarsRequestEntry entry)
        {
            // NT8 8.1: BarsRequest exposes its loaded data via request.Bars.
            var bars = entry.Request != null ? entry.Request.Bars : null;
            if (bars == null || bars.Count == 0)
            {
                logInfo("VLBarsSubscriptionManager: historical " + entry.Key + " — zero bars");
                return;
            }

            var barsList = new List<object>(bars.Count);
            long lastT = 0;
            // FLAG: NT8 API — `bars.Count` and the indexed accessors
            // (bars.GetTime(i), bars.GetOpen(i), ...) are the standard
            // NT8 8.1 API. If the operator's build exposes only `bars[i]`
            // with named properties, this loop needs a minor refactor.
            for (int i = 0; i < bars.Count; i++)
            {
                long t = ToUtcEpochMs(bars.GetTime(i), bars.TradingHours);
                if (t <= entry.LastEmittedTimeUtcMs) continue; // dedup guard
                lastT = t;
                barsList.Add(BuildBarObject(t,
                    bars.GetOpen(i), bars.GetHigh(i),
                    bars.GetLow(i),  bars.GetClose(i),
                    bars.GetVolume(i)));
            }

            entry.HistoricalSent = true;
            if (lastT > entry.LastEmittedTimeUtcMs)
                entry.LastEmittedTimeUtcMs = lastT;

            var payload = new Dictionary<string, object>
            {
                ["symbol"]    = entry.Symbol,
                ["timeframe"] = entry.Timeframe,
                ["bars"]      = barsList
            };
            sendFrame("bars_historical", payload);

            logInfo("VLBarsSubscriptionManager: emitted bars_historical "
                    + entry.Key + " bars=" + barsList.Count);
        }

        private void OnBarsUpdate(BarsRequestEntry entry, BarsUpdateEventArgs args)
        {
            // Historical-load dedup: ignore updates that fire BEFORE the
            // initial Request callback returns. NT8 may fire .Update during
            // backfill on some builds; we drop those, the historical batch
            // already covered the same bar times.
            if (!entry.HistoricalSent) return;

            // NT8 8.1: BarsUpdateEventArgs exposes MinIndex + MaxIndex only.
            // The source Bars collection hangs off the BarsRequest, NOT off
            // the event args (confirmed via operator compile-check —
            // CS1061 'BarsUpdateEventArgs' does not contain a definition
            // for 'Bars'). We read from the captured entry.Request.Bars.
            var bars = entry.Request != null ? entry.Request.Bars : null;
            int minIdx = args.MinIndex;
            int maxIdx = args.MaxIndex;

            if (bars == null || minIdx < 0 || maxIdx < minIdx) return;

            var emitted = new List<object>();
            long maxEmittedT = entry.LastEmittedTimeUtcMs;
            for (int i = minIdx; i <= maxIdx && i < bars.Count; i++)
            {
                long t = ToUtcEpochMs(bars.GetTime(i), bars.TradingHours);
                // We DO emit updates to a bar whose time we've already seen —
                // that's the "this bar is still forming" case. Only dedup
                // against bars STRICTLY OLDER than the latest historical t,
                // which means we never roll back into the historical batch.
                if (t < entry.LastEmittedTimeUtcMs) continue;
                emitted.Add(BuildBarObject(t,
                    bars.GetOpen(i), bars.GetHigh(i),
                    bars.GetLow(i),  bars.GetClose(i),
                    bars.GetVolume(i)));
                if (t > maxEmittedT) maxEmittedT = t;
            }
            if (emitted.Count == 0) return;

            entry.LastEmittedTimeUtcMs = maxEmittedT;

            var payload = new Dictionary<string, object>
            {
                ["symbol"]    = entry.Symbol,
                ["timeframe"] = entry.Timeframe,
                ["bars"]      = emitted
            };
            sendFrame("bar_update", payload);
        }

        // ==============================================================
        // Reconnect handling — called by VLTraderTCPClient when the broker
        // connection (NT8 ↔ data provider, NOT the Go TCP) reconnects
        // ==============================================================

        /// <summary>
        /// Re-create every active BarsRequest. NT8 stops delivering updates
        /// to the original request after a connection drop and gives no
        /// explicit signal — so on reconnect we tear down and re-open every
        /// subscription, preserving the LastEmittedTimeUtcMs cursor so the
        /// Go side doesn't see duplicate historical bars.
        /// </summary>
        public void OnConnectionReconnected()
        {
            lock (subsLock)
            {
                var toRecreate = new List<BarsRequestEntry>(active.Values);
                foreach (var entry in toRecreate)
                {
                    DisposeEntry(entry);
                    active.Remove(entry.Key);
                    logInfo("VLBarsSubscriptionManager: reconnect — re-subscribing " + entry.Key);

                    var instrument = Instrument.GetInstrument(entry.Symbol);
                    if (instrument == null)
                    {
                        logWarn("VLBarsSubscriptionManager: reconnect — instrument "
                                + entry.Symbol + " not found");
                        continue;
                    }
                    // Preserve LastEmittedTimeUtcMs on the new entry so the
                    // freshly-fired historical load can dedup against what
                    // the Go side already received.
                    Subscribe(entry.Symbol, entry.Timeframe, DEFAULT_BARS_BACK, instrument);
                    BarsRequestEntry fresh;
                    if (active.TryGetValue(entry.Key, out fresh))
                    {
                        fresh.LastEmittedTimeUtcMs = entry.LastEmittedTimeUtcMs;
                    }
                }
            }
        }

        /// <summary>
        /// Tear down all subscriptions. Called by VLTraderTCPClient at
        /// AddOn-Terminated time.
        /// </summary>
        public void DisposeAll()
        {
            lock (subsLock)
            {
                foreach (var kv in active) DisposeEntry(kv.Value);
                active.Clear();
            }
        }

        // ==============================================================
        // Timeframe → BarsPeriod mapping
        // ==============================================================

        /// <summary>
        /// Map a Go-side timeframe string (matching
        /// store/strategy.go::normalizeTimeframe) to an NT8 BarsPeriod.
        /// Throws on unsupported values — the caller drops the subscription
        /// and logs.
        /// </summary>
        private static BarsPeriod MapTimeframe(string tf)
        {
            if (string.IsNullOrEmpty(tf))
                throw new ArgumentException("empty timeframe");

            switch (tf)
            {
                // Sub-hour: native Minute multipliers.
                case "1m":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 1   };
                case "3m":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 3   };
                case "5m":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 5   };
                case "15m": return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 15  };
                case "30m": return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 30  };

                // Hourly: Minute*60 is universally supported in NT8. Using
                // BarsPeriodType.Hour with Value=N also works on most builds
                // but Minute multipliers avoid an "unknown enum" risk on
                // older builds.
                case "1h":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 60  };
                case "2h":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 120 };
                case "4h":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 240 };
                case "6h":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 360 };
                case "8h":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 480 };
                case "12h": return new BarsPeriod { BarsPeriodType = BarsPeriodType.Minute, Value = 720 };

                // Daily / weekly.
                case "1d":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Day,    Value = 1   };
                case "3d":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Day,    Value = 3   };
                case "1w":  return new BarsPeriod { BarsPeriodType = BarsPeriodType.Week,   Value = 1   };

                default:
                    throw new ArgumentException("unsupported timeframe: " + tf);
            }
        }

        // ==============================================================
        // Helpers
        // ==============================================================

        /// <summary>
        /// Convert NT8's local-time bar timestamp to UTC epoch milliseconds.
        /// `tradingHours.TimeZoneInfo` is the canonical source per the NT8
        /// help guide ("How Bars Are Built" — bar times reflect the chart's
        /// configured time zone, which derives from the Trading Hours
        /// template attached to the BarsRequest).
        /// </summary>
        private static long ToUtcEpochMs(DateTime localTime, TradingHours tradingHours)
        {
            DateTime utc;
            if (localTime.Kind == DateTimeKind.Utc)
            {
                utc = localTime;
            }
            else
            {
                TimeZoneInfo tz = null;
                if (tradingHours != null)
                {
                    // FLAG: NT8 API — `TradingHours.TimeZoneInfo` is the typical
                    // accessor. Some builds expose it as `tradingHours.TimeZone`
                    // (no "Info" suffix). If the operator hits a compile error,
                    // swap to that name.
                    tz = tradingHours.TimeZoneInfo;
                }
                if (tz == null)
                {
                    // Fall back to the machine's local zone — wrong if the
                    // bars were stamped in CT and the machine is in another
                    // zone, but better than throwing.
                    tz = TimeZoneInfo.Local;
                }
                utc = TimeZoneInfo.ConvertTimeToUtc(
                    DateTime.SpecifyKind(localTime, DateTimeKind.Unspecified), tz);
            }
            var unixEpoch = new DateTime(1970, 1, 1, 0, 0, 0, DateTimeKind.Utc);
            return (long)((utc - unixEpoch).TotalMilliseconds);
        }

        private static Dictionary<string, object> BuildBarObject(
            long tUtcMs, double o, double h, double l, double c, double v)
        {
            return new Dictionary<string, object>
            {
                ["t"] = tUtcMs,
                ["o"] = o,
                ["h"] = h,
                ["l"] = l,
                ["c"] = c,
                ["v"] = v
            };
        }

        private static string TryGetString(Dictionary<string, object> obj, string key)
        {
            object v;
            if (obj != null && obj.TryGetValue(key, out v) && v != null) return v.ToString();
            return null;
        }

        private static int TryGetInt(Dictionary<string, object> obj, string key, int dflt)
        {
            object v;
            if (obj != null && obj.TryGetValue(key, out v) && v != null)
            {
                if (v is long lo)   return (int)lo;
                if (v is int io)    return io;
                if (v is double doo) return (int)doo;
                int n;
                if (int.TryParse(v.ToString(), out n)) return n;
            }
            return dflt;
        }

        private static List<string> TryGetStringList(Dictionary<string, object> obj, string key)
        {
            object v;
            if (obj == null || !obj.TryGetValue(key, out v) || v == null) return null;
            var list = v as List<object>;
            if (list == null) return null;
            var result = new List<string>(list.Count);
            foreach (var item in list)
            {
                if (item == null) continue;
                var s = item.ToString();
                if (!string.IsNullOrEmpty(s)) result.Add(s);
            }
            return result;
        }

        /// <summary>
        /// Per-subscription state. Lives in the active dictionary for the
        /// lifetime of the BarsRequest. Mutated from NT8's data thread under
        /// subsLock (or — once the BarsRequest is firing — directly from the
        /// event callback; LastEmittedTimeUtcMs is the only field touched by
        /// that path).
        /// </summary>
        private class BarsRequestEntry
        {
            public string       Key;
            public string       Symbol;
            public string       Timeframe;
            public BarsRequest  Request;
            public long         LastEmittedTimeUtcMs;
            public bool         HistoricalSent;
        }
    }
}
