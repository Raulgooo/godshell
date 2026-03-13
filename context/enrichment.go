package ctxengine

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// ── DNS cache ──────────────────────────────────────────────────────────────

var dnsCache sync.Map // map[string]string → ip → domain

// reverseDNS performs a cached reverse DNS lookup.
// Returns the domain or empty string if lookup fails.
func reverseDNS(ip string) string {
	if v, ok := dnsCache.Load(ip); ok {
		return v.(string)
	}
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		dnsCache.Store(ip, "")
		return ""
	}
	domain := strings.TrimSuffix(names[0], ".")
	dnsCache.Store(ip, domain)
	return domain
}

// ── /proc/net/tcp parser ───────────────────────────────────────────────────

type tcpEntry struct {
	LocalIP    string
	LocalPort  uint16
	RemoteIP   string
	RemotePort uint16
	State      uint8 // 1=ESTABLISHED, 2=SYN_SENT, etc.
	TxQueue    uint64
	RxQueue    uint64
	Family     uint16 // 2=AF_INET, 10=AF_INET6
}

// enrichConnect reads <proc>/<pid>/net/tcp and tcp6 to find the newest
// connection. Returns nil if the process is gone or has no connections.
func enrichConnect(procPath string, pid uint32) *ConnectDetail {
	entries := parseProcNetTCP(procPath, pid, false)                   // IPv4
	entries = append(entries, parseProcNetTCP(procPath, pid, true)...) // IPv6

	if len(entries) > 0 {
		// Pick the best entry: prefer ESTABLISHED (1) or SYN_SENT (2)
		var best *tcpEntry
		for i := range entries {
			e := &entries[i]
			if e.State == 1 || e.State == 2 {
				if best == nil || e.State < best.State {
					best = e
				}
			}
		}
		if best == nil {
			best = &entries[len(entries)-1] // Fall back to last entry
		}

		domain := reverseDNS(best.RemoteIP)

		return &ConnectDetail{
			IP:        best.RemoteIP,
			Port:      best.RemotePort,
			Domain:    domain,
			BytesSent: best.TxQueue,
			BytesRecv: best.RxQueue,
			Family:    best.Family,
		}
	}

	// Fallback to <proc>/<pid>/net/unix for local IPC
	path := parseProcNetUnix(procPath, pid)
	if path != "" {
		return &ConnectDetail{
			Family:     1, // AF_UNIX
			UnixSocket: path,
		}
	}

	return nil
}

// parseProcNetUnix reads <proc>/<pid>/net/unix and returns the path
// of the last connected unix socket (skips empty paths and anonymous sockets).
func parseProcNetUnix(procPath string, pid uint32) string {
	path := fmt.Sprintf("%s/%d/net/unix", procPath, pid)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	var lastPath string

	// Skip header line
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// Format: Num RefCount Protocol Flags Type St Inode Path
		// Path is the 8th field (index 7), but it can be empty
		if len(fields) >= 8 {
			p := fields[7]
			// Skip anonymous sockets (start with @ in net/unix)
			if !strings.HasPrefix(p, "@") {
				lastPath = p
			}
		}
	}
	return lastPath
}

// parseProcNetTCP reads <proc>/<pid>/net/tcp or tcp6 and returns parsed entries.
func parseProcNetTCP(procPath string, pid uint32, ipv6 bool) []tcpEntry {
	path := fmt.Sprintf("%s/%d/net/tcp", procPath, pid)
	family := uint16(2) // AF_INET
	if ipv6 {
		path = fmt.Sprintf("%s/%d/net/tcp6", procPath, pid)
		family = 10 // AF_INET6
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	var entries []tcpEntry

	// Skip header line (line 0)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		localIP, localPort := parseHexAddr(fields[1], ipv6)
		remoteIP, remotePort := parseHexAddr(fields[2], ipv6)

		var state uint8
		fmt.Sscanf(fields[3], "%X", &state)

		var txQueue, rxQueue uint64
		if len(fields) > 4 {
			queues := strings.Split(fields[4], ":")
			if len(queues) == 2 {
				fmt.Sscanf(queues[0], "%X", &txQueue)
				fmt.Sscanf(queues[1], "%X", &rxQueue)
			}
		}

		// Skip LISTEN sockets and loopback
		if state == 10 { // LISTEN
			continue
		}
		if remoteIP == "127.0.0.1" || remoteIP == "::1" || remoteIP == "0.0.0.0" || remoteIP == "::" {
			continue
		}

		entries = append(entries, tcpEntry{
			LocalIP:    localIP,
			LocalPort:  localPort,
			RemoteIP:   remoteIP,
			RemotePort: remotePort,
			State:      state,
			TxQueue:    txQueue,
			RxQueue:    rxQueue,
			Family:     family,
		})
	}

	return entries
}

// parseHexAddr parses "0100007F:0050" format from /proc/net/tcp.
// IPv4 addresses are stored as little-endian 32-bit hex.
// IPv6 addresses are stored as 4 groups of little-endian 32-bit hex.
func parseHexAddr(s string, ipv6 bool) (string, uint16) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", 0
	}

	var port uint16
	fmt.Sscanf(parts[1], "%X", &port)

	if ipv6 {
		return parseHexIPv6(parts[0]), port
	}
	return parseHexIPv4(parts[0]), port
}

// parseHexIPv4 converts "0100007F" → "127.0.0.1" (little-endian).
func parseHexIPv4(s string) string {
	if len(s) != 8 {
		return s
	}
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 4 {
		return s
	}
	// Little-endian: reverse byte order
	return fmt.Sprintf("%d.%d.%d.%d", b[3], b[2], b[1], b[0])
}

// parseHexIPv6 converts the 32-char hex string to an IPv6 address.
func parseHexIPv6(s string) string {
	if len(s) != 32 {
		return s
	}
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 16 {
		return s
	}
	// Each 4-byte group is little-endian
	for i := 0; i < 16; i += 4 {
		b[i], b[i+3] = b[i+3], b[i]
		b[i+1], b[i+2] = b[i+2], b[i+1]
	}
	ip := net.IP(b)
	return ip.String()
}
