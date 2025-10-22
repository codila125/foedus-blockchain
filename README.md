# ⚡ Foedus Blockchain

> A modern, production-ready blockchain implementation with smart contracts, milestone-based payments, and REST API support.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Foedus is a complete blockchain implementation featuring Proof-of-Work consensus, UTXO transaction model, smart contracts with milestone tracking, and both CLI and REST API interfaces. Built in Go for performance and simplicity.

---

## 🌟 Features

- ⛏️ **Proof-of-Work Consensus** - SHA-256 based mining with adjustable difficulty
- 💰 **UTXO Transaction Model** - Bitcoin-style unspent transaction outputs
- 📜 **Smart Contracts** - Milestone-based contracts for project delivery and payments
- 🔐 **ECDSA Cryptography** - Secure wallet management with P-256 curve
- 🌳 **Merkle Trees** - Efficient transaction verification
- 🗄️ **PebbleDB Storage** - High-performance persistent storage
- 🌐 **REST API** - Complete HTTP interface for all operations
- 🎨 **CLI Interface** - User-friendly command-line tools
- 🔗 **Multi-Node Support** - Run multiple nodes with isolated data
- 📦 **Protocol Buffers** - Efficient serialization for contracts

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.25+** installed ([download](https://go.dev/dl/))
- **Git** for cloning the repository

### Installation

```bash
# Clone the repository
git clone https://github.com/codila125/foedus-blockchain.git
cd foedus-blockchain

# Download dependencies
go mod download

# Build the project
go build -o blockchain .
```

### First Steps

```bash
# 1. Set your node ID (used for data isolation)
export NODE_ID=3000

# 2. Create your first wallet
./blockchain createwallet
# Output: New address: 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S

# 3. Initialize the blockchain with genesis block
./blockchain createblockchain -address 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S

# 4. Check your balance (you'll have the genesis reward)
./blockchain getbalance -address 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S
# Output: Balance: 100

# 5. View the blockchain
./blockchain printchain
```

### Start API Server

```bash
# Start REST API server (runs on port specified by NODE_ID)
export NODE_ID=3000
./blockchain

# Server starts at http://localhost:3000
# Access API: curl http://localhost:3000/blockchain/printchain
```

---

## 📖 Documentation

Each module has detailed documentation in its respective folder:

| Module | Description | Documentation |
|--------|-------------|---------------|
| 🔐 **[Wallet](wallet/)** | ECDSA key management, address generation | [README](wallet/README.md) |
| ⛓️ **[Blockchain](blockchain/)** | Core blockchain, PoW, transactions, contracts | [README](blockchain/README.md) |
| 🎨 **[CLI](cli/)** | Command-line interface and commands | [README](cli/README.md) |
| 🌐 **[API](api/)** | REST API endpoints and handlers | [README](api/README.md) |
| 🗄️ **[Database](database/)** | PebbleDB wrapper and operations | [README](database/README.md) |
| 🌳 **[Merkle](merkle/)** | Merkle tree implementation | [README](merkle/README.md) |
| 📦 **[Protobuf](protobuf/)** | Protocol Buffer schemas and generated code | [README](protobuf/README.md) |
| 🗂️ **[Temp](temp/)** | Node data storage and management | [README](temp/README.md) |
| 📄 **[Proto](proto/)** | Protocol Buffer definitions | [README](proto/README.md) |

---

## 🎯 Usage Examples

### CLI Operations

```bash
# Create multiple wallets
./blockchain createwallet  # Wallet 1
./blockchain createwallet  # Wallet 2

# List all addresses
./blockchain listaddresses

# Send coins between wallets
./blockchain send -from SENDER_ADDR -to RECEIVER_ADDR -amount 50 -mine

# Check balance
./blockchain getbalance -address YOUR_ADDR

# View entire blockchain
./blockchain printchain

# Reindex UTXO set
./blockchain reindex

# Start mining node
./blockchain startnode -miner YOUR_ADDR
```

### REST API Operations

```bash
# Create wallet
curl http://localhost:3000/blockchain/createwallet

# List addresses
curl http://localhost:3000/blockchain/listaddresses

# Get balance
curl http://localhost:3000/blockchain/getbalance/YOUR_ADDR

# View blockchain
curl http://localhost:3000/blockchain/printchain

# Create smart contract
curl -X POST http://localhost:3000/blockchain/createcontract/CREATOR_ADDR \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Website Development",
    "description": "Full-stack web application",
    "milestones": [{
      "title": "Design Phase",
      "description": "UI/UX design",
      "value": 1000,
      "dueDate": 1735689600
    }],
    "parties": [{
      "address": "CONTRACTOR_ADDR",
      "role": "CONTRACTOR"
    }]
  }'

# Approve contract
curl http://localhost:3000/blockchain/approvecontract/PARTY_ADDR/CONTRACT_ID

# Get contract details
curl http://localhost:3000/blockchain/getcontract/CONTRACT_ID
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         User Layer                          │
│                    CLI / REST API / UI                      │
└────────────────────┬───────────────────────────────────────┘
                     │
┌────────────────────┴───────────────────────────────────────┐
│                    Application Layer                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Wallet  │  │   API    │  │   CLI    │  │ Contracts│  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└────────────────────┬───────────────────────────────────────┘
                     │
┌────────────────────┴───────────────────────────────────────┐
│                     Core Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  Blockchain  │  │ Transactions │  │  Proof-of-   │    │
│  │   & Blocks   │  │   & UTXO     │  │     Work     │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└────────────────────┬───────────────────────────────────────┘
                     │
┌────────────────────┴───────────────────────────────────────┐
│                  Persistence Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   PebbleDB   │  │    Merkle    │  │   Protobuf   │    │
│  │   Database   │  │     Tree     │  │ Serialization│    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 Configuration

### Environment Variables

```bash
# Required: Node identifier (also used as API port)
export NODE_ID=3000

# Optional: Enable debug logging
export DEBUG=true
```

### Blockchain Parameters

- **Difficulty**: 12 (adjustable in `blockchain/proof.go`)
- **Genesis Reward**: 100 coins (configurable in `blockchain/transaction.go`)
- **Database Path**: `temp/blocks_<NODE_ID>/`
- **Wallet Storage**: `wallets.dat`

---

## 🌐 Multi-Node Setup

Run multiple nodes for testing or development:

```bash
# Terminal 1 - Mining Node
export NODE_ID=3000
./blockchain createblockchain -address ADDR1
./blockchain startnode -miner ADDR1

# Terminal 2 - Regular Node
export NODE_ID=3001
./blockchain createblockchain -address ADDR2
./blockchain startnode

# Terminal 3 - API Server
export NODE_ID=3002
./blockchain
# Access at http://localhost:3002
```

---

## 🧪 Testing API with Bruno

A complete API collection is included for testing:

1. Install [Bruno](https://www.usebruno.com/)
2. Open `api/foedus-bruno-api/` collection
3. Update environment variables
4. Execute requests

Included requests:
- Create Wallet
- List Addresses
- Get Balance
- Print Chain
- Create Contract
- Approve Contract
- Get Contract
- Approve Milestone

---

## 📊 Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.25+ | High performance, simplicity |
| **Database** | PebbleDB | Fast key-value storage |
| **Crypto** | ECDSA P-256 | Wallet security |
| **Hashing** | SHA-256 | PoW & Merkle trees |
| **API** | Chi Router | REST endpoints |
| **Serialization** | Protocol Buffers | Efficient data encoding |
| **Encoding** | Base58 | Human-readable addresses |

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📝 Logging

Foedus uses structured, prefixed logging for easy debugging:

- `[BLOCKCHAIN]` - Blockchain operations
- `[TRANSACTION]` - Transaction processing
- `[MINING]` - Mining activities
- `[NETWORK]` - Network communications
- `[UTXO]` - UTXO set operations
- `[SERVER]` - API server events
- `[WALLET]` - Wallet operations

---

## 🗺️ Roadmap

- [x] Basic blockchain with PoW
- [x] UTXO transaction model
- [x] Wallet management
- [x] CLI interface
- [x] REST API
- [x] Smart contracts with milestones
- [ ] P2P network synchronization
- [ ] WebSocket support for real-time updates
- [ ] Web UI dashboard
- [ ] Enhanced contract scripting
- [ ] Performance optimizations
- [ ] Comprehensive test suite

---

## ⚠️ Security Notice

**This is an educational and experimental blockchain implementation.**

- ✅ Great for learning blockchain concepts
- ✅ Suitable for development and testing
- ✅ Good foundation for custom blockchain projects
- ❌ Not audited for production financial systems
- ❌ Do not use for real monetary transactions without security review

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Bitcoin whitepaper for UTXO model inspiration
- Go community for excellent tools and libraries
- PebbleDB team for high-performance storage
- Protocol Buffers for efficient serialization

---

## 📞 Support

- 📖 **Documentation**: Each module has detailed README
- 🐛 **Issues**: [GitHub Issues](https://github.com/codila125/foedus-blockchain/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/codila125/foedus-blockchain/discussions)

---

<div align="center">

**Built with ❤️ using Go**

⭐ Star this repo if you find it helpful!

</div>