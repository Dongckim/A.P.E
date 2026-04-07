package api

import (
	"net/http"

	"github.com/dongchankim/ape/internal/postgres"
	"github.com/dongchankim/ape/internal/s3"
)

// NewRouter creates the API route multiplexer.
func NewRouter(connMgr *ConnectionManager, s3Client s3.S3Client, pgClient postgres.Client) *http.ServeMux {
	mux := http.NewServeMux()

	ec2 := NewEC2Handler(connMgr)

	// Health
	mux.HandleFunc("/api/health", HandleHealth)

	// Connection management (GET list, DELETE remove)
	mux.HandleFunc("/api/connections", connMgr.HandleConnections)

	// EC2 file operations
	mux.HandleFunc("/api/ec2/files", ec2.HandleListFiles)       // GET  ?path=
	mux.HandleFunc("/api/ec2/file", ec2.handleFile)             // GET/PUT/DELETE/PATCH ?path=
	mux.HandleFunc("/api/ec2/upload", ec2.HandleUploadFile)     // POST multipart
	mux.HandleFunc("/api/ec2/download", ec2.HandleDownloadFile) // GET  ?path=
	mux.HandleFunc("/api/ec2/mkdir", ec2.HandleMkdir)           // POST ?path=
	mux.HandleFunc("/api/ec2/stat", ec2.HandleStat)             // GET  ?path=

	// Dashboard
	dash := NewDashboardHandler(connMgr)
	mux.HandleFunc("/api/dashboard/overview", dash.HandleOverview)
	mux.HandleFunc("/api/dashboard/services", dash.HandleServices)
	mux.HandleFunc("/api/dashboard/git", dash.HandleGitLog)
	mux.HandleFunc("/api/dashboard/processes", dash.HandleProcesses)

	// S3 operations
	if s3Client != nil {
		s3h := NewS3Handler(s3Client)
		mux.HandleFunc("/api/s3/buckets", s3h.HandleListBuckets)     // GET
		mux.HandleFunc("/api/s3/objects", s3h.HandleListObjects)     // GET  ?bucket=&prefix=
		mux.HandleFunc("/api/s3/upload", s3h.HandleUploadObject)     // POST ?bucket=&key= multipart
		mux.HandleFunc("/api/s3/download", s3h.HandleDownloadObject) // GET  ?bucket=&key=
		mux.HandleFunc("/api/s3/object", s3h.HandleDeleteObject)     // DELETE ?bucket=&key=
		mux.HandleFunc("/api/s3/presign", s3h.HandlePresignDownload) // GET  ?bucket=&key=&expiry=
	}

	// RDS PostgreSQL operations
	if pgClient != nil {
		rds := NewRDSHandler(pgClient)
		mux.HandleFunc("/api/rds/overview", rds.HandleOverview) // GET
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
