package ctxengine

import (
	"encoding/json"
	"sync"
	"time"
)

// This is not a tree, its a flat index used as a graph, a real tree would be taxing in performance.
type EffectKind uint8

const (
	EffectOpen    EffectKind = 0
	EffectConnect EffectKind = 1
)

// Effect structure. every process has a map of effects
type Effect struct {
	Kind   EffectKind
	Target string // Path
	Count  uint64 // Number of times the effect has taken place
	First  time.Time
	Last   time.Time
	Detail EffectDetail //Metadata for the effect.
}

type EffectDetail interface {
	effectDetail()
}

type rawEffect struct {
	Kind   EffectKind      `json:"Kind"`
	Target string          `json:"Target"`
	Count  uint64          `json:"Count"`
	First  time.Time       `json:"First"`
	Last   time.Time       `json:"Last"`
	Detail json.RawMessage `json:"Detail"`
}

func (e *Effect) UnmarshalJSON(data []byte) error {
	var raw rawEffect
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.Kind = raw.Kind
	e.Target = raw.Target
	e.Count = raw.Count
	e.First = raw.First
	e.Last = raw.Last

	if len(raw.Detail) > 0 && string(raw.Detail) != "null" {
		switch e.Kind {
		case EffectOpen:
			var detail OpenDetail
			if err := json.Unmarshal(raw.Detail, &detail); err != nil {
				return err
			}
			e.Detail = detail
		case EffectConnect:
			var detail ConnectDetail
			if err := json.Unmarshal(raw.Detail, &detail); err != nil {
				return err
			}
			e.Detail = detail
		}
	}

	return nil
}

type OpenDetail struct {
	Flags int // 0 = READ ONLY, 1 = WRITE ONLY, 2 = READ WRITE
}

func (OpenDetail) effectDetail() {}

type ConnectDetail struct {
	IP         string
	Port       uint16
	Domain     string // reverse DNS, cached
	BytesSent  uint64 // tx_queue from /proc/net/tcp
	BytesRecv  uint64 // rx_queue from /proc/net/tcp
	Family     uint16 // AF_INET=2, AF_INET6=10, AF_UNIX=1
	UnixSocket string // path for unix domain sockets
}

func (ConnectDetail) effectDetail() {}

// Definition for a process.
type ProcessNode struct {
	PID            uint32
	PPID           uint32
	BinaryPath     string
	Comm           string
	Effects        map[string]*Effect
	StartTimestamp time.Time
	EndTimestamp   time.Time
	ChildrenPID    []uint32
	CpuUsage       float64
	MemoryUsage    uint64
	IsEnriched     bool
}

// Clone deeply copies an Effect.
func (e *Effect) Clone() *Effect {
	if e == nil {
		return nil
	}
	clone := *e
	// EffectDetail is an interface, but the concrete types (OpenDetail, ConnectDetail) are structs containing only value types (no pointers/slices/maps).
	// So a shallow copy of the interface value is sufficient for our current implementations.
	return &clone
}

// Clone deeply copies a ProcessNode.
func (n *ProcessNode) Clone() *ProcessNode {
	if n == nil {
		return nil
	}

	clone := *n

	if n.Effects != nil {
		clone.Effects = make(map[string]*Effect, len(n.Effects))
		for k, v := range n.Effects {
			clone.Effects[k] = v.Clone()
		}
	}

	if n.ChildrenPID != nil {
		clone.ChildrenPID = make([]uint32, len(n.ChildrenPID))
		copy(clone.ChildrenPID, n.ChildrenPID)
	}

	return &clone
}

type ProcessTree struct {
	mu     sync.RWMutex
	ByPID  map[uint32]*ProcessNode
	Ghosts map[uint32]*ProcessNode

	// Dynamic Configuration
	MaxEffectsPerProcess int
	CaptureNetwork       bool
	CaptureFileIO        bool
	IgnoredProcesses     map[string]struct{}
	ProcPath             string
	SysPath              string
	StracePath           string
}
