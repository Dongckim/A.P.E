package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager()
	if cm == nil {
		t.Fatal("expected non-nil ConnectionManager")
	}
}

func TestDefaultReturnsNilWhenEmpty(t *testing.T) {
	cm := NewConnectionManager()
	if cm.Default() != nil {
		t.Fatal("expected nil default on empty manager")
	}
}

func TestListReturnsEmptySlice(t *testing.T) {
	cm := NewConnectionManager()
	list := cm.List()
	if len(list) != 0 {
		t.Fatalf("expected 0 connections, got %d", len(list))
	}
}

func TestGetReturnsNilForMissingID(t *testing.T) {
	cm := NewConnectionManager()
	if cm.Get("nonexistent") != nil {
		t.Fatal("expected nil for non-existent ID")
	}
}

func TestRemoveReturnsFalseForMissingID(t *testing.T) {
	cm := NewConnectionManager()
	if cm.Remove("nonexistent") {
		t.Fatal("expected false when removing non-existent ID")
	}
}

func TestCloseAllOnEmpty(t *testing.T) {
	cm := NewConnectionManager()
	cm.CloseAll() // should not panic
}

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatal("expected data to be a map")
	}
	if data["status"] != "ok" {
		t.Fatalf("expected status 'ok', got '%v'", data["status"])
	}
}

func TestHandleConnectionsGetEmpty(t *testing.T) {
	cm := NewConnectionManager()
	req := httptest.NewRequest(http.MethodGet, "/api/connections", nil)
	w := httptest.NewRecorder()

	cm.HandleConnections(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleConnectionsDeleteNoID(t *testing.T) {
	cm := NewConnectionManager()
	req := httptest.NewRequest(http.MethodDelete, "/api/connections", nil)
	w := httptest.NewRecorder()

	cm.HandleConnections(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleConnectionsDeleteNotFound(t *testing.T) {
	cm := NewConnectionManager()
	req := httptest.NewRequest(http.MethodDelete, "/api/connections?id=fake", nil)
	w := httptest.NewRecorder()

	cm.HandleConnections(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleConnectionsMethodNotAllowed(t *testing.T) {
	cm := NewConnectionManager()
	req := httptest.NewRequest(http.MethodPost, "/api/connections", nil)
	w := httptest.NewRecorder()

	cm.HandleConnections(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
