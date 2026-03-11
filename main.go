package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"godshell/cmd"
	"godshell/config"
	ctxengine "godshell/context"
	"godshell/llm"
	"godshell/observer"
	"godshell/store"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ─────────────────────────────────────────────────────────────────

var (
	promptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true)
	shellStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	hintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	okStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
	bodyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	scanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6655")).Italic(true)

	// Tool card: red background pill with white text for the name portion
	toolCardNameStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#CC2200")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1)

	// The metadata portion after the pill — dimmer, on default bg
	toolCardMetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#999999")).
				Italic(true)
	toolCardHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF8866")).
				Bold(true)
)

// ── Streaming messages ─────────────────────────────────────────────────────

// toolCallMsg sent when a tool starts — just signals spinner label change.
type toolCallMsg struct {
	name string
}

// toolDoneMsg sent after a tool completes — renders the card with metadata.
type toolDoneMsg struct {
	card string // fully rendered card line
}

// aiDoneMsg when the final text response arrives.
type aiDoneMsg struct {
	text string
	err  error
}

// ── Other messages ─────────────────────────────────────────────────────────

type snapshotSavedMsg struct {
	id  int64
	err error
}
type snapshotLoadedMsg struct {
	snap *ctxengine.FrozenSnapshot
	err  error
}
type snapshotsListedMsg struct {
	metas []store.SnapshotMeta
	err   error
}

// ── View mode ──────────────────────────────────────────────────────────────

type viewMode int

const (
	modeChat viewMode = iota
	modeLoader
)

// ── Model ──────────────────────────────────────────────────────────────────

type model struct {
	textInput textinput.Model
	spinner   spinner.Model
	vp        viewport.Model
	history   []string
	waiting   bool
	mode      viewMode
	statusMsg string
	width     int
	height    int
	ready     bool

	// what tool is currently running (shown on spinner line)
	activeTool string

	snapTable table.Model
	snapMetas []store.SnapshotMeta

	tree   *ctxengine.ProcessTree
	conv   *llm.Conversation
	client llm.Client
	snap   *ctxengine.FrozenSnapshot
	cfg    config.Config

	msgChan chan tea.Msg
}

func initialModel(tree *ctxengine.ProcessTree, client llm.Client, snap *ctxengine.FrozenSnapshot, cfg config.Config) model {
	ti := textinput.New()
	ti.Placeholder = "ask godshell something..."
	ti.Prompt = promptStyle.Render("❯ ")
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 80

	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))

	conv := llm.NewConversation(snap)

	return model{
		textInput: ti, spinner: sp,
		history: []string{}, mode: modeChat,
		tree: tree, conv: conv, client: client, snap: snap, cfg: cfg,
		msgChan: make(chan tea.Msg, 64),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// ── Update ─────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := m.vpHeight()
		if !m.ready {
			m.vp = viewport.New(msg.Width, h)
			m.vp.SetContent(m.renderHistory())
			m.ready = true
		} else {
			m.vp.Width = msg.Width
			m.vp.Height = h
		}
		m.textInput.Width = msg.Width - 6
		m.syncVP()
		return m, nil

	case toolCallMsg:
		m.activeTool = msg.name
		return m, listenChan(m.msgChan)

	case toolDoneMsg:
		m.history = append(m.history, msg.card)
		m.syncVP()
		return m, listenChan(m.msgChan)

	case aiDoneMsg:
		m.waiting = false
		m.activeTool = ""
		if msg.err != nil {
			m.history = append(m.history, errorStyle.Render("✗ "+msg.err.Error()))
		} else if msg.text != "" {
			m.history = append(m.history,
				shellStyle.Render("godshell❯ ")+bodyStyle.Render(msg.text),
			)
		}
		m.syncVP()
		return m, nil

	case snapshotSavedMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render("save failed: " + msg.err.Error())
		} else {
			m.statusMsg = okStyle.Render(fmt.Sprintf("✓ saved (id=%d)", msg.id))
		}
		return m, nil
	case snapshotLoadedMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render("load failed: " + msg.err.Error())
		} else {
			m.snap = msg.snap
			m.conv.UpdateSnapshot(msg.snap)
			m.statusMsg = okStyle.Render("✓ snapshot loaded")
		}
		return m, nil
	case snapshotsListedMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render("list: " + msg.err.Error())
			return m, nil
		}
		m.snapMetas = msg.metas
		m.snapTable = buildSnapshotTable(msg.metas)
		m.mode = modeLoader
		m.textInput.Blur()
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeLoader {
			return m.updateLoader(msg)
		}
		return m.updateChat(msg)
	}

	sp, spCmd := m.spinner.Update(msg)
	m.spinner = sp
	cmds = append(cmds, spCmd)
	if m.mode == modeChat {
		ti, tiCmd := m.textInput.Update(msg)
		m.textInput = ti
		cmds = append(cmds, tiCmd)
	}
	return m, tea.Batch(cmds...)
}

