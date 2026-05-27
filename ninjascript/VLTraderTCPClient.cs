// VLTraderTCPClient.cs — Plan 1.5 TCP NinjaTrader bridge (alternative to CSV).
//
// NinjaScript AddOn that connects to the VL Trader Go bot via local TCP
// (127.0.0.1:36974) and exchanges length-prefixed JSON frames with the Go-side
// server in provider/ninjatrader/tcp_server.go. This file CANNOT be compiled
// in WSL2 — the operator compiles on Windows via NT8's NinjaScript editor (F5)
// or Visual Studio. See vltrader_tcp_README.md for the install procedure and
// vltrader_tcp_PROTOCOL.md for the wire-format spec.
//
// Spec: docs/superpowers/plans/2026-05-22-nq-databento-ninjatrader.md L4343-4447.
// ADR : docs/adr/ADR-001-csv-bridge-vs-tcp.md.

#region Using declarations
using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Net.Sockets;
using System.Text;
using System.Threading;
using NinjaTrader.Cbi;
using NinjaTrader.Core;
using NinjaTrader.NinjaScript;
#endregion

namespace NinjaTrader.NinjaScript.AddOns
{
    /// <summary>
    /// VLTrader TCP client. Loaded automatically by NT8 when dropped in
    /// Documents\NinjaTrader 8\bin\Custom\AddOns\. Lifecycle is driven by
    /// OnStateChange (SetDefaults / Active / Terminated).
    /// </summary>
    public class VLTraderTCPClient : AddOnBase
    {
        // === Connection constants (must match Go-side tcp_server.go) ===
        private const string GO_SERVER_HOST          = "127.0.0.1";
        private const int    GO_SERVER_PORT          = 36974; // NOT NT's ATI port 36973
        private const int    HEARTBEAT_INTERVAL_MS   = 30000; // spec L4408
        private const int    RECONNECT_INTERVAL_MS   = 5000;  // spec L4415
        private const int    STALE_SIGNAL_AGE_SECONDS = 60;   // spec L4414
        private const int    MAX_FRAME_BYTES         = 1 << 20; // 1 MB, spec L4376

        // === State ===
        private TcpClient                client;
        private NetworkStream            stream;
        private Thread                   readerThread;
        private Thread                   heartbeatThread;
        private CancellationTokenSource  cts;
        private Account                  account;   // primary trading account
        private readonly object          writeLock = new object();

        // Track signal_id → original entry price for slippage calculation.
        private readonly Dictionary<string, double> signalEntryByOco = new Dictionary<string, double>();
        private readonly Dictionary<string, double> signalTickSizeByOco = new Dictionary<string, double>();
        private readonly object signalMapLock = new object();

        protected override void OnStateChange()
        {
            if (State == State.SetDefaults)
            {
                Name = "VLTrader TCP Client";
                Description = "Connects to the VL Trader Go bot via local TCP (Plan 1.5). " +
                              "Receives signals over the wire, places OCO bracket orders, " +
                              "and emits fills back. Falls back to CSV bridge if disabled.";
            }
            else if (State == State.Active)
            {
                cts = new CancellationTokenSource();
                ResolveAccount();
                if (account != null)
                {
                    account.OrderUpdate += OnOrderUpdate;
                }
                readerThread = new Thread(() => RunConnectionLoop(cts.Token))
                {
                    IsBackground = true,
                    Name = "VLTrader-TCP-Reader"
                };
                heartbeatThread = new Thread(() => RunHeartbeatLoop(cts.Token))
                {
                    IsBackground = true,
                    Name = "VLTrader-TCP-Heartbeat"
                };
                readerThread.Start();
                heartbeatThread.Start();
                LogInfo("VLTraderTCPClient: AddOn Active, connecting to "
                        + GO_SERVER_HOST + ":" + GO_SERVER_PORT);
            }
            else if (State == State.Terminated)
            {
                try { cts?.Cancel(); } catch { }
                if (account != null)
                {
                    try { account.OrderUpdate -= OnOrderUpdate; } catch { }
                }
                try { stream?.Close(); } catch { }
                try { client?.Close(); } catch { }
                LogInfo("VLTraderTCPClient: AddOn Terminated");
            }
        }

