package ctxengine

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// FrozenSnapshot represents an immutable point-in-time capture of the system state.
// It is detached from the live BPF events and can safely be queried iteratively.
type FrozenSnapshot struct {
	Timestamp time.Time
	ByPID     map[uint32]*ProcessNode
	Ghosts    map[uint32]*ProcessNode
}

// TakeSnapshot deep copies the current live ProcessTree and returns a safe, frozen copy.
func (t *ProcessTree) TakeSnapshot() *FrozenSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	frozen := &FrozenSnapshot{
		Timestamp: time.Now(),
		ByPID:     make(map[uint32]*ProcessNode, len(t.ByPID)),
		Ghosts:    make(map[uint32]*ProcessNode, len(t.Ghosts)),
	}

	for pid, node := range t.ByPID {
		frozen.ByPID[pid] = node.Clone()
	}
	for pid, node := range t.Ghosts {
		frozen.Ghosts[pid] = node.Clone()
	}

	return frozen
}

// Inspect returns a deeply formatted text view of a single process and its immediate context.
func (fs *FrozenSnapshot) Inspect(pid uint32) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		// Check ghosts
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Error] Process %d not found in snapshot", pid)
		}
	}

	var b strings.Builder

	state := "ACTIVE"
	uptime := time.Since(p.StartTimestamp).Round(time.Millisecond)
	if !p.EndTimestamp.IsZero() {
		state = "EXITED"
		uptime = p.EndTimestamp.Sub(p.StartTimestamp).Round(time.Millisecond)
	}

	b.WriteString(fmt.Sprintf("[Process Inspect: %d] (Snapshot from %s)\n", p.PID, fs.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Command:  %s\n", p.Comm))
	if p.BinaryPath != "" {
		b.WriteString(fmt.Sprintf("Binary:   %s\n", p.BinaryPath))
	}

	stateStr := fmt.Sprintf("State:    %s (uptime: %v, started: %s", state, uptime, p.StartTimestamp.Format("15:04:05.000"))
	if !p.EndTimestamp.IsZero() {
		stateStr += fmt.Sprintf(", ended: %s", p.EndTimestamp.Format("15:04:05.000"))
	}
	b.WriteString(stateStr + ")\n")

	if p.CpuUsage > 0 || p.MemoryUsage > 0 {
		b.WriteString(fmt.Sprintf("CPU:      %.1f%%  |  Mem: %d KB\n", p.CpuUsage, p.MemoryUsage))
	}
	b.WriteString("\n")

	// Parent info
	parentStr := fmt.Sprintf("%d", p.PPID)
	if p.PPID > 0 {
		if parent, found := fs.ByPID[p.PPID]; found {
			parentStr = fmt.Sprintf("%s (%d)", parent.Comm, p.PPID)
		} else if parent, found := fs.Ghosts[p.PPID]; found {
			parentStr = fmt.Sprintf("%s (%d) [EXITED]", parent.Comm, p.PPID)
		}
	}
	b.WriteString(fmt.Sprintf("Parent:   %s\n", parentStr))

	// Children info
	if len(p.ChildrenPID) > 0 {
		b.WriteString("Children:\n")
		var childPIDs []uint32
		childPIDs = append(childPIDs, p.ChildrenPID...)
		sort.Slice(childPIDs, func(i, j int) bool { return childPIDs[i] < childPIDs[j] })
		for _, cpid := range childPIDs {
			if child, found := fs.ByPID[cpid]; found {
				b.WriteString(fmt.Sprintf("  - %s (%d) [ACTIVE]\n", child.Comm, cpid))
			} else if child, found := fs.Ghosts[cpid]; found {
				b.WriteString(fmt.Sprintf("  - %s (%d) [EXITED]\n", child.Comm, cpid))
			} else {
				b.WriteString(fmt.Sprintf("  - (%d) [UNKNOWN]\n", cpid))
			}
		}
	} else {
		b.WriteString("Children: None\n")
	}
	b.WriteString("\n")

	// Effects
	var netEffects []string
	var fileEffects []string

	var effectList []*Effect
	for _, eff := range p.Effects {
		effectList = append(effectList, eff)
	}
	sort.Slice(effectList, func(i, j int) bool {
		if effectList[i].Kind != effectList[j].Kind {
			return effectList[i].Kind < effectList[j].Kind
		}
		return effectList[i].Target < effectList[j].Target
	})

	for _, eff := range effectList {
		duration := time.Since(eff.Last).Round(time.Second)
		seenStr := fmt.Sprintf("%v ago", duration)
		if duration < time.Second {
			seenStr = "just now"
		}

		lastStr := eff.Last.Format("15:04:05")
		if eff.Kind == EffectConnect {
			if strings.HasPrefix(eff.Target, "unix:") {
				netEffects = append(netEffects, fmt.Sprintf("  - unix socket: %s (last: %s, seen %s, count %d)", eff.Target[5:], lastStr, seenStr, eff.Count))
			} else {
				netEffects = append(netEffects, fmt.Sprintf("  - connect: %s (last: %s, seen %s, count %d)", eff.Target, lastStr, seenStr, eff.Count))
			}
		} else if eff.Kind == EffectOpen {
			fileEffects = append(fileEffects, fmt.Sprintf("  - open: %s (last: %s, seen %s, count %d)", eff.Target, lastStr, seenStr, eff.Count))
		}
	}

	if len(netEffects) > 0 {
		b.WriteString("NETWORK EFFECTS:\n")
		b.WriteString(strings.Join(netEffects, "\n") + "\n\n")
	}

	if len(fileEffects) > 0 {
		b.WriteString("FILE EFFECTS:\n")
		b.WriteString(strings.Join(fileEffects, "\n") + "\n")
	}

	if len(netEffects) == 0 && len(fileEffects) == 0 {
		b.WriteString("No observed file or network effects.\n")
	}

	return b.String()
}

// ProcessFamily returns a text-formatted lineage tree for the given PID.
func (fs *FrozenSnapshot) ProcessFamily(pid uint32) string {
	// Helper to lookup processes, including dynamically resolving untracked ones
	untracked := make(map[uint32]*ProcessNode)

	// We declare childrenMap early so readUntracked can add to it
	childrenMap := make(map[uint32][]uint32)

	readUntracked := func(pid uint32) *ProcessNode {
		if p, ok := untracked[pid]; ok {
			return p
		}
		commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err != nil {
			return nil
		}
		comm := strings.TrimSpace(string(commBytes))

		statusBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
		var ppid uint32
		if err == nil {
			for _, line := range strings.Split(string(statusBytes), "\n") {
				if strings.HasPrefix(line, "PPid:") {
					fields := strings.Fields(line)
					if len(fields) == 2 {
						fmt.Sscanf(fields[1], "%d", &ppid)
					}
					break
				}
			}
		}

		node := &ProcessNode{
			PID:  pid,
			PPID: ppid,
			Comm: comm + " [UNTRACKED]",
		}
		untracked[pid] = node
		// Add this newly discovered node to its parent's children list so tree prints it
		childrenMap[ppid] = append(childrenMap[ppid], pid)
		return node
	}

	getNode := func(pid uint32) (*ProcessNode, bool) {
		if p, ok := fs.ByPID[pid]; ok {
			return p, true
		}
		if p, ok := fs.Ghosts[pid]; ok {
			return p, true
		}
		if p, ok := untracked[pid]; ok {
			return p, true
		}
		return nil, false
	}

	target, ok := getNode(pid)
	if !ok {
		target = readUntracked(pid)
		if target == nil {
			return fmt.Sprintf("[Error] Process %d not found in snapshot or /proc", pid)
		}
	}

	// 1. Build a robust children map from ALL processes in the snapshot
	for p_id, p := range fs.ByPID {
		childrenMap[p.PPID] = append(childrenMap[p.PPID], p_id)
	}
	for p_id, p := range fs.Ghosts {
		childrenMap[p.PPID] = append(childrenMap[p.PPID], p_id)
	}

	// 2. Walk up to find the root and record the ancestry chain
	root := target
	chain := make(map[uint32]bool)
	chain[target.PID] = true

	for {
		if root.PPID == 0 || root.PPID == 1 || root.PPID == 2 {
			break
		}
		parent, ok := getNode(root.PPID)
		if !ok {
			// Try to dynamically resolve it from /proc
			parent = readUntracked(root.PPID)
			if parent == nil {
				break
			}
		}
		root = parent
		chain[root.PID] = true
	}

	// 3. To handle missing children of target (e.g. clone without execve), Try probing /proc/<pid>/task/<pid>/children
	resolveChildren := func(pid uint32) {
		childBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/task/%d/children", pid, pid))
		if err == nil {
			for _, childParts := range strings.Fields(string(childBytes)) {
				var cpid uint32
				fmt.Sscanf(childParts, "%d", &cpid)
				if _, ok := getNode(cpid); !ok {
					// We found a child that is untracked! Read it.
					if childProc := readUntracked(cpid); childProc != nil {
						// readUntracked handles adding it to childrenMap
					}
				}
			}
		}
	}

	// Resolve untracked children for target and its direct ancestors to round out the family view
	for ancestorPID := range chain {
		resolveChildren(ancestorPID)
	}

	// Sort children for deterministic output
	for _, children := range childrenMap {
		sort.Slice(children, func(i, j int) bool { return children[i] < children[j] })
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Process Family: %d] (Snapshot from %s)\n", pid, fs.Timestamp.Format(time.RFC3339)))

	// Helper for recursive printing
	var printTree func(node *ProcessNode, prefix string, isLast bool, inSubtree bool)
	printTree = func(node *ProcessNode, prefix string, isLast bool, inSubtree bool) {
		marker := "├──"
		if isLast {
			marker = "└──"
		}

		nameDisplay := node.Comm
		if node.PID == target.PID {
			nameDisplay += " [TARGET]"
		}

		b.WriteString(fmt.Sprintf("%s%s %s (%d)\n", prefix, marker, nameDisplay, node.PID))

		childPrefix := prefix + "│    "
		if isLast {
			childPrefix = prefix + "     "
		}

		var toPrint []*ProcessNode
		for _, cpid := range childrenMap[node.PID] {
			child, found := getNode(cpid)
			if found {
				// Print if:
				// 1. The child is in the direct ancestry chain (leads to target)
				// 2. The node being evaluated is the target's parent (so we print siblings)
				// 3. We are already in the target's subtree
				// 4. The node being evaluated IS the target (so we print its direct children)
				if chain[cpid] || node.PID == target.PPID || inSubtree || node.PID == target.PID {
					// Ensure no infinite recursion from bad PPID resolution
					if cpid != node.PID {
						toPrint = append(toPrint, child)
					}
				}
			}
		}

		nextInSubtree := inSubtree || node.PID == target.PID
		for i, child := range toPrint {
			printTree(child, childPrefix, i == len(toPrint)-1, nextInSubtree)
		}
	}

	nameDisplay := root.Comm
	if root.PID == target.PID {
		nameDisplay += " [TARGET]"
	}
	b.WriteString(fmt.Sprintf("%s (%d)\n", nameDisplay, root.PID))

	var rootToPrint []*ProcessNode
	for _, cpid := range childrenMap[root.PID] {
		child, found := getNode(cpid)
		if found {
			if chain[cpid] || root.PID == target.PPID || root.PID == target.PID {
				if cpid != root.PID {
					rootToPrint = append(rootToPrint, child)
				}
			}
		}
	}

	nextInSubtreeRoot := root.PID == target.PID
	for i, child := range rootToPrint {
		printTree(child, "  ", i == len(rootToPrint)-1, nextInSubtreeRoot)
	}

	return b.String()
}

// Search returns all processes matching the pattern (name, binary, or effect).
func (fs *FrozenSnapshot) Search(pattern string) string {
	lowerPattern := strings.ToLower(pattern)
	var matches []*ProcessNode

	matchNode := func(p *ProcessNode) bool {
		if strings.Contains(strings.ToLower(p.Comm), lowerPattern) {
			return true
		}
		if strings.Contains(strings.ToLower(p.BinaryPath), lowerPattern) {
			return true
		}
		for _, eff := range p.Effects {
			kindStr := "open"
			if eff.Kind == EffectConnect {
				kindStr = "connect"
			}
			if strings.Contains(strings.ToLower(eff.Target), lowerPattern) || strings.Contains(kindStr, lowerPattern) {
				return true
			}
		}
		return false
	}

	for _, p := range fs.ByPID {
		if matchNode(p) {
			matches = append(matches, p)
		}
	}
	for _, p := range fs.Ghosts {
		if matchNode(p) {
			matches = append(matches, p)
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("[Search Results for %q] (Snapshot from %s)\nNo matching processes found.", pattern, fs.Timestamp.Format(time.RFC3339))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Search Results for %q] (Snapshot from %s)\n", pattern, fs.Timestamp.Format(time.RFC3339)))

	for _, p := range matches {
		hitEffect := false
		for _, eff := range p.Effects {
			kindStr := "open"
			if eff.Kind == EffectConnect {
				kindStr = "connect"
			}

			// Match either the target or the kind (e.g. searching for "connect")
			if strings.Contains(strings.ToLower(eff.Target), lowerPattern) || strings.Contains(kindStr, lowerPattern) {
				lastStr := eff.Last.Format("15:04:05")
				b.WriteString(fmt.Sprintf("[MATCH] %s (%d) %s: %s (last: %s, count %d)\n", p.Comm, p.PID, kindStr, eff.Target, lastStr, eff.Count))
				hitEffect = true
			}
		}
		// If it only matched via Comm or BinaryPath and not an explicit effect
		if !hitEffect {
			b.WriteString(fmt.Sprintf("[MATCH] %s (%d) [Matched via process name or binary path]\n", p.Comm, p.PID))
		}
	}

	return b.String()
}

// ReadProcessMaps provides a condensed summary of a process's memory layout from /proc/[pid]/maps.
// It groups contiguous regions that map to the same backing file/pseudo-path,
// aggregating the permissions (e.g. if one segment is r--p and the next is r-xp, the aggregate shows 'rx').
func (fs *FrozenSnapshot) ReadProcessMaps(pid uint32) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Process Maps: %d] Process not found in snapshot.", pid)
		}
	}

	mapsPath := fmt.Sprintf("/proc/%d/maps", pid)
	content, err := os.ReadFile(mapsPath)
	if err != nil {
		return fmt.Sprintf("[Process Maps: %d] (%s)\nError reading maps: %v", pid, p.Comm, err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Process Maps: %d] (%s) (Snapshot from %s)\n", pid, p.Comm, fs.Timestamp.Format(time.RFC3339)))

	if p.BinaryPath != "" {
		b.WriteString(fmt.Sprintf("Binary: %s\n", p.BinaryPath))
	}
	b.WriteString("\n")

	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	type mapSummary struct {
		startAddr string
		endAddr   string
		perms     map[rune]bool
		path      string
		sizeBytes uint64
	}

	var summaries []mapSummary
	var current *mapSummary

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// address   perms offset  dev   inode   pathname
		// 00400000-00452000 r-xp 00000000 08:02 173521      /usr/bin/dbus-daemon
		addrs := strings.Split(fields[0], "-")
		if len(addrs) != 2 {
			continue
		}

		var start, end uint64
		fmt.Sscanf(addrs[0], "%x", &start)
		fmt.Sscanf(addrs[1], "%x", &end)
		size := end - start

		permsStr := fields[1]
		path := "[anonymous]"
		if len(fields) >= 6 {
			path = strings.Join(fields[5:], " ") // handle paths with spaces
		}

		if current != nil && current.path == path {
			// Extend current region
			current.endAddr = addrs[1]
			current.sizeBytes += size
			for _, p := range permsStr {
				if p != '-' && p != 'p' && p != 's' { // Keep r,w,x
					current.perms[p] = true
				}
			}
		} else {
			// Save previous and start new
			if current != nil {
				summaries = append(summaries, *current)
			}

			pMap := make(map[rune]bool)
			for _, p := range permsStr {
				if p != '-' && p != 'p' && p != 's' {
					pMap[p] = true
				}
			}

			current = &mapSummary{
				startAddr: addrs[0],
				endAddr:   addrs[1],
				perms:     pMap,
				path:      path,
				sizeBytes: size,
			}
		}
	}

	if current != nil {
		summaries = append(summaries, *current)
	}

	// Formatting
	b.WriteString(fmt.Sprintf("%-33s %-6s %-10s %s\n", "ADDRESS RANGE", "PERMS", "SIZE", "PATH"))
	b.WriteString(strings.Repeat("-", 80) + "\n")

	for _, s := range summaries {
		// Construct perm string: 'r', 'w', 'x' or '-'
		pStr := ""
		if s.perms['r'] {
			pStr += "r"
		} else {
			pStr += "-"
		}
		if s.perms['w'] {
			pStr += "w"
		} else {
			pStr += "-"
		}
		if s.perms['x'] {
			pStr += "x"
		} else {
			pStr += "-"
		}

		var sizeStr string
		if s.sizeBytes > 1024*1024 {
			sizeStr = fmt.Sprintf("%d MB", s.sizeBytes/(1024*1024))
		} else if s.sizeBytes > 1024 {
			sizeStr = fmt.Sprintf("%d KB", s.sizeBytes/1024)
		} else {
			sizeStr = fmt.Sprintf("%d B", s.sizeBytes)
		}

		rangeStr := fmt.Sprintf("%s-%s", s.startAddr, s.endAddr)

		// Truncate very long paths for UI neatness, keep the end usually (e.g. library names)
		displayPath := s.path
		if len(displayPath) > 60 {
			displayPath = "..." + displayPath[len(displayPath)-57:]
		}

		b.WriteString(fmt.Sprintf("%-33s %-6s %-10s %s\n", rangeStr, pStr, sizeStr, displayPath))
	}

	return b.String()
}

