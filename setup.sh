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

# 1. Check for sudo
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}Error: This script must be run as root (use sudo).${NC}"
   exit 1
fi

# 2. System Dependencies
echo -e "${YELLOW}[1/3] Installing system dependencies...${NC}"
apt-get update
apt-get install -y clang llvm libelf-dev libbpf-dev \
                   linux-headers-$(uname -r) build-essential pkg-config \
                   libglib2.0-dev libdbus-1-dev

# 3. Check for Go
echo -e "${YELLOW}[2/3] Verifying Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed.${NC}"
    echo -e "Please install Go 1.22+ from https://go.dev/doc/install"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "Found Go version: ${GO_VERSION}"

# 4. Build Godshell
echo -e "${YELLOW}[3/4] Compiling Godshell (eBPF + Go binary)...${NC}"
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
