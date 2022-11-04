// Package kvstore
package kvmem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/ohler55/ojg/oj"
	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
)

// MaxLimit is the default max nr of items to return in list or query
const MaxLimit = 1000

type BucketMap map[string][]byte

// KVMemStore is an embedded, file backed, in-memory, lightweight, and very fast key-value bucket
// store, intended for persisting and reading small simple datasets such as application state.
// This is for simple use-cases, but man its also faster than anything else I've tested.
//
// Changes are periodically persisted to file in the background.
//
// The current implementation is a simple in-memory store using maps. Its operations are atomic,
// data will be consistent, has isolation between transactions, and is eventually durable.
// Durability is guaranteed after a flush operation that writes changes to disk.
//
// Crude performance estimate using a data items of 1KB each
// using a i5-4570S @2.90GHz cpu.
// Read: Approx 0.019 msec (avg of 100K reads)
// Write: Approx 0.016 msec (avg of 100K writes)
// List: First  1000 records (1MB data)
//    Approx 18 msec with a dataset of 1K records
//    Approx 20 msec with a dataset of 10K records
//    Approx 66 msec with a dataset of 100K records
// Query: jsonPath: `$[?(@.properties.title.name=="title1")]`
//    Approx 1.8 msec with a dataset of 1K records (1 result)
//    Approx 23 msec with a dataset of 10K records (1 result)
//    Approx 260 msec with a dataset of 100K records (1 result)
//
//
// A good overview of jsonpath implementations can be found here:
// > https://cburgmer.github.io/json-path-comparison/
// Two good options for jsonpath queries:
//  > github.com/ohler55/ojg/jp
//  > github.com/PaesslerAG/jsonpath
//
// Note that future implementations of this service can change the storage media used while
// maintaining API compatibility.
type KVMemStore struct {
	clientID string
	// collection of buckets, each being a map.
	buckets              map[string]BucketMap
	storePath            string
	mutex                sync.RWMutex // simple locking is still fast enough
	updateCount          int          // nr of updates since last save
	backgroundLoopEnded  chan bool
	backgroundLoopEnding chan bool
	writeDelay           time.Duration // delay before writing changes
	// cache for parsed json strings for faster query
	jsonCache map[string]interface{}
}

// openStoreFile loads the store JSON content into a map.
// If the store doesn't exist it is created
// not concurrent safe
func openStoreFile(storePath string) (docs map[string]BucketMap, err error) {
	docs, err = readStoreFile(storePath)
	if err != nil {
		docs = make(map[string]BucketMap)
		err = writeStoreFile(storePath, docs)
	}
	return docs, err
}

// readStoreFile loads the store JSON content into a map
// not concurrent safe
func readStoreFile(storePath string) (docs map[string]BucketMap, err error) {
	docs = make(map[string]BucketMap)
	var rawData []byte
	rawData, err = os.ReadFile(storePath)
	if err == nil {
		err = json.Unmarshal(rawData, &docs)

		if err != nil {
			// todo: chain errors
			logrus.Errorf("failed read store '%s', error %s. Recover with an empty store. Sorry.", storePath, err)
		}
	}
	return docs, err
}

// writeStoreFile writes the store to file.
// This creates the folder if it doesn't exist. (the parent must exist)
// Note this is not concurrent safe. Callers must lock or create a shallow copy of the buckets.
//
//   storePath is the full path to the file
//   docs contains an object map of the store objects
func writeStoreFile(storePath string, docs map[string]BucketMap) error {
	logrus.Infof("writeStoreFile: Flush changes to json store at '%s'", storePath)

	// create the folder if needed
	storeFolder := path.Dir(storePath)
	_, err := os.Stat(storeFolder)
	if os.IsNotExist(err) {
		// folder doesn't exist. Attempt to create it
		logrus.Warningf("Store folder '%s' does not exist. Creating it now.", storeFolder)
		err = os.Mkdir(storeFolder, 0700)
	}
	// If the folder can't be created we're dead in the water
	if err != nil {
		err = fmt.Errorf("unable to create the store folder at '%s': %s", storeFolder, err)
	}
	if err != nil {
		return err
	}

	// serialize the data to json for writing. Use indent for testing and debugging
	rawData, err := oj.Marshal(docs)
	//rawData, err := json.MarshalIndent(docs, "  ", "  ")
	if err != nil {
		// yeah this is pretty fatal too
		logrus.Panicf("Unable to marshal documents while saving store to %s: %s", storePath, err)
	}
	// First write content to temp file
	// The temp file is opened with 0600 permissions
	tmpName := storePath + ".tmp"
	err = ioutil.WriteFile(tmpName, rawData, 0600)
	if err != nil {
		// ouch, wth?
		err := fmt.Errorf("error while creating tempfile for jsonstore: %s", err)
		logrus.Error(err)
		return err
	}

	// move the temp file to the final store file.
	// this replaces the file if it already exists
	err = os.Rename(tmpName, storePath)
	if err != nil {
		err := fmt.Errorf("error while moving tempfile to jsonstore '%s': %s", storePath, err)
		logrus.Error(err)
		return err
	}
	return nil
}

