# Merkle Module

Binary hash tree for transaction verification.

## Features

- SHA-256 based Merkle tree
- Efficient proof verification
- Block integrity validation

## Usage

```go
// Build tree from transaction IDs
txIDs := [][]byte{tx1.ID, tx2.ID, tx3.ID}
tree := merkle.NewMerkleTree(txIDs)

// Get root hash
rootHash := tree.RootNode.Data
```

## How It Works

```
       Root Hash
       /       \
    Hash01    Hash23
    /    \    /    \
  H0    H1  H2    H3
  |     |   |     |
 Tx0   Tx1 Tx2   Tx3
```

## Structure

```
merkle/
└── merkle.go  # Tree implementation (100% test coverage)
```
