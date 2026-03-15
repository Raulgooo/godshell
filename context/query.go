package ctxengine

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

// FrozenSnapshot represents an immutable point-in-time capture of the system state.
// It is detached from the live BPF events and can safely be queried iteratively.
type FrozenSnapshot struct {
	Timestamp time.Time
	ByPID     map[uint32]*ProcessNode
	Ghosts    map[uint32]*ProcessNode
	StracePath string
}

// TakeSnapshot deep copies the current live ProcessTree and returns a safe, frozen copy.
func (t *ProcessTree) TakeSnapshot() *FrozenSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()

	frozen := &FrozenSnapshot{
		Timestamp: time.Now(),
		ByPID:     make(map[uint32]*ProcessNode, len(t.ByPID)),
		Ghosts:    make(map[uint32]*ProcessNode, len(t.Ghosts)),
		StracePath: t.StracePath,
	}

	for pid, node := range t.ByPID {
		if !node.IsEnriched {
			// Slow enrichment while holding RLock is bad, but TakeSnapshot is the only place.
			// Better: Upgrade to Write Lock for enrichment, then downgrade?
			// For simplicity and since TakeSnapshot is "poll" time, we just do it.
			// To be safe, we release lock, enrich, and re-acquire? No, TakeSnapshot needs consistency.
			// Actually handlers.go functions use t.mu.Lock(), so we MUST release RLock first to avoid deadlock.
			t.mu.RUnlock()
			t.EnrichProcessMetadata(pid)
			// Enrich connection effects
			for key, eff := range node.Effects {
				if eff.Kind == EffectConnect {
					t.EnrichConnectionMetadata(pid, key, nil)
				}
			}
			t.mu.RLock()
			if n, ok := t.ByPID[pid]; ok {
				n.IsEnriched = true
			}
		}
		// Re-check node exists after lock dance
		if n, ok := t.ByPID[pid]; ok {
			frozen.ByPID[pid] = n.Clone()
		}
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
		permsStr  string
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

		// Bugfix: Only collapse if BOTH path AND permissions match.
		// This prevents merging r-x and rw- into a fake rwx block.
		if current != nil && current.path == path && current.permsStr == permsStr {
			// Extend current region
			current.endAddr = addrs[1]
			current.sizeBytes += size
		} else {
			// Save previous and start new
			if current != nil {
				summaries = append(summaries, *current)
			}

			current = &mapSummary{
				startAddr: addrs[0],
				endAddr:   addrs[1],
				permsStr:  permsStr,
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
		pStr := s.permsStr

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

// HashBinary computes the SHA256 hash of the executable associated with a process.
func (fs *FrozenSnapshot) HashBinary(pid uint32) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Errorf("[Hash Binary: %d] Process not found in snapshot.", pid).Error()
		}
	}

	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	// If process is a ghost (or permission denied), try its recorded BinaryPath
	if _, err := os.Stat(exePath); os.IsNotExist(err) || err != nil {
		if p.BinaryPath != "" {
			exePath = p.BinaryPath
		} else {
			return fmt.Sprintf("[Hash Binary: %d] (%s)\nCannot find executable path to hash. Process may have exited and no binary path was recorded.", pid, p.Comm)
		}
	}

	// Double check if we can read the file we resolved
	file, err := os.Open(exePath)
	if err != nil {
		return fmt.Sprintf("[Hash Binary: %d] (%s)\nError opening executable %s: %v", pid, p.Comm, exePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Sprintf("[Hash Binary: %d] (%s)\nError hashing executable %s: %v", pid, p.Comm, exePath, err)
	}

	hashStr := hex.EncodeToString(hash.Sum(nil))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Hash Binary: %d] (%s) (Snapshot from %s)\n", pid, p.Comm, fs.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Binary: %s\n", exePath))
	b.WriteString(fmt.Sprintf("SHA256: %s\n", hashStr))

	return b.String()
}

