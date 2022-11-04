package bucketstore_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hiveot/hub.go/pkg/logging"
	"github.com/hiveot/hub.go/pkg/thing"
	"github.com/hiveot/hub.go/pkg/vocab"
	"github.com/hiveot/hub/internal/bucketstore"
	"github.com/hiveot/hub/internal/bucketstore/kvmem"
)

var testBucketID = "default"

var testBackendType = bucketstore.BackendKVStore
var testBackendPath = "/tmp/test-kvstore.json"

//var testBackendType = bucketstore.BackendBBolt
//var testBackendPath = "/tmp/test-bolt.db"

//var testBackendType = bucketstore.BackendPebble
//var testBackendPath = "/tmp/test-pebble/"

const (
	doc1ID = "doc1"
	doc2ID = "doc2"
)

var doc1 = []byte(`{
  "id": "doc1",
  "title": "Title of doc 1",
  "@type": "sensor",
  "properties": 
     { "title": {
         "name": "title1" 
       }
     }
}`)
var doc2 = []byte(`{
  "id": "doc2",
  "title": "Title of doc 2",
  "properties": [
     { "title": "title2" }
  ]
}`)

// Create the bucket store using the backend
func createNewStore(backend string, storePath string) (store bucketstore.IBucketStore, err error) {
	_ = os.Remove(storePath)
	if backend == bucketstore.BackendKVStore {
		store = kvmem.NewKVStore("testclientkv", storePath)
		//} else if backend == bucketstore.BackendBBolt {
		//	store = bolts.NewBoltBucketStore("testclientbbolt", storePath)
		//} else if backend == bucketstore.BackendPebble {
		//	store = pebbles.NewPebbleBucketStore("testclientpebble", storePath)
	}
	return store, err
}

// Create a TD document
func createTD(id string) *thing.ThingDescription {
	td := &thing.ThingDescription{
		ID:         id,
		Title:      fmt.Sprintf("test TD %s", id),
		AtType:     string(vocab.DeviceTypeSensor),
		Properties: make(map[string]*thing.PropertyAffordance),
		Events:     make(map[string]*thing.EventAffordance),
	}
	td.Properties[vocab.PropNameTitle] = &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Title:       "Sensor title",
			Description: "This is a smart sensor",
			Type:        vocab.WoTDataTypeString,
			Default:     "Default value",
		},
	}
	td.Properties[vocab.PropNameSoftwareVersion] = &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Title:       "Version",
			Description: "Embedded firmware",
			Type:        vocab.WoTDataTypeString,
			Default:     "Default value",
			Const:       "v1.0",
		},
	}
	td.Events[vocab.PropNameValue] = &thing.EventAffordance{
		Title:       "Event 1",
		Description: "Name of this event",
		Data: thing.DataSchema{
			Type:        vocab.WoTDataTypeString,
			Const:       "123",
			Title:       "Event name data",
			Description: "String with friendly name of the event"},
	}
	td.Events[vocab.PropNameBattery] = &thing.EventAffordance{
		Title: "Event 2",
		Data: thing.DataSchema{
			Type:        vocab.WoTDataTypeInteger,
			Title:       "Battery level",
			Unit:        vocab.UnitNamePercent,
			Description: "Battery level update in % of device"},
	}
	return td
}

// AddDocs adds documents doc1, doc2 and given nr additional docs
func addDocs(store bucketstore.IBucketStore, count int) error {
	// these docs have values used for testing
	err := store.Set(testBucketID, doc1ID, doc1)
	err = store.Set(testBucketID, doc2ID, doc2)
	if err != nil {
		return err
	}
	docs := make(map[string][]byte)
	// TODO: use SetMultiple for performance
	// fill remainder with generated docs
	// don't sort order of id
	for i := count; i > 2; i-- {
		id := fmt.Sprintf("addDocs-%6d", i)
		td := createTD(id)
		_ = td
		jsonDoc := []byte("hello world")
		jsonDoc, _ = json.Marshal(td) // 900msec
		docs[id] = jsonDoc
		if err != nil {
			panic(fmt.Sprintf("Unmarshal failed: %s", err))
		}
	}
	err = store.SetMultiple(testBucketID, docs)
	return nil
}

