package bolts

import (
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"github.com/hiveot/hub/internal/bucketstore"
)

// BoltStore implements the IBucketStore API using the embedded bolt database
// This uses the BBolt package which is a derivative of BoltDB
// Estimates using a data items of 1KB each, using a i5-4570S @2.90GHz cpu.
// note that these times are more than 10x faster than what a marshaller would take to serialize this data.
//
// Dataset size: 1K, read 1K
//   Write 1K records: 5.5 msec,     5500   usec/op   (one transaction per write)
//   Read 1K records: 0.8 msec,         0.8 usec/op
//   Create cursor: 0.006 msec,         6   usec/op
//   Iterate 1K records: 0.22 msec      0.2 usec/op
// Dataset size: 10K records (
//   Write 1 record: 5.7 msec        5700   usec/op (one transaction per write)
//   Read 10K records: 8.8 msec         0.9 usec/op
//   Create cursor: 0.01 msec          10   usec/op
//   Iterate 10K records: 2.2 msec      0.2 usec/op
// Dataset size: 100K records
//   Write 1 record: 11.9 msec      11900   usec/op (one transaction per write)
//   Read 100K records: 114 msec        1.1 usec/op
//   Create cursor: 0.01 msec          10   usec/op
//   Iterate 100K records: 17 msec     0.17 usec/op
// Dataset size: 1M records
//   Write 1 record: 38 msec (batch=1)          38000 usec/op   (* - very slow .. disk access?)
//   Write 1 record: 0.6 msec (batch=100)         600 usec/op
//   Write 1M records: 178 sec (batch=10000)      180 usec/op
//   Write 1M records: 116 sec (batch=50000)      116 usec/op
//   Read 1M records: 1400msec         1.4 usec/op
//   Create cursor: 0.01 msec            10 usec/op
//   Iterate 1M records: 144 msec      0.14 usec/op
type BoltStore struct {
	// the underlying database
	boltDB *bbolt.DB
	// client this store is for. Intended for debugging and logging.
	clientID string
	// storePath with the location of the database
	storePath string
}

// Close the store and flush changes to disk
func (bs *BoltStore) Close() error {
	logrus.Infof("closing store for client '%s'", bs.clientID)
	err := bs.boltDB.Close()
	return err
}

// GetReadBucket returns a bucket to use for storage
//  returns a bucket or nil if it doesn't exist
func (bs *BoltStore) GetReadBucket(bucketID string) (bucket bucketstore.IBucket) {

	var boltBucket *bbolt.Bucket
	tx, err := bs.boltDB.Begin(false)
	if err != nil {
		// what to do here?
		panic("unable to start transaction. unable to recover")
	}
	boltBucket = tx.Bucket([]byte(bucketID))
	if boltBucket == nil {
		tx.Rollback()
		return nil
	}
	logrus.Infof("Opening bucket '%s' of client '%s", bucketID, bs.clientID)
	bucket = NewBoltBucket(bs.clientID, bucketID, boltBucket)
	return bucket
}

// GetWriteBucket returns a bucket to use for writing to storage.
// If the bucket doesn't exist it is created.
//  only a single writable bucket can be used at the same time. See bbolt doc for detail
//  returns a bucket or nil if it doesn't exist and can't be created
func (bs *BoltStore) GetWriteBucket(bucketID string) (bucket bucketstore.IBucket) {

	var boltBucket *bbolt.Bucket
	tx, err := bs.boltDB.Begin(true)
	if err != nil {
		// what to do here?
		panic("unable to start transaction. unable to recover")
	}
	boltBucket, err = tx.CreateBucketIfNotExists([]byte(bucketID))
	if err != nil {
		logrus.Errorf("unexpected error creating bucket '%s' for client '%s': %s", bucketID, bs.clientID, err)
		return nil
	}
	//logrus.Infof("Opening bucket '%s' of client '%s", bucketID, bs.clientID)
	bucket = NewBoltBucket(bs.clientID, bucketID, boltBucket)
	return bucket
}

// Open the store
func (bs *BoltStore) Open() (err error) {
	logrus.Infof("Opening bboltDB store for client %s", bs.clientID)

	options := &bbolt.Options{
		Timeout:        10,                    // wait max 1 sec for a file lock
		NoFreelistSync: false,                 // consider true for increased write performance
		FreelistType:   bbolt.FreelistMapType, // performant even for large DB
		//InitialMmapSize: 0,  // when is this useful to set?
	}
	bs.boltDB, err = bbolt.Open(bs.storePath, 0600, options)

	if err != nil {
		logrus.Errorf("Error opening bboltDB for client %s: %s", bs.clientID, err)
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
