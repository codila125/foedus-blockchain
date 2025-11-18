# Foedus Blockchain

![Cover](images/cover-template.png)

[![Go Version](https://img.shields.io/badge/go-1.25.1+-blue.svg)](https://golang.org) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![PebbleDB](https://img.shields.io/badge/PebbleDB-green.svg)](https://github.com/cockroachdb/pebble) [![libp2p](https://img.shields.io/badge/libp2p-purple.svg)](https://github.com/libp2p/go-libp2p) [![Chi](https://img.shields.io/badge/Chi-red.svg)](https://github.com/go-chi/chi) [![Protocol Buffers](https://img.shields.io/badge/Protocol%20Buffers-orange.svg)](https://protobuf.dev/)

A educational purpose blockchain implementation in Go featuring UTXO transaction model, Proof-of-Work consensus, milestone-based smart contracts, and peer-to-peer networking.

## Overview

Foedus is a decentralized blockchain platform designed for secure, trustless project collaboration through smart contracts. It combines proven cryptographic primitives (Ed25519 signatures, SHA-256 hashing) with modern distributed systems technologies (libp2p, Protocol Buffers) to deliver a robust foundation for contract-based agreements.

**Key Features:**
- **UTXO Model**: Bitcoin-style unspent transaction output architecture for parallel transaction processing
- **Proof-of-Work**: SHA-256 mining with configurable difficulty (default: 12)
- **Smart Contracts**: Milestone-based project agreements with multi-party signature validation
- **P2P Network**: libp2p-based decentralized communication with NAT traversal
- **Dual Interface**: Command-line tools and RESTful HTTP API
- **Persistent Storage**: PebbleDB LSM-tree database for blockchain state

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Application Layer                   │
│  ┌──────────────┐              ┌──────────────────────┐ │
│  │   CLI Mode   │              │   API Server Mode    │ │
│  │  (commands)  │              │  (HTTP endpoints)    │ │
│  └──────────────┘              └──────────────────────┘ │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────┐
│                   Blockchain Core                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │
│  │   Wallet    │  │ Blockchain  │  │  Smart Contract │  │
│  │  (Ed25519)  │  │ (UTXO/PoW)  │  │   (Milestones)  │  │
│  └─────────────┘  └─────────────┘  └─────────────────┘  │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────┐
│               Infrastructure Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │
│  │   Network   │  │  Database   │  │  Merkle Tree    │  │
│  │  (libp2p)   │  │  (PebbleDB) │  │  (SHA-256)      │  │
│  └─────────────┘  └─────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites
- Go 1.25.1 or higher
- 200MB disk space for blockchain data

### Installation

```bash
# Clone the repository
git clone https://github.com/codila125/foedus-blockchain.git
cd foedus-blockchain

# Install dependencies
go mod download

# Build the binary
go build -o foedus
```

### Basic Usage

**Create a wallet:**
```bash
./foedus createwallet
# Output: Your new address: 1A2B3C4D5E6F7G8H9I0J...
```

**Start API server:**
```bash
export NODE_ID=3000
./foedus
# API listening on http://localhost:3000
```

**Mine the genesis block:**
```bash
./foedus createblockchain -address <your-address>
```

**Send tokens:**
```bash
./foedus send -from <sender> -to <receiver> -amount 10
```

**Check balance:**
```bash
./foedus getbalance -address <your-address>
```

## Dockerized Local Network

Spin up a ready-to-mine playground with the included multi-stage `Dockerfile` and bootstrap script (`scripts/bootstrap-network.sh`). The container automates the full workflow:

1. Sets `NODE_ID=3000`, creates a wallet, and initializes the blockchain with that address.
2. Starts the API/source server (port `3000`, libp2p source on `8006`) and captures its multiaddress.
3. Sets `NODE_ID=3001` and launches a miner that auto-connects to the captured multiaddress (libp2p port `8007`).

### Build the image
```bash
docker build -t foedus-blockchain .
```

### Run the paired source + miner stack
```bash
docker run --rm -p 3000:3000 -p 8006:8006 -p 8007:8007 foedus-blockchain
```
> **Note:** The image declares `VOLUME /app/temp` and `VOLUME /var/log/foedus`, so Docker automatically creates anonymous volumes the first time the container runs. Those volumes stick around only while the container exists. If you want persistence without naming the volumes yourself, simply drop `--rm` and restart the same container ID:
> ```bash
> docker run -d --name foedus -p 3000:3000 -p 8006:8006 -p 8007:8007 foedus-blockchain
> # later
> docker stop foedus
> docker start foedus
> ```
> Removing the container (`docker rm foedus`) also deletes its anonymous volumes, so use named volumes if you need to recreate containers frequently.

### Run with explicit persistent volumes
```bash
docker volume create foedus-data
docker volume create foedus-logs
docker run --rm \
  -p 3000:3000 -p 8006:8006 -p 8007:8007 \
  -v foedus-data:/app/temp \
  -v foedus-logs:/var/log/foedus \
  foedus-blockchain
```

Container logs stream from `/var/log/foedus/source.log` and `/var/log/foedus/miner.log`, so `docker logs -f <container>` shows both servers plus bootstrap progress. The API remains available on `http://localhost:3000`, while libp2p peers can dial the exposed 8006/8007 ports.

**Customization tips:**
- Override defaults with env vars, e.g. `-e SOURCE_NODE_ID=4000 -e MINER_NODE_ID=4001 -e SOURCE_WAIT_SECS=120`.
- The entrypoint script can be reused locally: `scripts/bootstrap-network.sh` assumes it runs from the repo root (or `/app` in the container).

## Module Documentation

The codebase is organized into specialized modules with dedicated READMEs:

- **[api/](api/)** - RESTful HTTP server with Chi router (8 endpoints)
- **[blockchain/](blockchain/)** - Core UTXO/PoW implementation with smart contracts
- **[cli/](cli/)** - Command-line interface (8 commands)
- **[database/](database/)** - PebbleDB wrapper with batch operations
- **[merkle/](merkle/)** - Binary hash tree for transaction verification
- **[network/](network/)** - libp2p P2P layer with source/miner node types
- **[wallet/](wallet/)** - Ed25519 key management with Base58 addresses

## Multi-Node Deployment

Run multiple nodes for distributed consensus:

```bash
# Terminal 1: Source node
export NODE_ID=3000
./foedus

# Terminal 2: Miner node (auto-syncs with source)
export NODE_ID=3001
./foedus
```

Nodes automatically discover and sync blockchain state using libp2p protocol `/foedus/1.0.0`.

## Smart Contracts

Create milestone-based project contracts:

```bash
# API endpoint
POST /api/createcontract
{
  "creator": "1A2B3C...",
  "participant": "1X9Y8Z...",
  "amount": 1000,
  "milestones": [
    {"description": "Design phase", "amount": 300},
    {"description": "Development", "amount": 500},
    {"description": "Testing", "amount": 200}
  ]
}
```

Contract workflow:
1. Creator proposes contract → Pending
2. Participant approves → Active
3. Milestones completed sequentially with multi-signature approval
4. Funds released incrementally

## API Reference

**Core Endpoints:**
- `POST /api/createwallet` - Generate new Ed25519 wallet
- `GET /api/getbalance/:address` - Query UTXO balance
- `POST /api/send` - Submit signed transaction
- `POST /api/createcontract` - Propose new contract
- `POST /api/approvecontract` - Accept contract terms
- `POST /api/approvemilestone` - Complete milestone
- `GET /api/getcontract/:contractID` - Query contract state
- `GET /api/printchain` - Export full blockchain

See [api/README.md](api/README.md) for detailed specifications.

## Development

**Project Structure:**
```
foedus-blockchain/
├── main.go          # Entry point (CLI or API mode)
├── api/             # HTTP server
├── blockchain/      # Core consensus and state
├── cli/             # Command-line interface
├── database/        # PebbleDB wrapper
├── merkle/          # Merkle tree verification
├── network/         # P2P communication
├── wallet/          # Key management
└── proto/           # Protocol Buffer schemas
```

**Run Tests:**
```bash
go test ./...
```

## Technical Specifications

- **Consensus**: Proof-of-Work (SHA-256, difficulty 12)
- **Cryptography**: Ed25519 signatures, SHA-256 hashing
- **Storage**: PebbleDB LSM-tree with atomic batch writes
- **Networking**: libp2p with TCP transport (ports 8006/8007)
- **Serialization**: Protocol Buffers v3
- **Address Format**: Base58Check with 4-byte checksum