        // === Account resolution ===
        // For SIM testing: prefer "Sim101", else the first available account.
        private void ResolveAccount()
        {
            lock (Account.All)
            {
                foreach (var a in Account.All)
                {
                    if (a.Name == "Sim101") { account = a; return; }
                }
                if (Account.All.Count > 0) account = Account.All[0];
            }
            if (account == null)
            {
                LogWarn("VLTraderTCPClient: no account resolved; signals will be rejected");
            }
            else
            {
                LogInfo("VLTraderTCPClient: using account " + account.Name);
            }
        }

        // === Connection loop with reconnect (spec L4415: every 5s) ===
        private void RunConnectionLoop(CancellationToken ct)
        {
            while (!ct.IsCancellationRequested)
            {
                try
                {
                    client = new TcpClient();
                    client.Connect(GO_SERVER_HOST, GO_SERVER_PORT);
                    stream = client.GetStream();
                    LogInfo("VLTraderTCPClient: connected");
                    RunReadLoop(ct);
                }
                catch (Exception ex)
                {
                    LogWarn("VLTraderTCPClient: connect failed: " + ex.Message);
                }
                finally
                {
                    try { stream?.Close(); } catch { }
                    try { client?.Close(); } catch { }
                    stream = null;
                    client = null;
                }
                if (!ct.IsCancellationRequested)
                {
                    Thread.Sleep(RECONNECT_INTERVAL_MS);
                }
            }
        }

        // === Read loop: 4-byte BE length prefix + JSON body (spec L4376) ===
        private void RunReadLoop(CancellationToken ct)
        {
            byte[] header = new byte[4];
            while (!ct.IsCancellationRequested && stream != null)
            {
                if (!ReadExact(stream, header, 4)) return;
                uint length =
                    ((uint)header[0] << 24) |
                    ((uint)header[1] << 16) |
                    ((uint)header[2] << 8)  |
                    ((uint)header[3]);
                if (length > MAX_FRAME_BYTES)
                {
                    // Spec L4416: oversized frame is a protocol error; close.
                    LogWarn("VLTraderTCPClient: oversized frame " + length + " bytes — closing");
                    return;
                }
                byte[] body = new byte[length];
                if (!ReadExact(stream, body, (int)length)) return;
                string json = Encoding.UTF8.GetString(body);
                HandleFrame(json);
            }
        }

        private bool ReadExact(NetworkStream s, byte[] buf, int n)
        {
            int got = 0;
            while (got < n)
            {
                int r;
                try
                {
                    r = s.Read(buf, got, n - got);
                }
                catch (Exception ex)
                {
                    LogWarn("VLTraderTCPClient: read err " + ex.Message);
                    return false;
                }
                if (r <= 0) return false;
                got += r;
            }
            return true;
        }

        // === Frame dispatch ===
        private void HandleFrame(string json)
        {
            try
            {
                var root = (Dictionary<string, object>)new JsonParser(json).Parse();
                string type = GetString(root, "type");
                Dictionary<string, object> payload = null;
                if (root.ContainsKey("payload"))
                {
                    payload = root["payload"] as Dictionary<string, object>;
                }
                if (type == "signal")
                {
                    HandleSignal(payload);
                }
                else if (type == "heartbeat")
                {
                    // Ack incoming heartbeats per spec L4410.
                    SendAck("heartbeat");
                }
                else if (type == "ack")
                {
                    // Server-side ack of our heartbeat/fill — informational only.
                }
                else
                {
                    LogWarn("VLTraderTCPClient: unknown frame type " + type);
                }
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: malformed frame: " + ex.Message);
            }
        }

