package kvstore_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/hiveot/hub/internal/kvstore"
	"github.com/hiveot/hub.grpc/go/thing"

	"github.com/hiveot/hub.go/pkg/vocab"
)

const (
	doc1ID        = "doc1"
	doc2ID        = "doc2"
	jsonStoreFile = "/tmp/hivehub-kvstore_test.json"
)

const doc1 = `{
  "id": "doc1",
  "title": "Title of doc 1",
  "@type": "sensor",
  "properties": 
     { "title": {
         "name": "title1" 
       }
     }
}`
const doc2 = `{
  "id": "doc2",
  "title": "Title of doc 2",
  "properties": [
     { "title": "title2" }
  ]
}`

// Create a TD document
func createTD(id string) *thing.ThingDescription {
	td := &thing.ThingDescription{
		Id:         id,
		Title:      fmt.Sprintf("test TD %s", id),
		AtType:     string(vocab.DeviceTypeSensor),
		Properties: make(map[string]*thing.PropertyAffordance),
		Events:     make(map[string]*thing.EventAffordance),
	}
	td.Properties[vocab.PropNameTitle] = &thing.PropertyAffordance{
		Title:       "Sensor title",
		Description: "This is a smart sensor",
		Type:        vocab.WoTDataTypeString,
		Default:     "Default value",
	}
	td.Properties[vocab.PropNameSoftwareVersion] = &thing.PropertyAffordance{
		Title:       "Version",
		Description: "Embedded firmware",
		Type:        vocab.WoTDataTypeString,
		Default:     "Default value",
		Const:       "v1.0",
	}
	td.Events[vocab.PropNameValue] = &thing.EventAffordance{
		Title:       "Event 1",
		Description: "Name of this event",
		Data: &thing.DataSchema{
			Type:        vocab.WoTDataTypeString,
			Const:       "123",
			Title:       "Event name data",
			Description: "String with friendly name of the event"},
	}
	td.Events[vocab.PropNameBattery] = &thing.EventAffordance{
		Title: "Event 2",
		Data: &thing.DataSchema{
			Type:        vocab.WoTDataTypeInteger,
			Title:       "Battery level",
			Unit:        vocab.UnitNamePercent,
			Description: "Battery level update in % of device"},
	}
	return td
}

// AddDocs adds documents doc1, doc2 and given nr additional docs
func addDocs(store *kvstore.KVStore, count int) error {
	// these docs have values used for testing
	err := store.Write(doc1ID, doc1)
	err = store.Write(doc2ID, doc2)
	if err != nil {
		return err
	}
	// fill remainder with generated docs
	// don't sort order of id
	for i := count; i > 2; i-- {
		id := fmt.Sprintf("doc-%d", i)
		td := createTD(id)
		// td is a protobuf document with protbuf json annotations
		jsonDoc, _ := protojson.Marshal(td)
		_ = store.Write(id, string(jsonDoc))
	}
	return nil
}

func createNewStore() (*kvstore.KVStore, error) {
	_ = os.Remove(jsonStoreFile)
	store, err := kvstore.NewKVStore(jsonStoreFile)
	return store, err
}

// Generic directory store testcases
func TestStartStop(t *testing.T) {
	store, err := createNewStore()
	require.NoError(t, err)
	store.SetWriteDelay(10 * time.Millisecond)

	err = store.Start()
	time.Sleep(time.Millisecond * 20)
	assert.NoError(t, err)
	err = store.Stop()
	assert.NoError(t, err)

	// store should now exist and restart succeed
	assert.FileExists(t, jsonStoreFile)
	store, err = kvstore.NewKVStore(jsonStoreFile)
	err = store.Start()
	time.Sleep(time.Millisecond * 20)
	assert.NoError(t, err)
	err = store.Stop()
	assert.NoError(t, err)

	// corrupted file should not stop the service (or should it?)
	_ = ioutil.WriteFile(jsonStoreFile, []byte("-invalid json"), 0600)
	store, err = kvstore.NewKVStore(jsonStoreFile)
	assert.NoError(t, err)
}

