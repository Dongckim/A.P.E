package api

import (
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/dongchankim/ape/internal/s3"
)

// S3Handler handles S3 API endpoints.
type S3Handler struct {
	client s3.S3Client
}

// NewS3Handler creates a new S3Handler.
func NewS3Handler(client s3.S3Client) *S3Handler {
	return &S3Handler{client: client}
}

// HandleListBuckets handles GET /api/s3/buckets
func (h *S3Handler) HandleListBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	buckets, err := h.client.ListBuckets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, buckets)
}

// HandleListObjects handles GET /api/s3/objects?bucket=...&prefix=...
func (h *S3Handler) HandleListObjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		writeError(w, http.StatusBadRequest, "bucket is required")
		return
	}

	prefix := r.URL.Query().Get("prefix")

	objects, err := h.client.ListObjects(r.Context(), bucket, prefix)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, objects)
}

// HandleUploadObject handles POST /api/s3/upload?bucket=...&key=...
// Accepts multipart/form-data with a "file" field.
func (h *S3Handler) HandleUploadObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")
	if bucket == "" || key == "" {
		writeError(w, http.StatusBadRequest, "bucket and key are required")
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

	if err := h.client.UploadObject(r.Context(), bucket, key, file); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bucket":   bucket,
		"key":      key,
		"filename": header.Filename,
		"size":     header.Size,
	})
}

// HandleDownloadObject handles GET /api/s3/download?bucket=...&key=...
func (h *S3Handler) HandleDownloadObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")
	if bucket == "" || key == "" {
		writeError(w, http.StatusBadRequest, "bucket and key are required")
		return
	}

	filename := path.Base(key)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	if err := h.client.DownloadObject(r.Context(), bucket, key, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleDeleteObject handles DELETE /api/s3/object?bucket=...&key=...
func (h *S3Handler) HandleDeleteObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")
	if bucket == "" || key == "" {
		writeError(w, http.StatusBadRequest, "bucket and key are required")
		return
	}

	if err := h.client.DeleteObject(r.Context(), bucket, key); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"deleted": bucket + "/" + key,
	})
}

// HandlePresignDownload handles GET /api/s3/presign?bucket=...&key=...&expiry=...
// expiry is in seconds (default: 3600 = 1 hour).
func (h *S3Handler) HandlePresignDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bucket := r.URL.Query().Get("bucket")
	key := r.URL.Query().Get("key")
	if bucket == "" || key == "" {
		writeError(w, http.StatusBadRequest, "bucket and key are required")
		return
	}

	expiry := 3600
	if e := r.URL.Query().Get("expiry"); e != "" {
		if v, err := strconv.Atoi(e); err == nil && v > 0 {
			expiry = v
		}
	}

	url, err := h.client.PresignDownload(r.Context(), bucket, key, time.Duration(expiry)*time.Second)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"url":        url,
		"expires_in": expiry,
	})
}
