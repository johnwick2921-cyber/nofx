#!/usr/bin/env bash
# crypto-union-gate.sh — the C13 union gate (plan v10 FINAL; DS-102 owns it).
#
# READ-ONLY over the repo. Writes only a /tmp scratch listing for the
# enumeration (mktemp + trap), never the tree, the DB or the box.
#
# Usage:  scripts/crypto-union-gate.sh <table-A> <table-B> <table-C>
#
# CANONICAL TABLE FORMAT (CTO ruling, C13 table review 2026-10-01 — CR-B's
# markdown pipe row is canonical; one row per hit LINE):
#   branch-point: <40-hex sha>          # the part's branch point
#   integrator-tip: <40-hex sha>        # the integrator tip it was generated against
#   paths: <space-separated pathspecs>  # the part's swept paths
#   regex: <the assembled literal>      # MUST equal the guard's export, byte-for-byte
#   | path | line | token | DELETE|CUT|KEEP | OWNER | reason |
#   # a table with zero rows writes an explicit line instead:
#   0 hits
#   # OWNER names the part that owns the line (CR-A|CR-B|CR-C). A table MAY
#   # list a line it does not own as a CEDED row (OWNER = the owning part).
#   # EXACTLY ONE owner per hit line across the union — a double-claim FAILS.
#
# Gate rules (C13 invariants 1-5 + the CTO's content-asserts):
#   (1) every table exists and parses: >=1 row or the explicit "0 hits" line;
#       any line that is not a header/comment/canonical row is a FAIL.
#   (2) two shas: `git merge-base --is-ancestor <branch-point> HEAD` holds AND
#       `git diff --name-only <integrator-tip>..HEAD -- <paths>` is empty.
#   (3) git failure = FAIL, never a skip.
#   (4) canary: every KEEP row's token is re-found by the sweep in its file;
#       if the union holds NO KEEP rows, the guard test's sentinels are the
#       canary (printed as n/a here).
#   (5) ONE regex: each table's `regex:` header equals the guard's exported
#       literal (extracted programmatically from branding/no_crypto.go), byte
#       for byte — a hand-typed copy fails.
#   Sweep: git ls-files -z, >=2000-file floor, GNU grep -E -i -I -n (binary
#   files skipped). Every hit line must have rows in the union, EXACTLY ONE
#   distinct owner across the union, and be covered by a KEEP row whose token
#   the line contains (case-insensitive) — a DELETE/CUT-owned hit means the
#   cut has not happened at this head.
#   Content asserts (single-quoted 'mixed' sites the regex cannot see):
#       web/src/components/plan/ExecutorVerdict.tsx   arm.state === 'mixed'      -> must be PRESENT (KEEP)
#       web/src/components/trader/TraderConfigModal.tsx source_type === 'mixed'  -> if present, the union must name it (CUT); absent = cut complete
set -u

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT" || exit 3
GUARD=branding/no_crypto.go

fail=0
ok()   { echo "PASS $1"; }
bad()  { echo "FAIL $1"; fail=$((fail+1)); }

# -- regex literal, extracted PROGRAMMATICALLY from the guard (never re-typed)
lit=$(sed -n 's/^const SweepRegexLiteral = `\(.*\)`$/\1/p' "$GUARD")
if [ -z "$lit" ]; then bad "cannot extract SweepRegexLiteral from $GUARD"; echo "== $fail FAIL"; exit 1; fi
ok "sweep literal extracted from the guard ($(printf %s "$lit" | wc -c) bytes)"

# -- sweep the tracked tree (invariant 3 + the 2000 floor)
LIST=$(mktemp) || { bad "mktemp failed"; echo "== $fail FAIL"; exit 1; }
trap 'rm -f "$LIST"' EXIT
git ls-files -z > "$LIST" || { bad "git ls-files failed (never skip)"; echo "== $fail FAIL"; exit 1; }
nfiles=$(tr -cd '\0' < "$LIST" | wc -c)
if [ "$nfiles" -lt 2000 ]; then bad "enumeration floor: $nfiles tracked files < 2000"; fi
ok "enumerated $nfiles tracked files (floor 2000)"

