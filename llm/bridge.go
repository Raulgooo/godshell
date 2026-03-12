package llm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"godshell/client"
	ctxengine "godshell/context"
	"godshell/intel"
)

type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	// For tool integration
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`       // For tool responses
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"` // For tool calls from assistant
}

type Conversation struct {
	History         []Message
	CurrentSnapshot *ctxengine.FrozenSnapshot
	DaemonClient    *client.DaemonClient // optional: if set, proxy tool calls here
	Intel           *intel.Client        // optional: threat intel client
}

func NewConversation(snap *ctxengine.FrozenSnapshot, daemonClient *client.DaemonClient) *Conversation {
	c := &Conversation{
		History:         []Message{},
		CurrentSnapshot: snap,
		DaemonClient:    daemonClient,
	}

	// Master system prompt to establish investigatory identity
	c.History = append(c.History, Message{
		Role: RoleSystem,
		Content: `You are Godshell, an elite AI Investigatory Agent operating on immutable system snapshots.
You handle process analysis, network forensics, reverse engineering, debugging, and cybersecurity investigations.

CORE DIRECTIVES
1. Never fabricate data. If it is not in the snapshot, it does not exist.
2. The snapshot is your sole source of truth.
3. When uncertain, say so.
4. Cross-reference PIDs, ports, paths, and parent/child trees to build inferences.
5. Always cite which tool or field led you to each conclusion.
6. Be concise. No walls of text.
7. Attack agressively, but wise, never judge or assume eagerly.
8. 

OUTPUT RULES
- No markdown. No headers, no bold, no bullet symbols, no code fences.
- Lead with the finding, then the evidence. Never the other way around.
- Be terse. If the answer is two lines, write two lines.
- Use plain numbered lists only when sequence or enumeration genuinely matters.
- Never pad, summarize what was already said, or add closing remarks.
- If you have a theory, state it clearly and back it up with evidence.
- If you want the user to give you info, or write a command, please prompt the user shortly.
INVESTIGATION METHODOLOGY
- Triage first, drill second.
- Chase the lineage. Suspicious behavior rarely starts where it appears.
- Correlate everything. A process touching a network socket and an unusual path is more damning than either alone.
- Ghosts matter. A process that exited is not innocent.
- Absence is evidence. If something expected is missing, flag it.
`})

	return c
}

// GetToolDefinitions returns the JSON schema for the tools available to the LLM.
func (c *Conversation) GetToolDefinitions() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "summary",
				Description: "Get a high-level summary of all active applications, process groups, and recently exited processes.",
				Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "inspect",
				Description: "Get deep metadata for a specific PID, including full command line, parent info, and collapsed file/network effects.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to inspect"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "search",
				Description: "Search for processes matching a regex pattern against names, paths, or connection/file targets.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {"type": "string", "description": "Regex pattern to search for"}
					},
					"required": ["query"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "family",
				Description: "Get the process lineage (ancestors and descendants) for a specific PID.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to trace"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "get_maps",
				Description: "Get a condensed summary of a process's memory layout (heap, stack, libraries) from /proc/pid/maps.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to inspect"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "get_libraries",
				Description: "Resolve the dynamically linked shared objects (ldd) for a process's binary.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to inspect"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "trace",
				Description: "Run a scoped 5-second strace against a PID to see current system call activity.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to trace"},
						"duration": {"type": "integer", "description": "Tracing duration in seconds (default 5)"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "read_file",
				Description: "Read a chunk of a file from disk. Helpful for config files or sensitive paths identified in snapshots.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"path": {"type": "string", "description": "Absolute path to the file"},
						"offset": {"type": "integer", "description": "Byte offset to start reading from"},
						"limit": {"type": "integer", "description": "Maximum bytes to read (default 4096)"}
					},
					"required": ["path"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "read_memory",
				Description: "Read raw process memory directly from /proc/pid/mem at a specific address.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to read from"},
						"address_hex": {"type": "string", "description": "Memory address in hex exactly as shown in get_maps output. Example: '55bd07d14000'. Do NOT convert to decimal. Do NOT add 0x prefix."},
						"size": {"type": "integer", "description": "Number of bytes to read. Default 1024, max 65536."}
					},
					"required": ["pid", "address_hex"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "gohash_binary",
				Description: "Computes the SHA-256 hash of the executable associated with a process for reputation checks or integrity analysis.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to hash its binary"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "goread_shell_history",
				Description: "Retrieves the last N lines of a user's shell history (.bash_history, .zsh_history). Critical to understand what commands a user (or attacker) typed.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"user": {"type": "string", "description": "The username whose history you want to read"},
						"limit": {"type": "integer", "description": "The number of recent lines to read (default 50)"}
					},
					"required": ["user"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "gonetwork_state",
				Description: "Extracts active network connections and their state (ESTABLISHED, LISTEN, CLOSE_WAIT) directly from /proc/pid/net/tcp and tcp6 for C2 and backdoor detection.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to inspect network state for"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "goread_environ",
				Description: "Parses and returns the environment variables for a process from /proc/pid/environ to hunt for hardcoded credentials or tokens.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pid": {"type": "integer", "description": "The process ID to inspect its environment"}
					},
					"required": ["pid"]
				}`),
			},
		},
		{
			Type: "function",
			Function: Function{
				Name:        "goextract_strings",
				Description: "Extracts ASCII and Unicode strings from a binary file to find embedded URLs, keys, or messages.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"path": {"type": "string", "description": "The absolute path to the file to run strings on"},
						"min_length": {"type": "integer", "description": "The minimum length of strings to extract (default 8)"}
					},
					"required": ["path"]
				}`),
			},
		},
		// Disabled tools for this version:
		// browser_map
		// ssl_intercept
		// virustotal_hash
		// abuseipdb_ip
		// all report_* cards
	}
}

