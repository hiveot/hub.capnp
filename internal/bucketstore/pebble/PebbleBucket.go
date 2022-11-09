package pebble

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
)

// PebbleBucket represents a transactional bucket using Pebble
// Buckets are not supported in Pebble so these are simulated by prefixing all keys with "{bucketID}$"
// Each write operation is its own transaction.
type PebbleBucket struct {
	db           *pebble.DB
	bucketPrefix string // the prefix to apply to all keys of this bucket
	bucketID     string
	clientID     string
	closed       bool
}

// Close the bucket
func (bucket *PebbleBucket) Close() (err error) {
	if bucket.closed {
		err = fmt.Errorf("bucket '%s' of client '%s' is already closed", bucket.bucketID, bucket.clientID)
	}
	bucket.closed = true
	return err
}

// Commit changes to the bucket
//func (bucket *PebbleBucket) Commit() (err error) {
//	// this is just for error detection
//	if !bucket.writable {
//		err = fmt.Errorf("cant commit as bucket '%s' of client '%s' is not writable",
//			bucket.bucketID, bucket.clientID)
//	}
//	if bucket.closed {
//		err = fmt.Errorf("bucket '%s' of client '%s' is already closed", bucket.bucketID, bucket.clientID)
//	}
//	return err
//}

// Cursor provides an iterator for the bucket using a pebble iterator with prefix bounds
func (bucket *PebbleBucket) Cursor() (bucketstore.IBucketCursor, error) {
	// bucket prefix is {bucketID}$
	// range bounds end at {bucketID}@
	opts := &pebble.IterOptions{
		LowerBound:      []byte(bucket.bucketPrefix),
		UpperBound:      []byte(bucket.bucketID + "@"), // this key never exists
		TableFilter:     nil,
		PointKeyFilters: nil,
		RangeKeyFilters: nil,
		KeyTypes:        0,
		RangeKeyMasking: pebble.RangeKeyMasking{
			Suffix: nil,
			Filter: nil,
		},
		OnlyReadGuaranteedDurable: false,
		UseL6Filters:              false,
	}
	bucketIterator := bucket.db.NewIter(opts)
	cursor := NewPebbleCursor(bucket.clientID, bucket.bucketID, bucket.bucketPrefix, bucketIterator)
	return cursor, nil
}

// Delete removes the key-value pair from the bucket store
func (bucket *PebbleBucket) Delete(key string) (err error) {
	bucketKey := bucket.bucketPrefix + key
	opts := &pebble.WriteOptions{}
	err = bucket.db.Delete([]byte(bucketKey), opts)
	return err
}

// Get returns the document for the given key
func (bucket *PebbleBucket) Get(key string) (doc []byte, err error) {
	bucketKey := bucket.bucketPrefix + key
	byteValue, closer, err := bucket.db.Get([]byte(bucketKey))
	if err == nil {
		doc = bytes.NewBuffer(byteValue).Bytes()
		err = closer.Close()
	} else if errors.Is(err, pebble.ErrNotFound) {
		// return doc nil if not found
		err = nil
		doc = nil
	}
	return doc, err
}

// GetMultiple returns a batch of documents with existing keys
func (bucket *PebbleBucket) GetMultiple(keys []string) (docs map[string][]byte, err error) {

	docs = make(map[string][]byte)
	batch := bucket.db.NewIndexedBatch()
	for _, key := range keys {
		bucketKey := bucket.bucketPrefix + key
		value, closer, err2 := batch.Get([]byte(bucketKey))
		if err2 == nil {
			docs[key] = bytes.NewBuffer(value).Bytes()
			closer.Close()
		}
	}
	err = batch.Close()
	return docs, err
}

// ID returns the bucket ID
func (bucket *PebbleBucket) ID() string {
	return bucket.bucketID
}

// Set sets a document with the given key
func (bucket *PebbleBucket) Set(key string, doc []byte) error {
	if key == "" {
		err := fmt.Errorf("empty key '%s' for bucket '%s' and client '%s'",
			key, bucket.bucketID, bucket.clientID)
		return err
	}
	bucketKey := bucket.bucketPrefix + key
	opts := &pebble.WriteOptions{}
	err := bucket.db.Set([]byte(bucketKey), doc, opts)
	return err

}

// SetMultiple sets multiple documents in a batch update
func (bucket *PebbleBucket) SetMultiple(docs map[string][]byte) (err error) {

	batch := bucket.db.NewBatch()
	for key, value := range docs {
		bucketKey := bucket.bucketPrefix + key
		opts := &pebble.WriteOptions{}
		err = batch.Set([]byte(bucketKey), value, opts)
		if err != nil {
			logrus.Errorf("failed set multiple for client '%s: %s", bucket.clientID, err)
			_ = batch.Close()
			return err
		}
	}
	err = bucket.db.Apply(batch, &pebble.WriteOptions{})
	_ = batch.Close()
	return err
}

// NewPebbleBucket creates a new bucket
func NewPebbleBucket(clientID, bucketID string, pebbleDB *pebble.DB) *PebbleBucket {
	srv := &PebbleBucket{
		clientID:     clientID,
		bucketID:     bucketID,
		db:           pebbleDB,
		bucketPrefix: bucketID + "$",
	}
	return srv
}
