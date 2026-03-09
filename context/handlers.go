package ctxengine

import (
	"fmt"
	"os"
	"strings"
	"time"

	"godshell/observer"
)

func NewProcessTree() *ProcessTree {
	return &ProcessTree{
		ByPID:  make(map[uint32]*ProcessNode),
		Ghosts: make(map[uint32]*ProcessNode),
	}
}

// HandleEvent routes an observer event to the correct handler.
func (t *ProcessTree) HandleEvent(e observer.Event) {
	switch e.Type {
	case observer.EventExec:
		t.handleExec(e)
	case observer.EventOpen:
		t.handleOpen(e)
	case observer.EventExit:
		t.handleExit(e)
	case observer.EventConnect:
		t.handleConnect(e)
	}
}

// handleExec creates a new ProcessNode from an execve event.
func (t *ProcessTree) handleExec(e observer.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	proc := &ProcessNode{
		PID:            e.Pid,
		BinaryPath:     e.PathStr(),
		Comm:           e.CommStr(),
		Effects:        make(map[string]*Effect),
		StartTimestamp: time.Now(),
	}

	// Read PPID from /proc/<pid>/status
	proc.PPID = readPPID(e.Pid)

	// Link to parent's children list if parent is tracked
	if parent, ok := t.ByPID[proc.PPID]; ok {
		parent.ChildrenPID = append(parent.ChildrenPID, proc.PID)
	}

	t.ByPID[proc.PID] = proc
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

	// Enrich from /proc/<pid>/net/tcp — best-effort, may fail for short-lived procs
	detail := enrichConnect(e.Pid)
	if detail == nil {
		detail = &ConnectDetail{}
	}

	// Key by remote ip:port, unix socket, or fallback timestamp for dedup
	var key string
	if detail.IP != "" {
		key = fmt.Sprintf("%s:%d", detail.IP, detail.Port)
		if detail.Domain != "" {
			key = fmt.Sprintf("%s (%s)", key, detail.Domain)
		}
	} else if detail.UnixSocket != "" {
		key = fmt.Sprintf("unix:%s", detail.UnixSocket)
	} else {
		key = fmt.Sprintf("connect:%d:%d", e.Pid, time.Now().UnixNano())
	}

	if eff, ok := proc.Effects[key]; ok {
		eff.Count++
		eff.Last = time.Now()
		// Update bytes if we have newer data
		if d, ok := eff.Detail.(ConnectDetail); ok {
			d.BytesSent += detail.BytesSent
			d.BytesRecv += detail.BytesRecv
			eff.Detail = d
		}
	} else {
		proc.Effects[key] = &Effect{
			Kind:   EffectConnect,
			Target: key,
			Count:  1,
			First:  time.Now(),
			Last:   time.Now(),
			Detail: *detail,
		}
	}
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
		BinaryPath:     readBinaryPath(e.Pid),
		Effects:        make(map[string]*Effect),
		StartTimestamp: time.Now(),
		PPID:           readPPID(e.Pid),
	}
	t.ByPID[e.Pid] = proc
	return proc
}

// readBinaryPath reads the exe symlink from /proc/<pid>/exe.
// Returns empty string if the process is gone or unreadable.
func readBinaryPath(pid uint32) string {
	target, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return ""
	}
	return target
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

// readPPID reads the parent PID from /proc/<pid>/status.
// Returns 0 if the process is gone or unreadable.
func readPPID(pid uint32) uint32 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
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
