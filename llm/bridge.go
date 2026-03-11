package llm

import (
	"encoding/json"
	"fmt"
	ctxengine "godshell/context"
	"strconv"
	"strings"
	"time"
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
}

func NewConversation(snap *ctxengine.FrozenSnapshot) *Conversation {
	c := &Conversation{
		History:         []Message{},
		CurrentSnapshot: snap,
	}

	// Master system prompt to establish investigatory identity
	c.History = append(c.History, Message{
		Role: RoleSystem,
		Content: "You are the Godshell AI Investigatory Agent. You operate on immutable system snapshots. " +
			"Use the provided tools to explore process behavior, network connections, and file system activity. " +
			"Identify anomalies, suspicious behaviors, or performance bottlenecks. " +
			"The user can ask you for reverse engineering tasks, debugging tasks, or general cybersec and analysis tasks. " +
			"these are you DIRECTIVES: " +
			"You must NEVER lie or generate fake data. " +
			"The snapshot is your source of truth. " +
			"If you are not sure, be honest about it. " +
			"	CrossReference data to make inferences. " +
			"In your answers you must mention where you digged, and ground your claims in what you saw.",
	},
	)

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
		{
			Type: "function",
			Function: Function{
				Name:        "browser_map",
				Description: "Gets an abstract tree of all running Chrome and Firefox processes. Maps PIDs to URLs (for Chrome tabs) and network roles (for Firefox) to cut through process noise and identify targets for interception.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {}
				}`),
			},
		},
	}
}

// UpdateSnapshot replaces the current snapshot with a fresh one.
// This is useful for longitudinal analysis where the LLM compares context before and after an event.
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
	if c.CurrentSnapshot == nil {
		return "", fmt.Errorf("no snapshot loaded")
	}

	switch toolName {
	case "summary":
		return c.CurrentSnapshot.SummaryJSON()
	case "inspect":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		// JSON numbers unmarshal to float64 by default
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.InspectJSON(uint32(pid))
	case "search":
		query, ok := args["query"].(string)
		if !ok {
			return "", fmt.Errorf("missing query argument")
		}
		return c.CurrentSnapshot.SearchJSON(query)
	case "family":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.ProcessFamilyJSON(uint32(pid))
	case "get_maps":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.ReadProcessMaps(uint32(pid)), nil
	case "get_libraries":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.GetLinkedLibraries(uint32(pid)), nil
	case "trace":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		dur := 5.0
		if d, ok := args["duration"].(float64); ok {
			dur = d
		}
		return c.CurrentSnapshot.TraceSyscalls(uint32(pid), int(dur)), nil
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
		return c.CurrentSnapshot.ReadFile(path, int64(off), int64(lim)), nil
	case "read_memory":
		pidVal := args["pid"]
		pid, _ := pidVal.(float64)
		addrStr, ok := args["address_hex"].(string)
		if !ok {
			return "", fmt.Errorf("missing or invalid address_hex argument")
		}
		// Sanitize input from LLM just in case it adds 0x or whitespace
		addrStr = strings.TrimPrefix(strings.TrimSpace(addrStr), "0x")
		addr, err := strconv.ParseUint(addrStr, 16, 64)
		if err != nil {
			return "", fmt.Errorf("invalid address_hex format '%s': %v", addrStr, err)
		}
		size := int64(1024)
		if s, ok := args["size"].(float64); ok {
			size = int64(s)
		}
		return c.CurrentSnapshot.ReadMemory(uint32(pid), addr, size), nil
	case "gohash_binary":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.HashBinary(uint32(pid)), nil
	case "goread_shell_history":
		user, ok := args["user"].(string)
		if !ok {
			return "", fmt.Errorf("missing user argument")
		}
		limit := 50.0
		if l, ok := args["limit"].(float64); ok {
			limit = l
		}
		return c.CurrentSnapshot.ReadShellHistory(user, int(limit)), nil
	case "gonetwork_state":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.NetworkState(uint32(pid)), nil
	case "goread_environ":
		pidVal, ok := args["pid"]
		if !ok {
			return "", fmt.Errorf("missing pid argument")
		}
		pid, _ := pidVal.(float64)
		return c.CurrentSnapshot.ReadEnviron(uint32(pid)), nil
	case "goextract_strings":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("missing path argument")
		}
		minLength := 8.0
		if m, ok := args["min_length"].(float64); ok {
			minLength = m
		}
		return c.CurrentSnapshot.ExtractStrings(path, int(minLength)), nil
	case "browser_map":
		// Directly maps state from the OS using pure Go; doesnt need a context snapshot directly.
		// However, returning as JSON allows the LLM to inspect it easily.
		// We'll call the browser package. Wait, bridge shouldn't import it directly if it's meant to be contextualized.
		// As a quick tool, we can expose it via Snapshot or directly.
		return c.CurrentSnapshot.BrowserMapJSON()
	default:

		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}
