# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -extldflags '-static'" \
    -o /build/api ./cmd/api

# ── Production stage ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS production

# Copy timezone data and TLS certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/api /api

# The distroless image runs as uid 65532 (nonroot) by default
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/api"]
