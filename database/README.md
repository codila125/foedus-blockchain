# Database Module

PebbleDB storage layer for blockchain persistence.

## Features

- High-performance LSM-tree storage
- Atomic batch operations
- Auto-flush capability
- Crash-safe writes

## Usage

```go
// Open database
db, err := database.OpenDB("./data/blocks_3008")
defer db.Close()

// Read/Write
value, err := db.Get(key)
err := db.Put(key, value)
err := db.Delete(key)

// Batch operations
batch := database.NewBatchWriter(db, 100) // flush every 100 ops
batch.Put(key, value)
batch.Flush()
```

## Structure

```
database/
├── db.go     # PebbleDB wrapper
└── batch.go  # Batch writer
```
