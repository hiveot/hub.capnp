package dirfilestore_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ohler55/ojg/jp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/hub/thingdir/pkg/dirstore"
	"github.com/wostzone/hub/thingdir/pkg/dirstore/dirfilestore"
)

const (
	Thing1ID        = "thing1"
	Thing2ID        = "thing2"
	PropTitle1Value = "title1"
)

// AddTDs adds two TDs with two properties each: Title and version
func addTDs(store *dirfilestore.DirFileStore) {
	id1 := Thing1ID
	td1 := thing.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
	prop1 := td1.AddProperty(PropTitle1Value, PropTitle1Value, vocab.WoTDataTypeString)

	prop2 := td1.AddProperty(vocab.PropNameSoftwareVersion, "Thing version", vocab.WoTDataTypeString)

	id2 := Thing2ID
	td2 := thing.CreateTD(id2, "test TD", vocab.DeviceTypeSensor)
	td2.UpdateProperty(PropTitle1Value, prop1)
	td2.UpdateProperty(vocab.PropNameSoftwareVersion, prop2)

	tdd := td1.AsMap()
	_ = store.Replace(id1, tdd)
	tdd = td2.AsMap()
	_ = store.Replace(id2, tdd)

}

func makeFileStore() *dirfilestore.DirFileStore {
	filename := "/tmp/test-dirfilestore.json"
	_ = os.Remove(filename) // remove if exist
	store := dirfilestore.NewDirFileStore(filename)
	return store
}

// func readStoredFile() ([]byte, error) {
// 	filename := "/tmp/test-dirfilestore.json"
// 	data, err := ioutil.ReadFile(filename)
// 	return data, err
// }

// Generic directory store testcases
func TestFileStoreStartStop(t *testing.T) {
	fileStore := makeFileStore()
	dirstore.DirStoreStartStop(t, fileStore)
}

func TestCreateStoreBadFolder(t *testing.T) {
	filename := "/folder/does/notexist/dirfilestore.json"
	store := dirfilestore.NewDirFileStore(filename)
	err := store.Open()
	assert.Error(t, err)
}

func TestCreateStoreReadOnlyFolder(t *testing.T) {
	filename := "/var/dirfilestore.json"
	store := dirfilestore.NewDirFileStore(filename)
	err := store.Open()
	assert.Error(t, err)
}

func TestCreateStoreCantReadFile(t *testing.T) {
	filename := "/var"
	store := dirfilestore.NewDirFileStore(filename)
	err := store.Open()
	assert.Error(t, err)
}

func TestFileStoreWrite(t *testing.T) {
	fileStore := makeFileStore()
	dirstore.DirStoreCrud(t, fileStore)
}

func TestList(t *testing.T) {
	fileStore := makeFileStore()
	_ = fileStore.Open()
	addTDs(fileStore)

	items := fileStore.List(0, 0, nil)
	assert.Greater(t, len(items), 0)
	item1 := items[0]
	require.NotNil(t, item1)

}

