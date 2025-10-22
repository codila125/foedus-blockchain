// Package database implements the database layer for the Foedus blockchain using PebbleDB.
package database

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

type PebbleDB struct {
	db *pebble.DB
}

// DBExists checks if a PebbleDB exists at the specified path.
func DBExists(dbPath string) bool {
	// Check if directory exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}
	// Check for Pebble marker files (MANIFEST or CURRENT)
	manifestPath := filepath.Join(dbPath, "CURRENT")
	_, err := os.Stat(manifestPath)
	return err == nil
}

// OpenDB opens a PebbleDB at the specified path.
func OpenDB(path string) (*PebbleDB, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &PebbleDB{db: db}, nil
}

// Close closes the PebbleDB.
func (pdb *PebbleDB) Close() error {
	return pdb.db.Close()
}

// Set sets a key-value pair in the PebbleDB.
func (pdb *PebbleDB) Set(key, value []byte) error {
	return pdb.db.Set(key, value, pebble.Sync)
}

// Get retrieves the value for a given key from the PebbleDB.
func (pdb *PebbleDB) Get(key []byte) ([]byte, error) {
	value, closer, err := pdb.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = closer.Close()
	}()
	return value, nil
}

// Delete removes a key-value pair from the PebbleDB.
func (pdb *PebbleDB) Delete(key []byte) error {
	return pdb.db.Delete(key, pebble.Sync)
}

// Batch creates a new batch for batch operations.
func (pdb *PebbleDB) Batch() *pebble.Batch {
	return pdb.db.NewBatch()
}

// ApplyBatch applies a batch of operations to the PebbleDB.
func (pdb *PebbleDB) ApplyBatch(batch *pebble.Batch) error {
	return pdb.db.Apply(batch, &pebble.WriteOptions{Sync: true})
}

func (pdb *PebbleDB) GetRawDB() *pebble.DB {
	return pdb.db
}
