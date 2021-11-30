package oobclient_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/idprov/pkg/oobclient"
	"github.com/wostzone/hub/lib/client/pkg/config"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
)

const testPort = 9999

var idProvServerAddrPort = fmt.Sprintf("%s:%d", testenv.ServerAddress, testPort)

var clientCertFolder = ""
var serverCertFolder = ""
var testCerts testenv.TestCerts

// var mock1 *tlsserver.TLSServer

// This must match the use of deviceID of the protocol definition
var mockDirectory = idprovclient.GetDirectoryMessage{
	Endpoints: idprovclient.DirectoryEndpoints{
		GetDirectory:            idprovclient.IDProvDirectoryPath,
		PostProvisioningRequest: "/idprov/provreq",
		GetDeviceStatus:         "/idprov/status/{deviceID}",
		PostOobSecret:           "/idprov/oobsecret",
	},
	CaCertPEM: nil,
	Version:   "1",
}

func startTestServer(mux *http.ServeMux) (*http.Server, error) {
	var err error

	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(testCerts.CaCert)
	// serverCertPath := path.Join(serverCertFolder, certsetup.HubCertFile)
	// serverKeyPath := path.Join(serverCertFolder, certsetup.HubKeyFile)
	// serverCert, _ := certs.LoadTLSCertFromPEM(serverCertPath, serverKeyPath)
	// serverTLSCert := testenv.X509ToTLS(certs.ServerCert, nil)
	serverTLSConf := &tls.Config{
		Certificates:       []tls.Certificate{*testCerts.ServerCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}
	httpServer := &http.Server{
		Addr:        idProvServerAddrPort,
		ReadTimeout: 5 * time.Minute, // 5 min to allow for delays when testing
		// WriteTimeout: 10 * time.Second,
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

// easy cleanup for existing device certificate
func removeDeviceCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(clientCertFolder, "*.pem")).Output()
}

// func removeServerCerts() {
// 	exec.Command("sh", "-c", "rm -f "+path.Join(serverCertFolder, "*.pem")).Output()
// }

// TestMain prepares certificates
// NOTE: Don't run tests in parallel as each test can create and delete certificates
func TestMain(m *testing.M) {
	config.SetLogging("info", "")

	testCerts = testenv.CreateCertBundle()
	// no need to save the certificates to test the client
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	clientCertFolder = path.Join(homeFolder, "client")
	serverCertFolder = path.Join(homeFolder, "certs")

	res := m.Run()
	// mock1.Stop()
	os.Exit(res)
}

func TestStartStopOOBClient(t *testing.T) {

	// initial usage has no client cert
	removeDeviceCerts()

	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	mux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})

	oobClient := oobclient.NewOOBClient(
		idProvServerAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err = oobClient.Start()
	assert.NoError(t, err)

	// start twice should be fine
	err = oobClient.Start()
	assert.NoError(t, err)

	// get the IDProv directory should always work
	dir := oobClient.Directory()
	assert.NotNil(t, dir)

	oobClient.Stop()
	srv.Close()
}

func TestStartStopOOBBadPath(t *testing.T) {
	// initial usage has no client cert
	removeDeviceCerts()

	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	mux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})

	oobClient := oobclient.NewOOBClient(
		idProvServerAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err = oobClient.Start()
	assert.NoError(t, err)
	// Post something
	_, err = oobClient.Post("/notavalid/path", "payload")
	assert.Error(t, err)

	oobClient.Stop()
	srv.Close()
}

func TestStartBadCert(t *testing.T) {

	// the server certificate cannot be used as a client cert
	oobClient := oobclient.NewOOBClient(idProvServerAddrPort,
		testCerts.ServerCert, testCerts.CaCert)
	err := oobClient.Start()
	assert.Error(t, err)

	oobClient.Stop()
}

func TestPostOOB(t *testing.T) {
	const deviceID1 = "device1"
	const deviceSecret = "secret1"
	var oobMsg idprovclient.PostOobSecretMessage
	var rxDeviceID string
	var rxDeviceSecret string
	removeDeviceCerts()

	// separate mock server as testing from commandline gives 404 error
	mux := http.NewServeMux()
	srv, err := startTestServer(mux)
	assert.NoError(t, err)
	mux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})
	mux.HandleFunc(mockDirectory.Endpoints.PostOobSecret, func(resp http.ResponseWriter, req *http.Request) {
		nrClientCerts := len(req.TLS.PeerCertificates)
		logrus.Infof("TestPostOOB: %d client certificates provided", nrClientCerts)
		logrus.Infof("TestPostOOB: from device %s, secret=%s", rxDeviceID, rxDeviceSecret)

		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		json.Unmarshal(body, &oobMsg)
		rxDeviceID = oobMsg.DeviceID
		rxDeviceSecret = string(oobMsg.Secret)
		// _, _ = resp.Write(msg)
	})

	oobClient := oobclient.NewOOBClient(
		idProvServerAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err = oobClient.Start()
	assert.NoError(t, err)

	_, err = oobClient.PostOOB(deviceID1, deviceSecret)
	assert.NoError(t, err)
	assert.Equal(t, deviceID1, rxDeviceID)
	assert.Equal(t, deviceSecret, rxDeviceSecret)

	srv.Close()
	oobClient.Stop()
}

func TestBadServerAddress(t *testing.T) {
	badAddrPort := "127.0.0.2" // assume this doesnt exist
	oobClient := oobclient.NewOOBClient(
		badAddrPort, testCerts.PluginCert, testCerts.CaCert)
	err := oobClient.Start()
	assert.Error(t, err)

	// _, err = oobClient.PostOOB("deviceid", "secret")
	// assert.Error(t, err)

	oobClient.Stop()
}
