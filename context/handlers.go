package ctxengine

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"godshell/config"
	"godshell/observer"
)

func NewProcessTree(cfg config.Config) *ProcessTree {
	ignored := make(map[string]struct{})
	for _, p := range cfg.IgnoredProcesses {
		ignored[p] = struct{}{}
	}

	t := &ProcessTree{
		ByPID:                make(map[uint32]*ProcessNode),
		Ghosts:               make(map[uint32]*ProcessNode),
		MaxEffectsPerProcess: cfg.MaxEffectsPerProcess,
		CaptureNetwork:       cfg.CaptureNetwork,
		CaptureFileIO:        cfg.CaptureFileIO,
		IgnoredProcesses:     ignored,
		ProcPath:             cfg.ProcPath,
		SysPath:              cfg.SysPath,
	}
	return t
}

func (t *ProcessTree) EnrichProcessMetadata(pid uint32) {
	// Re-verify PID still exists in our tree
	t.mu.Lock()
	proc, ok := t.ByPID[pid]
	if !ok {
		t.mu.Unlock()
		return
	}
	t.mu.Unlock()

	// Slow /proc reads outside the lock
	comm := readCmdline(t.ProcPath, pid, proc.Comm)
	ppid := readPPID(t.ProcPath, pid)
	bin := readBinaryPath(t.ProcPath, pid)

	t.mu.Lock()
	defer t.mu.Unlock()
	// Check again if still there
	if proc, ok = t.ByPID[pid]; ok {
		proc.Comm = comm
		proc.PPID = ppid
		proc.BinaryPath = bin

		// Update parent's children if needed
		if ppid > 0 {
			if parent, ok := t.ByPID[ppid]; ok {
				found := false
				for _, cpid := range parent.ChildrenPID {
					if cpid == pid {
						found = true
						break
					}
				}
				if !found {
					parent.ChildrenPID = append(parent.ChildrenPID, pid)
				}
			}
		}
	}
}

func (t *ProcessTree) EnrichConnectionMetadata(pid uint32, key string, bpfDetail *ConnectDetail) {
	var detail *ConnectDetail
	if bpfDetail != nil && bpfDetail.IP != "" {
		detail = bpfDetail
		// Still do DNS lookup if possible
		detail.Domain = reverseDNS(detail.IP)
	} else {
		detail = enrichConnect(t.ProcPath, pid)
	}

	if detail == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	proc, ok := t.ByPID[pid]
	if !ok {
		proc, ok = t.Ghosts[pid]
		if !ok {
			return
		}
	}

	eff, ok := proc.Effects[key]
	if !ok {
		return
	}

	// Update the existing effect with enriched data
	if d, ok := eff.Detail.(ConnectDetail); ok {
		// Preserve captured stats but update metadata
		detail.BytesSent += d.BytesSent
		detail.BytesRecv += d.BytesRecv
		eff.Detail = *detail
	} else {
		eff.Detail = *detail
	}

	// Re-key the effect if it was a fallback key and we now have a better one
	if strings.HasPrefix(key, "connect:") && detail.IP != "" {
		newKey := fmt.Sprintf("%s:%d", detail.IP, detail.Port)
		if detail.Domain != "" {
			newKey = fmt.Sprintf("%s (%s)", newKey, detail.Domain)
		}
		delete(proc.Effects, key)
		eff.Target = newKey
		proc.Effects[newKey] = eff
	}
}

// HandleEvent routes an observer event to the correct handler.
func (t *ProcessTree) HandleEvent(e observer.Event) {
	switch e.Type {
	case observer.EventExec:
		t.handleExec(e)
	case observer.EventOpen:
		if t.CaptureFileIO {
			t.handleOpen(e)
		}
	case observer.EventExit:
		t.handleExit(e)
	case observer.EventConnect:
		if t.CaptureNetwork {
			t.handleConnect(e)
		}
	}
}

func (t *ProcessTree) handleExec(e observer.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	comm := e.CommStr()
	if _, ok := t.IgnoredProcesses[comm]; ok {
		return
	}

	proc := &ProcessNode{
		PID:            e.Pid,
		BinaryPath:     e.PathStr(),
		Comm:           comm,
		Effects:        make(map[string]*Effect),
		StartTimestamp: time.Now(),
	}

	t.ByPID[e.Pid] = proc
}

// handleOpen upserts a file-open effect on the process.
func (t *ProcessTree) handleOpen(e observer.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc := t.getOrCreateProc(e)
	path := e.PathStr()

	if eff, ok := proc.Effects[path]; ok {
		eff.Count++
		eff.Last = time.Now()
	} else {
		if len(proc.Effects) >= t.MaxEffectsPerProcess {
			return
		}
		proc.Effects[path] = &Effect{
			Kind:   EffectOpen,
			Target: path,
			Count:  1,
			First:  time.Now(),
			Last:   time.Now(),
			Detail: OpenDetail{},
		}
	}
}

// handleExit moves a process from ByPID to Ghosts.
func (t *ProcessTree) handleExit(e observer.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc, ok := t.ByPID[e.Pid]
	if !ok {
		return // process was never tracked (uid < 1000 exec, etc.)
	}

	proc.EndTimestamp = time.Now()
	t.Ghosts[e.Pid] = proc
	delete(t.ByPID, e.Pid)
}

