package pebble

import (
	"errors"
	"fmt"
	"os"

	"github.com/cockroachdb/pebble"
	"golang.org/x/sys/unix"

	"github.com/hiveot/hub/internal/bucketstore"
)

// PebbleStore implements the IBucketStore API using the embedded CockroachDB pebble database
//
// The following benchmark are made using BucketBench_test.go
// Performance is stellar! Fast, efficient data storage and low memory usage compared to the others.
// Estimates are made using a i5-4570S @2.90GHz cpu. Document size is 100 bytes.
//
// Create&commit write bucket, no data changes  (fast since pebbles doesn't use transactions for this)
//   Dataset 1K,        0.1 us/op
//   Dataset 10K,       0.1 us/op
//   Dataset 100K       0.1 us/op
//   Dataset 1M         0.1 us/op
//
// Create&close read-only bucket  (fast since pebbles doesn't use transactions for this)
//   Dataset 1K,        0.1 us/op
//   Dataset 10K,       0.1 us/op
//   Dataset 100K       0.1 us/op
//   Dataset 1M         0.1 us/op
//
// Get read-bucket 1 record
//   Dataset 1K,       14 us/op
//   Dataset 10K,       7 us/op
//   Dataset 100K       5 us/op
//   Dataset 1M        20 us/op
//
// Set write-bucket 1 record
//   Dataset 1K,         3.6 us/op
//   Dataset 10K,        2.8 us/op
//   Dataset 100K        2.8 us/op
//   Dataset 1M          5.5 us/op
//   Dataset 10M        27   us/op
//
// Seek, 1 record
//   Dataset 1K,        20 us/op
//   Dataset 10K,       76 us/op
//   Dataset 100K      123 us/op
//   Dataset 1M        175 us/op
//   Dataset 10M       144 us/op
//
// See https://pkg.go.dev/github.com/cockroachdb/pebble for Pebble's documentation.
//
// TODO: as this is a very crude implementation, it is lacking in many ways:
// 1. Add transaction support using batch for buckets. bucket.Close(false) should rollback.
// 2. Better error checking
// 3. Better test cases that really test proper values and edge cases
// 4. Does seek and range iteration behave correctly at boundaries?
// 5. ...
type PebbleStore struct {
	clientID       string
	storeDirectory string
	db             *pebble.DB
}

func (store *PebbleStore) Close() error {
	err := store.db.Close()
	return err
}

// GetReadBucket returns a read-only bucket
// Pebble doesn't support buckets so just use key prefixe. This implies a bucket always exists.
//  returns a bucket or nil if it doesn't exist
func (store *PebbleStore) GetReadBucket(bucketID string) (bucket bucketstore.IBucket) {
	bucket = NewPebbleBucket(store.clientID, bucketID, store.db, false)
	return bucket
}

// GetWriteBucket returns a writable bucket. A bucket is created if it doesn't exist.
func (store *PebbleStore) GetWriteBucket(bucketID string) (bucket bucketstore.IBucket) {
	pb := NewPebbleBucket(store.clientID, bucketID, store.db, true)
	return pb
}

// Open the store
func (store *PebbleStore) Open() (err error) {
	options := &pebble.Options{}
	// pebble.Open will panic if the store directory is readonly, so check ahead to return an error
	stat, err := os.Stat(store.storeDirectory)
	// if the path exists, it must be a directory
	if err == nil {
		if !stat.IsDir() {
			err = fmt.Errorf("can't open store. '%s' is not a directory", store.storeDirectory)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// if the path doesn't exist, create a directory with mode 0700
		err = os.MkdirAll(store.storeDirectory, 0700)
	}
	// path must be writable to avoid a panic
	if err == nil {
		err = unix.Access(store.storeDirectory, unix.W_OK)
	}
	if err == nil {
		store.db, err = pebble.Open(store.storeDirectory, options)
	}
	return err
}

// NewPebbleStore creates a storage database with bucket support.
//  clientID that owns the database
//  storeDirectory is the directory (not file) holding the database
func NewPebbleStore(clientID, storeDirectory string) *PebbleStore {
	srv := &PebbleStore{
		clientID:       clientID,
		storeDirectory: storeDirectory,
	}
	return srv
}
