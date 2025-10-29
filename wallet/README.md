# Wallet Module

Cryptographic wallet management for Ed25519 key pairs, address generation, and signature operations.

## Overview

The wallet module provides secure key management and address generation for the Foedus blockchain. It uses Ed25519 elliptic curve cryptography for digital signatures and Base58 encoding for human-readable addresses with built-in checksums for validation.

## Architecture

```
wallet/
├── wallet.go      # Core wallet structure and key pair generation
├── wallets.go     # Wallet collection management
├── utils.go       # Base58 encoding/decoding utilities
└── files.go       # Wallet persistence (Protocol Buffers)
```

**Components:**
- **Wallet**: Individual wallet with Ed25519 key pair
- **Wallets**: Collection manager for multiple wallets
- **Address System**: Base58-encoded addresses with version and checksum
- **File Persistence**: Protobuf serialization for secure storage

**Key Structure:**
```go
type Wallet struct {
    PublicKey  ed25519.PublicKey   // 32 bytes
    PrivateKey ed25519.PrivateKey  // 64 bytes
}

type Wallets struct {
    Wallets map[string]*Wallet  // Address -> Wallet mapping
}
```

## Core Operations

### Create Wallet

```go
// Generate new wallet with key pair
wallet := MakeWallet()

// Get address
address := wallet.Address()
// Returns: Base58-encoded address (e.g., "1A2B3C...")
```

### Wallet Collection Management

```go
// Load existing wallets or create empty collection
wallets, err := CreateWallets(nodeID)

// Add new wallet
address := wallets.AddWallet()

// Get specific wallet
wallet, err := wallets.GetWallet(address)

// List all addresses
addresses := wallets.GetAllAddresses()

// Save to disk
err = wallets.SaveFile(nodeID)
```

### Address Validation

```go
### Address Validation

```go
if !ValidateAddress(address) {
    log.Fatal("Invalid address")
}
```

## Security Features

**Private Key Protection:**
- File permissions: 0600 (owner read/write only)
- Private keys never transmitted over network
- Secure random generation with `crypto/rand`

**Address Validation:**
- Checksum prevents typos and corruption
- Version byte enables network identification
- Ed25519 provides strong cryptographic security

## Dependencies
```

## Address Generation

**Process:**
1. Generate Ed25519 key pair (public + private keys)
2. Hash public key with SHA-256
3. Add version byte (0x00) as prefix
4. Compute checksum (double SHA-256, first 4 bytes)
5. Append checksum to versioned hash
6. Encode with Base58

**Address Structure:**
```
[1 byte version][32 bytes public key hash][4 bytes checksum]
```

**Base58 Encoding**: Uses charset without confusing characters (0, O, I, l excluded)

## Key Generation

**Ed25519 Algorithm:**
- Public key: 32 bytes
- Private key: 64 bytes
- Fast signature generation and verification
- High security (equivalent to 128-bit symmetric encryption)

```go
public, private := NewKeyPair()
// Uses crypto/rand for secure random generation
```

## Address Validation

**Checksum Verification:**
```go
func ValidateAddress(address string) bool {
    // 1. Base58 decode
    fullHash := Base58Decode(address)
    
    // 2. Extract components
    actualChecksum := fullHash[len-4:]
    version := fullHash[0]
    pubKeyHash := fullHash[1:len-4]
    
    // 3. Recompute checksum
    targetChecksum := Checksum(version + pubKeyHash)
    
    // 4. Compare
    return actualChecksum == targetChecksum
}
```

**Checksum Algorithm**: Double SHA-256, first 4 bytes

## File Persistence

**Storage Location:**
```
./temp/wallets_{NODE_ID}.data
```

**Format**: Protocol Buffers (binary serialization)

**Security**: File permissions set to 0600 (owner read/write only)

**Operations:**
- `LoadFile(nodeID)` - Load wallets from disk
- `SaveFile(nodeID)` - Save wallets to disk

