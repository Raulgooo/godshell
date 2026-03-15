#!/bin/bash
set -e

# Godshell V1 Setup Script
# Automated environment configuration for Ubuntu/Debian systems.

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}--- Godshell Environment Setup ---${NC}"

# 1. OS and Sudo Check
if [[ "$OSTYPE" != "linux-gnu"* ]]; then
    echo -e "${RED}Error: Godshell requires Linux. Found: $OSTYPE${NC}"
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Error: This script must be run as root (use sudo).${NC}"
   exit 1
fi

# 2. Kernel and BTF Check
echo -e "${YELLOW}[1/5] Verifying kernel and BTF support...${NC}"
KERNEL_VERSION=$(uname -r | cut -d. -f1-2)
MIN_KERNEL="5.8"

if [ "$(printf '%s\n' "$MIN_KERNEL" "$KERNEL_VERSION" | sort -V | head -n1)" != "$MIN_KERNEL" ]; then
    echo -e "${RED}Error: Kernel version must be >= 5.8. Found: $KERNEL_VERSION${NC}"
    exit 1
fi

if [ ! -f /sys/kernel/btf/vmlinux ]; then
    echo -e "${RED}Error: BTF (BPF Type Format) is not enabled on this kernel.${NC}"
    echo -e "BTF is required for Godshell's eBPF features. Please enable CONFIG_DEBUG_INFO_BTF."
    exit 1
fi
echo -e "Found Kernel ${KERNEL_VERSION} with BTF support."

# 3. System Dependencies
echo -e "${YELLOW}[2/5] Installing system dependencies...${NC}"
apt-get update
apt-get install -y clang llvm libelf-dev libbpf-dev \
                   linux-headers-$(uname -r) build-essential pkg-config \
                   libglib2.0-dev libdbus-1-dev

# 4. Check for Go
echo -e "${YELLOW}[3/5] Verifying Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed.${NC}"
    echo -e "Please install Go 1.22+ from https://go.dev/doc/install"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MIN_GO="1.22"

if [ "$(printf '%s\n' "$MIN_GO" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_GO" ]; then
    echo -e "${RED}Error: Go version must be >= $MIN_GO. Found: $GO_VERSION${NC}"
    exit 1
fi
echo -e "Found Go version: ${GO_VERSION}"

# 5. Build Tools (bpf2go)
echo -e "${YELLOW}[4/5] Installing build-time tools (bpf2go)...${NC}"
go install github.com/cilium/ebpf/cmd/bpf2go@latest

# 6. Build Godshell
echo -e "${YELLOW}[5/5] Compiling Godshell (eBPF + Go binary)...${NC}"
make clean
make

echo -e ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  GODSHELL SETUP SUCCESSFUL!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e ""
echo -e "To start the investigation:"
echo -e "  ${YELLOW}sudo ./godshell${NC}"
echo -e ""
echo -e "To run in daemon mode:"
echo -e "  ${YELLOW}sudo ./godshell daemon${NC}"
echo -e ""
