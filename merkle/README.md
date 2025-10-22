# 🌳 Merkle Tree Module

> Cryptographic data structure for efficient transaction verification.

Merkle trees enable compact representation and quick verification of large transaction sets. A single 32-byte root hash can verify thousands of transactions, making light clients and SPV possible.

---

## ✨ Features

- 🔐 **Cryptographic Security** - SHA-256 based tree construction
- 📊 **Efficient Verification** - O(log n) proof size
- 🎯 **Tamper Detection** - Any change invalidates the root
- 💾 **Space Efficient** - Store only root hash in block header
- ⚡ **Quick Validation** - Fast transaction existence proofs
- 📱 **SPV Support** - Enable lightweight blockchain clients

## 🏗️ Architecture

### Core Component

- **`merkle.go`** - Merkle tree and node implementation

### Key Structures

```go
type MerkleTree struct {
    RootNode *MerkleNode  // Root of the tree
}

type MerkleNode struct {
    Left  *MerkleNode    // Left child
    Right *MerkleNode    // Right child
    Data  []byte         // SHA-256 hash
}
```

## 🎯 How It Works

### Tree Construction Process

1. **Create leaf nodes** from transaction IDs
2. **Pair adjacent nodes** and hash them together
3. **Build parent nodes** from paired hashes
4. **Repeat until single root** node remains
5. **Root hash** represents all transactions

### Visual Example

```
Transactions: [Tx1, Tx2, Tx3, Tx4]

                  Root Hash
                 /          \
           Hash(12)          Hash(34)
           /      \          /      \
      Hash(1)  Hash(2)  Hash(3)  Hash(4)
         |        |        |        |
       Tx1      Tx2      Tx3      Tx4
```

### Odd Number of Transactions

When there's an odd number of transactions, the last one is duplicated:

```
Transactions: [Tx1, Tx2, Tx3]

                  Root Hash
                 /          \
           Hash(12)          Hash(33)
           /      \          /      \
      Hash(1)  Hash(2)  Hash(3)  Hash(3)
         |        |        |        |
       Tx1      Tx2      Tx3      Tx3 (duplicate)
```

## 🚀 Usage

### Creating a Merkle Tree

```go
import "github.com/codila125/foedus-blockchain/merkle"

// Collect transaction IDs
txIDs := [][]byte{
    tx1.ID,
    tx2.ID,
    tx3.ID,
    tx4.ID,
}

// Create Merkle tree
tree := merkle.NewMerkleTree(txIDs)

// Get root hash
rootHash := tree.RootNode.Data
fmt.Printf("Merkle Root: %x\n", rootHash)
```

### In Block Creation

```go
// From blockchain/block.go
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
    block := &Block{
        Transactions: txs,
        PrevHash:     prevHash,
    }
    
    // Create Merkle tree from transactions
    block.HashTransactions()
    
    return block
}

func (b *Block) HashTransactions() []byte {
    var txHashes [][]byte
    
    for _, tx := range b.Transactions {
        txHashes = append(txHashes, tx.ID)
    }
    
    tree := merkle.NewMerkleTree(txHashes)
    return tree.RootNode.Data
}
```

### Verification Example

```go
// Verify transaction exists in block
func VerifyTransaction(block *Block, txID []byte) bool {
    // Collect all transaction IDs
    var txIDs [][]byte
    for _, tx := range block.Transactions {
        txIDs = append(txIDs, tx.ID)
    }
    
    // Rebuild Merkle tree
    tree := merkle.NewMerkleTree(txIDs)
    
    // Compare root with block's stored hash
    return bytes.Equal(tree.RootNode.Data, block.Hash)
}
```

## 🔐 Security Properties

### Data Integrity
- **Any change** to a transaction changes the root hash
- **Tamper detection** is immediate and conclusive
- **Cryptographic guarantee** via SHA-256

### Efficient Verification
- **Logarithmic proof size**: O(log n) vs O(n)
- **Quick verification**: Don't need all transactions
- **Partial data**: Can verify single transaction

### Example Scenarios

#### Scenario 1: Valid Block
```
Original:  [Tx1, Tx2, Tx3, Tx4] → Root: 0xABCD...
Verify:    [Tx1, Tx2, Tx3, Tx4] → Root: 0xABCD... ✓ Valid
```

#### Scenario 2: Modified Transaction
```
Original:  [Tx1, Tx2, Tx3, Tx4] → Root: 0xABCD...
Verify:    [Tx1, Tx2', Tx3, Tx4] → Root: 0x1234... ✗ Invalid
```

