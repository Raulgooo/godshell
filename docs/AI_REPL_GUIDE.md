# AI Agent REPL Implementation Guide

This document summarizes the core infrastructure implemented for the **Godshell AI Agent REPL**. Use this as a reference when building the `llm/bridge.go` execution loop and conversation manager.

---

## 1. Core Architecture

The AI Agent REPL operates on **Frozen Snapshots**. These are immutable, point-in-time clones of the system state.

1.  **Freeze**: A `FrozenSnapshot` is taken (`tree.TakeSnapshot()`).
2.  **Tooling**: The LLM queries this snapshot using high-signal JSON methods.
3.  **Persistence**: Snapshots can be saved/loaded from a SQLite database.

---

## 2. Structured JSON Endpoints

To provide the LLM with "TUI-like" high-signal data (collapsed noise, grouped apps), use the following methods on a `FrozenSnapshot` instance.

### Methods (`context/snapshot.go` & `context/query.go`)

| Method                   | Description                                                                       |
| :----------------------- | :-------------------------------------------------------------------------------- |
| `SummaryJSON()`          | Returns grouped applications, active processes, and recently exited ghosts.       |
| `InspectJSON(pid)`       | Returns deep metadata for a single PID, including collapsed file/network effects. |
| `SearchJSON(query)`      | Regex search across command names, binary paths, and effect targets (IPs/Files).  |
| `ProcessFamilyJSON(pid)` | returns a recursive lineage tree (parent -> target -> children).                  |

---

## 3. Data Transfer Objects (JSON Models)

All models are defined in `context/snapshot.go`.

### `SnapshotSummaryJSON`

```go
type SnapshotSummaryJSON struct {
    Timestamp      time.Time      `json:"timestamp"`
    ActiveGroups   []AppGroupJSON `json:"active_groups"`
    RecentlyExited []GhostJSON    `json:"recently_exited"`
}
```

### `ProcessInspectJSON`

```go
type ProcessInspectJSON struct {
    PID            uint32       `json:"pid"`
    PPID           uint32       `json:"ppid"`
    Comm           string       `json:"comm"`
    BinaryPath     string       `json:"binary_path,omitempty"`
    State          string       `json:"state"` // "ACTIVE" or "EXITED"
    Uptime         string       `json:"uptime"`
    StartTime      time.Time    `json:"start_time"`
    CpuUsage       float64      `json:"cpu_usage_percent"`
    MemoryUsage    uint64       `json:"memory_usage_kb"`
    Parent         string       `json:"parent"` // "name (pid)"
    Children       []ChildJSON  `json:"children"`
    NetworkEffects []EffectJSON `json:"network_effects"`
    FileEffects    []EffectJSON `json:"file_effects"`
}
```

### `EffectJSON` (Collapsed Data)

```go
type EffectJSON struct {
    Label    string `json:"label"`      // Collapse target (e.g. "/tmp/*")
    Count    uint64 `json:"count"`      // Total operations
    Unique   int    `json:"unique_targets,omitempty"`
    Category string `json:"category"`   // "open" or "connect"
    Subtype  string `json:"subtype,omitempty"` // "unix_socket" or "network"
}
```

---

## 4. Snapshot Persistence (`store/store.go`)

The system uses a SQLite database (`godshell.db`) to persist snapshots as JSON blobs.

- **`store.Init(path)`**: Initializes the DB and `snapshots` table.
- **`store.SaveSnapshot(label, snap)`**: Persists a snapshot. Returns the `id`.
- **`store.LoadSnapshot(id)`**: Fetches a snapshot by ID and performs a full unmarshal (handling the `EffectDetail` interface automatically).
- **`store.ListSnapshots()`**: Returns metadata (`id`, `label`, `timestamp`) for all stored captures.

---

## 5. Conversational Loop Strategy

When implementing the REPL in `llm/bridge.go`, follow this flow:

### A. The System Prompt

Inject a system message explaining the environment:

> "You are an AI Security Engineer. You have access to Godshell snapshots. Use the provided tools (`summary`, `inspect`, `search`, `family`) to investigate system behavior. A snapshot from <TIMESTAMP> is currently loaded."

### B. Tool Mapping

Map LLM function calls directly to the `FrozenSnapshot` JSON methods.

- `list_processes()` -> `snap.SummaryJSON()`
- `inspect_process(pid)` -> `snap.InspectJSON(pid)`
- `find_activity(query)` -> `snap.SearchJSON(query)`

### C. Context Injection

Do not send the raw snapshot JSON (multi-megabyte). Only send the output of the tools the LLM explicitly calls. This keeps the token count low and signal high.

### D. Multi-Snapshot Reasoning

If the user asks "What changed since the last snapshot?", the bridge should:

1.  Load Snapshot A.
2.  Run tool (e.g. `SummaryJSON`).
3.  Load Snapshot B.
4.  Run tool again.
5.  Let the LLM compare the diff.

---

## 6. Development Status

- [x] **JSON Serialization**: `tree.go` handles interface unmarshaling.
- [x] **JSON Methods**: Implemented in `snapshot.go` and `query.go`.
- [x] **SQLite Store**: Implemented in `store/store.go`.
- [/] **REPL Bridge**: _Next Objective._
