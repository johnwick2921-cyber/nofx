#!/usr/bin/env bash
# PRE-REOPEN F5c (2026-08-28) — WSL2 clock-drift remediation.
#
# WSL2 guests do not get a real hypervisor clock; after host suspend/resume the
# guest clock can drift hard enough to desync NT8 bar timestamps (the Level 1
# systemd-timesyncd path is unreliable inside WSL2).
#
# Run ONCE with sudo (root timer is installed via systemd --system):
#   sudo bash deploy/fix-wsl2-clock.sh
#
# What it does:
#  1. Installs chrony (WSL2-friendly), synced to time.windows.com (the host's
#     own clock), with `makestep 1 -1` — step freely on boot, never slewing.
#  2. Adds a 10-minute root cron step-sync as a belt-and-suspenders fallback:
#     `hwclock -s --utc` re-reads the RTC (which Windows keeps warm).
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "run with sudo: sudo bash $0" >&2
  exit 1
fi

echo "[1/3] installing chrony..."
if command -v apt-get >/dev/null 2>&1; then
  apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq chrony
else
  echo "no apt-get found — install chrony manually, then re-run" >&2
  exit 1
fi

echo "[2/3] configuring chrony for WSL2..."
cat > /etc/chrony/chrony.conf <<'EOF'
# PRE-REOPEN F5c — WSL2 clock hardening.
pool time.windows.com iburst maxsources 4
# Step the clock freely if far off; never trust a drifting WSL clock.
makestep 1 -1
# Local RTC is maintained by the Windows host — use it as a second opinion.
refclock SHM 0 offset 0.0 refid RTC
driftfile /var/lib/chrony/chrony.drift
EOF

echo "[3/3] enabling + starting chrony..."
systemctl enable --now chrony || service chrony start || true
sleep 2
chronyc makestep || true
chronyc tracking || true

# Belt-and-suspenders: re-read the host-maintained RTC every 10 minutes.
CRON_LINE="*/10 * * * * root /sbin/hwclock -s --utc >/dev/null 2>&1"
if [ -f /etc/cron.d/fix-wsl2-clock ]; then
  rm -f /etc/cron.d/fix-wsl2-clock
fi
printf '%s\n' "$CRON_LINE" > /etc/cron.d/fix-wsl2-clock
chmod 644 /etc/cron.d/fix-wsl2-clock

echo
echo "done — chrony enabled + RTC step-sync every 10 min."
echo "verify with: chronyc tracking"
