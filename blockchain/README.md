# ⛓️ Blockchain Module

> The core engine of Foedus - implementing Proof-of-Work consensus, UTXO transactions, and milestone-based smart contracts.

This module is the heart of the blockchain, combining Bitcoin-inspired UTXO model with modern smart contract capabilities for real-world project delivery and payments.

---

## ✨ Features

- ⛏️ **Proof-of-Work Mining** - SHA-256 based consensus with adjustable difficulty
- 💰 **UTXO Model** - Bitcoin-style unspent transaction outputs
- 📜 **Smart Contracts** - Milestone-based payments for projects
- 🔗 **Chain Validation** - Cryptographic integrity verification
- 🗄️ **PebbleDB Storage** - High-performance persistent storage
- 🌳 **Merkle Trees** - Efficient transaction verification
- 🔐 **Digital Signatures** - Secure transaction authorization

## 🏗️ Architecture

### Core Components

- **`blockchain.go`** - Blockchain structure and chain management
- **`block.go`** - Block structure and creation
- **`transaction.go`** - Transaction logic and UTXO handling
- **`proof.go`** - Proof-of-Work mining algorithm
- **`contract.go`** - Smart contract and milestone structures
- **`utxo.go`** - UTXO set management
- **`iterator.go`** - Blockchain traversal
- **`serialize.go`** - Data serialization/deserialization

### Key Structures

```go
type BlockChain struct {
    LastHash []byte           // Hash of the last block
    Database *database.PebbleDB  // Persistent storage
}

type Block struct {
    Hash         []byte
    Transactions []*Transaction
    PrevHash     []byte
    Nonce        int
    Height       int
}

type Transaction struct {
    ID      []byte
    Inputs  []TxInput
    Outputs []TxOutput
}
```

## 🚀 Usage

### CLI Commands

#### Create a Blockchain
```bash
# Initialize a new blockchain with genesis block
# Sends initial reward to specified address
export NODE_ID=3000
./blockchain createblockchain -address YOUR_ADDRESS

# Example output:
# [BLOCKCHAIN] Initializing new blockchain for node 3000
# [BLOCKCHAIN] Genesis block created - Hash: 00001a2b3c...
```

#### Send Transactions
```bash
# Send coins from one address to another
./blockchain send -from SENDER_ADDR -to RECEIVER_ADDR -amount 50

# Mine transaction immediately
./blockchain send -from SENDER_ADDR -to RECEIVER_ADDR -amount 50 -mine

# Example output:
# [TRANSACTION] New transaction created: 3d4e5f6g...
# Success! Transaction included in block
```

#### Check Balance
```bash
# Get balance for a wallet address
./blockchain getbalance -address YOUR_ADDRESS

# Example output:
# Balance of YOUR_ADDRESS: 100
```

#### Display Blockchain
```bash
# Print all blocks in the chain
./blockchain printchain

# Example output:
# Block Hash: 00001a2b3c...
# Previous Hash: 0000000000...
# Height: 0
# Transactions: 1
# PoW: true
# -------------------
```

#### Reindex UTXO Set
```bash
# Rebuild the UTXO set from the blockchain
./blockchain reindex

# Use when UTXO set becomes corrupted or outdated
```

### REST API Endpoints

#### Get Blockchain
```bash
# Retrieve all blocks in the chain
curl http://localhost:3000/blockchain/printchain

# Response: Array of blocks with full details
```

#### Check Balance
```bash
# Get balance for a specific address
curl http://localhost:3000/blockchain/getbalance/YOUR_ADDRESS

# Response:
# {
#   "balance": 100
# }
```

#### Create Contract
```bash
# Create a new smart contract with milestones
curl -X POST http://localhost:3000/blockchain/createcontract/CREATOR_ADDRESS \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Website Development",
    "description": "Build e-commerce site",
    "milestones": [
      {
        "title": "Design Phase",
        "description": "Complete UI/UX design",
        "value": 1000,
        "dueDate": 1735689600
      }
    ],
    "parties": [
      {
        "address": "CONTRACTOR_ADDRESS",
        "role": "CONTRACTOR"
      }
    ]
  }'

# Response:
# {
#   "contractID": "a1b2c3d4...",
#   "status": "success"
# }
```