        // === Signal handling: OCO bracket placement (spec L4387-4396) ===
        private void HandleSignal(Dictionary<string, object> p)
        {
            if (p == null)
            {
                LogWarn("VLTraderTCPClient: signal payload missing or not an object");
                return;
            }
            string symbol;
            string side;
            int qty;
            double entry;
            double sl;
            double tp;
            string signalId;
            string ts;
            try
            {
                symbol   = GetString(p, "symbol");
                side     = GetString(p, "side");                 // "long" | "short"
                qty      = GetInt(p, "quantity");
                entry    = GetDouble(p, "entry");
                sl       = GetDouble(p, "stop_loss");
                tp       = GetDouble(p, "take_profit");
                signalId = GetString(p, "signal_id");
                ts       = GetString(p, "timestamp");
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: signal payload missing field: " + ex.Message);
                return;
            }

            // Spec L4414: stale signal rejection at 60s.
            if (DateTime.TryParse(ts, null,
                System.Globalization.DateTimeStyles.RoundtripKind, out var sigTime))
            {
                var ageSec = (DateTime.UtcNow - sigTime.ToUniversalTime()).TotalSeconds;
                if (ageSec > STALE_SIGNAL_AGE_SECONDS)
                {
                    LogWarn("VLTraderTCPClient: stale signal " + signalId
                            + " (age " + ageSec.ToString("F1") + "s) — rejecting");
                    SendFillFrame(signalId, 0.0, side, qty, 0.0, "rejected");
                    return;
                }
            }

            if (account == null)
            {
                LogWarn("VLTraderTCPClient: no account — rejecting signal " + signalId);
                SendFillFrame(signalId, 0.0, side, qty, 0.0, "rejected");
                return;
            }

            var instrument = Instrument.GetInstrument(symbol);
            if (instrument == null)
            {
                LogWarn("VLTraderTCPClient: instrument " + symbol + " not found — rejecting");
                SendFillFrame(signalId, 0.0, side, qty, 0.0, "rejected");
                return;
            }

            // Track entry + tick size for slippage computation in fill emission.
            double tickSize = 0.25; // sensible NQ default; overridden if instrument exposes it
            try { tickSize = instrument.MasterInstrument.TickSize; } catch { }
            lock (signalMapLock)
            {
                signalEntryByOco[signalId] = entry;
                signalTickSizeByOco[signalId] = tickSize;
            }

            // OCO bracket: entry market + SL stop + TP limit, all sharing the
            // signal_id as the OCO group id so fills can be correlated.
            try
            {
                var ocoId = signalId;
                var entryAction = side == "long" ? OrderAction.Buy : OrderAction.SellShort;
                var exitAction  = side == "long" ? OrderAction.Sell : OrderAction.BuyToCover;

                var entryOrder = account.CreateOrder(
                    instrument, entryAction, OrderType.Market, OrderEntry.Manual,
                    TimeInForce.Day, qty, 0, 0, ocoId, signalId,
                    Core.Globals.MaxDate, null);

                var slOrder = account.CreateOrder(
                    instrument, exitAction, OrderType.StopMarket, OrderEntry.Manual,
                    TimeInForce.Day, qty, 0, sl, ocoId, signalId + "-sl",
                    Core.Globals.MaxDate, null);

                var tpOrder = account.CreateOrder(
                    instrument, exitAction, OrderType.Limit, OrderEntry.Manual,
                    TimeInForce.Day, qty, tp, 0, ocoId, signalId + "-tp",
                    Core.Globals.MaxDate, null);

                account.Submit(new[] { entryOrder, slOrder, tpOrder });
                LogInfo("VLTraderTCPClient: submitted OCO bracket signal_id=" + signalId
                        + " " + side + " " + qty + " " + symbol
                        + " entry≈" + entry + " sl=" + sl + " tp=" + tp);
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: order submit failed: " + ex.Message);
                SendFillFrame(signalId, 0.0, side, qty, 0.0, "rejected");
            }
        }