// ExtractStrings runs the `strings` command on a binary to extract printable characters.
func (fs *FrozenSnapshot) ExtractStrings(path string, minLength int) string {
	if minLength <= 0 {
		minLength = 8 // Default min length
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) || err != nil {
		return fmt.Sprintf("[Extract Strings: %s]\nCannot access file: %v", path, err)
	}

	cmd := exec.Command("strings", "-a", "-n", fmt.Sprintf("%d", minLength), path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("[Extract Strings: %s]\nError running strings: %v\nOutput: %s", path, err, string(output))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Extract Strings: %s] (Min Length: %d) (Snapshot from %s)\n\n", path, minLength, fs.Timestamp.Format(time.RFC3339)))

	// Limit output to prevent LLM context explosion.
	lines := strings.Split(string(output), "\n")
	limit := 200
	if len(lines) > limit {
		b.WriteString(strings.Join(lines[:limit], "\n"))
		b.WriteString(fmt.Sprintf("\n... (Output truncated, %d total strings found) ...\n", len(lines)))
	} else {
		b.WriteString(string(output))
	}

	return b.String()
}

// ReadShellHistory retrieves the last N lines of a user's shell history.
func (fs *FrozenSnapshot) ReadShellHistory(username string, limit int) string {
	if limit <= 0 {
		limit = 50 // default to last 50 lines
	}

	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Sprintf("[Read Shell History: %s]\nError looking up user: %v", username, err)
	}

	historyFiles := []string{
		filepath.Join(u.HomeDir, ".zsh_history"),
		filepath.Join(u.HomeDir, ".bash_history"),
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Read Shell History: %s] (Limit: %d lines) (Snapshot from %s)\n\n", username, limit, fs.Timestamp.Format(time.RFC3339)))

	foundAny := false
	for _, historyFile := range historyFiles {
		content, err := os.ReadFile(historyFile)
		if err != nil {
			continue // skip if not found or cannot read
		}

		foundAny = true
		b.WriteString(fmt.Sprintf("--- From: %s ---\n", historyFile))

		lines := strings.Split(string(content), "\n")
		// Remove empty trailing line if present
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		startIdx := 0
		if len(lines) > limit {
			startIdx = len(lines) - limit
		}

		for _, line := range lines[startIdx:] {
			// Basic cleanup of zsh timestamp format: ": 1690000000:0;command"
			if strings.HasPrefix(line, ": ") {
				if parts := strings.SplitN(line, ";", 2); len(parts) == 2 {
					line = parts[1]
				}
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	if !foundAny {
		b.WriteString("Could not read any standard shell history files (.zsh_history, .bash_history) for the user.\nCheck permissions or if the user uses a different shell.\n")
	}

	return b.String()
}

// NetworkState extracts active connections and their detailed state directly from /proc/<pid>/net/tcp and tcp6.
func (fs *FrozenSnapshot) NetworkState(pid uint32) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Network State: %d] Process not found in snapshot.", pid)
		}
	}

	netFiles := []string{
		fmt.Sprintf("/proc/%d/net/tcp", pid),
		fmt.Sprintf("/proc/%d/net/tcp6", pid),
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Network State: %d] (%s) (Snapshot from %s)\n\n", pid, p.Comm, fs.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("%-25s %-25s %s\n", "LOCAL_ADDRESS", "REMOTE_ADDRESS", "STATE"))
	b.WriteString(strings.Repeat("-", 60) + "\n")

	// Helper to parse hex IP and Port
	parseHexIPPort := func(hexStr string, isV6 bool) string {
		parts := strings.Split(hexStr, ":")
		if len(parts) != 2 {
			return hexStr
		}
		ipHex := parts[0]
		portHex := parts[1]

		var ipStr string
		if isV6 {
			// Basic formatting for IPv6 (not fully compliant, but enough for readable logs)
			// Reverse every 4 hex chars because of endianness in /proc/net/tcp6
			ipStr = ipHex // Keep raw hex for now for simplicity, it's a known limitation
		} else {
			if len(ipHex) == 8 {
				ipParsed := []byte{0, 0, 0, 0}
				fmt.Sscanf(ipHex, "%02x%02x%02x%02x", &ipParsed[3], &ipParsed[2], &ipParsed[1], &ipParsed[0])
				ipStr = fmt.Sprintf("%d.%d.%d.%d", ipParsed[0], ipParsed[1], ipParsed[2], ipParsed[3])
			} else {
				ipStr = ipHex
			}
		}

		var port int
		fmt.Sscanf(portHex, "%X", &port)

		return fmt.Sprintf("%s:%d", ipStr, port)
	}

	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}

	connCount := 0
	for _, netFile := range netFiles {
		isV6 := strings.HasSuffix(netFile, "6")
		content, err := os.ReadFile(netFile)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		// Skip header line
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}

			localAddr := parseHexIPPort(fields[1], isV6)
			remoteAddr := parseHexIPPort(fields[2], isV6)
			stateHex := fields[3]
			stateStr := states[stateHex]
			if stateStr == "" {
				stateStr = stateHex
			}

			b.WriteString(fmt.Sprintf("%-25s %-25s %s\n", localAddr, remoteAddr, stateStr))
			connCount++
		}
	}

	if connCount == 0 {
		b.WriteString("No active TCP connections found in /proc/<pid>/net/tcp.\n")
	}

	return b.String()
}

