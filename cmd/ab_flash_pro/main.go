// Command ab_flash_pro (RESEARCH-FLASH-AB, 2026-09-26, DS-101) — offline
// DeepSeek Flash vs Pro planner A/B. Iterates the corpus
// (planner_rejected_prompts attempt=1, prompt > 100 KB, since 2026-09-10),
// calls each arm at reasoning_effort=max with thinking enabled, and runs each
// answer through the offline judge with the LIVE strategy's AuthoringOpts:
// kernel.ParsePlanDocForAuthoring (maxLevels/scenarioCap/opts resolved from the
// strategy row in the DB COPY) + kernel.ValidatePlanDocWithFactsMachine with the
// stored facts JSON (machine map nil — the machine table is NOT persisted, so
// the machine-label legs are skipped exactly as the nil-map contract states).
//
// Arms: (a) deepseek-v4-pro cap 65536; (b) deepseek-flash cap 65536;
// (c) deepseek-flash cap 131072 ONLY on rows where (b) finished length.
// System prompt = the live plannerSystemPrompt verbatim (trader/auto_trader_
// planner.go:28); user prompt = the stored prompt_text verbatim.
//
// DB: the COPY only (ab.db). Crypto key: read-only from the .env given via
// -env (default /home/hoang/nofx/.env); the key is never printed or written.
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"nofx/config"
	"nofx/crypto"
	"nofx/kernel"
	"nofx/market"
	"nofx/store"
	"nofx/trader"

	"github.com/joho/godotenv"
)

const systemPrompt = "You are a disciplined CME index-futures day-plan reasoner. Output ONLY the single JSON object requested — reasoning first, then the answer fields. No prose outside the JSON."

const deepseekURL = "https://api.deepseek.com/chat/completions"

type arm struct {
	label string
	model string
	cap   int
}

type result struct {
	RowID          int
	Arm            string
	Model          string
	Cap            int
	WallS          float64
	TTFBS          float64
	PromptTok      int
	CompletionTok  int
	ReasoningChars int
	Finish         string
	Pass           bool
	FirstReject    string
	CorePass       bool
	CoreReject     string
	FullPass       bool
	FullReject     string
	EntryDistsPts  string // per scenario, joined with ';'
	EntryDistsATR  string // per scenario in xATR5m, joined with ';'
	BornDead       string // yes/no/na + reason (born-dead leg)
	Err            string
	cfg            *store.StrategyConfig // not serialized; stage-2 judge input
	prompt         string                // the stored verbatim prompt
	rawContent     string                // the AI's RAW answer, saved per call
	doc            *kernel.PlanDoc       // parsed doc, for stage-2 judge legs
	facts          kernel.PlanFacts
	maxLevels      int
	scenarioCap    int
	symbol         string
	session        string
	tradeDate      string
	readAt         time.Time
}

type apiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type apiChoice struct {
	Message struct {
		Content          string `json:"content"`
		ReasoningContent string `json:"reasoning_content"`
	} `json:"message"`
	FinishReason *string `json:"finish_reason"`
}

type apiResp struct {
	Choices []apiChoice `json:"choices"`
	Usage   apiUsage    `json:"usage"`
}

