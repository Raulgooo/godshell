package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"godshell/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ─────────────────────────────────────────────────────────────────

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AA88FF"))
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(24)
	activeLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true).Width(24)
	savedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	hintStyle2  = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true)
)

// field IDs
const (
	fieldAPIKey = iota
	fieldModel
	fieldEndpoint
	fieldSnapshotIntervalMin
	fieldGhostProcessRetention
	fieldDBSnapshotRetention
	fieldSnapshotConcurrency
	fieldDBPath
	fieldSocketPath
	fieldProcPath
	fieldSysPath
	fieldEnrichmentWorkers
	fieldMaxEffectsPerProcess
	fieldCaptureNetwork
	fieldCaptureFileIO
	fieldIgnoredProcesses
	fieldCount
)

var fieldNames = [fieldCount]string{
	"API Key",
	"Model",
	"Endpoint",
	"Snapshot Interval (min, auto-save)",
	"Ghost Process Retention (sec)",
	"DB Snapshot Retention (hours)",
	"Snapshot Concurrency",
	"DB Path",
	"Socket Path",
	"Proc Path",
	"Sys Path",
	"Enrichment Workers",
	"Max Effects Per Process",
	"Capture Network (true/false)",
	"Capture File I/O (true/false)",
	"Ignored Processes (CSV)",
}

type configModel struct {
	inputs  [fieldCount]textinput.Model
	focused int
	saved   bool
	err     string
}

func newConfigModel(cfg config.Config) configModel {
	var inputs [fieldCount]textinput.Model

	for i := 0; i < fieldCount; i++ {
		t := textinput.New()
		t.Prompt = ""
		t.CharLimit = 256
		t.Width = 50
		inputs[i] = t
	}

	// Pre-populate
	inputs[fieldAPIKey].SetValue(cfg.APIKey)
	inputs[fieldAPIKey].EchoMode = textinput.EchoPassword
	inputs[fieldAPIKey].EchoCharacter = '•'
	inputs[fieldAPIKey].Placeholder = "sk-or-..."

	inputs[fieldModel].SetValue(cfg.Model)
	inputs[fieldModel].Placeholder = "google/gemini-3.1-flash-lite-preview"

	inputs[fieldEndpoint].SetValue(cfg.Endpoint)
	inputs[fieldEndpoint].Placeholder = "https://openrouter.ai/api/v1/chat/completions"

	inputs[fieldSnapshotIntervalMin].SetValue(strconv.Itoa(cfg.SnapshotIntervalMin))
	inputs[fieldSnapshotIntervalMin].Placeholder = "0 (manual only)"

	inputs[fieldGhostProcessRetention].SetValue(strconv.Itoa(cfg.GhostProcessRetentionSec))
	inputs[fieldGhostProcessRetention].Placeholder = "60"

	inputs[fieldDBSnapshotRetention].SetValue(strconv.Itoa(cfg.DBSnapshotRetentionHours))
	inputs[fieldDBSnapshotRetention].Placeholder = "24"

	inputs[fieldSnapshotConcurrency].SetValue(strconv.Itoa(cfg.SnapshotConcurrency))
	inputs[fieldSnapshotConcurrency].Placeholder = "1"

	inputs[fieldDBPath].SetValue(cfg.DBPath)
	inputs[fieldDBPath].Placeholder = "~/.config/godshell/godshell.db"

	inputs[fieldSocketPath].SetValue(cfg.SocketPath)
	inputs[fieldSocketPath].Placeholder = "/var/run/godshell.sock"

	inputs[fieldProcPath].SetValue(cfg.ProcPath)
	inputs[fieldProcPath].Placeholder = "/proc"

	inputs[fieldSysPath].SetValue(cfg.SysPath)
	inputs[fieldSysPath].Placeholder = "/sys"

	inputs[fieldEnrichmentWorkers].SetValue(strconv.Itoa(cfg.EnrichmentWorkers))
	inputs[fieldEnrichmentWorkers].Placeholder = "4"

	inputs[fieldMaxEffectsPerProcess].SetValue(strconv.Itoa(cfg.MaxEffectsPerProcess))
	inputs[fieldMaxEffectsPerProcess].Placeholder = "1000"

	inputs[fieldCaptureNetwork].SetValue(strconv.FormatBool(cfg.CaptureNetwork))
	inputs[fieldCaptureNetwork].Placeholder = "true"

	inputs[fieldCaptureFileIO].SetValue(strconv.FormatBool(cfg.CaptureFileIO))
	inputs[fieldCaptureFileIO].Placeholder = "true"

	inputs[fieldIgnoredProcesses].SetValue(strings.Join(cfg.IgnoredProcesses, ","))
	inputs[fieldIgnoredProcesses].Placeholder = "godshell,systemd"

	inputs[0].Focus()

	return configModel{inputs: inputs, focused: 0}
}