// GetLinkedLibraries resolves the dynamically linked shared objects for a given process's binary
// by wrapping the standard `ldd` command.
func (fs *FrozenSnapshot) GetLinkedLibraries(pid uint32) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Linked Libraries: %d] Process not found in snapshot.", pid)
		}
	}

	if p.BinaryPath == "" {
		return fmt.Sprintf("[Linked Libraries: %d] (%s)\nNo binary path recorded. Cannot run ldd.", pid, p.Comm)
	}

	cmd := exec.Command("ldd", p.BinaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("[Linked Libraries: %d] (%s) %s\nError running ldd: %v\nOutput: %s",
			pid, p.Comm, p.BinaryPath, err, string(output))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Linked Libraries: %d] (%s) (Snapshot from %s)\n", pid, p.Comm, fs.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Binary: %s\n\n", p.BinaryPath))

	b.WriteString(fmt.Sprintf("%-30s %s\n", "LIBRARY", "RESOLVED PATH"))
	b.WriteString(strings.Repeat("-", 80) + "\n")

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Example ldd output lines:
		// linux-vdso.so.1 (0x00007ffe345bd000)
		// libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f5d4a600000)
		// /lib64/ld-linux-x86-64.so.2 (0x00007f5d4a83b000)

		parts := strings.Split(line, " => ")
		if len(parts) == 2 {
			libName := strings.TrimSpace(parts[0])
			pathPart := strings.TrimSpace(parts[1])

			// Extract just the path before the hex address
			pathOnly := strings.Split(pathPart, " (")[0]

			b.WriteString(fmt.Sprintf("%-30s %s\n", libName, pathOnly))
		} else {
			// Lines without " => " usually mean it's an absolute path already or vdso
			pathOnly := strings.Split(line, " (")[0]
			b.WriteString(fmt.Sprintf("%-30s [builtin or absolute]\n", pathOnly))
		}
	}

	return b.String()
}

