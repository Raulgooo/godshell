# 🟡 Demo 07 — "The Microservice That Can't Connect"

**Difficulty:** Medium  
**Category:** DevOps / service debugging  
**Estimated setup:** 3 minutes  
**Godshell tools exercised:** `summary`, `search`, `inspect`, `family`, `gonetwork_state`, `goread_environ`, `trace`, `read_file`

---

## Scenario

You have a web application that depends on three backend services:
a database (PostgreSQL), a cache (Redis), and an API gateway. The app
keeps crashing on startup because **one of the dependencies isn't reachable**.

Instead of checking each service manually, you ask Godshell to figure out
which connection is failing and why.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_07_microservice_debug.sh
# Simulates a microservice that fails to connect to one of its dependencies

set -e

APP_DIR=$(mktemp -d /tmp/godshell_svc_XXXX)
REDIS_PORT=6379
POSTGRES_PORT=5432
API_PORT=8080

echo "🌐 Setting up microservice connectivity demo..."

# 1. Simulate "PostgreSQL" — a listener on 5432 (healthy)
echo "[✓] Starting fake PostgreSQL on port $POSTGRES_PORT..."
(while true; do echo "OK" | ncat -l -p $POSTGRES_PORT -w 2 2>/dev/null; done) &
PG_PID=$!

# 2. DON'T start Redis — this is the broken dependency
echo "[✗] Redis on port $REDIS_PORT is NOT running (intentional)"

# 3. Simulate "API Gateway" — a listener on 8080 (healthy)
echo "[✓] Starting fake API Gateway on port $API_PORT..."
(while true; do echo '{"status":"ok"}' | ncat -l -p $API_PORT -w 2 2>/dev/null; done) &
API_PID=$!

sleep 1

# 4. Create the app config
cat > "$APP_DIR/config.yaml" << EOF
app:
  name: order-service
  port: 3000

dependencies:
  postgres:
    host: 127.0.0.1
    port: $POSTGRES_PORT
    database: orders
    user: app_user
    password: s3cret_db_pass
  redis:
    host: 127.0.0.1
    port: $REDIS_PORT
  api_gateway:
    host: 127.0.0.1
    port: $API_PORT
    path: /api/v2
EOF

# 5. Launch the "application" that tries to connect to all three
echo "[*] Launching order-service..."
(
  export APP_NAME="order-service"
  export DB_PASSWORD="s3cret_db_pass"
  export REDIS_URL="redis://127.0.0.1:$REDIS_PORT"
  export API_GATEWAY="http://127.0.0.1:$API_PORT"

  while true; do
    # Read config
    cat "$APP_DIR/config.yaml" > /dev/null 2>&1

    # Try PostgreSQL (succeeds)
    echo "SELECT 1" | nc -w 1 127.0.0.1 $POSTGRES_PORT > /dev/null 2>&1 && echo "[app] postgres: OK" > /dev/null

    # Try Redis (fails — connection refused)
    echo "PING" | nc -w 1 127.0.0.1 $REDIS_PORT > /dev/null 2>&1 || echo "[app] redis: FAILED" > /dev/null

    # Try API Gateway (succeeds)
    echo "GET /health" | nc -w 1 127.0.0.1 $API_PORT > /dev/null 2>&1 && echo "[app] api: OK" > /dev/null

    # Simulate crash and retry
    echo "[app] startup failed — retrying in 5s..." > /dev/null
    sleep 5
  done
) &
APP_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  order-service PID:  $APP_PID                                    "
echo "║  PostgreSQL PID:     $PG_PID  (port $POSTGRES_PORT ✓)           "
echo "║  Redis:              NOT RUNNING (port $REDIS_PORT ✗)            "
echo "║  API Gateway PID:    $API_PID  (port $API_PORT ✓)               "
echo "║  Config: $APP_DIR/config.yaml                                    "
echo "║                                                                  "
echo "║  Now ask Godshell:                                               "
echo "║                                                                  "
echo "║  > my order-service keeps crashing, what's going on?             "
echo "║  > what network connections is it making?                        "
echo "║  > which connection is failing?                                  "
echo "║  > read the service config file                                  "
echo "║  > what environment variables does it have?                      "
echo "║  > trace its syscalls for 5 seconds                              "
echo "║                                                                  "
echo "║  To clean up:  kill $APP_PID $PG_PID $API_PID                   "
echo "║                rm -rf $APP_DIR                                   "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

| #   | Prompt                                              | Tool                                       | What It Reveals                                                              |
| --- | --------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------- |
| 1   | `my order-service keeps crashing, what's going on?` | `search("order-service")` → `inspect(pid)` | Finds the process, shows it's connecting to 3 endpoints                      |
| 2   | `what network connections is it making?`            | `gonetwork_state(pid)`                     | Shows connections to :5432 and :8080 succeed, but :6379 is absent or refused |
| 3   | `trace its syscalls to see the failure`             | `trace(pid, 5)`                            | Shows `connect()` returning `ECONNREFUSED` for port 6379                     |
| 4   | `read its config file`                              | `read_file(config.yaml)`                   | Reveals all three dependency endpoints + **hardcoded password**              |
| 5   | `what are its environment variables?`               | `goread_environ(pid)`                      | Shows `REDIS_URL`, `DB_PASSWORD` in the process environment                  |
| 6   | `show its process family`                           | `family(pid)`                              | Parent chain reveals how the service was launched                            |

---

## Expected Outcome

Godshell should:

1. **Identify the connectivity failure** — port 6379 (Redis) has no listener.
2. **Show successful connections** to PostgreSQL (:5432) and API Gateway (:8080) for contrast.
3. **Reveal the `ECONNREFUSED`** in the strace output — proving Redis isn't running.
4. **Expose the configuration** — including the hardcoded database password.
5. **Diagnose the root cause** without the user needing to check each service individually.

This is a classic **"it works on my machine"** debugging scenario that every
developer has faced.

---

## Cleanup

```bash
kill %1 %2 %3 2>/dev/null
rm -rf /tmp/godshell_svc_*
```
