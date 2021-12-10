// Package dirfilestore
// This is just a simple in-memory store that is loaded from file and written regularly after updates.
//
// The jsonpath query feature is provided by a library that works with the in-memory object store.
// A good overview of implementations can be found here:
// > https://cburgmer.github.io/json-path-comparison/
//
// Two good options for jsonpath queries:
//  > github.com/ohler55/ojg/jp
//  > github.com/PaesslerAG/jsonpath

package dirfilestore

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/lib/client/pkg/td"
)

// DefaultListLimit is the default max nr of items to return in list
const DefaultListLimit = 100

// DirFileStore is a crude little file based Directory store
// Intended as a testing MVP for the directory service
// Implements the IDirStore interface
type DirFileStore struct {
	docs                 map[string]interface{} // documents by ID
	storePath            string
	mutex                sync.RWMutex
	maxLimit             int // default maximum for the limit value in list and queries
	updateCount          int // nr of updates since last save
	backgroundLoopEnded  chan bool
	backgroundLoopEnding chan bool
}

func createStoreFile(filePath string) error {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// new file
		fp, err2 := os.Create(filePath)
		if err2 == nil {
			fp.Chmod(0600)
			fp.Write([]byte("{}"))
			fp.Close()
		}
		err = err2
	}
	if err != nil {
		logrus.Errorf("createStoreFile: %s", err)
	}
	return err
}

// createStoreFolder creates the folder for the store if it doesn't exist
// The parent folder must exist otherwise this fails
func createStoreFolder(storeFolder string) error {
	_, err := os.Stat(storeFolder)
	if os.IsNotExist(err) {
		err = os.Mkdir(storeFolder, os.ModeDir)
	}
	if err != nil {
		logrus.Errorf("createStoreFolder. Error %s", err)
	}
	return err
}

// readStoreFile loads the store JSON content into a map
func readStoreFile(storePath string) (docs map[string]interface{}, err error) {
	var rawData []byte
	rawData, err = os.ReadFile(storePath)

	if err == nil {
		err = json.Unmarshal(rawData, &docs)
	}

	if err != nil {
		logrus.Infof("DirFileStore.readStoreFile: failed read store '%s', error %s", storePath, err)
		docs = make(map[string]interface{})
	}
	return docs, err
}

// writeStoreFile writes the store to file
func writeStoreFile(storePath string, docs map[string]interface{}) error {
	logrus.Infof("writeStoreFile: Writing Thing Directory to '%s'", storePath)
	rawData, err := json.MarshalIndent(docs, "  ", "  ")
	if err == nil {
		// FIXME: write to temp file and rename instead of writing to file directly. See pw and acl store

		// only allow this user access
		err = os.WriteFile(storePath, rawData, 0600)
	}
	if err != nil {
		logrus.Errorf("DirFileStore.save: Error while saving store to %s: %s", storePath, err)
	}
	return err
}

// AutoSaveLoop periodically saves changes to the directory
func (store *DirFileStore) AutoSaveLoop() {
	logrus.Infof("AutoSaveLoop: autosave loop started")

	defer close(store.backgroundLoopEnded)

	for {
		select {
		case <-store.backgroundLoopEnding:
			logrus.Infof("AutoSaveLoop: autosave loop ended")
			return
		default:
			store.mutex.Lock()
			if store.updateCount > 0 {
				err := writeStoreFile(store.storePath, store.docs)
				if err != nil {
					logrus.Errorf("DirFileStore.AutoSaveLoop: Writing directory to %s failed: %s", store.storePath, err)
				}
				store.updateCount = 0
			}
			store.mutex.Unlock()
			// does this need to be configurable?
			time.Sleep(time.Second * 3)
		}
	}
}

// Close the store
func (store *DirFileStore) Close() {
	logrus.Infof("DirFileStore.Close: Closing directory")
	store.backgroundLoopEnding <- true

	// wait for the background loop to end
	<-store.backgroundLoopEnded

	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.updateCount > 0 {
		err := writeStoreFile(store.storePath, store.docs)
		if err != nil {
			logrus.Errorf("DirFileStore.Close: Writing directory to %s failed: %s", store.storePath, err)
		}
	}
}

