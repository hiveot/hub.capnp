package idprovclient_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/lib/client/pkg/certsclient"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
)

// no hostname in certs
const idProvServerAddr = "127.0.0.1:4444"

var clientCertFolder = ""

// var serverCertFolder = ""
var caCertPath = ""

// var caKeyPath = ""
var clientCertPath = ""
var clientKeyPath = ""

// var mock1 *tlsserver.TLSServer

// This must match the use of deviceID of the protocol definition
var mockDirectory = idprovclient.GetDirectoryMessage{
	Endpoints: idprovclient.DirectoryEndpoints{
		GetDirectory:            idprovclient.IDProvDirectoryPath,
		PostProvisioningRequest: "/idprov/provreq",
		GetDeviceStatus:         "/idprov/status/{deviceID}",
		PostOobSecret:           "/idprov/oobSecret",
		// PostOOB:    "/idprov/oob",
	},
	CaCertPEM: nil, // create in testmain
	Version:   "1",
}

var testCerts testenv.TestCerts

func startTestServer(mux *http.ServeMux) (*http.Server, error) {
	var err error
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(testCerts.CaCert)

	// serverTLSCert := testenv.X509ToTLS(certs.ServerCert, nil)
	serverTLSConf := &tls.Config{
		Certificates:       []tls.Certificate{*testCerts.ServerCert},
		ClientAuth:         tls.VerifyClientCertIfGiven,
		ClientCAs:          caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	httpServer := &http.Server{
		Addr: idProvServerAddr,
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

// easy cleanup for existing device certificate
func removeDeviceCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(clientCertFolder, "*.pem")).Output()
}

// func removeServerCerts() {
// 	exec.Command("sh", "-c", "rm -f "+path.Join(serverCertFolder, "*.pem")).Output()
// }

// TestMain creates the environment for running the client tests
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of idprov client ------")
	// hostnames := []string{idProvServerAddr}
	cwd, _ := os.Getwd()
	homeFolder := path.Join(cwd, "../../test")
	clientCertFolder = path.Join(homeFolder, "clientcerts") // where the client saves its issued certificate
	clientCertPath = path.Join(clientCertFolder, "clientCert.pem")
	clientKeyPath = path.Join(clientCertFolder, "clientKey.pem")
	caCertPath = path.Join(clientCertFolder, "caCert.pem")

	testCerts = testenv.CreateCertBundle()
	mockDirectory.CaCertPEM = []byte(certsclient.X509CertToPEM(testCerts.CaCert))

	res := m.Run()

	os.Exit(res)
}

func TestStartStop(t *testing.T) {
	const deviceID = "device1"
	serverMux := http.NewServeMux()
	mock1, err := startTestServer(serverMux)
	assert.NoError(t, err)
	serverMux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})

	// no client cert
	removeDeviceCerts()
	idpClient := idprovclient.NewIDProvClient(deviceID, idProvServerAddr,
		clientCertPath, clientKeyPath, caCertPath)
	// this should load the CA certificate in the client folder
	err = idpClient.Start()
	assert.NoError(t, err)
	// starting twice should be fine
	err = idpClient.Start()
	assert.NoError(t, err)

	pk := idpClient.PublicKeyPEM()
	assert.NotNil(t, pk)
	dir := idpClient.Directory()
	assert.NotNil(t, dir)
	// obip := idpClient.OutboundIP()
	// assert.NotEmpty(t, obip)
	_, err = os.Stat(caCertPath)
	assert.NoErrorf(t, err, "CA cert wasn't obtained and stored")

	idpClient.Stop()
	_ = mock1.Close()
}

func TestBadDirectory(t *testing.T) {
	const deviceID = "device1"

	serverMux := http.NewServeMux()
	mock1, _ := startTestServer(serverMux)
	serverMux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg := "{this is a bad directory}"
		_, _ = resp.Write([]byte(msg))
	})
	removeDeviceCerts()
	idpClient := idprovclient.NewIDProvClient(deviceID, idProvServerAddr,
		clientCertPath, clientKeyPath, caCertPath)
	// start gets the directory which will fail
	err := idpClient.Start()
	assert.Error(t, err)

	idpClient.Stop()
	_ = mock1.Close()
}

func TestBadClientCertFolder(t *testing.T) {
	const deviceID = "device1"

	idpClient := idprovclient.NewIDProvClient(
		deviceID, idProvServerAddr,
		"/bad/client/certpath.pem", clientKeyPath, caCertPath)
	// start gets the directory which will fail
	err := idpClient.Start()
	assert.Error(t, err)

	idpClient.Stop()
}

