#!/usr/bin/env bash
#
# C1 — nofx SQLite online backup + retention prune.
#
# Uses the SQLite *online backup API* (via python3 stdlib — no sqlite3 CLI needed,
# and consistent even while the bot is writing) to snapshot data/data.db, verifies
# the copy's integrity, gzips it, and prunes to a daily+weekly retention window.
#
# Layout under $NOFX_BACKUP_DIR (default ~/nofx-backups/auto):
#   daily/nofx-YYYY-MM-DD_HHMMSS.db.gz    every run   → keep newest KEEP_DAILY
#   weekly/nofx-YYYY-MM-DD_HHMMSS.db.gz   1 per ISO wk → keep newest KEEP_WEEKLY
#
# Driven by the user systemd timer nofx-backup.timer (05:00 + 17:30 CT). Safe to
# run by hand any time. Exits non-zero (and keeps nothing partial) on any failure.
set -euo pipefail

DB="${NOFX_DB:-/home/hoang/nofx/data/data.db}"
DB_RESEARCH="${NOFX_DB_RESEARCH:-/home/hoang/nofx/data/data.db.research.db}"
ROOT="${NOFX_BACKUP_DIR:-$HOME/nofx-backups/auto}"
KEEP_DAILY="${NOFX_KEEP_DAILY:-14}"
KEEP_WEEKLY="${NOFX_KEEP_WEEKLY:-8}"
# P2-5: the research ledger keeps its own knobs so the owner can decide its
# retention independently (defaults mirror the main DB).
KEEP_RESEARCH_DAILY="${NOFX_KEEP_RESEARCH_DAILY:-$KEEP_DAILY}"
KEEP_RESEARCH_WEEKLY="${NOFX_KEEP_RESEARCH_WEEKLY:-$KEEP_WEEKLY}"

DAILY_DIR="$ROOT/daily"
WEEKLY_DIR="$ROOT/weekly"
mkdir -p "$DAILY_DIR" "$WEEKLY_DIR"

ts="$(date +%Y-%m-%d_%H%M%S)"

# backup_one <src> <prefix> — online backup + quick_check on the COPY + gzip.
# Writes to a .partial file first so a crash never leaves a truncated backup
# that looks complete. Runs niced: the research DB is 200+ GB and a full-copy
# stall must not starve the bot on the same box.
backup_one() {
  local src="$1" prefix="$2" tmp final
  tmp="$DAILY_DIR/.${prefix}-${ts}.db.partial"
  final="$DAILY_DIR/${prefix}-${ts}.db.gz"

  nice -n 19 ionice -c 3 python3 - "$src" "$tmp" <<'PY'
import sqlite3, sys
src, dst = sys.argv[1], sys.argv[2]
s = sqlite3.connect(src)
d = sqlite3.connect(dst)
ok = "fail"
try:
    with d:
        s.backup(d)            # atomic, consistent snapshot via the backup API
    ok = d.execute("PRAGMA quick_check").fetchone()[0]
finally:
    d.close(); s.close()
if ok != "ok":
    sys.exit("backup integrity check FAILED: %s" % ok)
PY

  gzip -f "$tmp"                  # -> $tmp.gz
  mv -f "$tmp.gz" "$final"
  echo "nofx-backup: wrote $final ($(du -h "$final" | cut -f1))"
}

# prune <dir> <keep> <prefix> — keep the newest N (names sort lexically).
prune() {
  local dir="$1" keep="$2" prefix="$3" f n=0
  # shellcheck disable=SC2012
  for f in $(ls -1 "$dir"/${prefix}-*.db.gz 2>/dev/null | sort -r); do
    n=$((n + 1))
    if [[ $n -gt $keep ]]; then
      rm -f "$f"
      echo "nofx-backup: pruned $(basename "$f")"
    fi
  done
}

# promote_weekly <prefix> <src> — one snapshot per ISO week (first run wins).
promote_weekly() {
  local prefix="$1" src="$2" week_tag weekly
  week_tag="$(date +%G-W%V)"     # e.g. 2026-W33
  if ! ls "$WEEKLY_DIR"/${prefix}-*."${week_tag}".db.gz >/dev/null 2>&1; then
    weekly="$WEEKLY_DIR/${prefix}-${ts}.${week_tag}.db.gz"
    cp -f "$src" "$weekly"
    echo "nofx-backup: promoted weekly $weekly"
  fi
}

if [[ ! -f "$DB" ]]; then
  echo "nofx-backup: DB not found at $DB" >&2
  exit 1
fi

backup_one "$DB" "nofx"
promote_weekly "nofx" "$DAILY_DIR/nofx-${ts}.db.gz"
prune "$DAILY_DIR" "$KEEP_DAILY" "nofx"
prune "$WEEKLY_DIR" "$KEEP_WEEKLY" "nofx"

if [[ -f "$DB_RESEARCH" ]]; then
  backup_one "$DB_RESEARCH" "research"
  promote_weekly "research" "$DAILY_DIR/research-${ts}.db.gz"
  prune "$DAILY_DIR" "$KEEP_RESEARCH_DAILY" "research"
  prune "$WEEKLY_DIR" "$KEEP_RESEARCH_WEEKLY" "research"
else
  echo "nofx-backup: research DB absent at $DB_RESEARCH — skipped (a machine without the research ledger is not a failure)"
fi

echo "nofx-backup: done ($(ls -1 "$DAILY_DIR"/nofx-*.db.gz 2>/dev/null | wc -l) daily, $(ls -1 "$WEEKLY_DIR"/nofx-*.db.gz 2>/dev/null | wc -l) weekly)"
