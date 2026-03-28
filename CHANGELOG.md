# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-03-28

### Added
- Saved connections support via `~/.ape/config.yaml`
- Show saved connections at startup for quick selection
- Prompt to save new connections after successful SSH connect

## [0.2.0] - 2026-03-28

### Added
- Server dashboard with live monitoring
- Real-time server metrics display

### Fixed
- Null safety on all dashboard components (prevent white screen crash)
- Dashboard routes moved to `/api/dashboard/*` to avoid path conflicts
- API request routing before SPA fallback
- Non-JSON response handling

## [0.1.0] - 2026-03-28

### Added
- Interactive CLI with SSH/SFTP connection prompts
- EC2 file browser (list, read, write, upload, download, delete, rename)
- S3 bucket browser (list buckets, objects, upload, download, delete)
- React frontend with Finder-style file explorer
- Monaco editor with syntax highlighting
- Drag & drop file upload with progress bar
- Right-click context menu
- Multi-select support (Shift/Cmd+click)
- Keyboard shortcuts (Cmd+N, Delete, Cmd+C)
- Multiple EC2 connection support (`/add`, `/list`, `/status`)
- Single binary with embedded frontend
- Dockerfile for containerized deployment
- CI/CD pipeline (lint, test, build, release)

[0.3.0]: https://github.com/Dongckim/A.P.E/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Dongckim/A.P.E/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Dongckim/A.P.E/releases/tag/v0.1.0
