package jsonstore_test

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

	"github.com/wostzone/wost.grpc/go/thing"

	//"github.com/wostzone/wost-go/pkg/thing"
	"github.com/wostzone/wost-go/pkg/vocab"
	"github.com/wostzone/wost.grpc/go/svc"
	"svc/jsonstore"
)

const (
	doc1ID        = "doc1"
	doc2ID        = "doc2"
	doc1Title     = "Title of doc 1"
	doc2Title     = "Title of doc 2"
	jsonStoreFile = "/tmp/jsonstore_test.json"
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
func addDocs(store *jsonstore.JsonStore, count int) error {
	// these docs have values used for testing
	_, err := store.Write(nil, &svc.Write_Args{Key: doc1ID, JsonDoc: doc1})
	_, err = store.Write(nil, &svc.Write_Args{Key: doc2ID, JsonDoc: doc2})
	if err != nil {
		return err
	}
	// fill remainder with generated docs
	// dont sort order of id
	for i := count; i > 2; i-- {
		id := fmt.Sprintf("doc-%d", i)
		td := createTD(id)
		//jsonDoc, _ := json.MarshalIndent(td, "  ", "  ")
		jsonDoc, _ := protojson.Marshal(td)
		args := &svc.Write_Args{Key: id, JsonDoc: string(jsonDoc)}
		_, _ = store.Write(nil, args)
	}
	return nil
}

// func readStoredFile() ([]byte, error) {
// 	filename := "/tmp/test-dirfilestore.json"
// 	data, err := ioutil.ReadFile(filename)
// 	return data, err
// }

// Generic directory store testcases
func TestStartStop(t *testing.T) {
	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	store.SetWriteDelay(10 * time.Millisecond)

	err = store.Start()
	time.Sleep(time.Millisecond * 20)
	assert.NoError(t, err)
	err = store.Stop()
	assert.NoError(t, err)

	// store should now exist and restart succeed
	assert.FileExists(t, jsonStoreFile)
	store, err = jsonstore.NewJsonStore(jsonStoreFile)
	err = store.Start()
	time.Sleep(time.Millisecond * 20)
	assert.NoError(t, err)
	err = store.Stop()
	assert.NoError(t, err)

	// corrupted file should not stop the service (or should it?)
	ioutil.WriteFile(jsonStoreFile, []byte("-invalid json"), 0600)
	store, err = jsonstore.NewJsonStore(jsonStoreFile)
	assert.NoError(t, err)
}

func TestCreateStoreBadFolder(t *testing.T) {
	filename := "/folder/does/notexist/dirfilestore.json"
	_, err := jsonstore.NewJsonStore(filename)
	assert.Error(t, err)
}

func TestCreateStoreReadOnlyFolder(t *testing.T) {
	filename := "/var/dirfilestore.json"
	_, err := jsonstore.NewJsonStore(filename)
	assert.Error(t, err)
}

func TestCreateStoreCantReadFile(t *testing.T) {
	filename := "/var"
	_, err := jsonstore.NewJsonStore(filename)
	assert.Error(t, err)
	assert.Error(t, err)
}

func TestWriteRead(t *testing.T) {
	newTitle := "new title"

	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	store.SetWriteDelay(time.Millisecond * 10)
	require.NoError(t, err)
	store.Start()
	err = addDocs(store, 0)
	require.NoError(t, err)

	// Replace
	doc3 := fmt.Sprintf(`{"id":"%s","title":"%s"}`, doc1ID, newTitle)
	_, err = store.Write(nil, &svc.Write_Args{Key: doc1ID, JsonDoc: doc3})
	assert.NoError(t, err)

	// wait for write
	time.Sleep(time.Millisecond * 100)

	// Read
	resp, err := store.Read(&svc.Read_Args{Key: doc1ID})
	assert.NoError(t, err)
	assert.Equal(t, doc3, resp.JsonDoc)

	// Delete
	_, err = store.Remove(nil, &svc.Remove_Args{Key: doc1ID})
	assert.NoError(t, err)
	err = store.Stop()
	assert.NoError(t, err)

	// Read again should fail
	_, err = store.Read(&svc.Read_Args{Key: doc1ID})
	assert.Error(t, err)
}

func TestWriteBadData(t *testing.T) {
	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	// missing doc
	_, err = store.Write(nil, &svc.Write_Args{Key: doc1ID})
	assert.Error(t, err)
	// not json
	_, err = store.Write(nil, &svc.Write_Args{Key: doc1ID, JsonDoc: "notjson"})
	assert.Error(t, err)
	// missing key
	_, err = store.Write(nil, &svc.Write_Args{JsonDoc: "{}"})
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	var docs []interface{}
	var total = 100
	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	err = addDocs(store, total)
	require.NoError(t, err)

	resp, err := store.List(nil, &svc.List_Args{})
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.Equal(t, total, len(docs))
}

func TestListWithLimit(t *testing.T) {
	var docs []interface{}
	const total = 100

	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	err = addDocs(store, total)
	require.NoError(t, err)

	resp, err := store.List(nil, &svc.List_Args{Offset: 10, Limit: 20})
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(docs))

	// based on count of 100
	resp, err = store.List(nil, &svc.List_Args{Offset: total - 10, Limit: 20})
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(docs))

	// based on count of 100
	resp, err = store.List(nil, &svc.List_Args{Offset: 105})
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(docs))
}

