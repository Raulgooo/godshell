// Package ssl provides a Go-side observer that attaches eBPF uprobes to
// libssl.so (OpenSSL), libnss3.so (Firefox NSS), and Go crypto/tls binaries,
// then streams plaintext SSL events for HTTP reconstruction.
package ssl

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

//go:embed ssl_bpfel.o
var sslBPFBytes []byte

// DirWrite / DirRead match ssl.bpf.c constants.
const (
	DirWrite = 0
	DirRead  = 1
)

// Event is the Go mirror of struct ssl_event in ssl.bpf.c.
// All fields must match size/alignment exactly.
type Event struct {
	Ts        uint64
	PID       uint32
	TID       uint32
	Direction uint8
	Pad       [3]uint8
	DataLen   uint32
	Comm      [16]byte
	Data      [4096]byte
}

// ParsedEvent is a decoded SSL event with string fields.
type ParsedEvent struct {
	Timestamp time.Time
	PID       uint32
	Comm      string
	Direction int // DirWrite or DirRead
	Data      []byte
}

// Observer holds BPF objects and all active uprobe links.
type Observer struct {
	spec   *ebpf.CollectionSpec
	coll   *ebpf.Collection
	rd     *ringbuf.Reader
	links  []link.Link
	Events chan ParsedEvent
}

// New creates and loads the SSL BPF collection but does not attach anything yet.
func New() (*Observer, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("RemoveMemlock: %w", err)
	}

	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(sslBPFBytes))
	if err != nil {
		return nil, fmt.Errorf("load ssl bpf spec: %w", err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return nil, fmt.Errorf("instantiate ssl bpf collection: %w", err)
	}

	rd, err := ringbuf.NewReader(coll.Maps["ssl_events"])
	if err != nil {
		coll.Close()
		return nil, fmt.Errorf("open ring buffer: %w", err)
	}

	return &Observer{
		spec:   spec,
		coll:   coll,
		rd:     rd,
		Events: make(chan ParsedEvent, 1024),
	}, nil
}

// AttachPID attaches uprobes to all relevant libraries/binaries used by the
// given PID. It is safe to call for multiple PIDs.
func (o *Observer) AttachPID(pid int) error {
	// Resolve the set of library paths we need to hook from /proc/pid/maps
	libs, err := parseMapsLibs(pid)
	if err != nil {
		return fmt.Errorf("parse maps for PID %d: %w", pid, err)
	}

	var lastErr error

	// OpenSSL (libssl.so)
	if ssl, ok := libs["libssl"]; ok {
		if err := o.attachLib(ssl, "SSL_write", "uprobe/SSL_write", pid); err != nil {
			lastErr = err
		}
		if err := o.attachLib(ssl, "SSL_read", "uprobe/SSL_read", pid); err != nil {
			lastErr = err
		}
		if err := o.attachRetLib(ssl, "SSL_read", "uretprobe/SSL_read", pid); err != nil {
			lastErr = err
		}
	}

	// NSS (libnss3.so — Firefox)
	if nss, ok := libs["libnss3"]; ok {
		if err := o.attachLib(nss, "PR_Write", "uprobe/PR_Write", pid); err != nil {
			lastErr = err
		}
		if err := o.attachLib(nss, "PR_Read", "uprobe/PR_Read", pid); err != nil {
			lastErr = err
		}
		if err := o.attachRetLib(nss, "PR_Read", "uretprobe/PR_Read", pid); err != nil {
			lastErr = err
		}
	}

	// Go crypto/tls — detect by looking for Go build ID in the executable
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	if isGoBinary(exePath) {
		if sym, off, err := findGoTLSSymbol(exePath, "crypto/tls.(*Conn).Write"); err == nil {
			if err := o.attachOffset(exePath, sym, off, "uprobe/go_tls_write", pid); err != nil {
				lastErr = err
			}
		}
		if sym, off, err := findGoTLSSymbol(exePath, "crypto/tls.(*Conn).Read"); err == nil {
			if err := o.attachOffset(exePath, sym, off, "uprobe/go_tls_read", pid); err != nil {
				lastErr = err
			}
			if err := o.attachRetOffset(exePath, sym, off, "uretprobe/go_tls_read", pid); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}

// Start begins reading events from the ring buffer in a goroutine.
func (o *Observer) Start() {
	go func() {
		for {
			record, err := o.rd.Read()
			if err != nil {
				return // reader closed
			}
			var e Event
			if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &e); err != nil {
				continue
			}
			comm := strings.TrimRight(string(e.Comm[:]), "\x00")
			data := make([]byte, e.DataLen)
			copy(data, e.Data[:e.DataLen])
			o.Events <- ParsedEvent{
				Timestamp: time.Now(),
				PID:       e.PID,
				Comm:      comm,
				Direction: int(e.Direction),
				Data:      data,
			}
		}
	}()
}

// Close detaches all probes and releases BPF resources.
func (o *Observer) Close() {
	o.rd.Close()
	for _, l := range o.links {
		l.Close()
	}
	o.coll.Close()
}

// ── Attach helpers ─────────────────────────────────────────────────────────

func (o *Observer) attachLib(libPath, symbol, progName string, pid int) error {
	prog, ok := o.coll.Programs[progName]
	if !ok {
		return fmt.Errorf("program %s not found in BPF collection", progName)
	}
	ex, err := link.OpenExecutable(libPath)
	if err != nil {
		return fmt.Errorf("open executable %s: %w", libPath, err)
	}
	var l link.Link
	if pid > 0 {
		l, err = ex.Uprobe(symbol, prog, &link.UprobeOptions{PID: pid})
	} else {
		l, err = ex.Uprobe(symbol, prog, nil)
	}
	if err != nil {
		return fmt.Errorf("uprobe %s:%s: %w", libPath, symbol, err)
	}
	o.links = append(o.links, l)
	return nil
}

func (o *Observer) attachRetLib(libPath, symbol, progName string, pid int) error {
	prog, ok := o.coll.Programs[progName]
	if !ok {
		return fmt.Errorf("program %s not found in BPF collection", progName)
	}
	ex, err := link.OpenExecutable(libPath)
	if err != nil {
		return fmt.Errorf("open executable %s: %w", libPath, err)
	}
	var (
		l link.Link
	)
	if pid > 0 {
		l, err = ex.Uretprobe(symbol, prog, &link.UprobeOptions{PID: pid})
	} else {
		l, err = ex.Uretprobe(symbol, prog, nil)
	}
	if err != nil {
		return fmt.Errorf("uretprobe %s:%s: %w", libPath, symbol, err)
	}
	o.links = append(o.links, l)
	return nil
}

func (o *Observer) attachOffset(binPath, symbol string, offset uint64, progName string, pid int) error {
	prog, ok := o.coll.Programs[progName]
	if !ok {
		return fmt.Errorf("program %s not found in BPF collection", progName)
	}
	ex, err := link.OpenExecutable(binPath)
	if err != nil {
		return fmt.Errorf("open executable %s: %w", binPath, err)
	}
	var l link.Link
	if pid > 0 {
		l, err = ex.Uprobe(symbol, prog, &link.UprobeOptions{PID: pid, Offset: offset})
	} else {
		l, err = ex.Uprobe(symbol, prog, &link.UprobeOptions{Offset: offset})
	}
	if err != nil {
		return fmt.Errorf("uprobe go %s offset=%d: %w", symbol, offset, err)
	}
	o.links = append(o.links, l)
	return nil
}

func (o *Observer) attachRetOffset(binPath, symbol string, offset uint64, progName string, pid int) error {
	prog, ok := o.coll.Programs[progName]
	if !ok {
		return fmt.Errorf("program %s not found in BPF collection", progName)
	}
	ex, err := link.OpenExecutable(binPath)
	if err != nil {
		return fmt.Errorf("open executable %s: %w", binPath, err)
	}
	var (
		l    link.Link
		err2 error
	)
	if pid > 0 {
		l, err2 = ex.Uretprobe(symbol, prog, &link.UprobeOptions{PID: pid, Offset: offset})
	} else {
		l, err2 = ex.Uretprobe(symbol, prog, &link.UprobeOptions{Offset: offset})
	}
	if err2 != nil {
		return fmt.Errorf("uretprobe go %s offset=%d: %w", symbol, offset, err2)
	}
	o.links = append(o.links, l)
	return nil
}

// ── /proc helpers ──────────────────────────────────────────────────────────

// parseMapsLibs extracts unique library base names → resolved paths from
// /proc/pid/maps. Returns e.g. {"libssl": "/usr/lib/libssl.so.3"}.
func parseMapsLibs(pid int) (map[string]string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 6 {
			continue
		}
		path := fields[5]
		if !strings.HasSuffix(path, ".so") && !strings.Contains(path, ".so.") {
			continue
		}
		base := filepath.Base(path)
		// Extract base name without version suffix e.g. "libssl" from "libssl.so.3"
		name := base
		if idx := strings.Index(name, ".so"); idx >= 0 {
			name = name[:idx]
		}
		// Prefer the first mapping found (usually the main .so file)
		if _, seen := result[name]; !seen {
			result[name] = path
		}
	}
	return result, nil
}

