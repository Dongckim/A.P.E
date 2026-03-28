.PHONY: build build-frontend dev dev-frontend clean test release install

# Build frontend then Go binary (single file with embedded UI)
build: build-frontend
	go build -o bin/ape main.go
	@echo "\n  ✓ Built bin/ape (frontend embedded)"

# Build frontend assets
build-frontend:
	cd frontend && npm install && npm run build

# Run Go backend in dev mode (uses embedded frontend)
dev: build-frontend
	go run main.go

# Run frontend dev server with hot reload (separate terminal)
dev-frontend:
	cd frontend && npm run dev

# Run all tests
test:
	go vet ./...
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/

# Cross-compile for release
release: build-frontend
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/ape-darwin-arm64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/ape-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -ldflags="-s -w" -o bin/ape-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux  GOARCH=arm64 go build -ldflags="-s -w" -o bin/ape-linux-arm64 main.go
	@echo "\n  ✓ Binaries in bin/"

# Install locally
install: build
	cp bin/ape /usr/local/bin/ape
	@echo "  ✓ Installed to /usr/local/bin/ape"
