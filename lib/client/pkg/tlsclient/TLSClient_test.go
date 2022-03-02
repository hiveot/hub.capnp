package tlsclient_test

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
	"github.com/wostzone/hub/lib/client/pkg/tlsclient"
)

// test hostname and port
var testAddress string

// These are set in TestMain
var homeFolder string

// CA, server and plugin test certificate
var certs testenv.TestCerts
var serverTLSConf *tls.Config

func startTestServer(mux *http.ServeMux) (*http.Server, error) {
	var err error
	httpServer := &http.Server{
		Addr: testAddress,
		// ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when testing
		// WriteTimeout: 10 * time.Second,
		// Handler:   srv.router,
		TLSConfig: serverTLSConf,
		Handler:   mux,
	}
	go func() {
		err = httpServer.ListenAndServeTLS("", "")
		logrus.Errorf("startTestServer: %s", err)
	}()
	// Catch any startup errors
	time.Sleep(100 * time.Millisecond)
	return httpServer, err
}

// TestMain runs a http server
// Used for all test cases in this package
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of httpauthhandler ------")
	testAddress = "127.0.0.1:9888"
	// hostnames := []string{testAddress}

	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	certs = testenv.CreateCertBundle()

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(certs.CaCert)

	// serverTLSCert := testenv.X509ToTLS(certsclient.ServerCert, nil)
	serverTLSConf = &tls.Config{
		Certificates:       []tls.Certificate{*certs.ServerCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	res := m.Run()

	time.Sleep(time.Second)
	os.Exit(res)
}

func TestNoCA(t *testing.T) {
	path1 := "/hello"
	path1Hit := 0

	// setup server and client environment
	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	mux.HandleFunc(path1, func(http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthCert: path1 hit")
		path1Hit++
	})
	assert.NoError(t, err)
	//
	cl := tlsclient.NewTLSClient(testAddress, nil)
	err = cl.ConnectWithClientCert(certs.PluginCert)
	assert.NoError(t, err)

	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 1, path1Hit)
	cl.Close()

	_, err = cl.ConnectWithLoginID("", "", "", tlsclient.AuthMethodNone)
	assert.NoError(t, err)

	_, err = cl.Get(path1)
	assert.NoError(t, err)
	assert.Equal(t, 2, path1Hit)

	cl.Close()
	srv.Close()
}

// Test certificate based authentication
func TestAuthClientCert(t *testing.T) {
	path1 := "/test1"
	path1Hit := 0

	// setup server and client environment
	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	//
	mux.HandleFunc(path1, func(http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthClientCert: path1 hit")
		path1Hit++
	})
	//
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	err = cl.ConnectWithClientCert(certs.PluginCert)
	assert.NoError(t, err)

	clientCert := cl.Certificate()
	assert.NotNil(t, clientCert)

	//
	_, err = cl.Get(path1)
	assert.NoError(t, err)
	_, err = cl.Post(path1, "")
	assert.NoError(t, err)
	_, err = cl.Put(path1, "")
	assert.NoError(t, err)
	_, err = cl.Delete(path1, "")
	assert.NoError(t, err)
	_, err = cl.Patch(path1, "")
	assert.NoError(t, err)
	assert.Equal(t, 5, path1Hit)

	cl.Close()
	srv.Close()
}

func TestNotStarted(t *testing.T) {
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	_, err := cl.Get("/notstarted")
	assert.Error(t, err)
	cl.Close()
}
func TestNoClientCert(t *testing.T) {
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	err := cl.ConnectWithClientCert(nil)
	assert.Error(t, err)
	cl.Close()
}

func TestBadClientCert(t *testing.T) {
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)

	// use cert not signed by the CA
	otherCA, otherKey := testenv.CreateCA()
	otherCert := testenv.CreateTlsCert("name", "ou", false,
		certs.PluginKey, otherCA, otherKey)

	err := cl.ConnectWithClientCert(otherCert)
	assert.Error(t, err)
	cl.Close()
}

func TestNoServer(t *testing.T) {
	// setup server and client environm
	//
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	err := cl.ConnectWithClientCert(certs.PluginCert)
	assert.NoError(t, err)
	_, err = cl.Get("/noserver")
	assert.Error(t, err)
	cl.Close()
}
func TestCert404(t *testing.T) {
	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)

	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)

	err = cl.ConnectWithClientCert(certs.PluginCert)
	assert.NoError(t, err)

	_, err = cl.Get("/pathnotfound")
	assert.Error(t, err)

	cl.Close()
	srv.Close()
}

func TestAuthBasic(t *testing.T) {
	path2 := "/test2"
	path2Hit := 0
	user1 := "user1"
	password1 := "password1"

	// setup server and client environment
	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	//
	mux.HandleFunc(path2, func(resp http.ResponseWriter, req *http.Request) {
		logrus.Infof("TestAuthBasic: path1 hit")
		username, password, ok := req.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, user1, username)
		assert.Equal(t, password1, password)
		path2Hit++
	})
	//
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	_, err = cl.ConnectWithLoginID(user1, password1, "", tlsclient.AuthMethodBasic)
	assert.NoError(t, err)

	//
	_, err = cl.Get(path2)
	assert.NoError(t, err)
	assert.Equal(t, 1, path2Hit)

	cl.Close()
	srv.Close()
}