func TestCreateStoreBadFolder(t *testing.T) {
	filename := "/folder/does/not/exist/store.json"
	_, err := kvstore.NewKVStore(filename)
	assert.Error(t, err)
}

func TestCreateStoreReadOnlyFolder(t *testing.T) {
	filename := "/var/jsonstore.json"
	_, err := kvstore.NewKVStore(filename)
	assert.Error(t, err)
}

func TestCreateStoreCantReadFile(t *testing.T) {
	filename := "/var"
	_, err := kvstore.NewKVStore(filename)
	assert.Error(t, err)
	assert.Error(t, err)
}

func TestWriteRead(t *testing.T) {
	newTitle := "new title"

	store, err := createNewStore()
	assert.NoError(t, err)
	store.SetWriteDelay(time.Millisecond * 10)
	require.NoError(t, err)
	_ = store.Start()
	err = addDocs(store, 0)
	require.NoError(t, err)

	// Replace
	doc3 := fmt.Sprintf(`{"id":"%s","title":"%s"}`, doc1ID, newTitle)
	err = store.Write(doc1ID, doc3)
	assert.NoError(t, err)

	// Read
	resp, err := store.Read(doc1ID)
	assert.NoError(t, err)
	assert.Equal(t, doc3, resp)

	// Delete
	store.Remove(doc1ID)
	assert.NoError(t, err)
	err = store.Stop()
	assert.NoError(t, err)

	// Read again should fail
	_, err = store.Read(doc1ID)
	assert.Error(t, err)
}

func TestWriteBadData(t *testing.T) {
	store, err := createNewStore()
	require.NoError(t, err)
	// not json
	err = store.Write(doc1ID, "not-json")
	assert.NoError(t, err)
	// missing key
	err = store.Write("", "{}")
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	var total = 100
	store, err := createNewStore()
	require.NoError(t, err)
	err = addDocs(store, total)
	require.NoError(t, err)

	resp, err := store.List(nil, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, total, len(resp))
}

func TestListWithLimit(t *testing.T) {
	var docs []interface{}
	const total = 100

	store, err := createNewStore()
	require.NoError(t, err)
	err = addDocs(store, total)
	require.NoError(t, err)

	resp, err := store.List(nil, 10, 20)
	require.NoError(t, err)
	assert.Equal(t, 10, len(resp))

	// based on count of 100
	resp, err = store.List(nil, 20, total-10)
	require.NoError(t, err)
	assert.Equal(t, 10, len(resp))

	// based on count of 100
	resp, err = store.List(nil, 0, 105)
	require.NoError(t, err)
	assert.Equal(t, 0, len(docs))
}

func TestQuery(t *testing.T) {
	store, err := createNewStore()
	require.NoError(t, err)
	err = addDocs(store, 20)
	require.NoError(t, err)

	// filter on key 'id' == doc1
	//args := &svc.Query_Args{JsonPathQuery: `$[?(@.id=="doc1")]`}
	jsonPath := `$[?(@.id=="doc1")]`
	resp, err := store.Query(jsonPath, 0, 0, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, resp)

	// regular nested filter comparison. note that a TD does not hold values
	jsonPath = `$[?(@.properties.title.name=="title1")]`
	resp, err = store.Query(jsonPath, 0, 0, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, resp)

	// filter with nested notation. some examples that return a list of TDs matching the filter
	//res, err = fileStore.Query(`$[?(@.properties.title.value=="title1")]`, 0, 0)
	// res, err = fileStore.Query(`$[?(@.*.title.value=="title1")]`, 0, 0)
	// res, err = fileStore.Query(`$[?(@['properties']['title']['value']=="title1")]`, 0, 0)
	jsonPath = `$[?(@..title.name=="title1")]`
	resp, err = store.Query(jsonPath, 0, 0, nil)
	assert.NoError(t, err)

	// these only return the properties - not good
	// res, err = fileStore.Query(`$.*.properties[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
	//res, err = fileStore.Query(`$.*.*[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
	// res, err = fileStore.Query(`$[?(@...value=="title1")]`, 0, 0)
	assert.NotEmpty(t, resp)

	// filter with bracket notation
	jsonPath = `$[?(@["id"]=="doc1")]`
	resp, err = store.Query(jsonPath, 0, 0, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, resp)

	// filter with bracket notation and current object literal (for search @type)
	// only supported by: ohler55/ojg
	jsonPath = `$[?(@['@type']=="sensor")]`
	resp, err = store.Query(jsonPath, 0, 0, nil)
	assert.NoError(t, err)
	assert.Greater(t, len(resp), 1)

	// bad query expression
	jsonPath = `$[?(.id=="doc1")]`
	resp, err = store.Query(jsonPath, 0, 0, nil)
	assert.Error(t, err)
}

