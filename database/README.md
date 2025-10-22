# 🗄️ Database Module

> High-performance persistent storage layer powered by PebbleDB.

This module wraps PebbleDB with a clean, blockchain-optimized interface. It handles all data persistence including blocks, transactions, UTXO sets, and smart contracts with atomic operations and crash recovery.

---

## ✨ Features

- ⚡ **PebbleDB Engine** - LSM-tree based key-value store
- 🔄 **Atomic Batches** - All-or-nothing write operations
- 💾 **Persistent Storage** - Survives crashes and restarts
- 🔒 **Lock Protection** - Prevents concurrent database access
- ✅ **Data Integrity** - Checksums and corruption detection
- 🚀 **High Performance** - Optimized for blockchain workloads

## 🏗️ Architecture

### Core Component

- **`db.go`** - PebbleDB wrapper with all database operations

### Key Structure

```go
type PebbleDB struct {
    db *pebble.DB  // Underlying PebbleDB instance
}
```

## 🚀 Usage

### Opening a Database

#### CLI Usage
```bash
# Database automatically created when blockchain is initialized
export NODE_ID=3000
./blockchain createblockchain -address YOUR_ADDRESS

# Database location: temp/blocks_3000/
```

#### API Usage
```bash
# Start API server (database loads automatically)
export NODE_ID=3000
./blockchain

# Database location: temp/blocks_3000/
```

### Programmatic Usage

```go
import "github.com/codila125/foedus-blockchain/database"

// Open or create database
db, err := database.OpenDB("./temp/blocks_3000")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Check if database exists
exists := database.DBExists("./temp/blocks_3000")
if exists {
    fmt.Println("Database found")
}
```

## 🔧 Operations

### Basic Operations

#### Set (Store)
```go
// Store a key-value pair
key := []byte("block:00001a2b3c")
value := []byte("block data here")
err := db.Set(key, value)
if err != nil {
    log.Fatal(err)
}
```

#### Get (Retrieve)
```go
// Retrieve value by key
key := []byte("block:00001a2b3c")
value, err := db.Get(key)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Value: %s\n", value)
```

#### Delete (Remove)
```go
// Delete a key-value pair
key := []byte("block:00001a2b3c")
err := db.Delete(key)
if err != nil {
    log.Fatal(err)
}
```

### Batch Operations

```go
// Create a new batch
batch := db.Batch()

// Add multiple operations
batch.Set([]byte("key1"), []byte("value1"), nil)
batch.Set([]byte("key2"), []byte("value2"), nil)
batch.Set([]byte("key3"), []byte("value3"), nil)

// Apply all operations atomically
err := db.ApplyBatch(batch)
if err != nil {
    log.Fatal(err)
}

// Always close the batch
batch.Close()
```

### Database Lifecycle

```go
// Open database
db, err := database.OpenDB("./temp/blocks_3000")

// Check if database exists before opening
if database.DBExists("./temp/blocks_3000") {
    db, err = database.OpenDB("./temp/blocks_3000")
}

// Close database when done
defer db.Close()
```

## 📊 Storage Structure

### Database Layout

```
temp/blocks_3000/
├── 000045.sst        # Sorted String Table files
├── CURRENT           # Points to current manifest
├── LOCK              # Database lock file
├── MANIFEST-000039   # Metadata manifest
├── MANIFEST-000043   # Metadata manifest
└── OPTIONS-000044    # Database options
```

### Key Naming Conventions

| Prefix | Usage | Example |
|--------|-------|---------|
| `lh` | Last hash pointer | `lh` → block hash |
| Block hash | Block data | `00001a2b...` → block bytes |
| `utxo-` | UTXO set | `utxo-txid` → outputs |
| `contract-` | Contracts | `contract-id` → contract data |

## 💡 Common Patterns

### Blockchain Block Storage

```go
// Store a block
block := &Block{...}
serialized := block.SerializeBlock()
err := db.Set(block.Hash, serialized)

// Store last hash pointer
err = db.Set([]byte("lh"), block.Hash)
```

### UTXO Set Management

