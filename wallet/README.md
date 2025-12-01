# Wallet Module

Cryptographic key management with Ed25519.

## Features

- Ed25519 key pair generation
- Base58Check address encoding
- Secure wallet persistence
- Address validation

## Usage

```go
// Create wallet
wallet := wallet.MakeWallet()
address := wallet.Address()

// Manage multiple wallets
wallets, _ := wallet.CreateWallets(nodeID)
address := wallets.AddWallet()
wallets.SaveFile(nodeID)

// Validate address
isValid := wallet.ValidateAddress(address)
```

## Address Format

```
Version (1 byte) + Public Key Hash (20 bytes) + Checksum (4 bytes)
→ Base58 encoded
```

## Structure

```
wallet/
├── wallet.go   # Key pair generation
├── wallets.go  # Collection management
├── utils.go    # Base58 encoding
└── files.go    # Persistence
```