func TestMain(m *testing.M) {
	logging.SetLogging("info", "")

	res := m.Run()
	os.Exit(res)
}

// Generic directory store testcases
func TestStartStop(t *testing.T) {
	store, err := createNewStore(testBackendType, testBackendPath)
	require.NoError(t, err)

	err = store.Open()
	assert.NoError(t, err)
	err = store.Close()
	assert.NoError(t, err)

	//// store should now exist and reopen succeed
	//assert.FileExists(t, bucketstoreFile)
	store, err = createNewStore(testBackendType, testBackendPath)
	//store, err = kvstore.NewKVStore(jsonStoreFile)
	err = store.Open()
	//time.Sleep(time.Millisecond * 20)
	assert.NoError(t, err)
	err = store.Close()
	assert.NoError(t, err)

	// corrupted file should not stop the service (or should it?)
	//_ = ioutil.WriteFile(jsonStoreFile, []byte("-invalid json"), 0600)
	//store, err = kvstore.NewKVStore(jsonStoreFile)
	//assert.NoError(t, err)
}

func TestCreateStoreBadFolder(t *testing.T) {
	filename := "/folder/does/not/exist/store.json"
	store, err := createNewStore(testBackendType, filename)
	assert.NoError(t, err)
	err = store.Open()
	assert.Error(t, err)
}

func TestCreateStoreReadOnlyFolder(t *testing.T) {
	filename := "/var/jsonstore.json"
	store, err := createNewStore(testBackendType, filename)
	err = store.Open()
	assert.Error(t, err)
}

func TestCreateStoreCantReadFile(t *testing.T) {
	filename := "/var"
	store, err := createNewStore(testBackendType, filename)
	err = store.Open()
	assert.Error(t, err)
}

func TestWriteRead(t *testing.T) {
	const id1 = "id1"
	const id5 = "id5"
	const id22 = "id22"

	store, err := createNewStore(testBackendType, testBackendPath)
	assert.NoError(t, err)
	err = store.Open()
	require.NoError(t, err)
	err = addDocs(store, 3)
	require.NoError(t, err)

	// write docs
	td1 := createTD(id1)
	td1json, _ := json.Marshal(td1)
	err = store.Set(testBucketID, id1, td1json)
	assert.NoError(t, err)
	td22 := createTD(id22)
	td22json, _ := json.Marshal(td22)
	err = store.Set(testBucketID, id22, td22json)
	assert.NoError(t, err)
	td5 := createTD(id5)
	td5json, _ := json.Marshal(td5)
	err = store.Set(testBucketID, id5, td5json)
	assert.NoError(t, err)

	// kvstore writes to backend in autosave loop
	// needs to be tested
	time.Sleep(time.Second * 4)

	// close and reopen
	err = store.Close()
	assert.NoError(t, err)
	time.Sleep(time.Second)
	err = store.Open()
	require.NoError(t, err)

	// Read and compare
	resp, found, err := store.Get(testBucketID, id22)
	if assert.True(t, found) {
		assert.Equal(t, td22json, resp)
	}
	resp, found, err = store.Get(testBucketID, id1)
	if assert.True(t, found) {
		assert.Equal(t, td1json, resp)
	}
	resp, found, err = store.Get(testBucketID, id5)
	if assert.True(t, found) {
		assert.Equal(t, td5json, resp)
	}
	// Delete
	err = store.Delete(testBucketID, id1)
	assert.NoError(t, err)
	err = store.Close()
	assert.NoError(t, err)

	// Read again should fail
	// (pebble throws a panic :(
	//_, err = store.Get(testBucketID, doc1ID)
	//assert.Error(t, err)
}

