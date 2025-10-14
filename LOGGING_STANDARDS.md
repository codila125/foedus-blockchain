# Blockchain Logging Standards

## Overview
This document describes the standardized logging format used throughout the Go Blockchain project.

## Log Format Structure

All logs follow a consistent format:
```
[CATEGORY] <Symbol> Message with relevant details
```

## Log Categories

### 1. **[BLOCKCHAIN]** - Core blockchain operations
- Blockchain initialization
- Block addition and validation
- Chain state changes

**Examples:**
```
[BLOCKCHAIN] Initializing new blockchain for node 3000
[BLOCKCHAIN] Genesis block created - Hash: 00017d1...
[BLOCKCHAIN] Blockchain loaded successfully with height 4
[BLOCKCHAIN] Block added and chain updated - Hash: 00011a1..., Height: 4
```

### 2. **[MINING]** - Mining and block creation
- Mining process initiation
- Transaction verification during mining
- Block creation and completion

**Examples:**
```
[MINING] Starting block mining with 2 transaction(s)
[MINING] Processing 3 transaction(s) from mempool
[MINING] ✓ Transaction 682b526c... verified
[MINING] ✗ Transaction 27850fbdf... rejected (invalid)
[MINING] ✓ Block mined successfully - Hash: 00074d0e..., Transactions: 3
```

### 3. **[TRANSACTION]** - Transaction operations
- Transaction creation
- Transaction signing and verification
- Fund transfers

**Examples:**
```
[TRANSACTION] Transaction created - ID: 682b526c..., Amount: 10
[TRANSACTION] Insufficient funds - Required: 100, Available: 20
[TRANSACTION] Signing failed - Invalid parent transaction: 5625e39c...
```

### 4. **[UTXO]** - UTXO set management
- UTXO reindexing
- UTXO set updates
- UTXO operations

**Examples:**
```
[UTXO] Starting UTXO set reindexing
[UTXO] UTXO set reindexed successfully - 15 transaction(s) in set
[UTXO] Updating UTXO set with block 00074d0e...
[UTXO] UTXO set updated successfully
```

### 5. **[NETWORK]** - Network communications
- Sending and receiving data
- Node connections
- Peer synchronization

**Examples:**
```
[NETWORK] → Sending block 00074d0e... (height: 4) to localhost:3000
[NETWORK] ← Received block 0002209d... (height: 3) from localhost:3000
[NETWORK] → Requesting blocks from localhost:3000
[NETWORK] ← Received version from localhost:4000 (peer height: 5, local height: 3)
[NETWORK] Local blockchain is behind by 2 block(s), requesting blocks
[NETWORK] Blockchains are synchronized with localhost:3000
[NETWORK] Node localhost:4000 is unavailable, removing from known nodes
```

### 6. **[VERIFY]** - Transaction and block verification
- Signature verification
- Transaction validation
- Parent transaction lookup

**Examples:**
```
[VERIFY] Transaction verification failed - Parent transaction 27850fbdf... not found
[VERIFY] Transaction 682b526c... signature verification failed
```

### 7. **[DATABASE]** - Database operations
- Database connections
- Lock file handling
- Database recovery

**Examples:**
```
[DATABASE] Database locked, attempting recovery
[DATABASE] Retrying database connection after removing lock file
[DATABASE] Database opened successfully after retry
```

### 8. **[SYSTEM]** - System-level operations
- Node startup and shutdown
- Configuration
- System events

**Examples:**
```
[SYSTEM] ═══════════════════════════════════════════════════
[SYSTEM] Starting blockchain node: localhost:3000
[SYSTEM] Mining enabled - Rewards to: 1A2B3C4D...
[SYSTEM] ═══════════════════════════════════════════════════
[SYSTEM] Shutting down node, closing database...
[SYSTEM] Database closed successfully
```

### 9. **[CLI]** - Command-line interface operations
- User commands
- Wallet operations
- Balance queries

**Examples:**
```
[CLI] Creating new blockchain for address: 1A2B3C4D...
[CLI] ✓ Blockchain created successfully
[CLI] Fetching balance for address: 1A2B3C4D...
[CLI] Balance retrieved: 50
[CLI] Initiating transaction: 10 from 1A2B3C4D... to 5E6F7G8H...
[CLI] ✓ Transaction mined in block 00074d0e...
```

### 10. **[WALLET]** - Wallet operations
- Wallet creation
- Key management
- Address generation

**Examples:**
```
[WALLET] New wallet created successfully
```

## Symbols Used

- **→** : Outgoing action (sending data)
- **←** : Incoming action (receiving data)
- **✓** : Success indicator
- **✗** : Failure/rejection indicator
- **═** : Visual separator for important events

## Best Practices

1. **Consistency**: Always use the correct category prefix
2. **Clarity**: Include relevant identifiers (hashes, addresses, amounts)
3. **Brevity**: Keep messages concise but informative
4. **Context**: Provide enough context to understand the action
5. **Visibility**: Use symbols to make critical operations stand out
6. **Hierarchy**: Use log levels appropriately:
   - `log.Printf()` for normal operations
   - `log.Panic()` for critical errors

## Hash Display

For readability, transaction and block hashes are typically truncated:
- Full hash: `682b526ccd01e208217792f9192a5401568e60541e279cd5343152dbf556ba64`
- Displayed: `682b526c...` (first 8 characters)

However, the code uses full hashes with `%x` format for complete visibility.

## Examples of Complete Operation Flows

### Mining a Block
```
[MINING] Processing 2 transaction(s) from mempool
[MINING] ✓ Transaction 682b526c... verified
[MINING] ✓ Transaction 14b44888... verified
[MINING] Mining block with 2 valid transaction(s)
[MINING] Starting block mining with 3 transaction(s)
[BLOCKCHAIN] Block mined successfully - Hash: 00011a10..., Height: 4
[UTXO] Updating UTXO set with block 00011a10...
[UTXO] UTXO set updated successfully
[MINING] ✓ Block mined successfully - Hash: 00011a10..., Transactions: 3
[MINING] Mempool cleared, 0 transaction(s) remaining
[NETWORK] → Sending inventory of 1 block to localhost:3000
```

### Synchronizing Nodes
```
[NETWORK] Connecting to central node: localhost:3000
[NETWORK] → Sending version (height: 2) to localhost:3000
[NETWORK] ← Received version from localhost:3000 (peer height: 4, local height: 2)
[NETWORK] Local blockchain is behind by 2 block(s), requesting blocks
[NETWORK] → Requesting blocks from localhost:3000
[NETWORK] ← Received inventory with 2 block from localhost:3000
[NETWORK] Requesting first block 0003b217... (total: 2)
[NETWORK] ← Received block 0003b217... (height: 3) from localhost:3000
[BLOCKCHAIN] Block added and chain updated - Hash: 0003b217..., Height: 3
```

---

**Document Version:** 1.0  
**Last Updated:** October 14, 2025
