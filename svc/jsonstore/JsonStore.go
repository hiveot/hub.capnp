// Package jsonstore
package jsonstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/wostzone/wost.grpc/go/svc"
)

// MaxLimit is the default max nr of items to return in list or query
const MaxLimit = 1000

// JsonStore is a simple lightweight storage service intended for basic read, write, list and jsonpath query
// operations. This is not recommended for large documents or large amount of data.
//
// The current implementation is a simple in-memory store whose changes are regularly written to disk.
// Its operations are atomic, data will be consistent, has isolation between transactions, and is eventually durable.
// Durability is guaranteed after a flush operation that writes changes to disk.
//
// Crude performance estimate using a data items of 1KB each
// using a i5-4570S @2.90GHz cpu.
// Read: Approx 0.019 msec (avg of 100K reads)
// Write: Approx 0.016 msec (avg of 100K writes)
// List: limit of 1000 (1MB data)
//    Approx 18 msec with a dataset of 1K records
//    Approx 20 msec with a dataset of 10K records
//    Approx 66 msec with a dataset of 100K records
// Query: jsonPath: `$[?(@.properties.title.name=="title1")]`
//    Approx 1.8 msec with a dataset of 1K records (1 result)
//    Approx 23 msec with a dataset of 10K records (1 result)
//    Approx 260 msec with a dataset of 100K records (1 result)
//
//
// It is recommended to use redis, unless:
//  - a list operation is needed
//  - jsonpath query is needed
//  - a low memory footprint is needed
//
// A good overview of jsonpath implementations can be found here:
// > https://cburgmer.github.io/json-path-comparison/
// Two good options for jsonpath queries:
//  > github.com/ohler55/ojg/jp
//  > github.com/PaesslerAG/jsonpath
//
// Note that future implementations of this service can change the storage media used while
// maintaining API compatibility.
type JsonStore struct {
	svc.UnimplementedJsonStoreServer // implement the gRPC API

	docs                 map[string]interface{} // json documents
	storePath            string
	mutex                sync.RWMutex
	limit                int // default limit value in list and queries
	updateCount          int // nr of updates since last save
	backgroundLoopEnded  chan bool
	backgroundLoopEnding chan bool
	writeDelay           time.Duration // delay before writing changes
}

// return the list of keys of a given map.
func getKeys(m map[string]interface{}) []string {
	keyList := make([]string, len(m))
	i := 0
	for key := range m {
		keyList[i] = key
		i++
	}
	return keyList
}

// openStoreFile loads the store JSON content into a map.
// If the store file doesn't exist or is corrupt it will be re-created.
func openStoreFile(storePath string) (docs map[string]interface{}, err error) {
	docs, err = readStoreFile(storePath)
	if err != nil {
		docs = make(map[string]interface{})
		err = writeStoreFile(storePath, docs)
	}
	return docs, err
}

