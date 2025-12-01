# Foedus Blockchain v0.4.0 — Release Notes

<p align="center">
  <img src="https://img.shields.io/badge/release-v0.4.0-blue.svg" alt="Release v0.4.0">
  <img src="https://img.shields.io/badge/status-production--ready-brightgreen.svg" alt="Production Ready">
</p>

---

## 🎉 Production Ready Release

**Foedus v0.4.0** is our first production-ready release, delivering immutable contract infrastructure for government bodies and enterprises. This version provides a stable, secure platform for managing legally binding agreements with full audit trails.

---

## ✨ What's New

### Production-Grade Error Handling
All core functions now return proper errors instead of terminating execution. This enables graceful degradation and better integration with upstream services.

- `NewBlockChain()` and `ContinueBlockChain()` return typed errors
- Custom error types: `ErrBlockchainExists`, `ErrBlockchainNotFound`
- Consistent error wrapping throughout the codebase

### Health Check System
Built-in health monitoring for container orchestration:

```bash
# CLI health check
./foedus --health

# HTTP endpoint
curl http://localhost:3000/health
# {"status":"ok"}
```

- Docker `HEALTHCHECK` integration
- Kubernetes-ready liveness probes
- Graceful shutdown with configurable timeout

### Contract Cancellation
Securely cancel active contracts and recover undistributed funds:

```bash
POST /blockchain/cancelcontract
{
  "contract_id": "...",
  "creator": "creator_address"
}
```

- Authorization verification for contract parties
- Automatic fund recovery to originating wallet
- Immutable cancellation records on-chain

### Enhanced Stability
- Graceful shutdown handling with `SIGINT`/`SIGTERM`
- Concurrent resource cleanup with timeout protection
- Improved database connection management

---

## 🔧 Improvements

- **API**: Health endpoint added at `/health`
- **CLI**: `--health` flag for container health checks
- **Docker**: Multi-stage build with non-root user
- **Network**: Better peer connection error handling
- **Logging**: Structured log output with context tags

---

## 📊 Test Coverage

| Module | Coverage |
|--------|----------|
| merkle | 100% |
| wallet | 89% |
| database | 86% |
| blockchain | 28% |
| cli | 25% |

---

## 🚀 Upgrade Guide

### From v0.3.x

1. Pull the latest image:
   ```bash
   docker pull codila125/foedus-blockchain:0.4.0
   ```

2. Stop existing container:
   ```bash
   docker stop foedus
   ```

3. Start with new version:
   ```bash
   docker run -d \
     --name foedus \
     -p 3000:3000 \
     -v foedus-data:/app/data \
     codila125/foedus-blockchain:0.4.0
   ```

**Note**: Blockchain data is backward compatible. No migration required.

---

## 📦 Installation

### Docker (Recommended)
```bash
docker pull codila125/foedus-blockchain:0.4.0
docker run -d -p 3000:3000 --name foedus codila125/foedus-blockchain:0.4.0
```

### From Source
```bash
git clone https://github.com/codila125/foedus-blockchain.git
cd foedus-blockchain
git checkout v0.4.0
go build -o foedus
```

---

## 🔗 Links

- **Docker Hub**: [codila125/foedus-blockchain](https://hub.docker.com/r/codila125/foedus-blockchain)
- **GitHub**: [codila125/foedus-blockchain](https://github.com/codila125/foedus-blockchain)
- **Documentation**: [README.md](README.md)

---

<p align="center">
  <strong>Foedus v0.4.0</strong> — Immutable Contracts for Government & Enterprise
</p>