func listenChan(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

func (m *model) vpHeight() int {
	chrome := 5
	if m.waiting {
		chrome += 1
	}
	h := m.height - chrome
	if h < 4 {
		h = 4
	}
	return h
}

func (m *model) syncVP() {
	m.vp.SetContent(m.renderHistory())
	m.vp.GotoBottom()
}

func (m model) renderHistory() string {
	if m.vp.Width <= 0 {
		return strings.Join(m.history, "\n")
	}
	// Wrap each history item individually to the viewport width.
	// We subtract 2 for a tiny margin.
	style := lipgloss.NewStyle().Width(m.vp.Width - 2)
	var wrapped []string
	for _, item := range m.history {
		wrapped = append(wrapped, style.Render(item))
	}
	return strings.Join(wrapped, "\n")
}

// ── Key handlers ───────────────────────────────────────────────────────────

func (m model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.textInput.Value() == "" && !m.waiting {
		switch msg.String() {
		case "r":
			m.snap = m.tree.TakeSnapshot()
			m.conv.UpdateSnapshot(m.snap)
			m.statusMsg = okStyle.Render("↻ snapshot refreshed")
			return m, nil
		case "s":
			snap := m.snap
			return m, func() tea.Msg {
				id, err := store.SaveSnapshot("manual", snap)
				return snapshotSavedMsg{id: id, err: err}
			}
		case "l":
			return m, func() tea.Msg {
				metas, err := store.ListSnapshots()
				return snapshotsListedMsg{metas: metas, err: err}
			}
		case "up", "k", "down", "j", "pgup", "pgdown":
			vp, cmd := m.vp.Update(msg)
			m.vp = vp
			return m, cmd
		}
	}
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		input := strings.TrimSpace(m.textInput.Value())
		m.textInput.SetValue("")
		m.statusMsg = ""
		if input == "" {
			return m, nil
		}
		return m.handleInput(input)
	}
	ti, cmd := m.textInput.Update(msg)
	m.textInput = ti
	return m, cmd
}

func (m model) updateLoader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = modeChat
		m.textInput.Focus()
		return m, nil
	case tea.KeyEnter:
		row := m.snapTable.Cursor()
		if row >= 0 && row < len(m.snapMetas) {
			meta := m.snapMetas[row]
			m.mode = modeChat
			m.textInput.Focus()
			return m, func() tea.Msg {
				snap, err := store.LoadSnapshot(meta.ID)
				return snapshotLoadedMsg{snap: snap, err: err}
			}
		}
	case tea.KeyCtrlC:
		return m, tea.Quit
	}
	t, cmd := m.snapTable.Update(msg)
	m.snapTable = t
	return m, cmd
}

