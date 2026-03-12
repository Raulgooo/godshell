package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	ctxengine "godshell/context"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init initializes the SQLite database at the specified path.
func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Wait for a bit to ensure the connection is ready
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		label TEXT,
		timestamp DATETIME,
		data TEXT
	);`

	if _, err = db.Exec(query); err != nil {
		return fmt.Errorf("failed to create snapshots table: %w", err)
	}

	return nil
}

// SaveSnapshot serializes a FrozenSnapshot to JSON and persists it to SQLite.
func SaveSnapshot(label string, snap *ctxengine.FrozenSnapshot) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	res, err := db.Exec("INSERT INTO snapshots (label, timestamp, data) VALUES (?, ?, ?)",
		label, snap.Timestamp, string(data))
	if err != nil {
		return 0, fmt.Errorf("failed to insert snapshot: %w", err)
	}

	return res.LastInsertId()
}

// LoadSnapshot retrieves a FrozenSnapshot by ID and deserializes the JSON data.
func LoadSnapshot(id int64) (*ctxengine.FrozenSnapshot, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var dataStr string
	err := db.QueryRow("SELECT data FROM snapshots WHERE id = ?", id).Scan(&dataStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("snapshot %d not found", id)
		}
		return nil, fmt.Errorf("failed to query snapshot: %w", err)
	}

	var snap ctxengine.FrozenSnapshot
	if err := json.Unmarshal([]byte(dataStr), &snap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}

	return &snap, nil
}

// SnapshotMeta represents metadata for a stored snapshot.
type SnapshotMeta struct {
	ID        int64     `json:"id"`
	Label     string    `json:"label"`
	Timestamp time.Time `json:"timestamp"`
}

// ListSnapshots returns a list of all stored snapshots (metadata only).
func ListSnapshots() ([]SnapshotMeta, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query("SELECT id, label, timestamp FROM snapshots ORDER BY timestamp DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	var list []SnapshotMeta
	for rows.Next() {
		var m SnapshotMeta
		if err := rows.Scan(&m.ID, &m.Label, &m.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan snapshot meta: %w", err)
		}
		list = append(list, m)
	}
	return list, nil
}
