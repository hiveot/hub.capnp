package bolts

import (
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"github.com/hiveot/hub/internal/bucketstore"
)

// BoltStore implements the IBucketStore API using the embedded bolt database
// This uses the BBolt package which is a derivative of BoltDB
// Estimates using a i5-4570S @2.90GHz cpu.
//
// Create&commit write bucket, no data changes
//   Dataset 1K,        3.7 ms/op
//   Dataset 10K,       3.7 ms/op
//   Dataset 100K      12   ms/op
//   Dataset 1M        45   ms/op
//
// Create&close read-only bucket
//   Dataset 1K,        0.8 us/op
//   Dataset 10K,       0.8 us/op
//   Dataset 100K       0.8 us/op
//   Dataset 1M         0.7 us/op
//
// Get read-bucket 1 record
//   Dataset 1K,        1.3 us/op
//   Dataset 10K,       1.3 us/op
//   Dataset 100K       1.4 us/op
//   Dataset 1M         1.2 us/op
//
// Set write-bucket 1 record, 100 byte records
//   Dataset 1K,          4.7 ms/op
//   Dataset 10K,         5.2 ms/op
//   Dataset 100K        12   ms/op
//   Dataset 1M          41   ms/op
//   Dataset 10M         62   ms/op
//
// Seek               read-only bucket      writable bucket/rollback
//   Dataset 1K,        1.3 us/op                 2 us/op
//   Dataset 10K,       1.3 us/op                 2 us/op
//   Dataset 100K       1.4 us/op               120 us/op
//   Dataset 1M         1.3 us/op              1161 us/op
//

//
type BoltStore struct {
	// the underlying database
	boltDB *bbolt.DB
	// client this store is for. Intended for debugging and logging.
	clientID string
	// storePath with the location of the database
	storePath string
}

// Close the store and flush changes to disk
func (store *BoltStore) Close() error {
	logrus.Infof("closing store for client '%s'", store.clientID)
	err := store.boltDB.Close()
	return err
}

// GetReadBucket returns a bucket to use for storage
//  returns a bucket or nil if it doesn't exist
func (store *BoltStore) GetReadBucket(bucketID string) (bucket bucketstore.IBucket) {

	var boltBucket *bbolt.Bucket
	tx, err := store.boltDB.Begin(false)
	if err != nil {
		// what to do here?
		panic("unable to start transaction. unable to recover")
	}
	boltBucket = tx.Bucket([]byte(bucketID))
	if boltBucket == nil {
		tx.Rollback()
		return nil
	}
	logrus.Infof("Opening bucket '%s' of client '%s", bucketID, store.clientID)
	bucket = NewBoltBucket(store.clientID, bucketID, boltBucket)
	return bucket
}

// GetWriteBucket returns a bucket to use for writing to storage.
// If the bucket doesn't exist it is created.
//  only a single writable bucket can be used at the same time. See bbolt doc for detail
//  returns a bucket or nil if it doesn't exist and can't be created
func (store *BoltStore) GetWriteBucket(bucketID string) (bucket bucketstore.IBucket) {

	var boltBucket *bbolt.Bucket
	tx, err := store.boltDB.Begin(true)
	if err != nil {
		// what to do here?
		panic("unable to start transaction. unable to recover")
	}
	boltBucket, err = tx.CreateBucketIfNotExists([]byte(bucketID))
	if err != nil {
		logrus.Errorf("unexpected error creating bucket '%s' for client '%s': %s", bucketID, store.clientID, err)
		return nil
	}
	//logrus.Infof("Opening bucket '%s' of client '%s", bucketID, store.clientID)
	bucket = NewBoltBucket(store.clientID, bucketID, boltBucket)
	return bucket
}

// Open the store
func (store *BoltStore) Open() (err error) {
	logrus.Infof("Opening bboltDB store for client %s", store.clientID)

	options := &bbolt.Options{
		Timeout:        10,                    // wait max 1 sec for a file lock
		NoFreelistSync: false,                 // consider true for increased write performance
		FreelistType:   bbolt.FreelistMapType, // performant even for large DB
		//InitialMmapSize: 0,  // when is this useful to set?
	}
	store.boltDB, err = bbolt.Open(store.storePath, 0600, options)

	if err != nil {
		logrus.Errorf("Error opening bboltDB for client %s: %s", store.clientID, err)
	}
	return err
}

// NewBoltStore creates a state storage server instance.
//  storePath is the file holding the database
func NewBoltStore(clientID, storePath string) *BoltStore {
	srv := &BoltStore{
		clientID:  clientID,
		storePath: storePath,
	}
	return srv
}
