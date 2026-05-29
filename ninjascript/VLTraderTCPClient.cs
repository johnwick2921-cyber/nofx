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

        // Pending protective brackets, keyed by signal_id. The SL/TP are placed
        // only AFTER the entry fills (see OnOrderUpdate) — placing them in the
        // entry's OCO group caused NT to cancel them the instant the entry
        // filled (OCO = one-cancels-other), leaving the position unprotected.
        private readonly Dictionary<string, PendingBracket> pendingBrackets = new Dictionary<string, PendingBracket>();

        private class PendingBracket
        {
            public Instrument  Instrument;
            public OrderAction ExitAction;
            public int         Qty;
            public double      Sl;
            public double      Tp;
        }

        // Plan 4.4 Stage 1 — multi-timeframe BarsRequest subscriptions. Owns
        // its own state; calls back into SendFrame() which serializes through
        // writeLock so bar frames cannot interleave with signal/fill bytes.
        private VLBarsSubscriptionManager barsManager;

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
                    // Plan 4.11 — emit the real account balance on cash/PnL
                    // changes so the dashboard reflects the live SIM account.
                    account.AccountItemUpdate += OnAccountItemUpdate;
                }
                // Plan 4.4 Stage 1 — instantiate the bars manager BEFORE the
                // reader thread starts so an early bars_subscribe frame has a
                // destination. SendFrame is wired through this instance so
                // bar frames inherit the same writeLock + encoder used by the
                // proven signal/fill/heartbeat path.
                barsManager = new VLBarsSubscriptionManager(SendFrame, LogInfo, LogWarn);
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
                    try { account.AccountItemUpdate -= OnAccountItemUpdate; } catch { }
                }
                try { barsManager?.DisposeAll(); } catch { }
                try { stream?.Close(); } catch { }
                try { client?.Close(); } catch { }
                LogInfo("VLTraderTCPClient: AddOn Terminated");
            }
        }

        // === Account resolution ===
        // Selection order (first match wins):
        //   1. The account name in %USERPROFILE%\NofxTrader\account.txt
        //      (operator-editable — one line, e.g. "Sim101" or "MyPropAcct").
        //   2. "Sim101" (SIM default).
        //   3. The first available account.
        private void ResolveAccount()
        {
            string preferred = null;
            try
            {
                string path = Path.Combine(
                    Environment.GetEnvironmentVariable("USERPROFILE") ?? "",
                    "NofxTrader", "account.txt");
                if (File.Exists(path))
                {
                    preferred = File.ReadAllText(path).Trim();
                    if (preferred.Length == 0) preferred = null;
                }
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: could not read account.txt: " + ex.Message);
            }

            lock (Account.All)
            {
                if (preferred != null)
                {
                    foreach (var a in Account.All)
                        if (a.Name == preferred) { account = a; break; }
                    if (account == null)
                        LogWarn("VLTraderTCPClient: account.txt requested '" + preferred
                                + "' but no such account — falling back");
                }
                if (account == null)
                    foreach (var a in Account.All)
                        if (a.Name == "Sim101") { account = a; break; }
                if (account == null && Account.All.Count > 0) account = Account.All[0];
            }

            if (account == null)
            {
                LogWarn("VLTraderTCPClient: no account resolved; signals will be rejected");
            }
            else
            {
                LogInfo("VLTraderTCPClient: using account " + account.Name
                        + (preferred != null && account.Name == preferred ? " (from account.txt)" : ""));
            }
        }

        // === SIM detection ===
        // Returns true if the account is a simulation account. Tries Account.Simulation
        // property first (NT8 8.1+), then falls back to name pattern matching (Sim*).
        private bool IsSimAccount(Account a)
        {
            if (a == null) return false;
            try
            {
                // NT8 8.1+ exposes Account.Simulation property
                var simProp = a.GetType().GetProperty("Simulation");
                if (simProp != null && simProp.PropertyType == typeof(bool))
                {
                    return (bool)simProp.GetValue(a, null);
                }
            }
            catch { }
            // Fallback: name-based detection (Sim prefix is NT8 standard for SIM accounts)
            return a.Name != null && a.Name.StartsWith("Sim");
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
                    LogInfo("VLTraderTCPClient: CONNECTED");
                    LogInfo("VLTraderTCPClient: About to call SendAccountsList");

                    // Trigger on connect in case accounts are already loaded
                    SendAccountsList();
                    LogInfo("VLTraderTCPClient: SendAccountsList completed");

                    // Plan 4.11 — push the current account balance immediately
                    // on (re)connect so the dashboard shows the real SIM
                    // account without waiting for the next AccountItemUpdate.
                    SendAccountBalance();
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
                else if (type == "close_position")
                {
                    HandleClosePosition(payload);
                }
                else if (type == "account_select")
                {
                    HandleAccountSelect(payload);
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
                else if (type == "bars_subscribe")
                {
                    // Plan 4.4 Stage 1 — forward to the bars manager.
                    barsManager?.HandleBarsSubscribe(payload);
                }
                else if (type == "bars_unsubscribe")
                {
                    // Plan 4.4 Stage 1 — forward to the bars manager.
                    barsManager?.HandleBarsUnsubscribe(payload);
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

            // Resolve the bare root ("MNQ") to the NT8 front-month
            // contract ("MNQ 06-26") via VLContractResolver (date-derived,
            // auto-rolls). NT8's Instrument.GetInstrument requires the
            // qualified contract; a bare root returns null. Canonical
            // symbol on the wire stays "MNQ"; contract form only exists
            // inside this GetInstrument call.
            string contract = VLContractResolver.ResolveFrontMonthContract(symbol);
            LogInfo("VLTraderTCPClient: resolved " + symbol + " -> " + contract);
            var instrument = Instrument.GetInstrument(contract);
            if (instrument == null)
            {
                LogWarn("VLTraderTCPClient: instrument " + symbol
                        + " (resolved to " + contract + ") not found — rejecting");
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

            // Submit the ENTRY only; the protective SL/TP are placed once the
            // entry fills (OnOrderUpdate → SubmitBracketOnEntryFill). Submitting
            // all three in one OCO group cancels the SL/TP the instant the
            // Market entry fills (OCO = one-cancels-other) — the position then
            // has no working stop/target in NT8 and never auto-closes. The
            // entry stands alone (empty OCO); SL+TP get their own OCO pair.
            try
            {
                var entryAction = side == "long" ? OrderAction.Buy : OrderAction.SellShort;
                var exitAction  = side == "long" ? OrderAction.Sell : OrderAction.BuyToCover;

                var entryOrder = account.CreateOrder(
                    instrument, entryAction, OrderType.Market, OrderEntry.Manual,
                    TimeInForce.Day, qty, 0, 0, string.Empty, signalId,
                    Core.Globals.MaxDate, null);

                lock (signalMapLock)
                {
                    pendingBrackets[signalId] = new PendingBracket
                    {
                        Instrument = instrument,
                        ExitAction = exitAction,
                        Qty        = qty,
                        Sl         = sl,
                        Tp         = tp,
                    };
                }

                account.Submit(new[] { entryOrder });
                LogInfo("VLTraderTCPClient: submitted entry signal_id=" + signalId
                        + " " + side + " " + qty + " " + symbol
                        + " entry≈" + entry + " (SL=" + sl + " TP=" + tp + " on fill)");
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: order submit failed: " + ex.Message);
                lock (signalMapLock) { pendingBrackets.Remove(signalId); }
                SendFillFrame(signalId, 0.0, side, qty, 0.0, "rejected");
            }
        }

        // === Manual close: flatten the position + cancel its working orders ===
        // account.Flatten closes the position at market AND cancels the bracket's
        // SL/TP, so no orphaned exit orders are left to re-open a position. The
        // flatten's market fill arrives in OnOrderUpdate as an exit (Sell /
        // BuyToCover) and is reported as a position_close with reason "manual".
        private void HandleClosePosition(Dictionary<string, object> p)
        {
            if (p == null) { LogWarn("VLTraderTCPClient: close_position empty payload"); return; }
            if (account == null) { LogWarn("VLTraderTCPClient: close_position no account"); return; }
            try
            {
                string symbol = GetString(p, "symbol");
                string contract = VLContractResolver.ResolveFrontMonthContract(symbol);
                var instrument = Instrument.GetInstrument(contract);
                if (instrument == null)
                {
                    LogWarn("VLTraderTCPClient: close_position instrument not found " + symbol);
                    return;
                }
                account.Flatten(new[] { instrument });
                LogInfo("VLTraderTCPClient: flatten " + symbol + " (" + contract + ")");
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: close_position failed: " + ex.Message);
            }
        }

        // === Account switching (Go-server → C#-AddOn command) ===
        // Unsubscribe from the old account's events, resolve the new account by name,
        // subscribe to its events, clear pending state tied to the old account,
        // and re-emit account_balance for the new account.
        private void HandleAccountSelect(Dictionary<string, object> p)
        {
            if (p == null) { LogWarn("VLTraderTCPClient: account_select empty payload"); return; }
            try
            {
                string requestedAccount = GetString(p, "account");
                if (string.IsNullOrEmpty(requestedAccount))
                {
                    LogWarn("VLTraderTCPClient: account_select missing account name");
                    return;
                }

                // Resolve the requested account from Account.All
                Account newAccount = null;
                lock (Account.All)
                {
                    foreach (var a in Account.All)
                        if (a.Name == requestedAccount) { newAccount = a; break; }
                }

                if (newAccount == null)
                {
                    LogWarn("VLTraderTCPClient: account_select account not found: " + requestedAccount);
                    return;
                }

                // If already on that account, no-op
                if (account == newAccount)
                {
                    LogInfo("VLTraderTCPClient: account_select already on " + requestedAccount);
                    return;
                }

                // Unsubscribe from the old account's events
                if (account != null)
                {
                    try { account.OrderUpdate -= OnOrderUpdate; } catch { }
                    try { account.AccountItemUpdate -= OnAccountItemUpdate; } catch { }
                }

                // Switch to the new account and subscribe to its events
                account = newAccount;
                account.OrderUpdate += OnOrderUpdate;
                account.AccountItemUpdate += OnAccountItemUpdate;

                // Clear pending brackets and signal maps (tied to old account context)
                lock (signalMapLock)
                {
                    signalEntryByOco.Clear();
                    signalTickSizeByOco.Clear();
                    pendingBrackets.Clear();
                }

                // Re-emit account_balance immediately for the new account
                SendAccountBalance();
                LogInfo("VLTraderTCPClient: switched to account " + account.Name);
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: account_select failed: " + ex.Message);
            }
        }

        // === Fill subscription (spec L4398-4406) ===
        private void OnOrderUpdate(object sender, OrderEventArgs e)
        {
            // Only act on filled/rejected states; ignore accepted/working.
            if (e.OrderState != OrderState.Filled
                && e.OrderState != OrderState.Rejected
                && e.OrderState != OrderState.PartFilled)
            {
                return;
            }
            if (e.Order == null) return;

            // Classify by ORDER ACTION, not name: Buy/SellShort = entry,
            // Sell/BuyToCover = exit (an SL/TP leg OR a manual flatten — a
            // flatten order may have no -sl/-tp name, so name alone is not
            // enough). The order Name still carries the -sl/-tp suffix when
            // present → the exit reason; otherwise the exit is a "manual" close.
            string orderName = e.Order.Name ?? "";
            string ocoId     = e.Order.Oco ?? "";
            string signalId  = orderName.Length > 0 ? orderName : ocoId;
            string exitReason = null;
            if (signalId.EndsWith("-sl")) { exitReason = "sl"; signalId = signalId.Substring(0, signalId.Length - 3); }
            else if (signalId.EndsWith("-tp")) { exitReason = "tp"; signalId = signalId.Substring(0, signalId.Length - 3); }

            var action = e.Order.OrderAction;
            bool isExit = (action == OrderAction.Sell || action == OrderAction.BuyToCover);

            // Exit fill (SL, TP, or manual flatten) → the position closed. Emit
            // position_close with the real exit price (reason from the leg name,
            // else "manual"). Only on a full Filled — PartFilled is transient (NT
            // sends a final Filled; per-partial would duplicate the close) and a
            // rejection is an error. Held side is opposite the exit action.
            if (isExit)
            {
                if (e.OrderState == OrderState.Filled)
                {
                    string positionSide = (action == OrderAction.BuyToCover) ? "short" : "long";
                    string rootSymbol = "";
                    try { rootSymbol = e.Order.Instrument.MasterInstrument.Name; } catch { }
                    SendPositionCloseFrame(signalId, rootSymbol, positionSide,
                                           e.AverageFillPrice, e.Filled, exitReason ?? "manual");
                }
                return;
            }

            // Entry leg (Buy = long, SellShort = short) → fill-frame path.
            string status;
            switch (e.OrderState)
            {
                case OrderState.Filled:     status = "filled";   break;
                case OrderState.Rejected:   status = "rejected"; break;
                case OrderState.PartFilled: status = "partial";  break;
                default:                    return;
            }

            // Entry just filled → place the protective SL/TP now that the
            // position exists (deferred from HandleSignal to dodge the OCO
            // cancel-on-entry-fill bug). On rejection, drop the pending bracket.
            if (e.OrderState == OrderState.Filled)
            {
                SubmitBracketOnEntryFill(signalId);
            }
            else if (e.OrderState == OrderState.Rejected)
            {
                lock (signalMapLock) { pendingBrackets.Remove(signalId); }
            }

            string sideStr = (action == OrderAction.Buy) ? "long" : "short";

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

        // Place the protective SL + TP once the entry has filled. They share
        // their OWN OCO group (signal_id-exit) so one cancels the other when
        // triggered, and — crucially — they are NOT in the entry's OCO group,
        // so the entry fill no longer cancels them. Names keep the -sl/-tp
        // suffix so OnOrderUpdate routes their fills to position_close.
        private void SubmitBracketOnEntryFill(string signalId)
        {
            PendingBracket b;
            lock (signalMapLock)
            {
                if (!pendingBrackets.TryGetValue(signalId, out b)) return;
                pendingBrackets.Remove(signalId); // idempotent: only place once
            }
            try
            {
                string exitOco = signalId + "-exit";
                var slOrder = account.CreateOrder(
                    b.Instrument, b.ExitAction, OrderType.StopMarket, OrderEntry.Manual,
                    TimeInForce.Day, b.Qty, 0, b.Sl, exitOco, signalId + "-sl",
                    Core.Globals.MaxDate, null);
                var tpOrder = account.CreateOrder(
                    b.Instrument, b.ExitAction, OrderType.Limit, OrderEntry.Manual,
                    TimeInForce.Day, b.Qty, b.Tp, 0, exitOco, signalId + "-tp",
                    Core.Globals.MaxDate, null);
                account.Submit(new[] { slOrder, tpOrder });
                LogInfo("VLTraderTCPClient: placed protective bracket signal_id=" + signalId
                        + " sl=" + b.Sl + " tp=" + b.Tp);
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: bracket submit failed signal_id=" + signalId
                        + ": " + ex.Message);
            }
        }

        // Position-history fix — emit a position_close frame when an OCO exit
        // leg (SL/TP) fills. position_side is the side that was HELD. The Go
        // side computes realized PnL against the recorded entry × the futures
        // point value, so we send only the raw exit price here.
        private void SendPositionCloseFrame(string signalId, string symbol,
                                            string positionSide, double exitPrice,
                                            int qty, string exitReason)
        {
            var payload = new Dictionary<string, object>
            {
                ["signal_id"]     = signalId,
                ["symbol"]        = symbol ?? "",
                ["position_side"] = positionSide,
                ["exit_price"]    = exitPrice,
                ["quantity"]      = qty,
                ["exit_reason"]   = exitReason ?? "",
                ["exit_time"]     = DateTime.UtcNow.ToString("o")
            };
            WriteEnvelope("position_close", payload);
        }

        // Emit the accounts_list frame reporting all available NT accounts with
        // SIM detection. Fired on connect so the Go server can populate the UI
        // account selector. Also used on account list change (if NT fires an event).
        // WriteEnvelope no-ops if not yet connected.
        // Handler for when accounts become available
        private void OnAccountStatusUpdate(object sender, AccountStatusEventArgs e)
        {
            LogInfo(string.Format("@@@ OnAccountStatusUpdate FIRED: {0}, status={1}", e.Account.Name, e.Status));
            SendAccountsList();
        }

        private bool IsRealAccount(Account a)
        {
            if (a == null) return false;

            // Skip NT8 internal/test/bracket accounts (auto-generated).
            if (a.Name.StartsWith("FTPROPLUSM") || a.Name.StartsWith("FTPROPLUS") ||
                a.Name.StartsWith("TAKEPROFIT") || a.Name.StartsWith("TDFYSL") ||
                a.Name.StartsWith("TDFYG") || a.Name.StartsWith("MFFU"))
                return false;

            // Include only real trading accounts (SIM and live funded accounts).
            // Exclude the massive number of latent LFE slots that NT8 creates for
            // each possible margin level but never uses. Real accounts are:
            // - Sim* (simulation)
            // - LFE* followed by exactly 15 digits (live funded, e.g. LFE05060792090061)
            // - Playback* (backtest/replay accounts)
            // - Backtest (generic backtest)
            if (a.Name.StartsWith("Sim") || a.Name.StartsWith("Playback") ||
                a.Name == "Backtest")
                return true;

            // For LFE accounts: only include if it's a real account number (15 digits).
            // Exclude latent LFE slots that don't match the pattern.
            if (a.Name.StartsWith("LFE"))
            {
                string suffix = a.Name.Substring(3);
                // Real LFE accounts have exactly 15-17 digit account numbers
                if (suffix.Length >= 15 && suffix.Length <= 17)
                {
                    // Check if it's all digits (real account number)
                    bool isNumeric = true;
                    foreach (char c in suffix)
                    {
                        if (!char.IsDigit(c))
                        {
                            isNumeric = false;
                            break;
                        }
                    }
                    if (isNumeric) return true;
                }
                return false;
            }

            return false;
        }

        private void SendAccountsList()
        {
            LogInfo("@@@ SendAccountsList START");
            var accountsList = new List<object>();
            try
            {
                lock (Account.All)
                {
                    LogInfo(string.Format("@@@ Account.All.Count={0}", Account.All.Count));
                    foreach (var a in Account.All)
                    {
                        if (!IsRealAccount(a)) continue;
                        if (accountsList.Count < 5)  // Log first 5
                            LogInfo(string.Format("@@@   account[{0}]: {1}", accountsList.Count, a.Name));
                        accountsList.Add(new Dictionary<string, object>
                        {
                            ["name"]   = a.Name,
                            ["is_sim"] = IsSimAccount(a)
                        });
                    }
                }
            }
            catch (Exception ex)
            {
                LogWarn(string.Format("@@@ lock failed: {0}", ex.Message));
                return;
            }

            LogInfo(string.Format("@@@ built list with {0} accounts, calling WriteEnvelope", accountsList.Count));
            LogInfo(string.Format("@@@ stream status: {0}", stream == null ? "NULL" : "CONNECTED"));
            var payload = new Dictionary<string, object> { ["accounts"] = accountsList };
            try
            {
                WriteEnvelope("accounts_list", payload);
                LogInfo(string.Format("@@@ WriteEnvelope succeeded, sent {0} accounts", accountsList.Count));
            }
            catch (Exception ex)
            {
                LogWarn(string.Format("@@@ WriteEnvelope FAILED: {0}", ex.Message));
            }
        }

        // Plan 4.11 — emit the real NT SIM account balance as an account_balance
        // frame (replaces the Go-side $50k mock). Fired on (re)connect and on
        // AccountItemUpdate. WriteEnvelope no-ops if not yet connected.
        //
        // FLAG: NT8 API — account.Get(AccountItem.X, Currency.UsDollar) and the
        // AccountItem enum (CashValue/BuyingPower/RealizedProfitLoss/
        // UnrealizedProfitLoss) match NT8 8.1. If the operator's build differs,
        // adjust the enum names here (compile error will point right at it).
        // NetLiquidation is derived (cash + unrealized) rather than via a
        // possibly-absent AccountItem.NetLiquidation.
        private void SendAccountBalance()
        {
            if (account == null) return;
            try
            {
                double cash       = account.Get(AccountItem.CashValue, Currency.UsDollar);
                double buying     = account.Get(AccountItem.BuyingPower, Currency.UsDollar);
                double realized   = account.Get(AccountItem.RealizedProfitLoss, Currency.UsDollar);
                double unrealized = account.Get(AccountItem.UnrealizedProfitLoss, Currency.UsDollar);
                var payload = new Dictionary<string, object>
                {
                    ["account"]         = account.Name,
                    ["cash_value"]      = cash,
                    ["buying_power"]    = buying,
                    ["realized_pnl"]    = realized,
                    ["unrealized_pnl"]  = unrealized,
                    ["net_liquidation"] = cash + unrealized
                };
                WriteEnvelope("account_balance", payload);
            }
            catch (Exception ex)
            {
                LogWarn("VLTraderTCPClient: account_balance emit failed: " + ex.Message);
            }
        }

        private void OnAccountItemUpdate(object sender, AccountItemEventArgs e)
        {
            // Emit only on cash/PnL changes to limit frame churn.
            if (e.AccountItem == AccountItem.CashValue
                || e.AccountItem == AccountItem.BuyingPower
                || e.AccountItem == AccountItem.RealizedProfitLoss
                || e.AccountItem == AccountItem.UnrealizedProfitLoss)
            {
                SendAccountBalance();
            }
        }

        private void SendAck(string acks)
        {
            WriteEnvelope("ack", new Dictionary<string, object> { ["acks"] = acks });
        }

        /// <summary>
        /// Plan 4.4 Stage 1 — entry point for VLBarsSubscriptionManager.
        /// Routes its bar frames (bars_historical, bar_update) through the
        /// SAME WriteEnvelope path the signal/fill/heartbeat code uses, so
        /// bar bytes inherit the writeLock and the hand-rolled encoder.
        /// </summary>
        internal void SendFrame(string type, Dictionary<string, object> payload)
        {
            WriteEnvelope(type, payload);
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
            int beatCount = 0;
            while (!ct.IsCancellationRequested)
            {
                Thread.Sleep(HEARTBEAT_INTERVAL_MS);
                if (ct.IsCancellationRequested) return;
                WriteEnvelope("heartbeat", new Dictionary<string, object>());

                // Periodic re-emit of accounts every 3 heartbeats (90s) to ensure
                // Go always has them after a restart (not just on connect/change)
                beatCount++;
                if (beatCount % 3 == 0)
                {
                    SendAccountsList();
                }
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
            // Plan 4.4 bar-payload fix: List<object> / arrays. Must come AFTER
            // string + Dictionary (string is IEnumerable<char>; Dictionary is
            // IEnumerable<KeyValuePair>). Without this, bars_historical /
            // bar_update payloads (["bars"] = List<object>) get stringified
            // via the .ToString() fallback to "System.Collections.Generic.
            // List`1[System.Object]", and Go rejects them with
            // "cannot unmarshal string into ... []Bar".
            if (v is System.Collections.IEnumerable e)
            {
                sb.Append('[');
                bool firstItem = true;
                foreach (var item in e)
                {
                    if (!firstItem) sb.Append(',');
                    firstItem = false;
                    AppendValue(sb, item);
                }
                sb.Append(']');
                return;
            }
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
