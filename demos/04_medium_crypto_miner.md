# 🟡 Demo 04 — "The Crypto Miner in Disguise"

**Difficulty:** Medium  
**Category:** Malware analysis / binary forensics  
**Estimated setup:** 3 minutes  
**Godshell tools exercised:** `summary`, `search`, `inspect`, `gohash_binary`, `goextract_strings`, `get_libraries`, `get_maps`, `goread_environ`, `gonetwork_state`

---

## Scenario

Someone dropped a binary on your server that pretends to be a legitimate
system utility (`system-update`) but is actually a crypto miner. It:

- Burns CPU on mathematical operations (mining simulation)
- Connects to a suspicious external "pool" (simulated with a local listener)
- Reads environment variables looking for cloud credentials
- Has suspicious strings embedded in the binary

You need Godshell to find it, identify it, and prove it's malicious.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_04_crypto_miner.sh
# Deploys a fake crypto miner disguised as a system utility

set -e

MINER_DIR=$(mktemp -d /tmp/godshell_miner_XXXX)
POOL_PORT=3333

echo "⛏️  Setting up crypto miner demo..."

# 1. Create the fake miner binary (a bash script pretending to be compiled)
cat > "$MINER_DIR/system-update" << 'MINER_SCRIPT'
#!/usr/bin/env bash
# This is a simulated crypto miner for demo purposes only.
# In reality, it would be a compiled binary.

# Suspicious environment variable harvesting
AWS_KEY="${AWS_ACCESS_KEY_ID:-not_found}"
AWS_SECRET="${AWS_SECRET_ACCESS_KEY:-not_found}"

# Simulate mining — CPU-intensive but harmless
while true; do
  # Fake "hashing" — just busy work
  for i in $(seq 1 5000); do
    echo -n "$i" | sha256sum > /dev/null 2>&1
  done

  # Periodically "phone home" to the pool
  echo "STRATUM_SUBSCRIBE" | nc -w 1 127.0.0.1 3333 2>/dev/null || true

  # Touch some files to look legit
  ls /var/log/ > /dev/null 2>&1
  cat /proc/cpuinfo > /dev/null 2>&1
done
MINER_SCRIPT
chmod +x "$MINER_DIR/system-update"

# 2. Start the fake mining pool listener
echo "[*] Starting fake mining pool on port $POOL_PORT..."
(while true; do echo '{"result":true}' | ncat -l -p $POOL_PORT -w 2 2>/dev/null; done) &
POOL_PID=$!
sleep 1

# 3. Launch the disguised miner with suspicious env vars
echo "[*] Launching 'system-update' (the disguised miner)..."
AWS_ACCESS_KEY_ID="AKIA_FAKE_DEMO_KEY_12345" \
AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/FAKE_DEMO_SECRET" \
MINING_POOL="stratum+tcp://127.0.0.1:3333" \
WALLET_ADDR="44AFFq5kSiGBoZ4NMDwYtN18obc8AemS33DBLWs3H7otXft3" \
"$MINER_DIR/system-update" &
MINER_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  Fake miner PID:   $MINER_PID                                   "
echo "║  Mining pool PID:  $POOL_PID (port $POOL_PORT)                   "
echo "║  Binary location:  $MINER_DIR/system-update                     "
echo "║                                                                  "
echo "║  Now ask Godshell:                                               "
echo "║                                                                  "
echo "║  > what's using the most CPU right now?                          "
echo "║  > inspect that system-update process                            "
echo "║  > what strings are embedded in the binary?                      "
echo "║  > what are its environment variables?                           "
echo "║  > hash the binary for reputation lookup                         "
echo "║  > what network connections does it have?                        "
echo "║  > show me its memory layout                                     "
echo "║  > what shared libraries is it linked to?                        "
echo "║                                                                  "
echo "║  To clean up:  kill $MINER_PID $POOL_PID                        "
echo "║                rm -rf $MINER_DIR                                 "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

This is a **malware triage workflow** — you're building a case:

| #   | Prompt                                   | Tool                      | Evidence Gathered                                                                     |
| --- | ---------------------------------------- | ------------------------- | ------------------------------------------------------------------------------------- |
| 1   | `what's using the most CPU?`             | `summary`                 | `system-update` appears high — suspicious name for a script                           |
| 2   | `inspect that process`                   | `inspect(pid)`            | Full cmdline, parent, file effects show `/proc/cpuinfo` reads and `/var/log` scanning |
| 3   | `extract the strings from that binary`   | `goextract_strings(path)` | Reveals `STRATUM_SUBSCRIBE`, `sha256sum`, mining pool references                      |
| 4   | `what are its environment variables?`    | `goread_environ(pid)`     | **Jackpot**: `AWS_ACCESS_KEY_ID`, `MINING_POOL`, `WALLET_ADDR` exposed                |
| 5   | `hash the binary`                        | `gohash_binary(pid)`      | SHA-256 hash for VirusTotal/reputation check                                          |
| 6   | `what network connections does it have?` | `gonetwork_state(pid)`    | Shows connection to port 3333 — known mining pool port                                |
| 7   | `show the memory map`                    | `get_maps(pid)`           | Memory layout analysis                                                                |
| 8   | `what libraries does it link to?`        | `get_libraries(pid)`      | Shared object analysis                                                                |

---

## Expected Outcome

Godshell should:

1. **Spot the CPU anomaly** — `system-update` consuming disproportionate resources.
2. **Reveal the mining infrastructure** through embedded strings (`STRATUM`, `sha256sum`).
3. **Expose stolen credentials** in the environment (`AWS_ACCESS_KEY_ID`).
4. **Show the C2/pool connection** on port 3333.
5. **Provide the binary hash** for external reputation databases.

This exercises **7 different tools** in a single investigatory flow, proving
Godshell can conduct comprehensive malware triage from a single snapshot.

---

## Cleanup

```bash
kill %1 %2 2>/dev/null
rm -rf /tmp/godshell_miner_*
```
