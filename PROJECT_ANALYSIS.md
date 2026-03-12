# Godshell Project Analysis

## Overview

**Godshell** is a high-performance, AI-native system forensics and investigatory platform designed for security engineers and system administrators. It bridges the gap between low-level system instrumentation (eBPF) and high-level reasoning (LLMs) by providing an interactive terminal environment where users can investigate system behavior using natural language.

The project treats the operating system state as a series of "immutable snapshots," allowing an AI agent to perform deep, reproducible forensics on processes, networks, and memory without the volatility of a live system interfering with the investigation.

## Tech Stack

The project is built with a modern, performance-oriented stack:

- **Primary Language**: Go (v1.21+) - for the daemon, client, and TUI.
- **Instrumentation**: eBPF (C / CO-RE) - used for low-overhead, kernel-level observation of process lifecycles and system events.
- **TUI Framework**: Charmbracelet Stack (`bubbletea`, `lipgloss`, `bubbles`) - providing a rich, responsive, and aesthetically pleasing terminal interface.
- **AI Engine**: LLM Integration (OpenRouter/Anthropic) - acts as the "brain" of the investigation, capable of autonomous tool execution.
- **Storage**: SQLite - manages snapshot persistence and metadata.
- **System Integration**: Deep integration with Linux `/proc` filesystem and kernel internals.
- **External Intel**: Hooks for VirusTotal and AbuseIPDB for reputation scoring.

## Capabilities

Godshell is designed for "Search and Destroy" style forensics:

1.  **System Snapping**: Captures the entire state of the system (process tree, open files, network sockets) into an immutable JSON snapshot.
2.  **Conversational Forensics**: An LLM agent ("Godshell") that can answer questions like _"Why is this process talking to an IP in Russia?"_ or _"Did this binary exist before the reboot?"_
3.  **Deep Inspection Tools**:
    - **Memory Mapping**: Real-time parsing of `/proc/pid/maps` and raw memory reading from `/proc/pid/mem`.
    - **Syscall Tracing**: On-demand 5-second `strace`-like capabilities against any PID.
    - **Lineage Tracking**: Reconstructing process parentage even for processes that have already exited.
    - **Network Forensics**: Direct extraction of TCP/UDP states and connection targets.
    - **Binary Integrity**: SHA-256 hashing and string extraction from executables.
4.  **Remote Operations**: A Daemon/Client architecture allowing godshell to monitor remote servers and expose the telemetry to a local TUI.

## Current Limitations

- **Linux Exclusive**: Due to heavy eBPF and `/proc` reliance, it cannot run on macOS or Windows.
- **Privilege Requirements**: Requires root/sudo for kernel-level monitoring and memory reading.
- **Snapshot-Centric**: While powerful for forensics, it lacks a "live streaming" mode for real-time EDR-style alerting.
- **Tooling Sandbox**: Several advanced tools (like SSL interception and automated report generation) are currently disabled or in a prototype state.
- **UX Complexity**: The 2-panel TUI requires familiarity with vim-like navigation and terminal hotkeys.

## Potential Improvements

- **Live Event Streaming**: Integrate eBPF ring buffers for real-time event logging alongside snapshotting.
- **Advanced Memory Forensics**: Add YARA rule scanning against process memory and automated shellcode detection.
- **Cross-Snapshot Diffs**: Implement a high-level diffing engine to show exactly what changed in the system between two points in time.
- **Automated Forensics Reports**: Re-enable and polish the `report_*` tools to generate professional Markdown/PDF summaries of an investigation.
- **Container/Namespace Support**: Better visibility into Docker/K8s environments by mapping PIDs back to container IDs and namespaces.
- **Graph Visualization**: An optional web-based UI or ASCII graph representing process relationships and network flows.