// TraceSyscalls runs a scoped strace against the target process for the specified duration and returns the summary.
func (fs *FrozenSnapshot) TraceSyscalls(pid uint32, durationSeconds int) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Trace Syscalls: %d] Process not found in snapshot.", pid)
		}
	}

	if durationSeconds <= 0 {
		durationSeconds = 5 // Default
	}

	cmd := exec.Command("strace", "-c", "-p", fmt.Sprintf("%d", pid))

	// strace prints its summary to stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("[Trace Syscalls: %d] (%s)\nFailed to start strace: %v\nMake sure you have privileges and strace is installed.", pid, p.Comm, err)
	}

	// We can't use context.WithTimeout because it sends SIGKILL which ruins the -c summary.
	// Instead, send SIGINT explicitly after the duration.
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(durationSeconds) * time.Second):
		// Interrupt strace to trigger the summary output
		if cmd.Process != nil {
			cmd.Process.Signal(os.Interrupt)
		}
		<-done // Wait for it to exit
	case err := <-done:
		if err != nil {
			return fmt.Sprintf("[Trace Syscalls: %d] (%s)\nstrace exited early with error: %v\nOutput: %s", pid, p.Comm, err, stderr.String())
		}
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Trace Syscalls: %d] (%s) (Snapshot from %s)\n", pid, p.Comm, fs.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Duration: %d seconds\n\n", durationSeconds))

	outStr := stderr.String()
	if outStr == "" {
		outStr = "No syscalls captured or process unreachable by strace.\n"
	}
	b.WriteString(outStr)

	return b.String()
}

