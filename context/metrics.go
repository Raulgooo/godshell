package ctxengine

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// prevCPU stores the previous CPU tick values for delta calculation.
type prevCPU struct {
	utime  uint64
	stime  uint64
	sample time.Time
}

// RefreshMetrics runs in a background goroutine, polling /proc/<pid>/stat
// and /proc/<pid>/status for CPU and memory metrics of all tracked processes.
func (t *ProcessTree) RefreshMetrics(interval time.Duration) {
	prev := make(map[uint32]*prevCPU)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		t.mu.Lock()
		for pid, proc := range t.ByPID {
			cpu := readCPUTicks(pid)
			mem := readMemoryRSS(pid)

			proc.MemoryUsage = mem

			// Calculate CPU percentage from tick delta
			if cpu != nil {
				if p, ok := prev[pid]; ok {
					dt := time.Since(p.sample).Seconds()
					if dt > 0 {
						tickDelta := float64((cpu.utime - p.utime) + (cpu.stime - p.stime))
						// Clock ticks per second (typically 100 on Linux)
						proc.CpuUsage = (tickDelta / 100.0 / dt) * 100.0
					}
				}
				prev[pid] = cpu
			}
		}

		// Clean up prev entries for PIDs no longer tracked
		for pid := range prev {
			if _, ok := t.ByPID[pid]; !ok {
				delete(prev, pid)
			}
		}
		t.mu.Unlock()
	}
}

// readCPUTicks reads utime and stime from /proc/<pid>/stat.
// Format: pid (comm) state ppid ... utime(14) stime(15) ...
func readCPUTicks(pid uint32) *prevCPU {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return nil
	}

	// Find the end of (comm) — it can contain spaces and parens
	s := string(data)
	closeParen := strings.LastIndex(s, ")")
	if closeParen < 0 || closeParen+2 >= len(s) {
		return nil
	}

	// Fields after (comm) start at index 2 (state=0-indexed after comm)
	fields := strings.Fields(s[closeParen+2:])
	// utime is field index 11 (= position 14 in full stat minus 3 for pid, comm, state)
	// stime is field index 12
	if len(fields) < 13 {
		return nil
	}

	utime, err1 := strconv.ParseUint(fields[11], 10, 64)
	stime, err2 := strconv.ParseUint(fields[12], 10, 64)
	if err1 != nil || err2 != nil {
		return nil
	}

	return &prevCPU{
		utime:  utime,
		stime:  stime,
		sample: time.Now(),
	}
}

// readMemoryRSS reads VmRSS from /proc/<pid>/status in KB.
func readMemoryRSS(pid uint32) uint64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					return val // KB
				}
			}
		}
	}
	return 0
}
