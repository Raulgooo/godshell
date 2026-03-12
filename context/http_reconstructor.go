package ctxengine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	sslobs "godshell/observer/ssl"
)

// ── Types ───────────────────────────────────────────────────────────────────

// EndpointStats tracks how many times an API endpoint was called and
// whether it used authentication headers.
type EndpointStats struct {
	Method       string
	PathTemplate string // normalized: /users/{id}/posts
	Count        int
	AuthSeen     bool
	AuthType     string // "Bearer", "Basic", "ApiKey", etc.
	LastSeen     time.Time
	StatusCodes  map[int]int // status → count
}

// APIMap aggregates reconstructed API calls per process.
// Key: "METHOD /path/template"
type APIMap struct {
	mu        sync.Mutex
	Endpoints map[string]*EndpointStats
	PID       uint32
	Comm      string
	Duration  time.Duration
	CaptureAt time.Time
}

func newAPIMap(pid uint32, comm string) *APIMap {
	return &APIMap{
		Endpoints: make(map[string]*EndpointStats),
		PID:       pid,
		Comm:      comm,
		CaptureAt: time.Now(),
	}
}

// ── Path normalization ───────────────────────────────────────────────────────

var (
	reUUID    = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	reNumeric = regexp.MustCompile(`/\d+(/|$)`)
	reHex     = regexp.MustCompile(`/[0-9a-f]{16,}`)
)

// normalizePath replaces dynamic segments with placeholders.
// e.g.: /users/12345/profile → /users/{id}/profile
//
//	/api/v2/messages/a3b4c5d6e7f8a3b4c5d6e7f8 → /api/v2/messages/{id}
func normalizePath(p string) string {
	p = reUUID.ReplaceAllString(p, "{uuid}")
	p = reNumeric.ReplaceAllStringFunc(p, func(s string) string {
		if strings.HasSuffix(s, "/") {
			return "/{id}/"
		}
		return "/{id}"
	})
	p = reHex.ReplaceAllString(p, "/{id}")
	// Strip query string for template purposes
	if idx := strings.Index(p, "?"); idx >= 0 {
		p = p[:idx]
	}
	return p
}

// detectAuth extracts the auth type from request headers.
func detectAuth(req *http.Request) (bool, string) {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		// Check API key headers
		for _, h := range []string{"X-Api-Key", "X-API-Key", "Api-Key", "X-Auth-Token"} {
			if req.Header.Get(h) != "" {
				return true, "ApiKey"
			}
		}
		return false, ""
	}
	if strings.HasPrefix(auth, "Bearer ") {
		return true, "Bearer"
	}
	if strings.HasPrefix(auth, "Basic ") {
		return true, "Basic"
	}
	return true, "Other"
}

// ── Per-PID stream buffer ────────────────────────────────────────────────────

// streamBuffer accumulates raw SSL bytes per PID until parseable.
type streamBuffer struct {
	readBuf  []byte
	writeBuf []byte
}

// ── SSLIntercept: the main entry point ──────────────────────────────────────

// SSLIntercept intercepts TLS traffic for the given PID for `duration` seconds
// and returns a formatted APIMap string for the LLM.
// Must run as root (requires eBPF uprobe attach).
func (fs *FrozenSnapshot) SSLIntercept(pid uint32, duration int) (string, error) {
	if duration <= 0 {
		duration = 10
	}
	if duration > 120 {
		duration = 120
	}

	// Verify PID exists in snapshot
	var comm string
	if p, ok := fs.ByPID[pid]; ok {
		comm = p.Comm
	} else if p, ok := fs.Ghosts[pid]; ok {
		comm = p.Comm
	} else {
		return "", fmt.Errorf("PID %d not found in snapshot", pid)
	}

	obs, err := sslobs.New()
	if err != nil {
		return "", fmt.Errorf("create SSL observer: %w", err)
	}
	defer obs.Close()

	if err := obs.AttachPID(int(pid)); err != nil {
		// Non-fatal — process may not use any SSL library
		// We continue anyway; if no events arrive, we report that.
	}

	obs.Start()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	defer cancel()

	apiMap := newAPIMap(pid, comm)
	buffers := make(map[uint32]*streamBuffer) // TID → buffer

	deadline := time.Now().Add(time.Duration(duration) * time.Second)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			goto done
		case evt, ok := <-obs.Events:
			if !ok {
				goto done
			}
			if evt.PID != pid {
				continue
			}
			buf, exists := buffers[uint32(evt.PID)]
			if !exists {
				buf = &streamBuffer{}
				buffers[uint32(evt.PID)] = buf
			}
			if evt.Direction == sslobs.DirWrite {
				buf.writeBuf = append(buf.writeBuf, evt.Data...)
				tryParseRequest(buf, apiMap)
			} else {
				buf.readBuf = append(buf.readBuf, evt.Data...)
				tryParseResponse(buf, apiMap)
			}
		}
	}