// UpdateSnapshot replaces the current snapshot with a fresh one.
func (c *Conversation) UpdateSnapshot(newSnap *ctxengine.FrozenSnapshot) {
	c.CurrentSnapshot = newSnap
	c.History = append(c.History, Message{
		Role: RoleSystem,
		Content: fmt.Sprintf("CONTEXT REFRESH: A new system snapshot was taken at %s. "+
			"Future tool calls will now reflect this updated state.", newSnap.Timestamp.Format(time.RFC3339)),
	})
}

// ExecuteTool dispatches LLM tool calls to the appropriate Snapshot JSON methods.
func (c *Conversation) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
	// If the tool is disabled, block it here.
	switch toolName {
	case "browser_map", "ssl_intercept", "virustotal_hash", "abuseipdb_ip",
		"report_text", "report_behaviour", "report_family", "report_network", "report_threat", "report_system_state":
		return "", fmt.Errorf("tool '%s' is not available in this version", toolName)
	}

	// Intel tools don't need a snapshot or daemon; handle them here first.
	// (Hidden but kept for easier reactivation)
	/*
		case "virustotal_hash":
			sha256, _ := args["sha256"].(string)
			if sha256 == "" { return "", fmt.Errorf("missing sha256") }
			if c.Intel == nil { return "VirusTotal: no API key", nil }
			r := c.Intel.LookupHash(sha256)
			return fmt.Sprintf("VirusTotal summary: %s", r.Summary), nil
		case "abuseipdb_ip":
			ip, _ := args["ip"].(string)
			if ip == "" { return "", fmt.Errorf("missing ip") }
			if c.Intel == nil { return "AbuseIPDB: no API key", nil }
			r := c.Intel.LookupIP(ip)
			return fmt.Sprintf("AbuseIPDB score: %d/100", r.AbuseScore), nil
	*/

	// If we are functioning as a CLI connected to a daemon, proxy privileged requests.
	if c.DaemonClient != nil && IsPrivilegedTool(toolName) {
		return c.DaemonClient.ExecuteTool(toolName, args)
	}

	if c.CurrentSnapshot == nil {
		return "", fmt.Errorf("no snapshot loaded")
	}

	return ExecuteToolOnSnapshot(toolName, args, c.CurrentSnapshot)
}

// ExecuteToolOnSnapshot statically evaluates a tool against a given snapshot.
func ExecuteToolOnSnapshot(toolName string, args map[string]interface{}, snap *ctxengine.FrozenSnapshot) (string, error) {
	switch toolName {
	case "summary":
		return snap.SummaryJSON()
	case "inspect":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return snap.InspectJSON(uint32(pid))
	case "search":
		query, ok := args["query"].(string)
		if !ok {
			return "", fmt.Errorf("missing query argument")
		}
		return snap.SearchJSON(query)
	case "family":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return snap.ProcessFamilyJSON(uint32(pid))
	case "get_maps":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		return snap.ReadProcessMaps(uint32(pid)), nil
	case "get_libraries":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		return snap.GetLinkedLibraries(uint32(pid)), nil
	case "trace":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		dur := 5.0
		if d, ok := args["duration"].(float64); ok {
			dur = d
		}
		return snap.TraceSyscalls(uint32(pid), int(dur)), nil
	case "read_file":
		path, _ := args["path"].(string)
		off := 0.0
		if o, ok := args["offset"].(float64); ok {
			off = o
		}
		lim := 4096.0
		if l, ok := args["limit"].(float64); ok {
			lim = l
		}
		return snap.ReadFile(path, int64(off), int64(lim)), nil
	case "read_memory":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		addrStr, ok := args["address_hex"].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid address_hex argument")
		}
		addrStr = strings.TrimPrefix(strings.TrimSpace(addrStr), "0x")
		addr, err := strconv.ParseUint(addrStr, 16, 64)
		if err != nil {
			return "", fmt.Errorf("invalid address_hex format '%s': %v", addrStr, err)
		}
		size := int64(1024)
		if s, ok := args["size"].(float64); ok {
			size = int64(s)
		}
		return snap.ReadMemory(uint32(pid), addr, size), nil
	case "gohash_binary":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return snap.HashBinary(uint32(pid)), nil
	case "goread_shell_history":
		user, ok := args["user"].(string)
		if !ok {
			return "", fmt.Errorf("missing user argument")
		}
		limit := 50.0
		if l, ok := args["limit"].(float64); ok {
			limit = l
		}
		return snap.ReadShellHistory(user, int(limit)), nil
	case "gonetwork_state":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return snap.NetworkState(uint32(pid)), nil
	case "goread_environ":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return snap.ReadEnviron(uint32(pid)), nil
	case "goextract_strings":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("missing path argument")
		}
		minLength := 8.0
		if m, ok := args["min_length"].(float64); ok {
			minLength = m
		}
		return snap.ExtractStrings(path, int(minLength)), nil
	default:
		return "", fmt.Errorf("unknown or disabled tool: %s", toolName)
	}
}

// IsPrivilegedTool returns true if the tool needs to run as root/daemon
func IsPrivilegedTool(name string) bool {
	switch name {
	case "summary", "inspect", "search", "family", "browser_map",
		"report_text", "report_behaviour", "report_family", "report_network", "report_threat", "report_system_state":
		return false
	case "get_maps", "get_libraries", "trace", "read_file", "read_memory",
		"gohash_binary", "goread_shell_history", "gonetwork_state",
		"goread_environ", "goextract_strings", "ssl_intercept":
		return true
	default:
		return false
	}
}
