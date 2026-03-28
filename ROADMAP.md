# A.P.E Roadmap

## Overview

2-week MVP development plan for A.P.E (AWS Platform Explorer).

**Goal**: A working CLI tool that connects to EC2 via SSH interactively and serves a Finder-like web GUI at localhost:9000 for file management + S3 browsing.

---

## Phase 1 — Go Setup + EC2 SFTP (Day 1–3)

### Day 1: Go project init + interactive CLI
- [ ] Initialize Go module (`go mod init github.com/USERNAME/ape`)
- [ ] Set up project folder structure
- [ ] Implement interactive CLI prompt
  - Display ASCII art banner (ape face + APE block text)
  - Prompt for SSH key path (default: `~/.ssh/id_rsa`)
  - Prompt for EC2 host
  - Prompt for username (default: `ubuntu`)
  - Prompt for SSH port (default: `22`)
  - On invalid input → show error + re-prompt that field
- [ ] Implement command loop after connection (`ape ▸` prompt)
  - `/add` — connect additional EC2
  - `/list` — list active connections
  - `/status` — show connection info
  - `/h` — help
  - `/q` — quit and close all connections
- [ ] Start basic HTTP server on localhost:9000
- [ ] Serve a "Hello from A.P.E" page to verify it works
- **Milestone**: `ape` starts interactive prompt, connects, and opens localhost:9000

### Day 2: SSH/SFTP connection
- [ ] Implement SSH client using `golang.org/x/crypto/ssh`
  - Read key from `~/.ssh/id_rsa` or `-i` flag
  - Support passphrase-protected keys
  - Handle `known_hosts` verification
- [ ] Establish SFTP session over SSH using `github.com/pkg/sftp`
- [ ] Implement `ListDirectory(path string)` — returns file/folder list with metadata
  - Name, size, permissions, modified date, is_dir
- [ ] Create GET `/api/ec2/files?path=...` endpoint
- **Milestone**: curl `localhost:3000/api/ec2/files?path=/home/ubuntu` returns JSON file list

### Day 3: SFTP file operations
- [ ] Implement `ReadFile(path string)` — returns file content
- [ ] Implement `WriteFile(path string, content []byte)` — save/overwrite file
- [ ] Implement `UploadFile(path string, reader io.Reader)` — upload from multipart
- [ ] Implement `DownloadFile(path string)` — stream file to HTTP response
- [ ] Implement `DeleteFile(path string)` — delete file or directory (recursive)
- [ ] Implement `RenameFile(oldPath, newPath string)` — rename/move
- [ ] Wire all operations to REST API endpoints
- **Milestone**: All EC2 file CRUD operations work via curl/Postman

---

## Phase 2 — S3 Integration (Day 4–5)

### Day 4: S3 client setup
- [ ] Add `aws-sdk-go-v2` dependency
- [ ] Implement S3 client using local `~/.aws/credentials`
- [ ] Implement `ListBuckets()` — returns all S3 buckets
- [ ] Implement `ListObjects(bucket, prefix string)` — returns objects with folder-like navigation
- [ ] Create GET `/api/s3/buckets` and GET `/api/s3/objects` endpoints
- **Milestone**: curl returns list of S3 buckets and objects

### Day 5: S3 CRUD operations
- [ ] Implement `UploadObject(bucket, key string, reader io.Reader)`
- [ ] Implement `DownloadObject(bucket, key string)` — stream to response
- [ ] Implement `DeleteObject(bucket, key string)`
- [ ] Implement presigned URL generation for large downloads
- [ ] Wire to REST API endpoints
- [ ] Add connection management endpoints (list/add/remove)
- **Milestone**: Full backend API complete (EC2 + S3), all testable via curl

---

## Phase 3 — React Frontend (Day 6–10)

### Day 6: React project init + layout
- [ ] Initialize Vite + React + TypeScript project in `frontend/`
- [ ] Set up Tailwind CSS
- [ ] Create main layout: sidebar + content area
- [ ] Build Sidebar component
  - EC2 connections list
  - S3 buckets list
  - "Add connection" button
