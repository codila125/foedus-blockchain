package database

import (
	"fmt"
	"github.com/cockroachdb/pebble"
)

// BatchWriter handles batch writing to database with automatic flushing
type BatchWriter struct {
	rawDB      *pebble.DB
	batch      *pebble.Batch
	count      int
	flushEvery int
}

// NewBatchWriter creates a new batch writer with specified flush interval
func NewBatchWriter(rawDB *pebble.DB, flushEvery int) *BatchWriter {
	return &BatchWriter{
		rawDB:      rawDB,
		batch:      rawDB.NewBatch(),
		flushEvery: flushEvery,
	}
}

// Write adds a key-value pair to the batch and flushes if needed
func (bw *BatchWriter) Write(key, value []byte) error {
	if err := bw.batch.Set(key, value, nil); err != nil {
		return fmt.Errorf("failed to write to batch: %w", err)
	}

	bw.count++

	// Auto-flush every N writes
	if bw.count%bw.flushEvery == 0 {
		if err := bw.Flush(false); err != nil {
			return err
		}
	}

	return nil
}

// Flush applies the current batch to the database
func (bw *BatchWriter) Flush(sync bool) error {
	opts := &pebble.WriteOptions{Sync: sync}
	if err := bw.rawDB.Apply(bw.batch, opts); err != nil {
		return fmt.Errorf("failed to apply batch: %w", err)
	}
	bw.batch.Close()
	bw.batch = bw.rawDB.NewBatch()
	return nil
}

// Close flushes remaining data and closes the batch
func (bw *BatchWriter) Close(sync bool) error {
	if err := bw.Flush(sync); err != nil {
		return err
	}
	bw.batch.Close()
	return nil
}