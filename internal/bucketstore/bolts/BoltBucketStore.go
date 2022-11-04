package bolts

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

// BoltBucketStore implements the IBucketStore API using the embedded bolt database
// This uses the BBolt package which is a derivative of BoltDB
type BoltBucketStore struct {
	// the underlying database
	boltDB *bbolt.DB
	// client this store is for. Intended for debugging and logging.
	clientID string
	// storePath with the location of the databse
	storePath string
}

// Close the store and flush changes to disk
func (bs *BoltBucketStore) Close() error {
	err := bs.boltDB.Close()
	return err
}

// Delete a key in the bucket
func (bs *BoltBucketStore) Delete(bucketID string, key string) error {

	tx, err := bs.boltDB.Begin(true)
	if err != nil {
		return err
	}
	bucket := tx.Bucket([]byte(bucketID))
	if bucket != nil {
		// only returns an error if the transaction was readonly, which it isn't
		_ = bucket.Delete([]byte(key))
	}
	err = tx.Commit()
	return err
}

// Get reads a document with the given key
func (bs *BoltBucketStore) Get(bucketID string, key string) (val []byte, err error) {

	tx, err := bs.boltDB.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	bucket := tx.Bucket([]byte(bucketID))
	if bucket == nil {
		err = fmt.Errorf("bucket '%s' for client '%s' doesn't exist", bucketID, bs.clientID)
		logrus.Info(err.Error())
		return nil, err
	}
	byteValue := bucket.Get([]byte(key))
	if byteValue == nil {
		err = fmt.Errorf("key '%s' in bucket '%s' for client '%s' doesn't exist", key, bucketID, bs.clientID)
		logrus.Info(err.Error())
		return nil, err
	}
	// byteValue is only valid within the transaction
	val = bytes.NewBuffer(byteValue).Bytes()
	return val, err
}

// GetMultiple returns a batch of documents with existing keys
func (bs *BoltBucketStore) GetMultiple(bucketID string, keys []string) (docs map[string][]byte, err error) {
	docs = make(map[string][]byte)
	tx, err := bs.boltDB.Begin(false)
	if err != nil {
		return docs, err
	}
	defer tx.Rollback()
	bucket := tx.Bucket([]byte(bucketID))
	if bucket == nil {
		err = fmt.Errorf("bucket '%s' for client '%s' doesn't exist", bucketID, bs.clientID)
		logrus.Info(err.Error())
		return docs, err
	}
	for _, key := range keys {
		byteValue := bucket.Get([]byte(key))
		// simply ignore non existing keys and log as info
		if byteValue == nil {
			logrus.Infof("key '%s' in bucket '%s' for client '%s' doesn't exist", key, bucketID, bs.clientID)
		} else {
			// byteValue is only valid within the transaction
			val := bytes.NewBuffer(byteValue).Bytes()
			docs[key] = val
		}
	}
	return docs, err
}

// Open the store
func (bs *BoltBucketStore) Open() (err error) {
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

// Set writes a document with the given key
func (bs *BoltBucketStore) Set(bucketID string, key string, value []byte) error {

	err := bs.boltDB.Update(func(tx *bbolt.Tx) error {
		var err error
		bucket := tx.Bucket([]byte(bucketID))
		if bucket == nil {
			bucket, err = tx.CreateBucketIfNotExists([]byte(bucketID))
		}
		if err == nil {
			err = bucket.Put([]byte(key), value)
		}
		return err
	})

	//tx, err := bs.boltDB.Begin(true)
	//if err != nil {
	//	return err
	//}
	//bucket, err := tx.CreateBucketIfNotExists([]byte(bucketID))
	//if err == nil {
	//	err = bucket.Put([]byte(key), value)
	//}
	//err = tx.Commit()
	return err
}

// SetMultiple writes a multiple documents in a single transaction
func (bs *BoltBucketStore) SetMultiple(bucketID string, docs map[string][]byte) error {

	tx, err := bs.boltDB.Begin(true)
	if err != nil {
		return err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte(bucketID))
	if err == nil {
		for key, value := range docs {
			err = bucket.Put([]byte(key), value)
			if err != nil {
				logrus.Errorf("invalid key or value for key '%s': %v", key, err)
				_ = tx.Rollback()
				return err
			}
		}
	}
	err = tx.Commit()
	return err
}

// NewBoltBucketStore creates a state storage server instance.
//  storePath is the file holding the database
func NewBoltBucketStore(clientID, storePath string) *BoltBucketStore {
	srv := &BoltBucketStore{
		clientID:  clientID,
		storePath: storePath,
	}
	return srv
}
