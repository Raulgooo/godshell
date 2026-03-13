# Godshell V1 Test Recommendations

To ensure long-term stability and catch regressions in the kernel-to-LLM pipeline, we recommend implementing the following tests:

## 1. Context Engine Unit Tests

- **Tree Reconstruction**: Validate that `BuildProcessTree` correctly links parents and children even when some processes in the middle are missing from the snapshot.
- **Snapshot Serialization**: Ensure `FrozenSnapshot` can be JSON-serialized and deserialized without losing event metadata.

## 2. eBPF Integration Tests (Mocked)

- **Ring Buffer Decoder**: Create a test that feeds raw bytes (representing kernel structs) into the `observer` package and verifies that the correct `Event` structs are emitted.
- **Filtering Logic**: Verify that events from `godshell`'s own PID are correctly discarded when the BPF variable is set.

## 3. Tool Execution Safety

- **Truncation Logic**: Verify that extremely large tool outputs (e.g., a massive `read_file` or `read_memory`) are truncated correctly before being sent to the LLM to avoid context window overflows.
- **Permission Mapping**: Test that privileged tools (`read_memory`, `scan_heap`) return a clean "Permission Denied" error when run without root, rather than crashing.

## 4. TUI Polish

- **Resize Handling**: Test that the viewport and side panel adjust their widths correctly when `WindowSizeMsg` is received.
- **Unified Navigation**: Verify that `j/k` keys navigate the process list only when the text input is blurred.

## 5. End-to-End Forensics (The "Ghost" Test)

- **Scripted Investigation**: A test that:
  1. Starts a dummy process.
  2. Takes a snapshot.
  3. Programmatically asks the LLM "What is the PID of [dummy name]?"
  4. Asserts the LLM response contains the correct PID.

---

_Targeting 80%+ coverage on the `context/` and `llm/` packages should be the priority for v1._