// handleConnect records a connect effect, enriched with IP/port/DNS from procfs.
func (t *ProcessTree) handleConnect(e observer.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc := t.getOrCreateProc(e)

	// Use a fallback key until enriched
	key := fmt.Sprintf("connect:%d:%d", e.Pid, time.Now().UnixNano())

	if len(proc.Effects) >= t.MaxEffectsPerProcess {
		return
	}

	proc.Effects[key] = &Effect{
		Kind:   EffectConnect,
		Target: key,
		Count:  1,
		First:  time.Now(),
		Last:   time.Now(),
		Detail: ConnectDetail{},
	}

	// In deferred enrichment mode, we don't do anything here.
	// Metadata will be gathered during TakeSnapshot.
}

// getOrCreateProc returns the ProcessNode for the given PID,
// creating a stub if the PID wasn't seen via exec (e.g., pre-existing process).
func (t *ProcessTree) getOrCreateProc(e observer.Event) *ProcessNode {
	if proc, ok := t.ByPID[e.Pid]; ok {
		return proc
	}
	proc := &ProcessNode{
		PID:            e.Pid,
		Comm:           e.CommStr(),
		Effects:        make(map[string]*Effect),
		StartTimestamp: time.Now(),
	}
	t.ByPID[e.Pid] = proc
	return proc
}

// readBinaryPath reads the exe symlink from <proc>/<pid>/exe.
// Returns empty string if the process is gone or unreadable.
func readBinaryPath(procPath string, pid uint32) string {
	target, err := os.Readlink(fmt.Sprintf("%s/%d/exe", procPath, pid))
	if err != nil {
		return ""
	}
	return target
}

// readCmdline reads the full command line from <proc>/<pid>/cmdline.
// It replaces null bytes with spaces to return a clean string.
func readCmdline(procPath string, pid uint32, fallback string) string {
	cmdlineBytes, err := os.ReadFile(fmt.Sprintf("%s/%d/cmdline", procPath, pid))
	if err != nil || len(cmdlineBytes) == 0 {
		return fallback
	}

	cmdline := strings.ReplaceAll(string(cmdlineBytes), "\x00", " ")
	cmdline = strings.TrimSpace(cmdline)
	if cmdline == "" {
		return fallback
	}
	return cmdline
}

// EvictGhosts runs in a background goroutine, pruning dead processes
// older than maxAge every 10 seconds.
func (t *ProcessTree) EvictGhosts(maxAge time.Duration) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		t.mu.Lock()
		for pid, proc := range t.Ghosts {
			if time.Since(proc.EndTimestamp) > maxAge {
				delete(t.Ghosts, pid)
			}
		}
		t.mu.Unlock()
	}
}

// readPPID reads the parent PID from <proc>/<pid>/status.
// Returns 0 if the process is gone or unreadable.
func readPPID(procPath string, pid uint32) uint32 {
	data, err := os.ReadFile(fmt.Sprintf("%s/%d/status", procPath, pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PPid:") {
			fields := strings.Fields(line)
			if len(fields) == 2 {
				var ppid uint32
				fmt.Sscanf(fields[1], "%d", &ppid)
				return ppid
			}
		}
	}
	return 0
}

// DiscoverExistingProcesses performs a one-time sweep of /proc to populate the tree
// with processes that started before Godshell.
func (t *ProcessTree) DiscoverExistingProcesses() {
	entries, err := os.ReadDir(t.ProcPath)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pidInt, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		pid := uint32(pidInt)

		t.mu.Lock()
		if _, ok := t.ByPID[pid]; ok {
			t.mu.Unlock()
			continue
		}
		t.mu.Unlock()

		// Metadata gathering (comm, ppid, bin)
		comm := readCmdline(t.ProcPath, pid, "")
		if comm == "" {
			commBytes, _ := os.ReadFile(fmt.Sprintf("%s/%d/comm", t.ProcPath, pid))
			comm = strings.TrimSpace(string(commBytes))
		}
		if comm == "" {
			comm = fmt.Sprintf("pid:%d", pid)
		}

		ppid := readPPID(t.ProcPath, pid)
		bin := readBinaryPath(t.ProcPath, pid)

		t.mu.Lock()
		if _, ok := t.ByPID[pid]; !ok {
			node := &ProcessNode{
				PID:            pid,
				PPID:           ppid,
				Comm:           comm,
				BinaryPath:     bin,
				Effects:        make(map[string]*Effect),
				StartTimestamp: time.Now(), // Estimated
				IsEnriched:     true,
			}
			t.ByPID[pid] = node

			// Wire up children
			if ppid > 0 {
				if parent, ok := t.ByPID[ppid]; ok {
					found := false
					for _, cpid := range parent.ChildrenPID {
						if cpid == pid {
							found = true
							break
						}
					}
					if !found {
						parent.ChildrenPID = append(parent.ChildrenPID, pid)
					}
				}
			}
		}
		t.mu.Unlock()
	}
}
