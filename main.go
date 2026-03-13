package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"godshell/client"
	"godshell/cmd"
	"godshell/config"
	ctxengine "godshell/context"
	"godshell/daemon"
	"godshell/intel"
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
	selStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#CC2200")).
			Foreground(lipgloss.Color("#FFFFFF")).
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
type snapshotDeletedMsg struct {
	id  int64
	err error
}

// ── View mode ──────────────────────────────────────────────────────────────

type viewMode int

const (
	modeChat viewMode = iota
	modeLoader
	modeConfirmContext
)

// procEntry is one row in the left-panel process list.
type procEntry struct {
	PID     uint32
	Comm    string
	Mem     string // e.g. "142 MB" or ""
	CPU     string // e.g. "3.2%" or ""
	Depth   int    // Indentation level in the tree
	HasKids bool
}

// leftPanelWidth is the fixed width of the process tree panel.
const leftPanelWidth = 36

// ── Model ──────────────────────────────────────────────────────────────────

type model struct {
	textInput     textinput.Model
	spinner       spinner.Model
	vp            viewport.Model
	history       []string
	waiting       bool
	mode          viewMode
	statusMsg     string
	width         int
	height        int
	ready         bool
	showSidePanel bool

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

	// ── Left panel: process navigation ──────────────────────────────
	procList    []procEntry     // flattened, viewable list of processes
	procCursor  int             // index into procList
	procScroll  int             // topmost visible row index
	selectedPID uint32          // currently highlighted PID (0 = none)
	expandedMap map[uint32]bool // PIDs that are expanded in the tree

	// ── Focus & Direct Output ───────────────────────────────────────────
	focusPanel       focusMode
	lastDirectTool   string
	lastDirectResult string
	confirmPrompt    string
}

type focusMode int

const (
	focusProcs focusMode = iota // left panel active
	focusChat                   // right panel active
)

func initialModel(tree *ctxengine.ProcessTree, client llm.Client, snap *ctxengine.FrozenSnapshot, cfg config.Config) model {
	ti := textinput.New()
	ti.Placeholder = "ask godshell something..."
	ti.Prompt = promptStyle.Render("❯ ")
	ti.CharLimit = 512
	ti.Width = 80

	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))

	conv := llm.NewConversation(snap, nil) // TUI will update this to use DaemonClient
	if cfg.VTApiKey != "" || cfg.AbuseIPKey != "" {
		conv.Intel = intel.New(cfg.VTApiKey, cfg.AbuseIPKey)
	}

	m := model{
		textInput: ti, spinner: sp,
		history: []string{}, mode: modeChat,
		tree: tree, conv: conv, client: client, snap: snap, cfg: cfg,
		msgChan:     make(chan tea.Msg, 64),
		expandedMap: make(map[uint32]bool),
		focusPanel:  focusChat, // start by chatting gg
	}
	if snap != nil {
		m.procList = buildProcTree(snap, m.expandedMap)
		if len(m.procList) > 0 {
			m.selectedPID = m.procList[0].PID
		}
	}
	return m
}

// initialDaemonModel replaces the local tree and snapshot with ones fetched via client
func initialDaemonModel(daemonClient *client.DaemonClient, llmClient llm.Client, snap *ctxengine.FrozenSnapshot, cfg config.Config) model {
	ti := textinput.New()
	ti.Placeholder = "ask godshell something..."
	ti.Prompt = promptStyle.Render("❯ ")
	ti.Focus()
	ti.CharLimit = 512
	ti.Width = 80

	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))

	conv := llm.NewConversation(snap, daemonClient)
	if cfg.VTApiKey != "" || cfg.AbuseIPKey != "" {
		conv.Intel = intel.New(cfg.VTApiKey, cfg.AbuseIPKey)
	}

	m := model{
		textInput: ti, spinner: sp,
		history: []string{}, mode: modeChat,
		// Note: no local process tree running in daemon client mode
		conv: conv, client: llmClient, snap: snap, cfg: cfg,
		msgChan:     make(chan tea.Msg, 64),
		expandedMap: make(map[uint32]bool),
		focusPanel:  focusProcs,
	}
	if snap != nil {
		m.procList = buildProcTree(snap, m.expandedMap)
	}
	return m
}

func (m model) Init() tea.Cmd {
	// We deliberately omit textinput.Blink and m.spinner.Tick here.
	// Continuous commands cause the terminal to re-render in a loop,
	// which breaks native text-selection in standard terminal emulators.
	return nil
}

// ── Update ─────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		chatW := m.chatWidth()
		h := m.vpHeight()
		if !m.ready {
			m.vp = viewport.New(chatW, h)
			m.vp.SetContent(m.renderHistory())
			m.ready = true
		} else {
			m.vp.Width = chatW
			m.vp.Height = h
		}
		m.textInput.Width = chatW - 6
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
			m.procList = buildProcTree(msg.snap, m.expandedMap)
			m.procCursor = 0
			m.procScroll = 0
			if len(m.procList) > 0 {
				m.selectedPID = m.procList[0].PID
			}
			m.statusMsg = okStyle.Render("✓ snapshot loaded")
		}
		return m, nil
	case snapshotsListedMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("✕ loader: %v", msg.err))
		} else {
			m.snapMetas = msg.metas
			m.snapTable = buildSnapshotTable(m.snapMetas)
		}
	case snapshotDeletedMsg:
		if msg.err != nil {
			m.statusMsg = errorStyle.Render(fmt.Sprintf("✕ delete: %v", msg.err))
		} else {
			m.statusMsg = okStyle.Render(fmt.Sprintf("✓ deleted snapshot %d", msg.id))
			// Refresh the list automatically
			return m, func() tea.Msg {
				var metas []store.SnapshotMeta
				var err error
				if m.conv.DaemonClient != nil {
					metas, err = m.conv.DaemonClient.ListSnapshots()
				} else {
					metas, err = store.ListSnapshots()
				}
				return snapshotsListedMsg{metas: metas, err: err}
			}
		}
		m.mode = modeLoader
		m.textInput.Blur()
		return m, nil

	case tea.KeyMsg:
		if m.mode == modeLoader {
			return m.updateLoader(msg)
		}
		return m.updateChat(msg)

	case tea.MouseMsg:
		// Handle scroll wheel without requiring full mouse tracking
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.vp.LineUp(3)
		case tea.MouseButtonWheelDown:
			m.vp.LineDown(3)
		}
		return m, nil
	}

	if m.waiting {
		sp, spCmd := m.spinner.Update(msg)
		m.spinner = sp
		cmds = append(cmds, spCmd)
	}
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

