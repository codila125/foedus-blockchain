# syntax=docker/dockerfile:1.6
# Multi-stage build for Foedus blockchain
FROM golang:1.25.1-alpine AS builder
WORKDIR /src

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/foedus

FROM alpine:3.22 AS runner
ENV FOEDUS_HOME=/app \
    FOEDUS_BIN=/usr/local/bin/foedus \
    FOEDUS_LOG_DIR=/var/log/foedus \
    SOURCE_NODE_ID=3000 \
    MINER_NODE_ID=3001

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S foedus \
    && adduser -S -G foedus foedus \
    && mkdir -p /app/temp ${FOEDUS_LOG_DIR}

COPY --from=builder /out/foedus /usr/local/bin/foedus
COPY scripts /app/scripts

RUN chmod +x /app/scripts/*.sh \
    && chown -R foedus:foedus /app /var/log/foedus /usr/local/bin/foedus

VOLUME ["/app/temp", "/var/log/foedus"]

EXPOSE 3000 8006 8007
ENTRYPOINT ["/app/scripts/bootstrap-network.sh"]
USER foedus
