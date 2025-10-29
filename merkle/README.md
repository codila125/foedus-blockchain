# Merkle Module

Merkle tree implementation for efficient transaction and contract verification in blocks.

## Overview

The merkle module provides a binary hash tree data structure used to create compact cryptographic proofs of data integrity. Each block in the Foedus blockchain uses Merkle roots to represent all transactions and contracts, enabling efficient verification without storing complete data sets.

## Architecture

```
merkle/
└── merkle.go    # Merkle tree and node implementation
```

**Components:**
- **MerkleTree**: Binary hash tree with root node
- **MerkleNode**: Tree node containing hash data and child references

**Key Structure:**
```go
type MerkleTree struct {
    RootNode *MerkleNode  // Root hash of entire tree
}

type MerkleNode struct {
    Left  *MerkleNode    // Left child node
    Right *MerkleNode    // Right child node
    Data  []byte         // SHA-256 hash (32 bytes)
}
```

## Core Operations

### Build Merkle Tree

```go
// Create tree from transaction IDs
txIDs := [][]byte{tx1.ID, tx2.ID, tx3.ID, tx4.ID}
tree := NewMerkleTree(txIDs)

// Access root hash
rootHash := tree.RootNode.Data
```

### Tree Construction Algorithm

1. Create leaf nodes by hashing each data item
2. If odd number of items, duplicate last item
3. Pair adjacent nodes and hash concatenation
4. Repeat pairing until single root remains

**Example with 4 transactions:**
```
         Root
        /    \
      H12    H34
     /  \   /  \
    H1  H2 H3  H4
    |   |  |   |
   TX1 TX2 TX3 TX4
```

**Example with 3 transactions (duplicates last):**
```
         Root
        /    \
      H12    H33
     /  \   /  \
    H1  H2 H3  H3*
    |   |  |   |
   TX1 TX2 TX3 TX3 (duplicated)
```

## Usage in Blockchain

### Block Transaction Root

```go
// Get all transaction IDs from block
txHashes := make([][]byte, len(block.Transactions))
for i, tx := range block.Transactions {
    txHashes[i] = tx.ID
}

// Build Merkle tree
tree := NewMerkleTree(txHashes)

// Store root in block (used in PoW)
block.TxRoot = tree.RootNode.Data
```

### Block Contract Root

```go
// Get all contract IDs from block
contractHashes := make([][]byte, len(block.Contracts))
for i, contract := range block.Contracts {
    contractHashes[i] = contract.ID
}

// Build Merkle tree
tree := NewMerkleTree(contractHashes)

// Store root in block
block.ContractRoot = tree.RootNode.Data
```

### Empty Block Handling

```go
// No transactions/contracts
if len(transactions) == 0 {
    return []byte{}  // Empty root
}

// Has transactions/contracts
tree := NewMerkleTree(transactionIDs)
return tree.RootNode.Data
```

## Implementation Details

### Node Creation

**Leaf Node (has data):**
```go
node := NewMerkleNode(nil, nil, data)
// Hashes: SHA256(data)
```

**Internal Node (has children):**
```go
node := NewMerkleNode(leftChild, rightChild, nil)
// Hashes: SHA256(leftChild.Data + rightChild.Data)
```

### Odd Number Handling

When building tree with odd number of leaves:
```go
if len(data)%2 != 0 {
    data = append(data, data[len(data)-1])  // Duplicate last
}
```

This ensures all levels have pairs for proper tree construction.

## Benefits

**Data Integrity:**
- Any change to a transaction invalidates the root hash
- Tamper-evident: changing one leaf changes entire path to root

**Efficiency:**
- Store only 32-byte root instead of all transaction data
- Verify transaction existence with O(log n) hashes
- Compact proofs for lightweight clients

**Use Cases:**
1. Transaction/contract verification
2. Proof of Work (Merkle roots in block hash)
3. Data integrity detection

## Algorithm Complexity

- **Construction**: O(n) - Linear in number of leaves
- **Space**: O(n) - Stores all nodes in tree
- **Verification**: O(log n) - Path from leaf to root
- **Root Access**: O(1) - Direct pointer

**Hash Function**: SHA-256 (32-byte output)

## Dependencies

- `crypto/sha256` - Cryptographic hash function

## Related Modules

- **Blockchain** - Uses Merkle roots for transaction/contract integrity
- **Proof** - Merkle roots included in PoW computation
