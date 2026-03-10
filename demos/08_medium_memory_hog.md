# 🟡 Demo 08 — "Who's Using All The RAM?"

**Difficulty:** Medium  
**Category:** Performance / memory investigation  
**Estimated setup:** 2 minutes  
**Godshell tools exercised:** `summary`, `inspect`, `search`, `get_maps`, `get_libraries`, `trace`, `family`

---

## Scenario

Your machine is sluggish and `free -h` shows almost no available memory.
Something is eating your RAM, but you don't know what. Is it a memory leak?
A legitimate workload? A browser with too many tabs?

You ask Godshell to figure it out. It already has CPU and memory metrics for
every tracked process, and can dive into memory maps and linked libraries
for a full memory forensic analysis.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_08_memory_hog.sh
# Simulates a memory-hogging application with growing allocations

set -e

echo "🧠 Starting memory hog simulation..."

# 1. The "leaking app" — allocates memory in a growing array
#    Python is ideal here because we can control allocations precisely
python3 -c "
import time
import os

data = []
print(f'[memleak] PID={os.getpid()} — allocating memory...')

# Allocate ~10MB every 2 seconds
for i in range(50):
    chunk = bytearray(10 * 1024 * 1024)  # 10 MB
    data.append(chunk)
    time.sleep(2)

# Hold the memory forever
print(f'[memleak] Holding {len(data) * 10} MB — press Ctrl+C to stop')
time.sleep(99999)
" &
LEAK_PID=$!

# 2. A "normal" app for comparison — uses modest memory
python3 -c "
import time
import os

# Just sit here using a little memory
small_data = bytearray(5 * 1024 * 1024)  # 5 MB total
print(f'[normal-app] PID={os.getpid()} — using 5MB (stable)')
time.sleep(99999)
" &
NORMAL_PID=$!

# 3. A helper that reads files to generate open events
(
  while true; do
    cat /proc/meminfo > /dev/null 2>&1
    cat /proc/loadavg > /dev/null 2>&1
    sleep 10
  done
) &
MONITOR_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  Memory leak sim PID:  $LEAK_PID  (grows ~10MB every 2s)        "
echo "║  Normal app PID:       $NORMAL_PID  (stable 5MB)                "
echo "║  System monitor PID:   $MONITOR_PID                             "
echo "║                                                                  "
echo "║  Wait ~30 seconds for the leak to grow, then ask Godshell:      "
echo "║                                                                  "
echo "║  > which process is using the most memory?                       "
echo "║  > inspect the biggest memory consumer                          "
echo "║  > show me its memory layout (maps)                             "
echo "║  > what shared libraries does it use?                            "
echo "║  > trace its syscalls — is it still allocating?                  "
echo "║  > compare it to the normal app's memory                        "
echo "║                                                                  "
echo "║  To clean up:  kill $LEAK_PID $NORMAL_PID $MONITOR_PID          "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

| #   | Prompt                                                | Tool                  | What It Reveals                                       |
| --- | ----------------------------------------------------- | --------------------- | ----------------------------------------------------- |
| 1   | `which process is using the most memory?`             | `summary`             | Python process shows high memory metric vs. others    |
| 2   | `inspect that python process`                         | `inspect(pid)`        | Shows growing RSS, cmdline reveals the script         |
| 3   | `show me its memory map`                              | `get_maps(pid)`       | Massive `[heap]` region visible — the smoking gun     |
| 4   | `what libraries is it using?`                         | `get_libraries(pid)`  | Python's libpython, libc — confirms it's a Python app |
| 5   | `trace its syscalls for 5 seconds`                    | `trace(pid, 5)`       | Shows `brk()` or `mmap()` calls — active allocation   |
| 6   | `now show the normal app's memory map for comparison` | `get_maps(other_pid)` | Small, stable heap — healthy baseline                 |
| 7   | `show the process families of both`                   | `family(pid)`         | Confirms both are separate Python interpreters        |

---

## Expected Outcome

Godshell should:

1. **Rank processes by memory** — the leaking process at the top with growing RSS.
2. **Reveal the heap explosion** via `get_maps` — the `[heap]` region is disproportionately large.
3. **Show active allocation** via `trace` — `brk()`/`mmap()` syscalls proving it's still growing.
4. **Enable comparison** between the leaker and the stable app.
5. **Identify Python** as the runtime via `get_libraries`.

This is a scenario **every developer and sysadmin** encounters regularly.

---

## Cleanup

```bash
kill %1 %2 %3 2>/dev/null
```
