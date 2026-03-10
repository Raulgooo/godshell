# 🔴 Demo 10 — "Full Stack Debugging: The Broken Deploy"

**Difficulty:** Hard  
**Category:** Full-stack debugging / DevOps incident  
**Estimated setup:** 5 minutes  
**Godshell tools exercised:** `summary`, `search`, `inspect`, `family`, `gonetwork_state`, `goread_environ`, `trace`, `read_file`, `get_maps`, `get_libraries`, `gohash_binary`, `goextract_strings`

---

## Scenario

You just deployed a new version of your application. It's broken:

1. **The web server** starts but can't bind to its port (something else is
   using it — an old zombie from the previous deploy).
2. **The worker process** starts but crashes because it reads a stale config
   file that points to the wrong database.
3. **The migration script** ran but failed silently — it touched the DB
   config but didn't actually apply the schema changes.
4. **A monitoring agent** keeps restarting because its binary was
   accidentally overwritten during the deploy.

You need Godshell to untangle this mess and find all four problems without
manually SSH-ing around and checking each component.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_10_broken_deploy.sh
# Simulates a multi-component broken deployment

set -e

DEPLOY_DIR=$(mktemp -d /tmp/godshell_deploy_XXXX)
WEB_PORT=9090

echo "🚀 Setting up broken deploy simulation..."
echo "   Deploy directory: $DEPLOY_DIR"
echo ""

# ─── Create the "application" files ────────────────────────────────────

mkdir -p "$DEPLOY_DIR/config" "$DEPLOY_DIR/bin" "$DEPLOY_DIR/logs"

# Good config (what the new version expects)
cat > "$DEPLOY_DIR/config/app.prod.yaml" << 'EOF'
server:
  port: 9090
  host: 0.0.0.0

database:
  host: db-prod-v2.internal
  port: 5432
  name: myapp_v2
  user: deploy_user
  password: new_secure_pass_2026

worker:
  concurrency: 8
  queue: jobs_v2

monitoring:
  agent: datadog
  api_key: dd-FAKE-api-key-demo
EOF

# Stale config (leftover from previous version — the bug)
cat > "$DEPLOY_DIR/config/app.yaml" << 'EOF'
server:
  port: 9090
  host: 0.0.0.0

database:
  host: db-prod-v1.internal
  port: 5432
  name: myapp_v1
  user: old_user
  password: old_password_2025

worker:
  concurrency: 4
  queue: jobs_v1

monitoring:
  agent: datadog
  api_key: dd-STALE-revoked-key
EOF

# Create fake binaries
cat > "$DEPLOY_DIR/bin/web-server" << 'SCRIPT'
#!/bin/bash
# Simulated web server
echo "[web] Starting web server on port 9090..."
cat /tmp/godshell_deploy_*/config/app.prod.yaml > /dev/null 2>&1
# Try to bind — will fail because zombie has the port
nc -l -p 9090 -w 60 2>/dev/null || echo "[web] ERROR: port 9090 already in use"
SCRIPT
chmod +x "$DEPLOY_DIR/bin/web-server"

cat > "$DEPLOY_DIR/bin/worker" << 'SCRIPT'
#!/bin/bash
# Simulated worker that reads the WRONG config
echo "[worker] Starting worker..."
# BUG: reads app.yaml instead of app.prod.yaml
cat /tmp/godshell_deploy_*/config/app.yaml > /dev/null 2>&1
echo "[worker] Connecting to database..."
# Tries to connect to the old DB (will fail)
nc -w 1 127.0.0.1 15432 2>/dev/null || echo "[worker] ERROR: cannot reach db-prod-v1.internal"
sleep 999
SCRIPT
chmod +x "$DEPLOY_DIR/bin/worker"

cat > "$DEPLOY_DIR/bin/migrate" << 'SCRIPT'
#!/bin/bash
# Migration script that "runs" but doesn't actually do anything useful
echo "[migrate] Running database migrations..."
cat /tmp/godshell_deploy_*/config/app.prod.yaml > /dev/null 2>&1
sleep 2
echo "[migrate] Migration complete (0 changes applied)"
# Exits with 0 even though nothing happened — silent failure
SCRIPT
chmod +x "$DEPLOY_DIR/bin/migrate"