// readStoreFile loads the store JSON content into a map
func readStoreFile(storePath string) (docs map[string]interface{}, err error) {
	docs = make(map[string]interface{})
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
func writeStoreFile(storePath string, docs map[string]interface{}) error {
	logrus.Infof("writeStoreFile: Writing data to json store to '%s'", storePath)

	// create the folder if needed
	storeFolder := path.Dir(storePath)
	_, err := os.Stat(storeFolder)
	if os.IsNotExist(err) {
		err = os.Mkdir(storeFolder, os.ModeDir)
	}
	if err != nil {
		logrus.Errorf("createStoreFolder. Error %s", err)
	}

	// serialize the data to json for writing. Use indent for testing and debugging
	rawData, err := json.MarshalIndent(docs, "  ", "  ")
	if err != nil {
		err := fmt.Errorf("writeStoreFile: Error while saving store to %s: %s", storePath, err)
		logrus.Error(err)
		return err
	}
	// First write content to temp file
	// The temp file is opened with 0600 permissions
	tmpName := storePath + ".tmp"
	err = ioutil.WriteFile(tmpName, rawData, 0600)
	if err != nil {
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
func (store *JsonStore) autoSaveLoop() {
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
				var indexCopy = make(map[string]interface{})
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
// This returns an empty list if offset is equal or larger than the available nr of documents
// Note that paging slows performance with larger datasets (10K+) due to sorting of keys
func (store *JsonStore) List(_ context.Context, args *svc.List_Args) (*svc.ResultValues, error) {
	if args.Limit <= 0 {
		args.Limit = int32(store.limit)
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	// Use the given keys or get them all
	// return results in given key order
	keyList := args.Keys
	if keyList == nil {
		keyList = getKeys(store.docs)
		// sort the keys when using paging, eg offset or limit < size
		if args.Offset > 0 || int(args.Limit) < len(store.docs) {
			sort.Strings(keyList)
		}
	}
	// apply paging
	nrResults := len(keyList) - int(args.Offset)
	if nrResults < 0 {
		nrResults = 0
	}
	if nrResults > int(args.Limit) {
		nrResults = int(args.Limit)
	}
	// collect the result
	items := make([]interface{}, nrResults)
	for index := 0; index < nrResults; index++ {
		key := keyList[int(args.Offset)+index]
		items[index] = store.docs[key]
	}
	jsonDocs, _ := json.Marshal(items)
	res := &svc.ResultValues{
		JsonDocs:  string(jsonDocs),
		Available: int32(nrResults),
	}
	return res, nil
}

// Query for documents using JSONPATH
//
// This implementation iterates the documents and matches each in turn.
//
// Eg `$[? @.properties.deviceType=="sensor"]`
//  jsonPath contains the query for each document.
//  offset contains the offset in the list of results, sorted by ID
//  limit contains the maximum or of responses, 0 for the default 100
func (store *JsonStore) Query(_ context.Context, args *svc.Query_Args) (*svc.ResultValues, error) {

	//  "github.com/PaesslerAG/jsonpath" - just works, amazing!
	// Unfortunately no filter with bracket notation $[? @.["title"]=="my title"]
	// res, err := jsonpath.Get(jsonPath, store.docs)
	// github.com/ohler55/ojg/jp - seems to work with in-mem maps, no @token in bracket notation
	//logrus.Infof("jsonPath='%s', limit=%d", args.JsonPathQuery, args.Limit)
	jpExpr, err := jp.ParseString(args.JsonPathQuery)
	if err != nil {
		return nil, err
	}
	if args.Limit == 0 {
		args.Limit = int32(store.limit)
	}
	store.mutex.RLock()

	// build an object tree of potential documents to query.
	// If a list is given then use the list, otherwise include all documents.
	var potentialDocs = make(map[string]interface{})
	if args.Keys != nil {
		for _, key := range args.Keys {
			doc, hasDoc := store.docs[key]
			if hasDoc {
				potentialDocs[key] = doc
			}
		}
	} else {
		// when paging, use sorted keys
		if len(store.docs) > int(args.Limit) || args.Offset > 0 {
			// to apply paging the keys need to be sorted
			keys := getKeys(store.docs)
			sort.Strings(keys)
			for index := 0; index < len(keys); index++ {
				key := keys[index]
				potentialDocs[key] = store.docs[key]
			}
		} else {
			// get all docs
			for key, doc := range store.docs {
				potentialDocs[key] = doc
			}
		}
	}
	defer store.mutex.RUnlock()

	// the query
	validDocs := jpExpr.Get(potentialDocs)
	// Apply paging
	if args.Offset > 0 || len(validDocs) > int(args.Limit) {
		nrResults := len(validDocs) - int(args.Offset)
		if nrResults > int(args.Limit) {
			nrResults = int(args.Limit)
		}
		if int(args.Offset) >= len(validDocs) {
			validDocs = make([]interface{}, 0)
		} else {
			validDocs = validDocs[args.Offset : int(args.Offset)+nrResults]
		}
	}

	jsonDocs, _ := json.Marshal(validDocs)

	// finally, collect the documents themselves
	qResults := &svc.ResultValues{
		JsonDocs:  string(jsonDocs),
		Available: int32(len(validDocs)),
	}
	//for _, key := range resultKeys {
	//	qResults.Items[key] = store.jsonDocs[key]
	//}
	return qResults, err
}

// Read an object by its ID
func (store *JsonStore) Read(args *svc.Read_Args) (*svc.ResultValue, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	res := &svc.ResultValue{}

	doc, ok := store.docs[args.Key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	jsonDoc, _ := json.Marshal(doc)
	res.JsonDoc = string(jsonDoc)
	return res, nil
}

// Remove a document from the store
// Also succeeds if the document doesn't exist
func (store *JsonStore) Remove(_ context.Context, args *svc.Remove_Args) (*emptypb.Empty, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.docs, args.Key)
	store.updateCount++
	return nil, nil
}

// RemoveAll empties the store. Intended for testing.
//func (store *JsonStore) RemoveAll(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
//	store.mutex.Lock()
//	defer store.mutex.Unlock()
//	store.docs = make(map[string]interface{})
//	store.updateCount++
//	return nil, nil
//}
//
//// Size returns the number of items in the store
//func (store *JsonStore) Size(context.Context, *emptypb.Empty) (*svc.SizeResult, error) {
//	store.mutex.RLock()
//	defer store.mutex.RUnlock()
//	res := &svc.SizeResult{
//		Count: int32(len(store.docs)),
//	}
//	return res, nil
//}

// SetWriteDelay sets the delay for writing after a change
func (store *JsonStore) SetWriteDelay(delay time.Duration) {
	store.writeDelay = delay
}

// Start the store background loop for saving changes
func (store *JsonStore) Start() error {
	logrus.Infof("JsonStore.Open: Opening store from '%s'", store.storePath)
	var err error
	store.mutex.Lock()
	defer store.mutex.Unlock()

	go store.autoSaveLoop()
	return err
}

// Stop the background update.
// If any changes are remaining then write to disk now.
func (store *JsonStore) Stop() error {
	var err error
	logrus.Infof("Ending background loop")
	store.backgroundLoopEnding <- true

	// wait for the background loop to end
	<-store.backgroundLoopEnded

	store.mutex.Lock()
	defer store.mutex.Unlock()

	// flush any remaining changes
	if store.updateCount > 0 {
		err = writeStoreFile(store.storePath, store.docs)
	}
	return err
}

// Write a document to the store. If the document exists it is replaced.
//
// Write operations are non-blocking and handled as follows:
//  1. Unmarshal to validate the json
//  2. Lock the store while storing both the json text and object in the index and increase the change count.
//
//  A background process periodically checks the change count. When increased:
//  1. Lock the store while copying the index. Unlock when done.
//  2. Stream the in-memory json documents to a temp file.
//  3. If success, move the temp file to the store file using the OS atomic move operation.
//
func (store *JsonStore) Write(_ context.Context, args *svc.Write_Args) (*emptypb.Empty, error) {
	if args.JsonDoc == "" || args.Key == "" {
		err := fmt.Errorf("JsonStore.Write: key='%s' parameter error", args.Key)
		return nil, err
	}
	// validate the json
	var jsonObj interface{}
	err := json.Unmarshal([]byte(args.JsonDoc), &jsonObj)
	if err != nil {
		err := fmt.Errorf("JsonStore.Write: key='%s' Invalid json: ", err)
		return nil, err
	}
	// store the document and object
	store.mutex.Lock()
	defer store.mutex.Unlock()
	//store.jsonDocs[args.Key] = args.JsonDoc
	store.docs[args.Key] = jsonObj
	store.updateCount++
	return nil, nil
}

// NewJsonStore creates a JSON store instance and load it with saved documents.
// Run Start to start the background loop and Stop to end it.
//
//  storeFile path to storage file
//  writeDelayMsec max delay before flushing changes to disk. Default 3000
func NewJsonStore(storePath string) (store *JsonStore, err error) {
	writeDelay := time.Duration(3000) * time.Millisecond
	store = &JsonStore{
		//jsonDocs:             make(map[string]string),
		docs:                 make(map[string]interface{}),
		storePath:            storePath,
		limit:                MaxLimit,
		backgroundLoopEnding: make(chan bool),
		backgroundLoopEnded:  make(chan bool),
		mutex:                sync.RWMutex{},
		writeDelay:           writeDelay,
	}
	store.docs, err = openStoreFile(store.storePath)
	return store, err
}
