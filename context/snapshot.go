package ctxengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"godshell/observer/browser"
)

// ── Noise filter ───────────────────────────────────────────────────────────

// noisePatterns are path prefixes/suffixes that carry zero information for
// the LLM. Every process loads libc, every shell probes locales — skip them.
var noisePrefixes = []string{
	"/etc/ld.so",
	"/usr/lib/libc.so",
	"/usr/lib/libm.so",
	"/usr/lib/libdl.so",
	"/usr/lib/libpthread.so",
	"/usr/lib/librt.so",
	"/usr/lib/libreadline.so",
	"/usr/lib/libncursesw.so",
	"/usr/lib/libacl.so",
	"/usr/lib/locale/",
	"/usr/lib/gconv/",
	"/usr/share/locale/",
	"/proc/self/",
}

var noiseSuffixes = []string{
	".so",
	"ld.so.cache",
	"locale-archive",
	"gconv-modules.cache",
	"locale.alias",
}

func isNoise(path string) bool {
	if path == "" || path == "stat" || path == "status" ||
		path == "environ" || path == "cmdline" || path == "maps" || path == "fd" {
		return true
	}
	for _, p := range noisePrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	for _, s := range noiseSuffixes {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	if strings.Contains(path, ".so.") {
		return true
	}
	return false
}

// ── Path collapsing ────────────────────────────────────────────────────────

// collapsed represents a group of similar effects merged into one line.
type collapsed struct {
	Label    string // display name (e.g., "process monitoring")
	Pattern  string // underlying regex or norm path (e.g., "/proc/*/stat")
	Original string // the raw target before collapsing
	Count    uint64 // sum of individual effect counts
	Kind     EffectKind
	Unique   int // number of distinct targets collapsed into this label
}

var numericSegment = regexp.MustCompile(`/\d+/`)

// collapseEffects groups effects with similar paths.
// e.g., /proc/12345/cmdline + /proc/67890/cmdline → /proc/*/cmdline ×N
func collapseEffects(effects map[string]*Effect) []collapsed {
	groups := make(map[string]*collapsed) // normalized key → collapsed

	for _, eff := range effects {
		if eff.Kind == EffectOpen && isNoise(eff.Target) {
			continue
		}

		var key string
		var label string
		var original = eff.Target

		if eff.Kind == EffectConnect {
			// For connections, try semantic label first, fall back to target
			label = semanticLabel(eff.Target, EffectConnect)
			if label != "" {
				key = label // group all HTTPS connections under one "HTTPS" line
			} else {
				key = eff.Target
				label = eff.Target
			}
		} else {
			// Normalize numeric path segments for grouping
			normalized := numericSegment.ReplaceAllString(eff.Target, "/*/")
			label = semanticLabel(normalized, EffectOpen)
			if label != "" {
				key = label // group all "process monitoring" effects under one line
			} else {
				key = normalized
				label = normalized
				original = normalized
			}
		}

		if g, ok := groups[key]; ok {
			g.Count += eff.Count
			g.Unique++
			g.Original = original // keep last
		} else {
			groups[key] = &collapsed{
				Label:    label,
				Pattern:  key,
				Original: original,
				Count:    eff.Count,
				Kind:     eff.Kind,
				Unique:   1,
			}
		}
	}

	// Sort by count descending
	result := make([]collapsed, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

// ── Semantic labels ────────────────────────────────────────────────────────

type semanticRule struct {
	pattern *regexp.Regexp
	label   string
	kind    EffectKind
}

var basicSemanticRules = []semanticRule{
	{regexp.MustCompile(`/.*/stat$`), "process monitoring", EffectOpen},
	{regexp.MustCompile(`/.*/status$`), "process monitoring", EffectOpen},
	{regexp.MustCompile(`/.*/cmdline`), "process monitoring", EffectOpen},
	{regexp.MustCompile(`/stat$`), "system CPU stats", EffectOpen},
	{regexp.MustCompile(`/meminfo`), "memory monitoring", EffectOpen},
	{regexp.MustCompile(`/cpuinfo`), "CPU info polling", EffectOpen},
	{regexp.MustCompile(`/devices/system/cpu/.*/cpufreq`), "CPU frequency polling", EffectOpen},
	{regexp.MustCompile(`/class/power_supply`), "battery monitoring", EffectOpen},
	{regexp.MustCompile(`/dev/shm/\.org\.chromium`), "Chromium shared memory IPC", EffectOpen},
	{regexp.MustCompile(`/dev/shm/\.com\.google\.Chrome`), "Chrome shared memory IPC", EffectOpen},
	{regexp.MustCompile(`\.indexeddb\.`), "IndexedDB storage", EffectOpen},
	{regexp.MustCompile(`\.gemini/antigravity`), "Antigravity workspace", EffectOpen},
	{regexp.MustCompile(`:443\b`), "HTTPS", EffectConnect},
	{regexp.MustCompile(`:80\b`), "HTTP", EffectConnect},
	{regexp.MustCompile(`:53\b`), "DNS query", EffectConnect},
}

func semanticLabel(path string, kind EffectKind) string {
	for _, rule := range basicSemanticRules {
		if rule.kind == kind && rule.pattern.MatchString(path) {
			return rule.label
		}
	}
	return ""
}

// ── Ghost grouping ─────────────────────────────────────────────────────────

type ghostGroup struct {
	Binary   string
	Count    int
	ParentID uint32
	Names    map[string]int // comm → count of processes
}

func groupGhosts(ghosts map[uint32]*ProcessNode) []ghostGroup {
	groups := make(map[string]*ghostGroup) // binary path → group

	for _, p := range ghosts {
		key := p.BinaryPath
		if key == "" {
			key = p.Comm
		}
		// Normalize — use just basename
		base := filepath.Base(key)

		if g, ok := groups[base]; ok {
			g.Count++
		} else {
			groups[base] = &ghostGroup{
				Binary:   key,
				Count:    1,
				ParentID: p.PPID,
				Names:    make(map[string]int),
			}
		}
		groups[base].Names[p.Comm]++
	}

	result := make([]ghostGroup, 0, len(groups))
	for _, g := range groups {
		result = append(result, *g)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	return result
}

// ── Process summary line ───────────────────────────────────────────────────

// processLine formats the header for a process: "name (pid) parent:name [metrics]"
func processLine(p *ProcessNode, t *ProcessTree, effects []collapsed, grouped bool) string {
	name := p.Comm

	// Parent info
	parentName := ""
	if p.PPID > 0 {
		if parent, ok := t.ByPID[p.PPID]; ok {
			parentName = parent.Comm
		} else if parent, ok := t.Ghosts[p.PPID]; ok {
			parentName = parent.Comm
		} else if p.PPID == 1 {
			parentName = "init"
		} else {
			// Fallback: the parent exists but we didn't catch its exec (e.g., it started before godshell)
			// Read <proc>/<ppid>/stat to get its name
			statPath := fmt.Sprintf("%s/%d/stat", t.ProcPath, p.PPID)
			if data, err := os.ReadFile(statPath); err == nil {
				// /proc/[pid]/stat format starts with: ppid (comm) state ...
				// Extract the part between the first '(' and the last ')'
				start := strings.IndexByte(string(data), '(')
				end := strings.LastIndexByte(string(data), ')')
				if start != -1 && end != -1 && start < end {
					parentName = string(data[start+1 : end])
				}
			}
			if parentName == "" {
				parentName = fmt.Sprintf("pid:%d", p.PPID)
			}
		}
	}

	// Metrics
	var metrics []string
	if p.CpuUsage > 0.1 {
		metrics = append(metrics, fmt.Sprintf("%.1f%% CPU", p.CpuUsage))
	}
	if p.MemoryUsage > 1024 {
		metrics = append(metrics, fmt.Sprintf("%d MB", p.MemoryUsage/1024))
	} else if p.MemoryUsage > 0 {
		metrics = append(metrics, fmt.Sprintf("%d KB", p.MemoryUsage))
	}

	header := fmt.Sprintf("%s (%d)", name, p.PID)
	if parentName != "" {
		header += fmt.Sprintf(" parent:%s", parentName)
	}
	return header
}

// processLineFrozen formats the header for a process: "name (pid) parent:name [metrics]"
func processLineFrozen(p *ProcessNode, t *FrozenSnapshot, effects []collapsed, grouped bool) string {
	name := p.Comm
	if !grouped && len(p.BinaryPath) > 0 {
		name += "//" + p.BinaryPath
	}

	// Parent info
	parentName := ""
	if p.PPID > 0 {
		if parent, ok := t.ByPID[p.PPID]; ok {
			parentName = parent.Comm
		} else if parent, ok := t.Ghosts[p.PPID]; ok {
			parentName = parent.Comm
		} else if p.PPID == 1 {
			parentName = "init"
		} else {
			statPath := fmt.Sprintf("%s/%d/stat", "/proc", p.PPID) // Frozen snapshot doesn't have t.ProcPath
			// For frozen snapshots, we might not have the original ProcPath stored.
			// Let's just use /proc as a fallback for now, or we should have stored it.
			if data, err := os.ReadFile(statPath); err == nil {
				start := strings.IndexByte(string(data), '(')
				end := strings.LastIndexByte(string(data), ')')
				if start != -1 && end != -1 && start < end {
					parentName = string(data[start+1 : end])
				}
			}
			if parentName == "" {
				parentName = fmt.Sprintf("pid:%d", p.PPID)
			}
		}
	}

	// Metrics
	var metrics []string

	uptime := time.Since(p.StartTimestamp).Round(time.Millisecond)
	stateStr := fmt.Sprintf("start:%s", p.StartTimestamp.Format("15:04:05"))
	if !p.EndTimestamp.IsZero() {
		uptime = p.EndTimestamp.Sub(p.StartTimestamp).Round(time.Millisecond)
		stateStr += fmt.Sprintf(" end:%s", p.EndTimestamp.Format("15:04:05"))
	}
	metrics = append(metrics, fmt.Sprintf("up:%v", uptime))
	metrics = append(metrics, stateStr)

	if p.CpuUsage > 0.1 {
		metrics = append(metrics, fmt.Sprintf("%.1f%% CPU", p.CpuUsage))
	}
	if p.MemoryUsage > 1024 {
		metrics = append(metrics, fmt.Sprintf("%d MB", p.MemoryUsage/1024))
	} else if p.MemoryUsage > 0 {
		metrics = append(metrics, fmt.Sprintf("%d KB", p.MemoryUsage))
	}

	header := fmt.Sprintf("%s (%d)", name, p.PID)
	if parentName != "" {
		header += fmt.Sprintf(" parent:%s", parentName)
	}
	if len(metrics) > 0 {
		header += " [" + strings.Join(metrics, ", ") + "]"
	}

	return header
}

// ── Main Snapshot ──────────────────────────────────────────────────────────

// groupKey returns a string to cluster processes by "App".
// It uses either the root ancestor Comm name or the process's own Comm/Binary.
func groupKey(p *ProcessNode, t *ProcessTree) string {
	// Walk up to find the highest non-generic ancestor
	curr := p
	for {
		if curr.PPID == 0 || curr.PPID == 1 || curr.PPID == 2 { // init/kthreadd
			break
		}
		parent, ok := t.ByPID[curr.PPID]
		if !ok {
			break
		}

		// Don't group under generic system daemons like systemd, bash, or sshd
		if parent.Comm == "systemd" || parent.Comm == "bash" || parent.Comm == "zsh" || parent.Comm == "sshd" || parent.Comm == "tmux: server" {
			break
		}

		curr = parent
	}

	name := curr.Comm
	if len(curr.BinaryPath) > 0 {
		name = curr.BinaryPath
	}
	return name
}

// groupKeyFrozen returns a string to cluster processes by "App" for a frozen snapshot.
func groupKeyFrozen(p *ProcessNode, fs *FrozenSnapshot) string {
	curr := p
	for {
		if curr.PPID == 0 || curr.PPID == 1 || curr.PPID == 2 {
			break
		}
		parent, ok := fs.ByPID[curr.PPID]
		if !ok {
			break
		}

		if parent.Comm == "systemd" || parent.Comm == "bash" || parent.Comm == "zsh" || parent.Comm == "sshd" || parent.Comm == "tmux: server" {
			break
		}

		curr = parent
	}

	name := curr.Comm
	if len(curr.BinaryPath) > 0 {
		name = curr.BinaryPath
	}
	return name
}

// Summary generates a compact, LLM-ready summary of the frozen process graph.
// Target: ~30-40 lines, ~1500 tokens. Every line carries semantic meaning.
func (fs *FrozenSnapshot) Summary() string {
	var b strings.Builder

	// Header
	ghostCount := len(fs.Ghosts)
	activeCount := len(fs.ByPID)
	b.WriteString(fmt.Sprintf("SYSTEM STATE (%d active processes, %d exited in last 60s)\n",
		activeCount, ghostCount))
	b.WriteString(fmt.Sprintf("Snapshot Timestamp: %s\n\n", fs.Timestamp.Format(time.RFC3339)))

	// Group active processes by "App"
	type appGroup struct {
		Name  string
		Procs []*ProcessNode
	}
	groupsMap := make(map[string]*appGroup)

	// Since we are operating on a frozen snapshot, groupKey needs access to ByPID and Ghosts
	// We'll pass the whole frozen snapshot instead of a live tree
	for _, p := range fs.ByPID {
		k := groupKeyFrozen(p, fs)
		if _, ok := groupsMap[k]; !ok {
			groupsMap[k] = &appGroup{Name: k, Procs: make([]*ProcessNode, 0)}
		}
		groupsMap[k].Procs = append(groupsMap[k].Procs, p)
	}

	var groups []*appGroup
	for _, g := range groupsMap {
		groups = append(groups, g)
	}

	// Sort groups by number of processes (descending)
	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].Procs) > len(groups[j].Procs)
	})

	for _, g := range groups {
		if len(g.Procs) > 1 {
			// Print App Header for grouped processes
			var pids []string
			for _, p := range g.Procs {
				pids = append(pids, fmt.Sprintf("%d", p.PID))
			}
			b.WriteString(fmt.Sprintf("[App Group: %s] → PIDs %s\n", g.Name, strings.Join(pids, ", ")))
		}

		// Sort processes within the group by effect count
		sort.Slice(g.Procs, func(i, j int) bool {
			return totalEffects(g.Procs[i]) > totalEffects(g.Procs[j])
		})

		for _, p := range g.Procs {
			effects := collapseEffects(p.Effects)
			header := processLineFrozen(p, fs, effects, len(g.Procs) > 1)
			if len(g.Procs) > 1 {
				b.WriteString("  " + header + "\n")
			} else {
				b.WriteString(header + "\n")
			}

			// Show top 3 collapsed effects
			effLimit := 3
			if len(effects) < effLimit {
				effLimit = len(effects)
			}
			for _, eff := range effects[:effLimit] {
				prefix := "  ├──"
				if effLimit == 1 || &eff == &effects[effLimit-1] {
					prefix = "  └──"
				}
				if len(g.Procs) == 1 {
					if effLimit == 1 || &eff == &effects[effLimit-1] {
						prefix = "└──"
					} else {
						prefix = "├──"
					}
				}

				display := eff.Label
				if eff.Kind == EffectConnect {
					if strings.HasPrefix(eff.Original, "unix:") {
						display = fmt.Sprintf("unix socket: %s", eff.Original[5:])
					} else {
						display = fmt.Sprintf("network: %s", eff.Label)
					}
				}
				if eff.Unique > 1 {
					b.WriteString(fmt.Sprintf("%s %s ×%d (%d targets)\n", prefix, display, eff.Count, eff.Unique))
				} else {
					b.WriteString(fmt.Sprintf("%s %s ×%d\n", prefix, display, eff.Count))
				}
			}
			remaining := len(effects) - effLimit
			if remaining > 0 {
				b.WriteString(fmt.Sprintf("%s +%d more activities\n", func() string {
					if len(g.Procs) == 1 {
						return "    └──"
					}
					return "      └──"
				}(), remaining))
			}
		}
		b.WriteString("\n")
	}

	// Ghost summary — grouped by app but individual lines
	if ghostCount > 0 {
		b.WriteString("RECENTLY EXITED:\n")
		var ghosts []*ProcessNode
		for _, g := range fs.Ghosts {
			ghosts = append(ghosts, g)
		}
		// Sort ghosts by exit time, most recent first
		sort.Slice(ghosts, func(i, j int) bool {
			return ghosts[i].EndTimestamp.After(ghosts[j].EndTimestamp)
		})

		for _, g := range ghosts {
			name := g.Comm
			if len(g.BinaryPath) > 0 {
				name = filepath.Base(g.BinaryPath)
			}
			duration := g.EndTimestamp.Sub(g.StartTimestamp).Round(time.Millisecond)

			// Summarize activity
			var opens, connects uint64
			for _, eff := range g.Effects {
				if eff.Kind == EffectOpen {
					opens += eff.Count
				} else if eff.Kind == EffectConnect {
					connects += eff.Count
				}
			}

			actStr := fmt.Sprintf("%d ops", opens)
			if connects > 0 {
				actStr += fmt.Sprintf(", %d connects", connects)
			}

			b.WriteString(fmt.Sprintf("  %s (%d) — %v lifespan [%s]\n", name, g.PID, duration, actStr))
		}
	}

	return b.String()
}

