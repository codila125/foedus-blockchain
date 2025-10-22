# 🔐 Wallet Module

> Secure cryptographic key management and address generation for the Foedus blockchain.

The wallet module implements industry-standard ECDSA (Elliptic Curve Digital Signature Algorithm) on the P-256 curve, providing Bitcoin-style addresses with Base58 encoding and checksum verification.

---

## ✨ Features

- 🔑 **ECDSA Key Pairs** - P-256 curve cryptography
- 📇 **Base58 Addresses** - Human-readable wallet addresses
- ✅ **Checksum Validation** - Typo-proof address verification
- ✍️ **Digital Signatures** - Transaction authentication
- 💾 **Multi-Wallet Storage** - Manage multiple wallets in one file
- 🛡️ **Secure Generation** - Cryptographically secure random keys

## 🏗️ Architecture

### Core Components

- **`wallet.go`** - Wallet structure and cryptographic operations
- **`wallets.go`** - Multi-wallet management and persistence
- **`utils.go`** - Base58 encoding/decoding utilities
- **`files.go`** - File operations for wallet storage

### Key Structures

```go
type Wallet struct {
    PrivateKey []byte  // ECDSA private key
    PublicKey  []byte  // ECDSA public key (X,Y coordinates)
}

type Wallets struct {
    Wallets map[string]*Wallet  // Address -> Wallet mapping
}
```

## 🚀 Usage

### CLI Commands

#### Create a New Wallet
```bash
# Create a new wallet and receive an address
./blockchain createwallet

# Example output:
# New address: 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S
```

#### List All Wallet Addresses
```bash
# Display all stored wallet addresses
./blockchain listaddresses

# Example output:
# 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S
# 1Z2Y3X4W5V6U7T8S9R0Q1P2O3N4M5L6K7J8I9H
```

### REST API Endpoints

#### Create Wallet
```bash
# Generate a new wallet via API
curl http://localhost:3000/blockchain/createwallet

# Response:
# {
#   "address": "1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S"
# }
```

#### List Addresses
```bash
# Retrieve all wallet addresses
curl http://localhost:3000/blockchain/listaddresses

# Response:
# {
#   "addresses": [
#     "1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S",
#     "1Z2Y3X4W5V6U7T8S9R0Q1P2O3N4M5L6K7J8I9H"
#   ]
# }
```

## 🔑 Address Format

Foedus uses a Bitcoin-style address format:

1. **Public Key** → SHA-256 hash
2. **Version byte** (0x00) prepended
3. **Checksum** (first 4 bytes of double SHA-256) appended
4. **Base58** encoding applied

**Example Address**: `1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S`

## 🛡️ Security Features

- **ECDSA P-256 Curve** - Industry-standard elliptic curve
- **SHA-256 Hashing** - Cryptographic hash functions
- **Checksum Validation** - Prevents address typos
- **Local Storage** - Wallets stored in `wallets.dat` file

## 📁 File Structure

```
wallet/
├── wallet.go      # Core wallet logic and cryptography
├── wallets.go     # Multi-wallet management
├── utils.go       # Base58 encoding utilities
├── files.go       # File I/O operations
└── README.md      # This file
```

## 💡 Common Operations

### Validate an Address
```go
isValid := wallet.ValidateAddress("1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S")
if isValid {
    fmt.Println("Valid address")
}
```

### Create and Store a Wallet
```go
wallets := wallet.CreateWallets()
address := wallets.AddWallet()
wallets.SaveFile()
```

## ⚠️ Important Notes

- **Backup your `wallets.dat` file** - Contains all private keys
- **Never share private keys** - Keep them secure and private
- **Addresses are case-sensitive** - Always verify before transactions
- **Loss of private key = Loss of funds** - No recovery mechanism

## 🔗 Dependencies

- `crypto/ecdsa` - Elliptic curve cryptography
- `crypto/sha256` - SHA-256 hashing
- `crypto/rand` - Secure random number generation

## 📚 Related Modules

- **Blockchain** - Uses wallets for transaction signing
- **API** - Exposes wallet operations via HTTP
- **CLI** - Command-line wallet management

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
