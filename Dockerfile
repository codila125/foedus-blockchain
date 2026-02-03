# syntax=docker/dockerfile:1.6
#
# Multi-stage Dockerfile for Foedus Blockchain
#

# ============================================================================
# Stage 1: Builder
# ============================================================================
FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine3.22 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# Build environment variables
ENV CGO_ENABLED=0

# Cache dependencies layer
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Build application
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags="-s -w -X main.Version=${VERSION:-dev}" \
    -o /out/foedus

# ============================================================================
# Stage 2: Runtime
# ============================================================================
FROM --platform=$TARGETPLATFORM alpine:3.22

LABEL maintainer="codila125" \
    description="Foedus Blockchain - Immutable Contract Infrastructure for Enterprise" \
    version="0.4.0"

# Environment variables
ENV FOEDUS_HOME=/app \
    FOEDUS_BIN=/usr/local/bin/foedus \
    FOEDUS_LOG_DIR=/var/log/foedus \
    FOEDUS_DATA_DIR=/app/data \
    SOURCE_NODE_ID=3008 \
    MINER_NODE_ID=3011

WORKDIR /app

# Install runtime dependencies and create app user
RUN apk add --no-cache \
    ca-certificates \
    tini \
    && addgroup -S foedus \
    && adduser -S -G foedus foedus \
    && mkdir -p /app/temp "${FOEDUS_LOG_DIR}" "${FOEDUS_DATA_DIR}" \
    && chown -R foedus:foedus /app /var/log/foedus

# Copy binary from builder
COPY --from=builder --chown=foedus:foedus /out/foedus /usr/local/bin/foedus

# Copy scripts and set permissions
COPY --chown=foedus:foedus scripts /app/scripts
RUN chmod +x /app/scripts/*.sh

# Health check for container orchestration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD foedus --health || exit 1

# Data volumes - only persistent data and logs require volumes
VOLUME ["/app/data", "/var/log/foedus"]

# Security: Run as non-root user
USER foedus

# Expose ports
EXPOSE 3008 3009 3010 3011

# Use tini as PID 1 for proper signal handling
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/scripts/bootstrap-network.sh"]
