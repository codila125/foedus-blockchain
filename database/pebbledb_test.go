package database

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// DBExists Tests
// =============================================================================

func TestDBExists_NonExistent(t *testing.T) {
	exists := DBExists("/nonexistent/path/to/db")
	if exists {
		t.Error("DBExists should return false for non-existent path")
	}
}

func TestDBExists_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	exists := DBExists(tmpDir)
	if exists {
		t.Error("DBExists should return false for empty directory without CURRENT file")
	}
}

func TestDBExists_ValidDatabase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-valid-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	_ = db.Close()

	exists := DBExists(tmpDir)
	if !exists {
		t.Error("DBExists should return true for valid database")
	}
}

// =============================================================================
// OpenDB Tests
// =============================================================================

func TestOpenDB_NewDatabase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-new-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	dbPath := filepath.Join(tmpDir, "testdb")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open new database: %v", err)
	}
	defer func() { _ = db.Close() }()

	if db == nil {
		t.Fatal("OpenDB should return non-nil database")
	}
	if db.db == nil {
		t.Error("Internal pebble.DB should not be nil")
	}
}

func TestOpenDB_ExistingDatabase(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-existing-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db1, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	_ = db1.Close()

	db2, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer func() { _ = db2.Close() }()

	if db2 == nil {
		t.Error("OpenDB should return non-nil database when reopening")
	}
}

// =============================================================================
// PebbleDB.Close Tests
// =============================================================================

func TestPebbleDB_Close(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-close-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}
}

// =============================================================================
// PebbleDB.Get Tests
// =============================================================================

func TestPebbleDB_GetNonExistentKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-get-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, err = db.Get([]byte("nonexistent"))
	if err == nil {
		t.Error("Get should return error for non-existent key")
	}
}

func TestPebbleDB_SetAndGet(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-setget-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	key := []byte("testkey")
	value := []byte("testvalue")

	batch := db.Batch()
	if err := batch.Set(key, value, nil); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}
	if err := batch.Commit(nil); err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}

	retrieved, err := db.Get(key)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if !bytes.Equal(retrieved, value) {
		t.Errorf("Retrieved value %s does not match original %s", retrieved, value)
	}
}

func TestPebbleDB_MultipleKeyValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-multi-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	batch := db.Batch()
	for k, v := range data {
		if err := batch.Set([]byte(k), []byte(v), nil); err != nil {
			t.Fatalf("Failed to set %s: %v", k, err)
		}
	}
	if err := batch.Commit(nil); err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}

	for k, expected := range data {
		retrieved, err := db.Get([]byte(k))
		if err != nil {
			t.Errorf("Failed to get %s: %v", k, err)
			continue
		}
		if string(retrieved) != expected {
			t.Errorf("Key %s: expected %s, got %s", k, expected, string(retrieved))
		}
	}
}

func TestPebbleDB_BinaryData(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-binary-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	key := []byte{0x00, 0x01, 0x02, 0x03}
	value := []byte{0xFF, 0xFE, 0x00, 0xFD, 0xFC}

	batch := db.Batch()
	if err := batch.Set(key, value, nil); err != nil {
		t.Fatalf("Failed to set binary value: %v", err)
	}
	if err := batch.Commit(nil); err != nil {
		t.Fatalf("Failed to commit batch: %v", err)
	}

	retrieved, err := db.Get(key)
	if err != nil {
		t.Fatalf("Failed to get binary value: %v", err)
	}

	if !bytes.Equal(retrieved, value) {
		t.Error("Binary data was not stored/retrieved correctly")
	}
}

// =============================================================================
// PebbleDB.Batch Tests
// =============================================================================

func TestPebbleDB_Batch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-batch-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	batch := db.Batch()
	if batch == nil {
		t.Error("Batch should return non-nil batch")
	}
	_ = batch.Close()
}

// =============================================================================
// PebbleDB.GetRawDB Tests
// =============================================================================

func TestPebbleDB_GetRawDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-rawdb-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	rawDB := db.GetRawDB()
	if rawDB == nil {
		t.Error("GetRawDB should return non-nil pebble.DB")
	}
}

// =============================================================================
// PebbleDB Benchmarks
// =============================================================================

func BenchmarkPebbleDB_Write(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-write-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := db.Batch()
		key := []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
		_ = batch.Set(key, value, nil)
		_ = batch.Commit(nil)
	}
}

func BenchmarkPebbleDB_Read(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-read-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	key := []byte("benchmark-key")
	value := []byte("benchmark-value")
	batch := db.Batch()
	_ = batch.Set(key, value, nil)
	_ = batch.Commit(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.Get(key)
	}
}
