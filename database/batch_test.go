package database

import (
	"bytes"
	"os"
	"testing"
)

// =============================================================================
// NewBatchWriter Tests
// =============================================================================

func TestNewBatchWriter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-batchwriter-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 100)
	if bw == nil {
		t.Fatal("NewBatchWriter should return non-nil BatchWriter")
	}
	if bw.flushEvery != 100 {
		t.Errorf("Expected flushEvery to be 100, got %d", bw.flushEvery)
	}
	if bw.batch == nil {
		t.Error("BatchWriter should have non-nil batch")
	}
	_ = bw.Close(false)
}

func TestNewBatchWriter_DifferentThresholds(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-bwthresh-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	testCases := []int{1, 10, 100, 1000}
	for _, threshold := range testCases {
		bw := NewBatchWriter(db.GetRawDB(), threshold)
		if bw.flushEvery != threshold {
			t.Errorf("Expected flushEvery to be %d, got %d", threshold, bw.flushEvery)
		}
		_ = bw.Close(false)
	}
}

// =============================================================================
// BatchWriter.Write Tests
// =============================================================================

func TestBatchWriter_Write(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-bwwrite-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 100)

	err = bw.Write([]byte("key"), []byte("value"))
	if err != nil {
		t.Errorf("Write should not return error: %v", err)
	}
	_ = bw.Close(true)

	retrieved, err := db.Get([]byte("key"))
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if string(retrieved) != "value" {
		t.Errorf("Expected 'value', got '%s'", string(retrieved))
	}
}

func TestBatchWriter_WriteMultiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-bwmulti-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 100)

	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range data {
		if err := bw.Write([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Write failed for %s: %v", k, err)
		}
	}
	_ = bw.Close(true)

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

// =============================================================================
// BatchWriter.AutoFlush Tests
// =============================================================================

func TestBatchWriter_AutoFlush(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-autoflush-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 5)

	for i := 0; i < 5; i++ {
		key := []byte{byte(i)}
		value := []byte{byte(i + 100)}
		if err := bw.Write(key, value); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	// After 5 writes, data should be flushed and readable
	retrieved, err := db.Get([]byte{0})
	if err != nil {
		t.Errorf("Auto-flush should have persisted data: %v", err)
	}
	if len(retrieved) > 0 && retrieved[0] != 100 {
		t.Errorf("Expected value 100, got %d", retrieved[0])
	}

	_ = bw.Close(false)
}

func TestBatchWriter_AutoFlushMultipleTimes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-multiflush-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 3)

	// Write 9 entries (triggers 3 auto-flushes)
	for i := 0; i < 9; i++ {
		key := []byte{byte(i)}
		value := []byte{byte(i * 10)}
		if err := bw.Write(key, value); err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
	}
	_ = bw.Close(true)

	// Verify all entries
	for i := 0; i < 9; i++ {
		key := []byte{byte(i)}
		expected := []byte{byte(i * 10)}
		retrieved, err := db.Get(key)
		if err != nil {
			t.Errorf("Failed to get entry %d: %v", i, err)
			continue
		}
		if !bytes.Equal(retrieved, expected) {
			t.Errorf("Entry %d: expected %v, got %v", i, expected, retrieved)
		}
	}
}

// =============================================================================
// BatchWriter.Flush Tests
// =============================================================================

func TestBatchWriter_Flush(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-flush-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 1000)

	if err := bw.Write([]byte("flushkey"), []byte("flushvalue")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := bw.Flush(true); err != nil {
		t.Errorf("Flush should not return error: %v", err)
	}

	retrieved, err := db.Get([]byte("flushkey"))
	if err != nil {
		t.Fatalf("Failed to get value after flush: %v", err)
	}
	if string(retrieved) != "flushvalue" {
		t.Errorf("Expected 'flushvalue', got '%s'", string(retrieved))
	}

	_ = bw.Close(false)
}

func TestBatchWriter_FlushSync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-flushsync-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 1000)

	if err := bw.Write([]byte("synckey"), []byte("syncvalue")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Test sync flush (true)
	if err := bw.Flush(true); err != nil {
		t.Errorf("Sync flush should not return error: %v", err)
	}

	_ = bw.Close(false)
}

func TestBatchWriter_FlushAsync(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-flushasync-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 1000)

	if err := bw.Write([]byte("asynckey"), []byte("asyncvalue")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Test async flush (false)
	if err := bw.Flush(false); err != nil {
		t.Errorf("Async flush should not return error: %v", err)
	}

	_ = bw.Close(false)
}

// =============================================================================
// BatchWriter.Close Tests
// =============================================================================

func TestBatchWriter_Close(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-bwclose-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 1000)

	if err := bw.Write([]byte("closekey"), []byte("closevalue")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := bw.Close(true); err != nil {
		t.Errorf("Close should not return error: %v", err)
	}

	retrieved, err := db.Get([]byte("closekey"))
	if err != nil {
		t.Fatalf("Failed to get value after close: %v", err)
	}
	if string(retrieved) != "closevalue" {
		t.Errorf("Expected 'closevalue', got '%s'", string(retrieved))
	}
}

func TestBatchWriter_CloseEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-closeempty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 100)

	// Close without any writes
	if err := bw.Close(true); err != nil {
		t.Errorf("Close on empty batch should not return error: %v", err)
	}
}

// =============================================================================
// BatchWriter Large Dataset Tests
// =============================================================================

func TestBatchWriter_LargeDataset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pebble-test-large-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 50)

	// Write 200 entries (triggers 4 auto-flushes)
	for i := 0; i < 200; i++ {
		key := []byte{byte(i / 256), byte(i % 256)}
		value := []byte{byte(i % 256)}
		if err := bw.Write(key, value); err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
	}
	_ = bw.Close(true)

	// Verify sample entries
	checkIndices := []int{0, 49, 100, 150, 199}
	for _, i := range checkIndices {
		key := []byte{byte(i / 256), byte(i % 256)}
		expected := []byte{byte(i % 256)}
		retrieved, err := db.Get(key)
		if err != nil {
			t.Errorf("Failed to get entry %d: %v", i, err)
			continue
		}
		if !bytes.Equal(retrieved, expected) {
			t.Errorf("Entry %d: expected %v, got %v", i, expected, retrieved)
		}
	}
}

// =============================================================================
// BatchWriter Benchmarks
// =============================================================================

func BenchmarkBatchWriter_Write(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-bw-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 100)
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
		_ = bw.Write(key, value)
	}
	b.StopTimer()
	_ = bw.Close(false)
}

func BenchmarkBatchWriter_WriteSmallBatch(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-bwsmall-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 10)
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
		_ = bw.Write(key, value)
	}
	b.StopTimer()
	_ = bw.Close(false)
}

func BenchmarkBatchWriter_WriteLargeBatch(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "pebble-bench-bwlarge-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	db, err := OpenDB(tmpDir)
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	bw := NewBatchWriter(db.GetRawDB(), 1000)
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
		_ = bw.Write(key, value)
	}
	b.StopTimer()
	_ = bw.Close(false)
}
