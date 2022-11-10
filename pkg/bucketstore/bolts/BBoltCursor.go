package bolts

import (
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

// BBoltCursor is a wrapper around the bbolt cursor to map to the IBucketCursor API and to
// ensure its transaction is released after the cursor is no longer used.
// This implements the IBucketCursor API
type BBoltCursor struct {
	bucket *bbolt.Bucket
	cursor *bbolt.Cursor
}

// First moves the cursor to the first item
func (bbc *BBoltCursor) First() (key string, value []byte) {
	if bbc.cursor == nil {
		return "", nil
	}
	k, v := bbc.cursor.First()
	return string(k), v
}

// Last moves the cursor to the last item
func (bbc *BBoltCursor) Last() (key string, value []byte) {
	if bbc.cursor == nil {
		return "", nil
	}
	k, v := bbc.cursor.Last()
	return string(k), v
}

// Next iterates to the next key from the current cursor
func (bbc *BBoltCursor) Next() (key string, value []byte) {
	if bbc.cursor == nil {
		return "", nil
	}
	k, v := bbc.cursor.Next()
	return string(k), v
}

// NextN increases the cursor position N times and return the encountered key-value pairs
func (bbc *BBoltCursor) NextN(steps uint) (docs map[string][]byte, endReached bool) {
	docs = make(map[string][]byte)
	if bbc.cursor == nil {
		return nil, true
	}
	for i := uint(0); i < steps; i++ {
		key, value := bbc.cursor.Next()
		if key == nil {
			endReached = true
			break
		}
		docs[string(key)] = value
	}
	return docs, endReached
}

// Prev iterations to the previous key from the current cursor
func (bbc *BBoltCursor) Prev() (key string, value []byte) {
	if bbc.cursor == nil {
		return "", nil
	}
	k, v := bbc.cursor.Prev()
	return string(k), v
}

// PrevN decreases the cursor position N times and return the encountered key-value pairs
func (bbc *BBoltCursor) PrevN(steps uint) (docs map[string][]byte, beginReached bool) {
	docs = make(map[string][]byte)
	if bbc.cursor == nil {
		return nil, true
	}

	for i := uint(0); i < steps; i++ {
		key, value := bbc.cursor.Prev()
		if key == nil {
			beginReached = true
			break
		}
		docs[string(key)] = value
	}
	return docs, beginReached
}

// Release the cursor
// This ends the bbolt bucket transaction
func (bbc *BBoltCursor) Release() {
	logrus.Infof("releasing bucket cursor")
	if bbc.bucket != nil {
		bbc.bucket.Tx().Rollback()
		bbc.cursor = nil
		bbc.bucket = nil
	}
}

// Seek returns a cursor with Next() and Prev() iterators
func (bbc *BBoltCursor) Seek(searchKey string) (key string, value []byte) {
	if bbc.cursor == nil {
		return "", nil
	}
	k, v := bbc.cursor.Seek([]byte(searchKey))
	return string(k), v
}

func NewBBoltCursor(bucket *bbolt.Bucket) *BBoltCursor {
	var bbCursor *bbolt.Cursor = nil
	if bucket != nil {
		bbCursor = bucket.Cursor()
	}
	bbc := &BBoltCursor{
		bucket: bucket,
		cursor: bbCursor,
	}

	return bbc
}