# row index:  tbl|file|line -> disposition (owner/token kept for messages)
declare -A row_dispo row_owner row_token row_line
allrows=0; keepcount=0
for tbl in "$@"; do
  [ -f "$tbl" ] || { bad "table missing: $tbl"; continue; }
  bp=$(sed -n 's/^branch-point: *\([0-9a-f]\{40\}\)$/\1/p' "$tbl" | head -1)
  tip=$(sed -n 's/^integrator-tip: *\([0-9a-f]\{40\}\)$/\1/p' "$tbl" | head -1)
  paths=$(sed -n 's/^paths: *//p' "$tbl" | head -1)
  treg=$(sed -n 's/^regex: *//p' "$tbl" | head -1)
  [ -n "$bp" ] || bad "$tbl: header branch-point missing/not 40-hex"
  [ -n "$tip" ] || bad "$tbl: header integrator-tip missing/not 40-hex"
  [ -n "$paths" ] || bad "$tbl: header paths missing"
  if [ -n "$treg" ]; then
    if [ "$treg" = "$lit" ]; then ok "$tbl: regex header equals the guard literal byte-for-byte"; else
      bad "$tbl: regex header != guard literal (hand-typed copy?)"; fi
  else bad "$tbl: header regex missing"; fi
  if [ -n "$bp" ] && ! git merge-base --is-ancestor "$bp" HEAD 2>/dev/null; then
    bad "$tbl: HEAD does not descend from branch-point $bp"; fi
  if [ -n "$tip" ] && [ -n "$paths" ]; then
    d=$(git diff --name-only "$tip"..HEAD -- $paths 2>/dev/null)
    if [ -n "$d" ]; then bad "$tbl: stale table — diff $tip..HEAD over its paths is non-empty: $(echo "$d" | head -1)"; else
      ok "$tbl: diff $tip..HEAD over its paths is empty"; fi
  fi
  n=0
  while IFS= read -r line; do
    case "$line" in
      ""|"#"*|branch-point:*|integrator-tip:*|paths:*|regex:*|"0 hits") continue;;
      "|"*)
        # | path | line | token | DISP | OWNER | reason |  (trailing pipe optional)
        body=${line#|}; body=${body%|}
        IFS='|' read -r p ln tok disp owner reason <<EOF
$body
EOF
        p=$(echo "$p" | xargs); ln=$(echo "$ln" | xargs); tok=$(echo "$tok" | xargs)
        disp=$(echo "$disp" | xargs); owner=$(echo "$owner" | xargs)
        reason=$(echo "$reason" | xargs)
        if [ -z "$p" ] || [ -z "$ln" ] || [ -z "$tok" ] || [ -z "$disp" ] || [ -z "$owner" ]; then
          bad "$tbl: unparseable row (canonical: | path | line | token | DISP | OWNER | reason |): $line"
          continue
        fi
        case "$disp" in
          DELETE|CUT|KEEP) ;;
          *) bad "$tbl: bad disposition '$disp' in row $p:$ln"; continue;;
        esac
        k="$tbl|$p|$ln"
        if [ -n "${row_dispo[$k]:-}" ]; then
          bad "$tbl: DUPLICATE row for $p:$ln (the same table lists it twice)"
          continue
        fi
        row_dispo[$k]="$disp"; row_owner[$k]="$owner"; row_token[$k]="$tok"; row_line[$k]="$p|$ln"
        n=$((n+1)); allrows=$((allrows+1))
        [ "$disp" = "KEEP" ] && keepcount=$((keepcount+1))
        ;;
      *)
        bad "$tbl: unparseable line (not a canonical pipe row): $(echo "$line" | cut -c1-80)"
        ;;
    esac
  done < "$tbl"
  if [ "$n" -eq 0 ]; then
    grep -q '^0 hits' "$tbl" && ok "$tbl: explicit 0-hits table" || bad "$tbl: parses to 0 rows and no explicit '0 hits' line"
  else
    ok "$tbl: parses, $n rows"
  fi
