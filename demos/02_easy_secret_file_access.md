# 🟢 Demo 02 — "Who Touched My SSH Keys?"

**Difficulty:** Easy  
**Category:** File access auditing  
**Estimated setup:** 1 minute  
**Godshell tools exercised:** `search`, `inspect`, `read_file`, `goread_shell_history`

---

## Scenario

You're a sysadmin and you want to know if any process has been reading
your SSH private keys, `.env` files, or other sensitive paths.

Traditional approach: set up `auditd`, write rules, parse logs.  
Godshell approach: just ask.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_02_secret_sniff.sh
# Simulates programs accessing sensitive files

set -e

echo "🕵️  Starting secret file access simulation..."

SECRETS_DIR=$(mktemp -d /tmp/godshell_demo_XXXX)

# 1. Create fake sensitive files
mkdir -p "$SECRETS_DIR/.ssh"
echo "-----BEGIN OPENSSH PRIVATE KEY-----" > "$SECRETS_DIR/.ssh/id_rsa"
echo "fake-key-content-do-not-use"       >> "$SECRETS_DIR/.ssh/id_rsa"
echo "-----END OPENSSH PRIVATE KEY-----" >> "$SECRETS_DIR/.ssh/id_rsa"
chmod 600 "$SECRETS_DIR/.ssh/id_rsa"

echo "DATABASE_URL=postgres://admin:supersecret@prod-db:5432/app" > "$SECRETS_DIR/.env"
echo "API_KEY=sk-live-FAKE1234567890" >> "$SECRETS_DIR/.env"

# 2. Simulate a "suspicious" script that reads these files
(
  while true; do
    cat "$SECRETS_DIR/.ssh/id_rsa" > /dev/null 2>&1
    cat "$SECRETS_DIR/.env" > /dev/null 2>&1
    sleep 5
  done
) &
SNIFF_PID=$!

# 3. Simulate a legitimate process too (e.g., ssh-agent checking keys)
(
  while true; do
    ls "$SECRETS_DIR/.ssh/" > /dev/null 2>&1
    sleep 10
  done
) &
LEGIT_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║  Sniffer running as PID: $SNIFF_PID                     "
echo "║  Legit process as PID:   $LEGIT_PID                     "
echo "║  Secrets dir: $SECRETS_DIR                               "
echo "║                                                          "
echo "║  Now ask Godshell:                                       "
echo "║                                                          "
echo "║  > who has been reading SSH keys?                        "
echo "║  > search for any process touching .env files            "
echo "║  > what's inside that .env file?                         "
echo "║  > show me the shell history for $(whoami)               "
echo "║                                                          "
echo "║  To clean up:  kill $SNIFF_PID $LEGIT_PID               "
echo "║                rm -rf $SECRETS_DIR                       "
echo "╚══════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

| #   | Prompt                                        | Expected Tool Chain                                      |
| --- | --------------------------------------------- | -------------------------------------------------------- |
| 1   | `has any process accessed SSH keys recently?` | `search("ssh\|id_rsa")` → finds the sniffer PID          |
| 2   | `who is reading .env files?`                  | `search(".env")` → identifies the sniffer                |
| 3   | `inspect that suspicious process`             | `inspect(pid)` → full metadata + all file effects        |
| 4   | `read the .env file it was accessing`         | `read_file(path)` → reveals the hardcoded credentials    |
| 5   | `show me my recent shell history`             | `goread_shell_history(user)` → context on what was typed |

---

## Expected Outcome

Godshell should:

1. **Find the sniffer process** via its `openat` effects on sensitive paths.
2. **Distinguish** the sniffer from the legitimate `ls` process (different access patterns).
3. **Read the .env contents** via `read_file` to show what secrets are exposed.
4. Show that **no `auditd` rules needed** — the eBPF observer already captured
   every `openat` call.

---

## Cleanup

```bash
kill %1 %2 2>/dev/null
rm -rf /tmp/godshell_demo_*
```
