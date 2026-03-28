# A.P.E — AWS Platform Explorer

> Mac Finder-like GUI for managing your AWS EC2 files and S3 buckets — right from your browser.

A.P.E is a self-hosted, open-source CLI tool that lets you visually manage files on your AWS EC2 instances and S3 buckets through a clean web interface. No more memorizing terminal commands — just connect and browse.

## How it works

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
  Opening browser...
```

1. Run `ape` — interactive prompt starts
2. Enter your SSH key path, EC2 host, and username
3. A.P.E uses your local SSH key to connect via SFTP
4. A local web server starts at `localhost:9000` and opens your browser
5. Browse, upload, download, edit, and delete files — all through a GUI
6. Type `/q` to quit, `/add` to connect more servers

**Your SSH keys never leave your machine.** A.P.E runs entirely on your local Mac/Linux and connects to EC2 via SSH/SFTP. No keys are transmitted over the network.

## Architecture

```
┌─────────────────────────────────────────────┐
│  Your Mac / Linux                           │
│                                             │
│  ┌─────────┐    ┌──────────────────────┐    │
│  │ Browser  │◄──►│  A.P.E Go Server    │    │
│  │ :9000    │    │  (localhost only)    │    │
│  └─────────┘    └──────┬───────┬───────┘    │
│                        │       │            │
└────────────────────────┼───────┼────────────┘
                   SSH/SFTP   AWS SDK
                         │       │
                   ┌─────▼──┐  ┌─▼────┐
                   │  EC2   │  │  S3  │
                   │ files  │  │bucket│
                   └────────┘  └──────┘
```

- **Frontend**: React + TypeScript (Finder-style file explorer)
- **Backend**: Go (HTTP server + SSH/SFTP client + AWS SDK)
- **Protocol**: SSH/SFTP for EC2 files, AWS SDK Go v2 for S3
- **Security**: Keys stay local, server binds to localhost only

## Features (MVP)

### EC2 File Management (via SFTP)
- [ ] Browse files and folders (grid/list view)
- [ ] Upload files (drag & drop)
- [ ] Download files
- [ ] Delete files and folders
- [ ] Rename files and folders
- [ ] Edit text files (built-in editor)

### S3 Management (via AWS SDK)
- [ ] List buckets
- [ ] Browse objects in a bucket
- [ ] Upload objects
- [ ] Download objects
- [ ] Delete objects

### General
- [ ] Multiple EC2 connections (sidebar navigation)
- [ ] Connection management (add/remove servers)
- [ ] Dark mode support
- [ ] Context menu (right-click actions)

## Quick Start

### Prerequisites
- Go 1.22+
- Node.js 20+ (for frontend build)
- SSH key pair configured for your EC2 instances
- AWS credentials (`~/.aws/credentials`) for S3 access

### Install

```bash
# From source
git clone https://github.com/Dongckim/ape.git
cd ape
make build

# Or download binary (coming soon)
# curl -L https://github.com/Dongckim/ape/releases/latest/download/ape-darwin-arm64 -o ape
# chmod +x ape
```

### Usage

```bash
# Start A.P.E (interactive mode)
ape

# Once running, available commands:
#   /add     — connect additional EC2
#   /list    — list active connections
#   /status  — show connection info
#   /h       — help
#   /q       — quit A.P.E
```

## Project Structure

```
ape/
├── main.go                  # CLI entry point
├── go.mod
├── go.sum
├── Makefile                 # Build commands
├── Dockerfile
├── docker-compose.yml
│
├── cmd/                     # CLI commands
│   └── root.go              # Interactive prompt + commands
│
├── internal/                # Private application code
│   ├── api/                 # HTTP handlers
│   │   ├── router.go        # Route registration
│   │   ├── ec2_handler.go   # EC2 file operation endpoints
│   │   ├── s3_handler.go    # S3 operation endpoints
│   │   └── ws_handler.go    # WebSocket for live updates
│   │
│   ├── sftp/                # SFTP client
│   │   ├── client.go        # SSH connection + session pool
│   │   └── operations.go    # File CRUD operations
│   │
│   ├── s3/                  # S3 client
│   │   ├── client.go        # AWS SDK setup
│   │   └── operations.go    # Bucket & object operations
│   │
│   ├── config/              # App configuration
│   │   └── config.go
│   │
│   └── server/              # Web server
│       └── server.go        # HTTP server setup + static file serving
│
├── frontend/                # React + TypeScript
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── src/
│       ├── App.tsx
│       ├── components/
│       │   ├── FileExplorer.tsx    # Finder-style grid/list view
│       │   ├── Sidebar.tsx         # EC2/S3 navigation
│       │   ├── TextEditor.tsx      # Monaco-based text editor
│       │   ├── ContextMenu.tsx     # Right-click menu
│       │   ├── Toolbar.tsx         # Action bar (upload, new folder, etc.)
│       │   └── ConnectionForm.tsx  # Add new EC2 connection
│       ├── api/
│       │   └── client.ts           # API call functions
│       ├── hooks/
│       │   └── useFileSystem.ts    # File operations hook
│       └── types/
│           └── index.ts            # Shared TypeScript types
│
└── docs/
    ├── ROADMAP.md
    ├── CONTRIBUTING.md
    └── API.md
```

## API Endpoints

### EC2 File Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/ec2/files?path=/home/ubuntu` | List directory contents |
| GET | `/api/ec2/file?path=/home/ubuntu/app.py` | Read file content |
| POST | `/api/ec2/upload` | Upload file (multipart) |
| GET | `/api/ec2/download?path=...` | Download file |
| DELETE | `/api/ec2/file?path=...` | Delete file or folder |
| PUT | `/api/ec2/file` | Save edited file content |
| PATCH | `/api/ec2/rename` | Rename file or folder |

### S3 Operations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/s3/buckets` | List all buckets |
| GET | `/api/s3/objects?bucket=...&prefix=...` | List objects |
| POST | `/api/s3/upload?bucket=...` | Upload object |
| GET | `/api/s3/download?bucket=...&key=...` | Download object |
| DELETE | `/api/s3/object?bucket=...&key=...` | Delete object |

### Connection Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/connections` | List active connections |
| POST | `/api/connections` | Add new EC2 connection |
| DELETE | `/api/connections/:id` | Disconnect |

## Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| CLI | Go + interactive prompt | Single binary, cross-platform |
| Backend | Go net/http | Fast, no framework needed |
| SSH/SFTP | golang.org/x/crypto/ssh + pkg/sftp | Standard Go SSH library |
| AWS | aws-sdk-go-v2 | Official AWS SDK for Go |
| Frontend | React 18 + TypeScript | Type-safe, component-based |
| Bundler | Vite | Fast dev server + builds |
| Editor | Monaco Editor | VS Code's editor component |
| Styling | Tailwind CSS | Utility-first, rapid UI dev |

## Roadmap

See [ROADMAP.md](docs/ROADMAP.md) for the detailed development plan.

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

MIT License — see [LICENSE](LICENSE) for details.

---

**A.P.E** — Because managing AWS shouldn't require memorizing commands. 🦍