```go
// Store UTXO
utxoKey := []byte("utxo-" + txID)
utxoData := SerializeUTXO(utxo)
err := db.Set(utxoKey, utxoData)

// Retrieve UTXO
data, err := db.Get(utxoKey)
utxo := DeserializeUTXO(data)
```

### Batch Block Mining

```go
// Mining stores multiple records atomically
rawDB := db.GetRawDB()
batch := rawDB.NewBatch()

// Store new block
batch.Set(block.Hash, block.SerializeBlock(), nil)

// Update last hash
batch.Set([]byte("lh"), block.Hash, nil)

// Apply atomically
err := rawDB.Apply(batch, &pebble.WriteOptions{Sync: true})
batch.Close()
```

## ⚙️ Configuration

### Database Options

PebbleDB is initialized with default options optimized for blockchain storage:

```go
db, err := pebble.Open(path, &pebble.Options{})
```

### Sync Options

All operations use synchronous writes for data integrity:

```go
// Individual operations
db.Set(key, value, pebble.Sync)

// Batch operations
db.Apply(batch, &pebble.WriteOptions{Sync: true})
```

## 🔐 Data Integrity

### Features

- **Atomic batch operations** - All-or-nothing writes
- **Synchronous writes** - Data flushed to disk immediately
- **Lock file protection** - Prevents concurrent access
- **Checksum verification** - Detects data corruption
- **Crash recovery** - Automatic recovery from unclean shutdown

### Backup Strategy

```bash
# Backup entire database directory
cp -r temp/blocks_3000 backup/blocks_3000_$(date +%Y%m%d)

# Or use rsync for efficiency
rsync -av temp/blocks_3000/ backup/blocks_3000/
```

## 🔍 Debugging

### Check Database Existence
```go
if database.DBExists("./temp/blocks_3000") {
    fmt.Println("Database exists")
} else {
    fmt.Println("Database not found")
}
```

### View Database Files
```bash
# List database files
ls -lh temp/blocks_3000/

# Check database size
du -sh temp/blocks_3000/
```

### Monitor Database Growth
```bash
# Watch database size in real-time
watch -n 1 'du -sh temp/blocks_3000/'
```

## 📁 File Structure

```
database/
├── db.go         # PebbleDB wrapper implementation
└── README.md     # This file
```

## 🚨 Error Handling

### Common Errors

#### Database Already Open
```
Error: lock: resource temporarily unavailable
Solution: Close existing database connection or use different NODE_ID
```

#### Permission Denied
```
Error: permission denied
Solution: Check directory permissions (chmod 755 temp/)
```

#### Corruption Detected
```
Error: pebble: corruption detected
Solution: Restore from backup or rebuild blockchain
```

## 🔗 Dependencies

- `github.com/cockroachdb/pebble` - High-performance key-value store
- `os` - File system operations
- `path/filepath` - Path manipulation

## 📚 Related Modules

- **Blockchain** - Primary consumer of database operations
- **CLI** - Initializes and manages database lifecycle
- **API** - Server-side database access

## 🎯 Best Practices

### ✅ Do
- Always close database connections with `defer db.Close()`
- Use batch operations for multiple writes
- Check `DBExists()` before opening
- Use synchronous writes for critical data
- Regular backups of database directory

### ❌ Don't
- Don't open same database from multiple processes
- Don't manually edit database files
- Don't ignore error returns
- Don't forget to close batches
- Don't delete database while node is running

## 📈 Performance Tips

1. **Batch Writes**: Use batch operations for multiple writes
2. **Key Design**: Use short, consistent key prefixes
3. **Read Optimization**: Cache frequently accessed data
4. **Cleanup**: Regularly compact database (automatic in PebbleDB)

## 🔄 Migration & Recovery

### Rebuild Database
```bash
# Stop node
# Delete database
rm -rf temp/blocks_3000/

# Reinitialize blockchain
./blockchain createblockchain -address YOUR_ADDRESS
```

### Reindex UTXO
```bash
# Rebuild UTXO set from blockchain
./blockchain reindex
```

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
