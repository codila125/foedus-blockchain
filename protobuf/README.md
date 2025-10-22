# 📦 Protocol Buffers Module

> Generated Go code for efficient, type-safe data serialization.

This module contains auto-generated code from Protocol Buffer definitions. It provides 3-10x smaller encoding than JSON and 20-100x faster parsing, perfect for deterministic contract IDs and future network communication.

---

## ✨ Features

- 🚀 **High Performance** - Binary encoding beats JSON/XML
- 🔒 **Type Safety** - Compile-time type checking
- 🌍 **Cross-Platform** - Works with multiple languages
- 📝 **Schema Evolution** - Backward/forward compatibility
- 🎯 **Deterministic** - Same data = same serialization
- 📦 **Compact Size** - Minimal bandwidth usage

## 🏗️ Architecture

### Core Components

#### Proto Definitions (`proto/`)
- **`core.proto`** - Contract and milestone structures
- **`block.proto`** - Block and transaction structures
- **`wallet.proto`** - Wallet and address structures

#### Generated Code (`protobuf/`)
- **`core.pb.go`** - Generated from core.proto
- **`block.pb.go`** - Generated from block.proto
- **`wallet.pb.go`** - Generated from wallet.proto

## 📄 Protocol Buffer Schemas

### Core Proto (Contracts)

```protobuf
// core.proto
syntax = "proto3";

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

### Block Proto (Future Use)

```protobuf
// block.proto
syntax = "proto3";

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
```

### Wallet Proto (Future Use)

```protobuf
// wallet.proto
syntax = "proto3";

message Wallet {
    bytes private_key = 1;
    bytes public_key = 2;
    string address = 3;
}

message Wallets {
    map<string, Wallet> wallets = 1;
}
```

## 🚀 Usage

### In Contract Operations

```go
import (
    "github.com/codila125/foedus-blockchain/blockchain"
    "github.com/codila125/foedus-blockchain/protobuf"
    "google.golang.org/protobuf/proto"
)

// Convert contract to protobuf for ID generation
func GenerateContractID(contract *blockchain.Contract) []byte {
    // Convert to protobuf message
    pbMilestones := make([]*protobuf.MilestoneCore, len(contract.Milestones))
    for i, m := range contract.Milestones {
        pbMilestones[i] = &protobuf.MilestoneCore{
            Title:       m.Title,
            Description: m.Description,
            Value:       int32(m.Value),
            CreatedAt:   m.CreatedAt,
        }
    }
    
    pbParties := make([]*protobuf.PartyCoreData, len(contract.Parties))
    for i, p := range contract.Parties {
        pbParties[i] = &protobuf.PartyCoreData{
            Address:   p.Address,
            Role:      string(p.Role),
            PublicKey: p.PublicKey,
        }
    }
    
    pbContract := &protobuf.ContractCore{
        Title:          contract.Title,
        Description:    contract.Description,
        CreatedAt:      contract.CreatedAt,
        Milestones:     pbMilestones,
        Parties:        pbParties,
        Terms:          contract.Terms,
        CreatorAddress: contract.CreatorAddress,
        Attachments:    contract.Attachments,
    }
    
    // Serialize to bytes
    data, err := proto.Marshal(pbContract)
    if err != nil {
        log.Panic(err)
    }
    
    // Hash for contract ID
    hash := sha256.Sum256(data)
    return hash[:]
}
```

### Serialization Example

```go
// Create protobuf message
milestone := &protobuf.MilestoneCore{
    Title:       "Design Phase",
    Description: "Complete UI/UX",
    Value:       1000,
    CreatedAt:   time.Now().Unix(),
}

// Serialize to bytes
data, err := proto.Marshal(milestone)
if err != nil {
    log.Fatal(err)
}

// Deserialize from bytes
newMilestone := &protobuf.MilestoneCore{}
err = proto.Unmarshal(data, newMilestone)
if err != nil {
    log.Fatal(err)
}
```

## 🔧 Generating Code

### Prerequisites
```bash
# Install Protocol Buffer compiler
brew install protobuf  # macOS
# or
apt-get install protobuf-compiler  # Linux

# Install Go protobuf plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### Generate Go Code

```bash
# From project root directory
cd proto

# Generate core.pb.go
protoc --go_out=../protobuf --go_opt=paths=source_relative core.proto

# Generate block.pb.go
protoc --go_out=../protobuf --go_opt=paths=source_relative block.proto

# Generate wallet.pb.go
protoc --go_out=../protobuf --go_opt=paths=source_relative wallet.proto

# Or all at once
protoc --go_out=../protobuf --go_opt=paths=source_relative *.proto
```

### Makefile (Recommended)

```makefile
# Makefile
.PHONY: proto

proto:
	@echo "Generating protobuf code..."
	protoc --go_out=./protobuf --go_opt=paths=source_relative proto/*.proto
	@echo "Done!"
```

