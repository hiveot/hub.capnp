// Package kvstore
package kvstore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/sirupsen/logrus"
)

// MaxLimit is the default max nr of items to return in list or query
const MaxLimit = 1000

// KVStore is a simple lightweight key-value storage library that distinguishes itself by:
//  1. being able to list available documents using offset and limit.
//  2. jsonpath query the values of the documents, if those are valid serialized json
//  3. no dependencies other than the jsonpath query
//  4. periodically persists the data to file
//
// It is intended for storing and querying small simple datasets with few writes and mostly read operations.
//
// The current implementation is a simple in-memory store using maps, whose changes are regularly written to disk.
// Its operations are atomic, data will be consistent, has isolation between transactions, and is eventually durable.
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
type KVStore struct {
	docs                 map[string]string // json documents
	storePath            string
	mutex                sync.RWMutex
	limit                int // default limit value in list and queries
	updateCount          int // nr of updates since last save
	backgroundLoopEnded  chan bool
	backgroundLoopEnding chan bool
	writeDelay           time.Duration // delay before writing changes
	// cache for parsed json strings for faster query
	jsonCache map[string]interface{}
}

// return the list of keys of a given map.
func getKeys(m map[string]string) []string {
	keyList := make([]string, len(m))
	i := 0
	for key := range m {
		keyList[i] = key
		i++
	}
	return keyList
}

// openStoreFile loads the store JSON content into a map.
// If the store doesn't exist it is created
func openStoreFile(storePath string) (docs map[string]string, err error) {
	docs, err = readStoreFile(storePath)
	if err != nil {
		docs = make(map[string]string)
		err = writeStoreFile(storePath, docs)
	}
	return docs, err
}

// readStoreFile loads the store JSON content into a map
func readStoreFile(storePath string) (docs map[string]string, err error) {
	docs = make(map[string]string)
	var rawData []byte
	rawData, err = os.ReadFile(storePath)
	if err == nil {
		err = json.Unmarshal(rawData, &docs)

		if err != nil {
			logrus.Errorf("readStoreFile: failed read store '%s', error %s. Recover with an empty store. Sorry.", storePath, err)
		}
	}
	return docs, err
}

// writeStoreFile writes the store to file.
// This creates the folder if it doesn't exist. (the parent must exist)
//   storePath is the full path to the file
//   docs contains an object map of the store objects
func writeStoreFile(storePath string, docs map[string]string) error {
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
		logrus.Panicf("Unable to create the store folder at '%s'. Error %s", storeFolder, err)
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
		err := fmt.Errorf("writeStoreFile: Error while creating tempfile for jsonstore: %s", err)
		logrus.Error(err)
		return err
	}

	// move the temp file to the final store file.
	// this replaces the file if it already exists
	err = os.Rename(tmpName, storePath)
	if err != nil {
		err := fmt.Errorf("writeStoreFile: Error while moving tempfile to jsonstore '%s': %s", storePath, err)
		logrus.Error(err)
		return err
	}
	return nil
}

