package bolts

import (
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"github.com/hiveot/hub/pkg/bucketstore"
)

// BoltStore implements the IBucketStore API using the embedded bolt database
// This uses the BBolt package which is a derivative of BoltDB
// Estimates using a i5-4570S @2.90GHz cpu.
//
// Create & close bucket
//   Dataset 1K,        0.8 us/op
//   Dataset 10K,       0.8 us/op
//   Dataset 100K       0.8 us/op
//   Dataset 1M         0.7 us/op
//
// Bucket Get 1 record
//   Dataset 1K,        1.3 us/op
//   Dataset 10K,       1.3 us/op
//   Dataset 100K       1.4 us/op
//   Dataset 1M         1.3 us/op
//
// Bucket Set 1 record
//   Dataset 1K,          5.1 ms/op
//   Dataset 10K,         5.1 ms/op
//   Dataset 100K        12   ms/op
//   Dataset 1M          47   ms/op
//   Dataset 10M         62   ms/op
//
// Seek
//   Dataset 1K,        1.6 us/op
//   Dataset 10K,       1.8 us/op
//   Dataset 100K       2.0 us/op
//   Dataset 1M         1.7 us/op
//

type BoltStore struct {
	// the underlying database
	boltDB *bbolt.DB
	// client this store is for. Intended for debugging and logging.
	clientID string
	// storePath with the location of the database
	storePath string
	// for preventing deadlocks when closing the store. panic instead
	bucketRefCount int32
}

// Close the store and flush changes to disk
// Since boltDB locks transactions on close, this runs in the background.
// Close() returns before closing is completed.
func (store *BoltStore) Close() (err error) {
	br := atomic.LoadInt32(&store.bucketRefCount)
	logrus.Infof("closing store for client '%s'. Refcnt=%d", store.clientID, br)
	//close with wait until all transactions are completed ...
	// so it might hang forever if not all transactions are released.
	//err = store.boltDB.Close()
	err2 := make(chan error, 1)
	go func() {
		err2 <- store.boltDB.Close()
	}()
	select {
	case <-time.After(10 * time.Second):
		panic("BoltDB is not closing")
	case err = <-err2:
		return err
	}
	return err
}

// GetBucket returns a bucket to use for writing to storage.
// This does not yet create the bucket in the database until an operation takes place on the bucket.
func (store *BoltStore) GetBucket(bucketID string) (bucket bucketstore.IBucket) {

	//logrus.Infof("Opening bucket '%s' of client '%s", bucketID, store.clientID)
	bucket = NewBoltBucket(store.clientID, bucketID, store.boltDB, store.onBucketReleased)

	atomic.AddInt32(&store.bucketRefCount, 1)
	return bucket
}

// track bucket references
func (store *BoltStore) onBucketReleased(bucket bucketstore.IBucket) {
	atomic.AddInt32(&store.bucketRefCount, -1)
}

// Open the store
func (store *BoltStore) Open() (err error) {
	logrus.Infof("Opening bboltDB store for client %s", store.clientID)

	// make sure the folder exists
	storeDir := path.Dir(store.storePath)
	err = os.MkdirAll(storeDir, 0700)
	if err != nil {
		logrus.Errorf("Failed ensuring folder exists: %s", err)
	}

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
//
//	storePath is the file holding the database
func NewBoltStore(clientID, storePath string) *BoltStore {
	srv := &BoltStore{
		clientID:  clientID,
		storePath: storePath,
	}
	return srv
}