// get device status using a valid client certificate
func TestGetDeviceStatus(t *testing.T) {
	const deviceID1 = "device1"
	var rxDeviceID string
	var nrClientCerts int
	var err error

	serverMux := http.NewServeMux()
	mock1, _ := startTestServer(serverMux)
	serverMux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})
	statusPath := strings.Replace(mockDirectory.Endpoints.GetDeviceStatus, "{deviceID}", deviceID1, 1)
	serverMux.HandleFunc(statusPath, func(resp http.ResponseWriter, req *http.Request) {
		nrClientCerts += len(req.TLS.PeerCertificates)
		// fake get status endpoint
		logrus.Infof("Mock: GET device status: %d client certificates provided", nrClientCerts)
		// we would not have come here if the deviceID was different
		rxDeviceID = deviceID1
		stat := idprovclient.GetDeviceStatusMessage{
			DeviceID: deviceID1,
			Status:   idprovclient.ProvisionStatusWaiting,
		}
		msg, _ := json.Marshal(stat)
		_, _ = resp.Write(msg)
	})
	// invoke the fake get status endpoint - not using a client cert should work
	removeDeviceCerts()
	idpClient := idprovclient.NewIDProvClient(deviceID1, idProvServerAddr,
		clientCertPath, clientKeyPath, caCertPath)
	err = idpClient.Start()
	require.NoError(t, err)
	// this should have downloaded the CA certificate
	assert.FileExists(t, caCertPath)
	stat2, err := idpClient.GetDeviceStatus(deviceID1)
	require.NoError(t, err)
	assert.Equal(t, deviceID1, stat2.DeviceID)
	assert.Equal(t, deviceID1, rxDeviceID)
	idpClient.Stop()
	_ = mock1.Close()
}

// Test the provisioning process
func TestProvision(t *testing.T) {
	const deviceID1 = "device1"
	const deviceSecret = "secret1"
	var provReq idprovclient.PostProvisionRequestMessage
	var rxDeviceID string
	var rxSignature string
	var responseStatus = idprovclient.ProvisionStatusWaiting
	// var clientCertPem string

	// create own mock server to deal with race condition during test
	serverMux := http.NewServeMux()
	mock1, _ := startTestServer(serverMux)
	serverMux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})
	serverMux.HandleFunc(mockDirectory.Endpoints.PostProvisioningRequest, func(resp http.ResponseWriter, req *http.Request) {

		hasClientCert := len(req.TLS.PeerCertificates) > 0
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		err = json.Unmarshal(body, &provReq)
		assert.NoError(t, err)
		rxDeviceID = provReq.DeviceID
		rxSignature = provReq.Signature
		logrus.Infof("TestProvision: from device '%s'", rxDeviceID)

		// reply
		respMsg := idprovclient.PostProvisionResponseMessage{
			RetrySec:  1,
			Status:    responseStatus,
			CaCertPEM: string(mockDirectory.CaCertPEM),
			// ClientCertPEM: clientCertPem,
		}
		// if status is accepted then generate a cert
		// next, create a new pretent certificate to be returned in the idprov request
		if responseStatus == idprovclient.ProvisionStatusApproved {
			clientPubKey, _ := certsclient.PublicKeyFromPEM(provReq.PublicKeyPEM)
			clientCert, _, _ := testenv.CreateX509Cert(rxDeviceID, testenv.OUDevice, false,
				clientPubKey, testCerts.CaCert, testCerts.CaKey)
			respMsg.ClientCertPEM = certsclient.X509CertToPEM(clientCert)
		}
		serialized, _ := json.Marshal(respMsg)
		// the protocol only uses
		deviceSecretToUse := ""
		if !hasClientCert {
			deviceSecretToUse = deviceSecret
		}
		signature, _ := idprovclient.Sign(string(serialized), deviceSecretToUse)
		respMsg.Signature = signature
		msg, _ := json.Marshal(respMsg)
		_, _ = resp.Write(msg)
	})

	// start with a new client and request a certificate
	// this should return waiting
	removeDeviceCerts()
	idpClient := idprovclient.NewIDProvClient(deviceID1, idProvServerAddr,
		clientCertPath, clientKeyPath, caCertPath)
	err := idpClient.Start()
	require.NoError(t, err)

	response, err := idpClient.PostProvisioningRequest("", deviceSecret)
	require.NoError(t, err)
	assert.Equal(t, deviceID1, rxDeviceID)
	assert.NotEmpty(t, rxSignature)
	// assert.NotEmpty(t, response.Signature)
	assert.Equal(t, idprovclient.ProvisionStatusWaiting, response.Status)

	// Next, admin posts an OOB secret and try again. this time it should result in a cert
	responseStatus = idprovclient.ProvisionStatusApproved
	response, err = idpClient.PostProvisioningRequest("", deviceSecret)
	require.NoError(t, err)
	assert.Equal(t, deviceID1, rxDeviceID)
	// assert.Equal(t, deviceSecret, rxDeviceSecret)
	assert.Equal(t, idprovclient.ProvisionStatusApproved, response.Status)
	assert.FileExistsf(t, clientCertPath, "Expected client certificate file")
	// assert.NotNil(t, idpClient.ClientCert())

	// refresh the cert
	response, err = idpClient.PostProvisioningRequest("", deviceSecret)
	require.NoError(t, err)
	assert.Equal(t, idprovclient.ProvisionStatusApproved, response.Status)

	idpClient.Stop()
	_ = mock1.Close()
}

