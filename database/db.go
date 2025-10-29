// Package database implements the persistence layer for the Foedus blockchain using PebbleDB,
// a high-performance embedded key-value store.
package database

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

// PebbleDB wraps a Pebble database instance providing blockchain storage operations.
type PebbleDB struct {
	db *pebble.DB
}

// DBExists checks whether a PebbleDB database exists at the specified path
// by verifying the presence of required Pebble marker files.
func DBExists(dbPath string) bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false
	}
	manifestPath := filepath.Join(dbPath, "CURRENT")
	_, err := os.Stat(manifestPath)
	return err == nil
}

// OpenDB opens or creates a PebbleDB at the specified path.
func OpenDB(path string) (*PebbleDB, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	return &PebbleDB{db: db}, nil
}

// Close closes the PebbleDB connection and flushes any pending writes.
func (pdb *PebbleDB) Close() error {
	return pdb.db.Close()
}

// Get retrieves a value by key from the database. The caller must close the returned closer.
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

// Batch creates a new batch for atomic write operations.
func (pdb *PebbleDB) Batch() *pebble.Batch {
	return pdb.db.NewBatch()
}

// GetRawDB returns the underlying Pebble database instance for advanced operations.
func (pdb *PebbleDB) GetRawDB() *pebble.DB {
	return pdb.db
}
