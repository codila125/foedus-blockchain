# 🎨 CLI Module

> Powerful and intuitive command-line interface for Foedus blockchain operations.

The CLI module transforms complex blockchain operations into simple, memorable commands. Manage wallets, send transactions, mine blocks, and inspect the chain—all from your terminal.

---

## ✨ Features

- 🚀 **Simple Commands** - Easy-to-remember blockchain operations
- 📊 **Formatted Output** - Beautiful, readable terminal displays
- ✅ **Input Validation** - Prevent errors before execution
- 🎯 **Smart Defaults** - Sensible behavior out of the box
- 📖 **Built-in Help** - Comprehensive usage information
- 🔄 **Multi-Node Support** - Manage multiple blockchain instances

## 🏗️ Architecture

### Core Components

- **`cli.go`** - Main CLI logic and command routing
- **`command.go`** - Command execution functions
- **`format.go`** - Output formatting utilities

### Key Structure

```go
type CommandLine struct{}
```

## 🚀 Available Commands

### Wallet Management

#### Create Wallet
```bash
# Generate a new wallet with ECDSA key pair
./blockchain createwallet

# Output:
# New address: 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S
```

#### List Addresses
```bash
# Display all wallet addresses in the system
./blockchain listaddresses

# Output:
# 1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S
# 1Z2Y3X4W5V6U7T8S9R0Q1P2O3N4M5L6K7J8I9H
```

### Blockchain Management

#### Create Blockchain
```bash
# Initialize a new blockchain with genesis block
# The specified address receives the genesis reward
export NODE_ID=3000
./blockchain createblockchain -address YOUR_ADDRESS

# Output:
# [BLOCKCHAIN] Initializing new blockchain for node 3000
# [BLOCKCHAIN] Genesis block created
# Blockchain created successfully!
```

#### Print Chain
```bash
# Display all blocks in the blockchain
./blockchain printchain

# Output:
# ============================================================
# Block Hash: 00001a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v
# Previous Hash: 0000000000000000000000000000000000000000000000000
# Height: 0
# Timestamp: 2025-10-23 10:30:00
# Transactions: 1
# Proof of Work: true
# ============================================================
```

#### Reindex UTXO
```bash
# Rebuild the UTXO (Unspent Transaction Output) set
# Use when UTXO set becomes corrupted or out of sync
./blockchain reindex

# Output:
# [UTXO] Reindexing UTXO set...
# [UTXO] Found 10 unspent outputs
# [UTXO] Reindex complete!
```

### Transaction Operations

#### Send Coins
```bash
# Send coins from one address to another
./blockchain send -from SENDER_ADDRESS -to RECEIVER_ADDRESS -amount 50

# Mine immediately on local node
./blockchain send -from SENDER_ADDRESS -to RECEIVER_ADDRESS -amount 50 -mine

# Output:
# [TRANSACTION] Creating new transaction...
# [TRANSACTION] Transaction ID: 3d4e5f6g7h8i9j0k...
# Success! Transaction sent
```

#### Check Balance
```bash
# Get balance for a specific wallet address
./blockchain getbalance -address YOUR_ADDRESS

# Output:
# Balance of YOUR_ADDRESS: 150 coins
```

### Network Operations

#### Start Node
```bash
# Start a blockchain node
export NODE_ID=3000
./blockchain startnode

# Start node with mining enabled
./blockchain startnode -miner YOUR_MINER_ADDRESS

# Output:
# [NODE] Starting node 3000...
# [NODE] Node ready at :3000
# [MINING] Mining enabled for address: YOUR_MINER_ADDRESS
```

## 📖 Command Reference

### General Syntax
```bash
./blockchain <command> [options]
```

### Command List

