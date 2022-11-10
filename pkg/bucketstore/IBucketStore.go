// Package bucketstore is a storage library for use by the services.
// The bucket store is primarily used by the state service which provides this store as a service to multiple clients.
// This package defines an API to use the store with several implementations.
package bucketstore

// Available embedded bucket store implementations with low memory overhead
const (
	BackendKVStore = "kvbtree" // fastest and best for small to medium amounts of data (dependent on available memory)
	BackendBBolt   = "bbolt"   // slow on writes but otherwise a good choice
	BackendPebble  = "pebble"  // a good middle ground between performance and memory
	// for consideration
	// badger: takes a lot of memory; some concerns about stability (based on comments)
	// encoding/gob encoder for serialization: interesting for kvmem serialization
	// akrylysov/pogreb: abandoned?
)

// BucketStoreStatus of the store
type BucketStoreStatus struct {
	// Number of buckets in the store
	NrBuckets int
	// Number of keys in the store
	NrKeys int64
	//
	MemSize int64
}

// IBucketStore defines the interface to a simple key-value embedded bucket store.
// Based on the boltDB API, this:
//   - organizes data into buckets
//   - open/close buckets as a transaction, if transactions are available
//   - get/set single or multiple key/value pairs
//   - delete key/value
//   - cursor based seek and iteration
//     Streaming data into a bucket is not supported
//     Various implementations are available to the services to use.
type IBucketStore interface {
	// GetBucket returns a bucket to use.
	// This creates the bucket if it doesn't exist.
	// Use bucket.Close() to close the bucket and release its resources.
	GetBucket(bucketID string) (bucket IBucket)

	// Close the store and release its resources
	Close() error

	// Open the store
	Open() error

	// Status returns the application state status
	//Status() BucketStoreStatus
}

// IBucket defines the interface to a store key-value bucket
type IBucket interface {

	// Close the bucket and release its resources
	// If commit is true and transactions are support then this commits the transaction.
	// use false to rollback the transaction. For readonly buckets commit returns an error
	Close() error

	// Cursor creates a new bucket cursor for iterating the bucket
	// cursor.Close must be called after use to release any read transactions
	// returns an error the cursor cannot be created
	Cursor() (IBucketCursor, error)

	// Delete removes the key-value pair from the bucket store
	// Returns nil if the key is deleted or doesn't exist.
	// Returns an error if the key cannot be deleted.
	Delete(key string) (err error)

	// Get returns the document for the given key
	// Returns nil if the key isn't found in the bucket
	// Returns an error if the database cannot be read
	Get(key string) (value []byte, err error)

	// GetMultiple returns a batch of documents with existing keys
	// if a key does not exist it will not be included in the result.
	// An error is return if the database cannot be read.
	GetMultiple(keys []string) (keyValues map[string][]byte, err error)

	// ID returns the bucket's ID
	ID() string

	// Set sets a document with the given key
	// An error is returned if either the bucketID or the key is empty
	Set(key string, value []byte) error

	// SetMultiple sets multiple documents in a batch update
	// If the transaction fails an error is returned and no changes are made.
	SetMultiple(docs map[string][]byte) (err error)

	// Status returns the bucket status
	//Status() BucketStoreStatus
}

// IBucketCursor provides the prev/next cursor based iterator on a range
type IBucketCursor interface {

	// First positions the cursor at the first key in the ordered list
	First() (key string, value []byte)

	// Last positions the cursor at the last key in the ordered list
	Last() (key string, value []byte)

	// Next moves the cursor to the next key from the current cursor
	// First() or Seek must have been called first.
	Next() (key string, value []byte)

	// NextN moves the cursor to the next N places from the current cursor
	// and return a map with the N key-value pairs.
	// endReached is true if the iterator has reached the end.
	// Intended to speed up with batch iterations over rpc.
	NextN(steps uint) (docs map[string][]byte, endReached bool)

	// Prev moves the cursor to the previous key from the current cursor
	// Last() or Seek must have been called first.
	Prev() (key string, value []byte)

	// PrevN moves the cursor back N places from the current cursor
	// and return a map with the N key-value pairs.
	// beginReached is true if the iterator has reached the beginning
	// Intended to speed up with batch iterations over rpc.
	PrevN(steps uint) (docs map[string][]byte, startReached bool)

	// Release close the cursor and release its resources.
	// This invalidates all values obtained from the cursor
	Release()

	// Seek positions the cursor at the given searchKey and corresponding value.
	// If the key is not found, the next key is returned.
	// cursor.Close must be invoked after use in order to close any read transactions.
	Seek(searchKey string) (key string, value []byte)

	// Stream the content of the cursor to the provided function
	//Stream(cb func(key string, value []byte, done bool) error)
}
