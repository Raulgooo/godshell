// Package intel provides clients for VirusTotal and AbuseIPDB threat intel lookups.
// Results are cached aggressively to stay within free-tier rate limits.
package intel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ── Cache ──────────────────────────────────────────────────────────────────

type cachedResult struct {
	Value     interface{}
	CachedAt  time.Time
	ExpiresAt time.Time
}

// Client holds API keys and the shared result cache.
type Client struct {
	VTKey      string // VirusTotal API key
	AbuseIPKey string // AbuseIPDB API key
	httpClient *http.Client

	mu    sync.RWMutex
	cache map[string]*cachedResult

	// Rate limit for VirusTotal free tier: 4 req/min
	vtLimiter *rateLimiter
}

type rateLimiter struct {
	mu       sync.Mutex
	lastReqs []time.Time
	maxReqs  int
	window   time.Duration
}

func newRateLimiter(maxReqs int, window time.Duration) *rateLimiter {
	return &rateLimiter{maxReqs: maxReqs, window: window}
}

func (r *rateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	// Remove expired timestamps
	cutoff := now.Add(-r.window)
	i := 0
	for i < len(r.lastReqs) && r.lastReqs[i].Before(cutoff) {
		i++
	}
	r.lastReqs = r.lastReqs[i:]
	// If at limit, sleep until oldest expires
	if len(r.lastReqs) >= r.maxReqs {
		wait := r.window - now.Sub(r.lastReqs[0]) + time.Millisecond*50
		if wait > 0 {
			r.mu.Unlock()
			time.Sleep(wait)
			r.mu.Lock()
		}
	}
	r.lastReqs = append(r.lastReqs, time.Now())
}

// New creates a new intel Client. Keys can be empty strings (lookups return graceful empty).
func New(vtKey, abuseIPKey string) *Client {
	return &Client{
		VTKey:      vtKey,
		AbuseIPKey: abuseIPKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      make(map[string]*cachedResult),
		vtLimiter:  newRateLimiter(4, time.Minute),
	}
}

func (c *Client) cacheGet(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.cache[key]
	if !ok || time.Now().After(r.ExpiresAt) {
		return nil, false
	}
	return r.Value, true
}

