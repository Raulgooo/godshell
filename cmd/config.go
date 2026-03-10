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
	fieldSnapshotInterval
	fieldSnapshotRetention
	fieldSnapshotConcurrency
	fieldDBPath
	fieldCount
)

var fieldNames = [fieldCount]string{
	"API Key",
	"Model",
	"Endpoint",
	"Snapshot Interval (sec)",
	"Snapshot Retention (sec)",
	"Snapshot Concurrency",
	"DB Path",
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

	inputs[fieldSnapshotInterval].SetValue(strconv.Itoa(cfg.SnapshotIntervalSec))
	inputs[fieldSnapshotInterval].Placeholder = "0 (manual only)"

	inputs[fieldSnapshotRetention].SetValue(strconv.Itoa(cfg.SnapshotRetentionSec))
	inputs[fieldSnapshotRetention].Placeholder = "60"

	inputs[fieldSnapshotConcurrency].SetValue(strconv.Itoa(cfg.SnapshotConcurrency))
	inputs[fieldSnapshotConcurrency].Placeholder = "1"

	inputs[fieldDBPath].SetValue(cfg.DBPath)
	inputs[fieldDBPath].Placeholder = "~/.config/godshell/godshell.db"

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
	interval, err := strconv.Atoi(strings.TrimSpace(m.inputs[fieldSnapshotInterval].Value()))
	if err != nil {
		return config.Config{}, fmt.Errorf("snapshot interval must be a number")
	}
	retention, err := strconv.Atoi(strings.TrimSpace(m.inputs[fieldSnapshotRetention].Value()))
	if err != nil {
		return config.Config{}, fmt.Errorf("snapshot retention must be a number")
	}
	concurrency, err := strconv.Atoi(strings.TrimSpace(m.inputs[fieldSnapshotConcurrency].Value()))
	if err != nil {
		return config.Config{}, fmt.Errorf("snapshot concurrency must be a number")
	}

	return config.Config{
		APIKey:               strings.TrimSpace(m.inputs[fieldAPIKey].Value()),
		Model:                strings.TrimSpace(m.inputs[fieldModel].Value()),
		Endpoint:             strings.TrimSpace(m.inputs[fieldEndpoint].Value()),
		SnapshotIntervalSec:  interval,
		SnapshotRetentionSec: retention,
		SnapshotConcurrency:  concurrency,
		DBPath:               strings.TrimSpace(m.inputs[fieldDBPath].Value()),
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
