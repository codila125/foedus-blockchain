# Database Module

PebbleDB wrapper providing persistent key-value storage for the Foedus blockchain.

## Overview

The database module is a thin abstraction layer over PebbleDB, offering blockchain-optimized storage operations. It manages all persistent data including blocks, transactions, UTXO sets, and contract states with atomic batch operations and automatic flushing.

## Architecture

```
database/
├── db.go      # PebbleDB wrapper with core operations
└── batch.go   # Batch writer with auto-flush capability
```

**Components:**
- **PebbleDB**: Main wrapper for database instance with CRUD operations
- **BatchWriter**: Efficient batch operations with automatic flushing at intervals

**Key Structure:**
```go
type PebbleDB struct {
    db *pebble.DB  // Underlying PebbleDB instance
}

type BatchWriter struct {
    rawDB      *pebble.DB
    batch      *pebble.Batch
    count      int
    flushEvery int  // Auto-flush threshold
}
```

## Core Operations

### Database Management

**Check Existence:**
```go
exists := DBExists("./temp/blocks_3000")
// Returns true if database exists at path
```

**Open Database:**
```go
db, err := OpenDB("./temp/blocks_3000")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

**Close Database:**
```go
err := db.Close()
// Flushes pending writes and releases resources
```

### Basic Operations

**Get Value:**
```go
value, err := db.Get([]byte("key"))
if err != nil {
    // Handle error (key not found, etc.)
}
```

**Get Raw DB:**
```go
rawDB := db.GetRawDB()
// Access underlying Pebble instance for advanced operations
```

### Batch Operations

**Simple Batch:**
```go
batch := db.Batch()
batch.Set(key1, value1, nil)
batch.Set(key2, value2, nil)
rawDB.Apply(batch, &pebble.WriteOptions{Sync: true})
batch.Close()
```

**Auto-Flush Batch Writer:**
```go
// Flush every 1000 writes
writer := NewBatchWriter(db.GetRawDB(), 1000)

for i := 0; i < 10000; i++ {
    err := writer.Write(key, value)
    if err != nil {
        log.Fatal(err)
    }
}

// Flush remaining data
writer.Close(true)  // true = sync to disk
```

## Usage in Blockchain

**Block Storage:**
```go
// Store block
batch := db.Batch()
batch.Set(block.Hash, block.SerializeBlock(), nil)
batch.Set([]byte("lh"), block.Hash, nil)  // Last hash
rawDB.Apply(batch, &pebble.WriteOptions{Sync: true})
batch.Close()

// Retrieve block
blockData, err := db.Get(blockHash)
block := DeserializeBlock(blockData)
```

**UTXO Set Management:**
```go
// Store UTXO with prefix
key := append([]byte("utxo-"), txID...)
batch.Set(key, utxoData, nil)

// Retrieve UTXO
utxoKey := append([]byte("utxo-"), txID...)
utxoData, err := db.Get(utxoKey)
```

**Contract Storage:**
```go
// Store incomplete contract
key := append([]byte("icct-"), contractID...)
batch.Set(key, contractData, nil)
```

## Storage Structure

**Data Directory:**
```
temp/blocks_{NODE_ID}/
├── 000001.sst        # Sorted String Tables (data files)
├── 000002.sst
├── CURRENT           # Current manifest pointer
├── LOCK              # Database lock file
├── MANIFEST-000003   # Manifest file (metadata)
└── OPTIONS-000004    # PebbleDB options
```

**Key Prefixes:**
- Block data: Raw block hash as key
- Last hash: `lh` key
- UTXO entries: `utxo-{txID}`
- Contract entries: `icct-{contractID}`

## Features

**PebbleDB Benefits:**
- LSM-tree architecture for write-optimized performance
- Automatic compression and compaction
- Point lookups and range scans
- Crash recovery with write-ahead logging
- Atomic batch operations

**Batch Writing:**
- Auto-flush at configurable intervals
- Reduces memory pressure for large writes
- Sync or async flush modes
- Efficient for bulk operations (reindexing, initial sync)

## Performance Features

**PebbleDB Benefits:**
- LSM-tree architecture for write-optimized performance
- Automatic compression and compaction
- Crash recovery with write-ahead logging

**Batch Writing:**
- Auto-flush at configurable intervals
- Efficient for bulk operations (reindexing, network sync)

**Best Practices:**
- Use batch operations for multiple writes
- Sync to disk for critical data (blocks, state changes)
- Close database connections properly

## Dependencies

- `github.com/cockroachdb/pebble` - Key-value storage engine

## Related Modules

- **Blockchain** - Block and state storage
- **UTXO** - Unspent output indexing  
- **ICCT** - Incomplete contract tracking