        // === Fill subscription (spec L4398-4406) ===
        private void OnOrderUpdate(object sender, OrderEventArgs e)
        {
            // Only emit fill frames for filled or rejected states; ignore the
            // intermediate accepted/working transitions.
            if (e.OrderState != OrderState.Filled
                && e.OrderState != OrderState.Rejected
                && e.OrderState != OrderState.PartFilled)
            {
                return;
            }

            // Oco group for the entry order is the signal_id; child orders use
            // signal_id-sl / signal_id-tp. We only emit a fill for the entry
            // and exits — both share the same Oco group so the Go side can
            // correlate via signal_id.
            string ocoId = e.Order != null ? e.Order.Oco : null;
            if (string.IsNullOrEmpty(ocoId)) return;

            // Trim "-sl" / "-tp" suffix if present (these are exit-leg fills).
            string signalId = ocoId;
            if (signalId.EndsWith("-sl") || signalId.EndsWith("-tp"))
            {
                signalId = signalId.Substring(0, signalId.Length - 3);
            }

            string status;
            switch (e.OrderState)
            {
                case OrderState.Filled:     status = "filled";   break;
                case OrderState.Rejected:   status = "rejected"; break;
                case OrderState.PartFilled: status = "partial";  break;
                default:                    return;
            }

            string sideStr = (e.Order.OrderAction == OrderAction.Buy
                              || e.Order.OrderAction == OrderAction.BuyToCover)
                              ? "long"
                              : "short";

            double slippageTicks = 0.0;
            lock (signalMapLock)
            {
                if (signalEntryByOco.TryGetValue(signalId, out var origEntry)
                    && signalTickSizeByOco.TryGetValue(signalId, out var tick)
                    && tick > 0)
                {
                    slippageTicks = (e.AverageFillPrice - origEntry) / tick;
                }
            }

            SendFillFrame(signalId, e.AverageFillPrice, sideStr, e.Filled,
                          slippageTicks, status);
        }

        // === Write helpers ===
        private void SendFillFrame(string signalId, double fillPrice, string side,
                                   int qty, double slippageTicks, string status)
        {
            var payload = new Dictionary<string, object>
            {
                ["signal_id"]      = signalId,
                ["fill_price"]     = fillPrice,
                ["fill_time"]      = DateTime.UtcNow.ToString("o"),
                ["side"]           = side,
                ["quantity"]       = qty,
                ["slippage_ticks"] = slippageTicks,
                ["status"]         = status
            };
            WriteEnvelope("fill", payload);
        }

        private void SendAck(string acks)
        {
            WriteEnvelope("ack", new Dictionary<string, object> { ["acks"] = acks });
        }

        private void WriteEnvelope(string type, Dictionary<string, object> payload)
        {
            if (stream == null) return;
            try
            {
                string json = EncodeFrame(type, payload);
                byte[] body = Encoding.UTF8.GetBytes(json);
                if (body.Length > MAX_FRAME_BYTES)
                {
                    LogWarn("VLTraderTCPClient: outgoing frame too large " + body.Length);
                    return;
                }
                byte[] header = new byte[]
                {
                    (byte)((body.Length >> 24) & 0xFF),
                    (byte)((body.Length >> 16) & 0xFF),
                    (byte)((body.Length >> 8)  & 0xFF),
                    (byte)( body.Length        & 0xFF)
                };
                lock (writeLock)
                {
                    stream.Write(header, 0, 4);
                    stream.Write(body, 0, body.Length);
                    stream.Flush();
                }
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: write err " + ex.Message);
            }
        }

        // === Heartbeat loop (spec L4408: 30s interval) ===
        private void RunHeartbeatLoop(CancellationToken ct)
        {
            while (!ct.IsCancellationRequested)
            {
                Thread.Sleep(HEARTBEAT_INTERVAL_MS);
                if (ct.IsCancellationRequested) return;
                WriteEnvelope("heartbeat", new Dictionary<string, object>());
            }
        }

        // === Log helpers — NT8 Log API varies slightly across versions; isolate ===
        private static void LogInfo(string msg)
        {
            try { NinjaScript.Log(msg, LogLevel.Information); } catch { }
        }

        private static void LogWarn(string msg)
        {
            try { NinjaScript.Log(msg, LogLevel.Warning); } catch { }
        }

        // ==============================================================
        // Hand-rolled JSON encoder + parser (no external deps).
        // NT8 user AddOns cannot resolve Newtonsoft.Json reliably, so we
        // implement only what the 4-message wire protocol needs:
        // {type, payload} envelopes with flat snake_case payloads.
        // On-wire bytes must match Go encoding/json: snake_case keys,
        // invariant-culture numbers, compact (no pretty-print), UTF-8.
        // ==============================================================

        private static string EncodeFrame(string type, Dictionary<string, object> payload)
        {
            var sb = new StringBuilder();
            sb.Append('{');
            sb.Append("\"type\":");
            AppendString(sb, type);
            sb.Append(",\"payload\":");
            AppendObject(sb, payload);
            sb.Append('}');
            return sb.ToString();
        }

