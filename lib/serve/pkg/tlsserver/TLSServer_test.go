package tlsserver_test

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
	"github.com/wostzone/hub/lib/serve/pkg/tlsserver"
)

var serverAddress string
var serverPort uint = 4444
var clientHostPort string
var testCerts testenv.TestCerts

// These are set in TestMain
var homeFolder string
var serverCertFolder string

// TestMain runs a http server
// Used for all test cases in this package
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of TLSServer_test.go ------")
	// serverAddress = hubnet.GetOutboundIP("").String()
	// use the localhost interface for testing
	serverAddress = "127.0.0.1"
	// hostnames := []string{serverAddress}
	clientHostPort = fmt.Sprintf("%s:%d", serverAddress, serverPort)

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	serverCertFolder = path.Join(homeFolder, "certs")

	testCerts = testenv.CreateCertBundle()
	res := m.Run()

	time.Sleep(time.Second)
	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)
	err := srv.Start()
	assert.NoError(t, err)
	srv.Stop()
}

func TestNoCA(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, nil)
	err := srv.Start()
	assert.Error(t, err)
	srv.Stop()
}
func TestNoServerCert(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		nil, testCerts.CaCert)
	err := srv.Start()
	assert.Error(t, err)
	srv.Stop()
}

// Connect without authentication
func TestNoAuth(t *testing.T) {
	path1 := "/hello"
	path1Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)

	srv.AddHandlerNoAuth(path1, func(http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthCert: path1 hit")
		path1Hit++
	})
	err := srv.Start()
	assert.NoError(t, err)

	cl := tlsclient.NewTLSClient(clientHostPort, nil)
	require.NoError(t, err)
	cl.ConnectNoAuth()
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 1, path1Hit)

	cl.Close()
	srv.Stop()
}

// Test with invalid login authentication
func TestUnauthorized(t *testing.T) {
	path1 := "/test1"
	loginID1 := "user1"
	password1 := "user1pass"

	// setup server and client environment
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)

	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestNoAuth: path1 hit")
		assert.Fail(t, "did not expect the request to pass")
	})
	//
	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	assert.NoError(t, err)

	// AuthMethodNone creates a client without any authentication method
	_, err = cl.ConnectWithLoginID(loginID1, password1, "", tlsclient.AuthMethodNone)
	assert.NoError(t, err)

	// ... which causes any request to fail
	_, err = cl.Get(path1)
	assert.Error(t, err)

	cl.Close()
	srv.Stop()
}

func TestCertAuth(t *testing.T) {
	path1 := "/hello"
	path1Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)
	err := srv.Start()
	assert.NoError(t, err)
	// handler can be added any time
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthCert: path1 hit")
		path1Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	require.NoError(t, err)
	err = cl.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 1, path1Hit)

	cl.Close()
	srv.Stop()
}

// Test valid authentication using JWT
func TestJWTLogin(t *testing.T) {
	user1 := "user1"
	user1Pass := "pass1"
	loginHit := 0
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)
	srv.EnableJwtAuth(&testCerts.ServerKey.PublicKey)
	srv.EnableJwtIssuer(testCerts.ServerKey, func(loginID1, password string) bool {
		loginHit++
		return loginID1 == user1 && password == user1Pass
	})
	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	require.NoError(t, err)

	// first show that an incorrect password fails
	_, err = cl.ConnectWithLoginID(user1, "wrongpassword")
	assert.Error(t, err)
	assert.Equal(t, 1, loginHit)
	// this request should be unauthorized
	_, err = cl.Get(path2)
	assert.Error(t, err)
	assert.Equal(t, 0, path2Hit) // should not increase
	cl.Close()

	// try again with the correct password
	_, err = cl.ConnectWithLoginID(user1, user1Pass)
	assert.NoError(t, err)
	assert.Equal(t, 2, loginHit)

	// use access token
	_, err = cl.Get(path2)
	require.NoError(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Stop()
}

