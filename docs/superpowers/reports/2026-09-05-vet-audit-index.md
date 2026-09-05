# NOFX / VL Intelligent — bộ báo cáo audit theo prompt gốc

**Trạng thái: đang đối chiếu bản hoàn thiện; chưa phải biên bản kết thúc.**

## Đọc bản nào trước?

Đọc [phần 09 — đánh giá toàn hệ thống](2026-09-05-vet-09-top-ten.md) trước. Phần này giải thích phương pháp giao dịch, các điểm chưa nhất quán, ưu tiên đề xuất và những quyết định cần chủ hệ thống xem xét. Sau đó dùng mục lục để kiểm tra từng luận điểm và dữ liệu gốc.

## Phạm vi

Tôi đánh giá bot như một phương pháp giao dịch: điều kiện thị trường, mức giá, điểm vào, xác nhận, stop, mục tiêu, lý do thoát, tin tức, volume, phiên giao dịch, giới hạn rủi ro và vai trò của AI/con người. Kiểm tra mã là cách xác minh bot thực sự làm gì; nó không thay thế phần đánh giá phương pháp.

Chỉ tài liệu audit và bản sao bằng chứng được cập nhật. Không sửa mã giao dịch, cấu hình, prompt đang dùng, cơ sở dữ liệu gốc, lệnh hay dịch vụ. Những câu “tôi đề xuất” là nội dung báo cáo, chưa phải hành động đã thực hiện.

## Cách hiểu kết luận

- **Đã xác minh:** có mã nguồn hoặc truy vấn và kết quả cụ thể.
- **Suy luận:** giải thích phù hợp với bằng chứng nhưng chưa chứng minh quan hệ nhân quả.
- **Giả thuyết cần thử:** một lựa chọn để so sánh, không phải chiến lược đã có lời.
- **Không đo được từ dữ liệu hiện có:** báo cáo nêu rõ thiếu gì; không thay số thiếu bằng số giả định rồi gọi là kết quả thực.

Mẫu hiệu suất hợp lệ theo yêu cầu là 58 giao dịch trong 12 ngày phiên CME, P&L đã hiệu chỉnh −$466.43. Không đủ căn cứ kết luận có lợi thế. Mẫu này cũng không chứng minh chính sách strict hiện tại có lời hay thua, vì chưa có giao dịch ghi nhận sau mốc thực thi strict đã đối chiếu. Bản hoàn thiện sửa cách dùng nhầm mẫu 65 giao dịch ở các báo cáo trước.

## Mục lục mười phần

1. [Phương pháp giao dịch và lợi thế](2026-09-05-vet-01-way-it-trades.md)
2. [Các mức giá và cách chọn mức](2026-09-05-vet-02-levels.md)
3. [Quyết định: AI, bộ máy và con người](2026-09-05-vet-03-decisions.md)
4. [Theo dõi trong phiên và tổng kết ngày](2026-09-05-vet-04-monitoring.md)
5. [Khớp lệnh, stop, mục tiêu và thoát lệnh](2026-09-05-vet-05-execution.md)
6. [Rủi ro, chuỗi thua và điều kiện tăng quy mô](2026-09-05-vet-06-risk.md)
7. [Chỉ dẫn cho AI và bản viết lại để xem xét](2026-09-05-vet-07-prompts.md)
8. [Diễn biến và khả năng giao dịch trong ba ngày](2026-09-05-vet-08-stretch.md)
9. [Đánh giá toàn hệ thống, mười ưu tiên và đề xuất quyết định](2026-09-05-vet-09-top-ten.md)
10. [Ý tưởng nghiên cứu và thử nghiệm một tháng](2026-09-05-vet-10-ideas.md)

## Đối chiếu 47 nhóm yêu cầu trong prompt gốc

Bảng này sẽ được chốt sau khi từng phần được kiểm tra. Một báo cáo có tồn tại không đồng nghĩa mọi phép đo đã thực hiện được; những giới hạn dữ liệu phải có câu trả lời rõ ràng.