**Protobuf Structure:**
```protobuf
message Wallets {
    map<string, Wallet> wallets = 1;
}

message Wallet {
    bytes private_key = 1;
    bytes public_key = 2;
}
```

## Usage Examples

### Create and Save Wallet

```go
// Initialize wallet collection
wallets, err := CreateWallets("3000")
if err != nil {
    // No existing wallets, starts empty
}

// Create new wallet
address := wallets.AddWallet()
log.Printf("New address: %s", address)

// Save to disk
err = wallets.SaveFile("3000")
if err != nil {
    log.Fatal(err)
}
```

### Load and Use Wallet

```go
// Load existing wallets
wallets, err := CreateWallets("3000")
if err != nil {
    log.Fatal(err)
}

// Get specific wallet
wallet, err := wallets.GetWallet(address)
if err != nil {
    log.Fatal("Wallet not found")
}

// Use private key for signing
signature := ed25519.Sign(wallet.PrivateKey, message)
```

### Validate Address

```go
if !ValidateAddress(address) {
    log.Fatal("Invalid address")
}

// Safe to use address
```

## Integration with Blockchain

### Transaction Signing

```go
// Get sender's wallet
wallet, err := wallets.GetWallet(fromAddress)

// Create transaction
tx := NewTransaction(&wallet, toAddress, amount, utxoSet)

// Sign transaction (uses wallet.PrivateKey internally)
chain.SignTransaction(tx, wallet.PrivateKey)
```

### Address Ownership Verification

```go
// Extract public key hash from address
pubKeyHash := Base58Decode(address)
pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]

// Compare with transaction output
output.IsLockedWithKey(pubKeyHash)  // Check ownership
```

### UTXO Queries

```go
// Get balance by address
pubKeyHash := PublicKeyHash(wallet.PublicKey)
utxos := utxoSet.FindUnspentTransactions(pubKeyHash)

balance := 0
for _, utxo := range utxos {
    balance += utxo.Value
}
```

## Security Considerations

**Private Key Protection:**
- Never expose private keys
- Store in encrypted file with restricted permissions (0600)
- Private keys never transmitted over network

**Address Validation:**
- Always validate addresses before use
- Checksum prevents typos and corruption
- Version byte enables network identification

**Key Generation:**
- Uses `crypto/rand` for secure randomness
- Ed25519 provides strong cryptographic security
- Each wallet has unique key pair

## Utility Functions

### Base58 Encoding/Decoding

```go
// Encode bytes to Base58
encoded := Base58Encode([]byte{0x01, 0x02, 0x03})

// Decode Base58 to bytes
decoded := Base58Decode([]byte("base58string"))
```

### Hashing

```go
// SHA-256 hash of public key
pubKeyHash := PublicKeyHash(publicKey)

// Generate checksum (double SHA-256)
checksum := Checksum(payload)
```

## Constants

```go
const (
    version        = 0x00  // Address version byte
    checksumLength = 4     // Checksum size in bytes
)
```

**Version Byte**: 0x00 for mainnet (configurable for testnet/other networks)

## Error Handling

**Wallet Not Found:**
```go
wallet, err := wallets.GetWallet(address)
if err != nil {
    // Wallet doesn't exist in collection
}
```

**Invalid Address:**
```go
if !ValidateAddress(address) {
    // Address checksum failed
    // Possible typo or corruption
}
```

**File Operations:**
- LoadFile returns error if file doesn't exist (expected for first run)
- SaveFile returns error if directory creation or write fails

## Dependencies

- `crypto/ed25519` - Elliptic curve cryptography
- `crypto/sha256` - Hashing algorithm
- `crypto/rand` - Secure random generation
- `github.com/mr-tron/base58` - Base58 encoding
- `google.golang.org/protobuf` - Binary serialization

## Related Modules

- **Blockchain** - Transaction signing and validation
- **CLI** - Wallet creation and management commands
- **API** - Wallet operations via HTTP endpoints
