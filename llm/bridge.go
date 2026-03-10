package llm

import (
	"encoding/json"
	"fmt"
	ctxengine "godshell/context"
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
						"address": {"type": "integer", "description": "The memory address (in decimal/integer form)"},
						"size": {"type": "integer", "description": "Number of bytes to read (default 1024)"}
					},
					"required": ["pid", "address"]
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
		addrVal := args["address"]
		addr, _ := addrVal.(float64)
		size := 1024.0
		if s, ok := args["size"].(float64); ok {
			size = s
		}
		return c.CurrentSnapshot.ReadMemory(uint32(pid), uint64(addr), int64(size)), nil
	default:

		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}
