package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dongchankim/ape/internal/api"
	"github.com/dongchankim/ape/internal/s3"
)

type Server struct {
	httpServer *http.Server
	ConnMgr    *api.ConnectionManager
}

func New(addr string, connMgr *api.ConnectionManager, s3Client s3.S3Client) *Server {
	apiMux := api.NewRouter(connMgr, s3Client)

	mux := http.NewServeMux()

	// API routes
	mux.Handle("/api/", apiMux)

	// Serve React frontend from frontend/dist
	distPath := findDistDir()
	if distPath != "" {
		slog.Info("serving frontend", "path", distPath)
		frontendFS := http.FileServer(http.Dir(distPath))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// If the file exists, serve it directly (JS, CSS, assets)
			path := distPath + r.URL.Path
			if _, err := os.Stat(path); err == nil && r.URL.Path != "/" {
				frontendFS.ServeHTTP(w, r)
				return
			}
			// Otherwise serve index.html (SPA fallback)
			http.ServeFile(w, r, distPath+"/index.html")
		})
	} else {
		slog.Warn("frontend/dist not found — run 'cd frontend && npm run build'")
		mux.HandleFunc("/", fallbackPage)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		ConnMgr: connMgr,
	}
}

func findDistDir() string {
	candidates := []string{
		"frontend/dist",
		"../frontend/dist",
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			if entries, _ := fs.Glob(os.DirFS(p), "index.html"); len(entries) > 0 {
				return p
			}
		}
	}
	return ""
}

func fallbackPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>A.P.E</title>
<style>
  body { font-family: -apple-system, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #0f172a; color: #e2e8f0; }
  .c { text-align: center; } h1 { font-size: 2rem; } p { color: #94a3b8; } code { background: #1e293b; padding: 4px 8px; border-radius: 4px; color: #22d3ee; }
</style>
</head>
<body>
  <div class="c">
    <h1>A.P.E</h1>
    <p>Frontend not built yet. Run:</p>
    <p><code>cd frontend && npm install && npm run build</code></p>
    <p>Then restart <code>ape</code></p>
  </div>
</body>
</html>`))
}

func (s *Server) Start() error {
	slog.Info("web UI ready", "url", "http://"+s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	s.ConnMgr.CloseAll()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
