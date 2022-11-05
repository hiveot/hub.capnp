// Package kvmem
package kvmem

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
)

// KVMemBucket is an in-memory bucket for the KVMemBucket
type KVMemBucket struct {
	BucketID string            `json:"bucketID"`
	ClientID string            `json:"clientID"`
	refCount int               // simple ref count for error detection
	KVMap    map[string][]byte `json:"kvMap"`
	mutex    sync.RWMutex
	// cache for parsed json strings for faster query
	//queryCache map[string]interface{}

	// update handler callback to notify bucket owner
	updated func(bucket *KVMemBucket)
	// any SetXYZ method fails if writable is false
	writable bool
}

// Close the bucket and release its resources
// commit is not used as this store doesn't handle transactions.
// This decreases the refCount and detects an error if below 0
func (bucket *KVMemBucket) Close(commit bool) (err error) {
	// there are no transactions to commit
	_ = commit
	if commit && !bucket.writable {
		// this is just for error detection
		return fmt.Errorf("cant commit as bucket '%s' of client '%s' is not writable",
			bucket.BucketID, bucket.ClientID)
	}

	logrus.Infof("closing bucket '%s' of client '%s'", bucket.BucketID, bucket.ClientID)
	// this just lowers the refCount to detect leaks
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	bucket.refCount--
	if bucket.refCount < 0 {
		err = fmt.Errorf("bucket '%s' of client '%s' closed more often than opened",
			bucket.BucketID, bucket.ClientID)
	}
	return err
}

// Cursor returns a new cursor for iterating the bucket.
// The cursor MUST be closed after use to release its memory.
//
// This implementation is brute force. It generates a sorted list of key/values for use by the cursor.
// The cursor makes a shallow copy of the store. Store mutations are not reflected in the cursor.
//
// This should be fast enough for many use-cases. 100K records takes around 27msec on an i5@2.9GHz
//
// This returns a cursor with Next() and Prev() iterators
func (bucket *KVMemBucket) Cursor() (cursor bucketstore.IBucketCursor) {

	bucket.mutex.RLock()
	defer bucket.mutex.RUnlock()

	// build an ordered key list and shallow copy of the store
	sortedKeys := Map2SortedKeys(bucket.KVMap)
	cursor = NewKVCursor(bucket, sortedKeys)
	return cursor
}

// Delete a document from the bucket
// Also succeeds if the document doesn't exist
func (bucket *KVMemBucket) Delete(key string) error {
	if !bucket.writable {
		return fmt.Errorf("bucket '%s' of client '%s' is not writable", bucket.BucketID, bucket.ClientID)
	}

	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	logrus.Infof("Deleting key '%s' from bucket '%s'", key, bucket.BucketID)
	delete(bucket.KVMap, key)
	bucket.updated(bucket)
	return nil
}

// Get an object by its ID
// returns an error if the key does not exist.
func (bucket *KVMemBucket) Get(key string) (doc []byte, err error) {
	var found bool
	bucket.mutex.RLock()
	defer bucket.mutex.RUnlock()

	doc, found = bucket.KVMap[key]
	if !found {
		logrus.Debugf("key '%s' not found in map", key)
	}
	return doc, nil
}

// GetMultiple returns a batch of documents for the given key
// The document can be any text.
func (bucket *KVMemBucket) GetMultiple(keys []string) (docs map[string][]byte, err error) {

	bucket.mutex.RLock()
	defer bucket.mutex.RUnlock()
	docs = make(map[string][]byte)

	for _, key := range keys {
		val, found := bucket.KVMap[key]
		if found {
			docs[key] = val
		}
	}
	return docs, err
}

