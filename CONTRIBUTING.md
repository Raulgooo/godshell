# Contributing to Godshell

Thank you for your interest in Godshell! We welcome contributions that help make system forensics more accessible and powerful.

## Development Setup

1. **Prerequisites**:
   - Linux Kernel 5.8+ with BTF enabled
   - Go 1.22+
   - Clang/LLVM 14+
   - `libbpf-dev`
2. **Build**:
   ```bash
   make
   ```
3. **Run Tests**:
   ```bash
   make test
   ```

## Contribution Flow

1. **Fork the Repo**: Create a personal fork of the project.
2. **Create a Feature Branch**: `git checkout -b feature/amazing-new-observer`.
3. **Commit Changes**: Use descriptive commit messages.
4. **Push & PR**: Open a Pull Request against the `main` branch.

## Standards

- **eBPF**: Keep probes minimal and performant. Avoid heavy computation in kernel space.
- **Go**: Follow standard Go formatting (`go fmt`) and idiomatic patterns.
- **Testing**: New features should include relevant unit tests or recorded event fixtures.

## Reporting Issues

Use the GitHub Issue tracker to report bugs or suggest features. Please include:

- Your kernel version (`uname -r`)
- Godshell version/commit
- Steps to reproduce (if applicable)

---

_Godshell is an investigatory tool. Use responsibly._