func (m *model) chatWidth() int {
	if !m.showSidePanel {
		return m.width - 2 // small padding
	}
	w := m.width - leftPanelWidth - 1 // -1 for divider
	if w < 20 {
		w = 20
	}
	return w
}

func (m *model) vpHeight() int {
	// Calculate chrome:
	// 1 (header) + 1 (input) + 1 (bar) + 1 (margin/selInfo)
	chrome := 5
	if m.waiting {
		chrome += 1
	}
	if m.showSidePanel {
		// +3 for the Enter/Ctrl/Tips info block
		chrome += 4
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
	w := m.vp.Width
	if w <= 0 {
		w = m.chatWidth()
	}
	if w <= 2 {
		return strings.Join(m.history, "\n")
	}
	style := lipgloss.NewStyle().Width(w - 2)
	var wrapped []string
	for _, item := range m.history {
		wrapped = append(wrapped, style.Render(item))
	}
	return strings.Join(wrapped, "\n")
}

// ── Key handlers ───────────────────────────────────────────────────────────

func (m model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ── Global Hotkeys ───────────────────────────────────────────────────────
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		// Deselect process and blur text input (Normal mode navigation)
		m.selectedPID = 0
		m.textInput.Blur()
		m.textInput.SetValue("")
		m.statusMsg = statusStyle.Render("press i/Enter to chat")
		return m, nil
	// Ctrl+F removed, no longer needed as panels are completely unified
	case tea.KeyCtrlR:
		var newSnap *ctxengine.FrozenSnapshot
		if m.conv.DaemonClient != nil {
			var err error
			newSnap, err = m.conv.DaemonClient.GetSnapshot()
			if err != nil {
				m.statusMsg = errorStyle.Render("↻ refresh failed: " + err.Error())
				return m, nil
			}
		} else {
			newSnap = m.tree.TakeSnapshot()
		}
		m.snap = newSnap
		m.conv.UpdateSnapshot(newSnap)
		m.procList = buildProcTree(newSnap, m.expandedMap)
		m.procCursor = 0
		m.procScroll = 0
		m.statusMsg = okStyle.Render("↻ snapshot refreshed")
		return m, nil
	case tea.KeyCtrlS:
		snap := m.snap
		return m, func() tea.Msg {
			if m.conv.DaemonClient != nil {
				id, err := m.conv.DaemonClient.SaveSnapshot("manual", snap)
				return snapshotSavedMsg{id: id, err: err}
			}
			id, err := store.SaveSnapshot("manual", snap)
			return snapshotSavedMsg{id: id, err: err}
		}
	case tea.KeyCtrlL:
		return m, func() tea.Msg {
			if m.conv.DaemonClient != nil {
				metas, err := m.conv.DaemonClient.ListSnapshots()
				return snapshotsListedMsg{metas: metas, err: err}
			}
			metas, err := store.ListSnapshots()
			return snapshotsListedMsg{metas: metas, err: err}
		}
	//case tea.KeyCtrlP: // user requested get maps not on v1
	//	if m.selectedPID != 0 {
	//		return m.fireDirectTool("get_maps", map[string]interface{}{"pid": float64(m.selectedPID)})
	//	}
	// case tea.KeyCtrlT: //ssl intercept not on v1
	//	if m.selectedPID != 0 {
	//		return m.fireDirectTool("ssl_intercept", map[string]interface{}{"pid": float64(m.selectedPID), "duration": float64(10)})
	//	}
	//case tea.KeyCtrlY: // user requested get shell history not on v1
	//	if m.selectedPID != 0 {
	//		user := "root"
	//		return m.fireDirectTool("goread_shell_history", map[string]interface{}{"user": user, "limit": float64(50)})
	//	}
	//case tea.KeyCtrlA: // user requested add last direct result to context not on v1
	//	if m.lastDirectTool != "" && m.lastDirectResult != "" {
	//		// Add the last direct result directly to the conversation history
	//		m.conv.History = append(m.conv.History, llm.Message{
	//			Role:    llm.RoleSystem,
	//			Content: fmt.Sprintf("ACTIVE CONTEXT: User manually explicitly executed tool '%s' and wanted to add its result to your context:\n%s", m.lastDirectTool, m.lastDirectResult),
	//		})
	//		m.history = append(m.history, hintStyle.Render("✓ Added last inspect/tool result to godshell's context."))
	//		m.syncVP()
	//		return m, nil
	//	}
	case tea.KeyCtrlP:
		// toggle side panel
		if m.showSidePanel == false {
			m.showSidePanel = true
			m.focusPanel = focusProcs
			m.textInput.Blur()
		} else {
			m.showSidePanel = false
			m.focusPanel = focusChat
			m.textInput.Focus()
		}
		// RE-CALCULATE dimensions for components!
		newChatW := m.chatWidth()
		m.vp.Width = newChatW
		m.vp.Height = m.vpHeight()
		m.textInput.Width = newChatW - 6

		// Sync viewport on toggle
		m.syncVP()
		return m, tea.ClearScreen
	}

	// ── Unified Navigation & Operations ──────────────────────────────────────

	// If text input is NOT focused, allow single-key vim navigation
	if !m.textInput.Focused() {
		switch msg.String() {
		case "j", "down":
			if m.procCursor < len(m.procList)-1 {
				m.procCursor++
				if m.procCursor >= m.procScroll+m.vpHeight() {
					m.procScroll++
				}
				if len(m.procList) > 0 {
					m.selectedPID = m.procList[m.procCursor].PID
				}
			}
			return m, nil
		case "k", "up":
			if m.procCursor > 0 {
				m.procCursor--
				if m.procCursor < m.procScroll {
					m.procScroll--
				}
				if len(m.procList) > 0 {
					m.selectedPID = m.procList[m.procCursor].PID
				}
			}
			return m, nil
		case "i", "/":
			m.textInput.Focus()
			m.statusMsg = ""
			m.selectedPID = 0
			return m, nil
		}
	}

	// Always allow Ctrl+J/Ctrl+K and Up/Down for side panel regardless of focus
	if msg.Type == tea.KeyCtrlJ {
		if m.procCursor < len(m.procList)-1 {
			m.procCursor++
			if m.procCursor >= m.procScroll+m.vpHeight() {
				m.procScroll++
			}
			if len(m.procList) > 0 {
				m.selectedPID = m.procList[m.procCursor].PID
			}
		}
		return m, nil
	} else if msg.String() == "up" || msg.Type == tea.KeyCtrlK {
		if m.procCursor > 0 {
			m.procCursor--
			if m.procCursor < m.procScroll {
				m.procScroll--
			}
			if len(m.procList) > 0 {
				m.selectedPID = m.procList[m.procCursor].PID
			}
		}
		return m, nil
	}

	// Chat history scrolling
	if msg.Type == tea.KeyCtrlU || msg.Type == tea.KeyPgUp {
		m.vp.LineUp(3)
		return m, nil
	} else if msg.Type == tea.KeyCtrlD || msg.Type == tea.KeyPgDown {
		m.vp.LineDown(3)
		return m, nil
	}

	// Enter Key: Contextual Action
	if msg.Type == tea.KeyEnter {
		if !m.textInput.Focused() {
			m.textInput.Focus()
			m.statusMsg = ""
		}

		input := strings.TrimSpace(m.textInput.Value())
		m.textInput.SetValue("")
		m.statusMsg = ""

		if input != "" {
			// User typed a message, send it to LLM
			if m.selectedPID != 0 && m.snap != nil {
				var selectedComm string
				for _, e := range m.procList {
					if e.PID == m.selectedPID {
						selectedComm = e.Comm
						break
					}
				}
				if selectedComm != "" {
					m.conv.History = append(m.conv.History, llm.Message{
						Role:    llm.RoleSystem,
						Content: fmt.Sprintf("ACTIVE CONTEXT: User has selected process PID=%d (%s). Treat this as the subject of investigation unless overridden by the user's message.", m.selectedPID, selectedComm),
					})
				}
			}
			return m.handleInput(input)
		}
		return m, nil
	}

	// Pass any remaining typing directly to text input
	ti, cmd := m.textInput.Update(msg)
	m.textInput = ti
	return m, cmd
}

