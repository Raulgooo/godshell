# 👻 The Ghost in the Machine: Godshell Ultimate Demo

This scenario demonstrates a "Hacker News breaking" capability of Godshell: **detecting and deconstructing an in-memory "ghost" process that has deleted its own binary from disk.**

Standard Linux tools (`ps`, `ls`, `top`) will see a process, but they won't easily show you _what_ it is or _where_ it came from once the file is gone. Godshell uses eBPF and direct memory access to solve the mystery.

---

## 1. Environment Setup

### 🚀 Launch the "Ghost"

Run the following command to launch a detached, self-deleting malicious process.

```bash
python3 GHOST_EXPLOIT.py
```

_Expected Output:_ `Malicious process launched and detached. Binary at /tmp/.syslog-helper will be deleted.`

### 🔍 Verify the "Ghost" exists

Try to find the binary:

```bash
ls -l /tmp/.syslog-helper
# Output: ls: cannot access '/tmp/.syslog-helper': No such file or directory
```

Check the process list:

```bash
ps aux | grep GHOST
# You might see the process, but the 'COMMAND' will look confused or point to a deleted file.
```

---

## 2. The Godshell Investigation

Now, open **Godshell** and prepare to be amazed.

### 📥 The Initial Prompt

Ask Godshell to find the anomaly:

> "Godshell, I suspect there is a 'ghost' process running on my system. A binary was executed and then deleted from /tmp. Find it, tell me its PID, and explain what it is doing right now."

---

## 3. What Godshell will do (The "Magic")

Watch the tool cards as Godshell performs the following steps autonomously:

1.  **`summary`**: It will scan the process tree and likely notice a process with a "deleted" executable (Linux marks these as `(deleted)` in `/proc/pid/exe`).
2.  **`inspect`**: It will see the process lineage and realize it was launched from `/tmp/.syslog-helper`.
3.  **`gonetwork_state`**: It will discover the process has an active connection to `8.8.8.8:53` (our simulated C2).
4.  **`trace`**: It will run a 5-second trace and see the process aggressively calling `openat()` on various files in `/home`, looking for `.env` files.
5.  **`read_memory`**: Since the file is deleted on disk, Godshell can read the _in-memory_ strings of the process to find hardcoded URLs, keys, or the original script content.

---

## 4. Conclusion

By the end of the session, Godshell will have reconstructed the entire attack:

- **Identifier**: PID and disguised name.
- **Modus Operandi**: Self-deletion and masquerading.
- **Intent**: Data exfiltration (searching for `.env` files) and persistence (dialing out).

**This is why Godshell is different: it doesn't just look at files; it looks at the living system.**
