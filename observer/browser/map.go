package browser

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Process represents a categorized browser process (Chrome or Firefox)
type Process struct {
	PID       int    `json:"pid"`
	Browser   string `json:"browser"` // "chrome" or "firefox"
	Type      string `json:"type"`    // e.g. "browser", "renderer", "gpu", "utility", "network"
	URL       string `json:"url,omitempty"` // Extracted from --site-for-process (Chrome only)
	Sandboxed bool   `json:"sandboxed"`
	ParentPID int    `json:"parent_pid"`
}

// MapProcesses walks /proc and builds a map of active Chrome/Firefox processes
func MapProcesses() ([]*Process, error) {
	var procs []*Process

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		cmdline, err := readCmdline(pid)
		if err != nil {
			continue
		}

		if len(cmdline) == 0 {
			continue
		}

		// Check if Chromium/Chrome
		if isChromium(cmdline) {
			procs = append(procs, parseChrome(pid, cmdline))
			continue
		}

		// Check if Firefox
		if isFirefox(cmdline) {
			procs = append(procs, parseFirefox(pid, cmdline))
			continue
		}
	}

	return procs, nil
}

func readCmdline(pid int) ([]string, error) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return nil, err
	}
	// /proc/*/cmdline is null-delimited
	parts := bytes.Split(data, []byte{0})
	var result []string
	for _, p := range parts {
		if len(p) > 0 {
			result = append(result, string(p))
		}
	}
	return result, nil
}

func readStatPPID(pid int) int {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0
	}
	parts := strings.Fields(string(data))
	if len(parts) > 3 {
		ppid, _ := strconv.Atoi(parts[3])
		return ppid
	}
	return 0
}

func isChromium(cmdline []string) bool {
	bin := filepath.Base(cmdline[0])
	return bin == "chrome" || bin == "chromium" || bin == "Brave-Browser" || bin == "msedge"
}

func parseChrome(pid int, cmdline []string) *Process {
	p := &Process{
		PID:       pid,
		Browser:   "chrome",
		ParentPID: readStatPPID(pid),
	}

	p.Type = extractFlag(cmdline, "--type")
	if p.Type == "" {
		p.Type = "browser" // Main process
	}

	p.URL = extractFlag(cmdline, "--site-for-process")
	if p.URL == "" {
		// Fallback to app-id or other indicators if needed
		appID := extractFlag(cmdline, "--app-id")
		if appID != "" {
			p.URL = "app-id: " + appID
		}
	}

	p.Sandboxed = containsFlag(cmdline, "--sandbox-type=renderer") || extractFlag(cmdline, "--type") == "renderer"

	return p
}

func isFirefox(cmdline []string) bool {
	bin := filepath.Base(cmdline[0])
	return bin == "firefox" || bin == "firefox-bin"
}

func parseFirefox(pid int, cmdline []string) *Process {
	p := &Process{
		PID:       pid,
		Browser:   "firefox",
		ParentPID: readStatPPID(pid),
	}

	contentProc := false
	for _, arg := range cmdline {
		if arg == "-contentproc" {
			contentProc = true
			break
		}
	}

	if !contentProc {
		p.Type = "browser" // Main process
		return p
	}

	// Determine Firefox role from trailing args
	// Common trailing args: socket, rdd, web, extension
	role := "unknown"
	for i := len(cmdline) - 1; i >= 0; i-- {
		arg := cmdline[i]
		if arg == "socket" {
			role = "network_socket"
			break
		} else if arg == "web" || arg == "webIsolated" {
			role = "web_content"
			break
		} else if arg == "extension" {
			role = "extension"
			break
		} else if arg == "rdd" {
			role = "data_decoder" // RDD Process
			break
		} else if arg == "gpu" {
			role = "gpu"
			break
		}
	}
	p.Type = role
	p.Sandboxed = true

	return p
}

func extractFlag(cmdline []string, flag string) string {
	prefix := flag + "="
	for _, arg := range cmdline {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}
	return ""
}

func containsFlag(cmdline []string, flag string) bool {
	for _, arg := range cmdline {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}
