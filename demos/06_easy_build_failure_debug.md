# 🟢 Demo 06 — "Why Did My Build Fail?"

**Difficulty:** Easy  
**Category:** Developer debugging  
**Estimated setup:** 2 minutes  
**Godshell tools exercised:** `summary`, `search`, `inspect`, `family`, `read_file`, `goread_environ`

---

## Scenario

You just ran a build and it failed. You're not sure why — was it a missing
dependency? Wrong config? Environment variable issue?

Instead of re-running with `--verbose`, scrolling through 500 lines of output,
and guessing, you ask Godshell: **"what went wrong?"**

Godshell has been watching every file the build touched, every subprocess it
spawned, and every one that exited. It knows exactly what happened.

---

## Setup Script

```bash
#!/usr/bin/env bash
# demos/setup_06_build_fail.sh
# Simulates a multi-step build pipeline that fails at step 3

set -e

BUILD_DIR=$(mktemp -d /tmp/godshell_build_XXXX)

echo "🔨 Setting up failing build simulation..."

# 1. Create a fake project structure
mkdir -p "$BUILD_DIR/src" "$BUILD_DIR/build" "$BUILD_DIR/deps"

cat > "$BUILD_DIR/src/main.c" << 'EOF'
#include <stdio.h>
#include <missing_header.h>  // This will cause the failure

int main() {
    printf("Hello from fake project\n");
    return 0;
}
EOF

cat > "$BUILD_DIR/Makefile" << 'MAKEFILE'
all: deps lint compile link

deps:
	@echo "Checking dependencies..."
	@ls /usr/include/stdio.h > /dev/null 2>&1
	@sleep 1
	@echo "Dependencies OK"

lint:
	@echo "Running linter..."
	@sleep 1
	@echo "Lint passed"

compile:
	@echo "Compiling src/main.c..."
	@sleep 1
	@gcc -c src/main.c -o build/main.o 2>&1 || { echo "COMPILE FAILED"; exit 1; }

link:
	@echo "Linking..."
	@gcc build/main.o -o build/app 2>&1
MAKEFILE

# 2. Create a "config" file that a real build might read
cat > "$BUILD_DIR/.buildrc" << 'EOF'
CC=gcc
CFLAGS=-Wall -Werror -O2
TARGET=x86_64-linux-gnu
BUILD_TYPE=release
PARALLEL_JOBS=4
EOF

# 3. Run the failing build (this generates events for Godshell)
echo "[*] Running the build (it WILL fail at compile step)..."
echo ""

cd "$BUILD_DIR"
(
  # Read config
  cat "$BUILD_DIR/.buildrc" > /dev/null

  # Step 1: deps check
  ls /usr/include/stdio.h > /dev/null 2>&1 || true
  sleep 1

  # Step 2: lint (succeeds)
  cat "$BUILD_DIR/src/main.c" > /dev/null
  sleep 1

  # Step 3: compile (fails)
  gcc -c "$BUILD_DIR/src/main.c" -o "$BUILD_DIR/build/main.o" 2>/dev/null
  COMPILE_EXIT=$?

  if [ $COMPILE_EXIT -ne 0 ]; then
    echo "BUILD FAILED at compile step (exit code: $COMPILE_EXIT)" >&2
  fi

  sleep 999
) &
BUILD_PID=$!

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  Build process PID: $BUILD_PID                                   "
echo "║  Build dir: $BUILD_DIR                                           "
echo "║                                                                  "
echo "║  Now ask Godshell:                                               "
echo "║                                                                  "
echo "║  > what build processes just ran?                                "
echo "║  > which build step failed? (check recently exited processes)    "
echo "║  > what files did the build pipeline read?                       "
echo "║  > read the source file that was being compiled                  "
echo "║  > what's in the build config file?                              "
echo "║  > what compiler was invoked and what were its arguments?        "
echo "║                                                                  "
echo "║  To clean up:  kill $BUILD_PID; rm -rf $BUILD_DIR                "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

wait
```

---

## What to Ask Godshell

| #   | Prompt                                           | Tool                  | What It Reveals                                                |
| --- | ------------------------------------------------ | --------------------- | -------------------------------------------------------------- |
| 1   | `what build processes just ran?`                 | `summary`             | Shows active + recently exited processes — `gcc`, `make`, etc. |
| 2   | `show recently exited processes — did any fail?` | `summary`             | Ghost list shows `gcc` exited quickly (short lifespan = crash) |
| 3   | `inspect the gcc process`                        | `inspect(pid)`        | Full cmdline reveals which file was being compiled             |
| 4   | `show the process family tree`                   | `family(pid)`         | bash → make → gcc chain visible                                |
| 5   | `read the source file it was compiling`          | `read_file(path)`     | The `#include <missing_header.h>` line is immediately visible  |
| 6   | `read the build config`                          | `read_file(.buildrc)` | Shows the compiler flags being used                            |
| 7   | `what environment was the build using?`          | `goread_environ(pid)` | Full env — PATH, CC, CFLAGS etc.                               |

---

## Expected Outcome

Godshell should:

1. **Show `gcc` as a ghost** (recently exited) with a very short lifespan — indicating a crash.
2. **Reveal the compile target** via `inspect` — the exact source file and flags.
3. **Let you read the offending source** — the `#include <missing_header.h>` is the smoking gun.
4. **Show the full build pipeline** — deps → lint → compile (failed) — via the process family tree.

This proves Godshell is useful for **everyday developer debugging**, not just security.

---

## Cleanup

```bash
kill %1 2>/dev/null
rm -rf /tmp/godshell_build_*
```