func main() {
	dbPath := flag.String("db", "ab.db", "the DB COPY to read")
	envPath := flag.String("env", "/home/hoang/nofx/.env", "the LIVE .env, read-only")
	rowsFlag := flag.String("rows", "all", "comma-separated corpus ids, or 'all'")
	armsFlag := flag.String("arms", "a,b", "arm subset: a,b,c")
	conc := flag.Int("conc", 3, "max concurrent API calls")
	out := flag.String("out", "ab_results.csv", "per-row CSV")
	rejudge := flag.String("rejudge", "", "re-judge the raw outputs in this dir with ZERO new API calls and write -out")
	flag.Parse()

	if *rejudge != "" {
		runRejudge(*rejudge, *out)
		return
	}

	_ = godotenv.Load(*envPath)
	config.Init()
	cs, err := crypto.NewCryptoService()
	if err != nil {
		fmt.Fprintln(os.Stderr, "crypto:", err)
		os.Exit(2)
	}
	crypto.SetGlobalCryptoService(cs)

	st, err := store.New(*dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "store:", err)
		os.Exit(2)
	}
	defer st.Close()

	rows, err := st.PlannerRejected().CorpusRows("2026-09-10", 1, 100000)
	if err != nil {
		fmt.Fprintln(os.Stderr, "corpus:", err)
		os.Exit(2)
	}
	if *rowsFlag != "all" {
		want := map[int]bool{}
		for _, p := range strings.Split(*rowsFlag, ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
				want[n] = true
			}
		}
		kept := rows[:0]
		for _, r := range rows {
			if want[int(r.ID)] {
				kept = append(kept, r)
			}
		}
		rows = kept
	}
	fmt.Printf("corpus rows selected: %d\n", len(rows))

	modelRow, err := st.AIModel().GetAnyEnabled()
	if err != nil || modelRow == nil {
		fmt.Fprintln(os.Stderr, "no enabled AI model on the copy:", err)
		os.Exit(2)
	}
	key := modelRow.APIKey.String()
	if key == "" {
		fmt.Fprintln(os.Stderr, "enabled AI model has no API key")
		os.Exit(2)
	}

	arms := map[string]arm{
		"a": {label: "a-pro-65536", model: "deepseek-v4-pro", cap: 65536},
		"b": {label: "b-flash-65536", model: "deepseek-flash", cap: 65536},
		"c": {label: "c-flash-131072", model: "deepseek-flash", cap: 131072},
	}
	var wanted []arm
	for _, p := range strings.Split(*armsFlag, ",") {
		if a, ok := arms[strings.TrimSpace(p)]; ok {
			wanted = append(wanted, a)
		}
	}

	// Incremental CSV: every result lands on disk the moment its call returns,
	// so a mid-run crash loses nothing (the full-run is hours long).
	csvF, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "csv:", err)
		os.Exit(2)
	}
	csvW := csv.NewWriter(csvF)
	_ = csvW.Write(csvHeader())
	csvW.Flush()

	sem := make(chan struct{}, *conc)
	var mu sync.Mutex
	var judgeMu sync.Mutex // the shims swap a package-global bars provider
	var results []result
	var wg sync.WaitGroup
	for _, r := range rows {
		opts, maxLevels, scenarioCap, cfg, optErr := resolveJudgeParams(st, r.TraderID)
		if optErr != nil {
			fmt.Fprintf(os.Stderr, "row %d: judge params: %v\n", r.ID, optErr)
			continue
		}
		for _, a := range wanted {
			wg.Add(1)
			go func(r store.PlannerRejectedPrompt, a arm) {
				defer wg.Done()
				sem <- struct{}{}
				res := runArm(r, a, key, opts, maxLevels, scenarioCap, cfg)
				<-sem
				judgeMu.Lock()
				res.judgeExtra(st, cfg)
				judgeMu.Unlock()
				if res.CorePass && res.FullReject == "" {
					res.FullPass = true
				}
				res.saveRaw()
				mu.Lock()
				results = append(results, res)
				_ = csvW.Write(res.csvRow())
				csvW.Flush()
				mu.Unlock()
				fmt.Printf("row %d arm %s: pass=%v wall=%.1fs finish=%s born=%s %s\n",
					r.ID, a.label, res.Pass, res.WallS, res.Finish, res.BornDead, truncate(res.FirstReject, 80))
			}(r, a)
		}
	}
	wg.Wait()
	csvW.Flush()
	_ = csvF.Close()

	printTable(results)
}

