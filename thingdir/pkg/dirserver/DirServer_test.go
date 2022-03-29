package dirserver_test

import (
	"encoding/json"
	"fmt"
	"github.com/wostzone/hub/authn/pkg/jwtissuer"
	"github.com/wostzone/hub/lib/client/pkg/thing"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
	"github.com/wostzone/hub/thingdir/pkg/dirserver"
)

const testDirectoryPort = 9990
const testDirectoryServiceInstanceID = "directory"
const testServiceDiscoveryName = "thingdir"
const serverAddress = "127.0.0.1"

var serverHostPort = fmt.Sprintf("%s:%d", serverAddress, testDirectoryPort)

var testCerts testenv.TestCerts

var homeFolder string

// var caCertPath string
var directoryServer *dirserver.DirectoryServer

// var pluginCertPath string
// var pluginKeyPath string
var storeFolder string

// TD's for testing. Expect 2 sensors in this list
var tdDefs = []struct {
	id         string
	deviceType vocab.DeviceType
	name       string
}{
	{"thing1", vocab.DeviceTypeBeacon, "a beacon"},
	{"thing2", vocab.DeviceTypeSensor, "hallway sensor"},
	{"thing3", vocab.DeviceTypeSensor, "garage sensor"},
	{"thing4", vocab.DeviceTypeNetSwitch, "main switch"},
}

// authentication result
//var authenticateResult = true

// authorization result
var authorizeResult = true

// Add a bunch of TDs
func AddTds(srv *dirserver.DirectoryServer) {
	for _, tdoc := range tdDefs {
		var tdMap map[string]interface{}
		td1 := thing.CreateTD(tdoc.id, "test thing", tdoc.deviceType)
		tdJson, _ := json.Marshal(td1)
		_ = json.Unmarshal(tdJson, &tdMap)
		_ = srv.UpdateTD(tdoc.id, tdMap)
	}
}

// Authenticator for testing of authentication of type 'authenticate.VerifyUsernamePassword'
//func authenticator(username string, password string) bool {
//	return authenticateResult
//}

// Authorizer for testing of authorization of type 'authorize.ValidateAuthorization'
func authorizer(userID string, certOU string, thingID string, authType string) bool {
	_ = userID
	_ = certOU
	_ = thingID
	_ = authType
	return authorizeResult
}

// TestMain runs a directory server for use by the test cases in this package
// This uses the directory client in testing
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryServer ------")

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	storeFolder = path.Join(homeFolder, "config")

	testCerts = testenv.CreateCertBundle()

	directoryServer = dirserver.NewDirectoryServer(
		testDirectoryServiceInstanceID,
		storeFolder,
		serverAddress, testDirectoryPort,
		testServiceDiscoveryName,
		testCerts.ServerCert, testCerts.CaCert,
		authorizer)
	_ = directoryServer.Start()

	res := m.Run()

	directoryServer.Stop()
	os.Exit(res)
}

func TestStartStop(t *testing.T) {

	// test start/stop separate from TestMain
	mydirserver := dirserver.NewDirectoryServer(
		testDirectoryServiceInstanceID,
		storeFolder,
		serverAddress, testDirectoryPort+1,
		testServiceDiscoveryName,
		testCerts.ServerCert, testCerts.CaCert,
		//authenticator,
		authorizer)
	err := mydirserver.Start()
	assert.NoError(t, err)

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)

	a := directoryServer.Address()
	assert.Equal(t, serverAddress, a)

	// Client start only succeeds if server is running
	err = dirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	dirClient.Close()
	mydirserver.Stop()

}

func TestUpdate(t *testing.T) {
	thingID1 := "thing1"
	deviceType1 := vocab.DeviceTypeSensor

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	// Create
	td1 := thing.CreateTD(thingID1, "test thing", deviceType1)
	tdMap := td1.AsMap()
	err = directoryServer.UpdateTD(thingID1, tdMap)
	assert.NoError(t, err)

	// get result
	td2, err := dirClient.GetTD(thingID1)
	assert.NoError(t, err)
	assert.Equal(t, td1.ID, td2.ID)

	dirClient.Close()
}

func TestBadUpdate(t *testing.T) {
	thingID1 := "thing1"

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	// Create
	// td1 := td.CreateTD(thingID1, deviceType1)

	err = directoryServer.UpdateTD(thingID1, nil)
	assert.Error(t, err)
	dirClient.Close()

	// use TLS client directly to circumvent type checking
	badUpdate := "hello world"
	tlsClient := tlsclient.NewTLSClient(serverHostPort, testCerts.CaCert)
	err = tlsClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)
	patchPath := strings.Replace(dirclient.RouteThingID, "{thingID}", thingID1, 1)
	_, err = tlsClient.Put(patchPath, badUpdate)
	assert.Error(t, err)

	// test incorrect command
	url := fmt.Sprintf("https://%s%s", serverHostPort, patchPath)
	_, err = tlsClient.Invoke("BADMETHOD", url, badUpdate)
	assert.Error(t, err)

	tlsClient.Close()

}

