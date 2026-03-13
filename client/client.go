package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	ctxengine "godshell/context"
	"godshell/store"
)

type DaemonClient struct {
	httpClient *http.Client
	socketPath string
}

func NewDaemonClient(socketPath string) *DaemonClient {
	return &DaemonClient{
		socketPath: socketPath,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
		},
	}
}

// GetSnapshot retrieves the latest ProcessTree snapshot from the daemon
func (c *DaemonClient) GetSnapshot() (*ctxengine.FrozenSnapshot, error) {
	resp, err := c.httpClient.Get("http://localhost/api/snapshot")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch snapshot from daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}

	var snap ctxengine.FrozenSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return nil, fmt.Errorf("failed to decode snapshot json: %w", err)
	}

	return &snap, nil
}

// ExecuteTool forwards a privileged LLM tool call to the daemon.
func (c *DaemonClient) ExecuteTool(toolName string, args map[string]interface{}) (string, error) {
	reqBody, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool args: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/api/tools?name="+toolName, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create tool request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("daemon tool execution failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read daemon response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("daemon tool error (%d): %s", resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}

// SaveSnapshot proxies saving a snapshot to the daemon's DB.
func (c *DaemonClient) SaveSnapshot(label string, snap *ctxengine.FrozenSnapshot) (int64, error) {
	reqBody, err := json.Marshal(snap)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	url := "http://localhost/api/db/save?label=" + label
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create save request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("daemon save failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("daemon save error (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, fmt.Errorf("failed to decode daemon save response: %w", err)
	}

	return res.ID, nil
}

// LoadSnapshot proxies loading a snapshot from the daemon's DB.
func (c *DaemonClient) LoadSnapshot(id int64) (*ctxengine.FrozenSnapshot, error) {
	url := fmt.Sprintf("http://localhost/api/db/load?id=%d", id)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request snapshot from daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("daemon load error (%d): %s", resp.StatusCode, string(body))
	}

	var snap ctxengine.FrozenSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return nil, fmt.Errorf("failed to decode daemon load response: %w", err)
	}

	return &snap, nil
}

// ListSnapshots proxies listing snapshots from the daemon's DB.
func (c *DaemonClient) ListSnapshots() ([]store.SnapshotMeta, error) {
	resp, err := c.httpClient.Get("http://localhost/api/db/list")
	if err != nil {
		return nil, fmt.Errorf("failed to request snapshot list from daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("daemon list error (%d): %s", resp.StatusCode, string(body))
	}

	var metas []store.SnapshotMeta
	if err := json.NewDecoder(resp.Body).Decode(&metas); err != nil {
		return nil, fmt.Errorf("failed to decode daemon list response: %w", err)
	}

	return metas, nil
}

// DeleteSnapshot proxies deleting a snapshot from the daemon's DB.
func (c *DaemonClient) DeleteSnapshot(id int64) error {
	url := fmt.Sprintf("http://localhost/api/db/delete?id=%d", id)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("daemon delete failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("daemon delete error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}
