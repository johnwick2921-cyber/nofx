# NOFX auto-start on reboot (WSL2 + Windows)

Goal: after a Windows restart + login, the full stack returns with **zero manual
steps** — WSL boots, `nofx-bin` + the frontend run as services (auto-restart on
crash), NT8 relaunches and the AddOn reconnects (it retries every 5s —
`VLTraderTCPClient.cs` `RECONNECT_INTERVAL_MS`).

Status on this machine: WSL distro **Ubuntu-24.04**, systemd **already enabled**
(`/etc/wsl.conf [boot] systemd=true`) — no `wsl --shutdown` needed.
`NT_TRANSPORT=tcp` was moved into `.env` (it lived only in `~/.bashrc`, which
services never read — that would have silently fallen back to the CSV bridge).

## 1. WSL side — one command (needs sudo)

```bash
cd /home/hoang/nofx && sudo bash deploy/install-autostart.sh
```

Installs + enables two units (templates in `deploy/`):

| Unit | What | Log |
|---|---|---|
| `nofx.service` | `./nofx-bin` from `/home/hoang/nofx` (reads `.env`, `data/data.db`) | `/tmp/backend.log` (+ `journalctl -u nofx` survives reboots) |
| `nofx-web.service` | `npm run dev` in `web/` (vite :3000, /api → :8080) | `/tmp/frontend.log` |

Both: `Restart=on-failure`, `RestartSec=5`.

Verify after install:
```bash
systemctl status nofx nofx-web --no-pager | grep Active
ss -tlnp | grep -E ':(8080|3000|36974)'    # 36974 binds once the NT trader loads
# crash-restart proof:
sudo kill -9 $(pgrep -x nofx-bin); sleep 6; pgrep -x nofx-bin && echo RESTARTED
```

## 2. Windows side — Task Scheduler (boots WSL at login)

1. Start → "Task Scheduler" → **Create Task** (not Basic):
   - **Name:** `Start nofx WSL`
   - **General:** Run only when user is logged on.
   - **Triggers:** New → *At log on* → Specific user (you) → ✅ *Delay task for:* `30 seconds`.
   - **Actions:** New → Program: `C:\Windows\System32\wsl.exe`
     Arguments: `-d Ubuntu-24.04 --exec /bin/true`
   - **Conditions:** untick "Start the task only if the computer is on AC power" (laptops).
2. That command boots the WSL VM; systemd then auto-starts `nofx` + `nofx-web`,
   and the running services keep the VM alive.

## 3. Windows side — NT8 auto-start + auto-connect

1. `Win+R` → `shell:startup` → Enter → copy a **NinjaTrader 8 shortcut** into
   that folder (NT8 then launches at every login; the AddOn loads with it and
   retries the backend every 5s until connected).
2. Inside NT8: **Tools → Options → General →** set **"On startup, connect to:"**
   = your SIM/data connection — so the feed reconnects without clicks.

## 4. Optional: zero-login boot (headless after power loss)

Everything above fires **at login**. For unattended recovery (power returns →
machine boots → stack up with nobody at the keyboard) enable Windows
auto-login: `Win+R` → `netplwiz` → untick "Users must enter a user name and
password…" → enter your password once. **Security caveat:** anyone who powers
on the PC is logged in as you. Default recommendation: leave login required.

## 5. Full-reboot test (the real proof)

Restart Windows → log in → wait ~1 minute → check:
- `wsl -d Ubuntu-24.04 -- systemctl is-active nofx nofx-web` → both `active`
- `:36974` listening; NT8 open; AddOn log shows CONNECTED (no "actively refused" loop)
- bars flowing (market open) or clean idle (closed); UI at `http://localhost:3000`

SIM-only: nothing here touches trading code; the live-account block is
unchanged. Rollback: `sudo systemctl disable --now nofx nofx-web` and launch
manually as before.
