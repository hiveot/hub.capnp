package kvbtree

import (
	"github.com/tidwall/btree"

	"github.com/hiveot/hub/pkg/bucketstore"
)

type KVBTreeCursor struct {
	bucket bucketstore.IBucket
	kviter btree.MapIter[string, []byte]
}

// First moves the cursor to the first item
func (cursor *KVBTreeCursor) First() (key string, value []byte) {
	cursor.kviter.First()
	key = cursor.kviter.Key()
	value = cursor.kviter.Value()
	return
}

// Last moves the cursor to the last item
func (cursor *KVBTreeCursor) Last() (key string, value []byte) {
	cursor.kviter.Last()
	key = cursor.kviter.Key()
	value = cursor.kviter.Value()
	return
}

// Next increases the cursor position and return the next key and value
// If the end is reached the returned key is empty
func (cursor *KVBTreeCursor) Next() (key string, value []byte) {
	endreached := !cursor.kviter.Next()
	if endreached {
		return "", nil
	}
	key = cursor.kviter.Key()
	value = cursor.kviter.Value()
	return key, value
}

// NextN increases the cursor position N times and return the encountered key-value pairs
func (cursor *KVBTreeCursor) NextN(steps uint) (docs map[string][]byte, endReached bool) {
	docs = make(map[string][]byte)
	for i := uint(0); i < steps; i++ {
		endReached = !cursor.kviter.Next()
		if endReached {
			break
		}
		key := cursor.kviter.Key()
		value := cursor.kviter.Value()
		docs[key] = value
	}
	return docs, endReached
}

// Prev decreases the cursor position and return the previous key and value
// If the head is reached the returned key is empty
func (cursor *KVBTreeCursor) Prev() (key string, value []byte) {
	startreached := !cursor.kviter.Prev()
	if startreached {
		return "", nil
	}
	key = cursor.kviter.Key()
	value = cursor.kviter.Value()
	return key, value
}

// PrevN decreases the cursor position N times and return the encountered key-value pairs
func (cursor *KVBTreeCursor) PrevN(steps uint) (docs map[string][]byte, beginReached bool) {
	docs = make(map[string][]byte)
	for i := uint(0); i < steps; i++ {
		beginReached = !cursor.kviter.Prev()
		if beginReached {
			break
		}
		key := cursor.kviter.Key()
		value := cursor.kviter.Value()
		docs[key] = value
	}
	return docs, beginReached
}

// Release the cursor capability
func (cursor *KVBTreeCursor) Release() {
}

// Seek positions the cursor at the given searchKey.
// This implementation is brute force. It generates a sorted list of orderedKeys for use by the cursor.
// This should still be fast enough for most cases. (test shows around 500msec for 1 million orderedKeys).
//
//	BucketID to seach for. Returns and error if the bucket is not found
//	key is the starting point. If key doesn't exist, the next closest key will be used.
//
// This returns a cursor with Next() and Prev() iterators
func (cursor *KVBTreeCursor) Seek(searchKey string) (key string, value []byte) {
	//var err error
	cursor.kviter.Seek(searchKey)
	key = cursor.kviter.Key()
	value = cursor.kviter.Value()
	return key, value
}

// NewKVCursor create a new bucket cursor for the KV store.
// Cursor.Close() must be called to release the resources.
//
//	bucket is the bucket holding the data
//	orderedKeys is a snapshot of the keys in ascending order
//
// func NewKVCursor(bucket bucketstore.IBucket, orderedKeys []string, kvbtree btree.Map[string, []byte]) *KVBTreeCursor {
func NewKVCursor(bucket bucketstore.IBucket, kvIter btree.MapIter[string, []byte]) *KVBTreeCursor {
	cursor := &KVBTreeCursor{
		bucket: bucket,
		kviter: kvIter,
	}
	return cursor
}
