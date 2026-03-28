package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dongchankim/ape/frontend"
	"github.com/dongchankim/ape/internal/api"
	"github.com/dongchankim/ape/internal/s3"
)

type Server struct {
	httpServer *http.Server
	ConnMgr    *api.ConnectionManager
}

func New(addr string, connMgr *api.ConnectionManager, s3Client s3.S3Client) *Server {
	apiMux := api.NewRouter(connMgr, s3Client)

	// Load embedded frontend
	var frontendHandler http.Handler
	distFS, err := fs.Sub(frontend.Dist, "dist")
	if err != nil {
		slog.Error("failed to load embedded frontend", "err", err)
		frontendHandler = http.HandlerFunc(fallbackPage)
	} else {
		slog.Info("serving embedded frontend")
		fileServer := http.FileServer(http.FS(distFS))
		frontendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			if f, err := distFS.Open(path[1:]); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
			// SPA fallback
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
		})
	}

	// Single top-level handler: API first, then frontend
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			apiMux.ServeHTTP(w, r)
			return
		}
		frontendHandler.ServeHTTP(w, r)
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
		ConnMgr: connMgr,
	}
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
    <p>Frontend not built. Run:</p>
    <p><code>cd frontend && npm install && npm run build</code></p>
    <p>Then <code>make build</code></p>
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