func (m model) handleInput(input string) (tea.Model, tea.Cmd) {
	switch input {
	case "exit", "quit":
		return m, tea.Quit
	case "help":
		m.history = append(m.history, helpText())
		m.syncVP()
		return m, nil
	case "clear":
		m.history = []string{}
		m.syncVP()
		return m, nil
	default:
		m.history = append(m.history, promptStyle.Render("you❯ ")+input)
		m.waiting = true
		m.activeTool = ""
		m.syncVP()

		conv := m.conv
		client := m.client
		ch := m.msgChan
		conv.History = append(conv.History, llm.Message{
			Role: llm.RoleUser, Content: input,
		})
		go streamingChatLoop(client, conv, ch)
		return m, listenChan(ch)
	}
}

// ── View ───────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.mode == modeLoader {
		return m.viewLoader()
	}
	return m.viewChat()
}

func (m model) viewChat() string {
	if !m.ready {
		return "initializing..."
	}
	var b strings.Builder

	b.WriteString(headerStyle.Render("⚡ godshell") + "\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width)) + "\n")
	b.WriteString(m.vp.View() + "\n")

	if m.waiting {
		label := "Scanning system snapshot..."
		if m.activeTool != "" {
			label = "Running " + m.activeTool + "..."
		}
		b.WriteString("  " + m.spinner.View() + " " + scanStyle.Render(label) + "\n")
	}

	b.WriteString(m.textInput.View() + "\n")

	scroll := ""
	if m.vp.TotalLineCount() > m.vp.Height {
		pct := int(m.vp.ScrollPercent() * 100)
		scroll = dimStyle.Render(fmt.Sprintf(" %d%%", pct))
	}
	bar := statusStyle.Render("[r]efresh  [s]ave  [l]oad  ↑↓ scroll  ctrl+c quit") + scroll
	if m.statusMsg != "" {
		bar = m.statusMsg + "  │  " + bar
	}
	b.WriteString(bar)
	return b.String()
}

func (m model) viewLoader() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("📂 Snapshot Loader") + "\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width)) + "\n\n")
	if len(m.snapMetas) == 0 {
		b.WriteString(hintStyle.Render("  no snapshots saved. press [s] in chat to save one.") + "\n")
	} else {
		b.WriteString(m.snapTable.View() + "\n")
	}
	b.WriteString("\n" + statusStyle.Render("↑↓ navigate  enter load  esc back") + "\n")
	return b.String()
}

// ── Streaming chat loop ────────────────────────────────────────────────────

func streamingChatLoop(client llm.Client, conv *llm.Conversation, ch chan<- tea.Msg) {
	for {
		resp, err := client.Chat(conv.History, conv.GetToolDefinitions())
		if err != nil {
			ch <- aiDoneMsg{err: err}
			return
		}
		conv.History = append(conv.History, *resp)

		if len(resp.ToolCalls) == 0 {
			ch <- aiDoneMsg{text: resp.Content}
			return
		}

		for _, tc := range resp.ToolCalls {
			var args map[string]interface{}
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			if args == nil {
				args = map[string]interface{}{}
			}

			// Signal tool start → updates spinner label
			ch <- toolCallMsg{name: tc.Function.Name}

			result, execErr := conv.ExecuteTool(tc.Function.Name, args)
			if execErr != nil {
				result = fmt.Sprintf("Error: %v", execErr)
			}

			conv.History = append(conv.History, llm.Message{
				Role: llm.RoleTool, ToolCallID: tc.ID,
				Name: tc.Function.Name, Content: result,
			})

			// Build a red card with result metadata
			card := buildToolCard(tc.Function.Name, args, result)
			ch <- toolDoneMsg{card: card}
		}
	}
}

// ── Tool card builder ──────────────────────────────────────────────────────

// buildToolCard creates a compact tool card with red pill + brief result metadata.
func buildToolCard(name string, args map[string]interface{}, result string) string {
	icon := iconForTool(name)
	pill := toolCardNameStyle.Render(icon + " " + name)
	meta := extractMeta(name, args, result)

	if meta == "" {
		return "  " + pill
	}
	return "  " + pill + "  " + toolCardMetaStyle.Render(meta)
}