func TestQuery(t *testing.T) {
	fileStore := makeFileStore()
	err := fileStore.Open()
	require.NoError(t, err)
	addTDs(fileStore)
	// dirstore.DirStoreCrud(t, fileStore)

	t1 := time.Now()
	var i int
	for i = 0; i < 1; i++ {

		// regular filter
		res, err := fileStore.Query(`$[?(@.id=="thing1")]`, 0, 1, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// regular nested filter comparison. note that a TD does not hold values
		res, err = fileStore.Query(`$[?(@.properties.title1.title=="title1")]`, 0, 0, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with nested notation. some examples that return a list of TDs matching the filter
		//res, err = fileStore.Query(`$[?(@.properties.title.value=="title1")]`, 0, 0)
		// res, err = fileStore.Query(`$[?(@.*.title.value=="title1")]`, 0, 0)
		// res, err = fileStore.Query(`$[?(@['properties']['title']['value']=="title1")]`, 0, 0)
		res, err = fileStore.Query(`$[?(@..title1.title=="title1")]`, 0, 0, nil)

		// these only return the properties - not good
		// res, err = fileStore.Query(`$.*.properties[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
		//res, err = fileStore.Query(`$.*.*[?(@.value=="title1")]`, 0, 0) // returns list of props, not tds
		// res, err = fileStore.Query(`$[?(@...value=="title1")]`, 0, 0)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with bracket notation
		res, err = fileStore.Query(`$[?(@["id"]=="thing1")]`, 0, 0, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with bracket notation and current object literal (for search @type)
		// only supported by ohler55/ojg
		res, err = fileStore.Query(`$[?(@['@type']=="sensor")]`, 0, 1, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// bad query expression
		_, err = fileStore.Query(`$[?(.id=="thing1")]`, 0, 0, nil)
		assert.Error(t, err)
	}
	d1 := time.Since(t1)
	logrus.Infof("TestQuery, %d runs: %d msec", i, d1.Milliseconds())
	// logrus.Infof("Query Results:\n%v", res)
	fileStore.Close()
}

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

func TestQueryBracketNotationB(t *testing.T) {
	//store := make(map[string]interface{})
	queryString := "$[?(@['@type']==\"sensor\")]"
	id1 := "thing1"
	td1 := thing.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
	td1.AddProperty(vocab.PropNameTitle, "Sensor title", vocab.WoTDataTypeString)
	td1.AddProperty(vocab.PropNameValue, "Sensor value", vocab.WoTDataTypeNumber)

	id2 := "thing2"
	td2 := thing.CreateTD(id2, "test TD 2", vocab.DeviceTypeSensor)
	td2.AddProperty("title", "The switch", vocab.WoTDataTypeBool)

	fileStore := makeFileStore()
	_ = fileStore.Open()

	// replace will add if it doesn't exist
	_ = fileStore.Replace(id1, td1.AsMap())
	_ = fileStore.Replace(id2, td2.AsMap())

	// query returns 2 sensors. The order is random
	res, err := fileStore.Query(queryString, 0, 2, nil)
	assert.NoError(t, err)
	require.Equal(t, 2, len(res))

	// both results are of device type sensor
	readTD1 := res[0].(map[string]interface{})
	read1type := readTD1["@type"]
	assert.Equal(t, string(vocab.DeviceTypeSensor), read1type)

	fileStore.Close()

}

func TestQueryValueProp(t *testing.T) {
	queryString := "$[?(@.properties..title=='The switch')]"
	id1 := "thing1"
	td1 := thing.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
	td1.AddProperty(vocab.PropNameTitle, "Device title", vocab.WoTDataTypeNumber)
	td1.AddProperty(vocab.PropNameValue, "Sensor value", vocab.WoTDataTypeString)

	td2Json := `{
		"id": "thing2",
		"type": "type2",
		"@type": "sensor", 
		"properties": { 
		  "title": {  
		   "title": "The switch" 
		   	} 
		  }
		}`
	var td2 map[string]interface{}
	err := json.Unmarshal([]byte(td2Json), &td2)
	assert.NoError(t, err)

	fileStore := makeFileStore()
	_ = fileStore.Open()

	// dirstore.DirStoreCrud(t, fileStore)
	_ = fileStore.Replace(id1, td1.AsMap())
	_ = fileStore.Replace("thing2", td2)

	res, err := fileStore.Query(queryString, 0, 2, nil)
	require.NoError(t, err)
	require.NotEmpty(t, res)
	resJson, _ := json.MarshalIndent(res, " ", " ")
	fmt.Println(string(resJson))
	// logrus.Infof("query result: %s", resJson)

	fileStore.Close()
}

func TestQueryAclFilter(t *testing.T) {
	queryString := "$..id"
	fileStore := makeFileStore()
	_ = fileStore.Open()
	addTDs(fileStore)

	// result of a normal query
	result, err := fileStore.Query(queryString, 0, 0, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// authorize access to Thing1 only
	result, err = fileStore.Query(queryString, 0, 0,
		func(thingID string) bool {
			return thingID == Thing1ID
		})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

// Test of merging two TDs
// TODO: improve tests to verify a correct merge
func TestPatch(t *testing.T) {
	fileStore := makeFileStore()
	_ = fileStore.Open()
	// step 1: create TDs
	addTDs(fileStore)

	// step 2: create an updated TD with a property named title2 with 'description2' as title
	id2 := "thing2"
	td2 := thing.CreateTD(id2, "test TD", vocab.DeviceTypeSensor)
	td2.AddProperty("title2", "description2", vocab.WoTDataTypeString)

	// step 3: patch should merge existing properties and add the 'title2' property
	_ = fileStore.Patch(id2, td2.AsMap())

	// step 4: get the resulting TD
	td2b, err := fileStore.Get(id2)
	assert.NoError(t, err)
	thing2 := td2b.(map[string]interface{})

	// step 5: get the combined properties
	props, found := thing2["properties"].(map[string]interface{})
	t1, found := props[PropTitle1Value]
	assert.True(t, found, "Expected both title 1 and title 2 to still exist")
	assert.NotNil(t, t1)

	t2, found := props["title2"]
	assert.True(t, found, "Expected both title 1 and title 2 to still exist")
	assert.NotNil(t, t2)

	//val, found = thing.GetPropertyValue(thing2, "title2")
	//assert.True(t, found, "Expected propery title2 to exist")
	//assert.Equal(t, "value2", val)
	fileStore.Close()
}

func TestBadPatch(t *testing.T) {
	fileStore := makeFileStore()
	_ = fileStore.Open()

	id2 := "thing2"
	err := fileStore.Patch(id2, nil)
	assert.Error(t, err)
}

func TestBadReplace(t *testing.T) {
	fileStore := makeFileStore()
	_ = fileStore.Open()

	id2 := "thing2"
	err := fileStore.Replace(id2, nil)
	assert.Error(t, err)
}
