# 🟢 Demo 09 — "What Just Happened on My Machine?"

**Difficulty:** Easy  
**Category:** Sysadmin / process archaeology  
**Estimated setup:** 2 minutes  
**Godshell tools exercised:** `summary`, `search`, `inspect`, `family`, `goread_shell_history`

---

## Scenario

You stepped away from your machine for 20 minutes. When you came back,
something changed: a new window appeared, a notification fired, or a
cron job ran. You have no idea what happened.

Traditional approach: dig through syslog, journalctl, .bash_history…  
Godshell approach: **"what happened while I was gone?"**

Because Godshell has been watching every `execve`, `openat`, `exit`, and
`connect` call, it can reconstruct a complete timeline of activity.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_09_process_archaeology.sh
# Simulates a burst of background system activity

set -e

echo "🏛️  Simulating background system activity..."
echo "   (These events will be visible in Godshell's ghost/exited list)"
echo ""

# 1. Simulate a system update check
(
  echo "[cron] Checking for updates..."
  cat /etc/os-release > /dev/null 2>&1
  cat /etc/apt/sources.list > /dev/null 2>&1 || true
  ls /var/cache/apt/ > /dev/null 2>&1 || true
  sleep 2
) &
wait $!
echo "[✓] Update check completed (now a ghost process)"

# 2. Simulate a log rotation
(
  echo "[logrotate] Rotating logs..."
  ls /var/log/ > /dev/null 2>&1
  cat /var/log/syslog > /dev/null 2>&1 || true
  cat /var/log/auth.log > /dev/null 2>&1 || true
  sleep 1
) &
wait $!
echo "[✓] Log rotation completed (now a ghost process)"

# 3. Simulate a backup script
BACKUP_DIR=$(mktemp -d /tmp/godshell_backup_XXXX)
(
  echo "[backup] Running nightly backup..."
  cp /etc/hostname "$BACKUP_DIR/" 2>/dev/null || true
  cp /etc/hosts "$BACKUP_DIR/" 2>/dev/null || true
  tar czf "$BACKUP_DIR/config_backup.tar.gz" -C /etc hostname hosts 2>/dev/null || true
  sleep 1
) &
wait $!
echo "[✓] Backup completed (files at $BACKUP_DIR)"

# 4. Simulate a Docker health check (even without Docker)
(
  echo "[healthcheck] Service health check..."
  curl -s --connect-timeout 1 http://127.0.0.1:80/health 2>/dev/null || true
  curl -s --connect-timeout 1 http://127.0.0.1:8080/status 2>/dev/null || true
  sleep 1
) &
wait $!
echo "[✓] Health check completed"

# 5. Leave one "suspicious" long-running process
(
  while true; do
    cat /etc/passwd > /dev/null 2>&1
    sleep 30
  done
) &
LINGERER_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  All background events have fired.                              "
echo "║  4 processes ran and exited (visible as ghosts).                "
echo "║  1 lingering process still running: PID $LINGERER_PID           "
echo "║                                                                  "
echo "║  Now ask Godshell:                                               "
echo "║                                                                  "
echo "║  > what happened on my machine in the last 5 minutes?           "
echo "║  > show me all recently exited processes                        "
echo "║  > what files did the backup script touch?                      "
echo "║  > did anything make network requests?                          "
echo "║  > is there anything still running that shouldn't be?           "
echo "║  > show my recent shell history                                 "
echo "║                                                                  "
echo "║  To clean up:  kill $LINGERER_PID; rm -rf $BACKUP_DIR           "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

| #   | Prompt                                                   | Tool                            | What It Reveals                                                  |
| --- | -------------------------------------------------------- | ------------------------------- | ---------------------------------------------------------------- |
| 1   | `what happened on this machine in the last few minutes?` | `summary`                       | Ghost list shows 4 recently exited processes + 1 active lingerer |
| 2   | `inspect the backup process`                             | `inspect(ghostPid)`             | Shows `tar`, `cp` commands and what files were copied            |
| 3   | `did anything make network requests?`                    | `search("curl\|http\|connect")` | Health check process tried port 80 and 8080                      |
| 4   | `what's this process still reading /etc/passwd?`         | `inspect(lingerer_pid)`         | Reveals the lingering suspicious process                         |
| 5   | `show my recent commands`                                | `goread_shell_history(user)`    | Shows the demo setup script was run                              |

---

## Expected Outcome

Godshell should reconstruct a **complete timeline** of what happened:

1. An update check ran and read `/etc/os-release` and apt sources
2. A log rotation script accessed `/var/log/`
3. A backup script copied config files and created a tarball
4. A health check made HTTP requests to localhost
5. A suspicious process is **still lingering** and reading `/etc/passwd`

All without `journalctl`, `syslog`, or any manual forensics.

---

## Cleanup

```bash
kill %1 2>/dev/null
rm -rf /tmp/godshell_backup_*
```