func TestWriteBadData(t *testing.T) {
	store, err := createNewStore(testBackendType, testBackendPath)
	require.NoError(t, err)
	err = store.Open()
	defer store.Close()
	// not json
	err = store.Set(testBucketID, doc1ID, []byte("not-json"))
	assert.NoError(t, err)
	// missing key
	err = store.Set(testBucketID, "", []byte("{}"))
	assert.Error(t, err)

}

func TestWriteReadMultiple(t *testing.T) {
	const id1 = "id1"
	const id5 = "id5"
	const id22 = "id22"
	docs := make(map[string][]byte)

	store, err := createNewStore(testBackendType, testBackendPath)
	assert.NoError(t, err)
	err = store.Open()
	require.NoError(t, err)
	err = addDocs(store, 3)
	require.NoError(t, err)

	// write docs
	docs[id1], _ = json.Marshal(createTD(id1))
	docs[id22], _ = json.Marshal(createTD(id22))
	docs[id5], _ = json.Marshal(createTD(id5))
	err = store.SetMultiple(testBucketID, docs)
	assert.NoError(t, err)

	// Read and compare

	resp, err := store.GetMultiple(testBucketID, []string{id22, id1, id5})
	assert.NoError(t, err)
	assert.Equal(t, docs[id1], resp[id1])
	assert.Equal(t, docs[id5], resp[id5])
	assert.Equal(t, docs[id22], resp[id22])

	// Delete
	err = store.Delete(testBucketID, id1)
	assert.NoError(t, err)
	resp2, err := store.GetMultiple(testBucketID, []string{id22, id1, id5})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp2))

	err = store.Close()
	assert.NoError(t, err)

	// Read again should fail
	// (pebble throws a panic :(
	//_, err = store.Get(testBucketID, doc1ID)
	//assert.Error(t, err)
}

func TestSeek(t *testing.T) {
	const count = 10000
	const seekCount = 3000
	const base = 4200

	store, err := createNewStore(testBackendType, testBackendPath)
	require.NoError(t, err)
	err = store.Open()
	require.NoError(t, err)
	defer store.Close()

	err = addDocs(store, count)
	require.NoError(t, err)

	// This format must match that of addDocs
	id := fmt.Sprintf("addDocs-%6d", base)

	t1 := time.Now()
	// seek forward
	cursor, err := store.Seek(testBucketID, id)
	require.NoError(t, err)
	assert.Equal(t, id, cursor.Key())
	assert.NotEmpty(t, cursor.Value())

	for i := 0; i < seekCount; i++ {
		k, v := cursor.Next()

		id := fmt.Sprintf("addDocs-%6d", base+i+1)
		require.Equal(t, id, k)
		assert.NotEmpty(t, v)
	}
	d1 := time.Now().Sub(t1)

	// seek beyond should succeed
	for cursor.Key() != "" {
		cursor.Next()
	}
	cursor.Close()

	// seek backwards
	id = fmt.Sprintf("addDocs-%6d", base)
	cursor, err = store.Seek(testBucketID, id)
	require.NoError(t, err)
	assert.Equal(t, id, cursor.Key())

	for i := 0; i < seekCount; i++ {
		k, v := cursor.Prev()

		id := fmt.Sprintf("addDocs-%6d", base-i-1)
		require.Equal(t, id, k)
		assert.NotEmpty(t, v)
	}
	cursor.Close()
	logrus.Infof("1 seek and %d iterations with %d documents: %dmsec", seekCount, count, d1.Milliseconds())

}