// Query for documents using JSONPATH
//
// This returns a cursor for a set of parsed documents that match.
// Note that the orderedKeys of the cursor are index numbers, not actual document orderedKeys.
//
// This parses the value into a json document. The parsed document is cached so successive queries
// will be faster.
//
// Eg `$[? @.properties.deviceType=="sensor"]`
//
//  jsonPath contains the query for each document.
//  offset contains the offset in the list of results, sorted by ID
//  limit contains the maximum or of responses, 0 for the default 100
//  orderedKeys can be used to limit the result to documents with the given orderedKeys. Use nil to ignore
//func (bucket *KVMemBucket) Query(
//	BucketID string, jsonPath string, orderedKeys []string) (cursor bucketstore.IBucketCursor, err error) {
//
//	//  "github.com/PaesslerAG/jsonpath" - just works, amazing!
//	// Unfortunately no filter with bracket notation $[? @.["title"]=="my title"]
//	// res, err := jsonpath.Get(jsonPath, store.docs)
//	// github.com/ohler55/ojg/jp - seems to work with in-mem maps, no @token in bracket notation
//	//logrus.Infof("jsonPath='%s', limit=%d", args.JsonPathQuery, args.Limit)
//
//	jpExpr, err := jp.ParseString(jsonPath)
//	if err != nil {
//		return nil, err
//	}
//	store.mutex.RLock()
//
//	// build an object tree of potential documents to query
//	var potentialDocs = make(map[string][]byte)
//
//	// no use to continue if the bucket doesn't exist
//	bucket, found := store.getBucket(BucketID, false)
//	if !found {
//		return cursor, fmt.Errorf("bucket '%s' not found", BucketID)
//	}
//	// when the list of orderedKeys is given, reduce to those that actually exist
//	if orderedKeys != nil {
//		for _, key := range orderedKeys {
//			doc, exists := bucket[key]
//			if exists {
//				potentialDocs[key] = doc
//			}
//		}
//	} else {
//		// get all docs
//		for key, docString := range bucket {
//			potentialDocs[key] = docString
//		}
//	}
//
//	// unlock now we have a copy of the document list
//	store.mutex.RUnlock()
//
//	// the query requires a parsed version of json docs
//	var docsToQuery = make(map[string]interface{})
//	for key, jsonDoc := range potentialDocs {
//		doc, found := store.jsonCache[key]
//		if found {
//			// use cached doc
//			docsToQuery[key] = doc
//		} else {
//			// parse and store
//			doc, err = oj.ParseString(string(jsonDoc))
//			if err == nil {
//				docsToQuery[key] = doc
//				store.jsonCache[key] = doc
//			}
//		}
//	}
//	// A big problem with jp.Get is that it returns an interface and we lose the orderedKeys.
//	// The only option is to query each document in order to retain the orderedKeys. That however affects jsonPath formulation.
//	validDocs := jpExpr.Get(docsToQuery)
//
//	// return the json docs instead of the interface.
//	// FIXME: Unfortunately that means marshalling again as we lost the orderedKeys... :(
//	cursorMap := make(map[string][]byte, 0)
//	cursorKeys := make([]string, len(validDocs))
//	for i, validDoc := range validDocs {
//		key := strconv.Itoa(i)
//		cursorKeys[i] = key
//		jsonDoc, _ := oj.Marshal(validDoc)
//		cursorMap[key] = jsonDoc
//	}
//	cursor = NewKVCursor(cursorMap, cursorKeys, 0)
//	return cursor, err
//}

//// Size returns the number of items in the store
//func (bucket *KVMemBucket) Size(context.Context, *emptypb.Empty) (*svc.SizeResult, error) {
//	store.mutex.RLock()
//	defer bucket.mutex.RUnlock()
//	res := &svc.SizeResult{
//		Count: int32(len(bucket.kvPairs)),
//	}
//	return res, nil
//}

func (bucket *KVMemBucket) ID() string {
	return bucket.BucketID
}

// increment the ref counter when a new bucket is requested
func (bucket *KVMemBucket) incrRefCounter() {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	bucket.refCount++
}

// Set writes a document to the store. If the document exists it is replaced.
//
//  A background process periodically checks the change count. When increased:
//  1. Lock the store while copying the index. Unlock when done.
//  2. Stream the in-memory json documents to a temp file.
//  3. If success, move the temp file to the store file using the OS atomic move operation.
//
func (bucket *KVMemBucket) Set(key string, doc []byte) error {
	if key == "" {
		return fmt.Errorf("missing key")
	} else if !bucket.writable {
		return fmt.Errorf("bucket '%s' of client '%s' is not writable", bucket.BucketID, bucket.ClientID)
	}

	// store the document and object
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	bucket.KVMap[key] = doc
	bucket.updated(bucket)
	return nil
}

func (bucket *KVMemBucket) setUpdateHandler(handler func(bucket *KVMemBucket)) {
	bucket.updated = handler
}

// SetMultiple writes a batch of key-values
func (bucket *KVMemBucket) SetMultiple(docs map[string][]byte) (err error) {
	if !bucket.writable {
		return fmt.Errorf("bucket '%s' of client '%s' is not writable", bucket.BucketID, bucket.ClientID)
	}
	// store the document and object
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	for k, v := range docs {
		bucket.KVMap[k] = v
	}
	bucket.updated(bucket)
	return nil
}

// returns a shallow copy of the bucket
func (bucket *KVMemBucket) makeShallowCopy() *KVMemBucket {
	bucket.mutex.RLock()
	defer bucket.mutex.RUnlock()
	shallowCopy := &KVMemBucket{
		BucketID: bucket.BucketID,
		ClientID: bucket.ClientID,
		refCount: 0,
		KVMap:    make(map[string][]byte),
		mutex:    sync.RWMutex{},
		updated:  nil, // not for production
		writable: false,
	}
	// shallow copy each bucket kv pairs as well
	for k, v := range bucket.KVMap {
		shallowCopy.KVMap[k] = v
	}
	return shallowCopy

}
