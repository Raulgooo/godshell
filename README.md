# godshell - speaking directly to your kernel

- A new way to interact with your kernel
  ![Godshell Demo](demo.gif)

**godshell.**

Godshell is a process observation system that uses eBPF to watch your kernel's internal events since boot, and models the system state as a snapshot ready to be queried by an LLM.

Think of it as a TUI for your kernel, but instead of using a CLI, you use natural language to query the system state. The use cases are very broad because you can ask anything about the system state, and the LLM can make correlations that a human might miss.

I started this project because I realized that LLM terminals were stupid. Don't get me wrong, LLM tools for devs are great, but they are still missing that "system understanding" part. They need to probe using human commands to understand your system. Why would you make an LLM query your system the way a human would? It's just inefficient.

## That's godshell's purpose: Becoming an inference layer on top of the OS

## Features

godshell is currently very alpha, but still has some cool features that i managed to find useful:

- snapshots: godshell allows you to take snapshots, manually or automatically every x minutes, and query them using an LLM
- process panel: godshell allows you to view the process tree, and selecting a process you find interesting for the LLM to analyze (Tree is WIP!)
- the godshell agent is packed with a set of tools that range from general analysis to specific forensics tasks, such as:
  - fileless malware detection
  - memory string extraction
  - lineage tracking
  - network connections

## Demos

Here i wanted to show some cool features of godshell in action, I was able to make more amusing PoCs but I couldn't get good footage of them, so here I have some PoCs that I managed to record.

### Fileless malware detection

![Fileless malware detection](demo.gif)

### ghost/recently exited processes (It can navigate them!)

![Ghost processes](deadprocs.gif)

### Connection analysis of short-lived processes(it also works with active ones!)

![Connection analysis](slp.gif)

## **When I am able to, I'll record more demos to show-off godshell's capabilities**

## Architecture

godshell is composed of two main parts:

- The godshell daemon: This is a systemd service that runs in the background and collects events from the kernel using eBPF
- The godshell TUI: This is a TUI tool that allows you to interact with the godshell daemon, it is almost as speaking with your kernel.

godshell daemon collects events from the kernel using eBPF and stores them in a SQLite database, it also exposes a UNIX socket via HTTP to communicate with the TUI. Currently it serves 4 main tracepoints, but I plan to add more in the future, including uprobes.

## Configuration

For configuration, run `godshell config`.

## Installation

### Option 1: One-Line Installation (Recommended)

Download the pre-built package for your distribution from the [latest release](https://github.com/raulgooo/godshell/releases/latest) and install it:

- **Debian/Ubuntu**: `sudo dpkg -i godshell_*.deb`
- **RHEL/CentOS**: `sudo rpm -i godshell_*.rpm`
- **Arch Linux**: `sudo pacman -U godshell_*.pkg.tar.zst`
- **Alpine**: `apk add godshell_*.apk`

Installing via package automatically sets up the **Godshell Daemon** as a `systemd` service.

### Option 2: Build from Source

Godshell requires a modern Linux kernel (5.8+) with BTF enabled.

```bash
git clone https://github.com/raulgooo/godshell
cd godshell
sudo ./setup.sh
```

---

---

## 📖 Usage

```bash
sudo godshell daemon(if not running)

```

```bash
godshell

```

ensure to run `godshell config` for your first time running godshell.

---

## Roadmap

- [ ] **Add more RE tools**: Add automatic API mapping via ssl libraries and more reverse engineering tools.
- [ ] **Add better memory analysis tools**: I plan to add more memory analysis tools to the daemon, including uprobes.
- [ ] **Add more tracepoints**: I plan to add more tracepoints to the daemon, including uprobes.
- [ ] **Cross-Snapshot Diffs**: See exactly what changed between two points in time.
- [ ] **YARA Integration**: Automatically scan process memory for malware signatures.
- [ ] **Container/K8s Support**: Map PIDs back to container IDs and namespaces.
- [ ] **Live Alerting**: EDR-style real-time notifications for suspicious eBPF events.

---

## License

MIT. See [LICENSE](LICENSE) for details.

_Godshell is an experimental Tool. Use responsibly._
