# --- Stage 1: Build frontend ---
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- Stage 2: Build Go binary ---
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ape main.go

# --- Stage 3: Minimal runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/ape /usr/local/bin/ape
ENTRYPOINT ["ape"]