// ── DumpDebug (raw view) ───────────────────────────────────────────────────

// DumpDebug prints the unfiltered process graph to stdout.
// Use for development; the LLM consumes Snapshot() instead.
func (t *ProcessTree) DumpDebug() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fmt.Printf("\n\033[1;36m═══ Process Graph (%d active, %d ghosts) ═══\033[0m\n",
		len(t.ByPID), len(t.Ghosts))

	procs := make([]*ProcessNode, 0, len(t.ByPID))
	for _, p := range t.ByPID {
		procs = append(procs, p)
	}
	sort.Slice(procs, func(i, j int) bool {
		return totalEffects(procs[i]) > totalEffects(procs[j])
	})

	limit := 15
	if len(procs) < limit {
		limit = len(procs)
	}
	for _, p := range procs[:limit] {
		printProcess(p, false)
	}
	if len(procs) > 15 {
		fmt.Printf("  \033[2m... and %d more processes\033[0m\n", len(procs)-15)
	}

	if len(t.Ghosts) > 0 {
		fmt.Printf("\n\033[1;33m─── Ghosts (recently exited) ───\033[0m\n")
		ghosts := make([]*ProcessNode, 0, len(t.Ghosts))
		for _, p := range t.Ghosts {
			ghosts = append(ghosts, p)
		}
		sort.Slice(ghosts, func(i, j int) bool {
			return ghosts[i].EndTimestamp.After(ghosts[j].EndTimestamp)
		})
		limit := 10
		if len(ghosts) < limit {
			limit = len(ghosts)
		}
		for _, p := range ghosts[:limit] {
			printProcess(p, true)
		}
	}
	fmt.Println()
}