// Get a document by its ID
//  id of the thing to look up
// Returns an error if it doesn't exist
func (store *DirFileStore) Get(thingID string) (interface{}, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	doc, ok := store.docs[thingID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return doc, nil
}

// List returns a list of documents
//  offset is the offset in the document list that is sorted by document ID
//  limit is the maximum nr of documents to return or 0 for the default
//  aclFilter filters the things by ID. Use nil to ignore.
// This returns an empty list if offset is equal or larger than the available nr of documents
func (store *DirFileStore) List(offset int, limit int, aclFilter func(thingID string) bool) []interface{} {

	if limit <= 0 {
		limit = DefaultListLimit
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	keyList := make([]string, 0)
	for key := range store.docs {
		if aclFilter == nil || aclFilter(key) {
			keyList = append(keyList, key)
		}
	}
	sort.Strings(keyList)
	sortedDocs := make([]interface{}, len(keyList))

	for index := offset; index < len(keyList) && index < offset+limit; index++ {
		key := keyList[index]
		sortedDocs[index] = store.docs[key]
	}

	return sortedDocs
}

// Open the store
// Returns error if it can't be opened or already open
func (store *DirFileStore) Open() error {
	logrus.Infof("DirFileStore.Open: Opening Thing Directory from '%s'", store.storePath)
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// create the folder if needed
	storeFolder := path.Dir(store.storePath)
	err := createStoreFolder(storeFolder)
	if err == nil {
		err = createStoreFile(store.storePath)
	}
	if err == nil {
		store.docs, err = readStoreFile(store.storePath)
	}
	go store.AutoSaveLoop()
	return err
}

// Patch a document
// Returns an error if it doesn't exist
func (store *DirFileStore) Patch(id string, src map[string]interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	logrus.Infof("DirFileStore.Patch: ID=%s", id)

	if src == nil || id == "" {
		err := fmt.Errorf("DirFileStore.Patch: id='%s' parameter error", id)
		return err
	}
	// the new doc is merged into the original
	dest := store.docs[id].(map[string]interface{})

	if err := mergo.Map(&dest, src, mergo.WithOverride); err != nil {
		return err
	}

	store.updateCount++

	return nil
}

// Query for documents using JSONPATH
// Eg `$[? @.properties.deviceType=="sensor"]`
//  jsonPath contains the query
//  offset contains the offset in the list of results, sorted by ID
//  limit contains the maximum or of responses, 0 for the default 100
func (store *DirFileStore) Query(jsonPath string, offset int, limit int,
	aclFilter func(thingID string) bool) ([]interface{}, error) {
	//  "github.com/PaesslerAG/jsonpath" - just works, amazing!
	// Unfortunately no filter with bracket notation $[? @.["title"]=="my title"]
	// res, err := jsonpath.Get(jsonPath, store.docs)
	// github.com/ohler55/ojg/jp - seems to work with in-mem maps, no @token in bracket notation
	logrus.Infof("DirFileStore.Query: jsonPath='%s', offset=%d, limit=%d", jsonPath, offset, limit)
	jpExpr, err := jp.ParseString(jsonPath)
	if err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = store.maxLimit
	}

	// Before querying the list of available documents must be reduced to those that the
	// user has access to.
	docsToQuery := store.docs
	if aclFilter != nil {
		// the aclFilter must be efficient
		filterResults := make(map[string]interface{})
		for _, tdDoc := range store.docs {
			// this is suppoed to be a valid TD document but need to make sure
			// thingTD, ok := tdDoc.(td.ThingTD)
			thingTD, ok := tdDoc.(map[string]interface{})
			if ok {
				thingID := td.GetID(thingTD)
				if aclFilter(thingID) {
					filterResults[thingID] = thingTD
				}
			}
		}
		docsToQuery = filterResults
	}

	// for key, item := range store.docs {
	// 	logrus.Infof("store item: key='%s', val='%v'", key, item)
	// }
	// Note: store.docs is a map but query returns a list. The key is lost
	// Does the same query always returns the same order?
	// TODO: sort the result to ensure same results when using paging
	qResults := jpExpr.Get(docsToQuery)

	if offset > 0 || limit < len(qResults) {
		qResults = qResults[offset:limit]
	}
	return qResults, err
}

// Remove a document from the store
// Also succeeds if the document doesn't exist
func (store *DirFileStore) Remove(id string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	delete(store.docs, id)
	store.updateCount++
}

// Replace a document
// The document does not have to exist
func (store *DirFileStore) Replace(id string, document map[string]interface{}) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if document == nil || id == "" {
		err := fmt.Errorf("DirFileStore.Replace: id='%s' parameter error", id)
		return err
	}
	store.docs[id] = document
	store.updateCount++
	return nil
}

// Size returns the number of items in the store
func (store *DirFileStore) Size() int {
	return len(store.docs)
}

// NewDirFileStore creates a new directory file store instance
//  filePath path to JSON store file
func NewDirFileStore(jsonFilePath string) *DirFileStore {
	store := DirFileStore{
		docs:                 make(map[string]interface{}),
		storePath:            jsonFilePath,
		maxLimit:             100,
		backgroundLoopEnding: make(chan bool),
		backgroundLoopEnded:  make(chan bool),
	}
	return &store
}
