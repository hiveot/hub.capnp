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
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/hub/thingdir/pkg/dirstore"
	"github.com/wostzone/hub/thingdir/pkg/dirstore/dirfilestore"
)

const (
	Thing1ID        = "thing1"
	Thing2ID        = "thing2"
	PropNameTitle   = "title"
	PropTitle1Value = "title1"
	PropNameVersion = "version"
)

// AddTDs adds two TDs with two properties each: Title and version
func addTDs(store *dirfilestore.DirFileStore) {
	id1 := Thing1ID
	td1 := td.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)

	prop1 := td.CreateProperty(PropNameTitle, "Property title", vocab.PropertyTypeAttr)
	td.SetPropertyValue(prop1, PropTitle1Value)
	td.AddTDProperty(td1, PropNameTitle, prop1)

	prop2 := td.CreateProperty(PropNameVersion, "Thing version", vocab.PropertyTypeAttr)
	td.SetPropertyValue(prop2, "version1")
	td.AddTDProperty(td1, PropNameVersion, prop2)

	id2 := Thing2ID
	td2 := td.CreateTD(id2, "test TD", vocab.DeviceTypeSensor)
	td.AddTDProperty(td2, PropNameTitle, prop1)
	td.AddTDProperty(td2, PropNameVersion, prop2)

	tdd := map[string]interface{}(td1)
	_ = store.Replace(id1, tdd)
	tdd = map[string]interface{}(td2)
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

		// regular nested filter comparison
		res, err = fileStore.Query(`$[?(@.properties.title.value=="title1")]`, 0, 0, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		// filter with nested notation. some examples that return a list of TDs matching the filter
		//res, err = fileStore.Query(`$[?(@.properties.title.value=="title1")]`, 0, 0)
		// res, err = fileStore.Query(`$[?(@.*.title.value=="title1")]`, 0, 0)
		// res, err = fileStore.Query(`$[?(@['properties']['title']['value']=="title1")]`, 0, 0)
		res, err = fileStore.Query(`$[?(@..title.value=="title1")]`, 0, 0, nil)

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
	query1 := `$[?(@['type']=="type1")]`
	query2 := `$[?(@['@type']=="sensor")]`

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
	queryString := "$[?(@['@type']==\"sensor\")]"
	id1 := "thing1"
	td1 := td.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
	titleProp := td.CreateProperty("Title", "Sensor title", vocab.PropertyTypeAttr)
	td.AddTDProperty(td1, "title", titleProp)
	valueProp := td.CreateProperty("value", "Sensor value", vocab.PropertyTypeSensor)
	td.AddTDProperty(td1, "title", valueProp)

	id2 := "thing2"
	td2 := make(map[string]interface{})
	td2["id"] = "thing2"
	td2["type"] = "type2"
	td2[vocab.WoTAtType] = "sensor"
	td2["actions"] = make(map[string]interface{})
	td2["properties"] = make(map[string]interface{})
	td.AddTDProperty(td2, "title", "The switch")

	fileStore := makeFileStore()
	_ = fileStore.Open()

	// replace will add if it doesn't exist
	_ = fileStore.Replace(id1, td1)
	_ = fileStore.Replace(id2, td2)

	// query returns 2 sensors. not sure about the sort order
	res, err := fileStore.Query(queryString, 0, 2, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	// item1 := res[0].(map[string]interface{})
	// item1ID := item1["id"]
	// assert.(t, id1, item1ID)
	// assert.Equal(t, res[0], td1)

	fileStore.Close()

}

func TestQueryValueProp(t *testing.T) {
	queryString := "$[?(@.properties..title=='The switch')]"
	id1 := "thing1"
	td1 := td.CreateTD(id1, "test TD", vocab.DeviceTypeSensor)
	titleProp := td.CreateProperty("Title", "Device title", vocab.PropertyTypeAttr)
	td.AddTDProperty(td1, "title", titleProp)
	valueProp := td.CreateProperty("value", "Sensor value", vocab.PropertyTypeSensor)
	td.AddTDProperty(td1, string(vocab.PropertyTypeSensor), valueProp)

	id2 := "thing2"
	td2 := make(map[string]interface{})
	td2["id"] = "thing2"
	td2["type"] = "type2"
	td2[vocab.WoTAtType] = "sensor"
	td2["actions"] = make(map[string]interface{})
	td2["properties"] = make(map[string]interface{})
	td.AddTDProperty(td2, "title", "The switch")

	fileStore := makeFileStore()
	_ = fileStore.Open()

	// dirstore.DirStoreCrud(t, fileStore)
	_ = fileStore.Replace(id1, td1)
	_ = fileStore.Replace(id2, td2)

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
	addTDs(fileStore)

	id2 := "thing2"
	td2 := td.CreateTD(id2, "test TD", vocab.DeviceTypeSensor)
	prop2 := td.CreateProperty("title2", "description2", vocab.PropertyTypeAttr)
	td.SetPropertyValue(prop2, "value2")
	td.AddTDProperty(td2, "title2", prop2)
	_ = fileStore.Patch(id2, td2)

	td2b, err := fileStore.Get(id2)
	assert.NoError(t, err)
	thing2 := td2b.(map[string]interface{})

	val, found := td.GetPropertyValue(thing2, PropNameTitle)
	assert.True(t, found, "Expected propery title1 to still exist")
	assert.Equal(t, PropTitle1Value, val)

	val, found = td.GetPropertyValue(thing2, "title2")
	assert.True(t, found, "Expected propery title2 to exist")
	// thing2b := td2b.(td.ThingTD)
	// val := thing2b["title2"]
	assert.Equal(t, "value2", val)
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
