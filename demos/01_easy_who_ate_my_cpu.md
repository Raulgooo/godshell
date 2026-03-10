# 🟢 Demo 01 — "Who Ate My CPU?"

**Difficulty:** Easy  
**Category:** Performance triage  
**Estimated setup:** 2 minutes  
**Godshell tools exercised:** `summary`, `inspect`, `search`, `family`

---

## Scenario

You are a developer who just noticed your laptop fans spinning like crazy.
Instead of manually running `top`, `htop`, or guessing, you ask Godshell.

Godshell has been watching every process launch since boot. It already knows
who's eating CPU, what spawned it, and what files it has been touching.

---

## Setup Script

Run the following **before** launching Godshell (or in a separate terminal
while Godshell is already running):

```bash
#!/usr/bin/env bash
# demos/setup_01_cpu_hog.sh
# Creates a realistic CPU hog that mimics a runaway build process

set -e

echo "🔥 Starting CPU hog simulation..."

# 1. Simulate a runaway compiler — uses 100% of one core
#    Reads random source files to generate file-open events
(
  while true; do
    # Busy loop to burn CPU
    for i in $(seq 1 10000); do
      : $((i * i))
    done
    # Touch some files so Godshell sees openat events
    cat /etc/hostname > /dev/null 2>&1
    cat /proc/self/status > /dev/null 2>&1
  done
) &
HOG_PID=$!

# 2. Simulate a second process — a "build watcher" that spawns children
(
  while true; do
    sleep 2
    echo "checking..." > /dev/null
    ls /tmp > /dev/null 2>&1
  done
) &
WATCHER_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════╗"
echo "║  CPU hog running as PID: $HOG_PID                       "
echo "║  Watcher running as PID: $WATCHER_PID                   "
echo "║                                                          "
echo "║  Now launch Godshell and try:                            "
echo "║                                                          "
echo "║  > why is my CPU usage so high?                          "
echo "║  > what process is using the most CPU?                   "
echo "║  > what files has PID $HOG_PID been opening?             "
echo "║  > show me the process family for PID $HOG_PID           "
echo "║                                                          "
echo "║  To clean up:  kill $HOG_PID $WATCHER_PID                "
echo "╚══════════════════════════════════════════════════════════╝"
echo ""
echo "Press Ctrl+C to stop everything."

wait
```

---

## What to Ask Godshell

Once the hog is running and Godshell has captured a snapshot:

| #   | Prompt                             | What Godshell Does                                                    |
| --- | ---------------------------------- | --------------------------------------------------------------------- |
| 1   | `why is my CPU so high?`           | Calls `summary` → identifies the bash loop at the top of the CPU list |
| 2   | `inspect PID <hog_pid>`            | Calls `inspect` → shows full cmdline, parent info, file effects       |
| 3   | `what spawned that process?`       | Calls `family` → shows the parent/child lineage                       |
| 4   | `what files has it been touching?` | Already visible in `inspect`, or calls `search` to cross-reference    |

---

## Expected Outcome

Godshell should:

1. **Immediately identify the busy-loop bash** process at the top of CPU usage
   without running `top` or `ps`.
2. **Show the parent chain** — your terminal → bash → the hog script.
3. **List file effects** — the `openat` events for `/etc/hostname`, `/proc/self/status`,
   etc. that prove the process is actively doing I/O.
4. Demonstrate that **zero probing commands** were needed. Godshell already knew.

---

## Cleanup

```bash
# Kill both background processes
kill %1 %2 2>/dev/null
# Or if you noted the PIDs:
# kill $HOG_PID $WATCHER_PID
```