// resolveJudgeParams re-resolves the live write-site judge inputs from the
// strategy row bound to the trader (the CTO's harness gap): AuthoringOpts
// {MinRR, EntryPolicyDefault, MinHoldMin}, maxLevels and scenarioCap.
func resolveJudgeParams(st *store.Store, traderID string) (kernel.AuthoringOpts, int, int, *store.StrategyConfig, error) {
	tr, err := st.Trader().GetByID(traderID)
	if err != nil || tr == nil {
		return kernel.AuthoringOpts{}, 0, 0, nil, fmt.Errorf("trader %s: %v", traderID, err)
	}
	strategy, err := st.Strategy().Get(tr.UserID, tr.StrategyID)
	if err != nil || strategy == nil {
		return kernel.AuthoringOpts{}, 0, 0, nil, fmt.Errorf("strategy %s: %v", tr.StrategyID, err)
	}
	var cfg store.StrategyConfig
	if err := json.Unmarshal([]byte(strategy.Config), &cfg); err != nil {
		return kernel.AuthoringOpts{}, 0, 0, nil, fmt.Errorf("strategy config: %v", err)
	}
	dp := cfg.DayPlan
	policy, _ := store.ResolveEntryPolicyDefault(dp)
	hold, _ := store.ResolveMinHoldMin(dp)
	minRR, _ := store.ResolveMinRiskReward(&cfg)
	opts := kernel.AuthoringOpts{MinRR: minRR, EntryPolicyDefault: policy, MinHoldMin: hold}
	maxLevels := kernel.DefaultMaxLevels
	if dp != nil && dp.MaxLevels > 0 {
		maxLevels = dp.MaxLevels
	}
	scenarioCap := store.DefaultScenarioCap
	if dp != nil {
		scenarioCap = dp.ScenarioCapResolved()
	}
	return opts, maxLevels, scenarioCap, &cfg, nil
}