| Yêu cầu | Nội dung cần trả lời | Báo cáo | Tình trạng |
|---|---|---|---|
| 1.1 | Describe the book as you'd brief a risk manager: what it fades, what it follows, when, with what stop and target, how often. Pull the real distribution since 2026-08-15: trades by condition · side · session · hold time · realized R (trader_positions ⋈ plans ⋈ trade_excursions). | [Phần 01](2026-09-05-vet-01-way-it-trades.md) | Đang kiểm tra |
| 1.2 | Where is the edge, if anywhere? Test each candidate: (a) fades — touch_outcomes hold by kind × ordinal; (b) breaks — the same inverted; (c) time of day — realized R by session and hour; (d) regime — realized R by machine regime label and realized- vol tercile. Wilson, n; "no verdict" below n=30. | [Phần 01](2026-09-05-vet-01-way-it-trades.md) | Đang kiểm tra |
| 1.3 | Is one contract + resting limits + 1.5×ATR5m floor + 2.0 R:R + flat 14:45 a SHAPE that can make money on NQ? What would you change in the shape, not the parameters? | [Phần 01](2026-09-05-vet-01-way-it-trades.md) | Đang kiểm tra |
| 1.4 | Reward side: plans claim median 2.55:1, book realizes 1.66:1, 3 of 36 reached 2R off the minimum stop — verify all three. Where should targets come from: the model, the level map, the MFE distribution (trade_excursions p50/p80/p95 by condition), or a fixed multiple? | [Phần 01](2026-09-05-vet-01-way-it-trades.md) | Đang kiểm tra |
| 1.5 | Three things a 30-year hand sees that the builders don't, each with the query that shows it. | [Phần 01](2026-09-05-vet-01-way-it-trades.md) | Đang kiểm tra |
| 2.1 | Inventory every seated kind (kernel/levels_*.go): definition (file:line), window, grade terms, seat frequency (candidate_ pool), touch count, hold/break with n + Wilson, ambiguous share. One table. | [Phần 02](2026-09-05-vet-02-levels.md) | Đang kiểm tra |
| 2.2 | Which kinds would you seat on NQ from experience; which does the tape say; where they disagree, which you trust and why. | [Phần 02](2026-09-05-vet-02-levels.md) | Đang kiểm tra |
| 2.3 | The grading ladders (zone size, TF, freshness, anchor decay, confluence, HTF ×1.2) are all [I]. Per term: any evidence in the store it changes outcomes (score-at-seat vs outcome; cut reasons vs what the cut levels did next)? | [Phần 02](2026-09-05-vet-02-levels.md) | Đang kiểm tra |
| 2.4 | The veteran review says retire eVWAP, VWAP±2σ, EQL, ONH, ONL, PDL as decision inputs (n=2–5) and that RTH-L breaks 68% (n=63) but is traded as a fade. Verify both. Then rule as a trader: retire / keep-measuring / flip. | [Phần 02](2026-09-05-vet-02-levels.md) | Đang kiểm tra |
| 2.5 | What's missing: volume-profile depth, prior-week structure, overnight-vs-RTH range, opening drive, time-of-day volatility profile, gap fill/no-fill? For each: [R]/[I], and what you'd need to see before seating it. | [Phần 02](2026-09-05-vet-02-levels.md) | Đang kiểm tra |
| 2.6 | Seats: 8 of a 12 cap, ranked by an [I] score. Would you seat fewer, more, or by a different rule (nearest-N, confluence- only, volume-anchored)? | [Phần 02](2026-09-05-vet-02-levels.md) | Đang kiểm tra |
| 3.1 | The division of labor: planner (LLM, one plan per session, re-plans on level-event wakes), executor (2-min LLM loop; under strict only plan arms trade), gates. Draw it from the code (file:line), then say whether you'd draw it that way. | [Phần 03](2026-09-05-vet-03-decisions.md) | Đang kiểm tra |
| 3.2 | Would you let an LLM author plans at all? If yes: what exactly should it author (direction? levels? scenarios? targets? nothing but the narrative?) and what must be mechanical? Measure: from planner_rejected_prompts + plans, the reject rate by rule, the re-author cost in seconds, the share of scenarios whose trigger ever fired (plans ⋈ bars). | [Phần 03](2026-09-05-vet-03-decisions.md) | Đang kiểm tra |
| 3.3 | Bias: both bias systems calibrated non-predictive (weekly anti-predictive; machine tree 0.48–0.54). The card now shows "AI x · tree y" as labels. Should any bias exist? What would you replace it with, and what's the evidence bar? | [Phần 03](2026-09-05-vet-03-decisions.md) | Đang kiểm tra |
| 3.4 | The gate chain, leg by leg: which legs protect money (cite a refusal that saved a loss from the store), which are engineering, which would you remove or reorder. Count refusals by leg since 09-02 with the counterfactual (the two-day audit's refusals CSVs). | [Phần 03](2026-09-05-vet-03-decisions.md) | Đang kiểm tra |
| 3.5 | Where does human judgment go back in — a morning review, a session veto, a kill switch, plan approval? Design it in one paragraph. | [Phần 03](2026-09-05-vet-03-decisions.md) | Đang kiểm tra |
| 4.1 | Inventory what the owner sees: plan card (fields), expectancy table, instruments drawer, boot lines, Guide, Telegram/alerts. Quote each surface's source (file:line) and what it reads. | [Phần 04](2026-09-05-vet-04-monitoring.md) | Đang kiểm tra |
| 4.2 | What a trader needs on screen DURING a session that isn't there: open risk in $, distance to stop/target in pts and ticks, session P&L vs the limit, arms resting with age, the broker book vs the ledger, last fill slippage, planner state (in flight / next read / wake suppressed), regime, the day's range vs ATR. Mock the layout in text. | [Phần 04](2026-09-05-vet-04-monitoring.md) | Đang kiểm tra |
| 4.3 | What's on screen that's noise or misleading (mock strings, stale labels, rates without n, "DESCRIPTIVE ONLY" rows a trader can't use). | [Phần 04](2026-09-05-vet-04-monitoring.md) | Đang kiểm tra |
| 4.4 | The 09-03 outage: 113 minutes silent, 50 minutes blind, no alert. What would your desk have had that fired? (The owner has declined a new alert channel — design within what exists.) | [Phần 04](2026-09-05-vet-04-monitoring.md) | Đang kiểm tra |
| 4.5 | End-of-day review: what one page would you read at 15:00 CT every day? Specify the queries. | [Phần 04](2026-09-05-vet-04-monitoring.md) | Đang kiểm tra |
| 5.1 | Trace one arm end to end with row ids: authored → composed (0B stop) → gated → placed (limit/stop-entry) → order_update → fill → materialized → excursions → exit. Quote each hop. | [Phần 05](2026-09-05-vet-05-execution.md) | Đang kiểm tra |
| 5.2 | Fill quality: for every fill, entry vs the bar's range at fill time; count fills at the bar's extreme; the AddOn's slippageTicks (never read by Go — verify) — what does the NT8 log say the slippage was? What does SIM hide vs live (queue, partial fills, stop-market gaps, the 14:45 flat into thin tape)? | [Phần 05](2026-09-05-vet-05-execution.md) | Đang kiểm tra |
| 5.3 | The marketable guard: 5 of 15 arms cancelled "accepted through — never placed", one over 1.7 pts, 3 filled; verify. Trader's ruling: cancel / bounded market entry within N ticks / stop- limit with a cap — and the N you'd set on MNQ, with the reasoning. | [Phần 05](2026-09-05-vet-05-execution.md) | Đang kiểm tra |
| 5.4 | The funnel: authored → armed → placed → reached → filled → won. Build the table since 09-02. Where does it leak most? | [Phần 05](2026-09-05-vet-05-execution.md) | Đang kiểm tra |
| 5.5 | Exits: stop composition (anchor + 1.5×ATR5m), BE/trail suspended, EOD flat. From trade_excursions: MAE of winners vs losers (p50/p80/p95), stop-hit share, target-hit share. Is the floor inside or outside what winners need? What exit design would you run, and what evidence gate before switching? | [Phần 05](2026-09-05-vet-05-execution.md) | Đang kiểm tra |
| 5.6 | Live-money breakpoints: the day this runs real money, what breaks first? Order type, size, the AddOn, reconnects, the broker snapshot cadence, the daily-limit leg (blocks entries, doesn't flatten). | [Phần 05](2026-09-05-vet-05-execution.md) | Đang kiểm tra |
| 6.1 | Re-run ~/nofx-analysis/mc-drawdown against the current store (by session-day, not trade); report expectancy CI, drawdown quantiles at 20/50/100 trades, streak probabilities, effective sample in DAYS. | [Phần 06](2026-09-05-vet-06-risk.md) | Đang kiểm tra |
| 6.2 | The Stage-A rule (1 contract until n≥30 with a positive lower- CI expectancy): the veteran calls it the best rule in the checklist. Agree? What's the honest go/no-go for live money, in numbers (days, n, lower CI, max DD, worst day)? | [Phần 06](2026-09-05-vet-06-risk.md) | Đang kiểm tra |
| 6.3 | Daily limit $450 (now wired, owner turning it on), no trade cap (owner-ruled): design the rest of a professional's risk framework at this size — per-trade risk, daily, weekly, kill switch, size ladder, what resets it. | [Phần 06](2026-09-05-vet-06-risk.md) | Đang kiểm tra |
| 6.4 | Sizing: when, if ever, does this go to 2 contracts, and what must be true? Write the ladder. | [Phần 06](2026-09-05-vet-06-risk.md) | Đang kiểm tra |
| 6.5 | Correlation and regime: one instrument, one strategy — what kills it (a regime it's never seen; a vol shift; a roll week; FOMC). What would you refuse to trade through, and how would the system know? | [Phần 06](2026-09-05-vet-06-risk.md) | Đang kiểm tra |
| 7.1 | Render the planner prompt and the executor prompt at a real read (planner_read_facts / planner_rejected_prompts). Measure: tokens, share facts vs instruction, MUST/NEVER/SHOULD counts, sentences per paragraph, the Rules paragraph length. | [Phần 07](2026-09-05-vet-07-prompts.md) | Đang kiểm tra |
| 7.2 | Read both as a junior's checklist: what's asked that no one could answer, what contradicts itself, what's missing, what you'd delete. Quote every line you'd cut. | [Phần 07](2026-09-05-vet-07-prompts.md) | Đang kiểm tra |
| 7.3 | Rejects by rule (last 7 days, n): which prompt sentences cause them; the class-45 pattern (a rule the validator knows and the prompt withholds) — any instances left? | [Phần 07](2026-09-05-vet-07-prompts.md) | Đang kiểm tra |
| 7.4 | Rewrite the planner prompt's instruction half to what YOU would give a junior: same information, half the length, numbered, no ALL-CAPS. Deliver it as an appendix (not applied). | [Phần 07](2026-09-05-vet-07-prompts.md) | Đang kiểm tra |
| 7.5 | The laws (AUDIT-CHECKLIST PART 2–3): which protect P&L, which are hygiene, which are missing from a trader's view. | [Phần 07](2026-09-05-vet-07-prompts.md) | Đang kiểm tra |
| 7.6 | Settings (/api/config/resolved): which knobs a trader would turn daily/weekly/never; which shouldn't exist. | [Phần 07](2026-09-05-vet-07-prompts.md) | Đang kiểm tra |
| 8.1 | Re-derive 09-02 → 09-04 from the store and bars yourself: every trade, every refusal, every arm, every plan version, the tape per session. Agree or disagree with the two-day audit's findings (its 5/55/30 split has no denominator — don't quote it; count in one currency: opportunities). | [Phần 08](2026-09-05-vet-08-stretch.md) | Đang kiểm tra |
| 8.2 | The trader's question: 10:00 CT on 09-03, +483-pt day, long bias, empty book 09:20–11:58 while price ran +199. What would you have done, and what would the system have needed to do it? | [Phần 08](2026-09-05-vet-08-stretch.md) | Đang kiểm tra |
| 8.3 | Replay those three days under the CURRENT rules (strict, reclaim stop-entry, arms-follow-bias warn, daily-limit leg, reaper on snapshot): count what would have been armed, filled, refused, and the P&L in points. Ids. | [Phần 08](2026-09-05-vet-08-stretch.md) | Đang kiểm tra |
| 8.4 | The outage and the blind boot: what would your desk's runbook have been at 12:30 when the log went silent? | [Phần 08](2026-09-05-vet-08-stretch.md) | Đang kiểm tra |
| 9.1 | Read final sections 1–8 on dev; ten ranked recommendations with what, why evidence label, implementation category, metric; at least two removals; checklist cross-check. | [Phần 09](2026-09-05-vet-09-top-ten.md) | Đang kiểm tra |
| 9.2 | Owner yes/no rulings with recommendation and reason; marketable tick cap, six level kinds, FVG block, Rules paragraph, wake blackout, acceptance/hold armability, target ownership. | [Phần 09](2026-09-05-vet-09-top-ten.md) | Đang kiểm tra |
| 10.1 | What other bot traders you know do that this doesn't: entry timing, regime filters, volatility sizing, session selection, order types, review loops, how they use an LLM. Each labelled [R]/[I], each with what it would take here. | [Phần 10](2026-09-05-vet-10-ideas.md) | Đang kiểm tra |
| 10.2 | What this system does that you have never seen work. Name it, say why, say what would change your mind. | [Phần 10](2026-09-05-vet-10-ideas.md) | Đang kiểm tra |
| 10.3 | If you had this codebase and one month: the single project you'd run, its evidence gate, and its kill criterion. | [Phần 10](2026-09-05-vet-10-ideas.md) | Đang kiểm tra |

## Biên bản hoàn tất

Chưa chốt. Bản cuối sẽ có phiên bản báo cáo, kiểm tra phạm vi chỉ tài liệu, kết quả health và giới hạn còn lại.
