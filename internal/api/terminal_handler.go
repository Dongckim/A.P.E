package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"golang.org/x/crypto/ssh"
	"nhooyr.io/websocket"

	sftppkg "github.com/dongchankim/ape/internal/sftp"
)

// TerminalHandler handles WebSocket-based SSH terminal sessions.
type TerminalHandler struct {
	connMgr *ConnectionManager
}

// NewTerminalHandler creates a new TerminalHandler.
func NewTerminalHandler(connMgr *ConnectionManager) *TerminalHandler {
	return &TerminalHandler{connMgr: connMgr}
}

// resizeMsg is the JSON payload for terminal resize events.
type resizeMsg struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func (h *TerminalHandler) getClient(r *http.Request) *sftppkg.Client {
	connID := r.URL.Query().Get("conn")
	if connID != "" {
		return h.connMgr.GetRaw(connID)
	}
	return h.connMgr.DefaultRaw()
}

// HandleTerminal handles GET /api/ec2/terminal — upgrades to WebSocket and
// bridges it to an SSH PTY session on the connected EC2 instance.
func (h *TerminalHandler) HandleTerminal(w http.ResponseWriter, r *http.Request) {
	client := h.getClient(r)
	if client == nil {
		writeError(w, http.StatusBadRequest, "no active connection")
		return
	}

	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// localhost only — no origin check needed.
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("websocket accept failed", "err", err)
		return
	}
	defer ws.Close(websocket.StatusNormalClosure, "session ended")

	session, err := client.NewSession()
	if err != nil {
		slog.Error("failed to create SSH session", "err", err)
		ws.Close(websocket.StatusInternalError, "failed to create SSH session")
		return
	}
	defer session.Close()

	// Request PTY with reasonable defaults; client will send a resize soon.
	if err := session.RequestPty("xterm-256color", 24, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		slog.Error("failed to request PTY", "err", err)
		ws.Close(websocket.StatusInternalError, "failed to request PTY")
		return
	}

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		slog.Error("failed to get stdin pipe", "err", err)
		return
	}
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		slog.Error("failed to get stdout pipe", "err", err)
		return
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		slog.Error("failed to get stderr pipe", "err", err)
		return
	}

	if err := session.Shell(); err != nil {
		slog.Error("failed to start shell", "err", err)
		ws.Close(websocket.StatusInternalError, "failed to start shell")
		return
	}

	slog.Info("terminal session started", "remote", client.Config().Host)

	// Use a standalone context so WebSocket I/O is not tied to the HTTP
	// request lifecycle (r.Context() may be cancelled after hijack).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// SSH stdout → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 8192)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				if writeErr := ws.Write(ctx, websocket.MessageBinary, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// SSH stderr → WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				if writeErr := ws.Write(ctx, websocket.MessageBinary, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// WebSocket → SSH stdin (binary = keystrokes, text = control JSON)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stdinPipe.Close()
		for {
			msgType, data, err := ws.Read(ctx)
			if err != nil {
				return
			}

			if msgType == websocket.MessageText {
				var msg resizeMsg
				if json.Unmarshal(data, &msg) == nil && msg.Type == "resize" {
					if msg.Cols > 0 && msg.Rows > 0 {
						_ = session.WindowChange(msg.Rows, msg.Cols)
					}
				}
				continue
			}

			// Binary message = raw terminal input
			if _, err := stdinPipe.Write(data); err != nil {
				return
			}
		}
	}()

	// Wait for shell to exit, then clean up.
	_ = session.Wait()
	slog.Info("terminal session ended", "remote", client.Config().Host)
	cancel()
	ws.Close(websocket.StatusNormalClosure, "shell exited")
	_, _ = io.ReadAll(stdoutPipe)
	_, _ = io.ReadAll(stderrPipe)
	wg.Wait()
}
