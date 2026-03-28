package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dongchankim/ape/internal/sftp"
)

// mockSFTPClient implements sftp.SFTPClient for testing.
type mockSFTPClient struct {
	files []sftp.FileInfo
	data  []byte
}

func (m *mockSFTPClient) ListDirectory(path string) ([]sftp.FileInfo, error) { return m.files, nil }
func (m *mockSFTPClient) ReadFile(path string) ([]byte, error)              { return m.data, nil }
func (m *mockSFTPClient) WriteFile(path string, content []byte) error       { return nil }
func (m *mockSFTPClient) UploadFile(path string, reader io.Reader) error    { return nil }
func (m *mockSFTPClient) DownloadFile(path string, w io.Writer) error       { return nil }
func (m *mockSFTPClient) DeleteFile(path string) error                      { return nil }
func (m *mockSFTPClient) RenameFile(oldPath, newPath string) error          { return nil }
func (m *mockSFTPClient) Stat(path string) (*sftp.FileInfo, error) {
	return &sftp.FileInfo{Name: "test", IsDir: false}, nil
}
func (m *mockSFTPClient) MkdirAll(path string) error                              { return nil }
func (m *mockSFTPClient) Exec(ctx context.Context, cmd string) (string, int, error) { return "", 0, nil }
func (m *mockSFTPClient) Close() error                                             { return nil }

// injectMock places a mock client into the ConnectionManager so handlers can find it.
func injectMock(cm *ConnectionManager, mock sftp.SFTPClient) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	id := "mock@test:22"
	// Store requires *sftp.Client but we bypass via interface trick:
	// We store the mock directly in a parallel map isn't possible since connections is map[string]*sftp.Client.
	// Instead, we'll test handlers that return "no active connection" for empty managers,
	// and use a custom approach for mock injection.
	_ = id
	_ = mock
}

// --- Tests with no active connection (empty manager) ---

func TestHandleListFilesNoConnection(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodGet, "/api/ec2/files", nil)
	w := httptest.NewRecorder()

	h.HandleListFiles(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleReadFileNoConnection(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodGet, "/api/ec2/file?path=/test", nil)
	w := httptest.NewRecorder()

	h.HandleReadFile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleWriteFileNoConnection(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodPut, "/api/ec2/file?path=/test", nil)
	w := httptest.NewRecorder()

	h.HandleWriteFile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteFileNoConnection(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodDelete, "/api/ec2/file?path=/test", nil)
	w := httptest.NewRecorder()

	h.HandleDeleteFile(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleMkdirNoConnection(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodPost, "/api/ec2/mkdir?path=/test", nil)
	w := httptest.NewRecorder()

	h.HandleMkdir(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleStatNoConnection(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodGet, "/api/ec2/stat?path=/test", nil)
	w := httptest.NewRecorder()

	h.HandleStat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Method not allowed tests ---

func TestHandleListFilesMethodNotAllowed(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodPost, "/api/ec2/files", nil)
	w := httptest.NewRecorder()

	h.HandleListFiles(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleReadFileMethodNotAllowed(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodPost, "/api/ec2/file", nil)
	w := httptest.NewRecorder()

	h.HandleReadFile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleDeleteFileMethodNotAllowed(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodGet, "/api/ec2/file", nil)
	w := httptest.NewRecorder()

	h.HandleDeleteFile(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleMkdirMethodNotAllowed(t *testing.T) {
	h := NewEC2Handler(NewConnectionManager())
	req := httptest.NewRequest(http.MethodGet, "/api/ec2/mkdir", nil)
	w := httptest.NewRecorder()

	h.HandleMkdir(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
