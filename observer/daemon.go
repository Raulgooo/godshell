package observer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

// Run loads the BPF programs, attaches the four tracepoints, and streams
// kernel events to out. It blocks until ctx is cancelled or a fatal error
// occurs. Call from main with sudo / CAP_BPF.
func Run(ctx context.Context, out chan<- Event) error {
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("remove memlock rlimit: %w", err)
	}

	// Load the CollectionSpec so we can inject the self-PID global
	// before instantiating the programs in the kernel.
	spec, err := loadBpf()
	if err != nil {
		return fmt.Errorf("load bpf spec: %w", err)
	}

	// Write godshell's own PID into the BPF global variable so all
	// handlers can filter self-events at the kernel level.
	selfPID := uint32(os.Getpid())
	if v, ok := spec.Variables["godshell_pid"]; ok {
		if err := v.Set(selfPID); err != nil {
			return fmt.Errorf("set godshell_pid: %w", err)
		}
	} else {
		return fmt.Errorf("BPF variable 'godshell_pid' not found in spec")
	}

	var objs bpfObjects
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{}); err != nil {
		return fmt.Errorf("load bpf objects: %w", err)
	}
	defer objs.Close()

	// Attach each BPF program to its tracepoint.
	execveLink, err := link.Tracepoint("syscalls", "sys_enter_execve", objs.TraceExecve, nil)
	if err != nil {
		return fmt.Errorf("attach sys_enter_execve: %w", err)
	}
	defer execveLink.Close()

	openLink, err := link.Tracepoint("syscalls", "sys_enter_openat", objs.TraceOpenat, nil)
	if err != nil {
		return fmt.Errorf("attach sys_enter_openat: %w", err)
	}
	defer openLink.Close()

	exitLink, err := link.Tracepoint("sched", "sched_process_exit", objs.TraceExit, nil)
	if err != nil {
		return fmt.Errorf("attach sched_process_exit: %w", err)
	}
	defer exitLink.Close()

	connectLink, err := link.Tracepoint("syscalls", "sys_enter_connect", objs.TraceConnect, nil)
	if err != nil {
		return fmt.Errorf("attach sys_enter_connect: %w", err)
	}
	defer connectLink.Close()

	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("open ring buffer reader: %w", err)
	}
	defer rd.Close()

	go func() {
		<-ctx.Done()
		rd.Close()
	}()

	log.Printf("observers attached: execve openat exit connect (self PID %d, kernel-filtered)", selfPID)

	for {
		record, err := rd.Read()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("ring buffer read: %w", err)
		}

		var e Event
		if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &e); err != nil {
			log.Printf("warn: decode event: %v", err)
			continue
		}

		select {
		case out <- e:
		case <-ctx.Done():
			return nil
		}
	}
}
