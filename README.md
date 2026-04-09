# A.P.E

[![CI](https://github.com/Dongckim/A.P.E/actions/workflows/ci.yml/badge.svg)](https://github.com/Dongckim/A.P.E/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Dongckim/A.P.E)](https://github.com/Dongckim/A.P.E/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white)](https://react.dev)

**AWS Platform Explorer** — browse EC2 files, S3 buckets, and RDS PostgreSQL from your browser. One binary, no AWS Console.

> Tired of `ssh -L` in one terminal, `psql` in another, hunting endpoints in the AWS Console, and adding your laptop's IP to a Security Group every time you switch wifi? A.P.E gives you a Finder-style web GUI that runs locally and just works.

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

         AWS Platform Explorer
                v0.5.0

  ⏺ Connected to ubuntu@54.123.45.67
  ⏺ PostgreSQL connected (tunneled via 54.123.45.67)
  ⏺ Web UI ready at http://localhost:9000
```

One binary. SSH into your EC2. Get a Finder-like GUI at `localhost:9000` for files, S3, and your RDS PostgreSQL — even when RDS lives in a private subnet.

## 🎥 Demo (v0.5.0 — RDS PostgreSQL)

https://github.com/Dongckim/A.P.E/releases/download/v0.5.0/ape-v0.5.0-rds-demo.mp4

> *Click through every database, drill into any schema, and auto-generate the entire ERD with one click.*

## What's new in v0.5.0

A.P.E now speaks RDS PostgreSQL — and it does it without making you fight Security Groups, ssh tunnels, or third-party clients:

- 🔌 **Interactive RDS setup at startup** — host, user, db, hidden password input. No env vars to babysit.
- 🚇 **In-process SSH tunnel** — if your RDS is in a private subnet, A.P.E opens an `ssh -L`-equivalent *inside the same process* through the existing EC2 SSH session. From your laptop, RDS is just `localhost`. **You don't even need a Security Group inbound rule** — the tunnel uses EC2's internal VPC route.
- 🗂 **Multi-database overview** — every database on the instance with size. Click to switch; a connection factory caches per-DB clients so credentials are entered once.
- 🔍 **3-level drill-down** — Database → Schema → Table → Columns + sample rows. Primary keys marked with 🔑, foreign keys flagged.
- 🌿 **Auto-generated ERD per schema** — Mermaid renders the entity-relationship diagram from `pg_catalog`. One click and you see how every table connects.
- 💾 **Saved connections** — `~/.ape/config.yaml` stores host/port/user/db/sslmode/tunnel preference. Passwords are *never* persisted, always re-prompted at startup.

## Quick Start

### Homebrew (macOS)

```bash
brew tap Dongckim/tap
brew install ape
ape
```

### Manual Download

```bash
# macOS Apple Silicon
curl -sL https://github.com/Dongckim/A.P.E/releases/latest/download/ape-darwin-arm64.tar.gz | tar xz
./ape
```

### Build from Source

```bash
git clone https://github.com/Dongckim/A.P.E.git && cd A.P.E
make build
./bin/ape
```

## What it does

```
┌──────────────┐         ┌─────────────────┐
│   Browser    │◄───────►│   A.P.E (Go)    │
│    :9000     │         │    localhost    │
└──────────────┘         └─┬─────┬─────┬───┘
                       SSH/SFTP │   AWS SDK
                          │     │     │
                     ┌────▼───┐ │  ┌──▼───┐
                     │  EC2   │ │  │  S3  │
                     └────┬───┘ │  └──────┘
                          │     │
                          ▼     │
                    (SSH tunnel forwards
                     PostgreSQL through
                     the EC2 bastion —
                     RDS never sees your
                     laptop's IP)
                          │     │
                     ┌────▼─────▼───┐
                     │  RDS         │
                     │  PostgreSQL  │
                     │  (private)   │
                     └──────────────┘
```

- **EC2** — browse, upload, download, edit, rename, delete files via SFTP
- **S3** — list buckets, navigate objects, upload/download/delete
- **RDS PostgreSQL** — overview, multi-DB switch, schema/table drill-down, auto-generated ERD
- **Editor** — built-in Monaco editor with syntax highlighting
- **Security** — SSH keys never leave your machine. RDS passwords never written to disk. Server binds to localhost only.

## Features

### File explorer
- Finder-style grid + list view
- Drag & drop upload with progress bar
- Right-click context menu
- Monaco text editor (`Cmd+S` to save)
- Multi-select (Shift / Cmd+click)
- Keyboard shortcuts (`Cmd+N`, `Delete`, `Cmd+C`, `?`)

### S3
- Bucket browser with prefix navigation
- Upload / download / delete
- Presigned URLs

### RDS PostgreSQL (new in v0.5.0)
- Interactive connection setup (no env vars required)
- In-process SSH tunnel through EC2 bastion (no `ssh -L` needed, no SG rule required)
- Database list with size, click-to-switch
- Schema drill-down with row count + size
- Table detail: columns (PK/FK marks), nullable, defaults, sample rows
- Auto-generated ERD per schema (Mermaid)
- Copy ERD source to clipboard for Notion / dbdiagram.io / docs
- Saved connections in `~/.ape/config.yaml` (passwords never stored)

### Other
- Multiple EC2 connections
- Single ~16MB binary (frontend embedded)
- Cross-platform (macOS, Linux, x86_64 + arm64)

## CLI Commands

```
ape ▸ /add      connect additional EC2
ape ▸ /list     list active connections
ape ▸ /status   show connection info
ape ▸ /h        show help
ape ▸ /q        quit

ape --version   print version
ape --help      print usage
```

## Tech

**Backend**: Go · `net/http` · `golang.org/x/crypto/ssh` · `github.com/pkg/sftp` · `aws-sdk-go-v2` · `jackc/pgx/v5`

**Frontend**: React 19 · TypeScript · Tailwind · Vite · Monaco Editor · Mermaid

**Distribution**: GoReleaser · Homebrew tap

## Why A.P.E for RDS

Most production RDS instances live in private subnets behind an EC2 bastion. The usual workflow:

1. Log into AWS Console (auth expired *again*)
2. Add your laptop's IP to the RDS Security Group (it changed since coffee ☕)
3. Copy the endpoint
4. Open pgAdmin / DBeaver / TablePlus, reconfigure for the 12th time
5. Realize it's still private → `ssh -L` in another terminal
6. Re-type the password you swore you'd remember

A.P.E collapses all of that into:

```bash
brew install ape && ape
```

The trick: A.P.E reuses your existing EC2 SSH session as a transport for PostgreSQL. The connection from `pgx` flows out to a localhost TCP listener that A.P.E binds, gets relayed through the SSH session to the RDS endpoint, and the SSL handshake happens end-to-end with the real RDS server. From your laptop's perspective it's `127.0.0.1`. From RDS's perspective it's the EC2's internal IP. **Neither end ever needs to know your laptop's public IP.**

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and PRs welcome.

## License

MIT
