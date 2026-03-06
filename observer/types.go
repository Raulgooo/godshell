package observer

import "strings"

// EventKind identifies the type of kernel event.
type EventKind uint8

const (
	EventExec    EventKind = 0
	EventOpen    EventKind = 1
	EventExit    EventKind = 2
	EventConnect EventKind = 3
)

// Event is the Go mirror of the C `struct event` in observer.bpf.c.
// Field order and types must match the C struct exactly so binary.Read
// deserialises the ring buffer bytes correctly.
type Event struct {
	Ts   uint64
	Pid  uint32
	Uid  uint32
	Type EventKind
	Comm [16]byte
	Path [256]byte
}

// CommStr returns the null-terminated process name as a Go string.
func (e *Event) CommStr() string {
	return strings.TrimRight(string(e.Comm[:]), "\x00")
}

// PathStr returns the null-terminated path as a Go string.
func (e *Event) PathStr() string {
	return strings.TrimRight(string(e.Path[:]), "\x00")
}

// KindStr returns a human-readable event type label.
func (e *Event) KindStr() string {
	switch e.Type {
	case EventExec:
		return "exec"
	case EventOpen:
		return "open"
	case EventExit:
		return "exit"
	case EventConnect:
		return "connect"
	default:
		return "unknown"
	}
}