// ReadFile provides a safe reading mechanism for the LLM to read chunks of files
// that may have been accessed by processes. It attempts to read as UTF-8, but
// falls back to a hexdump if unprintable characters dominate.
func (fs *FrozenSnapshot) ReadFile(path string, offset int64, limit int64) string {
	if limit <= 0 {
		limit = 4096 // 4KB default
	}
	if limit > 1024*1024 { // Max 1MB chunks to prevent memory explosion
		limit = 1024 * 1024
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("[Read File: %s] Error opening file: %v", path, err)
	}
	defer file.Close()

	if offset > 0 {
		if _, err := file.Seek(offset, 0); err != nil {
			return fmt.Sprintf("[Read File: %s] Error seeking to offset %d: %v", path, offset, err)
		}
	}

	buf := make([]byte, limit)
	n, err := file.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return fmt.Sprintf("[Read File: %s] Error reading file: %v", path, err)
	}
	if n == 0 {
		return fmt.Sprintf("[Read File: %s]\n(Empty or EOF reached at offset %d)", path, offset)
	}

	buf = buf[:n]

	// Check if the output is mostly printable text
	unprintable := 0
	for _, b := range buf {
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			unprintable++
		}
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[Read File: %s] (Offset: %d, Read: %d bytes)\n\n", path, offset, n))

	// If > 10% unprintable, use hex dump
	if float64(unprintable)/float64(n) > 0.10 {
		builder.WriteString("(Binary content detected. Rendering Hex Dump)\n")
		builder.WriteString(hex.Dump(buf))
	} else {
		builder.WriteString(string(buf))
	}

	return builder.String()
}