// autoSaveLoop periodically saves changes to the store
func (store *KVMemStore) autoSaveLoop() {
	logrus.Infof("auto-save loop started")

	defer close(store.backgroundLoopEnded)

	for {
		select {
		case <-store.backgroundLoopEnding:
			logrus.Infof("Autosave loop ended")
			return
		case <-time.After(store.writeDelay):
			store.mutex.Lock()
			if store.updateCount > 0 {
				// make a shallow copy for writing to avoid a lock during write to disk
				var indexCopy = make(map[string]BucketMap)
				for bucketID, bucket := range store.buckets {
					bucketCopy := make(BucketMap)
					indexCopy[bucketID] = bucketCopy
					for k, v := range bucket {
						bucketCopy[k] = v
					}
				}
				store.updateCount = 0
				store.mutex.Unlock()

				// nothing we can do here. error is already logged
				_ = writeStoreFile(store.storePath, indexCopy)
			} else {
				store.mutex.Unlock()
			}
		}
	}
}

// Close the store and stop the background update.
// If any changes are remaining then write to disk now.
func (store *KVMemStore) Close() error {
	var err error
	logrus.Infof("closing store for client '%s'", store.clientID)

	if store.buckets == nil {
		panic("store already closed")
	}

	store.backgroundLoopEnding <- true

	// wait for the background loop to end
	<-store.backgroundLoopEnded

	store.mutex.Lock()
	defer store.mutex.Unlock()

	// flush any remaining changes
	if store.updateCount > 0 {
		err = writeStoreFile(store.storePath, store.buckets)
	}
	store.buckets = nil
	logrus.Infof("Store '%s' close completed. Background loop ended", store.storePath)
	return err
}

// Delete a document from the store
// Also succeeds if the document doesn't exist
func (store *KVMemStore) Delete(bucketID string, key string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	logrus.Infof("Deleting key '%s' from bucket '%s'", key, bucketID)
	bucket, found := store.getBucket(bucketID, false)
	if found {
		delete(bucket, key)
	}
	store.updateCount++
	return nil
}

// Get an object by its ID
// returns an error if the key does not exist.
func (store *KVMemStore) Get(bucketID, key string) (doc []byte, found bool, err error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	bucket, found := store.getBucket(bucketID, false)
	if found {
		doc, found = bucket[key]
	}
	return doc, found, nil
}

// get a bucket or create if it doesn't exist
func (store *KVMemStore) getBucket(bucketID string, createIfNotExist bool) (bucket BucketMap, found bool) {
	if store.buckets == nil {
		panic("store is not open")
	}
	bucket, found = store.buckets[bucketID]
	if !found && createIfNotExist {
		bucket = BucketMap{}
		store.buckets[bucketID] = bucket
		found = true
	}
	return bucket, found
}

// GetMultiple returns a batch of documents for the given key
// The document can be any text.
func (store *KVMemStore) GetMultiple(
	bucketID string, keys []string) (docs map[string][]byte, err error) {

	store.mutex.RLock()
	defer store.mutex.RUnlock()
	docs = make(map[string][]byte)

	bucket, found := store.getBucket(bucketID, false)
	if found {
		for _, key := range keys {
			val, found := bucket[key]
			if found {
				docs[key] = val
			}
		}
	}
	return docs, err
}

// Open the store and start the background loop for saving changes
func (store *KVMemStore) Open() error {
	logrus.Infof("Opening store from '%s'", store.storePath)
	var err error
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.buckets != nil {
		panic("store already open")
	}
	store.buckets, err = openStoreFile(store.storePath)
	store.backgroundLoopEnding = make(chan bool)
	store.backgroundLoopEnded = make(chan bool)
	if err == nil {
		go store.autoSaveLoop()
	}
	// allow a context switch to start the autoSaveLoop to avoid problems
	// if the store is closed immediately.
	//time.Sleep(time.Millisecond)
	return err
}

