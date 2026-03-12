package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "godshell"
	keyringUser    = "api_key"
)

// Config stores persistent application settings.
type Config struct {
	// APIKey is exclusively managed via OS keyring and environment variables,
	// and never written to plain-text config files.
	APIKey string `json:"-"`

	// Threat intel keys — loaded from environment variables only, never written to disk.
	VTApiKey   string `json:"-"` // VIRUSTOTAL_API_KEY
	AbuseIPKey string `json:"-"` // ABUSEIPDB_API_KEY

	Model                string `json:"model"`
	Endpoint             string `json:"endpoint"`
	SnapshotIntervalSec  int    `json:"snapshot_interval_sec"`
	SnapshotRetentionSec int    `json:"snapshot_retention_sec"`
	SnapshotConcurrency  int    `json:"snapshot_concurrency"`
	DBPath               string `json:"db_path"`
}

func Default() Config {
	return Config{
		Model:                "google/gemini-2.5-flash-preview",
		Endpoint:             "https://openrouter.ai/api/v1/chat/completions",
		SnapshotIntervalSec:  0,
		SnapshotRetentionSec: 60,
		SnapshotConcurrency:  1,
		DBPath:               filepath.Join(configDir(), "godshell.db"),
	}
}

func configDir() string {
	home, _ := os.UserHomeDir()

	// If running under sudo, prefer the caller's home directory.
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if home == "/root" {
			if _, err := os.Stat("/home/" + sudoUser); err == nil {
				home = "/home/" + sudoUser
			}
		}
	}

	return filepath.Join(home, ".config", "godshell")
}

func ConfigPath() string {
	return filepath.Join(configDir(), "config.json")
}

func EnsureDir() error {
	return os.MkdirAll(configDir(), 0700)
}

// Load reads config from disk and the API key from the OS keyring.
func Load() (Config, error) {
	cfg := Default()

	// Ensure DBUS session is bridged if running under sudo
	bridgeDBus()

	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// Try keyring as the primary source
	keyringKey := loadKeyring()
	if keyringKey != "" {
		cfg.APIKey = keyringKey
	}

	// Threat intel keys from environment (never stored on disk)
	cfg.VTApiKey = os.Getenv("VIRUSTOTAL_API_KEY")
	cfg.AbuseIPKey = os.Getenv("ABUSEIPDB_API_KEY")

	return cfg, nil
}

// Save writes config to disk and the API key to the OS keyring.
func Save(cfg Config) error {
	if err := EnsureDir(); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	bridgeDBus()

	if cfg.APIKey != "" {
		if err := keyring.Set(keyringService, keyringUser, cfg.APIKey); err != nil {
			// If keyring fails, we don't have a fallback anymore per user request.
			// But we warn them so they know it might prompt again.
			fmt.Fprintf(os.Stderr, "warning: keyring set failed: %v\n", err)
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(ConfigPath(), data, 0600)
}

// bridgeDBus ensures that if we are running as root via sudo, we point
// to the user's DBUS session bus so we can access their keyring.
func bridgeDBus() {
	uid := os.Getenv("SUDO_UID")
	if uid == "" {
		// If NOT sudo, only return if ALREADY set to a valid address
		if os.Getenv("DBUS_SESSION_BUS_ADDRESS") != "" {
			return
		}
		// Try to find it if missing (unlikely on modern systems but helpful)
		if addr, err := os.UserHomeDir(); err == nil {
			bus := fmt.Sprintf("unix:path=/run/user/%d/bus", os.Getuid())
			if _, err := os.Stat(fmt.Sprintf("/run/user/%d/bus", os.Getuid())); err == nil {
				os.Setenv("DBUS_SESSION_BUS_ADDRESS", bus)
			}
			_ = addr
		}
		return
	}

	// If SUDO_UID is present, ALWAYS try to override to that user's session bus
	busPath := fmt.Sprintf("unix:path=/run/user/%s/bus", uid)
	if _, err := os.Stat(fmt.Sprintf("/run/user/%s/bus", uid)); err == nil {
		os.Setenv("DBUS_SESSION_BUS_ADDRESS", busPath)
	}
}

func loadKeyring() string {
	bridgeDBus() // Ensure bridge is applied before any keyring call
	secret, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return ""
	}
	return secret
}
