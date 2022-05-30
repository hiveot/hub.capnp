// Package idprovclient_test with Test Main shared by all tests
package idprovclient_test

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/idprovclient"
	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/testenv"
	"os"
	"path"
	"testing"
)

// IoT device test environment certificate files and folders
var (
	clientCertFolder = ""
	caCertPath       = ""
	clientCertPath   = ""
	clientKeyPath    = ""
)

// server environment certificates
var testCerts testenv.TestCerts = testenv.CreateCertBundle()

// This must match the use of {deviceID} of the protocol definition
var mockDirectory = idprovclient.GetDirectoryMessage{
	Endpoints: idprovclient.DirectoryEndpoints{
		GetDirectory:            idprovclient.IDProvDirectoryPath,
		PostProvisioningRequest: "/idprov/provreq",
		GetDeviceStatus:         "/idprov/status/{deviceID}",
		PostOobSecret:           "/idprov/oobSecret",
		// PostOOB:    "/idprov/oob",
	},
	CaCertPEM: []byte(certsclient.X509CertToPEM(testCerts.CaCert)),
	Version:   "1",
}

// TestMain creates the environment for running the client tests
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	logrus.Infof("------ TestMain of idprov client ------")

	// hostnames := []string{idProvServerAddr}
	//cwd, _ := os.Getwd()
	//homeFolder := path.Join(cwd, "../../test")
	//clientCertFolder = path.Join(homeFolder, "clientcerts")
	tempFolder := os.TempDir()
	clientCertFolder := path.Join(tempFolder, "wost-idprovclient-test")
	_ = os.MkdirAll(clientCertFolder, 0700)
	clientCertPath = path.Join(clientCertFolder, "clientCert.pem")
	clientKeyPath = path.Join(clientCertFolder, "clientKey.pem")
	caCertPath = path.Join(clientCertFolder, "caCert.pem")

	//testCerts = testenv.CreateCertBundle()
	//mockDirectory.CaCertPEM = []byte(certsclient.X509CertToPEM(testCerts.CaCert))

	res := m.Run()
	if res == 0 {
		_ = os.RemoveAll(clientCertFolder)
	}
	os.Exit(res)
}