// ReadEnviron parses and returns the environment variables for a process.
func (fs *FrozenSnapshot) ReadEnviron(pid uint32) string {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return fmt.Sprintf("[Read Environ: %d] Process not found in snapshot.", pid)
		}
	}

	envPath := fmt.Sprintf("/proc/%d/environ", pid)
	content, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Sprintf("[Read Environ: %d] (%s)\nError reading %s: %v\n(Check permissions, or process may have exited)", pid, p.Comm, envPath, err)
	}

	if len(content) == 0 {
		return fmt.Sprintf("[Read Environ: %d] (%s)\nEnvironment is empty.", pid, p.Comm)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[Read Environ: %d] (%s) (Snapshot from %s)\n\n", pid, p.Comm, fs.Timestamp.Format(time.RFC3339)))

	// /proc/<pid>/environ is null-byte separated
	vars := bytes.Split(content, []byte{0})

	// Convert to string slice and sort for readability
	var strVars []string
	for _, v := range vars {
		s := string(v)
		if s != "" { // final split might be empty
			strVars = append(strVars, s)
		}
	}
	sort.Strings(strVars)

	for _, v := range strVars {
		b.WriteString(v)
		b.WriteString("\n")
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

	cmd := exec.Command(fs.StracePath, "-c", "-p", fmt.Sprintf("%d", pid))

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

	data, err := fs.ReadMemoryRaw(pid, address, size)
	if err != nil {
		return fmt.Sprintf("[Read Memory: %d] (%s)\nError reading memory at 0x%x: %v\n(Check permissions, or process may have died/be protected)", pid, p.Comm, address, err)
	}

	if len(data) == 0 {
		return fmt.Sprintf("[Read Memory: %d] (%s)\n(0 bytes read at address 0x%x. Unmapped region?)", pid, p.Comm, address)
	}

	printable := 0
	for _, b := range data {
		if b >= 32 && b <= 126 {
			printable++
		}
	}
	pct := (printable * 100) / len(data)

	return fmt.Sprintf("PID %d @ 0x%x · %d bytes · %d%% printable\n\n%s",
		pid, address, len(data), pct, hex.Dump(data))
}

// ptraceAttach attaches to the process. Note: requires CAP_SYS_PTRACE or root.
func ptraceAttach(pid uint32) error {
	runtime.LockOSThread()
	return syscall.PtraceAttach(int(pid))
}

// ptraceDetach detaches from the process.
func ptraceDetach(pid uint32) error {
	defer runtime.UnlockOSThread()
	return syscall.PtraceDetach(int(pid))
}

// ReadMemoryRaw is a helper that returns raw bytes or an error.
// It uses ptrace to ensure consistent access via /proc/pid/mem.
func (fs *FrozenSnapshot) ReadMemoryRaw(pid uint32, address uint64, size int64) ([]byte, error) {
	if err := ptraceAttach(pid); err != nil {
		return nil, fmt.Errorf("ptrace attach failed: %w", err)
	}
	// Wait for process to stop
	var status syscall.WaitStatus
	if _, err := syscall.Wait4(int(pid), &status, 0, nil); err != nil {
		ptraceDetach(pid)
		return nil, fmt.Errorf("waitpid failed: %w", err)
	}

	defer ptraceDetach(pid)
	return fs.ReadMemoryRawLocked(pid, address, size)
}

// ReadMemoryRawLocked reads memory from /proc/pid/mem assuming ptrace is already attached.
func (fs *FrozenSnapshot) ReadMemoryRawLocked(pid uint32, address uint64, size int64) ([]byte, error) {
	memPath := fmt.Sprintf("/proc/%d/mem", pid)
	file, err := os.Open(memPath)
	if err != nil {
		return nil, fmt.Errorf("open /proc/mem: %w", err)
	}
	defer file.Close()

	if _, err := file.Seek(int64(address), 0); err != nil {
		return nil, fmt.Errorf("seek error: %w", err)
	}

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read error: %w", err)
	}
	return buf[:n], nil
}

