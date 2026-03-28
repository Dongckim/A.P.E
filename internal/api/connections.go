package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/dongchankim/ape/internal/sftp"
)

// ConnectionManager manages active SFTP connections.
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*sftp.Client
	order       []string // insertion order for default selection
}

// NewConnectionManager creates a new ConnectionManager.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*sftp.Client),
	}
}

// Add registers a new SFTP client with a generated ID.
func (m *ConnectionManager) Add(client *sftp.Client) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := client.Config()
	id := fmt.Sprintf("%s@%s:%s", cfg.Username, cfg.Host, cfg.Port)
	m.connections[id] = client
	m.order = append(m.order, id)
	return id
}

// Get returns the SFTP client for the given ID.
func (m *ConnectionManager) Get(id string) sftp.SFTPClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.connections[id]
	if !ok {
		return nil
	}
	return c
}

// Default returns the first (most recently added) connection, or nil.
func (m *ConnectionManager) Default() sftp.SFTPClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.order) == 0 {
		return nil
	}
	return m.connections[m.order[0]]
}

// ConnectionInfo is a summary of a connection for API responses.
type ConnectionInfo struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	Username string `json:"username"`
	Port     string `json:"port"`
}

// List returns info about all active connections.
func (m *ConnectionManager) List() []ConnectionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]ConnectionInfo, 0, len(m.order))
	for _, id := range m.order {
		c := m.connections[id]
		cfg := c.Config()
		infos = append(infos, ConnectionInfo{
			ID:       id,
			Host:     cfg.Host,
			Username: cfg.Username,
			Port:     cfg.Port,
		})
	}
	return infos
}

// CloseAll disconnects every active session.
func (m *ConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.connections {
		c.Close()
	}
	m.connections = make(map[string]*sftp.Client)
	m.order = nil
}

// HandleListConnections handles GET /api/connections
func (m *ConnectionManager) HandleListConnections(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, m.List())
}

// HandleHealth handles GET /api/health
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Data: map[string]string{"status": "ok"}, Error: ""})
}
