package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			apiMux.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>A.P.E — AWS Platform Explorer</title>
<style>
  body { font-family: -apple-system, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #0f172a; color: #e2e8f0; }
  .container { text-align: center; }
  h1 { font-size: 2.5rem; margin-bottom: 0.5rem; }
  p { color: #94a3b8; font-size: 1.1rem; }
  pre { font-size: 1.2rem; color: #22d3ee; }
</style>
</head>
<body>
  <div class="container">
    <pre>🦍 A.P.E</pre>
    <h1>AWS Platform Explorer</h1>
    <p>v0.1.0 — Hello from A.P.E!</p>
    <p>Web UI coming soon.</p>
  </div>
</body>
</html>`)
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		ConnMgr: connMgr,
	}
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