- [ ] Set up API client (`frontend/src/api/client.ts`)
- [ ] Configure Vite proxy to Go backend (`localhost:9000/api`)
- **Milestone**: App shell renders with sidebar showing mock data

### Day 7: File explorer component
- [ ] Build FileExplorer component
  - Grid view (icon + filename, like Finder)
  - List view (table with name, size, date, permissions)
  - Toggle between grid/list view
- [ ] Implement breadcrumb navigation (path bar)
- [ ] Double-click folder → navigate into it
- [ ] Single-click file → select it
- [ ] Multi-select with Shift/Cmd+click
- [ ] Connect to real API data
- **Milestone**: Browse EC2 file system through the web UI

### Day 8: File operations UI
- [ ] Drag & drop file upload with progress bar
- [ ] Context menu (right-click)
  - Open / Download / Rename / Delete
- [ ] "New Folder" button in toolbar
- [ ] Download: click triggers browser download
- [ ] Delete: confirmation dialog → API call
- [ ] Rename: inline editing
- **Milestone**: Upload, download, delete, rename all work in UI

### Day 9: Text editor + S3 browser
- [ ] Integrate Monaco Editor for text file editing
  - Click text file → opens in editor panel
  - Save button → PUT to API
  - Syntax highlighting based on file extension
- [ ] Build S3 browser view
  - Reuse FileExplorer component
  - Bucket list → object list navigation
  - Upload/download/delete for S3 objects
- **Milestone**: Edit files in browser, S3 browsing works

### Day 10: Polish + connection management
- [ ] Connection settings modal
  - Add new EC2 connection (host, user, key path)
  - Remove connection
  - Show connection status (connected/disconnected)
- [ ] Error handling
  - Connection failed → clear error message
  - Permission denied → show which file
  - Network timeout → retry option
- [ ] Loading states (skeletons, spinners)
- [ ] Empty states (no files, no connections)
- [ ] Keyboard shortcuts (Cmd+C copy path, Delete, Cmd+N new folder)
- **Milestone**: Working MVP with polished UX

---

## Phase 4 — Ship It (Day 11–14)

### Day 11: Build pipeline
- [ ] Embed frontend build into Go binary (`embed` package)
- [ ] Create Makefile with targets: `build`, `dev`, `test`, `clean`
- [ ] Cross-compile for macOS (arm64, amd64) and Linux
- [ ] Create Dockerfile for containerized usage
- [ ] Test on fresh machine (clean install experience)

### Day 12: Documentation
- [ ] Polish README.md with screenshots/GIF
- [ ] Write CONTRIBUTING.md
- [ ] Write API.md with all endpoints documented
- [ ] Add inline code comments for complex logic
- [ ] Create `examples/` folder with usage examples

### Day 13: GitHub setup
- [ ] Create GitHub repository
- [ ] Add MIT LICENSE
- [ ] Set up GitHub Actions CI (build + test)
- [ ] Create GitHub Release with binaries
- [ ] Add issue templates (bug report, feature request)
- [ ] Add `.goreleaser.yml` for automated releases

### Day 14: Launch
- [ ] Record demo GIF for README
- [ ] Write launch post (Reddit r/golang, r/aws, Hacker News)
- [ ] Final testing on macOS + Linux
- [ ] Tag v0.1.0 release
- **Milestone**: Public GitHub repo live with downloadable binary

---

## Post-MVP (Future)

- [ ] Go rewrite optimization (connection pooling, caching)
- [ ] File preview (images, PDFs, markdown)
- [ ] Terminal panel (embedded SSH terminal)
- [ ] Multi-tab interface
- [ ] File search across EC2
- [ ] Transfer files between EC2 and S3
- [ ] AWS IAM role support
- [ ] Config file (`~/.ape/config.yaml`) for saved connections
- [ ] Auto-update mechanism
