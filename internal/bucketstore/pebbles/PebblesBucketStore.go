package pebbles

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/cockroachdb/pebble"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type PebbleBucketStore struct {
	clientID       string
	storeDirectory string
	db             *pebble.DB
}

func (ps *PebbleBucketStore) Close() error {
	err := ps.db.Close()
	return err
}

// Delete removes the key-value pair from the bucket store
func (ps *PebbleBucketStore) Delete(bucketID string, key string) (err error) {
	valueKey := fmt.Sprintf("%s.%s", bucketID, key)
	opts := &pebble.WriteOptions{}
	err = ps.db.Delete([]byte(valueKey), opts)
	return err
}

// Get returns the document for the given key
func (ps *PebbleBucketStore) Get(bucketID string, key string) (doc []byte, err error) {
	valueKey := fmt.Sprintf("%s.%s", bucketID, key)
	byteValue, closer, err := ps.db.Get([]byte(valueKey))
	if err == nil {
		doc = bytes.NewBuffer(byteValue).Bytes()
		err = closer.Close()
	}
	return doc, err
}

// GetMultiple returns a batch of documents with existing keys
func (ps *PebbleBucketStore) GetMultiple(
	bucketID string, keys []string) (docs map[string][]byte, err error) {
	docs = make(map[string][]byte)
	batch := ps.db.NewIndexedBatch()
	for _, key := range keys {
		valueKey := fmt.Sprintf("%s.%s", bucketID, key)
		value, closer, err2 := batch.Get([]byte(valueKey))
		if err2 == nil {
			docs[key] = bytes.NewBuffer(value).Bytes()
			closer.Close()
		}
	}
	err = batch.Close()
	return docs, err
}

// Open the store
func (ps *PebbleBucketStore) Open() (err error) {
	options := &pebble.Options{}
	// pebble.Open will panic if the store directory is readonly, so check ahead to return an error
	stat, err := os.Stat(ps.storeDirectory)
	// if the path exists, it must be a directory
	if err == nil {
		if !stat.IsDir() {
			err = fmt.Errorf("can't open store. '%s' is not a directory", ps.storeDirectory)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// if the path doesn't exist, create a directory with mode 0700
		err = os.MkdirAll(ps.storeDirectory, 0700)
	}
	// path must be writable to avoid a panic
	if err == nil {
		err = unix.Access(ps.storeDirectory, unix.W_OK)
	}
	if err == nil {
		ps.db, err = pebble.Open(ps.storeDirectory, options)
	}
	return err
}

// Seek provides an iterator starting at a key
//func (ps *PebbleBucketStore) Seek(key string) IBucketCursor{
//}

// Set sets a document with the given key
func (ps *PebbleBucketStore) Set(bucketID string, key string, doc []byte) error {
	if bucketID == "" || key == "" {
		err := fmt.Errorf("empty bucketID '%s' or key '%s'", bucketID, key)
		return err
	}
	valueKey := fmt.Sprintf("%s.%s", bucketID, key)
	opts := &pebble.WriteOptions{}
	err := ps.db.Set([]byte(valueKey), []byte(doc), opts)
	return err

}

// SetMultiple sets multiple documents in a batch update
func (ps *PebbleBucketStore) SetMultiple(bucketID string, docs map[string][]byte) (err error) {
	batch := ps.db.NewBatch()
	for key, value := range docs {
		valueKey := fmt.Sprintf("%s.%s", bucketID, key)
		opts := &pebble.WriteOptions{}
		err = ps.db.Set([]byte(valueKey), value, opts)
		if err != nil {
			logrus.Errorf("failed set multiple for client '%s: %s", ps.clientID, err)
			_ = batch.Close()
			return err
		}
	}
	err = ps.db.Apply(batch, &pebble.WriteOptions{})
	_ = batch.Close()
	return err
}

// NewPebbleBucketStore creates a storage database with bucket support.
//  clientID that owns the database
//  storeDirectory is the directory (not file) holding the database
func NewPebbleBucketStore(clientID, storeDirectory string) *PebbleBucketStore {
	srv := &PebbleBucketStore{
		clientID:       clientID,
		storeDirectory: storeDirectory,
	}
	return srv
}