#### Approve Contract
```bash
# Party approves and signs the contract
curl http://localhost:3000/blockchain/approvecontract/PARTY_ADDRESS/CONTRACT_ID

# Response:
# {
#   "status": "Contract approved and signed"
# }
```

#### Get Contract
```bash
# Retrieve contract details
curl http://localhost:3000/blockchain/getcontract/CONTRACT_ID

# Response: Full contract with all milestones and parties
```

#### Approve Milestone
```bash
# Approve a milestone completion
curl -X POST http://localhost:3000/blockchain/approvemilestone/CONTRACT_ID/MILESTONE_ID/APPROVER_ADDRESS

# Response:
# {
#   "status": "Milestone approved"
# }
```

## 🎯 Proof-of-Work

Foedus uses SHA-256-based Proof-of-Work with adjustable difficulty:

```go
const Difficulty = 12  // Number of leading zeros required
```

**Mining Process:**
1. Collect pending transactions
2. Build a new block
3. Find nonce that produces hash with required leading zeros
4. Broadcast new block to network

## 💰 UTXO Model

**Unspent Transaction Output** model ensures:
- No double-spending
- Efficient balance calculation
- Transaction privacy
- Simplified verification

**Transaction Flow:**
1. Select unspent outputs (inputs)
2. Create new outputs (recipients)
3. Sign with private key
4. Broadcast to network
5. Update UTXO set after mining

## 📜 Smart Contracts

Foedus implements milestone-based contracts for:
- Freelance work
- Project delivery
- Escrow services
- Multi-party agreements

**Contract Lifecycle:**
1. **DRAFT** - Created, awaiting signatures
2. **ACTIVE** - All parties signed, milestones in progress
3. **COMPLETED** - All milestones completed
4. **CANCELLED** - Contract terminated

**Milestone Statuses:**
- **ACTIVE** - In progress
- **COMPLETED** - Finished and approved
- **CANCELLED** - Milestone cancelled

## 📁 File Structure

```
blockchain/
├── blockchain.go    # Chain management and persistence
├── block.go         # Block structure and creation
├── transaction.go   # Transaction logic
├── proof.go         # Proof-of-Work algorithm
├── contract.go      # Smart contract structures
├── utxo.go          # UTXO set management
├── tx.go            # Transaction helpers
├── iterator.go      # Chain traversal
├── serialize.go     # Serialization utilities
├── hash.go          # Hashing functions
├── operation.go     # Contract operations
└── README.md        # This file
```

## 🔐 Security Features

- **Cryptographic signing** of all transactions
- **Merkle tree** for transaction integrity
- **Proof-of-Work** prevents chain manipulation
- **UTXO validation** prevents double-spending
- **Address validation** before transactions

## 💡 Common Operations

### Mine a Block
```go
block := Block{
    Hash:         []byte{},
    Transactions: txs,
    PrevHash:     prevHash,
}
pow := NewProofOfWork(&block)
nonce, hash := pow.Run()  // Mining happens here
```

### Validate Chain
```go
iter := chain.Iterator()
for {
    block := iter.Next()
    pow := NewProofOfWork(block)
    if !pow.Validate() {
        return false
    }
}
```

## ⚙️ Configuration

- **Difficulty**: Adjustable in `proof.go` (default: 12)
- **Database Path**: `temp/blocks_<NODE_ID>`
- **Genesis Reward**: Configurable in `transaction.go`

## 🔗 Dependencies

- `github.com/cockroachdb/pebble` - Database storage
- `crypto/sha256` - Hashing algorithm
- `encoding/gob` - Data serialization

## 📚 Related Modules

- **Wallet** - Transaction signing and address generation
- **Database** - Persistent storage layer
- **Merkle** - Transaction verification
- **API** - HTTP interface for blockchain operations
- **CLI** - Command-line blockchain management

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
