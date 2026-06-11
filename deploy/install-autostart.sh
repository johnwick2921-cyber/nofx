#!/usr/bin/env bash
# NOFX auto-start installer (WSL side). Run ONCE as root:
#   sudo bash deploy/install-autostart.sh
#
# What it does:
#   1. Stops any manually-launched nofx-bin / vite (frees ports 8080/3000/36974).
#   2. Installs nofx.service + nofx-web.service into /etc/systemd/system.
#   3. Enables + starts both (auto-start on every WSL boot; auto-restart ≤5s on crash).
# Idempotent — safe to re-run after editing the unit files.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "→ stopping any manually-launched processes..."
pkill -x nofx-bin 2>/dev/null && sleep 2 || true
pkill -f "node /home/hoang/nofx/web/node_modules/.bin/vite" 2>/dev/null || true
pkill -f "npm run dev" 2>/dev/null || true
sleep 1

echo "→ installing systemd units..."
cp deploy/nofx.service deploy/nofx-web.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now nofx nofx-web

sleep 4
echo
echo "→ status:"
systemctl --no-pager --lines=0 status nofx | head -3
systemctl --no-pager --lines=0 status nofx-web | head -3
echo
echo "→ listeners (expect :8080 now; :36974 binds when the ninjatrader trader loads):"
ss -tlnp | grep -E ':(8080|3000|36974)' || true
echo
echo "✓ done. Crash-proof: kill the PID and watch systemd bring it back:"
echo "    pgrep -x nofx-bin; sudo kill -9 \$(pgrep -x nofx-bin); sleep 6; pgrep -x nofx-bin"
