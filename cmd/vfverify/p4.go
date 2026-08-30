package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"nofx/crypto"
	"nofx/kernel"
	"nofx/mcp"
	_ "nofx/mcp/provider" // registers deepseek
)

const plannerSystemPrompt = "You are a disciplined CME index-futures day-plan reasoner. Output ONLY the single JSON object requested — reasoning first, then the answer fields. No prose outside the JSON."

// runP4 — MANUAL PLANNER TEST (one real AI call, sandboxed).
func runP4() {
	// ── sandbox write-path proof (sandbox DB ONLY) ─────────────────────────
	sdb := openRW(sandboxDBPath)
	defer sdb.Close()
	beforeTraders := qint(sdb, `SELECT COUNT(*) FROM traders`)
	sdb.Exec(`INSERT OR IGNORE INTO traders (id,user_id,name,ai_model_id,exchange_id,strategy_id,initial_balance,is_running)
	          VALUES ('trader-1','default','vfverify-test-trader','8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek','ninjatrader','',10000,0)`)
	afterTraders := qint(sdb, `SELECT COUNT(*) FROM traders`)
	fmt.Printf("=== P4 sandbox write proof: traders %d → %d (test row 'trader-1' inserted in SANDBOX only)\n", beforeTraders, afterTraders)

	// ── decrypt the rotated DeepSeek key from the SANDBOX ai_models ────────
	var cipher string
	if err := sdb.QueryRow(`SELECT api_key FROM ai_models WHERE id='8ef641a7-815c-4bb5-9798-b070b67d7998_deepseek'`).Scan(&cipher); err != nil {
		must(err)
	}
	cs, err := crypto.NewCryptoService()
	must(err)
	key, err := cs.DecryptFromStorage(cipher)
	must(err)
	fmt.Printf("decrypted DeepSeek key: %s…%s (len %d, ENC:v1 AES-GCM, DATA_ENCRYPTION_KEY from live .env)\n",
		key[:5], key[len(key)-4:], len(key))

	// ── assemble the SAME validator context the write path uses ───────────
	ldb := openRO(liveDBPath)
	defer ldb.Close()
	now := time.Now()
	as := assemble(ldb, "ASIA", now.In(kernel.CTLocation()).Format("2006-01-02"), now, 0)
	promptBytes, err := os.ReadFile(promptPath)
	must(err)
	prompt := string(promptBytes)
	fmt.Printf("P3-rendered prompt loaded (%d bytes) — built from LIVE bars read-only\n", len(prompt))

	// ── the ONE real AI call ───────────────────────────────────────────────
	client := mcp.NewAIClientByProvider("deepseek")
	if client == nil {
		fmt.Println("FATAL: deepseek provider not registered")
		return
	}
	mcp.ApplyThinking(client, "", "")
	client.SetAPIKey(key, "", "")
	restore := mcp.ApplyMaxTokens(client, 65536)
	defer restore()
	start := time.Now()
	raw, err := client.CallWithMessages(plannerSystemPrompt, prompt)
	fmt.Printf("AI call: err=%v · elapsed %.1fs · raw len=%d\n", err, time.Since(start).Seconds(), len(raw))
	finish := mcp.LastFinishReason(client)
	fmt.Printf("finish_reason=%q (length == truncation → would retry)\n", finish)

	// ── validator chain (the write path) ───────────────────────────────────
	maxLevels, scenarioCap := as.in.MaxLevels, as.in.ScenarioCap
	sideQuota := as.sideQuota
	doc, perr := kernel.ParsePlanDocCapped(raw, maxLevels, scenarioCap)
	if perr != nil {
		fmt.Printf("ParsePlanDocCapped (schema gate): REJECTED: %v\n", perr)
	} else {
		fmt.Printf("ParsePlanDocCapped (schema gate incl. 9-condition enum): ACCEPTED (levels %d ≤ %d · scenarios %d)\n",
			len(doc.Levels), maxLevels, len(doc.Scenarios))
	}
	if doc != nil {
		if collapsed, n := kernel.CollapsePlanLevels(doc.Levels, kernel.LevelClusterTicks*0.25); n > 0 {
			doc.Levels = collapsed
			fmt.Printf("CollapsePlanLevels: merged %d near-duplicate level(s)\n", n)
		}
		if mis := kernel.MislabeledStructuralLevels(doc, as.machineLabels); len(mis) > 0 {
			fmt.Printf("MislabeledStructuralLevels: REJECT %v\n", mis)
		} else {
			fmt.Println("MislabeledStructuralLevels: clean")
		}
		thin, verr := kernel.ValidatePlanDocWithFactsMachine(doc, as.facts, as.machineLabels, sideQuota, maxLevels, scenarioCap)
		if verr != nil {
			fmt.Printf("ValidatePlanDocWithFactsMachine: REJECTED: %v\n", verr)
		} else {
			fmt.Printf("ValidatePlanDocWithFactsMachine: ACCEPTED (sideQuota=%d, thin-side notes: %v)\n", sideQuota, thin)
		}
		atr5m := as.atr5m
		for _, w := range kernel.ArmFeasibilityWarnings(doc, atr5m, 2.0, kernel.MinSLATRMult()) {
			fmt.Printf("arm feasibility WARN: %s\n", w)
		}
		if kernel.HasFvgScenario(doc) {
			origin := map[string]bool{}
			for _, lbl := range as.machineLabels {
				origin[lbl] = true
			}
			if verr := kernel.ValidateFvgEntryScenarios(doc, as.bars1m, symbol, origin, now); verr != nil {
				fmt.Printf("ValidateFvgEntryScenarios: REJECTED: %v\n", verr)
			} else {
				fmt.Println("ValidateFvgEntryScenarios: ACCEPTED")
			}
		} else {
			fmt.Println("ValidateFvgEntryScenarios: n/a (no fvg_entry scenario)")
		}
		if kernel.HasBreakdownScenario(doc) {
			if verr := kernel.ValidateBreakdownContinueScenarios(doc, as.bars1m, atr5m, as.facts.Price, now.UnixMilli()); verr != nil {
				fmt.Printf("ValidateBreakdownContinueScenarios: REJECTED: %v\n", verr)
			} else {
				fmt.Println("ValidateBreakdownContinueScenarios: ACCEPTED")
			}
		} else {
			fmt.Println("ValidateBreakdownContinueScenarios: n/a (no waterfall scenario)")
		}
		for _, m := range kernel.RoleMismatches(doc) {
			fmt.Printf("role mismatch WARN: %s\n", m)
		}
		for _, m := range kernel.ChainWarnings(*doc) {
			fmt.Printf("chain WARN: %s\n", m)
		}
		for _, m := range kernel.FantasyTargetWarnings(*doc) {
			fmt.Printf("fantasy-target WARN: %s\n", m)
		}

		// bias-tree branch + candle citations
		if strings.Contains(strings.ToLower(doc.Reasoning), "bias-tree") {
			fmt.Println("bias-tree branch: CITED ✓ — " + firstWords(doc.Reasoning, 200))
		} else {
			fmt.Println("bias-tree branch: NOT CITED ✗ (contract violation)")
		}
		cited := 0
		for _, s := range doc.Scenarios {
			for _, line := range strings.Split(s.Trigger+"\n"+s.Invalid, "\n") {
				if strings.Contains(strings.ToLower(line), "per candles") {
					cited++
				}
			}
		}
		fmt.Printf("\"per candles\" citation present in %d scenario line(s) — %s\n",
			cited, map[bool]string{true: "the eyes work ✓", false: "ABSENT (honest report)"}[cited > 0])

		// scenario table + arm math
		fmt.Println("\n=== SCENARIO LIST (the dress-rehearsal plan we never act on) ===")
		for _, s := range doc.Scenarios {
			line := fmt.Sprintf("%s %s %s", s.ID, s.Condition, s.Direction)
			if s.Arm != nil && s.Arm.Enabled {
				dist := s.Arm.Entry - s.Arm.Stop
				rr := (s.Arm.Target - s.Arm.Entry) / dist
				well := dist > 0
				if strings.EqualFold(s.Direction, "short") {
					dist = s.Arm.Stop - s.Arm.Entry
					rr = (s.Arm.Entry - s.Arm.Target) / dist
					well = dist > 0 && s.Arm.Target < s.Arm.Entry && s.Arm.Entry < s.Arm.Stop
				} else {
					well = dist > 0 && s.Arm.Stop < s.Arm.Entry && s.Arm.Entry < s.Arm.Target
				}
				minSL := kernel.MinSLATRMult() * atr5m
				feasible := well && rr+1e-9 >= 2.0 && dist+1e-9 >= minSL
				line += fmt.Sprintf(" arm[entry=%.2f stop=%.2f target=%.2f] R:R=%.2f wellFormed=%v stopDist=%.2f (≥%.2f=1×ATR5m? %v) gateFeasible=%v",
					s.Arm.Entry, s.Arm.Stop, s.Arm.Target, rr, well, dist, minSL, dist+1e-9 >= minSL, feasible)
			} else {
				line += " no-arm (AI path)"
			}
			fmt.Println("  " + line)
		}
		b, _ := json.MarshalIndent(doc, "", "  ")
		must(os.WriteFile("/tmp/nofx-vf-p4-doc.json", b, 0o600))
		fmt.Println("\nfull parsed doc: /tmp/nofx-vf-p4-doc.json · raw response: /tmp/nofx-vf-p4-raw.txt")
		must(os.WriteFile("/tmp/nofx-vf-p4-raw.txt", []byte(raw), 0o600))
	}

	// ── WEEKLY dry-run render (NO AI, NO execution) ────────────────────────
	fmt.Println("\n=== WEEKLY dry-run render (facts from live bars, read-only) ===")
	wf := kernel.ComputeWeeklyFacts(as.bars1m, now, as.price)
	fmt.Printf("completed weeks=%d (exactly 2 expected — bars start 08-19) · thin_history=%v · price=%.2f\n",
		len(wf.Weeks), wf.ThinHistory, wf.Price)
	fmt.Println(wf.SectionsText)
	fmt.Printf("WeeklyRefSet entries: %d (draw-eligible references)\n", len(kernel.WeeklyRefSet(wf)))
}