//func TestList(t *testing.T) {
//	var total = 100
//	store, err := createNewStore(testBackendType, testBackendFile)
//	require.NoError(t, err)
//	err = addDocs(store, total)
//	require.NoError(t, err)
//
//	resp, err := store.List(0, 0, nil)
//	require.NoError(t, err)
//	assert.Equal(t, total, len(resp))
//}
//
//func TestListWithLimit(t *testing.T) {
//	var docs []interface{}
//	const total = 100
//
//	store, err := createNewStore(testBackendType, testBackendFile)
//	require.NoError(t, err)
//	err = addDocs(store, total)
//	require.NoError(t, err)
//
//	resp, err := store.List(10, 20, nil)
//	require.NoError(t, err)
//	assert.Equal(t, 10, len(resp))
//
//	// based on count of 100
//	resp, err = store.List(20, total-10, nil)
//	require.NoError(t, err)
//	assert.Equal(t, 10, len(resp))
//
//	// based on count of 100
//	resp, err = store.List(0, 105, nil)
//	require.NoError(t, err)
//	assert.Equal(t, 0, len(docs))
//}

//func TestQuery(t *testing.T) {
//	store, err := createNewStore()
//	require.NoError(t, err)
//	err = addDocs(store, 20)
//	require.NoError(t, err)
//
//	// filter on key 'id' == doc1
//	//args := &svc.Query_Args{JsonPathQuery: `$[?(@.id=="doc1")]`}
//	jsonPath := `$[?(@.id=="doc1")]`
//	resp, err := store.Query(jsonPath, 0, 0, nil)
//	require.NoError(t, err)
//	assert.NotEmpty(t, resp)
//
//	// regular nested filter comparison. note that a TD does not hold values
//	jsonPath = `$[?(@.properties.title.name=="title1")]`
//	resp, err = store.Query(jsonPath, 0, 0, nil)
//	require.NoError(t, err)
//	assert.NotEmpty(t, resp)
//
//	// filter with nested notation. some examples that return a list of TDs matching the filter
//	//res, err = fileStore.Query(`$[?(@.properties.title.value=="title1")]`, 0, 0)
//	// res, err = fileStore.Query(`$[?(@.*.title.value=="title1")]`, 0, 0)
//	// res, err = fileStore.Query(`$[?(@['properties']['title']['value']=="title1")]`, 0, 0)
//	jsonPath = `$[?(@..title.name=="title1")]`
//	resp, err = store.Query(jsonPath, 0, 0, nil)
//	assert.NoError(t, err)
//
//	// these only return the properties - not good
//	// res, err = fileStore.Query(`$.*.properties[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
//	//res, err = fileStore.Query(`$.*.*[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
//	// res, err = fileStore.Query(`$[?(@...value=="title1")]`, 0, 0)
//	assert.NotEmpty(t, resp)
//
//	// filter with bracket notation
//	jsonPath = `$[?(@["id"]=="doc1")]`
//	resp, err = store.Query(jsonPath, 0, 0, nil)
//	require.NoError(t, err)
//	assert.NotEmpty(t, resp)
//
//	// filter with bracket notation and current object literal (for search @type)
//	// only supported by: ohler55/ojg
//	jsonPath = `$[?(@['@type']=="sensor")]`
//	resp, err = store.Query(jsonPath, 0, 0, nil)
//	assert.NoError(t, err)
//	assert.Greater(t, len(resp), 1)
//
//	// bad query expression
//	jsonPath = `$[?(.id=="doc1")]`
//	resp, err = store.Query(jsonPath, 0, 0, nil)
//	assert.Error(t, err)
//}

// tests to figure out how to use jp parse with bracket notation
//func TestQueryBracketNotationA(t *testing.T) {
//	store := make(map[string]interface{})
//	query1 := `$[?(@['type']=="type1")]`
//	query2 := `$[?(@['@type']=="sensor")]`
//
//	jsonDoc := `{
//		"thing1": {
//			"id": "thing1",
//			"type": "type1",
//			"@type": "sensor",
//			"properties": {
//				"title": "title1"
//			}
//		},
//		"thing2": {
//			"id": "thing2",
//			"type": "type2",
//			"@type": "sensor",
//			"properties": {
//				"title": "title1"
//			}
//		}
//	}`
//
//	err := json.Unmarshal([]byte(jsonDoc), &store)
//	assert.NoError(t, err)
//
//	jpExpr, err := jp.ParseString(query1)
//	assert.NoError(t, err)
//	result := jpExpr.Get(store)
//	assert.NotEmpty(t, result)
//
//	jpExpr, err = jp.ParseString(query2)
//	assert.NoError(t, err)
//	result = jpExpr.Get(store)
//	assert.NotEmpty(t, result)
//}