Usage:
```bash
make proto
```

## 💡 Advantages of Protocol Buffers

### Performance
- **3-10x smaller** than JSON
- **20-100x faster** than XML
- **Binary format** for efficient transmission
- **Optimized parsing** and generation

### Compatibility
- **Cross-language** support (Go, Python, Java, C++, etc.)
- **Forward/backward** compatibility
- **Versioning** support
- **Schema evolution** without breaking changes

### Type Safety
- **Strongly typed** fields
- **Compile-time checking**
- **IDE support** with auto-completion
- **Clear contracts** between services

## 🔍 Use Cases in Foedus

### 1. Contract ID Generation
```go
// Deterministic contract IDs using protobuf serialization
contractID := GenerateContractID(contract)
```

### 2. Network Transmission (Future)
```go
// Efficient block transmission between nodes
block := &protobuf.Block{...}
data, _ := proto.Marshal(block)
network.Send(peer, data)
```

### 3. Data Persistence (Optional)
```go
// Alternative to gob encoding
contract := &protobuf.ContractCore{...}
data, _ := proto.Marshal(contract)
db.Set([]byte("contract:"+id), data)
```

### 4. API Responses (Future)
```go
// Efficient API response format
func GetContract(w http.ResponseWriter, r *http.Request) {
    contract := &protobuf.ContractCore{...}
    data, _ := proto.Marshal(contract)
    w.Header().Set("Content-Type", "application/x-protobuf")
    w.Write(data)
}
```

## 📊 Comparison

| Format | Size | Speed | Human-Readable | Type Safety |
|--------|------|-------|----------------|-------------|
| **Protobuf** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐⭐⭐⭐ |
| **JSON** | ⭐⭐⭐ | ⭐⭐⭐ | ✅ | ⭐⭐⭐ |
| **XML** | ⭐⭐ | ⭐⭐ | ✅ | ⭐⭐⭐⭐ |
| **Gob** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ❌ | ⭐⭐⭐⭐ |

## 📁 File Structure

```
proto/
├── core.proto       # Contract definitions
├── block.proto      # Block definitions
├── wallet.proto     # Wallet definitions
└── README.md        # Proto documentation

protobuf/
├── core.pb.go       # Generated code
├── block.pb.go      # Generated code
├── wallet.pb.go     # Generated code
└── README.md        # This file
```

## 🔗 Proto Field Tags

### Field Rules
- **optional** - Field may or may not be present
- **required** - Field must be present (proto2 only)
- **repeated** - Field can appear multiple times (array)

### Field Numbers
- **1-15**: Use for frequently used fields (1 byte encoding)
- **16-2047**: Use for less frequent fields (2 bytes encoding)
- **Never change** field numbers after deployment

### Example
```protobuf
message Contract {
    string id = 1;           // Most important, 1 byte
    string title = 2;        // Frequently used, 1 byte
    repeated Milestone milestones = 15;  // Array, 1 byte
    string description = 16;  // Less frequent, 2 bytes
}
```

## 🛠️ Development Workflow

### 1. Modify Proto Files
```bash
# Edit proto/core.proto
vim proto/core.proto
```

### 2. Generate Code
```bash
# Regenerate Go code
make proto
# or
protoc --go_out=./protobuf --go_opt=paths=source_relative proto/*.proto
```

### 3. Update Application Code
```go
// Use new fields in application
contract.NewField = "value"
```

### 4. Test Changes
```bash
go test ./...
```

## 🎯 Best Practices

### ✅ Do
- Use meaningful field names
- Document complex fields
- Version your proto files
- Keep proto files simple
- Use appropriate field numbers
- Test backward compatibility

### ❌ Don't
- Don't reuse field numbers
- Don't change field types
- Don't remove required fields
- Don't use reserved keywords
- Don't forget to regenerate code

## 🔐 Security Considerations

- Protobuf **doesn't validate** data integrity
- Always **validate** deserialized data
- Use **checksums/signatures** for tampering detection
- **Limit message size** to prevent DoS
- **Sanitize** user input before serialization

## 🔗 Dependencies

- `google.golang.org/protobuf` - Protocol Buffer runtime
- `protoc` - Protocol Buffer compiler
- `protoc-gen-go` - Go code generator

## 📚 Related Modules

- **Blockchain** - Uses protobuf for contract IDs
- **Network** (future) - Will use for node communication
- **API** (future) - Can use for efficient responses

## 📖 Further Reading

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [Go Protocol Buffers Tutorial](https://protobuf.dev/getting-started/gotutorial/)
- [Proto3 Language Guide](https://protobuf.dev/programming-guides/proto3/)
- [Best Practices](https://protobuf.dev/programming-guides/dos-donts/)

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