func printProcess(p *ProcessNode, ghost bool) {
	name := p.Comm
	if p.BinaryPath != "" {
		name = p.BinaryPath
	}

	icon := "●"
	if ghost {
		icon = "○"
	}

	opens := 0
	connects := 0
	for _, eff := range p.Effects {
		switch eff.Kind {
		case EffectOpen:
			opens += int(eff.Count)
		case EffectConnect:
			connects += int(eff.Count)
		}
	}

	parts := []string{fmt.Sprintf("%d opens", opens)}
	if connects > 0 {
		parts = append(parts, fmt.Sprintf("%d connects", connects))
	}
	if p.CpuUsage > 0.1 {
		parts = append(parts, fmt.Sprintf("%.1f%% CPU", p.CpuUsage))
	}
	if p.MemoryUsage > 0 {
		if p.MemoryUsage > 1024 {
			parts = append(parts, fmt.Sprintf("%d MB RSS", p.MemoryUsage/1024))
		} else {
			parts = append(parts, fmt.Sprintf("%d KB RSS", p.MemoryUsage))
		}
	}
	fmt.Printf("  %s \033[1m%-30s\033[0m pid=%-6d ppid=%-6d [%s]\n",
		icon, name, p.PID, p.PPID, strings.Join(parts, ", "))

	effects := make([]*Effect, 0, len(p.Effects))
	for _, eff := range p.Effects {
		effects = append(effects, eff)
	}
	sort.Slice(effects, func(i, j int) bool {
		return effects[i].Count > effects[j].Count
	})
	effLimit := 5
	if len(effects) < effLimit {
		effLimit = len(effects)
	}
	for _, eff := range effects[:effLimit] {
		kindLabel := "open"
		if eff.Kind == EffectConnect {
			kindLabel = "conn"
		}
		fmt.Printf("      [%s] %s ×%d\n", kindLabel, eff.Target, eff.Count)
	}
	if len(effects) > 5 {
		fmt.Printf("      \033[2m... +%d more effects\033[0m\n", len(effects)-5)
	}
}

