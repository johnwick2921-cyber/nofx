#!/usr/bin/env bash
# install-tree-guard.sh — installs the tree-guard user timer. NO SUDO
# (mirrors install-clock-guard.sh; linger is already enabled for user hoang so
# user timers fire without a login session).
#
#   bash deploy/install-tree-guard.sh
#
# Verify:  systemctl --user list-timers | grep tree-guard
# Logs:    journalctl --user -u nofx-tree-guard.service -f
# Ad hoc:  bash deploy/nofx-tree-guard.sh --once
#
# INSTALLING IS NOT A TREE WRITE: the units live under ~/.config/systemd/user
# and the state file lives under ~/nofx-backups. The guard never writes into the
# tree it observes.
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UNIT_DIR="$HOME/.config/systemd/user"

mkdir -p "$UNIT_DIR"
install -m 0644 "$REPO/deploy/systemd-user/nofx-tree-guard.service" "$UNIT_DIR/"
install -m 0644 "$REPO/deploy/systemd-user/nofx-tree-guard.timer" "$UNIT_DIR/"
chmod +x "$REPO/deploy/nofx-tree-guard.sh"

systemctl --user daemon-reload
systemctl --user enable --now nofx-tree-guard.timer
# Prime one run so the state file and the first journal line exist immediately —
# an installed guard that has never run is indistinguishable from a broken one.
systemctl --user start nofx-tree-guard.service || true

echo "tree-guard installed. Next runs:"
systemctl --user list-timers --no-pager | grep -E "NEXT|tree-guard" || true
echo
echo "Last verdict:"
cat "$HOME/nofx-backups/tree-guard/state" 2>/dev/null | head -3 || echo "(no state yet)"