// ListHeapRegions returns a JSON summary of scan-worthy memory regions for a process.
func (fs *FrozenSnapshot) ListHeapRegions(pid uint32) (string, error) {
	if _, ok := fs.ByPID[pid]; !ok {
		if _, ok := fs.Ghosts[pid]; !ok {
			return "", fmt.Errorf("process %d not found", pid)
		}
	}

	mapsPath := fmt.Sprintf("/proc/%d/maps", pid)
	content, err := os.ReadFile(mapsPath)
	if err != nil {
		return "", fmt.Errorf("error reading maps: %w", err)
	}

	var regions []RegionSummary
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		perms := fields[1]
		// Consider regions that are private AND either writable or executable
		if !strings.HasSuffix(perms, "p") || (!strings.Contains(perms, "w") && !strings.Contains(perms, "x")) {
			continue
		}

		addrRange := strings.Split(fields[0], "-")
		if len(addrRange) != 2 {
			continue
		}
		var start, end uint64
		fmt.Sscanf(addrRange[0], "%x", &start)
		fmt.Sscanf(addrRange[1], "%x", &end)

		label := ""
		if len(fields) >= 6 {
			label = strings.Join(fields[5:], " ")
		}
		if label == "" {
			label = "[anon]"
		}

		regions = append(regions, RegionSummary{
			StartAddr: start,
			EndAddr:   end,
			Size:      int64(end - start),
			Perms:     perms,
			Label:     label,
		})
	}

	data, err := json.MarshalIndent(regions, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadRegion reads a specific chunk of memory and returns it in a requested format.
func (fs *FrozenSnapshot) ReadRegion(pid uint32, startAddr uint64, size int64, encoding string) (string, error) {
	data, err := fs.ReadMemoryRaw(pid, startAddr, size)
	if err != nil {
		return "", err
	}

	switch encoding {
	case "hex":
		return hex.Dump(data), nil
	case "utf8":
		return string(data), nil
	case "strings":
		// Simple strings implementation
		var result []string
		re := regexp.MustCompile(`[\w\.\-/]{4,}`)
		matches := re.FindAll(data, -1)
		for _, m := range matches {
			result = append(result, string(m))
		}
		return strings.Join(result, "\n"), nil
	default:
		return hex.Dump(data), nil
	}
}

// ScanHeap identifies all suspicious regions and scans them, returning a structured JSON result.
func (fs *FrozenSnapshot) ScanHeap(pid uint32, mode string) (string, error) {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return "", fmt.Errorf("process %d not found", pid)
		}
	}

	result := HeapScanResult{
		PID:      pid,
		Comm:     p.Comm,
		Regions:  []RegionSummary{},
		Findings: []Finding{},
	}

	regionsRaw, err := fs.ListHeapRegions(pid)
	if err != nil {
		result.Error = err.Error()
		b, _ := json.Marshal(result)
		return string(b), nil
	}

	var regions []RegionSummary
	json.Unmarshal([]byte(regionsRaw), &regions)

	// Filter regions for ScanHeap (exclude obvious system libraries, keep anon)
	var filtered []RegionSummary
	for _, reg := range regions {
		isLib := strings.HasPrefix(reg.Label, "/usr/lib") || strings.HasPrefix(reg.Label, "/lib")
		isBinary := reg.Label == p.BinaryPath
		isAnon := strings.HasPrefix(reg.Label, "[anon]") || strings.Contains(reg.Label, "anon")
		isHeap := strings.HasPrefix(reg.Label, "[heap]")
		isStack := strings.HasPrefix(reg.Label, "[stack]")

		if !isLib && (isAnon || isHeap || isStack || !strings.HasPrefix(reg.Label, "/")) {
			filtered = append(filtered, reg)
		} else if isBinary && strings.Contains(reg.Perms, "w") {
			// Include writable sections of the binary itself
			filtered = append(filtered, reg)
		}
	}
	result.Regions = filtered

	// Patterns
	reIP := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	reURL := regexp.MustCompile(`https?://[^\s"'>]+`)
	rePath := regexp.MustCompile(`/(?:tmp|var/tmp|home|etc)/[^\s"'>]+`)
	reKey := regexp.MustCompile(`(?i)(?:key|secret|token|auth|passwd|password)["']?\s*[:=]\s*["']?([a-zA-Z0-9\-_+/=]{16,})|([a-zA-Z0-9\_]{8,255}_[0-9]{10,}_[A-Z0-9]{10,})`)

	foundMap := make(map[string]bool)
	stats := ScanStats{
		RegionsScanned: len(filtered),
	}

	// Attach ptrace once for the whole scan session
	if err := ptraceAttach(pid); err == nil {
		var status syscall.WaitStatus
		syscall.Wait4(int(pid), &status, 0, nil)
		defer ptraceDetach(pid)
	} else {
		// If attach fails, we try reading without it, might still work for some regions or if privileged
		stats.ReadErrors++
		result.Findings = append(result.Findings, Finding{
			Kind:    "PTRACE_FAILED",
			Value:   err.Error(),
			Address: 0,
			Region:  "",
			Score:   0,
		})
	}

	for _, reg := range filtered {
		// Adaptive strategy for ScanHeap
		var chunks []struct {
			off  int64
			size int64
		}

		if reg.Size < 2*1024*1024 {
			chunks = append(chunks, struct {
				off  int64
				size int64
			}{0, reg.Size})
		} else if mode == "deep" {
			// Deep: Scan in 1MB chunks every 4MB
			for off := int64(0); off < reg.Size; off += 4 * 1024 * 1024 {
				chunkSize := int64(1024 * 1024)
				if off+chunkSize > reg.Size {
					chunkSize = reg.Size - off
				}
				chunks = append(chunks, struct {
					off  int64
					size int64
				}{off, chunkSize})
			}
		} else {
			// Quick (default): Scan first 1MB and last 1MB of large regions
			chunks = append(chunks, struct {
				off  int64
				size int64
			}{0, 1024 * 1024})
			chunks = append(chunks, struct {
				off  int64
				size int64
			}{reg.Size - 1024*1024, 1024 * 1024})
			stats.Truncated = true
		}

		for _, chk := range chunks {
			if chk.size <= 0 {
				continue
			}
			data, readErr := fs.ReadMemoryRawLocked(pid, reg.StartAddr+uint64(chk.off), chk.size)
			if readErr != nil {
				stats.ReadErrors++
				continue
			}
			stats.BytesRead += int64(len(data))
			strData := string(data)

			addFinding := func(kind, val string, addr uint64) {
				if !foundMap[val] {
					foundMap[val] = true
					score := 50
					if kind == "POTENT_KEY" {
						score = 90
					} else if kind == "URL" && (strings.Contains(val, "api") || strings.Contains(val, "secret")) {
						score = 80
					}

					result.Findings = append(result.Findings, Finding{
						Kind:    kind,
						Value:   val,
						Address: addr,
						Region:  reg.Label,
						Score:   score,
					},
					)
					stats.FindingsTotal++
				}
			}

			for _, m := range reIP.FindAllStringIndex(strData, -1) {
				addFinding("IP_ADDR", strData[m[0]:m[1]], reg.StartAddr+uint64(chk.off)+uint64(m[0]))
			}
			for _, m := range reURL.FindAllStringIndex(strData, -1) {
				addFinding("URL", strData[m[0]:m[1]], reg.StartAddr+uint64(chk.off)+uint64(m[0]))
			}
			for _, m := range rePath.FindAllStringIndex(strData, -1) {
				addFinding("SENS_PATH", strData[m[0]:m[1]], reg.StartAddr+uint64(chk.off)+uint64(m[0]))
			}
			for _, m := range reKey.FindAllStringSubmatchIndex(strData, -1) {
				if len(m) >= 4 {
					addFinding("POTENT_KEY", strData[m[2]:m[3]], reg.StartAddr+uint64(chk.off)+uint64(m[2]))
				}
			}

			if stats.FindingsTotal > 200 {
				stats.Truncated = true
				break
			}
		}
		if stats.FindingsTotal > 200 {
			break
		}
	}

	result.Stats = stats
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// ── JSON Endpoints ─────────────────────────────────────────────────────────

