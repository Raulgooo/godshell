# Godshell Demos

Ten scenarios to test Godshell's capabilities, covering cybersecurity,
debugging, performance, and DevOps. Each demo includes a setup script
and an investigation playbook.

> **Prerequisite**: Godshell must be running with root/`CAP_BPF` so the eBPF
> observers are attached. Run each setup script in a **separate terminal**.

---

## Scenarios

### 🔒 Cybersecurity

| #                                       | Difficulty | Name                             | Category                  | Key Tools                                      |
| --------------------------------------- | :--------: | -------------------------------- | ------------------------- | ---------------------------------------------- |
| [01](01_easy_who_ate_my_cpu.md)         |  🟢 Easy   | **Who Ate My CPU?**              | Performance triage        | `summary` `inspect` `family`                   |
| [02](02_easy_secret_file_access.md)     |  🟢 Easy   | **Who Touched My SSH Keys?**     | File access audit         | `search` `read_file` `goread_shell_history`    |
| [03](03_medium_reverse_shell.md)        | 🟡 Medium  | **Catch the Reverse Shell**      | Intrusion detection       | `gonetwork_state` `trace` `goread_environ`     |
| [04](04_medium_crypto_miner.md)         | 🟡 Medium  | **The Crypto Miner in Disguise** | Malware analysis          | `gohash_binary` `goextract_strings` `get_maps` |
| [05](05_hard_full_incident_response.md) |  🔴 Hard   | **Full Incident Response**       | Kill chain reconstruction | All 14 tools                                   |

### 🛠️ Debugging & DevOps

| #                                     | Difficulty | Name                                    | Category             | Key Tools                                 |
| ------------------------------------- | :--------: | --------------------------------------- | -------------------- | ----------------------------------------- |
| [06](06_easy_build_failure_debug.md)  |  🟢 Easy   | **Why Did My Build Fail?**              | Dev debugging        | `inspect` `family` `read_file`            |
| [07](07_medium_microservice_debug.md) | 🟡 Medium  | **The Microservice That Can't Connect** | Service debugging    | `gonetwork_state` `trace` `read_file`     |
| [08](08_medium_memory_hog.md)         | 🟡 Medium  | **Who's Using All The RAM?**            | Memory investigation | `get_maps` `get_libraries` `trace`        |
| [09](09_easy_process_archaeology.md)  |  🟢 Easy   | **What Just Happened on My Machine?**   | Process timeline     | `summary` `search` `goread_shell_history` |
| [10](10_hard_broken_deploy.md)        |  🔴 Hard   | **The Broken Deploy**                   | Full-stack debugging | 12 tools across 4 problems                |

---

## Quick Start

```bash
# Terminal 1 — start Godshell
sudo ./godshell

# Terminal 2 — run any demo setup
bash demos/setup_XX_name.sh
```

Then ask Godshell the questions listed in each demo's playbook.
