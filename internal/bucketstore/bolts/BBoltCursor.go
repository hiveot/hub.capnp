package bolts

import "go.etcd.io/bbolt"

// BBoltCursor is a wrapper around the bbolt cursor to map to the IBucketCursor API and to
// ensure its transaction is released after the cursor is no longer used.
// This implements the IBucketCursor API
type BBoltCursor struct {
	bucket *bbolt.Bucket
	cursor *bbolt.Cursor
}

// Close the cursor
// This does not release the transaction that created the bucket
func (cursor *BBoltCursor) Close() error {
	cursor.cursor = nil
	cursor.bucket = nil
	return nil
}

// First moves the cursor to the first item
func (cursor *BBoltCursor) First() (key string, value []byte) {
	k, v := cursor.cursor.First()
	return string(k), v
}

// Last moves the cursor to the last item
func (cursor *BBoltCursor) Last() (key string, value []byte) {
	k, v := cursor.cursor.Last()
	return string(k), v
}

// Next iterates to the next key from the current cursor
func (cursor *BBoltCursor) Next() (key string, value []byte) {
	k, v := cursor.cursor.Next()
	return string(k), v
}

// Prev iterations to the previous key from the current cursor
func (cursor *BBoltCursor) Prev() (key string, value []byte) {
	k, v := cursor.cursor.Prev()
	return string(k), v
}

// Seek returns a cursor with Next() and Prev() iterators
func (cursor *BBoltCursor) Seek(searchKey string) (key string, value []byte) {
	k, v := cursor.cursor.Seek([]byte(searchKey))
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