// InspectJSON returns a structured JSON variant of Inspect()
func (fs *FrozenSnapshot) InspectJSON(pid uint32) (string, error) {
	p, ok := fs.ByPID[pid]
	if !ok {
		p, ok = fs.Ghosts[pid]
		if !ok {
			return "", fmt.Errorf("process %d not found", pid)
		}
	}

	state := "ACTIVE"
	uptime := time.Since(p.StartTimestamp).Round(time.Millisecond)
	if !p.EndTimestamp.IsZero() {
		state = "EXITED"
		uptime = p.EndTimestamp.Sub(p.StartTimestamp).Round(time.Millisecond)
	}

	resp := ProcessInspectJSON{
		PID:            p.PID,
		PPID:           p.PPID,
		Comm:           p.Comm,
		BinaryPath:     p.BinaryPath,
		State:          state,
		Uptime:         uptime.String(),
		StartTime:      p.StartTimestamp,
		CpuUsage:       p.CpuUsage,
		MemoryUsage:    p.MemoryUsage,
		Children:       []ChildJSON{},
		NetworkEffects: []EffectJSON{},
		FileEffects:    []EffectJSON{},
	}
	if !p.EndTimestamp.IsZero() {
		resp.EndTime = &p.EndTimestamp
	}

	// Parent info
	resp.Parent = fmt.Sprintf("%d", p.PPID)
	if p.PPID > 0 {
		if parent, found := fs.ByPID[p.PPID]; found {
			resp.Parent = fmt.Sprintf("%s (%d)", parent.Comm, p.PPID)
		} else if parent, found := fs.Ghosts[p.PPID]; found {
			resp.Parent = fmt.Sprintf("%s (%d) [EXITED]", parent.Comm, p.PPID)
		}
	}

	// Children info
	for _, cpid := range p.ChildrenPID {
		childState := "UNKNOWN"
		comm := ""
		if child, found := fs.ByPID[cpid]; found {
			childState = "ACTIVE"
			comm = child.Comm
		} else if child, found := fs.Ghosts[cpid]; found {
			childState = "EXITED"
			comm = child.Comm
		}
		resp.Children = append(resp.Children, ChildJSON{
			PID:   cpid,
			Comm:  comm,
			State: childState,
		})
	}

	// Effects
	collapsedEffs := collapseEffects(p.Effects)
	for _, eff := range collapsedEffs {
		effJ := EffectJSON{
			Label:  eff.Label,
			Count:  eff.Count,
			Unique: eff.Unique,
		}
		if eff.Kind == EffectOpen {
			effJ.Category = "open"
			resp.FileEffects = append(resp.FileEffects, effJ)
		} else {
			effJ.Category = "connect"
			if strings.HasPrefix(eff.Original, "unix:") {
				effJ.Subtype = "unix_socket"
				effJ.Label = eff.Original[5:]
			} else {
				effJ.Subtype = "network"
			}
			resp.NetworkEffects = append(resp.NetworkEffects, effJ)
		}
	}

	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// SearchJSON returns search results in structured JSON format
func (fs *FrozenSnapshot) SearchJSON(query string) (string, error) {
	results := []SearchMatchJSON{}
	re, err := regexp.Compile("(?i)" + query)
	if err != nil {
		return "", err
	}

	for _, p := range fs.ByPID {
		match := false
		if re.MatchString(p.Comm) || re.MatchString(p.BinaryPath) {
			match = true
			results = append(results, SearchMatchJSON{
				PID:       p.PID,
				Comm:      p.Comm,
				MatchKind: "process",
			})
		}

		for _, eff := range p.Effects {
			if re.MatchString(eff.Target) {
				matchKind := "effect"
				if match {
					matchKind = "both"
				}

				effKind := "open"
				if eff.Kind == EffectConnect {
					effKind = "connect"
				}

				results = append(results, SearchMatchJSON{
					PID:        p.PID,
					Comm:       p.Comm,
					MatchKind:  matchKind,
					EffectKind: effKind,
					Target:     eff.Target,
					Count:      eff.Count,
					LastSeen:   eff.Last.Format(time.RFC3339),
				})
			}
		}
	}

	// Also search ghosts
	for _, p := range fs.Ghosts {
		// Same logic as active
		if re.MatchString(p.Comm) || re.MatchString(p.BinaryPath) {
			results = append(results, SearchMatchJSON{
				PID:       p.PID,
				Comm:      p.Comm,
				MatchKind: "process",
			})
		}
		for _, eff := range p.Effects {
			if re.MatchString(eff.Target) {
				effKind := "open"
				if eff.Kind == EffectConnect {
					effKind = "connect"
				}
				results = append(results, SearchMatchJSON{
					PID:        p.PID,
					Comm:       p.Comm,
					MatchKind:  "effect",
					EffectKind: effKind,
					Target:     eff.Target,
					Count:      eff.Count,
					LastSeen:   eff.Last.Format(time.RFC3339),
				})
			}
		}
	}

	bytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ProcessFamilyJSON returns the process lineage in a recursive JSON structure
func (fs *FrozenSnapshot) ProcessFamilyJSON(pid uint32) (string, error) {
	// We reuse the logic from ProcessFamily but structure it for JSON
	// Helper to lookup processes, including dynamically resolving untracked ones
	untracked := make(map[uint32]*ProcessNode)
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
			return "", fmt.Errorf("process %d not found", pid)
		}
	}

	for p_id, p := range fs.ByPID {
		childrenMap[p.PPID] = append(childrenMap[p.PPID], p_id)
	}
	for p_id, p := range fs.Ghosts {
		childrenMap[p.PPID] = append(childrenMap[p.PPID], p_id)
	}

	// Walk up to root
	root := target
	chain := make(map[uint32]bool)
	chain[target.PID] = true
	for {
		if root.PPID == 0 || root.PPID == 1 || root.PPID == 2 {
			break
		}
		parent, ok := getNode(root.PPID)
		if !ok {
			parent = readUntracked(root.PPID)
			if parent == nil {
				break
			}
		}
		root = parent
		chain[root.PID] = true
	}

	// Build recursive JSON
	var buildNode func(pid uint32) *FamilyNodeJSON
	buildNode = func(pid uint32) *FamilyNodeJSON {
		node, _ := getNode(pid)
		fNode := &FamilyNodeJSON{
			PID:      pid,
			Comm:     node.Comm,
			IsTarget: pid == target.PID,
		}

		children := childrenMap[pid]
		sort.Slice(children, func(i, j int) bool { return children[i] < children[j] })

		for _, cpid := range children {
			// Optimization: only recurse into active/ghost/untracked nodes we care about
			// or if they are in the ancestry chain we've already resolved.
			// Actually for family view we want all children.
			fNode.Children = append(fNode.Children, buildNode(cpid))
		}
		return fNode
	}

	rootJSON := buildNode(root.PID)
	bytes, err := json.MarshalIndent(rootJSON, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