cat > "$DEPLOY_DIR/bin/monitor-agent" << 'SCRIPT'
#!/bin/bash
# Monitoring agent that crash loops
echo "[monitor] Starting monitoring agent..."
# Reads the stale config with revoked API key
cat /tmp/godshell_deploy_*/config/app.yaml > /dev/null 2>&1
sleep 1
echo "[monitor] ERROR: API key rejected, shutting down"
exit 1
SCRIPT
chmod +x "$DEPLOY_DIR/bin/monitor-agent"

# ─── Problem 1: Zombie process holding the port ───────────────────────

echo "[*] Problem 1: Creating zombie process on port $WEB_PORT..."
(ncat -l -k -p $WEB_PORT > /dev/null 2>&1) &
ZOMBIE_PID=$!
sleep 1

# ─── Problem 2: Worker reads stale config ─────────────────────────────

echo "[*] Problem 2: Launching worker with stale config..."
(
  export APP_CONFIG="$DEPLOY_DIR/config/app.yaml"  # WRONG! Should be app.prod.yaml
  export DB_HOST="db-prod-v1.internal"
  export DEPLOY_VERSION="v2.1.0"
  "$DEPLOY_DIR/bin/worker"
) &
WORKER_PID=$!

# ─── Problem 3: Migration that does nothing ───────────────────────────

echo "[*] Problem 3: Running silent-fail migration..."
(
  "$DEPLOY_DIR/bin/migrate"
) &
MIGRATE_PID=$!

# ─── Problem 4: Monitor agent crash loop ──────────────────────────────

echo "[*] Problem 4: Starting crash-looping monitor agent..."
(
  for i in $(seq 1 5); do
    "$DEPLOY_DIR/bin/monitor-agent" 2>/dev/null
    sleep 2
  done
  sleep 999
) &
MONITOR_PID=$!

# ─── Now try to start the web server (it will fail) ──────────────────

echo "[*] Attempting to start web server (will fail — port busy)..."
(
  export DEPLOY_VERSION="v2.1.0"
  export APP_ENV="production"
  "$DEPLOY_DIR/bin/web-server" 2>/dev/null
  sleep 999
) &
WEB_PID=$!

sleep 3

echo ""
echo "╔═════════════════════════════════════════════════════════════════════╗"
echo "║  🚀 BROKEN DEPLOY SIMULATION ACTIVE                                "
echo "║                                                                     "
echo "║  Problem 1 — Port zombie:    PID $ZOMBIE_PID (holding :$WEB_PORT)  "
echo "║  Problem 2 — Stale config:   PID $WORKER_PID (reads app.yaml)      "
echo "║  Problem 3 — Silent migrate: PID $MIGRATE_PID (did nothing)        "
echo "║  Problem 4 — Crash loop:     PID $MONITOR_PID (monitor-agent)      "
echo "║  Failed web server:          PID $WEB_PID                           "
echo "║  Deploy dir:                 $DEPLOY_DIR                            "
echo "║                                                                     "
echo "║  ── Investigation Playbook ──────────────────────────────────────── "
echo "║                                                                     "
echo "║  TRIAGE:                                                            "
echo "║  > give me a full system summary — what's running?                 "
echo "║  > are there any processes that recently crashed?                   "
echo "║  > what's listening on port 9090?                                   "
echo "║                                                                     "
echo "║  DIAGNOSE:                                                          "
echo "║  > the web server can't start — why?                               "
echo "║  > the worker is connecting to the wrong database — show config    "
echo "║  > the migration script ran but did nothing — inspect it           "
echo "║  > the monitor agent keeps restarting — why?                       "
echo "║                                                                     "
echo "║  EVIDENCE:                                                         "
echo "║  > read the stale config file vs the correct one                   "
echo "║  > compare the env vars across the worker and web server           "
echo "║  > trace the worker's syscalls                                     "
echo "║  > extract strings from the monitor binary                        "
echo "║                                                                     "
echo "║  To clean up:                                                       "
echo "║  kill $ZOMBIE_PID $WORKER_PID $MIGRATE_PID $MONITOR_PID $WEB_PID  "
echo "║  rm -rf $DEPLOY_DIR                                                "
echo "╚═════════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## Investigation Playbook

