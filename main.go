package main

import (
	"bufio"
	"context"
	"fmt"
	ctxengine "godshell/context"
	"godshell/observer"
	"godshell/store"
	"log"

	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Starting godshell...")
	fmt.Println("Press 'd' + Enter for raw dump, 's' + Enter for smart snapshot, 'q' to quit")

	tree := ctxengine.NewProcessTree()
	go tree.EvictGhosts(60 * time.Second)
	go tree.RefreshMetrics(5 * time.Second)

	if err := store.Init("godshell.db"); err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	events := make(chan observer.Event, 256)
	go func() {
		if err := observer.Run(ctx, events); err != nil {
			log.Fatalf("observer: %v", err)
		}
	}()

	// Read keyboard commands in a goroutine
	keys := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			keys <- scanner.Text()
		}
	}()

	for {
		select {
		case e := <-events:
			tree.HandleEvent(e)
		case cmd := <-keys:
			switch cmd {
			case "d":
				// fmt.Print(tree.DumpDebug()) // Removed DumpDebug in favor of snapshot.
				fmt.Println("Debug dump is deprecated. Use 's' for smart snapshot.")
			case "s":
				snap := tree.TakeSnapshot()
				fmt.Print("\033[H\033[2J") // Clear terminal
				fmt.Println(snap.Summary())
				fmt.Println("=== POINT-IN-TIME SNAPSHOT MODE ===")
				fmt.Println("Commands: [i]nspect <pid>, [f]amily <pid>, [s]earch <pattern>, [q]uit snapshot")

			snapshotLoop:
				for {
					select {
					case e := <-events:
						tree.HandleEvent(e) // keep processing live events in the background
					case cmd := <-keys:
						if cmd == "" {
							continue
						}

						parts := strings.SplitN(cmd, " ", 2)
						action := parts[0]
						arg := ""
						if len(parts) > 1 {
							arg = strings.TrimSpace(parts[1])
						}

						switch action {
						case "i":
							var pid uint32
							fmt.Sscanf(arg, "%d", &pid)
							fmt.Println(snap.Inspect(pid))
						case "f":
							var pid uint32
							fmt.Sscanf(arg, "%d", &pid)
							fmt.Println(snap.ProcessFamily(pid))
						case "s":
							fmt.Println(snap.Search(arg))
						case "m":
							var pid uint32
							fmt.Sscanf(arg, "%d", &pid)
							fmt.Println(snap.ReadProcessMaps(pid))
						case "l":
							var pid uint32
							fmt.Sscanf(arg, "%d", &pid)
							fmt.Println(snap.GetLinkedLibraries(pid))
						case "t":
							var pid uint32
							fmt.Sscanf(arg, "%d", &pid)
							fmt.Println(snap.TraceSyscalls(pid, 5))
						case "c":
							parts := strings.Fields(arg)
							if len(parts) >= 1 {
								path := parts[0]
								var offset, limit int64 = 0, 4096
								if len(parts) >= 2 {
									fmt.Sscanf(parts[1], "%d", &offset)
								}
								if len(parts) >= 3 {
									fmt.Sscanf(parts[2], "%d", &limit)
								}
								fmt.Println(snap.ReadFile(path, offset, limit))
							} else {
								fmt.Println("Usage: c <path> [offset] [limit]")
							}
						case "r":
							parts := strings.Fields(arg)
							if len(parts) >= 2 {
								var pid uint32
								var address uint64
								var size int64 = 1024
								fmt.Sscanf(parts[0], "%d", &pid)

								// Parse hex address robustly handling optional 0x prefix
								addrStr := strings.TrimPrefix(parts[1], "0x")
								fmt.Sscanf(addrStr, "%x", &address)

								if len(parts) >= 3 {
									fmt.Sscanf(parts[2], "%d", &size)
								}
								fmt.Println(snap.ReadMemory(pid, address, size))
							} else {
								fmt.Println("Usage: r <pid> <address_hex> [size]")
							}
						case "save":
							id, err := store.SaveSnapshot(arg, snap)
							if err != nil {
								fmt.Printf("Save error: %v\n", err)
							} else {
								fmt.Printf("Snapshot saved as ID %d (label: %s)\n", id, arg)
							}
						case "load":
							var id int64
							fmt.Sscanf(arg, "%d", &id)
							newSnap, err := store.LoadSnapshot(id)
							if err != nil {
								fmt.Printf("Load error: %v\n", err)
							} else {
								snap = newSnap
								fmt.Printf("Switched to loaded snapshot %d (%s)\n", id, snap.Timestamp.Format(time.RFC3339))
							}
						case "list":
							list, err := store.ListSnapshots()
							if err != nil {
								fmt.Printf("List error: %v\n", err)
							} else {
								fmt.Println("Stored Snapshots:")
								for _, m := range list {
									fmt.Printf("  [%d] %s (%v)\n", m.ID, m.Label, m.Timestamp.Format(time.Kitchen))
								}
							}
						case "q":
							fmt.Println("Exited snapshot mode.")
							break snapshotLoop

						default:
							fmt.Println("Unknown command inside snapshot mode.")
							fmt.Println("Commands: [i]nspect, [f]amily, [s]earch, [m]aps, [l]ibraries, [t]race, [c]at file, [r]ead mem, [q]uit")
						}
					case <-ctx.Done():
						return
					}
				}

			case "q":
				return
			default:
				fmt.Println("[d]ump  [s]napshot  [quit]")
			}
		case <-ctx.Done():
			return
		}
	}
}