func TestAuthJWT(t *testing.T) {
	pathLogin1 := tlsclient.DefaultJWTLoginPath
	pathLogin2 := "/login2"
	path3 := "/test3"
	path3Hit := 0
	user1 := "user1"
	password1 := "password1"
	secret := make([]byte, 64)
	rand.Read(secret)

	// setup server and client environment
	mux := http.NewServeMux()
	// Handle a login
	mux.HandleFunc(pathLogin1, func(resp http.ResponseWriter, req *http.Request) {
		var authMsg tlsclient.JwtAuthLogin
		logrus.Infof("TestAuthJWT: login")
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		err = json.Unmarshal(body, &authMsg)
		assert.NoError(t, err)
		assert.Equal(t, user1, authMsg.LoginID)
		assert.Equal(t, password1, authMsg.Password)
		if authMsg.LoginID == user1 {
			claims := jwt.StandardClaims{
				Id:      user1,
				Issuer:  "me",
				Subject: "accessToken",
				// In JWT, the expiry time is expressed as unix milliseconds
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: time.Now().Add(time.Second).Unix(),
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			accessToken, err := token.SignedString(secret)
			assert.NoError(t, err)
			response := tlsclient.JwtAuthResponse{
				AccessToken:  accessToken,
				RefreshToken: accessToken,
				RefreshURL:   fmt.Sprintf("https://%s/refresh", testAddress),
			}
			data, _ := json.Marshal(response)
			resp.Write(data)
		} else {
			// write nothing
			_ = err
		}
		path3Hit++
	})
	// a second login function that returns nothing
	mux.HandleFunc(pathLogin2, func(resp http.ResponseWriter, req *http.Request) {
	})

	mux.HandleFunc(path3, func(http.ResponseWriter, *http.Request) {
		logrus.Infof("TestAuthJWT: path3 hit")
		path3Hit++
	})
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	//
	//
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	_, err = cl.ConnectWithLoginID(user1, password1)
	assert.NoError(t, err)

	_, err = cl.Get(path3)
	assert.NoError(t, err)
	assert.Equal(t, 2, path3Hit)

	cl.Close()
	srv.Close()
}

func TestAuthRefreshJWT(t *testing.T) {
	pathLogin1 := tlsclient.DefaultJWTLoginPath
	//pathLogin1 := "/login"
	//pathRefresh1 := "/refresh"
	pathRefresh1 := tlsclient.DefaultJWTRefreshPath
	accessToken1 := "accessToken1"
	refreshToken1 := "refreshToken1"
	accessToken2 := "accessToken2"
	refreshToken2 := "refreshToken2"

	// setup server and client environment
	mux := http.NewServeMux()
	mux.HandleFunc(pathLogin1, func(resp http.ResponseWriter, req *http.Request) {
		logrus.Infof("TestAuthRefreshJWT: login")
		tokens := tlsclient.JwtAuthResponse{AccessToken: accessToken1, RefreshToken: refreshToken1}
		data, _ := json.Marshal(tokens)
		resp.Write(data)
	})
	mux.HandleFunc(pathRefresh1, func(resp http.ResponseWriter, req *http.Request) {
		logrus.Infof("TestAuthRefreshJWT: refresh")
		tokens := tlsclient.JwtAuthResponse{AccessToken: accessToken2, RefreshToken: refreshToken2}
		data, _ := json.Marshal(tokens)
		resp.Write(data)
	})
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	//
	//
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	_, err = cl.ConnectWithLoginID("user1", "pw")
	assert.NoError(t, err)

	refreshTokens, err := cl.RefreshJWTTokens("")
	assert.NoError(t, err)
	assert.Equal(t, refreshToken2, string(refreshTokens.RefreshToken))

	cl.Close()
	srv.Close()
}

func TestAuthJWTFail(t *testing.T) {
	pathLogin1 := tlsclient.DefaultJWTLoginPath
	pathLogin2 := "/login2"
	user1 := "user1"
	password1 := "password1"

	// setup server and client environment
	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	//
	mux.HandleFunc(pathLogin1, func(resp http.ResponseWriter, req *http.Request) {
		logrus.Infof("TestAuthJWTFail: login")
		resp.Write([]byte("invalid token"))
	})

	mux.HandleFunc(pathLogin2, func(resp http.ResponseWriter, req *http.Request) {
		logrus.Infof("TestAuthJWTFail: path4 hit, should fail")
		resp.WriteHeader(http.StatusUnauthorized)
	})
	//
	cl := tlsclient.NewTLSClient(testAddress, certs.CaCert)
	_, err = cl.ConnectWithLoginID(user1, password1, "", tlsclient.AuthMethodJwt)
	assert.Error(t, err)

	// cl.SetJWTAuthPaths(pathLogin2, "/wrongrefreshpath")
	_, err = cl.ConnectWithLoginID(user1, password1, testAddress+"/wrongaddress")
	assert.Error(t, err)
	// refresh fails cause path not found
	_, err = cl.RefreshJWTTokens("/wrongrefreshpath")
	assert.Error(t, err)

	cl.Close()
	srv.Close()
}