### Phase 1 — Triage

| #   | Prompt                                      | Tool               | Finding                                                                        |
| --- | ------------------------------------------- | ------------------ | ------------------------------------------------------------------------------ |
| 1   | `system summary — what just happened?`      | `summary`          | Multiple processes, some ghosts (monitor crash-looping, migration exited fast) |
| 2   | `any processes recently crashed or exited?` | `summary` (ghosts) | `monitor-agent` appears multiple times as ghosts — crash loop!                 |
| 3   | `search for anything on port 9090`          | `search("9090")`   | Finds the zombie ncat AND the failed web-server                                |

### Phase 2 — Diagnose Each Problem

| #   | Prompt                                                     | Tool                          | Root Cause                                                   |
| --- | ---------------------------------------------------------- | ----------------------------- | ------------------------------------------------------------ |
| 4   | `inspect the zombie on port 9090`                          | `inspect` → `gonetwork_state` | Old ncat still bound to :9090 — blocking the new deploy      |
| 5   | `inspect the worker — what config is it reading?`          | `inspect(worker_pid)`         | File effects show it opened `app.yaml` not `app.prod.yaml`   |
| 6   | `read app.yaml and compare with app.prod.yaml`             | `read_file` × 2               | Stale config points to v1 database & old credentials         |
| 7   | `what env vars does the worker have?`                      | `goread_environ`              | `APP_CONFIG` set to wrong path, `DB_HOST` is the old one     |
| 8   | `inspect the migration — did it actually change anything?` | `inspect(migrate_pid)`        | Very short lifespan, only read the config, no DB connections |
| 9   | `why does the monitor keep crashing?`                      | `inspect` ghost instances     | Each ghost reads `app.yaml` → uses revoked API key           |
| 10  | `extract strings from the monitor binary`                  | `goextract_strings`           | Reveals hardcoded error message about API key rejection      |

### Phase 3 — Evidence Gathering

| #   | Prompt                                                  | Tool                                             |
| --- | ------------------------------------------------------- | ------------------------------------------------ |
| 11  | `trace the worker for 5 seconds`                        | `trace` — shows failed `connect()` syscalls      |
| 12  | `show the process family for the crash-looping monitor` | `family` — reveals the restart loop parent       |
| 13  | `hash the web-server and worker binaries`               | `gohash_binary` — forensic hashes for comparison |
| 14  | `show the memory layout of the worker`                  | `get_maps` — confirms Python/bash runtime        |
| 15  | `what libraries does the web server use?`               | `get_libraries` — linked library audit           |

---

## Expected Outcome — Four Root Causes

| Problem              | Root Cause                                         | How Godshell Found It                               |
| -------------------- | -------------------------------------------------- | --------------------------------------------------- |
| **Port conflict**    | Old ncat zombie holding :9090                      | `search("9090")` + `gonetwork_state`                |
| **Stale config**     | Worker reads `app.yaml` instead of `app.prod.yaml` | `inspect` file effects + `read_file` comparison     |
| **Silent migration** | Script exits 0 without DB changes                  | `inspect` shows no network effects, short lifespan  |
| **Crash loop**       | Monitor uses revoked API key from stale config     | Multiple ghosts + `goextract_strings` reveals error |

This exercises **12 different tools** across a single realistic investigation
and demonstrates Godshell's value for **deployment debugging** — a scenario
that costs engineering teams hours every week.

---

## Cleanup

```bash
kill %1 %2 %3 %4 %5 2>/dev/null
rm -rf /tmp/godshell_deploy_*
```