func (c *Client) cachePut(key string, val interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = &cachedResult{
		Value:     val,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
}

// ── VirusTotal ─────────────────────────────────────────────────────────────

// VTResult summarizes VirusTotal verdict for a hash or domain.
type VTResult struct {
	Malicious  int    `json:"malicious"`
	Suspicious int    `json:"suspicious"`
	Harmless   int    `json:"harmless"`
	Undetected int    `json:"undetected"`
	Summary    string `json:"summary"`
}

// VTEmpty returns a VTResult indicating no data was available.
func VTEmpty() *VTResult {
	return &VTResult{Summary: "no VT data (no API key or lookup failed)"}
}

// LookupHash queries VirusTotal for a SHA-256 hash.
// Returns cached result if available. Non-blocking on API failure.
func (c *Client) LookupHash(sha256 string) *VTResult {
	if c.VTKey == "" || sha256 == "" {
		return VTEmpty()
	}
	key := "vt:hash:" + sha256
	if v, ok := c.cacheGet(key); ok {
		return v.(*VTResult)
	}

	c.vtLimiter.Wait()
	url := fmt.Sprintf("https://www.virustotal.com/api/v3/files/%s", sha256)
	result := c.vtGET(url)
	c.cachePut(key, result, time.Hour)
	return result
}

// LookupDomain queries VirusTotal for a domain.
func (c *Client) LookupDomain(domain string) *VTResult {
	if c.VTKey == "" || domain == "" {
		return VTEmpty()
	}
	key := "vt:domain:" + domain
	if v, ok := c.cacheGet(key); ok {
		return v.(*VTResult)
	}

	c.vtLimiter.Wait()
	url := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s", domain)
	result := c.vtGET(url)
	c.cachePut(key, result, time.Hour)
	return result
}

func (c *Client) vtGET(url string) *VTResult {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return VTEmpty()
	}
	req.Header.Set("x-apikey", c.VTKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return VTEmpty()
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &VTResult{Summary: "not found in VirusTotal"}
	}
	if resp.StatusCode != 200 {
		return &VTResult{Summary: fmt.Sprintf("VT API error %d", resp.StatusCode)}
	}

	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Data struct {
			Attributes struct {
				LastAnalysisStats struct {
					Malicious  int `json:"malicious"`
					Suspicious int `json:"suspicious"`
					Harmless   int `json:"harmless"`
					Undetected int `json:"undetected"`
				} `json:"last_analysis_stats"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return VTEmpty()
	}

	s := payload.Data.Attributes.LastAnalysisStats
	r := &VTResult{
		Malicious:  s.Malicious,
		Suspicious: s.Suspicious,
		Harmless:   s.Harmless,
		Undetected: s.Undetected,
	}
	if s.Malicious > 0 {
		r.Summary = fmt.Sprintf("⚠ MALICIOUS: %d/72 engines flagged", s.Malicious)
	} else if s.Suspicious > 0 {
		r.Summary = fmt.Sprintf("suspicious: %d engines", s.Suspicious)
	} else if s.Harmless > 0 {
		r.Summary = fmt.Sprintf("clean (%d harmless)", s.Harmless)
	} else {
		r.Summary = "no detections"
	}
	return r
}

// ── AbuseIPDB ──────────────────────────────────────────────────────────────

// AbuseResult summarizes AbuseIPDB verdict for an IP address.
type AbuseResult struct {
	IP           string `json:"ip_address"`
	AbuseScore   int    `json:"abuse_confidence_score"`
	TotalReports int    `json:"total_reports"`
	CountryCode  string `json:"country_code"`
	ISP          string `json:"isp"`
	Summary      string `json:"summary"`
}

// AbuseEmpty returns an AbuseResult indicating no data.
func AbuseEmpty() *AbuseResult {
	return &AbuseResult{Summary: "no AbuseIPDB data (no API key or lookup failed)"}
}

// LookupIP queries AbuseIPDB for an IP address reputation.
func (c *Client) LookupIP(ip string) *AbuseResult {
	if c.AbuseIPKey == "" || ip == "" {
		return AbuseEmpty()
	}
	// Skip private/loopback
	if isPrivateIP(ip) {
		return &AbuseResult{IP: ip, Summary: "private IP — no lookup needed"}
	}
	key := "abuse:" + ip
	if v, ok := c.cacheGet(key); ok {
		return v.(*AbuseResult)
	}

	req, err := http.NewRequest("GET", "https://api.abuseipdb.com/api/v2/check", nil)
	if err != nil {
		return AbuseEmpty()
	}
	q := req.URL.Query()
	q.Add("ipAddress", ip)
	q.Add("maxAgeInDays", "90")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Key", c.AbuseIPKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return AbuseEmpty()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &AbuseResult{Summary: fmt.Sprintf("AbuseIPDB API error %d", resp.StatusCode)}
	}

	body, _ := io.ReadAll(resp.Body)
	var payload struct {
		Data struct {
			IPAddress            string `json:"ipAddress"`
			AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
			TotalReports         int    `json:"totalReports"`
			CountryCode          string `json:"countryCode"`
			ISP                  string `json:"isp"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return AbuseEmpty()
	}

	d := payload.Data
	r := &AbuseResult{
		IP:           d.IPAddress,
		AbuseScore:   d.AbuseConfidenceScore,
		TotalReports: d.TotalReports,
		CountryCode:  d.CountryCode,
		ISP:          d.ISP,
	}

	if r.TotalReports == 0 {
		r.Summary = fmt.Sprintf("clean (%s, %s)", r.ISP, r.CountryCode)
	} else if r.AbuseScore >= 75 {
		r.Summary = fmt.Sprintf("⚠ HIGH RISK: score=%d, %d reports (%s)", r.AbuseScore, r.TotalReports, r.ISP)
	} else if r.AbuseScore >= 25 {
		r.Summary = fmt.Sprintf("suspicious: score=%d, %d reports (%s)", r.AbuseScore, r.TotalReports, r.ISP)
	} else {
		r.Summary = fmt.Sprintf("low risk: score=%d, %d reports (%s)", r.AbuseScore, r.TotalReports, r.ISP)
	}
	c.cachePut(key, r, time.Hour)
	return r
}

// isPrivateIP returns true for loopback and RFC1918 addresses.
func isPrivateIP(ip string) bool {
	private := []string{
		"127.", "::1", "10.", "192.168.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
	}
	for _, prefix := range private {
		if len(ip) >= len(prefix) && ip[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
