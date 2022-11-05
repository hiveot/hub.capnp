package bolts

import (
	"bytes"

	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"github.com/hiveot/hub/internal/bucketstore"
)

// BoltBucket implements the IBucket API using the embedded bolt database
type BoltBucket struct {
	// the underlying bucket
	bucket *bbolt.Bucket
	// client this bucket is for. Intended for debugging and logging.
	clientID string
	// ID of the bucket. Intended for debugging and logging.
	bucketID string
}

// Close the bucket and the transaction that created it
// if commit is false then close with rollback instead of commit
func (bb *BoltBucket) Close(commit bool) (err error) {
	//logrus.Infof("Closing bucket '%s' of client '%s", bb.bucketID, bb.clientID)
	if commit {
		err = bb.bucket.Tx().Commit()
	} else {
		err = bb.bucket.Tx().Rollback()
	}
	return err
}

// Cursor returns a new cursor for iterating the bucket.
// The cursor MUST be closed after use to release its memory.
//
// This returns a cursor with Next() and Prev() iterators
func (bb *BoltBucket) Cursor() (cursor bucketstore.IBucketCursor) {

	cursor = NewBBoltCursor(bb.bucket)
	return cursor
}

// Delete a key in the bucket
func (bb *BoltBucket) Delete(key string) (err error) {
	err = bb.bucket.Delete([]byte(key))
	return err
}

// Get reads a document with the given key
// returns nil if the key doesn't exist
func (bb *BoltBucket) Get(key string) (val []byte, err error) {

	byteValue := bb.bucket.Get([]byte(key))
	if byteValue == nil {
		//err = fmt.Errorf("key '%s' in bucket '%s' for client '%s' doesn't exist", key, bb.bucketID, bb.clientID)
		//logrus.Info(err.Error())
		return nil, nil
	}
	// byteValue is only valid within the transaction
	val = bytes.NewBuffer(byteValue).Bytes()
	return val, err
}

// GetMultiple returns a batch of documents with existing keys
func (bb *BoltBucket) GetMultiple(keys []string) (docs map[string][]byte, err error) {
	docs = make(map[string][]byte)

	for _, key := range keys {
		byteValue := bb.bucket.Get([]byte(key))
		// simply ignore non existing keys and log as info
		if byteValue == nil {
			//logrus.Infof("key '%s' in bucket '%s' for client '%s' doesn't exist", key, bb.bucketID, bb.clientID)
		} else {
			// byteValue is only valid within the transaction
			val := bytes.NewBuffer(byteValue).Bytes()
			docs[key] = val
		}
	}
	return docs, err
}

// ID returns the bucket's ID
func (bb *BoltBucket) ID() string {
	return bb.bucketID
}

// Set writes a document with the given key
func (bb *BoltBucket) Set(key string, value []byte) (err error) {

	err = bb.bucket.Put([]byte(key), value)
	return err
}

// SetMultiple writes a multiple documents in a single transaction
// This returns an error as soon as an invalid key is encountered.
// Cancel this bucket with Close(false) if this returns an error.
func (bb *BoltBucket) SetMultiple(docs map[string][]byte) (err error) {
	logrus.Infof("%d docs", len(docs))
	for key, value := range docs {
		err = bb.bucket.Put([]byte(key), value)
		if err != nil {
			logrus.Errorf("error put client '%s' value for key '%s' in bucket '%s': %v", bb.clientID, key, bb.bucketID, err)
			//_ = bb.bucket.Tx().Rollback()
			return err
		}
	}
	return err
}

// NewBoltBucket creates a new bucket
//  storePath is the file holding the database
func NewBoltBucket(clientID, bucketID string, boltBucket *bbolt.Bucket) *BoltBucket {
	srv := &BoltBucket{
		bucket:   boltBucket,
		clientID: clientID,
		bucketID: bucketID,
	}
	return srv
}
