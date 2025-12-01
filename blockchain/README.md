# Blockchain Module

Core consensus engine with UTXO transactions and smart contracts.

## Features

- **Proof-of-Work** consensus with SHA-256
- **UTXO Model** for parallel transaction processing
- **Smart Contracts** with milestone tracking
- **Merkle Trees** for transaction verification

## Key Components

| File | Purpose |
|------|---------|
| `blockchain.go` | Chain management |
| `block.go` | Block structure |
| `transaction.go` | UTXO transactions |
| `contract.go` | Smart contracts |
| `proof.go` | Mining algorithm |
| `utxo.go` | UTXO indexing |

## Usage

```go
// Create new blockchain
chain, err := blockchain.NewBlockChain(address, nodeID)

// Continue existing chain
chain, err := blockchain.ContinueBlockChain(nodeID)

// Mine block
block := chain.MineBlock(transactions, contracts)
```

## Error Types

```go
var ErrBlockchainExists = errors.New("blockchain already exists")
var ErrBlockchainNotFound = errors.New("blockchain not found")
```