func (m configModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			return m, tea.Quit

		case tea.KeyCtrlS, tea.KeyEnter:
			// Save
			cfg, err := m.toConfig()
			if err != nil {
				m.err = err.Error()
				return m, nil
			}
			if err := config.Save(cfg); err != nil {
				m.err = err.Error()
				return m, nil
			}
			m.saved = true
			m.err = ""
			return m, tea.Quit

		case tea.KeyTab, tea.KeyDown:
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % fieldCount
			m.inputs[m.focused].Focus()
			return m, nil

		case tea.KeyShiftTab, tea.KeyUp:
			m.inputs[m.focused].Blur()
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			m.inputs[m.focused].Focus()
			return m, nil
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m configModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("⚙  godshell config"))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 56))
	b.WriteString("\n\n")

	for i := 0; i < fieldCount; i++ {
		label := labelStyle.Render(fieldNames[i] + ":")
		if i == m.focused {
			label = activeLabel.Render("▸ " + fieldNames[i] + ":")
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", label, m.inputs[i].View()))
	}

	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 56))
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString(errStyle.Render("  ✗ " + m.err))
		b.WriteString("\n")
	}

	if m.saved {
		b.WriteString(savedStyle.Render("  ✓ saved to " + config.ConfigPath()))
		b.WriteString("\n")
	}

	b.WriteString(hintStyle2.Render("  tab/↑↓ navigate • enter/ctrl+s save • esc quit"))
	b.WriteString("\n")

	return b.String()
}

func (m configModel) toConfig() (config.Config, error) {
	interval, _ := strconv.Atoi(strings.TrimSpace(m.inputs[fieldSnapshotIntervalMin].Value()))
	ghostRetention, _ := strconv.Atoi(strings.TrimSpace(m.inputs[fieldGhostProcessRetention].Value()))
	dbRetention, _ := strconv.Atoi(strings.TrimSpace(m.inputs[fieldDBSnapshotRetention].Value()))
	concurrency, _ := strconv.Atoi(strings.TrimSpace(m.inputs[fieldSnapshotConcurrency].Value()))

	dbPath := strings.TrimSpace(m.inputs[fieldDBPath].Value())

	workers, _ := strconv.Atoi(strings.TrimSpace(m.inputs[fieldEnrichmentWorkers].Value()))
	maxEffs, _ := strconv.Atoi(strings.TrimSpace(m.inputs[fieldMaxEffectsPerProcess].Value()))
	capNet, _ := strconv.ParseBool(strings.TrimSpace(m.inputs[fieldCaptureNetwork].Value()))
	capFile, _ := strconv.ParseBool(strings.TrimSpace(m.inputs[fieldCaptureFileIO].Value()))

	ignoredStr := strings.TrimSpace(m.inputs[fieldIgnoredProcesses].Value())
	var ignored []string
	if ignoredStr != "" {
		for _, s := range strings.Split(ignoredStr, ",") {
			ignored = append(ignored, strings.TrimSpace(s))
		}
	}

	return config.Config{
		APIKey:                   strings.TrimSpace(m.inputs[fieldAPIKey].Value()),
		Model:                    strings.TrimSpace(m.inputs[fieldModel].Value()),
		Endpoint:                 strings.TrimSpace(m.inputs[fieldEndpoint].Value()),
		SnapshotIntervalMin:      interval,
		GhostProcessRetentionSec: ghostRetention,
		DBSnapshotRetentionHours: dbRetention,
		SnapshotConcurrency:      concurrency,
		DBPath:                   dbPath,
		SocketPath:               strings.TrimSpace(m.inputs[fieldSocketPath].Value()),
		ProcPath:                 strings.TrimSpace(m.inputs[fieldProcPath].Value()),
		SysPath:                  strings.TrimSpace(m.inputs[fieldSysPath].Value()),
		EnrichmentWorkers:        workers,
		MaxEffectsPerProcess:     maxEffs,
		CaptureNetwork:           capNet,
		CaptureFileIO:            capFile,
		IgnoredProcesses:         ignored,
	}, nil
}

// RunConfigEditor is the entry point called from main.
func RunConfigEditor() error {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("warning: %v (using defaults)\n", err)
		cfg = config.Default()
	}

	m := newConfigModel(cfg)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
