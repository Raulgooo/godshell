# Godshell: Capabilities and Tooling Summary

Godshell is an eBPF-powered, terminal-based observability tool designed to provide deep, real-time insights into system activity. It is specifically built to cater to two primary audiences: human operators requiring immediate situational awareness, and Large Language Models (LLMs) acting as autonomous security or debugging agents.

At its core, Godshell functions by attaching eBPF tracepoints directly to the Linux kernel to monitor critical system calls (`execve`, `openat`, `connect`, `sched_process_exit`). This allows it to bypass user-space tracing overhead and capture high-fidelity behavioral events as they occur.

## The Concept of a "Frozen Snapshot"

Instead of forcing users or LLMs to parse a chaotic, infinitely scrolling stream of logs, Godshell centers around the concept of a **"Point-in-Time Snapshot."**

When an operator triggers a snapshot (by pressing `s`), Godshell halts the rendering of live events and creates a deep copy of the system's current state. This frozen snapshot aggregates all historical context—who spawned what, what files were opened, what network connections were made—into a highly structured, immutable view.

This approach completely eliminates race conditions when analyzing the state of the machine. An LLM can take its time querying the snapshot without the ground truth changing beneath it.

---

## Tooling and Query Arsenal

Once inside Snapshot Mode, Godshell exposes a suite of query commands designed to drill down into specific processes, memory regions, and system artifacts.

### 1. `[i]nspect <pid>` (Process Inspection)

- **Purpose:** Deep-dive into a specific process and its behavioral signature.
- **Output:** Generates a structured view showing the process's command line, binary path, CPU/Memory estimations, child processes, and a complete history of its network and filesystem interactions since Godshell started (or the process spawned). Effects include timestamps, occurrence counts, and aggregated network traffic (bytes sent/received).

### 2. `[f]amily <pid>` (Lineage Tracking)

- **Purpose:** Understand the execution ancestry of a target process.
- **Output:** Renders a visual ASCII tree strictly isolated to the target's direct lineage. It shows the target's parents all the way to PID 1, its immediate siblings, and the full tree of its spawned descendants. Crucially, it dynamically resolves missing parent or child links directly from `/proc/` if the eBPF layer missed the initial fork/exec event.

### 3. `[s]earch <pattern>` (Behavioral Grep)

- **Purpose:** Global search across all active and recently exited processes.
- **Output:** Returns a flat, easily parseable list of processes that match the input string. Importantly, Godshell doesn't just match process names or binaries; it searches the _effects_. For example, `s connect` will list every process that made a network connection, and `s /etc/passwd` will reveal who touched that file.

### 4. `[m]aps <pid>` (Memory Layout Analysis)

- **Purpose:** Expose the internal memory layout of a process.
- **Output:** Parses `/proc/<pid>/maps` and condenses it into a human-readable summary. It groups memory regions by binary/library and aggregates their total sizes and permissions (e.g., highlighting `rwxp` regions which are indicative of JIT compilers or potential injection vectors).

### 5. `[l]ibraries <pid>` (Linked Object Discovery)

- **Purpose:** Reveal the dynamic dependencies of a running executable.
- **Output:** Resolves the process's executable path via `/proc/<pid>/exe` and safely executes `ldd` against it, returning the list of shared object (`.so`) libraries the binary relies on.

### 6. `[t]race <pid>` (System Call Tracing)

- **Purpose:** Capture a short burst of raw syscall activity for a target process.
- **Output:** Spawns a background `strace -c -p <pid>` instance for 5 seconds. Once finished, it dumps a highly structured histogram showing which syscalls the process is spamming, the success/error rates, and the total time spent in kernel space. Useful for instantly diagnosing hanging threads (`futex`) or heavy I/O polling.

### 7. `[c]at file <path> [offset] [limit]` (Safe File Reading)

- **Purpose:** Safely extract localized file content.
- **Output:** Attempts to read and decode a file via UTF-8. If the file is binary or unstructured, it automatically falls back to generating a safe, formatted hex dump. It respects optional offset and chunk limits to prevent localized memory exhaustion from massive files.

### 8. `[r]ead mem <pid> <address_hex> [size]` (Memory Extraction)

- **Purpose:** Read raw binary data directly from a living process's memory space.
- **Output:** Opens `/proc/<pid>/mem` and extracts a requested number of bytes starting from a specific virtual address. Outputs a structured, 16-byte aligned hex dump alongside the ASCII representation, similar to `xxd` or `pwndbg`.

---

## An Honest Summary of Godshell

**What it truly is:**
Godshell is a highly effective, specialized bridge between raw Linux kernel observables and Large Language Model reasoning architectures. It transforms the overwhelming noise of an operating system into structured, interrogable data structures.

It is brilliant for answering localized, behavioral questions:

- _"Who spawned this shell, what configs did it read, and what IPs did it call?"_
- _"Is this Python process spending all its time waiting on I/O, or is it spinning in a tight CPU loop?"_
- _"What is the exact hex value living at this mapped segment in this active Nginx worker?"_

**What it is NOT (Current Limitations):**

1.  **Not a Perfect Historical Sentinel:** Godshell currently relies heavily on `sys_enter_execve`. If a process started before Godshell was launched, or if an application heavily utilizes pure `fork`/`clone` (like Chrome or Postgres) without calling `execve`, Godshell is blind to their inherent creation events. While we added "Lazy Resolution" to query `/proc` when explicitly asked, the baseline state is not fully captured on boot.
2.  **Not an Intervention Engine:** Godshell strictly observes. It cannot kill processes, block network packets, or modify memory. It is a read-only oracle.
3.  **Potential eBPF Overhead:** While eBPF is incredibly fast, heavy tracing on highly concurrent servers (e.g., thousands of `openat` calls per second) will inherently impact system latency and memory usage within the ring buffer.
4.  **Single Node Only:** It provides deep introspection into the _local_ machine. It lacks any distributed or clustered context.

Ultimately, Godshell succeeds as an interactive, intelligent magnifying glass for the Linux OS, designed from the ground up to assume the user asking the questions is an AI trying to solve a complex security or operational puzzle.
