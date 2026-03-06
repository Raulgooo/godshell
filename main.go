package main

import (
	"context"
	"fmt"
	"godshell/observer"
	"log"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	events := make(chan observer.Event, 256)

	go func() {
		if err := observer.Run(ctx, events); err != nil {
			log.Fatalf("observer: %v", err)
		}
	}()

	for {
		select {
		case e := <-events:
			path := e.PathStr()
			// Drop high-volume noise. Real filtering belongs in BPF C code; this is the Go-side quick filter.
			if strings.HasPrefix(path, "/proc/") ||
				strings.HasPrefix(path, "/sys/") ||
				strings.HasPrefix(path, "/dev/") ||
				strings.HasPrefix(path, "/usr/share/locale/") ||
				strings.HasPrefix(path, "/usr/lib/locale/") ||
				strings.HasPrefix(path, "/usr/lib/gconv/") ||
				strings.HasSuffix(path, ".so") ||
				strings.Contains(path, ".so.") ||
				// Relative proc pseudo-files opened by ps/procfs readers via dirfd
				path == "stat" || path == "status" || path == "environ" ||
				path == "cmdline" || path == "maps" || path == "fd" ||
				path == "" {
				continue
			}
			fmt.Printf("[%s] pid=%-6d uid=%-4d %-16s %s\n",
				e.KindStr(), e.Pid, e.Uid, e.CommStr(), path)
		case <-ctx.Done():
			return
		}
	}
}
