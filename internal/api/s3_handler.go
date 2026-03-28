package api

import (
	"net/http"

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
