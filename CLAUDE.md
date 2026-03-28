# CLAUDE.md — A.P.E Project Context

## What is this project?

A.P.E (AWS Platform Explorer) is a CLI tool that provides a Mac Finder-like web GUI for managing AWS EC2 files and S3 buckets. Users run `ape` from their terminal, go through interactive prompts for SSH key/host/username, and a browser-based file manager opens at localhost:9000.

## Architecture (Option A — runs on user's local machine)

- **CLI + Backend**: Go (single binary)
  - Interactive CLI prompts (no subcommands — just `ape`)
  - HTTP server using `net/http` (no framework)
  - SSH/SFTP via `golang.org/x/crypto/ssh` + `github.com/pkg/sftp`
  - S3 via `aws-sdk-go-v2`
  - Frontend embedded in binary via Go `embed` package

- **Frontend**: React 18 + TypeScript + Vite
  - Tailwind CSS for styling
  - Monaco Editor for text editing
  - Located in `frontend/` directory

## Key Design Decisions

1. **No key transmission**: SSH keys stay on the user's local machine. The Go backend reads keys from `~/.ssh/` directly via `os.ReadFile()`. Keys are used in memory only and never stored, copied, or sent over the network.

2. **localhost only**: The web server binds to `127.0.0.1:9000` only. No external access. Port 9000 is the default (configurable).

3. **Single binary**: Frontend is built and embedded into the Go binary using `//go:embed`. Users download one file.

4. **No database**: Connection state lives in memory. Config file at `~/.ape/config.yaml` for saved connections (post-MVP).

5. **Interactive CLI**: No subcommands. User runs `ape`, gets prompted for key path, host, username, port. On failure, re-prompts the failed field only. After connection, enters REPL mode with `/add`, `/list`, `/status`, `/q` commands.

## CLI Flow

```
$ ape

          ▄▄██████████▄▄
        ▄████████████████▄
       ████████████████████
       ███  (◕)    (◕)  ███
       ████     ▄▄     ████
        ████ ┌──────┐ ████
         ████│ ━━━━ │████
          ▀██└──────┘██▀
            ▀████████▀

       ██████  ██████  ██████
       ██  ██  ██  ██  ██
       ██████  ██████  ████
       ██  ██  ██      ██
       ██  ██  ██      ██████

         AWS Platform Explorer
               v0.1.0

? SSH key path (~/.ssh/id_rsa): ~/.ssh/my-key.pem
? EC2 host: 54.123.45.67
? Username (ubuntu):
? SSH port (22):

Connecting to ubuntu@54.123.45.67:22 ...
✓ Connected!
✓ Web UI ready at http://localhost:9000

──────────────────────────────────────────
A.P.E is running. Commands:
  /add     — connect additional EC2
  /list    — list active connections
  /status  — show connection info
  /q       — quit A.P.E
──────────────────────────────────────────

ape ▸
```

### CLI behavior rules:
- Default values: key=`~/.ssh/id_rsa`, username=`ubuntu`, port=`22`
- On connection failure: show error, re-prompt only the failed field
- On success: auto-open browser to `localhost:9000`
- `/add`: prompt for new EC2 connection (same flow)
- `/list`: show all active connections with status
- `/status`: show current connection details
- `/h`: show help
- `/q`: disconnect all, exit

## Project Structure

```
ape/
├── main.go              # Entry point
├── cmd/                 # CLI (interactive prompts + REPL)
│   └── root.go
├── internal/
│   ├── api/             # HTTP handlers (ec2, s3, connections)
│   ├── sftp/            # SSH/SFTP client and operations
│   ├── s3/              # AWS S3 client and operations
│   ├── config/          # App config
│   └── server/          # HTTP server setup
└── frontend/            # React + TypeScript app
    └── src/
        ├── components/  # UI components
        ├── api/         # API client
        ├── hooks/       # Custom hooks
        └── types/       # TypeScript types
```

## Coding Conventions

### Go
- Use standard library where possible (net/http over gin/echo)
- Error handling: always wrap errors with context (`fmt.Errorf("failed to list dir %s: %w", path, err)`)
- Use `internal/` for all non-CLI code
- Interfaces for testability (SFTPClient interface, S3Client interface)
- Context propagation: pass `context.Context` through all operations
- Logging: use `slog` (Go 1.21+ structured logging)

### TypeScript / React
- Functional components only, no class components
- Custom hooks for business logic (`useFileSystem`, `useS3`)
- TypeScript strict mode enabled
- API calls in `frontend/src/api/client.ts` — never in components directly
- Tailwind for all styling, no CSS files

### API Design
- REST endpoints under `/api/`
- EC2 operations: `/api/ec2/...`
- S3 operations: `/api/s3/...`
- Connection management: `/api/connections/...`
- JSON responses with consistent shape: `{ "data": ..., "error": "..." }`
- HTTP status codes: 200 (ok), 400 (bad request), 404 (not found), 500 (server error)

## Common Tasks

### Run development

```bash
# Backend (from project root)
go run main.go

# Frontend (separate terminal)
cd frontend && npm run dev
```

### Build for production

```bash
# Build frontend first
cd frontend && npm run build

# Build Go binary (embeds frontend)
go build -o ape main.go
```

### Add a new API endpoint

1. Add handler function in `internal/api/ec2_handler.go` or `s3_handler.go`
2. Register route in `internal/api/router.go`
3. Add corresponding operation in `internal/sftp/operations.go` or `internal/s3/operations.go`
4. Add TypeScript type in `frontend/src/types/index.ts`
5. Add API call function in `frontend/src/api/client.ts`

### Add a new CLI command

1. Edit `cmd/root.go`
2. Add case in the REPL switch (e.g. `/mycommand`)
3. Implement handler function

## Dependencies

### Go
- `golang.org/x/crypto/ssh` — SSH client
- `github.com/pkg/sftp` — SFTP over SSH
- `github.com/aws/aws-sdk-go-v2` — AWS SDK
- `github.com/rs/cors` — CORS middleware (dev only)

### Frontend (npm)
- `react`, `react-dom` — UI framework
- `@monaco-editor/react` — Code editor
- `tailwindcss` — Styling
- `lucide-react` — Icons

## Current Sprint

Working on: Phase 1 — Go Setup + EC2 SFTP (Day 1-3)
See docs/ROADMAP.md for full plan.
