// Package kvstore
package kvmem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hiveot/hub/internal/bucketstore"
)

// KVMemStore is an embedded, file backed, in-memory, lightweight, and very fast key-value bucket
// store.
// This is intended for simple use-cases with up to 100K records.
// Interestingly, this simple brute-force store using maps is faster than anything else I've tested and even
// scales up to 1M records. Pretty much all you need for basic databases.
//
// Changes are periodically persisted to file in the background.
//
// Limitations:
//   - No transaction support (basic usage, remember)
//   - Changes are periodically (default 3 second) written to disk
//
// Improvements for future considerations:
//   - Append-only to reduce disk writes for larger databases
//
// Crude performance using random read/writes:
// Todo: use proper benchmark
// Estimates using a data items of 1KB each, using a i5-4570S @2.90GHz cpu.
// note that these times are more than 10x faster than what a marshaller would take to serialize this data.
//
// Create&commit bucket, no data changes  (basically all overhead)
//   Dataset 1K,        0.2 us/op
//   Dataset 10K,       0.2 us/op
//   Dataset 100K       0.2 us/op
//   Dataset 1M         0.2 us/op
//
// Get bucket 1 record
//   Dataset 1K,        0.3 us/op
//   Dataset 10K,       0.3 us/op
//   Dataset 100K       0.3 us/op
//   Dataset 1M         0.3 us/op
//
// Set bucket 1 record
//   Dataset 1K,         0.3 us/op
//   Dataset 10K,        0.3 us/op
//   Dataset 100K        0.3 us/op
//   Dataset 1M          0.3 us/op
//
// Seek, 1 record
//   Dataset 1K,         163 us/op
//   Dataset 10K,          2 ms/op
//   Dataset 100K         28 ms/op  (*! cursor sort is slow.)
//   Dataset 1M          519 ms/op  (*! cursor sort is slow.)
//
//
// --- about jsonpath ---
// This was experimental because of the W3C WoT recommendation, and seems to work well.
// However this is shelved as the Hub has no use-case for it and the other stores don't support it.
//
// Query: jsonPath: `$[?(@.properties.title.name=="title1")]`
//    Approx 1.8 msec with a dataset of 1K records (1 result)
//    Approx 23 msec with a dataset of 10K records (1 result)
//    Approx 260 msec with a dataset of 100K records (1 result)
//
// A good overview of jsonpath implementations can be found here:
// > https://cburgmer.github.io/json-path-comparison/
// Two good options for jsonpath queries:
//  > github.com/ohler55/ojg/jp
//  > github.com/PaesslerAG/jsonpath
// Note that future implementations of this service can change the storage media used while
// maintaining API compatibility.
type KVMemStore struct {
	clientID string
	// collection of buckets, each being a map.
	buckets              map[string]*KVMemBucket
	storePath            string
	mutex                sync.RWMutex // simple locking is still fast enough
	updateCount          int32        // nr of updates since last save
	backgroundLoopEnded  chan bool
	backgroundLoopEnding chan bool
	writeDelay           time.Duration // delay before writing changes
	// cache for parsed json strings for faster query
	jsonCache map[string]interface{}
}

// makeShallowCopy returns a shallow copy of the given bucket map
// this copies the keys and values into a new map and bucket. Since keys are strings they
// are a true copy. Values are (immutable) byte arrays so not duplicated.
func makeShallowCopy(source map[string]*KVMemBucket) map[string]*KVMemBucket {

	var shallowCopy = make(map[string]*KVMemBucket)

	for bucketID, bucket := range source {
		bucketCopy := bucket.makeShallowCopy()
		shallowCopy[bucketID] = bucketCopy
	}
	return shallowCopy
}

// openStoreFile loads the store JSON content into a map.
// If the store doesn't exist it is created
// not concurrent safe
func openStoreFile(storePath string) (docs map[string]*KVMemBucket, err error) {
	docs, err = readStoreFile(storePath)
	if err != nil {
		docs = make(map[string]*KVMemBucket)
		err = writeStoreFile(storePath, docs)
	}
	return docs, err
}

// readStoreFile loads the store JSON content into a map
// not concurrent safe
func readStoreFile(storePath string) (docs map[string]*KVMemBucket, err error) {
	docs = make(map[string]*KVMemBucket)
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
func writeStoreFile(storePath string, docs map[string]*KVMemBucket) error {
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
	//rawData, err := oj.Marshal(docs)
	rawData, err := json.MarshalIndent(docs, "  ", "  ")
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
			if atomic.LoadInt32(&store.updateCount) > int32(0) {
				// make a shallow copy for writing to avoid a lock during write to disk
				shallowCopy := makeShallowCopy(store.buckets)
				atomic.StoreInt32(&store.updateCount, 0)
				store.mutex.Unlock()

				// nothing we can do here. error is already logged
				// FIXME: use separate write lock
				_ = writeStoreFile(store.storePath, shallowCopy)
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

// GetBucket returns a bucket and creates it if it doesn't exist
func (store *KVMemStore) GetBucket(bucketID string) (bucket bucketstore.IBucket) {

	if store.buckets == nil {
		panic("store is not open")
	}
	kvBucket, _ := store.buckets[bucketID]
	if kvBucket == nil {
		kvBucket = &KVMemBucket{
			BucketID: bucketID,
			ClientID: store.clientID,
			refCount: 0,
			KVMap:    make(map[string][]byte),
			mutex:    sync.RWMutex{},
			writable: true,
		}
		store.buckets[bucketID] = kvBucket
		bucket = kvBucket
	}
	if kvBucket != nil {
		// first time use of this bucket it might not have the callback yet.
		// also, after loading the store all handlers are empty
		kvBucket.setUpdateHandler(store.onBucketUpdated)
		kvBucket.incrRefCounter()
		// FIXME: this will make existing readonly buckets writable
		// The capability to use the bucket should be separated from the bucket itself
		kvBucket.writable = true
	}
	return kvBucket
}

// callback handler for notification that a bucket has been modified
func (store *KVMemStore) onBucketUpdated(bucket *KVMemBucket) {
	atomic.AddInt32(&store.updateCount, 1)
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

//// Size returns the number of items in the store
//func (store *KVMemStore) Size(context.Context, *emptypb.Empty) (*svc.SizeResult, error) {
//	store.mutex.RLock()
//	defer store.mutex.RUnlock()
//	res := &svc.SizeResult{
//		Count: int32(len(store.docs)),
//	}
//	return res, nil
//}

// SetWriteDelay sets the delay for writing after a change
func (store *KVMemStore) SetWriteDelay(delay time.Duration) {
	store.writeDelay = delay
}

// NewKVStore creates a store instance and load it with saved documents.
// Run Start to start the background loop and Stop to end it.
//  ClientID service or user for debugging and logging
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