func TestQuery(t *testing.T) {
	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	err = addDocs(store, 20)
	require.NoError(t, err)

	var docs []interface{}

	// filter on key 'id' == doc1
	//args := &svc.Query_Args{JsonPathQuery: `$[?(@.id=="doc1")]`}
	args := &svc.Query_Args{JsonPathQuery: `$[?(@.id=="doc1")]`, Limit: 10, Offset: 0}
	resp, err := store.Query(nil, args)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.NotEmpty(t, docs)

	// regular nested filter comparison. note that a TD does not hold values
	args.JsonPathQuery = `$[?(@.properties.title.name=="title1")]`
	resp, err = store.Query(nil, args)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NotEmpty(t, docs)

	// filter with nested notation. some examples that return a list of TDs matching the filter
	//res, err = fileStore.Query(`$[?(@.properties.title.value=="title1")]`, 0, 0)
	// res, err = fileStore.Query(`$[?(@.*.title.value=="title1")]`, 0, 0)
	// res, err = fileStore.Query(`$[?(@['properties']['title']['value']=="title1")]`, 0, 0)
	args.JsonPathQuery = `$[?(@..title.name=="title1")]`
	resp, err = store.Query(nil, args)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)

	// these only return the properties - not good
	// res, err = fileStore.Query(`$.*.properties[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
	//res, err = fileStore.Query(`$.*.*[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
	// res, err = fileStore.Query(`$[?(@...value=="title1")]`, 0, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, docs)

	// filter with bracket notation
	args.JsonPathQuery = `$[?(@["id"]=="doc1")]`
	resp, err = store.Query(nil, args)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.NotEmpty(t, docs)

	// filter with bracket notation and current object literal (for search @type)
	// only supported by ohler55/ojg
	args.JsonPathQuery = `$[?(@['@type']=="sensor")]`
	resp, err = store.Query(nil, args)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.Greater(t, len(docs), 1)

	// bad query expression
	args.JsonPathQuery = `$[?(.id=="doc1")]`
	resp, err = store.Query(nil, args)
	assert.Error(t, err)
}

