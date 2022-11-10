package pebble

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/pebble"
	"github.com/sirupsen/logrus"
)

type PebbleCursor struct {
	//db       *pebble.DB
	//bucket   *PebbleBucket
	bucketPrefix string // prefix to remove from keys returned by get/set/seek/first/lasst
	bucketID     string
	clientID     string
	iterator     *pebble.Iterator
}

// First moves the cursor to the first item
func (cursor *PebbleCursor) First() (key string, value []byte) {
	isValid := cursor.iterator.First()
	_ = isValid
	return cursor.getKV()
}

// Return the iterator current key and value
// This removes the bucket prefix
func (cursor *PebbleCursor) getKV() (key string, value []byte) {
	k := string(cursor.iterator.Key())
	v, err := cursor.iterator.ValueAndErr()
	if strings.HasPrefix(k, cursor.bucketPrefix) {
		key = k[len(cursor.bucketPrefix):]
	} else {
		err = fmt.Errorf("bucket key '%s' has no prefix '%s'", k, cursor.bucketPrefix)
	}
	// what to do in case of error?
	_ = err
	return key, v
}

// Last moves the cursor to the last item
func (cursor *PebbleCursor) Last() (key string, value []byte) {
	cursor.iterator.Last()
	return cursor.getKV()
}

// Next iterates to the next key from the current cursor
func (cursor *PebbleCursor) Next() (key string, value []byte) {
	isValid := cursor.iterator.Next()
	if !isValid {
		return "", nil
	}
	return cursor.getKV()
}

// NextN increases the cursor position N times and return the encountered key-value pairs
func (cursor *PebbleCursor) NextN(steps uint) (docs map[string][]byte, endReached bool) {
	docs = make(map[string][]byte)
	for i := uint(0); i < steps; i++ {
		endReached = !cursor.iterator.Next()
		if endReached {
			break
		}
		key, value := cursor.getKV()
		docs[key] = value
	}
	return docs, endReached
}

// Prev iterations to the previous key from the current cursor
func (cursor *PebbleCursor) Prev() (key string, value []byte) {
	isValid := cursor.iterator.Prev()
	if !isValid {
		return "", nil
	}
	return cursor.getKV()
}

// PrevN decreases the cursor position N times and return the encountered key-value pairs
func (cursor *PebbleCursor) PrevN(steps uint) (docs map[string][]byte, beginReached bool) {
	docs = make(map[string][]byte)
	for i := uint(0); i < steps; i++ {
		beginReached = !cursor.iterator.Prev()
		if beginReached {
			break
		}
		key, value := cursor.getKV()
		docs[key] = value
	}
	return docs, beginReached
}

// Release the cursor
func (cursor *PebbleCursor) Release() {
	err := cursor.iterator.Close()
	if err != nil {
		logrus.Errorf("unexpected error releasing cursor: %v", err)
	}
}

// Seek returns a cursor with Next() and Prev() iterators
func (cursor *PebbleCursor) Seek(searchKey string) (key string, value []byte) {
	bucketKey := cursor.bucketPrefix + searchKey
	cursor.iterator.SeekGE([]byte(bucketKey))
	return cursor.getKV()
}

func NewPebbleCursor(clientID, bucketID string, bucketPrefix string, iterator *pebble.Iterator) *PebbleCursor {
	cursor := &PebbleCursor{
		bucketPrefix: bucketPrefix,
		bucketID:     bucketID,
		clientID:     clientID,
		iterator:     iterator,
	}
	return cursor
}