done
[ "$allrows" -eq 0 ] && bad "union holds no rows and no table is an explicit 0-hits table"

# -- the sweep: every hit line has rows, ONE owner, and a KEEP row covers it
hits=0
while IFS= read -r -d '' f; do
  case "$f" in
    *_test.go) continue;;   # swept scope: Go *_test.go excluded
  esac
  out=$(grep -n -I -i -E "$lit" -- "$f" 2>/dev/null) || continue
  while IFS= read -r line; do
    ln=${line%%:*}
    hits=$((hits+1))
    k="$f|$ln"
    # collect this line's rows across ALL tables (ceded rows share one owner)
    owners=""; disp=""; tok=""
    for key in "${!row_dispo[@]}"; do
      [ "${row_line[$key]}" = "$k" ] || continue
      o=${row_owner[$key]}
      case " $owners " in *" $o "*) ;; *) owners="$owners $o";; esac
      disp=${row_dispo[$key]}; tok=${row_token[$key]}
    done
    owners=$(echo "$owners" | xargs)
    if [ -z "$owners" ]; then
      bad "UNLISTED hit: $f:$ln: $(echo "$line" | cut -c1-90)"
      continue
    fi
    nowners=$(echo "$owners" | wc -w)
    if [ "$nowners" -gt 1 ]; then
      bad "DOUBLE-CLAIM hit: $f:$ln owned by [$owners] — exactly one owner per line"
      continue
    fi
    if [ "$disp" = "KEEP" ] && printf '%s' "$line" | grep -qiF -- "$tok"; then
      continue
    fi
    bad "hit not covered by a KEEP row (owner $owners says $disp — the cut must have removed it): $f:$ln: $(echo "$line" | cut -c1-90)"
  done <<EOF
$out
EOF
done < "$LIST"
ok "sweep complete: $hits hit lines, $keepcount KEEP rows in the union"

# -- invariant (4): canary over the union KEEP rows
can=0
for k in "${!row_dispo[@]}"; do
  [ "${row_dispo[$k]}" = "KEEP" ] || continue
  p=${row_line[$k]%%|*}
  if [ -f "$p" ] && grep -q -I -i -F -- "${row_token[$k]}" "$p" 2>/dev/null; then can=$((can+1)); else
    bad "canary: KEEP row token '${row_token[$k]}' not re-found in $p — sweep or table is wrong"; fi
done
[ "$can" -gt 0 ] && ok "canary: all $can KEEP tokens re-found" || echo "NOTE canary n/a (no KEEP rows) — the guard test's sentinels are the canary"

# -- content asserts (CTO ruling 2026-10-01): single-quoted 'mixed' sites
f_keep="web/src/components/plan/ExecutorVerdict.tsx";  n_keep="arm.state === 'mixed'"
f_cut="web/src/components/trader/TraderConfigModal.tsx"; n_cut="source_type === 'mixed'"
if [ -f "$f_keep" ]; then
  if grep -qF -- "$n_keep" "$f_keep" 2>/dev/null; then
    ok "content-assert: ExecutorVerdict 'mixed' present (KEEP — its CR-C table row is the requirement)"
  else bad "content-assert: ExecutorVerdict lost its plan-state 'mixed' line (KEEP site vanished)"; fi
else bad "content-assert: $f_keep missing at HEAD"; fi
if [ -f "$f_cut" ] && grep -qF -- "$n_cut" "$f_cut" 2>/dev/null; then
  found=""
  for key in "${!row_dispo[@]}"; do
    case "${row_line[$key]}" in "$f_cut|"*) found=yes;; esac
  done
  if [ "$found" = "yes" ]; then
    ok "content-assert: TraderConfigModal 'mixed' still present but the union names it (CUT pending)"
  else bad "content-assert: TraderConfigModal 'mixed' present with NO row — a lane forgot the CUT"; fi
else
  ok "content-assert: TraderConfigModal 'mixed' gone (CUT complete)"
fi

echo "== $fail FAIL"
exit "$fail"
