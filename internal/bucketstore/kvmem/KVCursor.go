package kvmem

import (
	"sort"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
)

type KVBucketCursor struct {
	orderedKeys []string
	bucket      bucketstore.IBucket
	index       int
}

// First moves the cursor to the first item
func (cursor *KVBucketCursor) First() (key string, value []byte) {
	cursor.index = 0
	key = cursor.orderedKeys[cursor.index]
	value, _ = cursor.bucket.Get(key)
	return
}

// Last moves the cursor to the last item
func (cursor *KVBucketCursor) Last() (key string, value []byte) {
	cursor.index = len(cursor.orderedKeys) - 1
	if cursor.index >= 0 {
		key = cursor.orderedKeys[cursor.index]
		value, _ = cursor.bucket.Get(key)
	}
	return
}

// Next increases the cursor position and return the next key and value
// If the end is reached the returned key is empty
func (cursor *KVBucketCursor) Next() (key string, value []byte) {
	if cursor.index < len(cursor.orderedKeys) {
		cursor.index++
	}
	if cursor.index >= len(cursor.orderedKeys) {
		key = ""
		value = nil
	} else {
		key = cursor.orderedKeys[cursor.index]
		value, _ = cursor.bucket.Get(key)
	}
	return key, value
}

// Prev decreases the cursor position and return the previous key and value
// If the head is reached the returned key is empty
func (cursor *KVBucketCursor) Prev() (key string, value []byte) {
	if cursor.index >= 0 {
		cursor.index--
	}
	if cursor.index < 0 || len(cursor.orderedKeys) == 0 {
		key = ""
		value = nil
	} else {
		key = cursor.orderedKeys[cursor.index]
		value, _ = cursor.bucket.Get(key)
	}
	return key, value
}

// Release the cursor capability
func (cursor *KVBucketCursor) Release() {
	cursor.orderedKeys = nil
}

// Seek positions the cursor at the given searchKey.
// This implementation is brute force. It generates a sorted list of orderedKeys for use by the cursor.
// This should still be fast enough for most cases. (test shows around 500msec for 1 million orderedKeys).
//
//  BucketID to seach for. Returns and error if the bucket is not found
//  key is the starting point. If key doesn't exist, the next closest key will be used.
//
// This returns a cursor with Next() and Prev() iterators
func (cursor *KVBucketCursor) Seek(searchKey string) (key string, value []byte) {
	var err error
	cursor.index = sort.SearchStrings(cursor.orderedKeys, searchKey)
	if cursor.index < len(cursor.orderedKeys) {
		key = cursor.orderedKeys[cursor.index]
		value, err = cursor.bucket.Get(key)
		if err != nil {
			logrus.Error(err)
		}
	}
	return key, value
}

// NewKVCursor create a new bucket cursor for the KV store.
// Cursor.Close() must be called to release the resources.
//
//  bucket is the bucket holding the data
//  orderedKeys is a snapshot of the keys in ascending order
func NewKVCursor(bucket bucketstore.IBucket, orderedKeys []string) *KVBucketCursor {
	cursor := &KVBucketCursor{
		bucket:      bucket,
		orderedKeys: orderedKeys,
		index:       0,
	}
	return cursor
}