// ReadMemory allows reading raw process memory from /proc/[pid]/mem.
// Given the binary nature of raw memory, it exclusively outputs hex dumps.
func (fs *FrozenSnapshot) ReadMemory(pid uint32, address uint64, size int64) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Read Memory: %d] Process not found in snapshot.", pid)
		}
	}

	if size <= 0 {
		size = 1024 // 1KB default
	}
	if size > 1024*1024 { // Cap at 1MB
		size = 1024 * 1024
	}

	memPath := fmt.Sprintf("/proc/%d/mem", pid)
	file, err := os.Open(memPath)
	if err != nil {
		return fmt.Sprintf("[Read Memory: %d] (%s)\nError opening %s: %v\n(Check permissions, or process may have died/be protected)", pid, p.Comm, memPath, err)
	}
	defer file.Close()

	if _, err := file.Seek(int64(address), 0); err != nil {
		return fmt.Sprintf("[Read Memory: %d] (%s)\nError seeking to address 0x%x: %v\n(Address may be unmapped or invalid)", pid, p.Comm, address, err)
	}

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return fmt.Sprintf("[Read Memory: %d] (%s)\nError reading memory at 0x%x: %v", pid, p.Comm, address, err)
	}

	if n == 0 {
		return fmt.Sprintf("[Read Memory: %d] (%s)\n(0 bytes read at address 0x%x. Unmapped region?)", pid, p.Comm, address)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[Read Memory: %d] (%s) (Address: 0x%x, Read: %d bytes)\n\n", pid, p.Comm, address, n))
	builder.WriteString(hex.Dump(buf[:n]))

	return builder.String()
}