func TestPatch(t *testing.T) {

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	AddTds(directoryServer)

	// Change the device type to sensor using patch
	thingID1 := tdDefs[0].id
	td1 := thing.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	td1.UpdateProperty("name1", &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{
			Title:    "Just a name",
			ReadOnly: true, // this is an attribute
		},
	})
	err = directoryServer.PatchTD(thingID1, td1.AsMap())
	assert.NoError(t, err)

	//props1 := td1["properties"].(map[string]interface{})
	//nameProp1 := props1["name"].(map[string]interface{})
	//nameProp1val := nameProp1["title"]

	// check result
	td2, err := dirClient.GetTD(thingID1)
	assert.NoError(t, err)
	assert.Equal(t, td1.ID, td2.ID)
	assert.Equal(t, string(vocab.DeviceTypeSensor), td2.AtType)

	props1 := td1.GetProperty("name1")
	props2 := td2.GetProperty("name1")
	assert.Equal(t, props1.Title, props2.Title)

	// cleanup
	dirClient.Close()
}

func TestBadPatch(t *testing.T) {

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	AddTds(directoryServer)
	thingID1 := tdDefs[0].id
	td1 := thing.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	td1.UpdateProperty("name", &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{Title: "title1", ReadOnly: true},
	})

	err = directoryServer.PatchTD(thingID1, nil)
	assert.Error(t, err)
	dirClient.Close()

	// use TLS client directly to circumvent type checking
	badPatch := "hello world"
	tlsClient := tlsclient.NewTLSClient(serverHostPort, testCerts.CaCert)
	err = tlsClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)
	patchPath := strings.Replace(dirclient.RouteThingID, "{thingID}", thingID1, 1)
	_, err = tlsClient.Patch(patchPath, badPatch)
	assert.Error(t, err)

	tlsClient.Close()
}

func TestQueryAndList(t *testing.T) {
	const query = `$[?(@['@type']=='sensor')]`

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	AddTds(directoryServer)

	// Client start only succeeds if server is running
	// err = dirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	// expect 2 sensors
	td2, err := dirClient.QueryTDs(query, 0, 99999)
	require.NoError(t, err)
	assert.Equal(t, 2, len(td2))

	// test list
	td3, err := dirClient.ListTDs(0, 99999)
	require.NoError(t, err)
	assert.Equal(t, len(tdDefs), len(td3))

	// test offset
	td4, err := dirClient.ListTDs(1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, len(td4))

	td5, err := dirClient.ListTDs(len(tdDefs), 1)
	require.NoError(t, err)
	assert.Equal(t, 0, len(td5))

	dirClient.Close()
}

func TestDelete(t *testing.T) {
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	AddTds(directoryServer)

	// Client start only succeeds if server is running
	// err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	// expect 4 items
	tds, err := dirClient.ListTDs(0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(tdDefs), len(tds))

	// remove 1 sensor
	err = dirClient.Delete(tdDefs[1].id)
	assert.NoError(t, err)
	tds, err = dirClient.ListTDs(0, 0)
	require.NoError(t, err)
	assert.Equal(t, len(tdDefs)-1, len(tds))

	// deleting a non existing ID is not an error
	err = dirClient.Delete("notavalidID")
	require.NoError(t, err)

	dirClient.Close()
}

func TestBadRequest(t *testing.T) {
	const query = `$[?(badquery@['@type']=='sensor')]`

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	_, err = dirClient.QueryTDs(query, 0, 99999)
	assert.Error(t, err)

	// test list
	_, err = dirClient.ListTDs(-1, 0)
	require.Error(t, err)

	_, err = dirClient.GetTD("notavalidID")
	require.Error(t, err)
	dirClient.Close()
}

func TestNotAuthenticated(t *testing.T) {
	loginID := "user1"
	accessToken := "badtoken123"
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)

	// this doesn't return an error until the client is used
	dirClient.ConnectWithJwtToken(loginID, accessToken)
	//assert.Error(t, err)

	_, err := dirClient.ListTDs(0, 0)
	assert.Error(t, err)

	dirClient.Close()
}

func TestNotAuthorized(t *testing.T) {
	thingID1 := tdDefs[0].id
	loginID := "user1"
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	td1 := thing.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	td1.UpdateProperty("name", &thing.PropertyAffordance{
		DataSchema: thing.DataSchema{Title: "name1", ReadOnly: true},
	})

	authorizeResult = false
	// the Directory Server uses the server cert public key to verify the token
	issuer := jwtissuer.NewJWTIssuer("test",
		testCerts.ServerKey, 10, 10, nil)
	accessToken, _, _ := issuer.CreateJWTTokens(loginID)

	dirClient.ConnectWithJwtToken(loginID, accessToken)

	// expect empty list as the user is authenticated but not authorized
	tds, err := dirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	assert.Empty(t, tds)

	_, err = dirClient.GetTD(thingID1)
	assert.Error(t, err)

	err = dirClient.Delete(thingID1)
	assert.Error(t, err)

	//err = dirClient.UpdateTD(thingID1, td1)
	//assert.Error(t, err)
	//
	//err = dirClient.PatchTD(thingID1, td1)
	//assert.Error(t, err)

	dirClient.Close()
}
