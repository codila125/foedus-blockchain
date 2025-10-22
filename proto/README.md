# 📐 Proto Definitions

Protocol Buffer schemas for the Foedus blockchain. These `.proto` files define the structure of data used throughout the system, particularly for smart contracts and future network communication.

## 📋 Overview

This directory contains:
- **Type-safe schemas** for blockchain data structures
- **Language-agnostic definitions** for cross-platform compatibility
- **Version-controlled contracts** ensuring backward compatibility
- **Efficient serialization** specifications

---

## 📄 Proto Files

### `core.proto` - Smart Contracts
Defines contract and milestone structures for deterministic ID generation.

```protobuf
syntax = "proto3";
package protobuf;
option go_package = "github.com/codila125/foedus-blockchain/protobuf";

message ContractCore {
    string title = 1;
    string description = 2;
    int64 created_at = 3;
    repeated MilestoneCore milestones = 4;
    repeated PartyCoreData parties = 5;
    bytes terms = 6;
    string creator_address = 7;
    repeated bytes attachments = 8;
}

message MilestoneCore {
    string title = 1;
    string description = 2;
    int32 value = 3;
    int64 created_at = 4;
}

message PartyCoreData {
    string address = 1;
    string role = 2;
    bytes public_key = 3;
}
```

**Usage**: Contract ID generation, ensuring immutable core data

### `block.proto` - Blockchain Structures
Defines blocks and transactions for future network communication.

```protobuf
syntax = "proto3";
package protobuf;
option go_package = "github.com/codila125/foedus-blockchain/protobuf";

message Block {
    bytes hash = 1;
    repeated Transaction transactions = 2;
    bytes prev_hash = 3;
    int32 nonce = 4;
    int32 height = 5;
    int64 timestamp = 6;
}

message Transaction {
    bytes id = 1;
    repeated TxInput inputs = 2;
    repeated TxOutput outputs = 3;
}

message TxInput {
    bytes tx_id = 1;
    int32 out_idx = 2;
    bytes signature = 3;
    bytes pub_key = 4;
}

message TxOutput {
    int32 value = 1;
    bytes pub_key_hash = 2;
}
```

**Usage**: Block transmission between nodes, efficient serialization

### `wallet.proto` - Wallet Structures
Defines wallet data for potential encrypted storage or transmission.

```protobuf
syntax = "proto3";
package protobuf;
option go_package = "github.com/codila125/foedus-blockchain/protobuf";

message Wallet {
    bytes private_key = 1;
    bytes public_key = 2;
    string address = 3;
}

message Wallets {
    map<string, Wallet> wallets = 1;
}
```

**Usage**: Future encrypted wallet storage or backup

---

## 🔧 Generating Go Code

### Prerequisites

```bash
# Install Protocol Buffer compiler
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# Verify installation
protoc --version
# Output: libprotoc 3.x.x or higher
```

```bash
# Install Go plugin for protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Ensure it's in your PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Generate Code

```bash
# From project root
cd foedus-blockchain

