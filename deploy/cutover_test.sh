#!/usr/bin/env bash
# cutover_test.sh — pins for the OLD_SHA reconcile (preboot finding [23]).
#
# The way back is decided by the DISK binary's vcs.revision, but the disk binary
# is not the running build: a crashed staging leaves a never-proven file on disk
# while the old process keeps serving. The reconcile refuses unless the disk
# revision agrees with BOTH /api/health (the running process's own revision) and
# the RELEASE marker, and refuses outright when neither answers.
#
# Three fixtures, each driving the PRODUCTION script (not a re-implementation):
#   F1 all agree (disk == health == RELEASE)      -> proceeds, dry run completes
#   F2 disk != health                             -> refuses, names the mismatch
#   F3 neither health nor RELEASE answers         -> refuses, cannot reconcile
set -uo pipefail

CUTOVER_SH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/cutover.sh"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PASS=0; FAIL=0
ok()   { PASS=$((PASS+1)); printf '  ok   %s\n' "$1"; }
bad()  { FAIL=$((FAIL+1)); printf '  FAIL %s\n     %s\n' "$1" "${2:-}"; }
check(){ if [ "$2" = "$3" ]; then ok "$1"; else bad "$1" "want [$3] got [$2]"; fi; }
has()  { case "$2" in *"$3"*) ok "$1";; *) bad "$1" "missing [$3] in: $2";; esac; }
hasnt(){ case "$2" in *"$3"*) bad "$1" "found [$3] in: $2";; *) ok "$1";; esac; }

if [ ! -f "$CUTOVER_SH" ]; then
  printf 'FAIL: %s does not exist — every pin below would be vacuous\n' "$CUTOVER_SH"
  exit 1
fi

WORK="$(mktemp -d)"
SRV_PID=""
trap '[ -n "$SRV_PID" ] && kill "$SRV_PID" 2>/dev/null; rm -rf "$WORK"' EXIT

# --- a stamped binary: go only stamps vcs.revision in a PLAIN checkout (it
# refuses in a git worktree, where .git is a file — proven by control
# experiment), so build the fixture in a throwaway git repo. The stamp VALUE is
# arbitrary: every fixture reads the same binary's sha back out, so the three
# legs are self-consistent.
STAMP_REPO="$WORK/stamprepo"
mkdir -p "$STAMP_REPO/m"
( cd "$STAMP_REPO" && git init -q
  cd "$STAMP_REPO/m"
  go mod init fixture.test/stamp >/dev/null 2>&1
  printf 'package main\nfunc main() {}\n' > main.go
  git add -A
  git -c user.email=fixture@test -c user.name=fixture commit -qm stamp )
BIN="$WORK/nofx-bin"
( cd "$STAMP_REPO/m" && go build -o "$BIN" . ) \
  || { echo "FAIL: cannot build the stamped fixture binary"; exit 1; }
SHA="$(go version -m "$BIN" 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i ~ /^vcs.revision=/){sub(/^vcs.revision=/,"",$i); print $i; exit}}')"
[ -n "$SHA" ] || { echo "FAIL: fixture binary carries no vcs.revision"; exit 1; }
SHORT="${SHA:0:12}"
OTHER="$(printf 'd%s' "$SHA" | cut -c1-40)"   # same length, different value

mkdir -p "$WORK/inst"
cp "$BIN" "$WORK/inst/nofx-bin"
# the dist preflight greps the bundle for the sha being installed
mkdir -p "$WORK/inst/web/dist"
printf 'fixture bundle carrying %s\n' "$SHA" > "$WORK/inst/web/dist/index.js"
export NOFX_INSTALL="$WORK/inst"
export NOFX_CUTOVER_TOKEN="fixture-token-not-a-secret"

# --- a local HTTP server: /health serves a revision, /gate serves ready ------
# The python interpreter is launched DIRECTLY with '&' (no function wrapper):
# $! is python's pid, so the cleanup kill reaches it. A function wrapper would
# make $! a subshell that exits before the kill, leaking a live server.
SRV_PY="$WORK/server.py"
cat > "$SRV_PY" <<'PY'
import http.server, sys
health_body = sys.argv[1].encode()
class H(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith('/health'):
            body = health_body
        else:
            body = b'{"ready": true, "legs": {}}'
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *a): pass
srv = http.server.HTTPServer(('127.0.0.1', 0), H)
open(sys.argv[2], 'w').write(str(srv.server_port))
srv.serve_forever()
PY

start_server() { # $1 = health revision
  [ -n "$SRV_PID" ] && kill "$SRV_PID" 2>/dev/null && wait "$SRV_PID" 2>/dev/null
  rm -f "$WORK/port"
  PORTFILE="$WORK/port"
  python3 "$SRV_PY" "$1" "$PORTFILE" &
  SRV_PID=$!
  for _ in $(seq 1 50); do [ -f "$PORTFILE" ] && break; sleep 0.1; done
  [ -f "$PORTFILE" ] || { echo "FAIL: fixture server did not start"; exit 1; }
  PORT="$(cat "$PORTFILE")"
  export NOFX_HEALTH_URL="http://127.0.0.1:$PORT/health"
  export NOFX_GATE_URL="http://127.0.0.1:$PORT/gate"
}

run_cutover() { # prints the script's combined output, sets RC
  OUT="$(bash "$CUTOVER_SH" --dry-run "$SHA" "$BIN" 2>&1)"; RC=$?
}

echo "== F1: disk == health == RELEASE -> proceeds (dry run completes) =="
start_server "{\"revision\":\"$SHA\"}"
printf '%s\n' "$SHA" > "$WORK/inst/RELEASE"
run_cutover
check "F1 rc is 0"        "$RC" "0"
has   "F1 prints reconciled" "$OUT" "current reconciled"
has   "F1 dry run completes" "$OUT" "dry run complete"
hasnt "F1 no refusal"        "$OUT" "refusing"

echo "== F2: disk != health -> refuses, names the mismatch =="
start_server "{\"revision\":\"$OTHER\"}"
rm -f "$WORK/inst/RELEASE"
run_cutover
check "F2 rc is nonzero"   "$([ $RC -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "F2 names the mismatch" "$OUT" "RUNNING process reports"
has   "F2 refuses the cutover" "$OUT" "refusing"
hasnt "F2 no dry-run completion" "$OUT" "dry run complete"

echo "== F3: neither health nor RELEASE answers -> refuses =="
export NOFX_HEALTH_URL="http://127.0.0.1:9/health"   # discard port: nothing listens
export NOFX_GATE_URL="http://127.0.0.1:9/gate"
rm -f "$WORK/inst/RELEASE"
run_cutover
check "F3 rc is nonzero"   "$([ $RC -ne 0 ] && echo nonzero || echo zero)" "nonzero"
has   "F3 cannot reconcile"   "$OUT" "cannot reconcile OLD_SHA=$SHORT"
has   "F3 refuses the cutover" "$OUT" "refusing"
hasnt "F3 no dry-run completion" "$OUT" "dry run complete"

printf '\n== %d pass / %d fail ==\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ] || exit 1