func TestBadAddress(t *testing.T) {
	const deviceID = "device1"

	badAddr := "10.10.255.254" // assume this doesnt exist
	// no client cert
	removeDeviceCerts()
	idpClient := idprovclient.NewIDProvClient(deviceID, badAddr,
		clientCertPath, clientKeyPath, caCertPath)

	err := idpClient.Start()
	assert.Error(t, err)

	_, err = idpClient.GetDeviceStatus("badid")
	assert.Error(t, err)

	_, err = idpClient.PostProvisioningRequest("badid", "notasecret")
	assert.Error(t, err)

	idpClient.Stop()
}

func TestBadCACert(t *testing.T) {
	const deviceID = "device1"

	removeDeviceCerts()
	// use a file that exists but is not a cert
	idpClient := idprovclient.NewIDProvClient(deviceID, idProvServerAddr,
		clientCertPath, clientKeyPath, "/root")

	err := idpClient.Start()
	assert.Error(t, err)

	idpClient.Stop()
}

func TestBadKeys(t *testing.T) {
	const deviceID = "device1"
	// no client cert
	removeDeviceCerts()
	// use a file that exists but is not a cert
	idpClient := idprovclient.NewIDProvClient(deviceID, idProvServerAddr,
		clientCertPath, "/root/clientKeyPath", caCertPath)

	err := idpClient.Start()
	assert.Error(t, err)

	idpClient.Stop()
}

func TestBadCertPaths(t *testing.T) {
	const deviceID1 = "device1"
	const deviceSecret = "secret1"
	var provReq idprovclient.PostProvisionRequestMessage

	serverMux := http.NewServeMux()
	mock1, _ := startTestServer(serverMux)
	serverMux.HandleFunc(idprovclient.IDProvDirectoryPath, func(resp http.ResponseWriter, req *http.Request) {
		msg, _ := json.Marshal(mockDirectory)
		_, _ = resp.Write(msg)
	})
	serverMux.HandleFunc(mockDirectory.Endpoints.PostProvisioningRequest, func(resp http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		err = json.Unmarshal(body, &provReq)
		assert.NoError(t, err)

		// reply
		respMsg := idprovclient.PostProvisionResponseMessage{
			RetrySec:  1,
			Status:    idprovclient.ProvisionStatusApproved,
			CaCertPEM: string(mockDirectory.CaCertPEM),
			// ClientCertPEM: clientCertPem,
		}
		// if status is accepted then generate a cert
		// next, create a new pretent certificate to be returned in the idprov request
		clientPubKey, _ := certsclient.PublicKeyFromPEM(provReq.PublicKeyPEM)
		clientCert, _, _ := testenv.CreateX509Cert(deviceID1, testenv.OUDevice, false,
			clientPubKey, testCerts.CaCert, testCerts.CaKey)
		respMsg.ClientCertPEM = certsclient.X509CertToPEM(clientCert)
		serialized, _ := json.Marshal(respMsg)
		signature, _ := idprovclient.Sign(string(serialized), deviceSecret)
		respMsg.Signature = signature
		msg, _ := json.Marshal(respMsg)
		_, _ = resp.Write(msg)
	})
	// no client cert
	removeDeviceCerts()
	// use a file that exists but is not a cert
	idpClient := idprovclient.NewIDProvClient(deviceID1, idProvServerAddr,
		"/root/badclientcertPath", clientKeyPath, caCertPath)

	err := idpClient.Start()
	require.NoError(t, err)

	_, err = idpClient.PostProvisioningRequest("", deviceSecret)
	assert.Error(t, err)

	idpClient.Stop()
	_ = mock1.Close()
}