// isGoBinary returns true if the ELF at path contains a Go build ID section.
func isGoBinary(path string) bool {
	f, err := elf.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	return f.Section(".go.buildinfo") != nil || f.Section(".gosymtab") != nil
}

// findGoTLSSymbol resolves the virtual address of a Go symbol in an ELF binary.
// Returns (symbolName, offset, error). The offset is relative to the file start
// for use with link.Uprobe.
func findGoTLSSymbol(path, symbolName string) (string, uint64, error) {
	f, err := elf.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	syms, err := f.Symbols()
	if err != nil {
		// Try dynamic symbols if regular symbols table is absent
		syms, err = f.DynamicSymbols()
		if err != nil {
			return "", 0, fmt.Errorf("no symbol table in %s", path)
		}
	}

	for _, s := range syms {
		if s.Name == symbolName {
			// Convert virtual address to file offset via section info
			offset, err := vAddrToFileOffset(f, s.Value)
			if err != nil {
				return s.Name, s.Value, nil // fall back to VA
			}
			return s.Name, offset, nil
		}
	}
	return "", 0, fmt.Errorf("symbol %q not found in %s", symbolName, path)
}

// vAddrToFileOffset converts an ELF virtual address to its file offset.
func vAddrToFileOffset(f *elf.File, vaddr uint64) (uint64, error) {
	for _, prog := range f.Progs {
		if prog.Type != elf.PT_LOAD {
			continue
		}
		if vaddr >= prog.Vaddr && vaddr < prog.Vaddr+prog.Filesz {
			return vaddr - prog.Vaddr + prog.Off, nil
		}
	}
	return 0, fmt.Errorf("vaddr 0x%x not found in any LOAD segment", vaddr)
}

// ── Utility ────────────────────────────────────────────────────────────────

// AttachAllSSLProcesses discovers all running processes that link libssl,
// libnss3, or are Go binaries, and attaches uprobes to all of them.
// This is used for system-wide interception when no specific PID is given.
func AttachAllSSLProcesses(o *Observer) {
	entries, _ := os.ReadDir("/proc")
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		// Best-effort: ignore errors for individual PIDs
		_ = o.AttachPID(pid)
	}
}
