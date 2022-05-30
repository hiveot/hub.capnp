// Package idprovserver_test with TestMain test setup
package idprovserver_test

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/idprov/pkg/idprovserver"
	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/testenv"
	"os"
	"path"
	"testing"
	"time"
)

const idProvTestAddr = "127.0.0.1"
const idProvTestPort = 9880

const clientOobSecret = "secret1"
const certValidityDays = 1

var idProvTestAddrPort = fmt.Sprintf("%s:%d", idProvTestAddr, idProvTestPort)

// IoT device client environment certificate files and folders
var (
	clientCertFolder  string
	device1CaCertPath string
	device1CertPath   string
	device1KeyPath    string
)

// cert folder for CA, server and plugin certificates
//var serverCertFolder string

var idpConfig idprovserver.IDProvConfig
var testCerts testenv.TestCerts

// var idProvServerIP = idprovserver.GetIPAddr("")
// var idProvServerAddr = idProvServerIP.String()
var idpServer *idprovserver.IDProvServer

// TestMain runs a idProv server, gets the directory for futher calls
// Used for all test cases in this package
// NOTE: Don't run tests in parallel as each test creates and deletes certificates
func TestMain(m *testing.M) {
	logrus.Infof("------ TestMain of idprovserver ------")
	// hostnames := []string{idProvTestAddr}
	logging.SetLogging("info", "")

	tempFolder := os.TempDir()
	testFolder := path.Join(tempFolder, "wost-idprovserver-test")
	certStoreFolder := path.Join(testFolder, "certstore")
	clientCertFolder = path.Join(testFolder, "clientcert")
	_ = os.MkdirAll(certStoreFolder, 0700)
	_ = os.MkdirAll(clientCertFolder, 0700)

	// Start test with new certificates
	// logrus.Infof("Creating certificate bundle for names: %s", hostnames)
	removeDeviceCerts()
	// removeServerCerts()
	testCerts = testenv.CreateCertBundle()

	// location where the client saves the provisioned certificates
	device1CertPath = path.Join(clientCertFolder, "device1Cert.pem")
	device1KeyPath = path.Join(clientCertFolder, "device1Key.pem")
	device1CaCertPath = path.Join(clientCertFolder, "caCert.pem")

	idpConfig = idprovserver.IDProvConfig{
		IdpAddress:       idProvTestAddr,
		IdpPort:          idProvTestPort,
		CertStoreFolder:  certStoreFolder,
		CertValidityDays: certValidityDays,
		ServiceName:      "test", // discovery record looks like _test._idprov._tcp
	}
	idpServer = idprovserver.NewIDProvServer(&idpConfig,
		testCerts.ServerCert, testCerts.CaCert, testCerts.CaKey)

	_ = idpServer.Start()

	res := m.Run()

	idpServer.Stop()
	time.Sleep(time.Second)
	if res == 0 {
		_ = os.RemoveAll(testFolder)
	} else {
		logrus.Print("Test files can be found in ", testFolder)
	}

	os.Exit(res)
}