// tests to figure out how to use jp parse with bracket notation
func TestQueryBracketNotationA(t *testing.T) {
	store := make(map[string]interface{})
	query1 := `$[?(@['type']=="type1")]`
	query2 := `$[?(@['@type']=="sensor")]`

	jsonDoc := `{
		"thing1": {
			"id": "thing1",
			"type": "type1",
			"@type": "sensor",
			"properties": {
				"title": "title1"
			}
		},
		"thing2": {
			"id": "thing2",
			"type": "type2",
			"@type": "sensor",
			"properties": {
				"title": "title1"
			}
		}
	}`

	err := json.Unmarshal([]byte(jsonDoc), &store)
	assert.NoError(t, err)

	jpExpr, err := jp.ParseString(query1)
	assert.NoError(t, err)
	result := jpExpr.Get(store)
	assert.NotEmpty(t, result)

	jpExpr, err = jp.ParseString(query2)
	assert.NoError(t, err)
	result = jpExpr.Get(store)
	assert.NotEmpty(t, result)
}

// tests to figure out how to use jp parse with bracket notation
func TestQueryBracketNotationB(t *testing.T) {
	//store := make(map[string]interface{})
	queryString := "$[?(@['@type']==\"sensor\")]"
	id1 := "thing1"
	id2 := "thing2"
	td1 := thing.ThingDescription{
		Id:         id1,
		Title:      "test TD 1",
		AtType:     string(vocab.DeviceTypeSensor),
		Properties: make(map[string]*thing.PropertyAffordance),
	}
	//td1 := thing.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
	td1.Properties[vocab.PropNameTitle] = &thing.PropertyAffordance{
		Title: "Sensor title",
		Type:  vocab.WoTDataTypeString,
	}
	td1.Properties[vocab.PropNameValue] = &thing.PropertyAffordance{
		Title: "Sensor value",
		Type:  vocab.WoTDataTypeNumber,
	}

	td2 := thing.ThingDescription{
		Id:         id2,
		Title:      "test TD 2",
		AtType:     string(vocab.DeviceTypeSensor),
		Properties: make(map[string]*thing.PropertyAffordance),
	}
	td2.Properties[vocab.PropNameTitle] = &thing.PropertyAffordance{
		Title: "The switch",
		Type:  vocab.WoTDataTypeBool,
	}

	store, err := createNewStore()
	require.NoError(t, err)

	//td1json, err := json.MarshalIndent(td1, "", "")
	td1json, err := protojson.Marshal(&td1)
	td2json, err := protojson.Marshal(&td2)
	_ = store.Write(id1, string(td1json))
	err = store.Write(id2, string(td2json))
	assert.NoError(t, err)

	// query returns 2 sensors.
	resp, err := store.Query(queryString, 0, 0, nil)
	require.NoError(t, err)
	require.Equal(t, 2, len(resp))

	var readTD1 thing.ThingDescription
	err = protojson.Unmarshal([]byte(resp[0]), &readTD1)
	require.NoError(t, err)
	read1type := readTD1.AtType
	assert.Equal(t, string(vocab.DeviceTypeSensor), read1type)
}

