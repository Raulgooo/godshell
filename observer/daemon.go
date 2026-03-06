package observer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
)

// Run loads the BPF programs, attaches the four tracepoints, and streams
// kernel events to out. It blocks until ctx is cancelled or a fatal error
// occurs. Call from main with sudo / CAP_BPF.
func Run(ctx context.Context, out chan<- Event) error {
	// Kernels older than 5.11 require the memlock rlimit to be removed
	// before BPF maps can be allocated. No-op on newer kernels.
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("remove memlock rlimit: %w", err)
	}

	// Load the pre-compiled BPF bytecode (embedded in the binary by bpf2go)
	// and instantiate all programs and maps in the kernel.
	var objs bpfObjects
	if err := loadBpfObjects(&objs, nil); err != nil {
		return fmt.Errorf("load bpf objects: %w", err)
	}
	defer objs.Close()

	// Attach each BPF program to its tracepoint.
	// The link must be kept alive for the duration of the run; closing it
	// detaches the tracepoint automatically.
	execLink, err := link.Tracepoint("sched", "sched_process_exec", objs.TraceExec, nil)
	if err != nil {
		return fmt.Errorf("attach sched_process_exec: %w", err)
	}
	defer execLink.Close()

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

	// Open a reader on the ring buffer map. Events written by the BPF
	// handlers will be available here.
	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("open ring buffer reader: %w", err)
	}
	defer rd.Close()

	// Unblock rd.Read() when the context is cancelled so the loop can exit.
	go func() {
		<-ctx.Done()
		rd.Close()
	}()

	log.Println("observers attached: exec openat exit connect")

	for {
		record, err := rd.Read()
		if err != nil {
			if ctx.Err() != nil {
				return nil // clean shutdown, not an error
			}
			return fmt.Errorf("ring buffer read: %w", err)
		}

		// Deserialise the raw bytes into an Event. binary.Read consumes
		// exactly sizeof(Event) bytes; any tail padding from the C struct
		// is silently ignored.
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
