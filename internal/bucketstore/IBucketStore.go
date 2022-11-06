package bucketstore

// Available embedded bucket store implementations with low memory overhead
const (
	BackendKVStore = "kvstore" // fast but limited to available memory
	BackendBBolt   = "bbolt"   // slow on writes but otherwise a good choice
	BackendPebble  = "pebble"  //
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

// IBucketStore defines the interface to a simple to use key-value embedded bucket stores.
// Based on the bbolt API.
// - organize data in buckets
// - get/set single or multiple key/value pairs
// - delete key/value
// - cursor based seek and iteration
//  Various implementations are available to the services to use.
//  Pipelining is not yet supported
//  Transactions are used when available
type IBucketStore interface {
	// GetReadBucket returns a read-only bucket to use.
	// This creates a transaction for reading bucket data from the database.
	// Use bucket.Close() to close the transaction.
	// Returns the bucket, or nil if the bucket does not exist. The Pebble implementation
	// returns an empty bucket if one doesn't exist.
	GetReadBucket(bucketID string) (bucket IBucket)

	// GetWriteBucket returns the bucket to use. It is created if it doesn't exist.
	// This creates a transaction for reading bucket data from the database.
	// Use bucket.Close() to close the transaction.
	// Returns the bucket, or nil in case of an internal database error
	GetWriteBucket(bucketID string) (bucket IBucket)

	// Close the store and release its resources
	Close() error

	// Open the store
	Open() error

	// Status returns the application state status
	//Status() BucketStoreStatus
}

// IBucketCursor provides the prev/next cursor on a range
type IBucketCursor interface {
	// Close the cursor and release its resources.
	// This invalidates all values obtained from the cursor
	Close() error

	// First positions the cursor at the first key in the ordered list
	First() (key string, value []byte)

	// Last positions the cursor at the last key in the ordered list
	Last() (key string, value []byte)

	// Next moves the cursor to the next key from the current cursor
	Next() (key string, value []byte)

	// Prev moves the cursor to the previous key from the current cursor
	Prev() (key string, value []byte)

	// Seek positions the cursor at the given searchKey and corresponding value.
	// If the key is not found, the next key is returned.
	// cursor.Close must be invoked after use in order to close any read transactions.
	Seek(searchKey string) (key string, value []byte)
}

// IBucket defines the interface to a store key-value bucket
type IBucket interface {

	// Close the bucket and release its resources
	// If commit is true and transactions are support then this commits the transaction.
	// use false to rollback the transaction. For readonly buckets commit returns an error
	Close(commit bool) error

	// Cursor creates a new bucket cursor for iterating the bucket
	// cursor.Close must be called after use to release any read transactions
	Cursor() IBucketCursor

	// Delete removes the key-value pair from the bucket store
	// Returns nil if the key is deleted or doesn't exist.
	// Returns an error if the key cannot be deleted.
	Delete(key string) (err error)

	// Get returns the document for the given key
	// Returns nil if the key isn't found in the bucket
	// Returns an error if the database cannot be read
	Get(key string) (doc []byte, err error)

	// GetMultiple returns a batch of documents with existing keys
	// if a key does not exist it will not be included in the result.
	// An error is return if the database cannot be read.
	GetMultiple(keys []string) (docs map[string][]byte, err error)

	// ID returns the bucket's ID
	ID() string

	// Set sets a document with the given key
	// An error is returned if either the bucketID or the key is empty
	Set(key string, doc []byte) error

	// SetMultiple sets multiple documents in a batch update
	// If the transaction fails an error is returned and no changes are made.
	SetMultiple(docs map[string][]byte) (err error)

	// Status returns the bucket status
	//Status() BucketStoreStatus
}