// test query with reduced list of IDs
func TestQueryFiltered(t *testing.T) {
	queryString := "$..id"

	store, err := createNewStore()
	require.NoError(t, err)
	_ = addDocs(store, 10)

	// result of a normal query
	resp, err := store.Query(queryString, 0, 0, nil)
	require.NoError(t, err)
	assert.Equal(t, 10, len(resp))
}

// crude read/write performance test
func TestReadWritePerf(t *testing.T) {
	const iterations = 100000
	store, err := createNewStore()
	require.NoError(t, err)
	_ = store.Start()
	err = addDocs(store, 0)
	require.NoError(t, err)
	tdDoc := createTD("id")
	doc3, _ := json.Marshal(tdDoc)

	t1 := time.Now()
	var i int
	for i = 0; i < iterations; i++ {
		// write
		docID := fmt.Sprintf("doc-%d", i)
		tdDoc.Id = docID
		//doc3, _ := json.Marshal(tdDoc)
		//doc3 := fmt.Sprintf(`{"id":"%s","title":"%s-%d"}`, docID, "Hello ", i)
		err = store.Write(docID, string(doc3))
		assert.NoError(t, err)
	}
	d1 := time.Since(t1)
	t2 := time.Now()
	for i = 0; i < iterations; i++ {
		// Read
		docID := fmt.Sprintf("doc-%d", i)
		resp, err := store.Read(docID)
		assert.NoError(t, err)
		require.NotEmpty(t, resp)
	}
	d2 := time.Since(t2)
	t3 := time.Now()
	resp, err := store.List(nil, 1000, 0)
	assert.NoError(t, err)
	require.NotEmpty(t, resp)
	d3 := time.Since(t3)
	_ = store.Stop()
	// 100K writes: 93 msec, 100K reads 77msec, List first 1K: 31msec
	// 1M writes 1.1 sec, 1M reads 0.8 sec. list first 1K: 0.5 sec
	// 2M writes 2.2 sec, 1M reads 2.8 sec. list first 1K: 1.1 sec
	logrus.Infof("TestQuery, %d write: %d msec; %d reads: %d msec. List first 1K: %d msec",
		iterations, d1.Milliseconds(), iterations, d2.Milliseconds(), d3.Milliseconds())
}

// crude list performance test
func TestPerfList(t *testing.T) {
	const iterations = 1000
	const dataSize = 1000
	const listLimit = 1000

	store, err := createNewStore()
	require.NoError(t, err)
	_ = store.Start()
	err = addDocs(store, dataSize)
	require.NoError(t, err)

	t1 := time.Now()
	var i int
	for i = 0; i < iterations; i++ {
		// List
		resp2, err := store.List(nil, listLimit, 0)
		assert.NoError(t, err)
		require.NotEmpty(t, resp2)
	}
	d1 := time.Since(t1)
	_ = store.Stop()
	logrus.Infof("TestList, %d runs: %d msec.", i, d1.Milliseconds())
}

// crude query performance test
func TestPerfQuery(t *testing.T) {
	const iterations = 1000
	const dataSize = 1000
	jsonPath := `$[?(@.properties.title.name=="title1")]`

	store, err := createNewStore()
	require.NoError(t, err)
	_ = store.Start()
	err = addDocs(store, dataSize)
	require.NoError(t, err)

	// skip first run as it caches docs
	_, _ = store.Query(jsonPath, 0, 0, nil)
	t1 := time.Now()
	var i int
	for i = 0; i < iterations; i++ {
		// filter on key 'id' == doc1
		//jsonPath := `$[?(@.id=="doc1")]`,
		//JsonPathQuery: `$..id`,
		resp, err := store.Query(jsonPath, 0, 0, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, resp)
	}
	d1 := time.Since(t1)
	_ = store.Stop()
	// 1000 runs on 1K records: 1.4 sec
	// 100 runs on 10K records: 2.0 sec
	// 10 runs on 100K records: 2.5 sec
	// 1 run on 1M records: 3.6 sec
	logrus.Infof("TestQuery, %d runs: %d msec.", i, d1.Milliseconds())
}
