package api

import (
	"net/http"

	"github.com/dongchankim/ape/internal/s3"
)

// NewRouter creates the API route multiplexer.
func NewRouter(connMgr *ConnectionManager, s3Client s3.S3Client) *http.ServeMux {
	mux := http.NewServeMux()

	ec2 := NewEC2Handler(connMgr)

	// Health
	mux.HandleFunc("/api/health", HandleHealth)

	// Connection management
	mux.HandleFunc("/api/connections", connMgr.HandleListConnections)

	// EC2 file operations
	mux.HandleFunc("/api/ec2/files", ec2.HandleListFiles)      // GET  ?path=
	mux.HandleFunc("/api/ec2/file", ec2.handleFile)             // GET/PUT/DELETE/PATCH ?path=
	mux.HandleFunc("/api/ec2/upload", ec2.HandleUploadFile)     // POST multipart
	mux.HandleFunc("/api/ec2/download", ec2.HandleDownloadFile) // GET  ?path=
	mux.HandleFunc("/api/ec2/mkdir", ec2.HandleMkdir)           // POST ?path=
	mux.HandleFunc("/api/ec2/stat", ec2.HandleStat)             // GET  ?path=

	// S3 operations
	if s3Client != nil {
		s3h := NewS3Handler(s3Client)
		mux.HandleFunc("/api/s3/buckets", s3h.HandleListBuckets) // GET
		mux.HandleFunc("/api/s3/objects", s3h.HandleListObjects)  // GET ?bucket=&prefix=
	}

	return mux
}

// handleFile routes /api/ec2/file to the correct handler by HTTP method.
func (h *EC2Handler) handleFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.HandleReadFile(w, r)
	case http.MethodPut:
		h.HandleWriteFile(w, r)
	case http.MethodDelete:
		h.HandleDeleteFile(w, r)
	case http.MethodPatch:
		h.HandleRenameFile(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