func (m model) updateLoader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeChat
		m.textInput.Focus()
		return m, nil
	case "enter":
		row := m.snapTable.Cursor()
		if row >= 0 && row < len(m.snapMetas) {
			meta := m.snapMetas[row]
			m.mode = modeChat
			m.textInput.Focus()
			return m, func() tea.Msg {
				if m.conv.DaemonClient != nil {
					snap, err := m.conv.DaemonClient.LoadSnapshot(meta.ID)
					return snapshotLoadedMsg{snap: snap, err: err}
				}
				snap, err := store.LoadSnapshot(meta.ID)
				return snapshotLoadedMsg{snap: snap, err: err}
			}
		}
	case "x", "d":
		row := m.snapTable.Cursor()
		if row >= 0 && row < len(m.snapMetas) {
			id := m.snapMetas[row].ID
			return m, func() tea.Msg {
				var err error
				if m.conv.DaemonClient != nil {
					err = m.conv.DaemonClient.DeleteSnapshot(id)
				} else {
					err = store.DeleteSnapshot(id)
				}
				return snapshotDeletedMsg{id: id, err: err}
			}
		}
	case "ctrl+c":
		return m, tea.Quit
	}
	t, cmd := m.snapTable.Update(msg)
	m.snapTable = t
	return m, cmd
}

//func (m model) updateConfirmContext(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
//	switch msg.String() {
//	case "y", "Y", "enter": // Accept
//		m.mode = modeChat
//		m.textInput.Focus()
//		if m.lastDirectTool != "" && m.lastDirectResult != "" {
//			m.conv.History = append(m.conv.History, llm.Message{
//				Role:    llm.RoleSystem,
//				Content: fmt.Sprintf("ACTIVE CONTEXT: User explicitly inspected process %d, here is the result:\n%s", m.selectedPID, m.lastDirectResult),
//			})
//			m.history = append(m.history, okStyle.Render("✓ Added process inspect output to context."))
//			m.syncVP()
//		}
//		return m, nil
//	case "n", "N", "esc", "ctrl+c": // Decline
//		m.mode = modeChat
//		m.textInput.Focus()
//		m.history = append(m.history, hintStyle.Render("✕ Ignored context addition."))
//		m.syncVP()
//		return m, nil
//	}
//	return m, nil
//}

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
		return m, tea.Batch(listenChan(ch), m.spinner.Tick)
	}
}

