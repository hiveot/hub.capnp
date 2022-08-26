package selfsigned_test

import (
	"crypto/x509"
	"os"
	"path"
	"testing"

	"github.com/wostzone/wost-go/pkg/certsclient"
	"github.com/wostzone/wost-go/pkg/logging"

	"github.com/wostzone/hub/pkg/svc/certsvc/selfsigned"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tempFolder string
var certFolder string

const TempCertDurationDays = 1

// removeCerts easy cleanup for existing device certificate
//func removeServerCerts() {
//	_, _ = exec.Command("sh", "-c", "rm -f "+path.Join(certFolder, "*.pem")).Output()
//}

// TestMain clears the certs folder for clean testing
func TestMain(m *testing.M) {
	logging.SetLogging("info", "")
	tempFolder := path.Join(os.TempDir(), "wost-certs-test")
	// clean start
	os.RemoveAll(tempFolder)
	certFolder = path.Join(tempFolder, "certs")
	_ = os.MkdirAll(certFolder, 0700)
	logging.SetLogging("info", "")
	//removeServerCerts()

	res := m.Run()
	if res == 0 {
		os.RemoveAll(tempFolder)
	}
	os.Exit(res)
}

func TestCreateCA(t *testing.T) {
	// test creating hub CA certificate
	caCert, caKeys, err := selfsigned.CreateHubCA(1)
	assert.NoError(t, err)
	require.NotNil(t, caCert)
	require.NotNil(t, caKeys)
}

func TestClientCertBadCA(t *testing.T) {
	clientID := "client1"
	ou := certsclient.OUClient
	caCert, caKey, err := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()

	clientCert, err := selfsigned.CreateClientCert(clientID, ou,
		&keys.PublicKey, nil, caKey, TempCertDurationDays)
	assert.Error(t, err)
	assert.Empty(t, clientCert)

	clientCert, err = selfsigned.CreateClientCert(clientID, ou,
		&keys.PublicKey, caCert, nil, TempCertDurationDays)
	assert.Error(t, err)
	assert.Empty(t, clientCert)
}

func TestCreateServiceCert(t *testing.T) {
	// test creating hub certificate
	const serviceID = "testService"
	names := []string{"127.0.0.1", "localhost"}
	caCert, caKey, err := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()

	serviceCert, err := selfsigned.CreateServiceCert(
		serviceID, names, &keys.PublicKey, caCert, caKey, 1)
	require.NoError(t, err)
	require.NotNil(t, serviceCert)
	require.NotNil(t, serviceCert.PublicKey)

	// verify service certificate against CA
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)
	opts := x509.VerifyOptions{
		Roots:     caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	_, err = serviceCert.Verify(opts)
	assert.NoError(t, err)
}

// test with bad parameters
func TestServiceCertBadParms(t *testing.T) {
	const serviceID = "testService"
	hostnames := []string{"127.0.0.1"}
	caCert, caKey, err := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()
	// missing CA private key
	hubCert, err := selfsigned.CreateServiceCert(
		serviceID, hostnames, &keys.PublicKey, caCert, nil, 1)
	require.Error(t, err)
	require.Empty(t, hubCert)

	// missing service ID
	hubCert, err = selfsigned.CreateServiceCert(
		"", hostnames, nil, nil, caKey, 1)
	require.Error(t, err)
	require.Empty(t, hubCert)

	// missing public key
	hubCert, err = selfsigned.CreateServiceCert(
		serviceID, hostnames, nil, nil, caKey, 1)
	require.Error(t, err)
	require.Empty(t, hubCert)

	// missing CA certificate
	hubCert, err = selfsigned.CreateServiceCert(
		serviceID, hostnames, &keys.PublicKey, nil, caKey, 1)
	require.Error(t, err)
	require.Empty(t, hubCert)

	// Bad CA certificate
	badCa := x509.Certificate{}
	hubCert, err = selfsigned.CreateServiceCert(
		serviceID, hostnames, &keys.PublicKey, &badCa, caKey, 1)
	require.Error(t, err)
	require.Empty(t, hubCert)
}
func TestCreateClientCert(t *testing.T) {
	clientID := "plugin1"
	ou := certsclient.OUClient
	// test creating hub certificate
	caCert, caKeys, err := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()

	clientCert, err := selfsigned.CreateClientCert(
		clientID, ou, &keys.PublicKey, caCert, caKeys, 1)
	require.NoErrorf(t, err, "TestServiceCert: Failed creating server certificate")
	require.NotNil(t, clientCert)

	// verify client certificate against CA
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)
	opts := x509.VerifyOptions{
		Roots:     caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	_, err = clientCert.Verify(opts)
	assert.NoError(t, err)
}

func TestCreateDeviceCert(t *testing.T) {
	deviceID := "device1"
	ou := certsclient.OUIoTDevice
	// test creating hub certificate
	caCert, caKeys, err := selfsigned.CreateHubCA(1)
	keys := certsclient.CreateECDSAKeys()

	deviceCert, err := selfsigned.CreateClientCert(
		deviceID, ou, &keys.PublicKey, caCert, caKeys, 1)
	require.NoErrorf(t, err, "TestServiceCert: Failed creating device certificate")
	require.NotNil(t, deviceCert)

	// verify certificate against CA
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)
	opts := x509.VerifyOptions{
		Roots:     caCertPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	_, err = deviceCert.Verify(opts)
	assert.NoError(t, err)
}

//
//func TestCreateBundle(t *testing.T) {
//	hostnames := []string{"127.0.0.1"}
//
//	// test creating hub CA certificate
//	err := selfsigned.CreateCertificateBundle(hostnames, certFolder, true)
//	require.NoError(t, err)
//}
//
//func TestCreateBundleBadFolder(t *testing.T) {
//	hostnames := []string{"127.0.0.1"}
//
//	// test creating hub CA certificate
//	err := selfsigned.CreateCertificateBundle(hostnames, "/not/a/valid/folder", true)
//	require.Error(t, err)
//}
//
//func TestCreateBundleBadNames(t *testing.T) {
//	// test creating hub CA certificate
//	err := selfsigned.CreateCertificateBundle(nil, certFolder, true)
//	require.Error(t, err)
//}
