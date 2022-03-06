package dirserver_test

import (
	"fmt"
	"github.com/wostzone/hub/authn/pkg/jwtissuer"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/td"
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
var authenticateResult = true

// authorization result
var authorizeResult = true

// Add a bunch of TDs
func AddTds(client *dirclient.DirClient) {
	for _, tdDef := range tdDefs {
		td1 := td.CreateTD(tdDef.id, "test thing", tdDef.deviceType)
		td.AddTDProperty(td1, "name", td.CreateProperty(tdDef.name, "", vocab.PropertyTypeAttr))
		client.UpdateTD(tdDef.id, td1)
	}
}

// Authenticator for testing of authentication of type 'authenticate.VerifyUsernamePassword'
func authenticator(username string, password string) bool {
	return authenticateResult
}

// Authorizer for testing of authorization of type 'authorize.ValidateAuthorization'
func authorizer(userID string, certOU string,
	thingID string, writing bool, writeType string) bool {
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
		//authenticator,
		authorizer)
	directoryServer.Start()

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
	td1 := td.CreateTD(thingID1, "test thing", deviceType1)
	err = dirClient.UpdateTD(thingID1, td1)
	assert.NoError(t, err)

	// get result
	td2, err := dirClient.GetTD(thingID1)
	assert.NoError(t, err)
	assert.Equal(t, td1["id"], td2["id"])

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

	err = dirClient.UpdateTD(thingID1, nil)
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

	AddTds(dirClient)

	// Change the device type to sensor using patch
	thingID1 := tdDefs[0].id
	td1 := td.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	td.AddTDProperty(td1, "name", td.CreateProperty("name1", "just a name", vocab.PropertyTypeAttr))

	err = dirClient.PatchTD(thingID1, td1)
	assert.NoError(t, err)
	props1 := td1["properties"].(map[string]interface{})
	nameProp1 := props1["name"].(map[string]interface{})
	nameProp1val := nameProp1["title"]

	// check result
	td2, err := dirClient.GetTD(thingID1)
	assert.NoError(t, err)
	assert.Equal(t, td1["id"], td2["id"])
	assert.Equal(t, string(vocab.DeviceTypeSensor), td2["@type"])
	props2 := td2["properties"].(map[string]interface{})
	nameProp2 := props2["name"].(map[string]interface{})
	nameProp2val := nameProp2["title"]
	assert.NotEmpty(t, nameProp2val)
	assert.Equal(t, nameProp1val, nameProp2val)
	dirClient.Close()
}

func TestBadPatch(t *testing.T) {

	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)

	// Client start only succeeds if server is running
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	AddTds(dirClient)
	thingID1 := tdDefs[0].id
	td1 := td.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	td.AddTDProperty(td1, "name", td.CreateProperty("name1", "just a name", vocab.PropertyTypeAttr))

	err = dirClient.PatchTD(thingID1, nil)
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
	AddTds(dirClient)

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
	AddTds(dirClient)

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
	password := "pass1"
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)

	authenticateResult = false

	err := dirClient.ConnectWithLoginID(loginID, password)
	assert.Error(t, err)

	_, err = dirClient.ListTDs(0, 0)
	assert.Error(t, err)

	dirClient.Close()
}

func TestNotAuthorized(t *testing.T) {
	thingID1 := tdDefs[0].id
	loginID := "user1"
	dirClient := dirclient.NewDirClient(serverHostPort, testCerts.CaCert)
	td1 := td.CreateTD(thingID1, "test thing", vocab.DeviceTypeSensor)
	td.AddTDProperty(td1, "name", td.CreateProperty("name1", "just a name", vocab.PropertyTypeAttr))

	authenticateResult = true
	authorizeResult = false
	// the Directory Server uses the server cert public key to verify the token
	issuer := jwtissuer.NewJWTIssuer("test", testCerts.ServerKey, nil)
	accessToken, _, _ := issuer.CreateJWTTokens(loginID)

	// authenticationResult is true so login should succeed
	// FIXME: fix authentication so this will work
	// - authenticate via dirClient or must an authserver be used?
	//err := dirClient.ConnectWithLoginID(loginID, password)
	err := dirClient.ConnectWithJwtToken(loginID, accessToken)
	assert.NoError(t, err)

	// expect empty list as the user is authenticated but not authorized
	tds, err := dirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	assert.Empty(t, tds)

	_, err = dirClient.GetTD(thingID1)
	assert.Error(t, err)

	err = dirClient.Delete(thingID1)
	assert.Error(t, err)

	err = dirClient.UpdateTD(thingID1, td1)
	assert.Error(t, err)

	err = dirClient.PatchTD(thingID1, td1)
	assert.Error(t, err)

	dirClient.Close()
}
