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
	k, v := bbc.cursor.First()
	return string(k), v
}

// Last moves the cursor to the last item
func (bbc *BBoltCursor) Last() (key string, value []byte) {
	k, v := bbc.cursor.Last()
	return string(k), v
}

// Next iterates to the next key from the current cursor
func (bbc *BBoltCursor) Next() (key string, value []byte) {
	k, v := bbc.cursor.Next()
	return string(k), v
}

// Prev iterations to the previous key from the current cursor
func (bbc *BBoltCursor) Prev() (key string, value []byte) {
	k, v := bbc.cursor.Prev()
	return string(k), v
}

// Release the cursor
// This ends the bbolt bucket transaction
func (bbc *BBoltCursor) Release() {
	logrus.Infof("releasing bucket cursor")
	bbc.bucket.Tx().Rollback()
	bbc.cursor = nil
	bbc.bucket = nil
}

// Seek returns a cursor with Next() and Prev() iterators
func (bbc *BBoltCursor) Seek(searchKey string) (key string, value []byte) {
	k, v := bbc.cursor.Seek([]byte(searchKey))
	return string(k), v
}

func NewBBoltCursor(bucket *bbolt.Bucket) *BBoltCursor {
	newCursor := bucket.Cursor()
	bbc := &BBoltCursor{
		bucket: bucket,
		cursor: newCursor,
	}
	return bbc
}