| Command | Description | Required Flags | Optional Flags |
|---------|-------------|----------------|----------------|
| `createwallet` | Generate new wallet | - | - |
| `listaddresses` | Show all addresses | - | - |
| `createblockchain` | Initialize chain | `-address` | - |
| `printchain` | Display all blocks | - | - |
| `getbalance` | Check balance | `-address` | - |
| `send` | Send transaction | `-from`, `-to`, `-amount` | `-mine` |
| `reindex` | Rebuild UTXO set | - | - |
| `startnode` | Start network node | - | `-miner` |

## 🔧 Environment Variables

### NODE_ID (Required)
```bash
# Set the node identifier (used for data directory)
export NODE_ID=3000

# Different nodes need different IDs
export NODE_ID=3001  # For second node
export NODE_ID=3002  # For third node
```

This creates separate data directories:
- `temp/blocks_3000/`
- `temp/blocks_3001/`
- `temp/blocks_3002/`

## 💡 Usage Examples

### Complete Workflow

#### 1. Setup Environment
```bash
export NODE_ID=3000
```

#### 2. Create Wallet
```bash
./blockchain createwallet
# Output: New address: 1A2B3C...
```

#### 3. Initialize Blockchain
```bash
./blockchain createblockchain -address 1A2B3C...
```

#### 4. Check Initial Balance
```bash
./blockchain getbalance -address 1A2B3C...
# Output: Balance: 100 (genesis reward)
```

#### 5. Create Second Wallet
```bash
./blockchain createwallet
# Output: New address: 1Z2Y3X...
```

#### 6. Send Transaction
```bash
./blockchain send -from 1A2B3C... -to 1Z2Y3X... -amount 50 -mine
```

#### 7. Verify Balances
```bash
./blockchain getbalance -address 1A2B3C...
# Output: Balance: 50

./blockchain getbalance -address 1Z2Y3X...
# Output: Balance: 50
```

### Multi-Node Setup

#### Node 1 (Mining Node)
```bash
export NODE_ID=3000
./blockchain createblockchain -address YOUR_ADDRESS
./blockchain startnode -miner YOUR_ADDRESS
```

#### Node 2 (Regular Node)
```bash
export NODE_ID=3001
./blockchain startnode
```

## 🎯 Best Practices

### ✅ Do
- Always set `NODE_ID` before running commands
- Validate addresses before sending transactions
- Backup `wallets.dat` and `temp/blocks_*` directories
- Use `-mine` flag for quick local testing
- Regularly reindex UTXO set for data integrity

### ❌ Don't
- Share private keys or `wallets.dat` file
- Use same `NODE_ID` for multiple nodes
- Send more coins than available balance
- Delete blockchain data without backup

## 🐛 Troubleshooting

### Error: "NODE_ID env. var is not set!"
```bash
# Solution: Set the environment variable
export NODE_ID=3000
```

### Error: "Blockchain already exists"
```bash
# Solution: Use different NODE_ID or delete existing data
rm -rf temp/blocks_3000
```

### Error: "Not enough funds"
```bash
# Solution: Check balance first
./blockchain getbalance -address YOUR_ADDRESS
```

### Error: "Invalid wallet address"
```bash
# Solution: Verify address with listaddresses
./blockchain listaddresses
```

## 📁 File Structure

```
cli/
├── cli.go        # Main CLI logic and command parsing
├── command.go    # Command execution functions
├── format.go     # Output formatting utilities
└── README.md     # This file
```

## 🔗 Dependencies

- `flag` - Command-line flag parsing
- `fmt` - Formatted I/O
- `os` - Environment variables and exit codes

## 📚 Related Modules

- **Blockchain** - Core blockchain operations
- **Wallet** - Wallet management
- **API** - Alternative REST interface

## 🆘 Help & Support

For command help:
```bash
./blockchain
# Displays usage information and available commands
```

For detailed logging, check the console output with prefixes:
- `[BLOCKCHAIN]` - Blockchain operations
- `[TRANSACTION]` - Transaction processing
- `[MINING]` - Mining activities
- `[NETWORK]` - Network communications
- `[UTXO]` - UTXO set operations

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
