package dirclient_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/td"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/vocab"
	"github.com/wostzone/hub/lib/serve/pkg/tlsserver"
	"github.com/wostzone/hub/thingdir/pkg/dirclient"
)

var testDirectoryAddr string

const testDirectoryPort = 9990

var testCerts testenv.TestCerts

// var caCertPath string
// var pluginCertPath string
// var pluginKeyPath string

/*
 */
func startTestServer() *tlsserver.TLSServer {
	server := tlsserver.NewTLSServer(testDirectoryAddr, testDirectoryPort,
		testCerts.ServerCert, testCerts.CaCert)

	// Todo test with auth

	server.Start()
	return server
}

// test setup. Run tests with -p 1 as this test environment doesn't handle concurrent tests
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of DirectoryClient ------")
	testDirectoryAddr = "127.0.0.1"
	testCerts = testenv.CreateCertBundle()

	res := m.Run()
	os.Exit(res)
}

func TestConnectClose(t *testing.T) {
	// launch a server to receive requests
	server := startTestServer()
	server.AddHandler(dirclient.RouteThings, func(userID string, resp http.ResponseWriter, req *http.Request) {
		tds := []map[string]interface{}{}
		data, _ := json.Marshal(tds)
		resp.Write([]byte(data))
		//return
	})
	//
	hostPort := fmt.Sprintf("%s:%d", testDirectoryAddr, testDirectoryPort)
	dirClient := dirclient.NewDirClient(hostPort, testCerts.CaCert)

	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)
	_, err = dirClient.ListTDs(0, 0)
	assert.NoError(t, err)
	dirClient.Close()

	// server isn't setup with username password login so login endpoint is not found
	dirClient = dirclient.NewDirClient(hostPort, testCerts.CaCert)
	err = dirClient.ConnectWithLoginID("user1", "pass1")
	assert.Error(t, err)
	// without server auth this should not succeed
	_, err = dirClient.ListTDs(0, 0)
	assert.Error(t, err)

	dirClient.Close()
	server.Stop()
}

func TestUpdateTD(t *testing.T) {
	var receivedTD td.ThingTD
	var body []byte
	var receivedPatch bool
	var err2 error
	const id1 = "thing1"

	server := startTestServer()
	server.AddHandler(dirclient.RouteThingID, func(userID string, response http.ResponseWriter, request *http.Request) {
		logrus.Infof("TestUpdateTD: %s %s", request.Method, request.RequestURI)

		if request.Method == "POST" {
			body, err2 = ioutil.ReadAll(request.Body)
			if err2 == nil {
				err2 = json.Unmarshal(body, &receivedTD)
			}
		} else if request.Method == "PATCH" {
			body, err2 = ioutil.ReadAll(request.Body)
			if err2 == nil {
				err2 = json.Unmarshal(body, &receivedTD)
			}
			receivedPatch = true
		} else if request.Method == "GET" {
			parts := strings.Split(request.URL.Path, "/")
			id := parts[len(parts)-1]
			assert.Equal(t, id1, id)
			//return the previously sent td
			msg, _ := json.Marshal(receivedTD)
			response.Write(msg)
		}
		assert.NoError(t, err2)
	})
	hostPort := fmt.Sprintf("%s:%d", testDirectoryAddr, testDirectoryPort)
	dirClient := dirclient.NewDirClient(hostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	// write a TD document
	td := td.CreateTD(id1, "test sensor TD", vocab.DeviceTypeSensor)
	err = dirClient.UpdateTD(id1, td)
	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.Equal(t, id1, receivedTD["id"])

	// check result
	receivedTD2, err := dirClient.GetTD(id1)
	assert.NoError(t, err)
	assert.Equal(t, id1, receivedTD2["id"])

	// patch the a TD document
	err = dirClient.PatchTD(id1, td)
	assert.NoError(t, err)
	assert.True(t, receivedPatch)

	dirClient.Close()
	server.Stop()

}

func TestQueryAndList(t *testing.T) {
	const query = "$.hello.world"
	server := startTestServer()
	// if no thingID is specified then this is a request for a list or a query
	server.AddHandler(dirclient.RouteThings, func(userID string, response http.ResponseWriter, request *http.Request) {
		logrus.Infof("TestQuery: %s %s", request.Method, request.RequestURI)

		if request.Method == "GET" {
			q := request.URL.Query().Get(dirclient.ParamQuery)
			thd := td.CreateTD("thing1", "Test TD", vocab.DeviceTypeSensor)
			prop := td.CreateProperty("query", "", vocab.PropertyTypeAttr)
			td.SetPropertyDataTypeString(prop, 0, 0)
			td.SetPropertyValue(prop, q)
			td.AddTDProperty(thd, dirclient.ParamQuery, prop)
			tdList := []td.ThingTD{thd}
			data, _ := json.Marshal(tdList)
			response.Write(data)
		} else {
			server.WriteBadRequest(response, "Only GET is supported")
		}
	})

	hostPort := fmt.Sprintf("%s:%d", testDirectoryAddr, testDirectoryPort)
	dirClient := dirclient.NewDirClient(hostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	// test query
	td2, err := dirClient.QueryTDs(query, 0, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, td2)
	val, _ := td.GetPropertyValue(td2[0], dirclient.ParamQuery)
	assert.Equal(t, query, val)

	// test list
	td3, err := dirClient.ListTDs(0, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, td3)

	_, err = dirClient.GetTD("notexistingtd")
	require.Error(t, err)

	dirClient.Close()
	server.Stop()
}

func TestDelete(t *testing.T) {
	const thingID1 = "thing1"
	var idToDelete string

	server := startTestServer()
	server.AddHandler(dirclient.RouteThingID, func(userID string, response http.ResponseWriter, request *http.Request) {
		if request.Method == "DELETE" {
			parts := strings.Split(request.URL.Path, "/")
			idToDelete = parts[len(parts)-1]
		} else {
			server.WriteBadRequest(response, "wrong method: "+request.Method)
		}
	})

	hostPort := fmt.Sprintf("%s:%d", testDirectoryAddr, testDirectoryPort)
	dirClient := dirclient.NewDirClient(hostPort, testCerts.CaCert)
	err := dirClient.ConnectWithClientCert(testCerts.PluginCert)
	require.NoError(t, err)

	err = dirClient.Delete(thingID1)
	require.NoError(t, err)
	assert.Equal(t, thingID1, idToDelete)

	dirClient.Close()
	server.Stop()
}
