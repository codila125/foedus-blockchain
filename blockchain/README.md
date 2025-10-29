# Blockchain Module

Core blockchain implementation with Proof-of-Work consensus, UTXO transaction model, and milestone-based smart contracts.

## Overview

The blockchain module provides the foundational data structures and algorithms for the Foedus blockchain. It combines Bitcoin's UTXO model for value transfer with custom smart contracts supporting milestone-based project agreements.

## Architecture

```
blockchain/
├── blockchain.go     # Chain management, mining, validation
├── block.go          # Block structure and creation
├── transaction.go    # Transaction creation, signing, verification
├── contract.go       # Smart contract structures and statuses
├── operation.go      # Contract operations and lifecycle
├── proof.go          # Proof-of-Work mining algorithm
├── utxo.go           # UTXO set management and indexing
├── icct.go           # Incomplete Contract Tracking (ICCT) set
├── tx.go             # Transaction inputs/outputs
├── hash.go           # Hashing utilities (Merkle roots)
├── iterator.go       # Blockchain traversal
└── serialize.go      # Binary serialization
```

**Core Data Structures:**

```go
type BlockChain struct {
    LastHash []byte             // Most recent block hash
    Database *database.PebbleDB // Persistent storage (PebbleDB)
}

type Block struct {
    Hash         []byte         // Block hash (PoW result)
    Transactions []*Transaction // UTXO transactions
    Contracts    []*Contract    // Smart contracts
    PrevHash     []byte         // Previous block hash
    Nonce        int            // PoW nonce
    Height       int            // Block number
    Timestamp    int64          // Creation time
}

type Transaction struct {
    ID      []byte      // Transaction hash
    Inputs  []TxInput   // References to UTXOs
    Outputs []TxOutput  // New UTXOs
}
```

## Key Features

**Proof-of-Work Consensus:**
- SHA-256 hashing with adjustable difficulty (default: 12)
- Nonce-based mining algorithm
- Block validation with PoW verification

**UTXO Transaction Model:**
- Bitcoin-style unspent transaction outputs
- Cryptographic signature verification (Ed25519)
- Double-spending prevention through UTXO tracking
- Efficient balance calculation

**Smart Contracts:**
- Milestone-based project agreements
- Multi-party signature support
- Contract lifecycle management (DRAFT → ACTIVE → COMPLETED)
- Evidence-based milestone approval

**State Management:**
- UTXO Set: Indexed unspent outputs for fast lookups
- ICCT Set: Incomplete Contract Tracking for active contracts
- Merkle tree-based transaction/contract integrity

## Core Operations

### Blockchain Initialization

```go
// Create new blockchain with genesis block
chain := NewBlockChain(address, nodeID)

// Continue existing blockchain
chain := ContinueBlockChain(nodeID)
```

### Transaction Processing

```go
// Create transaction
tx := NewTransaction(wallet, toAddress, amount, utxoSet)

// Sign transaction
chain.SignTransaction(tx, privateKey)

// Verify transaction
valid := chain.VerifyTransaction(tx, txMap)
```

### Block Mining

```go
// Mine new block with transactions and contracts
block := chain.MineBlock(transactions, contracts)

// Proof-of-Work
pow := NewProof(block)
nonce, hash := pow.Run()

// Validate block
valid := pow.Validate()
```

### Smart Contract Operations

```go
// Create contract
contract := CreateContract(title, description, wallet, milestones, parties, terms, attachments)

// Approve contract
err := contract.ApproveContract(wallet)

// Approve milestone
updatedContract, err := contract.ApproveMilestone(wallet, milestoneID, evidence)

// Find contract
contract, err := chain.FindContract(contractID)
```

### UTXO Management

```go
// Initialize UTXO set
utxoSet := UTXOSet{Blockchain: chain}

// Reindex from blockchain
utxoSet.Reindex()

// Find spendable outputs
acc, outputs := utxoSet.FindSpendableOutputs(pubKeyHash, amount)

// Update after mining
utxoSet.Update(block)
```

## Contract Lifecycle

**Statuses:**
- `DRAFT` - Created, awaiting party signatures
- `ACTIVE` - All parties signed, milestones in progress
- `COMPLETED` - All milestones fulfilled
- `CANCELLED` - Terminated prematurely

**Milestone Statuses:**
- `ACTIVE` - In progress
- `COMPLETED` - Delivered and approved
- `CANCELLED` - Voided

**Workflow:**
1. Creator initiates contract with milestones and parties
2. Parties review and sign contract (DRAFT → ACTIVE)
3. Contractor delivers milestone with evidence
4. Parties approve milestone completion
5. Payment released upon approval
6. Contract completes when all milestones done

## Validation Rules

**Transaction Validation:**
- All inputs reference valid UTXOs
- Input signatures verified with Ed25519 public keys
- Sum of inputs ≥ sum of outputs
- No double-spending within block

**Contract Validation:**
- All party addresses valid
- Creator signature present
- Milestone values > 0
- Contract ID matches hash

**Block Validation:**
- Previous block hash exists
- PoW hash meets difficulty target
- All transactions valid
- All contracts valid
- Merkle roots match computed values

## Storage Structure

```
temp/blocks_{NODE_ID}/
├── 000001.sst     # Sorted string tables
├── MANIFEST       # Database manifest
└── OPTIONS        # PebbleDB configuration
```

**Key Prefixes**: Block data (hash), UTXO (`utxo-{txID}`), ICCT (`icct-{contractID}`), Last hash (`lh`)

## Dependencies

- `github.com/cockroachdb/pebble` - Persistent storage
- `crypto/ed25519` - Digital signatures
- `crypto/sha256` - Cryptographic hashing
- `encoding/gob` - Binary serialization

## Related Modules

- **Wallet** - Transaction signing and verification
- **Network** - Block and contract propagation
- **Database** - Persistent storage layer
- **Merkle** - Transaction/contract verification