// autoSaveLoop periodically saves changes to the store
func (store *KVStore) autoSaveLoop() {
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
				var indexCopy = make(map[string]string)
				for index, element := range store.docs {
					indexCopy[index] = element
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

// List returns a list of documents
//
//  offset contains the offset in the list of results, sorted by ID
//  limit contains the maximum or of responses, 0 for the default 100
//  keys can be used to limit the result to the given document keys. Use nil to ignore.
//
// This returns an empty list if offset is equal or larger than the available nr of documents
// Note that paging slows performance with larger datasets (100K+) due to sorting of keys
func (store *KVStore) List(limit int, offset int, keys []string) (docs map[string]string, err error) {
	if limit <= 0 {
		limit = store.limit
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	// Use the given keys or get them all
	// return results in given key order
	keyList := keys
	if keyList == nil {
		keyList = getKeys(store.docs)
		// sort the keys when using paging, eg offset or limit < size
		if offset > 0 || int(limit) < len(store.docs) {
			sort.Strings(keyList)
		}
	}
	// apply paging
	nrResults := len(keyList) - int(offset)
	if nrResults < 0 {
		nrResults = 0
	}
	if nrResults > int(limit) {
		nrResults = int(limit)
	}
	// collect the result
	docs = make(map[string]string, nrResults)
	for index := 0; index < nrResults; index++ {
		key := keyList[int(offset)+index]
		docs[key] = store.docs[key]
	}
	return docs, err
}

// Query for documents using JSONPATH
// This returns a set of parsed documents that match.
// This parses the value into a json document. The parsed document is cached so successive queries
// will be faster.
//
// Eg `$[? @.properties.deviceType=="sensor"]`
//
//  jsonPath contains the query for each document.
//  offset contains the offset in the list of results, sorted by ID
//  limit contains the maximum or of responses, 0 for the default 100
//  keys can be used to limit the result to documents with the given keys. Use nil to ignore
func (store *KVStore) Query(jsonPath string, offset int, limit int, keys []string) (docs []string, err error) {
	//  "github.com/PaesslerAG/jsonpath" - just works, amazing!
	// Unfortunately no filter with bracket notation $[? @.["title"]=="my title"]
	// res, err := jsonpath.Get(jsonPath, store.docs)
	// github.com/ohler55/ojg/jp - seems to work with in-mem maps, no @token in bracket notation
	//logrus.Infof("jsonPath='%s', limit=%d", args.JsonPathQuery, args.Limit)
	jpExpr, err := jp.ParseString(jsonPath)
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = store.limit
	}
	store.mutex.RLock()

	// build an object tree of potential documents to query
	var potentialDocs = make(map[string]string)

	// when the list of keys is given, reduce to those that actually exist
	if keys != nil {
		for _, key := range keys {
			doc, exists := store.docs[key]
			if exists {
				potentialDocs[key] = doc
			}
		}
	} else if len(store.docs) > limit || offset > 0 {
		// when paging, use sorted keys
		// to apply paging the keys need to be sorted
		keys := getKeys(store.docs)
		sort.Strings(keys)
		for _, key := range keys {
			potentialDocs[key] = store.docs[key]
		}
	} else {
		// get all docs
		for key, docString := range store.docs {
			potentialDocs[key] = docString
		}
	}

	// unlock now we have a copy of the document list
	store.mutex.RUnlock()

	// the query requires a parsed version of json docs
	var docsToQuery = make(map[string]interface{})
	for key, jsonText := range potentialDocs {
		doc, found := store.jsonCache[key]
		if found {
			// use cached doc
			docsToQuery[key] = doc
		} else {
			// parse and store
			doc, err = oj.ParseString(jsonText)
			if err == nil {
				docsToQuery[key] = doc
				store.jsonCache[key] = doc
			}
		}
	}
	// A big problem with jp.Get is that it returns an interface and we lose the keys.
	// The only option is to query each document in order to retain the keys. That however affects jsonPath formulation.
	validDocs := jpExpr.Get(docsToQuery)

	// Apply paging to the result
	if offset > 0 || len(validDocs) > int(limit) {
		nrResults := len(validDocs) - int(offset)
		if nrResults > int(limit) {
			nrResults = int(limit)
		}
		if int(offset) >= len(validDocs) {
			validDocs = make([]interface{}, 0)
		} else {
			validDocs = validDocs[offset : int(offset)+nrResults]
		}
	}
	// return the json docs instead of the interface.
	// Unfortunately that means marshalling again as we lost the keys... :(
	jsonDocs := make([]string, 0)
	for _, validDoc := range validDocs {
		jsonDoc, _ := oj.Marshal(validDoc)
		jsonDocs = append(jsonDocs, string(jsonDoc))
	}

	return jsonDocs, err
}

// Read an object by its ID
func (store *KVStore) Read(key string) (string, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	doc, ok := store.docs[key]
	if !ok {
		return doc, fmt.Errorf("key '%s' not found", key)
	}
	return doc, nil
}

// Remove a document from the store
// Also succeeds if the document doesn't exist
func (store *KVStore) Remove(key string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.docs, key)
	store.updateCount++
}

// RemoveAll empties the store. Intended for testing.
//func (store *KVStore) RemoveAll(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
//	store.mutex.Lock()
//	defer store.mutex.Unlock()
//	store.docs = make(map[string]interface{})
//	store.updateCount++
//	return nil, nil
//}
//
//// Size returns the number of items in the store
//func (store *KVStore) Size(context.Context, *emptypb.Empty) (*svc.SizeResult, error) {
//	store.mutex.RLock()
//	defer store.mutex.RUnlock()
//	res := &svc.SizeResult{
//		Count: int32(len(store.docs)),
//	}
//	return res, nil
//}

// SetWriteDelay sets the delay for writing after a change
func (store *KVStore) SetWriteDelay(delay time.Duration) {
	store.writeDelay = delay
}

// Start the store background loop for saving changes
func (store *KVStore) Start() error {
	logrus.Infof("KVStore.Start: Opening store from '%s'", store.storePath)
	var err error
	store.mutex.Lock()
	defer store.mutex.Unlock()

	go store.autoSaveLoop()
	return err
}

// Stop the background update.
// If any changes are remaining then write to disk now.
func (store *KVStore) Stop() error {
	var err error
	logrus.Infof("KVStore.Stop")
	store.backgroundLoopEnding <- true

	// wait for the background loop to end
	<-store.backgroundLoopEnded

	store.mutex.Lock()
	defer store.mutex.Unlock()

	// flush any remaining changes
	if store.updateCount > 0 {
		err = writeStoreFile(store.storePath, store.docs)
	}
	logrus.Infof("KVStore.Stop completed. Background loop ended")
	return err
}

// Write a document to the store. If the document exists it is replaced.
//
//  A background process periodically checks the change count. When increased:
//  1. Lock the store while copying the index. Unlock when done.
//  2. Stream the in-memory json documents to a temp file.
//  3. If success, move the temp file to the store file using the OS atomic move operation.
//
func (store *KVStore) Write(key string, doc string) error {
	if key == "" {
		return fmt.Errorf("KVStore.Write: missing key")
	}

	// store the document and object
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.docs[key] = doc
	store.updateCount++
	return nil
}

// NewKVStore creates a store instance and load it with saved documents.
// Run Start to start the background loop and Stop to end it.
//
//  storeFile path to storage file
//  writeDelayMsec max delay before flushing changes to disk. Default 3000
func NewKVStore(storePath string) (store *KVStore, err error) {
	writeDelay := time.Duration(3000) * time.Millisecond
	store = &KVStore{
		//jsonDocs:             make(map[string]string),
		docs:                 make(map[string]string),
		storePath:            storePath,
		limit:                MaxLimit,
		backgroundLoopEnding: make(chan bool),
		backgroundLoopEnded:  make(chan bool),
		mutex:                sync.RWMutex{},
		writeDelay:           writeDelay,
		jsonCache:            make(map[string]interface{}),
	}
	store.docs, err = openStoreFile(store.storePath)
	return store, err
}