// ── View ───────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.mode == modeLoader {
		return m.viewLoader()
	} else if m.mode == modeConfirmContext {
		return m.viewConfirmContext()
	}
	return m.viewChat()
}

func (m model) viewConfirmContext() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("⚡ godshell popup") + "\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width)) + "\n\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF8866")).
		Padding(1, 2)

	b.WriteString(boxStyle.Render(m.confirmPrompt) + "\n\n")
	b.WriteString(statusStyle.Render("  [Y]es to add to context  |  [N]o to discard"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m model) viewChat() string {
	if !m.ready {
		return "initializing..."
	}

	// ── Left panel: process list ─────────────────────────────────────
	var leftPanel string
	if m.showSidePanel == true {

		leftPanel = m.renderProcPanel()
	}

	// ── Right panel: chat ────────────────────────────────────────────
	var chatB strings.Builder
	chatB.WriteString(headerStyle.Render(" godshell") + "\n")
	chatB.WriteString(m.vp.View() + "\n")
	if m.waiting {
		label := "Scanning system snapshot..."
		if m.activeTool != "" {
			label = "Running " + m.activeTool + "..."
		}
		chatB.WriteString("  " + m.spinner.View() + " " + scanStyle.Render(label) + "\n")
	}
	chatB.WriteString(m.textInput.View() + "\n")
	scroll := ""
	if m.vp.TotalLineCount() > m.vp.Height {
		pct := int(m.vp.ScrollPercent() * 100)
		scroll = dimStyle.Render(fmt.Sprintf(" %d%%", pct))
	}
	selInfo := ""
	if m.selectedPID != 0 {
		for _, e := range m.procList {
			if e.PID == m.selectedPID {
				selInfo = okStyle.Render(fmt.Sprintf("[sel:%s/%d]", e.Comm, e.PID)) + "  "
				break
			}
		}
	}
	var bar string

	if m.showSidePanel {
		bar = statusStyle.Render("Ctrl+([r]efresh [s]ave [l]oad [p]anel) Esc to toggle focus") + scroll
	} else {
		bar = statusStyle.Render("Ctrl+([r]efresh [s]ave [l]oad [p]anel) type help for more info, Esc to toggle focus") + scroll
	}

	if m.statusMsg != "" {
		bar = m.statusMsg + "  │  " + bar
	}

	chatB.WriteString(bar)
	chatB.WriteString("\n")
	chatB.WriteString(selInfo)

	if m.showSidePanel {
		tips := statusStyle.Render("Tip: If select not working, move across the processes a bit")

		chatB.WriteString("\n")
		chatB.WriteString("Enter to select a process, Esc to deselect\n")
		chatB.WriteString("Ctrl + j/k to move across processes or up/down keys\n")
		chatB.WriteString(tips)
	}

	// ── Side-by-side join ────────────────────────────────────────────
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		dimStyle.Render("│"),
		chatB.String(),
	)
}

func (m model) viewLoader() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("Snapshot Loader") + "\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width)) + "\n\n")
	if len(m.snapMetas) == 0 {
		b.WriteString(hintStyle.Render("  no snapshots saved. press [s] in chat to save one.") + "\n")
	} else {
		b.WriteString(m.snapTable.View() + "\n")
	}
	b.WriteString("\n" + statusStyle.Render("↑↓ navigate  enter load  [x] delete  esc back") + "\n")
	return b.String()
}

// ── Streaming chat loop ────────────────────────────────────────────────────