func totalEffects(p *ProcessNode) int {
	total := 0
	for _, eff := range p.Effects {
		total += int(eff.Count)
	}
	return total
}

// ── JSON Endpoints ─────────────────────────────────────────────────────────

type EffectJSON struct {
	Label    string `json:"label"`
	Count    uint64 `json:"count"`
	Unique   int    `json:"unique_targets,omitempty"`
	Category string `json:"category"`          // "open" or "connect"
	Subtype  string `json:"subtype,omitempty"` // "unix_socket" or "network"
}

type ProcessJSON struct {
	PID         uint32       `json:"pid"`
	PPID        uint32       `json:"ppid"`
	Name        string       `json:"name"`
	TotalOpens  int          `json:"total_opens"`
	TotalConns  int          `json:"total_connections"`
	CpuUsage    float64      `json:"cpu_usage_percent"`
	MemoryUsage uint64       `json:"memory_usage_kb"`
	TopEffects  []EffectJSON `json:"top_effects"`
	MoreEffects int          `json:"more_effects_count"`
}

type ProcessInspectJSON struct {
	PID            uint32       `json:"pid"`
	PPID           uint32       `json:"ppid"`
	Comm           string       `json:"comm"`
	BinaryPath     string       `json:"binary_path,omitempty"`
	State          string       `json:"state"`
	Uptime         string       `json:"uptime"`
	StartTime      time.Time    `json:"start_time"`
	EndTime        *time.Time   `json:"end_time,omitempty"`
	CpuUsage       float64      `json:"cpu_usage_percent"`
	MemoryUsage    uint64       `json:"memory_usage_kb"`
	Parent         string       `json:"parent"` // formatted name (pid)
	Children       []ChildJSON  `json:"children"`
	NetworkEffects []EffectJSON `json:"network_effects"`
	FileEffects    []EffectJSON `json:"file_effects"`
}