        private static void AppendObject(StringBuilder sb, Dictionary<string, object> obj)
        {
            sb.Append('{');
            if (obj != null)
            {
                bool first = true;
                foreach (var kv in obj)
                {
                    if (!first) sb.Append(',');
                    first = false;
                    AppendString(sb, kv.Key);
                    sb.Append(':');
                    AppendValue(sb, kv.Value);
                }
            }
            sb.Append('}');
        }

        private static void AppendValue(StringBuilder sb, object v)
        {
            if (v == null) { sb.Append("null"); return; }
            if (v is string s) { AppendString(sb, s); return; }
            if (v is bool b) { sb.Append(b ? "true" : "false"); return; }
            if (v is int i) { sb.Append(i.ToString(CultureInfo.InvariantCulture)); return; }
            if (v is long l) { sb.Append(l.ToString(CultureInfo.InvariantCulture)); return; }
            if (v is double d) { sb.Append(d.ToString("R", CultureInfo.InvariantCulture)); return; }
            if (v is float f) { sb.Append(((double)f).ToString("R", CultureInfo.InvariantCulture)); return; }
            if (v is decimal m) { sb.Append(m.ToString(CultureInfo.InvariantCulture)); return; }
            if (v is Dictionary<string, object> nested) { AppendObject(sb, nested); return; }
            AppendString(sb, v.ToString());
        }

        private static void AppendString(StringBuilder sb, string s)
        {
            sb.Append('"');
            if (s != null)
            {
                foreach (char c in s)
                {
                    switch (c)
                    {
                        case '"':  sb.Append("\\\""); break;
                        case '\\': sb.Append("\\\\"); break;
                        case '\n': sb.Append("\\n"); break;
                        case '\r': sb.Append("\\r"); break;
                        case '\t': sb.Append("\\t"); break;
                        case '\b': sb.Append("\\b"); break;
                        case '\f': sb.Append("\\f"); break;
                        default:
                            if (c < 0x20) sb.AppendFormat("\\u{0:x4}", (int)c);
                            else sb.Append(c);
                            break;
                    }
                }
            }
            sb.Append('"');
        }

        // === Field-extraction helpers (typed reads from parsed dictionary) ===
        private static string GetString(Dictionary<string, object> obj, string key)
        {
            if (obj != null && obj.TryGetValue(key, out var v) && v != null) return v.ToString();
            return null;
        }

        private static double GetDouble(Dictionary<string, object> obj, string key)
        {
            if (obj != null && obj.TryGetValue(key, out var v) && v != null)
            {
                if (v is double d) return d;
                if (v is long l) return (double)l;
                if (v is int i) return (double)i;
                if (v is decimal m) return (double)m;
                return Convert.ToDouble(v, CultureInfo.InvariantCulture);
            }
            return 0.0;
        }

        private static int GetInt(Dictionary<string, object> obj, string key)
        {
            if (obj != null && obj.TryGetValue(key, out var v) && v != null)
            {
                if (v is long l) return (int)l;
                if (v is int i) return i;
                if (v is double d) return (int)d;
                return Convert.ToInt32(v, CultureInfo.InvariantCulture);
            }
            return 0;
        }

        // === Recursive-descent JSON parser ===
        private class JsonParser
        {
            private readonly string s;
            private int pos;

            public JsonParser(string input)
            {
                s = input ?? string.Empty;
                pos = 0;
            }

            public object Parse()
            {
                SkipWs();
                return ParseValue();
            }

            private object ParseValue()
            {
                SkipWs();
                if (pos >= s.Length) throw new Exception("unexpected EOF");
                char c = s[pos];
                if (c == '{') return ParseObject();
                if (c == '"') return ParseString();
                if (c == 't' || c == 'f') return ParseBool();
                if (c == 'n') return ParseNull();
                if (c == '-' || (c >= '0' && c <= '9')) return ParseNumber();
                if (c == '[') return ParseArray();
                throw new Exception("unexpected char '" + c + "' at " + pos);
            }