// extractMeta pulls a one-line summary from the tool result for display on the card.
func extractMeta(name string, args map[string]interface{}, result string) string {
	switch name {
	case "summary":
		return metaSummary(result)
	case "inspect":
		return metaInspect(result, args)
	case "search":
		return metaSearch(result, args)
	case "family":
		return metaFamily(result, args)
	case "get_maps":
		return metaMaps(result, args)
	case "get_libraries":
		return metaLibraries(result)
	case "trace":
		return metaTrace(result, args)
	case "read_file":
		return metaReadFile(result, args)
	case "read_memory":
		return metaReadMemory(result, args)
	case "gohash_binary":
		return metaHash(result, args)
	case "goread_shell_history":
		return metaShellHistory(result, args)
	case "gonetwork_state":
		return metaNetwork(result, args)
	case "goread_environ":
		return metaEnviron(result, args)
	case "goextract_strings":
		return metaStrings(result, args)
	case "browser_map":
		return metaBrowserMap(result)
	}
	return trunc(firstNonEmptyLine(result), 60)
}

// ── Per-tool metadata extractors ───────────────────────────────────────────

func metaBrowserMap(result string) string {
	var procs []struct {
		Browser string `json:"browser"`
		Type    string `json:"type"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal([]byte(result), &procs); err == nil {
		chrome := 0
		firefox := 0
		tabs := 0
		for _, p := range procs {
			if p.Browser == "chrome" {
				chrome++
				if p.URL != "" {
					tabs++
				}
			} else {
				firefox++
				if p.Type == "web_content" {
					tabs++
				}
			}
		}
		parts := []string{}
		if chrome > 0 {
			parts = append(parts, fmt.Sprintf("%d chrome", chrome))
		}
		if firefox > 0 {
			parts = append(parts, fmt.Sprintf("%d firefox", firefox))
		}
		if tabs > 0 {
			parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("%d tabs", tabs)))
		}
		if len(parts) == 0 {
			return "no browsers found"
		}
		return strings.Join(parts, " · ")
	}
	return "browser map generated"
}

func metaSummary(result string) string {
	// Parse the actual SnapshotSummaryJSON
	var summary struct {
		ActiveGroups []struct {
			Name      string `json:"group_name"`
			Processes []struct {
				PID   uint32  `json:"pid"`
				Name  string  `json:"name"`
				CPU   float64 `json:"cpu_usage_percent"`
				MemKB uint64  `json:"memory_usage_kb"`
				Opens int     `json:"total_opens"`
				Conns int     `json:"total_connections"`
			} `json:"processes"`
		} `json:"active_groups"`
		Ghosts []struct {
			Name string `json:"name"`
		} `json:"recently_exited"`
	}
	if json.Unmarshal([]byte(result), &summary) == nil {
		totalProcs := 0
		var topName string
		var topCPU float64
		for _, g := range summary.ActiveGroups {
			totalProcs += len(g.Processes)
			for _, p := range g.Processes {
				if p.CPU > topCPU {
					topCPU = p.CPU
					topName = p.Name
				}
			}
		}
		parts := []string{
			fmt.Sprintf("%d groups", len(summary.ActiveGroups)),
			fmt.Sprintf("%d processes", totalProcs),
		}
		if len(summary.Ghosts) > 0 {
			parts = append(parts, fmt.Sprintf("%d exited", len(summary.Ghosts)))
		}
		if topName != "" && topCPU > 0.5 {
			parts = append(parts, toolCardHighStyle.Render(
				fmt.Sprintf("top: %s %.1f%%", trunc(topName, 20), topCPU)))
		}
		return strings.Join(parts, " · ")
	}
	lines := nonEmptyLines(result)
	return fmt.Sprintf("%d entries", len(lines))
}

func metaInspect(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	var data struct {
		Comm        string     `json:"comm"`
		Binary      string     `json:"binary_path"`
		State       string     `json:"state"`
		CPU         float64    `json:"cpu_usage_percent"`
		MemKB       uint64     `json:"memory_usage_kb"`
		Parent      string     `json:"parent"`
		Uptime      string     `json:"uptime"`
		Children    []struct{} `json:"children"`
		NetEffects  []struct{} `json:"network_effects"`
		FileEffects []struct{} `json:"file_effects"`
	}
	if json.Unmarshal([]byte(result), &data) == nil {
		name := data.Comm
		if name == "" {
			name = data.Binary
		}
		totalEffects := len(data.NetEffects) + len(data.FileEffects)
		parts := []string{}
		if name != "" {
			parts = append(parts, toolCardHighStyle.Render(name))
		}
		parts = append(parts, dimStyle.Render("PID "+pid))
		if data.State != "" {
			parts = append(parts, data.State)
		}
		if data.CPU > 0.1 {
			parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("%.1f%% CPU", data.CPU)))
		}
		if data.MemKB > 1024 {
			parts = append(parts, fmt.Sprintf("%d MB", data.MemKB/1024))
		} else if data.MemKB > 0 {
			parts = append(parts, fmt.Sprintf("%d KB", data.MemKB))
		}
		parts = append(parts, fmt.Sprintf("%d effects", totalEffects))
		if len(data.Children) > 0 {
			parts = append(parts, fmt.Sprintf("%d children", len(data.Children)))
		}
		if data.Parent != "" {
			parts = append(parts, "parent:"+trunc(data.Parent, 20))
		}
		return strings.Join(parts, " · ")
	}
	return "PID " + pid + " · " + trunc(firstNonEmptyLine(result), 50)
}

func metaSearch(result string, args map[string]interface{}) string {
	query := argStr(args, "query")
	var items []struct {
		PID       uint32 `json:"pid"`
		Comm      string `json:"comm"`
		MatchKind string `json:"match_kind"`
		Target    string `json:"target"`
	}
	if json.Unmarshal([]byte(result), &items) == nil {
		procMatches := 0
		effectMatches := 0
		var names []string
		seen := map[string]bool{}
		for _, m := range items {
			if m.MatchKind == "process" || m.MatchKind == "both" {
				procMatches++
			} else {
				effectMatches++
			}
			if !seen[m.Comm] {
				names = append(names, m.Comm)
				seen[m.Comm] = true
			}
		}
		parts := []string{fmt.Sprintf("/%s/", query)}
		if procMatches > 0 {
			parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("%d proc matches", procMatches)))
		}
		if effectMatches > 0 {
			parts = append(parts, fmt.Sprintf("%d effect matches", effectMatches))
		}
		if len(names) > 0 {
			top := names
			if len(top) > 3 {
				top = top[:3]
			}
			parts = append(parts, strings.Join(top, ", "))
		}
		return strings.Join(parts, " · ")
	}
	return fmt.Sprintf("/%s/ → %d results", query, len(nonEmptyLines(result)))
}

func metaFamily(result string, args map[string]interface{}) string {
	// FamilyNodeJSON is recursive: {pid, comm, is_target, children:[...]}
	// We want to show: parent → TARGET → child1, child2
	type familyNode struct {
		PID      uint32       `json:"pid"`
		Comm     string       `json:"comm"`
		IsTarget bool         `json:"is_target"`
		Children []familyNode `json:"children"`
	}

	var root familyNode
	if json.Unmarshal([]byte(result), &root) != nil {
		return "PID " + argStr(args, "pid") + " lineage"
	}

	// Walk the tree to find the target, its parent, and its children
	var targetComm, parentComm string
	var childNames []string

	var walk func(node *familyNode, parent string)
	walk = func(node *familyNode, parent string) {
		if node.IsTarget {
			targetComm = node.Comm
			parentComm = parent
			for _, c := range node.Children {
				childNames = append(childNames, c.Comm)
			}
			return
		}
		for i := range node.Children {
			walk(&node.Children[i], node.Comm)
		}
	}
	walk(&root, "")

	if targetComm == "" {
		targetComm = root.Comm
	}

	var chain []string
	if parentComm != "" {
		chain = append(chain, parentComm)
	}
	chain = append(chain, toolCardHighStyle.Render(targetComm))
	if len(childNames) > 0 {
		top := childNames
		if len(top) > 3 {
			top = top[:3]
		}
		chain = append(chain, strings.Join(top, ", "))
		if len(childNames) > 3 {
			chain = append(chain, fmt.Sprintf("+%d more", len(childNames)-3))
		}
	}
	return strings.Join(chain, " → ")
}

func metaMaps(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	lines := nonEmptyLines(result)
	regions := len(lines)
	hasHeap := strings.Contains(result, "[heap]")
	hasStack := strings.Contains(result, "[stack]")
	libCount := 0
	for _, l := range lines {
		if strings.Contains(l, ".so") {
			libCount++
		}
	}
	parts := []string{fmt.Sprintf("PID %s", pid), fmt.Sprintf("%d regions", regions)}
	if hasHeap {
		parts = append(parts, "heap")
	}
	if hasStack {
		parts = append(parts, "stack")
	}
	if libCount > 0 {
		parts = append(parts, fmt.Sprintf("%d libs mapped", libCount))
	}
	return strings.Join(parts, " · ")
}

func metaLibraries(result string) string {
	lines := nonEmptyLines(result)
	libs := 0
	var names []string
	for _, l := range lines {
		if strings.Contains(l, ".so") {
			libs++
			// Extract lib name
			parts := strings.Fields(l)
			for _, p := range parts {
				if strings.Contains(p, ".so") {
					segs := strings.Split(p, "/")
					names = append(names, segs[len(segs)-1])
					break
				}
			}
		}
	}
	if libs == 0 {
		return fmt.Sprintf("%d entries", len(lines))
	}
	meta := fmt.Sprintf("%d shared libraries", libs)
	if len(names) > 0 {
		top := names
		if len(top) > 3 {
			top = top[:3]
		}
		meta += " · " + strings.Join(top, ", ")
	}
	return meta
}

func metaTrace(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	lines := nonEmptyLines(result)
	syscalls := map[string]int{}
	for _, l := range lines {
		if idx := strings.Index(l, "("); idx > 0 {
			sc := strings.TrimSpace(l[:idx])
			if sc != "" && !strings.Contains(sc, " ") {
				syscalls[sc]++
			}
		}
	}
	// Find top syscall
	topSC := ""
	topCount := 0
	for sc, c := range syscalls {
		if c > topCount {
			topCount = c
			topSC = sc
		}
	}
	parts := []string{
		fmt.Sprintf("PID %s", pid),
		fmt.Sprintf("%d calls", len(lines)),
		fmt.Sprintf("%d unique syscalls", len(syscalls)),
	}
	if topSC != "" {
		parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("top: %s ×%d", topSC, topCount)))
	}
	return strings.Join(parts, " · ")
}

func metaReadFile(result string, args map[string]interface{}) string {
	path := argStr(args, "path")
	segs := strings.Split(path, "/")
	filename := segs[len(segs)-1]
	lineCount := len(nonEmptyLines(result))
	preview := trunc(strings.TrimSpace(firstNonEmptyLine(result)), 30)
	parts := []string{filename, fmt.Sprintf("%d bytes", len(result)), fmt.Sprintf("%d lines", lineCount)}
	if preview != "" {
		parts = append(parts, hintStyle.Render(preview))
	}
	return strings.Join(parts, " · ")
}

func metaReadMemory(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	addr := argStr(args, "address_hex")
	// Count printable strings in the dump
	printable := 0
	for _, c := range result {
		if c >= 32 && c < 127 {
			printable++
		}
	}
	pct := 0
	if len(result) > 0 {
		pct = printable * 100 / len(result)
	}
	return fmt.Sprintf("PID %s @ 0x%s · %d bytes · %d%% printable", pid, addr, len(result), pct)
}

func metaHash(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	hash := strings.TrimSpace(result)
	if len(hash) > 16 {
		hash = hash[:16] + "…"
	}
	return fmt.Sprintf("PID %s · sha256:%s", pid, hash)
}

func metaShellHistory(result string, args map[string]interface{}) string {
	user := argStr(args, "user")
	lines := nonEmptyLines(result)
	if len(lines) == 0 {
		return user + " · empty"
	}
	// Show last 2 commands
	last := lines[len(lines)-1]
	parts := []string{user, fmt.Sprintf("%d cmds", len(lines))}
	parts = append(parts, "last: "+trunc(strings.TrimSpace(last), 35))
	if len(lines) >= 2 {
		prev := trunc(strings.TrimSpace(lines[len(lines)-2]), 25)
		parts = append(parts, "prev: "+prev)
	}
	return strings.Join(parts, " · ")
}

func metaNetwork(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	// gonetwork_state returns plain text, not JSON — parse lines
	lines := nonEmptyLines(result)
	est := 0
	listen := 0
	closeWait := 0
	for _, l := range lines {
		lu := strings.ToUpper(l)
		if strings.Contains(lu, "ESTABLISHED") {
			est++
		}
		if strings.Contains(lu, "LISTEN") {
			listen++
		}
		if strings.Contains(lu, "CLOSE_WAIT") {
			closeWait++
		}
	}
	parts := []string{fmt.Sprintf("PID %s", pid), fmt.Sprintf("%d connections", len(lines))}
	if est > 0 {
		parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("%d ESTABLISHED", est)))
	}
	if listen > 0 {
		parts = append(parts, fmt.Sprintf("%d LISTEN", listen))
	}
	if closeWait > 0 {
		parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("%d CLOSE_WAIT", closeWait)))
	}
	return strings.Join(parts, " · ")
}

func metaEnviron(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	lines := nonEmptyLines(result)
	// Flag sensitive-looking vars
	sensitive := 0
	for _, l := range lines {
		lu := strings.ToUpper(l)
		if strings.Contains(lu, "KEY") || strings.Contains(lu, "SECRET") ||
			strings.Contains(lu, "TOKEN") || strings.Contains(lu, "PASS") ||
			strings.Contains(lu, "CREDENTIAL") {
			sensitive++
		}
	}
	parts := []string{fmt.Sprintf("PID %s · %d vars", pid, len(lines))}
	if sensitive > 0 {
		parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("⚠ %d sensitive", sensitive)))
	}
	return strings.Join(parts, " · ")
}

func metaStrings(result string, args map[string]interface{}) string {
	path := argStr(args, "path")
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	lines := nonEmptyLines(result)
	// Flag URLs & interesting patterns
	urls := 0
	for _, l := range lines {
		if strings.Contains(l, "http") || strings.Contains(l, "://") {
			urls++
		}
	}
	meta := fmt.Sprintf("%s · %d strings", filename, len(lines))
	if urls > 0 {
		meta += " · " + toolCardHighStyle.Render(fmt.Sprintf("%d URLs", urls))
	}
	return meta
}

// ── Tool icons ─────────────────────────────────────────────────────────────

func iconForTool(name string) string {
	switch name {
	case "summary":
		return "◈"
	case "inspect":
		return "◎"
	case "search":
		return "⌕"
	case "family":
		return "⎇"
	case "get_maps":
		return "▤"
	case "get_libraries":
		return "⧉"
	case "trace":
		return "⚡"
	case "read_file":
		return "📄"
	case "read_memory":
		return "⬡"
	case "gohash_binary":
		return "#"
	case "goread_shell_history":
		return "⏎"
	case "gonetwork_state":
		return "🔗"
	case "goread_environ":
		return "⊕"
	case "goextract_strings":
		return "✂"
	default:
		return "⚙"
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func firstNonEmptyLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			return t
		}
	}
	return ""
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func argStr(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func jsonStr(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func jsonFloat(m map[string]interface{}, key string) float64 {
	v, ok := m[key].(float64)
	if !ok {
		return 0
	}
	return v
}

// ── Snapshot table ─────────────────────────────────────────────────────────

func buildSnapshotTable(metas []store.SnapshotMeta) table.Model {
	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "Label", Width: 20},
		{Title: "Timestamp", Width: 28},
	}
	rows := make([]table.Row, len(metas))
	for i, m := range metas {
		rows[i] = table.Row{
			fmt.Sprintf("%d", m.ID),
			m.Label,
			m.Timestamp.Format("2006-01-02 15:04:05"),
		}
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#FF4444"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#CC2200")).
		Bold(true)
	t.SetStyles(s)
	return t
}

// ── Help ───────────────────────────────────────────────────────────────────

func helpText() string {
	return hintStyle.Render(strings.Join([]string{
		"hotkeys (when input is empty):",
		"  r        — refresh system snapshot",
		"  s        — save snapshot to SQLite",
		"  l        — load a saved snapshot",
		"  ↑/↓/k/j  — scroll chat history",
		"",
		"typed commands:",
		"  clear    — clear chat history",
		"  help     — show this",
		"  exit     — quit",
		"  (anything else is sent to the LLM)",
	}, "\n"))
}

// ── main ───────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			if err := cmd.RunConfigEditor(); err != nil {
				fmt.Fprintf(os.Stderr, "config error: %v\n", err)
				os.Exit(1)
			}
			return
		case "--help", "-h":
			fmt.Println("Usage: godshell [command]")
			fmt.Println()
			fmt.Println("Commands:")
			fmt.Println("  (none)    Start the AI REPL (requires sudo for eBPF)")
			fmt.Println("  config    Open interactive config editor")
			fmt.Println()
			fmt.Println("Environment overrides:")
			fmt.Println("  OPENROUTER_API_KEY    API key (overrides config)")
			fmt.Println("  GODSHELL_MODEL        Model name (overrides config)")
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			os.Exit(1)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v (using defaults)\n", err)
		cfg = config.Default()
	}
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	if mod := os.Getenv("GODSHELL_MODEL"); mod != "" {
		cfg.Model = mod
	}
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, errorStyle.Render("✗ API key not found"))
		fmt.Fprintln(os.Stderr, hintStyle.Render("  Please set your OpenRouter API key using: godshell config"))
		fmt.Fprintln(os.Stderr, hintStyle.Render("  Or set the environment variable: export OPENROUTER_API_KEY=sk-or-..."))
		os.Exit(1)
	}

	if err := store.Init(cfg.DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: store init: %v (save/load disabled)\n", err)
	}

	tree := ctxengine.NewProcessTree()
	events := make(chan observer.Event, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := observer.Run(ctx, events); err != nil {
			fmt.Fprintf(os.Stderr, "observer: %v\n", err)
		}
	}()
	go func() {
		for e := range events {
			tree.HandleEvent(e)
		}
	}()
	go tree.RefreshMetrics(2 * time.Second)
	go tree.EvictGhosts(time.Duration(cfg.SnapshotRetentionSec) * time.Second)

	fmt.Printf("⚡ godshell  model=%s\n", cfg.Model)
	time.Sleep(1 * time.Second)
	snap := tree.TakeSnapshot()
	client := llm.NewOpenAIClient(cfg.APIKey, cfg.Endpoint, cfg.Model)

	m := initialModel(tree, client, snap, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
