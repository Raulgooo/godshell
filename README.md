# godshell

A shell that knows what is happening on your system before you ask.

Most LLM terminals are blind. They run `top`, `ps`, `df`, `dmesg` to figure
out what is going on every time you ask something. GODshell has been watching
your system since it started. No probing. No roundtrips. Just answers.

```
you: why is my PC slow?

godshell: rust-analyzer (PID 9012) has been at 89% CPU for the last
          14 minutes, triggered after you opened godshell/src/main.rs.
          It is indexing the workspace. Kill it?
```

---

## How it works

GODshell runs two things in parallel:

- An **eBPF observer** attached to kernel tracepoints. It watches file access,
  process launches, exits, and network connections at the kernel level, with no
  overhead on your processes.

- A **context engine** that combines those events with live system metrics
  (CPU, RAM, top processes, active connections) into a snapshot.

When you ask something, the snapshot goes into the LLM prompt. The LLM already
has the answer. It does not need to explore.

```
kernel
  └── eBPF programs (C)
        openat / execve / exit / tcp_connect
        └── ring buffer
              └── observer daemon (Go)
                    └── event store (SQLite)
                          └── context engine
                                └── LLM prompt
                                      └── you
```

---

## Requirements

- Linux kernel 5.8 or later
- BTF enabled (`/sys/kernel/btf/vmlinux` must exist)
- Go 1.22+
- clang 14+
- libbpf-dev
- Root or `CAP_BPF` capability to load eBPF programs
- [ollama](https://ollama.ai) running locally

---

## Install

```bash
# system dependencies
sudo apt install clang llvm libelf-dev libbpf-dev \
                 linux-headers-$(uname -r) build-essential

# verify BTF
ls /sys/kernel/btf/vmlinux

# ollama
curl -fsSL https://ollama.ai/install.sh | sh
ollama pull qwen2.5:7b
ollama serve &

# build godshell
git clone https://github.com/youruser/godshell
cd godshell
make
```

---

## Usage

```bash
sudo ./godshell
```

Root is required to load the eBPF programs. After the observers are attached,
the shell prints a prompt. You can type natural language or use the built-in
commands below.

```
godshell v0.1
observers attached: openat execve exit tcp_connect
model: qwen2.5:7b

> why is my PC slow?
> what opened ~/.ssh/id_rsa in the last 10 minutes?
> what have I run since 3pm?
> which app is using the most network?
```

### Built-in commands

```
context        print the current system snapshot the LLM sees
history [n]    show last n kernel events (default 50)
clear          clear the screen
exit           quit
```

Any input that is not a built-in goes to the LLM with the current context.

---

## Configuration

Configuration lives in `~/.config/godshell/config.toml`. Created on first run
with defaults.

```toml
[llm]
model   = "qwen2.5:7b"
host    = "http://localhost:11434"
timeout = 30   # seconds

[observer]
ring_buffer_mb = 16
max_events     = 5000   # events kept in SQLite
ignore_pids    = []     # PIDs to exclude from events

[context]
top_processes     = 8   # processes shown in snapshot
recent_events     = 30  # events included in each prompt
lookback_minutes  = 60  # how far back the context engine looks
```

---

## Architecture

### eBPF observers (`ebpf/observer.bpf.c`)

Four tracepoints, compiled to bytecode and embedded in the Go binary:

| Tracepoint                   | What it captures                     |
| ---------------------------- | ------------------------------------ |
| `syscalls/sys_enter_openat`  | Every file opened, by which process  |
| `syscalls/sys_enter_execve`  | Every program launched, with argv[0] |
| `sched/sched_process_exit`   | Every process exit with PID and name |
| `syscalls/sys_enter_connect` | Every outbound TCP connection        |

All events go through a BPF ring buffer to the observer daemon. Filtering
happens in kernel space to avoid saturating the ring buffer with irrelevant
events (kernel threads, godshell's own syscalls).

### Observer daemon (`observer/`)

Loads the eBPF object, attaches all tracepoints, reads the ring buffer in a
goroutine, and writes events to the event store. Runs for the lifetime of the
process.

### Event store (`store/`)

SQLite database at `/var/lib/godshell/events.db`. Schema:

```sql
CREATE TABLE events (
    id        INTEGER PRIMARY KEY,
    ts        INTEGER NOT NULL,   -- nanoseconds since boot
    ts_wall   INTEGER NOT NULL,   -- unix timestamp
    pid       INTEGER NOT NULL,
    ppid      INTEGER,
    uid       INTEGER,
    comm      TEXT NOT NULL,      -- process name
    type      TEXT NOT NULL,      -- open | exec | exit | connect
    data      TEXT                -- path, argv, or remote addr
);

CREATE INDEX idx_ts   ON events(ts_wall);
CREATE INDEX idx_pid  ON events(pid);
CREATE INDEX idx_type ON events(type);
CREATE INDEX idx_comm ON events(comm);
```

### Context engine (`context/`)

Builds the snapshot string that goes into every LLM prompt. Called on each
query, takes about 5ms. Combines:

- Live CPU, RAM, disk, and network metrics via gopsutil
- Top 8 processes by CPU
- Last N kernel events from the store, filtered to remove noise

### LLM bridge (`llm/`)

Thin HTTP client for the ollama API. Streams the response token by token to
stdout. Prepends the system prompt and context snapshot to every user query.
Maintains a short conversation history (last 6 turns) so follow-up questions
work naturally.

---

## Development

```bash
# build everything (compiles eBPF + Go)
make

# build only the eBPF bytecode
make ebpf

# run with verbose event output
sudo ./godshell -v

# run against a specific ollama model
sudo ./godshell -model llama3.1:8b

# run tests (does not require root, uses recorded event fixtures)
make test
```

### Working in a VM

eBPF programs that pass the verifier are safe, but during development you will
write programs that crash the kernel. Use a VM.

```bash
# create a dev VM with virt-manager or:
sudo apt install qemu-kvm libvirt-daemon-system virt-manager
virt-manager
```

Build and test inside the VM. Once the observers are stable, you can run on
bare metal.

### Adding a new observer

1. Add the tracepoint handler to `ebpf/observer.bpf.c`
2. Add the event type constant to `observer/types.go`
3. Handle the new type in the ring buffer reader in `observer/daemon.go`
4. Add any relevant context extraction to `context/snapshot.go`
5. Run `make ebpf` to recompile the bytecode

---

## Roadmap

### GODshell (now)

REPL that runs on top of your shell. You switch to it to ask questions,
then go back to bash. The observer runs the whole time. This validates the
core idea: a system-aware LLM that does not probe.

### GODshell-C (next)

Wrapper around bash. GODshell becomes your shell. Commands that look like
real shell commands (binary exists in PATH) pass through to bash directly.
Everything else goes to the LLM. No context switching. One place for
everything.

```
$ ls -la              → passes to bash, you see normal output
$ git status          → passes to bash
$ why did that fail?  → goes to LLM with the last command's stderr in context
$ clean up the logs   → LLM generates the command, asks for confirmation
```

### LLMINUX (long term)

GODshell is step 1. The full vision is an OS-level MCP layer for Linux:
a privileged daemon (`mcpd`) that exposes every kernel subsystem as MCP tools
to a local LLM. GODshell proves the value of the context engine. mcpd
industrializes it.

---

## Why not just use an existing AI terminal?

Existing tools (aider, Claude in terminal, shell-gpt) are useful. They are
also blind. They see what you explicitly give them. GODshell's observer has
been watching since boot. The difference shows up in questions like:

- "what process touched that file?"
- "why did that command fail 20 minutes ago?"
- "what changed on this machine in the last hour?"

No existing terminal tool can answer those without running commands first.

---

## License

MIT
