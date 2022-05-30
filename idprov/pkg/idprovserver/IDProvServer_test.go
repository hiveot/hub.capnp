package idprovserver_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/hub/idprov/pkg/idprovserver"
	"os/exec"
	"path"
	"testing"
)

// easy cleanup for existing device certificate
func removeDeviceCerts() {
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(clientCertFolder, "*.pem")).Output()
	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(idpConfig.CertStoreFolder, "*.pem")).Output()
}

// func removeServerCerts() {
// 	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(serverCertFolder, "*.pem")).Output()
// }

// test using the idprov client
func TestStartStopIDProvClient(t *testing.T) {
	// start without existing client cert
	deviceID1 := "device1"
	removeDeviceCerts()
	idprovClient := idprovclient.NewIDProvClient(deviceID1,
		idProvTestAddrPort,
		device1CertPath, device1KeyPath, device1CaCertPath)

	// Client start only succeeds if server is running
	err := idprovClient.Start()
	assert.NoError(t, err)

	idprovClient.Stop()
}

func TestStartStopAlreadyRunning(t *testing.T) {
	const testDiscoveryType = "_test._idprov._tcp"

	idpServer2 := idprovserver.NewIDProvServer(&idpConfig,
		testCerts.ServerCert,
		testCerts.CaCert, testCerts.CaKey)

	// can't listen on the same port twice
	err := idpServer2.Start()
	assert.Error(t, err)
	idpServer2.Stop()
}

func TestStartStopMissingPort(t *testing.T) {

	config2 := idpConfig
	config2.IdpPort = 0
	idpServer2 := idprovserver.NewIDProvServer(&config2,
		testCerts.ServerCert,
		testCerts.CaCert, testCerts.CaKey)
	err := idpServer2.Start()
	// should have used default port
	assert.NoError(t, err)
	idpServer2.Stop()
}

func TestStartStopJustPort(t *testing.T) {

	// listen on the port on all addresses. Discovery will fail though
	config2 := idpConfig
	config2.IdpAddress = ""
	config2.IdpPort = idProvTestPort + 1
	idpServer2 := idprovserver.NewIDProvServer(&config2,
		testCerts.ServerCert,
		testCerts.CaCert, testCerts.CaKey)

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
	//
	config2 := idpConfig
	config2.IdpAddress = idProvTestAddr
	config2.IdpPort = idProvTestPort2
	idpServer2 := idprovserver.NewIDProvServer(&config2,
		testCerts.ServerCert,
		nil, testCerts.CaKey)

	err := idpServer2.Start()
	assert.Error(t, err)

	idpServer2 = idprovserver.NewIDProvServer(&config2,
		testCerts.ServerCert,
		testCerts.CaCert, nil)

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
	err := idpClient.Start()
	assert.NoError(t, err)
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
	err = idpClient.Start()
	assert.NoError(t, err)

	// authentication is required
	_, err = idpClient.GetDeviceStatus(deviceID1)
	assert.Error(t, err)
	idpClient.Stop()

}