// Query for documents using JSONPATH
//
// This returns a cursor for a set of parsed documents that match.
// Note that the keys of the cursor are index numbers, not actual document keys.
//
// This parses the value into a json document. The parsed document is cached so successive queries
// will be faster.
//
// Eg `$[? @.properties.deviceType=="sensor"]`
//
//  jsonPath contains the query for each document.
//  offset contains the offset in the list of results, sorted by ID
//  limit contains the maximum or of responses, 0 for the default 100
//  keys can be used to limit the result to documents with the given keys. Use nil to ignore
//func (store *KVMemStore) Query(
//	bucketID string, jsonPath string, keys []string) (cursor bucketstore.IBucketCursor, err error) {
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
//	bucket, found := store.getBucket(bucketID, false)
//	if !found {
//		return cursor, fmt.Errorf("bucket '%s' not found", bucketID)
//	}
//	// when the list of keys is given, reduce to those that actually exist
//	if keys != nil {
//		for _, key := range keys {
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
//	// A big problem with jp.Get is that it returns an interface and we lose the keys.
//	// The only option is to query each document in order to retain the keys. That however affects jsonPath formulation.
//	validDocs := jpExpr.Get(docsToQuery)
//
//	// return the json docs instead of the interface.
//	// FIXME: Unfortunately that means marshalling again as we lost the keys... :(
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

// RemoveAll empties the store. Intended for testing.
//func (store *KVMemStore) RemoveAll(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
//	store.mutex.Lock()
//	defer store.mutex.Unlock()
//	store.docs = make(map[string]interface{})
//	store.updateCount++
//	return nil, nil
//}
//
//// Size returns the number of items in the store
//func (store *KVMemStore) Size(context.Context, *emptypb.Empty) (*svc.SizeResult, error) {
//	store.mutex.RLock()
//	defer store.mutex.RUnlock()
//	res := &svc.SizeResult{
//		Count: int32(len(store.docs)),
//	}
//	return res, nil
//}

// Seek returns an iterator
// This implementation is brute force. It generates a sorted list of keys for use by the cursor.
// This should still be fast enough for most cases. (test shows around 500msec for 1 million keys).
//
//  bucketID to seach for. Returns and error if the bucket is not found
//  key is the starting point. If key doesn't exist, the next closest key will be used.
//
// This returns a cursor with Next() and Prev() iterators
func (store *KVMemStore) Seek(
	bucketID string, key string) (cursor bucketstore.IBucketCursor, err error) {

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	// build a sorted key list
	bucket, found := store.getBucket(bucketID, false)
	if !found {
		err := fmt.Errorf("bucket '%s not found", bucketID)
		return cursor, err
	}
	sortedKeys := Map2SortedKeys(bucket)
	i := sort.SearchStrings(sortedKeys, key)
	cursor = NewKVCursor(bucket, sortedKeys, i)
	return cursor, err
}

// SetWriteDelay sets the delay for writing after a change
func (store *KVMemStore) SetWriteDelay(delay time.Duration) {
	store.writeDelay = delay
}

// Set writes a document to the store. If the document exists it is replaced.
//
//  A background process periodically checks the change count. When increased:
//  1. Lock the store while copying the index. Unlock when done.
//  2. Stream the in-memory json documents to a temp file.
//  3. If success, move the temp file to the store file using the OS atomic move operation.
//
func (store *KVMemStore) Set(bucketID, key string, doc []byte) error {
	if key == "" {
		return fmt.Errorf("missing key")
	}

	// store the document and object
	store.mutex.Lock()
	defer store.mutex.Unlock()

	bucket, _ := store.getBucket(bucketID, true)
	bucket[key] = doc
	store.updateCount++
	return nil
}

// SetMultiple writes a batch of key-values
func (store *KVMemStore) SetMultiple(
	bucketID string, docs map[string][]byte) (err error) {

	// store the document and object
	store.mutex.Lock()
	defer store.mutex.Unlock()
	bucket, _ := store.getBucket(bucketID, true)
	for k, v := range docs {
		bucket[k] = v
	}
	store.updateCount++
	return nil
}

// NewKVStore creates a store instance and load it with saved documents.
// Run Start to start the background loop and Stop to end it.
//  clientID service or user for debugging and logging
//  storeFile path to storage file
func NewKVStore(clientID, storePath string) (store *KVMemStore) {
	writeDelay := time.Duration(3000) * time.Millisecond
	store = &KVMemStore{
		//jsonDocs:             make(map[string]string),
		clientID:             clientID,
		buckets:              nil, // will be set after open
		storePath:            storePath,
		backgroundLoopEnding: nil,
		backgroundLoopEnded:  nil,
		mutex:                sync.RWMutex{},
		writeDelay:           writeDelay,
		jsonCache:            make(map[string]interface{}),
	}
	return store
}