// tests to figure out how to use jp parse with bracket notation
//func TestQueryBracketNotationB(t *testing.T) {
//	//store := make(map[string]interface{})
//	queryString := "$[?(@['@type']==\"sensor\")]"
//	id1 := "thing1"
//	id2 := "thing2"
//	td1 := thing.ThingDescription{
//		ID:         id1,
//		Title:      "test TD 1",
//		AtType:     string(vocab.DeviceTypeSensor),
//		Properties: make(map[string]*thing.PropertyAffordance),
//	}
//	//td1 := thing.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
//	td1.Properties[vocab.PropNameTitle] = &thing.PropertyAffordance{
//		DataSchema: thing.DataSchema{
//			Title: "Sensor title",
//			Type:  vocab.WoTDataTypeString,
//		},
//	}
//	td1.Properties[vocab.PropNameValue] = &thing.PropertyAffordance{
//		DataSchema: thing.DataSchema{
//			Title: "Sensor value",
//			Type:  vocab.WoTDataTypeNumber,
//		},
//	}
//
//	td2 := thing.ThingDescription{
//		ID:         id2,
//		Title:      "test TD 2",
//		AtType:     string(vocab.DeviceTypeSensor),
//		Properties: make(map[string]*thing.PropertyAffordance),
//	}
//	td2.Properties[vocab.PropNameTitle] = &thing.PropertyAffordance{
//		DataSchema: thing.DataSchema{
//			Title: "The switch",
//			Type:  vocab.WoTDataTypeBool,
//		},
//	}
//
//	store, err := createNewStore()
//	require.NoError(t, err)
//
//	//td1json, err := json.MarshalIndent(td1, "", "")
//	td1json, err := json.Marshal(&td1)
//	td2json, err := json.Marshal(&td2)
//	_ = store.Write(id1, string(td1json))
//	err = store.Write(id2, string(td2json))
//	assert.NoError(t, err)
//
//	// query returns 2 sensors.
//	resp, err := store.Query(queryString, 0, 0, nil)
//	require.NoError(t, err)
//	require.Equal(t, 2, len(resp))
//
//	var readTD1 thing.ThingDescription
//	err = json.Unmarshal([]byte(resp[0]), &readTD1)
//	require.NoError(t, err)
//	read1type := readTD1.AtType
//	assert.Equal(t, string(vocab.DeviceTypeSensor), read1type)
//}

// test query with reduced list of IDs
//func TestQueryFiltered(t *testing.T) {
//	queryString := "$..id"
//
//	store, err := createNewStore()
//	require.NoError(t, err)
//	_ = addDocs(store, 10)
//
//	// result of a normal query
//	resp, err := store.Query(queryString, 0, 0, nil)
//	require.NoError(t, err)
//	assert.Equal(t, 10, len(resp))
//}

