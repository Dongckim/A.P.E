package api

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"

	"github.com/dongchankim/ape/internal/sftp"
)

// EC2Handler handles EC2 file operation endpoints.
type EC2Handler struct {
	connMgr *ConnectionManager
}

// NewEC2Handler creates a new EC2Handler.
func NewEC2Handler(connMgr *ConnectionManager) *EC2Handler {
	return &EC2Handler{connMgr: connMgr}
}

func (h *EC2Handler) getClient(r *http.Request) sftp.SFTPClient {
	connID := r.URL.Query().Get("conn")
	if connID != "" {
		return h.connMgr.Get(connID)
	}
	return h.connMgr.Default()
}

// HandleListFiles handles GET /api/ec2/files?path=...&conn=...
func (h *EC2Handler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	files, err := client.ListDirectory(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, files)
}

// HandleReadFile handles GET /api/ec2/file?path=...
func (h *EC2Handler) HandleReadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	content, err := client.ReadFile(path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"path":    path,
		"content": string(content),
		"size":    len(content),
	})
}

// HandleWriteFile handles PUT /api/ec2/file?path=...
// Body: JSON { "content": "file content here" }
func (h *EC2Handler) HandleWriteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	if err := client.WriteFile(path, []byte(body.Content)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"path":    path,
		"written": len(body.Content),
	})
}

// HandleUploadFile handles POST /api/ec2/upload?path=...
// Accepts multipart/form-data with a "file" field.
func (h *EC2Handler) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	destPath := r.URL.Query().Get("path")
	if destPath == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	// 32 MB max
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field: "+err.Error())
		return
	}
	defer file.Close()

	// If destPath is a directory, append the original filename
	info, statErr := client.Stat(destPath)
	if statErr == nil && info.IsDir {
		destPath = filepath.Join(destPath, header.Filename)
	}

	if err := client.UploadFile(destPath, file); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"path":     destPath,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// HandleDownloadFile handles GET /api/ec2/download?path=...
// Streams the file as an octet-stream attachment.
func (h *EC2Handler) HandleDownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	info, err := client.Stat(path)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if info.IsDir {
		writeError(w, http.StatusBadRequest, "cannot download a directory")
		return
	}

	filename := filepath.Base(path)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	if err := client.DownloadFile(path, w); err != nil {
		// Headers already sent, can't write JSON error — log only
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleDeleteFile handles DELETE /api/ec2/file?path=...
func (h *EC2Handler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	if err := client.DeleteFile(path); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"deleted": path,
	})
}

// HandleRenameFile handles PATCH /api/ec2/file?path=...&dest=...
func (h *EC2Handler) HandleRenameFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	oldPath := r.URL.Query().Get("path")
	newPath := r.URL.Query().Get("dest")
	if oldPath == "" || newPath == "" {
		writeError(w, http.StatusBadRequest, "both 'path' and 'dest' query params are required")
		return
	}

	if err := client.RenameFile(oldPath, newPath); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"old_path": oldPath,
		"new_path": newPath,
	})
}

// HandleMkdir handles POST /api/ec2/mkdir?path=...
func (h *EC2Handler) HandleMkdir(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	if err := client.MkdirAll(path); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"created": path,
	})
}

// HandleStat handles GET /api/ec2/stat?path=...
func (h *EC2Handler) HandleStat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	info, err := client.Stat(path)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// Response is the standard API response shape.
type Response struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Data: data, Error: ""})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Data: nil, Error: msg})
}

// Ensure io import is used (for HandleUploadFile file.Close via multipart).
var _ io.Reader
