# ─── Stage 1: Build ─────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags '-static'" \
    -o /app/bin/todo-app ./cmd/api

# ─── Stage 2: Runtime ───────────────────────────────────────────────────────
FROM scratch

# Copy CA certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/bin/todo-app /todo-app

EXPOSE 8080

ENTRYPOINT ["/todo-app"]
