# CLI Module

Command-line interface for interacting with the Foedus blockchain.

## Overview

The CLI module provides a terminal-based interface for blockchain operations including wallet management, transaction processing, blockchain initialization, and network node control. Commands are validated and executed through a flag-based argument parser.

## Architecture

```
cli/
├── cli.go        # Command parser and router
├── command.go    # Command implementations
└── format.go     # Output formatting utilities
```

**Components:**
- **CommandLine**: Entry point for parsing and dispatching commands
- **Command Functions**: Execute blockchain operations (wallet, transactions, mining)
- **Formatters**: Pretty-print blocks, transactions, and contracts

## Commands

### Blockchain Management

**Create Blockchain**
```bash
export NODE_ID=3000
./blockchain createblockchain -address <ADDRESS>
```
Initializes new blockchain with genesis block. Sends initial mining reward to specified address. Creates UTXO and ICCT index sets.

**Print Chain**
```bash
./blockchain printchain
```
Displays complete blockchain with formatted blocks, transactions, and contracts. Shows PoW validation status for each block.

**Reindex**
```bash
./blockchain reindex
```
Rebuilds UTXO and ICCT sets from blockchain data. Use after network sync or to resolve index corruption.

### Wallet Operations

**Create Wallet**
```bash
./blockchain createwallet
```
Generates new Ed25519 key pair and returns wallet address. Wallet persisted to node storage.

**List Addresses**
```bash
./blockchain listaddresses
```
Displays all wallet addresses stored on the node.

**Get Balance**
```bash
./blockchain getbalance -address <ADDRESS>
```
Calculates address balance by summing unspent transaction outputs (UTXOs).

### Transaction Processing

**Send Coins**
```bash
./blockchain send -from <FROM_ADDRESS> -to <TO_ADDRESS> -amount <AMOUNT>
```
Creates transaction, mines new block locally, and updates UTXO set. Validates sender and receiver addresses before execution.

### Network Operations

**Start Node**
```bash
./blockchain startnode -source <MULTIADDRESS>
```
Launches node and connects to P2P network. Optional `-source` flag specifies mining mode with source node multiaddress.

## Command Reference

| Command | Flags | Description |
|---------|-------|-------------|
| `createwallet` | - | Generate new Ed25519 wallet |
| `listaddresses` | - | List all wallet addresses |
| `createblockchain` | `-address` | Initialize blockchain with genesis block |
| `getbalance` | `-address` | Get address balance from UTXO set |
| `send` | `-from`, `-to`, `-amount` | Create and mine transaction |
| `printchain` | - | Display all blocks with formatting |
| `reindex` | - | Rebuild UTXO and ICCT index sets |
| `startnode` | `-source` | Start network node (optional mining) |

## Usage Examples

**Complete Workflow:**
```bash
# Set node identifier
export NODE_ID=3000

# Create wallet
./blockchain createwallet
# Output: New address: 1A2B3C...

# Initialize blockchain
./blockchain createblockchain -address 1A2B3C...

# Check balance (genesis reward)
./blockchain getbalance -address 1A2B3C...
# Output: Balance: 20

# Create second wallet
./blockchain createwallet
# Output: New address: 1Z2Y3X...

# Send transaction
./blockchain send -from 1A2B3C... -to 1Z2Y3X... -amount 10

# Verify balances
./blockchain getbalance -address 1Z2Y3X...
# Output: Balance: 10
```

## Output Formatting

The CLI provides formatted output for blockchain data:

**Block Display:**
- Block header with hash, height, timestamp, nonce
- PoW validation status (✓ VALID / ✗ INVALID)
- Transaction list with inputs/outputs
- Contract list with parties, milestones, signatures

**Contract Display:**
- Contract ID, title, status, creator
- Party list with roles and signature status
- Milestone list with values and completion status

**Transaction Display:**
- Transaction ID
- Inputs (coinbase or UTXO references)
- Outputs (recipient public key hash and value)

## Dependencies

- `flag` - CLI argument parsing

## Related Modules

- **API** - HTTP REST interface (alternative to CLI)
- **Blockchain** - Core blockchain operations
- **Wallet** - Key generation and address validation
- **Network** - Node networking
- **Network** - P2P communication layer

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