// runArm makes ONE call and runs the offline judge on the answer.
func runArm(r store.PlannerRejectedPrompt, a arm, key string, opts kernel.AuthoringOpts, maxLevels, scenarioCap int, cfg *store.StrategyConfig) result {
	res := result{RowID: int(r.ID), Arm: a.label, Model: a.model, Cap: a.cap}
	body := map[string]any{
		"model": a.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": r.PromptText},
		},
		"temperature":      0.5,
		"max_tokens":       a.cap,
		"stream":           false,
		"thinking":         map[string]any{"type": "enabled"},
		"reasoning_effort": "max",
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", deepseekURL, bytes.NewReader(raw))
	if err != nil {
		res.Err = err.Error()
		return res
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		res.Err = err.Error()
		res.WallS = time.Since(start).Seconds()
		return res
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	res.WallS = time.Since(start).Seconds()
	if resp.StatusCode != http.StatusOK {
		res.Err = fmt.Sprintf("status %d: %.200s", resp.StatusCode, b)
		return res
	}
	var parsed apiResp
	_ = json.Unmarshal(b, &parsed)
	res.PromptTok = parsed.Usage.PromptTokens
	res.CompletionTok = parsed.Usage.CompletionTokens
	if len(parsed.Choices) == 0 {
		res.Err = "no choices in response"
		return res
	}
	ch := parsed.Choices[0]
	res.ReasoningChars = len(ch.Message.ReasoningContent)
	if ch.FinishReason != nil {
		res.Finish = *ch.FinishReason
	}
	content := ch.Message.Content
	res.cfg = cfg
	res.rawContent = content
	res.session, res.tradeDate = r.Session, r.TradeDate
	res.prompt = r.PromptText
	res.symbol = symbolFromPrompt(r.PromptText)
	res.readAt = readAtFromPrompt(r.PromptText, r.TradeDate)
	judge(&res, content, r.Facts, opts, maxLevels, scenarioCap)
	return res
}

// judge runs the attempt-1 legs of the live write-site chain that are runnable
// offline: parse (with the LIVE AuthoringOpts — the CTO's harness gap) and the
// facts validator with the stored facts (machine map nil → skipped per the
// nil-map contract). requiredBias / machine-label / FVG / breakdown legs are
// NOT runnable offline (their inputs are not persisted) — see the report.
func judge(res *result, content, factsJSON string, opts kernel.AuthoringOpts, maxLevels, scenarioCap int) {
	if content == "" {
		res.FirstReject = "empty content"
		return
	}
	d, perr := kernel.ParsePlanDocForAuthoring(content, maxLevels, scenarioCap, opts)
	if perr != nil {
		res.FirstReject = "parse: " + perr.Error()
		res.CoreReject = res.FirstReject
		return
	}
	var facts kernel.PlanFacts
	if factsJSON != "" {
		if ferr := json.Unmarshal([]byte(factsJSON), &facts); ferr != nil {
			res.FirstReject = "facts unmarshal: " + ferr.Error()
			return
		}
	}
	if verr := kernel.ValidatePlanDocWithFactsMachine(d, facts, nil, maxLevels, scenarioCap); verr != nil {
		res.FirstReject = "facts: " + verr.Error()
		res.CoreReject = res.FirstReject
		return
	}
	res.Pass = true
	res.CorePass = true
	res.doc, res.facts = d, facts
	res.maxLevels, res.scenarioCap = maxLevels, scenarioCap
	// Per-scenario entry distance from the read price, in pts and xATR5m
	// (facts.DATR is the read-time daily ATR proxy the live gate uses).
	var pts, atrs []string
	for _, sc := range d.Scenarios {
		entry := 0.0
		if sc.Arm != nil {
			entry = sc.Arm.Entry
		}
		dist := entry - facts.Price
		pts = append(pts, fmt.Sprintf("%.2f", dist))
		if facts.DATR > 0 {
			atrs = append(atrs, fmt.Sprintf("%.2f", dist/facts.DATR))
		} else {
			atrs = append(atrs, "n/a")
		}
	}
	res.EntryDistsPts = strings.Join(pts, ";")
	res.EntryDistsATR = strings.Join(atrs, ";")
}

func csvHeader() []string {
	return []string{"row_id", "arm", "model", "cap", "wall_s", "ttfb_s", "prompt_tok", "completion_tok", "reasoning_chars", "finish", "core_pass", "core_reject", "full_pass", "full_reject", "entry_dists_pts", "entry_dists_xatr5m", "born_dead", "err"}
}

func (r result) csvRow() []string {
	return []string{
		strconv.Itoa(r.RowID), r.Arm, r.Model, strconv.Itoa(r.Cap),
		fmt.Sprintf("%.2f", r.WallS), fmt.Sprintf("%.2f", r.TTFBS),
		strconv.Itoa(r.PromptTok), strconv.Itoa(r.CompletionTok), strconv.Itoa(r.ReasoningChars),
		r.Finish, strconv.FormatBool(r.CorePass), r.CoreReject, strconv.FormatBool(r.FullPass), r.FullReject, r.EntryDistsPts, r.EntryDistsATR, r.BornDead, r.Err,
	}
}

// saveRaw persists the call's RAW answer (plan JSON + finish + usage) to
// ab_raw/<row>-<arm>.json so the outputs can be re-judged offline with zero
// new API calls (CTO 03:50Z).
func (r result) saveRaw() {
	if err := os.MkdirAll("ab_raw", 0o755); err != nil {
		return
	}
	rec := map[string]any{
		"row_id": r.RowID, "arm": r.Arm, "model": r.Model, "cap": r.Cap,
		"wall_s": r.WallS, "finish": r.Finish,
		"prompt_tok": r.PromptTok, "completion_tok": r.CompletionTok,
		"reasoning_chars": r.ReasoningChars,
		"core_pass":       r.CorePass, "core_reject": r.CoreReject,
		"full_pass": r.FullPass, "full_reject": r.FullReject,
		"born_dead": r.BornDead, "content": r.rawContent,
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	_ = os.WriteFile(fmt.Sprintf("ab_raw/%d-%s.json", r.RowID, r.Arm), b, 0o644)
}

func printTable(results []result) {
	fmt.Println("\narm | n | pass | cap_hit | median_wall | p90_wall | completion_tok_med")
	byArm := map[string][]result{}
	for _, r := range results {
		byArm[r.Arm] = append(byArm[r.Arm], r)
	}
	keys := make([]string, 0, len(byArm))
	for k := range byArm {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		rs := byArm[k]
		sort.Slice(rs, func(i, j int) bool { return rs[i].WallS < rs[j].WallS })
		pass, capHit := 0, 0
		for _, r := range rs {
			if r.Pass {
				pass++
			}
			if r.Finish == "length" {
				capHit++
			}
		}
		med := rs[len(rs)/2].WallS
		p90 := rs[int(0.9*float64(len(rs)-1))].WallS
		fmt.Printf("%-16s | %d | %d/%d | %d/%d | %.0fs | %.0fs | %d\n",
			k, len(rs), pass, len(rs), capHit, len(rs), med, p90, rs[len(rs)/2].CompletionTok)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// runRejudge (CTO 04:39Z) re-runs the core + write-time judges over the saved
// raw outputs with ZERO new API calls, and rewrites -out. It exists because a
// born-dead plan must never read as a pass: the live-run CSV had full_pass=true
// beside born_dead=yes, and the fix must be applied to the SAVED rows too.
func runRejudge(rawDir, out string) {
	dir, err := os.ReadDir(rawDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "raw dir:", err)
		os.Exit(2)
	}
	st, err := store.New("ab.db")
	if err != nil {
		fmt.Fprintln(os.Stderr, "store:", err)
		os.Exit(2)
	}
	defer st.Close()
	corpus, err := st.PlannerRejected().CorpusRows("2026-09-10", 1, 100000)
	if err != nil {
		fmt.Fprintln(os.Stderr, "corpus:", err)
		os.Exit(2)
	}
	byID := map[int]store.PlannerRejectedPrompt{}
	for _, r := range corpus {
		byID[int(r.ID)] = r
	}

	csvF, err := os.Create(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "csv:", err)
		os.Exit(2)
	}
	csvW := csv.NewWriter(csvF)
	_ = csvW.Write(csvHeader())

	names := []string{}
	for _, e := range dir {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var results []result
	for _, name := range names {
		b, err := os.ReadFile(rawDir + "/" + name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal(b, &rec); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			continue
		}
		rowID := int(rec["row_id"].(float64))
		row, ok := byID[rowID]
		if !ok {
			fmt.Fprintf(os.Stderr, "%s: row %d not in corpus\n", name, rowID)
			continue
		}
		opts, maxLevels, scenarioCap, cfg, optErr := resolveJudgeParams(st, row.TraderID)
		if optErr != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, optErr)
			continue
		}
		res := result{
			RowID: rowID,
			Arm:   rec["arm"].(string), Model: rec["model"].(string),
			Cap: int(rec["cap"].(float64)), WallS: rec["wall_s"].(float64),
			Finish: rec["finish"].(string),
		}
		if v, ok := rec["prompt_tok"].(float64); ok {
			res.PromptTok = int(v)
		}
		if v, ok := rec["completion_tok"].(float64); ok {
			res.CompletionTok = int(v)
		}
		if v, ok := rec["reasoning_chars"].(float64); ok {
			res.ReasoningChars = int(v)
		}
		res.rawContent, _ = rec["content"].(string)
		res.cfg = cfg
		res.session, res.tradeDate = row.Session, row.TradeDate
		res.prompt = row.PromptText
		res.symbol = symbolFromPrompt(row.PromptText)
		res.readAt = readAtFromPrompt(row.PromptText, row.TradeDate)
		judge(&res, res.rawContent, row.Facts, opts, maxLevels, scenarioCap)
		res.judgeExtra(st, cfg)
		if res.CorePass && res.FullReject == "" {
			res.FullPass = true
		}
		results = append(results, res)
		_ = csvW.Write(res.csvRow())
		fmt.Printf("rejudge %s: core=%v full=%v born=%s %s\n", name, res.CorePass, res.FullPass, res.BornDead, truncate(res.FullReject, 80))
	}
	csvW.Flush()
	_ = csvF.Close()
	printTable(results)
}

var clockRE = regexp.MustCompile(`clock (\d{1,2}:\d{2}) CT`)
var symbolRE = regexp.MustCompile(`CME (\w+) futures`)

// judgeExtra runs the write-time legs that need the DB copy's bars and the
// research shims: ATR5m from the copy's 1m tape, the REAL write-time
// feasibility + zone verdicts, and the REAL born-dead check with the publish
// clock = read time + this call's wall time. Caller holds judgeMu (the shims
// swap the package-global bars provider).
func (res *result) judgeExtra(st *store.Store, cfg *store.StrategyConfig) {
	if !res.Pass || res.doc == nil {
		res.BornDead = "na (never reached the write-time legs)"
		return
	}
	readAt := res.readAt
	if readAt.IsZero() {
		res.BornDead = "na (read clock not recoverable from the prompt)"
	}
	bars := fetchBars(st, res.symbol, readAt, res.WallS)
	atr5m := trader.ResearchArmSeamATR5mFromBars(bars)
	if atr5m > 0 {
		if feas := trader.ResearchWriteTimeFeasibilityVerdicts(res.symbol, res.doc, atr5m, cfg, res.session); len(feas) > 0 {
			res.Pass = false
			res.FirstReject = "write-time feasibility: " + feas[0]
			res.FullReject = res.FirstReject
		}
	}
	if res.Pass && !readAt.IsZero() && len(bars) > 0 {
		res.FullPass = true // provisional: cleared by either refusal below
		publish := readAt.Add(time.Duration(res.WallS * float64(time.Second)))
		_, berr := trader.ResearchValidateAuthoredScenariosAt(st, res.symbol, res.doc, res.session, res.tradeDate, readAt, publish, bars)
		if berr != nil {
			res.Pass = false
			res.FullPass = false // CTO 04:39Z — a born-dead plan can never read as a pass
			res.FullReject = "born-dead/flip-met: " + berr.Error()
			if res.FirstReject == "" {
				res.FirstReject = res.FullReject
			}
			res.BornDead = "yes (" + berr.Error() + ")"
		} else {
			res.BornDead = "no"
		}
	}
}

// symbolFromPrompt reads the instrument from the prompt header
// ("# DAY-PLAN READER — CME MNQ futures"); empty on a shape mismatch (the
// harness then skips the bars-dependent legs and names it).
func symbolFromPrompt(prompt string) string {
	if m := symbolRE.FindStringSubmatch(prompt); m != nil {
		return m[1]
	}
	return ""
}

// readAtFromPrompt recovers the read clock: the stored facts snapshot cannot
// hold ReadAt (json:"-"), so it is parsed from the prompt's own clock line
// ("clock 16:33 CT (21:33 UTC)") plus the row's trade date. Zero on any shape
// mismatch — the bars-dependent legs are skipped and the row says so.
func readAtFromPrompt(prompt, tradeDate string) time.Time {
	m := clockRE.FindStringSubmatch(prompt)
	if m == nil {
		return time.Time{}
	}
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		return time.Time{}
	}
	t, err := time.ParseInLocation("2006-01-02 15:04", tradeDate+" "+m[1], loc)
	if err != nil {
		return time.Time{}
	}
	return t
}

// fetchBars reads 1m bars for the symbol from the DB COPY around the read
// window (30m lookback for the ATR + the call's wall time after the read).
func fetchBars(st *store.Store, symbol string, readAt time.Time, wallS float64) []market.Kline {
	if st == nil || readAt.IsZero() {
		return nil
	}
	from := readAt.Add(-30 * time.Minute).UnixMilli()
	to := readAt.Add(time.Duration(wallS*float64(time.Second)) + 5*time.Minute).UnixMilli()
	rows, err := st.GormDB().Raw(
		"SELECT open_time_ms, o, h, l, c FROM bars WHERE symbol = ? AND tf = '1m' AND open_time_ms BETWEEN ? AND ? ORDER BY open_time_ms",
		symbol, from, to).Rows()
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []market.Kline
	for rows.Next() {
		var ms int64
		var o, h, l, c float64
		if rows.Scan(&ms, &o, &h, &l, &c) == nil {
			out = append(out, market.Kline{OpenTime: ms, CloseTime: ms + 59_999, Open: o, High: h, Low: l, Close: c})
		}
	}
	return out
}