type ChildJSON struct {
	PID   uint32 `json:"pid"`
	Comm  string `json:"comm"`
	State string `json:"state"`
}

type SearchMatchJSON struct {
	PID        uint32 `json:"pid"`
	Comm       string `json:"comm"`
	MatchKind  string `json:"match_kind"` // "process" or "effect"
	EffectKind string `json:"effect_kind,omitempty"`
	Target     string `json:"target,omitempty"`
	Count      uint64 `json:"count,omitempty"`
	LastSeen   string `json:"last_seen,omitempty"`
}

type FamilyNodeJSON struct {
	PID      uint32            `json:"pid"`
	Comm     string            `json:"comm"`
	IsTarget bool              `json:"is_target"`
	Children []*FamilyNodeJSON `json:"children,omitempty"`
}

type AppGroupJSON struct {
	Name      string        `json:"group_name"`
	PIDs      []uint32      `json:"pids"`
	Processes []ProcessJSON `json:"processes"`
}

type GhostJSON struct {
	Name     string `json:"name"`
	PID      uint32 `json:"pid"`
	Duration string `json:"lifespan_duration"`
	Opens    uint64 `json:"total_opens"`
	Connects uint64 `json:"total_connections"`
}

type SnapshotSummaryJSON struct {
	Timestamp      time.Time      `json:"timestamp"`
	ActiveGroups   []AppGroupJSON `json:"active_groups"`
	RecentlyExited []GhostJSON    `json:"recently_exited,omitempty"`
}