done:
	apiMap.Duration = time.Since(apiMap.CaptureAt)
	return formatAPIMap(apiMap), nil
}

// ── HTTP parsing ─────────────────────────────────────────────────────────────

func tryParseRequest(buf *streamBuffer, m *APIMap) {
	if len(buf.writeBuf) == 0 {
		return
	}
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf.writeBuf)))
	if err != nil {
		// Data may be incomplete — keep accumulating
		// But if buffer is too large (> 64KB), reset to avoid memory leak
		if len(buf.writeBuf) > 65536 {
			buf.writeBuf = buf.writeBuf[len(buf.writeBuf)-4096:]
		}
		return
	}

	method := req.Method
	path := normalizePath(req.URL.Path)
	host := req.Host
	if host == "" {
		host = req.Header.Get("Host")
	}
	key := fmt.Sprintf("%s %s", method, path)

	authSeen, authType := detectAuth(req)

	m.mu.Lock()
	ep, ok := m.Endpoints[key]
	if !ok {
		ep = &EndpointStats{
			Method:       method,
			PathTemplate: path,
			StatusCodes:  make(map[int]int),
		}
		m.Endpoints[key] = ep
	}
	ep.Count++
	ep.LastSeen = time.Now()
	if authSeen {
		ep.AuthSeen = true
		ep.AuthType = authType
	}
	_ = host
	m.mu.Unlock()

	// Consume parsed bytes — approximate by content length + header
	// Reset buffer after a successful parse to avoid re-parsing
	buf.writeBuf = nil
}

func tryParseResponse(buf *streamBuffer, m *APIMap) {
	if len(buf.readBuf) == 0 {
		return
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(buf.readBuf)), nil)
	if err != nil {
		if len(buf.readBuf) > 65536 {
			buf.readBuf = buf.readBuf[len(buf.readBuf)-4096:]
		}
		return
	}

	m.mu.Lock()
	// Attribute to last endpoint seen — responses don't carry method/path
	// Just record the status code globally for now
	for _, ep := range m.Endpoints {
		if ep.StatusCodes == nil {
			ep.StatusCodes = make(map[int]int)
		}
		ep.StatusCodes[resp.StatusCode]++
		break // attribute to most recent
	}
	m.mu.Unlock()

	buf.readBuf = nil
}

// ── Output formatting ─────────────────────────────────────────────────────────

func formatAPIMap(m *APIMap) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[SSL Intercept: PID %d (%s)] Duration: %v, Captured at: %s\n",
		m.PID, m.Comm, m.Duration.Round(time.Millisecond), m.CaptureAt.Format("15:04:05")))
	b.WriteString(fmt.Sprintf("Endpoints discovered: %d\n\n", len(m.Endpoints)))

	if len(m.Endpoints) == 0 {
		b.WriteString("No HTTP/1.1 traffic observed. The process may:\n")
		b.WriteString("  - Use a library not attached (e.g. embedded BoringSSL)\n")
		b.WriteString("  - Not have made any requests during the capture window\n")
		b.WriteString("  - Use HTTP/2 (not yet supported — only HTTP/1.1 reconstructed)\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%-8s %-50s %-8s %-12s %s\n",
		"METHOD", "PATH TEMPLATE", "CALLS", "AUTH", "STATUS CODES"))
	b.WriteString(strings.Repeat("─", 90) + "\n")

	for _, ep := range m.Endpoints {
		authStr := "none"
		if ep.AuthSeen {
			authStr = ep.AuthType
		}
		var statusParts []string
		for status, count := range ep.StatusCodes {
			statusParts = append(statusParts, fmt.Sprintf("%d×%d", status, count))
		}
		statusStr := strings.Join(statusParts, ", ")
		if statusStr == "" {
			statusStr = "—"
		}

		path := ep.PathTemplate
		if len(path) > 50 {
			path = path[:47] + "..."
		}
		b.WriteString(fmt.Sprintf("%-8s %-50s %-8d %-12s %s\n",
			ep.Method, path, ep.Count, authStr, statusStr))
	}

	// Security signals
	b.WriteString("\nSecurity signals:\n")
	unauthCount := 0
	for _, ep := range m.Endpoints {
		if !ep.AuthSeen {
			unauthCount++
			b.WriteString(fmt.Sprintf("  ⚠ Unauthenticated endpoint: %s %s\n", ep.Method, ep.PathTemplate))
		}
	}
	if unauthCount == 0 {
		b.WriteString("  ✓ All observed endpoints used authentication\n")
	}

	return b.String()
}
