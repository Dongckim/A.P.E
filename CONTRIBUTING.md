# Contributing to A.P.E

Thanks for your interest in contributing to A.P.E! Here's how to get started.

## Development Setup

### Prerequisites
- Go 1.22+
- Node.js 20+
- An AWS EC2 instance with SSH access (for testing)

### Clone and build

```bash
git clone https://github.com/Dongckim/ape.git
cd ape
go mod download

cd frontend
npm install
cd ..
```

### Run in development mode

You'll need two terminals:

```bash
# Terminal 1: Go backend
go run main.go

# Terminal 2: React frontend (with hot reload)
cd frontend && npm run dev
```

The frontend dev server proxies API calls to the Go backend at port 9000.

## Making Changes

### Branch naming
- `feat/description` — new features
- `fix/description` — bug fixes
- `docs/description` — documentation
- `refactor/description` — code refactoring

### Commit messages
Use conventional commits:
```
feat: add file rename operation
fix: handle SSH timeout gracefully
docs: update API endpoint documentation
refactor: extract SFTP connection pool
```

### Pull request process
1. Fork the repo
2. Create your branch (`git checkout -b feat/my-feature`)
3. Make your changes
4. Run tests (`go test ./...` and `cd frontend && npm test`)
5. Commit and push
6. Open a Pull Request with a clear description

## Code Style

### Go
- Run `gofmt` before committing
- Run `go vet ./...` for static analysis
- Use meaningful variable names
- Add comments for exported functions

### TypeScript
- Run `npm run lint` before committing
- Use TypeScript strict mode
- Prefer named exports over default exports

## Reporting Issues

When reporting a bug, please include:
- OS and version
- Go version (`go version`)
- Node.js version (`node --version`)
- Steps to reproduce
- Expected vs actual behavior
- Error logs (if any)

## Feature Requests

Open an issue with the `enhancement` label. Describe:
- What problem does this solve?
- How should it work from a user's perspective?
- Any technical considerations?

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