// tests to figure out how to use jpquery with bracket notation
func TestQueryBracketNotationA(t *testing.T) {
	store := make(map[string]interface{})
	query1 := `$[?(@['type']=="type1")]`
	query2 := `$[?(@['@type']=="sensor")]`

	jsondoc := `{
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

	err := json.Unmarshal([]byte(jsondoc), &store)
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

// tests to figure out how to use jpquery with bracket notation
func TestQueryBracketNotationB(t *testing.T) {
	var docs []interface{}

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

	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)

	//td1json, err := json.MarshalIndent(td1, "", "")
	td1json, err := protojson.Marshal(&td1)
	td2json, err := protojson.Marshal(&td2)
	_, _ = store.Write(nil, &svc.Write_Args{Key: id1, JsonDoc: string(td1json)})
	_, err = store.Write(nil, &svc.Write_Args{Key: id2, JsonDoc: string(td2json)})
	assert.NoError(t, err)

	// query returns 2 sensors. The order is random
	resp, err := store.Query(nil, &svc.Query_Args{JsonPathQuery: queryString})
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	require.Equal(t, 2, len(docs))

	// both results are of device type sensor
	readTD1 := docs[0].(map[string]interface{})
	require.NotNil(t, readTD1)
	read1type := readTD1["@type"]
	assert.Equal(t, string(vocab.DeviceTypeSensor), read1type)
}

// test query with reduced list of IDs
func TestQueryFiltered(t *testing.T) {
	var docs []interface{}
	queryString := "$..id"

	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	addDocs(store, 10)

	// result of a normal query
	resp, err := store.Query(nil, &svc.Query_Args{JsonPathQuery: queryString})
	require.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(docs))

	// authorize access to Thing1 only
	validKeys := []string{doc1ID}
	resp, err = store.Query(nil, &svc.Query_Args{JsonPathQuery: queryString, Keys: validKeys})
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
	assert.Equal(t, 1, len(docs))
}

// crude read/write performance test
func TestReadWritePerf(t *testing.T) {
	const iterations = 10000
	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	store.Start()
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
		_, err = store.Write(nil, &svc.Write_Args{Key: docID, JsonDoc: string(doc3)})
		assert.NoError(t, err)
	}
	d1 := time.Since(t1)
	t2 := time.Now()
	for i = 0; i < iterations; i++ {
		// Read
		docID := fmt.Sprintf("doc-%d", i)
		resp, err := store.Read(&svc.Read_Args{Key: docID})
		assert.NoError(t, err)
		require.NotEmpty(t, resp.JsonDoc)
	}
	d2 := time.Since(t2)
	store.Stop()
	logrus.Infof("TestQuery, %d runs. Write: %d msec; Reads: %d msec",
		i, d1.Milliseconds(), d2.Milliseconds())
}

// crude list performance test
func TestPerfList(t *testing.T) {
	const iterations = 100
	const datasize = 1000
	const listLimit = 1000

	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	store.Start()
	err = addDocs(store, datasize)
	require.NoError(t, err)

	t1 := time.Now()
	var i int
	for i = 0; i < iterations; i++ {
		// List
		listArgs := &svc.List_Args{Limit: listLimit}
		resp2, err := store.List(nil, listArgs)
		assert.NoError(t, err)
		require.NotEmpty(t, resp2.JsonDocs)
	}
	d1 := time.Since(t1)
	store.Stop()
	logrus.Infof("TestList, %d runs: %d msec.", i, d1.Milliseconds())
}

// crude query performance test
func TestPerfQuery(t *testing.T) {
	const iterations = 1000
	const dataSize = 1000

	_ = os.Remove(jsonStoreFile)
	store, err := jsonstore.NewJsonStore(jsonStoreFile)
	require.NoError(t, err)
	store.Start()
	err = addDocs(store, dataSize)
	require.NoError(t, err)

	t1 := time.Now()
	var i int
	for i = 0; i < iterations; i++ {

		// filter on key 'id' == doc1
		var docs []interface{}
		args := &svc.Query_Args{
			JsonPathQuery: `$[?(@.properties.title.name=="title1")]`,
			//JsonPathQuery: `$[?(@.id=="doc1")]`,
			//JsonPathQuery: `$..id`,
			Offset: 0,
			//Keys:          []string{doc1ID},
		}
		resp, err := store.Query(nil, args)
		require.NoError(t, err)
		err = json.Unmarshal([]byte(resp.JsonDocs), &docs)
		assert.NoError(t, err)
		assert.NotEmpty(t, docs)
	}
	d1 := time.Since(t1)
	store.Stop()
	logrus.Infof("TestQuery, %d runs: %d msec.", i, d1.Milliseconds())
}
