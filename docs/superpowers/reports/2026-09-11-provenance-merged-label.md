# Provenance validator × W3's merged label — one small wave (2026-09-11)

**Branch** `fix/provenance-merged-label` · lane `provenance-2bdef526/nofx-59[3f0200]` · claim `4f6a1e15` · base dev `f5034570` · read-only findings first, one-line fix second.

## The check and the render, quoted

**The validator** — `kernel.MislabeledStructuralLevels` (`kernel/plan_doc.go:813`), called at every plan write (`trader/auto_trader_planner.go:1667`):

```go
mp := structuralPrefix(ml)      // machine table's label
lp := structuralPrefix(l.Label) // the plan level's label
if (mp != "" || lp != "") && mp != lp { … "%.2f labeled %q but the machine table says %q" }
```
with, before this wave,
```go
func structuralPrefix(label string) string {
	l := strings.TrimSpace(label)
	if i := strings.Index(l, "·"); i > 0 {
		l = l[:i]            // "RTH-H · EQL·4h · EQH·15m" → "RTH-H " — trailing space
	}
	if structuralLabels[l] { return l }
	return ""
}
```

**W3's render** — `MapCandidate.NamesLine` (`kernel/map_candidates.go:94-99`): `strings.Join(c.Names, " · ")`, rendered into the model's map block at `renderMapBlock` (`:353`, `names := c.NamesLine()`; `[merged x%d]` appended when merged). "Names carries EVERY merged reference's label, strongest first" — the PRIMARY is the first component.

**The defect:** the render separates with `" · "` (space, dot, space); the validator split at `"·"` and did not trim, so the primary read as `"RTH-H "`, which is not in `structuralLabels`, so `lp=""` while `mp="RTH-H"` → mismatch. Every merged label on a structural row was rejected. Live: LONDON and NY 2026-09-11 attempt 1 both rejected on exactly this (`29275.25 labeled "RTH-H · EQL·4h · EQH·15m" but the machine table says "RTH-H"; 29038.00 labeled "PDL · SWG-L·15m · EQL·1h · EQL·15m" but the machine table says "PDL"`), each costing a repair round. Class 97's shape: two readers of one label with separator conventions that happened to differ.

## The fix (one line) and the pins

`l = strings.TrimSpace(l[:i])`. Pins (`kernel/plan_doc_provenance_merged_test.go`), RED first — the RED reproduced the live rejection text verbatim:
- `TestProvenanceAcceptsMergedLabelWithMatchingPrimary` — the two live labels pass against `RTH-H` / `PDL`.
- `TestProvenanceStillRejectsWrongPrimary` — `"EQH·15m · RTH-H"` against `RTH-H` (a structural anchor demoted behind another) still fails; LONDON v1's phantom (`"PDH"` over a zone row) still fails.
- `TestStructuralPrefixTrimsTheMergedSeparator` — the token table.
- Mutation via `scripts/mutate.sh` (the trim removed) → KILLED. `go test ./kernel ./trader` green.

## Finding for the owner, not a fix: the deadline has no room for a second slow attempt

NY 2026-09-11, attempt 1: `ai_call model=deepseek-v4-pro duration_ms=892488 finish_reason=stop ok=true http_status=200 completion=48001 reasoning_chars=155731` — **892.5 s of the 1200 s `AI_PLAN_TOTAL_DEADLINE_SECS`** (`stream idle=30s total=1200s`). It was then rejected on the label above and the repair (45.1 s) landed v1 at 08:16:32, 15 m 42 s after the read began. One more attempt of that length and the read has no repair round left inside the deadline; the deadline is per call, so the READ can run 2× or 3× 1200 s across attempts, but the session's usable window does not. LONDON's attempt 1 the same morning ran 682 s. Reasoning at `max` with 110–156k reasoning characters is the driver; nothing timed out, nothing was refused. Recorded here for the owner's ruling on the reasoning cap or the deadline; not changed.

## Rollback
Revert the one line; the pins go RED again.
