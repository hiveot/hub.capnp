package kvmem

type KVBucketCursor struct {
	keys         []string
	bucket       map[string][]byte
	index        int
	currentKey   string
	currentValue []byte
}

// Close the cursor
func (cursor *KVBucketCursor) Close() {

}

func (cursor *KVBucketCursor) Key() string {
	return cursor.currentKey
}

// Next increases the cursor position and return the next key and value
// If the end is reached the returned key is empty
func (cursor *KVBucketCursor) Next() (key string, value []byte) {
	if cursor.index < len(cursor.keys) {
		cursor.index++
	}
	if cursor.index >= len(cursor.keys) {
		cursor.currentKey = ""
		cursor.currentValue = nil
	} else {
		cursor.currentKey = cursor.keys[cursor.index]
		cursor.currentValue = cursor.bucket[cursor.currentKey]
	}
	return cursor.currentKey, cursor.currentValue
}

// Prev decreases the cursor position and return the previous key and value
// If the head is reached the returned key is empty
func (cursor *KVBucketCursor) Prev() (key string, value []byte) {
	if cursor.index >= 0 {
		cursor.index--
	}
	if cursor.index < 0 || len(cursor.keys) == 0 {
		cursor.currentKey = ""
		cursor.currentValue = nil
	} else {
		cursor.currentKey = cursor.keys[cursor.index]
		cursor.currentValue = cursor.bucket[cursor.currentKey]
	}
	return cursor.currentKey, cursor.currentValue
}

func (cursor *KVBucketCursor) Value() []byte {
	return cursor.currentValue
}

func NewKVCursor(bucket map[string][]byte, keys []string, index int) *KVBucketCursor {
	cursor := &KVBucketCursor{
		keys:         keys,
		bucket:       bucket,
		index:        index,
		currentKey:   "",
		currentValue: nil,
	}
	if index >= 0 && index < len(keys) {
		cursor.currentKey = keys[index]
		cursor.currentValue = bucket[cursor.currentKey]
	}
	return cursor
}
