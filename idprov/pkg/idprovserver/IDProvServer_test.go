package idprovserver_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/idprov/pkg/idprovserver"
	"github.com/wostzone/hub/lib/client/pkg/testenv"
)

const idProvTestAddr = "127.0.0.1"
const idProvTestPort = 9880

var idProvTestAddrPort = fmt.Sprintf("%s:%d", idProvTestAddr, idProvTestPort)

const clientOobSecret = "secret1"
const certValidityDays = 3
const idprovServiceID = "idprov"

// These are set in TestMain
var clientCertFolder string
var serverCertFolder string

var certStoreFolder string
var device1CaCertPath string
var device1CertPath string
var device1KeyPath string

var testCerts testenv.TestCerts

// var idProvServerIP = idprovserver.GetIPAddr("")
// var idProvServerAddr = idProvServerIP.String()
var homeFolder string
var idpServer *idprovserver.IDProvServer

// easy cleanup for existing device certificate
func removeDeviceCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(clientCertFolder, "*.pem")).Output()
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certStoreFolder, "*.pem")).Output()
}

// func removeServerCerts() {
// 	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(serverCertFolder, "*.pem")).Output()
// }

// TestMain runs a idProv server, gets the directory for futher calls
// Used for all test cases in this package
// NOTE: Don't run tests in parallel as each test creates and deletes certificates
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of idprovserver ------")
	// hostnames := []string{idProvTestAddr}

	const testDiscoveryType = "_test._idprov._tcp"
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../../test")
	serverCertFolder = path.Join(homeFolder, "certs")
	certStoreFolder = path.Join(homeFolder, "certstore")
	clientCertFolder = path.Join(homeFolder, "client")

	// Start test with new certificates
	// logrus.Infof("Creating certificate bundle for names: %s", hostnames)
	removeDeviceCerts()
	// removeServerCerts()
	testCerts = testenv.CreateCertBundle()

	// location where the client saves the provisioned certificates
	device1CertPath = path.Join(clientCertFolder, "device1Cert.pem")
	device1KeyPath = path.Join(clientCertFolder, "device1Key.pem")
	device1CaCertPath = path.Join(clientCertFolder, "caCert.pem")

	idpServer = idprovserver.NewIDProvServer(idprovServiceID,
		idProvTestAddr, idProvTestPort,
		testCerts.ServerCert, testCerts.CaCert, testCerts.CaKey,
		certStoreFolder, certValidityDays,
		testDiscoveryType)

	idpServer.Start()
	res := m.Run()
	idpServer.Stop()
	time.Sleep(time.Second)
	os.Exit(res)
}

func TestStartStopIDProvClient(t *testing.T) {
	// start without existing client cert
	deviceID1 := "device1"
	removeDeviceCerts()
	idprovClient := idprovclient.NewIDProvClient(deviceID1, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)

	// Client start only succeeds if server is running
	err := idprovClient.Start()
	assert.NoError(t, err)

	idprovClient.Stop()
	//// stop the server within the testcase (to count for coverage)
	//idpServer.Stop()
}

func TestStartStopAlreadyRunning(t *testing.T) {
	const testDiscoveryType = "_test._idprov._tcp"

	idpServer2 := idprovserver.NewIDProvServer(idprovServiceID,
		idProvTestAddr, idProvTestPort,
		testCerts.ServerCert,
		testCerts.CaCert, testCerts.CaKey,
		certStoreFolder, certValidityDays,
		testDiscoveryType)

	// can't listen on the same port twice
	err := idpServer2.Start()
	assert.Error(t, err)
	idpServer2.Stop()
}

func TestStartStopMissingPort(t *testing.T) {
	const testDiscoveryType = "_test._idprov._tcp"

	idpServer2 := idprovserver.NewIDProvServer(idprovServiceID,
		idProvTestAddr, 0,
		testCerts.ServerCert,
		testCerts.CaCert, testCerts.CaKey,
		certStoreFolder, certValidityDays,
		testDiscoveryType)

	err := idpServer2.Start()
	assert.Error(t, err)
}

func TestStartStopJustPort(t *testing.T) {
	const testDiscoveryType = "_test._idprov._tcp"

	// listen on the port on all addresses. Discovery will fail though
	idpServer2 := idprovserver.NewIDProvServer(idprovServiceID,
		"", idProvTestPort+1,
		testCerts.ServerCert,
		testCerts.CaCert, testCerts.CaKey,
		certStoreFolder, certValidityDays,
		testDiscoveryType)

	err := idpServer2.Start()
	assert.NoError(t, err)
}

func TestStartStopBadCertFolder(t *testing.T) {
	deviceID1 := "device1"
	removeDeviceCerts()
	idprovClient := idprovclient.NewIDProvClient(deviceID1,
		idProvTestAddrPort,
		"/bad/cert/path", "/bad/key", device1CaCertPath)
	err := idprovClient.Start()
	assert.Error(t, err)

	idprovClient.Stop()
}

func TestStartStopMissingCA(t *testing.T) {
	const testDiscoveryType = "_test._idprov._tcp"
	const idProvTestPort2 = 9998

	idpServer2 := idprovserver.NewIDProvServer(idprovServiceID,
		idProvTestAddr, idProvTestPort2,
		testCerts.ServerCert,
		nil, testCerts.CaKey,
		certStoreFolder, certValidityDays,
		testDiscoveryType)

	err := idpServer2.Start()
	assert.Error(t, err)

	idpServer2 = idprovserver.NewIDProvServer(idprovServiceID,
		idProvTestAddr, idProvTestPort2,
		testCerts.ServerCert,
		testCerts.CaCert, nil,
		certStoreFolder, certValidityDays,
		testDiscoveryType)

	err = idpServer2.Start()
	assert.Error(t, err)

	//
	idpServer2.Stop()
}

// func TestOutboundIP(t *testing.T) {

// 	// test server has an IP
// 	ip := hubnet.GetOutboundIP("")
// 	assert.NotEmpty(t, ip)

// 	// test invalid destination
// 	ip = hubnet.GetOutboundIP("---")
// 	assert.Empty(t, ip)
// }

func TestGetDirectory(t *testing.T) {
	deviceID1 := "device1"
	removeDeviceCerts()
	idpClient := idprovclient.NewIDProvClient(deviceID1,
		idProvTestAddrPort, device1CertPath, device1KeyPath, device1CaCertPath)
	idpClient.Start()
	//
	directory, err := idpClient.GetDirectory()
	assert.NoError(t, err)
	assert.NotNil(t, directory)

	serverDir := idpServer.Directory()
	assert.Equal(t, serverDir.Version, directory.Version)
	assert.Equal(t, serverDir.Services, directory.Services)
	assert.Equal(t, serverDir.Endpoints.GetDirectory, directory.Endpoints.GetDirectory)
	idpClient.Stop()
}

func TestGetDeviceStatusFailAuth(t *testing.T) {
	const deviceID1 = "device1"
	var err error

	removeDeviceCerts()

	idpClient := idprovclient.NewIDProvClient(deviceID1, idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)
	idpClient.Start()

	// authentication is required
	_, err = idpClient.GetDeviceStatus(deviceID1)
	assert.Error(t, err)
	idpClient.Stop()

}