# Generate all proto files
protoc --go_out=./protobuf \
       --go_opt=paths=source_relative \
       proto/*.proto

# Or generate individually
protoc --go_out=./protobuf --go_opt=paths=source_relative proto/core.proto
protoc --go_out=./protobuf --go_opt=paths=source_relative proto/block.proto
protoc --go_out=./protobuf --go_opt=paths=source_relative proto/wallet.proto
```

### Makefile Approach

Create a `Makefile` in the project root:

```makefile
.PHONY: proto clean-proto

proto:
	@echo "🔄 Generating Protocol Buffer code..."
	@protoc --go_out=./protobuf \
	        --go_opt=paths=source_relative \
	        proto/*.proto
	@echo "✅ Done! Generated code in protobuf/"

clean-proto:
	@echo "🧹 Cleaning generated proto files..."
	@rm -f protobuf/*.pb.go
	@echo "✅ Cleaned!"
```

Usage:
```bash
make proto        # Generate code
make clean-proto  # Remove generated files
```

---

## 📊 Field Numbering Strategy

Protocol Buffers use field numbers for serialization. **Once assigned, never change them!**

### Best Practices

| Range | Usage | Encoding Cost |
|-------|-------|---------------|
| **1-15** | Most frequently used fields | 1 byte |
| **16-2047** | Less frequent fields | 2 bytes |
| **2048+** | Rarely used fields | 3+ bytes |

### Example Strategy

```protobuf
message Contract {
    // Critical fields (1-15) - 1 byte encoding
    string id = 1;              // Most accessed
    string title = 2;           // Very common
    string status = 3;          // Frequently checked
    int64 created_at = 4;       // Always present
    repeated Milestone milestones = 5;  // Core data
    
    // Important fields (16-2047) - 2 bytes encoding
    string description = 16;    // Moderately used
    repeated Party parties = 17;
    bytes terms = 18;
    
    // Optional/rare fields (2048+) - 3+ bytes encoding
    bytes metadata = 2048;      // Rarely used
}
```

---

## 🎯 Usage in Foedus

### Contract ID Generation

```go
import (
    "crypto/sha256"
    "github.com/codila125/foedus-blockchain/protobuf"
    "google.golang.org/protobuf/proto"
)

func GenerateContractID(contract *blockchain.Contract) []byte {
    // Convert to protobuf
    pbContract := &protobuf.ContractCore{
        Title:          contract.Title,
        Description:    contract.Description,
        CreatedAt:      contract.CreatedAt,
        // ... more fields
    }
    
    // Serialize deterministically
    data, err := proto.Marshal(pbContract)
    if err != nil {
        log.Panic(err)
    }
    
    // Generate ID from hash
    hash := sha256.Sum256(data)
    return hash[:]
}
```

### Future: Block Transmission

```go
// Serialize block for network transmission
func SerializeBlock(block *blockchain.Block) ([]byte, error) {
    pbBlock := &protobuf.Block{
        Hash:     block.Hash,
        PrevHash: block.PrevHash,
        Nonce:    int32(block.Nonce),
        Height:   int32(block.Height),
        // ... transactions
    }
    
    return proto.Marshal(pbBlock)
}

// Deserialize received block
func DeserializeBlock(data []byte) (*blockchain.Block, error) {
    pbBlock := &protobuf.Block{}
    err := proto.Unmarshal(data, pbBlock)
    if err != nil {
        return nil, err
    }
    
    // Convert to internal structure
    return &blockchain.Block{
        Hash:     pbBlock.Hash,
        PrevHash: pbBlock.PrevHash,
        // ...
    }, nil
}
```

---

## 🔄 Modifying Proto Files

### Adding New Fields

✅ **Safe Operations:**
```protobuf
message Contract {
    string title = 1;
    string description = 2;
    
    // ✅ Add new field with new number
    string category = 3;  // Safe!
}
```

❌ **Unsafe Operations:**
```protobuf
message Contract {
    string title = 1;
    // ❌ DON'T delete this field!
    // string description = 2;  
    
    // ❌ DON'T reuse field number 2!
    int32 version = 2;  // DANGEROUS!
}
```

### Deprecating Fields

Use `reserved` to prevent field reuse:

```protobuf
message Contract {
    reserved 2, 5 to 10;  // Reserve field numbers
    reserved "old_field"; // Reserve field names
    
    string title = 1;
    string new_field = 3;
    // Fields 2, 5-10 cannot be reused
}
```

---

## 📐 Proto Syntax Reference

### Basic Types

| Proto Type | Go Type | Description |
|------------|---------|-------------|
| `bool` | `bool` | Boolean value |
| `int32` | `int32` | 32-bit integer |
| `int64` | `int64` | 64-bit integer |
| `uint32` | `uint32` | Unsigned 32-bit |
| `uint64` | `uint64` | Unsigned 64-bit |
| `float` | `float32` | Floating point |
| `double` | `float64` | Double precision |
| `string` | `string` | UTF-8 string |
| `bytes` | `[]byte` | Byte array |

### Field Rules

```protobuf
message Example {
    string required_field = 1;           // Single value (proto3 default)
    repeated string array_field = 2;     // Array/list
    map<string, int32> map_field = 3;    // Map/dictionary
}
```

### Nested Messages

```protobuf
message Contract {
    string title = 1;
    repeated Milestone milestones = 2;
    
    message Milestone {
        string title = 1;
        int32 value = 2;
    }
}
```

---

## 🧪 Testing Proto Changes

### Test Serialization

```go
func TestProtoSerialization(t *testing.T) {
    // Create message
    original := &protobuf.ContractCore{
        Title:       "Test Contract",
        Description: "Testing proto",
        CreatedAt:   time.Now().Unix(),
    }
    
    // Serialize
    data, err := proto.Marshal(original)
    if err != nil {
        t.Fatal(err)
    }
    
    // Deserialize
    decoded := &protobuf.ContractCore{}
    err = proto.Unmarshal(data, decoded)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify
    if decoded.Title != original.Title {
        t.Errorf("Expected %s, got %s", original.Title, decoded.Title)
    }
}
```

### Test Backward Compatibility

```go
func TestBackwardCompatibility(t *testing.T) {
    // Old version (without new field)
    old := &protobuf.ContractCore{
        Title: "Old Contract",
    }
    oldData, _ := proto.Marshal(old)
    
    // New version should handle old data
    new := &protobuf.ContractCore{}
    err := proto.Unmarshal(oldData, new)
    if err != nil {
        t.Fatal("Failed to decode old format")
    }
}
```

---

## 📁 File Structure

```
proto/
├── core.proto       # Contract and milestone definitions
├── block.proto      # Block and transaction definitions
├── wallet.proto     # Wallet structure definitions
└── README.md        # This file
```

**Generated code goes to:** `protobuf/` directory

---

## 🎯 Best Practices

### ✅ Do
- Use descriptive field names
- Document complex fields with comments
- Reserve deprecated field numbers
- Keep messages focused and simple
- Version your proto files
- Test backward compatibility
- Use appropriate field numbers (1-15 for frequent)

### ❌ Don't
- Don't change field numbers
- Don't reuse field numbers
- Don't change field types
- Don't remove required fields
- Don't use large default values
- Don't create deeply nested structures
- Don't forget to regenerate after changes

---

## 🔗 Package Options

Standard options used in Foedus:

```protobuf
syntax = "proto3";                                              // Use proto3 syntax
package protobuf;                                               // Package name
option go_package = "github.com/codila125/foedus-blockchain/protobuf";  // Go import path
```

---

## 📚 Additional Resources

### Official Documentation
- [Protocol Buffers](https://protobuf.dev/)
- [Proto3 Language Guide](https://protobuf.dev/programming-guides/proto3/)
- [Go Tutorial](https://protobuf.dev/getting-started/gotutorial/)

### Best Practices
- [Proto Style Guide](https://protobuf.dev/programming-guides/style/)
- [API Best Practices](https://protobuf.dev/programming-guides/api/)
- [Dos and Don'ts](https://protobuf.dev/programming-guides/dos-donts/)

### Performance
- [Encoding Guide](https://protobuf.dev/programming-guides/encoding/)
- [Optimization Tips](https://protobuf.dev/programming-guides/techniques/)

---

## 🔧 Troubleshooting

### Issue: `protoc: command not found`
```bash
# Install protoc compiler
brew install protobuf  # macOS
```

### Issue: `protoc-gen-go: program not found`
```bash
# Install Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Add to PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Issue: Import path errors
```bash
# Ensure correct go_package option in .proto file
option go_package = "github.com/codila125/foedus-blockchain/protobuf";
```

### Issue: Generated code out of sync
```bash
# Regenerate all proto files
make proto
# or
protoc --go_out=./protobuf --go_opt=paths=source_relative proto/*.proto
```

---

## 📝 Version History

| Version | Changes | Date |
|---------|---------|------|
| 1.0 | Initial proto definitions | Oct 2025 |
| - | Contract core structure | - |
| - | Block and transaction types | - |
| - | Wallet structures | - |

---

*For generated Go code documentation, see [protobuf/README.md](../protobuf/README.md)*

*For more information, see the main [Foedus Blockchain README](../README.md)*