func streamingChatLoop(client llm.Client, conv *llm.Conversation, ch chan<- tea.Msg) {
	for {
		// Ensure history is within limits before chatting
		conv.ManageHistory()

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

			// Check if this is a report rendering tool
			var card string
			switch tc.Function.Name {
			case "report_text":
				card = renderReportText(args)
			case "report_behaviour":
				card = renderReportBehaviour(args)
			case "report_family":
				card = renderReportFamily(args)
			case "report_network":
				card = renderReportNetwork(args)
			case "report_threat":
				card = renderReportThreat(args)
			case "report_system_state":
				card = renderReportSystemState(args)
			default:
				// Build a standard red pill card with result metadata
				card = buildToolCard(tc.Function.Name, args, result)
			}

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
		return "PID " + argStr(args, "pid") + " · Inspect Completed" // Replaced by beautiful card for the chat body directly
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
	case "ssl_intercept":
		return metaSSLIntercept(result, args)
	case "scan_heap":
		return metaScanHeap(result, args)
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

func metaSSLIntercept(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	lines := nonEmptyLines(result)
	// Look for "Endpoints discovered: N" line
	for _, line := range lines {
		if strings.HasPrefix(line, "Endpoints discovered:") {
			parts := []string{"PID " + pid}
			parts = append(parts, toolCardHighStyle.Render(strings.TrimPrefix(line, "Endpoints discovered: ")+" endpoints"))
			// Check for unauthenticated signal
			for _, l := range lines {
				if strings.Contains(l, "Unauthenticated endpoint") {
					parts = append(parts, errorStyle.Render("⚠ unauth routes detected"))
					break
				}
			}
			return strings.Join(parts, " · ")
		}
	}
	return "PID " + pid + " · " + trunc(firstNonEmptyLine(result), 50)
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

func renderBeautifulInspect(result string, pid string) string {
	var data struct {
		Comm     string  `json:"comm"`
		Binary   string  `json:"binary_path"`
		State    string  `json:"state"`
		CPU      float64 `json:"cpu_usage_percent"`
		MemKB    uint64  `json:"memory_usage_kb"`
		Parent   string  `json:"parent"`
		Uptime   string  `json:"uptime"`
		Children []struct {
			PID  int    `json:"pid"`
			Comm string `json:"comm"`
		} `json:"children"`
		NetEffects  []ctxengine.EffectJSON `json:"network_effects"`
		FileEffects []ctxengine.EffectJSON `json:"file_effects"`
	}

	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return errorStyle.Render("Failed to parse inspect output")
	}

	var sb strings.Builder

	title := toolCardNameStyle.Render(fmt.Sprintf(" ◎ Inspect PID %s ", pid))
	nameText := data.Comm
	if nameText == "" {
		nameText = data.Binary
	}

	sb.WriteString(title + " " + shellStyle.Render(nameText) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")

	sb.WriteString(fmt.Sprintf("%s %s\n", toolCardHighStyle.Render("State: "), data.State))
	sb.WriteString(fmt.Sprintf("%s %.1f%%\n", toolCardHighStyle.Render("CPU:   "), data.CPU))
	sb.WriteString(fmt.Sprintf("%s %d KB\n", toolCardHighStyle.Render("Mem:   "), data.MemKB))
	if data.Parent != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", toolCardHighStyle.Render("Parent:"), data.Parent))
	}

	if len(data.Children) > 0 {
		sb.WriteString(fmt.Sprintf("%s (%d)\n", toolCardHighStyle.Render("Children:"), len(data.Children)))
		max := len(data.Children)
		if max > 5 {
			max = 5
		} // show up to 5
		for i := 0; i < max; i++ {
			c := data.Children[i]
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  ├─ %d: %s\n", c.PID, c.Comm)))
		}
		if len(data.Children) > 5 {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  └─ ... %d more\n", len(data.Children)-5)))
		}
	}

	if len(data.NetEffects) > 0 {
		sb.WriteString(fmt.Sprintf("%s (%d groups)\n", toolCardHighStyle.Render("Network:"), len(data.NetEffects)))
		for _, e := range data.NetEffects {
			sb.WriteString(hintStyle.Render(fmt.Sprintf("  • %s ×%d\n", e.Label, e.Count)))
		}
	}

	if len(data.FileEffects) > 0 {
		sb.WriteString(fmt.Sprintf("%s (%d groups)\n", toolCardHighStyle.Render("Files:"), len(data.FileEffects)))
		max := len(data.FileEffects)
		if max > 5 {
			max = 5
		}
		for i := 0; i < max; i++ {
			e := data.FileEffects[i]
			sb.WriteString(hintStyle.Render(fmt.Sprintf("  • %s ×%d\n", e.Label, e.Count)))
		}
		if len(data.FileEffects) > 5 {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  └─ ... %d more groups\n", len(data.FileEffects)-5)))
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).Render(sb.String())
}

// ── LLM Report Renderers ───────────────────────────────────────────────────

func renderReportText(args map[string]interface{}) string {
	title := argStr(args, "title")
	text := argStr(args, "text")

	var sb strings.Builder
	sb.WriteString(toolCardNameStyle.Render(fmt.Sprintf(" 📝 %s ", title)) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")
	sb.WriteString(text + "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).Render(sb.String())
}

func renderReportBehaviour(args map[string]interface{}) string {
	procName := argStr(args, "process_name")
	reasons := argStr(args, "reasons")

	var sb strings.Builder
	sb.WriteString(toolCardNameStyle.Render(fmt.Sprintf(" 🧠 Behaviour Analysis: %s ", procName)) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")

	if raw, ok := args["insights"].([]interface{}); ok {
		for _, r := range raw {
			if str, ok := r.(string); ok {
				sb.WriteString(hintStyle.Render("  • "+str) + "\n")
			}
		}
	}
	sb.WriteString("\n" + toolCardHighStyle.Render("Conclusion/Reasons: ") + reasons + "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).Render(sb.String())
}

func renderReportFamily(args map[string]interface{}) string {
	target := argStr(args, "target_process")

	var sb strings.Builder
	sb.WriteString(toolCardNameStyle.Render(fmt.Sprintf(" ⎇ Lineage: %s ", target)) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")

	if raw, ok := args["parents"].([]interface{}); ok && len(raw) > 0 {
		sb.WriteString(toolCardHighStyle.Render("Ancestors:\n"))
		for _, r := range raw {
			if str, ok := r.(string); ok {
				sb.WriteString(dimStyle.Render("  ↑ "+str) + "\n")
			}
		}
	}

	sb.WriteString(selStyle.Render("  "+target) + "\n")

	if raw, ok := args["children"].([]interface{}); ok && len(raw) > 0 {
		sb.WriteString(toolCardHighStyle.Render("Descendants:\n"))
		for _, r := range raw {
			if str, ok := r.(string); ok {
				sb.WriteString(dimStyle.Render("  ↓ "+str) + "\n")
			}
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).Render(sb.String())
}

func renderReportNetwork(args map[string]interface{}) string {
	summary := argStr(args, "summary")

	var sb strings.Builder
	sb.WriteString(toolCardNameStyle.Render(" 🔗 Network Intelligence ") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")
	sb.WriteString(summary + "\n\n")

	if raw, ok := args["connections"].([]interface{}); ok {
		sb.WriteString(toolCardHighStyle.Render("Connections:\n"))
		for _, r := range raw {
			if cmap, ok := r.(map[string]interface{}); ok {
				addr := argStr(cmap, "address")
				state := argStr(cmap, "state")
				sus, _ := cmap["is_suspicious"].(bool)

				prefix := "  ○ "
				if sus {
					prefix = errorStyle.Render("  ⚠ ")
				}
				sb.WriteString(fmt.Sprintf("%s%-20s %s\n", prefix, addr, dimStyle.Render(state)))
			}
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).Render(sb.String())
}

func renderReportThreat(args map[string]interface{}) string {
	explanation := argStr(args, "threat_explanation")
	harm := argStr(args, "harm_level")
	conclusion := argStr(args, "conclusion")

	var sb strings.Builder
	titleBar := " 🛡️ Threat Assessment "
	if harm == "High" || harm == "Critical" {
		titleBar = errorStyle.Render(" ☠ CRITICAL THREAT DETECTED ")
	} else if harm == "Medium" {
		titleBar = " ⚠ Suspicious Activity "
	}
	sb.WriteString(titleBar + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")

	sb.WriteString(toolCardHighStyle.Render(fmt.Sprintf("Level: %s", harm)) + "\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", explanation))

	if raw, ok := args["involved_processes"].([]interface{}); ok && len(raw) > 0 {
		sb.WriteString(toolCardHighStyle.Render("Involved Processes:\n"))
		for _, r := range raw {
			if str, ok := r.(string); ok {
				sb.WriteString(dimStyle.Render("  • "+str) + "\n")
			}
		}
	}

	if raw, ok := args["involved_effects"].([]interface{}); ok && len(raw) > 0 {
		sb.WriteString(toolCardHighStyle.Render("Involved Effects:\n"))
		for _, r := range raw {
			if str, ok := r.(string); ok {
				sb.WriteString(dimStyle.Render("  • "+str) + "\n")
			}
		}
	}

	sb.WriteString("\n" + errorStyle.Render("Verdict: ") + conclusion + "\n")

	borderColor := "#555555"
	if harm == "High" || harm == "Critical" {
		borderColor = "#FF0000"
	} else if harm == "Medium" {
		borderColor = "#FFAA00"
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).Render(sb.String())
}

func renderReportSystemState(args map[string]interface{}) string {
	susp := argStr(args, "suspicious_behaviour")
	work := argStr(args, "user_work_summary")
	ghosts := argStr(args, "ghosts_summary")

	threatIdx := 0
	if val, ok := args["threat_index"].(float64); ok {
		threatIdx = int(val)
	}

	var sb strings.Builder
	sb.WriteString(toolCardNameStyle.Render(fmt.Sprintf(" ◈ System State (Threat Index: %d/10) ", threatIdx)) + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", 40)) + "\n")

	sb.WriteString(toolCardHighStyle.Render("Current Activity: ") + work + "\n\n")

	if raw, ok := args["top_processes"].([]interface{}); ok && len(raw) > 0 {
		sb.WriteString(toolCardHighStyle.Render("Top Processes:\n"))
		for _, r := range raw {
			if str, ok := r.(string); ok {
				sb.WriteString(dimStyle.Render("  • "+str) + "\n")
			}
		}
		sb.WriteString("\n")
	}

	if susp != "" {
		sb.WriteString(toolCardHighStyle.Render("Anomalies: ") + errorStyle.Render(susp) + "\n\n")
	}

	if ghosts != "" {
		sb.WriteString(toolCardHighStyle.Render("Ghosts / Recent Exits: ") + hintStyle.Render(ghosts) + "\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		Padding(0, 1).Render(sb.String())
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
	case "scan_heap":
		return "🗜"
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

// ANSI escape sequence regex
var ansiRegex = `\x1b\[[0-9;]*m`
var _ansiRegexCompiled *strings.Replacer

func stripAnsi(str string) string {
	// Simple manual strip, or use regex
	// A basic implementation ignoring escape codes for the length requirement
	var b strings.Builder
	inEsc := false
	for _, ch := range str {
		if ch == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '~' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

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

// ── Process panel helpers ──────────────────────────────────────────────────

// buildProcTree materializes a flattened, viewable list of procEntry from a snapshot,
// respecting the tree hierarchy and expansion state.
func buildProcTree(snap *ctxengine.FrozenSnapshot, expanded map[uint32]bool) []procEntry {
	if snap == nil {
		return nil
	}

	// 1. Find all root nodes (processes where PPID is 0, or PPID is not in snapshot)
	// and build a map of parent -> sorted children.
	// Since ByPID is a map, we need to sort to make the tree stable.
	var roots []uint32
	childrenMap := make(map[uint32][]uint32)

	// Collect and sort all PIDs for stable iteration
	var allPIDs []uint32
	for pid := range snap.ByPID {
		allPIDs = append(allPIDs, pid)
	}
	// Sort by PID ascending
	for i := 1; i < len(allPIDs); i++ {
		for j := i; j > 0 && allPIDs[j] < allPIDs[j-1]; j-- {
			allPIDs[j], allPIDs[j-1] = allPIDs[j-1], allPIDs[j]
		}
	}

	for _, pid := range allPIDs {
		p := snap.ByPID[pid]
		_, hasParent := snap.ByPID[p.PPID]
		if p.PPID == 0 || p.PPID == 2 || !hasParent {
			roots = append(roots, pid)
		} else {
			childrenMap[p.PPID] = append(childrenMap[p.PPID], pid)
		}
	}

	var entries []procEntry
	var walk func(pid uint32, depth int)
	walk = func(pid uint32, depth int) {
		pStr := snap.ByPID[pid]
		kids, hasKids := childrenMap[pid]

		e := procEntry{
			PID:     pid,
			Comm:    pStr.Comm,
			Depth:   depth,
			HasKids: hasKids,
		}
		if pStr.MemoryUsage > 1024 {
			e.Mem = fmt.Sprintf("%dM", pStr.MemoryUsage/1024)
		} else if pStr.MemoryUsage > 0 {
			e.Mem = fmt.Sprintf("%dK", pStr.MemoryUsage)
		}
		if pStr.CpuUsage > 0.1 {
			e.CPU = fmt.Sprintf("%.0f%%", pStr.CpuUsage)
		}
		entries = append(entries, e)

		// Recursive step
		if expanded[pid] && hasKids {
			for _, childPID := range kids {
				walk(childPID, depth+1)
			}
		}
	}

	for _, rootPID := range roots {
		walk(rootPID, 0)
	}

	return entries
}

// renderProcPanel renders the left process-tree panel as a fixed-width block.
func (m model) renderProcPanel() string {
	panelW := leftPanelWidth
	panelH := m.height - 1 // leave 1 row for status bar

	// Styles
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	// selStyle is now a global style
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	dimRowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	dimTargetStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	var lines []string

	titleText := "PROCESSES"
	title := titleStyle.Render(titleText)
	lines = append(lines, title)

	borderColor := "#FF4444"
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor)).Render(strings.Repeat("─", panelW)))

	if len(m.procList) == 0 {
		lines = append(lines, dimRowStyle.Render("  no snapshot"))
	} else {
		visible := panelH - 2 // title + divider
		if visible < 1 {
			visible = 1
		}
		end := m.procScroll + visible
		if end > len(m.procList) {
			end = len(m.procList)
		}
		for i := m.procScroll; i < end; i++ {
			e := m.procList[i]

			// Build tree prefix
			var prefix string
			if e.Depth > 0 {
				prefix = strings.Repeat("  ", e.Depth-1) + "├─ "
			}
			var toggle string
			if e.HasKids {
				if m.expandedMap[e.PID] {
					toggle = "▼ "
				} else {
					toggle = "▶ "
				}
			} else {
				toggle = "  "
			}

			// Calculate remaining space for Comm
			comm := e.Comm
			maxCommLen := 14 - len(prefix) - len(toggle)
			if maxCommLen < 4 {
				maxCommLen = 4
			}
			if len(comm) > maxCommLen {
				comm = comm[:maxCommLen-1] + "…"
			}

			pidStr := fmt.Sprintf("%5d", e.PID)
			meta := ""
			if e.CPU != "" {
				meta += e.CPU + " "
			}
			if e.Mem != "" {
				meta += e.Mem
			}

			// Pad row to panel width
			treeComm := fmt.Sprintf("%s%s%s", prefix, toggle, comm)
			rowMain := fmt.Sprintf(" %-14s %s", treeComm, pidStr)

			// Apply dim styling to tree prefixes for visual clarity
			rowMain = strings.Replace(rowMain, "├─", dimTargetStyle.Render("├─"), 1)

			rowMeta := metaStyle.Render(fmt.Sprintf("%-6s", meta))

			var row string
			if i == m.procCursor {
				row = selStyle.Render(fmt.Sprintf("%-*s", panelW-7, stripAnsi(rowMain))) + rowMeta
			} else {
				row = normalStyle.Render(fmt.Sprintf("%-*s", panelW-7, rowMain)) + rowMeta
			}
			lines = append(lines, row)
		}
		// Scroll indicator
		if len(m.procList) > (panelH - 2) {
			scrollPct := 0
			if len(m.procList) > 1 {
				scrollPct = m.procCursor * 100 / (len(m.procList) - 1)
			}
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor)).Render(fmt.Sprintf("  ── %d%% ──", scrollPct)))
		}
	}

	// Fill remaining height with empty lines
	for len(lines) < panelH {
		lines = append(lines, "")
	}

	return lipgloss.NewStyle().
		Width(panelW).
		MaxWidth(panelW).
		Render(strings.Join(lines, "\n"))
}

// fireDirectTool fires a tool locally and synchronously, skipping the LLM,
// and renders the output directly into the chat history view as a beautiful card.
func (m *model) fireDirectTool(toolName string, args map[string]interface{}) (tea.Model, tea.Cmd) {
	if m.snap == nil {
		m.history = append(m.history, errorStyle.Render("No snapshot available."))
		m.syncVP()
		return m, nil
	}

	cmdLabel := fmt.Sprintf("%s(pid=%v)", toolName, args["pid"])
	m.history = append(m.history, promptStyle.Render("you❯ ")+cmdLabel)

	m.waiting = true
	m.activeTool = toolName
	m.syncVP()

	// Run tool in background to avoid blocking the UI
	return m, func() tea.Msg {
		var output string
		var err error

		if m.conv.DaemonClient != nil && llm.IsPrivilegedTool(toolName) {
			// Ask daemon to run it
			output, err = m.conv.DaemonClient.ExecuteTool(toolName, args)
			if err != nil {
				output = fmt.Sprintf("Error executing tool on daemon: %v", err)
			}
		} else {
			// Run locally
			output, err = llm.ExecuteToolOnSnapshot(toolName, args, m.snap)
			if err != nil {
				output = fmt.Sprintf("Error executing tool locally: %v", err)
			}
		}

		// Save the output so user can add it to context with Ctrl+A or prompt
		result := directToolResultMsg{
			toolName: toolName,
			output:   output,
			args:     args,
		}
		return result
	}
}

type directToolResultMsg struct {
	toolName string
	output   string
	args     map[string]interface{}
}

//func (m model) updateDirectToolResult(msg directToolResultMsg) (tea.Model, tea.Cmd) {
//	m.waiting = false
//	m.activeTool = ""
//	m.lastDirectTool = msg.toolName
//	m.lastDirectResult = msg.output

//	if msg.toolName == "inspect" {
//		pid := argStr(msg.args, "pid")
//		card := renderBeautifulInspect(msg.output, pid)
//		m.history = append(m.history, card)
//		m.syncVP()

// Setup the confirmation popup
//		m.mode = modeConfirmContext
//		m.confirmPrompt = fmt.Sprintf("Add output of `inspect PID %s` to godshell's LLM context?", pid)
//		m.textInput.Blur()
//		return m, nil
//	}

// Not inspect, just dump it
//		m.history = append(m.history, msg.output)
//	m.syncVP()
//	return m, nil
//}

// ── Help ───────────────────────────────────────────────────────────────────

func helpText() string {
	return hintStyle.Render(strings.Join([]string{
		"hotkeys:",
		"  ctrl+esc   — toggle focus betweeen processes and chat",
		"  ctrl+c   — quit godshell",
		"  ctrl+r   — refresh system snapshot",
		"  ctrl+s   — save snapshot to SQLite",
		"  ctrl+l   — load a saved snapshot",
		"",
		"  type clear to clear chat history",
		"",
		"process panel navigation:",
		"  ctrl+j   — navigate down",
		"  ctrl+k   — navigate up",
		"  enter    — select process for LLM to focus on.",
		"",
		"chat panel controls (when focused):",
		"  pgup     — scroll history up",
		"  pgdown   — scroll history down",
		"  mouse up/down - scroll up/down",
	}, "\n"))
}

// ── main ───────────────────────────────────────────────────────────────────

func safePrefix(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:8] + "…"
}

func runDaemon() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v (using defaults)\n", err)
		cfg = config.Default()
	}

	if err := store.Init(cfg.DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: store init: %v (save/load disabled)\n", err)
	}

	// Change ownership/permissions so non-root can read/write if they access it directly later
	_ = os.Chmod(cfg.DBPath, 0666)

	tree := ctxengine.NewProcessTree(cfg)
	events := make(chan observer.Event, 1000)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := observer.Run(ctx, events); err != nil {
			fmt.Fprintf(os.Stderr, "observer error (are you root?): %v\n", err)
			os.Exit(1)
		}
	}()
	go func() {
		for e := range events {
			tree.HandleEvent(e)
		}
	}()
	go tree.RefreshMetrics(2 * time.Second)
	go tree.EvictGhosts(time.Duration(cfg.GhostProcessRetentionSec) * time.Second)

	// Auto-snapshotting every SnapshotIntervalMin (if > 0)
	if cfg.SnapshotIntervalMin > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.SnapshotIntervalMin) * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				snap := tree.TakeSnapshot()
				_, _ = store.SaveSnapshot("auto", snap)
			}
		}()
	}

	// Periodic DB pruning every hour
	if cfg.DBSnapshotRetentionHours > 0 {
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				_, _ = store.PruneSnapshots(cfg.DBSnapshotRetentionHours)
			}
		}()
	}

	srv := daemon.NewServer(cfg.SocketPath, tree)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "daemon server error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "daemon":
			runDaemon()
			return
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
			fmt.Println("  (none)    Start the AI REPL (connects to daemon socket)")
			fmt.Println("  daemon    Run the eBPF observer background process (requires sudo)")
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

	// Connect to Daemon
	daemonClient := client.NewDaemonClient(cfg.SocketPath)
	snap, err := daemonClient.GetSnapshot()
	if err != nil {
		fmt.Fprint(os.Stderr, errorStyle.Render("✗ Failed to connect to godshell daemon")+"\n")
		fmt.Fprint(os.Stderr, hintStyle.Render("  Ensure the daemon is running in another terminal:\n"))
		fmt.Fprint(os.Stderr, hintStyle.Render("    sudo godshell daemon\n"))
		fmt.Fprint(os.Stderr, hintStyle.Render(fmt.Sprintf("  Error details: %v\n", err)))
		os.Exit(1)
	}

	fmt.Printf("⚡ godshell  model=%s  key_len=%d  key_prefix=%s\n", cfg.Model, len(cfg.APIKey), safePrefix(cfg.APIKey))
	llmClient := llm.NewOpenAIClient(cfg.APIKey, cfg.Endpoint, cfg.Model)

	m := initialDaemonModel(daemonClient, llmClient, snap, cfg)
	// WithAltScreen: full-screen TUI
	// WithMouseCellMotion: allows mouse scrolling and clicking (Hold Shift to do native terminal select)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
func metaScanHeap(result string, args map[string]interface{}) string {
	pid := argStr(args, "pid")
	lines := nonEmptyLines(result)

	findings := 0
	ips := 0
	urls := 0
	keys := 0

	for _, line := range lines {
		if strings.Contains(line, "IP_ADDR") {
			ips++
			findings++
		}
		if strings.Contains(line, "URL") {
			urls++
			findings++
		}
		if strings.Contains(line, "POTENT_KEY") {
			keys++
			findings++
		}
	}

	if findings == 0 {
		return "PID " + pid + " · No patterns found"
	}

	parts := []string{"PID " + pid}
	if ips > 0 {
		parts = append(parts, fmt.Sprintf("%d IPs", ips))
	}
	if urls > 0 {
		parts = append(parts, fmt.Sprintf("%d URLs", urls))
	}
	if keys > 0 {
		parts = append(parts, toolCardHighStyle.Render(fmt.Sprintf("%d Keys", keys)))
	}

	return strings.Join(parts, " · ")
}