// Memory Region metadata
type RegionSummary struct {
	StartAddr uint64 `json:"start"`
	EndAddr   uint64 `json:"end"`
	Size      int64  `json:"size"`
	Perms     string `json:"perms"`
	Label     string `json:"label"`
}

// A single sensitive finding
type Finding struct {
	Kind    string `json:"kind"` // IP_ADDR, URL, POTENT_KEY, SENS_PATH
	Value   string `json:"value"`
	Address uint64 `json:"address"`
	Region  string `json:"region"`
	Score   int    `json:"score"`
}

// Statistics for a scan operation
type ScanStats struct {
	RegionsScanned int   `json:"regions_scanned"`
	BytesRead      int64 `json:"bytes_read"`
	FindingsTotal  int   `json:"findings_total"`
	Truncated      bool  `json:"truncated"`
	ReadErrors     int   `json:"read_errors"`
}

// Top-level structured result for LLM consumption
type HeapScanResult struct {
	PID      uint32          `json:"pid"`
	Comm     string          `json:"comm"`
	Error    string          `json:"error,omitempty"`
	Regions  []RegionSummary `json:"regions"`
	Findings []Finding       `json:"findings"`
	Stats    ScanStats       `json:"stats"`
}

// BrowserMapJSON returns a structured JSON map of running Chrome/Firefox processes.
func (fs *FrozenSnapshot) BrowserMapJSON() (string, error) {
	// Call to our new pure-go parsing package
	// We need to import godshell/observer/browser
	procs, err := browser.MapProcesses()
	if err != nil {
		return "", fmt.Errorf("failed to map browsers: %w", err)
	}

	data, err := json.MarshalIndent(procs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SummaryJSON returns a structured JSON equivalent of Summary()
func (fs *FrozenSnapshot) SummaryJSON() (string, error) {
	summary := SnapshotSummaryJSON{
		Timestamp:      fs.Timestamp,
		ActiveGroups:   []AppGroupJSON{},
		RecentlyExited: []GhostJSON{},
	}

	// Group strictly matching TUI logic
	type appGroup struct {
		Name  string
		Procs []*ProcessNode
	}
	groupsMap := make(map[string]*appGroup)

	for _, p := range fs.ByPID {
		k := groupKeyFrozen(p, fs)
		if _, ok := groupsMap[k]; !ok {
			groupsMap[k] = &appGroup{Name: k, Procs: make([]*ProcessNode, 0)}
		}
		groupsMap[k].Procs = append(groupsMap[k].Procs, p)
	}

	var groups []*appGroup
	for _, g := range groupsMap {
		groups = append(groups, g)
	}

	sort.Slice(groups, func(i, j int) bool {
		return len(groups[i].Procs) > len(groups[j].Procs)
	})

	for _, g := range groups {
		groupJson := AppGroupJSON{
			Name:      g.Name,
			PIDs:      []uint32{},
			Processes: []ProcessJSON{},
		}
		for _, p := range g.Procs {
			groupJson.PIDs = append(groupJson.PIDs, p.PID)
		}

		sort.Slice(g.Procs, func(i, j int) bool {
			return totalEffects(g.Procs[i]) > totalEffects(g.Procs[j])
		})

		for _, p := range g.Procs {
			effects := collapseEffects(p.Effects)

			opens := 0
			connects := 0
			for _, eff := range p.Effects {
				if eff.Kind == EffectOpen {
					opens += int(eff.Count)
				} else {
					connects += int(eff.Count)
				}
			}

			procJson := ProcessJSON{
				PID:         p.PID,
				PPID:        p.PPID,
				Name:        p.Comm,
				TotalOpens:  opens,
				TotalConns:  connects,
				CpuUsage:    p.CpuUsage,
				MemoryUsage: p.MemoryUsage,
				TopEffects:  []EffectJSON{},
			}
			if p.BinaryPath != "" {
				procJson.Name = p.BinaryPath
			}

			effLimit := 3
			if len(effects) < effLimit {
				effLimit = len(effects)
			}

			for _, eff := range effects[:effLimit] {
				effJson := EffectJSON{
					Label:  eff.Label,
					Count:  eff.Count,
					Unique: eff.Unique,
				}
				if eff.Kind == EffectOpen {
					effJson.Category = "open"
				} else {
					effJson.Category = "connect"
					if strings.HasPrefix(eff.Original, "unix:") {
						effJson.Subtype = "unix_socket"
						effJson.Label = eff.Original[5:]
					} else {
						effJson.Subtype = "network"
					}
				}
				procJson.TopEffects = append(procJson.TopEffects, effJson)
			}
			procJson.MoreEffects = len(effects) - effLimit
			groupJson.Processes = append(groupJson.Processes, procJson)
		}
		summary.ActiveGroups = append(summary.ActiveGroups, groupJson)
	}

	// Ghosts matching TUI logic
	var ghosts []*ProcessNode
	for _, g := range fs.Ghosts {
		ghosts = append(ghosts, g)
	}
	sort.Slice(ghosts, func(i, j int) bool {
		return ghosts[i].EndTimestamp.After(ghosts[j].EndTimestamp)
	})

	for _, g := range ghosts {
		name := g.Comm
		if len(g.BinaryPath) > 0 {
			name = filepath.Base(g.BinaryPath)
		}
		var opens, connects uint64
		for _, eff := range g.Effects {
			if eff.Kind == EffectOpen {
				opens += eff.Count
			} else {
				connects += eff.Count
			}
		}

		summary.RecentlyExited = append(summary.RecentlyExited, GhostJSON{
			Name:     name,
			PID:      g.PID,
			Duration: g.EndTimestamp.Sub(g.StartTimestamp).Round(time.Millisecond).String(),
			Opens:    opens,
			Connects: connects,
		})
	}

	bytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// suppress unused import
var _ = time.Now