#### Scenario 3: Missing Transaction
```
Original:  [Tx1, Tx2, Tx3, Tx4] → Root: 0xABCD...
Verify:    [Tx1, Tx2, Tx3] → Root: 0x5678... ✗ Invalid
```

## 💡 Advantages

### Space Efficiency
- Store only **root hash** in block header (32 bytes)
- Don't need to store entire tree
- Reconstruct tree when needed

### Verification Speed
- **Constant time** root verification
- **Logarithmic** transaction existence proof
- **Parallel** verification possible

### Data Privacy
- Can prove transaction exists **without revealing** all transactions
- Light clients can verify **without full blockchain**
- **Merkle proofs** enable SPV (Simplified Payment Verification)

## 🔍 Use Cases in Foedus

### 1. Block Validation
```go
// When receiving a new block
func ValidateBlock(block *Block) bool {
    // Rebuild Merkle tree
    var txIDs [][]byte
    for _, tx := range block.Transactions {
        txIDs = append(txIDs, tx.ID)
    }
    tree := merkle.NewMerkleTree(txIDs)
    
    // Verify root matches
    return bytes.Equal(tree.RootNode.Data, block.Hash)
}
```

### 2. Transaction Verification
```go
// Verify transaction is in blockchain
func TransactionInChain(chain *BlockChain, txID []byte) bool {
    iter := chain.Iterator()
    
    for {
        block := iter.Next()
        for _, tx := range block.Transactions {
            if bytes.Equal(tx.ID, txID) {
                return true
            }
        }
        if len(block.PrevHash) == 0 {
            break
        }
    }
    return false
}
```

### 3. Light Client Support
```go
// Light client only needs block headers with Merkle roots
type BlockHeader struct {
    MerkleRoot []byte
    PrevHash   []byte
    Timestamp  int64
}
```

## 📊 Performance Characteristics

| Operation | Complexity | Description |
|-----------|-----------|-------------|
| Tree Construction | O(n) | Build tree from n transactions |
| Root Retrieval | O(1) | Get root hash instantly |
| Transaction Proof | O(log n) | Prove transaction exists |
| Verification | O(log n) | Verify transaction in block |

## 🧪 Testing Examples

### Test Tree Construction
```go
func TestMerkleTree() {
    // Create test data
    data := [][]byte{
        []byte("Transaction 1"),
        []byte("Transaction 2"),
        []byte("Transaction 3"),
        []byte("Transaction 4"),
    }
    
    // Build tree
    tree := merkle.NewMerkleTree(data)
    
    // Verify root exists
    if tree.RootNode == nil {
        t.Error("Root node is nil")
    }
    
    // Verify root has data
    if len(tree.RootNode.Data) != 32 {
        t.Error("Root hash incorrect length")
    }
}
```

### Test Odd Number of Transactions
```go
func TestOddTransactions() {
    data := [][]byte{
        []byte("Tx1"),
        []byte("Tx2"),
        []byte("Tx3"),
    }
    
    tree := merkle.NewMerkleTree(data)
    
    // Should still create valid tree
    if tree.RootNode == nil {
        t.Error("Failed with odd transactions")
    }
}
```

## 📁 File Structure

```
merkle/
├── merkle.go     # Merkle tree implementation
└── README.md     # This file
```

## 🔗 Dependencies

- `crypto/sha256` - SHA-256 hashing algorithm

## 📚 Related Modules

- **Blockchain** - Uses Merkle trees in blocks
- **Transaction** - Provides data for tree construction
- **Block** - Stores Merkle root

## 🎯 Best Practices

### ✅ Do
- Build tree from transaction IDs (hashes)
- Store only root hash in block
- Rebuild tree for verification
- Use SHA-256 for hashing

### ❌ Don't
- Don't store entire tree structure
- Don't use weak hash functions
- Don't skip tree validation
- Don't modify tree after creation

## 🔬 Advanced Concepts

### Merkle Proof
A Merkle proof is a set of hashes needed to verify a transaction:

```
To prove Tx1 exists:
- Need: Hash(2), Hash(34)
- Compute: Hash(1) → Hash(12) → Root
- Compare with stored root
```

### Simplified Payment Verification (SPV)
Light clients can verify transactions without downloading full blockchain:
1. Download block headers (with Merkle roots)
2. Request Merkle proof for transaction
3. Verify proof against stored root

## 📖 Further Reading

- [Bitcoin Whitepaper](https://bitcoin.org/bitcoin.pdf) - Section 7: Reclaiming Disk Space
- [Merkle Tree Explained](https://en.wikipedia.org/wiki/Merkle_tree)
- [SPV Clients](https://bitcoin.org/en/operating-modes-guide#simplified-payment-verification-spv)

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
