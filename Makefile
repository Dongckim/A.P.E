.PHONY: build dev dev-frontend clean test

# Build frontend and embed into Go binary
build: build-frontend
	go build -o bin/ape main.go

# Build frontend assets
build-frontend:
	cd frontend && npm run build

# Run Go backend in dev mode
dev:
	go run main.go connect $(ARGS)

# Run frontend dev server (separate terminal)
dev-frontend:
	cd frontend && npm run dev

# Run all tests
test:
	go test ./...
	cd frontend && npm test

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/dist/

# Cross-compile for release
release:
	GOOS=darwin GOARCH=arm64 go build -o bin/ape-darwin-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/ape-darwin-amd64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/ape-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o bin/ape-linux-arm64 main.go

# Install locally
install: build
	cp bin/ape /usr/local/bin/ape
