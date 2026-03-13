# Godshell Build System

GO := go
EBPF_GEN := github.com/cilium/ebpf/cmd/bpf2go
BINARY := godshell

.PHONY: all
all: ebpf build

.PHONY: ebpf
ebpf:
	@echo "Generating eBPF bytecode..."
	@$(GO) generate ./...

.PHONY: build
build:
	@echo "Building godshell..."
	@$(GO) build -o $(BINARY) main.go

.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test ./...

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY)
	@rm -f observer/bpf_bpfeb.go observer/bpf_bpfeb.o
	@rm -f observer/bpf_bpfel.go observer/bpf_bpfel.o
	@rm -f godshell_* test_heap_bin
	@find . -name "*.o" -delete

.PHONY: help
help:
	@echo "Godshell Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all     - Build eBPF and Go binary (default)"
	@echo "  ebpf    - Generate eBPF bytecode using bpf2go"
	@echo "  build   - Build the godshell Go binary"
	@echo "  test    - Run all Go tests"
	@echo "  clean   - Remove binaries and generated files"
