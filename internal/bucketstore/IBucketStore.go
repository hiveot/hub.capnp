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

// IBucketCursor provides the prev/next cursor on a range
type IBucketCursor interface {
	// Key provides the current cursor key
	Key() string
	// Value provides the current cursor value
	Value() []byte

	// Close the iterator
	Close()
	// Prev iterations to the previous key from the current cursor
	Prev() (key string, value []byte)
	// Next iterates to the next key from the current cursor
	Next() (key string, value []byte)
}

// IBucketStore defines the interface to a simple to use key-value embedded bucket stores.
// - organize data in buckets
// - get/set single or multiple key/value pairs
// - delete key/value
// - key range scan (todo)
// - forward iteration from key (todo)
// - reverse iteration from key (todo)
// Various implementations are available to the services to use.
// Pipelining
//
type IBucketStore interface {
	// Close the store and release its resources
	Close() error

	// Delete removes the key-value pair from the bucket store
	// Returns nil if the key is deleted or doesn't exist.
	// Returns an error if the key cannot be deleted.
	Delete(bucketID string, key string) (err error)

	// Get returns the document for the given key
	// Returns 'found' if the key exists in the bucket
	// Returns an error if the database cannot be read
	Get(bucketID string, key string) (doc []byte, found bool, err error)

	// GetMultiple returns a batch of documents with existing keys
	// if a key does not exist it will not be included in the result.
	// An error is return if the database cannot be read.
	GetMultiple(bucketID string, keys []string) (docs map[string][]byte, err error)

	// Keys returns a list of document keys in the store
	//Keys(ctx context.Context) (keys []string, err error)

	// Open the store
	Open() error

	// Seek provides an iterator starting at a key
	Seek(bucketID, key string) (cursor IBucketCursor, err error)

	// Set sets a document with the given key
	// An error is returned if either the bucketID or the key is empty
	Set(bucketID string, key string, doc []byte) error

	// SetMultiple sets multiple documents in a batch update
	// If the transaction fails an error is returned and no changes are made.
	SetMultiple(bucketID string, docs map[string][]byte) (err error)

	// Status returns the application state status
	//Status() BucketStoreStatus
}
