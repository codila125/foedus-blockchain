# Multi-stage build for Foedus blockchain
FROM golang:1.25.1-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/foedus

FROM alpine:3.20 AS runner
ENV FOEDUS_HOME=/app \
    FOEDUS_BIN=/usr/local/bin/foedus \
    FOEDUS_LOG_DIR=/var/log/foedus \
    SOURCE_NODE_ID=3000 \
    MINER_NODE_ID=3001

WORKDIR /app

RUN apk add --no-cache bash ca-certificates coreutils grep

COPY --from=builder /out/foedus /usr/local/bin/foedus
COPY scripts /app/scripts

RUN chmod +x /app/scripts/*.sh \
    && mkdir -p /app/temp ${FOEDUS_LOG_DIR}

EXPOSE 3000 8006 8007
ENTRYPOINT ["/app/scripts/bootstrap-network.sh"]