            private Dictionary<string, object> ParseObject()
            {
                var result = new Dictionary<string, object>();
                pos++; // consume '{'
                SkipWs();
                if (pos < s.Length && s[pos] == '}') { pos++; return result; }
                while (true)
                {
                    SkipWs();
                    string key = ParseString();
                    SkipWs();
                    if (pos >= s.Length || s[pos] != ':') throw new Exception("expected ':' at " + pos);
                    pos++; // consume ':'
                    SkipWs();
                    object value = ParseValue();
                    result[key] = value;
                    SkipWs();
                    if (pos < s.Length && s[pos] == ',') { pos++; continue; }
                    if (pos < s.Length && s[pos] == '}') { pos++; return result; }
                    throw new Exception("expected ',' or '}' at " + pos);
                }
            }

            private string ParseString()
            {
                if (pos >= s.Length || s[pos] != '"') throw new Exception("expected string at " + pos);
                pos++; // consume opening '"'
                var sb = new StringBuilder();
                while (pos < s.Length && s[pos] != '"')
                {
                    if (s[pos] == '\\' && pos + 1 < s.Length)
                    {
                        char esc = s[pos + 1];
                        switch (esc)
                        {
                            case '"':  sb.Append('"'); pos += 2; break;
                            case '\\': sb.Append('\\'); pos += 2; break;
                            case '/':  sb.Append('/'); pos += 2; break;
                            case 'n':  sb.Append('\n'); pos += 2; break;
                            case 'r':  sb.Append('\r'); pos += 2; break;
                            case 't':  sb.Append('\t'); pos += 2; break;
                            case 'b':  sb.Append('\b'); pos += 2; break;
                            case 'f':  sb.Append('\f'); pos += 2; break;
                            case 'u':
                                if (pos + 5 < s.Length)
                                {
                                    sb.Append((char)Convert.ToInt32(s.Substring(pos + 2, 4), 16));
                                    pos += 6;
                                }
                                else
                                {
                                    pos++;
                                }
                                break;
                            default: sb.Append(esc); pos += 2; break;
                        }
                    }
                    else
                    {
                        sb.Append(s[pos]);
                        pos++;
                    }
                }
                if (pos >= s.Length) throw new Exception("unterminated string");
                pos++; // consume closing '"'
                return sb.ToString();
            }

            private object ParseNumber()
            {
                int start = pos;
                if (s[pos] == '-') pos++;
                while (pos < s.Length &&
                       (char.IsDigit(s[pos]) || s[pos] == '.' ||
                        s[pos] == 'e' || s[pos] == 'E' ||
                        s[pos] == '+' || s[pos] == '-'))
                {
                    pos++;
                }
                string num = s.Substring(start, pos - start);
                if (num.IndexOfAny(new[] { '.', 'e', 'E' }) >= 0)
                {
                    return double.Parse(num, CultureInfo.InvariantCulture);
                }
                if (long.TryParse(num, NumberStyles.Integer, CultureInfo.InvariantCulture, out long l))
                {
                    return l;
                }
                return double.Parse(num, CultureInfo.InvariantCulture);
            }

            private bool ParseBool()
            {
                if (pos + 4 <= s.Length && s.Substring(pos, 4) == "true") { pos += 4; return true; }
                if (pos + 5 <= s.Length && s.Substring(pos, 5) == "false") { pos += 5; return false; }
                throw new Exception("invalid bool at " + pos);
            }

            private object ParseNull()
            {
                if (pos + 4 <= s.Length && s.Substring(pos, 4) == "null") { pos += 4; return null; }
                throw new Exception("invalid null at " + pos);
            }

            private List<object> ParseArray()
            {
                var result = new List<object>();
                pos++; // consume '['
                SkipWs();
                if (pos < s.Length && s[pos] == ']') { pos++; return result; }
                while (true)
                {
                    SkipWs();
                    result.Add(ParseValue());
                    SkipWs();
                    if (pos < s.Length && s[pos] == ',') { pos++; continue; }
                    if (pos < s.Length && s[pos] == ']') { pos++; return result; }
                    throw new Exception("expected ',' or ']' at " + pos);
                }
            }

            private void SkipWs()
            {
                while (pos < s.Length &&
                       (s[pos] == ' ' || s[pos] == '\t' ||
                        s[pos] == '\n' || s[pos] == '\r'))
                {
                    pos++;
                }
            }
        }
    }
}