func TestJWTRefresh(t *testing.T) {
	user1 := "user1"
	user1Pass := "pass1"
	loginHit := 0
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)

	srv.EnableJwtAuth(&testCerts.ServerKey.PublicKey)
	srv.EnableJwtIssuer(testCerts.ServerKey, func(loginID1, password string) bool {
		loginHit++
		return loginID1 == user1 && password == user1Pass
	})

	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	require.NoError(t, err)

	_, err = cl.ConnectWithLoginID(user1, user1Pass)
	assert.NoError(t, err)
	assert.Equal(t, 1, loginHit)

	rt, err := cl.RefreshJWTTokens("")
	assert.NoError(t, err)
	assert.NotNil(t, rt)

	// use access token
	_, err = cl.Get(path2)
	require.NoError(t, err)
	assert.Equal(t, 1, path2Hit)
	srv.Stop()

}

func TestQueryParams(t *testing.T) {
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)
	err := srv.Start()
	assert.NoError(t, err)
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		// query string
		q1 := srv.GetQueryString(req, "query1", "")
		assert.Equal(t, "bob", q1)
		// fail not a number
		_, err := srv.GetQueryInt(req, "query1", 0) // not a number
		assert.Error(t, err)
		// query of number
		q2, _ := srv.GetQueryInt(req, "query2", 0)
		assert.Equal(t, 3, q2)
		// default should work
		q3 := srv.GetQueryString(req, "query3", "default")
		assert.Equal(t, "default", q3)
		// multiple parameters fail
		_, err = srv.GetQueryInt(req, "multi", 0)
		assert.Error(t, err)
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	require.NoError(t, err)
	err = cl.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	_, err = cl.Get(fmt.Sprintf("%s?query1=bob&query2=3&multi=a&multi=b", path2))
	assert.NoError(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Stop()
}

func TestWriteResponse(t *testing.T) {
	path2 := "/hello"
	path2Hit := 0
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)
	err := srv.Start()
	assert.NoError(t, err)
	srv.AddHandler(path2, func(userID string, resp http.ResponseWriter, req *http.Request) {
		srv.WriteBadRequest(resp, "bad request")
		srv.WriteInternalError(resp, "internal error")
		srv.WriteNotFound(resp, "not found")
		srv.WriteNotImplemented(resp, "not implemented")
		srv.WriteUnauthorized(resp, "unauthorized")
		path2Hit++
	})

	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	require.NoError(t, err)
	err = cl.ConnectWithClientCert(testCerts.PluginCert)
	assert.NoError(t, err)

	_, err = cl.Get(path2)
	assert.Error(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Stop()
}

func TestBadPort(t *testing.T) {
	srv := tlsserver.NewTLSServer(serverAddress, 1, // bad port
		testCerts.ServerCert, testCerts.CaCert)

	err := srv.Start()
	assert.Error(t, err)
}

// Test BASIC authentication
func TestBasicAuth(t *testing.T) {
	path1 := "/test1"
	path1Hit := 0
	loginID1 := "user1"
	password1 := "user1pass"

	// setup server and client environment
	srv := tlsserver.NewTLSServer(serverAddress, serverPort,
		testCerts.ServerCert, testCerts.CaCert)
	srv.EnableBasicAuth(func(userID, password string) bool {
		path1Hit++
		return userID == loginID1 && password == password1
	})
	err := srv.Start()
	assert.NoError(t, err)
	//
	srv.AddHandler(path1, func(string, http.ResponseWriter, *http.Request) {
		logrus.Infof("TestBasicAuth: path1 hit")
		path1Hit++
	})
	//
	cl := tlsclient.NewTLSClient(clientHostPort, testCerts.CaCert)
	assert.NoError(t, err)
	_, err = cl.ConnectWithLoginID(loginID1, password1, "", tlsclient.AuthMethodBasic)
	assert.NoError(t, err)

	// test the auth with a GET request
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 2, path1Hit)

	// test a failed login
	cl.Close()
	_, err = cl.ConnectWithLoginID(loginID1, "wrongpassword", "", tlsclient.AuthMethodBasic)
	assert.NoError(t, err)
	_, err = cl.Get(path1)
	assert.Error(t, err)
	assert.Equal(t, 3, path1Hit) // should not increase

	cl.Close()
	srv.Stop()
}