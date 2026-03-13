package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	ctxengine "godshell/context"
	"godshell/llm"
	"godshell/store"
)

type Server struct {
	SocketPath string
	Tree       *ctxengine.ProcessTree
	Client     *http.Client
}

func NewServer(socketPath string, tree *ctxengine.ProcessTree) *Server {
	return &Server{
		SocketPath: socketPath,
		Tree:       tree,
	}
}

func (s *Server) Start() error {
	// Remove existing socket if it exists
	if err := os.RemoveAll(s.SocketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket %s: %w", s.SocketPath, err)
	}

	listener, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", s.SocketPath, err)
	}

	// Make socket accessible only to root, or change to 0666 if we want non-root to access it.
	// We want non-root TUI to access it, so we'll set 0666. The TUI is running as mortal user.
	if err := os.Chmod(s.SocketPath, 0666); err != nil {
		return fmt.Errorf("failed to chmod socket %s: %w", s.SocketPath, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/snapshot", s.handleSnapshot)
	mux.HandleFunc("/api/tools", s.handleToolProxy)
	mux.HandleFunc("/api/db/save", s.handleDBSave)
	mux.HandleFunc("/api/db/load", s.handleDBLoad)
	mux.HandleFunc("/api/db/list", s.handleDBList)
	mux.HandleFunc("/api/db/delete", s.handleDBDelete)

	fmt.Printf("Daemon listening on %s\n", s.SocketPath)
	return http.Serve(listener, mux)
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snap := s.Tree.TakeSnapshot()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snap); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode snapshot: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleToolProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toolName := r.URL.Query().Get("name")
	if toolName == "" {
		http.Error(w, "Missing tool name parameter", http.StatusBadRequest)
		return
	}

	var args map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse tool args: %v", err), http.StatusBadRequest)
		return
	}

	// Take a fresh snapshot just for this proxy execution
	// (this ensures memory/maps tools run against bleeding-edge data)
	snap := s.Tree.TakeSnapshot()

	result, execErr := llm.ExecuteToolOnSnapshot(toolName, args, snap)
	if execErr != nil {
		http.Error(w, fmt.Sprintf("Tool execution failed: %v", execErr), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, result)
}

func (s *Server) handleDBSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	label := r.URL.Query().Get("label")
	if label == "" {
		label = "manual"
	}

	var snap ctxengine.FrozenSnapshot
	if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode snapshot json: %v", err), http.StatusBadRequest)
		return
	}

	id, err := store.SaveSnapshot(label, &snap)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"id": %d}`, id)
}

func (s *Server) handleDBLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	snap, err := store.LoadSnapshot(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snap); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode snapshot: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleDBList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metas, err := store.ListSnapshots()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list snapshots: %v", err), http.StatusInternalServerError)
		return
	}

	// Just return an empty array instead of null if empty
	if metas == nil {
		metas = []store.SnapshotMeta{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metas); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode snapshot list: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleDBDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	if err := store.DeleteSnapshot(id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete snapshot: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