// crude read/write performance test
func TestReadWritePerf(t *testing.T) {
	const iterations = 1000
	store, err := createNewStore(testBackendType, testBackendPath)
	require.NoError(t, err)
	_ = store.Open()
	t0 := time.Now()
	err = addDocs(store, iterations)
	require.NoError(t, err)
	d0 := time.Now().Sub(t0)

	tdDoc := createTD("id")
	doc3, _ := json.Marshal(tdDoc)
	docData := make(map[string][]byte)

	t1 := time.Now()
	var i int
	for i = 0; i < iterations; i++ {
		// write
		docID := fmt.Sprintf("doc-%d", i)
		tdDoc.ID = docID
		//doc3, _ := json.Marshal(tdDoc)
		//doc3 := fmt.Sprintf(`{"id":"%s","title":"%s-%d"}`, docID, "Hello ", i)
		err = store.Set(testBucketID, docID, doc3)
		docData[docID] = doc3
		assert.NoError(t, err)
	}
	d1 := time.Since(t1)
	t2 := time.Now()
	for i = 0; i < iterations; i++ {
		// Read
		docID := fmt.Sprintf("doc-%d", i)
		resp, found, err := store.Get(testBucketID, docID)
		if assert.True(t, found) {
			assert.NoError(t, err)
			require.NotEmpty(t, resp)
			require.Equal(t, docData[docID], resp)
		}
	}
	d2 := time.Since(t2)
	//t3 := time.Now()
	//resp, err := store.List(1000, 0, nil)
	//assert.NoError(t, err)
	//require.NotEmpty(t, resp)
	//d3 := time.Since(t3)
	_ = store.Close()
	// 100K writes: 93 msec, 100K reads 77msec, List first 1K: 31msec
	// 1M writes 1.1 sec, 1M reads 0.8 sec. list first 1K: 0.5 sec
	// 2M writes 2.2 sec, 1M reads 2.8 sec. list first 1K: 1.1 sec
	//logrus.Infof("TestQuery, %d write: %d msec; %d reads: %d msec. List first 1K: %d msec",
	//	iterations, d1.Milliseconds(), iterations, d2.Milliseconds(), d3.Milliseconds())
	logrus.Infof("'%d' addDocs (incl marshal): %d msec", iterations, d0.Milliseconds())
	logrus.Infof("%d write: %d msec; %d reads: %d msec",
		iterations, d1.Milliseconds(), iterations, d2.Milliseconds())
}

// crude list performance test
//func TestPerfList(t *testing.T) {
//	const iterations = 1000
//	const dataSize = 1000
//	const listLimit = 1000
//
//	store, err := createNewStore()
//	require.NoError(t, err)
//	_ = store.Start()
//	err = addDocs(store, dataSize)
//	require.NoError(t, err)
//
//	t1 := time.Now()
//	var i int
//	for i = 0; i < iterations; i++ {
//		// List
//		resp2, err := store.List(listLimit, 0, nil)
//		assert.NoError(t, err)
//		require.NotEmpty(t, resp2)
//	}
//	d1 := time.Since(t1)
//	_ = store.Stop()
//	logrus.Infof("TestList, %d runs: %d msec.", i, d1.Milliseconds())
//}

// crude query performance test
//func TestPerfQuery(t *testing.T) {
//	const iterations = 1000
//	const dataSize = 1000
//	jsonPath := `$[?(@.properties.title.name=="title1")]`
//
//	store, err := createNewStore()
//	require.NoError(t, err)
//	_ = store.Open()
//	err = addDocs(store, dataSize)
//	require.NoError(t, err)
//
//	// skip first run as it caches docs
//	_, _ = store.Query(jsonPath, 0, 0, nil)
//	t1 := time.Now()
//	var i int
//	for i = 0; i < iterations; i++ {
//		// filter on key 'id' == doc1
//		//jsonPath := `$[?(@.id=="doc1")]`,
//		//JsonPathQuery: `$..id`,
//		resp, err := store.Query(jsonPath, 0, 0, nil)
//		require.NoError(t, err)
//		assert.NotEmpty(t, resp)
//	}
//	d1 := time.Since(t1)
//	_ = store.OpenClose()
//	// 1000 runs on 1K records: 1.4 sec
//	// 100 runs on 10K records: 2.0 sec
//	// 10 runs on 100K records: 2.5 sec
//	// 1 run on 1M records: 3.6 sec
//	logrus.Infof("TestQuery, %d runs: %d msec.", i, d1.Milliseconds())
//}
